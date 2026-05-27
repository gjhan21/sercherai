package repo

import (
	"strings"
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestBuildStrategyForecastL3ResearchPackIncludesStockDomainEvidence(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	run, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:    model.StrategyForecastL3TargetTypeStock,
		TargetID:      "sr_001",
		TargetKey:     "600519.SH",
		TargetLabel:   "贵州茅台",
		TriggerType:   model.StrategyForecastL3TriggerTypeAdminManual,
		RequestUserID: "admin_001",
		Reason:        "manual deep forecast",
	})
	if err != nil {
		t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
	}

	pack, err := buildStrategyForecastL3ResearchPack(repo, run)
	if err != nil {
		t.Fatalf("buildStrategyForecastL3ResearchPack() error = %v", err)
	}

	if pack.StockEvidence.Fundamental.Summary == "" {
		t.Fatalf("expected stock fundamental evidence summary, got %+v", pack.StockEvidence)
	}
	if len(pack.StockEvidence.Technical.SupportingPoints) == 0 {
		t.Fatalf("expected stock technical evidence supporting points, got %+v", pack.StockEvidence.Technical)
	}
	if len(pack.StockEvidence.Flow.SupportingPoints) == 0 {
		t.Fatalf("expected stock flow evidence supporting points, got %+v", pack.StockEvidence.Flow)
	}
	if len(pack.StockEvidence.Valuation.SupportingPoints) == 0 {
		t.Fatalf("expected stock valuation evidence supporting points, got %+v", pack.StockEvidence.Valuation)
	}
	if len(pack.StockEvidence.Event.SupportingPoints) == 0 {
		t.Fatalf("expected stock event evidence supporting points, got %+v", pack.StockEvidence.Event)
	}
}

func TestBuildStrategyForecastL3ResearchPackPrefersEvidenceOverInternalHandoffReason(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	run, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:    model.StrategyForecastL3TargetTypeStock,
		TargetID:      "sr_001",
		TargetKey:     "600519.SH",
		TargetLabel:   "贵州茅台",
		TriggerType:   model.StrategyForecastL3TriggerTypeUserRequest,
		RequestUserID: "user_001",
		Source:        "RECOMMENDATION",
		SourcePath:    "/recommendations/sr_001",
		Reason:        "from strategies focused handoff",
	})
	if err != nil {
		t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
	}

	pack, err := buildStrategyForecastL3ResearchPack(repo, run)
	if err != nil {
		t.Fatalf("buildStrategyForecastL3ResearchPack() error = %v", err)
	}

	if strings.Contains(strings.ToLower(pack.CoreThesis), "focused handoff") {
		t.Fatalf("expected internal handoff reason to be excluded from core thesis, got %q", pack.CoreThesis)
	}
	if !strings.Contains(pack.CoreThesis, "基本面") {
		t.Fatalf("expected evidence-backed thesis, got %q", pack.CoreThesis)
	}
}

func TestBuildStrategyForecastL3ResearchPackIncludesFuturesDomainEvidence(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	run, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:    model.StrategyForecastL3TargetTypeFutures,
		TargetID:      "fs_001",
		TargetKey:     "IF2603",
		TargetLabel:   "股指趋势跟踪",
		TriggerType:   model.StrategyForecastL3TriggerTypeAdminManual,
		RequestUserID: "admin_001",
		Reason:        "manual deep forecast",
	})
	if err != nil {
		t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
	}

	pack, err := buildStrategyForecastL3ResearchPack(repo, run)
	if err != nil {
		t.Fatalf("buildStrategyForecastL3ResearchPack() error = %v", err)
	}

	if pack.FuturesEvidence.SupplyDemand.Summary == "" {
		t.Fatalf("expected futures supply-demand evidence summary, got %+v", pack.FuturesEvidence)
	}
	if len(pack.FuturesEvidence.TermStructure.SupportingPoints) == 0 {
		t.Fatalf("expected futures term-structure evidence supporting points, got %+v", pack.FuturesEvidence.TermStructure)
	}
	if len(pack.FuturesEvidence.TapeTechnical.SupportingPoints) == 0 {
		t.Fatalf("expected futures tape-technical evidence supporting points, got %+v", pack.FuturesEvidence.TapeTechnical)
	}
	if len(pack.FuturesEvidence.PositionFlow.SupportingPoints) == 0 {
		t.Fatalf("expected futures position-flow evidence supporting points, got %+v", pack.FuturesEvidence.PositionFlow)
	}
	if len(pack.FuturesEvidence.MacroEvent.SupportingPoints) == 0 {
		t.Fatalf("expected futures macro-event evidence supporting points, got %+v", pack.FuturesEvidence.MacroEvent)
	}
}

func TestBuildStrategyForecastL3ResearchPackSkipsMisleadingZeroEvidenceWhenMetaMissing(t *testing.T) {
	insight := model.StockRecommendationInsight{
		Recommendation: model.StockRecommendation{
			Name:          "测试股票",
			ReasonSummary: "仅保留主线结论。",
		},
		Detail: model.StockRecommendationDetail{
			RiskNote: "关注风险边界。",
		},
		Explanation: model.StrategyClientExplanation{
			ConsensusSummary: "主线结论仍需验证。",
			EvaluationMeta:   map[string]any{},
		},
	}

	evidence := buildStrategyForecastL3StockEvidence(insight)
	joined := strings.Join(append(append([]string{}, evidence.Technical.SupportingPoints...), evidence.Valuation.SupportingPoints...), " | ")
	if strings.Contains(joined, "0.00") {
		t.Fatalf("expected no fabricated zero-valued evidence, got %q", joined)
	}
}

func TestBuildStrategyForecastL3StockEvidenceSkipsQuantOnlyFundamentalClaim(t *testing.T) {
	insight := model.StockRecommendationInsight{
		Recommendation: model.StockRecommendation{
			Name:          "测试股票",
			ReasonSummary: "20日动量23.42%，量比4.56；主力净流入100.00，资金面偏强。",
		},
		Explanation: model.StrategyClientExplanation{
			ConsensusSummary: "资金面偏强，趋势延续。",
		},
	}

	evidence := buildStrategyForecastL3StockEvidence(insight)
	if evidence.Fundamental.Summary != "" || len(evidence.Fundamental.SupportingPoints) != 0 {
		t.Fatalf("expected technical/flow-only statement not to be labelled fundamental, got %+v", evidence.Fundamental)
	}
}

func TestCompactStrategyForecastL3HighlightsRemovesDuplicatesAndBroadMarketNews(t *testing.T) {
	got := compactStrategyForecastL3Highlights([]string{
		"京东方A介绍：面板业务修复",
		"京东方A介绍：面板业务修复与近期进展",
		"45股特大单净流入超2亿元",
		"今日这些个股异动 主力抛售电子、计算机板块",
	})

	if len(got) != 1 || got[0] != "京东方A介绍：面板业务修复" {
		t.Fatalf("expected only concise symbol-specific highlight, got %+v", got)
	}
}

func TestCompactStrategyForecastL3NarrativeClausesRemovesRepeatedRiskPoint(t *testing.T) {
	got := compactStrategyForecastL3NarrativeClauses("风险级别 中风险；观察名单可作为下一轮事件驱动补位池。；按校准后置信度降低节奏执行；观察名单可作为下一轮事件驱动补位池。")
	if strings.Count(got, "观察名单可作为下一轮事件驱动补位池") != 1 {
		t.Fatalf("expected repeated risk clause to be removed, got %q", got)
	}
}

func TestLocalSynthesisForecastL3AdapterDoesNotPresentQuantThesisAsIndustryEvidence(t *testing.T) {
	roles := (localSynthesisForecastL3Adapter{}).RunDeepForecast(strategyForecastL3ResearchPack{
		TargetType: model.StrategyForecastL3TargetTypeStock,
		CoreThesis: "20日动量23.42%，量比4.56，资金面偏强。",
	})

	for _, role := range roles {
		if role.Role != "INDUSTRY" {
			continue
		}
		if role.Stance == "BULLISH" || strings.Contains(role.Summary, "动量") || strings.Contains(role.Summary, "资金面") {
			t.Fatalf("expected industry role not to claim quantitative thesis as industry evidence, got %+v", role)
		}
		return
	}
	t.Fatal("expected industry role in local synthesis")
}

func TestParseStrategyForecastL3ValidationPayloadAcceptsFencedJSON(t *testing.T) {
	got, err := parseStrategyForecastL3ValidationPayload("```json\n{\"verdict\":\"保持观察\",\"llm_summary\":\"证据仍需确认\"}\n```")
	if err != nil {
		t.Fatalf("parseStrategyForecastL3ValidationPayload() error = %v", err)
	}
	if got.Verdict != "保持观察" || got.LLMSummary != "证据仍需确认" {
		t.Fatalf("expected fenced validation payload to parse, got %+v", got)
	}
}

func TestExecuteForecastL3RunBuildsReportAndLogs(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	run, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:     model.StrategyForecastL3TargetTypeStock,
		TargetID:       "sr_001",
		TargetKey:      "600519.SH",
		TargetLabel:    "贵州茅台",
		TriggerType:    model.StrategyForecastL3TriggerTypeAdminManual,
		RequestUserID:  "admin_001",
		OperatorUserID: "admin_001",
		PriorityScore:  0.83,
		Reason:         "manual deep forecast",
	})
	if err != nil {
		t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
	}

	count, err := repo.ExecuteQueuedStrategyForecastL3Runs(1, "system")
	if err != nil {
		t.Fatalf("ExecuteQueuedStrategyForecastL3Runs() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one queued run to be executed, got %d", count)
	}

	detail, err := repo.GetStrategyForecastL3RunDetail(run.ID)
	if err != nil {
		t.Fatalf("GetStrategyForecastL3RunDetail() error = %v", err)
	}
	if detail.Run.Status != model.StrategyForecastL3StatusSucceeded {
		t.Fatalf("expected succeeded run, got %+v", detail.Run)
	}
	if detail.Run.ReportRef == nil || detail.Run.ReportRef.ReportID == "" {
		t.Fatalf("expected report ref on executed run, got %+v", detail.Run)
	}
	if detail.Report == nil || detail.Report.ExecutiveSummary == "" {
		t.Fatalf("expected report snapshot to be persisted, got %+v", detail)
	}
	if len(detail.Logs) < 5 {
		t.Fatalf("expected fixed orchestration logs, got %+v", detail.Logs)
	}
}
