#!/usr/bin/env python3
"""超短线 T+1 回测烟幕测试 —— 使用合成数据验证 7-策略管道端到端运行。

用途: 在无 Go 后端 / 无 Tushare 数据 / 无 pydantic 的环境下验证回测引擎和策略管道。
生成 30 只合成 A 股的 90 个交易日数据，每日价格随机偏移 0-2%，
执行完整流程: MarketSeed 生成 → 7策略并行挖掘 → 跨策略融合 → 组合风控 → T+1 收益计算 → 统计。

用法:
    cd services/strategy-engine
    python3 -m app.tools.smoke_test_backtest --days 90 --seeds 30 --format summary

此脚本仅依赖标准库 + app.domain (纯 dataclasses)，无需 pydantic/fastapi/uvicorn。
"""
from __future__ import annotations

import argparse
import logging
import random
import sys
from collections import defaultdict
from dataclasses import dataclass, field
from datetime import date, timedelta
from typing import Any

# ——— Bootstrap: add parent to path, shim datetime.UTC for Python 3.10 ———
import os as _os

_PARENT = _os.path.dirname(_os.path.dirname(_os.path.dirname(_os.path.abspath(__file__))))
if _PARENT not in sys.path:
    sys.path.insert(0, _PARENT)

import datetime as _dt
if not hasattr(_dt, "UTC"):
    _dt.UTC = _dt.timezone.utc  # type: ignore

# 导入纯标准库域模块（已验证无 pydantic 依赖）
from app.domain.models import MarketSeed, MarketSeedLoadResult, StockFeature  # noqa: E402
from app.domain.seeds.intraday_seed_miner import IntradaySeedMiner  # noqa: E402
from app.domain.decision.intraday_decision_fusion import IntradayDecisionFusion  # noqa: E402

logger = logging.getLogger(__name__)

# ————————————————————————————————————————————
# 内联类型（避免导入依赖 pydantic 的模块）
# ————————————————————————————————————————————

RISK_ORDER = {"LOW": 0, "MEDIUM": 1, "HIGH": 2}


@dataclass(slots=True)
class TradeRecord:
    symbol: str
    name: str
    trade_date: str
    strategy: str
    score: float
    portfolio_role: str
    buy_price: float
    sell_price: float
    return_pct: float
    is_win: bool
    regime: str
    is_multi_strategy: bool


@dataclass(slots=True)
class StrategyStats:
    strategy_name: str
    total_signals: int = 0
    wins: int = 0
    hit_rate: float = 0.0
    avg_return: float = 0.0
    total_return: float = 0.0
    max_return: float = 0.0
    min_return: float = 0.0
    avg_score: float = 0.0
    regime_stats: dict[str, dict[str, float]] = field(default_factory=dict)


@dataclass(slots=True)
class RegimeStats:
    regime: str
    total_signals: int = 0
    wins: int = 0
    hit_rate: float = 0.0
    avg_return: float = 0.0
    strategy_contributions: dict[str, int] = field(default_factory=dict)


@dataclass(slots=True)
class BacktestResult:
    start_date: str
    end_date: str
    trading_days: int
    active_days: int
    total_trades: int
    win_trades: int
    overall_hit_rate: float
    overall_avg_return: float
    overall_cumulative_return: float
    max_drawdown: float
    best_day: str
    best_day_return: float
    worst_day: str
    worst_day_return: float
    multi_strategy_trades: int
    multi_strategy_hit_rate: float
    multi_strategy_avg_return: float
    strategy_stats: dict[str, StrategyStats]
    regime_stats: dict[str, RegimeStats]
    daily_returns: list[dict[str, Any]]
    trade_records: list[TradeRecord]
    warnings: list[str]


@dataclass(slots=True)
class _GuardResult:
    portfolio: list[StockFeature]
    watchlist: list[StockFeature]
    warnings: list[str]


# ————————————————————————————————————————————
# 轻量 PortfolioGuard（替代 pydantic 版本）
# ————————————————————————————————————————————


class PortfolioGuard:
    """简化版组合风控，逻辑与 app.domain.risk.portfolio_guard 一致。"""

    def apply(self, features: list[StockFeature], max_risk: str = "MEDIUM", min_score: float = 65.0,
              limit: int = 5, max_per_strategy: int = 2, max_per_sector: int = 1) -> _GuardResult:
        warnings: list[str] = []

        # Risk gate
        max_risk_ord = RISK_ORDER[max_risk]
        allowed = [f for f in features if RISK_ORDER.get(f.risk_level, 1) <= max_risk_ord]
        if not allowed:
            warnings.append("按风险等级过滤后无结果，回退到原始排序候选。")
            allowed = list(features)

        # Score gate
        scored = [f for f in allowed if f.score >= min_score]
        if scored:
            allowed = scored[:max(limit, len(scored))]

        # Strategy-level diversification cap (INTRADAY_T1 specific)
        strategy_counts: dict[str, int] = defaultdict(int)
        sector_counts: dict[str, int] = defaultdict(int)
        portfolio: list[StockFeature] = []
        watchlist: list[StockFeature] = []

        for feat in allowed:
            strat = feat.strategy_name or "unknown"
            sector = feat.sector or "other"
            if len(portfolio) < limit:
                if strategy_counts[strat] < max_per_strategy and sector_counts[sector] < max_per_sector:
                    portfolio.append(feat)
                    strategy_counts[strat] += 1
                    sector_counts[sector] += 1
            else:
                watchlist.append(feat)

        return _GuardResult(portfolio=portfolio, watchlist=watchlist, warnings=warnings)


# ————————————————————————————————————————————
# 市场环境检测
# ————————————————————————————————————————————

def detect_regime(seeds: list[MarketSeed]) -> str:
    if not seeds:
        return "UNKNOWN"
    n = len(seeds)
    avg_m5 = sum(s.momentum5 for s in seeds) / n
    avg_m20 = sum(s.momentum20 for s in seeds) / n
    avg_vol20 = sum(s.volatility20 for s in seeds) / n
    if avg_vol20 > 12:
        return "RISK_OFF"
    if avg_m20 < -5 and avg_vol20 > 8:
        return "DEFENSIVE"
    if avg_m5 > 3 and avg_m20 > 0:
        return "UPTREND"
    if 2 < abs(avg_m5) < 6:
        return "ROTATION"
    if avg_m5 > 4 and avg_vol20 > 8:
        return "EVENT_DRIVEN"
    return "ROTATION"


# ————————————————————————————————————————————
# 合成数据
# ————————————————————————————————————————————

STOCK_POOL: list[dict[str, Any]] = [
    {"symbol": "600519.SH", "name": "贵州茅台", "sector": "白酒", "industry": "食品饮料",
     "pe_ttm": 26.8, "pb": 9.6, "listing_days": 8000, "avg_turnover20": 3_500_000_000,
     "theme_tags": ["白酒", "消费龙头", "权重股"]},
    {"symbol": "601318.SH", "name": "中国平安", "sector": "保险", "industry": "金融",
     "pe_ttm": 10.5, "pb": 1.2, "listing_days": 6000, "avg_turnover20": 2_800_000_000,
     "theme_tags": ["保险", "金融科技", "权重股"]},
    {"symbol": "600036.SH", "name": "招商银行", "sector": "银行", "industry": "金融",
     "pe_ttm": 8.9, "pb": 0.9, "listing_days": 7000, "avg_turnover20": 2_200_000_000,
     "theme_tags": ["银行", "零售银行", "权重股"]},
    {"symbol": "300750.SZ", "name": "宁德时代", "sector": "锂电池", "industry": "新能源",
     "pe_ttm": 38.2, "pb": 7.6, "listing_days": 1500, "avg_turnover20": 5_500_000_000,
     "theme_tags": ["锂电池", "新能源", "创业板龙头"]},
    {"symbol": "000333.SZ", "name": "美的集团", "sector": "家电", "industry": "家用电器",
     "pe_ttm": 13.2, "pb": 2.7, "listing_days": 9000, "avg_turnover20": 1_800_000_000,
     "theme_tags": ["家电", "消费", "深股通"]},
    {"symbol": "688981.SH", "name": "中芯国际", "sector": "半导体", "industry": "电子",
     "pe_ttm": 58.6, "pb": 5.1, "listing_days": 800, "avg_turnover20": 3_200_000_000,
     "theme_tags": ["半导体", "芯片", "科创板龙头"]},
    {"symbol": "002594.SZ", "name": "比亚迪", "sector": "新能源汽车", "industry": "汽车",
     "pe_ttm": 31.5, "pb": 6.4, "listing_days": 4000, "avg_turnover20": 6_200_000_000,
     "theme_tags": ["新能源汽车", "电池", "深股通"]},
    {"symbol": "601012.SH", "name": "隆基绿能", "sector": "光伏", "industry": "新能源",
     "pe_ttm": 42.1, "pb": 2.2, "listing_days": 3000, "avg_turnover20": 3_800_000_000,
     "theme_tags": ["光伏", "新能源", "权重股"]},
    {"symbol": "600276.SH", "name": "恒瑞医药", "sector": "创新药", "industry": "医药",
     "pe_ttm": 34.2, "pb": 6.0, "listing_days": 7500, "avg_turnover20": 1_500_000_000,
     "theme_tags": ["创新药", "医药", "沪股通"]},
    {"symbol": "601888.SH", "name": "中国中免", "sector": "免税", "industry": "商贸零售",
     "pe_ttm": 51.0, "pb": 4.4, "listing_days": 3500, "avg_turnover20": 2_100_000_000,
     "theme_tags": ["免税", "旅游", "沪股通"]},
    {"symbol": "000858.SZ", "name": "五粮液", "sector": "白酒", "industry": "食品饮料",
     "pe_ttm": 22.5, "pb": 5.8, "listing_days": 9000, "avg_turnover20": 4_200_000_000,
     "theme_tags": ["白酒", "消费", "深股通"]},
    {"symbol": "002415.SZ", "name": "海康威视", "sector": "安防", "industry": "信息技术",
     "pe_ttm": 28.7, "pb": 4.5, "listing_days": 4500, "avg_turnover20": 2_600_000_000,
     "theme_tags": ["安防", "AI", "深股通"]},
    {"symbol": "000651.SZ", "name": "格力电器", "sector": "家电", "industry": "家用电器",
     "pe_ttm": 8.5, "pb": 1.8, "listing_days": 10000, "avg_turnover20": 2_000_000_000,
     "theme_tags": ["家电", "消费", "深股通"]},
    {"symbol": "600900.SH", "name": "长江电力", "sector": "水电", "industry": "公用事业",
     "pe_ttm": 25.1, "pb": 3.2, "listing_days": 8000, "avg_turnover20": 1_200_000_000,
     "theme_tags": ["水电", "高股息", "权重股"]},
    {"symbol": "300059.SZ", "name": "东方财富", "sector": "券商", "industry": "金融",
     "pe_ttm": 35.4, "pb": 5.6, "listing_days": 3500, "avg_turnover20": 8_500_000_000,
     "theme_tags": ["券商", "互联网金融", "创业板龙头"]},
    {"symbol": "601899.SH", "name": "紫金矿业", "sector": "黄金", "industry": "有色金属",
     "pe_ttm": 18.9, "pb": 3.8, "listing_days": 5000, "avg_turnover20": 3_400_000_000,
     "theme_tags": ["黄金", "有色金属", "沪股通"]},
    {"symbol": "002230.SZ", "name": "科大讯飞", "sector": "AI", "industry": "信息技术",
     "pe_ttm": 120.5, "pb": 8.2, "listing_days": 4500, "avg_turnover20": 3_000_000_000,
     "theme_tags": ["AI", "语音识别", "深股通"]},
    {"symbol": "688111.SH", "name": "金山办公", "sector": "SaaS", "industry": "信息技术",
     "pe_ttm": 85.3, "pb": 12.5, "listing_days": 700, "avg_turnover20": 1_800_000_000,
     "theme_tags": ["SaaS", "办公", "科创板"]},
    {"symbol": "300124.SZ", "name": "汇川技术", "sector": "工控", "industry": "机械设备",
     "pe_ttm": 42.1, "pb": 7.8, "listing_days": 4000, "avg_turnover20": 1_500_000_000,
     "theme_tags": ["工控", "机器人", "创业板"]},
    {"symbol": "002475.SZ", "name": "立讯精密", "sector": "消费电子", "industry": "电子",
     "pe_ttm": 25.3, "pb": 4.2, "listing_days": 3500, "avg_turnover20": 3_600_000_000,
     "theme_tags": ["消费电子", "苹果链", "深股通"]},
    {"symbol": "600809.SH", "name": "山西汾酒", "sector": "白酒", "industry": "食品饮料",
     "pe_ttm": 32.1, "pb": 9.2, "listing_days": 9500, "avg_turnover20": 2_400_000_000,
     "theme_tags": ["白酒", "消费", "沪股通"]},
    {"symbol": "603259.SH", "name": "药明康德", "sector": "CXO", "industry": "医药",
     "pe_ttm": 28.4, "pb": 4.1, "listing_days": 1800, "avg_turnover20": 2_200_000_000,
     "theme_tags": ["CXO", "医药外包", "沪股通"]},
    {"symbol": "002714.SZ", "name": "牧原股份", "sector": "养猪", "industry": "农林牧渔",
     "pe_ttm": 15.2, "pb": 2.8, "listing_days": 3000, "avg_turnover20": 1_900_000_000,
     "theme_tags": ["养猪", "农林牧渔", "深股通"]},
    {"symbol": "601857.SH", "name": "中国石油", "sector": "石油", "industry": "采掘",
     "pe_ttm": 12.3, "pb": 0.9, "listing_days": 5000, "avg_turnover20": 1_400_000_000,
     "theme_tags": ["石油", "央企", "权重股"]},
    {"symbol": "000725.SZ", "name": "京东方A", "sector": "面板", "industry": "电子",
     "pe_ttm": 20.4, "pb": 1.1, "listing_days": 6000, "avg_turnover20": 2_500_000_000,
     "theme_tags": ["面板", "OLED", "深股通"]},
    {"symbol": "002371.SZ", "name": "北方华创", "sector": "半导体设备", "industry": "电子",
     "pe_ttm": 68.2, "pb": 9.8, "listing_days": 3500, "avg_turnover20": 2_800_000_000,
     "theme_tags": ["半导体设备", "芯片", "深股通"]},
    {"symbol": "300274.SZ", "name": "阳光电源", "sector": "逆变器", "industry": "新能源",
     "pe_ttm": 35.6, "pb": 8.4, "listing_days": 3000, "avg_turnover20": 4_500_000_000,
     "theme_tags": ["逆变器", "光伏", "储能"]},
    {"symbol": "601225.SH", "name": "陕西煤业", "sector": "煤炭", "industry": "采掘",
     "pe_ttm": 6.5, "pb": 1.4, "listing_days": 3000, "avg_turnover20": 1_100_000_000,
     "theme_tags": ["煤炭", "高股息", "沪股通"]},
    {"symbol": "600030.SH", "name": "中信证券", "sector": "券商", "industry": "金融",
     "pe_ttm": 16.8, "pb": 1.3, "listing_days": 8000, "avg_turnover20": 3_200_000_000,
     "theme_tags": ["券商", "龙头", "权重股"]},
    {"symbol": "300498.SZ", "name": "温氏股份", "sector": "养猪", "industry": "农林牧渔",
     "pe_ttm": 18.5, "pb": 3.1, "listing_days": 2000, "avg_turnover20": 1_600_000_000,
     "theme_tags": ["养猪", "农业", "创业板"]},
]


def make_seed(tmpl: dict[str, Any], trade_date: str, price: float, rng: random.Random) -> MarketSeed:
    # 约 30% 的股票处于"调整中"状态（更容易触发超跌反弹/缩量止跌/均线支撑等策略）
    is_correcting = rng.random() < 0.30

    if is_correcting:
        # 处于调整状态：连跌、缩量、K线偏弱
        m5 = round(rng.uniform(-6, 1), 2)
        m20 = round(rng.uniform(-5, 10), 2)
        vol20 = round(rng.uniform(2, 10), 2)
        m1 = round(rng.uniform(-4, 2), 2)
        m2 = round(rng.uniform(-6, 1), 2)
        m3 = round(rng.uniform(-10, -3), 2)  # 近3日下跌，触发超跌/均线支撑
        cd = rng.randint(2, 4)  # 连跌 2-4 天
        body = round(rng.uniform(-5, 1), 2)
        ls = round(rng.uniform(0.5, 5), 2)  # 下影线较长（多头承接）
        us = round(rng.uniform(0, 2), 2)
        volume_ratio = round(rng.uniform(0.25, 0.55), 2)  # 缩量
        volume_20d_min = rng.choice([1, 1, 1, 2, 3])  # 大概率是 20 日最低量
        dev20 = round(rng.uniform(-1.5, 1.5), 2)  # 贴近 MA20
        trend_strength = round(rng.uniform(-1, 8), 2)
    else:
        # 正常状态：涨跌随机
        m5 = round(rng.uniform(-8, 12), 2)
        m20 = round(rng.uniform(-12, 18), 2)
        vol20 = round(rng.uniform(2, 16), 2)
        m1 = round(rng.uniform(-5, 6), 2)
        m2 = round(rng.uniform(-8, 10), 2)
        m3 = round(rng.uniform(-12, 14), 2)
        cd = rng.randint(0, 4)
        body = round(rng.uniform(-6, 6), 2)
        ls = round(rng.uniform(0, 5), 2)
        us = round(rng.uniform(0, 4), 2)
        volume_ratio = round(rng.uniform(0.3, 3.5), 2)
        volume_20d_min = rng.randint(1, 20)
        dev20 = round(rng.uniform(-18, 18), 2)
        trend_strength = round(rng.uniform(-5, 10), 2)

    is_lu = rng.random() < 0.08
    on_top = rng.random() < 0.05

    return MarketSeed(
        symbol=tmpl["symbol"], name=tmpl["name"], trade_date=trade_date,
        close_price=round(price, 2), momentum5=m5, momentum20=m20, volatility20=vol20,
        volume_ratio=volume_ratio,
        drawdown20=round(rng.uniform(-20, 0), 2),
        trend_strength=trend_strength,
        net_mf_amount=round(rng.uniform(-50_000_000, 200_000_000), 0),
        pe_ttm=tmpl["pe_ttm"], pb=tmpl["pb"],
        turnover_rate=round(rng.uniform(0.2, 8.0), 2),
        news_heat=rng.randint(0, 8),
        positive_news_rate=round(rng.uniform(0.3, 0.9), 2),
        listing_days=tmpl["listing_days"], avg_turnover20=tmpl["avg_turnover20"],
        industry=tmpl["industry"], sector=tmpl["sector"],
        theme_tags=list(tmpl["theme_tags"]),
        momentum1=m1, momentum2=m2, momentum3=m3,
        consecutive_down_days=cd, candle_body_pct=body,
        lower_shadow_pct=ls, upper_shadow_pct=us,
        is_bullish=m1 > 0 and body > 0,
        is_doji=abs(body) < 0.3,
        is_engulfing_bullish=m1 > 0 and body > 0 and rng.random() < 0.2,
        deviation_ma5=round(rng.uniform(-8, 8), 2),
        deviation_ma10=round(rng.uniform(-12, 12), 2),
        deviation_ma20=dev20,
        deviation_ma60=round(rng.uniform(-25, 25), 2),
        volume_20d_min_rank=volume_20d_min,
        volume_contraction_days=rng.randint(0, 7),
        is_limit_up=is_lu,
        lu_time_rank=rng.randint(1, 10) if is_lu else 0,
        seal_order_ratio=round(rng.uniform(0.5, 1.0), 2) if is_lu else 0.0,
        limit_up_days=rng.randint(1, 3) if is_lu else 0,
        is_opened=rng.random() < 0.3 if is_lu else False,
        is_natural_limit=rng.random() < 0.85,
        on_top_list=on_top,
        top_net_amount=round(rng.uniform(5_000_000, 500_000_000), 0) if on_top else 0.0,
        top_buy_sell_ratio=round(rng.uniform(0.8, 3.0), 2) if on_top else 0.0,
    )


# ————————————————————————————————————————————
# 合成 SeedLoader
# ————————————————————————————————————————————

class SyntheticSeedLoader:
    def __init__(self, pool: list[dict[str, Any]], rng_seed: int = 42) -> None:
        self._pool = pool
        self._rng = random.Random(rng_seed)
        self._prices: dict[str, float] = {
            item["symbol"]: round(self._rng.uniform(5, 300), 2) for item in pool
        }
        self._call_count = 0

    def load(self, trade_date: str) -> MarketSeedLoadResult:
        self._call_count += 1
        for sym in self._prices:
            self._prices[sym] = round(self._prices[sym] * (1 + self._rng.uniform(-0.02, 0.02)), 2)
        seeds = [make_seed(tmpl, trade_date, self._prices[tmpl["symbol"]], self._rng) for tmpl in self._pool]
        return MarketSeedLoadResult(seeds=seeds, meta={"source": "synthetic-smoke-test"})


# ————————————————————————————————————————————
# 日期工具
# ————————————————————————————————————————————

def trading_dates_between(start_str: str, end_str: str) -> list[str]:
    start = date.fromisoformat(start_str)
    end = date.fromisoformat(end_str)
    dates: list[str] = []
    d = start
    while d <= end:
        if d.weekday() < 5:
            dates.append(d.isoformat())
        d += timedelta(days=1)
    return dates


def next_trading_date(td: str) -> str:
    d = date.fromisoformat(td) + timedelta(days=1)
    while d.weekday() >= 5:
        d += timedelta(days=1)
    return d.isoformat()


# ————————————————————————————————————————————
# 统计计算（与 intraday_backtest.py 一致）
# ————————————————————————————————————————————

def compute_strategy_stats(trades: list[TradeRecord]) -> dict[str, StrategyStats]:
    by_strat: dict[str, list[TradeRecord]] = defaultdict(list)
    for t in trades:
        by_strat[t.strategy].append(t)
    result: dict[str, StrategyStats] = {}
    for name, group in by_strat.items():
        wins = sum(1 for t in group if t.is_win)
        returns = [t.return_pct for t in group]
        scores = [t.score for t in group]
        regime_groups: dict[str, list[float]] = defaultdict(list)
        for t in group:
            regime_groups[t.regime].append(t.return_pct)
        reg_stats = {}
        for reg, rets in regime_groups.items():
            reg_stats[reg] = {
                "count": len(rets),
                "hit_rate": round(sum(1 for r in rets if r > 0) / len(rets), 4) if rets else 0.0,
                "avg_return": round(sum(rets) / len(rets), 4) if rets else 0.0,
            }
        result[name] = StrategyStats(
            strategy_name=name, total_signals=len(group), wins=wins,
            hit_rate=round(wins / len(group), 4) if group else 0.0,
            avg_return=round(sum(returns) / len(returns), 4) if returns else 0.0,
            total_return=round(sum(returns), 4),
            max_return=round(max(returns), 4) if returns else 0.0,
            min_return=round(min(returns), 4) if returns else 0.0,
            avg_score=round(sum(scores) / len(scores), 2) if scores else 0.0,
            regime_stats=reg_stats,
        )
    return result


def compute_regime_stats(trades: list[TradeRecord]) -> dict[str, RegimeStats]:
    by_regime: dict[str, list[TradeRecord]] = defaultdict(list)
    for t in trades:
        by_regime[t.regime].append(t)
    result: dict[str, RegimeStats] = {}
    for regime, group in by_regime.items():
        wins = sum(1 for t in group if t.is_win)
        returns = [t.return_pct for t in group]
        contrib = defaultdict(int)
        for t in group:
            contrib[t.strategy] += 1
        result[regime] = RegimeStats(
            regime=regime, total_signals=len(group), wins=wins,
            hit_rate=round(wins / len(group), 4) if group else 0.0,
            avg_return=round(sum(returns) / len(returns), 4) if returns else 0.0,
            strategy_contributions=dict(contrib),
        )
    return result


def compute_result(start_date: str, end_date: str, trading_days: int,
                   trades: list[TradeRecord], daily_returns: list[dict[str, Any]],
                   warnings: list[str]) -> BacktestResult:
    pt = [t for t in trades if t.portfolio_role in ("CORE", "SATELLITE")]
    active_days = len({t.trade_date for t in pt})
    total_trades = len(pt)
    win_trades = sum(1 for t in pt if t.is_win)
    hit_rate = win_trades / total_trades if total_trades > 0 else 0.0
    avg_return = sum(t.return_pct for t in pt) / total_trades if total_trades > 0 else 0.0

    cumulative = 0.0
    peak = 0.0
    max_dd = 0.0
    best_day = worst_day = ""
    best_ret = -999.0
    worst_ret = 999.0
    for day in daily_returns:
        cumulative += day["avg_return"]
        peak = max(peak, cumulative)
        max_dd = max(max_dd, peak - cumulative)
        if day["avg_return"] > best_ret:
            best_ret, best_day = day["avg_return"], day["date"]
        if day["avg_return"] < worst_ret:
            worst_ret, worst_day = day["avg_return"], day["date"]

    multi = [t for t in pt if t.is_multi_strategy]
    m_hit = sum(1 for t in multi if t.is_win) / len(multi) if multi else 0.0
    m_avg = sum(t.return_pct for t in multi) / len(multi) if multi else 0.0

    return BacktestResult(
        start_date=start_date, end_date=end_date, trading_days=trading_days,
        active_days=active_days, total_trades=total_trades, win_trades=win_trades,
        overall_hit_rate=round(hit_rate, 4),
        overall_avg_return=round(avg_return, 4),
        overall_cumulative_return=round(cumulative, 4),
        max_drawdown=round(max_dd, 4),
        best_day=best_day, best_day_return=round(best_ret, 4),
        worst_day=worst_day, worst_day_return=round(worst_ret, 4),
        multi_strategy_trades=len(multi),
        multi_strategy_hit_rate=round(m_hit, 4),
        multi_strategy_avg_return=round(m_avg, 4),
        strategy_stats=compute_strategy_stats(trades),
        regime_stats=compute_regime_stats(trades),
        daily_returns=daily_returns, trade_records=trades, warnings=warnings,
    )


# ————————————————————————————————————————————
# 报告格式化（与 backtest_reporter.py 一致）
# ————————————————————————————————————————————

def format_summary(result: BacktestResult) -> str:
    lines = [
        "=" * 72,
        "  超短线 T+1 回测报告（烟幕测试 · 合成数据）",
        "=" * 72,
        "",
        f"  回测区间: {result.start_date} → {result.end_date}",
        f"  交易日数: {result.trading_days}  有信号天数: {result.active_days}",
        f"  总交易: {result.total_trades}  胜: {result.win_trades}  "
        f"胜率: {result.overall_hit_rate:.2%}",
        f"  平均收益: {result.overall_avg_return:+.2f}%  "
        f"累计收益: {result.overall_cumulative_return:+.2f}%  "
        f"最大回撤: {result.max_drawdown:.2f}%",
        f"  最佳交易日: {result.best_day} ({result.best_day_return:+.2f}%)  "
        f"最差交易日: {result.worst_day} ({result.worst_day_return:+.2f}%)",
        "",
    ]

    if result.multi_strategy_trades > 0:
        lines.append(f"  多策略共振: {result.multi_strategy_trades}笔  "
                     f"胜率: {result.multi_strategy_hit_rate:.2%}  "
                     f"平均: {result.multi_strategy_avg_return:+.2f}%")
        lines.append("")

    # 按策略统计
    lines.append("  ── 策略表现 ──")
    lines.append(f"  {'策略':<18s} {'信号':>5s} {'胜':>5s} {'胜率':>7s} {'平均':>8s} {'累计':>8s} {'最高':>8s} {'最低':>8s} {'均分':>6s}")
    lines.append("  " + "-" * 68)
    for name, s in sorted(result.strategy_stats.items()):
        lines.append(
            f"  {s.strategy_name:<18s} {s.total_signals:>5d} {s.wins:>5d} "
            f"{s.hit_rate:>6.1%} {s.avg_return:>+7.2f}% {s.total_return:>+7.2f}% "
            f"{s.max_return:>+7.2f}% {s.min_return:>+7.2f}% {s.avg_score:>5.1f}"
        )
        # 按市场环境子统计
        if len(s.regime_stats) > 1:
            for reg, rs in sorted(s.regime_stats.items()):
                lines.append(f"    └ {reg:<12s}  {rs['count']:>3d}笔  "
                             f"胜率: {rs['hit_rate']:.1%}  平均: {rs['avg_return']:+.2f}%")

    lines.append("")

    # 按市场环境统计
    lines.append("  ── 市场环境 ──")
    for name, rs in sorted(result.regime_stats.items()):
        lines.append(f"  {rs.regime:<18s} {rs.total_signals:>5d}笔  "
                     f"胜率: {rs.hit_rate:.1%}  平均: {rs.avg_return:+.2f}%  "
                     f"分布: {dict(rs.strategy_contributions)}")

    lines.append("")

    # 日收益概览
    if result.daily_returns:
        monthly: dict[str, list[float]] = defaultdict(list)
        for d in result.daily_returns:
            monthly[d["date"][:7]].append(d["avg_return"])
        lines.append("  ── 月度收益 ──")
        for m, rets in sorted(monthly.items()):
            avg_m = sum(rets) / len(rets) if rets else 0
            lines.append(f"  {m}: {len(rets)}天  "
                         f"均值: {avg_m:+.2f}%  "
                         f"胜率: {sum(1 for r in rets if r > 0)/len(rets):.1%}")

    if result.warnings:
        lines.append("")
        lines.append("  ── 警告 ──")
        for w in result.warnings[-10:]:  # 最多显示最后 10 条
            lines.append(f"  ⚠ {w}")

    lines.append("")
    lines.append("=" * 72)
    lines.append("  注意: 此为合成数据烟幕测试，收益不反映真实市场表现。")
    lines.append("  目的: 验证策略管道 (7策略 → 融合 → 风控 → 收益计算) 端到端运行正确。")
    lines.append("=" * 72)
    return "\n".join(lines)


def result_to_dict(result: BacktestResult) -> dict:
    return {
        "meta": {
            "start_date": result.start_date, "end_date": result.end_date,
            "trading_days": result.trading_days, "active_days": result.active_days,
            "total_trades": result.total_trades, "win_trades": result.win_trades,
            "hit_rate": result.overall_hit_rate,
            "avg_return": result.overall_avg_return,
            "cumulative_return": result.overall_cumulative_return,
            "max_drawdown": result.max_drawdown,
            "best_day": result.best_day, "best_day_return": result.best_day_return,
            "worst_day": result.worst_day, "worst_day_return": result.worst_day_return,
            "multi_strategy_trades": result.multi_strategy_trades,
            "multi_strategy_hit_rate": result.multi_strategy_hit_rate,
            "multi_strategy_avg_return": result.multi_strategy_avg_return,
            "source": "synthetic-smoke-test",
        },
        "strategies": {
            name: {
                "signals": s.total_signals, "wins": s.wins,
                "hit_rate": s.hit_rate, "avg_return": s.avg_return,
                "total_return": s.total_return, "max_return": s.max_return,
                "min_return": s.min_return, "avg_score": s.avg_score,
                "by_regime": s.regime_stats,
            }
            for name, s in result.strategy_stats.items()
        },
        "regimes": {
            name: {
                "signals": rs.total_signals, "wins": rs.wins,
                "hit_rate": rs.hit_rate, "avg_return": rs.avg_return,
                "strategy_distribution": rs.strategy_contributions,
            }
            for name, rs in result.regime_stats.items()
        },
        "trades": [
            {
                "symbol": t.symbol, "name": t.name, "date": t.trade_date,
                "strategy": t.strategy, "score": t.score,
                "role": t.portfolio_role, "buy": t.buy_price, "sell": t.sell_price,
                "return_pct": t.return_pct, "is_win": t.is_win,
                "regime": t.regime, "multi_strategy": t.is_multi_strategy,
            }
            for t in result.trade_records
        ],
        "daily_returns": result.daily_returns,
        "warnings": result.warnings,
    }


def trades_to_csv(result: BacktestResult) -> str:
    lines = ["symbol,name,date,strategy,score,role,buy_price,sell_price,return_pct,is_win,regime,multi_strategy"]
    for t in result.trade_records:
        lines.append(
            f"{t.symbol},{t.name},{t.trade_date},{t.strategy},{t.score},"
            f"{t.portfolio_role},{t.buy_price},{t.sell_price},"
            f"{t.return_pct},{1 if t.is_win else 0},{t.regime},"
            f"{1 if t.is_multi_strategy else 0}"
        )
    return "\n".join(lines) + "\n"


# ————————————————————————————————————————————
# 主回测循环
# ————————————————————————————————————————————

def run_smoke_backtest(
    num_seeds: int = 30, num_days: int = 90, limit: int = 5,
    min_score: float = 65.0, max_risk_level: str = "MEDIUM",
    max_per_strategy: int = 2, max_per_sector: int = 1,
    rng_seed: int = 42,
) -> BacktestResult:
    total_ticks = num_days * 2
    print(f"\n{'='*70}")
    print(f"  超短线 T+1 回测烟幕测试")
    print(f"  股票: {num_seeds}只 | 交易日: {num_days}天 | 总tick: {total_ticks}")
    print(f"{'='*70}\n")

    pool = STOCK_POOL[:num_seeds]
    loader = SyntheticSeedLoader(pool, rng_seed=rng_seed)

    # 日期范围
    today = date.today()
    end_str = today.isoformat()
    start_d = today
    cnt = 0
    while cnt < num_days + 1:
        start_d -= timedelta(days=1)
        if start_d.weekday() < 5:
            cnt += 1
    trade_dates = trading_dates_between(start_d.isoformat(), end_str)
    trade_dates = trade_dates[-num_days:] if len(trade_dates) > num_days else trade_dates

    print(f"  回测时段: {trade_dates[0]} → {trade_dates[-1]}")
    print(f"  参数: limit={limit}  min_score={min_score}  "
          f"max_risk={max_risk_level}  max_per_s={max_per_strategy}  "
          f"max_per_sec={max_per_sector}")
    print(f"  [进度] ", end="", flush=True)

    miner = IntradaySeedMiner()
    fusion = IntradayDecisionFusion(
        max_per_strategy=max_per_strategy, max_per_sector=max_per_sector,
        limit=limit, min_score=min_score,
    )
    guard = PortfolioGuard()

    all_trades: list[TradeRecord] = []
    daily_returns: list[dict[str, Any]] = []
    warnings_list: list[str] = []

    for i, td in enumerate(trade_dates[:-1]):  # 最后一天无 T+1 价格
        load_result = loader.load(td)
        seeds = load_result.seeds
        regime = detect_regime(seeds)

        mining = miner.mine(seeds, regime)
        fused = fusion.fuse(mining.strategy_results, regime)
        guarded = guard.apply(
            fused.candidates, max_risk_level, min_score, limit,
            max_per_strategy, max_per_sector,
        )

        # 获取 T+1 价格
        next_date = next_trading_date(td)
        next_load = loader.load(next_date)
        next_prices: dict[str, float] = {}
        for s in next_load.seeds:
            if s.close_price > 0:
                next_prices[s.symbol] = s.close_price

        # 生成交易记录
        for item in guarded.portfolio + guarded.watchlist:
            bp = item.close_price
            sp = next_prices.get(item.symbol, bp)
            ret = (sp / bp - 1) * 100 if bp > 0 else 0.0
            strat = (item.strategy_name or "").strip() or "unknown"
            all_trades.append(TradeRecord(
                symbol=item.symbol, name=item.name, trade_date=td,
                strategy=strat, score=item.score,
                portfolio_role=item.portfolio_role,
                buy_price=bp, sell_price=sp, return_pct=round(ret, 4),
                is_win=ret > 0, regime=regime,
                is_multi_strategy=getattr(item, "multi_strategy_hit", 0) > 1,
            ))

        # 日总结
        day_pt = [t for t in all_trades if t.trade_date == td and t.portfolio_role in ("CORE", "SATELLITE")]
        if day_pt:
            daily_returns.append({
                "date": td, "trades": len(day_pt),
                "avg_return": round(sum(t.return_pct for t in day_pt) / len(day_pt), 4),
                "win_rate": round(sum(1 for t in day_pt if t.is_win) / len(day_pt), 4),
            })

        if (i + 1) % 10 == 0:
            print(f"{i+1}/{len(trade_dates)-1} ", end="", flush=True)

    print("\n")

    return compute_result(
        start_date=trade_dates[0], end_date=trade_dates[-1],
        trading_days=len(trade_dates) - 1, trades=all_trades,
        daily_returns=daily_returns, warnings=warnings_list,
    )


# ————————————————————————————————————————————
# CLI
# ————————————————————————————————————————————

def main() -> None:
    ap = argparse.ArgumentParser(description="超短线 T+1 回测烟幕测试（合成数据）")
    ap.add_argument("--days", type=int, default=90, help="交易日数 (default: 90)")
    ap.add_argument("--seeds", type=int, default=30, help="股票数量 (default: 30)")
    ap.add_argument("--limit", type=int, default=5, help="每日推荐数 (default: 5)")
    ap.add_argument("--min-score", type=float, default=65.0, help="最低评分 (default: 65)")
    ap.add_argument("--max-risk", type=str, default="MEDIUM", help="最大风险等级")
    ap.add_argument("--max-per-strategy", type=int, default=2, help="每策略最大输出")
    ap.add_argument("--max-per-sector", type=int, default=1, help="每板块最大输出")
    ap.add_argument("--seed", type=int, default=42, help="随机种子")
    ap.add_argument("--format", choices=["summary", "json", "csv"], default="summary")
    ap.add_argument("--output", type=str, default="", help="输出文件路径")

    args = ap.parse_args()

    result = run_smoke_backtest(
        num_seeds=args.seeds, num_days=args.days, limit=args.limit,
        min_score=args.min_score, max_risk_level=args.max_risk,
        max_per_strategy=args.max_per_strategy, max_per_sector=args.max_per_sector,
        rng_seed=args.seed,
    )

    if args.format == "csv":
        output = trades_to_csv(result)
    elif args.format == "json":
        import json as _json
        output = _json.dumps(result_to_dict(result), ensure_ascii=False, indent=2)
    else:
        output = format_summary(result)

    if args.output:
        with open(args.output, "w", encoding="utf-8") as f:
            f.write(output)
        print(f"结果已保存至: {args.output}")
    else:
        print(output)


if __name__ == "__main__":
    main()
