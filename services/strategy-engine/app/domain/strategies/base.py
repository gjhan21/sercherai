from __future__ import annotations

from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from typing import Optional

from app.domain.models import MarketSeed, StockFeature


@dataclass(slots=True)
class StrategyResult:
    strategy_key: str
    strategy_name: str
    candidates: list[StockFeature] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


class BaseIntradayStrategy(ABC):
    """Base class for all T+1 intraday strategies."""

    def __init__(self, max_candidates: int = 10):
        self._max_candidates = max_candidates

    @property
    @abstractmethod
    def strategy_key(self) -> str:
        ...

    @property
    @abstractmethod
    def strategy_name(self) -> str:
        ...

    @abstractmethod
    def check_entry(self, seed: MarketSeed) -> tuple[bool, str]:
        """Return (passed, reason_if_failed)."""
        ...

    @abstractmethod
    def compute_score(self, seed: MarketSeed) -> float:
        """Compute raw score 0-100."""
        ...

    def regime_fit(self, market_regime: str) -> bool:
        """Whether this strategy is active in the given market regime."""
        active_regimes = self.active_regimes()
        if not active_regimes:
            return True
        return market_regime in active_regimes

    def active_regimes(self) -> set[str]:
        """Override to restrict to specific market regimes."""
        return set()

    def run(self, seeds: list[MarketSeed], market_regime: str) -> StrategyResult:
        warnings: list[str] = []
        if not self.regime_fit(market_regime):
            return StrategyResult(
                strategy_key=self.strategy_key,
                strategy_name=self.strategy_name,
                warnings=[f"{self.strategy_name}: market regime {market_regime} not active"],
            )

        passed: list[tuple[MarketSeed, float]] = []
        for seed in seeds:
            ok, reason = self.check_entry(seed)
            if not ok:
                continue
            score = self.compute_score(seed)
            passed.append((seed, score))

        passed.sort(key=lambda x: -x[1])
        passed = passed[: self._max_candidates]

        candidates: list[StockFeature] = []
        for seed, score in passed:
            feature = StockFeature(
                symbol=seed.symbol,
                name=seed.name,
                trade_date=seed.trade_date,
                close_price=seed.close_price,
                momentum5=seed.momentum5,
                momentum20=seed.momentum20,
                volatility20=seed.volatility20,
                volume_ratio=seed.volume_ratio,
                drawdown20=seed.drawdown20,
                trend_strength=seed.trend_strength,
                net_mf_amount=seed.net_mf_amount,
                pe_ttm=seed.pe_ttm,
                pb=seed.pb,
                turnover_rate=seed.turnover_rate,
                news_heat=seed.news_heat,
                positive_news_rate=seed.positive_news_rate,
                trend_score=0,
                flow_score=0,
                value_score=0,
                quality_score=0,
                news_score=0,
                event_score=0,
                resonance_score=0,
                liquidity_risk_score=0,
                quant_score=score,
                score=score,
                risk_level="MEDIUM",
                listing_days=seed.listing_days,
                avg_turnover20=seed.avg_turnover20,
                suspended_proxy=seed.suspended_proxy,
                st_risk_proxy=seed.st_risk_proxy,
                industry=seed.industry,
                sector=seed.sector,
                theme_tags=list(seed.theme_tags),
                risk_flags=list(seed.risk_flags),
                risk_adjustment_score=0.0,
                strategy_name=self.strategy_key,
                raw_intraday_score=score,
                # Intraday / T+1 features — full passthrough from MarketSeed
                momentum1=seed.momentum1,
                momentum2=seed.momentum2,
                momentum3=seed.momentum3,
                consecutive_down_days=seed.consecutive_down_days,
                candle_body_pct=seed.candle_body_pct,
                lower_shadow_pct=seed.lower_shadow_pct,
                upper_shadow_pct=seed.upper_shadow_pct,
                is_bullish=seed.is_bullish,
                is_doji=seed.is_doji,
                is_engulfing_bullish=seed.is_engulfing_bullish,
                deviation_ma5=seed.deviation_ma5,
                deviation_ma10=seed.deviation_ma10,
                deviation_ma20=seed.deviation_ma20,
                deviation_ma60=seed.deviation_ma60,
                volume_20d_min_rank=seed.volume_20d_min_rank,
                volume_contraction_days=seed.volume_contraction_days,
                is_limit_up=seed.is_limit_up,
                lu_time_rank=seed.lu_time_rank,
                seal_order_ratio=seed.seal_order_ratio,
                limit_up_days=seed.limit_up_days,
                is_opened=seed.is_opened,
                is_natural_limit=seed.is_natural_limit,
                on_top_list=seed.on_top_list,
                top_net_amount=seed.top_net_amount,
                top_buy_sell_ratio=seed.top_buy_sell_ratio,
            )
            feature.reasons = self._build_reasons(seed, score)
            feature.reason_summary = "；".join(feature.reasons[:3])
            candidates.append(feature)

        return StrategyResult(
            strategy_key=self.strategy_key,
            strategy_name=self.strategy_name,
            candidates=candidates,
            warnings=warnings,
        )

    def _build_reasons(self, seed: MarketSeed, score: float) -> list[str]:
        return [f"{self.strategy_name}评分 {score:.1f}"]

    @staticmethod
    def _clamp(value: float, lo: float = 0, hi: float = 100) -> float:
        return max(lo, min(hi, value))
