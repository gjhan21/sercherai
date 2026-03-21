from __future__ import annotations

from app.domain.models import StockFeature


class StockSelector:
    def select(self, features: list[StockFeature], limit: int) -> list[StockFeature]:
        ranked = sorted(
            features,
            key=lambda item: (-item.score, -item.quant_score, -item.momentum20, item.symbol),
        )
        return ranked[:limit]
