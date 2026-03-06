package repo

import (
	"context"

	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	"gorm.io/gorm"
)

type MonoRepo struct {
	*gorm.DB
}

func NewMonoRepo(db *gorm.DB) *MonoRepo {
	return &MonoRepo{
		db,
	}
}

func (c MonoRepo) WithTx(ctx context.Context, tx *gorm.DB) *MonoRepo {
	return &MonoRepo{
		tx.WithContext(ctx),
	}
}

func (m *MonoRepo) MigrateDomain() error {
	err := m.AutoMigrate(&domain.Payment{})
	if err != nil {
		return err
	}

	err = m.AutoMigrate(&domain.PaymentEvent{})
	if err != nil {
		return err
	}
	return nil

}
