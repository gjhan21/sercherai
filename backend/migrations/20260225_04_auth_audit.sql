-- Auth audit migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS auth_login_logs (
  id         varchar(32) PRIMARY KEY,
  user_id    varchar(32),
  phone      varchar(20),
  action     varchar(16) NOT NULL,
  status     varchar(16) NOT NULL,
  reason     varchar(256),
  ip         varchar(64),
  user_agent varchar(256),
  created_at datetime NOT NULL,
  INDEX idx_user_created_at (user_id, created_at),
  INDEX idx_phone_created_at (phone, created_at),
  INDEX idx_action_status_created_at (action, status, created_at)
);
