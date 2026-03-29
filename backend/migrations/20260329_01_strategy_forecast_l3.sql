CREATE TABLE IF NOT EXISTS strategy_forecast_l3_runs (
  id varchar(64) NOT NULL,
  target_type varchar(16) NOT NULL,
  target_id varchar(64) DEFAULT NULL,
  target_key varchar(64) NOT NULL,
  target_label varchar(128) DEFAULT NULL,
  trigger_type varchar(32) NOT NULL,
  request_user_id varchar(64) DEFAULT NULL,
  operator_user_id varchar(64) DEFAULT NULL,
  engine_key varchar(64) NOT NULL DEFAULT 'LOCAL_SYNTHESIS',
  status varchar(16) NOT NULL DEFAULT 'QUEUED',
  priority_score decimal(8,4) NOT NULL DEFAULT 0,
  reason varchar(512) DEFAULT NULL,
  failure_reason text,
  context_meta_json json DEFAULT NULL,
  summary_json json DEFAULT NULL,
  report_ref_json json DEFAULT NULL,
  queued_at datetime DEFAULT NULL,
  started_at datetime DEFAULT NULL,
  finished_at datetime DEFAULT NULL,
  cancelled_at datetime DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_strategy_forecast_l3_runs_status (status, created_at),
  KEY idx_strategy_forecast_l3_runs_target (target_type, target_key, created_at),
  KEY idx_strategy_forecast_l3_runs_request_user (request_user_id, created_at),
  KEY idx_strategy_forecast_l3_runs_trigger (trigger_type, status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS strategy_forecast_l3_reports (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  version int NOT NULL DEFAULT 1,
  executive_summary text,
  primary_scenario varchar(64) DEFAULT NULL,
  alternative_scenarios_json json DEFAULT NULL,
  trigger_checklist_json json DEFAULT NULL,
  invalidation_signals_json json DEFAULT NULL,
  role_disagreements_json json DEFAULT NULL,
  action_guidance_json json DEFAULT NULL,
  markdown_body longtext,
  html_body longtext,
  summary_json json DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_strategy_forecast_l3_reports_run_version (run_id, version),
  KEY idx_strategy_forecast_l3_reports_run_id (run_id, updated_at),
  CONSTRAINT fk_strategy_forecast_l3_reports_run
    FOREIGN KEY (run_id) REFERENCES strategy_forecast_l3_runs(id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS strategy_forecast_l3_logs (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  step_key varchar(64) NOT NULL,
  status varchar(16) NOT NULL DEFAULT 'PENDING',
  message varchar(1024) DEFAULT NULL,
  payload_json json DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_strategy_forecast_l3_logs_run (run_id, created_at),
  CONSTRAINT fk_strategy_forecast_l3_logs_run
    FOREIGN KEY (run_id) REFERENCES strategy_forecast_l3_runs(id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS strategy_forecast_l3_learning_records (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  target_type varchar(16) NOT NULL,
  target_key varchar(64) NOT NULL,
  scenario_hit tinyint(1) NOT NULL DEFAULT 0,
  trigger_hit tinyint(1) NOT NULL DEFAULT 0,
  invalidation_early tinyint(1) NOT NULL DEFAULT 0,
  bias_label varchar(64) DEFAULT NULL,
  role_effectiveness_json json DEFAULT NULL,
  summary_text varchar(1024) DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_strategy_forecast_l3_learning_target (target_type, target_key, created_at),
  KEY idx_strategy_forecast_l3_learning_run (run_id, created_at),
  CONSTRAINT fk_strategy_forecast_l3_learning_run
    FOREIGN KEY (run_id) REFERENCES strategy_forecast_l3_runs(id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO scheduler_job_definitions
  (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_forecast_l3_dispatch_pending', 'forecast_l3_dispatch_pending', 'Forecast L3 Dispatch Pending', 'GROWTH', '0 */10 * * * *', 'DISABLED', NULL, 'system', NOW(), NOW()),
  ('jobdef_forecast_l3_quality_backfill', 'forecast_l3_quality_backfill', 'Forecast L3 Quality Backfill', 'GROWTH', '0 15 * * * *', 'DISABLED', NULL, 'system', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  module = VALUES(module),
  cron_expr = VALUES(cron_expr),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO rbac_permissions (code, name, module, action, description, status, created_at, updated_at)
VALUES
  ('forecast_l3.view', 'Forecast L3 View', 'FORECAST', 'VIEW', 'view forecast l3 runs, reports and quality summary', 'ACTIVE', NOW(), NOW()),
  ('forecast_l3.edit', 'Forecast L3 Edit', 'FORECAST', 'EDIT', 'create, retry and cancel forecast l3 runs', 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  module = VALUES(module),
  action = VALUES(action),
  description = VALUES(description),
  status = VALUES(status),
  updated_at = VALUES(updated_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_super_admin', p.code, NOW()
FROM rbac_permissions p
WHERE p.code IN ('forecast_l3.view', 'forecast_l3.edit')
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_ops_admin', p.code, NOW()
FROM rbac_permissions p
WHERE p.code IN ('forecast_l3.view', 'forecast_l3.edit')
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_auditor', p.code, NOW()
FROM rbac_permissions p
WHERE p.code IN ('forecast_l3.view')
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES
  ('cfg_growth_forecast_l3_enabled', 'growth.forecast_l3.enabled', 'false', 'enable forecast l3 runtime', 'system', NOW()),
  ('cfg_growth_forecast_l3_admin_manual_enabled', 'growth.forecast_l3.admin_manual_enabled', 'true', 'allow admin manual forecast l3 runs', 'system', NOW()),
  ('cfg_growth_forecast_l3_user_request_enabled', 'growth.forecast_l3.user_request_enabled', 'false', 'allow user-request forecast l3 runs', 'system', NOW()),
  ('cfg_growth_forecast_l3_auto_priority_enabled', 'growth.forecast_l3.auto_priority_enabled', 'false', 'allow auto-priority forecast l3 runs', 'system', NOW()),
  ('cfg_growth_forecast_l3_client_read_enabled', 'growth.forecast_l3.client_read_enabled', 'true', 'allow client read side forecast l3 summary', 'system', NOW()),
  ('cfg_growth_forecast_l3_require_vip_for_full_report', 'growth.forecast_l3.require_vip_for_full_report', 'true', 'require vip for full forecast l3 report', 'system', NOW()),
  ('cfg_growth_forecast_l3_max_active_runs', 'growth.forecast_l3.max_active_runs', '2', 'max concurrently active forecast l3 runs', 'system', NOW()),
  ('cfg_growth_forecast_l3_max_runs_per_day', 'growth.forecast_l3.max_runs_per_day', '24', 'max total forecast l3 runs per day', 'system', NOW()),
  ('cfg_growth_forecast_l3_max_user_runs_per_day', 'growth.forecast_l3.max_user_runs_per_day', '1', 'max user-request forecast l3 runs per day', 'system', NOW()),
  ('cfg_growth_forecast_l3_min_priority_threshold', 'growth.forecast_l3.min_priority_threshold', '0.70', 'min priority threshold for forecast l3 auto dispatch', 'system', NOW()),
  ('cfg_growth_forecast_l3_dispatch_enabled', 'growth.forecast_l3.dispatch.enabled', 'true', 'enable forecast l3 dispatch worker', 'system', NOW()),
  ('cfg_growth_forecast_l3_dispatch_interval_minutes', 'growth.forecast_l3.dispatch.interval_minutes', '5', 'dispatch worker interval minutes', 'system', NOW()),
  ('cfg_growth_forecast_l3_quality_enabled', 'growth.forecast_l3.quality.enabled', 'true', 'enable forecast l3 quality worker', 'system', NOW()),
  ('cfg_growth_forecast_l3_quality_interval_minutes', 'growth.forecast_l3.quality.interval_minutes', '60', 'quality worker interval minutes', 'system', NOW()),
  ('cfg_growth_forecast_l3_default_engine_key', 'growth.forecast_l3.default_engine_key', 'LOCAL_SYNTHESIS', 'default engine key for forecast l3', 'system', NOW())
ON DUPLICATE KEY UPDATE
  config_value = VALUES(config_value),
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);
