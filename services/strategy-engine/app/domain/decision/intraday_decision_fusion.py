from __future__ import annotations

from dataclasses import dataclass, field

from app.domain.models import StockFeature
from app.domain.strategies.base import StrategyResult


_STRATEGY_LETTER: dict[str, str] = {
    "oversold_bounce": "A",
    "volume_exhaustion": "B",
    "capital_divergence": "C",
    "ma_support": "D",
    "bullish_engulfing": "E",
    "limit_up_next_day": "F",
    "strong_pullback": "G",
}

_REGIME_WEIGHTS: dict[str, dict[str, float]] = {
    "UPTREND":      {"A": 0.15, "B": 0.10, "C": 0.10, "D": 0.25, "E": 0.10, "F": 0.25, "G": 0.20},
    "ROTATION":     {"A": 0.25, "B": 0.20, "C": 0.20, "D": 0.15, "E": 0.15, "F": 0.10, "G": 0.10},
    "EVENT_DRIVEN": {"A": 0.15, "B": 0.10, "C": 0.15, "D": 0.10, "E": 0.20, "F": 0.25, "G": 0.10},
    "DEFENSIVE":    {"A": 0.0,  "B": 0.30, "C": 0.25, "D": 0.0,  "E": 0.15, "F": 0.0,  "G": 0.0},
    "RISK_OFF":     {"A": 0.0,  "B": 0.0,  "C": 0.0,  "D": 0.0,  "E": 0.0,  "F": 0.0,  "G": 0.0},
}

DEFAULT_MULTI_STRATEGY_BONUS = 5
DEFAULT_MAX_PER_STRATEGY = 2
DEFAULT_MAX_PER_SECTOR = 1
DEFAULT_LIMIT = 5
DEFAULT_MIN_SCORE = 65


@dataclass(slots=True)
class IntradayFusionResult:
    candidates: list[StockFeature] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class IntradayDecisionFusion:
    """T+1 多策略跨策略融合器。

    职责：
    1. 策略内评分归一化（min-max → 0-100）
    2. 按市场环境分配策略权重
    3. 跨策略去重（同一股票取最高分，加多策略共振分）
    4. 分散化约束（同策略/同板块上限）
    5. 输出最终候选列表
    """

    def __init__(
        self,
        multi_strategy_bonus: float = DEFAULT_MULTI_STRATEGY_BONUS,
        max_per_strategy: int = DEFAULT_MAX_PER_STRATEGY,
        max_per_sector: int = DEFAULT_MAX_PER_SECTOR,
        limit: int = DEFAULT_LIMIT,
        min_score: float = DEFAULT_MIN_SCORE,
    ) -> None:
        self._multi_strategy_bonus = multi_strategy_bonus
        self._max_per_strategy = max_per_strategy
        self._max_per_sector = max_per_sector
        self._limit = limit
        self._min_score = min_score

    def fuse(
        self,
        strategy_results: dict[str, StrategyResult],
        market_regime: str,
    ) -> IntradayFusionResult:
        """将各策略候选池融合为统一排名列表。"""
        weights = _REGIME_WEIGHTS.get(market_regime, _REGIME_WEIGHTS["RISK_OFF"])
        warnings: list[str] = []

        # Step 1: 策略内归一化
        normalized: dict[str, list[StockFeature]] = {}
        for key, result in strategy_results.items():
            if not result.candidates:
                continue
            letter = _STRATEGY_LETTER.get(key, "?")
            if weights.get(letter, 0) <= 0:
                continue
            normalized[key] = self._normalize_within_strategy(result.candidates)

        if not normalized:
            return IntradayFusionResult(
                warnings=["融合阶段无有效候选：所有策略在当前市场环境下均关闭或无产出"],
            )

        # Step 2: 加权计算 unified_score
        for key, candidates in normalized.items():
            letter = _STRATEGY_LETTER.get(key, "?")
            w = weights.get(letter, 0)
            for item in candidates:
                item.normalized_intraday_score = item.score
                item.score = round(item.score * w, 2)

        # Step 3: 跨策略去重，追踪多策略共振
        # 同一股票可能被多个策略选中，取最高加权分，记录命中策略数
        merged: dict[str, StockFeature] = {}
        hit_counts: dict[str, int] = {}
        hit_strategies: dict[str, list[str]] = {}
        for key, candidates in normalized.items():
            for item in candidates:
                sym = item.symbol
                if sym not in merged or item.score > merged[sym].score:
                    merged[sym] = item
                    hit_strategies[sym] = [key]
                elif item.score == merged[sym].score:
                    hit_strategies[sym].append(key)
                hit_counts[sym] = hit_counts.get(sym, 0) + 1

        # Apply multi-strategy resonance bonus
        for sym, count in hit_counts.items():
            feat = merged[sym]
            feat.multi_strategy_hit = count
            if count >= 2:
                bonus = self._multi_strategy_bonus
                feat.score = round(feat.score + bonus, 2)
                strategies = " + ".join(hit_strategies.get(sym, []))
                feat.reasons.append(f"多策略共振({count}策略: {strategies}) +{bonus}分")

        # Sort by unified score descending
        ranked = sorted(merged.values(), key=lambda x: -x.score)

        # Step 4: min_score filter
        scored = [item for item in ranked if item.score >= self._min_score]
        if not scored:
            warnings.append(f"最低分数 {self._min_score} 过滤后无候选，已下调至 {self._min_score - 10}")
            scored = [item for item in ranked if item.score >= self._min_score - 10]

        # Step 5: 分散化约束
        diversified = self._diversify(scored)
        if len(diversified) < min(self._limit, len(scored)):
            diversified = self._backfill(diversified, scored, self._limit)

        final = diversified[: self._limit]

        # Assign ranks
        for i, item in enumerate(final):
            item.portfolio_role = "CORE" if i < 2 else "SATELLITE"
            if not item.reason_summary:
                item.reason_summary = "；".join(item.reasons[:3])

        return IntradayFusionResult(candidates=final, warnings=warnings)

    def _normalize_within_strategy(
        self, candidates: list[StockFeature]
    ) -> list[StockFeature]:
        """策略内 min-max 归一化到 0-100。"""
        if len(candidates) <= 1:
            for item in candidates:
                item.score = 100.0 if item.score > 0 else 50.0
            return candidates
        scores = [item.score for item in candidates]
        min_s, max_s = min(scores), max(scores)
        span = max_s - min_s
        for item in candidates:
            normalized = ((item.score - min_s) / span * 100) if span > 0 else 50.0
            item.score = round(normalized, 2)
        return candidates

    def _diversify(self, candidates: list[StockFeature]) -> list[StockFeature]:
        """分散化：同策略上限 + 同板块上限。"""
        selected: list[StockFeature] = []
        strategy_count: dict[str, int] = {}
        sector_count: dict[str, int] = {}
        for item in candidates:
            strat = item.strategy_name or "unknown"
            sector = (item.sector or item.industry or "").strip() or "UNKNOWN"
            if (
                strategy_count.get(strat, 0) < self._max_per_strategy
                and sector_count.get(sector, 0) < self._max_per_sector
            ):
                selected.append(item)
                strategy_count[strat] = strategy_count.get(strat, 0) + 1
                sector_count[sector] = sector_count.get(sector, 0) + 1
            if len(selected) >= self._limit:
                return selected
        return selected

    def _backfill(
        self, current: list[StockFeature], ordered: list[StockFeature], limit: int
    ) -> list[StockFeature]:
        seen = {item.symbol for item in current}
        result = list(current)
        for item in ordered:
            if item.symbol in seen:
                continue
            result.append(item)
            seen.add(item.symbol)
            if len(result) >= limit:
                break
        return result
