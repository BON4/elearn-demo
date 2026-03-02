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
	CoursePublishedEventType = "CoursePublished"
	CourseDraftedEventType   = "CourseDrafted"
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
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	AggregateID   uuid.UUID
	Type          string
	Payload       json.RawMessage
	Status        OutboxEventStatus
	SchemaVersion int
	CreatedAt     time.Time
	RetryCount    int64
	LastError     string
}

func CoursePublishedEvent(courseID uuid.UUID, courseVersion int64) *OutboxEvent {
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
		Version:     courseVersion,
	})
	return e
}

func CourseDraftedEvent(courseID uuid.UUID, courseVersion int64) *OutboxEvent {
	now := time.Now()
	e := &OutboxEvent{
		ID:            uuid.New(),
		AggregateID:   courseID,
		Type:          CourseDraftedEventType,
		Status:        Pending,
		SchemaVersion: contracts.CourseDraftedEventSchemaVersion,
		CreatedAt:     now,
	}
	e.SetPayload(contracts.CourseDraftedEventPayload{
		CourseID:    courseID,
		DraftededAt: now,
		Version:     courseVersion,
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
