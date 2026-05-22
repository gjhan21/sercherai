from __future__ import annotations

from dataclasses import dataclass, field

from app.domain.market.market_daily_analyzer import MarketAnalysisResult
from app.domain.models import StockFeature


@dataclass(slots=True)
class SwingRecommendationResult:
    recommendations: list[StockFeature]
    summary: dict[str, int | str] = field(default_factory=dict)


class SwingRecommendationHead:
    def select(self, features: list[StockFeature], market: MarketAnalysisResult) -> SwingRecommendationResult:
        filtered = [
            item for item in sorted(features, key=lambda feature: (feature.trend_score, feature.momentum20), reverse=True)
            if item.momentum20 > 4 and item.trend_score >= 55 and item.drawdown20 < 10
        ]
        recommendations = filtered[:5]
        for item in recommendations:
            item.recommendation_head = "SWING_AUXILIARY"
            item.selection_layer = "L2_AUXILIARY"
        return SwingRecommendationResult(
            recommendations=recommendations,
            summary={
                "market_trend_state": market.trend_state,
                "recommendation_count": len(recommendations),
            },
        )
