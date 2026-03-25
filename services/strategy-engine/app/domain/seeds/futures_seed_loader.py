from __future__ import annotations

import json
from typing import Any
from urllib import error as urllib_error
from urllib import request as urllib_request

from app.domain.models import FuturesSeed, FuturesSeedLoadResult
from app.schemas.futures import FuturesStrategyPayload
from app.settings import Settings, get_settings


DEFAULT_FUTURES_SEEDS: tuple[FuturesSeed, ...] = (
    FuturesSeed("IF2606", "沪深300股指", "2026-03-17", 3988.0, -0.32, 1.8, 1.9, 6.2, 1.28, 0.55, 0.18, 0.42, "TREND"),
    FuturesSeed("IH2606", "上证50股指", "2026-03-17", 2752.4, -0.18, 1.5, 1.5, 5.1, 1.16, 0.44, 0.12, 0.36, "TREND"),
    FuturesSeed("IC2606", "中证500股指", "2026-03-17", 5426.8, 0.22, 2.4, -1.2, 4.8, 1.22, -0.38, -0.08, -0.22, "WEAK"),
    FuturesSeed("IM2606", "中证1000股指", "2026-03-17", 5872.0, 0.35, 3.1, -1.7, 7.3, 1.34, -0.46, -0.16, -0.28, "VOLATILE"),
    FuturesSeed("AU2606", "沪金主力", "2026-03-17", 536.7, -0.12, 2.2, 1.1, 3.2, 1.08, 0.25, 0.22, 0.18, "DEFENSIVE"),
)


class FuturesSeedLoader:
    def __init__(self, settings: Settings | None = None) -> None:
        self._settings = settings or get_settings()

    def load(self, payload: FuturesStrategyPayload) -> FuturesSeedLoadResult:
        try:
            return self._load_from_go_backend(payload)
        except Exception as exc:
            if self._settings.allow_sample_futures_seeds:
                return FuturesSeedLoadResult(
                    seeds=self._filter_sample_seeds(payload),
                    meta={"source": "sample-fallback"},
                    warnings=[f"futures context api unavailable, fallback to sample seeds: {exc}"],
                )
            raise ValueError(f"futures context api unavailable: {exc}") from exc

    def _load_from_go_backend(self, payload: FuturesStrategyPayload) -> FuturesSeedLoadResult:
        base_url = self._settings.go_backend_base_url.strip().rstrip("/")
        if not base_url:
            raise ValueError("STRATEGY_ENGINE_GO_BACKEND_BASE_URL is not configured")
        response = self._post_context_request(
            f"{base_url}/internal/v1/strategy-engine/context/futures-strategy",
            {
                "trade_date": payload.trade_date,
                "contracts": payload.contracts,
                "limit": payload.limit,
                "allow_mock_fallback_on_short_history": payload.allow_mock_fallback_on_short_history,
            },
        )
        seeds_payload = response.get("seeds")
        if not isinstance(seeds_payload, list) or not seeds_payload:
            raise ValueError("go backend returned empty futures seeds")
        meta = response.get("meta") if isinstance(response.get("meta"), dict) else {}
        warnings = [str(item).strip() for item in meta.get("warnings", []) if str(item).strip()]
        return FuturesSeedLoadResult(
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

    def _filter_sample_seeds(self, payload: FuturesStrategyPayload) -> list[FuturesSeed]:
        include = {item.strip().upper() for item in payload.contracts if item.strip()}
        result = []
        for item in DEFAULT_FUTURES_SEEDS:
            if include and item.contract not in include:
                continue
            result.append(item)
        return result

    def _parse_seed(self, item: dict[str, Any]) -> FuturesSeed:
        return FuturesSeed(
            contract=str(item.get("contract", "")).strip().upper(),
            name=str(item.get("name", "")).strip(),
            trade_date=str(item.get("trade_date", "")).strip(),
            last_price=_as_float(item.get("last_price")),
            basis_pct=_as_float(item.get("basis_pct")),
            volatility14=_as_float(item.get("volatility14")),
            trend_strength=_as_float(item.get("trend_strength")),
            oi_change_pct=_as_float(item.get("oi_change_pct")),
            volume_ratio=_as_float(item.get("volume_ratio")),
            flow_bias=_as_float(item.get("flow_bias")),
            carry_pct=_as_float(item.get("carry_pct")),
            news_bias=_as_float(item.get("news_bias")),
            regime=str(item.get("regime", "")).strip().upper() or "DEFENSIVE",
            turnover_ratio=_as_float(item.get("turnover_ratio"), 1.0),
            term_structure_pct=_as_float(item.get("term_structure_pct")),
            curve_slope_pct=_as_float(item.get("curve_slope_pct")),
            inventory_level=_as_float(item.get("inventory_level")),
            inventory_change_pct=_as_float(item.get("inventory_change_pct")),
            inventory_pressure=_as_float(item.get("inventory_pressure")),
            inventory_focus_area=str(item.get("inventory_focus_area", "")).strip(),
            inventory_focus_warehouse=str(item.get("inventory_focus_warehouse", "")).strip(),
            inventory_focus_brand=str(item.get("inventory_focus_brand", "")).strip(),
            inventory_focus_place=str(item.get("inventory_focus_place", "")).strip(),
            inventory_focus_grade=str(item.get("inventory_focus_grade", "")).strip(),
            inventory_area_share=_as_float(item.get("inventory_area_share")),
            inventory_warehouse_share=_as_float(item.get("inventory_warehouse_share")),
            inventory_brand_share=_as_float(item.get("inventory_brand_share")),
            inventory_place_share=_as_float(item.get("inventory_place_share")),
            inventory_grade_share=_as_float(item.get("inventory_grade_share")),
            inventory_concentration=_as_float(item.get("inventory_concentration")),
            inventory_warehouse_shift=_as_float(item.get("inventory_warehouse_shift")),
            inventory_persistence_days=int(item.get("inventory_persistence_days") or 0),
            inventory_brand_grade_bias=_as_float(item.get("inventory_brand_grade_bias")),
            inventory_brand_grade_summary=str(item.get("inventory_brand_grade_summary", "")).strip(),
            basis_term_alignment=_as_float(item.get("basis_term_alignment")),
            cross_contract_linkage=_as_float(item.get("cross_contract_linkage")),
            structure_signal_summary=str(item.get("structure_signal_summary", "")).strip(),
            spread_pressure=_as_float(item.get("spread_pressure")),
            spread_percentile=_as_float(item.get("spread_percentile")),
            spread_pair=str(item.get("spread_pair", "")).strip().upper(),
        )


def _as_float(value: Any, default: float = 0.0) -> float:
    if value is None:
        return default
    return float(value)
