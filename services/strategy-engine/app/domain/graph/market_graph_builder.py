from __future__ import annotations

from collections import Counter
from typing import Any

from app.schemas.research import ResearchGraphEntity, ResearchGraphRelation, ResearchGraphSnapshot


class MarketGraphBuilder:
    def build_stock(
        self,
        features: list,
        *,
        run_id: str,
        trade_date: str,
        market_regime: str,
    ) -> ResearchGraphSnapshot:
        entity_map: dict[tuple[str, str], ResearchGraphEntity] = {}
        relation_map: dict[tuple[str, str, str, str, str], ResearchGraphRelation] = {}
        theme_counter: Counter[str] = Counter()
        industry_counter: Counter[str] = Counter()

        regime_entity = ResearchGraphEntity(
            entity_type="Policy",
            entity_key=f"REGIME:{market_regime or 'BASE'}",
            label=f"{_stock_market_regime_label(market_regime)} 市场状态",
            asset_domain="cross",
            tags=["market_regime"],
            meta={"market_regime": market_regime or "BASE"},
        )
        _append_entity(entity_map, regime_entity)

        for item in features:
            stock = ResearchGraphEntity(
                entity_type="Stock",
                entity_key=item.symbol,
                label=item.name,
                asset_domain="stock",
                tags=[value for value in (item.portfolio_role, item.risk_level) if value],
                meta={
                    "score": round(float(getattr(item, "score", 0.0)), 2),
                    "quant_score": round(float(getattr(item, "quant_score", 0.0)), 2),
                    "reason_summary": getattr(item, "reason_summary", ""),
                    "risk_summary": _stock_risk_summary(item),
                },
            )
            _append_entity(entity_map, stock)
            _append_relation(
                relation_map,
                ResearchGraphRelation(
                    relation_type="CONFIRMS_SIGNAL",
                    source_type="Policy",
                    source_key=regime_entity.entity_key,
                    target_type="Stock",
                    target_key=item.symbol,
                    strength=_score_strength(getattr(item, "score", 0.0)),
                    note=f"{_stock_market_regime_label(market_regime)}下的研究确认",
                ),
            )

            if getattr(item, "industry", ""):
                industry_key = str(item.industry).strip().upper()
                industry_counter[item.industry] += 1
                _append_entity(
                    entity_map,
                    ResearchGraphEntity(
                        entity_type="Industry",
                        entity_key=industry_key,
                        label=item.industry,
                        asset_domain="stock",
                    ),
                )
                _append_relation(
                    relation_map,
                    ResearchGraphRelation(
                        relation_type="BELONGS_TO",
                        source_type="Stock",
                        source_key=item.symbol,
                        target_type="Industry",
                        target_key=industry_key,
                        strength=0.9,
                    ),
                )

            if getattr(item, "sector", ""):
                sector_key = f"SECTOR:{str(item.sector).strip().upper()}"
                _append_entity(
                    entity_map,
                    ResearchGraphEntity(
                        entity_type="ConceptTheme",
                        entity_key=sector_key,
                        label=item.sector,
                        asset_domain="stock",
                        tags=["sector"],
                    ),
                )
                _append_relation(
                    relation_map,
                    ResearchGraphRelation(
                        relation_type="BENEFITS_FROM",
                        source_type="Stock",
                        source_key=item.symbol,
                        target_type="ConceptTheme",
                        target_key=sector_key,
                        strength=0.72,
                        note="板块景气与主题共振",
                    ),
                )

            for tag in getattr(item, "theme_tags", [])[:3]:
                normalized_tag = str(tag).strip()
                if not normalized_tag:
                    continue
                theme_counter[normalized_tag] += 1
                theme_key = f"THEME:{normalized_tag.upper()}"
                _append_entity(
                    entity_map,
                    ResearchGraphEntity(
                        entity_type="ConceptTheme",
                        entity_key=theme_key,
                        label=normalized_tag,
                        asset_domain="stock",
                        tags=["theme"],
                    ),
                )
                _append_relation(
                    relation_map,
                    ResearchGraphRelation(
                        relation_type="BENEFITS_FROM",
                        source_type="Stock",
                        source_key=item.symbol,
                        target_type="ConceptTheme",
                        target_key=theme_key,
                        strength=0.78,
                        note="题材共振增强候选得分",
                    ),
                )

            if getattr(item, "news_heat", 0) >= 3:
                event_key = f"EVENT:{item.symbol}:{trade_date}"
                _append_entity(
                    entity_map,
                    ResearchGraphEntity(
                        entity_type="Event",
                        entity_key=event_key,
                        label=f"{item.name} 资讯催化",
                        asset_domain="cross",
                        tags=["news"],
                        meta={"news_heat": getattr(item, "news_heat", 0)},
                    ),
                )
                _append_relation(
                    relation_map,
                    ResearchGraphRelation(
                        relation_type="AFFECTED_BY",
                        source_type="Stock",
                        source_key=item.symbol,
                        target_type="Event",
                        target_key=event_key,
                        strength=min(1.0, 0.5 + getattr(item, "news_heat", 0) * 0.1),
                        note="资讯热度达到图谱催化阈值",
                    ),
                )

        entities = list(entity_map.values())
        relations = list(relation_map.values())
        related_entities = _pick_related_entities(entities, theme_counter, industry_counter)
        summary = (
            f"股票图谱快照：{len(features)} 只研究标的，"
            f"{len([item for item in entities if item.entity_type == 'Industry'])} 个行业，"
            f"{len([item for item in entities if item.entity_type == 'ConceptTheme'])} 个题材/板块，"
            f"市场状态 {_stock_market_regime_label(market_regime)}。"
        )
        return ResearchGraphSnapshot(
            run_id=run_id,
            asset_domain="stock",
            trade_date=trade_date,
            summary=summary,
            related_entities=related_entities,
            entities=entities,
            relations=relations,
            meta={
                "market_regime": market_regime or "BASE",
                "top_themes": [name for name, _ in theme_counter.most_common(5)],
                "top_industries": [name for name, _ in industry_counter.most_common(5)],
                "asset_count": len(features),
            },
        )

    def build_futures(
        self,
        features: list,
        *,
        run_id: str,
        trade_date: str,
        market_regime: str,
    ) -> ResearchGraphSnapshot:
        entity_map: dict[tuple[str, str], ResearchGraphEntity] = {}
        relation_map: dict[tuple[str, str, str, str, str], ResearchGraphRelation] = {}
        commodity_counter: Counter[str] = Counter()

        regime_entity = ResearchGraphEntity(
            entity_type="Policy",
            entity_key=f"FUTURES_REGIME:{market_regime or 'BASE'}",
            label=f"{_futures_market_regime_label(market_regime)} 期货状态",
            asset_domain="cross",
            tags=["market_regime"],
            meta={"market_regime": market_regime or "BASE"},
        )
        _append_entity(entity_map, regime_entity)

        long_count = 0
        short_count = 0
        for item in features:
            if getattr(item, "direction", "") == "LONG":
                long_count += 1
            elif getattr(item, "direction", "") == "SHORT":
                short_count += 1

            contract = ResearchGraphEntity(
                entity_type="FuturesContract",
                entity_key=item.contract,
                label=item.name,
                asset_domain="futures",
                tags=[value for value in (item.direction, item.risk_level, item.portfolio_role) if value],
                meta={
                    "conviction_score": round(float(getattr(item, "conviction_score", 0.0)), 2),
                    "reason_summary": getattr(item, "reason_summary", ""),
                    "risk_summary": _futures_risk_summary(item),
                },
            )
            _append_entity(entity_map, contract)
            _append_relation(
                relation_map,
                ResearchGraphRelation(
                    relation_type="CONFIRMS_SIGNAL",
                    source_type="Policy",
                    source_key=regime_entity.entity_key,
                    target_type="FuturesContract",
                    target_key=item.contract,
                    strength=_score_strength(getattr(item, "conviction_score", 0.0)),
                    note=f"{_futures_market_regime_label(market_regime)}下的方向确认",
                ),
            )

            commodity_key = _commodity_key(item.contract)
            commodity_label = _commodity_label(item.name, commodity_key)
            commodity_counter[commodity_label] += 1
            _append_entity(
                entity_map,
                ResearchGraphEntity(
                    entity_type="Commodity",
                    entity_key=commodity_key,
                    label=commodity_label,
                    asset_domain="cross",
                ),
            )
            _append_relation(
                relation_map,
                ResearchGraphRelation(
                    relation_type="BELONGS_TO",
                    source_type="FuturesContract",
                    source_key=item.contract,
                    target_type="Commodity",
                    target_key=commodity_key,
                    strength=0.92,
                ),
            )

            if getattr(item, "inventory_focus_area", ""):
                event_key = f"SUPPLY:{commodity_key}:{trade_date}"
                _append_entity(
                    entity_map,
                    ResearchGraphEntity(
                        entity_type="SupplyChainNode",
                        entity_key=event_key,
                        label=f"{commodity_label} 库存/仓单结构",
                        asset_domain="cross",
                        tags=["inventory"],
                        meta={
                            "focus_area": getattr(item, "inventory_focus_area", ""),
                            "inventory_pressure": getattr(item, "inventory_pressure", 0),
                        },
                    ),
                )
                _append_relation(
                    relation_map,
                    ResearchGraphRelation(
                        relation_type="SUPPLIES_TO",
                        source_type="SupplyChainNode",
                        source_key=event_key,
                        target_type="FuturesContract",
                        target_key=item.contract,
                        strength=min(1.0, 0.55 + abs(getattr(item, "inventory_pressure", 0.0))),
                        note="库存与仓单结构影响合约方向",
                    ),
                )

            if getattr(item, "spread_pair", ""):
                pair_key = f"SPREAD:{str(item.spread_pair).strip().upper()}"
                _append_entity(
                    entity_map,
                    ResearchGraphEntity(
                        entity_type="Index",
                        entity_key=pair_key,
                        label=str(item.spread_pair),
                        asset_domain="cross",
                        tags=["spread"],
                    ),
                )
                _append_relation(
                    relation_map,
                    ResearchGraphRelation(
                        relation_type="CORRELATED_WITH",
                        source_type="FuturesContract",
                        source_key=item.contract,
                        target_type="Index",
                        target_key=pair_key,
                        strength=min(1.0, 0.4 + abs(getattr(item, "spread_pressure", 0.0))),
                        note="跨期价差关联",
                    ),
                )

        entities = list(entity_map.values())
        relations = list(relation_map.values())
        related_entities = _pick_related_entities(entities, commodity_counter, Counter())
        summary = (
            f"期货图谱快照：多头 {long_count}，空头 {short_count}，"
            f"{len([item for item in entities if item.entity_type == 'Commodity'])} 个商品锚点，"
            f"市场状态 {_futures_market_regime_label(market_regime)}。"
        )
        return ResearchGraphSnapshot(
            run_id=run_id,
            asset_domain="futures",
            trade_date=trade_date,
            summary=summary,
            related_entities=related_entities,
            entities=entities,
            relations=relations,
            meta={
                "market_regime": market_regime or "BASE",
                "top_commodities": [name for name, _ in commodity_counter.most_common(5)],
                "asset_count": len(features),
                "long_count": long_count,
                "short_count": short_count,
            },
        )


def _append_entity(store: dict[tuple[str, str], ResearchGraphEntity], entity: ResearchGraphEntity) -> None:
    store[(entity.entity_type, entity.entity_key)] = entity


def _append_relation(
    store: dict[tuple[str, str, str, str, str], ResearchGraphRelation],
    relation: ResearchGraphRelation,
) -> None:
    key = (
        relation.relation_type,
        relation.source_type,
        relation.source_key,
        relation.target_type,
        relation.target_key,
    )
    store[key] = relation


def _pick_related_entities(
    entities: list[ResearchGraphEntity],
    primary_counter: Counter[str],
    secondary_counter: Counter[str],
) -> list[ResearchGraphEntity]:
    priority_labels = {name for name, _ in primary_counter.most_common(8)}
    priority_labels.update(name for name, _ in secondary_counter.most_common(4))
    selected: list[ResearchGraphEntity] = []
    for item in entities:
        if item.entity_type in {"Stock", "FuturesContract"}:
            continue
        if item.label in priority_labels or len(selected) < 12:
            selected.append(item)
        if len(selected) >= 12:
            break
    return selected


def _score_strength(value: Any) -> float:
    try:
        score = float(value)
    except (TypeError, ValueError):
        return 0.5
    return round(max(0.1, min(score / 100, 1.0)), 2)


def _stock_risk_summary(item: Any) -> str:
    parts = [f"风险级别 {getattr(item, 'risk_level', '')}".strip()]
    volatility = getattr(item, "volatility20", 0.0)
    drawdown = getattr(item, "drawdown20", 0.0)
    parts.append(f"波动率 {volatility:.2f}%")
    parts.append(f"回撤 {drawdown:.2f}%")
    return "；".join([part for part in parts if part])


def _futures_risk_summary(item: Any) -> str:
    parts = [f"风险级别 {getattr(item, 'risk_level', '')}".strip()]
    volatility = getattr(item, "volatility14", 0.0)
    basis_pct = getattr(item, "basis_pct", 0.0)
    parts.append(f"波动率 {volatility:.2f}%")
    parts.append(f"基差 {basis_pct:.2f}%")
    return "；".join([part for part in parts if part])


def _commodity_key(contract: str) -> str:
    prefix = []
    for char in contract:
        if char.isalpha():
            prefix.append(char)
        else:
            break
    text = "".join(prefix).upper()
    return text or contract.upper()


def _commodity_label(name: str, commodity_key: str) -> str:
    name = str(name).strip()
    if not name:
        return commodity_key
    return name[:8]


def _stock_market_regime_label(value: str) -> str:
    mapping = {
        "UPTREND": "上升趋势",
        "ROTATION": "轮动切换",
        "EVENT_DRIVEN": "事件驱动",
        "DEFENSIVE": "防御修复",
        "RISK_OFF": "风险回避",
        "BASE": "基准状态",
    }
    key = str(value or "").strip().upper()
    return mapping.get(key, key or "基准状态")


def _futures_market_regime_label(value: str) -> str:
    mapping = {
        "BASE": "基准状态",
        "TREND_CONTINUE": "趋势延续",
        "POLICY_POSITIVE": "政策利多",
        "POLICY_NEGATIVE": "政策利空",
        "SUPPLY_SHOCK": "供给冲击",
        "LIQUIDITY_SHOCK": "流动性冲击",
    }
    key = str(value or "").strip().upper()
    return mapping.get(key, key or "基准状态")
