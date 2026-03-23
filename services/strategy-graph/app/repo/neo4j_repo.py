from __future__ import annotations

import json
from uuid import uuid4

from neo4j import GraphDatabase

from app.schemas.graph import (
    GraphEntity,
    GraphRelation,
    GraphSnapshot,
    GraphSnapshotWriteRequest,
    GraphSubgraphResponse,
)
from app.settings import Settings


class Neo4jGraphRepository:
    def __init__(self, settings: Settings) -> None:
        self._driver = GraphDatabase.driver(
            settings.neo4j_uri,
            auth=(settings.neo4j_user, settings.neo4j_password),
        )
        self._database = settings.neo4j_database
        self._query_limit = settings.query_limit

    def backend_name(self) -> str:
        return "neo4j"

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
        with self._driver.session(database=self._database) as session:
            session.execute_write(self._write_snapshot_tx, snapshot)
        return snapshot

    @staticmethod
    def _write_snapshot_tx(tx, snapshot: GraphSnapshot) -> None:
        tx.run(
            """
MERGE (s:GraphSnapshot {snapshot_id: $snapshot_id})
SET s.run_id = $run_id,
    s.asset_domain = $asset_domain,
    s.trade_date = $trade_date,
    s.summary = $summary,
    s.meta_json = $meta_json,
    s.created_at = datetime($created_at)
WITH s
OPTIONAL MATCH (s)-[r:CONTAINS]->()
DELETE r
""",
            snapshot_id=snapshot.snapshot_id,
            run_id=snapshot.run_id,
            asset_domain=snapshot.asset_domain,
            trade_date=snapshot.trade_date,
            summary=snapshot.summary,
            meta_json=json.dumps(snapshot.meta, ensure_ascii=False),
            created_at=snapshot.created_at,
        )
        for entity in snapshot.entities:
            tx.run(
                """
MERGE (e:GraphEntity {asset_domain: $asset_domain, entity_type: $entity_type, entity_key: $entity_key})
SET e.label = $label,
    e.tags_json = $tags_json,
    e.meta_json = $meta_json
WITH e
MATCH (s:GraphSnapshot {snapshot_id: $snapshot_id})
MERGE (s)-[:CONTAINS]->(e)
""",
                snapshot_id=snapshot.snapshot_id,
                asset_domain=entity.asset_domain or snapshot.asset_domain,
                entity_type=entity.entity_type,
                entity_key=entity.entity_key,
                label=entity.label,
                tags_json=json.dumps(entity.tags, ensure_ascii=False),
                meta_json=json.dumps(entity.meta, ensure_ascii=False),
            )
        tx.run(
            "MATCH ()-[r:GRAPH_REL {snapshot_id: $snapshot_id}]-() DELETE r",
            snapshot_id=snapshot.snapshot_id,
        )
        for relation in snapshot.relations:
            tx.run(
                """
MATCH (a:GraphEntity {entity_type: $source_type, entity_key: $source_key})
MATCH (b:GraphEntity {entity_type: $target_type, entity_key: $target_key})
MERGE (a)-[r:GRAPH_REL {
  snapshot_id: $snapshot_id,
  relation_type: $relation_type,
  source_type: $source_type,
  source_key: $source_key,
  target_type: $target_type,
  target_key: $target_key
}]->(b)
SET r.strength = $strength,
    r.note = $note,
    r.meta_json = $meta_json
""",
                snapshot_id=snapshot.snapshot_id,
                relation_type=relation.relation_type,
                source_type=relation.source_type,
                source_key=relation.source_key,
                target_type=relation.target_type,
                target_key=relation.target_key,
                strength=relation.strength,
                note=relation.note,
                meta_json=json.dumps(relation.meta, ensure_ascii=False),
            )

    def get_snapshot(self, snapshot_id: str) -> GraphSnapshot | None:
        with self._driver.session(database=self._database) as session:
            return session.execute_read(self._get_snapshot_tx, snapshot_id.strip())

    @staticmethod
    def _get_snapshot_tx(tx, snapshot_id: str) -> GraphSnapshot | None:
        base = tx.run(
            """
MATCH (s:GraphSnapshot {snapshot_id: $snapshot_id})
RETURN s.snapshot_id AS snapshot_id, s.run_id AS run_id, s.asset_domain AS asset_domain,
       s.trade_date AS trade_date, COALESCE(s.summary, '') AS summary,
       COALESCE(s.meta_json, '{}') AS meta_json, toString(s.created_at) AS created_at
LIMIT 1
""",
            snapshot_id=snapshot_id,
        ).single()
        if base is None:
            return None
        entity_rows = tx.run(
            """
MATCH (s:GraphSnapshot {snapshot_id: $snapshot_id})-[:CONTAINS]->(e:GraphEntity)
RETURN e.entity_type AS entity_type, e.entity_key AS entity_key, COALESCE(e.label, '') AS label,
       COALESCE(e.asset_domain, '') AS asset_domain, COALESCE(e.tags_json, '[]') AS tags_json,
       COALESCE(e.meta_json, '{}') AS meta_json
ORDER BY e.entity_type, e.entity_key
""",
            snapshot_id=snapshot_id,
        )
        relation_rows = tx.run(
            """
MATCH (:GraphEntity)-[r:GRAPH_REL {snapshot_id: $snapshot_id}]->(:GraphEntity)
RETURN r.relation_type AS relation_type, r.source_type AS source_type, r.source_key AS source_key,
       r.target_type AS target_type, r.target_key AS target_key, COALESCE(r.strength, 1.0) AS strength,
       COALESCE(r.note, '') AS note, COALESCE(r.meta_json, '{}') AS meta_json
ORDER BY r.relation_type, r.source_key, r.target_key
""",
            snapshot_id=snapshot_id,
        )
        entities = [_row_to_entity(item) for item in entity_rows]
        related_entities = entities[: min(len(entities), 12)]
        relations = [_row_to_relation(item) for item in relation_rows]
        return GraphSnapshot(
            snapshot_id=base["snapshot_id"],
            run_id=base["run_id"],
            asset_domain=base["asset_domain"],
            trade_date=base["trade_date"],
            summary=base["summary"],
            related_entities=related_entities,
            entities=entities,
            relations=relations,
            meta=json.loads(base["meta_json"] or "{}"),
            created_at=base["created_at"],
        )

    def query_subgraph(
        self,
        *,
        entity_type: str,
        entity_key: str,
        depth: int,
        asset_domain: str,
    ) -> GraphSubgraphResponse:
        with self._driver.session(database=self._database) as session:
            return session.execute_read(
                self._query_subgraph_tx,
                entity_type.strip(),
                entity_key.strip(),
                max(depth, 1),
                asset_domain.strip(),
                self._query_limit,
            )

    @staticmethod
    def _query_subgraph_tx(tx, entity_type: str, entity_key: str, depth: int, asset_domain: str, query_limit: int) -> GraphSubgraphResponse:
        match_filter = ""
        params = {
            "entity_type": entity_type,
            "entity_key": entity_key,
            "depth": depth,
            "query_limit": query_limit,
        }
        if asset_domain:
            match_filter = " AND root.asset_domain = $asset_domain"
            params["asset_domain"] = asset_domain
        root_row = tx.run(
            f"""
MATCH (root:GraphEntity {{entity_type: $entity_type, entity_key: $entity_key}})
WHERE 1 = 1{match_filter}
RETURN root.entity_type AS entity_type, root.entity_key AS entity_key, COALESCE(root.label, '') AS label,
       COALESCE(root.asset_domain, '') AS asset_domain, COALESCE(root.tags_json, '[]') AS tags_json,
       COALESCE(root.meta_json, '{{}}') AS meta_json
LIMIT 1
""",
            **params,
        ).single()
        if root_row is None:
            return GraphSubgraphResponse(backend="neo4j")
        graph_rows = tx.run(
            f"""
MATCH (root:GraphEntity {{entity_type: $entity_type, entity_key: $entity_key}})
WHERE 1 = 1{match_filter}
MATCH p=(root)-[rels:GRAPH_REL*1..{depth}]-(neighbor:GraphEntity)
UNWIND nodes(p) AS entity
UNWIND rels AS rel
RETURN DISTINCT
  entity.entity_type AS entity_type,
  entity.entity_key AS entity_key,
  COALESCE(entity.label, '') AS label,
  COALESCE(entity.asset_domain, '') AS asset_domain,
  COALESCE(entity.tags_json, '[]') AS tags_json,
  COALESCE(entity.meta_json, '{{}}') AS meta_json,
  rel.relation_type AS relation_type,
  rel.source_type AS source_type,
  rel.source_key AS source_key,
  rel.target_type AS target_type,
  rel.target_key AS target_key,
  COALESCE(rel.strength, 1.0) AS strength,
  COALESCE(rel.note, '') AS note,
  COALESCE(rel.meta_json, '{{}}') AS relation_meta_json,
  rel.snapshot_id AS snapshot_id
LIMIT $query_limit
""",
            **params,
        )
        entity_map: dict[tuple[str, str], GraphEntity] = {}
        relation_map: dict[tuple[str, str, str, str, str], GraphRelation] = {}
        snapshot_ids: set[str] = set()
        for row in graph_rows:
            entity = GraphEntity(
                entity_type=row["entity_type"],
                entity_key=row["entity_key"],
                label=row["label"],
                asset_domain=row["asset_domain"],
                tags=json.loads(row["tags_json"] or "[]"),
                meta=json.loads(row["meta_json"] or "{}"),
            )
            entity_map[(entity.entity_type, entity.entity_key)] = entity
            relation = GraphRelation(
                relation_type=row["relation_type"],
                source_type=row["source_type"],
                source_key=row["source_key"],
                target_type=row["target_type"],
                target_key=row["target_key"],
                strength=float(row["strength"] or 1.0),
                note=row["note"],
                meta=json.loads(row["relation_meta_json"] or "{}"),
            )
            relation_map[
                (
                    relation.relation_type,
                    relation.source_type,
                    relation.source_key,
                    relation.target_type,
                    relation.target_key,
                )
            ] = relation
            if row["snapshot_id"]:
                snapshot_ids.add(row["snapshot_id"])
        root = _row_to_entity(root_row)
        entity_map[(root.entity_type, root.entity_key)] = root
        return GraphSubgraphResponse(
            entity=root,
            entities=list(entity_map.values()),
            relations=list(relation_map.values()),
            matched_snapshot_ids=sorted(snapshot_ids),
            backend="neo4j",
        )


def _row_to_entity(row) -> GraphEntity:
    return GraphEntity(
        entity_type=row["entity_type"],
        entity_key=row["entity_key"],
        label=row["label"],
        asset_domain=row["asset_domain"],
        tags=json.loads(row["tags_json"] or "[]"),
        meta=json.loads(row["meta_json"] or "{}"),
    )


def _row_to_relation(row) -> GraphRelation:
    return GraphRelation(
        relation_type=row["relation_type"],
        source_type=row["source_type"],
        source_key=row["source_key"],
        target_type=row["target_type"],
        target_key=row["target_key"],
        strength=float(row["strength"] or 1.0),
        note=row["note"],
        meta=json.loads(row["meta_json"] or "{}"),
    )
