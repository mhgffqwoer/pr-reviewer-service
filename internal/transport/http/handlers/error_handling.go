package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/mhgffqwoer/pr-service/internal/models"
	"github.com/mhgffqwoer/pr-service/internal/services"
	"github.com/mhgffqwoer/pr-service/pkg/logger"
)

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
