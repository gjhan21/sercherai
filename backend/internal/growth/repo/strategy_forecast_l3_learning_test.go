package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestBuildStrategyForecastL3RunReviewGradesScenarioAndTriggerHit(t *testing.T) {
	record := model.StrategyForecastL3LearningRecord{
		RunID:             "l3run_review_a",
		TargetType:        model.StrategyForecastL3TargetTypeStock,
		TargetKey:         "600519.SH",
		ScenarioHit:       true,
		TriggerHit:        true,
		InvalidationEarly: false,
		BiasLabel:         "UNDERCONFIRMED",
		RoleEffectiveness: map[string]float64{"TECHNICAL": 0.78},
		Summary:           "主情景得到验证",
		CreatedAt:         "2026-05-24T12:00:00Z",
	}

	review := buildStrategyForecastL3RunReview(record)
	if review.ReviewGrade != "A" || review.ReviewScore != 85 {
		t.Fatalf("expected A/85, got %+v", review)
	}
	if review.ReviewVerdict == "" {
		t.Fatalf("expected human-readable review verdict, got %+v", review)
	}
}

func TestBuildStrategyForecastL3RunReviewPenalizesEarlyInvalidation(t *testing.T) {
	record := model.StrategyForecastL3LearningRecord{
		RunID:             "l3run_review_d",
		TargetType:        model.StrategyForecastL3TargetTypeFutures,
		TargetKey:         "AU2408",
		ScenarioHit:       false,
		TriggerHit:        false,
		InvalidationEarly: true,
		BiasLabel:         "RISK_FIRST",
		Summary:           "风险边界过早触发",
		CreatedAt:         "2026-05-24T13:00:00Z",
	}

	review := buildStrategyForecastL3RunReview(record)
	if review.ReviewGrade != "D" || review.ReviewScore != 35 {
		t.Fatalf("expected D/35, got %+v", review)
	}
	if len(review.ReviewNotes) == 0 {
		t.Fatalf("expected review notes, got %+v", review)
	}
}

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
