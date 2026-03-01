SET @cover_col_exists = (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'news_articles'
    AND COLUMN_NAME = 'cover_url'
);

SET @cover_col_sql = IF(
  @cover_col_exists = 0,
  'ALTER TABLE news_articles ADD COLUMN cover_url varchar(512) NULL AFTER content',
  'SELECT 1'
);

PREPARE stmt FROM @cover_col_sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
