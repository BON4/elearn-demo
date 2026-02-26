package domain

import (
	"time"

	"github.com/google/uuid"
)

var (
	CoursePublishedProcessedEventType = "published"
)

type ProcessedEvent struct {
	ID          uuid.UUID `gorm:"primaryKey"`
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
