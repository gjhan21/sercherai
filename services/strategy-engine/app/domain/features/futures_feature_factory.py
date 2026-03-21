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
            trend_score = _clamp(55 + item.trend_strength * 16 - item.volatility14 * 3, 0, 100)
            flow_score = _clamp(
                52
                + item.flow_bias * 26
                + item.oi_change_pct * 1.8
                + item.volume_ratio * 6
                + (item.turnover_ratio - 1) * 10
                + item.inventory_pressure * (9 + inventory_structure_share * 6)
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
                + item.spread_pressure * 0.14
            )
            conviction_score = _clamp(trend_score * 0.35 + flow_score * 0.30 + carry_score * 0.15 + news_score * 0.20, 0, 100)

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
                spread_pressure=round(item.spread_pressure, 2),
                spread_percentile=round(item.spread_percentile, 2),
                spread_pair=item.spread_pair,
            )
            feature.reasons = _build_reasons(feature)
            feature.reason_summary = "；".join(feature.reasons[:3])
            features.append(feature)
        return features


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
    return reasons[:8]


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


def _clamp(value: float, min_value: float, max_value: float) -> float:
    if value < min_value:
        return min_value
    if value > max_value:
        return max_value
    return value
