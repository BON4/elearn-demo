package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/BON4/elearn-demo/contracts"
	"github.com/google/uuid"
)

type OutboxEventStatus string

const (
	Pending   OutboxEventStatus = "pending"
	Processed OutboxEventStatus = "processed"
	Failed    OutboxEventStatus = "failed"
)

const (
	MAX_FAILED_EVENT_RETRY_COUNT = 200
)

var (
	CoursePublishedEventType   = "CoursePublished"
	CourseUnPublishedEventType = "CourseUnPublished"
)

// Ошибки домена
var (
	ErrInvalidEventType   = errors.New("invalid outbox event type")
	ErrInvalidAggregateID = errors.New("invalid aggregate ID")
	ErrInvalidPayload     = errors.New("payload cannot be empty")
	ErrInvalidEventStatus = errors.New("invalid event status")
	ErrAlreadyProcessed   = errors.New("event already processed")
)

type OutboxEvent struct {
	ID            uuid.UUID         `json:"id"`
	AggregateID   uuid.UUID         `json:"aggregate_id"`
	Type          string            `json:"type"`
	Payload       json.RawMessage   `json:"payload" gorm:"type:jsonb"`
	Status        OutboxEventStatus `json:"status"`
	SchemaVersion int               `json:"version"`
	CreatedAt     time.Time         `json:"created_at"`
	RetryCount    int64             `json:"retry_count"`
	LastError     string            `json:"last_error"`
}

func CoursePublishedEvent(courseID uuid.UUID) *OutboxEvent {
	now := time.Now()
	e := &OutboxEvent{
		ID:            uuid.New(),
		AggregateID:   courseID,
		Type:          CoursePublishedEventType,
		Status:        Pending,
		SchemaVersion: contracts.CoursePublishedEventSchemaVersion,
		CreatedAt:     now,
	}
	e.SetPayload(contracts.CoursePublishedEventPayload{
		CourseID:    courseID,
		PublishedAt: now,
	})
	return e
}

func (e *OutboxEvent) Validate() error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	if e.AggregateID == uuid.Nil {
		return ErrInvalidAggregateID
	}
	if e.Type == "" {
		return ErrInvalidEventType
	}
	if len(e.Payload) == 0 {
		return ErrInvalidPayload
	}
	switch e.Status {
	case Pending, Processed, Failed:
	default:
		return ErrInvalidStatus
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	return nil
}

func (e *OutboxEvent) MarkProcessed() {
	e.Status = Processed
}

func (e *OutboxEvent) MarkFailed(err error) {
	e.Status = Failed
	e.LastError = err.Error()
}

func (e *OutboxEvent) MarkFailedWithRetry(err error) {
	e.RetryCount++
	e.LastError = err.Error()
	if e.RetryCount > MAX_FAILED_EVENT_RETRY_COUNT {
		e.Status = Failed
	}
}

func (e *OutboxEvent) SetPayload(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	e.Payload = json.RawMessage(data)
	return nil
}
