CREATE TABLE IF NOT EXISTS news_sync_checkpoints (
  sync_key           varchar(64) PRIMARY KEY,
  cursor_updated_at  bigint NOT NULL DEFAULT 0,
  cursor_source_id   bigint NOT NULL DEFAULT 0,
  last_run_at        datetime NULL,
  last_success_at    datetime NULL,
  last_status        varchar(16) NOT NULL DEFAULT 'IDLE',
  last_error         varchar(255) NULL,
  synced_articles    bigint NOT NULL DEFAULT 0,
  synced_attachments bigint NOT NULL DEFAULT 0,
  created_at         datetime NOT NULL,
  updated_at         datetime NOT NULL
);

INSERT INTO scheduler_job_definitions
  (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_doc_fast_news_incremental', 'doc_fast_news_incremental', 'DocFast资讯增量同步', 'NEWS', 'EVERY_100_MINUTES', 'ACTIVE', NULL, 'system', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  module = VALUES(module),
  cron_expr = VALUES(cron_expr),
  status = VALUES(status),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES
  ('cfg_news_sync_doc_fast_enabled', 'news.sync.doc_fast.enabled', 'true', 'doc_fast 新闻增量同步开关', 'system', NOW()),
  ('cfg_news_sync_doc_fast_interval', 'news.sync.doc_fast.interval_minutes', '100', 'doc_fast 新闻增量同步周期(分钟)', 'system', NOW()),
  ('cfg_news_sync_doc_fast_batch', 'news.sync.doc_fast.batch_size', '200', 'doc_fast 新闻增量同步单批条数', 'system', NOW()),
  ('cfg_news_sync_doc_fast_source_base', 'news.sync.doc_fast.source_base_url', 'https://img.cloudup518.top', 'doc_fast 资源URL前缀', 'system', NOW()),
  ('cfg_news_sync_doc_fast_author', 'news.sync.doc_fast.author_id', 'admin_001', 'doc_fast 同步文章作者ID', 'system', NOW())
ON DUPLICATE KEY UPDATE
  config_value = VALUES(config_value),
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);
