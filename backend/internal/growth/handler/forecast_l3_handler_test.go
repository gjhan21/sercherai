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

	req := httptest.NewRequest(http.MethodPost, "/api/v1/forecast/runs", bytes.NewBufferString(`{"target_type":"STOCK","target_key":"600519.SH","target_label":"贵州茅台","priority_score":0.81,"reason":"need deeper view"}`))
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

func TestAdminListForecastL3RunsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)
	growthRepo := repo.NewInMemoryGrowthRepo()
	growthService := service.NewGrowthService(growthRepo)
	adminHandler := NewAdminGrowthHandler(growthService, config.Config{})

	if _, err := growthRepo.CreateStrategyForecastL3Run(repoRunInput("STOCK", "600519.SH", "admin_001")); err != nil {
		t.Fatalf("seed run: %v", err)
	}

	router := gin.New()
	router.GET("/api/v1/admin/forecast/runs", adminHandler.ListForecastL3Runs)

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
	adminHandler := NewAdminGrowthHandler(growthService, config.Config{})

	run, err := growthRepo.CreateStrategyForecastL3Run(repoRunInput("FUTURES", "RB2609", "admin_001"))
	if err != nil {
		t.Fatalf("seed run: %v", err)
	}
	if _, err := growthRepo.CancelStrategyForecastL3Run(run.ID, "admin_001", "stop first"); err != nil {
		t.Fatalf("cancel run before retry: %v", err)
	}

	router := gin.New()
	attachUserID(router, "admin_001")
	router.POST("/api/v1/admin/forecast/runs/:id/retry", adminHandler.RetryForecastL3Run)

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
		TargetKey:      targetKey,
		TargetLabel:    targetKey,
		TriggerType:    "ADMIN_MANUAL",
		RequestUserID:  userID,
		OperatorUserID: userID,
		PriorityScore:  0.8,
		Reason:         "seeded handler test run",
	}
}
