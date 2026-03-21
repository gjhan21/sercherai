from datetime import UTC, datetime

from fastapi import APIRouter

from app.main import get_container
from app.schemas.job import HealthResponse

router = APIRouter(prefix="/internal/v1", tags=["health"])


@router.get("/health", response_model=HealthResponse)
def get_health() -> HealthResponse:
    container = get_container()
    settings = container.settings
    return HealthResponse(
        service=settings.service_name,
        environment=settings.environment,
        status="ok",
        supported_job_types=list(settings.supported_job_types),
        job_counts=container.job_store.status_counts(),
        now=datetime.now(UTC),
    )
