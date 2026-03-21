from __future__ import annotations

from app.domain.models import MarketSeed, StockFeature


class StockFeatureFactory:
    def build(self, seeds: list[MarketSeed]) -> list[StockFeature]:
        if not seeds:
            return []

        momentum5_values = [item.momentum5 for item in seeds]
        momentum20_values = [item.momentum20 for item in seeds]
        volatility_values = [item.volatility20 for item in seeds]
        volume_ratio_values = [item.volume_ratio for item in seeds]
        drawdown_values = [item.drawdown20 for item in seeds]
        trend_values = [item.trend_strength for item in seeds]
        pe_values = [item.pe_ttm for item in seeds if item.pe_ttm > 0]
        pb_values = [item.pb for item in seeds if item.pb > 0]
        turnover_values = [item.turnover_rate for item in seeds if item.turnover_rate > 0]
        moneyflow_values = [item.net_mf_amount for item in seeds]
        news_heat_values = [float(item.news_heat) for item in seeds if item.news_heat > 0]

        features: list[StockFeature] = []
        for item in seeds:
            m5_norm = _normalize(item.momentum5, momentum5_values)
            m20_norm = _normalize(item.momentum20, momentum20_values)
            vol_norm = _normalize(item.volatility20, volatility_values, invert=True)
            volume_norm = _normalize(item.volume_ratio, volume_ratio_values)
            drawdown_norm = _normalize(item.drawdown20, drawdown_values, invert=True)
            trend_norm = _normalize(item.trend_strength, trend_values)

            trend_score = (m5_norm * 0.25 + m20_norm * 0.30 + vol_norm * 0.15 + drawdown_norm * 0.15 + trend_norm * 0.15) * 100

            net_moneyflow_norm = _normalize(item.net_mf_amount, moneyflow_values)
            flow_score = (volume_norm * 0.42 + net_moneyflow_norm * 0.58) * 100

            pe_norm = _normalize(item.pe_ttm, pe_values, invert=True) if item.pe_ttm > 0 else 0.48
            pb_norm = _normalize(item.pb, pb_values, invert=True) if item.pb > 0 else 0.5
            turnover_norm = _normalize(item.turnover_rate, turnover_values) if item.turnover_rate > 0 else 0.45
            value_score = (pe_norm * 0.45 + pb_norm * 0.25 + turnover_norm * 0.30) * 100
            quality_score = value_score

            heat_norm = _normalize(float(item.news_heat), news_heat_values) if item.news_heat > 0 else 0.45
            positive_rate = item.positive_news_rate if item.positive_news_rate > 0 else 0.5
            news_score = (heat_norm * 0.45 + _clamp(positive_rate, 0, 1) * 0.55) * 100
            event_score = (heat_norm * 0.55 + _clamp(positive_rate, 0, 1) * 0.45) * 100
            resonance_score = (
                trend_norm * 0.25
                + volume_norm * 0.20
                + net_moneyflow_norm * 0.20
                + heat_norm * 0.15
                + _clamp(positive_rate, 0, 1) * 0.20
            ) * 100
            liquidity_risk_score = (
                turnover_norm * 0.45
                + volume_norm * 0.25
                + vol_norm * 0.15
                + drawdown_norm * 0.15
            ) * 100
            risk_adjustment_score = _build_risk_adjustment_score(item, liquidity_risk_score)

            quant_score = trend_score * 0.45 + flow_score * 0.25 + quality_score * 0.20 + news_score * 0.10
            quant_score -= _quant_risk_penalty(item)
            quant_score = _clamp(quant_score, 0, 100)

            feature = StockFeature(
                symbol=item.symbol,
                name=item.name,
                trade_date=item.trade_date,
                close_price=round(item.close_price, 3),
                momentum5=round(item.momentum5, 2),
                momentum20=round(item.momentum20, 2),
                volatility20=round(item.volatility20, 2),
                volume_ratio=round(item.volume_ratio, 2),
                drawdown20=round(item.drawdown20, 2),
                trend_strength=round(item.trend_strength, 2),
                net_mf_amount=round(item.net_mf_amount, 2),
                pe_ttm=round(item.pe_ttm, 2),
                pb=round(item.pb, 2),
                turnover_rate=round(item.turnover_rate, 2),
                news_heat=item.news_heat,
                positive_news_rate=round(item.positive_news_rate, 4),
                trend_score=round(trend_score, 2),
                flow_score=round(flow_score, 2),
                value_score=round(value_score, 2),
                quality_score=round(quality_score, 2),
                news_score=round(news_score, 2),
                event_score=round(event_score, 2),
                resonance_score=round(resonance_score, 2),
                liquidity_risk_score=round(liquidity_risk_score, 2),
                quant_score=round(quant_score, 2),
                score=round(quant_score, 2),
                risk_level=_classify_quant_risk(item.volatility20, item.drawdown20),
                listing_days=item.listing_days,
                avg_turnover20=round(item.avg_turnover20, 2),
                suspended_proxy=item.suspended_proxy,
                st_risk_proxy=item.st_risk_proxy,
                industry=item.industry,
                sector=item.sector,
                theme_tags=list(item.theme_tags),
                risk_flags=list(item.risk_flags),
                risk_adjustment_score=round(risk_adjustment_score, 2),
            )
            feature.reasons = _build_quant_reasons(feature)
            feature.positive_reasons = list(feature.reasons[:4])
            feature.veto_reasons = _build_veto_reasons(feature)
            feature.evidence_cards = _build_evidence_cards(feature)
            feature.evidence_summary = "；".join((feature.positive_reasons or feature.reasons)[:2])
            feature.reason_summary = "；".join(feature.reasons[:3])
            features.append(feature)

        return features


def _classify_quant_risk(volatility20: float, drawdown20: float) -> str:
    if volatility20 >= 4.0 or drawdown20 >= 12:
        return "HIGH"
    if volatility20 >= 2.5 or drawdown20 >= 7:
        return "MEDIUM"
    return "LOW"


def _quant_risk_penalty(item: MarketSeed) -> float:
    penalty = 0.0
    if item.volatility20 >= 5.0:
        penalty += 4.0
    if item.drawdown20 >= 16:
        penalty += 5.0
    if item.net_mf_amount < 0:
        penalty += 2.5
    if item.pe_ttm > 0 and item.pe_ttm >= 60:
        penalty += 2.0
    if item.news_heat >= 3 and item.positive_news_rate > 0 and item.positive_news_rate < 0.3:
        penalty += 2.0
    return penalty


def _build_quant_reasons(item: StockFeature) -> list[str]:
    reasons: list[str] = []
    if item.momentum20 >= 6:
        reasons.append(f"20日动量{item.momentum20:.2f}%，中期趋势较强")
    if item.volume_ratio >= 1.2:
        reasons.append(f"量比{item.volume_ratio:.2f}，成交活跃度提升")
    if item.net_mf_amount > 0:
        reasons.append(f"主力净流入{item.net_mf_amount:.2f}，资金面偏强")
    elif item.net_mf_amount < 0:
        reasons.append(f"主力净流出{abs(item.net_mf_amount):.2f}，需控制仓位")
    if item.pe_ttm > 0:
        if item.pe_ttm < 30:
            reasons.append(f"PE(TTM) {item.pe_ttm:.2f}，估值处于可接受区间")
        else:
            reasons.append(f"PE(TTM) {item.pe_ttm:.2f}，估值偏高需跟踪兑现")
    if item.news_heat > 0:
        reasons.append(f"近14天资讯热度{item.news_heat}，正面占比{item.positive_news_rate * 100:.0f}%")
    if item.listing_days > 0:
        reasons.append(f"上市约{item.listing_days}天，样本稳定性可用于近端筛选")
    if item.industry:
        reasons.append(f"行业归属 {item.industry}")
    if item.theme_tags:
        reasons.append(f"题材标签 {', '.join(item.theme_tags[:2])}")
    if item.volatility20 <= 2.5 and item.drawdown20 <= 7:
        reasons.append("波动与回撤控制在可接受范围")
    if not reasons:
        reasons.append(f"趋势评分{item.trend_score:.2f}，资金评分{item.flow_score:.2f}，估值评分{item.value_score:.2f}")
    return reasons


def _build_veto_reasons(item: StockFeature) -> list[str]:
    reasons: list[str] = list(item.risk_flags)
    if item.suspended_proxy:
        reasons.append("成交量或成交额异常，存在停牌/流动性失真代理信号")
    if item.st_risk_proxy:
        reasons.append("存在 ST 或风险警示代理信号")
    if item.volatility20 >= 4.5:
        reasons.append(f"20日波动率 {item.volatility20:.2f}% 偏高")
    if item.drawdown20 >= 10:
        reasons.append(f"20日回撤 {item.drawdown20:.2f}% 偏大")
    if item.net_mf_amount < 0:
        reasons.append("主力资金净流入转负，需要控制仓位")
    return reasons[:4]


def _build_evidence_cards(item: StockFeature) -> list[dict[str, str]]:
    cards = [
        {
            "title": "趋势",
            "value": f"{item.trend_score:.1f}",
            "note": f"20日动量 {item.momentum20:.2f}% / 趋势强度 {item.trend_strength:.2f}",
        },
        {
            "title": "资金",
            "value": f"{item.flow_score:.1f}",
            "note": f"量比 {item.volume_ratio:.2f} / 净流入 {item.net_mf_amount:.2f}",
        },
        {
            "title": "质量",
            "value": f"{item.quality_score:.1f}",
            "note": f"PE {item.pe_ttm:.2f} / PB {item.pb:.2f}",
        },
        {
            "title": "风险修正",
            "value": f"{item.risk_adjustment_score:.1f}",
            "note": f"20日均成交额 {item.avg_turnover20:.0f} / 风险级别 {item.risk_level}",
        },
    ]
    if item.theme_tags:
        cards.append(
            {
                "title": "题材/共振",
                "value": " / ".join(item.theme_tags[:2]),
                "note": f"行业 {item.industry or '-'} / 板块 {item.sector or '-'}",
            }
        )
    return cards


def _build_risk_adjustment_score(item: MarketSeed, liquidity_risk_score: float) -> float:
    score = liquidity_risk_score
    if item.listing_days > 0 and item.listing_days < 120:
        score -= 18
    elif item.listing_days > 365:
        score += 3
    if item.avg_turnover20 > 0:
        if item.avg_turnover20 < 50_000_000:
            score -= 12
        elif item.avg_turnover20 >= 200_000_000:
            score += 6
    if item.suspended_proxy:
        score -= 35
    if item.st_risk_proxy:
        score -= 30
    if item.risk_flags:
        score -= min(10, len(item.risk_flags) * 3)
    return _clamp(score, 0, 100)


def _normalize(value: float, values: list[float], invert: bool = False) -> float:
    if not values:
        return 0.5
    min_value = min(values)
    max_value = max(values)
    span = max_value - min_value
    if abs(span) < 1e-9:
        return 0.5
    normalized = (value - min_value) / span
    if invert:
        normalized = 1 - normalized
    return _clamp(normalized, 0, 1)


def _clamp(value: float, min_value: float, max_value: float) -> float:
    if value < min_value:
        return min_value
    if value > max_value:
        return max_value
    return value
