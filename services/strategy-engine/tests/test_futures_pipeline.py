from app.domain.features.futures_feature_factory import FuturesFeatureFactory
from app.domain.models import FuturesSeed
from app.domain.pipelines.futures_strategy_pipeline import FuturesStrategyPipeline
from app.domain.reports.futures_report_builder import FuturesReportBuilder
from app.domain.risk.leverage_guard import LeverageGuard
from app.domain.seeds.futures_seed_loader import FuturesSeedLoader
from app.domain.selectors.futures_selector import FuturesSelector
from app.settings import Settings


def test_futures_strategy_pipeline_builds_publish_payloads() -> None:
    pipeline = FuturesStrategyPipeline(
        futures_seed_loader=FuturesSeedLoader(
            settings=Settings(
                go_backend_base_url="",
                allow_sample_futures_seeds=True,
            )
        ),
        futures_feature_factory=FuturesFeatureFactory(),
        futures_selector=FuturesSelector(),
        leverage_guard=LeverageGuard(),
        futures_report_builder=FuturesReportBuilder(),
    )

    report, warnings = pipeline.run(
        {
            "trade_date": "2026-03-17",
            "limit": 2,
            "contracts": ["IF2606", "IC2606", "AU2606"],
            "max_risk_level": "HIGH",
            "min_confidence": 50,
        }
    )

    assert report.selected_count == 2
    assert len(report.publish_payloads) == 2
    assert any("fallback to sample seeds" in item for item in warnings)
    assert all(item.strategy.status == "PUBLISHED" for item in report.publish_payloads)
    assert all(item.guidance.entry_range for item in report.publish_payloads)
    assert report.context_meta["source"] == "sample-fallback"


def test_futures_feature_factory_uses_term_structure_and_turnover_confirmation() -> None:
    feature = FuturesFeatureFactory().build(
        [
            FuturesSeed(
                contract="IF2606",
                name="沪深300股指",
                trade_date="2026-03-19",
                last_price=4012.8,
                basis_pct=-0.18,
                volatility14=1.62,
                trend_strength=0.92,
                oi_change_pct=4.8,
                volume_ratio=1.14,
                flow_bias=0.36,
                carry_pct=0.24,
                news_bias=0.5,
                regime="TREND",
                turnover_ratio=1.24,
                term_structure_pct=0.58,
                curve_slope_pct=0.93,
                inventory_level=1460,
                inventory_change_pct=-2.67,
                inventory_pressure=0.13,
                inventory_focus_area="华东",
                inventory_focus_warehouse="中储1号",
                inventory_focus_brand="央企品牌A",
                inventory_focus_place="上海",
                inventory_focus_grade="标准品",
                inventory_area_share=1.0,
                inventory_warehouse_share=0.67,
                inventory_brand_share=0.67,
                inventory_place_share=0.67,
                inventory_grade_share=0.67,
                spread_pressure=0.21,
                spread_percentile=0.82,
                spread_pair="IF2606/IF2609",
            )
        ]
    )[0]

    assert feature.turnover_ratio == 1.24
    assert feature.term_structure_pct == 0.58
    assert feature.curve_slope_pct == 0.93
    assert feature.inventory_pressure == 0.13
    assert feature.spread_pressure == 0.21
    assert feature.carry_score > 50
    assert any("成交额比" in reason for reason in feature.reasons)
    assert any("近远月斜率" in reason for reason in feature.reasons)
    assert any("全曲线斜率" in reason for reason in feature.reasons)
    assert any("仓单库存" in reason for reason in feature.reasons)
    assert any("仓单结构" in reason for reason in feature.reasons)
    assert any("产地上海" in reason or "等级标准品" in reason for reason in feature.reasons)
    assert any("关联价差" in reason for reason in feature.reasons)
