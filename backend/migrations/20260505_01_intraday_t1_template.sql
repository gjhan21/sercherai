-- 注册超短线 T+1 模板到 stock_selection_profile_templates

INSERT INTO stock_selection_profile_templates (
  id,
  template_key,
  name,
  description,
  market_regime_bias,
  is_default,
  status,
  universe_defaults_json,
  seed_defaults_json,
  factor_defaults_json,
  portfolio_defaults_json,
  publish_defaults_json,
  updated_by,
  updated_at,
  created_at
) VALUES (
  'sstpl_intraday_t1',
  'INTRADAY_T1',
  '超短线T+1多策略',
  '今日买入明日卖出，七大策略并行（超跌反弹/缩量止跌/资金背离/均线支撑/阳包阴/涨停板次日/强势回踩），按市场环境动态加权融合。',
  'ROTATION',
  0,
  'ACTIVE',
  '{"universe_scope":"CN_A_ALL","min_listing_days":60,"min_avg_turnover":100000000,"exclude_st":true,"exclude_suspended":true,"price_min":3,"price_max":500,"volatility_max":10}',
  '{"strategy_mode":"intraday","bucket_limit":36,"seed_pool_cap":180}',
  '{"lookback_days":60,"quant_weight":0.60,"event_weight":0.15,"resonance_weight":0.15,"liquidity_risk_weight":0.10}',
  '{"limit":5,"watchlist_limit":5,"max_risk_level":"HIGH","min_score":65,"max_symbol_per_bucket":2,"max_symbols_per_sector":1,"diversify_by_strategy":true,"max_per_strategy":2,"multi_strategy_bonus":5}',
  '{"review_required":true,"allow_auto_publish":false}',
  'system-bootstrap',
  NOW(),
  NOW()
)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  description = VALUES(description),
  market_regime_bias = VALUES(market_regime_bias),
  is_default = VALUES(is_default),
  status = VALUES(status),
  universe_defaults_json = VALUES(universe_defaults_json),
  seed_defaults_json = VALUES(seed_defaults_json),
  factor_defaults_json = VALUES(factor_defaults_json),
  portfolio_defaults_json = VALUES(portfolio_defaults_json),
  publish_defaults_json = VALUES(publish_defaults_json),
  updated_by = VALUES(updated_by),
  updated_at = VALUES(updated_at);
