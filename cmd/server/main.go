package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpapi "github.com/mhgffqwoer/pr-service/internal/adapters/http"
	"github.com/mhgffqwoer/pr-service/internal/adapters/postgres"
	"github.com/mhgffqwoer/pr-service/internal/config"
	"github.com/mhgffqwoer/pr-service/internal/domain/services"
	"github.com/mhgffqwoer/pr-service/internal/logger"
	"go.uber.org/zap"
)

func main() {
	if err := config.Init(); err != nil {
		panic(fmt.Sprintf("Config init failed: %v", err))
	}

	cfg := config.Get()

	log := logger.InitLogger(cfg.Logging)
	defer func() { _ = log.Sync() }()

	pool, err := postgres.Connect(cfg.Database)
	if err != nil || pool == nil {
		log.Fatalw("Failed to connect to database", zap.Error(err))
	}
	defer func() { _ = pool.Close() }()

	teamRepo := postgres.NewTeamRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	prRepo := postgres.NewPullRequestRepository(pool)

	service := services.NewService(prRepo, userRepo, teamRepo)

	mux := http.NewServeMux()
	h := httpapi.NewHandlers(service)
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/team/add", h.CreateTeam)
	mux.HandleFunc("/team/get", h.GetTeam)
	mux.HandleFunc("/users/setIsActive", h.SetUserActive)
	mux.HandleFunc("/users/getReview", h.GetReview)
	mux.HandleFunc("/pullRequest/create", h.CreatePR)
	mux.HandleFunc("/pullRequest/merge", h.MergePR)
	mux.HandleFunc("/pullRequest/reassign", h.ReassignPR)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}

	go func() {
		logger.Get().Infow("Server is running", "port", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Get().Fatalw("Server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Get().Infow("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Get().Fatalw("Server forced to shutdown", zap.Error(err))
	}
	logger.Get().Infow("Server exiting")
}
