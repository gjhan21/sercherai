SET @payload_echo_snapshot_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'strategy_job_artifacts'
    AND column_name = 'payload_echo_snapshot'
);

SET @payload_echo_snapshot_sql := IF(
  @payload_echo_snapshot_exists = 0,
  'ALTER TABLE strategy_job_artifacts ADD COLUMN payload_echo_snapshot json DEFAULT NULL AFTER result_summary',
  'SELECT 1'
);

PREPARE payload_echo_snapshot_stmt FROM @payload_echo_snapshot_sql;
EXECUTE payload_echo_snapshot_stmt;
DEALLOCATE PREPARE payload_echo_snapshot_stmt;
