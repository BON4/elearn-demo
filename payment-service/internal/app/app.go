package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/BON4/elearn-demo/payment-service/internal/config"
	"github.com/BON4/elearn-demo/payment-service/internal/handlers"
	"github.com/BON4/elearn-demo/payment-service/internal/infra"
	"github.com/BON4/elearn-demo/payment-service/internal/logger"
	"github.com/BON4/elearn-demo/payment-service/internal/payment"
	outbox "github.com/BON4/elearn-demo/payment-service/internal/producer"
	"github.com/BON4/elearn-demo/payment-service/internal/repo"
	"github.com/BON4/elearn-demo/payment-service/internal/server"
	"github.com/BON4/elearn-demo/payment-service/internal/service"
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

	paymentProvider := payment.NewMockPaymentProvider(cfg.PaymentProviderWebhookURL)

	eventsServise := service.NewEventService(rp)
	coursesService := service.NewCourseServiceClient(cfg.CoursesUrl)

	paymentsService := service.NewPaymentsService(
		rp,
		eventsServise,
		paymentProvider,
		coursesService,
	)

	producerWg := sync.WaitGroup{}
	producerWorker := outbox.NewProducerWorker(
		eventsServise,
		rbmq,
		cfg.ProducerInterval,
	)

	h := handlers.NewHandler(paymentsService, eventsServise, producerWorker, lg)
	srv := server.NewServer(*cfg, h, lg)

	go func() {
		if err := srv.StartBlocking(); err != nil {
			lg.WithError(err).Error("server start error")
		}
	}()

	producerWg.Go(func() {
		producerWorker.Run(ctx)
	})

	<-ctx.Done()

	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(stopCtx); err != nil {
		lg.WithError(err).Error("shutdown error")
	}

	producerWg.Wait()

	lg.Println("Server stopped")
	return nil
}
