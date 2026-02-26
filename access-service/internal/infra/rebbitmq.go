package infra

import (
	"context"
	"fmt"

	"github.com/BON4/elearn-demo/course-service/internal/domain"
	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	CoursePublishedRoutingKey   = "course.published"
	CourseUnPublishedRoutingKey = "course.unpublished"
)

func CourseEventTypeToRoutingKey(evtT string) (string, error) {
	switch evtT {
	case domain.CoursePublishedEventType:
		return CoursePublishedRoutingKey, nil
	case domain.CourseUnPublishedEventType:
		return CourseUnPublishedRoutingKey, nil
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
