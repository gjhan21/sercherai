from app.domain.features.stock_feature_factory import StockFeatureFactory
from app.domain.heads.short_term_recommendation_head import ShortTermRecommendationHead
from app.domain.market.market_daily_analyzer import MarketDailyAnalyzer
from app.domain.models import MarketSeed


def _seed(symbol: str, *, momentum5: float, momentum20: float, trend_strength: float, volume_ratio: float, drawdown20: float = 4.0) -> MarketSeed:
    return MarketSeed(
        symbol=symbol,
        name=symbol,
        trade_date="2026-05-20",
        close_price=10.0,
        momentum5=momentum5,
        momentum20=momentum20,
        volatility20=2.2,
        volume_ratio=volume_ratio,
        drawdown20=drawdown20,
        trend_strength=trend_strength,
        net_mf_amount=1_500_000.0,
        pe_ttm=20.0,
        pb=2.0,
        turnover_rate=1.5,
        news_heat=1,
        positive_news_rate=0.6,
        avg_turnover20=900_000_000,
        industry="电子",
        sector="芯片",
    )


def test_short_term_head_selects_at_most_two_from_shared_pool() -> None:
    seeds = [
        _seed("A", momentum5=5.0, momentum20=12.0, trend_strength=82.0, volume_ratio=1.6),
        _seed("B", momentum5=4.5, momentum20=10.0, trend_strength=78.0, volume_ratio=1.5),
        _seed("C", momentum5=3.8, momentum20=8.5, trend_strength=74.0, volume_ratio=1.4),
    ]
    features = StockFeatureFactory().build(seeds)
    market = MarketDailyAnalyzer().analyze(seeds)

    intraday_inputs = {
        "A": {"minute_confirmation_score": 86, "vwap_supported": True, "opening_range_intact": True},
        "B": {"minute_confirmation_score": 80, "vwap_supported": True, "opening_range_intact": True},
        "C": {"minute_confirmation_score": 58, "vwap_supported": False, "opening_range_intact": False},
    }

    result = ShortTermRecommendationHead().select(features, market, intraday_inputs)

    assert len(result.recommendations) == 2
    assert [item.symbol for item in result.recommendations] == ["A", "B"]
    assert all(item.recommendation_head == "SHORT_TERM_PRIMARY" for item in result.recommendations)


def test_short_term_head_rejects_weak_intraday_structure() -> None:
    seeds = [_seed("A", momentum5=5.0, momentum20=12.0, trend_strength=82.0, volume_ratio=1.6)]
    features = StockFeatureFactory().build(seeds)
    market = MarketDailyAnalyzer().analyze(seeds)

    result = ShortTermRecommendationHead().select(
        features,
        market,
        {"A": {"minute_confirmation_score": 40, "vwap_supported": False, "opening_range_intact": False}},
    )

    assert result.recommendations == []
