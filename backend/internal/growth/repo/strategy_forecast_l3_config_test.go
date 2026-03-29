package repo

import "testing"

func TestParseForecastL3RuntimeConfig(t *testing.T) {
	config := parseForecastL3RuntimeConfig(map[string]string{
		"growth.forecast_l3.enabled":                     "true",
		"growth.forecast_l3.admin_manual_enabled":        "true",
		"growth.forecast_l3.user_request_enabled":        "false",
		"growth.forecast_l3.auto_priority_enabled":       "true",
		"growth.forecast_l3.client_read_enabled":         "true",
		"growth.forecast_l3.require_vip_for_full_report": "true",
		"growth.forecast_l3.max_active_runs":             "6",
		"growth.forecast_l3.max_runs_per_day":            "48",
		"growth.forecast_l3.max_user_runs_per_day":       "2",
		"growth.forecast_l3.min_priority_threshold":      "0.70",
		"growth.forecast_l3.dispatch.enabled":            "true",
		"growth.forecast_l3.dispatch.interval_minutes":   "8",
		"growth.forecast_l3.quality.enabled":             "true",
		"growth.forecast_l3.quality.interval_minutes":    "30",
		"growth.forecast_l3.default_engine_key":          "LOCAL_SYNTHESIS",
	})

	if !config.Enabled || !config.AdminManualEnabled || !config.AutoPriorityEnabled {
		t.Fatalf("expected l3 runtime switches to be enabled, got %+v", config)
	}
	if config.UserRequestEnabled {
		t.Fatalf("expected user-request switch to stay disabled when configured, got %+v", config)
	}
	if !config.ClientReadEnabled || !config.RequireVIPForFullReport {
		t.Fatalf("expected read-side switches to be parsed, got %+v", config)
	}
	if config.MaxActiveRuns != 6 || config.MaxRunsPerDay != 48 || config.MaxUserRunsPerDay != 2 {
		t.Fatalf("expected l3 runtime thresholds to be parsed, got %+v", config)
	}
	if config.MinPriorityThreshold != 0.70 {
		t.Fatalf("expected l3 priority threshold to be parsed, got %+v", config)
	}
	if !config.DispatchEnabled || config.DispatchIntervalMinutes != 8 {
		t.Fatalf("expected dispatch worker config to be parsed, got %+v", config)
	}
	if !config.QualityEnabled || config.QualityIntervalMinutes != 30 {
		t.Fatalf("expected quality worker config to be parsed, got %+v", config)
	}
	if config.DefaultEngineKey != "LOCAL_SYNTHESIS" {
		t.Fatalf("expected default engine key to be parsed, got %+v", config)
	}
}

func TestParseForecastL3RuntimeConfigFallsBackToDefaults(t *testing.T) {
	config := parseForecastL3RuntimeConfig(map[string]string{
		"growth.forecast_l3.max_active_runs":           "-1",
		"growth.forecast_l3.max_runs_per_day":          "99999",
		"growth.forecast_l3.max_user_runs_per_day":     "0",
		"growth.forecast_l3.min_priority_threshold":    "9.99",
		"growth.forecast_l3.dispatch.interval_minutes": "0",
		"growth.forecast_l3.quality.interval_minutes":  "-5",
		"growth.forecast_l3.default_engine_key":        "   ",
	})

	if config.MaxActiveRuns != defaultForecastL3RuntimeConfig.MaxActiveRuns {
		t.Fatalf("expected invalid max-active-runs to fall back to default, got %+v", config)
	}
	if config.MaxRunsPerDay <= defaultForecastL3RuntimeConfig.MaxRunsPerDay {
		t.Fatalf("expected oversized max-runs-per-day to be clamped above default, got %+v", config)
	}
	if config.MaxUserRunsPerDay != defaultForecastL3RuntimeConfig.MaxUserRunsPerDay {
		t.Fatalf("expected invalid user-run limit to fall back to default, got %+v", config)
	}
	if config.MinPriorityThreshold <= defaultForecastL3RuntimeConfig.MinPriorityThreshold {
		t.Fatalf("expected oversized priority threshold to be clamped above default, got %+v", config)
	}
	if config.DispatchIntervalMinutes != defaultForecastL3RuntimeConfig.DispatchIntervalMinutes {
		t.Fatalf("expected invalid dispatch interval to fall back to default, got %+v", config)
	}
	if config.QualityIntervalMinutes != defaultForecastL3RuntimeConfig.QualityIntervalMinutes {
		t.Fatalf("expected invalid quality interval to fall back to default, got %+v", config)
	}
	if config.DefaultEngineKey != defaultForecastL3RuntimeConfig.DefaultEngineKey {
		t.Fatalf("expected blank engine key to fall back to default, got %+v", config)
	}
}
