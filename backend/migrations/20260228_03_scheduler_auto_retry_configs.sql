INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES
  ('cfg_scheduler_auto_retry_enabled', 'scheduler.auto_retry.enabled', 'true', '调度任务自动重试开关', 'system', NOW()),
  ('cfg_scheduler_auto_retry_max_retries', 'scheduler.auto_retry.max_retries', '2', '调度任务自动重试最大次数', 'system', NOW()),
  ('cfg_scheduler_auto_retry_backoff_seconds', 'scheduler.auto_retry.backoff_seconds', '2', '调度任务自动重试退避秒数', 'system', NOW()),
  ('cfg_scheduler_auto_retry_jobs', 'scheduler.auto_retry.jobs', 'daily_stock_quant_pipeline', '开启自动重试的任务清单，逗号分隔', 'system', NOW())
ON DUPLICATE KEY UPDATE
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);
