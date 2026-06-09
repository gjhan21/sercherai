CREATE TABLE IF NOT EXISTS stock_simulated_positions (
    id VARCHAR(32) PRIMARY KEY,
    reco_id VARCHAR(32) NOT NULL, -- 关联 stock_recommendations.id
    symbol VARCHAR(32) NOT NULL,
    name VARCHAR(64) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'HOLDING', -- HOLDING, CLOSED
    open_date DATE NOT NULL,
    open_price DECIMAL(10, 4) NOT NULL,
    current_price DECIMAL(10, 4) NOT NULL,
    close_date DATE,
    close_price DECIMAL(10, 4),
    take_profit_price DECIMAL(10, 4),
    stop_loss_price DECIMAL(10, 4),
    quantity DECIMAL(12, 2) NOT NULL DEFAULT 1000.00, -- 默认模拟买入1000股
    cost_basis DECIMAL(14, 4) NOT NULL,              -- open_price * quantity
    close_value DECIMAL(14, 4),                     -- close_price * quantity
    return_rate DECIMAL(8, 4) NOT NULL DEFAULT 0.0000, -- (current_price - open_price) / open_price
    max_drawdown DECIMAL(8, 4) NOT NULL DEFAULT 0.0000, -- 持有期间历史最大回撤
    hold_days INT NOT NULL DEFAULT 0,
    close_reason VARCHAR(32), -- TAKE_PROFIT, STOP_LOSS, EXPIRED, MANUAL
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_status (status),
    KEY idx_symbol (symbol)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
