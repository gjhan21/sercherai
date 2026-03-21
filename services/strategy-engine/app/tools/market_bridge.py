#!/usr/bin/env python3
from __future__ import annotations

import argparse
import hashlib
import json
import sys
from datetime import datetime, timedelta
from typing import Iterable


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Market data bridge for external providers")
    subparsers = parser.add_subparsers(dest="command", required=True)

    subparsers.add_parser("healthcheck", help="Verify akshare can be imported")

    stock_daily = subparsers.add_parser("stock_daily", help="Fetch stock daily bars via AkShare")
    stock_daily.add_argument("--symbols", default="", help="Comma-separated stock symbols")
    stock_daily.add_argument("--days", type=int, default=120, help="Lookback days")

    futures_daily = subparsers.add_parser("futures_daily", help="Fetch futures daily bars via AkShare")
    futures_daily.add_argument("--symbols", default="", help="Comma-separated futures symbols")
    futures_daily.add_argument("--days", type=int, default=120, help="Lookback days")

    market_news = subparsers.add_parser("market_news", help="Fetch market news via AkShare")
    market_news.add_argument("--symbols", default="", help="Comma-separated stock symbols")
    market_news.add_argument("--days", type=int, default=7, help="Lookback days")
    market_news.add_argument("--limit", type=int, default=50, help="Maximum returned items")

    return parser.parse_args()


def import_akshare():
    try:
        import akshare as ak  # type: ignore
    except Exception as exc:  # pragma: no cover - runtime dependency
        print(str(exc), file=sys.stderr)
        raise
    return ak


def split_symbols(raw: str) -> list[str]:
    items: list[str] = []
    seen: set[str] = set()
    for part in raw.replace(";", ",").replace("|", ",").split(","):
        normalized = part.strip().upper()
        if not normalized or normalized in seen:
            continue
        seen.add(normalized)
        items.append(normalized)
    return items


def normalize_akshare_stock_symbol(symbol: str) -> str:
    normalized = symbol.strip().upper()
    if "." in normalized:
        normalized = normalized.split(".", 1)[0]
    if normalized.startswith("SH") or normalized.startswith("SZ") or normalized.startswith("BJ"):
        normalized = normalized[2:]
    return normalized


def normalize_akshare_futures_symbol(symbol: str) -> str:
    normalized = symbol.strip().upper()
    if "." in normalized:
        normalized = normalized.split(".", 1)[0]
    return normalized


def row_value(row, candidates: Iterable[str], default=None):
    for candidate in candidates:
        if candidate in row and row[candidate] not in ("", None):
            value = row[candidate]
            try:
                if value != value:  # NaN
                    continue
            except Exception:
                pass
            return value
    return default


def safe_float(value, default: float = 0.0) -> float:
    try:
        if value is None or value == "":
            return default
        return float(value)
    except Exception:
        return default


def safe_text(value) -> str:
    if value is None:
        return ""
    return str(value).strip()


def parse_time_text(value) -> str:
    text = safe_text(value)
    if not text:
        return ""
    for fmt in (
        "%Y-%m-%d %H:%M:%S",
        "%Y-%m-%d %H:%M",
        "%Y/%m/%d %H:%M:%S",
        "%Y/%m/%d %H:%M",
        "%Y-%m-%d",
        "%Y/%m/%d",
    ):
        try:
            return datetime.strptime(text, fmt).isoformat()
        except ValueError:
            continue
    return text


def build_external_id(url: str, title: str, published_at: str) -> str:
    base = f"{url}|{title}|{published_at}"
    return hashlib.sha1(base.encode("utf-8")).hexdigest()[:24]


def command_healthcheck() -> dict:
    ak = import_akshare()
    return {
        "status": "ok",
        "provider": "AKSHARE",
        "version": getattr(ak, "__version__", "")
    }


def command_stock_daily(symbols: list[str], days: int) -> dict:
    ak = import_akshare()
    start_date = (datetime.now() - timedelta(days=max(days + 10, 20))).strftime("%Y%m%d")
    end_date = datetime.now().strftime("%Y%m%d")
    items: list[dict] = []

    for raw_symbol in symbols:
        symbol = normalize_akshare_stock_symbol(raw_symbol)
        if not symbol:
            continue
        try:
            frame = ak.stock_zh_a_hist(
                symbol=symbol,
                period="daily",
                start_date=start_date,
                end_date=end_date,
                adjust=""
            )
        except Exception:
            continue
        if frame is None or frame.empty:
            continue

        date_column = "日期" if "日期" in frame.columns else "date"
        frame = frame.sort_values(by=date_column)
        prev_close = 0.0
        for _, row in frame.iterrows():
            trade_date_text = safe_text(row_value(row, ["日期", "date", "Date"]))
            if not trade_date_text:
                continue
            open_price = safe_float(row_value(row, ["开盘", "open", "Open"]))
            high_price = safe_float(row_value(row, ["最高", "high", "High"]))
            low_price = safe_float(row_value(row, ["最低", "low", "Low"]))
            close_price = safe_float(row_value(row, ["收盘", "close", "Close"]))
            if close_price <= 0:
                continue
            volume = safe_float(row_value(row, ["成交量", "volume", "Volume"]))
            turnover = safe_float(row_value(row, ["成交额", "amount", "Turnover"]))
            if prev_close <= 0:
                prev_close = open_price if open_price > 0 else close_price

            items.append(
                {
                    "instrument_key": raw_symbol,
                    "external_symbol": symbol,
                    "trade_date": trade_date_text,
                    "open_price": open_price,
                    "high_price": high_price,
                    "low_price": low_price,
                    "close_price": close_price,
                    "prev_close_price": prev_close,
                    "settle_price": 0.0,
                    "prev_settle_price": 0.0,
                    "volume": volume,
                    "turnover": turnover,
                    "open_interest": 0.0,
                }
            )
            prev_close = close_price

    return {"items": items}


def command_futures_daily(symbols: list[str], days: int) -> dict:
    ak = import_akshare()
    items: list[dict] = []

    for raw_symbol in symbols:
        symbol = normalize_akshare_futures_symbol(raw_symbol)
        if not symbol:
            continue
        try:
            frame = ak.futures_zh_daily_sina(symbol=symbol)
        except Exception:
            continue
        if frame is None or frame.empty:
            continue

        date_column = "date" if "date" in frame.columns else "日期"
        frame = frame.sort_values(by=date_column).tail(max(days, 1))
        prev_close = 0.0
        prev_settle = 0.0
        for _, row in frame.iterrows():
            trade_date_text = safe_text(row_value(row, ["date", "日期", "Date"]))
            if not trade_date_text:
                continue
            open_price = safe_float(row_value(row, ["open", "开盘", "Open"]))
            high_price = safe_float(row_value(row, ["high", "最高", "High"]))
            low_price = safe_float(row_value(row, ["low", "最低", "Low"]))
            close_price = safe_float(row_value(row, ["close", "收盘", "Close"]))
            if close_price <= 0:
                continue
            volume = safe_float(row_value(row, ["volume", "成交量", "Volume"]))
            settle_price = safe_float(row_value(row, ["settle", "结算价", "Settle"]))
            open_interest = safe_float(row_value(row, ["hold", "open_interest", "持仓量", "OpenInterest"]))
            if prev_close <= 0:
                prev_close = open_price if open_price > 0 else close_price
            if prev_settle <= 0:
                prev_settle = settle_price if settle_price > 0 else prev_close

            items.append(
                {
                    "instrument_key": raw_symbol,
                    "external_symbol": symbol,
                    "trade_date": trade_date_text,
                    "open_price": open_price,
                    "high_price": high_price,
                    "low_price": low_price,
                    "close_price": close_price,
                    "prev_close_price": prev_close,
                    "settle_price": settle_price if settle_price > 0 else close_price,
                    "prev_settle_price": prev_settle,
                    "volume": volume,
                    "turnover": 0.0,
                    "open_interest": open_interest,
                }
            )
            prev_close = close_price
            prev_settle = settle_price if settle_price > 0 else close_price

    return {"items": items}


def command_market_news(symbols: list[str], days: int, limit: int) -> dict:
    ak = import_akshare()
    frame = None
    loaders = []
    if hasattr(ak, "stock_info_global_em"):
        loaders.append(lambda: ak.stock_info_global_em())
    if hasattr(ak, "stock_info_global_cls"):
        loaders.append(lambda: ak.stock_info_global_cls())

    for loader in loaders:
        try:
            frame = loader()
        except Exception:
            continue
        if frame is not None and not frame.empty:
            break

    if frame is None or frame.empty:
        return {"items": []}

    cutoff = datetime.now() - timedelta(days=max(days, 1))
    symbol_keywords = [normalize_akshare_stock_symbol(symbol) for symbol in symbols if symbol]
    items: list[dict] = []

    for _, row in frame.iterrows():
        title = safe_text(row_value(row, ["标题", "title", "Title", "新闻标题"]))
        summary = safe_text(row_value(row, ["内容", "摘要", "summary", "内容摘要"]))
        url = safe_text(row_value(row, ["链接", "url", "Url", "URL"]))
        published_at = parse_time_text(row_value(row, ["发布时间", "时间", "datetime", "publish_time", "发布时间戳"]))
        if not title:
            continue
        published_dt = None
        if published_at:
            try:
                published_dt = datetime.fromisoformat(published_at)
            except ValueError:
                published_dt = None
        if published_dt and published_dt < cutoff:
            continue

        matched_symbols: list[str] = []
        for keyword in symbol_keywords:
            if keyword and (keyword in title or keyword in summary):
                matched_symbols.append(keyword)
        if symbol_keywords and not matched_symbols:
            continue

        items.append(
            {
                "external_id": build_external_id(url, title, published_at or title),
                "news_type": "MARKET",
                "title": title,
                "summary": summary[:500],
                "content": summary,
                "url": url,
                "primary_symbol": matched_symbols[0] if matched_symbols else "",
                "symbols": matched_symbols,
                "published_at": published_at or datetime.now().isoformat(),
            }
        )
        if len(items) >= max(limit, 1):
            break

    return {"items": items}


def main() -> int:
    args = parse_args()
    if args.command == "healthcheck":
        payload = command_healthcheck()
    elif args.command == "stock_daily":
        payload = command_stock_daily(split_symbols(args.symbols), args.days)
    elif args.command == "futures_daily":
        payload = command_futures_daily(split_symbols(args.symbols), args.days)
    elif args.command == "market_news":
        payload = command_market_news(split_symbols(args.symbols), args.days, args.limit)
    else:  # pragma: no cover - argparse should prevent this
        raise ValueError(f"Unsupported command: {args.command}")

    print(json.dumps(payload, ensure_ascii=False))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
