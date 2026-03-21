CREATE TABLE IF NOT EXISTS stock_selection_profiles (
  id varchar(64) NOT NULL,
  name varchar(128) NOT NULL,
  status varchar(16) NOT NULL DEFAULT 'ACTIVE',
  is_default tinyint(1) NOT NULL DEFAULT 0,
  selection_mode_default varchar(16) NOT NULL DEFAULT 'AUTO',
  universe_scope varchar(32) NOT NULL DEFAULT 'CN_A_ALL',
  universe_config json DEFAULT NULL,
  seed_mining_config json DEFAULT NULL,
  factor_config json DEFAULT NULL,
  portfolio_config json DEFAULT NULL,
  publish_config json DEFAULT NULL,
  description varchar(255) DEFAULT NULL,
  updated_by varchar(64) NOT NULL,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_stock_selection_profiles_name (name),
  KEY idx_stock_selection_profiles_default (status, is_default),
  KEY idx_stock_selection_profiles_mode (selection_mode_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_selection_profile_versions (
  id varchar(64) NOT NULL,
  profile_id varchar(64) NOT NULL,
  version_no int NOT NULL,
  snapshot_json json NOT NULL,
  change_note varchar(255) DEFAULT NULL,
  created_by varchar(64) NOT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_stock_selection_profile_versions_profile_version (profile_id, version_no),
  CONSTRAINT fk_stock_selection_profile_versions_profile_id FOREIGN KEY (profile_id) REFERENCES stock_selection_profiles (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO stock_selection_profiles (
  id,
  name,
  status,
  is_default,
  selection_mode_default,
  universe_scope,
  universe_config,
  seed_mining_config,
  factor_config,
  portfolio_config,
  publish_config,
  description,
  updated_by,
  updated_at,
  created_at
) VALUES (
  'profile_default_stock_auto',
  '默认自动选股',
  'ACTIVE',
  1,
  'AUTO',
  'CN_A_ALL',
  '{"universe_scope":"CN_A_ALL","min_listing_days":180,"min_avg_turnover":50000000,"exclude_st":true,"exclude_suspended":true}',
  '{"mode":"AUTO","trend_enabled":true,"money_flow_enabled":true,"quality_enabled":true,"event_enabled":true,"linkage_enabled":true}',
  '{"model_version":"v1","groups":["trend","volatility","liquidity","money_flow","quality","event"]}',
  '{"limit":5,"max_risk_level":"MEDIUM","min_score":75}',
  '{"review_required":true,"allow_auto_publish":false}',
  '阶段9第一轮默认自动选股 profile。',
  'system-bootstrap',
  NOW(),
  NOW()
) ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  status = VALUES(status),
  is_default = VALUES(is_default),
  selection_mode_default = VALUES(selection_mode_default),
  universe_scope = VALUES(universe_scope),
  universe_config = VALUES(universe_config),
  seed_mining_config = VALUES(seed_mining_config),
  factor_config = VALUES(factor_config),
  portfolio_config = VALUES(portfolio_config),
  publish_config = VALUES(publish_config),
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO stock_selection_profile_versions (
  id,
  profile_id,
  version_no,
  snapshot_json,
  change_note,
  created_by,
  created_at
) VALUES (
  'profile_default_stock_auto_v1',
  'profile_default_stock_auto',
  1,
  '{
    "id":"profile_default_stock_auto",
    "name":"默认自动选股",
    "status":"ACTIVE",
    "is_default":true,
    "selection_mode_default":"AUTO",
    "universe_scope":"CN_A_ALL",
    "universe_config":{"universe_scope":"CN_A_ALL","min_listing_days":180,"min_avg_turnover":50000000,"exclude_st":true,"exclude_suspended":true},
    "seed_mining_config":{"mode":"AUTO","trend_enabled":true,"money_flow_enabled":true,"quality_enabled":true,"event_enabled":true,"linkage_enabled":true},
    "factor_config":{"model_version":"v1","groups":["trend","volatility","liquidity","money_flow","quality","event"]},
    "portfolio_config":{"limit":5,"max_risk_level":"MEDIUM","min_score":75},
    "publish_config":{"review_required":true,"allow_auto_publish":false}
  }',
  '系统初始化默认自动选股 profile。',
  'system-bootstrap',
  NOW()
) ON DUPLICATE KEY UPDATE
  snapshot_json = VALUES(snapshot_json),
  change_note = VALUES(change_note),
  created_by = VALUES(created_by),
  created_at = VALUES(created_at);
