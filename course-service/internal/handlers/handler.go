package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	courseService CourseService
	logger        *logrus.Logger
	producer      Producer
}

func NewHandler(coursesService CourseService, producer Producer, logger *logrus.Logger) *Handler {
	return &Handler{
		courseService: coursesService,
		producer:      producer,
		logger:        logger,
	}
}

func (h *Handler) Register(r *gin.Engine) {
	group := r.Group("/courses")
	{
		group.POST("", h.CreateCourse)
		group.GET("/:id", h.GetCourse)
		group.POST("/:id/publish", h.PublishCourse)
		group.POST("/:id/draft", h.DraftCourse)
	}

	group = r.Group("/tests")
	{
		group.POST("/pause-worker", h.TestPauseWorker)
		group.POST("/resume-worker", h.TestResumeWorker)
	}
}
