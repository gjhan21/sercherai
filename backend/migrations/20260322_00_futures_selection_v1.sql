CREATE TABLE IF NOT EXISTS futures_selection_profile_templates (
  id varchar(64) NOT NULL,
  template_key varchar(64) NOT NULL,
  name varchar(128) NOT NULL,
  description varchar(255) DEFAULT NULL,
  market_regime_bias varchar(32) DEFAULT NULL,
  is_default tinyint(1) NOT NULL DEFAULT 0,
  status varchar(16) NOT NULL DEFAULT 'ACTIVE',
  universe_defaults_json json DEFAULT NULL,
  factor_defaults_json json DEFAULT NULL,
  portfolio_defaults_json json DEFAULT NULL,
  publish_defaults_json json DEFAULT NULL,
  updated_by varchar(64) DEFAULT NULL,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_futures_selection_profile_templates_key (template_key),
  KEY idx_futures_selection_profile_templates_default (status, is_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS futures_selection_profiles (
  id varchar(64) NOT NULL,
  name varchar(128) NOT NULL,
  template_id varchar(64) DEFAULT NULL,
  status varchar(16) NOT NULL DEFAULT 'ACTIVE',
  is_default tinyint(1) NOT NULL DEFAULT 0,
  style_default varchar(16) NOT NULL DEFAULT 'balanced',
  contract_scope varchar(32) NOT NULL DEFAULT 'DOMINANT_ALL',
  universe_config json DEFAULT NULL,
  factor_config json DEFAULT NULL,
  portfolio_config json DEFAULT NULL,
  publish_config json DEFAULT NULL,
  description varchar(255) DEFAULT NULL,
  updated_by varchar(64) NOT NULL,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_futures_selection_profiles_name (name),
  KEY idx_futures_selection_profiles_default (status, is_default),
  CONSTRAINT fk_futures_selection_profiles_template_id FOREIGN KEY (template_id) REFERENCES futures_selection_profile_templates (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS futures_selection_profile_versions (
  id varchar(64) NOT NULL,
  profile_id varchar(64) NOT NULL,
  version_no int NOT NULL,
  snapshot_json json NOT NULL,
  change_note varchar(255) DEFAULT NULL,
  created_by varchar(64) NOT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_futures_selection_profile_versions_profile_version (profile_id, version_no),
  CONSTRAINT fk_futures_selection_profile_versions_profile_id FOREIGN KEY (profile_id) REFERENCES futures_selection_profiles (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS futures_selection_runs (
  run_id varchar(64) NOT NULL,
  trade_date date NOT NULL,
  job_id varchar(64) DEFAULT NULL,
  profile_id varchar(64) NOT NULL,
  profile_version int NOT NULL DEFAULT 1,
  template_id varchar(64) DEFAULT NULL,
  market_regime varchar(32) DEFAULT NULL,
  style varchar(16) NOT NULL DEFAULT 'balanced',
  contract_scope varchar(32) NOT NULL DEFAULT 'DOMINANT_ALL',
  status varchar(32) NOT NULL DEFAULT 'QUEUED',
  result_summary text,
  warning_messages json DEFAULT NULL,
  warning_count int NOT NULL DEFAULT 0,
  universe_count int NOT NULL DEFAULT 0,
  candidate_count int NOT NULL DEFAULT 0,
  selected_count int NOT NULL DEFAULT 0,
  publish_count int NOT NULL DEFAULT 0,
  context_meta json DEFAULT NULL,
  template_snapshot json DEFAULT NULL,
  compare_summary json DEFAULT NULL,
  started_at datetime DEFAULT NULL,
  completed_at datetime DEFAULT NULL,
  created_by varchar(64) DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (run_id),
  UNIQUE KEY uk_futures_selection_runs_job_id (job_id),
  KEY idx_futures_selection_runs_trade_date (trade_date),
  KEY idx_futures_selection_runs_status (status, trade_date),
  KEY idx_futures_selection_runs_profile (profile_id, profile_version),
  CONSTRAINT fk_futures_selection_runs_profile_id FOREIGN KEY (profile_id) REFERENCES futures_selection_profiles (id),
  CONSTRAINT fk_futures_selection_runs_template_id FOREIGN KEY (template_id) REFERENCES futures_selection_profile_templates (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS futures_selection_run_stage_logs (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  stage_key varchar(32) NOT NULL,
  stage_order int NOT NULL DEFAULT 0,
  status varchar(32) NOT NULL DEFAULT 'PENDING',
  input_count int NOT NULL DEFAULT 0,
  output_count int NOT NULL DEFAULT 0,
  duration_ms bigint NOT NULL DEFAULT 0,
  detail_message text,
  payload_snapshot json DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_futures_selection_run_stage_logs_run_stage (run_id, stage_key),
  KEY idx_futures_selection_run_stage_logs_run_id (run_id),
  CONSTRAINT fk_futures_selection_run_stage_logs_run_id FOREIGN KEY (run_id) REFERENCES futures_selection_runs (run_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS futures_selection_run_candidates (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  contract varchar(32) NOT NULL,
  name varchar(128) NOT NULL,
  stage varchar(32) NOT NULL,
  score decimal(10,4) NOT NULL DEFAULT 0,
  direction varchar(16) NOT NULL DEFAULT 'NEUTRAL',
  risk_level varchar(16) NOT NULL DEFAULT 'MEDIUM',
  selected tinyint(1) NOT NULL DEFAULT 0,
  rank_no int NOT NULL DEFAULT 0,
  reason_summary varchar(512) DEFAULT NULL,
  evidence_summary varchar(512) DEFAULT NULL,
  portfolio_role varchar(16) DEFAULT NULL,
  risk_summary varchar(512) DEFAULT NULL,
  factor_breakdown_json json DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_futures_selection_run_candidates_run_contract_stage (run_id, contract, stage),
  KEY idx_futures_selection_run_candidates_run_stage (run_id, stage, rank_no),
  CONSTRAINT fk_futures_selection_run_candidates_run_id FOREIGN KEY (run_id) REFERENCES futures_selection_runs (run_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS futures_selection_run_portfolio (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  contract varchar(32) NOT NULL,
  name varchar(128) NOT NULL,
  rank_no int NOT NULL DEFAULT 0,
  score decimal(10,4) NOT NULL DEFAULT 0,
  direction varchar(16) NOT NULL DEFAULT 'NEUTRAL',
  risk_level varchar(16) NOT NULL DEFAULT 'MEDIUM',
  position_range varchar(64) DEFAULT NULL,
  reason_summary varchar(512) DEFAULT NULL,
  evidence_summary varchar(512) DEFAULT NULL,
  portfolio_role varchar(16) DEFAULT NULL,
  risk_summary varchar(512) DEFAULT NULL,
  factor_breakdown_json json DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_futures_selection_run_portfolio_run_contract (run_id, contract),
  KEY idx_futures_selection_run_portfolio_run_rank (run_id, rank_no),
  CONSTRAINT fk_futures_selection_run_portfolio_run_id FOREIGN KEY (run_id) REFERENCES futures_selection_runs (run_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS futures_selection_publish_reviews (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  review_status varchar(16) NOT NULL DEFAULT 'PENDING',
  reviewer varchar(64) DEFAULT NULL,
  review_note text,
  override_reason text,
  publish_id varchar(64) DEFAULT NULL,
  publish_version int NOT NULL DEFAULT 0,
  published_contract_snapshot json DEFAULT NULL,
  approved_at datetime DEFAULT NULL,
  rejected_at datetime DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_futures_selection_publish_reviews_run_id (run_id),
  KEY idx_futures_selection_publish_reviews_status (review_status, updated_at),
  CONSTRAINT fk_futures_selection_publish_reviews_run_id FOREIGN KEY (run_id) REFERENCES futures_selection_runs (run_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS futures_selection_run_evidence (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  contract varchar(32) NOT NULL,
  stage varchar(32) NOT NULL,
  name varchar(128) DEFAULT NULL,
  portfolio_role varchar(16) DEFAULT NULL,
  evidence_summary varchar(512) DEFAULT NULL,
  evidence_cards_json json DEFAULT NULL,
  positive_reasons_json json DEFAULT NULL,
  veto_reasons_json json DEFAULT NULL,
  risk_flags_json json DEFAULT NULL,
  related_entities_json json DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_futures_selection_run_evidence_run_contract_stage (run_id, contract, stage),
  KEY idx_futures_selection_run_evidence_run_id (run_id),
  CONSTRAINT fk_futures_selection_run_evidence_run_id FOREIGN KEY (run_id) REFERENCES futures_selection_runs (run_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS futures_selection_run_evaluations (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  contract varchar(32) NOT NULL,
  horizon_day int NOT NULL,
  evaluation_scope varchar(16) NOT NULL DEFAULT 'PORTFOLIO',
  name varchar(128) DEFAULT NULL,
  entry_date date DEFAULT NULL,
  exit_date date DEFAULT NULL,
  entry_price decimal(18,6) NOT NULL DEFAULT 0,
  exit_price decimal(18,6) NOT NULL DEFAULT 0,
  return_pct decimal(18,6) NOT NULL DEFAULT 0,
  excess_return_pct decimal(18,6) NOT NULL DEFAULT 0,
  max_drawdown_pct decimal(18,6) NOT NULL DEFAULT 0,
  hit_flag tinyint(1) NOT NULL DEFAULT 0,
  benchmark_symbol varchar(32) DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_futures_selection_run_evaluations_run_contract_scope_horizon (run_id, contract, evaluation_scope, horizon_day),
  KEY idx_futures_selection_run_evaluations_run_id (run_id),
  CONSTRAINT fk_futures_selection_run_evaluations_run_id FOREIGN KEY (run_id) REFERENCES futures_selection_runs (run_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO futures_selection_profile_templates (
  id,
  template_key,
  name,
  description,
  market_regime_bias,
  is_default,
  status,
  universe_defaults_json,
  factor_defaults_json,
  portfolio_defaults_json,
  publish_defaults_json,
  updated_by,
  updated_at,
  created_at
) VALUES (
  'fstpl_balanced_trend',
  'BALANCED_TREND',
  '均衡趋势',
  '默认期货研究模板，兼顾趋势、风险与流动性约束。',
  'BASE',
  1,
  'ACTIVE',
  '{"contract_scope":"DOMINANT_ALL","allow_mock_fallback_on_short_history":true,"style":"balanced"}',
  '{"min_confidence":55}',
  '{"limit":3,"max_risk_level":"HIGH"}',
  '{"review_required":true,"allow_auto_publish":false}',
  'system-bootstrap',
  NOW(),
  NOW()
) ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  description = VALUES(description),
  market_regime_bias = VALUES(market_regime_bias),
  is_default = VALUES(is_default),
  status = VALUES(status),
  universe_defaults_json = VALUES(universe_defaults_json),
  factor_defaults_json = VALUES(factor_defaults_json),
  portfolio_defaults_json = VALUES(portfolio_defaults_json),
  publish_defaults_json = VALUES(publish_defaults_json),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO futures_selection_profiles (
  id,
  name,
  template_id,
  status,
  is_default,
  style_default,
  contract_scope,
  universe_config,
  factor_config,
  portfolio_config,
  publish_config,
  description,
  updated_by,
  updated_at,
  created_at
) VALUES (
  'profile_default_futures_auto',
  '默认智能期货',
  'fstpl_balanced_trend',
  'ACTIVE',
  1,
  'balanced',
  'DOMINANT_ALL',
  '{"contract_scope":"DOMINANT_ALL","allow_mock_fallback_on_short_history":true}',
  '{"min_confidence":55}',
  '{"limit":3,"max_risk_level":"HIGH"}',
  '{"review_required":true,"allow_auto_publish":false}',
  '双资产研究平台默认期货 profile。',
  'system-bootstrap',
  NOW(),
  NOW()
) ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  template_id = VALUES(template_id),
  status = VALUES(status),
  is_default = VALUES(is_default),
  style_default = VALUES(style_default),
  contract_scope = VALUES(contract_scope),
  universe_config = VALUES(universe_config),
  factor_config = VALUES(factor_config),
  portfolio_config = VALUES(portfolio_config),
  publish_config = VALUES(publish_config),
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO futures_selection_profile_versions (
  id,
  profile_id,
  version_no,
  snapshot_json,
  change_note,
  created_by,
  created_at
) VALUES (
  'profile_default_futures_auto_v1',
  'profile_default_futures_auto',
  1,
  '{
    "id":"profile_default_futures_auto",
    "name":"默认智能期货",
    "template_id":"fstpl_balanced_trend",
    "status":"ACTIVE",
    "is_default":true,
    "style_default":"balanced",
    "contract_scope":"DOMINANT_ALL",
    "universe_config":{"contract_scope":"DOMINANT_ALL","allow_mock_fallback_on_short_history":true},
    "factor_config":{"min_confidence":55},
    "portfolio_config":{"limit":3,"max_risk_level":"HIGH"},
    "publish_config":{"review_required":true,"allow_auto_publish":false}
  }',
  '系统初始化默认智能期货 profile。',
  'system-bootstrap',
  NOW()
) ON DUPLICATE KEY UPDATE
  snapshot_json = VALUES(snapshot_json),
  change_note = VALUES(change_note),
  created_by = VALUES(created_by),
  created_at = VALUES(created_at);

INSERT INTO rbac_permissions (code, name, module, action, description, status, created_at, updated_at)
VALUES
  ('futures_selection.view', 'Futures Selection View', 'FUTURES_SELECTION', 'VIEW', 'view futures selection module', 'ACTIVE', NOW(), NOW()),
  ('futures_selection.manage', 'Futures Selection Manage', 'FUTURES_SELECTION', 'MANAGE', 'manage futures selection runs', 'ACTIVE', NOW(), NOW())
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
WHERE p.code IN ('futures_selection.view', 'futures_selection.manage')
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_ops_admin', p.code, NOW()
FROM rbac_permissions p
WHERE p.code IN ('futures_selection.view', 'futures_selection.manage')
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);

INSERT INTO rbac_role_permissions (role_id, permission_code, created_at)
SELECT 'role_auditor', p.code, NOW()
FROM rbac_permissions p
WHERE p.code IN ('futures_selection.view')
ON DUPLICATE KEY UPDATE created_at = VALUES(created_at);
