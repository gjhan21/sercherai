package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminListSystemConfigs(keyword string, page int, pageSize int) ([]model.SystemConfig, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if keyword != "" {
		like := "%" + keyword + "%"
		filter += " AND (config_key LIKE ? OR description LIKE ?)"
		args = append(args, like, like)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM system_configs"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, config_key, config_value, description, updated_by, updated_at
FROM system_configs` + filter + `
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.SystemConfig, 0)
	for rows.Next() {
		var item model.SystemConfig
		var desc sql.NullString
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.ConfigKey, &item.ConfigValue, &desc, &item.UpdatedBy, &updatedAt); err != nil {
			return nil, 0, err
		}
		if desc.Valid {
			item.Description = desc.String
		}
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminUpsertSystemConfig(configKey string, configValue string, description string, operator string) error {
	if operator == "" {
		operator = "admin_unknown"
	}
	_, err := r.db.Exec(`
INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	config_value = VALUES(config_value),
	description = VALUES(description),
	updated_by = VALUES(updated_by),
	updated_at = VALUES(updated_at)`,
		newID("cfg"), configKey, configValue, description, operator, time.Now(),
	)
	return err
}

func (r *MySQLGrowthRepo) AdminListReviewTasks(module string, status string, submitterID string, reviewerID string, page int, pageSize int) ([]model.ReviewTask, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if status != "" {
		if strings.Contains(status, ",") {
			filter += " AND FIND_IN_SET(status, ?) > 0"
		} else {
			filter += " AND status = ?"
		}
		args = append(args, strings.ToUpper(status))
	}
	if submitterID != "" {
		filter += " AND submitter_id = ?"
		args = append(args, submitterID)
	}
	if reviewerID != "" {
		if strings.ToUpper(reviewerID) == "NULL" {
			filter += " AND (reviewer_id IS NULL OR reviewer_id = '')"
		} else {
			filter += " AND reviewer_id = ?"
			args = append(args, reviewerID)
		}
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM review_tasks"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, module, target_id, submitter_id, reviewer_id, status, submit_note, review_note, submitted_at, reviewed_at
FROM review_tasks` + filter + `
ORDER BY submitted_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.ReviewTask, 0)
	for rows.Next() {
		var item model.ReviewTask
		var reviewerID, submitNote, reviewNote sql.NullString
		var submittedAt time.Time
		var reviewedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Module, &item.TargetID, &item.SubmitterID, &reviewerID, &item.Status, &submitNote, &reviewNote, &submittedAt, &reviewedAt); err != nil {
			return nil, 0, err
		}
		if reviewerID.Valid {
			item.ReviewerID = reviewerID.String
		}
		if submitNote.Valid {
			item.SubmitNote = submitNote.String
		}
		if reviewNote.Valid {
			item.ReviewNote = reviewNote.String
		}
		item.SubmittedAt = submittedAt.Format(time.RFC3339)
		if reviewedAt.Valid {
			item.ReviewedAt = reviewedAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminSubmitReviewTask(module string, targetID string, submitterID string, reviewerID string, submitNote string) (string, error) {
	module = strings.ToUpper(strings.TrimSpace(module))
	if module == "" {
		return "", errors.New("module is required")
	}
	var pendingCount int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM review_tasks WHERE module = ? AND target_id = ? AND status = 'PENDING'", module, targetID).Scan(&pendingCount); err != nil {
		return "", err
	}
	if pendingCount > 0 {
		return "", errors.New("pending review task already exists")
	}

	// Move target into REVIEWING state when submitting for review.
	if err := r.applyModuleTargetStatus(r.db, module, targetID, "REVIEWING"); err != nil {
		return "", err
	}

	id := newID("rt")
	now := time.Now()
	_, err := r.db.Exec(`
INSERT INTO review_tasks (id, module, target_id, submitter_id, reviewer_id, status, submit_note, submitted_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 'PENDING', ?, ?, ?, ?)`,
		id, module, targetID, submitterID, reviewerID, submitNote, now, now, now,
	)
	if err != nil {
		return "", err
	}
	_ = r.AdminCreateWorkflowMessage(
		id,
		targetID,
		module,
		reviewerID,
		submitterID,
		"REVIEW_SUBMITTED",
		"内容已提交审核",
		"模块 "+module+" 的目标 "+targetID+" 已提交审核",
	)
	return id, nil
}

func (r *MySQLGrowthRepo) AdminAssignReviewTask(reviewID string, reviewerID string) error {
	_, err := r.db.Exec(
		"UPDATE review_tasks SET reviewer_id = ?, updated_at = ? WHERE id = ? AND status = 'PENDING'",
		reviewerID, time.Now(), reviewID,
	)
	if err != nil {
		return err
	}
	_ = r.AdminCreateWorkflowMessage(
		reviewID,
		reviewID,
		"WORKFLOW",
		reviewerID,
		"",
		"REVIEW_ASSIGNED",
		"审核任务已分配",
		"你有新的审核任务 "+reviewID,
	)
	return nil
}

func (r *MySQLGrowthRepo) AdminReviewTaskDecision(reviewID string, status string, reviewerID string, reviewNote string) error {
	status = strings.ToUpper(strings.TrimSpace(status))
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var module, targetID, currentStatus, submitterID string
	var assignedReviewer sql.NullString
	err = tx.QueryRow("SELECT module, target_id, submitter_id, reviewer_id, status FROM review_tasks WHERE id = ? FOR UPDATE", reviewID).
		Scan(&module, &targetID, &submitterID, &assignedReviewer, &currentStatus)
	if err != nil {
		return err
	}
	if strings.ToUpper(currentStatus) != "PENDING" {
		return errors.New("review task is not pending")
	}
	if assignedReviewer.Valid && strings.TrimSpace(assignedReviewer.String) != "" && assignedReviewer.String != reviewerID {
		return errors.New("reviewer mismatch")
	}

	targetStatus := "DRAFT"
	if status == "APPROVED" {
		targetStatus = "PUBLISHED"
	}
	if err = r.applyModuleTargetStatus(tx, module, targetID, targetStatus); err != nil {
		return err
	}

	_, err = tx.Exec(`
UPDATE review_tasks
SET status = ?, reviewer_id = ?, review_note = ?, reviewed_at = ?, updated_at = ?
WHERE id = ?`,
		status, reviewerID, reviewNote, time.Now(), time.Now(), reviewID,
	)
	if err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return err
	}
	level := "INFO"
	if status != "APPROVED" {
		level = "WARNING"
	}
	_ = r.AdminCreateAuditEvent(model.AdminAuditEvent{
		EventDomain: "RESEARCH",
		EventType:   "REVIEW_" + status,
		Level:       level,
		Module:      strings.ToUpper(strings.TrimSpace(module)),
		ObjectType:  "REVIEW_TASK",
		ObjectID:    reviewID,
		ActorUserID: strings.TrimSpace(reviewerID),
		Title:       "审核结果通知",
		Summary:     "审核结果为 " + status + "，目标 " + targetID,
		Detail:      truncateByRunes(normalizeUTF8Text(reviewNote), 512),
		Status:      "OPEN",
		Metadata: map[string]any{
			"review_id":     reviewID,
			"target_id":     targetID,
			"target_module": module,
			"submitter_id":  submitterID,
			"review_status": status,
		},
	})
	return nil
}

func (r *MySQLGrowthRepo) GetSchedulerJobNameByRunID(runID string) (string, error) {
	var jobName string
	if err := r.db.QueryRow("SELECT job_name FROM scheduler_job_runs WHERE id = ?", runID).Scan(&jobName); err != nil {
		return "", err
	}
	return jobName, nil
}

func (r *MySQLGrowthRepo) AdminListSchedulerJobRuns(jobName string, status string, page int, pageSize int) ([]model.SchedulerJobRun, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if jobName != "" {
		filter += " AND job_name = ?"
		args = append(args, jobName)
	}
	if status != "" {
		filter += " AND status = ?"
		args = append(args, strings.ToUpper(status))
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, parent_run_id, retry_count, job_name, trigger_source, status, started_at, finished_at, result_summary, error_message, operator_id
FROM scheduler_job_runs` + filter + `
ORDER BY started_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.SchedulerJobRun, 0)
	for rows.Next() {
		var item model.SchedulerJobRun
		var finishedAt sql.NullTime
		var resultSummary, errorMsg, operatorID, parentRunID sql.NullString
		var startedAt time.Time
		if err := rows.Scan(&item.ID, &parentRunID, &item.RetryCount, &item.JobName, &item.TriggerSource, &item.Status, &startedAt, &finishedAt, &resultSummary, &errorMsg, &operatorID); err != nil {
			return nil, 0, err
		}
		item.StartedAt = startedAt.Format(time.RFC3339)
		if parentRunID.Valid {
			item.ParentRunID = parentRunID.String
		}
		if finishedAt.Valid {
			item.FinishedAt = finishedAt.Time.Format(time.RFC3339)
		}
		if resultSummary.Valid {
			item.ResultSummary = resultSummary.String
		}
		if errorMsg.Valid {
			item.ErrorMessage = errorMsg.String
		}
		if operatorID.Valid {
			item.OperatorID = operatorID.String
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminListNewsSyncRunDetails(runID string, syncType string, source string, symbol string, status string, page int, pageSize int) ([]model.NewsSyncRunDetail, int, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return []model.NewsSyncRunDetail{}, 0, nil
	}
	offset := (page - 1) * pageSize
	args := []interface{}{runID}
	filter := " WHERE run_id = ?"
	if syncType = strings.ToUpper(strings.TrimSpace(syncType)); syncType != "" {
		filter += " AND sync_type = ?"
		args = append(args, syncType)
	}
	if source = strings.TrimSpace(source); source != "" {
		filter += " AND source LIKE ?"
		args = append(args, "%"+source+"%")
	}
	if symbol = strings.ToUpper(strings.TrimSpace(symbol)); symbol != "" {
		filter += " AND symbol = ?"
		args = append(args, symbol)
	}
	if status = strings.ToUpper(strings.TrimSpace(status)); status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM news_sync_run_details"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT
	id, run_id, job_name, sync_type, source, symbol, status,
	fetched_count, upserted_count, failed_count,
	warning_text, error_text, started_at, finished_at, created_at, updated_at
FROM news_sync_run_details` + filter + `
ORDER BY created_at DESC, sync_type ASC, source ASC, symbol ASC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.NewsSyncRunDetail, 0, pageSize)
	for rows.Next() {
		var item model.NewsSyncRunDetail
		var sourceVal, symbolVal, warningVal, errorVal sql.NullString
		var finishedAt sql.NullTime
		var startedAt, createdAt, updatedAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.JobName,
			&item.SyncType,
			&sourceVal,
			&symbolVal,
			&item.Status,
			&item.FetchedCount,
			&item.UpsertedCount,
			&item.FailedCount,
			&warningVal,
			&errorVal,
			&startedAt,
			&finishedAt,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, err
		}
		if sourceVal.Valid {
			item.Source = sourceVal.String
		}
		if symbolVal.Valid {
			item.Symbol = symbolVal.String
		}
		if warningVal.Valid {
			item.WarningText = warningVal.String
		}
		if errorVal.Valid {
			item.ErrorText = errorVal.String
		}
		item.StartedAt = startedAt.Format(time.RFC3339)
		if finishedAt.Valid {
			item.FinishedAt = finishedAt.Time.Format(time.RFC3339)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminCreateNewsSyncRunDetails(runID string, details []model.NewsSyncRunDetail) error {
	runID = strings.TrimSpace(runID)
	if runID == "" || len(details) == 0 {
		return nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
INSERT INTO news_sync_run_details
	(id, run_id, job_name, sync_type, source, symbol, status, fetched_count, upserted_count, failed_count, warning_text, error_text, started_at, finished_at, created_at, updated_at)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	job_name = VALUES(job_name),
	sync_type = VALUES(sync_type),
	source = VALUES(source),
	symbol = VALUES(symbol),
	status = VALUES(status),
	fetched_count = VALUES(fetched_count),
	upserted_count = VALUES(upserted_count),
	failed_count = VALUES(failed_count),
	warning_text = VALUES(warning_text),
	error_text = VALUES(error_text),
	started_at = VALUES(started_at),
	finished_at = VALUES(finished_at),
	updated_at = VALUES(updated_at)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, detail := range details {
		id := strings.TrimSpace(detail.ID)
		if id == "" {
			id = newID("nsd")
		}
		jobName := strings.TrimSpace(detail.JobName)
		if jobName == "" {
			jobName = tushareNewsIncrementalJobName
		}
		syncType := normalizeTushareSyncType(detail.SyncType)
		if syncType == "" {
			syncType = strings.ToUpper(strings.TrimSpace(detail.SyncType))
		}
		if syncType == "" {
			syncType = tushareSyncTypeNewsBrief
		}
		status := strings.ToUpper(strings.TrimSpace(detail.Status))
		if status == "" {
			if detail.FailedCount > 0 || strings.TrimSpace(detail.ErrorText) != "" {
				status = "FAILED"
			} else {
				status = "SUCCESS"
			}
		}
		startedAt := parseDateTimeOrDefault(detail.StartedAt, time.Now())
		finishedAt := parseDateTimeOrDefault(detail.FinishedAt, startedAt)
		if strings.TrimSpace(detail.FinishedAt) == "" {
			finishedAt = startedAt
		}
		warningText := truncateByRunes(normalizeUTF8Text(detail.WarningText), 512)
		errorText := truncateByRunes(normalizeUTF8Text(detail.ErrorText), 512)
		source := truncateByRunes(normalizeUTF8Text(detail.Source), 128)
		symbol := truncateByRunes(normalizeTushareSymbol(detail.Symbol), 32)
		now := time.Now()
		_, err := stmt.Exec(
			id,
			runID,
			jobName,
			syncType,
			nullableString(source),
			nullableString(symbol),
			status,
			maxInt(0, detail.FetchedCount),
			maxInt(0, detail.UpsertedCount),
			maxInt(0, detail.FailedCount),
			nullableString(warningText),
			nullableString(errorText),
			startedAt,
			finishedAt,
			now,
			now,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MySQLGrowthRepo) AdminCreateSchedulerJobRun(jobName string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	id := newID("jr")
	now := time.Now()
	jobName = truncateByRunes(normalizeUTF8Text(jobName), 64)
	operatorID = truncateByRunes(normalizeUTF8Text(operatorID), 32)
	safeResultSummary := truncateByRunes(normalizeUTF8Text(resultSummary), 512)
	safeErrorMessage := truncateByRunes(normalizeUTF8Text(errorMessage), 512)
	var finishedAt interface{} = nil
	if strings.ToUpper(status) != "RUNNING" {
		finishedAt = now
	}
	_, err := r.db.Exec(`
INSERT INTO scheduler_job_runs (id, parent_run_id, retry_count, job_name, trigger_source, status, started_at, finished_at, result_summary, error_message, operator_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, nil, 0, jobName, strings.ToUpper(triggerSource), strings.ToUpper(status), now, finishedAt, safeResultSummary, safeErrorMessage, operatorID, now,
	)
	if err != nil {
		return "", err
	}
	_, _ = r.db.Exec(`
UPDATE scheduler_job_definitions
SET last_run_at = ?, updated_at = ?
WHERE job_name = ?`, now, now, strings.TrimSpace(jobName))
	_ = r.AdminCreateWorkflowMessage(
		"",
		id,
		"SYSTEM",
		operatorID,
		operatorID,
		"JOB_TRIGGERED",
		"任务已触发",
		"任务 "+jobName+" 已触发，状态 "+strings.ToUpper(status),
	)
	if strings.ToUpper(status) == "FAILED" {
		_ = r.AdminCreateWorkflowMessage(
			"",
			id,
			"SYSTEM",
			operatorID,
			operatorID,
			"JOB_FAILED",
			"任务执行失败",
			"任务 "+jobName+" 执行失败："+safeErrorMessage,
		)
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminRetrySchedulerJobRun(runID string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	var jobName string
	var prevRetry int
	err := r.db.QueryRow("SELECT job_name, retry_count FROM scheduler_job_runs WHERE id = ?", runID).Scan(&jobName, &prevRetry)
	if err != nil {
		return "", err
	}
	id := newID("jr")
	now := time.Now()
	operatorID = truncateByRunes(normalizeUTF8Text(operatorID), 32)
	safeResultSummary := truncateByRunes(normalizeUTF8Text(resultSummary), 512)
	safeErrorMessage := truncateByRunes(normalizeUTF8Text(errorMessage), 512)
	upperStatus := strings.ToUpper(status)
	var finishedAt interface{} = nil
	if upperStatus != "RUNNING" {
		finishedAt = now
	}
	upperTriggerSource := strings.ToUpper(strings.TrimSpace(triggerSource))
	if upperTriggerSource == "" {
		upperTriggerSource = "MANUAL"
	}
	_, err = r.db.Exec(`
INSERT INTO scheduler_job_runs (id, parent_run_id, retry_count, job_name, trigger_source, status, started_at, finished_at, result_summary, error_message, operator_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, runID, prevRetry+1, jobName, upperTriggerSource, upperStatus, now, finishedAt, safeResultSummary, safeErrorMessage, operatorID, now,
	)
	if err != nil {
		return "", err
	}
	_, _ = r.db.Exec(`
UPDATE scheduler_job_definitions
SET last_run_at = ?, updated_at = ?
WHERE job_name = ?`, now, now, strings.TrimSpace(jobName))
	_ = r.AdminCreateWorkflowMessage(
		"",
		id,
		"SYSTEM",
		operatorID,
		operatorID,
		"JOB_RETRIED",
		"任务重跑已触发",
		"任务 "+jobName+" 基于运行 "+runID+" 重跑，状态 "+upperStatus,
	)
	if upperStatus == "FAILED" {
		_ = r.AdminCreateWorkflowMessage(
			"",
			id,
			"SYSTEM",
			operatorID,
			operatorID,
			"JOB_FAILED",
			"任务重跑失败",
			"任务 "+jobName+" 重跑失败："+safeErrorMessage,
		)
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateSchedulerJobRun(runID string, status string, resultSummary string, errorMessage string) error {
	now := time.Now()
	safeResultSummary := truncateByRunes(normalizeUTF8Text(resultSummary), 512)
	safeErrorMessage := truncateByRunes(normalizeUTF8Text(errorMessage), 512)
	upperStatus := strings.ToUpper(status)
	var finishedAt interface{} = nil
	if upperStatus != "RUNNING" {
		finishedAt = now
	}

	var jobName string
	var operatorID string
	err := r.db.QueryRow("SELECT job_name, operator_id FROM scheduler_job_runs WHERE id = ?", runID).Scan(&jobName, &operatorID)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
UPDATE scheduler_job_runs
SET status = ?, finished_at = ?, result_summary = ?, error_message = ?
WHERE id = ?`, upperStatus, finishedAt, safeResultSummary, safeErrorMessage, runID)
	if err != nil {
		return err
	}

	if upperStatus == "SUCCESS" {
		_ = r.AdminCreateWorkflowMessage(
			"",
			runID,
			"SYSTEM",
			operatorID,
			operatorID,
			"JOB_COMPLETED",
			"任务执行成功",
			"任务 "+jobName+" 执行成功："+safeResultSummary,
		)
	} else if upperStatus == "FAILED" {
		_ = r.AdminCreateWorkflowMessage(
			"",
			runID,
			"SYSTEM",
			operatorID,
			operatorID,
			"JOB_FAILED",
			"任务执行失败",
			"任务 "+jobName+" 执行失败："+safeErrorMessage,
		)
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminListSchedulerJobDefinitions(status string, module string, page int, pageSize int) ([]model.SchedulerJobDefinition, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND status = ?"
		args = append(args, strings.ToUpper(status))
	}
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_definitions"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at
FROM scheduler_job_definitions` + filter + `
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.SchedulerJobDefinition, 0)
	for rows.Next() {
		var item model.SchedulerJobDefinition
		var lastRunAt sql.NullTime
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.JobName, &item.DisplayName, &item.Module, &item.CronExpr, &item.Status, &lastRunAt, &item.UpdatedBy, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		if lastRunAt.Valid {
			item.LastRunAt = lastRunAt.Time.Format(time.RFC3339)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateSchedulerJobDefinition(item model.SchedulerJobDefinition, operatorID string) (string, error) {
	id := newID("jobdef")
	now := time.Now()
	_, err := r.db.Exec(`
INSERT INTO scheduler_job_definitions (id, job_name, display_name, module, cron_expr, status, updated_by, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, item.JobName, item.DisplayName, strings.ToUpper(item.Module), item.CronExpr, strings.ToUpper(item.Status), operatorID, now, now,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateSchedulerJobDefinition(id string, item model.SchedulerJobDefinition, operatorID string) error {
	_, err := r.db.Exec(`
UPDATE scheduler_job_definitions
SET display_name = ?, module = ?, cron_expr = ?, status = ?, updated_by = ?, updated_at = ?
WHERE id = ?`,
		item.DisplayName, strings.ToUpper(item.Module), item.CronExpr, strings.ToUpper(item.Status), operatorID, time.Now(), id,
	)
	return err
}

func (r *MySQLGrowthRepo) AdminUpdateSchedulerJobDefinitionStatus(id string, status string, operatorID string) error {
	_, err := r.db.Exec(
		"UPDATE scheduler_job_definitions SET status = ?, updated_by = ?, updated_at = ? WHERE id = ?",
		strings.ToUpper(status), operatorID, time.Now(), id,
	)
	return err
}

func (r *MySQLGrowthRepo) AdminDeleteSchedulerJobDefinition(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return sql.ErrNoRows
	}
	result, err := r.db.Exec("DELETE FROM scheduler_job_definitions WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminListWorkflowMessages(module string, eventType string, isRead string, receiverID string, page int, pageSize int) ([]model.WorkflowMessage, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if eventType != "" {
		filter += " AND event_type = ?"
		args = append(args, strings.ToUpper(eventType))
	}
	if receiverID != "" {
		filter += " AND receiver_id = ?"
		args = append(args, receiverID)
	}
	if isRead != "" {
		switch strings.ToLower(isRead) {
		case "true", "1":
			filter += " AND is_read = 1"
		case "false", "0":
			filter += " AND is_read = 0"
		}
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM workflow_messages"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, review_id, target_id, module, receiver_id, sender_id, event_type, title, content, is_read, created_at, read_at
FROM workflow_messages` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.WorkflowMessage, 0)
	for rows.Next() {
		var item model.WorkflowMessage
		var reviewID, receiverID, senderID sql.NullString
		var readAt sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &reviewID, &item.TargetID, &item.Module, &receiverID, &senderID, &item.EventType, &item.Title, &item.Content, &item.IsRead, &createdAt, &readAt); err != nil {
			return nil, 0, err
		}
		if reviewID.Valid {
			item.ReviewID = reviewID.String
		}
		if receiverID.Valid {
			item.ReceiverID = receiverID.String
		}
		if senderID.Valid {
			item.SenderID = senderID.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		if readAt.Valid {
			item.ReadAt = readAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCountUnreadWorkflowMessages(module string, eventType string, receiverID string) (int, error) {
	args := []interface{}{}
	filter := " WHERE is_read = 0"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if eventType != "" {
		filter += " AND event_type = ?"
		args = append(args, strings.ToUpper(eventType))
	}
	if receiverID != "" {
		filter += " AND receiver_id = ?"
		args = append(args, receiverID)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM workflow_messages"+filter, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *MySQLGrowthRepo) AdminUpdateWorkflowMessageRead(id string, isRead bool) error {
	readAt := interface{}(nil)
	if isRead {
		readAt = time.Now()
	}
	_, err := r.db.Exec("UPDATE workflow_messages SET is_read = ?, read_at = ? WHERE id = ?", isRead, readAt, id)
	return err
}

func (r *MySQLGrowthRepo) AdminBulkReadWorkflowMessages(module string, eventType string, receiverID string) (int64, error) {
	args := []interface{}{}
	filter := " WHERE is_read = 0"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if eventType != "" {
		filter += " AND event_type = ?"
		args = append(args, strings.ToUpper(eventType))
	}
	if receiverID != "" {
		filter += " AND receiver_id = ?"
		args = append(args, receiverID)
	}
	query := "UPDATE workflow_messages SET is_read = 1, read_at = ?" + filter
	args = append([]interface{}{time.Now()}, args...)
	res, err := r.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) AdminCreateWorkflowMessage(reviewID string, targetID string, module string, receiverID string, senderID string, eventType string, title string, content string) error {
	reviewID = truncateByRunes(normalizeUTF8Text(reviewID), 64)
	targetID = truncateByRunes(normalizeUTF8Text(targetID), 64)
	module = truncateByRunes(strings.ToUpper(normalizeUTF8Text(module)), 32)
	receiverID = truncateByRunes(normalizeUTF8Text(receiverID), 32)
	senderID = truncateByRunes(normalizeUTF8Text(senderID), 32)
	eventType = truncateByRunes(strings.ToUpper(normalizeUTF8Text(eventType)), 32)
	title = truncateByRunes(normalizeUTF8Text(title), 128)
	content = truncateByRunes(normalizeUTF8Text(content), 512)

	var reviewVal interface{} = nil
	if reviewID != "" {
		reviewVal = reviewID
	}
	var receiverVal interface{} = nil
	if receiverID != "" {
		receiverVal = receiverID
	}
	var senderVal interface{} = nil
	if senderID != "" {
		senderVal = senderID
	}
	_, err := r.db.Exec(`
INSERT INTO workflow_messages (id, review_id, target_id, module, receiver_id, sender_id, event_type, title, content, is_read, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?)`,
		newID("wm"), reviewVal, targetID, module, receiverVal, senderVal, eventType, title, content, time.Now(),
	)
	return err
}

func (r *MySQLGrowthRepo) AdminGetWorkflowMetrics(module string, receiverID string) (model.WorkflowMetrics, error) {
	result := model.WorkflowMetrics{}
	args := []interface{}{}
	filter := " WHERE 1=1"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM review_tasks"+filter+" AND status = 'PENDING'", args...).Scan(&result.PendingReviews); err != nil {
		return result, err
	}

	todayStart := time.Now().Truncate(24 * time.Hour)
	tomorrowStart := todayStart.Add(24 * time.Hour)
	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM review_tasks"+filter+" AND status = 'APPROVED' AND reviewed_at >= ? AND reviewed_at < ?",
		append(args, todayStart, tomorrowStart)...,
	).Scan(&result.ApprovedToday); err != nil {
		return result, err
	}
	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM review_tasks"+filter+" AND status = 'REJECTED' AND reviewed_at >= ? AND reviewed_at < ?",
		append(args, todayStart, tomorrowStart)...,
	).Scan(&result.RejectedToday); err != nil {
		return result, err
	}

	msgArgs := []interface{}{}
	msgFilter := " WHERE 1=1"
	if module != "" {
		msgFilter += " AND module = ?"
		msgArgs = append(msgArgs, strings.ToUpper(module))
	}
	if receiverID != "" {
		msgFilter += " AND receiver_id = ?"
		msgArgs = append(msgArgs, receiverID)
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM workflow_messages"+msgFilter, msgArgs...).Scan(&result.TotalMessages); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM workflow_messages"+msgFilter+" AND is_read = 0", msgArgs...).Scan(&result.UnreadMessages); err != nil {
		return result, err
	}
	return result, nil
}

func (r *MySQLGrowthRepo) AdminGetSchedulerJobMetrics(jobName string) (model.SchedulerJobMetrics, error) {
	result := model.SchedulerJobMetrics{}
	normalizedJobName := strings.TrimSpace(jobName)
	args := []interface{}{}
	filter := " WHERE started_at >= ? AND started_at < ?"
	todayStart := time.Now().Truncate(24 * time.Hour)
	tomorrowStart := todayStart.Add(24 * time.Hour)
	args = append(args, todayStart, tomorrowStart)
	if normalizedJobName != "" {
		filter += " AND job_name = ?"
		args = append(args, normalizedJobName)
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter, args...).Scan(&result.TodayTotal); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND status = 'SUCCESS'", args...).Scan(&result.TodaySuccess); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND status = 'FAILED'", args...).Scan(&result.TodayFailed); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND status = 'RUNNING'", args...).Scan(&result.TodayRunning); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND retry_count > 0", args...).Scan(&result.RetryTotal); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND retry_count > 0 AND status = 'SUCCESS'", args...).Scan(&result.RetrySuccess); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND retry_count > 0 AND status = 'FAILED'", args...).Scan(&result.RetryFailed); err != nil {
		return result, err
	}
	if result.RetryTotal > 0 {
		result.RetryHitRate = roundTo(float64(result.RetrySuccess)/float64(result.RetryTotal), 4)
	} else {
		result.RetryHitRate = 0
	}
	var avgRetry sql.NullFloat64
	if err := r.db.QueryRow("SELECT AVG(retry_count) FROM scheduler_job_runs"+filter+" AND retry_count > 0", args...).Scan(&avgRetry); err != nil {
		return result, err
	}
	if avgRetry.Valid {
		result.AvgRetryCount = roundTo(avgRetry.Float64, 2)
	}

	jobStatsRows, err := r.db.Query(`
SELECT
	job_name,
	COUNT(*) AS today_total,
	SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS today_success,
	SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END) AS today_failed,
	SUM(CASE WHEN status = 'RUNNING' THEN 1 ELSE 0 END) AS today_running,
	SUM(CASE WHEN retry_count > 0 THEN 1 ELSE 0 END) AS retry_total,
	SUM(CASE WHEN retry_count > 0 AND status = 'SUCCESS' THEN 1 ELSE 0 END) AS retry_success,
	SUM(CASE WHEN retry_count > 0 AND status = 'FAILED' THEN 1 ELSE 0 END) AS retry_failed,
	SUM(CASE WHEN retry_count > 0 AND trigger_source = 'SYSTEM' THEN 1 ELSE 0 END) AS auto_retry_total,
	AVG(CASE WHEN retry_count > 0 THEN retry_count ELSE NULL END) AS avg_retry_count
FROM scheduler_job_runs`+filter+`
GROUP BY job_name
ORDER BY retry_total DESC, job_name ASC`, args...)
	if err != nil {
		return result, err
	}
	defer jobStatsRows.Close()
	jobStatsMap := make(map[string]*model.SchedulerJobRetryStat)
	for jobStatsRows.Next() {
		var (
			item                model.SchedulerJobRetryStat
			todayTotal          int64
			todaySuccess        int64
			todayFailed         int64
			todayRunning        int64
			retryTotal          int64
			retrySuccess        int64
			retryFailed         int64
			autoRetryTotal      int64
			avgRetryCountPerJob sql.NullFloat64
		)
		if err := jobStatsRows.Scan(
			&item.JobName,
			&todayTotal,
			&todaySuccess,
			&todayFailed,
			&todayRunning,
			&retryTotal,
			&retrySuccess,
			&retryFailed,
			&autoRetryTotal,
			&avgRetryCountPerJob,
		); err != nil {
			return result, err
		}
		item.TodayTotal = int(todayTotal)
		item.TodaySuccess = int(todaySuccess)
		item.TodayFailed = int(todayFailed)
		item.TodayRunning = int(todayRunning)
		item.RetryTotal = int(retryTotal)
		item.RetrySuccess = int(retrySuccess)
		item.RetryFailed = int(retryFailed)
		item.AutoRetryTotal = int(autoRetryTotal)
		if item.RetryTotal > 0 {
			item.RetryHitRate = roundTo(float64(item.RetrySuccess)/float64(item.RetryTotal), 4)
		}
		if avgRetryCountPerJob.Valid {
			item.AvgRetryCount = roundTo(avgRetryCountPerJob.Float64, 2)
		}
		jobStatsMap[item.JobName] = &item
	}
	if err := jobStatsRows.Err(); err != nil {
		return result, err
	}

	if err := r.fillSchedulerRecoveryStats(jobStatsMap, todayStart, tomorrowStart, normalizedJobName); err != nil {
		return result, err
	}

	jobStats := make([]model.SchedulerJobRetryStat, 0, len(jobStatsMap))
	for _, item := range jobStatsMap {
		jobStats = append(jobStats, *item)
	}
	sort.Slice(jobStats, func(i, j int) bool {
		if jobStats[i].RetryTotal == jobStats[j].RetryTotal {
			if jobStats[i].RecoveryTotal == jobStats[j].RecoveryTotal {
				return jobStats[i].JobName < jobStats[j].JobName
			}
			return jobStats[i].RecoveryTotal > jobStats[j].RecoveryTotal
		}
		return jobStats[i].RetryTotal > jobStats[j].RetryTotal
	})
	result.JobRetryStats = jobStats
	for _, item := range jobStats {
		result.AutoRetryTotal += item.AutoRetryTotal
		result.RecoveryTotal += item.RecoveryTotal
		result.RecoverySuccess += item.RecoverySuccess
	}
	if result.RecoveryTotal > 0 {
		result.RecoveryHitRate = roundTo(float64(result.RecoverySuccess)/float64(result.RecoveryTotal), 4)
	}

	failScopeDays := 7
	failSince := time.Now().AddDate(0, 0, -failScopeDays)
	failArgs := []interface{}{failSince}
	failFilter := " WHERE status = 'FAILED' AND started_at >= ?"
	if normalizedJobName != "" {
		failFilter += " AND job_name = ?"
		failArgs = append(failArgs, normalizedJobName)
	}
	failRows, err := r.db.Query(`
SELECT job_name, error_message, started_at
FROM scheduler_job_runs`+failFilter+`
ORDER BY started_at DESC
LIMIT 800`, failArgs...)
	if err != nil {
		return result, err
	}
	defer failRows.Close()
	reasonAggMap := make(map[string]*schedulerFailureReasonAggregate)
	jobReasonAggMap := make(map[string]*schedulerFailureReasonAggregate)
	for failRows.Next() {
		var (
			jobNameValue string
			errorMessage sql.NullString
			startedAt    time.Time
		)
		if err := failRows.Scan(&jobNameValue, &errorMessage, &startedAt); err != nil {
			return result, err
		}
		reason := normalizeSchedulerFailureReason(errorMessage.String)
		agg := reasonAggMap[reason]
		if agg == nil {
			agg = &schedulerFailureReasonAggregate{Reason: reason}
			reasonAggMap[reason] = agg
		}
		agg.Count++
		if agg.LastOccurredAt.IsZero() || startedAt.After(agg.LastOccurredAt) {
			agg.LastOccurredAt = startedAt
		}
		jobReasonKey := strings.TrimSpace(jobNameValue) + "|" + reason
		jobAgg := jobReasonAggMap[jobReasonKey]
		if jobAgg == nil {
			jobAgg = &schedulerFailureReasonAggregate{
				JobName: strings.TrimSpace(jobNameValue),
				Reason:  reason,
			}
			jobReasonAggMap[jobReasonKey] = jobAgg
		}
		jobAgg.Count++
		if jobAgg.LastOccurredAt.IsZero() || startedAt.After(jobAgg.LastOccurredAt) {
			jobAgg.LastOccurredAt = startedAt
		}
	}
	if err := failRows.Err(); err != nil {
		return result, err
	}
	reasons := make([]model.SchedulerJobFailureReason, 0, len(reasonAggMap))
	for _, agg := range reasonAggMap {
		reasons = append(reasons, model.SchedulerJobFailureReason{
			Reason:         agg.Reason,
			Count:          agg.Count,
			LastOccurredAt: agg.LastOccurredAt.Format(time.RFC3339),
		})
	}
	sort.Slice(reasons, func(i, j int) bool {
		if reasons[i].Count == reasons[j].Count {
			return reasons[i].Reason < reasons[j].Reason
		}
		return reasons[i].Count > reasons[j].Count
	})
	if len(reasons) > 8 {
		reasons = reasons[:8]
	}
	jobReasons := make([]model.SchedulerJobFailureByJob, 0, len(jobReasonAggMap))
	for _, agg := range jobReasonAggMap {
		jobReasons = append(jobReasons, model.SchedulerJobFailureByJob{
			JobName:        agg.JobName,
			Reason:         agg.Reason,
			Count:          agg.Count,
			LastOccurredAt: agg.LastOccurredAt.Format(time.RFC3339),
		})
	}
	sort.Slice(jobReasons, func(i, j int) bool {
		if jobReasons[i].Count == jobReasons[j].Count {
			if jobReasons[i].JobName == jobReasons[j].JobName {
				return jobReasons[i].Reason < jobReasons[j].Reason
			}
			return jobReasons[i].JobName < jobReasons[j].JobName
		}
		return jobReasons[i].Count > jobReasons[j].Count
	})
	if len(jobReasons) > 36 {
		jobReasons = jobReasons[:36]
	}
	result.FailureReasons = reasons
	result.JobFailureReasons = jobReasons
	result.FailureReasonScope = fmt.Sprintf("LAST_%d_DAYS", failScopeDays)
	return result, nil
}

type schedulerJobRecoveryState struct {
	JobName       string
	HasRetry      bool
	MaxRetryCount int
	FinalStatus   string
}

type schedulerJobRetryChild struct {
	ID          string
	ParentRunID string
	JobName     string
	Status      string
	RetryCount  int
}

func (r *MySQLGrowthRepo) fillSchedulerRecoveryStats(
	jobStatsMap map[string]*model.SchedulerJobRetryStat,
	startedAt time.Time,
	endedAt time.Time,
	jobName string,
) error {
	args := []interface{}{startedAt, endedAt}
	filter := ""
	if jobName != "" {
		filter += " AND job_name = ?"
		args = append(args, jobName)
	}
	rootRows, err := r.db.Query(`
SELECT id, job_name
FROM scheduler_job_runs
WHERE retry_count = 0
	AND status = 'FAILED'
	AND started_at >= ?
	AND started_at < ?`+filter, args...)
	if err != nil {
		return err
	}
	defer rootRows.Close()

	recoveryStates := make(map[string]*schedulerJobRecoveryState)
	runToRoot := make(map[string]string)
	frontier := make([]string, 0)
	visited := make(map[string]struct{})
	for rootRows.Next() {
		var (
			rootID      string
			rootJobName string
		)
		if err := rootRows.Scan(&rootID, &rootJobName); err != nil {
			return err
		}
		rootID = strings.TrimSpace(rootID)
		if rootID == "" {
			continue
		}
		rootJobName = strings.TrimSpace(rootJobName)
		recoveryStates[rootID] = &schedulerJobRecoveryState{
			JobName:       rootJobName,
			HasRetry:      false,
			MaxRetryCount: 0,
			FinalStatus:   "FAILED",
		}
		runToRoot[rootID] = rootID
		frontier = append(frontier, rootID)
		visited[rootID] = struct{}{}
	}
	if err := rootRows.Err(); err != nil {
		return err
	}
	if len(frontier) == 0 {
		return nil
	}

	const batchSize = 200
	for len(frontier) > 0 {
		nextFrontier := make([]string, 0)
		for i := 0; i < len(frontier); i += batchSize {
			end := i + batchSize
			if end > len(frontier) {
				end = len(frontier)
			}
			children, err := r.loadSchedulerJobRetryChildren(frontier[i:end])
			if err != nil {
				return err
			}
			for _, child := range children {
				if _, seen := visited[child.ID]; seen {
					continue
				}
				rootID, ok := runToRoot[child.ParentRunID]
				if !ok || rootID == "" {
					continue
				}
				visited[child.ID] = struct{}{}
				runToRoot[child.ID] = rootID
				nextFrontier = append(nextFrontier, child.ID)

				state := recoveryStates[rootID]
				if state == nil {
					continue
				}
				if state.JobName == "" {
					state.JobName = strings.TrimSpace(child.JobName)
				}
				if child.RetryCount > 0 {
					state.HasRetry = true
				}
				childStatus := strings.ToUpper(strings.TrimSpace(child.Status))
				if child.RetryCount > state.MaxRetryCount {
					state.MaxRetryCount = child.RetryCount
					state.FinalStatus = childStatus
					continue
				}
				if child.RetryCount == state.MaxRetryCount && state.MaxRetryCount > 0 {
					if state.FinalStatus != "SUCCESS" && childStatus == "SUCCESS" {
						state.FinalStatus = childStatus
					}
				}
			}
		}
		frontier = nextFrontier
	}

	for _, state := range recoveryStates {
		if !state.HasRetry {
			continue
		}
		jobNameValue := strings.TrimSpace(state.JobName)
		if jobNameValue == "" {
			jobNameValue = "UNKNOWN"
		}
		item := jobStatsMap[jobNameValue]
		if item == nil {
			item = &model.SchedulerJobRetryStat{JobName: jobNameValue}
			jobStatsMap[jobNameValue] = item
		}
		item.RecoveryTotal++
		if strings.EqualFold(state.FinalStatus, "SUCCESS") {
			item.RecoverySuccess++
		}
		if item.RecoveryTotal > 0 {
			item.RecoveryHitRate = roundTo(float64(item.RecoverySuccess)/float64(item.RecoveryTotal), 4)
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) loadSchedulerJobRetryChildren(parentRunIDs []string) ([]schedulerJobRetryChild, error) {
	if len(parentRunIDs) == 0 {
		return []schedulerJobRetryChild{}, nil
	}
	placeholder := strings.TrimSuffix(strings.Repeat("?,", len(parentRunIDs)), ",")
	query := fmt.Sprintf(`
SELECT id, parent_run_id, job_name, status, retry_count
FROM scheduler_job_runs
WHERE parent_run_id IN (%s)
ORDER BY retry_count ASC, id ASC`, placeholder)
	args := make([]interface{}, 0, len(parentRunIDs))
	for _, id := range parentRunIDs {
		args = append(args, id)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]schedulerJobRetryChild, 0)
	for rows.Next() {
		var item schedulerJobRetryChild
		if err := rows.Scan(&item.ID, &item.ParentRunID, &item.JobName, &item.Status, &item.RetryCount); err != nil {
			return nil, err
		}
		item.ID = strings.TrimSpace(item.ID)
		item.ParentRunID = strings.TrimSpace(item.ParentRunID)
		item.JobName = strings.TrimSpace(item.JobName)
		item.Status = strings.TrimSpace(item.Status)
		if item.ID == "" || item.ParentRunID == "" {
			continue
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type schedulerFailureReasonAggregate struct {
	JobName        string
	Reason         string
	Count          int
	LastOccurredAt time.Time
}

func normalizeSchedulerFailureReason(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "UNKNOWN_ERROR"
	}
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "tushare") && strings.Contains(lower, "token"):
		return "TUSHARE_TOKEN_INVALID_OR_MISSING"
	case strings.Contains(lower, "error 2003"), strings.Contains(lower, "can’t connect"), strings.Contains(lower, "can't connect"):
		return "MYSQL_CONNECTION_FAILED"
	case strings.Contains(lower, "error 1146"):
		return "TABLE_NOT_FOUND"
	case strings.Contains(lower, "timeout"):
		return "UPSTREAM_TIMEOUT"
	case strings.Contains(lower, "status: 5"), strings.Contains(lower, "status code: 5"), strings.Contains(lower, "internal server error"):
		return "UPSTREAM_5XX"
	case strings.Contains(lower, "permission"), strings.Contains(lower, "forbidden"), strings.Contains(lower, "unauthorized"):
		return "PERMISSION_DENIED"
	case strings.Contains(lower, "unknown job"):
		return "UNKNOWN_JOB_NAME"
	default:
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.Join(strings.Fields(text), " ")
		if len(text) > 96 {
			text = text[:96]
		}
		return text
	}
}

type sqlExecer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func (r *MySQLGrowthRepo) applyModuleTargetStatus(execer sqlExecer, module string, targetID string, status string) error {
	var (
		query string
		args  []interface{}
	)
	switch strings.ToUpper(module) {
	case "NEWS":
		query = "UPDATE news_articles SET status = ?, updated_at = ? WHERE id = ?"
		args = []interface{}{status, time.Now(), targetID}
	case "STOCK":
		query = "UPDATE stock_recommendations SET status = ? WHERE id = ?"
		args = []interface{}{status, targetID}
	case "FUTURES":
		query = "UPDATE futures_strategies SET status = ? WHERE id = ?"
		args = []interface{}{status, targetID}
	case "STOCK_EVENT":
		query = `UPDATE stock_event_clusters
SET cluster_status = ?, review_status = ?, updated_at = ?
WHERE id = ?`
		clusterStatus := "CLUSTERED"
		reviewStatus := "PENDING"
		switch strings.ToUpper(strings.TrimSpace(status)) {
		case "PUBLISHED":
			clusterStatus = "REVIEWED"
			reviewStatus = "APPROVED"
		case "DRAFT":
			clusterStatus = "REJECTED"
			reviewStatus = "REJECTED"
		}
		args = []interface{}{clusterStatus, reviewStatus, time.Now(), targetID}
	default:
		return errors.New("unsupported module")
	}
	res, err := execer.Exec(query, args...)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("target not found")
	}
	return nil
}
