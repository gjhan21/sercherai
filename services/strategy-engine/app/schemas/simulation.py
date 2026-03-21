from __future__ import annotations

from pydantic import BaseModel, Field


class ScenarioOutcome(BaseModel):
    scenario: str
    thesis: str
    score_adjustment: float
    action: str
    risk_signal: str


class ScenarioTemplateConfig(BaseModel):
    scenario: str
    label: str = ""
    thesis_template: str
    action: str
    risk_signal: str
    score_bias: float = 0


class AgentOpinion(BaseModel):
    agent: str
    stance: str
    confidence: float
    summary: str
    veto: bool = False


class SimulationCard(BaseModel):
    asset_key: str
    asset_type: str
    scenarios: list[ScenarioOutcome] = Field(default_factory=list)
    agents: list[AgentOpinion] = Field(default_factory=list)
    consensus_action: str
    vetoed: bool = False
    veto_reason: str = ""
