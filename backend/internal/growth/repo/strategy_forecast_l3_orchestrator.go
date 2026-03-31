package repo

import (
	"fmt"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
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

	report := buildStrategyForecastL3Report(run, pack, roles, now)
	logs = append(logs, newStrategyForecastL3Log(run.ID, "BUILD_REPORT", "SUCCESS", "structured report built", map[string]any{
		"report_id":        report.ID,
		"primary_scenario": report.PrimaryScenario,
	}, now))

	run.Status = model.StrategyForecastL3StatusSucceeded
	run.FailureReason = ""
	run.EngineKey = firstNonEmpty(run.EngineKey, model.StrategyForecastL3EngineLocalSynthesis)
	run.FinishedAt = now.Format(time.RFC3339)
	run.UpdatedAt = now.Format(time.RFC3339)
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
