package contracts

import (
	"time"

	"github.com/google/uuid"
)

const (
	CoursesExchange = "domain.events"
)

const (
	CoursePublishedEventSchemaVersion = 1
	CourseDraftedEventSchemaVersion   = 1
)

type CoursePublishedEventPayload struct {
	CourseID    uuid.UUID
	PublishedAt time.Time
	Version     int64
}

type CourseDraftedEventPayload struct {
	CourseID    uuid.UUID
	DraftededAt time.Time
	Version     int64
}
