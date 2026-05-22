package worker

import (
	"log"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/growth/service"
)

const (
	docFastIncrementalJobName            = "doc_fast_news_incremental"
	docFastIncrementalDefaultMinutes     = 100
	docFastIncrementalMaxMinutes         = 24 * 60
	tushareNewsIncrementalJobName        = "tushare_news_incremental"
	tushareNewsIncrementalDefaultMinutes = 20
	tushareNewsIncrementalMaxMinutes     = 24 * 60
	vipLifecycleJobName                  = "vip_membership_lifecycle"
	vipLifecycleDefaultMinutes           = 30
	vipLifecycleMaxMinutes               = 24 * 60
	forecastL3DispatchJobName            = "forecast_l3_dispatch_pending"
	forecastL3DispatchDefaultMinutes     = 5
	forecastL3QualityJobName             = "forecast_l3_quality_backfill"
	forecastL3QualityDefaultMinutes      = 60
)

func StartAll(growthSvc service.GrowthService) {
	StartDocFastIncrementalSyncWorker(growthSvc)
	StartTushareNewsIncrementalSyncWorker(growthSvc)
	StartVIPMembershipLifecycleWorker(growthSvc)
	StartForecastL3DispatchWorker(growthSvc)
	StartForecastL3QualityWorker(growthSvc)
}

func StartDocFastIncrementalSyncWorker(growthSvc service.GrowthService) {
	go func() {
		log.Printf("[scheduler] start doc_fast incremental worker")
		for {
			enabled, intervalMinutes := loadDocFastIncrementalWorkerConfig(growthSvc)
			if enabled {
				runDocFastIncrementalJob(growthSvc, "SYSTEM_TIMER")
			}
			if intervalMinutes <= 0 {
				intervalMinutes = docFastIncrementalDefaultMinutes
			}
			time.Sleep(time.Duration(intervalMinutes) * time.Minute)
		}
	}()
}

func runDocFastIncrementalJob(growthSvc service.GrowthService, triggerSource string) {
	summary, runErr := growthSvc.AdminSyncDocFastNewsIncremental(0)
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	_, logErr := growthSvc.AdminCreateSchedulerJobRun(
		docFastIncrementalJobName,
		triggerSource,
		status,
		summary,
		errorMessage,
		"system",
	)
	if logErr != nil {
		log.Printf("[scheduler] create job run failed(%s): %v", docFastIncrementalJobName, logErr)
	}
	if runErr != nil {
		log.Printf("[scheduler] job failed(%s): %v", docFastIncrementalJobName, runErr)
		return
	}
	log.Printf("[scheduler] job success(%s): %s", docFastIncrementalJobName, strings.TrimSpace(summary))
}

func StartTushareNewsIncrementalSyncWorker(growthSvc service.GrowthService) {
	go func() {
		log.Printf("[scheduler] start tushare news incremental worker")
		for {
			enabled, intervalMinutes := loadTushareNewsIncrementalWorkerConfig(growthSvc)
			if enabled {
				runTushareNewsIncrementalJob(growthSvc, "SYSTEM_TIMER")
			}
			if intervalMinutes <= 0 {
				intervalMinutes = tushareNewsIncrementalDefaultMinutes
			}
			time.Sleep(time.Duration(intervalMinutes) * time.Minute)
		}
	}()
}

func runTushareNewsIncrementalJob(growthSvc service.GrowthService, triggerSource string) {
	summary, details, runErr := growthSvc.AdminSyncTushareNewsIncrementalWithOptions(model.TushareNewsSyncOptions{})
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	runID, logErr := growthSvc.AdminCreateSchedulerJobRun(
		tushareNewsIncrementalJobName,
		triggerSource,
		status,
		summary,
		errorMessage,
		"system",
	)
	if logErr != nil {
		log.Printf("[scheduler] create job run failed(%s): %v", tushareNewsIncrementalJobName, logErr)
	} else if len(details) > 0 {
		if detailErr := growthSvc.AdminCreateNewsSyncRunDetails(runID, details); detailErr != nil {
			log.Printf("[scheduler] create news sync details failed(%s): %v", runID, detailErr)
		}
	}
	if runErr != nil {
		log.Printf("[scheduler] job failed(%s): %v", tushareNewsIncrementalJobName, runErr)
		return
	}
	log.Printf("[scheduler] job success(%s): %s", tushareNewsIncrementalJobName, strings.TrimSpace(summary))
}

func StartVIPMembershipLifecycleWorker(growthSvc service.GrowthService) {
	go func() {
		log.Printf("[scheduler] start vip membership lifecycle worker")
		for {
			enabled, intervalMinutes := loadVIPMembershipLifecycleWorkerConfig(growthSvc)
			if enabled {
				runVIPMembershipLifecycleJob(growthSvc, "SYSTEM_TIMER")
			}
			if intervalMinutes <= 0 {
				intervalMinutes = vipLifecycleDefaultMinutes
			}
			time.Sleep(time.Duration(intervalMinutes) * time.Minute)
		}
	}()
}

func runVIPMembershipLifecycleJob(growthSvc service.GrowthService, triggerSource string) {
	summary, runErr := growthSvc.AdminRunVIPMembershipLifecycle()
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	_, logErr := growthSvc.AdminCreateSchedulerJobRun(
		vipLifecycleJobName,
		triggerSource,
		status,
		summary,
		errorMessage,
		"system",
	)
	if logErr != nil {
		log.Printf("[scheduler] create job run failed(%s): %v", vipLifecycleJobName, logErr)
	}
	if runErr != nil {
		log.Printf("[scheduler] job failed(%s): %v", vipLifecycleJobName, runErr)
		return
	}
	log.Printf("[scheduler] job success(%s): %s", vipLifecycleJobName, strings.TrimSpace(summary))
}

func StartForecastL3DispatchWorker(growthSvc service.GrowthService) {
	go func() {
		log.Printf("[scheduler] start forecast l3 dispatch worker")
		for {
			enabled, intervalMinutes := loadForecastL3DispatchWorkerConfig(growthSvc)
			if enabled {
				runForecastL3DispatchJob(growthSvc, "SYSTEM_TIMER")
			}
			if intervalMinutes <= 0 {
				intervalMinutes = forecastL3DispatchDefaultMinutes
			}
			time.Sleep(time.Duration(intervalMinutes) * time.Minute)
		}
	}()
}

func runForecastL3DispatchJob(growthSvc service.GrowthService, triggerSource string) {
	count, runErr := growthSvc.ExecuteQueuedStrategyForecastL3Runs(10, "system")
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	summary := "executed forecast l3 runs: " + strconv.Itoa(count)
	_, logErr := growthSvc.AdminCreateSchedulerJobRun(
		forecastL3DispatchJobName,
		triggerSource,
		status,
		summary,
		errorMessage,
		"system",
	)
	if logErr != nil {
		log.Printf("[scheduler] create job run failed(%s): %v", forecastL3DispatchJobName, logErr)
	}
	if runErr != nil {
		log.Printf("[scheduler] job failed(%s): %v", forecastL3DispatchJobName, runErr)
		return
	}
	log.Printf("[scheduler] job success(%s): %s", forecastL3DispatchJobName, summary)
}

func StartForecastL3QualityWorker(growthSvc service.GrowthService) {
	go func() {
		log.Printf("[scheduler] start forecast l3 quality worker")
		for {
			enabled, intervalMinutes := loadForecastL3QualityWorkerConfig(growthSvc)
			if enabled {
				runForecastL3QualityJob(growthSvc, "SYSTEM_TIMER")
			}
			if intervalMinutes <= 0 {
				intervalMinutes = forecastL3QualityDefaultMinutes
			}
			time.Sleep(time.Duration(intervalMinutes) * time.Minute)
		}
	}()
}

func runForecastL3QualityJob(growthSvc service.GrowthService, triggerSource string) {
	count, runErr := growthSvc.RunStrategyForecastL3QualityBackfill(20, "system")
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	summary := "quality backfill forecast l3 records: " + strconv.Itoa(count)
	_, logErr := growthSvc.AdminCreateSchedulerJobRun(
		forecastL3QualityJobName,
		triggerSource,
		status,
		summary,
		errorMessage,
		"system",
	)
	if logErr != nil {
		log.Printf("[scheduler] create job run failed(%s): %v", forecastL3QualityJobName, logErr)
	}
	if runErr != nil {
		log.Printf("[scheduler] job failed(%s): %v", forecastL3QualityJobName, runErr)
		return
	}
	log.Printf("[scheduler] job success(%s): %s", forecastL3QualityJobName, summary)
}

func loadDocFastIncrementalWorkerConfig(growthSvc service.GrowthService) (bool, int) {
	enabled := true
	intervalMinutes := docFastIncrementalDefaultMinutes

	items, _, err := growthSvc.AdminListSystemConfigs("news.sync.doc_fast.", 1, 200)
	if err != nil {
		return enabled, intervalMinutes
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case "news.sync.doc_fast.enabled":
			enabled = parseRouterBoolConfig(value, enabled)
		case "news.sync.doc_fast.interval_minutes":
			intervalMinutes = parseRouterIntConfig(value, intervalMinutes)
		}
	}
	if intervalMinutes <= 0 {
		intervalMinutes = docFastIncrementalDefaultMinutes
	}
	if intervalMinutes > docFastIncrementalMaxMinutes {
		intervalMinutes = docFastIncrementalMaxMinutes
	}
	return enabled, intervalMinutes
}

func loadForecastL3DispatchWorkerConfig(growthSvc service.GrowthService) (bool, int) {
	enabled := false
	intervalMinutes := forecastL3DispatchDefaultMinutes

	items, _, err := growthSvc.AdminListSystemConfigs("growth.forecast_l3.", 1, 200)
	if err != nil {
		return enabled, intervalMinutes
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case "growth.forecast_l3.enabled":
			enabled = parseRouterBoolConfig(value, enabled)
		case "growth.forecast_l3.dispatch.enabled":
			enabled = parseRouterBoolConfig(value, enabled)
		case "growth.forecast_l3.dispatch.interval_minutes":
			intervalMinutes = parseRouterIntConfig(value, intervalMinutes)
		}
	}
	if intervalMinutes <= 0 {
		intervalMinutes = forecastL3DispatchDefaultMinutes
	}
	return enabled, intervalMinutes
}

func loadForecastL3QualityWorkerConfig(growthSvc service.GrowthService) (bool, int) {
	enabled := false
	intervalMinutes := forecastL3QualityDefaultMinutes

	items, _, err := growthSvc.AdminListSystemConfigs("growth.forecast_l3.", 1, 200)
	if err != nil {
		return enabled, intervalMinutes
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case "growth.forecast_l3.enabled":
			enabled = parseRouterBoolConfig(value, enabled)
		case "growth.forecast_l3.quality.enabled":
			enabled = parseRouterBoolConfig(value, enabled)
		case "growth.forecast_l3.quality.interval_minutes":
			intervalMinutes = parseRouterIntConfig(value, intervalMinutes)
		}
	}
	if intervalMinutes <= 0 {
		intervalMinutes = forecastL3QualityDefaultMinutes
	}
	return enabled, intervalMinutes
}

func loadTushareNewsIncrementalWorkerConfig(growthSvc service.GrowthService) (bool, int) {
	enabled := true
	intervalMinutes := tushareNewsIncrementalDefaultMinutes

	items, _, err := growthSvc.AdminListSystemConfigs("news.sync.tushare.", 1, 200)
	if err != nil {
		return enabled, intervalMinutes
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case "news.sync.tushare.enabled":
			enabled = parseRouterBoolConfig(value, enabled)
		case "news.sync.tushare.interval_minutes":
			intervalMinutes = parseRouterIntConfig(value, intervalMinutes)
		}
	}
	if intervalMinutes <= 0 {
		intervalMinutes = tushareNewsIncrementalDefaultMinutes
	}
	if intervalMinutes > tushareNewsIncrementalMaxMinutes {
		intervalMinutes = tushareNewsIncrementalMaxMinutes
	}
	return enabled, intervalMinutes
}

func loadVIPMembershipLifecycleWorkerConfig(growthSvc service.GrowthService) (bool, int) {
	enabled := true
	intervalMinutes := vipLifecycleDefaultMinutes

	items, _, err := growthSvc.AdminListSystemConfigs("membership.vip.lifecycle.", 1, 50)
	if err != nil {
		return enabled, intervalMinutes
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case "membership.vip.lifecycle.enabled":
			enabled = parseRouterBoolConfig(value, enabled)
		case "membership.vip.lifecycle.interval_minutes":
			intervalMinutes = parseRouterIntConfig(value, intervalMinutes)
		}
	}
	if intervalMinutes <= 0 {
		intervalMinutes = vipLifecycleDefaultMinutes
	}
	if intervalMinutes > vipLifecycleMaxMinutes {
		intervalMinutes = vipLifecycleMaxMinutes
	}
	return enabled, intervalMinutes
}

func parseRouterBoolConfig(raw string, fallback bool) bool {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" {
		return fallback
	}
	switch text {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func parseRouterIntConfig(raw string, fallback int) int {
	text := strings.TrimSpace(raw)
	if text == "" {
		return fallback
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return fallback
	}
	return value
}
