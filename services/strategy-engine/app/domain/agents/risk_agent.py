from __future__ import annotations

from app.schemas.simulation import AgentOpinion


class RiskAgent:
    name = "risk"

    def evaluate(self, context: dict) -> AgentOpinion:
        risk_level = context.get("risk_level", "MEDIUM")
        shock_drawdown = float(context.get("shock_adjustment", -5))
        if risk_level == "HIGH" and shock_drawdown <= -12:
            return AgentOpinion(
                agent=self.name,
                stance="NEGATIVE",
                confidence=88,
                summary="极端情景下损失放大，当前建议直接否决。",
                veto=True,
            )
        if risk_level == "HIGH":
            return AgentOpinion(agent=self.name, stance="NEGATIVE", confidence=78, summary="风险等级偏高，只适合轻仓跟踪。")
        if risk_level == "LOW":
            return AgentOpinion(agent=self.name, stance="POSITIVE", confidence=69, summary="风险可控，具备继续跟踪条件。")
        return AgentOpinion(agent=self.name, stance="CAUTION", confidence=60, summary="风险中性，需搭配场景止损。")
