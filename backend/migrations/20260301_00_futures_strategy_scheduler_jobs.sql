INSERT INTO scheduler_job_definitions
  (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_futures_strategy_generate', 'futures_strategy_generate', '期货策略生成任务', 'FUTURES', '0 35 8 * * *', 'ACTIVE', NULL, 'system', NOW(), NOW()),
  ('jobdef_futures_strategy_evaluate', 'futures_strategy_evaluate', '期货策略评估任务', 'FUTURES', '0 45 8 * * *', 'ACTIVE', NULL, 'system', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  module = VALUES(module),
  cron_expr = VALUES(cron_expr),
  status = VALUES(status),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES
  (
    'cfg_futures_strategy_score_weights',
    'futures.strategy.score_weights',
    '{"trend":0.25,"structure":0.2,"flow":0.15,"risk":0.2,"news":0.1,"performance":0.1}',
    '期货策略评分因子权重(JSON，和为1)',
    'system',
    NOW()
  )
ON DUPLICATE KEY UPDATE
  config_value = VALUES(config_value),
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);
