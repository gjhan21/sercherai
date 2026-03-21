CREATE TABLE IF NOT EXISTS experiment_events (
  id VARCHAR(32) NOT NULL PRIMARY KEY,
  experiment_key VARCHAR(64) NOT NULL,
  variant_key VARCHAR(32) NOT NULL,
  event_type VARCHAR(32) NOT NULL,
  page_key VARCHAR(64) NOT NULL,
  target_key VARCHAR(64) NULL,
  user_stage VARCHAR(32) NULL,
  anonymous_id VARCHAR(64) NULL,
  session_id VARCHAR(64) NULL,
  pathname VARCHAR(255) NULL,
  referrer VARCHAR(255) NULL,
  metadata_json TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_experiment_events_created_at (created_at),
  KEY idx_experiment_events_exp_variant (experiment_key, variant_key, created_at),
  KEY idx_experiment_events_event_type (event_type, created_at),
  KEY idx_experiment_events_stage (user_stage, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
