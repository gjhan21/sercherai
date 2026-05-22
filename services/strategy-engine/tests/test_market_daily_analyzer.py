from app.domain.market.market_daily_analyzer import MarketDailyAnalyzer
from app.domain.models import MarketSeed


def _seed(
    symbol: str,
    *,
    momentum5: float,
    momentum20: float,
    volatility20: float,
    volume_ratio: float = 1.2,
    drawdown20: float = 5.0,
    trend_strength: float = 70.0,
    net_mf_amount: float = 1_000_000.0,
    news_heat: int = 1,
    positive_news_rate: float = 0.6,
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
        news_heat=news_heat,
        positive_news_rate=positive_news_rate,
    )


def test_market_daily_analyzer_classifies_uptrend_and_attack() -> None:
    analyzer = MarketDailyAnalyzer()
    result = analyzer.analyze(
        [
            _seed("A", momentum5=4.0, momentum20=12.0, volatility20=2.1),
            _seed("B", momentum5=3.5, momentum20=10.0, volatility20=2.3),
            _seed("C", momentum5=3.0, momentum20=9.0, volatility20=2.0),
        ]
    )

    assert result.trend_state == "UPTREND"
    assert result.next_day_rhythm == "ATTACK"
    assert result.risk_posture == "NORMAL"
    assert result.breadth_summary["positive_flow_ratio"] > 0.5


def test_market_daily_analyzer_classifies_range_neutral() -> None:
    analyzer = MarketDailyAnalyzer()
    result = analyzer.analyze(
        [
            _seed("A", momentum5=0.8, momentum20=2.5, volatility20=2.6, net_mf_amount=300_000),
            _seed("B", momentum5=0.3, momentum20=1.8, volatility20=2.7, net_mf_amount=-100_000),
            _seed("C", momentum5=0.1, momentum20=1.2, volatility20=2.4, net_mf_amount=100_000),
        ]
    )

    assert result.trend_state == "RANGE_NEUTRAL"
    assert result.next_day_rhythm == "NEUTRAL"


def test_market_daily_analyzer_classifies_risk_off_and_defense() -> None:
    analyzer = MarketDailyAnalyzer()
    result = analyzer.analyze(
        [
            _seed("A", momentum5=-5.0, momentum20=-8.0, volatility20=5.8, drawdown20=16.0, net_mf_amount=-2_000_000),
            _seed("B", momentum5=-4.0, momentum20=-6.5, volatility20=5.5, drawdown20=14.0, net_mf_amount=-1_500_000),
            _seed("C", momentum5=-3.5, momentum20=-7.0, volatility20=5.2, drawdown20=15.0, net_mf_amount=-1_000_000),
        ]
    )

    assert result.trend_state == "RISK_OFF"
    assert result.next_day_rhythm == "DEFENSE"
    assert result.risk_posture == "TIGHT"
