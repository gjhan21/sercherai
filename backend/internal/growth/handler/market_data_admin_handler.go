package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
)

func (h *AdminGrowthHandler) ListMarketDataQualityLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
	hours := 0
	if raw := strings.TrimSpace(c.Query("hours")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "hours must be a positive integer", Data: struct{}{}})
			return
		}
		hours = parsed
	}
	items, total, err := h.service.AdminListMarketDataQualityLogs(
		c.Query("asset_class"),
		c.Query("data_kind"),
		c.Query("severity"),
		c.Query("issue_code"),
		hours,
		page,
		pageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) GetMarketDataQualitySummary(c *gin.Context) {
	hours := 24
	if raw := strings.TrimSpace(c.Query("hours")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "hours must be a positive integer", Data: struct{}{}})
			return
		}
		hours = parsed
	}
	item, err := h.service.AdminGetMarketDataQualitySummary(c.Query("asset_class"), hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) GetMarketDerivedTruthSummary(c *gin.Context) {
	assetClass := strings.TrimSpace(c.Query("asset_class"))
	if assetClass == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "asset_class is required", Data: struct{}{}})
		return
	}
	item, err := h.service.AdminGetMarketDerivedTruthSummary(assetClass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) RebuildStockDerivedTruth(c *gin.Context) {
	h.rebuildMarketDerivedTruth(c, "STOCK", "STOCK", "REBUILD_DERIVED_TRUTH", "STOCK_QUOTES")
}

func (h *AdminGrowthHandler) RebuildFuturesDerivedTruth(c *gin.Context) {
	h.rebuildMarketDerivedTruth(c, "FUTURES", "FUTURES", "REBUILD_DERIVED_TRUTH", "FUTURES_QUOTES")
}

func (h *AdminGrowthHandler) rebuildMarketDerivedTruth(c *gin.Context, assetClass string, module string, action string, targetType string) {
	var req dto.MarketDerivedTruthRebuildRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	result, err := h.service.AdminRebuildMarketDerivedTruth(assetClass, req.TradeDate, req.Days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	summary := fmt.Sprintf("truth_bars=%d,status=%d,mappings=%d", result.TruthBarCount, result.StockStatusCount, result.FuturesMappingCount)
	reason := fmt.Sprintf("trade_date=%s,days=%d,warnings=%s", result.TradeDate, result.Days, strings.Join(result.Warnings, " | "))
	h.writeOperationLog(c, module, action, targetType, result.TradeDate, "", summary, reason)
	c.JSON(http.StatusOK, dto.OK(result))
}
