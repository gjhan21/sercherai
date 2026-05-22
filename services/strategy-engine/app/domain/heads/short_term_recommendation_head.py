from __future__ import annotations

from dataclasses import dataclass, field

from app.domain.market.market_daily_analyzer import MarketAnalysisResult
from app.domain.models import StockFeature


@dataclass(slots=True)
class ShortTermRecommendationResult:
    recommendations: list[StockFeature]
    summary: dict[str, int | str] = field(default_factory=dict)


class ShortTermRecommendationHead:
    def select(
        self,
        features: list[StockFeature],
        market: MarketAnalysisResult,
        intraday_inputs: dict[str, dict[str, object]],
    ) -> ShortTermRecommendationResult:
        candidates: list[StockFeature] = []
        for item in sorted(features, key=lambda feature: (feature.score, feature.momentum20), reverse=True):
            intraday = intraday_inputs.get(item.symbol, {})
            minute_score = int(intraday.get("minute_confirmation_score", 0) or 0)
            vwap_supported = bool(intraday.get("vwap_supported", False))
            opening_range_intact = bool(intraday.get("opening_range_intact", False))
            if minute_score < 70 or not vwap_supported or not opening_range_intact:
                continue
            cloned = item
            cloned.portfolio_role = "CORE"
            cloned.recommendation_head = "SHORT_TERM_PRIMARY"
            cloned.selection_layer = "L2_PRIMARY"
            candidates.append(cloned)
            if len(candidates) >= 2:
                break
        return ShortTermRecommendationResult(
            recommendations=candidates,
            summary={
                "market_rhythm": market.next_day_rhythm,
                "recommendation_count": len(candidates),
            },
        )
