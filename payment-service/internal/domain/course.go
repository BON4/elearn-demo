package domain

import (
	"errors"

	"github.com/google/uuid"
)

type CourseRM struct {
	ID     uuid.UUID
	Status string
}

var (
	ErrCourseCanNotBePurchesed = errors.New("this course can not be purchesed")
)

func (c *CourseRM) Purchese() error {
	if c.Status != "published" {
		return ErrCourseCanNotBePurchesed
	}

	return nil
}
