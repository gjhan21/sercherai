package repo

import (
	"database/sql"
	"strconv"
	"strings"

	"sercherai/backend/internal/growth/model"
)

const (
	forecastL2EnabledConfigKey                = "growth.forecast_l2.enabled"
	forecastL2RelationshipEnabledConfigKey    = "growth.forecast_l2.relationship_snapshot_enabled"
	forecastL2StableScenariosEnabledConfigKey = "growth.forecast_l2.stable_scenarios_enabled"
	forecastL2VetoThresholdConfigKey          = "growth.forecast_l2.veto_confidence_threshold"
)

type forecastL2RuntimeConfig struct {
	Enabled                     bool
	RelationshipSnapshotEnabled bool
	StableScenariosEnabled      bool
	VetoConfidenceThreshold     float64
}

var defaultForecastL2RuntimeConfig = forecastL2RuntimeConfig{
	Enabled:                     true,
	RelationshipSnapshotEnabled: true,
	StableScenariosEnabled:      true,
	VetoConfidenceThreshold:     0.35,
}

func (r *MySQLGrowthRepo) loadForecastL2RuntimeConfig() forecastL2RuntimeConfig {
	if r == nil || r.db == nil {
		return defaultForecastL2RuntimeConfig
	}

	rows, err := r.db.Query(`
SELECT config_key, config_value
FROM system_configs
WHERE config_key LIKE 'growth.forecast_l2.%'`)
	if err != nil {
		return defaultForecastL2RuntimeConfig
	}
	defer rows.Close()

	values := make(map[string]string)
	for rows.Next() {
		var configKey string
		var configValue sql.NullString
		if err := rows.Scan(&configKey, &configValue); err != nil {
			return defaultForecastL2RuntimeConfig
		}
		values[strings.TrimSpace(configKey)] = strings.TrimSpace(configValue.String)
	}
	if err := rows.Err(); err != nil {
		return defaultForecastL2RuntimeConfig
	}
	return parseForecastL2RuntimeConfig(values)
}

func parseForecastL2RuntimeConfig(values map[string]string) forecastL2RuntimeConfig {
	config := defaultForecastL2RuntimeConfig
	config.Enabled = parseForecastConfigBool(values[forecastL2EnabledConfigKey], config.Enabled)
	config.RelationshipSnapshotEnabled = parseForecastConfigBool(
		values[forecastL2RelationshipEnabledConfigKey],
		config.RelationshipSnapshotEnabled,
	)
	config.StableScenariosEnabled = parseForecastConfigBool(
		values[forecastL2StableScenariosEnabledConfigKey],
		config.StableScenariosEnabled,
	)
	config.VetoConfidenceThreshold = parseForecastConfigFloat(
		values[forecastL2VetoThresholdConfigKey],
		config.VetoConfidenceThreshold,
		0.05,
		0.95,
	)
	return config
}

func parseForecastConfigBool(raw string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return fallback
	case "1", "true", "yes", "on", "y":
		return true
	case "0", "false", "no", "off", "n":
		return false
	default:
		return fallback
	}
}

func parseForecastConfigFloat(raw string, fallback float64, min float64, max float64) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return fallback
	}
	if parsed < min {
		return min
	}
	if parsed > max {
		return max
	}
	return parsed
}

func applyForecastL2DisplayConfigToExplanation(
	explanation *model.StrategyClientExplanation,
	config forecastL2RuntimeConfig,
) {
	if explanation == nil {
		return
	}
	applyForecastL2DisplayConfigFields(
		explanation.AgentOpinions,
		&explanation.RelationshipSnapshot,
		&explanation.ScenarioSnapshots,
		&explanation.ScenarioMeta,
		config,
	)
}

func applyForecastL2DisplayConfigToVersionHistoryItem(
	item *model.StrategyVersionHistoryItem,
	config forecastL2RuntimeConfig,
) {
	if item == nil {
		return
	}
	applyForecastL2DisplayConfigFields(
		item.AgentOpinions,
		&item.RelationshipSnapshot,
		&item.ScenarioSnapshots,
		&item.ScenarioMeta,
		config,
	)
}

func applyForecastL2DisplayConfigFields(
	agentOpinions []model.StrategyExplanationAgentOpinion,
	relationshipSnapshot *model.StrategyExplanationRelationshipSnapshot,
	scenarioSnapshots *[]model.StrategyExplanationScenarioSnapshot,
	scenarioMeta *model.StrategyExplanationScenarioMeta,
	config forecastL2RuntimeConfig,
) {
	if relationshipSnapshot == nil || scenarioSnapshots == nil || scenarioMeta == nil {
		return
	}
	if !config.Enabled {
		*relationshipSnapshot = model.StrategyExplanationRelationshipSnapshot{}
		*scenarioSnapshots = nil
		*scenarioMeta = model.StrategyExplanationScenarioMeta{}
		return
	}
	if !config.RelationshipSnapshotEnabled {
		*relationshipSnapshot = model.StrategyExplanationRelationshipSnapshot{}
	}
	if !config.StableScenariosEnabled {
		*scenarioSnapshots = nil
		scenarioMeta.PrimaryScenario = ""
		scenarioMeta.ScenarioConfidenceSpread = 0
	}
	if scenarioMeta.Vetoed && highestVetoConfidence(agentOpinions) < config.VetoConfidenceThreshold {
		scenarioMeta.Vetoed = false
		scenarioMeta.VetoReason = ""
	}
}

func highestVetoConfidence(agentOpinions []model.StrategyExplanationAgentOpinion) float64 {
	high := 0.0
	for _, opinion := range agentOpinions {
		if !opinion.Veto {
			continue
		}
		if opinion.Confidence > high {
			high = opinion.Confidence
		}
	}
	return high
}
