-- Auth risk config migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS auth_risk_configs (
  id                     varchar(32) PRIMARY KEY,
  phone_fail_threshold   int NOT NULL,
  ip_fail_threshold      int NOT NULL,
  ip_phone_threshold     int NOT NULL,
  lock_seconds           int NOT NULL,
  updated_at             datetime NOT NULL
);

INSERT INTO auth_risk_configs (
  id, phone_fail_threshold, ip_fail_threshold, ip_phone_threshold, lock_seconds, updated_at
)
VALUES ('default', 5, 20, 5, 900, NOW())
ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at);
