-- Auth risk audit migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS auth_risk_config_logs (
  id                 varchar(32) PRIMARY KEY,
  operator_user_id   varchar(32) NOT NULL,
  old_phone_fail     int NOT NULL,
  old_ip_fail        int NOT NULL,
  old_ip_phone       int NOT NULL,
  old_lock_seconds   int NOT NULL,
  new_phone_fail     int NOT NULL,
  new_ip_fail        int NOT NULL,
  new_ip_phone       int NOT NULL,
  new_lock_seconds   int NOT NULL,
  created_at         datetime NOT NULL,
  INDEX idx_operator_created_at (operator_user_id, created_at)
);
