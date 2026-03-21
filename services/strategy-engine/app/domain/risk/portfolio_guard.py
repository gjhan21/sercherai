from __future__ import annotations

from dataclasses import dataclass

from app.domain.models import StockFeature
from app.schemas.stock import StockSelectionPayload


RISK_ORDER = {"LOW": 0, "MEDIUM": 1, "HIGH": 2}


@dataclass(slots=True)
class PortfolioGuardResult:
    portfolio: list[StockFeature]
    watchlist: list[StockFeature]
    warnings: list[str]


class PortfolioGuard:
    def apply(self, features: list[StockFeature], payload: StockSelectionPayload) -> PortfolioGuardResult:
        warnings: list[str] = []
        allowed = [item for item in features if _passes_quant_risk_gate(item)]
        if len(allowed) != len(features):
            warnings.append("已剔除未通过基础风险门槛的股票样本。")

        max_risk = RISK_ORDER[payload.max_risk_level]
        allowed = [item for item in allowed if RISK_ORDER.get(item.risk_level, 1) <= max_risk]
        if not allowed:
            warnings.append("按风险等级过滤后无结果，已回退到原始排序候选。")
            allowed = list(features)

        min_score_filtered = [item for item in allowed if item.score >= payload.min_score]
        if len(min_score_filtered) >= min(payload.limit, len(allowed)):
            allowed = min_score_filtered
        elif min_score_filtered:
            warnings.append("满足最小分数的样本不足目标组合数量，已按原始排序补足候选。")
            allowed = _backfill(min_score_filtered, allowed, payload.limit)
        else:
            warnings.append("按最小分数过滤后无结果，已回退到风险过滤结果。")

        diversified = _select_diversified(
            allowed,
            payload.limit,
            max_symbol_per_bucket=payload.max_symbol_per_bucket,
            max_symbols_per_sector=payload.max_symbols_per_sector,
        )
        if len(diversified) < min(payload.limit, len(allowed)):
            warnings.append("分散化约束后数量不足，已补足高分候选。")
            diversified = _backfill(diversified, allowed, payload.limit)

        portfolio = diversified[: payload.limit]
        watchlist = _build_watchlist(allowed, portfolio, payload.watchlist_limit)
        _assign_portfolio_roles(portfolio, watchlist)
        return PortfolioGuardResult(portfolio=portfolio, watchlist=watchlist, warnings=warnings)


def _passes_quant_risk_gate(item: StockFeature) -> bool:
    if item.close_price <= 0:
        return False
    if item.volatility20 >= 9.0:
        return False
    if item.drawdown20 >= 25:
        return False
    if item.volume_ratio <= 0.15:
        return False
    if item.momentum20 <= -12:
        return False
    if item.pe_ttm > 0 and item.pe_ttm >= 200:
        return False
    if item.pb > 0 and item.pb >= 30:
        return False
    if item.news_heat >= 5 and item.positive_news_rate > 0 and item.positive_news_rate < 0.2:
        return False
    return True


def _bucket(symbol: str) -> str:
    if symbol.startswith("68") and symbol.endswith(".SH"):
        return "KCB"
    if symbol.startswith("30") and symbol.endswith(".SZ"):
        return "CYB"
    if symbol.startswith("60") and symbol.endswith(".SH"):
        return "MAIN_SH"
    if (symbol.startswith("00") or symbol.startswith("002")) and symbol.endswith(".SZ"):
        return "MAIN_SZ"
    return "OTHER"


def _select_diversified(
    features: list[StockFeature],
    limit: int,
    *,
    max_symbol_per_bucket: int,
    max_symbols_per_sector: int,
) -> list[StockFeature]:
    if len(features) <= limit:
        return list(features)
    cap_by_bucket = max(1, max_symbol_per_bucket)
    cap_by_sector = max(1, max_symbols_per_sector)
    selected: list[StockFeature] = []
    overflow: list[StockFeature] = []
    bucket_count: dict[str, int] = {}
    sector_count: dict[str, int] = {}
    theme_count: dict[str, int] = {}
    for item in features:
        bucket = _bucket(item.symbol)
        sector = (item.sector or item.industry or "").strip().upper() or "UNKNOWN"
        primary_theme = (item.theme_tags[0] if item.theme_tags else "").strip().upper() or "UNTHEMED"
        if (
            bucket_count.get(bucket, 0) < cap_by_bucket
            and sector_count.get(sector, 0) < cap_by_sector
            and theme_count.get(primary_theme, 0) < cap_by_sector
        ):
            selected.append(item)
            bucket_count[bucket] = bucket_count.get(bucket, 0) + 1
            sector_count[sector] = sector_count.get(sector, 0) + 1
            theme_count[primary_theme] = theme_count.get(primary_theme, 0) + 1
        else:
            overflow.append(item)
        if len(selected) >= limit:
            return selected
    return selected + overflow


def _backfill(current: list[StockFeature], ordered: list[StockFeature], limit: int) -> list[StockFeature]:
    seen = {item.symbol for item in current}
    result = list(current)
    for item in ordered:
        if item.symbol in seen:
            continue
        result.append(item)
        seen.add(item.symbol)
        if len(result) >= limit:
            break
    return result


def _build_watchlist(ordered: list[StockFeature], portfolio: list[StockFeature], limit: int) -> list[StockFeature]:
    if limit <= 0:
        return []
    selected_symbols = {item.symbol for item in portfolio}
    watchlist: list[StockFeature] = []
    for item in ordered:
        if item.symbol in selected_symbols:
            continue
        watchlist.append(item)
        if len(watchlist) >= limit:
            break
    return watchlist


def _assign_portfolio_roles(portfolio: list[StockFeature], watchlist: list[StockFeature]) -> None:
    core_count = max(1, min(2, len(portfolio)))
    for index, item in enumerate(portfolio):
        item.portfolio_role = "CORE" if index < core_count else "SATELLITE"
        item.watchlist = False
    for item in watchlist:
        item.portfolio_role = "WATCHLIST"
        item.watchlist = True
