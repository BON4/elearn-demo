package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	accessService AccessService
	logger        *logrus.Logger
	consumer      Consumer
}

func NewHandler(accessService AccessService, consumer Consumer, logger *logrus.Logger) *Handler {
	return &Handler{
		accessService: accessService,
		consumer:      consumer,
		logger:        logger,
	}
}

func (h *Handler) Register(r *gin.Engine) {
	group := r.Group("/access")
	{
		group.GET("/:user_id/:course_id", h.GetUserAccess)
		group.GET("/:user_id", h.GetUserAccessList)
	}

	group = r.Group("/tests")
	{
		group.POST("/pause-worker", h.TestPauseWorker)
		group.POST("/resume-worker", h.TestResumeWorker)
	}
}
