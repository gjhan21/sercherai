-- Settlement and risk extension migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS reward_wallets (
  id               varchar(32) PRIMARY KEY,
  user_id          varchar(32) NOT NULL UNIQUE,
  cash_balance     decimal(12,2) NOT NULL DEFAULT 0,
  cash_frozen      decimal(12,2) NOT NULL DEFAULT 0,
  coupon_balance   decimal(12,2) NOT NULL DEFAULT 0,
  vip_days_balance int NOT NULL DEFAULT 0,
  updated_at       datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS reward_wallet_txns (
  id         varchar(32) PRIMARY KEY,
  wallet_id  varchar(32) NOT NULL,
  txn_type   varchar(16) NOT NULL,
  amount     decimal(12,2) NOT NULL,
  status     varchar(16) NOT NULL,
  ref_id     varchar(64),
  created_at datetime NOT NULL,
  FOREIGN KEY (wallet_id) REFERENCES reward_wallets(id),
  INDEX idx_wallet_created_at (wallet_id, created_at)
);

CREATE TABLE IF NOT EXISTS withdraw_requests (
  id            varchar(32) PRIMARY KEY,
  user_id       varchar(32) NOT NULL,
  amount        decimal(12,2) NOT NULL,
  status        varchar(16) NOT NULL,
  review_reason varchar(256),
  applied_at    datetime NOT NULL,
  reviewed_at   datetime,
  INDEX idx_user_applied_at (user_id, applied_at)
);

CREATE TABLE IF NOT EXISTS payment_callback_logs (
  id              varchar(32) PRIMARY KEY,
  pay_channel     varchar(16) NOT NULL,
  order_no        varchar(64) NOT NULL,
  channel_txn_no  varchar(64) NOT NULL,
  sign_verified   tinyint(1) NOT NULL,
  idempotency_key varchar(128) NOT NULL,
  callback_status varchar(16) NOT NULL,
  created_at      datetime NOT NULL,
  UNIQUE KEY uk_idempotency_key (idempotency_key)
);

CREATE TABLE IF NOT EXISTS reconciliation_records (
  id          varchar(32) PRIMARY KEY,
  pay_channel varchar(16) NOT NULL,
  batch_date  date NOT NULL,
  status      varchar(16) NOT NULL,
  diff_count  int NOT NULL DEFAULT 0,
  created_at  datetime NOT NULL,
  UNIQUE KEY uk_channel_batch (pay_channel, batch_date)
);

CREATE TABLE IF NOT EXISTS risk_rule_configs (
  id           varchar(32) PRIMARY KEY,
  rule_code    varchar(32) NOT NULL UNIQUE,
  rule_name    varchar(128) NOT NULL,
  threshold    int NOT NULL,
  status       varchar(16) NOT NULL,
  effective_at datetime NOT NULL,
  updated_at   datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS risk_hit_logs (
  id         varchar(32) PRIMARY KEY,
  rule_code  varchar(32) NOT NULL,
  user_id    varchar(32) NOT NULL,
  event_id   varchar(64) NOT NULL,
  risk_level varchar(16) NOT NULL,
  status     varchar(32) NOT NULL,
  created_at datetime NOT NULL,
  INDEX idx_user_created_at (user_id, created_at),
  INDEX idx_rule_status (rule_code, status)
);

CREATE TABLE IF NOT EXISTS attachment_download_logs (
  id            varchar(32) PRIMARY KEY,
  user_id       varchar(32) NOT NULL,
  attachment_id varchar(32) NOT NULL,
  article_id    varchar(32) NOT NULL,
  downloaded_at datetime NOT NULL,
  INDEX idx_user_downloaded_at (user_id, downloaded_at)
);

