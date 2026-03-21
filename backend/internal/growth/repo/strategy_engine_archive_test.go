package repo

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"sercherai/backend/internal/growth/model"
)

const (
	publishRecordSnapshotQueryPattern  = `(?s)SELECT\s+j\.publish_id,\s+j\.job_id,\s+r\.job_type,\s+j\.publish_version.*WHERE j\.publish_id = \?`
	publishReplaySnapshotQueryPattern  = `(?s)SELECT\s+publish_id,\s+job_id,\s+publish_version,\s+COALESCE\(operator, ''\),\s+force_publish.*WHERE publish_id = \?`
	publishHistorySnapshotQueryPattern = `(?s)SELECT\s+j\.publish_id,\s+j\.job_id,\s+r\.job_type,\s+j\.publish_version.*WHERE r\.job_type = \?`
	jobSnapshotQueryPattern            = `(?s)SELECT\s+r\.job_id,\s+r\.job_type,\s+r\.status,\s+COALESCE\(r\.requested_by, ''\),\s+COALESCE\(r\.trace_id, ''\).*WHERE r\.job_id = \?`
)

func TestAdminListStrategyEnginePublishHistoryBackfillsRemoteSummariesIntoLocalSnapshots(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildRemotePublishRecordDetail(
		"publish_history_001",
		"job_history_001",
		9,
		"2026-03-19",
		2,
		2,
		[]string{"600519.SH", "300750.SZ"},
		[]string{"风控过滤 1 条"},
		[]string{"688981.SH"},
	)
	remoteHistory := strategyEnginePublishHistoryResponse{
		Records: []strategyEnginePublishRecordSummary{
			{
				PublishID:     record.PublishID,
				JobID:         record.JobID,
				JobType:       record.JobType,
				Version:       record.Version,
				CreatedAt:     record.CreatedAt,
				TradeDate:     record.TradeDate,
				ReportSummary: "REMOTE SUMMARY",
				SelectedCount: record.SelectedCount,
				AssetKeys:     record.AssetKeys,
				PayloadCount:  record.PayloadCount,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/history/stock-selection":
			_ = json.NewEncoder(w).Encode(remoteHistory)
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/records/"+record.PublishID:
			_ = json.NewEncoder(w).Encode(record)
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/jobs/"+record.JobID:
			_ = json.NewEncoder(w).Encode(strategyEngineAdminJobRecord{
				JobID:       record.JobID,
				JobType:     record.JobType,
				Status:      "SUCCEEDED",
				RequestedBy: "ops-admin",
				TraceID:     "trace-" + record.JobID,
				Payload: map[string]any{
					"trade_date":   record.TradeDate,
					"seed_symbols": record.AssetKeys,
				},
				Result: &strategyEngineAdminJobResult{
					Summary: "history backfill job",
					Artifacts: map[string]any{
						"report": map[string]any{
							"trade_date":       record.TradeDate,
							"selected_count":   record.SelectedCount,
							"publish_payloads": record.PublishPayloads,
							"candidates": []map[string]any{
								{"symbol": firstAssetKey(record.AssetKeys)},
							},
						},
					},
					Warnings: []string{"history backfill warning"},
				},
				CreatedAt:  record.CreatedAt,
				StartedAt:  record.CreatedAt,
				FinishedAt: record.CreatedAt,
			})
		case strings.HasPrefix(r.URL.Path, "/internal/v1/jobs/"):
			http.NotFound(w, r)
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

	mock.ExpectQuery(publishHistorySnapshotQueryPattern).
		WithArgs("stock-selection").
		WillReturnRows(newPublishHistorySummaryRows())
	mock.ExpectQuery(publishRecordSnapshotQueryPattern).
		WithArgs(record.PublishID).
		WillReturnError(sql.ErrNoRows)

	items, err := repo.AdminListStrategyEnginePublishHistory("stock-selection")
	if err != nil {
		t.Fatalf("AdminListStrategyEnginePublishHistory returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one publish history item, got %+v", items)
	}
	if items[0].ReportSummary == "REMOTE SUMMARY" {
		t.Fatalf("expected backfilled summary instead of raw remote summary, got %+v", items[0])
	}
	if items[0].ReportSummary != record.ReportSummary {
		t.Fatalf("unexpected backfilled summary payload: %+v", items[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminCompareStrategyEnginePublishVersionsUsesPublishRecordsWithoutRemoteCompare(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	leftRecord := buildRemotePublishRecordDetail(
		"publish_hist_left",
		"job_hist_left",
		7,
		"2026-03-17",
		2,
		2,
		[]string{"600519.SH", "601318.SH"},
		[]string{"风控过滤 2 条", "仓位需要复核"},
		[]string{"688981.SH"},
	)
	rightRecord := buildRemotePublishRecordDetail(
		"publish_hist_right",
		"job_hist_right",
		8,
		"2026-03-18",
		3,
		3,
		[]string{"600519.SH", "300750.SZ"},
		[]string{"风控过滤 1 条"},
		[]string{"601318.SH", "002594.SZ"},
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/records/"+leftRecord.PublishID:
			_ = json.NewEncoder(w).Encode(leftRecord)
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/records/"+rightRecord.PublishID:
			_ = json.NewEncoder(w).Encode(rightRecord)
		case r.Method == http.MethodPost && r.URL.Path == "/internal/v1/publish/compare":
			t.Fatalf("unexpected remote compare call: %s", r.URL.Path)
		case strings.HasPrefix(r.URL.Path, "/internal/v1/jobs/"):
			t.Fatalf("unexpected remote job fetch: %s", r.URL.Path)
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

	expectPublishRecordBackfillViaLocalJob(mock, leftRecord)
	expectPublishRecordBackfillViaLocalJob(mock, rightRecord)

	result, err := repo.AdminCompareStrategyEnginePublishVersions(leftRecord.PublishID, rightRecord.PublishID)
	if err != nil {
		t.Fatalf("AdminCompareStrategyEnginePublishVersions returned error: %v", err)
	}

	if result.LeftPublishID != leftRecord.PublishID || result.RightPublishID != rightRecord.PublishID {
		t.Fatalf("unexpected compare ids: %+v", result)
	}
	if result.LeftVersion != 7 || result.RightVersion != 8 {
		t.Fatalf("unexpected compare versions: %+v", result)
	}
	if result.SelectedCountDelta != 1 || result.PayloadCountDelta != 1 {
		t.Fatalf("unexpected compare deltas: %+v", result)
	}
	if result.WarningCountDelta != -1 || result.VetoCountDelta != 1 {
		t.Fatalf("unexpected replay deltas: %+v", result)
	}
	if strings.Join(result.AddedAssets, ",") != "300750.SZ" {
		t.Fatalf("unexpected added assets: %+v", result.AddedAssets)
	}
	if strings.Join(result.RemovedAssets, ",") != "601318.SH" {
		t.Fatalf("unexpected removed assets: %+v", result.RemovedAssets)
	}
	if strings.Join(result.SharedAssets, ",") != "600519.SH" {
		t.Fatalf("unexpected shared assets: %+v", result.SharedAssets)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminGetStrategyEnginePublishReplayBackfillsFromPublishRecordBeforeRemoteReplay(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildRemotePublishRecordDetail(
		"publish_replay_001",
		"job_replay_001",
		11,
		"2026-03-18",
		2,
		2,
		[]string{"600519.SH", "300750.SZ"},
		[]string{"风控过滤 2 条", "复核后允许发布"},
		[]string{"688981.SH"},
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/records/"+record.PublishID:
			_ = json.NewEncoder(w).Encode(record)
		case strings.HasSuffix(r.URL.Path, "/replay"):
			t.Fatalf("unexpected remote replay fetch: %s", r.URL.Path)
		case strings.HasPrefix(r.URL.Path, "/internal/v1/jobs/"):
			t.Fatalf("unexpected remote job fetch: %s", r.URL.Path)
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

	mock.ExpectQuery(publishReplaySnapshotQueryPattern).
		WithArgs(record.PublishID).
		WillReturnError(sql.ErrNoRows)
	expectPublishRecordBackfillViaLocalJob(mock, record)
	mock.ExpectQuery(publishReplaySnapshotQueryPattern).
		WithArgs(record.PublishID).
		WillReturnRows(newPublishReplayRows(record))

	replay, err := repo.AdminGetStrategyEnginePublishReplay(record.PublishID)
	if err != nil {
		t.Fatalf("AdminGetStrategyEnginePublishReplay returned error: %v", err)
	}

	if replay.PublishID != record.PublishID || replay.JobID != record.JobID {
		t.Fatalf("unexpected replay identity: %+v", replay)
	}
	if replay.PublishVersion != record.Version {
		t.Fatalf("unexpected replay version: %+v", replay)
	}
	if replay.StorageSource != "LOCAL_ARCHIVED" {
		t.Fatalf("expected local archived replay, got %+v", replay)
	}
	if replay.WarningCount != 2 || len(replay.WarningMessages) != 2 || len(replay.VetoedAssets) != 1 {
		t.Fatalf("unexpected replay payload: %+v", replay)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func expectPublishRecordBackfillViaLocalJob(mock sqlmock.Sqlmock, record strategyEnginePublishRecordDetail) {
	mock.ExpectQuery(publishRecordSnapshotQueryPattern).
		WithArgs(record.PublishID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newJobSnapshotRows(record))
	mock.ExpectExec(`INSERT INTO strategy_job_runs`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_artifacts`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_replays`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(publishRecordSnapshotQueryPattern).
		WithArgs(record.PublishID).
		WillReturnError(sql.ErrNoRows)
}

func newJobSnapshotRows(record strategyEnginePublishRecordDetail) *sqlmock.Rows {
	payloadJSON := mustJSON(map[string]any{
		"trade_date":   record.TradeDate,
		"seed_symbols": record.AssetKeys,
	})
	payloadEchoJSON := mustJSON(map[string]any{
		"trade_date": record.TradeDate,
		"job_id":     record.JobID,
	})
	warningsJSON := mustJSON([]string{"已有本地快照"})
	artifactsJSON := mustJSON(map[string]any{
		"report": map[string]any{
			"trade_date":       record.TradeDate,
			"selected_count":   record.SelectedCount,
			"publish_payloads": record.PublishPayloads,
			"candidates": []map[string]any{
				{"symbol": firstAssetKey(record.AssetKeys)},
			},
		},
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
		warningsJSON,
		artifactsJSON,
	)
}

func newPublishReplayRows(record strategyEnginePublishRecordDetail) *sqlmock.Rows {
	replayJSON := mustJSON(map[string]any{
		"warning_count":      record.Replay.WarningCount,
		"warning_messages":   record.Replay.WarningMessages,
		"vetoed_assets":      record.Replay.VetoedAssets,
		"invalidated_assets": record.Replay.InvalidatedAssets,
		"notes":              record.Replay.Notes,
	})
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
	}).AddRow(
		record.PublishID,
		record.JobID,
		record.Version,
		"",
		false,
		"",
		"",
		replayJSON,
		record.CreatedAt,
	)
}

func newPublishHistorySummaryRows(items ...model.StrategyEnginePublishRecordSummary) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		"publish_id",
		"job_id",
		"job_type",
		"publish_version",
		"created_at",
		"trade_date",
		"result_summary",
		"selected_count",
		"payload_count",
		"asset_keys",
	})
	for _, item := range items {
		rows.AddRow(
			item.PublishID,
			item.JobID,
			item.JobType,
			item.Version,
			item.CreatedAt,
			item.TradeDate,
			item.ReportSummary,
			item.SelectedCount,
			item.PayloadCount,
			mustJSON(item.AssetKeys),
		)
	}
	return rows
}

func buildRemotePublishRecordDetail(
	publishID string,
	jobID string,
	version int,
	tradeDate string,
	selectedCount int,
	payloadCount int,
	assetKeys []string,
	warningMessages []string,
	vetoedAssets []string,
) strategyEnginePublishRecordDetail {
	publishPayloads := make([]map[string]any, 0, len(assetKeys))
	for _, assetKey := range assetKeys {
		publishPayloads = append(publishPayloads, map[string]any{
			"recommendation": map[string]any{
				"symbol": assetKey,
			},
		})
	}
	reportSnapshot := map[string]any{
		"trade_date":       tradeDate,
		"publish_payloads": publishPayloads,
		"candidates": []map[string]any{
			{"symbol": firstAssetKey(assetKeys)},
		},
	}
	return strategyEnginePublishRecordDetail{
		PublishID:       publishID,
		JobID:           jobID,
		JobType:         "stock-selection",
		Version:         version,
		CreatedAt:       tradeDate + "T08:30:00Z",
		TradeDate:       tradeDate,
		ReportSummary:   "生成策略发布快照",
		SelectedCount:   selectedCount,
		AssetKeys:       assetKeys,
		PayloadCount:    payloadCount,
		Markdown:        "# 发布报告",
		HTML:            "<p>发布报告</p>",
		PublishPayloads: publishPayloads,
		ReportSnapshot:  reportSnapshot,
		Replay: strategyEnginePublishReplay{
			WarningCount:      len(warningMessages),
			WarningMessages:   warningMessages,
			VetoedAssets:      vetoedAssets,
			InvalidatedAssets: []string{},
			Notes:             []string{"来源于远端发布记录"},
		},
	}
}

func mustJSON(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(body)
}

func firstAssetKey(assetKeys []string) string {
	if len(assetKeys) == 0 {
		return ""
	}
	return assetKeys[0]
}
