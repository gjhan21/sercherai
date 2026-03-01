CREATE TABLE IF NOT EXISTS stock_market_quotes (
  id               varchar(32) PRIMARY KEY,
  symbol           varchar(16) NOT NULL,
  trade_date       date NOT NULL,
  open_price       decimal(14,4) NOT NULL,
  high_price       decimal(14,4) NOT NULL,
  low_price        decimal(14,4) NOT NULL,
  close_price      decimal(14,4) NOT NULL,
  prev_close_price decimal(14,4),
  volume           bigint NOT NULL DEFAULT 0,
  turnover         decimal(20,2) NOT NULL DEFAULT 0,
  source_key       varchar(64) NOT NULL,
  created_at       datetime NOT NULL,
  updated_at       datetime NOT NULL,
  UNIQUE KEY uk_symbol_trade_date (symbol, trade_date),
  INDEX idx_trade_date (trade_date),
  INDEX idx_symbol_trade_date (symbol, trade_date)
);
