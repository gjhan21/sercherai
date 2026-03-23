package repo

import (
	"reflect"
	"testing"
	"time"
)

func TestBuildStockStatusTruthRecordUsesNameMetadataAndBar(t *testing.T) {
	record := buildStockStatusTruthRecord(stockStatusTruthInput{
		TradeDate:      time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local),
		Symbol:         "600001.SH",
		DisplayName:    "*ST测试",
		SelectedSource: "TUSHARE",
		Volume:         0,
		Turnover:       0,
		ListDate:       time.Date(2020, 1, 2, 0, 0, 0, 0, time.Local),
		MetadataJSON:   `{"risk_warning":true}`,
	})

	if record.Symbol != "600001.SH" {
		t.Fatalf("expected symbol 600001.SH, got %s", record.Symbol)
	}
	if !record.IsSuspended {
		t.Fatal("expected suspended flag to be true")
	}
	if !record.IsST {
		t.Fatal("expected ST flag to be true")
	}
	if !record.RiskWarning {
		t.Fatal("expected risk warning flag to be true")
	}
	if record.ListDate != "2020-01-02" {
		t.Fatalf("expected list date 2020-01-02, got %s", record.ListDate)
	}
	if !reflect.DeepEqual(record.ReasonCodes, []string{"SUSPENDED_PROXY", "ST_NAME_PREFIX", "RISK_WARNING"}) {
		t.Fatalf("unexpected reason codes: %#v", record.ReasonCodes)
	}
}

func TestBuildFuturesContractMappingRecordPrefersHighestTurnoverAndEarliestExpiry(t *testing.T) {
	record, ok := buildFuturesContractMappingRecord(
		time.Date(2026, 3, 22, 0, 0, 0, 0, time.Local),
		[]futuresContractSnapshot{
			{InstrumentKey: "IF2606.CFX", ProductKey: "IF", ExchangeCode: "CFX", Turnover: 1000000, OpenInterest: 320000, SelectedSource: "TUSHARE"},
			{InstrumentKey: "IF2605.CFX", ProductKey: "IF", ExchangeCode: "CFX", Turnover: 800000, OpenInterest: 280000, SelectedSource: "TUSHARE"},
			{InstrumentKey: "IF2604.CFX", ProductKey: "IF", ExchangeCode: "CFX", Turnover: 300000, OpenInterest: 120000, SelectedSource: "TUSHARE"},
		},
	)
	if !ok {
		t.Fatal("expected mapping record to be built")
	}
	if record.ProductKey != "IF" || record.ExchangeCode != "CFX" {
		t.Fatalf("unexpected mapping identity: %+v", record)
	}
	if record.DominantInstrumentKey != "IF2606.CFX" {
		t.Fatalf("expected dominant IF2606.CFX, got %s", record.DominantInstrumentKey)
	}
	if record.SecondaryInstrumentKey != "IF2605.CFX" {
		t.Fatalf("expected secondary IF2605.CFX, got %s", record.SecondaryInstrumentKey)
	}
	if record.NearInstrumentKey != "IF2604.CFX" {
		t.Fatalf("expected near IF2604.CFX, got %s", record.NearInstrumentKey)
	}
	if record.MappingMethod != "TURNOVER_OPEN_INTEREST" {
		t.Fatalf("expected TURNOVER_OPEN_INTEREST mapping method, got %s", record.MappingMethod)
	}
}

func TestNormalizeMarketDataQualitySeverityDefaultsToWarn(t *testing.T) {
	if got := normalizeMarketDataQualitySeverity("error"); got != "ERROR" {
		t.Fatalf("expected ERROR, got %s", got)
	}
	if got := normalizeMarketDataQualitySeverity("bad-value"); got != "WARN" {
		t.Fatalf("expected WARN fallback, got %s", got)
	}
}
