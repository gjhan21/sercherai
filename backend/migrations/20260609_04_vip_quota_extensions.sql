-- Add new columns for download, forecast, and stock recommendation quotas to vip_quota_configs
SET @has_download_limit := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'vip_quota_configs'
    AND COLUMN_NAME = 'download_limit'
);
SET @sql_dl := IF(
  @has_download_limit = 0,
  'ALTER TABLE vip_quota_configs ADD COLUMN download_limit INT NOT NULL DEFAULT 0 AFTER news_subscribe_limit',
  'SELECT 1'
);
PREPARE stmt_dl FROM @sql_dl;
EXECUTE stmt_dl;
DEALLOCATE PREPARE stmt_dl;

SET @has_forecast_limit := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'vip_quota_configs'
    AND COLUMN_NAME = 'forecast_limit'
);
SET @sql_fl := IF(
  @has_forecast_limit = 0,
  'ALTER TABLE vip_quota_configs ADD COLUMN forecast_limit INT NOT NULL DEFAULT 0 AFTER download_limit',
  'SELECT 1'
);
PREPARE stmt_fl FROM @sql_fl;
EXECUTE stmt_fl;
DEALLOCATE PREPARE stmt_fl;

SET @has_stock_reco_limit := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'vip_quota_configs'
    AND COLUMN_NAME = 'stock_reco_limit'
);
SET @sql_srl := IF(
  @has_stock_reco_limit = 0,
  'ALTER TABLE vip_quota_configs ADD COLUMN stock_reco_limit INT NOT NULL DEFAULT 0 AFTER forecast_limit',
  'SELECT 1'
);
PREPARE stmt_srl FROM @sql_srl;
EXECUTE stmt_srl;
DEALLOCATE PREPARE stmt_srl;

-- Add new columns for download, forecast, and stock recommendation usages to user_quota_usages
SET @has_download_used := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_quota_usages'
    AND COLUMN_NAME = 'download_used'
);
SET @sql_du := IF(
  @has_download_used = 0,
  'ALTER TABLE user_quota_usages ADD COLUMN download_used INT NOT NULL DEFAULT 0 AFTER news_subscribe_used',
  'SELECT 1'
);
PREPARE stmt_du FROM @sql_du;
EXECUTE stmt_du;
DEALLOCATE PREPARE stmt_du;

SET @has_forecast_used := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_quota_usages'
    AND COLUMN_NAME = 'forecast_used'
);
SET @sql_fu := IF(
  @has_forecast_used = 0,
  'ALTER TABLE user_quota_usages ADD COLUMN forecast_used INT NOT NULL DEFAULT 0 AFTER download_used',
  'SELECT 1'
);
PREPARE stmt_fu FROM @sql_fu;
EXECUTE stmt_fu;
DEALLOCATE PREPARE stmt_fu;

SET @has_stock_reco_used := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'user_quota_usages'
    AND COLUMN_NAME = 'stock_reco_used'
);
SET @sql_sru := IF(
  @has_stock_reco_used = 0,
  'ALTER TABLE user_quota_usages ADD COLUMN stock_reco_used INT NOT NULL DEFAULT 0 AFTER forecast_used',
  'SELECT 1'
);
PREPARE stmt_sru FROM @sql_sru;
EXECUTE stmt_sru;
DEALLOCATE PREPARE stmt_sru;
