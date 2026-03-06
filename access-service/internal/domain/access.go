package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AccessStatus string

const (
	AccessGranted AccessStatus = "granted"
	AccessRevoked AccessStatus = "revoked"
)

type AccessType string

const (
	AccessPayment AccessType = "payment"
	AccessSystem  AccessType = "system"
)

var (
	ErrInvalidAccessStatus = errors.New("invalid access status")
)

type UserCourseAccess struct {
	UserID       uuid.UUID    `gorm:"type:uuid;primaryKey"`
	CourseID     uuid.UUID    `gorm:"type:uuid;primaryKey"`
	AccessStatus AccessStatus `gorm:"type:varchar(20)"`
	AccessType   AccessType   `gorm:"type:varchar(20)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUserCourseAccess(userID, courseID uuid.UUID, accType AccessType) *UserCourseAccess {
	now := time.Now()
	return &UserCourseAccess{
		UserID:       userID,
		CourseID:     courseID,
		AccessStatus: AccessGranted,
		AccessType:   accType,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (a *UserCourseAccess) Validate() error {
	switch a.AccessStatus {
	case AccessGranted, AccessRevoked:
	default:
		return ErrInvalidAccessStatus
	}
	return nil
}

func (a *UserCourseAccess) Revoke() {
	a.AccessStatus = AccessRevoked
	a.UpdatedAt = time.Now()
}

func (a *UserCourseAccess) Grant() {
	a.AccessStatus = AccessGranted
	a.UpdatedAt = time.Now()
}
