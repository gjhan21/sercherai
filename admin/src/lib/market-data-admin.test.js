import test from "node:test";
import assert from "node:assert/strict";

import {
  buildMarketCenterQualitySummaryReturnQuery,
  buildMarketCenterQualityReturnQuery,
  buildMarketCoverageAssetRows,
  buildMarketCoverageOverviewCards,
  buildMarketQualityDrillQuery,
  validateMarketBackfillLongHistoryInput,
  buildUniverseSnapshotDigest,
  DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS,
  MARKET_QUALITY_LOOKBACK_OPTIONS,
  areMarketQualityFiltersEqual,
  buildMarketCenterRouteQuery,
  buildMarketQualityDialogTitle,
  buildMarketQualityRouteQuery,
  buildMarketQualityBucketSummary,
  collectMarketQualityIssueOptions,
  formatMarketAssetScopeSummary,
  formatMarketBackfillStatusLabel,
  formatMarketBackfillRunTypeLabel,
  formatMarketBackfillStageLabel,
  formatMarketCenterReturnLabel,
  formatMarketQualityLookbackLabel,
  formatMarketQualityPayload,
  formatTruthRebuildSuccessMessage,
  marketBackfillDetailStatusTagType,
  marketBackfillStatusTagType,
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

test("market backfill labels format run type and stage in Chinese", () => {
  assert.equal(formatMarketBackfillRunTypeLabel("full"), "全量回填");
  assert.equal(formatMarketBackfillRunTypeLabel("INCREMENTAL"), "增量回填");
  assert.equal(formatMarketBackfillRunTypeLabel("REBUILD_ONLY"), "仅重建");
  assert.equal(formatMarketBackfillStageLabel("universe"), "Universe");
  assert.equal(formatMarketBackfillStageLabel("DAILY_BASIC"), "Daily Basic");
  assert.equal(formatMarketBackfillStageLabel("coverage_summary"), "覆盖率汇总");
  assert.equal(formatMarketBackfillStatusLabel("skipped"), "已跳过");
});

test("marketBackfillDetailStatusTagType maps run detail status to element tag types", () => {
  assert.equal(marketBackfillDetailStatusTagType("success"), "success");
  assert.equal(marketBackfillDetailStatusTagType("running"), "warning");
  assert.equal(marketBackfillDetailStatusTagType("failed"), "danger");
  assert.equal(marketBackfillDetailStatusTagType("skipped"), "info");
  assert.equal(marketBackfillDetailStatusTagType("unknown"), "info");
});

test("validateMarketBackfillLongHistoryInput rejects unsupported long history combinations", () => {
  assert.equal(
    validateMarketBackfillLongHistoryInput({
      run_type: "FULL",
      asset_scope: ["STOCK"],
      source_key: "AKSHARE",
      trade_date_from: "2024-01-01",
      trade_date_to: "2025-01-05"
    }),
    "超过 365 天的股票长历史回补当前仅支持 TUSHARE 数据源"
  );

  assert.equal(
    validateMarketBackfillLongHistoryInput({
      run_type: "INCREMENTAL",
      asset_scope: ["STOCK"],
      source_key: "TUSHARE",
      trade_date_from: "2024-01-01",
      trade_date_to: "2025-01-05"
    }),
    "超过 365 天的股票长历史回补当前只允许 FULL 运行类型"
  );

  assert.equal(
    validateMarketBackfillLongHistoryInput({
      run_type: "FULL",
      asset_scope: ["STOCK"],
      source_key: "TUSHARE",
      trade_date_from: "2024-01-01",
      trade_date_to: "2025-01-05",
      stages: ["MASTER", "TRUTH"]
    }),
    "长历史股票回补的阶段范围必须包含 QUOTES"
  );

  assert.equal(
    validateMarketBackfillLongHistoryInput({
      run_type: "FULL",
      asset_scope: ["STOCK"],
      source_key: "TUSHARE",
      trade_date_from: "2024-01-01",
      trade_date_to: "2025-01-05",
      stages: ["QUOTES", "TRUTH"]
    }),
    ""
  );
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

test("buildMarketCoverageOverviewCards builds Chinese market coverage cards", () => {
  assert.deepEqual(
    buildMarketCoverageOverviewCards({
      total_universe_count: 4200,
      quotes_coverage_count: 3980,
      daily_basic_coverage_count: 1860,
      moneyflow_coverage_count: 1750,
      latest_trade_date: "2026-03-24",
      canonical_key_gap_count: 12,
      display_name_gap_count: 8,
      list_date_gap_count: 3
    }),
    [
      {
        key: "universe",
        title: "证券全集",
        value: "4200",
        tone: "primary",
        helper: "当前快照里的可跟踪标的总数"
      },
      {
        key: "quotes",
        title: "行情覆盖",
        value: "3980",
        tone: "success",
        helper: "已落到 truth 的行情标的数"
      },
      {
        key: "enhancement",
        title: "增强因子覆盖",
        value: "1860 / 1750",
        tone: "warning",
        helper: "分别对应 daily_basic / moneyflow 覆盖数"
      },
      {
        key: "gaps",
        title: "主数据缺口",
        value: "12 / 8 / 3",
        tone: "danger",
        helper: "分别对应 canonical key / 名称 / 上市日期缺口"
      },
      {
        key: "latest_trade_date",
        title: "最新交易日",
        value: "2026-03-24",
        tone: "gold",
        helper: "用于判断行情覆盖是否更新到最近交易日"
      }
    ]
  );
});

test("buildMarketCoverageAssetRows and scope helpers format asset level Chinese summaries", () => {
  assert.deepEqual(
    buildMarketCoverageAssetRows({
      asset_items: [
        {
          asset_type: "ETF",
          universe_count: 120,
          master_coverage_count: 120,
          quotes_coverage_count: 118,
          daily_basic_coverage_count: 0,
          moneyflow_coverage_count: 0,
          latest_trade_date: "2026-03-24"
        },
        {
          asset_type: "STOCK",
          universe_count: 3800,
          master_coverage_count: 3788,
          quotes_coverage_count: 3610,
          daily_basic_coverage_count: 1860,
          moneyflow_coverage_count: 1750,
          latest_trade_date: "2026-03-24"
        }
      ]
    }),
    [
      {
        key: "STOCK",
        label: "股票",
        universeCount: 3800,
        masterCoverageCount: 3788,
        quotesCoverageCount: 3610,
        dailyBasicCoverageCount: 1860,
        moneyflowCoverageCount: 1750,
        latestTradeDate: "2026-03-24"
      },
      {
        key: "ETF",
        label: "ETF",
        universeCount: 120,
        masterCoverageCount: 120,
        quotesCoverageCount: 118,
        dailyBasicCoverageCount: 0,
        moneyflowCoverageCount: 0,
        latestTradeDate: "2026-03-24"
      }
    ]
  );

  assert.equal(formatMarketAssetScopeSummary(["stock", "index", "cbond"]), "股票 / 指数 / 可转债");
  assert.equal(formatMarketAssetScopeSummary([]), "未设置");
});

test("market backfill helpers format status and snapshot digest in Chinese", () => {
  assert.equal(formatMarketBackfillStatusLabel("PARTIAL_SUCCESS"), "部分完成");
  assert.equal(marketBackfillStatusTagType("FAILED"), "danger");
  assert.equal(marketBackfillStatusTagType("RUNNING"), "warning");
  assert.equal(
    buildUniverseSnapshotDigest({
      scope: ["STOCK", "INDEX", "ETF"],
      snapshot_date: "2026-03-24",
      source_key: "TUSHARE"
    }),
    "2026-03-24 · 股票 / 指数 / ETF · TUSHARE"
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
