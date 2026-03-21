from __future__ import annotations


class MarketGraphBuilder:
    def build_stock(self, features: list) -> str:
        strong_trend = sum(1 for item in features if getattr(item, "trend_strength", 0) >= 1.5)
        high_risk = sum(1 for item in features if getattr(item, "risk_level", "") == "HIGH")
        return f"股票市场图谱：趋势节点 {strong_trend}，高风险节点 {high_risk}。"

    def build_futures(self, features: list) -> str:
        long_count = sum(1 for item in features if getattr(item, "direction", "") == "LONG")
        short_count = sum(1 for item in features if getattr(item, "direction", "") == "SHORT")
        return f"期货市场图谱：多头节点 {long_count}，空头节点 {short_count}。"
