package repo

import (
	"database/sql"
	"strings"

	"sercherai/backend/internal/growth/model"
)

const (
	forecastL3EnabledConfigKey                 = "growth.forecast_l3.enabled"
	forecastL3AdminManualEnabledConfigKey      = "growth.forecast_l3.admin_manual_enabled"
	forecastL3UserRequestEnabledConfigKey      = "growth.forecast_l3.user_request_enabled"
	forecastL3AutoPriorityEnabledConfigKey     = "growth.forecast_l3.auto_priority_enabled"
	forecastL3ClientReadEnabledConfigKey       = "growth.forecast_l3.client_read_enabled"
	forecastL3RequireVIPForFullReportConfigKey = "growth.forecast_l3.require_vip_for_full_report"
	forecastL3MaxActiveRunsConfigKey           = "growth.forecast_l3.max_active_runs"
	forecastL3MaxRunsPerDayConfigKey           = "growth.forecast_l3.max_runs_per_day"
	forecastL3MaxUserRunsPerDayConfigKey       = "growth.forecast_l3.max_user_runs_per_day"
	forecastL3MinPriorityThresholdConfigKey    = "growth.forecast_l3.min_priority_threshold"
	forecastL3DispatchEnabledConfigKey         = "growth.forecast_l3.dispatch.enabled"
	forecastL3DispatchIntervalMinutesConfigKey = "growth.forecast_l3.dispatch.interval_minutes"
	forecastL3QualityEnabledConfigKey          = "growth.forecast_l3.quality.enabled"
	forecastL3QualityIntervalMinutesConfigKey  = "growth.forecast_l3.quality.interval_minutes"
	forecastL3DefaultEngineKeyConfigKey        = "growth.forecast_l3.default_engine_key"
)

type forecastL3RuntimeConfig struct {
	Enabled                 bool
	AdminManualEnabled      bool
	UserRequestEnabled      bool
	AutoPriorityEnabled     bool
	ClientReadEnabled       bool
	RequireVIPForFullReport bool
	MaxActiveRuns           int
	MaxRunsPerDay           int
	MaxUserRunsPerDay       int
	MinPriorityThreshold    float64
	DispatchEnabled         bool
	DispatchIntervalMinutes int
	QualityEnabled          bool
	QualityIntervalMinutes  int
	DefaultEngineKey        string
}

var defaultForecastL3RuntimeConfig = forecastL3RuntimeConfig{
	Enabled:                 false,
	AdminManualEnabled:      true,
	UserRequestEnabled:      false,
	AutoPriorityEnabled:     false,
	ClientReadEnabled:       true,
	RequireVIPForFullReport: true,
	MaxActiveRuns:           2,
	MaxRunsPerDay:           24,
	MaxUserRunsPerDay:       1,
	MinPriorityThreshold:    0.70,
	DispatchEnabled:         true,
	DispatchIntervalMinutes: 5,
	QualityEnabled:          true,
	QualityIntervalMinutes:  60,
	DefaultEngineKey:        model.StrategyForecastL3EngineLocalSynthesis,
}

func (r *MySQLGrowthRepo) loadForecastL3RuntimeConfig() forecastL3RuntimeConfig {
	if r == nil || r.db == nil {
		return defaultForecastL3RuntimeConfig
	}

	rows, err := r.db.Query(`
SELECT config_key, config_value
FROM system_configs
WHERE config_key LIKE 'growth.forecast_l3.%'`)
	if err != nil {
		return defaultForecastL3RuntimeConfig
	}
	defer rows.Close()

	values := make(map[string]string)
	for rows.Next() {
		var configKey string
		var configValue sql.NullString
		if err := rows.Scan(&configKey, &configValue); err != nil {
			return defaultForecastL3RuntimeConfig
		}
		values[strings.TrimSpace(configKey)] = strings.TrimSpace(configValue.String)
	}
	if err := rows.Err(); err != nil {
		return defaultForecastL3RuntimeConfig
	}
	return parseForecastL3RuntimeConfig(values)
}

func parseForecastL3RuntimeConfig(values map[string]string) forecastL3RuntimeConfig {
	config := defaultForecastL3RuntimeConfig
	config.Enabled = parseForecastConfigBool(values[forecastL3EnabledConfigKey], config.Enabled)
	config.AdminManualEnabled = parseForecastConfigBool(values[forecastL3AdminManualEnabledConfigKey], config.AdminManualEnabled)
	config.UserRequestEnabled = parseForecastConfigBool(values[forecastL3UserRequestEnabledConfigKey], config.UserRequestEnabled)
	config.AutoPriorityEnabled = parseForecastConfigBool(values[forecastL3AutoPriorityEnabledConfigKey], config.AutoPriorityEnabled)
	config.ClientReadEnabled = parseForecastConfigBool(values[forecastL3ClientReadEnabledConfigKey], config.ClientReadEnabled)
	config.RequireVIPForFullReport = parseForecastConfigBool(
		values[forecastL3RequireVIPForFullReportConfigKey],
		config.RequireVIPForFullReport,
	)
	config.MaxActiveRuns = parseForecastConfigInt(values[forecastL3MaxActiveRunsConfigKey], config.MaxActiveRuns, 1, 20)
	config.MaxRunsPerDay = parseForecastConfigInt(values[forecastL3MaxRunsPerDayConfigKey], config.MaxRunsPerDay, 1, 500)
	config.MaxUserRunsPerDay = parseForecastConfigInt(
		values[forecastL3MaxUserRunsPerDayConfigKey],
		config.MaxUserRunsPerDay,
		1,
		20,
	)
	config.MinPriorityThreshold = parseForecastConfigFloat(
		values[forecastL3MinPriorityThresholdConfigKey],
		config.MinPriorityThreshold,
		0.10,
		0.99,
	)
	config.DispatchEnabled = parseForecastConfigBool(values[forecastL3DispatchEnabledConfigKey], config.DispatchEnabled)
	config.DispatchIntervalMinutes = parseForecastConfigInt(
		values[forecastL3DispatchIntervalMinutesConfigKey],
		config.DispatchIntervalMinutes,
		1,
		1440,
	)
	config.QualityEnabled = parseForecastConfigBool(values[forecastL3QualityEnabledConfigKey], config.QualityEnabled)
	config.QualityIntervalMinutes = parseForecastConfigInt(
		values[forecastL3QualityIntervalMinutesConfigKey],
		config.QualityIntervalMinutes,
		1,
		1440,
	)
	config.DefaultEngineKey = parseForecastConfigString(
		values[forecastL3DefaultEngineKeyConfigKey],
		config.DefaultEngineKey,
	)
	return config
}

func parseForecastConfigString(raw string, fallback string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(raw))
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func (c forecastL3RuntimeConfig) TriggerEnabled(triggerType string) bool {
	if !c.Enabled {
		return false
	}
	switch strings.ToUpper(strings.TrimSpace(triggerType)) {
	case model.StrategyForecastL3TriggerTypeAdminManual:
		return c.AdminManualEnabled
	case model.StrategyForecastL3TriggerTypeAutoPriority:
		return c.AutoPriorityEnabled
	case model.StrategyForecastL3TriggerTypeUserRequest:
		return c.UserRequestEnabled
	default:
		return false
	}
}
