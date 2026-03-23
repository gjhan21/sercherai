package handler

import (
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

func (h *AdminGrowthHandler) GetMarketProviderGovernanceOverview(c *gin.Context) {
	hours := 24
	if raw := strings.TrimSpace(c.Query("hours")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "hours must be a positive integer", Data: struct{}{}})
			return
		}
		hours = parsed
	}
	item, err := h.service.AdminGetMarketProviderGovernanceOverview(c.Query("asset_class"), c.Query("data_kind"), hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) ListMarketProviderCapabilities(c *gin.Context) {
	items, err := h.service.AdminListMarketProviderCapabilities(
		c.Query("provider_key"),
		c.Query("asset_class"),
		c.Query("data_kind"),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) ListMarketProviderRoutingPolicies(c *gin.Context) {
	items, err := h.service.AdminListMarketProviderRoutingPolicies(c.Query("asset_class"), c.Query("data_kind"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) UpdateMarketProviderRoutingPolicy(c *gin.Context) {
	policyKey := strings.TrimSpace(c.Param("policy_key"))
	if policyKey == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "policy_key is required", Data: struct{}{}})
		return
	}
	var req model.MarketProviderRoutingPolicy
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if strings.TrimSpace(req.DataKind) == "" || strings.TrimSpace(req.PrimaryProviderKey) == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "data_kind and primary_provider_key are required", Data: struct{}{}})
		return
	}
	item, err := h.service.AdminUpsertMarketProviderRoutingPolicy(policyKey, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) CreateMarketDataBackfillRun(c *gin.Context) {
	var req dto.MarketDataBackfillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	item, err := h.service.AdminCreateMarketDataBackfillRun(model.MarketBackfillCreateInput{
		RunType:               req.RunType,
		AssetScope:            req.AssetScope,
		SourceKey:             req.SourceKey,
		TradeDateFrom:         req.TradeDateFrom,
		TradeDateTo:           req.TradeDateTo,
		BatchSize:             req.BatchSize,
		Stages:                req.Stages,
		ForceRefreshUniverse:  req.ForceRefreshUniverse,
		RebuildTruthAfterSync: req.RebuildTruthAfterSync,
	}, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET_DATA", "CREATE_BACKFILL_RUN", "MARKET_BACKFILL_RUN", item.ID, "", item.Status, strings.Join(item.AssetScope, ","))
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"run_id":               item.ID,
		"scheduler_run_id":     item.SchedulerRunID,
		"universe_snapshot_id": item.UniverseSnapshotID,
		"status":               item.Status,
	}))
}

func (h *AdminGrowthHandler) ListMarketDataBackfillRuns(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListMarketDataBackfillRuns(
		c.Query("status"),
		c.Query("run_type"),
		c.Query("asset_type"),
		c.Query("source_key"),
		page,
		pageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) GetMarketDataBackfillRun(c *gin.Context) {
	item, err := h.service.AdminGetMarketDataBackfillRun(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) ListMarketDataBackfillRunDetails(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListMarketDataBackfillRunDetails(
		c.Param("id"),
		c.Query("stage"),
		c.Query("asset_type"),
		c.Query("status"),
		page,
		pageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) RetryMarketDataBackfillRun(c *gin.Context) {
	var req dto.MarketDataBackfillRetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	item, err := h.service.AdminRetryMarketDataBackfillRun(c.Param("id"), model.MarketBackfillRetryInput{
		RetryMode: req.RetryMode,
		Stage:     req.Stage,
		AssetType: req.AssetType,
		BatchKeys: req.BatchKeys,
	}, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET_DATA", "RETRY_BACKFILL_RUN", "MARKET_BACKFILL_RUN", item.ID, "", item.Status, req.RetryMode)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"run_id":               item.ID,
		"scheduler_run_id":     item.SchedulerRunID,
		"universe_snapshot_id": item.UniverseSnapshotID,
		"status":               item.Status,
	}))
}

func (h *AdminGrowthHandler) ListMarketUniverseSnapshots(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListMarketUniverseSnapshots(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) GetMarketUniverseSnapshot(c *gin.Context) {
	snapshot, items, err := h.service.AdminGetMarketUniverseSnapshot(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"snapshot": snapshot, "items": items}))
}

func (h *AdminGrowthHandler) GetMarketCoverageSummary(c *gin.Context) {
	item, err := h.service.AdminGetMarketCoverageSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) SyncMarketDataMaster(c *gin.Context) {
	var req dto.MarketDataSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	snapshot, itemsByAsset, assetScope, sourceKey, err := h.resolveMarketStageSyncContext(req, "TUSHARE", operator)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	result := model.MarketSyncResult{
		DataKind:           "INSTRUMENT_MASTER",
		RequestedSourceKey: sourceKey,
		Results:            make([]model.MarketSourceSyncItemResult, 0, len(assetScope)),
	}
	for _, assetType := range assetScope {
		piece, err := h.service.AdminSyncMarketMasterDetailed(assetType, sourceKey, marketStageInstrumentKeysFromItems(itemsByAsset[assetType]))
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		mergeMarketSyncResult(&result, piece)
	}
	h.writeOperationLog(c, "MARKET_DATA", "SYNC_MASTER", "MARKET_MASTER", snapshot.ID, "", fmt.Sprintf("assets=%d", len(assetScope)), fmt.Sprintf("source=%s", sourceKey))
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"snapshot_id": snapshot.ID,
		"asset_scope": assetScope,
		"result":      result,
	}))
}

func (h *AdminGrowthHandler) SyncMarketDataQuotes(c *gin.Context) {
	var req dto.MarketDataSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	snapshot, itemsByAsset, assetScope, sourceKey, err := h.resolveMarketStageSyncContext(req, "TUSHARE", operator)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	result := model.MarketSyncResult{
		DataKind:           "DAILY_BARS",
		RequestedSourceKey: sourceKey,
		Results:            make([]model.MarketSourceSyncItemResult, 0, len(assetScope)),
	}
	for _, assetType := range assetScope {
		piece, err := h.service.AdminSyncMarketQuotesDetailed(assetType, sourceKey, marketStageInstrumentKeysFromItems(itemsByAsset[assetType]), req.Days)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		mergeMarketSyncResult(&result, piece)
	}
	h.writeOperationLog(c, "MARKET_DATA", "SYNC_QUOTES", "MARKET_QUOTES", snapshot.ID, "", fmt.Sprintf("bars=%d", result.BarCount), fmt.Sprintf("source=%s,assets=%d", sourceKey, len(assetScope)))
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"snapshot_id": snapshot.ID,
		"asset_scope": assetScope,
		"result":      result,
	}))
}

func (h *AdminGrowthHandler) SyncMarketDataDailyBasic(c *gin.Context) {
	var req dto.MarketDataSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	snapshot, itemsByAsset, assetScope, sourceKey, err := h.resolveMarketStageSyncContext(req, "TUSHARE", operator)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	result := model.MarketSyncResult{
		DataKind:           "DAILY_BASIC",
		RequestedSourceKey: sourceKey,
		Results:            make([]model.MarketSourceSyncItemResult, 0, len(assetScope)),
	}
	for _, assetType := range assetScope {
		piece, err := h.service.AdminSyncMarketDailyBasicDetailed(assetType, sourceKey, marketStageInstrumentKeysFromItems(itemsByAsset[assetType]), req.Days)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		mergeMarketSyncResult(&result, piece)
	}
	h.writeOperationLog(c, "MARKET_DATA", "SYNC_DAILY_BASIC", "MARKET_DAILY_BASIC", snapshot.ID, "", fmt.Sprintf("count=%d", result.BarCount), fmt.Sprintf("source=%s,assets=%d", sourceKey, len(assetScope)))
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"snapshot_id": snapshot.ID,
		"asset_scope": assetScope,
		"result":      result,
	}))
}

func (h *AdminGrowthHandler) SyncMarketDataMoneyflow(c *gin.Context) {
	var req dto.MarketDataSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	snapshot, itemsByAsset, assetScope, sourceKey, err := h.resolveMarketStageSyncContext(req, "TUSHARE", operator)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	result := model.MarketSyncResult{
		DataKind:           "MONEYFLOW",
		RequestedSourceKey: sourceKey,
		Results:            make([]model.MarketSourceSyncItemResult, 0, len(assetScope)),
	}
	for _, assetType := range assetScope {
		piece, err := h.service.AdminSyncMarketMoneyflowDetailed(assetType, sourceKey, marketStageInstrumentKeysFromItems(itemsByAsset[assetType]), req.Days)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		mergeMarketSyncResult(&result, piece)
	}
	h.writeOperationLog(c, "MARKET_DATA", "SYNC_MONEYFLOW", "MARKET_MONEYFLOW", snapshot.ID, "", fmt.Sprintf("count=%d", result.BarCount), fmt.Sprintf("source=%s,assets=%d", sourceKey, len(assetScope)))
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"snapshot_id": snapshot.ID,
		"asset_scope": assetScope,
		"result":      result,
	}))
}

func (h *AdminGrowthHandler) RebuildMarketDataTruth(c *gin.Context) {
	var req dto.MarketDataSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	snapshot, itemsByAsset, assetScope, sourceKey, err := h.resolveMarketStageSyncContext(req, "TUSHARE", operator)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	tradeDateFrom, tradeDateTo := normalizeMarketTruthDateRange(req)

	result := model.MarketSyncResult{
		DataKind:           "TRUTH_REBUILD",
		RequestedSourceKey: sourceKey,
		Results:            make([]model.MarketSourceSyncItemResult, 0, len(assetScope)),
	}
	for _, assetType := range assetScope {
		piece, err := h.service.AdminRebuildMarketDailyTruthDetailed(assetType, sourceKey, marketStageInstrumentKeysFromItems(itemsByAsset[assetType]), tradeDateFrom, tradeDateTo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		mergeMarketSyncResult(&result, piece)
	}
	h.writeOperationLog(c, "MARKET_DATA", "REBUILD_TRUTH", "MARKET_DAILY_BAR_TRUTH", snapshot.ID, "", fmt.Sprintf("truth=%d", result.TruthCount), fmt.Sprintf("source=%s,from=%s,to=%s", sourceKey, tradeDateFrom, tradeDateTo))
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"snapshot_id": snapshot.ID,
		"asset_scope": assetScope,
		"result":      result,
	}))
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

func normalizeMarketStageAssetType(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "STOCK":
		return "STOCK"
	case "INDEX":
		return "INDEX"
	case "ETF":
		return "ETF"
	case "LOF":
		return "LOF"
	case "CBOND":
		return "CBOND"
	default:
		return ""
	}
}

func normalizeMarketStageAssetScope(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, value := range values {
		normalized := normalizeMarketStageAssetType(value)
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

func normalizeMarketStageSymbols(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, value := range values {
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

func marketStageInstrumentKeysFromItems(items []model.MarketUniverseSnapshotItem) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		key := strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		if key == "" {
			continue
		}
		result = append(result, key)
	}
	return result
}

func mergeMarketSyncResult(total *model.MarketSyncResult, piece model.MarketSyncResult) {
	if total.AssetClass == "" {
		total.AssetClass = piece.AssetClass
	} else if piece.AssetClass != "" && total.AssetClass != piece.AssetClass {
		total.AssetClass = ""
	}
	if total.DataKind == "" {
		total.DataKind = piece.DataKind
	}
	if total.RequestedSourceKey == "" {
		total.RequestedSourceKey = piece.RequestedSourceKey
	}
	total.ResolvedSourceKeys = appendUniqueMarketSourceKeys(total.ResolvedSourceKeys, piece.ResolvedSourceKeys...)
	total.BarCount += piece.BarCount
	total.NewsCount += piece.NewsCount
	total.TruthCount += piece.TruthCount
	total.InventoryCount += piece.InventoryCount
	total.SnapshotCount += piece.SnapshotCount
	total.Results = append(total.Results, piece.Results...)
}

func appendUniqueMarketSourceKeys(base []string, values ...string) []string {
	seen := make(map[string]struct{}, len(base)+len(values))
	result := make([]string, 0, len(base)+len(values))
	for _, value := range base {
		normalized := strings.ToUpper(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	for _, value := range values {
		normalized := strings.ToUpper(strings.TrimSpace(value))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func normalizeMarketTruthDateRange(req dto.MarketDataSyncRequest) (string, string) {
	tradeDateFrom := strings.TrimSpace(req.TradeDateFrom)
	tradeDateTo := strings.TrimSpace(req.TradeDateTo)
	if tradeDateFrom != "" || tradeDateTo != "" {
		return tradeDateFrom, tradeDateTo
	}
	if req.Days <= 0 {
		return "", ""
	}
	to := time.Now()
	from := to.AddDate(0, 0, -(req.Days - 1))
	return from.Format("2006-01-02"), to.Format("2006-01-02")
}

func (h *AdminGrowthHandler) resolveMarketStageSyncContext(req dto.MarketDataSyncRequest, defaultSourceKey string, operator string) (model.MarketUniverseSnapshot, map[string][]model.MarketUniverseSnapshotItem, []string, string, error) {
	normalizedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	if normalizedSourceKey == "" {
		normalizedSourceKey = strings.ToUpper(strings.TrimSpace(defaultSourceKey))
	}

	var (
		snapshot model.MarketUniverseSnapshot
		items    []model.MarketUniverseSnapshotItem
		err      error
	)
	if strings.TrimSpace(req.UniverseSnapshotID) != "" {
		snapshot, items, err = h.service.AdminGetMarketUniverseSnapshot(strings.TrimSpace(req.UniverseSnapshotID))
	} else {
		scope := normalizeMarketStageAssetScope(req.AssetScope)
		if len(scope) == 0 {
			return model.MarketUniverseSnapshot{}, nil, nil, "", fmt.Errorf("universe_snapshot_id or asset_scope is required")
		}
		snapshot, items, err = h.service.AdminBuildMarketUniverseSnapshot(normalizedSourceKey, scope, operator)
	}
	if err != nil {
		return model.MarketUniverseSnapshot{}, nil, nil, "", err
	}
	if normalizedSourceKey == "" {
		normalizedSourceKey = strings.ToUpper(strings.TrimSpace(snapshot.SourceKey))
	}

	effectiveScope := normalizeMarketStageAssetScope(req.AssetScope)
	if len(effectiveScope) == 0 {
		effectiveScope = normalizeMarketStageAssetScope(snapshot.Scope)
	}
	symbolFilters := normalizeMarketStageSymbols(req.Symbols)
	symbolSet := make(map[string]struct{}, len(symbolFilters))
	for _, symbol := range symbolFilters {
		symbolSet[symbol] = struct{}{}
	}

	grouped := make(map[string][]model.MarketUniverseSnapshotItem, len(effectiveScope))
	finalScope := make([]string, 0, len(effectiveScope))
	for _, assetType := range effectiveScope {
		filtered := make([]model.MarketUniverseSnapshotItem, 0)
		for _, item := range items {
			if item.AssetType != assetType {
				continue
			}
			if len(symbolSet) > 0 {
				instrumentKey := strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
				externalSymbol := strings.ToUpper(strings.TrimSpace(item.ExternalSymbol))
				if _, ok := symbolSet[instrumentKey]; !ok {
					if _, ok := symbolSet[externalSymbol]; !ok {
						continue
					}
				}
			}
			filtered = append(filtered, item)
		}
		if len(filtered) == 0 {
			continue
		}
		grouped[assetType] = filtered
		finalScope = append(finalScope, assetType)
	}
	if len(finalScope) == 0 {
		return model.MarketUniverseSnapshot{}, nil, nil, "", fmt.Errorf("no target instruments resolved")
	}
	return snapshot, grouped, finalScope, normalizedSourceKey, nil
}
