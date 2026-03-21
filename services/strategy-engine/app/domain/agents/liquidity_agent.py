from __future__ import annotations

from app.schemas.simulation import AgentOpinion


class LiquidityAgent:
    name = "liquidity"

    def evaluate(self, context: dict) -> AgentOpinion:
        volume_ratio = float(context.get("volume_ratio", 0))
        oi_change = float(context.get("oi_change_pct", 0))
        if volume_ratio >= 1.2 and oi_change >= 0:
            return AgentOpinion(agent=self.name, stance="POSITIVE", confidence=73, summary="流动性与增仓信号配合，执行阻力较小。")
        if volume_ratio < 0.95:
            return AgentOpinion(agent=self.name, stance="NEGATIVE", confidence=67, summary="流动性不足，容易放大滑点与波动。")
        return AgentOpinion(agent=self.name, stance="CAUTION", confidence=57, summary="流动性一般，建议控制节奏和仓位。")
