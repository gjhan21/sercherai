package handler

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/repo"
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

		today := time.Now()
		start := today.AddDate(0, 0, -365)
		// Count trading days (weekdays only)
		totalDays := 0
		for d := start; !d.After(today); d = d.AddDate(0, 0, 1) {
			if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
				totalDays++
			}
		}

		syncState.mu.Lock()
		syncState.Total = totalDays
		syncState.mu.Unlock()

		for d := start; !d.After(today); d = d.AddDate(0, 0, 1) {
			if d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
				continue
			}
			dateStr := d.Format("20060102")
			err := repo.FetchAndSaveStockQuotesByDate(token, dateStr)
			syncState.mu.Lock()
			if err != nil {
				syncState.Failed++
				syncState.FailedCodes = append(syncState.FailedCodes, dateStr)
			} else {
				syncState.Completed++
			}
			syncState.mu.Unlock()
		}
	}()

	c.JSON(http.StatusOK, dto.APIResponse{Code: 0, Message: "full sync started (by date mode)", Data: syncProgressResponse()})
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

		latestDate := repo.GetLatestTradeDate()
		startDate := latestDate
		if startDate == "" {
			startDate = time.Now().AddDate(0, 0, -7).Format("20060102")
		} else {
			t, err := time.Parse("20060102", startDate)
			if err == nil {
				startDate = t.AddDate(0, 0, 1).Format("20060102")
			}
		}
		endDate := time.Now().Format("20060102")

		current := startDate
		for current <= endDate {
			err := repo.FetchAndSaveStockQuotesByDate(token, current)
			syncState.mu.Lock()
			if err != nil {
				syncState.Failed++
				syncState.FailedCodes = append(syncState.FailedCodes, current)
			} else {
				syncState.Completed++
			}
			syncState.mu.Unlock()

			t, _ := time.Parse("20060102", current)
			current = t.AddDate(0, 0, 1).Format("20060102")
		}
		_ = repo.RebuildStockQuotesTruth()
	}()

	c.JSON(http.StatusOK, dto.APIResponse{Code: 0, Message: "incremental sync started", Data: syncProgressResponse()})
}

func (h *AdminMarketDataHandler) GetSyncProgress(c *gin.Context) {
	c.JSON(http.StatusOK, dto.APIResponse{Code: 0, Message: "ok", Data: syncProgressResponse()})
}
