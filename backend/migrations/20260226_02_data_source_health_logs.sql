-- Data source health check logs for MySQL 8.x

CREATE TABLE IF NOT EXISTS data_source_health_logs (
  id          varchar(64) PRIMARY KEY,
  source_key  varchar(64) NOT NULL,
  status      varchar(16) NOT NULL,
  reachable   tinyint(1) NOT NULL DEFAULT 0,
  http_status int,
  latency_ms  bigint NOT NULL DEFAULT 0,
  message     varchar(512),
  checked_at  datetime NOT NULL,
  INDEX idx_source_checked_at (source_key, checked_at),
  INDEX idx_checked_at (checked_at)
);
