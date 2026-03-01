-- Builtin data sources: mock + tushare

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_data_source_mock_stock',
  'data_source.mock_stock',
  '{"name":"Mock Stock Quotes","source_type":"STOCK","status":"ACTIVE","config":{"provider":"MOCK","endpoint":"http://127.0.0.1:18080/healthz","retry_times":0,"fail_threshold":5,"retry_interval_ms":200,"health_timeout_ms":3000,"alert_receiver_id":"admin_001"}}',
  '内置模拟股票行情数据源',
  'system',
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'data_source.mock_stock'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_data_source_tushare',
  'data_source.tushare',
  '{"name":"Tushare","source_type":"STOCK","status":"ACTIVE","config":{"provider":"TUSHARE","endpoint":"https://api.tushare.pro","token":"","retry_times":1,"fail_threshold":3,"retry_interval_ms":500,"health_timeout_ms":8000,"alert_receiver_id":"admin_001"}}',
  '内置Tushare股票行情数据源',
  'system',
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'data_source.tushare'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_stock_quotes_default_source',
  'stock.quotes.default_source_key',
  'TUSHARE',
  '股票行情默认数据源',
  'system',
  NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'stock.quotes.default_source_key'
);
