from __future__ import annotations

from datetime import UTC, datetime
from typing import Any

from pydantic import BaseModel, Field


class ResearchGraphEntity(BaseModel):
    entity_type: str
    entity_key: str
    label: str
    asset_domain: str = ""
    tags: list[str] = Field(default_factory=list)
    meta: dict[str, Any] = Field(default_factory=dict)


class ResearchGraphRelation(BaseModel):
    relation_type: str
    source_type: str
    source_key: str
    target_type: str
    target_key: str
    strength: float = 1.0
    note: str = ""
    meta: dict[str, Any] = Field(default_factory=dict)


class ResearchGraphSnapshot(BaseModel):
    snapshot_id: str = ""
    run_id: str
    asset_domain: str
    trade_date: str
    summary: str = ""
    related_entities: list[ResearchGraphEntity] = Field(default_factory=list)
    entities: list[ResearchGraphEntity] = Field(default_factory=list)
    relations: list[ResearchGraphRelation] = Field(default_factory=list)
    meta: dict[str, Any] = Field(default_factory=dict)
    created_at: str = Field(default_factory=lambda: datetime.now(UTC).isoformat())


class MemoryFeedbackItem(BaseModel):
    title: str
    level: str = "INFO"
    detail: str
    suggestion: str = ""
    source: str = ""


class MemoryFeedback(BaseModel):
    summary: str = ""
    suggestions: list[str] = Field(default_factory=list)
    failure_signals: list[str] = Field(default_factory=list)
    items: list[MemoryFeedbackItem] = Field(default_factory=list)


class StrategyGraphWriteResult(BaseModel):
    snapshot_id: str = ""
    backend: str = ""
    node_count: int = 0
    relation_count: int = 0
    status: str = "SKIPPED"
    error_message: str = ""
