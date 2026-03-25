from __future__ import annotations

from app.domain.models import FuturesFeature


class FuturesSelector:
    def select(self, features: list[FuturesFeature], limit: int) -> list[FuturesFeature]:
        ranked = sorted(
            features,
            key=lambda item: (
                item.direction == "NEUTRAL",
                -item.conviction_score,
                -item.structure_depth_score,
                item.contract,
            ),
        )
        shortlist_size = max(limit * 2, limit)
        return ranked[:shortlist_size]
