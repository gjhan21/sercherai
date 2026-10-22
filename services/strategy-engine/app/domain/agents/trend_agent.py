from __future__ import annotations

from app.schemas.simulation import AgentOpinion


class TrendAgent:
    name = "trend"

    def evaluate(self, context: dict) -> AgentOpinion:
        score = float(context.get("trend_strength", 0))
        direction = context.get("direction", "")
        velocity = float(context.get("price_velocity", 0))
        acceleration = float(context.get("price_acceleration", 0))
        divergence = float(context.get("vel_amp_divergence", 0))

        # 场景一：强势动能共振突破 (趋势分数正向, 速度快, 加速度为正)
        if score >= 1.0 and velocity > 0.02 and acceleration > 0:
            return AgentOpinion(
                agent=self.name,
                stance="POSITIVE",
                confidence=85,
                summary=f"趋势共振走强。上涨速度 {velocity*100:.1f}% 且加速度转正，动能正加速爆发。"
            )

        # 场景二：顶背离衰竭防护 (趋势分高，但速度极慢且背离度低)
        if score >= 1.2 and divergence < -0.05 and velocity < 0.005:
            return AgentOpinion(
                agent=self.name,
                stance="NEGATIVE",
                confidence=80,
                summary=f"警惕趋势见顶！幅度扩张但速度严重衰竭(背离度 {divergence:.3f})，多头动能枯竭。"
            )

        # 场景三：失速下跌加速 (速度大幅为负且加速度下行)
        if velocity < -0.015 and acceleration < 0:
            return AgentOpinion(
                agent=self.name,
                stance="NEGATIVE",
                confidence=78,
                summary=f"下跌动能恶化。下跌速度 {velocity*100:.1f}% 且负向加速，反身性抛压严重。"
            )

        # 默认旧策略回退
        if score >= 1.5 or direction == "LONG":
            return AgentOpinion(agent=self.name, stance="POSITIVE", confidence=78, summary="趋势结构延续，顺势逻辑成立。")
        if score <= -1.0 or direction == "SHORT":
            return AgentOpinion(agent=self.name, stance="NEGATIVE", confidence=74, summary="趋势转弱，反身性压力偏大。")
        return AgentOpinion(agent=self.name, stance="CAUTION", confidence=58, summary="趋势尚未完全展开，需要结合其他信号。")
