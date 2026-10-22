package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) ListStockRecommendations(userID string, tradeDate string, page int, pageSize int) ([]model.StockRecommendation, int, error) {
	offset := (page - 1) * pageSize
	var latestTradeDate sql.NullString
	if tradeDate == "" {
		if err := r.db.QueryRow(`
SELECT DATE_FORMAT(MAX(valid_from), '%Y-%m-%d')
FROM stock_recommendations
WHERE status IN ('PUBLISHED', 'ACTIVE', 'TRACKING')`).Scan(&latestTradeDate); err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.db.QueryRow(`
SELECT DATE_FORMAT(MAX(valid_from), '%Y-%m-%d')
FROM stock_recommendations
WHERE status IN ('PUBLISHED', 'ACTIVE', 'TRACKING')
  AND DATE(valid_from) <= ?`, tradeDate).Scan(&latestTradeDate); err != nil {
			return nil, 0, err
		}
	}
	if latestTradeDate.Valid {
		tradeDate = strings.TrimSpace(latestTradeDate.String)
	}
	args := []interface{}{}
	filter := " WHERE r.status IN ('PUBLISHED', 'ACTIVE', 'TRACKING')"
	if tradeDate != "" {
		filter += " AND DATE(r.valid_from) = ?"
		args = append(args, tradeDate)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM stock_recommendations r"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT r.id, r.symbol, r.name, r.score, r.risk_level, COALESCE(r.position_range, ''), r.valid_from, r.valid_to, r.status, COALESCE(r.reason_summary, ''),
       COALESCE(r.source_type, ''), COALESCE(r.strategy_version, ''), COALESCE(r.performance_label, ''),
       COALESCE(d.take_profit, ''), COALESCE(d.stop_loss, '')
FROM stock_recommendations r
LEFT JOIN stock_reco_details d ON d.reco_id = r.id` + filter + `
ORDER BY r.score DESC, r.created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.StockRecommendation, 0)
	for rows.Next() {
		var item model.StockRecommendation
		var validFrom, validTo time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Symbol,
			&item.Name,
			&item.Score,
			&item.RiskLevel,
			&item.PositionRange,
			&validFrom,
			&validTo,
			&item.Status,
			&item.ReasonSummary,
			&item.SourceType,
			&item.StrategyVersion,
			&item.PerformanceLabel,
			&item.TakeProfit,
			&item.StopLoss,
		); err != nil {
			return nil, 0, err
		}
		item.ValidFrom = validFrom.Format(time.RFC3339)
		item.ValidTo = validTo.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) ListStockRecommendationHistory(userID string, outcome string, tradeDateFrom string, tradeDateTo string, page int, pageSize int) ([]model.StockRecommendationHistoryItem, model.StockRecommendationHistorySummary, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE r.status IN ('PUBLISHED', 'ACTIVE', 'TRACKING', 'HIT_TAKE_PROFIT', 'HIT_STOP_LOSS', 'INVALIDATED', 'REVIEWED')"
	if tradeDateFrom != "" {
		filter += " AND DATE(r.valid_from) >= ?"
		args = append(args, tradeDateFrom)
	}
	if tradeDateTo != "" {
		filter += " AND DATE(r.valid_from) <= ?"
		args = append(args, tradeDateTo)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM stock_recommendations r"+filter, args...).Scan(&total); err != nil {
		return nil, model.StockRecommendationHistorySummary{}, 0, err
	}

	query := `
SELECT r.id, r.symbol, r.name, r.score, r.risk_level, COALESCE(r.position_range, ''), r.valid_from, r.valid_to, r.status,
       COALESCE(r.source_type, ''), COALESCE(r.strategy_version, ''), COALESCE(r.performance_label, ''),
       COALESCE(d.take_profit, ''), COALESCE(d.stop_loss, '')
FROM stock_recommendations r
LEFT JOIN stock_reco_details d ON d.reco_id = r.id` + filter + `
ORDER BY r.valid_from DESC, r.created_at DESC
LIMIT ? OFFSET ?`
	queryArgs := append(append([]interface{}{}, args...), pageSize, offset)
	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, model.StockRecommendationHistorySummary{}, 0, err
	}
	defer rows.Close()

	type historyBaseRow struct {
		item      model.StockRecommendationHistoryItem
		validFrom time.Time
		validTo   time.Time
	}

	baseRows := make([]historyBaseRow, 0)
	symbolSet := make(map[string]struct{})
	var earliestStart time.Time
	var latestEnd time.Time
	for rows.Next() {
		var row historyBaseRow
		if err := rows.Scan(
			&row.item.ID,
			&row.item.Symbol,
			&row.item.Name,
			&row.item.Score,
			&row.item.RiskLevel,
			&row.item.PositionRange,
			&row.validFrom,
			&row.validTo,
			&row.item.Status,
			&row.item.SourceType,
			&row.item.StrategyVersion,
			&row.item.PerformanceLabel,
			&row.item.TakeProfit,
			&row.item.StopLoss,
		); err != nil {
			return nil, model.StockRecommendationHistorySummary{}, 0, err
		}
		row.item.ValidFrom = row.validFrom.Format(time.RFC3339)
		row.item.ValidTo = row.validTo.Format(time.RFC3339)
		row.item.IsClosed = stockRecommendationHistoryIsClosed(row.item.Status)
		baseRows = append(baseRows, row)
		symbolSet[row.item.Symbol] = struct{}{}
		if earliestStart.IsZero() || row.validFrom.Before(earliestStart) {
			earliestStart = row.validFrom
		}
		if latestEnd.IsZero() || row.validTo.After(latestEnd) {
			latestEnd = row.validTo
		}
	}
	if err := rows.Err(); err != nil {
		return nil, model.StockRecommendationHistorySummary{}, 0, err
	}
	if len(baseRows) == 0 {
		return []model.StockRecommendationHistoryItem{}, model.StockRecommendationHistorySummary{}, total, nil
	}

	if tradeDateFrom != "" {
		if parsed, err := time.Parse("2006-01-02", tradeDateFrom); err == nil && (earliestStart.IsZero() || parsed.Before(earliestStart)) {
			earliestStart = parsed
		}
	}
	if tradeDateTo != "" {
		if parsed, err := time.Parse("2006-01-02", tradeDateTo); err == nil && (latestEnd.IsZero() || parsed.After(latestEnd)) {
			latestEnd = parsed
		}
	}

	symbols := make([]string, 0, len(symbolSet))
	for symbol := range symbolSet {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	quotesBySymbol, err := r.listStockRecommendationHistoryQuotes(symbols, earliestStart, latestEnd)
	if err != nil {
		return nil, model.StockRecommendationHistorySummary{}, 0, err
	}

	allItems := make([]model.StockRecommendationHistoryItem, 0, len(baseRows))
	for _, row := range baseRows {
		item := row.item
		quotes := quotesBySymbol[item.Symbol]
		entryQuote := findStockRecommendationHistoryEntryQuote(quotes, row.validFrom)
		latestQuote := findStockRecommendationHistoryLatestQuote(quotes, row.validTo, item.IsClosed)
		if entryQuote != nil {
			item.EntryPrice = entryQuote.ClosePrice
		}
		if latestQuote != nil {
			item.LatestPrice = latestQuote.ClosePrice
		}
		item.ReturnPct = calculateStockRecommendationHistoryReturnPct(item.EntryPrice, item.LatestPrice)
		item.MaxDrawdownPct = calculateStockRecommendationHistoryMaxDrawdownPct(quotes, row.validFrom, latestQuote)
		item.Outcome = mapStockRecommendationHistoryOutcome(item.Status, item.PerformanceLabel)
		allItems = append(allItems, item)
	}

	filteredItems := make([]model.StockRecommendationHistoryItem, 0, len(allItems))
	for _, item := range allItems {
		if outcome == "" || strings.EqualFold(item.Outcome, outcome) {
			filteredItems = append(filteredItems, item)
		}
	}

	summary := buildStockRecommendationHistorySummary(filteredItems)
	return filteredItems, summary, total, nil
}

type stockRecommendationHistoryQuote struct {
	TradeDate  time.Time
	ClosePrice float64
}

func (r *MySQLGrowthRepo) listStockRecommendationHistoryQuotes(symbols []string, start time.Time, end time.Time) (map[string][]stockRecommendationHistoryQuote, error) {
	result := make(map[string][]stockRecommendationHistoryQuote, len(symbols))
	if len(symbols) == 0 || start.IsZero() || end.IsZero() {
		return result, nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(symbols)), ",")
	query := fmt.Sprintf(`
SELECT symbol, trade_date, close_price
FROM stock_market_quotes
WHERE symbol IN (%s) AND trade_date >= ? AND trade_date <= ?
ORDER BY symbol ASC, trade_date ASC`, placeholders)
	args := make([]interface{}, 0, len(symbols)+2)
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	args = append(args, start.Format("2006-01-02"), end.Format("2006-01-02"))
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var symbol string
		var tradeDate time.Time
		var closePrice float64
		if err := rows.Scan(&symbol, &tradeDate, &closePrice); err != nil {
			return nil, err
		}
		result[symbol] = append(result[symbol], stockRecommendationHistoryQuote{
			TradeDate:  tradeDate,
			ClosePrice: closePrice,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func findStockRecommendationHistoryEntryQuote(quotes []stockRecommendationHistoryQuote, validFrom time.Time) *stockRecommendationHistoryQuote {
	if len(quotes) == 0 {
		return nil
	}
	startDate := dateOnlyTime(validFrom)
	for index := range quotes {
		if !quotes[index].TradeDate.Before(startDate) {
			return &quotes[index]
		}
	}
	return &quotes[0]
}

func findStockRecommendationHistoryLatestQuote(quotes []stockRecommendationHistoryQuote, validTo time.Time, isClosed bool) *stockRecommendationHistoryQuote {
	if len(quotes) == 0 {
		return nil
	}
	if !isClosed {
		return &quotes[len(quotes)-1]
	}
	endDate := dateOnlyTime(validTo)
	for index := len(quotes) - 1; index >= 0; index-- {
		if !quotes[index].TradeDate.After(endDate) {
			return &quotes[index]
		}
	}
	return &quotes[len(quotes)-1]
}

func calculateStockRecommendationHistoryReturnPct(entryPrice float64, latestPrice float64) float64 {
	if entryPrice <= 0 || latestPrice <= 0 {
		return 0
	}
	return roundTo(((latestPrice-entryPrice)/entryPrice)*100, 2)
}

func calculateStockRecommendationHistoryMaxDrawdownPct(quotes []stockRecommendationHistoryQuote, validFrom time.Time, latestQuote *stockRecommendationHistoryQuote) float64 {
	if len(quotes) == 0 || latestQuote == nil {
		return 0
	}
	startDate := dateOnlyTime(validFrom)
	endDate := dateOnlyTime(latestQuote.TradeDate)
	peak := 0.0
	maxDrawdown := 0.0
	for _, quote := range quotes {
		if quote.TradeDate.Before(startDate) || quote.TradeDate.After(endDate) {
			continue
		}
		if quote.ClosePrice > peak {
			peak = quote.ClosePrice
		}
		if peak <= 0 {
			continue
		}
		drawdown := ((quote.ClosePrice - peak) / peak) * 100
		if drawdown < maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	return roundTo(maxDrawdown, 2)
}

func mapStockRecommendationHistoryOutcome(status string, performanceLabel string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "HIT_TAKE_PROFIT":
		return "success"
	case "HIT_STOP_LOSS", "INVALIDATED":
		return "fail"
	case "ACTIVE", "TRACKING", "PUBLISHED":
		return "ongoing"
	case "REVIEWED":
		label := strings.ToUpper(strings.TrimSpace(performanceLabel))
		switch label {
		case "OUTPERFORM", "HIT", "GOOD", "SUCCESS":
			return "success"
		case "UNDERPERFORM", "MISS", "BAD", "FAIL":
			return "fail"
		default:
			return "neutral"
		}
	default:
		return "neutral"
	}
}

func buildStockRecommendationHistorySummary(items []model.StockRecommendationHistoryItem) model.StockRecommendationHistorySummary {
	summary := model.StockRecommendationHistorySummary{
		TotalCount: len(items),
	}
	returnSum := 0.0
	for _, item := range items {
		switch item.Outcome {
		case "success":
			summary.SuccessCount++
		case "fail":
			summary.FailCount++
		case "ongoing":
			summary.OngoingCount++
		default:
			summary.NeutralCount++
		}
		returnSum += item.ReturnPct
		if item.ReturnPct > summary.MaxReturnPct {
			summary.MaxReturnPct = item.ReturnPct
		}
		if item.MaxDrawdownPct < summary.MaxDrawdownPct {
			summary.MaxDrawdownPct = item.MaxDrawdownPct
		}
	}
	closedCount := summary.SuccessCount + summary.FailCount + summary.NeutralCount
	if closedCount > 0 {
		summary.WinRate = roundTo((float64(summary.SuccessCount)/float64(closedCount))*100, 2)
	}
	if summary.TotalCount > 0 {
		summary.AvgReturnPct = roundTo(returnSum/float64(summary.TotalCount), 2)
	}
	return summary
}

func stockRecommendationHistoryIsClosed(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED", "REVIEWED":
		return true
	default:
		return false
	}
}

func dateOnlyTime(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func (r *MySQLGrowthRepo) GetStockRecommendationDetail(userID string, recoID string) (model.StockRecommendationDetail, error) {
	var item model.StockRecommendationDetail
	var takeProfit, stopLoss, riskNote sql.NullString
	err := r.db.QueryRow(`
SELECT d.reco_id, d.tech_score, d.fund_score, d.sentiment_score, d.money_flow_score, d.take_profit, d.stop_loss, d.risk_note
FROM stock_reco_details d
JOIN stock_recommendations r ON r.id = d.reco_id
WHERE d.reco_id = ? AND r.status IN ('PUBLISHED', 'ACTIVE', 'TRACKING', 'HIT_TAKE_PROFIT', 'HIT_STOP_LOSS', 'INVALIDATED', 'REVIEWED')`, recoID).Scan(
		&item.RecoID, &item.TechScore, &item.FundScore, &item.SentimentScore, &item.MoneyFlowScore, &takeProfit, &stopLoss, &riskNote,
	)
	if err != nil {
		return model.StockRecommendationDetail{}, err
	}
	if takeProfit.Valid {
		item.TakeProfit = takeProfit.String
	}
	if stopLoss.Valid {
		item.StopLoss = stopLoss.String
	}
	if riskNote.Valid {
		item.RiskNote = riskNote.String
	}

	if userID != "" {
		allowed, err := r.CheckAndConsumeQuota(userID, "stock_reco", recoID)
		if err != nil {
			return model.StockRecommendationDetail{}, err
		}
		if !allowed {
			return model.StockRecommendationDetail{}, errors.New("quota exceeded")
		}

		_, _ = r.db.Exec(`
INSERT INTO browse_histories (id, user_id, content_type, content_id, source_page, viewed_at)
VALUES (?, ?, 'STOCK', ?, '/recommendation', ?)`,
			newID("bh"), userID, recoID, time.Now(),
		)
	}

	return item, nil
}

func (r *MySQLGrowthRepo) GetStockRecommendationPerformance(userID string, recoID string) ([]model.RecommendationPerformancePoint, error) {
	var score sql.NullFloat64
	var validFrom, validTo time.Time
	err := r.db.QueryRow(`
SELECT score, valid_from, valid_to
FROM stock_recommendations
WHERE id = ? AND status IN ('PUBLISHED', 'ACTIVE', 'TRACKING', 'HIT_TAKE_PROFIT', 'HIT_STOP_LOSS', 'INVALIDATED', 'REVIEWED', 'REALTIME_CACHED')`, recoID).Scan(&score, &validFrom, &validTo)
	if err != nil {
		return nil, err
	}

	base := 0.0
	if score.Valid {
		base = (score.Float64 - 50) / 100.0
	}
	points := make([]model.RecommendationPerformancePoint, 0, 5)
	for day := 0; day < 5; day++ {
		current := validFrom.AddDate(0, 0, day)
		if current.After(validTo) {
			break
		}
		value := base * (float64(day+1) / 5.0)
		value = float64(int(value*10000)) / 10000
		points = append(points, model.RecommendationPerformancePoint{
			Date:   current.Format("2006-01-02"),
			Return: value,
		})
	}
	if len(points) == 0 {
		points = append(points, model.RecommendationPerformancePoint{
			Date:   validFrom.Format("2006-01-02"),
			Return: base,
		})
	}
	return points, nil
}

func (r *MySQLGrowthRepo) GetStockRecommendationInsight(userID string, recoID string) (model.StockRecommendationInsight, error) {
	var item model.StockRecommendation
	var positionRange, reasonSummary sql.NullString
	var validFrom, validTo time.Time
	if err := r.db.QueryRow(`
SELECT id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary
FROM stock_recommendations
WHERE id = ? AND status IN ('PUBLISHED', 'ACTIVE', 'TRACKING', 'HIT_TAKE_PROFIT', 'HIT_STOP_LOSS', 'INVALIDATED', 'REVIEWED', 'REALTIME_CACHED')`, recoID).Scan(
		&item.ID, &item.Symbol, &item.Name, &item.Score, &item.RiskLevel, &positionRange, &validFrom, &validTo, &item.Status, &reasonSummary,
	); err != nil {
		return model.StockRecommendationInsight{}, err
	}
	if positionRange.Valid {
		item.PositionRange = positionRange.String
	}
	if reasonSummary.Valid {
		item.ReasonSummary = reasonSummary.String
	}
	item.ValidFrom = validFrom.Format(time.RFC3339)
	item.ValidTo = validTo.Format(time.RFC3339)

	detail, detailErr := r.GetStockRecommendationDetail(userID, recoID)
	if detailErr != nil {
		if errors.Is(detailErr, sql.ErrNoRows) {
			detail = estimateStockRecoDetailByScore(recoID, item.Score)
		} else {
			return model.StockRecommendationInsight{}, detailErr
		}
	}
	if detail.RecoID == "" {
		detail.RecoID = recoID
	}

	performancePoints, err := r.GetStockRecommendationPerformance(userID, recoID)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}
	benchmarkPoints, benchmarkSymbol, benchmarkSource, err := r.buildBenchmarkSeries(performancePoints, validFrom, validTo)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}
	performanceStats := summarizeStockPerformance(performancePoints, benchmarkPoints, benchmarkSymbol, benchmarkSource)

	scoreFramework := buildStockRecoScoreFramework(item.Score, detail)
	relatedNews, err := r.listStockRelatedNews(item.Symbol, item.Name, validFrom, 6)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}
	explanation := r.buildStockStrategyExplanation(item, detail)
	if contexts, contextErr := r.listStrategyEngineAssetContexts("stock-selection", item.Symbol, defaultVersionHistoryLimit); contextErr == nil {
		applyStrategyVersionDiffToExplanation(&explanation, contexts, item.Symbol)
		applyStrategyL1HistoryFromContexts(&explanation, contexts, item.Symbol)
	}

	return model.StockRecommendationInsight{
		Recommendation:   item,
		Detail:           detail,
		ScoreFramework:   scoreFramework,
		RelatedNews:      relatedNews,
		Performance:      performancePoints,
		Benchmark:        benchmarkPoints,
		PerformanceStats: performanceStats,
		Explanation:      explanation,
		GeneratedAt:      time.Now().Format(time.RFC3339),
	}, nil
}

func (r *MySQLGrowthRepo) GetStockRecommendationVersionHistory(userID string, recoID string) ([]model.StrategyVersionHistoryItem, error) {
	var item model.StockRecommendation
	var validFrom time.Time
	var reasonSummary, strategyVersion sql.NullString
	if err := r.db.QueryRow(`
SELECT id, symbol, valid_from, reason_summary, strategy_version
FROM stock_recommendations
WHERE id = ? AND status IN ('PUBLISHED', 'ACTIVE', 'TRACKING', 'HIT_TAKE_PROFIT', 'HIT_STOP_LOSS', 'INVALIDATED', 'REVIEWED', 'REALTIME_CACHED')`,
		recoID,
	).Scan(&item.ID, &item.Symbol, &validFrom, &reasonSummary, &strategyVersion); err != nil {
		return nil, err
	}
	item.ValidFrom = validFrom.Format(time.RFC3339)
	if reasonSummary.Valid {
		item.ReasonSummary = reasonSummary.String
	}
	if strategyVersion.Valid {
		item.StrategyVersion = strategyVersion.String
	}

	detail, detailErr := r.GetStockRecommendationDetail(userID, recoID)
	if detailErr != nil && !errors.Is(detailErr, sql.ErrNoRows) {
		return nil, detailErr
	}
	explanation := r.buildStockStrategyExplanation(item, detail)
	l3Config := r.loadForecastL3RuntimeConfig()
	contexts, err := r.listStrategyEngineAssetContexts("stock-selection", item.Symbol, defaultVersionHistoryLimit)
	if err != nil {
		return nil, err
	}
	if len(contexts) == 0 {
		items := []model.StrategyVersionHistoryItem{
			buildFallbackVersionHistoryItem(
				explanation.PublishID,
				explanation.JobID,
				dateOnly(item.ValidFrom),
				explanation.PublishVersion,
				explanation.GeneratedAt,
				item.StrategyVersion,
				item.ReasonSummary,
				explanation,
			),
		}
		if l3Config.ClientReadEnabled {
			r.attachLatestStrategyForecastL3ToHistoryItems(items, model.StrategyForecastL3TargetTypeStock, item.Symbol)
		}
		return items, nil
	}

	items := make([]model.StrategyVersionHistoryItem, 0, len(contexts))
	l1Config := r.loadForecastL1RuntimeConfig()
	l2Config := r.loadForecastL2RuntimeConfig()
	for _, ctx := range contexts {
		historyItem := buildStrategyVersionHistoryItem(
			ctx,
			asString(ctx.asset["reason_summary"]),
			firstNonEmpty(asString(ctx.asset["strategy_version"]), item.StrategyVersion),
			[]string{
				"市场种子输入",
				"特征工程与打分",
				"多角色评审",
				"情景模拟",
				"风险过滤与发布",
			},
		)
		if summary, summaryErr := r.loadStockSelectionEvaluationSummaryByContext(ctx, item.Symbol); summaryErr == nil {
			enrichStockSelectionVersionHistoryEvaluationMeta(&historyItem, summary)
		}
		applyForecastL1DisplayConfigToVersionHistoryItem(&historyItem, mapValue(ctx.record.ReportSnapshot["memory_feedback"]), l1Config)
		applyForecastL2DisplayConfigToVersionHistoryItem(&historyItem, l2Config)
		items = append(items, historyItem)
	}
	applyStrategyVersionDiffToHistoryItems(items, contexts, item.Symbol)
	applyStrategyL1HistoryToHistoryItems(items, contexts, item.Symbol)
	if l3Config.ClientReadEnabled {
		r.attachLatestStrategyForecastL3ToHistoryItems(items, model.StrategyForecastL3TargetTypeStock, item.Symbol)
	}
	return items, nil
}

func estimateStockRecoDetailByScore(recoID string, totalScore float64) model.StockRecommendationDetail {
	base := clampFloat(totalScore, 55, 98)
	return model.StockRecommendationDetail{
		RecoID:         recoID,
		TechScore:      roundTo(clampFloat(base-2, 50, 99), 2),
		FundScore:      roundTo(clampFloat(base+1, 50, 99), 2),
		SentimentScore: roundTo(clampFloat(base-4, 50, 99), 2),
		MoneyFlowScore: roundTo(clampFloat(base-1, 50, 99), 2),
		TakeProfit:     "上涨 6%-10% 分批止盈",
		StopLoss:       "跌破关键支撑位止损",
		RiskNote:       "评分由推荐总分估算，建议结合盘中波动与仓位纪律执行。",
	}
}

func buildStockRecoScoreFramework(totalScore float64, detail model.StockRecommendationDetail) model.StockRecommendationScoreFramework {
	factors := []model.StockRecommendationFactorScore{
		{Key: "tech", Label: "技术因子", Weight: 0.30, Score: detail.TechScore},
		{Key: "fund", Label: "基本面因子", Weight: 0.30, Score: detail.FundScore},
		{Key: "sentiment", Label: "情绪因子", Weight: 0.20, Score: detail.SentimentScore},
		{Key: "flow", Label: "资金流因子", Weight: 0.20, Score: detail.MoneyFlowScore},
	}
	sumWeight := 0.0
	weightedScore := 0.0
	for index := range factors {
		if factors[index].Score <= 0 {
			continue
		}
		factors[index].Contribution = roundTo(factors[index].Weight*factors[index].Score, 2)
		weightedScore += factors[index].Contribution
		sumWeight += factors[index].Weight
	}
	if sumWeight > 0 {
		weightedScore = weightedScore / sumWeight
	}
	weightedScore = roundTo(weightedScore, 2)
	return model.StockRecommendationScoreFramework{
		Method:        "growth-v1 (tech30 + fund30 + sentiment20 + flow20)",
		TotalScore:    roundTo(totalScore, 2),
		WeightedScore: weightedScore,
		ScoreGap:      roundTo(totalScore-weightedScore, 2),
		Factors:       factors,
	}
}

func buildEstimatedBenchmark(points []model.RecommendationPerformancePoint) []model.RecommendationPerformancePoint {
	result := make([]model.RecommendationPerformancePoint, 0, len(points))
	for _, point := range points {
		result = append(result, model.RecommendationPerformancePoint{
			Date:   point.Date,
			Return: roundTo(clampFloat(point.Return*0.55, -0.2, 0.2), 4),
		})
	}
	return result
}

func (r *MySQLGrowthRepo) buildBenchmarkSeries(points []model.RecommendationPerformancePoint, validFrom time.Time, validTo time.Time) ([]model.RecommendationPerformancePoint, string, string, error) {
	estimated := buildEstimatedBenchmark(points)
	if len(points) == 0 {
		return estimated, "000300.SH", "estimated: no strategy points", nil
	}

	startDate := validFrom
	endDate := validTo
	if startDate.IsZero() || endDate.IsZero() || endDate.Before(startDate) {
		for index := range points {
			ts, err := time.Parse("2006-01-02", points[index].Date)
			if err != nil {
				continue
			}
			if startDate.IsZero() || ts.Before(startDate) {
				startDate = ts
			}
			if endDate.IsZero() || ts.After(endDate) {
				endDate = ts
			}
		}
	}
	if startDate.IsZero() || endDate.IsZero() || endDate.Before(startDate) {
		return estimated, "000300.SH", "estimated: invalid date range", nil
	}

	candidates := []struct {
		Symbol string
		Name   string
	}{
		{Symbol: "000300.SH", Name: "CSI300"},
		{Symbol: "000905.SH", Name: "CSI500"},
	}

	rows, err := r.db.Query(`
SELECT symbol, trade_date, close_price, prev_close_price
FROM stock_market_quotes
WHERE symbol IN (?, ?)
  AND trade_date >= ?
  AND trade_date <= ?
ORDER BY symbol ASC, trade_date ASC`,
		candidates[0].Symbol, candidates[1].Symbol, startDate, endDate,
	)
	if err != nil {
		if isTableNotFoundError(err) {
			return estimated, "000300.SH", "estimated: stock_market_quotes table missing", nil
		}
		return nil, "", "", err
	}
	defer rows.Close()

	quoteMap := map[string]map[string]float64{
		candidates[0].Symbol: {},
		candidates[1].Symbol: {},
	}
	lastClose := map[string]float64{}
	for rows.Next() {
		var symbol string
		var tradeDate time.Time
		var closePrice, prevClosePrice sql.NullFloat64
		if err := rows.Scan(&symbol, &tradeDate, &closePrice, &prevClosePrice); err != nil {
			return nil, "", "", err
		}
		if !closePrice.Valid || closePrice.Float64 <= 0 {
			continue
		}
		ret := math.NaN()
		if prevClosePrice.Valid && prevClosePrice.Float64 > 0 {
			ret = closePrice.Float64/prevClosePrice.Float64 - 1
		} else if previous, exists := lastClose[symbol]; exists && previous > 0 {
			ret = closePrice.Float64/previous - 1
		}
		lastClose[symbol] = closePrice.Float64
		if !math.IsNaN(ret) {
			dateKey := tradeDate.Format("2006-01-02")
			quoteMap[symbol][dateKey] = roundTo(clampFloat(ret, -0.2, 0.2), 4)
		}
	}

	perfDates := make([]string, 0, len(points))
	for _, point := range points {
		if point.Date == "" {
			continue
		}
		dateKey := point.Date
		if ts, err := time.Parse("2006-01-02", point.Date); err == nil {
			dateKey = ts.Format("2006-01-02")
		}
		perfDates = append(perfDates, dateKey)
	}

	bestSymbol := candidates[0].Symbol
	bestName := candidates[0].Name
	bestCoverage := -1
	for _, candidate := range candidates {
		coverage := 0
		for _, date := range perfDates {
			if _, ok := quoteMap[candidate.Symbol][date]; ok {
				coverage++
			}
		}
		if coverage > bestCoverage {
			bestCoverage = coverage
			bestSymbol = candidate.Symbol
			bestName = candidate.Name
		}
	}
	if bestCoverage <= 0 {
		return estimated, "000300.SH", "estimated: no benchmark quotes matched", nil
	}

	result := make([]model.RecommendationPerformancePoint, 0, len(points))
	actualCount := 0
	for index, point := range points {
		dateKey := point.Date
		if ts, err := time.Parse("2006-01-02", point.Date); err == nil {
			dateKey = ts.Format("2006-01-02")
		}
		value := estimated[index].Return
		if actual, ok := quoteMap[bestSymbol][dateKey]; ok {
			value = actual
			actualCount++
		}
		result = append(result, model.RecommendationPerformancePoint{
			Date:   point.Date,
			Return: value,
		})
	}

	source := fmt.Sprintf("actual: %s (%d/%d points), missing filled by estimated", bestName, actualCount, len(points))
	if actualCount == len(points) {
		source = fmt.Sprintf("actual: %s", bestName)
	}
	return result, bestSymbol, source, nil
}

func summarizeStockPerformance(points []model.RecommendationPerformancePoint, benchmark []model.RecommendationPerformancePoint, benchmarkSymbol string, benchmarkSource string) model.StockRecommendationPerformanceSummary {
	if len(points) == 0 {
		return model.StockRecommendationPerformanceSummary{
			BenchmarkSymbol: benchmarkSymbol,
			BenchmarkSource: benchmarkSource,
		}
	}
	sampleDays := len(points)
	wins := 0
	sumDaily := 0.0
	curve := 1.0
	peak := 1.0
	maxDrawdown := 0.0
	for _, point := range points {
		if point.Return > 0 {
			wins++
		}
		sumDaily += point.Return
		curve *= 1 + point.Return
		if curve > peak {
			peak = curve
		}
		if peak > 0 {
			dd := (peak - curve) / peak
			if dd > maxDrawdown {
				maxDrawdown = dd
			}
		}
	}

	benchmarkCurve := 1.0
	for _, point := range benchmark {
		benchmarkCurve *= 1 + point.Return
	}

	cumulative := curve - 1
	benchmarkCumulative := benchmarkCurve - 1
	return model.StockRecommendationPerformanceSummary{
		SampleDays:                sampleDays,
		WinRate:                   roundTo(float64(wins)/float64(sampleDays), 4),
		AvgDailyReturn:            roundTo(sumDaily/float64(sampleDays), 4),
		CumulativeReturn:          roundTo(cumulative, 4),
		BenchmarkCumulativeReturn: roundTo(benchmarkCumulative, 4),
		ExcessReturn:              roundTo(cumulative-benchmarkCumulative, 4),
		MaxDrawdown:               roundTo(maxDrawdown, 4),
		BenchmarkSymbol:           benchmarkSymbol,
		BenchmarkSource:           benchmarkSource,
	}
}

func (r *MySQLGrowthRepo) listStockRelatedNews(symbol string, name string, anchor time.Time, limit int) ([]model.StockRecommendationRelatedNews, error) {
	keywords := buildStockNewsKeywords(symbol, name)
	if len(keywords) == 0 {
		return []model.StockRecommendationRelatedNews{}, nil
	}
	if limit <= 0 {
		limit = 6
	}
	if anchor.IsZero() {
		anchor = time.Now()
	}
	windowStart := anchor.AddDate(0, 0, -45)

	conditions := make([]string, 0, len(keywords))
	args := make([]interface{}, 0, 1+len(keywords)*3+1)
	args = append(args, windowStart)
	for _, keyword := range keywords {
		conditions = append(conditions, "(na.title LIKE ? OR COALESCE(na.summary, '') LIKE ? OR COALESCE(na.content, '') LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}

	query := `
SELECT na.id, na.title, na.summary, nc.name AS source, na.visibility, na.published_at
FROM news_articles na
LEFT JOIN news_categories nc ON nc.id = na.category_id
WHERE na.status = 'PUBLISHED'
  AND na.published_at >= ?
  AND (` + strings.Join(conditions, " OR ") + `)
ORDER BY na.published_at DESC
LIMIT ?`
	args = append(args, limit*4)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StockRecommendationRelatedNews, 0, limit)
	for rows.Next() {
		var item model.StockRecommendationRelatedNews
		var summary, source sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Title, &summary, &source, &item.Visibility, &publishedAt); err != nil {
			return nil, err
		}
		if summary.Valid {
			item.Summary = truncateByRunes(normalizeUTF8Text(summary.String), 120)
		}
		if source.Valid {
			item.Source = source.String
		}
		if item.Source == "" {
			item.Source = "资讯中心"
		}
		if publishedAt.Valid {
			item.PublishedAt = publishedAt.Time.Format(time.RFC3339)
			item.RelevanceScore = scoreNewsRelevance(item.Title, item.Summary, keywords, publishedAt.Time, anchor)
		} else {
			item.RelevanceScore = scoreNewsRelevance(item.Title, item.Summary, keywords, time.Time{}, anchor)
		}
		items = append(items, item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].RelevanceScore == items[j].RelevanceScore {
			return items[i].PublishedAt > items[j].PublishedAt
		}
		return items[i].RelevanceScore > items[j].RelevanceScore
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func buildStockNewsKeywords(symbol string, name string) []string {
	items := make([]string, 0, 4)
	seen := make(map[string]struct{})
	push := func(value string) {
		key := strings.TrimSpace(value)
		if key == "" {
			return
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		items = append(items, key)
	}

	normalizedName := strings.TrimSpace(name)
	push(normalizedName)

	rawSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	push(rawSymbol)
	if idx := strings.Index(rawSymbol, "."); idx > 0 {
		push(rawSymbol[:idx])
	}
	return items
}

func scoreNewsRelevance(title string, summary string, keywords []string, publishedAt time.Time, anchor time.Time) float64 {
	titleLower := strings.ToLower(title)
	summaryLower := strings.ToLower(summary)
	score := 0.0
	for _, keyword := range keywords {
		key := strings.ToLower(strings.TrimSpace(keyword))
		if key == "" {
			continue
		}
		if strings.Contains(titleLower, key) {
			score += 0.45
		}
		if strings.Contains(summaryLower, key) {
			score += 0.20
		}
	}

	if !publishedAt.IsZero() {
		diff := anchor.Sub(publishedAt)
		if diff < 0 {
			diff = -diff
		}
		hours := diff.Hours()
		if hours <= 72 {
			score += 0.2
		} else if hours <= 168 {
			score += 0.1
		}
	}
	return roundTo(clampFloat(score, 0, 1), 4)
}
