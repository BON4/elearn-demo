package domain

import (
	"errors"
	"time"

	"github.com/BON4/elearn-demo/contracts"
	"github.com/google/uuid"
)

type CourseRM struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Status      CourseStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt time.Time
	DraftedAt   time.Time
	Version     int64
}

type CourseStatus string

const (
	Drafted   CourseStatus = "drafted"
	Published CourseStatus = "published"
)

var (
	ErrInvalidStatus = errors.New("invalid course status")
)

func (c *CourseRM) Validate() error {
	switch c.Status {
	case Drafted, Published:
	default:
		return ErrInvalidStatus
	}

	return nil
}

func PublishedCourceFromContract(ctr contracts.CoursePublishedEventPayload) *CourseRM {
	now := time.Now()
	return &CourseRM{
		ID:          ctr.CourseID,
		Status:      Published,
		CreatedAt:   now,
		UpdatedAt:   now,
		PublishedAt: ctr.PublishedAt,
		Version:     ctr.Version,
	}
}

func DraftedCourceFromContract(ctr contracts.CourseDraftedEventPayload) *CourseRM {
	now := time.Now()
	return &CourseRM{
		ID:        ctr.CourseID,
		Status:    Drafted,
		CreatedAt: now,
		UpdatedAt: now,
		DraftedAt: ctr.DraftededAt,
		Version:   ctr.Version,
	}
}
