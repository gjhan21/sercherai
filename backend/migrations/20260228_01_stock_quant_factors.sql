CREATE TABLE IF NOT EXISTS stock_daily_basic (
  id            varchar(32) PRIMARY KEY,
  symbol        varchar(16) NOT NULL,
  trade_date    date NOT NULL,
  turnover_rate decimal(10,4),
  volume_ratio  decimal(10,4),
  pe_ttm        decimal(14,4),
  pb            decimal(14,4),
  total_mv      decimal(20,4),
  circ_mv       decimal(20,4),
  source_key    varchar(64) NOT NULL,
  created_at    datetime NOT NULL,
  updated_at    datetime NOT NULL,
  UNIQUE KEY uk_symbol_trade_date (symbol, trade_date),
  INDEX idx_trade_date (trade_date),
  INDEX idx_symbol_trade_date (symbol, trade_date)
);

CREATE TABLE IF NOT EXISTS stock_moneyflow_daily (
  id             varchar(32) PRIMARY KEY,
  symbol         varchar(16) NOT NULL,
  trade_date     date NOT NULL,
  net_mf_amount  decimal(20,4),
  buy_lg_amount  decimal(20,4),
  sell_lg_amount decimal(20,4),
  buy_elg_amount decimal(20,4),
  sell_elg_amount decimal(20,4),
  source_key     varchar(64) NOT NULL,
  created_at     datetime NOT NULL,
  updated_at     datetime NOT NULL,
  UNIQUE KEY uk_symbol_trade_date (symbol, trade_date),
  INDEX idx_trade_date (trade_date),
  INDEX idx_symbol_trade_date (symbol, trade_date)
);

CREATE TABLE IF NOT EXISTS stock_news_raw (
  id           varchar(32) PRIMARY KEY,
  source_key   varchar(64) NOT NULL,
  symbol       varchar(16) NOT NULL,
  published_at datetime NOT NULL,
  title        varchar(512) NOT NULL,
  content      mediumtext,
  url          varchar(512),
  sentiment    varchar(16) NOT NULL DEFAULT 'NEUTRAL',
  created_at   datetime NOT NULL,
  updated_at   datetime NOT NULL,
  UNIQUE KEY uk_source_symbol_title_time (source_key, symbol, published_at, title(191)),
  INDEX idx_symbol_published_at (symbol, published_at),
  INDEX idx_published_at (published_at)
);

CREATE TABLE IF NOT EXISTS stock_factor_snapshot (
  id             varchar(32) PRIMARY KEY,
  symbol         varchar(16) NOT NULL,
  trade_date     date NOT NULL,
  total_score    decimal(8,4) NOT NULL,
  trend_score    decimal(8,4),
  flow_score     decimal(8,4),
  value_score    decimal(8,4),
  news_score     decimal(8,4),
  risk_level     varchar(16),
  reason_summary varchar(512),
  reasons_json   json,
  source_key     varchar(64) NOT NULL,
  created_at     datetime NOT NULL,
  updated_at     datetime NOT NULL,
  UNIQUE KEY uk_symbol_trade_date (symbol, trade_date),
  INDEX idx_trade_date (trade_date),
  INDEX idx_total_score (trade_date, total_score)
);

CREATE TABLE IF NOT EXISTS stock_rank_daily (
  id             varchar(32) PRIMARY KEY,
  trade_date     date NOT NULL,
  rank_no        int NOT NULL,
  symbol         varchar(16) NOT NULL,
  name           varchar(64),
  total_score    decimal(8,4) NOT NULL,
  risk_level     varchar(16),
  reason_summary varchar(512),
  payload_json   json,
  created_at     datetime NOT NULL,
  updated_at     datetime NOT NULL,
  UNIQUE KEY uk_trade_date_rank (trade_date, rank_no),
  UNIQUE KEY uk_trade_date_symbol (trade_date, symbol),
  INDEX idx_symbol_trade_date (symbol, trade_date)
);
