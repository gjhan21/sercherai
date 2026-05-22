package repo

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"sercherai/backend/internal/growth/model"
)

const stockSelectionOverviewCoverageQueryPattern = `(?s)SELECT\s+run_id\s+FROM stock_selection_runs.*LIMIT 24`
const stockSelectionOverviewEvaluationSummaryQueryPattern = `(?s)SELECT\s+horizon_day,.*FROM stock_selection_run_evaluations.*WHERE evaluation_scope IN \('PORTFOLIO', 'SHORT_TERM_PRIMARY'\).*COALESCE\(entry_price, 0\) > 0.*COALESCE\(exit_price, 0\) > 0.*GROUP BY horizon_day.*ORDER BY horizon_day ASC`

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

func TestInMemoryStockSelectionArtifactsExposeDualLayerFields(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	candidates, err := repo.AdminListStockSelectionRunCandidates("ssr_demo_001")
	if err != nil {
		t.Fatalf("list in-memory candidates: %v", err)
	}
	if len(candidates) == 0 {
		t.Fatalf("expected demo candidates")
	}
	if candidates[0].RecommendationHead != "SHORT_TERM_PRIMARY" {
		t.Fatalf("expected recommendation head SHORT_TERM_PRIMARY, got %q", candidates[0].RecommendationHead)
	}
	if candidates[0].SelectionLayer == "" || candidates[0].TechnicalPattern == "" {
		t.Fatalf("expected dual-layer candidate metadata, got %+v", candidates[0])
	}

	evaluations, err := repo.AdminListStockSelectionRunEvaluations("ssr_demo_001", "")
	if err != nil {
		t.Fatalf("list in-memory evaluations: %v", err)
	}
	if len(evaluations) == 0 {
		t.Fatalf("expected demo evaluations")
	}
	if evaluations[0].HeadLabel == "" || evaluations[0].HoldingContract == "" {
		t.Fatalf("expected dual-layer evaluation metadata, got %+v", evaluations[0])
	}
}

func TestStockSelectionEvaluationScopeHelpersMapLegacyScopes(t *testing.T) {
	if got := stockSelectionDisplayEvaluationScope("PORTFOLIO"); got != "SHORT_TERM_PRIMARY" {
		t.Fatalf("expected PORTFOLIO -> SHORT_TERM_PRIMARY, got %q", got)
	}
	if got := stockSelectionDisplayEvaluationScope("CANDIDATE"); got != "CANDIDATE_POOL" {
		t.Fatalf("expected CANDIDATE -> CANDIDATE_POOL, got %q", got)
	}
	if got := stockSelectionHeadLabel("SHORT_TERM_PRIMARY"); got != "超短线主推荐" {
		t.Fatalf("unexpected head label: %q", got)
	}
	if got := stockSelectionHoldingContract("SWING_AUXILIARY"); got != "3-10D" {
		t.Fatalf("unexpected holding contract: %q", got)
	}
}

func TestStockSelectionEvaluationRowIsMaterialized(t *testing.T) {
	if !stockSelectionEvaluationRowIsMaterialized(model.StockSelectionRunEvaluation{
		EntryDate:  "2026-05-20",
		ExitDate:   "2026-05-21",
		EntryPrice: 12.3,
		ExitPrice:  12.9,
	}) {
		t.Fatalf("expected populated evaluation row to be materialized")
	}
	if stockSelectionEvaluationRowIsMaterialized(model.StockSelectionRunEvaluation{
		EntryDate:  "2026-05-20",
		ExitDate:   "2026-05-21",
		EntryPrice: 0,
		ExitPrice:  12.9,
	}) {
		t.Fatalf("expected zero-price placeholder row to stay non-materialized")
	}
}

func TestStockSelectionEvaluationMaterializedPredicate(t *testing.T) {
	got := stockSelectionEvaluationMaterializedPredicate("e")
	want := "COALESCE(e.entry_price, 0) > 0 AND COALESCE(e.exit_price, 0) > 0 AND e.entry_date IS NOT NULL AND e.exit_date IS NOT NULL"
	if got != want {
		t.Fatalf("unexpected materialized predicate: %q", got)
	}
}

func TestAdminListStockSelectionRunEvaluationsSkipsPlaceholderRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(`SELECT id, run_id, symbol, horizon_day, evaluation_scope, COALESCE\(name, ''\)`).
		WithArgs("ssr_eval_rows").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "run_id", "symbol", "horizon_day", "evaluation_scope", "name",
			"entry_date", "exit_date", "entry_price", "exit_price",
			"return_pct", "excess_return_pct", "max_drawdown_pct", "hit_flag", "benchmark_symbol",
			"head_label", "holding_contract", "created_at", "updated_at",
		}).AddRow(
			"sseval_placeholder",
			"ssr_eval_rows",
			"600519.SH",
			1,
			"SHORT_TERM_PRIMARY",
			"贵州茅台",
			"",
			"",
			0,
			0,
			0,
			0,
			0,
			false,
			"000300.SH",
			"超短线主推荐",
			"1-2D",
			"2026-05-20T10:00:00Z",
			"2026-05-20T10:00:00Z",
		).AddRow(
			"sseval_real",
			"ssr_eval_rows",
			"000858.SZ",
			1,
			"SWING_AUXILIARY",
			"五粮液",
			"2026-05-20",
			"2026-05-21",
			132.1,
			134.8,
			0.0204,
			0.0111,
			-0.008,
			true,
			"000300.SH",
			"短波段辅助",
			"3-10D",
			"2026-05-21T10:00:00Z",
			"2026-05-21T10:00:00Z",
		))

	items, err := repo.AdminListStockSelectionRunEvaluations("ssr_eval_rows", "")
	if err != nil {
		t.Fatalf("list evaluations: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected only materialized rows, got %d", len(items))
	}
	if items[0].Symbol != "000858.SZ" {
		t.Fatalf("expected real row to survive filtering, got %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestBuildStockSelectionEvaluationStatusMapTracksPartialAndReady(t *testing.T) {
	rows := []model.StockSelectionRunEvaluation{
		{Symbol: "600519.SH", EvaluationScope: "SHORT_TERM_PRIMARY", HorizonDay: 1},
		{Symbol: "600519.SH", EvaluationScope: "SHORT_TERM_PRIMARY", HorizonDay: 3},
		{Symbol: "000858.SZ", EvaluationScope: "SWING_AUXILIARY", HorizonDay: 1},
		{Symbol: "000858.SZ", EvaluationScope: "SWING_AUXILIARY", HorizonDay: 3},
		{Symbol: "000858.SZ", EvaluationScope: "SWING_AUXILIARY", HorizonDay: 5},
		{Symbol: "000858.SZ", EvaluationScope: "SWING_AUXILIARY", HorizonDay: 10},
		{Symbol: "000858.SZ", EvaluationScope: "SWING_AUXILIARY", HorizonDay: 20},
		{Symbol: "300750.SZ", EvaluationScope: "CANDIDATE", HorizonDay: 1},
	}

	statusMap := buildStockSelectionEvaluationStatusMap(rows)

	if got := statusMap[stockSelectionEvaluationStatusMapKey("600519.SH", "SHORT_TERM_PRIMARY")]; got != "PARTIAL" {
		t.Fatalf("expected partial short-term status, got %q", got)
	}
	if got := statusMap[stockSelectionEvaluationStatusMapKey("000858.SZ", "SWING_AUXILIARY")]; got != "READY" {
		t.Fatalf("expected ready swing status, got %q", got)
	}
	if got := statusMap[stockSelectionEvaluationStatusMapKey("300750.SZ", "CANDIDATE_POOL")]; got != "PARTIAL" {
		t.Fatalf("expected legacy candidate scope to map to PARTIAL, got %q", got)
	}
}
