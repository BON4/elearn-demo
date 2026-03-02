package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresClient struct {
	DB *gorm.DB
}

func tryToConnect(ctx context.Context, dsn string) (*gorm.DB, error) {
	var gormDB *gorm.DB
	var err error
	for i := range 5 {
		gormDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			return gormDB, nil
		}
		logrus.WithError(err).Warnf("postgres not ready, attempt %d/5, retrying...", i+1)
		time.Sleep(2 * time.Second)
	}
	return nil, err
}

func NewPostgres(ctx context.Context, dsn string) (*PostgresClient, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	gormDB, err := tryToConnect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	// Проверка соединения
	if err := ping(ctx, sqlDB); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	// Базовые настройки пула (демо, но правильно)
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &PostgresClient{
		DB: gormDB,
	}, nil
}

func ping(ctx context.Context, db interface {
	PingContext(context.Context) error
}) error {
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("postgres ping failed: %w", err)
	}
	return nil
}

func (c *PostgresClient) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
