CREATE TABLE IF NOT EXISTS market_rhythm_tasks (
  id VARCHAR(32) NOT NULL PRIMARY KEY,
  task_date DATE NOT NULL,
  slot VARCHAR(16) NOT NULL,
  task_key VARCHAR(64) NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'TODO',
  owner VARCHAR(64) NULL,
  notes TEXT NULL,
  source_links_json TEXT NULL,
  completed_at DATETIME NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  UNIQUE KEY uk_market_rhythm_tasks_date_slot_key (task_date, slot, task_key),
  KEY idx_market_rhythm_tasks_date_slot (task_date, slot),
  KEY idx_market_rhythm_tasks_status_date (status, task_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
