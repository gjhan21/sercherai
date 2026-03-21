from __future__ import annotations

from datetime import datetime, timedelta

from app.domain.models import FuturesFeature
from app.schemas.futures import (
    FuturesGuidanceWriteModel,
    FuturesPublishPayload,
    FuturesStrategyCandidate,
    FuturesStrategyPayload,
    FuturesStrategyReport,
    FuturesStrategyWriteModel,
)


class FuturesReportBuilder:
    def build(self, payload: FuturesStrategyPayload, features: list[FuturesFeature], warnings: list[str]) -> FuturesStrategyReport:
        trade_date = payload.trade_date or datetime.now().strftime("%Y-%m-%d")
        valid_from = f"{trade_date}T00:00:00Z"
        valid_to = (datetime.strptime(trade_date, "%Y-%m-%d") + timedelta(days=1)).strftime("%Y-%m-%dT00:00:00Z")

        strategies = []
        publish_payloads = []
        risk_counter: dict[str, int] = {"LOW": 0, "MEDIUM": 0, "HIGH": 0}
        for item in features:
            risk_counter[item.risk_level] = risk_counter.get(item.risk_level, 0) + 1
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

        report_summary = f"本次期货 MVP 生成 {len(strategies)} 条策略，聚焦方向、价位和风险控制建议。"
        if warnings:
            report_summary = f"{report_summary} 风控提醒：{'；'.join(warnings)}"
        risk_summary = f"LOW {risk_counter['LOW']} / MEDIUM {risk_counter['MEDIUM']} / HIGH {risk_counter['HIGH']}"
        return FuturesStrategyReport(
            trade_date=trade_date,
            report_summary=report_summary,
            risk_summary=risk_summary,
            selected_count=len(strategies),
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
    return base


def _price_range(price: float) -> str:
    lower = round(price * 0.998, 2)
    upper = round(price * 1.002, 2)
    return f"{lower}-{upper}"
