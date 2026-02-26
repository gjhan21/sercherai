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

ALTER TABLE scheduler_job_runs
  ADD COLUMN IF NOT EXISTS parent_run_id varchar(64) NULL AFTER id,
  ADD COLUMN IF NOT EXISTS retry_count int NOT NULL DEFAULT 0 AFTER parent_run_id;
