from app.domain.models import MarketSeedLoadResult
from app.domain.seeds.market_seed_loader import MarketSeedLoader
from app.schemas.stock import StockSelectionPayload
from app.settings import Settings


def test_market_seed_loader_uses_sample_fallback_when_enabled() -> None:
    loader = MarketSeedLoader(
        settings=Settings(
            go_backend_base_url="",
            allow_sample_stock_seeds=True,
        )
    )

    result = loader.load(
        StockSelectionPayload(
            trade_date="2026-03-17",
            seed_symbols=["600519.SH", "601318.SH"],
            excluded_symbols=["601318.SH"],
            limit=5,
        )
    )

    assert isinstance(result, MarketSeedLoadResult)
    assert [item.symbol for item in result.seeds] == ["600519.SH"]
    assert result.meta["source"] == "sample-fallback"
    assert any("fallback to sample seeds" in item for item in result.warnings)


def test_market_seed_loader_reads_go_backend_context(monkeypatch) -> None:
    loader = MarketSeedLoader(
        settings=Settings(
            go_backend_base_url="http://127.0.0.1:8080",
            allow_sample_stock_seeds=False,
        )
    )

    def fake_post(url: str, payload: dict):
        assert url == "http://127.0.0.1:8080/internal/v1/strategy-engine/context/stock-selection"
        assert payload["trade_date"] == "2026-03-19"
        assert payload["selection_mode"] == "MANUAL"
        assert payload["seed_symbols"] == ["600519.SH"]
        assert payload["universe_scope"] == ""
        assert payload["profile_id"] == ""
        return {
            "seeds": [
                {
                    "symbol": "600519.SH",
                    "name": "贵州茅台",
                    "trade_date": "2026-03-18",
                    "close_price": 1710.2,
                    "momentum5": 2.1,
                    "momentum20": 6.8,
                    "volatility20": 1.6,
                    "volume_ratio": 1.12,
                    "drawdown20": 3.4,
                    "trend_strength": 2.2,
                    "net_mf_amount": 9800.5,
                    "pe_ttm": 25.8,
                    "pb": 9.2,
                    "turnover_rate": 0.8,
                    "news_heat": 3,
                    "positive_news_rate": 0.67,
                }
            ],
            "meta": {
                "selected_trade_date": "2026-03-18",
                "price_source": "TUSHARE",
                "news_window_days": 14,
                "listing_days_filter_applied": False,
                "warnings": ["market news partially missing"],
            },
        }

    monkeypatch.setattr(loader, "_post_context_request", fake_post)

    result = loader.load(
        StockSelectionPayload(
            trade_date="2026-03-19",
            seed_symbols=["600519.SH"],
            limit=3,
        )
    )

    assert len(result.seeds) == 1
    assert result.seeds[0].symbol == "600519.SH"
    assert result.meta["selected_trade_date"] == "2026-03-18"
    assert result.meta["listing_days_filter_applied"] is False
    assert result.warnings == ["market news partially missing"]


def test_market_seed_loader_prefers_debug_seed_symbols(monkeypatch) -> None:
    loader = MarketSeedLoader(
        settings=Settings(
            go_backend_base_url="",
            allow_sample_stock_seeds=True,
        )
    )

    result = loader.load(
        StockSelectionPayload(
            trade_date="2026-03-19",
            selection_mode="DEBUG",
            debug_seed_symbols=["300750.SZ"],
            seed_symbols=["600519.SH", "601318.SH"],
            limit=3,
        )
    )

    assert [item.symbol for item in result.seeds] == ["300750.SZ"]


def test_market_seed_loader_reads_auto_universe_from_go_backend(monkeypatch) -> None:
    loader = MarketSeedLoader(
        settings=Settings(
            go_backend_base_url="http://127.0.0.1:8080",
            allow_sample_stock_seeds=False,
        )
    )

    def fake_post(url: str, payload: dict):
        assert url == "http://127.0.0.1:8080/internal/v1/strategy-engine/context/stock-selection"
        assert payload["selection_mode"] == "AUTO"
        assert payload["seed_symbols"] == []
        return {
            "seeds": [
                {
                    "symbol": "600519.SH",
                    "name": "贵州茅台",
                    "trade_date": "2026-03-18",
                    "close_price": 1710.2,
                    "momentum5": 2.1,
                    "momentum20": 6.8,
                    "volatility20": 1.6,
                    "volume_ratio": 1.12,
                    "drawdown20": 3.4,
                    "trend_strength": 2.2,
                    "net_mf_amount": 9800.5,
                    "pe_ttm": 25.8,
                    "pb": 9.2,
                    "turnover_rate": 0.8,
                    "news_heat": 3,
                    "positive_news_rate": 0.67,
                },
                {
                    "symbol": "601318.SH",
                    "name": "中国平安",
                    "trade_date": "2026-03-18",
                    "close_price": 52.6,
                    "momentum5": 2.4,
                    "momentum20": 7.1,
                    "volatility20": 1.8,
                    "volume_ratio": 1.25,
                    "drawdown20": 3.0,
                    "trend_strength": 2.0,
                    "net_mf_amount": 8800.3,
                    "pe_ttm": 11.0,
                    "pb": 1.3,
                    "turnover_rate": 1.1,
                    "news_heat": 2,
                    "positive_news_rate": 0.62,
                },
            ],
            "meta": {
                "selected_trade_date": "2026-03-18",
                "price_source": "TUSHARE",
                "news_window_days": 14,
            },
        }

    monkeypatch.setattr(loader, "_post_context_request", fake_post)

    result = loader.load(
        StockSelectionPayload(
            trade_date="2026-03-19",
            selection_mode="AUTO",
            limit=5,
        )
    )

    assert [item.symbol for item in result.seeds] == ["600519.SH", "601318.SH"]
    assert result.meta["selected_trade_date"] == "2026-03-18"
