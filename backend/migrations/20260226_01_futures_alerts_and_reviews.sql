-- Futures alerts and review tables for MySQL 8.x

CREATE TABLE IF NOT EXISTS futures_alerts (
  id         varchar(32) PRIMARY KEY,
  user_id    varchar(32) NOT NULL,
  contract   varchar(32) NOT NULL,
  alert_type varchar(16) NOT NULL,
  threshold  decimal(10,2),
  status     varchar(16) NOT NULL,
  created_at datetime NOT NULL,
  INDEX idx_user_contract (user_id, contract),
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS futures_reviews (
  id           varchar(32) PRIMARY KEY,
  strategy_id  varchar(32) NOT NULL,
  hit_rate     decimal(6,3),
  pnl          decimal(10,2),
  max_drawdown decimal(6,3),
  review_date  datetime NOT NULL,
  INDEX idx_strategy_review (strategy_id, review_date),
  FOREIGN KEY (strategy_id) REFERENCES futures_strategies(id)
);
