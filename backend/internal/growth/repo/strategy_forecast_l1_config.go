package repo

import (
	"database/sql"
	"strings"

	"sercherai/backend/internal/growth/model"
)

const (
	forecastL1EnabledConfigKey                 = "growth.forecast_l1.enabled"
	forecastL1ExplanationEnabledConfigKey      = "growth.forecast_l1.explanation_enabled"
	forecastL1MemoryMinSamplesConfigKey        = "growth.forecast_l1.memory_feedback_min_samples"
	forecastL1AdvisoryPriorityThresholdConfigKey = "growth.forecast_l1.advisory_priority_threshold"
)

type forecastL1RuntimeConfig struct {
	Enabled                  bool
	ExplanationEnabled       bool
	MemoryFeedbackMinSamples int
	AdvisoryPriorityThreshold float64
}

var defaultForecastL1RuntimeConfig = forecastL1RuntimeConfig{
	Enabled:                  true,
	ExplanationEnabled:       true,
	MemoryFeedbackMinSamples: 5,
	AdvisoryPriorityThreshold: 0.55,
}

func (r *MySQLGrowthRepo) loadForecastL1RuntimeConfig() forecastL1RuntimeConfig {
	if r == nil || r.db == nil {
		return defaultForecastL1RuntimeConfig
	}

	rows, err := r.db.Query(`
SELECT config_key, config_value
FROM system_configs
WHERE config_key LIKE 'growth.forecast_l1.%'`)
	if err != nil {
		return defaultForecastL1RuntimeConfig
	}
	defer rows.Close()

	values := make(map[string]string)
	for rows.Next() {
		var configKey string
		var configValue sql.NullString
		if err := rows.Scan(&configKey, &configValue); err != nil {
			return defaultForecastL1RuntimeConfig
		}
		values[strings.TrimSpace(configKey)] = strings.TrimSpace(configValue.String)
	}
	if err := rows.Err(); err != nil {
		return defaultForecastL1RuntimeConfig
	}
	return parseForecastL1RuntimeConfig(values)
}

func parseForecastL1RuntimeConfig(values map[string]string) forecastL1RuntimeConfig {
	config := defaultForecastL1RuntimeConfig
	config.Enabled = parseForecastConfigBool(values[forecastL1EnabledConfigKey], config.Enabled)
	config.ExplanationEnabled = parseForecastConfigBool(values[forecastL1ExplanationEnabledConfigKey], config.ExplanationEnabled)
	config.MemoryFeedbackMinSamples = parseForecastConfigInt(
		values[forecastL1MemoryMinSamplesConfigKey],
		config.MemoryFeedbackMinSamples,
		1,
		100,
	)
	config.AdvisoryPriorityThreshold = parseForecastConfigFloat(
		values[forecastL1AdvisoryPriorityThresholdConfigKey],
		config.AdvisoryPriorityThreshold,
		0.1,
		0.95,
	)
	return config
}

func parseForecastConfigInt(raw string, fallback int, min int, max int) int {
	parsed := asInt(raw)
	if parsed <= 0 {
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

func applyForecastL1DisplayConfigToExplanation(
	explanation *model.StrategyClientExplanation,
	rawMemoryFeedback map[string]any,
	config forecastL1RuntimeConfig,
) {
	if explanation == nil {
		return
	}
	applyForecastL1DisplayConfigFields(
		&explanation.ResearchOutline,
		&explanation.ActiveThesisCards,
		&explanation.HistoricalThesisCards,
		&explanation.WatchSignals,
		&explanation.MemoryFeedback,
		&explanation.ConfidenceCalibration,
		&explanation.EvaluationMeta,
		rawMemoryFeedback,
		config,
	)
}

func applyForecastL1DisplayConfigToVersionHistoryItem(
	item *model.StrategyVersionHistoryItem,
	rawMemoryFeedback map[string]any,
	config forecastL1RuntimeConfig,
) {
	if item == nil {
		return
	}
	applyForecastL1DisplayConfigFields(
		&item.ResearchOutline,
		&item.ActiveThesisCards,
		&item.HistoricalThesisCards,
		&item.WatchSignals,
		&item.MemoryFeedback,
		&item.ConfidenceCalibration,
		&item.EvaluationMeta,
		rawMemoryFeedback,
		config,
	)
}

func applyForecastL1DisplayConfigFields(
	researchOutline *[]model.StrategyResearchOutlineStep,
	activeThesis *[]model.StrategyExplanationThesisCard,
	historicalThesis *[]model.StrategyExplanationThesisCard,
	watchSignals *[]model.StrategyExplanationWatchSignal,
	memoryFeedback *model.StrategyExplanationMemoryFeedback,
	confidenceCalibration *model.StrategyExplanationConfidenceCalibration,
	evaluationMeta *map[string]any,
	rawMemoryFeedback map[string]any,
	config forecastL1RuntimeConfig,
) {
	if researchOutline == nil || activeThesis == nil || historicalThesis == nil || watchSignals == nil || memoryFeedback == nil || confidenceCalibration == nil || evaluationMeta == nil {
		return
	}

	if !config.Enabled || !config.ExplanationEnabled {
		*researchOutline = nil
		*activeThesis = nil
		*historicalThesis = nil
		*watchSignals = nil
		*memoryFeedback = model.StrategyExplanationMemoryFeedback{}
		*confidenceCalibration = model.StrategyExplanationConfidenceCalibration{}
		return
	}

	if *evaluationMeta == nil {
		*evaluationMeta = map[string]any{}
	}

	if sampleCount, ok := forecastMemoryFeedbackSampleCount(rawMemoryFeedback); ok && sampleCount < config.MemoryFeedbackMinSamples {
		*memoryFeedback = model.StrategyExplanationMemoryFeedback{}
	}

	if confidenceCalibration.AdvisoryOnly && confidenceCalibration.AdjustedConfidence > 0 &&
		confidenceCalibration.AdjustedConfidence <= config.AdvisoryPriorityThreshold {
		(*evaluationMeta)["advisory_priority"] = "HIGH"
	}
}

func forecastMemoryFeedbackSampleCount(rawMemoryFeedback map[string]any) (int, bool) {
	if len(rawMemoryFeedback) == 0 {
		return 0, false
	}
	for _, key := range []string{"sample_count", "samples", "total_samples", "count"} {
		if count := asInt(rawMemoryFeedback[key]); count > 0 {
			return count, true
		}
	}
	meta := mapValue(rawMemoryFeedback["meta"])
	for _, key := range []string{"sample_count", "samples", "total_samples", "count"} {
		if count := asInt(meta[key]); count > 0 {
			return count, true
		}
	}
	return 0, false
}
