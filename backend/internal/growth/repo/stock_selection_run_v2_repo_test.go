package repo

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const stockSelectionOverviewCoverageQueryPattern = `(?s)SELECT\s+run_id\s+FROM stock_selection_runs.*LIMIT 24`
const stockSelectionOverviewEvaluationSummaryQueryPattern = `(?s)SELECT\s+horizon_day,.*FROM stock_selection_run_evaluations.*WHERE evaluation_scope = 'PORTFOLIO'.*GROUP BY horizon_day.*ORDER BY horizon_day ASC`

func TestLoadStockSelectionOverviewEvaluationSummaryBuildsStableHorizonMap(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(stockSelectionOverviewCoverageQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"run_id"}))

	mock.ExpectQuery(stockSelectionOverviewEvaluationSummaryQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{
			"horizon_day",
			"count",
			"avg_return_pct",
			"avg_excess_return_pct",
			"hit_rate",
			"avg_max_drawdown_pct",
			"worst_max_drawdown_pct",
			"generated_at",
		}).AddRow(
			1,
			6,
			0.0123,
			0.0045,
			0.66,
			-0.018,
			-0.031,
			"2026-03-22T00:30:00Z",
		).AddRow(
			5,
			4,
			0.0345,
			0.0112,
			0.75,
			-0.027,
			-0.046,
			"2026-03-22T00:30:00Z",
		))

	summary, err := repo.loadStockSelectionOverviewEvaluationSummary()
	if err != nil {
		t.Fatalf("load overview evaluation summary: %v", err)
	}

	row1, ok := summary["1"].(map[string]any)
	if !ok {
		t.Fatalf("expected 1-day row map, got %#v", summary["1"])
	}
	if got := asInt(row1["sample_count"]); got != 6 {
		t.Fatalf("expected 1-day sample_count=6, got %d", got)
	}
	if got := asFloat(row1["avg_return_pct"]); got != 0.0123 {
		t.Fatalf("expected 1-day avg_return_pct=0.0123, got %v", got)
	}
	if got := asFloat(row1["hit_rate"]); got != 0.66 {
		t.Fatalf("expected 1-day hit_rate=0.66, got %v", got)
	}

	row3, ok := summary["3"].(map[string]any)
	if !ok {
		t.Fatalf("expected 3-day row map, got %#v", summary["3"])
	}
	if got := asInt(row3["sample_count"]); got != 0 {
		t.Fatalf("expected 3-day sample_count=0 default, got %d", got)
	}

	row5, ok := summary["5"].(map[string]any)
	if !ok {
		t.Fatalf("expected 5-day row map, got %#v", summary["5"])
	}
	if got := asFloat(row5["avg_excess_return_pct"]); got != 0.0112 {
		t.Fatalf("expected 5-day avg_excess_return_pct=0.0112, got %v", got)
	}
	if got := asFloat(row5["worst_max_drawdown_pct"]); got != -0.046 {
		t.Fatalf("expected 5-day worst_max_drawdown_pct=-0.046, got %v", got)
	}

	for _, horizon := range []string{"1", "3", "5", "10", "20"} {
		if _, ok := summary[horizon]; !ok {
			t.Fatalf("expected stable summary row for horizon %s", horizon)
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
