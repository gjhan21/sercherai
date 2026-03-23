SET @ddl := IF (
  EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'market_instruments'
      AND column_name = 'selected_source_key'
  ),
  'SELECT 1',
  'ALTER TABLE market_instruments ADD COLUMN selected_source_key varchar(64) DEFAULT NULL AFTER exchange_code'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := IF (
  EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'market_instruments'
      AND column_name = 'product_key'
  ),
  'SELECT 1',
  'ALTER TABLE market_instruments ADD COLUMN product_key varchar(64) DEFAULT NULL AFTER selected_source_key'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := IF (
  EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'market_instruments'
      AND column_name = 'list_date'
  ),
  'SELECT 1',
  'ALTER TABLE market_instruments ADD COLUMN list_date date DEFAULT NULL AFTER product_key'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := IF (
  EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'market_instruments'
      AND column_name = 'delist_date'
  ),
  'SELECT 1',
  'ALTER TABLE market_instruments ADD COLUMN delist_date date DEFAULT NULL AFTER list_date'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := IF (
  EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'market_instruments'
      AND column_name = 'truth_version'
  ),
  'SELECT 1',
  'ALTER TABLE market_instruments ADD COLUMN truth_version bigint NOT NULL DEFAULT 0 AFTER delist_date'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := IF (
  EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'market_instruments'
      AND column_name = 'quality_score'
  ),
  'SELECT 1',
  'ALTER TABLE market_instruments ADD COLUMN quality_score decimal(6,2) NOT NULL DEFAULT 0 AFTER truth_version'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := IF (
  EXISTS (
    SELECT 1
    FROM information_schema.columns
    WHERE table_schema = DATABASE()
      AND table_name = 'market_instruments'
      AND column_name = 'source_updated_at'
  ),
  'SELECT 1',
  'ALTER TABLE market_instruments ADD COLUMN source_updated_at datetime DEFAULT NULL AFTER quality_score'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := IF (
  EXISTS (
    SELECT 1
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'market_instruments'
      AND index_name = 'idx_market_instrument_source'
  ),
  'SELECT 1',
  'ALTER TABLE market_instruments ADD INDEX idx_market_instrument_source (asset_class, selected_source_key)'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @ddl := IF (
  EXISTS (
    SELECT 1
    FROM information_schema.statistics
    WHERE table_schema = DATABASE()
      AND table_name = 'market_instruments'
      AND index_name = 'idx_market_instrument_product'
  ),
  'SELECT 1',
  'ALTER TABLE market_instruments ADD INDEX idx_market_instrument_product (asset_class, product_key)'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

CREATE TABLE IF NOT EXISTS market_instrument_source_facts (
  id                varchar(32) PRIMARY KEY,
  asset_class       varchar(16) NOT NULL,
  instrument_key    varchar(64) NOT NULL,
  source_key        varchar(64) NOT NULL,
  external_symbol   varchar(128) DEFAULT NULL,
  display_name      varchar(128) DEFAULT NULL,
  exchange_code     varchar(32) DEFAULT NULL,
  product_key       varchar(64) DEFAULT NULL,
  list_date         date DEFAULT NULL,
  delist_date       date DEFAULT NULL,
  status            varchar(16) NOT NULL DEFAULT 'ACTIVE',
  metadata_json     text,
  quality_score     decimal(6,2) NOT NULL DEFAULT 0,
  source_updated_at datetime DEFAULT NULL,
  fetched_at        datetime NOT NULL,
  created_at        datetime NOT NULL,
  updated_at        datetime NOT NULL,
  UNIQUE KEY uk_market_instrument_source_fact (asset_class, instrument_key, source_key),
  INDEX idx_market_instrument_source_fact_lookup (asset_class, instrument_key),
  INDEX idx_market_instrument_source_fact_source (source_key, source_updated_at),
  INDEX idx_market_instrument_source_fact_symbol (source_key, external_symbol)
);

UPDATE market_instruments
SET
  selected_source_key = COALESCE(NULLIF(selected_source_key, ''), 'LOCAL_PLACEHOLDER'),
  product_key = COALESCE(NULLIF(product_key, ''), SUBSTRING_INDEX(instrument_key, '.', 1)),
  truth_version = CASE
    WHEN truth_version IS NULL OR truth_version = 0 THEN UNIX_TIMESTAMP(COALESCE(updated_at, NOW()))
    ELSE truth_version
  END,
  quality_score = CASE
    WHEN quality_score IS NULL OR quality_score = 0 THEN 0.20
    ELSE quality_score
  END,
  source_updated_at = COALESCE(source_updated_at, updated_at);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_market_instrument_stock_priority',
  'market.instrument.stock.source_priority',
  'TUSHARE,AKSHARE,TICKERMD,MYSELF,MOCK',
  '股票主数据多源优先级',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.instrument.stock.source_priority'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_market_instrument_futures_priority',
  'market.instrument.futures.source_priority',
  'TUSHARE,AKSHARE,TICKERMD,MYSELF,MOCK',
  '期货主数据多源优先级',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.instrument.futures.source_priority'
);
