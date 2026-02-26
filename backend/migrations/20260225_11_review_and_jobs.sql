-- Review workflow and scheduler job run logs for MySQL 8.x

CREATE TABLE IF NOT EXISTS review_tasks (
  id            varchar(64) PRIMARY KEY,
  module        varchar(32) NOT NULL,
  target_id     varchar(64) NOT NULL,
  submitter_id  varchar(32) NOT NULL,
  reviewer_id   varchar(32),
  status        varchar(16) NOT NULL,
  submit_note   varchar(512),
  review_note   varchar(512),
  submitted_at  datetime NOT NULL,
  reviewed_at   datetime,
  created_at    datetime NOT NULL,
  updated_at    datetime NOT NULL,
  INDEX idx_module_target_status (module, target_id, status),
  INDEX idx_module_status_submitted (module, status, submitted_at),
  INDEX idx_submitter (submitter_id),
  INDEX idx_reviewer (reviewer_id)
);

CREATE TABLE IF NOT EXISTS scheduler_job_runs (
  id             varchar(64) PRIMARY KEY,
  job_name       varchar(64) NOT NULL,
  trigger_source varchar(32) NOT NULL,
  status         varchar(16) NOT NULL,
  started_at     datetime NOT NULL,
  finished_at    datetime,
  result_summary varchar(512),
  error_message  varchar(512),
  operator_id    varchar(32),
  created_at     datetime NOT NULL,
  INDEX idx_job_status_started (job_name, status, started_at),
  INDEX idx_trigger_started (trigger_source, started_at)
);
