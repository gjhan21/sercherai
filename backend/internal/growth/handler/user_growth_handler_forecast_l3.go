package handler

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
)

func (h *UserGrowthHandler) CreateForecastL3Run(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req dto.StrategyForecastL3CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	run, err := h.service.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:    req.TargetType,
		TargetID:      req.TargetID,
		TargetKey:     req.TargetKey,
		TargetLabel:   req.TargetLabel,
		TriggerType:   firstNonEmpty(strings.TrimSpace(req.TriggerType), model.StrategyForecastL3TriggerTypeUserRequest),
		RequestUserID: userID,
		PriorityScore: req.PriorityScore,
		Reason:        req.Reason,
	})
	if err != nil {
		if isForecastL3BadRequest(err) {
			c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: err.Error(), Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(run))
}

func (h *UserGrowthHandler) ListForecastL3Runs(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListStrategyForecastL3Runs(
		userID,
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

func (h *UserGrowthHandler) GetForecastL3RunDetail(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	detail, err := h.service.GetStrategyForecastL3RunDetailForUser(c.Param("id"), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "forecast run not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if strings.TrimSpace(detail.Run.RequestUserID) != "" && detail.Run.RequestUserID != userID {
		c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "forecast run not found", Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(detail))
}

func isForecastL3BadRequest(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "invalid") || 
		strings.Contains(msg, "required") || 
		strings.Contains(msg, "limit reached") ||
		strings.Contains(msg, "disabled")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
