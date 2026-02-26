package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/BON4/elearn-demo/access-service/internal/domain"
	"github.com/BON4/elearn-demo/access-service/internal/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CoursesService struct {
	db *repo.MonoRepo
}

func NewCoursesService(rp *repo.MonoRepo) *CoursesService {
	return &CoursesService{
		db: rp,
	}
}

var (
	ErrEventAlreadyProcessed = errors.New("event already have been processed, idempotency violation")
)

func (c *CoursesService) ProcessPublishedCourseEvent(ctx context.Context, course domain.CourseRM, eventID uuid.UUID) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		rp := c.db.WithTx(tx)
		txC := NewCoursesService(rp)

		evt := domain.NewProcessedEvent(eventID, domain.CoursePublishedProcessedEventType)
		err := txC.db.Model(domain.ProcessedEvent{}).Create(evt).Error
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrEventAlreadyProcessed
			}
			return fmt.Errorf("failed to save processed event: %w", err)
		}

		err = txC.db.Model(domain.CourseRM{}).Save(course).Error
		if err != nil {
			return fmt.Errorf("failed to save cource: %w", err)
		}

		return nil
	})
}

func (c *CoursesService) ProcessedEventExists(eventID uuid.UUID) (bool, error) {
	var count int64
	err := c.db.Where("id = ?", eventID).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
