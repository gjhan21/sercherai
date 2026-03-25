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

func newUserGrowthHandlerForTest(t *testing.T) *UserGrowthHandler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return NewUserGrowthHandler(service.NewGrowthService(repo.NewInMemoryGrowthRepo()), config.Config{})
}

func newAdminGrowthHandlerForTest(t *testing.T) *AdminGrowthHandler {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return NewAdminGrowthHandler(service.NewGrowthService(repo.NewInMemoryGrowthRepo()), config.Config{})
}

func attachUserID(router *gin.Engine, userID string) {
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
}

func TestListPublicCommunityTopicsOK(t *testing.T) {
	growthHandler := newUserGrowthHandlerForTest(t)
	router := gin.New()
	router.GET("/api/v1/public/community/topics", growthHandler.ListPublicCommunityTopics)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/public/community/topics?topic_type=STOCK", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got := int(payload["code"].(float64)); got != 0 {
		t.Fatalf("expected code 0, got %d", got)
	}
}

func TestCreateCommunityTopicRequiresAuth(t *testing.T) {
	growthHandler := newUserGrowthHandlerForTest(t)
	router := gin.New()
	router.POST("/api/v1/community/topics", growthHandler.CreateCommunityTopic)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/community/topics",
		bytes.NewBufferString(`{"title":"测试观点","topic_type":"STOCK","stance":"WATCH","target_type":"STOCK","target_id":"600519"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestCreateCommunityCommentCreatesVisibleReplyNotification(t *testing.T) {
	growthRepo := repo.NewInMemoryGrowthRepo()
	growthService := service.NewGrowthService(growthRepo)

	_, err := growthService.CreateCommunityComment(model.CommunityCommentCreateInput{
		UserID:  "u_demo_002",
		TopicID: "ct_demo_001",
		Content: "继续观察量价结构。",
	})
	if err != nil {
		t.Fatalf("CreateCommunityComment() error = %v", err)
	}

	items, total, err := growthRepo.ListMessages("u_demo_001", 1, 20)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if total < 2 || len(items) < 2 {
		t.Fatalf("expected community notification to be visible in messages, got total=%d len=%d", total, len(items))
	}

	found := false
	for _, item := range items {
		if item.Type == "COMMUNITY" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected community notification message to be returned")
	}
}

func TestAdminListCommunityTopicsOK(t *testing.T) {
	adminHandler := newAdminGrowthHandlerForTest(t)
	router := gin.New()
	router.GET("/api/v1/admin/community/topics", adminHandler.ListCommunityTopics)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/community/topics?status=PUBLISHED", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if got := int(payload["code"].(float64)); got != 0 {
		t.Fatalf("expected code 0, got %d", got)
	}
}

func TestAdminReviewCommunityReportOK(t *testing.T) {
	adminHandler := newAdminGrowthHandlerForTest(t)
	router := gin.New()
	router.PUT("/api/v1/admin/community/reports/:id/review", adminHandler.ReviewCommunityReport)

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/admin/community/reports/cr_demo_001/review",
		bytes.NewBufferString(`{"status":"RESOLVED","review_note":"已记录并处理"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAdminListCommunityCommentsIncludesTopicContext(t *testing.T) {
	adminHandler := newAdminGrowthHandlerForTest(t)
	router := gin.New()
	router.GET("/api/v1/admin/community/comments", adminHandler.ListCommunityComments)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/community/comments", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Items []model.CommunityComment `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if len(payload.Data.Items) == 0 {
		t.Fatal("expected comment rows")
	}
	item := payload.Data.Items[0]
	if item.TopicTitle == "" {
		t.Fatal("expected topic title in admin comment context")
	}
	if item.TopicSummary == "" {
		t.Fatal("expected topic summary in admin comment context")
	}
	if item.LinkedTarget.TargetSnapshot == "" {
		t.Fatal("expected linked target snapshot in admin comment context")
	}
}

func TestAdminListCommunityReportsIncludesTargetContext(t *testing.T) {
	adminHandler := newAdminGrowthHandlerForTest(t)
	router := gin.New()
	router.GET("/api/v1/admin/community/reports", adminHandler.ListCommunityReports)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/community/reports", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Items []model.CommunityReport `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if len(payload.Data.Items) == 0 {
		t.Fatal("expected report rows")
	}
	item := payload.Data.Items[0]
	if item.TopicTitle == "" {
		t.Fatal("expected topic title in admin report context")
	}
	if item.TargetContent == "" {
		t.Fatal("expected target content in admin report context")
	}
	if item.LinkedTarget.TargetSnapshot == "" {
		t.Fatal("expected linked target snapshot in admin report context")
	}
}

func TestListMyCommunityTopicsReturnsOnlyCurrentUserTopics(t *testing.T) {
	growthHandler := newUserGrowthHandlerForTest(t)
	router := gin.New()
	attachUserID(router, "u_demo_002")
	router.GET("/api/v1/community/me/topics", growthHandler.ListMyCommunityTopics)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/community/me/topics", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Items []model.CommunityTopicListItem `json:"items"`
			Total int                           `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if payload.Data.Total == 0 || len(payload.Data.Items) == 0 {
		t.Fatal("expected my topic rows")
	}
	for _, item := range payload.Data.Items {
		if item.UserID != "u_demo_002" {
			t.Fatalf("expected only current user topics, got %q", item.UserID)
		}
	}
}

func TestListMyCommunityCommentsReturnsOnlyCurrentUserComments(t *testing.T) {
	growthHandler := newUserGrowthHandlerForTest(t)
	router := gin.New()
	attachUserID(router, "u_demo_002")
	router.GET("/api/v1/community/me/comments", growthHandler.ListMyCommunityComments)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/community/me/comments", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var payload struct {
		Code int `json:"code"`
		Data struct {
			Items []model.CommunityComment `json:"items"`
			Total int                     `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if payload.Data.Total == 0 || len(payload.Data.Items) == 0 {
		t.Fatal("expected my comment rows")
	}
	for _, item := range payload.Data.Items {
		if item.UserID != "u_demo_002" {
			t.Fatalf("expected only current user comments, got %q", item.UserID)
		}
	}
}
