from __future__ import annotations

from datetime import datetime
from typing import Any

from pydantic import BaseModel, Field


class PublishReplay(BaseModel):
    warning_count: int = 0
    warning_messages: list[str] = Field(default_factory=list)
    vetoed_assets: list[str] = Field(default_factory=list)
    invalidated_assets: list[str] = Field(default_factory=list)
    notes: list[str] = Field(default_factory=list)


class PublishPolicyConfig(BaseModel):
    max_risk_level: str = "MEDIUM"
    max_warning_count: int = 3
    allow_vetoed_publish: bool = False
    default_publisher: str = "strategy-engine"
    override_note_template: str = ""


class PublishJobRequest(BaseModel):
    requested_by: str = "system"
    force: bool = False
    override_reason: str = ""
    policy: PublishPolicyConfig = Field(default_factory=PublishPolicyConfig)


class PublishRecord(BaseModel):
    publish_id: str
    job_id: str
    job_type: str
    version: int
    created_at: datetime
    trade_date: str = ""
    report_summary: str
    selected_count: int = 0
    asset_keys: list[str] = Field(default_factory=list)
    payload_count: int = 0
    markdown: str
    html: str
    publish_payloads: list[dict[str, Any]] = Field(default_factory=list)
    report_snapshot: dict[str, Any] = Field(default_factory=dict)
    replay: PublishReplay = Field(default_factory=PublishReplay)


class PublishRecordSummary(BaseModel):
    publish_id: str
    job_id: str
    job_type: str
    version: int
    created_at: datetime
    trade_date: str = ""
    report_summary: str
    selected_count: int = 0
    asset_keys: list[str] = Field(default_factory=list)
    payload_count: int = 0


class PublishHistoryResponse(BaseModel):
    job_type: str
    records: list[PublishRecordSummary] = Field(default_factory=list)


class PublishCompareRequest(BaseModel):
    left_publish_id: str
    right_publish_id: str


class PublishCompareResponse(BaseModel):
    left_publish_id: str
    right_publish_id: str
    left_version: int
    right_version: int
    selected_count_delta: int
    payload_count_delta: int
    warning_count_delta: int
    veto_count_delta: int
    added_assets: list[str] = Field(default_factory=list)
    removed_assets: list[str] = Field(default_factory=list)
    shared_assets: list[str] = Field(default_factory=list)
