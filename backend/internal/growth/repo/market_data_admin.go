package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func normalizeMarketQualityLogFilter(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeMarketDerivedTruthAssetClass(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case marketAssetClassStock:
		return marketAssetClassStock
	case marketAssetClassFutures:
		return marketAssetClassFutures
	default:
		return ""
	}
}

func buildMarketDerivedTruthWarnings(message string) []string {
	message = strings.TrimSpace(message)
	if message == "" {
		return nil
	}
	return []string{message}
}

func normalizeMarketQualitySummaryAssetClass(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", nil
	}
	normalized := normalizeMarketDerivedTruthAssetClass(trimmed)
	if normalized == "" {
		return "", fmt.Errorf("unsupported asset class: %s", value)
	}
	return normalized, nil
}

func resolveMarketDerivedTruthIssueCode(assetClass string) string {
	switch normalizeMarketDerivedTruthAssetClass(assetClass) {
	case marketAssetClassStock:
		return "DERIVED_STOCK_STATUS_REBUILT"
	case marketAssetClassFutures:
		return "DERIVED_FUTURES_MAPPING_REBUILT"
	default:
		return ""
	}
}

func parseMarketDerivedTruthSummaryPayload(payload string) model.MarketDerivedTruthSummary {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return model.MarketDerivedTruthSummary{}
	}
	var item model.MarketDerivedTruthSummary
	if err := json.Unmarshal([]byte(payload), &item); err != nil {
		return model.MarketDerivedTruthSummary{}
	}
	return item
}

func buildMarketQualitySummaryWhereClause(assetClass string) (string, []any, error) {
	normalizedAssetClass, err := normalizeMarketQualitySummaryAssetClass(assetClass)
	if err != nil {
		return "", nil, err
	}
	filters := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if normalizedAssetClass != "" {
		filters = append(filters, "asset_class = ?")
		args = append(args, normalizedAssetClass)
	}
	filters = append(filters, "created_at >= ?")
	return " WHERE " + strings.Join(filters, " AND "), args, nil
}

func (r *MySQLGrowthRepo) AdminListMarketDataQualityLogs(assetClass string, dataKind string, severity string, issueCode string, hours int, page int, pageSize int) ([]model.MarketDataQualityLog, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 200 {
		pageSize = 200
	}
	offset := (page - 1) * pageSize

	filters := make([]string, 0, 5)
	args := make([]any, 0, 7)

	if normalized := normalizeMarketDerivedTruthAssetClass(assetClass); normalized != "" {
		filters = append(filters, "asset_class = ?")
		args = append(args, normalized)
	}
	if normalized := normalizeMarketQualityLogFilter(dataKind); normalized != "" {
		filters = append(filters, "data_kind = ?")
		args = append(args, normalized)
	}
	if normalized := normalizeMarketQualityLogFilter(severity); normalized != "" {
		filters = append(filters, "severity = ?")
		args = append(args, normalized)
	}
	if normalized := normalizeMarketQualityLogFilter(issueCode); normalized != "" {
		filters = append(filters, "issue_code = ?")
		args = append(args, normalized)
	}
	if hours > 0 {
		filters = append(filters, "created_at >= ?")
		args = append(args, time.Now().Add(-time.Duration(hours)*time.Hour))
	}

	whereClause := ""
	if len(filters) > 0 {
		whereClause = " WHERE " + strings.Join(filters, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM market_data_quality_logs" + whereClause
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		if isMarketStatusSchemaCompatError(err) {
			return []model.MarketDataQualityLog{}, 0, nil
		}
		return nil, 0, err
	}

	queryArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := r.db.Query(`
SELECT id, asset_class, data_kind, instrument_key, trade_date, source_key, severity, issue_code, issue_message, COALESCE(CAST(payload_json AS CHAR), ''), created_at
FROM market_data_quality_logs`+whereClause+`
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		if isMarketStatusSchemaCompatError(err) {
			return []model.MarketDataQualityLog{}, 0, nil
		}
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.MarketDataQualityLog, 0)
	for rows.Next() {
		var (
			item           model.MarketDataQualityLog
			assetValue     sql.NullString
			instrumentKey  sql.NullString
			tradeDateValue sql.NullTime
			sourceKeyValue sql.NullString
			issueMessage   sql.NullString
			payloadValue   sql.NullString
			createdAt      time.Time
		)
		if err := rows.Scan(
			&item.ID,
			&assetValue,
			&item.DataKind,
			&instrumentKey,
			&tradeDateValue,
			&sourceKeyValue,
			&item.Severity,
			&item.IssueCode,
			&issueMessage,
			&payloadValue,
			&createdAt,
		); err != nil {
			return nil, 0, err
		}
		if assetValue.Valid {
			item.AssetClass = assetValue.String
		}
		if instrumentKey.Valid {
			item.InstrumentKey = instrumentKey.String
		}
		if tradeDateValue.Valid {
			item.TradeDate = tradeDateValue.Time.Format("2006-01-02")
		}
		if sourceKeyValue.Valid {
			item.SourceKey = sourceKeyValue.String
		}
		if issueMessage.Valid {
			item.IssueMessage = issueMessage.String
		}
		if payloadValue.Valid {
			item.Payload = payloadValue.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminGetMarketDataQualitySummary(assetClass string, hours int) (model.MarketDataQualitySummary, error) {
	if hours <= 0 {
		hours = 24
	}
	summary := model.MarketDataQualitySummary{
		LookbackHours: hours,
	}
	normalizedAssetClass, err := normalizeMarketQualitySummaryAssetClass(assetClass)
	if err != nil {
		return summary, err
	}
	summary.AssetClass = normalizedAssetClass

	whereClause, args, err := buildMarketQualitySummaryWhereClause(assetClass)
	if err != nil {
		return summary, err
	}
	lookbackStart := time.Now().Add(-time.Duration(hours) * time.Hour)
	args = append(args, lookbackStart)

	err = r.db.QueryRow(`
SELECT COUNT(*),
       COALESCE(SUM(CASE WHEN severity = 'ERROR' THEN 1 ELSE 0 END), 0),
       COALESCE(SUM(CASE WHEN severity = 'WARN' THEN 1 ELSE 0 END), 0),
       COALESCE(SUM(CASE WHEN severity = 'INFO' THEN 1 ELSE 0 END), 0),
       COUNT(DISTINCT CASE WHEN COALESCE(source_key, '') = '' THEN NULL ELSE source_key END)
FROM market_data_quality_logs`+whereClause, args...).Scan(
		&summary.TotalCount,
		&summary.ErrorCount,
		&summary.WarnCount,
		&summary.InfoCount,
		&summary.DistinctSourceCount,
	)
	if err != nil {
		if isMarketStatusSchemaCompatError(err) {
			return summary, nil
		}
		return summary, err
	}

	latestArgs := append([]any{}, args...)
	var (
		latestSourceValue   sql.NullString
		latestSeverityValue sql.NullString
		latestIssueValue    sql.NullString
		latestMessageValue  sql.NullString
		latestTradeDate     sql.NullTime
		latestCreatedAt     time.Time
	)
	err = r.db.QueryRow(`
SELECT source_key, severity, issue_code, issue_message, trade_date, created_at
FROM market_data_quality_logs`+whereClause+`
ORDER BY created_at DESC, id DESC
LIMIT 1`, latestArgs...).Scan(
		&latestSourceValue,
		&latestSeverityValue,
		&latestIssueValue,
		&latestMessageValue,
		&latestTradeDate,
		&latestCreatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		if isMarketStatusSchemaCompatError(err) {
			return summary, nil
		}
		return summary, err
	}
	if err == nil {
		if latestSourceValue.Valid {
			summary.LatestSourceKey = latestSourceValue.String
		}
		if latestSeverityValue.Valid {
			summary.LatestSeverity = latestSeverityValue.String
		}
		if latestIssueValue.Valid {
			summary.LatestIssueCode = latestIssueValue.String
		}
		if latestMessageValue.Valid {
			summary.LatestIssueMessage = latestMessageValue.String
		}
		if latestTradeDate.Valid {
			summary.LatestTradeDate = latestTradeDate.Time.Format("2006-01-02")
		}
		summary.LatestCreatedAt = latestCreatedAt.Format(time.RFC3339)
	}

	errorWhereClause := whereClause + " AND severity = 'ERROR'"
	errorArgs := append([]any{}, args...)
	var (
		errorSourceValue  sql.NullString
		errorIssueValue   sql.NullString
		errorMessageValue sql.NullString
		errorCreatedAt    time.Time
	)
	err = r.db.QueryRow(`
SELECT source_key, issue_code, issue_message, created_at
FROM market_data_quality_logs`+errorWhereClause+`
ORDER BY created_at DESC, id DESC
LIMIT 1`, errorArgs...).Scan(
		&errorSourceValue,
		&errorIssueValue,
		&errorMessageValue,
		&errorCreatedAt,
	)
	if err != nil && err != sql.ErrNoRows {
		if isMarketStatusSchemaCompatError(err) {
			return summary, nil
		}
		return summary, err
	}
	if err == nil {
		if errorSourceValue.Valid {
			summary.LatestErrorSourceKey = errorSourceValue.String
		}
		if errorIssueValue.Valid {
			summary.LatestErrorIssueCode = errorIssueValue.String
		}
		if errorMessageValue.Valid {
			summary.LatestErrorMessage = errorMessageValue.String
		}
		summary.LatestErrorCreatedAt = errorCreatedAt.Format(time.RFC3339)
	}

	return summary, nil
}

func (r *MySQLGrowthRepo) AdminGetMarketDerivedTruthSummary(assetClass string) (*model.MarketDerivedTruthSummary, error) {
	normalizedAssetClass := normalizeMarketDerivedTruthAssetClass(assetClass)
	issueCode := resolveMarketDerivedTruthIssueCode(assetClass)
	if normalizedAssetClass == "" || issueCode == "" {
		return nil, fmt.Errorf("unsupported asset class: %s", assetClass)
	}

	var (
		assetValue    sql.NullString
		sourceValue   sql.NullString
		issueValue    sql.NullString
		messageValue  sql.NullString
		payloadValue  sql.NullString
		createdAtTime time.Time
	)
	err := r.db.QueryRow(`
SELECT asset_class, source_key, issue_code, issue_message, COALESCE(CAST(payload_json AS CHAR), ''), created_at
FROM market_data_quality_logs
WHERE asset_class = ? AND source_key = ? AND issue_code = ?
ORDER BY created_at DESC, id DESC
LIMIT 1`,
		normalizedAssetClass,
		"LOCAL_TRUTH",
		issueCode,
	).Scan(&assetValue, &sourceValue, &issueValue, &messageValue, &payloadValue, &createdAtTime)
	if err != nil {
		if err == sql.ErrNoRows || isMarketStatusSchemaCompatError(err) {
			return nil, nil
		}
		return nil, err
	}

	item := parseMarketDerivedTruthSummaryPayload(payloadValue.String)
	item.AssetClass = normalizedAssetClass
	if sourceValue.Valid {
		item.SourceKey = sourceValue.String
	}
	if issueValue.Valid {
		item.IssueCode = issueValue.String
	}
	if messageValue.Valid {
		item.IssueMessage = messageValue.String
	}
	item.CreatedAt = createdAtTime.Format(time.RFC3339)
	return &item, nil
}

func (r *MySQLGrowthRepo) AdminRebuildMarketDerivedTruth(assetClass string, tradeDate string, days int) (model.MarketDerivedTruthRebuildResult, error) {
	normalizedAssetClass := normalizeMarketDerivedTruthAssetClass(assetClass)
	if normalizedAssetClass == "" {
		return model.MarketDerivedTruthRebuildResult{}, fmt.Errorf("unsupported asset class: %s", assetClass)
	}
	if days <= 0 {
		days = 1
	}
	if days > 365 {
		days = 365
	}

	result := model.MarketDerivedTruthRebuildResult{
		AssetClass: normalizedAssetClass,
		Days:       days,
	}

	endTradeDate, err := r.resolveMarketDerivedTruthEndTradeDate(normalizedAssetClass, tradeDate)
	if err != nil {
		if isMarketStatusSchemaCompatError(err) {
			result.Warnings = buildMarketDerivedTruthWarnings("本地 truth 行情表尚未就绪，已跳过派生重建。")
			return result, nil
		}
		return result, err
	}
	if endTradeDate.IsZero() {
		result.Warnings = buildMarketDerivedTruthWarnings("未找到可用于重建的本地 truth 行情数据。")
		return result, nil
	}

	startTradeDate := endTradeDate.AddDate(0, 0, -(days - 1))
	result.TradeDate = endTradeDate.Format("2006-01-02")
	result.StartDate = startTradeDate.Format("2006-01-02")
	result.EndDate = result.TradeDate

	truthBars, err := r.loadMarketDailyBarTruthForRebuild(normalizedAssetClass, startTradeDate, endTradeDate)
	if err != nil {
		if isMarketStatusSchemaCompatError(err) {
			result.Warnings = buildMarketDerivedTruthWarnings("本地 truth 行情表尚未就绪，已跳过派生重建。")
			return result, nil
		}
		return result, err
	}
	result.TruthBarCount = len(truthBars)
	if len(truthBars) == 0 {
		result.Warnings = buildMarketDerivedTruthWarnings("重建窗口内没有可用 truth 行情数据。")
		return result, nil
	}

	switch normalizedAssetClass {
	case marketAssetClassStock:
		count, rebuildErr := r.rebuildStockStatusTruth(truthBars)
		if rebuildErr != nil {
			if isMarketStatusSchemaCompatError(rebuildErr) {
				result.Warnings = buildMarketDerivedTruthWarnings("股票状态 truth 表尚未就绪，已跳过写入。")
				return result, nil
			}
			return result, rebuildErr
		}
		result.StockStatusCount = count
		r.insertMarketDataQualityLog(
			normalizedAssetClass,
			marketDataKindDailyBars,
			"",
			result.TradeDate,
			"LOCAL_TRUTH",
			"INFO",
			"DERIVED_STOCK_STATUS_REBUILT",
			fmt.Sprintf("rebuilt stock status truth for %d rows", count),
			marshalMarketDerivedTruthRebuildPayload(result),
		)
	case marketAssetClassFutures:
		count, rebuildErr := r.rebuildFuturesContractMappings(truthBars)
		if rebuildErr != nil {
			if isMarketStatusSchemaCompatError(rebuildErr) {
				result.Warnings = buildMarketDerivedTruthWarnings("期货主力映射表尚未就绪，已跳过写入。")
				return result, nil
			}
			return result, rebuildErr
		}
		result.FuturesMappingCount = count
		r.insertMarketDataQualityLog(
			normalizedAssetClass,
			marketDataKindDailyBars,
			"",
			result.TradeDate,
			"LOCAL_TRUTH",
			"INFO",
			"DERIVED_FUTURES_MAPPING_REBUILT",
			fmt.Sprintf("rebuilt futures dominant mappings for %d rows", count),
			marshalMarketDerivedTruthRebuildPayload(result),
		)
	}

	return result, nil
}

func (r *MySQLGrowthRepo) resolveMarketDerivedTruthEndTradeDate(assetClass string, tradeDate string) (time.Time, error) {
	requestedTradeDate := strings.TrimSpace(tradeDate)
	var (
		query string
		args  []any
	)
	if requestedTradeDate == "" {
		query = `
SELECT MAX(trade_date)
FROM market_daily_bar_truth
WHERE asset_class = ?`
		args = []any{assetClass}
	} else {
		query = `
SELECT MAX(trade_date)
FROM market_daily_bar_truth
WHERE asset_class = ? AND trade_date <= ?`
		args = []any{assetClass, requestedTradeDate}
	}
	var selectedTradeDate sql.NullTime
	if err := r.db.QueryRow(query, args...).Scan(&selectedTradeDate); err != nil {
		return time.Time{}, err
	}
	if !selectedTradeDate.Valid {
		return time.Time{}, nil
	}
	return selectedTradeDate.Time, nil
}

func (r *MySQLGrowthRepo) loadMarketDailyBarTruthForRebuild(assetClass string, startDate time.Time, endDate time.Time) ([]model.MarketDailyBar, error) {
	rows, err := r.db.Query(`
SELECT asset_class, instrument_key, trade_date, selected_source_key, open_price, high_price, low_price, close_price, prev_close_price, settle_price, prev_settle_price, volume, turnover, open_interest
FROM market_daily_bar_truth
WHERE asset_class = ? AND trade_date BETWEEN ? AND ?
ORDER BY trade_date ASC, instrument_key ASC`,
		assetClass,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.MarketDailyBar, 0)
	for rows.Next() {
		var (
			item            model.MarketDailyBar
			tradeDateValue  time.Time
			selectedSource  sql.NullString
			prevCloseVal    sql.NullFloat64
			settleVal       sql.NullFloat64
			prevSettleVal   sql.NullFloat64
			openInterestVal sql.NullFloat64
		)
		if err := rows.Scan(
			&item.AssetClass,
			&item.InstrumentKey,
			&tradeDateValue,
			&selectedSource,
			&item.OpenPrice,
			&item.HighPrice,
			&item.LowPrice,
			&item.ClosePrice,
			&prevCloseVal,
			&settleVal,
			&prevSettleVal,
			&item.Volume,
			&item.Turnover,
			&openInterestVal,
		); err != nil {
			return nil, err
		}
		item.TradeDate = tradeDateValue.Format("2006-01-02")
		if selectedSource.Valid {
			item.SourceKey = selectedSource.String
		}
		item.PrevClosePrice = sqlNullFloat(prevCloseVal)
		item.SettlePrice = sqlNullFloat(settleVal)
		item.PrevSettlePrice = sqlNullFloat(prevSettleVal)
		if openInterestVal.Valid {
			item.OpenInterest = openInterestVal.Float64
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *InMemoryGrowthRepo) AdminListMarketDataQualityLogs(assetClass string, dataKind string, severity string, issueCode string, hours int, page int, pageSize int) ([]model.MarketDataQualityLog, int, error) {
	items := []model.MarketDataQualityLog{
		{
			ID:            "mdq_demo_001",
			AssetClass:    marketAssetClassStock,
			DataKind:      marketDataKindDailyBars,
			InstrumentKey: "600519.SH",
			TradeDate:     "2026-03-22",
			SourceKey:     "TUSHARE",
			Severity:      "WARN",
			IssueCode:     "SOURCE_FETCH_FAILED",
			IssueMessage:  "上游返回超时，已由本地 truth 保持旧值。",
			Payload:       `{"attempt":1}`,
			CreatedAt:     time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
		},
		{
			ID:            "mdq_demo_002",
			AssetClass:    marketAssetClassFutures,
			DataKind:      marketDataKindDailyBars,
			InstrumentKey: "IF2606.CFX",
			TradeDate:     "2026-03-22",
			SourceKey:     "MYSELF",
			Severity:      "INFO",
			IssueCode:     "DERIVED_MAPPING_REBUILT",
			IssueMessage:  "主力映射已按本地 truth 重建。",
			CreatedAt:     time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
		},
	}
	filtered := make([]model.MarketDataQualityLog, 0, len(items))
	lookbackStart := time.Time{}
	if hours > 0 {
		lookbackStart = time.Now().Add(-time.Duration(hours) * time.Hour)
	}
	for _, item := range items {
		if normalized := normalizeMarketDerivedTruthAssetClass(assetClass); normalized != "" && item.AssetClass != normalized {
			continue
		}
		if normalized := normalizeMarketQualityLogFilter(dataKind); normalized != "" && item.DataKind != normalized {
			continue
		}
		if normalized := normalizeMarketQualityLogFilter(severity); normalized != "" && item.Severity != normalized {
			continue
		}
		if normalized := normalizeMarketQualityLogFilter(issueCode); normalized != "" && item.IssueCode != normalized {
			continue
		}
		if !lookbackStart.IsZero() {
			createdAt, err := time.Parse(time.RFC3339, item.CreatedAt)
			if err != nil || createdAt.Before(lookbackStart) {
				continue
			}
		}
		filtered = append(filtered, item)
	}
	return filtered, len(filtered), nil
}

func (r *InMemoryGrowthRepo) AdminGetMarketDataQualitySummary(assetClass string, hours int) (model.MarketDataQualitySummary, error) {
	if hours <= 0 {
		hours = 24
	}
	normalizedAssetClass, err := normalizeMarketQualitySummaryAssetClass(assetClass)
	if err != nil {
		return model.MarketDataQualitySummary{}, err
	}
	now := time.Now()
	item := model.MarketDataQualitySummary{
		AssetClass:           normalizedAssetClass,
		LookbackHours:        hours,
		TotalCount:           7,
		ErrorCount:           1,
		WarnCount:            4,
		InfoCount:            2,
		DistinctSourceCount:  3,
		LatestSourceKey:      "MYSELF",
		LatestSeverity:       "WARN",
		LatestIssueCode:      "BAR_UPSERT_RETRIED",
		LatestIssueMessage:   "upsert retried with fallback",
		LatestTradeDate:      "2026-03-22",
		LatestCreatedAt:      now.Add(-15 * time.Minute).Format(time.RFC3339),
		LatestErrorSourceKey: "TUSHARE",
		LatestErrorIssueCode: "SOURCE_FETCH_FAILED",
		LatestErrorMessage:   "upstream timeout",
		LatestErrorCreatedAt: now.Add(-35 * time.Minute).Format(time.RFC3339),
	}
	if normalizedAssetClass == "" {
		item.TotalCount = 13
		item.ErrorCount = 2
		item.WarnCount = 8
		item.InfoCount = 3
		item.DistinctSourceCount = 4
	}
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminGetMarketDerivedTruthSummary(assetClass string) (*model.MarketDerivedTruthSummary, error) {
	normalizedAssetClass := normalizeMarketDerivedTruthAssetClass(assetClass)
	if normalizedAssetClass == "" {
		return nil, fmt.Errorf("unsupported asset class: %s", assetClass)
	}
	item := &model.MarketDerivedTruthSummary{
		AssetClass:    normalizedAssetClass,
		SourceKey:     "LOCAL_TRUTH",
		TradeDate:     "2026-03-22",
		StartDate:     "2026-03-20",
		EndDate:       "2026-03-22",
		Days:          3,
		TruthBarCount: 12,
		CreatedAt:     time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
	}
	if normalizedAssetClass == marketAssetClassStock {
		item.IssueCode = "DERIVED_STOCK_STATUS_REBUILT"
		item.IssueMessage = "rebuilt stock status truth for 12 rows"
		item.StockStatusCount = 12
		return item, nil
	}
	item.IssueCode = "DERIVED_FUTURES_MAPPING_REBUILT"
	item.IssueMessage = "rebuilt futures dominant mappings for 4 rows"
	item.FuturesMappingCount = 4
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminRebuildMarketDerivedTruth(assetClass string, tradeDate string, days int) (model.MarketDerivedTruthRebuildResult, error) {
	normalizedAssetClass := normalizeMarketDerivedTruthAssetClass(assetClass)
	if normalizedAssetClass == "" {
		return model.MarketDerivedTruthRebuildResult{}, fmt.Errorf("unsupported asset class: %s", assetClass)
	}
	if days <= 0 {
		days = 1
	}
	result := model.MarketDerivedTruthRebuildResult{
		AssetClass: normalizedAssetClass,
		TradeDate:  "2026-03-22",
		StartDate:  "2026-03-21",
		EndDate:    "2026-03-22",
		Days:       days,
	}
	if strings.TrimSpace(tradeDate) != "" {
		result.TradeDate = strings.TrimSpace(tradeDate)
		result.StartDate = result.TradeDate
		result.EndDate = result.TradeDate
	}
	if normalizedAssetClass == marketAssetClassStock {
		result.TruthBarCount = 4
		result.StockStatusCount = 4
		return result, nil
	}
	result.TruthBarCount = 3
	result.FuturesMappingCount = 1
	return result, nil
}

func marshalMarketDerivedTruthRebuildPayload(result model.MarketDerivedTruthRebuildResult) string {
	payload, err := json.Marshal(map[string]any{
		"asset_class":           result.AssetClass,
		"trade_date":            result.TradeDate,
		"start_date":            result.StartDate,
		"end_date":              result.EndDate,
		"days":                  result.Days,
		"truth_bar_count":       result.TruthBarCount,
		"stock_status_count":    result.StockStatusCount,
		"futures_mapping_count": result.FuturesMappingCount,
		"warnings":              result.Warnings,
	})
	if err != nil {
		return ""
	}
	return string(payload)
}
