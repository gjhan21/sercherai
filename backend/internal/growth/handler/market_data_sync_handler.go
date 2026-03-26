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

func isTushareNewsPermissionDeniedError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(strings.TrimSpace(err.Error()))
	if text == "" {
		return false
	}
	return strings.Contains(text, "tushare error(anns_d)") &&
		(strings.Contains(text, "没有接口访问权限") || strings.Contains(text, "no permission"))
}

func preferAkshareWhenDefaultRequested(requestedSourceKey string, resolvedSourceKey string) string {
	if strings.TrimSpace(requestedSourceKey) != "" {
		return resolvedSourceKey
	}
	if strings.EqualFold(strings.TrimSpace(resolvedSourceKey), "TUSHARE") {
		return "AKSHARE"
	}
	return resolvedSourceKey
}

func marketSyncResultCount(result model.MarketSyncResult) int {
	if result.TruthCount > 0 {
		return result.TruthCount
	}
	if result.BarCount > 0 {
		return result.BarCount
	}
	if result.NewsCount > 0 {
		return result.NewsCount
	}
	if result.InventoryCount > 0 {
		return result.InventoryCount
	}
	return 0
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
	fallbackUsed := false
	fallbackReason := ""
	effectiveSourceKey := sourceKey
	if err == nil && strings.EqualFold(sourceKey, "TUSHARE") && len(contracts) == 0 && marketSyncResultCount(result) == 0 {
		fallbackResult, fallbackErr := h.service.AdminSyncFuturesQuotes("AUTO", contracts, days)
		if fallbackErr == nil && marketSyncResultCount(fallbackResult) > 0 {
			result = fallbackResult
			fallbackUsed = true
			fallbackReason = "tushare_empty_result"
			effectiveSourceKey = strings.ToUpper(strings.TrimSpace(fallbackResult.SelectedSource))
			if effectiveSourceKey == "" {
				effectiveSourceKey = "AUTO"
			}
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	count := marketSyncResultCount(result)
	beforeValue := requestedSourceKey
	if beforeValue == "" {
		beforeValue = "DEFAULT"
	}
	reason := fmt.Sprintf("days=%d,contracts=%d", days, len(contracts))
	if fallbackUsed {
		reason += ",fallback_from=TUSHARE"
	}
	h.writeOperationLog(c, "FUTURES", "SYNC_QUOTES", "FUTURES_QUOTES", effectiveSourceKey, beforeValue, "count="+strconv.Itoa(count), reason)
	response := gin.H{
		"count":                count,
		"source_key":           effectiveSourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"contracts":            contracts,
		"result":               result,
	}
	if fallbackUsed {
		response["fallback_used"] = true
		response["fallback_source_key"] = effectiveSourceKey
		response["fallback_reason"] = fallbackReason
	}
	c.JSON(http.StatusOK, dto.OK(response))
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
	effectiveSourceKey := sourceKey
	fallbackSourceKey := ""
	result, err := h.service.AdminSyncMarketNews(sourceKey, symbols, days, limit)
	if err != nil && strings.EqualFold(sourceKey, "TUSHARE") && isTushareNewsPermissionDeniedError(err) {
		fallbackSourceKey = h.resolveDefaultConfigValue("market.news.default_source_key", "AKSHARE")
		if strings.EqualFold(fallbackSourceKey, "TUSHARE") || strings.TrimSpace(fallbackSourceKey) == "" {
			fallbackSourceKey = "AKSHARE"
		}
		result, err = h.service.AdminSyncMarketNews(fallbackSourceKey, symbols, days, limit)
		if err == nil {
			effectiveSourceKey = fallbackSourceKey
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	reason := fmt.Sprintf("days=%d,symbols=%d,limit=%d", days, len(symbols), limit)
	if fallbackSourceKey != "" && strings.EqualFold(effectiveSourceKey, fallbackSourceKey) {
		reason += fmt.Sprintf(",fallback_from=%s", sourceKey)
	}
	h.writeOperationLog(
		c,
		"NEWS",
		"SYNC_MARKET_NEWS",
		"MARKET_NEWS",
		effectiveSourceKey,
		requestedSourceKey,
		fmt.Sprintf("count=%d", result.NewsCount),
		reason,
	)
	response := gin.H{
		"count":                result.NewsCount,
		"source_key":           effectiveSourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"limit":                limit,
		"symbols":              symbols,
		"result":               result,
	}
	if fallbackSourceKey != "" && strings.EqualFold(effectiveSourceKey, fallbackSourceKey) {
		response["fallback_used"] = true
		response["fallback_source_key"] = fallbackSourceKey
		response["fallback_reason"] = "tushare_anns_d_permission_denied"
	}
	c.JSON(http.StatusOK, dto.OK(response))
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
		sourceKey = h.resolveDefaultConfigValue("stock.daily_basic.default_source_key", "AKSHARE")
		sourceKey = preferAkshareWhenDefaultRequested(requestedSourceKey, sourceKey)
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
		sourceKey = h.resolveDefaultConfigValue("stock.moneyflow.default_source_key", "AKSHARE")
		sourceKey = preferAkshareWhenDefaultRequested(requestedSourceKey, sourceKey)
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
		sourceKey = h.resolveDefaultConfigValue("stock.news.default_source_key", "AKSHARE")
		sourceKey = preferAkshareWhenDefaultRequested(requestedSourceKey, sourceKey)
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
		sourceKey = h.resolveDefaultConfigValue("stock.master.default_source_key", "AKSHARE")
		sourceKey = preferAkshareWhenDefaultRequested(requestedSourceKey, sourceKey)
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
