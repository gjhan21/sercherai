package repo

import (
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const strategyAgentProfilesAllQueryPattern = `SELECT config_key, config_value, description, updated_by, updated_at\s+FROM system_configs\s+WHERE config_key LIKE \?`

func TestResolveActiveStrategyAgentProfileMaterializesBuiltInFuturesProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(strategyAgentProfilesAllQueryPattern).
		WithArgs(strategyAgentProfileConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyAgentProfileConfigPrefix+"agent_default_futures",
			sqlmock.AnyArg(),
			"系统自动落库的期货默认角色配置，允许短历史时受控回退 MOCK。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyAgentProfileConfigPrefix+"agent_default_all",
			sqlmock.AnyArg(),
			"系统自动落库的默认多角色评审配置。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(strategyAgentProfilesAllQueryPattern).
		WithArgs(strategyAgentProfileConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}).
			AddRow(
				strategyAgentProfileConfigPrefix+"agent_default_futures",
				`{"id":"agent_default_futures","name":"期货默认五角评审","target_type":"FUTURES","status":"ACTIVE","is_default":true,"enabled_agents":["trend","event","liquidity","risk","basis"],"positive_threshold":3,"negative_threshold":2,"allow_veto":true,"allow_mock_fallback_on_short_history":true,"description":"系统自动落库的期货默认角色配置，允许短历史时受控回退 MOCK。","updated_by":"system-bootstrap"}`,
				"系统自动落库的期货默认角色配置，允许短历史时受控回退 MOCK。",
				"system-bootstrap",
				time.Date(2026, 3, 20, 9, 0, 0, 0, time.Local),
			).
			AddRow(
				strategyAgentProfileConfigPrefix+"agent_default_all",
				`{"id":"agent_default_all","name":"默认五角评审","target_type":"ALL","status":"ACTIVE","is_default":true,"enabled_agents":["trend","event","liquidity","risk","basis"],"positive_threshold":3,"negative_threshold":2,"allow_veto":true,"allow_mock_fallback_on_short_history":false,"description":"系统自动落库的默认多角色评审配置。","updated_by":"system-bootstrap"}`,
				"系统自动落库的默认多角色评审配置。",
				"system-bootstrap",
				time.Date(2026, 3, 20, 9, 0, 0, 0, time.Local),
			))

	item, err := repo.ResolveActiveStrategyAgentProfile("FUTURES")
	if err != nil {
		t.Fatalf("ResolveActiveStrategyAgentProfile returned error: %v", err)
	}
	if item == nil || item.TargetType != "FUTURES" {
		t.Fatalf("expected FUTURES profile, got %+v", item)
	}
	if !item.AllowMockFallbackOnShortHistory {
		t.Fatalf("expected built-in futures profile to enable short-history fallback, got %+v", item)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAdminListStrategyAgentProfilesReturnsMaterializedFuturesProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(strategyAgentProfilesAllQueryPattern).
		WithArgs(strategyAgentProfileConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyAgentProfileConfigPrefix+"agent_default_futures",
			sqlmock.AnyArg(),
			"系统自动落库的期货默认角色配置，允许短历史时受控回退 MOCK。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(strategyPersistConfigExecPattern).
		WithArgs(
			sqlmock.AnyArg(),
			strategyAgentProfileConfigPrefix+"agent_default_all",
			sqlmock.AnyArg(),
			"系统自动落库的默认多角色评审配置。",
			strategyBuiltInPolicyOperator,
			sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(strategyAgentProfilesAllQueryPattern).
		WithArgs(strategyAgentProfileConfigPrefix + "%").
		WillReturnRows(sqlmock.NewRows([]string{"config_key", "config_value", "description", "updated_by", "updated_at"}).
			AddRow(
				strategyAgentProfileConfigPrefix+"agent_default_futures",
				`{"id":"agent_default_futures","name":"期货默认五角评审","target_type":"FUTURES","status":"ACTIVE","is_default":true,"enabled_agents":["trend","event","liquidity","risk","basis"],"positive_threshold":3,"negative_threshold":2,"allow_veto":true,"allow_mock_fallback_on_short_history":true,"description":"系统自动落库的期货默认角色配置，允许短历史时受控回退 MOCK。","updated_by":"system-bootstrap"}`,
				"系统自动落库的期货默认角色配置，允许短历史时受控回退 MOCK。",
				"system-bootstrap",
				time.Date(2026, 3, 20, 9, 0, 0, 0, time.Local),
			).
			AddRow(
				strategyAgentProfileConfigPrefix+"agent_default_all",
				`{"id":"agent_default_all","name":"默认五角评审","target_type":"ALL","status":"ACTIVE","is_default":true,"enabled_agents":["trend","event","liquidity","risk","basis"],"positive_threshold":3,"negative_threshold":2,"allow_veto":true,"allow_mock_fallback_on_short_history":false,"description":"系统自动落库的默认多角色评审配置。","updated_by":"system-bootstrap"}`,
				"系统自动落库的默认多角色评审配置。",
				"system-bootstrap",
				time.Date(2026, 3, 20, 9, 0, 0, 0, time.Local),
			))

	items, total, err := repo.AdminListStrategyAgentProfiles("FUTURES", "ACTIVE", 1, 20)
	if err != nil {
		t.Fatalf("AdminListStrategyAgentProfiles returned error: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("expected one futures agent profile, got total=%d items=%d", total, len(items))
	}
	if !items[0].AllowMockFallbackOnShortHistory {
		t.Fatalf("expected materialized futures profile toggle, got %+v", items[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
