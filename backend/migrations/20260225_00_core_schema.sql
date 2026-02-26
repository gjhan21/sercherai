-- Core schema migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS users (
  id            varchar(32) PRIMARY KEY,
  phone         varchar(20) NOT NULL UNIQUE,
  email         varchar(128),
  password_hash varchar(128) NOT NULL,
  status        varchar(16) NOT NULL,
  kyc_status    varchar(16) NOT NULL,
  member_level  varchar(16) NOT NULL,
  created_at    datetime NOT NULL,
  updated_at    datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS kyc_records (
  id           varchar(32) PRIMARY KEY,
  user_id      varchar(32) NOT NULL,
  real_name    varchar(64) NOT NULL,
  id_number    varchar(32) NOT NULL,
  provider     varchar(32) NOT NULL,
  status       varchar(16) NOT NULL,
  reason       varchar(256),
  submitted_at datetime NOT NULL,
  reviewed_at  datetime,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS stock_recommendations (
  id             varchar(32) PRIMARY KEY,
  symbol         varchar(16) NOT NULL,
  name           varchar(64) NOT NULL,
  score          decimal(6,2) NOT NULL,
  risk_level     varchar(16) NOT NULL,
  position_range varchar(32),
  valid_from     datetime NOT NULL,
  valid_to       datetime NOT NULL,
  status         varchar(16) NOT NULL,
  reason_summary varchar(512),
  created_at     datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS stock_reco_details (
  reco_id           varchar(32) PRIMARY KEY,
  tech_score        decimal(6,2),
  fund_score        decimal(6,2),
  sentiment_score   decimal(6,2),
  money_flow_score  decimal(6,2),
  take_profit       varchar(256),
  stop_loss         varchar(256),
  risk_note         varchar(512),
  FOREIGN KEY (reco_id) REFERENCES stock_recommendations(id)
);

CREATE TABLE IF NOT EXISTS futures_strategies (
  id             varchar(32) PRIMARY KEY,
  contract       varchar(32) NOT NULL,
  name           varchar(64),
  direction      varchar(16) NOT NULL,
  risk_level     varchar(16) NOT NULL,
  position_range varchar(32),
  valid_from     datetime NOT NULL,
  valid_to       datetime NOT NULL,
  status         varchar(16) NOT NULL,
  reason_summary varchar(512)
);

CREATE TABLE IF NOT EXISTS arbitrage_recos (
  id           varchar(32) PRIMARY KEY,
  type         varchar(16) NOT NULL,
  contract_a   varchar(32) NOT NULL,
  contract_b   varchar(32) NOT NULL,
  spread       decimal(10,2),
  percentile   decimal(6,4),
  entry_point  decimal(10,2),
  exit_point   decimal(10,2),
  stop_point   decimal(10,2),
  trigger_rule varchar(256),
  status       varchar(16) NOT NULL
);

CREATE TABLE IF NOT EXISTS futures_guidances (
  id                varchar(32) PRIMARY KEY,
  contract          varchar(32) NOT NULL,
  guidance_direction varchar(16) NOT NULL,
  position_level    varchar(16) NOT NULL,
  entry_range       varchar(64),
  take_profit_range varchar(64),
  stop_loss_range   varchar(64),
  risk_level        varchar(16) NOT NULL,
  invalid_condition varchar(512),
  valid_to          datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS market_events (
  id           varchar(32) PRIMARY KEY,
  event_type   varchar(32) NOT NULL,
  symbol       varchar(32) NOT NULL,
  summary      varchar(512),
  trigger_rule varchar(256),
  source       varchar(64),
  created_at   datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS public_holdings (
  id           varchar(32) PRIMARY KEY,
  holder       varchar(128) NOT NULL,
  symbol       varchar(32) NOT NULL,
  ratio        decimal(6,3),
  disclosed_at datetime NOT NULL,
  source       varchar(64)
);

CREATE TABLE IF NOT EXISTS futures_positions_public (
  id             varchar(32) PRIMARY KEY,
  contract       varchar(32) NOT NULL,
  long_position  decimal(16,2),
  short_position decimal(16,2),
  disclosed_at   datetime NOT NULL,
  source         varchar(64)
);

CREATE TABLE IF NOT EXISTS subscriptions (
  id        varchar(32) PRIMARY KEY,
  user_id   varchar(32) NOT NULL,
  type      varchar(32) NOT NULL,
  scope     varchar(64),
  frequency varchar(16) NOT NULL,
  status    varchar(16) NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS messages (
  id          varchar(32) PRIMARY KEY,
  user_id     varchar(32) NOT NULL,
  title       varchar(128) NOT NULL,
  content     text NOT NULL,
  type        varchar(16) NOT NULL,
  read_status varchar(16) NOT NULL,
  created_at  datetime NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS membership_products (
  id     varchar(32) PRIMARY KEY,
  name   varchar(64) NOT NULL,
  price  decimal(10,2) NOT NULL,
  status varchar(16) NOT NULL
);

CREATE TABLE IF NOT EXISTS membership_orders (
  id         varchar(32) PRIMARY KEY,
  user_id    varchar(32) NOT NULL,
  product_id varchar(32) NOT NULL,
  amount     decimal(10,2) NOT NULL,
  status     varchar(16) NOT NULL,
  paid_at    datetime,
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (product_id) REFERENCES membership_products(id)
);

CREATE TABLE IF NOT EXISTS vip_quota_configs (
  id                   varchar(32) PRIMARY KEY,
  member_level         varchar(16) NOT NULL,
  doc_read_limit       int NOT NULL,
  news_subscribe_limit int NOT NULL,
  reset_cycle          varchar(16) NOT NULL,
  status               varchar(16) NOT NULL,
  effective_at         datetime NOT NULL,
  updated_at           datetime NOT NULL,
  UNIQUE KEY uk_member_level_effective (member_level, effective_at),
  INDEX idx_member_level_status (member_level, status)
);

CREATE TABLE IF NOT EXISTS user_quota_usages (
  id                  varchar(32) PRIMARY KEY,
  user_id             varchar(32) NOT NULL,
  member_level        varchar(16) NOT NULL,
  period_key          varchar(16) NOT NULL,
  doc_read_used       int NOT NULL DEFAULT 0,
  news_subscribe_used int NOT NULL DEFAULT 0,
  updated_at          datetime NOT NULL,
  UNIQUE KEY uk_user_period (user_id, period_key),
  INDEX idx_member_level_period (member_level, period_key),
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS news_categories (
  id         varchar(32) PRIMARY KEY,
  name       varchar(64) NOT NULL,
  slug       varchar(64) NOT NULL UNIQUE,
  sort       int NOT NULL DEFAULT 0,
  visibility varchar(16) NOT NULL,
  status     varchar(16) NOT NULL,
  created_at datetime NOT NULL,
  updated_at datetime NOT NULL
);

CREATE TABLE IF NOT EXISTS news_articles (
  id           varchar(32) PRIMARY KEY,
  category_id  varchar(32) NOT NULL,
  title        varchar(256) NOT NULL,
  summary      varchar(512),
  content      mediumtext NOT NULL,
  tags         json,
  visibility   varchar(16) NOT NULL,
  status       varchar(16) NOT NULL,
  published_at datetime,
  author_id    varchar(32) NOT NULL,
  created_at   datetime NOT NULL,
  updated_at   datetime NOT NULL,
  FOREIGN KEY (category_id) REFERENCES news_categories(id)
);

CREATE TABLE IF NOT EXISTS news_attachments (
  id         varchar(32) PRIMARY KEY,
  article_id varchar(32) NOT NULL,
  file_name  varchar(256) NOT NULL,
  file_url   varchar(512) NOT NULL,
  file_size  bigint NOT NULL,
  mime_type  varchar(128),
  created_at datetime NOT NULL,
  FOREIGN KEY (article_id) REFERENCES news_articles(id)
);

