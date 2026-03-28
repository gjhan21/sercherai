package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestBuildExplanationConfidenceCalibrationIsAdvisoryOnly(t *testing.T) {
	explanation := model.StrategyClientExplanation{
		ConfidenceReason: "趋势延续",
		MemoryFeedback: model.StrategyExplanationMemoryFeedback{
			Summary: "高波动模板近几次回撤偏大",
		},
		EvaluationMeta: map[string]any{
			"status": "COMPLETED",
			"5": map[string]any{
				"return_pct":       0.02,
				"max_drawdown_pct": -0.11,
			},
		},
	}

	calibration := buildExplanationConfidenceCalibration(explanation)
	if !calibration.AdvisoryOnly {
		t.Fatalf("expected advisory_only=true")
	}
	if calibration.AdjustedConfidence >= calibration.BaseConfidence {
		t.Fatalf("expected drawdown-heavy case to reduce confidence: %+v", calibration)
	}
	if len(calibration.Drivers) == 0 {
		t.Fatalf("expected confidence drivers")
	}
}
