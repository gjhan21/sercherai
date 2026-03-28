package repo

import "sercherai/backend/internal/growth/model"

func buildStockAgentOpinions(ctx strategyEngineAssetContext) ([]model.StrategyExplanationAgentOpinion, model.StrategyExplanationScenarioMeta) {
	opinions := []model.StrategyExplanationAgentOpinion{
		{
			Role:       "FLOW",
			Agent:      "FLOW",
			Stance:     "SUPPORT",
			Confidence: 0.68,
			Summary:    firstNonEmpty(asString(ctx.asset["reason_summary"]), "量价与资金延续原方向"),
		},
		{
			Role:       "THEME",
			Agent:      "THEME",
			Stance:     stanceFromTags(stringSlice(ctx.asset["theme_tags"])),
			Confidence: 0.61,
			Summary:    firstNonEmpty(firstNonEmpty(stringSlice(ctx.asset["theme_tags"])...), "主题线索仍可跟踪"),
		},
		{
			Role:       "RISK",
			Agent:      "RISK",
			Stance:     riskStance(ctx),
			Confidence: 0.73,
			Summary:    firstNonEmpty(firstNonEmpty(stringSlice(ctx.asset["risk_flags"])...), firstNonEmpty(stringSlice(ctx.asset["invalidations"])...), "风险边界正常"),
			Veto:       len(stringSlice(ctx.asset["risk_flags"])) >= 3,
		},
	}
	meta := buildScenarioConsensusMeta(opinions)
	return opinions, meta
}

func buildFuturesAgentOpinions(ctx strategyEngineAssetContext) ([]model.StrategyExplanationAgentOpinion, model.StrategyExplanationScenarioMeta) {
	opinions := []model.StrategyExplanationAgentOpinion{
		{
			Role:       "DIRECTION",
			Agent:      "DIRECTION",
			Stance:     "SUPPORT",
			Confidence: 0.66,
			Summary:    firstNonEmpty(asString(ctx.asset["reason_summary"]), "方向逻辑仍然成立"),
		},
		{
			Role:       "SUPPLY",
			Agent:      "SUPPLY",
			Stance:     "WATCH",
			Confidence: 0.58,
			Summary:    firstNonEmpty(asString(ctx.asset["inventory_factor_summary"]), "供需与库存需要继续观察"),
		},
		{
			Role:       "RISK",
			Agent:      "RISK",
			Stance:     riskStance(ctx),
			Confidence: 0.71,
			Summary:    firstNonEmpty(firstNonEmpty(stringSlice(ctx.asset["risk_flags"])...), firstNonEmpty(stringSlice(ctx.asset["invalidations"])...), "风险边界正常"),
			Veto:       len(stringSlice(ctx.asset["risk_flags"])) >= 3,
		},
	}
	meta := buildScenarioConsensusMeta(opinions)
	return opinions, meta
}

func buildScenarioConsensusMeta(opinions []model.StrategyExplanationAgentOpinion) model.StrategyExplanationScenarioMeta {
	meta := model.StrategyExplanationScenarioMeta{
		ConsensusAction: "继续观察",
	}
	if len(opinions) == 0 {
		return meta
	}
	support := 0
	oppose := 0
	confidences := make([]float64, 0, len(opinions))
	for _, opinion := range opinions {
		confidences = append(confidences, opinion.Confidence)
		switch opinion.Stance {
		case "SUPPORT":
			support++
		case "OPPOSE":
			oppose++
		}
		if opinion.Veto {
			meta.Vetoed = true
			meta.VetoReason = firstNonEmpty(opinion.Summary, opinion.Role)
		}
	}
	if support >= oppose {
		meta.ConsensusAction = "按计划执行"
	} else {
		meta.ConsensusAction = "降低风险暴露"
	}
	meta.ScenarioConfidenceSpread = confidenceSpread(confidences)
	return meta
}

func confidenceSpread(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	high := values[0]
	low := values[0]
	for _, value := range values[1:] {
		if value > high {
			high = value
		}
		if value < low {
			low = value
		}
	}
	return roundTo(high-low, 2)
}

func stanceFromTags(tags []string) string {
	if len(compactStrings(tags)) == 0 {
		return "WATCH"
	}
	return "SUPPORT"
}

func riskStance(ctx strategyEngineAssetContext) string {
	if len(stringSlice(ctx.asset["risk_flags"])) > 0 || len(stringSlice(ctx.asset["invalidations"])) > 0 {
		return "WATCH"
	}
	return "SUPPORT"
}
