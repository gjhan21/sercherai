CREATE TABLE IF NOT EXISTS strategy_job_runs (
  job_id varchar(64) NOT NULL,
  job_type varchar(64) NOT NULL,
  status varchar(32) NOT NULL,
  requested_by varchar(64) DEFAULT NULL,
  trace_id varchar(128) DEFAULT NULL,
  trade_date date DEFAULT NULL,
  payload_snapshot json DEFAULT NULL,
  config_refs json DEFAULT NULL,
  publish_policy_preview json DEFAULT NULL,
  error_message text,
  remote_created_at datetime DEFAULT NULL,
  remote_started_at datetime DEFAULT NULL,
  remote_finished_at datetime DEFAULT NULL,
  synced_at datetime NOT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (job_id),
  KEY idx_strategy_job_runs_type_status (job_type, status),
  KEY idx_strategy_job_runs_trade_date (trade_date),
  KEY idx_strategy_job_runs_remote_created_at (remote_created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS strategy_job_artifacts (
  job_id varchar(64) NOT NULL,
  result_summary text,
  warning_messages json DEFAULT NULL,
  artifacts_snapshot json DEFAULT NULL,
  report_snapshot json DEFAULT NULL,
  selected_count int NOT NULL DEFAULT 0,
  payload_count int NOT NULL DEFAULT 0,
  asset_keys json DEFAULT NULL,
  synced_at datetime NOT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (job_id),
  KEY idx_strategy_job_artifacts_selected_count (selected_count),
  CONSTRAINT fk_strategy_job_artifacts_job_id FOREIGN KEY (job_id) REFERENCES strategy_job_runs (job_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS strategy_job_replays (
  id varchar(64) NOT NULL,
  job_id varchar(64) NOT NULL,
  publish_id varchar(64) NOT NULL,
  operator varchar(64) DEFAULT NULL,
  force_publish tinyint(1) NOT NULL DEFAULT 0,
  override_reason text,
  policy_snapshot json DEFAULT NULL,
  replay_snapshot json DEFAULT NULL,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  KEY idx_strategy_job_replays_job_id (job_id),
  KEY idx_strategy_job_replays_publish_id (publish_id),
  CONSTRAINT fk_strategy_job_replays_job_id FOREIGN KEY (job_id) REFERENCES strategy_job_runs (job_id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
