package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Consumer interface {
	PauseWorker()
	ResumeWorker()
}

// POST /tests/pause-worker
func (h *Handler) TestPauseWorker(c *gin.Context) {
	h.consumer.PauseWorker()
	c.JSON(http.StatusOK, gin.H{"status": "paused"})
}

// POST /tests/resume-worker
func (h *Handler) TestResumeWorker(c *gin.Context) {
	h.consumer.ResumeWorker()
	c.JSON(http.StatusOK, gin.H{"status": "resumed"})
}
