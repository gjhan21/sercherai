SET @has_users_vip_started_at := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'vip_started_at'
);
SET @sql_users_vip_started_at := IF(
  @has_users_vip_started_at = 0,
  'ALTER TABLE users ADD COLUMN vip_started_at datetime NULL',
  'SELECT 1'
);
PREPARE stmt_users_vip_started_at FROM @sql_users_vip_started_at;
EXECUTE stmt_users_vip_started_at;
DEALLOCATE PREPARE stmt_users_vip_started_at;

SET @has_users_vip_expire_at := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'vip_expire_at'
);
SET @sql_users_vip_expire_at := IF(
  @has_users_vip_expire_at = 0,
  'ALTER TABLE users ADD COLUMN vip_expire_at datetime NULL',
  'SELECT 1'
);
PREPARE stmt_users_vip_expire_at FROM @sql_users_vip_expire_at;
EXECUTE stmt_users_vip_expire_at;
DEALLOCATE PREPARE stmt_users_vip_expire_at;

SET @has_users_vip_remind_3d_at := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'vip_remind_3d_at'
);
SET @sql_users_vip_remind_3d_at := IF(
  @has_users_vip_remind_3d_at = 0,
  'ALTER TABLE users ADD COLUMN vip_remind_3d_at datetime NULL',
  'SELECT 1'
);
PREPARE stmt_users_vip_remind_3d_at FROM @sql_users_vip_remind_3d_at;
EXECUTE stmt_users_vip_remind_3d_at;
DEALLOCATE PREPARE stmt_users_vip_remind_3d_at;

SET @has_users_vip_remind_1d_at := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND COLUMN_NAME = 'vip_remind_1d_at'
);
SET @sql_users_vip_remind_1d_at := IF(
  @has_users_vip_remind_1d_at = 0,
  'ALTER TABLE users ADD COLUMN vip_remind_1d_at datetime NULL',
  'SELECT 1'
);
PREPARE stmt_users_vip_remind_1d_at FROM @sql_users_vip_remind_1d_at;
EXECUTE stmt_users_vip_remind_1d_at;
DEALLOCATE PREPARE stmt_users_vip_remind_1d_at;

SET @has_idx_users_member_vip_expire := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'users'
    AND INDEX_NAME = 'idx_users_member_vip_expire'
);
SET @sql_idx_users_member_vip_expire := IF(
  @has_idx_users_member_vip_expire = 0,
  'ALTER TABLE users ADD INDEX idx_users_member_vip_expire (member_level, vip_expire_at)',
  'SELECT 1'
);
PREPARE stmt_idx_users_member_vip_expire FROM @sql_idx_users_member_vip_expire;
EXECUTE stmt_idx_users_member_vip_expire;
DEALLOCATE PREPARE stmt_idx_users_member_vip_expire;

INSERT INTO scheduler_job_definitions
  (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_vip_membership_lifecycle', 'vip_membership_lifecycle', 'VIP会员生命周期任务', 'SYSTEM', 'EVERY_30_MINUTES', 'ACTIVE', NULL, 'system', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  module = VALUES(module),
  cron_expr = VALUES(cron_expr),
  status = VALUES(status),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);

INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES
  ('cfg_vip_lifecycle_enabled', 'membership.vip.lifecycle.enabled', 'true', 'VIP会员生命周期任务开关', 'system', NOW()),
  ('cfg_vip_lifecycle_interval_minutes', 'membership.vip.lifecycle.interval_minutes', '30', 'VIP会员生命周期任务间隔(分钟)', 'system', NOW()),
  ('cfg_vip_lifecycle_remind_days_3', 'membership.vip.lifecycle.remind_days_3', '3', 'VIP到期提前3天提醒', 'system', NOW()),
  ('cfg_vip_lifecycle_remind_days_1', 'membership.vip.lifecycle.remind_days_1', '1', 'VIP到期提前1天提醒', 'system', NOW())
ON DUPLICATE KEY UPDATE
  config_value = VALUES(config_value),
  description = VALUES(description),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);
