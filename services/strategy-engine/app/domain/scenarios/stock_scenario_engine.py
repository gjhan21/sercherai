from __future__ import annotations

from app.domain.agents.agent_panel import AgentPanel
from app.schemas.simulation import ScenarioOutcome, ScenarioTemplateConfig, SimulationCard


class StockScenarioEngine:
    def __init__(self, agent_panel: AgentPanel) -> None:
        self._agent_panel = agent_panel

    def simulate(self, features: list, agent_options: dict | None = None) -> list[SimulationCard]:
        cards = []
        scenario_templates = _normalize_templates((agent_options or {}).get("scenario_templates", []))
        for item in features:
            scenarios = _build_stock_scenarios(item, scenario_templates)
            opinions, consensus, vetoed, veto_reason = self._agent_panel.review(
                {
                    "trend_strength": item.trend_strength,
                    "risk_level": item.risk_level,
                    "news_heat": item.news_heat,
                    "volume_ratio": item.volume_ratio,
                    "pe_ttm": item.pe_ttm,
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
