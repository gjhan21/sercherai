from __future__ import annotations

from app.domain.models import FuturesFeature
from app.schemas.futures import FuturesStrategyPayload


RISK_ORDER = {"LOW": 0, "MEDIUM": 1, "HIGH": 2}


class LeverageGuard:
    def apply(self, features: list[FuturesFeature], payload: FuturesStrategyPayload) -> tuple[list[FuturesFeature], list[str]]:
        warnings: list[str] = []
        filtered = [item for item in features if item.direction != "NEUTRAL"]
        if not filtered:
            warnings.append("方向信号不足，已回退到中性候选。")
            filtered = list(features)

        max_risk = RISK_ORDER[payload.max_risk_level]
        by_risk = [item for item in filtered if RISK_ORDER.get(item.risk_level, 1) <= max_risk]
        if by_risk:
            filtered = by_risk
        else:
            warnings.append("按风险等级过滤后无结果，已回退到原始期货候选。")

        by_confidence = [item for item in filtered if item.conviction_score >= payload.min_confidence]
        if by_confidence:
            filtered = by_confidence
        else:
            warnings.append("按最小置信度过滤后无结果，已回退到风险过滤结果。")

        result = []
        for item in filtered[: payload.limit]:
            if item.risk_level == "LOW":
                item.position_range = "15%-20%"
                item.position_level = "中仓"
            elif item.risk_level == "HIGH":
                item.position_range = "6%-10%"
                item.position_level = "轻仓"
            else:
                item.position_range = "10%-15%"
                item.position_level = "轻中仓"
            result.append(item)
        return result, warnings
