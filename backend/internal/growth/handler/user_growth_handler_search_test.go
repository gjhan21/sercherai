package handler

import (
	"testing"

	"sercherai/backend/internal/growth/model"
)

func TestNormalizeSearchKeyword(t *testing.T) {
	if got := normalizeSearchKeyword("  宁德   时代  "); got != "宁德 时代" {
		t.Fatalf("expected normalized keyword, got %q", got)
	}
	if got := normalizeSearchKeyword("  "); got != "" {
		t.Fatalf("expected empty keyword, got %q", got)
	}
}

func TestFilterStockRecommendationsByKeyword(t *testing.T) {
	items := []model.StockRecommendation{
		{ID: "1", Symbol: "300750.SZ", Name: "宁德时代", ReasonSummary: "趋势延续"},
		{ID: "2", Symbol: "600036.SH", Name: "招商银行", ReasonSummary: "高股息防御"},
		{ID: "3", Symbol: "000001.SZ", Name: "平安银行", ReasonSummary: "估值修复"},
	}

	matched, total := filterStockRecommendationsByKeyword(items, "银行", 2)
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if len(matched) != 2 {
		t.Fatalf("expected len=2, got %d", len(matched))
	}
	if matched[0].ID != "2" || matched[1].ID != "3" {
		t.Fatalf("unexpected order/content: %+v", matched)
	}
}

func TestFilterStockRecommendationsByKeywordMatchesReasonSummary(t *testing.T) {
	items := []model.StockRecommendation{
		{ID: "1", Symbol: "300750.SZ", Name: "宁德时代", ReasonSummary: "锂电景气修复"},
		{ID: "2", Symbol: "600036.SH", Name: "招商银行", ReasonSummary: "高股息防御"},
	}

	matched, total := filterStockRecommendationsByKeyword(items, "锂电", 5)
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(matched) != 1 || matched[0].ID != "1" {
		t.Fatalf("unexpected matched result: %+v", matched)
	}
}

func TestFilterFuturesStrategiesByKeyword(t *testing.T) {
	items := []model.FuturesStrategy{
		{ID: "f1", Contract: "RB2505", Name: "螺纹钢策略", ReasonSummary: "价差均值回归"},
		{ID: "f2", Contract: "IF2506", Name: "股指策略", ReasonSummary: "趋势跟随"},
	}

	matched, total := filterFuturesStrategiesByKeyword(items, "RB", 5)
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(matched) != 1 || matched[0].ID != "f1" {
		t.Fatalf("unexpected matched result: %+v", matched)
	}
}

func TestFilterFuturesStrategiesByKeywordMatchesDirectionAndReason(t *testing.T) {
	items := []model.FuturesStrategy{
		{ID: "f1", Contract: "RB2505", Name: "螺纹钢策略", Direction: "LONG", ReasonSummary: "价差均值回归"},
		{ID: "f2", Contract: "IF2506", Name: "股指策略", Direction: "SHORT", ReasonSummary: "趋势跟随"},
	}

	matched, total := filterFuturesStrategiesByKeyword(items, "short", 5)
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(matched) != 1 || matched[0].ID != "f2" {
		t.Fatalf("unexpected matched result: %+v", matched)
	}
}

func TestNormalizeSearchModeDefaultsToSuggest(t *testing.T) {
	if got := normalizeSearchMode(""); got != "suggest" {
		t.Fatalf("expected suggest, got %q", got)
	}
	if got := normalizeSearchMode("full"); got != "full" {
		t.Fatalf("expected full, got %q", got)
	}
	if got := normalizeSearchMode("unknown"); got != "suggest" {
		t.Fatalf("expected suggest fallback, got %q", got)
	}
}
