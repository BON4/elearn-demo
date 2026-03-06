package service

import (
	"context"

	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	"github.com/BON4/elearn-demo/payment-service/internal/repo"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
)

type EventService struct {
	db *repo.MonoRepo
}

func NewEventService(rp *repo.MonoRepo) *EventService {
	return &EventService{db: rp}
}

// ProcessBatch принимает callback, который обрабатывает события.
// Сервис сам открывает транзакцию, делает FOR UPDATE SKIP LOCKED,
// и после callback обновляет статусы и коммитит.
// Если callback возвращает ошибку — транзакция откатывается.
func (s *EventService) ProcessPaymentEventBatch(ctx context.Context, limit int, fn func([]*domain.PaymentEvent) error) error {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var events []*domain.PaymentEvent
	if err := tx.Clauses(clause.Locking{
		Strength: "UPDATE",
		Options:  "SKIP LOCKED",
	}).Where("status = ?", domain.Pending).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(events) == 0 {
		tx.Rollback()
		return nil
	}

	if err := fn(events); err != nil {
		tx.Rollback()
		return err
	}

	for _, e := range events {
		if err := tx.Model(e).Updates(map[string]any{
			"status":      e.Status,
			"retry_count": e.RetryCount,
			"last_error":  e.LastError,
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

func (c *EventService) CreatePaymentEvent(ctx context.Context, event *domain.PaymentEvent) error {
	err := event.Validate()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"event_id":     event.ID,
		"event_status": event.Status,
		"event_type":   event.Type,
	}).Info("created payment event")

	return c.db.Create(event).Error
}
