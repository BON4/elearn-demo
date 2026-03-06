package server

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/BON4/elearn-demo/payment-service/internal/config"
	"github.com/BON4/elearn-demo/payment-service/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
	cfg           config.Config
	logger        *logrus.Logger
	engine        *gin.Engine
	srv           *http.Server
	shutDownQueue []func(context.Context) error
}

func NewServer(
	cfg config.Config,
	h *handlers.Handler,
	lg *logrus.Logger,
) *Server {
	g := gin.New()

	s := &Server{
		cfg:    cfg,
		logger: lg,
		engine: g,
	}

	s.srv = &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      s.engine,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	h.Register(s.engine)

	return s
}

func (s *Server) StartBlocking() error {
	s.logger.Infof("starting server on %s", s.srv.Addr)
	return s.srv.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	var shutDownErrors = []error{}
	for _, f := range s.shutDownQueue {
		if err := f(ctx); err != nil {
			shutDownErrors = append(shutDownErrors, err)
		}
	}
	return errors.Join(append(shutDownErrors, s.srv.Shutdown(ctx))...)
}
