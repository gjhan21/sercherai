package scheduler

import "testing"

func TestNormalizeCronExpr(t *testing.T) {
	cases := []struct {
		name       string
		raw        string
		want       string
		translated bool
	}{
		{name: "empty", raw: "  ", want: "", translated: false},
		{name: "passthrough six field", raw: "0 30 8 * * *", want: "0 30 8 * * *", translated: false},
		{name: "passthrough trims", raw: "  0 0 */2 * * *\n", want: "0 0 */2 * * *", translated: false},
		{name: "every 30 minutes", raw: "EVERY_30_MINUTES", want: "0 */30 * * * *", translated: true},
		{name: "case insensitive", raw: "every_20_minutes", want: "0 */20 * * * *", translated: true},
		{name: "every hour alias", raw: "EVERY_HOUR", want: "0 0 * * * *", translated: true},
		{name: "100 minutes falls back to 2h", raw: "EVERY_100_MINUTES", want: "0 0 */2 * * *", translated: true},
		{name: "unknown preset is left alone", raw: "EVERY_FORTNIGHT", want: "EVERY_FORTNIGHT", translated: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, translated := normalizeCronExpr(tc.raw)
			if got != tc.want {
				t.Fatalf("normalizeCronExpr(%q) = %q, want %q", tc.raw, got, tc.want)
			}
			if translated != tc.translated {
				t.Fatalf("normalizeCronExpr(%q) translated = %v, want %v", tc.raw, translated, tc.translated)
			}
		})
	}
}
