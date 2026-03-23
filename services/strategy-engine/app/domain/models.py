from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any


@dataclass(slots=True)
class MarketSeed:
    symbol: str
    name: str
    trade_date: str
    close_price: float
    momentum5: float
    momentum20: float
    volatility20: float
    volume_ratio: float
    drawdown20: float
    trend_strength: float
    net_mf_amount: float
    pe_ttm: float
    pb: float
    turnover_rate: float
    news_heat: int
    positive_news_rate: float
    listing_days: int = 0
    avg_turnover20: float = 0.0
    suspended_proxy: bool = False
    st_risk_proxy: bool = False
    industry: str = ""
    sector: str = ""
    theme_tags: list[str] = field(default_factory=list)
    risk_flags: list[str] = field(default_factory=list)


@dataclass(slots=True)
class MarketSeedLoadResult:
    seeds: list[MarketSeed]
    meta: dict[str, Any] = field(default_factory=dict)
    warnings: list[str] = field(default_factory=list)


@dataclass(slots=True)
class StockFeature:
    symbol: str
    name: str
    trade_date: str
    close_price: float
    momentum5: float
    momentum20: float
    volatility20: float
    volume_ratio: float
    drawdown20: float
    trend_strength: float
    net_mf_amount: float
    pe_ttm: float
    pb: float
    turnover_rate: float
    news_heat: int
    positive_news_rate: float
    trend_score: float
    flow_score: float
    value_score: float
    quality_score: float
    news_score: float
    event_score: float
    resonance_score: float
    liquidity_risk_score: float
    quant_score: float
    score: float
    risk_level: str
    reasons: list[str] = field(default_factory=list)
    reason_summary: str = ""
    listing_days: int = 0
    avg_turnover20: float = 0.0
    suspended_proxy: bool = False
    st_risk_proxy: bool = False
    industry: str = ""
    sector: str = ""
    theme_tags: list[str] = field(default_factory=list)
    risk_flags: list[str] = field(default_factory=list)
    risk_adjustment_score: float = 0.0
    positive_reasons: list[str] = field(default_factory=list)
    veto_reasons: list[str] = field(default_factory=list)
    evidence_cards: list[dict[str, Any]] = field(default_factory=list)
    evidence_summary: str = ""
    portfolio_role: str = ""
    watchlist: bool = False

    def factor_breakdown(self) -> dict[str, float]:
        return {
            "trend": round(self.trend_score, 2),
            "money_flow": round(self.flow_score, 2),
            "quality": round(self.quality_score, 2),
            "event": round(self.event_score, 2),
            "resonance": round(self.resonance_score, 2),
            "risk_adjustment": round(self.risk_adjustment_score, 2),
            "total_score": round(self.score, 2),
        }


@dataclass(slots=True)
class FuturesSeed:
    contract: str
    name: str
    trade_date: str
    last_price: float
    basis_pct: float
    volatility14: float
    trend_strength: float
    oi_change_pct: float
    volume_ratio: float
    flow_bias: float
    carry_pct: float
    news_bias: float
    regime: str
    turnover_ratio: float = 0.0
    term_structure_pct: float = 0.0
    curve_slope_pct: float = 0.0
    inventory_level: float = 0.0
    inventory_change_pct: float = 0.0
    inventory_pressure: float = 0.0
    inventory_focus_area: str = ""
    inventory_focus_warehouse: str = ""
    inventory_focus_brand: str = ""
    inventory_focus_place: str = ""
    inventory_focus_grade: str = ""
    inventory_area_share: float = 0.0
    inventory_warehouse_share: float = 0.0
    inventory_brand_share: float = 0.0
    inventory_place_share: float = 0.0
    inventory_grade_share: float = 0.0
    spread_pressure: float = 0.0
    spread_percentile: float = 0.0
    spread_pair: str = ""


@dataclass(slots=True)
class FuturesSeedLoadResult:
    seeds: list[FuturesSeed]
    meta: dict[str, Any] = field(default_factory=dict)
    warnings: list[str] = field(default_factory=list)


@dataclass(slots=True)
class FuturesFeature:
    contract: str
    name: str
    trade_date: str
    last_price: float
    basis_pct: float
    volatility14: float
    trend_strength: float
    oi_change_pct: float
    volume_ratio: float
    flow_bias: float
    carry_pct: float
    news_bias: float
    regime: str
    direction: str
    trend_score: float
    flow_score: float
    carry_score: float
    news_score: float
    conviction_score: float
    risk_level: str
    entry_price: float
    take_profit_price: float
    stop_loss_price: float
    position_range: str = ""
    position_level: str = ""
    turnover_ratio: float = 0.0
    term_structure_pct: float = 0.0
    curve_slope_pct: float = 0.0
    inventory_level: float = 0.0
    inventory_change_pct: float = 0.0
    inventory_pressure: float = 0.0
    inventory_focus_area: str = ""
    inventory_focus_warehouse: str = ""
    inventory_focus_brand: str = ""
    inventory_focus_place: str = ""
    inventory_focus_grade: str = ""
    inventory_area_share: float = 0.0
    inventory_warehouse_share: float = 0.0
    inventory_brand_share: float = 0.0
    inventory_place_share: float = 0.0
    inventory_grade_share: float = 0.0
    spread_pressure: float = 0.0
    spread_percentile: float = 0.0
    spread_pair: str = ""
    reasons: list[str] = field(default_factory=list)
    reason_summary: str = ""
    positive_reasons: list[str] = field(default_factory=list)
    veto_reasons: list[str] = field(default_factory=list)
    evidence_cards: list[dict[str, Any]] = field(default_factory=list)
    evidence_summary: str = ""
    portfolio_role: str = ""
    risk_flags: list[str] = field(default_factory=list)
    related_entities: list[dict[str, Any]] = field(default_factory=list)

    def factor_breakdown(self) -> dict[str, float]:
        return {
            "trend": round(self.trend_score, 2),
            "money_flow": round(self.flow_score, 2),
            "basis_term": round(self.carry_score, 2),
            "event": round(self.news_score, 2),
            "conviction": round(self.conviction_score, 2),
            "total_score": round(self.conviction_score, 2),
        }
