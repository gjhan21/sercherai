package repo

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

const (
	marketDataKindDailyBasic               = "DAILY_BASIC"
	marketDataKindMoneyflow                = "MONEYFLOW"
	marketStockDailyBasicPriorityConfigKey = "market.stock.daily_basic.source_priority"
	marketStockMoneyflowPriorityConfigKey  = "market.stock.moneyflow.source_priority"
)

func marketAssetEnhancementSupport(assetType string) (dailyBasic bool, moneyflow bool) {
	switch normalizeUniverseAssetType(assetType) {
	case "STOCK":
		return true, true
	default:
		return false, false
	}
}

func normalizeMarketEnhancementAssetType(assetType string) string {
	if strings.TrimSpace(assetType) == "" {
		return "STOCK"
	}
	return normalizeUniverseAssetType(assetType)
}

func normalizeMarketBackfillWindowDays(tradeDateFrom string, tradeDateTo string) int {
	const defaultDays = 120
	from := strings.TrimSpace(tradeDateFrom)
	to := strings.TrimSpace(tradeDateTo)
	if from == "" || to == "" {
		return defaultDays
	}
	fromTime, err := time.ParseInLocation("2006-01-02", from, time.Local)
	if err != nil {
		return defaultDays
	}
	toTime, err := time.ParseInLocation("2006-01-02", to, time.Local)
	if err != nil || toTime.Before(fromTime) {
		return defaultDays
	}
	days := int(toTime.Sub(fromTime).Hours()/24) + 1
	if days <= 0 {
		return defaultDays
	}
	if days > 365 {
		return 365
	}
	return days
}

func (r *MySQLGrowthRepo) AdminSyncMarketDailyBasicDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error) {
	return r.syncMarketEnhancementDetailed(assetType, sourceKey, instrumentKeys, days, marketDataKindDailyBasic)
}

func (r *MySQLGrowthRepo) AdminSyncMarketMoneyflowDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error) {
	return r.syncMarketEnhancementDetailed(assetType, sourceKey, instrumentKeys, days, marketDataKindMoneyflow)
}

func (r *MySQLGrowthRepo) syncMarketEnhancementDetailed(assetType string, sourceKey string, instrumentKeys []string, days int, dataKind string) (model.MarketSyncResult, error) {
	normalizedAssetType := normalizeMarketEnhancementAssetType(assetType)
	if normalizedAssetType == "" {
		return model.MarketSyncResult{}, fmt.Errorf("asset_type is required")
	}
	result := model.MarketSyncResult{
		AssetClass:         normalizedAssetType,
		DataKind:           dataKind,
		RequestedSourceKey: strings.ToUpper(strings.TrimSpace(sourceKey)),
		Results:            make([]model.MarketSourceSyncItemResult, 0, 1),
	}

	dailyBasicSupported, moneyflowSupported := marketAssetEnhancementSupport(normalizedAssetType)
	supported := (dataKind == marketDataKindDailyBasic && dailyBasicSupported) || (dataKind == marketDataKindMoneyflow && moneyflowSupported)
	if !supported {
		skippedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
		if skippedSourceKey == "" {
			skippedSourceKey = "AUTO"
		}
		result.Results = append(result.Results, model.MarketSourceSyncItemResult{
			SourceKey: skippedSourceKey,
			Status:    "SKIPPED",
			Message:   fmt.Sprintf("%s does not support %s in phase 1", normalizedAssetType, strings.ToLower(dataKind)),
		})
		return result, nil
	}

	normalizedSymbols := normalizeStockSymbolList(instrumentKeys)
	if len(normalizedSymbols) == 0 {
		normalizedSymbols = defaultMockStockSymbols()
	}
	if days <= 0 {
		days = 120
	}
	if days > 365 {
		days = 365
	}

	if err := r.syncMarketInstrumentMasterData(marketAssetClassStock, sourceKey, normalizedSymbols); err != nil {
		return result, err
	}

	routeConfigKey := marketStockDailyBasicPriorityConfigKey
	if dataKind == marketDataKindMoneyflow {
		routeConfigKey = marketStockMoneyflowPriorityConfigKey
	}
	sourceKeys := r.resolveRequestedMarketSourceKeys(sourceKey, routeConfigKey, []string{"TUSHARE"})
	result.ResolvedSourceKeys = sourceKeys

	totalCount := 0
	totalSnapshots := 0
	successes := 0
	failures := make([]string, 0)

	for _, resolvedSourceKey := range sourceKeys {
		sourceItem, err := r.getDataSourceBySourceKey(resolvedSourceKey)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
			result.Results = append(result.Results, model.MarketSourceSyncItemResult{
				SourceKey: resolvedSourceKey,
				Status:    "FAILED",
				Message:   err.Error(),
			})
			continue
		}

		itemsCount, payload, message, err := r.runStockEnhancementSyncForSource(sourceItem, resolvedSourceKey, normalizedSymbols, days, dataKind)
		status := "SUCCESS"
		if err != nil {
			status = "FAILED"
			message = err.Error()
			failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
		} else {
			successes++
			totalCount += itemsCount
		}
		if snapshotErr := r.insertMarketSourceSnapshot(
			resolvedSourceKey,
			marketAssetClassStock,
			dataKind,
			"",
			"",
			status,
			message,
			payload,
			time.Now(),
		); snapshotErr == nil {
			totalSnapshots++
		}
		result.Results = append(result.Results, model.MarketSourceSyncItemResult{
			SourceKey:     resolvedSourceKey,
			Status:        status,
			BarCount:      itemsCount,
			SnapshotCount: 1,
			Message:       message,
		})
	}

	result.BarCount = totalCount
	result.SnapshotCount = totalSnapshots
	if successes == 0 && len(failures) > 0 {
		return result, errors.New(strings.Join(failures, "; "))
	}
	return result, nil
}

func (r *MySQLGrowthRepo) runStockEnhancementSyncForSource(sourceItem model.DataSource, resolvedSourceKey string, symbols []string, days int, dataKind string) (int, string, string, error) {
	provider := strings.ToUpper(parseDataSourceStringConfig(sourceItem.Config, "provider", "vendor"))
	if provider == "" {
		provider = strings.ToUpper(strings.TrimSpace(sourceItem.SourceKey))
	}
	sourceKey := canonicalMarketSourceKey(resolvedSourceKey, provider)
	timeoutMS := parseDataSourceTimeoutMS(sourceItem.Config)

	switch dataKind {
	case marketDataKindDailyBasic:
		var items []stockDailyBasicPoint
		var err error
		switch provider {
		case "TUSHARE":
			token := strings.TrimSpace(parseDataSourceStringConfig(sourceItem.Config, "token", "api_token", "tushare_token"))
			if token == "" {
				token = strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
			}
			items, err = fetchStockDailyBasicsFromTushare(token, sourceKey, symbols, days, timeoutMS)
		case "MOCK":
			items = buildMockStockDailyBasics(sourceKey, symbols, days)
		default:
			err = fmt.Errorf("unsupported daily_basic provider: %s", provider)
		}
		if err != nil {
			return 0, "", "", err
		}
		count, err := r.upsertStockDailyBasics(items)
		payload := marshalJSONSilently(map[string]any{
			"source_key": sourceKey,
			"data_kind":  dataKind,
			"symbol_n":   len(symbols),
			"days":       days,
			"item_count": len(items),
		})
		return count, payload, "ok", err
	case marketDataKindMoneyflow:
		var items []stockMoneyflowPoint
		var err error
		switch provider {
		case "TUSHARE":
			token := strings.TrimSpace(parseDataSourceStringConfig(sourceItem.Config, "token", "api_token", "tushare_token"))
			if token == "" {
				token = strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
			}
			items, err = fetchStockMoneyflowsFromTushare(token, sourceKey, symbols, days, timeoutMS)
		case "MOCK":
			items = buildMockStockMoneyflows(sourceKey, symbols, days)
		default:
			err = fmt.Errorf("unsupported moneyflow provider: %s", provider)
		}
		if err != nil {
			return 0, "", "", err
		}
		count, err := r.upsertStockMoneyflows(items)
		payload := marshalJSONSilently(map[string]any{
			"source_key": sourceKey,
			"data_kind":  dataKind,
			"symbol_n":   len(symbols),
			"days":       days,
			"item_count": len(items),
		})
		return count, payload, "ok", err
	default:
		return 0, "", "", fmt.Errorf("unsupported data kind: %s", dataKind)
	}
}

func buildMockStockDailyBasics(sourceKey string, symbols []string, days int) []stockDailyBasicPoint {
	if days <= 0 {
		days = 30
	}
	start := time.Now().AddDate(0, 0, -(days - 1))
	items := make([]stockDailyBasicPoint, 0, len(symbols)*days)
	for symbolIndex, symbol := range symbols {
		for dayIndex := 0; dayIndex < days; dayIndex++ {
			tradeDate := start.AddDate(0, 0, dayIndex)
			items = append(items, stockDailyBasicPoint{
				Symbol:       strings.ToUpper(strings.TrimSpace(symbol)),
				TradeDate:    tradeDate,
				TurnoverRate: roundTo(1.2+float64(symbolIndex)*0.1+float64(dayIndex)*0.02, 4),
				VolumeRatio:  roundTo(0.9+float64(symbolIndex)*0.05, 4),
				PeTTM:        roundTo(12.5+float64(symbolIndex), 4),
				PB:           roundTo(1.5+float64(symbolIndex)*0.08, 4),
				TotalMV:      roundTo(1200+float64(symbolIndex)*120, 4),
				CircMV:       roundTo(900+float64(symbolIndex)*90, 4),
				SourceKey:    sourceKey,
			})
		}
	}
	return items
}

func buildMockStockMoneyflows(sourceKey string, symbols []string, days int) []stockMoneyflowPoint {
	if days <= 0 {
		days = 30
	}
	start := time.Now().AddDate(0, 0, -(days - 1))
	items := make([]stockMoneyflowPoint, 0, len(symbols)*days)
	for symbolIndex, symbol := range symbols {
		for dayIndex := 0; dayIndex < days; dayIndex++ {
			base := float64(100 + symbolIndex*10 + dayIndex)
			tradeDate := start.AddDate(0, 0, dayIndex)
			items = append(items, stockMoneyflowPoint{
				Symbol:        strings.ToUpper(strings.TrimSpace(symbol)),
				TradeDate:     tradeDate,
				NetMFAmount:   roundTo(base*1.3, 4),
				BuyLGAmount:   roundTo(base*0.8, 4),
				SellLGAmount:  roundTo(base*0.4, 4),
				BuyELGAmount:  roundTo(base*0.6, 4),
				SellELGAmount: roundTo(base*0.3, 4),
				SourceKey:     sourceKey,
			})
		}
	}
	return items
}

func (r *InMemoryGrowthRepo) AdminSyncMarketDailyBasicDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error) {
	return buildInMemoryMarketEnhancementSyncResult(assetType, sourceKey, instrumentKeys, days, marketDataKindDailyBasic), nil
}

func (r *InMemoryGrowthRepo) AdminSyncMarketMoneyflowDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error) {
	return buildInMemoryMarketEnhancementSyncResult(assetType, sourceKey, instrumentKeys, days, marketDataKindMoneyflow), nil
}

func buildInMemoryMarketEnhancementSyncResult(assetType string, sourceKey string, instrumentKeys []string, days int, dataKind string) model.MarketSyncResult {
	normalizedAssetType := normalizeMarketEnhancementAssetType(assetType)
	if normalizedAssetType == "" {
		normalizedAssetType = "STOCK"
	}
	normalizedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	if normalizedSourceKey == "" {
		normalizedSourceKey = "TUSHARE"
	}
	result := model.MarketSyncResult{
		AssetClass:         normalizedAssetType,
		DataKind:           dataKind,
		RequestedSourceKey: normalizedSourceKey,
		ResolvedSourceKeys: []string{normalizedSourceKey},
		Results:            make([]model.MarketSourceSyncItemResult, 0, 1),
	}
	dailyBasicSupported, moneyflowSupported := marketAssetEnhancementSupport(normalizedAssetType)
	supported := (dataKind == marketDataKindDailyBasic && dailyBasicSupported) || (dataKind == marketDataKindMoneyflow && moneyflowSupported)
	if !supported {
		result.Results = append(result.Results, model.MarketSourceSyncItemResult{
			SourceKey: normalizedSourceKey,
			Status:    "SKIPPED",
			Message:   fmt.Sprintf("%s does not support %s in phase 1", normalizedAssetType, strings.ToLower(dataKind)),
		})
		return result
	}
	normalizedSymbols := normalizeStockSymbolList(instrumentKeys)
	if len(normalizedSymbols) == 0 {
		normalizedSymbols = defaultMockStockSymbols()
	}
	if days <= 0 {
		days = 120
	}
	count := len(normalizedSymbols) * days
	result.BarCount = count
	result.Results = append(result.Results, model.MarketSourceSyncItemResult{
		SourceKey: normalizedSourceKey,
		Status:    "SUCCESS",
		BarCount:  count,
		Message:   "in-memory enhancement sync",
	})
	return result
}

func (r *InMemoryGrowthRepo) executeMarketDataBackfillRun(runID string) (model.MarketBackfillRun, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, ok := r.marketBackfillRuns[strings.TrimSpace(runID)]
	if !ok {
		return model.MarketBackfillRun{}, fmt.Errorf("backfill run not found")
	}
	snapshotItems := append([]model.MarketUniverseSnapshotItem(nil), r.marketUniverseItems[run.UniverseSnapshotID]...)
	byAsset := groupMarketUniverseItemsByAsset(snapshotItems)
	now := time.Now()
	nowText := now.Format(time.RFC3339)
	windowDays := normalizeMarketBackfillWindowDays(run.TradeDateFrom, run.TradeDateTo)
	assetScope := normalizeMarketBackfillAssetScope(run.AssetScope)
	if len(assetScope) == 0 {
		assetScope = orderedMarketUniverseAssets(byAsset)
	}

	details := append([]model.MarketBackfillRunDetail(nil), r.marketBackfillRunDetails[run.ID]...)
	progress := run.StageProgress

	masterDetails := make([]model.MarketBackfillRunDetail, 0, len(assetScope))
	for _, assetType := range assetScope {
		items := byAsset[assetType]
		masterDetails = append(masterDetails, newBackfillStageDetail(run, "MASTER", assetType, run.SourceKey, items, len(items), len(items), 0, "SUCCESS", nowText, "master synchronized"))
	}
	details = append(details, masterDetails...)
	progress = updateMarketBackfillStageProgress(progress, "MASTER", "SUCCESS", len(assetScope), len(assetScope), 0, 0)

	quotesDetails := make([]model.MarketBackfillRunDetail, 0, len(assetScope))
	for _, assetType := range assetScope {
		items := byAsset[assetType]
		quotesCount := len(items) * windowDays
		quotesDetails = append(quotesDetails, newBackfillStageDetail(run, "QUOTES", assetType, run.SourceKey, items, quotesCount, quotesCount, quotesCount, "SUCCESS", nowText, "quotes synchronized"))
	}
	details = append(details, quotesDetails...)
	progress = updateMarketBackfillStageProgress(progress, "QUOTES", "SUCCESS", len(assetScope), len(assetScope), 0, 0)

	dailyBasicDetails := make([]model.MarketBackfillRunDetail, 0, len(assetScope))
	dailyBasicCompleted := 0
	dailyBasicSkipped := 0
	for _, assetType := range assetScope {
		items := byAsset[assetType]
		syncResult := buildInMemoryMarketEnhancementSyncResult(assetType, run.SourceKey, universeItemsToInstrumentKeys(items), windowDays, marketDataKindDailyBasic)
		status := syncResult.Results[0].Status
		if status == "SUCCESS" {
			dailyBasicCompleted++
		}
		if status == "SKIPPED" {
			dailyBasicSkipped++
		}
		dailyBasicDetails = append(dailyBasicDetails, newBackfillStageDetail(run, "DAILY_BASIC", assetType, run.SourceKey, items, syncResult.BarCount, syncResult.BarCount, 0, status, nowText, syncResult.Results[0].Message))
	}
	details = append(details, dailyBasicDetails...)
	progress = updateMarketBackfillStageProgress(progress, "DAILY_BASIC", "SUCCESS", len(assetScope), dailyBasicCompleted, 0, dailyBasicSkipped)

	moneyflowDetails := make([]model.MarketBackfillRunDetail, 0, len(assetScope))
	moneyflowCompleted := 0
	moneyflowSkipped := 0
	for _, assetType := range assetScope {
		items := byAsset[assetType]
		syncResult := buildInMemoryMarketEnhancementSyncResult(assetType, run.SourceKey, universeItemsToInstrumentKeys(items), windowDays, marketDataKindMoneyflow)
		status := syncResult.Results[0].Status
		if status == "SUCCESS" {
			moneyflowCompleted++
		}
		if status == "SKIPPED" {
			moneyflowSkipped++
		}
		moneyflowDetails = append(moneyflowDetails, newBackfillStageDetail(run, "MONEYFLOW", assetType, run.SourceKey, items, syncResult.BarCount, syncResult.BarCount, 0, status, nowText, syncResult.Results[0].Message))
	}
	details = append(details, moneyflowDetails...)
	progress = updateMarketBackfillStageProgress(progress, "MONEYFLOW", "SUCCESS", len(assetScope), moneyflowCompleted, 0, moneyflowSkipped)

	truthDetails := make([]model.MarketBackfillRunDetail, 0, len(assetScope))
	for _, assetType := range assetScope {
		items := byAsset[assetType]
		truthCount := len(items) * windowDays
		truthDetails = append(truthDetails, newBackfillStageDetail(run, "TRUTH", assetType, run.SourceKey, items, 0, 0, truthCount, "SUCCESS", nowText, "truth rebuilt"))
	}
	details = append(details, truthDetails...)
	progress = updateMarketBackfillStageProgress(progress, "TRUTH", "SUCCESS", len(assetScope), len(assetScope), 0, 0)

	details = append(details, model.MarketBackfillRunDetail{
		ID:             "mbd_" + strings.ToLower(strings.ReplaceAll(newID("detail"), "_", "")),
		RunID:          run.ID,
		SchedulerRunID: run.SchedulerRunID,
		Stage:          "COVERAGE_SUMMARY",
		BatchKey:       "COVERAGE-SUMMARY-001",
		SourceKey:      run.SourceKey,
		Status:         "SUCCESS",
		WarningText:    "coverage summary refreshed",
		StartedAt:      nowText,
		FinishedAt:     nowText,
		CreatedAt:      nowText,
		UpdatedAt:      nowText,
	})
	progress = updateMarketBackfillStageProgress(progress, "COVERAGE_SUMMARY", "SUCCESS", 1, 1, 0, 0)

	run.Status = "SUCCESS"
	run.CurrentStage = "COVERAGE_SUMMARY"
	run.StageProgress = progress
	if run.Summary == nil {
		run.Summary = make(map[string]any)
	}
	run.Summary["executed"] = true
	run.Summary["window_days"] = windowDays
	run.UpdatedAt = nowText
	run.FinishedAt = nowText
	r.marketBackfillRuns[run.ID] = run
	r.marketBackfillRunDetails[run.ID] = details
	return run, nil
}

func groupMarketUniverseItemsByAsset(items []model.MarketUniverseSnapshotItem) map[string][]model.MarketUniverseSnapshotItem {
	result := make(map[string][]model.MarketUniverseSnapshotItem)
	for _, item := range items {
		result[item.AssetType] = append(result[item.AssetType], item)
	}
	return result
}

func universeItemsToInstrumentKeys(items []model.MarketUniverseSnapshotItem) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.InstrumentKey) == "" {
			continue
		}
		result = append(result, item.InstrumentKey)
	}
	return result
}

func orderedMarketUniverseAssets(items map[string][]model.MarketUniverseSnapshotItem) []string {
	assets := make([]string, 0, len(items))
	for assetType := range items {
		assets = append(assets, assetType)
	}
	return normalizeMarketBackfillAssetScope(assets)
}

func updateMarketBackfillStageProgress(progress []model.MarketBackfillStageProgress, stage string, status string, total int, completed int, failed int, skipped int) []model.MarketBackfillStageProgress {
	for idx := range progress {
		if progress[idx].Stage != stage {
			continue
		}
		progress[idx].Status = status
		progress[idx].TotalBatches = total
		progress[idx].CompletedBatches = completed
		progress[idx].FailedBatches = failed
		progress[idx].SkippedBatches = skipped
		return progress
	}
	return append(progress, model.MarketBackfillStageProgress{
		Stage:            stage,
		Status:           status,
		TotalBatches:     total,
		CompletedBatches: completed,
		FailedBatches:    failed,
		SkippedBatches:   skipped,
	})
}

func newBackfillStageDetail(run model.MarketBackfillRun, stage string, assetType string, sourceKey string, items []model.MarketUniverseSnapshotItem, fetchedCount int, upsertedCount int, truthCount int, status string, nowText string, message string) model.MarketBackfillRunDetail {
	sample := make([]string, 0, minInt(len(items), 3))
	for idx := 0; idx < len(items) && idx < 3; idx++ {
		sample = append(sample, items[idx].InstrumentKey)
	}
	detail := model.MarketBackfillRunDetail{
		ID:             "mbd_" + strings.ToLower(strings.ReplaceAll(newID("detail"), "_", "")),
		RunID:          run.ID,
		SchedulerRunID: run.SchedulerRunID,
		Stage:          stage,
		AssetType:      assetType,
		BatchKey:       fmt.Sprintf("%s-%s-001", stage, assetType),
		SourceKey:      sourceKey,
		SymbolCount:    len(items),
		SymbolSample:   sample,
		Status:         status,
		FetchedCount:   fetchedCount,
		UpsertedCount:  upsertedCount,
		TruthCount:     truthCount,
		StartedAt:      nowText,
		FinishedAt:     nowText,
		CreatedAt:      nowText,
		UpdatedAt:      nowText,
	}
	if status == "SKIPPED" {
		detail.WarningText = message
	}
	return detail
}
