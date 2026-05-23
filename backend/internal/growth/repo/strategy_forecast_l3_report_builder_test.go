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
