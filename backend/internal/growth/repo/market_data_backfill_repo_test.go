package repo

import (
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"sercherai/backend/internal/growth/model"
)

const marketBackfillRunByIDQueryPattern = `(?s)SELECT id, scheduler_run_id, run_type, COALESCE\(CAST\(asset_scope AS CHAR\), ''\),\s*COALESCE\(DATE_FORMAT\(trade_date_from, '%Y-%m-%d'\), ''\),\s*COALESCE\(DATE_FORMAT\(trade_date_to, '%Y-%m-%d'\), ''\),\s*COALESCE\(source_key, ''\), batch_size,\s*universe_snapshot_id, status, current_stage, COALESCE\(CAST\(stage_progress_json AS CHAR\), ''\),\s*COALESCE\(CAST\(summary_json AS CHAR\), ''\), COALESCE\(error_message, ''\), COALESCE\(created_by, ''\),\s*created_at, updated_at, finished_at\s*FROM market_backfill_runs\s*WHERE id = \?`

const marketBackfillSnapshotItemsQueryPattern = `(?s)SELECT id, snapshot_id, asset_type, instrument_key, COALESCE\(external_symbol, ''\), COALESCE\(display_name, ''\),\s*COALESCE\(exchange_code, ''\), COALESCE\(status, ''\), COALESCE\(DATE_FORMAT\(list_date, '%Y-%m-%d'\), ''\),\s*COALESCE\(DATE_FORMAT\(delist_date, '%Y-%m-%d'\), ''\), COALESCE\(CAST\(raw_metadata_json AS CHAR\), ''\), created_at\s*FROM market_universe_snapshot_items\s*WHERE snapshot_id = \?\s*ORDER BY asset_type ASC, instrument_key ASC`

const marketInstrumentSourceFactsQueryPattern = `(?s)SELECT\s+asset_class,.*?source_updated_at\s+FROM market_instrument_source_facts`

const marketBackfillDetailInsertPattern = `INSERT INTO market_backfill_run_details`

const marketBackfillRunUpdatePattern = `UPDATE market_backfill_runs`

const schedulerJobRunUpdatePattern = `UPDATE scheduler_job_runs`

const marketInstrumentsUpsertPattern = `INSERT INTO market_instruments`

const marketInstrumentFactsUpsertPattern = `INSERT INTO market_instrument_source_facts`

const marketSymbolAliasUpsertPattern = `INSERT INTO market_symbol_aliases`

const marketDailyBarsUpsertPattern = `INSERT INTO market_daily_bars`

const marketDailyBarTruthInsertPattern = `INSERT INTO market_daily_bar_truth`

const stockDailyBasicUpsertPattern = `INSERT INTO stock_daily_basic`

const stockMoneyflowUpsertPattern = `INSERT INTO stock_moneyflow_daily`

func TestAdminBuildMarketUniverseSnapshotIncludesRequestedAssetScope(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	snapshot, items, err := repo.AdminBuildMarketUniverseSnapshot("TUSHARE", []string{"STOCK", "INDEX", "ETF"}, "tester")
	if err != nil {
		t.Fatalf("AdminBuildMarketUniverseSnapshot returned error: %v", err)
	}
	if snapshot.ID == "" {
		t.Fatal("expected snapshot id")
	}
	if len(snapshot.Scope) != 3 {
		t.Fatalf("expected 3 scope items, got %+v", snapshot.Scope)
	}
	if len(snapshot.AssetSummaries) != 3 {
		t.Fatalf("expected 3 asset summaries, got %+v", snapshot.AssetSummaries)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 universe items, got %+v", items)
	}
	if items[0].SnapshotID != snapshot.ID {
		t.Fatalf("expected snapshot id %s, got %+v", snapshot.ID, items[0])
	}
}

func TestAdminBuildMarketUniverseSnapshotRejectsEmptyScope(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	_, _, err := repo.AdminBuildMarketUniverseSnapshot("TUSHARE", nil, "tester")
	if err == nil {
		t.Fatal("expected error when asset scope is empty")
	}
}

func TestAdminCreateMarketDataBackfillRunBuildsUniverseSnapshotPerAsset(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	run, err := repo.AdminCreateMarketDataBackfillRun(model.MarketBackfillCreateInput{
		RunType:    "FULL",
		AssetScope: []string{"STOCK", "INDEX", "ETF"},
		SourceKey:  "TUSHARE",
		BatchSize:  200,
	}, "tester")
	if err != nil {
		t.Fatalf("AdminCreateMarketDataBackfillRun returned error: %v", err)
	}
	if run.UniverseSnapshotID == "" {
		t.Fatal("expected universe snapshot id")
	}
	if run.CurrentStage != "COVERAGE_SUMMARY" {
		t.Fatalf("expected final stage COVERAGE_SUMMARY after immediate execution, got %s", run.CurrentStage)
	}
	if len(run.StageProgress) == 0 || run.StageProgress[0].Stage != "UNIVERSE" || run.StageProgress[0].Status != "SUCCESS" {
		t.Fatalf("expected universe stage success progress, got %+v", run.StageProgress)
	}

	snapshot, items, err := repo.AdminGetMarketUniverseSnapshot(run.UniverseSnapshotID)
	if err != nil {
		t.Fatalf("AdminGetMarketUniverseSnapshot returned error: %v", err)
	}
	if len(snapshot.AssetSummaries) != 3 {
		t.Fatalf("expected 3 asset summaries, got %+v", snapshot.AssetSummaries)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 snapshot items, got %+v", items)
	}

	details, total, err := repo.AdminListMarketDataBackfillRunDetails(run.ID, "UNIVERSE", "", "", 1, 10)
	if err != nil {
		t.Fatalf("AdminListMarketDataBackfillRunDetails returned error: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected 3 universe details, got %d", total)
	}
	if len(details) != 3 {
		t.Fatalf("expected 3 returned details, got %+v", details)
	}
}

func TestAdminSyncMarketDailyBasicDetailedSkipsNonStockAssets(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	result, err := repo.AdminSyncMarketDailyBasicDetailed("INDEX", "TUSHARE", []string{"000300.SH"}, 30)
	if err != nil {
		t.Fatalf("AdminSyncMarketDailyBasicDetailed returned error: %v", err)
	}
	if result.DataKind != "DAILY_BASIC" {
		t.Fatalf("expected data kind DAILY_BASIC, got %s", result.DataKind)
	}
	if result.BarCount != 0 {
		t.Fatalf("expected no synced rows for INDEX, got %d", result.BarCount)
	}
	if len(result.Results) != 1 || result.Results[0].Status != "SKIPPED" {
		t.Fatalf("expected skipped result, got %+v", result.Results)
	}
}

func TestAdminSyncMarketMoneyflowDetailedSupportsStocks(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	result, err := repo.AdminSyncMarketMoneyflowDetailed("STOCK", "TUSHARE", []string{"600519.SH", "000001.SZ"}, 20)
	if err != nil {
		t.Fatalf("AdminSyncMarketMoneyflowDetailed returned error: %v", err)
	}
	if result.DataKind != "MONEYFLOW" {
		t.Fatalf("expected data kind MONEYFLOW, got %s", result.DataKind)
	}
	if result.BarCount != 40 {
		t.Fatalf("expected 40 synced rows, got %d", result.BarCount)
	}
	if len(result.Results) != 1 || result.Results[0].Status != "SUCCESS" {
		t.Fatalf("expected success result, got %+v", result.Results)
	}
}

func TestBuildMockMarketDailyBarsPreservesRequestedAssetType(t *testing.T) {
	bars := buildMockMarketDailyBars("INDEX", "MOCK", []string{"000300.SH"}, 2)
	if len(bars) != 2 {
		t.Fatalf("expected 2 mock bars, got %d", len(bars))
	}
	for _, bar := range bars {
		if bar.AssetClass != "INDEX" {
			t.Fatalf("expected mock bar asset class INDEX, got %+v", bar)
		}
	}
}

func TestExecuteMarketDataBackfillRunMarksEnhancementStagesBySupportMatrix(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	executed, err := repo.AdminCreateMarketDataBackfillRun(model.MarketBackfillCreateInput{
		RunType:    "FULL",
		AssetScope: []string{"STOCK", "INDEX"},
		SourceKey:  "TUSHARE",
		BatchSize:  200,
	}, "tester")
	if err != nil {
		t.Fatalf("AdminCreateMarketDataBackfillRun returned error: %v", err)
	}
	if executed.Status != "SUCCESS" {
		t.Fatalf("expected SUCCESS run, got %+v", executed)
	}
	if executed.CurrentStage != "COVERAGE_SUMMARY" {
		t.Fatalf("expected final current stage COVERAGE_SUMMARY, got %s", executed.CurrentStage)
	}

	dailyBasicDetails, total, err := repo.AdminListMarketDataBackfillRunDetails(executed.ID, "DAILY_BASIC", "", "", 1, 10)
	if err != nil {
		t.Fatalf("AdminListMarketDataBackfillRunDetails daily basic returned error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 daily basic details, got %d", total)
	}
	if dailyBasicDetails[0].Status != "SUCCESS" && dailyBasicDetails[1].Status != "SUCCESS" {
		t.Fatalf("expected one success detail, got %+v", dailyBasicDetails)
	}
	if dailyBasicDetails[0].Status != "SKIPPED" && dailyBasicDetails[1].Status != "SKIPPED" {
		t.Fatalf("expected one skipped detail, got %+v", dailyBasicDetails)
	}

	moneyflowDetails, total, err := repo.AdminListMarketDataBackfillRunDetails(executed.ID, "MONEYFLOW", "", "", 1, 10)
	if err != nil {
		t.Fatalf("AdminListMarketDataBackfillRunDetails moneyflow returned error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 moneyflow details, got %d", total)
	}
	if moneyflowDetails[0].Status != "SUCCESS" && moneyflowDetails[1].Status != "SUCCESS" {
		t.Fatalf("expected one success moneyflow detail, got %+v", moneyflowDetails)
	}
	if moneyflowDetails[0].Status != "SKIPPED" && moneyflowDetails[1].Status != "SKIPPED" {
		t.Fatalf("expected one skipped moneyflow detail, got %+v", moneyflowDetails)
	}
}

func TestAdminCreateMarketDataBackfillRunExecutesImmediately(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	run, err := repo.AdminCreateMarketDataBackfillRun(model.MarketBackfillCreateInput{
		RunType:       "FULL",
		AssetScope:    []string{"STOCK", "INDEX"},
		SourceKey:     "MOCK",
		TradeDateFrom: "2026-03-23",
		TradeDateTo:   "2026-03-24",
		BatchSize:     200,
	}, "tester")
	if err != nil {
		t.Fatalf("AdminCreateMarketDataBackfillRun returned error: %v", err)
	}
	if run.Status != "SUCCESS" {
		t.Fatalf("expected immediate SUCCESS run, got %+v", run)
	}
	if run.CurrentStage != "COVERAGE_SUMMARY" {
		t.Fatalf("expected final stage COVERAGE_SUMMARY, got %s", run.CurrentStage)
	}
}

func TestAdminCreateMarketDataBackfillRunRejectsLongHistoryWithNonTushareSource(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	_, err := repo.AdminCreateMarketDataBackfillRun(model.MarketBackfillCreateInput{
		RunType:       "FULL",
		AssetScope:    []string{"STOCK"},
		SourceKey:     "AKSHARE",
		TradeDateFrom: "2024-01-01",
		TradeDateTo:   "2025-01-05",
		BatchSize:     100,
	}, "tester")
	if err == nil {
		t.Fatal("expected long-history validation error")
	}
	if !strings.Contains(err.Error(), "当前仅支持 TUSHARE") {
		t.Fatalf("expected tushare-only validation, got %v", err)
	}
}

func TestAdminCreateMarketDataBackfillRunRejectsLongHistoryWithoutQuotesStage(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	_, err := repo.AdminCreateMarketDataBackfillRun(model.MarketBackfillCreateInput{
		RunType:       "FULL",
		AssetScope:    []string{"STOCK"},
		SourceKey:     "TUSHARE",
		TradeDateFrom: "2024-01-01",
		TradeDateTo:   "2025-01-05",
		BatchSize:     100,
		Stages:        []string{"MASTER", "TRUTH"},
	}, "tester")
	if err == nil {
		t.Fatal("expected missing quotes validation error")
	}
	if !strings.Contains(err.Error(), "必须包含 QUOTES") {
		t.Fatalf("expected quotes-stage validation, got %v", err)
	}
}

func TestExecuteMarketDataBackfillRunSplitsLongHistoryStockQuotesAndSkipsUnsupportedStages(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	run, err := repo.AdminCreateMarketDataBackfillRun(model.MarketBackfillCreateInput{
		RunType:               "FULL",
		AssetScope:            []string{"STOCK"},
		SourceKey:             "TUSHARE",
		TradeDateFrom:         "2024-01-01",
		TradeDateTo:           "2025-01-05",
		BatchSize:             1,
		RebuildTruthAfterSync: false,
	}, "tester")
	if err != nil {
		t.Fatalf("AdminCreateMarketDataBackfillRun returned error: %v", err)
	}

	longHistoryMode, ok := run.Summary["long_history_mode"].(bool)
	if !ok || !longHistoryMode {
		t.Fatalf("expected long_history_mode=true in summary, got %+v", run.Summary)
	}
	if run.CurrentStage != "COVERAGE_SUMMARY" {
		t.Fatalf("expected final stage COVERAGE_SUMMARY, got %s", run.CurrentStage)
	}

	quotesDetails, total, err := repo.AdminListMarketDataBackfillRunDetails(run.ID, "QUOTES", "STOCK", "", 1, 20)
	if err != nil {
		t.Fatalf("AdminListMarketDataBackfillRunDetails returned error: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected 3 long-history quote chunks, got %d", total)
	}
	for _, detail := range quotesDetails {
		if detail.TradeDateFrom == "" || detail.TradeDateTo == "" {
			t.Fatalf("expected quote chunk date range, got %+v", detail)
		}
		if detail.SymbolCount != 1 {
			t.Fatalf("expected stock batch size 1, got %+v", detail)
		}
		if detail.Status != "SUCCESS" {
			t.Fatalf("expected quote chunk success, got %+v", detail)
		}
	}

	dailyBasicDetails, total, err := repo.AdminListMarketDataBackfillRunDetails(run.ID, "DAILY_BASIC", "STOCK", "", 1, 10)
	if err != nil {
		t.Fatalf("AdminListMarketDataBackfillRunDetails daily basic returned error: %v", err)
	}
	if total != 1 || len(dailyBasicDetails) != 1 {
		t.Fatalf("expected 1 daily basic detail, got total=%d items=%+v", total, dailyBasicDetails)
	}
	if dailyBasicDetails[0].Status != "SKIPPED" || !strings.Contains(dailyBasicDetails[0].WarningText, "长历史") {
		t.Fatalf("expected long-history daily basic skip, got %+v", dailyBasicDetails[0])
	}

	moneyflowDetails, total, err := repo.AdminListMarketDataBackfillRunDetails(run.ID, "MONEYFLOW", "STOCK", "", 1, 10)
	if err != nil {
		t.Fatalf("AdminListMarketDataBackfillRunDetails moneyflow returned error: %v", err)
	}
	if total != 1 || len(moneyflowDetails) != 1 {
		t.Fatalf("expected 1 moneyflow detail, got total=%d items=%+v", total, moneyflowDetails)
	}
	if moneyflowDetails[0].Status != "SKIPPED" || !strings.Contains(moneyflowDetails[0].WarningText, "长历史") {
		t.Fatalf("expected long-history moneyflow skip, got %+v", moneyflowDetails[0])
	}

	truthDetails, total, err := repo.AdminListMarketDataBackfillRunDetails(run.ID, "TRUTH", "STOCK", "", 1, 10)
	if err != nil {
		t.Fatalf("AdminListMarketDataBackfillRunDetails truth returned error: %v", err)
	}
	if total != 1 || len(truthDetails) != 1 {
		t.Fatalf("expected 1 truth detail, got total=%d items=%+v", total, truthDetails)
	}
	if truthDetails[0].Status != "SKIPPED" {
		t.Fatalf("expected truth skipped when rebuild_truth_after_sync=false, got %+v", truthDetails[0])
	}
}

func TestExecuteMarketDataBackfillRunLongHistoryRebuildsTruthFromQuoteChunksWhenRequested(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	run, err := repo.AdminCreateMarketDataBackfillRun(model.MarketBackfillCreateInput{
		RunType:               "FULL",
		AssetScope:            []string{"STOCK"},
		SourceKey:             "TUSHARE",
		TradeDateFrom:         "2024-01-01",
		TradeDateTo:           "2025-01-05",
		BatchSize:             1,
		RebuildTruthAfterSync: true,
	}, "tester")
	if err != nil {
		t.Fatalf("AdminCreateMarketDataBackfillRun returned error: %v", err)
	}

	quotesDetails, total, err := repo.AdminListMarketDataBackfillRunDetails(run.ID, "QUOTES", "STOCK", "", 1, 20)
	if err != nil {
		t.Fatalf("AdminListMarketDataBackfillRunDetails quotes returned error: %v", err)
	}
	if total != 3 || len(quotesDetails) != 3 {
		t.Fatalf("expected 3 long-history quote details, got total=%d items=%+v", total, quotesDetails)
	}
	expectedTruthCount := 0
	for _, detail := range quotesDetails {
		expectedTruthCount += detail.UpsertedCount
	}

	truthDetails, total, err := repo.AdminListMarketDataBackfillRunDetails(run.ID, "TRUTH", "STOCK", "", 1, 10)
	if err != nil {
		t.Fatalf("AdminListMarketDataBackfillRunDetails truth returned error: %v", err)
	}
	if total != 1 || len(truthDetails) != 1 {
		t.Fatalf("expected 1 truth detail, got total=%d items=%+v", total, truthDetails)
	}
	if truthDetails[0].Status != "SUCCESS" {
		t.Fatalf("expected truth success when rebuild_truth_after_sync=true, got %+v", truthDetails[0])
	}
	if truthDetails[0].TruthCount != expectedTruthCount {
		t.Fatalf("expected truth count %d from quote chunks, got %+v", expectedTruthCount, truthDetails[0])
	}
}

func TestAdminRetryMarketDataBackfillRunExecutesImmediately(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	base, err := repo.AdminCreateMarketDataBackfillRun(model.MarketBackfillCreateInput{
		RunType:       "FULL",
		AssetScope:    []string{"STOCK", "INDEX"},
		SourceKey:     "MOCK",
		TradeDateFrom: "2026-03-23",
		TradeDateTo:   "2026-03-24",
		BatchSize:     200,
	}, "tester")
	if err != nil {
		t.Fatalf("AdminCreateMarketDataBackfillRun returned error: %v", err)
	}

	retried, err := repo.AdminRetryMarketDataBackfillRun(base.ID, model.MarketBackfillRetryInput{
		RetryMode: "FROM_STAGE",
		Stage:     "QUOTES",
	}, "tester")
	if err != nil {
		t.Fatalf("AdminRetryMarketDataBackfillRun returned error: %v", err)
	}
	if retried.ID == base.ID {
		t.Fatalf("expected new run id, got %+v", retried)
	}
	if retried.Status != "SUCCESS" {
		t.Fatalf("expected retried run success, got %+v", retried)
	}
	if retried.CurrentStage != "COVERAGE_SUMMARY" {
		t.Fatalf("expected retried run final stage COVERAGE_SUMMARY, got %s", retried.CurrentStage)
	}
}

func TestMySQLExecuteMarketDataBackfillRunCompletesMockPipeline(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()
	mock.MatchExpectationsInOrder(false)

	repo := &MySQLGrowthRepo{db: db}
	createdAt := time.Date(2026, 3, 24, 9, 0, 0, 0, time.Local)

	mock.ExpectQuery(marketBackfillRunByIDQueryPattern).
		WithArgs("mbr_001").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"scheduler_run_id",
			"run_type",
			"asset_scope",
			"trade_date_from",
			"trade_date_to",
			"source_key",
			"batch_size",
			"universe_snapshot_id",
			"status",
			"current_stage",
			"stage_progress_json",
			"summary_json",
			"error_message",
			"created_by",
			"created_at",
			"updated_at",
			"finished_at",
		}).AddRow(
			"mbr_001",
			"jr_001",
			"FULL",
			`["STOCK","INDEX"]`,
			"2026-03-23",
			"2026-03-24",
			"MOCK",
			200,
			"mus_001",
			"RUNNING",
			"MASTER",
			`[{"stage":"UNIVERSE","status":"SUCCESS","total_batches":2,"completed_batches":2},{"stage":"MASTER","status":"PENDING"},{"stage":"QUOTES","status":"PENDING"},{"stage":"DAILY_BASIC","status":"PENDING"},{"stage":"MONEYFLOW","status":"PENDING"},{"stage":"TRUTH","status":"PENDING"},{"stage":"COVERAGE_SUMMARY","status":"PENDING"}]`,
			`{"requested_stages":["MASTER","QUOTES","DAILY_BASIC","MONEYFLOW","TRUTH"],"asset_scope":["STOCK","INDEX"],"universe_item_count":2}`,
			"",
			"tester",
			createdAt,
			createdAt,
			nil,
		))

	mock.ExpectQuery(marketBackfillSnapshotItemsQueryPattern).
		WithArgs("mus_001").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"snapshot_id",
			"asset_type",
			"instrument_key",
			"external_symbol",
			"display_name",
			"exchange_code",
			"status",
			"list_date",
			"delist_date",
			"raw_metadata_json",
			"created_at",
		}).
			AddRow("musi_stock", "mus_001", "STOCK", "600519.SH", "600519.SH", "贵州茅台", "SH", "ACTIVE", "2001-08-27", "", `{"industry":"白酒"}`, createdAt).
			AddRow("musi_index", "mus_001", "INDEX", "000300.SH", "000300.SH", "沪深300", "SH", "ACTIVE", "2005-04-08", "", `{"category":"broad_index"}`, createdAt))

	for range []int{0, 1} {
		mock.ExpectExec(marketInstrumentsUpsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	for range []int{0, 1} {
		mock.ExpectExec(marketInstrumentFactsUpsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	for range []int{0, 1} {
		mock.ExpectExec(marketSymbolAliasUpsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}

	mock.ExpectQuery(marketInstrumentSourceFactsQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{
			"asset_class",
			"instrument_key",
			"source_key",
			"external_symbol",
			"display_name",
			"exchange_code",
			"product_key",
			"list_date",
			"delist_date",
			"status",
			"metadata_json",
			"quality_score",
			"source_updated_at",
		}).AddRow("INDEX", "000300.SH", "MOCK", "000300.SH", "沪深300", "SH", "INDEX:000300", "2005-04-08", "", "ACTIVE", `{"category":"broad_index"}`, 0.93, createdAt))
	mock.ExpectExec(`UPDATE market_instruments`).WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectQuery(marketInstrumentSourceFactsQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{
			"asset_class",
			"instrument_key",
			"source_key",
			"external_symbol",
			"display_name",
			"exchange_code",
			"product_key",
			"list_date",
			"delist_date",
			"status",
			"metadata_json",
			"quality_score",
			"source_updated_at",
		}).AddRow("STOCK", "600519.SH", "MOCK", "600519.SH", "贵州茅台", "SH", "STOCK:600519", "2001-08-27", "", "ACTIVE", `{"industry":"白酒"}`, 0.96, createdAt))
	mock.ExpectExec(`UPDATE market_instruments`).WillReturnResult(sqlmock.NewResult(1, 1))

	for range []int{0, 1} {
		mock.ExpectExec(marketBackfillDetailInsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	for range []int{0, 1, 2, 3} {
		mock.ExpectExec(marketDailyBarsUpsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	for range []int{0, 1} {
		mock.ExpectExec(marketBackfillDetailInsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	for range []int{0, 1} {
		mock.ExpectExec(stockDailyBasicUpsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	for range []int{0, 1} {
		mock.ExpectExec(marketBackfillDetailInsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	for range []int{0, 1} {
		mock.ExpectExec(stockMoneyflowUpsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	for range []int{0, 1} {
		mock.ExpectExec(marketBackfillDetailInsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	for _, assetType := range []string{"INDEX", "STOCK"} {
		instrumentKey := map[string]string{"INDEX": "000300.SH", "STOCK": "600519.SH"}[assetType]
		for _, payload := range []struct {
			tradeDate time.Time
			open      float64
			high      float64
			low       float64
			close     float64
			prevClose float64
			volume    int
			turnover  int
		}{
			{tradeDate: time.Date(2026, 3, 23, 0, 0, 0, 0, time.Local), open: 10, high: 11, low: 9, close: 10.5, prevClose: 10, volume: 1000, turnover: 10000},
			{tradeDate: time.Date(2026, 3, 24, 0, 0, 0, 0, time.Local), open: 10.5, high: 11.2, low: 10.1, close: 10.9, prevClose: 10.5, volume: 1100, turnover: 11990},
		} {
			mock.ExpectQuery(`SELECT instrument_key, external_symbol, trade_date, open_price, high_price, low_price, close_price`).
				WithArgs(assetType, sqlmock.AnyArg(), sqlmock.AnyArg()).
				WillReturnRows(sqlmock.NewRows([]string{
					"instrument_key",
					"external_symbol",
					"trade_date",
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
					"source_key",
				}).AddRow(
					instrumentKey,
					instrumentKey,
					payload.tradeDate,
					payload.open,
					payload.high,
					payload.low,
					payload.close,
					payload.prevClose,
					0,
					0,
					payload.volume,
					payload.turnover,
					0,
					"MOCK",
				))
			mock.ExpectExec(marketDailyBarTruthInsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
		}
	}
	for range []int{0, 1} {
		mock.ExpectExec(marketBackfillDetailInsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	}
	mock.ExpectExec(marketBackfillDetailInsertPattern).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(marketBackfillRunUpdatePattern).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(schedulerJobRunUpdatePattern).WillReturnResult(sqlmock.NewResult(1, 1))

	run, err := repo.executeMarketDataBackfillRun("mbr_001")
	if err != nil {
		t.Fatalf("executeMarketDataBackfillRun returned error: %v", err)
	}
	if run.Status != "SUCCESS" {
		t.Fatalf("expected SUCCESS run, got %+v", run)
	}
	if run.CurrentStage != "COVERAGE_SUMMARY" {
		t.Fatalf("expected final stage COVERAGE_SUMMARY, got %s", run.CurrentStage)
	}
	if got := run.Summary["window_days"]; got != 2 {
		t.Fatalf("expected window_days summary 2, got %+v", run.Summary)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
