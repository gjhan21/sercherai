INSERT INTO scheduler_job_definitions
  (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_stock_master_full_sync', 'stock_master_full_sync', '股票主数据全量同步', 'STOCK', '0 10 3 * * 0', 'ACTIVE', NULL, 'system', NOW(), NOW()),
  ('jobdef_stock_master_incremental_sync', 'stock_master_incremental_sync', '股票主数据增量同步', 'STOCK', '0 40 6 * * 1-5', 'ACTIVE', NULL, 'system', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  module = VALUES(module),
  cron_expr = VALUES(cron_expr),
  status = VALUES(status),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO scheduler_job_definitions
  (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_stock_quotes_incremental_sync', 'stock_quotes_incremental_sync', '股票日线增量同步', 'STOCK', '0 45 16 * * 1-5', 'ACTIVE', NULL, 'system', NOW(), NOW()),
  ('jobdef_stock_daily_basic_incremental_sync', 'stock_daily_basic_incremental_sync', '股票日度基础因子增量同步', 'STOCK', '0 5 17 * * 1-5', 'ACTIVE', NULL, 'system', NOW(), NOW()),
  ('jobdef_stock_moneyflow_incremental_sync', 'stock_moneyflow_incremental_sync', '股票资金流向增量同步', 'STOCK', '0 20 17 * * 1-5', 'ACTIVE', NULL, 'system', NOW(), NOW()),
  ('jobdef_stock_news_incremental_sync', 'stock_news_incremental_sync', '股票公告资讯增量同步', 'STOCK', '0 40 17 * * 1-5', 'ACTIVE', NULL, 'system', NOW(), NOW()),
  ('jobdef_stock_truth_rebuild', 'stock_truth_rebuild', '股票真相重建', 'STOCK', '0 55 17 * * 1-5', 'ACTIVE', NULL, 'system', NOW(), NOW()),
  ('jobdef_stock_data_backfill', 'stock_data_backfill', '股票市场数据补数', 'STOCK', '0 30 4 * * 0', 'ACTIVE', NULL, 'system', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  module = VALUES(module),
  cron_expr = VALUES(cron_expr),
  status = VALUES(status),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_stock_master_default_source',
  'stock.master.default_source_key',
  'TUSHARE',
  '股票主数据默认数据源',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'stock.master.default_source_key'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_stock_master_sync_batch_size',
  'market.stock.master.sync_batch_size',
  '500',
  '股票主数据全量同步批次大小',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.stock.master.sync_batch_size'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_stock_master_incremental_lookback_days',
  'market.stock.master.incremental_lookback_days',
  '7',
  '股票主数据增量同步回看天数',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.stock.master.incremental_lookback_days'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_stock_master_missing_profile_backfill',
  'market.stock.master.backfill_company_profile',
  'true',
  '股票主数据同步后是否补充公司画像',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.stock.master.backfill_company_profile'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_stock_daily_basic_default_source',
  'stock.daily_basic.default_source_key',
  'TUSHARE',
  '股票 daily_basic 默认数据源',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'stock.daily_basic.default_source_key'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_stock_moneyflow_default_source',
  'stock.moneyflow.default_source_key',
  'TUSHARE',
  '股票 moneyflow 默认数据源',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'stock.moneyflow.default_source_key'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_stock_news_default_source',
  'stock.news.default_source_key',
  'TUSHARE',
  '股票公告资讯默认数据源',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'stock.news.default_source_key'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_stock_factor_source_priority',
  'market.stock.factor.source_priority',
  'TUSHARE',
  '股票增强因子数据源优先级',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.stock.factor.source_priority'
);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
SELECT
  'cfg_stock_news_source_priority',
  'market.stock.news.source_priority',
  'TUSHARE',
  '股票公告资讯数据源优先级',
  'system',
  NOW()
FROM DUAL
WHERE NOT EXISTS (
  SELECT 1 FROM system_configs WHERE config_key = 'market.stock.news.source_priority'
);
