package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"pickup-helper/internal/config"

	"github.com/gin-gonic/gin"
)

const shutdownTimeout = 10 * time.Second

// Server wraps a gin engine with an *http.Server and supports graceful shutdown.
type Server struct {
	engine  *gin.Engine
	httpSrv *http.Server
}

// New constructs a Server bound to cfg.Server.Port.
func New(cfg *config.Config, engine *gin.Engine) *Server {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	return &Server{
		engine: engine,
		httpSrv: &http.Server{
			Addr:         addr,
			Handler:      engine,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
	}
}

// Addr returns the address the server listens on.
func (s *Server) Addr() string {
	return s.httpSrv.Addr
}

// Run blocks until ctx is cancelled. On cancellation it attempts graceful
// shutdown with a 10s budget.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return s.httpSrv.Shutdown(shutdownCtx)
	case err := <-errCh:
		return err
	}
}
