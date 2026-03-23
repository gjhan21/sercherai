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
    assert len(report.stage_logs) == 10
    assert any(item.stage_key == "GRAPH_ENRICHMENT" for item in report.stage_logs)
    assert any(item.stage_key == "MEMORY_FEEDBACK" for item in report.stage_logs)
    assert any(item.stage == "UNIVERSE" for item in report.candidate_snapshots)
    assert len(report.portfolio_entries) == report.selected_count
    assert report.evidence_records
    assert report.evaluation_summary["status"] == "PENDING"
    assert report.graph_summary
    assert "市场状态" in report.graph_summary
    assert all(regime not in report.graph_summary for regime in ["UPTREND", "ROTATION", "EVENT_DRIVEN", "DEFENSIVE", "RISK_OFF"])
    assert report.related_entities
    assert any("市场状态" in (item.label or "") for item in report.related_entities)
    assert report.graph_entities
    assert report.graph_relations
    assert report.memory_feedback.summary
    assert "市场状态" in report.memory_feedback.summary
    assert all(regime not in report.memory_feedback.summary for regime in ["UPTREND", "ROTATION", "EVENT_DRIVEN", "DEFENSIVE", "RISK_OFF"])
    assert any("图谱增强完成" == item.detail_message for item in report.stage_logs)
    assert any("已登记日终异步评估补写" == item.detail_message for item in report.stage_logs)
    assert report.context_meta["run_id"].startswith("stock-")
