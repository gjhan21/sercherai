SET @profiles_market_analysis_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_profiles'
    AND COLUMN_NAME = 'market_analysis_config'
);
SET @profiles_market_analysis_sql = IF(
  @profiles_market_analysis_exists = 0,
  'ALTER TABLE stock_selection_profiles ADD COLUMN market_analysis_config json NULL AFTER publish_config',
  'SELECT 1'
);
PREPARE stock_selection_profiles_market_analysis_stmt FROM @profiles_market_analysis_sql;
EXECUTE stock_selection_profiles_market_analysis_stmt;
DEALLOCATE PREPARE stock_selection_profiles_market_analysis_stmt;

SET @profiles_candidate_pool_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_profiles'
    AND COLUMN_NAME = 'candidate_pool_config'
);
SET @profiles_candidate_pool_sql = IF(
  @profiles_candidate_pool_exists = 0,
  'ALTER TABLE stock_selection_profiles ADD COLUMN candidate_pool_config json NULL AFTER market_analysis_config',
  'SELECT 1'
);
PREPARE stock_selection_profiles_candidate_pool_stmt FROM @profiles_candidate_pool_sql;
EXECUTE stock_selection_profiles_candidate_pool_stmt;
DEALLOCATE PREPARE stock_selection_profiles_candidate_pool_stmt;

SET @profiles_short_term_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_profiles'
    AND COLUMN_NAME = 'short_term_head_config'
);
SET @profiles_short_term_sql = IF(
  @profiles_short_term_exists = 0,
  'ALTER TABLE stock_selection_profiles ADD COLUMN short_term_head_config json NULL AFTER candidate_pool_config',
  'SELECT 1'
);
PREPARE stock_selection_profiles_short_term_stmt FROM @profiles_short_term_sql;
EXECUTE stock_selection_profiles_short_term_stmt;
DEALLOCATE PREPARE stock_selection_profiles_short_term_stmt;

SET @profiles_swing_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_profiles'
    AND COLUMN_NAME = 'swing_head_config'
);
SET @profiles_swing_sql = IF(
  @profiles_swing_exists = 0,
  'ALTER TABLE stock_selection_profiles ADD COLUMN swing_head_config json NULL AFTER short_term_head_config',
  'SELECT 1'
);
PREPARE stock_selection_profiles_swing_stmt FROM @profiles_swing_sql;
EXECUTE stock_selection_profiles_swing_stmt;
DEALLOCATE PREPARE stock_selection_profiles_swing_stmt;

SET @templates_market_analysis_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_profile_templates'
    AND COLUMN_NAME = 'market_analysis_defaults_json'
);
SET @templates_market_analysis_sql = IF(
  @templates_market_analysis_exists = 0,
  'ALTER TABLE stock_selection_profile_templates ADD COLUMN market_analysis_defaults_json json NULL AFTER publish_defaults_json',
  'SELECT 1'
);
PREPARE stock_selection_templates_market_analysis_stmt FROM @templates_market_analysis_sql;
EXECUTE stock_selection_templates_market_analysis_stmt;
DEALLOCATE PREPARE stock_selection_templates_market_analysis_stmt;

SET @templates_candidate_pool_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_profile_templates'
    AND COLUMN_NAME = 'candidate_pool_defaults_json'
);
SET @templates_candidate_pool_sql = IF(
  @templates_candidate_pool_exists = 0,
  'ALTER TABLE stock_selection_profile_templates ADD COLUMN candidate_pool_defaults_json json NULL AFTER market_analysis_defaults_json',
  'SELECT 1'
);
PREPARE stock_selection_templates_candidate_pool_stmt FROM @templates_candidate_pool_sql;
EXECUTE stock_selection_templates_candidate_pool_stmt;
DEALLOCATE PREPARE stock_selection_templates_candidate_pool_stmt;

SET @templates_short_term_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_profile_templates'
    AND COLUMN_NAME = 'short_term_head_defaults_json'
);
SET @templates_short_term_sql = IF(
  @templates_short_term_exists = 0,
  'ALTER TABLE stock_selection_profile_templates ADD COLUMN short_term_head_defaults_json json NULL AFTER candidate_pool_defaults_json',
  'SELECT 1'
);
PREPARE stock_selection_templates_short_term_stmt FROM @templates_short_term_sql;
EXECUTE stock_selection_templates_short_term_stmt;
DEALLOCATE PREPARE stock_selection_templates_short_term_stmt;

SET @templates_swing_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_profile_templates'
    AND COLUMN_NAME = 'swing_head_defaults_json'
);
SET @templates_swing_sql = IF(
  @templates_swing_exists = 0,
  'ALTER TABLE stock_selection_profile_templates ADD COLUMN swing_head_defaults_json json NULL AFTER short_term_head_defaults_json',
  'SELECT 1'
);
PREPARE stock_selection_templates_swing_stmt FROM @templates_swing_sql;
EXECUTE stock_selection_templates_swing_stmt;
DEALLOCATE PREPARE stock_selection_templates_swing_stmt;

SET @evidence_head_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_run_evidence'
    AND COLUMN_NAME = 'recommendation_head'
);
SET @evidence_head_sql = IF(
  @evidence_head_exists = 0,
  'ALTER TABLE stock_selection_run_evidence ADD COLUMN recommendation_head varchar(32) NULL AFTER risk_flags_json',
  'SELECT 1'
);
PREPARE stock_selection_evidence_head_stmt FROM @evidence_head_sql;
EXECUTE stock_selection_evidence_head_stmt;
DEALLOCATE PREPARE stock_selection_evidence_head_stmt;

SET @evidence_layer_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_run_evidence'
    AND COLUMN_NAME = 'selection_layer'
);
SET @evidence_layer_sql = IF(
  @evidence_layer_exists = 0,
  'ALTER TABLE stock_selection_run_evidence ADD COLUMN selection_layer varchar(32) NULL AFTER recommendation_head',
  'SELECT 1'
);
PREPARE stock_selection_evidence_layer_stmt FROM @evidence_layer_sql;
EXECUTE stock_selection_evidence_layer_stmt;
DEALLOCATE PREPARE stock_selection_evidence_layer_stmt;

SET @evidence_pattern_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_run_evidence'
    AND COLUMN_NAME = 'technical_pattern'
);
SET @evidence_pattern_sql = IF(
  @evidence_pattern_exists = 0,
  'ALTER TABLE stock_selection_run_evidence ADD COLUMN technical_pattern varchar(64) NULL AFTER selection_layer',
  'SELECT 1'
);
PREPARE stock_selection_evidence_pattern_stmt FROM @evidence_pattern_sql;
EXECUTE stock_selection_evidence_pattern_stmt;
DEALLOCATE PREPARE stock_selection_evidence_pattern_stmt;

SET @evaluations_head_label_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_run_evaluations'
    AND COLUMN_NAME = 'head_label'
);
SET @evaluations_head_label_sql = IF(
  @evaluations_head_label_exists = 0,
  'ALTER TABLE stock_selection_run_evaluations ADD COLUMN head_label varchar(64) NULL AFTER benchmark_symbol',
  'SELECT 1'
);
PREPARE stock_selection_evaluations_head_label_stmt FROM @evaluations_head_label_sql;
EXECUTE stock_selection_evaluations_head_label_stmt;
DEALLOCATE PREPARE stock_selection_evaluations_head_label_stmt;

SET @evaluations_scope_length = (
  SELECT COALESCE(CHARACTER_MAXIMUM_LENGTH, 0)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_run_evaluations'
    AND COLUMN_NAME = 'evaluation_scope'
  LIMIT 1
);
SET @evaluations_scope_sql = IF(
  @evaluations_scope_length > 0 AND @evaluations_scope_length < 32,
  'ALTER TABLE stock_selection_run_evaluations MODIFY COLUMN evaluation_scope varchar(32) NOT NULL DEFAULT ''PORTFOLIO''',
  'SELECT 1'
);
PREPARE stock_selection_evaluations_scope_stmt FROM @evaluations_scope_sql;
EXECUTE stock_selection_evaluations_scope_stmt;
DEALLOCATE PREPARE stock_selection_evaluations_scope_stmt;

SET @evaluations_holding_contract_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'stock_selection_run_evaluations'
    AND COLUMN_NAME = 'holding_contract'
);
SET @evaluations_holding_contract_sql = IF(
  @evaluations_holding_contract_exists = 0,
  'ALTER TABLE stock_selection_run_evaluations ADD COLUMN holding_contract varchar(32) NULL AFTER head_label',
  'SELECT 1'
);
PREPARE stock_selection_evaluations_holding_contract_stmt FROM @evaluations_holding_contract_sql;
EXECUTE stock_selection_evaluations_holding_contract_stmt;
DEALLOCATE PREPARE stock_selection_evaluations_holding_contract_stmt;
