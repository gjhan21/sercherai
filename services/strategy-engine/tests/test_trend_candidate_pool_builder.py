from app.domain.candidates.trend_candidate_pool_builder import TrendCandidatePoolBuilder
from app.domain.features.stock_feature_factory import StockFeatureFactory
from app.domain.market.market_daily_analyzer import MarketDailyAnalyzer
from app.domain.models import MarketSeed


def _seed(
    symbol: str,
    *,
    momentum5: float,
    momentum20: float,
    volatility20: float,
    volume_ratio: float,
    drawdown20: float,
    trend_strength: float,
    net_mf_amount: float,
    avg_turnover20: float = 800_000_000,
    industry: str = "电子",
    sector: str = "芯片",
) -> MarketSeed:
    return MarketSeed(
        symbol=symbol,
        name=symbol,
        trade_date="2026-05-20",
        close_price=10.0,
        momentum5=momentum5,
        momentum20=momentum20,
        volatility20=volatility20,
        volume_ratio=volume_ratio,
        drawdown20=drawdown20,
        trend_strength=trend_strength,
        net_mf_amount=net_mf_amount,
        pe_ttm=20.0,
        pb=2.0,
        turnover_rate=1.5,
        news_heat=1,
        positive_news_rate=0.6,
        avg_turnover20=avg_turnover20,
        industry=industry,
        sector=sector,
    )


def test_trend_candidate_pool_builder_prefers_trend_and_excludes_weak_structure() -> None:
    seeds = [
        _seed("STRONG1", momentum5=5.0, momentum20=12.0, volatility20=2.1, volume_ratio=1.6, drawdown20=4.0, trend_strength=80.0, net_mf_amount=2_000_000),
        _seed("STRONG2", momentum5=4.2, momentum20=10.5, volatility20=2.4, volume_ratio=1.4, drawdown20=5.0, trend_strength=76.0, net_mf_amount=1_500_000),
        _seed("WEAK1", momentum5=-2.0, momentum20=-4.0, volatility20=5.1, volume_ratio=0.7, drawdown20=18.0, trend_strength=20.0, net_mf_amount=-1_000_000),
    ]
    features = StockFeatureFactory().build(seeds)
    market = MarketDailyAnalyzer().analyze(seeds)

    result = TrendCandidatePoolBuilder().build(features, market)

    assert result.coarse_pool
    assert result.watch_pool
    assert result.watch_pool[0].symbol == "STRONG1"
    assert "WEAK1" not in {item.symbol for item in result.watch_pool}
    assert result.summary["watch_pool_count"] <= result.summary["focus_pool_count"] <= result.summary["coarse_pool_count"]


def test_trend_candidate_pool_builder_tightens_pool_in_risk_off_market() -> None:
    seeds = [
        _seed("A", momentum5=-4.0, momentum20=-6.0, volatility20=5.4, volume_ratio=0.8, drawdown20=14.0, trend_strength=25.0, net_mf_amount=-1_000_000),
        _seed("B", momentum5=-3.5, momentum20=-5.2, volatility20=5.1, volume_ratio=0.9, drawdown20=12.0, trend_strength=30.0, net_mf_amount=-800_000),
        _seed("C", momentum5=1.5, momentum20=4.0, volatility20=2.0, volume_ratio=1.3, drawdown20=4.0, trend_strength=72.0, net_mf_amount=1_200_000),
    ]
    features = StockFeatureFactory().build(seeds)
    market = MarketDailyAnalyzer().analyze(seeds)

    result = TrendCandidatePoolBuilder().build(features, market)

    assert market.trend_state == "RISK_OFF"
    assert result.summary["risk_posture"] == "TIGHT"
    assert result.summary["watch_pool_limit"] <= 10
