package repo

import "testing"

func TestBuildStockAgentOpinionsReturnsConsensusAndVeto(t *testing.T) {
	ctx := strategyEngineAssetContext{
		asset: map[string]any{
			"reason_summary": "趋势延续但高位分歧",
			"risk_flags":     []any{"高位分歧"},
			"invalidations":  []any{"跌破 5 日线"},
			"theme_tags":     []any{"白酒"},
		},
	}

	opinions, meta := buildStockAgentOpinions(ctx)
	if len(opinions) < 3 {
		t.Fatalf("expected multiple opinions, got %+v", opinions)
	}
	if meta.ConsensusAction == "" {
		t.Fatalf("expected consensus action, got %+v", meta)
	}
}
