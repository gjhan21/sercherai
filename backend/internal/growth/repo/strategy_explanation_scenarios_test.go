package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestBuildStockScenarioSnapshotsReturnsBullBaseBear(t *testing.T) {
	ctx := strategyEngineAssetContext{
		record: model.StrategyEnginePublishRecord{TradeDate: "2026-03-29"},
		asset: map[string]any{
			"symbol":         "600519.SH",
			"reason_summary": "资金回流叠加趋势延续",
			"risk_summary":   "跌破 5 日线失效",
			"invalidations":  []any{"跌破 5 日线"},
		},
	}

	scenarios, meta := buildStockScenarioSnapshots(ctx)
	if len(scenarios) != 3 {
		t.Fatalf("expected 3 stable scenarios, got %+v", scenarios)
	}
	if scenarios[0].Scenario == "" || scenarios[0].Trigger == "" || scenarios[0].ActionSuggestion == "" {
		t.Fatalf("expected complete scenario contract, got %+v", scenarios[0])
	}
	if meta.PrimaryScenario == "" {
		t.Fatalf("expected scenario meta, got %+v", meta)
	}
}

func TestBuildStockScenarioSnapshotsUsesBaseAsPrimaryScenario(t *testing.T) {
	ctx := strategyEngineAssetContext{
		record: model.StrategyEnginePublishRecord{TradeDate: "2026-03-29"},
		asset: map[string]any{
			"symbol":         "600519.SH",
			"reason_summary": "资金回流叠加趋势延续",
			"risk_summary":   "跌破 5 日线失效",
			"invalidations":  []any{"跌破 5 日线"},
		},
	}

	_, meta := buildStockScenarioSnapshots(ctx)
	if meta.PrimaryScenario != "base" {
		t.Fatalf("expected base to be the primary scenario, got %+v", meta)
	}
}
