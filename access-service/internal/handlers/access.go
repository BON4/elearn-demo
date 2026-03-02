package handlers

import (
	"net/http"
	"time"

	"github.com/BON4/elearn-demo/access-service/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AccessService interface {
	GetUserAccessList(userID uuid.UUID) ([]*domain.UserCourseAccess, error)
	GetUserAccess(userID, courseID uuid.UUID) (*domain.UserCourseAccess, error)
}

type AccessDTO struct {
	UserID       uuid.UUID
	CourseID     uuid.UUID
	AccessStatus string
	CreatedAt    time.Time
}

func NewAccessDTO(d *domain.UserCourseAccess) *AccessDTO {
	a := AccessDTO{}
	a.UserID = d.UserID
	a.CourseID = d.CourseID
	a.AccessStatus = string(d.AccessStatus)
	a.CreatedAt = d.CreatedAt
	return &a
}

// GET /access/:user_id/all
func (h *Handler) GetUserAccessList(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	list, err := h.accessService.GetUserAccessList(userID)
	if err != nil {
		h.logger.WithError(err).Error("failed to get access list")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	dtos := make([]*AccessDTO, 0, len(list))
	for _, l := range list {
		dtos = append(dtos, NewAccessDTO(l))
	}
	c.JSON(http.StatusOK, gin.H{"data": dtos})
}

type AccessStatusDTO struct {
	AccessStatus string
}

func NewAccessStatusDTO(s domain.AccessStatus) *AccessStatusDTO {
	return &AccessStatusDTO{
		AccessStatus: string(s),
	}
}

// GET /access/:user_id/:course_id
func (h *Handler) GetUserAccess(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	courseID, err := uuid.Parse(c.Param("course_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid course_id"})
		return
	}

	access, err := h.accessService.GetUserAccess(userID, courseID)
	if err != nil {
		h.logger.WithError(err).Errorf("failed to get access user=%s course=%s", userID, courseID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	if access == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "access record not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": NewAccessStatusDTO(access.AccessStatus)})
}
