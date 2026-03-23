package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const marketQualityLogsCountQueryPattern = `SELECT COUNT\(\*\) FROM market_data_quality_logs`
const marketQualityLogsListQueryPattern = `(?s)SELECT id, asset_class, data_kind, instrument_key, trade_date, source_key, severity, issue_code, issue_message, COALESCE\(CAST\(payload_json AS CHAR\), ''\), created_at\s+FROM market_data_quality_logs`
const marketDerivedTruthSummaryQueryPattern = `(?s)SELECT asset_class, source_key, issue_code, issue_message, COALESCE\(CAST\(payload_json AS CHAR\), ''\), created_at\s+FROM market_data_quality_logs`
const marketQualitySummaryCountQueryPattern = `(?s)SELECT COUNT\(\*\),\s*COALESCE\(SUM\(CASE WHEN severity = 'ERROR' THEN 1 ELSE 0 END\), 0\),\s*COALESCE\(SUM\(CASE WHEN severity = 'WARN' THEN 1 ELSE 0 END\), 0\),\s*COALESCE\(SUM\(CASE WHEN severity = 'INFO' THEN 1 ELSE 0 END\), 0\),\s*COUNT\(DISTINCT CASE WHEN COALESCE\(source_key, ''\) = '' THEN NULL ELSE source_key END\)\s+FROM market_data_quality_logs`
const marketQualitySummaryLatestQueryPattern = `(?s)SELECT source_key, severity, issue_code, issue_message, trade_date, created_at\s+FROM market_data_quality_logs`
const marketQualitySummaryLatestErrorQueryPattern = `(?s)SELECT source_key, issue_code, issue_message, created_at\s+FROM market_data_quality_logs`
const marketDerivedTruthMaxTradeDateQueryPattern = `SELECT MAX\(trade_date\)\s+FROM market_daily_bar_truth`
const marketDerivedTruthTruthBarsQueryPattern = `(?s)SELECT asset_class, instrument_key, trade_date, selected_source_key, open_price, high_price, low_price, close_price, prev_close_price, settle_price, prev_settle_price, volume, turnover, open_interest\s+FROM market_daily_bar_truth`
const marketDerivedTruthStockBaseQueryPattern = `SELECT instrument_key, display_name, list_date, COALESCE\(CAST\(metadata_json AS CHAR\), ''\)\s+FROM market_instruments`
const marketDerivedTruthStockStatusUpsertPattern = `INSERT INTO stock_status_truth`
const marketDerivedTruthFuturesMappingUpsertPattern = `INSERT INTO futures_contract_mappings`
const marketQualityLogInsertPattern = `INSERT INTO market_data_quality_logs`

func TestAdminListMarketDataQualityLogsAppliesFiltersAndMapsRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	createdAt := time.Date(2026, 3, 22, 9, 30, 0, 0, time.Local)
	tradeDate := time.Date(2026, 3, 21, 0, 0, 0, 0, time.Local)

	mock.ExpectQuery(marketQualityLogsCountQueryPattern).
		WithArgs("STOCK", "DAILY_BARS", "WARN", "SOURCE_FETCH_FAILED").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(marketQualityLogsListQueryPattern).
		WithArgs("STOCK", "DAILY_BARS", "WARN", "SOURCE_FETCH_FAILED", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"asset_class",
			"data_kind",
			"instrument_key",
			"trade_date",
			"source_key",
			"severity",
			"issue_code",
			"issue_message",
			"payload_json",
			"created_at",
		}).AddRow(
			"mdq_001",
			"STOCK",
			"DAILY_BARS",
			"600519.SH",
			tradeDate,
			"TUSHARE",
			"WARN",
			"SOURCE_FETCH_FAILED",
			"upstream timeout",
			`{"attempt":1}`,
			createdAt,
		))

	items, total, err := repo.AdminListMarketDataQualityLogs("stock", "daily_bars", "warn", "source_fetch_failed", 0, 1, 20)
	if err != nil {
		t.Fatalf("AdminListMarketDataQualityLogs returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].AssetClass != "STOCK" || items[0].IssueCode != "SOURCE_FETCH_FAILED" {
		t.Fatalf("unexpected quality log item: %+v", items[0])
	}
	if items[0].TradeDate != "2026-03-21" {
		t.Fatalf("expected trade_date 2026-03-21, got %q", items[0].TradeDate)
	}
	if items[0].CreatedAt == "" {
		t.Fatalf("expected created_at to be set, got %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminListMarketDataQualityLogsAppliesLookbackHours(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(marketQualityLogsCountQueryPattern).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(marketQualityLogsListQueryPattern).
		WithArgs(sqlmock.AnyArg(), 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"asset_class",
			"data_kind",
			"instrument_key",
			"trade_date",
			"source_key",
			"severity",
			"issue_code",
			"issue_message",
			"payload_json",
			"created_at",
		}))

	_, _, err = repo.AdminListMarketDataQualityLogs("", "", "", "", 72, 1, 20)
	if err != nil {
		t.Fatalf("AdminListMarketDataQualityLogs returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminGetMarketDerivedTruthSummaryReadsLatestRebuildLog(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	createdAt := time.Date(2026, 3, 22, 9, 45, 0, 0, time.Local)

	mock.ExpectQuery(marketDerivedTruthSummaryQueryPattern).
		WithArgs("STOCK", "LOCAL_TRUTH", "DERIVED_STOCK_STATUS_REBUILT").
		WillReturnRows(sqlmock.NewRows([]string{
			"asset_class",
			"source_key",
			"issue_code",
			"issue_message",
			"payload_json",
			"created_at",
		}).AddRow(
			"STOCK",
			"LOCAL_TRUTH",
			"DERIVED_STOCK_STATUS_REBUILT",
			"rebuilt stock status truth for 12 rows",
			`{"asset_class":"STOCK","trade_date":"2026-03-22","start_date":"2026-03-20","end_date":"2026-03-22","days":3,"truth_bar_count":120,"stock_status_count":12,"warnings":["schema compat"]}`,
			createdAt,
		))

	summary, err := repo.AdminGetMarketDerivedTruthSummary("stock")
	if err != nil {
		t.Fatalf("AdminGetMarketDerivedTruthSummary returned error: %v", err)
	}
	if summary == nil {
		t.Fatalf("expected summary, got nil")
	}
	if summary.AssetClass != "STOCK" || summary.IssueCode != "DERIVED_STOCK_STATUS_REBUILT" {
		t.Fatalf("unexpected summary head: %+v", summary)
	}
	if summary.TradeDate != "2026-03-22" || summary.StartDate != "2026-03-20" || summary.EndDate != "2026-03-22" {
		t.Fatalf("unexpected summary dates: %+v", summary)
	}
	if summary.TruthBarCount != 120 || summary.StockStatusCount != 12 {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}
	if len(summary.Warnings) != 1 || summary.Warnings[0] != "schema compat" {
		t.Fatalf("unexpected summary warnings: %+v", summary)
	}
	if summary.CreatedAt == "" {
		t.Fatalf("expected created_at to be set, got %+v", summary)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminGetMarketDataQualitySummaryAggregatesCountsAndLatestEntries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	latestTradeDate := time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local)
	latestCreatedAt := time.Date(2026, 3, 22, 9, 55, 0, 0, time.Local)
	latestErrorAt := time.Date(2026, 3, 22, 9, 12, 0, 0, time.Local)

	mock.ExpectQuery(marketQualitySummaryCountQueryPattern).
		WithArgs("STOCK", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"total_count",
			"error_count",
			"warn_count",
			"info_count",
			"distinct_source_count",
		}).AddRow(9, 2, 5, 2, 3))

	mock.ExpectQuery(marketQualitySummaryLatestQueryPattern).
		WithArgs("STOCK", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"source_key",
			"severity",
			"issue_code",
			"issue_message",
			"trade_date",
			"created_at",
		}).AddRow(
			"MYSELF",
			"WARN",
			"BAR_UPSERT_RETRIED",
			"upsert retried with fallback",
			latestTradeDate,
			latestCreatedAt,
		))

	mock.ExpectQuery(marketQualitySummaryLatestErrorQueryPattern).
		WithArgs("STOCK", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"source_key",
			"issue_code",
			"issue_message",
			"created_at",
		}).AddRow(
			"TUSHARE",
			"SOURCE_FETCH_FAILED",
			"upstream timeout",
			latestErrorAt,
		))

	summary, err := repo.AdminGetMarketDataQualitySummary("stock", 24)
	if err != nil {
		t.Fatalf("AdminGetMarketDataQualitySummary returned error: %v", err)
	}
	if summary.AssetClass != "STOCK" || summary.LookbackHours != 24 {
		t.Fatalf("unexpected summary head: %+v", summary)
	}
	if summary.TotalCount != 9 || summary.ErrorCount != 2 || summary.WarnCount != 5 || summary.InfoCount != 2 {
		t.Fatalf("unexpected summary counts: %+v", summary)
	}
	if summary.DistinctSourceCount != 3 {
		t.Fatalf("unexpected source count: %+v", summary)
	}
	if summary.LatestIssueCode != "BAR_UPSERT_RETRIED" || summary.LatestSourceKey != "MYSELF" {
		t.Fatalf("unexpected latest event fields: %+v", summary)
	}
	if summary.LatestTradeDate != "2026-03-22" || summary.LatestCreatedAt == "" {
		t.Fatalf("unexpected latest event time fields: %+v", summary)
	}
	if summary.LatestErrorIssueCode != "SOURCE_FETCH_FAILED" || summary.LatestErrorSourceKey != "TUSHARE" {
		t.Fatalf("unexpected latest error fields: %+v", summary)
	}
	if summary.LatestErrorCreatedAt == "" {
		t.Fatalf("expected latest error created_at, got %+v", summary)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminGetMarketProviderGovernanceOverviewCombinesQualitySummaryAndProviderScores(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	latestTradeDate := time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local)
	latestCreatedAt := time.Date(2026, 3, 22, 10, 0, 0, 0, time.Local)
	latestErrorAt := time.Date(2026, 3, 22, 9, 15, 0, 0, time.Local)
	providerUpdatedAt := time.Date(2026, 3, 22, 9, 5, 0, 0, time.Local)
	providerLatestIssueAt := time.Date(2026, 3, 22, 8, 55, 0, 0, time.Local)
	truthCreatedAt := time.Date(2026, 3, 22, 10, 5, 0, 0, time.Local)

	mock.ExpectQuery(marketQualitySummaryCountQueryPattern).
		WithArgs("STOCK", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"total_count",
			"error_count",
			"warn_count",
			"info_count",
			"distinct_source_count",
		}).AddRow(4, 1, 2, 1, 2))

	mock.ExpectQuery(marketQualitySummaryLatestQueryPattern).
		WithArgs("STOCK", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"source_key",
			"severity",
			"issue_code",
			"issue_message",
			"trade_date",
			"created_at",
		}).AddRow(
			"TUSHARE",
			"WARN",
			"BAR_UPSERT_RETRIED",
			"upsert retried with fallback",
			latestTradeDate,
			latestCreatedAt,
		))

	mock.ExpectQuery(marketQualitySummaryLatestErrorQueryPattern).
		WithArgs("STOCK", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"source_key",
			"issue_code",
			"issue_message",
			"created_at",
		}).AddRow(
			"TUSHARE",
			"SOURCE_FETCH_FAILED",
			"upstream timeout",
			latestErrorAt,
		))

	mock.ExpectQuery(marketProviderCapabilityListQueryPattern).
		WithArgs("STOCK", "DAILY_BARS").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider_key",
			"asset_class",
			"data_kind",
			"supports_sync",
			"supports_truth_rebuild",
			"supports_context_seed",
			"supports_research_run",
			"supports_backfill",
			"supports_batch",
			"supports_intraday",
			"supports_history",
			"supports_metadata_enrichment",
			"requires_auth",
			"fallback_allowed",
			"priority_weight",
			"updated_at",
		}).AddRow(
			"TUSHARE",
			"STOCK",
			"DAILY_BARS",
			true,
			true,
			true,
			true,
			true,
			true,
			false,
			true,
			false,
			true,
			true,
			100,
			providerUpdatedAt,
		))

	mock.ExpectQuery(marketProviderQualityAggregateQueryPattern).
		WithArgs("STOCK", "DAILY_BARS", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"source_key",
			"total_count",
			"error_count",
			"warn_count",
			"latest_created_at",
		}).AddRow(
			"TUSHARE",
			2,
			1,
			1,
			providerLatestIssueAt,
		))

	mock.ExpectQuery(marketProviderQualityLatestIssueQueryPattern).
		WithArgs("STOCK", "DAILY_BARS", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"source_key",
			"issue_code",
			"created_at",
		}).AddRow(
			"TUSHARE",
			"SOURCE_FETCH_FAILED",
			providerLatestIssueAt,
		))

	mock.ExpectQuery(marketDerivedTruthSummaryQueryPattern).
		WithArgs("STOCK", "LOCAL_TRUTH", "DERIVED_STOCK_STATUS_REBUILT").
		WillReturnRows(sqlmock.NewRows([]string{
			"asset_class",
			"source_key",
			"issue_code",
			"issue_message",
			"payload_json",
			"created_at",
		}).AddRow(
			"STOCK",
			"LOCAL_TRUTH",
			"DERIVED_STOCK_STATUS_REBUILT",
			"rebuilt stock status truth for 12 rows",
			`{"asset_class":"STOCK","trade_date":"2026-03-22","days":1,"truth_bar_count":120,"stock_status_count":12}`,
			truthCreatedAt,
		))

	overview, err := repo.AdminGetMarketProviderGovernanceOverview("stock", "daily_bars", 24)
	if err != nil {
		t.Fatalf("AdminGetMarketProviderGovernanceOverview returned error: %v", err)
	}
	if overview.AssetClass != "STOCK" || overview.DataKind != "DAILY_BARS" || overview.LookbackHours != 24 {
		t.Fatalf("unexpected overview head: %+v", overview)
	}
	if overview.QualitySummary.TotalCount != 4 || overview.QualitySummary.LatestIssueCode != "BAR_UPSERT_RETRIED" {
		t.Fatalf("unexpected quality summary in overview: %+v", overview.QualitySummary)
	}
	if len(overview.ProviderScores) != 1 || overview.ProviderScores[0].ProviderKey != "TUSHARE" {
		t.Fatalf("unexpected provider scores in overview: %+v", overview.ProviderScores)
	}
	if overview.ProviderScores[0].GovernanceSuggestion == "" {
		t.Fatalf("expected governance suggestion in overview provider score, got %+v", overview.ProviderScores[0])
	}
	if overview.LatestDerivedTruth == nil || overview.LatestDerivedTruth.IssueCode != "DERIVED_STOCK_STATUS_REBUILT" {
		t.Fatalf("unexpected latest derived truth in overview: %+v", overview.LatestDerivedTruth)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminRebuildMarketDerivedTruthStocksUsesTruthBarsAndInstrumentMaster(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(marketDerivedTruthMaxTradeDateQueryPattern).
		WithArgs(marketAssetClassStock).
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local)))
	mock.ExpectQuery(marketDerivedTruthTruthBarsQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-21", "2026-03-22").
		WillReturnRows(sqlmock.NewRows([]string{
			"asset_class",
			"instrument_key",
			"trade_date",
			"selected_source_key",
			"open_price",
			"high_price",
			"low_price",
			"close_price",
			"prev_close_price",
			"settle_price",
			"prev_settle_price",
			"volume",
			"turnover",
			"open_interest",
		}).AddRow(
			marketAssetClassStock, "600519.SH", time.Date(2026, 3, 21, 0, 0, 0, 0, time.Local), "TUSHARE", 1700, 1710, 1695, 1705, 1698, 0, 0, 1000000, 980000000, 0,
		).AddRow(
			marketAssetClassStock, "600519.SH", time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local), "TUSHARE", 1705, 1715, 1700, 1712, 1705, 0, 0, 1100000, 990000000, 0,
		))
	mock.ExpectQuery(marketDerivedTruthStockBaseQueryPattern).
		WithArgs(marketAssetClassStock, "600519.SH").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "display_name", "list_date", "metadata_json"}).
			AddRow("600519.SH", "贵州茅台", time.Date(2001, 8, 27, 0, 0, 0, 0, time.Local), `{"risk_warning":false}`))
	mock.ExpectExec(marketDerivedTruthStockStatusUpsertPattern).
		WithArgs(
			sqlmock.AnyArg(),
			"2026-03-21",
			"600519.SH",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			0,
			0,
			0,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(marketDerivedTruthStockStatusUpsertPattern).
		WithArgs(
			sqlmock.AnyArg(),
			"2026-03-22",
			"600519.SH",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			0,
			0,
			0,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(marketQualityLogInsertPattern).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"DAILY_BARS",
			nil,
			sqlmock.AnyArg(),
			"LOCAL_TRUTH",
			"INFO",
			"DERIVED_STOCK_STATUS_REBUILT",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := repo.AdminRebuildMarketDerivedTruth(marketAssetClassStock, "", 2)
	if err != nil {
		t.Fatalf("AdminRebuildMarketDerivedTruth returned error: %v", err)
	}
	if result.AssetClass != marketAssetClassStock {
		t.Fatalf("expected STOCK asset class, got %+v", result)
	}
	if result.EndDate != "2026-03-22" || result.StartDate != "2026-03-21" {
		t.Fatalf("unexpected rebuild date window: %+v", result)
	}
	if result.TruthBarCount != 2 || result.StockStatusCount != 2 {
		t.Fatalf("unexpected stock rebuild counts: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminRebuildMarketDerivedTruthStocksAllowsNullSettlementFields(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(marketDerivedTruthMaxTradeDateQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-22").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local)))
	mock.ExpectQuery(marketDerivedTruthTruthBarsQueryPattern).
		WithArgs(marketAssetClassStock, "2026-03-22", "2026-03-22").
		WillReturnRows(sqlmock.NewRows([]string{
			"asset_class",
			"instrument_key",
			"trade_date",
			"selected_source_key",
			"open_price",
			"high_price",
			"low_price",
			"close_price",
			"prev_close_price",
			"settle_price",
			"prev_settle_price",
			"volume",
			"turnover",
			"open_interest",
		}).AddRow(
			marketAssetClassStock, "600519.SH", time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local), "TUSHARE", 1705, 1715, 1700, 1712, 1705, nil, nil, 1100000, 990000000, 0,
		))
	mock.ExpectQuery(marketDerivedTruthStockBaseQueryPattern).
		WithArgs(marketAssetClassStock, "600519.SH").
		WillReturnRows(sqlmock.NewRows([]string{"instrument_key", "display_name", "list_date", "metadata_json"}).
			AddRow("600519.SH", "贵州茅台", time.Date(2001, 8, 27, 0, 0, 0, 0, time.Local), `{"risk_warning":false}`))
	mock.ExpectExec(marketDerivedTruthStockStatusUpsertPattern).
		WithArgs(
			sqlmock.AnyArg(),
			"2026-03-22",
			"600519.SH",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			0,
			0,
			0,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(marketQualityLogInsertPattern).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"DAILY_BARS",
			nil,
			sqlmock.AnyArg(),
			"LOCAL_TRUTH",
			"INFO",
			"DERIVED_STOCK_STATUS_REBUILT",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := repo.AdminRebuildMarketDerivedTruth(marketAssetClassStock, "2026-03-22", 1)
	if err != nil {
		t.Fatalf("AdminRebuildMarketDerivedTruth returned error: %v", err)
	}
	if result.TruthBarCount != 1 || result.StockStatusCount != 1 {
		t.Fatalf("unexpected stock rebuild counts: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminRebuildMarketDerivedTruthFuturesBuildsDominantMappings(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(marketDerivedTruthMaxTradeDateQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-22").
		WillReturnRows(sqlmock.NewRows([]string{"max_trade_date"}).AddRow(time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local)))
	mock.ExpectQuery(marketDerivedTruthTruthBarsQueryPattern).
		WithArgs(marketAssetClassFutures, "2026-03-22", "2026-03-22").
		WillReturnRows(sqlmock.NewRows([]string{
			"asset_class",
			"instrument_key",
			"trade_date",
			"selected_source_key",
			"open_price",
			"high_price",
			"low_price",
			"close_price",
			"prev_close_price",
			"settle_price",
			"prev_settle_price",
			"volume",
			"turnover",
			"open_interest",
		}).AddRow(
			marketAssetClassFutures, "IF2606.CFX", time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local), "TUSHARE", 0, 0, 0, 0, 0, 0, 0, 10000, 1200000, 320000,
		).AddRow(
			marketAssetClassFutures, "IF2605.CFX", time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local), "TUSHARE", 0, 0, 0, 0, 0, 0, 0, 9000, 900000, 280000,
		))
	mock.ExpectExec(marketDerivedTruthFuturesMappingUpsertPattern).
		WithArgs(
			sqlmock.AnyArg(),
			"2026-03-22",
			"IF",
			"CFX",
			"IF2606.CFX",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"TURNOVER_OPEN_INTEREST",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(marketQualityLogInsertPattern).
		WithArgs(
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"DAILY_BARS",
			nil,
			sqlmock.AnyArg(),
			"LOCAL_TRUTH",
			"INFO",
			"DERIVED_FUTURES_MAPPING_REBUILT",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	result, err := repo.AdminRebuildMarketDerivedTruth(marketAssetClassFutures, "2026-03-22", 1)
	if err != nil {
		t.Fatalf("AdminRebuildMarketDerivedTruth returned error: %v", err)
	}
	if result.AssetClass != marketAssetClassFutures {
		t.Fatalf("expected FUTURES asset class, got %+v", result)
	}
	if result.TruthBarCount != 2 || result.FuturesMappingCount != 1 {
		t.Fatalf("unexpected futures rebuild counts: %+v", result)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
