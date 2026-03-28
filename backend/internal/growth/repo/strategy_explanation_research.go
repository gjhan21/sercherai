package repo

import (
	"strings"

	"sercherai/backend/internal/growth/model"
)

func buildStockResearchBlocks(
	ctx strategyEngineAssetContext,
	report map[string]any,
	previous *strategyEngineAssetContext,
	evaluation map[string]any,
) (
	outline []model.StrategyResearchOutlineStep,
	active []model.StrategyExplanationThesisCard,
	historical []model.StrategyExplanationThesisCard,
	watch []model.StrategyExplanationWatchSignal,
) {
	memory := buildExplanationMemoryFeedback(mapValue(report["memory_feedback"]))
	reasonSummary := asString(ctx.asset["reason_summary"])
	riskSummary := firstNonEmpty(asString(ctx.asset["risk_summary"]), asString(report["risk_summary"]))
	themeText := strings.Join(compactStrings(append(stringSlice(ctx.asset["theme_tags"]), stringSlice(ctx.asset["sector_tags"])...)), " / ")

	outline = []model.StrategyResearchOutlineStep{
		{Slot: "TREND", Title: "趋势与价量结构", Summary: firstNonEmpty(reasonSummary, "当前查看价格结构与趋势延续。"), Status: "ACTIVE", EvidenceHint: "理由摘要"},
		{Slot: "FLOW", Title: "资金承接与流向", Summary: firstNonEmpty(themeText, "结合资金承接确认主线延续性。"), Status: "ACTIVE", EvidenceHint: "主题与板块"},
		{Slot: "FUNDAMENTAL", Title: "估值与基本面约束", Summary: firstNonEmpty(themeText, "检查估值与基本面约束是否支持当前结论。"), Status: "ACTIVE", EvidenceHint: "行业标签"},
		{Slot: "CATALYST", Title: "行业 / 主题 / 事件催化", Summary: firstNonEmpty(memory.Summary, "跟踪主题与事件催化是否继续强化。"), Status: "ACTIVE", EvidenceHint: "记忆反馈"},
		{Slot: "RISK", Title: "风险边界与失效条件", Summary: firstNonEmpty(riskSummary, "明确风险边界与失效条件。"), Status: "WATCH", EvidenceHint: "风险摘要"},
	}

	active = compactThesisCards([]model.StrategyExplanationThesisCard{
		{
			Key:            "stock-reason",
			Title:          "当前主理由",
			Summary:        firstNonEmpty(reasonSummary, "当前推荐继续沿用现有主逻辑。"),
			Status:         "ACTIVE",
			EvidenceSource: "reason_summary",
			Note:           firstNonEmpty(themeText, "当前理由仍在执行区间内。"),
		},
		{
			Key:            "stock-theme",
			Title:          "主题与板块承接",
			Summary:        firstNonEmpty(themeText, "主题与板块承接仍支持当前结论。"),
			Status:         "ACTIVE",
			EvidenceSource: "theme_tags",
			Note:           "用于确认当前主线没有明显转弱。",
		},
	})

	historical = buildHistoricalThesisCards(previous, memory, "过去有效但当前已弱化的股票理由")
	watch = buildWatchSignals(
		append(stringSlice(ctx.asset["invalidations"]), stringSlice(ctx.asset["risk_flags"])...),
		memory.Suggestions,
		riskSummary,
	)
	if len(watch) == 0 {
		watch = []model.StrategyExplanationWatchSignal{
			{
				Title:      "继续跟踪风险边界",
				SignalType: "WATCH",
				Trigger:    firstNonEmpty(riskSummary, "观察当前逻辑是否出现明显转弱"),
				Action:     "若出现失效信号，回到版本历史核对旧逻辑变化",
				Priority:   "MEDIUM",
			},
		}
	}
	_ = evaluation
	return outline, active, historical, watch
}

func buildFuturesResearchBlocks(
	ctx strategyEngineAssetContext,
	report map[string]any,
	previous *strategyEngineAssetContext,
	evaluation map[string]any,
) (
	outline []model.StrategyResearchOutlineStep,
	active []model.StrategyExplanationThesisCard,
	historical []model.StrategyExplanationThesisCard,
	watch []model.StrategyExplanationWatchSignal,
) {
	memory := buildExplanationMemoryFeedback(mapValue(report["memory_feedback"]))
	reasonSummary := asString(ctx.asset["reason_summary"])
	riskSummary := firstNonEmpty(asString(ctx.asset["risk_summary"]), asString(report["risk_summary"]))
	inventorySummary := asString(ctx.asset["inventory_factor_summary"])
	structureSummary := asString(ctx.asset["structure_factor_summary"])

	outline = []model.StrategyResearchOutlineStep{
		{Slot: "DIRECTION", Title: "方向与关键价位", Summary: firstNonEmpty(reasonSummary, "先确认方向和关键价位。"), Status: "ACTIVE", EvidenceHint: "方向理由"},
		{Slot: "SUPPLY", Title: "供需 / 库存 / 产业链线索", Summary: firstNonEmpty(inventorySummary, "结合库存和产业链线索确认供需方向。"), Status: "ACTIVE", EvidenceHint: "库存摘要"},
		{Slot: "STRUCTURE", Title: "基差 / 期限结构 / 价差线索", Summary: firstNonEmpty(structureSummary, "观察期限结构与价差是否继续同向。"), Status: "ACTIVE", EvidenceHint: "结构摘要"},
		{Slot: "MACRO", Title: "宏观 / 政策 / 事件扰动", Summary: firstNonEmpty(memory.Summary, "检查宏观和政策扰动是否改变节奏。"), Status: "WATCH", EvidenceHint: "记忆反馈"},
		{Slot: "RISK", Title: "风险边界与失效条件", Summary: firstNonEmpty(riskSummary, "明确期货执行边界与失效条件。"), Status: "WATCH", EvidenceHint: "风险摘要"},
	}

	active = compactThesisCards([]model.StrategyExplanationThesisCard{
		{
			Key:            "futures-reason",
			Title:          "当前方向理由",
			Summary:        firstNonEmpty(reasonSummary, "当前方向逻辑继续成立。"),
			Status:         "ACTIVE",
			EvidenceSource: "reason_summary",
			Note:           firstNonEmpty(structureSummary, inventorySummary),
		},
		{
			Key:            "futures-structure",
			Title:          "结构与库存线索",
			Summary:        firstNonEmpty(structureSummary, inventorySummary, "结构与库存线索仍支持当前观点。"),
			Status:         "ACTIVE",
			EvidenceSource: "structure_factor_summary",
			Note:           "用于确认方向与节奏没有显著背离。",
		},
	})

	historical = buildHistoricalThesisCards(previous, memory, "过去有效但当前已弱化的期货理由")
	watch = buildWatchSignals(
		append(stringSlice(ctx.asset["invalidations"]), stringSlice(ctx.asset["risk_flags"])...),
		memory.Suggestions,
		riskSummary,
	)
	if len(watch) == 0 {
		watch = []model.StrategyExplanationWatchSignal{
			{
				Title:      "继续跟踪关键价位",
				SignalType: "WATCH",
				Trigger:    firstNonEmpty(riskSummary, "观察关键价位是否失守"),
				Action:     "如出现失效信号，重新确认方向与执行价位",
				Priority:   "MEDIUM",
			},
		}
	}
	_ = evaluation
	return outline, active, historical, watch
}

func buildHistoricalThesisCards(
	previous *strategyEngineAssetContext,
	memory model.StrategyExplanationMemoryFeedback,
	fallbackTitle string,
) []model.StrategyExplanationThesisCard {
	cards := []model.StrategyExplanationThesisCard{}
	if previous != nil {
		prevReason := asString(previous.asset["reason_summary"])
		if prevReason != "" {
			cards = append(cards, model.StrategyExplanationThesisCard{
				Key:            "previous-reason",
				Title:          "上一版主理由",
				Summary:        prevReason,
				Status:         "HISTORICAL",
				EvidenceSource: "version_history",
				Note:           "当前版本需要和旧理由一起对照理解变化。",
			})
		}
	}
	for _, signal := range memory.FailureSignals {
		cards = append(cards, model.StrategyExplanationThesisCard{
			Key:            "memory-failure-" + signal,
			Title:          fallbackTitle,
			Summary:        signal,
			Status:         "HISTORICAL",
			EvidenceSource: "memory_feedback.failure_signals",
			Note:           firstNonEmpty(memory.Summary, "该逻辑过去曾成立，但当前需要谨慎处理。"),
		})
	}
	if len(cards) == 0 && memory.Summary != "" {
		cards = append(cards, model.StrategyExplanationThesisCard{
			Key:            "memory-summary",
			Title:          fallbackTitle,
			Summary:        memory.Summary,
			Status:         "HISTORICAL",
			EvidenceSource: "memory_feedback.summary",
			Note:           "作为历史复盘提示使用。",
		})
	}
	return compactThesisCards(cards)
}

func buildWatchSignals(baseSignals []string, suggestions []string, riskSummary string) []model.StrategyExplanationWatchSignal {
	result := []model.StrategyExplanationWatchSignal{}
	for _, item := range compactStrings(baseSignals) {
		result = append(result, model.StrategyExplanationWatchSignal{
			Title:      "失效条件",
			SignalType: "INVALIDATION",
			Trigger:    item,
			Action:     "命中后停止沿用旧结论，并回看版本变化",
			Priority:   "HIGH",
		})
	}
	for _, item := range compactStrings(suggestions) {
		result = append(result, model.StrategyExplanationWatchSignal{
			Title:      "观察信号",
			SignalType: "WATCH",
			Trigger:    item,
			Action:     "继续跟踪该信号，确认逻辑是否继续强化",
			Priority:   "MEDIUM",
		})
	}
	if len(result) == 0 && strings.TrimSpace(riskSummary) != "" {
		result = append(result, model.StrategyExplanationWatchSignal{
			Title:      "风险边界",
			SignalType: "WATCH",
			Trigger:    riskSummary,
			Action:     "如风险边界被触发，应暂停沿用当前结论",
			Priority:   "MEDIUM",
		})
	}
	return result
}

func compactThesisCards(items []model.StrategyExplanationThesisCard) []model.StrategyExplanationThesisCard {
	result := make([]model.StrategyExplanationThesisCard, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		key := firstNonEmpty(item.Key, item.Title, item.Summary)
		if key == "" || strings.TrimSpace(item.Summary) == "" {
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
