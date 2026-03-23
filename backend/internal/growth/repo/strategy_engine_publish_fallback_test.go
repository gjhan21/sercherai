package repo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestAdminPublishStrategyEngineJobFallsBackToArchivedSnapshotWhenLiveJobMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildRemotePublishRecordDetail(
		"publish_local_fallback_seed",
		"job_local_fallback_001",
		0,
		"2026-03-23",
		2,
		2,
		[]string{"600519.SH", "300750.SZ"},
		nil,
		nil,
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/internal/v1/publish/jobs/"+record.JobID:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"job not found"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	repo := &MySQLGrowthRepo{
		db: db,
		strategyEngine: &strategyEngineClient{
			baseURL:      server.URL,
			httpClient:   &http.Client{Timeout: 2 * time.Second},
			pollInterval: 10 * time.Millisecond,
		},
	}

	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newArchivedPublishJobSnapshotRows(record))
	mock.ExpectQuery(jobReplayListQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newEmptyStrategyReplayRows())
	expectResolveActiveStrategyPublishPolicy(mock, "ALL", `{"id":"policy_default_all","name":"默认发布门槛","target_type":"ALL","status":"ACTIVE","is_default":true,"max_risk_level":"MEDIUM","max_warning_count":3,"allow_vetoed_publish":false,"default_publisher":"strategy-engine","override_note_template":"人工覆盖发布，需记录原因与复盘结论。","description":"系统自动落库的默认发布策略，可在后台继续编辑与审计。","updated_by":"system-bootstrap"}`)
	mock.ExpectExec(`INSERT INTO strategy_job_runs`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_artifacts`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_replays`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	published, err := repo.AdminPublishStrategyEngineJob(record.JobID, "ops-admin", false, "")
	if err != nil {
		t.Fatalf("AdminPublishStrategyEngineJob returned error: %v", err)
	}

	if strings.TrimSpace(published.PublishID) == "" {
		t.Fatalf("expected fallback publish id, got %+v", published)
	}
	if published.JobID != record.JobID || published.JobType != "stock-selection" {
		t.Fatalf("unexpected fallback publish identity: %+v", published)
	}
	if published.Version != 1 {
		t.Fatalf("expected fallback publish version 1, got %+v", published)
	}
	if published.TradeDate != record.TradeDate || published.ReportSummary != record.ReportSummary {
		t.Fatalf("unexpected fallback publish summary: %+v", published)
	}
	if published.SelectedCount != record.SelectedCount || published.PayloadCount != record.PayloadCount {
		t.Fatalf("unexpected fallback publish counters: %+v", published)
	}
	if len(published.PublishPayloads) != record.PayloadCount {
		t.Fatalf("expected %d publish payloads, got %+v", record.PayloadCount, published.PublishPayloads)
	}
	if published.Replay.WarningCount != 0 || len(published.Replay.WarningMessages) != 0 {
		t.Fatalf("expected empty warning replay, got %+v", published.Replay)
	}
	if published.Replay.ForcePublish {
		t.Fatalf("expected policy publish instead of force publish, got %+v", published.Replay)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminPublishStrategyEngineJobFallbackStillAppliesPublishPolicy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildRemotePublishRecordDetail(
		"publish_local_blocked_seed",
		"job_local_blocked_001",
		0,
		"2026-03-23",
		2,
		2,
		[]string{"600519.SH", "300750.SZ"},
		[]string{"warning-1", "warning-2", "warning-3", "warning-4"},
		nil,
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/internal/v1/publish/jobs/"+record.JobID:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"job not found"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	repo := &MySQLGrowthRepo{
		db: db,
		strategyEngine: &strategyEngineClient{
			baseURL:      server.URL,
			httpClient:   &http.Client{Timeout: 2 * time.Second},
			pollInterval: 10 * time.Millisecond,
		},
	}

	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newArchivedPublishJobSnapshotRows(record))
	mock.ExpectQuery(jobReplayListQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newEmptyStrategyReplayRows())
	expectResolveActiveStrategyPublishPolicy(mock, "ALL", `{"id":"policy_default_all","name":"默认发布门槛","target_type":"ALL","status":"ACTIVE","is_default":true,"max_risk_level":"MEDIUM","max_warning_count":3,"allow_vetoed_publish":false,"default_publisher":"strategy-engine","override_note_template":"人工覆盖发布，需记录原因与复盘结论。","description":"系统自动落库的默认发布策略，可在后台继续编辑与审计。","updated_by":"system-bootstrap"}`)

	_, err = repo.AdminPublishStrategyEngineJob(record.JobID, "ops-admin", false, "")
	if err == nil {
		t.Fatal("expected publish policy conflict error, got nil")
	}
	if !strings.Contains(err.Error(), "发布策略拦截") || !strings.Contains(err.Error(), "警告数量 4 超过阈值 3") {
		t.Fatalf("expected publish policy conflict detail, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func expectResolveActiveStrategyPublishPolicy(mock sqlmock.Sqlmock, targetType string, configValue string) {
	rows := sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}).
		AddRow(
			strategyPublishPolicyConfigPrefix+"policy_default_"+strings.ToLower(targetType),
			configValue,
			"系统自动落库的默认发布策略，可在后台继续编辑与审计。",
			"system-bootstrap",
			time.Date(2026, 3, 23, 9, 0, 0, 0, time.Local),
		)
	mock.ExpectQuery(strategyPublishPoliciesAllQueryPattern).
		WithArgs(strategyPublishPolicyConfigPrefix + "%").
		WillReturnRows(rows)
	mock.ExpectQuery(strategyPublishPoliciesAllQueryPattern).
		WithArgs(strategyPublishPolicyConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}).
			AddRow(
				strategyPublishPolicyConfigPrefix+"policy_default_"+strings.ToLower(targetType),
				configValue,
				"系统自动落库的默认发布策略，可在后台继续编辑与审计。",
				"system-bootstrap",
				time.Date(2026, 3, 23, 9, 0, 0, 0, time.Local),
			))
}

func newEmptyStrategyReplayRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"publish_id",
		"job_id",
		"publish_version",
		"operator",
		"force_publish",
		"override_reason",
		"policy_snapshot",
		"replay_snapshot",
		"created_at",
	})
}

func newArchivedPublishJobSnapshotRows(record strategyEnginePublishRecordDetail) *sqlmock.Rows {
	payloadJSON := mustJSON(map[string]any{
		"trade_date":   record.TradeDate,
		"seed_symbols": record.AssetKeys,
	})
	payloadEchoJSON := mustJSON(map[string]any{
		"trade_date": record.TradeDate,
		"job_id":     record.JobID,
	})
	artifactsJSON := mustJSON(map[string]any{
		"report": record.ReportSnapshot,
	})
	return sqlmock.NewRows([]string{
		"job_id",
		"job_type",
		"status",
		"requested_by",
		"trace_id",
		"trade_date",
		"payload_snapshot",
		"error_message",
		"remote_created_at",
		"remote_started_at",
		"remote_finished_at",
		"synced_at",
		"result_summary",
		"payload_echo_snapshot",
		"warning_messages",
		"artifacts_snapshot",
	}).AddRow(
		record.JobID,
		record.JobType,
		"SUCCEEDED",
		"ops-admin",
		"trace-"+record.JobID,
		record.TradeDate,
		payloadJSON,
		"",
		record.CreatedAt,
		record.CreatedAt,
		record.CreatedAt,
		record.CreatedAt,
		record.ReportSummary,
		payloadEchoJSON,
		mustJSON(record.Replay.WarningMessages),
		artifactsJSON,
	)
}
