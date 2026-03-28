package repo

import (
	"strings"
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestApplyL1AdvisoryMemoryAdjustmentsAppendsRiskBoundaryAndWatchSignal(t *testing.T) {
	explanation := model.StrategyClientExplanation{
		ConfidenceReason: "当前趋势与资金共振",
		RiskBoundary:     "跌破 5 日线失效",
		MemoryFeedback: model.StrategyExplanationMemoryFeedback{
			Summary:        "过去同类题材高波动、验证慢",
			Suggestions:    []string{"缩短验证窗口"},
			FailureSignals: []string{"高位放量不涨"},
		},
	}
	evaluation := map[string]any{
		"status": "COMPLETED",
		"5": map[string]any{"return_pct": 0.03, "max_drawdown_pct": -0.09},
	}

	applyL1AdvisoryMemoryAdjustments(&explanation, evaluation)
	if !strings.Contains(explanation.RiskBoundary, "缩短验证窗口") {
		t.Fatalf("expected memory feedback to update risk boundary: %+v", explanation)
	}
	if len(explanation.WatchSignals) == 0 {
		t.Fatalf("expected memory feedback to create watch signals: %+v", explanation)
	}
	if len(explanation.RiskFlags) == 0 {
		t.Fatalf("expected memory feedback to append risk flags: %+v", explanation)
	}
}
