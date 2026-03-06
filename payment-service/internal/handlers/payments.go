package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PaymentsService interface {
	GetPayment(ctx context.Context, paymentID uuid.UUID) (*domain.Payment, error)
	GetUserPayments(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*domain.Payment, error)

	MarkProcessedPayment(ctx context.Context, paymentID uuid.UUID) error
	MarkSuccesedPayment(ctx context.Context, paymentID uuid.UUID) error
	MarkRefoundedPayment(ctx context.Context, paymentID uuid.UUID) error

	MakePaymentRefoundRequest(ctx context.Context, paymentID uuid.UUID) error
	CreatePayment(ctx context.Context, userID, courseID uuid.UUID, amount int64, currency string, provider string) (*domain.Payment, error)
}

type EventsService interface {
	CreatePaymentEvent(ctx context.Context, event *domain.PaymentEvent) error
}

type CreatePaymentRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	CourseID uuid.UUID `json:"course_id" binding:"required"`
	Amount   int64     `json:"amount" binding:"required,gt=0"`
	Currency string    `json:"currency" binding:"required"`
	Provider string    `json:"provider" binding:"required"`
}

type PaymentResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	CourseID  uuid.UUID `json:"course_id"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func paymentToDTO(p *domain.Payment) *PaymentResponse {
	return &PaymentResponse{
		ID:        p.ID,
		UserID:    p.UserID,
		CourseID:  p.CourseID,
		Amount:    p.Amount,
		Currency:  p.Currency,
		Status:    string(p.Status),
		CreatedAt: p.CreatedAt,
	}
}

func (h *Handler) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment, err := h.paymentsService.CreatePayment(
		c.Request.Context(),
		req.UserID,
		req.CourseID,
		req.Amount,
		req.Currency,
		req.Provider,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, paymentToDTO(payment))
}

func (h *Handler) ProcessPayment(c *gin.Context) {
	paymentIDStr := c.Param("payment_id")

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment_id"})
		return
	}

	h.paymentsService.MarkProcessedPayment(c, paymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *Handler) GetPayment(c *gin.Context) {
	paymentIDStr := c.Param("payment_id")

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment_id"})
		return
	}

	payment, err := h.paymentsService.GetPayment(
		c.Request.Context(),
		paymentID,
	)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, paymentToDTO(payment))
}

func (h *Handler) GetUserPayments(c *gin.Context) {
	userIDStr := c.Param("user_id")

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	payments, err := h.paymentsService.GetUserPayments(
		c.Request.Context(),
		userID,
		offset,
		limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := make([]*PaymentResponse, 0, len(payments))

	for _, p := range payments {
		resp = append(resp, paymentToDTO(p))
	}

	c.JSON(http.StatusOK, resp)
}

type ProviderWebhookRequest struct {
	PaymentID         uuid.UUID `json:"payment_id"`
	Status            string    `json:"status"`
	ProviderPaymentID string    `json:"provider_payment_id"`
}

func (h *Handler) ProviderWebhook(c *gin.Context) {
	var req ProviderWebhookRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	switch req.Status {

	case "succeeded":

		err := h.paymentsService.MarkSuccesedPayment(
			ctx,
			req.PaymentID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	case "refunded":

		err := h.paymentsService.MarkRefoundedPayment(
			ctx,
			req.PaymentID,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown status"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) RefoundPayment(c *gin.Context) {
	paymentIDStr := c.Param("payment_id")

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment_id"})
		return
	}

	err = h.paymentsService.MakePaymentRefoundRequest(c, paymentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payment_id"})
		return
	}

	c.Status(http.StatusOK)
}
