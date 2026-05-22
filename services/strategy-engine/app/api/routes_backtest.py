"""超短线 T+1 回测 API 路由

POST /api/v1/backtest/intraday-t1
  - 触发回测并返回结果

GET /api/v1/backtest/intraday-t1/status
  - 查询异步回测状态 (TODO: 支持异步)
"""

from __future__ import annotations

import logging
from typing import Any

from fastapi import APIRouter, Query
from pydantic import BaseModel, Field

from app.domain.backtesting.backtest_reporter import (
    backtest_result_to_dict,
    backtest_to_json,
    backtest_trades_to_csv,
    format_backtest_summary,
)
from app.domain.backtesting.intraday_backtest import IntradayBacktestRunner
from app.domain.seeds.market_seed_loader import MarketSeedLoader

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/backtest", tags=["backtest"])


class BacktestRequest(BaseModel):
    start_date: str = Field(..., description="回测起始日期 (YYYY-MM-DD)")
    end_date: str = Field(..., description="回测结束日期 (YYYY-MM-DD)")
    limit: int = Field(default=5, ge=1, le=20, description="每日推荐数量")
    min_score: float = Field(default=65.0, ge=0, le=100, description="最低评分")
    max_risk_level: str = Field(default="MEDIUM", description="最大风险等级 (LOW/MEDIUM/HIGH)")
    max_per_strategy: int = Field(default=2, ge=1, le=10, description="每策略最大输出")
    max_per_sector: int = Field(default=1, ge=1, le=10, description="每板块最大输出")
    format: str = Field(default="json", description="输出格式 (json/csv/summary)")


@router.post("/intraday-t1")
async def run_intraday_t1_backtest(req: BacktestRequest) -> dict[str, Any]:
    """运行超短线 T+1 回测。

    返回按策略和按市场环境分层的详细统计结果。
    回测使用当日收盘价买入、次日收盘价卖出的模拟逻辑。
    """
    logger.info("backtest request: %s ~ %s, limit=%d, format=%s", req.start_date, req.end_date, req.limit, req.format)

    runner = IntradayBacktestRunner(MarketSeedLoader())
    result = runner.run(
        start_date=req.start_date,
        end_date=req.end_date,
        limit=req.limit,
        min_score=req.min_score,
        max_risk_level=req.max_risk_level,
        max_per_strategy=req.max_per_strategy,
        max_symbols_per_sector=req.max_per_sector,
    )

    if req.format == "csv":
        return {
            "format": "csv",
            "data": backtest_trades_to_csv(result),
            "summary": format_backtest_summary(result),
        }
    elif req.format == "summary":
        return {
            "format": "summary",
            "data": format_backtest_summary(result),
        }
    else:
        return {
            "format": "json",
            "data": backtest_result_to_dict(result),
        }


class BacktestDatesResponse(BaseModel):
    available_start: str
    available_end: str
    trading_days: int


@router.get("/intraday-t1/available-dates")
async def get_available_backtest_dates() -> BacktestDatesResponse:
    """返回可用于回测的日期范围（最近 180 个交易日）。"""
    from datetime import date, timedelta

    end = date.today()
    start = end - timedelta(days=270)  # about 180 trading days

    trading_count = 0
    d = start
    while d <= end:
        if d.weekday() < 5:
            trading_count += 1
        d += timedelta(days=1)

    return BacktestDatesResponse(
        available_start=start.isoformat(),
        available_end=end.isoformat(),
        trading_days=trading_count,
    )
