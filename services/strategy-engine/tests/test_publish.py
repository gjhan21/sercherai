from fastapi.testclient import TestClient

from app.main import app, get_container
from app.schemas.job import JobResult


client = TestClient(app)


def setup_function() -> None:
    get_container.cache_clear()


def test_publish_stock_job_and_query_replay_history() -> None:
    accepted = client.post(
        "/internal/v1/jobs/stock-selection",
        json={
            "requested_by": "tester",
            "payload": {
                "trade_date": "2026-03-17",
                "limit": 3,
                "max_risk_level": "MEDIUM",
                "seed_symbols": ["600519.SH", "601318.SH", "300750.SZ"],
            },
        },
    ).json()

    publish_response = client.post(f"/internal/v1/publish/jobs/{accepted['job_id']}")
    assert publish_response.status_code == 200
    record = publish_response.json()
    assert record["job_type"] == "stock-selection"
    assert record["version"] == 1
    assert record["payload_count"] == 3
    assert record["markdown"]
    assert record["html"].startswith("<html>")

    replay_response = client.get(f"/internal/v1/publish/records/{record['publish_id']}/replay")
    assert replay_response.status_code == 200
    replay = replay_response.json()
    assert replay["warning_count"] >= 0
    assert len(replay["invalidated_assets"]) == 3

    history_response = client.get("/internal/v1/publish/history/stock-selection")
    assert history_response.status_code == 200
    history = history_response.json()
    assert history["job_type"] == "stock-selection"
    assert len(history["records"]) == 1
    assert history["records"][0]["publish_id"] == record["publish_id"]


def test_compare_two_published_stock_versions() -> None:
    first_job = client.post(
        "/internal/v1/jobs/stock-selection",
        json={
            "requested_by": "tester",
            "payload": {
                "trade_date": "2026-03-17",
                "limit": 2,
                "max_risk_level": "MEDIUM",
                "seed_symbols": ["600519.SH", "601318.SH"],
            },
        },
    ).json()
    second_job = client.post(
        "/internal/v1/jobs/stock-selection",
        json={
            "requested_by": "tester",
                "payload": {
                    "trade_date": "2026-03-17",
                    "limit": 2,
                    "max_risk_level": "MEDIUM",
                    "seed_symbols": ["600519.SH", "600036.SH"],
                },
            },
        ).json()

    first_publish = client.post(f"/internal/v1/publish/jobs/{first_job['job_id']}").json()
    second_publish = client.post(f"/internal/v1/publish/jobs/{second_job['job_id']}").json()

    compare_response = client.post(
        "/internal/v1/publish/compare",
        json={
            "left_publish_id": first_publish["publish_id"],
            "right_publish_id": second_publish["publish_id"],
        },
    )
    assert compare_response.status_code == 200
    compare = compare_response.json()
    assert compare["left_version"] == 1
    assert compare["right_version"] == 2
    assert compare["added_assets"] == ["600036.SH"]
    assert compare["removed_assets"] == ["601318.SH"]
    assert compare["shared_assets"] == ["600519.SH"]


def test_publish_policy_can_block_and_force_publish() -> None:
    container = get_container()
    created = container.job_store.create_job(
        job_type="stock-selection",
        requested_by="tester",
        trace_id="trace-block-001",
        payload={"trade_date": "2026-03-17"},
    )
    container.job_store.mark_succeeded(
        created.job_id,
        JobResult(
            summary="manual stock report",
            artifacts={
                "report": {
                    "trade_date": "2026-03-17",
                    "report_summary": "测试发布策略拦截",
                    "selected_count": 1,
                    "candidates": [
                        {
                            "symbol": "600519.SH",
                            "risk_level": "HIGH",
                            "invalidations": ["失效条件"],
                        }
                    ],
                    "publish_payloads": [
                        {
                            "recommendation": {"symbol": "600519.SH"},
                        }
                    ],
                    "simulations": [
                        {
                            "asset_key": "600519.SH",
                            "vetoed": True,
                        }
                    ],
                }
            },
            warnings=["高风险提醒"],
        ),
    )

    blocked = client.post(
        f"/internal/v1/publish/jobs/{created.job_id}",
        json={
            "requested_by": "ops-admin",
            "policy": {
                "max_risk_level": "MEDIUM",
                "max_warning_count": 0,
                "allow_vetoed_publish": False,
            },
        },
    )
    assert blocked.status_code == 409
    assert "发布策略拦截" in blocked.json()["detail"]

    forced = client.post(
        f"/internal/v1/publish/jobs/{created.job_id}",
        json={
            "requested_by": "ops-admin",
            "force": True,
            "override_reason": "人工确认题材驱动有效",
            "policy": {
                "max_risk_level": "MEDIUM",
                "max_warning_count": 0,
                "allow_vetoed_publish": False,
                "override_note_template": "覆盖后需重点观察次日表现。",
            },
        },
    )
    assert forced.status_code == 200
    record = forced.json()
    assert any("人工覆盖发布原因" in item for item in record["replay"]["notes"])
    assert any("覆盖后需重点观察次日表现" in item for item in record["replay"]["notes"])
