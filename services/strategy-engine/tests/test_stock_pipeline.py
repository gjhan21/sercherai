from app.domain.features.stock_feature_factory import StockFeatureFactory
from app.domain.pipelines.stock_selection_pipeline import StockSelectionPipeline
from app.domain.reports.stock_report_builder import StockReportBuilder
from app.domain.risk.portfolio_guard import PortfolioGuard
from app.domain.seeds.market_seed_loader import MarketSeedLoader
from app.domain.selectors.stock_selector import StockSelector
from app.settings import Settings


def test_stock_selection_pipeline_filters_symbols_and_builds_publish_payloads() -> None:
    pipeline = StockSelectionPipeline(
        market_seed_loader=MarketSeedLoader(
            settings=Settings(
                go_backend_base_url="",
                allow_sample_stock_seeds=True,
            )
        ),
        stock_feature_factory=StockFeatureFactory(),
        stock_selector=StockSelector(),
        portfolio_guard=PortfolioGuard(),
        stock_report_builder=StockReportBuilder(),
    )

    report, warnings = pipeline.run(
        {
            "trade_date": "2026-03-17",
            "limit": 3,
            "seed_symbols": ["600519.SH", "601318.SH", "600036.SH", "601888.SH"],
            "excluded_symbols": ["601888.SH"],
            "max_risk_level": "MEDIUM",
            "min_score": 70,
        }
    )

    assert any("fallback to sample seeds" in item for item in warnings)
    assert report.selected_count >= 1
    assert all(item.symbol in {"600519.SH", "601318.SH", "600036.SH"} for item in report.candidates)
    assert all(item.symbol != "601888.SH" for item in report.candidates)
    assert all(item.recommendation.strategy_version.startswith("stock-selection-v2/") for item in report.publish_payloads)
    assert report.context_meta["source"] == "sample-fallback"
    assert report.stage_counts["UNIVERSE"] >= report.stage_counts["SEED_POOL"] >= report.stage_counts["PORTFOLIO"]
    assert "CANDIDATE_POOL" in report.stage_durations_ms
    assert report.market_regime in {"UPTREND", "ROTATION", "EVENT_DRIVEN", "DEFENSIVE", "RISK_OFF"}
    assert len(report.stage_logs) == 6
    assert any(item.stage == "UNIVERSE" for item in report.candidate_snapshots)
    assert len(report.portfolio_entries) == report.selected_count
    assert report.evidence_records
    assert report.evaluation_summary["status"] == "PENDING"
