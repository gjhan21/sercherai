package repo

import (
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"sercherai/backend/internal/growth/model"
)

const stockEventClusterUpsertPattern = `INSERT INTO stock_event_clusters`
const stockEventItemDeletePattern = `DELETE FROM stock_event_items WHERE cluster_id = \?`
const stockEventItemInsertPattern = `INSERT INTO stock_event_items`
const stockEventEntityDeletePattern = `DELETE FROM stock_event_entities WHERE cluster_id = \?`
const stockEventEntityInsertPattern = `INSERT INTO stock_event_entities`
const stockEventEdgeDeletePattern = `DELETE FROM stock_event_edges WHERE cluster_id = \?`
const stockEventEdgeInsertPattern = `INSERT INTO stock_event_edges`
const stockEventClusterSelectPattern = `(?s)SELECT\s+c\.id,.*FROM stock_event_clusters c`
const stockEventItemSelectPattern = `SELECT id, cluster_id, source_key, source_item_id, title, summary, primary_symbol, symbols_json, metadata_json, published_at, created_at, updated_at FROM stock_event_items WHERE cluster_id = \? ORDER BY published_at DESC, id ASC`
const stockEventEntitySelectPattern = `SELECT id, cluster_id, entity_type, entity_key, label, symbol, sector_label, topic_label, confidence_score, metadata_json, created_at, updated_at FROM stock_event_entities WHERE cluster_id = \? ORDER BY confidence_score DESC, id ASC`
const stockEventEdgeSelectPattern = `SELECT id, cluster_id, subject_entity_id, object_entity_id, relation_type, relation_weight, metadata_json, created_at, updated_at FROM stock_event_edges WHERE cluster_id = \? ORDER BY relation_weight DESC, id ASC`
const stockEventReviewInsertPattern = `INSERT INTO stock_event_reviews`
const stockEventClusterReviewUpdatePattern = `UPDATE stock_event_clusters SET cluster_status = \?, review_status = \?, updated_at = NOW\(\) WHERE id = \?`
const stockEventClusterCountPattern = `SELECT COUNT\(DISTINCT c\.id\) FROM stock_event_clusters c`
const stockEventClusterListPattern = `(?s)SELECT\s+c\.id,.*FROM stock_event_clusters c`

func TestAdminUpsertStockEventClusterPersistsMembersEntitiesAndEdges(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	cluster := model.StockEventCluster{
		ID:            "sec_demo_001",
		ClusterKey:    "earnings:600519.SH:2026Q1",
		EventType:     "EARNINGS",
		Title:         "贵州茅台一季报超预期",
		Summary:       "营收与利润均超市场预期",
		Source:        "MARKET_NEWS",
		PrimarySymbol: "600519.SH",
		SectorLabel:   "消费",
		TopicLabel:    "白酒",
		Status:        "CLUSTERED",
		ReviewStatus:  "PENDING",
		NewsCount:     2,
		Confidence:    0.91,
		Metadata:      map[string]any{"source_count": float64(2)},
		Items: []model.StockEventItem{
			{ID: "sei_demo_001", SourceKey: "TUSHARE", SourceItemID: "news_001", Title: "一季报超预期", Summary: "摘要1", PrimarySymbol: "600519.SH", Symbols: []string{"600519.SH"}, PublishedAt: "2026-03-23T09:00:00+08:00"},
			{ID: "sei_demo_002", SourceKey: "DOCFAST", SourceItemID: "news_002", Title: "渠道反馈积极", Summary: "摘要2", PrimarySymbol: "600519.SH", Symbols: []string{"600519.SH", "000858.SZ"}, PublishedAt: "2026-03-23T09:05:00+08:00"},
		},
		Entities: []model.StockEventEntity{
			{ID: "see_demo_company", EntityType: "COMPANY", EntityKey: "company:600519.SH", Label: "贵州茅台", Symbol: "600519.SH", Confidence: 0.98},
			{ID: "see_demo_topic", EntityType: "TOPIC", EntityKey: "topic:white-liquor", Label: "白酒", TopicLabel: "白酒", Confidence: 0.88},
		},
		Edges: []model.StockEventEdge{
			{ID: "seg_demo_001", SubjectEntityID: "see_demo_company", ObjectEntityID: "see_demo_topic", RelationType: "BELONGS_TO_TOPIC", Weight: 0.83},
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec(stockEventClusterUpsertPattern).
		WithArgs(
			cluster.ID,
			cluster.ClusterKey,
			cluster.EventType,
			cluster.Title,
			cluster.Summary,
			cluster.Source,
			cluster.PrimarySymbol,
			cluster.SectorLabel,
			cluster.TopicLabel,
			cluster.Status,
			cluster.ReviewStatus,
			cluster.NewsCount,
			cluster.Confidence,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stockEventItemDeletePattern).
		WithArgs(cluster.ID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(stockEventItemInsertPattern).
		WithArgs("sei_demo_001", cluster.ID, "TUSHARE", "news_001", "一季报超预期", "摘要1", "600519.SH", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stockEventItemInsertPattern).
		WithArgs("sei_demo_002", cluster.ID, "DOCFAST", "news_002", "渠道反馈积极", "摘要2", "600519.SH", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stockEventEntityDeletePattern).
		WithArgs(cluster.ID).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectExec(stockEventEntityInsertPattern).
		WithArgs("see_demo_company", cluster.ID, "COMPANY", "company:600519.SH", "贵州茅台", "600519.SH", "", "", 0.98, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stockEventEntityInsertPattern).
		WithArgs("see_demo_topic", cluster.ID, "TOPIC", "topic:white-liquor", "白酒", "", "", "白酒", 0.88, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stockEventEdgeDeletePattern).
		WithArgs(cluster.ID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(stockEventEdgeInsertPattern).
		WithArgs("seg_demo_001", cluster.ID, "see_demo_company", "see_demo_topic", "BELONGS_TO_TOPIC", 0.83, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	stored, err := repo.AdminUpsertStockEventCluster(cluster)
	if err != nil {
		t.Fatalf("upsert stock event cluster: %v", err)
	}
	if stored.ID != cluster.ID {
		t.Fatalf("expected cluster id %s, got %s", cluster.ID, stored.ID)
	}
	if len(stored.Items) != 2 || len(stored.Entities) != 2 || len(stored.Edges) != 1 {
		t.Fatalf("expected nested members to be preserved, got items=%d entities=%d edges=%d", len(stored.Items), len(stored.Entities), len(stored.Edges))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminGetStockEventClusterLoadsNestedTruthSource(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	reviewedAt := time.Date(2026, 3, 23, 10, 0, 0, 0, time.FixedZone("CST", 8*3600))
	publishedAt := time.Date(2026, 3, 23, 9, 30, 0, 0, time.FixedZone("CST", 8*3600))
	createdAt := time.Date(2026, 3, 23, 9, 0, 0, 0, time.FixedZone("CST", 8*3600))

	mock.ExpectQuery(stockEventClusterSelectPattern).
		WithArgs("sec_demo_001").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "cluster_key", "event_type", "title", "summary", "source", "primary_symbol", "sector_label", "topic_label", "cluster_status", "review_status", "news_count", "confidence_score", "metadata_json", "published_at", "created_at", "updated_at", "latest_review_id", "latest_reviewer", "latest_review_note", "latest_review_status", "latest_review_metadata", "latest_reviewed_at", "latest_review_created_at", "latest_review_updated_at",
		}).AddRow(
			"sec_demo_001", "earnings:600519.SH:2026Q1", "EARNINGS", "贵州茅台一季报超预期", "营收与利润均超市场预期", "MARKET_NEWS", "600519.SH", "消费", "白酒", "REVIEWED", "APPROVED", 2, 0.91, `{"source_count":2}`, publishedAt, createdAt, reviewedAt,
			"ser_demo_001", "reviewer_001", "确认进入白酒景气主线", "APPROVED", `{"manual":true}`, reviewedAt, reviewedAt, reviewedAt,
		))
	mock.ExpectQuery(stockEventItemSelectPattern).
		WithArgs("sec_demo_001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cluster_id", "source_key", "source_item_id", "title", "summary", "primary_symbol", "symbols_json", "metadata_json", "published_at", "created_at", "updated_at"}).
			AddRow("sei_demo_001", "sec_demo_001", "TUSHARE", "news_001", "一季报超预期", "摘要1", "600519.SH", `[
"600519.SH"]`, `{"sentiment":"positive"}`, publishedAt, createdAt, reviewedAt))
	mock.ExpectQuery(stockEventEntitySelectPattern).
		WithArgs("sec_demo_001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cluster_id", "entity_type", "entity_key", "label", "symbol", "sector_label", "topic_label", "confidence_score", "metadata_json", "created_at", "updated_at"}).
			AddRow("see_demo_company", "sec_demo_001", "COMPANY", "company:600519.SH", "贵州茅台", "600519.SH", "", "", 0.98, `{"source":"reviewed"}`, createdAt, reviewedAt))
	mock.ExpectQuery(stockEventEdgeSelectPattern).
		WithArgs("sec_demo_001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "cluster_id", "subject_entity_id", "object_entity_id", "relation_type", "relation_weight", "metadata_json", "created_at", "updated_at"}).
			AddRow("seg_demo_001", "sec_demo_001", "see_demo_company", "see_demo_topic", "BELONGS_TO_TOPIC", 0.83, `{"evidence":"reviewed"}`, createdAt, reviewedAt))

	cluster, err := repo.AdminGetStockEventCluster("sec_demo_001")
	if err != nil {
		t.Fatalf("get stock event cluster: %v", err)
	}
	if cluster.ID != "sec_demo_001" || cluster.ReviewStatus != "APPROVED" {
		t.Fatalf("unexpected cluster: %+v", cluster)
	}
	if cluster.LatestReview == nil || cluster.LatestReview.Reviewer != "reviewer_001" {
		t.Fatalf("expected latest review, got %+v", cluster.LatestReview)
	}
	if len(cluster.Items) != 1 || cluster.Items[0].Symbols[0] != "600519.SH" {
		t.Fatalf("unexpected cluster items: %+v", cluster.Items)
	}
	if len(cluster.Entities) != 1 || cluster.Entities[0].EntityKey != "company:600519.SH" {
		t.Fatalf("unexpected entities: %+v", cluster.Entities)
	}
	if len(cluster.Edges) != 1 || cluster.Edges[0].RelationType != "BELONGS_TO_TOPIC" {
		t.Fatalf("unexpected edges: %+v", cluster.Edges)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminCreateStockEventReviewTransitionsClusterStatusAndSupportsFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectBegin()
	mock.ExpectExec(stockEventClusterReviewUpdatePattern).
		WithArgs("REVIEWED", "APPROVED", "sec_demo_001").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(stockEventReviewInsertPattern).
		WithArgs(sqlmock.AnyArg(), "sec_demo_001", "APPROVED", "reviewer_001", "进入正式事件池", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	review, err := repo.AdminCreateStockEventReview(model.StockEventReview{
		ClusterID:      "sec_demo_001",
		ReviewStatus:   "APPROVED",
		Reviewer:       "reviewer_001",
		ReviewNote:     "进入正式事件池",
		ReviewMetadata: map[string]any{"manual": true},
	})
	if err != nil {
		t.Fatalf("create stock event review: %v", err)
	}
	if review.ClusterID != "sec_demo_001" || review.ReviewStatus != "APPROVED" {
		t.Fatalf("unexpected review result: %+v", review)
	}

	mock.ExpectQuery(stockEventClusterCountPattern).
		WithArgs("APPROVED", "EARNINGS", "HIGH", "600519.SH", "600519.SH", "消费", "白酒").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(stockEventClusterListPattern).
		WithArgs("APPROVED", "EARNINGS", "HIGH", "600519.SH", "600519.SH", "消费", "白酒", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "cluster_key", "event_type", "title", "summary", "source", "primary_symbol", "sector_label", "topic_label", "cluster_status", "review_status", "news_count", "confidence_score", "metadata_json", "published_at", "created_at", "updated_at", "latest_review_id", "latest_reviewer", "latest_review_note", "latest_review_status", "latest_review_metadata", "latest_reviewed_at", "latest_review_created_at", "latest_review_updated_at",
		}).AddRow(
			"sec_demo_001", "earnings:600519.SH:2026Q1", "EARNINGS", "贵州茅台一季报超预期", "营收与利润均超市场预期", "MARKET_NEWS", "600519.SH", "消费", "白酒", "REVIEWED", "APPROVED", 2, 0.91, `{"source_count":2,"review_priority":"HIGH"}`, time.Time{}, time.Time{}, time.Time{},
			"ser_demo_001", "reviewer_001", "进入正式事件池", "APPROVED", `{"manual":true}`, time.Time{}, time.Time{}, time.Time{},
		))

	items, total, err := repo.AdminListStockEventClusters(model.StockEventQuery{
		ReviewStatus:   "APPROVED",
		EventType:      "EARNINGS",
		ReviewPriority: "HIGH",
		Symbol:         "600519.SH",
		Sector:         "消费",
		Topic:          "白酒",
		Page:           1,
		PageSize:       20,
	})
	if err != nil {
		t.Fatalf("list stock event clusters: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 approved event cluster, got total=%d len=%d", total, len(items))
	}
	if items[0].PrimarySymbol != "600519.SH" || items[0].SectorLabel != "消费" || items[0].TopicLabel != "白酒" {
		t.Fatalf("unexpected filtered cluster: %+v", items[0])
	}
	if items[0].Metadata["review_priority"] != "HIGH" {
		t.Fatalf("expected review_priority HIGH, got %+v", items[0].Metadata)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminCreateStockEventReviewReturnsErrNoRowsWhenClusterMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectBegin()
	mock.ExpectExec(stockEventClusterReviewUpdatePattern).
		WithArgs("REVIEWED", "APPROVED", "sec_missing_001").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	_, err = repo.AdminCreateStockEventReview(model.StockEventReview{
		ClusterID:    "sec_missing_001",
		ReviewStatus: "APPROVED",
		Reviewer:     "reviewer_001",
		ReviewNote:   "事件不存在时不应写入 review",
	})
	if err == nil {
		t.Fatalf("expected sql.ErrNoRows when cluster is missing")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
