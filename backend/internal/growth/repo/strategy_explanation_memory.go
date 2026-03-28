package repo

import (
	"fmt"
	"strings"

	"sercherai/backend/internal/growth/model"
)

func normalizeExplanationEvaluationSummary(summary map[string]any) map[string]any {
	if len(summary) == 0 {
		return map[string]any{"status": "PENDING"}
	}
	result := make(map[string]any, len(summary)+1)
	for key, value := range summary {
		result[key] = value
	}
	if strings.TrimSpace(asString(result["status"])) == "" {
		result["status"] = "COMPLETED"
	}
	return result
}

func buildMemoryFailureSignals(feedback model.StrategyExplanationMemoryFeedback, evaluation map[string]any) []string {
	signals := append([]string{}, feedback.FailureSignals...)
	horizon5 := mapValue(evaluation["5"])
	if drawdown := asFloat(horizon5["max_drawdown_pct"]); drawdown <= -0.08 {
		signals = append(signals, fmt.Sprintf("5日最大回撤偏大(%.1f%%)", drawdown*100))
	}
	if returnPct := asFloat(horizon5["return_pct"]); returnPct < 0 {
		signals = append(signals, fmt.Sprintf("5日收益转负(%.1f%%)", returnPct*100))
	}
	return compactStrings(signals)
}

func buildMemorySuggestionSignals(feedback model.StrategyExplanationMemoryFeedback) []model.StrategyExplanationWatchSignal {
	result := make([]model.StrategyExplanationWatchSignal, 0, len(feedback.Suggestions))
	for _, suggestion := range compactStrings(feedback.Suggestions) {
		result = append(result, model.StrategyExplanationWatchSignal{
			Title:      "记忆反馈建议",
			SignalType: "MEMORY_ADVISORY",
			Trigger:    suggestion,
			Action:     "缩短验证周期",
			Priority:   "NORMAL",
		})
	}
	return result
}

func applyL1AdvisoryMemoryAdjustments(explanation *model.StrategyClientExplanation, evaluation map[string]any) {
	if explanation == nil {
		return
	}
	evaluation = normalizeExplanationEvaluationSummary(evaluation)
	explanation.EvaluationMeta = evaluation

	failureSignals := buildMemoryFailureSignals(explanation.MemoryFeedback, evaluation)
	suggestionSignals := buildMemorySuggestionSignals(explanation.MemoryFeedback)

	if len(failureSignals) > 0 {
		for _, signal := range failureSignals {
			explanation.WatchSignals = append(explanation.WatchSignals, model.StrategyExplanationWatchSignal{
				Title:      "历史失效信号",
				SignalType: "INVALIDATION",
				Trigger:    signal,
				Action:     "降低暴露",
				Priority:   "HIGH",
			})
		}
		explanation.RiskFlags = compactStrings(append(explanation.RiskFlags, failureSignals...))
	}
	if len(suggestionSignals) > 0 {
		explanation.WatchSignals = append(explanation.WatchSignals, suggestionSignals...)
	}

	boundaryParts := compactStrings([]string{
		explanation.RiskBoundary,
		firstNonEmpty(strings.Join(compactStrings(explanation.MemoryFeedback.Suggestions), "；"), ""),
	})
	if len(boundaryParts) > 0 {
		explanation.RiskBoundary = strings.Join(boundaryParts, "；")
	}
	if explanation.MemoryFeedback.Summary != "" && !strings.Contains(explanation.ConfidenceReason, explanation.MemoryFeedback.Summary) {
		explanation.ConfidenceReason = strings.Trim(strings.Join(compactStrings([]string{
			explanation.ConfidenceReason,
			"记忆反馈：" + explanation.MemoryFeedback.Summary,
		}), "；"), "；")
	}
	explanation.WatchSignals = compactExplanationWatchSignals(explanation.WatchSignals)
}

func compactExplanationWatchSignals(items []model.StrategyExplanationWatchSignal) []model.StrategyExplanationWatchSignal {
	if len(items) == 0 {
		return nil
	}
	result := make([]model.StrategyExplanationWatchSignal, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		key := strings.TrimSpace(item.SignalType) + "|" + strings.TrimSpace(item.Trigger)
		if key == "|" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}
