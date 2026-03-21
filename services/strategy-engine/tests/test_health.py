from fastapi.testclient import TestClient

from app.main import app, get_container


client = TestClient(app)


def setup_function() -> None:
    get_container.cache_clear()


def test_health_endpoint_returns_service_status() -> None:
    response = client.get("/internal/v1/health")

    assert response.status_code == 200
    body = response.json()
    assert body["status"] == "ok"
    assert body["service"] == "strategy-engine"
    assert "stock-selection" in body["supported_job_types"]
    assert "futures-strategy" in body["supported_job_types"]
