from __future__ import annotations

from datetime import UTC, datetime
from typing import Any, Literal

from pydantic import BaseModel, Field, model_validator

from app.schemas.research import MemoryFeedback, ResearchGraphEntity, ResearchGraphRelation
from app.schemas.simulation import ScenarioTemplateConfig, SimulationCard

RiskLevel = Literal["LOW", "MEDIUM", "HIGH"]
SelectionMode = Literal["AUTO", "MANUAL", "DEBUG"]
MarketRegime = Literal["UPTREND", "ROTATION", "EVENT_DRIVEN", "DEFENSIVE", "RISK_OFF"]
PortfolioRole = Literal["CORE", "SATELLITE", "WATCHLIST"]


class StockSelectionPayload(BaseModel):
    run_id: str = ""
    trade_date: str = ""
    selection_mode: SelectionMode = "AUTO"
    universe_scope: str = ""
    profile_id: str = ""
    template_id: str = ""
    template_key: str = ""
    template_name: str = ""
    template_snapshot: dict[str, Any] = Field(default_factory=dict)
    debug_seed_symbols: list[str] = Field(default_factory=list)
    min_listing_days: int = Field(default=180, ge=0, le=3650)
    min_avg_turnover: float = Field(default=50_000_000, ge=0)
    exclude_st: bool = True
    exclude_suspended: bool = True
    price_min: float = Field(default=0, ge=0)
    price_max: float = Field(default=0, ge=0)
    volatility_min: float = Field(default=0, ge=0)
    volatility_max: float = Field(default=0, ge=0)
    industry_whitelist: list[str] = Field(default_factory=list)
    industry_blacklist: list[str] = Field(default_factory=list)
    sector_whitelist: list[str] = Field(default_factory=list)
    sector_blacklist: list[str] = Field(default_factory=list)
    theme_whitelist: list[str] = Field(default_factory=list)
    theme_blacklist: list[str] = Field(default_factory=list)
    limit: int = Field(default=5, ge=1, le=10)
    lookback_days: int = Field(default=120, ge=30, le=365)
    market_scope: str = ""
    seed_symbols: list[str] = Field(default_factory=list)
    excluded_symbols: list[str] = Field(default_factory=list)
    bucket_limit: int = Field(default=36, ge=1, le=120)
    seed_pool_cap: int = Field(default=180, ge=1, le=500)
    candidate_pool_limit: int = Field(default=30, ge=1, le=120)
    watchlist_limit: int = Field(default=5, ge=0, le=20)
    trend_bias: float = Field(default=1.0, ge=0.5, le=1.5)
    money_flow_bias: float = Field(default=1.0, ge=0.5, le=1.5)
    quality_bias: float = Field(default=1.0, ge=0.5, le=1.5)
    event_bias: float = Field(default=1.0, ge=0.5, le=1.5)
    resonance_bias: float = Field(default=1.0, ge=0.5, le=1.5)
    quant_weight: float = Field(default=0.70, ge=0, le=1)
    event_weight: float = Field(default=0.10, ge=0, le=1)
    resonance_weight: float = Field(default=0.10, ge=0, le=1)
    liquidity_risk_weight: float = Field(default=0.10, ge=0, le=1)
    max_symbol_per_bucket: int = Field(default=2, ge=1, le=10)
    max_symbols_per_sector: int = Field(default=2, ge=1, le=10)
    enabled_agents: list[str] = Field(default_factory=list)
    positive_threshold: int = Field(default=3, ge=1, le=5)
    negative_threshold: int = Field(default=2, ge=1, le=5)
    allow_veto: bool = True
    scenario_templates: list[ScenarioTemplateConfig] = Field(default_factory=list)
    max_risk_level: RiskLevel = "MEDIUM"
    min_score: float = Field(default=75, ge=0, le=100)
    compare_with_last_published: bool = False
    dry_run: bool = False

    @model_validator(mode="before")
    @classmethod
    def apply_compat_defaults(cls, raw: Any) -> Any:
        if not isinstance(raw, dict):
            return raw
        data = dict(raw)
        mode = str(data.get("selection_mode", "")).strip().upper()
        if mode not in {"AUTO", "MANUAL", "DEBUG"}:
            if _has_symbols(data.get("debug_seed_symbols")):
                mode = "DEBUG"
            elif _has_symbols(data.get("seed_symbols")):
                mode = "MANUAL"
            else:
                mode = "AUTO"
        data["selection_mode"] = mode
        if not str(data.get("universe_scope", "")).strip():
            data["universe_scope"] = str(data.get("market_scope", "")).strip()
        return data

    def effective_seed_symbols(self) -> list[str]:
        if self.selection_mode == "DEBUG" and self.debug_seed_symbols:
            return _normalize_symbols(self.debug_seed_symbols)
        if self.selection_mode in {"MANUAL", "DEBUG"}:
            return _normalize_symbols(self.seed_symbols)
        return []

    def effective_universe_scope(self) -> str:
        scope = self.universe_scope.strip()
        if scope:
            return scope
        return self.market_scope.strip()


class StockCandidate(BaseModel):
    symbol: str
    name: str
    score: float
    risk_level: RiskLevel
    position_range: str
    reason_summary: str
    invalidations: list[str] = Field(default_factory=list)
    take_profit: str
    stop_loss: str
    evidence_summary: str = ""
    portfolio_role: PortfolioRole = "SATELLITE"
    risk_summary: str = ""
    evidence_cards: list[dict[str, Any]] = Field(default_factory=list)
    positive_reasons: list[str] = Field(default_factory=list)
    veto_reasons: list[str] = Field(default_factory=list)
    theme_tags: list[str] = Field(default_factory=list)
    sector_tags: list[str] = Field(default_factory=list)
    risk_flags: list[str] = Field(default_factory=list)
    evaluation_status: str = "PENDING"


class StockRecommendationWriteModel(BaseModel):
    symbol: str
    name: str
    score: float
    risk_level: RiskLevel
    position_range: str
    valid_from: str
    valid_to: str
    status: str
    reason_summary: str
    source_type: str
    strategy_version: str
    reviewer: str = ""
    publisher: str
    review_note: str = ""
    performance_label: str


class StockRecommendationDetailWriteModel(BaseModel):
    tech_score: float
    fund_score: float
    sentiment_score: float
    money_flow_score: float
    take_profit: str
    stop_loss: str
    risk_note: str


class StockPublishPayload(BaseModel):
    recommendation: StockRecommendationWriteModel
    detail: StockRecommendationDetailWriteModel


class StockStageLog(BaseModel):
    stage_key: str
    stage_order: int
    status: str = "SUCCEEDED"
    input_count: int = 0
    output_count: int = 0
    duration_ms: int = 0
    detail_message: str = ""
    payload_snapshot: dict[str, Any] = Field(default_factory=dict)


class StockCandidateSnapshot(BaseModel):
    symbol: str
    name: str
    stage: str
    quant_score: float
    total_score: float = 0
    risk_level: RiskLevel
    selected: bool = False
    rank: int = 0
    reason_summary: str = ""
    evidence_summary: str = ""
    portfolio_role: str = ""
    risk_summary: str = ""
    factor_breakdown_json: dict[str, Any] = Field(default_factory=dict)


class StockPortfolioEntry(BaseModel):
    symbol: str
    name: str
    rank: int
    quant_score: float
    total_score: float = 0
    risk_level: RiskLevel
    weight_suggestion: str = ""
    reason_summary: str = ""
    evidence_summary: str = ""
    portfolio_role: PortfolioRole = "SATELLITE"
    risk_summary: str = ""
    evidence_cards: list[dict[str, Any]] = Field(default_factory=list)
    positive_reasons: list[str] = Field(default_factory=list)
    veto_reasons: list[str] = Field(default_factory=list)
    theme_tags: list[str] = Field(default_factory=list)
    sector_tags: list[str] = Field(default_factory=list)
    risk_flags: list[str] = Field(default_factory=list)
    evaluation_status: str = "PENDING"
    factor_breakdown_json: dict[str, Any] = Field(default_factory=dict)


class StockEvidenceRecord(BaseModel):
    symbol: str
    name: str
    stage: str
    portfolio_role: str = ""
    evidence_summary: str = ""
    evidence_cards: list[dict[str, Any]] = Field(default_factory=list)
    positive_reasons: list[str] = Field(default_factory=list)
    veto_reasons: list[str] = Field(default_factory=list)
    theme_tags: list[str] = Field(default_factory=list)
    sector_tags: list[str] = Field(default_factory=list)
    risk_flags: list[str] = Field(default_factory=list)


class StockEvaluationRecord(BaseModel):
    symbol: str
    name: str
    horizon_day: int
    evaluation_scope: str
    entry_date: str = ""
    exit_date: str = ""
    entry_price: float = 0
    exit_price: float = 0
    return_pct: float = 0
    excess_return_pct: float = 0
    max_drawdown_pct: float = 0
    hit_flag: bool = False
    benchmark_symbol: str = ""


class StockSelectionReport(BaseModel):
    trade_date: str
    generated_at: str = Field(default_factory=lambda: datetime.now(UTC).isoformat())
    report_summary: str
    risk_summary: str
    selected_count: int
    market_regime: MarketRegime = "ROTATION"
    graph_summary: str = ""
    graph_snapshot_id: str = ""
    consensus_summary: str = ""
    context_meta: dict[str, Any] = Field(default_factory=dict)
    template_snapshot: dict[str, Any] = Field(default_factory=dict)
    evaluation_summary: dict[str, Any] = Field(default_factory=dict)
    related_entities: list[ResearchGraphEntity] = Field(default_factory=list)
    graph_entities: list[ResearchGraphEntity] = Field(default_factory=list)
    graph_relations: list[ResearchGraphRelation] = Field(default_factory=list)
    memory_feedback: MemoryFeedback = Field(default_factory=MemoryFeedback)
    stage_counts: dict[str, int] = Field(default_factory=dict)
    stage_durations_ms: dict[str, int] = Field(default_factory=dict)
    stage_logs: list[StockStageLog] = Field(default_factory=list)
    simulations: list[SimulationCard] = Field(default_factory=list)
    evidence_records: list[StockEvidenceRecord] = Field(default_factory=list)
    evaluation_records: list[StockEvaluationRecord] = Field(default_factory=list)
    candidate_snapshots: list[StockCandidateSnapshot] = Field(default_factory=list)
    portfolio_entries: list[StockPortfolioEntry] = Field(default_factory=list)
    candidates: list[StockCandidate] = Field(default_factory=list)
    publish_payloads: list[StockPublishPayload] = Field(default_factory=list)


def _has_symbols(value: Any) -> bool:
    if not isinstance(value, list):
        return False
    return any(str(item).strip() for item in value)


def _normalize_symbols(items: list[str]) -> list[str]:
    result: list[str] = []
    seen: set[str] = set()
    for item in items:
        symbol = str(item).strip().upper()
        if not symbol or symbol in seen:
            continue
        seen.add(symbol)
        result.append(symbol)
    return result
