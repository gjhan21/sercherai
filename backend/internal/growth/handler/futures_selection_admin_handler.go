package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/platform/utils"
)

type AdminFuturesSelectionHandler struct {
	AdminBaseHandler
}

func NewAdminFuturesSelectionHandler(base *AdminBaseHandler) *AdminFuturesSelectionHandler {
	return &AdminFuturesSelectionHandler{AdminBaseHandler: *base}
}

type adminFuturesSelectionRunRequest struct {
	TradeDate                string `json:"trade_date"`
	ProfileID                string `json:"profile_id"`
	TemplateID               string `json:"template_id"`
	CompareWithLastPublished bool   `json:"compare_with_last_published"`
	DryRun                   bool   `json:"dry_run"`
}

type adminFuturesSelectionProfileRequest struct {
	Name            string         `json:"name"`
	TemplateID      string         `json:"template_id"`
	Status          string         `json:"status"`
	IsDefault       bool           `json:"is_default"`
	StyleDefault    string         `json:"style_default"`
	ContractScope   string         `json:"contract_scope"`
	UniverseConfig  map[string]any `json:"universe_config"`
	FactorConfig    map[string]any `json:"factor_config"`
	PortfolioConfig map[string]any `json:"portfolio_config"`
	PublishConfig   map[string]any `json:"publish_config"`
	Description     string         `json:"description"`
	ChangeNote      string         `json:"change_note"`
}

type adminFuturesSelectionRollbackRequest struct {
	VersionNo  int    `json:"version_no"`
	ChangeNote string `json:"change_note"`
}

type adminFuturesSelectionTemplateRequest struct {
	TemplateKey       string         `json:"template_key"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	MarketRegimeBias  string         `json:"market_regime_bias"`
	IsDefault         bool           `json:"is_default"`
	Status            string         `json:"status"`
	UniverseDefaults  map[string]any `json:"universe_defaults_json"`
	FactorDefaults    map[string]any `json:"factor_defaults_json"`
	PortfolioDefaults map[string]any `json:"portfolio_defaults_json"`
	PublishDefaults   map[string]any `json:"publish_defaults_json"`
}

type adminFuturesSelectionApproveRequest struct {
	ReviewNote     string `json:"review_note"`
	Force          bool   `json:"force"`
	OverrideReason string `json:"override_reason"`
}

type adminFuturesSelectionRejectRequest struct {
	ReviewNote string `json:"review_note"`
}

func (h *AdminFuturesSelectionHandler) GetFuturesSelectionOverview(c *gin.Context) {
	data, err := h.service.AdminGetFuturesSelectionOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

func (h *AdminFuturesSelectionHandler) ListFuturesSelectionRuns(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListFuturesSelectionRuns(c.Query("status"), c.Query("review_status"), c.Query("profile_id"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminFuturesSelectionHandler) CreateFuturesSelectionRun(c *gin.Context) {
	var req adminFuturesSelectionRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminCreateFuturesSelectionRun(model.FuturesSelectionRunCreateRequest{
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

func (h *AdminFuturesSelectionHandler) GetFuturesSelectionRun(c *gin.Context) {
	item, err := h.service.AdminGetFuturesSelectionRun(c.Param("run_id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "futures selection run not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminFuturesSelectionHandler) CompareFuturesSelectionRuns(c *gin.Context) {
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
	data, err := h.service.AdminCompareFuturesSelectionRuns(runIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

func (h *AdminFuturesSelectionHandler) ListFuturesSelectionProfiles(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListFuturesSelectionProfiles(c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminFuturesSelectionHandler) ListFuturesSelectionProfileVersions(c *gin.Context) {
	items, err := h.service.AdminListFuturesSelectionProfileVersions(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminFuturesSelectionHandler) CreateFuturesSelectionProfile(c *gin.Context) {
	var req adminFuturesSelectionProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminCreateFuturesSelectionProfile(futuresSelectionProfileFromRequest(req, operator), req.ChangeNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminFuturesSelectionHandler) UpdateFuturesSelectionProfile(c *gin.Context) {
	var req adminFuturesSelectionProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminUpdateFuturesSelectionProfile(c.Param("id"), futuresSelectionProfileFromRequest(req, operator), req.ChangeNote)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "futures selection profile not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminFuturesSelectionHandler) PublishFuturesSelectionProfile(c *gin.Context) {
	operator := currentAdminOperator(c)
	item, err := h.service.AdminPublishFuturesSelectionProfile(c.Param("id"), operator)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "futures selection profile not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminFuturesSelectionHandler) RollbackFuturesSelectionProfile(c *gin.Context) {
	var req adminFuturesSelectionRollbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminRollbackFuturesSelectionProfile(c.Param("id"), req.VersionNo, req.ChangeNote, operator)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "futures selection profile version not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminFuturesSelectionHandler) ListFuturesSelectionRunCandidates(c *gin.Context) {
	items, err := h.service.AdminListFuturesSelectionRunCandidates(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	stage := strings.ToUpper(strings.TrimSpace(c.Query("stage")))
	if stage == "" {
		c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
		return
	}
	filtered := make([]model.FuturesSelectionCandidateSnapshot, 0, len(items))
	for _, item := range items {
		if strings.EqualFold(item.Stage, stage) {
			filtered = append(filtered, item)
		}
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": filtered}))
}

func (h *AdminFuturesSelectionHandler) ListFuturesSelectionRunPortfolio(c *gin.Context) {
	items, err := h.service.AdminListFuturesSelectionRunPortfolio(c.Param("run_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminFuturesSelectionHandler) ListFuturesSelectionRunEvidence(c *gin.Context) {
	items, err := h.service.AdminListFuturesSelectionRunEvidence(c.Param("run_id"), c.Query("contract"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminFuturesSelectionHandler) ListFuturesSelectionRunEvaluations(c *gin.Context) {
	items, err := h.service.AdminListFuturesSelectionRunEvaluations(c.Param("run_id"), c.Query("contract"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminFuturesSelectionHandler) ListFuturesSelectionProfileTemplates(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListFuturesSelectionProfileTemplates(c.Query("status"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminFuturesSelectionHandler) CreateFuturesSelectionProfileTemplate(c *gin.Context) {
	var req adminFuturesSelectionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	item, err := h.service.AdminCreateFuturesSelectionProfileTemplate(futuresSelectionTemplateFromRequest(req, currentAdminOperator(c)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminFuturesSelectionHandler) UpdateFuturesSelectionProfileTemplate(c *gin.Context) {
	var req adminFuturesSelectionTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	item, err := h.service.AdminUpdateFuturesSelectionProfileTemplate(c.Param("id"), futuresSelectionTemplateFromRequest(req, currentAdminOperator(c)))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "futures selection template not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminFuturesSelectionHandler) SetDefaultFuturesSelectionProfileTemplate(c *gin.Context) {
	item, err := h.service.AdminSetDefaultFuturesSelectionProfileTemplate(c.Param("id"), currentAdminOperator(c))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "futures selection template not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminFuturesSelectionHandler) ListFuturesSelectionEvaluationLeaderboard(c *gin.Context) {
	items, err := h.service.AdminListFuturesSelectionEvaluationLeaderboard(c.Query("template_id"), c.Query("profile_id"), c.Query("market_regime"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminFuturesSelectionHandler) ApproveFuturesSelectionReview(c *gin.Context) {
	var req adminFuturesSelectionApproveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminApproveFuturesSelectionReview(c.Param("run_id"), operator, req.ReviewNote, req.Force, req.OverrideReason)
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

func (h *AdminFuturesSelectionHandler) RejectFuturesSelectionReview(c *gin.Context) {
	var req adminFuturesSelectionRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator := currentAdminOperator(c)
	item, err := h.service.AdminRejectFuturesSelectionReview(c.Param("run_id"), operator, req.ReviewNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminFuturesSelectionHandler) GenerateDailyFuturesStrategies(c *gin.Context) {
	tradeDate := c.Query("trade_date")
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	result, err := h.service.AdminGenerateDailyFuturesStrategies(tradeDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	afterValue := "count=" + strconv.Itoa(result.Count)
	if strings.TrimSpace(result.PublishID) != "" {
		afterValue += " publish_id=" + result.PublishID
	}
	if strings.TrimSpace(result.GenerationMode) != "" {
		afterValue += " generation_mode=" + result.GenerationMode
	}
	h.writeOperationLog(c, "FUTURES", "GENERATE_DAILY_STRATEGIES", "BATCH", tradeDate, "", afterValue, "")
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":           result.Count,
		"publish_id":      result.PublishID,
		"publish_version": result.PublishVersion,
		"report_summary":  result.ReportSummary,
		"generation_mode": result.GenerationMode,
		"archive_enabled": result.ArchiveEnabled,
	}))
}

func (h *AdminFuturesSelectionHandler) ListFuturesStrategies(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	contract := c.Query("contract")
	items, total, err := h.service.AdminListFuturesStrategies(status, contract, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminFuturesSelectionHandler) CreateFuturesStrategy(c *gin.Context) {
	var req dto.FuturesStrategyRequest
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
	id, err := h.service.AdminCreateFuturesStrategy(model.FuturesStrategy{
		Contract:      req.Contract,
		Name:          req.Name,
		Direction:     req.Direction,
		RiskLevel:     req.RiskLevel,
		PositionRange: req.PositionRange,
		ValidFrom:     validFrom,
		ValidTo:       validTo,
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

func (h *AdminFuturesSelectionHandler) UpdateFuturesStrategyStatus(c *gin.Context) {
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

func futuresSelectionProfileFromRequest(req adminFuturesSelectionProfileRequest, operator string) model.FuturesSelectionProfile {
	return model.FuturesSelectionProfile{
		Name:            req.Name,
		TemplateID:      req.TemplateID,
		Status:          req.Status,
		IsDefault:       req.IsDefault,
		StyleDefault:    req.StyleDefault,
		ContractScope:   req.ContractScope,
		UniverseConfig:  req.UniverseConfig,
		FactorConfig:    req.FactorConfig,
		PortfolioConfig: req.PortfolioConfig,
		PublishConfig:   req.PublishConfig,
		Description:     req.Description,
		UpdatedBy:       operator,
	}
}

func futuresSelectionTemplateFromRequest(req adminFuturesSelectionTemplateRequest, operator string) model.FuturesSelectionProfileTemplate {
	return model.FuturesSelectionProfileTemplate{
		TemplateKey:       req.TemplateKey,
		Name:              req.Name,
		Description:       req.Description,
		MarketRegimeBias:  req.MarketRegimeBias,
		IsDefault:         req.IsDefault,
		Status:            req.Status,
		UniverseDefaults:  req.UniverseDefaults,
		FactorDefaults:    req.FactorDefaults,
		PortfolioDefaults: req.PortfolioDefaults,
		PublishDefaults:   req.PublishDefaults,
		UpdatedBy:         operator,
	}
}

func (h *AdminFuturesSelectionHandler) GetFuturesSimulatedOverview(c *gin.Context) {
	data, err := h.service.AdminGetFuturesSimulatedOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

func (h *AdminFuturesSelectionHandler) ListFuturesSimulatedPositions(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	contract := c.Query("contract")
	items, total, err := h.service.AdminListFuturesSimulatedPositions(status, contract, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminFuturesSelectionHandler) SettleSimulatedPositions(c *gin.Context) {
	tradeDate := c.Query("trade_date")
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	if err := h.service.AdminSettlementFuturesSimulatedPositions(tradeDate); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

