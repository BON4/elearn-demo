package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type MockPaymentProvider struct {
	webhookURL string
	client     *http.Client
}

func NewMockPaymentProvider(webhookURL string) *MockPaymentProvider {
	return &MockPaymentProvider{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type MockPaymentProviderWebhookResponse struct {
	PaymentID         uuid.UUID `json:"payment_id"`
	Status            string    `json:"status"` // "succeeded" или "refunded"
	ProviderPaymentID string    `json:"provider_payment_id"`
}

func (m *MockPaymentProvider) MakePayment(ctx context.Context, payment *domain.Payment) error {
	log.WithField("payment_id", payment.ID).Info("payment creation requested")

	go m.resolvePayment(payment.ID, "succeeded")

	return nil
}

func (m *MockPaymentProvider) RefoundPayment(ctx context.Context, payment *domain.Payment) error {
	log.WithField("payment_id", payment.ID).Info("payment refund requested")

	go m.resolvePayment(payment.ID, "refunded")

	return nil
}

func (m *MockPaymentProvider) resolvePayment(paymentID uuid.UUID, status string) {
	time.Sleep(2 * time.Second)

	payload := MockPaymentProviderWebhookResponse{
		PaymentID:         paymentID,
		Status:            status,
		ProviderPaymentID: uuid.NewString(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		log.WithError(err).Error("mock_provider: failed to marshal payload")
		return
	}

	req, err := http.NewRequest(http.MethodPost, m.webhookURL, bytes.NewBuffer(body))
	if err != nil {
		log.WithError(err).Error("mock_provider: failed to create request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		log.WithError(err).Errorf("mock_provider: failed to send webhook to %s", m.webhookURL)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Errorf("mock_provider: webhook returned non-200 status: %d", resp.StatusCode)
		return
	}

	log.WithFields(log.Fields{
		"payment_id": paymentID,
		"status":     status,
	}).Info("mock_provider: webhook sent successfully")
}
