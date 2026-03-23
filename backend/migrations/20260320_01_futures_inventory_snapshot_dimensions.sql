SET @brand_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'futures_inventory_snapshots'
    AND COLUMN_NAME = 'brand'
);
SET @brand_sql = IF(
  @brand_exists = 0,
  'ALTER TABLE futures_inventory_snapshots ADD COLUMN brand varchar(128) DEFAULT NULL AFTER area',
  'SELECT 1'
);
PREPARE futures_inventory_brand_stmt FROM @brand_sql;
EXECUTE futures_inventory_brand_stmt;
DEALLOCATE PREPARE futures_inventory_brand_stmt;

SET @place_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'futures_inventory_snapshots'
    AND COLUMN_NAME = 'place'
);
SET @place_sql = IF(
  @place_exists = 0,
  'ALTER TABLE futures_inventory_snapshots ADD COLUMN place varchar(128) DEFAULT NULL AFTER brand',
  'SELECT 1'
);
PREPARE futures_inventory_place_stmt FROM @place_sql;
EXECUTE futures_inventory_place_stmt;
DEALLOCATE PREPARE futures_inventory_place_stmt;

SET @grade_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'futures_inventory_snapshots'
    AND COLUMN_NAME = 'grade'
);
SET @grade_sql = IF(
  @grade_exists = 0,
  'ALTER TABLE futures_inventory_snapshots ADD COLUMN grade varchar(64) DEFAULT NULL AFTER place',
  'SELECT 1'
);
PREPARE futures_inventory_grade_stmt FROM @grade_sql;
EXECUTE futures_inventory_grade_stmt;
DEALLOCATE PREPARE futures_inventory_grade_stmt;

ALTER TABLE futures_inventory_snapshots
  DROP INDEX uk_futures_inventory_snapshot,
  ADD UNIQUE KEY uk_futures_inventory_snapshot (symbol, trade_date, warehouse_id, warehouse, area, brand, place, grade, source_key);
