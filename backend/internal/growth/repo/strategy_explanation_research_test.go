package repo

import (
	"strings"
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestBuildStockResearchBlocksCreatesActiveHistoricalAndWatchLayers(t *testing.T) {
	report := map[string]any{
		"market_regime": "ROTATION",
		"memory_feedback": map[string]any{
			"summary":         "高波动题材需要缩短验证窗口",
			"failure_signals": []any{"题材切换过快"},
		},
	}
	ctx := strategyEngineAssetContext{
		record: model.StrategyEnginePublishRecord{TradeDate: "2026-03-28", ReportSnapshot: report},
		asset: map[string]any{
			"symbol":         "600519.SH",
			"reason_summary": "资金回流叠加趋势延续",
			"risk_summary":   "跌破 5 日线需减仓",
		},
	}

	outline, active, historical, watch := buildStockResearchBlocks(ctx, "600519.SH")

	if len(outline) == 0 || outline[0].Slot == "" {
		t.Fatalf("expected research outline to be built: %+v", outline)
	}
	if len(active) != 1 || !strings.Contains(active[0].Summary, "趋势") {
		t.Fatalf("expected active thesis to be created from reason summary: %+v", active)
	}
	if len(historical) != 1 || historical[0].Status != "WEAKENED" {
		t.Fatalf("expected historical thesis derived from memory feedback: %+v", historical)
	}
	if len(watch) != 1 || watch[0].SignalType != "INVALIDATION" {
		t.Fatalf("expected watch signal derived from risk or memory feedback: %+v", watch)
	}
}

func TestBuildFuturesResearchBlocksUsesStructureAndInventory(t *testing.T) {
	report := map[string]any{
		"market_regime": "TREND",
	}
	ctx := strategyEngineAssetContext{
		record: model.StrategyEnginePublishRecord{TradeDate: "2026-03-28", ReportSnapshot: report},
		asset: map[string]any{
			"contract":                 "RB2505",
			"reason_summary":           "库存继续回落，结构维持升水",
			"structure_factor_summary": "期限结构维持升水",
			"inventory_factor_summary": "库存持续去化",
		},
	}

	outline, active, historical, watch := buildFuturesResearchBlocks(ctx, "RB2505")

	if len(outline) == 0 || outline[0].Slot == "" {
		t.Fatalf("expected futures research outline to be built: %+v", outline)
	}
	if len(active) < 1 {
		t.Fatalf("expected futures active thesis cards: %+v", active)
	}
	if len(historical) != 0 {
		t.Fatalf("expected no historical thesis when memory feedback is absent: %+v", historical)
	}
	if len(watch) == 0 {
		t.Fatalf("expected futures watch signals even without memory feedback: %+v", watch)
	}
	if active[0].EvidenceSource == "" {
		t.Fatalf("expected evidence source to be filled for futures thesis cards: %+v", active)
	}
}
