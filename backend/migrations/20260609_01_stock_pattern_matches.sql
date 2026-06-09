CREATE TABLE IF NOT EXISTS stock_pattern_matches (
    id VARCHAR(32) PRIMARY KEY,
    source_symbol VARCHAR(32) NOT NULL,    -- 目标股票代码 (如 300750.SZ)
    match_symbol VARCHAR(32) NOT NULL,     -- 匹配到的历史股票代码
    match_date DATE NOT NULL,              -- 匹配历史片段的结束日期
    similarity DECIMAL(6, 4) NOT NULL,     -- 相似度 (0.85 ~ 1.0000)
    lookback INT NOT NULL DEFAULT 20,      -- 窗口长度 (20 或 60)
    next_7d_returns JSON NOT NULL,         -- 匹配历史之后 7 天的收益率序列 (如 [1.2, -0.5, ...])
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY ukey_src_match (source_symbol, match_symbol, lookback),
    KEY idx_source_symbol (source_symbol)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
