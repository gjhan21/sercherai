CREATE TABLE IF NOT EXISTS futures_simulated_positions (
    id VARCHAR(32) PRIMARY KEY,
    strategy_id VARCHAR(32) NOT NULL,              -- 关联 futures_strategies.id
    contract VARCHAR(32) NOT NULL,                 -- 期货合约代码 (如 IF2606)
    name VARCHAR(128) NOT NULL,                    -- 合约名称
    direction VARCHAR(16) NOT NULL DEFAULT 'LONG', -- 交易方向 (LONG/SHORT)
    status VARCHAR(16) NOT NULL DEFAULT 'HOLDING', -- 状态 (HOLDING/CLOSED)
    open_date DATE NOT NULL,                       -- 建仓日期
    open_price DECIMAL(12, 4) NOT NULL,            -- 建仓价
    current_price DECIMAL(12, 4) NOT NULL,         -- 现价/结算价
    close_date DATE,                               -- 平仓日期
    close_price DECIMAL(12, 4),                    -- 平仓价
    take_profit_price DECIMAL(12, 4),              -- 止盈触发价
    stop_loss_price DECIMAL(12, 4),                -- 止损触发价
    quantity DECIMAL(12, 2) NOT NULL DEFAULT 10.00,-- 模拟交易量 (默认 10 手)
    cost_basis DECIMAL(16, 4) NOT NULL,            -- 初始持仓价值 (open_price * quantity)
    close_value DECIMAL(16, 4),                    -- 平仓结余价值
    return_rate DECIMAL(8, 4) NOT NULL DEFAULT 0.0000, -- 收益率
    max_drawdown DECIMAL(8, 4) NOT NULL DEFAULT 0.0000, -- 最大回撤率
    hold_days INT NOT NULL DEFAULT 0,              -- 持有天数
    close_reason VARCHAR(32),                      -- 平仓原因 (TAKE_PROFIT, STOP_LOSS, EXPIRED, MANUAL)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_status (status),
    KEY idx_contract (contract)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
