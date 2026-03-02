package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/BON4/elearn-demo/access-service/internal/domain"
	"github.com/BON4/elearn-demo/contracts"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type CourcesService interface {
	ProcessPublishedCourseEvent(ctx context.Context, course domain.CourseRM, eventID uuid.UUID) error
	ProcessDraftedCourseEvent(ctx context.Context, course domain.CourseRM, eventID uuid.UUID) error
}

type Consumer struct {
	broker        *amqp.Connection
	consumeCh     *amqp.Channel
	courseService CourcesService
	queueName     string
	consumerTag   string
	interval      time.Duration
	paused        *atomic.Bool
}

const (
	CoursePublishedQueue = "course-access"
)

func NewConsumer(
	conn *amqp.Connection,
	consumerTag string,
	service CourcesService,
	interval time.Duration,
) *Consumer {
	paused := atomic.Bool{}
	paused.Store(false)

	return &Consumer{
		broker:        conn,
		courseService: service,
		consumerTag:   consumerTag,
		interval:      interval,
		paused:        &paused,
	}
}

func (w *Consumer) PauseWorker() {
	w.paused.Store(true)
	log.Info("consumer worker paused")
}

func (w *Consumer) ResumeWorker() {
	w.paused.Store(false)
	log.Info("consumer worker resumed")
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if c.paused.Load() {
			continue
		}

		err := c.HandleCoursesExchange(ctx)
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

func (c *Consumer) setupQueue() (<-chan amqp.Delivery, error) {
	if c.consumeCh == nil || c.consumeCh.IsClosed() {
		ch, err := c.broker.Channel()
		if err != nil {
			return nil, fmt.Errorf("open channel: %w", err)
		}
		c.consumeCh = ch
	}

	queue, err := c.consumeCh.QueueDeclare(
		CoursePublishedQueue,
		true, false, false, false, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	if err := c.consumeCh.QueueBind(
		queue.Name,
		contracts.CoursePublishedRoutingKey,
		contracts.CoursesExchange,
		false, nil,
	); err != nil {
		return nil, fmt.Errorf("bind queue: %w", err)
	}

	if err := c.consumeCh.QueueBind(
		queue.Name,
		contracts.CourseDraftedRoutingKey,
		contracts.CoursesExchange,
		false, nil,
	); err != nil {
		return nil, fmt.Errorf("bind queue: %w", err)
	}

	msgs, err := c.consumeCh.Consume(
		queue.Name,
		c.consumerTag,
		false, false, false, false, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}

	return msgs, nil
}

func (c *Consumer) HandleCoursesExchange(ctx context.Context) error {
	msgs, err := c.setupQueue()
	if err != nil {
		return err
	}

	defer func() {
		if c.consumeCh != nil && !c.consumeCh.IsClosed() {
			c.consumeCh.Close()
		}
		c.consumeCh = nil
	}()

	log.Info("consumer started, waiting for messages")

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case msg, ok := <-msgs:
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
