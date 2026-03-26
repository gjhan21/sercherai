package repo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func withTestTushareRetryConfig(serverURL string, fn func()) {
	previousEndpoint := tushareAPIEndpoint
	previousMaxAttempts := tushareMaxRetryAttempts
	previousRetryBaseDelay := tushareRetryBaseDelay
	previousRateLimitDelay := tushareRateLimitDelay
	previousMinIntervalDefault := tushareMinIntervalDefault
	previousMinIntervalByAPI := tushareMinIntervalByAPI
	previousLastCallAt := tushareLastCallAt

	tushareAPIEndpoint = serverURL
	tushareMaxRetryAttempts = 3
	tushareRetryBaseDelay = time.Millisecond
	tushareRateLimitDelay = time.Millisecond
	tushareMinIntervalDefault = 0
	tushareMinIntervalByAPI = map[string]time.Duration{}
	tushareLastCallAt = map[string]time.Time{}

	defer func() {
		tushareAPIEndpoint = previousEndpoint
		tushareMaxRetryAttempts = previousMaxAttempts
		tushareRetryBaseDelay = previousRetryBaseDelay
		tushareRateLimitDelay = previousRateLimitDelay
		tushareMinIntervalDefault = previousMinIntervalDefault
		tushareMinIntervalByAPI = previousMinIntervalByAPI
		tushareLastCallAt = previousLastCallAt
	}()

	fn()
}

func TestCallTushareAPIRetriesRateLimitError(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		current := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if current == 1 {
			_, _ = w.Write([]byte(`{"code":-2001,"msg":"抱歉，您每分钟最多访问该接口500次","data":{"fields":[],"items":[]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok","data":{"fields":[],"items":[]}}`))
	}))
	defer server.Close()

	withTestTushareRetryConfig(server.URL, func() {
		client := &http.Client{Timeout: 2 * time.Second}
		if _, err := callTushareAPI(client, "token", "daily", map[string]string{"trade_date": "20260325"}, ""); err != nil {
			t.Fatalf("callTushareAPI returned error: %v", err)
		}
	})

	if got := calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls with one retry, got %d", got)
	}
}

func TestCallTushareAPIRetriesHTTP502(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		current := calls.Add(1)
		if current == 1 {
			w.WriteHeader(http.StatusBadGateway)
			_, _ = w.Write([]byte(`bad gateway`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"msg":"ok","data":{"fields":[],"items":[]}}`))
	}))
	defer server.Close()

	withTestTushareRetryConfig(server.URL, func() {
		client := &http.Client{Timeout: 2 * time.Second}
		if _, err := callTushareAPI(client, "token", "daily", map[string]string{"trade_date": "20260325"}, ""); err != nil {
			t.Fatalf("callTushareAPI returned error: %v", err)
		}
	})

	if got := calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls with one retry, got %d", got)
	}
}

func TestCallTushareAPIDoesNotRetryBusinessError(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":-1,"msg":"invalid token","data":{"fields":[],"items":[]}}`))
	}))
	defer server.Close()

	withTestTushareRetryConfig(server.URL, func() {
		client := &http.Client{Timeout: 2 * time.Second}
		_, err := callTushareAPI(client, "token", "daily", map[string]string{"trade_date": "20260325"}, "")
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		expected := "tushare error(daily): invalid token"
		if err.Error() != expected {
			t.Fatalf("expected %q, got %q", expected, err.Error())
		}
	})

	if got := calls.Load(); got != 1 {
		t.Fatalf("expected 1 call without retry, got %d", got)
	}
}

func TestShouldRetryTushareRequestOnEOF(t *testing.T) {
	if !shouldRetryTushareRequest(fmt.Errorf("Post \"https://api.tushare.pro\": EOF"), 0) {
		t.Fatalf("expected EOF to be retryable")
	}
}
