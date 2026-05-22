from __future__ import annotations

import logging
from collections import defaultdict
from dataclasses import dataclass, field
from datetime import date, timedelta
from typing import Any

from app.domain.decision.intraday_decision_fusion import IntradayDecisionFusion
from app.domain.models import MarketSeed, StockFeature
from app.domain.risk.portfolio_guard import PortfolioGuard
from app.domain.seeds.intraday_seed_miner import IntradaySeedMiner
from app.domain.seeds.market_seed_loader import MarketSeedLoader
from app.schemas.stock import StockSelectionPayload

logger = logging.getLogger(__name__)


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


class IntradayBacktestRunner:
    def __init__(self, seed_loader: MarketSeedLoader | None = None) -> None:
        self._seed_loader = seed_loader or MarketSeedLoader()

    def run(
        self,
        start_date: str,
        end_date: str,
        *,
        limit: int = 5,
        min_score: float = 65.0,
        max_risk_level: str = "MEDIUM",
        watchlist_limit: int = 10,
        max_symbol_per_bucket: int = 2,
        max_symbols_per_sector: int = 1,
        max_per_strategy: int = 2,
        trade_dates: list[str] | None = None,
    ) -> BacktestResult:
        if trade_dates is None:
            trade_dates = _trading_dates_between(start_date, end_date)
        if not trade_dates:
            raise ValueError(f"no trading dates between {start_date} and {end_date}")

        all_trades: list[TradeRecord] = []
        daily_returns: list[dict[str, Any]] = []
        warnings: list[str] = []

        for trade_date in trade_dates:
            try:
                day_trades = self._run_single_day(
                    trade_date=trade_date,
                    limit=limit,
                    min_score=min_score,
                    max_risk_level=max_risk_level,
                    watchlist_limit=watchlist_limit,
                    max_symbol_per_bucket=max_symbol_per_bucket,
                    max_symbols_per_sector=max_symbols_per_sector,
                    max_per_strategy=max_per_strategy,
                )
                all_trades.extend(day_trades)
                daily_returns.append(self._day_summary(trade_date, day_trades))
            except Exception as exc:
                warnings.append(f"{trade_date}: {exc}")
                logger.warning("backtest day %s failed: %s", trade_date, exc)

        return self._compute_result(
            start_date=start_date,
            end_date=end_date,
            trading_days=len(trade_dates),
            trades=all_trades,
            daily_returns=daily_returns,
            warnings=warnings,
        )

    def _run_single_day(
        self,
        trade_date: str,
        *,
        limit: int,
        min_score: float,
        max_risk_level: str,
        watchlist_limit: int,
        max_symbol_per_bucket: int,
        max_symbols_per_sector: int,
        max_per_strategy: int,
    ) -> list[TradeRecord]:
        payload = StockSelectionPayload(
            template_key="INTRADAY_T1",
            trade_date=trade_date,
            limit=limit,
            min_score=min_score,
            max_risk_level=max_risk_level,
            watchlist_limit=watchlist_limit,
            max_symbol_per_bucket=max_symbol_per_bucket,
            max_symbols_per_sector=max_symbols_per_sector,
        )

        load_result = self._seed_loader.load(payload)
        seeds = load_result.seeds

        regime = self._detect_regime(seeds)
        miner = IntradaySeedMiner()
        mining_result = miner.mine(seeds, regime)
        fusion = IntradayDecisionFusion()
        fusion_result = fusion.fuse(mining_result.strategy_results, regime)
        guard = PortfolioGuard()
        guarded = guard.apply(fusion_result.candidates, payload)

        next_date = _next_trade_date(trade_date)
        next_prices = self._load_next_prices(guarded.portfolio + guarded.watchlist, next_date)

        trades: list[TradeRecord] = []
        for item in guarded.portfolio + guarded.watchlist:
            buy_price = item.close_price
            sell_price = next_prices.get(item.symbol, buy_price)
            return_pct = (sell_price / buy_price - 1) * 100 if buy_price > 0 else 0.0
            strategy = (item.strategy_name or "").strip() or "unknown"
            trades.append(TradeRecord(
                symbol=item.symbol,
                name=item.name,
                trade_date=trade_date,
                strategy=strategy,
                score=item.score,
                portfolio_role=item.portfolio_role,
                buy_price=buy_price,
                sell_price=sell_price,
                return_pct=round(return_pct, 4),
                is_win=return_pct > 0,
                regime=regime,
                is_multi_strategy=getattr(item, "multi_strategy_hit", 0) > 1,
            ))

        return trades

    def _load_next_prices(self, items: list[StockFeature], next_date: str) -> dict[str, float]:
        if not items:
            return {}
        try:
            payload = StockSelectionPayload(
                template_key="INTRADAY_T1",
                trade_date=next_date,
                limit=len(items),
                seed_symbols=[item.symbol for item in items],
            )
            load_result = self._seed_loader.load(payload)
            price_map: dict[str, float] = {}
            for seed in load_result.seeds:
                if seed.close_price > 0:
                    price_map[seed.symbol] = seed.close_price
            return price_map
        except Exception as exc:
            logger.warning("failed to load next-day prices for %s: %s", next_date, exc)
            return {}

    def _detect_regime(self, seeds: list[MarketSeed]) -> str:
        if not seeds:
            return "UNKNOWN"
        avg_momentum5 = sum(s.momentum5 for s in seeds if hasattr(s, "momentum5")) / max(len(seeds), 1)
        avg_momentum20 = sum(s.momentum20 for s in seeds if hasattr(s, "momentum20")) / max(len(seeds), 1)
        avg_volatility20 = sum(s.volatility20 for s in seeds if hasattr(s, "volatility20")) / max(len(seeds), 1)

        if avg_volatility20 > 12:
            return "RISK_OFF"
        if avg_momentum20 < -5 and avg_volatility20 > 8:
            return "DEFENSIVE"
        if avg_momentum5 > 3 and avg_momentum20 > 0:
            return "UPTREND"
        if 2 < abs(avg_momentum5) < 6:
            return "ROTATION"
        if avg_momentum5 > 4 and avg_volatility20 > 8:
            return "EVENT_DRIVEN"
        return "ROTATION"

    def _day_summary(self, trade_date: str, trades: list[TradeRecord]) -> dict[str, Any]:
        portfolio_trades = [t for t in trades if t.portfolio_role in ("CORE", "SATELLITE")]
        if not portfolio_trades:
            return {"date": trade_date, "trades": 0, "avg_return": 0.0, "win_rate": 0.0}
        avg_ret = sum(t.return_pct for t in portfolio_trades) / len(portfolio_trades)
        wins = sum(1 for t in portfolio_trades if t.is_win)
        return {
            "date": trade_date,
            "trades": len(portfolio_trades),
            "avg_return": round(avg_ret, 4),
            "win_rate": round(wins / len(portfolio_trades), 4),
        }

    def _compute_result(
        self,
        start_date: str,
        end_date: str,
        trading_days: int,
        trades: list[TradeRecord],
        daily_returns: list[dict[str, Any]],
        warnings: list[str],
    ) -> BacktestResult:
        portfolio_trades = [t for t in trades if t.portfolio_role in ("CORE", "SATELLITE")]
        active_days = len({t.trade_date for t in portfolio_trades})

        total_trades = len(portfolio_trades)
        win_trades = sum(1 for t in portfolio_trades if t.is_win)
        overall_hit_rate = win_trades / total_trades if total_trades > 0 else 0.0
        overall_avg_return = sum(t.return_pct for t in portfolio_trades) / total_trades if total_trades > 0 else 0.0

        cumulative = 0.0
        peak = 0.0
        max_drawdown = 0.0
        best_day = ""
        best_day_return = -999.0
        worst_day = ""
        worst_day_return = 999.0
        for day in daily_returns:
            cumulative += day["avg_return"]
            if cumulative > peak:
                peak = cumulative
            dd = peak - cumulative
            if dd > max_drawdown:
                max_drawdown = dd
            if day["avg_return"] > best_day_return:
                best_day_return = day["avg_return"]
                best_day = day["date"]
            if day["avg_return"] < worst_day_return:
                worst_day_return = day["avg_return"]
                worst_day = day["date"]

        multi_trades = [t for t in portfolio_trades if t.is_multi_strategy]
        multi_hit_rate = sum(1 for t in multi_trades if t.is_win) / len(multi_trades) if multi_trades else 0.0
        multi_avg_return = sum(t.return_pct for t in multi_trades) / len(multi_trades) if multi_trades else 0.0

        strategy_stats = self._compute_strategy_stats(trades)
        regime_stats = self._compute_regime_stats(trades)

        return BacktestResult(
            start_date=start_date,
            end_date=end_date,
            trading_days=trading_days,
            active_days=active_days,
            total_trades=total_trades,
            win_trades=win_trades,
            overall_hit_rate=round(overall_hit_rate, 4),
            overall_avg_return=round(overall_avg_return, 4),
            overall_cumulative_return=round(cumulative, 4),
            max_drawdown=round(max_drawdown, 4),
            best_day=best_day,
            best_day_return=round(best_day_return, 4),
            worst_day=worst_day,
            worst_day_return=round(worst_day_return, 4),
            multi_strategy_trades=len(multi_trades),
            multi_strategy_hit_rate=round(multi_hit_rate, 4),
            multi_strategy_avg_return=round(multi_avg_return, 4),
            strategy_stats=strategy_stats,
            regime_stats=regime_stats,
            daily_returns=daily_returns,
            trade_records=trades,
            warnings=warnings,
        )

    def _compute_strategy_stats(self, trades: list[TradeRecord]) -> dict[str, StrategyStats]:
        by_strategy: dict[str, list[TradeRecord]] = defaultdict(list)
        for t in trades:
            by_strategy[t.strategy].append(t)

        result: dict[str, StrategyStats] = {}
        for name, group in by_strategy.items():
            wins = sum(1 for t in group if t.is_win)
            returns = [t.return_pct for t in group]
            scores = [t.score for t in group]
            regime_groups: dict[str, list[float]] = defaultdict(list)
            for t in group:
                regime_groups[t.regime].append(t.return_pct)

            regime_stats = {}
            for reg, rets in regime_groups.items():
                regime_stats[reg] = {
                    "count": len(rets),
                    "hit_rate": round(sum(1 for r in rets if r > 0) / len(rets), 4),
                    "avg_return": round(sum(rets) / len(rets), 4),
                }

            result[name] = StrategyStats(
                strategy_name=name,
                total_signals=len(group),
                wins=wins,
                hit_rate=round(wins / len(group), 4) if group else 0.0,
                avg_return=round(sum(returns) / len(returns), 4) if returns else 0.0,
                total_return=round(sum(returns), 4),
                max_return=round(max(returns), 4) if returns else 0.0,
                min_return=round(min(returns), 4) if returns else 0.0,
                avg_score=round(sum(scores) / len(scores), 2) if scores else 0.0,
                regime_stats=regime_stats,
            )
        return result

    def _compute_regime_stats(self, trades: list[TradeRecord]) -> dict[str, RegimeStats]:
        by_regime: dict[str, list[TradeRecord]] = defaultdict(list)
        for t in trades:
            by_regime[t.regime].append(t)

        result: dict[str, RegimeStats] = {}
        for regime, group in by_regime.items():
            wins = sum(1 for t in group if t.is_win)
            returns = [t.return_pct for t in group]
            strategy_contrib = defaultdict(int)
            for t in group:
                strategy_contrib[t.strategy] += 1

            result[regime] = RegimeStats(
                regime=regime,
                total_signals=len(group),
                wins=wins,
                hit_rate=round(wins / len(group), 4) if group else 0.0,
                avg_return=round(sum(returns) / len(returns), 4) if returns else 0.0,
                strategy_contributions=dict(strategy_contrib),
            )
        return result


def _trading_dates_between(start_date: str, end_date: str) -> list[str]:
    start = date.fromisoformat(start_date)
    end = date.fromisoformat(end_date)
    if end < start:
        raise ValueError(f"end_date {end_date} must be >= start_date {start_date}")
    dates: list[str] = []
    current = start
    while current <= end:
        if current.weekday() < 5:
            dates.append(current.isoformat())
        current += timedelta(days=1)
    return dates


def _next_trade_date(trade_date: str) -> str:
    d = date.fromisoformat(trade_date) + timedelta(days=1)
    while d.weekday() >= 5:
        d += timedelta(days=1)
    return d.isoformat()
