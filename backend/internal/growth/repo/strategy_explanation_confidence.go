package repo

import (
	"fmt"
	"strings"

	"sercherai/backend/internal/growth/model"
)

func buildExplanationConfidenceCalibration(explanation model.StrategyClientExplanation) model.StrategyExplanationConfidenceCalibration {
	base := deriveBaseConfidenceScore(explanation)
	drivers, adjusted := buildConfidenceDrivers(explanation, base)
	return model.StrategyExplanationConfidenceCalibration{
		BaseConfidence:     base,
		AdjustedConfidence: adjusted,
		Drivers:            drivers,
		AdvisoryOnly:       true,
	}
}

func deriveBaseConfidenceScore(explanation model.StrategyClientExplanation) float64 {
	base := 0.52
	if len(explanation.ActiveThesisCards) > 0 {
		base += 0.06
	}
	if len(explanation.EvidenceCards) >= 2 {
		base += 0.06
	} else if len(explanation.EvidenceCards) == 1 {
		base += 0.03
	}
	if len(explanation.RelatedEvents) > 0 {
		base += 0.04
	}
	if len(explanation.Invalidations) > 0 {
		base -= 0.03
	}
	return roundTo(clampFloat(base, 0.35, 0.88), 2)
}

func buildConfidenceDrivers(
	explanation model.StrategyClientExplanation,
	base float64,
) ([]model.StrategyExplanationConfidenceDriver, float64) {
	drivers := make([]model.StrategyExplanationConfidenceDriver, 0, 6)
	adjusted := base
	appendDriver := func(label string, impact float64, note string, sourceKey string) {
		if impact == 0 {
			return
		}
		adjusted += impact
		drivers = append(drivers, model.StrategyExplanationConfidenceDriver{
			Label:     label,
			Impact:    roundTo(impact, 2),
			Note:      note,
			SourceKey: sourceKey,
		})
	}

	if len(explanation.RelatedEvents) > 0 {
		appendDriver("事件佐证", 0.03, "已存在审核通过的关联事件", "related_events")
	}
	if summary := strings.TrimSpace(explanation.MemoryFeedback.Summary); summary != "" {
		appendDriver("历史记忆反馈", -0.05, summary, "memory_feedback")
	}
	horizon5 := mapValue(explanation.EvaluationMeta["5"])
	if drawdown := asFloat(horizon5["max_drawdown_pct"]); drawdown <= -0.08 {
		appendDriver("回撤校准", -0.08, fmt.Sprintf("5日最大回撤 %.1f%%", drawdown*100), "evaluation_meta.5.max_drawdown_pct")
	}
	if returnPct := asFloat(horizon5["return_pct"]); returnPct < 0 {
		appendDriver("收益校准", -0.04, fmt.Sprintf("5日收益 %.1f%%", returnPct*100), "evaluation_meta.5.return_pct")
	}
	if len(explanation.RiskFlags) >= 3 {
		appendDriver("风险提示密集", -0.03, "风险提示较多", "risk_flags")
	}
	if len(explanation.Invalidations) == 0 && len(explanation.ActiveThesisCards) > 0 {
		appendDriver("理由集中", 0.02, "当前有效理由较清晰", "active_thesis_cards")
	}

	return drivers, roundTo(clampFloat(adjusted, 0.2, 0.95), 2)
}

func applyConfidenceCalibrationToExplanation(explanation *model.StrategyClientExplanation) {
	if explanation == nil {
		return
	}
	calibration := buildExplanationConfidenceCalibration(*explanation)
	explanation.ConfidenceCalibration = calibration

	if calibration.AdjustedConfidence < calibration.BaseConfidence {
		note := fmt.Sprintf("校准后置信度 %.2f", calibration.AdjustedConfidence)
		explanation.ConfidenceReason = strings.Trim(strings.Join(compactStrings([]string{
			explanation.ConfidenceReason,
			note,
		}), "；"), "；")
		explanation.RiskBoundary = strings.Trim(strings.Join(compactStrings([]string{
			explanation.RiskBoundary,
			"按校准后置信度降低节奏执行",
		}), "；"), "；")
		explanation.RiskFlags = compactStrings(append(explanation.RiskFlags, "置信度经 advisory 校准下调"))
		explanation.WatchSignals = compactExplanationWatchSignals(append(explanation.WatchSignals, model.StrategyExplanationWatchSignal{
			Title:      "置信度校准",
			SignalType: "CONFIDENCE_CALIBRATION",
			Trigger:    note,
			Action:     "降低预期并缩短观察窗口",
			Priority:   "HIGH",
		}))
		return
	}
	explanation.ConfidenceReason = strings.Trim(strings.Join(compactStrings([]string{
		explanation.ConfidenceReason,
		fmt.Sprintf("校准后置信度 %.2f", calibration.AdjustedConfidence),
	}), "；"), "；")
}
