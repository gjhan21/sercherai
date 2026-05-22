from __future__ import annotations

from app.domain.models import MarketSeed
from app.domain.strategies.base import BaseIntradayStrategy


class OversoldBounceStrategy(BaseIntradayStrategy):
    """策略A: 超跌反弹 — 连续下跌后卖压衰竭，技术性反弹"""

    @property
    def strategy_key(self) -> str:
        return "oversold_bounce"

    @property
    def strategy_name(self) -> str:
        return "超跌反弹"

    def active_regimes(self) -> set[str]:
        return {"UPTREND", "ROTATION", "EVENT_DRIVEN"}

    def check_entry(self, seed: MarketSeed) -> tuple[bool, str]:
        if seed.consecutive_down_days < 3:
            return False, f"连跌天数 {seed.consecutive_down_days} < 3"
        cum3d = seed.momentum3
        if cum3d > -4 or cum3d < -15:
            return False, f"近3日累计跌幅 {cum3d:.1f}% 不在 [-4%, -15%] 范围"
        if seed.volume_ratio >= 0.6:
            return False, f"量比 {seed.volume_ratio:.2f} >= 0.6，未缩量"
        if seed.momentum1 <= -7:
            return False, f"当日跌幅 {seed.momentum1:.1f}% <= -7%，可能跌停"
        if seed.st_risk_proxy:
            return False, "ST 标的排除"
        if seed.avg_turnover20 < 150_000_000:
            return False, f"日均成交额 {seed.avg_turnover20:.0f} < 1.5亿"
        return True, ""

    def compute_score(self, seed: MarketSeed) -> float:
        score = 0.0
        down = seed.consecutive_down_days
        if down == 4:
            score += 25
        elif down == 5:
            score += 20
        elif down == 3:
            score += 15
        elif down == 6:
            score += 10
        else:
            score += 5

        vr = seed.volume_ratio
        if vr < 0.3:
            score += 30
        elif vr < 0.5:
            score += 20
        elif vr < 0.6:
            score += 10

        if seed.lower_shadow_pct > seed.upper_shadow_pct * 2:
            score += 20 if seed.is_bullish else 10

        dev20 = seed.deviation_ma20
        if dev20 < -8:
            score += 25
        elif dev20 < -5:
            score += 20
        elif dev20 < -3:
            score += 15
        elif dev20 < -1:
            score += 10

        cum3d = seed.momentum3
        if -10 <= cum3d <= -6:
            score += 20
        elif -6 < cum3d <= -4:
            score += 15
        elif -15 <= cum3d < -10:
            score += 10

        if seed.is_bullish:
            score += 10

        return self._clamp(score)

    def _build_reasons(self, seed: MarketSeed, score: float) -> list[str]:
        reasons = [f"超跌反弹评分 {score:.1f}"]
        reasons.append(f"连跌{seed.consecutive_down_days}天，近3日跌{seed.momentum3:.1f}%")
        reasons.append(f"量比{seed.volume_ratio:.2f}，距MA20偏离{seed.deviation_ma20:.1f}%")
        if seed.is_bullish:
            reasons.append("当日收阳，有企稳迹象")
        return reasons
