-- 超短线 T+1 选股数据表
-- stock_limit_up_daily: 涨停板榜单 (kpl_list 接口)
-- stock_top_list_daily:  龙虎榜每日明细 (top_list 接口)

CREATE TABLE IF NOT EXISTS stock_limit_up_daily (
  id varchar(64) NOT NULL,
  trade_date date NOT NULL,
  ts_code varchar(32) NOT NULL,
  name varchar(64) NOT NULL DEFAULT '',
  lu_time varchar(8) DEFAULT '',
  open_time varchar(8) DEFAULT '',
  last_time varchar(8) DEFAULT '',
  tag varchar(32) DEFAULT '',
  theme varchar(128) DEFAULT '',
  status varchar(32) DEFAULT '',
  limit_order decimal(18,2) DEFAULT 0,
  lu_limit_order decimal(18,2) DEFAULT 0,
  bid_amount decimal(18,2) DEFAULT 0,
  bid_change decimal(18,2) DEFAULT 0,
  lu_desc varchar(512) DEFAULT '',
  pct_chg decimal(10,4) DEFAULT 0,
  close decimal(18,6) DEFAULT 0,
  amount decimal(18,2) DEFAULT 0,
  float_mv decimal(18,2) DEFAULT 0,
  turnover_rate decimal(10,4) DEFAULT 0,
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_stock_limit_up_daily_date_code (trade_date, ts_code),
  KEY idx_stock_limit_up_daily_date (trade_date),
  KEY idx_stock_limit_up_daily_theme (trade_date, theme)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS stock_top_list_daily (
  id varchar(64) NOT NULL,
  trade_date date NOT NULL,
  ts_code varchar(32) NOT NULL,
  name varchar(64) NOT NULL DEFAULT '',
  close decimal(18,6) DEFAULT 0,
  pct_chg decimal(10,4) DEFAULT 0,
  amount decimal(18,2) DEFAULT 0,
  buy_amount decimal(18,2) DEFAULT 0,
  sell_amount decimal(18,2) DEFAULT 0,
  net_amount decimal(18,2) DEFAULT 0,
  reason varchar(256) DEFAULT '',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_stock_top_list_daily_date_code (trade_date, ts_code),
  KEY idx_stock_top_list_daily_date (trade_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
