package repo

import (
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
