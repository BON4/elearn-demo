package handlers

import (
	"context"
	"net/http"

	"github.com/BON4/elearn-demo/course-service/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CourseService interface {
	CreateCourse(ctx context.Context, title string, description string, authorID uuid.UUID) (*domain.Course, error)
	GetCourse(ctx context.Context, courseID uuid.UUID) (*domain.Course, error)
	PublishCourse(ctx context.Context, courseID uuid.UUID) error
}

type CreateCourseRequest struct {
	Title       string    `json:"title" binding:"required,min=3"`
	Description string    `json:"description" binding:"required,min=10"`
	AuthorID    uuid.UUID `json:"author_id" binding:"required"`
}

type CreateLessonRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (h *Handler) PublishCourse(c *gin.Context) {
	courseUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not valid"})
		return
	}

	err = h.courseService.PublishCourse(c.Request.Context(), courseUUID)
	if err != nil {
		h.logger.
			WithError(err).
			WithField("course_id", courseUUID).
			Error("failed to publish course")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish course"})
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *Handler) GetCourse(c *gin.Context) {
	courseID := c.Param("id")
	courseUUID, err := uuid.Parse(courseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is not valid"})
		return
	}

	course, err := h.courseService.GetCourse(c.Request.Context(), courseUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get course"})
		return
	}

	if course == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "course not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          course.ID,
		"title":       course.Title,
		"description": course.Description,
		"author_id":   course.AuthorID,
		"status":      course.Status,
	})
}

func (h *Handler) CreateCourse(c *gin.Context) {
	var req CreateCourseRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	course, err := h.courseService.CreateCourse(
		c.Request.Context(),
		req.Title,
		req.Description,
		req.AuthorID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to create course",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":          course.ID,
		"title":       course.Title,
		"description": course.Description,
		"author_id":   course.AuthorID,
	})
}
