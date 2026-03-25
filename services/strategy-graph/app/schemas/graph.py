from __future__ import annotations

from datetime import UTC, datetime
from typing import Any

from pydantic import BaseModel, Field


class GraphEntity(BaseModel):
    entity_type: str
    entity_key: str
    label: str
    asset_domain: str = ""
    tags: list[str] = Field(default_factory=list)
    meta: dict[str, Any] = Field(default_factory=dict)


class GraphRelation(BaseModel):
    relation_type: str
    source_type: str
    source_key: str
    target_type: str
    target_key: str
    strength: float = 1.0
    note: str = ""
    meta: dict[str, Any] = Field(default_factory=dict)


class GraphSnapshotWriteRequest(BaseModel):
    snapshot_id: str = ""
    run_id: str
    asset_domain: str
    trade_date: str
    summary: str = ""
    related_entities: list[GraphEntity] = Field(default_factory=list)
    entities: list[GraphEntity] = Field(default_factory=list)
    relations: list[GraphRelation] = Field(default_factory=list)
    meta: dict[str, Any] = Field(default_factory=dict)


class GraphSnapshot(BaseModel):
    snapshot_id: str
    run_id: str
    asset_domain: str
    trade_date: str
    summary: str = ""
    related_entities: list[GraphEntity] = Field(default_factory=list)
    entities: list[GraphEntity] = Field(default_factory=list)
    relations: list[GraphRelation] = Field(default_factory=list)
    meta: dict[str, Any] = Field(default_factory=dict)
    created_at: str = Field(default_factory=lambda: datetime.now(UTC).isoformat())


class GraphSnapshotWriteResponse(BaseModel):
    snapshot_id: str
    node_count: int
    relation_count: int
    backend: str


class GraphSubgraphResponse(BaseModel):
    entity: GraphEntity | None = None
    entities: list[GraphEntity] = Field(default_factory=list)
    relations: list[GraphRelation] = Field(default_factory=list)
    matched_snapshot_ids: list[str] = Field(default_factory=list)
    backend: str


class ReviewedEventWriteRequest(BaseModel):
    cluster_id: str
    approved: bool = True
    trade_date: str = ""
    summary: str = ""
    entities: list[GraphEntity] = Field(default_factory=list)
    relations: list[GraphRelation] = Field(default_factory=list)
    meta: dict[str, Any] = Field(default_factory=dict)


class ReviewedEventWriteResponse(BaseModel):
    snapshot_id: str
    cluster_id: str
    node_count: int
    relation_count: int
    backend: str


class HealthResponse(BaseModel):
    service: str
    environment: str
    status: str
    backend: str
    now: str = Field(default_factory=lambda: datetime.now(UTC).isoformat())
