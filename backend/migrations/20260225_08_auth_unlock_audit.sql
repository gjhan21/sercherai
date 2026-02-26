-- Auth unlock audit migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS auth_unlock_logs (
  id               varchar(32) PRIMARY KEY,
  operator_user_id varchar(32) NOT NULL,
  phone            varchar(20),
  ip               varchar(64),
  reason           varchar(256),
  created_at       datetime NOT NULL,
  INDEX idx_operator_created_at (operator_user_id, created_at),
  INDEX idx_phone_created_at (phone, created_at),
  INDEX idx_ip_created_at (ip, created_at)
);
