package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Course struct {
	ID          uuid.UUID
	Title       string
	Description string
	AuthorID    uuid.UUID
	Status      CourseStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CourseStatus string

const (
	Draft     CourseStatus = "draft"
	Published CourseStatus = "published"
)

var (
	ErrInvalidTitle       = errors.New("invalid title")
	ErrInvalidDescription = errors.New("invalid description")
	ErrInvalidAuthorID    = errors.New("invalid author id")
	ErrInvalidStatus      = errors.New("invalid course status")
	ErrAlreadyPublished   = errors.New("course is already published")
)

func (c *Course) Validate() error {
	if c.Title == "" {
		return ErrInvalidTitle
	}
	if len(c.Title) > 200 {
		return ErrInvalidTitle
	}

	if c.Description == "" || len(c.Description) < 10 {
		return ErrInvalidDescription
	}

	if c.AuthorID == uuid.Nil {
		return ErrInvalidAuthorID
	}

	switch c.Status {
	case Draft, Published:
	default:
		return ErrInvalidStatus
	}

	return nil
}

func (c *Course) Publish() error {
	if c.Status == Published {
		return ErrAlreadyPublished
	}

	if err := c.Validate(); err != nil {
		return err
	}

	c.Status = Published
	c.UpdatedAt = time.Now()

	return nil
}
