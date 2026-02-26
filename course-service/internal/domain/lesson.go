package domain

import "github.com/google/uuid"

type Lesson struct {
	ID       uuid.UUID
	CourseID uuid.UUID
	Title    string
	Position int
}
