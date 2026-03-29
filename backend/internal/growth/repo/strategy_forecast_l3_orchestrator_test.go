package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

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
