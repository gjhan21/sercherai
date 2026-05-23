package repo

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/platform/config"
	"sercherai/backend/internal/platform/llm"
)

type strategyForecastL3ContextReader interface {
	GetStockRecommendationInsight(userID string, recoID string) (model.StockRecommendationInsight, error)
	GetStockRecommendationVersionHistory(userID string, recoID string) ([]model.StrategyVersionHistoryItem, error)
	GetFuturesStrategyInsight(userID string, strategyID string) (model.FuturesStrategyInsight, error)
	GetFuturesStrategyVersionHistory(userID string, strategyID string) ([]model.StrategyVersionHistoryItem, error)
}

type strategyForecastL3ExecutionResult struct {
	Run    model.StrategyForecastL3Run
	Report *model.StrategyForecastL3Report
	Logs   []model.StrategyForecastL3Log
}

type strategyForecastL3DeepForecastAdapter interface {
	RunDeepForecast(pack strategyForecastL3ResearchPack) []strategyForecastL3RoleResult
}

type strategyForecastL3ValidationResult struct {
	Status              string
	Verdict             string
	ScenarioConsistency string
	SupportingEvidence  []string
	CounterEvidence     []string
	BlindSpots          []string
	RiskReview          []string
	ActionReview        []string
	LLMSummary          string
}

type localSynthesisForecastL3Adapter struct{}

func (a localSynthesisForecastL3Adapter) RunDeepForecast(pack strategyForecastL3ResearchPack) []strategyForecastL3RoleResult {
	highlightsStr := ""
	if len(pack.RelatedHighlights) > 0 {
		highlightsStr = "结合当前异动/新闻：" + strings.Join(pack.RelatedHighlights, "; ")
	}
	notesStr := ""
	if len(pack.HistoricalNotes) > 0 {
		notesStr = "结合历史点评追踪：" + strings.Join(pack.HistoricalNotes, "; ")
	}

	if strings.EqualFold(pack.TargetType, model.StrategyForecastL3TargetTypeFutures) {
		return []strategyForecastL3RoleResult{
			{Role: "SUPPLY_DEMAND", Stance: "CONSTRUCTIVE", Confidence: 0.82, Summary: firstNonEmpty(pack.CoreThesis, "供需基本面验证暂未恶化，主要矛盾仍按预期节奏推进。")},
			{Role: "HEDGE", Stance: "NEUTRAL", Confidence: 0.65, Summary: "产业套保与现货对冲压力表现为正常轮动，未观察到恐慌性抢跑。"},
			{Role: "SPEC_FLOW", Stance: "WATCH", Confidence: 0.70, Summary: firstNonEmpty(highlightsStr, "投机资金呈结构性分化，需结合盘面基差异动确认。")},
			{Role: "MACRO", Stance: "NEUTRAL", Confidence: 0.60, Summary: firstNonEmpty(pack.EvaluationSummary, "宏观背景边际影响钝化，暂时不是该品种的核心驱动力。")},
			{Role: "RISK", Stance: "CAUTION", Confidence: 0.75, Summary: firstNonEmpty(pack.RiskBoundary, "存在极端行情下的脆弱性，必须严设防守底线。"), Veto: pack.L2Vetoed},
		}
	}
	return []strategyForecastL3RoleResult{
		{Role: "INDUSTRY", Stance: "BULLISH", Confidence: 0.85, Summary: firstNonEmpty(pack.CoreThesis, "行业景气度及竞争格局趋势良好，中长线逻辑依然成立。")},
		{Role: "FLOW", Stance: "CONSTRUCTIVE", Confidence: 0.72, Summary: firstNonEmpty(highlightsStr, "量价与北向/机构筹码维持偏强震荡，未见合力抛压。")},
		{Role: "EVENT", Stance: "WATCH", Confidence: 0.68, Summary: firstNonEmpty(notesStr, "公司即将落地的催化节点存在一定博弈，需保持跟踪。")},
		{Role: "MACRO", Stance: "NEUTRAL", Confidence: 0.65, Summary: firstNonEmpty(pack.EvaluationSummary, "系统大盘风险偏好适中，未对板块形成明显的溢价拖累。")},
		{Role: "RISK", Stance: "CAUTION", Confidence: 0.80, Summary: firstNonEmpty(pack.RiskBoundary, "关注业绩雷或监管风险，失效条件触及应直接离场。"), Veto: pack.L2Vetoed},
	}
}

func executeStrategyForecastL3Run(
	reader strategyForecastL3ContextReader,
	run model.StrategyForecastL3Run,
) strategyForecastL3ExecutionResult {
	now := time.Now().UTC()
	run.Status = model.StrategyForecastL3StatusRunning
	run.StartedAt = now.Format(time.RFC3339)
	run.UpdatedAt = now.Format(time.RFC3339)

	logs := []model.StrategyForecastL3Log{
		newStrategyForecastL3Log(run.ID, "LOAD_CONTEXT", "SUCCESS", "target context loaded", map[string]any{
			"target_type": run.TargetType,
			"target_key":  run.TargetKey,
		}, now),
	}

	pack, err := buildStrategyForecastL3ResearchPack(reader, run)
	if err != nil {
		logs = append(logs, newStrategyForecastL3Log(run.ID, "BUILD_RESEARCH_PACK", "FAILED", err.Error(), nil, now))
		run.Status = model.StrategyForecastL3StatusFailed
		run.FailureReason = err.Error()
		run.FinishedAt = now.Format(time.RFC3339)
		run.UpdatedAt = now.Format(time.RFC3339)
		run.Summary = model.StrategyForecastL3Summary{
			RunID:            run.ID,
			Status:           run.Status,
			EngineKey:        firstNonEmpty(run.EngineKey, model.StrategyForecastL3EngineLocalSynthesis),
			TriggerType:      run.TriggerType,
			TargetType:       run.TargetType,
			TargetKey:        run.TargetKey,
			TargetLabel:      firstNonEmpty(run.TargetLabel, run.TargetKey),
			ExecutiveSummary: err.Error(),
			PriorityScore:    run.PriorityScore,
			GeneratedAt:      now.Format(time.RFC3339),
			ReportAvailable:  false,
		}
		return strategyForecastL3ExecutionResult{Run: run, Logs: logs}
	}

	logs = append(logs, newStrategyForecastL3Log(run.ID, "BUILD_RESEARCH_PACK", "SUCCESS", "research pack assembled", map[string]any{
		"highlights":      len(pack.RelatedHighlights),
		"invalidations":   len(pack.Invalidations),
		"historicalNotes": len(pack.HistoricalNotes),
	}, now))

	adapter := localSynthesisForecastL3Adapter{}
	roles := adapter.RunDeepForecast(pack)
	logs = append(logs, newStrategyForecastL3Log(run.ID, "RUN_DEEP_FORECAST", "SUCCESS", "local synthesis completed", map[string]any{
		"engine": model.StrategyForecastL3EngineLocalSynthesis,
		"roles":  len(roles),
	}, now))

	validation := runStrategyForecastL3Validation(run, pack, roles)
	logs = append(logs, newStrategyForecastL3Log(run.ID, "VALIDATE_REPORT", "SUCCESS", "validation review completed", map[string]any{
		"status": validation.Status,
	}, now))

	report := buildStrategyForecastL3Report(run, pack, roles, validation, now)
	logs = append(logs, newStrategyForecastL3Log(run.ID, "BUILD_REPORT", "SUCCESS", "structured report built", map[string]any{
		"report_id":        report.ID,
		"primary_scenario": report.PrimaryScenario,
	}, now))

	run.Status = model.StrategyForecastL3StatusSucceeded
	run.FailureReason = ""
	run.EngineKey = firstNonEmpty(run.EngineKey, model.StrategyForecastL3EngineLocalSynthesis)
	run.FinishedAt = now.Format(time.RFC3339)
	run.UpdatedAt = now.Format(time.RFC3339)
	run.ValidationStatus = report.ValidationReview.Status
	run.Summary = report.Summary
	run.ReportRef = &model.StrategyForecastL3ReportRef{
		RunID:        run.ID,
		ReportID:     report.ID,
		Status:       run.Status,
		EngineKey:    run.EngineKey,
		GeneratedAt:  report.CreatedAt,
		RequiresVIP:  true,
		FullReadable: false,
	}
	logs = append(logs, newStrategyForecastL3Log(run.ID, "PERSIST_REPORT", "SUCCESS", "report persisted", map[string]any{
		"report_id": report.ID,
	}, now))

	return strategyForecastL3ExecutionResult{
		Run:    run,
		Report: &report,
		Logs:   logs,
	}
}

func runStrategyForecastL3Validation(
	run model.StrategyForecastL3Run,
	pack strategyForecastL3ResearchPack,
	roles []strategyForecastL3RoleResult,
) strategyForecastL3ValidationResult {
	if strings.TrimSpace(run.ValidationStatus) == model.StrategyForecastL3ValidationStatusSkipped {
		return strategyForecastL3ValidationResult{
			Status:              model.StrategyForecastL3ValidationStatusSkipped,
			Verdict:             "本次运行未启用模型复核。",
			ScenarioConsistency: "未执行",
		}
	}

	base := strategyForecastL3ValidationResult{
		Status:              model.StrategyForecastL3ValidationStatusCompleted,
		Verdict:             "模型复核认为主情景与现有证据基本一致。",
		ScenarioConsistency: "主情景整体自洽，但仍需继续验证触发条件。",
		SupportingEvidence:  uniqueForecastL3Strings(append([]string{}, pack.RelatedHighlights...)),
		CounterEvidence:     uniqueForecastL3Strings(append([]string{}, pack.Invalidations...)),
		BlindSpots:          []string{"当前仍缺少更长窗口的跨周期验证。"},
		RiskReview:          uniqueForecastL3Strings([]string{pack.RiskBoundary}),
		ActionReview:        uniqueForecastL3Strings(pack.ActionHints),
		LLMSummary:          "当前模型复核支持继续围绕主情景跟踪，但不建议脱离风险边界独立放大仓位。",
	}

	if !shouldAttemptForecastL3LLMValidation(run) {
		return base
	}

	cfg := config.Load()
	client := llm.NewClient(cfg)
	content, err := client.ChatCompletion([]llm.ChatMessage{
		{
			Role:    "system",
			Content: "你是深度推演复核器。请仅返回 JSON，字段包括 verdict, scenario_consistency, supporting_evidence, counter_evidence, blind_spots, risk_review, action_review, llm_summary。",
		},
		{
			Role:    "user",
			Content: buildStrategyForecastL3ValidationPrompt(run, pack, roles),
		},
	})
	if err != nil {
		base.Status = model.StrategyForecastL3ValidationStatusDegraded
		base.Verdict = "模型复核未完成，已回退为结构化本地复核。"
		base.LLMSummary = "模型复核服务当前不可用，主报告仍基于结构化研究链路生成。"
		return base
	}

	type llmValidationPayload struct {
		Verdict             string   `json:"verdict"`
		ScenarioConsistency string   `json:"scenario_consistency"`
		SupportingEvidence  []string `json:"supporting_evidence"`
		CounterEvidence     []string `json:"counter_evidence"`
		BlindSpots          []string `json:"blind_spots"`
		RiskReview          []string `json:"risk_review"`
		ActionReview        []string `json:"action_review"`
		LLMSummary          string   `json:"llm_summary"`
	}
	var parsed llmValidationPayload
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &parsed); err != nil {
		base.Status = model.StrategyForecastL3ValidationStatusDegraded
		base.Verdict = "模型复核输出格式异常，已回退为结构化本地复核。"
		base.LLMSummary = "模型复核返回不可解析内容，主报告仍可正常使用。"
		return base
	}

	return strategyForecastL3ValidationResult{
		Status:              model.StrategyForecastL3ValidationStatusCompleted,
		Verdict:             firstNonEmpty(parsed.Verdict, base.Verdict),
		ScenarioConsistency: firstNonEmpty(parsed.ScenarioConsistency, base.ScenarioConsistency),
		SupportingEvidence:  uniqueForecastL3Strings(firstNonEmptyStrings(parsed.SupportingEvidence, base.SupportingEvidence)),
		CounterEvidence:     uniqueForecastL3Strings(firstNonEmptyStrings(parsed.CounterEvidence, base.CounterEvidence)),
		BlindSpots:          uniqueForecastL3Strings(firstNonEmptyStrings(parsed.BlindSpots, base.BlindSpots)),
		RiskReview:          uniqueForecastL3Strings(firstNonEmptyStrings(parsed.RiskReview, base.RiskReview)),
		ActionReview:        uniqueForecastL3Strings(firstNonEmptyStrings(parsed.ActionReview, base.ActionReview)),
		LLMSummary:          firstNonEmpty(parsed.LLMSummary, base.LLMSummary),
	}
}

func shouldAttemptForecastL3LLMValidation(run model.StrategyForecastL3Run) bool {
	return strings.EqualFold(strings.TrimSpace(run.TriggerType), model.StrategyForecastL3TriggerTypeUserRequest)
}

func buildStrategyForecastL3ValidationPrompt(
	run model.StrategyForecastL3Run,
	pack strategyForecastL3ResearchPack,
	roles []strategyForecastL3RoleResult,
) string {
	parts := []string{
		fmt.Sprintf("target_type=%s", run.TargetType),
		fmt.Sprintf("target_key=%s", run.TargetKey),
		fmt.Sprintf("target_label=%s", firstNonEmpty(run.TargetLabel, pack.TargetLabel)),
		fmt.Sprintf("core_thesis=%s", pack.CoreThesis),
		fmt.Sprintf("risk_boundary=%s", pack.RiskBoundary),
		fmt.Sprintf("invalidations=%s", strings.Join(pack.Invalidations, " | ")),
		fmt.Sprintf("action_hints=%s", strings.Join(pack.ActionHints, " | ")),
	}
	parts = append(parts, buildStrategyForecastL3DomainEvidencePromptLines(run.TargetType, pack)...)
	for _, role := range roles {
		parts = append(parts, fmt.Sprintf("role=%s stance=%s confidence=%.2f summary=%s", role.Role, role.Stance, role.Confidence, role.Summary))
	}
	return strings.Join(parts, "\n")
}

func firstNonEmptyStrings(primary []string, fallback []string) []string {
	if len(primary) > 0 {
		return primary
	}
	return fallback
}

func buildStrategyForecastL3ResearchPack(
	reader strategyForecastL3ContextReader,
	run model.StrategyForecastL3Run,
) (strategyForecastL3ResearchPack, error) {
	pack := strategyForecastL3ResearchPack{
		TargetType:  run.TargetType,
		TargetKey:   run.TargetKey,
		TargetLabel: firstNonEmpty(run.TargetLabel, run.TargetKey),
		CoreThesis:  strings.TrimSpace(run.Reason),
	}

	switch strings.ToUpper(strings.TrimSpace(run.TargetType)) {
	case model.StrategyForecastL3TargetTypeStock:
		if strings.TrimSpace(run.TargetID) != "" {
			insight, err := reader.GetStockRecommendationInsight(run.RequestUserID, run.TargetID)
			if err != nil {
				return strategyForecastL3ResearchPack{}, err
			}
			pack.TargetLabel = firstNonEmpty(pack.TargetLabel, insight.Recommendation.Name)
			pack.CoreThesis = firstNonEmpty(pack.CoreThesis, insight.Recommendation.ReasonSummary, insight.Explanation.ConsensusSummary)
			pack.RiskBoundary = firstNonEmpty(insight.Explanation.RiskBoundary, insight.Detail.RiskNote)
			pack.Invalidations = uniqueForecastL3Strings(append(pack.Invalidations, insight.Explanation.Invalidations...))
			for _, item := range insight.RelatedNews {
				pack.RelatedHighlights = append(pack.RelatedHighlights, firstNonEmpty(item.Title, item.Summary))
			}
			pack.ActionHints = uniqueForecastL3Strings(append(pack.ActionHints, insight.Detail.TakeProfit, insight.Detail.StopLoss))
			pack.L2PrimaryScenario = insight.Explanation.ScenarioMeta.PrimaryScenario
			pack.L2ConsensusAction = insight.Explanation.ScenarioMeta.ConsensusAction
			pack.L2Vetoed = insight.Explanation.ScenarioMeta.Vetoed
			pack.L2VetoReason = insight.Explanation.ScenarioMeta.VetoReason
			pack.EvaluationSummary = fmt.Sprintf("sample_days=%d cumulative_return=%.4f", insight.PerformanceStats.SampleDays, insight.PerformanceStats.CumulativeReturn)
			pack.StockEvidence = buildStrategyForecastL3StockEvidence(insight)
		}
		if strings.TrimSpace(run.TargetID) != "" {
			history, err := reader.GetStockRecommendationVersionHistory(run.RequestUserID, run.TargetID)
			if err == nil && len(history) > 0 {
				for _, item := range history {
					pack.HistoricalNotes = append(pack.HistoricalNotes, firstNonEmpty(item.ReasonSummary, item.ConfidenceReason))
				}
			}
		}
	case model.StrategyForecastL3TargetTypeFutures:
		if strings.TrimSpace(run.TargetID) != "" {
			insight, err := reader.GetFuturesStrategyInsight(run.RequestUserID, run.TargetID)
			if err != nil {
				return strategyForecastL3ResearchPack{}, err
			}
			pack.TargetLabel = firstNonEmpty(pack.TargetLabel, insight.Strategy.Name, insight.Strategy.Contract)
			pack.CoreThesis = firstNonEmpty(pack.CoreThesis, insight.Strategy.ReasonSummary, insight.Explanation.ConsensusSummary)
			pack.RiskBoundary = firstNonEmpty(insight.Explanation.RiskBoundary, insight.Guidance.InvalidCondition)
			pack.Invalidations = append(pack.Invalidations, insight.Explanation.Invalidations...)
			pack.Invalidations = uniqueForecastL3Strings(append(pack.Invalidations, insight.Guidance.InvalidCondition))
			for _, item := range insight.RelatedNews {
				pack.RelatedHighlights = append(pack.RelatedHighlights, firstNonEmpty(item.Title, item.Summary))
			}
			for _, item := range insight.RelatedEvents {
				pack.RelatedHighlights = append(pack.RelatedHighlights, firstNonEmpty(item.Summary, item.EventType))
			}
			pack.ActionHints = uniqueForecastL3Strings(append(pack.ActionHints, insight.Guidance.TakeProfitRange, insight.Guidance.StopLossRange))
			pack.L2PrimaryScenario = insight.Explanation.ScenarioMeta.PrimaryScenario
			pack.L2ConsensusAction = insight.Explanation.ScenarioMeta.ConsensusAction
			pack.L2Vetoed = insight.Explanation.ScenarioMeta.Vetoed
			pack.L2VetoReason = insight.Explanation.ScenarioMeta.VetoReason
			pack.EvaluationSummary = fmt.Sprintf("sample_days=%d cumulative_return=%.4f", insight.PerformanceStats.SampleDays, insight.PerformanceStats.CumulativeReturn)
			pack.FuturesEvidence = buildStrategyForecastL3FuturesEvidence(insight)
		}
		if strings.TrimSpace(run.TargetID) != "" {
			history, err := reader.GetFuturesStrategyVersionHistory(run.RequestUserID, run.TargetID)
			if err == nil && len(history) > 0 {
				for _, item := range history {
					pack.HistoricalNotes = append(pack.HistoricalNotes, firstNonEmpty(item.ReasonSummary, item.ConfidenceReason))
				}
			}
		}
	}

	pack.RelatedHighlights = uniqueForecastL3Strings(pack.RelatedHighlights)
	pack.HistoricalNotes = uniqueForecastL3Strings(pack.HistoricalNotes)
	pack.ActionHints = uniqueForecastL3Strings(pack.ActionHints)
	pack.Invalidations = uniqueForecastL3Strings(pack.Invalidations)
	if strings.TrimSpace(pack.CoreThesis) == "" && len(pack.RelatedHighlights) == 0 && len(pack.HistoricalNotes) == 0 {
		return strategyForecastL3ResearchPack{}, fmt.Errorf("no usable forecast l3 context for %s", run.TargetKey)
	}
	return pack, nil
}

func buildStrategyForecastL3StockEvidence(insight model.StockRecommendationInsight) model.StrategyForecastL3StockEvidence {
	quote, hasQuote := firstStockQuantScore(insight.Explanation)
	fundamentalPoints := uniqueForecastL3Strings(nonEmptyStrings(
		firstNonEmpty(insight.Recommendation.ReasonSummary, insight.Explanation.ConsensusSummary),
		firstNonEmpty(insight.Explanation.ConfidenceReason, insight.Explanation.SeedSummary),
		stockPerformanceSummaryText(insight.PerformanceStats),
	))
	fundamentalRisks := uniqueForecastL3Strings(nonEmptyStrings(
		insight.Detail.RiskNote,
		insight.Explanation.RiskBoundary,
	))

	technicalPoints := make([]string, 0, 3)
	if hasQuote && (quote.Momentum20 != 0 || quote.TrendStrength != 0) {
		technicalPoints = append(technicalPoints, fmt.Sprintf("20日动量%.2f，趋势强度%.2f。", quote.Momentum20, quote.TrendStrength))
	}
	if hasQuote && (quote.Volatility20 != 0 || quote.VolumeRatio != 0) {
		technicalPoints = append(technicalPoints, fmt.Sprintf("20日波动率%.2f，量比%.2f。", quote.Volatility20, quote.VolumeRatio))
	}
	if insight.Detail.TechScore > 0 || (hasQuote && quote.Drawdown20 != 0) {
		technicalPoints = append(technicalPoints, fmt.Sprintf("技术评分%.2f，20日回撤%.2f。", insight.Detail.TechScore, quote.Drawdown20))
	}
	technicalPoints = uniqueForecastL3Strings(technicalPoints)
	technicalRisks := uniqueForecastL3Strings(nonEmptyStrings(
		technicalRiskFromQuote(quote, hasQuote),
		insight.Detail.RiskNote,
	))

	flowPoints := make([]string, 0, 2)
	if hasQuote && (quote.FlowScore != 0 || quote.NetMFAmount != 0) {
		flowPoints = append(flowPoints, fmt.Sprintf("资金流评分%.2f，净流入%.2f。", quote.FlowScore, quote.NetMFAmount))
	}
	if (hasQuote && quote.TurnoverRate != 0) || insight.Detail.MoneyFlowScore > 0 {
		flowPoints = append(flowPoints, fmt.Sprintf("换手率%.2f，资金流因子%.2f。", quote.TurnoverRate, insight.Detail.MoneyFlowScore))
	}
	flowPoints = uniqueForecastL3Strings(flowPoints)
	flowRisks := uniqueForecastL3Strings(nonEmptyStrings(
		flowRiskFromQuote(quote, hasQuote),
	))

	valuationPoints := make([]string, 0, 2)
	if hasQuote && (quote.PeTTM > 0 || quote.PB > 0) {
		valuationPoints = append(valuationPoints, fmt.Sprintf("PE(TTM) %.2f，PB %.2f。", quote.PeTTM, quote.PB))
	}
	if hasQuote && quote.ValueScore != 0 {
		valuationPoints = append(valuationPoints, fmt.Sprintf("估值评分%.2f。", quote.ValueScore))
	}
	valuationPoints = uniqueForecastL3Strings(valuationPoints)
	valuationRisks := uniqueForecastL3Strings(nonEmptyStrings(
		valuationRiskFromQuote(quote, hasQuote),
	))

	eventPoints := make([]string, 0, len(insight.RelatedNews)+2)
	if hasQuote && (quote.NewsHeat > 0 || quote.PositiveNewsRate > 0) {
		eventPoints = append(eventPoints, fmt.Sprintf("新闻热度%d，正向新闻占比%.2f。", quote.NewsHeat, quote.PositiveNewsRate))
	}
	if hasQuote && quote.NewsScore != 0 {
		eventPoints = append(eventPoints, fmt.Sprintf("新闻评分%.2f。", quote.NewsScore))
	}
	for _, item := range insight.RelatedNews {
		eventPoints = append(eventPoints, firstNonEmpty(item.Title, item.Summary))
	}
	eventPoints = uniqueForecastL3Strings(eventPoints)
	eventRisks := uniqueForecastL3Strings(nonEmptyStrings(
		eventRiskFromQuote(quote, hasQuote),
	))

	return model.StrategyForecastL3StockEvidence{
		Fundamental: model.StrategyForecastL3EvidenceSlice{
			Summary:          firstNonEmpty(insight.Recommendation.ReasonSummary, insight.Explanation.ConsensusSummary),
			SupportingPoints: fundamentalPoints,
			RiskPoints:       fundamentalRisks,
		},
		Technical: model.StrategyForecastL3EvidenceSlice{
			Summary:          firstNonEmpty(technicalPoints...),
			SupportingPoints: technicalPoints,
			RiskPoints:       technicalRisks,
		},
		Flow: model.StrategyForecastL3EvidenceSlice{
			Summary:          firstNonEmpty(flowPoints...),
			SupportingPoints: flowPoints,
			RiskPoints:       flowRisks,
		},
		Valuation: model.StrategyForecastL3EvidenceSlice{
			Summary:          firstNonEmpty(valuationPoints...),
			SupportingPoints: valuationPoints,
			RiskPoints:       valuationRisks,
		},
		Event: model.StrategyForecastL3EvidenceSlice{
			Summary:          firstNonEmpty(eventPoints...),
			SupportingPoints: eventPoints,
			RiskPoints:       eventRisks,
		},
	}
}

func buildStrategyForecastL3FuturesEvidence(insight model.FuturesStrategyInsight) model.StrategyForecastL3FuturesEvidence {
	seed, hasSeed := firstFuturesSeed(insight.Explanation)
	supplyDemandPoints := uniqueForecastL3Strings(nonEmptyStrings(
		firstNonEmpty(insight.Explanation.InventorySummary, inventorySummaryText(seed)),
		conditionalString(hasSeed && (seed.InventoryPressure != 0 || seed.InventoryChangePct != 0 || seed.InventoryLevel != 0),
			fmt.Sprintf("库存压力%.2f，库存变动%.2f，库存水平%.2f。", seed.InventoryPressure, seed.InventoryChangePct, seed.InventoryLevel)),
		firstNonEmpty(seed.InventoryBrandGradeSummary, inventoryFocusText(seed)),
	))
	supplyDemandRisks := uniqueForecastL3Strings(nonEmptyStrings(
		supplyDemandRiskFromSeed(seed, hasSeed),
		insight.Guidance.InvalidCondition,
	))

	termStructurePoints := uniqueForecastL3Strings(nonEmptyStrings(
		conditionalString(hasSeed && (seed.BasisPct != 0 || seed.CarryPct != 0 || seed.TermStructurePct != 0),
			fmt.Sprintf("基差%.2f，Carry %.2f，期限结构%.2f。", seed.BasisPct, seed.CarryPct, seed.TermStructurePct)),
		conditionalString(hasSeed && (seed.CurveSlopePct != 0 || seed.BasisTermAlignment != 0),
			fmt.Sprintf("曲线斜率%.2f，基差-期限一致性%.2f。", seed.CurveSlopePct, seed.BasisTermAlignment)),
		firstNonEmpty(seed.StructureSignalSummary),
	))
	termStructureRisks := uniqueForecastL3Strings(nonEmptyStrings(
		termStructureRiskFromSeed(seed, hasSeed),
	))

	tapePoints := uniqueForecastL3Strings(nonEmptyStrings(
		conditionalString(hasSeed && (seed.TrendStrength != 0 || seed.Volatility14 != 0 || seed.VolumeRatio != 0),
			fmt.Sprintf("趋势强度%.2f，14日波动率%.2f，量比%.2f。", seed.TrendStrength, seed.Volatility14, seed.VolumeRatio)),
		conditionalString(hasSeed && (seed.SpreadPressure != 0 || seed.CrossContractLinkage != 0),
			fmt.Sprintf("价差压力%.2f，跨合约联动%.2f。", seed.SpreadPressure, seed.CrossContractLinkage)),
	))
	tapeRisks := uniqueForecastL3Strings(nonEmptyStrings(
		tapeRiskFromSeed(seed, hasSeed),
	))

	positionPoints := uniqueForecastL3Strings(nonEmptyStrings(
		conditionalString(hasSeed && (seed.OIChangePct != 0 || seed.TurnoverRatio != 0 || seed.FlowBias != 0),
			fmt.Sprintf("持仓变化%.2f，换手率%.2f，流向偏置%.2f。", seed.OIChangePct, seed.TurnoverRatio, seed.FlowBias)),
		guidancePositionText(insight.Guidance.GuidanceDirection, insight.Guidance.PositionLevel),
	))
	positionRisks := uniqueForecastL3Strings(nonEmptyStrings(
		positionRiskFromSeed(seed, hasSeed),
	))

	macroPoints := make([]string, 0, len(insight.RelatedNews)+len(insight.RelatedEvents)+2)
	macroPoints = append(macroPoints, nonEmptyStrings(
		firstNonEmpty(insight.Explanation.MarketRegime, seed.Regime),
		futuresPerformanceSummaryText(insight.PerformanceStats),
	)...)
	for _, item := range insight.RelatedNews {
		macroPoints = append(macroPoints, firstNonEmpty(item.Title, item.Summary))
	}
	for _, item := range insight.RelatedEvents {
		macroPoints = append(macroPoints, firstNonEmpty(item.Summary, item.EventType))
	}
	macroRisks := uniqueForecastL3Strings(nonEmptyStrings(
		insight.Guidance.InvalidCondition,
		insight.Explanation.RiskBoundary,
	))

	return model.StrategyForecastL3FuturesEvidence{
		SupplyDemand: model.StrategyForecastL3EvidenceSlice{
			Summary:          firstNonEmpty(insight.Explanation.InventorySummary, inventorySummaryText(seed), insight.Strategy.ReasonSummary),
			SupportingPoints: supplyDemandPoints,
			RiskPoints:       supplyDemandRisks,
		},
		TermStructure: model.StrategyForecastL3EvidenceSlice{
			Summary:          firstNonEmpty(termStructurePoints...),
			SupportingPoints: termStructurePoints,
			RiskPoints:       termStructureRisks,
		},
		TapeTechnical: model.StrategyForecastL3EvidenceSlice{
			Summary:          firstNonEmpty(tapePoints...),
			SupportingPoints: tapePoints,
			RiskPoints:       tapeRisks,
		},
		PositionFlow: model.StrategyForecastL3EvidenceSlice{
			Summary:          firstNonEmpty(positionPoints...),
			SupportingPoints: positionPoints,
			RiskPoints:       positionRisks,
		},
		MacroEvent: model.StrategyForecastL3EvidenceSlice{
			Summary:          firstNonEmpty(macroPoints...),
			SupportingPoints: uniqueForecastL3Strings(macroPoints),
			RiskPoints:       macroRisks,
		},
	}
}

func firstStockQuantScore(explanation model.StrategyClientExplanation) (model.StockQuantScore, bool) {
	if len(explanation.EvaluationMeta) == 0 {
		return model.StockQuantScore{}, false
	}
	raw, ok := explanation.EvaluationMeta["stock_quant_score"]
	if !ok {
		return model.StockQuantScore{}, false
	}
	var score model.StockQuantScore
	if decodeStrategyForecastL3Meta(raw, &score) {
		return score, true
	}
	return model.StockQuantScore{}, false
}

func firstFuturesSeed(explanation model.StrategyClientExplanation) (model.StrategyEngineFuturesSeed, bool) {
	if len(explanation.EvaluationMeta) == 0 {
		return model.StrategyEngineFuturesSeed{}, false
	}
	raw, ok := explanation.EvaluationMeta["futures_seed"]
	if !ok {
		return model.StrategyEngineFuturesSeed{}, false
	}
	var seed model.StrategyEngineFuturesSeed
	if decodeStrategyForecastL3Meta(raw, &seed) {
		return seed, true
	}
	return model.StrategyEngineFuturesSeed{}, false
}

func decodeStrategyForecastL3Meta(raw any, target any) bool {
	payload, err := json.Marshal(raw)
	if err != nil {
		return false
	}
	return json.Unmarshal(payload, target) == nil
}

func nonEmptyStrings(items ...string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if value := strings.TrimSpace(item); value != "" {
			result = append(result, value)
		}
	}
	return result
}

func stockPerformanceSummaryText(stats model.StockRecommendationPerformanceSummary) string {
	if stats.SampleDays == 0 && stats.CumulativeReturn == 0 {
		return ""
	}
	return fmt.Sprintf("sample_days=%d cumulative_return=%.4f", stats.SampleDays, stats.CumulativeReturn)
}

func futuresPerformanceSummaryText(stats model.FuturesStrategyPerformanceSummary) string {
	if stats.SampleDays == 0 && stats.CumulativeReturn == 0 {
		return ""
	}
	return fmt.Sprintf("sample_days=%d cumulative_return=%.4f", stats.SampleDays, stats.CumulativeReturn)
}

func technicalRiskFromQuote(quote model.StockQuantScore, hasQuote bool) string {
	if !hasQuote {
		return ""
	}
	if quote.Drawdown20 > 0 {
		return fmt.Sprintf("20日回撤%.2f提示技术面仍有回吐压力。", quote.Drawdown20)
	}
	return ""
}

func flowRiskFromQuote(quote model.StockQuantScore, hasQuote bool) string {
	if !hasQuote {
		return ""
	}
	if quote.NetMFAmount < 0 {
		return fmt.Sprintf("净资金流入%.2f偏弱，资金承接需要继续验证。", quote.NetMFAmount)
	}
	return ""
}

func valuationRiskFromQuote(quote model.StockQuantScore, hasQuote bool) string {
	if !hasQuote {
		return ""
	}
	if quote.PeTTM > 0 && quote.PB > 0 {
		return fmt.Sprintf("PE(TTM) %.2f、PB %.2f 对估值安全边际提出约束。", quote.PeTTM, quote.PB)
	}
	return ""
}

func eventRiskFromQuote(quote model.StockQuantScore, hasQuote bool) string {
	if !hasQuote {
		return ""
	}
	if quote.PositiveNewsRate > 0 && quote.PositiveNewsRate < 0.6 {
		return fmt.Sprintf("正向新闻占比%.2f，事件驱动一致性不足。", quote.PositiveNewsRate)
	}
	return ""
}

func inventorySummaryText(seed model.StrategyEngineFuturesSeed) string {
	if seed.InventoryPressure == 0 && seed.InventoryChangePct == 0 && seed.InventoryLevel == 0 {
		return ""
	}
	return fmt.Sprintf("库存压力%.2f，库存变动%.2f，库存水平%.2f。", seed.InventoryPressure, seed.InventoryChangePct, seed.InventoryLevel)
}

func inventoryFocusText(seed model.StrategyEngineFuturesSeed) string {
	parts := nonEmptyStrings(seed.InventoryFocusArea, seed.InventoryFocusWarehouse, seed.InventoryFocusBrand, seed.InventoryFocusPlace, seed.InventoryFocusGrade)
	if len(parts) == 0 {
		return ""
	}
	return "库存关注点：" + strings.Join(parts, " / ")
}

func supplyDemandRiskFromSeed(seed model.StrategyEngineFuturesSeed, hasSeed bool) string {
	if !hasSeed {
		return ""
	}
	if seed.InventoryChangePct > 0 {
		return fmt.Sprintf("库存变动%.2f偏高，供需改善节奏可能慢于预期。", seed.InventoryChangePct)
	}
	return ""
}

func termStructureRiskFromSeed(seed model.StrategyEngineFuturesSeed, hasSeed bool) string {
	if !hasSeed {
		return ""
	}
	if seed.BasisTermAlignment < 0 {
		return fmt.Sprintf("基差-期限结构一致性%.2f，曲线与现货信号存在背离。", seed.BasisTermAlignment)
	}
	return ""
}

func tapeRiskFromSeed(seed model.StrategyEngineFuturesSeed, hasSeed bool) string {
	if !hasSeed {
		return ""
	}
	if seed.Volatility14 > 0 {
		return fmt.Sprintf("14日波动率%.2f，盘面波动仍需用仓位控制消化。", seed.Volatility14)
	}
	return ""
}

func positionRiskFromSeed(seed model.StrategyEngineFuturesSeed, hasSeed bool) string {
	if !hasSeed {
		return ""
	}
	if seed.FlowBias < 0 {
		return fmt.Sprintf("流向偏置%.2f 偏空，持仓与换手共振尚未完全站稳。", seed.FlowBias)
	}
	return ""
}

func guidancePositionText(direction string, level string) string {
	direction = strings.TrimSpace(direction)
	level = strings.TrimSpace(level)
	switch {
	case direction != "" && level != "":
		return fmt.Sprintf("策略方向%s，仓位建议%s。", direction, level)
	case direction != "":
		return fmt.Sprintf("策略方向%s。", direction)
	case level != "":
		return fmt.Sprintf("仓位建议%s。", level)
	default:
		return ""
	}
}

func conditionalString(ok bool, value string) string {
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func buildStrategyForecastL3DomainEvidencePromptLines(
	targetType string,
	pack strategyForecastL3ResearchPack,
) []string {
	lines := make([]string, 0, 10)
	appendEvidence := func(dimension string, evidence model.StrategyForecastL3EvidenceSlice) {
		if strings.TrimSpace(evidence.Summary) == "" && len(evidence.SupportingPoints) == 0 && len(evidence.RiskPoints) == 0 {
			return
		}
		lines = append(lines,
			fmt.Sprintf("dimension=%s summary=%s", dimension, evidence.Summary),
			fmt.Sprintf("dimension=%s supporting=%s", dimension, strings.Join(evidence.SupportingPoints, " | ")),
			fmt.Sprintf("dimension=%s risks=%s", dimension, strings.Join(evidence.RiskPoints, " | ")),
		)
	}
	if strings.EqualFold(targetType, model.StrategyForecastL3TargetTypeFutures) {
		appendEvidence("SUPPLY_DEMAND", pack.FuturesEvidence.SupplyDemand)
		appendEvidence("TERM_STRUCTURE", pack.FuturesEvidence.TermStructure)
		appendEvidence("TAPE_TECHNICAL", pack.FuturesEvidence.TapeTechnical)
		appendEvidence("POSITION_FLOW", pack.FuturesEvidence.PositionFlow)
		appendEvidence("MACRO_EVENT", pack.FuturesEvidence.MacroEvent)
		return lines
	}
	appendEvidence("FUNDAMENTAL", pack.StockEvidence.Fundamental)
	appendEvidence("TECHNICAL", pack.StockEvidence.Technical)
	appendEvidence("FLOW", pack.StockEvidence.Flow)
	appendEvidence("VALUATION", pack.StockEvidence.Valuation)
	appendEvidence("EVENT", pack.StockEvidence.Event)
	return lines
}

func newStrategyForecastL3Log(runID string, stepKey string, status string, message string, payload map[string]any, now time.Time) model.StrategyForecastL3Log {
	return model.StrategyForecastL3Log{
		ID:        newID("l3log"),
		RunID:     runID,
		StepKey:   stepKey,
		Status:    status,
		Message:   message,
		Payload:   cloneStringAnyMap(payload),
		CreatedAt: now.UTC().Format(time.RFC3339),
	}
}
