package repo

import (
	"errors"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestSplitSourcePriorityList(t *testing.T) {
	items := splitSourcePriorityList(" TUSHARE, AKSHARE | TICKERMD ; TUSHARE ")
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0] != "TUSHARE" || items[1] != "AKSHARE" || items[2] != "TICKERMD" {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestDeriveDefaultExternalSymbol(t *testing.T) {
	if got := deriveDefaultExternalSymbol("AKSHARE", marketAssetClassStock, "000001.SZ"); got != "000001" {
		t.Fatalf("unexpected akshare stock symbol: %s", got)
	}
	if got := deriveDefaultExternalSymbol("TICKERMD", marketAssetClassStock, "600519.SH"); got != "sh600519" {
		t.Fatalf("unexpected tickermd stock symbol: %s", got)
	}
	if got := deriveDefaultExternalSymbol("TUSHARE", marketAssetClassFutures, "AU2406.SHF"); got != "AU2406.SHF" {
		t.Fatalf("unexpected tushare futures symbol: %s", got)
	}
	if got := deriveDefaultExternalSymbol("AKSHARE", marketAssetClassFutures, "AU2406.SHF"); got != "au2406" {
		t.Fatalf("unexpected akshare futures symbol: %s", got)
	}
}

func TestNormalizeFuturesInventorySymbolList(t *testing.T) {
	items := normalizeFuturesInventorySymbolList([]string{"rb2405", " RB ", "au2406.shf", "IF2606"})
	if len(items) != 3 {
		t.Fatalf("expected 3 unique roots, got %d", len(items))
	}
	if items[0] != "RB" || items[1] != "AU" || items[2] != "IF" {
		t.Fatalf("unexpected normalized inventory symbols: %#v", items)
	}
}

func TestBuildMockFuturesInventorySnapshots(t *testing.T) {
	items := buildMockFuturesInventorySnapshots("MOCK", []string{"RB", "AU"}, 3)
	if len(items) != 6 {
		t.Fatalf("expected 6 mock inventory snapshots, got %d", len(items))
	}
	if items[0].Symbol != "RB" || items[0].SourceKey != "MOCK" {
		t.Fatalf("unexpected first mock inventory snapshot: %+v", items[0])
	}
	if items[0].Brand == "" || items[0].Place == "" || items[0].Grade == "" {
		t.Fatalf("expected mock inventory dimensions to be populated, got %+v", items[0])
	}
	if items[1].TradeDate == items[0].TradeDate {
		t.Fatalf("expected consecutive trade dates, got %+v", items[:2])
	}
}

func TestBuildFuturesInventoryExternalIDIncludesDimensions(t *testing.T) {
	first := buildFuturesInventoryExternalID("TUSHARE", "RB", "2026-03-20", "仓库A", "A1", "华东", "品牌甲", "上海", "标准")
	second := buildFuturesInventoryExternalID("TUSHARE", "RB", "2026-03-20", "仓库A", "A1", "华东", "品牌乙", "上海", "标准")
	if first == second {
		t.Fatalf("expected distinct external ids for different inventory dimensions, got %s", first)
	}
}

func TestParseTickerMDDailyBars(t *testing.T) {
	rows := [][]interface{}{
		{1.741248e+09, 10.8, 11.0, 10.7, 10.9, "2025-03-18 15:00:00", 1000.0},
		{1.7413344e+09, 10.9, 11.1, 10.8, 11.0, "2025-03-19 15:00:00", 1200.0},
	}

	items := parseTickerMDDailyBars(marketAssetClassStock, "TICKERMD", "000001.SZ", "sz000001", rows)
	if len(items) != 2 {
		t.Fatalf("expected 2 bars, got %d", len(items))
	}
	if items[0].TradeDate != "2025-03-18" || items[1].TradeDate != "2025-03-19" {
		t.Fatalf("unexpected trade dates: %#v", items)
	}
	if items[1].PrevClosePrice != 10.9 {
		t.Fatalf("expected prev close 10.9, got %.4f", items[1].PrevClosePrice)
	}
	if items[1].Volume != 1200 {
		t.Fatalf("expected volume 1200, got %d", items[1].Volume)
	}
}

func TestCanonicalMarketSourceKey(t *testing.T) {
	if got := canonicalMarketSourceKey("mock_stock", "MOCK"); got != "MOCK" {
		t.Fatalf("expected MOCK, got %s", got)
	}
	if got := canonicalMarketSourceKey("akshare", "AKSHARE"); got != "AKSHARE" {
		t.Fatalf("expected AKSHARE, got %s", got)
	}
}

func TestBuildDataSourceLookupCandidatesIncludesMockAlias(t *testing.T) {
	items := buildDataSourceLookupCandidates("MOCK")
	joined := strings.Join(items, ",")
	if !strings.Contains(joined, "mock_stock") {
		t.Fatalf("expected mock_stock alias in candidates, got %v", items)
	}
}

func TestDecodeTickerMDDailyRowsFromEnvelope(t *testing.T) {
	body := []byte(`{"data":[[1741248000,10.8,11.0,10.7,10.9,"2025-03-18 15:00:00",1000]]}`)
	rows, err := decodeTickerMDDailyRows(body)
	if err != nil {
		t.Fatalf("decode rows: %v", err)
	}
	if len(rows) != 1 || len(rows[0]) < 7 {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestDecodeTickerMDDailyRowsFromSymbolMap(t *testing.T) {
	body := []byte(`{"sz000001":[[1741248000,10.8,11.0,10.7,10.9,"2025-03-18 15:00:00",1000]]}`)
	rows, err := decodeTickerMDDailyRows(body)
	if err != nil {
		t.Fatalf("decode symbol rows: %v", err)
	}
	if len(rows) != 1 || len(rows[0]) < 7 {
		t.Fatalf("unexpected rows: %#v", rows)
	}
}

func TestDecodeTickerMDDailyRowsReturnsClearError(t *testing.T) {
	body := []byte(`{"code":500,"msg":"您的IP未授权"}`)
	_, err := decodeTickerMDDailyRows(body)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "未授权") {
		t.Fatalf("expected unauthorized error, got %v", err)
	}
}

func TestConvertAkshareBridgeDailyBarsForFutures(t *testing.T) {
	items := convertAkshareBridgeDailyBars(
		marketAssetClassFutures,
		"AKSHARE",
		[]string{"IF2606.CFX"},
		map[string]string{"IF2606.CFX": "if2606"},
		[]pythonBridgeDailyBarItem{
			{
				InstrumentKey:   "IF2606.CFX",
				ExternalSymbol:  "IF2606",
				TradeDate:       "2026-03-18",
				OpenPrice:       4578.0,
				HighPrice:       4582.2,
				LowPrice:        4529.6,
				ClosePrice:      4569.6,
				PrevClosePrice:  4557.8,
				SettlePrice:     4568.4,
				PrevSettlePrice: 4559.2,
				Volume:          43797,
				OpenInterest:    125306,
			},
		},
	)
	if len(items) != 1 {
		t.Fatalf("expected one futures bar, got %d", len(items))
	}
	if items[0].AssetClass != marketAssetClassFutures || items[0].InstrumentKey != "IF2606.CFX" {
		t.Fatalf("unexpected futures identity: %+v", items[0])
	}
	if items[0].SettlePrice != 4568.4 || items[0].PrevSettlePrice != 4559.2 {
		t.Fatalf("unexpected futures settle fields: %+v", items[0])
	}
	if items[0].OpenInterest != 125306 {
		t.Fatalf("expected open interest 125306, got %.4f", items[0].OpenInterest)
	}
}

func TestMyselfStockAPISymbol(t *testing.T) {
	if got := myselfStockAPISymbol("600519.SH", ""); got != "sh600519" {
		t.Fatalf("expected sh600519, got %s", got)
	}
	if got := myselfStockAPISymbol("000001.SZ", ""); got != "sz000001" {
		t.Fatalf("expected sz000001, got %s", got)
	}
}

func TestFetchStockQuotesFromTushareDateRangeUsesExplicitDates(t *testing.T) {
	var requestPayload struct {
		APIName string            `json:"api_name"`
		Params  map[string]string `json:"params"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if err := json.Unmarshal(body, &requestPayload); err != nil {
			t.Fatalf("decode request payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"code":0,"msg":"","data":{"fields":["ts_code","trade_date","open","high","low","close","pre_close","vol","amount"],"items":[["600519.SH","20240105",10,11,9,10.5,10,100,1000]]}}`)
	}))
	defer server.Close()

	previousEndpoint := tushareAPIEndpoint
	tushareAPIEndpoint = server.URL
	defer func() {
		tushareAPIEndpoint = previousEndpoint
	}()

	items, err := fetchStockQuotesFromTushareDateRange("token", "TUSHARE", []string{"600519.SH"}, "20240101", "20240131", 500)
	if err != nil {
		t.Fatalf("fetchStockQuotesFromTushareDateRange returned error: %v", err)
	}
	if requestPayload.APIName != "daily" {
		t.Fatalf("expected daily api_name, got %+v", requestPayload)
	}
	if requestPayload.Params["start_date"] != "20240101" || requestPayload.Params["end_date"] != "20240131" {
		t.Fatalf("expected explicit tushare date range, got %+v", requestPayload.Params)
	}
	if len(items) != 1 || items[0].TradeDate != "2024-01-05" {
		t.Fatalf("unexpected quote items: %+v", items)
	}
}

func TestParseMyselfTencentStockDailyPayload(t *testing.T) {
	body := []byte(`{"code":0,"msg":"","data":{"sh600519":{"day":[["2026-03-17","1468.000","1485.000","1498.070","1461.190","49454.000"],["2026-03-18","1489.000","1468.800","1496.500","1463.150","35551.000"]]}}}`)
	items, err := parseMyselfTencentStockDailyPayload("600519.SH", "MYSELF", body, 2)
	if err != nil {
		t.Fatalf("parse tencent stock payload: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 stock quotes, got %d", len(items))
	}
	if items[0].Symbol != "600519.SH" || items[1].PrevClosePrice != 1485 {
		t.Fatalf("unexpected tencent quote items: %+v", items)
	}
}

func TestParseMyselfSinaStockDailyPayload(t *testing.T) {
	body := []byte(`[{"day":"2026-03-17","open":"1468.000","high":"1498.070","low":"1461.190","close":"1485.000","volume":"4945361"},{"day":"2026-03-18","open":"1489.000","high":"1496.500","low":"1463.150","close":"1468.800","volume":"3555100"}]`)
	items, err := parseMyselfSinaStockDailyPayload("600519.SH", "MYSELF", body, 2)
	if err != nil {
		t.Fatalf("parse sina stock payload: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 stock quotes, got %d", len(items))
	}
	if items[1].PrevClosePrice != 1485 || items[1].Volume != 3555100 {
		t.Fatalf("unexpected sina stock quotes: %+v", items)
	}
}

func TestParseMyselfSinaFuturesDailyPayload(t *testing.T) {
	body := []byte(`/*<script>location.href='//sina.com';</script>*/
var _TEST=([{"d":"2026-03-17","o":"4600.000","h":"4648.600","l":"4556.600","c":"4557.800","v":"44272","p":"121279","s":"0.000"},{"d":"2026-03-18","o":"4578.000","h":"4582.200","l":"4529.600","c":"4569.600","v":"43797","p":"125306","s":"0.000"}]);`)
	items, err := parseMyselfSinaFuturesDailyPayload("IF2606.CFX", "MYSELF", "IF2606", body, 2)
	if err != nil {
		t.Fatalf("parse sina futures payload: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 futures bars, got %d", len(items))
	}
	if items[0].PrevClosePrice != 4600 || items[1].PrevClosePrice != 4557.8 {
		t.Fatalf("unexpected futures prev close chain: %+v", items)
	}
	if items[1].OpenInterest != 125306 || items[1].SettlePrice != 4569.6 {
		t.Fatalf("unexpected futures bar fields: %+v", items[1])
	}
}

const marketProviderRoutingPriorityQueryPattern = `(?s)SELECT primary_provider_key,\s*COALESCE\(CAST\(fallback_provider_keys_json AS CHAR\), ''\),\s*fallback_allowed,\s*mock_allowed\s+FROM market_provider_routing_policies`

func assertTableMissingSQLError() error {
	return errors.New("Error 1146 (42S02): Table 'sercherai.market_provider_routing_policies' doesn't exist")
}

func TestResolveRequestedMarketSourceKeysWithGovernanceUsesRoutingPolicyForAuto(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(marketProviderRoutingPriorityQueryPattern).
		WithArgs("STOCK", "DAILY_BARS").
		WillReturnRows(sqlmock.NewRows([]string{
			"primary_provider_key",
			"fallback_provider_keys_json",
			"fallback_allowed",
			"mock_allowed",
		}).AddRow("TUSHARE", `["AKSHARE","TICKERMD"]`, true, false))

	items := repo.resolveRequestedMarketSourceKeysWithGovernance("AUTO", "STOCK", "DAILY_BARS", marketStockPriorityConfigKey, []string{"TUSHARE", "AKSHARE"})

	if len(items) != 3 {
		t.Fatalf("expected 3 routed items, got %d", len(items))
	}
	if items[0] != "TUSHARE" || items[1] != "AKSHARE" || items[2] != "TICKERMD" {
		t.Fatalf("unexpected routed items: %#v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestResolveRequestedMarketSourceKeysWithGovernanceFallsBackToLegacyPriorityWhenPolicyMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(marketProviderRoutingPriorityQueryPattern).
		WithArgs("FUTURES", "DAILY_BARS").
		WillReturnError(assertTableMissingSQLError())
	mock.ExpectQuery(`(?s)SELECT config_value\s+FROM system_configs`).
		WithArgs(marketFuturesPriorityConfigKey).
		WillReturnRows(sqlmock.NewRows([]string{"config_value"}).AddRow("TUSHARE,MYSELF,MOCK"))

	items := repo.resolveRequestedMarketSourceKeysWithGovernance("AUTO", "FUTURES", "DAILY_BARS", marketFuturesPriorityConfigKey, []string{"TUSHARE", "TICKERMD"})

	if len(items) != 3 {
		t.Fatalf("expected legacy fallback items, got %d", len(items))
	}
	if items[0] != "TUSHARE" || items[1] != "MYSELF" || items[2] != "MOCK" {
		t.Fatalf("unexpected legacy fallback items: %#v", items)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestBuildMarketSourceRoutingSummaryUsesGovernedAutoPriority(t *testing.T) {
	summary := buildMarketSourceRoutingSummary("AUTO", []string{"TUSHARE", "AKSHARE", "TICKERMD"}, marketAssetClassStock, marketDataKindDailyBars)

	if summary.SelectedSource != "TUSHARE" {
		t.Fatalf("expected selected source TUSHARE, got %s", summary.SelectedSource)
	}
	if len(summary.FallbackSourceKeys) != 2 || summary.FallbackSourceKeys[0] != "AKSHARE" || summary.FallbackSourceKeys[1] != "TICKERMD" {
		t.Fatalf("unexpected fallback chain: %#v", summary.FallbackSourceKeys)
	}
	if summary.RoutingPolicyKey != "market.stock.daily" {
		t.Fatalf("expected stock routing policy key, got %s", summary.RoutingPolicyKey)
	}
	if summary.DecisionReason != "governed_auto_priority" {
		t.Fatalf("expected governed auto decision reason, got %s", summary.DecisionReason)
	}
}

func TestBuildMarketSourceRoutingSummaryKeepsExplicitSourceDecision(t *testing.T) {
	summary := buildMarketSourceRoutingSummary("MYSELF", []string{"MYSELF"}, marketAssetClassFutures, marketDataKindDailyBars)

	if summary.SelectedSource != "MYSELF" {
		t.Fatalf("expected selected source MYSELF, got %s", summary.SelectedSource)
	}
	if len(summary.FallbackSourceKeys) != 0 {
		t.Fatalf("expected no fallback chain for explicit source, got %#v", summary.FallbackSourceKeys)
	}
	if summary.RoutingPolicyKey != "market.futures.daily" {
		t.Fatalf("expected futures routing policy key, got %s", summary.RoutingPolicyKey)
	}
	if summary.DecisionReason != "explicit_source" {
		t.Fatalf("expected explicit source decision reason, got %s", summary.DecisionReason)
	}
}
