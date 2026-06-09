import pytest
from unittest.mock import MagicMock, patch
import httpx

from app.core.llm_client import LLMClient
from app.settings import Settings
from app.domain.pipelines.stock_selection_pipeline import StockSelectionPipeline
from app.domain.features.stock_feature_factory import StockFeatureFactory
from app.domain.reports.stock_report_builder import StockReportBuilder
from app.domain.risk.portfolio_guard import PortfolioGuard
from app.domain.seeds.market_seed_loader import MarketSeedLoader
from app.domain.selectors.stock_selector import StockSelector


def test_llm_client_fallback_when_no_key() -> None:
    client = LLMClient()
    client.settings.llm_api_key = ""
    client.settings.enable_llm_review = True

    res = client.review_stock(
        symbol="600519.SH",
        name="贵州茅台",
        metrics={"quant_score": 85.0},
        local_opinions=["trend智能体认为：向上 (POSITIVE)"],
        market_regime="UPTREND",
    )

    assert res["rating"] == "观望"
    assert "本地智能体审议结论" in res["analysis"]


@patch("httpx.Client.post")
def test_llm_client_successful_response(mock_post: MagicMock) -> None:
    # 模拟正常的 HTTP 200 返回，且带 ```json 标记
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.json.return_value = {
        "choices": [
            {
                "message": {
                    "content": (
                        "```json\n"
                        "{\n"
                        '  "rating": "强烈推荐",\n'
                        '  "summary": "技术共振爆发，资金流入显著",\n'
                        '  "analysis": "大容量趋势向上，多指标共振良好。"\n'
                        "}\n"
                        "```"
                    )
                }
            }
        ]
    }
    mock_post.return_value = mock_resp

    client = LLMClient()
    client.settings.llm_api_key = "mock_key"
    client.settings.enable_llm_review = True

    res = client.review_stock(
        symbol="600519.SH",
        name="贵州茅台",
        metrics={"quant_score": 85.0},
        local_opinions=["trend智能体认为：向上"],
        market_regime="UPTREND",
    )

    assert res["rating"] == "强烈推荐"
    assert res["summary"] == "技术共振爆发，资金流入显著"
    assert res["analysis"] == "大容量趋势向上，多指标共振良好。"


@patch("httpx.Client.post")
def test_llm_client_http_timeout_fallback(mock_post: MagicMock) -> None:
    # 模拟超时异常
    mock_post.side_effect = httpx.TimeoutException("Connection timeout")

    client = LLMClient()
    client.settings.llm_api_key = "mock_key"
    client.settings.enable_llm_review = True

    res = client.review_stock(
        symbol="600519.SH",
        name="贵州茅台",
        metrics={"quant_score": 85.0},
        local_opinions=["trend：向上 (POSITIVE)"],
        market_regime="UPTREND",
    )

    # 应该正常降级，不发生崩溃
    assert res["rating"] == "观望"
    assert "本地智能体审议结论" in res["analysis"]


@patch("httpx.Client.post")
def test_stock_pipeline_with_llm_integration(mock_post: MagicMock) -> None:
    # 模拟大模型成功返回的 JSON 结构
    mock_resp = MagicMock()
    mock_resp.status_code = 200
    mock_resp.json.return_value = {
        "choices": [
            {
                "message": {
                    "content": (
                        "{\n"
                        '  "rating": "强烈推荐",\n'
                        '  "summary": "AI 强烈推荐测试标的",\n'
                        '  "analysis": "AI 终审通过，建议重点跟踪。"\n'
                        "}"
                    )
                }
            }
        ]
    }
    mock_post.return_value = mock_resp

    from app.settings import get_settings
    settings = get_settings()
    old_key = settings.llm_api_key
    old_enabled = settings.enable_llm_review
    settings.llm_api_key = "mock_key"
    settings.enable_llm_review = True

    pipeline = StockSelectionPipeline(
        market_seed_loader=MarketSeedLoader(
            settings=Settings(
                go_backend_base_url="",
                allow_sample_stock_seeds=True,
                llm_api_key="mock_key",
                enable_llm_review=True,
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
            "seed_symbols": ["600519.SH", "601318.SH"],
            "max_risk_level": "MEDIUM",
            "min_score": 70,
            "llm_config": {
                "api_key": "mock_key_from_payload",
                "base_url": "https://api.openai.com/v1",
                "model_name": "gpt-4o-mini"
            }
        }
    )

    assert pipeline._llm_client.settings.llm_api_key == "mock_key_from_payload"
    assert report.selected_count >= 1
    # 验证是否成功回填大模型数据
    try:
        assert any("AI 强烈推荐测试标的" in item["value"] for entry in report.portfolio_entries for item in entry.evidence_cards)
        assert any("AI 终审判定" in item["title"] for entry in report.portfolio_entries for item in entry.evidence_cards)
        assert "AI强烈推荐" in report.report_summary
    finally:
        settings.llm_api_key = old_key
        settings.enable_llm_review = old_enabled
