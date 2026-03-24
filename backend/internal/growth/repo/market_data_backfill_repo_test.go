package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

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
	if run.CurrentStage != "MASTER" {
		t.Fatalf("expected current stage MASTER after snapshot build, got %s", run.CurrentStage)
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

func TestExecuteMarketDataBackfillRunMarksEnhancementStagesBySupportMatrix(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	run, err := repo.AdminCreateMarketDataBackfillRun(model.MarketBackfillCreateInput{
		RunType:    "FULL",
		AssetScope: []string{"STOCK", "INDEX"},
		SourceKey:  "TUSHARE",
		BatchSize:  200,
	}, "tester")
	if err != nil {
		t.Fatalf("AdminCreateMarketDataBackfillRun returned error: %v", err)
	}

	executed, err := repo.executeMarketDataBackfillRun(run.ID)
	if err != nil {
		t.Fatalf("executeMarketDataBackfillRun returned error: %v", err)
	}
	if executed.Status != "SUCCESS" {
		t.Fatalf("expected SUCCESS run, got %+v", executed)
	}
	if executed.CurrentStage != "COVERAGE_SUMMARY" {
		t.Fatalf("expected final current stage COVERAGE_SUMMARY, got %s", executed.CurrentStage)
	}

	dailyBasicDetails, total, err := repo.AdminListMarketDataBackfillRunDetails(run.ID, "DAILY_BASIC", "", "", 1, 10)
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

	moneyflowDetails, total, err := repo.AdminListMarketDataBackfillRunDetails(run.ID, "MONEYFLOW", "", "", 1, 10)
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
