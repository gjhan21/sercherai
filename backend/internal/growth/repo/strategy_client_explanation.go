package repo

import (
	"strconv"
	"strings"

	"sercherai/backend/internal/growth/model"
)

type strategyEngineAssetContext struct {
	record     model.StrategyEnginePublishRecord
	job        model.StrategyEngineJobRecord
	asset      map[string]any
	simulation map[string]any
}

const defaultVersionHistoryLimit = 6

func (r *MySQLGrowthRepo) buildStockStrategyExplanation(item model.StockRecommendation, detail model.StockRecommendationDetail) model.StrategyClientExplanation {
	explanation := model.StrategyClientExplanation{
		SeedSummary:      "系统从日内股票候选池中完成筛选，并给出当前推荐结论。",
		SeedHighlights:   []string{item.Symbol},
		RiskFlags:        compactStrings([]string{detail.RiskNote}),
		Invalidations:    fallbackStockInvalidations(item, detail),
		ConfidenceReason: item.ReasonSummary,
		WorkloadSummary: model.StrategyWorkloadSummary{
			SeedCount:      1,
			CandidateCount: 1,
			SelectedCount:  1,
			AgentCount:     0,
			ScenarioCount:  0,
			FilterSteps: []string{
				"候选池初始化",
				"技术面/资金面/情绪面评分",
				"风险约束过滤",
				"生成推荐结论",
			},
		},
		StrategyVersion: item.StrategyVersion,
		GeneratedAt:     item.ValidFrom,
	}

	ctx, err := r.findStrategyEngineAssetContext("stock-selection", item.Symbol, dateOnly(item.ValidFrom))
	if err != nil || ctx == nil {
		return explanation
	}
	explanation = mergeStrategyExplanation(explanation, buildStrategyExplanationFromContext(
		ctx,
		item.Symbol,
		item.ReasonSummary,
		item.StrategyVersion,
		[]string{
			"市场种子输入",
			"特征工程与打分",
			"多角色评审",
			"情景模拟",
			"风险过滤与发布",
		},
	))
	if summary, summaryErr := r.loadStockSelectionEvaluationSummaryByContext(*ctx, item.Symbol); summaryErr == nil {
		enrichStockSelectionEvaluationMetaFromSummary(&explanation, summary)
	}
	return explanation
}

func (r *MySQLGrowthRepo) buildFuturesStrategyExplanation(item model.FuturesStrategy, guidance model.FuturesGuidance) model.StrategyClientExplanation {
	explanation := model.StrategyClientExplanation{
		SeedSummary:      "系统从期货合约池中完成方向与风险筛选，并输出执行建议。",
		SeedHighlights:   []string{item.Contract},
		RiskFlags:        compactStrings([]string{guidance.RiskLevel, guidance.InvalidCondition}),
		Invalidations:    compactStrings([]string{guidance.InvalidCondition}),
		ConfidenceReason: item.ReasonSummary,
		WorkloadSummary: model.StrategyWorkloadSummary{
			SeedCount:      1,
			CandidateCount: 1,
			SelectedCount:  1,
			AgentCount:     0,
			ScenarioCount:  0,
			FilterSteps: []string{
				"合约种子初始化",
				"方向与价位评估",
				"多角色评审",
				"场景推演",
				"杠杆与风险过滤",
			},
		},
		StrategyVersion: "futures-mvp-v1",
		GeneratedAt:     item.ValidFrom,
	}

	ctx, err := r.findStrategyEngineAssetContext("futures-strategy", item.Contract, dateOnly(item.ValidFrom))
	if err != nil || ctx == nil {
		return explanation
	}
	return mergeStrategyExplanation(explanation, buildStrategyExplanationFromContext(
		ctx,
		item.Contract,
		item.ReasonSummary,
		"futures-mvp-v1",
		[]string{
			"合约池初始化",
			"方向/价位特征评估",
			"多角色评审",
			"情景推演",
			"风险与发布过滤",
		},
	))
}

func (r *MySQLGrowthRepo) findStrategyEngineAssetContext(jobType string, assetKey string, tradeDate string) (*strategyEngineAssetContext, error) {
	ctx, err := r.findLocalStrategyEngineAssetContext(jobType, assetKey, tradeDate)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		return ctx, nil
	}
	return r.findRemoteStrategyEngineAssetContext(jobType, assetKey, tradeDate)
}

func (r *MySQLGrowthRepo) findRemoteStrategyEngineAssetContext(jobType string, assetKey string, tradeDate string) (*strategyEngineAssetContext, error) {
	if r.strategyEngine == nil || strings.TrimSpace(assetKey) == "" {
		return nil, nil
	}
	history, err := r.strategyEngine.listPublishHistory(jobType)
	if err != nil {
		return nil, err
	}

	var fallbackRecord *model.StrategyEnginePublishRecordSummary
	for _, item := range history {
		if !containsString(item.AssetKeys, assetKey) {
			continue
		}
		if tradeDate != "" && item.TradeDate == tradeDate {
			record, err := r.backfillStrategyEnginePublishRecord(item.PublishID)
			if err != nil {
				return nil, err
			}
			job, err := r.resolveStrategyEngineArchiveJob(item.JobID)
			if err != nil {
				return buildStrategyEngineAssetContext(record, model.StrategyEngineJobRecord{}, assetKey), nil
			}
			return buildStrategyEngineAssetContext(record, job, assetKey), nil
		}
		if fallbackRecord == nil {
			copyItem := item
			fallbackRecord = &copyItem
		}
	}

	if fallbackRecord == nil {
		return nil, nil
	}
	record, err := r.backfillStrategyEnginePublishRecord(fallbackRecord.PublishID)
	if err != nil {
		return nil, err
	}
	job, err := r.resolveStrategyEngineArchiveJob(fallbackRecord.JobID)
	if err != nil {
		return buildStrategyEngineAssetContext(record, model.StrategyEngineJobRecord{}, assetKey), nil
	}
	return buildStrategyEngineAssetContext(record, job, assetKey), nil
}

func (r *MySQLGrowthRepo) listStrategyEngineAssetContexts(jobType string, assetKey string, limit int) ([]strategyEngineAssetContext, error) {
	items, err := r.listLocalStrategyEngineAssetContexts(jobType, assetKey, limit)
	if err != nil {
		return nil, err
	}
	if len(items) > 0 {
		return items, nil
	}
	return r.listRemoteStrategyEngineAssetContexts(jobType, assetKey, limit)
}

func (r *MySQLGrowthRepo) listRemoteStrategyEngineAssetContexts(jobType string, assetKey string, limit int) ([]strategyEngineAssetContext, error) {
	if r.strategyEngine == nil || strings.TrimSpace(assetKey) == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = defaultVersionHistoryLimit
	}
	history, err := r.strategyEngine.listPublishHistory(jobType)
	if err != nil {
		return nil, err
	}

	result := make([]strategyEngineAssetContext, 0, limit)
	for _, item := range history {
		if !containsString(item.AssetKeys, assetKey) {
			continue
		}
		record, err := r.backfillStrategyEnginePublishRecord(item.PublishID)
		if err != nil {
			return nil, err
		}
		job, err := r.resolveStrategyEngineArchiveJob(item.JobID)
		if err != nil {
			job = model.StrategyEngineJobRecord{}
		}
		ctx := buildStrategyEngineAssetContext(record, job, assetKey)
		if ctx == nil {
			continue
		}
		result = append(result, *ctx)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func (r *MySQLGrowthRepo) findLocalStrategyEngineAssetContext(jobType string, assetKey string, tradeDate string) (*strategyEngineAssetContext, error) {
	items, err := r.listLocalStrategyEngineAssetContexts(jobType, assetKey, defaultVersionHistoryLimit)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, nil
	}
	if strings.TrimSpace(tradeDate) != "" {
		for _, item := range items {
			if item.record.TradeDate == tradeDate {
				copyItem := item
				return &copyItem, nil
			}
		}
	}
	copyItem := items[0]
	return &copyItem, nil
}

func (r *MySQLGrowthRepo) listLocalStrategyEngineAssetContexts(jobType string, assetKey string, limit int) ([]strategyEngineAssetContext, error) {
	if strings.TrimSpace(jobType) == "" || strings.TrimSpace(assetKey) == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = defaultVersionHistoryLimit
	}

	rows, err := r.db.Query(`
SELECT
  r.job_id,
  COALESCE(CAST(r.payload_snapshot AS CHAR), ''),
  COALESCE(DATE_FORMAT(r.remote_created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.trade_date, '%Y-%m-%d'), ''),
  COALESCE(CAST(a.report_snapshot AS CHAR), ''),
  COALESCE(j.publish_id, ''),
  COALESCE(j.publish_version, 0),
  COALESCE(DATE_FORMAT(j.created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(CAST(j.replay_snapshot AS CHAR), '')
FROM strategy_job_runs r
JOIN strategy_job_artifacts a ON a.job_id = r.job_id
LEFT JOIN strategy_job_replays j ON j.job_id = r.job_id
WHERE r.job_type = ?
ORDER BY COALESCE(j.publish_version, 0) DESC, j.created_at DESC, COALESCE(r.remote_created_at, r.created_at) DESC, r.job_id DESC`, jobType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]strategyEngineAssetContext, 0, limit)
	for rows.Next() {
		var jobID string
		var payloadText string
		var remoteCreatedAt string
		var tradeDateText string
		var reportText string
		var publishID string
		var publishVersion int
		var replayCreatedAt string
		var replayText string
		if err := rows.Scan(
			&jobID,
			&payloadText,
			&remoteCreatedAt,
			&tradeDateText,
			&reportText,
			&publishID,
			&publishVersion,
			&replayCreatedAt,
			&replayText,
		); err != nil {
			return nil, err
		}

		report := unmarshalStrategySnapshotMap(reportText)
		if !containsString(strategySnapshotAssetKeys(report), assetKey) {
			continue
		}

		replay := buildLocalStrategyPublishReplay(publishID, jobID, publishVersion, replayCreatedAt, replayText)
		selectedCount, payloadCount := strategySnapshotCounts(report)
		job := model.StrategyEngineJobRecord{
			JobID:     jobID,
			JobType:   jobType,
			Payload:   unmarshalStrategySnapshotMap(payloadText),
			CreatedAt: remoteCreatedAt,
		}
		record := model.StrategyEnginePublishRecord{
			PublishID:      publishID,
			JobID:          jobID,
			JobType:        jobType,
			Version:        publishVersion,
			CreatedAt:      firstNonEmpty(replayCreatedAt, remoteCreatedAt),
			TradeDate:      tradeDateText,
			SelectedCount:  selectedCount,
			PayloadCount:   payloadCount,
			AssetKeys:      strategySnapshotAssetKeys(report),
			ReportSnapshot: report,
			Replay:         replay,
		}
		ctx := buildStrategyEngineAssetContext(record, job, assetKey)
		if ctx == nil || (len(ctx.asset) == 0 && len(ctx.simulation) == 0) {
			continue
		}
		result = append(result, *ctx)
		if len(result) >= limit {
			break
		}
	}
	return result, rows.Err()
}

func buildLocalStrategyPublishReplay(publishID string, jobID string, publishVersion int, createdAt string, replayText string) model.StrategyEnginePublishReplay {
	replay := model.StrategyEnginePublishReplay{
		PublishID:      strings.TrimSpace(publishID),
		JobID:          strings.TrimSpace(jobID),
		PublishVersion: publishVersion,
		CreatedAt:      strings.TrimSpace(createdAt),
		StorageSource:  "LOCAL_ARCHIVED",
	}
	snapshot := unmarshalStrategySnapshotMap(replayText)
	replay.WarningCount = asInt(snapshot["warning_count"])
	replay.WarningMessages = stringSlice(snapshot["warning_messages"])
	replay.VetoedAssets = stringSlice(snapshot["vetoed_assets"])
	replay.InvalidatedAssets = stringSlice(snapshot["invalidated_assets"])
	replay.Notes = stringSlice(snapshot["notes"])
	return replay
}

func buildStrategyEngineAssetContext(
	record model.StrategyEnginePublishRecord,
	job model.StrategyEngineJobRecord,
	assetKey string,
) *strategyEngineAssetContext {
	report := record.ReportSnapshot
	candidates := sliceOfMaps(report["candidates"])
	if len(candidates) == 0 {
		candidates = sliceOfMaps(report["strategies"])
	}
	simulations := sliceOfMaps(report["simulations"])

	ctx := &strategyEngineAssetContext{
		record: record,
		job:    job,
	}
	for _, item := range candidates {
		key := asString(item["symbol"])
		if key == "" {
			key = asString(item["contract"])
		}
		if key == assetKey {
			ctx.asset = item
			break
		}
	}
	for _, item := range simulations {
		if asString(item["asset_key"]) == assetKey {
			ctx.simulation = item
			break
		}
	}
	return ctx
}

func buildStrategyExplanationFromContext(
	ctx *strategyEngineAssetContext,
	assetKey string,
	confidenceReason string,
	strategyVersion string,
	filterSteps []string,
) model.StrategyClientExplanation {
	report := ctx.record.ReportSnapshot
	seedHighlights := extractSeedHighlights(ctx.job.Payload)
	simulation := model.StrategyExplanationSimulation{
		AssetKey:        assetKey,
		AssetType:       asString(ctx.simulation["asset_type"]),
		Scenarios:       buildExplanationScenarios(sliceOfMaps(ctx.simulation["scenarios"])),
		Agents:          buildExplanationAgents(sliceOfMaps(ctx.simulation["agents"])),
		ConsensusAction: asString(ctx.simulation["consensus_action"]),
		Vetoed:          asBool(ctx.simulation["vetoed"]),
		VetoReason:      asString(ctx.simulation["veto_reason"]),
	}
	agentOpinions := simulation.Agents
	if confidenceReason == "" {
		confidenceReason = asString(ctx.asset["reason_summary"])
	}

	return model.StrategyClientExplanation{
		SeedSummary:      buildSeedSummary(seedHighlights, report),
		SeedHighlights:   seedHighlights,
		GraphSummary:     asString(report["graph_summary"]),
		ConsensusSummary: asString(report["consensus_summary"]),
		Simulations:      compactSimulations([]model.StrategyExplanationSimulation{simulation}),
		AgentOpinions:    agentOpinions,
		RiskFlags: compactStrings(append(
			append(append([]string{}, ctx.record.Replay.WarningMessages...), ctx.record.Replay.Notes...),
			stringSlice(ctx.asset["risk_flags"])...,
		)),
		Invalidations:    compactStrings(stringSlice(ctx.asset["invalidations"])),
		ConfidenceReason: confidenceReason,
		MarketRegime:     asString(report["market_regime"]),
		EvidenceCards:    buildExplanationEvidenceCards(sliceOfMaps(ctx.asset["evidence_cards"])),
		PortfolioRole:    asString(ctx.asset["portfolio_role"]),
		RiskBoundary:     firstNonEmpty(asString(ctx.asset["risk_summary"]), asString(report["risk_summary"])),
		ThemeTags:        stringSlice(ctx.asset["theme_tags"]),
		SectorTags:       stringSlice(ctx.asset["sector_tags"]),
		EvaluationMeta:   mapValue(report["evaluation_summary"]),
		WorkloadSummary: model.StrategyWorkloadSummary{
			SeedCount:      len(seedHighlights),
			CandidateCount: len(sliceOfMaps(report["candidates"])) + len(sliceOfMaps(report["strategies"])),
			SelectedCount:  ctx.record.SelectedCount,
			AgentCount:     len(agentOpinions),
			ScenarioCount:  len(simulation.Scenarios),
			FilterSteps:    filterSteps,
		},
		StrategyVersion: strategyVersion,
		PublishID:       ctx.record.PublishID,
		JobID:           ctx.record.JobID,
		TradeDate:       ctx.record.TradeDate,
		PublishVersion:  ctx.record.Version,
		GeneratedAt:     firstNonEmpty(asString(report["generated_at"]), ctx.record.CreatedAt),
	}
}

func buildStrategyVersionHistoryItem(
	ctx strategyEngineAssetContext,
	confidenceReason string,
	strategyVersion string,
	filterSteps []string,
) model.StrategyVersionHistoryItem {
	explanation := buildStrategyExplanationFromContext(&ctx, firstNonEmpty(asString(ctx.asset["symbol"]), asString(ctx.asset["contract"])), confidenceReason, strategyVersion, filterSteps)
	return model.StrategyVersionHistoryItem{
		PublishID:        ctx.record.PublishID,
		JobID:            ctx.record.JobID,
		TradeDate:        ctx.record.TradeDate,
		PublishVersion:   ctx.record.Version,
		CreatedAt:        ctx.record.CreatedAt,
		StrategyVersion:  explanation.StrategyVersion,
		ReasonSummary:    asString(ctx.asset["reason_summary"]),
		ConfidenceReason: explanation.ConfidenceReason,
		ConsensusSummary: explanation.ConsensusSummary,
		MarketRegime:     explanation.MarketRegime,
		PortfolioRole:    explanation.PortfolioRole,
		RiskBoundary:     explanation.RiskBoundary,
		ThemeTags:        explanation.ThemeTags,
		SectorTags:       explanation.SectorTags,
		RiskFlags:        explanation.RiskFlags,
		Invalidations:    explanation.Invalidations,
		EvaluationMeta:   explanation.EvaluationMeta,
		GeneratedAt:      explanation.GeneratedAt,
	}
}

func buildFallbackVersionHistoryItem(
	publishID string,
	jobID string,
	tradeDate string,
	publishVersion int,
	createdAt string,
	strategyVersion string,
	reasonSummary string,
	explanation model.StrategyClientExplanation,
) model.StrategyVersionHistoryItem {
	return model.StrategyVersionHistoryItem{
		PublishID:        publishID,
		JobID:            jobID,
		TradeDate:        firstNonEmpty(tradeDate, explanation.TradeDate),
		PublishVersion:   publishVersion,
		CreatedAt:        createdAt,
		StrategyVersion:  firstNonEmpty(explanation.StrategyVersion, strategyVersion),
		ReasonSummary:    reasonSummary,
		ConfidenceReason: firstNonEmpty(explanation.ConfidenceReason, reasonSummary),
		ConsensusSummary: explanation.ConsensusSummary,
		MarketRegime:     explanation.MarketRegime,
		PortfolioRole:    explanation.PortfolioRole,
		RiskBoundary:     explanation.RiskBoundary,
		ThemeTags:        explanation.ThemeTags,
		SectorTags:       explanation.SectorTags,
		RiskFlags:        explanation.RiskFlags,
		Invalidations:    explanation.Invalidations,
		EvaluationMeta:   explanation.EvaluationMeta,
		GeneratedAt:      firstNonEmpty(explanation.GeneratedAt, createdAt),
	}
}

func mergeStrategyExplanation(base model.StrategyClientExplanation, live model.StrategyClientExplanation) model.StrategyClientExplanation {
	if live.SeedSummary != "" {
		base.SeedSummary = live.SeedSummary
	}
	if len(live.SeedHighlights) > 0 {
		base.SeedHighlights = live.SeedHighlights
	}
	if live.GraphSummary != "" {
		base.GraphSummary = live.GraphSummary
	}
	if live.ConsensusSummary != "" {
		base.ConsensusSummary = live.ConsensusSummary
	}
	if len(live.Simulations) > 0 {
		base.Simulations = live.Simulations
	}
	if len(live.AgentOpinions) > 0 {
		base.AgentOpinions = live.AgentOpinions
	}
	if len(live.RiskFlags) > 0 {
		base.RiskFlags = live.RiskFlags
	}
	if len(live.Invalidations) > 0 {
		base.Invalidations = live.Invalidations
	}
	if live.ConfidenceReason != "" {
		base.ConfidenceReason = live.ConfidenceReason
	}
	if live.MarketRegime != "" {
		base.MarketRegime = live.MarketRegime
	}
	if len(live.EvidenceCards) > 0 {
		base.EvidenceCards = live.EvidenceCards
	}
	if live.PortfolioRole != "" {
		base.PortfolioRole = live.PortfolioRole
	}
	if live.RiskBoundary != "" {
		base.RiskBoundary = live.RiskBoundary
	}
	if len(live.ThemeTags) > 0 {
		base.ThemeTags = live.ThemeTags
	}
	if len(live.SectorTags) > 0 {
		base.SectorTags = live.SectorTags
	}
	if len(live.EvaluationMeta) > 0 {
		base.EvaluationMeta = live.EvaluationMeta
	}
	base.WorkloadSummary = live.WorkloadSummary
	if live.StrategyVersion != "" {
		base.StrategyVersion = live.StrategyVersion
	}
	if live.PublishID != "" {
		base.PublishID = live.PublishID
	}
	if live.JobID != "" {
		base.JobID = live.JobID
	}
	if live.TradeDate != "" {
		base.TradeDate = live.TradeDate
	}
	if live.PublishVersion > 0 {
		base.PublishVersion = live.PublishVersion
	}
	if live.GeneratedAt != "" {
		base.GeneratedAt = live.GeneratedAt
	}
	return base
}

func buildExplanationScenarios(items []map[string]any) []model.StrategyExplanationScenario {
	result := make([]model.StrategyExplanationScenario, 0, len(items))
	for _, item := range items {
		result = append(result, model.StrategyExplanationScenario{
			Scenario:        asString(item["scenario"]),
			Thesis:          asString(item["thesis"]),
			ScoreAdjustment: asFloat(item["score_adjustment"]),
			Action:          asString(item["action"]),
			RiskSignal:      asString(item["risk_signal"]),
		})
	}
	return result
}

func buildExplanationEvidenceCards(items []map[string]any) []model.StrategyExplanationEvidenceCard {
	result := make([]model.StrategyExplanationEvidenceCard, 0, len(items))
	for _, item := range items {
		card := model.StrategyExplanationEvidenceCard{
			Title: asString(item["title"]),
			Value: asString(item["value"]),
			Note:  asString(item["note"]),
		}
		if card.Title == "" && card.Value == "" && card.Note == "" {
			continue
		}
		result = append(result, card)
	}
	return result
}

func buildExplanationAgents(items []map[string]any) []model.StrategyExplanationAgentOpinion {
	result := make([]model.StrategyExplanationAgentOpinion, 0, len(items))
	for _, item := range items {
		result = append(result, model.StrategyExplanationAgentOpinion{
			Agent:      asString(item["agent"]),
			Stance:     asString(item["stance"]),
			Confidence: asFloat(item["confidence"]),
			Summary:    asString(item["summary"]),
			Veto:       asBool(item["veto"]),
		})
	}
	return result
}

func buildSeedSummary(seedHighlights []string, report map[string]any) string {
	candidateCount := len(sliceOfMaps(report["candidates"])) + len(sliceOfMaps(report["strategies"]))
	contextMeta := mapValue(report["context_meta"])
	contextSummary := buildContextSummary(contextMeta)
	if len(seedHighlights) == 0 {
		if contextSummary != "" {
			return "系统基于" + contextSummary + "完成本次策略筛选。"
		}
		return "系统基于内部候选池完成本次策略筛选。"
	}
	summary := "本次先处理 " + strconvItoa(len(seedHighlights)) + " 个种子输入，再筛到 " + strconvItoa(candidateCount) + " 个可解释候选。"
	if contextSummary != "" {
		summary += " 数据窗口：" + contextSummary + "。"
	}
	return summary
}

func extractSeedHighlights(payload map[string]any) []string {
	if items := stringSlice(payload["seed_symbols"]); len(items) > 0 {
		return items
	}
	if items := stringSlice(payload["contracts"]); len(items) > 0 {
		return items
	}
	return []string{}
}

func fallbackStockInvalidations(item model.StockRecommendation, detail model.StockRecommendationDetail) []string {
	return compactStrings([]string{
		detail.StopLoss,
		"若核心逻辑转弱或风险等级抬升，应停止沿用旧结论。",
		item.ReasonSummary,
	})
}

func compactSimulations(items []model.StrategyExplanationSimulation) []model.StrategyExplanationSimulation {
	result := make([]model.StrategyExplanationSimulation, 0, len(items))
	for _, item := range items {
		if item.AssetKey == "" && len(item.Scenarios) == 0 && len(item.Agents) == 0 {
			continue
		}
		result = append(result, item)
	}
	return result
}

func compactStrings(items []string) []string {
	result := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func buildContextSummary(contextMeta map[string]any) string {
	if len(contextMeta) == 0 {
		return ""
	}
	parts := make([]string, 0, 2)
	selectedTradeDate := asString(contextMeta["selected_trade_date"])
	priceSource := asString(contextMeta["price_source"])
	if selectedTradeDate != "" || priceSource != "" {
		priceText := selectedTradeDate
		if priceText == "" {
			priceText = "最新交易日"
		}
		if priceSource != "" {
			priceText += " 行情(" + priceSource + ")"
		} else {
			priceText += " 行情"
		}
		parts = append(parts, priceText)
	}
	if newsWindowDays := asInt(contextMeta["news_window_days"]); newsWindowDays > 0 {
		parts = append(parts, "近 "+strconvItoa(newsWindowDays)+" 天资讯")
	}
	return strings.Join(parts, " + ")
}

func sliceOfMaps(value any) []map[string]any {
	raw, ok := value.([]any)
	if !ok {
		if typed, ok := value.([]map[string]any); ok {
			return typed
		}
		return []map[string]any{}
	}
	result := make([]map[string]any, 0, len(raw))
	for _, item := range raw {
		if mapped, ok := item.(map[string]any); ok {
			result = append(result, mapped)
		}
	}
	return result
}

func mapValue(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{}
}

func stringSlice(value any) []string {
	raw, ok := value.([]any)
	if !ok {
		if typed, ok := value.([]string); ok {
			return typed
		}
		return []string{}
	}
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		if text := strings.TrimSpace(asString(item)); text != "" {
			result = append(result, text)
		}
	}
	return result
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if strings.TrimSpace(item) == strings.TrimSpace(target) {
			return true
		}
	}
	return false
}

func asString(value any) string {
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return ""
}

func asFloat(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		return 0
	}
}

func asBool(value any) bool {
	v, ok := value.(bool)
	return ok && v
}

func firstNonEmpty(items ...string) string {
	for _, item := range items {
		if strings.TrimSpace(item) != "" {
			return strings.TrimSpace(item)
		}
	}
	return ""
}

func dateOnly(value string) string {
	if len(value) >= 10 {
		return value[:10]
	}
	return strings.TrimSpace(value)
}

func strconvItoa(v int) string {
	return strconv.Itoa(v)
}
