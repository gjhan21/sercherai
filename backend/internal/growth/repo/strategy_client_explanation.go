package repo

import (
	"database/sql"
	"sort"
	"strconv"
	"strings"
	"time"

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
		EvaluationMeta:  normalizeExplanationEvaluationSummary(nil),
		ResearchOutline: []model.StrategyResearchOutlineStep{
			{Slot: "TREND", Title: "趋势与结构", Summary: firstNonEmpty(item.ReasonSummary, "当前推荐理由待确认"), Status: "ACTIVE"},
		},
		ActiveThesisCards: []model.StrategyExplanationThesisCard{
			{Key: "fallback_reason", Title: "当前理由", Summary: firstNonEmpty(item.ReasonSummary, "当前推荐理由待确认"), Status: "ACTIVE", EvidenceSource: "RECOMMENDATION"},
		},
		WatchSignals: buildFallbackWatchSignals(fallbackStockInvalidations(item, detail)),
	}
	applyStrategyL1HistoryFallback(&explanation)
	applyConfidenceCalibrationToExplanation(&explanation)
	if r.db == nil {
		return explanation
	}
	l1Config := r.loadForecastL1RuntimeConfig()
	l2Config := r.loadForecastL2RuntimeConfig()
	l3Config := r.loadForecastL3RuntimeConfig()

	ctx, err := r.findStrategyEngineAssetContext("stock-selection", item.Symbol, dateOnly(item.ValidFrom))
	if err != nil || ctx == nil {
		if l3Config.ClientReadEnabled {
			r.attachLatestStrategyForecastL3ToExplanation(&explanation, model.StrategyForecastL3TargetTypeStock, item.Symbol)
		}
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
		applyL1AdvisoryMemoryAdjustments(&explanation, summary)
	}
	if relatedEvents, eventCards, eventErr := r.listReviewedStockEventEvidence(item.Symbol, item.ValidFrom); eventErr == nil {
		explanation.RelatedEvents = relatedEvents
		explanation.EventEvidenceCards = eventCards
	}
	applyForecastL1DisplayConfigToExplanation(&explanation, mapValue(ctx.record.ReportSnapshot["memory_feedback"]), l1Config)
	applyForecastL2DisplayConfigToExplanation(&explanation, l2Config)
	if l3Config.ClientReadEnabled {
		r.attachLatestStrategyForecastL3ToExplanation(&explanation, model.StrategyForecastL3TargetTypeStock, item.Symbol)
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
		EvaluationMeta:  normalizeExplanationEvaluationSummary(nil),
		ResearchOutline: []model.StrategyResearchOutlineStep{
			{Slot: "STRUCTURE", Title: "结构与供需", Summary: firstNonEmpty(item.ReasonSummary, "当前策略理由待确认"), Status: "ACTIVE"},
		},
		ActiveThesisCards: []model.StrategyExplanationThesisCard{
			{Key: "fallback_reason", Title: "当前理由", Summary: firstNonEmpty(item.ReasonSummary, "当前策略理由待确认"), Status: "ACTIVE", EvidenceSource: "STRATEGY"},
		},
		WatchSignals: buildFallbackWatchSignals(compactStrings([]string{guidance.InvalidCondition})),
	}
	applyStrategyL1HistoryFallback(&explanation)
	applyConfidenceCalibrationToExplanation(&explanation)
	if r.db == nil {
		return explanation
	}
	l1Config := r.loadForecastL1RuntimeConfig()
	l2Config := r.loadForecastL2RuntimeConfig()
	l3Config := r.loadForecastL3RuntimeConfig()

	ctx, err := r.findStrategyEngineAssetContext("futures-strategy", item.Contract, dateOnly(item.ValidFrom))
	if err != nil || ctx == nil {
		if l3Config.ClientReadEnabled {
			r.attachLatestStrategyForecastL3ToExplanation(&explanation, model.StrategyForecastL3TargetTypeFutures, item.Contract)
		}
		return explanation
	}
	explanation = mergeStrategyExplanation(explanation, buildStrategyExplanationFromContext(
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
	applyForecastL1DisplayConfigToExplanation(&explanation, mapValue(ctx.record.ReportSnapshot["memory_feedback"]), l1Config)
	applyForecastL2DisplayConfigToExplanation(&explanation, l2Config)
	if l3Config.ClientReadEnabled {
		r.attachLatestStrategyForecastL3ToExplanation(&explanation, model.StrategyForecastL3TargetTypeFutures, item.Contract)
	}
	return explanation
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
	evaluationMeta := normalizeExplanationEvaluationSummary(mapValue(report["evaluation_summary"]))
	seedHighlights := extractSeedHighlights(ctx.job.Payload)
	evidenceCards := buildExplanationEvidenceCards(sliceOfMaps(ctx.asset["evidence_cards"]))
	relatedEntities := mergeExplanationRelatedEntities(
		buildExplanationRelatedEntities(sliceOfMaps(report["related_entities"])),
		buildExplanationRelatedEntities(sliceOfMaps(ctx.asset["related_entities"])),
	)
	structureSummary := firstNonEmpty(
		asString(ctx.asset["structure_factor_summary"]),
		evidenceCardNoteByTitle(evidenceCards, "结构联动"),
	)
	inventorySummary := firstNonEmpty(
		asString(ctx.asset["inventory_factor_summary"]),
		evidenceCardNoteByTitle(evidenceCards, "库存画像"),
	)
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
	var researchOutline []model.StrategyResearchOutlineStep
	var activeThesis []model.StrategyExplanationThesisCard
	var historicalThesis []model.StrategyExplanationThesisCard
	var watchSignals []model.StrategyExplanationWatchSignal
	var relationshipSnapshot model.StrategyExplanationRelationshipSnapshot
	var scenarioSnapshots []model.StrategyExplanationScenarioSnapshot
	var scenarioMeta model.StrategyExplanationScenarioMeta
	var scenarioSnapshotMeta model.StrategyExplanationScenarioMeta
	var agentScenarioMeta model.StrategyExplanationScenarioMeta
	var l2AgentOpinions []model.StrategyExplanationAgentOpinion
	if asString(ctx.asset["symbol"]) != "" {
		researchOutline, activeThesis, historicalThesis, watchSignals = buildStockResearchBlocks(*ctx, assetKey)
		relationshipSnapshot = buildStockRelationshipSnapshot(*ctx, relatedEntities)
		scenarioSnapshots, scenarioSnapshotMeta = buildStockScenarioSnapshots(*ctx)
		l2AgentOpinions, agentScenarioMeta = buildStockAgentOpinions(*ctx)
	} else {
		researchOutline, activeThesis, historicalThesis, watchSignals = buildFuturesResearchBlocks(*ctx, assetKey)
		relationshipSnapshot = buildFuturesRelationshipSnapshot(*ctx, relatedEntities, inventorySummary, structureSummary)
		scenarioSnapshots, scenarioSnapshotMeta = buildFuturesScenarioSnapshots(*ctx)
		l2AgentOpinions, agentScenarioMeta = buildFuturesAgentOpinions(*ctx)
	}
	scenarioMeta = mergeScenarioMeta(scenarioSnapshotMeta, agentScenarioMeta)
	if scenarioMeta.PrimaryScenario == "" {
		scenarioMeta = buildScenarioMeta(scenarioSnapshots)
	}
	if len(l2AgentOpinions) == 0 {
		l2AgentOpinions = agentOpinions
	}
	if scenarioMeta.ConsensusAction == "" && len(scenarioSnapshots) > 1 {
		scenarioMeta.ConsensusAction = scenarioSnapshots[1].ActionSuggestion
	}
	if scenarioMeta.PrimaryScenario == "" && len(scenarioSnapshots) > 0 {
		scenarioMeta.PrimaryScenario = preferredPrimaryScenario(scenarioSnapshots)
	}

	explanation := model.StrategyClientExplanation{
		SeedSummary:      buildSeedSummary(seedHighlights, report),
		SeedHighlights:   seedHighlights,
		GraphSummary:     asString(report["graph_summary"]),
		GraphSnapshotID:  asString(report["graph_snapshot_id"]),
		ConsensusSummary: asString(report["consensus_summary"]),
		Simulations:      compactSimulations([]model.StrategyExplanationSimulation{simulation}),
		AgentOpinions:    l2AgentOpinions,
		RiskFlags: compactStrings(append(
			append(append([]string{}, ctx.record.Replay.WarningMessages...), ctx.record.Replay.Notes...),
			stringSlice(ctx.asset["risk_flags"])...,
		)),
		Invalidations:         compactStrings(stringSlice(ctx.asset["invalidations"])),
		ConfidenceReason:      confidenceReason,
		MarketRegime:          asString(report["market_regime"]),
		EvidenceCards:         evidenceCards,
		PortfolioRole:         asString(ctx.asset["portfolio_role"]),
		RiskBoundary:          firstNonEmpty(asString(ctx.asset["risk_summary"]), asString(report["risk_summary"])),
		ThemeTags:             stringSlice(ctx.asset["theme_tags"]),
		SectorTags:            stringSlice(ctx.asset["sector_tags"]),
		SupplyChainNotes:      buildFuturesSupplyChainNotes(relatedEntities, evidenceCards, inventorySummary, structureSummary),
		StructureSummary:      structureSummary,
		InventorySummary:      inventorySummary,
		ResearchOutline:       researchOutline,
		ActiveThesisCards:     activeThesis,
		HistoricalThesisCards: historicalThesis,
		WatchSignals:          watchSignals,
		RelationshipSnapshot:  relationshipSnapshot,
		ScenarioSnapshots:     scenarioSnapshots,
		ScenarioMeta:          scenarioMeta,
		RelatedEntities:       relatedEntities,
		MemoryFeedback:        buildExplanationMemoryFeedback(mapValue(report["memory_feedback"])),
		EvaluationMeta:        evaluationMeta,
		WorkloadSummary: model.StrategyWorkloadSummary{
			SeedCount:      len(seedHighlights),
			CandidateCount: len(sliceOfMaps(report["candidates"])) + len(sliceOfMaps(report["strategies"])),
			SelectedCount:  ctx.record.SelectedCount,
			AgentCount:     maxInt(len(agentOpinions), len(l2AgentOpinions)),
			ScenarioCount:  maxInt(len(simulation.Scenarios), len(scenarioSnapshots)),
			FilterSteps:    filterSteps,
		},
		StrategyVersion: firstNonEmpty(asString(ctx.asset["strategy_version"]), strategyVersion),
		PublishID:       ctx.record.PublishID,
		JobID:           ctx.record.JobID,
		TradeDate:       ctx.record.TradeDate,
		PublishVersion:  ctx.record.Version,
		GeneratedAt:     firstNonEmpty(asString(report["generated_at"]), ctx.record.CreatedAt),
	}
	applyL1AdvisoryMemoryAdjustments(&explanation, evaluationMeta)
	applyConfidenceCalibrationToExplanation(&explanation)
	return explanation
}

func buildFallbackWatchSignals(invalidations []string) []model.StrategyExplanationWatchSignal {
	result := make([]model.StrategyExplanationWatchSignal, 0, len(invalidations))
	for _, item := range compactStrings(invalidations) {
		result = append(result, model.StrategyExplanationWatchSignal{
			Title:      "基础失效信号",
			SignalType: "INVALIDATION",
			Trigger:    item,
			Action:     "重新验证",
			Priority:   "HIGH",
		})
	}
	return result
}

func applyStrategyL1HistoryFallback(explanation *model.StrategyClientExplanation) {
	if explanation == nil || len(explanation.HistoricalThesisCards) > 0 || len(explanation.Invalidations) == 0 {
		return
	}
	explanation.HistoricalThesisCards = []model.StrategyExplanationThesisCard{
		{
			Key:     "fallback_invalidation",
			Title:   "历史弱化理由",
			Summary: explanation.Invalidations[0],
			Status:  "WEAKENED",
			Note:    "缺少更完整历史上下文时，先用当前失效边界兜底。",
		},
	}
}

func applyStrategyL1HistoryFromContexts(explanation *model.StrategyClientExplanation, contexts []strategyEngineAssetContext, assetKey string) {
	if explanation == nil {
		return
	}
	if len(contexts) > 1 {
		previous := contexts[1]
		summary := strings.TrimSpace(asString(previous.asset["reason_summary"]))
		note := strings.Join(compactStrings(stringSlice(previous.asset["invalidations"])), "；")
		if summary != "" {
			explanation.HistoricalThesisCards = append(explanation.HistoricalThesisCards, model.StrategyExplanationThesisCard{
				Key:     "previous_context",
				Title:   "上一版理由",
				Summary: summary,
				Status:  "WEAKENED",
				Note:    note,
			})
		}
	}
	applyStrategyL1HistoryFallback(explanation)
}

func applyStrategyL1HistoryToHistoryItems(items []model.StrategyVersionHistoryItem, contexts []strategyEngineAssetContext, assetKey string) {
	for index := range items {
		if len(items[index].HistoricalThesisCards) == 0 {
			if len(contexts) > index+1 {
				previous := contexts[index+1]
				summary := strings.TrimSpace(asString(previous.asset["reason_summary"]))
				note := strings.Join(compactStrings(stringSlice(previous.asset["invalidations"])), "；")
				if summary != "" {
					items[index].HistoricalThesisCards = append(items[index].HistoricalThesisCards, model.StrategyExplanationThesisCard{
						Key:     "previous_context",
						Title:   "上一版理由",
						Summary: summary,
						Status:  "WEAKENED",
						Note:    note,
					})
				}
			}
			if len(items[index].HistoricalThesisCards) == 0 && len(items[index].Invalidations) > 0 {
				items[index].HistoricalThesisCards = append(items[index].HistoricalThesisCards, model.StrategyExplanationThesisCard{
					Key:     "fallback_invalidation",
					Title:   "历史弱化理由",
					Summary: items[index].Invalidations[0],
					Status:  "WEAKENED",
					Note:    "缺少更完整历史上下文时，先用当前失效边界兜底。",
				})
			}
		}
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
		PublishID:             ctx.record.PublishID,
		JobID:                 ctx.record.JobID,
		TradeDate:             ctx.record.TradeDate,
		PublishVersion:        ctx.record.Version,
		CreatedAt:             ctx.record.CreatedAt,
		StrategyVersion:       explanation.StrategyVersion,
		ReasonSummary:         asString(ctx.asset["reason_summary"]),
		ConfidenceReason:      explanation.ConfidenceReason,
		ConsensusSummary:      explanation.ConsensusSummary,
		GraphSummary:          explanation.GraphSummary,
		GraphSnapshotID:       explanation.GraphSnapshotID,
		MarketRegime:          explanation.MarketRegime,
		PortfolioRole:         explanation.PortfolioRole,
		RiskBoundary:          explanation.RiskBoundary,
		ThemeTags:             explanation.ThemeTags,
		SectorTags:            explanation.SectorTags,
		ResearchOutline:       explanation.ResearchOutline,
		ActiveThesisCards:     explanation.ActiveThesisCards,
		HistoricalThesisCards: explanation.HistoricalThesisCards,
		WatchSignals:          explanation.WatchSignals,
		ConfidenceCalibration: explanation.ConfidenceCalibration,
		RelationshipSnapshot:  explanation.RelationshipSnapshot,
		ScenarioSnapshots:     explanation.ScenarioSnapshots,
		ScenarioMeta:          explanation.ScenarioMeta,
		AgentOpinions:         explanation.AgentOpinions,
		RelatedEntities:       explanation.RelatedEntities,
		MemoryFeedback:        explanation.MemoryFeedback,
		RiskFlags:             explanation.RiskFlags,
		Invalidations:         explanation.Invalidations,
		EvaluationMeta:        explanation.EvaluationMeta,
		DeepForecastSummary:   cloneStrategyForecastL3SummaryPtr(explanation.DeepForecastSummary),
		DeepForecastReportRef: cloneStrategyForecastL3ReportRefPtr(explanation.DeepForecastReportRef),
		VersionDiff:           explanation.VersionDiff,
		GeneratedAt:           explanation.GeneratedAt,
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
		PublishID:             publishID,
		JobID:                 jobID,
		TradeDate:             firstNonEmpty(tradeDate, explanation.TradeDate),
		PublishVersion:        publishVersion,
		CreatedAt:             createdAt,
		StrategyVersion:       firstNonEmpty(explanation.StrategyVersion, strategyVersion),
		ReasonSummary:         reasonSummary,
		ConfidenceReason:      firstNonEmpty(explanation.ConfidenceReason, reasonSummary),
		ConsensusSummary:      explanation.ConsensusSummary,
		GraphSummary:          explanation.GraphSummary,
		GraphSnapshotID:       explanation.GraphSnapshotID,
		MarketRegime:          explanation.MarketRegime,
		PortfolioRole:         explanation.PortfolioRole,
		RiskBoundary:          explanation.RiskBoundary,
		ThemeTags:             explanation.ThemeTags,
		SectorTags:            explanation.SectorTags,
		ResearchOutline:       explanation.ResearchOutline,
		ActiveThesisCards:     explanation.ActiveThesisCards,
		HistoricalThesisCards: explanation.HistoricalThesisCards,
		WatchSignals:          explanation.WatchSignals,
		ConfidenceCalibration: explanation.ConfidenceCalibration,
		RelationshipSnapshot:  explanation.RelationshipSnapshot,
		ScenarioSnapshots:     explanation.ScenarioSnapshots,
		ScenarioMeta:          explanation.ScenarioMeta,
		AgentOpinions:         explanation.AgentOpinions,
		RelatedEntities:       explanation.RelatedEntities,
		MemoryFeedback:        explanation.MemoryFeedback,
		RiskFlags:             explanation.RiskFlags,
		Invalidations:         explanation.Invalidations,
		EvaluationMeta:        explanation.EvaluationMeta,
		DeepForecastSummary:   cloneStrategyForecastL3SummaryPtr(explanation.DeepForecastSummary),
		DeepForecastReportRef: cloneStrategyForecastL3ReportRefPtr(explanation.DeepForecastReportRef),
		VersionDiff:           explanation.VersionDiff,
		GeneratedAt:           firstNonEmpty(explanation.GeneratedAt, createdAt),
	}
}

func (r *MySQLGrowthRepo) attachLatestStrategyForecastL3ToExplanation(
	explanation *model.StrategyClientExplanation,
	targetType string,
	targetKey string,
) {
	if explanation == nil {
		return
	}
	run, err := r.findLatestReadableStrategyForecastL3Run(targetType, targetKey)
	if err != nil || strings.TrimSpace(run.ID) == "" {
		return
	}
	explanation.DeepForecastSummary = cloneStrategyForecastL3SummaryPtr(&run.Summary)
	explanation.DeepForecastReportRef = cloneStrategyForecastL3ReportRefPtr(run.ReportRef)
}

func (r *MySQLGrowthRepo) attachLatestStrategyForecastL3ToHistoryItems(
	items []model.StrategyVersionHistoryItem,
	targetType string,
	targetKey string,
) {
	if len(items) == 0 {
		return
	}
	run, err := r.findLatestReadableStrategyForecastL3Run(targetType, targetKey)
	if err != nil || strings.TrimSpace(run.ID) == "" {
		return
	}
	for index := range items {
		items[index].DeepForecastSummary = cloneStrategyForecastL3SummaryPtr(&run.Summary)
		items[index].DeepForecastReportRef = cloneStrategyForecastL3ReportRefPtr(run.ReportRef)
	}
}

func (r *MySQLGrowthRepo) findLatestReadableStrategyForecastL3Run(targetType string, targetKey string) (model.StrategyForecastL3Run, error) {
	if r == nil || r.db == nil || strings.TrimSpace(targetType) == "" || strings.TrimSpace(targetKey) == "" {
		return model.StrategyForecastL3Run{}, sql.ErrNoRows
	}
	row := r.db.QueryRow(`
SELECT
	id,
	target_type,
	COALESCE(target_id, ''),
	target_key,
	COALESCE(target_label, ''),
	trigger_type,
	COALESCE(request_user_id, ''),
	COALESCE(operator_user_id, ''),
	engine_key,
	status,
	priority_score,
	COALESCE(reason, ''),
	COALESCE(failure_reason, ''),
	COALESCE(CAST(context_meta_json AS CHAR), ''),
	COALESCE(CAST(summary_json AS CHAR), ''),
	COALESCE(CAST(report_ref_json AS CHAR), ''),
	queued_at,
	started_at,
	finished_at,
	cancelled_at,
	created_at,
	updated_at
FROM strategy_forecast_l3_runs
WHERE target_type = ?
  AND target_key = ?
  AND status IN (?, ?, ?)
ORDER BY
  CASE status
    WHEN 'RUNNING' THEN 0
    WHEN 'QUEUED' THEN 1
    ELSE 2
  END,
  created_at DESC,
  id DESC
LIMIT 1`,
		strings.TrimSpace(targetType),
		strings.TrimSpace(targetKey),
		model.StrategyForecastL3StatusRunning,
		model.StrategyForecastL3StatusQueued,
		model.StrategyForecastL3StatusSucceeded,
	)
	return scanStrategyForecastL3Run(row)
}

func cloneStrategyForecastL3SummaryPtr(input *model.StrategyForecastL3Summary) *model.StrategyForecastL3Summary {
	if input == nil {
		return nil
	}
	copyItem := *input
	return &copyItem
}

func cloneStrategyForecastL3ReportRefPtr(input *model.StrategyForecastL3ReportRef) *model.StrategyForecastL3ReportRef {
	if input == nil {
		return nil
	}
	copyItem := *input
	return &copyItem
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
	if live.GraphSnapshotID != "" {
		base.GraphSnapshotID = live.GraphSnapshotID
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
	if live.RelationshipSnapshot.RelationshipCount > 0 || len(live.RelationshipSnapshot.Nodes) > 0 {
		base.RelationshipSnapshot = live.RelationshipSnapshot
	}
	if len(live.ScenarioSnapshots) > 0 {
		base.ScenarioSnapshots = live.ScenarioSnapshots
	}
	if live.ScenarioMeta.PrimaryScenario != "" || live.ScenarioMeta.ConsensusAction != "" || live.ScenarioMeta.Vetoed {
		base.ScenarioMeta = live.ScenarioMeta
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
	if len(live.ResearchOutline) > 0 {
		base.ResearchOutline = live.ResearchOutline
	}
	if len(live.ActiveThesisCards) > 0 {
		base.ActiveThesisCards = live.ActiveThesisCards
	}
	if len(live.HistoricalThesisCards) > 0 {
		base.HistoricalThesisCards = live.HistoricalThesisCards
	}
	if len(live.WatchSignals) > 0 {
		base.WatchSignals = live.WatchSignals
	}
	if live.ConfidenceCalibration.AdjustedConfidence > 0 || live.ConfidenceCalibration.BaseConfidence > 0 {
		base.ConfidenceCalibration = live.ConfidenceCalibration
	}
	if len(live.SupplyChainNotes) > 0 {
		base.SupplyChainNotes = live.SupplyChainNotes
	}
	if live.StructureSummary != "" {
		base.StructureSummary = live.StructureSummary
	}
	if live.InventorySummary != "" {
		base.InventorySummary = live.InventorySummary
	}
	if len(live.RelatedEntities) > 0 {
		base.RelatedEntities = live.RelatedEntities
	}
	if len(live.RelatedEvents) > 0 {
		base.RelatedEvents = live.RelatedEvents
	}
	if len(live.EventEvidenceCards) > 0 {
		base.EventEvidenceCards = live.EventEvidenceCards
	}
	if live.MemoryFeedback.Summary != "" || len(live.MemoryFeedback.Items) > 0 {
		base.MemoryFeedback = live.MemoryFeedback
	}
	if len(live.EvaluationMeta) > 0 {
		base.EvaluationMeta = live.EvaluationMeta
	}
	if hasStrategyVersionDiff(live.VersionDiff) {
		base.VersionDiff = live.VersionDiff
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

func (r *MySQLGrowthRepo) listReviewedStockEventEvidence(symbol string, anchorTime string) ([]model.StrategyExplanationRelatedEvent, []model.StrategyExplanationEvidenceCard, error) {
	if r.db == nil {
		return nil, nil, nil
	}
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, nil, nil
	}
	endAt := time.Now()
	if parsed, ok := parseRFC3339(anchorTime); ok {
		endAt = parsed
	}
	startAt := endAt.AddDate(0, 0, -14)

	rows, err := r.db.Query(`
SELECT
  c.id,
  c.title,
  c.event_type,
  COALESCE(c.primary_symbol, ''),
  COALESCE(c.topic_label, ''),
  COALESCE(c.sector_label, ''),
  COALESCE(JSON_UNQUOTE(JSON_EXTRACT(c.metadata_json, '$.review_priority')), ''),
  COALESCE(lr.review_note, ''),
  c.published_at
FROM stock_event_clusters c
LEFT JOIN (
  SELECT r1.*
  FROM stock_event_reviews r1
  INNER JOIN (
    SELECT cluster_id, MAX(created_at) AS latest_created_at
    FROM stock_event_reviews
    GROUP BY cluster_id
  ) latest ON latest.cluster_id = r1.cluster_id AND latest.latest_created_at = r1.created_at
) lr ON lr.cluster_id = c.id
WHERE c.review_status = 'APPROVED'
  AND (
    c.primary_symbol = ?
    OR EXISTS (
      SELECT 1 FROM stock_event_entities se
      WHERE se.cluster_id = c.id AND se.symbol = ?
    )
  )
  AND COALESCE(c.published_at, c.updated_at) >= ?
  AND COALESCE(c.published_at, c.updated_at) <= ?
ORDER BY CASE COALESCE(JSON_UNQUOTE(JSON_EXTRACT(c.metadata_json, '$.review_priority')), '') WHEN 'HIGH' THEN 0 ELSE 1 END ASC,
         COALESCE(c.published_at, c.updated_at) DESC,
         c.id DESC
LIMIT 3`, symbol, symbol, startAt, endAt)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	events := make([]model.StrategyExplanationRelatedEvent, 0, 3)
	cards := make([]model.StrategyExplanationEvidenceCard, 0, 3)
	for rows.Next() {
		var event model.StrategyExplanationRelatedEvent
		var reviewNote sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(
			&event.ClusterID,
			&event.Title,
			&event.EventType,
			&event.PrimarySymbol,
			&event.TopicLabel,
			&event.SectorLabel,
			&event.ReviewPriority,
			&reviewNote,
			&publishedAt,
		); err != nil {
			return nil, nil, err
		}
		event.ReviewNote = strings.TrimSpace(reviewNote.String)
		event.PublishedAt = formatNullTime(publishedAt)
		event.Tags = compactStrings([]string{event.EventType, event.TopicLabel, event.SectorLabel, event.ReviewPriority})
		events = append(events, event)

		noteParts := compactStrings([]string{
			event.ReviewNote,
			firstNonEmpty(event.TopicLabel, event.SectorLabel),
			event.PublishedAt,
		})
		cards = append(cards, model.StrategyExplanationEvidenceCard{
			Title: firstNonEmpty(event.EventType, "EVENT"),
			Value: event.Title,
			Note:  strings.Join(noteParts, " · "),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return events, cards, nil
}

type strategyVersionDiffAsset struct {
	Key           string
	Rank          int
	ReasonSummary string
	RiskSummary   string
}

func applyStrategyVersionDiffToExplanation(explanation *model.StrategyClientExplanation, contexts []strategyEngineAssetContext, assetKey string) {
	if explanation == nil || len(contexts) < 2 {
		return
	}
	diff := buildStrategyVersionDiffFromContexts(contexts[0], contexts[1], assetKey)
	if hasStrategyVersionDiff(diff) {
		explanation.VersionDiff = diff
	}
}

func applyStrategyVersionDiffToHistoryItems(items []model.StrategyVersionHistoryItem, contexts []strategyEngineAssetContext, assetKey string) {
	limit := len(items)
	if len(contexts)-1 < limit {
		limit = len(contexts) - 1
	}
	for index := 0; index < limit; index++ {
		diff := buildStrategyVersionDiffFromContexts(contexts[index], contexts[index+1], assetKey)
		if hasStrategyVersionDiff(diff) {
			items[index].VersionDiff = diff
		}
	}
}

func buildStrategyVersionDiffFromContexts(current strategyEngineAssetContext, previous strategyEngineAssetContext, assetKey string) model.StrategyVersionDiff {
	currentAssets := buildStrategyVersionDiffAssets(current.record.ReportSnapshot)
	previousAssets := buildStrategyVersionDiffAssets(previous.record.ReportSnapshot)
	if len(currentAssets) == 0 && len(previousAssets) == 0 {
		return model.StrategyVersionDiff{}
	}

	currentKeys := sortedStrategyVersionDiffAssetKeys(currentAssets)
	previousKeys := sortedStrategyVersionDiffAssetKeys(previousAssets)
	previousSet := make(map[string]struct{}, len(previousKeys))
	currentSet := make(map[string]struct{}, len(currentKeys))
	for _, key := range previousKeys {
		previousSet[key] = struct{}{}
	}
	for _, key := range currentKeys {
		currentSet[key] = struct{}{}
	}

	added := make([]string, 0)
	removed := make([]string, 0)
	promoted := make([]string, 0)
	downgradeReasons := make([]string, 0)

	for _, key := range currentKeys {
		if _, ok := previousSet[key]; !ok {
			added = append(added, key)
		}
	}
	for _, key := range previousKeys {
		if _, ok := currentSet[key]; !ok {
			removed = append(removed, key)
		}
	}

	assetKey = strings.ToUpper(strings.TrimSpace(assetKey))
	for _, key := range currentKeys {
		currentAsset, ok := currentAssets[key]
		if !ok {
			continue
		}
		previousAsset, existsBefore := previousAssets[key]
		if !existsBefore {
			continue
		}
		if currentAsset.Rank > 0 && previousAsset.Rank > 0 {
			if currentAsset.Rank < previousAsset.Rank {
				promoted = append(promoted, key)
			}
			if currentAsset.Rank > previousAsset.Rank {
				downgradeReasons = append(downgradeReasons, key+" 排位由 "+strconvItoa(previousAsset.Rank)+" 调整到 "+strconvItoa(currentAsset.Rank))
			}
		}
		if assetKey != "" && key == assetKey {
			if previousAsset.ReasonSummary != "" && currentAsset.ReasonSummary != "" && previousAsset.ReasonSummary != currentAsset.ReasonSummary {
				downgradeReasons = append(downgradeReasons, key+" 理由更新为："+currentAsset.ReasonSummary)
			}
			if previousAsset.RiskSummary != "" && currentAsset.RiskSummary != "" && previousAsset.RiskSummary != currentAsset.RiskSummary {
				downgradeReasons = append(downgradeReasons, key+" 风险边界更新为："+currentAsset.RiskSummary)
			}
		}
	}

	currentAssetChange := ""
	if assetKey != "" {
		if containsString(added, assetKey) {
			currentAssetChange = "ADDED"
		} else if containsString(promoted, assetKey) {
			currentAssetChange = "PROMOTED"
		} else {
			for _, reason := range downgradeReasons {
				if strings.HasPrefix(reason, assetKey+" ") {
					currentAssetChange = "WEAKENED"
					break
				}
			}
			if currentAssetChange == "" {
				if _, ok := currentAssets[assetKey]; ok {
					currentAssetChange = "UNCHANGED"
				}
			}
		}
	}

	summaryParts := make([]string, 0, 4)
	if len(added) > 0 {
		summaryParts = append(summaryParts, "新增 "+strconvItoa(len(added))+" 个")
	}
	if len(removed) > 0 {
		summaryParts = append(summaryParts, "移除 "+strconvItoa(len(removed))+" 个")
	}
	if len(promoted) > 0 {
		summaryParts = append(summaryParts, "上调 "+strconvItoa(len(promoted))+" 个")
	}
	if len(downgradeReasons) > 0 {
		summaryParts = append(summaryParts, "下降提示 "+strconvItoa(len(downgradeReasons))+" 条")
	}

	return model.StrategyVersionDiff{
		ComparePublishID:   previous.record.PublishID,
		CompareVersion:     previous.record.Version,
		Added:              added,
		Removed:            removed,
		Promoted:           compactStrings(promoted),
		DowngradeReasons:   compactStrings(downgradeReasons),
		CurrentAssetChange: strings.TrimSpace(currentAssetChange),
		Summary:            strings.Join(summaryParts, " · "),
	}
}

func buildStrategyVersionDiffAssets(report map[string]any) map[string]strategyVersionDiffAsset {
	result := map[string]strategyVersionDiffAsset{}
	mergeAsset := func(key string, rank int, reasonSummary string, riskSummary string) {
		key = strings.ToUpper(strings.TrimSpace(key))
		if key == "" {
			return
		}
		current := result[key]
		current.Key = key
		if rank > 0 && (current.Rank == 0 || rank < current.Rank) {
			current.Rank = rank
		}
		if strings.TrimSpace(reasonSummary) != "" {
			current.ReasonSummary = strings.TrimSpace(reasonSummary)
		}
		if strings.TrimSpace(riskSummary) != "" {
			current.RiskSummary = strings.TrimSpace(riskSummary)
		}
		result[key] = current
	}

	collect := func(items []map[string]any) {
		for _, item := range items {
			mergeAsset(
				firstNonEmpty(asString(item["symbol"]), asString(item["contract"])),
				asInt(item["rank"]),
				asString(item["reason_summary"]),
				asString(item["risk_summary"]),
			)
		}
	}

	collect(sliceOfMaps(report["portfolio_entries"]))
	collect(sliceOfMaps(report["candidates"]))
	collect(sliceOfMaps(report["strategies"]))

	for _, item := range sliceOfMaps(report["publish_payloads"]) {
		recommendation := mapValue(item["recommendation"])
		strategy := mapValue(item["strategy"])
		if len(recommendation) > 0 {
			mergeAsset(asString(recommendation["symbol"]), 0, asString(recommendation["reason_summary"]), "")
		}
		if len(strategy) > 0 {
			mergeAsset(asString(strategy["contract"]), 0, asString(strategy["reason_summary"]), "")
		}
	}

	return result
}

func sortedStrategyVersionDiffAssetKeys(items map[string]strategyVersionDiffAsset) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func hasStrategyVersionDiff(diff model.StrategyVersionDiff) bool {
	return strings.TrimSpace(diff.Summary) != "" ||
		strings.TrimSpace(diff.ComparePublishID) != "" ||
		len(diff.Added) > 0 ||
		len(diff.Removed) > 0 ||
		len(diff.Promoted) > 0 ||
		len(diff.DowngradeReasons) > 0 ||
		strings.TrimSpace(diff.CurrentAssetChange) != ""
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

func buildExplanationRelatedEntities(items []map[string]any) []model.StrategyExplanationRelatedEntity {
	result := make([]model.StrategyExplanationRelatedEntity, 0, len(items))
	for _, item := range items {
		entity := model.StrategyExplanationRelatedEntity{
			EntityType:  asString(item["entity_type"]),
			EntityKey:   asString(item["entity_key"]),
			Label:       asString(item["label"]),
			AssetDomain: asString(item["asset_domain"]),
			Tags:        stringSlice(item["tags"]),
			Meta:        mapValue(item["meta"]),
		}
		if entity.EntityType == "" && entity.EntityKey == "" && entity.Label == "" {
			continue
		}
		result = append(result, entity)
	}
	return result
}

func mergeExplanationRelatedEntities(groups ...[]model.StrategyExplanationRelatedEntity) []model.StrategyExplanationRelatedEntity {
	result := make([]model.StrategyExplanationRelatedEntity, 0)
	seen := make(map[string]struct{})
	for _, group := range groups {
		for _, item := range group {
			key := strings.TrimSpace(strings.Join([]string{item.EntityType, item.EntityKey, item.Label}, "|"))
			if key == "||" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func evidenceCardNoteByTitle(cards []model.StrategyExplanationEvidenceCard, title string) string {
	want := strings.TrimSpace(title)
	for _, card := range cards {
		if strings.TrimSpace(card.Title) == want {
			return firstNonEmpty(card.Note, card.Value)
		}
	}
	return ""
}

func buildFuturesSupplyChainNotes(
	relatedEntities []model.StrategyExplanationRelatedEntity,
	evidenceCards []model.StrategyExplanationEvidenceCard,
	inventorySummary string,
	structureSummary string,
) []string {
	notes := make([]string, 0, 6)
	seen := make(map[string]struct{})
	appendNote := func(text string) {
		text = strings.TrimSpace(text)
		if text == "" {
			return
		}
		if _, ok := seen[text]; ok {
			return
		}
		seen[text] = struct{}{}
		notes = append(notes, text)
	}

	for _, entity := range relatedEntities {
		label := firstNonEmpty(entity.Label, entity.EntityKey)
		switch strings.TrimSpace(entity.EntityType) {
		case "Commodity":
			appendNote("商品：" + label)
		case "SupplyChainNode":
			appendNote("商品链节点：" + label)
		case "SpreadPair", "Index":
			appendNote("结构联动：" + label)
		case "DeliveryPlace":
			appendNote("交割地：" + label)
		case "Warehouse":
			appendNote("仓库：" + label)
		case "Brand":
			appendNote("品牌：" + label)
		case "Grade":
			appendNote("等级：" + label)
		}
	}
	if len(notes) == 0 {
		appendNote(inventorySummary)
		appendNote(structureSummary)
		appendNote(evidenceCardNoteByTitle(evidenceCards, "库存画像"))
		appendNote(evidenceCardNoteByTitle(evidenceCards, "结构联动"))
	}
	if len(notes) > 6 {
		return notes[:6]
	}
	return notes
}

func buildExplanationMemoryFeedback(item map[string]any) model.StrategyExplanationMemoryFeedback {
	result := model.StrategyExplanationMemoryFeedback{
		Summary:        asString(item["summary"]),
		Suggestions:    stringSlice(item["suggestions"]),
		FailureSignals: stringSlice(item["failure_signals"]),
		Items:          []model.StrategyExplanationMemoryFeedbackItem{},
	}
	for _, raw := range sliceOfMaps(item["items"]) {
		entry := model.StrategyExplanationMemoryFeedbackItem{
			Title:      asString(raw["title"]),
			Level:      asString(raw["level"]),
			Detail:     asString(raw["detail"]),
			Suggestion: asString(raw["suggestion"]),
			Source:     asString(raw["source"]),
		}
		if entry.Title == "" && entry.Detail == "" {
			continue
		}
		result.Items = append(result.Items, entry)
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
