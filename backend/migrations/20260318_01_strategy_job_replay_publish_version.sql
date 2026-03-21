SET @strategy_job_replays_publish_version_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'strategy_job_replays'
    AND column_name = 'publish_version'
);

SET @strategy_job_replays_publish_version_sql := IF(
  @strategy_job_replays_publish_version_exists = 0,
  'ALTER TABLE strategy_job_replays ADD COLUMN publish_version int NOT NULL DEFAULT 0 AFTER publish_id',
  'SELECT 1'
);

PREPARE strategy_job_replays_publish_version_stmt FROM @strategy_job_replays_publish_version_sql;
EXECUTE strategy_job_replays_publish_version_stmt;
DEALLOCATE PREPARE strategy_job_replays_publish_version_stmt;
