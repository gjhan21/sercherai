from __future__ import annotations

from datetime import datetime
from typing import Optional

from app.domain.agents.agent_panel import AgentPanel
from app.domain.graph.market_graph_builder import MarketGraphBuilder
from app.domain.scenarios.futures_scenario_engine import FuturesScenarioEngine
from app.domain.features.futures_feature_factory import FuturesFeatureFactory
from app.domain.reports.futures_report_builder import FuturesReportBuilder
from app.domain.risk.leverage_guard import LeverageGuard
from app.domain.seeds.futures_seed_loader import FuturesSeedLoader
from app.domain.selectors.futures_selector import FuturesSelector
from app.schemas.futures import FuturesStrategyPayload, FuturesStrategyReport


class FuturesStrategyPipeline:
    def __init__(
        self,
        futures_seed_loader: FuturesSeedLoader,
        futures_feature_factory: FuturesFeatureFactory,
        futures_selector: FuturesSelector,
        leverage_guard: LeverageGuard,
        futures_report_builder: FuturesReportBuilder,
        futures_scenario_engine: Optional[FuturesScenarioEngine] = None,
        market_graph_builder: Optional[MarketGraphBuilder] = None,
    ) -> None:
        self._futures_seed_loader = futures_seed_loader
        self._futures_feature_factory = futures_feature_factory
        self._futures_selector = futures_selector
        self._leverage_guard = leverage_guard
        self._futures_report_builder = futures_report_builder
        self._futures_scenario_engine = futures_scenario_engine or FuturesScenarioEngine(agent_panel=AgentPanel())
        self._market_graph_builder = market_graph_builder or MarketGraphBuilder()

    def run(self, raw_payload: dict) -> tuple[FuturesStrategyReport, list[str]]:
        payload = FuturesStrategyPayload.model_validate(raw_payload)
        if not payload.trade_date:
            payload.trade_date = datetime.now().strftime("%Y-%m-%d")
        agent_options = {
            "enabled_agents": payload.enabled_agents,
            "positive_threshold": payload.positive_threshold,
            "negative_threshold": payload.negative_threshold,
            "allow_veto": payload.allow_veto,
            "scenario_templates": [item.model_dump(mode="json") for item in payload.scenario_templates],
        }

        seed_result = self._futures_seed_loader.load(payload)
        if not seed_result.seeds:
            raise ValueError("futures strategy seed set is empty")

        features = self._futures_feature_factory.build(seed_result.seeds)
        ranked = self._futures_selector.select(features, payload.limit)
        guarded, warnings = self._leverage_guard.apply(ranked, payload)
        warnings = [*seed_result.warnings, *warnings]
        report = self._futures_report_builder.build(payload, guarded, warnings)
        report.context_meta = seed_result.meta
        report.simulations = self._futures_scenario_engine.simulate(guarded, agent_options)
        report.graph_summary = self._market_graph_builder.build_futures(guarded)
        report.consensus_summary = _build_consensus_summary(report.simulations)
        return report, warnings


def _build_consensus_summary(simulations: list) -> str:
    veto_count = sum(1 for item in simulations if item.vetoed)
    go_count = sum(1 for item in simulations if item.consensus_action == "GO")
    watch_count = sum(1 for item in simulations if item.consensus_action == "WATCH")
    return f"多代理收敛：GO {go_count}，WATCH {watch_count}，VETO {veto_count}。"
