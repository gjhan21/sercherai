-- System configs migration for MySQL 8.x

CREATE TABLE IF NOT EXISTS system_configs (
  id          varchar(64) PRIMARY KEY,
  config_key  varchar(64) NOT NULL UNIQUE,
  config_value text NOT NULL,
  description varchar(256),
  updated_by  varchar(32) NOT NULL,
  updated_at  datetime NOT NULL
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES
  ('cfg_stock_model', 'stock.model.version', 'v1', '股票推荐模型版本', 'system', NOW()),
  ('cfg_futures_model', 'futures.model.version', 'v1', '期货策略模型版本', 'system', NOW())
ON DUPLICATE KEY UPDATE updated_at = VALUES(updated_at);
