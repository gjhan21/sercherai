from __future__ import annotations

from typing import Protocol

from app.schemas.graph import GraphSnapshot, GraphSnapshotWriteRequest, GraphSubgraphResponse


class GraphRepository(Protocol):
    def backend_name(self) -> str:
        ...

    def write_snapshot(self, request: GraphSnapshotWriteRequest) -> GraphSnapshot:
        ...

    def get_snapshot(self, snapshot_id: str) -> GraphSnapshot | None:
        ...

    def query_subgraph(
        self,
        *,
        entity_type: str,
        entity_key: str,
        depth: int,
        asset_domain: str,
    ) -> GraphSubgraphResponse:
        ...
