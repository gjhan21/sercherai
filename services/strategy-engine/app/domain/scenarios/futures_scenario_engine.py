from __future__ import annotations

from app.domain.agents.agent_panel import AgentPanel
from app.schemas.simulation import ScenarioOutcome, ScenarioTemplateConfig, SimulationCard


class FuturesScenarioEngine:
    def __init__(self, agent_panel: AgentPanel) -> None:
        self._agent_panel = agent_panel

    def simulate(self, features: list, agent_options: dict | None = None) -> list[SimulationCard]:
        cards = []
        scenario_templates = _normalize_templates((agent_options or {}).get("scenario_templates", []))
        for item in features:
            scenarios = _build_futures_scenarios(item, scenario_templates)
            opinions, consensus, vetoed, veto_reason = self._agent_panel.review(
                {
                    "direction": item.direction,
                    "trend_strength": item.trend_strength,
                    "risk_level": item.risk_level,
                    "news_bias": item.news_bias,
                    "volume_ratio": item.volume_ratio,
                    "oi_change_pct": item.oi_change_pct,
                    "basis_pct": item.basis_pct,
                    "carry_pct": item.carry_pct,
                    "term_structure_pct": item.term_structure_pct,
                    "curve_slope_pct": item.curve_slope_pct,
                    "inventory_pressure": item.inventory_pressure,
                    "inventory_focus_area": item.inventory_focus_area,
                    "inventory_focus_warehouse": item.inventory_focus_warehouse,
                    "inventory_focus_brand": item.inventory_focus_brand,
                    "inventory_focus_place": item.inventory_focus_place,
                    "inventory_focus_grade": item.inventory_focus_grade,
                    "inventory_area_share": item.inventory_area_share,
                    "inventory_warehouse_share": item.inventory_warehouse_share,
                    "inventory_brand_share": item.inventory_brand_share,
                    "inventory_place_share": item.inventory_place_share,
                    "inventory_grade_share": item.inventory_grade_share,
                    "basis_term_alignment": item.basis_term_alignment,
                    "cross_contract_linkage": item.cross_contract_linkage,
                    "structure_signal_summary": item.structure_signal_summary,
                    "spread_pressure": item.spread_pressure,
                    "shock_adjustment": min(s.score_adjustment for s in scenarios if s.scenario == "shock"),
                },
                agent_options,
            )
            cards.append(
                SimulationCard(
                    asset_key=item.contract,
                    asset_type="futures",
                    scenarios=scenarios,
                    agents=opinions,
                    consensus_action=consensus,
                    vetoed=vetoed,
                    veto_reason=veto_reason,
                )
            )
        return cards


def _build_futures_scenarios(item, templates: dict[str, ScenarioTemplateConfig]) -> list[ScenarioOutcome]:
    direction_sign = -1 if item.direction == "SHORT" else 1
    structure_bonus = max(item.structure_depth_score - 50, 0) / 20
    structure_thesis = "结构联动强化主方向。" if item.structure_factor_summary else "趋势顺畅推进，主方向胜率抬升。"
    bull = round(direction_sign * (5 + abs(item.trend_strength) * 1.8 + structure_bonus), 2)
    base = round(direction_sign * (2 + abs(item.trend_strength) * 0.8 + structure_bonus * 0.4), 2)
    bear = round(-direction_sign * (4 + item.volatility14), 2)
    shock = round(-8 - item.volatility14 * 1.5, 2)
    return [
        _build_outcome("bull", bull, structure_thesis, "顺势推进", "中", templates),
        _build_outcome("base", base, "常态波动下仍可按计划执行。", "按计划执行", "中", templates),
        _build_outcome("bear", bear, "反向波动会压缩盈亏比，需快速收缩。", "收缩仓位", "中高", templates),
        _build_outcome("shock", shock, "高波动冲击时先执行止损与降杠杆。", "停止新增", "高", templates),
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
