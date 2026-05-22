#!/usr/bin/env python3
"""超短线 T+1 回测 CLI 工具

用法:
    python -m app.tools.backtest_runner --start 2025-11-01 --end 2026-05-01
    python -m app.tools.backtest_runner --start 2025-11-01 --end 2026-05-01 --output report.json
    python -m app.tools.backtest_runner --start 2025-11-01 --end 2026-05-01 --output report.csv --format csv

选项:
    --start         回测起始日期 (YYYY-MM-DD)
    --end           回测结束日期 (YYYY-MM-DD)
    --limit         每日推荐数量 (默认 5)
    --min-score     最低评分 (默认 65)
    --max-risk      最大风险等级 (LOW/MEDIUM/HIGH, 默认 MEDIUM)
    --output        输出文件路径
    --format        输出格式 (json/csv/summary, 默认 summary)
"""

from __future__ import annotations

import argparse
import logging
import sys

from app.domain.backtesting.backtest_reporter import (
    backtest_result_to_dict,
    backtest_to_json,
    backtest_trades_to_csv,
    format_backtest_summary,
)
from app.domain.backtesting.intraday_backtest import IntradayBacktestRunner
from app.domain.seeds.market_seed_loader import MarketSeedLoader

logger = logging.getLogger(__name__)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="超短线 T+1 回测工具")
    parser.add_argument("--start", required=True, help="回测起始日期 (YYYY-MM-DD)")
    parser.add_argument("--end", required=True, help="回测结束日期 (YYYY-MM-DD)")
    parser.add_argument("--limit", type=int, default=5, help="每日推荐数量")
    parser.add_argument("--min-score", type=float, default=65.0, help="最低评分")
    parser.add_argument("--max-risk", type=str, default="MEDIUM", help="最大风险等级")
    parser.add_argument("--max-per-strategy", type=int, default=2, help="每策略最大输出数")
    parser.add_argument("--max-per-sector", type=int, default=1, help="每板块最大输出数")
    parser.add_argument("--output", type=str, default="", help="输出文件路径")
    parser.add_argument("--format", type=str, default="summary", choices=["summary", "json", "csv"])

    args = parser.parse_args(argv)

    logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
    logger.info("starting backtest: %s ~ %s", args.start, args.end)

    runner = IntradayBacktestRunner(MarketSeedLoader())
    result = runner.run(
        start_date=args.start,
        end_date=args.end,
        limit=args.limit,
        min_score=args.min_score,
        max_risk_level=args.max_risk,
        max_per_strategy=args.max_per_strategy,
        max_symbols_per_sector=args.max_per_sector,
    )

    if args.output:
        content: str
        if args.format == "json":
            content = backtest_to_json(result)
        elif args.format == "csv":
            content = backtest_trades_to_csv(result)
        else:
            content = format_backtest_summary(result)

        with open(args.output, "w", encoding="utf-8") as f:
            f.write(content)
        logger.info("report saved to %s", args.output)
    else:
        print(format_backtest_summary(result))

    return 0


if __name__ == "__main__":
    sys.exit(main())
