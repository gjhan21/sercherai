package repo

import (
	"database/sql"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const strategyPublishPoliciesAllQueryPattern = `SELECT config_key, config_value, description, updated_by, updated_at\s+FROM system_configs\s+WHERE config_key LIKE \?`
const strategyPersistConfigExecPattern = `INSERT INTO system_configs \(id, config_key, config_value, description, updated_by, updated_at\)`

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
