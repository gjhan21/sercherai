from __future__ import annotations

from app.domain.models import StockFeature
from app.schemas.stock import StockSelectionPayload


class StockDecisionFusion:
    def fuse(self, features: list[StockFeature], payload: StockSelectionPayload) -> list[StockFeature]:
        weight_total = (
            payload.quant_weight
            + payload.event_weight
            + payload.resonance_weight
            + payload.liquidity_risk_weight
        )
        if weight_total <= 0:
            quant_weight = 0.70
            event_weight = 0.10
            resonance_weight = 0.10
            risk_weight = 0.10
        else:
            quant_weight = payload.quant_weight / weight_total
            event_weight = payload.event_weight / weight_total
            resonance_weight = payload.resonance_weight / weight_total
            risk_weight = payload.liquidity_risk_weight / weight_total
        for item in features:
            item.score = round(
                item.quant_score * quant_weight
                + item.event_score * event_weight
                + item.resonance_score * resonance_weight
                + item.risk_adjustment_score * risk_weight,
                2,
            )
        return features
