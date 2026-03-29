package repo

import "sercherai/backend/internal/growth/model"

func buildStockScenarioSnapshots(ctx strategyEngineAssetContext) ([]model.StrategyExplanationScenarioSnapshot, model.StrategyExplanationScenarioMeta) {
	scenarios := buildScenarioSnapshotsFromContext(
		firstNonEmpty(asString(ctx.asset["reason_summary"]), "主逻辑延续"),
		firstNonEmpty(asString(ctx.asset["risk_summary"]), "关注风险边界"),
		compactStrings(stringSlice(ctx.asset["invalidations"])),
	)
	return scenarios, buildScenarioMeta(scenarios)
}

func buildFuturesScenarioSnapshots(ctx strategyEngineAssetContext) ([]model.StrategyExplanationScenarioSnapshot, model.StrategyExplanationScenarioMeta) {
	scenarios := buildScenarioSnapshotsFromContext(
		firstNonEmpty(asString(ctx.asset["reason_summary"]), "方向逻辑延续"),
		firstNonEmpty(asString(ctx.asset["risk_summary"]), "关注波动与止损"),
		compactStrings(stringSlice(ctx.asset["invalidations"])),
	)
	return scenarios, buildScenarioMeta(scenarios)
}

func buildScenarioSnapshotsFromContext(thesis string, riskBoundary string, invalidations []string) []model.StrategyExplanationScenarioSnapshot {
	invalidation := firstNonEmpty(firstNonEmpty(invalidations...), riskBoundary, "失效条件待补充")
	return []model.StrategyExplanationScenarioSnapshot{
		{
			Scenario:           "bull",
			Thesis:             thesis,
			Trigger:            "趋势与证据继续强化",
			ConfirmationSignal: "量价/结构继续共振",
			InvalidationSignal: invalidation,
			ExpectedWindow:     "1-2 周",
			ActionSuggestion:   "顺势跟踪",
			Confidence:         0.72,
		},
		{
			Scenario:           "base",
			Thesis:             thesis,
			Trigger:            "核心逻辑维持不变",
			ConfirmationSignal: "未出现新的失效信号",
			InvalidationSignal: invalidation,
			ExpectedWindow:     "3-5 个交易日",
			ActionSuggestion:   "按计划执行",
			Confidence:         0.64,
		},
		{
			Scenario:           "bear",
			Thesis:             "风险边界被触发后，原路径弱化",
			Trigger:            invalidation,
			ConfirmationSignal: "风险信号持续放大",
			InvalidationSignal: "重新站回关键边界",
			ExpectedWindow:     "1-3 个交易日",
			ActionSuggestion:   "收缩风险暴露",
			Confidence:         0.38,
		},
	}
}

func buildScenarioMeta(items []model.StrategyExplanationScenarioSnapshot) model.StrategyExplanationScenarioMeta {
	if len(items) == 0 {
		return model.StrategyExplanationScenarioMeta{}
	}
	high := items[0].Confidence
	low := items[0].Confidence
	for _, item := range items[1:] {
		if item.Confidence > high {
			high = item.Confidence
		}
		if item.Confidence < low {
			low = item.Confidence
		}
	}
	primaryScenario := preferredPrimaryScenario(items)
	return model.StrategyExplanationScenarioMeta{
		PrimaryScenario:          primaryScenario,
		ConsensusAction:          items[1].ActionSuggestion,
		Vetoed:                   false,
		VetoReason:               "",
		ScenarioConfidenceSpread: high - low,
	}
}

func preferredPrimaryScenario(items []model.StrategyExplanationScenarioSnapshot) string {
	for _, item := range items {
		if item.Scenario == "base" {
			return item.Scenario
		}
	}
	if len(items) == 0 {
		return ""
	}
	return items[0].Scenario
}

func mergeScenarioMeta(snapshotMeta model.StrategyExplanationScenarioMeta, agentMeta model.StrategyExplanationScenarioMeta) model.StrategyExplanationScenarioMeta {
	meta := snapshotMeta
	if meta.PrimaryScenario == "" {
		meta.PrimaryScenario = agentMeta.PrimaryScenario
	}
	if agentMeta.ConsensusAction != "" {
		meta.ConsensusAction = agentMeta.ConsensusAction
	}
	if agentMeta.Vetoed {
		meta.Vetoed = true
	}
	if agentMeta.VetoReason != "" {
		meta.VetoReason = agentMeta.VetoReason
	}
	if meta.ScenarioConfidenceSpread == 0 {
		meta.ScenarioConfidenceSpread = agentMeta.ScenarioConfidenceSpread
	}
	return meta
}
