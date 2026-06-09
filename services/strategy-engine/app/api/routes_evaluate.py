from __future__ import annotations

import logging
import copy
from datetime import datetime, timedelta
from typing import Any, Dict, List, Optional
from pydantic import BaseModel, Field
from fastapi import APIRouter, HTTPException, status

from app.domain.seeds.market_seed_loader import MarketSeedLoader, DEFAULT_MARKET_SEEDS
from app.domain.features.stock_feature_factory import StockFeatureFactory
from app.domain.scenarios.stock_scenario_engine import StockScenarioEngine
from app.domain.agents.agent_panel import AgentPanel
from app.core.llm_client import LLMClient
from app.settings import get_settings
from app.domain.models import MarketSeed
from app.domain.pipelines.stock_selection_pipeline import _detect_market_regime

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/internal/v1/strategy", tags=["evaluate"])


class EvaluateStockRequest(BaseModel):
    symbol: str = Field(..., description="Stock symbol (e.g. 300750.SZ)")
    trade_date: Optional[str] = Field(default=None, description="Trade date YYYY-MM-DD")
    llm_config: Optional[Dict[str, str]] = Field(default=None, description="LLM settings overrides")


def _risk_controls(risk_level: str) -> tuple[str, str, str]:
    if risk_level == "LOW":
        return "上涨8%-12%分批止盈", "回撤3%止损", "10%-15%"
    if risk_level == "HIGH":
        return "上涨12%-18%动态止盈", "回撤7%止损", "5%-8%"
    return "上涨10%-15%分批止盈", "回撤5%止损", "8%-12%"


def _clamp(value: float, min_value: float, max_value: float) -> float:
    return max(min_value, min(max_value, value))


@router.post("/evaluate-stock", status_code=status.HTTP_200_OK)
def evaluate_stock(req: EvaluateStockRequest) -> Dict[str, Any]:
    symbol = req.symbol.strip().upper()
    if not symbol:
        raise HTTPException(status_code=400, detail="symbol is required")

    trade_date = req.trade_date
    if not trade_date:
        trade_date = datetime.now().strftime("%Y-%m-%d")

    settings = get_settings()
    loader = MarketSeedLoader(settings=settings)

    # 1. 加载个股历史种子数据 (用于时间序列归一化与模型训练预测)
    try:
        seeds = loader.load_history(symbol=symbol, trade_date=trade_date, limit=120)
    except Exception as exc:
        logger.warning("Failed to load stock history for %s: %s", symbol, exc)
        seeds = []

    # 如果没有找到历史数据且允许样例，则生成假数据方便调试
    if not seeds and settings.allow_sample_stock_seeds:
        sample = None
        for s in DEFAULT_MARKET_SEEDS:
            if s.symbol == symbol:
                sample = s
                break
        if not sample:
            sample = MarketSeed(
                symbol=symbol,
                name=f"模拟{symbol}",
                trade_date=trade_date,
                close_price=100.0,
                momentum5=1.2,
                momentum20=3.5,
                volatility20=2.0,
                volume_ratio=1.1,
                drawdown20=4.5,
                trend_strength=1.5,
                net_mf_amount=800.0,
                pe_ttm=25.0,
                pb=3.0,
                turnover_rate=1.5,
                news_heat=2,
                positive_news_rate=0.6,
            )

        base_date = datetime.strptime(trade_date, "%Y-%m-%d")
        seeds = []
        for i in range(120):
            curr_date = (base_date - timedelta(days=i)).strftime("%Y-%m-%d")
            seed_copy = copy.deepcopy(sample)
            seed_copy.trade_date = curr_date
            seed_copy.close_price = round(sample.close_price * (1 + (i % 12 - 6) * 0.01), 3)
            seed_copy.momentum20 = round(sample.momentum20 + (i % 8 - 4) * 0.2, 2)
            seeds.append(seed_copy)

    if not seeds:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"No stock history found for symbol {symbol} on date {trade_date}",
        )

    # 按交易日期升序排序
    seeds = sorted(seeds, key=lambda s: s.trade_date)

    # 2. 提取特征
    factory = StockFeatureFactory()
    features = factory.build(seeds)
    if not features:
        raise HTTPException(
            status_code=status.HTTP_422_UNPROCESSABLE_ENTITY,
            detail=f"Failed to build features from seeds for symbol {symbol}",
        )
    latest_feature = features[-1]

    # 3. 研判大盘状态
    from app.schemas.stock import StockSelectionPayload
    market_regime = _detect_market_regime(seeds, StockSelectionPayload())

    # 4. 多智能体初审 & 场景推演
    agent_panel = AgentPanel()
    scenario_engine = StockScenarioEngine(agent_panel=agent_panel)
    simulation_cards = scenario_engine.simulate([latest_feature])
    sim_card = simulation_cards[0]

    # 5. 调用大语言模型（LLM）进行临门一脚终审
    llm_client = LLMClient()
    # 动态覆盖大模型配置（若 Go 后端提供了参数）
    if req.llm_config:
        if req.llm_config.get("api_key"):
            llm_client.settings.llm_api_key = req.llm_config.get("api_key")
        if req.llm_config.get("base_url"):
            llm_client.settings.llm_base_url = req.llm_config.get("base_url")
        if req.llm_config.get("model_name"):
            llm_client.settings.llm_model = req.llm_config.get("model_name")

    # 组织智能体意见列表
    local_opinions = []
    for op in sim_card.agents:
        local_opinions.append(f"{op.agent}智能体认为：{op.summary} ({op.stance})")

    metrics = {
        "quant_score": latest_feature.quant_score,
        "total_score": latest_feature.score,
        "risk_level": latest_feature.risk_level,
        "momentum5": getattr(latest_feature, "momentum5", 0.0),
        "momentum20": getattr(latest_feature, "momentum20", 0.0),
        "volatility20": getattr(latest_feature, "volatility20", 0.0),
        "volume_ratio": getattr(latest_feature, "volume_ratio", 1.0),
        "drawdown20": getattr(latest_feature, "drawdown20", 0.0),
        "pe_ttm": getattr(latest_feature, "pe_ttm", 0.0),
        "pb": getattr(latest_feature, "pb", 0.0),
        "industry": getattr(latest_feature, "industry", ""),
        "sector": getattr(latest_feature, "sector", ""),
        "theme_tags": getattr(latest_feature, "theme_tags", []),
        "risk_flags": getattr(latest_feature, "risk_flags", []),
    }

    # 大模型审查
    llm_res = llm_client.review_stock(
        symbol=symbol,
        name=latest_feature.name,
        metrics=metrics,
        local_opinions=local_opinions,
        market_regime=market_regime,
    )

    # 6. 计算风控指标与退场条件
    take_profit, stop_loss, position_range = _risk_controls(latest_feature.risk_level)

    # 计算失效条件
    invalidations = ["20日动量跌破0，趋势逻辑失效", "主力资金转负且量比回落到1以下"]
    if latest_feature.risk_level == "LOW":
        invalidations.append("回撤超过3%或跌破短期支撑位")
    elif latest_feature.risk_level == "HIGH":
        invalidations.append("回撤超过7%或高波动放大")
    else:
        invalidations.append("回撤超过5%或波动显著放大")
    if latest_feature.news_heat >= 4:
        invalidations.append("热点舆情转负，正面占比跌破35%")
    invalidations.extend(latest_feature.veto_reasons[:2])

    # 7. 构建场景快照
    scenario_snapshots = []
    for outcome in sim_card.scenarios:
        sc = outcome.scenario
        trigger = "突破前高且成交量放大"
        conf = "成交量比放大至1.5倍以上，主力资金加速流入"
        inval = "未突破即放量回落，或跌破起涨点"
        window = "3-5个交易日"
        conf_val = 0.75

        if sc == "base":
            trigger = "维持均线附近震荡，缩量整理"
            conf = "波动率降低，缩量蓄势，量比维持在0.8-1.2"
            inval = "选择破位向下，波动率异常放大"
            window = "5-10个交易日"
            conf_val = 0.80
        elif sc == "bear":
            trigger = "跌破5日/10日均线，且成交量萎缩"
            conf = "主力资金持续净流出，日内反弹无量"
            inval = "放量阳线突破均线压制，趋势反转"
            window = "2-3个交易日"
            conf_val = 0.65
        elif sc == "shock":
            trigger = "突发板块重大利空，或破位大阴线跌破多根长短期均线"
            conf = "高开低走或直接低开低走，破位成交量异常放大"
            inval = "迅速收回破位失地，形成长下影线金针探底"
            window = "1-2个交易日"
            conf_val = 0.50

        scenario_snapshots.append({
            "scenario": sc,
            "thesis": outcome.thesis,
            "trigger": trigger,
            "confirmation_signal": conf,
            "invalidation_signal": inval,
            "expected_window": window,
            "action_suggestion": outcome.action,
            "confidence": conf_val,
        })

    # 组装智能体意见
    agent_opinions_data = []
    for op in sim_card.agents:
        agent_opinions_data.append({
            "agent": op.agent,
            "stance": op.stance,
            "confidence": float(op.confidence),
            "summary": op.summary,
            "veto": op.veto,
        })

    # 组装返回数据结构
    go_response = {
        "recommendation": {
            "symbol": symbol,
            "name": latest_feature.name,
            "score": float(latest_feature.score),
            "risk_level": latest_feature.risk_level,
            "position_range": position_range,
            "reason_summary": f"{latest_feature.reason_summary} | AI终审研判：{llm_res['analysis']}",
        },
        "detail": {
            "tech_score": float(round(_clamp(latest_feature.trend_score, 55, 98), 2)),
            "fund_score": float(round(_clamp(latest_feature.quality_score, 52, 97), 2)),
            "sentiment_score": float(round(_clamp(latest_feature.event_score, 50, 96), 2)),
            "money_flow_score": float(round(_clamp(latest_feature.flow_score, 50, 97), 2)),
            "take_profit": take_profit,
            "stop_loss": stop_loss,
            "risk_note": "量化信号存在失效风险，仅供参考，不构成投资建议",
        },
        "explanation": {
            "seed_summary": f"个股收盘价：{latest_feature.close_price}，20日动量：{latest_feature.momentum20}%，量比：{latest_feature.volume_ratio}。",
            "graph_summary": f"与行业 '{latest_feature.industry}' 其它标的呈现趋势共振，资金净流向偏{'多' if latest_feature.net_mf_amount > 0 else '空'}。",
            "consensus_summary": f"多代理收敛：通过 {sum(1 for op in sim_card.agents if op.stance == 'POSITIVE')}，观望 {sum(1 for op in sim_card.agents if op.stance == 'NEGATIVE')}，否决 {sum(1 for op in sim_card.agents if op.veto)}。",
            "risk_flags": latest_feature.risk_flags,
            "invalidations": invalidations,
            "agent_opinions": agent_opinions_data,
            "scenario_snapshots": scenario_snapshots,
            "confidence_calibration": {
                "base_confidence": 0.8,
                "adjusted_confidence": 0.85,
                "drivers": [
                    {
                        "label": "技术趋势得分",
                        "impact": 0.05,
                        "note": "中期趋势较好对置信度有正向修正作用",
                        "source_key": "trend",
                    }
                ],
                "advisory_only": False,
            },
        },
    }

    return go_response
