package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
)

func (h *AdminGrowthHandler) CreateForecastL3Run(c *gin.Context) {
	var req dto.StrategyForecastL3CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorUserID := currentForecastL3Operator(c)
	run, err := h.service.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		TargetKey:      req.TargetKey,
		TargetLabel:    req.TargetLabel,
		TriggerType:    firstNonEmpty(strings.TrimSpace(req.TriggerType), model.StrategyForecastL3TriggerTypeAdminManual),
		RequestUserID:  operatorUserID,
		OperatorUserID: operatorUserID,
		PriorityScore:  req.PriorityScore,
		Reason:         req.Reason,
	})
	if err != nil {
		if isForecastL3BadRequest(err) {
			c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: err.Error(), Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "FORECAST", "CREATE_L3_RUN", "FORECAST_L3_RUN", run.ID, "", run.Status, run.TargetKey)
	c.JSON(http.StatusOK, dto.OK(run))
}

func (h *AdminGrowthHandler) ListForecastL3Runs(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListStrategyForecastL3Runs(
		strings.TrimSpace(c.Query("user_id")),
		c.Query("status"),
		c.Query("target_type"),
		c.Query("trigger_type"),
		page,
		pageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) GetForecastL3RunDetail(c *gin.Context) {
	detail, err := h.service.GetStrategyForecastL3RunDetail(c.Param("id"))
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "forecast run not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(detail))
}

func (h *AdminGrowthHandler) RetryForecastL3Run(c *gin.Context) {
	var req dto.StrategyForecastL3RetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	run, err := h.service.RetryStrategyForecastL3Run(c.Param("id"), currentForecastL3Operator(c), req.Reason)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "forecast run not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "FORECAST", "RETRY_L3_RUN", "FORECAST_L3_RUN", run.ID, "", run.Status, req.Reason)
	c.JSON(http.StatusOK, dto.OK(run))
}

func (h *AdminGrowthHandler) CancelForecastL3Run(c *gin.Context) {
	var req dto.StrategyForecastL3CancelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	run, err := h.service.CancelStrategyForecastL3Run(c.Param("id"), currentForecastL3Operator(c), req.Reason)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "forecast run not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "FORECAST", "CANCEL_L3_RUN", "FORECAST_L3_RUN", run.ID, "", run.Status, req.Reason)
	c.JSON(http.StatusOK, dto.OK(run))
}

func (h *AdminGrowthHandler) ListForecastL3Quality(c *gin.Context) {
	days := 0
	if raw := strings.TrimSpace(c.Query("days")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			days = parsed
		}
	}
	items, err := h.service.ListStrategyForecastL3QualitySummaries(c.Query("target_type"), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func currentForecastL3Operator(c *gin.Context) string {
	if value, ok := c.Get("user_id"); ok {
		if userID, castOK := value.(string); castOK && strings.TrimSpace(userID) != "" {
			return strings.TrimSpace(userID)
		}
	}
	if value := strings.TrimSpace(c.Query("operator_user_id")); value != "" {
		return value
	}
	return "system"
}
