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

	relatedEntities := []model.StrategyExplanationRelatedEntity{
		{
			EntityType: "ConceptTheme",
			EntityKey:  "HIGH-END-LIQUOR",
			Label:      "高端白酒",
		},
	}

	snapshot := buildStockRelationshipSnapshot(ctx, relatedEntities)
	if snapshot.AssetKey != "600519.SH" {
		t.Fatalf("expected asset key to be carried, got %+v", snapshot)
	}
	if snapshot.RelationshipCount < 4 {
		t.Fatalf("expected stock relationship snapshot to collect multiple nodes, got %+v", snapshot)
	}
}

func TestBuildFuturesRelationshipSnapshotCollectsSupplyChainAndStructureSignals(t *testing.T) {
	ctx := strategyEngineAssetContext{
		record: model.StrategyEnginePublishRecord{TradeDate: "2026-03-29"},
		asset: map[string]any{
			"contract": "RB2609",
			"name":     "螺纹钢主力",
		},
		simulation: map[string]any{
			"related_events": []any{
				map[string]any{
					"title":      "地产施工节奏回暖",
					"event_type": "MACRO",
				},
			},
		},
	}

	relatedEntities := []model.StrategyExplanationRelatedEntity{
		{
			EntityType: "Commodity",
			EntityKey:  "RB",
			Label:      "螺纹钢",
		},
		{
			EntityType: "Warehouse",
			EntityKey:  "WAREHOUSE:SH",
			Label:      "上海主仓",
		},
	}

	snapshot := buildFuturesRelationshipSnapshot(
		ctx,
		relatedEntities,
		"社会库存连续三周去化",
		"RB-HC 价差结构继续走强",
	)
	if snapshot.AssetKey != "RB2609" {
		t.Fatalf("expected contract key to be carried, got %+v", snapshot)
	}
	if snapshot.RelationshipCount < 5 {
		t.Fatalf("expected futures relationship snapshot to collect supply chain and structure nodes, got %+v", snapshot)
	}
}
