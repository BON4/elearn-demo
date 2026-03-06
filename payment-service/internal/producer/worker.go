package outbox

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/BON4/elearn-demo/contracts"
	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	"github.com/BON4/elearn-demo/payment-service/internal/infra"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

type EventService interface {
	ProcessPaymentEventBatch(ctx context.Context, limit int, fn func([]*domain.PaymentEvent) error) error
}

type Worker struct {
	eventsSrv EventService
	broker    *amqp.Connection
	interval  time.Duration
	publishCh *amqp.Channel
	initErr   error
	paused    *atomic.Bool
}

func NewProducerWorker(
	eventsSrv EventService,
	brokker *amqp.Connection,
	interval time.Duration,
) *Worker {
	paused := atomic.Bool{}
	paused.Store(false)
	return &Worker{
		eventsSrv: eventsSrv,
		broker:    brokker,
		interval:  interval,
		paused:    &paused,
	}
}

func (w *Worker) PauseWorker() {
	w.paused.Store(true)
	log.Info("producer worker paused")
}

func (w *Worker) ResumeWorker() {
	w.paused.Store(false)
	log.Info("producer worker resumed")
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
		contracts.PaymentsExchange,
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

	err := w.eventsSrv.ProcessPaymentEventBatch(ctx, 100, func(events []*domain.PaymentEvent) error {
		for _, evt := range events {
			if err := w.publishEvent(ctx, evt); err != nil {
				evt.MarkFailedWithRetry(err)
				log.WithError(err).
					WithField("payment_event_uuid", evt.ID).
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

func (w *Worker) publishEvent(ctx context.Context, evt *domain.PaymentEvent) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	routingKey, err := infra.PaymentEventTypeToRoutingKey(evt.Type)
	if err != nil {
		return err
	}

	err = w.publishCh.PublishWithContext(ctx, contracts.PaymentsExchange, routingKey, false, false, amqp.Publishing{
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
		WithField("payment_event_uuid", evt.ID).
		WithField("payment_event_exchange", contracts.PaymentsExchange).
		WithField("payment_event_exchange_key", routingKey).
		Info("event was published")

	return nil
}
