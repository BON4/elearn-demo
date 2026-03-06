package domain

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/BON4/elearn-demo/contracts"
	"github.com/google/uuid"
)

type EventStatus string

const (
	Pending   EventStatus = "pending"
	Processed EventStatus = "processed"
	Failed    EventStatus = "failed"
)

type EventType string

const (
	PaymentSucceededEventType EventType = "SUCCEEDED"
	PaymentRefundedEventType  EventType = "REFUNDED"
)

const (
	MAX_FAILED_EVENT_RETRY_COUNT = 200
)

var (
	ErrInvalidEventID       = errors.New("invalid event id")
	ErrInvalidEventType     = errors.New("invalid event type")
	ErrInvalidEventStatus   = errors.New("invalid event status")
	ErrInvalidPayload       = errors.New("invalid payload")
	ErrInvalidSchemaVersion = errors.New("invalid schema version")
)

type PaymentEvent struct {
	ID            uuid.UUID
	Type          EventType
	Status        EventStatus
	Payload       json.RawMessage
	SchemaVersion int
	CreatedAt     time.Time
	RetryCount    int64
	LastError     string
}

func (e *PaymentEvent) SetPayload(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	e.Payload = json.RawMessage(data)
	return nil
}

func (e *PaymentEvent) MarkProcessed() {
	e.Status = Processed
}

func (e *PaymentEvent) MarkFailed(err error) {
	e.Status = Failed
	e.LastError = err.Error()
}

func (e *PaymentEvent) MarkFailedWithRetry(err error) {
	e.RetryCount++
	e.LastError = err.Error()
	if e.RetryCount > MAX_FAILED_EVENT_RETRY_COUNT {
		e.Status = Failed
	}
}

func (e *PaymentEvent) Validate() error {
	if e.ID == uuid.Nil {
		return ErrInvalidEventID
	}

	if e.Payload == nil {
		return ErrInvalidPayload
	}

	if e.CreatedAt.IsZero() {
		return errors.New("created_at is required")
	}

	switch e.Status {
	case Pending, Processed, Failed:
	default:
		return ErrInvalidEventStatus
	}

	switch e.Type {

	case PaymentSucceededEventType:

		if e.SchemaVersion != contracts.PaymentSucceededEventSchemaVersion {
			return ErrInvalidSchemaVersion
		}

		var payload contracts.PaymentSucceededEventPayload
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			return ErrInvalidPayload
		}

		if err := payload.Validate(); err != nil {
			return err
		}

	case PaymentRefundedEventType:

		if e.SchemaVersion != contracts.PaymentRefoundedEventSchemaVersion {
			return ErrInvalidSchemaVersion
		}

		var payload contracts.PaymentRefoundedEventPayload
		if err := json.Unmarshal(e.Payload, &payload); err != nil {
			return ErrInvalidPayload
		}

		if err := payload.Validate(); err != nil {
			return err
		}

	default:
		return ErrInvalidEventType
	}

	return nil
}

func NewPaymentSucceededEvent(
	paymentID uuid.UUID,
	userID uuid.UUID,
	courseID uuid.UUID,
	amount int64,
	currency string,
	paymentVersion PaymentVersion,
) *PaymentEvent {
	now := time.Now()

	e := &PaymentEvent{
		ID:            uuid.New(),
		Type:          PaymentSucceededEventType,
		Status:        Pending,
		SchemaVersion: contracts.PaymentSucceededEventSchemaVersion,
		CreatedAt:     now,
	}
	e.SetPayload(contracts.PaymentSucceededEventPayload{
		PaymentID: paymentID,
		UserID:    userID,
		CourseID:  courseID,
		Amount:    amount,
		Currency:  currency,
		Version:   int64(paymentVersion),
	})

	return e
}

func NewPaymentRefoundedEvent(
	paymentID uuid.UUID,
	userID uuid.UUID,
	courseID uuid.UUID,
	paymentVersion PaymentVersion,
) *PaymentEvent {
	now := time.Now()

	e := &PaymentEvent{
		ID:            uuid.New(),
		Type:          PaymentRefundedEventType,
		Status:        Pending,
		SchemaVersion: contracts.PaymentRefoundedEventSchemaVersion,
		CreatedAt:     now,
	}
	e.SetPayload(contracts.PaymentRefoundedEventPayload{
		PaymentID: paymentID,
		UserID:    userID,
		CourseID:  courseID,
		Version:   int64(paymentVersion),
	})

	return e
}
