from fastapi import APIRouter, HTTPException, Query, Request

from app.schemas.graph import (
    GraphSnapshot,
    GraphSnapshotWriteRequest,
    GraphSnapshotWriteResponse,
    GraphSubgraphResponse,
    ReviewedEventWriteRequest,
    ReviewedEventWriteResponse,
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


@router.post("/reviewed-events", response_model=ReviewedEventWriteResponse)
def write_reviewed_event(
    request: Request, payload: ReviewedEventWriteRequest
) -> ReviewedEventWriteResponse:
    repo = request.app.state.repo
    snapshot = repo.write_snapshot(
        GraphSnapshotWriteRequest(
            snapshot_id=f"reviewed-event-{payload.cluster_id.strip()}",
            run_id=payload.cluster_id.strip(),
            asset_domain="stock",
            trade_date=payload.trade_date,
            summary=payload.summary,
            related_entities=payload.entities[: min(len(payload.entities), 12)]
            if payload.approved
            else [],
            entities=payload.entities if payload.approved else [],
            relations=payload.relations if payload.approved else [],
            meta={
                **payload.meta,
                "cluster_id": payload.cluster_id.strip(),
                "source_kind": "reviewed_event",
                "approved": payload.approved,
            },
        )
    )
    return ReviewedEventWriteResponse(
        snapshot_id=snapshot.snapshot_id,
        cluster_id=payload.cluster_id.strip(),
        node_count=len(snapshot.entities),
        relation_count=len(snapshot.relations),
        backend=repo.backend_name(),
    )
