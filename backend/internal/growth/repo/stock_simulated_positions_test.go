package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestAdminListStockSimulatedPositions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM stock_simulated_positions WHERE 1=1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT id, reco_id, symbol, name, status, open_date, open_price, current_price`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "reco_id", "symbol", "name", "status", "open_date", "open_price", "current_price",
			"close_date", "close_price", "take_profit_price", "stop_loss_price", "quantity", "cost_basis",
			"close_value", "return_rate", "max_drawdown", "hold_days", "close_reason", "created_at",
		}).AddRow(
			"sp_001", "reco_001", "600519.SH", "贵州茅台", "HOLDING", time.Now(), 1700.0, 1750.0,
			"", 0.0, 1850.0, 1600.0, 1000.0, 1700000.0, 0.0, 0.0294, 0.0, 3, "", time.Now(),
		))

	items, total, err := repo.AdminListStockSimulatedPositions("", "", 1, 10)
	if err != nil {
		t.Fatalf("AdminListStockSimulatedPositions error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(items) != 1 || items[0].Symbol != "600519.SH" {
		t.Errorf("unexpected items returned: %+v", items)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminGetStockSimulatedOverview(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM stock_simulated_positions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM stock_simulated_positions WHERE status = 'HOLDING'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM stock_simulated_positions WHERE status = 'CLOSED'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM stock_simulated_positions WHERE status = 'CLOSED' AND return_rate > 0`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT AVG\(return_rate\), MAX\(return_rate\), MIN\(return_rate\), AVG\(hold_days\)`).
		WillReturnRows(sqlmock.NewRows([]string{"avg_return", "max_profit", "max_loss", "avg_hold_days"}).
			AddRow(0.05, 0.12, -0.05, 10.0))

	mock.ExpectQuery(`SELECT SUM\(cost_basis\)`).
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_current_value"}).
			AddRow(500000.0, 520000.0))

	overview, err := repo.AdminGetStockSimulatedOverview()
	if err != nil {
		t.Fatalf("AdminGetStockSimulatedOverview error: %v", err)
	}

	if overview.TotalTrades != 5 {
		t.Errorf("expected total trades 5, got %d", overview.TotalTrades)
	}
	if overview.ActiveHoldings != 2 {
		t.Errorf("expected active holdings 2, got %d", overview.ActiveHoldings)
	}
	if overview.WinRate != 2.0/3.0 {
		t.Errorf("expected win rate %f, got %f", 2.0/3.0, overview.WinRate)
	}
	if overview.AverageReturn != 0.05 {
		t.Errorf("expected average return 0.05, got %f", overview.AverageReturn)
	}
	if overview.TotalReturn != 0.04 { // (520000 - 500000) / 500000 = 0.04
		t.Errorf("expected total return 0.04, got %f", overview.TotalReturn)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminAutoOpenSimulatedPositions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	tradeDate := "2026-06-09"
	tDate, _ := time.Parse("2006-01-02", tradeDate)

	mock.ExpectQuery(`SELECT id, symbol, name, take_profit, stop_loss FROM stock_recommendations`).
		WithArgs(tDate, tDate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "symbol", "name", "take_profit", "stop_loss"}).
			AddRow("reco_001", "600519.SH", "贵州茅台", "1850.00", "1600.00"))

	mock.ExpectQuery(`SELECT close_price FROM market_daily_bar_truth`).
		WithArgs("600519.SH", "600519.SH.%", tDate).
		WillReturnRows(sqlmock.NewRows([]string{"close_price"}).AddRow(1700.0))

	mock.ExpectExec(`INSERT INTO stock_simulated_positions`).
		WithArgs(sqlmock.AnyArg(), "reco_001", "600519.SH", "贵州茅台", tDate, 1700.0, 1700.0, 1850.0, 1600.0, 1000.0, 1700000.0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AdminAutoOpenSimulatedPositions(tradeDate)
	if err != nil {
		t.Fatalf("AdminAutoOpenSimulatedPositions error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminSettlementSimulatedPositions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	tradeDate := "2026-06-09"
	tDate, _ := time.Parse("2006-01-02", tradeDate)

	mock.ExpectQuery(`SELECT p\.id, p\.reco_id, p\.symbol, p\.open_price, p\.take_profit_price, p\.stop_loss_price`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "reco_id", "symbol", "open_price", "take_profit_price", "stop_loss_price", "quantity", "hold_days", "valid_to",
		}).AddRow("sp_001", "reco_001", "600519.SH", 1700.0, 1850.0, 1600.0, 1000.0, 3, "2026-06-15"))

	// Settlement high price >= take profit triggers CLOSED with TAKE_PROFIT
	mock.ExpectQuery(`SELECT open_price, high_price, low_price, close_price FROM market_daily_bar_truth`).
		WithArgs("600519.SH", "600519.SH.%", tDate).
		WillReturnRows(sqlmock.NewRows([]string{"open_price", "high_price", "low_price", "close_price"}).
			AddRow(1710.0, 1860.0, 1705.0, 1800.0))

	mock.ExpectQuery(`SELECT MAX\(close_price\) FROM market_daily_bar_truth`).
		WithArgs("600519.SH", "600519.SH.%", "sp_001", tDate).
		WillReturnRows(sqlmock.NewRows([]string{"max_price"}).AddRow(1800.0))

	mock.ExpectExec(`UPDATE stock_simulated_positions SET status = 'CLOSED'`).
		WithArgs(1800.0, tDate, 1850.0, 1850.0*1000.0, (1850.0-1700.0)/1700.0, (1800.0-1800.0)/1800.0, 4, "TAKE_PROFIT", "sp_001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AdminSettlementSimulatedPositions(tradeDate)
	if err != nil {
		t.Fatalf("AdminSettlementSimulatedPositions error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
