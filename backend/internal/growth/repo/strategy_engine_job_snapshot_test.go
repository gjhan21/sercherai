package repo

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const (
	jobReplayListQueryPattern    = `(?s)SELECT\s+publish_id,\s+job_id,\s+publish_version,\s+COALESCE\(operator, ''\),\s+force_publish.*FROM strategy_job_replays\s+WHERE job_id = \?`
	jobSnapshotCountQueryPattern = `(?s)SELECT COUNT\(\*\) FROM strategy_job_runs r WHERE`
	jobSnapshotListQueryPattern  = `(?s)SELECT\s+r\.job_id,\s+r\.job_type,\s+r\.status,\s+COALESCE\(r\.requested_by, ''\),\s+COALESCE\(r\.trace_id, ''\),\s+COALESCE\(DATE_FORMAT\(r\.trade_date, '%Y-%m-%d'\), ''\).*FROM strategy_job_runs r`
)

func TestAdminGetStrategyEngineJobUsesLocalPayloadEchoSnapshotWithoutRemoteFetch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildRemotePublishRecordDetail(
		"publish_job_payload_echo_001",
		"job_payload_echo_001",
		5,
		"2026-03-18",
		2,
		2,
		[]string{"600519.SH", "300750.SZ"},
		[]string{"本地 payload echo warning"},
		[]string{},
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected remote request: %s %s", r.Method, r.URL.Path)
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
		WillReturnRows(newJobSnapshotRows(record))
	mock.ExpectQuery(jobReplayListQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(sqlmock.NewRows([]string{
			"publish_id",
			"job_id",
			"publish_version",
			"operator",
			"force_publish",
			"override_reason",
			"policy_snapshot",
			"replay_snapshot",
			"created_at",
		}).AddRow(
			record.PublishID,
			record.JobID,
			record.Version,
			"ops-admin",
			false,
			"",
			"",
			mustJSON(map[string]any{
				"warning_count":      1,
				"warning_messages":   []string{"本地 payload echo warning"},
				"vetoed_assets":      []string{},
				"invalidated_assets": []string{},
				"notes":              []string{"本地归档 replay"},
			}),
			record.CreatedAt,
		))

	item, err := repo.AdminGetStrategyEngineJob(record.JobID)
	if err != nil {
		t.Fatalf("AdminGetStrategyEngineJob returned error: %v", err)
	}
	if item.JobID != record.JobID || item.StorageSource != "LOCAL_ARCHIVED" {
		t.Fatalf("unexpected local job snapshot identity: %+v", item)
	}
	if item.TradeDate != record.TradeDate || item.ResultSummary != record.ReportSummary {
		t.Fatalf("expected top-level summary fields from local snapshot, got %+v", item)
	}
	if item.SelectedCount != record.SelectedCount || item.PayloadCount != record.PayloadCount || item.WarningCount != 1 {
		t.Fatalf("unexpected local summary counters: %+v", item)
	}
	if item.PublishCount != 1 || item.LatestPublishID != record.PublishID || item.LatestPublishVersion != record.Version {
		t.Fatalf("expected publish summary fields from replay snapshot, got %+v", item)
	}
	if item.LatestPublishMode != "POLICY" || item.LatestPublishSource != "LOCAL_ARCHIVED" || item.LatestPublishAt != record.CreatedAt {
		t.Fatalf("unexpected latest publish summary: %+v", item)
	}
	if item.Result == nil {
		t.Fatalf("expected result payload from local snapshot, got %+v", item)
	}
	if tradeDate := item.Result.PayloadEcho["trade_date"]; tradeDate != record.TradeDate {
		t.Fatalf("expected payload_echo trade_date %s, got %+v", record.TradeDate, item.Result.PayloadEcho)
	}
	if jobID := item.Result.PayloadEcho["job_id"]; jobID != record.JobID {
		t.Fatalf("expected payload_echo job_id %s, got %+v", record.JobID, item.Result.PayloadEcho)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminGetStrategyEngineJobBackfillsPayloadEchoSnapshotFromRemote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	jobID := "job_remote_payload_echo_001"
	payloadEcho := map[string]any{
		"trade_date": "2026-03-19",
		"job_id":     jobID,
		"seed_count": 3,
	}
	payloadEchoJSON := mustJSON(payloadEcho)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/jobs/"+jobID:
			_ = json.NewEncoder(w).Encode(strategyEngineAdminJobRecord{
				JobID:       jobID,
				JobType:     "stock-selection",
				Status:      "SUCCEEDED",
				RequestedBy: "ops-admin",
				TraceID:     "trace-" + jobID,
				Payload: map[string]any{
					"seed_symbols": []string{"600519.SH", "300750.SZ"},
				},
				Result: &strategyEngineAdminJobResult{
					Summary:     "远端任务回填",
					PayloadEcho: payloadEcho,
					Artifacts: map[string]any{
						"report": map[string]any{
							"trade_date":       "2026-03-19",
							"selected_count":   1,
							"publish_payloads": []map[string]any{{"recommendation": map[string]any{"symbol": "600519.SH"}}},
							"candidates": []map[string]any{
								{"symbol": "600519.SH"},
							},
						},
					},
					Warnings: []string{"payload echo should be archived"},
				},
				CreatedAt:  "2026-03-19T09:00:00Z",
				StartedAt:  "2026-03-19T09:00:10Z",
				FinishedAt: "2026-03-19T09:00:20Z",
			})
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
		WithArgs(jobID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO strategy_job_runs`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_artifacts`).
		WithArgs(
			jobID,
			"远端任务回填",
			payloadEchoJSON,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
			1,
			1,
			sqlmock.AnyArg(),
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(jobReplayListQueryPattern).
		WithArgs(jobID).
		WillReturnRows(sqlmock.NewRows([]string{
			"publish_id",
			"job_id",
			"publish_version",
			"operator",
			"force_publish",
			"override_reason",
			"policy_snapshot",
			"replay_snapshot",
			"created_at",
		}))

	item, err := repo.AdminGetStrategyEngineJob(jobID)
	if err != nil {
		t.Fatalf("AdminGetStrategyEngineJob returned error: %v", err)
	}
	if item.StorageSource != "REMOTE_BACKFILLED" {
		t.Fatalf("expected remote backfilled storage source, got %+v", item)
	}
	if item.TradeDate != "2026-03-19" || item.ResultSummary != "远端任务回填" {
		t.Fatalf("expected remote top-level summary fields, got %+v", item)
	}
	if item.SelectedCount != 1 || item.PayloadCount != 1 || item.WarningCount != 1 {
		t.Fatalf("unexpected remote summary counters: %+v", item)
	}
	if item.PublishCount != 0 || item.LatestPublishID != "" || item.LatestPublishMode != "" {
		t.Fatalf("expected remote job without replay summary to stay empty, got %+v", item)
	}
	if item.Result == nil {
		t.Fatalf("expected remote payload echo in result, got %+v", item)
	}
	if tradeDate := item.Result.PayloadEcho["trade_date"]; tradeDate != "2026-03-19" {
		t.Fatalf("unexpected payload_echo trade_date: %+v", item.Result.PayloadEcho)
	}
	if seedCount := item.Result.PayloadEcho["seed_count"]; seedCount != float64(3) {
		t.Fatalf("unexpected payload_echo seed_count: %+v", item.Result.PayloadEcho)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminListStrategyEngineJobsUsesTopLevelSnapshotSummaries(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildRemotePublishRecordDetail(
		"publish_job_list_001",
		"job_list_001",
		6,
		"2026-03-20",
		2,
		2,
		[]string{"600519.SH", "300750.SZ"},
		[]string{"列表 warning 1", "列表 warning 2"},
		[]string{},
	)

	replayRows := sqlmock.NewRows([]string{
		"publish_id",
		"job_id",
		"publish_version",
		"operator",
		"force_publish",
		"override_reason",
		"policy_snapshot",
		"replay_snapshot",
		"created_at",
	}).AddRow(
		record.PublishID,
		record.JobID,
		record.Version,
		"ops-admin",
		false,
		"",
		"",
		mustJSON(map[string]any{
			"warning_count":      2,
			"warning_messages":   []string{"列表 warning 1", "列表 warning 2"},
			"vetoed_assets":      []string{},
			"invalidated_assets": []string{},
			"notes":              []string{"本地归档 replay"},
		}),
		record.CreatedAt,
	)

	repo := &MySQLGrowthRepo{db: db}

	mock.ExpectQuery(jobSnapshotCountQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(jobSnapshotListQueryPattern).
		WithArgs(20, 0).
		WillReturnRows(newJobSnapshotRows(record))
	mock.ExpectQuery(jobReplayListQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(replayRows)

	items, total, err := repo.AdminListStrategyEngineJobs("", "", 1, 20)
	if err != nil {
		t.Fatalf("AdminListStrategyEngineJobs returned error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected one local snapshot job, got total=%d items=%+v", total, items)
	}
	if items[0].TradeDate != record.TradeDate || items[0].SelectedCount != record.SelectedCount {
		t.Fatalf("expected top-level list summaries, got %+v", items[0])
	}
	if items[0].PayloadCount != record.PayloadCount || items[0].WarningCount != 1 {
		t.Fatalf("unexpected top-level list counters: %+v", items[0])
	}
	if items[0].ResultSummary != record.ReportSummary || len(items[0].Replays) != 1 {
		t.Fatalf("expected summary text and replay list, got %+v", items[0])
	}
	if items[0].PublishCount != 1 || items[0].LatestPublishID != record.PublishID || items[0].LatestPublishVersion != record.Version {
		t.Fatalf("expected list publish summary fields, got %+v", items[0])
	}
	if items[0].LatestPublishMode != "POLICY" || items[0].LatestPublishSource != "LOCAL_ARCHIVED" || items[0].LatestPublishAt != record.CreatedAt {
		t.Fatalf("unexpected latest publish summary fields: %+v", items[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
