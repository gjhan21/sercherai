package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func normalizeAdminAuditEvent(item model.AdminAuditEvent) model.AdminAuditEvent {
	item.EventDomain = truncateByRunes(strings.ToUpper(normalizeUTF8Text(strings.TrimSpace(item.EventDomain))), 32)
	if item.EventDomain == "" {
		item.EventDomain = "SYSTEM"
	}
	item.EventType = truncateByRunes(strings.ToUpper(normalizeUTF8Text(strings.TrimSpace(item.EventType))), 64)
	item.Level = truncateByRunes(strings.ToUpper(normalizeUTF8Text(strings.TrimSpace(item.Level))), 16)
	if item.Level == "" {
		item.Level = "INFO"
	}
	item.Module = truncateByRunes(strings.ToUpper(normalizeUTF8Text(strings.TrimSpace(item.Module))), 32)
	item.ObjectType = truncateByRunes(strings.ToUpper(normalizeUTF8Text(strings.TrimSpace(item.ObjectType))), 64)
	item.ObjectID = truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.ObjectID)), 128)
	item.ActorUserID = truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.ActorUserID)), 64)
	item.Title = truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Title)), 128)
	item.Summary = truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Summary)), 512)
	item.Detail = truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Detail)), 2048)
	item.Status = truncateByRunes(strings.ToUpper(normalizeUTF8Text(strings.TrimSpace(item.Status))), 16)
	if item.Status == "" {
		item.Status = "OPEN"
	}
	item.DedupeKey = truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.DedupeKey)), 191)
	return item
}

func normalizeAdminAuditEventFilter(filter model.AdminAuditEventFilter) model.AdminAuditEventFilter {
	filter.EventDomain = strings.ToUpper(strings.TrimSpace(filter.EventDomain))
	filter.EventType = strings.ToUpper(strings.TrimSpace(filter.EventType))
	filter.Level = strings.ToUpper(strings.TrimSpace(filter.Level))
	filter.Module = strings.ToUpper(strings.TrimSpace(filter.Module))
	filter.ObjectType = strings.ToUpper(strings.TrimSpace(filter.ObjectType))
	filter.ObjectID = strings.TrimSpace(filter.ObjectID)
	filter.ActorUserID = strings.TrimSpace(filter.ActorUserID)
	filter.Status = strings.ToUpper(strings.TrimSpace(filter.Status))
	return filter
}

func marshalAdminAuditEventMetadata(metadata map[string]any) string {
	if len(metadata) == 0 {
		return ""
	}
	body, err := json.Marshal(metadata)
	if err != nil {
		return ""
	}
	return string(body)
}

func parseAdminAuditEventMetadata(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil
	}
	return result
}

func auditEventMetadataString(metadata map[string]any, key string) string {
	if len(metadata) == 0 {
		return ""
	}
	raw, ok := metadata[key]
	if !ok || raw == nil {
		return ""
	}
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	case fmt.Stringer:
		return strings.TrimSpace(value.String())
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func buildAuditEventWorkflowRoute(item model.AdminAuditEvent) (model.WorkflowMessage, bool) {
	receiverID := auditEventMetadataString(item.Metadata, "receiver_id")
	if receiverID == "" {
		receiverID = auditEventMetadataString(item.Metadata, "submitter_id")
	}
	if receiverID == "" && strings.EqualFold(item.EventDomain, "PUBLISH") {
		receiverID = strings.TrimSpace(item.ActorUserID)
	}
	if receiverID == "" {
		return model.WorkflowMessage{}, false
	}

	module := strings.ToUpper(strings.TrimSpace(item.Module))
	if module == "" {
		module = "SYSTEM"
	}
	content := strings.TrimSpace(item.Summary)
	if detail := strings.TrimSpace(item.Detail); detail != "" {
		if content != "" {
			content += "；" + detail
		} else {
			content = detail
		}
	}
	if content == "" {
		content = strings.TrimSpace(item.Title)
	}

	message := model.WorkflowMessage{
		ReviewID:   auditEventMetadataString(item.Metadata, "review_id"),
		TargetID:   strings.TrimSpace(item.ObjectID),
		Module:     module,
		ReceiverID: receiverID,
		SenderID:   strings.TrimSpace(item.ActorUserID),
		EventType:  strings.ToUpper(strings.TrimSpace(item.EventType)),
		Title:      strings.TrimSpace(item.Title),
		Content:    content,
	}
	if message.TargetID == "" {
		message.TargetID = item.ID
	}
	return message, true
}

func auditEventLevelRank(level string) int {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "CRITICAL":
		return 3
	case "WARNING":
		return 2
	default:
		return 1
	}
}

func cloneAuditEventMetadata(metadata map[string]any) map[string]any {
	if len(metadata) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(metadata))
	for key, value := range metadata {
		cloned[key] = value
	}
	return cloned
}

func mergeOpenAuditEvent(existing model.AdminAuditEvent, incoming model.AdminAuditEvent, now time.Time) model.AdminAuditEvent {
	merged := existing
	if auditEventLevelRank(incoming.Level) >= auditEventLevelRank(existing.Level) {
		merged.Level = incoming.Level
	}
	if strings.TrimSpace(incoming.ActorUserID) != "" {
		merged.ActorUserID = incoming.ActorUserID
	}
	if strings.TrimSpace(incoming.Title) != "" {
		merged.Title = incoming.Title
	}
	if strings.TrimSpace(incoming.Summary) != "" {
		merged.Summary = incoming.Summary
	}
	if strings.TrimSpace(incoming.Detail) != "" {
		merged.Detail = incoming.Detail
	}
	metadata := cloneAuditEventMetadata(existing.Metadata)
	for key, value := range incoming.Metadata {
		metadata[key] = value
	}
	occurrenceCount := intValue(metadata["occurrence_count"])
	if occurrenceCount <= 0 {
		occurrenceCount = 1
	}
	metadata["occurrence_count"] = occurrenceCount + 1
	if stringValue(metadata["first_seen_at"]) == "" {
		metadata["first_seen_at"] = firstNonEmpty(existing.CreatedAt, now.Format(time.RFC3339))
	}
	metadata["last_seen_at"] = now.Format(time.RFC3339)
	merged.Metadata = metadata
	return merged
}

func (r *MySQLGrowthRepo) routeAuditEventWorkflowMessage(item model.AdminAuditEvent) {
	message, ok := buildAuditEventWorkflowRoute(item)
	if !ok {
		return
	}
	_ = r.AdminCreateWorkflowMessage(
		message.ReviewID,
		message.TargetID,
		message.Module,
		message.ReceiverID,
		message.SenderID,
		message.EventType,
		message.Title,
		message.Content,
	)
}

func (r *MySQLGrowthRepo) hasOpenAuditEventWithDedupeKey(dedupeKey string) (bool, error) {
	dedupeKey = strings.TrimSpace(dedupeKey)
	if dedupeKey == "" {
		return false, nil
	}
	var count int
	if err := r.db.QueryRow(`
SELECT COUNT(*)
FROM admin_audit_events
WHERE dedupe_key = ? AND status = 'OPEN'`, dedupeKey).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MySQLGrowthRepo) findOpenAuditEventByDedupeKey(dedupeKey string) (*model.AdminAuditEvent, error) {
	dedupeKey = strings.TrimSpace(dedupeKey)
	if dedupeKey == "" {
		return nil, nil
	}
	var (
		item        model.AdminAuditEvent
		detail      sql.NullString
		metadataRaw string
	)
	err := r.db.QueryRow(`
SELECT id, event_domain, event_type, level, module, object_type, object_id, actor_user_id, title, summary, detail, status, COALESCE(metadata_json, ''), DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
FROM admin_audit_events
WHERE dedupe_key = ? AND status = 'OPEN'
ORDER BY created_at DESC, id DESC
LIMIT 1`, dedupeKey).Scan(
		&item.ID,
		&item.EventDomain,
		&item.EventType,
		&item.Level,
		&item.Module,
		&item.ObjectType,
		&item.ObjectID,
		&item.ActorUserID,
		&item.Title,
		&item.Summary,
		&detail,
		&item.Status,
		&metadataRaw,
		&item.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if detail.Valid {
		item.Detail = detail.String
	}
	item.Metadata = parseAdminAuditEventMetadata(metadataRaw)
	item.DedupeKey = dedupeKey
	return &item, nil
}

func (r *MySQLGrowthRepo) updateOpenAuditEvent(existing model.AdminAuditEvent, incoming model.AdminAuditEvent) error {
	now := time.Now()
	merged := mergeOpenAuditEvent(existing, incoming, now)
	metadataText := marshalAdminAuditEventMetadata(merged.Metadata)
	_, err := r.db.Exec(`
UPDATE admin_audit_events
SET level = ?, actor_user_id = ?, title = ?, summary = ?, detail = ?, metadata_json = ?
WHERE id = ?`,
		merged.Level,
		merged.ActorUserID,
		merged.Title,
		merged.Summary,
		nullableString(merged.Detail),
		nullableString(metadataText),
		existing.ID,
	)
	return err
}

func (r *MySQLGrowthRepo) AdminCreateAuditEvent(item model.AdminAuditEvent) error {
	item = normalizeAdminAuditEvent(item)
	if item.Title == "" || item.EventType == "" {
		return nil
	}
	if item.DedupeKey != "" {
		existing, err := r.findOpenAuditEventByDedupeKey(item.DedupeKey)
		if err != nil {
			return err
		}
		if existing != nil {
			return r.updateOpenAuditEvent(*existing, item)
		}
	}
	metadataText := marshalAdminAuditEventMetadata(item.Metadata)
	_, err := r.db.Exec(`
INSERT INTO admin_audit_events (id, event_domain, event_type, level, module, object_type, object_id, actor_user_id, title, summary, detail, status, metadata_json, dedupe_key, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("ae"),
		item.EventDomain,
		item.EventType,
		item.Level,
		item.Module,
		item.ObjectType,
		item.ObjectID,
		item.ActorUserID,
		item.Title,
		item.Summary,
		nullableString(item.Detail),
		item.Status,
		nullableString(metadataText),
		item.DedupeKey,
		time.Now(),
	)
	if err == nil {
		r.routeAuditEventWorkflowMessage(item)
	}
	return err
}

func (r *MySQLGrowthRepo) AdminListAuditEvents(filter model.AdminAuditEventFilter, page int, pageSize int) ([]model.AdminAuditEvent, int, error) {
	filter = normalizeAdminAuditEventFilter(filter)
	offset := (page - 1) * pageSize
	args := make([]interface{}, 0, 8)
	where := " WHERE 1=1"
	if filter.EventDomain != "" {
		where += " AND event_domain = ?"
		args = append(args, filter.EventDomain)
	}
	if filter.Level != "" {
		where += " AND level = ?"
		args = append(args, filter.Level)
	}
	if filter.EventType != "" {
		where += " AND event_type = ?"
		args = append(args, filter.EventType)
	}
	if filter.Module != "" {
		where += " AND module = ?"
		args = append(args, filter.Module)
	}
	if filter.ObjectType != "" {
		where += " AND object_type = ?"
		args = append(args, filter.ObjectType)
	}
	if filter.ObjectID != "" {
		where += " AND object_id = ?"
		args = append(args, filter.ObjectID)
	}
	if filter.ActorUserID != "" {
		where += " AND actor_user_id = ?"
		args = append(args, filter.ActorUserID)
	}
	if filter.Status != "" {
		where += " AND status = ?"
		args = append(args, filter.Status)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM admin_audit_events"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT id, event_domain, event_type, level, module, object_type, object_id, actor_user_id, title, summary, detail, status, COALESCE(metadata_json, ''), created_at
FROM admin_audit_events`+where+`
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.AdminAuditEvent, 0, pageSize)
	for rows.Next() {
		var item model.AdminAuditEvent
		var detail, metadataText sql.NullString
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.EventDomain,
			&item.EventType,
			&item.Level,
			&item.Module,
			&item.ObjectType,
			&item.ObjectID,
			&item.ActorUserID,
			&item.Title,
			&item.Summary,
			&detail,
			&item.Status,
			&metadataText,
			&createdAt,
		); err != nil {
			return nil, 0, err
		}
		if detail.Valid {
			item.Detail = detail.String
		}
		if metadataText.Valid {
			item.Metadata = parseAdminAuditEventMetadata(metadataText.String)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminGetAuditEventSummary() (model.AdminAuditEventSummary, error) {
	summary := model.AdminAuditEventSummary{
		ByDomain: map[string]int{},
	}
	if err := r.db.QueryRow(`
SELECT
  COUNT(*),
  COALESCE(SUM(CASE WHEN status = 'OPEN' THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN level = 'INFO' THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN level = 'WARNING' THEN 1 ELSE 0 END), 0),
  COALESCE(SUM(CASE WHEN level = 'CRITICAL' THEN 1 ELSE 0 END), 0)
FROM admin_audit_events`).Scan(
		&summary.TotalCount,
		&summary.OpenCount,
		&summary.InfoCount,
		&summary.WarningCount,
		&summary.CriticalCount,
	); err != nil {
		return summary, err
	}
	rows, err := r.db.Query(`
SELECT event_domain, COUNT(*)
FROM admin_audit_events
GROUP BY event_domain`)
	if err != nil {
		return summary, err
	}
	defer rows.Close()
	for rows.Next() {
		var domain string
		var count int
		if err := rows.Scan(&domain, &count); err != nil {
			return summary, err
		}
		summary.ByDomain[domain] = count
	}
	return summary, rows.Err()
}

func strategyConfigAuditObjectType(configKey string) string {
	switch {
	case strings.HasPrefix(configKey, strategySeedSetConfigPrefix):
		return "STRATEGY_SEED_SET"
	case strings.HasPrefix(configKey, strategyAgentProfileConfigPrefix):
		return "STRATEGY_AGENT_PROFILE"
	case strings.HasPrefix(configKey, strategyScenarioTemplateConfigPrefix):
		return "STRATEGY_SCENARIO_TEMPLATE"
	case strings.HasPrefix(configKey, strategyPublishPolicyConfigPrefix):
		return "STRATEGY_PUBLISH_POLICY"
	default:
		return "SYSTEM_CONFIG"
	}
}

func strategyConfigAuditObjectID(configKey string) string {
	switch {
	case strings.HasPrefix(configKey, strategySeedSetConfigPrefix):
		return strings.TrimPrefix(configKey, strategySeedSetConfigPrefix)
	case strings.HasPrefix(configKey, strategyAgentProfileConfigPrefix):
		return strings.TrimPrefix(configKey, strategyAgentProfileConfigPrefix)
	case strings.HasPrefix(configKey, strategyScenarioTemplateConfigPrefix):
		return strings.TrimPrefix(configKey, strategyScenarioTemplateConfigPrefix)
	case strings.HasPrefix(configKey, strategyPublishPolicyConfigPrefix):
		return strings.TrimPrefix(configKey, strategyPublishPolicyConfigPrefix)
	default:
		return configKey
	}
}

func (r *MySQLGrowthRepo) recordStrategyConfigBootstrapAuditEvent(configKey string, description string, operator string) {
	configKey = strings.TrimSpace(configKey)
	if configKey == "" {
		return
	}
	objectType := strategyConfigAuditObjectType(configKey)
	objectID := strategyConfigAuditObjectID(configKey)
	_ = r.AdminCreateAuditEvent(model.AdminAuditEvent{
		EventDomain: "RESEARCH",
		EventType:   "DEFAULT_STRATEGY_CONFIG_MATERIALIZED",
		Level:       "INFO",
		Module:      "STRATEGY_ENGINE",
		ObjectType:  objectType,
		ObjectID:    objectID,
		ActorUserID: strings.TrimSpace(operator),
		Title:       "默认策略配置已自动落库",
		Summary:     "系统已自动落库默认策略配置 " + objectType + " / " + objectID,
		Status:      "OPEN",
		DedupeKey:   "strategy_config_bootstrap:" + configKey,
		Metadata: map[string]any{
			"config_key":   configKey,
			"description":  strings.TrimSpace(description),
			"bootstrap_by": strings.TrimSpace(operator),
		},
	})
}
