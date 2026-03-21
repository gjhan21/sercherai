from __future__ import annotations

from app.domain.agents.basis_agent import BasisAgent
from app.domain.agents.event_agent import EventAgent
from app.domain.agents.liquidity_agent import LiquidityAgent
from app.domain.agents.risk_agent import RiskAgent
from app.domain.agents.trend_agent import TrendAgent
from app.schemas.simulation import AgentOpinion


class AgentPanel:
    def __init__(self) -> None:
        self._agents = [
            TrendAgent(),
            EventAgent(),
            LiquidityAgent(),
            RiskAgent(),
            BasisAgent(),
        ]
        self._agent_map = {agent.name: agent for agent in self._agents}

    def review(self, context: dict, options: dict | None = None) -> tuple[list[AgentOpinion], str, bool, str]:
        options = options or {}
        enabled_names = [item.strip().lower() for item in options.get("enabled_agents", []) if str(item).strip()]
        selected_agents = self._agents
        if enabled_names:
            selected_agents = [self._agent_map[name] for name in enabled_names if name in self._agent_map]
        if not selected_agents:
            selected_agents = self._agents

        allow_veto = bool(options.get("allow_veto", True))
        positive_threshold = max(int(options.get("positive_threshold", 3) or 3), 1)
        negative_threshold = max(int(options.get("negative_threshold", 2) or 2), 1)

        opinions = [agent.evaluate(context) for agent in selected_agents]
        vetoes = [item for item in opinions if item.veto]
        if allow_veto and vetoes:
            return opinions, "REJECT", True, vetoes[0].summary

        positive = sum(1 for item in opinions if item.stance == "POSITIVE")
        negative = sum(1 for item in opinions if item.stance == "NEGATIVE")
        if positive >= positive_threshold:
            return opinions, "GO", False, ""
        if negative >= negative_threshold:
            return opinions, "HOLD", False, ""
        return opinions, "WATCH", False, ""
