package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

const (
	marketDataKindInstrumentMaster         = "INSTRUMENT_MASTER"
	marketDataKindDailyBasic               = "DAILY_BASIC"
	marketDataKindMoneyflow                = "MONEYFLOW"
	marketDataKindTruthRebuild             = "TRUTH_REBUILD"
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

func normalizeMarketStageSyncAssetType(assetType string) string {
	return normalizeUniverseAssetType(assetType)
}

func normalizeMarketStageInstrumentKeys(assetType string, instrumentKeys []string) []string {
	normalizedAssetType := normalizeMarketStageSyncAssetType(assetType)
	if normalizedAssetType == "STOCK" {
		return normalizeStockSymbolList(instrumentKeys)
	}
	seen := make(map[string]struct{}, len(instrumentKeys))
	items := make([]string, 0, len(instrumentKeys))
	for _, value := range instrumentKeys {
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

func marketStageSyncPriority(assetType string) (string, []string) {
	if normalizeMarketStageSyncAssetType(assetType) == "STOCK" {
		return marketStockPriorityConfigKey, []string{"TUSHARE", "AKSHARE", "TICKERMD", "MOCK"}
	}
	return marketStockPriorityConfigKey, []string{"TUSHARE", "AKSHARE", "TICKERMD", "MOCK"}
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

func (r *MySQLGrowthRepo) AdminSyncMarketMasterDetailed(assetType string, sourceKey string, instrumentKeys []string) (model.MarketSyncResult, error) {
	normalizedAssetType := normalizeMarketStageSyncAssetType(assetType)
	if normalizedAssetType == "" {
		return model.MarketSyncResult{}, fmt.Errorf("asset_type is required")
	}
	normalizedKeys := normalizeMarketStageInstrumentKeys(normalizedAssetType, instrumentKeys)
	if len(normalizedKeys) == 0 {
		return model.MarketSyncResult{}, fmt.Errorf("instrument_keys is required")
	}
	normalizedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	if normalizedSourceKey == "" {
		normalizedSourceKey = "TUSHARE"
	}
	if err := r.syncMarketInstrumentMasterData(normalizedAssetType, normalizedSourceKey, normalizedKeys); err != nil {
		return model.MarketSyncResult{}, err
	}
	return model.MarketSyncResult{
		AssetClass:         normalizedAssetType,
		DataKind:           marketDataKindInstrumentMaster,
		RequestedSourceKey: normalizedSourceKey,
		ResolvedSourceKeys: []string{normalizedSourceKey},
		SnapshotCount:      len(normalizedKeys),
		Results: []model.MarketSourceSyncItemResult{
			{
				SourceKey:     normalizedSourceKey,
				Status:        "SUCCESS",
				SnapshotCount: len(normalizedKeys),
				Message:       "master synchronized",
			},
		},
	}, nil
}

func (r *MySQLGrowthRepo) AdminSyncMarketQuotesDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error) {
	normalizedAssetType := normalizeMarketStageSyncAssetType(assetType)
	if normalizedAssetType == "" {
		return model.MarketSyncResult{}, fmt.Errorf("asset_type is required")
	}
	normalizedKeys := normalizeMarketStageInstrumentKeys(normalizedAssetType, instrumentKeys)
	if len(normalizedKeys) == 0 {
		return model.MarketSyncResult{}, fmt.Errorf("instrument_keys is required")
	}
	routeConfigKey, defaultPriority := marketStageSyncPriority(normalizedAssetType)
	return r.syncMarketDailyBars(normalizedAssetType, sourceKey, normalizedKeys, days, routeConfigKey, defaultPriority)
}

func (r *MySQLGrowthRepo) AdminSyncMarketDailyBasicDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error) {
	return r.syncMarketEnhancementDetailed(assetType, sourceKey, instrumentKeys, days, marketDataKindDailyBasic)
}

func (r *MySQLGrowthRepo) AdminSyncMarketMoneyflowDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error) {
	return r.syncMarketEnhancementDetailed(assetType, sourceKey, instrumentKeys, days, marketDataKindMoneyflow)
}

func (r *MySQLGrowthRepo) AdminRebuildMarketDailyTruthDetailed(assetType string, sourceKey string, instrumentKeys []string, tradeDateFrom string, tradeDateTo string) (model.MarketSyncResult, error) {
	normalizedAssetType := normalizeMarketStageSyncAssetType(assetType)
	if normalizedAssetType == "" {
		return model.MarketSyncResult{}, fmt.Errorf("asset_type is required")
	}
	normalizedKeys := normalizeMarketStageInstrumentKeys(normalizedAssetType, instrumentKeys)
	if len(normalizedKeys) == 0 {
		return model.MarketSyncResult{}, fmt.Errorf("instrument_keys is required")
	}
	normalizedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	routeConfigKey, defaultPriority := marketStageSyncPriority(normalizedAssetType)
	priority := defaultPriority
	if normalizedSourceKey != "" {
		priority = []string{normalizedSourceKey}
	} else {
		priority = r.loadMarketSourcePriority(routeConfigKey, defaultPriority)
	}
	result := model.MarketSyncResult{
		AssetClass:         normalizedAssetType,
		DataKind:           marketDataKindTruthRebuild,
		RequestedSourceKey: normalizedSourceKey,
		ResolvedSourceKeys: priority,
		Results:            make([]model.MarketSourceSyncItemResult, 0, 1),
	}
	touched, err := r.loadTouchedBarKeysForTruthRebuild(normalizedAssetType, normalizedKeys, tradeDateFrom, tradeDateTo)
	if err != nil {
		return result, err
	}
	resultSourceKey := normalizedSourceKey
	if resultSourceKey == "" && len(priority) > 0 {
		resultSourceKey = priority[0]
	}
	if len(touched) == 0 {
		result.Results = append(result.Results, model.MarketSourceSyncItemResult{
			SourceKey: resultSourceKey,
			Status:    "SUCCESS",
			Message:   "no bars matched current truth rebuild scope",
		})
		return result, nil
	}
	selectedBars, err := r.rebuildMarketDailyBarTruth(normalizedAssetType, touched, priority)
	if err != nil {
		return result, err
	}
	result.TruthCount = len(selectedBars)
	result.Results = append(result.Results, model.MarketSourceSyncItemResult{
		SourceKey:  resultSourceKey,
		Status:     "SUCCESS",
		TruthCount: len(selectedBars),
		Message:    "truth rebuilt",
	})
	return result, nil
}

func (r *MySQLGrowthRepo) loadTouchedBarKeysForTruthRebuild(assetType string, instrumentKeys []string, tradeDateFrom string, tradeDateTo string) (map[string]marketTouchedBarKey, error) {
	normalizedAssetType := normalizeMarketStageSyncAssetType(assetType)
	normalizedKeys := normalizeMarketStageInstrumentKeys(normalizedAssetType, instrumentKeys)
	if normalizedAssetType == "" || len(normalizedKeys) == 0 {
		return map[string]marketTouchedBarKey{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(normalizedKeys)), ",")
	args := make([]any, 0, len(normalizedKeys)+3)
	args = append(args, normalizedAssetType)
	for _, instrumentKey := range normalizedKeys {
		args = append(args, instrumentKey)
	}
	query := fmt.Sprintf(`
SELECT DISTINCT instrument_key, DATE_FORMAT(trade_date, '%%Y-%%m-%%d')
FROM market_daily_bars
WHERE asset_class = ? AND instrument_key IN (%s)`, placeholders)
	if trimmed := strings.TrimSpace(tradeDateFrom); trimmed != "" {
		query += " AND trade_date >= ?"
		args = append(args, trimmed)
	}
	if trimmed := strings.TrimSpace(tradeDateTo); trimmed != "" {
		query += " AND trade_date <= ?"
		args = append(args, trimmed)
	}
	query += "\nORDER BY instrument_key ASC, trade_date ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]marketTouchedBarKey)
	for rows.Next() {
		var instrumentKey string
		var tradeDate string
		if err := rows.Scan(&instrumentKey, &tradeDate); err != nil {
			return nil, err
		}
		instrumentKey = strings.ToUpper(strings.TrimSpace(instrumentKey))
		tradeDate = strings.TrimSpace(tradeDate)
		if instrumentKey == "" || tradeDate == "" {
			continue
		}
		result[instrumentKey+"|"+tradeDate] = marketTouchedBarKey{
			InstrumentKey: instrumentKey,
			TradeDate:     tradeDate,
		}
	}
	return result, rows.Err()
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

func (r *MySQLGrowthRepo) executeMarketDataBackfillRun(runID string) (model.MarketBackfillRun, error) {
	run, snapshotItems, err := r.loadMarketBackfillExecutionContext(runID)
	if err != nil {
		return model.MarketBackfillRun{}, err
	}

	byAsset := groupMarketUniverseItemsByAsset(snapshotItems)
	assetScope := normalizeMarketBackfillAssetScope(run.AssetScope)
	if len(assetScope) == 0 {
		assetScope = orderedMarketUniverseAssets(byAsset)
	}
	windowDays := normalizeMarketBackfillWindowDays(run.TradeDateFrom, run.TradeDateTo)
	progress := run.StageProgress
	if len(progress) == 0 {
		progress = buildMarketBackfillStageProgressAfterUniverse(assetScope)
	}
	if run.Summary == nil {
		run.Summary = make(map[string]any)
	}

	nowText := time.Now().Format(time.RFC3339)

	masterDetails, err := r.runMarketMasterStage(run, byAsset, assetScope, nowText)
	if err != nil {
		return r.failMarketBackfillRun(run, progress, "MASTER", err)
	}
	progress = updateMarketBackfillProgressFromDetails(progress, "MASTER", masterDetails)

	quotesDetails, quoteTruthCounts, quoteTouchedByAsset, err := r.runMarketQuotesStage(run, byAsset, assetScope, windowDays, nowText)
	if err != nil {
		return r.failMarketBackfillRun(run, progress, "QUOTES", err)
	}
	progress = updateMarketBackfillProgressFromDetails(progress, "QUOTES", quotesDetails)

	dailyBasicDetails, err := r.runMarketDailyBasicStage(run, byAsset, assetScope, windowDays, nowText)
	if err != nil {
		return r.failMarketBackfillRun(run, progress, "DAILY_BASIC", err)
	}
	progress = updateMarketBackfillProgressFromDetails(progress, "DAILY_BASIC", dailyBasicDetails)

	moneyflowDetails, err := r.runMarketMoneyflowStage(run, byAsset, assetScope, windowDays, nowText)
	if err != nil {
		return r.failMarketBackfillRun(run, progress, "MONEYFLOW", err)
	}
	progress = updateMarketBackfillProgressFromDetails(progress, "MONEYFLOW", moneyflowDetails)

	truthDetails, err := r.runMarketTruthStage(run, byAsset, assetScope, quoteTruthCounts, quoteTouchedByAsset, nowText)
	if err != nil {
		return r.failMarketBackfillRun(run, progress, "TRUTH", err)
	}
	progress = updateMarketBackfillProgressFromDetails(progress, "TRUTH", truthDetails)

	coverageDetail, err := r.finalizeMarketCoverageSummaryStage(run, nowText)
	if err != nil {
		return r.failMarketBackfillRun(run, progress, "COVERAGE_SUMMARY", err)
	}
	progress = updateMarketBackfillProgressFromDetails(progress, "COVERAGE_SUMMARY", []model.MarketBackfillRunDetail{coverageDetail})

	run.Status = "SUCCESS"
	run.CurrentStage = "COVERAGE_SUMMARY"
	run.StageProgress = progress
	run.Summary["executed"] = true
	run.Summary["window_days"] = windowDays
	run.Summary["asset_scope"] = assetScope
	run.Summary["stage_detail_count"] = len(masterDetails) + len(quotesDetails) + len(dailyBasicDetails) + len(moneyflowDetails) + len(truthDetails) + 1
	run.UpdatedAt = nowText
	run.FinishedAt = nowText
	run.ErrorMessage = ""
	if err := r.updateMarketBackfillRunExecutionState(run); err != nil {
		return model.MarketBackfillRun{}, err
	}
	return run, nil
}

func (r *MySQLGrowthRepo) loadMarketBackfillExecutionContext(runID string) (model.MarketBackfillRun, []model.MarketUniverseSnapshotItem, error) {
	var (
		run               model.MarketBackfillRun
		assetScopeJSON    string
		stageProgressJSON string
		summaryJSON       string
		finishedAt        sql.NullTime
		createdAt         time.Time
		updatedAt         time.Time
	)
	err := r.db.QueryRow(`
SELECT id, scheduler_run_id, run_type, COALESCE(CAST(asset_scope AS CHAR), ''),
       COALESCE(DATE_FORMAT(trade_date_from, '%Y-%m-%d'), ''),
       COALESCE(DATE_FORMAT(trade_date_to, '%Y-%m-%d'), ''),
       COALESCE(source_key, ''), batch_size, universe_snapshot_id, status, current_stage,
       COALESCE(CAST(stage_progress_json AS CHAR), ''),
       COALESCE(CAST(summary_json AS CHAR), ''), COALESCE(error_message, ''), COALESCE(created_by, ''),
       created_at, updated_at, finished_at
FROM market_backfill_runs
WHERE id = ?`, strings.TrimSpace(runID)).
		Scan(
			&run.ID,
			&run.SchedulerRunID,
			&run.RunType,
			&assetScopeJSON,
			&run.TradeDateFrom,
			&run.TradeDateTo,
			&run.SourceKey,
			&run.BatchSize,
			&run.UniverseSnapshotID,
			&run.Status,
			&run.CurrentStage,
			&stageProgressJSON,
			&summaryJSON,
			&run.ErrorMessage,
			&run.CreatedBy,
			&createdAt,
			&updatedAt,
			&finishedAt,
		)
	if err != nil {
		return model.MarketBackfillRun{}, nil, err
	}
	run.AssetScope = unmarshalStringSlice(assetScopeJSON)
	run.StageProgress = unmarshalStageProgress(stageProgressJSON)
	run.Summary = unmarshalSummaryMap(summaryJSON)
	run.CreatedAt = createdAt.Format(time.RFC3339)
	run.UpdatedAt = updatedAt.Format(time.RFC3339)
	run.FinishedAt = parseNullableTimeRFC3339(finishedAt)

	rows, err := r.db.Query(`
SELECT id, snapshot_id, asset_type, instrument_key, COALESCE(external_symbol, ''), COALESCE(display_name, ''),
       COALESCE(exchange_code, ''), COALESCE(status, ''), COALESCE(DATE_FORMAT(list_date, '%Y-%m-%d'), ''),
       COALESCE(DATE_FORMAT(delist_date, '%Y-%m-%d'), ''), COALESCE(CAST(raw_metadata_json AS CHAR), ''), created_at
FROM market_universe_snapshot_items
WHERE snapshot_id = ?
ORDER BY asset_type ASC, instrument_key ASC`, run.UniverseSnapshotID)
	if err != nil {
		return model.MarketBackfillRun{}, nil, err
	}
	defer rows.Close()

	items := make([]model.MarketUniverseSnapshotItem, 0)
	for rows.Next() {
		var (
			item          model.MarketUniverseSnapshotItem
			itemCreatedAt time.Time
		)
		if err := rows.Scan(
			&item.ID,
			&item.SnapshotID,
			&item.AssetType,
			&item.InstrumentKey,
			&item.ExternalSymbol,
			&item.DisplayName,
			&item.ExchangeCode,
			&item.Status,
			&item.ListDate,
			&item.DelistDate,
			&item.MetadataJSON,
			&itemCreatedAt,
		); err != nil {
			return model.MarketBackfillRun{}, nil, err
		}
		item.CreatedAt = itemCreatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return run, items, rows.Err()
}

func (r *MySQLGrowthRepo) runMarketMasterStage(run model.MarketBackfillRun, byAsset map[string][]model.MarketUniverseSnapshotItem, assetScope []string, nowText string) ([]model.MarketBackfillRunDetail, error) {
	details := make([]model.MarketBackfillRunDetail, 0, len(assetScope))
	for _, assetType := range assetScope {
		items := byAsset[assetType]
		instrumentKeys := universeItemsToInstrumentKeys(items)
		if err := r.upsertMarketInstruments(assetType, instrumentKeys); err != nil {
			return nil, err
		}
		facts := buildMarketInstrumentSourceFactsFromUniverseItems(run.SourceKey, assetType, snapshotItemsToUniverseSourceItems(items), time.Now())
		if err := r.upsertMarketInstrumentSourceFacts(facts); err != nil {
			return nil, err
		}
		if err := r.upsertMarketSymbolAliasesFromInstrumentFacts(facts); err != nil {
			return nil, err
		}
		priority := []string{strings.ToUpper(strings.TrimSpace(run.SourceKey))}
		if priority[0] == "" {
			priority[0] = "MOCK"
		}
		if err := r.rebuildMarketInstrumentTruth(assetType, instrumentKeys, priority); err != nil {
			return nil, err
		}
		detail := newBackfillStageDetail(run, "MASTER", assetType, run.SourceKey, items, len(items), len(items), 0, "SUCCESS", nowText, "master synchronized")
		if err := r.insertMarketBackfillRunDetail(detail); err != nil {
			return nil, err
		}
		details = append(details, detail)
	}
	return details, nil
}

func (r *MySQLGrowthRepo) runMarketQuotesStage(run model.MarketBackfillRun, byAsset map[string][]model.MarketUniverseSnapshotItem, assetScope []string, windowDays int, nowText string) ([]model.MarketBackfillRunDetail, map[string]int, map[string]map[string]marketTouchedBarKey, error) {
	details := make([]model.MarketBackfillRunDetail, 0, len(assetScope))
	truthCounts := make(map[string]int, len(assetScope))
	touchedByAsset := make(map[string]map[string]marketTouchedBarKey, len(assetScope))
	for _, assetType := range assetScope {
		items := byAsset[assetType]
		instrumentKeys := universeItemsToInstrumentKeys(items)
		normalizedSource := strings.ToUpper(strings.TrimSpace(run.SourceKey))
		if normalizedSource == "" {
			normalizedSource = "MOCK"
		}

		fetchedCount := 0
		upsertedCount := 0
		status := "SUCCESS"
		message := "quotes synchronized"

		if normalizedSource == "MOCK" {
			bars := buildMockMarketDailyBars(assetType, normalizedSource, instrumentKeys, windowDays)
			count, err := r.upsertMarketDailyBars(bars)
			if err != nil {
				return nil, nil, nil, err
			}
			fetchedCount = len(bars)
			upsertedCount = count
			touchedByAsset[assetType] = buildTouchedBarKeysFromBars(bars)
		} else {
			routeConfigKey, defaultPriority := resolveMarketBackfillQuoteRoute(assetType)
			syncResult, err := r.syncMarketDailyBars(assetType, normalizedSource, instrumentKeys, windowDays, routeConfigKey, defaultPriority)
			if err != nil {
				return nil, nil, nil, err
			}
			fetchedCount = syncResult.BarCount
			upsertedCount = syncResult.BarCount
			truthCounts[assetType] = syncResult.TruthCount
			if len(syncResult.Results) > 0 && strings.TrimSpace(syncResult.Results[0].Message) != "" {
				message = syncResult.Results[0].Message
			}
			if len(syncResult.Results) > 0 && strings.TrimSpace(syncResult.Results[0].Status) != "" {
				status = syncResult.Results[0].Status
			}
		}

		detail := newBackfillStageDetail(run, "QUOTES", assetType, normalizedSource, items, fetchedCount, upsertedCount, 0, status, nowText, message)
		if err := r.insertMarketBackfillRunDetail(detail); err != nil {
			return nil, nil, nil, err
		}
		details = append(details, detail)
	}
	return details, truthCounts, touchedByAsset, nil
}

func (r *MySQLGrowthRepo) runMarketDailyBasicStage(run model.MarketBackfillRun, byAsset map[string][]model.MarketUniverseSnapshotItem, assetScope []string, windowDays int, nowText string) ([]model.MarketBackfillRunDetail, error) {
	return r.runMarketEnhancementStage(run, byAsset, assetScope, windowDays, nowText, marketDataKindDailyBasic)
}

func (r *MySQLGrowthRepo) runMarketMoneyflowStage(run model.MarketBackfillRun, byAsset map[string][]model.MarketUniverseSnapshotItem, assetScope []string, windowDays int, nowText string) ([]model.MarketBackfillRunDetail, error) {
	return r.runMarketEnhancementStage(run, byAsset, assetScope, windowDays, nowText, marketDataKindMoneyflow)
}

func (r *MySQLGrowthRepo) runMarketEnhancementStage(run model.MarketBackfillRun, byAsset map[string][]model.MarketUniverseSnapshotItem, assetScope []string, windowDays int, nowText string, dataKind string) ([]model.MarketBackfillRunDetail, error) {
	details := make([]model.MarketBackfillRunDetail, 0, len(assetScope))
	normalizedSource := strings.ToUpper(strings.TrimSpace(run.SourceKey))
	if normalizedSource == "" {
		normalizedSource = "MOCK"
	}
	for _, assetType := range assetScope {
		items := byAsset[assetType]
		instrumentKeys := universeItemsToInstrumentKeys(items)
		dailyBasicSupported, moneyflowSupported := marketAssetEnhancementSupport(assetType)
		supported := (dataKind == marketDataKindDailyBasic && dailyBasicSupported) || (dataKind == marketDataKindMoneyflow && moneyflowSupported)
		if !supported {
			detail := newBackfillStageDetail(run, dataKind, assetType, normalizedSource, items, 0, 0, 0, "SKIPPED", nowText, fmt.Sprintf("%s does not support %s in phase 1", assetType, strings.ToLower(dataKind)))
			if err := r.insertMarketBackfillRunDetail(detail); err != nil {
				return nil, err
			}
			details = append(details, detail)
			continue
		}

		count := 0
		if normalizedSource == "MOCK" {
			var err error
			switch dataKind {
			case marketDataKindDailyBasic:
				count, err = r.upsertStockDailyBasics(buildMockStockDailyBasics(normalizedSource, normalizeStockSymbolList(instrumentKeys), windowDays))
			case marketDataKindMoneyflow:
				count, err = r.upsertStockMoneyflows(buildMockStockMoneyflows(normalizedSource, normalizeStockSymbolList(instrumentKeys), windowDays))
			}
			if err != nil {
				return nil, err
			}
		} else {
			var (
				syncResult model.MarketSyncResult
				err        error
			)
			if dataKind == marketDataKindDailyBasic {
				syncResult, err = r.AdminSyncMarketDailyBasicDetailed(assetType, normalizedSource, instrumentKeys, windowDays)
			} else {
				syncResult, err = r.AdminSyncMarketMoneyflowDetailed(assetType, normalizedSource, instrumentKeys, windowDays)
			}
			if err != nil {
				return nil, err
			}
			count = syncResult.BarCount
		}

		detail := newBackfillStageDetail(run, dataKind, assetType, normalizedSource, items, count, count, 0, "SUCCESS", nowText, fmt.Sprintf("%s synchronized", strings.ToLower(dataKind)))
		if err := r.insertMarketBackfillRunDetail(detail); err != nil {
			return nil, err
		}
		details = append(details, detail)
	}
	return details, nil
}

func (r *MySQLGrowthRepo) runMarketTruthStage(run model.MarketBackfillRun, byAsset map[string][]model.MarketUniverseSnapshotItem, assetScope []string, quoteTruthCounts map[string]int, quoteTouchedByAsset map[string]map[string]marketTouchedBarKey, nowText string) ([]model.MarketBackfillRunDetail, error) {
	details := make([]model.MarketBackfillRunDetail, 0, len(assetScope))
	normalizedSource := strings.ToUpper(strings.TrimSpace(run.SourceKey))
	if normalizedSource == "" {
		normalizedSource = "MOCK"
	}
	for _, assetType := range assetScope {
		items := byAsset[assetType]
		truthCount := quoteTruthCounts[assetType]
		if touched := quoteTouchedByAsset[assetType]; len(touched) > 0 {
			selectedBars, err := r.rebuildMarketDailyBarTruth(assetType, touched, []string{normalizedSource})
			if err != nil {
				return nil, err
			}
			truthCount = len(selectedBars)
		}
		detail := newBackfillStageDetail(run, "TRUTH", assetType, normalizedSource, items, 0, 0, truthCount, "SUCCESS", nowText, "truth rebuilt")
		if err := r.insertMarketBackfillRunDetail(detail); err != nil {
			return nil, err
		}
		details = append(details, detail)
	}
	return details, nil
}

func (r *MySQLGrowthRepo) finalizeMarketCoverageSummaryStage(run model.MarketBackfillRun, nowText string) (model.MarketBackfillRunDetail, error) {
	detail := model.MarketBackfillRunDetail{
		ID:             newID("mbd"),
		RunID:          run.ID,
		SchedulerRunID: run.SchedulerRunID,
		Stage:          "COVERAGE_SUMMARY",
		BatchKey:       "COVERAGE-SUMMARY-001",
		SourceKey:      strings.ToUpper(strings.TrimSpace(run.SourceKey)),
		Status:         "SUCCESS",
		WarningText:    "coverage summary refreshed",
		StartedAt:      nowText,
		FinishedAt:     nowText,
		CreatedAt:      nowText,
		UpdatedAt:      nowText,
	}
	return detail, r.insertMarketBackfillRunDetail(detail)
}

func (r *MySQLGrowthRepo) failMarketBackfillRun(run model.MarketBackfillRun, progress []model.MarketBackfillStageProgress, stage string, cause error) (model.MarketBackfillRun, error) {
	nowText := time.Now().Format(time.RFC3339)
	progress = updateMarketBackfillStageProgress(progress, stage, "FAILED", 1, 0, 1, 0)
	run.Status = "FAILED"
	run.CurrentStage = stage
	run.StageProgress = progress
	run.ErrorMessage = cause.Error()
	run.UpdatedAt = nowText
	run.FinishedAt = nowText
	if err := r.updateMarketBackfillRunExecutionState(run); err != nil {
		return model.MarketBackfillRun{}, err
	}
	return model.MarketBackfillRun{}, cause
}

func (r *MySQLGrowthRepo) updateMarketBackfillRunExecutionState(run model.MarketBackfillRun) error {
	now := time.Now()
	finishedAt := interface{}(nil)
	if strings.TrimSpace(run.FinishedAt) != "" {
		if parsed, err := parseFlexibleDateTime(run.FinishedAt); err == nil {
			finishedAt = parsed
		} else {
			finishedAt = now
		}
	}
	_, err := r.db.Exec(`
UPDATE market_backfill_runs
SET status = ?, current_stage = ?, stage_progress_json = ?, summary_json = ?, error_message = ?, updated_at = ?, finished_at = ?
WHERE id = ?`,
		run.Status,
		run.CurrentStage,
		marshalJSONText(run.StageProgress),
		marshalJSONText(run.Summary),
		nullableString(strings.TrimSpace(run.ErrorMessage)),
		now,
		finishedAt,
		run.ID,
	)
	if err != nil {
		return err
	}
	resultSummary := fmt.Sprintf("market backfill %s at %s", strings.ToUpper(strings.TrimSpace(run.Status)), strings.ToUpper(strings.TrimSpace(run.CurrentStage)))
	return r.updateSchedulerJobRunExecutionState(run.SchedulerRunID, run.Status, resultSummary, run.ErrorMessage)
}

func (r *MySQLGrowthRepo) updateSchedulerJobRunExecutionState(runID string, status string, resultSummary string, errorMessage string) error {
	if strings.TrimSpace(runID) == "" {
		return nil
	}
	now := time.Now()
	finishedAt := interface{}(nil)
	upperStatus := strings.ToUpper(strings.TrimSpace(status))
	if upperStatus == "" {
		upperStatus = "RUNNING"
	}
	if upperStatus != "RUNNING" {
		finishedAt = now
	}
	_, err := r.db.Exec(`
UPDATE scheduler_job_runs
SET status = ?, result_summary = ?, error_message = ?, finished_at = ?
WHERE id = ?`,
		upperStatus,
		nullableString(truncateByRunes(normalizeUTF8Text(resultSummary), 512)),
		nullableString(truncateByRunes(normalizeUTF8Text(errorMessage), 512)),
		finishedAt,
		strings.TrimSpace(runID),
	)
	return err
}

func (r *MySQLGrowthRepo) insertMarketBackfillRunDetail(detail model.MarketBackfillRunDetail) error {
	now := time.Now()
	startedAt := parseBackfillDetailTime(detail.StartedAt, now)
	finishedAt := nullableTimeValue(detail.FinishedAt)
	_, err := r.db.Exec(`
INSERT INTO market_backfill_run_details (
	id, run_id, scheduler_run_id, stage, asset_type, batch_key, source_key, symbol_count, symbol_sample,
	trade_date_from, trade_date_to, status, fetched_count, upserted_count, truth_count, warning_text,
	error_text, started_at, finished_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		detail.ID,
		detail.RunID,
		nullableString(detail.SchedulerRunID),
		detail.Stage,
		nullableString(detail.AssetType),
		nullableString(detail.BatchKey),
		nullableString(detail.SourceKey),
		detail.SymbolCount,
		nullableString(marshalJSONText(detail.SymbolSample)),
		nullableString(strings.TrimSpace(detail.TradeDateFrom)),
		nullableString(strings.TrimSpace(detail.TradeDateTo)),
		detail.Status,
		detail.FetchedCount,
		detail.UpsertedCount,
		detail.TruthCount,
		nullableString(detail.WarningText),
		nullableString(detail.ErrorText),
		startedAt,
		finishedAt,
		now,
		now,
	)
	return err
}

func snapshotItemsToUniverseSourceItems(items []model.MarketUniverseSnapshotItem) []marketUniverseSourceItem {
	result := make([]marketUniverseSourceItem, 0, len(items))
	for _, item := range items {
		result = append(result, marketUniverseSourceItem{
			AssetType:      item.AssetType,
			InstrumentKey:  item.InstrumentKey,
			ExternalSymbol: item.ExternalSymbol,
			DisplayName:    item.DisplayName,
			ExchangeCode:   item.ExchangeCode,
			Status:         item.Status,
			ListDate:       item.ListDate,
			DelistDate:     item.DelistDate,
			MetadataJSON:   item.MetadataJSON,
		})
	}
	return result
}

func buildTouchedBarKeysFromBars(items []model.MarketDailyBar) map[string]marketTouchedBarKey {
	result := make(map[string]marketTouchedBarKey, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.InstrumentKey) == "" || strings.TrimSpace(item.TradeDate) == "" {
			continue
		}
		key := strings.ToUpper(strings.TrimSpace(item.InstrumentKey)) + "|" + strings.TrimSpace(item.TradeDate)
		result[key] = marketTouchedBarKey{
			InstrumentKey: strings.ToUpper(strings.TrimSpace(item.InstrumentKey)),
			TradeDate:     strings.TrimSpace(item.TradeDate),
		}
	}
	return result
}

func updateMarketBackfillProgressFromDetails(progress []model.MarketBackfillStageProgress, stage string, details []model.MarketBackfillRunDetail) []model.MarketBackfillStageProgress {
	completed, failed, skipped := summarizeMarketBackfillDetailStatuses(details)
	status := "SUCCESS"
	if failed > 0 && completed > 0 {
		status = "PARTIAL_SUCCESS"
	} else if failed > 0 {
		status = "FAILED"
	}
	return updateMarketBackfillStageProgress(progress, stage, status, len(details), completed, failed, skipped)
}

func summarizeMarketBackfillDetailStatuses(details []model.MarketBackfillRunDetail) (completed int, failed int, skipped int) {
	for _, detail := range details {
		switch detail.Status {
		case "FAILED":
			failed++
		case "SKIPPED":
			skipped++
		default:
			completed++
		}
	}
	return completed, failed, skipped
}

func resolveMarketBackfillQuoteRoute(assetType string) (string, []string) {
	switch normalizeMarketBackfillAssetType(assetType) {
	case "STOCK", "INDEX", "ETF", "LOF", "CBOND":
		return marketStockPriorityConfigKey, []string{"TUSHARE", "AKSHARE", "TICKERMD", "MYSELF", "MOCK"}
	default:
		return marketStockPriorityConfigKey, []string{"MOCK"}
	}
}

func parseBackfillDetailTime(raw string, fallback time.Time) time.Time {
	if parsed, err := parseFlexibleDateTime(strings.TrimSpace(raw)); err == nil {
		return parsed
	}
	return fallback
}

func nullableTimeValue(raw string) interface{} {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	if parsed, err := parseFlexibleDateTime(strings.TrimSpace(raw)); err == nil {
		return parsed
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminSyncMarketMasterDetailed(assetType string, sourceKey string, instrumentKeys []string) (model.MarketSyncResult, error) {
	normalizedAssetType := normalizeMarketStageSyncAssetType(assetType)
	if normalizedAssetType == "" {
		return model.MarketSyncResult{}, fmt.Errorf("asset_type is required")
	}
	normalizedKeys := normalizeMarketStageInstrumentKeys(normalizedAssetType, instrumentKeys)
	if len(normalizedKeys) == 0 {
		normalizedKeys = []string{inMemoryUniverseTemplateForAsset(normalizedAssetType).InstrumentKey}
	}
	normalizedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	if normalizedSourceKey == "" {
		normalizedSourceKey = "TUSHARE"
	}
	return model.MarketSyncResult{
		AssetClass:         normalizedAssetType,
		DataKind:           marketDataKindInstrumentMaster,
		RequestedSourceKey: normalizedSourceKey,
		ResolvedSourceKeys: []string{normalizedSourceKey},
		SnapshotCount:      len(normalizedKeys),
		Results: []model.MarketSourceSyncItemResult{
			{
				SourceKey:     normalizedSourceKey,
				Status:        "SUCCESS",
				SnapshotCount: len(normalizedKeys),
				Message:       "in-memory master sync",
			},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminSyncMarketQuotesDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error) {
	normalizedAssetType := normalizeMarketStageSyncAssetType(assetType)
	if normalizedAssetType == "" {
		return model.MarketSyncResult{}, fmt.Errorf("asset_type is required")
	}
	normalizedKeys := normalizeMarketStageInstrumentKeys(normalizedAssetType, instrumentKeys)
	if len(normalizedKeys) == 0 {
		normalizedKeys = []string{inMemoryUniverseTemplateForAsset(normalizedAssetType).InstrumentKey}
	}
	normalizedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	if normalizedSourceKey == "" {
		normalizedSourceKey = "MOCK"
	}
	if days <= 0 {
		days = 120
	}
	if days > 365 {
		days = 365
	}
	bars := buildMockMarketDailyBars(normalizedAssetType, normalizedSourceKey, normalizedKeys, days)
	return model.MarketSyncResult{
		AssetClass:         normalizedAssetType,
		DataKind:           marketDataKindDailyBars,
		RequestedSourceKey: normalizedSourceKey,
		ResolvedSourceKeys: []string{normalizedSourceKey},
		BarCount:           len(bars),
		TruthCount:         len(bars),
		Results: []model.MarketSourceSyncItemResult{
			{
				SourceKey:  normalizedSourceKey,
				Status:     "SUCCESS",
				BarCount:   len(bars),
				TruthCount: len(bars),
				Message:    "in-memory quotes sync",
			},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminSyncMarketDailyBasicDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error) {
	return buildInMemoryMarketEnhancementSyncResult(assetType, sourceKey, instrumentKeys, days, marketDataKindDailyBasic), nil
}

func (r *InMemoryGrowthRepo) AdminSyncMarketMoneyflowDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error) {
	return buildInMemoryMarketEnhancementSyncResult(assetType, sourceKey, instrumentKeys, days, marketDataKindMoneyflow), nil
}

func (r *InMemoryGrowthRepo) AdminRebuildMarketDailyTruthDetailed(assetType string, sourceKey string, instrumentKeys []string, tradeDateFrom string, tradeDateTo string) (model.MarketSyncResult, error) {
	normalizedAssetType := normalizeMarketStageSyncAssetType(assetType)
	if normalizedAssetType == "" {
		return model.MarketSyncResult{}, fmt.Errorf("asset_type is required")
	}
	normalizedKeys := normalizeMarketStageInstrumentKeys(normalizedAssetType, instrumentKeys)
	if len(normalizedKeys) == 0 {
		normalizedKeys = []string{inMemoryUniverseTemplateForAsset(normalizedAssetType).InstrumentKey}
	}
	normalizedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	if normalizedSourceKey == "" {
		normalizedSourceKey = "MOCK"
	}
	days := normalizeMarketBackfillWindowDays(tradeDateFrom, tradeDateTo)
	bars := buildMockMarketDailyBars(normalizedAssetType, normalizedSourceKey, normalizedKeys, days)
	truthCount := len(buildTouchedBarKeysFromBars(bars))
	return model.MarketSyncResult{
		AssetClass:         normalizedAssetType,
		DataKind:           marketDataKindTruthRebuild,
		RequestedSourceKey: normalizedSourceKey,
		ResolvedSourceKeys: []string{normalizedSourceKey},
		TruthCount:         truthCount,
		Results: []model.MarketSourceSyncItemResult{
			{
				SourceKey:  normalizedSourceKey,
				Status:     "SUCCESS",
				TruthCount: truthCount,
				Message:    "in-memory truth rebuild",
			},
		},
	}, nil
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
