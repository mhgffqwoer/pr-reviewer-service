package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mhgffqwoer/pr-service/internal/models"
	"github.com/mhgffqwoer/pr-service/internal/logger"
)

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
