CREATE TABLE IF NOT EXISTS experiment_order_attributions (
  order_no VARCHAR(32) NOT NULL PRIMARY KEY,
  experiment_key VARCHAR(64) NOT NULL,
  variant_key VARCHAR(32) NOT NULL,
  page_key VARCHAR(64) NOT NULL,
  target_key VARCHAR(64) NULL,
  user_stage VARCHAR(32) NULL,
  anonymous_id VARCHAR(64) NULL,
  session_id VARCHAR(64) NULL,
  pathname VARCHAR(255) NULL,
  referrer VARCHAR(255) NULL,
  conversion_type VARCHAR(32) NOT NULL,
  metadata_json TEXT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  KEY idx_experiment_order_attr_experiment (experiment_key, variant_key, created_at),
  KEY idx_experiment_order_attr_stage (user_stage, created_at),
  KEY idx_experiment_order_attr_conversion (conversion_type, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
