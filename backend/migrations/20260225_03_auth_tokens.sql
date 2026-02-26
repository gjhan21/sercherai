-- Auth tokens migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id          varchar(32) PRIMARY KEY,
  user_id     varchar(32) NOT NULL,
  token_hash  varchar(128) NOT NULL UNIQUE,
  expires_at  datetime NOT NULL,
  revoked     tinyint(1) NOT NULL DEFAULT 0,
  replaced_by varchar(32),
  revoked_at  datetime,
  created_at  datetime NOT NULL,
  INDEX idx_user_created_at (user_id, created_at),
  INDEX idx_expires_at (expires_at)
);
