import pytest

from app.main import get_container
from app.settings import get_settings


@pytest.fixture(autouse=True)
def _test_settings(monkeypatch):
    monkeypatch.setenv("STRATEGY_ENGINE_ALLOW_SAMPLE_STOCK_SEEDS", "true")
    monkeypatch.setenv("STRATEGY_ENGINE_ALLOW_SAMPLE_FUTURES_SEEDS", "true")
    monkeypatch.delenv("STRATEGY_ENGINE_GO_BACKEND_BASE_URL", raising=False)
    monkeypatch.delenv("STRATEGY_ENGINE_GRAPH_SERVICE_BASE_URL", raising=False)
    get_settings.cache_clear()
    get_container.cache_clear()
    yield
    get_settings.cache_clear()
    get_container.cache_clear()
