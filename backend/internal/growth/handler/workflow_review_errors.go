package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
)

func resolveWorkflowReviewHTTPError(err error) (int, int) {
	if err == nil {
		return http.StatusInternalServerError, 50001
	}
	if errors.Is(err, sql.ErrNoRows) {
		return http.StatusNotFound, 40401
	}

	message := strings.ToLower(strings.TrimSpace(err.Error()))
	switch {
	case strings.Contains(message, "not found"):
		return http.StatusNotFound, 40401
	case strings.Contains(message, "already exists"),
		strings.Contains(message, "not pending"),
		strings.Contains(message, "reviewer mismatch"):
		return http.StatusConflict, 40901
	case strings.Contains(message, "module is required"),
		strings.Contains(message, "unsupported module"):
		return http.StatusBadRequest, 40001
	default:
		return http.StatusInternalServerError, 50001
	}
}
