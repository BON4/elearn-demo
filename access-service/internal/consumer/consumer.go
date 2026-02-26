package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/BON4/elearn-demo/access-service/internal/domain"
	"github.com/BON4/elearn-demo/access-service/internal/service"
	"github.com/BON4/elearn-demo/contracts"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type CourcesService interface {
	ProcessPublishedCourseEvent(ctx context.Context, course domain.CourseRM, eventID uuid.UUID) error
}

type Consumer struct {
	broker        *amqp.Connection
	consumeCh     *amqp.Channel
	courseService CourcesService
	queueName     string
	consumerTag   string
}

func NewConsumer(
	conn *amqp.Connection,
	queueName string,
	consumerTag string,
	service CourcesService,
) *Consumer {
	return &Consumer{
		broker:        conn,
		courseService: service,
		queueName:     queueName,
		consumerTag:   consumerTag,
	}
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		err := c.HandleCoursesExchange(ctx)
		if err == nil || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return
		}

		log.WithError(err).Warn("consumer error, restarting in 5s...")

		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (c *Consumer) setupQueue() (<-chan amqp.Delivery, error) {
	// Переоткрываем канал если упал
	if c.consumeCh == nil || c.consumeCh.IsClosed() {
		ch, err := c.broker.Channel()
		if err != nil {
			return nil, fmt.Errorf("open channel: %w", err)
		}
		c.consumeCh = ch
	}

	queue, err := c.consumeCh.QueueDeclare(
		c.queueName,
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

				if errors.Is(err, service.ErrEventAlreadyProcessed) {
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

	switch msg.RoutingKey {
	case contracts.CoursePublishedRoutingKey:
		return c.handleCoursePublishedMessage(ctx, msg)
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

	return c.courseService.ProcessPublishedCourseEvent(processCtx, *domain.CourceFromContract(coursePayload), eventID)
}
