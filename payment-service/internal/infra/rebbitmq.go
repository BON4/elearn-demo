package infra

import (
	"context"
	"fmt"

	"github.com/BON4/elearn-demo/contracts"
	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

func PaymentEventTypeToRoutingKey(evtT domain.EventType) (string, error) {
	switch evtT {
	case domain.PaymentSucceededEventType:
		return contracts.PaymentSucceededRoutingKey, nil
	case domain.PaymentRefundedEventType:
		return contracts.PaymentRefoundedRoutingKey, nil
	}
	return "", fmt.Errorf("uknown routing key for event type: %s", evtT)
}

func NewRabbitMQ(ctx context.Context, url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("rabbitmq: dial: %w", err)
	}

	go func() {
		<-ctx.Done()
		conn.Close()
	}()

	return conn, nil
}
