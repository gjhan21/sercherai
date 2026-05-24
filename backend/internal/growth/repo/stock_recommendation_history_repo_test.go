package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const stockRecommendationHistoryCountQueryPattern = `(?s)SELECT COUNT\(\*\)\s+FROM stock_recommendations r`
const stockRecommendationHistoryListQueryPattern = `(?s)SELECT r\.id,\s*r\.symbol,\s*r\.name,\s*r\.score,\s*r\.risk_level,\s*COALESCE\(r\.position_range, ''\),\s*r\.valid_from,\s*r\.valid_to,\s*r\.status,\s*COALESCE\(r\.source_type, ''\),\s*COALESCE\(r\.strategy_version, ''\),\s*COALESCE\(r\.performance_label, ''\),\s*COALESCE\(d\.take_profit, ''\),\s*COALESCE\(d\.stop_loss, ''\)\s+FROM stock_recommendations r\s+LEFT JOIN stock_reco_details d ON d\.reco_id = r\.id`
const stockRecommendationHistoryQuotesQueryPattern = `(?s)SELECT symbol,\s*trade_date,\s*close_price\s+FROM stock_market_quotes`

func TestListStockRecommendationHistoryBuildsRealPerformanceItems(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	validFrom := time.Date(2026, 5, 20, 9, 0, 0, 0, time.Local)
	validTo := time.Date(2026, 5, 24, 15, 0, 0, 0, time.Local)

	mock.ExpectQuery(stockRecommendationHistoryCountQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(stockRecommendationHistoryListQueryPattern).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "symbol", "name", "score", "risk_level", "position_range", "valid_from", "valid_to", "status", "source_type", "strategy_version", "performance_label", "take_profit", "stop_loss",
		}).AddRow(
			"sr_hist_001", "600519.SH", "贵州茅台", 88.0, "MEDIUM", "10%-15%", validFrom, validTo, "HIT_TAKE_PROFIT", "SYSTEM", "daily-v2", "OUTPERFORM", "上涨 10% 分批止盈", "回撤 5% 止损",
		))
	mock.ExpectQuery(stockRecommendationHistoryQuotesQueryPattern).
		WithArgs("600519.SH", "2026-05-20", "2026-05-24").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "close_price"}).
			AddRow("600519.SH", time.Date(2026, 5, 20, 0, 0, 0, 0, time.Local), 10.0).
			AddRow("600519.SH", time.Date(2026, 5, 21, 0, 0, 0, 0, time.Local), 12.0).
			AddRow("600519.SH", time.Date(2026, 5, 22, 0, 0, 0, 0, time.Local), 11.0).
			AddRow("600519.SH", time.Date(2026, 5, 23, 0, 0, 0, 0, time.Local), 9.0).
			AddRow("600519.SH", time.Date(2026, 5, 24, 0, 0, 0, 0, time.Local), 13.0))

	items, summary, total, err := repo.ListStockRecommendationHistory("u_demo_001", "", "", "", 1, 20)
	if err != nil {
		t.Fatalf("ListStockRecommendationHistory() error = %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected total=1 len=1, got total=%d len=%d", total, len(items))
	}
	if items[0].EntryPrice != 10 || items[0].LatestPrice != 13 {
		t.Fatalf("expected real entry/latest prices, got %+v", items[0])
	}
	if items[0].ReturnPct != 30 {
		t.Fatalf("expected return 30%%, got %+v", items[0])
	}
	if items[0].MaxDrawdownPct != -25 {
		t.Fatalf("expected max drawdown -25%%, got %+v", items[0])
	}
	if items[0].Outcome != "success" {
		t.Fatalf("expected success outcome, got %+v", items[0])
	}
	if summary.TotalCount != 1 || summary.SuccessCount != 1 {
		t.Fatalf("expected summary to reflect item, got %+v", summary)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListStockRecommendationHistoryBuildsSummaryFromAllMatchedRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	firstFrom := time.Date(2026, 5, 20, 9, 0, 0, 0, time.Local)
	firstTo := time.Date(2026, 5, 24, 15, 0, 0, 0, time.Local)
	secondFrom := time.Date(2026, 5, 19, 9, 0, 0, 0, time.Local)
	secondTo := time.Date(2026, 5, 24, 15, 0, 0, 0, time.Local)

	mock.ExpectQuery(stockRecommendationHistoryCountQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(stockRecommendationHistoryListQueryPattern).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "symbol", "name", "score", "risk_level", "position_range", "valid_from", "valid_to", "status", "source_type", "strategy_version", "performance_label", "take_profit", "stop_loss",
		}).
			AddRow("sr_hist_001", "600519.SH", "贵州茅台", 88.0, "MEDIUM", "10%-15%", firstFrom, firstTo, "HIT_TAKE_PROFIT", "SYSTEM", "daily-v2", "OUTPERFORM", "上涨 10% 分批止盈", "回撤 5% 止损").
			AddRow("sr_hist_002", "300750.SZ", "宁德时代", 76.0, "HIGH", "5%-8%", secondFrom, secondTo, "HIT_STOP_LOSS", "SYSTEM", "daily-v2", "UNDERPERFORM", "上涨 8% 分批止盈", "回撤 4% 止损"))
	mock.ExpectQuery(stockRecommendationHistoryQuotesQueryPattern).
		WithArgs("300750.SZ", "600519.SH", "2026-05-19", "2026-05-24").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "close_price"}).
			AddRow("300750.SZ", time.Date(2026, 5, 19, 0, 0, 0, 0, time.Local), 20.0).
			AddRow("300750.SZ", time.Date(2026, 5, 20, 0, 0, 0, 0, time.Local), 18.0).
			AddRow("300750.SZ", time.Date(2026, 5, 21, 0, 0, 0, 0, time.Local), 17.0).
			AddRow("300750.SZ", time.Date(2026, 5, 24, 0, 0, 0, 0, time.Local), 16.0).
			AddRow("600519.SH", time.Date(2026, 5, 20, 0, 0, 0, 0, time.Local), 10.0).
			AddRow("600519.SH", time.Date(2026, 5, 21, 0, 0, 0, 0, time.Local), 11.0).
			AddRow("600519.SH", time.Date(2026, 5, 24, 0, 0, 0, 0, time.Local), 13.0))

	_, summary, total, err := repo.ListStockRecommendationHistory("u_demo_001", "", "", "", 1, 20)
	if err != nil {
		t.Fatalf("ListStockRecommendationHistory() error = %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if summary.TotalCount != 2 || summary.SuccessCount != 1 || summary.FailCount != 1 {
		t.Fatalf("expected success/fail summary, got %+v", summary)
	}
	if summary.WinRate != 50 {
		t.Fatalf("expected 50%% win rate, got %+v", summary)
	}
	if summary.AvgReturnPct != 5 {
		t.Fatalf("expected 5%% average return, got %+v", summary)
	}
	if summary.MaxReturnPct != 30 || summary.MaxDrawdownPct != -20 {
		t.Fatalf("expected extreme stats from all rows, got %+v", summary)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListStockRecommendationHistoryMapsOutcomeFromStatusAndPerformanceLabel(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	validFrom := time.Date(2026, 5, 20, 9, 0, 0, 0, time.Local)
	validTo := time.Date(2026, 5, 24, 15, 0, 0, 0, time.Local)

	mock.ExpectQuery(stockRecommendationHistoryCountQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(stockRecommendationHistoryListQueryPattern).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "symbol", "name", "score", "risk_level", "position_range", "valid_from", "valid_to", "status", "source_type", "strategy_version", "performance_label", "take_profit", "stop_loss",
		}).
			AddRow("sr_hist_001", "600519.SH", "贵州茅台", 88.0, "MEDIUM", "10%-15%", validFrom, validTo, "REVIEWED", "SYSTEM", "daily-v2", "OUTPERFORM", "上涨 10% 分批止盈", "回撤 5% 止损").
			AddRow("sr_hist_002", "300750.SZ", "宁德时代", 76.0, "HIGH", "5%-8%", validFrom, validTo, "REVIEWED", "SYSTEM", "daily-v2", "UNDERPERFORM", "上涨 8% 分批止盈", "回撤 4% 止损").
			AddRow("sr_hist_003", "002594.SZ", "比亚迪", 82.0, "LOW", "8%-10%", validFrom, validTo, "TRACKING", "SYSTEM", "daily-v2", "ESTIMATED", "上涨 6% 分批止盈", "回撤 3% 止损"))
	mock.ExpectQuery(stockRecommendationHistoryQuotesQueryPattern).
		WithArgs("002594.SZ", "300750.SZ", "600519.SH", "2026-05-20", "2026-05-24").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "close_price"}).
			AddRow("002594.SZ", time.Date(2026, 5, 20, 0, 0, 0, 0, time.Local), 10.0).
			AddRow("002594.SZ", time.Date(2026, 5, 24, 0, 0, 0, 0, time.Local), 10.0).
			AddRow("300750.SZ", time.Date(2026, 5, 20, 0, 0, 0, 0, time.Local), 20.0).
			AddRow("300750.SZ", time.Date(2026, 5, 24, 0, 0, 0, 0, time.Local), 18.0).
			AddRow("600519.SH", time.Date(2026, 5, 20, 0, 0, 0, 0, time.Local), 10.0).
			AddRow("600519.SH", time.Date(2026, 5, 24, 0, 0, 0, 0, time.Local), 11.0))

	items, _, _, err := repo.ListStockRecommendationHistory("u_demo_001", "", "", "", 1, 20)
	if err != nil {
		t.Fatalf("ListStockRecommendationHistory() error = %v", err)
	}
	if items[0].Outcome != "success" {
		t.Fatalf("expected reviewed outperform to map to success, got %+v", items[0])
	}
	if items[1].Outcome != "fail" {
		t.Fatalf("expected reviewed underperform to map to fail, got %+v", items[1])
	}
	if items[2].Outcome != "ongoing" {
		t.Fatalf("expected tracking to map to ongoing, got %+v", items[2])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
