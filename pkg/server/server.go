package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mhgffqwoer/pr-service/pkg/config"
	"github.com/mhgffqwoer/pr-service/pkg/logger"
	"go.uber.org/zap"
)

type Server struct {
	srv *http.Server
}

func New(cfg *config.ServerConfig, handler http.Handler) *Server {
	return &Server{
		srv: &http.Server{
			Addr:              fmt.Sprintf(":%d", cfg.Port),
			Handler:           handler,
			ReadHeaderTimeout: 30 * time.Second,
		},
	}
}

func (s *Server) Start() {
	go func() {
		logger.Get().Infow("Server is running", "port", s.srv.Addr)
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Get().Fatalw("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Get().Infow("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		logger.Get().Fatalw("Server forced to shutdown", zap.Error(err))
	}
	logger.Get().Infow("Server exiting")
}
