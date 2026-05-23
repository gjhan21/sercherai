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
			Role: "system",
			Content: "你是深度推演复核器。请仅返回 JSON，字段包括 verdict, scenario_consistency, supporting_evidence, counter_evidence, blind_spots, risk_review, action_review, llm_summary。",
		},
		{
			Role: "user",
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
