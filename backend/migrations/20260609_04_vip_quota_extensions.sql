-- Add new columns for download, forecast, and stock recommendation quotas to vip_quota_configs
ALTER TABLE vip_quota_configs ADD COLUMN download_limit INT NOT NULL DEFAULT 0 AFTER news_subscribe_limit;
ALTER TABLE vip_quota_configs ADD COLUMN forecast_limit INT NOT NULL DEFAULT 0 AFTER download_limit;
ALTER TABLE vip_quota_configs ADD COLUMN stock_reco_limit INT NOT NULL DEFAULT 0 AFTER forecast_limit;

-- Add new columns for download, forecast, and stock recommendation usages to user_quota_usages
ALTER TABLE user_quota_usages ADD COLUMN download_used INT NOT NULL DEFAULT 0 AFTER news_subscribe_used;
ALTER TABLE user_quota_usages ADD COLUMN forecast_used INT NOT NULL DEFAULT 0 AFTER download_used;
ALTER TABLE user_quota_usages ADD COLUMN stock_reco_used INT NOT NULL DEFAULT 0 AFTER forecast_used;
