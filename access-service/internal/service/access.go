package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/BON4/elearn-demo/access-service/internal/domain"
	"github.com/BON4/elearn-demo/access-service/internal/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AccessService struct {
	db *repo.MonoRepo
}

func NewAccessService(rp *repo.MonoRepo) *AccessService {
	return &AccessService{
		db: rp,
	}
}

func (p *AccessService) withRepo(rp *repo.MonoRepo) *AccessService {
	return &AccessService{
		db: rp,
	}
}

func (c *AccessService) GetUserAccessList(userID uuid.UUID) ([]*domain.UserCourseAccess, error) {
	accesses := make([]*domain.UserCourseAccess, 0)
	err := c.db.Model(&domain.UserCourseAccess{}).Where("user_id = ?", userID).Find(&accesses).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user accesses: %w", err)
	}

	return accesses, nil
}

func (c *AccessService) GetUserAccess(userID, courseID uuid.UUID) (*domain.UserCourseAccess, error) {
	var access domain.UserCourseAccess
	err := c.db.
		Model(&domain.UserCourseAccess{}).
		Where("user_id = ?", userID).
		Where("course_id = ?", courseID).
		First(&access).
		Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user access: %w", err)
	}

	return &access, nil
}

func (c *AccessService) UserHasAccess(userID, courseID uuid.UUID) (bool, error) {
	var count int64
	err := c.db.
		Model(&domain.UserCourseAccess{}).
		Where("user_id = ?", userID).
		Where("course_id = ?", courseID).
		Where("access_status = ?", domain.AccessGranted).
		Count(&count).
		Error
	if err != nil {
		return false, fmt.Errorf("failed to get user access: %w", err)
	}

	return count > 0, nil
}

func (c *AccessService) grantAccess(ctx context.Context, courseID uuid.UUID, userID uuid.UUID, accType domain.AccessType) error {
	access := domain.NewUserCourseAccess(userID, courseID, accType)
	access.Grant()

	err := c.db.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "course_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"access_status": access.AccessStatus,
				"updated_at":    access.UpdatedAt,
				"access_type":   access.AccessType,
			})}).
		Create(&access).
		Error
	if err != nil {
		return fmt.Errorf("failed to create cource: %w", err)
	}

	return nil
}

func (c *AccessService) revokeAccess(ctx context.Context, courseID uuid.UUID, userID uuid.UUID, accType domain.AccessType) error {
	access := domain.NewUserCourseAccess(userID, courseID, accType)
	access.Revoke()

	err := c.db.
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}, {Name: "course_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"access_status": access.AccessStatus,
				"updated_at":    access.UpdatedAt,
				"access_type":   access.AccessType,
			})}).
		Create(&access).
		Error
	if err != nil {
		return fmt.Errorf("failed to create cource: %w", err)
	}

	return nil
}

func (c *AccessService) ProcessPaymentSuccesedEvent(ctx context.Context, courseID uuid.UUID, userID uuid.UUID, eventID uuid.UUID) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		rp := c.db.WithTx(ctx, tx)
		txC := c.withRepo(rp)

		evt := domain.NewProcessedEvent(eventID, domain.PaymentSucceededProcessedEventType)
		res := txC.db.
			Model(domain.ProcessedEvent{}).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(evt)
		if res.Error != nil {
			return fmt.Errorf("failed to save processed event: %w", res.Error)
		}

		if res.RowsAffected == 0 {
			return domain.ErrEventAlreadyProcessed
		}

		err := txC.grantAccess(ctx, courseID, userID, domain.AccessPayment)
		if err != nil {
			return fmt.Errorf("failed to grant access: %w", err)
		}

		return nil
	})
}

func (c *AccessService) ProcessPaymentRefoundedEvent(ctx context.Context, courseID uuid.UUID, userID uuid.UUID, eventID uuid.UUID) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		rp := c.db.WithTx(ctx, tx)
		txC := c.withRepo(rp)

		evt := domain.NewProcessedEvent(eventID, domain.PaymentRefoundedProcessedEventType)
		res := txC.db.
			Model(domain.ProcessedEvent{}).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(evt)
		if res.Error != nil {
			return fmt.Errorf("failed to save processed event: %w", res.Error)
		}

		if res.RowsAffected == 0 {
			return domain.ErrEventAlreadyProcessed
		}

		err := txC.revokeAccess(ctx, courseID, userID, domain.AccessPayment)
		if err != nil {
			return fmt.Errorf("failed to grant access: %w", err)
		}

		return nil
	})
}
