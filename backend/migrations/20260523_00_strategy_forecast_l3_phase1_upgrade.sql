SET @headline_verdict_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'strategy_forecast_l3_reports'
    AND column_name = 'headline_verdict'
);

SET @alter_sql := IF(
  @headline_verdict_exists = 0,
  'ALTER TABLE strategy_forecast_l3_reports ADD COLUMN headline_verdict TEXT NULL AFTER version, ADD COLUMN state_assessment_json JSON NULL AFTER primary_scenario, ADD COLUMN dimension_evidence_json JSON NULL AFTER state_assessment_json, ADD COLUMN scenario_assessment_json JSON NULL AFTER dimension_evidence_json, ADD COLUMN validation_review_json JSON NULL AFTER scenario_assessment_json',
  'SELECT 1'
);

PREPARE alter_stmt FROM @alter_sql;
EXECUTE alter_stmt;
DEALLOCATE PREPARE alter_stmt;

