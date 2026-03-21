from __future__ import annotations

import json
from typing import Any
from urllib import error as urllib_error
from urllib import request as urllib_request

from app.domain.models import MarketSeed, MarketSeedLoadResult
from app.schemas.stock import StockSelectionPayload
from app.settings import Settings, get_settings


DEFAULT_MARKET_SEEDS: tuple[MarketSeed, ...] = (
    MarketSeed("600519.SH", "贵州茅台", "2026-03-17", 1728.5, 3.2, 8.6, 1.9, 1.28, 3.1, 2.4, 18234.5, 26.8, 9.6, 0.86, 4, 0.75),
    MarketSeed("601318.SH", "中国平安", "2026-03-17", 49.6, 2.8, 7.9, 2.1, 1.35, 4.0, 2.0, 10453.8, 10.5, 1.2, 1.05, 3, 0.67),
    MarketSeed("600036.SH", "招商银行", "2026-03-17", 36.2, 2.1, 6.8, 2.0, 1.22, 3.6, 1.8, 7234.2, 8.9, 0.9, 0.96, 2, 0.50),
    MarketSeed("300750.SZ", "宁德时代", "2026-03-17", 211.8, 3.9, 9.5, 2.8, 1.54, 6.4, 2.9, 21230.0, 38.2, 7.6, 1.48, 5, 0.60),
    MarketSeed("000333.SZ", "美的集团", "2026-03-17", 71.4, 1.9, 6.1, 1.7, 1.11, 3.5, 1.6, 5312.7, 13.2, 2.7, 0.82, 2, 0.50),
    MarketSeed("688981.SH", "中芯国际", "2026-03-17", 91.2, 4.1, 10.8, 3.6, 1.62, 8.4, 3.0, 15890.2, 58.6, 5.1, 2.13, 5, 0.58),
    MarketSeed("002594.SZ", "比亚迪", "2026-03-17", 246.7, 3.5, 8.2, 2.9, 1.43, 6.9, 2.5, 17340.6, 31.5, 6.4, 1.27, 4, 0.63),
    MarketSeed("601012.SH", "隆基绿能", "2026-03-17", 18.4, 1.4, 4.8, 3.4, 0.96, 9.8, 1.2, -2350.0, 42.1, 2.2, 1.38, 4, 0.42),
    MarketSeed("600276.SH", "恒瑞医药", "2026-03-17", 55.6, 2.4, 7.1, 2.3, 1.19, 5.2, 1.9, 6180.5, 34.2, 6.0, 0.92, 3, 0.57),
    MarketSeed("601888.SH", "中国中免", "2026-03-17", 79.1, 1.1, 3.6, 4.2, 0.88, 12.4, 0.9, -1250.0, 51.0, 4.4, 1.03, 4, 0.38),
)


class MarketSeedLoader:
    def __init__(self, settings: Settings | None = None) -> None:
        self._settings = settings or get_settings()

    def load(self, payload: StockSelectionPayload) -> MarketSeedLoadResult:
        try:
            return self._load_from_go_backend(payload)
        except Exception as exc:
            if self._settings.allow_sample_stock_seeds:
                return MarketSeedLoadResult(
                    seeds=self._filter_sample_seeds(payload),
                    meta={"source": "sample-fallback"},
                    warnings=[f"stock context api unavailable, fallback to sample seeds: {exc}"],
                )
            raise ValueError(f"stock context api unavailable: {exc}") from exc

    def _load_from_go_backend(self, payload: StockSelectionPayload) -> MarketSeedLoadResult:
        base_url = self._settings.go_backend_base_url.strip().rstrip("/")
        if not base_url:
            raise ValueError("STRATEGY_ENGINE_GO_BACKEND_BASE_URL is not configured")
        response = self._post_context_request(
            f"{base_url}/internal/v1/strategy-engine/context/stock-selection",
            {
                "trade_date": payload.trade_date,
                "selection_mode": payload.selection_mode,
                "universe_scope": payload.effective_universe_scope(),
                "profile_id": payload.profile_id,
                "debug_seed_symbols": payload.debug_seed_symbols,
                "seed_symbols": payload.effective_seed_symbols(),
                "excluded_symbols": payload.excluded_symbols,
                "limit": payload.limit,
                "market_scope": payload.market_scope or payload.effective_universe_scope(),
                "min_listing_days": payload.min_listing_days,
                "min_avg_turnover": payload.min_avg_turnover,
                "exclude_st": payload.exclude_st,
                "exclude_suspended": payload.exclude_suspended,
                "price_min": payload.price_min,
                "price_max": payload.price_max,
                "volatility_min": payload.volatility_min,
                "volatility_max": payload.volatility_max,
                "industry_whitelist": payload.industry_whitelist,
                "industry_blacklist": payload.industry_blacklist,
                "sector_whitelist": payload.sector_whitelist,
                "sector_blacklist": payload.sector_blacklist,
                "theme_whitelist": payload.theme_whitelist,
                "theme_blacklist": payload.theme_blacklist,
            },
        )
        seeds_payload = response.get("seeds")
        if not isinstance(seeds_payload, list) or not seeds_payload:
            raise ValueError("go backend returned empty stock seeds")
        meta = response.get("meta") if isinstance(response.get("meta"), dict) else {}
        warnings = [str(item).strip() for item in meta.get("warnings", []) if str(item).strip()]
        return MarketSeedLoadResult(
            seeds=[self._parse_seed(item) for item in seeds_payload if isinstance(item, dict)],
            meta=meta,
            warnings=warnings,
        )

    def _post_context_request(self, url: str, payload: dict[str, Any]) -> dict[str, Any]:
        body = json.dumps(payload).encode("utf-8")
        req = urllib_request.Request(
            url,
            data=body,
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        timeout_seconds = self._settings.go_backend_timeout_ms / 1000.0
        try:
            with urllib_request.urlopen(req, timeout=timeout_seconds) as resp:
                raw = resp.read().decode("utf-8")
        except urllib_error.HTTPError as exc:  # pragma: no cover - exercised via ValueError path
            detail = exc.read().decode("utf-8", errors="ignore")
            raise ValueError(f"http {exc.code}: {detail or exc.reason}") from exc
        except urllib_error.URLError as exc:  # pragma: no cover - exercised via ValueError path
            raise ValueError(str(exc.reason)) from exc
        try:
            data = json.loads(raw)
        except json.JSONDecodeError as exc:
            raise ValueError("go backend returned invalid json") from exc
        if not isinstance(data, dict):
            raise ValueError("go backend returned unexpected payload")
        return data

    def _filter_sample_seeds(self, payload: StockSelectionPayload) -> list[MarketSeed]:
        include = {item.strip().upper() for item in payload.effective_seed_symbols() if item.strip()}
        exclude = {item.strip().upper() for item in payload.excluded_symbols if item.strip()}

        result = []
        for item in DEFAULT_MARKET_SEEDS:
            if include and item.symbol not in include:
                continue
            if item.symbol in exclude:
                continue
            result.append(item)
        return result

    def _parse_seed(self, item: dict[str, Any]) -> MarketSeed:
        return MarketSeed(
            symbol=str(item.get("symbol", "")).strip().upper(),
            name=str(item.get("name", "")).strip(),
            trade_date=str(item.get("trade_date", "")).strip(),
            close_price=_as_float(item.get("close_price")),
            momentum5=_as_float(item.get("momentum5")),
            momentum20=_as_float(item.get("momentum20")),
            volatility20=_as_float(item.get("volatility20")),
            volume_ratio=_as_float(item.get("volume_ratio")),
            drawdown20=_as_float(item.get("drawdown20")),
            trend_strength=_as_float(item.get("trend_strength")),
            net_mf_amount=_as_float(item.get("net_mf_amount")),
            pe_ttm=_as_float(item.get("pe_ttm")),
            pb=_as_float(item.get("pb")),
            turnover_rate=_as_float(item.get("turnover_rate")),
            news_heat=int(item.get("news_heat", 0) or 0),
            positive_news_rate=_as_float(item.get("positive_news_rate"), default=0.5),
            listing_days=_as_int(item.get("listing_days")),
            avg_turnover20=_as_float(item.get("avg_turnover_20")),
            suspended_proxy=_as_bool(item.get("suspended_proxy")),
            st_risk_proxy=_as_bool(item.get("st_risk_proxy")),
            industry=str(item.get("industry", "")).strip(),
            sector=str(item.get("sector", "")).strip(),
            theme_tags=_as_string_list(item.get("theme_tags")),
            risk_flags=_as_string_list(item.get("risk_flags")),
        )


def _as_float(value: Any, default: float = 0.0) -> float:
    if value is None:
        return default
    return float(value)


def _as_int(value: Any, default: int = 0) -> int:
    if value is None:
        return default
    return int(value)


def _as_bool(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    if value in {None, ""}:
        return False
    if isinstance(value, (int, float)):
        return value != 0
    return str(value).strip().lower() in {"1", "true", "yes", "y"}


def _as_string_list(value: Any) -> list[str]:
    if isinstance(value, list):
        return [str(item).strip() for item in value if str(item).strip()]
    if isinstance(value, str):
        return [item.strip() for item in value.split(",") if item.strip()]
    return []
