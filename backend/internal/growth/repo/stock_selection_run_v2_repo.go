package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) resolveStockSelectionTemplateForRun(runTemplateID string, profileTemplateID string) (*model.StockSelectionProfileTemplate, error) {
	for _, candidate := range []string{strings.TrimSpace(runTemplateID), strings.TrimSpace(profileTemplateID)} {
		if candidate == "" {
			continue
		}
		item, err := r.getStockSelectionProfileTemplate(candidate)
		if err != nil {
			return nil, err
		}
		return &item, nil
	}
	return r.getDefaultStockSelectionProfileTemplate()
}

func mergeStockSelectionConfigMaps(base map[string]any, override map[string]any) map[string]any {
	result := make(map[string]any)
	for key, value := range base {
		result[key] = value
	}
	for key, value := range override {
		result[key] = value
	}
	return result
}

func copyStockSelectionPayloadFields(payload map[string]any, source map[string]any, keys []string) {
	for _, key := range keys {
		value, exists := source[key]
		if !exists || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) == "" {
				continue
			}
			payload[key] = strings.TrimSpace(typed)
		case []string:
			if len(typed) == 0 {
				continue
			}
			payload[key] = typed
		case []any:
			if len(typed) == 0 {
				continue
			}
			payload[key] = typed
		default:
			payload[key] = value
		}
	}
}

func parseJSONMapSlice(text string) []map[string]any {
	text = strings.TrimSpace(text)
	if text == "" || text == "null" {
		return nil
	}
	var items []map[string]any
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		return nil
	}
	return items
}

func buildPublishedPortfolioSnapshot(items []model.StockSelectionPortfolioEntry) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"symbol":            item.Symbol,
			"name":              item.Name,
			"rank":              item.Rank,
			"quant_score":       item.QuantScore,
			"risk_level":        item.RiskLevel,
			"weight_suggestion": item.WeightSuggestion,
			"reason_summary":    item.ReasonSummary,
			"portfolio_role":    item.PortfolioRole,
		})
	}
	return result
}

func buildStockSelectionCompareSummary(previousSnapshot []map[string]any, current []strategyEngineStockPortfolioEntry) map[string]any {
	currentSymbols := make([]string, 0, len(current))
	currentSet := make(map[string]struct{}, len(current))
	for _, item := range current {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		currentSymbols = append(currentSymbols, symbol)
		currentSet[symbol] = struct{}{}
	}
	prevSet := make(map[string]struct{}, len(previousSnapshot))
	for _, item := range previousSnapshot {
		symbol := strings.ToUpper(strings.TrimSpace(asString(item["symbol"])))
		if symbol == "" {
			continue
		}
		prevSet[symbol] = struct{}{}
	}
	added := make([]string, 0)
	removed := make([]string, 0)
	for _, symbol := range currentSymbols {
		if _, ok := prevSet[symbol]; !ok {
			added = append(added, symbol)
		}
	}
	for symbol := range prevSet {
		if _, ok := currentSet[symbol]; !ok {
			removed = append(removed, symbol)
		}
	}
	sort.Strings(removed)
	return map[string]any{
		"current_symbols": currentSymbols,
		"added_symbols":   added,
		"removed_symbols": removed,
	}
}

func (r *MySQLGrowthRepo) loadLatestApprovedPortfolioSnapshot(excludeRunID string) ([]map[string]any, error) {
	var snapshotText sql.NullString
	err := r.db.QueryRow(`
SELECT COALESCE(CAST(published_portfolio_snapshot AS CHAR), '')
FROM stock_selection_publish_reviews
WHERE review_status = 'APPROVED' AND run_id <> ?
ORDER BY approved_at DESC, updated_at DESC
LIMIT 1`, strings.TrimSpace(excludeRunID)).Scan(&snapshotText)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return parseJSONMapSlice(snapshotText.String), nil
}

func (r *MySQLGrowthRepo) loadLatestApprovedPortfolioEntries() ([]model.StockSelectionPortfolioEntry, error) {
	snapshot, err := r.loadLatestApprovedPortfolioSnapshot("")
	if err != nil {
		return nil, err
	}
	items := make([]model.StockSelectionPortfolioEntry, 0, len(snapshot))
	for _, entry := range snapshot {
		items = append(items, model.StockSelectionPortfolioEntry{
			Symbol:           asString(entry["symbol"]),
			Name:             asString(entry["name"]),
			Rank:             asInt(entry["rank"]),
			QuantScore:       asFloat(entry["quant_score"]),
			RiskLevel:        asString(entry["risk_level"]),
			WeightSuggestion: asString(entry["weight_suggestion"]),
			ReasonSummary:    asString(entry["reason_summary"]),
			PortfolioRole:    asString(entry["portfolio_role"]),
		})
	}
	return items, nil
}

func (r *MySQLGrowthRepo) loadStockSelectionRunEvidenceMap(runID string) (map[string]model.StockSelectionRunEvidence, error) {
	items, err := r.AdminListStockSelectionRunEvidence(runID, "")
	if err != nil {
		return nil, err
	}
	result := make(map[string]model.StockSelectionRunEvidence, len(items))
	for _, item := range items {
		key := stockSelectionEvidenceMapKey(item.Symbol, item.Stage)
		result[key] = item
	}
	return result, nil
}

func stockSelectionEvidenceMapKey(symbol string, stage string) string {
	return strings.ToUpper(strings.TrimSpace(symbol)) + "::" + strings.ToUpper(strings.TrimSpace(stage))
}

func (r *MySQLGrowthRepo) loadStockSelectionRunEvaluationStatusMap(runID string) (map[string]string, error) {
	rows, err := r.db.Query(`
SELECT symbol, COUNT(*)
FROM stock_selection_run_evaluations
WHERE run_id = ?
GROUP BY symbol`, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]string{}
	for rows.Next() {
		var symbol string
		var count int
		if err := rows.Scan(&symbol, &count); err != nil {
			return nil, err
		}
		if count > 0 {
			result[strings.ToUpper(strings.TrimSpace(symbol))] = "READY"
		}
	}
	return result, rows.Err()
}

func buildPreviousPublishDiff(symbol string, previousSnapshot []map[string]any) map[string]any {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return map[string]any{}
	}
	for _, item := range previousSnapshot {
		if strings.ToUpper(strings.TrimSpace(asString(item["symbol"]))) == symbol {
			return map[string]any{
				"status":          "UNCHANGED",
				"previous_rank":   asInt(item["rank"]),
				"previous_reason": asString(item["reason_summary"]),
			}
		}
	}
	return map[string]any{"status": "ADDED"}
}

func (r *MySQLGrowthRepo) AdminListStockSelectionRunEvidence(runID string, symbol string) ([]model.StockSelectionRunEvidence, error) {
	args := []any{strings.TrimSpace(runID)}
	query := `
SELECT id, run_id, symbol, stage, COALESCE(name, ''), COALESCE(portfolio_role, ''), COALESCE(evidence_summary, ''),
       COALESCE(CAST(evidence_cards_json AS CHAR), ''), COALESCE(CAST(positive_reasons_json AS CHAR), ''),
       COALESCE(CAST(veto_reasons_json AS CHAR), ''), COALESCE(CAST(theme_tags_json AS CHAR), ''),
       COALESCE(CAST(sector_tags_json AS CHAR), ''), COALESCE(CAST(risk_flags_json AS CHAR), ''),
       DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM stock_selection_run_evidence
WHERE run_id = ?`
	if strings.TrimSpace(symbol) != "" {
		query += " AND symbol = ?"
		args = append(args, strings.ToUpper(strings.TrimSpace(symbol)))
	}
	query += " ORDER BY FIELD(stage, 'PORTFOLIO', 'CANDIDATE_POOL', 'SEED_POOL', 'UNIVERSE'), symbol ASC"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.StockSelectionRunEvidence, 0)
	for rows.Next() {
		var item model.StockSelectionRunEvidence
		var cardsJSON, positiveJSON, vetoJSON, themeJSON, sectorJSON, riskJSON string
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.Symbol,
			&item.Stage,
			&item.Name,
			&item.PortfolioRole,
			&item.EvidenceSummary,
			&cardsJSON,
			&positiveJSON,
			&vetoJSON,
			&themeJSON,
			&sectorJSON,
			&riskJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.EvidenceCards = parseJSONMapSlice(cardsJSON)
		item.PositiveReasons = parseJSONStringList(positiveJSON)
		item.VetoReasons = parseJSONStringList(vetoJSON)
		item.ThemeTags = parseJSONStringList(themeJSON)
		item.SectorTags = parseJSONStringList(sectorJSON)
		item.RiskFlags = parseJSONStringList(riskJSON)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) AdminListStockSelectionRunEvaluations(runID string, symbol string) ([]model.StockSelectionRunEvaluation, error) {
	_ = r.ensureStockSelectionRunEvaluations(runID)
	args := []any{strings.TrimSpace(runID)}
	query := `
SELECT id, run_id, symbol, horizon_day, evaluation_scope, COALESCE(name, ''),
       COALESCE(DATE_FORMAT(entry_date, '%Y-%m-%d'), ''), COALESCE(DATE_FORMAT(exit_date, '%Y-%m-%d'), ''),
       entry_price, exit_price, return_pct, excess_return_pct, max_drawdown_pct, hit_flag, COALESCE(benchmark_symbol, ''),
       DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM stock_selection_run_evaluations
WHERE run_id = ?`
	if strings.TrimSpace(symbol) != "" {
		query += " AND symbol = ?"
		args = append(args, strings.ToUpper(strings.TrimSpace(symbol)))
	}
	query += " ORDER BY symbol ASC, evaluation_scope ASC, horizon_day ASC"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.StockSelectionRunEvaluation, 0)
	for rows.Next() {
		var item model.StockSelectionRunEvaluation
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.Symbol,
			&item.HorizonDay,
			&item.EvaluationScope,
			&item.Name,
			&item.EntryDate,
			&item.ExitDate,
			&item.EntryPrice,
			&item.ExitPrice,
			&item.ReturnPct,
			&item.ExcessReturnPct,
			&item.MaxDrawdownPct,
			&item.HitFlag,
			&item.BenchmarkSymbol,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) AdminCompareStockSelectionRuns(runIDs []string) (model.StockSelectionRunCompareResult, error) {
	normalized := make([]string, 0, len(runIDs))
	seen := map[string]struct{}{}
	for _, item := range runIDs {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}
	result := model.StockSelectionRunCompareResult{Items: []model.StockSelectionRunCompareItem{}}
	if len(normalized) == 0 {
		return result, nil
	}
	var previousSymbols []string
	for index, runID := range normalized {
		run, err := r.AdminGetStockSelectionRun(runID)
		if err != nil {
			return result, err
		}
		portfolio, err := r.AdminListStockSelectionRunPortfolio(runID)
		if err != nil {
			return result, err
		}
		currentSymbols := make([]string, 0, len(portfolio))
		for _, item := range portfolio {
			currentSymbols = append(currentSymbols, item.Symbol)
		}
		added, removed := diffStringSlices(previousSymbols, currentSymbols)
		result.Items = append(result.Items, model.StockSelectionRunCompareItem{
			RunID:            run.RunID,
			TradeDate:        run.TradeDate,
			ProfileID:        run.ProfileID,
			TemplateID:       run.TemplateID,
			TemplateName:     run.TemplateName,
			MarketRegime:     run.MarketRegime,
			Status:           run.Status,
			ReviewStatus:     run.ReviewStatus,
			SelectedCount:    run.SelectedCount,
			PortfolioSymbols: currentSymbols,
			AddedSymbols:     added,
			RemovedSymbols:   removed,
		})
		if index == 0 {
			result.BaseRunID = run.RunID
		}
		previousSymbols = currentSymbols
	}
	return result, nil
}

func diffStringSlices(previous []string, current []string) ([]string, []string) {
	prevSet := map[string]struct{}{}
	currentSet := map[string]struct{}{}
	for _, item := range previous {
		prevSet[strings.ToUpper(strings.TrimSpace(item))] = struct{}{}
	}
	for _, item := range current {
		currentSet[strings.ToUpper(strings.TrimSpace(item))] = struct{}{}
	}
	added := make([]string, 0)
	removed := make([]string, 0)
	for _, item := range current {
		key := strings.ToUpper(strings.TrimSpace(item))
		if _, ok := prevSet[key]; !ok {
			added = append(added, key)
		}
	}
	for _, item := range previous {
		key := strings.ToUpper(strings.TrimSpace(item))
		if _, ok := currentSet[key]; !ok {
			removed = append(removed, key)
		}
	}
	return added, removed
}

func (r *MySQLGrowthRepo) AdminListStockSelectionEvaluationLeaderboard(templateID string, profileID string, marketRegime string) ([]model.StockSelectionEvaluationLeaderboardItem, error) {
	r.ensureStockSelectionEvaluationLeaderboardCoverage(templateID, profileID, marketRegime)
	args := []any{}
	conditions := []string{}
	if strings.TrimSpace(templateID) != "" {
		conditions = append(conditions, "r.template_id = ?")
		args = append(args, strings.TrimSpace(templateID))
	}
	if strings.TrimSpace(profileID) != "" {
		conditions = append(conditions, "r.profile_id = ?")
		args = append(args, strings.TrimSpace(profileID))
	}
	if strings.TrimSpace(marketRegime) != "" {
		conditions = append(conditions, "r.market_regime = ?")
		args = append(args, strings.ToUpper(strings.TrimSpace(marketRegime)))
	}
	filter := ""
	if len(conditions) > 0 {
		filter = " WHERE " + strings.Join(conditions, " AND ")
	}

	rows, err := r.db.Query(`
SELECT
  COALESCE(r.template_id, ''),
  COALESCE(t.name, ''),
  r.profile_id,
  COALESCE(p.name, ''),
  COALESCE(r.market_regime, ''),
  e.horizon_day,
  e.return_pct,
  e.hit_flag,
  e.max_drawdown_pct
FROM stock_selection_run_evaluations e
JOIN stock_selection_runs r ON r.run_id = e.run_id
LEFT JOIN stock_selection_profiles p ON p.id = r.profile_id
LEFT JOIN stock_selection_profile_templates t ON t.id = r.template_id`+filter+`
ORDER BY COALESCE(r.template_id, ''), r.profile_id, COALESCE(r.market_regime, ''), e.horizon_day`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type horizonAggregate struct {
		returnSum float64
		hitSum    float64
		count     int
	}
	type leaderboardKey struct {
		templateID   string
		profileID    string
		marketRegime string
	}
	type leaderboardAggregate struct {
		model.StockSelectionEvaluationLeaderboardItem
		horizon map[int]*horizonAggregate
	}
	aggregateMap := map[leaderboardKey]*leaderboardAggregate{}
	order := make([]leaderboardKey, 0)
	for rows.Next() {
		var templateIDValue, templateName, profileIDValue, profileName, regime string
		var horizonDay int
		var returnPct, maxDrawdown float64
		var hitFlag bool
		if err := rows.Scan(&templateIDValue, &templateName, &profileIDValue, &profileName, &regime, &horizonDay, &returnPct, &hitFlag, &maxDrawdown); err != nil {
			return nil, err
		}
		key := leaderboardKey{templateID: templateIDValue, profileID: profileIDValue, marketRegime: regime}
		entry, ok := aggregateMap[key]
		if !ok {
			entry = &leaderboardAggregate{
				StockSelectionEvaluationLeaderboardItem: model.StockSelectionEvaluationLeaderboardItem{
					TemplateID:       templateIDValue,
					TemplateName:     templateName,
					ProfileID:        profileIDValue,
					ProfileName:      profileName,
					MarketRegime:     regime,
					ReturnByHorizon:  map[string]float64{},
					HitRateByHorizon: map[string]float64{},
				},
				horizon: map[int]*horizonAggregate{},
			}
			aggregateMap[key] = entry
			order = append(order, key)
		}
		if _, exists := entry.horizon[horizonDay]; !exists {
			entry.horizon[horizonDay] = &horizonAggregate{}
		}
		entry.horizon[horizonDay].returnSum += returnPct
		if hitFlag {
			entry.horizon[horizonDay].hitSum += 1
		}
		entry.horizon[horizonDay].count++
		entry.SampleCount++
		if maxDrawdown < entry.MaxDrawdownPct {
			entry.MaxDrawdownPct = maxDrawdown
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]model.StockSelectionEvaluationLeaderboardItem, 0, len(order))
	for _, key := range order {
		entry := aggregateMap[key]
		for horizonDay, aggregate := range entry.horizon {
			if aggregate.count == 0 {
				continue
			}
			horizonKey := fmt.Sprintf("%d", horizonDay)
			entry.ReturnByHorizon[horizonKey] = aggregate.returnSum / float64(aggregate.count)
			entry.HitRateByHorizon[horizonKey] = aggregate.hitSum / float64(aggregate.count)
		}
		items = append(items, entry.StockSelectionEvaluationLeaderboardItem)
	}
	return items, nil
}

func (r *MySQLGrowthRepo) enrichStockSelectionCandidateSnapshots(runID string, items []model.StockSelectionCandidateSnapshot) {
	evidenceMap, _ := r.loadStockSelectionRunEvidenceMap(runID)
	evaluationMap, _ := r.loadStockSelectionRunEvaluationStatusMap(runID)
	previousSnapshot, _ := r.loadLatestApprovedPortfolioSnapshot(runID)
	for index := range items {
		key := stockSelectionEvidenceMapKey(items[index].Symbol, items[index].Stage)
		if evidence, ok := evidenceMap[key]; ok {
			items[index].EvidenceSummary = evidence.EvidenceSummary
			items[index].PortfolioRole = evidence.PortfolioRole
			items[index].RiskSummary = strings.Join(evidence.RiskFlags, "；")
		}
		if items[index].PortfolioRole == "" && items[index].Stage == "PORTFOLIO" {
			items[index].PortfolioRole = "CORE"
		}
		if status, ok := evaluationMap[strings.ToUpper(strings.TrimSpace(items[index].Symbol))]; ok {
			items[index].EvaluationStatus = status
		} else {
			items[index].EvaluationStatus = "PENDING"
		}
		items[index].PreviousPublishDiff = buildPreviousPublishDiff(items[index].Symbol, previousSnapshot)
	}
}

func (r *MySQLGrowthRepo) enrichStockSelectionPortfolioEntries(runID string, items []model.StockSelectionPortfolioEntry) {
	evidenceMap, _ := r.loadStockSelectionRunEvidenceMap(runID)
	evaluationMap, _ := r.loadStockSelectionRunEvaluationStatusMap(runID)
	previousSnapshot, _ := r.loadLatestApprovedPortfolioSnapshot(runID)
	for index := range items {
		key := stockSelectionEvidenceMapKey(items[index].Symbol, "PORTFOLIO")
		if evidence, ok := evidenceMap[key]; ok {
			items[index].EvidenceSummary = evidence.EvidenceSummary
			items[index].PortfolioRole = evidence.PortfolioRole
			items[index].RiskSummary = strings.Join(evidence.RiskFlags, "；")
		}
		if items[index].PortfolioRole == "" {
			items[index].PortfolioRole = "CORE"
		}
		if status, ok := evaluationMap[strings.ToUpper(strings.TrimSpace(items[index].Symbol))]; ok {
			items[index].EvaluationStatus = status
		} else {
			items[index].EvaluationStatus = "PENDING"
		}
		items[index].PreviousPublishDiff = buildPreviousPublishDiff(items[index].Symbol, previousSnapshot)
	}
}
