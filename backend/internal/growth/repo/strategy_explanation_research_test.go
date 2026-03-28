package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestBuildStockResearchBlocksCreatesActiveHistoricalAndWatchLayers(t *testing.T) {
	report := map[string]any{
		"market_regime": "ROTATION",
		"memory_feedback": map[string]any{
			"summary":         "高波动题材需要缩短验证窗口",
			"failure_signals": []any{"题材切换过快"},
			"suggestions":     []any{"观察放量后的承接强度"},
		},
	}
	ctx := strategyEngineAssetContext{
		record: model.StrategyEnginePublishRecord{TradeDate: "2026-03-28"},
		asset: map[string]any{
			"symbol":         "600519.SH",
			"reason_summary": "资金回流叠加趋势延续",
			"risk_summary":   "跌破 5 日线失效",
			"risk_flags":     []any{"高位分歧"},
			"invalidations":  []any{"跌破 5 日线"},
			"theme_tags":     []any{"消费龙头"},
			"sector_tags":    []any{"消费"},
		},
	}

	outline, active, historical, watch := buildStockResearchBlocks(ctx, report, nil, nil)
	if len(outline) != 5 {
		t.Fatalf("expected 5 stock research outline steps, got %+v", outline)
	}
	if len(active) == 0 {
		t.Fatalf("expected stock active thesis cards, got %+v", active)
	}
	if len(historical) == 0 {
		t.Fatalf("expected stock historical thesis cards, got %+v", historical)
	}
	if len(watch) == 0 {
		t.Fatalf("expected stock watch signals, got %+v", watch)
	}
}

func TestBuildFuturesResearchBlocksCreatesDomainSpecificOutline(t *testing.T) {
	report := map[string]any{
		"market_regime": "TREND_CONTINUE",
		"memory_feedback": map[string]any{
			"summary":         "趋势方向正确，但节奏验证偏慢",
			"failure_signals": []any{"宏观扰动放大回撤"},
			"suggestions":     []any{"紧盯基差和主仓切换"},
		},
	}
	ctx := strategyEngineAssetContext{
		record: model.StrategyEnginePublishRecord{TradeDate: "2026-03-28"},
		asset: map[string]any{
			"contract":                "RB2609",
			"reason_summary":          "供需收紧叠加基差修复，方向继续偏多",
			"risk_summary":            "跌破 3220 则失效",
			"risk_flags":              []any{"主仓切换临近"},
			"invalidations":           []any{"跌破 3220 则失效"},
			"inventory_factor_summary": "库存连续去化",
			"structure_factor_summary": "期限结构保持同向",
		},
	}

	outline, active, historical, watch := buildFuturesResearchBlocks(ctx, report, nil, nil)
	if len(outline) != 5 {
		t.Fatalf("expected 5 futures research outline steps, got %+v", outline)
	}
	if len(active) == 0 {
		t.Fatalf("expected futures active thesis cards, got %+v", active)
	}
	if len(historical) == 0 {
		t.Fatalf("expected futures historical thesis cards, got %+v", historical)
	}
	if len(watch) == 0 {
		t.Fatalf("expected futures watch signals, got %+v", watch)
	}
}
