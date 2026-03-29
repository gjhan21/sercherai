package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"sercherai/backend/internal/growth/model"
)

const forecastL3RunInsertPattern = `INSERT INTO strategy_forecast_l3_runs`
const forecastL3RunCountQueryPattern = `SELECT COUNT\(\*\) FROM strategy_forecast_l3_runs`
const forecastL3ConfigQueryPattern = `(?s)SELECT\s+config_key,\s*config_value\s+FROM system_configs\s+WHERE config_key LIKE 'growth\.forecast_l3\.%'`
const forecastL3ActiveCountQueryPattern = `SELECT COUNT\(\*\) FROM strategy_forecast_l3_runs WHERE status IN \(\?,\?\)`
const forecastL3TodayCountQueryPattern = `SELECT COUNT\(\*\) FROM strategy_forecast_l3_runs WHERE DATE\(created_at\) = CURDATE\(\)`
const forecastL3RunListQueryPattern = `(?s)SELECT\s+id,\s*target_type,\s*COALESCE\(target_id, ''\),\s*target_key,`
const forecastL3RunByIDQueryPattern = `(?s)SELECT\s+id,\s*target_type,\s*COALESCE\(target_id, ''\),\s*target_key,.*FROM strategy_forecast_l3_runs\s+WHERE id = \?`
const forecastL3ReportByRunIDQueryPattern = `(?s)SELECT\s+id,\s*run_id,\s*version,\s*COALESCE\(executive_summary, ''\),.*FROM strategy_forecast_l3_reports\s+WHERE run_id = \?\s+ORDER BY version DESC LIMIT 1`
const forecastL3LogsByRunIDQueryPattern = `(?s)SELECT\s+id,\s*run_id,\s*step_key,\s*status,\s*COALESCE\(message, ''\),.*FROM strategy_forecast_l3_logs\s+WHERE run_id = \?\s+ORDER BY created_at ASC, id ASC`

func expectForecastL3CreateGuards(mock sqlmock.Sqlmock) {
	mock.ExpectQuery(forecastL3ConfigQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value"}).
			AddRow("growth.forecast_l3.enabled", "true").
			AddRow("growth.forecast_l3.admin_manual_enabled", "true").
			AddRow("growth.forecast_l3.user_request_enabled", "true").
			AddRow("growth.forecast_l3.max_active_runs", "4").
			AddRow("growth.forecast_l3.max_runs_per_day", "24").
			AddRow("growth.forecast_l3.max_user_runs_per_day", "2").
			AddRow("growth.forecast_l3.default_engine_key", "LOCAL_SYNTHESIS"))
	mock.ExpectQuery(forecastL3ActiveCountQueryPattern).
		WithArgs(model.StrategyForecastL3StatusQueued, model.StrategyForecastL3StatusRunning).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery(forecastL3TodayCountQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
}

func TestCreateStrategyForecastL3RunPersistsQueuedRecord(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	expectForecastL3CreateGuards(mock)
	mock.ExpectExec(forecastL3RunInsertPattern).
		WillReturnResult(sqlmock.NewResult(1, 1))

	run, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:     model.StrategyForecastL3TargetTypeStock,
		TargetID:       "reco_001",
		TargetKey:      "600519.SH",
		TargetLabel:    "贵州茅台",
		TriggerType:    model.StrategyForecastL3TriggerTypeAdminManual,
		RequestUserID:  "admin_001",
		OperatorUserID: "admin_001",
		PriorityScore:  0.82,
		Reason:         "manual deep forecast",
		ContextMeta:    map[string]any{"source": "admin"},
	})
	if err != nil {
		t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
	}
	if run.Status != model.StrategyForecastL3StatusQueued {
		t.Fatalf("expected queued run, got %+v", run)
	}
	if run.EngineKey != model.StrategyForecastL3EngineLocalSynthesis {
		t.Fatalf("expected local synthesis engine, got %+v", run)
	}
	if run.TargetKey != "600519.SH" || run.TriggerType != model.StrategyForecastL3TriggerTypeAdminManual {
		t.Fatalf("expected persisted run to echo target and trigger, got %+v", run)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListStrategyForecastL3RunsBuildsSummaryAndReportRef(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(forecastL3RunCountQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(forecastL3RunListQueryPattern).
		WithArgs(20, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"target_type",
			"target_id",
			"target_key",
			"target_label",
			"trigger_type",
			"request_user_id",
			"operator_user_id",
			"engine_key",
			"status",
			"priority_score",
			"reason",
			"failure_reason",
			"context_meta_json",
			"summary_json",
			"report_ref_json",
			"queued_at",
			"started_at",
			"finished_at",
			"cancelled_at",
			"created_at",
			"updated_at",
		}).AddRow(
			"l3run_demo_001",
			"STOCK",
			"reco_001",
			"600519.SH",
			"贵州茅台",
			"ADMIN_MANUAL",
			"admin_001",
			"admin_001",
			"LOCAL_SYNTHESIS",
			"SUCCEEDED",
			0.88,
			"manual deep forecast",
			"",
			`{"source":"admin"}`,
			`{"run_id":"l3run_demo_001","status":"SUCCEEDED","engine_key":"LOCAL_SYNTHESIS","trigger_type":"ADMIN_MANUAL","target_type":"STOCK","target_key":"600519.SH","target_label":"贵州茅台","executive_summary":"趋势延续，但需要确认量能。","primary_scenario":"base","action_guidance":"先看确认再加仓","confidence_label":"MEDIUM","priority_score":0.88,"generated_at":"2026-03-29T10:00:00Z","report_available":true}`,
			`{"run_id":"l3run_demo_001","report_id":"l3report_demo_001","status":"SUCCEEDED","engine_key":"LOCAL_SYNTHESIS","generated_at":"2026-03-29T10:00:00Z","requires_vip":true,"full_readable":false}`,
			time.Date(2026, 3, 29, 9, 30, 0, 0, time.UTC),
			time.Date(2026, 3, 29, 9, 31, 0, 0, time.UTC),
			time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC),
			nil,
			time.Date(2026, 3, 29, 9, 30, 0, 0, time.UTC),
			time.Date(2026, 3, 29, 10, 0, 0, 0, time.UTC),
		))

	items, total, err := repo.ListStrategyForecastL3Runs("", "", "", "", 1, 20)
	if err != nil {
		t.Fatalf("ListStrategyForecastL3Runs() error = %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected single run row, got total=%d items=%d", total, len(items))
	}
	if items[0].Summary.ExecutiveSummary == "" || items[0].ReportRef == nil {
		t.Fatalf("expected summary and report ref to be materialized, got %+v", items[0])
	}
	if items[0].ReportRef.ReportID != "l3report_demo_001" {
		t.Fatalf("expected report ref to carry report id, got %+v", items[0].ReportRef)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGetStrategyForecastL3RunDetailAggregatesReportAndLogs(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(forecastL3RunByIDQueryPattern).
		WithArgs("l3run_demo_001").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"target_type",
			"target_id",
			"target_key",
			"target_label",
			"trigger_type",
			"request_user_id",
			"operator_user_id",
			"engine_key",
			"status",
			"priority_score",
			"reason",
			"failure_reason",
			"context_meta_json",
			"summary_json",
			"report_ref_json",
			"queued_at",
			"started_at",
			"finished_at",
			"cancelled_at",
			"created_at",
			"updated_at",
		}).AddRow(
			"l3run_demo_001",
			"FUTURES",
			"futures_001",
			"RB2609",
			"螺纹主力",
			"USER_REQUEST",
			"user_001",
			"",
			"LOCAL_SYNTHESIS",
			"SUCCEEDED",
			0.76,
			"user deep forecast",
			"",
			`{"source":"client"}`,
			`{"run_id":"l3run_demo_001","status":"SUCCEEDED","engine_key":"LOCAL_SYNTHESIS","trigger_type":"USER_REQUEST","target_type":"FUTURES","target_key":"RB2609","target_label":"螺纹主力","executive_summary":"高位分歧扩大，先看基差和库存。","primary_scenario":"base","action_guidance":"观察主情景确认","confidence_label":"MEDIUM","priority_score":0.76,"generated_at":"2026-03-29T11:00:00Z","report_available":true}`,
			`{"run_id":"l3run_demo_001","report_id":"l3report_demo_001","status":"SUCCEEDED","engine_key":"LOCAL_SYNTHESIS","generated_at":"2026-03-29T11:00:00Z","requires_vip":false,"full_readable":true}`,
			time.Date(2026, 3, 29, 10, 30, 0, 0, time.UTC),
			time.Date(2026, 3, 29, 10, 31, 0, 0, time.UTC),
			time.Date(2026, 3, 29, 11, 0, 0, 0, time.UTC),
			nil,
			time.Date(2026, 3, 29, 10, 30, 0, 0, time.UTC),
			time.Date(2026, 3, 29, 11, 0, 0, 0, time.UTC),
		))
	mock.ExpectQuery(forecastL3ReportByRunIDQueryPattern).
		WithArgs("l3run_demo_001").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"run_id",
			"version",
			"executive_summary",
			"primary_scenario",
			"alternative_scenarios_json",
			"trigger_checklist_json",
			"invalidation_signals_json",
			"role_disagreements_json",
			"action_guidance_json",
			"markdown_body",
			"html_body",
			"summary_json",
			"created_at",
			"updated_at",
		}).AddRow(
			"l3report_demo_001",
			"l3run_demo_001",
			1,
			"高位分歧扩大，先看基差和库存。",
			"base",
			`[{"name":"bull","probability":0.22,"thesis":"补涨延续","action":"跟随"},{"name":"bear","probability":0.18,"thesis":"高位回撤","action":"收缩"}]`,
			`[{"label":"库存","status":"WATCH","note":"继续跟踪","trigger":"库存拐点"}]`,
			`["跌破关键支撑","基差快速恶化"]`,
			`[{"role":"RISK","stance":"CAUTION","summary":"回撤风险放大","veto":false}]`,
			`["等库存确认","缩短验证周期"]`,
			"# Deep Forecast",
			"<h1>Deep Forecast</h1>",
			`{"run_id":"l3run_demo_001","status":"SUCCEEDED","engine_key":"LOCAL_SYNTHESIS","trigger_type":"USER_REQUEST","target_type":"FUTURES","target_key":"RB2609","target_label":"螺纹主力","executive_summary":"高位分歧扩大，先看基差和库存。","primary_scenario":"base","action_guidance":"观察主情景确认","confidence_label":"MEDIUM","priority_score":0.76,"generated_at":"2026-03-29T11:00:00Z","report_available":true}`,
			time.Date(2026, 3, 29, 11, 0, 0, 0, time.UTC),
			time.Date(2026, 3, 29, 11, 0, 0, 0, time.UTC),
		))
	mock.ExpectQuery(forecastL3LogsByRunIDQueryPattern).
		WithArgs("l3run_demo_001").
		WillReturnRows(sqlmock.NewRows([]string{
			"id",
			"run_id",
			"step_key",
			"status",
			"message",
			"payload_json",
			"created_at",
		}).AddRow(
			"log_demo_001",
			"l3run_demo_001",
			"BUILD_RESEARCH_PACK",
			"SUCCESS",
			"context ready",
			`{"sources":4}`,
			time.Date(2026, 3, 29, 10, 32, 0, 0, time.UTC),
		))

	detail, err := repo.GetStrategyForecastL3RunDetail("l3run_demo_001")
	if err != nil {
		t.Fatalf("GetStrategyForecastL3RunDetail() error = %v", err)
	}
	if detail.Run.ID != "l3run_demo_001" {
		t.Fatalf("expected run detail to include run, got %+v", detail)
	}
	if detail.Report == nil || detail.Report.MarkdownBody == "" {
		t.Fatalf("expected report snapshot in detail, got %+v", detail)
	}
	if len(detail.Logs) != 1 || detail.Logs[0].StepKey != "BUILD_RESEARCH_PACK" {
		t.Fatalf("expected detail logs to be loaded, got %+v", detail.Logs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestInMemoryStrategyForecastL3RunLifecycle(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	run, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:    model.StrategyForecastL3TargetTypeStock,
		TargetKey:     "000001.SZ",
		TargetLabel:   "平安银行",
		TriggerType:   model.StrategyForecastL3TriggerTypeUserRequest,
		RequestUserID: "user_001",
		PriorityScore: 0.61,
		Reason:        "need deeper view",
	})
	if err != nil {
		t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
	}
	if run.ID == "" || run.Status != model.StrategyForecastL3StatusQueued {
		t.Fatalf("expected created in-memory run to be queued with id, got %+v", run)
	}

	items, total, err := repo.ListStrategyForecastL3Runs("user_001", "", "", "", 1, 20)
	if err != nil {
		t.Fatalf("ListStrategyForecastL3Runs() error = %v", err)
	}
	if total < 1 || len(items) < 1 {
		t.Fatalf("expected in-memory list to include created run, got total=%d items=%d", total, len(items))
	}

	cancelled, err := repo.CancelStrategyForecastL3Run(run.ID, "admin_001", "manual cancel")
	if err != nil {
		t.Fatalf("CancelStrategyForecastL3Run() error = %v", err)
	}
	if cancelled.Status != model.StrategyForecastL3StatusCancelled {
		t.Fatalf("expected cancelled in-memory run, got %+v", cancelled)
	}

	detail, err := repo.GetStrategyForecastL3RunDetail(run.ID)
	if err != nil {
		t.Fatalf("GetStrategyForecastL3RunDetail() error = %v", err)
	}
	if detail.Run.ID != run.ID {
		t.Fatalf("expected detail to carry run id, got %+v", detail)
	}
}
