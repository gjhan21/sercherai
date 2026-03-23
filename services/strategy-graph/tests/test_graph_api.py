from fastapi.testclient import TestClient

from app.main import build_app
from app.repo.inmemory import InMemoryGraphRepository


def test_graph_snapshot_write_and_read_roundtrip() -> None:
    app = build_app(repo=InMemoryGraphRepository())
    client = TestClient(app)

    write = client.post(
        "/internal/v1/graph/snapshots",
        json={
            "run_id": "ssr_demo_001",
            "asset_domain": "stock",
            "trade_date": "2026-03-22",
            "summary": "股票图谱快照",
            "related_entities": [
                {"entity_type": "Industry", "entity_key": "BANK", "label": "银行"}
            ],
            "entities": [
                {"entity_type": "Stock", "entity_key": "600036.SH", "label": "招商银行", "asset_domain": "stock"},
                {"entity_type": "Industry", "entity_key": "BANK", "label": "银行", "asset_domain": "stock"},
            ],
            "relations": [
                {
                    "relation_type": "BELONGS_TO",
                    "source_type": "Stock",
                    "source_key": "600036.SH",
                    "target_type": "Industry",
                    "target_key": "BANK",
                    "strength": 0.9,
                }
            ],
        },
    )
    assert write.status_code == 200
    payload = write.json()
    assert payload["snapshot_id"]
    assert payload["node_count"] == 2
    assert payload["relation_count"] == 1

    read = client.get(f"/internal/v1/graph/snapshots/{payload['snapshot_id']}")
    assert read.status_code == 200
    snapshot = read.json()
    assert snapshot["run_id"] == "ssr_demo_001"
    assert len(snapshot["entities"]) == 2
    assert snapshot["summary"] == "股票图谱快照"


def test_graph_subgraph_query_returns_neighbors() -> None:
    app = build_app(repo=InMemoryGraphRepository())
    client = TestClient(app)
    client.post(
        "/internal/v1/graph/snapshots",
        json={
            "run_id": "fsr_demo_001",
            "asset_domain": "futures",
            "trade_date": "2026-03-22",
            "summary": "期货图谱快照",
            "entities": [
                {"entity_type": "FuturesContract", "entity_key": "AU2606", "label": "沪金AU2606", "asset_domain": "futures"},
                {"entity_type": "Commodity", "entity_key": "GOLD", "label": "黄金", "asset_domain": "cross"},
                {"entity_type": "Policy", "entity_key": "LIQUIDITY", "label": "流动性政策", "asset_domain": "cross"},
            ],
            "relations": [
                {
                    "relation_type": "BELONGS_TO",
                    "source_type": "FuturesContract",
                    "source_key": "AU2606",
                    "target_type": "Commodity",
                    "target_key": "GOLD",
                },
                {
                    "relation_type": "AFFECTED_BY",
                    "source_type": "FuturesContract",
                    "source_key": "AU2606",
                    "target_type": "Policy",
                    "target_key": "LIQUIDITY",
                },
            ],
        },
    )

    resp = client.get(
        "/internal/v1/graph/subgraph",
        params={"entity_type": "FuturesContract", "entity_key": "AU2606", "depth": 1},
    )
    assert resp.status_code == 200
    payload = resp.json()
    assert payload["entity"]["entity_key"] == "AU2606"
    assert len(payload["entities"]) >= 3
    assert len(payload["relations"]) == 2
    assert payload["backend"] == "inmemory"


def test_graph_subgraph_query_filters_by_entity_asset_domain() -> None:
    app = build_app(repo=InMemoryGraphRepository())
    client = TestClient(app)
    client.post(
        "/internal/v1/graph/snapshots",
        json={
            "run_id": "ssr_demo_cross_001",
            "asset_domain": "stock",
            "trade_date": "2026-03-22",
            "summary": "跨资产实体查询",
            "entities": [
                {
                    "entity_type": "Policy",
                    "entity_key": "REGIME:DEFENSIVE",
                    "label": "DEFENSIVE 市场状态",
                    "asset_domain": "cross",
                },
                {
                    "entity_type": "Stock",
                    "entity_key": "600036",
                    "label": "招商银行",
                    "asset_domain": "stock",
                },
            ],
            "relations": [
                {
                    "relation_type": "CONFIRMS_SIGNAL",
                    "source_type": "Policy",
                    "source_key": "REGIME:DEFENSIVE",
                    "target_type": "Stock",
                    "target_key": "600036",
                    "strength": 0.8,
                }
            ],
        },
    )

    resp = client.get(
        "/internal/v1/graph/subgraph",
        params={
            "entity_type": "Policy",
            "entity_key": "REGIME:DEFENSIVE",
            "asset_domain": "cross",
            "depth": 1,
        },
    )
    assert resp.status_code == 200
    payload = resp.json()
    assert payload["entity"]["entity_key"] == "REGIME:DEFENSIVE"
    assert payload["entity"]["asset_domain"] == "cross"
    assert payload["matched_snapshot_ids"]
    assert len(payload["relations"]) == 1
