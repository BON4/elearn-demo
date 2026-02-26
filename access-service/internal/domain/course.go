package domain

import (
	"errors"
	"time"

	"github.com/BON4/elearn-demo/contracts"
	"github.com/google/uuid"
)

type CourseRM struct {
	ID          uuid.UUID
	Status      CourseStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt time.Time
}

type CourseStatus string

const (
	Draft     CourseStatus = "draft"
	Published CourseStatus = "published"
)

var (
	ErrInvalidStatus = errors.New("invalid course status")
)

func (c *CourseRM) Validate() error {
	switch c.Status {
	case Draft, Published:
	default:
		return ErrInvalidStatus
	}

	return nil
}

func CourceFromContract(ctr contracts.CoursePublishedEventPayload) *CourseRM {
	now := time.Now()
	return &CourseRM{
		ID:          ctr.CourseID,
		Status:      Published,
		CreatedAt:   now,
		UpdatedAt:   now,
		PublishedAt: ctr.PublishedAt,
	}
}
