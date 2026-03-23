from fastapi import APIRouter, HTTPException, Query, Request

from app.schemas.graph import (
    GraphSnapshot,
    GraphSnapshotWriteRequest,
    GraphSnapshotWriteResponse,
    GraphSubgraphResponse,
)

router = APIRouter(prefix="/internal/v1/graph", tags=["graph"])


@router.post("/snapshots", response_model=GraphSnapshotWriteResponse)
def write_snapshot(request: Request, payload: GraphSnapshotWriteRequest) -> GraphSnapshotWriteResponse:
    repo = request.app.state.repo
    snapshot = repo.write_snapshot(payload)
    return GraphSnapshotWriteResponse(
        snapshot_id=snapshot.snapshot_id,
        node_count=len(snapshot.entities),
        relation_count=len(snapshot.relations),
        backend=repo.backend_name(),
    )


@router.get("/snapshots/{snapshot_id}", response_model=GraphSnapshot)
def get_snapshot(request: Request, snapshot_id: str) -> GraphSnapshot:
    repo = request.app.state.repo
    snapshot = repo.get_snapshot(snapshot_id)
    if snapshot is None:
        raise HTTPException(status_code=404, detail="graph snapshot not found")
    return snapshot


@router.get("/subgraph", response_model=GraphSubgraphResponse)
def get_subgraph(
    request: Request,
    entity_type: str = Query(..., min_length=1),
    entity_key: str = Query(..., min_length=1),
    depth: int = Query(1, ge=1, le=2),
    asset_domain: str = Query(""),
) -> GraphSubgraphResponse:
    repo = request.app.state.repo
    return repo.query_subgraph(
        entity_type=entity_type,
        entity_key=entity_key,
        depth=depth,
        asset_domain=asset_domain,
    )
