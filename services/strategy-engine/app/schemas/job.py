from datetime import datetime
from enum import Enum
from typing import Any
from uuid import uuid4

from pydantic import BaseModel, Field


class JobStatus(str, Enum):
    QUEUED = "QUEUED"
    RUNNING = "RUNNING"
    SUCCEEDED = "SUCCEEDED"
    FAILED = "FAILED"


class JobSubmissionRequest(BaseModel):
    requested_by: str = Field(default="system", min_length=1, max_length=128)
    trace_id: str = Field(default_factory=lambda: f"trace-{uuid4().hex[:12]}")
    payload: dict[str, Any] = Field(default_factory=dict)


class JobAcceptedResponse(BaseModel):
    job_id: str
    job_type: str
    status: JobStatus
    trace_id: str
    created_at: datetime


class JobResult(BaseModel):
    summary: str
    payload_echo: dict[str, Any] = Field(default_factory=dict)
    artifacts: dict[str, Any] = Field(default_factory=dict)
    warnings: list[str] = Field(default_factory=list)


class JobRecord(BaseModel):
    job_id: str
    job_type: str
    status: JobStatus
    requested_by: str
    trace_id: str
    payload: dict[str, Any] = Field(default_factory=dict)
    result: JobResult | None = None
    error_message: str | None = None
    created_at: datetime
    started_at: datetime | None = None
    finished_at: datetime | None = None


class JobListResponse(BaseModel):
    items: list[JobRecord] = Field(default_factory=list)
    total: int
    page: int
    page_size: int


class HealthResponse(BaseModel):
    service: str
    environment: str
    status: str
    supported_job_types: list[str]
    job_counts: dict[str, int]
    now: datetime
