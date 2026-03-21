package repo

import "testing"

func TestStrategyConfigKeysFitSystemConfigKeyLimit(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		configBase string
		idPrefix   string
	}{
		{name: "seed", configBase: strategySeedSetConfigPrefix, idPrefix: "sd"},
		{name: "agent", configBase: strategyAgentProfileConfigPrefix, idPrefix: "ag"},
		{name: "scenario", configBase: strategyScenarioTemplateConfigPrefix, idPrefix: "sc"},
		{name: "policy", configBase: strategyPublishPolicyConfigPrefix, idPrefix: "po"},
	}

	seen := map[string]struct{}{}
	for _, tc := range cases {
		id := newStrategyConfigID(tc.idPrefix)
		if _, ok := seen[id]; ok {
			t.Fatalf("%s id should be unique, got duplicate %q", tc.name, id)
		}
		seen[id] = struct{}{}

		configKey := tc.configBase + id
		if got := len(configKey); got > 64 {
			t.Fatalf("%s config_key too long: got %d chars for %q", tc.name, got, configKey)
		}
	}
}
