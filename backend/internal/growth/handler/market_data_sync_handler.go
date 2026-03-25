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
	"sercherai/backend/internal/growth/model"
)

func normalizeFuturesContracts(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	items := make([]string, 0, len(raw))
	for _, value := range raw {
		normalized := strings.ToUpper(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	return items
}

func (h *AdminGrowthHandler) resolveDefaultConfigValue(configKey string, fallback string) string {
	items, _, err := h.service.AdminListSystemConfigs(configKey, 1, 20)
	if err != nil {
		return fallback
	}
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.ConfigKey), strings.TrimSpace(configKey)) {
			value := strings.ToUpper(strings.TrimSpace(item.ConfigValue))
			if value != "" {
				return value
			}
		}
	}
	return fallback
}

func (h *AdminGrowthHandler) SyncFuturesQuotes(c *gin.Context) {
	var req dto.FuturesQuoteSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultConfigValue("futures.quotes.default_source_key", "TUSHARE")
	}
	days := req.Days
	if days <= 0 {
		days = 120
	}
	if days > 365 {
		days = 365
	}
	contracts := normalizeFuturesContracts(req.Contracts)
	result, err := h.service.AdminSyncFuturesQuotes(sourceKey, contracts, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	count := result.TruthCount
	if count <= 0 {
		count = result.BarCount
	}
	beforeValue := requestedSourceKey
	if beforeValue == "" {
		beforeValue = "DEFAULT"
	}
	reason := fmt.Sprintf("days=%d,contracts=%d", days, len(contracts))
	h.writeOperationLog(c, "FUTURES", "SYNC_QUOTES", "FUTURES_QUOTES", sourceKey, beforeValue, "count="+strconv.Itoa(count), reason)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":                count,
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"contracts":            contracts,
		"result":               result,
	}))
}

func normalizeFuturesInventorySymbols(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	items := make([]string, 0, len(raw))
	for _, value := range raw {
		normalized := strings.ToUpper(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		letters := make([]rune, 0, len(normalized))
		for _, ch := range normalized {
			if ch >= 'A' && ch <= 'Z' {
				letters = append(letters, ch)
			} else {
				break
			}
		}
		if len(letters) > 0 {
			normalized = string(letters)
		}
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	return items
}

func (h *AdminGrowthHandler) SyncFuturesInventory(c *gin.Context) {
	var req dto.FuturesInventorySyncRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultConfigValue("futures.inventory.default_source_key", "TUSHARE")
	}
	days := req.Days
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	symbols := normalizeFuturesInventorySymbols(req.Symbols)
	result, err := h.service.AdminSyncFuturesInventory(sourceKey, symbols, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	count := result.InventoryCount
	beforeValue := requestedSourceKey
	if beforeValue == "" {
		beforeValue = "DEFAULT"
	}
	reason := fmt.Sprintf("days=%d,symbols=%d", days, len(symbols))
	h.writeOperationLog(c, "FUTURES", "SYNC_INVENTORY", "FUTURES_INVENTORY", sourceKey, beforeValue, "count="+strconv.Itoa(count), reason)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":                count,
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"symbols":              symbols,
		"result":               result,
	}))
}

func (h *AdminGrowthHandler) SyncMarketNewsSource(c *gin.Context) {
	var req dto.MarketNewsSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultConfigValue("market.news.default_source_key", "AKSHARE")
	}
	days := req.Days
	if days <= 0 {
		days = 7
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	symbols := normalizeStockSymbols(req.Symbols)
	result, err := h.service.AdminSyncMarketNews(sourceKey, symbols, days, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(
		c,
		"NEWS",
		"SYNC_MARKET_NEWS",
		"MARKET_NEWS",
		sourceKey,
		requestedSourceKey,
		fmt.Sprintf("count=%d", result.NewsCount),
		fmt.Sprintf("days=%d,symbols=%d,limit=%d", days, len(symbols), limit),
	)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":                result.NewsCount,
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"limit":                limit,
		"symbols":              symbols,
		"result":               result,
	}))
}

func (h *AdminGrowthHandler) SyncStockDailyBasics(c *gin.Context) {
	var req dto.StockMarketDataSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultConfigValue("stock.daily_basic.default_source_key", "TUSHARE")
	}
	days := req.Days
	if days <= 0 {
		days = 120
	}
	if days > 365 {
		days = 365
	}
	symbols := normalizeStockSymbols(req.Symbols)
	result, err := h.service.AdminSyncStockDailyBasics(sourceKey, symbols, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "STOCK", "SYNC_DAILY_BASIC", "STOCK_DAILY_BASIC", sourceKey, requestedSourceKey, fmt.Sprintf("count=%d", result.TruthCount), fmt.Sprintf("days=%d,symbols=%d", days, len(symbols)))
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":                result.TruthCount,
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"symbols":              symbols,
		"result":               result,
	}))
}

func (h *AdminGrowthHandler) SyncStockMoneyflows(c *gin.Context) {
	var req dto.StockMarketDataSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultConfigValue("stock.moneyflow.default_source_key", "TUSHARE")
	}
	days := req.Days
	if days <= 0 {
		days = 120
	}
	if days > 365 {
		days = 365
	}
	symbols := normalizeStockSymbols(req.Symbols)
	result, err := h.service.AdminSyncStockMoneyflows(sourceKey, symbols, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "STOCK", "SYNC_MONEYFLOW", "STOCK_MONEYFLOW", sourceKey, requestedSourceKey, fmt.Sprintf("count=%d", result.TruthCount), fmt.Sprintf("days=%d,symbols=%d", days, len(symbols)))
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":                result.TruthCount,
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"symbols":              symbols,
		"result":               result,
	}))
}

func (h *AdminGrowthHandler) SyncStockNewsSource(c *gin.Context) {
	var req dto.StockMarketDataSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultConfigValue("stock.news.default_source_key", "TUSHARE")
	}
	days := req.Days
	if days <= 0 {
		days = 7
	}
	if days > 30 {
		days = 30
	}
	symbols := normalizeStockSymbols(req.Symbols)
	result, err := h.service.AdminSyncStockNewsRaw(sourceKey, symbols, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "STOCK", "SYNC_STOCK_NEWS", "STOCK_NEWS_RAW", sourceKey, requestedSourceKey, fmt.Sprintf("count=%d", result.NewsCount), fmt.Sprintf("days=%d,symbols=%d", days, len(symbols)))
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":                result.NewsCount,
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"symbols":              symbols,
		"result":               result,
	}))
}

func (h *AdminGrowthHandler) BackfillStockMarketData(c *gin.Context) {
	var req dto.StockMarketDataBackfillRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultConfigValue("stock.master.default_source_key", "TUSHARE")
	}
	days := req.Days
	if days <= 0 {
		days = 120
	}
	if days > 365 {
		days = 365
	}
	symbols := normalizeStockSymbols(req.Symbols)

	masterResult, err := h.service.AdminSyncStockInstrumentMaster(sourceKey, symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	var quotesSyncResult model.MarketSyncResult
	if len(symbols) == 0 {
		quotesSyncResult, err = h.service.AdminSyncStockQuotesFromMaster(sourceKey, days)
	} else {
		quotesSyncResult, err = h.service.AdminSyncStockQuotesDetailed(sourceKey, symbols, days)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	dailyBasicResult, err := h.service.AdminSyncStockDailyBasics(sourceKey, symbols, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	moneyflowResult, err := h.service.AdminSyncStockMoneyflows(sourceKey, symbols, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	newsDays := days
	if newsDays > 30 {
		newsDays = 30
	}
	newsResult, err := h.service.AdminSyncStockNewsRaw(sourceKey, symbols, newsDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	h.writeOperationLog(c, "STOCK", "BACKFILL_MARKET_DATA", "STOCK_BACKFILL", sourceKey, requestedSourceKey, fmt.Sprintf("master=%d", masterResult.TruthCount), fmt.Sprintf("days=%d,symbols=%d", days, len(symbols)))
	marketResult := gin.H{
		"master_result":      masterResult,
		"quotes_result":      quotesSyncResult,
		"daily_basic_result": dailyBasicResult,
		"moneyflow_result":   moneyflowResult,
		"news_result":        newsResult,
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"symbols":              symbols,
		"result":               marketResult,
		"master_result":        masterResult,
		"quotes_result":        quotesSyncResult,
		"daily_basic_result":   dailyBasicResult,
		"moneyflow_result":     moneyflowResult,
		"news_result":          newsResult,
	}))
}
