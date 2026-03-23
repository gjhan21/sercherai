package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"testing"
)

func TestResolveWorkflowReviewHTTPErrorTargetNotFound(t *testing.T) {
	status, code := resolveWorkflowReviewHTTPError(errors.New("target not found"))
	if status != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", status)
	}
	if code != 40401 {
		t.Fatalf("expected code 40401, got %d", code)
	}
}

func TestResolveWorkflowReviewHTTPErrorPendingTaskConflict(t *testing.T) {
	status, code := resolveWorkflowReviewHTTPError(errors.New("pending review task already exists"))
	if status != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", status)
	}
	if code != 40901 {
		t.Fatalf("expected code 40901, got %d", code)
	}
}

func TestResolveWorkflowReviewHTTPErrorMissingReviewTask(t *testing.T) {
	status, code := resolveWorkflowReviewHTTPError(sql.ErrNoRows)
	if status != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", status)
	}
	if code != 40401 {
		t.Fatalf("expected code 40401, got %d", code)
	}
}

func TestResolveWorkflowReviewHTTPErrorFallbackInternal(t *testing.T) {
	status, code := resolveWorkflowReviewHTTPError(errors.New("unexpected database error"))
	if status != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", status)
	}
	if code != 50001 {
		t.Fatalf("expected code 50001, got %d", code)
	}
}
