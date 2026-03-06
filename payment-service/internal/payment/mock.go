package payment

import (
	"context"
	"time"

	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	log "github.com/sirupsen/logrus"
)

type MockPaymentProvider struct{}

func NewMockPaymentProvider() *MockPaymentProvider {
	return &MockPaymentProvider{}
}

func (m *MockPaymentProvider) MakePayment(ctx context.Context, payment *domain.Payment) error {
	time.Sleep(time.Second)

	log.WithFields(log.Fields{
		"payment_id": payment.ID,
		"user_id":    payment.UserID,
		"course_id":  payment.CourseID,
		"status":     payment.Status,
		"amount":     payment.Amount,
		"provider":   payment.Provider,
	}).Info("payment creation requested")

	return nil
}
func (m *MockPaymentProvider) RefoundPayment(ctx context.Context, payment *domain.Payment) error {
	time.Sleep(time.Second)

	log.WithFields(log.Fields{
		"payment_id": payment.ID,
		"user_id":    payment.UserID,
		"course_id":  payment.CourseID,
		"status":     payment.Status,
		"amount":     payment.Amount,
		"provider":   payment.Provider,
	}).Info("payment refound requested")

	return nil
}
