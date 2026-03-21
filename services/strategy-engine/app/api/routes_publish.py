from fastapi import APIRouter, HTTPException, status

from app.main import get_container
from app.schemas.publish import (
    PublishCompareRequest,
    PublishCompareResponse,
    PublishHistoryResponse,
    PublishJobRequest,
    PublishRecord,
    PublishReplay,
)

router = APIRouter(prefix="/internal/v1/publish", tags=["publish"])


@router.post("/jobs/{job_id}", response_model=PublishRecord)
def publish_job(job_id: str, payload: PublishJobRequest | None = None) -> PublishRecord:
    container = get_container()
    job_record = container.job_store.get_job(job_id)
    if job_record is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="job not found")

    try:
        return container.go_backend_publisher.publish_job(job_record, payload)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail=str(exc)) from exc


@router.get("/records/{publish_id}", response_model=PublishRecord)
def get_publish_record(publish_id: str) -> PublishRecord:
    container = get_container()
    record = container.go_backend_publisher.get_record(publish_id)
    if record is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="publish record not found")
    return record


@router.get("/records/{publish_id}/replay", response_model=PublishReplay)
def get_publish_replay(publish_id: str) -> PublishReplay:
    container = get_container()
    record = container.go_backend_publisher.get_record(publish_id)
    if record is None:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail="publish record not found")
    return record.replay


@router.get("/history/{job_type}", response_model=PublishHistoryResponse)
def get_publish_history(job_type: str) -> PublishHistoryResponse:
    container = get_container()
    return PublishHistoryResponse(job_type=job_type, records=container.go_backend_publisher.list_history(job_type))


@router.post("/compare", response_model=PublishCompareResponse)
def compare_publish_versions(payload: PublishCompareRequest) -> PublishCompareResponse:
    container = get_container()
    try:
        return container.go_backend_publisher.compare(payload.left_publish_id, payload.right_publish_id)
    except ValueError as exc:
        raise HTTPException(status_code=status.HTTP_404_NOT_FOUND, detail=str(exc)) from exc
