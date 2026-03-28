package repo

import (
	"strings"

	"sercherai/backend/internal/growth/model"
)

func buildStockResearchBlocks(
	ctx strategyEngineAssetContext,
	assetKey string,
) ([]model.StrategyResearchOutlineStep, []model.StrategyExplanationThesisCard, []model.StrategyExplanationThesisCard, []model.StrategyExplanationWatchSignal) {
	report := ctx.record.ReportSnapshot
	reasonSummary := strings.TrimSpace(asString(ctx.asset["reason_summary"]))
	if reasonSummary == "" {
		reasonSummary = strings.TrimSpace(asString(report["reason_summary"]))
	}
	marketRegime := strings.TrimSpace(asString(report["market_regime"]))
	riskSummary := strings.TrimSpace(firstNonEmpty(asString(ctx.asset["risk_summary"]), asString(report["risk_summary"])))
	memory := mapValue(report["memory_feedback"])
	failureSignals := compactStrings(stringSlice(memory["failure_signals"]))

	outline := []model.StrategyResearchOutlineStep{
		{
			Slot:    "TREND",
			Title:   "趋势与结构",
			Summary: firstNonEmpty(reasonSummary, "趋势与结构仍需验证"),
			Status:  "ACTIVE",
		},
	}
	if marketRegime != "" {
		outline = append(outline, model.StrategyResearchOutlineStep{
			Slot:    "REGIME",
			Title:   "市场状态",
			Summary: marketRegime,
			Status:  "MONITOR",
		})
	}

	active := []model.StrategyExplanationThesisCard{}
	if reasonSummary != "" {
		active = append(active, model.StrategyExplanationThesisCard{
			Key:            "primary_reason",
			Title:          "当前理由",
			Summary:        reasonSummary,
			Status:         "ACTIVE",
			EvidenceSource: firstNonEmpty(marketRegime, "REPORT"),
		})
	}

	historical := []model.StrategyExplanationThesisCard{}
	if len(failureSignals) > 0 {
		historical = append(historical, model.StrategyExplanationThesisCard{
			Key:     "memory_feedback",
			Title:   "历史弱化理由",
			Summary: failureSignals[0],
			Status:  "WEAKENED",
			Note:    strings.TrimSpace(asString(memory["summary"])),
		})
	}

	watch := []model.StrategyExplanationWatchSignal{}
	trigger := riskSummary
	if trigger == "" && len(failureSignals) > 0 {
		trigger = failureSignals[0]
	}
	if trigger != "" {
		watch = append(watch, model.StrategyExplanationWatchSignal{
			Title:      "风险触发",
			SignalType: "INVALIDATION",
			Trigger:    trigger,
			Action:     "降低暴露",
			Priority:   "HIGH",
		})
	}

	return outline, active, historical, watch
}

func buildFuturesResearchBlocks(
	ctx strategyEngineAssetContext,
	assetKey string,
) ([]model.StrategyResearchOutlineStep, []model.StrategyExplanationThesisCard, []model.StrategyExplanationThesisCard, []model.StrategyExplanationWatchSignal) {
	report := ctx.record.ReportSnapshot
	reasonSummary := strings.TrimSpace(asString(ctx.asset["reason_summary"]))
	structureSummary := strings.TrimSpace(asString(ctx.asset["structure_factor_summary"]))
	inventorySummary := strings.TrimSpace(asString(ctx.asset["inventory_factor_summary"]))
	marketRegime := strings.TrimSpace(asString(report["market_regime"]))
	memory := mapValue(report["memory_feedback"])
	failureSignals := compactStrings(stringSlice(memory["failure_signals"]))

	outline := []model.StrategyResearchOutlineStep{
		{
			Slot:    "STRUCTURE",
			Title:   "结构与供需",
			Summary: firstNonEmpty(structureSummary, inventorySummary, reasonSummary, "结构信号待确认"),
			Status:  "ACTIVE",
		},
	}
	if marketRegime != "" {
		outline = append(outline, model.StrategyResearchOutlineStep{
			Slot:    "REGIME",
			Title:   "市场状态",
			Summary: marketRegime,
			Status:  "MONITOR",
		})
	}

	active := []model.StrategyExplanationThesisCard{}
	if structureSummary != "" {
		active = append(active, model.StrategyExplanationThesisCard{
			Key:            "structure",
			Title:          "结构联动",
			Summary:        structureSummary,
			Status:         "ACTIVE",
			EvidenceSource: "STRUCTURE",
		})
	}
	if inventorySummary != "" {
		active = append(active, model.StrategyExplanationThesisCard{
			Key:            "inventory",
			Title:          "库存画像",
			Summary:        inventorySummary,
			Status:         "ACTIVE",
			EvidenceSource: "INVENTORY",
		})
	}
	if len(active) == 0 && reasonSummary != "" {
		active = append(active, model.StrategyExplanationThesisCard{
			Key:            "reason",
			Title:          "当前理由",
			Summary:        reasonSummary,
			Status:         "ACTIVE",
			EvidenceSource: "REASON",
		})
	}

	historical := []model.StrategyExplanationThesisCard{}
	if len(failureSignals) > 0 {
		historical = append(historical, model.StrategyExplanationThesisCard{
			Key:     "memory_feedback",
			Title:   "历史弱化理由",
			Summary: failureSignals[0],
			Status:  "WEAKENED",
			Note:    strings.TrimSpace(asString(memory["summary"])),
		})
	}

	watch := []model.StrategyExplanationWatchSignal{}
	trigger := strings.TrimSpace(firstNonEmpty(structureSummary, inventorySummary, reasonSummary))
	if trigger == "" {
		trigger = "关注结构与库存变化"
	}
	watch = append(watch, model.StrategyExplanationWatchSignal{
		Title:      "结构关注",
		SignalType: "INVALIDATION",
		Trigger:    trigger,
		Action:     "缩短验证周期",
		Priority:   "NORMAL",
	})

	return outline, active, historical, watch
}
