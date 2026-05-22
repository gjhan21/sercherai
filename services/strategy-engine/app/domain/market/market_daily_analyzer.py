from __future__ import annotations

from dataclasses import dataclass, field

from app.domain.models import MarketSeed


@dataclass(slots=True)
class MarketAnalysisResult:
    trend_state: str
    next_day_rhythm: str
    risk_posture: str
    summary: str
    breadth_summary: dict[str, float] = field(default_factory=dict)


class MarketDailyAnalyzer:
    def analyze(self, seeds: list[MarketSeed]) -> MarketAnalysisResult:
        if not seeds:
            return MarketAnalysisResult(
                trend_state="RANGE_NEUTRAL",
                next_day_rhythm="NEUTRAL",
                risk_posture="NORMAL",
                summary="样本为空，按中性市场处理。",
                breadth_summary={},
            )

        avg_m5 = sum(item.momentum5 for item in seeds) / len(seeds)
        avg_m20 = sum(item.momentum20 for item in seeds) / len(seeds)
        avg_vol = sum(item.volatility20 for item in seeds) / len(seeds)
        positive_flow_ratio = sum(1 for item in seeds if item.net_mf_amount > 0) / len(seeds)
        positive_m20_ratio = sum(1 for item in seeds if item.momentum20 > 0) / len(seeds)

        if avg_vol >= 4.0 and avg_m20 <= -2.0 and positive_flow_ratio <= 0.4:
            trend_state = "RISK_OFF"
        elif avg_m20 >= 8.0 and positive_flow_ratio >= 0.6:
            trend_state = "UPTREND"
        elif avg_m20 >= 4.0 and positive_flow_ratio >= 0.55:
            trend_state = "RANGE_STRONG"
        elif avg_m20 <= -1.0:
            trend_state = "RANGE_WEAK"
        else:
            trend_state = "RANGE_NEUTRAL"

        if trend_state == "UPTREND":
            next_day_rhythm = "ATTACK"
            risk_posture = "NORMAL"
        elif trend_state == "RISK_OFF":
            next_day_rhythm = "DEFENSE"
            risk_posture = "TIGHT"
        elif trend_state == "RANGE_STRONG":
            next_day_rhythm = "REPAIR"
            risk_posture = "NORMAL"
        elif trend_state == "RANGE_WEAK":
            next_day_rhythm = "DEFENSE"
            risk_posture = "TIGHT"
        else:
            next_day_rhythm = "NEUTRAL"
            risk_posture = "NORMAL"

        breadth_summary = {
            "avg_momentum5": round(avg_m5, 4),
            "avg_momentum20": round(avg_m20, 4),
            "avg_volatility20": round(avg_vol, 4),
            "positive_flow_ratio": round(positive_flow_ratio, 4),
            "positive_momentum20_ratio": round(positive_m20_ratio, 4),
        }
        summary = (
            f"市场状态 {trend_state}，次日节奏 {next_day_rhythm}，"
            f"20日动量均值 {avg_m20:.2f}，正向资金占比 {positive_flow_ratio:.2%}。"
        )
        return MarketAnalysisResult(
            trend_state=trend_state,
            next_day_rhythm=next_day_rhythm,
            risk_posture=risk_posture,
            summary=summary,
            breadth_summary=breadth_summary,
        )
