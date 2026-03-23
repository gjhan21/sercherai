from __future__ import annotations

from collections import deque
from uuid import uuid4

from app.schemas.graph import (
    GraphEntity,
    GraphRelation,
    GraphSnapshot,
    GraphSnapshotWriteRequest,
    GraphSubgraphResponse,
)


class InMemoryGraphRepository:
    def __init__(self) -> None:
        self._snapshots: dict[str, GraphSnapshot] = {}

    def backend_name(self) -> str:
        return "inmemory"

    def write_snapshot(self, request: GraphSnapshotWriteRequest) -> GraphSnapshot:
        snapshot_id = request.snapshot_id.strip() or f"gss_{uuid4().hex[:16]}"
        snapshot = GraphSnapshot(
            snapshot_id=snapshot_id,
            run_id=request.run_id,
            asset_domain=request.asset_domain,
            trade_date=request.trade_date,
            summary=request.summary,
            related_entities=request.related_entities,
            entities=request.entities,
            relations=request.relations,
            meta=request.meta,
        )
        self._snapshots[snapshot_id] = snapshot
        return snapshot

    def get_snapshot(self, snapshot_id: str) -> GraphSnapshot | None:
        return self._snapshots.get(snapshot_id)

    def query_subgraph(
        self,
        *,
        entity_type: str,
        entity_key: str,
        depth: int,
        asset_domain: str,
    ) -> GraphSubgraphResponse:
        normalized_type = entity_type.strip().upper()
        normalized_key = entity_key.strip().upper()
        normalized_asset_domain = asset_domain.strip().upper()
        snapshots = list(self._snapshots.values())
        entity_index: dict[tuple[str, str], GraphEntity] = {}
        relation_index: list[tuple[str, GraphRelation]] = []
        matched_snapshot_ids: list[str] = []
        root: GraphEntity | None = None

        for snapshot in snapshots:
            entity_map = {
                (item.entity_type.strip().upper(), item.entity_key.strip().upper()): item
                for item in snapshot.entities
            }
            root_entity = entity_map.get((normalized_type, normalized_key))
            if root_entity is None:
                continue
            root_asset_domain = (
                root_entity.asset_domain.strip().upper()
                or snapshot.asset_domain.strip().upper()
            )
            if normalized_asset_domain and root_asset_domain != normalized_asset_domain:
                continue
            matched_snapshot_ids.append(snapshot.snapshot_id)
            if root is None:
                root = root_entity
            entity_index.update(entity_map)
            for relation in snapshot.relations:
                relation_index.append((snapshot.snapshot_id, relation))

        if root is None:
            return GraphSubgraphResponse(backend=self.backend_name())

        adjacency: dict[tuple[str, str], list[tuple[tuple[str, str], GraphRelation]]] = {}
        for _, relation in relation_index:
            source_key = (relation.source_type.strip().upper(), relation.source_key.strip().upper())
            target_key = (relation.target_type.strip().upper(), relation.target_key.strip().upper())
            adjacency.setdefault(source_key, []).append((target_key, relation))
            adjacency.setdefault(target_key, []).append((source_key, relation))

        visited = {(normalized_type, normalized_key)}
        queue = deque([((normalized_type, normalized_key), 0)])
        entities = [root]
        relations: list[GraphRelation] = []
        seen_relations: set[tuple[str, str, str, str, str]] = set()

        while queue:
            current, current_depth = queue.popleft()
            if current_depth >= max(depth, 1):
                continue
            for neighbor, relation in adjacency.get(current, []):
                relation_key = (
                    relation.relation_type,
                    relation.source_type,
                    relation.source_key,
                    relation.target_type,
                    relation.target_key,
                )
                if relation_key not in seen_relations:
                    seen_relations.add(relation_key)
                    relations.append(relation)
                if neighbor in visited:
                    continue
                visited.add(neighbor)
                if neighbor in entity_index:
                    entities.append(entity_index[neighbor])
                queue.append((neighbor, current_depth + 1))

        return GraphSubgraphResponse(
            entity=root,
            entities=entities,
            relations=relations,
            matched_snapshot_ids=matched_snapshot_ids,
            backend=self.backend_name(),
        )
