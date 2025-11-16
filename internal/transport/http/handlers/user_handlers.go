package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mhgffqwoer/pr-service/pkg/logger"
)

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
