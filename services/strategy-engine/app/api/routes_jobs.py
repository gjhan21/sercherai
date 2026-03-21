from fastapi import APIRouter, BackgroundTasks, HTTPException, Query, status

from app.main import get_container
from app.schemas.job import JobAcceptedResponse, JobListResponse, JobRecord, JobSubmissionRequest

router = APIRouter(prefix="/internal/v1/jobs", tags=["jobs"])


@router.post("/stock-selection", response_model=JobAcceptedResponse, status_code=status.HTTP_202_ACCEPTED)
def create_stock_selection_job(payload: JobSubmissionRequest, background_tasks: BackgroundTasks) -> JobAcceptedResponse:
    return _create_job("stock-selection", payload, background_tasks)


@router.post("/futures-strategy", response_model=JobAcceptedResponse, status_code=status.HTTP_202_ACCEPTED)
def create_futures_strategy_job(payload: JobSubmissionRequest, background_tasks: BackgroundTasks) -> JobAcceptedResponse:
    return _create_job("futures-strategy", payload, background_tasks)


@router.get("", response_model=JobListResponse)
def list_jobs(
    job_type: str = Query(default=""),
    job_status: str = Query(default="", alias="status"),
    page: int = Query(default=1, ge=1),
    page_size: int = Query(default=20, ge=1, le=200),
) -> JobListResponse:
    container = get_container()
    items, total = container.job_store.list_jobs(
        job_type=job_type,
        status=job_status,
        page=page,
        page_size=page_size,
    )
    return JobListResponse(items=items, total=total, page=page, page_size=page_size)


@router.get("/{job_id}", response_model=JobRecord)
def get_job(job_id: str) -> JobRecord:
    container = get_container()
    record = container.job_store.get_job(job_id)
    if record is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="job not found")
    return record


def _create_job(job_type: str, payload: JobSubmissionRequest, background_tasks: BackgroundTasks) -> JobAcceptedResponse:
    container = get_container()
    record = container.job_store.create_job(
        job_type=job_type,
        requested_by=payload.requested_by,
        trace_id=payload.trace_id,
        payload=payload.payload,
    )
    background_tasks.add_task(container.job_runner.run, record.job_id)
    return JobAcceptedResponse(
        job_id=record.job_id,
        job_type=record.job_type,
        status=record.status,
        trace_id=record.trace_id,
        created_at=record.created_at,
    )
