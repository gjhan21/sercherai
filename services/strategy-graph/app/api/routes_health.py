from fastapi import APIRouter, Request

from app.schemas.graph import HealthResponse

router = APIRouter(tags=["health"])


@router.get("/health", response_model=HealthResponse)
def health(request: Request) -> HealthResponse:
    settings = request.app.state.settings
    repo = request.app.state.repo
    return HealthResponse(
        service=settings.service_name,
        environment=settings.environment,
        status="ok",
        backend=repo.backend_name(),
    )
