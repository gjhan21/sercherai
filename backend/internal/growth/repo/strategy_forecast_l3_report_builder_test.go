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

	report := buildStrategyForecastL3Report(run, pack, roles, time.Date(2026, 3, 29, 12, 0, 0, 0, time.UTC))
	if report.ExecutiveSummary == "" || report.PrimaryScenario == "" {
		t.Fatalf("expected report summary and scenario to be built, got %+v", report)
	}
	if len(report.ActionGuidance) == 0 || len(report.TriggerChecklist) == 0 {
		t.Fatalf("expected action guidance and trigger checklist, got %+v", report)
	}
	if !strings.Contains(report.MarkdownBody, "## Action Guidance") {
		t.Fatalf("expected markdown to contain action guidance section, got %q", report.MarkdownBody)
	}
	if !strings.Contains(report.HTMLBody, "<h2>Primary Scenario</h2>") {
		t.Fatalf("expected html to contain scenario section, got %q", report.HTMLBody)
	}
}
