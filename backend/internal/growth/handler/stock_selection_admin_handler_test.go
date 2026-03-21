package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/repo"
	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/config"
)

func newStockSelectionTestHandler() *AdminGrowthHandler {
	return NewAdminGrowthHandler(service.NewGrowthService(repo.NewInMemoryGrowthRepo()), config.Config{})
}

func TestGetStockSelectionOverview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/stock-selection/overview", nil)

	newStockSelectionTestHandler().GetStockSelectionOverview(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if code, ok := payload["code"].(float64); !ok || code != 0 {
		t.Fatalf("expected success code, got %#v", payload["code"])
	}
}

func TestCreateAndApproveStockSelectionRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	runRecorder := httptest.NewRecorder()
	runCtx, _ := gin.CreateTestContext(runRecorder)
	runCtx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/stock-selection/runs", strings.NewReader(`{"trade_date":"2026-03-21"}`))
	runCtx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateStockSelectionRun(runCtx)

	if runRecorder.Code != http.StatusOK {
		t.Fatalf("expected create run 200, got %d", runRecorder.Code)
	}

	reviewRecorder := httptest.NewRecorder()
	reviewCtx, _ := gin.CreateTestContext(reviewRecorder)
	reviewCtx.Params = gin.Params{{Key: "run_id", Value: "ssr_demo_001"}}
	reviewCtx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/stock-selection/reviews/ssr_demo_001/approve", strings.NewReader(`{"review_note":"通过"}`))
	reviewCtx.Request.Header.Set("Content-Type", "application/json")

	handler.ApproveStockSelectionReview(reviewCtx)

	if reviewRecorder.Code != http.StatusOK {
		t.Fatalf("expected approve review 200, got %d", reviewRecorder.Code)
	}
}
