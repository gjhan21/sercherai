from __future__ import annotations

from app.domain.models import MarketSeed
from app.domain.strategies.base import BaseIntradayStrategy


class BullishEngulfingStrategy(BaseIntradayStrategy):
    """策略E: 阳包阴 — K线吞没反转形态"""

    @property
    def strategy_key(self) -> str:
        return "bullish_engulfing"

    @property
    def strategy_name(self) -> str:
        return "阳包阴"

    def active_regimes(self) -> set[str]:
        return {"UPTREND", "ROTATION", "EVENT_DRIVEN", "DEFENSIVE"}

    def check_entry(self, seed: MarketSeed) -> tuple[bool, str]:
        if not seed.is_engulfing_bullish:
            return False, "不构成阳包阴形态"
        if seed.volume_ratio < 1.2:
            return False, f"量比 {seed.volume_ratio:.2f} < 1.2，未放量确认"
        if seed.st_risk_proxy:
            return False, "ST 标的排除"
        return True, ""

    def compute_score(self, seed: MarketSeed) -> float:
        score = 0.0

        engulf_pct = seed.momentum1
        if engulf_pct > 2:
            score += 25
        elif engulf_pct > 1:
            score += 20
        elif engulf_pct > 0.5:
            score += 15

        vr = seed.volume_ratio
        if vr > 2.5:
            score += 25
        elif vr > 1.8:
            score += 20
        elif vr > 1.2:
            score += 15

        dev20 = seed.deviation_ma20
        if dev20 < -5:
            score += 25
        elif dev20 < -2:
            score += 20
        elif dev20 < 0:
            score += 10

        score += 10  # base engulfing

        if seed.net_mf_amount > 0:
            score += 15

        if seed.upper_shadow_pct < 0.2:
            score += 10

        return self._clamp(score)

    def _build_reasons(self, seed: MarketSeed, score: float) -> list[str]:
        reasons = [f"阳包阴评分 {score:.1f}"]
        reasons.append(f"放量{seed.volume_ratio:.2f}倍吞没，涨幅{seed.momentum1:.1f}%")
        if seed.deviation_ma20 < -2:
            reasons.append(f"低位反转，距MA20偏离{seed.deviation_ma20:.1f}%")
        return reasons
