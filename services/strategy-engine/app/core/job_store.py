from __future__ import annotations

from collections import Counter
from datetime import UTC, datetime
from threading import Lock
from uuid import uuid4

from app.schemas.job import JobRecord, JobResult, JobStatus


class InMemoryJobStore:
    def __init__(self) -> None:
        self._jobs: dict[str, JobRecord] = {}
        self._lock = Lock()

    def create_job(self, job_type: str, requested_by: str, trace_id: str, payload: dict) -> JobRecord:
        now = datetime.now(UTC)
        record = JobRecord(
            job_id=f"job_{uuid4().hex}",
            job_type=job_type,
            status=JobStatus.QUEUED,
            requested_by=requested_by,
            trace_id=trace_id,
            payload=payload,
            created_at=now,
        )
        with self._lock:
            self._jobs[record.job_id] = record
        return record.model_copy(deep=True)

    def get_job(self, job_id: str) -> JobRecord | None:
        with self._lock:
            record = self._jobs.get(job_id)
            if record is None:
                return None
            return record.model_copy(deep=True)

    def list_jobs(
        self,
        *,
        job_type: str = "",
        status: str = "",
        page: int = 1,
        page_size: int = 20,
    ) -> tuple[list[JobRecord], int]:
        page = max(page, 1)
        page_size = max(page_size, 1)
        job_type = job_type.strip().lower()
        status = status.strip().upper()

        with self._lock:
            records = [record.model_copy(deep=True) for record in self._jobs.values()]

        records.sort(key=lambda item: item.created_at, reverse=True)
        filtered: list[JobRecord] = []
        for item in records:
            if job_type and item.job_type.lower() != job_type:
                continue
            if status and item.status.value.upper() != status:
                continue
            filtered.append(item)

        total = len(filtered)
        start = (page - 1) * page_size
        end = start + page_size
        return filtered[start:end], total

    def mark_running(self, job_id: str) -> JobRecord | None:
        with self._lock:
            record = self._jobs.get(job_id)
            if record is None:
                return None
            record.status = JobStatus.RUNNING
            record.started_at = datetime.now(UTC)
            return record.model_copy(deep=True)

    def mark_succeeded(self, job_id: str, result: JobResult) -> JobRecord | None:
        with self._lock:
            record = self._jobs.get(job_id)
            if record is None:
                return None
            record.status = JobStatus.SUCCEEDED
            record.result = result
            record.finished_at = datetime.now(UTC)
            return record.model_copy(deep=True)

    def mark_failed(self, job_id: str, error_message: str) -> JobRecord | None:
        with self._lock:
            record = self._jobs.get(job_id)
            if record is None:
                return None
            record.status = JobStatus.FAILED
            record.error_message = error_message
            record.finished_at = datetime.now(UTC)
            return record.model_copy(deep=True)

    def status_counts(self) -> dict[str, int]:
        with self._lock:
            counts = Counter(job.status.value for job in self._jobs.values())
        return {
            JobStatus.QUEUED.value: counts.get(JobStatus.QUEUED.value, 0),
            JobStatus.RUNNING.value: counts.get(JobStatus.RUNNING.value, 0),
            JobStatus.SUCCEEDED.value: counts.get(JobStatus.SUCCEEDED.value, 0),
            JobStatus.FAILED.value: counts.get(JobStatus.FAILED.value, 0),
        }
