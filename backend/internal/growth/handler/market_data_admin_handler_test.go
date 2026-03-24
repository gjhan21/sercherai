package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func createMarketBackfillRunForTest(t *testing.T, handler *AdminGrowthHandler) (string, string) {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/market-data/backfill",
		strings.NewReader(`{"run_type":"FULL","asset_scope":["STOCK","INDEX"],"source_key":"TUSHARE","batch_size":200}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateMarketDataBackfillRun(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected create run 200, got %d", recorder.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			RunID              string `json:"run_id"`
			UniverseSnapshotID string `json:"universe_snapshot_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if payload.Code != 0 || payload.Data.RunID == "" {
		t.Fatalf("unexpected create payload: %+v", payload)
	}
	return payload.Data.RunID, payload.Data.UniverseSnapshotID
}

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

func TestCreateMarketDataBackfillRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/market-data/backfill",
		strings.NewReader(`{"run_type":"FULL","asset_scope":["STOCK","INDEX"],"source_key":"TUSHARE","batch_size":200}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateMarketDataBackfillRun(ctx)

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
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data payload, got %#v", payload["data"])
	}
	if _, ok := data["run_id"].(string); !ok {
		t.Fatalf("expected run_id in response, got %#v", data["run_id"])
	}
}

func TestCreateMarketDataBackfillRunRejectsMissingAssetScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/market-data/backfill",
		strings.NewReader(`{"run_type":"FULL","source_key":"TUSHARE"}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateMarketDataBackfillRun(ctx)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}
}

func TestGetMarketCoverageSummary(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/data-sources/market-coverage-summary", nil)

	handler.GetMarketCoverageSummary(ctx)

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

func TestListMarketDataBackfillRuns(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()
	runID, _ := createMarketBackfillRunForTest(t, handler)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/market-data/backfill-runs", nil)

	handler.ListMarketDataBackfillRuns(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			Total int `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Code != 0 || payload.Data.Total == 0 {
		t.Fatalf("unexpected list payload: %+v", payload)
	}
	found := false
	for _, item := range payload.Data.Items {
		if item.ID == runID {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected run %s in payload: %+v", runID, payload.Data.Items)
	}
}

func TestRetryMarketDataBackfillRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()
	runID, _ := createMarketBackfillRunForTest(t, handler)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: runID}}
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/market-data/backfill-runs/"+runID+"/retry", strings.NewReader(`{"retry_mode":"FAILED_ONLY"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.RetryMarketDataBackfillRun(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			RunID string `json:"run_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Code != 0 || payload.Data.RunID == "" || payload.Data.RunID == runID {
		t.Fatalf("unexpected retry payload: %+v", payload)
	}
}

func TestGetMarketUniverseSnapshot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()
	_, snapshotID := createMarketBackfillRunForTest(t, handler)

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: snapshotID}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/market-data/universe-snapshots/"+snapshotID, nil)

	handler.GetMarketUniverseSnapshot(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Snapshot struct {
				ID string `json:"id"`
			} `json:"snapshot"`
			Items []map[string]any `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Code != 0 || payload.Data.Snapshot.ID != snapshotID || len(payload.Data.Items) == 0 {
		t.Fatalf("unexpected snapshot payload: %+v", payload)
	}
}

func TestSyncMarketDataMaster(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/market-data/master/sync",
		strings.NewReader(`{"source_key":"TUSHARE","asset_scope":["STOCK","INDEX"]}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.SyncMarketDataMaster(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			SnapshotID string `json:"snapshot_id"`
			Result     struct {
				DataKind      string `json:"data_kind"`
				SnapshotCount int    `json:"snapshot_count"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Code != 0 || payload.Data.SnapshotID == "" || payload.Data.Result.DataKind == "" || payload.Data.Result.SnapshotCount <= 0 {
		t.Fatalf("unexpected master sync payload: %+v", payload)
	}
}

func TestSyncMarketDataQuotes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/market-data/quotes/sync",
		strings.NewReader(`{"source_key":"MOCK","asset_scope":["INDEX"],"symbols":["000300.SH"],"days":2}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.SyncMarketDataQuotes(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Result struct {
				DataKind string `json:"data_kind"`
				BarCount int    `json:"bar_count"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Code != 0 || payload.Data.Result.DataKind == "" || payload.Data.Result.BarCount <= 0 {
		t.Fatalf("unexpected quotes sync payload: %+v", payload)
	}
}

func TestSyncMarketDataDailyBasic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/market-data/daily-basic/sync",
		strings.NewReader(`{"source_key":"TUSHARE","asset_scope":["INDEX"],"symbols":["000300.SH"],"days":30}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.SyncMarketDataDailyBasic(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Result struct {
				DataKind string `json:"data_kind"`
				Results  []struct {
					Status string `json:"status"`
				} `json:"results"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Code != 0 || payload.Data.Result.DataKind == "" || len(payload.Data.Result.Results) == 0 || payload.Data.Result.Results[0].Status == "" {
		t.Fatalf("unexpected daily basic sync payload: %+v", payload)
	}
}

func TestSyncMarketDataMoneyflow(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/market-data/moneyflow/sync",
		strings.NewReader(`{"source_key":"TUSHARE","asset_scope":["STOCK"],"symbols":["600519.SH","000001.SZ"],"days":20}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.SyncMarketDataMoneyflow(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Result struct {
				DataKind string `json:"data_kind"`
				BarCount int    `json:"bar_count"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Code != 0 || payload.Data.Result.DataKind == "" || payload.Data.Result.BarCount <= 0 {
		t.Fatalf("unexpected moneyflow sync payload: %+v", payload)
	}
}

func TestRebuildMarketDataTruth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newStockSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/market-data/truth/rebuild",
		strings.NewReader(`{"source_key":"MOCK","asset_scope":["STOCK","INDEX"],"trade_date_from":"2026-03-23","trade_date_to":"2026-03-24"}`),
	)
	ctx.Request.Header.Set("Content-Type", "application/json")

	handler.RebuildMarketDataTruth(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Result struct {
				DataKind   string `json:"data_kind"`
				TruthCount int    `json:"truth_count"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Code != 0 || payload.Data.Result.DataKind == "" || payload.Data.Result.TruthCount <= 0 {
		t.Fatalf("unexpected truth rebuild payload: %+v", payload)
	}
}
