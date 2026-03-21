package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

var stockSelectionEvaluationHorizons = []int{1, 3, 5, 10, 20}

type stockSelectionEvaluationTarget struct {
	Symbol string
	Name   string
	Scope  string
}

type stockSelectionEvaluationBar struct {
	TradeDate  time.Time
	ClosePrice float64
	LowPrice   float64
}

func (r *MySQLGrowthRepo) ensureStockSelectionRunEvaluations(runID string) error {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil
	}
	tradeDate, benchmarkHint, err := r.loadStockSelectionEvaluationRunMeta(runID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if tradeDate.IsZero() {
		return nil
	}

	targets, err := r.loadStockSelectionEvaluationTargets(runID)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		return nil
	}

	symbols := make([]string, 0, len(targets))
	for _, item := range targets {
		symbols = append(symbols, item.Symbol)
	}
	benchmarkCandidates := buildStockSelectionEvaluationBenchmarkCandidates(benchmarkHint)
	symbols = append(symbols, benchmarkCandidates...)
	symbols = normalizeStockSymbolList(symbols)

	priceMap, err := r.loadStockSelectionEvaluationBarMap(symbols, tradeDate)
	if err != nil {
		return err
	}
	if len(priceMap) == 0 {
		return nil
	}

	benchmarkSymbol := selectStockSelectionEvaluationBenchmarkSymbol(benchmarkCandidates, priceMap, tradeDate)
	records := buildStockSelectionEvaluationRecords(runID, tradeDate, benchmarkSymbol, priceMap, targets)
	if len(records) == 0 {
		return nil
	}
	return r.replaceStockSelectionRunEvaluations(runID, records)
}

func (r *MySQLGrowthRepo) ensureStockSelectionEvaluationLeaderboardCoverage(templateID string, profileID string, marketRegime string) {
	args := []any{}
	conditions := []string{}
	if strings.TrimSpace(templateID) != "" {
		conditions = append(conditions, "template_id = ?")
		args = append(args, strings.TrimSpace(templateID))
	}
	if strings.TrimSpace(profileID) != "" {
		conditions = append(conditions, "profile_id = ?")
		args = append(args, strings.TrimSpace(profileID))
	}
	if strings.TrimSpace(marketRegime) != "" {
		conditions = append(conditions, "market_regime = ?")
		args = append(args, strings.ToUpper(strings.TrimSpace(marketRegime)))
	}
	query := `SELECT run_id FROM stock_selection_runs`
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY trade_date DESC, created_at DESC LIMIT 24"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var runID string
		if err := rows.Scan(&runID); err != nil {
			return
		}
		_ = r.ensureStockSelectionRunEvaluations(runID)
	}
}

func (r *MySQLGrowthRepo) loadStockSelectionEvaluationRunMeta(runID string) (time.Time, string, error) {
	var tradeDate time.Time
	var contextText string
	var reportText string
	err := r.db.QueryRow(`
SELECT r.trade_date, COALESCE(CAST(r.context_meta AS CHAR), ''), COALESCE(CAST(a.report_snapshot AS CHAR), '')
FROM stock_selection_runs r
LEFT JOIN strategy_job_artifacts a ON a.job_id = r.job_id
WHERE r.run_id = ?
LIMIT 1`, runID).Scan(&tradeDate, &contextText, &reportText)
	if err != nil {
		return time.Time{}, "", err
	}
	contextMeta := parseJSONMap(contextText)
	report := unmarshalStrategySnapshotMap(reportText)
	reportContext := mapValue(report["context_meta"])
	evaluationSummary := mapValue(report["evaluation_summary"])
	selectedTradeDate := firstNonEmpty(
		asString(reportContext["selected_trade_date"]),
		asString(contextMeta["selected_trade_date"]),
		tradeDate.Format("2006-01-02"),
	)
	if parsedTradeDate, err := time.Parse("2006-01-02", dateOnly(selectedTradeDate)); err == nil {
		tradeDate = parsedTradeDate
	}
	return tradeDate, canonicalizeStockSelectionBenchmarkSymbol(asString(evaluationSummary["benchmark_symbol"])), nil
}

func (r *MySQLGrowthRepo) loadStockSelectionEvaluationTargets(runID string) ([]stockSelectionEvaluationTarget, error) {
	targets := make([]stockSelectionEvaluationTarget, 0, 32)

	portfolioRows, err := r.db.Query(`
SELECT symbol, COALESCE(name, '')
FROM stock_selection_run_portfolio
WHERE run_id = ?
ORDER BY rank_no ASC, symbol ASC`, runID)
	if err != nil {
		return nil, err
	}
	for portfolioRows.Next() {
		var symbol, name string
		if err := portfolioRows.Scan(&symbol, &name); err != nil {
			portfolioRows.Close()
			return nil, err
		}
		targets = append(targets, stockSelectionEvaluationTarget{
			Symbol: strings.ToUpper(strings.TrimSpace(symbol)),
			Name:   strings.TrimSpace(name),
			Scope:  "PORTFOLIO",
		})
	}
	if err := portfolioRows.Err(); err != nil {
		portfolioRows.Close()
		return nil, err
	}
	portfolioRows.Close()

	candidateRows, err := r.db.Query(`
SELECT symbol, COALESCE(name, '')
FROM stock_selection_run_candidates
WHERE run_id = ? AND stage = 'CANDIDATE_POOL'
ORDER BY rank_no ASC, symbol ASC`, runID)
	if err != nil {
		return nil, err
	}
	for candidateRows.Next() {
		var symbol, name string
		if err := candidateRows.Scan(&symbol, &name); err != nil {
			candidateRows.Close()
			return nil, err
		}
		targets = append(targets, stockSelectionEvaluationTarget{
			Symbol: strings.ToUpper(strings.TrimSpace(symbol)),
			Name:   strings.TrimSpace(name),
			Scope:  "CANDIDATE",
		})
	}
	if err := candidateRows.Err(); err != nil {
		candidateRows.Close()
		return nil, err
	}
	candidateRows.Close()
	return targets, nil
}

func (r *MySQLGrowthRepo) loadStockSelectionEvaluationBarMap(symbols []string, tradeDate time.Time) (map[string][]stockSelectionEvaluationBar, error) {
	symbols = normalizeStockSymbolList(symbols)
	if len(symbols) == 0 {
		return map[string][]stockSelectionEvaluationBar{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	args := make([]any, 0, len(symbols)+2)
	args = append(args, marketAssetClassStock)
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	args = append(args, tradeDate.Format("2006-01-02"))

	query := fmt.Sprintf(`
SELECT instrument_key, trade_date, close_price, COALESCE(low_price, close_price)
FROM market_daily_bar_truth
WHERE asset_class = ?
  AND instrument_key IN (%s)
  AND trade_date >= ?
ORDER BY instrument_key ASC, trade_date ASC`, placeholders)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string][]stockSelectionEvaluationBar{}
	for rows.Next() {
		var symbol string
		var bar stockSelectionEvaluationBar
		var closePrice, lowPrice sql.NullFloat64
		if err := rows.Scan(&symbol, &bar.TradeDate, &closePrice, &lowPrice); err != nil {
			return nil, err
		}
		if !closePrice.Valid || closePrice.Float64 <= 0 {
			continue
		}
		bar.ClosePrice = closePrice.Float64
		if lowPrice.Valid && lowPrice.Float64 > 0 {
			bar.LowPrice = lowPrice.Float64
		} else {
			bar.LowPrice = bar.ClosePrice
		}
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		result[symbol] = append(result[symbol], bar)
	}
	return result, rows.Err()
}

func (r *MySQLGrowthRepo) replaceStockSelectionRunEvaluations(runID string, records []model.StockSelectionRunEvaluation) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM stock_selection_run_evaluations WHERE run_id = ?", runID); err != nil {
		return err
	}
	stmt, err := tx.Prepare(`
INSERT INTO stock_selection_run_evaluations (
  id, run_id, symbol, horizon_day, evaluation_scope, name, entry_date, exit_date,
  entry_price, exit_price, return_pct, excess_return_pct, max_drawdown_pct, hit_flag, benchmark_symbol,
  created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, item := range records {
		if _, err := stmt.Exec(
			item.ID,
			runID,
			item.Symbol,
			item.HorizonDay,
			item.EvaluationScope,
			item.Name,
			item.EntryDate,
			item.ExitDate,
			item.EntryPrice,
			item.ExitPrice,
			item.ReturnPct,
			item.ExcessReturnPct,
			item.MaxDrawdownPct,
			item.HitFlag,
			item.BenchmarkSymbol,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MySQLGrowthRepo) loadStockSelectionEvaluationSummaryByContext(ctx strategyEngineAssetContext, symbol string) (map[string]any, error) {
	runID, err := r.findStockSelectionRunIDByStrategyContext(ctx)
	if err != nil || strings.TrimSpace(runID) == "" {
		return nil, err
	}
	return r.loadStockSelectionEvaluationSummary(runID, symbol, "PORTFOLIO")
}

func (r *MySQLGrowthRepo) findStockSelectionRunIDByStrategyContext(ctx strategyEngineAssetContext) (string, error) {
	publishID := strings.TrimSpace(ctx.record.PublishID)
	jobID := strings.TrimSpace(ctx.record.JobID)
	tradeDate := strings.TrimSpace(ctx.record.TradeDate)
	if publishID == "" && jobID == "" {
		return "", nil
	}
	var runID string
	err := r.db.QueryRow(`
SELECT r.run_id
FROM stock_selection_runs r
LEFT JOIN stock_selection_publish_reviews rv ON rv.run_id = r.run_id
WHERE (? <> '' AND rv.publish_id = ?)
   OR (? <> '' AND r.job_id = ?)
   OR (? <> '' AND ? <> '' AND r.job_id = ? AND DATE_FORMAT(r.trade_date, '%Y-%m-%d') = ?)
ORDER BY rv.approved_at DESC, r.completed_at DESC, r.created_at DESC
LIMIT 1`,
		publishID, publishID,
		jobID, jobID,
		jobID, tradeDate, jobID, tradeDate,
	).Scan(&runID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(runID), nil
}

func (r *MySQLGrowthRepo) loadStockSelectionEvaluationSummary(runID string, symbol string, scope string) (map[string]any, error) {
	runID = strings.TrimSpace(runID)
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	scope = strings.ToUpper(strings.TrimSpace(scope))
	if runID == "" || symbol == "" {
		return nil, nil
	}
	if scope == "" {
		scope = "PORTFOLIO"
	}
	if err := r.ensureStockSelectionRunEvaluations(runID); err != nil {
		return nil, err
	}
	rows, err := r.AdminListStockSelectionRunEvaluations(runID, symbol)
	if err != nil {
		return nil, err
	}
	filtered := make([]model.StockSelectionRunEvaluation, 0, len(rows))
	for _, item := range rows {
		if strings.EqualFold(item.Symbol, symbol) && strings.EqualFold(item.EvaluationScope, scope) {
			filtered = append(filtered, item)
		}
	}
	return buildStockSelectionEvaluationSummary(filtered), nil
}

func buildStockSelectionEvaluationRecords(
	runID string,
	tradeDate time.Time,
	benchmarkSymbol string,
	priceMap map[string][]stockSelectionEvaluationBar,
	targets []stockSelectionEvaluationTarget,
) []model.StockSelectionRunEvaluation {
	records := make([]model.StockSelectionRunEvaluation, 0, len(targets)*len(stockSelectionEvaluationHorizons))
	benchmarkSymbol = strings.ToUpper(strings.TrimSpace(benchmarkSymbol))
	benchmarkBars := priceMap[benchmarkSymbol]
	benchmarkByDate := buildStockSelectionEvaluationBarDateMap(benchmarkBars)

	for _, target := range targets {
		symbol := strings.ToUpper(strings.TrimSpace(target.Symbol))
		if symbol == "" {
			continue
		}
		bars := priceMap[symbol]
		entryIndex := findStockSelectionEvaluationEntryIndex(bars, tradeDate)
		if entryIndex < 0 {
			continue
		}
		entryBar := bars[entryIndex]
		if entryBar.ClosePrice <= 0 {
			continue
		}
		benchmarkEntry, benchmarkHasEntry := benchmarkByDate[entryBar.TradeDate.Format("2006-01-02")]
		for _, horizon := range stockSelectionEvaluationHorizons {
			exitIndex := entryIndex + horizon
			if exitIndex >= len(bars) {
				continue
			}
			exitBar := bars[exitIndex]
			if exitBar.ClosePrice <= 0 {
				continue
			}
			returnPct := roundTo(exitBar.ClosePrice/entryBar.ClosePrice-1, 6)
			maxDrawdown := roundTo(calcStockSelectionMaxDrawdown(bars[entryIndex:exitIndex+1], entryBar.ClosePrice), 6)
			excessReturn := returnPct
			if benchmarkHasEntry && benchmarkEntry.ClosePrice > 0 {
				if benchmarkExit, ok := benchmarkByDate[exitBar.TradeDate.Format("2006-01-02")]; ok && benchmarkExit.ClosePrice > 0 {
					benchmarkReturn := benchmarkExit.ClosePrice/benchmarkEntry.ClosePrice - 1
					excessReturn = roundTo(returnPct-benchmarkReturn, 6)
				}
			}
			records = append(records, model.StockSelectionRunEvaluation{
				ID:              newID("ssev"),
				RunID:           runID,
				Symbol:          symbol,
				Name:            target.Name,
				HorizonDay:      horizon,
				EvaluationScope: target.Scope,
				EntryDate:       entryBar.TradeDate.Format("2006-01-02"),
				ExitDate:        exitBar.TradeDate.Format("2006-01-02"),
				EntryPrice:      roundTo(entryBar.ClosePrice, 6),
				ExitPrice:       roundTo(exitBar.ClosePrice, 6),
				ReturnPct:       returnPct,
				ExcessReturnPct: excessReturn,
				MaxDrawdownPct:  maxDrawdown,
				HitFlag:         returnPct > 0,
				BenchmarkSymbol: benchmarkSymbol,
			})
		}
	}

	sort.SliceStable(records, func(i, j int) bool {
		if records[i].Symbol == records[j].Symbol {
			if records[i].EvaluationScope == records[j].EvaluationScope {
				return records[i].HorizonDay < records[j].HorizonDay
			}
			return records[i].EvaluationScope < records[j].EvaluationScope
		}
		return records[i].Symbol < records[j].Symbol
	})
	return records
}

func buildStockSelectionEvaluationSummary(rows []model.StockSelectionRunEvaluation) map[string]any {
	if len(rows) == 0 {
		return nil
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].HorizonDay < rows[j].HorizonDay
	})
	summary := map[string]any{
		"status":           "PENDING",
		"benchmark_symbol": "",
		"message":          "评估生成中",
	}
	readyCount := 0
	hasHorizon5 := false
	readySet := map[int]struct{}{}
	for _, row := range rows {
		summary[strconv.Itoa(row.HorizonDay)] = map[string]any{
			"status":            "READY",
			"entry_date":        row.EntryDate,
			"exit_date":         row.ExitDate,
			"entry_price":       row.EntryPrice,
			"exit_price":        row.ExitPrice,
			"return_pct":        row.ReturnPct,
			"excess_return_pct": row.ExcessReturnPct,
			"max_drawdown_pct":  row.MaxDrawdownPct,
			"hit_flag":          row.HitFlag,
			"evaluation_scope":  row.EvaluationScope,
			"benchmark_symbol":  row.BenchmarkSymbol,
		}
		readySet[row.HorizonDay] = struct{}{}
		if row.BenchmarkSymbol != "" && summary["benchmark_symbol"] == "" {
			summary["benchmark_symbol"] = row.BenchmarkSymbol
		}
		readyCount++
		if row.HorizonDay == 5 {
			hasHorizon5 = true
		}
	}
	for _, horizon := range stockSelectionEvaluationHorizons {
		key := strconv.Itoa(horizon)
		if _, ok := summary[key]; !ok {
			summary[key] = map[string]any{"status": "PENDING"}
		}
	}
	if hasHorizon5 || len(readySet) == len(stockSelectionEvaluationHorizons) {
		summary["status"] = "READY"
		if len(readySet) == len(stockSelectionEvaluationHorizons) {
			summary["message"] = "已完成 1/3/5/10/20 日评估"
		} else {
			summary["message"] = "已回写阶段性评估结果"
		}
	} else {
		summary["message"] = "评估生成中，已回写部分 horizon"
	}
	summary["ready_count"] = readyCount
	return summary
}

func buildStockSelectionEvaluationBarDateMap(bars []stockSelectionEvaluationBar) map[string]stockSelectionEvaluationBar {
	result := make(map[string]stockSelectionEvaluationBar, len(bars))
	for _, item := range bars {
		result[item.TradeDate.Format("2006-01-02")] = item
	}
	return result
}

func findStockSelectionEvaluationEntryIndex(bars []stockSelectionEvaluationBar, tradeDate time.Time) int {
	for index, item := range bars {
		if item.TradeDate.Equal(tradeDate) || item.TradeDate.After(tradeDate) {
			return index
		}
	}
	return -1
}

func calcStockSelectionMaxDrawdown(bars []stockSelectionEvaluationBar, entryPrice float64) float64 {
	if entryPrice <= 0 {
		return 0
	}
	maxDrawdown := 0.0
	for _, item := range bars {
		lowPrice := item.LowPrice
		if lowPrice <= 0 {
			lowPrice = item.ClosePrice
		}
		if lowPrice <= 0 {
			continue
		}
		drawdown := lowPrice/entryPrice - 1
		if drawdown < maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	return maxDrawdown
}

func buildStockSelectionEvaluationBenchmarkCandidates(hint string) []string {
	items := []string{
		canonicalizeStockSelectionBenchmarkSymbol(hint),
		"000300.SH",
		"000905.SH",
		"000001.SH",
		"399001.SZ",
	}
	return normalizeStockSymbolList(items)
}

func canonicalizeStockSelectionBenchmarkSymbol(value string) string {
	text := strings.ToUpper(strings.TrimSpace(value))
	if text == "" {
		return ""
	}
	switch {
	case strings.HasPrefix(text, "SH") && len(text) == 8:
		return text[2:] + ".SH"
	case strings.HasPrefix(text, "SZ") && len(text) == 8:
		return text[2:] + ".SZ"
	case strings.HasPrefix(text, "BJ") && len(text) == 8:
		return text[2:] + ".BJ"
	default:
		return text
	}
}

func selectStockSelectionEvaluationBenchmarkSymbol(candidates []string, priceMap map[string][]stockSelectionEvaluationBar, tradeDate time.Time) string {
	bestSymbol := ""
	bestCoverage := -1
	for _, symbol := range candidates {
		bars := priceMap[strings.ToUpper(strings.TrimSpace(symbol))]
		entryIndex := findStockSelectionEvaluationEntryIndex(bars, tradeDate)
		if entryIndex < 0 {
			continue
		}
		coverage := len(bars) - entryIndex
		if coverage > bestCoverage {
			bestCoverage = coverage
			bestSymbol = strings.ToUpper(strings.TrimSpace(symbol))
		}
	}
	return bestSymbol
}

func enrichStockSelectionEvaluationMetaFromSummary(explanation *model.StrategyClientExplanation, summary map[string]any) {
	if explanation == nil || len(summary) == 0 {
		return
	}
	if explanation.EvaluationMeta == nil {
		explanation.EvaluationMeta = map[string]any{}
	}
	for key, value := range summary {
		explanation.EvaluationMeta[key] = value
	}
}

func enrichStockSelectionVersionHistoryEvaluationMeta(item *model.StrategyVersionHistoryItem, summary map[string]any) {
	if item == nil || len(summary) == 0 {
		return
	}
	if item.EvaluationMeta == nil {
		item.EvaluationMeta = map[string]any{}
	}
	for key, value := range summary {
		item.EvaluationMeta[key] = value
	}
}

func marshalStockSelectionEvaluationSummary(summary map[string]any) string {
	if len(summary) == 0 {
		return ""
	}
	body, err := json.Marshal(summary)
	if err != nil {
		return ""
	}
	return string(body)
}
