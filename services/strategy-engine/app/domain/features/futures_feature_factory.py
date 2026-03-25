from __future__ import annotations

from app.domain.models import FuturesFeature, FuturesSeed


class FuturesFeatureFactory:
    def build(self, seeds: list[FuturesSeed]) -> list[FuturesFeature]:
        features: list[FuturesFeature] = []
        for item in seeds:
            structure_alignment = _structure_alignment(item.carry_pct, item.term_structure_pct)
            curve_alignment = _structure_alignment(item.carry_pct, item.curve_slope_pct)
            inventory_structure_share = max(
                item.inventory_area_share,
                item.inventory_warehouse_share,
                item.inventory_brand_share,
                item.inventory_place_share,
                item.inventory_grade_share,
            )
            inventory_depth_score = _inventory_depth_score(item)
            inventory_depth_edge = (inventory_depth_score - 50) / 50 if inventory_depth_score else 0.0
            structure_depth_score = _structure_depth_score(item)
            structure_depth_edge = (structure_depth_score - 50) / 50 if structure_depth_score else 0.0
            trend_score = _clamp(55 + item.trend_strength * 16 - item.volatility14 * 3, 0, 100)
            flow_score = _clamp(
                52
                + item.flow_bias * 26
                + item.oi_change_pct * 1.8
                + item.volume_ratio * 6
                + (item.turnover_ratio - 1) * 10
                + item.inventory_pressure * (9 + inventory_structure_share * 6)
                + inventory_depth_edge * 8
                + structure_depth_edge * 7
                + item.spread_pressure * 12,
                0,
                100,
            )
            carry_score = _clamp(
                50
                + item.carry_pct * 16
                + item.term_structure_pct * 9
                + item.curve_slope_pct * 5
                + item.inventory_pressure * (8 + inventory_structure_share * 5)
                + inventory_depth_edge * 6
                + structure_depth_edge * 12
                + item.spread_pressure * 10
                - abs(item.basis_pct) * 12
                + structure_alignment * 5
                + curve_alignment * 4,
                0,
                100,
            )
            news_score = _clamp(50 + item.news_bias * 20, 0, 100)
            directional_edge = (
                item.trend_strength * 0.45
                + item.flow_bias * 0.35
                + item.carry_pct * 0.12
                + item.news_bias * 0.08
                + item.inventory_pressure * (0.10 + inventory_structure_share * 0.08)
                + inventory_depth_edge * 0.12
                + structure_depth_edge * 0.10
                + item.spread_pressure * 0.14
            )
            conviction_score = _clamp(
                trend_score * 0.30
                + flow_score * 0.25
                + carry_score * 0.15
                + structure_depth_score * 0.15
                + news_score * 0.15
                + inventory_depth_score * 0.10,
                0,
                100,
            )

            direction = "NEUTRAL"
            if directional_edge >= 0.18 or item.regime in {"TREND", "DEFENSIVE"}:
                direction = "LONG"
            elif directional_edge <= -0.18 or item.regime in {"WEAK", "VOLATILE"}:
                direction = "SHORT"

            risk_level = _classify_futures_risk(item.volatility14, item.volume_ratio, item.basis_pct)
            entry_price, take_profit_price, stop_loss_price = _price_levels(item.last_price, item.volatility14, direction)
            feature = FuturesFeature(
                contract=item.contract,
                name=item.name,
                trade_date=item.trade_date,
                last_price=round(item.last_price, 2),
                basis_pct=round(item.basis_pct, 2),
                volatility14=round(item.volatility14, 2),
                trend_strength=round(item.trend_strength, 2),
                oi_change_pct=round(item.oi_change_pct, 2),
                volume_ratio=round(item.volume_ratio, 2),
                flow_bias=round(item.flow_bias, 2),
                carry_pct=round(item.carry_pct, 2),
                news_bias=round(item.news_bias, 2),
                regime=item.regime,
                direction=direction,
                trend_score=round(trend_score, 2),
                flow_score=round(flow_score, 2),
                carry_score=round(carry_score, 2),
                news_score=round(news_score, 2),
                conviction_score=round(conviction_score, 2),
                risk_level=risk_level,
                entry_price=entry_price,
                take_profit_price=take_profit_price,
                stop_loss_price=stop_loss_price,
                turnover_ratio=round(item.turnover_ratio, 2),
                term_structure_pct=round(item.term_structure_pct, 2),
                curve_slope_pct=round(item.curve_slope_pct, 2),
                inventory_level=round(item.inventory_level, 2),
                inventory_change_pct=round(item.inventory_change_pct, 2),
                inventory_pressure=round(item.inventory_pressure, 2),
                inventory_focus_area=item.inventory_focus_area,
                inventory_focus_warehouse=item.inventory_focus_warehouse,
                inventory_focus_brand=item.inventory_focus_brand,
                inventory_focus_place=item.inventory_focus_place,
                inventory_focus_grade=item.inventory_focus_grade,
                inventory_area_share=round(item.inventory_area_share, 2),
                inventory_warehouse_share=round(item.inventory_warehouse_share, 2),
                inventory_brand_share=round(item.inventory_brand_share, 2),
                inventory_place_share=round(item.inventory_place_share, 2),
                inventory_grade_share=round(item.inventory_grade_share, 2),
                inventory_concentration=round(item.inventory_concentration, 2),
                inventory_warehouse_shift=round(item.inventory_warehouse_shift, 2),
                inventory_persistence_days=item.inventory_persistence_days,
                inventory_brand_grade_bias=round(item.inventory_brand_grade_bias, 2),
                inventory_brand_grade_summary=item.inventory_brand_grade_summary,
                inventory_depth_score=round(inventory_depth_score, 2),
                basis_term_alignment=round(item.basis_term_alignment, 2),
                cross_contract_linkage=round(item.cross_contract_linkage, 2),
                structure_signal_summary=item.structure_signal_summary,
                structure_depth_score=round(structure_depth_score, 2),
                spread_pressure=round(item.spread_pressure, 2),
                spread_percentile=round(item.spread_percentile, 2),
                spread_pair=item.spread_pair,
            )
            feature.inventory_factor_summary = _build_inventory_factor_summary(feature)
            feature.structure_factor_summary = _build_structure_factor_summary(feature)
            feature.reasons = _build_reasons(feature)
            feature.positive_reasons = list(feature.reasons[:4])
            feature.veto_reasons = _build_veto_reasons(feature)
            feature.evidence_cards = _build_evidence_cards(feature)
            summary_parts = []
            if feature.structure_factor_summary:
                summary_parts.append(feature.structure_factor_summary)
            if feature.inventory_factor_summary:
                summary_parts.append(feature.inventory_factor_summary)
            summary_parts.extend((feature.positive_reasons or feature.reasons)[:2])
            feature.evidence_summary = "；".join(summary_parts[:3])
            feature.risk_flags = _build_risk_flags(feature)
            feature.related_entities = _build_related_entities(feature)
            feature.reason_summary = "；".join(feature.reasons[:3])
            features.append(feature)
        return features


def _inventory_depth_score(item: FuturesSeed) -> float:
    persistence_edge = min(max(item.inventory_persistence_days, 0), 5) * (1 if item.inventory_pressure >= 0 else -1) * 4
    return _clamp(
        50
        + item.inventory_pressure * 20
        + item.inventory_concentration * 18
        + item.inventory_warehouse_shift * 60
        + item.inventory_brand_grade_bias * 14
        + persistence_edge,
        0,
        100,
    )


def _structure_depth_score(item: FuturesSeed) -> float:
    return _clamp(
        50
        + item.basis_term_alignment * 24
        - item.cross_contract_linkage * 18
        + abs(item.term_structure_pct) * 5
        + abs(item.curve_slope_pct) * 3
        - abs(item.basis_pct) * 4,
        0,
        100,
    )


def _classify_futures_risk(volatility14: float, volume_ratio: float, basis_pct: float) -> str:
    if volatility14 >= 3.5 or abs(basis_pct) >= 0.8 or volume_ratio >= 1.45:
        return "HIGH"
    if volatility14 >= 2.2 or abs(basis_pct) >= 0.35:
        return "MEDIUM"
    return "LOW"


def _price_levels(last_price: float, volatility14: float, direction: str) -> tuple[float, float, float]:
    unit = max(last_price * max(volatility14, 1.0) / 100, last_price * 0.002)
    if direction == "SHORT":
        return round(last_price + unit * 0.2, 2), round(last_price - unit * 1.6, 2), round(last_price + unit * 0.9, 2)
    if direction == "NEUTRAL":
        return round(last_price, 2), round(last_price + unit, 2), round(last_price - unit, 2)
    return round(last_price - unit * 0.2, 2), round(last_price + unit * 1.6, 2), round(last_price - unit * 0.9, 2)


def _build_reasons(item: FuturesFeature) -> list[str]:
    reasons = [f"趋势强度{item.trend_strength:.2f}，方向偏{_direction_cn(item.direction)}"]
    if item.carry_pct != 0 or item.term_structure_pct != 0:
        structure_text = "同向确认" if _structure_alignment(item.carry_pct, item.term_structure_pct) > 0 else "存在分歧"
        reasons.append(
            f"期限结构{item.carry_pct:.2f}%，近远月斜率{item.term_structure_pct:.2f}%，跨期结构{structure_text}"
        )
    if item.curve_slope_pct != 0:
        slope_text = "远月升水延续" if item.curve_slope_pct > 0 else "远月贴水加深"
        reasons.append(f"全曲线斜率{item.curve_slope_pct:.2f}%，{slope_text}")
    if item.inventory_level > 0 or item.inventory_change_pct != 0:
        inventory_text = "仓单去化偏多" if item.inventory_pressure > 0 else "仓单累库偏空"
        reasons.append(f"仓单库存{item.inventory_level:.0f}，变化{item.inventory_change_pct:.2f}%，{inventory_text}")
    inventory_structure = _inventory_structure_text(item)
    if inventory_structure:
        reasons.append(inventory_structure)
    if item.inventory_concentration >= 0.55:
        reasons.append(f"库存集中度{item.inventory_concentration:.0%}，结构博弈更聚焦")
    if abs(item.inventory_warehouse_shift) >= 0.05 and item.inventory_focus_warehouse:
        shift_text = "仓库迁移至主仓" if item.inventory_warehouse_shift > 0 else "仓库分散至其他库点"
        reasons.append(
            f"仓库迁移：{item.inventory_focus_warehouse}份额变动{item.inventory_warehouse_shift:+.0%}，{shift_text}"
        )
    if item.inventory_persistence_days >= 2:
        persistence_text = "去库" if item.inventory_pressure >= 0 else "累库"
        reasons.append(f"库存趋势已连续{item.inventory_persistence_days}日{persistence_text}")
    if item.inventory_brand_grade_summary:
        reasons.append(item.inventory_brand_grade_summary)
    if item.structure_signal_summary:
        reasons.append(f"结构联动：{item.structure_signal_summary}")
    if item.oi_change_pct > 0:
        reasons.append(f"持仓变化{item.oi_change_pct:.2f}%，增仓配合方向信号")
    if item.volume_ratio >= 1.15:
        reasons.append(f"量比{item.volume_ratio:.2f}，流动性确认较好")
    if item.turnover_ratio >= 1.1:
        reasons.append(f"成交额比{item.turnover_ratio:.2f}，资金活跃度同步放大")
    if item.spread_pair and item.spread_percentile > 0:
        if item.spread_pressure >= 0.08:
            spread_text = "当前合约处于价差回归受益腿"
        elif item.spread_pressure <= -0.08:
            spread_text = "当前合约处于价差压力腿"
        else:
            spread_text = "跨期价差信号中性"
        reasons.append(f"关联价差{item.spread_pair}分位{item.spread_percentile:.2f}，{spread_text}")
    if item.news_bias != 0:
        bias_text = "偏多" if item.news_bias > 0 else "偏空"
        reasons.append(f"资讯情绪{bias_text}，新闻偏置{item.news_bias:.2f}")
    return reasons[:14]


def _inventory_structure_text(item: FuturesFeature) -> str:
    parts: list[str] = []
    if item.inventory_focus_area and item.inventory_area_share > 0:
        parts.append(f"区域{item.inventory_focus_area}占比{item.inventory_area_share:.0%}")
    if item.inventory_focus_warehouse and item.inventory_warehouse_share > 0:
        parts.append(f"仓库{item.inventory_focus_warehouse}占比{item.inventory_warehouse_share:.0%}")
    if item.inventory_focus_brand and item.inventory_brand_share > 0:
        parts.append(f"品牌{item.inventory_focus_brand}占比{item.inventory_brand_share:.0%}")
    if item.inventory_focus_place and item.inventory_place_share > 0:
        parts.append(f"产地{item.inventory_focus_place}占比{item.inventory_place_share:.0%}")
    if item.inventory_focus_grade and item.inventory_grade_share > 0:
        parts.append(f"等级{item.inventory_focus_grade}占比{item.inventory_grade_share:.0%}")
    if not parts:
        return ""
    return "仓单结构：" + "，".join(parts[:5])


def _build_inventory_factor_summary(item: FuturesFeature) -> str:
    parts: list[str] = []
    if item.inventory_brand_grade_summary:
        parts.append(item.inventory_brand_grade_summary)
    if item.inventory_focus_warehouse and abs(item.inventory_warehouse_shift) >= 0.05:
        parts.append(f"主仓{item.inventory_focus_warehouse}份额{item.inventory_warehouse_shift:+.0%}")
    if item.inventory_concentration >= 0.55:
        parts.append(f"集中度{item.inventory_concentration:.0%}")
    if item.inventory_persistence_days >= 2:
        persistence_text = "去库" if item.inventory_pressure >= 0 else "累库"
        parts.append(f"连续{item.inventory_persistence_days}日{persistence_text}")
    return "；".join(parts[:3])


def _build_structure_factor_summary(item: FuturesFeature) -> str:
    parts: list[str] = []
    if item.structure_signal_summary:
        parts.append(item.structure_signal_summary)
    if item.basis_term_alignment:
        parts.append(f"结构对齐{item.basis_term_alignment:.2f}")
    if item.spread_pair and item.cross_contract_linkage:
        parts.append(f"{item.spread_pair}联动{item.cross_contract_linkage:.2f}")
    return "；".join(parts[:3])


def _structure_alignment(carry_pct: float, term_structure_pct: float) -> float:
    if carry_pct == 0 or term_structure_pct == 0:
        return 0.0
    if carry_pct > 0 and term_structure_pct > 0:
        return 1.0
    if carry_pct < 0 and term_structure_pct < 0:
        return 1.0
    return -1.0


def _direction_cn(direction: str) -> str:
    if direction == "LONG":
        return "多"
    if direction == "SHORT":
        return "空"
    return "中性"


def _build_veto_reasons(item: FuturesFeature) -> list[str]:
    reasons: list[str] = []
    if item.risk_level == "HIGH":
        reasons.append("风险等级偏高，仅适合轻仓跟踪")
    if item.volume_ratio < 0.95:
        reasons.append("量比不足，流动性确认偏弱")
    if abs(item.basis_pct) >= 0.8:
        reasons.append(f"基差 {item.basis_pct:.2f}% 偏大，结构波动放大")
    if item.inventory_pressure <= -0.18:
        reasons.append("库存或仓单压力偏空")
    if item.spread_pressure <= -0.12:
        reasons.append("跨期价差不利于当前方向")
    return reasons[:4]


def _build_risk_flags(item: FuturesFeature) -> list[str]:
    flags: list[str] = []
    if item.risk_level == "HIGH":
        flags.append("高波动")
    if abs(item.basis_pct) >= 0.8:
        flags.append("基差放大")
    if item.inventory_pressure <= -0.18:
        flags.append("库存压力")
    if item.spread_pressure <= -0.12:
        flags.append("价差逆风")
    return flags[:4]


def _build_evidence_cards(item: FuturesFeature) -> list[dict[str, str]]:
    cards = [
        {
            "title": "趋势",
            "value": f"{item.trend_score:.1f}",
            "note": f"趋势强度 {item.trend_strength:.2f} / 方向 {_direction_cn(item.direction)}",
        },
        {
            "title": "流动性/资金",
            "value": f"{item.flow_score:.1f}",
            "note": f"量比 {item.volume_ratio:.2f} / 持仓变化 {item.oi_change_pct:.2f}%",
        },
        {
            "title": "基差/期限结构",
            "value": f"{item.carry_score:.1f}",
            "note": f"基差 {item.basis_pct:.2f}% / 期限结构 {item.term_structure_pct:.2f}%",
        },
        {
            "title": "事件/库存",
            "value": f"{item.news_score:.1f}",
            "note": f"资讯偏置 {item.news_bias:.2f} / 库存压力 {item.inventory_pressure:.2f}",
        },
    ]
    if item.inventory_factor_summary:
        cards.append(
            {
                "title": "库存画像",
                "value": f"{item.inventory_depth_score:.1f}",
                "note": item.inventory_factor_summary,
            }
        )
    if item.structure_factor_summary:
        cards.append(
            {
                "title": "结构联动",
                "value": f"{item.structure_depth_score:.1f}",
                "note": item.structure_factor_summary,
            }
        )
    if item.spread_pair:
        cards.append(
            {
                "title": "跨期关联",
                "value": item.spread_pair,
                "note": f"价差分位 {item.spread_percentile:.2f} / 压力 {item.spread_pressure:.2f}",
            }
        )
    return cards


def _build_related_entities(item: FuturesFeature) -> list[dict[str, str]]:
    result: list[dict[str, str]] = []
    result.append({"entity_type": "Commodity", "label": item.name[:8] or item.contract, "entity_key": item.contract})
    if item.inventory_focus_area:
        result.append({"entity_type": "SupplyChainNode", "label": item.inventory_focus_area, "entity_key": item.inventory_focus_area})
    if item.inventory_focus_warehouse:
        result.append({"entity_type": "SupplyChainNode", "label": item.inventory_focus_warehouse, "entity_key": item.inventory_focus_warehouse})
    if item.spread_pair:
        result.append({"entity_type": "Index", "label": item.spread_pair, "entity_key": item.spread_pair})
    return result[:4]


def _clamp(value: float, min_value: float, max_value: float) -> float:
    if value < min_value:
        return min_value
    if value > max_value:
        return max_value
    return value
