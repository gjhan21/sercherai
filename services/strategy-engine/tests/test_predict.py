from fastapi.testclient import TestClient
from unittest.mock import patch
from app.main import app, get_container
from app.domain.models import MarketSeed

client = TestClient(app)


def setup_function() -> None:
    get_container.cache_clear()


def test_predict_endpoint_returns_predictions() -> None:
    # Build 100 dummy seeds to trigger the predictor training logic
    mock_seeds = []
    for i in range(100):
        mock_seeds.append(MarketSeed(
            symbol="600519.SH",
            name="贵州茅台",
            trade_date=f"2026-01-{i:02d}",
            close_price=1700.0 + i * 2.0,
            momentum5=1.0,
            momentum20=2.0,
            volatility20=1.5,
            volume_ratio=1.1,
            drawdown20=2.0,
            trend_strength=0.5,
            net_mf_amount=1000.0,
            pe_ttm=25.0,
            pb=8.0,
            turnover_rate=0.8,
            news_heat=3,
            positive_news_rate=0.6,
            momentum1=0.2,
            momentum2=0.4,
            momentum3=0.6,
            candle_body_pct=0.01,
            lower_shadow_pct=0.005,
            upper_shadow_pct=0.005,
            deviation_ma5=0.1,
            deviation_ma10=0.2,
            deviation_ma20=0.3,
            deviation_ma60=0.4
        ))

    with patch("app.domain.seeds.market_seed_loader.MarketSeedLoader.load_history", return_value=mock_seeds):
        response = client.get("/internal/v1/predict/stock-7d?symbol=600519.SH&trade_date=2026-03-17")
        assert response.status_code == 200
        body = response.json()
        assert "median" in body
        assert "upper_75" in body
        assert "lower_25" in body
        assert len(body["median"]) == 7
        assert len(body["upper_75"]) == 7
        assert len(body["lower_25"]) == 7
        assert "avg_return" in body
        assert "win_rate" in body
