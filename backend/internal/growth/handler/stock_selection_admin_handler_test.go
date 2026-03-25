package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/growth/repo"
	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/config"
)

func newStockSelectionTestHandler() *AdminGrowthHandler {
	return NewAdminGrowthHandler(service.NewGrowthService(repo.NewInMemoryGrowthRepo()), config.Config{})
}

func newStockSelectionTestHandlerWithRepo() (*AdminGrowthHandler, *repo.InMemoryGrowthRepo) {
	repository := repo.NewInMemoryGrowthRepo()
	return NewAdminGrowthHandler(service.NewGrowthService(repository), config.Config{}), repository
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

func TestListAndReviewStockEventClusters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, repository := newStockSelectionTestHandlerWithRepo()
	_, err := repository.AdminUpsertStockEventCluster(model.StockEventCluster{
		ID:            "sec_demo_001",
		ClusterKey:    "draft:event:001",
		EventType:     "NEWS",
		Title:         "公司动态更新",
		PrimarySymbol: "600519.SH",
		Status:        "CLUSTERED",
		ReviewStatus:  "PENDING",
		Confidence:    0.62,
		Metadata: map[string]any{
			"review_priority":     "HIGH",
			"review_reason_codes": []string{"LOW_CONFIDENCE", "GENERIC_EVENT_TYPE"},
			"draft_source":        "market_news_sync",
		},
	})
	if err != nil {
		t.Fatalf("seed stock event cluster: %v", err)
	}

	listRecorder := httptest.NewRecorder()
	listCtx, _ := gin.CreateTestContext(listRecorder)
	listCtx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/stock-selection/events?review_status=PENDING&review_priority=HIGH", nil)

	handler.ListStockEventClusters(listCtx)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected list events 200, got %d", listRecorder.Code)
	}
	var listPayload struct {
		Code int `json:"code"`
		Data struct {
			Items []model.StockEventCluster `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listRecorder.Body.Bytes(), &listPayload); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listPayload.Code != 0 || len(listPayload.Data.Items) != 1 {
		t.Fatalf("unexpected list response: %+v", listPayload)
	}

	reviewRecorder := httptest.NewRecorder()
	reviewCtx, _ := gin.CreateTestContext(reviewRecorder)
	reviewCtx.Params = gin.Params{{Key: "id", Value: "sec_demo_001"}}
	reviewCtx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/stock-selection/events/sec_demo_001/review", strings.NewReader(`{"review_status":"APPROVED","review_note":"事件成立","reviewer":"reviewer_001","review_metadata":{"manual":true}}`))
	reviewCtx.Request.Header.Set("Content-Type", "application/json")

	handler.ReviewStockEventCluster(reviewCtx)

	if reviewRecorder.Code != http.StatusOK {
		t.Fatalf("expected review event 200, got %d", reviewRecorder.Code)
	}
	reviewTasks, total, err := repository.AdminListReviewTasks("STOCK_EVENT", "", "", "", 1, 20)
	if err != nil {
		t.Fatalf("list stock event review tasks: %v", err)
	}
	if total != 1 || len(reviewTasks) != 1 || reviewTasks[0].Status != "APPROVED" {
		t.Fatalf("expected approved stock event review task, got total=%d items=%+v", total, reviewTasks)
	}

	detailRecorder := httptest.NewRecorder()
	detailCtx, _ := gin.CreateTestContext(detailRecorder)
	detailCtx.Params = gin.Params{{Key: "id", Value: "sec_demo_001"}}
	detailCtx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/stock-selection/events/sec_demo_001", nil)

	handler.GetStockEventCluster(detailCtx)

	if detailRecorder.Code != http.StatusOK {
		t.Fatalf("expected get event 200, got %d", detailRecorder.Code)
	}
	var detailPayload struct {
		Code int                     `json:"code"`
		Data model.StockEventCluster `json:"data"`
	}
	if err := json.Unmarshal(detailRecorder.Body.Bytes(), &detailPayload); err != nil {
		t.Fatalf("decode detail response: %v", err)
	}
	if detailPayload.Code != 0 || detailPayload.Data.ReviewStatus != "APPROVED" || detailPayload.Data.Status != "REVIEWED" {
		t.Fatalf("unexpected detail payload: %+v", detailPayload)
	}
}
