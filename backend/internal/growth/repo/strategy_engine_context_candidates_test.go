package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const stockSelectionContextCandidatesAutoQueryPattern = `(?s)SELECT t\.instrument_key,.*FROM market_daily_bar_truth t.*WHERE t\.asset_class = \? AND t\.trade_date = \?.*ORDER BY t\.turnover DESC, t\.instrument_key ASC LIMIT \?`

func TestLoadStrategyStockContextCandidatesDedupesMixedSymbolFormatsInAutoMode(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	selectedTradeDate := time.Date(2026, 2, 24, 0, 0, 0, 0, time.Local)
	mock.ExpectQuery(stockSelectionContextCandidatesAutoQueryPattern).
		WithArgs(marketAssetClassStock, "2026-02-24", 20).
		WillReturnRows(sqlmock.NewRows([]string{
			"instrument_key",
			"display_name",
			"selected_source_key",
			"volume",
			"turnover",
			"metadata_json",
		}).AddRow(
			"600036",
			"招商银行",
			"AKSHARE",
			1100000,
			990000000,
			`{"industry":"银行","sector":"金融"}`,
		).AddRow(
			"600036.SH",
			"招商银行",
			"MYSELF",
			1000000,
			980000000,
			`{"industry":"银行","sector":"金融"}`,
		).AddRow(
			"000333.SZ",
			"美的集团",
			"AKSHARE",
			1200000,
			860000000,
			`{"industry":"家电","sector":"消费"}`,
		))

	items, err := repo.loadStrategyStockContextCandidates(selectedTradeDate, nil, map[string]struct{}{}, 20)
	if err != nil {
		t.Fatalf("load candidates: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 deduped candidates, got %d (%+v)", len(items), items)
	}
	if items[0].Symbol != "600036.SH" {
		t.Fatalf("expected canonical ts_code variant to be kept, got %+v", items[0])
	}
	if items[1].Symbol != "000333.SZ" {
		t.Fatalf("expected unrelated symbol to remain, got %+v", items[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
