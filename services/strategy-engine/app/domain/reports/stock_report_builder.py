from __future__ import annotations

from datetime import datetime, timedelta

from app.domain.models import StockFeature
from app.schemas.stock import (
    StockCandidate,
    StockCandidateSnapshot,
    StockEvidenceRecord,
    StockEvaluationRecord,
    StockPortfolioEntry,
    StockPublishPayload,
    StockRecommendationDetailWriteModel,
    StockRecommendationWriteModel,
    StockSelectionPayload,
    StockSelectionReport,
    StockStageLog,
)


class StockReportBuilder:
    def build(
        self,
        payload: StockSelectionPayload,
        features: list[StockFeature],
        warnings: list[str],
        *,
        market_regime: str = "ROTATION",
        template_snapshot: dict | None = None,
        evaluation_summary: dict | None = None,
        stage_counts: dict[str, int] | None = None,
        stage_durations_ms: dict[str, int] | None = None,
        stage_logs: list[StockStageLog] | None = None,
        evidence_records: list[StockEvidenceRecord] | None = None,
        evaluation_records: list[StockEvaluationRecord] | None = None,
        candidate_snapshots: list[StockCandidateSnapshot] | None = None,
        watchlist: list[StockFeature] | None = None,
    ) -> StockSelectionReport:
        trade_date = payload.trade_date or datetime.now().strftime("%Y-%m-%d")
        valid_from = f"{trade_date}T00:00:00Z"
        valid_to = (datetime.strptime(trade_date, "%Y-%m-%d") + timedelta(days=1)).strftime("%Y-%m-%dT00:00:00Z")
        watchlist = watchlist or []
        strategy_version = _strategy_version(payload, market_regime)

        candidates = []
        publish_payloads = []
        portfolio_entries = []
        risk_counter: dict[str, int] = {"LOW": 0, "MEDIUM": 0, "HIGH": 0}
        for index, item in enumerate(features, start=1):
            risk_counter[item.risk_level] = risk_counter.get(item.risk_level, 0) + 1
            take_profit, stop_loss, position_range = _risk_controls(item.risk_level)
            factor_breakdown = item.factor_breakdown()
            candidates.append(
                StockCandidate(
                    symbol=item.symbol,
                    name=item.name,
                    score=item.score,
                    risk_level=item.risk_level,
                    position_range=position_range,
                    reason_summary=item.reason_summary,
                    invalidations=_invalidations(item),
                    take_profit=take_profit,
                    stop_loss=stop_loss,
                    evidence_summary=item.evidence_summary,
                    portfolio_role=item.portfolio_role or _portfolio_role_for_index(index),
                    risk_summary=_item_risk_summary(item),
                    evidence_cards=item.evidence_cards,
                    positive_reasons=item.positive_reasons,
                    veto_reasons=item.veto_reasons,
                    theme_tags=item.theme_tags,
                    sector_tags=_sector_tags(item),
                    risk_flags=item.risk_flags,
                    evaluation_status="PENDING",
                )
            )
            portfolio_entries.append(
                StockPortfolioEntry(
                    symbol=item.symbol,
                    name=item.name,
                    rank=index,
                    quant_score=item.quant_score,
                    total_score=item.score,
                    risk_level=item.risk_level,
                    weight_suggestion=position_range,
                    reason_summary=item.reason_summary,
                    evidence_summary=item.evidence_summary,
                    portfolio_role=item.portfolio_role or _portfolio_role_for_index(index),
                    risk_summary=_item_risk_summary(item),
                    evidence_cards=item.evidence_cards,
                    positive_reasons=item.positive_reasons,
                    veto_reasons=item.veto_reasons,
                    theme_tags=item.theme_tags,
                    sector_tags=_sector_tags(item),
                    risk_flags=item.risk_flags,
                    evaluation_status="PENDING",
                    factor_breakdown_json=factor_breakdown,
                )
            )
            publish_payloads.append(
                StockPublishPayload(
                    recommendation=StockRecommendationWriteModel(
                        symbol=item.symbol,
                        name=item.name,
                        score=item.score,
                        risk_level=item.risk_level,
                        position_range=position_range,
                        valid_from=valid_from,
                        valid_to=valid_to,
                        status="PUBLISHED",
                        reason_summary=item.reason_summary,
                        source_type="SYSTEM",
                        strategy_version=strategy_version,
                        publisher="strategy-engine",
                        review_note="由 strategy-engine 智能选股 V2 自动生成，待人工复核发布。",
                        performance_label="ESTIMATED",
                    ),
                    detail=StockRecommendationDetailWriteModel(
                        tech_score=round(_clamp(item.trend_score, 55, 98), 2),
                        fund_score=round(_clamp(item.quality_score, 52, 97), 2),
                        sentiment_score=round(_clamp(item.event_score, 50, 96), 2),
                        money_flow_score=round(_clamp(item.flow_score, 50, 97), 2),
                        take_profit=take_profit,
                        stop_loss=stop_loss,
                        risk_note="量化信号存在失效风险，仅供参考，不构成投资建议",
                    ),
                )
            )

        stage_counts = stage_counts or {}
        universe_count = stage_counts.get("UNIVERSE", 0)
        seed_pool_count = stage_counts.get("SEED_POOL", 0)
        candidate_count = stage_counts.get("CANDIDATE_POOL", 0)
        report_summary = (
            f"本次处于 {market_regime} 市场状态，从 {universe_count} 只股票池经 {seed_pool_count} 只种子池、"
            f"{candidate_count} 只候选池，最终形成 {len(candidates)} 只可发布组合，并保留 {len(watchlist)} 只观察名单。"
        )
        if warnings:
            report_summary = f"{report_summary} 风控提醒：{'；'.join(warnings)}"

        risk_summary = (
            f"市场状态 {market_regime}；LOW {risk_counter['LOW']} / MEDIUM {risk_counter['MEDIUM']} / HIGH {risk_counter['HIGH']}；"
            f"观察名单 {len(watchlist)}"
        )
        return StockSelectionReport(
            trade_date=trade_date,
            report_summary=report_summary,
            risk_summary=risk_summary,
            selected_count=len(candidates),
            market_regime=market_regime,
            template_snapshot=template_snapshot or {},
            evaluation_summary=evaluation_summary or {},
            stage_counts=stage_counts,
            stage_durations_ms=stage_durations_ms or {},
            stage_logs=stage_logs or [],
            evidence_records=evidence_records or [],
            evaluation_records=evaluation_records or [],
            candidate_snapshots=candidate_snapshots or [],
            portfolio_entries=portfolio_entries,
            candidates=candidates,
            publish_payloads=publish_payloads,
        )


def _risk_controls(risk_level: str) -> tuple[str, str, str]:
    if risk_level == "LOW":
        return "上涨8%-12%分批止盈", "回撤3%止损", "10%-15%"
    if risk_level == "HIGH":
        return "上涨12%-18%动态止盈", "回撤7%止损", "5%-8%"
    return "上涨10%-15%分批止盈", "回撤5%止损", "8%-12%"


def _invalidations(item: StockFeature) -> list[str]:
    invalidations = ["20日动量跌破0，趋势逻辑失效", "主力资金转负且量比回落到1以下"]
    if item.risk_level == "LOW":
        invalidations.append("回撤超过3%或跌破短期支撑位")
    elif item.risk_level == "HIGH":
        invalidations.append("回撤超过7%或高波动放大")
    else:
        invalidations.append("回撤超过5%或波动显著放大")
    if item.news_heat >= 4:
        invalidations.append("热点舆情转负，正面占比跌破35%")
    invalidations.extend(item.veto_reasons[:2])
    return invalidations


def _clamp(value: float, min_value: float, max_value: float) -> float:
    if value < min_value:
        return min_value
    if value > max_value:
        return max_value
    return value


def _strategy_version(payload: StockSelectionPayload, market_regime: str) -> str:
    template_key = str(payload.template_key or payload.template_snapshot.get("template_key", "")).strip()
    if template_key:
        return f"stock-selection-v2/{template_key}"
    if market_regime:
        return f"stock-selection-v2/{market_regime.lower()}"
    return "stock-selection-v2/default"


def _item_risk_summary(item: StockFeature) -> str:
    parts = [f"风险级别 {item.risk_level}", f"波动率 {item.volatility20:.2f}%", f"回撤 {item.drawdown20:.2f}%"]
    if item.risk_flags:
        parts.append("风险旗标 " + " / ".join(item.risk_flags[:2]))
    return "；".join(parts)


def _sector_tags(item: StockFeature) -> list[str]:
    return [value for value in (item.industry, item.sector) if value]


def _portfolio_role_for_index(index: int) -> str:
    return "CORE" if index <= 2 else "SATELLITE"
