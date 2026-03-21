ALTER TABLE futures_inventory_snapshots
  ADD COLUMN IF NOT EXISTS brand varchar(128) DEFAULT NULL AFTER area,
  ADD COLUMN IF NOT EXISTS place varchar(128) DEFAULT NULL AFTER brand,
  ADD COLUMN IF NOT EXISTS grade varchar(64) DEFAULT NULL AFTER place;

ALTER TABLE futures_inventory_snapshots
  DROP INDEX uk_futures_inventory_snapshot,
  ADD UNIQUE KEY uk_futures_inventory_snapshot (symbol, trade_date, warehouse_id, warehouse, area, brand, place, grade, source_key);
