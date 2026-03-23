CREATE TABLE IF NOT EXISTS market_provider_registry (
    provider_key VARCHAR(64) PRIMARY KEY,
    provider_name VARCHAR(128) NOT NULL,
    provider_type VARCHAR(32) NOT NULL DEFAULT 'API',
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE',
    auth_mode VARCHAR(32) NOT NULL DEFAULT 'NONE',
    endpoint VARCHAR(512) NOT NULL DEFAULT '',
    timeout_ms INT NOT NULL DEFAULT 10000,
    retry_policy_json JSON NULL,
    health_policy_json JSON NULL,
    rate_limit_policy_json JSON NULL,
    cost_tier VARCHAR(32) NOT NULL DEFAULT 'FREE',
    supports_truth_write TINYINT(1) NOT NULL DEFAULT 0,
    supports_manual_sync TINYINT(1) NOT NULL DEFAULT 1,
    supports_auto_sync TINYINT(1) NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS market_provider_capabilities (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    provider_key VARCHAR(64) NOT NULL,
    asset_class VARCHAR(32) NOT NULL DEFAULT '',
    data_kind VARCHAR(64) NOT NULL,
    supports_sync TINYINT(1) NOT NULL DEFAULT 0,
    supports_truth_rebuild TINYINT(1) NOT NULL DEFAULT 0,
    supports_context_seed TINYINT(1) NOT NULL DEFAULT 0,
    supports_research_run TINYINT(1) NOT NULL DEFAULT 0,
    supports_backfill TINYINT(1) NOT NULL DEFAULT 0,
    supports_batch TINYINT(1) NOT NULL DEFAULT 0,
    supports_intraday TINYINT(1) NOT NULL DEFAULT 0,
    supports_history TINYINT(1) NOT NULL DEFAULT 1,
    supports_metadata_enrichment TINYINT(1) NOT NULL DEFAULT 0,
    requires_auth TINYINT(1) NOT NULL DEFAULT 0,
    fallback_allowed TINYINT(1) NOT NULL DEFAULT 1,
    priority_weight INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_market_provider_capability (provider_key, asset_class, data_kind)
);

CREATE TABLE IF NOT EXISTS market_provider_routing_policies (
    policy_key VARCHAR(128) PRIMARY KEY,
    asset_class VARCHAR(32) NOT NULL DEFAULT '',
    data_kind VARCHAR(64) NOT NULL,
    primary_provider_key VARCHAR(64) NOT NULL,
    fallback_provider_keys_json JSON NULL,
    fallback_allowed TINYINT(1) NOT NULL DEFAULT 1,
    mock_allowed TINYINT(1) NOT NULL DEFAULT 0,
    quality_threshold DECIMAL(6,4) NOT NULL DEFAULT 0.0000,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS market_provider_quality_scores (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    provider_key VARCHAR(64) NOT NULL,
    asset_class VARCHAR(32) NOT NULL DEFAULT '',
    data_kind VARCHAR(64) NOT NULL,
    freshness_score DECIMAL(6,4) NOT NULL DEFAULT 0.0000,
    coverage_score DECIMAL(6,4) NOT NULL DEFAULT 0.0000,
    trust_score DECIMAL(6,4) NOT NULL DEFAULT 0.0000,
    stability_score DECIMAL(6,4) NOT NULL DEFAULT 0.0000,
    overall_score DECIMAL(6,4) NOT NULL DEFAULT 0.0000,
    score_reasons_json JSON NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY uk_market_provider_quality_score (provider_key, asset_class, data_kind)
);

INSERT INTO market_provider_registry (
    provider_key, provider_name, provider_type, status, auth_mode, endpoint, timeout_ms,
    retry_policy_json, health_policy_json, rate_limit_policy_json, cost_tier,
    supports_truth_write, supports_manual_sync, supports_auto_sync
)
VALUES
    ('TUSHARE', 'Tushare Pro', 'API', 'ACTIVE', 'TOKEN', 'https://api.tushare.pro', 12000, JSON_OBJECT('max_retries', 3), JSON_OBJECT('timeout_ms', 3000), JSON_OBJECT('rpm', 120), 'PAID', 1, 1, 1),
    ('AKSHARE', 'AkShare', 'API', 'ACTIVE', 'NONE', 'https://akshare.akfamily.xyz', 12000, JSON_OBJECT('max_retries', 2), JSON_OBJECT('timeout_ms', 3000), JSON_OBJECT('rpm', 90), 'FREE', 1, 1, 1),
    ('TICKERMD', 'TickerMD', 'API', 'ACTIVE', 'TOKEN', 'https://api.tickermd.com', 12000, JSON_OBJECT('max_retries', 2), JSON_OBJECT('timeout_ms', 3000), JSON_OBJECT('rpm', 60), 'PAID', 1, 1, 1),
    ('MYSELF', 'Myself Bridge', 'BRIDGE', 'ACTIVE', 'NONE', 'internal://myself', 8000, JSON_OBJECT('max_retries', 1), JSON_OBJECT('timeout_ms', 2000), JSON_OBJECT('rpm', 120), 'FREE', 1, 1, 1),
    ('MOCK', 'Mock Source', 'INTERNAL', 'ACTIVE', 'NONE', '', 1000, JSON_OBJECT('max_retries', 0), JSON_OBJECT('timeout_ms', 500), JSON_OBJECT('rpm', 9999), 'FREE', 0, 1, 0)
ON DUPLICATE KEY UPDATE
    provider_name = VALUES(provider_name),
    provider_type = VALUES(provider_type),
    status = VALUES(status),
    auth_mode = VALUES(auth_mode),
    endpoint = VALUES(endpoint),
    timeout_ms = VALUES(timeout_ms),
    retry_policy_json = VALUES(retry_policy_json),
    health_policy_json = VALUES(health_policy_json),
    rate_limit_policy_json = VALUES(rate_limit_policy_json),
    cost_tier = VALUES(cost_tier),
    supports_truth_write = VALUES(supports_truth_write),
    supports_manual_sync = VALUES(supports_manual_sync),
    supports_auto_sync = VALUES(supports_auto_sync);

INSERT INTO market_provider_capabilities (
    provider_key, asset_class, data_kind, supports_sync, supports_truth_rebuild, supports_context_seed,
    supports_research_run, supports_backfill, supports_batch, supports_intraday, supports_history,
    supports_metadata_enrichment, requires_auth, fallback_allowed, priority_weight
)
VALUES
    ('TUSHARE', 'STOCK', 'DAILY_BARS', 1, 1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 100),
    ('AKSHARE', 'STOCK', 'DAILY_BARS', 1, 1, 1, 1, 1, 1, 0, 1, 1, 0, 1, 90),
    ('TICKERMD', 'STOCK', 'DAILY_BARS', 1, 1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 80),
    ('MYSELF', 'STOCK', 'DAILY_BARS', 1, 1, 1, 1, 1, 1, 0, 1, 0, 0, 1, 70),
    ('MOCK', 'STOCK', 'DAILY_BARS', 1, 1, 1, 1, 1, 1, 0, 1, 0, 0, 1, 10),
    ('TUSHARE', 'FUTURES', 'DAILY_BARS', 1, 1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 100),
    ('TICKERMD', 'FUTURES', 'DAILY_BARS', 1, 1, 1, 1, 1, 1, 0, 1, 0, 1, 1, 90),
    ('AKSHARE', 'FUTURES', 'DAILY_BARS', 1, 1, 1, 1, 1, 1, 0, 1, 0, 0, 1, 80),
    ('MYSELF', 'FUTURES', 'DAILY_BARS', 1, 1, 1, 1, 1, 1, 0, 1, 0, 0, 1, 75),
    ('MOCK', 'FUTURES', 'DAILY_BARS', 1, 1, 1, 1, 1, 1, 0, 1, 0, 0, 1, 10),
    ('TUSHARE', 'STOCK', 'INSTRUMENT_MASTER', 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 1, 100),
    ('TUSHARE', 'FUTURES', 'INSTRUMENT_MASTER', 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 1, 100),
    ('AKSHARE', '', 'NEWS_ITEMS', 1, 0, 1, 1, 1, 1, 0, 1, 1, 0, 1, 100),
    ('TUSHARE', '', 'NEWS_ITEMS', 1, 0, 1, 1, 1, 1, 0, 1, 0, 1, 1, 80),
    ('TUSHARE', 'FUTURES', 'FUTURES_INVENTORY', 1, 0, 1, 1, 1, 1, 0, 1, 1, 1, 1, 100),
    ('MOCK', 'FUTURES', 'FUTURES_INVENTORY', 1, 0, 1, 1, 1, 1, 0, 1, 0, 0, 1, 10)
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
    ('market.stock.daily', 'STOCK', 'DAILY_BARS', 'TUSHARE', JSON_ARRAY('AKSHARE', 'TICKERMD'), 1, 0, 0.8000),
    ('market.futures.daily', 'FUTURES', 'DAILY_BARS', 'TUSHARE', JSON_ARRAY('TICKERMD', 'AKSHARE'), 1, 0, 0.8000),
    ('market.news', '', 'NEWS_ITEMS', 'AKSHARE', JSON_ARRAY('TUSHARE'), 1, 0, 0.7000),
    ('market.instrument.stock', 'STOCK', 'INSTRUMENT_MASTER', 'TUSHARE', JSON_ARRAY('AKSHARE', 'TICKERMD', 'MYSELF'), 1, 0, 0.8000),
    ('market.instrument.futures', 'FUTURES', 'INSTRUMENT_MASTER', 'TUSHARE', JSON_ARRAY('AKSHARE', 'TICKERMD', 'MYSELF'), 1, 0, 0.8000),
    ('market.futures.inventory', 'FUTURES', 'FUTURES_INVENTORY', 'TUSHARE', JSON_ARRAY('MOCK'), 1, 1, 0.6000)
ON DUPLICATE KEY UPDATE
    primary_provider_key = VALUES(primary_provider_key),
    fallback_provider_keys_json = VALUES(fallback_provider_keys_json),
    fallback_allowed = VALUES(fallback_allowed),
    mock_allowed = VALUES(mock_allowed),
    quality_threshold = VALUES(quality_threshold);
