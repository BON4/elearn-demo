package service

import (
	"context"
	"fmt"

	"github.com/BON4/elearn-demo/access-service/internal/domain"
	"github.com/BON4/elearn-demo/access-service/internal/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CoursesService struct {
	db *repo.MonoRepo
}

func NewCoursesService(rp *repo.MonoRepo) *CoursesService {
	return &CoursesService{
		db: rp,
	}
}

func (c *CoursesService) ProcessPublishedCourseEvent(ctx context.Context, course domain.CourseRM, eventID uuid.UUID) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		rp := c.db.WithTx(tx)
		txC := NewCoursesService(rp)

		evt := domain.NewProcessedEvent(eventID, domain.CoursePublishedProcessedEventType)
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

		var (
			existsCourse domain.CourseRM
		)
		resFind := txC.db.
			Where("id = ?", course.ID).
			Find(&existsCourse)
		if resFind.Error != nil {
			return res.Error
		}

		exists := resFind.RowsAffected > 0

		if exists && existsCourse.Version >= course.Version {
			return nil
		}

		err := txC.db.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.Assignments(map[string]any{
					"status":       course.Status,
					"version":      course.Version,
					"updated_at":   course.UpdatedAt,
					"published_at": course.PublishedAt,
				})}).
			Create(&course).
			Error
		if err != nil {
			return fmt.Errorf("failed to create cource: %w", err)
		}

		return nil
	})
}

func (c *CoursesService) ProcessDraftedCourseEvent(ctx context.Context, course domain.CourseRM, eventID uuid.UUID) error {
	return c.db.Transaction(func(tx *gorm.DB) error {
		rp := c.db.WithTx(tx)
		txC := NewCoursesService(rp)

		evt := domain.NewProcessedEvent(eventID, domain.CourseDraftededProcessedEventType)
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

		var (
			existsCourse domain.CourseRM
		)
		resFind := txC.db.
			Where("id = ?", course.ID).
			Find(&existsCourse)
		if resFind.Error != nil {
			return res.Error
		}

		exists := resFind.RowsAffected > 0

		if exists && existsCourse.Version >= course.Version {
			return nil
		}

		err := txC.db.
			Clauses(clause.OnConflict{
				Columns: []clause.Column{{Name: "id"}},
				DoUpdates: clause.Assignments(map[string]any{
					"status":     course.Status,
					"version":    course.Version,
					"updated_at": course.UpdatedAt,
				})}).
			Create(&course).
			Error
		if err != nil {
			return fmt.Errorf("failed to create cource: %w", err)
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
