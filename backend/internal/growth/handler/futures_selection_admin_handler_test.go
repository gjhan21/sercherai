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

func newFuturesSelectionTestHandler() *AdminGrowthHandler {
	return NewAdminGrowthHandler(service.NewGrowthService(repo.NewInMemoryGrowthRepo()), config.Config{})
}

func TestGetFuturesSelectionOverview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/futures-selection/overview", nil)

	newFuturesSelectionTestHandler().GetFuturesSelectionOverview(ctx)

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

func TestCreateAndApproveFuturesSelectionRun(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newFuturesSelectionTestHandler()

	runRecorder := httptest.NewRecorder()
	runCtx, _ := gin.CreateTestContext(runRecorder)
	runCtx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/futures-selection/runs", strings.NewReader(`{"trade_date":"2026-03-21"}`))
	runCtx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateFuturesSelectionRun(runCtx)

	if runRecorder.Code != http.StatusOK {
		t.Fatalf("expected create run 200, got %d", runRecorder.Code)
	}

	reviewRecorder := httptest.NewRecorder()
	reviewCtx, _ := gin.CreateTestContext(reviewRecorder)
	reviewCtx.Params = gin.Params{{Key: "run_id", Value: "fsr_demo_001"}}
	reviewCtx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/futures-selection/reviews/fsr_demo_001/approve", strings.NewReader(`{"review_note":"通过"}`))
	reviewCtx.Request.Header.Set("Content-Type", "application/json")

	handler.ApproveFuturesSelectionReview(reviewCtx)

	if reviewRecorder.Code != http.StatusOK {
		t.Fatalf("expected approve review 200, got %d", reviewRecorder.Code)
	}
}

func TestListFuturesSelectionProfilesAndEvaluation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newFuturesSelectionTestHandler()

	profileRecorder := httptest.NewRecorder()
	profileCtx, _ := gin.CreateTestContext(profileRecorder)
	profileCtx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/futures-selection/profiles", nil)

	handler.ListFuturesSelectionProfiles(profileCtx)

	if profileRecorder.Code != http.StatusOK {
		t.Fatalf("expected list profiles 200, got %d", profileRecorder.Code)
	}

	evaluationRecorder := httptest.NewRecorder()
	evaluationCtx, _ := gin.CreateTestContext(evaluationRecorder)
	evaluationCtx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/futures-selection/evaluation/leaderboard", nil)

	handler.ListFuturesSelectionEvaluationLeaderboard(evaluationCtx)

	if evaluationRecorder.Code != http.StatusOK {
		t.Fatalf("expected evaluation leaderboard 200, got %d", evaluationRecorder.Code)
	}
}

func TestListFuturesSelectionProfileVersions(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newFuturesSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "profile_default_futures_auto"}}
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/futures-selection/profiles/profile_default_futures_auto/versions", nil)

	handler.ListFuturesSelectionProfileVersions(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected list profile versions 200, got %d", recorder.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := payload["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected data payload, got %#v", payload["data"])
	}
	items, ok := data["items"].([]any)
	if !ok || len(items) == 0 {
		t.Fatalf("expected non-empty versions list, got %#v", data["items"])
	}
}

func TestListCreateAndSetDefaultFuturesSelectionTemplates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newFuturesSelectionTestHandler()

	listRecorder := httptest.NewRecorder()
	listCtx, _ := gin.CreateTestContext(listRecorder)
	listCtx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/futures-selection/templates", nil)

	handler.ListFuturesSelectionProfileTemplates(listCtx)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected list templates 200, got %d", listRecorder.Code)
	}

	createRecorder := httptest.NewRecorder()
	createCtx, _ := gin.CreateTestContext(createRecorder)
	createCtx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/futures-selection/templates", strings.NewReader(`{
	  "template_key":"MANUAL_TREND",
	  "name":"手工趋势模板",
	  "description":"用于测试期货模板创建",
	  "market_regime_bias":"TREND_CONTINUE",
	  "status":"ACTIVE",
	  "is_default":false,
	  "universe_defaults_json":{"style":"trend","contract_scope":"MANUAL","contracts":["IH2506"],"allow_mock_fallback_on_short_history":true},
	  "factor_defaults_json":{"min_confidence":66},
	  "portfolio_defaults_json":{"limit":2,"max_risk_level":"MEDIUM"},
	  "publish_defaults_json":{"review_required":true,"allow_auto_publish":false}
	}`))
	createCtx.Request.Header.Set("Content-Type", "application/json")

	handler.CreateFuturesSelectionProfileTemplate(createCtx)

	if createRecorder.Code != http.StatusOK {
		t.Fatalf("expected create template 200, got %d", createRecorder.Code)
	}

	setDefaultRecorder := httptest.NewRecorder()
	setDefaultCtx, _ := gin.CreateTestContext(setDefaultRecorder)
	setDefaultCtx.Params = gin.Params{{Key: "id", Value: "fstpl_demo_001"}}
	setDefaultCtx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/admin/futures-selection/templates/fstpl_demo_001/set-default", nil)

	handler.SetDefaultFuturesSelectionProfileTemplate(setDefaultCtx)

	if setDefaultRecorder.Code != http.StatusOK {
		t.Fatalf("expected set default template 200, got %d", setDefaultRecorder.Code)
	}
}

func TestCompareFuturesSelectionRuns(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newFuturesSelectionTestHandler()

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/v1/admin/futures-selection/runs/compare?run_ids=fsr_demo_001", nil)

	handler.CompareFuturesSelectionRuns(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected compare runs 200, got %d", recorder.Code)
	}
}
