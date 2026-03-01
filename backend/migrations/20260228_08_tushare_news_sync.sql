INSERT INTO news_categories
  (id, name, slug, sort, visibility, status, created_at, updated_at)
SELECT
  'nc_ts_news_brief', '新闻快讯', 'ts-news-brief', 210, 'PUBLIC', 'PUBLISHED', NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM news_categories WHERE slug = 'ts-news-brief'
);

INSERT INTO news_categories
  (id, name, slug, sort, visibility, status, created_at, updated_at)
SELECT
  'nc_ts_news_major', '新闻通讯', 'ts-news-major', 220, 'PUBLIC', 'PUBLISHED', NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM news_categories WHERE slug = 'ts-news-major'
);

INSERT INTO news_categories
  (id, name, slug, sort, visibility, status, created_at, updated_at)
SELECT
  'nc_ts_report_research', '券商研报', 'ts-report-research', 230, 'VIP', 'PUBLISHED', NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM news_categories WHERE slug = 'ts-report-research'
);

INSERT INTO news_categories
  (id, name, slug, sort, visibility, status, created_at, updated_at)
SELECT
  'nc_ts_report_forecast', '盈利预测', 'ts-report-forecast', 240, 'VIP', 'PUBLISHED', NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM news_categories WHERE slug = 'ts-report-forecast'
);

INSERT INTO news_categories
  (id, name, slug, sort, visibility, status, created_at, updated_at)
SELECT
  'nc_ts_announcement', '上市公司公告', 'ts-announcement', 250, 'PUBLIC', 'PUBLISHED', NOW(), NOW()
WHERE NOT EXISTS (
  SELECT 1 FROM news_categories WHERE slug = 'ts-announcement'
);

INSERT INTO scheduler_job_definitions
  (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_tushare_news_incremental', 'tushare_news_incremental', 'Tushare资讯增量同步', 'NEWS', 'EVERY_20_MINUTES', 'ACTIVE', NULL, 'system', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  module = VALUES(module),
  cron_expr = VALUES(cron_expr),
  status = VALUES(status),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES
  ('cfg_news_sync_tushare_enabled', 'news.sync.tushare.enabled', 'true', 'Tushare资讯增量同步开关', 'system', NOW()),
  ('cfg_news_sync_tushare_interval', 'news.sync.tushare.interval_minutes', '20', 'Tushare资讯增量同步周期(分钟)', 'system', NOW()),
  ('cfg_news_sync_tushare_batch', 'news.sync.tushare.batch_size', '200', 'Tushare资讯单次同步条数上限', 'system', NOW()),
  ('cfg_news_sync_tushare_timeout', 'news.sync.tushare.timeout_ms', '12000', 'Tushare资讯接口超时时间毫秒', 'system', NOW()),
  ('cfg_news_sync_tushare_author', 'news.sync.tushare.author_id', 'admin_001', 'Tushare资讯同步文章作者ID', 'system', NOW()),
  ('cfg_news_sync_tushare_brief_sources', 'news.sync.tushare.sources.news_brief', 'cls,yicai,sina,eastmoney', '新闻快讯来源列表(英文逗号分隔)', 'system', NOW()),
  ('cfg_news_sync_tushare_major_sources', 'news.sync.tushare.sources.news_major', '新华网,财联社,第一财经,新浪财经,华尔街见闻', '新闻通讯来源列表(中文逗号分隔)', 'system', NOW()),
  ('cfg_news_sync_tushare_brief_lookback', 'news.sync.tushare.lookback_hours.news_brief', '8', '新闻快讯回溯小时数', 'system', NOW()),
  ('cfg_news_sync_tushare_major_lookback', 'news.sync.tushare.lookback_hours.news_major', '24', '新闻通讯回溯小时数', 'system', NOW()),
  ('cfg_news_sync_tushare_research_lookback', 'news.sync.tushare.lookback_days.research_report', '7', '券商研报回溯天数', 'system', NOW()),
  ('cfg_news_sync_tushare_forecast_lookback', 'news.sync.tushare.lookback_days.report_rc', '7', '盈利预测回溯天数', 'system', NOW()),
  ('cfg_news_sync_tushare_announcement_lookback', 'news.sync.tushare.lookback_days.announcement', '2', '上市公司公告回溯天数', 'system', NOW()),
  ('cfg_news_sync_tushare_visibility_brief', 'news.sync.tushare.visibility.news_brief', 'PUBLIC', '新闻快讯文章可见性', 'system', NOW()),
  ('cfg_news_sync_tushare_visibility_major', 'news.sync.tushare.visibility.news_major', 'PUBLIC', '新闻通讯文章可见性', 'system', NOW()),
  ('cfg_news_sync_tushare_visibility_research', 'news.sync.tushare.visibility.research_report', 'VIP', '券商研报文章可见性', 'system', NOW()),
  ('cfg_news_sync_tushare_visibility_forecast', 'news.sync.tushare.visibility.report_rc', 'VIP', '盈利预测文章可见性', 'system', NOW()),
  ('cfg_news_sync_tushare_visibility_announcement', 'news.sync.tushare.visibility.announcement', 'PUBLIC', '上市公司公告文章可见性', 'system', NOW())
ON DUPLICATE KEY UPDATE
  config_value = VALUES(config_value),
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);
