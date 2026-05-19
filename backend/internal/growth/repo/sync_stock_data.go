package repo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// tushareRateLimiter enforces max 480 calls per minute (80% of Tushare's 500/min limit)
var tushareRateLimiter = sync.Mutex{}
var tushareCallTimestamps []time.Time

func waitTushareSlot() {
	tushareRateLimiter.Lock()
	defer tushareRateLimiter.Unlock()

	now := time.Now()
	// Remove timestamps older than 1 minute
	cutoff := now.Add(-1 * time.Minute)
	var recent []time.Time
	for _, t := range tushareCallTimestamps {
		if t.After(cutoff) {
			recent = append(recent, t)
		}
	}
	tushareCallTimestamps = recent

	// If we've made 480+ calls in the last minute, wait
	if len(tushareCallTimestamps) >= 480 {
		oldest := tushareCallTimestamps[0]
		waitTime := time.Until(oldest.Add(1 * time.Minute))
		if waitTime > 0 {
			time.Sleep(waitTime + 100*time.Millisecond)
		}
		// Clear and retry
		tushareCallTimestamps = nil
	}

	tushareCallTimestamps = append(tushareCallTimestamps, now)
}

func callTushare(token, apiName string, params map[string]string, fields string) (*tushareStdResponse, error) {
	body := map[string]interface{}{
		"api_name": apiName,
		"token":    token,
		"params":   params,
	}
	if fields != "" {
		body["fields"] = fields
	}
	payload, _ := json.Marshal(body)

	waitTushareSlot()

	var lastErr error
	for attempt := 0; attempt < 4; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 700 * time.Millisecond)
		}

		resp, err := http.Post("https://api.tushare.pro", "application/json", strings.NewReader(string(payload)))
		if err != nil {
			lastErr = fmt.Errorf("tushare http error: %w", err)
			continue
		}

		var result tushareStdResponse
		if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
			resp.Body.Close()
			lastErr = fmt.Errorf("tushare decode error: %w", decodeErr)
			continue
		}
		resp.Body.Close()

		if result.Code == 0 {
			return &result, nil
		}

		if isRateLimitErr(result.Msg) {
			time.Sleep(time.Duration(attempt+1) * 5 * time.Second)
			lastErr = fmt.Errorf("tushare rate limited: %s", result.Msg)
			continue
		}

		return nil, fmt.Errorf("tushare error code=%d: %s", result.Code, result.Msg)
	}
	return nil, lastErr
}

func isRateLimitErr(msg string) bool {
	text := strings.ToLower(msg)
	keywords := []string{"最多访问该接口", "触发访问频次", "请求过于频繁", "rate limit", "too many requests"}
	for _, k := range keywords {
		if strings.Contains(text, k) {
			return true
		}
	}
	if strings.Contains(text, "每分钟") && strings.Contains(text, "访问") {
		return true
	}
	return false
}

func FetchAllStockSymbolsFromTushare() ([]string, error) {
	token := os.Getenv("TUSHARE_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TUSHARE_TOKEN not configured")
	}
	result, err := callTushare(token, "stock_basic", map[string]string{"limit": "5000", "offset": "0"}, "ts_code,symbol")
	if err != nil {
		return nil, err
	}
	var symbols []string
	for _, item := range result.Data.Items {
		if len(item) > 0 {
			if code, ok := item[0].(string); ok && code != "" {
				symbols = append(symbols, code)
			}
		}
	}
	return symbols, nil
}

func FetchAndSaveStockQuotesFromTushare(token, symbol, startDate, endDate string) error {
	result, err := callTushare(token, "daily", map[string]string{
		"ts_code": symbol, "start_date": startDate, "end_date": endDate,
	}, "ts_code,trade_date,open,high,low,close,pre_close,vol,amount")
	if err != nil {
		return err
	}
	if len(result.Data.Items) == 0 {
		return nil
	}
	fieldMap := map[string]int{}
	for i, f := range result.Data.Fields {
		fieldMap[f] = i
	}
	for _, item := range result.Data.Items {
		close := getFloatField(item, fieldMap, "close")
		if close <= 0 {
			continue
		}
		_ = upsertMarketDailyBar(
			getStringField(item, fieldMap, "ts_code"),
			getStringField(item, fieldMap, "trade_date"),
			getFloatField(item, fieldMap, "open"),
			getFloatField(item, fieldMap, "high"),
			getFloatField(item, fieldMap, "low"),
			close,
			getFloatField(item, fieldMap, "vol"),
			getFloatField(item, fieldMap, "amount"),
		)
	}
	return nil
}

func FetchAndSaveStockQuotesByDate(token, date string) error {
	result, err := callTushare(token, "daily", map[string]string{"trade_date": date},
		"ts_code,trade_date,open,high,low,close,pre_close,vol,amount")
	if err != nil {
		return err
	}
	if len(result.Data.Items) == 0 {
		return nil
	}
	fieldMap := map[string]int{}
	for i, f := range result.Data.Fields {
		fieldMap[f] = i
	}
	for _, item := range result.Data.Items {
		close := getFloatField(item, fieldMap, "close")
		if close <= 0 {
			continue
		}
		_ = upsertMarketDailyBar(
			getStringField(item, fieldMap, "ts_code"),
			getStringField(item, fieldMap, "trade_date"),
			getFloatField(item, fieldMap, "open"),
			getFloatField(item, fieldMap, "high"),
			getFloatField(item, fieldMap, "low"),
			close,
			getFloatField(item, fieldMap, "vol"),
			getFloatField(item, fieldMap, "amount"),
		)
	}
	return nil
}

func GetLatestTradeDate() string     { return "" }
func RebuildStockQuotesTruth() error { return nil }

func upsertMarketDailyBar(symbol, date string, open, high, low, close, vol, amount float64) error {
	return nil
}

func getStringField(item []interface{}, fieldMap map[string]int, key string) string {
	if idx, ok := fieldMap[key]; ok && idx < len(item) {
		if s, ok := item[idx].(string); ok {
			return s
		}
	}
	return ""
}

func getFloatField(item []interface{}, fieldMap map[string]int, key string) float64 {
	if idx, ok := fieldMap[key]; ok && idx < len(item) {
		switch v := item[idx].(type) {
		case float64:
			return v
		case string:
			var f float64
			fmt.Sscanf(v, "%f", &f)
			return f
		}
	}
	return 0
}
