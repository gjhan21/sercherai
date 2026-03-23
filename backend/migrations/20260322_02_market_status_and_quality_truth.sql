CREATE TABLE IF NOT EXISTS stock_status_truth (
  id                       varchar(32) PRIMARY KEY,
  trade_date               date NOT NULL,
  instrument_key           varchar(64) NOT NULL,
  list_date                date DEFAULT NULL,
  selected_source_key      varchar(64) DEFAULT NULL,
  is_suspended             tinyint(1) NOT NULL DEFAULT 0,
  is_st                    tinyint(1) NOT NULL DEFAULT 0,
  risk_warning             tinyint(1) NOT NULL DEFAULT 0,
  status_reason_codes_json text,
  metadata_json            text,
  created_at               datetime NOT NULL,
  updated_at               datetime NOT NULL,
  UNIQUE KEY uk_stock_status_truth (trade_date, instrument_key),
  INDEX idx_stock_status_truth_symbol (instrument_key, trade_date),
  INDEX idx_stock_status_truth_flags (trade_date, is_suspended, is_st, risk_warning)
);

CREATE TABLE IF NOT EXISTS futures_contract_mappings (
  id                     varchar(32) PRIMARY KEY,
  trade_date             date NOT NULL,
  product_key            varchar(64) NOT NULL,
  exchange_code          varchar(32) NOT NULL,
  dominant_instrument_key varchar(64) NOT NULL,
  secondary_instrument_key varchar(64) DEFAULT NULL,
  near_instrument_key    varchar(64) DEFAULT NULL,
  selected_source_key    varchar(64) DEFAULT NULL,
  mapping_method         varchar(64) NOT NULL DEFAULT 'TURNOVER_OPEN_INTEREST',
  metadata_json          text,
  created_at             datetime NOT NULL,
  updated_at             datetime NOT NULL,
  UNIQUE KEY uk_futures_contract_mapping (trade_date, product_key, exchange_code),
  INDEX idx_futures_contract_mapping_trade (trade_date, exchange_code),
  INDEX idx_futures_contract_mapping_dominant (dominant_instrument_key, trade_date)
);

CREATE TABLE IF NOT EXISTS market_data_quality_logs (
  id             varchar(32) PRIMARY KEY,
  asset_class    varchar(16) DEFAULT NULL,
  data_kind      varchar(32) NOT NULL,
  instrument_key varchar(64) DEFAULT NULL,
  trade_date     date DEFAULT NULL,
  source_key     varchar(64) DEFAULT NULL,
  severity       varchar(16) NOT NULL DEFAULT 'WARN',
  issue_code     varchar(64) NOT NULL,
  issue_message  varchar(255) DEFAULT NULL,
  payload_json   text,
  created_at     datetime NOT NULL,
  INDEX idx_market_quality_lookup (data_kind, created_at),
  INDEX idx_market_quality_issue (issue_code, created_at),
  INDEX idx_market_quality_instrument (asset_class, instrument_key, trade_date)
);
