package worker

import (
	"fmt"
	"log"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/utils"
)

const (
	stockDefaultSourceConfigKey               = "stock.quotes.default_source_key"
	stockDefaultSourceFallback                = "TUSHARE"
	schedulerJobDailyFuturesStrategy          = "daily_futures_strategy"
	schedulerJobFuturesStrategyGenerate       = "futures_strategy_generate"
	schedulerJobFuturesStrategyEvaluate       = "futures_strategy_evaluate"
	schedulerAutoRetryEnabledConfigKey        = "scheduler.auto_retry.enabled"
	schedulerAutoRetryMaxRetriesConfigKey     = "scheduler.auto_retry.max_retries"
	schedulerAutoRetryBackoffSecondsConfigKey = "scheduler.auto_retry.backoff_seconds"
	schedulerAutoRetryJobsConfigKey           = "scheduler.auto_retry.jobs"
	schedulerAutoRetryDefaultJob              = "daily_stock_quant_pipeline"
)

type JobExecutionRequest struct {
	RunID         string
	JobName       string
	OperatorID    string
	TriggerSource string
	SyncOptions   model.TushareNewsSyncOptions
}

var JobQueue = make(chan JobExecutionRequest, 100)

// StartJobExecutorWorkers starts background goroutines to consume tasks from JobQueue
func StartJobExecutorWorkers(concurrency int, svc service.GrowthService) {
	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			log.Printf("[worker] job execution worker %d started", workerID)
			for req := range JobQueue {
				log.Printf("[worker] worker %d picked up job: %s (runID: %s)", workerID, req.JobName, req.RunID)
				executeJobWithAutoRetry(svc, req)
			}
		}(i)
	}
}

func executeJobWithAutoRetry(svc service.GrowthService, req JobExecutionRequest) {
	runID := req.RunID
	var err error
	if runID == "" {
		// If timer triggered, first create the database record as RUNNING
		runID, err = svc.AdminCreateSchedulerJobRun(req.JobName, req.TriggerSource, "RUNNING", "Timer triggered job started", "", req.OperatorID)
		if err != nil {
			log.Printf("[worker] failed to create scheduler job run for %s: %v", req.JobName, err)
			return
		}
	}

	execResult, runErr := runSchedulerJob(svc, req.JobName, req.SyncOptions)
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	resultSummary := execResult.Summary

	// Update execution status in DB
	err = svc.AdminUpdateSchedulerJobRun(runID, status, resultSummary, errorMessage)
	if err != nil {
		log.Printf("[worker] failed to update scheduler job run status for %s (runID: %s): %v", req.JobName, runID, err)
	}
	if len(execResult.NewsSyncDetails) > 0 {
		if detailErr := svc.AdminCreateNewsSyncRunDetails(runID, execResult.NewsSyncDetails); detailErr != nil {
			log.Printf("[worker] failed to create news sync run details for %s (runID: %s): %v", req.JobName, runID, detailErr)
		}
	}

	if status == "FAILED" {
		// Auto retry logic
		executeSchedulerAutoRetry(svc, req.JobName, runID, status, resultSummary, errorMessage, req.OperatorID, req.SyncOptions)
	}
}

func executeSchedulerAutoRetry(svc service.GrowthService, jobName string, baseRunID string, baseStatus string, baseSummary string, baseError string, operator string, syncOptions model.TushareNewsSyncOptions) {
	policy := resolveSchedulerAutoRetryPolicy(svc, jobName)
	if !policy.Enabled {
		return
	}
	currentRunID := baseRunID
	currentStatus := baseStatus
	for attempt := 1; attempt <= policy.MaxRetries && currentStatus == "FAILED"; attempt++ {
		if policy.BackoffSeconds > 0 {
			time.Sleep(time.Duration(policy.BackoffSeconds*attempt) * time.Second)
		}
		log.Printf("[worker] auto retry job %s, attempt %d/%d", jobName, attempt, policy.MaxRetries)
		execResult, runErr := runSchedulerJob(svc, jobName, syncOptions)
		summary := execResult.Summary
		status := "SUCCESS"
		errorMessage := ""
		if runErr != nil {
			status = "FAILED"
			errorMessage = runErr.Error()
		}
		newRunID, createErr := svc.AdminRetrySchedulerJobRun(currentRunID, "SYSTEM", status, summary, errorMessage, operator)
		if createErr != nil {
			log.Printf("[worker] failed to create retry scheduler job run for %s: %v", jobName, createErr)
			break
		}
		if len(execResult.NewsSyncDetails) > 0 {
			if detailErr := svc.AdminCreateNewsSyncRunDetails(newRunID, execResult.NewsSyncDetails); detailErr != nil {
				log.Printf("[worker] failed to create news sync run details in retry for %s: %v", jobName, detailErr)
			}
		}
		currentRunID = newRunID
		currentStatus = status
	}
}

type schedulerJobExecutionResult struct {
	Summary         string
	NewsSyncDetails []model.NewsSyncRunDetail
}

type schedulerAutoRetryPolicy struct {
	Enabled        bool
	MaxRetries     int
	BackoffSeconds int
}

func resolveSchedulerAutoRetryPolicy(svc service.GrowthService, jobName string) schedulerAutoRetryPolicy {
	policy := schedulerAutoRetryPolicy{
		Enabled:        strings.EqualFold(strings.TrimSpace(jobName), schedulerAutoRetryDefaultJob),
		MaxRetries:     2,
		BackoffSeconds: 2,
	}
	items, _, err := svc.AdminListSystemConfigs("scheduler.auto_retry", 1, 200)
	if err != nil || len(items) == 0 {
		return policy
	}
	allowedJobs := map[string]struct{}{
		strings.ToLower(strings.TrimSpace(schedulerAutoRetryDefaultJob)): {},
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case strings.ToLower(schedulerAutoRetryEnabledConfigKey):
			policy.Enabled = utils.ParseConfigBool(value, policy.Enabled)
		case strings.ToLower(schedulerAutoRetryMaxRetriesConfigKey):
			policy.MaxRetries = utils.ParseConfigInt(value, policy.MaxRetries)
		case strings.ToLower(schedulerAutoRetryBackoffSecondsConfigKey):
			policy.BackoffSeconds = utils.ParseConfigInt(value, policy.BackoffSeconds)
		case strings.ToLower(schedulerAutoRetryJobsConfigKey):
			allowedJobs = map[string]struct{}{}
			for _, name := range strings.Split(value, ",") {
				normalized := strings.ToLower(strings.TrimSpace(name))
				if normalized == "" {
					continue
				}
				allowedJobs[normalized] = struct{}{}
			}
		}
	}
	if policy.MaxRetries < 0 {
		policy.MaxRetries = 0
	}
	if policy.MaxRetries > 5 {
		policy.MaxRetries = 5
	}
	if policy.BackoffSeconds < 0 {
		policy.BackoffSeconds = 0
	}
	if policy.BackoffSeconds > 60 {
		policy.BackoffSeconds = 60
	}
	if len(allowedJobs) > 0 {
		if _, ok := allowedJobs[strings.ToLower(strings.TrimSpace(jobName))]; !ok {
			policy.Enabled = false
		}
	}
	if policy.MaxRetries <= 0 {
		policy.Enabled = false
	}
	return policy
}

func runSchedulerJob(svc service.GrowthService, jobName string, syncOptions model.TushareNewsSyncOptions) (schedulerJobExecutionResult, error) {
	switch strings.ToLower(strings.TrimSpace(jobName)) {
	case "daily_stock_quant_pipeline":
		tradeDate := time.Now().Format("2006-01-02")
		sourceKey := strings.ToUpper(strings.TrimSpace(resolveDefaultStockQuoteSourceKey(svc)))
		if sourceKey == "" {
			sourceKey = stockDefaultSourceFallback
		}
		usedSourceKey := sourceKey
		quoteCount, err := svc.AdminSyncStockQuotes(sourceKey, nil, 180)
		if err != nil && sourceKey != "MOCK" {
			fallbackCount, fallbackErr := svc.AdminSyncStockQuotes("MOCK", nil, 180)
			if fallbackErr != nil {
				return schedulerJobExecutionResult{}, fmt.Errorf("sync quotes failed(%s): %v, fallback MOCK failed: %w", sourceKey, err, fallbackErr)
			}
			quoteCount = fallbackCount
			usedSourceKey = "MOCK"
		}
		if err != nil && sourceKey == "MOCK" {
			return schedulerJobExecutionResult{}, err
		}
		// Settle simulated positions after successful daily quotes sync
		if settleErr := svc.AdminSettlementSimulatedPositions(tradeDate); settleErr != nil {
			log.Printf("[worker] simulated positions settlement failed for %s: %v", tradeDate, settleErr)
		}
		topItems, err := svc.AdminGetQuantTopStocks(10, 180)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		recoResult, err := svc.AdminGenerateDailyStockRecommendations(tradeDate)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{
			Summary: fmt.Sprintf(
				"trade_date=%s source=%s quotes=%d top=%d recommendations=%d",
				tradeDate,
				usedSourceKey,
				quoteCount,
				len(topItems),
				recoResult.Count,
			),
		}, nil
	case "daily_stock_recommendation":
		tradeDate := time.Now().Format("2006-01-02")
		result, err := svc.AdminGenerateDailyStockRecommendations(tradeDate)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{Summary: fmt.Sprintf("generated %d recommendations", result.Count)}, nil
	case schedulerJobDailyFuturesStrategy, schedulerJobFuturesStrategyGenerate:
		tradeDate := time.Now().Format("2006-01-02")
		result, err := svc.AdminGenerateDailyFuturesStrategies(tradeDate)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{
			Summary: fmt.Sprintf("trade_date=%s generated=%d", tradeDate, result.Count),
		}, nil
	case schedulerJobFuturesStrategyEvaluate:
		return schedulerJobExecutionResult{Summary: "Futures strategy evaluation triggered"}, nil
	case "doc_fast_news_incremental":
		summary, err := svc.AdminSyncDocFastNewsIncremental(0)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{Summary: summary}, nil
	case "tushare_news_incremental":
		summary, details, err := svc.AdminSyncTushareNewsIncrementalWithOptions(syncOptions)
		if err != nil {
			return schedulerJobExecutionResult{Summary: summary, NewsSyncDetails: details}, err
		}
		return schedulerJobExecutionResult{Summary: summary, NewsSyncDetails: details}, nil
	case "vip_membership_lifecycle":
		summary, err := svc.AdminRunVIPMembershipLifecycle()
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{Summary: summary}, nil
	default:
		return schedulerJobExecutionResult{}, fmt.Errorf("unsupported job: %s", jobName)
	}
}

func resolveDefaultStockQuoteSourceKey(svc service.GrowthService) string {
	items, _, err := svc.AdminListSystemConfigs(stockDefaultSourceConfigKey, 1, 10)
	if err != nil || len(items) == 0 {
		return stockDefaultSourceFallback
	}
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.ConfigKey), stockDefaultSourceConfigKey) {
			return strings.TrimSpace(item.ConfigValue)
		}
	}
	return stockDefaultSourceFallback
}
