package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	paymentsService PaymentsService
	eventsService   EventsService
	producer        Producer
	logger          *logrus.Logger
}

func NewHandler(paymentsService PaymentsService, eventsService EventsService, producer Producer, logger *logrus.Logger) *Handler {
	return &Handler{
		paymentsService: paymentsService,
		eventsService:   eventsService,
		producer:        producer,
		logger:          logger,
	}
}

func (h *Handler) Register(r *gin.Engine) {
	group := r.Group("/payments")
	{
		group.POST("", h.CreatePayment)
		group.POST("/:payment_id", h.GetPayment)
		group.POST("/:payment_id/process", h.ProcessPayment)
		group.POST("/:payment_id/refund", h.RefoundPayment)

		group.POST("/webhook/:provider", h.ProviderWebhook)
		group.POST("/user/:user_id", h.GetUserPayments)
	}

	group = r.Group("/tests")
	{
		group.POST("/pause-worker", h.TestPauseWorker)
		group.POST("/resume-worker", h.TestResumeWorker)
	}
}
