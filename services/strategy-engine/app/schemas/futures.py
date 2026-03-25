from __future__ import annotations

from datetime import UTC, datetime
from typing import Any
from typing import Literal

from pydantic import BaseModel, Field

from app.schemas.research import MemoryFeedback, ResearchGraphEntity, ResearchGraphRelation
from app.schemas.simulation import ScenarioTemplateConfig, SimulationCard

RiskLevel = Literal["LOW", "MEDIUM", "HIGH"]
Direction = Literal["LONG", "SHORT", "NEUTRAL"]


class FuturesStrategyPayload(BaseModel):
    run_id: str = ""
    trade_date: str = ""
    profile_id: str = ""
    template_id: str = ""
    template_key: str = ""
    template_name: str = ""
    template_snapshot: dict[str, Any] = Field(default_factory=dict)
    limit: int = Field(default=3, ge=1, le=5)
    contracts: list[str] = Field(default_factory=list)
    allow_mock_fallback_on_short_history: bool = False
    enabled_agents: list[str] = Field(default_factory=list)
    positive_threshold: int = Field(default=3, ge=1, le=5)
    negative_threshold: int = Field(default=2, ge=1, le=5)
    allow_veto: bool = True
    scenario_templates: list[ScenarioTemplateConfig] = Field(default_factory=list)
    max_risk_level: RiskLevel = "HIGH"
    style: Literal["trend", "balanced"] = "trend"
    min_confidence: float = Field(default=55, ge=0, le=100)
    compare_with_last_published: bool = False
    dry_run: bool = False


class FuturesStrategyCandidate(BaseModel):
    contract: str
    name: str
    direction: Direction
    entry_price: float
    take_profit_price: float
    stop_loss_price: float
    risk_level: RiskLevel
    position_range: str
    reason_summary: str
    invalidations: list[str] = Field(default_factory=list)
    evidence_summary: str = ""
    inventory_factor_summary: str = ""
    structure_factor_summary: str = ""
    portfolio_role: str = "SATELLITE"
    risk_summary: str = ""
    evidence_cards: list[dict[str, Any]] = Field(default_factory=list)
    positive_reasons: list[str] = Field(default_factory=list)
    veto_reasons: list[str] = Field(default_factory=list)
    risk_flags: list[str] = Field(default_factory=list)
    related_entities: list[dict[str, Any]] = Field(default_factory=list)
    evaluation_status: str = "PENDING"


class FuturesStrategyWriteModel(BaseModel):
    contract: str
    name: str
    direction: Direction
    risk_level: RiskLevel
    position_range: str
    valid_from: str
    valid_to: str
    status: str
    reason_summary: str


class FuturesGuidanceWriteModel(BaseModel):
    contract: str
    guidance_direction: Direction
    position_level: str
    entry_range: str
    take_profit_range: str
    stop_loss_range: str
    risk_level: RiskLevel
    invalid_condition: str
    valid_to: str


class FuturesPublishPayload(BaseModel):
    strategy: FuturesStrategyWriteModel
    guidance: FuturesGuidanceWriteModel


class FuturesStageLog(BaseModel):
    stage_key: str
    stage_order: int
    status: str = "SUCCEEDED"
    input_count: int = 0
    output_count: int = 0
    duration_ms: int = 0
    detail_message: str = ""
    payload_snapshot: dict[str, Any] = Field(default_factory=dict)


class FuturesCandidateSnapshot(BaseModel):
    contract: str
    name: str
    stage: str
    score: float
    risk_level: RiskLevel
    selected: bool = False
    rank: int = 0
    direction: Direction = "NEUTRAL"
    reason_summary: str = ""
    evidence_summary: str = ""
    inventory_factor_summary: str = ""
    structure_factor_summary: str = ""
    portfolio_role: str = ""
    risk_summary: str = ""
    factor_breakdown_json: dict[str, Any] = Field(default_factory=dict)


class FuturesPortfolioEntry(BaseModel):
    contract: str
    name: str
    rank: int
    direction: Direction
    score: float
    risk_level: RiskLevel
    position_range: str = ""
    reason_summary: str = ""
    evidence_summary: str = ""
    inventory_factor_summary: str = ""
    structure_factor_summary: str = ""
    portfolio_role: str = "SATELLITE"
    risk_summary: str = ""
    evidence_cards: list[dict[str, Any]] = Field(default_factory=list)
    positive_reasons: list[str] = Field(default_factory=list)
    veto_reasons: list[str] = Field(default_factory=list)
    risk_flags: list[str] = Field(default_factory=list)
    related_entities: list[dict[str, Any]] = Field(default_factory=list)
    evaluation_status: str = "PENDING"
    factor_breakdown_json: dict[str, Any] = Field(default_factory=dict)


class FuturesEvidenceRecord(BaseModel):
    contract: str
    name: str
    stage: str
    portfolio_role: str = ""
    evidence_summary: str = ""
    inventory_factor_summary: str = ""
    structure_factor_summary: str = ""
    evidence_cards: list[dict[str, Any]] = Field(default_factory=list)
    positive_reasons: list[str] = Field(default_factory=list)
    veto_reasons: list[str] = Field(default_factory=list)
    risk_flags: list[str] = Field(default_factory=list)
    related_entities: list[dict[str, Any]] = Field(default_factory=list)


class FuturesEvaluationRecord(BaseModel):
    contract: str
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


class FuturesStrategyReport(BaseModel):
    trade_date: str
    generated_at: str = Field(default_factory=lambda: datetime.now(UTC).isoformat())
    report_summary: str
    risk_summary: str
    selected_count: int
    market_regime: str = "BASE"
    context_meta: dict[str, Any] = Field(default_factory=dict)
    graph_summary: str = ""
    graph_snapshot_id: str = ""
    consensus_summary: str = ""
    template_snapshot: dict[str, Any] = Field(default_factory=dict)
    evaluation_summary: dict[str, Any] = Field(default_factory=dict)
    related_entities: list[ResearchGraphEntity] = Field(default_factory=list)
    graph_entities: list[ResearchGraphEntity] = Field(default_factory=list)
    graph_relations: list[ResearchGraphRelation] = Field(default_factory=list)
    memory_feedback: MemoryFeedback = Field(default_factory=MemoryFeedback)
    stage_counts: dict[str, int] = Field(default_factory=dict)
    stage_durations_ms: dict[str, int] = Field(default_factory=dict)
    stage_logs: list[FuturesStageLog] = Field(default_factory=list)
    simulations: list[SimulationCard] = Field(default_factory=list)
    evidence_records: list[FuturesEvidenceRecord] = Field(default_factory=list)
    evaluation_records: list[FuturesEvaluationRecord] = Field(default_factory=list)
    candidate_snapshots: list[FuturesCandidateSnapshot] = Field(default_factory=list)
    portfolio_entries: list[FuturesPortfolioEntry] = Field(default_factory=list)
    strategies: list[FuturesStrategyCandidate] = Field(default_factory=list)
    publish_payloads: list[FuturesPublishPayload] = Field(default_factory=list)
