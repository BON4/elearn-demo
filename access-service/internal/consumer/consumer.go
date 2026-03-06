package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/BON4/elearn-demo/access-service/internal/domain"
	"github.com/BON4/elearn-demo/contracts"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type CourcesService interface {
	ProcessPublishedCourseEvent(ctx context.Context, course domain.CourseRM, eventID uuid.UUID) error
	ProcessDraftedCourseEvent(ctx context.Context, course domain.CourseRM, eventID uuid.UUID) error
}

type AccessService interface {
	ProcessPaymentSuccesedEvent(ctx context.Context, courseID uuid.UUID, userID uuid.UUID, eventID uuid.UUID) error
	ProcessPaymentRefoundedEvent(ctx context.Context, courseID uuid.UUID, userID uuid.UUID, eventID uuid.UUID) error
}

type Consumer struct {
	broker        *amqp.Connection
	courseService CourcesService
	accessService AccessService
	consumerTag   string
	interval      time.Duration
	paused        *atomic.Bool
	resumeCh      chan struct{}
}

const (
	CourseQueue  = "course-access"
	PaymentQueue = "payment-access"
)

func NewConsumer(
	conn *amqp.Connection,
	consumerTag string,
	courseService CourcesService,
	accessService AccessService,
	interval time.Duration,
) *Consumer {
	paused := atomic.Bool{}
	paused.Store(false)

	return &Consumer{
		broker:        conn,
		courseService: courseService,
		accessService: accessService,
		consumerTag:   consumerTag,
		interval:      interval,
		paused:        &paused,
		resumeCh:      make(chan struct{}, 1),
	}
}

func (w *Consumer) PauseWorker() {
	w.paused.Store(true)
	log.Info("consumer worker paused")
}

func (w *Consumer) ResumeWorker() {
	w.paused.Store(false)
	log.Info("consumer worker resumed")
	select {
	case w.resumeCh <- struct{}{}:
	default:
	}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if c.paused.Load() {
			select {
			case <-ctx.Done():
				return
			case <-c.resumeCh:
			}
			continue
		}

		g, gCtx := errgroup.WithContext(ctx)

		g.Go(func() error {
			return c.HandleCoursesExchange(gCtx)
		})

		g.Go(func() error {
			return c.HandlePaymentsExchange(gCtx)
		})

		err := g.Wait()
		if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}

		log.WithError(err).Warn("consumer error, restarting in 5s...")

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * c.interval):
		}
	}
}

func (c *Consumer) setupPaymentQueue() (*amqp.Channel, <-chan amqp.Delivery, error) {
	ch, err := c.broker.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("open channel: %w", err)
	}

	paymentQueue, err := ch.QueueDeclare(
		PaymentQueue,
		true, false, false, false, nil,
	)
	if err != nil {
		ch.Close()
		return nil, nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(
		paymentQueue.Name,
		contracts.PaymentSucceededRoutingKey,
		contracts.PaymentsExchange,
		false, nil,
	); err != nil {
		ch.Close()
		return nil, nil, fmt.Errorf("bind queue: %w", err)
	}

	if err := ch.QueueBind(
		paymentQueue.Name,
		contracts.PaymentRefoundedRoutingKey,
		contracts.PaymentsExchange,
		false, nil,
	); err != nil {
		ch.Close()
		return nil, nil, fmt.Errorf("bind queue: %w", err)
	}

	msgs, err := ch.Consume(
		PaymentQueue,
		c.consumerTag,
		false, false, false, false, nil,
	)
	if err != nil {
		ch.Close()
		return nil, nil, fmt.Errorf("consume: %w", err)
	}

	return ch, msgs, nil
}

func (c *Consumer) setupCourseQueue() (*amqp.Channel, <-chan amqp.Delivery, error) {
	ch, err := c.broker.Channel()
	if err != nil {
		return nil, nil, fmt.Errorf("open channel: %w", err)
	}

	courseQueue, err := ch.QueueDeclare(
		CourseQueue,
		true, false, false, false, nil,
	)
	if err != nil {
		ch.Close()
		return nil, nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := ch.QueueBind(
		courseQueue.Name,
		contracts.CoursePublishedRoutingKey,
		contracts.CoursesExchange,
		false, nil,
	); err != nil {
		ch.Close()
		return nil, nil, fmt.Errorf("bind queue: %w", err)
	}

	if err := ch.QueueBind(
		courseQueue.Name,
		contracts.CourseDraftedRoutingKey,
		contracts.CoursesExchange,
		false, nil,
	); err != nil {
		ch.Close()
		return nil, nil, fmt.Errorf("bind queue: %w", err)
	}

	msgs, err := ch.Consume(
		CourseQueue,
		c.consumerTag,
		false, false, false, false, nil,
	)
	if err != nil {
		ch.Close()
		return nil, nil, fmt.Errorf("consume: %w", err)
	}

	return ch, msgs, nil
}

func (c *Consumer) HandlePaymentsExchange(ctx context.Context) error {
	ch, paymentMsg, err := c.setupPaymentQueue()
	if err != nil {
		return err
	}
	defer ch.Close()

	log.Info("payment consumer started, waiting for messages")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case msg, ok := <-paymentMsg:
			if !ok {
				return errors.New("delivery channel closed")
			}

			if err := c.handleMessage(ctx, msg); err != nil {
				log.WithError(err).
					WithFields(log.Fields{
						"routing_key": msg.RoutingKey,
						"message_id":  msg.MessageId,
					}).Error("failed to handle message")

				if errors.Is(err, domain.ErrEventAlreadyProcessed) {
					_ = msg.Ack(false)
					continue
				}

				_ = msg.Nack(false, false)
				continue
			}

			_ = msg.Ack(false)
		}
	}
}

func (c *Consumer) HandleCoursesExchange(ctx context.Context) error {
	ch, courseMsg, err := c.setupCourseQueue()
	if err != nil {
		return err
	}
	defer ch.Close()

	log.Info("course consumer started, waiting for messages")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case msg, ok := <-courseMsg:
			if !ok {
				return errors.New("delivery channel closed")
			}

			if err := c.handleMessage(ctx, msg); err != nil {
				log.WithError(err).
					WithFields(log.Fields{
						"routing_key": msg.RoutingKey,
						"message_id":  msg.MessageId,
					}).Error("failed to handle message")

				if errors.Is(err, domain.ErrEventAlreadyProcessed) {
					_ = msg.Ack(false)
					continue
				}

				_ = msg.Nack(false, false)
				continue
			}

			_ = msg.Ack(false)
		}
	}
}

func (c *Consumer) handleMessage(ctx context.Context, msg amqp.Delivery) error {
	if msg.MessageId == "" {
		return errors.New("missing message id")
	}

	log.WithField("route", msg.RoutingKey).
		WithField("msg_id", msg.MessageId).
		Info("got message")

	switch msg.RoutingKey {
	case contracts.CoursePublishedRoutingKey:
		return c.handleCoursePublishedMessage(ctx, msg)
	case contracts.CourseDraftedRoutingKey:
		return c.handleCourseDraftedMessage(ctx, msg)
	case contracts.PaymentSucceededRoutingKey:
		return c.handlePaymentSucceededMessage(ctx, msg)
	case contracts.PaymentRefoundedRoutingKey:
		return c.handlePaymentRefoundedMessage(ctx, msg)
	default:
		log.WithField("routing_key", msg.RoutingKey).Info("uknown routing key, skipping")
		return nil
	}
}

func (c *Consumer) handleCoursePublishedMessage(ctx context.Context, msg amqp.Delivery) error {
	var coursePayload contracts.CoursePublishedEventPayload
	if err := json.Unmarshal(msg.Body, &coursePayload); err != nil {
		return err
	}

	eventID, err := uuid.Parse(msg.MessageId)
	if err != nil {
		return err
	}

	processCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.courseService.ProcessPublishedCourseEvent(processCtx, *domain.PublishedCourceFromContract(coursePayload), eventID)
}

func (c *Consumer) handleCourseDraftedMessage(ctx context.Context, msg amqp.Delivery) error {
	var coursePayload contracts.CourseDraftedEventPayload
	if err := json.Unmarshal(msg.Body, &coursePayload); err != nil {
		return err
	}

	eventID, err := uuid.Parse(msg.MessageId)
	if err != nil {
		return err
	}

	processCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.courseService.ProcessDraftedCourseEvent(processCtx, *domain.DraftedCourceFromContract(coursePayload), eventID)
}

func (c *Consumer) handlePaymentSucceededMessage(ctx context.Context, msg amqp.Delivery) error {
	var paymentPayload contracts.PaymentSucceededEventPayload
	if err := json.Unmarshal(msg.Body, &paymentPayload); err != nil {
		return err
	}

	eventID, err := uuid.Parse(msg.MessageId)
	if err != nil {
		return err
	}

	processCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.accessService.ProcessPaymentSuccesedEvent(processCtx, paymentPayload.CourseID, paymentPayload.UserID, eventID)
}

func (c *Consumer) handlePaymentRefoundedMessage(ctx context.Context, msg amqp.Delivery) error {
	var paymentPayload contracts.PaymentSucceededEventPayload
	if err := json.Unmarshal(msg.Body, &paymentPayload); err != nil {
		return err
	}

	eventID, err := uuid.Parse(msg.MessageId)
	if err != nil {
		return err
	}

	processCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return c.accessService.ProcessPaymentRefoundedEvent(processCtx, paymentPayload.CourseID, paymentPayload.UserID, eventID)
}
