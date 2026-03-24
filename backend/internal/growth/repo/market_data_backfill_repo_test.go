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
