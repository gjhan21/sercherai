package repo

import (
	"strings"
	"testing"
	"time"

	"sercherai/backend/internal/growth/model"
)

func TestBuildForecastL3ReportSynthesizesExecutiveSummaryAndMarkdown(t *testing.T) {
	run := model.StrategyForecastL3Run{
		ID:            "l3run_report_test",
		TargetType:    model.StrategyForecastL3TargetTypeStock,
		TargetKey:     "600519.SH",
		TargetLabel:   "贵州茅台",
		TriggerType:   model.StrategyForecastL3TriggerTypeAdminManual,
		EngineKey:     model.StrategyForecastL3EngineLocalSynthesis,
		Status:        model.StrategyForecastL3StatusRunning,
		PriorityScore: 0.84,
	}
	pack := strategyForecastL3ResearchPack{
		TargetType:        model.StrategyForecastL3TargetTypeStock,
		TargetKey:         "600519.SH",
		TargetLabel:       "贵州茅台",
		CoreThesis:        "当前主逻辑仍由景气和资金回流支撑。",
		RiskBoundary:      "跌破关键支撑位则主情景失效。",
		Invalidations:     []string{"跌破关键支撑位", "资金回流转负"},
		RelatedHighlights: []string{"机构调研热度回升", "白酒板块成交额回暖"},
		ActionHints:       []string{"先看量能确认", "缩短验证窗口"},
	}
	roles := []strategyForecastL3RoleResult{
		{Role: "INDUSTRY", Stance: "BULLISH", Confidence: 0.73, Summary: "行业景气仍在上行区间。"},
		{Role: "FLOW", Stance: "CONSTRUCTIVE", Confidence: 0.68, Summary: "资金回流继续但尚未形成加速。"},
		{Role: "RISK", Stance: "CAUTION", Confidence: 0.62, Summary: "高位追涨赔率一般，需要确认成交额。"},
	}

	report := buildStrategyForecastL3Report(run, pack, roles, strategyForecastL3ValidationResult{
		Status:              model.StrategyForecastL3ValidationStatusCompleted,
		Verdict:             "模型复核认为主情景与现有证据基本一致。",
		ScenarioConsistency: "主情景整体自洽。",
		SupportingEvidence:  []string{"机构调研热度回升"},
		CounterEvidence:     []string{"资金回流转负"},
		BlindSpots:          []string{"缺少更长窗口验证"},
		RiskReview:          []string{"跌破关键支撑位则主情景失效。"},
		ActionReview:        []string{"先看量能确认"},
		LLMSummary:          "当前复核支持继续围绕主情景跟踪。",
	}, time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC))
	if report.ExecutiveSummary == "" || report.PrimaryScenario == "" {
		t.Fatalf("expected report summary and scenario to be built, got %+v", report)
	}
	if len(report.ActionGuidance) == 0 || len(report.TriggerChecklist) == 0 {
		t.Fatalf("expected action guidance and trigger checklist, got %+v", report)
	}
	if report.StateAssessment == nil || report.HeadlineVerdict == "" {
		t.Fatalf("expected structured report state and headline verdict, got %+v", report)
	}
	if len(report.DimensionEvidence) == 0 {
		t.Fatalf("expected structured dimension evidence, got %+v", report)
	}
	if report.ScenarioAssessment == nil {
		t.Fatalf("expected scenario assessment, got %+v", report)
	}
	if report.ValidationReview == nil {
		t.Fatalf("expected validation review placeholder, got %+v", report)
	}
	if strings.Contains(report.MarkdownBody, "Action Guidance") || strings.Contains(report.MarkdownBody, "Alternative Scenarios") {
		t.Fatalf("expected markdown body to be localized in chinese, got %q", report.MarkdownBody)
	}
	if !strings.Contains(report.MarkdownBody, "## 综合应对与操作指引") {
		t.Fatalf("expected markdown to contain chinese action guidance section, got %q", report.MarkdownBody)
	}
	if !strings.Contains(report.HTMLBody, "<h2>主线推演</h2>") {
		t.Fatalf("expected html to contain chinese scenario section, got %q", report.HTMLBody)
	}
}

func TestBuildForecastL3ReportBuildsStockDimensionEvidenceFromDomainEvidence(t *testing.T) {
	report := buildStrategyForecastL3Report(
		model.StrategyForecastL3Run{
			ID:          "l3run_stock_report",
			TargetType:  model.StrategyForecastL3TargetTypeStock,
			TargetKey:   "600519.SH",
			TargetLabel: "贵州茅台",
			TriggerType: model.StrategyForecastL3TriggerTypeUserRequest,
			Status:      model.StrategyForecastL3StatusRunning,
		},
		strategyForecastL3ResearchPack{
			TargetType:   model.StrategyForecastL3TargetTypeStock,
			TargetKey:    "600519.SH",
			TargetLabel:  "贵州茅台",
			CoreThesis:   "估值承压但资金仍在支撑。",
			RiskBoundary: "跌破关键支撑位则主线失效。",
			StockEvidence: model.StrategyForecastL3StockEvidence{
				Fundamental: model.StrategyForecastL3EvidenceSlice{Summary: "盈利质量稳定。", SupportingPoints: []string{"绩效回撤可控"}},
				Technical:   model.StrategyForecastL3EvidenceSlice{Summary: "20日动量维持正值。", SupportingPoints: []string{"趋势强度抬升"}},
				Flow:        model.StrategyForecastL3EvidenceSlice{Summary: "主力净流入修复。", SupportingPoints: []string{"换手温和放大"}},
				Valuation:   model.StrategyForecastL3EvidenceSlice{Summary: "PE处于可跟踪区间。", SupportingPoints: []string{"PB未失控"}},
				Event:       model.StrategyForecastL3EvidenceSlice{Summary: "新闻偏正向。", SupportingPoints: []string{"资讯热度回升"}},
			},
		},
		nil,
		strategyForecastL3ValidationResult{},
		time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
	)

	if got := len(report.DimensionEvidence); got < 5 {
		t.Fatalf("expected at least five stock dimensions, got %d (%+v)", got, report.DimensionEvidence)
	}
	if report.ScenarioAssessment == nil || len(report.ScenarioAssessment.ActionPlan) == 0 {
		t.Fatalf("expected structured scenario assessment, got %+v", report.ScenarioAssessment)
	}
	if report.StateAssessment == nil || report.StateAssessment.CurrentState == "" {
		t.Fatalf("expected current state assessment, got %+v", report.StateAssessment)
	}
	if !containsDimension(report.DimensionEvidence, "TECHNICAL") || !containsDimension(report.DimensionEvidence, "VALUATION") {
		t.Fatalf("expected stock dimensions to include technical and valuation, got %+v", report.DimensionEvidence)
	}
	if !containsForecastString(report.ScenarioAssessment.TriggerConditions, "趋势强度抬升") {
		t.Fatalf("expected stock scenario triggers to include evidence-derived trigger, got %+v", report.ScenarioAssessment.TriggerConditions)
	}
	if !containsForecastString(report.ScenarioAssessment.ActionPlan, "优先等待技术面与资金面继续共振确认。") {
		t.Fatalf("expected stock scenario actions to include evidence-derived action, got %+v", report.ScenarioAssessment.ActionPlan)
	}
}

func TestBuildForecastL3ReportKeepsDegradedValidationOutOfHeadline(t *testing.T) {
	report := buildStrategyForecastL3Report(
		model.StrategyForecastL3Run{
			ID:          "l3run_degraded_validation",
			TargetType:  model.StrategyForecastL3TargetTypeStock,
			TargetKey:   "000725.SZ",
			TargetLabel: "京东方A",
			EngineKey:   model.StrategyForecastL3EngineLocalSynthesis,
		},
		strategyForecastL3ResearchPack{
			TargetType:  model.StrategyForecastL3TargetTypeStock,
			TargetKey:   "000725.SZ",
			TargetLabel: "京东方A",
			CoreThesis:  "面板价格修复与资金回流仍需继续确认。",
		},
		nil,
		strategyForecastL3ValidationResult{
			Status:     model.StrategyForecastL3ValidationStatusDegraded,
			Verdict:    "模型复核输出格式异常，已回退为结构化本地复核。",
			LLMSummary: "模型复核返回不可解析内容，主报告仍可正常使用。",
		},
		time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC),
	)

	if strings.Contains(report.HeadlineVerdict, "格式异常") || strings.Contains(report.HeadlineVerdict, "模型复核") {
		t.Fatalf("expected degraded validation to stay out of headline, got %q", report.HeadlineVerdict)
	}
	if !strings.Contains(report.HeadlineVerdict, "面板价格修复") {
		t.Fatalf("expected headline to use core thesis, got %q", report.HeadlineVerdict)
	}
}

func TestBuildForecastL3ReportDoesNotPublishSyntheticPrecision(t *testing.T) {
	report := buildStrategyForecastL3Report(
		model.StrategyForecastL3Run{
			ID:          "l3run_local_synthesis",
			TargetType:  model.StrategyForecastL3TargetTypeStock,
			TargetKey:   "000725.SZ",
			TargetLabel: "京东方A",
			EngineKey:   model.StrategyForecastL3EngineLocalSynthesis,
		},
		strategyForecastL3ResearchPack{
			TargetType:  model.StrategyForecastL3TargetTypeStock,
			TargetKey:   "000725.SZ",
			TargetLabel: "京东方A",
			CoreThesis:  "等待更多结构化证据确认。",
		},
		[]strategyForecastL3RoleResult{
			{Role: "FLOW", Stance: "WATCH", Confidence: 0.72, Summary: "资金方向仍待确认。"},
		},
		strategyForecastL3ValidationResult{Status: model.StrategyForecastL3ValidationStatusDegraded},
		time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC),
	)

	if report.Summary.ConfidenceLabel != "" {
		t.Fatalf("expected no computed confidence label for local synthesis, got %q", report.Summary.ConfidenceLabel)
	}
	for _, item := range report.AlternativeScenarios {
		if item.Probability != 0 {
			t.Fatalf("expected scenario probability to be omitted without model calculation, got %+v", report.AlternativeScenarios)
		}
	}
	if strings.Contains(report.MarkdownBody, "发生概率") || strings.Contains(report.HTMLBody, "发生概率") {
		t.Fatalf("expected generated bodies not to present synthetic probability, got %q", report.MarkdownBody)
	}
}

func TestBuildForecastL3ValidationPromptIncludesDomainEvidence(t *testing.T) {
	prompt := buildStrategyForecastL3ValidationPrompt(
		model.StrategyForecastL3Run{
			TargetType:  model.StrategyForecastL3TargetTypeFutures,
			TargetKey:   "AU2408",
			TargetLabel: "沪金主力",
		},
		strategyForecastL3ResearchPack{
			TargetType:  model.StrategyForecastL3TargetTypeFutures,
			TargetKey:   "AU2408",
			TargetLabel: "沪金主力",
			CoreThesis:  "基差与库存共同支撑近月。",
			FuturesEvidence: model.StrategyForecastL3FuturesEvidence{
				SupplyDemand:  model.StrategyForecastL3EvidenceSlice{Summary: "库存压力缓和。", SupportingPoints: []string{"库存延续去化"}},
				TermStructure: model.StrategyForecastL3EvidenceSlice{Summary: "近月结构占优。", SupportingPoints: []string{"基差与期限结构同向"}},
			},
		},
		nil,
	)

	if !strings.Contains(prompt, "库存压力缓和") || !strings.Contains(prompt, "近月结构占优") {
		t.Fatalf("expected domain evidence in validation prompt, got %q", prompt)
	}
}

func TestBuildForecastL3ReportBuildsFuturesDimensionEvidenceFromDomainEvidence(t *testing.T) {
	report := buildStrategyForecastL3Report(
		model.StrategyForecastL3Run{
			ID:          "l3run_futures_report",
			TargetType:  model.StrategyForecastL3TargetTypeFutures,
			TargetKey:   "AU2408",
			TargetLabel: "沪金主力",
			TriggerType: model.StrategyForecastL3TriggerTypeUserRequest,
			Status:      model.StrategyForecastL3StatusRunning,
		},
		strategyForecastL3ResearchPack{
			TargetType:   model.StrategyForecastL3TargetTypeFutures,
			TargetKey:    "AU2408",
			TargetLabel:  "沪金主力",
			CoreThesis:   "近月结构和库存去化仍支撑主线。",
			RiskBoundary: "期限结构转弱且流向翻空，则主线失效。",
			FuturesEvidence: model.StrategyForecastL3FuturesEvidence{
				SupplyDemand:  model.StrategyForecastL3EvidenceSlice{Summary: "库存压力缓和。", SupportingPoints: []string{"库存延续去化"}, RiskPoints: []string{"若库存反弹则供需修复放缓"}},
				TermStructure: model.StrategyForecastL3EvidenceSlice{Summary: "近月结构占优。", SupportingPoints: []string{"基差与期限结构同向"}, RiskPoints: []string{"若结构背离则主线减弱"}},
				TapeTechnical: model.StrategyForecastL3EvidenceSlice{Summary: "趋势维持。", SupportingPoints: []string{"盘面波动受控"}},
				PositionFlow:  model.StrategyForecastL3EvidenceSlice{Summary: "持仓资金偏多。", SupportingPoints: []string{"流向偏置仍偏多"}},
				MacroEvent:    model.StrategyForecastL3EvidenceSlice{Summary: "宏观事件中性偏稳。", SupportingPoints: []string{"风险偏好未明显恶化"}},
			},
		},
		nil,
		strategyForecastL3ValidationResult{},
		time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
	)

	if !containsDimension(report.DimensionEvidence, "SUPPLY_DEMAND") || !containsDimension(report.DimensionEvidence, "TERM_STRUCTURE") {
		t.Fatalf("expected futures dimensions to include supply-demand and term-structure, got %+v", report.DimensionEvidence)
	}
	if report.StateAssessment == nil || report.StateAssessment.CurrentState == "" {
		t.Fatalf("expected futures current state assessment, got %+v", report.StateAssessment)
	}
	if !containsForecastString(report.ScenarioAssessment.InvalidationConditions, "若结构背离则主线减弱") {
		t.Fatalf("expected futures invalidation conditions to include evidence-derived risk, got %+v", report.ScenarioAssessment.InvalidationConditions)
	}
}

func TestBuildForecastL3ReportFallsBackToRoleEvidenceWhenDomainEvidenceMissing(t *testing.T) {
	report := buildStrategyForecastL3Report(
		model.StrategyForecastL3Run{
			ID:          "l3run_fallback_report",
			TargetType:  model.StrategyForecastL3TargetTypeStock,
			TargetKey:   "000001.SZ",
			TargetLabel: "平安银行",
			TriggerType: model.StrategyForecastL3TriggerTypeAdminManual,
			Status:      model.StrategyForecastL3StatusRunning,
		},
		strategyForecastL3ResearchPack{
			TargetType:        model.StrategyForecastL3TargetTypeStock,
			TargetKey:         "000001.SZ",
			TargetLabel:       "平安银行",
			CoreThesis:        "等待更多证据确认。",
			RelatedHighlights: []string{"资金面仍待确认"},
			Invalidations:     []string{"跌破关键支撑位"},
		},
		[]strategyForecastL3RoleResult{
			{Role: "FLOW", Stance: "WATCH", Confidence: 0.62, Summary: "资金面仍待确认。"},
		},
		strategyForecastL3ValidationResult{},
		time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
	)

	if len(report.DimensionEvidence) == 0 {
		t.Fatalf("expected fallback role-driven dimension evidence, got %+v", report.DimensionEvidence)
	}
	if report.DimensionEvidence[0].Summary == "" {
		t.Fatalf("expected fallback dimension evidence summary, got %+v", report.DimensionEvidence[0])
	}
}

func containsDimension(items []model.StrategyForecastL3DimensionEvidence, target string) bool {
	for _, item := range items {
		if item.Dimension == target {
			return true
		}
	}
	return false
}

func containsForecastString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
