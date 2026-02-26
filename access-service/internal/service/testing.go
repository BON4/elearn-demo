package service

import (
	"github.com/BON4/elearn-demo/course-service/internal/outbox"
)

type TestingService struct {
	outBoxWorker *outbox.Worker
}

func NewTestingService(wrk *outbox.Worker) *TestingService {
	return &TestingService{
		outBoxWorker: wrk,
	}
}

func (t *TestingService) PauseWorker() {
	t.outBoxWorker.Pause()
}

func (t *TestingService) ResumeWorker() {
	t.outBoxWorker.Resume()
}
