package repo

import (
	"encoding/json"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"sercherai/backend/internal/growth/model"
)

const adminAuditEventListQueryPattern = `(?s)SELECT\s+id,\s+event_domain,\s+event_type,\s+level,\s+module,\s+object_type,\s+object_id,\s+actor_user_id,\s+title,\s+summary,\s+detail,\s+status,\s+COALESCE\(metadata_json, ''\),\s+created_at\s+FROM admin_audit_events`

func TestAdminCreateAndListAuditEventsInMemory(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	event := model.AdminAuditEvent{
		EventDomain: "DATA",
		EventType:   "DATA_SOURCE_UNHEALTHY",
		Level:       "WARNING",
		Module:      "SYSTEM",
		ObjectType:  "DATA_SOURCE",
		ObjectID:    "TUSHARE",
		ActorUserID: "system",
		Title:       "数据源健康告警",
		Summary:     "数据源 TUSHARE 连续失败 3 次",
		Status:      "OPEN",
		Metadata: map[string]any{
			"consecutive_failures": 3,
		},
	}

	if err := repo.AdminCreateAuditEvent(event); err != nil {
		t.Fatalf("create audit event: %v", err)
	}

	items, total, err := repo.AdminListAuditEvents(model.AdminAuditEventFilter{EventDomain: "DATA", Level: "WARNING"}, 1, 20)
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 audit event, got total=%d items=%d", total, len(items))
	}
	if items[0].EventType != "DATA_SOURCE_UNHEALTHY" {
		t.Fatalf("unexpected audit event: %+v", items[0])
	}
	if got, ok := items[0].Metadata["consecutive_failures"].(int); !ok || got != 3 {
		t.Fatalf("expected consecutive_failures=3, got %+v", items[0].Metadata)
	}
}

func TestAdminListAuditEventsMySQLAppliesFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	filter := model.AdminAuditEventFilter{
		EventDomain: "PUBLISH",
		Level:       "INFO",
		EventType:   "STRATEGY_JOB_PUBLISHED",
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM admin_audit_events WHERE 1=1 AND event_domain = ? AND level = ? AND event_type = ?")).
		WithArgs("PUBLISH", "INFO", "STRATEGY_JOB_PUBLISHED").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	metadataText, _ := json.Marshal(map[string]any{"publish_id": "pub_001"})
	mock.ExpectQuery(adminAuditEventListQueryPattern).
		WithArgs("PUBLISH", "INFO", "STRATEGY_JOB_PUBLISHED", 20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "event_domain", "event_type", "level", "module", "object_type", "object_id", "actor_user_id",
			"title", "summary", "detail", "status", "metadata_json", "created_at",
		}).AddRow(
			"ae_001", "PUBLISH", "STRATEGY_JOB_PUBLISHED", "INFO", "STRATEGY_ENGINE", "STRATEGY_JOB", "job_001", "tester",
			"策略发布完成", "job_001 已发布到 publish_id pub_001", "", "OPEN", string(metadataText), time.Date(2026, 3, 24, 10, 0, 0, 0, time.Local),
		))

	items, total, err := repo.AdminListAuditEvents(filter, 1, 20)
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected 1 audit event, got total=%d items=%d", total, len(items))
	}
	if items[0].Metadata["publish_id"] != "pub_001" {
		t.Fatalf("expected metadata to round-trip, got %+v", items[0].Metadata)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPersistStrategyConfigWritesBootstrapAuditEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE config_value = VALUES(config_value), description = VALUES(description), updated_by = VALUES(updated_by), updated_at = VALUES(updated_at)`)).
		WithArgs(
			sqlmock.AnyArg(),
			strategyPublishPolicyConfigPrefix+"policy_default_all",
			sqlmock.AnyArg(),
			"系统自动落库的默认发布策略，可在后台继续编辑与审计。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id, event_domain, event_type, level, module, object_type, object_id, actor_user_id, title, summary, detail, status, COALESCE(metadata_json, ''), DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
FROM admin_audit_events
WHERE dedupe_key = ? AND status = 'OPEN'
ORDER BY created_at DESC, id DESC
LIMIT 1`)).
		WithArgs("strategy_config_bootstrap:" + strategyPublishPolicyConfigPrefix + "policy_default_all").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "event_domain", "event_type", "level", "module", "object_type", "object_id", "actor_user_id",
			"title", "summary", "detail", "status", "metadata_json", "created_at",
		}))
	mock.ExpectExec(regexp.QuoteMeta(`
INSERT INTO admin_audit_events (id, event_domain, event_type, level, module, object_type, object_id, actor_user_id, title, summary, detail, status, metadata_json, dedupe_key, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)).
		WithArgs(
			sqlmock.AnyArg(),
			"RESEARCH",
			"DEFAULT_STRATEGY_CONFIG_MATERIALIZED",
			"INFO",
			"STRATEGY_ENGINE",
			"STRATEGY_PUBLISH_POLICY",
			"policy_default_all",
			strategyBuiltInPolicyOperator,
			"默认策略配置已自动落库",
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			"OPEN",
			sqlmock.AnyArg(),
			"strategy_config_bootstrap:"+strategyPublishPolicyConfigPrefix+"policy_default_all",
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.persistStrategyConfig(
		strategyPublishPolicyConfigPrefix+"policy_default_all",
		map[string]any{"id": "policy_default_all", "name": "默认发布门槛"},
		"系统自动落库的默认发布策略，可在后台继续编辑与审计。",
		strategyBuiltInPolicyOperator,
	)
	if err != nil {
		t.Fatalf("persistStrategyConfig: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminCreateAuditEventRoutesWorkflowMessageByReceiverMetadata(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	err := repo.AdminCreateAuditEvent(model.AdminAuditEvent{
		EventDomain: "DATA",
		EventType:   "DATA_SOURCE_UNHEALTHY",
		Level:       "WARNING",
		Module:      "SYSTEM",
		ObjectType:  "DATA_SOURCE",
		ObjectID:    "TUSHARE",
		ActorUserID: "system",
		Title:       "数据源健康告警",
		Summary:     "TUSHARE 连续失败 3 次",
		Metadata: map[string]any{
			"receiver_id": "ops_admin",
		},
	})
	if err != nil {
		t.Fatalf("create audit event: %v", err)
	}

	items, total, err := repo.AdminListWorkflowMessages("SYSTEM", "DATA_SOURCE_UNHEALTHY", "", "ops_admin", 1, 20)
	if err != nil {
		t.Fatalf("list workflow messages: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected routed workflow message, got total=%d items=%d", total, len(items))
	}
	if items[0].Title != "数据源健康告警" || items[0].ReceiverID != "ops_admin" {
		t.Fatalf("unexpected routed workflow message: %+v", items[0])
	}
}

func TestAdminCreateAuditEventRoutesWorkflowMessageBySubmitterMetadata(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	err := repo.AdminCreateAuditEvent(model.AdminAuditEvent{
		EventDomain: "RESEARCH",
		EventType:   "REVIEW_REJECTED",
		Level:       "WARNING",
		Module:      "STOCK",
		ObjectType:  "REVIEW_TASK",
		ObjectID:    "rt_001",
		ActorUserID: "reviewer_001",
		Title:       "审核结果通知",
		Summary:     "审核结果为 REJECTED，目标 run_001",
		Metadata: map[string]any{
			"submitter_id": "submitter_001",
		},
	})
	if err != nil {
		t.Fatalf("create audit event: %v", err)
	}

	items, total, err := repo.AdminListWorkflowMessages("STOCK", "REVIEW_REJECTED", "", "submitter_001", 1, 20)
	if err != nil {
		t.Fatalf("list workflow messages: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected review routed workflow message, got total=%d items=%d", total, len(items))
	}
	if items[0].SenderID != "reviewer_001" {
		t.Fatalf("expected sender reviewer_001, got %+v", items[0])
	}
}

func TestAdminGetAuditEventSummaryInMemory(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	_ = repo.AdminCreateAuditEvent(model.AdminAuditEvent{
		EventDomain: "DATA",
		EventType:   "DATA_SOURCE_UNHEALTHY",
		Level:       "WARNING",
		Title:       "数据源健康告警",
		Status:      "OPEN",
	})
	_ = repo.AdminCreateAuditEvent(model.AdminAuditEvent{
		EventDomain: "PUBLISH",
		EventType:   "STRATEGY_JOB_FORCE_PUBLISHED",
		Level:       "WARNING",
		Title:       "策略人工覆盖发布",
		Status:      "OPEN",
	})
	_ = repo.AdminCreateAuditEvent(model.AdminAuditEvent{
		EventDomain: "RESEARCH",
		EventType:   "DEFAULT_STRATEGY_CONFIG_MATERIALIZED",
		Level:       "INFO",
		Title:       "默认策略配置已自动落库",
		Status:      "RESOLVED",
	})

	summary, err := repo.AdminGetAuditEventSummary()
	if err != nil {
		t.Fatalf("get audit event summary: %v", err)
	}
	if summary.TotalCount != 3 || summary.OpenCount != 2 || summary.WarningCount != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.ByDomain["DATA"] != 1 || summary.ByDomain["PUBLISH"] != 1 || summary.ByDomain["RESEARCH"] != 1 {
		t.Fatalf("unexpected domain summary: %+v", summary.ByDomain)
	}
}

func TestAdminCreateAuditEventDedupesByOpenDedupeKey(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	first := model.AdminAuditEvent{
		EventDomain: "DATA",
		EventType:   "DATA_SOURCE_UNHEALTHY",
		Level:       "WARNING",
		Module:      "SYSTEM",
		ObjectType:  "DATA_SOURCE",
		ObjectID:    "TUSHARE",
		ActorUserID: "system",
		Title:       "数据源健康告警",
		Summary:     "连续失败",
		Status:      "OPEN",
		DedupeKey:   "data_source_health:TUSHARE:unhealthy",
		Metadata: map[string]any{
			"receiver_id": "ops_admin",
		},
	}
	second := first
	second.Summary = "再次失败"

	if err := repo.AdminCreateAuditEvent(first); err != nil {
		t.Fatalf("create first event: %v", err)
	}
	if err := repo.AdminCreateAuditEvent(second); err != nil {
		t.Fatalf("create second event: %v", err)
	}

	events, total, err := repo.AdminListAuditEvents(model.AdminAuditEventFilter{EventDomain: "DATA"}, 1, 20)
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if total != 1 || len(events) != 1 {
		t.Fatalf("expected deduped audit events, got total=%d items=%d", total, len(events))
	}

	messages, messageTotal, err := repo.AdminListWorkflowMessages("SYSTEM", "DATA_SOURCE_UNHEALTHY", "", "ops_admin", 1, 20)
	if err != nil {
		t.Fatalf("list workflow messages: %v", err)
	}
	if messageTotal != 1 || len(messages) != 1 {
		t.Fatalf("expected deduped workflow messages, got total=%d items=%d", messageTotal, len(messages))
	}
}

func TestAdminCreateAuditEventDedupedOpenEventEscalatesMetadata(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	first := model.AdminAuditEvent{
		EventDomain: "DATA",
		EventType:   "DATA_SOURCE_UNHEALTHY",
		Level:       "WARNING",
		Module:      "SYSTEM",
		ObjectType:  "DATA_SOURCE",
		ObjectID:    "TUSHARE",
		ActorUserID: "system",
		Title:       "数据源健康告警",
		Summary:     "首次失败",
		Status:      "OPEN",
		DedupeKey:   "data_source_health:TUSHARE:unhealthy",
		Metadata: map[string]any{
			"receiver_id": "ops_admin",
		},
	}
	second := first
	second.Level = "CRITICAL"
	second.Summary = "连续失败升级"

	if err := repo.AdminCreateAuditEvent(first); err != nil {
		t.Fatalf("create first event: %v", err)
	}
	if err := repo.AdminCreateAuditEvent(second); err != nil {
		t.Fatalf("create second event: %v", err)
	}

	events, total, err := repo.AdminListAuditEvents(model.AdminAuditEventFilter{EventDomain: "DATA"}, 1, 20)
	if err != nil {
		t.Fatalf("list audit events: %v", err)
	}
	if total != 1 || len(events) != 1 {
		t.Fatalf("expected single merged event, got total=%d items=%d", total, len(events))
	}
	if events[0].Level != "CRITICAL" {
		t.Fatalf("expected escalated level CRITICAL, got %+v", events[0])
	}
	if events[0].Summary != "连续失败升级" {
		t.Fatalf("expected latest summary to win, got %+v", events[0])
	}
	if got := intValue(events[0].Metadata["occurrence_count"]); got != 2 {
		t.Fatalf("expected occurrence_count=2, got %+v", events[0].Metadata)
	}
	if stringValue(events[0].Metadata["last_seen_at"]) == "" {
		t.Fatalf("expected last_seen_at to be recorded, got %+v", events[0].Metadata)
	}
}

func TestAdminCreateAuditEventMySQLUpdatesOpenDedupeEvent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(regexp.QuoteMeta(`
SELECT id, event_domain, event_type, level, module, object_type, object_id, actor_user_id, title, summary, detail, status, COALESCE(metadata_json, ''), DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
FROM admin_audit_events
WHERE dedupe_key = ? AND status = 'OPEN'
ORDER BY created_at DESC, id DESC
LIMIT 1`)).
		WithArgs("data_source_health:TUSHARE:unhealthy").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "event_domain", "event_type", "level", "module", "object_type", "object_id", "actor_user_id",
			"title", "summary", "detail", "status", "metadata_json", "created_at",
		}).AddRow(
			"ae_001", "DATA", "DATA_SOURCE_UNHEALTHY", "WARNING", "SYSTEM", "DATA_SOURCE", "TUSHARE", "system",
			"数据源健康告警", "首次失败", "", "OPEN", `{"receiver_id":"ops_admin"}`, "2026-03-24T09:00:00Z",
		))
	mock.ExpectExec(regexp.QuoteMeta(`
UPDATE admin_audit_events
SET level = ?, actor_user_id = ?, title = ?, summary = ?, detail = ?, metadata_json = ?
WHERE id = ?`)).
		WithArgs(
			"CRITICAL",
			"system",
			"数据源健康告警",
			"连续失败升级",
			nil,
			sqlmock.AnyArg(),
			"ae_001",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.AdminCreateAuditEvent(model.AdminAuditEvent{
		EventDomain: "DATA",
		EventType:   "DATA_SOURCE_UNHEALTHY",
		Level:       "CRITICAL",
		Module:      "SYSTEM",
		ObjectType:  "DATA_SOURCE",
		ObjectID:    "TUSHARE",
		ActorUserID: "system",
		Title:       "数据源健康告警",
		Summary:     "连续失败升级",
		Status:      "OPEN",
		DedupeKey:   "data_source_health:TUSHARE:unhealthy",
		Metadata: map[string]any{
			"receiver_id": "ops_admin",
		},
	})
	if err != nil {
		t.Fatalf("create audit event: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
