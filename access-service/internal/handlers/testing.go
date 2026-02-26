package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type TestingService interface {
	PauseWorker()
	ResumeWorker()
}

func (h *Handler) TestPauseWorker(c *gin.Context) {
	h.testing.PauseWorker()
	c.Status(http.StatusOK)
}

func (h *Handler) TestResumeWorker(c *gin.Context) {
	h.testing.ResumeWorker()
	c.Status(http.StatusOK)
}
