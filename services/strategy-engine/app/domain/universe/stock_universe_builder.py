from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any

from app.domain.models import MarketSeed, MarketSeedLoadResult
from app.schemas.stock import StockSelectionPayload


@dataclass(slots=True)
class StockUniverseBuildResult:
    seeds: list[MarketSeed]
    meta: dict[str, Any] = field(default_factory=dict)
    warnings: list[str] = field(default_factory=list)


class StockUniverseBuilder:
    def build(self, payload: StockSelectionPayload, seed_result: MarketSeedLoadResult) -> StockUniverseBuildResult:
        excluded = {item.strip().upper() for item in payload.excluded_symbols if item.strip()}
        industry_whitelist = _normalize_text_set(payload.industry_whitelist)
        industry_blacklist = _normalize_text_set(payload.industry_blacklist)
        sector_whitelist = _normalize_text_set(payload.sector_whitelist)
        sector_blacklist = _normalize_text_set(payload.sector_blacklist)
        theme_whitelist = _normalize_text_set(payload.theme_whitelist)
        theme_blacklist = _normalize_text_set(payload.theme_blacklist)
        listing_days_filter_applied = _listing_days_filter_applied(seed_result.meta)
        deduped: list[MarketSeed] = []
        seen: set[str] = set()
        filter_counters: dict[str, int] = {
            "excluded_symbols": 0,
            "duplicates": 0,
            "listing_days": 0,
            "avg_turnover20": 0,
            "suspended": 0,
            "st": 0,
            "price_range": 0,
            "volatility_range": 0,
            "industry": 0,
            "sector": 0,
            "theme": 0,
        }
        duplicate_count = 0
        for item in seed_result.seeds:
            symbol = item.symbol.strip().upper()
            if not symbol:
                continue
            if symbol in excluded:
                filter_counters["excluded_symbols"] += 1
                continue
            if symbol in seen:
                duplicate_count += 1
                continue
            blocked, reason = _should_filter_seed(
                item,
                payload,
                listing_days_filter_applied,
                industry_whitelist,
                industry_blacklist,
                sector_whitelist,
                sector_blacklist,
                theme_whitelist,
                theme_blacklist,
            )
            if blocked:
                filter_counters[reason] = filter_counters.get(reason, 0) + 1
                continue
            seen.add(symbol)
            deduped.append(item)
        filter_counters["duplicates"] = duplicate_count

        warnings = list(seed_result.warnings)
        if filter_counters["excluded_symbols"] > 0:
            warnings.append(f"universe builder excluded {filter_counters['excluded_symbols']} explicit symbols")
        if filter_counters["duplicates"] > 0:
            warnings.append(f"universe builder removed {filter_counters['duplicates']} duplicated symbols")
        for key, label in (
            ("listing_days", "上市天数"),
            ("avg_turnover20", "20日均成交额"),
            ("suspended", "停牌/流动性代理"),
            ("st", "ST/风险警示"),
            ("price_range", "价格区间"),
            ("volatility_range", "波动率区间"),
            ("industry", "行业筛选"),
            ("sector", "板块筛选"),
            ("theme", "题材筛选"),
        ):
            count = filter_counters.get(key, 0)
            if count > 0:
                warnings.append(f"universe builder 因 {label} 过滤 {count} 只股票")

        meta = dict(seed_result.meta)
        meta.setdefault("selection_mode", payload.selection_mode)
        meta.setdefault("universe_scope", payload.effective_universe_scope())
        meta.setdefault("profile_id", payload.profile_id)
        meta.setdefault("template_key", payload.template_key)
        meta["universe_count"] = len(deduped)
        meta["universe_filters"] = {
            "price_min": payload.price_min,
            "price_max": payload.price_max,
            "volatility_min": payload.volatility_min,
            "volatility_max": payload.volatility_max,
            "min_listing_days": payload.min_listing_days,
            "min_avg_turnover": payload.min_avg_turnover,
        }
        return StockUniverseBuildResult(
            seeds=deduped,
            meta=meta,
            warnings=warnings,
        )


def _should_filter_seed(
    item: MarketSeed,
    payload: StockSelectionPayload,
    listing_days_filter_applied: bool,
    industry_whitelist: set[str],
    industry_blacklist: set[str],
    sector_whitelist: set[str],
    sector_blacklist: set[str],
    theme_whitelist: set[str],
    theme_blacklist: set[str],
) -> tuple[bool, str]:
    if (
        listing_days_filter_applied
        and payload.min_listing_days > 0
        and item.listing_days > 0
        and item.listing_days < payload.min_listing_days
    ):
        return True, "listing_days"
    if payload.min_avg_turnover > 0 and item.avg_turnover20 > 0 and item.avg_turnover20 < payload.min_avg_turnover:
        return True, "avg_turnover20"
    if payload.exclude_suspended and item.suspended_proxy:
        return True, "suspended"
    if payload.exclude_st and item.st_risk_proxy:
        return True, "st"
    if payload.price_min > 0 and item.close_price < payload.price_min:
        return True, "price_range"
    if payload.price_max > 0 and item.close_price > payload.price_max:
        return True, "price_range"
    if payload.volatility_min > 0 and item.volatility20 < payload.volatility_min:
        return True, "volatility_range"
    if payload.volatility_max > 0 and item.volatility20 > payload.volatility_max:
        return True, "volatility_range"

    industry = item.industry.strip().upper()
    sector = item.sector.strip().upper()
    themes = {tag.strip().upper() for tag in item.theme_tags if tag.strip()}
    if industry_whitelist and industry not in industry_whitelist:
        return True, "industry"
    if industry_blacklist and industry in industry_blacklist:
        return True, "industry"
    if sector_whitelist and sector not in sector_whitelist:
        return True, "sector"
    if sector_blacklist and sector in sector_blacklist:
        return True, "sector"
    if theme_whitelist and not (themes & theme_whitelist):
        return True, "theme"
    if theme_blacklist and (themes & theme_blacklist):
        return True, "theme"
    return False, ""


def _normalize_text_set(items: list[str]) -> set[str]:
    return {str(item).strip().upper() for item in items if str(item).strip()}


def _listing_days_filter_applied(meta: dict[str, Any]) -> bool:
    raw = meta.get("listing_days_filter_applied")
    if isinstance(raw, bool):
        return raw
    if raw not in {None, ""}:
        return str(raw).strip().lower() in {"1", "true", "yes", "y"}
    warnings = meta.get("warnings")
    if isinstance(warnings, list):
        for item in warnings:
            text = str(item).strip().lower()
            if "skipped min_listing_days proxy" in text:
                return False
    return True
