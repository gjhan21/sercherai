import test from "node:test";
import assert from "node:assert/strict";

import {
  buildMarketCenterQualitySummaryReturnQuery,
  buildMarketCenterQualityReturnQuery,
  buildMarketQualityDrillQuery,
  DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS,
  MARKET_QUALITY_LOOKBACK_OPTIONS,
  areMarketQualityFiltersEqual,
  buildMarketCenterRouteQuery,
  buildMarketQualityDialogTitle,
  buildMarketQualityRouteQuery,
  buildMarketQualityBucketSummary,
  collectMarketQualityIssueOptions,
  formatMarketCenterReturnLabel,
  formatMarketQualityLookbackLabel,
  formatMarketQualityPayload,
  formatTruthRebuildSuccessMessage,
  marketQualitySeverityTagType,
  normalizeMarketCenterRouteState,
  normalizeMarketCenterTab,
  normalizeMarketQualityLookbackHours,
  normalizeMarketQualityRouteContext,
  normalizeMarketQualityRouteFilters
} from "./market-data-admin.js";

test("marketQualitySeverityTagType maps warn error info to element tag types", () => {
  assert.equal(marketQualitySeverityTagType("warn"), "warning");
  assert.equal(marketQualitySeverityTagType("ERROR"), "danger");
  assert.equal(marketQualitySeverityTagType("Info"), "success");
  assert.equal(marketQualitySeverityTagType("unknown"), "info");
});

test("formatTruthRebuildSuccessMessage formats stock rebuild summary", () => {
  assert.equal(
    formatTruthRebuildSuccessMessage("STOCK", {
      truth_bar_count: 12
    }),
    "股票状态 truth 重建完成，处理 12 条 truth 行情"
  );
});

test("formatTruthRebuildSuccessMessage formats futures rebuild summary", () => {
  assert.equal(
    formatTruthRebuildSuccessMessage("FUTURES", {
      truth_bar_count: 8
    }),
    "期货主力映射重建完成，处理 8 条 truth 行情"
  );
});

test("formatMarketQualityPayload pretty prints json payload", () => {
  assert.equal(
    formatMarketQualityPayload('{"asset_class":"STOCK","truth_bar_count":12}'),
    '{\n  "asset_class": "STOCK",\n  "truth_bar_count": 12\n}'
  );
});

test("formatMarketQualityPayload falls back to trimmed raw string", () => {
  assert.equal(formatMarketQualityPayload("  upstream timeout  "), "upstream timeout");
  assert.equal(formatMarketQualityPayload(""), "");
});

test("collectMarketQualityIssueOptions sorts by frequency then code", () => {
  assert.deepEqual(
    collectMarketQualityIssueOptions(
      [
        { issue_code: "SOURCE_FETCH_FAILED" },
        { issue_code: "BAR_UPSERT_FAILED" },
        { issue_code: "SOURCE_FETCH_FAILED" },
        { issue_code: "DERIVED_STOCK_STATUS_REBUILT" },
        { issue_code: "BAR_UPSERT_FAILED" },
        { issue_code: "BAR_UPSERT_FAILED" }
      ],
      2
    ),
    [
      { value: "BAR_UPSERT_FAILED", label: "BAR_UPSERT_FAILED (3)" },
      { value: "SOURCE_FETCH_FAILED", label: "SOURCE_FETCH_FAILED (2)" }
    ]
  );
});

test("buildMarketQualityDialogTitle uses asset and issue code when present", () => {
  assert.equal(
    buildMarketQualityDialogTitle({
      asset_class: "STOCK",
      issue_code: "SOURCE_FETCH_FAILED"
    }),
    "质量日志详情 · STOCK · SOURCE_FETCH_FAILED"
  );
  assert.equal(buildMarketQualityDialogTitle({}), "质量日志详情");
});

test("buildMarketQualityRouteQuery trims values and omits empty fields", () => {
  assert.deepEqual(
    buildMarketQualityRouteQuery({
      asset_class: " stock ",
      severity: " warn ",
      issue_code: " source_fetch_failed ",
      hours: 24
    }),
    {
      asset_class: "STOCK",
      severity: "WARN",
      issue_code: "SOURCE_FETCH_FAILED",
      hours: "24"
    }
  );
});

test("normalizeMarketQualityLookbackHours falls back to default for invalid values", () => {
  assert.equal(DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS, 24);
  assert.deepEqual(
    MARKET_QUALITY_LOOKBACK_OPTIONS.map((item) => item.value),
    [24, 72, 168]
  );
  assert.equal(normalizeMarketQualityLookbackHours(" 72 "), 72);
  assert.equal(normalizeMarketQualityLookbackHours(0), 24);
  assert.equal(normalizeMarketQualityLookbackHours("bad"), 24);
});

test("formatMarketQualityLookbackLabel prefers preset labels and falls back to hours", () => {
  assert.equal(formatMarketQualityLookbackLabel(24), "近 24 小时");
  assert.equal(formatMarketQualityLookbackLabel("72"), "近 3 天");
  assert.equal(formatMarketQualityLookbackLabel(96), "近 96 小时");
});

test("market center route helpers normalize tab and quality state", () => {
  assert.equal(normalizeMarketCenterTab("FUTURES"), "futures");
  assert.equal(normalizeMarketCenterTab("bad"), "stocks");
  assert.deepEqual(
    buildMarketCenterRouteQuery({
      tab: "futures",
      quality_hours: "72",
      issue_code: " source_fetch_failed "
    }),
    {
      tab: "futures",
      quality_hours: "72",
      issue_code: "SOURCE_FETCH_FAILED"
    }
  );
  assert.deepEqual(
    normalizeMarketCenterRouteState({
      tab: "futures",
      quality_hours: "168",
      issue_code: "bar_upsert_failed"
    }),
    {
      tab: "futures",
      quality_hours: 168,
      issue_code: "BAR_UPSERT_FAILED"
    }
  );
});

test("market quality route context keeps market-center return info", () => {
  assert.deepEqual(
    normalizeMarketQualityRouteContext({
      return_to: "market-center",
      return_tab: "futures"
    }),
    {
      return_to: "market-center",
      return_tab: "futures"
    }
  );
  assert.deepEqual(normalizeMarketQualityRouteContext({ return_to: "dashboard" }), {});
  assert.equal(formatMarketCenterReturnLabel("futures"), "返回期货行情中心");
  assert.equal(formatMarketCenterReturnLabel("stocks"), "返回股票行情中心");
});

test("buildMarketQualityDrillQuery merges route filters with market-center return context", () => {
  assert.deepEqual(
    buildMarketQualityDrillQuery({
      asset_class: "stock",
      severity: "warn",
      issue_code: "source_fetch_failed",
      hours: 72,
      return_tab: "stocks"
    }),
    {
      return_to: "market-center",
      return_tab: "stocks",
      asset_class: "STOCK",
      severity: "WARN",
      issue_code: "SOURCE_FETCH_FAILED",
      hours: "72"
    }
  );
});

test("buildMarketCenterQualityReturnQuery maps quality row back to matching market-center tab", () => {
  assert.deepEqual(
    buildMarketCenterQualityReturnQuery(
      {
        asset_class: "FUTURES",
        issue_code: "source_fetch_failed"
      },
      168
    ),
    {
      tab: "futures",
      quality_hours: "168",
      issue_code: "SOURCE_FETCH_FAILED"
    }
  );
  assert.deepEqual(
    buildMarketCenterQualityReturnQuery(
      {
        asset_class: "stock",
        issue_code: "bar_upsert_failed"
      },
      24
    ),
    {
      tab: "stocks",
      quality_hours: "24",
      issue_code: "BAR_UPSERT_FAILED"
    }
  );
});

test("buildMarketCenterQualitySummaryReturnQuery uses latest issue or latest error for market-center jump", () => {
  assert.deepEqual(
    buildMarketCenterQualitySummaryReturnQuery(
      "stock",
      {
        latest_issue_code: "bar_upsert_retried",
        latest_error_issue_code: "source_fetch_failed"
      },
      72,
      "latest"
    ),
    {
      tab: "stocks",
      quality_hours: "72",
      issue_code: "BAR_UPSERT_RETRIED"
    }
  );
  assert.deepEqual(
    buildMarketCenterQualitySummaryReturnQuery(
      "futures",
      {
        latest_issue_code: "mapping_rebuilt",
        latest_error_issue_code: "source_fetch_failed"
      },
      24,
      "error"
    ),
    {
      tab: "futures",
      quality_hours: "24",
      issue_code: "SOURCE_FETCH_FAILED"
    }
  );
});

test("normalizeMarketQualityRouteFilters normalizes query values", () => {
  assert.deepEqual(
    normalizeMarketQualityRouteFilters({
      asset_class: "futures",
      data_kind: " daily_bars ",
      severity: "error",
      issue_code: "bar_upsert_failed",
      hours: " 72 "
    }),
    {
      asset_class: "FUTURES",
      data_kind: "DAILY_BARS",
      severity: "ERROR",
      issue_code: "BAR_UPSERT_FAILED",
      hours: 72
    }
  );
});

test("buildMarketQualityBucketSummary groups and sorts by count", () => {
  assert.deepEqual(
    buildMarketQualityBucketSummary(
      [
        { source_key: "MYSELF" },
        { source_key: "TUSHARE" },
        { source_key: "MYSELF" },
        { source_key: "SINA" },
        { source_key: "MYSELF" },
        { source_key: "" }
      ],
      (item) => item.source_key,
      2
    ),
    [
      { value: "MYSELF", count: 3, label: "MYSELF" },
      { value: "SINA", count: 1, label: "SINA" }
    ]
  );
});

test("areMarketQualityFiltersEqual compares normalized route and form filters", () => {
  assert.equal(
    areMarketQualityFiltersEqual(
      {
        asset_class: "stock",
        data_kind: "daily_bars",
        severity: "warn",
        issue_code: "source_fetch_failed",
        hours: "24"
      },
      {
        asset_class: " STOCK ",
        data_kind: "DAILY_BARS",
        severity: " WARN ",
        issue_code: " SOURCE_FETCH_FAILED ",
        hours: 24
      }
    ),
    true
  );
  assert.equal(
    areMarketQualityFiltersEqual(
      { asset_class: "STOCK", issue_code: "SOURCE_FETCH_FAILED", hours: 24 },
      { asset_class: "FUTURES", issue_code: "SOURCE_FETCH_FAILED", hours: 72 }
    ),
    false
  );
});
