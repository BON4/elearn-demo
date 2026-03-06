package service

import (
	"context"

	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	"github.com/google/uuid"
)

func (c *PaymentsService) GetPayment(ctx context.Context, paymentID uuid.UUID) (*domain.Payment, error) {
	var domainPayment domain.Payment
	err := c.db.Where("id = ?", paymentID).First(&domainPayment).Error
	if err != nil {
		return nil, err
	}

	return &domainPayment, nil
}

func (c *PaymentsService) GetUserPayments(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.Payment, error) {
	var domainPayments []*domain.Payment
	err := c.db.Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&domainPayments).Error
	if err != nil {
		return nil, err
	}

	return domainPayments, nil
}
