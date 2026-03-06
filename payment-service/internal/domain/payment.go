package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type PaymentStatus string

const (
	PaymentStatusCreated    PaymentStatus = "CREATED"
	PaymentStatusProcessing PaymentStatus = "PROCESSING"
	PaymentStatusSucceeded  PaymentStatus = "SUCCEEDED"
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusCancelled  PaymentStatus = "CANCELLED"
	PaymentStatusRefunded   PaymentStatus = "REFUNDED"
)

func (s PaymentStatus) IsFinal() bool {
	return s == PaymentStatusSucceeded ||
		s == PaymentStatusFailed ||
		s == PaymentStatusCancelled ||
		s == PaymentStatusRefunded
}

type PaymentVersion int64

type Payment struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID            uuid.UUID
	CourseID          uuid.UUID
	Amount            int64 // cents
	Currency          string
	Status            PaymentStatus
	Provider          string
	ProviderPaymentID string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Version           PaymentVersion
}

func NewPayment(
	userID uuid.UUID,
	courseID uuid.UUID,
	amount int64,
	currency string,
	provider string,
) *Payment {
	now := time.Now()

	return &Payment{
		ID:        uuid.New(),
		UserID:    userID,
		CourseID:  courseID,
		Amount:    amount,
		Currency:  currency,
		Status:    PaymentStatusCreated,
		Provider:  provider,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (p *Payment) MarkProcessing() error {
	if p.Status != PaymentStatusCreated && p.Status != PaymentStatusProcessing {
		return ErrInvalidStatusTransition
	}

	p.Status = PaymentStatusProcessing
	p.touch()
	return p.Validate()
}

func (p *Payment) MarkSucceeded(providerPaymentID string) error {
	if p.Status.IsFinal() {
		return ErrPaymentAlreadyFinal
	}

	p.Status = PaymentStatusSucceeded
	p.ProviderPaymentID = providerPaymentID
	p.touch()
	return p.Validate()
}

func (p *Payment) MarkFailed() error {
	if p.Status.IsFinal() {
		return ErrPaymentAlreadyFinal
	}

	p.Status = PaymentStatusFailed
	p.touch()
	return p.Validate()
}

func (p *Payment) Refund() error {
	if p.Status != PaymentStatusSucceeded {
		return ErrRefundNotAllowed
	}

	p.Status = PaymentStatusRefunded
	p.touch()
	return p.Validate()
}

func (p *Payment) IncVersion() {
	p.Version++
}

func (p *Payment) touch() {
	p.UpdatedAt = time.Now()
}

var (
	ErrInvalidPaymentID     = errors.New("invalid payment id")
	ErrInvalidUserID        = errors.New("invalid user id")
	ErrInvalidCourseID      = errors.New("invalid course id")
	ErrInvalidAmount        = errors.New("invalid payment amount")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrInvalidProvider      = errors.New("invalid provider")
	ErrInvalidPaymentStatus = errors.New("invalid payment status")
	ErrInvalidCreatedAt     = errors.New("invalid created_at")
	ErrInvalidUpdatedAt     = errors.New("invalid updated_at")
	ErrInvalidTimestamps    = errors.New("updated_at before created_at")
	ErrImmutableField       = errors.New("immutable field modification")
	ErrInvalidVersion       = errors.New("invalid version")
	ErrInvalidPayment       = errors.New("invalid payment")
)

func (p *Payment) ValidateForUpdate(old *Payment) error {
	if old == nil {
		return ErrInvalidPayment
	}

	if p.ID != old.ID {
		return ErrImmutableField
	}

	if p.UserID != old.UserID {
		return ErrImmutableField
	}

	if p.CourseID != old.CourseID {
		return ErrImmutableField
	}

	if p.Amount != old.Amount {
		return ErrImmutableField
	}

	if p.Currency != old.Currency {
		return ErrImmutableField
	}

	if p.Provider != old.Provider {
		return ErrImmutableField
	}

	if !old.Status.IsFinal() && old.Status != p.Status {
		// status change allowed through domain methods
	}

	if old.Status.IsFinal() && old.Status != p.Status {
		return ErrPaymentAlreadyFinal
	}

	if p.CreatedAt != old.CreatedAt {
		return ErrImmutableField
	}

	if p.Version <= old.Version {
		return ErrInvalidVersion
	}

	return nil
}

func (p *Payment) Validate() error {
	if p.ID == uuid.Nil {
		return ErrInvalidPaymentID
	}

	if p.UserID == uuid.Nil {
		return ErrInvalidUserID
	}

	if p.CourseID == uuid.Nil {
		return ErrInvalidCourseID
	}

	if p.Amount <= 0 {
		return ErrInvalidAmount
	}

	if p.Currency == "" {
		return ErrInvalidCurrency
	}

	if p.Provider == "" {
		return ErrInvalidProvider
	}

	if p.Status == "" {
		return ErrInvalidPaymentStatus
	}

	if p.CreatedAt.IsZero() {
		return ErrInvalidCreatedAt
	}

	if p.UpdatedAt.IsZero() {
		return ErrInvalidUpdatedAt
	}

	if p.UpdatedAt.Before(p.CreatedAt) {
		return ErrInvalidTimestamps
	}

	return nil
}
