package repo

import (
	"database/sql"
	"reflect"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"sercherai/backend/internal/growth/model"
)

const strategyStockContextTradeDateQueryPattern = `SELECT MAX\(trade_date\)\s+FROM market_daily_bar_truth`
const strategyStockContextCandidatesQueryPattern = `SELECT t\.instrument_key,\s+COALESCE\(NULLIF\(mi\.display_name, ''\), t\.instrument_key\) AS display_name,\s+t\.selected_source_key,\s+COALESCE\(t\.volume, 0\),\s+COALESCE\(t\.turnover, 0\),\s+COALESCE\(CAST\(mi\.metadata_json AS CHAR\), ''\)\s+FROM market_daily_bar_truth`
const strategyStockContextListingDateQueryPattern = `SELECT instrument_key, MIN\(trade_date\)\s+FROM market_daily_bar_truth`
const strategyStockContextCoverageStartQueryPattern = `SELECT MIN\(trade_date\)\s+FROM market_daily_bar_truth\s+WHERE asset_class = \?`
const strategyStockContextHistoryQueryPattern = `SELECT instrument_key, trade_date, open_price, high_price, low_price, close_price, prev_close_price, volume, turnover\s+FROM market_daily_bar_truth`
const strategyStockContextDailyBasicQueryPattern = `SELECT t\.symbol, t\.trade_date, t\.turnover_rate, t\.volume_ratio, t\.pe_ttm, t\.pb, t\.total_mv, t\.circ_mv, t\.source_key\s+FROM stock_daily_basic`
const strategyStockContextMoneyflowQueryPattern = `SELECT t\.symbol, t\.trade_date, t\.net_mf_amount, t\.buy_lg_amount, t\.sell_lg_amount, t\.buy_elg_amount, t\.sell_elg_amount, t\.source_key\s+FROM stock_moneyflow_daily`
const strategyStockContextNewsQueryPattern = `SELECT primary_symbol, symbols_json, title\s+FROM market_news_items`
const strategyFuturesContextTradeDateQueryPattern = `SELECT MAX\(trade_date\)\s+FROM market_daily_bar_truth`
const strategyFuturesContextCandidatesQueryPattern = `SELECT t\.instrument_key,\s+SUBSTRING_INDEX\(t\.instrument_key, '\.', 1\) AS contract_key,\s+COALESCE\(NULLIF\(mi\.display_name, ''\), SUBSTRING_INDEX\(t\.instrument_key, '\.', 1\)\) AS display_name,\s+t\.selected_source_key\s+FROM market_daily_bar_truth`
const strategyFuturesContextHistoryQueryPattern = `SELECT instrument_key, trade_date, close_price, prev_close_price, settle_price, prev_settle_price, volume, turnover, open_interest\s+FROM market_daily_bar_truth`
const strategyFuturesContextSupplementQueryPattern = `SELECT instrument_key, trade_date, source_key, close_price, settle_price, prev_settle_price, turnover, open_interest\s+FROM market_daily_bars`
const strategyFuturesContextNewsQueryPattern = `SELECT primary_symbol, symbols_json, title\s+FROM market_news_items`
const strategyFuturesContextTermStructureQueryPattern = `SELECT instrument_key, close_price\s+FROM market_daily_bar_truth`
const strategyFuturesContextInventoryQueryPattern = `SELECT symbol, trade_date, warehouse, area, brand, place, grade, receipt_volume, previous_volume, change_volume\s+FROM futures_inventory_snapshots`
const strategyFuturesContextSpreadQueryPattern = `SELECT contract_a, contract_b, percentile, status\s+FROM arbitrage_recos`

func TestResolveStrategyStockSelectionModeKeepsLegacySeedCompatibility(t *testing.T) {
	if got := resolveStrategyStockSelectionMode("", []string{"600519.SH"}, nil); got != "MANUAL" {
		t.Fatalf("expected MANUAL for legacy seed payload, got %s", got)
	}
	if got := resolveStrategyStockSelectionMode("", nil, []string{"300750.SZ"}); got != "DEBUG" {
		t.Fatalf("expected DEBUG for debug symbols, got %s", got)
	}
	if got := resolveStrategyStockSelectionMode("", nil, nil); got != "AUTO" {
		t.Fatalf("expected AUTO when no symbols are provided, got %s", got)
	}
}

func TestStrategyEngineStockSelectionContextRequestNormalizedBackfillsDefaults(t *testing.T) {
	got := (model.StrategyEngineStockSelectionContextRequest{
		TradeDate:        " 2026-03-21 ",
		MarketScope:      "CN_A_MAIN",
		SeedSymbols:      []string{" 600519.sh ", "600519.SH"},
		ExcludedSymbols:  []string{" 300750.sz ", "300750.SZ"},
		MinListingDays:   0,
		MinAvgTurnover:   0,
		SelectionMode:    "",
		DebugSeedSymbols: nil,
	}).Normalized()

	if got.TradeDate != "2026-03-21" {
		t.Fatalf("expected trimmed trade date, got %q", got.TradeDate)
	}
	if got.ProfileID != model.StrategyEngineDefaultStockSelectionProfileID {
		t.Fatalf("expected default profile id, got %q", got.ProfileID)
	}
	if got.SelectionMode != model.StrategyEngineStockSelectionModeManual {
		t.Fatalf("expected MANUAL mode, got %q", got.SelectionMode)
	}
	if got.UniverseScope != "CN_A_MAIN" || got.MarketScope != "CN_A_MAIN" {
		t.Fatalf("expected market/universe scope to align, got universe=%q market=%q", got.UniverseScope, got.MarketScope)
	}
	if got.MinListingDays != model.StrategyEngineDefaultStockMinListingDays {
		t.Fatalf("expected default min listing days, got %d", got.MinListingDays)
	}
	if got.MinAvgTurnover != model.StrategyEngineDefaultStockMinAvgTurnover {
		t.Fatalf("expected default min avg turnover, got %v", got.MinAvgTurnover)
	}
	if !reflect.DeepEqual(got.SeedSymbols, []string{"600519.SH"}) {
		t.Fatalf("unexpected normalized seed symbols: %#v", got.SeedSymbols)
	}
	if !reflect.DeepEqual(got.ExcludedSymbols, []string{"300750.SZ"}) {
		t.Fatalf("unexpected normalized excluded symbols: %#v", got.ExcludedSymbols)
	}
}

func TestStrategyEngineStockSelectionContextRequestNormalizedPrefersDebugSymbols(t *testing.T) {
	got := (model.StrategyEngineStockSelectionContextRequest{
		DebugSeedSymbols: []string{" 300750.sz ", "300750.SZ"},
		SeedSymbols:      []string{"600519.SH"},
	}).Normalized()

	if got.SelectionMode != model.StrategyEngineStockSelectionModeDebug {
		t.Fatalf("expected DEBUG mode, got %q", got.SelectionMode)
	}
	if got.UniverseScope != model.StrategyEngineDefaultStockUniverseScope {
		t.Fatalf("expected default universe scope, got %q", got.UniverseScope)
	}
	if got.MarketScope != model.StrategyEngineDefaultStockUniverseScope {
		t.Fatalf("expected market scope fallback, got %q", got.MarketScope)
	}
	if !reflect.DeepEqual(got.DebugSeedSymbols, []string{"300750.SZ"}) {
		t.Fatalf("unexpected normalized debug seed symbols: %#v", got.DebugSeedSymbols)
	}
}

func TestResolveStrategyStockContextSymbolsUsesDebugSymbolsFirst(t *testing.T) {
	got := resolveStrategyStockContextSymbols("DEBUG", []string{"600519.SH"}, []string{"300750.SZ", "300750.SZ"})
	if len(got) != 1 || got[0] != "300750.SZ" {
		t.Fatalf("expected debug symbols to win, got %#v", got)
	}
}

func TestBuildStrategyEngineStockSelectionContextUsesTruthBarsAndNews(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	selectedTradeDate := time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(strategyStockContextTradeDateQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-19", "600519.SH", "300750.SZ").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(selectedTradeDate))

	mock.ExpectQuery(strategyStockContextCandidatesQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-18", "600519.SH", "300750.SZ").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "display_name", "selected_source_key", "volume", "turnover", "metadata_json"}).
			AddRow("600519.SH", "贵州茅台", "TUSHARE", 1000000, 980000000, `{"industry":"白酒","sector":"消费","theme_tags":["高股息","消费龙头"],"risk_flags":["机构抱团"]}`).
			AddRow("300750.SZ", "宁德时代", "TUSHARE", 900000, 880000000, ""))

	mock.ExpectQuery(strategyStockContextHistoryQueryPattern).
		WithArgs(marketAssetClassStock, "600519.SH", "300750.SZ", "2026-01-17", "2026-03-18").
		WillReturnRows(buildStrategyStockContextHistoryRows())

	mock.ExpectQuery(strategyStockContextDailyBasicQueryPattern).
		WithArgs("600519.SH", "300750.SZ", "2026-03-18", "600519.SH", "300750.SZ").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "turnover_rate", "volume_ratio", "pe_ttm", "pb", "total_mv", "circ_mv", "source_key"}).
			AddRow("600519.SH", selectedTradeDate, 0.82, 1.18, 25.6, 9.1, 1000000, 800000, "TUSHARE").
			AddRow("300750.SZ", selectedTradeDate, 1.36, 1.42, 32.4, 7.4, 1200000, 900000, "TUSHARE"))

	mock.ExpectQuery(strategyStockContextMoneyflowQueryPattern).
		WithArgs("600519.SH", "300750.SZ", "2026-03-18", "600519.SH", "300750.SZ").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "net_mf_amount", "buy_lg_amount", "sell_lg_amount", "buy_elg_amount", "sell_elg_amount", "source_key"}).
			AddRow("600519.SH", selectedTradeDate, 9800.5, 0, 0, 0, 0, "TUSHARE").
			AddRow("300750.SZ", selectedTradeDate, 15200.7, 0, 0, 0, 0, "TUSHARE"))

	mock.ExpectQuery(strategyStockContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}).
			AddRow(sql.NullString{String: "600519.SH", Valid: true}, sql.NullString{String: "", Valid: false}, "公司增长签约").
			AddRow(sql.NullString{String: "600519.SH", Valid: true}, sql.NullString{String: "", Valid: false}, "公司风险提示").
			AddRow(sql.NullString{String: "", Valid: false}, sql.NullString{String: `["300750.SZ"]`, Valid: true}, "宁德时代中标创新高"))

	ctx, err := repo.BuildStrategyEngineStockSelectionContext(model.StrategyEngineStockSelectionContextRequest{
		TradeDate:   "2026-03-19",
		SeedSymbols: []string{"600519.SH", "300750.SZ"},
		Limit:       2,
	})
	if err != nil {
		t.Fatalf("BuildStrategyEngineStockSelectionContext returned error: %v", err)
	}
	if ctx.Meta.SelectedTradeDate != "2026-03-18" {
		t.Fatalf("expected selected trade date 2026-03-18, got %s", ctx.Meta.SelectedTradeDate)
	}
	if ctx.Meta.PriceSource != "TUSHARE" {
		t.Fatalf("expected TUSHARE price source, got %s", ctx.Meta.PriceSource)
	}
	if ctx.Meta.NewsWindowDays != 14 {
		t.Fatalf("expected 14-day news window, got %d", ctx.Meta.NewsWindowDays)
	}
	if len(ctx.Meta.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", ctx.Meta.Warnings)
	}
	if len(ctx.Seeds) != 2 {
		t.Fatalf("expected 2 seeds, got %d", len(ctx.Seeds))
	}
	if ctx.Seeds[0].Symbol != "600519.SH" || ctx.Seeds[0].TurnoverRate != 0.82 {
		t.Fatalf("unexpected first seed: %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].NewsHeat != 2 || ctx.Seeds[0].PositiveNewsRate != 0.5 {
		t.Fatalf("unexpected first seed news fields: %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].ListingDays <= 0 || ctx.Seeds[0].AvgTurnover20 <= 0 {
		t.Fatalf("expected derived listing/turnover fields, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].Industry != "白酒" || ctx.Seeds[0].Sector != "消费" {
		t.Fatalf("expected metadata industry/sector fields, got %+v", ctx.Seeds[0])
	}
	if !reflect.DeepEqual(ctx.Seeds[0].ThemeTags, []string{"高股息", "消费龙头"}) {
		t.Fatalf("unexpected theme tags: %#v", ctx.Seeds[0].ThemeTags)
	}
	if !reflect.DeepEqual(ctx.Seeds[0].RiskFlags, []string{"机构抱团"}) {
		t.Fatalf("unexpected risk flags: %#v", ctx.Seeds[0].RiskFlags)
	}
	if ctx.Seeds[1].Symbol != "300750.SZ" || ctx.Seeds[1].NewsHeat != 1 {
		t.Fatalf("unexpected second seed: %+v", ctx.Seeds[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineStockSelectionContextFallsBackToNeutralNewsSignals(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	selectedTradeDate := time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(strategyStockContextTradeDateQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-19", "600519.SH").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(selectedTradeDate))
	mock.ExpectQuery(strategyStockContextCandidatesQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-18", "600519.SH").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "display_name", "selected_source_key", "volume", "turnover", "metadata_json"}).
			AddRow("600519.SH", "贵州茅台", "TUSHARE", 1000000, 980000000, ""))
	mock.ExpectQuery(strategyStockContextHistoryQueryPattern).
		WithArgs(marketAssetClassStock, "600519.SH", "2026-01-17", "2026-03-18").
		WillReturnRows(buildStrategySingleSymbolHistoryRows("600519.SH", 120, 2.2, 1000, 25))
	mock.ExpectQuery(strategyStockContextDailyBasicQueryPattern).
		WithArgs("600519.SH", "2026-03-18", "600519.SH").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "turnover_rate", "volume_ratio", "pe_ttm", "pb", "total_mv", "circ_mv", "source_key"}).
			AddRow("600519.SH", selectedTradeDate, 0.75, 1.0, 24.8, 8.8, 1000000, 800000, "TUSHARE"))
	mock.ExpectQuery(strategyStockContextMoneyflowQueryPattern).
		WithArgs("600519.SH", "2026-03-18", "600519.SH").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "net_mf_amount", "buy_lg_amount", "sell_lg_amount", "buy_elg_amount", "sell_elg_amount", "source_key"}).
			AddRow("600519.SH", selectedTradeDate, 8600.2, 0, 0, 0, 0, "TUSHARE"))
	mock.ExpectQuery(strategyStockContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}))

	ctx, err := repo.BuildStrategyEngineStockSelectionContext(model.StrategyEngineStockSelectionContextRequest{
		TradeDate:   "2026-03-19",
		SeedSymbols: []string{"600519.SH"},
		Limit:       1,
	})
	if err != nil {
		t.Fatalf("BuildStrategyEngineStockSelectionContext returned error: %v", err)
	}
	if len(ctx.Seeds) != 1 {
		t.Fatalf("expected 1 seed, got %d", len(ctx.Seeds))
	}
	if ctx.Seeds[0].NewsHeat != 0 || ctx.Seeds[0].PositiveNewsRate != 0.5 {
		t.Fatalf("expected neutral news fallback, got %+v", ctx.Seeds[0])
	}
	if !strings.Contains(strings.Join(ctx.Meta.Warnings, " | "), "fallback to neutral defaults") {
		t.Fatalf("expected news fallback warning, got %#v", ctx.Meta.Warnings)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineStockSelectionContextAutoModeAppliesUniverseFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	selectedTradeDate := time.Date(2026, 3, 19, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(strategyStockContextTradeDateQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-19").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(selectedTradeDate))
	mock.ExpectQuery(strategyStockContextCandidatesQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-19", 180).
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "display_name", "selected_source_key", "volume", "turnover", "metadata_json"}).
			AddRow("600519.SH", "贵州茅台", "TUSHARE", 1000000, 980000000, "").
			AddRow("600001.SH", "*ST测试", "TUSHARE", 900000, 760000000, `{"risk_warning":true}`).
			AddRow("300750.SZ", "宁德时代", "TUSHARE", 0, 0, ""))
	mock.ExpectQuery(strategyStockContextHistoryQueryPattern).
		WithArgs(marketAssetClassStock, "600519.SH", "600001.SH", "300750.SZ", "2026-01-18", "2026-03-19").
		WillReturnRows(func() *sqlmock.Rows {
			rows := sqlmock.NewRows([]string{"instrument_key", "trade_date", "open_price", "high_price", "low_price", "close_price", "prev_close_price", "volume", "turnover"})
			rows = appendStrategyHistoryRows(rows, "600519.SH", 120, 1.8, 1000, 26)
			rows = appendStrategyHistoryRows(rows, "600001.SH", 15, 0.2, 900, 26)
			rows = appendStrategyHistoryRows(rows, "300750.SZ", 220, 2.2, 800, 26)
			return rows
		}())
	mock.ExpectQuery(strategyStockContextListingDateQueryPattern).
		WithArgs(marketAssetClassStock, "600519.SH", "600001.SH", "300750.SZ").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "min_trade_date"}).
			AddRow("600519.SH", time.Date(2025, 1, 1, 0, 0, 0, 0, time.Local)).
			AddRow("600001.SH", time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)).
			AddRow("300750.SZ", time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)))
	mock.ExpectQuery(strategyStockContextCoverageStartQueryPattern).
		WithArgs(marketAssetClassStock).
		WillReturnRows(sqlmock.NewRows([]string{"min_trade_date"}).AddRow(time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)))
	mock.ExpectQuery(strategyStockContextDailyBasicQueryPattern).
		WithArgs("600519.SH", "600001.SH", "300750.SZ", "2026-03-19", "600519.SH", "600001.SH", "300750.SZ").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "turnover_rate", "volume_ratio", "pe_ttm", "pb", "total_mv", "circ_mv", "source_key"}).
			AddRow("600519.SH", selectedTradeDate, 0.82, 1.18, 25.6, 9.1, 1000000, 800000, "TUSHARE"))
	mock.ExpectQuery(strategyStockContextMoneyflowQueryPattern).
		WithArgs("600519.SH", "600001.SH", "300750.SZ", "2026-03-19", "600519.SH", "600001.SH", "300750.SZ").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "net_mf_amount", "buy_lg_amount", "sell_lg_amount", "buy_elg_amount", "sell_elg_amount", "source_key"}).
			AddRow("600519.SH", selectedTradeDate, 9800.5, 0, 0, 0, 0, "TUSHARE"))
	mock.ExpectQuery(strategyStockContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}))

	ctx, err := repo.BuildStrategyEngineStockSelectionContext(model.StrategyEngineStockSelectionContextRequest{
		TradeDate:      "2026-03-19",
		Limit:          2,
		MinAvgTurnover: 100000,
	})
	if err != nil {
		t.Fatalf("BuildStrategyEngineStockSelectionContext returned error: %v", err)
	}
	if len(ctx.Seeds) != 1 {
		t.Fatalf("expected only one filtered seed, got %d", len(ctx.Seeds))
	}
	if ctx.Seeds[0].Symbol != "600519.SH" {
		t.Fatalf("expected valid symbol 600519.SH, got %+v", ctx.Seeds[0])
	}
	joinedWarnings := strings.Join(ctx.Meta.Warnings, " | ")
	if !strings.Contains(joinedWarnings, "suspended") || !strings.Contains(joinedWarnings, "ST symbols") {
		t.Fatalf("expected auto filter warnings, got %#v", ctx.Meta.Warnings)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineStockSelectionContextAutoModeSkipsListingDaysWhenTruthCoverageTooShort(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	selectedTradeDate := time.Date(2026, 3, 20, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(strategyStockContextTradeDateQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-21").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(selectedTradeDate))
	mock.ExpectQuery(strategyStockContextCandidatesQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-20", 180).
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "display_name", "selected_source_key", "volume", "turnover", "metadata_json"}).
			AddRow("600519", "贵州茅台", "AKSHARE", 26132, 3782818260, "").
			AddRow("300750", "宁德时代", "AKSHARE", 537297, 22252017693, ""))
	mock.ExpectQuery(strategyStockContextHistoryQueryPattern).
		WithArgs(marketAssetClassStock, "600519", "300750", "2026-01-19", "2026-03-20").
		WillReturnRows(func() *sqlmock.Rows {
			rows := sqlmock.NewRows([]string{"instrument_key", "trade_date", "open_price", "high_price", "low_price", "close_price", "prev_close_price", "volume", "turnover"})
			rows = appendStrategyHistoryRows(rows, "600519", 1400, 2.5, 200000, 30)
			rows = appendStrategyHistoryRows(rows, "300750", 380, 3.2, 500000, 30)
			return rows
		}())
	mock.ExpectQuery(strategyStockContextListingDateQueryPattern).
		WithArgs(marketAssetClassStock, "600519", "300750").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "min_trade_date"}).
			AddRow("600519", time.Date(2025, 11, 11, 0, 0, 0, 0, time.Local)).
			AddRow("300750", time.Date(2025, 11, 11, 0, 0, 0, 0, time.Local)))
	mock.ExpectQuery(strategyStockContextCoverageStartQueryPattern).
		WithArgs(marketAssetClassStock).
		WillReturnRows(sqlmock.NewRows([]string{"min_trade_date"}).AddRow(time.Date(2025, 11, 11, 0, 0, 0, 0, time.Local)))
	mock.ExpectQuery(strategyStockContextDailyBasicQueryPattern).
		WithArgs("600519", "300750", "2026-03-20", "600519", "300750").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "turnover_rate", "volume_ratio", "pe_ttm", "pb", "total_mv", "circ_mv", "source_key"}).
			AddRow("600519", selectedTradeDate, 0.82, 1.18, 25.6, 9.1, 1000000, 800000, "AKSHARE").
			AddRow("300750", selectedTradeDate, 1.36, 1.42, 32.4, 7.4, 1200000, 900000, "AKSHARE"))
	mock.ExpectQuery(strategyStockContextMoneyflowQueryPattern).
		WithArgs("600519", "300750", "2026-03-20", "600519", "300750").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "net_mf_amount", "buy_lg_amount", "sell_lg_amount", "buy_elg_amount", "sell_elg_amount", "source_key"}).
			AddRow("600519", selectedTradeDate, 9800.5, 0, 0, 0, 0, "AKSHARE").
			AddRow("300750", selectedTradeDate, 15200.7, 0, 0, 0, 0, "AKSHARE"))
	mock.ExpectQuery(strategyStockContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}))

	ctx, err := repo.BuildStrategyEngineStockSelectionContext(model.StrategyEngineStockSelectionContextRequest{
		TradeDate:      "2026-03-21",
		SelectionMode:  model.StrategyEngineStockSelectionModeAuto,
		Limit:          2,
		MinListingDays: 180,
	})
	if err != nil {
		t.Fatalf("BuildStrategyEngineStockSelectionContext returned error: %v", err)
	}
	if len(ctx.Seeds) != 2 {
		t.Fatalf("expected coverage guard to keep both seeds, got %d", len(ctx.Seeds))
	}
	if ctx.Meta.ListingDaysFilterApplied {
		t.Fatalf("expected listing days proxy filter to be disabled when truth coverage is too short")
	}
	if !strings.Contains(strings.Join(ctx.Meta.Warnings, " | "), "skipped min_listing_days proxy") {
		t.Fatalf("expected coverage guard warning, got %#v", ctx.Meta.Warnings)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineStockSelectionContextErrorsWhenTruthIsMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(strategyStockContextTradeDateQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-19", "600519.SH").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(nil))

	_, err = repo.BuildStrategyEngineStockSelectionContext(model.StrategyEngineStockSelectionContextRequest{
		TradeDate:   "2026-03-19",
		SeedSymbols: []string{"600519.SH"},
	})
	if err == nil {
		t.Fatal("expected missing truth error, got nil")
	}
	if !strings.Contains(err.Error(), "no stock truth bars available") {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineFuturesStrategyContextUsesTruthBarsAndNews(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	selectedTradeDate := time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(strategyFuturesContextTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "MOCK", "IF2606", "AU2606").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(selectedTradeDate))
	mock.ExpectQuery(strategyFuturesContextCandidatesQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-18", "MOCK", "IF2606", "AU2606").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "contract_key", "display_name", "selected_source_key"}).
			AddRow("IF2606.CFX", "IF2606", "沪深300股指", "TUSHARE").
			AddRow("AU2606.SHF", "AU2606", "沪金主力", "TUSHARE"))
	mock.ExpectQuery(strategyFuturesContextHistoryQueryPattern).
		WithArgs(marketAssetClassFutures, "IF2606.CFX", "AU2606.SHF", "2026-02-01", "2026-03-18").
		WillReturnRows(buildStrategyFuturesContextHistoryRows())
	mock.ExpectQuery(strategyFuturesContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}).
			AddRow(sql.NullString{String: "IF2606", Valid: true}, sql.NullString{String: "", Valid: false}, "股指期货突破创新高").
			AddRow(sql.NullString{String: "AU2606.SHF", Valid: true}, sql.NullString{String: "", Valid: false}, "黄金波动风险提示"))
	mock.ExpectQuery(strategyFuturesContextTermStructureQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-18", "MOCK", "IF%", "AU%").
		WillReturnRows(buildStrategyFuturesTermStructureRows(
			"IF2606.CFX", 4012.8,
			"IF2609.CFX", 4036.2,
			"AU2606.SHF", 536.7,
			"AU2608.SHF", 539.1,
		))
	mock.ExpectQuery(strategyFuturesContextInventoryQueryPattern).
		WithArgs("2026-03-18", "IF", "AU").
		WillReturnRows(buildStrategyFuturesInventoryRows(
			"IF", time.Date(2026, 3, 17, 0, 0, 0, 0, time.Local), "中储1号", "华东", "央企品牌A", "上海", "标准品", 900, 920, -20,
			"IF", time.Date(2026, 3, 17, 0, 0, 0, 0, time.Local), "中储2号", "华东", "央企品牌B", "无锡", "标准品", 600, 600, 0,
			"IF", time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local), "中储1号", "华东", "央企品牌A", "上海", "标准品", 980, 1020, -40,
			"IF", time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local), "中储2号", "华东", "央企品牌B", "无锡", "交割品", 480, 480, 0,
			"AU", time.Date(2026, 3, 17, 0, 0, 0, 0, time.Local), "国储A", "华南", "金品牌A", "深圳", "一等", 1100, 1080, 20,
			"AU", time.Date(2026, 3, 17, 0, 0, 0, 0, time.Local), "国储B", "华中", "金品牌B", "武汉", "二等", 1200, 1200, 0,
			"AU", time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local), "国储A", "华南", "金品牌A", "深圳", "一等", 1360, 1300, 60,
			"AU", time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local), "国储B", "华中", "金品牌B", "武汉", "二等", 1000, 1000, 0,
		))
	mock.ExpectQuery(strategyFuturesContextSpreadQueryPattern).
		WithArgs("IF2606", "AU2606", "IF2606", "AU2606").
		WillReturnRows(sqlmock.NewRows([]string{"contract_a", "contract_b", "percentile", "status"}).
			AddRow("IF2606", "IF2609", 0.82, "WATCH").
			AddRow("AU2606", "AU2608", 0.31, "WATCH"))

	ctx, err := repo.BuildStrategyEngineFuturesStrategyContext(model.StrategyEngineFuturesStrategyContextRequest{
		TradeDate: "2026-03-19",
		Contracts: []string{"IF2606", "AU2606"},
		Limit:     2,
	})
	if err != nil {
		t.Fatalf("BuildStrategyEngineFuturesStrategyContext returned error: %v", err)
	}
	if ctx.Meta.SelectedTradeDate != "2026-03-18" {
		t.Fatalf("expected selected trade date 2026-03-18, got %s", ctx.Meta.SelectedTradeDate)
	}
	if ctx.Meta.PriceSource != "TUSHARE" {
		t.Fatalf("expected TUSHARE price source, got %s", ctx.Meta.PriceSource)
	}
	if ctx.Meta.NewsWindowDays != 14 {
		t.Fatalf("expected 14-day news window, got %d", ctx.Meta.NewsWindowDays)
	}
	if len(ctx.Meta.Warnings) != 0 {
		t.Fatalf("expected no warnings, got %#v", ctx.Meta.Warnings)
	}
	if len(ctx.Seeds) != 2 {
		t.Fatalf("expected 2 seeds, got %d", len(ctx.Seeds))
	}
	if ctx.Seeds[0].Contract != "IF2606" || ctx.Seeds[0].NewsBias <= 0 {
		t.Fatalf("unexpected first futures seed: %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[1].Contract != "AU2606" || ctx.Seeds[1].NewsBias >= 0 {
		t.Fatalf("unexpected second futures seed: %+v", ctx.Seeds[1])
	}
	if ctx.Seeds[0].InventoryFocusArea != "华东" || ctx.Seeds[0].InventoryFocusWarehouse != "中储1号" || ctx.Seeds[0].InventoryFocusBrand != "央企品牌A" {
		t.Fatalf("expected IF inventory focus dimensions, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].InventoryFocusPlace != "上海" || ctx.Seeds[0].InventoryFocusGrade != "标准品" {
		t.Fatalf("expected IF place/grade focus dimensions, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].InventoryWarehouseShare <= 0.6 || ctx.Seeds[0].InventoryBrandShare <= 0.6 {
		t.Fatalf("expected IF inventory concentration shares, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].InventoryPlaceShare <= 0.6 || ctx.Seeds[0].InventoryGradeShare <= 0.6 {
		t.Fatalf("expected IF place/grade concentration shares, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].TermStructurePct <= 0 || ctx.Seeds[0].TurnoverRatio <= 0 {
		t.Fatalf("expected term structure and turnover ratios, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].CurveSlopePct <= 0 || ctx.Seeds[0].SpreadPressure >= 0 || ctx.Seeds[0].SpreadPair == "" {
		t.Fatalf("expected curve slope and spread pressure enrichment, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].InventoryPressure <= 0 || ctx.Seeds[1].InventoryPressure >= 0 {
		t.Fatalf("expected inventory pressure enrichment, got %+v / %+v", ctx.Seeds[0], ctx.Seeds[1])
	}
	if ctx.Seeds[1].SpreadPressure <= 0 {
		t.Fatalf("expected AU seed to receive positive spread pressure, got %+v", ctx.Seeds[1])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineFuturesStrategyContextFallsBackToNeutralNewsSignals(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	selectedTradeDate := time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(strategyFuturesContextTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "MOCK", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(selectedTradeDate))
	mock.ExpectQuery(strategyFuturesContextCandidatesQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-18", "MOCK", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "contract_key", "display_name", "selected_source_key"}).
			AddRow("IF2606.CFX", "IF2606", "沪深300股指", "TUSHARE"))
	mock.ExpectQuery(strategyFuturesContextHistoryQueryPattern).
		WithArgs(marketAssetClassFutures, "IF2606.CFX", "2026-02-01", "2026-03-18").
		WillReturnRows(buildStrategySingleFuturesHistoryRows("IF2606.CFX", 3980, 6, 200000, 150000))
	mock.ExpectQuery(strategyFuturesContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}))
	mock.ExpectQuery(strategyFuturesContextTermStructureQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-18", "MOCK", "IF%").
		WillReturnRows(buildStrategyFuturesTermStructureRows(
			"IF2606.CFX", 4012.8,
			"IF2609.CFX", 4036.2,
		))
	mock.ExpectQuery(strategyFuturesContextInventoryQueryPattern).
		WithArgs("2026-03-18", "IF").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "total_receipt_volume", "total_previous_volume", "total_change_volume"}))
	mock.ExpectQuery(strategyFuturesContextSpreadQueryPattern).
		WithArgs("IF2606", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"contract_a", "contract_b", "percentile", "status"}).
			AddRow("IF2606", "IF2609", 0.86, "WATCH"))

	ctx, err := repo.BuildStrategyEngineFuturesStrategyContext(model.StrategyEngineFuturesStrategyContextRequest{
		TradeDate: "2026-03-19",
		Contracts: []string{"IF2606"},
		Limit:     1,
	})
	if err != nil {
		t.Fatalf("BuildStrategyEngineFuturesStrategyContext returned error: %v", err)
	}
	if len(ctx.Seeds) != 1 {
		t.Fatalf("expected 1 seed, got %d", len(ctx.Seeds))
	}
	if ctx.Seeds[0].NewsBias != 0 {
		t.Fatalf("expected neutral futures news fallback, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].CurveSlopePct <= 0 {
		t.Fatalf("expected curve slope metric, got %+v", ctx.Seeds[0])
	}
	if !strings.Contains(strings.Join(ctx.Meta.Warnings, " | "), "fallback to neutral defaults") {
		t.Fatalf("expected news fallback warning, got %#v", ctx.Meta.Warnings)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineFuturesStrategyContextErrorsWhenTruthIsMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(strategyFuturesContextTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "MOCK", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(nil))
	mock.ExpectQuery(strategyFuturesContextTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(nil))

	_, err = repo.BuildStrategyEngineFuturesStrategyContext(model.StrategyEngineFuturesStrategyContextRequest{
		TradeDate: "2026-03-19",
		Contracts: []string{"IF2606"},
	})
	if err == nil {
		t.Fatal("expected missing futures truth error, got nil")
	}
	if !strings.Contains(err.Error(), "no futures truth bars available") {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineFuturesStrategyContextPrefersRealTruthBeforeMock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	realTradeDate := time.Date(2024, 6, 18, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(strategyFuturesContextTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "MOCK").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(realTradeDate))
	mock.ExpectQuery(strategyFuturesContextCandidatesQueryPattern).
		WithArgs(marketAssetClassFutures, "2024-06-18", "MOCK", 6).
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "contract_key", "display_name", "selected_source_key"}).
			AddRow("IF2406.CFX", "IF2406", "股指主力", "AKSHARE").
			AddRow("AU2406.SHF", "AU2406", "沪金主力", "AKSHARE"))
	mock.ExpectQuery(strategyFuturesContextHistoryQueryPattern).
		WithArgs(marketAssetClassFutures, "IF2406.CFX", "AU2406.SHF", "2024-05-04", "2024-06-18").
		WillReturnRows(buildStrategySpecificFuturesContextHistoryRows([]string{"IF2406.CFX", "AU2406.SHF"}))
	mock.ExpectQuery(strategyFuturesContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}))
	mock.ExpectQuery(strategyFuturesContextTermStructureQueryPattern).
		WithArgs(marketAssetClassFutures, "2024-06-18", "MOCK", "IF%", "AU%").
		WillReturnRows(buildStrategyFuturesTermStructureRows(
			"IF2406.CFX", 3988.0,
			"IF2409.CFX", 4018.0,
			"AU2406.SHF", 530.6,
			"AU2408.SHF", 533.1,
		))
	mock.ExpectQuery(strategyFuturesContextInventoryQueryPattern).
		WithArgs("2024-06-18", "IF", "AU").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "total_receipt_volume", "total_previous_volume", "total_change_volume"}))
	mock.ExpectQuery(strategyFuturesContextSpreadQueryPattern).
		WithArgs("IF2406", "AU2406", "IF2406", "AU2406").
		WillReturnRows(sqlmock.NewRows([]string{"contract_a", "contract_b", "percentile", "status"}))

	ctx, err := repo.BuildStrategyEngineFuturesStrategyContext(model.StrategyEngineFuturesStrategyContextRequest{
		TradeDate: "2026-03-19",
		Limit:     2,
	})
	if err != nil {
		t.Fatalf("BuildStrategyEngineFuturesStrategyContext returned error: %v", err)
	}
	if ctx.Meta.SelectedTradeDate != "2024-06-18" {
		t.Fatalf("expected real selected trade date, got %s", ctx.Meta.SelectedTradeDate)
	}
	if ctx.Meta.PriceSource != "AKSHARE" {
		t.Fatalf("expected AKSHARE price source, got %s", ctx.Meta.PriceSource)
	}
	for _, warning := range ctx.Meta.Warnings {
		if strings.Contains(warning, "fallback to MOCK truth source") {
			t.Fatalf("expected no mock fallback warning, got %#v", ctx.Meta.Warnings)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineFuturesStrategyContextExplicitContractsFallBackToMockWithWarning(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mockTradeDate := time.Date(2026, 3, 19, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(strategyFuturesContextTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "MOCK", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(nil))
	mock.ExpectQuery(strategyFuturesContextTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(mockTradeDate))
	mock.ExpectQuery(strategyFuturesContextCandidatesQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "contract_key", "display_name", "selected_source_key"}).
			AddRow("IF2606.CFX", "IF2606", "沪深300股指", "MOCK"))
	mock.ExpectQuery(strategyFuturesContextHistoryQueryPattern).
		WithArgs(marketAssetClassFutures, "IF2606.CFX", "2026-02-02", "2026-03-19").
		WillReturnRows(buildStrategySingleFuturesHistoryRows("IF2606.CFX", 3980, 6, 200000, 150000))
	mock.ExpectQuery(strategyFuturesContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}))
	mock.ExpectQuery(strategyFuturesContextTermStructureQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "IF%").
		WillReturnRows(buildStrategyFuturesTermStructureRows(
			"IF2606.CFX", 4012.8,
			"IF2609.CFX", 4036.2,
		))
	mock.ExpectQuery(strategyFuturesContextInventoryQueryPattern).
		WithArgs("2026-03-19", "IF").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "total_receipt_volume", "total_previous_volume", "total_change_volume"}))
	mock.ExpectQuery(strategyFuturesContextSpreadQueryPattern).
		WithArgs("IF2606", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"contract_a", "contract_b", "percentile", "status"}))

	ctx, err := repo.BuildStrategyEngineFuturesStrategyContext(model.StrategyEngineFuturesStrategyContextRequest{
		TradeDate: "2026-03-19",
		Contracts: []string{"IF2606"},
		Limit:     1,
	})
	if err != nil {
		t.Fatalf("BuildStrategyEngineFuturesStrategyContext returned error: %v", err)
	}
	if ctx.Meta.PriceSource != "MOCK" {
		t.Fatalf("expected MOCK price source after fallback, got %s", ctx.Meta.PriceSource)
	}
	if !strings.Contains(strings.Join(ctx.Meta.Warnings, " | "), "fallback to MOCK truth source") {
		t.Fatalf("expected mock fallback warning, got %#v", ctx.Meta.Warnings)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineFuturesStrategyContextAllowsMockFallbackOnShortHistory(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	realTradeDate := time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local)
	mockTradeDate := time.Date(2026, 3, 19, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(strategyFuturesContextTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "MOCK", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(realTradeDate))
	mock.ExpectQuery(strategyFuturesContextCandidatesQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-18", "MOCK", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "contract_key", "display_name", "selected_source_key"}).
			AddRow("IF2606.CFX", "IF2606", "沪深300股指", "TUSHARE"))
	mock.ExpectQuery(strategyFuturesContextHistoryQueryPattern).
		WithArgs(marketAssetClassFutures, "IF2606.CFX", "2026-02-01", "2026-03-18").
		WillReturnRows(buildStrategyShortFuturesHistoryRows("IF2606.CFX", 8))
	mock.ExpectQuery(strategyFuturesContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}))
	mock.ExpectQuery(strategyFuturesContextTermStructureQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-18", "MOCK", "IF%").
		WillReturnRows(buildStrategyFuturesTermStructureRows(
			"IF2606.CFX", 3998.0,
			"IF2609.CFX", 4016.0,
		))
	mock.ExpectQuery(strategyFuturesContextInventoryQueryPattern).
		WithArgs("2026-03-18", "IF").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "total_receipt_volume", "total_previous_volume", "total_change_volume"}))
	mock.ExpectQuery(strategyFuturesContextSpreadQueryPattern).
		WithArgs("IF2606", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"contract_a", "contract_b", "percentile", "status"}))
	mock.ExpectQuery(strategyFuturesContextTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "MOCK", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(mockTradeDate))
	mock.ExpectQuery(strategyFuturesContextCandidatesQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "contract_key", "display_name", "selected_source_key"}).
			AddRow("IF2606.CFX", "IF2606", "沪深300股指", "MOCK"))
	mock.ExpectQuery(strategyFuturesContextHistoryQueryPattern).
		WithArgs(marketAssetClassFutures, "IF2606.CFX", "2026-02-02", "2026-03-19").
		WillReturnRows(buildStrategySingleFuturesHistoryRows("IF2606.CFX", 3980, 6, 200000, 150000))
	mock.ExpectQuery(strategyFuturesContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}))
	mock.ExpectQuery(strategyFuturesContextTermStructureQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "IF%").
		WillReturnRows(buildStrategyFuturesTermStructureRows(
			"IF2606.CFX", 4012.8,
			"IF2609.CFX", 4036.2,
		))
	mock.ExpectQuery(strategyFuturesContextInventoryQueryPattern).
		WithArgs("2026-03-19", "IF").
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "trade_date", "total_receipt_volume", "total_previous_volume", "total_change_volume"}))
	mock.ExpectQuery(strategyFuturesContextSpreadQueryPattern).
		WithArgs("IF2606", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"contract_a", "contract_b", "percentile", "status"}))

	ctx, err := repo.BuildStrategyEngineFuturesStrategyContext(model.StrategyEngineFuturesStrategyContextRequest{
		TradeDate:                       "2026-03-19",
		Contracts:                       []string{"IF2606"},
		Limit:                           1,
		AllowMockFallbackOnShortHistory: true,
	})
	if err != nil {
		t.Fatalf("BuildStrategyEngineFuturesStrategyContext returned error: %v", err)
	}
	if ctx.Meta.SelectedTradeDate != "2026-03-19" {
		t.Fatalf("expected mock fallback trade date 2026-03-19, got %s", ctx.Meta.SelectedTradeDate)
	}
	if ctx.Meta.PriceSource != "MOCK" {
		t.Fatalf("expected MOCK price source after short-history fallback, got %s", ctx.Meta.PriceSource)
	}
	if !strings.Contains(strings.Join(ctx.Meta.Warnings, " | "), "after explicit opt-in") {
		t.Fatalf("expected explicit opt-in fallback warning, got %#v", ctx.Meta.Warnings)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestBuildStrategyEngineFuturesStrategyContextEnrichesFactorsFromRawBars(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	selectedTradeDate := time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(strategyFuturesContextTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-19", "MOCK", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(selectedTradeDate))
	mock.ExpectQuery(strategyFuturesContextCandidatesQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-18", "MOCK", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "contract_key", "display_name", "selected_source_key"}).
			AddRow("IF2606.CFX", "IF2606", "沪深300股指", "AKSHARE"))
	mock.ExpectQuery(strategyFuturesContextHistoryQueryPattern).
		WithArgs(marketAssetClassFutures, "IF2606.CFX", "2026-02-01", "2026-03-18").
		WillReturnRows(buildStrategySingleFuturesHistoryRowsWithTurnover("IF2606.CFX", 3980, 6, 200000, 0, 150000, 0, true))
	mock.ExpectQuery(strategyFuturesContextSupplementQueryPattern).
		WithArgs(marketAssetClassFutures, "IF2606.CFX", "2026-02-01", "2026-03-18").
		WillReturnRows(buildStrategySingleFuturesSupplementRows("IF2606.CFX", 3980, 6, 200000, 150000))
	mock.ExpectQuery(strategyFuturesContextNewsQueryPattern).
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"primary_symbol", "symbols_json", "title"}))
	mock.ExpectQuery(strategyFuturesContextTermStructureQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-18", "MOCK", "IF%").
		WillReturnRows(buildStrategyFuturesTermStructureRows(
			"IF2606.CFX", 4012.8,
			"IF2609.CFX", 4036.2,
		))
	mock.ExpectQuery(strategyFuturesContextInventoryQueryPattern).
		WithArgs("2026-03-18", "IF").
		WillReturnRows(buildStrategyFuturesInventoryRows(
			"IF", time.Date(2026, 3, 17, 0, 0, 0, 0, time.Local), "中储1号", "华东", "央企品牌A", "上海", "标准品", 900, 920, -20,
			"IF", time.Date(2026, 3, 17, 0, 0, 0, 0, time.Local), "中储2号", "华东", "央企品牌B", "无锡", "标准品", 600, 600, 0,
			"IF", time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local), "中储1号", "华东", "央企品牌A", "上海", "标准品", 980, 1020, -40,
			"IF", time.Date(2026, 3, 18, 0, 0, 0, 0, time.Local), "中储2号", "华东", "央企品牌B", "无锡", "交割品", 480, 480, 0,
		))
	mock.ExpectQuery(strategyFuturesContextSpreadQueryPattern).
		WithArgs("IF2606", "IF2606").
		WillReturnRows(sqlmock.NewRows([]string{"contract_a", "contract_b", "percentile", "status"}).
			AddRow("IF2606", "IF2609", 0.86, "WATCH"))

	ctx, err := repo.BuildStrategyEngineFuturesStrategyContext(model.StrategyEngineFuturesStrategyContextRequest{
		TradeDate: "2026-03-19",
		Contracts: []string{"IF2606"},
		Limit:     1,
	})
	if err != nil {
		t.Fatalf("BuildStrategyEngineFuturesStrategyContext returned error: %v", err)
	}
	if len(ctx.Seeds) != 1 {
		t.Fatalf("expected 1 seed, got %d", len(ctx.Seeds))
	}
	if ctx.Seeds[0].BasisPct == 0 {
		t.Fatalf("expected enriched non-zero basis pct, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].CarryPct == 0 {
		t.Fatalf("expected enriched non-zero carry pct, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].FlowBias == 0 {
		t.Fatalf("expected enriched non-zero flow bias, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].TermStructurePct <= 0 {
		t.Fatalf("expected non-zero term structure pct, got %+v", ctx.Seeds[0])
	}
	if ctx.Seeds[0].CurveSlopePct <= 0 || ctx.Seeds[0].SpreadPressure >= 0 || ctx.Seeds[0].InventoryPressure <= 0 {
		t.Fatalf("expected curve slope, positive inventory pressure and negative spread pressure, got %+v", ctx.Seeds[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func buildStrategyStockContextHistoryRows() *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"instrument_key", "trade_date", "open_price", "high_price", "low_price", "close_price", "prev_close_price", "volume", "turnover"})
	rows = appendStrategyHistoryRows(rows, "600519.SH", 120, 1.6, 1000, 26)
	rows = appendStrategyHistoryRows(rows, "300750.SZ", 220, 2.2, 1800, 26)
	return rows
}

func buildStrategySingleSymbolHistoryRows(symbol string, startClose float64, closeStep float64, startVolume float64, count int) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"instrument_key", "trade_date", "open_price", "high_price", "low_price", "close_price", "prev_close_price", "volume", "turnover"})
	return appendStrategyHistoryRows(rows, symbol, startClose, closeStep, startVolume, count)
}

func appendStrategyHistoryRows(rows *sqlmock.Rows, symbol string, startClose float64, closeStep float64, startVolume float64, count int) *sqlmock.Rows {
	tradeDate := time.Date(2026, 2, 15, 0, 0, 0, 0, time.Local)
	prevClose := startClose - closeStep
	for index := 0; index < count; index++ {
		closePrice := startClose + float64(index)*closeStep
		openPrice := closePrice - 0.6
		highPrice := closePrice + 1.1
		lowPrice := closePrice - 1.2
		volume := startVolume + float64(index)*18
		turnover := closePrice * volume
		rows.AddRow(symbol, tradeDate.AddDate(0, 0, index), openPrice, highPrice, lowPrice, closePrice, prevClose, volume, turnover)
		prevClose = closePrice
	}
	return rows
}

func buildStrategyFuturesContextHistoryRows() *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"instrument_key", "trade_date", "close_price", "prev_close_price", "settle_price", "prev_settle_price", "volume", "turnover", "open_interest"})
	rows = appendStrategyFuturesHistoryRows(rows, "IF2606.CFX", 3920, 6, 180000, 140000, 0, 0, 26, false)
	rows = appendStrategyFuturesHistoryRows(rows, "AU2606.SHF", 532, 1.8, 80000, 60000, 0, 0, 26, false)
	return rows
}

func buildStrategySpecificFuturesContextHistoryRows(instrumentKeys []string) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"instrument_key", "trade_date", "close_price", "prev_close_price", "settle_price", "prev_settle_price", "volume", "turnover", "open_interest"})
	for _, instrumentKey := range instrumentKeys {
		startClose := 3000.0
		closeStep := 5.0
		startVolume := 90000.0
		startOpenInterest := 70000.0
		if strings.Contains(instrumentKey, "AU") {
			startClose = 540
			closeStep = 1.6
			startVolume = 75000
			startOpenInterest = 65000
		}
		rows = appendStrategyFuturesHistoryRows(rows, instrumentKey, startClose, closeStep, startVolume, startOpenInterest, 0, 0, 26, false)
	}
	return rows
}

func buildStrategySingleFuturesHistoryRows(instrumentKey string, startClose float64, closeStep float64, startVolume float64, startOpenInterest float64) *sqlmock.Rows {
	return buildStrategySingleFuturesHistoryRowsWithTurnover(instrumentKey, startClose, closeStep, startVolume, 0, startOpenInterest, 0, false)
}

func buildStrategySingleFuturesHistoryRowsWithTurnover(instrumentKey string, startClose float64, closeStep float64, startVolume float64, startTurnover float64, startOpenInterest float64, turnoverStep float64, flatSettle bool) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"instrument_key", "trade_date", "close_price", "prev_close_price", "settle_price", "prev_settle_price", "volume", "turnover", "open_interest"})
	return appendStrategyFuturesHistoryRows(rows, instrumentKey, startClose, closeStep, startVolume, startOpenInterest, startTurnover, turnoverStep, 26, flatSettle)
}

func buildStrategyShortFuturesHistoryRows(instrumentKey string, count int) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"instrument_key", "trade_date", "close_price", "prev_close_price", "settle_price", "prev_settle_price", "volume", "turnover", "open_interest"})
	return appendStrategyFuturesHistoryRows(rows, instrumentKey, 3980, 6, 200000, 150000, 0, 0, count, false)
}

func buildStrategySingleFuturesSupplementRows(instrumentKey string, startClose float64, closeStep float64, startVolume float64, startOpenInterest float64) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"instrument_key", "trade_date", "source_key", "close_price", "settle_price", "prev_settle_price", "turnover", "open_interest"})
	tradeDate := time.Date(2026, 2, 15, 0, 0, 0, 0, time.Local)
	prevSettle := startClose - closeStep*1.25
	for index := 0; index < 26; index++ {
		closePrice := startClose + float64(index)*closeStep
		settlePrice := closePrice - closeStep*0.35
		turnover := closePrice * (startVolume + float64(index)*1200)
		openInterest := startOpenInterest + float64(index)*850
		rows.AddRow(instrumentKey, tradeDate.AddDate(0, 0, index), "TUSHARE", closePrice, settlePrice, prevSettle, turnover, openInterest)
		prevSettle = settlePrice
	}
	return rows
}

func buildStrategyFuturesTermStructureRows(values ...any) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"instrument_key", "close_price"})
	for index := 0; index+1 < len(values); index += 2 {
		instrumentKey, _ := values[index].(string)
		closePrice, _ := values[index+1].(float64)
		rows.AddRow(instrumentKey, closePrice)
	}
	return rows
}

func buildStrategyFuturesInventoryRows(values ...any) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{"symbol", "trade_date", "warehouse", "area", "brand", "place", "grade", "receipt_volume", "previous_volume", "change_volume"})
	for index := 0; index+9 < len(values); index += 10 {
		symbol, _ := values[index].(string)
		tradeDate, _ := values[index+1].(time.Time)
		warehouse, _ := values[index+2].(string)
		area, _ := values[index+3].(string)
		brand, _ := values[index+4].(string)
		place, _ := values[index+5].(string)
		grade, _ := values[index+6].(string)
		receiptVolume := inventoryTestFloat(values[index+7])
		previousVolume := inventoryTestFloat(values[index+8])
		changeVolume := inventoryTestFloat(values[index+9])
		rows.AddRow(symbol, tradeDate, warehouse, area, brand, place, grade, receiptVolume, previousVolume, changeVolume)
	}
	return rows
}

func inventoryTestFloat(value any) float64 {
	switch item := value.(type) {
	case float64:
		return item
	case int:
		return float64(item)
	case int64:
		return float64(item)
	default:
		return 0
	}
}

func appendStrategyFuturesHistoryRows(rows *sqlmock.Rows, instrumentKey string, startClose float64, closeStep float64, startVolume float64, startOpenInterest float64, startTurnover float64, turnoverStep float64, count int, flatSettle bool) *sqlmock.Rows {
	tradeDate := time.Date(2026, 2, 15, 0, 0, 0, 0, time.Local)
	prevClose := startClose - closeStep
	prevSettle := prevClose - closeStep*0.2
	for index := 0; index < count; index++ {
		closePrice := startClose + float64(index)*closeStep
		settlePrice := closePrice - closeStep*0.15
		if flatSettle {
			settlePrice = closePrice
		}
		volume := startVolume + float64(index)*1200
		openInterest := startOpenInterest + float64(index)*850
		turnover := startTurnover + float64(index)*turnoverStep
		if startTurnover == 0 && turnoverStep == 0 && !flatSettle {
			turnover = closePrice * volume
		}
		rows.AddRow(instrumentKey, tradeDate.AddDate(0, 0, index), closePrice, prevClose, settlePrice, prevSettle, volume, turnover, openInterest)
		prevClose = closePrice
		prevSettle = settlePrice
	}
	return rows
}
