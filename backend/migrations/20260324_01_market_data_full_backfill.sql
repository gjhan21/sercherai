-- Market data full backfill foundation for MySQL 8.x

CREATE TABLE IF NOT EXISTS market_universe_snapshots (
  id           varchar(64) PRIMARY KEY,
  scope        json NOT NULL,
  source_key   varchar(64) NOT NULL,
  snapshot_date date NOT NULL,
  summary_json json NULL,
  created_by   varchar(64) NULL,
  created_at   datetime NOT NULL,
  INDEX idx_market_universe_snapshot_date (snapshot_date, source_key)
);

CREATE TABLE IF NOT EXISTS market_universe_snapshot_items (
  id               varchar(64) PRIMARY KEY,
  snapshot_id      varchar(64) NOT NULL,
  asset_type       varchar(32) NOT NULL,
  instrument_key   varchar(128) NOT NULL,
  external_symbol  varchar(128) NULL,
  display_name     varchar(255) NULL,
  exchange_code    varchar(32) NULL,
  status           varchar(32) NULL,
  list_date        date NULL,
  delist_date      date NULL,
  raw_metadata_json json NULL,
  created_at       datetime NOT NULL,
  UNIQUE KEY uk_market_universe_snapshot_item (snapshot_id, asset_type, instrument_key),
  INDEX idx_market_universe_snapshot_asset (snapshot_id, asset_type),
  CONSTRAINT fk_market_universe_snapshot_items_snapshot
    FOREIGN KEY (snapshot_id) REFERENCES market_universe_snapshots(id)
    ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS market_backfill_runs (
  id                  varchar(64) PRIMARY KEY,
  scheduler_run_id    varchar(64) NOT NULL,
  run_type            varchar(32) NOT NULL,
  asset_scope         json NOT NULL,
  trade_date_from     date NULL,
  trade_date_to       date NULL,
  source_key          varchar(64) NOT NULL,
  batch_size          int NOT NULL DEFAULT 200,
  universe_snapshot_id varchar(64) NOT NULL,
  status              varchar(32) NOT NULL,
  current_stage       varchar(32) NOT NULL,
  stage_progress_json json NULL,
  summary_json        json NULL,
  error_message       varchar(1024) NULL,
  created_by          varchar(64) NULL,
  created_at          datetime NOT NULL,
  updated_at          datetime NOT NULL,
  finished_at         datetime NULL,
  INDEX idx_market_backfill_run_status (status, current_stage, created_at),
  INDEX idx_market_backfill_scheduler_run (scheduler_run_id),
  CONSTRAINT fk_market_backfill_runs_snapshot
    FOREIGN KEY (universe_snapshot_id) REFERENCES market_universe_snapshots(id)
);

CREATE TABLE IF NOT EXISTS market_backfill_run_details (
  id                varchar(64) PRIMARY KEY,
  run_id            varchar(64) NOT NULL,
  scheduler_run_id  varchar(64) NULL,
  stage             varchar(32) NOT NULL,
  asset_type        varchar(32) NULL,
  batch_key         varchar(128) NULL,
  source_key        varchar(64) NULL,
  symbol_count      int NOT NULL DEFAULT 0,
  symbol_sample     json NULL,
  trade_date_from   date NULL,
  trade_date_to     date NULL,
  status            varchar(32) NOT NULL,
  fetched_count     int NOT NULL DEFAULT 0,
  upserted_count    int NOT NULL DEFAULT 0,
  truth_count       int NOT NULL DEFAULT 0,
  warning_text      varchar(1024) NULL,
  error_text        varchar(1024) NULL,
  started_at        datetime NOT NULL,
  finished_at       datetime NULL,
  created_at        datetime NOT NULL,
  updated_at        datetime NOT NULL,
  INDEX idx_market_backfill_details_stage (run_id, stage, asset_type, status),
  INDEX idx_market_backfill_details_scheduler (scheduler_run_id, stage, status),
  CONSTRAINT fk_market_backfill_run_details_run
    FOREIGN KEY (run_id) REFERENCES market_backfill_runs(id)
    ON DELETE CASCADE
);

INSERT INTO scheduler_job_definitions
  (id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at)
VALUES
  ('jobdef_market_data_full_backfill', 'market_data_full_backfill', '市场数据全量回填', 'SYSTEM', '0 0 2 * * *', 'DISABLED', NULL, 'system', NOW(), NOW()),
  ('jobdef_market_data_incremental_sync', 'market_data_incremental_sync', '市场数据增量同步', 'SYSTEM', '0 30 6 * * 1-5', 'ACTIVE', NULL, 'system', NOW(), NOW()),
  ('jobdef_market_data_truth_rebuild', 'market_data_truth_rebuild', '市场数据 truth 重建', 'SYSTEM', '0 45 6 * * 1-5', 'ACTIVE', NULL, 'system', NOW(), NOW())
ON DUPLICATE KEY UPDATE
  display_name = VALUES(display_name),
  module = VALUES(module),
  cron_expr = VALUES(cron_expr),
  status = VALUES(status),
  updated_at = VALUES(updated_at);
