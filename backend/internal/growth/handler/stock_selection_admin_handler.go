package handler

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/growth/swarm"
	"sercherai/backend/internal/platform/config"
	"sercherai/backend/internal/platform/llm"
	"sercherai/backend/internal/platform/utils"
)

type AdminStockSelectionHandler struct {
	AdminBaseHandler
}

func NewAdminStockSelectionHandler(base *AdminBaseHandler) *AdminStockSelectionHandler {
	return &AdminStockSelectionHandler{AdminBaseHandler: *base}
}

type adminStockSelectionRunRequest struct {
	TradeDate                string `json:"trade_date"`
	ProfileID                string `json:"profile_id"`
	TemplateID               string `json:"template_id"`
	CompareWithLastPublished bool   `json:"compare_with_last_published"`
	DryRun                   bool   `json:"dry_run"`
}

type adminStockSelectionProfileRequest struct {
	Name                 string         `json:"name"`
	TemplateID           string         `json:"template_id"`
	Status               string         `json:"status"`
	IsDefault            bool           `json:"is_default"`
	SelectionModeDefault string         `json:"selection_mode_default"`
	UniverseScope        string         `json:"universe_scope"`
	UniverseConfig       map[string]any `json:"universe_config"`
	SeedMiningConfig     map[string]any `json:"seed_mining_config"`
	FactorConfig         map[string]any `json:"factor_config"`
	PortfolioConfig      map[string]any `json:"portfolio_config"`
	PublishConfig        map[string]any `json:"publish_config"`
	MarketAnalysisConfig map[string]any `json:"market_analysis_config"`
	CandidatePoolConfig  map[string]any `json:"candidate_pool_config"`
	ShortTermHeadConfig  map[string]any `json:"short_term_head_config"`
	SwingHeadConfig      map[string]any `json:"swing_head_config"`
	Description          string         `json:"description"`
	ChangeNote           string         `json:"change_note"`
}

type adminStockSelectionTemplateRequest struct {
	TemplateKey       string         `json:"template_key"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	MarketRegimeBias  string         `json:"market_regime_bias"`
	IsDefault         bool           `json:"is_default"`
	Status            string         `json:"status"`
	UniverseDefaults  map[string]any `json:"universe_defaults_json"`
	SeedDefaults      map[string]any `json:"seed_defaults_json"`
	FactorDefaults    map[string]any `json:"factor_defaults_json"`
	PortfolioDefaults map[string]any `json:"portfolio_defaults_json"`
	PublishDefaults   map[string]any `json:"publish_defaults_json"`
	MarketAnalysisDefaults map[string]any `json:"market_analysis_defaults_json"`
	CandidatePoolDefaults  map[string]any `json:"candidate_pool_defaults_json"`
	ShortTermHeadDefaults  map[string]any `json:"short_term_head_defaults_json"`
	SwingHeadDefaults      map[string]any `json:"swing_head_defaults_json"`
}

type adminStockSelectionRollbackRequest struct {
	VersionNo  int    `json:"version_no"`
	ChangeNote string `json:"change_note"`
}

type adminStockSelectionApproveRequest struct {
	ReviewNote     string `json:"review_note"`
	Force          bool   `json:"force"`
	OverrideReason string `json:"override_reason"`
}

type adminStockSelectionRejectRequest struct {
	ReviewNote string `json:"review_note"`
}

type adminStockEventReviewRequest struct {
	ReviewStatus   string         `json:"review_status"`
	ReviewNote     string         `json:"review_note"`
	Reviewer       string         `json:"reviewer"`
	ReviewMetadata map[string]any `json:"review_metadata"`
}

func (h *AdminStockSelectionHandler) GetStockSelectionOverview(c *gin.Context) {
	data, err := h.service.AdminGetStockSelectionOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

func (h *AdminStockSelectionHandler) ListStockEventClusters(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListStockEventClusters(model.StockEventQuery{
		ReviewStatus:   c.Query("review_status"),
		EventType:      c.Query("event_type"),
		ReviewPriority: c.Query("review_priority"),
		Symbol:         c.Query("symbol"),
		Sector:         c.Query("sector"),
		Topic:          c.Query("topic"),
		Page:           page,
		PageSize:       pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminStockSelectionHandler) GetStockEventCluster(c *gin.Context) {
	item, err := h.service.AdminGetStockEventCluster(c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "stock event cluster not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) ReviewStockEventCluster(c *gin.Context) {
	var req adminStockEventReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	reviewer := strings.TrimSpace(req.Reviewer)
	if reviewer == "" {
		reviewer = currentAdminOperator(c)
	}
	item, err := h.service.AdminReviewStockEventCluster(c.Param("id"), model.StockEventReview{
		ReviewStatus:   req.ReviewStatus,
		Reviewer:       reviewer,
		ReviewNote:     req.ReviewNote,
		ReviewMetadata: req.ReviewMetadata,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "stock event cluster not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) ListStockSelectionRuns(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListStockSelectionRuns(c.Query("status"), c.Query("review_status"), c.Query("profile_id"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminStockSelectionHandler) CreateStockSelectionRun(c *gin.Context) {
	var req adminStockSelectionRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminCreateStockSelectionRun(model.StockSelectionRunCreateRequest{
		TradeDate:                req.TradeDate,
		ProfileID:                req.ProfileID,
		TemplateID:               req.TemplateID,
		CompareWithLastPublished: req.CompareWithLastPublished,
		DryRun:                   req.DryRun,
	}, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) GetStockSelectionRun(c *gin.Context) {
	item, err := h.service.AdminGetStockSelectionRun(c.Param("run_id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "stock selection run not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) CompareStockSelectionRuns(c *gin.Context) {
	raw := strings.TrimSpace(c.Query("run_ids"))
	if raw == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "run_ids is required", Data: struct{}{}})
		return
	}
	runIDs := make([]string, 0)
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		runIDs = append(runIDs, item)
	}
	data, err := h.service.AdminCompareStockSelectionRuns(runIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

func (h *AdminStockSelectionHandler) ListStockSelectionProfiles(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListStockSelectionProfiles(c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminStockSelectionHandler) ListStockSelectionProfileVersions(c *gin.Context) {
	items, err := h.service.AdminListStockSelectionProfileVersions(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminStockSelectionHandler) CreateStockSelectionProfile(c *gin.Context) {
	var req adminStockSelectionProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminCreateStockSelectionProfile(stockSelectionProfileFromRequest(req, operator), req.ChangeNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) UpdateStockSelectionProfile(c *gin.Context) {
	var req adminStockSelectionProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminUpdateStockSelectionProfile(c.Param("id"), stockSelectionProfileFromRequest(req, operator), req.ChangeNote)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "stock selection profile not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) PublishStockSelectionProfile(c *gin.Context) {
	operator := currentAdminOperator(c)
	item, err := h.service.AdminPublishStockSelectionProfile(c.Param("id"), operator)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "stock selection profile not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) RollbackStockSelectionProfile(c *gin.Context) {
	var req adminStockSelectionRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminRollbackStockSelectionProfile(c.Param("id"), req.VersionNo, req.ChangeNote, operator)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "stock selection profile version not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) ListStockSelectionRunCandidates(c *gin.Context) {
	items, err := h.service.AdminListStockSelectionRunCandidates(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	stage := strings.ToUpper(strings.TrimSpace(c.Query("stage")))
	if stage == "" {
		c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
		return
	}
	filtered := make([]model.StockSelectionCandidateSnapshot, 0, len(items))
	for _, item := range items {
		if strings.EqualFold(item.Stage, stage) {
			filtered = append(filtered, item)
		}
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": filtered}))
}

func (h *AdminStockSelectionHandler) ListStockSelectionRunPortfolio(c *gin.Context) {
	items, err := h.service.AdminListStockSelectionRunPortfolio(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminStockSelectionHandler) ListStockSelectionRunEvidence(c *gin.Context) {
	items, err := h.service.AdminListStockSelectionRunEvidence(c.Param("run_id"), c.Query("symbol"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminStockSelectionHandler) ListStockSelectionRunEvaluations(c *gin.Context) {
	items, err := h.service.AdminListStockSelectionRunEvaluations(c.Param("run_id"), c.Query("symbol"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminStockSelectionHandler) ListStockSelectionProfileTemplates(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListStockSelectionProfileTemplates(c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminStockSelectionHandler) CreateStockSelectionProfileTemplate(c *gin.Context) {
	var req adminStockSelectionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	item, err := h.service.AdminCreateStockSelectionProfileTemplate(stockSelectionTemplateFromRequest(req, currentAdminOperator(c)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) UpdateStockSelectionProfileTemplate(c *gin.Context) {
	var req adminStockSelectionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	item, err := h.service.AdminUpdateStockSelectionProfileTemplate(c.Param("id"), stockSelectionTemplateFromRequest(req, currentAdminOperator(c)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "stock selection template not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) SetDefaultStockSelectionProfileTemplate(c *gin.Context) {
	item, err := h.service.AdminSetDefaultStockSelectionProfileTemplate(c.Param("id"), currentAdminOperator(c))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "stock selection template not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) ListStockSelectionEvaluationLeaderboard(c *gin.Context) {
	scope := strings.TrimSpace(c.Query("scope"))
	if scope == "" {
		scope = strings.TrimSpace(c.Query("evaluation_scope"))
	}
	items, err := h.service.AdminListStockSelectionEvaluationLeaderboard(c.Query("template_id"), c.Query("profile_id"), c.Query("market_regime"), scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminStockSelectionHandler) ListStockSelectionReviews(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListStockSelectionReviews(c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminStockSelectionHandler) ApproveStockSelectionReview(c *gin.Context) {
	var req adminStockSelectionApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminApproveStockSelectionReview(c.Param("run_id"), operator, req.ReviewNote, req.Force, req.OverrideReason)
	if err != nil {
		if detail, ok := extractStrategyPublishConflictDetail(err); ok {
			c.JSON(http.StatusConflict, dto.APIResponse{
				Code:    40901,
				Message: detail,
				Data: gin.H{
					"conflict_type": "PUBLISH_POLICY_BLOCKED",
					"detail":        detail,
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) RejectStockSelectionReview(c *gin.Context) {
	var req adminStockSelectionRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminRejectStockSelectionReview(c.Param("run_id"), operator, req.ReviewNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStockSelectionHandler) ListStockRecommendations(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListStockRecommendations(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminStockSelectionHandler) CreateStockRecommendation(c *gin.Context) {
	var req dto.StockRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	validFrom, err := normalizeAdminDateTime(req.ValidFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "valid_from " + err.Error(), Data: struct{}{}})
		return
	}
	validTo, err := normalizeAdminDateTime(req.ValidTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "valid_to " + err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	sourceType := strings.ToUpper(strings.TrimSpace(req.SourceType))
	if sourceType == "" {
		sourceType = "MANUAL"
	}
	performanceLabel := strings.ToUpper(strings.TrimSpace(req.PerformanceLabel))
	if performanceLabel == "" {
		performanceLabel = "PENDING"
	}
	strategyVersion := strings.TrimSpace(req.StrategyVersion)
	if strategyVersion == "" {
		strategyVersion = "manual-v1"
	}
	publisher := strings.TrimSpace(req.Publisher)
	if publisher == "" {
		publisher = operator
	}
	id, err := h.service.AdminCreateStockRecommendation(model.StockRecommendation{
		Symbol:           req.Symbol,
		Name:             req.Name,
		Score:            req.Score,
		RiskLevel:        req.RiskLevel,
		PositionRange:    req.PositionRange,
		ValidFrom:        validFrom,
		ValidTo:          validTo,
		Status:           req.Status,
		ReasonSummary:    req.ReasonSummary,
		SourceType:       sourceType,
		StrategyVersion:  strategyVersion,
		Reviewer:         strings.TrimSpace(req.Reviewer),
		Publisher:        publisher,
		ReviewNote:       strings.TrimSpace(req.ReviewNote),
		PerformanceLabel: performanceLabel,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "STOCK", "CREATE_RECOMMENDATION", "STOCK_RECOMMENDATION", id, "", req.Status, req.Symbol)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminStockSelectionHandler) UpdateStockRecommendationStatus(c *gin.Context) {
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

func (h *AdminStockSelectionHandler) GenerateStockRecommendationAIReview(c *gin.Context) {
	id := c.Param("id")

	// Phase 8-A: 集成 MiroFish 群体智能推演引擎
	cfg := config.Load()
	
	// 从管理端配置覆盖环境变量 (支持：llm.api_key, llm.base_url, llm.model_name)
	if llmConfigs, _, err := h.service.AdminListSystemConfigs("llm.", 1, 50); err == nil {
		for _, c := range llmConfigs {
			val := strings.TrimSpace(c.ConfigValue)
			if val == "" {
				continue
			}
			switch c.ConfigKey {
			case "llm.api_key":
				cfg.LLMAPIKey = val
			case "llm.base_url":
				cfg.LLMBaseURL = val
			case "llm.model_name":
				cfg.LLMModelName = val
			}
		}
	}
	
	llmClient := llm.NewClient(cfg)
	engine := swarm.NewSimulationEngine(llmClient)

	// 构造推演的事件背景（种子信息）
	// TODO: 后续深度融合时，应从 h.service.GetStockRecommendationDetail 获取真实的标的异动原因、新闻流和财务数据
	eventContext := fmt.Sprintf("标的 [ID: %s] 近期出现异常资金流入，且伴随行业利好政策传闻。由于部分指标已处于历史高位，市场产生较大分歧。请基于各自的立场给出观点。", id)

	simResult, err := engine.Run(eventContext)
	var generatedContent string
	if err != nil {
		generatedContent = fmt.Sprintf("【AI推演失败】原因：%v\n\n(请检查 LLM_API_KEY 配置及网络连接，当前为兜底Mock) 该标的在近期表现出强劲的趋势动量，主力资金持续净流入。短期内建议关注回撤风险，逢低可适当建仓。此复盘由 AI 自动生成，仅供参考。", err)
	} else {
		generatedContent = simResult.Summary
	}

	if err := h.service.AdminUpdateStockRecommendationAIReview(id, generatedContent); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "STOCK", "GENERATE_AI_REVIEW", "STOCK_RECOMMENDATION", id, "", "SUCCESS", "")
	c.JSON(http.StatusOK, dto.OK(gin.H{"ai_review_content": generatedContent}))
}

func (h *AdminStockSelectionHandler) SyncStockQuotes(c *gin.Context) {
	var req dto.StockQuoteSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultStockQuoteSourceKey()
	}
	if sourceKey == "" {
		sourceKey = "MOCK"
	}
	days := req.Days
	if days <= 0 {
		days = 120
	}
	if days > 365 {
		days = 365
	}
	symbols := normalizeStockSymbols(req.Symbols)
	syncMode := normalizeStockQuoteSyncMode(req.SyncMode, symbols)

	var (
		result model.MarketSyncResult
		err    error
	)
	if syncMode == "FULL_MARKET" {
		result, err = h.service.AdminSyncStockQuotesFromMaster(sourceKey, days)
	} else {
		result, err = h.service.AdminSyncStockQuotesDetailed(sourceKey, symbols, days)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	count := result.TruthCount
	if count <= 0 {
		count = result.BarCount
	}
	reason := fmt.Sprintf("days=%d,symbols=%d", days, len(symbols))
	h.writeOperationLog(c, "STOCK", "SYNC_QUOTES", "STOCK_QUOTES", sourceKey, requestedSourceKey, "count="+strconv.Itoa(count), reason)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":                count,
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"sync_mode":            syncMode,
		"symbols":              symbols,
		"result":               result,
	}))
}

func (h *AdminStockSelectionHandler) SyncStockInstrumentMaster(c *gin.Context) {
	var req dto.StockMasterSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultConfigValue("stock.master.default_source_key", "TUSHARE")
	}
	symbols := normalizeStockSymbols(req.Symbols)
	result, err := h.service.AdminSyncStockInstrumentMaster(sourceKey, symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	count := result.TruthCount
	reason := fmt.Sprintf("symbols=%d", len(symbols))
	h.writeOperationLog(c, "STOCK", "SYNC_INSTRUMENT_MASTER", "STOCK_MASTER", sourceKey, requestedSourceKey, "count="+strconv.Itoa(count), reason)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":                count,
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"symbols":              symbols,
		"result":               result,
	}))
}

func (h *AdminStockSelectionHandler) ListQuantTopStocks(c *gin.Context) {
	limit := 10
	if parsed, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("limit", "10"))); err == nil && parsed > 0 {
		limit = parsed
	}
	if limit > 50 {
		limit = 50
	}

	lookbackDays := 120
	if parsed, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("lookback_days", "120"))); err == nil && parsed > 0 {
		lookbackDays = parsed
	}
	if lookbackDays < 30 {
		lookbackDays = 30
	}
	if lookbackDays > 365 {
		lookbackDays = 365
	}

	items, err := h.service.AdminGetQuantTopStocks(limit, lookbackDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"items":         items,
		"limit":         limit,
		"lookback_days": lookbackDays,
		"total":         len(items),
	}))
}

func (h *AdminStockSelectionHandler) GenerateDailyStockRecommendations(c *gin.Context) {
	tradeDate := strings.TrimSpace(c.Query("trade_date"))
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	result, err := h.service.AdminGenerateDailyStockRecommendations(tradeDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "STOCK", "GENERATE_RECOMMENDATIONS", "DAILY_STOCK", tradeDate, "", "SUCCESS", strconv.Itoa(result.Count))
	c.JSON(http.StatusOK, dto.OK(result))
}

func (h *AdminStockSelectionHandler) GetStockSimulatedOverview(c *gin.Context) {
	data, err := h.service.AdminGetStockSimulatedOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

func (h *AdminStockSelectionHandler) ListStockSimulatedPositions(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	symbol := c.Query("symbol")
	items, total, err := h.service.AdminListStockSimulatedPositions(status, symbol, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminStockSelectionHandler) SettleSimulatedPositions(c *gin.Context) {
	tradeDate := strings.TrimSpace(c.Query("trade_date"))
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	if err := h.service.AdminSettlementSimulatedPositions(tradeDate); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "STOCK", "SETTLE_SIMULATED_POSITIONS", "STOCK_SIMULATED_POSITIONS", tradeDate, "", "SUCCESS", "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}


func stockSelectionProfileFromRequest(req adminStockSelectionProfileRequest, operator string) model.StockSelectionProfile {
	return model.StockSelectionProfile{
		Name:                 req.Name,
		TemplateID:           req.TemplateID,
		Status:               req.Status,
		IsDefault:            req.IsDefault,
		SelectionModeDefault: req.SelectionModeDefault,
		UniverseScope:        req.UniverseScope,
		UniverseConfig:       req.UniverseConfig,
		SeedMiningConfig:     req.SeedMiningConfig,
		FactorConfig:         req.FactorConfig,
		PortfolioConfig:      req.PortfolioConfig,
		PublishConfig:        req.PublishConfig,
		MarketAnalysisConfig: req.MarketAnalysisConfig,
		CandidatePoolConfig:  req.CandidatePoolConfig,
		ShortTermHeadConfig:  req.ShortTermHeadConfig,
		SwingHeadConfig:      req.SwingHeadConfig,
		Description:          req.Description,
		UpdatedBy:            operator,
	}
}

func stockSelectionTemplateFromRequest(req adminStockSelectionTemplateRequest, operator string) model.StockSelectionProfileTemplate {
	return model.StockSelectionProfileTemplate{
		TemplateKey:       req.TemplateKey,
		Name:              req.Name,
		Description:       req.Description,
		MarketRegimeBias:  req.MarketRegimeBias,
		IsDefault:         req.IsDefault,
		Status:            req.Status,
		UniverseDefaults:  req.UniverseDefaults,
		SeedDefaults:      req.SeedDefaults,
		FactorDefaults:    req.FactorDefaults,
		PortfolioDefaults: req.PortfolioDefaults,
		PublishDefaults:   req.PublishDefaults,
		MarketAnalysisDefaults: req.MarketAnalysisDefaults,
		CandidatePoolDefaults:  req.CandidatePoolDefaults,
		ShortTermHeadDefaults:  req.ShortTermHeadDefaults,
		SwingHeadDefaults:      req.SwingHeadDefaults,
		UpdatedBy:         operator,
	}
}

func (h *AdminStockSelectionHandler) resolveDefaultStockQuoteSourceKey() string {
	items, _, err := h.service.AdminListSystemConfigs("stock.quotes.default_source_key", 1, 10)
	if err != nil || len(items) == 0 {
		return "TUSHARE"
	}
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.ConfigKey), "stock.quotes.default_source_key") {
			return strings.TrimSpace(item.ConfigValue)
		}
	}
	return "TUSHARE"
}

func (h *AdminStockSelectionHandler) resolveDefaultConfigValue(configKey string, fallback string) string {
	items, _, err := h.service.AdminListSystemConfigs(configKey, 1, 10)
	if err != nil || len(items) == 0 {
		return fallback
	}
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.ConfigKey), configKey) {
			return strings.TrimSpace(item.ConfigValue)
		}
	}
	return fallback
}



func (h *AdminStockSelectionHandler) ListQuantEvaluation(c *gin.Context) {
	windowDays, topN := h.parseQuantEvaluationQuery(c)
	summary, points, riskItems, rotationItems, err := h.service.AdminGetQuantEvaluation(windowDays, topN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"summary":        summary,
		"items":          points,
		"risk_items":     riskItems,
		"rotation_items": rotationItems,
		"total":          len(points),
	}))
}

func (h *AdminStockSelectionHandler) ExportQuantEvaluationCSV(c *gin.Context) {
	windowDays, topN := h.parseQuantEvaluationQuery(c)
	summary, points, riskItems, rotationItems, err := h.service.AdminGetQuantEvaluation(windowDays, topN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"section", "field", "value"})
	_ = writer.Write([]string{"summary", "window_days", strconv.Itoa(summary.WindowDays)})
	_ = writer.Write([]string{"summary", "top_n", strconv.Itoa(summary.TopN)})
	_ = writer.Write([]string{"summary", "sample_days", strconv.Itoa(summary.SampleDays)})
	_ = writer.Write([]string{"summary", "sample_count", strconv.Itoa(summary.SampleCount)})
	_ = writer.Write([]string{"summary", "avg_return_5", fmt.Sprintf("%.6f", summary.AvgReturn5)})
	_ = writer.Write([]string{"summary", "hit_rate_5", fmt.Sprintf("%.6f", summary.HitRate5)})
	_ = writer.Write([]string{"summary", "max_drawdown_5", fmt.Sprintf("%.6f", summary.MaxDrawdown5)})
	_ = writer.Write([]string{"summary", "avg_return_10", fmt.Sprintf("%.6f", summary.AvgReturn10)})
	_ = writer.Write([]string{"summary", "hit_rate_10", fmt.Sprintf("%.6f", summary.HitRate10)})
	_ = writer.Write([]string{"summary", "max_drawdown_10", fmt.Sprintf("%.6f", summary.MaxDrawdown10)})
	_ = writer.Write([]string{"summary", "benchmark_avg_return_5", fmt.Sprintf("%.6f", summary.BenchmarkAvgReturn5)})
	_ = writer.Write([]string{"summary", "benchmark_avg_return_10", fmt.Sprintf("%.6f", summary.BenchmarkAvgReturn10)})
	_ = writer.Write([]string{"summary", "generated_at", summary.GeneratedAt})
	_ = writer.Write([]string{})

	_ = writer.Write([]string{
		"points_trade_date",
		"sample_count",
		"avg_return_5",
		"hit_rate_5",
		"benchmark_return_5",
		"avg_return_10",
		"hit_rate_10",
		"benchmark_return_10",
		"cumulative_return_5",
		"cumulative_benchmark_5",
		"cumulative_excess_5",
		"cumulative_return_10",
		"cumulative_benchmark_10",
		"cumulative_excess_10",
	})
	for _, item := range points {
		_ = writer.Write([]string{
			item.TradeDate,
			strconv.Itoa(item.SampleCount),
			fmt.Sprintf("%.6f", item.AvgReturn5),
			fmt.Sprintf("%.6f", item.HitRate5),
			fmt.Sprintf("%.6f", item.BenchmarkReturn),
			fmt.Sprintf("%.6f", item.AvgReturn10),
			fmt.Sprintf("%.6f", item.HitRate10),
			fmt.Sprintf("%.6f", item.BenchmarkReturn10),
			fmt.Sprintf("%.6f", item.CumulativeReturn5),
			fmt.Sprintf("%.6f", item.CumulativeBenchmark5),
			fmt.Sprintf("%.6f", item.CumulativeExcess5),
			fmt.Sprintf("%.6f", item.CumulativeReturn10),
			fmt.Sprintf("%.6f", item.CumulativeBenchmark10),
			fmt.Sprintf("%.6f", item.CumulativeExcess10),
		})
	}
	_ = writer.Write([]string{})

	_ = writer.Write([]string{"risk_level", "sample_count", "avg_return_5", "hit_rate_5", "avg_return_10", "hit_rate_10"})
	for _, item := range riskItems {
		_ = writer.Write([]string{
			item.RiskLevel,
			strconv.Itoa(item.SampleCount),
			fmt.Sprintf("%.6f", item.AvgReturn5),
			fmt.Sprintf("%.6f", item.HitRate5),
			fmt.Sprintf("%.6f", item.AvgReturn10),
			fmt.Sprintf("%.6f", item.HitRate10),
		})
	}
	_ = writer.Write([]string{})

	_ = writer.Write([]string{"rotation_trade_date", "top_symbols", "entered", "exited", "stayed_count", "changed_count"})
	for _, item := range rotationItems {
		_ = writer.Write([]string{
			item.TradeDate,
			strings.Join(item.TopSymbols, "|"),
			strings.Join(item.Entered, "|"),
			strings.Join(item.Exited, "|"),
			strconv.Itoa(item.StayedCount),
			strconv.Itoa(item.ChangedCount),
		})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	fileName := fmt.Sprintf("stock_quant_evaluation_%dd_top%d.csv", windowDays, topN)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.String(http.StatusOK, buf.String())
}

func (h *AdminStockSelectionHandler) parseQuantEvaluationQuery(c *gin.Context) (int, int) {
	windowDays := 60
	if parsed, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("days", "60"))); err == nil && parsed > 0 {
		windowDays = parsed
	}
	if windowDays < 20 {
		windowDays = 20
	}
	if windowDays > 365 {
		windowDays = 365
	}
	topN := 10
	if parsed, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("top_n", "10"))); err == nil && parsed > 0 {
		topN = parsed
	}
	if topN < 1 {
		topN = 1
	}
	if topN > 30 {
		topN = 30
	}
	return windowDays, topN
}

func normalizeStockQuoteSyncMode(raw string, symbols []string) string {
	mode := strings.ToUpper(strings.TrimSpace(raw))
	if mode == "FULL_MARKET" {
		return "FULL_MARKET"
	}
	if len(symbols) > 0 {
		return "DETAILED"
	}
	return "FULL_MARKET"
}
