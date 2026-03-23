package repo

import (
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const strategyPublishPoliciesAllQueryPattern = `SELECT config_key, config_value, description, updated_by, updated_at\s+FROM system_configs\s+WHERE config_key LIKE \?`
const strategyScenarioTemplatesAllQueryPattern = `SELECT config_key, config_value, description, updated_by, updated_at\s+FROM system_configs\s+WHERE config_key LIKE \?`
const strategyPersistConfigExecPattern = `INSERT INTO system_configs \(id, config_key, config_value, description, updated_by, updated_at\)`

func TestResolveActiveStrategyScenarioTemplateMaterializesBuiltInFuturesTemplate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(strategyScenarioTemplatesAllQueryPattern).
		WithArgs(strategyScenarioTemplateConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyScenarioTemplateConfigPrefix+"scenario_default_futures",
			sqlmock.AnyArg(),
			"系统自动落库的期货默认场景模板，可在后台继续编辑与审计。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(strategyScenarioTemplatesAllQueryPattern).
		WithArgs(strategyScenarioTemplateConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}).
			AddRow(
				strategyScenarioTemplateConfigPrefix+"scenario_default_futures",
				`{"id":"scenario_default_futures","name":"期货六世界模板","target_type":"FUTURES","status":"ACTIVE","is_default":true,"items":[{"scenario":"base","label":"常态","thesis_template":"主线逻辑延续但尚未出现强加速，继续跟踪基差、库存与持仓确认。","action":"观察","risk_signal":"中","score_bias":0},{"scenario":"trend_continue","label":"趋势延续","thesis_template":"趋势、持仓与成交同步确认，顺势策略仍可延展。","action":"顺势跟进","risk_signal":"中低","score_bias":0},{"scenario":"policy_positive","label":"政策利多","thesis_template":"政策边际改善强化需求或风险偏好，优先验证资金是否跟随。","action":"择机加仓","risk_signal":"中","score_bias":0},{"scenario":"policy_negative","label":"政策压制","thesis_template":"监管或政策扰动压制预期，需要收缩仓位并重估胜率。","action":"降杠杆","risk_signal":"中高","score_bias":0},{"scenario":"supply_shock","label":"供给冲击","thesis_template":"供给侧扰动抬升波动，关注库存、升贴水和跨品种传导。","action":"事件驱动","risk_signal":"高","score_bias":0},{"scenario":"liquidity_shock","label":"流动性冲击","thesis_template":"流动性恶化或风险偏好骤降，优先控制回撤与滑点风险。","action":"防守回避","risk_signal":"高","score_bias":0}],"description":"系统自动落库的期货默认场景模板，可在后台继续编辑与审计。","updated_by":"system-bootstrap"}`,
				"系统自动落库的期货默认场景模板，可在后台继续编辑与审计。",
				"system-bootstrap",
				time.Date(2026, 3, 22, 9, 0, 0, 0, time.Local),
			))

	item, err := repo.ResolveActiveStrategyScenarioTemplate("FUTURES")
	if err != nil {
		t.Fatalf("ResolveActiveStrategyScenarioTemplate returned error: %v", err)
	}
	if item == nil {
		t.Fatal("expected materialized futures scenario template, got nil")
	}
	if item.TargetType != "FUTURES" || item.ID != "scenario_default_futures" || len(item.Items) != 6 {
		t.Fatalf("unexpected futures scenario template: %+v", item)
	}
	if item.UpdatedBy != strategyBuiltInPolicyOperator {
		t.Fatalf("expected system-bootstrap updater, got %+v", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdminListStrategyScenarioTemplatesReturnsMaterializedDefaults(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(strategyScenarioTemplatesAllQueryPattern).
		WithArgs(strategyScenarioTemplateConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyScenarioTemplateConfigPrefix+"scenario_default_stock",
			sqlmock.AnyArg(),
			"系统自动落库的股票默认场景模板，可在后台继续编辑与审计。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyScenarioTemplateConfigPrefix+"scenario_default_futures",
			sqlmock.AnyArg(),
			"系统自动落库的期货默认场景模板，可在后台继续编辑与审计。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(strategyScenarioTemplatesAllQueryPattern).
		WithArgs(strategyScenarioTemplateConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}).
			AddRow(
				strategyScenarioTemplateConfigPrefix+"scenario_default_futures",
				`{"id":"scenario_default_futures","name":"期货六世界模板","target_type":"FUTURES","status":"ACTIVE","is_default":true,"items":[{"scenario":"base","label":"常态","thesis_template":"主线逻辑延续但尚未出现强加速，继续跟踪基差、库存与持仓确认。","action":"观察","risk_signal":"中","score_bias":0},{"scenario":"trend_continue","label":"趋势延续","thesis_template":"趋势、持仓与成交同步确认，顺势策略仍可延展。","action":"顺势跟进","risk_signal":"中低","score_bias":0},{"scenario":"policy_positive","label":"政策利多","thesis_template":"政策边际改善强化需求或风险偏好，优先验证资金是否跟随。","action":"择机加仓","risk_signal":"中","score_bias":0},{"scenario":"policy_negative","label":"政策压制","thesis_template":"监管或政策扰动压制预期，需要收缩仓位并重估胜率。","action":"降杠杆","risk_signal":"中高","score_bias":0},{"scenario":"supply_shock","label":"供给冲击","thesis_template":"供给侧扰动抬升波动，关注库存、升贴水和跨品种传导。","action":"事件驱动","risk_signal":"高","score_bias":0},{"scenario":"liquidity_shock","label":"流动性冲击","thesis_template":"流动性恶化或风险偏好骤降，优先控制回撤与滑点风险。","action":"防守回避","risk_signal":"高","score_bias":0}],"description":"系统自动落库的期货默认场景模板，可在后台继续编辑与审计。","updated_by":"system-bootstrap"}`,
				"系统自动落库的期货默认场景模板，可在后台继续编辑与审计。",
				"system-bootstrap",
				time.Date(2026, 3, 22, 9, 0, 0, 0, time.Local),
			).
			AddRow(
				strategyScenarioTemplateConfigPrefix+"scenario_default_stock",
				`{"id":"scenario_default_stock","name":"股票四象限模板","target_type":"STOCK","status":"ACTIVE","is_default":true,"items":[{"scenario":"bull","label":"进攻","thesis_template":"景气扩散与资金跟随，趋势继续强化。","action":"加仓","risk_signal":"低","score_bias":0},{"scenario":"base","label":"常态","thesis_template":"维持当前节奏，等待下一轮验证。","action":"持有","risk_signal":"中","score_bias":0},{"scenario":"bear","label":"收缩","thesis_template":"市场回撤导致估值与情绪压缩。","action":"减仓","risk_signal":"中高","score_bias":0},{"scenario":"shock","label":"防守","thesis_template":"黑天鹅或流动性冲击下先保命。","action":"回避","risk_signal":"高","score_bias":0}],"description":"系统自动落库的股票默认场景模板，可在后台继续编辑与审计。","updated_by":"system-bootstrap"}`,
				"系统自动落库的股票默认场景模板，可在后台继续编辑与审计。",
				"system-bootstrap",
				time.Date(2026, 3, 22, 9, 0, 0, 0, time.Local),
			))

	items, total, err := repo.AdminListStrategyScenarioTemplates("", "ACTIVE", 1, 20)
	if err != nil {
		t.Fatalf("AdminListStrategyScenarioTemplates returned error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 materialized scenario templates, got total=%d items=%d", total, len(items))
	}
	seenTargets := map[string]bool{}
	for _, item := range items {
		seenTargets[item.TargetType] = true
	}
	if !seenTargets["STOCK"] || !seenTargets["FUTURES"] {
		t.Fatalf("expected stock and futures templates, got %+v", items)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestResolveActiveStrategyPublishPolicyMaterializesBuiltInFuturesPolicy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(strategyPublishPoliciesAllQueryPattern).
		WithArgs(strategyPublishPolicyConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyPublishPolicyConfigPrefix+"policy_default_futures",
			sqlmock.AnyArg(),
			"系统自动落库的期货默认发布策略，可在后台继续编辑与审计。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyPublishPolicyConfigPrefix+"policy_default_all",
			sqlmock.AnyArg(),
			"系统自动落库的默认发布策略，可在后台继续编辑与审计。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(strategyPublishPoliciesAllQueryPattern).
		WithArgs(strategyPublishPolicyConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}).
			AddRow(
				strategyPublishPolicyConfigPrefix+"policy_default_futures",
				`{"id":"policy_default_futures","name":"期货默认发布门槛","target_type":"FUTURES","status":"ACTIVE","is_default":true,"max_risk_level":"HIGH","max_warning_count":5,"allow_vetoed_publish":true,"default_publisher":"strategy-engine","override_note_template":"期货策略默认允许带 veto 与 warning 发布，但必须保留完整风控说明与复盘链路。","description":"系统自动落库的期货默认发布策略，可在后台继续编辑与审计。","updated_by":"system-bootstrap"}`,
				"系统自动落库的期货默认发布策略，可在后台继续编辑与审计。",
				"system-bootstrap",
				time.Date(2026, 3, 19, 10, 0, 0, 0, time.Local),
			).
			AddRow(
				strategyPublishPolicyConfigPrefix+"policy_default_all",
				`{"id":"policy_default_all","name":"默认发布门槛","target_type":"ALL","status":"ACTIVE","is_default":true,"max_risk_level":"MEDIUM","max_warning_count":3,"allow_vetoed_publish":false,"default_publisher":"strategy-engine","override_note_template":"人工覆盖发布，需记录原因与复盘结论。","description":"系统自动落库的默认发布策略，可在后台继续编辑与审计。","updated_by":"system-bootstrap"}`,
				"系统自动落库的默认发布策略，可在后台继续编辑与审计。",
				"system-bootstrap",
				time.Date(2026, 3, 19, 10, 0, 0, 0, time.Local),
			))

	item, err := repo.ResolveActiveStrategyPublishPolicy("FUTURES")
	if err != nil {
		t.Fatalf("ResolveActiveStrategyPublishPolicy returned error: %v", err)
	}
	if item == nil {
		t.Fatal("expected materialized futures policy, got nil")
	}
	if item.TargetType != "FUTURES" || item.MaxRiskLevel != "HIGH" || !item.AllowVetoedPublish {
		t.Fatalf("unexpected futures policy: %+v", item)
	}
	if item.UpdatedBy != strategyBuiltInPolicyOperator {
		t.Fatalf("expected system-bootstrap updater, got %+v", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdminListStrategyPublishPoliciesReturnsMaterializedDefaults(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(strategyPublishPoliciesAllQueryPattern).
		WithArgs(strategyPublishPolicyConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyPublishPolicyConfigPrefix+"policy_default_all",
			sqlmock.AnyArg(),
			"系统自动落库的默认发布策略，可在后台继续编辑与审计。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyPublishPolicyConfigPrefix+"policy_default_futures",
			sqlmock.AnyArg(),
			"系统自动落库的期货默认发布策略，可在后台继续编辑与审计。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(strategyPublishPoliciesAllQueryPattern).
		WithArgs(strategyPublishPolicyConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}).
			AddRow(
				strategyPublishPolicyConfigPrefix+"policy_default_futures",
				`{"id":"policy_default_futures","name":"期货默认发布门槛","target_type":"FUTURES","status":"ACTIVE","is_default":true,"max_risk_level":"HIGH","max_warning_count":5,"allow_vetoed_publish":true,"default_publisher":"strategy-engine","override_note_template":"期货策略默认允许带 veto 与 warning 发布，但必须保留完整风控说明与复盘链路。","description":"系统自动落库的期货默认发布策略，可在后台继续编辑与审计。","updated_by":"system-bootstrap"}`,
				"系统自动落库的期货默认发布策略，可在后台继续编辑与审计。",
				"system-bootstrap",
				time.Date(2026, 3, 19, 10, 0, 0, 0, time.Local),
			).
			AddRow(
				strategyPublishPolicyConfigPrefix+"policy_default_all",
				`{"id":"policy_default_all","name":"默认发布门槛","target_type":"ALL","status":"ACTIVE","is_default":true,"max_risk_level":"MEDIUM","max_warning_count":3,"allow_vetoed_publish":false,"default_publisher":"strategy-engine","override_note_template":"人工覆盖发布，需记录原因与复盘结论。","description":"系统自动落库的默认发布策略，可在后台继续编辑与审计。","updated_by":"system-bootstrap"}`,
				"系统自动落库的默认发布策略，可在后台继续编辑与审计。",
				"system-bootstrap",
				time.Date(2026, 3, 19, 10, 0, 0, 0, time.Local),
			))

	items, total, err := repo.AdminListStrategyPublishPolicies("", "ACTIVE", 1, 20)
	if err != nil {
		t.Fatalf("AdminListStrategyPublishPolicies returned error: %v", err)
	}
	if total != 2 || len(items) != 2 {
		t.Fatalf("expected 2 materialized policies, got total=%d items=%d", total, len(items))
	}
	if items[0].UpdatedBy != strategyBuiltInPolicyOperator && items[1].UpdatedBy != strategyBuiltInPolicyOperator {
		t.Fatalf("expected materialized policy updater, got %+v", items)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestResolveActiveStrategyPublishPolicyReturnsPersistedAllPolicyForStock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(strategyPublishPoliciesAllQueryPattern).
		WithArgs(strategyPublishPolicyConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyPublishPolicyConfigPrefix+"policy_default_all",
			sqlmock.AnyArg(),
			"系统自动落库的默认发布策略，可在后台继续编辑与审计。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(strategyPublishPoliciesAllQueryPattern).
		WithArgs(strategyPublishPolicyConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}).
			AddRow(
				strategyPublishPolicyConfigPrefix+"policy_default_all",
				`{"id":"policy_default_all","name":"默认发布门槛","target_type":"ALL","status":"ACTIVE","is_default":true,"max_risk_level":"MEDIUM","max_warning_count":3,"allow_vetoed_publish":false,"default_publisher":"strategy-engine","override_note_template":"人工覆盖发布，需记录原因与复盘结论。","description":"系统自动落库的默认发布策略，可在后台继续编辑与审计。","updated_by":"system-bootstrap"}`,
				"系统自动落库的默认发布策略，可在后台继续编辑与审计。",
				"system-bootstrap",
				time.Date(2026, 3, 19, 10, 0, 0, 0, time.Local),
			))

	item, err := repo.ResolveActiveStrategyPublishPolicy("STOCK")
	if err != nil {
		t.Fatalf("ResolveActiveStrategyPublishPolicy returned error: %v", err)
	}
	if item == nil || item.TargetType != "ALL" {
		t.Fatalf("expected persisted ALL policy, got %+v", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestResolveActiveStrategyPublishPolicyReturnsSQLNotFoundWhenConfigListFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(strategyPublishPoliciesAllQueryPattern).
		WithArgs(strategyPublishPolicyConfigPrefix + "%").
		WillReturnError(sql.ErrNoRows)

	_, err = repo.ResolveActiveStrategyPublishPolicy("FUTURES")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
