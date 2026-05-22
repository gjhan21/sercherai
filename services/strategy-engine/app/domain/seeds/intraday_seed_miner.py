from __future__ import annotations

from dataclasses import dataclass, field

from app.domain.models import MarketSeed
from app.domain.strategies.base import StrategyResult
from app.domain.strategies.oversold_bounce import OversoldBounceStrategy
from app.domain.strategies.volume_exhaustion import VolumeExhaustionStrategy
from app.domain.strategies.capital_divergence import CapitalDivergenceStrategy
from app.domain.strategies.ma_support import MASupportStrategy
from app.domain.strategies.bullish_engulfing import BullishEngulfingStrategy
from app.domain.strategies.limit_up_next_day import LimitUpNextDayStrategy
from app.domain.strategies.strong_pullback import StrongPullbackStrategy


@dataclass(slots=True)
class IntradaySeedMiningResult:
    strategy_results: dict[str, StrategyResult] = field(default_factory=dict)
    warnings: list[str] = field(default_factory=list)


class IntradaySeedMiner:
    """并行运行全部7个T+1超短线策略，产出各策略独立候选池。

    替代原有的5桶 StockSeedMiner，用于 intraday_t1 模板。
    每个策略独立从全量 MarketSeed 中筛选、评分、排序，
    结果汇总后交由 IntradayDecisionFusion 做跨策略融合。
    """

    def __init__(self) -> None:
        self._strategies = [
            OversoldBounceStrategy(max_candidates=10),
            VolumeExhaustionStrategy(max_candidates=10),
            CapitalDivergenceStrategy(max_candidates=10),
            MASupportStrategy(max_candidates=10),
            BullishEngulfingStrategy(max_candidates=8),
            LimitUpNextDayStrategy(max_candidates=15),
            StrongPullbackStrategy(max_candidates=10),
        ]

    def mine(
        self,
        seeds: list[MarketSeed],
        market_regime: str,
    ) -> IntradaySeedMiningResult:
        if not seeds:
            return IntradaySeedMiningResult()

        all_warnings: list[str] = []
        strategy_results: dict[str, StrategyResult] = {}

        for strategy in self._strategies:
            result = strategy.run(seeds, market_regime)
            strategy_results[strategy.strategy_key] = result
            if result.warnings:
                all_warnings.extend(result.warnings)

        total_candidates = sum(
            len(r.candidates) for r in strategy_results.values()
        )
        if total_candidates == 0:
            all_warnings.append("全部T+1策略均未产出候选标的，请检查当日市场数据。")

        return IntradaySeedMiningResult(
            strategy_results=strategy_results,
            warnings=all_warnings,
        )
