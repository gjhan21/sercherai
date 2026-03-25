package repo

import (
	"bytes"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCanonicalStockInstrumentKey(t *testing.T) {
	cases := map[string]string{
		"600519":    "600519.SH",
		"000001":    "000001.SZ",
		"430047":    "430047.BJ",
		"sh600519":  "600519.SH",
		"sz000001":  "000001.SZ",
		"bj430047":  "430047.BJ",
		"600519.SH": "600519.SH",
		"":          "",
	}
	for raw, expected := range cases {
		if got := canonicalStockInstrumentKey(raw); got != expected {
			t.Fatalf("canonicalStockInstrumentKey(%q) = %q, want %q", raw, got, expected)
		}
	}
}

func TestNormalizeStockSymbolListCanonicalizesAndDeduplicates(t *testing.T) {
	items := normalizeStockSymbolList([]string{
		"600519",
		"600519.SH",
		"sh600519",
		"000001",
		"000001.SZ",
		"sz000001",
		"430047",
		"bj430047",
		"",
		"   ",
	})
	expected := []string{"600519.SH", "000001.SZ", "430047.BJ"}
	if !reflect.DeepEqual(items, expected) {
		t.Fatalf("unexpected normalized stock symbol list: %#v", items)
	}
}

func TestBuildTushareStockInstrumentFacts(t *testing.T) {
	fetchedAt := time.Date(2026, 3, 23, 10, 0, 0, 0, time.Local)
	facts := buildTushareStockInstrumentFacts([]tushareStockBasicRecord{
		{
			TSCode:     "600519.SH",
			Symbol:     "600519",
			Name:       "贵州茅台",
			Area:       "贵州",
			Industry:   "白酒",
			Market:     "主板",
			ListDate:   "20010827",
			ListStatus: "L",
			IsHS:       "H",
			Exchange:   "SSE",
		},
		{
			TSCode:     "000001.SZ",
			Symbol:     "000001",
			Name:       "平安银行",
			Area:       "深圳",
			Industry:   "银行",
			Market:     "主板",
			ListDate:   "19910403",
			ListStatus: "L",
			Exchange:   "SZSE",
		},
	}, fetchedAt)
	if len(facts) != 2 {
		t.Fatalf("expected 2 facts, got %d", len(facts))
	}
	if facts[0].InstrumentKey != "600519.SH" || facts[1].InstrumentKey != "000001.SZ" {
		t.Fatalf("unexpected instrument keys: %#v", facts)
	}
	if facts[0].DisplayName != "贵州茅台" || facts[1].DisplayName != "平安银行" {
		t.Fatalf("unexpected display names: %#v", facts)
	}
	if facts[0].ListDate != "2001-08-27" || facts[1].ListDate != "1991-04-03" {
		t.Fatalf("unexpected list dates: %#v", facts)
	}
	if facts[0].SourceUpdatedAt != fetchedAt || facts[1].SourceUpdatedAt != fetchedAt {
		t.Fatalf("unexpected source updated at: %#v", facts)
	}
}

func TestFetchStockInstrumentFactsFromTushareReturnsFullMarketFactsWhenKeysEmpty(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	requestCount := 0
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if !bytes.Contains(body, []byte(`"api_name":"stock_basic"`)) {
			t.Fatalf("expected stock_basic api call, got %s", string(body))
		}
		payload := `{"code":0,"msg":"","data":{"fields":["ts_code","symbol","name","area","industry","market","list_date","delist_date","list_status","is_hs","exchange"],"items":[["600519.SH","600519","贵州茅台","贵州","白酒","主板","20010827","","L","H","SSE"],["000001.SZ","000001","平安银行","深圳","银行","主板","19910403","","L","","SZSE"]]}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(payload)),
			Header:     make(http.Header),
		}, nil
	})

	facts, err := fetchStockInstrumentFactsFromTushare("token", "TUSHARE", nil, 1000)
	if err != nil {
		t.Fatalf("fetchStockInstrumentFactsFromTushare: %v", err)
	}
	if requestCount != 1 {
		t.Fatalf("expected 1 request, got %d", requestCount)
	}
	if len(facts) != 2 {
		t.Fatalf("expected 2 facts, got %d", len(facts))
	}
	if facts[0].SourceKey != "TUSHARE" || facts[1].SourceKey != "TUSHARE" {
		t.Fatalf("expected source key TUSHARE, got %#v", facts)
	}
	if facts[0].InstrumentKey != "600519.SH" || facts[1].InstrumentKey != "000001.SZ" {
		t.Fatalf("unexpected facts: %#v", facts)
	}
}

func TestFetchStockInstrumentFactsFromTushareCanonicalizesRequestedKeys(t *testing.T) {
	originalTransport := http.DefaultTransport
	defer func() {
		http.DefaultTransport = originalTransport
	}()

	requestBodies := make([]string, 0, 2)
	http.DefaultTransport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		requestBodies = append(requestBodies, string(body))
		switch len(requestBodies) {
		case 1:
			if !bytes.Contains(body, []byte(`"ts_code":"600519.SH"`)) {
				t.Fatalf("expected canonical ts_code in stock_basic request, got %s", string(body))
			}
			payload := `{"code":0,"msg":"","data":{"fields":["ts_code","symbol","name","area","industry","market","list_date","delist_date","list_status","is_hs","exchange"],"items":[["600519.SH","600519","贵州茅台","贵州","白酒","主板","20010827","","L","H","SSE"]]}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(payload)),
				Header:     make(http.Header),
			}, nil
		default:
			if !bytes.Contains(body, []byte(`"api_name":"stock_company"`)) || !bytes.Contains(body, []byte(`"ts_code":"600519.SH"`)) {
				t.Fatalf("expected canonical ts_code in stock_company request, got %s", string(body))
			}
			payload := `{"code":0,"msg":"","data":{"fields":["ts_code","exchange","chairman","manager","secretary","reg_capital","setup_date","province","city","website","email","employees","main_business","business_scope"],"items":[["600519.SH","SSE","","","","","","","","","","","",""]]}}`
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(payload)),
				Header:     make(http.Header),
			}, nil
		}
	})

	facts, err := fetchStockInstrumentFactsFromTushare("token", "TUSHARE", []string{"600519"}, 1000)
	if err != nil {
		t.Fatalf("fetchStockInstrumentFactsFromTushare: %v", err)
	}
	if len(requestBodies) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requestBodies))
	}
	if len(facts) != 1 || facts[0].InstrumentKey != "600519.SH" {
		t.Fatalf("unexpected facts: %#v", facts)
	}
}

func TestNormalizeMarketInstrumentKeysCanonicalizesStockKeys(t *testing.T) {
	items := normalizeMarketInstrumentKeys(marketAssetClassStock, []string{
		"600519",
		"600519.SH",
		"sh600519",
		"000001",
		"000001.SZ",
	})
	expected := []string{"600519.SH", "000001.SZ"}
	if !reflect.DeepEqual(items, expected) {
		t.Fatalf("unexpected market instrument keys: %#v", items)
	}
}

func TestMergeMarketInstrumentKeysFromFactsUsesFetchedStockUniverse(t *testing.T) {
	items := mergeMarketInstrumentKeysFromFacts(marketAssetClassStock, nil, []marketInstrumentSourceFact{
		{InstrumentKey: "600519.SH"},
		{InstrumentKey: "600519"},
		{InstrumentKey: "000001.SZ"},
		{InstrumentKey: "000001"},
	})
	expected := []string{"600519.SH", "000001.SZ"}
	if !reflect.DeepEqual(items, expected) {
		t.Fatalf("unexpected merged market instrument keys: %#v", items)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

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

func TestBuildMarketInstrumentSourceFactsFromUniverseItems(t *testing.T) {
	fetchedAt := time.Date(2026, 3, 24, 9, 30, 0, 0, time.Local)
	facts := buildMarketInstrumentSourceFactsFromUniverseItems("TUSHARE", "INDEX", []marketUniverseSourceItem{
		{
			AssetType:      "INDEX",
			InstrumentKey:  "000300.SH",
			ExternalSymbol: "000300.SH",
			DisplayName:    "沪深300",
			ExchangeCode:   "SH",
			Status:         "ACTIVE",
			ListDate:       "2005-04-08",
			MetadataJSON:   `{"category":"broad_index"}`,
		},
	}, fetchedAt)
	if len(facts) != 1 {
		t.Fatalf("expected 1 fact, got %+v", facts)
	}
	if facts[0].AssetClass != "INDEX" {
		t.Fatalf("expected asset class INDEX, got %s", facts[0].AssetClass)
	}
	if facts[0].InstrumentKey != "000300.SH" {
		t.Fatalf("expected instrument key 000300.SH, got %s", facts[0].InstrumentKey)
	}
	if facts[0].SourceKey != "TUSHARE" {
		t.Fatalf("expected source key TUSHARE, got %s", facts[0].SourceKey)
	}
	if facts[0].ProductKey != "000300" {
		t.Fatalf("expected product key 000300, got %s", facts[0].ProductKey)
	}
	if facts[0].SourceUpdatedAt.IsZero() || !facts[0].SourceUpdatedAt.Equal(fetchedAt) {
		t.Fatalf("expected fetched at %s, got %s", fetchedAt, facts[0].SourceUpdatedAt)
	}
}

func TestBuildFuturesInstrumentProfileSnapshotAggregatesContractChainAndInventoryDimensions(t *testing.T) {
	profile := buildFuturesInstrumentProfileSnapshot(
		"AU",
		[]marketInstrumentTruth{
			{
				AssetClass:      marketAssetClassFutures,
				InstrumentKey:   "AU2506.SHF",
				DisplayName:     "沪金2506",
				ExchangeCode:    "SHF",
				ProductKey:      "AU",
				SourceUpdatedAt: time.Date(2026, 3, 24, 9, 0, 0, 0, time.Local),
			},
			{
				AssetClass:      marketAssetClassFutures,
				InstrumentKey:   "AU2508.SHF",
				DisplayName:     "沪金2508",
				ExchangeCode:    "SHF",
				ProductKey:      "AU",
				SourceUpdatedAt: time.Date(2026, 3, 24, 9, 10, 0, 0, time.Local),
			},
		},
		[]futuresInstrumentInventoryProfileRow{
			{Symbol: "AU", Place: "上海", Warehouse: "上期所一库", Brand: "国标一号", Grade: "标准品"},
			{Symbol: "AU", Place: "深圳", Warehouse: "上期所二库", Brand: "国标一号", Grade: "标准品"},
		},
	)

	if profile.ProductKey != "AU" {
		t.Fatalf("expected product key AU, got %s", profile.ProductKey)
	}
	if profile.ExchangeCode != "SHF" {
		t.Fatalf("expected exchange code SHF, got %s", profile.ExchangeCode)
	}
	if len(profile.ContractChain) != 2 || profile.ContractChain[0] != "AU2506.SHF" {
		t.Fatalf("unexpected contract chain: %+v", profile.ContractChain)
	}
	if len(profile.DeliveryPlaces) != 2 || profile.DeliveryPlaces[0] != "上海" {
		t.Fatalf("unexpected delivery places: %+v", profile.DeliveryPlaces)
	}
	if len(profile.Warehouses) != 2 || len(profile.Brands) != 1 || len(profile.Grades) != 1 {
		t.Fatalf("unexpected profile dimensions: %+v", profile)
	}
	if len(profile.InventoryMetricKeys) != 3 {
		t.Fatalf("expected inventory metric keys, got %+v", profile.InventoryMetricKeys)
	}
	if profile.Metadata["contract_count"] != float64(2) {
		t.Fatalf("expected contract_count metadata, got %+v", profile.Metadata)
	}
}
