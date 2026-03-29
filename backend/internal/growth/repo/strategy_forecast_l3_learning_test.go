package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestRunForecastL3QualityBackfillWritesLearningRecord(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	_, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:     model.StrategyForecastL3TargetTypeFutures,
		TargetID:       "fs_001",
		TargetKey:      "IF2603",
		TargetLabel:    "股指趋势跟踪",
		TriggerType:    model.StrategyForecastL3TriggerTypeAdminManual,
		RequestUserID:  "admin_001",
		OperatorUserID: "admin_001",
		PriorityScore:  0.79,
		Reason:         "quality backfill test",
	})
	if err != nil {
		t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
	}
	if _, err := repo.ExecuteQueuedStrategyForecastL3Runs(5, "system"); err != nil {
		t.Fatalf("ExecuteQueuedStrategyForecastL3Runs() error = %v", err)
	}

	count, err := repo.RunStrategyForecastL3QualityBackfill(20, "system")
	if err != nil {
		t.Fatalf("RunStrategyForecastL3QualityBackfill() error = %v", err)
	}
	if count == 0 {
		t.Fatalf("expected learning records to be written")
	}

	items, err := repo.ListStrategyForecastL3QualitySummaries(model.StrategyForecastL3TargetTypeFutures, 30)
	if err != nil {
		t.Fatalf("ListStrategyForecastL3QualitySummaries() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected quality summary items")
	}
	if items[0].TotalRuns == 0 {
		t.Fatalf("expected non-zero quality summary totals, got %+v", items[0])
	}
}
