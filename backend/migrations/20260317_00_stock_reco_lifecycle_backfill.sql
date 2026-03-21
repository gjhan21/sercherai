UPDATE stock_recommendations
SET
  status = CASE
    WHEN COALESCE(NULLIF(TRIM(status), ''), '') = '' THEN 'DRAFT'
    WHEN UPPER(TRIM(status)) = 'ACTIVE' THEN 'PUBLISHED'
    ELSE UPPER(TRIM(status))
  END,
  source_type = CASE
    WHEN LOWER(TRIM(COALESCE(strategy_version, ''))) LIKE 'daily-%' THEN 'SYSTEM'
    WHEN COALESCE(NULLIF(TRIM(source_type), ''), '') = '' THEN 'MANUAL'
    ELSE UPPER(TRIM(source_type))
  END,
  strategy_version = CASE
    WHEN COALESCE(NULLIF(TRIM(strategy_version), ''), '') <> '' THEN TRIM(strategy_version)
    WHEN UPPER(TRIM(COALESCE(source_type, ''))) = 'SYSTEM' THEN 'daily-v1'
    ELSE 'manual-v1'
  END,
  reviewer = NULLIF(TRIM(reviewer), ''),
  publisher = CASE
    WHEN COALESCE(NULLIF(TRIM(publisher), ''), '') <> '' THEN TRIM(publisher)
    WHEN UPPER(TRIM(COALESCE(source_type, ''))) = 'SYSTEM' OR LOWER(TRIM(COALESCE(strategy_version, ''))) LIKE 'daily-%' THEN 'system'
    ELSE 'legacy-import'
  END,
  review_note = NULLIF(TRIM(review_note), ''),
  performance_label = CASE
    WHEN COALESCE(NULLIF(TRIM(performance_label), ''), '') <> '' THEN UPPER(TRIM(performance_label))
    WHEN UPPER(TRIM(COALESCE(status, ''))) = 'HIT_TAKE_PROFIT' THEN 'WIN'
    WHEN UPPER(TRIM(COALESCE(status, ''))) = 'HIT_STOP_LOSS' THEN 'LOSS'
    WHEN UPPER(TRIM(COALESCE(status, ''))) IN ('INVALIDATED', 'REVIEWED') THEN 'FLAT'
    WHEN UPPER(TRIM(COALESCE(source_type, ''))) = 'SYSTEM' OR LOWER(TRIM(COALESCE(strategy_version, ''))) LIKE 'daily-%' THEN 'ESTIMATED'
    ELSE 'PENDING'
  END;
