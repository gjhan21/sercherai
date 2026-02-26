-- Scheduler job definition management for MySQL 8.x

CREATE TABLE IF NOT EXISTS scheduler_job_definitions (
  id           varchar(64) PRIMARY KEY,
  job_name     varchar(64) NOT NULL UNIQUE,
  display_name varchar(128) NOT NULL,
  module       varchar(32) NOT NULL,
  cron_expr    varchar(64) NOT NULL,
  status       varchar(16) NOT NULL,
  last_run_at  datetime,
  updated_by   varchar(32) NOT NULL,
  created_at   datetime NOT NULL,
  updated_at   datetime NOT NULL,
  INDEX idx_module_status (module, status)
);

INSERT INTO scheduler_job_definitions
  (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_stock_daily', 'daily_stock_recommendation', '每日股票推荐生成', 'STOCK', '0 30 8 * * *', 'ACTIVE', NULL, 'system', NOW(), NOW()),
  ('jobdef_futures_daily', 'daily_futures_strategy', '每日期货策略生成', 'FUTURES', '0 35 8 * * *', 'ACTIVE', NULL, 'system', NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at);
