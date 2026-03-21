package repo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) upsertStrategyEngineJobSnapshot(record model.StrategyEngineJobRecord) error {
	if strings.TrimSpace(record.JobID) == "" {
		return sql.ErrNoRows
	}

	payloadJSON, err := marshalStrategySnapshotJSON(record.Payload)
	if err != nil {
		return err
	}
	configRefsJSON, err := marshalStrategySnapshotJSON(snapshotMap(record.Payload, "config_refs"))
	if err != nil {
		return err
	}
	publishPreviewJSON, err := marshalStrategySnapshotJSON(snapshotMap(record.Payload, "publish_policy_preview"))
	if err != nil {
		return err
	}

	tradeDate := firstNonEmptyString(
		strategySnapshotTradeDate(record.Payload),
		strategySnapshotTradeDate(record.ResultPayloadEcho()),
		strategySnapshotReportTradeDate(record.Result),
	)
	remoteCreatedAt, err := parseStrategySnapshotTimeNullable(record.CreatedAt)
	if err != nil {
		return err
	}
	remoteStartedAt, err := parseStrategySnapshotTimeNullable(record.StartedAt)
	if err != nil {
		return err
	}
	remoteFinishedAt, err := parseStrategySnapshotTimeNullable(record.FinishedAt)
	if err != nil {
		return err
	}
	syncedAt := time.Now().UTC()

	if _, err := r.db.Exec(`
INSERT INTO strategy_job_runs (
  job_id, job_type, status, requested_by, trace_id, trade_date,
  payload_snapshot, config_refs, publish_policy_preview,
  error_message, remote_created_at, remote_started_at, remote_finished_at, synced_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  job_type = VALUES(job_type),
  status = VALUES(status),
  requested_by = VALUES(requested_by),
  trace_id = VALUES(trace_id),
  trade_date = VALUES(trade_date),
  payload_snapshot = VALUES(payload_snapshot),
  config_refs = VALUES(config_refs),
  publish_policy_preview = VALUES(publish_policy_preview),
  error_message = VALUES(error_message),
  remote_created_at = VALUES(remote_created_at),
  remote_started_at = VALUES(remote_started_at),
  remote_finished_at = VALUES(remote_finished_at),
  synced_at = VALUES(synced_at),
  updated_at = VALUES(synced_at)`,
		record.JobID,
		record.JobType,
		record.Status,
		nullableString(record.RequestedBy),
		nullableString(record.TraceID),
		nullableDateString(tradeDate),
		nullableJSONString(payloadJSON),
		nullableJSONString(configRefsJSON),
		nullableJSONString(publishPreviewJSON),
		nullableString(record.ErrorMessage),
		remoteCreatedAt,
		remoteStartedAt,
		remoteFinishedAt,
		syncedAt,
	); err != nil {
		return err
	}

	if record.Result == nil {
		_, err = r.db.Exec(`DELETE FROM strategy_job_artifacts WHERE job_id = ?`, record.JobID)
		return err
	}

	artifactsJSON, err := marshalStrategySnapshotJSON(record.Result.Artifacts)
	if err != nil {
		return err
	}
	reportSnapshot := strategySnapshotReport(record.Result)
	reportJSON, err := marshalStrategySnapshotJSON(reportSnapshot)
	if err != nil {
		return err
	}
	warningsJSON, err := marshalStrategySnapshotJSON(record.Result.Warnings)
	if err != nil {
		return err
	}
	payloadEchoJSON, err := marshalStrategySnapshotJSON(record.Result.PayloadEcho)
	if err != nil {
		return err
	}
	assetKeysJSON, err := marshalStrategySnapshotJSON(strategySnapshotAssetKeys(reportSnapshot))
	if err != nil {
		return err
	}
	selectedCount, payloadCount := strategySnapshotCounts(reportSnapshot)

	_, err = r.db.Exec(`
INSERT INTO strategy_job_artifacts (
  job_id, result_summary, payload_echo_snapshot, warning_messages, artifacts_snapshot, report_snapshot,
  selected_count, payload_count, asset_keys, synced_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  result_summary = VALUES(result_summary),
  payload_echo_snapshot = VALUES(payload_echo_snapshot),
  warning_messages = VALUES(warning_messages),
  artifacts_snapshot = VALUES(artifacts_snapshot),
  report_snapshot = VALUES(report_snapshot),
  selected_count = VALUES(selected_count),
  payload_count = VALUES(payload_count),
  asset_keys = VALUES(asset_keys),
  synced_at = VALUES(synced_at),
  updated_at = VALUES(synced_at)`,
		record.JobID,
		nullableString(record.Result.Summary),
		nullableJSONString(payloadEchoJSON),
		nullableJSONString(warningsJSON),
		nullableJSONString(artifactsJSON),
		nullableJSONString(reportJSON),
		selectedCount,
		payloadCount,
		nullableJSONString(assetKeysJSON),
		syncedAt,
	)
	return err
}

func (r *MySQLGrowthRepo) listStrategyEngineJobSnapshots(jobType string, status string, page int, pageSize int) ([]model.StrategyEngineJobRecord, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	args := make([]any, 0, 4)
	filters := []string{"1=1"}
	if trimmed := strings.TrimSpace(jobType); trimmed != "" {
		filters = append(filters, "r.job_type = ?")
		args = append(args, trimmed)
	}
	if trimmed := strings.TrimSpace(status); trimmed != "" {
		filters = append(filters, "r.status = ?")
		args = append(args, trimmed)
	}
	where := strings.Join(filters, " AND ")

	var total int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM strategy_job_runs r WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []model.StrategyEngineJobRecord{}, 0, nil
	}

	queryArgs := append(append([]any{}, args...), pageSize, offset)
	rows, err := r.db.Query(`
SELECT
  r.job_id,
  r.job_type,
  r.status,
  COALESCE(r.requested_by, ''),
  COALESCE(r.trace_id, ''),
  COALESCE(DATE_FORMAT(r.trade_date, '%Y-%m-%d'), ''),
  COALESCE(CAST(r.payload_snapshot AS CHAR), ''),
  COALESCE(r.error_message, ''),
  COALESCE(DATE_FORMAT(r.remote_created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.remote_started_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.remote_finished_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.synced_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(a.result_summary, ''),
  COALESCE(CAST(a.payload_echo_snapshot AS CHAR), ''),
  COALESCE(CAST(a.warning_messages AS CHAR), ''),
  COALESCE(CAST(a.artifacts_snapshot AS CHAR), '')
FROM strategy_job_runs r
LEFT JOIN strategy_job_artifacts a ON a.job_id = r.job_id
WHERE `+where+`
ORDER BY COALESCE(r.remote_created_at, r.created_at) DESC, r.job_id DESC
LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.StrategyEngineJobRecord, 0, pageSize)
	for rows.Next() {
		item, err := scanStrategyEngineJobSnapshotRow(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, strategySnapshotJobRecordWithSource(item, "LOCAL_ARCHIVED"))
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) getStrategyEngineJobSnapshot(jobID string) (model.StrategyEngineJobRecord, error) {
	row := r.db.QueryRow(`
SELECT
  r.job_id,
  r.job_type,
  r.status,
  COALESCE(r.requested_by, ''),
  COALESCE(r.trace_id, ''),
  COALESCE(DATE_FORMAT(r.trade_date, '%Y-%m-%d'), ''),
  COALESCE(CAST(r.payload_snapshot AS CHAR), ''),
  COALESCE(r.error_message, ''),
  COALESCE(DATE_FORMAT(r.remote_created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.remote_started_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.remote_finished_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.synced_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(a.result_summary, ''),
  COALESCE(CAST(a.payload_echo_snapshot AS CHAR), ''),
  COALESCE(CAST(a.warning_messages AS CHAR), ''),
  COALESCE(CAST(a.artifacts_snapshot AS CHAR), '')
FROM strategy_job_runs r
LEFT JOIN strategy_job_artifacts a ON a.job_id = r.job_id
WHERE r.job_id = ?`, jobID)
	item, err := scanStrategyEngineJobSnapshotRow(row)
	if err != nil {
		return model.StrategyEngineJobRecord{}, err
	}
	return strategySnapshotJobRecordWithSource(item, "LOCAL_ARCHIVED"), nil
}

func (r *MySQLGrowthRepo) createStrategyEngineJobReplay(jobID string, publishRecord model.StrategyEnginePublishRecord, operator string, force bool, overrideReason string, policy *model.StrategyPublishPolicy) error {
	if strings.TrimSpace(jobID) == "" {
		return sql.ErrNoRows
	}
	replay := publishRecord.Replay
	replay.PublishID = publishRecord.PublishID
	replay.JobID = jobID
	replay.PublishVersion = publishRecord.Version
	replay.CreatedAt = firstNonEmpty(replay.CreatedAt, publishRecord.CreatedAt)
	return r.upsertStrategyEngineJobReplaySnapshot(jobID, replay, operator, force, overrideReason, buildStrategyPublishPolicyPreview(policy))
}

func (r *MySQLGrowthRepo) upsertStrategyEngineJobReplaySnapshot(jobID string, replay model.StrategyEnginePublishReplay, operator string, force bool, overrideReason string, policySnapshot map[string]any) error {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return sql.ErrNoRows
	}
	replay.PublishID = strings.TrimSpace(replay.PublishID)
	replay.JobID = firstNonEmpty(strings.TrimSpace(replay.JobID), jobID)

	policyJSON, err := marshalStrategySnapshotJSON(policySnapshot)
	if err != nil {
		return err
	}
	replayJSON, err := marshalStrategySnapshotJSON(map[string]any{
		"warning_count":      replay.WarningCount,
		"warning_messages":   replay.WarningMessages,
		"vetoed_assets":      replay.VetoedAssets,
		"invalidated_assets": replay.InvalidatedAssets,
		"notes":              replay.Notes,
	})
	if err != nil {
		return err
	}

	createdAt := time.Now().UTC()
	if parsedCreatedAt, parseErr := parseStrategySnapshotTimeNullable(replay.CreatedAt); parseErr != nil {
		return parseErr
	} else if parsedCreatedAt != nil {
		if parsed, ok := parsedCreatedAt.(time.Time); ok {
			createdAt = parsed
		}
	}

	_, err = r.db.Exec(`
INSERT INTO strategy_job_replays (
  id, job_id, publish_id, publish_version, operator, force_publish, override_reason,
  policy_snapshot, replay_snapshot, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  publish_id = VALUES(publish_id),
  publish_version = VALUES(publish_version),
  operator = COALESCE(VALUES(operator), operator),
  force_publish = CASE
    WHEN VALUES(operator) IS NULL AND VALUES(override_reason) IS NULL AND VALUES(policy_snapshot) IS NULL
      THEN force_publish
    ELSE VALUES(force_publish)
  END,
  override_reason = COALESCE(VALUES(override_reason), override_reason),
  policy_snapshot = COALESCE(VALUES(policy_snapshot), policy_snapshot),
  replay_snapshot = VALUES(replay_snapshot),
  created_at = VALUES(created_at)`,
		strategyJobReplaySnapshotID(jobID, replay.PublishID, replay.PublishVersion),
		jobID,
		replay.PublishID,
		replay.PublishVersion,
		nullableString(operator),
		force,
		nullableString(overrideReason),
		nullableJSONString(policyJSON),
		nullableJSONString(replayJSON),
		createdAt,
	)
	return err
}

func strategyJobReplaySnapshotID(jobID string, publishID string, publishVersion int) string {
	if strings.TrimSpace(publishID) != "" {
		return "job_replay_" + strings.TrimSpace(publishID)
	}
	if publishVersion > 0 {
		return "job_replay_" + strings.TrimSpace(jobID) + "_" + strconv.Itoa(publishVersion)
	}
	return "job_replay_" + strings.TrimSpace(jobID)
}

func (r *MySQLGrowthRepo) archiveStrategyEnginePublishSnapshot(record model.StrategyEnginePublishRecord, job model.StrategyEngineJobRecord) {
	if err := r.upsertStrategyEngineJobSnapshot(job); err != nil {
		return
	}
	if strings.TrimSpace(record.PublishID) == "" {
		return
	}
	replay := record.Replay
	replay.PublishID = record.PublishID
	replay.JobID = firstNonEmpty(replay.JobID, job.JobID)
	replay.PublishVersion = maxInt(replay.PublishVersion, record.Version)
	replay.CreatedAt = firstNonEmpty(replay.CreatedAt, record.CreatedAt)
	replay.StorageSource = "REMOTE_BACKFILLED"
	_ = r.upsertStrategyEngineJobReplaySnapshot(job.JobID, replay, "", false, "", nil)
}

func (r *MySQLGrowthRepo) listStrategyEnginePublishRecordSnapshots(jobType string) ([]model.StrategyEnginePublishRecordSummary, error) {
	jobType = strings.TrimSpace(jobType)
	if jobType == "" {
		return []model.StrategyEnginePublishRecordSummary{}, nil
	}

	rows, err := r.db.Query(`
SELECT
  j.publish_id,
  j.job_id,
  r.job_type,
  j.publish_version,
  COALESCE(DATE_FORMAT(j.created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.trade_date, '%Y-%m-%d'), ''),
  COALESCE(a.result_summary, ''),
  COALESCE(a.selected_count, 0),
  COALESCE(a.payload_count, 0),
  COALESCE(CAST(a.asset_keys AS CHAR), '')
FROM strategy_job_replays j
JOIN strategy_job_runs r ON r.job_id = j.job_id
LEFT JOIN strategy_job_artifacts a ON a.job_id = j.job_id
WHERE r.job_type = ?
  AND j.publish_id <> ''
ORDER BY j.publish_version DESC, j.created_at DESC, j.publish_id DESC`, jobType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyEnginePublishRecordSummary, 0)
	for rows.Next() {
		var item model.StrategyEnginePublishRecordSummary
		var assetKeysText string
		if err := rows.Scan(
			&item.PublishID,
			&item.JobID,
			&item.JobType,
			&item.Version,
			&item.CreatedAt,
			&item.TradeDate,
			&item.ReportSummary,
			&item.SelectedCount,
			&item.PayloadCount,
			&assetKeysText,
		); err != nil {
			return nil, err
		}
		item.AssetKeys = unmarshalStrategySnapshotStringList(assetKeysText)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) getStrategyEnginePublishRecordSnapshot(publishID string) (model.StrategyEnginePublishRecord, error) {
	publishID = strings.TrimSpace(publishID)
	if publishID == "" {
		return model.StrategyEnginePublishRecord{}, sql.ErrNoRows
	}

	var item model.StrategyEnginePublishRecord
	var reportText string
	var assetKeysText string
	var operator string
	var forcePublish bool
	var overrideReason string
	var policyText string
	var replayText string
	err := r.db.QueryRow(`
SELECT
  j.publish_id,
  j.job_id,
  r.job_type,
  j.publish_version,
  COALESCE(DATE_FORMAT(j.created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.trade_date, '%Y-%m-%d'), ''),
  COALESCE(a.result_summary, ''),
  COALESCE(a.selected_count, 0),
  COALESCE(a.payload_count, 0),
  COALESCE(CAST(a.asset_keys AS CHAR), ''),
  COALESCE(CAST(a.report_snapshot AS CHAR), ''),
  COALESCE(operator, ''),
  force_publish,
  COALESCE(override_reason, ''),
  COALESCE(CAST(policy_snapshot AS CHAR), ''),
  COALESCE(CAST(replay_snapshot AS CHAR), '')
FROM strategy_job_replays j
JOIN strategy_job_runs r ON r.job_id = j.job_id
LEFT JOIN strategy_job_artifacts a ON a.job_id = j.job_id
WHERE j.publish_id = ?
ORDER BY j.created_at DESC
LIMIT 1`, publishID).Scan(
		&item.PublishID,
		&item.JobID,
		&item.JobType,
		&item.Version,
		&item.CreatedAt,
		&item.TradeDate,
		&item.ReportSummary,
		&item.SelectedCount,
		&item.PayloadCount,
		&assetKeysText,
		&reportText,
		&operator,
		&forcePublish,
		&overrideReason,
		&policyText,
		&replayText,
	)
	if err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}

	item.ReportSnapshot = unmarshalStrategySnapshotMap(reportText)
	item.AssetKeys = unmarshalStrategySnapshotStringList(assetKeysText)
	if len(item.AssetKeys) == 0 {
		item.AssetKeys = strategySnapshotAssetKeys(item.ReportSnapshot)
	}
	item.PublishPayloads = sliceOfMaps(item.ReportSnapshot["publish_payloads"])
	item.Replay = hydrateStrategyPublishReplaySnapshot(model.StrategyEnginePublishReplay{
		PublishID:      item.PublishID,
		JobID:          item.JobID,
		PublishVersion: item.Version,
		CreatedAt:      item.CreatedAt,
	}, operator, forcePublish, overrideReason, policyText, replayText)
	return item, nil
}

func (r *MySQLGrowthRepo) getStrategyEngineJobReplayByPublishID(publishID string) (model.StrategyEnginePublishReplay, error) {
	publishID = strings.TrimSpace(publishID)
	if publishID == "" {
		return model.StrategyEnginePublishReplay{}, sql.ErrNoRows
	}

	var item model.StrategyEnginePublishReplay
	var operator string
	var forcePublish bool
	var overrideReason string
	var policyText string
	var replayText string
	if err := r.db.QueryRow(`
SELECT
  publish_id,
  job_id,
  publish_version,
  COALESCE(operator, ''),
  force_publish,
  COALESCE(override_reason, ''),
  COALESCE(CAST(policy_snapshot AS CHAR), ''),
  COALESCE(CAST(replay_snapshot AS CHAR), ''),
  COALESCE(DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), '')
FROM strategy_job_replays
WHERE publish_id = ?
ORDER BY created_at DESC
LIMIT 1`, publishID).Scan(
		&item.PublishID,
		&item.JobID,
		&item.PublishVersion,
		&operator,
		&forcePublish,
		&overrideReason,
		&policyText,
		&replayText,
		&item.CreatedAt,
	); err != nil {
		return model.StrategyEnginePublishReplay{}, err
	}
	return hydrateStrategyPublishReplaySnapshot(item, operator, forcePublish, overrideReason, policyText, replayText), nil
}

func (r *MySQLGrowthRepo) listStrategyEngineJobReplays(jobID string) ([]model.StrategyEnginePublishReplay, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return []model.StrategyEnginePublishReplay{}, nil
	}
	rows, err := r.db.Query(`
SELECT
  publish_id,
  job_id,
  publish_version,
  COALESCE(operator, ''),
  force_publish,
  COALESCE(override_reason, ''),
  COALESCE(CAST(policy_snapshot AS CHAR), ''),
  COALESCE(CAST(replay_snapshot AS CHAR), ''),
  COALESCE(DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), '')
FROM strategy_job_replays
WHERE job_id = ?
ORDER BY created_at DESC, publish_id DESC`, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyEnginePublishReplay, 0)
	for rows.Next() {
		var item model.StrategyEnginePublishReplay
		var operator string
		var forcePublish bool
		var overrideReason string
		var policyText string
		var replayText string
		if err := rows.Scan(
			&item.PublishID,
			&item.JobID,
			&item.PublishVersion,
			&operator,
			&forcePublish,
			&overrideReason,
			&policyText,
			&replayText,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, hydrateStrategyPublishReplaySnapshot(item, operator, forcePublish, overrideReason, policyText, replayText))
	}
	return items, rows.Err()
}

func hydrateStrategyPublishReplaySnapshot(item model.StrategyEnginePublishReplay, operator string, forcePublish bool, overrideReason string, policyText string, replayText string) model.StrategyEnginePublishReplay {
	item.Operator = operator
	item.ForcePublish = forcePublish
	item.OverrideReason = overrideReason
	item.PolicySnapshot = unmarshalStrategySnapshotMap(policyText)
	item.StorageSource = "LOCAL_ARCHIVED"

	replaySnapshot := unmarshalStrategySnapshotMap(replayText)
	item.WarningCount = asInt(replaySnapshot["warning_count"])
	item.WarningMessages = stringSlice(replaySnapshot["warning_messages"])
	item.VetoedAssets = stringSlice(replaySnapshot["vetoed_assets"])
	item.InvalidatedAssets = stringSlice(replaySnapshot["invalidated_assets"])
	item.Notes = stringSlice(replaySnapshot["notes"])
	return item
}

func scanStrategyEngineJobSnapshotRow(scanner interface{ Scan(dest ...any) error }) (model.StrategyEngineJobRecord, error) {
	var item model.StrategyEngineJobRecord
	var payloadText string
	var syncedAt string
	var resultSummary string
	var payloadEchoText string
	var warningsText string
	var artifactsText string
	if err := scanner.Scan(
		&item.JobID,
		&item.JobType,
		&item.Status,
		&item.RequestedBy,
		&item.TraceID,
		&item.TradeDate,
		&payloadText,
		&item.ErrorMessage,
		&item.CreatedAt,
		&item.StartedAt,
		&item.FinishedAt,
		&syncedAt,
		&resultSummary,
		&payloadEchoText,
		&warningsText,
		&artifactsText,
	); err != nil {
		return model.StrategyEngineJobRecord{}, err
	}
	item.SyncedAt = syncedAt
	item.Payload = unmarshalStrategySnapshotMap(payloadText)
	artifacts := unmarshalStrategySnapshotMap(artifactsText)
	payloadEcho := unmarshalStrategySnapshotMap(payloadEchoText)
	warnings := unmarshalStrategySnapshotStringList(warningsText)
	if resultSummary != "" || len(artifacts) > 0 || len(warnings) > 0 || len(payloadEcho) > 0 {
		item.Result = &model.StrategyEngineJobResult{
			Summary:     resultSummary,
			PayloadEcho: payloadEcho,
			Artifacts:   artifacts,
			Warnings:    warnings,
		}
	}
	return hydrateStrategyEngineJobSummary(item), nil
}

func strategySnapshotReport(result *model.StrategyEngineJobResult) map[string]any {
	if result == nil {
		return map[string]any{}
	}
	report, ok := result.Artifacts["report"].(map[string]any)
	if !ok || report == nil {
		return map[string]any{}
	}
	return report
}

func strategySnapshotCounts(report map[string]any) (int, int) {
	selectedCount := asInt(report["selected_count"])
	payloadCount := len(sliceOfMaps(report["publish_payloads"]))
	return selectedCount, payloadCount
}

func hydrateStrategyEngineJobSummary(item model.StrategyEngineJobRecord) model.StrategyEngineJobRecord {
	if strings.TrimSpace(item.TradeDate) == "" {
		item.TradeDate = firstNonEmptyString(
			strategySnapshotTradeDate(item.Payload),
			strategySnapshotTradeDate(item.ResultPayloadEcho()),
			strategySnapshotReportTradeDate(item.Result),
		)
	}
	if item.Result != nil {
		if strings.TrimSpace(item.ResultSummary) == "" {
			item.ResultSummary = strings.TrimSpace(item.Result.Summary)
		}
		item.WarningCount = len(item.Result.Warnings)
		report := strategySnapshotReport(item.Result)
		if len(report) > 0 {
			item.SelectedCount, item.PayloadCount = strategySnapshotCounts(report)
		}
	}
	item = hydrateStrategyEngineJobPublishSummary(item)
	return item
}

func hydrateStrategyEngineJobPublishSummary(item model.StrategyEngineJobRecord) model.StrategyEngineJobRecord {
	if len(item.Replays) == 0 {
		if item.PublishCount < 0 {
			item.PublishCount = 0
		}
		item.LatestPublishMode = normalizeStrategyPublishMode(item.LatestPublishMode)
		item.LatestPublishSource = strings.TrimSpace(item.LatestPublishSource)
		return item
	}

	replays := append([]model.StrategyEnginePublishReplay{}, item.Replays...)
	sort.SliceStable(replays, func(i, j int) bool {
		return strings.TrimSpace(replays[i].CreatedAt) > strings.TrimSpace(replays[j].CreatedAt)
	})
	latest := replays[0]
	item.PublishCount = len(replays)
	item.LatestPublishID = strings.TrimSpace(latest.PublishID)
	item.LatestPublishVersion = latest.PublishVersion
	item.LatestPublishAt = strings.TrimSpace(latest.CreatedAt)
	item.LatestPublishMode = normalizeStrategyReplayMode(latest)
	item.LatestPublishSource = strings.TrimSpace(latest.StorageSource)
	return item
}

func normalizeStrategyReplayMode(replay model.StrategyEnginePublishReplay) string {
	if replay.ForcePublish {
		return "OVERRIDE"
	}
	return "POLICY"
}

func normalizeStrategyPublishMode(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "OVERRIDE":
		return "OVERRIDE"
	case "POLICY":
		return "POLICY"
	default:
		return ""
	}
}

func strategySnapshotAssetKeys(report map[string]any) []string {
	keys := make([]string, 0)
	seen := make(map[string]struct{})
	appendKey := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		keys = append(keys, value)
	}

	for _, payload := range sliceOfMaps(report["publish_payloads"]) {
		if reco := mapOfAny(payload["recommendation"]); len(reco) > 0 {
			appendKey(asString(reco["symbol"]))
		}
		if strategy := mapOfAny(payload["strategy"]); len(strategy) > 0 {
			appendKey(asString(strategy["contract"]))
		}
	}
	for _, item := range sliceOfMaps(report["candidates"]) {
		appendKey(asString(item["symbol"]))
	}
	for _, item := range sliceOfMaps(report["strategies"]) {
		appendKey(asString(item["contract"]))
	}
	for _, item := range sliceOfMaps(report["simulations"]) {
		appendKey(asString(item["asset_key"]))
	}
	return keys
}

func strategySnapshotTradeDate(payload map[string]any) string {
	if payload == nil {
		return ""
	}
	return strings.TrimSpace(asString(payload["trade_date"]))
}

func strategySnapshotReportTradeDate(result *model.StrategyEngineJobResult) string {
	return strings.TrimSpace(asString(strategySnapshotReport(result)["trade_date"]))
}

func marshalStrategySnapshotJSON(value any) (string, error) {
	if value == nil {
		return "", nil
	}
	body, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func unmarshalStrategySnapshotMap(text string) map[string]any {
	text = strings.TrimSpace(text)
	if text == "" {
		return map[string]any{}
	}
	payload := make(map[string]any)
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return map[string]any{}
	}
	return payload
}

func unmarshalStrategySnapshotStringList(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		return nil
	}
	return items
}

func nullableJSONString(text string) any {
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return text
}

func nullableDateString(text string) any {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	return text
}

func parseStrategySnapshotTimeNullable(value string) (any, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	parsed, err := parseStrategyEngineTime(value)
	if err != nil {
		return nil, err
	}
	return parsed.UTC(), nil
}

func snapshotMap(payload map[string]any, key string) map[string]any {
	if payload == nil {
		return nil
	}
	if key == "" {
		return payload
	}
	mapped, ok := payload[key].(map[string]any)
	if !ok || mapped == nil {
		return nil
	}
	return mapped
}

func mapOfAny(value any) map[string]any {
	mapped, ok := value.(map[string]any)
	if !ok || mapped == nil {
		return map[string]any{}
	}
	return mapped
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func strategySnapshotJobRecordWithSource(item model.StrategyEngineJobRecord, source string) model.StrategyEngineJobRecord {
	item.StorageSource = source
	if item.SyncedAt == "" {
		item.SyncedAt = time.Now().UTC().Format(time.RFC3339)
	}
	return item
}

func isNotFoundStrategySnapshot(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func asInt(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float32:
		return int(typed)
	case float64:
		return int(typed)
	case json.Number:
		parsed, err := typed.Int64()
		if err == nil {
			return int(parsed)
		}
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		if err == nil {
			return parsed
		}
	}
	return 0
}
