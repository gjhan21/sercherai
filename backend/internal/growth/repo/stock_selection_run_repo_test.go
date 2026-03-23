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

func TestBuildStockSelectionRunContextMetaIncludesResearchFields(t *testing.T) {
	report := strategyEngineStockSelectionReport{
		GraphSummary:    "图谱显示机器人主线和算力题材继续共振。",
		GraphSnapshotID: "gss_demo_001",
		ContextMeta: map[string]any{
			"selected_trade_date": "2026-03-21",
			"graph_write_status":  "WRITTEN",
		},
		RelatedEntities: []map[string]any{
			{"label": "机器人", "entity_type": "ConceptTheme"},
			{"label": "算力", "entity_type": "ConceptTheme"},
		},
		MemoryFeedback: map[string]any{
			"summary":     "题材共振较强，但高位分歧需要继续跟踪。",
			"suggestions": []any{"下次降低高波动票权重", "优先观察成交额持续性"},
		},
	}

	contextMeta := buildStockSelectionRunContextMeta(report)
	if got := stringValue(contextMeta["selected_trade_date"]); got != "2026-03-21" {
		t.Fatalf("expected selected_trade_date to be preserved, got %q", got)
	}
	if got := stringValue(contextMeta["graph_summary"]); got != report.GraphSummary {
		t.Fatalf("expected graph_summary=%q, got %q", report.GraphSummary, got)
	}
	if got := stringValue(contextMeta["graph_snapshot_id"]); got != report.GraphSnapshotID {
		t.Fatalf("expected graph_snapshot_id=%q, got %q", report.GraphSnapshotID, got)
	}
	relatedEntities, ok := contextMeta["related_entities"].([]map[string]any)
	if !ok || len(relatedEntities) != 2 {
		t.Fatalf("expected 2 related_entities, got %#v", contextMeta["related_entities"])
	}
	memoryFeedback, ok := contextMeta["memory_feedback"].(map[string]any)
	if !ok {
		t.Fatalf("expected memory_feedback map, got %#v", contextMeta["memory_feedback"])
	}
	if got := stringValue(memoryFeedback["summary"]); got == "" {
		t.Fatalf("expected memory_feedback summary to be stored")
	}
}
