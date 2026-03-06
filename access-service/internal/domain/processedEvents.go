package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	CoursePublishedProcessedEventType  = "course-published"
	CourseDraftededProcessedEventType  = "course-drafted"
	PaymentSucceededProcessedEventType = "payment-succeeded"
	PaymentRefoundedProcessedEventType = "payment-refounded"
)

var (
	ErrEventAlreadyProcessed = errors.New("event already have been processed, idempotency violation")
)

type ProcessedEvent struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Type        string
	ProcessedAt time.Time
}

func NewProcessedEvent(id uuid.UUID, eventType string) *ProcessedEvent {
	return &ProcessedEvent{
		ID:          id,
		Type:        eventType,
		ProcessedAt: time.Now(),
	}
}
