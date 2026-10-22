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

func (r *MySQLGrowthRepo) AdminListMarketEvents(eventType string, symbol string, page int, pageSize int) ([]model.MarketEvent, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	trimmedType := strings.ToUpper(strings.TrimSpace(eventType))
	trimmedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	if trimmedType != "" {
		filter += " AND event_type = ?"
		args = append(args, trimmedType)
	}
	if trimmedSymbol != "" {
		filter += " AND symbol LIKE ?"
		args = append(args, "%"+trimmedSymbol+"%")
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM market_events"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, event_type, symbol, summary, trigger_rule, source, created_at
FROM market_events` + filter + `
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.MarketEvent, 0)
	for rows.Next() {
		var item model.MarketEvent
		var summary, triggerRule, source sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.EventType, &item.Symbol, &summary, &triggerRule, &source, &createdAt); err != nil {
			return nil, 0, err
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
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateMarketEvent(item model.MarketEvent) (string, error) {
	id := newID("me")
	_, err := r.db.Exec(`
INSERT INTO market_events (id, event_type, symbol, summary, trigger_rule, source, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id,
		strings.ToUpper(strings.TrimSpace(item.EventType)),
		strings.ToUpper(strings.TrimSpace(item.Symbol)),
		strings.TrimSpace(item.Summary),
		strings.TrimSpace(item.TriggerRule),
		strings.TrimSpace(item.Source),
		time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateMarketEvent(id string, item model.MarketEvent) error {
	result, err := r.db.Exec(`
UPDATE market_events
SET event_type = ?, symbol = ?, summary = ?, trigger_rule = ?, source = ?
WHERE id = ?`,
		strings.ToUpper(strings.TrimSpace(item.EventType)),
		strings.ToUpper(strings.TrimSpace(item.Symbol)),
		strings.TrimSpace(item.Summary),
		strings.TrimSpace(item.TriggerRule),
		strings.TrimSpace(item.Source),
		id,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminListStockRecommendations(status string, page int, pageSize int) ([]model.StockRecommendation, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM stock_recommendations"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, source_type, strategy_version, reviewer, publisher, review_note, performance_label
FROM stock_recommendations` + filter + `
ORDER BY created_at DESC
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
		var pos, reason, sourceType, strategyVersion, reviewer, publisher, reviewNote, performanceLabel sql.NullString
		var vf, vt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Symbol,
			&item.Name,
			&item.Score,
			&item.RiskLevel,
			&pos,
			&vf,
			&vt,
			&item.Status,
			&reason,
			&sourceType,
			&strategyVersion,
			&reviewer,
			&publisher,
			&reviewNote,
			&performanceLabel,
		); err != nil {
			return nil, 0, err
		}
		if pos.Valid {
			item.PositionRange = pos.String
		}
		if reason.Valid {
			item.ReasonSummary = reason.String
		}
		if sourceType.Valid {
			item.SourceType = sourceType.String
		}
		if strategyVersion.Valid {
			item.StrategyVersion = strategyVersion.String
		}
		if reviewer.Valid {
			item.Reviewer = reviewer.String
		}
		if publisher.Valid {
			item.Publisher = publisher.String
		}
		if reviewNote.Valid {
			item.ReviewNote = reviewNote.String
		}
		if performanceLabel.Valid {
			item.PerformanceLabel = performanceLabel.String
		}
		item.ValidFrom = vf.Format(time.RFC3339)
		item.ValidTo = vt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateStockRecommendation(item model.StockRecommendation) (string, error) {
	id := newID("sr")
	vf, err := parseFlexibleDateTime(item.ValidFrom)
	if err != nil {
		return "", err
	}
	vt, err := parseFlexibleDateTime(item.ValidTo)
	if err != nil {
		return "", err
	}
	sourceType := strings.ToUpper(strings.TrimSpace(item.SourceType))
	if sourceType == "" {
		sourceType = "MANUAL"
	}
	strategyVersion := strings.TrimSpace(item.StrategyVersion)
	if strategyVersion == "" {
		strategyVersion = "manual-v1"
	}
	reviewer := strings.TrimSpace(item.Reviewer)
	publisher := strings.TrimSpace(item.Publisher)
	reviewNote := strings.TrimSpace(item.ReviewNote)
	performanceLabel := strings.ToUpper(strings.TrimSpace(item.PerformanceLabel))
	if performanceLabel == "" {
		performanceLabel = "PENDING"
	}
	_, err = r.db.Exec(`
INSERT INTO stock_recommendations (id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, source_type, strategy_version, reviewer, publisher, review_note, performance_label, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id,
		item.Symbol,
		item.Name,
		item.Score,
		item.RiskLevel,
		item.PositionRange,
		vf,
		vt,
		item.Status,
		item.ReasonSummary,
		sourceType,
		strategyVersion,
		reviewer,
		publisher,
		reviewNote,
		performanceLabel,
		time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateStockRecommendationStatus(id string, status string) error {
	targetStatus := normalizeStockRecommendationLifecycleStatus(status)
	if targetStatus == "" {
		return errors.New("invalid stock recommendation status")
	}

	var currentStatus string
	if err := r.db.QueryRow("SELECT status FROM stock_recommendations WHERE id = ?", id).Scan(&currentStatus); err != nil {
		return err
	}
	currentStatus = normalizeStockRecommendationLifecycleStatus(currentStatus)
	if currentStatus == "" {
		return errors.New("current stock recommendation status is empty")
	}
	if !canTransitionStockRecommendationStatus(currentStatus, targetStatus) {
		return fmt.Errorf(
			"invalid status transition: %s -> %s (allowed: %s)",
			currentStatus,
			targetStatus,
			strings.Join(allowedStockRecommendationTransitions(currentStatus), ", "),
		)
	}

	_, err := r.db.Exec("UPDATE stock_recommendations SET status = ? WHERE id = ?", targetStatus, id)
	return err
}

func (r *MySQLGrowthRepo) AdminGetQuantTopStocks(limit int, lookbackDays int) ([]model.StockQuantScore, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if lookbackDays <= 0 {
		lookbackDays = 120
	}
	if lookbackDays < 30 {
		lookbackDays = 30
	}
	if lookbackDays > 365 {
		lookbackDays = 365
	}

	sinceDate := time.Now().AddDate(0, 0, -(lookbackDays + 40))
	quotesBySymbol, err := r.loadStockQuotesBySymbol(sinceDate)
	if err != nil {
		return nil, err
	}
	items := make([]model.StockQuantScore, 0, len(quotesBySymbol))
	for symbol, quotes := range quotesBySymbol {
		item, ok := buildStockQuantScore(symbol, quotes)
		if !ok {
			continue
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return []model.StockQuantScore{}, nil
	}
	dailyBasicMap, err := r.loadLatestStockDailyBasics(sinceDate)
	if err != nil {
		return nil, err
	}
	moneyflowMap, err := r.loadLatestStockMoneyflows(sinceDate)
	if err != nil {
		return nil, err
	}
	newsSignalMap, err := r.loadStockNewsSignals(time.Now().AddDate(0, 0, -14))
	if err != nil {
		return nil, err
	}
	for index := range items {
		if basic, ok := dailyBasicMap[items[index].Symbol]; ok {
			items[index].PeTTM = basic.PeTTM
			items[index].PB = basic.PB
			items[index].TurnoverRate = basic.TurnoverRate
		}
		if flow, ok := moneyflowMap[items[index].Symbol]; ok {
			items[index].NetMFAmount = flow.NetMFAmount
		}
		if signal, ok := newsSignalMap[items[index].Symbol]; ok {
			items[index].NewsHeat = signal.Heat
			items[index].PositiveNewsRate = signal.PositiveRate
		}
	}

	momentum5Range := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.Momentum5 })
	momentum20Range := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.Momentum20 })
	volatilityRange := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.Volatility20 })
	volumeRatioRange := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.VolumeRatio })
	drawdownRange := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.Drawdown20 })
	trendRange := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.TrendStrength })
	peRange := buildMetricRangeFromValues(extractPositiveValues(items, func(item model.StockQuantScore) float64 { return item.PeTTM }))
	pbRange := buildMetricRangeFromValues(extractPositiveValues(items, func(item model.StockQuantScore) float64 { return item.PB }))
	turnoverRateRange := buildMetricRangeFromValues(extractPositiveValues(items, func(item model.StockQuantScore) float64 { return item.TurnoverRate }))
	netMoneyflowRange := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.NetMFAmount })
	newsHeatRange := buildMetricRangeFromValues(extractPositiveValues(items, func(item model.StockQuantScore) float64 { return float64(item.NewsHeat) }))

	for index := range items {
		m5Norm := normalizeMetric(items[index].Momentum5, momentum5Range, false)
		m20Norm := normalizeMetric(items[index].Momentum20, momentum20Range, false)
		volNorm := normalizeMetric(items[index].Volatility20, volatilityRange, true)
		volumeNorm := normalizeMetric(items[index].VolumeRatio, volumeRatioRange, false)
		drawdownNorm := normalizeMetric(items[index].Drawdown20, drawdownRange, true)
		trendNorm := normalizeMetric(items[index].TrendStrength, trendRange, false)
		trendScore := (m5Norm*0.25 + m20Norm*0.30 + volNorm*0.15 + drawdownNorm*0.15 + trendNorm*0.15) * 100

		netMoneyflowNorm := 0.5
		if math.Abs(items[index].NetMFAmount) > 0.001 {
			netMoneyflowNorm = normalizeMetric(items[index].NetMFAmount, netMoneyflowRange, false)
		}
		flowScore := (volumeNorm*0.42 + netMoneyflowNorm*0.58) * 100

		peNorm := 0.48
		if items[index].PeTTM > 0 {
			peNorm = normalizeMetric(items[index].PeTTM, peRange, true)
		}
		pbNorm := 0.5
		if items[index].PB > 0 {
			pbNorm = normalizeMetric(items[index].PB, pbRange, true)
		}
		turnoverNorm := 0.45
		if items[index].TurnoverRate > 0 {
			turnoverNorm = normalizeMetric(items[index].TurnoverRate, turnoverRateRange, false)
		}
		valueScore := (peNorm*0.45 + pbNorm*0.25 + turnoverNorm*0.30) * 100

		heatNorm := 0.45
		if items[index].NewsHeat > 0 {
			heatNorm = normalizeMetric(float64(items[index].NewsHeat), newsHeatRange, false)
		}
		positiveRate := items[index].PositiveNewsRate
		if positiveRate <= 0 {
			positiveRate = 0.5
		}
		newsScore := (heatNorm*0.45 + clampFloat(positiveRate, 0, 1)*0.55) * 100

		score := trendScore*0.45 + flowScore*0.25 + valueScore*0.20 + newsScore*0.10
		score -= quantRiskPenalty(items[index])
		score = clampFloat(score, 0, 100)
		items[index].TrendScore = roundTo(trendScore, 2)
		items[index].FlowScore = roundTo(flowScore, 2)
		items[index].ValueScore = roundTo(valueScore, 2)
		items[index].NewsScore = roundTo(newsScore, 2)
		items[index].Score = roundTo(score, 2)
		items[index].Momentum5 = roundTo(items[index].Momentum5, 2)
		items[index].Momentum20 = roundTo(items[index].Momentum20, 2)
		items[index].Volatility20 = roundTo(items[index].Volatility20, 2)
		items[index].VolumeRatio = roundTo(items[index].VolumeRatio, 2)
		items[index].Drawdown20 = roundTo(items[index].Drawdown20, 2)
		items[index].TrendStrength = roundTo(items[index].TrendStrength, 2)
		items[index].NetMFAmount = roundTo(items[index].NetMFAmount, 2)
		items[index].PeTTM = roundTo(items[index].PeTTM, 2)
		items[index].PB = roundTo(items[index].PB, 2)
		items[index].TurnoverRate = roundTo(items[index].TurnoverRate, 2)
		items[index].PositiveNewsRate = roundTo(items[index].PositiveNewsRate, 4)
		items[index].ClosePrice = roundTo(items[index].ClosePrice, 3)
	}
	filteredItems := make([]model.StockQuantScore, 0, len(items))
	for _, item := range items {
		if !passesQuantRiskGate(item) {
			continue
		}
		filteredItems = append(filteredItems, item)
	}
	if len(filteredItems) == 0 {
		filteredItems = items
	}
	items = filteredItems

	sort.Slice(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			if items[i].Momentum20 == items[j].Momentum20 {
				return items[i].Symbol < items[j].Symbol
			}
			return items[i].Momentum20 > items[j].Momentum20
		}
		return items[i].Score > items[j].Score
	})

	if len(items) > limit {
		items = selectDiversifiedQuantTop(items, limit)
	}

	symbols := make([]string, 0, len(items))
	for _, item := range items {
		symbols = append(symbols, item.Symbol)
	}
	nameMap, err := r.lookupStockNames(symbols)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].Rank = index + 1
		if name, ok := nameMap[items[index].Symbol]; ok && strings.TrimSpace(name) != "" {
			items[index].Name = name
		}
		if strings.TrimSpace(items[index].Name) == "" {
			items[index].Name = items[index].Symbol
		}
		items[index].Reasons = buildQuantReasons(items[index])
		items[index].ReasonSummary = buildQuantReasonSummary(items[index])
	}
	if err := r.persistStockQuantSnapshots(items); err != nil && !isTableNotFoundError(err) {
		return nil, err
	}

	return items, nil
}

func (r *MySQLGrowthRepo) AdminGetQuantEvaluation(windowDays int, topN int) (model.StockQuantEvaluationSummary, []model.StockQuantEvaluationPoint, []model.StockQuantRiskPerformance, []model.StockQuantRotationPoint, error) {
	if windowDays <= 0 {
		windowDays = 60
	}
	if windowDays < 20 {
		windowDays = 20
	}
	if windowDays > 365 {
		windowDays = 365
	}
	if topN <= 0 {
		topN = 10
	}
	if topN > 30 {
		topN = 30
	}
	sinceDate := time.Now().AddDate(0, 0, -windowDays)
	rankRows, err := r.db.Query(`
SELECT trade_date, symbol, rank_no
FROM stock_rank_daily
WHERE trade_date >= ? AND rank_no <= ?
ORDER BY trade_date ASC, rank_no ASC`, sinceDate.Format("2006-01-02"), topN)
	if err != nil {
		if isTableNotFoundError(err) {
			return model.StockQuantEvaluationSummary{
				WindowDays:  windowDays,
				TopN:        topN,
				GeneratedAt: time.Now().Format(time.RFC3339),
			}, []model.StockQuantEvaluationPoint{}, []model.StockQuantRiskPerformance{}, []model.StockQuantRotationPoint{}, nil
		}
		return model.StockQuantEvaluationSummary{}, nil, nil, nil, err
	}
	defer rankRows.Close()

	type rankItem struct {
		TradeDate time.Time
		Symbol    string
		RankNo    int
	}
	ranks := make([]rankItem, 0, windowDays*topN)
	for rankRows.Next() {
		var (
			item   rankItem
			symbol string
		)
		if err := rankRows.Scan(&item.TradeDate, &symbol, &item.RankNo); err != nil {
			return model.StockQuantEvaluationSummary{}, nil, nil, nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if item.Symbol == "" {
			continue
		}
		ranks = append(ranks, item)
	}
	if len(ranks) == 0 {
		return model.StockQuantEvaluationSummary{
			WindowDays:  windowDays,
			TopN:        topN,
			GeneratedAt: time.Now().Format(time.RFC3339),
		}, []model.StockQuantEvaluationPoint{}, []model.StockQuantRiskPerformance{}, []model.StockQuantRotationPoint{}, nil
	}

	quoteSince := sinceDate.AddDate(0, 0, -20)
	quotesBySymbol, err := r.loadStockQuotesBySymbol(quoteSince)
	if err != nil {
		return model.StockQuantEvaluationSummary{}, nil, nil, nil, err
	}
	benchmarkSeries, err := r.loadBenchmarkQuoteSeries(quoteSince)
	if err != nil {
		return model.StockQuantEvaluationSummary{}, nil, nil, nil, err
	}
	riskLevelMap, err := r.loadQuantRiskLevelByDateSymbol(sinceDate, topN)
	if err != nil {
		return model.StockQuantEvaluationSummary{}, nil, nil, nil, err
	}

	type dayAggregate struct {
		Sum5           float64
		Cnt5           int
		Hit5           int
		Sum10          float64
		Cnt10          int
		Hit10          int
		Benchmark5Sum  float64
		Benchmark5Cnt  int
		Benchmark10Sum float64
		Benchmark10Cnt int
	}
	byDay := make(map[string]*dayAggregate)
	riskAggMap := make(map[string]*quantRiskAggregate)
	dayTopSymbols := make(map[string][]string)
	dayTopSeen := make(map[string]map[string]struct{})
	for _, item := range ranks {
		symbolQuotes, ok := quotesBySymbol[item.Symbol]
		if !ok || len(symbolQuotes) == 0 {
			continue
		}
		dayKey := item.TradeDate.Format("2006-01-02")
		if _, ok := dayTopSeen[dayKey]; !ok {
			dayTopSeen[dayKey] = make(map[string]struct{}, topN)
		}
		if _, exists := dayTopSeen[dayKey][item.Symbol]; !exists {
			dayTopSeen[dayKey][item.Symbol] = struct{}{}
			dayTopSymbols[dayKey] = append(dayTopSymbols[dayKey], item.Symbol)
		}

		agg := byDay[dayKey]
		if agg == nil {
			agg = &dayAggregate{}
			byDay[dayKey] = agg
		}
		riskLevel := resolveQuantRiskLevel(riskLevelMap[dayKey+"|"+item.Symbol])
		riskAgg := riskAggMap[riskLevel]
		if riskAgg == nil {
			riskAgg = &quantRiskAggregate{}
			riskAggMap[riskLevel] = riskAgg
		}
		if ret5, ok := calcForwardReturn(symbolQuotes, item.TradeDate, 5); ok {
			agg.Sum5 += ret5
			agg.Cnt5++
			if ret5 > 0 {
				agg.Hit5++
			}
			riskAgg.Sum5 += ret5
			riskAgg.Cnt5++
			if ret5 > 0 {
				riskAgg.Hit5++
			}
		}
		if ret10, ok := calcForwardReturn(symbolQuotes, item.TradeDate, 10); ok {
			agg.Sum10 += ret10
			agg.Cnt10++
			if ret10 > 0 {
				agg.Hit10++
			}
			riskAgg.Sum10 += ret10
			riskAgg.Cnt10++
			if ret10 > 0 {
				riskAgg.Hit10++
			}
		}
		if benchmark5, ok := calcForwardReturn(benchmarkSeries, item.TradeDate, 5); ok {
			agg.Benchmark5Sum += benchmark5
			agg.Benchmark5Cnt++
		}
		if benchmark10, ok := calcForwardReturn(benchmarkSeries, item.TradeDate, 10); ok {
			agg.Benchmark10Sum += benchmark10
			agg.Benchmark10Cnt++
		}
	}

	tradeDates := make([]string, 0, len(byDay))
	for day := range byDay {
		tradeDates = append(tradeDates, day)
	}
	sort.Strings(tradeDates)
	points := make([]model.StockQuantEvaluationPoint, 0, len(tradeDates))
	summary := model.StockQuantEvaluationSummary{
		WindowDays:  windowDays,
		TopN:        topN,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
	equityReturns5 := make([]float64, 0, len(tradeDates))
	equityReturns10 := make([]float64, 0, len(tradeDates))
	benchmarkReturns5 := make([]float64, 0, len(tradeDates))
	benchmarkReturns10 := make([]float64, 0, len(tradeDates))
	for _, tradeDate := range tradeDates {
		agg := byDay[tradeDate]
		point := model.StockQuantEvaluationPoint{TradeDate: tradeDate}
		if agg.Cnt5 > 0 {
			point.AvgReturn5 = roundTo(agg.Sum5/float64(agg.Cnt5), 4)
			point.HitRate5 = roundTo(float64(agg.Hit5)/float64(agg.Cnt5), 4)
			if agg.Benchmark5Cnt > 0 {
				point.BenchmarkReturn = roundTo(agg.Benchmark5Sum/float64(agg.Benchmark5Cnt), 4)
			}
			point.SampleCount = agg.Cnt5
			equityReturns5 = append(equityReturns5, point.AvgReturn5)
			benchmarkReturns5 = append(benchmarkReturns5, point.BenchmarkReturn)
			summary.AvgReturn5 += point.AvgReturn5
			summary.HitRate5 += point.HitRate5
			summary.SampleCount += agg.Cnt5
		}
		if agg.Cnt10 > 0 {
			point.AvgReturn10 = roundTo(agg.Sum10/float64(agg.Cnt10), 4)
			point.HitRate10 = roundTo(float64(agg.Hit10)/float64(agg.Cnt10), 4)
			equityReturns10 = append(equityReturns10, point.AvgReturn10)
			if agg.Benchmark10Cnt > 0 {
				point.BenchmarkReturn10 = roundTo(agg.Benchmark10Sum/float64(agg.Benchmark10Cnt), 4)
				benchmarkReturns10 = append(benchmarkReturns10, point.BenchmarkReturn10)
			}
			summary.AvgReturn10 += point.AvgReturn10
			summary.HitRate10 += point.HitRate10
		}
		points = append(points, point)
	}
	summary.SampleDays = len(points)
	if len(points) > 0 {
		if len(equityReturns5) > 0 {
			summary.AvgReturn5 = roundTo(summary.AvgReturn5/float64(len(equityReturns5)), 4)
			summary.HitRate5 = roundTo(summary.HitRate5/float64(len(equityReturns5)), 4)
		}
		if len(equityReturns10) > 0 {
			summary.AvgReturn10 = roundTo(summary.AvgReturn10/float64(len(equityReturns10)), 4)
			summary.HitRate10 = roundTo(summary.HitRate10/float64(len(equityReturns10)), 4)
		}
		summary.MaxDrawdown5 = roundTo(calcMaxDrawdownFromReturns(equityReturns5), 4)
		summary.MaxDrawdown10 = roundTo(calcMaxDrawdownFromReturns(equityReturns10), 4)
		summary.BenchmarkAvgReturn5 = roundTo(avgFloat(benchmarkReturns5), 4)
		summary.BenchmarkAvgReturn10 = roundTo(avgFloat(benchmarkReturns10), 4)
	}
	cum5 := 1.0
	cumBench5 := 1.0
	cum10 := 1.0
	cumBench10 := 1.0
	for index := range points {
		cum5 *= 1 + points[index].AvgReturn5
		cumBench5 *= 1 + points[index].BenchmarkReturn
		points[index].CumulativeReturn5 = roundTo(cum5-1, 4)
		points[index].CumulativeBenchmark5 = roundTo(cumBench5-1, 4)
		points[index].CumulativeExcess5 = roundTo(points[index].CumulativeReturn5-points[index].CumulativeBenchmark5, 4)
		cum10 *= 1 + points[index].AvgReturn10
		cumBench10 *= 1 + points[index].BenchmarkReturn10
		points[index].CumulativeReturn10 = roundTo(cum10-1, 4)
		points[index].CumulativeBenchmark10 = roundTo(cumBench10-1, 4)
		points[index].CumulativeExcess10 = roundTo(points[index].CumulativeReturn10-points[index].CumulativeBenchmark10, 4)
	}
	riskItems := buildQuantRiskPerformanceItems(riskAggMap)
	rotationItems := buildQuantRotationItems(tradeDates, dayTopSymbols)
	return summary, points, riskItems, rotationItems, nil
}

func (r *MySQLGrowthRepo) AdminGenerateDailyStockRecommendations(tradeDate string) (model.AdminDailyStockRecommendationGenerationResult, error) {
	if r.strategyEngine != nil {
		return r.generateDailyStockRecommendationsViaStrategyEngine(tradeDate)
	}
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	start, err := time.ParseInLocation("2006-01-02", tradeDate, time.Local)
	if err != nil {
		return model.AdminDailyStockRecommendationGenerationResult{}, err
	}
	end := start.Add(24 * time.Hour)
	quantItems, quantErr := r.AdminGetQuantTopStocks(10, 180)
	candidates := buildDailyStockCandidatesFromQuant(quantItems, start, end)
	if quantErr != nil || len(candidates) == 0 {
		candidates = buildDefaultDailyStockCandidates(start, end)
	}

	count := 0
	for _, candidate := range candidates {
		id := newID("sr")
		_, err := r.db.Exec(`
INSERT INTO stock_recommendations (id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, source_type, strategy_version, reviewer, publisher, review_note, performance_label, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id,
			candidate.Symbol,
			candidate.Name,
			candidate.Score,
			candidate.RiskLevel,
			candidate.PositionRange,
			candidate.ValidFrom,
			candidate.ValidTo,
			candidate.Status,
			candidate.ReasonSummary,
			"SYSTEM",
			"daily-v1",
			"",
			"system",
			"由每日量化流水线自动生成，待运营跟踪。",
			"ESTIMATED",
			time.Now(),
		)
		if err != nil {
			return model.AdminDailyStockRecommendationGenerationResult{}, err
		}
		_, _ = r.db.Exec(`
INSERT INTO stock_reco_details (reco_id, tech_score, fund_score, sentiment_score, money_flow_score, take_profit, stop_loss, risk_note)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE tech_score=VALUES(tech_score), fund_score=VALUES(fund_score), sentiment_score=VALUES(sentiment_score), money_flow_score=VALUES(money_flow_score), take_profit=VALUES(take_profit), stop_loss=VALUES(stop_loss), risk_note=VALUES(risk_note)`,
			id,
			candidate.TechScore,
			candidate.FundScore,
			candidate.SentimentScore,
			candidate.MoneyFlowScore,
			candidate.TakeProfit,
			candidate.StopLoss,
			candidate.RiskNote,
		)
		count++
	}
	return model.AdminDailyStockRecommendationGenerationResult{
		Count:          count,
		GenerationMode: "FALLBACK",
		ArchiveEnabled: false,
	}, nil
}

func (r *MySQLGrowthRepo) AdminGenerateDailyFuturesStrategies(tradeDate string) (model.AdminDailyFuturesStrategyGenerationResult, error) {
	if r.strategyEngine != nil {
		return r.generateDailyFuturesStrategiesViaStrategyEngine(tradeDate)
	}
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	start, err := time.ParseInLocation("2006-01-02", tradeDate, time.Local)
	if err != nil {
		return model.AdminDailyFuturesStrategyGenerationResult{}, err
	}
	end := start.Add(24 * time.Hour)
	samples := []model.FuturesStrategy{
		{Contract: "IF2603", Name: "股指趋势", Direction: "LONG", RiskLevel: "MEDIUM", PositionRange: "20%-30%", Status: "PUBLISHED", ReasonSummary: "趋势与量价结构一致"},
		{Contract: "IH2603", Name: "蓝筹回归", Direction: "LONG", RiskLevel: "LOW", PositionRange: "15%-25%", Status: "PUBLISHED", ReasonSummary: "估值回归驱动"},
		{Contract: "IC2603", Name: "中证成长", Direction: "SHORT", RiskLevel: "HIGH", PositionRange: "10%-20%", Status: "PUBLISHED", ReasonSummary: "风格切换压力"},
	}
	count := 0
	for _, s := range samples {
		id := newID("fs")
		_, err := r.db.Exec(`
INSERT INTO futures_strategies (id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, s.Contract, s.Name, s.Direction, s.RiskLevel, s.PositionRange, start, end, s.Status, s.ReasonSummary,
		)
		if err != nil {
			return model.AdminDailyFuturesStrategyGenerationResult{}, err
		}
		count++
	}
	return model.AdminDailyFuturesStrategyGenerationResult{
		Count:          count,
		GenerationMode: "FALLBACK",
		ArchiveEnabled: false,
	}, nil
}

func (r *MySQLGrowthRepo) AdminListFuturesStrategies(status string, contract string, page int, pageSize int) ([]model.FuturesStrategy, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	contract = strings.TrimSpace(contract)
	status = strings.TrimSpace(status)
	if contract != "" {
		filter += " AND contract = ?"
		args = append(args, contract)
	}
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
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

func (r *MySQLGrowthRepo) AdminCreateFuturesStrategy(item model.FuturesStrategy) (string, error) {
	id := newID("fs")
	vf, err := parseFlexibleDateTime(item.ValidFrom)
	if err != nil {
		return "", err
	}
	vt, err := parseFlexibleDateTime(item.ValidTo)
	if err != nil {
		return "", err
	}
	_, err = r.db.Exec(`
INSERT INTO futures_strategies (id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, item.Contract, item.Name, item.Direction, item.RiskLevel, item.PositionRange, vf, vt, item.Status, item.ReasonSummary,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateFuturesStrategyStatus(id string, status string) error {
	_, err := r.db.Exec("UPDATE futures_strategies SET status = ? WHERE id = ?", status, id)
	return err
}
