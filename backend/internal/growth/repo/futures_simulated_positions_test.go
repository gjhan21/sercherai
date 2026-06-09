package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestAdminListFuturesSimulatedPositions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM futures_simulated_positions WHERE 1=1`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT id, strategy_id, contract, name, direction, status`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "strategy_id", "contract", "name", "direction", "status",
			"open_date", "open_price", "current_price",
			"close_date", "close_price", "take_profit_price", "stop_loss_price",
			"quantity", "cost_basis", "close_value",
			"return_rate", "max_drawdown", "hold_days", "close_reason", "created_at",
		}).AddRow(
			"fp_001", "strat_001", "IF2606", "沪深300指数期货", "LONG", "HOLDING",
			"2026-06-09", 3500.0, 3550.0,
			"", 0.0, 3800.0, 3300.0,
			10.0, 35000.0, 0.0,
			0.0142, 0.0, 3, "", "2026-06-09T12:00:00Z",
		))

	items, total, err := repo.AdminListFuturesSimulatedPositions("", "", 1, 10)
	if err != nil {
		t.Fatalf("AdminListFuturesSimulatedPositions error: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
	if len(items) != 1 || items[0].Contract != "IF2606" {
		t.Errorf("unexpected items returned: %+v", items)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminGetFuturesSimulatedOverview(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM futures_simulated_positions`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM futures_simulated_positions WHERE status = 'HOLDING'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM futures_simulated_positions WHERE status = 'CLOSED'`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM futures_simulated_positions WHERE status = 'CLOSED' AND return_rate > 0`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT COALESCE\(AVG\(return_rate\), 0\) FROM futures_simulated_positions WHERE status = 'CLOSED'`).
		WillReturnRows(sqlmock.NewRows([]string{"avg_return"}).AddRow(0.05))

	mock.ExpectQuery(`SELECT COALESCE\(AVG\(hold_days\), 0\) FROM futures_simulated_positions WHERE status = 'CLOSED'`).
		WillReturnRows(sqlmock.NewRows([]string{"avg_hold_days"}).AddRow(10.0))

	mock.ExpectQuery(`SELECT COALESCE\(MAX\(return_rate\), 0\) FROM futures_simulated_positions WHERE status = 'CLOSED'`).
		WillReturnRows(sqlmock.NewRows([]string{"max_profit"}).AddRow(0.12))

	mock.ExpectQuery(`SELECT COALESCE\(MIN\(return_rate\), 0\) FROM futures_simulated_positions WHERE status = 'CLOSED'`).
		WillReturnRows(sqlmock.NewRows([]string{"max_loss"}).AddRow(-0.05))

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(cost_basis\), 0\)`).
		WillReturnRows(sqlmock.NewRows([]string{"total_cost", "total_current_value"}).
			AddRow(500000.0, 520000.0))

	overview, err := repo.AdminGetFuturesSimulatedOverview()
	if err != nil {
		t.Fatalf("AdminGetFuturesSimulatedOverview error: %v", err)
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

func TestAdminAutoOpenFuturesSimulatedPositions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	tradeDate := "2026-06-09"
	tDate, _ := time.Parse("2006-01-02", tradeDate)

	mock.ExpectQuery(`SELECT s.id, s.contract, s.name, s.direction, COALESCE\(g.take_profit_range, ''\), COALESCE\(g.stop_loss_range, ''\)`).
		WithArgs(tDate, tDate).
		WillReturnRows(sqlmock.NewRows([]string{"id", "contract", "name", "direction", "take_profit_range", "stop_loss_range"}).
			AddRow("strat_001", "IF2606", "沪深300指数期货", "LONG", "3800-4000", "3300-3400"))

	mock.ExpectQuery(`SELECT close_price FROM market_daily_bar_truth`).
		WithArgs("IF2606", "IF2606.%", tDate).
		WillReturnRows(sqlmock.NewRows([]string{"close_price"}).AddRow(3500.0))

	mock.ExpectExec(`INSERT INTO futures_simulated_positions`).
		WithArgs(sqlmock.AnyArg(), "strat_001", "IF2606", "沪深300指数期货", "LONG", tDate, 3500.0, 3500.0, 3800.0, 3300.0, 10.0, 35000.0, sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AdminAutoOpenFuturesSimulatedPositions(tradeDate)
	if err != nil {
		t.Fatalf("AdminAutoOpenFuturesSimulatedPositions error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminSettlementFuturesSimulatedPositions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	tradeDate := "2026-06-09"
	tDate, _ := time.Parse("2006-01-02", tradeDate)

	mock.ExpectQuery(`SELECT p.id, p.strategy_id, p.contract, p.direction, p.open_price, p.take_profit_price, p.stop_loss_price`).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "strategy_id", "contract", "direction", "open_price", "take_profit_price", "stop_loss_price", "quantity", "cost_basis", "hold_days", "valid_to",
		}).AddRow("fp_001", "strat_001", "IF2606", "LONG", 3500.0, 3800.0, 3300.0, 10.0, 35000.0, 3, "2026-06-15"))

	// Settlement high price >= take profit triggers CLOSED with TAKE_PROFIT
	mock.ExpectQuery(`SELECT open_price, high_price, low_price, close_price FROM market_daily_bar_truth`).
		WithArgs("IF2606", "IF2606.%", tDate).
		WillReturnRows(sqlmock.NewRows([]string{"open_price", "high_price", "low_price", "close_price"}).
			AddRow(3510.0, 3810.0, 3505.0, 3750.0))

	mock.ExpectQuery(`SELECT MAX\(close_price\)`).
		WithArgs("IF2606", "IF2606.%", "fp_001", tDate).
		WillReturnRows(sqlmock.NewRows([]string{"max_price"}).AddRow(3750.0))

	mock.ExpectExec(`UPDATE futures_simulated_positions SET status = 'CLOSED'`).
		WithArgs(3750.0, tDate, 3800.0, 3800.0*10.0, (3800.0-3500.0)/3500.0, (3750.0-3750.0)/3750.0, 4, "TAKE_PROFIT", "fp_001").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AdminSettlementFuturesSimulatedPositions(tradeDate)
	if err != nil {
		t.Fatalf("AdminSettlementFuturesSimulatedPositions error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
