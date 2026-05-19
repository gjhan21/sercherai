package handler

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/platform/oss"
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

type supportedSchedulerJob struct {
	JobName     string `json:"job_name"`
	DisplayName string `json:"display_name"`
	Module      string `json:"module"`
	AliasOf     string `json:"alias_of,omitempty"`
}

var supportedSchedulerJobs = []supportedSchedulerJob{
	{JobName: "daily_stock_quant_pipeline", DisplayName: "每日股票量化流水线", Module: "STOCK"},
	{JobName: "daily_stock_recommendation", DisplayName: "每日股票推荐", Module: "STOCK"},
	{JobName: schedulerJobDailyFuturesStrategy, DisplayName: "每日期货策略", Module: "FUTURES"},
	{JobName: schedulerJobFuturesStrategyGenerate, DisplayName: "期货策略生成(别名)", Module: "FUTURES", AliasOf: schedulerJobDailyFuturesStrategy},
	{JobName: schedulerJobFuturesStrategyEvaluate, DisplayName: "期货策略评估", Module: "FUTURES"},
	{JobName: "doc_fast_news_incremental", DisplayName: "DocFast资讯增量同步", Module: "NEWS"},
	{JobName: "tushare_news_incremental", DisplayName: "Tushare资讯增量同步", Module: "NEWS"},
	{JobName: "vip_membership_lifecycle", DisplayName: "VIP会员生命周期任务", Module: "SYSTEM"},
}

type AdminSystemHandler struct {
	AdminBaseHandler
}

func NewAdminSystemHandler(base *AdminBaseHandler) *AdminSystemHandler {
	return &AdminSystemHandler{AdminBaseHandler: *base}
}

func (h *AdminSystemHandler) ListSystemConfigs(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	keyword := c.Query("keyword")
	includeSensitive := utils.ParseConfigBool(c.Query("include_sensitive"), false)
	items, total, err := h.service.AdminListSystemConfigs(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	responseItems := sanitizeSystemConfigItems(items, includeSensitive)
	if includeSensitive {
		targetID := strings.TrimSpace(keyword)
		if targetID == "" {
			targetID = "ALL"
		}
		h.writeOperationLog(
			c,
			"SYSTEM",
			"VIEW_SENSITIVE_CONFIG_VALUES",
			"SYSTEM_CONFIG",
			targetID,
			"",
			fmt.Sprintf("count=%d", len(responseItems)),
			"include_sensitive=true",
		)
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": responseItems, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminSystemHandler) UpsertSystemConfig(c *gin.Context) {
	var req dto.SystemConfigUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpsertSystemConfig(req.ConfigKey, req.ConfigValue, req.Description, operator); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(
		c,
		"SYSTEM",
		"UPSERT_CONFIG",
		"SYSTEM_CONFIG",
		req.ConfigKey,
		"",
		maskSystemConfigValueForAudit(req.ConfigKey, req.ConfigValue),
		req.Description,
	)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminSystemHandler) ListDataSources(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListDataSources(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminSystemHandler) CreateDataSource(c *gin.Context) {
	var req dto.DataSourceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateDataSource(model.DataSource{
		SourceKey:  req.SourceKey,
		Name:       req.Name,
		SourceType: req.SourceType,
		Status:     req.Status,
		Config:     req.Config,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "exists") {
			c.JSON(http.StatusConflict, dto.APIResponse{Code: 40901, Message: err.Error(), Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "CREATE_DATA_SOURCE", "DATA_SOURCE", id, "", req.Status, req.SourceKey)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminSystemHandler) UpdateDataSource(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	var req dto.DataSourceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	err := h.service.AdminUpdateDataSource(sourceKey, model.DataSource{
		Name:       req.Name,
		SourceType: req.SourceType,
		Status:     req.Status,
		Config:     req.Config,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "UPDATE_DATA_SOURCE", "DATA_SOURCE", sourceKey, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminSystemHandler) DeleteDataSource(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	if err := h.service.AdminDeleteDataSource(sourceKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "DELETE_DATA_SOURCE", "DATA_SOURCE", sourceKey, "", "DELETED", "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminSystemHandler) CheckDataSourceHealth(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	item, err := h.service.AdminCheckDataSourceHealth(sourceKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "CHECK_DATA_SOURCE_HEALTH", "DATA_SOURCE", sourceKey, "", item.Status, item.Message)
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminSystemHandler) BatchCheckDataSourcesHealth(c *gin.Context) {
	var req dto.DataSourceBatchHealthCheckRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
			return
		}
	}
	items, err := h.service.AdminBatchCheckDataSourceHealth(req.SourceKeys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "BATCH_CHECK_DATA_SOURCE_HEALTH", "DATA_SOURCE", "BATCH", "", fmt.Sprintf("count=%d", len(items)), "")
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "count": len(items)}))
}

func (h *AdminSystemHandler) ListDataSourceHealthLogs(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListDataSourceHealthLogs(sourceKey, page, pageSize)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminSystemHandler) ListSchedulerJobDefinitions(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	module := c.Query("module")
	items, total, err := h.service.AdminListSchedulerJobDefinitions(status, module, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminSystemHandler) ListSupportedSchedulerJobs(c *gin.Context) {
	items := make([]supportedSchedulerJob, 0, len(supportedSchedulerJobs))
	moduleFilter := strings.ToUpper(strings.TrimSpace(c.Query("module")))
	for _, item := range supportedSchedulerJobs {
		if moduleFilter != "" && strings.ToUpper(item.Module) != moduleFilter {
			continue
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "total": len(items)}))
}

func (h *AdminSystemHandler) CreateSchedulerJobDefinition(c *gin.Context) {
	var req dto.SchedulerJobDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	id, err := h.service.AdminCreateSchedulerJobDefinition(model.SchedulerJobDefinition{
		JobName:     req.JobName,
		DisplayName: req.DisplayName,
		Module:      req.Module,
		CronExpr:    req.CronExpr,
		Status:      req.Status,
	}, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "CREATE_JOB_DEFINITION", "JOB_DEFINITION", id, "", req.Status, req.JobName)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminSystemHandler) UpdateSchedulerJobDefinition(c *gin.Context) {
	id := c.Param("id")
	var req dto.SchedulerJobDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpdateSchedulerJobDefinition(id, model.SchedulerJobDefinition{
		JobName:     req.JobName,
		DisplayName: req.DisplayName,
		Module:      req.Module,
		CronExpr:    req.CronExpr,
		Status:      req.Status,
	}, operator); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "UPDATE_JOB_DEFINITION", "JOB_DEFINITION", id, "", req.Status, req.CronExpr)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminSystemHandler) UpdateSchedulerJobDefinitionStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.SchedulerJobDefinitionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpdateSchedulerJobDefinitionStatus(id, req.Status, operator); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "UPDATE_JOB_DEFINITION_STATUS", "JOB_DEFINITION", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminSystemHandler) DeleteSchedulerJobDefinition(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if err := h.service.AdminDeleteSchedulerJobDefinition(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "job definition not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "DELETE_JOB_DEFINITION", "JOB_DEFINITION", id, "", "DELETED", "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminSystemHandler) ListSchedulerJobRuns(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	jobName := c.Query("job_name")
	status := c.Query("status")
	items, total, err := h.service.AdminListSchedulerJobRuns(jobName, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminSystemHandler) TriggerSchedulerJob(c *gin.Context) {
	var req dto.SchedulerTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	simulateStatus := strings.ToUpper(strings.TrimSpace(req.SimulateStatus))
	if simulateStatus != "" && h.cfg.AllowJobSimulation {
		id, err := h.service.AdminCreateSchedulerJobRun(req.JobName, req.TriggerSource, simulateStatus, req.ResultSummary, req.ErrorMessage, operator)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		h.writeOperationLog(c, "SCHEDULER", "TRIGGER_JOB", "JOB", req.JobName, "", simulateStatus, req.TriggerSource)
		c.JSON(http.StatusOK, dto.OK(gin.H{"id": id, "status": simulateStatus}))
		return
	}
	syncOptions := h.buildTushareNewsSyncOptions(req.NewsSources, req.Symbols, req.SyncTypes, req.BatchSize)
	execResult, err := h.runSchedulerJob(req.JobName, syncOptions)
	status := "SUCCESS"
	errorMessage := ""
	if err != nil {
		status = "FAILED"
		errorMessage = err.Error()
	}
	resultSummary := execResult.Summary
	id, err := h.service.AdminCreateSchedulerJobRun(req.JobName, req.TriggerSource, status, resultSummary, errorMessage, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if len(execResult.NewsSyncDetails) > 0 {
		if detailErr := h.service.AdminCreateNewsSyncRunDetails(id, execResult.NewsSyncDetails); detailErr != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: detailErr.Error(), Data: struct{}{}})
			return
		}
	}
	finalRunID, finalStatus, finalSummary, finalError, retryAttempts, retryErr := h.executeSchedulerAutoRetry(
		req.JobName,
		id,
		status,
		resultSummary,
		errorMessage,
		operator,
		syncOptions,
	)
	if retryErr != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: retryErr.Error(), Data: struct{}{}})
		return
	}
	reason := req.TriggerSource
	if retryAttempts > 0 {
		reason = fmt.Sprintf("%s,auto_retry=%d", req.TriggerSource, retryAttempts)
	}
	h.writeOperationLog(c, "SCHEDULER", "TRIGGER_JOB", "JOB", req.JobName, "", finalStatus, reason)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"id":                 finalRunID,
		"status":             finalStatus,
		"first_run_id":       id,
		"retry_attempts":     retryAttempts,
		"result_summary":     finalSummary,
		"error_message":      finalError,
		"auto_retry_applied": retryAttempts > 0,
	}))
}

func (h *AdminSystemHandler) RetrySchedulerJobRun(c *gin.Context) {
	runID := c.Param("id")
	var req dto.SchedulerRetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	simulateStatus := strings.ToUpper(strings.TrimSpace(req.SimulateStatus))
	if simulateStatus != "" && h.cfg.AllowJobSimulation {
		id, err := h.service.AdminRetrySchedulerJobRun(runID, "MANUAL", simulateStatus, req.ResultSummary, req.ErrorMessage, operator)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		h.writeOperationLog(c, "SCHEDULER", "RETRY_JOB", "JOB_RUN", runID, "", simulateStatus, req.ResultSummary)
		c.JSON(http.StatusOK, dto.OK(gin.H{"id": id, "status": simulateStatus}))
		return
	}
	jobName, err := h.service.GetSchedulerJobNameByRunID(runID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "job run not found", Data: struct{}{}})
		return
	}
	syncOptions := h.buildTushareNewsSyncOptions(req.NewsSources, req.Symbols, req.SyncTypes, req.BatchSize)
	execResult, runErr := h.runSchedulerJob(jobName, syncOptions)
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	resultSummary := execResult.Summary
	id, err := h.service.AdminRetrySchedulerJobRun(runID, "MANUAL", status, resultSummary, errorMessage, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if len(execResult.NewsSyncDetails) > 0 {
		if detailErr := h.service.AdminCreateNewsSyncRunDetails(id, execResult.NewsSyncDetails); detailErr != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: detailErr.Error(), Data: struct{}{}})
			return
		}
	}
	finalRunID, finalStatus, finalSummary, finalError, retryAttempts, retryErr := h.executeSchedulerAutoRetry(
		jobName,
		id,
		status,
		resultSummary,
		errorMessage,
		operator,
		syncOptions,
	)
	if retryErr != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: retryErr.Error(), Data: struct{}{}})
		return
	}
	reason := resultSummary
	if retryAttempts > 0 {
		reason = fmt.Sprintf("auto_retry=%d, %s", retryAttempts, strings.TrimSpace(resultSummary))
	}
	h.writeOperationLog(c, "SCHEDULER", "RETRY_JOB", "JOB_RUN", runID, "", finalStatus, reason)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"id":                 finalRunID,
		"status":             finalStatus,
		"first_run_id":       id,
		"retry_attempts":     retryAttempts,
		"result_summary":     finalSummary,
		"error_message":      finalError,
		"auto_retry_applied": retryAttempts > 0,
	}))
}

func (h *AdminSystemHandler) RetryNewsSyncItem(c *gin.Context) {
	runID := strings.TrimSpace(c.Param("id"))
	if runID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "run id is required", Data: struct{}{}})
		return
	}
	var req dto.RetryNewsSyncItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	jobName, err := h.service.GetSchedulerJobNameByRunID(runID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "job run not found", Data: struct{}{}})
		return
	}
	if !strings.EqualFold(strings.TrimSpace(jobName), "tushare_news_incremental") {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "only tushare_news_incremental supports fine-grained retry", Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)

	syncOptions := h.buildTushareNewsSyncOptions(
		[]string{req.Source},
		[]string{req.Symbol},
		[]string{req.SyncType},
		req.BatchSize,
	)
	execResult, runErr := h.runSchedulerJob(jobName, syncOptions)
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	resultSummary := execResult.Summary
	if resultSummary == "" {
		resultSummary = "retry single news sync item"
	}

	newRunID, err := h.service.AdminRetrySchedulerJobRun(runID, "MANUAL", status, resultSummary, errorMessage, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if len(execResult.NewsSyncDetails) > 0 {
		if detailErr := h.service.AdminCreateNewsSyncRunDetails(newRunID, execResult.NewsSyncDetails); detailErr != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: detailErr.Error(), Data: struct{}{}})
			return
		}
	}
	h.writeOperationLog(
		c,
		"SCHEDULER",
		"RETRY_NEWS_SYNC_ITEM",
		"JOB_RUN",
		runID,
		"",
		status,
		fmt.Sprintf("sync_type=%s source=%s symbol=%s", strings.TrimSpace(req.SyncType), strings.TrimSpace(req.Source), strings.TrimSpace(req.Symbol)),
	)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"id":             newRunID,
		"status":         status,
		"result_summary": resultSummary,
		"error_message":  errorMessage,
	}))
}

func (h *AdminSystemHandler) SchedulerJobMetrics(c *gin.Context) {
	jobName := c.Query("job_name")
	item, err := h.service.AdminGetSchedulerJobMetrics(jobName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminSystemHandler) ListNewsSyncRunDetails(c *gin.Context) {
	runID := strings.TrimSpace(c.Param("id"))
	if runID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "run id is required", Data: struct{}{}})
		return
	}
	page, pageSize := utils.ParsePage(c)
	syncType := c.Query("sync_type")
	source := c.Query("source")
	symbol := c.Query("symbol")
	status := c.Query("status")
	items, total, err := h.service.AdminListNewsSyncRunDetails(runID, syncType, source, symbol, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminSystemHandler) ExportSchedulerJobRunsCSV(c *gin.Context) {
	jobName := c.Query("job_name")
	status := c.Query("status")
	items, _, err := h.service.AdminListSchedulerJobRuns(jobName, status, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "parent_run_id", "retry_count", "job_name", "trigger_source", "status", "started_at", "finished_at", "result_summary", "error_message", "operator_id"})
	for _, it := range items {
		_ = writer.Write([]string{it.ID, it.ParentRunID, fmt.Sprintf("%d", it.RetryCount), it.JobName, it.TriggerSource, it.Status, it.StartedAt, it.FinishedAt, it.ResultSummary, it.ErrorMessage, it.OperatorID})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=admin_scheduler_job_runs.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminSystemHandler) GetExperimentAnalyticsSummary(c *gin.Context) {
	daysStr := strings.TrimSpace(c.DefaultQuery("days", "7"))
	days := 7
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
		days = d
	}
	if days > 30 {
		days = 30
	}
	summary, err := h.service.AdminGetExperimentAnalyticsSummary(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(summary))
}

func (h *AdminSystemHandler) TestOSSQiniuConfig(c *gin.Context) {
	cfg := h.resolveOSSUploadConfig()
	if !strings.EqualFold(strings.TrimSpace(cfg.Provider), "QINIU") {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "oss provider is not QINIU", Data: struct{}{}})
		return
	}
	if !cfg.Enabled {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "oss is disabled", Data: struct{}{}})
		return
	}
	objectKey := utils.JoinObjectPath(
		cfg.PathPrefix,
		"probe",
		time.Now().Format("20060102"),
		fmt.Sprintf("probe_%d_%s.txt", time.Now().Unix(), utils.RandomHex(3)),
	)
	payload := []byte(fmt.Sprintf("sercherai oss qiniu probe %s", time.Now().Format(time.RFC3339)))
	fileURL, err := h.uploadToQiniu(cfg, objectKey, payload, "text/plain")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "TEST_OSS_QINIU", "SYSTEM_CONFIG", "oss.provider", "", "SUCCESS", objectKey)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"provider":   "QINIU",
		"region":     cfg.Region,
		"bucket":     cfg.Bucket,
		"object_key": objectKey,
		"file_url":   fileURL,
		"size":       len(payload),
	}))
}

func (h *AdminSystemHandler) TestYolkPayConfig(c *gin.Context) {
	cfg, err := resolveYolkPayConfig(h.service)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if !cfg.PaymentEnabled {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "payment.enabled 未开启", Data: struct{}{}})
		return
	}
	if !cfg.Enabled {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "payment.channel.yolkpay.enabled 未开启", Data: struct{}{}})
		return
	}
	if strings.TrimSpace(cfg.PID) == "" || strings.TrimSpace(cfg.Key) == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "payment.channel.yolkpay.pid 或 key 未配置", Data: struct{}{}})
		return
	}

	queryEndpoint := strings.TrimRight(buildYolkPayGatewayURL(cfg.Gateway, "/api.php"), "/")
	values := url.Values{}
	values.Set("act", "query")
	values.Set("pid", strings.TrimSpace(cfg.PID))
	values.Set("key", strings.TrimSpace(cfg.Key))
	requestURL := queryEndpoint + "?" + values.Encode()

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	resp, err := (&http.Client{Timeout: 12 * time.Second}).Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, dto.APIResponse{Code: 50201, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 128*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(string(bodyBytes))
		if message == "" {
			message = resp.Status
		}
		c.JSON(http.StatusBadGateway, dto.APIResponse{Code: 50201, Message: message, Data: struct{}{}})
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		c.JSON(http.StatusBadGateway, dto.APIResponse{Code: 50201, Message: "蛋黄支付返回格式异常", Data: struct{}{}})
		return
	}
	if parseYolkPayCode(payload["code"]) != 1 {
		msg := stringifyYolkPayValue(payload["msg"])
		if msg == "" {
			msg = "蛋黄支付配置校验失败"
		}
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: msg, Data: payload})
		return
	}
	result := gin.H{
		"gateway": queryEndpoint,
		"pid":     stringifyYolkPayValue(payload["pid"]),
		"active":  stringifyYolkPayValue(payload["active"]),
		"money":   stringifyYolkPayValue(payload["money"]),
		"orders":  stringifyYolkPayValue(payload["orders"]),
		"raw":     payload,
	}
	pid := stringifyYolkPayValue(result["pid"])
	h.writeOperationLog(c, "SYSTEM", "TEST_PAYMENT_YOLKPAY", "SYSTEM_CONFIG", "payment.channel.yolkpay.pid", "", "SUCCESS", pid)
	c.JSON(http.StatusOK, dto.OK(result))
}

// Helpers
func (h *AdminSystemHandler) buildTushareNewsSyncOptions(newsSources []string, symbols []string, syncTypes []string, batchSize int) model.TushareNewsSyncOptions {
	opts := model.TushareNewsSyncOptions{
		BatchSize: batchSize,
		Sources:   uniqueNonEmptyStrings(newsSources),
		Symbols:   normalizeUpperValues(symbols),
		SyncTypes: normalizeUpperValues(syncTypes),
	}
	if opts.BatchSize < 0 {
		opts.BatchSize = 0
	}
	return opts
}

func (h *AdminSystemHandler) executeSchedulerAutoRetry(jobName string, baseRunID string, baseStatus string, baseSummary string, baseError string, operator string, syncOptions model.TushareNewsSyncOptions) (string, string, string, string, int, error) {
	finalRunID := strings.TrimSpace(baseRunID)
	finalStatus := strings.ToUpper(strings.TrimSpace(baseStatus))
	finalSummary := baseSummary
	finalError := baseError
	retryAttempts := 0
	if finalStatus != "FAILED" || finalRunID == "" {
		return finalRunID, finalStatus, finalSummary, finalError, retryAttempts, nil
	}
	policy := h.resolveSchedulerAutoRetryPolicy(jobName)
	if !policy.Enabled {
		return finalRunID, finalStatus, finalSummary, finalError, retryAttempts, nil
	}
	currentRunID := finalRunID
	for attempt := 1; attempt <= policy.MaxRetries && finalStatus == "FAILED"; attempt++ {
		if policy.BackoffSeconds > 0 {
			time.Sleep(time.Duration(policy.BackoffSeconds*attempt) * time.Second)
		}
		execResult, runErr := h.runSchedulerJob(jobName, syncOptions)
		summary := execResult.Summary
		status := "SUCCESS"
		errorMessage := ""
		if runErr != nil {
			status = "FAILED"
			errorMessage = runErr.Error()
		}
		newRunID, createErr := h.service.AdminRetrySchedulerJobRun(currentRunID, "SYSTEM", status, summary, errorMessage, operator)
		if createErr != nil {
			return finalRunID, finalStatus, finalSummary, finalError, retryAttempts, createErr
		}
		if len(execResult.NewsSyncDetails) > 0 {
			if detailErr := h.service.AdminCreateNewsSyncRunDetails(newRunID, execResult.NewsSyncDetails); detailErr != nil {
				return finalRunID, finalStatus, finalSummary, finalError, retryAttempts, detailErr
			}
		}
		retryAttempts = attempt
		currentRunID = newRunID
		finalRunID = newRunID
		finalStatus = status
		finalSummary = summary
		finalError = errorMessage
	}
	return finalRunID, finalStatus, finalSummary, finalError, retryAttempts, nil
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

func (h *AdminSystemHandler) resolveSchedulerAutoRetryPolicy(jobName string) schedulerAutoRetryPolicy {
	policy := schedulerAutoRetryPolicy{
		Enabled:        strings.EqualFold(strings.TrimSpace(jobName), schedulerAutoRetryDefaultJob),
		MaxRetries:     2,
		BackoffSeconds: 2,
	}
	items, _, err := h.service.AdminListSystemConfigs("scheduler.auto_retry", 1, 200)
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

func (h *AdminSystemHandler) runSchedulerJob(jobName string, syncOptions model.TushareNewsSyncOptions) (schedulerJobExecutionResult, error) {
	switch strings.ToLower(strings.TrimSpace(jobName)) {
	case "daily_stock_quant_pipeline":
		tradeDate := time.Now().Format("2006-01-02")
		sourceKey := strings.ToUpper(strings.TrimSpace(h.resolveDefaultStockQuoteSourceKey()))
		if sourceKey == "" {
			sourceKey = stockDefaultSourceFallback
		}
		usedSourceKey := sourceKey
		quoteCount, err := h.service.AdminSyncStockQuotes(sourceKey, nil, 180)
		if err != nil && sourceKey != "MOCK" {
			fallbackCount, fallbackErr := h.service.AdminSyncStockQuotes("MOCK", nil, 180)
			if fallbackErr != nil {
				return schedulerJobExecutionResult{}, fmt.Errorf("sync quotes failed(%s): %v, fallback MOCK failed: %w", sourceKey, err, fallbackErr)
			}
			quoteCount = fallbackCount
			usedSourceKey = "MOCK"
		}
		if err != nil && sourceKey == "MOCK" {
			return schedulerJobExecutionResult{}, err
		}
		topItems, err := h.service.AdminGetQuantTopStocks(10, 180)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		recoResult, err := h.service.AdminGenerateDailyStockRecommendations(tradeDate)
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
		result, err := h.service.AdminGenerateDailyStockRecommendations(tradeDate)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{Summary: fmt.Sprintf("generated %d recommendations", result.Count)}, nil
	case schedulerJobDailyFuturesStrategy, schedulerJobFuturesStrategyGenerate:
		tradeDate := time.Now().Format("2006-01-02")
		result, err := h.service.AdminGenerateDailyFuturesStrategies(tradeDate)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{
			Summary: fmt.Sprintf("trade_date=%s generated=%d", tradeDate, result.Count),
		}, nil
	case schedulerJobFuturesStrategyEvaluate:
		return schedulerJobExecutionResult{Summary: "Futures strategy evaluation triggered"}, nil
	case "doc_fast_news_incremental":
		summary, err := h.service.AdminSyncDocFastNewsIncremental(0)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{Summary: summary}, nil
	case "tushare_news_incremental":
		summary, details, err := h.service.AdminSyncTushareNewsIncrementalWithOptions(syncOptions)
		if err != nil {
			return schedulerJobExecutionResult{Summary: summary, NewsSyncDetails: details}, err
		}
		return schedulerJobExecutionResult{Summary: summary, NewsSyncDetails: details}, nil
	case "vip_membership_lifecycle":
		summary, err := h.service.AdminRunVIPMembershipLifecycle()
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{Summary: summary}, nil
	default:
		return schedulerJobExecutionResult{}, fmt.Errorf("unsupported job: %s", jobName)
	}
}

func (h *AdminSystemHandler) resolveDefaultStockQuoteSourceKey() string {
	items, _, err := h.service.AdminListSystemConfigs(stockDefaultSourceConfigKey, 1, 10)
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

func maskSystemConfigValueForAudit(configKey string, configValue string) string {
	if isSensitiveSystemConfigKey(configKey) {
		return maskSensitiveSystemConfigValue(configValue)
	}
	return configValue
}

func sanitizeSystemConfigItems(items []model.SystemConfig, includeSensitive bool) []model.SystemConfig {
	result := make([]model.SystemConfig, 0, len(items))
	for _, item := range items {
		result = append(result, sanitizeSystemConfigItem(item, includeSensitive))
	}
	return result
}

func sanitizeSystemConfigItem(item model.SystemConfig, includeSensitive bool) model.SystemConfig {
	sanitized := item
	if !includeSensitive && isSensitiveSystemConfigKey(item.ConfigKey) {
		sanitized.ConfigValue = maskSensitiveSystemConfigValue(item.ConfigValue)
	}
	return sanitized
}


func isSensitiveSystemConfigKey(rawKey string) bool {
	key := strings.ToLower(strings.TrimSpace(rawKey))
	if key == "" {
		return false
	}
	if strings.Contains(key, "source_key") || strings.Contains(key, "policy_key") {
		return false
	}
	if strings.Contains(key, "secret") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "password") ||
		strings.Contains(key, "private_key") ||
		strings.Contains(key, "access_key") ||
		strings.Contains(key, "api_v3_key") {
		return true
	}
	return strings.HasSuffix(key, ".key") || strings.HasSuffix(key, "_key")
}

func maskSensitiveSystemConfigValue(rawValue string) string {
	trimmed := strings.TrimSpace(rawValue)
	if trimmed == "" {
		return "-"
	}
	if len(trimmed) <= 6 {
		return fmt.Sprintf("****** (%d chars)", len(rawValue))
	}
	return fmt.Sprintf("%s***%s (%d chars)", trimmed[:2], trimmed[len(trimmed)-2:], len(rawValue))
}

func (h *AdminSystemHandler) uploadToQiniu(cfg ossUploadConfig, objectKey string, payload []byte, mimeType string) (string, error) {
	qiniuCfg := oss.QiniuConfig{
		AccessKey: cfg.AccessKey,
		SecretKey: cfg.SecretKey,
		Bucket:    cfg.Bucket,
		Domain:    cfg.Domain,
		Region:    cfg.Region,
		UseHTTPS:  cfg.UseHTTPS,
	}
	return oss.UploadToQiniu(qiniuCfg, objectKey, payload, mimeType)
}

func (h *AdminSystemHandler) writeOperationLog(c *gin.Context, module string, action string, targetType string, targetID string, beforeValue string, afterValue string, reason string) {
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = "admin_unknown"
	}
	_ = h.service.AdminCreateOperationLog(module, action, targetType, targetID, operator, beforeValue, afterValue, reason)
}
