from __future__ import annotations

from dataclasses import dataclass, field

from app.domain.market.market_daily_analyzer import MarketAnalysisResult
from app.domain.models import StockFeature


@dataclass(slots=True)
class TrendCandidatePoolResult:
    coarse_pool: list[StockFeature]
    focus_pool: list[StockFeature]
    watch_pool: list[StockFeature]
    summary: dict[str, int | str] = field(default_factory=dict)


class TrendCandidatePoolBuilder:
    def build(self, features: list[StockFeature], market: MarketAnalysisResult) -> TrendCandidatePoolResult:
        filtered = [item for item in features if _passes_trend_gate(item)]
        ranked = sorted(filtered, key=_trend_sort_key, reverse=True)
        if market.risk_posture == "TIGHT":
            coarse_limit, focus_limit, watch_limit = 50, 20, 10
        else:
            coarse_limit, focus_limit, watch_limit = 100, 30, 12
        coarse_pool = ranked[:coarse_limit]
        focus_pool = coarse_pool[:focus_limit]
        watch_pool = focus_pool[:watch_limit]
        return TrendCandidatePoolResult(
            coarse_pool=coarse_pool,
            focus_pool=focus_pool,
            watch_pool=watch_pool,
            summary={
                "risk_posture": market.risk_posture,
                "coarse_pool_count": len(coarse_pool),
                "focus_pool_count": len(focus_pool),
                "watch_pool_count": len(watch_pool),
                "watch_pool_limit": watch_limit,
            },
        )


def _passes_trend_gate(item: StockFeature) -> bool:
    if item.momentum20 <= 0:
        return False
    if item.trend_score < 50:
        return False
    if item.drawdown20 >= 12:
        return False
    if item.volatility20 >= 5.0:
        return False
    return True


def _trend_sort_key(item: StockFeature) -> tuple[float, float, float, float]:
    return (
        item.trend_score,
        item.quant_score,
        item.momentum20,
        item.flow_score,
    )
