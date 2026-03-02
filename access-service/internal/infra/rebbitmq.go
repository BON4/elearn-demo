package infra

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

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
