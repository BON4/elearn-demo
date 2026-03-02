package service

import (
	"errors"
	"fmt"

	"github.com/BON4/elearn-demo/access-service/internal/domain"
	"github.com/BON4/elearn-demo/access-service/internal/repo"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AccessService struct {
	db *repo.MonoRepo
}

func NewAccessService(rp *repo.MonoRepo) *AccessService {
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
