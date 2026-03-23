from __future__ import annotations

from functools import lru_cache

from fastapi import FastAPI

from app.api.routes_graph import router as graph_router
from app.api.routes_health import router as health_router
from app.repo.inmemory import InMemoryGraphRepository
from app.repo.neo4j_repo import Neo4jGraphRepository
from app.settings import Settings, get_settings


@lru_cache(maxsize=1)
def get_repo():
    settings = get_settings()
    if settings.neo4j_uri.strip():
        return Neo4jGraphRepository(settings)
    return InMemoryGraphRepository()


def build_app(settings: Settings | None = None, repo=None) -> FastAPI:
    app = FastAPI(
        title="strategy-graph",
        version="0.1.0",
        description="Internal graph service for strategy research snapshots",
    )
    app.state.settings = settings or get_settings()
    app.state.repo = repo or get_repo()
    app.include_router(health_router)
    app.include_router(graph_router)
    return app


app = build_app()
