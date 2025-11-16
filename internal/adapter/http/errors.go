package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"

	"prservice/internal/adapter/http/api"
	"prservice/internal/domain"
)

func writeError(w http.ResponseWriter, err error) {
	var resp api.ErrorResponse

	switch {
	case errors.Is(err, domain.ErrTeamExists):
		resp.Error.Code = "TEAM_EXISTS"
		resp.Error.Message = err.Error()
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, domain.ErrPRExists):
		resp.Error.Code = "PR_EXISTS"
		resp.Error.Message = err.Error()
		w.WriteHeader(http.StatusConflict)
	case errors.Is(err, domain.ErrPRMerged):
		resp.Error.Code = "PR_MERGED"
		resp.Error.Message = err.Error()
		w.WriteHeader(http.StatusConflict)
	case errors.Is(err, domain.ErrNotAssigned):
		resp.Error.Code = "NOT_ASSIGNED"
		resp.Error.Message = err.Error()
		w.WriteHeader(http.StatusConflict)
	case errors.Is(err, domain.ErrNoCandidate):
		resp.Error.Code = "NO_CANDIDATE"
		resp.Error.Message = err.Error()
		w.WriteHeader(http.StatusConflict)
	case errors.Is(err, domain.ErrNotFound):
		resp.Error.Code = "NOT_FOUND"
		resp.Error.Message = err.Error()
		w.WriteHeader(http.StatusNotFound)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error.Code = "NOT_FOUND"
		resp.Error.Message = "internal error"
	}

	_ = json.NewEncoder(w).Encode(resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
