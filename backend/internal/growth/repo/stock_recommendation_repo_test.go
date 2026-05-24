package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const stockRecommendationLatestTradeDateQueryPattern = `(?s)SELECT DATE_FORMAT\(MAX\(valid_from\), '%Y-%m-%d'\)\s+FROM stock_recommendations\s+WHERE status IN \('PUBLISHED', 'ACTIVE', 'TRACKING'\)`
const stockRecommendationLatestTradeDateBeforeQueryPattern = `(?s)SELECT DATE_FORMAT\(MAX\(valid_from\), '%Y-%m-%d'\)\s+FROM stock_recommendations\s+WHERE status IN \('PUBLISHED', 'ACTIVE', 'TRACKING'\)\s+AND DATE\(valid_from\) <= \?`
const stockRecommendationCountQueryPattern = `(?s)SELECT COUNT\(\*\)\s+FROM stock_recommendations`
const stockRecommendationListQueryPattern = `(?s)SELECT r\.id,\s*r\.symbol,\s*r\.name,\s*r\.score,\s*r\.risk_level,\s*COALESCE\(r\.position_range, ''\),\s*r\.valid_from,\s*r\.valid_to,\s*r\.status,\s*COALESCE\(r\.reason_summary, ''\),\s*COALESCE\(r\.source_type, ''\),\s*COALESCE\(r\.strategy_version, ''\),\s*COALESCE\(r\.performance_label, ''\),\s*COALESCE\(d\.take_profit, ''\),\s*COALESCE\(d\.stop_loss, ''\)\s+FROM stock_recommendations r\s+LEFT JOIN stock_reco_details d ON d\.reco_id = r\.id`

func TestListStockRecommendationsFiltersByTradeDateAndSortsByScore(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(stockRecommendationLatestTradeDateBeforeQueryPattern).
		WithArgs("2026-05-24").
		WillReturnRows(sqlmock.NewRows([]string{"trade_date"}).AddRow("2026-05-24"))
	mock.ExpectQuery(stockRecommendationCountQueryPattern).
		WithArgs("2026-05-24").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(stockRecommendationListQueryPattern).
		WithArgs("2026-05-24", 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "symbol", "name", "score", "risk_level", "position_range", "valid_from", "valid_to", "status", "reason_summary", "source_type", "strategy_version", "performance_label", "take_profit", "stop_loss",
		}).
			AddRow("sr_002", "300750.SZ", "宁德时代", 92.5, "MEDIUM", "10%-15%", time.Date(2026, 5, 24, 9, 0, 0, 0, time.Local), time.Date(2026, 5, 25, 15, 0, 0, 0, time.Local), "PUBLISHED", "量化评分高", "SYSTEM", "daily-v1", "ESTIMATED", "上涨10%-15%分批止盈", "回撤5%止损").
			AddRow("sr_001", "600519.SH", "贵州茅台", 88.4, "LOW", "8%-10%", time.Date(2026, 5, 24, 9, 0, 0, 0, time.Local), time.Date(2026, 5, 25, 15, 0, 0, 0, time.Local), "TRACKING", "估值低位", "SYSTEM", "daily-v1", "ESTIMATED", "上涨8%-12%分批止盈", "回撤3%止损"))

	items, total, err := repo.ListStockRecommendations("u_demo_001", "2026-05-24", 1, 10)
	if err != nil {
		t.Fatalf("ListStockRecommendations() error = %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected total=2 len=2, got total=%d len=%d", total, len(items))
	}
	if items[0].Score < items[1].Score {
		t.Fatalf("expected score-desc ordering, got %+v", items)
	}
	if items[0].ValidFrom[:10] != "2026-05-24" || items[1].ValidFrom[:10] != "2026-05-24" {
		t.Fatalf("expected trade-date constrained rows, got %+v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListStockRecommendationsIncludesDetailFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(stockRecommendationLatestTradeDateQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"trade_date"}).AddRow("2026-05-24"))
	mock.ExpectQuery(stockRecommendationCountQueryPattern).
		WithArgs("2026-05-24").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(stockRecommendationListQueryPattern).
		WithArgs("2026-05-24", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "symbol", "name", "score", "risk_level", "position_range", "valid_from", "valid_to", "status", "reason_summary", "source_type", "strategy_version", "performance_label", "take_profit", "stop_loss",
		}).
			AddRow("sr_003", "002594.SZ", "比亚迪", 85.3, "MEDIUM", "7%-11%", time.Date(2026, 5, 24, 9, 0, 0, 0, time.Local), time.Date(2026, 5, 25, 15, 0, 0, 0, time.Local), "ACTIVE", "销量韧性", "SYSTEM", "daily-v1", "ESTIMATED", "上涨10%-15%分批止盈", "回撤5%止损"))

	items, total, err := repo.ListStockRecommendations("u_demo_001", "", 1, 20)
	if err != nil {
		t.Fatalf("ListStockRecommendations() error = %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected single row, got total=%d len=%d", total, len(items))
	}
	if items[0].TakeProfit == "" || items[0].StopLoss == "" {
		t.Fatalf("expected detail fields to be populated, got %+v", items[0])
	}
	if items[0].SourceType == "" || items[0].StrategyVersion == "" || items[0].PerformanceLabel == "" {
		t.Fatalf("expected source/version/performance fields, got %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListStockRecommendationsFallsBackToLatestTradeDate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(stockRecommendationLatestTradeDateQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"trade_date"}).AddRow("2026-05-23"))
	mock.ExpectQuery(stockRecommendationCountQueryPattern).
		WithArgs("2026-05-23").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(stockRecommendationListQueryPattern).
		WithArgs("2026-05-23", 5, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "symbol", "name", "score", "risk_level", "position_range", "valid_from", "valid_to", "status", "reason_summary", "source_type", "strategy_version", "performance_label", "take_profit", "stop_loss",
		}).
			AddRow("sr_010", "688981.SH", "中芯国际", 80.8, "HIGH", "5%-8%", time.Date(2026, 5, 23, 9, 0, 0, 0, time.Local), time.Date(2026, 5, 24, 15, 0, 0, 0, time.Local), "PUBLISHED", "国产替代", "SYSTEM", "daily-v1", "ESTIMATED", "上涨12%-18%动态止盈", "回撤7%止损"))

	items, total, err := repo.ListStockRecommendations("u_demo_001", "", 1, 5)
	if err != nil {
		t.Fatalf("ListStockRecommendations() error = %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected single row, got total=%d len=%d", total, len(items))
	}
	if items[0].ValidFrom[:10] != "2026-05-23" {
		t.Fatalf("expected fallback latest trade date to be applied, got %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListStockRecommendationsFallsBackToLatestAvailableTradeDateOnExplicitTradeDate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(stockRecommendationLatestTradeDateBeforeQueryPattern).
		WithArgs("2026-05-24").
		WillReturnRows(sqlmock.NewRows([]string{"trade_date"}).AddRow("2026-05-22"))
	mock.ExpectQuery(stockRecommendationCountQueryPattern).
		WithArgs("2026-05-22").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(stockRecommendationListQueryPattern).
		WithArgs("2026-05-22", 6, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "symbol", "name", "score", "risk_level", "position_range", "valid_from", "valid_to", "status", "reason_summary", "source_type", "strategy_version", "performance_label", "take_profit", "stop_loss",
		}).
			AddRow("sr_018", "002371.SZ", "北方华创", 91.2, "MEDIUM", "8%-12%", time.Date(2026, 5, 22, 9, 0, 0, 0, time.Local), time.Date(2026, 5, 23, 15, 0, 0, 0, time.Local), "PUBLISHED", "设备景气延续", "SYSTEM", "daily-v1", "ESTIMATED", "上涨10%-15%分批止盈", "回撤5%止损"))

	items, total, err := repo.ListStockRecommendations("u_demo_001", "2026-05-24", 1, 6)
	if err != nil {
		t.Fatalf("ListStockRecommendations() error = %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected single fallback row, got total=%d len=%d", total, len(items))
	}
	if items[0].ValidFrom[:10] != "2026-05-22" {
		t.Fatalf("expected explicit trade date to fall back to 2026-05-22, got %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
