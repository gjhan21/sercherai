from __future__ import annotations

from app.domain.agents.agent_panel import AgentPanel
from app.domain.seeds.market_seed_loader import MarketSeedLoader
from app.schemas.simulation import ScenarioOutcome, ScenarioTemplateConfig, SimulationCard


class StockScenarioEngine:
    def __init__(self, agent_panel: AgentPanel) -> None:
        self._agent_panel = agent_panel

    def simulate(self, features: list, agent_options: dict | None = None) -> list[SimulationCard]:
        cards = []
        scenario_templates = _normalize_templates((agent_options or {}).get("scenario_templates", []))
        loader = MarketSeedLoader()
        for item in features:
            # 统一加载历史种子数据
            history_seeds = []
            try:
                history_seeds = loader.load_history(item.symbol, item.trade_date, limit=1250)
            except Exception:
                pass

            # Compute 5-year historical PE percentile
            pe_pct = -1.0
            if item.pe_ttm > 0 and history_seeds:
                pe_vals = [s.pe_ttm for s in history_seeds if s.pe_ttm > 0]
                if pe_vals:
                    below = sum(1 for v in pe_vals if v < item.pe_ttm)
                    pe_pct = below / len(pe_vals)
            item.pe_pct = pe_pct

            # 计算 K线运动学特征 (基于已加载的 history_seeds 提取最近5日)
            price_velocity = 0.0
            price_acceleration = 0.0
            vel_amp_divergence = 0.0
            try:
                if history_seeds and len(history_seeds) >= 5:
                    import numpy as np
                    sorted_seeds = sorted(history_seeds, key=lambda x: x.trade_date)
                    closes = [s.close_price for s in sorted_seeds[-5:]]
                    x = np.arange(5)
                    
                    # 1. 区间幅度
                    amplitude = (closes[-1] - closes[0]) / closes[0] if closes[0] > 0 else 0.0
                    
                    # 2. 最小二乘法拟合斜率 (速度)
                    slope, _ = np.polyfit(x, closes, 1)
                    price_velocity = slope / closes[-1] if closes[-1] > 0 else 0.0
                    
                    # 3. 加速度 (前 4 日拟合速度)
                    prev_slope, _ = np.polyfit(x[:-1], closes[:-1], 1)
                    prev_velocity = prev_slope / closes[-2] if closes[-2] > 0 else 0.0
                    price_acceleration = price_velocity - prev_velocity
                    
                    # 4. 背离度
                    vel_amp_divergence = price_velocity - amplitude
            except Exception:
                pass

            item.price_velocity = price_velocity
            item.price_acceleration = price_acceleration
            item.vel_amp_divergence = vel_amp_divergence

            scenarios = _build_stock_scenarios(item, scenario_templates)
            opinions, consensus, vetoed, veto_reason = self._agent_panel.review(
                {
                    "trend_strength": item.trend_strength,
                    "risk_level": item.risk_level,
                    "news_heat": item.news_heat,
                    "volume_ratio": item.volume_ratio,
                    "pe_ttm": item.pe_ttm,
                    "industry_pe_median": item.industry_pe_median,
                    "pe_pct": item.pe_pct,
                    "price_velocity": item.price_velocity,
                    "price_acceleration": item.price_acceleration,
                    "vel_amp_divergence": item.vel_amp_divergence,
                    "shock_adjustment": min(s.score_adjustment for s in scenarios if s.scenario == "shock"),
                },
                agent_options,
            )
            cards.append(
                SimulationCard(
                    asset_key=item.symbol,
                    asset_type="stock",
                    scenarios=scenarios,
                    agents=opinions,
                    consensus_action=consensus,
                    vetoed=vetoed,
                    veto_reason=veto_reason,
                )
            )
        return cards


def _build_stock_scenarios(item, templates: dict[str, ScenarioTemplateConfig]) -> list[ScenarioOutcome]:
    bull = round(4 + item.momentum20 * 0.3, 2)
    base = round(1 + item.trend_strength * 0.5, 2)
    bear = round(-4 - item.volatility20 * 0.8, 2)
    shock = round(-8 - item.drawdown20 * 0.6, 2)
    return [
        _build_outcome("bull", bull, "景气扩散与资金跟随，趋势继续强化。", "加仓", "低", templates),
        _build_outcome("base", base, "维持当前节奏，等待下一轮验证。", "持有", "中", templates),
        _build_outcome("bear", bear, "市场回撤导致估值与情绪压缩。", "减仓", "中高", templates),
        _build_outcome("shock", shock, "黑天鹅或流动性冲击下先保命。", "回避", "高", templates),
    ]


def _normalize_templates(raw_items: list) -> dict[str, ScenarioTemplateConfig]:
    result: dict[str, ScenarioTemplateConfig] = {}
    for raw in raw_items or []:
        item = raw if isinstance(raw, ScenarioTemplateConfig) else ScenarioTemplateConfig.model_validate(raw)
        result[item.scenario.strip().lower()] = item
    return result


def _build_outcome(
    scenario: str,
    score_adjustment: float,
    thesis: str,
    action: str,
    risk_signal: str,
    templates: dict[str, ScenarioTemplateConfig],
) -> ScenarioOutcome:
    template = templates.get(scenario)
    if template is None:
        return ScenarioOutcome(
            scenario=scenario,
            thesis=thesis,
            score_adjustment=score_adjustment,
            action=action,
            risk_signal=risk_signal,
        )
    return ScenarioOutcome(
        scenario=scenario,
        thesis=template.thesis_template or thesis,
        score_adjustment=round(score_adjustment + template.score_bias, 2),
        action=template.action or action,
        risk_signal=template.risk_signal or risk_signal,
    )
