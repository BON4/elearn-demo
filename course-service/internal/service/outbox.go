package service

import (
	"context"

	"github.com/BON4/elearn-demo/course-service/internal/domain"
	"github.com/BON4/elearn-demo/course-service/internal/repo"
	"gorm.io/gorm/clause"
)

type OutboxService struct {
	db *repo.MonoRepo
}

func NewOutboxService(rp *repo.MonoRepo) *OutboxService {
	return &OutboxService{db: rp}
}

// ProcessBatch принимает callback, который обрабатывает события.
// Сервис сам открывает транзакцию, делает FOR UPDATE SKIP LOCKED,
// и после callback обновляет статусы и коммитит.
// Если callback возвращает ошибку — транзакция откатывается.
func (s *OutboxService) ProcessBatch(ctx context.Context, limit int, fn func([]*domain.OutboxEvent) error) error {
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

	var events []*domain.OutboxEvent
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
