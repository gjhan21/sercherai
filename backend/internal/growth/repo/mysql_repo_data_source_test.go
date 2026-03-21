package repo

import "testing"

func TestMergeDataSourceConfigMapPreservesProviderSpecificFields(t *testing.T) {
	existing := map[string]interface{}{
		"provider":                     "MYSELF",
		"endpoint":                     "https://qt.gtimg.cn/q=s_sh000001",
		"stock_kline_endpoint_tencent": "https://web.ifzq.gtimg.cn/appstock/app/fqkline/get",
		"stock_kline_endpoint_sina":    "https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData",
		"retry_times":                  1,
	}
	incoming := map[string]interface{}{
		"endpoint":          "https://custom.example.com/health",
		"retry_times":       3,
		"health_timeout_ms": 9000,
	}

	merged := mergeDataSourceConfigMap(existing, incoming)

	if merged["provider"] != "MYSELF" {
		t.Fatalf("expected provider to be preserved, got %#v", merged["provider"])
	}
	if merged["stock_kline_endpoint_tencent"] != existing["stock_kline_endpoint_tencent"] {
		t.Fatalf("expected tencent endpoint to be preserved, got %#v", merged["stock_kline_endpoint_tencent"])
	}
	if merged["endpoint"] != "https://custom.example.com/health" {
		t.Fatalf("expected endpoint override to win, got %#v", merged["endpoint"])
	}
	if merged["retry_times"] != 3 {
		t.Fatalf("expected retry_times override to win, got %#v", merged["retry_times"])
	}
	if merged["health_timeout_ms"] != 9000 {
		t.Fatalf("expected new field to be merged in, got %#v", merged["health_timeout_ms"])
	}
}

func TestCloneDataSourceConfigMapReturnsIndependentCopy(t *testing.T) {
	original := map[string]interface{}{
		"provider": "TUSHARE",
		"token":    "secret-token",
	}

	clone := cloneDataSourceConfigMap(original)
	clone["token"] = "changed-token"

	if original["token"] != "secret-token" {
		t.Fatalf("expected original map to remain unchanged, got %#v", original["token"])
	}
}
