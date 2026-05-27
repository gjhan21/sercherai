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
const forecastL3ReportByRunIDQueryPattern = `(?s)SELECT\s+id,\s*run_id,\s*version,\s*COALESCE\(headline_verdict, ''\),\s*COALESCE\(executive_summary, ''\),.*FROM strategy_forecast_l3_reports\s+WHERE run_id = \?\s+ORDER BY version DESC LIMIT 1`
const forecastL3LogsByRunIDQueryPattern = `(?s)SELECT\s+id,\s*run_id,\s*step_key,\s*status,\s*COALESCE\(message, ''\),.*FROM strategy_forecast_l3_logs\s+WHERE run_id = \?\s+ORDER BY created_at ASC, id ASC`
const forecastL3VIPUserQueryPattern = `SELECT member_level, kyc_status, vip_expire_at FROM users WHERE id = \?`

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
		Source:         "ADMIN_CONSOLE",
		SourceID:       "seed-admin",
		SourcePath:     "/admin/forecast-lab",
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
	if run.Source != "ADMIN_CONSOLE" || run.ContextQuality != model.StrategyForecastL3ContextQualityFull {
		t.Fatalf("expected run source and context quality to be populated, got %+v", run)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestPersistMySQLStrategyForecastL3ExecutionRefreshesReportIdentityOnRetry(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectExec(`(?s)INSERT INTO strategy_forecast_l3_reports.*ON DUPLICATE KEY UPDATE\s+id = VALUES\(id\),\s+created_at = VALUES\(created_at\),`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`(?s)UPDATE strategy_forecast_l3_runs`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	result := strategyForecastL3ExecutionResult{
		Run: model.StrategyForecastL3Run{
			ID:        "l3run_retry",
			EngineKey: model.StrategyForecastL3EngineLocalSynthesis,
			Status:    model.StrategyForecastL3StatusSucceeded,
			ReportRef: &model.StrategyForecastL3ReportRef{
				RunID:    "l3run_retry",
				ReportID: "l3report_new",
			},
		},
		Report: &model.StrategyForecastL3Report{
			ID:        "l3report_new",
			RunID:     "l3run_retry",
			Version:   1,
			CreatedAt: time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
			UpdatedAt: time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
	}

	if err := repo.persistMySQLStrategyForecastL3Execution(result, "system"); err != nil {
		t.Fatalf("persistMySQLStrategyForecastL3Execution() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestCreateStrategyForecastL3RunRejectsUserRequestWithoutTargetID(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	_, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:    model.StrategyForecastL3TargetTypeStock,
		TargetKey:     "600519.SH",
		TargetLabel:   "贵州茅台",
		TriggerType:   model.StrategyForecastL3TriggerTypeUserRequest,
		RequestUserID: "user_001",
		PriorityScore: 0.61,
		Reason:        "need deeper view",
		Source:        "RECOMMENDATION",
		SourceID:      "reco_001",
		SourcePath:    "/recommendations",
	})
	if err == nil {
		t.Fatalf("expected missing target_id to fail for strict user-request context")
	}
}

func TestListStrategyForecastL3HistoryForTargetReturnsOnlySucceededRuns(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	seedForecastL3HistoryRunsForTarget(t, repo, "600519.SH", model.StrategyForecastL3TargetTypeStock)

	items, err := repo.ListStrategyForecastL3HistoryForTarget("user_001", model.StrategyForecastL3TargetTypeStock, "600519.SH", 1, 20)
	if err != nil {
		t.Fatalf("ListStrategyForecastL3HistoryForTarget() error = %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected successful history items")
	}
	for _, item := range items {
		if item.Status != model.StrategyForecastL3StatusSucceeded {
			t.Fatalf("expected succeeded status in history item, got %+v", item)
		}
		if item.Review == nil || item.Review.ReviewGrade == "" {
			t.Fatalf("expected review grade in history item, got %+v", item)
		}
	}
}

func TestBuildStrategyForecastL3HistoryCompareDefaultsToLatestVsPrevious(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	seedForecastL3HistoryRunsForTarget(t, repo, "AU2408", model.StrategyForecastL3TargetTypeFutures)

	compare, err := repo.GetStrategyForecastL3HistoryCompare("user_001", model.StrategyForecastL3TargetTypeFutures, "AU2408", "", "")
	if err != nil {
		t.Fatalf("GetStrategyForecastL3HistoryCompare() error = %v", err)
	}
	if compare.LeftRun == nil || compare.RightRun == nil {
		t.Fatalf("expected latest vs previous pair, got %+v", compare)
	}
	if len(compare.VerdictShift) == 0 {
		t.Fatalf("expected verdict shift summary, got %+v", compare)
	}
	if len(compare.EvidenceDiffs) == 0 {
		t.Fatalf("expected role effectiveness evidence diffs, got %+v", compare)
	}
	if compare.EvidenceDiffs[0].Dimension == "STATE" || compare.EvidenceDiffs[0].Dimension == "SCENARIO" {
		t.Fatalf("expected evidence diff dimensions, got %+v", compare.EvidenceDiffs)
	}
}

func TestListStrategyForecastL3HistoryForTargetFiltersByRequestUserID(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	seedForecastL3HistoryRunsForTarget(t, repo, "RB2609", model.StrategyForecastL3TargetTypeFutures)

	items, err := repo.ListStrategyForecastL3HistoryForTarget("another_user", model.StrategyForecastL3TargetTypeFutures, "RB2609", 1, 20)
	if err != nil {
		t.Fatalf("ListStrategyForecastL3HistoryForTarget() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no history for another user, got %+v", items)
	}
}

func TestGetStrategyForecastL3RunReviewRejectsOtherUsersRun(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	seedForecastL3HistoryRunsForTarget(t, repo, "CU2407", model.StrategyForecastL3TargetTypeFutures)

	items, err := repo.ListStrategyForecastL3HistoryForTarget("user_001", model.StrategyForecastL3TargetTypeFutures, "CU2407", 1, 20)
	if err != nil || len(items) == 0 {
		t.Fatalf("expected seeded history items, err=%v len=%d", err, len(items))
	}

	if _, err := repo.GetStrategyForecastL3RunReview(items[0].RunID, "another_user"); err == nil {
		t.Fatalf("expected review access to be denied for another user")
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
			"headline_verdict",
			"executive_summary",
			"primary_scenario",
			"state_assessment_json",
			"dimension_evidence_json",
			"scenario_assessment_json",
			"validation_review_json",
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
			"螺纹主力：模型复核认为主情景与现有证据基本一致。",
			"高位分歧扩大，先看基差和库存。",
			"base",
			`{"current_state":"base","risk_boundary":"跌破关键支撑","source":"RECOMMENDATION","context_quality":"FULL"}`,
			`[{"dimension":"SUPPLY_DEMAND","stance":"CONSTRUCTIVE","confidence":0.82,"summary":"供需基本面验证暂未恶化。","supporting_points":["库存"],"risk_points":["基差快速恶化"]}]`,
			`{"current_state":"base","primary_scenario":"base","secondary_scenarios":["bull","bear"],"trigger_conditions":["库存拐点"],"invalidation_conditions":["跌破关键支撑"],"action_plan":["等库存确认"]}`,
			`{"verdict":"模型复核认为主情景与现有证据基本一致。","scenario_consistency":"主情景整体自洽。","supporting_evidence":["库存拐点"],"counter_evidence":["基差快速恶化"],"blind_spots":["缺少更长窗口验证"],"risk_review":["关注回撤"],"action_review":["等库存确认"],"llm_summary":"当前复核支持继续跟踪。","status":"COMPLETED"}`,
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

func TestGetStrategyForecastL3RunDetailForUserHidesFullReportForNonVIP(t *testing.T) {
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
			"STOCK",
			"reco_001",
			"600519.SH",
			"贵州茅台",
			"USER_REQUEST",
			"user_nonvip_001",
			"",
			"LOCAL_SYNTHESIS",
			"SUCCEEDED",
			0.86,
			"user deep forecast",
			"",
			`{"source":"client"}`,
			`{"run_id":"l3run_demo_001","status":"SUCCEEDED","engine_key":"LOCAL_SYNTHESIS","trigger_type":"USER_REQUEST","target_type":"STOCK","target_key":"600519.SH","target_label":"贵州茅台","executive_summary":"摘要仍然可读。","primary_scenario":"base","action_guidance":"先等确认信号","confidence_label":"MEDIUM","priority_score":0.86,"generated_at":"2026-03-29T11:00:00Z","report_available":true}`,
			`{"run_id":"l3run_demo_001","report_id":"l3report_demo_001","status":"SUCCEEDED","engine_key":"LOCAL_SYNTHESIS","generated_at":"2026-03-29T11:00:00Z","requires_vip":true,"full_readable":true}`,
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
			"headline_verdict",
			"executive_summary",
			"primary_scenario",
			"state_assessment_json",
			"dimension_evidence_json",
			"scenario_assessment_json",
			"validation_review_json",
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
			"贵州茅台：模型复核认为主情景与现有证据基本一致。",
			"摘要仍然可读。",
			"base",
			`{"current_state":"base","risk_boundary":"跌破关键支撑","source":"RECOMMENDATION","context_quality":"FULL"}`,
			`[{"dimension":"FUNDAMENTAL","stance":"BULLISH","confidence":0.78,"summary":"基本面仍稳健。","supporting_points":["量能"],"risk_points":["跌破关键支撑"]}]`,
			`{"current_state":"base","primary_scenario":"base","secondary_scenarios":["bull"],"trigger_conditions":["放量确认"],"invalidation_conditions":["跌破关键支撑"],"action_plan":["等量能确认"]}`,
			`{"verdict":"模型复核认为主情景与现有证据基本一致。","scenario_consistency":"主情景整体自洽。","supporting_evidence":["放量确认"],"counter_evidence":["跌破关键支撑"],"blind_spots":["缺少更长窗口验证"],"risk_review":["关注回撤"],"action_review":["等量能确认"],"llm_summary":"当前复核支持继续跟踪。","status":"COMPLETED"}`,
			`[{"name":"bull","probability":0.22,"thesis":"补涨延续","action":"跟随"}]`,
			`[{"label":"量能","status":"WATCH","note":"继续跟踪","trigger":"放量确认"}]`,
			`["跌破关键支撑"]`,
			`[{"role":"RISK","stance":"CAUTION","summary":"回撤风险放大","veto":false}]`,
			`["等量能确认"]`,
			"# 完整深推演正文",
			"<h1>完整深推演正文</h1>",
			`{"run_id":"l3run_demo_001","status":"SUCCEEDED","engine_key":"LOCAL_SYNTHESIS","trigger_type":"USER_REQUEST","target_type":"STOCK","target_key":"600519.SH","target_label":"贵州茅台","executive_summary":"摘要仍然可读。","primary_scenario":"base","action_guidance":"先等确认信号","confidence_label":"MEDIUM","priority_score":0.86,"generated_at":"2026-03-29T11:00:00Z","report_available":true}`,
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
	mock.ExpectQuery(forecastL3ConfigQueryPattern).
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value"}).
			AddRow("growth.forecast_l3.require_vip_for_full_report", "true"))
	mock.ExpectQuery(forecastL3VIPUserQueryPattern).
		WithArgs("user_nonvip_001").
		WillReturnRows(sqlmock.NewRows([]string{"member_level", "kyc_status", "vip_expire_at"}).
			AddRow("FREE", "PASSED", nil))

	detail, err := repo.GetStrategyForecastL3RunDetailForUser("l3run_demo_001", "user_nonvip_001")
	if err != nil {
		t.Fatalf("GetStrategyForecastL3RunDetailForUser() error = %v", err)
	}
	if detail.Report == nil {
		t.Fatalf("expected report snapshot in detail, got %+v", detail)
	}
	if detail.Report.ExecutiveSummary == "" {
		t.Fatalf("expected executive summary to remain readable, got %+v", detail.Report)
	}
	if detail.Report.MarkdownBody != "" || detail.Report.HTMLBody != "" {
		t.Fatalf("expected full report body to be stripped for non-vip user, got %+v", detail.Report)
	}
	if detail.Run.ReportRef == nil || detail.Run.ReportRef.FullReadable {
		t.Fatalf("expected report ref to mark full report unreadable, got %+v", detail.Run.ReportRef)
	}
	if len(detail.Logs) != 1 {
		t.Fatalf("expected logs to remain available, got %+v", detail.Logs)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestInMemoryStrategyForecastL3RunLifecycle(t *testing.T) {
	repo := NewInMemoryGrowthRepo()

	run, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:    model.StrategyForecastL3TargetTypeStock,
		TargetID:      "sr_001",
		TargetKey:     "000001.SZ",
		TargetLabel:   "平安银行",
		Source:        "recommendations",
		SourceID:      "sr_001",
		SourcePath:    "/recommendations",
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

func seedForecastL3HistoryRunsForTarget(t *testing.T, repo *InMemoryGrowthRepo, targetKey string, targetType string) {
	t.Helper()

	for index, scenario := range []string{"base", "bull", "bear"} {
		run, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
			TargetType:    targetType,
			TargetID:      targetKey,
			TargetKey:     targetKey,
			TargetLabel:   targetKey,
			Source:        "RECOMMENDATION",
			SourceID:      "seed-history",
			SourcePath:    "/recommendations",
			TriggerType:   model.StrategyForecastL3TriggerTypeUserRequest,
			RequestUserID: "user_001",
			PriorityScore: 0.6 + float64(index)*0.05,
			Reason:        "seed history",
		})
		if err != nil {
			t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
		}
		repo.mu.Lock()
		stored := repo.forecastL3Runs[run.ID]
		stored.Status = model.StrategyForecastL3StatusSucceeded
		stored.FinishedAt = time.Date(2026, 5, 24, 10+index, 0, 0, 0, time.UTC).Format(time.RFC3339)
		stored.CreatedAt = time.Date(2026, 5, 24, 9+index, 0, 0, 0, time.UTC).Format(time.RFC3339)
		stored.UpdatedAt = stored.FinishedAt
		stored.Summary.ExecutiveSummary = "历史结论"
		stored.Summary.PrimaryScenario = scenario
		stored.Summary.ActionGuidance = "观察确认"
		stored.Summary.ReportAvailable = true
		repo.forecastL3Runs[run.ID] = stored
		repo.forecastL3Reports[run.ID] = model.StrategyForecastL3Report{
			ID:               "report_" + run.ID,
			RunID:            run.ID,
			Version:          1,
			HeadlineVerdict:  "结论 " + scenario,
			ExecutiveSummary: "历史结论",
			PrimaryScenario:  scenario,
			StateAssessment: &model.StrategyForecastL3StateAssessment{
				CurrentState: "等待确认",
			},
			DimensionEvidence: []model.StrategyForecastL3DimensionEvidence{
				{
					Dimension:  "TECHNICAL",
					Stance:     "WATCH",
					Summary:    "技术面观察",
					Confidence: 0.62 + float64(index)*0.05,
				},
			},
			Summary: model.StrategyForecastL3Summary{
				RunID:            run.ID,
				Status:           model.StrategyForecastL3StatusSucceeded,
				TargetType:       targetType,
				TargetKey:        targetKey,
				TargetLabel:      targetKey,
				ExecutiveSummary: "历史结论",
				PrimaryScenario:  scenario,
				ActionGuidance:   "观察确认",
				ReportAvailable:  true,
			},
			CreatedAt: stored.CreatedAt,
			UpdatedAt: stored.UpdatedAt,
		}
		repo.forecastL3Learning[run.ID] = []model.StrategyForecastL3LearningRecord{
			{
				ID:                "learn_" + run.ID,
				RunID:             run.ID,
				TargetType:        targetType,
				TargetKey:         targetKey,
				ScenarioHit:       index != 2,
				TriggerHit:        index == 0,
				InvalidationEarly: index == 2,
				BiasLabel:         []string{"UNDERCONFIRMED", "UNCALIBRATED", "RISK_FIRST"}[index],
				RoleEffectiveness: map[string]float64{"TECHNICAL": 0.7},
				Summary:           "历史复盘摘要",
				CreatedAt:         stored.FinishedAt,
				UpdatedAt:         stored.FinishedAt,
			},
		}
		repo.mu.Unlock()
	}

	failedRun, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:    targetType,
		TargetID:      targetKey,
		TargetKey:     targetKey,
		TargetLabel:   targetKey,
		Source:        "RECOMMENDATION",
		SourceID:      "seed-history-failed",
		SourcePath:    "/recommendations",
		TriggerType:   model.StrategyForecastL3TriggerTypeUserRequest,
		RequestUserID: "user_001",
		PriorityScore: 0.4,
		Reason:        "seed failed history",
	})
	if err != nil {
		t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
	}
	repo.mu.Lock()
	failedStored := repo.forecastL3Runs[failedRun.ID]
	failedStored.Status = model.StrategyForecastL3StatusFailed
	failedStored.UpdatedAt = time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC).Format(time.RFC3339)
	repo.forecastL3Runs[failedRun.ID] = failedStored
	repo.mu.Unlock()
}
