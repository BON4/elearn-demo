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
)

type CoursePublishedEventPayload struct {
	CourseID    uuid.UUID
	PublishedAt time.Time
}
