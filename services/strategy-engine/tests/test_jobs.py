from fastapi.testclient import TestClient

from app.main import app, get_container


client = TestClient(app)


def setup_function() -> None:
    get_container.cache_clear()


def test_create_stock_selection_job_and_query_status() -> None:
    create_response = client.post(
        "/internal/v1/jobs/stock-selection",
        json={
            "requested_by": "tester",
            "payload": {
                "trade_date": "2026-03-17",
                "limit": 5,
                "max_risk_level": "MEDIUM",
                "min_score": 80,
                "seed_symbols": ["600519.SH", "601318.SH", "600036.SH", "300750.SZ", "688981.SH", "000333.SZ"],
            },
        },
    )

    assert create_response.status_code == 202
    accepted = create_response.json()
    assert accepted["status"] == "QUEUED"

    detail_response = client.get(f"/internal/v1/jobs/{accepted['job_id']}")
    assert detail_response.status_code == 200
    detail = detail_response.json()
    assert detail["job_type"] == "stock-selection"
    assert detail["status"] == "SUCCEEDED"
    assert detail["result"]["payload_echo"]["trade_date"] == "2026-03-17"
    report = detail["result"]["artifacts"]["report"]
    assert report["trade_date"] == "2026-03-17"
    assert report["selected_count"] == 5
    assert len(report["candidates"]) == 5
    assert len(report["publish_payloads"]) == 5
    assert report["stage_counts"]["PORTFOLIO"] == 5
    assert report["market_regime"] in {"UPTREND", "ROTATION", "EVENT_DRIVEN", "DEFENSIVE", "RISK_OFF"}
    assert len(report["stage_logs"]) == 6
    assert len(report["portfolio_entries"]) == 5
    assert len(report["evidence_records"]) >= 5
    assert len(report["simulations"]) == 5
    assert report["graph_summary"]
    assert report["consensus_summary"]
    assert all("invalidations" in item for item in report["candidates"])
    assert all(item["risk_level"] in {"LOW", "MEDIUM"} for item in report["candidates"])
    assert report["publish_payloads"][0]["recommendation"]["source_type"] == "SYSTEM"


def test_create_futures_strategy_job_and_query_status() -> None:
    create_response = client.post(
        "/internal/v1/jobs/futures-strategy",
        json={
            "requested_by": "tester",
            "payload": {
                "trade_date": "2026-03-17",
                "limit": 3,
                "max_risk_level": "HIGH",
                "contracts": ["IF2606", "IH2606", "IC2606", "AU2606"],
            },
        },
    )

    assert create_response.status_code == 202
    accepted = create_response.json()
    assert accepted["job_type"] == "futures-strategy"

    detail_response = client.get(f"/internal/v1/jobs/{accepted['job_id']}")
    assert detail_response.status_code == 200
    detail = detail_response.json()
    assert detail["status"] == "SUCCEEDED"
    report = detail["result"]["artifacts"]["report"]
    assert report["trade_date"] == "2026-03-17"
    assert report["selected_count"] == 3
    assert len(report["strategies"]) == 3
    assert len(report["publish_payloads"]) == 3
    assert len(report["simulations"]) == 3
    assert report["graph_summary"]
    assert report["consensus_summary"]
    assert report["strategies"][0]["direction"] in {"LONG", "SHORT"}
    assert "entry_price" in report["strategies"][0]
    assert "take_profit_price" in report["strategies"][0]
    assert "stop_loss_price" in report["strategies"][0]


def test_list_jobs_supports_filters() -> None:
    client.post(
        "/internal/v1/jobs/stock-selection",
        json={
            "requested_by": "ops-admin",
            "payload": {
                "trade_date": "2026-03-17",
                "limit": 2,
                "max_risk_level": "MEDIUM",
                "min_score": 75,
                "seed_symbols": ["600519.SH", "601318.SH", "600036.SH"],
            },
        },
    )
    client.post(
        "/internal/v1/jobs/futures-strategy",
        json={
            "requested_by": "ops-admin",
            "payload": {
                "trade_date": "2026-03-17",
                "limit": 1,
                "max_risk_level": "HIGH",
                "contracts": ["IF2606", "AU2606"],
            },
        },
    )

    response = client.get(
        "/internal/v1/jobs",
        params={
            "job_type": "stock-selection",
            "status": "SUCCEEDED",
            "page": 1,
            "page_size": 10,
        },
    )

    assert response.status_code == 200
    payload = response.json()
    assert payload["total"] == 1
    assert payload["page"] == 1
    assert payload["page_size"] == 10
    assert len(payload["items"]) == 1
    assert payload["items"][0]["job_type"] == "stock-selection"
    assert payload["items"][0]["status"] == "SUCCEEDED"
    assert payload["items"][0]["result"]["summary"].startswith("stock-selection completed")
