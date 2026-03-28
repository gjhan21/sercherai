package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const stockSelectionEvaluationRunMetaQueryPattern = `(?s)SELECT r\.trade_date, COALESCE\(CAST\(r\.context_meta AS CHAR\), ''\), COALESCE\(CAST\(a\.report_snapshot AS CHAR\), ''\).*FROM stock_selection_runs r`

func TestLoadStockSelectionEvaluationRunMetaPrefersSelectedTradeDate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(stockSelectionEvaluationRunMetaQueryPattern).
		WithArgs("ssr_eval_demo").
		WillReturnRows(sqlmock.NewRows([]string{"trade_date", "context_meta", "report_snapshot"}).AddRow(
			time.Date(2026, 3, 21, 0, 0, 0, 0, time.Local),
			`{"selected_trade_date":"2026-03-20"}`,
			`{"context_meta":{"selected_trade_date":"2026-03-19"},"evaluation_summary":{"benchmark_symbol":"sh000001"}}`,
		))

	tradeDate, benchmarkSymbol, err := repo.loadStockSelectionEvaluationRunMeta("ssr_eval_demo")
	if err != nil {
		t.Fatalf("load run meta: %v", err)
	}
	if got := tradeDate.Format("2006-01-02"); got != "2026-03-19" {
		t.Fatalf("expected selected trade date from report context, got %s", got)
	}
	if benchmarkSymbol != "000001.SH" {
		t.Fatalf("expected canonicalized benchmark symbol, got %s", benchmarkSymbol)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestLoadStockSelectionEvaluationRunMetaFallsBackToRunContextAndTradeDate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(stockSelectionEvaluationRunMetaQueryPattern).
		WithArgs("ssr_eval_fallback").
		WillReturnRows(sqlmock.NewRows([]string{"trade_date", "context_meta", "report_snapshot"}).AddRow(
			time.Date(2026, 3, 21, 0, 0, 0, 0, time.Local),
			`{"selected_trade_date":"2026-03-20"}`,
			`{"evaluation_summary":{"benchmark_symbol":"000300.SH"}}`,
		))

	tradeDate, benchmarkSymbol, err := repo.loadStockSelectionEvaluationRunMeta("ssr_eval_fallback")
	if err != nil {
		t.Fatalf("load run meta: %v", err)
	}
	if got := tradeDate.Format("2006-01-02"); got != "2026-03-20" {
		t.Fatalf("expected selected trade date from persisted run context, got %s", got)
	}
	if benchmarkSymbol != "000300.SH" {
		t.Fatalf("expected benchmark symbol passthrough, got %s", benchmarkSymbol)
	}

	mock.ExpectQuery(stockSelectionEvaluationRunMetaQueryPattern).
		WithArgs("ssr_eval_plain").
		WillReturnRows(sqlmock.NewRows([]string{"trade_date", "context_meta", "report_snapshot"}).AddRow(
			time.Date(2026, 3, 21, 0, 0, 0, 0, time.Local),
			`{}`,
			`{}`,
		))

	tradeDate, benchmarkSymbol, err = repo.loadStockSelectionEvaluationRunMeta("ssr_eval_plain")
	if err != nil {
		t.Fatalf("load plain run meta: %v", err)
	}
	if got := tradeDate.Format("2006-01-02"); got != "2026-03-21" {
		t.Fatalf("expected fallback to run trade date, got %s", got)
	}
	if benchmarkSymbol != "" {
		t.Fatalf("expected empty benchmark symbol, got %s", benchmarkSymbol)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestNormalizeStockSelectionEvaluationSummary(t *testing.T) {
	empty := normalizeStockSelectionEvaluationSummary(nil)
	if empty["status"] != "PENDING" {
		t.Fatalf("expected pending summary for empty input, got %+v", empty)
	}

	summary := normalizeStockSelectionEvaluationSummary(map[string]any{
		"5": map[string]any{"return_pct": 0.04},
	})
	if summary["status"] != "COMPLETED" {
		t.Fatalf("expected completed summary when payload exists, got %+v", summary)
	}
}
