from __future__ import annotations

from app.schemas.simulation import AgentOpinion


class TrendAgent:
    name = "trend"

    def evaluate(self, context: dict) -> AgentOpinion:
        score = float(context.get("trend_strength", 0))
        direction = context.get("direction", "")
        if score >= 1.5 or direction == "LONG":
            return AgentOpinion(agent=self.name, stance="POSITIVE", confidence=78, summary="趋势结构延续，顺势逻辑成立。")
        if score <= -1.0 or direction == "SHORT":
            return AgentOpinion(agent=self.name, stance="NEGATIVE", confidence=74, summary="趋势转弱，反身性压力偏大。")
        return AgentOpinion(agent=self.name, stance="CAUTION", confidence=58, summary="趋势尚未完全展开，需要结合其他信号。")
