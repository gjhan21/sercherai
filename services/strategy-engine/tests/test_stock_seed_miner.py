from app.domain.features.stock_feature_factory import StockFeatureFactory
from app.domain.seeds.market_seed_loader import DEFAULT_MARKET_SEEDS
from app.domain.seeds.stock_seed_miner import StockSeedMiner


def test_stock_seed_miner_generates_bucketed_seed_pool() -> None:
    features = StockFeatureFactory().build(list(DEFAULT_MARKET_SEEDS))
    result = StockSeedMiner(bucket_limit=2, seed_pool_cap=5).mine(features)

    assert len(result.seed_pool) <= 5
    assert set(result.bucket_members.keys()) == {"trend", "money_flow", "quality", "event", "resonance"}
    assert all(len(items) <= 2 for items in result.bucket_members.values())
    assert len({item.symbol for item in result.seed_pool}) == len(result.seed_pool)
