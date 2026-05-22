from __future__ import annotations

import csv
import json
import logging
from io import StringIO
from typing import Any

from app.domain.backtesting.intraday_backtest import BacktestResult, StrategyStats, RegimeStats

logger = logging.getLogger(__name__)


def format_backtest_summary(result: BacktestResult) -> str:
    lines: list[str] = []
    lines.append("=" * 70)
    lines.append(f"  超短线 T+1 回测报告  |  {result.start_date} ~ {result.end_date}")
    lines.append("=" * 70)
    lines.append("")

    lines.append("--- 整体表现 ---")
    lines.append(f"  回测交易日数:        {result.trading_days}")
    lines.append(f"  有效荐股日数:        {result.active_days}")
    lines.append(f"  总交易笔数:          {result.total_trades}")
    lines.append(f"  胜率 (hit_rate_1):   {result.overall_hit_rate:.2%}")
    lines.append(f"  平均收益率:          {result.overall_avg_return:+.2%}")
    lines.append(f"  累计收益率:          {result.overall_cumulative_return:+.2%}")
    lines.append(f"  最大回撤:            {result.max_drawdown:.2%}")
    lines.append(f"  最佳日 ({result.best_day}):  {result.best_day_return:+.2%}")
    lines.append(f"  最差日 ({result.worst_day}):  {result.worst_day_return:+.2%}")
    lines.append("")

    lines.append("--- 多策略共振 ---")
    lines.append(f"  共振交易笔数:        {result.multi_strategy_trades}")
    lines.append(f"  共振胜率:            {result.multi_strategy_hit_rate:.2%}")
    lines.append(f"  共振平均收益:        {result.multi_strategy_avg_return:+.2%}")
    lines.append("")

    lines.append("--- 策略分层统计 ---")
    lines.append(f"  {'策略':<20s} {'信号':>5s} {'胜率':>8s} {'均收益':>8s} {'累计':>8s} {'最高':>8s} {'最低':>8s}")
    lines.append("  " + "-" * 65)
    for name, stats in sorted(result.strategy_stats.items()):
        lines.append(
            f"  {name:<20s} {stats.total_signals:>5d} {stats.hit_rate:>7.1%} {stats.avg_return:>+7.2%} "
            f"{stats.total_return:>+7.2%} {stats.max_return:>+7.2%} {stats.min_return:>+7.2%}"
        )
    lines.append("")

    lines.append("--- 市场环境分层统计 ---")
    for regime, stats in sorted(result.regime_stats.items()):
        lines.append(f"  [{regime}] 信号={stats.total_signals}  胜率={stats.hit_rate:.1%}  均收益={stats.avg_return:+.2%}")
        for strat, count in sorted(stats.strategy_contributions.items(), key=lambda x: -x[1]):
            lines.append(f"         {strat}: {count}笔")
    lines.append("")

    if result.warnings:
        lines.append(f"--- 警告 ({len(result.warnings)}条) ---")
        for w in result.warnings[:10]:
            lines.append(f"  - {w}")
        if len(result.warnings) > 10:
            lines.append(f"  ... 还有 {len(result.warnings) - 10} 条")
        lines.append("")

    lines.append("=" * 70)
    return "\n".join(lines)


def backtest_result_to_dict(result: BacktestResult) -> dict[str, Any]:
    return {
        "start_date": result.start_date,
        "end_date": result.end_date,
        "trading_days": result.trading_days,
        "active_days": result.active_days,
        "total_trades": result.total_trades,
        "win_trades": result.win_trades,
        "overall_hit_rate": result.overall_hit_rate,
        "overall_avg_return": result.overall_avg_return,
        "overall_cumulative_return": result.overall_cumulative_return,
        "max_drawdown": result.max_drawdown,
        "best_day": result.best_day,
        "best_day_return": result.best_day_return,
        "worst_day": result.worst_day,
        "worst_day_return": result.worst_day_return,
        "multi_strategy_trades": result.multi_strategy_trades,
        "multi_strategy_hit_rate": result.multi_strategy_hit_rate,
        "multi_strategy_avg_return": result.multi_strategy_avg_return,
        "strategy_stats": {
            name: {
                "strategy_name": s.strategy_name,
                "total_signals": s.total_signals,
                "wins": s.wins,
                "hit_rate": s.hit_rate,
                "avg_return": s.avg_return,
                "total_return": s.total_return,
                "max_return": s.max_return,
                "min_return": s.min_return,
                "avg_score": s.avg_score,
                "regime_stats": s.regime_stats,
            }
            for name, s in result.strategy_stats.items()
        },
        "regime_stats": {
            r: {
                "regime": rs.regime,
                "total_signals": rs.total_signals,
                "wins": rs.wins,
                "hit_rate": rs.hit_rate,
                "avg_return": rs.avg_return,
                "strategy_contributions": rs.strategy_contributions,
            }
            for r, rs in result.regime_stats.items()
        },
        "daily_returns": result.daily_returns,
        "warnings": result.warnings,
    }


def backtest_to_json(result: BacktestResult) -> str:
    return json.dumps(backtest_result_to_dict(result), ensure_ascii=False, indent=2)


def backtest_trades_to_csv(result: BacktestResult) -> str:
    output = StringIO()
    writer = csv.writer(output)
    writer.writerow([
        "trade_date", "symbol", "name", "strategy", "role", "score",
        "buy_price", "sell_price", "return_pct", "is_win", "regime", "multi_strategy",
    ])
    for t in result.trade_records:
        writer.writerow([
            t.trade_date, t.symbol, t.name, t.strategy, t.portfolio_role, t.score,
            t.buy_price, t.sell_price, t.return_pct, t.is_win, t.regime, t.is_multi_strategy,
        ])
    return output.getvalue()
