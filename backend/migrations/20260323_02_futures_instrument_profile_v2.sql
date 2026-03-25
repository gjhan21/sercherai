CREATE TABLE IF NOT EXISTS futures_instrument_profiles_v2 (
  asset_class            varchar(16) NOT NULL,
  product_key            varchar(64) NOT NULL,
  commodity_label        varchar(128) NOT NULL,
  exchange_code          varchar(32) DEFAULT NULL,
  contract_chain_json    text,
  delivery_places_json   text,
  warehouses_json        text,
  brands_json            text,
  grades_json            text,
  inventory_metric_keys_json text,
  metadata_json          longtext,
  source_updated_at      datetime DEFAULT NULL,
  created_at             datetime NOT NULL,
  updated_at             datetime NOT NULL,
  PRIMARY KEY (asset_class, product_key),
  INDEX idx_futures_instrument_profiles_exchange (exchange_code, updated_at)
);
