DELETE q1
FROM market_intraday_quotes q1
JOIN market_intraday_quotes q2
  ON q1.asset_class = q2.asset_class
 AND q1.instrument_key = q2.instrument_key
 AND q1.quote_time = q2.quote_time
 AND q1.source_key = q2.source_key
 AND (
   q1.updated_at < q2.updated_at
   OR (q1.updated_at = q2.updated_at AND q1.fetched_at < q2.fetched_at)
   OR (q1.updated_at = q2.updated_at AND q1.fetched_at = q2.fetched_at AND q1.id < q2.id)
 );

SET @uk_exists := (
  SELECT COUNT(*)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 'market_intraday_quotes'
    AND index_name = 'uk_market_intraday_quote'
);

SET @add_uk_sql := IF(
  @uk_exists = 0,
  'ALTER TABLE market_intraday_quotes ADD UNIQUE KEY uk_market_intraday_quote (asset_class, instrument_key, quote_time, source_key)',
  'SELECT 1'
);

PREPARE add_uk_stmt FROM @add_uk_sql;
EXECUTE add_uk_stmt;
DEALLOCATE PREPARE add_uk_stmt;

