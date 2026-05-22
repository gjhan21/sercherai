package handler

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/service"
)

func resolveToken(svc service.GrowthService) string {
	token := os.Getenv("TUSHARE_TOKEN")
	if token != "" {
		return token
	}
	if svc != nil {
		items, _, err := svc.AdminListSystemConfigs("tushare.token", 1, 1)
		if err == nil && len(items) > 0 {
			t := strings.TrimSpace(items[0].ConfigValue)
			if t != "" {
				return t
			}
		}
	}
	return ""
}

type SyncProgress struct {
	mu           sync.RWMutex
	Total        int      `json:"total"`
	Completed    int      `json:"completed"`
	Failed       int      `json:"failed"`
	FailedCodes  []string `json:"failed_codes"`
	Message      string   `json:"message"`
	Running      bool     `json:"running"`
	StartTime    string   `json:"start_time"`
	Elapsed      string   `json:"elapsed"`
	BatchSize    int      `json:"batch_size"`
	CurrentBatch int      `json:"current_batch"`
}

var syncState = &SyncProgress{}

func syncProgressResponse() gin.H {
	syncState.mu.RLock()
	defer syncState.mu.RUnlock()
	elapsed := ""
	if syncState.StartTime != "" {
		st, _ := time.Parse(time.RFC3339, syncState.StartTime)
		elapsed = time.Since(st).Round(time.Second).String()
	}
	return gin.H{
		"running":       syncState.Running,
		"total":         syncState.Total,
		"completed":     syncState.Completed,
		"failed":        syncState.Failed,
		"failed_codes":  syncState.FailedCodes,
		"message":       syncState.Message,
		"elapsed":       elapsed,
		"start_time":    syncState.StartTime,
		"batch_size":    syncState.BatchSize,
		"current_batch": syncState.CurrentBatch,
	}
}

func resolveDefaultStockQuoteSourceKey(svc service.GrowthService) string {
	if svc == nil {
		return "TUSHARE"
	}
	items, _, err := svc.AdminListSystemConfigs("stock.quotes.default_source_key", 1, 10)
	if err != nil || len(items) == 0 {
		return "TUSHARE"
	}
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.ConfigKey), "stock.quotes.default_source_key") {
			value := strings.ToUpper(strings.TrimSpace(item.ConfigValue))
			if value != "" {
				return value
			}
		}
	}
	return "TUSHARE"
}

func estimateIncrementalStockSyncDays(svc service.GrowthService, now time.Time) int {
	if svc == nil {
		return 7
	}
	summary, err := svc.AdminGetMarketCoverageSummary()
	if err != nil {
		return 7
	}
	latestTradeDate := strings.TrimSpace(summary.LatestTradeDate)
	if latestTradeDate == "" {
		return 7
	}
	latest, err := time.ParseInLocation("2006-01-02", latestTradeDate, now.Location())
	if err != nil {
		return 7
	}
	days := int(now.Sub(latest).Hours()/24) + 1
	if days < 1 {
		return 1
	}
	if days > 30 {
		return 30
	}
	return days
}

func (h *AdminMarketDataHandler) FullSyncStockQuotes(c *gin.Context) {
	token := resolveToken(h.service)
	if token == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "TUSHARE_TOKEN not configured", Data: struct{}{}})
		return
	}
	syncState.mu.Lock()
	if syncState.Running {
		syncState.mu.Unlock()
		c.JSON(http.StatusConflict, dto.APIResponse{Code: 40901, Message: "sync already running", Data: syncProgressResponse()})
		return
	}
	syncState.Running = true
	syncState.StartTime = time.Now().Format(time.RFC3339)
	syncState.Completed = 0
	syncState.Failed = 0
	syncState.FailedCodes = nil
	syncState.Message = ""
	syncState.Total = 0
	syncState.mu.Unlock()

	go func() {
		defer func() {
			syncState.mu.Lock()
			syncState.Running = false
			syncState.mu.Unlock()
		}()

		os.Setenv("TUSHARE_TOKEN", token)
		sourceKey := resolveDefaultStockQuoteSourceKey(h.service)
		days := 365

		syncState.mu.Lock()
		syncState.Total = days
		syncState.Message = "正在通过正式市场数据链路执行全量同步"
		syncState.mu.Unlock()

		result, err := h.service.AdminSyncStockQuotesFromMaster(sourceKey, days)
		syncState.mu.Lock()
		defer syncState.mu.Unlock()
		if err != nil {
			syncState.Failed = 1
			syncState.FailedCodes = []string{"FULL_MARKET"}
			syncState.Message = err.Error()
			return
		}
		syncState.Completed = result.TruthCount
		if syncState.Completed <= 0 {
			syncState.Completed = result.BarCount
		}
		syncState.Message = "全量同步完成，来源=" + sourceKey
	}()

	c.JSON(http.StatusOK, dto.APIResponse{Code: 0, Message: "full sync started", Data: syncProgressResponse()})
}

func (h *AdminMarketDataHandler) IncrementalSyncStockQuotes(c *gin.Context) {
	token := resolveToken(h.service)
	if token == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "TUSHARE_TOKEN not configured", Data: struct{}{}})
		return
	}
	syncState.mu.Lock()
	if syncState.Running {
		syncState.mu.Unlock()
		c.JSON(http.StatusConflict, dto.APIResponse{Code: 40901, Message: "sync already running", Data: syncProgressResponse()})
		return
	}
	syncState.Running = true
	syncState.StartTime = time.Now().Format(time.RFC3339)
	syncState.Completed = 0
	syncState.Failed = 0
	syncState.FailedCodes = nil
	syncState.Message = ""
	syncState.Total = 0
	syncState.mu.Unlock()

	go func() {
		defer func() {
			syncState.mu.Lock()
			syncState.Running = false
			syncState.mu.Unlock()
		}()

		os.Setenv("TUSHARE_TOKEN", token)
		now := time.Now()
		sourceKey := resolveDefaultStockQuoteSourceKey(h.service)
		days := estimateIncrementalStockSyncDays(h.service, now)

		syncState.mu.Lock()
		syncState.Total = days
		syncState.Message = "正在通过正式市场数据链路执行增量同步"
		syncState.mu.Unlock()

		result, err := h.service.AdminSyncStockQuotesFromMaster(sourceKey, days)
		syncState.mu.Lock()
		defer syncState.mu.Unlock()
		if err != nil {
			syncState.Failed = 1
			syncState.FailedCodes = []string{"INCREMENTAL"}
			syncState.Message = err.Error()
			return
		}
		syncState.Completed = result.TruthCount
		if syncState.Completed <= 0 {
			syncState.Completed = result.BarCount
		}
		syncState.Message = "增量同步完成，来源=" + sourceKey + "，窗口天数=" + strconv.Itoa(days)
	}()

	c.JSON(http.StatusOK, dto.APIResponse{Code: 0, Message: "incremental sync started", Data: syncProgressResponse()})
}

func (h *AdminMarketDataHandler) GetSyncProgress(c *gin.Context) {
	c.JSON(http.StatusOK, dto.APIResponse{Code: 0, Message: "ok", Data: syncProgressResponse()})
}
