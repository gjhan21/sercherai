from __future__ import annotations

from datetime import datetime
from time import perf_counter
from typing import Optional

from app.domain.agents.agent_panel import AgentPanel
from app.domain.features.futures_feature_factory import FuturesFeatureFactory
from app.domain.graph.market_graph_builder import MarketGraphBuilder
from app.domain.graph.strategy_graph_client import StrategyGraphClient
from app.domain.models import FuturesFeature, FuturesSeed
from app.domain.reports.futures_report_builder import FuturesReportBuilder
from app.domain.risk.leverage_guard import LeverageGuard
from app.domain.scenarios.futures_scenario_engine import FuturesScenarioEngine
from app.domain.seeds.futures_seed_loader import FuturesSeedLoader
from app.domain.selectors.futures_selector import FuturesSelector
from app.schemas.futures import (
    FuturesCandidateSnapshot,
    FuturesEvidenceRecord,
    FuturesStageLog,
    FuturesStrategyPayload,
    FuturesStrategyReport,
)
from app.schemas.research import MemoryFeedback, MemoryFeedbackItem, StrategyGraphWriteResult


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
        strategy_graph_client: Optional[StrategyGraphClient] = None,
    ) -> None:
        self._futures_seed_loader = futures_seed_loader
        self._futures_feature_factory = futures_feature_factory
        self._futures_selector = futures_selector
        self._leverage_guard = leverage_guard
        self._futures_report_builder = futures_report_builder
        self._futures_scenario_engine = futures_scenario_engine or FuturesScenarioEngine(agent_panel=AgentPanel())
        self._market_graph_builder = market_graph_builder or MarketGraphBuilder()
        self._strategy_graph_client = strategy_graph_client

    def run(self, raw_payload: dict) -> tuple[FuturesStrategyReport, list[str]]:
        payload = FuturesStrategyPayload.model_validate(raw_payload)
        if not payload.trade_date:
            payload.trade_date = datetime.now().strftime("%Y-%m-%d")
        run_id = payload.run_id.strip() or _build_run_id(payload)
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

        stage_counts: dict[str, int] = {}
        stage_durations_ms: dict[str, int] = {}
        stage_logs: list[FuturesStageLog] = []

        regime_start = perf_counter()
        market_regime = _detect_market_regime(seed_result.seeds, payload)
        regime_duration = _duration_ms(regime_start)
        stage_counts["MARKET_REGIME"] = len(seed_result.seeds)
        stage_durations_ms["MARKET_REGIME"] = regime_duration
        stage_logs.append(
            FuturesStageLog(
                stage_key="MARKET_REGIME",
                stage_order=1,
                input_count=len(seed_result.seeds),
                output_count=len(seed_result.seeds),
                duration_ms=regime_duration,
                detail_message=f"已识别期货市场状态：{_market_regime_label(market_regime)}",
                payload_snapshot={"template_key": payload.template_key, "style": payload.style},
            )
        )

        universe_start = perf_counter()
        features = self._futures_feature_factory.build(seed_result.seeds)
        universe_duration = _duration_ms(universe_start)
        stage_counts["UNIVERSE"] = len(features)
        stage_durations_ms["UNIVERSE"] = universe_duration
        stage_logs.append(
            FuturesStageLog(
                stage_key="UNIVERSE",
                stage_order=2,
                input_count=len(seed_result.seeds),
                output_count=len(features),
                duration_ms=universe_duration,
                detail_message=f"合约池标准化完成，风格 {payload.style}",
                payload_snapshot={"meta_source": seed_result.meta.get("source", "")},
            )
        )

        graph_start = perf_counter()
        graph_snapshot = self._market_graph_builder.build_futures(
            features,
            run_id=run_id,
            trade_date=payload.trade_date,
            market_regime=market_regime,
        )
        graph_duration = _duration_ms(graph_start)
        stage_counts["GRAPH_ENRICHMENT"] = len(graph_snapshot.entities)
        stage_durations_ms["GRAPH_ENRICHMENT"] = graph_duration
        stage_logs.append(
            FuturesStageLog(
                stage_key="GRAPH_ENRICHMENT",
                stage_order=3,
                input_count=len(features),
                output_count=len(graph_snapshot.entities),
                duration_ms=graph_duration,
                detail_message="期货图谱增强完成",
                payload_snapshot={
                    "related_entities": [item.label for item in graph_snapshot.related_entities[:6]],
                    "relation_count": len(graph_snapshot.relations),
                },
            )
        )

        candidate_start = perf_counter()
        ranked = self._futures_selector.select(features, payload.limit)
        candidate_duration = _duration_ms(candidate_start)
        stage_counts["CANDIDATE_POOL"] = len(ranked)
        stage_durations_ms["CANDIDATE_POOL"] = candidate_duration
        stage_logs.append(
            FuturesStageLog(
                stage_key="CANDIDATE_POOL",
                stage_order=4,
                input_count=len(features),
                output_count=len(ranked),
                duration_ms=candidate_duration,
                detail_message="已按置信度完成期货候选排序",
                payload_snapshot={"top_contracts": [item.contract for item in ranked[:10]]},
            )
        )

        portfolio_start = perf_counter()
        guarded, warnings = self._leverage_guard.apply(ranked, payload)
        portfolio_duration = _duration_ms(portfolio_start)
        stage_counts["PORTFOLIO"] = len(guarded)
        stage_durations_ms["PORTFOLIO"] = portfolio_duration
        stage_logs.append(
            FuturesStageLog(
                stage_key="PORTFOLIO",
                stage_order=5,
                input_count=len(ranked),
                output_count=len(guarded),
                duration_ms=portfolio_duration,
                detail_message=(
                    f"组合约束完成：持仓上限 {payload.limit}，"
                    f"最大风险 {_risk_level_label(payload.max_risk_level)}"
                ),
                payload_snapshot={"selected_contracts": [item.contract for item in guarded]},
            )
        )

        warnings = [*seed_result.warnings, *warnings]
        graph_write_result = _write_graph_snapshot(self._strategy_graph_client, graph_snapshot)
        if graph_write_result.snapshot_id:
            graph_snapshot.snapshot_id = graph_write_result.snapshot_id
        if graph_write_result.status == "FAILED" and graph_write_result.error_message:
            warnings.append(f"图谱快照写入失败：{graph_write_result.error_message}")

        candidate_snapshots = _build_candidate_snapshots(features, ranked, guarded)
        evidence_records = _build_evidence_records(ranked, guarded)
        evaluation_summary = _build_evaluation_summary(guarded)
        memory_feedback = _build_memory_feedback(market_regime=market_regime, warnings=warnings, portfolio=guarded)
        report = self._futures_report_builder.build(
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
            graph_snapshot=graph_snapshot,
            memory_feedback=memory_feedback,
        )
        stage_logs.append(
            FuturesStageLog(
                stage_key="REVIEW_PAYLOAD",
                stage_order=6,
                input_count=len(guarded),
                output_count=len(report.publish_payloads),
                duration_ms=0,
                detail_message="已生成可审核发布载荷",
                payload_snapshot={
                    "graph_snapshot_id": graph_snapshot.snapshot_id,
                    "publish_count": len(report.publish_payloads),
                },
            )
        )
        stage_logs.append(
            FuturesStageLog(
                stage_key="FORWARD_EVALUATION",
                stage_order=7,
                status="PENDING",
                input_count=len(guarded),
                output_count=0,
                duration_ms=0,
                detail_message="已登记日终异步评估补写",
                payload_snapshot={"horizons": [1, 3, 5, 10, 20]},
            )
        )
        stage_logs.append(
            FuturesStageLog(
                stage_key="MEMORY_FEEDBACK",
                stage_order=8,
                input_count=len(guarded),
                output_count=len(memory_feedback.items),
                duration_ms=0,
                detail_message=memory_feedback.summary,
                payload_snapshot={"suggestions": memory_feedback.suggestions[:3]},
            )
        )
        report.stage_logs = stage_logs
        report.context_meta = seed_result.meta
        report.context_meta["market_regime"] = market_regime
        report.context_meta["run_id"] = run_id
        report.context_meta["graph_snapshot_id"] = graph_snapshot.snapshot_id
        report.context_meta["graph_write_status"] = graph_write_result.status
        report.simulations = self._futures_scenario_engine.simulate(guarded, agent_options)
        report.consensus_summary = _build_consensus_summary(report.simulations)
        return report, warnings


def _build_consensus_summary(simulations: list) -> str:
    veto_count = sum(1 for item in simulations if item.vetoed)
    go_count = sum(1 for item in simulations if item.consensus_action == "GO")
    watch_count = sum(1 for item in simulations if item.consensus_action == "WATCH")
    return f"多代理收敛：通过 {go_count}，观察 {watch_count}，否决 {veto_count}。"


def _duration_ms(started_at: float) -> int:
    return int((perf_counter() - started_at) * 1000)


def _build_run_id(payload: FuturesStrategyPayload) -> str:
    trade_date = payload.trade_date or datetime.now().strftime("%Y-%m-%d")
    if payload.profile_id:
        return f"futures-{payload.profile_id}-{trade_date}"
    return f"futures-auto-{trade_date}"


def _detect_market_regime(seeds: list[FuturesSeed], payload: FuturesStrategyPayload) -> str:
    bias = str(payload.template_snapshot.get("market_regime_bias", "") or "").strip().upper()
    if bias in {"BASE", "TREND_CONTINUE", "POLICY_POSITIVE", "POLICY_NEGATIVE", "SUPPLY_SHOCK", "LIQUIDITY_SHOCK"}:
        return bias
    if not seeds:
        return "BASE"
    positive_news = sum(1 for item in seeds if item.news_bias >= 0.25)
    negative_news = sum(1 for item in seeds if item.news_bias <= -0.25)
    strong_trend = sum(1 for item in seeds if item.trend_strength >= 1.0)
    high_inventory_pressure = sum(1 for item in seeds if item.inventory_pressure <= -0.18)
    high_liquidity_shock = sum(1 for item in seeds if item.volume_ratio >= 1.5 and abs(item.basis_pct) >= 0.8)
    if high_liquidity_shock >= max(1, len(seeds) // 3):
        return "LIQUIDITY_SHOCK"
    if high_inventory_pressure >= max(1, len(seeds) // 3):
        return "SUPPLY_SHOCK"
    if positive_news > negative_news and positive_news >= max(1, len(seeds) // 3):
        return "POLICY_POSITIVE"
    if negative_news > positive_news and negative_news >= max(1, len(seeds) // 3):
        return "POLICY_NEGATIVE"
    if strong_trend >= max(1, len(seeds) // 2):
        return "TREND_CONTINUE"
    return "BASE"


def _build_candidate_snapshots(
    universe: list[FuturesFeature],
    candidate_pool: list[FuturesFeature],
    portfolio: list[FuturesFeature],
) -> list[FuturesCandidateSnapshot]:
    result: list[FuturesCandidateSnapshot] = []
    selected_contracts = {item.contract for item in candidate_pool}
    portfolio_contracts = {item.contract for item in portfolio}
    result.extend(_build_stage_snapshots("UNIVERSE", _rank_by_score(universe), selected_contracts))
    result.extend(_build_stage_snapshots("CANDIDATE_POOL", candidate_pool, portfolio_contracts))
    result.extend(_build_stage_snapshots("PORTFOLIO", portfolio, portfolio_contracts))
    return result


def _build_stage_snapshots(
    stage: str,
    ordered: list[FuturesFeature],
    selected_contracts: set[str],
) -> list[FuturesCandidateSnapshot]:
    snapshots: list[FuturesCandidateSnapshot] = []
    for index, item in enumerate(ordered, start=1):
        snapshots.append(
            FuturesCandidateSnapshot(
                contract=item.contract,
                name=item.name,
                stage=stage,
                score=item.conviction_score,
                risk_level=item.risk_level,
                selected=item.contract in selected_contracts,
                rank=index,
                direction=item.direction,
                reason_summary=item.reason_summary,
                evidence_summary=item.evidence_summary,
                inventory_factor_summary=item.inventory_factor_summary,
                structure_factor_summary=item.structure_factor_summary,
                portfolio_role=item.portfolio_role or ("CORE" if index <= 2 else "SATELLITE"),
                risk_summary=_item_risk_summary(item),
                factor_breakdown_json=item.factor_breakdown(),
            )
        )
    return snapshots


def _rank_by_score(features: list[FuturesFeature]) -> list[FuturesFeature]:
    return sorted(features, key=lambda item: (-item.conviction_score, item.contract))


def _build_evidence_records(
    candidate_pool: list[FuturesFeature],
    portfolio: list[FuturesFeature],
) -> list[FuturesEvidenceRecord]:
    items: list[FuturesEvidenceRecord] = []
    seen: set[tuple[str, str]] = set()
    for stage, features in (("CANDIDATE_POOL", candidate_pool), ("PORTFOLIO", portfolio)):
        for item in features:
            key = (stage, item.contract)
            if key in seen:
                continue
            seen.add(key)
            items.append(
                FuturesEvidenceRecord(
                    contract=item.contract,
                    name=item.name,
                    stage=stage,
                    portfolio_role=item.portfolio_role or "SATELLITE",
                    evidence_summary=item.evidence_summary,
                    inventory_factor_summary=item.inventory_factor_summary,
                    structure_factor_summary=item.structure_factor_summary,
                    evidence_cards=item.evidence_cards,
                    positive_reasons=item.positive_reasons,
                    veto_reasons=item.veto_reasons,
                    risk_flags=item.risk_flags,
                    related_entities=item.related_entities,
                )
            )
    return items


def _build_evaluation_summary(portfolio: list[FuturesFeature]) -> dict[str, object]:
    horizons = {str(day): {"status": "PENDING"} for day in (1, 3, 5, 10, 20)}
    return {
        "status": "PENDING",
        "message": "日终评估任务待补写",
        "benchmark_symbol": "FI00",
        "portfolio_count": len(portfolio),
        **horizons,
    }


def _build_template_snapshot(payload: FuturesStrategyPayload, market_regime: str) -> dict[str, object]:
    snapshot = dict(payload.template_snapshot)
    if payload.template_id:
        snapshot.setdefault("id", payload.template_id)
    if payload.template_key:
        snapshot.setdefault("template_key", payload.template_key)
    if payload.template_name:
        snapshot.setdefault("name", payload.template_name)
    snapshot.setdefault("market_regime_runtime", market_regime)
    return snapshot


def _item_risk_summary(item: FuturesFeature) -> str:
    parts = [f"风险 {item.risk_level}", f"波动 {item.volatility14:.2f}%"]
    if item.risk_flags:
        parts.append(" / ".join(item.risk_flags[:2]))
    return "；".join(parts)


def _write_graph_snapshot(
    strategy_graph_client: StrategyGraphClient | None,
    graph_snapshot,
) -> StrategyGraphWriteResult:
    if strategy_graph_client is None:
        return StrategyGraphWriteResult(status="SKIPPED")
    return strategy_graph_client.write_snapshot(graph_snapshot)


def _build_memory_feedback(
    *,
    market_regime: str,
    warnings: list[str],
    portfolio: list[FuturesFeature],
) -> MemoryFeedback:
    suggestions: list[str] = []
    failure_signals: list[str] = []
    items: list[MemoryFeedbackItem] = []
    high_risk_count = sum(1 for item in portfolio if item.risk_level == "HIGH")
    if high_risk_count >= max(1, len(portfolio) // 2):
        suggestions.append("下轮期货运行可提高最小置信度并继续收缩高波动仓位。")
        items.append(
            MemoryFeedbackItem(
                title="高风险合约占比偏高",
                level="WARN",
                detail=f"当前组合中高风险合约 {high_risk_count} 个。",
                suggestion="优先保留趋势与结构同时确认的合约。",
                source="portfolio",
            )
        )
    for warning in warnings[:3]:
        failure_signals.append(warning)
        items.append(
            MemoryFeedbackItem(
                title="风控提醒",
                level="WARN",
                detail=warning,
                suggestion="下次运行前复核最小置信度与风险阈值。",
                source="warnings",
            )
        )
    if market_regime == "SUPPLY_SHOCK":
        suggestions.append("当前偏供给冲击，可提高库存/仓单与价差相关因子权重。")
    elif market_regime == "LIQUIDITY_SHOCK":
        suggestions.append("当前偏流动性冲击，建议降低高波动合约的默认仓位。")
    summary = f"记忆反馈已生成：{len(items)} 条提示，市场状态 {_market_regime_label(market_regime)}。"
    return MemoryFeedback(
        summary=summary,
        suggestions=list(dict.fromkeys(suggestions))[:4],
        failure_signals=failure_signals[:4],
        items=items[:6],
    )


def _market_regime_label(value: str) -> str:
    mapping = {
        "BASE": "基准状态",
        "TREND_CONTINUE": "趋势延续",
        "POLICY_POSITIVE": "政策利多",
        "POLICY_NEGATIVE": "政策利空",
        "SUPPLY_SHOCK": "供给冲击",
        "LIQUIDITY_SHOCK": "流动性冲击",
    }
    key = str(value or "").strip().upper()
    return mapping.get(key, key or "-")


def _risk_level_label(value: str) -> str:
    mapping = {
        "LOW": "低风险",
        "MEDIUM": "中风险",
        "HIGH": "高风险",
    }
    key = str(value or "").strip().upper()
    return mapping.get(key, key or "-")
