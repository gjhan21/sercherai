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

func calcForwardReturn(quotes []stockQuoteCandle, tradeDate time.Time, horizon int) (float64, bool) {
	if horizon <= 0 || len(quotes) == 0 {
		return 0, false
	}
	indexByDate := make(map[string]int, len(quotes))
	for idx, quote := range quotes {
		indexByDate[quote.TradeDate.Format("2006-01-02")] = idx
	}
	key := tradeDate.Format("2006-01-02")
	startIndex, ok := indexByDate[key]
	if !ok {
		return 0, false
	}
	targetIndex := startIndex + horizon
	if targetIndex >= len(quotes) {
		return 0, false
	}
	startClose := quotes[startIndex].ClosePrice
	targetClose := quotes[targetIndex].ClosePrice
	if startClose <= 0 || targetClose <= 0 {
		return 0, false
	}
	return targetClose/startClose - 1, true
}

func buildStockQuantScore(symbol string, quotes []stockQuoteCandle) (model.StockQuantScore, bool) {
	if len(quotes) < 25 {
		return model.StockQuantScore{}, false
	}
	lastIndex := len(quotes) - 1
	index5 := lastIndex - 5
	index20 := lastIndex - 20
	if index5 < 0 || index20 < 0 {
		return model.StockQuantScore{}, false
	}
	latest := quotes[lastIndex]
	base5 := quotes[index5].ClosePrice
	base20 := quotes[index20].ClosePrice
	if base5 <= 0 || base20 <= 0 || latest.ClosePrice <= 0 {
		return model.StockQuantScore{}, false
	}

	momentum5 := (latest.ClosePrice/base5 - 1) * 100
	momentum20 := (latest.ClosePrice/base20 - 1) * 100

	returns := make([]float64, 0, 20)
	for idx := index20 + 1; idx <= lastIndex; idx++ {
		prev := quotes[idx-1].ClosePrice
		curr := quotes[idx].ClosePrice
		if prev <= 0 || curr <= 0 {
			continue
		}
		returns = append(returns, curr/prev-1)
	}
	if len(returns) < 10 {
		return model.StockQuantScore{}, false
	}
	volatility20 := calculateStdDev(returns) * 100

	volumeSum := 0.0
	volumeCount := 0
	startVolume := maxInt(0, lastIndex-20)
	for idx := startVolume; idx < lastIndex; idx++ {
		if quotes[idx].Volume <= 0 {
			continue
		}
		volumeSum += quotes[idx].Volume
		volumeCount++
	}
	avgVolume := volumeSum
	if volumeCount > 0 {
		avgVolume = volumeSum / float64(volumeCount)
	}
	volumeRatio := 1.0
	if avgVolume > 0 && latest.Volume > 0 {
		volumeRatio = latest.Volume / avgVolume
	}

	startDrawdown := maxInt(0, len(quotes)-20)
	peak := quotes[startDrawdown].ClosePrice
	maxDrawdown := 0.0
	for idx := startDrawdown; idx <= lastIndex; idx++ {
		closePrice := quotes[idx].ClosePrice
		if closePrice > peak {
			peak = closePrice
		}
		if peak > 0 {
			drawdown := (peak - closePrice) / peak * 100
			if drawdown > maxDrawdown {
				maxDrawdown = drawdown
			}
		}
	}

	ma5 := 0.0
	for idx := lastIndex - 4; idx <= lastIndex; idx++ {
		ma5 += quotes[idx].ClosePrice
	}
	ma5 /= 5
	ma20 := 0.0
	for idx := lastIndex - 19; idx <= lastIndex; idx++ {
		ma20 += quotes[idx].ClosePrice
	}
	ma20 /= 20
	trendStrength := 0.0
	if ma20 > 0 {
		trendStrength = (ma5/ma20 - 1) * 100
	}

	// --- Intraday / T+1 short-cycle feature computation ---

	// Short-cycle momentum
	momentum1 := 0.0
	if latest.PrevClosePrice > 0 {
		momentum1 = (latest.ClosePrice/latest.PrevClosePrice - 1) * 100
	}
	momentum2 := 0.0
	if lastIndex >= 2 && quotes[lastIndex-2].ClosePrice > 0 {
		momentum2 = (latest.ClosePrice/quotes[lastIndex-2].ClosePrice - 1) * 100
	}
	momentum3 := 0.0
	if lastIndex >= 3 && quotes[lastIndex-3].ClosePrice > 0 {
		momentum3 = (latest.ClosePrice/quotes[lastIndex-3].ClosePrice - 1) * 100
	}

	// Consecutive down days
	consecutiveDownDays := 0
	for idx := lastIndex; idx >= 0; idx-- {
		if quotes[idx].PrevClosePrice > 0 && quotes[idx].ClosePrice < quotes[idx].PrevClosePrice {
			consecutiveDownDays++
		} else {
			break
		}
	}

	// Candle body as percentage of close
	candleBodyPct := math.Abs(latest.ClosePrice-latest.OpenPrice) / math.Max(latest.ClosePrice, 0.01)
	isDoji := candleBodyPct < 0.003
	isBullish := latest.ClosePrice > latest.OpenPrice

	// Shadow percentages
	hlRange := math.Max(latest.HighPrice-latest.LowPrice, 0.01)
	lowerShadowPct := (math.Min(latest.OpenPrice, latest.ClosePrice) - latest.LowPrice) / hlRange
	upperShadowPct := (latest.HighPrice - math.Max(latest.OpenPrice, latest.ClosePrice)) / hlRange

	// Bullish engulfing: today bullish, yesterday bearish, today's close > yesterday's open, today's open < yesterday's close
	isEngulfingBullish := false
	if lastIndex >= 1 {
		prev := quotes[lastIndex-1]
		isEngulfingBullish = latest.ClosePrice > latest.OpenPrice && // today bullish
			prev.ClosePrice < prev.OpenPrice && // yesterday bearish
			latest.OpenPrice < prev.ClosePrice && // open below yesterday close
			latest.ClosePrice > prev.OpenPrice // close above yesterday open
	}

	// MA deviations
	ma10 := 0.0
	if lastIndex >= 9 {
		for idx := lastIndex - 9; idx <= lastIndex; idx++ {
			ma10 += quotes[idx].ClosePrice
		}
		ma10 /= 10
	}
	ma60 := 0.0
	if lastIndex >= 59 {
		for idx := lastIndex - 59; idx <= lastIndex; idx++ {
			ma60 += quotes[idx].ClosePrice
		}
		ma60 /= 60
	}
	deviationMA5 := 0.0
	if ma5 > 0 {
		deviationMA5 = (latest.ClosePrice/ma5 - 1) * 100
	}
	deviationMA10 := 0.0
	if ma10 > 0 {
		deviationMA10 = (latest.ClosePrice/ma10 - 1) * 100
	}
	deviationMA20 := 0.0
	if ma20 > 0 {
		deviationMA20 = (latest.ClosePrice/ma20 - 1) * 100
	}
	deviationMA60 := 0.0
	if ma60 > 0 {
		deviationMA60 = (latest.ClosePrice/ma60 - 1) * 100
	}

	// Volume rank in last 20 days (1 = lowest volume in 20)
	volume20dMinRank := computeVolume20dRank(quotes, lastIndex)

	// Consecutive days of declining volume
	volumeContractionDays := computeVolumeContractionDays(quotes, lastIndex)

	return model.StockQuantScore{
		Symbol:        symbol,
		TradeDate:     latest.TradeDate.Format("2006-01-02"),
		ClosePrice:    latest.ClosePrice,
		Momentum5:     momentum5,
		Momentum20:    momentum20,
		Volatility20:  volatility20,
		VolumeRatio:   volumeRatio,
		Drawdown20:    maxDrawdown,
		TrendStrength: trendStrength,
		RiskLevel:     classifyQuantRisk(volatility20, maxDrawdown),

		// Intraday / T+1 short-cycle features
		Momentum1:             momentum1,
		Momentum2:             momentum2,
		Momentum3:             momentum3,
		ConsecutiveDownDays:   consecutiveDownDays,
		CandleBodyPct:         candleBodyPct,
		LowerShadowPct:        lowerShadowPct,
		UpperShadowPct:        upperShadowPct,
		IsBullish:             isBullish,
		IsDoji:                isDoji,
		IsEngulfingBullish:    isEngulfingBullish,
		DeviationMA5:          deviationMA5,
		DeviationMA10:         deviationMA10,
		DeviationMA20:         deviationMA20,
		DeviationMA60:         deviationMA60,
		Volume20dMinRank:      volume20dMinRank,
		VolumeContractionDays: volumeContractionDays,
		// Limit-up / top-list fields are populated from DB data, not computed here
	}, true
}

// computeVolume20dRank returns the rank of the latest candle's volume within the last 20 candles.
// 1 = lowest volume, higher numbers = higher volume.
func computeVolume20dRank(quotes []stockQuoteCandle, lastIndex int) int {
	if len(quotes) < 2 {
		return 0
	}
	start := maxInt(0, lastIndex-19)
	latestVolume := quotes[lastIndex].Volume
	rank := 1
	for idx := start; idx < lastIndex; idx++ {
		if quotes[idx].Volume > 0 && quotes[idx].Volume < latestVolume {
			rank++
		}
	}
	return rank
}

// computeVolumeContractionDays counts consecutive days (backwards from lastIndex)
// where volume has been below its own 20-day average.
func computeVolumeContractionDays(quotes []stockQuoteCandle, lastIndex int) int {
	days := 0
	for idx := lastIndex; idx >= 0; idx-- {
		startAvg := maxInt(0, idx-19)
		sum := 0.0
		count := 0
		for j := startAvg; j < idx; j++ {
			if quotes[j].Volume > 0 {
				sum += quotes[j].Volume
				count++
			}
		}
		if count == 0 || quotes[idx].Volume <= 0 {
			break
		}
		avg := sum / float64(count)
		if quotes[idx].Volume < avg {
			days++
		} else {
			break
		}
	}
	return days
}

func classifyQuantRisk(volatility20 float64, drawdown20 float64) string {
	if volatility20 >= 4.0 || drawdown20 >= 12 {
		return "HIGH"
	}
	if volatility20 >= 2.5 || drawdown20 >= 7 {
		return "MEDIUM"
	}
	return "LOW"
}

func quantRiskPenalty(item model.StockQuantScore) float64 {
	penalty := 0.0
	if item.Volatility20 >= 5.0 {
		penalty += 4.0
	}
	if item.Drawdown20 >= 16 {
		penalty += 5.0
	}
	if item.NetMFAmount < 0 {
		penalty += 2.5
	}
	if item.PeTTM > 0 && item.PeTTM >= 60 {
		penalty += 2.0
	}
	if item.NewsHeat >= 3 && item.PositiveNewsRate > 0 && item.PositiveNewsRate < 0.3 {
		penalty += 2.0
	}
	return penalty
}

func passesQuantRiskGate(item model.StockQuantScore) bool {
	if item.ClosePrice <= 0 {
		return false
	}
	if item.Volatility20 >= 9.0 {
		return false
	}
	if item.Drawdown20 >= 25 {
		return false
	}
	if item.VolumeRatio <= 0.15 {
		return false
	}
	if item.Momentum20 <= -12 {
		return false
	}
	if item.PeTTM > 0 && item.PeTTM >= 200 {
		return false
	}
	if item.PB > 0 && item.PB >= 30 {
		return false
	}
	if item.NewsHeat >= 5 && item.PositiveNewsRate > 0 && item.PositiveNewsRate < 0.2 {
		return false
	}
	return true
}

func quantBucketBySymbol(symbol string) string {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if strings.HasPrefix(normalized, "68") && strings.HasSuffix(normalized, ".SH") {
		return "KCB"
	}
	if strings.HasPrefix(normalized, "30") && strings.HasSuffix(normalized, ".SZ") {
		return "CYB"
	}
	if strings.HasPrefix(normalized, "60") && strings.HasSuffix(normalized, ".SH") {
		return "MAIN_SH"
	}
	if (strings.HasPrefix(normalized, "00") || strings.HasPrefix(normalized, "002")) && strings.HasSuffix(normalized, ".SZ") {
		return "MAIN_SZ"
	}
	if strings.HasSuffix(normalized, ".BJ") {
		return "BJ"
	}
	if strings.HasSuffix(normalized, ".SH") {
		return "SH"
	}
	if strings.HasSuffix(normalized, ".SZ") {
		return "SZ"
	}
	return "OTHER"
}

func selectDiversifiedQuantTop(items []model.StockQuantScore, limit int) []model.StockQuantScore {
	if len(items) <= limit {
		return items
	}
	capByBucket := maxInt(1, int(math.Ceil(float64(limit)*0.4)))
	selected := make([]model.StockQuantScore, 0, limit)
	overflow := make([]model.StockQuantScore, 0, len(items))
	bucketCount := make(map[string]int)
	for _, item := range items {
		bucket := quantBucketBySymbol(item.Symbol)
		if bucketCount[bucket] < capByBucket {
			selected = append(selected, item)
			bucketCount[bucket]++
		} else {
			overflow = append(overflow, item)
		}
		if len(selected) >= limit {
			return selected
		}
	}
	for _, item := range overflow {
		if len(selected) >= limit {
			break
		}
		selected = append(selected, item)
	}
	if len(selected) > limit {
		selected = selected[:limit]
	}
	return selected
}

func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := 0.0
	for _, item := range values {
		mean += item
	}
	mean /= float64(len(values))
	sumSq := 0.0
	for _, item := range values {
		diff := item - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(values)))
}

func avgFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func calcMaxDrawdownFromReturns(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	equity := 1.0
	peak := 1.0
	maxDrawdown := 0.0
	for _, dayReturn := range returns {
		equity *= 1 + dayReturn
		if equity > peak {
			peak = equity
		}
		if peak <= 0 {
			continue
		}
		drawdown := (peak - equity) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	return maxDrawdown
}

func buildMetricRange(items []model.StockQuantScore, selector func(model.StockQuantScore) float64) metricRange {
	if len(items) == 0 {
		return metricRange{}
	}
	minVal := selector(items[0])
	maxVal := minVal
	for index := 1; index < len(items); index++ {
		value := selector(items[index])
		if value < minVal {
			minVal = value
		}
		if value > maxVal {
			maxVal = value
		}
	}
	return metricRange{Min: minVal, Max: maxVal}
}

func buildMetricRangeFromValues(values []float64) metricRange {
	if len(values) == 0 {
		return metricRange{}
	}
	minVal := values[0]
	maxVal := values[0]
	for index := 1; index < len(values); index++ {
		if values[index] < minVal {
			minVal = values[index]
		}
		if values[index] > maxVal {
			maxVal = values[index]
		}
	}
	return metricRange{Min: minVal, Max: maxVal}
}

func extractPositiveValues(items []model.StockQuantScore, selector func(model.StockQuantScore) float64) []float64 {
	result := make([]float64, 0, len(items))
	for _, item := range items {
		value := selector(item)
		if value > 0 {
			result = append(result, value)
		}
	}
	return result
}

func normalizeMetric(value float64, valueRange metricRange, invert bool) float64 {
	span := valueRange.Max - valueRange.Min
	if math.Abs(span) < 1e-9 {
		return 0.5
	}
	normalized := (value - valueRange.Min) / span
	if invert {
		normalized = 1 - normalized
	}
	return clampFloat(normalized, 0, 1)
}

func clampFloat(value float64, minValue float64, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func roundTo(value float64, precision int) float64 {
	if precision < 0 {
		precision = 0
	}
	factor := math.Pow(10, float64(precision))
	return math.Round(value*factor) / factor
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func sqlNullFloat(value sql.NullFloat64) float64 {
	if !value.Valid {
		return 0
	}
	return value.Float64
}

func isTableNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "error 1146") ||
		(strings.Contains(message, "doesn't exist") && strings.Contains(message, "table"))
}

func (r *MySQLGrowthRepo) lookupStockNames(symbols []string) (map[string]string, error) {
	result := make(map[string]string)
	symbols = normalizeStockSymbolList(symbols)
	if len(symbols) == 0 {
		return result, nil
	}
	for _, symbol := range symbols {
		result[symbol] = ""
	}
	placeholder := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	query := fmt.Sprintf(`
SELECT symbol, name
FROM stock_recommendations
WHERE symbol IN (%s)
ORDER BY created_at DESC`, placeholder)
	args := make([]interface{}, 0, len(symbols))
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var symbol string
		var name sql.NullString
		if err := rows.Scan(&symbol, &name); err != nil {
			return nil, err
		}
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if _, ok := result[symbol]; !ok {
			continue
		}
		if strings.TrimSpace(result[symbol]) != "" {
			continue
		}
		if name.Valid {
			result[symbol] = strings.TrimSpace(name.String)
		}
	}
	return result, nil
}

func buildQuantReasons(item model.StockQuantScore) []string {
	reasons := make([]string, 0, 6)
	if item.Momentum20 >= 6 {
		reasons = append(reasons, fmt.Sprintf("20日动量%.2f%%，中期趋势较强", item.Momentum20))
	}
	if item.VolumeRatio >= 1.2 {
		reasons = append(reasons, fmt.Sprintf("量比%.2f，成交活跃度提升", item.VolumeRatio))
	}
	if item.NetMFAmount > 0 {
		reasons = append(reasons, fmt.Sprintf("主力净流入%.2f，资金面偏强", item.NetMFAmount))
	} else if item.NetMFAmount < 0 {
		reasons = append(reasons, fmt.Sprintf("主力净流出%.2f，需控制仓位", math.Abs(item.NetMFAmount)))
	}
	if item.PeTTM > 0 {
		if item.PeTTM < 30 {
			reasons = append(reasons, fmt.Sprintf("PE(TTM) %.2f，估值处于可接受区间", item.PeTTM))
		} else {
			reasons = append(reasons, fmt.Sprintf("PE(TTM) %.2f，估值偏高需跟踪兑现", item.PeTTM))
		}
	}
	if item.NewsHeat > 0 {
		reasons = append(reasons, fmt.Sprintf("近14天资讯热度%d，正面占比%.0f%%", item.NewsHeat, item.PositiveNewsRate*100))
	}
	if item.Volatility20 <= 2.5 && item.Drawdown20 <= 7 {
		reasons = append(reasons, "波动与回撤控制在可接受范围")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, fmt.Sprintf("趋势评分%.2f，资金评分%.2f，估值评分%.2f", item.TrendScore, item.FlowScore, item.ValueScore))
	}
	return reasons
}

func buildQuantReasonSummary(item model.StockQuantScore) string {
	reasons := item.Reasons
	if len(reasons) == 0 {
		reasons = buildQuantReasons(item)
	}
	return strings.Join(firstNStrings(reasons, 3), "；")
}

func firstNStrings(items []string, count int) []string {
	if count <= 0 || len(items) == 0 {
		return []string{}
	}
	if len(items) <= count {
		return items
	}
	return items[:count]
}

func buildDailyStockCandidatesFromQuant(items []model.StockQuantScore, validFrom time.Time, validTo time.Time) []dailyStockCandidate {
	if len(items) == 0 {
		return nil
	}
	candidates := make([]dailyStockCandidate, 0, len(items))
	for _, item := range items {
		riskLevel := strings.ToUpper(strings.TrimSpace(item.RiskLevel))
		if riskLevel == "" {
			riskLevel = "MEDIUM"
		}
		positionRange := "8%-12%"
		takeProfit := "上涨10%-15%分批止盈"
		stopLoss := "回撤5%止损"
		switch riskLevel {
		case "LOW":
			positionRange = "10%-15%"
			takeProfit = "上涨8%-12%分批止盈"
			stopLoss = "回撤3%止损"
		case "HIGH":
			positionRange = "5%-8%"
			takeProfit = "上涨12%-18%动态止盈"
			stopLoss = "回撤7%止损"
		}
		score := item.Score
		if score <= 0 {
			score = 80
		}
		candidate := dailyStockCandidate{
			Symbol:        strings.ToUpper(strings.TrimSpace(item.Symbol)),
			Name:          strings.TrimSpace(item.Name),
			Score:         roundTo(score, 2),
			RiskLevel:     riskLevel,
			PositionRange: positionRange,
			ValidFrom:     validFrom,
			ValidTo:       validTo,
			Status:        "PUBLISHED",
			ReasonSummary: fmt.Sprintf("量化评分%.2f，%s", score, buildQuantReasonSummary(item)),
			TechScore:     roundTo(clampFloat(58+item.Momentum20*1.6+item.TrendStrength*2.8, 55, 98), 2),
			FundScore:     roundTo(clampFloat(56+item.Momentum20*1.4+item.Momentum5*1.1, 52, 97), 2),
			SentimentScore: roundTo(
				clampFloat(54+item.VolumeRatio*9+item.Momentum5*1.2-item.Volatility20*0.6, 50, 96),
				2,
			),
			MoneyFlowScore: roundTo(
				clampFloat(55+item.VolumeRatio*11+item.TrendStrength*1.5-item.Drawdown20*0.4, 50, 97),
				2,
			),
			TakeProfit: takeProfit,
			StopLoss:   stopLoss,
			RiskNote:   "量化信号存在失效风险，仅供参考，不构成投资建议",
		}
		if candidate.Name == "" {
			candidate.Name = candidate.Symbol
		}
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Symbol < candidates[j].Symbol
		}
		return candidates[i].Score > candidates[j].Score
	})
	return candidates
}

func buildDefaultDailyStockCandidates(validFrom time.Time, validTo time.Time) []dailyStockCandidate {
	samples := []dailyStockCandidate{
		{Symbol: "600519.SH", Name: "贵州茅台", Score: 91.1, RiskLevel: "MEDIUM", PositionRange: "10%-15%", Status: "PUBLISHED", ReasonSummary: "龙头估值修复"},
		{Symbol: "601318.SH", Name: "中国平安", Score: 88.4, RiskLevel: "MEDIUM", PositionRange: "8%-12%", Status: "PUBLISHED", ReasonSummary: "估值低位"},
		{Symbol: "600036.SH", Name: "招商银行", Score: 86.8, RiskLevel: "LOW", PositionRange: "8%-10%", Status: "PUBLISHED", ReasonSummary: "基本面稳健"},
		{Symbol: "600276.SH", Name: "恒瑞医药", Score: 84.5, RiskLevel: "MEDIUM", PositionRange: "6%-10%", Status: "PUBLISHED", ReasonSummary: "创新药预期"},
		{Symbol: "601012.SH", Name: "隆基绿能", Score: 83.2, RiskLevel: "HIGH", PositionRange: "5%-8%", Status: "PUBLISHED", ReasonSummary: "景气拐点博弈"},
		{Symbol: "000333.SZ", Name: "美的集团", Score: 82.1, RiskLevel: "LOW", PositionRange: "6%-10%", Status: "PUBLISHED", ReasonSummary: "外需改善"},
		{Symbol: "300750.SZ", Name: "宁德时代", Score: 87.5, RiskLevel: "MEDIUM", PositionRange: "8%-12%", Status: "PUBLISHED", ReasonSummary: "产业链回暖"},
		{Symbol: "002594.SZ", Name: "比亚迪", Score: 85.3, RiskLevel: "MEDIUM", PositionRange: "7%-11%", Status: "PUBLISHED", ReasonSummary: "销量韧性"},
		{Symbol: "688981.SH", Name: "中芯国际", Score: 80.8, RiskLevel: "HIGH", PositionRange: "5%-8%", Status: "PUBLISHED", ReasonSummary: "国产替代"},
		{Symbol: "601888.SH", Name: "中国中免", Score: 79.9, RiskLevel: "HIGH", PositionRange: "4%-7%", Status: "PUBLISHED", ReasonSummary: "消费复苏博弈"},
	}
	for index := range samples {
		samples[index].ValidFrom = validFrom
		samples[index].ValidTo = validTo
		samples[index].TechScore = roundTo(clampFloat(samples[index].Score-3, 50, 99), 2)
		samples[index].FundScore = roundTo(clampFloat(samples[index].Score-1, 50, 99), 2)
		samples[index].SentimentScore = roundTo(clampFloat(samples[index].Score-5, 50, 99), 2)
		samples[index].MoneyFlowScore = roundTo(clampFloat(samples[index].Score-2, 50, 99), 2)
		samples[index].TakeProfit = "上涨8%-12%分批止盈"
		samples[index].StopLoss = "跌破关键位止损"
		samples[index].RiskNote = "仅供参考，不构成投资建议"
	}
	return samples
}

func normalizeStockRecommendationLifecycleStatus(status string) string {
	return strings.ToUpper(strings.TrimSpace(status))
}

func allowedStockRecommendationTransitions(status string) []string {
	switch normalizeStockRecommendationLifecycleStatus(status) {
	case "DRAFT":
		return []string{"REVIEW_PENDING", "PUBLISHED", "DISABLED", "INVALIDATED"}
	case "REVIEW_PENDING":
		return []string{"DRAFT", "PUBLISHED", "DISABLED", "INVALIDATED"}
	case "PUBLISHED":
		return []string{"REVIEW_PENDING", "TRACKING", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED", "REVIEWED", "DISABLED"}
	case "TRACKING":
		return []string{"PUBLISHED", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED", "REVIEWED", "DISABLED"}
	case "HIT_TAKE_PROFIT":
		return []string{"TRACKING", "REVIEWED", "DISABLED"}
	case "HIT_STOP_LOSS":
		return []string{"TRACKING", "REVIEWED", "DISABLED"}
	case "INVALIDATED":
		return []string{"DRAFT", "REVIEWED", "DISABLED"}
	case "REVIEWED":
		return []string{"PUBLISHED", "TRACKING", "DISABLED"}
	case "ACTIVE":
		return []string{"PUBLISHED", "REVIEW_PENDING", "TRACKING", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED", "REVIEWED", "DISABLED"}
	case "DISABLED":
		return []string{"DRAFT", "REVIEW_PENDING", "PUBLISHED"}
	default:
		return []string{}
	}
}

func canTransitionStockRecommendationStatus(currentStatus string, targetStatus string) bool {
	currentStatus = normalizeStockRecommendationLifecycleStatus(currentStatus)
	targetStatus = normalizeStockRecommendationLifecycleStatus(targetStatus)
	if currentStatus == "" || targetStatus == "" {
		return false
	}
	if currentStatus == targetStatus {
		return true
	}
	for _, allowed := range allowedStockRecommendationTransitions(currentStatus) {
		if allowed == targetStatus {
			return true
		}
	}
	return false
}

func parseFlexibleDateTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, errors.New("datetime is required")
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.999999999Z07:00",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if layout == time.RFC3339 {
			if ts, err := time.Parse(layout, trimmed); err == nil {
				return ts, nil
			}
			continue
		}
		if ts, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime format: %s", trimmed)
}
