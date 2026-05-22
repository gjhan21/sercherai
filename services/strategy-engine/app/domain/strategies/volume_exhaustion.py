from __future__ import annotations

from app.domain.models import MarketSeed
from app.domain.strategies.base import BaseIntradayStrategy


class VolumeExhaustionStrategy(BaseIntradayStrategy):
    """策略B: 缩量止跌 — 成交量到极值，卖盘枯竭，地量见地价"""

    @property
    def strategy_key(self) -> str:
        return "volume_exhaustion"

    @property
    def strategy_name(self) -> str:
        return "缩量止跌"

    def active_regimes(self) -> set[str]:
        return {"UPTREND", "ROTATION", "DEFENSIVE", "EVENT_DRIVEN"}

    def check_entry(self, seed: MarketSeed) -> tuple[bool, str]:
        cum5d = seed.momentum5
        if cum5d < -3 or cum5d > 1:
            return False, f"近5日涨跌 {cum5d:.1f}% 不在 [-3%, +1%] 范围"
        if seed.volume_20d_min_rank != 1:
            return False, f"当日量非20日最低（排名{seed.volume_20d_min_rank}）"
        if seed.volume_ratio >= 0.5:
            return False, f"量比 {seed.volume_ratio:.2f} >= 0.5"
        # 前10日内出现过连续 >= 2日下跌，确认经历过调整
        if seed.consecutive_down_days < 2:
            return False, f"前10日内未出现连续>=2日下跌（当前连跌{seed.consecutive_down_days}天）"
        if seed.avg_turnover20 < 150_000_000:
            return False, f"日均成交额 {seed.avg_turnover20:.0f} < 1.5亿（流通市值不足50亿）"
        if seed.st_risk_proxy:
            return False, "ST 标的排除"
        return True, ""

    def compute_score(self, seed: MarketSeed) -> float:
        score = 0.0
        score += 30  # 20日最低

        if seed.is_doji:
            score += 20
        elif seed.is_bullish and seed.momentum1 < 1:
            score += 15
        elif not seed.is_bullish and seed.momentum1 > -1:
            score += 5

        dev20 = abs(seed.deviation_ma20)
        if dev20 <= 1:
            score += 15
        elif dev20 <= 2:
            score += 10

        if seed.volume_contraction_days >= 3:
            score += 10

        if seed.net_mf_amount > 10_000_000:
            score += 10
        if seed.net_mf_amount > 0:
            score += 20

        if 0 < seed.pe_ttm < 40:
            score += 5

        return self._clamp(score)

    def _build_reasons(self, seed: MarketSeed, score: float) -> list[str]:
        reasons = [f"缩量止跌评分 {score:.1f}"]
        reasons.append(f"量比为20日最低，量比{seed.volume_ratio:.2f}")
        reasons.append(f"距MA20偏离{seed.deviation_ma20:.1f}%，连续缩量{seed.volume_contraction_days}天")
        if seed.net_mf_amount > 0:
            reasons.append(f"主力净流入{seed.net_mf_amount:.0f}万")
        return reasons
