from app.domain.models import MarketSeedLoadResult
from app.domain.seeds.market_seed_loader import DEFAULT_MARKET_SEEDS
from app.domain.universe.stock_universe_builder import StockUniverseBuilder
from app.schemas.stock import StockSelectionPayload


def test_stock_universe_builder_dedupes_and_respects_explicit_exclusion() -> None:
    builder = StockUniverseBuilder()
    result = builder.build(
        StockSelectionPayload(
            trade_date="2026-03-19",
            selection_mode="AUTO",
            excluded_symbols=["601318.SH"],
        ),
        MarketSeedLoadResult(
            seeds=[
                DEFAULT_MARKET_SEEDS[0],
                DEFAULT_MARKET_SEEDS[1],
                DEFAULT_MARKET_SEEDS[0],
            ],
            meta={"source": "go-backend"},
        ),
    )

    assert [item.symbol for item in result.seeds] == ["600519.SH"]
    assert result.meta["universe_count"] == 1
    assert any("excluded" in item for item in result.warnings)
    assert any("duplicated" in item for item in result.warnings)


def test_stock_universe_builder_skips_listing_days_when_go_context_disables_proxy() -> None:
    builder = StockUniverseBuilder()
    result = builder.build(
        StockSelectionPayload(
            trade_date="2026-03-21",
            selection_mode="AUTO",
            min_listing_days=180,
            price_min=5,
            price_max=300,
        ),
        MarketSeedLoadResult(
            seeds=[
                DEFAULT_MARKET_SEEDS[1],
            ],
            meta={
                "selected_trade_date": "2026-03-20",
                "listing_days_filter_applied": False,
                "warnings": ["auto universe skipped min_listing_days proxy because stock truth coverage is only 129 days"],
            },
        ),
    )

    assert [item.symbol for item in result.seeds] == ["601318.SH"]
    assert result.meta["universe_count"] == 1
