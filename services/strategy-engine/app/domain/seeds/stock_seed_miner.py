from __future__ import annotations

from dataclasses import dataclass, field

from app.domain.models import StockFeature
from app.schemas.stock import StockSelectionPayload

DEFAULT_BUCKET_LIMIT = 36
DEFAULT_SEED_POOL_CAP = 180


@dataclass(slots=True)
class StockSeedMiningResult:
    seed_pool: list[StockFeature]
    bucket_members: dict[str, list[str]] = field(default_factory=dict)
    bucket_limits: dict[str, int] = field(default_factory=dict)


class StockSeedMiner:
    def __init__(self, bucket_limit: int = DEFAULT_BUCKET_LIMIT, seed_pool_cap: int = DEFAULT_SEED_POOL_CAP) -> None:
        self._bucket_limit = max(1, bucket_limit)
        self._seed_pool_cap = max(self._bucket_limit, seed_pool_cap)

    def mine(
        self,
        features: list[StockFeature],
        payload: StockSelectionPayload | None = None,
        market_regime: str = "ROTATION",
    ) -> StockSeedMiningResult:
        if not features:
            return StockSeedMiningResult(seed_pool=[])
        payload = payload or StockSelectionPayload(
            bucket_limit=self._bucket_limit,
            seed_pool_cap=self._seed_pool_cap,
        )

        definitions = (
            ("trend", lambda item: (item.trend_score, item.momentum20, item.momentum5)),
            ("money_flow", lambda item: (item.flow_score, item.net_mf_amount, item.volume_ratio)),
            ("quality", lambda item: (item.quality_score, -_risk_rank(item.risk_level), item.turnover_rate)),
            ("event", lambda item: (item.event_score, item.news_score, item.news_heat)),
            ("resonance", lambda item: (item.resonance_score, item.quant_score, item.flow_score)),
        )

        base_limit = max(1, payload.bucket_limit or self._bucket_limit)
        seed_pool_cap = max(base_limit, payload.seed_pool_cap or self._seed_pool_cap)
        bucket_limits = _resolve_bucket_limits(base_limit, payload, market_regime)
        seed_pool: list[StockFeature] = []
        seen: set[str] = set()
        bucket_members: dict[str, list[str]] = {}
        for bucket, key_fn in definitions:
            ranked = sorted(features, key=key_fn, reverse=True)
            bucket_pick = ranked[: bucket_limits[bucket]]
            bucket_members[bucket] = [item.symbol for item in bucket_pick]
            for item in bucket_pick:
                if item.symbol in seen:
                    continue
                seen.add(item.symbol)
                seed_pool.append(item)
                if len(seed_pool) >= seed_pool_cap:
                    return StockSeedMiningResult(
                        seed_pool=seed_pool,
                        bucket_members=bucket_members,
                        bucket_limits=bucket_limits,
                    )

        if len(seed_pool) < min(len(features), seed_pool_cap):
            ranked_all = sorted(features, key=lambda item: (item.quant_score, item.momentum20), reverse=True)
            for item in ranked_all:
                if item.symbol in seen:
                    continue
                seen.add(item.symbol)
                seed_pool.append(item)
                if len(seed_pool) >= seed_pool_cap:
                    break

        return StockSeedMiningResult(
            seed_pool=seed_pool,
            bucket_members=bucket_members,
            bucket_limits=bucket_limits,
        )


def _risk_rank(risk_level: str) -> int:
    if risk_level == "LOW":
        return 0
    if risk_level == "MEDIUM":
        return 1
    return 2


def _resolve_bucket_limits(base_limit: int, payload: StockSelectionPayload, market_regime: str) -> dict[str, int]:
    regime_bias = {
        "UPTREND": {"trend": 1.35, "money_flow": 1.15, "quality": 0.9, "event": 0.85, "resonance": 1.2},
        "ROTATION": {"trend": 1.0, "money_flow": 1.05, "quality": 1.0, "event": 0.95, "resonance": 1.1},
        "EVENT_DRIVEN": {"trend": 0.9, "money_flow": 0.95, "quality": 0.8, "event": 1.4, "resonance": 1.25},
        "DEFENSIVE": {"trend": 0.8, "money_flow": 0.9, "quality": 1.3, "event": 0.85, "resonance": 0.9},
        "RISK_OFF": {"trend": 0.7, "money_flow": 0.8, "quality": 1.35, "event": 0.75, "resonance": 0.8},
    }.get(market_regime, {})

    template_bias = _resolve_template_bias(payload.template_key, payload.template_name, payload.template_snapshot)
    payload_bias = {
        "trend": payload.trend_bias,
        "money_flow": payload.money_flow_bias,
        "quality": payload.quality_bias,
        "event": payload.event_bias,
        "resonance": payload.resonance_bias,
    }

    result: dict[str, int] = {}
    for bucket in ("trend", "money_flow", "quality", "event", "resonance"):
        multiplier = _clamp_multiplier(
            regime_bias.get(bucket, 1.0) * template_bias.get(bucket, 1.0) * payload_bias.get(bucket, 1.0)
        )
        result[bucket] = max(1, int(round(base_limit * multiplier)))
    return result


def _resolve_template_bias(template_key: str, template_name: str, snapshot: dict[str, object]) -> dict[str, float]:
    text = " ".join(
        item.strip().lower()
        for item in (
            template_key or "",
            template_name or "",
            str(snapshot.get("template_key", "") or ""),
            str(snapshot.get("name", "") or ""),
        )
        if item
    )
    if "trend" in text or "趋势成长" in text:
        return {"trend": 1.3, "money_flow": 1.1, "quality": 0.95, "event": 0.9, "resonance": 1.15}
    if "龙头共振" in text or "resonance" in text:
        return {"trend": 1.1, "money_flow": 1.0, "quality": 0.9, "event": 1.05, "resonance": 1.35}
    if "事件驱动" in text or "event" in text:
        return {"trend": 0.9, "money_flow": 0.95, "quality": 0.85, "event": 1.4, "resonance": 1.2}
    if "均衡稳健" in text or "balanced" in text:
        return {"trend": 1.0, "money_flow": 1.0, "quality": 1.1, "event": 0.95, "resonance": 1.0}
    if "行业轮动" in text or "rotation" in text:
        return {"trend": 1.0, "money_flow": 1.15, "quality": 1.0, "event": 0.95, "resonance": 1.15}
    return {}


def _clamp_multiplier(value: float) -> float:
    if value < 0.5:
        return 0.5
    if value > 1.5:
        return 1.5
    return value
