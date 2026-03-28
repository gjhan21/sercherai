package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestBuildStockRelationshipSnapshotCollectsThemeSectorAndEvents(t *testing.T) {
	ctx := strategyEngineAssetContext{
		record: model.StrategyEnginePublishRecord{TradeDate: "2026-03-29"},
		asset: map[string]any{
			"symbol":      "600519.SH",
			"name":        "贵州茅台",
			"theme_tags":  []any{"白酒"},
			"sector_tags": []any{"消费"},
		},
		simulation: map[string]any{
			"related_events": []any{
				map[string]any{
					"title":      "消费刺激预期升温",
					"event_type": "POLICY",
				},
			},
		},
	}

	snapshot := buildStockRelationshipSnapshot(ctx)
	if snapshot.AssetKey != "600519.SH" {
		t.Fatalf("expected asset key to be carried, got %+v", snapshot)
	}
	if snapshot.RelationshipCount < 3 {
		t.Fatalf("expected stock relationship snapshot to collect multiple nodes, got %+v", snapshot)
	}
}
