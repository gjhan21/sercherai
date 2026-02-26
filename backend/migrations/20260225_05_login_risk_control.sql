-- Login risk control migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS auth_login_failures (
  phone        varchar(20) PRIMARY KEY,
  fail_count   int NOT NULL DEFAULT 0,
  locked_until datetime,
  updated_at   datetime NOT NULL
);
