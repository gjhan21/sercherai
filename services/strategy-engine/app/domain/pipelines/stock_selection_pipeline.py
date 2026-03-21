from __future__ import annotations

from datetime import datetime
from time import perf_counter
from typing import Optional

from app.domain.agents.agent_panel import AgentPanel
from app.domain.decision.stock_decision_fusion import StockDecisionFusion
from app.domain.features.stock_feature_factory import StockFeatureFactory
from app.domain.graph.market_graph_builder import MarketGraphBuilder
from app.domain.models import MarketSeed, StockFeature
from app.domain.reports.stock_report_builder import StockReportBuilder
from app.domain.risk.portfolio_guard import PortfolioGuard
from app.domain.scenarios.stock_scenario_engine import StockScenarioEngine
from app.domain.seeds.market_seed_loader import MarketSeedLoader
from app.domain.seeds.stock_seed_miner import StockSeedMiner
from app.domain.selectors.stock_selector import StockSelector
from app.domain.universe.stock_universe_builder import StockUniverseBuilder
from app.schemas.stock import (
    StockCandidateSnapshot,
    StockEvidenceRecord,
    StockSelectionPayload,
    StockSelectionReport,
    StockStageLog,
)


class StockSelectionPipeline:
    def __init__(
        self,
        market_seed_loader: MarketSeedLoader,
        stock_feature_factory: StockFeatureFactory,
        stock_selector: StockSelector,
        portfolio_guard: PortfolioGuard,
        stock_report_builder: StockReportBuilder,
        stock_universe_builder: Optional[StockUniverseBuilder] = None,
        stock_seed_miner: Optional[StockSeedMiner] = None,
        stock_decision_fusion: Optional[StockDecisionFusion] = None,
        stock_scenario_engine: Optional[StockScenarioEngine] = None,
        market_graph_builder: Optional[MarketGraphBuilder] = None,
    ) -> None:
        self._market_seed_loader = market_seed_loader
        self._stock_feature_factory = stock_feature_factory
        self._stock_selector = stock_selector
        self._portfolio_guard = portfolio_guard
        self._stock_report_builder = stock_report_builder
        self._stock_universe_builder = stock_universe_builder or StockUniverseBuilder()
        self._stock_seed_miner = stock_seed_miner or StockSeedMiner()
        self._stock_decision_fusion = stock_decision_fusion or StockDecisionFusion()
        self._stock_scenario_engine = stock_scenario_engine or StockScenarioEngine(agent_panel=AgentPanel())
        self._market_graph_builder = market_graph_builder or MarketGraphBuilder()

    def run(self, raw_payload: dict) -> tuple[StockSelectionReport, list[str]]:
        payload = StockSelectionPayload.model_validate(raw_payload)
        if not payload.trade_date:
            payload.trade_date = datetime.now().strftime("%Y-%m-%d")
        agent_options = {
            "enabled_agents": payload.enabled_agents,
            "positive_threshold": payload.positive_threshold,
            "negative_threshold": payload.negative_threshold,
            "allow_veto": payload.allow_veto,
            "scenario_templates": [item.model_dump(mode="json") for item in payload.scenario_templates],
        }

        seed_result = self._market_seed_loader.load(payload)
        if not seed_result.seeds:
            raise ValueError("stock selection seed set is empty")

        stage_counts: dict[str, int] = {}
        stage_durations_ms: dict[str, int] = {}
        stage_logs: list[StockStageLog] = []

        regime_start = perf_counter()
        market_regime = _detect_market_regime(seed_result.seeds, payload)
        regime_duration = _duration_ms(regime_start)
        stage_counts["MARKET_REGIME"] = len(seed_result.seeds)
        stage_durations_ms["MARKET_REGIME"] = regime_duration
        stage_logs.append(
            StockStageLog(
                stage_key="MARKET_REGIME",
                stage_order=1,
                input_count=len(seed_result.seeds),
                output_count=len(seed_result.seeds),
                duration_ms=regime_duration,
                detail_message=f"detected market_regime={market_regime}",
                payload_snapshot={
                    "template_key": payload.template_key,
                    "template_name": payload.template_name,
                },
            )
        )

        universe_start = perf_counter()
        universe_result = self._stock_universe_builder.build(payload, seed_result)
        features = self._stock_feature_factory.build(universe_result.seeds)
        universe_duration = _duration_ms(universe_start)
        if not features:
            raise ValueError("stock selection universe is empty after normalization")
        stage_counts["UNIVERSE"] = len(features)
        stage_durations_ms["UNIVERSE"] = universe_duration
        stage_logs.append(
            StockStageLog(
                stage_key="UNIVERSE",
                stage_order=2,
                input_count=len(seed_result.seeds),
                output_count=len(features),
                duration_ms=universe_duration,
                detail_message=f"selection_mode={payload.selection_mode} universe_scope={payload.effective_universe_scope()}",
                payload_snapshot={
                    "profile_id": payload.profile_id,
                    "meta_source": universe_result.meta.get("source", ""),
                },
            )
        )

        theme_event_start = perf_counter()
        _enrich_theme_event_features(features, market_regime)
        theme_event_duration = _duration_ms(theme_event_start)
        stage_counts["THEME_EVENT"] = len(features)
        stage_durations_ms["THEME_EVENT"] = theme_event_duration
        stage_logs.append(
            StockStageLog(
                stage_key="THEME_EVENT",
                stage_order=3,
                input_count=len(features),
                output_count=len(features),
                duration_ms=theme_event_duration,
                detail_message="theme/event enrichment completed",
                payload_snapshot={"top_themes": _top_theme_tags(features)},
            )
        )

        seed_pool_start = perf_counter()
        seed_mining_result = self._stock_seed_miner.mine(features, payload, market_regime)
        seed_pool_duration = _duration_ms(seed_pool_start)
        seed_pool = seed_mining_result.seed_pool
        if not seed_pool:
            raise ValueError("stock selection seed pool is empty after mining")
        stage_counts["SEED_POOL"] = len(seed_pool)
        stage_durations_ms["SEED_POOL"] = seed_pool_duration
        stage_logs.append(
            StockStageLog(
                stage_key="SEED_POOL",
                stage_order=4,
                input_count=len(features),
                output_count=len(seed_pool),
                duration_ms=seed_pool_duration,
                detail_message="5 fixed buckets with market regime + template routing",
                payload_snapshot={
                    "bucket_members": seed_mining_result.bucket_members,
                    "bucket_limits": seed_mining_result.bucket_limits,
                },
            )
        )

        candidate_start = perf_counter()
        fused_seed_pool = self._stock_decision_fusion.fuse(seed_pool, payload)
        candidate_pool = self._stock_selector.select(fused_seed_pool, payload.candidate_pool_limit)
        candidate_duration = _duration_ms(candidate_start)
        stage_counts["CANDIDATE_POOL"] = len(candidate_pool)
        stage_durations_ms["CANDIDATE_POOL"] = candidate_duration
        stage_logs.append(
            StockStageLog(
                stage_key="CANDIDATE_POOL",
                stage_order=5,
                input_count=len(seed_pool),
                output_count=len(candidate_pool),
                duration_ms=candidate_duration,
                detail_message=(
                    f"fusion weights quant={payload.quant_weight:.2f},event={payload.event_weight:.2f},"
                    f"resonance={payload.resonance_weight:.2f},risk={payload.liquidity_risk_weight:.2f}"
                ),
                payload_snapshot={"top_symbols": [item.symbol for item in candidate_pool[:10]]},
            )
        )

        portfolio_start = perf_counter()
        guard_result = self._portfolio_guard.apply(candidate_pool, payload)
        guarded = guard_result.portfolio
        watchlist = guard_result.watchlist
        portfolio_duration = _duration_ms(portfolio_start)
        stage_counts["PORTFOLIO"] = len(guarded)
        stage_counts["WATCHLIST"] = len(watchlist)
        stage_durations_ms["PORTFOLIO"] = portfolio_duration
        stage_logs.append(
            StockStageLog(
                stage_key="PORTFOLIO",
                stage_order=6,
                input_count=len(candidate_pool),
                output_count=len(guarded),
                duration_ms=portfolio_duration,
                detail_message=f"portfolio limit={payload.limit} max_risk={payload.max_risk_level}",
                payload_snapshot={
                    "selected_symbols": [item.symbol for item in guarded],
                    "watchlist_symbols": [item.symbol for item in watchlist],
                },
            )
        )

        warnings = [*universe_result.warnings, *guard_result.warnings]
        candidate_snapshots = _build_candidate_snapshots(features, seed_pool, candidate_pool, guarded, watchlist)
        evidence_records = _build_evidence_records(candidate_pool, guarded, watchlist)
        evaluation_summary = _build_evaluation_summary(guarded, watchlist)
        report = self._stock_report_builder.build(
            payload,
            guarded,
            warnings,
            market_regime=market_regime,
            template_snapshot=_build_template_snapshot(payload, market_regime),
            evaluation_summary=evaluation_summary,
            stage_counts=stage_counts,
            stage_durations_ms=stage_durations_ms,
            stage_logs=stage_logs,
            evidence_records=evidence_records,
            evaluation_records=[],
            candidate_snapshots=candidate_snapshots,
            watchlist=watchlist,
        )
        report.context_meta = universe_result.meta
        report.context_meta["market_regime"] = market_regime
        if payload.template_key:
            report.context_meta["template_key"] = payload.template_key
        if payload.template_name:
            report.context_meta["template_name"] = payload.template_name
        report.simulations = self._stock_scenario_engine.simulate(guarded, agent_options)
        report.graph_summary = self._market_graph_builder.build_stock(guarded)
        report.consensus_summary = _build_consensus_summary(report.simulations)
        return report, warnings


def _build_consensus_summary(simulations: list) -> str:
    veto_count = sum(1 for item in simulations if item.vetoed)
    go_count = sum(1 for item in simulations if item.consensus_action == "GO")
    hold_count = sum(1 for item in simulations if item.consensus_action == "HOLD")
    return f"多代理收敛：GO {go_count}，HOLD {hold_count}，VETO {veto_count}。"


def _duration_ms(started_at: float) -> int:
    return int((perf_counter() - started_at) * 1000)


def _build_candidate_snapshots(
    universe: list[StockFeature],
    seed_pool: list[StockFeature],
    candidate_pool: list[StockFeature],
    portfolio: list[StockFeature],
    watchlist: list[StockFeature],
) -> list[StockCandidateSnapshot]:
    result: list[StockCandidateSnapshot] = []
    seed_symbols = {item.symbol for item in seed_pool}
    candidate_symbols = {item.symbol for item in candidate_pool}
    portfolio_symbols = {item.symbol for item in portfolio}
    watchlist_symbols = {item.symbol for item in watchlist}
    result.extend(_build_stage_snapshots("UNIVERSE", _rank_by_quant(universe), seed_symbols, watchlist_symbols))
    result.extend(_build_stage_snapshots("SEED_POOL", seed_pool, candidate_symbols, watchlist_symbols))
    result.extend(_build_stage_snapshots("CANDIDATE_POOL", candidate_pool, portfolio_symbols, watchlist_symbols))
    result.extend(_build_stage_snapshots("PORTFOLIO", portfolio, portfolio_symbols, watchlist_symbols))
    return result


def _build_stage_snapshots(
    stage: str,
    ordered: list[StockFeature],
    selected_symbols: set[str],
    watchlist_symbols: set[str],
) -> list[StockCandidateSnapshot]:
    snapshots: list[StockCandidateSnapshot] = []
    for index, item in enumerate(ordered, start=1):
        portfolio_role = item.portfolio_role
        if not portfolio_role and item.symbol in watchlist_symbols:
            portfolio_role = "WATCHLIST"
        snapshots.append(
            StockCandidateSnapshot(
                symbol=item.symbol,
                name=item.name,
                stage=stage,
                quant_score=item.quant_score,
                total_score=item.score,
                risk_level=item.risk_level,
                selected=item.symbol in selected_symbols,
                rank=index,
                reason_summary=item.reason_summary,
                evidence_summary=item.evidence_summary,
                portfolio_role=portfolio_role,
                risk_summary=_feature_risk_summary(item),
                factor_breakdown_json=item.factor_breakdown(),
            )
        )
    return snapshots


def _rank_by_quant(features: list[StockFeature]) -> list[StockFeature]:
    return sorted(features, key=lambda item: (-item.quant_score, -item.momentum20, item.symbol))


def _detect_market_regime(seeds: list[MarketSeed], payload: StockSelectionPayload) -> str:
    bias = str(payload.template_snapshot.get("market_regime_bias", "") or "").strip().upper()
    if bias in {"UPTREND", "ROTATION", "EVENT_DRIVEN", "DEFENSIVE", "RISK_OFF"}:
        return bias
    if not seeds:
        return "ROTATION"
    avg_momentum20 = sum(item.momentum20 for item in seeds) / len(seeds)
    avg_volatility20 = sum(item.volatility20 for item in seeds) / len(seeds)
    positive_flow_ratio = sum(1 for item in seeds if item.net_mf_amount > 0) / len(seeds)
    event_ratio = sum(1 for item in seeds if item.news_heat >= 3) / len(seeds)
    if avg_volatility20 >= 4.2 and avg_momentum20 < -1:
        return "RISK_OFF"
    if event_ratio >= 0.35 and positive_flow_ratio >= 0.4:
        return "EVENT_DRIVEN"
    if avg_momentum20 >= 5 and positive_flow_ratio >= 0.6:
        return "UPTREND"
    if avg_momentum20 <= 1.5 or positive_flow_ratio < 0.45:
        return "DEFENSIVE"
    return "ROTATION"


def _enrich_theme_event_features(features: list[StockFeature], market_regime: str) -> None:
    for item in features:
        theme_bonus = min(10.0, float(len(item.theme_tags)) * 2.0)
        event_bonus = 0.0
        resonance_bonus = 0.0
        if market_regime == "EVENT_DRIVEN":
            event_bonus += 6.0 if item.news_heat >= 3 else 1.5
            resonance_bonus += 3.0 if item.theme_tags else 0.0
        elif market_regime == "UPTREND":
            resonance_bonus += 3.5 if item.momentum20 >= 5 else 0.0
            event_bonus += 1.5 if item.net_mf_amount > 0 else 0.0
        elif market_regime == "DEFENSIVE":
            event_bonus += 1.5 if item.risk_level == "LOW" else -1.0
        elif market_regime == "RISK_OFF":
            event_bonus -= 2.5 if item.risk_level == "HIGH" else 0.0
            item.risk_adjustment_score = max(0.0, item.risk_adjustment_score - 8.0)
        item.event_score = round(min(100.0, max(0.0, item.event_score + theme_bonus + event_bonus)), 2)
        item.resonance_score = round(min(100.0, max(0.0, item.resonance_score + theme_bonus + resonance_bonus)), 2)
        if item.theme_tags:
            item.positive_reasons = list(dict.fromkeys([*item.positive_reasons, f"题材共振：{' / '.join(item.theme_tags[:2])}"]))
        item.evidence_summary = "；".join((item.positive_reasons or item.reasons)[:2])
        item.reason_summary = "；".join((item.positive_reasons or item.reasons)[:3])


def _top_theme_tags(features: list[StockFeature]) -> list[str]:
    counter: dict[str, int] = {}
    for item in features:
        for tag in item.theme_tags[:3]:
            counter[tag] = counter.get(tag, 0) + 1
    return [tag for tag, _count in sorted(counter.items(), key=lambda item: (-item[1], item[0]))[:5]]


def _build_evidence_records(
    candidate_pool: list[StockFeature],
    portfolio: list[StockFeature],
    watchlist: list[StockFeature],
) -> list[StockEvidenceRecord]:
    items: list[StockEvidenceRecord] = []
    seen: set[tuple[str, str]] = set()
    for stage, features in (
        ("CANDIDATE_POOL", candidate_pool),
        ("PORTFOLIO", portfolio),
    ):
        for item in features:
            key = (stage, item.symbol)
            if key in seen:
                continue
            seen.add(key)
            items.append(
                StockEvidenceRecord(
                    symbol=item.symbol,
                    name=item.name,
                    stage=stage,
                    portfolio_role=item.portfolio_role or ("WATCHLIST" if item.symbol in {x.symbol for x in watchlist} else "SATELLITE"),
                    evidence_summary=item.evidence_summary,
                    evidence_cards=item.evidence_cards,
                    positive_reasons=item.positive_reasons,
                    veto_reasons=item.veto_reasons,
                    theme_tags=item.theme_tags,
                    sector_tags=[value for value in (item.industry, item.sector) if value],
                    risk_flags=item.risk_flags,
                )
            )
    return items


def _build_evaluation_summary(portfolio: list[StockFeature], watchlist: list[StockFeature]) -> dict[str, object]:
    horizons = {str(day): {"status": "PENDING"} for day in (1, 3, 5, 10, 20)}
    return {
        "status": "PENDING",
        "message": "日终评估任务待补写",
        "benchmark_symbol": "sh000001",
        "portfolio_count": len(portfolio),
        "watchlist_count": len(watchlist),
        **horizons,
    }


def _build_template_snapshot(payload: StockSelectionPayload, market_regime: str) -> dict[str, object]:
    snapshot = dict(payload.template_snapshot)
    if payload.template_id:
        snapshot.setdefault("id", payload.template_id)
    if payload.template_key:
        snapshot.setdefault("template_key", payload.template_key)
    if payload.template_name:
        snapshot.setdefault("name", payload.template_name)
    snapshot.setdefault("market_regime_runtime", market_regime)
    return snapshot


def _feature_risk_summary(item: StockFeature) -> str:
    parts = [f"风险 {item.risk_level}", f"波动 {item.volatility20:.2f}%"]
    if item.risk_flags:
        parts.append(" / ".join(item.risk_flags[:2]))
    return "；".join(parts)
