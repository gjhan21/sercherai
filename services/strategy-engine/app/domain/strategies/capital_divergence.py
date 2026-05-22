from __future__ import annotations

from app.domain.models import MarketSeed
from app.domain.strategies.base import BaseIntradayStrategy


class CapitalDivergenceStrategy(BaseIntradayStrategy):
    """策略C: 资金背离 — 价跌资金进，主力逆势吸筹"""

    @property
    def strategy_key(self) -> str:
        return "capital_divergence"

    @property
    def strategy_name(self) -> str:
        return "资金背离"

    def active_regimes(self) -> set[str]:
        return {"UPTREND", "ROTATION", "DEFENSIVE"}

    def check_entry(self, seed: MarketSeed) -> tuple[bool, str]:
        if seed.net_mf_amount <= 0:
            return False, f"主力净流入 {seed.net_mf_amount:.0f} <= 0"
        if seed.momentum1 >= 0:
            return False, f"当日涨幅 {seed.momentum1:.2f}% >= 0，无背离"
        if seed.volume_ratio < 0.5:
            return False, f"量比 {seed.volume_ratio:.2f} < 0.5，太低"
        if seed.momentum5 < -10:
            return False, f"近5日跌幅 {seed.momentum5:.1f}% < -10%，可能崩盘"
        if seed.avg_turnover20 < 100_000_000:
            return False, f"日均成交额不足1亿"
        if seed.st_risk_proxy:
            return False, "ST 标的排除"
        return True, ""

    def compute_score(self, seed: MarketSeed) -> float:
        score = 0.0

        # We use consecutive_down_days as proxy for divergence days
        div_days = seed.consecutive_down_days
        if div_days >= 3:
            score += 30
        elif div_days >= 2:
            score += 20
        else:
            score += 10

        nf = seed.net_mf_amount
        if nf > 50_000_000:
            score += 30
        elif nf > 20_000_000:
            score += 20
        elif nf > 5_000_000:
            score += 10

        # Flow ratio proxy: net_mf / (avg_turnover20 * close_price)
        est_amount = seed.avg_turnover20 * seed.close_price if seed.avg_turnover20 > 0 and seed.close_price > 0 else 1e9
        flow_pct = abs(nf) / est_amount * 100 if est_amount > 0 else 0
        if flow_pct > 15:
            score += 20
        elif flow_pct > 8:
            score += 15
        elif flow_pct > 3:
            score += 10

        m1 = abs(seed.momentum1)
        if 0.5 <= m1 <= 3:
            score += 15
        elif 3 < m1 <= 5:
            score += 10

        if seed.on_top_list and seed.top_net_amount > 0:
            score += 15

        return self._clamp(score)

    def _build_reasons(self, seed: MarketSeed, score: float) -> list[str]:
        reasons = [f"资金背离评分 {score:.1f}"]
        reasons.append(f"价跌{abs(seed.momentum1):.1f}%，主力净流入{seed.net_mf_amount:.0f}万")
        if seed.on_top_list:
            reasons.append("龙虎榜加持")
        return reasons
