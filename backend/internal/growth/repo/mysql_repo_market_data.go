package repo

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"sercherai/backend/internal/growth/model"
)

type stockQuoteCandle struct {
	Symbol         string
	TradeDate      time.Time
	OpenPrice      float64
	HighPrice      float64
	LowPrice       float64
	ClosePrice     float64
	PrevClosePrice float64
	Volume         float64
	Turnover       float64
}

type stockDailyBasicPoint struct {
	Symbol       string
	TradeDate    time.Time
	TurnoverRate float64
	VolumeRatio  float64
	PeTTM        float64
	PB           float64
	TotalMV      float64
	CircMV       float64
	SourceKey    string
}

type stockMoneyflowPoint struct {
	Symbol        string
	TradeDate     time.Time
	NetMFAmount   float64
	BuyLGAmount   float64
	SellLGAmount  float64
	BuyELGAmount  float64
	SellELGAmount float64
	SourceKey     string
}

type stockLimitUpPoint struct {
	Symbol       string
	TradeDate    time.Time
	Name         string
	LuTime       string
	OpenTime     string
	LastTime     string
	Tag          string
	Theme        string
	Status       string
	LimitOrder   float64
	LuLimitOrder float64
	BidAmount    float64
	BidChange    float64
	FloatMV      float64
	PctChg       float64
}

type stockTopListPoint struct {
	Symbol     string
	TradeDate  time.Time
	Name       string
	BuyAmount  float64
	SellAmount float64
	NetAmount  float64
	Reason     string
}

type stockNewsRawPoint struct {
	SourceKey   string
	Symbol      string
	PublishedAt time.Time
	Title       string
	Content     string
	URL         string
	Sentiment   string
}

type stockNewsSignal struct {
	Heat         int
	PositiveRate float64
}

type quantRiskAggregate struct {
	Sum5  float64
	Cnt5  int
	Hit5  int
	Sum10 float64
	Cnt10 int
	Hit10 int
}

type metricRange struct {
	Min float64
	Max float64
}

type dailyStockCandidate struct {
	Symbol         string
	Name           string
	Score          float64
	RiskLevel      string
	PositionRange  string
	ValidFrom      time.Time
	ValidTo        time.Time
	Status         string
	ReasonSummary  string
	TechScore      float64
	FundScore      float64
	SentimentScore float64
	MoneyFlowScore float64
	TakeProfit     string
	StopLoss       string
	RiskNote       string
}

func normalizeStockSymbolList(symbols []string) []string {
	seen := make(map[string]struct{}, len(symbols))
	result := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		normalized := canonicalStockInstrumentKey(symbol)
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func defaultMockStockSymbols() []string {
	return []string{
		"600519.SH", "601318.SH", "600036.SH", "600276.SH", "601012.SH",
		"000333.SZ", "300750.SZ", "002594.SZ", "688981.SH", "601888.SH",
		"000858.SZ", "000001.SZ", "601166.SH", "300015.SZ", "000651.SZ",
	}
}

func symbolSeed(symbol string) float64 {
	sum := 0
	for _, ch := range symbol {
		sum += int(ch)
	}
	return float64((sum%97)+3) / 10.0
}

func mockBasePriceBySymbol(symbol string) float64 {
	if strings.HasSuffix(symbol, ".SZ") {
		return 18 + symbolSeed(symbol)*4.5
	}
	return 25 + symbolSeed(symbol)*5.2
}

func buildMockStockQuotes(symbols []string, days int, sourceKey string) []model.StockMarketQuote {
	now := time.Now()
	tradeDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	items := make([]model.StockMarketQuote, 0, len(symbols)*days)
	for _, symbol := range symbols {
		seed := symbolSeed(symbol)
		prevClose := mockBasePriceBySymbol(symbol)
		for offset := days - 1; offset >= 0; offset-- {
			currentDay := tradeDay.AddDate(0, 0, -offset)
			drift := 0.0008 + 0.0015*math.Sin(float64(days-offset)/11.0+seed/5.0)
			wave := 0.006 * math.Sin(float64(offset)/2.7+seed)
			noise := 0.003 * math.Cos(float64(offset)/1.9+seed*1.8)
			closePrice := prevClose * (1 + drift + wave + noise)
			if closePrice <= 0 {
				closePrice = prevClose
			}
			openPrice := prevClose * (1 + noise*0.6)
			highPrice := math.Max(openPrice, closePrice) * (1 + 0.002 + math.Abs(wave)*0.5)
			lowPrice := math.Min(openPrice, closePrice) * (1 - 0.002 - math.Abs(wave)*0.45)
			if lowPrice <= 0 {
				lowPrice = math.Min(openPrice, closePrice)
			}

			volumeBase := 1_800_000 + seed*260_000
			volumeWave := (1 + 0.35*math.Sin(float64(offset)/3.1+seed*0.7))
			volume := int64(math.Max(200_000, math.Round(volumeBase*volumeWave)))
			turnover := closePrice * float64(volume)

			items = append(items, model.StockMarketQuote{
				Symbol:         symbol,
				TradeDate:      currentDay.Format("2006-01-02"),
				OpenPrice:      roundTo(openPrice, 4),
				HighPrice:      roundTo(highPrice, 4),
				LowPrice:       roundTo(lowPrice, 4),
				ClosePrice:     roundTo(closePrice, 4),
				PrevClosePrice: roundTo(prevClose, 4),
				Volume:         volume,
				Turnover:       roundTo(turnover, 2),
				SourceKey:      sourceKey,
			})
			prevClose = closePrice
		}
	}
	return items
}

func fetchStockQuotesFromEndpoint(endpoint string, sourceKey string, symbols []string, days int, timeoutMS int) ([]model.StockMarketQuote, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return nil, err
	}
	query := parsed.Query()
	if len(symbols) > 0 {
		query.Set("symbols", strings.Join(symbols, ","))
	}
	if days > 0 {
		query.Set("days", strconv.Itoa(days))
	}
	parsed.RawQuery = query.Encode()
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	resp, err := client.Get(parsed.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("quotes endpoint status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type quoteItem struct {
		Symbol         string  `json:"symbol"`
		TradeDate      string  `json:"trade_date"`
		Date           string  `json:"date"`
		OpenPrice      float64 `json:"open_price"`
		HighPrice      float64 `json:"high_price"`
		LowPrice       float64 `json:"low_price"`
		ClosePrice     float64 `json:"close_price"`
		PrevClosePrice float64 `json:"prev_close_price"`
		Volume         float64 `json:"volume"`
		Turnover       float64 `json:"turnover"`
	}
	payload := struct {
		Data  []quoteItem `json:"data"`
		Items []quoteItem `json:"items"`
	}{}
	items := make([]quoteItem, 0)
	if err := json.Unmarshal(body, &payload); err == nil {
		if len(payload.Data) > 0 {
			items = payload.Data
		} else if len(payload.Items) > 0 {
			items = payload.Items
		}
	}
	if len(items) == 0 {
		if err := json.Unmarshal(body, &items); err != nil {
			return nil, err
		}
	}

	symbolFilter := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbolFilter[strings.ToUpper(strings.TrimSpace(symbol))] = struct{}{}
	}

	result := make([]model.StockMarketQuote, 0, len(items))
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		if len(symbolFilter) > 0 {
			if _, ok := symbolFilter[symbol]; !ok {
				continue
			}
		}
		dateText := strings.TrimSpace(item.TradeDate)
		if dateText == "" {
			dateText = strings.TrimSpace(item.Date)
		}
		if dateText == "" {
			continue
		}
		tradeDate, err := parseFlexibleDateTime(dateText)
		if err != nil {
			continue
		}
		closePrice := item.ClosePrice
		if closePrice <= 0 {
			continue
		}
		openPrice := item.OpenPrice
		if openPrice <= 0 {
			openPrice = closePrice
		}
		highPrice := item.HighPrice
		if highPrice <= 0 {
			highPrice = math.Max(openPrice, closePrice)
		}
		lowPrice := item.LowPrice
		if lowPrice <= 0 {
			lowPrice = math.Min(openPrice, closePrice)
		}
		prevClose := item.PrevClosePrice
		if prevClose <= 0 {
			prevClose = openPrice
		}
		volume := int64(math.Round(item.Volume))
		if volume < 0 {
			volume = 0
		}
		turnover := item.Turnover
		if turnover <= 0 && volume > 0 {
			turnover = closePrice * float64(volume)
		}
		result = append(result, model.StockMarketQuote{
			Symbol:         symbol,
			TradeDate:      tradeDate.Format("2006-01-02"),
			OpenPrice:      roundTo(openPrice, 4),
			HighPrice:      roundTo(highPrice, 4),
			LowPrice:       roundTo(lowPrice, 4),
			ClosePrice:     roundTo(closePrice, 4),
			PrevClosePrice: roundTo(prevClose, 4),
			Volume:         volume,
			Turnover:       roundTo(turnover, 2),
			SourceKey:      sourceKey,
		})
	}
	return result, nil
}

var tushareAPIEndpoint = "https://api.tushare.pro"

var (
	tushareThrottleMu sync.Mutex
	tushareLastCallAt = map[string]time.Time{}

	tushareMinIntervalDefault = 150 * time.Millisecond
	tushareMinIntervalByAPI   = map[string]time.Duration{
		"stock_basic": 220 * time.Millisecond,
		"daily":       150 * time.Millisecond,
		"daily_basic": 180 * time.Millisecond,
		"moneyflow":   180 * time.Millisecond,
		"anns_d":      220 * time.Millisecond,
	}

	tushareMaxRetryAttempts = 4
	tushareRetryBaseDelay   = 700 * time.Millisecond
	tushareRateLimitDelay   = 5 * time.Second
)

func shouldThrottleTushareAPI() bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(tushareAPIEndpoint)), "api.tushare.pro")
}

func tushareMinInterval(apiName string) time.Duration {
	interval := tushareMinIntervalDefault
	if value, ok := tushareMinIntervalByAPI[strings.ToLower(strings.TrimSpace(apiName))]; ok && value > interval {
		interval = value
	}
	return interval
}

func waitForTushareCallSlot(apiName string) {
	if !shouldThrottleTushareAPI() {
		return
	}
	interval := tushareMinInterval(apiName)
	if interval <= 0 {
		return
	}
	key := strings.ToLower(strings.TrimSpace(apiName))
	if key == "" {
		key = "_default"
	}
	for {
		now := time.Now()
		tushareThrottleMu.Lock()
		last := tushareLastCallAt[key]
		wait := interval - now.Sub(last)
		if last.IsZero() || wait <= 0 {
			tushareLastCallAt[key] = now
			tushareThrottleMu.Unlock()
			return
		}
		tushareThrottleMu.Unlock()
		time.Sleep(wait)
	}
}

func isTushareRetryableNetworkError(err error) bool {
	if err == nil {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	text := strings.ToLower(strings.TrimSpace(err.Error()))
	retryableKeywords := []string{
		"eof",
		"timeout",
		"connection reset",
		"broken pipe",
		"tls handshake timeout",
		"temporarily unavailable",
	}
	for _, keyword := range retryableKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func shouldRetryTushareRequest(err error, statusCode int) bool {
	if err == nil {
		return false
	}
	if statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError {
		return true
	}
	if isTushareRateLimitErrorText(err.Error()) {
		return true
	}
	return isTushareRetryableNetworkError(err)
}

func nextTushareRetryDelay(err error, attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if isTushareRateLimitErrorText(err.Error()) {
		return time.Duration(attempt) * tushareRateLimitDelay
	}
	delay := time.Duration(attempt) * tushareRetryBaseDelay
	if delay > 15*time.Second {
		return 15 * time.Second
	}
	return delay
}

func normalizeTushareTradeDateValue(value string, fieldName string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("%s is required", fieldName)
	}
	if strings.Contains(trimmed, "-") {
		parsed, err := time.ParseInLocation("2006-01-02", trimmed, time.Local)
		if err != nil {
			return "", fmt.Errorf("%s must be YYYY-MM-DD or YYYYMMDD", fieldName)
		}
		return parsed.Format("20060102"), nil
	}
	parsed, err := time.ParseInLocation("20060102", trimmed, time.Local)
	if err != nil {
		return "", fmt.Errorf("%s must be YYYY-MM-DD or YYYYMMDD", fieldName)
	}
	return parsed.Format("20060102"), nil
}

func parseStockQuotesFromTushareResponse(sourceKey string, symbols []string, parsed tushareStdResponse) []model.StockMarketQuote {
	if len(parsed.Data.Fields) == 0 || len(parsed.Data.Items) == 0 {
		return nil
	}
	symbolFilter := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbolFilter[strings.ToUpper(strings.TrimSpace(symbol))] = struct{}{}
	}
	fieldIndex := make(map[string]int, len(parsed.Data.Fields))
	for idx, field := range parsed.Data.Fields {
		fieldIndex[strings.TrimSpace(field)] = idx
	}
	result := make([]model.StockMarketQuote, 0, len(parsed.Data.Items))
	for _, row := range parsed.Data.Items {
		tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
		if !ok {
			continue
		}
		tsCode = strings.ToUpper(strings.TrimSpace(tsCode))
		if len(symbolFilter) > 0 {
			if _, exists := symbolFilter[tsCode]; !exists {
				continue
			}
		}
		tradeDateRaw, ok := tushareGetString(row, fieldIndex, "trade_date")
		if !ok {
			continue
		}
		tradeDate, err := time.ParseInLocation("20060102", strings.TrimSpace(tradeDateRaw), time.Local)
		if err != nil {
			continue
		}

		openPrice, ok := tushareGetFloat(row, fieldIndex, "open")
		if !ok {
			continue
		}
		highPrice, ok := tushareGetFloat(row, fieldIndex, "high")
		if !ok {
			continue
		}
		lowPrice, ok := tushareGetFloat(row, fieldIndex, "low")
		if !ok {
			continue
		}
		closePrice, ok := tushareGetFloat(row, fieldIndex, "close")
		if !ok || closePrice <= 0 {
			continue
		}
		prevClose, ok := tushareGetFloat(row, fieldIndex, "pre_close")
		if !ok || prevClose <= 0 {
			prevClose = openPrice
		}
		volLot, _ := tushareGetFloat(row, fieldIndex, "vol")
		amountK, _ := tushareGetFloat(row, fieldIndex, "amount")
		volumeShares := int64(math.Round(volLot * 100))
		if volumeShares < 0 {
			volumeShares = 0
		}
		turnover := amountK * 1000
		if turnover <= 0 && volumeShares > 0 {
			turnover = closePrice * float64(volumeShares)
		}

		result = append(result, model.StockMarketQuote{
			Symbol:         tsCode,
			TradeDate:      tradeDate.Format("2006-01-02"),
			OpenPrice:      roundTo(openPrice, 4),
			HighPrice:      roundTo(highPrice, 4),
			LowPrice:       roundTo(lowPrice, 4),
			ClosePrice:     roundTo(closePrice, 4),
			PrevClosePrice: roundTo(prevClose, 4),
			Volume:         volumeShares,
			Turnover:       roundTo(turnover, 2),
			SourceKey:      sourceKey,
		})
	}
	return result
}

func parseStockDailyBasicsFromTushareResponse(sourceKey string, symbols []string, parsed tushareStdResponse) []stockDailyBasicPoint {
	if len(parsed.Data.Fields) == 0 || len(parsed.Data.Items) == 0 {
		return nil
	}
	symbolFilter := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbolFilter[strings.ToUpper(strings.TrimSpace(symbol))] = struct{}{}
	}
	fieldIndex := make(map[string]int, len(parsed.Data.Fields))
	for idx, field := range parsed.Data.Fields {
		fieldIndex[strings.TrimSpace(field)] = idx
	}
	result := make([]stockDailyBasicPoint, 0, len(parsed.Data.Items))
	for _, row := range parsed.Data.Items {
		tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
		if !ok {
			continue
		}
		tsCode = strings.ToUpper(strings.TrimSpace(tsCode))
		if len(symbolFilter) > 0 {
			if _, exists := symbolFilter[tsCode]; !exists {
				continue
			}
		}
		tradeDateRaw, ok := tushareGetString(row, fieldIndex, "trade_date")
		if !ok {
			continue
		}
		tradeDate, err := time.ParseInLocation("20060102", tradeDateRaw, time.Local)
		if err != nil {
			continue
		}
		turnoverRate, _ := tushareGetFloat(row, fieldIndex, "turnover_rate")
		volumeRatio, _ := tushareGetFloat(row, fieldIndex, "volume_ratio")
		peTTM, _ := tushareGetFloat(row, fieldIndex, "pe_ttm")
		pb, _ := tushareGetFloat(row, fieldIndex, "pb")
		totalMV, _ := tushareGetFloat(row, fieldIndex, "total_mv")
		circMV, _ := tushareGetFloat(row, fieldIndex, "circ_mv")
		result = append(result, stockDailyBasicPoint{
			Symbol:       tsCode,
			TradeDate:    tradeDate,
			TurnoverRate: roundTo(turnoverRate, 4),
			VolumeRatio:  roundTo(volumeRatio, 4),
			PeTTM:        roundTo(peTTM, 4),
			PB:           roundTo(pb, 4),
			TotalMV:      roundTo(totalMV, 4),
			CircMV:       roundTo(circMV, 4),
			SourceKey:    sourceKey,
		})
	}
	return result
}

func parseStockMoneyflowsFromTushareResponse(sourceKey string, symbols []string, parsed tushareStdResponse) []stockMoneyflowPoint {
	if len(parsed.Data.Fields) == 0 || len(parsed.Data.Items) == 0 {
		return nil
	}
	symbolFilter := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbolFilter[strings.ToUpper(strings.TrimSpace(symbol))] = struct{}{}
	}
	fieldIndex := make(map[string]int, len(parsed.Data.Fields))
	for idx, field := range parsed.Data.Fields {
		fieldIndex[strings.TrimSpace(field)] = idx
	}
	result := make([]stockMoneyflowPoint, 0, len(parsed.Data.Items))
	for _, row := range parsed.Data.Items {
		tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
		if !ok {
			continue
		}
		tsCode = strings.ToUpper(strings.TrimSpace(tsCode))
		if len(symbolFilter) > 0 {
			if _, exists := symbolFilter[tsCode]; !exists {
				continue
			}
		}
		tradeDateRaw, ok := tushareGetString(row, fieldIndex, "trade_date")
		if !ok {
			continue
		}
		tradeDate, err := time.ParseInLocation("20060102", tradeDateRaw, time.Local)
		if err != nil {
			continue
		}
		netMFAmount, _ := tushareGetFloat(row, fieldIndex, "net_mf_amount")
		buyLGAmount, _ := tushareGetFloat(row, fieldIndex, "buy_lg_amount")
		sellLGAmount, _ := tushareGetFloat(row, fieldIndex, "sell_lg_amount")
		buyELGAmount, _ := tushareGetFloat(row, fieldIndex, "buy_elg_amount")
		sellELGAmount, _ := tushareGetFloat(row, fieldIndex, "sell_elg_amount")
		result = append(result, stockMoneyflowPoint{
			Symbol:        tsCode,
			TradeDate:     tradeDate,
			NetMFAmount:   roundTo(netMFAmount, 4),
			BuyLGAmount:   roundTo(buyLGAmount, 4),
			SellLGAmount:  roundTo(sellLGAmount, 4),
			BuyELGAmount:  roundTo(buyELGAmount, 4),
			SellELGAmount: roundTo(sellELGAmount, 4),
			SourceKey:     sourceKey,
		})
	}
	return result
}

// estimateTushareCalendarDays calculates the approximate number of calendar days
// between two YYYYMMDD date strings. Used to decide whether to loop by symbol or
// by trade date when calling Tushare APIs, minimizing total API call count.
func estimateTushareCalendarDays(startDate string, endDate string) int {
	start, err := time.ParseInLocation("20060102", strings.TrimSpace(startDate), time.Local)
	if err != nil {
		return 1
	}
	end, err := time.ParseInLocation("20060102", strings.TrimSpace(endDate), time.Local)
	if err != nil {
		return 1
	}
	days := int(end.Sub(start).Hours()/24) + 1
	if days <= 0 {
		return 1
	}
	return days
}

func fetchStockQuotesFromTushare(token string, sourceKey string, symbols []string, days int, timeoutMS int) ([]model.StockMarketQuote, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 120
	}
	startDate := time.Now().AddDate(0, 0, -(days + 20)).Format("20060102")
	endDate := time.Now().Format("20060102")
	return fetchStockQuotesFromTushareDateRange(token, sourceKey, symbols, startDate, endDate, timeoutMS)
}

func fetchStockQuotesFromTushareDateRange(token string, sourceKey string, symbols []string, startDate string, endDate string, timeoutMS int) ([]model.StockMarketQuote, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	normalizedStartDate, err := normalizeTushareTradeDateValue(startDate, "start_date")
	if err != nil {
		return nil, err
	}
	normalizedEndDate, err := normalizeTushareTradeDateValue(endDate, "end_date")
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}

	// Smart routing: choose the loop dimension that minimizes total API calls.
	// By-date mode: 1 API call per calendar day → efficient when many symbols, few days (e.g. daily incremental sync)
	// By-symbol mode: 1 API call per symbol → efficient when few symbols, many days (e.g. historical backfill)
	estimatedDays := estimateTushareCalendarDays(normalizedStartDate, normalizedEndDate)
	if len(symbols) > 1 && len(symbols) > estimatedDays {
		return fetchStockQuotesFromTushareByTradeDates(client, token, sourceKey, symbols, normalizedStartDate, normalizedEndDate)
	}

	result := make([]model.StockMarketQuote, 0, len(symbols)*maxInt(estimatedDays, 1))

	for _, symbol := range symbols {
		parsed, err := callTushareAPI(client, token, "daily", map[string]string{
			"ts_code":    strings.TrimSpace(symbol),
			"start_date": normalizedStartDate,
			"end_date":   normalizedEndDate,
		}, "ts_code,trade_date,open,high,low,close,pre_close,vol,amount")
		if err != nil {
			return nil, err
		}
		result = append(result, parseStockQuotesFromTushareResponse(sourceKey, []string{symbol}, parsed)...)
	}
	return result, nil
}

func fetchStockQuotesFromTushareByTradeDates(client *http.Client, token string, sourceKey string, symbols []string, startDate string, endDate string) ([]model.StockMarketQuote, error) {
	start, err := time.ParseInLocation("20060102", startDate, time.Local)
	if err != nil {
		return nil, err
	}
	end, err := time.ParseInLocation("20060102", endDate, time.Local)
	if err != nil {
		return nil, err
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end_date must be >= start_date")
	}

	result := make([]model.StockMarketQuote, 0, len(symbols)*int(end.Sub(start).Hours()/24+1))
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		tradeDate := current.Format("20060102")
		parsed, err := callTushareAPI(client, token, "daily", map[string]string{
			"trade_date": tradeDate,
		}, "ts_code,trade_date,open,high,low,close,pre_close,vol,amount")
		if err != nil {
			return nil, err
		}
		result = append(result, parseStockQuotesFromTushareResponse(sourceKey, symbols, parsed)...)
	}
	return result, nil
}

type tushareStdResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Fields []string        `json:"fields"`
		Items  [][]interface{} `json:"items"`
	} `json:"data"`
}

func callTushareAPI(client *http.Client, token string, apiName string, params map[string]string, fields string) (tushareStdResponse, error) {
	request := map[string]interface{}{
		"api_name": apiName,
		"token":    token,
		"params":   params,
		"fields":   fields,
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return tushareStdResponse{}, err
	}
	attempts := tushareMaxRetryAttempts
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		waitForTushareCallSlot(apiName)
		req, reqErr := http.NewRequest(http.MethodPost, tushareAPIEndpoint, bytes.NewReader(payload))
		if reqErr != nil {
			return tushareStdResponse{}, reqErr
		}
		req.Header.Set("Content-Type", "application/json")
		resp, doErr := client.Do(req)
		if doErr != nil {
			lastErr = doErr
			if attempt < attempts && shouldRetryTushareRequest(doErr, 0) {
				time.Sleep(nextTushareRetryDelay(doErr, attempt))
				continue
			}
			return tushareStdResponse{}, doErr
		}
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			if attempt < attempts && shouldRetryTushareRequest(readErr, 0) {
				time.Sleep(nextTushareRetryDelay(readErr, attempt))
				continue
			}
			return tushareStdResponse{}, readErr
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			statusErr := fmt.Errorf("tushare status: %s", resp.Status)
			lastErr = statusErr
			if attempt < attempts && shouldRetryTushareRequest(statusErr, resp.StatusCode) {
				time.Sleep(nextTushareRetryDelay(statusErr, attempt))
				continue
			}
			return tushareStdResponse{}, statusErr
		}
		parsed := tushareStdResponse{}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return tushareStdResponse{}, err
		}
		if parsed.Code != 0 {
			apiErr := fmt.Errorf("tushare error(%s): %s", apiName, strings.TrimSpace(parsed.Msg))
			lastErr = apiErr
			if attempt < attempts && shouldRetryTushareRequest(apiErr, 0) {
				time.Sleep(nextTushareRetryDelay(apiErr, attempt))
				continue
			}
			return tushareStdResponse{}, apiErr
		}
		return parsed, nil
	}
	if lastErr == nil {
		lastErr = errors.New("tushare request failed")
	}
	return tushareStdResponse{}, lastErr
}

func fetchStockDailyBasicsFromTushare(token string, sourceKey string, symbols []string, days int, timeoutMS int) ([]stockDailyBasicPoint, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 120
	}
	startDate := time.Now().AddDate(0, 0, -(days + 20)).Format("20060102")
	endDate := time.Now().Format("20060102")
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}

	// Smart routing: by-date when many symbols over few days; by-symbol when few symbols over many days
	estimatedDays := estimateTushareCalendarDays(startDate, endDate)
	if len(symbols) > 1 && len(symbols) > estimatedDays {
		return fetchStockDailyBasicsFromTushareByTradeDates(client, token, sourceKey, symbols, startDate, endDate)
	}

	result := make([]stockDailyBasicPoint, 0, len(symbols)*maxInt(estimatedDays, 1))
	for _, symbol := range symbols {
		parsed, err := callTushareAPI(client, token, "daily_basic", map[string]string{
			"ts_code":    symbol,
			"start_date": startDate,
			"end_date":   endDate,
		}, "ts_code,trade_date,turnover_rate,volume_ratio,pe_ttm,pb,total_mv,circ_mv")
		if err != nil {
			return nil, err
		}
		result = append(result, parseStockDailyBasicsFromTushareResponse(sourceKey, []string{symbol}, parsed)...)
	}
	return result, nil
}

func fetchStockMoneyflowsFromTushare(token string, sourceKey string, symbols []string, days int, timeoutMS int) ([]stockMoneyflowPoint, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 120
	}
	startDate := time.Now().AddDate(0, 0, -(days + 20)).Format("20060102")
	endDate := time.Now().Format("20060102")
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}

	// Smart routing: by-date when many symbols over few days; by-symbol when few symbols over many days
	estimatedDays := estimateTushareCalendarDays(startDate, endDate)
	if len(symbols) > 1 && len(symbols) > estimatedDays {
		return fetchStockMoneyflowsFromTushareByTradeDates(client, token, sourceKey, symbols, startDate, endDate)
	}

	result := make([]stockMoneyflowPoint, 0, len(symbols)*maxInt(estimatedDays, 1))
	for _, symbol := range symbols {
		parsed, err := callTushareAPI(client, token, "moneyflow", map[string]string{
			"ts_code":    symbol,
			"start_date": startDate,
			"end_date":   endDate,
		}, "ts_code,trade_date,net_mf_amount,buy_lg_amount,sell_lg_amount,buy_elg_amount,sell_elg_amount")
		if err != nil {
			return nil, err
		}
		result = append(result, parseStockMoneyflowsFromTushareResponse(sourceKey, []string{symbol}, parsed)...)
	}
	return result, nil
}

func fetchStockDailyBasicsFromTushareByTradeDates(client *http.Client, token string, sourceKey string, symbols []string, startDate string, endDate string) ([]stockDailyBasicPoint, error) {
	start, err := time.ParseInLocation("20060102", startDate, time.Local)
	if err != nil {
		return nil, err
	}
	end, err := time.ParseInLocation("20060102", endDate, time.Local)
	if err != nil {
		return nil, err
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end_date must be >= start_date")
	}

	result := make([]stockDailyBasicPoint, 0, len(symbols)*int(end.Sub(start).Hours()/24+1))
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		tradeDate := current.Format("20060102")
		parsed, err := callTushareAPI(client, token, "daily_basic", map[string]string{
			"trade_date": tradeDate,
		}, "ts_code,trade_date,turnover_rate,volume_ratio,pe_ttm,pb,total_mv,circ_mv")
		if err != nil {
			return nil, err
		}
		result = append(result, parseStockDailyBasicsFromTushareResponse(sourceKey, symbols, parsed)...)
	}
	return result, nil
}

func fetchStockMoneyflowsFromTushareByTradeDates(client *http.Client, token string, sourceKey string, symbols []string, startDate string, endDate string) ([]stockMoneyflowPoint, error) {
	start, err := time.ParseInLocation("20060102", startDate, time.Local)
	if err != nil {
		return nil, err
	}
	end, err := time.ParseInLocation("20060102", endDate, time.Local)
	if err != nil {
		return nil, err
	}
	if end.Before(start) {
		return nil, fmt.Errorf("end_date must be >= start_date")
	}

	result := make([]stockMoneyflowPoint, 0, len(symbols)*int(end.Sub(start).Hours()/24+1))
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		tradeDate := current.Format("20060102")
		parsed, err := callTushareAPI(client, token, "moneyflow", map[string]string{
			"trade_date": tradeDate,
		}, "ts_code,trade_date,net_mf_amount,buy_lg_amount,sell_lg_amount,buy_elg_amount,sell_elg_amount")
		if err != nil {
			return nil, err
		}
		result = append(result, parseStockMoneyflowsFromTushareResponse(sourceKey, symbols, parsed)...)
	}
	return result, nil
}

// fetchStockKPLListFromTushare fetches daily limit-up board (kpl_list) data from Tushare.
func fetchStockKPLListFromTushare(token string, sourceKey string, days int, timeoutMS int) ([]stockLimitUpPoint, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 5
	}
	startDate := time.Now().AddDate(0, 0, -days).Format("20060102")
	endDate := time.Now().Format("20060102")
	start, err := time.ParseInLocation("20060102", startDate, time.Local)
	if err != nil {
		return nil, err
	}
	end, err := time.ParseInLocation("20060102", endDate, time.Local)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	result := make([]stockLimitUpPoint, 0, 200)
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		tradeDate := current.Format("20060102")
		parsed, err := callTushareAPI(client, token, "kpl_list", map[string]string{
			"trade_date": tradeDate,
		}, "ts_code,name,trade_date,lu_time,open_time,last_time,tag,theme,status,limit_order,lu_limit_order,bid_amount,bid_change,lu_desc,pct_chg,close,amount,float_mv,turnover_rate")
		if err != nil {
			return nil, err
		}
		result = append(result, parseStockKPLListFromTushareResponse(sourceKey, tradeDate, parsed)...)
	}
	return result, nil
}

// parseStockKPLListFromTushareResponse parses kpl_list Tushare response into stockLimitUpPoint.
func parseStockKPLListFromTushareResponse(sourceKey string, tradeDate string, parsed tushareStdResponse) []stockLimitUpPoint {
	if len(parsed.Data.Fields) == 0 || len(parsed.Data.Items) == 0 {
		return nil
	}
	fieldIndex := make(map[string]int, len(parsed.Data.Fields))
	for idx, field := range parsed.Data.Fields {
		fieldIndex[strings.TrimSpace(field)] = idx
	}
	result := make([]stockLimitUpPoint, 0, len(parsed.Data.Items))
	parsedDate, dateErr := time.ParseInLocation("20060102", tradeDate, time.Local)
	for _, row := range parsed.Data.Items {
		tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
		if !ok {
			continue
		}
		tsCode = strings.ToUpper(strings.TrimSpace(tsCode))
		tradeDateEffective := parsedDate
		if dateErr != nil {
			if td, ok := tushareGetString(row, fieldIndex, "trade_date"); ok {
				if parsedDate2, err := time.ParseInLocation("20060102", td, time.Local); err == nil {
					tradeDateEffective = parsedDate2
				}
			}
		}
		name, _ := tushareGetString(row, fieldIndex, "name")
		luTime, _ := tushareGetString(row, fieldIndex, "lu_time")
		openTime, _ := tushareGetString(row, fieldIndex, "open_time")
		lastTime, _ := tushareGetString(row, fieldIndex, "last_time")
		tag, _ := tushareGetString(row, fieldIndex, "tag")
		theme, _ := tushareGetString(row, fieldIndex, "theme")
		status, _ := tushareGetString(row, fieldIndex, "status")
		limitOrder, _ := tushareGetFloat(row, fieldIndex, "limit_order")
		luLimitOrder, _ := tushareGetFloat(row, fieldIndex, "lu_limit_order")
		bidAmount, _ := tushareGetFloat(row, fieldIndex, "bid_amount")
		bidChange, _ := tushareGetFloat(row, fieldIndex, "bid_change")
		pctChg, _ := tushareGetFloat(row, fieldIndex, "pct_chg")
		floatMV, _ := tushareGetFloat(row, fieldIndex, "float_mv")
		result = append(result, stockLimitUpPoint{
			Symbol:       tsCode,
			TradeDate:    tradeDateEffective,
			Name:         name,
			LuTime:       luTime,
			OpenTime:     openTime,
			LastTime:     lastTime,
			Tag:          tag,
			Theme:        theme,
			Status:       status,
			LimitOrder:   roundTo(limitOrder, 4),
			LuLimitOrder: roundTo(luLimitOrder, 4),
			BidAmount:    roundTo(bidAmount, 4),
			BidChange:    roundTo(bidChange, 4),
			FloatMV:      roundTo(floatMV, 4),
			PctChg:       roundTo(pctChg, 4),
		})
	}
	return result
}

// fetchStockTopListFromTushare fetches daily dragon-tiger board (top_list) data from Tushare.
func fetchStockTopListFromTushare(token string, sourceKey string, days int, timeoutMS int) ([]stockTopListPoint, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 5
	}
	startDate := time.Now().AddDate(0, 0, -days).Format("20060102")
	endDate := time.Now().Format("20060102")
	start, err := time.ParseInLocation("20060102", startDate, time.Local)
	if err != nil {
		return nil, err
	}
	end, err := time.ParseInLocation("20060102", endDate, time.Local)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	result := make([]stockTopListPoint, 0, 200)
	for current := start; !current.After(end); current = current.AddDate(0, 0, 1) {
		tradeDate := current.Format("20060102")
		parsed, err := callTushareAPI(client, token, "top_list", map[string]string{
			"trade_date": tradeDate,
		}, "ts_code,name,trade_date,close,pct_chg,amount,buy_amount,sell_amount,net_amount,reason")
		if err != nil {
			return nil, err
		}
		result = append(result, parseStockTopListFromTushareResponse(sourceKey, tradeDate, parsed)...)
	}
	return result, nil
}

// parseStockTopListFromTushareResponse parses top_list Tushare response into stockTopListPoint.
func parseStockTopListFromTushareResponse(sourceKey string, tradeDate string, parsed tushareStdResponse) []stockTopListPoint {
	if len(parsed.Data.Fields) == 0 || len(parsed.Data.Items) == 0 {
		return nil
	}
	fieldIndex := make(map[string]int, len(parsed.Data.Fields))
	for idx, field := range parsed.Data.Fields {
		fieldIndex[strings.TrimSpace(field)] = idx
	}
	result := make([]stockTopListPoint, 0, len(parsed.Data.Items))
	parsedDate, dateErr := time.ParseInLocation("20060102", tradeDate, time.Local)
	for _, row := range parsed.Data.Items {
		tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
		if !ok {
			continue
		}
		tsCode = strings.ToUpper(strings.TrimSpace(tsCode))
		tradeDateEffective := parsedDate
		if dateErr != nil {
			if td, ok := tushareGetString(row, fieldIndex, "trade_date"); ok {
				if parsedDate2, err := time.ParseInLocation("20060102", td, time.Local); err == nil {
					tradeDateEffective = parsedDate2
				}
			}
		}
		name, _ := tushareGetString(row, fieldIndex, "name")
		buyAmount, _ := tushareGetFloat(row, fieldIndex, "buy_amount")
		sellAmount, _ := tushareGetFloat(row, fieldIndex, "sell_amount")
		netAmount, _ := tushareGetFloat(row, fieldIndex, "net_amount")
		reason, _ := tushareGetString(row, fieldIndex, "reason")
		result = append(result, stockTopListPoint{
			Symbol:     tsCode,
			TradeDate:  tradeDateEffective,
			Name:       name,
			BuyAmount:  roundTo(buyAmount, 4),
			SellAmount: roundTo(sellAmount, 4),
			NetAmount:  roundTo(netAmount, 4),
			Reason:     reason,
		})
	}
	return result
}

func fetchStockNewsFromTushare(token string, sourceKey string, symbols []string, days int, timeoutMS int) ([]stockNewsRawPoint, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 14
	}
	startDate := time.Now().AddDate(0, 0, -days).Format("20060102")
	endDate := time.Now().Format("20060102")
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	result := make([]stockNewsRawPoint, 0, len(symbols)*3)
	for _, symbol := range symbols {
		parsed, err := callTushareAPI(client, token, "anns_d", map[string]string{
			"ts_code":    symbol,
			"start_date": startDate,
			"end_date":   endDate,
		}, "ts_code,ann_date,title,url,rec_time")
		if err != nil {
			return nil, err
		}
		fieldIndex := make(map[string]int, len(parsed.Data.Fields))
		for idx, field := range parsed.Data.Fields {
			fieldIndex[strings.TrimSpace(field)] = idx
		}
		for _, row := range parsed.Data.Items {
			tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
			if !ok {
				continue
			}
			title, ok := tushareGetString(row, fieldIndex, "title")
			if !ok {
				continue
			}
			urlText, _ := tushareGetString(row, fieldIndex, "url")
			recTime, _ := tushareGetString(row, fieldIndex, "rec_time")
			annDate, _ := tushareGetString(row, fieldIndex, "ann_date")
			publishedAt := parseTushareNewsTime(recTime, annDate)
			sentiment := classifyNewsSentiment(title)
			result = append(result, stockNewsRawPoint{
				SourceKey:   sourceKey,
				Symbol:      strings.ToUpper(strings.TrimSpace(tsCode)),
				PublishedAt: publishedAt,
				Title:       strings.TrimSpace(title),
				URL:         strings.TrimSpace(urlText),
				Sentiment:   sentiment,
			})
		}
	}
	return result, nil
}

func parseTushareNewsTime(recTime string, annDate string) time.Time {
	recTime = strings.TrimSpace(recTime)
	annDate = strings.TrimSpace(annDate)
	if recTime != "" {
		if ts, err := time.ParseInLocation("2006-01-02 15:04:05", recTime, time.Local); err == nil {
			return ts
		}
		if ts, err := time.ParseInLocation("2006-01-02 15:04", recTime, time.Local); err == nil {
			return ts
		}
	}
	if annDate != "" {
		if ts, err := time.ParseInLocation("20060102", annDate, time.Local); err == nil {
			return ts
		}
	}
	return time.Now()
}

func classifyNewsSentiment(title string) string {
	text := strings.ToLower(strings.TrimSpace(title))
	if text == "" {
		return "NEUTRAL"
	}
	positiveKeywords := []string{"增长", "预增", "中标", "回购", "增持", "突破", "创新高", "上调", "盈利", "签约", "positive", "upgrade"}
	negativeKeywords := []string{"下滑", "预减", "亏损", "违约", "减持", "处罚", "风险", "暴跌", "下调", "诉讼", "negative", "downgrade"}
	positiveHit := 0
	negativeHit := 0
	for _, keyword := range positiveKeywords {
		if strings.Contains(text, keyword) {
			positiveHit++
		}
	}
	for _, keyword := range negativeKeywords {
		if strings.Contains(text, keyword) {
			negativeHit++
		}
	}
	if positiveHit > negativeHit {
		return "POSITIVE"
	}
	if negativeHit > positiveHit {
		return "NEGATIVE"
	}
	return "NEUTRAL"
}

func tushareGetString(row []interface{}, fieldIndex map[string]int, key string) (string, bool) {
	index, ok := fieldIndex[key]
	if !ok || index < 0 || index >= len(row) {
		return "", false
	}
	text := strings.TrimSpace(fmt.Sprintf("%v", row[index]))
	if text == "" || text == "<nil>" {
		return "", false
	}
	return text, true
}

func tushareGetFloat(row []interface{}, fieldIndex map[string]int, key string) (float64, bool) {
	index, ok := fieldIndex[key]
	if !ok || index < 0 || index >= len(row) {
		return 0, false
	}
	value := row[index]
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case json.Number:
		num, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return num, true
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return 0, false
		}
		num, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return 0, false
		}
		return num, true
	default:
		text := strings.TrimSpace(fmt.Sprintf("%v", value))
		if text == "" || text == "<nil>" {
			return 0, false
		}
		num, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return 0, false
		}
		return num, true
	}
}

func (r *MySQLGrowthRepo) upsertStockMarketQuotes(items []model.StockMarketQuote) (int, error) {
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		tradeDate, err := parseFlexibleDateTime(item.TradeDate)
		if err != nil {
			return affected, err
		}
		if item.ClosePrice <= 0 {
			continue
		}
		openPrice := item.OpenPrice
		if openPrice <= 0 {
			openPrice = item.ClosePrice
		}
		highPrice := item.HighPrice
		if highPrice <= 0 {
			highPrice = math.Max(openPrice, item.ClosePrice)
		}
		lowPrice := item.LowPrice
		if lowPrice <= 0 {
			lowPrice = math.Min(openPrice, item.ClosePrice)
		}
		prevClose := item.PrevClosePrice
		if prevClose <= 0 {
			prevClose = openPrice
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if sourceKey == "" {
			sourceKey = "MOCK"
		}
		_, err = r.db.Exec(`
INSERT INTO stock_market_quotes
  (id, symbol, trade_date, open_price, high_price, low_price, close_price, prev_close_price, volume, turnover, source_key, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  open_price = VALUES(open_price),
  high_price = VALUES(high_price),
  low_price = VALUES(low_price),
  close_price = VALUES(close_price),
  prev_close_price = VALUES(prev_close_price),
  volume = VALUES(volume),
  turnover = VALUES(turnover),
  source_key = VALUES(source_key),
  updated_at = VALUES(updated_at)`,
			newID("smq"),
			symbol,
			tradeDate.Format("2006-01-02"),
			openPrice,
			highPrice,
			lowPrice,
			item.ClosePrice,
			prevClose,
			item.Volume,
			item.Turnover,
			sourceKey,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) upsertStockDailyBasics(items []stockDailyBasicPoint) (int, error) {
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if sourceKey == "" {
			sourceKey = "TUSHARE"
		}
		_, err := r.db.Exec(`
INSERT INTO stock_daily_basic
  (id, symbol, trade_date, turnover_rate, volume_ratio, pe_ttm, pb, total_mv, circ_mv, source_key, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  turnover_rate = VALUES(turnover_rate),
  volume_ratio = VALUES(volume_ratio),
  pe_ttm = VALUES(pe_ttm),
  pb = VALUES(pb),
  total_mv = VALUES(total_mv),
  circ_mv = VALUES(circ_mv),
  source_key = VALUES(source_key),
  updated_at = VALUES(updated_at)`,
			newID("sdb"),
			symbol,
			item.TradeDate.Format("2006-01-02"),
			item.TurnoverRate,
			item.VolumeRatio,
			item.PeTTM,
			item.PB,
			item.TotalMV,
			item.CircMV,
			sourceKey,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) upsertStockMoneyflows(items []stockMoneyflowPoint) (int, error) {
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if sourceKey == "" {
			sourceKey = "TUSHARE"
		}
		_, err := r.db.Exec(`
INSERT INTO stock_moneyflow_daily
  (id, symbol, trade_date, net_mf_amount, buy_lg_amount, sell_lg_amount, buy_elg_amount, sell_elg_amount, source_key, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  net_mf_amount = VALUES(net_mf_amount),
  buy_lg_amount = VALUES(buy_lg_amount),
  sell_lg_amount = VALUES(sell_lg_amount),
  buy_elg_amount = VALUES(buy_elg_amount),
  sell_elg_amount = VALUES(sell_elg_amount),
  source_key = VALUES(source_key),
  updated_at = VALUES(updated_at)`,
			newID("smf"),
			symbol,
			item.TradeDate.Format("2006-01-02"),
			item.NetMFAmount,
			item.BuyLGAmount,
			item.SellLGAmount,
			item.BuyELGAmount,
			item.SellELGAmount,
			sourceKey,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) upsertStockNewsRaw(items []stockNewsRawPoint) (int, error) {
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		title := strings.TrimSpace(item.Title)
		if symbol == "" || title == "" || item.PublishedAt.IsZero() {
			continue
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if sourceKey == "" {
			sourceKey = "TUSHARE"
		}
		sentiment := strings.ToUpper(strings.TrimSpace(item.Sentiment))
		if sentiment == "" {
			sentiment = "NEUTRAL"
		}
		_, err := r.db.Exec(`
INSERT INTO stock_news_raw
  (id, source_key, symbol, published_at, title, content, url, sentiment, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  content = VALUES(content),
  url = VALUES(url),
  sentiment = VALUES(sentiment),
  updated_at = VALUES(updated_at)`,
			newID("snr"),
			sourceKey,
			symbol,
			item.PublishedAt,
			title,
			nullableString(item.Content),
			nullableString(item.URL),
			sentiment,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) upsertStockKPLList(items []stockLimitUpPoint) (int, error) {
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		_, err := r.db.Exec(`
INSERT INTO stock_limit_up_daily
  (id, trade_date, ts_code, name, lu_time, open_time, last_time, tag, theme, status,
   limit_order, lu_limit_order, bid_amount, bid_change, pct_chg, float_mv,
   created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  lu_time = VALUES(lu_time),
  open_time = VALUES(open_time),
  last_time = VALUES(last_time),
  tag = VALUES(tag),
  theme = VALUES(theme),
  status = VALUES(status),
  limit_order = VALUES(limit_order),
  lu_limit_order = VALUES(lu_limit_order),
  bid_amount = VALUES(bid_amount),
  bid_change = VALUES(bid_change),
  pct_chg = VALUES(pct_chg),
  float_mv = VALUES(float_mv),
  updated_at = VALUES(updated_at)`,
			newID("slu"),
			item.TradeDate.Format("2006-01-02"),
			symbol,
			item.Name,
			item.LuTime,
			item.OpenTime,
			item.LastTime,
			item.Tag,
			item.Theme,
			item.Status,
			item.LimitOrder,
			item.LuLimitOrder,
			item.BidAmount,
			item.BidChange,
			item.PctChg,
			item.FloatMV,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) upsertStockTopList(items []stockTopListPoint) (int, error) {
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		_, err := r.db.Exec(`
INSERT INTO stock_top_list_daily
  (id, trade_date, ts_code, name, buy_amount, sell_amount, net_amount, reason,
   created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  buy_amount = VALUES(buy_amount),
  sell_amount = VALUES(sell_amount),
  net_amount = VALUES(net_amount),
  reason = VALUES(reason),
  updated_at = VALUES(updated_at)`,
			newID("stl"),
			item.TradeDate.Format("2006-01-02"),
			symbol,
			item.Name,
			item.BuyAmount,
			item.SellAmount,
			item.NetAmount,
			item.Reason,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) loadLatestStockDailyBasics(sinceDate time.Time) (map[string]stockDailyBasicPoint, error) {
	rows, err := r.db.Query(`
SELECT t.symbol, t.trade_date, t.turnover_rate, t.volume_ratio, t.pe_ttm, t.pb, t.total_mv, t.circ_mv, t.source_key
FROM stock_daily_basic t
INNER JOIN (
  SELECT symbol, MAX(trade_date) AS latest_trade_date
  FROM stock_daily_basic
  WHERE trade_date >= ?
  GROUP BY symbol
) latest
ON latest.symbol = t.symbol AND latest.latest_trade_date = t.trade_date`, sinceDate.Format("2006-01-02"))
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]stockDailyBasicPoint{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]stockDailyBasicPoint)
	for rows.Next() {
		var (
			item         stockDailyBasicPoint
			turnoverRate sql.NullFloat64
			volumeRatio  sql.NullFloat64
			peTTM        sql.NullFloat64
			pb           sql.NullFloat64
			totalMV      sql.NullFloat64
			circMV       sql.NullFloat64
			sourceKey    sql.NullString
		)
		if err := rows.Scan(
			&item.Symbol,
			&item.TradeDate,
			&turnoverRate,
			&volumeRatio,
			&peTTM,
			&pb,
			&totalMV,
			&circMV,
			&sourceKey,
		); err != nil {
			return nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
		item.TurnoverRate = sqlNullFloat(turnoverRate)
		item.VolumeRatio = sqlNullFloat(volumeRatio)
		item.PeTTM = sqlNullFloat(peTTM)
		item.PB = sqlNullFloat(pb)
		item.TotalMV = sqlNullFloat(totalMV)
		item.CircMV = sqlNullFloat(circMV)
		if sourceKey.Valid {
			item.SourceKey = strings.ToUpper(strings.TrimSpace(sourceKey.String))
		}
		if item.Symbol == "" {
			continue
		}
		result[item.Symbol] = item
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadLatestStockMoneyflows(sinceDate time.Time) (map[string]stockMoneyflowPoint, error) {
	rows, err := r.db.Query(`
SELECT t.symbol, t.trade_date, t.net_mf_amount, t.buy_lg_amount, t.sell_lg_amount, t.buy_elg_amount, t.sell_elg_amount, t.source_key
FROM stock_moneyflow_daily t
INNER JOIN (
  SELECT symbol, MAX(trade_date) AS latest_trade_date
  FROM stock_moneyflow_daily
  WHERE trade_date >= ?
  GROUP BY symbol
) latest
ON latest.symbol = t.symbol AND latest.latest_trade_date = t.trade_date`, sinceDate.Format("2006-01-02"))
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]stockMoneyflowPoint{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]stockMoneyflowPoint)
	for rows.Next() {
		var (
			item      stockMoneyflowPoint
			netMF     sql.NullFloat64
			buyLG     sql.NullFloat64
			sellLG    sql.NullFloat64
			buyELG    sql.NullFloat64
			sellELG   sql.NullFloat64
			sourceKey sql.NullString
		)
		if err := rows.Scan(
			&item.Symbol,
			&item.TradeDate,
			&netMF,
			&buyLG,
			&sellLG,
			&buyELG,
			&sellELG,
			&sourceKey,
		); err != nil {
			return nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
		item.NetMFAmount = sqlNullFloat(netMF)
		item.BuyLGAmount = sqlNullFloat(buyLG)
		item.SellLGAmount = sqlNullFloat(sellLG)
		item.BuyELGAmount = sqlNullFloat(buyELG)
		item.SellELGAmount = sqlNullFloat(sellELG)
		if sourceKey.Valid {
			item.SourceKey = strings.ToUpper(strings.TrimSpace(sourceKey.String))
		}
		if item.Symbol == "" {
			continue
		}
		result[item.Symbol] = item
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadStockNewsSignals(sinceTime time.Time) (map[string]stockNewsSignal, error) {
	rows, err := r.db.Query(`
SELECT symbol,
       COUNT(*) AS heat,
       SUM(CASE WHEN sentiment = 'POSITIVE' THEN 1 ELSE 0 END) AS positive_cnt
FROM stock_news_raw
WHERE published_at >= ?
GROUP BY symbol`, sinceTime)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]stockNewsSignal{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]stockNewsSignal)
	for rows.Next() {
		var (
			symbol      string
			heat        int
			positiveCnt int
		)
		if err := rows.Scan(&symbol, &heat, &positiveCnt); err != nil {
			return nil, err
		}
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" || heat <= 0 {
			continue
		}
		result[symbol] = stockNewsSignal{
			Heat:         heat,
			PositiveRate: float64(positiveCnt) / float64(heat),
		}
	}
	return result, nil
}

func (r *MySQLGrowthRepo) persistStockQuantSnapshots(items []model.StockQuantScore) error {
	if len(items) == 0 {
		return nil
	}
	now := time.Now()
	for _, item := range items {
		tradeDate, err := parseFlexibleDateTime(item.TradeDate)
		if err != nil {
			continue
		}
		reasonsJSON, _ := json.Marshal(item.Reasons)
		payloadJSON, _ := json.Marshal(item)
		_, err = r.db.Exec(`
INSERT INTO stock_factor_snapshot
  (id, symbol, trade_date, total_score, trend_score, flow_score, value_score, news_score, risk_level, reason_summary, reasons_json, source_key, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  total_score = VALUES(total_score),
  trend_score = VALUES(trend_score),
  flow_score = VALUES(flow_score),
  value_score = VALUES(value_score),
  news_score = VALUES(news_score),
  risk_level = VALUES(risk_level),
  reason_summary = VALUES(reason_summary),
  reasons_json = VALUES(reasons_json),
  source_key = VALUES(source_key),
  updated_at = VALUES(updated_at)`,
			newID("sfs"),
			item.Symbol,
			tradeDate.Format("2006-01-02"),
			item.Score,
			item.TrendScore,
			item.FlowScore,
			item.ValueScore,
			item.NewsScore,
			item.RiskLevel,
			nullableString(item.ReasonSummary),
			string(reasonsJSON),
			"QUANT_V2",
			now,
			now,
		)
		if err != nil {
			return err
		}
		_, err = r.db.Exec(`
INSERT INTO stock_rank_daily
  (id, trade_date, rank_no, symbol, name, total_score, risk_level, reason_summary, payload_json, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  rank_no = VALUES(rank_no),
  name = VALUES(name),
  total_score = VALUES(total_score),
  risk_level = VALUES(risk_level),
  reason_summary = VALUES(reason_summary),
  payload_json = VALUES(payload_json),
  updated_at = VALUES(updated_at)`,
			newID("srd"),
			tradeDate.Format("2006-01-02"),
			item.Rank,
			item.Symbol,
			nullableString(item.Name),
			item.Score,
			item.RiskLevel,
			nullableString(item.ReasonSummary),
			string(payloadJSON),
			now,
			now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) loadStockQuotesBySymbol(sinceDate time.Time) (map[string][]stockQuoteCandle, error) {
	rows, err := r.db.Query(`
SELECT symbol, trade_date, open_price, high_price, low_price, close_price, prev_close_price, volume, turnover
FROM stock_market_quotes
WHERE trade_date >= ?
ORDER BY symbol ASC, trade_date ASC`, sinceDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string][]stockQuoteCandle)
	for rows.Next() {
		var item stockQuoteCandle
		var prevClose sql.NullFloat64
		var turnover sql.NullFloat64
		if err := rows.Scan(
			&item.Symbol,
			&item.TradeDate,
			&item.OpenPrice,
			&item.HighPrice,
			&item.LowPrice,
			&item.ClosePrice,
			&prevClose,
			&item.Volume,
			&turnover,
		); err != nil {
			return nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
		if prevClose.Valid {
			item.PrevClosePrice = prevClose.Float64
		}
		if turnover.Valid {
			item.Turnover = turnover.Float64
		}
		if item.Symbol == "" || item.ClosePrice <= 0 {
			continue
		}
		result[item.Symbol] = append(result[item.Symbol], item)
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadBenchmarkQuoteSeries(sinceDate time.Time) ([]stockQuoteCandle, error) {
	benchmarkSymbols := []string{
		"000300.SH",
		"399001.SZ",
		"000001.SH",
		"399006.SZ",
	}
	symbols := normalizeStockSymbolList(benchmarkSymbols)
	placeholder := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	args := make([]interface{}, 0, len(symbols)+1)
	args = append(args, sinceDate.Format("2006-01-02"))
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	query := fmt.Sprintf(`
SELECT symbol, trade_date, open_price, high_price, low_price, close_price, prev_close_price, volume, turnover
FROM stock_market_quotes
WHERE trade_date >= ? AND symbol IN (%s)
ORDER BY trade_date ASC, symbol ASC`, placeholder)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isTableNotFoundError(err) {
			return []stockQuoteCandle{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	seriesByDate := make(map[string][]float64)
	for rows.Next() {
		var (
			symbol     string
			tradeDate  time.Time
			openPrice  float64
			highPrice  float64
			lowPrice   float64
			closePrice float64
			prevClose  sql.NullFloat64
			volume     sql.NullInt64
			turnover   sql.NullFloat64
		)
		if err := rows.Scan(&symbol, &tradeDate, &openPrice, &highPrice, &lowPrice, &closePrice, &prevClose, &volume, &turnover); err != nil {
			return nil, err
		}
		if closePrice <= 0 {
			continue
		}
		key := tradeDate.Format("2006-01-02")
		seriesByDate[key] = append(seriesByDate[key], closePrice)
	}
	dateKeys := make([]string, 0, len(seriesByDate))
	for date := range seriesByDate {
		dateKeys = append(dateKeys, date)
	}
	sort.Strings(dateKeys)
	result := make([]stockQuoteCandle, 0, len(dateKeys))
	prev := 0.0
	for _, date := range dateKeys {
		values := seriesByDate[date]
		if len(values) == 0 {
			continue
		}
		meanClose := avgFloat(values)
		if meanClose <= 0 {
			continue
		}
		tradeDate, err := time.ParseInLocation("2006-01-02", date, time.Local)
		if err != nil {
			continue
		}
		if prev <= 0 {
			prev = meanClose
		}
		result = append(result, stockQuoteCandle{
			Symbol:         "BENCHMARK",
			TradeDate:      tradeDate,
			OpenPrice:      meanClose,
			HighPrice:      meanClose,
			LowPrice:       meanClose,
			ClosePrice:     meanClose,
			PrevClosePrice: prev,
		})
		prev = meanClose
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadQuantRiskLevelByDateSymbol(sinceDate time.Time, topN int) (map[string]string, error) {
	rows, err := r.db.Query(`
SELECT r.trade_date, r.symbol, COALESCE(NULLIF(s.risk_level, ''), 'MEDIUM') AS risk_level
FROM stock_rank_daily r
LEFT JOIN stock_factor_snapshot s
  ON s.trade_date = r.trade_date AND s.symbol = r.symbol
WHERE r.trade_date >= ? AND r.rank_no <= ?`, sinceDate.Format("2006-01-02"), topN)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]string)
	for rows.Next() {
		var (
			tradeDate time.Time
			symbol    string
			riskLevel sql.NullString
		)
		if err := rows.Scan(&tradeDate, &symbol, &riskLevel); err != nil {
			return nil, err
		}
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" {
			continue
		}
		risk := "MEDIUM"
		if riskLevel.Valid {
			risk = resolveQuantRiskLevel(riskLevel.String)
		}
		result[tradeDate.Format("2006-01-02")+"|"+symbol] = risk
	}
	return result, nil
}

func resolveQuantRiskLevel(raw string) string {
	risk := strings.ToUpper(strings.TrimSpace(raw))
	switch risk {
	case "LOW", "MEDIUM", "HIGH":
		return risk
	default:
		return "MEDIUM"
	}
}

func buildQuantRiskPerformanceItems(aggMap map[string]*quantRiskAggregate) []model.StockQuantRiskPerformance {
	if len(aggMap) == 0 {
		return []model.StockQuantRiskPerformance{}
	}
	orderedRiskLevels := []string{"LOW", "MEDIUM", "HIGH"}
	items := make([]model.StockQuantRiskPerformance, 0, len(orderedRiskLevels))
	for _, riskLevel := range orderedRiskLevels {
		agg := aggMap[riskLevel]
		if agg == nil {
			continue
		}
		item := model.StockQuantRiskPerformance{
			RiskLevel:   riskLevel,
			SampleCount: agg.Cnt5,
		}
		if agg.Cnt5 > 0 {
			item.AvgReturn5 = roundTo(agg.Sum5/float64(agg.Cnt5), 4)
			item.HitRate5 = roundTo(float64(agg.Hit5)/float64(agg.Cnt5), 4)
		}
		if agg.Cnt10 > 0 {
			item.AvgReturn10 = roundTo(agg.Sum10/float64(agg.Cnt10), 4)
			item.HitRate10 = roundTo(float64(agg.Hit10)/float64(agg.Cnt10), 4)
		}
		items = append(items, item)
	}
	return items
}

func buildQuantRotationItems(tradeDates []string, dayTopSymbols map[string][]string) []model.StockQuantRotationPoint {
	if len(tradeDates) == 0 {
		return []model.StockQuantRotationPoint{}
	}
	result := make([]model.StockQuantRotationPoint, 0, len(tradeDates))
	prevSymbols := []string{}
	prevSet := make(map[string]struct{})
	for _, tradeDate := range tradeDates {
		currSymbols := uniqueUpperSymbols(dayTopSymbols[tradeDate])
		currSet := make(map[string]struct{}, len(currSymbols))
		for _, symbol := range currSymbols {
			currSet[symbol] = struct{}{}
		}
		entered := make([]string, 0)
		for _, symbol := range currSymbols {
			if _, ok := prevSet[symbol]; !ok {
				entered = append(entered, symbol)
			}
		}
		exited := make([]string, 0)
		for _, symbol := range prevSymbols {
			if _, ok := currSet[symbol]; !ok {
				exited = append(exited, symbol)
			}
		}
		stayed := len(currSymbols) - len(entered)
		if stayed < 0 {
			stayed = 0
		}
		result = append(result, model.StockQuantRotationPoint{
			TradeDate:    tradeDate,
			TopSymbols:   currSymbols,
			Entered:      entered,
			Exited:       exited,
			StayedCount:  stayed,
			ChangedCount: len(entered) + len(exited),
		})
		prevSymbols = currSymbols
		prevSet = currSet
	}
	return result
}

func uniqueUpperSymbols(symbols []string) []string {
	seen := make(map[string]struct{}, len(symbols))
	result := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}
