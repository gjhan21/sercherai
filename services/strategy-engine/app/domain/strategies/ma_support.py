from __future__ import annotations

from app.domain.models import MarketSeed
from app.domain.strategies.base import BaseIntradayStrategy


class MASupportStrategy(BaseIntradayStrategy):
    """策略D: 均线支撑 — 上升趋势中回踩MA20，趋势中继"""

    @property
    def strategy_key(self) -> str:
        return "ma_support"

    @property
    def strategy_name(self) -> str:
        return "均线支撑"

    def active_regimes(self) -> set[str]:
        return {"UPTREND", "ROTATION"}

    def check_entry(self, seed: MarketSeed) -> tuple[bool, str]:
        if seed.trend_strength <= 0:
            return False, f"趋势强度 {seed.trend_strength:.2f} <= 0"
        dev20 = seed.deviation_ma20
        if dev20 < -1 or dev20 > 1:
            return False, f"距MA20偏离 {dev20:.1f}% 不在 [-1%, +1%]"
        if seed.momentum3 > -2 or seed.momentum3 < -8:
            return False, f"近3日跌幅 {seed.momentum3:.1f}% 不在 [-2%, -8%]"
        if seed.volume_ratio >= 0.8:
            return False, f"量比 {seed.volume_ratio:.2f} >= 0.8"
        if seed.momentum1 <= -5:
            return False, "当日跌幅过大，可能砸穿支撑"
        if seed.st_risk_proxy:
            return False, "ST 标的排除"
        return True, ""

    def compute_score(self, seed: MarketSeed) -> float:
        score = 0.0

        ts = seed.trend_strength
        if ts > 5:
            score += 25
        elif ts > 2:
            score += 20
        elif ts > 0:
            score += 15

        vr = seed.volume_ratio
        if vr < 0.4:
            score += 20
        elif vr < 0.6:
            score += 15
        elif vr < 0.8:
            score += 10

        if seed.lower_shadow_pct > seed.candle_body_pct * 2:
            score += 15

        score += 10  # base MA alignment

        if 0 < seed.pe_ttm < 40:
            score += 10

        if seed.net_mf_amount > 0:
            score += 10

        return self._clamp(score)

    def _build_reasons(self, seed: MarketSeed, score: float) -> list[str]:
        reasons = [f"均线支撑评分 {score:.1f}"]
        reasons.append(f"趋势强度{seed.trend_strength:.1f}，回踩MA20偏离{seed.deviation_ma20:.1f}%")
        reasons.append(f"缩量回调，量比{seed.volume_ratio:.2f}")
        return reasons
