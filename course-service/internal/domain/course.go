package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Course struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Title       string
	Description string
	AuthorID    uuid.UUID
	Status      CourseStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int64
}

type CourseStatus string

const (
	Draft     CourseStatus = "drafted"
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

func (c *Course) Draft() error {
	if err := c.Validate(); err != nil {
		return err
	}

	c.Status = Draft
	c.UpdatedAt = time.Now()

	return nil
}

func (c *Course) IncVersion() {
	c.Version++
}
