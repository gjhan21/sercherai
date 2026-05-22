from __future__ import annotations

from app.domain.models import MarketSeed
from app.domain.strategies.base import BaseIntradayStrategy


class LimitUpNextDayStrategy(BaseIntradayStrategy):
    """策略F: 涨停板次日 — 早封板+高封单比首板溢价"""

    @property
    def strategy_key(self) -> str:
        return "limit_up_next_day"

    @property
    def strategy_name(self) -> str:
        return "涨停板次日"

    def active_regimes(self) -> set[str]:
        return {"UPTREND", "ROTATION", "EVENT_DRIVEN"}

    def check_entry(self, seed: MarketSeed) -> tuple[bool, str]:
        if not seed.is_limit_up:
            return False, "当日未涨停"
        if not seed.is_natural_limit:
            return False, "一字板排除（次日无交易机会）"
        if seed.is_opened:
            return False, "当日开板过"
        if seed.limit_up_days >= 3:
            return False, f"连板 {seed.limit_up_days} >= 3，风险太高"
        if seed.seal_order_ratio < 0.003:
            return False, f"封单比 {seed.seal_order_ratio:.4f} < 0.3%"
        if seed.avg_turnover20 < 200_000_000:
            return False, "日均成交额 < 2亿"
        if seed.st_risk_proxy:
            return False, "ST 标的排除"
        return True, ""

    def compute_score(self, seed: MarketSeed) -> float:
        score = 0.0

        rank = seed.lu_time_rank
        if rank <= 3:
            score += 35
        elif rank <= 10:
            score += 25
        elif rank <= 30:
            score += 15
        else:
            score += 5

        sr = seed.seal_order_ratio
        if sr > 0.02:
            score += 30
        elif sr > 0.01:
            score += 20
        elif sr > 0.005:
            score += 10
        elif sr > 0.003:
            score += 5

        if seed.limit_up_days == 1:
            score += 15
        elif seed.limit_up_days == 2:
            score += 5

        if seed.news_heat >= 3:
            score += 10

        if seed.net_mf_amount > 0:
            score += 10

        if seed.on_top_list and seed.top_net_amount > 0:
            score += 10

        return self._clamp(score)

    def _build_reasons(self, seed: MarketSeed, score: float) -> list[str]:
        reasons = [f"涨停板次日评分 {score:.1f}"]
        reasons.append(f"涨停时间排名{seed.lu_time_rank}，封单比{seed.seal_order_ratio*100:.1f}%")
        reasons.append(f"{seed.limit_up_days}连板，{'首板' if seed.limit_up_days == 1 else '二板'}优选")
        return reasons
