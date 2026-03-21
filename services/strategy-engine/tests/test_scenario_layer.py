from app.domain.agents.agent_panel import AgentPanel
from app.domain.features.futures_feature_factory import FuturesFeatureFactory
from app.domain.models import FuturesSeed
from app.domain.scenarios.futures_scenario_engine import FuturesScenarioEngine


def test_risk_agent_can_veto_high_risk_futures_contract() -> None:
    feature = FuturesFeatureFactory().build(
        [
            FuturesSeed(
                contract="IM2606",
                name="中证1000股指",
                trade_date="2026-03-17",
                last_price=5872.0,
                basis_pct=1.4,
                volatility14=4.6,
                trend_strength=-1.8,
                oi_change_pct=8.0,
                volume_ratio=1.6,
                flow_bias=-0.5,
                carry_pct=-0.3,
                news_bias=-0.2,
                regime="VOLATILE",
            )
        ]
    )[0]

    simulations = FuturesScenarioEngine(agent_panel=AgentPanel()).simulate([feature])

    assert len(simulations) == 1
    assert simulations[0].vetoed is True
    assert simulations[0].consensus_action == "REJECT"


def test_futures_scenario_template_can_override_action_and_score_bias() -> None:
    feature = FuturesFeatureFactory().build(
        [
            FuturesSeed(
                contract="IF2606",
                name="沪深300股指",
                trade_date="2026-03-17",
                last_price=3612.0,
                basis_pct=0.2,
                volatility14=1.8,
                trend_strength=1.2,
                oi_change_pct=3.5,
                volume_ratio=1.3,
                flow_bias=0.6,
                carry_pct=0.1,
                news_bias=0.2,
                regime="TREND",
            )
        ]
    )[0]

    simulation = FuturesScenarioEngine(agent_panel=AgentPanel()).simulate(
        [feature],
        {
            "scenario_templates": [
                {
                    "scenario": "bull",
                    "label": "加速",
                    "thesis_template": "宏观共振下延续主升浪。",
                    "action": "扩大持仓",
                    "risk_signal": "低",
                    "score_bias": 2.5,
                }
            ]
        },
    )[0]

    bull = next(item for item in simulation.scenarios if item.scenario == "bull")
    assert bull.thesis == "宏观共振下延续主升浪。"
    assert bull.action == "扩大持仓"
    assert bull.score_adjustment > 9
