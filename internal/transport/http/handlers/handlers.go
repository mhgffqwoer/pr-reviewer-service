package handlers

import (
	"net/http"

	"github.com/mhgffqwoer/pr-service/internal/services"
	"github.com/mhgffqwoer/pr-service/pkg/logger"
)

type Handlers struct {
	teamService *services.TeamService
	userService *services.UserService
	prService   *services.PullRequestService
}

func New(service *services.Service) *Handlers {
	return &Handlers{
		teamService: service.TeamService,
		userService: service.UserService,
		prService:   service.PullRequestService,
	}
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	logger.Get().Debugw("Health check OK")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}
