-- Membership product/order extensions for MySQL 8.x

SET @has_mp_member_level := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'membership_products'
    AND COLUMN_NAME = 'member_level'
);
SET @sql_mp_member_level := IF(
  @has_mp_member_level = 0,
  'ALTER TABLE membership_products ADD COLUMN member_level varchar(16) NOT NULL DEFAULT ''VIP1''',
  'SELECT 1'
);
PREPARE stmt_mp_member_level FROM @sql_mp_member_level;
EXECUTE stmt_mp_member_level;
DEALLOCATE PREPARE stmt_mp_member_level;

SET @has_mp_duration_days := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'membership_products'
    AND COLUMN_NAME = 'duration_days'
);
SET @sql_mp_duration_days := IF(
  @has_mp_duration_days = 0,
  'ALTER TABLE membership_products ADD COLUMN duration_days int NOT NULL DEFAULT 30',
  'SELECT 1'
);
PREPARE stmt_mp_duration_days FROM @sql_mp_duration_days;
EXECUTE stmt_mp_duration_days;
DEALLOCATE PREPARE stmt_mp_duration_days;

SET @has_mo_order_no := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'membership_orders'
    AND COLUMN_NAME = 'order_no'
);
SET @sql_mo_order_no := IF(
  @has_mo_order_no = 0,
  'ALTER TABLE membership_orders ADD COLUMN order_no varchar(64)',
  'SELECT 1'
);
PREPARE stmt_mo_order_no FROM @sql_mo_order_no;
EXECUTE stmt_mo_order_no;
DEALLOCATE PREPARE stmt_mo_order_no;

SET @has_mo_pay_channel := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'membership_orders'
    AND COLUMN_NAME = 'pay_channel'
);
SET @sql_mo_pay_channel := IF(
  @has_mo_pay_channel = 0,
  'ALTER TABLE membership_orders ADD COLUMN pay_channel varchar(16)',
  'SELECT 1'
);
PREPARE stmt_mo_pay_channel FROM @sql_mo_pay_channel;
EXECUTE stmt_mo_pay_channel;
DEALLOCATE PREPARE stmt_mo_pay_channel;

SET @has_mo_created_at := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'membership_orders'
    AND COLUMN_NAME = 'created_at'
);
SET @sql_mo_created_at := IF(
  @has_mo_created_at = 0,
  'ALTER TABLE membership_orders ADD COLUMN created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP',
  'SELECT 1'
);
PREPARE stmt_mo_created_at FROM @sql_mo_created_at;
EXECUTE stmt_mo_created_at;
DEALLOCATE PREPARE stmt_mo_created_at;

SET @has_mo_updated_at := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'membership_orders'
    AND COLUMN_NAME = 'updated_at'
);
SET @sql_mo_updated_at := IF(
  @has_mo_updated_at = 0,
  'ALTER TABLE membership_orders ADD COLUMN updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP',
  'SELECT 1'
);
PREPARE stmt_mo_updated_at FROM @sql_mo_updated_at;
EXECUTE stmt_mo_updated_at;
DEALLOCATE PREPARE stmt_mo_updated_at;

SET @has_mo_order_no_uk := (
  SELECT COUNT(*)
  FROM information_schema.STATISTICS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'membership_orders'
    AND INDEX_NAME = 'uk_membership_order_no'
);
SET @sql_mo_order_no_uk := IF(
  @has_mo_order_no_uk = 0,
  'ALTER TABLE membership_orders ADD UNIQUE KEY uk_membership_order_no (order_no)',
  'SELECT 1'
);
PREPARE stmt_mo_order_no_uk FROM @sql_mo_order_no_uk;
EXECUTE stmt_mo_order_no_uk;
DEALLOCATE PREPARE stmt_mo_order_no_uk;
