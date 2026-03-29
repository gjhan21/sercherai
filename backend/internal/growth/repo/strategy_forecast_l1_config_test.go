package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestApplyForecastL1DisplayConfigDisablesExplanationBlocks(t *testing.T) {
	explanation := model.StrategyClientExplanation{
		ResearchOutline: []model.StrategyResearchOutlineStep{
			{Slot: "TREND", Title: "趋势与结构"},
		},
		ActiveThesisCards: []model.StrategyExplanationThesisCard{
			{Key: "active", Title: "当前理由"},
		},
		HistoricalThesisCards: []model.StrategyExplanationThesisCard{
			{Key: "history", Title: "历史弱化理由"},
		},
		WatchSignals: []model.StrategyExplanationWatchSignal{
			{SignalType: "INVALIDATION", Trigger: "跌破 5 日线"},
		},
		MemoryFeedback: model.StrategyExplanationMemoryFeedback{
			Summary:        "历史上高波动题材容易验证过慢",
			Suggestions:    []string{"缩短验证周期"},
			FailureSignals: []string{"放量不涨"},
		},
		ConfidenceCalibration: model.StrategyExplanationConfidenceCalibration{
			BaseConfidence:     0.62,
			AdjustedConfidence: 0.48,
			AdvisoryOnly:       true,
		},
	}

	applyForecastL1DisplayConfigToExplanation(
		&explanation,
		map[string]any{"sample_count": 8},
		forecastL1RuntimeConfig{
			Enabled:                 false,
			ExplanationEnabled:      true,
			MemoryFeedbackMinSamples: 5,
			AdvisoryPriorityThreshold: 0.55,
		},
	)

	if len(explanation.ResearchOutline) != 0 || len(explanation.ActiveThesisCards) != 0 || len(explanation.WatchSignals) != 0 {
		t.Fatalf("expected l1 explanation blocks to be cleared when l1 is disabled, got %+v", explanation)
	}
	if explanation.MemoryFeedback.Summary != "" || explanation.ConfidenceCalibration.AdjustedConfidence != 0 {
		t.Fatalf("expected l1 memory and confidence blocks to be cleared when l1 is disabled, got %+v", explanation)
	}
}

func TestApplyForecastL1DisplayConfigDropsMemoryFeedbackBelowSampleThreshold(t *testing.T) {
	explanation := model.StrategyClientExplanation{
		MemoryFeedback: model.StrategyExplanationMemoryFeedback{
			Summary:        "历史上容易高位放量不涨",
			Suggestions:    []string{"缩短验证周期"},
			FailureSignals: []string{"高位放量不涨"},
		},
		ConfidenceReason: "当前趋势仍在",
		RiskBoundary:     "跌破关键位失效",
	}

	applyForecastL1DisplayConfigToExplanation(
		&explanation,
		map[string]any{"sample_count": 3},
		forecastL1RuntimeConfig{
			Enabled:                 true,
			ExplanationEnabled:      true,
			MemoryFeedbackMinSamples: 5,
			AdvisoryPriorityThreshold: 0.55,
		},
	)

	if explanation.MemoryFeedback.Summary != "" || len(explanation.MemoryFeedback.Suggestions) != 0 {
		t.Fatalf("expected memory feedback to be hidden when sample count is below threshold, got %+v", explanation.MemoryFeedback)
	}
}

func TestApplyForecastL1DisplayConfigMarksHighAdvisorySamplesByThreshold(t *testing.T) {
	explanation := model.StrategyClientExplanation{
		EvaluationMeta: map[string]any{"status": "COMPLETED"},
		ConfidenceCalibration: model.StrategyExplanationConfidenceCalibration{
			BaseConfidence:     0.62,
			AdjustedConfidence: 0.42,
			AdvisoryOnly:       true,
		},
	}

	applyForecastL1DisplayConfigToExplanation(
		&explanation,
		nil,
		forecastL1RuntimeConfig{
			Enabled:                 true,
			ExplanationEnabled:      true,
			MemoryFeedbackMinSamples: 5,
			AdvisoryPriorityThreshold: 0.55,
		},
	)

	if asString(explanation.EvaluationMeta["advisory_priority"]) != "HIGH" {
		t.Fatalf("expected advisory threshold to mark high-priority advisory samples in evaluation meta, got %+v", explanation.EvaluationMeta)
	}
}
