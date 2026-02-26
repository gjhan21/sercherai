package handler

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/config"
)

type AdminGrowthHandler struct {
	service service.GrowthService
	cfg     config.Config
}

func NewAdminGrowthHandler(service service.GrowthService, cfg config.Config) *AdminGrowthHandler {
	return &AdminGrowthHandler{service: service, cfg: cfg}
}

func (h *AdminGrowthHandler) ListInviteRecords(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")

	items, total, err := h.service.AdminListInviteRecords(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ListRewardRecords(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")

	items, total, err := h.service.AdminListRewardRecords(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ReviewRewardRecord(c *gin.Context) {
	id := c.Param("id")
	var req dto.ReviewRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminReviewRewardRecord(id, req.Status, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListReconciliation(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListReconciliation(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) RetryReconciliation(c *gin.Context) {
	batchID := c.Param("batch_id")
	if err := h.service.AdminRetryReconciliation(batchID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListRiskRules(c *gin.Context) {
	items, err := h.service.AdminListRiskRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) CreateRiskRule(c *gin.Context) {
	var req dto.RiskRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateRiskRule(req.RuleCode, req.RuleName, req.Threshold, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "RISK", "CREATE_RULE", "RISK_RULE", id, "", req.Status, req.RuleCode)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateRiskRule(c *gin.Context) {
	id := c.Param("id")
	var req dto.RiskRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateRiskRule(id, req.Threshold, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListRiskHits(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListRiskHits(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ReviewRiskHit(c *gin.Context) {
	id := c.Param("id")
	var req dto.ReviewRiskHitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminReviewRiskHit(id, req.Status, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListWithdrawRequests(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListWithdrawRequests(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ReviewWithdrawRequest(c *gin.Context) {
	id := c.Param("id")
	var req dto.ReviewWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminReviewWithdrawRequest(id, req.Status, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListNewsCategories(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListNewsCategories(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateNewsCategory(c *gin.Context) {
	var req dto.NewsCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateNewsCategory(req.Name, req.Slug, req.Sort, req.Visibility, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "NEWS", "CREATE_CATEGORY", "NEWS_CATEGORY", id, "", req.Status, req.Slug)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateNewsCategory(c *gin.Context) {
	id := c.Param("id")
	var req dto.NewsCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateNewsCategory(id, req.Name, req.Slug, req.Sort, req.Visibility, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListNewsArticles(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	categoryID := c.Query("category_id")
	items, total, err := h.service.AdminListNewsArticles(status, categoryID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateNewsArticle(c *gin.Context) {
	var req dto.NewsArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator, _ := c.Get("user_id")
	authorID, _ := operator.(string)
	if authorID == "" {
		authorID = "admin_unknown"
	}
	id, err := h.service.AdminCreateNewsArticle(req.CategoryID, req.Title, req.Summary, req.Content, req.Visibility, req.Status, authorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateNewsArticle(c *gin.Context) {
	id := c.Param("id")
	var req dto.NewsArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateNewsArticle(id, req.CategoryID, req.Title, req.Summary, req.Content, req.Visibility, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) PublishNewsArticle(c *gin.Context) {
	id := c.Param("id")
	var req dto.NewsPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminPublishNewsArticle(id, req.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "article not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "NEWS", "PUBLISH_ARTICLE", "NEWS_ARTICLE", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) CreateNewsAttachment(c *gin.Context) {
	articleID := c.Param("id")
	var req dto.NewsAttachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateNewsAttachment(articleID, req.FileName, req.FileURL, req.FileSize, req.MimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) ListNewsAttachments(c *gin.Context) {
	articleID := c.Param("id")
	items, err := h.service.AdminListNewsAttachments(articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) DeleteNewsAttachment(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.AdminDeleteNewsAttachment(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40402, Message: "attachment not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "NEWS", "DELETE_ATTACHMENT", "NEWS_ATTACHMENT", id, "", "DELETED", "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListStockRecommendations(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListStockRecommendations(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateStockRecommendation(c *gin.Context) {
	var req dto.StockRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateStockRecommendation(model.StockRecommendation{
		Symbol:        req.Symbol,
		Name:          req.Name,
		Score:         req.Score,
		RiskLevel:     req.RiskLevel,
		PositionRange: req.PositionRange,
		ValidFrom:     req.ValidFrom,
		ValidTo:       req.ValidTo,
		Status:        req.Status,
		ReasonSummary: req.ReasonSummary,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateStockRecommendationStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.StockRecommendationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateStockRecommendationStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "STOCK", "UPDATE_RECOMMENDATION_STATUS", "STOCK_RECOMMENDATION", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) GenerateDailyStockRecommendations(c *gin.Context) {
	tradeDate := c.Query("trade_date")
	count, err := h.service.AdminGenerateDailyStockRecommendations(tradeDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "STOCK", "GENERATE_DAILY_RECOMMENDATIONS", "BATCH", tradeDate, "", "count="+strconv.Itoa(count), "")
	c.JSON(http.StatusOK, dto.OK(gin.H{"count": count}))
}

func (h *AdminGrowthHandler) ListFuturesStrategies(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	contract := c.Query("contract")
	items, total, err := h.service.AdminListFuturesStrategies(status, contract, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateFuturesStrategy(c *gin.Context) {
	var req dto.FuturesStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateFuturesStrategy(model.FuturesStrategy{
		Contract:      req.Contract,
		Name:          req.Name,
		Direction:     req.Direction,
		RiskLevel:     req.RiskLevel,
		PositionRange: req.PositionRange,
		ValidFrom:     req.ValidFrom,
		ValidTo:       req.ValidTo,
		Status:        req.Status,
		ReasonSummary: req.ReasonSummary,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "FUTURES", "CREATE_STRATEGY", "FUTURES_STRATEGY", id, "", req.Status, req.Contract)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateFuturesStrategyStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.FuturesStrategyStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateFuturesStrategyStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "FUTURES", "UPDATE_STRATEGY_STATUS", "FUTURES_STRATEGY", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListUsers(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	kycStatus := c.Query("kyc_status")
	memberLevel := c.Query("member_level")
	items, total, err := h.service.AdminListUsers(status, kycStatus, memberLevel, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ExportUsersCSV(c *gin.Context) {
	status := c.Query("status")
	kycStatus := c.Query("kyc_status")
	memberLevel := c.Query("member_level")
	items, _, err := h.service.AdminListUsers(status, kycStatus, memberLevel, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "phone", "email", "status", "kyc_status", "member_level", "created_at"})
	for _, it := range items {
		_ = writer.Write([]string{it.ID, it.Phone, it.Email, it.Status, it.KYCStatus, it.MemberLevel, it.CreatedAt})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=admin_users.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminGrowthHandler) UpdateUserStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateUserStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "USER", "UPDATE_STATUS", "USER", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) UpdateUserMemberLevel(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUserMemberLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateUserMemberLevel(id, req.MemberLevel); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "USER", "UPDATE_MEMBER_LEVEL", "USER", id, "", req.MemberLevel, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) UpdateUserKYCStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUserKYCStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateUserKYCStatus(id, req.KYCStatus); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "USER", "UPDATE_KYC_STATUS", "USER", id, "", req.KYCStatus, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) DashboardOverview(c *gin.Context) {
	item, err := h.service.AdminDashboardOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) ListOperationLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
	module := c.Query("module")
	action := c.Query("action")
	operator := c.Query("operator_user_id")
	items, total, err := h.service.AdminListOperationLogs(module, action, operator, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ExportOperationLogsCSV(c *gin.Context) {
	module := c.Query("module")
	action := c.Query("action")
	operator := c.Query("operator_user_id")
	items, _, err := h.service.AdminListOperationLogs(module, action, operator, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "module", "action", "target_type", "target_id", "operator_user_id", "before_value", "after_value", "reason", "created_at"})
	for _, it := range items {
		_ = writer.Write([]string{it.ID, it.Module, it.Action, it.TargetType, it.TargetID, it.OperatorUserID, it.BeforeValue, it.AfterValue, it.Reason, it.CreatedAt})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=admin_operation_logs.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminGrowthHandler) ListMembershipProducts(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListMembershipProducts(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateMembershipProduct(c *gin.Context) {
	var req dto.MembershipProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateMembershipProduct(req.Name, req.Price, req.Status, req.MemberLevel, req.DurationDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "CREATE_PRODUCT", "MEMBERSHIP_PRODUCT", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateMembershipProductStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.MembershipProductStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateMembershipProductStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "UPDATE_PRODUCT_STATUS", "MEMBERSHIP_PRODUCT", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListMembershipOrders(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	userID := c.Query("user_id")
	items, total, err := h.service.AdminListMembershipOrders(status, userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ExportMembershipOrdersCSV(c *gin.Context) {
	status := c.Query("status")
	userID := c.Query("user_id")
	items, _, err := h.service.AdminListMembershipOrders(status, userID, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "user_id", "product_id", "amount", "status", "paid_at", "created_at"})
	for _, it := range items {
		_ = writer.Write([]string{
			it.ID, it.UserID, it.ProductID,
			strconv.FormatFloat(it.Amount, 'f', -1, 64),
			it.Status, it.PaidAt, it.CreatedAt,
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=membership_orders.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminGrowthHandler) UpdateMembershipOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.MembershipOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateMembershipOrderStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "UPDATE_ORDER_STATUS", "MEMBERSHIP_ORDER", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListVIPQuotaConfigs(c *gin.Context) {
	page, pageSize := parsePage(c)
	memberLevel := c.Query("member_level")
	status := c.Query("status")
	items, total, err := h.service.AdminListVIPQuotaConfigs(memberLevel, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateVIPQuotaConfig(c *gin.Context) {
	var req dto.VIPQuotaConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if _, err := time.Parse(time.RFC3339, req.EffectiveAt); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "effective_at must be RFC3339 format", Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateVIPQuotaConfig(model.VIPQuotaConfig{
		MemberLevel:        req.MemberLevel,
		DocReadLimit:       req.DocReadLimit,
		NewsSubscribeLimit: req.NewsSubscribeLimit,
		ResetCycle:         req.ResetCycle,
		Status:             req.Status,
		EffectiveAt:        req.EffectiveAt,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "CREATE_VIP_QUOTA_CONFIG", "VIP_QUOTA_CONFIG", id, "", req.MemberLevel, req.EffectiveAt)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateVIPQuotaConfig(c *gin.Context) {
	id := c.Param("id")
	var req dto.VIPQuotaConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if _, err := time.Parse(time.RFC3339, req.EffectiveAt); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "effective_at must be RFC3339 format", Data: struct{}{}})
		return
	}
	err := h.service.AdminUpdateVIPQuotaConfig(id, model.VIPQuotaConfig{
		DocReadLimit:       req.DocReadLimit,
		NewsSubscribeLimit: req.NewsSubscribeLimit,
		ResetCycle:         req.ResetCycle,
		Status:             req.Status,
		EffectiveAt:        req.EffectiveAt,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "quota config not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "UPDATE_VIP_QUOTA_CONFIG", "VIP_QUOTA_CONFIG", id, "", req.Status, req.EffectiveAt)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListUserQuotas(c *gin.Context) {
	page, pageSize := parsePage(c)
	userID := strings.TrimSpace(c.Query("user_id"))
	periodKey := strings.TrimSpace(c.Query("period_key"))
	items, total, err := h.service.AdminListUserQuotaUsages(userID, periodKey, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) AdjustUserQuota(c *gin.Context) {
	userID := c.Param("user_id")
	var req dto.UserQuotaAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminAdjustUserQuota(userID, req.PeriodKey, req.DocReadDelta, req.NewsSubscribeDelta); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "user not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	summary := fmt.Sprintf("period=%s,doc_delta=%d,news_delta=%d", req.PeriodKey, req.DocReadDelta, req.NewsSubscribeDelta)
	h.writeOperationLog(c, "MEMBERSHIP", "ADJUST_USER_QUOTA", "USER_QUOTA", userID, "", summary, req.Reason)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListDataSources(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListDataSources(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateDataSource(c *gin.Context) {
	var req dto.DataSourceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateDataSource(model.DataSource{
		SourceKey:  req.SourceKey,
		Name:       req.Name,
		SourceType: req.SourceType,
		Status:     req.Status,
		Config:     req.Config,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "exists") {
			c.JSON(http.StatusConflict, dto.APIResponse{Code: 40901, Message: err.Error(), Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "CREATE_DATA_SOURCE", "DATA_SOURCE", id, "", req.Status, req.SourceKey)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateDataSource(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	var req dto.DataSourceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	err := h.service.AdminUpdateDataSource(sourceKey, model.DataSource{
		Name:       req.Name,
		SourceType: req.SourceType,
		Status:     req.Status,
		Config:     req.Config,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "UPDATE_DATA_SOURCE", "DATA_SOURCE", sourceKey, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) DeleteDataSource(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	if err := h.service.AdminDeleteDataSource(sourceKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "DELETE_DATA_SOURCE", "DATA_SOURCE", sourceKey, "", "DELETED", "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) CheckDataSourceHealth(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	item, err := h.service.AdminCheckDataSourceHealth(sourceKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "CHECK_DATA_SOURCE_HEALTH", "DATA_SOURCE", sourceKey, "", item.Status, item.Message)
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) ListSystemConfigs(c *gin.Context) {
	page, pageSize := parsePage(c)
	keyword := c.Query("keyword")
	items, total, err := h.service.AdminListSystemConfigs(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) UpsertSystemConfig(c *gin.Context) {
	var req dto.SystemConfigUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpsertSystemConfig(req.ConfigKey, req.ConfigValue, req.Description, operator); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "UPSERT_CONFIG", "SYSTEM_CONFIG", req.ConfigKey, "", req.ConfigValue, req.Description)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListReviewTasks(c *gin.Context) {
	page, pageSize := parsePage(c)
	module := c.Query("module")
	status := c.Query("status")
	submitterID := c.Query("submitter_id")
	reviewerID := c.Query("reviewer_id")
	items, total, err := h.service.AdminListReviewTasks(module, status, submitterID, reviewerID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ExportReviewTasksCSV(c *gin.Context) {
	module := c.Query("module")
	status := c.Query("status")
	submitterID := c.Query("submitter_id")
	reviewerID := c.Query("reviewer_id")
	items, _, err := h.service.AdminListReviewTasks(module, status, submitterID, reviewerID, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "module", "target_id", "submitter_id", "reviewer_id", "status", "submit_note", "review_note", "submitted_at", "reviewed_at"})
	for _, it := range items {
		_ = writer.Write([]string{it.ID, it.Module, it.TargetID, it.SubmitterID, it.ReviewerID, it.Status, it.SubmitNote, it.ReviewNote, it.SubmittedAt, it.ReviewedAt})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=review_tasks.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminGrowthHandler) WorkflowMetrics(c *gin.Context) {
	module := c.Query("module")
	receiverID := strings.TrimSpace(c.Query("receiver_id"))
	if receiverID == "" {
		operatorVal, _ := c.Get("user_id")
		receiverID, _ = operatorVal.(string)
	}
	item, err := h.service.AdminGetWorkflowMetrics(module, receiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) SubmitReviewTask(c *gin.Context) {
	var req dto.ReviewSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	id, err := h.service.AdminSubmitReviewTask(req.Module, req.TargetID, operator, req.ReviewerID, req.SubmitNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "WORKFLOW", "SUBMIT_REVIEW", strings.ToUpper(req.Module), req.TargetID, "", "REVIEWING", req.SubmitNote)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) ReviewTaskDecision(c *gin.Context) {
	id := c.Param("id")
	var req dto.ReviewDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminReviewTaskDecision(id, req.Status, operator, req.ReviewNote); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "WORKFLOW", "REVIEW_DECISION", "REVIEW_TASK", id, "PENDING", strings.ToUpper(req.Status), req.ReviewNote)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) AssignReviewTask(c *gin.Context) {
	id := c.Param("id")
	var req dto.ReviewAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminAssignReviewTask(id, req.ReviewerID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "WORKFLOW", "ASSIGN_REVIEW", "REVIEW_TASK", id, "", req.ReviewerID, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListSchedulerJobRuns(c *gin.Context) {
	page, pageSize := parsePage(c)
	jobName := c.Query("job_name")
	status := c.Query("status")
	items, total, err := h.service.AdminListSchedulerJobRuns(jobName, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ExportSchedulerJobRunsCSV(c *gin.Context) {
	jobName := c.Query("job_name")
	status := c.Query("status")
	items, _, err := h.service.AdminListSchedulerJobRuns(jobName, status, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "parent_run_id", "retry_count", "job_name", "trigger_source", "status", "started_at", "finished_at", "result_summary", "error_message", "operator_id"})
	for _, it := range items {
		_ = writer.Write([]string{it.ID, it.ParentRunID, strconv.Itoa(it.RetryCount), it.JobName, it.TriggerSource, it.Status, it.StartedAt, it.FinishedAt, it.ResultSummary, it.ErrorMessage, it.OperatorID})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=scheduler_job_runs.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminGrowthHandler) SchedulerJobMetrics(c *gin.Context) {
	jobName := c.Query("job_name")
	item, err := h.service.AdminGetSchedulerJobMetrics(jobName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) ListSchedulerJobDefinitions(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	module := c.Query("module")
	items, total, err := h.service.AdminListSchedulerJobDefinitions(status, module, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateSchedulerJobDefinition(c *gin.Context) {
	var req dto.SchedulerJobDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	id, err := h.service.AdminCreateSchedulerJobDefinition(model.SchedulerJobDefinition{
		JobName:     req.JobName,
		DisplayName: req.DisplayName,
		Module:      req.Module,
		CronExpr:    req.CronExpr,
		Status:      req.Status,
	}, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "CREATE_JOB_DEFINITION", "JOB_DEFINITION", id, "", req.Status, req.JobName)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateSchedulerJobDefinition(c *gin.Context) {
	id := c.Param("id")
	var req dto.SchedulerJobDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpdateSchedulerJobDefinition(id, model.SchedulerJobDefinition{
		JobName:     req.JobName,
		DisplayName: req.DisplayName,
		Module:      req.Module,
		CronExpr:    req.CronExpr,
		Status:      req.Status,
	}, operator); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "UPDATE_JOB_DEFINITION", "JOB_DEFINITION", id, "", req.Status, req.CronExpr)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) UpdateSchedulerJobDefinitionStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.SchedulerJobDefinitionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpdateSchedulerJobDefinitionStatus(id, req.Status, operator); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "UPDATE_JOB_DEFINITION_STATUS", "JOB_DEFINITION", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) TriggerSchedulerJob(c *gin.Context) {
	var req dto.SchedulerTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	simulateStatus := strings.ToUpper(strings.TrimSpace(req.SimulateStatus))
	if simulateStatus != "" && h.cfg.AllowJobSimulation {
		id, err := h.service.AdminCreateSchedulerJobRun(req.JobName, req.TriggerSource, simulateStatus, req.ResultSummary, req.ErrorMessage, operator)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		h.writeOperationLog(c, "SCHEDULER", "TRIGGER_JOB", "JOB", req.JobName, "", simulateStatus, req.TriggerSource)
		c.JSON(http.StatusOK, dto.OK(gin.H{"id": id, "status": simulateStatus}))
		return
	}
	resultSummary, err := h.runSchedulerJob(req.JobName)
	status := "SUCCESS"
	errorMessage := ""
	if err != nil {
		status = "FAILED"
		errorMessage = err.Error()
	}
	id, err := h.service.AdminCreateSchedulerJobRun(req.JobName, req.TriggerSource, status, resultSummary, errorMessage, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "TRIGGER_JOB", "JOB", req.JobName, "", status, req.TriggerSource)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id, "status": status}))
}

func (h *AdminGrowthHandler) RetrySchedulerJobRun(c *gin.Context) {
	runID := c.Param("id")
	var req dto.SchedulerRetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	simulateStatus := strings.ToUpper(strings.TrimSpace(req.SimulateStatus))
	if simulateStatus != "" && h.cfg.AllowJobSimulation {
		id, err := h.service.AdminRetrySchedulerJobRun(runID, simulateStatus, req.ResultSummary, req.ErrorMessage, operator)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		h.writeOperationLog(c, "SCHEDULER", "RETRY_JOB", "JOB_RUN", runID, "", simulateStatus, req.ResultSummary)
		c.JSON(http.StatusOK, dto.OK(gin.H{"id": id, "status": simulateStatus}))
		return
	}
	jobName, err := h.service.GetSchedulerJobNameByRunID(runID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "job run not found", Data: struct{}{}})
		return
	}
	resultSummary, runErr := h.runSchedulerJob(jobName)
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	id, err := h.service.AdminRetrySchedulerJobRun(runID, status, resultSummary, errorMessage, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "RETRY_JOB", "JOB_RUN", runID, "", status, resultSummary)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id, "status": status}))
}

func (h *AdminGrowthHandler) ListWorkflowMessages(c *gin.Context) {
	page, pageSize := parsePage(c)
	module := c.Query("module")
	eventType := c.Query("event_type")
	isRead := c.Query("is_read")
	receiverID := strings.TrimSpace(c.Query("receiver_id"))
	if receiverID == "" {
		operatorVal, _ := c.Get("user_id")
		receiverID, _ = operatorVal.(string)
	}
	items, total, err := h.service.AdminListWorkflowMessages(module, eventType, isRead, receiverID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ExportWorkflowMessagesCSV(c *gin.Context) {
	module := c.Query("module")
	eventType := c.Query("event_type")
	isRead := c.Query("is_read")
	receiverID := strings.TrimSpace(c.Query("receiver_id"))
	if receiverID == "" {
		operatorVal, _ := c.Get("user_id")
		receiverID, _ = operatorVal.(string)
	}
	items, _, err := h.service.AdminListWorkflowMessages(module, eventType, isRead, receiverID, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "review_id", "target_id", "module", "receiver_id", "sender_id", "event_type", "title", "content", "is_read", "created_at", "read_at"})
	for _, it := range items {
		isReadVal := "false"
		if it.IsRead {
			isReadVal = "true"
		}
		_ = writer.Write([]string{it.ID, it.ReviewID, it.TargetID, it.Module, it.ReceiverID, it.SenderID, it.EventType, it.Title, it.Content, isReadVal, it.CreatedAt, it.ReadAt})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=workflow_messages.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminGrowthHandler) UpdateWorkflowMessageRead(c *gin.Context) {
	id := c.Param("id")
	var req dto.WorkflowMessageReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateWorkflowMessageRead(id, req.IsRead); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) CountUnreadWorkflowMessages(c *gin.Context) {
	module := c.Query("module")
	eventType := c.Query("event_type")
	receiverID := strings.TrimSpace(c.Query("receiver_id"))
	if receiverID == "" {
		operatorVal, _ := c.Get("user_id")
		receiverID, _ = operatorVal.(string)
	}
	total, err := h.service.AdminCountUnreadWorkflowMessages(module, eventType, receiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"unread_count": total}))
}

func (h *AdminGrowthHandler) BulkReadWorkflowMessages(c *gin.Context) {
	var req dto.WorkflowMessageBulkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	receiverID := strings.TrimSpace(req.ReceiverID)
	if receiverID == "" {
		operatorVal, _ := c.Get("user_id")
		receiverID, _ = operatorVal.(string)
	}
	affected, err := h.service.AdminBulkReadWorkflowMessages(req.Module, req.EventType, receiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"affected": affected}))
}

func (h *AdminGrowthHandler) writeOperationLog(c *gin.Context, module string, action string, targetType string, targetID string, beforeValue string, afterValue string, reason string) {
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = "admin_unknown"
	}
	_ = h.service.AdminCreateOperationLog(module, action, targetType, targetID, operator, beforeValue, afterValue, reason)
}

func (h *AdminGrowthHandler) runSchedulerJob(jobName string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(jobName)) {
	case "daily_stock_recommendation":
		tradeDate := time.Now().Format("2006-01-02")
		count, err := h.service.AdminGenerateDailyStockRecommendations(tradeDate)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("generated %d recommendations", count), nil
	case "daily_futures_strategy":
		tradeDate := time.Now().Format("2006-01-02")
		count, err := h.service.AdminGenerateDailyFuturesStrategies(tradeDate)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("generated %d futures strategies", count), nil
	default:
		return "", fmt.Errorf("unknown job: %s", jobName)
	}
}
