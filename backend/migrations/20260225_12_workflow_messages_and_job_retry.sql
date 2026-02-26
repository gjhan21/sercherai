-- Workflow messages and scheduler retry support for MySQL 8.x

CREATE TABLE IF NOT EXISTS workflow_messages (
  id          varchar(64) PRIMARY KEY,
  review_id   varchar(64),
  target_id   varchar(64) NOT NULL,
  module      varchar(32) NOT NULL,
  receiver_id varchar(32),
  sender_id   varchar(32),
  event_type  varchar(32) NOT NULL,
  title       varchar(128) NOT NULL,
  content     varchar(512) NOT NULL,
  is_read     tinyint(1) NOT NULL DEFAULT 0,
  created_at  datetime NOT NULL,
  read_at     datetime,
  INDEX idx_receiver_read_created (receiver_id, is_read, created_at),
  INDEX idx_module_target_created (module, target_id, created_at),
  INDEX idx_event_created (event_type, created_at)
);

SET @has_parent_run_id := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'scheduler_job_runs'
    AND COLUMN_NAME = 'parent_run_id'
);
SET @sql_parent := IF(
  @has_parent_run_id = 0,
  'ALTER TABLE scheduler_job_runs ADD COLUMN parent_run_id varchar(64) NULL AFTER id',
  'SELECT 1'
);
PREPARE stmt_parent FROM @sql_parent;
EXECUTE stmt_parent;
DEALLOCATE PREPARE stmt_parent;

SET @has_retry_count := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'scheduler_job_runs'
    AND COLUMN_NAME = 'retry_count'
);
SET @sql_retry := IF(
  @has_retry_count = 0,
  'ALTER TABLE scheduler_job_runs ADD COLUMN retry_count int NOT NULL DEFAULT 0 AFTER parent_run_id',
  'SELECT 1'
);
PREPARE stmt_retry FROM @sql_retry;
EXECUTE stmt_retry;
DEALLOCATE PREPARE stmt_retry;
