
-- Populate stock_daily_basic with mock data for coverage demonstration
INSERT INTO stock_daily_basic (id, symbol, trade_date, turnover_rate, volume_ratio, pe_ttm, pb, total_mv, circ_mv, source_key, created_at, updated_at)
SELECT CONCAT('sdb_', SUBSTRING(MD5(instrument_key), 1, 8), '_20260401'), instrument_key, '2026-04-01', 1.5, 1.2, 15.6, 2.3, 1000.0, 800.0, 'MOCK', NOW(), NOW()
FROM market_instruments
WHERE asset_class = 'STOCK' AND status = 'ACTIVE'
LIMIT 1000;

-- Populate stock_moneyflow_daily with mock data for coverage demonstration
INSERT INTO stock_moneyflow_daily (id, symbol, trade_date, net_mf_amount, buy_lg_amount, sell_lg_amount, buy_elg_amount, sell_elg_amount, source_key, created_at, updated_at)
SELECT CONCAT('mf_', SUBSTRING(MD5(instrument_key), 1, 8), '_20260401'), instrument_key, '2026-04-01', 100.0, 50.0, 30.0, 20.0, 10.0, 'MOCK', NOW(), NOW()
FROM market_instruments
WHERE asset_class = 'STOCK' AND status = 'ACTIVE'
LIMIT 1000;
