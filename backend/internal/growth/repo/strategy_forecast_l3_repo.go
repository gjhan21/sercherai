package repo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

var errStrategyForecastL3InvalidInput = errors.New("invalid strategy forecast l3 input")

func (r *MySQLGrowthRepo) CreateStrategyForecastL3Run(input model.StrategyForecastL3RunCreateInput) (model.StrategyForecastL3Run, error) {
	config := r.loadForecastL3RuntimeConfig()
	run, now, err := buildStrategyForecastL3QueuedRun(input, config)
	if err != nil {
		return model.StrategyForecastL3Run{}, err
	}
	if !config.TriggerEnabled(run.TriggerType) {
		return model.StrategyForecastL3Run{}, fmt.Errorf("forecast l3 trigger is disabled: %s", run.TriggerType)
	}
	if activeCount, err := r.countStrategyForecastL3RunsByStatuses(model.StrategyForecastL3StatusQueued, model.StrategyForecastL3StatusRunning); err == nil {
		if activeCount >= config.MaxActiveRuns {
			return model.StrategyForecastL3Run{}, fmt.Errorf("forecast l3 active run limit reached")
		}
	} else {
		return model.StrategyForecastL3Run{}, err
	}
	if todayCount, err := r.countStrategyForecastL3RunsCreatedToday(""); err == nil {
		if todayCount >= config.MaxRunsPerDay {
			return model.StrategyForecastL3Run{}, fmt.Errorf("forecast l3 daily run limit reached")
		}
	} else {
		return model.StrategyForecastL3Run{}, err
	}
	if run.TriggerType == model.StrategyForecastL3TriggerTypeUserRequest && strings.TrimSpace(run.RequestUserID) != "" {
		allowed, err := r.CheckAndConsumeQuota(run.RequestUserID, "forecast", run.ID)
		if err != nil {
			return model.StrategyForecastL3Run{}, err
		}
		if !allowed {
			return model.StrategyForecastL3Run{}, errors.New("quota exceeded")
		}

		_, _ = r.db.Exec(`
INSERT INTO browse_histories (id, user_id, content_type, content_id, source_page, viewed_at)
VALUES (?, ?, 'FORECAST', ?, '/forecast', ?)`,
			newID("bh"), run.RequestUserID, run.ID, now,
		)

		userCount, err := r.countStrategyForecastL3RunsCreatedToday(run.RequestUserID)
		if err != nil {
			return model.StrategyForecastL3Run{}, err
		}
		if userCount >= config.MaxUserRunsPerDay {
			return model.StrategyForecastL3Run{}, fmt.Errorf("forecast l3 user daily run limit reached")
		}
	}
	if _, err := r.db.Exec(`
INSERT INTO strategy_forecast_l3_runs (
	id, target_type, target_id, target_key, target_label, trigger_type,
	request_user_id, operator_user_id, engine_key, status, priority_score,
	reason, failure_reason, context_meta_json, summary_json, report_ref_json,
	queued_at, started_at, finished_at, cancelled_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID,
		run.TargetType,
		nullableString(run.TargetID),
		run.TargetKey,
		nullableString(run.TargetLabel),
		run.TriggerType,
		nullableString(run.RequestUserID),
		nullableString(run.OperatorUserID),
		run.EngineKey,
		run.Status,
		run.PriorityScore,
		nullableString(run.Reason),
		nil,
		nullableForecastL3JSON(marshalJSONText(run.ContextMeta)),
		marshalJSONText(run.Summary),
		nil,
		now,
		nil,
		nil,
		nil,
		now,
		now,
	); err != nil {
		return model.StrategyForecastL3Run{}, err
	}
	return run, nil
}

func (r *MySQLGrowthRepo) ListStrategyForecastL3Runs(requestUserID string, status string, targetType string, triggerType string, page int, pageSize int) ([]model.StrategyForecastL3Run, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	whereClause, args := buildStrategyForecastL3RunFilters(requestUserID, status, targetType, triggerType)

	countQuery := `SELECT COUNT(*) FROM strategy_forecast_l3_runs` + whereClause
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	queryArgs := append([]interface{}{}, args...)
	offset := (page - 1) * pageSize
	queryArgs = append(queryArgs, pageSize, offset)
	rows, err := r.db.Query(`
SELECT
	id,
	target_type,
	COALESCE(target_id, ''),
	target_key,
	COALESCE(target_label, ''),
	trigger_type,
	COALESCE(request_user_id, ''),
	COALESCE(operator_user_id, ''),
	engine_key,
	status,
	priority_score,
	COALESCE(reason, ''),
	COALESCE(failure_reason, ''),
	COALESCE(CAST(context_meta_json AS CHAR), ''),
	COALESCE(CAST(summary_json AS CHAR), ''),
	COALESCE(CAST(report_ref_json AS CHAR), ''),
	queued_at,
	started_at,
	finished_at,
	cancelled_at,
	created_at,
	updated_at
FROM strategy_forecast_l3_runs`+whereClause+`
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.StrategyForecastL3Run, 0)
	for rows.Next() {
		item, err := scanStrategyForecastL3Run(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) GetLatestStrategyForecastL3Run(targetType string, targetID string) (model.StrategyForecastL3Run, error) {
	row := r.db.QueryRow(`
SELECT
	id,
	target_type,
	COALESCE(target_id, ''),
	target_key,
	COALESCE(target_label, ''),
	trigger_type,
	COALESCE(request_user_id, ''),
	COALESCE(operator_user_id, ''),
	engine_key,
	status,
	priority_score,
	COALESCE(reason, ''),
	COALESCE(failure_reason, ''),
	COALESCE(CAST(context_meta_json AS CHAR), ''),
	COALESCE(CAST(summary_json AS CHAR), ''),
	COALESCE(CAST(report_ref_json AS CHAR), ''),
	queued_at,
	started_at,
	finished_at,
	cancelled_at,
	created_at,
	updated_at
FROM strategy_forecast_l3_runs
WHERE target_type = ? AND target_id = ?
ORDER BY created_at DESC, id DESC
LIMIT 1`, targetType, targetID)

	return scanStrategyForecastL3Run(row)
}

func (r *MySQLGrowthRepo) GetStrategyForecastL3Run(runID string) (model.StrategyForecastL3Run, error) {
	row := r.db.QueryRow(`
SELECT
	id,
	target_type,
	COALESCE(target_id, ''),
	target_key,
	COALESCE(target_label, ''),
	trigger_type,
	COALESCE(request_user_id, ''),
	COALESCE(operator_user_id, ''),
	engine_key,
	status,
	priority_score,
	COALESCE(reason, ''),
	COALESCE(failure_reason, ''),
	COALESCE(CAST(context_meta_json AS CHAR), ''),
	COALESCE(CAST(summary_json AS CHAR), ''),
	COALESCE(CAST(report_ref_json AS CHAR), ''),
	queued_at,
	started_at,
	finished_at,
	cancelled_at,
	created_at,
	updated_at
FROM strategy_forecast_l3_runs
WHERE id = ?`, strings.TrimSpace(runID))
	return scanStrategyForecastL3Run(row)
}

func (r *MySQLGrowthRepo) GetStrategyForecastL3RunDetail(runID string) (model.StrategyForecastL3RunDetail, error) {
	run, err := r.GetStrategyForecastL3Run(runID)
	if err != nil {
		return model.StrategyForecastL3RunDetail{}, err
	}
	report, err := r.getLatestStrategyForecastL3Report(runID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.StrategyForecastL3RunDetail{}, err
	}
	logs, err := r.ListStrategyForecastL3Logs(runID)
	if err != nil {
		return model.StrategyForecastL3RunDetail{}, err
	}
	detail := model.StrategyForecastL3RunDetail{
		Run:  run,
		Logs: logs,
	}
	if err == nil {
		detail.Report = &report
	}
	return detail, nil
}

func (r *MySQLGrowthRepo) GetStrategyForecastL3RunDetailForUser(runID string, userID string) (model.StrategyForecastL3RunDetail, error) {
	detail, err := r.GetStrategyForecastL3RunDetail(runID)
	if err != nil {
		return model.StrategyForecastL3RunDetail{}, err
	}
	return r.applyStrategyForecastL3UserReadPolicy(detail, userID)
}

func (r *MySQLGrowthRepo) ListStrategyForecastL3HistoryForTarget(requestUserID string, targetType string, targetKey string, page int, pageSize int) ([]model.StrategyForecastL3HistoryItem, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	normalizedType := normalizeStrategyForecastL3TargetType(targetType)
	trimmedKey := strings.TrimSpace(targetKey)
	if normalizedType == "" || trimmedKey == "" {
		return []model.StrategyForecastL3HistoryItem{}, nil
	}

	whereClause := ` WHERE target_type = ? AND target_key = ? AND status = ?`
	args := []interface{}{normalizedType, trimmedKey, model.StrategyForecastL3StatusSucceeded}
	if userID := strings.TrimSpace(requestUserID); userID != "" {
		whereClause += ` AND request_user_id = ?`
		args = append(args, userID)
	}
	queryArgs := append([]interface{}{}, args...)
	offset := (page - 1) * pageSize
	queryArgs = append(queryArgs, pageSize, offset)
	rows, err := r.db.Query(`
SELECT
	id,
	target_type,
	COALESCE(target_id, ''),
	target_key,
	COALESCE(target_label, ''),
	trigger_type,
	COALESCE(request_user_id, ''),
	COALESCE(operator_user_id, ''),
	engine_key,
	status,
	priority_score,
	COALESCE(reason, ''),
	COALESCE(failure_reason, ''),
	COALESCE(CAST(context_meta_json AS CHAR), ''),
	COALESCE(CAST(summary_json AS CHAR), ''),
	COALESCE(CAST(report_ref_json AS CHAR), ''),
	queued_at,
	started_at,
	finished_at,
	cancelled_at,
	created_at,
	updated_at
FROM strategy_forecast_l3_runs`+whereClause+`
ORDER BY finished_at DESC, created_at DESC, id DESC
LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyForecastL3HistoryItem, 0)
	for rows.Next() {
		run, err := scanStrategyForecastL3Run(rows)
		if err != nil {
			return nil, err
		}
		item := model.StrategyForecastL3HistoryItem{
			RunID:            run.ID,
			TargetType:       run.TargetType,
			TargetKey:        run.TargetKey,
			TargetLabel:      run.TargetLabel,
			Status:           run.Status,
			CreatedAt:        run.CreatedAt,
			FinishedAt:       run.FinishedAt,
			ValidationStatus: run.ValidationStatus,
			ContextQuality:   run.ContextQuality,
		}
		report, err := r.getLatestStrategyForecastL3Report(run.ID)
		if err == nil {
			item.HeadlineVerdict = report.HeadlineVerdict
			item.PrimaryScenario = firstNonEmpty(report.PrimaryScenario, run.Summary.PrimaryScenario)
			if report.StateAssessment != nil {
				item.CurrentState = report.StateAssessment.CurrentState
			}
		}
		if item.HeadlineVerdict == "" {
			item.HeadlineVerdict = firstNonEmpty(run.Summary.ExecutiveSummary, run.Summary.ActionGuidance)
		}
		if item.PrimaryScenario == "" {
			item.PrimaryScenario = run.Summary.PrimaryScenario
		}
		if review, err := r.GetStrategyForecastL3RunReview(run.ID, requestUserID); err == nil {
			item.Review = &review
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MySQLGrowthRepo) GetStrategyForecastL3HistoryCompare(requestUserID string, targetType string, targetKey string, leftRunID string, rightRunID string) (model.StrategyForecastL3HistoryCompare, error) {
	items, err := r.ListStrategyForecastL3HistoryForTarget(requestUserID, targetType, targetKey, 1, 50)
	if err != nil {
		return model.StrategyForecastL3HistoryCompare{}, err
	}
	return buildStrategyForecastL3HistoryCompare(normalizeStrategyForecastL3TargetType(targetType), strings.TrimSpace(targetKey), items, leftRunID, rightRunID)
}

func (r *MySQLGrowthRepo) GetStrategyForecastL3RunReview(runID string, requestUserID string) (model.StrategyForecastL3RunReview, error) {
	detail, err := r.GetStrategyForecastL3RunDetail(runID)
	if err != nil {
		return model.StrategyForecastL3RunReview{}, err
	}
	if userID := strings.TrimSpace(requestUserID); userID != "" {
		if strings.TrimSpace(detail.Run.RequestUserID) != "" && detail.Run.RequestUserID != userID {
			return model.StrategyForecastL3RunReview{}, sql.ErrNoRows
		}
	}
	records, err := r.loadStrategyForecastL3LearningRecordsForRun(runID)
	if err != nil {
		return model.StrategyForecastL3RunReview{}, err
	}
	if len(records) == 0 {
		return model.StrategyForecastL3RunReview{}, sql.ErrNoRows
	}
	return buildStrategyForecastL3RunReview(records[0]), nil
}

func (r *MySQLGrowthRepo) RetryStrategyForecastL3Run(runID string, operatorUserID string, reason string) (model.StrategyForecastL3Run, error) {
	run, err := r.GetStrategyForecastL3Run(runID)
	if err != nil {
		return model.StrategyForecastL3Run{}, err
	}
	now := time.Now().UTC()
	run.Status = model.StrategyForecastL3StatusQueued
	run.OperatorUserID = strings.TrimSpace(operatorUserID)
	run.Reason = coalesceTrimmed(reason, run.Reason)
	run.FailureReason = ""
	run.ReportRef = nil
	run.StartedAt = ""
	run.FinishedAt = ""
	run.CancelledAt = ""
	run.QueuedAt = now.Format(time.RFC3339)
	run.UpdatedAt = now.Format(time.RFC3339)
	run.Summary.Status = run.Status
	run.Summary.ReportAvailable = false
	run.Summary.GeneratedAt = ""
	if _, err := r.db.Exec(`
UPDATE strategy_forecast_l3_runs
SET status = ?,
	operator_user_id = ?,
	reason = ?,
	failure_reason = NULL,
	summary_json = ?,
	report_ref_json = NULL,
	queued_at = ?,
	started_at = NULL,
	finished_at = NULL,
	cancelled_at = NULL,
	updated_at = ?
WHERE id = ?`,
		run.Status,
		nullableString(run.OperatorUserID),
		nullableString(run.Reason),
		marshalJSONText(run.Summary),
		now,
		now,
		run.ID,
	); err != nil {
		return model.StrategyForecastL3Run{}, err
	}
	return run, nil
}

func (r *MySQLGrowthRepo) CancelStrategyForecastL3Run(runID string, operatorUserID string, reason string) (model.StrategyForecastL3Run, error) {
	run, err := r.GetStrategyForecastL3Run(runID)
	if err != nil {
		return model.StrategyForecastL3Run{}, err
	}
	now := time.Now().UTC()
	run.Status = model.StrategyForecastL3StatusCancelled
	run.OperatorUserID = strings.TrimSpace(operatorUserID)
	run.Reason = coalesceTrimmed(reason, run.Reason)
	run.CancelledAt = now.Format(time.RFC3339)
	run.FinishedAt = now.Format(time.RFC3339)
	run.UpdatedAt = now.Format(time.RFC3339)
	run.Summary.Status = run.Status
	if _, err := r.db.Exec(`
UPDATE strategy_forecast_l3_runs
SET status = ?,
	operator_user_id = ?,
	reason = ?,
	cancelled_at = ?,
	finished_at = ?,
	updated_at = ?
WHERE id = ?`,
		run.Status,
		nullableString(run.OperatorUserID),
		nullableString(run.Reason),
		now,
		now,
		now,
		run.ID,
	); err != nil {
		return model.StrategyForecastL3Run{}, err
	}
	return run, nil
}

func (r *MySQLGrowthRepo) ListStrategyForecastL3Logs(runID string) ([]model.StrategyForecastL3Log, error) {
	rows, err := r.db.Query(`
SELECT
	id,
	run_id,
	step_key,
	status,
	COALESCE(message, ''),
	COALESCE(CAST(payload_json AS CHAR), ''),
	created_at
FROM strategy_forecast_l3_logs
WHERE run_id = ?
ORDER BY created_at ASC, id ASC`, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyForecastL3Log, 0)
	for rows.Next() {
		item, err := scanStrategyForecastL3Log(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MySQLGrowthRepo) ListStrategyForecastL3QualitySummaries(targetType string, days int) ([]model.StrategyForecastL3QualitySummary, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	items, err := r.loadStrategyForecastL3LearningRecords(targetType, days)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "doesn't exist") {
			return nil, nil
		}
		return nil, err
	}
	return summarizeStrategyForecastL3LearningRecords(items), nil
}

func (r *MySQLGrowthRepo) ExecuteQueuedStrategyForecastL3Runs(limit int, operatorUserID string) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	config := r.loadForecastL3RuntimeConfig()
	if !config.Enabled || !config.DispatchEnabled {
		return 0, nil
	}
	items, err := r.listQueuedStrategyForecastL3Runs(limit)
	if err != nil {
		return 0, err
	}
	processed := 0
	for _, item := range items {
		result := executeStrategyForecastL3Run(r, item)
		if err := r.persistMySQLStrategyForecastL3Execution(result, operatorUserID); err != nil {
			return processed, err
		}
		processed++
	}
	return processed, nil
}

func (r *MySQLGrowthRepo) RunStrategyForecastL3QualityBackfill(limit int, operatorUserID string) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	config := r.loadForecastL3RuntimeConfig()
	if !config.Enabled || !config.QualityEnabled {
		return 0, nil
	}
	items, total, err := r.ListStrategyForecastL3Runs("", model.StrategyForecastL3StatusSucceeded, "", "", 1, limit*3)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	created := 0
	for _, item := range items {
		if created >= limit {
			break
		}
		exists, err := r.hasStrategyForecastL3LearningRecord(item.ID)
		if err != nil {
			return created, err
		}
		if exists {
			continue
		}
		detail, err := r.GetStrategyForecastL3RunDetail(item.ID)
		if err != nil {
			return created, err
		}
		record := buildStrategyForecastL3LearningRecord(detail, time.Now().UTC())
		if err := r.insertStrategyForecastL3LearningRecord(record); err != nil {
			return created, err
		}
		created++
	}
	return created, nil
}

func (r *InMemoryGrowthRepo) CreateStrategyForecastL3Run(input model.StrategyForecastL3RunCreateInput) (model.StrategyForecastL3Run, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, _, err := buildStrategyForecastL3QueuedRun(input, defaultForecastL3RuntimeConfig)
	if err != nil {
		return model.StrategyForecastL3Run{}, err
	}
	r.forecastL3Runs[run.ID] = run
	return run, nil
}

func (r *InMemoryGrowthRepo) ListStrategyForecastL3Runs(requestUserID string, status string, targetType string, triggerType string, page int, pageSize int) ([]model.StrategyForecastL3Run, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	filtered := make([]model.StrategyForecastL3Run, 0, len(r.forecastL3Runs))
	for _, item := range r.forecastL3Runs {
		if !strategyForecastL3RunMatches(item, requestUserID, status, targetType, triggerType) {
			continue
		}
		filtered = append(filtered, cloneStrategyForecastL3Run(item))
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].CreatedAt == filtered[j].CreatedAt {
			return filtered[i].ID > filtered[j].ID
		}
		return filtered[i].CreatedAt > filtered[j].CreatedAt
	})
	total := len(filtered)
	start := (page - 1) * pageSize
	if start >= total {
		return []model.StrategyForecastL3Run{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return filtered[start:end], total, nil
}

func (r *InMemoryGrowthRepo) GetStrategyForecastL3Run(runID string) (model.StrategyForecastL3Run, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.forecastL3Runs[strings.TrimSpace(runID)]
	if !ok {
		return model.StrategyForecastL3Run{}, sql.ErrNoRows
	}
	return cloneStrategyForecastL3Run(item), nil
}

func (r *InMemoryGrowthRepo) GetStrategyForecastL3RunDetail(runID string) (model.StrategyForecastL3RunDetail, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, ok := r.forecastL3Runs[strings.TrimSpace(runID)]
	if !ok {
		return model.StrategyForecastL3RunDetail{}, sql.ErrNoRows
	}
	detail := model.StrategyForecastL3RunDetail{
		Run:  cloneStrategyForecastL3Run(run),
		Logs: cloneStrategyForecastL3Logs(r.forecastL3Logs[run.ID]),
	}
	if report, ok := r.forecastL3Reports[run.ID]; ok {
		reportCopy := cloneStrategyForecastL3Report(report)
		detail.Report = &reportCopy
	}
	return detail, nil
}

func (r *InMemoryGrowthRepo) GetStrategyForecastL3RunDetailForUser(runID string, userID string) (model.StrategyForecastL3RunDetail, error) {
	return r.GetStrategyForecastL3RunDetail(runID)
}

func (r *InMemoryGrowthRepo) ListStrategyForecastL3HistoryForTarget(requestUserID string, targetType string, targetKey string, page int, pageSize int) ([]model.StrategyForecastL3HistoryItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := buildStrategyForecastL3HistoryItemsFromInMemory(r, requestUserID, targetType, targetKey)
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []model.StrategyForecastL3HistoryItem{}, nil
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	cloned := make([]model.StrategyForecastL3HistoryItem, end-start)
	for index, item := range items[start:end] {
		cloned[index] = cloneStrategyForecastL3HistoryItem(item)
	}
	return cloned, nil
}

func (r *InMemoryGrowthRepo) GetStrategyForecastL3HistoryCompare(requestUserID string, targetType string, targetKey string, leftRunID string, rightRunID string) (model.StrategyForecastL3HistoryCompare, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	items := buildStrategyForecastL3HistoryItemsFromInMemory(r, requestUserID, targetType, targetKey)
	return buildStrategyForecastL3HistoryCompare(normalizeStrategyForecastL3TargetType(targetType), strings.TrimSpace(targetKey), items, leftRunID, rightRunID)
}

func (r *InMemoryGrowthRepo) GetStrategyForecastL3RunReview(runID string, requestUserID string) (model.StrategyForecastL3RunReview, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	runID = strings.TrimSpace(runID)
	run, ok := r.forecastL3Runs[runID]
	if !ok {
		return model.StrategyForecastL3RunReview{}, sql.ErrNoRows
	}
	if userID := strings.TrimSpace(requestUserID); userID != "" {
		if strings.TrimSpace(run.RequestUserID) != "" && run.RequestUserID != userID {
			return model.StrategyForecastL3RunReview{}, sql.ErrNoRows
		}
	}
	records := r.forecastL3Learning[runID]
	if len(records) == 0 {
		return model.StrategyForecastL3RunReview{}, sql.ErrNoRows
	}
	return buildStrategyForecastL3RunReview(records[0]), nil
}

func (r *InMemoryGrowthRepo) RetryStrategyForecastL3Run(runID string, operatorUserID string, reason string) (model.StrategyForecastL3Run, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, ok := r.forecastL3Runs[strings.TrimSpace(runID)]
	if !ok {
		return model.StrategyForecastL3Run{}, sql.ErrNoRows
	}
	now := time.Now().UTC().Format(time.RFC3339)
	run.Status = model.StrategyForecastL3StatusQueued
	run.OperatorUserID = strings.TrimSpace(operatorUserID)
	run.Reason = coalesceTrimmed(reason, run.Reason)
	run.FailureReason = ""
	run.ReportRef = nil
	run.QueuedAt = now
	run.StartedAt = ""
	run.FinishedAt = ""
	run.CancelledAt = ""
	run.UpdatedAt = now
	run.Summary.Status = run.Status
	run.Summary.GeneratedAt = ""
	run.Summary.ReportAvailable = false
	r.forecastL3Runs[run.ID] = run
	delete(r.forecastL3Reports, run.ID)
	delete(r.forecastL3Logs, run.ID)
	return cloneStrategyForecastL3Run(run), nil
}

func (r *InMemoryGrowthRepo) CancelStrategyForecastL3Run(runID string, operatorUserID string, reason string) (model.StrategyForecastL3Run, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	run, ok := r.forecastL3Runs[strings.TrimSpace(runID)]
	if !ok {
		return model.StrategyForecastL3Run{}, sql.ErrNoRows
	}
	now := time.Now().UTC().Format(time.RFC3339)
	run.Status = model.StrategyForecastL3StatusCancelled
	run.OperatorUserID = strings.TrimSpace(operatorUserID)
	run.Reason = coalesceTrimmed(reason, run.Reason)
	run.CancelledAt = now
	run.FinishedAt = now
	run.UpdatedAt = now
	run.Summary.Status = run.Status
	r.forecastL3Runs[run.ID] = run
	return cloneStrategyForecastL3Run(run), nil
}

func (r *InMemoryGrowthRepo) ListStrategyForecastL3Logs(runID string) ([]model.StrategyForecastL3Log, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return cloneStrategyForecastL3Logs(r.forecastL3Logs[strings.TrimSpace(runID)]), nil
}

func (r *InMemoryGrowthRepo) ListStrategyForecastL3QualitySummaries(targetType string, days int) ([]model.StrategyForecastL3QualitySummary, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	records := make([]model.StrategyForecastL3LearningRecord, 0)
	for _, items := range r.forecastL3Learning {
		for _, item := range items {
			if normalized := normalizeStrategyForecastL3TargetType(targetType); normalized != "" && item.TargetType != normalized {
				continue
			}
			records = append(records, item)
		}
	}
	return summarizeStrategyForecastL3LearningRecords(records), nil
}

func (r *InMemoryGrowthRepo) ExecuteQueuedStrategyForecastL3Runs(limit int, operatorUserID string) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	r.mu.Lock()
	queued := make([]model.StrategyForecastL3Run, 0, len(r.forecastL3Runs))
	for _, item := range r.forecastL3Runs {
		if item.Status == model.StrategyForecastL3StatusQueued {
			queued = append(queued, cloneStrategyForecastL3Run(item))
		}
	}
	sort.Slice(queued, func(i, j int) bool {
		if queued[i].CreatedAt == queued[j].CreatedAt {
			return queued[i].ID < queued[j].ID
		}
		return queued[i].CreatedAt < queued[j].CreatedAt
	})
	if limit < len(queued) {
		queued = queued[:limit]
	}
	r.mu.Unlock()

	processed := 0
	for _, item := range queued {
		result := executeStrategyForecastL3Run(r, item)
		r.mu.Lock()
		r.persistInMemoryStrategyForecastL3Execution(result)
		r.mu.Unlock()
		processed++
	}
	return processed, nil
}

func (r *InMemoryGrowthRepo) RunStrategyForecastL3QualityBackfill(limit int, operatorUserID string) (int, error) {
	if limit <= 0 {
		return 0, nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	keys := make([]string, 0, len(r.forecastL3Runs))
	for key := range r.forecastL3Runs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	created := 0
	for _, key := range keys {
		if created >= limit {
			break
		}
		run := r.forecastL3Runs[key]
		if run.Status != model.StrategyForecastL3StatusSucceeded {
			continue
		}
		if len(r.forecastL3Learning[key]) > 0 {
			continue
		}
		detail := model.StrategyForecastL3RunDetail{
			Run:  cloneStrategyForecastL3Run(run),
			Logs: cloneStrategyForecastL3Logs(r.forecastL3Logs[key]),
		}
		if report, ok := r.forecastL3Reports[key]; ok {
			reportCopy := cloneStrategyForecastL3Report(report)
			detail.Report = &reportCopy
		}
		record := buildStrategyForecastL3LearningRecord(detail, time.Now().UTC())
		r.forecastL3Learning[key] = append(r.forecastL3Learning[key], record)
		created++
	}
	return created, nil
}

func (r *MySQLGrowthRepo) getLatestStrategyForecastL3Report(runID string) (model.StrategyForecastL3Report, error) {
	row := r.db.QueryRow(`
SELECT
	id,
	run_id,
	version,
	COALESCE(headline_verdict, ''),
	COALESCE(executive_summary, ''),
	COALESCE(primary_scenario, ''),
	COALESCE(CAST(state_assessment_json AS CHAR), ''),
	COALESCE(CAST(dimension_evidence_json AS CHAR), ''),
	COALESCE(CAST(scenario_assessment_json AS CHAR), ''),
	COALESCE(CAST(validation_review_json AS CHAR), ''),
	COALESCE(CAST(alternative_scenarios_json AS CHAR), ''),
	COALESCE(CAST(trigger_checklist_json AS CHAR), ''),
	COALESCE(CAST(invalidation_signals_json AS CHAR), ''),
	COALESCE(CAST(role_disagreements_json AS CHAR), ''),
	COALESCE(CAST(action_guidance_json AS CHAR), ''),
	COALESCE(markdown_body, ''),
	COALESCE(html_body, ''),
	COALESCE(CAST(summary_json AS CHAR), ''),
	created_at,
	updated_at
FROM strategy_forecast_l3_reports
WHERE run_id = ?
ORDER BY version DESC
LIMIT 1`, strings.TrimSpace(runID))
	return scanStrategyForecastL3Report(row)
}

func (r *MySQLGrowthRepo) listQueuedStrategyForecastL3Runs(limit int) ([]model.StrategyForecastL3Run, error) {
	rows, err := r.db.Query(`
SELECT
	id,
	target_type,
	COALESCE(target_id, ''),
	target_key,
	COALESCE(target_label, ''),
	trigger_type,
	COALESCE(request_user_id, ''),
	COALESCE(operator_user_id, ''),
	engine_key,
	status,
	priority_score,
	COALESCE(reason, ''),
	COALESCE(failure_reason, ''),
	COALESCE(CAST(context_meta_json AS CHAR), ''),
	COALESCE(CAST(summary_json AS CHAR), ''),
	COALESCE(CAST(report_ref_json AS CHAR), ''),
	queued_at,
	started_at,
	finished_at,
	cancelled_at,
	created_at,
	updated_at
FROM strategy_forecast_l3_runs
WHERE status = ?
ORDER BY created_at ASC, id ASC
LIMIT ?`, model.StrategyForecastL3StatusQueued, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyForecastL3Run, 0)
	for rows.Next() {
		item, err := scanStrategyForecastL3Run(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MySQLGrowthRepo) countStrategyForecastL3RunsByStatuses(statuses ...string) (int, error) {
	if len(statuses) == 0 {
		return 0, nil
	}
	placeholders := make([]string, 0, len(statuses))
	args := make([]interface{}, 0, len(statuses))
	for _, status := range statuses {
		normalized := normalizeStrategyForecastL3Status(status)
		if normalized == "" {
			continue
		}
		placeholders = append(placeholders, "?")
		args = append(args, normalized)
	}
	if len(placeholders) == 0 {
		return 0, nil
	}
	var count int
	if err := r.db.QueryRow(
		`SELECT COUNT(*) FROM strategy_forecast_l3_runs WHERE status IN (`+strings.Join(placeholders, ",")+`)`,
		args...,
	).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MySQLGrowthRepo) countStrategyForecastL3RunsCreatedToday(requestUserID string) (int, error) {
	args := make([]interface{}, 0, 1)
	query := `SELECT COUNT(*) FROM strategy_forecast_l3_runs WHERE DATE(created_at) = CURDATE()`
	if strings.TrimSpace(requestUserID) != "" {
		query += ` AND request_user_id = ?`
		args = append(args, strings.TrimSpace(requestUserID))
	}
	var count int
	if err := r.db.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *MySQLGrowthRepo) persistMySQLStrategyForecastL3Execution(result strategyForecastL3ExecutionResult, operatorUserID string) error {
	if result.Report != nil {
		if _, err := r.db.Exec(`
INSERT INTO strategy_forecast_l3_reports (
	id, run_id, version, headline_verdict, executive_summary, primary_scenario,
	state_assessment_json, dimension_evidence_json, scenario_assessment_json, validation_review_json,
	alternative_scenarios_json, trigger_checklist_json, invalidation_signals_json,
	role_disagreements_json, action_guidance_json, markdown_body, html_body,
	summary_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	id = VALUES(id),
	created_at = VALUES(created_at),
	headline_verdict = VALUES(headline_verdict),
	executive_summary = VALUES(executive_summary),
	primary_scenario = VALUES(primary_scenario),
	state_assessment_json = VALUES(state_assessment_json),
	dimension_evidence_json = VALUES(dimension_evidence_json),
	scenario_assessment_json = VALUES(scenario_assessment_json),
	validation_review_json = VALUES(validation_review_json),
	alternative_scenarios_json = VALUES(alternative_scenarios_json),
	trigger_checklist_json = VALUES(trigger_checklist_json),
	invalidation_signals_json = VALUES(invalidation_signals_json),
	role_disagreements_json = VALUES(role_disagreements_json),
	action_guidance_json = VALUES(action_guidance_json),
	markdown_body = VALUES(markdown_body),
	html_body = VALUES(html_body),
	summary_json = VALUES(summary_json),
	updated_at = VALUES(updated_at)`,
			result.Report.ID,
			result.Report.RunID,
			result.Report.Version,
			nullableString(result.Report.HeadlineVerdict),
			nullableString(result.Report.ExecutiveSummary),
			nullableString(result.Report.PrimaryScenario),
			nullableForecastL3JSON(marshalJSONText(result.Report.StateAssessment)),
			nullableForecastL3JSON(marshalJSONText(result.Report.DimensionEvidence)),
			nullableForecastL3JSON(marshalJSONText(result.Report.ScenarioAssessment)),
			nullableForecastL3JSON(marshalJSONText(result.Report.ValidationReview)),
			nullableForecastL3JSON(marshalJSONText(result.Report.AlternativeScenarios)),
			nullableForecastL3JSON(marshalJSONText(result.Report.TriggerChecklist)),
			nullableForecastL3JSON(marshalJSONText(result.Report.InvalidationSignals)),
			nullableForecastL3JSON(marshalJSONText(result.Report.RoleDisagreements)),
			nullableForecastL3JSON(marshalJSONText(result.Report.ActionGuidance)),
			nullableString(result.Report.MarkdownBody),
			nullableString(result.Report.HTMLBody),
			nullableForecastL3JSON(marshalJSONText(result.Report.Summary)),
			parseForecastL3TimestampOrNow(result.Report.CreatedAt),
			parseForecastL3TimestampOrNow(result.Report.UpdatedAt),
		); err != nil {
			return err
		}
	}
	for _, item := range result.Logs {
		if _, err := r.db.Exec(`
INSERT INTO strategy_forecast_l3_logs (id, run_id, step_key, status, message, payload_json, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	status = VALUES(status),
	message = VALUES(message),
	payload_json = VALUES(payload_json),
	created_at = VALUES(created_at)`,
			item.ID,
			item.RunID,
			item.StepKey,
			item.Status,
			nullableString(item.Message),
			nullableForecastL3JSON(marshalJSONText(item.Payload)),
			parseForecastL3TimestampOrNow(item.CreatedAt),
		); err != nil {
			return err
		}
	}
	_, err := r.db.Exec(`
UPDATE strategy_forecast_l3_runs
SET operator_user_id = ?,
	engine_key = ?,
	status = ?,
	failure_reason = ?,
	summary_json = ?,
	report_ref_json = ?,
	started_at = ?,
	finished_at = ?,
	updated_at = ?
WHERE id = ?`,
		nullableString(firstNonEmpty(operatorUserID, result.Run.OperatorUserID)),
		result.Run.EngineKey,
		result.Run.Status,
		nullableString(result.Run.FailureReason),
		nullableForecastL3JSON(marshalJSONText(result.Run.Summary)),
		nullableForecastL3JSON(marshalJSONText(result.Run.ReportRef)),
		parseForecastL3TimestampOrNow(result.Run.StartedAt),
		parseForecastL3TimestampOrNull(result.Run.FinishedAt),
		parseForecastL3TimestampOrNow(result.Run.UpdatedAt),
		result.Run.ID,
	)
	return err
}

func (r *InMemoryGrowthRepo) persistInMemoryStrategyForecastL3Execution(result strategyForecastL3ExecutionResult) {
	r.forecastL3Runs[result.Run.ID] = cloneStrategyForecastL3Run(result.Run)
	if result.Report != nil {
		r.forecastL3Reports[result.Run.ID] = cloneStrategyForecastL3Report(*result.Report)
	}
	if len(result.Logs) > 0 {
		r.forecastL3Logs[result.Run.ID] = cloneStrategyForecastL3Logs(result.Logs)
	}
}

func buildStrategyForecastL3QueuedRun(input model.StrategyForecastL3RunCreateInput, config forecastL3RuntimeConfig) (model.StrategyForecastL3Run, time.Time, error) {
	targetType := normalizeStrategyForecastL3TargetType(input.TargetType)
	targetKey := strings.TrimSpace(input.TargetKey)
	if targetKey == "" {
		targetKey = strings.TrimSpace(input.TargetID)
	}
	if targetType == "" || targetKey == "" {
		return model.StrategyForecastL3Run{}, time.Time{}, errStrategyForecastL3InvalidInput
	}
	triggerType := normalizeStrategyForecastL3TriggerType(input.TriggerType)
	if triggerType == "" {
		triggerType = model.StrategyForecastL3TriggerTypeUserRequest
	}
	targetID := strings.TrimSpace(input.TargetID)
	source := strings.ToUpper(strings.TrimSpace(input.Source))
	sourceID := strings.TrimSpace(input.SourceID)
	sourcePath := strings.TrimSpace(input.SourcePath)
	contextQuality := model.StrategyForecastL3ContextQualityPartial
	if targetID != "" {
		contextQuality = model.StrategyForecastL3ContextQualityFull
	}
	if triggerType == model.StrategyForecastL3TriggerTypeUserRequest {
		if source == "" || sourcePath == "" {
			return model.StrategyForecastL3Run{}, time.Time{}, fmt.Errorf("%w: source is required for user-request deep forecast", errStrategyForecastL3InvalidInput)
		}
		if config.ContextRequireTargetID && targetID == "" {
			return model.StrategyForecastL3Run{}, time.Time{}, fmt.Errorf("%w: target_id is required for user-request deep forecast", errStrategyForecastL3InvalidInput)
		}
	}
	now := time.Now().UTC()
	runID := newID("l3run")
	engineKey := strings.TrimSpace(config.DefaultEngineKey)
	if engineKey == "" {
		engineKey = model.StrategyForecastL3EngineLocalSynthesis
	}
	contextMeta := cloneStringAnyMap(input.ContextMeta)
	if contextMeta == nil {
		contextMeta = map[string]any{}
	}
	if source != "" {
		contextMeta["source"] = source
	}
	if sourceID != "" {
		contextMeta["source_id"] = sourceID
	}
	if sourcePath != "" {
		contextMeta["source_path"] = sourcePath
	}
	contextMeta["context_quality"] = contextQuality
	validationStatus := model.StrategyForecastL3ValidationStatusSkipped
	if config.ValidationEnabled && (!config.ValidationUserRequestOnly || triggerType == model.StrategyForecastL3TriggerTypeUserRequest) {
		validationStatus = model.StrategyForecastL3ValidationStatusPending
	}
	run := model.StrategyForecastL3Run{
		ID:             runID,
		TargetType:     targetType,
		TargetID:       targetID,
		TargetKey:      targetKey,
		TargetLabel:    strings.TrimSpace(input.TargetLabel),
		Source:         source,
		TriggerType:    triggerType,
		RequestUserID:  strings.TrimSpace(input.RequestUserID),
		OperatorUserID: strings.TrimSpace(input.OperatorUserID),
		EngineKey:      engineKey,
		Status:         model.StrategyForecastL3StatusQueued,
		ContextQuality: contextQuality,
		ValidationStatus: validationStatus,
		PriorityScore:  input.PriorityScore,
		Reason:         strings.TrimSpace(input.Reason),
		ContextMeta:    contextMeta,
		QueuedAt:       now.Format(time.RFC3339),
		CreatedAt:      now.Format(time.RFC3339),
		UpdatedAt:      now.Format(time.RFC3339),
	}
	run.Summary = model.StrategyForecastL3Summary{
		RunID:           run.ID,
		Status:          run.Status,
		EngineKey:       run.EngineKey,
		TriggerType:     run.TriggerType,
		TargetType:      run.TargetType,
		TargetKey:       run.TargetKey,
		TargetLabel:     run.TargetLabel,
		Source:          run.Source,
		ContextQuality:  run.ContextQuality,
		ValidationStatus: run.ValidationStatus,
		PriorityScore:   run.PriorityScore,
		ReportAvailable: false,
	}
	return run, now, nil
}

func (r *MySQLGrowthRepo) applyStrategyForecastL3UserReadPolicy(detail model.StrategyForecastL3RunDetail, userID string) (model.StrategyForecastL3RunDetail, error) {
	if detail.Run.ReportRef == nil {
		return detail, nil
	}
	if !detail.Run.ReportRef.RequiresVIP {
		detail.Run.ReportRef.FullReadable = true
		return detail, nil
	}

	config := r.loadForecastL3RuntimeConfig()
	if !config.RequireVIPForFullReport {
		detail.Run.ReportRef.FullReadable = true
		return detail, nil
	}

	isVIP := false
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID != "" {
		var err error
		isVIP, err = r.isVIPUser(trimmedUserID)
		if err != nil {
			return model.StrategyForecastL3RunDetail{}, err
		}
	}
	if isVIP {
		detail.Run.ReportRef.FullReadable = true
		return detail, nil
	}

	detail.Run.ReportRef.FullReadable = false
	if detail.Report != nil {
		detail.Report.MarkdownBody = ""
		detail.Report.HTMLBody = ""
	}
	return detail, nil
}

func buildStrategyForecastL3RunFilters(requestUserID string, status string, targetType string, triggerType string) (string, []interface{}) {
	parts := make([]string, 0, 4)
	args := make([]interface{}, 0, 4)
	if userID := strings.TrimSpace(requestUserID); userID != "" {
		parts = append(parts, "request_user_id = ?")
		args = append(args, userID)
	}
	if normalized := normalizeStrategyForecastL3Status(status); normalized != "" {
		parts = append(parts, "status = ?")
		args = append(args, normalized)
	}
	if normalized := normalizeStrategyForecastL3TargetType(targetType); normalized != "" {
		parts = append(parts, "target_type = ?")
		args = append(args, normalized)
	}
	if normalized := normalizeStrategyForecastL3TriggerType(triggerType); normalized != "" {
		parts = append(parts, "trigger_type = ?")
		args = append(args, normalized)
	}
	if len(parts) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(parts, " AND "), args
}

func scanStrategyForecastL3Run(scanner interface {
	Scan(dest ...interface{}) error
}) (model.StrategyForecastL3Run, error) {
	var item model.StrategyForecastL3Run
	var contextMetaJSON sql.NullString
	var summaryJSON sql.NullString
	var reportRefJSON sql.NullString
	var queuedAt sql.NullTime
	var startedAt sql.NullTime
	var finishedAt sql.NullTime
	var cancelledAt sql.NullTime
	var createdAt time.Time
	var updatedAt time.Time
	if err := scanner.Scan(
		&item.ID,
		&item.TargetType,
		&item.TargetID,
		&item.TargetKey,
		&item.TargetLabel,
		&item.TriggerType,
		&item.RequestUserID,
		&item.OperatorUserID,
		&item.EngineKey,
		&item.Status,
		&item.PriorityScore,
		&item.Reason,
		&item.FailureReason,
		&contextMetaJSON,
		&summaryJSON,
		&reportRefJSON,
		&queuedAt,
		&startedAt,
		&finishedAt,
		&cancelledAt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return model.StrategyForecastL3Run{}, err
	}
	item.ContextMeta = parseStrategyForecastL3Map(contextMetaJSON.String)
	item.Summary = parseStrategyForecastL3Summary(summaryJSON.String)
	item.ReportRef = parseStrategyForecastL3ReportRef(reportRefJSON.String)
	item.QueuedAt = formatNullTime(queuedAt)
	item.StartedAt = formatNullTime(startedAt)
	item.FinishedAt = formatNullTime(finishedAt)
	item.CancelledAt = formatNullTime(cancelledAt)
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	if item.Summary.RunID == "" {
		item.Summary.RunID = item.ID
		item.Summary.Status = item.Status
		item.Summary.EngineKey = item.EngineKey
		item.Summary.TriggerType = item.TriggerType
		item.Summary.TargetType = item.TargetType
		item.Summary.TargetKey = item.TargetKey
		item.Summary.TargetLabel = item.TargetLabel
		item.Summary.PriorityScore = item.PriorityScore
		item.Summary.ReportAvailable = item.ReportRef != nil
	}
	return item, nil
}

func scanStrategyForecastL3Report(scanner interface {
	Scan(dest ...interface{}) error
}) (model.StrategyForecastL3Report, error) {
	var item model.StrategyForecastL3Report
	var stateAssessmentJSON sql.NullString
	var dimensionEvidenceJSON sql.NullString
	var scenarioAssessmentJSON sql.NullString
	var validationReviewJSON sql.NullString
	var alternativeScenariosJSON sql.NullString
	var triggerChecklistJSON sql.NullString
	var invalidationSignalsJSON sql.NullString
	var roleDisagreementsJSON sql.NullString
	var actionGuidanceJSON sql.NullString
	var summaryJSON sql.NullString
	var createdAt time.Time
	var updatedAt time.Time
	if err := scanner.Scan(
		&item.ID,
		&item.RunID,
		&item.Version,
		&item.HeadlineVerdict,
		&item.ExecutiveSummary,
		&item.PrimaryScenario,
		&stateAssessmentJSON,
		&dimensionEvidenceJSON,
		&scenarioAssessmentJSON,
		&validationReviewJSON,
		&alternativeScenariosJSON,
		&triggerChecklistJSON,
		&invalidationSignalsJSON,
		&roleDisagreementsJSON,
		&actionGuidanceJSON,
		&item.MarkdownBody,
		&item.HTMLBody,
		&summaryJSON,
		&createdAt,
		&updatedAt,
	); err != nil {
		return model.StrategyForecastL3Report{}, err
	}
	item.AlternativeScenarios = parseStrategyForecastL3ScenarioList(alternativeScenariosJSON.String)
	item.StateAssessment = parseStrategyForecastL3StateAssessment(stateAssessmentJSON.String)
	item.DimensionEvidence = parseStrategyForecastL3DimensionEvidenceList(dimensionEvidenceJSON.String)
	item.ScenarioAssessment = parseStrategyForecastL3ScenarioAssessment(scenarioAssessmentJSON.String)
	item.ValidationReview = parseStrategyForecastL3ValidationReview(validationReviewJSON.String)
	item.TriggerChecklist = parseStrategyForecastL3Checklist(triggerChecklistJSON.String)
	item.InvalidationSignals = parseStrategyForecastL3StringList(invalidationSignalsJSON.String)
	item.RoleDisagreements = parseStrategyForecastL3RoleDisagreements(roleDisagreementsJSON.String)
	item.ActionGuidance = parseStrategyForecastL3StringList(actionGuidanceJSON.String)
	item.Summary = parseStrategyForecastL3Summary(summaryJSON.String)
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
	if item.Summary.RunID == "" {
		item.Summary.RunID = item.RunID
		item.Summary.Status = model.StrategyForecastL3StatusSucceeded
		item.Summary.ExecutiveSummary = item.ExecutiveSummary
		item.Summary.PrimaryScenario = item.PrimaryScenario
		item.Summary.ReportAvailable = true
	}
	return item, nil
}

func scanStrategyForecastL3Log(scanner interface {
	Scan(dest ...interface{}) error
}) (model.StrategyForecastL3Log, error) {
	var item model.StrategyForecastL3Log
	var payloadJSON sql.NullString
	var createdAt time.Time
	if err := scanner.Scan(
		&item.ID,
		&item.RunID,
		&item.StepKey,
		&item.Status,
		&item.Message,
		&payloadJSON,
		&createdAt,
	); err != nil {
		return model.StrategyForecastL3Log{}, err
	}
	item.Payload = parseStrategyForecastL3Map(payloadJSON.String)
	item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
	return item, nil
}

func normalizeStrategyForecastL3TargetType(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case model.StrategyForecastL3TargetTypeStock:
		return model.StrategyForecastL3TargetTypeStock
	case model.StrategyForecastL3TargetTypeFutures:
		return model.StrategyForecastL3TargetTypeFutures
	default:
		return ""
	}
}

func normalizeStrategyForecastL3TriggerType(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case model.StrategyForecastL3TriggerTypeAdminManual:
		return model.StrategyForecastL3TriggerTypeAdminManual
	case model.StrategyForecastL3TriggerTypeAutoPriority:
		return model.StrategyForecastL3TriggerTypeAutoPriority
	case model.StrategyForecastL3TriggerTypeUserRequest:
		return model.StrategyForecastL3TriggerTypeUserRequest
	default:
		return ""
	}
}

func normalizeStrategyForecastL3Status(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case model.StrategyForecastL3StatusQueued:
		return model.StrategyForecastL3StatusQueued
	case model.StrategyForecastL3StatusRunning:
		return model.StrategyForecastL3StatusRunning
	case model.StrategyForecastL3StatusSucceeded:
		return model.StrategyForecastL3StatusSucceeded
	case model.StrategyForecastL3StatusFailed:
		return model.StrategyForecastL3StatusFailed
	case model.StrategyForecastL3StatusCancelled:
		return model.StrategyForecastL3StatusCancelled
	default:
		return ""
	}
}

func parseStrategyForecastL3Map(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var value map[string]any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil
	}
	return value
}

func parseStrategyForecastL3Summary(raw string) model.StrategyForecastL3Summary {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return model.StrategyForecastL3Summary{}
	}
	var item model.StrategyForecastL3Summary
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return model.StrategyForecastL3Summary{}
	}
	return item
}

func parseStrategyForecastL3ReportRef(raw string) *model.StrategyForecastL3ReportRef {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var item model.StrategyForecastL3ReportRef
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return nil
	}
	return &item
}

func parseStrategyForecastL3ScenarioList(raw string) []model.StrategyForecastL3Scenario {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []model.StrategyForecastL3Scenario
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func parseStrategyForecastL3StateAssessment(raw string) *model.StrategyForecastL3StateAssessment {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var item model.StrategyForecastL3StateAssessment
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return nil
	}
	return &item
}

func parseStrategyForecastL3DimensionEvidenceList(raw string) []model.StrategyForecastL3DimensionEvidence {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []model.StrategyForecastL3DimensionEvidence
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func parseStrategyForecastL3ScenarioAssessment(raw string) *model.StrategyForecastL3ScenarioAssessment {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var item model.StrategyForecastL3ScenarioAssessment
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return nil
	}
	return &item
}

func parseStrategyForecastL3ValidationReview(raw string) *model.StrategyForecastL3ValidationReview {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var item model.StrategyForecastL3ValidationReview
	if err := json.Unmarshal([]byte(raw), &item); err != nil {
		return nil
	}
	return &item
}

func parseStrategyForecastL3Checklist(raw string) []model.StrategyForecastL3ChecklistItem {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []model.StrategyForecastL3ChecklistItem
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func parseStrategyForecastL3RoleDisagreements(raw string) []model.StrategyForecastL3RoleDisagreement {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []model.StrategyForecastL3RoleDisagreement
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func parseStrategyForecastL3StringList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func nullableForecastL3JSON(body string) interface{} {
	body = strings.TrimSpace(body)
	if body == "" || body == "null" || body == "{}" || body == "[]" {
		return nil
	}
	return body
}

func cloneStringAnyMap(value map[string]any) map[string]any {
	if len(value) == 0 {
		return nil
	}
	body, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	var cloned map[string]any
	if err := json.Unmarshal(body, &cloned); err != nil {
		return nil
	}
	return cloned
}

func cloneStrategyForecastL3Run(item model.StrategyForecastL3Run) model.StrategyForecastL3Run {
	item.ContextMeta = cloneStringAnyMap(item.ContextMeta)
	if item.ReportRef != nil {
		reportRef := *item.ReportRef
		item.ReportRef = &reportRef
	}
	return item
}

func cloneStrategyForecastL3Report(item model.StrategyForecastL3Report) model.StrategyForecastL3Report {
	if len(item.AlternativeScenarios) > 0 {
		cloned := make([]model.StrategyForecastL3Scenario, len(item.AlternativeScenarios))
		copy(cloned, item.AlternativeScenarios)
		item.AlternativeScenarios = cloned
	}
	if len(item.TriggerChecklist) > 0 {
		cloned := make([]model.StrategyForecastL3ChecklistItem, len(item.TriggerChecklist))
		copy(cloned, item.TriggerChecklist)
		item.TriggerChecklist = cloned
	}
	if len(item.InvalidationSignals) > 0 {
		cloned := make([]string, len(item.InvalidationSignals))
		copy(cloned, item.InvalidationSignals)
		item.InvalidationSignals = cloned
	}
	if len(item.RoleDisagreements) > 0 {
		cloned := make([]model.StrategyForecastL3RoleDisagreement, len(item.RoleDisagreements))
		copy(cloned, item.RoleDisagreements)
		item.RoleDisagreements = cloned
	}
	if len(item.ActionGuidance) > 0 {
		cloned := make([]string, len(item.ActionGuidance))
		copy(cloned, item.ActionGuidance)
		item.ActionGuidance = cloned
	}
	return item
}

func cloneStrategyForecastL3Logs(items []model.StrategyForecastL3Log) []model.StrategyForecastL3Log {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]model.StrategyForecastL3Log, len(items))
	for i, item := range items {
		cloned[i] = item
		cloned[i].Payload = cloneStringAnyMap(item.Payload)
	}
	return cloned
}

func strategyForecastL3RunMatches(item model.StrategyForecastL3Run, requestUserID string, status string, targetType string, triggerType string) bool {
	if userID := strings.TrimSpace(requestUserID); userID != "" && item.RequestUserID != userID {
		return false
	}
	if normalized := normalizeStrategyForecastL3Status(status); normalized != "" && item.Status != normalized {
		return false
	}
	if normalized := normalizeStrategyForecastL3TargetType(targetType); normalized != "" && item.TargetType != normalized {
		return false
	}
	if normalized := normalizeStrategyForecastL3TriggerType(triggerType); normalized != "" && item.TriggerType != normalized {
		return false
	}
	return true
}

func buildStrategyForecastL3HistoryItemsFromInMemory(r *InMemoryGrowthRepo, requestUserID string, targetType string, targetKey string) []model.StrategyForecastL3HistoryItem {
	normalizedType := normalizeStrategyForecastL3TargetType(targetType)
	trimmedKey := strings.TrimSpace(targetKey)
	trimmedUserID := strings.TrimSpace(requestUserID)
	items := make([]model.StrategyForecastL3HistoryItem, 0)
	for _, run := range r.forecastL3Runs {
		if run.Status != model.StrategyForecastL3StatusSucceeded {
			continue
		}
		if trimmedUserID != "" && strings.TrimSpace(run.RequestUserID) != "" && run.RequestUserID != trimmedUserID {
			continue
		}
		if normalizedType != "" && run.TargetType != normalizedType {
			continue
		}
		if trimmedKey != "" && run.TargetKey != trimmedKey {
			continue
		}
		item := model.StrategyForecastL3HistoryItem{
			RunID:            run.ID,
			TargetType:       run.TargetType,
			TargetKey:        run.TargetKey,
			TargetLabel:      run.TargetLabel,
			Status:           run.Status,
			CreatedAt:        run.CreatedAt,
			FinishedAt:       run.FinishedAt,
			ValidationStatus: run.ValidationStatus,
			ContextQuality:   run.ContextQuality,
		}
		if report, ok := r.forecastL3Reports[run.ID]; ok {
			item.HeadlineVerdict = report.HeadlineVerdict
			item.PrimaryScenario = firstNonEmpty(report.PrimaryScenario, run.Summary.PrimaryScenario)
			if report.StateAssessment != nil {
				item.CurrentState = report.StateAssessment.CurrentState
			}
		}
		if item.HeadlineVerdict == "" {
			item.HeadlineVerdict = firstNonEmpty(run.Summary.ExecutiveSummary, run.Summary.ActionGuidance)
		}
		if item.PrimaryScenario == "" {
			item.PrimaryScenario = run.Summary.PrimaryScenario
		}
		if records := r.forecastL3Learning[run.ID]; len(records) > 0 {
			review := buildStrategyForecastL3RunReview(records[0])
			item.Review = &review
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		leftTime := firstNonEmpty(items[i].FinishedAt, items[i].CreatedAt)
		rightTime := firstNonEmpty(items[j].FinishedAt, items[j].CreatedAt)
		if leftTime == rightTime {
			return items[i].RunID > items[j].RunID
		}
		return leftTime > rightTime
	})
	return items
}

func buildStrategyForecastL3HistoryCompare(targetType string, targetKey string, items []model.StrategyForecastL3HistoryItem, leftRunID string, rightRunID string) (model.StrategyForecastL3HistoryCompare, error) {
	compare := model.StrategyForecastL3HistoryCompare{
		TargetType:     targetType,
		TargetKey:      strings.TrimSpace(targetKey),
		TimelineLength: len(items),
	}
	if len(items) == 0 {
		return compare, nil
	}
	left := resolveStrategyForecastL3HistoryItem(items, leftRunID, 0)
	right := resolveStrategyForecastL3HistoryItem(items, rightRunID, 1)
	if left == nil || right == nil {
		return compare, nil
	}
	leftClone := cloneStrategyForecastL3HistoryItem(*left)
	rightClone := cloneStrategyForecastL3HistoryItem(*right)
	compare.LeftRun = &leftClone
	compare.RightRun = &rightClone
	compare.VerdictShift = buildStrategyForecastL3VerdictShift(leftClone, rightClone)
	compare.EvidenceDiffs = buildStrategyForecastL3EvidenceDiffs(leftClone, rightClone)
	compare.ReviewSummary = buildStrategyForecastL3ReviewSummary(leftClone, rightClone)
	return compare, nil
}

func resolveStrategyForecastL3HistoryItem(items []model.StrategyForecastL3HistoryItem, runID string, fallbackIndex int) *model.StrategyForecastL3HistoryItem {
	trimmed := strings.TrimSpace(runID)
	if trimmed != "" {
		for index := range items {
			if items[index].RunID == trimmed {
				return &items[index]
			}
		}
	}
	if fallbackIndex < 0 || fallbackIndex >= len(items) {
		return nil
	}
	return &items[fallbackIndex]
}

func buildStrategyForecastL3VerdictShift(left model.StrategyForecastL3HistoryItem, right model.StrategyForecastL3HistoryItem) []string {
	shifts := make([]string, 0, 3)
	if left.HeadlineVerdict != right.HeadlineVerdict {
		shifts = append(shifts, fmt.Sprintf("核心判断由“%s”切换到“%s”。", fallbackForecastText(right.HeadlineVerdict), fallbackForecastText(left.HeadlineVerdict)))
	}
	if left.PrimaryScenario != right.PrimaryScenario {
		shifts = append(shifts, fmt.Sprintf("主情景由“%s”变为“%s”。", fallbackForecastText(right.PrimaryScenario), fallbackForecastText(left.PrimaryScenario)))
	}
	if len(shifts) == 0 {
		shifts = append(shifts, "最新一次与上一次的结论整体保持一致。")
	}
	return shifts
}

func buildStrategyForecastL3EvidenceDiffs(left model.StrategyForecastL3HistoryItem, right model.StrategyForecastL3HistoryItem) []model.StrategyForecastL3EvidenceDiff {
	leftReview := left.Review
	rightReview := right.Review
	if leftReview == nil && rightReview == nil {
		return nil
	}
	diffs := make([]model.StrategyForecastL3EvidenceDiff, 0)
	keys := map[string]struct{}{}
	if leftReview != nil {
		for key := range leftReview.RoleEffectiveness {
			keys[key] = struct{}{}
		}
	}
	if rightReview != nil {
		for key := range rightReview.RoleEffectiveness {
			keys[key] = struct{}{}
		}
	}
	sortedKeys := make([]string, 0, len(keys))
	for key := range keys {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Strings(sortedKeys)
	for _, key := range sortedKeys {
		leftValue, leftOK := 0.0, false
		rightValue, rightOK := 0.0, false
		if leftReview != nil {
			leftValue, leftOK = leftReview.RoleEffectiveness[key]
		}
		if rightReview != nil {
			rightValue, rightOK = rightReview.RoleEffectiveness[key]
		}
		if !leftOK && !rightOK {
			continue
		}
		diffs = append(diffs, model.StrategyForecastL3EvidenceDiff{
			Dimension:    key,
			ChangeLabel:  buildStrategyForecastL3NumericChangeLabel(leftValue, rightValue, leftOK, rightOK),
			Previous:     formatStrategyForecastL3Effectiveness(rightValue, rightOK),
			Current:      formatStrategyForecastL3Effectiveness(leftValue, leftOK),
			PreviousNote: buildStrategyForecastL3EffectivenessNote(rightValue, rightOK),
			CurrentNote:  buildStrategyForecastL3EffectivenessNote(leftValue, leftOK),
		})
	}
	return diffs
}

func buildStrategyForecastL3ReviewSummary(left model.StrategyForecastL3HistoryItem, right model.StrategyForecastL3HistoryItem) []string {
	if left.Review == nil || right.Review == nil {
		return nil
	}
	delta := left.Review.ReviewScore - right.Review.ReviewScore
	switch {
	case delta > 0:
		return []string{fmt.Sprintf("完整复盘评分由 %d 提升到 %d。", right.Review.ReviewScore, left.Review.ReviewScore)}
	case delta < 0:
		return []string{fmt.Sprintf("完整复盘评分由 %d 回落到 %d。", right.Review.ReviewScore, left.Review.ReviewScore)}
	default:
		return []string{fmt.Sprintf("完整复盘评分维持在 %d。", left.Review.ReviewScore)}
	}
}

func buildStrategyForecastL3ChangeLabel(current string, previous string) string {
	current = strings.TrimSpace(current)
	previous = strings.TrimSpace(previous)
	switch {
	case current == "" && previous == "":
		return "未提供"
	case current == previous:
		return "基本不变"
	case previous == "":
		return "新增"
	case current == "":
		return "弱化"
	default:
		return "变化"
	}
}

func buildStrategyForecastL3NumericChangeLabel(current float64, previous float64, currentOK bool, previousOK bool) string {
	switch {
	case !currentOK && !previousOK:
		return "未提供"
	case currentOK && previousOK && current == previous:
		return "基本不变"
	case !previousOK && currentOK:
		return "新增"
	case previousOK && !currentOK:
		return "弱化"
	default:
		return "变化"
	}
}

func formatStrategyForecastL3Effectiveness(value float64, ok bool) string {
	if !ok {
		return ""
	}
	return fmt.Sprintf("%.2f", value)
}

func buildStrategyForecastL3EffectivenessNote(value float64, ok bool) string {
	if !ok {
		return "未提供历史有效性样本"
	}
	switch {
	case value >= 0.8:
		return "该维度在历史复盘中有效性较强"
	case value >= 0.6:
		return "该维度在历史复盘中有效性中等"
	default:
		return "该维度在历史复盘中有效性偏弱"
	}
}

func cloneStrategyForecastL3HistoryItem(item model.StrategyForecastL3HistoryItem) model.StrategyForecastL3HistoryItem {
	if item.Review != nil {
		review := *item.Review
		review.ReviewNotes = append([]string(nil), review.ReviewNotes...)
		review.RoleEffectiveness = cloneForecastL3RoleEffectiveness(review.RoleEffectiveness)
		item.Review = &review
	}
	return item
}

func fallbackForecastText(value string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}
	return "未给出"
}

func coalesceTrimmed(primary string, fallback string) string {
	if trimmed := strings.TrimSpace(primary); trimmed != "" {
		return trimmed
	}
	return strings.TrimSpace(fallback)
}

func parseForecastL3TimestampOrNow(raw string) time.Time {
	if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(raw)); err == nil {
		return parsed.UTC()
	}
	return time.Now().UTC()
}

func parseForecastL3TimestampOrNull(raw string) interface{} {
	if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(raw)); err == nil {
		return parsed.UTC()
	}
	return nil
}
