package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mhgffqwoer/pr-service/internal/models"
	"github.com/mhgffqwoer/pr-service/pkg/logger"
)

func (h *Handlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
	}

	log := logger.Get()
	log.Debugw("POST /team/add: creating team")

	var reqTeam models.Team
	if err := json.NewDecoder(r.Body).Decode(&reqTeam); err != nil {
		logger.Get().Errorw("Invalid JSON", err)
		writeError(w, http.StatusBadRequest, models.ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid JSON",
		})
		return
	}

	createdTeam, err := h.teamService.AddTeam(&reqTeam)
	if err != nil {
		handleError(w, err, reqTeam.TeamName)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]*models.Team{"team": createdTeam})
}

func (h *Handlers) GetTeam(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
	}

	log := logger.Get()
	log.Debugw("GET /team/get: getting team by name")

	teamName := r.URL.Query().Get("team_name")
	team, err := h.teamService.GetTeam(teamName)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(team)
}
