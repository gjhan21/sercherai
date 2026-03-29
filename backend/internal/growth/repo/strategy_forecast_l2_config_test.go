package repo

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestApplyForecastL2DisplayConfigDisablesL2SummaryBlocks(t *testing.T) {
	relationship := model.StrategyExplanationRelationshipSnapshot{
		AssetKey:          "600519.SH",
		AssetType:         "stock",
		RelationshipCount: 3,
		Nodes: []model.StrategyExplanationRelationshipNode{
			{Type: "Company", Key: "600519.SH", Label: "贵州茅台"},
		},
	}
	scenarioSnapshots := []model.StrategyExplanationScenarioSnapshot{
		{Scenario: "bull", ActionSuggestion: "顺势跟踪"},
		{Scenario: "base", ActionSuggestion: "按计划执行"},
	}
	scenarioMeta := model.StrategyExplanationScenarioMeta{
		PrimaryScenario:          "base",
		ConsensusAction:          "按计划执行",
		Vetoed:                   true,
		VetoReason:               "风险边界触发",
		ScenarioConfidenceSpread: 0.33,
	}
	agentOpinions := []model.StrategyExplanationAgentOpinion{
		{Role: "RISK", Agent: "RISK", Confidence: 0.72, Veto: true, Summary: "风险边界触发"},
	}

	applyForecastL2DisplayConfigFields(
		agentOpinions,
		&relationship,
		&scenarioSnapshots,
		&scenarioMeta,
		forecastL2RuntimeConfig{
			Enabled:                    false,
			RelationshipSnapshotEnabled: true,
			StableScenariosEnabled:    true,
			VetoConfidenceThreshold:   0.35,
		},
	)

	if relationship.RelationshipCount != 0 || len(relationship.Nodes) != 0 {
		t.Fatalf("expected l2 relationship snapshot to be cleared when l2 is disabled, got %+v", relationship)
	}
	if len(scenarioSnapshots) != 0 {
		t.Fatalf("expected stable scenarios to be cleared when l2 is disabled, got %+v", scenarioSnapshots)
	}
	if scenarioMeta.PrimaryScenario != "" || scenarioMeta.ConsensusAction != "" || scenarioMeta.Vetoed {
		t.Fatalf("expected l2 scenario meta to be cleared when l2 is disabled, got %+v", scenarioMeta)
	}
}

func TestApplyForecastL2DisplayConfigHonorsVetoThresholdAndScenarioToggle(t *testing.T) {
	relationship := model.StrategyExplanationRelationshipSnapshot{
		AssetKey:          "RB2609",
		AssetType:         "futures",
		RelationshipCount: 4,
		Nodes: []model.StrategyExplanationRelationshipNode{
			{Type: "Contract", Key: "RB2609", Label: "螺纹钢主力"},
		},
	}
	scenarioSnapshots := []model.StrategyExplanationScenarioSnapshot{
		{Scenario: "bull", ActionSuggestion: "顺势跟踪"},
		{Scenario: "base", ActionSuggestion: "按计划执行"},
	}
	scenarioMeta := model.StrategyExplanationScenarioMeta{
		PrimaryScenario:          "base",
		ConsensusAction:          "按计划执行",
		Vetoed:                   true,
		VetoReason:               "风险代理提示回撤扩大",
		ScenarioConfidenceSpread: 0.28,
	}
	agentOpinions := []model.StrategyExplanationAgentOpinion{
		{Role: "RISK", Agent: "RISK", Confidence: 0.22, Veto: true, Summary: "风险代理提示回撤扩大"},
		{Role: "SUPPLY", Agent: "SUPPLY", Confidence: 0.61, Veto: false, Summary: "供需仍可跟踪"},
	}

	applyForecastL2DisplayConfigFields(
		agentOpinions,
		&relationship,
		&scenarioSnapshots,
		&scenarioMeta,
		forecastL2RuntimeConfig{
			Enabled:                    true,
			RelationshipSnapshotEnabled: false,
			StableScenariosEnabled:    false,
			VetoConfidenceThreshold:   0.35,
		},
	)

	if relationship.RelationshipCount != 0 || len(relationship.Nodes) != 0 {
		t.Fatalf("expected relationship snapshot to respect toggle, got %+v", relationship)
	}
	if len(scenarioSnapshots) != 0 {
		t.Fatalf("expected scenario snapshots to respect toggle, got %+v", scenarioSnapshots)
	}
	if scenarioMeta.PrimaryScenario != "" || scenarioMeta.ScenarioConfidenceSpread != 0 {
		t.Fatalf("expected hidden stable scenarios to clear primary scenario summary, got %+v", scenarioMeta)
	}
	if scenarioMeta.Vetoed || scenarioMeta.VetoReason != "" {
		t.Fatalf("expected low-confidence veto to be downgraded from summary meta, got %+v", scenarioMeta)
	}
	if scenarioMeta.ConsensusAction == "" {
		t.Fatalf("expected consensus action to remain available when only stable scenarios are hidden, got %+v", scenarioMeta)
	}
}
