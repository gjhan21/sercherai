from app.domain.features.stock_feature_factory import StockFeatureFactory
from app.domain.heads.swing_recommendation_head import SwingRecommendationHead
from app.domain.market.market_daily_analyzer import MarketDailyAnalyzer
from app.domain.models import MarketSeed


def _seed(symbol: str, *, momentum20: float, trend_strength: float, volume_ratio: float, drawdown20: float = 5.0) -> MarketSeed:
    return MarketSeed(
        symbol=symbol,
        name=symbol,
        trade_date="2026-05-20",
        close_price=10.0,
        momentum5=3.0,
        momentum20=momentum20,
        volatility20=2.1,
        volume_ratio=volume_ratio,
        drawdown20=drawdown20,
        trend_strength=trend_strength,
        net_mf_amount=1_200_000.0,
        pe_ttm=20.0,
        pb=2.0,
        turnover_rate=1.5,
        news_heat=1,
        positive_news_rate=0.6,
        avg_turnover20=900_000_000,
        industry="电子",
        sector="芯片",
    )


def test_swing_head_outputs_fixed_auxiliary_candidates_from_shared_pool() -> None:
    seeds = [
        _seed("A", momentum20=13.0, trend_strength=82.0, volume_ratio=1.5),
        _seed("B", momentum20=11.0, trend_strength=79.0, volume_ratio=1.4),
        _seed("C", momentum20=9.0, trend_strength=75.0, volume_ratio=1.3),
        _seed("D", momentum20=8.0, trend_strength=73.0, volume_ratio=1.2),
        _seed("E", momentum20=7.5, trend_strength=71.0, volume_ratio=1.1),
        _seed("F", momentum20=2.0, trend_strength=45.0, volume_ratio=0.9, drawdown20=10.0),
    ]
    features = StockFeatureFactory().build(seeds)
    market = MarketDailyAnalyzer().analyze(seeds)

    result = SwingRecommendationHead().select(features, market)

    assert 3 <= len(result.recommendations) <= 5
    assert all(item.recommendation_head == "SWING_AUXILIARY" for item in result.recommendations)
    assert "F" not in {item.symbol for item in result.recommendations}
