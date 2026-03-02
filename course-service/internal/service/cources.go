package service

import (
	"context"

	"github.com/BON4/elearn-demo/course-service/internal/domain"
	"github.com/BON4/elearn-demo/course-service/internal/repo"
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

func (c *CoursesService) CreateCourse(ctx context.Context, title string, description string, authorID uuid.UUID) (*domain.Course, error) {
	var domainCource = &domain.Course{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		Status:      domain.Draft,
		AuthorID:    authorID,
		Version:     0,
	}

	err := domainCource.Validate()
	if err != nil {
		return nil, err
	}

	err = c.db.Create(domainCource).Error
	if err != nil {
		return nil, err
	}

	return domainCource, nil
}

func (c *CoursesService) GetCourse(ctx context.Context, courseID uuid.UUID) (*domain.Course, error) {
	var domainCourse domain.Course
	err := c.db.Where("id = ?", courseID).First(&domainCourse).Error
	if err != nil {
		return nil, err
	}

	return &domainCourse, nil
}

func (c *CoursesService) SaveCourse(ctx context.Context, course *domain.Course) error {
	course.IncVersion()
	err := course.Validate()
	if err != nil {
		return err
	}

	return c.db.Save(course).Error
}

func (c *CoursesService) CreateOutboxEvent(ctx context.Context, event *domain.OutboxEvent) error {
	err := event.Validate()
	if err != nil {
		return err
	}

	return c.db.Create(event).Error
}

func (c *CoursesService) PublishCourse(ctx context.Context, courseID uuid.UUID) error {
	course, err := c.GetCourse(ctx, courseID)
	if err != nil {
		return err
	}

	return c.db.Transaction(func(tx *gorm.DB) error {
		rp := c.db.WithTx(tx)
		txC := NewCoursesService(rp)

		err = course.Publish()
		if err != nil {
			return err
		}

		err = txC.SaveCourse(ctx, course)
		if err != nil {
			return err
		}
		event := domain.CoursePublishedEvent(courseID, course.Version)

		err = txC.CreateOutboxEvent(ctx, event)
		if err != nil {
			return err
		}

		return nil
	})
}

func (c *CoursesService) DraftCourse(ctx context.Context, courseID uuid.UUID) error {
	course, err := c.GetCourse(ctx, courseID)
	if err != nil {
		return err
	}

	return c.db.Transaction(func(tx *gorm.DB) error {
		rp := c.db.WithTx(tx)
		txC := NewCoursesService(rp)

		err = course.Draft()
		if err != nil {
			return err
		}

		err = txC.SaveCourse(ctx, course)
		if err != nil {
			return err
		}
		event := domain.CourseDraftedEvent(courseID, course.Version)

		err = txC.CreateOutboxEvent(ctx, event)
		if err != nil {
			return err
		}

		return nil
	})
}
