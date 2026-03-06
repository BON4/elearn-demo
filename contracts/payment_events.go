package contracts

import (
	"errors"

	"github.com/google/uuid"
)

const (
	PaymentsExchange = "domain.events"
)

const (
	PaymentSucceededEventSchemaVersion = 1
	PaymentRefoundedEventSchemaVersion = 1
)

var ErrInvalidPayload = errors.New("invalid payload")

type PaymentSucceededEventPayload struct {
	PaymentID uuid.UUID
	UserID    uuid.UUID
	CourseID  uuid.UUID
	Amount    int64
	Currency  string
	Version   int64
}

func (p *PaymentSucceededEventPayload) Validate() error {
	if p.PaymentID == uuid.Nil ||
		p.UserID == uuid.Nil ||
		p.CourseID == uuid.Nil {
		return ErrInvalidPayload
	}

	return nil
}

type PaymentRefoundedEventPayload struct {
	PaymentID uuid.UUID
	UserID    uuid.UUID
	CourseID  uuid.UUID
	Version   int64
}

func (p *PaymentRefoundedEventPayload) Validate() error {
	if p.PaymentID == uuid.Nil ||
		p.UserID == uuid.Nil ||
		p.CourseID == uuid.Nil {
		return ErrInvalidPayload
	}

	return nil
}
