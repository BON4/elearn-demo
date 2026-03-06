package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Producer interface {
	PauseWorker()
	ResumeWorker()
}

func (h *Handler) TestPauseWorker(c *gin.Context) {
	h.producer.PauseWorker()
	c.JSON(http.StatusOK, gin.H{"status": "paused"})
}

func (h *Handler) TestResumeWorker(c *gin.Context) {
	h.producer.ResumeWorker()
	c.JSON(http.StatusOK, gin.H{"status": "resumed"})
}
