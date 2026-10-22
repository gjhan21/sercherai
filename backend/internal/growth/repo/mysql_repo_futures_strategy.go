package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) ListFuturesStrategies(userID string, contract string, status string, page int, pageSize int) ([]model.FuturesStrategy, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if contract != "" {
		filter += " AND contract = ?"
		args = append(args, contract)
	}
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	} else {
		filter += " AND status IN ('PUBLISHED', 'ACTIVE')"
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM futures_strategies"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary
FROM futures_strategies` + filter + `
ORDER BY valid_from DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.FuturesStrategy, 0)
	for rows.Next() {
		var item model.FuturesStrategy
		var name, positionRange, reasonSummary sql.NullString
		var validFrom, validTo time.Time
		if err := rows.Scan(&item.ID, &item.Contract, &name, &item.Direction, &item.RiskLevel, &positionRange, &validFrom, &validTo, &item.Status, &reasonSummary); err != nil {
			return nil, 0, err
		}
		if name.Valid {
			item.Name = name.String
		}
		if positionRange.Valid {
			item.PositionRange = positionRange.String
		}
		if reasonSummary.Valid {
			item.ReasonSummary = reasonSummary.String
		}
		item.ValidFrom = validFrom.Format(time.RFC3339)
		item.ValidTo = validTo.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) GetFuturesStrategyDetail(userID string, strategyID string) (model.FuturesStrategy, error) {
	var item model.FuturesStrategy
	var name, positionRange, reasonSummary sql.NullString
	var validFrom, validTo time.Time
	err := r.db.QueryRow(`
SELECT id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary
FROM futures_strategies
WHERE id = ?`, strategyID).Scan(
		&item.ID, &item.Contract, &name, &item.Direction, &item.RiskLevel, &positionRange, &validFrom, &validTo, &item.Status, &reasonSummary,
	)
	if err != nil {
		return model.FuturesStrategy{}, err
	}
	if name.Valid {
		item.Name = name.String
	}
	if positionRange.Valid {
		item.PositionRange = positionRange.String
	}
	if reasonSummary.Valid {
		item.ReasonSummary = reasonSummary.String
	}
	item.ValidFrom = validFrom.Format(time.RFC3339)
	item.ValidTo = validTo.Format(time.RFC3339)
	return item, nil
}

func (r *MySQLGrowthRepo) GetFuturesStrategyInsight(userID string, strategyID string) (model.FuturesStrategyInsight, error) {
	strategy, err := r.GetFuturesStrategyDetail(userID, strategyID)
	if err != nil {
		return model.FuturesStrategyInsight{}, err
	}

	guidance, _ := r.getLatestFuturesGuidanceByContract(strategy.Contract)

	validFrom, validTo := parseStrategyWindow(strategy.ValidFrom, strategy.ValidTo)
	performance, statsFromReview, err := r.buildFuturesStrategyPerformance(strategyID, strategy.RiskLevel, validFrom, validTo)
	if err != nil {
		return model.FuturesStrategyInsight{}, err
	}

	benchmarkSymbol := futuresBenchmarkSymbolByContract(strategy.Contract)
	benchmark, benchmarkSource, err := r.buildFuturesBenchmarkSeries(performance, benchmarkSymbol)
	if err != nil {
		return model.FuturesStrategyInsight{}, err
	}
	stats := summarizeFuturesPerformance(performance, benchmark, benchmarkSymbol, benchmarkSource, statsFromReview.MaxDrawdown)

	relatedEvents, err := r.listFuturesRelatedEvents(strategy.Contract, 6)
	if err != nil {
		return model.FuturesStrategyInsight{}, err
	}
	anchor := validFrom
	if anchor.IsZero() {
		anchor = time.Now()
	}
	relatedNews, err := r.listStockRelatedNews(strategy.Contract, strategy.Name, anchor, 6)
	if err != nil {
		return model.FuturesStrategyInsight{}, err
	}

	scoreWeights := r.loadFuturesStrategyScoreWeights()
	scoreFramework := buildFuturesScoreFramework(strategy, guidance, stats, relatedNews, relatedEvents, scoreWeights)
	explanation := r.buildFuturesStrategyExplanation(strategy, guidance)
	if contexts, contextErr := r.listStrategyEngineAssetContexts("futures-strategy", strategy.Contract, defaultVersionHistoryLimit); contextErr == nil {
		applyStrategyVersionDiffToExplanation(&explanation, contexts, strategy.Contract)
		applyStrategyL1HistoryFromContexts(&explanation, contexts, strategy.Contract)
	}
	return model.FuturesStrategyInsight{
		Strategy:         strategy,
		Guidance:         guidance,
		ScoreFramework:   scoreFramework,
		RelatedNews:      relatedNews,
		RelatedEvents:    relatedEvents,
		Performance:      performance,
		Benchmark:        benchmark,
		PerformanceStats: stats,
		Explanation:      explanation,
		GeneratedAt:      time.Now().Format(time.RFC3339),
	}, nil
}

func (r *MySQLGrowthRepo) GetFuturesStrategyVersionHistory(userID string, strategyID string) ([]model.StrategyVersionHistoryItem, error) {
	strategy, err := r.GetFuturesStrategyDetail(userID, strategyID)
	if err != nil {
		return nil, err
	}
	guidance, _ := r.getLatestFuturesGuidanceByContract(strategy.Contract)
	explanation := r.buildFuturesStrategyExplanation(strategy, guidance)
	l3Config := r.loadForecastL3RuntimeConfig()
	contexts, err := r.listStrategyEngineAssetContexts("futures-strategy", strategy.Contract, defaultVersionHistoryLimit)
	if err != nil {
		return nil, err
	}
	if len(contexts) == 0 {
		items := []model.StrategyVersionHistoryItem{
			buildFallbackVersionHistoryItem(
				explanation.PublishID,
				explanation.JobID,
				dateOnly(strategy.ValidFrom),
				explanation.PublishVersion,
				explanation.GeneratedAt,
				explanation.StrategyVersion,
				strategy.ReasonSummary,
				explanation,
			),
		}
		if l3Config.ClientReadEnabled {
			r.attachLatestStrategyForecastL3ToHistoryItems(items, model.StrategyForecastL3TargetTypeFutures, strategy.Contract)
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
			firstNonEmpty(asString(ctx.asset["strategy_version"]), explanation.StrategyVersion, "futures-mvp-v1"),
			[]string{
				"合约池初始化",
				"方向/价位特征评估",
				"多角色评审",
				"情景推演",
				"风险与发布过滤",
			},
		)
		applyForecastL1DisplayConfigToVersionHistoryItem(&historyItem, mapValue(ctx.record.ReportSnapshot["memory_feedback"]), l1Config)
		applyForecastL2DisplayConfigToVersionHistoryItem(&historyItem, l2Config)
		items = append(items, historyItem)
	}
	applyStrategyVersionDiffToHistoryItems(items, contexts, strategy.Contract)
	applyStrategyL1HistoryToHistoryItems(items, contexts, strategy.Contract)
	if l3Config.ClientReadEnabled {
		r.attachLatestStrategyForecastL3ToHistoryItems(items, model.StrategyForecastL3TargetTypeFutures, strategy.Contract)
	}
	return items, nil
}

func (r *MySQLGrowthRepo) getLatestFuturesGuidanceByContract(contract string) (model.FuturesGuidance, error) {
	var item model.FuturesGuidance
	var entryRange, takeProfitRange, stopLossRange, invalidCondition sql.NullString
	var validTo time.Time
	err := r.db.QueryRow(`
SELECT contract, guidance_direction, position_level, entry_range, take_profit_range, stop_loss_range, risk_level, invalid_condition, valid_to
FROM futures_guidances
WHERE contract = ?
ORDER BY valid_to DESC
LIMIT 1`, strings.TrimSpace(contract)).Scan(
		&item.Contract, &item.GuidanceDirection, &item.PositionLevel, &entryRange, &takeProfitRange, &stopLossRange, &item.RiskLevel, &invalidCondition, &validTo,
	)
	if err != nil {
		return model.FuturesGuidance{}, err
	}
	if entryRange.Valid {
		item.EntryRange = entryRange.String
	}
	if takeProfitRange.Valid {
		item.TakeProfitRange = takeProfitRange.String
	}
	if stopLossRange.Valid {
		item.StopLossRange = stopLossRange.String
	}
	if invalidCondition.Valid {
		item.InvalidCondition = invalidCondition.String
	}
	item.ValidTo = validTo.Format(time.RFC3339)
	return item, nil
}

type futuresPerformanceBuildMeta struct {
	MaxDrawdown float64
}

func (r *MySQLGrowthRepo) buildFuturesStrategyPerformance(strategyID string, riskLevel string, validFrom time.Time, validTo time.Time) ([]model.RecommendationPerformancePoint, futuresPerformanceBuildMeta, error) {
	rows, err := r.db.Query(`
SELECT review_date, pnl, max_drawdown
FROM futures_reviews
WHERE strategy_id = ?
ORDER BY review_date ASC`, strategyID)
	if err != nil {
		if !isTableNotFoundError(err) {
			return nil, futuresPerformanceBuildMeta{}, err
		}
		rows = nil
	}

	points := make([]model.RecommendationPerformancePoint, 0, 12)
	meta := futuresPerformanceBuildMeta{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var reviewDate time.Time
			var pnl, maxDrawdown sql.NullFloat64
			if err := rows.Scan(&reviewDate, &pnl, &maxDrawdown); err != nil {
				return nil, futuresPerformanceBuildMeta{}, err
			}
			value := normalizeFuturesPnLToReturn(pnl)
			points = append(points, model.RecommendationPerformancePoint{
				Date:   reviewDate.Format("2006-01-02"),
				Return: roundTo(value, 4),
			})
			if maxDrawdown.Valid {
				meta.MaxDrawdown = math.Max(meta.MaxDrawdown, math.Abs(maxDrawdown.Float64))
			}
		}
	}

	if len(points) > 0 {
		return points, meta, nil
	}
	return estimateFuturesPerformanceByRisk(riskLevel, validFrom, validTo), meta, nil
}

func normalizeFuturesPnLToReturn(pnl sql.NullFloat64) float64 {
	if !pnl.Valid {
		return 0
	}
	value := pnl.Float64
	if math.Abs(value) > 1.5 {
		value = value / 100.0
	}
	return clampFloat(value, -0.3, 0.3)
}

func estimateFuturesPerformanceByRisk(riskLevel string, validFrom time.Time, validTo time.Time) []model.RecommendationPerformancePoint {
	base := 0.009
	switch strings.ToUpper(strings.TrimSpace(riskLevel)) {
	case "LOW":
		base = 0.006
	case "HIGH":
		base = 0.013
	}

	start := validFrom
	if start.IsZero() {
		start = time.Now().AddDate(0, 0, -4)
	}
	end := validTo
	if end.IsZero() || end.Before(start) {
		end = start.AddDate(0, 0, 4)
	}

	points := make([]model.RecommendationPerformancePoint, 0, 5)
	current := start
	for index := 0; index < 5; index++ {
		if current.After(end) {
			break
		}
		value := base * (0.8 + float64(index)*0.12)
		if index == 2 {
			value = -value * 0.45
		}
		points = append(points, model.RecommendationPerformancePoint{
			Date:   current.Format("2006-01-02"),
			Return: roundTo(clampFloat(value, -0.3, 0.3), 4),
		})
		current = current.AddDate(0, 0, 1)
	}
	if len(points) == 0 {
		points = append(points, model.RecommendationPerformancePoint{
			Date:   start.Format("2006-01-02"),
			Return: roundTo(base, 4),
		})
	}
	return points
}

func parseStrategyWindow(validFrom string, validTo string) (time.Time, time.Time) {
	start, _ := time.Parse(time.RFC3339, strings.TrimSpace(validFrom))
	end, _ := time.Parse(time.RFC3339, strings.TrimSpace(validTo))
	return start, end
}

func futuresBenchmarkSymbolByContract(contract string) string {
	source := strings.ToUpper(strings.TrimSpace(contract))
	switch {
	case strings.HasPrefix(source, "IF"):
		return "000300.SH"
	case strings.HasPrefix(source, "IC"):
		return "000905.SH"
	case strings.HasPrefix(source, "IH"):
		return "000016.SH"
	case strings.HasPrefix(source, "IM"):
		return "000852.SH"
	default:
		return "000300.SH"
	}
}

func (r *MySQLGrowthRepo) buildFuturesBenchmarkSeries(points []model.RecommendationPerformancePoint, benchmarkSymbol string) ([]model.RecommendationPerformancePoint, string, error) {
	estimated := buildEstimatedBenchmark(points)
	if len(points) == 0 {
		return estimated, "estimated: no strategy points", nil
	}
	if benchmarkSymbol == "" {
		return estimated, "estimated: benchmark symbol missing", nil
	}

	minDate, maxDate := "", ""
	for _, point := range points {
		dateKey := strings.TrimSpace(point.Date)
		if dateKey == "" {
			continue
		}
		if minDate == "" || dateKey < minDate {
			minDate = dateKey
		}
		if maxDate == "" || dateKey > maxDate {
			maxDate = dateKey
		}
	}
	if minDate == "" || maxDate == "" {
		return estimated, "estimated: invalid point dates", nil
	}

	rows, err := r.db.Query(`
SELECT trade_date, close_price, prev_close_price
FROM stock_market_quotes
WHERE symbol = ?
  AND trade_date >= ?
  AND trade_date <= ?
ORDER BY trade_date ASC`, benchmarkSymbol, minDate, maxDate)
	if err != nil {
		if isTableNotFoundError(err) {
			return estimated, "estimated: stock_market_quotes table missing", nil
		}
		return nil, "", err
	}
	defer rows.Close()

	valueByDate := map[string]float64{}
	lastClose := 0.0
	for rows.Next() {
		var tradeDate time.Time
		var closePrice, prevClosePrice sql.NullFloat64
		if err := rows.Scan(&tradeDate, &closePrice, &prevClosePrice); err != nil {
			return nil, "", err
		}
		if !closePrice.Valid || closePrice.Float64 <= 0 {
			continue
		}
		ret := math.NaN()
		if prevClosePrice.Valid && prevClosePrice.Float64 > 0 {
			ret = closePrice.Float64/prevClosePrice.Float64 - 1
		} else if lastClose > 0 {
			ret = closePrice.Float64/lastClose - 1
		}
		lastClose = closePrice.Float64
		if !math.IsNaN(ret) {
			valueByDate[tradeDate.Format("2006-01-02")] = roundTo(clampFloat(ret, -0.2, 0.2), 4)
		}
	}

	result := make([]model.RecommendationPerformancePoint, 0, len(points))
	actualCount := 0
	for index, point := range points {
		value := estimated[index].Return
		if actual, ok := valueByDate[point.Date]; ok {
			value = actual
			actualCount++
		}
		result = append(result, model.RecommendationPerformancePoint{
			Date:   point.Date,
			Return: value,
		})
	}

	if actualCount == 0 {
		return result, "estimated: no benchmark quotes matched", nil
	}
	if actualCount < len(points) {
		return result, fmt.Sprintf("actual mixed: %d/%d points from %s", actualCount, len(points), benchmarkSymbol), nil
	}
	return result, fmt.Sprintf("actual: %s", benchmarkSymbol), nil
}

func summarizeFuturesPerformance(points []model.RecommendationPerformancePoint, benchmark []model.RecommendationPerformancePoint, benchmarkSymbol string, benchmarkSource string, overrideMaxDrawdown float64) model.FuturesStrategyPerformanceSummary {
	if len(points) == 0 {
		return model.FuturesStrategyPerformanceSummary{
			BenchmarkSymbol: benchmarkSymbol,
			BenchmarkSource: benchmarkSource,
		}
	}

	sampleDays := len(points)
	positiveDays := 0
	sumDaily := 0.0
	curve := 1.0
	peak := 1.0
	maxDrawdown := 0.0
	for _, point := range points {
		if point.Return > 0 {
			positiveDays++
		}
		sumDaily += point.Return
		curve *= 1 + point.Return
		if curve > peak {
			peak = curve
		}
		if peak > 0 {
			drawdown := (peak - curve) / peak
			if drawdown > maxDrawdown {
				maxDrawdown = drawdown
			}
		}
	}
	if overrideMaxDrawdown > 0 {
		maxDrawdown = math.Max(maxDrawdown, clampFloat(overrideMaxDrawdown, 0, 0.99))
	}

	benchmarkCurve := 1.0
	for _, point := range benchmark {
		benchmarkCurve *= 1 + point.Return
	}
	cumulative := curve - 1
	benchmarkCumulative := benchmarkCurve - 1
	return model.FuturesStrategyPerformanceSummary{
		SampleDays:                sampleDays,
		WinRate:                   roundTo(float64(positiveDays)/float64(sampleDays), 4),
		AvgDailyReturn:            roundTo(sumDaily/float64(sampleDays), 4),
		CumulativeReturn:          roundTo(cumulative, 4),
		BenchmarkCumulativeReturn: roundTo(benchmarkCumulative, 4),
		ExcessReturn:              roundTo(cumulative-benchmarkCumulative, 4),
		MaxDrawdown:               roundTo(maxDrawdown, 4),
		BenchmarkSymbol:           benchmarkSymbol,
		BenchmarkSource:           benchmarkSource,
	}
}

func (r *MySQLGrowthRepo) listFuturesRelatedEvents(contract string, limit int) ([]model.MarketEvent, error) {
	if limit <= 0 {
		limit = 6
	}
	rows, err := r.db.Query(`
SELECT id, event_type, symbol, summary, trigger_rule, source, created_at
FROM market_events
WHERE symbol = ? OR symbol = 'ALL'
ORDER BY created_at DESC
LIMIT ?`, strings.TrimSpace(contract), limit)
	if err != nil {
		if isTableNotFoundError(err) {
			return []model.MarketEvent{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	items := make([]model.MarketEvent, 0, limit)
	for rows.Next() {
		var item model.MarketEvent
		var summary, triggerRule, source sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.EventType, &item.Symbol, &summary, &triggerRule, &source, &createdAt); err != nil {
			return nil, err
		}
		if summary.Valid {
			item.Summary = summary.String
		}
		if triggerRule.Valid {
			item.TriggerRule = triggerRule.String
		}
		if source.Valid {
			item.Source = source.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, nil
}

const futuresStrategyScoreWeightsConfigKey = "futures.strategy.score_weights"

var futuresStrategyScoreWeightKeys = []string{
	"trend",
	"structure",
	"flow",
	"risk",
	"news",
	"performance",
}

func defaultFuturesStrategyScoreWeights() map[string]float64 {
	return map[string]float64{
		"trend":       0.25,
		"structure":   0.20,
		"flow":        0.15,
		"risk":        0.20,
		"news":        0.10,
		"performance": 0.10,
	}
}

func normalizeFuturesStrategyScoreWeightKey(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "trend", "trend_factor":
		return "trend"
	case "structure", "structure_factor":
		return "structure"
	case "flow", "flow_factor", "event", "events":
		return "flow"
	case "risk", "risk_factor", "risk_control":
		return "risk"
	case "news", "news_factor":
		return "news"
	case "performance", "performance_factor", "perf":
		return "performance"
	default:
		return ""
	}
}

func normalizeFuturesStrategyScoreWeights(raw map[string]float64) map[string]float64 {
	weights := defaultFuturesStrategyScoreWeights()
	if raw == nil {
		return weights
	}
	sum := 0.0
	for _, key := range futuresStrategyScoreWeightKeys {
		if value, ok := raw[key]; ok {
			weights[key] = clampFloat(value, 0, 1)
		}
		sum += weights[key]
	}
	if sum <= 0 {
		return defaultFuturesStrategyScoreWeights()
	}
	for _, key := range futuresStrategyScoreWeightKeys {
		weights[key] = weights[key] / sum
	}
	return weights
}

func parseFloatFromInterface(value interface{}) (float64, bool) {
	switch raw := value.(type) {
	case float64:
		if !math.IsNaN(raw) && !math.IsInf(raw, 0) {
			return raw, true
		}
	case float32:
		parsed := float64(raw)
		if !math.IsNaN(parsed) && !math.IsInf(parsed, 0) {
			return parsed, true
		}
	case int:
		return float64(raw), true
	case int32:
		return float64(raw), true
	case int64:
		return float64(raw), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err == nil && !math.IsNaN(parsed) && !math.IsInf(parsed, 0) {
			return parsed, true
		}
	}
	return 0, false
}

func parseFuturesStrategyScoreWeights(raw string) (map[string]float64, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, false
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return nil, false
	}
	weights := defaultFuturesStrategyScoreWeights()
	hasAny := false
	for key, value := range payload {
		normalizedKey := normalizeFuturesStrategyScoreWeightKey(key)
		if normalizedKey == "" {
			continue
		}
		parsed, ok := parseFloatFromInterface(value)
		if !ok {
			continue
		}
		if parsed > 1 {
			parsed = parsed / 100
		}
		weights[normalizedKey] = clampFloat(parsed, 0, 1)
		hasAny = true
	}
	if !hasAny {
		return nil, false
	}
	return normalizeFuturesStrategyScoreWeights(weights), true
}

func (r *MySQLGrowthRepo) loadFuturesStrategyScoreWeights() map[string]float64 {
	defaults := defaultFuturesStrategyScoreWeights()
	var configValue sql.NullString
	err := r.db.QueryRow(`
SELECT config_value
FROM system_configs
WHERE LOWER(config_key) = LOWER(?)
ORDER BY updated_at DESC
LIMIT 1`, futuresStrategyScoreWeightsConfigKey).Scan(&configValue)
	if err != nil || !configValue.Valid {
		return defaults
	}
	parsed, ok := parseFuturesStrategyScoreWeights(configValue.String)
	if !ok {
		return defaults
	}
	return parsed
}

func formatCompactPercent(value float64) string {
	percent := roundTo(value*100, 1)
	if math.Abs(percent-math.Round(percent)) < 0.05 {
		return fmt.Sprintf("%.0f", math.Round(percent))
	}
	return fmt.Sprintf("%.1f", percent)
}

func futuresStrategyScoreMethod(weights map[string]float64) string {
	return fmt.Sprintf(
		"futures-v1 (trend%s + structure%s + flow%s + risk%s + news%s + performance%s)",
		formatCompactPercent(weights["trend"]),
		formatCompactPercent(weights["structure"]),
		formatCompactPercent(weights["flow"]),
		formatCompactPercent(weights["risk"]),
		formatCompactPercent(weights["news"]),
		formatCompactPercent(weights["performance"]),
	)
}

func buildFuturesScoreFramework(strategy model.FuturesStrategy, guidance model.FuturesGuidance, stats model.FuturesStrategyPerformanceSummary, relatedNews []model.StockRecommendationRelatedNews, relatedEvents []model.MarketEvent, factorWeights map[string]float64) model.FuturesStrategyScoreFramework {
	weights := normalizeFuturesStrategyScoreWeights(factorWeights)

	trendScore := 68.0
	if strings.EqualFold(strategy.Direction, "LONG") || strings.EqualFold(strategy.Direction, "SHORT") {
		trendScore += 6
	}
	reasonText := strings.ToLower(strings.TrimSpace(strategy.ReasonSummary))
	if strings.Contains(reasonText, "趋势") {
		trendScore += 8
	}
	if strings.Contains(reasonText, "突破") {
		trendScore += 5
	}
	trendScore = clampFloat(trendScore, 45, 96)

	structureScore := 56.0
	if guidance.EntryRange != "" {
		structureScore += 12
	}
	if guidance.TakeProfitRange != "" {
		structureScore += 10
	}
	if guidance.StopLossRange != "" {
		structureScore += 10
	}
	if guidance.InvalidCondition != "" {
		structureScore += 6
	}
	structureScore = clampFloat(structureScore, 40, 95)

	eventScore := clampFloat(58+float64(len(relatedEvents))*5, 45, 90)

	riskScore := 74.0
	switch strings.ToUpper(strings.TrimSpace(strategy.RiskLevel)) {
	case "LOW":
		riskScore = 86
	case "MEDIUM":
		riskScore = 76
	case "HIGH":
		riskScore = 64
	}

	newsAvg := 0.0
	for _, item := range relatedNews {
		newsAvg += clampFloat(item.RelevanceScore, 0, 1)
	}
	if len(relatedNews) > 0 {
		newsAvg = newsAvg / float64(len(relatedNews))
	}
	newsScore := clampFloat(55+newsAvg*35, 45, 92)

	perfScore := clampFloat(50+stats.WinRate*25+stats.ExcessReturn*120, 40, 96)

	factors := []model.FuturesStrategyFactorScore{
		{Key: "trend", Label: "趋势因子", Weight: weights["trend"], Score: roundTo(trendScore, 2)},
		{Key: "structure", Label: "结构因子", Weight: weights["structure"], Score: roundTo(structureScore, 2)},
		{Key: "flow", Label: "资金与事件因子", Weight: weights["flow"], Score: roundTo(eventScore, 2)},
		{Key: "risk", Label: "风险控制因子", Weight: weights["risk"], Score: roundTo(riskScore, 2)},
		{Key: "news", Label: "资讯因子", Weight: weights["news"], Score: roundTo(newsScore, 2)},
		{Key: "performance", Label: "绩效因子", Weight: weights["performance"], Score: roundTo(perfScore, 2)},
	}

	total := 0.0
	weightSum := 0.0
	for index := range factors {
		factors[index].Contribution = roundTo(factors[index].Score*factors[index].Weight, 2)
		total += factors[index].Contribution
		weightSum += factors[index].Weight
	}
	weighted := total
	if weightSum > 0 {
		weighted = total / weightSum
	}
	weighted = roundTo(weighted, 2)
	totalScore := roundTo(weighted+0.4, 2)
	return model.FuturesStrategyScoreFramework{
		Method:        futuresStrategyScoreMethod(weights),
		TotalScore:    totalScore,
		WeightedScore: weighted,
		ScoreGap:      roundTo(totalScore-weighted, 2),
		Factors:       factors,
	}
}
