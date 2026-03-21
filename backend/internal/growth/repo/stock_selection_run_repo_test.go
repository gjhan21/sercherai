package repo

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const stockSelectionRunListCountQueryPattern = `SELECT COUNT\(\*\) FROM stock_selection_runs r LEFT JOIN stock_selection_publish_reviews rv ON rv\.run_id = r\.run_id`
const stockSelectionRunListQueryPattern = `(?s)SELECT\s+r\.run_id,.*COALESCE\(rv\.publish_version,\s*0\).*FROM stock_selection_runs r`
const stockSelectionReviewListCountQueryPattern = `SELECT COUNT\(\*\) FROM stock_selection_publish_reviews`
const stockSelectionReviewListQueryPattern = `(?s)SELECT id, run_id, review_status,.*COALESCE\(publish_version,\s*0\).*FROM stock_selection_publish_reviews`

func TestAdminListStockSelectionRunsHandlesPendingReviewWithoutPublishVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(stockSelectionRunListCountQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(stockSelectionRunListQueryPattern).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"run_id",
			"trade_date",
			"job_id",
			"profile_id",
			"profile_version",
			"template_id",
			"template_name",
			"market_regime",
			"selection_mode",
			"universe_scope",
			"status",
			"result_summary",
			"warning_messages",
			"warning_count",
			"universe_count",
			"seed_count",
			"candidate_count",
			"selected_count",
			"publish_count",
			"context_meta",
			"template_snapshot",
			"compare_summary",
			"started_at",
			"completed_at",
			"created_by",
			"created_at",
			"updated_at",
			"review_id",
			"review_status",
			"reviewer",
			"review_note",
			"override_reason",
			"publish_id",
			"publish_version",
			"published_portfolio_snapshot",
			"approved_at",
			"rejected_at",
			"review_created_at",
			"review_updated_at",
		}).AddRow(
			"ssr_demo_pending",
			"2026-03-21",
			"",
			"profile_default_stock_auto",
			1,
			"sstpl_balanced_steady",
			"均衡稳健",
			"ROTATION",
			"AUTO",
			"CN_A_ALL",
			"SUCCEEDED",
			"",
			"[]",
			0,
			180,
			180,
			30,
			5,
			0,
			"{}",
			"{}",
			"{}",
			"2026-03-21T09:00:00Z",
			"2026-03-21T09:00:05Z",
			"admin_001",
			"2026-03-21T09:00:00Z",
			"2026-03-21T09:00:05Z",
			"review_demo_pending",
			"PENDING",
			"",
			"",
			"",
			"",
			0,
			"[]",
			"",
			"",
			"2026-03-21T09:00:05Z",
			"2026-03-21T09:00:05Z",
		))

	items, total, err := repo.AdminListStockSelectionRuns("", "", "", 1, 20)
	if err != nil {
		t.Fatalf("list runs: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Review == nil {
		t.Fatalf("expected pending review to be attached")
	}
	if items[0].Review.PublishVersion != 0 {
		t.Fatalf("expected pending review publish version to default to 0, got %d", items[0].Review.PublishVersion)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminListStockSelectionReviewsHandlesPendingRowsWithoutPublishVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(stockSelectionReviewListCountQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(stockSelectionReviewListQueryPattern).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"run_id",
			"review_status",
			"reviewer",
			"review_note",
			"override_reason",
			"publish_id",
			"publish_version",
			"approved_at",
			"rejected_at",
			"created_at",
			"updated_at",
		}).AddRow(
			"review_demo_pending",
			"ssr_demo_pending",
			"PENDING",
			"",
			"",
			"",
			"",
			0,
			"",
			"",
			"2026-03-21T09:00:05Z",
			"2026-03-21T09:00:05Z",
		))

	items, total, err := repo.AdminListStockSelectionReviews("", 1, 20)
	if err != nil {
		t.Fatalf("list reviews: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].PublishVersion != 0 {
		t.Fatalf("expected pending review publish version to default to 0, got %d", items[0].PublishVersion)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
