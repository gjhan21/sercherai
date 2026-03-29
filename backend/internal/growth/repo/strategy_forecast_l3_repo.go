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
	COALESCE(executive_summary, ''),
	COALESCE(primary_scenario, ''),
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
	id, run_id, version, executive_summary, primary_scenario,
	alternative_scenarios_json, trigger_checklist_json, invalidation_signals_json,
	role_disagreements_json, action_guidance_json, markdown_body, html_body,
	summary_json, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	executive_summary = VALUES(executive_summary),
	primary_scenario = VALUES(primary_scenario),
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
			nullableString(result.Report.ExecutiveSummary),
			nullableString(result.Report.PrimaryScenario),
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
	now := time.Now().UTC()
	runID := newID("l3run")
	engineKey := strings.TrimSpace(config.DefaultEngineKey)
	if engineKey == "" {
		engineKey = model.StrategyForecastL3EngineLocalSynthesis
	}
	run := model.StrategyForecastL3Run{
		ID:             runID,
		TargetType:     targetType,
		TargetID:       strings.TrimSpace(input.TargetID),
		TargetKey:      targetKey,
		TargetLabel:    strings.TrimSpace(input.TargetLabel),
		TriggerType:    triggerType,
		RequestUserID:  strings.TrimSpace(input.RequestUserID),
		OperatorUserID: strings.TrimSpace(input.OperatorUserID),
		EngineKey:      engineKey,
		Status:         model.StrategyForecastL3StatusQueued,
		PriorityScore:  input.PriorityScore,
		Reason:         strings.TrimSpace(input.Reason),
		ContextMeta:    cloneStringAnyMap(input.ContextMeta),
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
		PriorityScore:   run.PriorityScore,
		ReportAvailable: false,
	}
	return run, now, nil
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
		&item.ExecutiveSummary,
		&item.PrimaryScenario,
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
