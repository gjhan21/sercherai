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
	if strings.EqualFold(pack.TargetType, model.StrategyForecastL3TargetTypeFutures) {
		return []strategyForecastL3RoleResult{
			{Role: "SUPPLY_DEMAND", Stance: "CONSTRUCTIVE", Confidence: 0.66, Summary: firstNonEmpty(pack.CoreThesis, "Supply and demand remain aligned with the base case.")},
			{Role: "HEDGE", Stance: "NEUTRAL", Confidence: 0.58, Summary: "Hedging pressure is stable and not yet disruptive."},
			{Role: "SPEC_FLOW", Stance: "WATCH", Confidence: 0.61, Summary: firstNonEmpty(firstString(pack.RelatedHighlights), "Speculative flow still needs confirmation.")},
			{Role: "MACRO", Stance: "NEUTRAL", Confidence: 0.57, Summary: firstNonEmpty(pack.EvaluationSummary, "Macro conditions remain watchable, not decisive.")},
			{Role: "RISK", Stance: "CAUTION", Confidence: 0.64, Summary: firstNonEmpty(pack.RiskBoundary, "Risk boundary should stay visible throughout execution."), Veto: pack.L2Vetoed},
		}
	}
	return []strategyForecastL3RoleResult{
		{Role: "INDUSTRY", Stance: "BULLISH", Confidence: 0.68, Summary: firstNonEmpty(pack.CoreThesis, "Industry context still supports the current thesis.")},
		{Role: "FLOW", Stance: "CONSTRUCTIVE", Confidence: 0.63, Summary: firstNonEmpty(firstString(pack.RelatedHighlights), "Flow remains constructive but needs confirmation.")},
		{Role: "EVENT", Stance: "WATCH", Confidence: 0.60, Summary: firstNonEmpty(firstString(pack.HistoricalNotes), "Event path should remain under review.")},
		{Role: "MACRO", Stance: "NEUTRAL", Confidence: 0.56, Summary: firstNonEmpty(pack.EvaluationSummary, "Macro conditions are supportive but not decisive.")},
		{Role: "RISK", Stance: "CAUTION", Confidence: 0.65, Summary: firstNonEmpty(pack.RiskBoundary, "Risk boundary must stay front and center."), Veto: pack.L2Vetoed},
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
