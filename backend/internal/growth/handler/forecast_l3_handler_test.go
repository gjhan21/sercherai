package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/growth/repo"
	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/config"
)

func TestCreateForecastL3RunRequiresAuth(t *testing.T) {
	growthHandler := newUserGrowthHandlerForTest(t)
	router := gin.New()
	router.POST("/api/v1/forecast/runs", growthHandler.CreateForecastL3Run)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/runs", bytes.NewBufferString(`{"target_type":"STOCK","target_key":"600519.SH"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateForecastL3RunReturnsQueuedRun(t *testing.T) {
	growthHandler := newUserGrowthHandlerForTest(t)
	router := gin.New()
	attachUserID(router, "user_001")
	router.POST("/api/v1/forecast/runs", growthHandler.CreateForecastL3Run)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/runs", bytes.NewBufferString(`{"target_type":"STOCK","target_id":"reco_001","target_key":"600519.SH","target_label":"贵州茅台","source":"RECOMMENDATION","source_id":"reco_001","source_path":"/recommendations","priority_score":0.81,"reason":"need deeper view"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if payload.Data.ID == "" || payload.Data.Status != "QUEUED" {
		t.Fatalf("expected queued run response, got %+v", payload.Data)
	}
}

func TestCreateForecastL3RunRejectsMissingStrictContextForUserRequest(t *testing.T) {
	growthHandler := newUserGrowthHandlerForTest(t)
	router := gin.New()
	attachUserID(router, "user_001")
	router.POST("/api/v1/forecast/runs", growthHandler.CreateForecastL3Run)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/runs", bytes.NewBufferString(`{"target_type":"STOCK","target_key":"600519.SH","target_label":"贵州茅台","reason":"need deeper view"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestListStockRecommendationHistoryRequiresAuth(t *testing.T) {
	growthHandler := newUserGrowthHandlerForTest(t)
	router := gin.New()
	router.GET("/api/v1/stocks/recommendations/history", growthHandler.ListStockRecommendationHistory)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stocks/recommendations/history", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestListStockRecommendationHistoryReturnsItemsAndSummary(t *testing.T) {
	growthHandler := newUserGrowthHandlerForTest(t)
	router := gin.New()
	attachUserID(router, "user_001")
	router.GET("/api/v1/stocks/recommendations/history", growthHandler.ListStockRecommendationHistory)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stocks/recommendations/history?page=1&page_size=20", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Items   []model.StockRecommendationHistoryItem    `json:"items"`
			Summary model.StockRecommendationHistorySummary   `json:"summary"`
			Total   int                                       `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if payload.Data.Total == 0 || len(payload.Data.Items) == 0 {
		t.Fatalf("expected non-empty history payload, got %+v", payload.Data)
	}
	if payload.Data.Summary.TotalCount == 0 {
		t.Fatalf("expected summary payload, got %+v", payload.Data.Summary)
	}
}

func TestListForecastL3HistoryReturnsSucceededTimeline(t *testing.T) {
	growthRepo := repo.NewInMemoryGrowthRepo()
	seedForecastHistoryHandlerRuns(t, growthRepo, "600519.SH", model.StrategyForecastL3TargetTypeStock)
	growthService := service.NewGrowthService(growthRepo)
	growthHandler := NewUserGrowthHandler(growthService, config.Config{})

	router := gin.New()
	attachUserID(router, "user_001")
	router.GET("/api/v1/forecast/targets/history", growthHandler.ListForecastL3History)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forecast/targets/history?target_type=STOCK&target_key=600519.SH", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Items []model.StrategyForecastL3HistoryItem `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 || len(payload.Data.Items) < 2 {
		t.Fatalf("expected successful history timeline, got %+v", payload.Data)
	}
}

func TestGetForecastL3HistoryCompareDefaultsToLatestVsPrevious(t *testing.T) {
	growthRepo := repo.NewInMemoryGrowthRepo()
	seedForecastHistoryHandlerRuns(t, growthRepo, "AU2408", model.StrategyForecastL3TargetTypeFutures)
	growthService := service.NewGrowthService(growthRepo)
	growthHandler := NewUserGrowthHandler(growthService, config.Config{})

	router := gin.New()
	attachUserID(router, "user_001")
	router.GET("/api/v1/forecast/targets/history/compare", growthHandler.GetForecastL3HistoryCompare)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forecast/targets/history/compare?target_type=FUTURES&target_key=AU2408", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Code int `json:"code"`
		Data model.StrategyForecastL3HistoryCompare `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 || payload.Data.LeftRun == nil || payload.Data.RightRun == nil {
		t.Fatalf("expected default latest vs previous compare, got %+v", payload.Data)
	}
}

func TestGetForecastL3RunReviewReturnsExplainableScore(t *testing.T) {
	growthRepo := repo.NewInMemoryGrowthRepo()
	seedForecastHistoryHandlerRuns(t, growthRepo, "RB2609", model.StrategyForecastL3TargetTypeFutures)
	items, err := growthRepo.ListStrategyForecastL3HistoryForTarget("user_001", model.StrategyForecastL3TargetTypeFutures, "RB2609", 1, 10)
	if err != nil || len(items) == 0 {
		t.Fatalf("seed history items error = %v items=%d", err, len(items))
	}
	growthService := service.NewGrowthService(growthRepo)
	growthHandler := NewUserGrowthHandler(growthService, config.Config{})

	router := gin.New()
	attachUserID(router, "user_001")
	router.GET("/api/v1/forecast/runs/:id/review", growthHandler.GetForecastL3RunReview)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forecast/runs/"+items[0].RunID+"/review", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Code int                             `json:"code"`
		Data model.StrategyForecastL3RunReview `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 || payload.Data.ReviewGrade == "" || payload.Data.ReviewScore == 0 {
		t.Fatalf("expected explainable review score, got %+v", payload.Data)
	}
}

func TestListForecastL3HistoryRejectsOtherUsersTargetHistory(t *testing.T) {
	growthRepo := repo.NewInMemoryGrowthRepo()
	seedForecastHistoryHandlerRuns(t, growthRepo, "600519.SH", model.StrategyForecastL3TargetTypeStock)
	growthService := service.NewGrowthService(growthRepo)
	growthHandler := NewUserGrowthHandler(growthService, config.Config{})

	router := gin.New()
	attachUserID(router, "another_user")
	router.GET("/api/v1/forecast/targets/history", growthHandler.ListForecastL3History)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forecast/targets/history?target_type=STOCK&target_key=600519.SH", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Items []model.StrategyForecastL3HistoryItem `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if len(payload.Data.Items) != 0 {
		t.Fatalf("expected no visible history for another user, got %+v", payload.Data.Items)
	}
}

func TestGetForecastL3RunReviewRejectsOtherUsersRun(t *testing.T) {
	growthRepo := repo.NewInMemoryGrowthRepo()
	seedForecastHistoryHandlerRuns(t, growthRepo, "RB2609", model.StrategyForecastL3TargetTypeFutures)
	items, err := growthRepo.ListStrategyForecastL3HistoryForTarget("user_001", model.StrategyForecastL3TargetTypeFutures, "RB2609", 1, 10)
	if err != nil || len(items) == 0 {
		t.Fatalf("seed history items error = %v items=%d", err, len(items))
	}
	growthService := service.NewGrowthService(growthRepo)
	growthHandler := NewUserGrowthHandler(growthService, config.Config{})

	router := gin.New()
	attachUserID(router, "another_user")
	router.GET("/api/v1/forecast/runs/:id/review", growthHandler.GetForecastL3RunReview)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/forecast/runs/"+items[0].RunID+"/review", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestAdminListForecastL3RunsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	growthRepo := repo.NewInMemoryGrowthRepo()
	growthService := service.NewGrowthService(growthRepo)
	adminHandlers := NewAdminHandlers(growthService, config.Config{})

	if _, err := growthRepo.CreateStrategyForecastL3Run(repoRunInput("STOCK", "600519.SH", "admin_001")); err != nil {
		t.Fatalf("seed run: %v", err)
	}

	router := gin.New()
	router.GET("/api/v1/admin/forecast/runs", adminHandlers.Forecast.ListForecastL3Runs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/forecast/runs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Items []map[string]any `json:"items"`
			Total int              `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if payload.Data.Total == 0 || len(payload.Data.Items) == 0 {
		t.Fatalf("expected admin list to return seeded forecast runs, got %+v", payload.Data)
	}
}

func TestAdminRetryForecastL3RunOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	growthRepo := repo.NewInMemoryGrowthRepo()
	growthService := service.NewGrowthService(growthRepo)
	adminHandlers := NewAdminHandlers(growthService, config.Config{})

	run, err := growthRepo.CreateStrategyForecastL3Run(repoRunInput("FUTURES", "RB2609", "admin_001"))
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}
	if _, err := growthRepo.CancelStrategyForecastL3Run(run.ID, "admin_001", "stop first"); err != nil {
		t.Fatalf("cancel run before retry: %v", err)
	}

	router := gin.New()
	attachUserID(router, "admin_001")
	router.POST("/api/v1/admin/forecast/runs/:id/retry", adminHandlers.Forecast.RetryForecastL3Run)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/forecast/runs/"+run.ID+"/retry", bytes.NewBufferString(`{"reason":"rerun after new data"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if payload.Data.ID != run.ID || payload.Data.Status != "QUEUED" {
		t.Fatalf("expected retried run to return queued status, got %+v", payload.Data)
	}
}

func repoRunInput(targetType string, targetKey string, userID string) model.StrategyForecastL3RunCreateInput {
	return model.StrategyForecastL3RunCreateInput{
		TargetType:     targetType,
		TargetID:       targetKey,
		TargetKey:      targetKey,
		TargetLabel:    targetKey,
		TriggerType:    "ADMIN_MANUAL",
		RequestUserID:  userID,
		OperatorUserID: userID,
		PriorityScore:  0.8,
		Reason:         "seeded handler test run",
		Source:         "ADMIN_CONSOLE",
		SourceID:       "seed-admin",
		SourcePath:     "/admin/forecast-lab",
	}
}

func seedForecastHistoryHandlerRuns(t *testing.T, growthRepo *repo.InMemoryGrowthRepo, targetKey string, targetType string) {
	t.Helper()

	for index, scenario := range []string{"base", "bull", "bear"} {
		run, err := growthRepo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
			TargetType:    targetType,
			TargetID:      targetKey,
			TargetKey:     targetKey,
			TargetLabel:   targetKey,
			Source:        "RECOMMENDATION",
			SourceID:      "seed-history",
			SourcePath:    "/recommendations",
			TriggerType:   model.StrategyForecastL3TriggerTypeUserRequest,
			RequestUserID: "user_001",
			PriorityScore: 0.6 + float64(index)*0.1,
			Reason:        "seed history handler",
		})
		if err != nil {
			t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
		}
		growthRepo.ExecuteQueuedStrategyForecastL3Runs(1, "system")
		detail, err := growthRepo.GetStrategyForecastL3RunDetail(run.ID)
		if err != nil {
			t.Fatalf("GetStrategyForecastL3RunDetail() error = %v", err)
		}
		growthRepo.RunStrategyForecastL3QualityBackfill(20, "system")
		detail.Run.Summary.PrimaryScenario = scenario
		_ = detail
	}
}
