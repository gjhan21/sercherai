from __future__ import annotations

from app.domain.models import MarketSeed
from app.domain.strategies.base import BaseIntradayStrategy


class StrongPullbackStrategy(BaseIntradayStrategy):
    """策略G: 强势回踩 — 前期强势股回调但支撑有效"""

    @property
    def strategy_key(self) -> str:
        return "strong_pullback"

    @property
    def strategy_name(self) -> str:
        return "强势回踩"

    def active_regimes(self) -> set[str]:
        return {"UPTREND", "ROTATION"}

    def check_entry(self, seed: MarketSeed) -> tuple[bool, str]:
        if seed.momentum20 < 8:
            return False, f"近20日涨幅 {seed.momentum20:.1f}% < 8%"
        if seed.momentum3 > -3 or seed.momentum3 < -10:
            return False, f"近3日跌幅 {seed.momentum3:.1f}% 不在 [-3%, -10%]"
        if seed.deviation_ma20 < -1:
            return False, f"已跌破MA20（偏离{seed.deviation_ma20:.1f}%）"
        if seed.volume_ratio >= 0.7:
            return False, f"量比 {seed.volume_ratio:.2f} >= 0.7"
        if seed.avg_turnover20 < 150_000_000:
            return False, "日均成交额 < 1.5亿"
        if seed.st_risk_proxy:
            return False, "ST 标的排除"
        return True, ""

    def compute_score(self, seed: MarketSeed) -> float:
        score = 0.0

        m20 = seed.momentum20
        if m20 > 25:
            score += 25
        elif m20 > 15:
            score += 20
        elif m20 > 10:
            score += 15
        else:
            score += 10

        vr = seed.volume_ratio
        if vr < 0.3:
            score += 20
        elif vr < 0.5:
            score += 15
        elif vr < 0.7:
            score += 10

        if seed.is_bullish:
            score += 15
        if seed.lower_shadow_pct > seed.upper_shadow_pct * 1.5:
            score += 10

        if abs(seed.deviation_ma20) <= 1:
            score += 15
        elif abs(seed.deviation_ma20) <= 2:
            score += 10

        if seed.theme_tags:
            score += 10

        if seed.net_mf_amount > 20_000_000:
            score += 15
        elif seed.net_mf_amount > 0:
            score += 5

        return self._clamp(score)

    def _build_reasons(self, seed: MarketSeed, score: float) -> list[str]:
        reasons = [f"强势回踩评分 {score:.1f}"]
        reasons.append(f"近20日涨{seed.momentum20:.1f}%，回调{abs(seed.momentum3):.1f}%")
        reasons.append(f"缩量回踩，量比{seed.volume_ratio:.2f}，距MA20偏离{seed.deviation_ma20:.1f}%")
        if seed.is_bullish:
            reasons.append("当日收阳企稳")
        return reasons
