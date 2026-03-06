package service

import (
	"context"
	"fmt"

	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	"github.com/BON4/elearn-demo/payment-service/internal/repo"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (c *PaymentsService) SavePayment(ctx context.Context, payment *domain.Payment) error {
	payment.IncVersion()
	err := payment.Validate()
	if err != nil {
		return err
	}

	oldPayment, err := c.GetPayment(ctx, payment.ID)
	if err != nil {
		return err
	}

	err = payment.ValidateForUpdate(oldPayment)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"payment_id": payment.ID,
		"user_id":    payment.UserID,
		"course_id":  payment.CourseID,
		"status":     payment.Status,
		"amount":     payment.Amount,
		"provider":   payment.Provider,
	}).Info("payment saved")

	return c.db.Save(payment).Error
}

func (c *PaymentsService) CreatePayment(ctx context.Context, userID, courseID uuid.UUID, amount int64, currency string, provider string) (*domain.Payment, error) {
	course, err := c.courseService.GetCourse(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("failed to get course: %w", err)
	}

	err = course.Purchese()
	if err != nil {
		return nil, err
	}

	var domainPayment = domain.NewPayment(
		userID,
		courseID,
		amount,
		currency,
		provider,
	)

	err = domainPayment.Validate()
	if err != nil {
		return nil, err
	}

	err = c.db.Create(domainPayment).Error
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"payment_id": domainPayment.ID,
		"user_id":    domainPayment.UserID,
		"course_id":  domainPayment.CourseID,
		"status":     domainPayment.Status,
		"amount":     domainPayment.Amount,
		"provider":   domainPayment.Provider,
	}).Info("payment created")

	return domainPayment, nil
}

func (c *PaymentsService) MarkProcessedPayment(ctx context.Context, paymentID uuid.UUID) error {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	txC := c.withRepo(repo.NewMonoRepo(tx))

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var payment domain.Payment

	err := tx.Clauses(clause.Locking{
		Strength: "UPDATE",
	}).Where("id = ?", paymentID).First(&payment).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = payment.MarkProcessing()
	if err != nil {
		tx.Rollback()
		return err
	}

	err = txC.SavePayment(ctx, &payment)
	if err != nil {
		tx.Rollback()
		return err
	}

	log.WithFields(log.Fields{
		"payment_id": payment.ID,
		"user_id":    payment.UserID,
		"course_id":  payment.CourseID,
		"status":     payment.Status,
		"amount":     payment.Amount,
		"provider":   payment.Provider,
	}).Info("payment marked as proccessed")

	err = tx.Commit().Error
	if err != nil {
		return err
	}

	err = c.paymentProvider.MakePayment(ctx, &payment)
	if err != nil {
		return err
	}

	return nil
}

func (c *PaymentsService) MarkSuccesedPayment(ctx context.Context, paymentID uuid.UUID) error {
	payment, err := c.GetPayment(ctx, paymentID)
	if err != nil {
		return err
	}

	return c.db.Transaction(func(tx *gorm.DB) error {
		rp := c.db.WithTx(ctx, tx)
		txC := c.withRepo(rp)

		err = payment.MarkSucceeded(paymentID.String())
		if err != nil {
			return err
		}

		err = txC.SavePayment(ctx, payment)
		if err != nil {
			return err
		}

		event := domain.NewPaymentSucceededEvent(
			paymentID,
			payment.UserID,
			payment.CourseID,
			payment.Amount,
			payment.Currency,
			payment.Version,
		)

		err = event.Validate()
		if err != nil {
			return err
		}

		err = c.eventProducer.CreatePaymentEvent(ctx, event)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"payment_id": payment.ID,
			"user_id":    payment.UserID,
			"course_id":  payment.CourseID,
			"status":     payment.Status,
			"amount":     payment.Amount,
			"provider":   payment.Provider,
		}).Info("payment marked as succesed")

		return nil
	})
}

func (c *PaymentsService) MarkRefoundedPayment(ctx context.Context, paymentID uuid.UUID) error {
	payment, err := c.GetPayment(ctx, paymentID)
	if err != nil {
		return err
	}

	return c.db.Transaction(func(tx *gorm.DB) error {
		rp := c.db.WithTx(ctx, tx)
		txC := c.withRepo(rp)

		err = payment.Refund()
		if err != nil {
			return err
		}

		err = txC.SavePayment(ctx, payment)
		if err != nil {
			return err
		}

		event := domain.NewPaymentRefoundedEvent(
			paymentID,
			payment.UserID,
			payment.CourseID,
			payment.Version,
		)

		err = event.Validate()
		if err != nil {
			return err
		}

		err = c.eventProducer.CreatePaymentEvent(ctx, event)
		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"payment_id": payment.ID,
			"user_id":    payment.UserID,
			"course_id":  payment.CourseID,
			"status":     payment.Status,
			"amount":     payment.Amount,
			"provider":   payment.Provider,
		}).Info("payment marked as refounded")

		return nil
	})
}

func (p *PaymentsService) MakePaymentRefoundRequest(ctx context.Context, paymentID uuid.UUID) error {
	payment, err := p.GetPayment(ctx, paymentID)
	if err != nil {
		return err
	}

	err = p.paymentProvider.RefoundPayment(ctx, payment)
	if err != nil {
		return err
	}

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
