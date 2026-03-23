package repo

import (
	"strings"
	"testing"
	"time"
)

func TestBuildTushareStockInstrumentFact(t *testing.T) {
	fetchedAt := time.Date(2026, 3, 22, 10, 30, 0, 0, time.Local)
	fact, ok := buildTushareStockInstrumentFact(
		tushareStockBasicRecord{
			TSCode:     "600519.SH",
			Symbol:     "600519",
			Name:       "贵州茅台",
			Area:       "贵州",
			Industry:   "白酒",
			Market:     "主板",
			ListDate:   "20010827",
			ListStatus: "L",
			IsHS:       "H",
		},
		&tushareStockCompanyRecord{
			TSCode:        "600519.SH",
			Exchange:      "SSE",
			Chairman:      "张三",
			Province:      "贵州",
			City:          "遵义",
			Website:       "https://www.moutaichina.com",
			MainBusiness:  "茅台酒及系列酒生产销售",
			BusinessScope: "白酒制造与销售",
		},
		fetchedAt,
	)
	if !ok {
		t.Fatal("expected stock fact to be built")
	}
	if fact.InstrumentKey != "600519.SH" {
		t.Fatalf("expected instrument key 600519.SH, got %s", fact.InstrumentKey)
	}
	if fact.DisplayName != "贵州茅台" {
		t.Fatalf("expected display name 贵州茅台, got %s", fact.DisplayName)
	}
	if fact.ExchangeCode != "SH" {
		t.Fatalf("expected exchange code SH, got %s", fact.ExchangeCode)
	}
	if fact.ProductKey != "600519" {
		t.Fatalf("expected product key 600519, got %s", fact.ProductKey)
	}
	if fact.ListDate != "2001-08-27" {
		t.Fatalf("expected list date 2001-08-27, got %s", fact.ListDate)
	}
	if fact.Status != "ACTIVE" {
		t.Fatalf("expected ACTIVE status, got %s", fact.Status)
	}
	if fact.SourceUpdatedAt.IsZero() || !fact.SourceUpdatedAt.Equal(fetchedAt) {
		t.Fatalf("expected source updated at %s, got %s", fetchedAt, fact.SourceUpdatedAt)
	}
	if !strings.Contains(fact.MetadataJSON, `"industry":"白酒"`) {
		t.Fatalf("expected industry in metadata, got %s", fact.MetadataJSON)
	}
	if !strings.Contains(fact.MetadataJSON, `"main_business":"茅台酒及系列酒生产销售"`) {
		t.Fatalf("expected main_business in metadata, got %s", fact.MetadataJSON)
	}
}

func TestBuildTushareFuturesInstrumentFact(t *testing.T) {
	fetchedAt := time.Date(2026, 3, 22, 10, 45, 0, 0, time.Local)
	fact, ok := buildTushareFuturesInstrumentFact(
		tushareFuturesBasicRecord{
			TSCode:        "IF2606.CFX",
			Symbol:        "IF2606",
			Exchange:      "CFX",
			Name:          "沪深300股指2606",
			FutCode:       "IF",
			TradeUnit:     "手",
			Multiplier:    300,
			QuoteUnit:     "点",
			ListDate:      "20260301",
			DelistDate:    "20260619",
			DModeDesc:     "实物交割",
			TradeTimeDesc: "09:30-11:30,13:00-15:00",
			QuoteUnitDesc: "300元/点",
			LastTradeDate: "20260619",
			DeliveryMonth: "202606",
		},
		fetchedAt,
	)
	if !ok {
		t.Fatal("expected futures fact to be built")
	}
	if fact.InstrumentKey != "IF2606.CFX" {
		t.Fatalf("expected instrument key IF2606.CFX, got %s", fact.InstrumentKey)
	}
	if fact.DisplayName != "沪深300股指2606" {
		t.Fatalf("expected display name 沪深300股指2606, got %s", fact.DisplayName)
	}
	if fact.ExchangeCode != "CFX" {
		t.Fatalf("expected exchange code CFX, got %s", fact.ExchangeCode)
	}
	if fact.ProductKey != "IF" {
		t.Fatalf("expected product key IF, got %s", fact.ProductKey)
	}
	if fact.ListDate != "2026-03-01" {
		t.Fatalf("expected list date 2026-03-01, got %s", fact.ListDate)
	}
	if fact.DelistDate != "2026-06-19" {
		t.Fatalf("expected delist date 2026-06-19, got %s", fact.DelistDate)
	}
	if fact.Status != "ACTIVE" {
		t.Fatalf("expected ACTIVE status for not-yet-delisted contract, got %s", fact.Status)
	}
	if !strings.Contains(fact.MetadataJSON, `"trade_unit":"手"`) {
		t.Fatalf("expected trade_unit metadata, got %s", fact.MetadataJSON)
	}
	if !strings.Contains(fact.MetadataJSON, `"multiplier":300`) {
		t.Fatalf("expected multiplier metadata, got %s", fact.MetadataJSON)
	}
}

func TestResolveMarketInstrumentTruthPrefersNamedSourceByPriority(t *testing.T) {
	fetchedAt := time.Date(2026, 3, 22, 11, 0, 0, 0, time.Local)
	truth := resolveMarketInstrumentTruth(
		marketAssetClassStock,
		"600519.SH",
		[]marketInstrumentSourceFact{
			{
				AssetClass:      marketAssetClassStock,
				InstrumentKey:   "600519.SH",
				SourceKey:       "AKSHARE",
				ExternalSymbol:  "600519",
				DisplayName:     "",
				ExchangeCode:    "SH",
				ProductKey:      "600519",
				Status:          "ACTIVE",
				QualityScore:    0.35,
				SourceUpdatedAt: fetchedAt,
			},
			{
				AssetClass:      marketAssetClassStock,
				InstrumentKey:   "600519.SH",
				SourceKey:       "TUSHARE",
				ExternalSymbol:  "600519.SH",
				DisplayName:     "贵州茅台",
				ExchangeCode:    "SH",
				ProductKey:      "600519",
				ListDate:        "2001-08-27",
				Status:          "ACTIVE",
				MetadataJSON:    `{"industry":"白酒"}`,
				QualityScore:    1,
				SourceUpdatedAt: fetchedAt,
			},
		},
		[]string{"AKSHARE", "TUSHARE", "TICKERMD"},
	)
	if truth.SelectedSourceKey != "TUSHARE" {
		t.Fatalf("expected TUSHARE truth source, got %s", truth.SelectedSourceKey)
	}
	if truth.DisplayName != "贵州茅台" {
		t.Fatalf("expected display name 贵州茅台, got %s", truth.DisplayName)
	}
	if truth.ListDate != "2001-08-27" {
		t.Fatalf("expected list date 2001-08-27, got %s", truth.ListDate)
	}
	if truth.QualityScore < 0.99 {
		t.Fatalf("expected quality score near 1, got %.2f", truth.QualityScore)
	}
}

func TestResolveMarketInstrumentTruthFallsBackToPlaceholder(t *testing.T) {
	truth := resolveMarketInstrumentTruth(
		marketAssetClassFutures,
		"AU2606.SHF",
		nil,
		[]string{"TUSHARE", "AKSHARE"},
	)
	if truth.SelectedSourceKey != "LOCAL_PLACEHOLDER" {
		t.Fatalf("expected LOCAL_PLACEHOLDER source, got %s", truth.SelectedSourceKey)
	}
	if truth.DisplayName != "AU2606.SHF" {
		t.Fatalf("expected placeholder display name, got %s", truth.DisplayName)
	}
	if truth.ExchangeCode != "SHF" {
		t.Fatalf("expected exchange code SHF, got %s", truth.ExchangeCode)
	}
	if truth.Status != "ACTIVE" {
		t.Fatalf("expected ACTIVE fallback status, got %s", truth.Status)
	}
}
