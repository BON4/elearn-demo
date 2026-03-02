package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/BON4/elearn-demo/access-service/internal/config"
	"github.com/BON4/elearn-demo/access-service/internal/consumer"
	"github.com/BON4/elearn-demo/access-service/internal/handlers"
	"github.com/BON4/elearn-demo/access-service/internal/infra"
	"github.com/BON4/elearn-demo/access-service/internal/logger"
	"github.com/BON4/elearn-demo/access-service/internal/repo"
	"github.com/BON4/elearn-demo/access-service/internal/server"
	"github.com/BON4/elearn-demo/access-service/internal/service"
)

func Run(ctx context.Context, cfg *config.Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	lg := logger.New()

	pgConn, err := infra.NewPostgres(ctx, cfg.DBUrl)
	if err != nil {
		return fmt.Errorf("failed to init postgres: %w", err)
	}

	rp := repo.NewMonoRepo(pgConn.DB)

	err = rp.MigrateDomain()
	if err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}

	rbmq, err := infra.NewRabbitMQ(ctx, cfg.RBBMQUrl)
	if err != nil {
		return fmt.Errorf("failed to init rbmq: %w", err)
	}

	coursesService := service.NewCoursesService(rp)
	accessService := service.NewAccessService(rp)
	consWorkerWg := sync.WaitGroup{}
	cons := consumer.NewConsumer(rbmq, "access", coursesService, cfg.ConsumerInterval)

	h := handlers.NewHandler(accessService, cons, lg)
	srv := server.NewServer(*cfg, h, lg)

	go func() {
		if err := srv.StartBlocking(); err != nil {
			lg.WithError(err).Error("server start error")
		}
	}()

	consWorkerWg.Go(func() {
		cons.Run(ctx)
	})

	<-ctx.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(stopCtx); err != nil {
		lg.WithError(err).Error("shutdown error")
	}

	consWorkerWg.Wait()

	lg.Println("Server stopped")
	return nil
}
