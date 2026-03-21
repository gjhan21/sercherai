from __future__ import annotations

from app.schemas.simulation import AgentOpinion


class EventAgent:
    name = "event"

    def evaluate(self, context: dict) -> AgentOpinion:
        news_bias = float(context.get("news_bias", 0))
        news_heat = float(context.get("news_heat", 0))
        if news_bias > 0.2 or news_heat >= 4:
            return AgentOpinion(agent=self.name, stance="POSITIVE", confidence=70, summary="事件/资讯面偏正向，利于观点扩散。")
        if news_bias < -0.15:
            return AgentOpinion(agent=self.name, stance="NEGATIVE", confidence=72, summary="事件扰动偏空，需要防范预期反转。")
        return AgentOpinion(agent=self.name, stance="CAUTION", confidence=55, summary="事件面中性，暂不单独放大仓位。")
