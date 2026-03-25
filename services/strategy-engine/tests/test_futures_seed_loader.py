from app.domain.models import FuturesSeedLoadResult
from app.domain.seeds.futures_seed_loader import FuturesSeedLoader
from app.schemas.futures import FuturesStrategyPayload
from app.settings import Settings


def test_futures_seed_loader_uses_sample_fallback_when_enabled() -> None:
    loader = FuturesSeedLoader(
        settings=Settings(
            go_backend_base_url="",
            allow_sample_futures_seeds=True,
        )
    )

    result = loader.load(
        FuturesStrategyPayload(
            trade_date="2026-03-17",
            contracts=["IF2606", "AU2606"],
            limit=5,
        )
    )

    assert isinstance(result, FuturesSeedLoadResult)
    assert [item.contract for item in result.seeds] == ["IF2606", "AU2606"]
    assert result.meta["source"] == "sample-fallback"
    assert any("fallback to sample seeds" in item for item in result.warnings)


def test_futures_seed_loader_reads_go_backend_context(monkeypatch) -> None:
    loader = FuturesSeedLoader(
        settings=Settings(
            go_backend_base_url="http://127.0.0.1:8080",
            allow_sample_futures_seeds=False,
        )
    )

    def fake_post(url: str, payload: dict):
        assert url == "http://127.0.0.1:8080/internal/v1/strategy-engine/context/futures-strategy"
        assert payload["trade_date"] == "2026-03-19"
        return {
            "seeds": [
                {
                    "contract": "IF2606",
                    "name": "沪深300股指",
                    "trade_date": "2026-03-18",
                    "last_price": 4012.8,
                    "basis_pct": -0.18,
                    "volatility14": 1.62,
                    "trend_strength": 0.92,
                    "oi_change_pct": 4.8,
                    "volume_ratio": 1.14,
                    "turnover_ratio": 1.21,
                    "flow_bias": 0.36,
                    "carry_pct": 0.24,
                    "term_structure_pct": 0.58,
                    "curve_slope_pct": 0.91,
                    "inventory_level": 1460,
                    "inventory_change_pct": -2.67,
                    "inventory_pressure": 0.13,
                    "inventory_focus_area": "华东",
                    "inventory_focus_warehouse": "中储1号",
                    "inventory_focus_brand": "央企品牌A",
                    "inventory_focus_place": "上海",
                    "inventory_focus_grade": "标准品",
                    "inventory_area_share": 1.0,
                    "inventory_warehouse_share": 0.67,
                    "inventory_brand_share": 0.67,
                    "inventory_place_share": 0.67,
                    "inventory_grade_share": 0.67,
                    "inventory_concentration": 1.0,
                    "inventory_warehouse_shift": 0.11,
                    "inventory_persistence_days": 3,
                    "inventory_brand_grade_summary": "品牌央企品牌A / 等级标准品占比 67%，已连续3日去库",
                    "basis_term_alignment": 0.86,
                    "cross_contract_linkage": -0.22,
                    "structure_signal_summary": "期限结构顺价差联动，当前合约位于受益腿",
                    "spread_pressure": 0.22,
                    "spread_percentile": 0.81,
                    "spread_pair": "IF2606/IF2609",
                    "news_bias": 0.5,
                    "regime": "TREND",
                }
            ],
            "meta": {
                "selected_trade_date": "2026-03-18",
                "price_source": "TUSHARE",
                "news_window_days": 14,
                "warnings": ["market news partially missing"],
            },
        }

    monkeypatch.setattr(loader, "_post_context_request", fake_post)

    result = loader.load(
        FuturesStrategyPayload(
            trade_date="2026-03-19",
            contracts=["IF2606"],
            limit=2,
        )
    )

    assert len(result.seeds) == 1
    assert result.seeds[0].contract == "IF2606"
    assert result.seeds[0].turnover_ratio == 1.21
    assert result.seeds[0].term_structure_pct == 0.58
    assert result.seeds[0].curve_slope_pct == 0.91
    assert result.seeds[0].inventory_level == 1460
    assert result.seeds[0].inventory_pressure == 0.13
    assert result.seeds[0].inventory_focus_area == "华东"
    assert result.seeds[0].inventory_focus_warehouse == "中储1号"
    assert result.seeds[0].inventory_focus_brand == "央企品牌A"
    assert result.seeds[0].inventory_focus_place == "上海"
    assert result.seeds[0].inventory_focus_grade == "标准品"
    assert result.seeds[0].inventory_warehouse_share == 0.67
    assert result.seeds[0].inventory_place_share == 0.67
    assert result.seeds[0].inventory_concentration == 1.0
    assert result.seeds[0].inventory_warehouse_shift == 0.11
    assert result.seeds[0].inventory_persistence_days == 3
    assert "央企品牌A" in result.seeds[0].inventory_brand_grade_summary
    assert result.seeds[0].basis_term_alignment == 0.86
    assert result.seeds[0].cross_contract_linkage == -0.22
    assert "价差联动" in result.seeds[0].structure_signal_summary
    assert result.seeds[0].spread_pressure == 0.22
    assert result.seeds[0].spread_pair == "IF2606/IF2609"
    assert result.meta["selected_trade_date"] == "2026-03-18"
    assert result.warnings == ["market news partially missing"]
