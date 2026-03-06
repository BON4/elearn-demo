package service

import (
	"context"

	"github.com/BON4/elearn-demo/payment-service/internal/domain"
	"github.com/BON4/elearn-demo/payment-service/internal/repo"
	"github.com/google/uuid"
)

type PaymentProvider interface {
	MakePayment(ctx context.Context, payment *domain.Payment) error
	RefoundPayment(ctx context.Context, payment *domain.Payment) error
}

type EventProducer interface {
	CreatePaymentEvent(ctx context.Context, event *domain.PaymentEvent) error
}

type CourseService interface {
	GetCourse(ctx context.Context, courseID uuid.UUID) (*domain.CourseRM, error)
}

type PaymentsService struct {
	db              *repo.MonoRepo
	eventProducer   EventProducer
	paymentProvider PaymentProvider
	courseService   CourseService
}

func NewPaymentsService(
	rp *repo.MonoRepo,
	eventProducer EventProducer,
	paymentProvider PaymentProvider,
	courseSrv CourseService,
) *PaymentsService {
	return &PaymentsService{
		db:              rp,
		eventProducer:   eventProducer,
		paymentProvider: paymentProvider,
		courseService:   courseSrv,
	}
}

func (p *PaymentsService) withRepo(rp *repo.MonoRepo) *PaymentsService {
	return &PaymentsService{
		db:              rp,
		eventProducer:   p.eventProducer,
		paymentProvider: p.paymentProvider,
		courseService:   p.courseService,
	}
}
