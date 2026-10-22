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


def test_structure_linkage_strengthens_basis_agent_and_bull_scenario() -> None:
    feature = FuturesFeatureFactory().build(
        [
            FuturesSeed(
                contract="IF2606",
                name="沪深300股指",
                trade_date="2026-03-17",
                last_price=4012.0,
                basis_pct=-0.25,
                volatility14=1.4,
                trend_strength=1.0,
                oi_change_pct=4.5,
                volume_ratio=1.18,
                flow_bias=0.42,
                carry_pct=0.36,
                news_bias=0.18,
                regime="TREND",
                turnover_ratio=1.2,
                term_structure_pct=0.72,
                curve_slope_pct=0.88,
                basis_term_alignment=0.9,
                cross_contract_linkage=-0.28,
                structure_signal_summary="期限结构顺价差联动，当前合约位于受益腿",
                spread_pressure=-0.26,
                spread_percentile=0.86,
                spread_pair="IF2606/IF2609",
            )
        ]
    )[0]

    simulation = FuturesScenarioEngine(agent_panel=AgentPanel()).simulate([feature])[0]

    bull = next(item for item in simulation.scenarios if item.scenario == "bull")
    basis_opinion = next(item for item in simulation.agents if item.agent == "basis")
    assert bull.score_adjustment > 7
    assert "结构联动" in bull.thesis
    assert basis_opinion.stance == "POSITIVE"
    assert "结构联动" in basis_opinion.summary


def test_stock_basis_agent_adaptive_valuation() -> None:
    from app.domain.agents.basis_agent import BasisAgent

    agent = BasisAgent()

    # Case 1: PE < Industry Median and PE percentile < 40% -> POSITIVE
    opinion1 = agent.evaluate({
        "pe_ttm": 15.0,
        "industry_pe_median": 20.0,
        "pe_pct": 0.35,
    })
    assert opinion1.stance == "POSITIVE"
    assert "低估" in opinion1.summary

    # Case 2: PE > Industry Median -> NEGATIVE
    opinion2 = agent.evaluate({
        "pe_ttm": 25.0,
        "industry_pe_median": 20.0,
        "pe_pct": 0.35,
    })
    assert opinion2.stance == "NEGATIVE"
    assert "偏高" in opinion2.summary

    # Case 3: PE percentile > 80% -> NEGATIVE
    opinion3 = agent.evaluate({
        "pe_ttm": 18.0,
        "industry_pe_median": 20.0,
        "pe_pct": 0.85,
    })
    assert opinion3.stance == "NEGATIVE"
    assert "偏高" in opinion3.summary

    # Case 4: PE percentile between 40% and 80%, PE < Industry Median -> CAUTION
    opinion4 = agent.evaluate({
        "pe_ttm": 18.0,
        "industry_pe_median": 20.0,
        "pe_pct": 0.50,
    })
    assert opinion4.stance == "CAUTION"
    assert "合理" in opinion4.summary

    # Case 5: PE is 0.0 (data missing or special case) and no futures parameters -> CAUTION
    opinion5 = agent.evaluate({
        "pe_ttm": 0.0,
    })
    assert opinion5.stance == "CAUTION"
    assert "暂无有效估值" in opinion5.summary

    # Case 6: PE is negative (loss-making stock) and no futures parameters -> CAUTION
    opinion6 = agent.evaluate({
        "pe_ttm": -15.4,
    })
    assert opinion6.stance == "CAUTION"
    assert "暂无有效估值" in opinion6.summary


def test_futures_regime_bias_does_not_override_direction() -> None:
    # Under regime TREND, directional_edge originally would be LONG regardless of edge score.
    # Now, it uses directional_edge_adjusted = directional_edge + 0.08
    
    # Case 1: directional_edge is highly negative (e.g. -0.45) and regime is TREND.
    # Adjusted edge = -0.45 + 0.08 = -0.37 <= -0.18 -> SHORT (not overridden to LONG)
    feature1 = FuturesFeatureFactory().build(
        [
            FuturesSeed(
                contract="FG2609",
                name="玻璃",
                trade_date="2026-03-17",
                last_price=1500.0,
                basis_pct=-0.1,
                volatility14=1.5,
                trend_strength=-1.5, # negative trend
                oi_change_pct=2.0,
                volume_ratio=1.1,
                flow_bias=-1.0,      # negative flow
                carry_pct=-0.5,
                news_bias=-0.5,
                regime="TREND",
            )
        ]
    )[0]
    assert feature1.direction == "SHORT"

    # Case 2: directional_edge is neutral but positive (e.g. 0.11), regime is TREND.
    # Adjusted edge = 0.11 + 0.08 = 0.19 >= 0.18 -> LONG (boosted to LONG)
    feature2 = FuturesFeatureFactory().build(
        [
            FuturesSeed(
                contract="FG2609",
                name="玻璃",
                trade_date="2026-03-17",
                last_price=1500.0,
                basis_pct=-0.01,
                volatility14=1.0,
                trend_strength=0.1,
                oi_change_pct=1.0,
                volume_ratio=1.0,
                flow_bias=0.15,
                carry_pct=0.05,
                news_bias=0.0,
                regime="TREND",
            )
        ]
    )[0]
    assert feature2.direction == "LONG"


def test_trend_agent_kinematics_scenarios() -> None:
    from app.domain.agents.trend_agent import TrendAgent
    agent = TrendAgent()

    # Case 1: 强势共振突破
    opinion1 = agent.evaluate({
        "trend_strength": 1.2,
        "price_velocity": 0.025,
        "price_acceleration": 0.002,
        "vel_amp_divergence": 0.01,
    })
    assert opinion1.stance == "POSITIVE"
    assert opinion1.confidence == 85
    assert "加速爆发" in opinion1.summary

    # Case 2: 顶背离衰竭防护
    opinion2 = agent.evaluate({
        "trend_strength": 1.3,
        "price_velocity": 0.002,
        "price_acceleration": -0.001,
        "vel_amp_divergence": -0.08,
    })
    assert opinion2.stance == "NEGATIVE"
    assert opinion2.confidence == 80
    assert "动能枯竭" in opinion2.summary

    # Case 3: 失速下跌加速
    opinion3 = agent.evaluate({
        "trend_strength": -0.8,
        "price_velocity": -0.02,
        "price_acceleration": -0.005,
        "vel_amp_divergence": -0.01,
    })
    assert opinion3.stance == "NEGATIVE"
    assert opinion3.confidence == 78
    assert "负向加速" in opinion3.summary

