-- Migration: Add market provider governance extensions for stock factors/news/list data domains.
-- Date: 2026-07-06

INSERT INTO market_provider_capabilities (
    provider_key, asset_class, data_kind, supports_sync, supports_truth_rebuild, supports_context_seed,
    supports_research_run, supports_backfill, supports_batch, supports_intraday, supports_history,
    supports_metadata_enrichment, requires_auth, fallback_allowed, priority_weight
)
VALUES
    ('TUSHARE', 'STOCK', 'STOCK_DAILY_BASIC', 1, 0, 1, 1, 1, 1, 0, 1, 0, 1, 1, 100),
    ('AKSHARE', 'STOCK', 'STOCK_DAILY_BASIC', 1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 90),
    ('MOCK', 'STOCK', 'STOCK_DAILY_BASIC', 1, 0, 1, 1, 1, 1, 0, 1, 0, 0, 1, 10),
    
    ('TUSHARE', 'STOCK', 'STOCK_MONEYFLOW', 1, 0, 1, 1, 1, 1, 0, 1, 0, 1, 1, 100),
    ('AKSHARE', 'STOCK', 'STOCK_MONEYFLOW', 1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 90),
    ('MOCK', 'STOCK', 'STOCK_MONEYFLOW', 1, 0, 1, 1, 1, 1, 0, 1, 0, 0, 1, 10),
    
    ('AKSHARE', 'STOCK', 'STOCK_NEWS_RAW', 1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 100),
    ('TUSHARE', 'STOCK', 'STOCK_NEWS_RAW', 1, 0, 1, 1, 1, 1, 0, 1, 0, 1, 1, 80),
    
    ('TUSHARE', 'STOCK', 'STOCK_KPL_LIST', 1, 0, 1, 1, 1, 1, 0, 1, 0, 1, 1, 100),
    ('AKSHARE', 'STOCK', 'STOCK_KPL_LIST', 1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 80),
    
    ('TUSHARE', 'STOCK', 'STOCK_TOP_LIST', 1, 0, 1, 1, 1, 1, 0, 1, 0, 1, 1, 100),
    ('AKSHARE', 'STOCK', 'STOCK_TOP_LIST', 1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 90)
ON DUPLICATE KEY UPDATE
    supports_sync = VALUES(supports_sync),
    supports_truth_rebuild = VALUES(supports_truth_rebuild),
    supports_context_seed = VALUES(supports_context_seed),
    supports_research_run = VALUES(supports_research_run),
    supports_backfill = VALUES(supports_backfill),
    supports_batch = VALUES(supports_batch),
    supports_intraday = VALUES(supports_intraday),
    supports_history = VALUES(supports_history),
    supports_metadata_enrichment = VALUES(supports_metadata_enrichment),
    requires_auth = VALUES(requires_auth),
    fallback_allowed = VALUES(fallback_allowed),
    priority_weight = VALUES(priority_weight);

INSERT INTO market_provider_routing_policies (
    policy_key, asset_class, data_kind, primary_provider_key, fallback_provider_keys_json, fallback_allowed, mock_allowed, quality_threshold
)
VALUES
    ('market.stock.daily_basic', 'STOCK', 'STOCK_DAILY_BASIC', 'TUSHARE', JSON_ARRAY('AKSHARE', 'MOCK'), 1, 1, 0.7500),
    ('market.stock.moneyflow', 'STOCK', 'STOCK_MONEYFLOW', 'TUSHARE', JSON_ARRAY('AKSHARE', 'MOCK'), 1, 1, 0.7500),
    ('market.stock.news_raw', 'STOCK', 'STOCK_NEWS_RAW', 'AKSHARE', JSON_ARRAY('TUSHARE'), 1, 0, 0.7000),
    ('market.stock.kpl_list', 'STOCK', 'STOCK_KPL_LIST', 'TUSHARE', JSON_ARRAY('AKSHARE'), 1, 0, 0.7000),
    ('market.stock.top_list', 'STOCK', 'STOCK_TOP_LIST', 'TUSHARE', JSON_ARRAY('AKSHARE'), 1, 0, 0.7000)
ON DUPLICATE KEY UPDATE
    primary_provider_key = VALUES(primary_provider_key),
    fallback_provider_keys_json = VALUES(fallback_provider_keys_json),
    fallback_allowed = VALUES(fallback_allowed),
    mock_allowed = VALUES(mock_allowed),
    quality_threshold = VALUES(quality_threshold);
