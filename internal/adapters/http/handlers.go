package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mhgffqwoer/pr-service/internal/logger"
	"github.com/mhgffqwoer/pr-service/internal/models"
	"github.com/mhgffqwoer/pr-service/internal/services"
)

type Handlers struct {
	teamService *services.TeamService
	userService *services.UserService
	prService   *services.PullRequestService
}

func NewHandlers(service *services.Service) *Handlers {
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

func handleError(w http.ResponseWriter, err error, opts ...string) {
	log := logger.Get()

	switch {
	case errors.Is(err, services.ErrPRExists):
		log.Errorw(err.Error())
		writeError(w, http.StatusConflict, models.ErrorResponse{
			Code:    "PR_EXISTS",
			Message: fmt.Sprintf("PR %s already exists", opts[0]),
		})
		return
	case errors.Is(err, services.ErrNotFound):
		log.Errorw(err.Error())
		writeError(w, http.StatusNotFound, models.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "resource not found",
		})
		return
	case errors.Is(err, services.ErrTeamExists):
		log.Errorw(err.Error())
		writeError(w, http.StatusInternalServerError, models.ErrorResponse{
			Code:    "TEAM_EXISTS",
			Message: fmt.Sprintf("%s already exists", opts[0]),
		})
		return
	default:
		log.Errorw(err.Error())
		writeError(w, http.StatusInternalServerError, models.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "internal server error",
		})
		return
	}
}

func writeError(w http.ResponseWriter, status int, err models.ErrorResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(err)
}

func (h *Handlers) CreatePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
	}

	log := logger.Get()
	log.Debugw("POST /pullRequest/create: creating PR")

	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	createdPR, err := h.prService.Create(req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		handleError(w, err, req.PullRequestID)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]*models.PullRequest{"pull_request": createdPR})
}

func (h *Handlers) MergePR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
	}

	log := logger.Get()
	log.Debugw("POST /pullRequest/merge: merging PR")

	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	pr, err := h.prService.Merge(req.PullRequestID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"pr": pr})
}

func (h *Handlers) ReassignPR(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
	}

	log := logger.Get()
	log.Debugw("POST /pullRequest/reassign: reassigning PR")

	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldReviewerID string `json:"old_reviewer_id"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	pr, newUserID, err := h.prService.Reassign(req.PullRequestID, req.OldReviewerID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"pr": pr, "replaced_by": newUserID})
}

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

func (h *Handlers) SetUserActive(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusNotFound)
	}

	log := logger.Get()
	log.Debugw("POST /users/setIsActive: setting user active status")

	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	user, err := h.userService.SetActive(req.UserID, req.IsActive)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(user)
}

func (h *Handlers) GetReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusNotFound)
	}

	log := logger.Get()
	log.Debugw("GET /users/getReview: getting user review list")

	userID := r.URL.Query().Get("user_id")
	reviews, err := h.userService.GetReviewe(userID)
	if err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"user_id": userID, "pull_requests": reviews})
}
