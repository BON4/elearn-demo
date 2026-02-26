package repo

import (
	"github.com/BON4/elearn-demo/course-service/internal/domain"
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

func (c MonoRepo) WithTx(tx *gorm.DB) *MonoRepo {
	return &MonoRepo{
		tx,
	}
}

func (m *MonoRepo) MigrateDomain() error {
	err := m.AutoMigrate(&domain.Course{})
	if err != nil {
		return err
	}

	err = m.AutoMigrate(&domain.OutboxEvent{})
	if err != nil {
		return err
	}
	return nil

}
