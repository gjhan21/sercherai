from __future__ import annotations

from datetime import datetime, timedelta

from app.domain.models import FuturesFeature
from app.schemas.futures import (
    FuturesCandidateSnapshot,
    FuturesEvaluationRecord,
    FuturesEvidenceRecord,
    FuturesGuidanceWriteModel,
    FuturesPortfolioEntry,
    FuturesPublishPayload,
    FuturesStageLog,
    FuturesStrategyCandidate,
    FuturesStrategyPayload,
    FuturesStrategyReport,
    FuturesStrategyWriteModel,
)
from app.schemas.research import MemoryFeedback, ResearchGraphSnapshot


class FuturesReportBuilder:
    def build(
        self,
        payload: FuturesStrategyPayload,
        features: list[FuturesFeature],
        warnings: list[str],
        *,
        market_regime: str = "BASE",
        template_snapshot: dict | None = None,
        evaluation_summary: dict | None = None,
        stage_counts: dict[str, int] | None = None,
        stage_durations_ms: dict[str, int] | None = None,
        stage_logs: list[FuturesStageLog] | None = None,
        evidence_records: list[FuturesEvidenceRecord] | None = None,
        evaluation_records: list[FuturesEvaluationRecord] | None = None,
        candidate_snapshots: list[FuturesCandidateSnapshot] | None = None,
        graph_snapshot: ResearchGraphSnapshot | None = None,
        memory_feedback: MemoryFeedback | None = None,
    ) -> FuturesStrategyReport:
        trade_date = payload.trade_date or datetime.now().strftime("%Y-%m-%d")
        valid_from = f"{trade_date}T00:00:00Z"
        valid_to = (datetime.strptime(trade_date, "%Y-%m-%d") + timedelta(days=1)).strftime("%Y-%m-%dT00:00:00Z")

        strategies = []
        publish_payloads = []
        portfolio_entries = []
        risk_counter: dict[str, int] = {"LOW": 0, "MEDIUM": 0, "HIGH": 0}
        for index, item in enumerate(features, start=1):
            risk_counter[item.risk_level] = risk_counter.get(item.risk_level, 0) + 1
            portfolio_role = item.portfolio_role or _portfolio_role_for_index(index)
            risk_summary = _item_risk_summary(item)
            strategies.append(
                FuturesStrategyCandidate(
                    contract=item.contract,
                    name=item.name,
                    direction=item.direction,
                    entry_price=item.entry_price,
                    take_profit_price=item.take_profit_price,
                    stop_loss_price=item.stop_loss_price,
                    risk_level=item.risk_level,
                    position_range=item.position_range,
                    reason_summary=item.reason_summary,
                    invalidations=_invalidations(item),
                    evidence_summary=item.evidence_summary,
                    portfolio_role=portfolio_role,
                    risk_summary=risk_summary,
                    evidence_cards=item.evidence_cards,
                    positive_reasons=item.positive_reasons,
                    veto_reasons=item.veto_reasons,
                    risk_flags=item.risk_flags,
                    related_entities=item.related_entities,
                    evaluation_status="PENDING",
                )
            )
            portfolio_entries.append(
                FuturesPortfolioEntry(
                    contract=item.contract,
                    name=item.name,
                    rank=index,
                    direction=item.direction,
                    score=item.conviction_score,
                    risk_level=item.risk_level,
                    position_range=item.position_range,
                    reason_summary=item.reason_summary,
                    evidence_summary=item.evidence_summary,
                    portfolio_role=portfolio_role,
                    risk_summary=risk_summary,
                    evidence_cards=item.evidence_cards,
                    positive_reasons=item.positive_reasons,
                    veto_reasons=item.veto_reasons,
                    risk_flags=item.risk_flags,
                    related_entities=item.related_entities,
                    evaluation_status="PENDING",
                    factor_breakdown_json=item.factor_breakdown(),
                )
            )
            publish_payloads.append(
                FuturesPublishPayload(
                    strategy=FuturesStrategyWriteModel(
                        contract=item.contract,
                        name=item.name,
                        direction=item.direction,
                        risk_level=item.risk_level,
                        position_range=item.position_range,
                        valid_from=valid_from,
                        valid_to=valid_to,
                        status="PUBLISHED",
                        reason_summary=item.reason_summary,
                    ),
                    guidance=FuturesGuidanceWriteModel(
                        contract=item.contract,
                        guidance_direction=item.direction,
                        position_level=item.position_level,
                        entry_range=_price_range(item.entry_price),
                        take_profit_range=_price_range(item.take_profit_price),
                        stop_loss_range=_price_range(item.stop_loss_price),
                        risk_level=item.risk_level,
                        invalid_condition="；".join(_invalidations(item)),
                        valid_to=valid_to,
                    ),
                )
            )

        stage_counts = stage_counts or {}
        universe_count = stage_counts.get("UNIVERSE", 0)
        candidate_count = stage_counts.get("CANDIDATE_POOL", len(features))
        market_regime_label = _market_regime_label(market_regime)
        report_summary = (
            f"本次处于 {market_regime_label}，从 {universe_count} 个合约池筛出 {candidate_count} 个方向候选，"
            f"最终形成 {len(strategies)} 条可发布策略。"
        )
        if warnings:
            report_summary = f"{report_summary} 风控提醒：{'；'.join(warnings)}"
        risk_summary = (
            f"市场状态 {market_regime_label}；低风险 {risk_counter['LOW']} / 中风险 {risk_counter['MEDIUM']} / 高风险 {risk_counter['HIGH']}"
        )
        return FuturesStrategyReport(
            trade_date=trade_date,
            report_summary=report_summary,
            risk_summary=risk_summary,
            selected_count=len(strategies),
            market_regime=market_regime,
            context_meta={},
            graph_summary=graph_snapshot.summary if graph_snapshot else "",
            graph_snapshot_id=graph_snapshot.snapshot_id if graph_snapshot else "",
            template_snapshot=template_snapshot or {},
            evaluation_summary=evaluation_summary or {},
            related_entities=list(graph_snapshot.related_entities) if graph_snapshot else [],
            graph_entities=list(graph_snapshot.entities) if graph_snapshot else [],
            graph_relations=list(graph_snapshot.relations) if graph_snapshot else [],
            memory_feedback=memory_feedback or MemoryFeedback(),
            stage_counts=stage_counts,
            stage_durations_ms=stage_durations_ms or {},
            stage_logs=stage_logs or [],
            evidence_records=evidence_records or [],
            evaluation_records=evaluation_records or [],
            candidate_snapshots=candidate_snapshots or [],
            portfolio_entries=portfolio_entries,
            strategies=strategies,
            publish_payloads=publish_payloads,
        )


def _invalidations(item: FuturesFeature) -> list[str]:
    if item.direction == "SHORT":
        base = ["价格重新站回入场带上沿", "持仓反向减仓且量比低于1", "基差快速反向收敛"]
    else:
        base = ["价格跌破入场带下沿", "增仓信号消失且量比低于1", "基差反向扩大并削弱顺势逻辑"]
    if item.risk_level == "HIGH":
        base.append("波动率继续放大，超出轻仓容忍区间")
    base.extend(item.veto_reasons[:2])
    return base


def _price_range(price: float) -> str:
    lower = round(price * 0.998, 2)
    upper = round(price * 1.002, 2)
    return f"{lower}-{upper}"


def _portfolio_role_for_index(index: int) -> str:
    return "CORE" if index <= 2 else "SATELLITE"


def _item_risk_summary(item: FuturesFeature) -> str:
    parts = [f"风险级别 {_risk_level_label(item.risk_level)}", f"波动率 {item.volatility14:.2f}%", f"基差 {item.basis_pct:.2f}%"]
    if item.risk_flags:
        parts.append("风险旗标 " + " / ".join(item.risk_flags[:2]))
    return "；".join(parts)


def _market_regime_label(value: str) -> str:
    return {
        "BASE": "基准状态",
        "TREND_CONTINUE": "趋势延续",
        "POLICY_POSITIVE": "政策利多",
        "POLICY_NEGATIVE": "政策利空",
        "SUPPLY_SHOCK": "供给冲击",
        "LIQUIDITY_SHOCK": "流动性冲击",
    }.get(str(value or "").strip().upper(), value)


def _risk_level_label(value: str) -> str:
    return {
        "LOW": "低风险",
        "MEDIUM": "中风险",
        "HIGH": "高风险",
    }.get(str(value or "").strip().upper(), value)
