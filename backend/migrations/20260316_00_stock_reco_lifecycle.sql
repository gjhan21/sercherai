SET @stock_reco_source_type_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'stock_recommendations'
    AND column_name = 'source_type'
);
SET @stock_reco_source_type_sql := IF(
  @stock_reco_source_type_exists = 0,
  'ALTER TABLE stock_recommendations ADD COLUMN source_type varchar(32) NULL AFTER reason_summary',
  'SELECT 1'
);
PREPARE stock_reco_source_type_stmt FROM @stock_reco_source_type_sql;
EXECUTE stock_reco_source_type_stmt;
DEALLOCATE PREPARE stock_reco_source_type_stmt;

SET @stock_reco_strategy_version_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'stock_recommendations'
    AND column_name = 'strategy_version'
);
SET @stock_reco_strategy_version_sql := IF(
  @stock_reco_strategy_version_exists = 0,
  'ALTER TABLE stock_recommendations ADD COLUMN strategy_version varchar(64) NULL AFTER source_type',
  'SELECT 1'
);
PREPARE stock_reco_strategy_version_stmt FROM @stock_reco_strategy_version_sql;
EXECUTE stock_reco_strategy_version_stmt;
DEALLOCATE PREPARE stock_reco_strategy_version_stmt;

SET @stock_reco_reviewer_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'stock_recommendations'
    AND column_name = 'reviewer'
);
SET @stock_reco_reviewer_sql := IF(
  @stock_reco_reviewer_exists = 0,
  'ALTER TABLE stock_recommendations ADD COLUMN reviewer varchar(64) NULL AFTER strategy_version',
  'SELECT 1'
);
PREPARE stock_reco_reviewer_stmt FROM @stock_reco_reviewer_sql;
EXECUTE stock_reco_reviewer_stmt;
DEALLOCATE PREPARE stock_reco_reviewer_stmt;

SET @stock_reco_publisher_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'stock_recommendations'
    AND column_name = 'publisher'
);
SET @stock_reco_publisher_sql := IF(
  @stock_reco_publisher_exists = 0,
  'ALTER TABLE stock_recommendations ADD COLUMN publisher varchar(64) NULL AFTER reviewer',
  'SELECT 1'
);
PREPARE stock_reco_publisher_stmt FROM @stock_reco_publisher_sql;
EXECUTE stock_reco_publisher_stmt;
DEALLOCATE PREPARE stock_reco_publisher_stmt;

SET @stock_reco_review_note_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'stock_recommendations'
    AND column_name = 'review_note'
);
SET @stock_reco_review_note_sql := IF(
  @stock_reco_review_note_exists = 0,
  'ALTER TABLE stock_recommendations ADD COLUMN review_note varchar(512) NULL AFTER publisher',
  'SELECT 1'
);
PREPARE stock_reco_review_note_stmt FROM @stock_reco_review_note_sql;
EXECUTE stock_reco_review_note_stmt;
DEALLOCATE PREPARE stock_reco_review_note_stmt;

SET @stock_reco_performance_label_exists := (
  SELECT COUNT(*)
  FROM information_schema.columns
  WHERE table_schema = DATABASE()
    AND table_name = 'stock_recommendations'
    AND column_name = 'performance_label'
);
SET @stock_reco_performance_label_sql := IF(
  @stock_reco_performance_label_exists = 0,
  'ALTER TABLE stock_recommendations ADD COLUMN performance_label varchar(16) NULL AFTER review_note',
  'SELECT 1'
);
PREPARE stock_reco_performance_label_stmt FROM @stock_reco_performance_label_sql;
EXECUTE stock_reco_performance_label_stmt;
DEALLOCATE PREPARE stock_reco_performance_label_stmt;
