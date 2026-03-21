CREATE TABLE IF NOT EXISTS market_instruments (
  id             varchar(32) PRIMARY KEY,
  asset_class    varchar(16) NOT NULL,
  instrument_key varchar(64) NOT NULL,
  display_name   varchar(128) NOT NULL DEFAULT '',
  exchange_code  varchar(32) DEFAULT NULL,
  status         varchar(16) NOT NULL DEFAULT 'ACTIVE',
  metadata_json  text,
  created_at     datetime NOT NULL,
  updated_at     datetime NOT NULL,
  UNIQUE KEY uk_market_instrument (asset_class, instrument_key),
  INDEX idx_market_instrument_status (asset_class, status)
);

CREATE TABLE IF NOT EXISTS market_symbol_aliases (
  id              varchar(32) PRIMARY KEY,
  asset_class     varchar(16) NOT NULL,
  instrument_key  varchar(64) NOT NULL,
  source_key      varchar(64) NOT NULL,
  external_symbol varchar(128) NOT NULL,
  status          varchar(16) NOT NULL DEFAULT 'ACTIVE',
  metadata_json   text,
  created_at      datetime NOT NULL,
  updated_at      datetime NOT NULL,
  UNIQUE KEY uk_market_symbol_alias (asset_class, instrument_key, source_key),
  INDEX idx_market_symbol_external (source_key, external_symbol),
  INDEX idx_market_symbol_status (asset_class, source_key, status)
);

CREATE TABLE IF NOT EXISTS market_daily_bars (
  id                varchar(32) PRIMARY KEY,
  asset_class       varchar(16) NOT NULL,
  instrument_key    varchar(64) NOT NULL,
  external_symbol   varchar(128) NOT NULL,
  trade_date        date NOT NULL,
  open_price        decimal(18,6) NOT NULL,
  high_price        decimal(18,6) NOT NULL,
  low_price         decimal(18,6) NOT NULL,
  close_price       decimal(18,6) NOT NULL,
  prev_close_price  decimal(18,6) DEFAULT NULL,
  settle_price      decimal(18,6) DEFAULT NULL,
  prev_settle_price decimal(18,6) DEFAULT NULL,
  volume            bigint NOT NULL DEFAULT 0,
  turnover          decimal(24,6) NOT NULL DEFAULT 0,
  open_interest     decimal(24,6) NOT NULL DEFAULT 0,
  source_key        varchar(64) NOT NULL,
  fetched_at        datetime NOT NULL,
  created_at        datetime NOT NULL,
  updated_at        datetime NOT NULL,
  UNIQUE KEY uk_market_daily_bar (asset_class, instrument_key, trade_date, source_key),
  INDEX idx_market_daily_bar_trade (asset_class, trade_date),
  INDEX idx_market_daily_bar_instrument (asset_class, instrument_key, trade_date)
);

CREATE TABLE IF NOT EXISTS market_daily_bar_truth (
  id                  varchar(32) PRIMARY KEY,
  asset_class         varchar(16) NOT NULL,
  instrument_key      varchar(64) NOT NULL,
  trade_date          date NOT NULL,
  selected_source_key varchar(64) NOT NULL,
  external_symbol     varchar(128) NOT NULL DEFAULT '',
  open_price          decimal(18,6) NOT NULL,
  high_price          decimal(18,6) NOT NULL,
  low_price           decimal(18,6) NOT NULL,
  close_price         decimal(18,6) NOT NULL,
  prev_close_price    decimal(18,6) DEFAULT NULL,
  settle_price        decimal(18,6) DEFAULT NULL,
  prev_settle_price   decimal(18,6) DEFAULT NULL,
  volume              bigint NOT NULL DEFAULT 0,
  turnover            decimal(24,6) NOT NULL DEFAULT 0,
  open_interest       decimal(24,6) NOT NULL DEFAULT 0,
  created_at          datetime NOT NULL,
  updated_at          datetime NOT NULL,
  UNIQUE KEY uk_market_daily_bar_truth (asset_class, instrument_key, trade_date),
  INDEX idx_market_daily_truth_trade (asset_class, trade_date)
);

CREATE TABLE IF NOT EXISTS market_intraday_quotes (
  id              varchar(32) PRIMARY KEY,
  asset_class     varchar(16) NOT NULL,
  instrument_key  varchar(64) NOT NULL,
  external_symbol varchar(128) NOT NULL,
  quote_time      datetime NOT NULL,
  last_price      decimal(18,6) NOT NULL,
  open_price      decimal(18,6) DEFAULT NULL,
  high_price      decimal(18,6) DEFAULT NULL,
  low_price       decimal(18,6) DEFAULT NULL,
  prev_close      decimal(18,6) DEFAULT NULL,
  bid1_price      decimal(18,6) DEFAULT NULL,
  ask1_price      decimal(18,6) DEFAULT NULL,
  bid1_volume     decimal(24,6) DEFAULT NULL,
  ask1_volume     decimal(24,6) DEFAULT NULL,
  volume          decimal(24,6) DEFAULT NULL,
  turnover        decimal(24,6) DEFAULT NULL,
  open_interest   decimal(24,6) DEFAULT NULL,
  source_key      varchar(64) NOT NULL,
  fetched_at      datetime NOT NULL,
  created_at      datetime NOT NULL,
  updated_at      datetime NOT NULL,
  INDEX idx_market_intraday_instrument (asset_class, instrument_key, quote_time),
  INDEX idx_market_intraday_source (source_key, quote_time)
);

CREATE TABLE IF NOT EXISTS market_news_items (
  id             varchar(32) PRIMARY KEY,
  source_key     varchar(64) NOT NULL,
  external_id    varchar(128) NOT NULL,
  news_type      varchar(32) NOT NULL DEFAULT 'MARKET',
  title          varchar(255) NOT NULL,
  summary        text,
  content        longtext,
  url            varchar(1024),
  primary_symbol varchar(64),
  symbols_json   text,
  published_at   datetime NOT NULL,
  metadata_json  text,
  created_at     datetime NOT NULL,
  updated_at     datetime NOT NULL,
  UNIQUE KEY uk_market_news_external (source_key, external_id),
  INDEX idx_market_news_published (published_at),
  INDEX idx_market_news_symbol (primary_symbol, published_at)
);

CREATE TABLE IF NOT EXISTS market_source_snapshots (
  id              varchar(32) PRIMARY KEY,
  source_key      varchar(64) NOT NULL,
  asset_class     varchar(16) DEFAULT NULL,
  data_kind       varchar(32) NOT NULL,
  instrument_key  varchar(64) DEFAULT NULL,
  external_symbol varchar(128) DEFAULT NULL,
  status          varchar(16) NOT NULL,
  error_message   varchar(255) DEFAULT NULL,
  payload_text    longtext,
  fetched_at      datetime NOT NULL,
  created_at      datetime NOT NULL,
  INDEX idx_market_snapshot_lookup (source_key, data_kind, fetched_at),
  INDEX idx_market_snapshot_instrument (asset_class, instrument_key, fetched_at)
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_data_source_akshare',
  'data_source.akshare',
  '{"name":"AkShare","source_type":"MARKET","status":"ACTIVE","config":{"provider":"AKSHARE","python_bin":"../services/strategy-engine/.venv/bin/python","bridge_script":"../services/strategy-engine/app/tools/market_bridge.py","retry_times":0,"fail_threshold":3,"retry_interval_ms":300,"health_timeout_ms":10000,"alert_receiver_id":"admin_001"}}',
  '内置 AkShare 多市场数据源',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'data_source.akshare'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_data_source_tickermd',
  'data_source.tickermd',
  '{"name":"TickerMD","source_type":"MARKET","status":"ACTIVE","config":{"provider":"TICKERMD","endpoint":"http://39.107.99.235:1008","quotes_endpoint":"http://39.107.99.235:1008/getQuote.php","kline_endpoint":"http://39.107.99.235:1008/redis.php","ws_endpoint":"ws://39.107.99.235/ws","retry_times":1,"fail_threshold":3,"retry_interval_ms":500,"health_timeout_ms":8000,"alert_receiver_id":"admin_001"}}',
  '内置 TickerMD 行情备份数据源',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'data_source.tickermd'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_market_stock_source_priority',
  'market.stock.daily.source_priority',
  'TUSHARE,AKSHARE,TICKERMD,MOCK',
  '股票日线多源优先级',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.stock.daily.source_priority'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_market_futures_source_priority',
  'market.futures.daily.source_priority',
  'TUSHARE,TICKERMD,AKSHARE,MOCK',
  '期货日线多源优先级',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.futures.daily.source_priority'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_market_news_source_priority',
  'market.news.source_priority',
  'AKSHARE,TUSHARE',
  '市场资讯多源优先级',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.news.source_priority'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_futures_quotes_default_source',
  'futures.quotes.default_source_key',
  'TUSHARE',
  '期货行情默认数据源',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'futures.quotes.default_source_key'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_market_news_default_source',
  'market.news.default_source_key',
  'AKSHARE',
  '市场资讯默认数据源',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.news.default_source_key'
);
