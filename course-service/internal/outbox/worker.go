package outbox

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/BON4/elearn-demo/contracts"
	"github.com/BON4/elearn-demo/course-service/internal/domain"
	"github.com/BON4/elearn-demo/course-service/internal/infra"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

type OutboxService interface {
	ProcessBatch(ctx context.Context, limit int, fn func([]*domain.OutboxEvent) error) error
}

type Worker struct {
	outboxSrv OutboxService
	broker    *amqp.Connection
	interval  time.Duration
	publishCh *amqp.Channel
	initErr   error
	paused    *atomic.Bool
}

func NewOutboxWorker(
	outboxSrv OutboxService,
	brokker *amqp.Connection,
	interval time.Duration,
) *Worker {
	paused := atomic.Bool{}
	paused.Store(false)
	return &Worker{
		outboxSrv: outboxSrv,
		broker:    brokker,
		interval:  interval,
		paused:    &paused,
	}
}

func (w *Worker) Pause() {
	w.paused.Store(true)
	log.Info("outbox worker paused")
}

func (w *Worker) Resume() {
	w.paused.Store(false)
	log.Info("outbox worker resumed")
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()
	defer func() {
		if w.publishCh != nil {
			w.publishCh.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if w.paused.Load() {
				continue
			}

			err := w.processBatch(ctx)
			if err != nil {
				log.WithError(err).Println("failed to process batch")
			}
		}
	}
}

func (w *Worker) setupExchange() error {
	if w.publishCh != nil && !w.publishCh.IsClosed() {
		return nil
	}

	ch, err := w.broker.Channel()
	if err != nil {
		return fmt.Errorf("open channel: %w", err)
	}

	if err := ch.ExchangeDeclare(
		contracts.CoursesExchange,
		"topic",
		true, false, false, false, nil,
	); err != nil {
		ch.Close()
		return fmt.Errorf("declare exchange: %w", err)
	}

	w.publishCh = ch
	return nil
}

func (w *Worker) processBatch(ctx context.Context) error {
	if err := w.setupExchange(); err != nil {
		return err
	}

	err := w.outboxSrv.ProcessBatch(ctx, 100, func(events []*domain.OutboxEvent) error {
		for _, evt := range events {
			if err := w.publishEvent(ctx, evt); err != nil {
				evt.MarkFailedWithRetry(err)
				log.WithError(err).
					WithField("outbox_event_uuid", evt.ID).
					Error("failed to pubish event, retrying")
			} else {
				evt.MarkProcessed()
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to process outbox events: %w", err)
	}

	return nil
}

func (w *Worker) publishEvent(ctx context.Context, evt *domain.OutboxEvent) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	routingKey, err := infra.CourseEventTypeToRoutingKey(evt.Type)
	if err != nil {
		return err
	}

	err = w.publishCh.PublishWithContext(ctx, contracts.CoursesExchange, routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        evt.Payload,
		MessageId:   evt.ID.String(),
		Headers: amqp.Table{
			"schema_version": evt.SchemaVersion,
		},
	})
	if err != nil {
		w.publishCh = nil // force channel reopen
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.
		WithField("outbox_event_uuid", evt.ID).
		WithField("outbox_event_exchange", contracts.CoursesExchange).
		WithField("outbox_event_exchange_key", routingKey).
		Info("event was published")

	return nil
}
