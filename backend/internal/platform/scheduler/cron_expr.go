package scheduler

import "strings"

// legacyCronPresets maps the schedule labels written by the early scheduler
// migrations onto six-field cron expressions.
//
// Some presets have no exact cron equivalent: a 100-minute cadence does not
// divide an hour, so every 100 minutes cannot be expressed by a step value.
// Those cases are mapped to the closest cadence a cron expression can reach
// and the deviation is stated here rather than hidden in a seed file.
var legacyCronPresets = map[string]string{
	"EVERY_MINUTE":      "0 * * * * *",
	"EVERY_5_MINUTES":   "0 */5 * * * *",
	"EVERY_10_MINUTES":  "0 */10 * * * *",
	"EVERY_15_MINUTES":  "0 */15 * * * *",
	"EVERY_20_MINUTES":  "0 */20 * * * *",
	"EVERY_30_MINUTES":  "0 */30 * * * *",
	"EVERY_HOUR":        "0 0 * * * *",
	"EVERY_60_MINUTES":  "0 0 * * * *",
	"EVERY_2_HOURS":     "0 0 */2 * * *",
	"EVERY_100_MINUTES": "0 0 */2 * * *", // every 2h: 100 minutes has no cron step that divides 60
	"EVERY_DAY":         "0 0 0 * * *",
	"DAILY":             "0 0 0 * * *",
}

// normalizeCronExpr turns a stored schedule into a cron expression.
//
// Legacy preset labels are translated so rows seeded before the admin form
// switched to raw cron keep running. Anything else is returned trimmed but
// otherwise untouched, so cron.AddFunc still reports a precise parse error for
// a genuinely malformed expression instead of this function guessing.
func normalizeCronExpr(raw string) (expr string, translated bool) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", false
	}
	if mapped, ok := legacyCronPresets[strings.ToUpper(trimmed)]; ok {
		return mapped, true
	}
	return trimmed, false
}
