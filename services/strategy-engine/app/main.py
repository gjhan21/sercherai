from __future__ import annotations

import logging
from dataclasses import dataclass
from functools import lru_cache

from fastapi import FastAPI

from app.domain.agents.agent_panel import AgentPanel
from app.core.publish_store import InMemoryPublishStore
from app.domain.decision.stock_decision_fusion import StockDecisionFusion
from app.domain.features.futures_feature_factory import FuturesFeatureFactory
from app.domain.features.stock_feature_factory import StockFeatureFactory
from app.domain.graph.market_graph_builder import MarketGraphBuilder
from app.domain.graph.strategy_graph_client import StrategyGraphClient
from app.domain.pipelines.futures_strategy_pipeline import FuturesStrategyPipeline
from app.domain.pipelines.stock_selection_pipeline import StockSelectionPipeline
from app.domain.publish.go_backend_publisher import GoBackendPublisher
from app.domain.reports.futures_report_builder import FuturesReportBuilder
from app.domain.reports.report_renderer import ReportRenderer
from app.domain.reports.stock_report_builder import StockReportBuilder
from app.domain.risk.leverage_guard import LeverageGuard
from app.domain.risk.portfolio_guard import PortfolioGuard
from app.domain.scenarios.futures_scenario_engine import FuturesScenarioEngine
from app.domain.scenarios.stock_scenario_engine import StockScenarioEngine
from app.domain.seeds.futures_seed_loader import FuturesSeedLoader
from app.domain.seeds.market_seed_loader import MarketSeedLoader
from app.domain.seeds.stock_seed_miner import StockSeedMiner
from app.domain.selectors.futures_selector import FuturesSelector
from app.domain.selectors.stock_selector import StockSelector
from app.domain.universe.stock_universe_builder import StockUniverseBuilder
from app.core.job_runner import JobRunner
from app.core.job_store import InMemoryJobStore
from app.settings import Settings, get_settings

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s [%(name)s] %(message)s",
)


@dataclass(slots=True)
class Container:
    settings: Settings
    job_store: InMemoryJobStore
    publish_store: InMemoryPublishStore
    job_runner: JobRunner
    go_backend_publisher: GoBackendPublisher
    stock_selection_pipeline: StockSelectionPipeline
    futures_strategy_pipeline: FuturesStrategyPipeline


@lru_cache(maxsize=1)
def get_container() -> Container:
    settings = get_settings()
    job_store = InMemoryJobStore()
    publish_store = InMemoryPublishStore()
    market_graph_builder = MarketGraphBuilder()
    strategy_graph_client = StrategyGraphClient(settings=settings)
    agent_panel = AgentPanel()
    report_renderer = ReportRenderer()
    stock_selection_pipeline = StockSelectionPipeline(
        market_seed_loader=MarketSeedLoader(settings=settings),
        stock_feature_factory=StockFeatureFactory(),
        stock_selector=StockSelector(),
        portfolio_guard=PortfolioGuard(),
        stock_report_builder=StockReportBuilder(),
        stock_universe_builder=StockUniverseBuilder(),
        stock_seed_miner=StockSeedMiner(),
        stock_decision_fusion=StockDecisionFusion(),
        stock_scenario_engine=StockScenarioEngine(agent_panel=agent_panel),
        market_graph_builder=market_graph_builder,
        strategy_graph_client=strategy_graph_client,
    )
    futures_strategy_pipeline = FuturesStrategyPipeline(
        futures_seed_loader=FuturesSeedLoader(settings=settings),
        futures_feature_factory=FuturesFeatureFactory(),
        futures_selector=FuturesSelector(),
        leverage_guard=LeverageGuard(),
        futures_report_builder=FuturesReportBuilder(),
        futures_scenario_engine=FuturesScenarioEngine(agent_panel=agent_panel),
        market_graph_builder=market_graph_builder,
        strategy_graph_client=strategy_graph_client,
    )
    job_runner = JobRunner(
        job_store,
        simulate_delay_seconds=settings.simulate_job_delay_seconds,
        stock_selection_pipeline=stock_selection_pipeline,
        futures_strategy_pipeline=futures_strategy_pipeline,
    )
    go_backend_publisher = GoBackendPublisher(
        store=publish_store,
        renderer=report_renderer,
    )
    return Container(
        settings=settings,
        job_store=job_store,
        publish_store=publish_store,
        job_runner=job_runner,
        go_backend_publisher=go_backend_publisher,
        stock_selection_pipeline=stock_selection_pipeline,
        futures_strategy_pipeline=futures_strategy_pipeline,
    )


app = FastAPI(
    title="strategy-engine",
    version="0.1.0",
    description="Standalone strategy engine service for job orchestration",
)

from app.api.routes_health import router as health_router  # noqa: E402
from app.api.routes_jobs import router as jobs_router  # noqa: E402
from app.api.routes_publish import router as publish_router  # noqa: E402

app.include_router(health_router)
app.include_router(jobs_router)
app.include_router(publish_router)
