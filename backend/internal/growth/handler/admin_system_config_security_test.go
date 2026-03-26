package handler

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestIsSensitiveSystemConfigKey(t *testing.T) {
	t.Parallel()

	cases := []struct {
		key       string
		sensitive bool
	}{
		{key: "payment.signing_secret", sensitive: true},
		{key: "oss.qiniu.access_key", sensitive: true},
		{key: "payment.channel.yolkpay.key", sensitive: true},
		{key: "payment.channel.wechat.private_key", sensitive: true},
		{key: "payment.channel.wechat.api_v3_key", sensitive: true},
		{key: "stock.quotes.default_source_key", sensitive: false},
		{key: "market.stock.daily.routing_policy_key", sensitive: false},
		{key: "scheduler.auto_retry.enabled", sensitive: false},
		{key: "oss.qiniu.domain", sensitive: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()
			got := isSensitiveSystemConfigKey(tc.key)
			if got != tc.sensitive {
				t.Fatalf("isSensitiveSystemConfigKey(%q)=%v, want %v", tc.key, got, tc.sensitive)
			}
		})
	}
}

func TestSanitizeSystemConfigItems(t *testing.T) {
	t.Parallel()

	items := []model.SystemConfig{
		{ConfigKey: "oss.qiniu.access_key", ConfigValue: "ABCDEFGH"},
		{ConfigKey: "oss.qiniu.domain", ConfigValue: "img.example.com"},
	}

	masked := sanitizeSystemConfigItems(items, false)
	if masked[0].ConfigValue == items[0].ConfigValue {
		t.Fatalf("expected sensitive config value to be masked")
	}
	if masked[1].ConfigValue != items[1].ConfigValue {
		t.Fatalf("non-sensitive value should remain unchanged")
	}

	raw := sanitizeSystemConfigItems(items, true)
	if raw[0].ConfigValue != items[0].ConfigValue {
		t.Fatalf("includeSensitive=true should keep raw value")
	}
}

func TestMaskSystemConfigValueForAudit(t *testing.T) {
	t.Parallel()

	if got := maskSystemConfigValueForAudit("payment.signing_secret", "secret-123456"); got == "secret-123456" {
		t.Fatalf("expected masked value for sensitive config")
	}
	if got := maskSystemConfigValueForAudit("oss.qiniu.domain", "img.example.com"); got != "img.example.com" {
		t.Fatalf("expected non-sensitive value unchanged, got %q", got)
	}
}
