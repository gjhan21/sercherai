from __future__ import annotations

from datetime import UTC, datetime
from typing import Any
from typing import Literal

from pydantic import BaseModel, Field

from app.schemas.simulation import ScenarioTemplateConfig, SimulationCard

RiskLevel = Literal["LOW", "MEDIUM", "HIGH"]
Direction = Literal["LONG", "SHORT", "NEUTRAL"]


class FuturesStrategyPayload(BaseModel):
    trade_date: str = ""
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


class FuturesStrategyReport(BaseModel):
    trade_date: str
    generated_at: str = Field(default_factory=lambda: datetime.now(UTC).isoformat())
    report_summary: str
    risk_summary: str
    selected_count: int
    context_meta: dict[str, Any] = Field(default_factory=dict)
    graph_summary: str = ""
    consensus_summary: str = ""
    simulations: list[SimulationCard] = Field(default_factory=list)
    strategies: list[FuturesStrategyCandidate] = Field(default_factory=list)
    publish_payloads: list[FuturesPublishPayload] = Field(default_factory=list)
