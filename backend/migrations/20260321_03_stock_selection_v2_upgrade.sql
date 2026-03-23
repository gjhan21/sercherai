SET @profiles_template_id_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_profiles'
    AND COLUMN_NAME = 'template_id'
);
SET @profiles_template_id_sql = IF(
  @profiles_template_id_exists = 0,
  'ALTER TABLE stock_selection_profiles ADD COLUMN template_id varchar(64) NULL AFTER name',
  'SELECT 1'
);
PREPARE stock_selection_profiles_template_id_stmt FROM @profiles_template_id_sql;
EXECUTE stock_selection_profiles_template_id_stmt;
DEALLOCATE PREPARE stock_selection_profiles_template_id_stmt;

SET @runs_template_id_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_runs'
    AND COLUMN_NAME = 'template_id'
);
SET @runs_template_id_sql = IF(
  @runs_template_id_exists = 0,
  'ALTER TABLE stock_selection_runs ADD COLUMN template_id varchar(64) NULL AFTER profile_version',
  'SELECT 1'
);
PREPARE stock_selection_runs_template_id_stmt FROM @runs_template_id_sql;
EXECUTE stock_selection_runs_template_id_stmt;
DEALLOCATE PREPARE stock_selection_runs_template_id_stmt;

SET @runs_market_regime_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_runs'
    AND COLUMN_NAME = 'market_regime'
);
SET @runs_market_regime_sql = IF(
  @runs_market_regime_exists = 0,
  'ALTER TABLE stock_selection_runs ADD COLUMN market_regime varchar(32) NULL AFTER template_id',
  'SELECT 1'
);
PREPARE stock_selection_runs_market_regime_stmt FROM @runs_market_regime_sql;
EXECUTE stock_selection_runs_market_regime_stmt;
DEALLOCATE PREPARE stock_selection_runs_market_regime_stmt;

SET @runs_template_snapshot_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_runs'
    AND COLUMN_NAME = 'template_snapshot'
);
SET @runs_template_snapshot_sql = IF(
  @runs_template_snapshot_exists = 0,
  'ALTER TABLE stock_selection_runs ADD COLUMN template_snapshot json NULL AFTER context_meta',
  'SELECT 1'
);
PREPARE stock_selection_runs_template_snapshot_stmt FROM @runs_template_snapshot_sql;
EXECUTE stock_selection_runs_template_snapshot_stmt;
DEALLOCATE PREPARE stock_selection_runs_template_snapshot_stmt;

SET @runs_compare_summary_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_runs'
    AND COLUMN_NAME = 'compare_summary'
);
SET @runs_compare_summary_sql = IF(
  @runs_compare_summary_exists = 0,
  'ALTER TABLE stock_selection_runs ADD COLUMN compare_summary json NULL AFTER template_snapshot',
  'SELECT 1'
);
PREPARE stock_selection_runs_compare_summary_stmt FROM @runs_compare_summary_sql;
EXECUTE stock_selection_runs_compare_summary_stmt;
DEALLOCATE PREPARE stock_selection_runs_compare_summary_stmt;

SET @reviews_published_snapshot_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_publish_reviews'
    AND COLUMN_NAME = 'published_portfolio_snapshot'
);
SET @reviews_published_snapshot_sql = IF(
  @reviews_published_snapshot_exists = 0,
  'ALTER TABLE stock_selection_publish_reviews ADD COLUMN published_portfolio_snapshot json NULL AFTER publish_version',
  'SELECT 1'
);
PREPARE stock_selection_reviews_published_snapshot_stmt FROM @reviews_published_snapshot_sql;
EXECUTE stock_selection_reviews_published_snapshot_stmt;
DEALLOCATE PREPARE stock_selection_reviews_published_snapshot_stmt;

CREATE TABLE IF NOT EXISTS stock_selection_profile_templates (
  id varchar(64) NOT NULL,
  template_key varchar(64) NOT NULL,
  name varchar(128) NOT NULL,
  description varchar(255) DEFAULT NULL,
  market_regime_bias varchar(32) DEFAULT NULL,
  is_default tinyint(1) NOT NULL DEFAULT 0,
  status varchar(16) NOT NULL DEFAULT 'ACTIVE',
  universe_defaults_json json DEFAULT NULL,
  seed_defaults_json json DEFAULT NULL,
  factor_defaults_json json DEFAULT NULL,
  portfolio_defaults_json json DEFAULT NULL,
  publish_defaults_json json DEFAULT NULL,
  updated_by varchar(64) DEFAULT NULL,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_stock_selection_profile_templates_key (template_key),
  KEY idx_stock_selection_profile_templates_default (status, is_default)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_selection_run_evidence (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  symbol varchar(32) NOT NULL,
  stage varchar(32) NOT NULL,
  name varchar(128) DEFAULT NULL,
  portfolio_role varchar(16) DEFAULT NULL,
  evidence_summary varchar(512) DEFAULT NULL,
  evidence_cards_json json DEFAULT NULL,
  positive_reasons_json json DEFAULT NULL,
  veto_reasons_json json DEFAULT NULL,
  theme_tags_json json DEFAULT NULL,
  sector_tags_json json DEFAULT NULL,
  risk_flags_json json DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_stock_selection_run_evidence_run_symbol_stage (run_id, symbol, stage),
  KEY idx_stock_selection_run_evidence_run_id (run_id),
  CONSTRAINT fk_stock_selection_run_evidence_run_id FOREIGN KEY (run_id) REFERENCES stock_selection_runs (run_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_selection_run_evaluations (
  id varchar(64) NOT NULL,
  run_id varchar(64) NOT NULL,
  symbol varchar(32) NOT NULL,
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
  UNIQUE KEY uk_stock_selection_run_evaluations_run_symbol_scope_horizon (run_id, symbol, evaluation_scope, horizon_day),
  KEY idx_stock_selection_run_evaluations_run_id (run_id),
  CONSTRAINT fk_stock_selection_run_evaluations_run_id FOREIGN KEY (run_id) REFERENCES stock_selection_runs (run_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO stock_selection_profile_templates (
  id,
  template_key,
  name,
  description,
  market_regime_bias,
  is_default,
  status,
  universe_defaults_json,
  seed_defaults_json,
  factor_defaults_json,
  portfolio_defaults_json,
  publish_defaults_json,
  updated_by,
  updated_at,
  created_at
) VALUES
  (
    'sstpl_trend_growth',
    'TREND_GROWTH',
    '趋势成长',
    '偏向中期趋势延续与成长弹性，适合强势市场。',
    'UPTREND',
    0,
    'ACTIVE',
    '{"universe_scope":"CN_A_ALL","min_listing_days":180,"min_avg_turnover":50000000,"exclude_st":true,"exclude_suspended":true,"price_min":8,"price_max":300,"volatility_max":8}',
    '{"bucket_limit":36,"seed_pool_cap":180,"trend_bias":1.3,"money_flow_bias":1.1,"quality_bias":0.8,"event_bias":0.9,"resonance_bias":1.2}',
    '{"lookback_days":120,"quant_weight":0.70,"event_weight":0.10,"resonance_weight":0.10,"liquidity_risk_weight":0.10}',
    '{"limit":5,"watchlist_limit":5,"max_risk_level":"MEDIUM","min_score":76,"max_symbol_per_bucket":2}',
    '{"review_required":true,"allow_auto_publish":false}',
    'system-bootstrap',
    NOW(),
    NOW()
  ),
  (
    'sstpl_leader_resonance',
    'LEADER_RESONANCE',
    '龙头共振',
    '偏向成交活跃、事件与趋势同步共振的龙头标的。',
    'ROTATION',
    0,
    'ACTIVE',
    '{"universe_scope":"CN_A_ALL","min_listing_days":120,"min_avg_turnover":80000000,"exclude_st":true,"exclude_suspended":true,"price_min":5,"price_max":500,"volatility_max":10}',
    '{"bucket_limit":36,"seed_pool_cap":180,"trend_bias":1.1,"money_flow_bias":1.2,"quality_bias":0.7,"event_bias":1.0,"resonance_bias":1.4}',
    '{"lookback_days":90,"quant_weight":0.68,"event_weight":0.12,"resonance_weight":0.12,"liquidity_risk_weight":0.08}',
    '{"limit":5,"watchlist_limit":5,"max_risk_level":"HIGH","min_score":74,"max_symbol_per_bucket":2}',
    '{"review_required":true,"allow_auto_publish":false}',
    'system-bootstrap',
    NOW(),
    NOW()
  ),
  (
    'sstpl_event_driven',
    'EVENT_DRIVEN',
    '事件驱动',
    '偏向资讯热度、事件确认和主题催化。',
    'EVENT_DRIVEN',
    0,
    'ACTIVE',
    '{"universe_scope":"CN_A_ALL","min_listing_days":90,"min_avg_turnover":40000000,"exclude_st":true,"exclude_suspended":true,"price_min":3,"price_max":200,"volatility_max":12}',
    '{"bucket_limit":36,"seed_pool_cap":180,"trend_bias":0.9,"money_flow_bias":1.0,"quality_bias":0.7,"event_bias":1.5,"resonance_bias":1.1}',
    '{"lookback_days":60,"quant_weight":0.62,"event_weight":0.18,"resonance_weight":0.12,"liquidity_risk_weight":0.08}',
    '{"limit":5,"watchlist_limit":5,"max_risk_level":"HIGH","min_score":72,"max_symbol_per_bucket":2}',
    '{"review_required":true,"allow_auto_publish":false}',
    'system-bootstrap',
    NOW(),
    NOW()
  ),
  (
    'sstpl_balanced_steady',
    'BALANCED_STEADY',
    '均衡稳健',
    '偏向平衡质量、趋势与风险约束，适合默认日常运行。',
    'DEFENSIVE',
    1,
    'ACTIVE',
    '{"universe_scope":"CN_A_ALL","min_listing_days":180,"min_avg_turnover":50000000,"exclude_st":true,"exclude_suspended":true,"price_min":5,"price_max":300,"volatility_max":8}',
    '{"bucket_limit":36,"seed_pool_cap":180,"trend_bias":1.0,"money_flow_bias":1.0,"quality_bias":1.0,"event_bias":1.0,"resonance_bias":1.0}',
    '{"lookback_days":120,"quant_weight":0.70,"event_weight":0.10,"resonance_weight":0.10,"liquidity_risk_weight":0.10}',
    '{"limit":5,"watchlist_limit":5,"max_risk_level":"MEDIUM","min_score":75,"max_symbol_per_bucket":2}',
    '{"review_required":true,"allow_auto_publish":false}',
    'system-bootstrap',
    NOW(),
    NOW()
  ),
  (
    'sstpl_sector_rotation',
    'SECTOR_ROTATION',
    '行业轮动',
    '偏向轮动市场里的行业切换和主题扩散。',
    'ROTATION',
    0,
    'ACTIVE',
    '{"universe_scope":"CN_A_ALL","min_listing_days":120,"min_avg_turnover":50000000,"exclude_st":true,"exclude_suspended":true,"price_min":4,"price_max":250,"volatility_max":10}',
    '{"bucket_limit":36,"seed_pool_cap":180,"trend_bias":1.0,"money_flow_bias":1.1,"quality_bias":0.9,"event_bias":0.9,"resonance_bias":1.2}',
    '{"lookback_days":90,"quant_weight":0.68,"event_weight":0.10,"resonance_weight":0.12,"liquidity_risk_weight":0.10}',
    '{"limit":5,"watchlist_limit":5,"max_risk_level":"MEDIUM","min_score":74,"max_symbol_per_bucket":2}',
    '{"review_required":true,"allow_auto_publish":false}',
    'system-bootstrap',
    NOW(),
    NOW()
  )
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  description = VALUES(description),
  market_regime_bias = VALUES(market_regime_bias),
  is_default = VALUES(is_default),
  status = VALUES(status),
  universe_defaults_json = VALUES(universe_defaults_json),
  seed_defaults_json = VALUES(seed_defaults_json),
  factor_defaults_json = VALUES(factor_defaults_json),
  portfolio_defaults_json = VALUES(portfolio_defaults_json),
  publish_defaults_json = VALUES(publish_defaults_json),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

UPDATE stock_selection_profiles
SET template_id = 'sstpl_balanced_steady'
WHERE id = 'profile_default_stock_auto' AND (template_id IS NULL OR template_id = '');
