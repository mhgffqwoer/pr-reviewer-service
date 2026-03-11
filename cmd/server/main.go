package main

import (
	"fmt"

	"github.com/mhgffqwoer/pr-service/internal/adapters/http"
	"github.com/mhgffqwoer/pr-service/internal/adapters/postgres"
	"github.com/mhgffqwoer/pr-service/internal/app"
	"github.com/mhgffqwoer/pr-service/internal/config"
	"github.com/mhgffqwoer/pr-service/internal/logger"
	"github.com/mhgffqwoer/pr-service/internal/services"
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

	r := http.NewRouter()
	h := http.NewHandlers(service)
	r.HandleFunc("/health", h.Health)
	r.HandleFunc("/team/add", h.CreateTeam)
	r.HandleFunc("/team/get", h.GetTeam)
	r.HandleFunc("/users/setIsActive", h.SetUserActive)
	r.HandleFunc("/users/getReview", h.GetReview)
	r.HandleFunc("/pullRequest/create", h.CreatePR)
	r.HandleFunc("/pullRequest/merge", h.MergePR)
	r.HandleFunc("/pullRequest/reassign", h.ReassignPR)

	srv := app.New(cfg.Server, r)
	srv.Start()
}
