package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
)

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

func (h *AdminGrowthHandler) GetStockSelectionOverview(c *gin.Context) {
	data, err := h.service.AdminGetStockSelectionOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

func (h *AdminGrowthHandler) ListStockSelectionRuns(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListStockSelectionRuns(c.Query("status"), c.Query("review_status"), c.Query("profile_id"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateStockSelectionRun(c *gin.Context) {
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

func (h *AdminGrowthHandler) GetStockSelectionRun(c *gin.Context) {
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

func (h *AdminGrowthHandler) CompareStockSelectionRuns(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListStockSelectionProfiles(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListStockSelectionProfiles(c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ListStockSelectionProfileVersions(c *gin.Context) {
	items, err := h.service.AdminListStockSelectionProfileVersions(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) CreateStockSelectionProfile(c *gin.Context) {
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

func (h *AdminGrowthHandler) UpdateStockSelectionProfile(c *gin.Context) {
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

func (h *AdminGrowthHandler) PublishStockSelectionProfile(c *gin.Context) {
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

func (h *AdminGrowthHandler) RollbackStockSelectionProfile(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListStockSelectionRunCandidates(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListStockSelectionRunPortfolio(c *gin.Context) {
	items, err := h.service.AdminListStockSelectionRunPortfolio(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) ListStockSelectionRunEvidence(c *gin.Context) {
	items, err := h.service.AdminListStockSelectionRunEvidence(c.Param("run_id"), c.Query("symbol"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) ListStockSelectionRunEvaluations(c *gin.Context) {
	items, err := h.service.AdminListStockSelectionRunEvaluations(c.Param("run_id"), c.Query("symbol"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) ListStockSelectionProfileTemplates(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListStockSelectionProfileTemplates(c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateStockSelectionProfileTemplate(c *gin.Context) {
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

func (h *AdminGrowthHandler) UpdateStockSelectionProfileTemplate(c *gin.Context) {
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

func (h *AdminGrowthHandler) SetDefaultStockSelectionProfileTemplate(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListStockSelectionEvaluationLeaderboard(c *gin.Context) {
	items, err := h.service.AdminListStockSelectionEvaluationLeaderboard(c.Query("template_id"), c.Query("profile_id"), c.Query("market_regime"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) ListStockSelectionReviews(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListStockSelectionReviews(c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ApproveStockSelectionReview(c *gin.Context) {
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

func (h *AdminGrowthHandler) RejectStockSelectionReview(c *gin.Context) {
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

func currentAdminOperator(c *gin.Context) string {
	if value, ok := c.Get("user_id"); ok {
		if operator, ok := value.(string); ok && strings.TrimSpace(operator) != "" {
			return strings.TrimSpace(operator)
		}
	}
	return "admin"
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
		UpdatedBy:         operator,
	}
}
