package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestListMarketDataQualityLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/data-sources/market-quality-logs?asset_class=stock&severity=warn", nil)

	handler.ListMarketDataQualityLogs(ctx)

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

func TestListMarketDataQualityLogsRejectsInvalidHours(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/data-sources/market-quality-logs?hours=-1", nil)

	handler.ListMarketDataQualityLogs(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if message, _ := payload["message"].(string); message != "hours must be a positive integer" {
		t.Fatalf("unexpected message: %#v", payload["message"])
	}
}

func TestGetMarketDerivedTruthSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/data-sources/market-derived-truth-summary?asset_class=stock", nil)

	handler.GetMarketDerivedTruthSummary(ctx)

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

func TestGetMarketDataQualitySummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/data-sources/market-quality-summary?asset_class=stock&hours=24", nil)

	handler.GetMarketDataQualitySummary(ctx)

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

func TestRebuildStockDerivedTruth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/stocks/quotes/rebuild-derived-truth", strings.NewReader(`{"trade_date":"2026-03-22","days":2}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.RebuildStockDerivedTruth(ctx)

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
