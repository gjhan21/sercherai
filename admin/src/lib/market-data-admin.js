export const DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS = 24;

export const MARKET_QUALITY_LOOKBACK_OPTIONS = [
  { label: "近 24 小时", value: 24 },
  { label: "近 3 天", value: 72 },
  { label: "近 7 天", value: 168 }
];

const MARKET_ASSET_LABELS = {
  STOCK: "股票",
  INDEX: "指数",
  ETF: "ETF",
  LOF: "LOF",
  CBOND: "可转债",
  FUTURES: "期货"
};

const MARKET_BACKFILL_STATUS_LABELS = {
  PENDING: "待执行",
  RUNNING: "进行中",
  PARTIAL_SUCCESS: "部分完成",
  SUCCESS: "已完成",
  FAILED: "失败",
  CANCELLED: "已取消",
  SKIPPED: "已跳过"
};

const MARKET_BACKFILL_RUN_TYPE_LABELS = {
  FULL: "全量回填",
  INCREMENTAL: "增量回填",
  REBUILD_ONLY: "仅重建"
};

const MARKET_BACKFILL_STAGE_LABELS = {
  UNIVERSE: "Universe",
  MASTER: "主数据",
  QUOTES: "行情",
  DAILY_BASIC: "Daily Basic",
  MONEYFLOW: "Moneyflow",
  TRUTH: "Truth",
  COVERAGE_SUMMARY: "覆盖率汇总"
};

const KNOWN_NON_TUSHARE_LONG_HISTORY_SOURCE_KEYS = new Set(["AKSHARE", "TICKERMD", "MYSELF", "MOCK"]);

function normalizeBackfillStageList(values = []) {
  const seen = new Set();
  return (Array.isArray(values) ? values : [])
    .map((item) => String(item || "").trim().toUpperCase())
    .filter((item) => {
      if (!item || seen.has(item)) {
        return false;
      }
      seen.add(item);
      return true;
    });
}

function parseBackfillDateValue(value) {
  const text = String(value || "").trim();
  if (!text) {
    return null;
  }
  const match = text.match(/^(\d{4})-(\d{2})-(\d{2})$/);
  if (!match) {
    return null;
  }
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const timestamp = Date.UTC(year, month - 1, day);
  const parsed = new Date(timestamp);
  if (Number.isNaN(parsed.getTime()) || parsed.getUTCFullYear() !== year || parsed.getUTCMonth() !== month - 1 || parsed.getUTCDate() !== day) {
    return null;
  }
  return timestamp;
}

export function normalizeMarketQualityLookbackHours(value, fallback = DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS) {
  const hours = Number(value);
  if (Number.isFinite(hours) && hours > 0) {
    return Math.trunc(hours);
  }
  return fallback;
}

export function formatMarketQualityLookbackLabel(value) {
  const hours = normalizeMarketQualityLookbackHours(value);
  return (
    MARKET_QUALITY_LOOKBACK_OPTIONS.find((item) => item.value === hours)?.label ||
    `近 ${hours} 小时`
  );
}

export function formatMarketAssetLabel(value) {
  const key = String(value || "").trim().toUpperCase();
  return MARKET_ASSET_LABELS[key] || key || "-";
}

export function formatMarketAssetScopeSummary(values = []) {
  const labels = (Array.isArray(values) ? values : [])
    .map((item) => formatMarketAssetLabel(item))
    .filter(Boolean);
  return labels.length ? labels.join(" / ") : "未设置";
}

export function formatMarketBackfillStatusLabel(value) {
  const key = String(value || "").trim().toUpperCase();
  return MARKET_BACKFILL_STATUS_LABELS[key] || key || "未知";
}

export function formatMarketBackfillRunTypeLabel(value) {
  const key = String(value || "").trim().toUpperCase();
  return MARKET_BACKFILL_RUN_TYPE_LABELS[key] || key || "-";
}

export function formatMarketBackfillStageLabel(value) {
  const key = String(value || "").trim().toUpperCase();
  return MARKET_BACKFILL_STAGE_LABELS[key] || key || "-";
}

export function getMarketBackfillDateSpanDays(from, to) {
  const fromDate = parseBackfillDateValue(from);
  const toDate = parseBackfillDateValue(to);
  if (!fromDate || !toDate) {
    return 0;
  }
  const diff = toDate - fromDate;
  if (diff < 0) {
    return -1;
  }
  return Math.floor(diff / 86400000) + 1;
}

export function validateMarketBackfillLongHistoryInput(payload = {}) {
  const spanDays = getMarketBackfillDateSpanDays(payload?.trade_date_from, payload?.trade_date_to);
  if (spanDays <= 365) {
    return "";
  }
  const assetScope = Array.from(
    new Set(
      (Array.isArray(payload?.asset_scope) ? payload.asset_scope : [])
        .map((item) => String(item || "").trim().toUpperCase())
        .filter(Boolean)
    )
  );
  if (assetScope.length !== 1 || assetScope[0] !== "STOCK") {
    return "超过 365 天的长历史回补目前只支持单一 STOCK 资产范围";
  }
  const sourceProvider = String(payload?.source_provider || "").trim().toUpperCase();
  const sourceKey = String(payload?.source_key || "").trim().toUpperCase() || "TUSHARE";
  if (sourceProvider && sourceProvider !== "TUSHARE") {
    return "超过 365 天的股票长历史回补当前仅支持 provider 为 TUSHARE 的数据源";
  }
  if (!sourceProvider && KNOWN_NON_TUSHARE_LONG_HISTORY_SOURCE_KEYS.has(sourceKey)) {
    return "超过 365 天的股票长历史回补当前仅支持 provider 为 TUSHARE 的数据源";
  }
  const runType = String(payload?.run_type || "").trim().toUpperCase() || "FULL";
  if (runType !== "FULL") {
    return "超过 365 天的股票长历史回补当前只允许 FULL 运行类型";
  }
  const stages = normalizeBackfillStageList(payload?.stages);
  if (stages.length > 0 && !stages.includes("QUOTES")) {
    return "长历史股票回补的阶段范围必须包含 QUOTES";
  }
  return "";
}

export function marketBackfillStatusTagType(value) {
  const key = String(value || "").trim().toUpperCase();
  if (key === "FAILED") return "danger";
  if (key === "PARTIAL_SUCCESS" || key === "RUNNING" || key === "PENDING") return "warning";
  if (key === "SUCCESS") return "success";
  return "info";
}

export function marketBackfillDetailStatusTagType(value) {
  const key = String(value || "").trim().toUpperCase();
  if (key === "FAILED") return "danger";
  if (key === "RUNNING" || key === "PENDING") return "warning";
  if (key === "SUCCESS") return "success";
  if (key === "SKIPPED") return "info";
  return "info";
}

export function buildMarketCoverageOverviewCards(summary = {}) {
  const latestTradeDate = String(summary?.latest_trade_date || "").trim() || "-";
  return [
    {
      key: "universe",
      title: "证券全集",
      value: String(Number(summary?.total_universe_count) || 0),
      tone: "primary",
      helper: "当前快照里的可跟踪标的总数"
    },
    {
      key: "quotes",
      title: "行情覆盖",
      value: String(Number(summary?.quotes_coverage_count) || 0),
      tone: "success",
      helper: "已落到 truth 的行情标的数"
    },
    {
      key: "enhancement",
      title: "增强因子覆盖",
      value: `${Number(summary?.daily_basic_coverage_count) || 0} / ${Number(summary?.moneyflow_coverage_count) || 0}`,
      tone: "warning",
      helper: "分别对应 daily_basic / moneyflow 覆盖数"
    },
    {
      key: "gaps",
      title: "主数据缺口",
      value: `${Number(summary?.canonical_key_gap_count) || 0} / ${Number(summary?.display_name_gap_count) || 0} / ${Number(summary?.list_date_gap_count) || 0}`,
      tone: "danger",
      helper: "分别对应 canonical key / 名称 / 上市日期缺口"
    },
    {
      key: "latest_trade_date",
      title: "最新交易日",
      value: latestTradeDate,
      tone: "gold",
      helper: "用于判断行情覆盖是否更新到最近交易日"
    }
  ];
}

export function buildMarketCoverageAssetRows(summary = {}) {
  return (Array.isArray(summary?.asset_items) ? summary.asset_items : [])
    .map((item) => ({
      key: String(item?.asset_type || "").trim().toUpperCase(),
      label: formatMarketAssetLabel(item?.asset_type),
      universeCount: Number(item?.universe_count) || 0,
      masterCoverageCount: Number(item?.master_coverage_count) || 0,
      quotesCoverageCount: Number(item?.quotes_coverage_count) || 0,
      dailyBasicCoverageCount: Number(item?.daily_basic_coverage_count) || 0,
      moneyflowCoverageCount: Number(item?.moneyflow_coverage_count) || 0,
      latestTradeDate: String(item?.latest_trade_date || "").trim() || "-"
    }))
    .sort((left, right) => {
      const stockBias = left.key === "STOCK" ? -1 : right.key === "STOCK" ? 1 : 0;
      if (stockBias !== 0) {
        return stockBias;
      }
      return left.key.localeCompare(right.key, "zh-Hans-CN");
    });
}

export function buildUniverseSnapshotDigest(snapshot = {}) {
  const snapshotDate = String(snapshot?.snapshot_date || "").trim() || "-";
  const sourceKey = String(snapshot?.source_key || "").trim().toUpperCase() || "-";
  return `${snapshotDate} · ${formatMarketAssetScopeSummary(snapshot?.scope)} · ${sourceKey}`;
}

export function normalizeMarketCenterTab(value) {
  const normalized = String(value || "").trim().toLowerCase();
  if (normalized === "futures") return "futures";
  if (normalized === "engine-config") return "engine-config";
  return "stocks";
}

export function normalizeMarketCenterView(value, hasPublishID = false) {
  const normalized = String(value || "").trim().toLowerCase();
  if (normalized === "replay") return "replay";
  if (normalized === "detail") return "detail";
  return hasPublishID ? "detail" : "";
}

export function normalizeMarketCenterJobType(value) {
  const normalized = String(value || "").trim().toUpperCase();
  if (!normalized) {
    return "";
  }
  if (normalized.includes("FUTURES")) {
    return "FUTURES_SELECTION";
  }
  if (normalized.includes("STOCK")) {
    return "STOCK_SELECTION";
  }
  return normalized;
}

export function normalizeMarketCenterConfigType(value) {
  const normalized = String(value || "").trim().toUpperCase().replace(/[-\s]+/g, "_");
  if (!normalized) {
    return "";
  }
  if (normalized === "SEED_SET" || normalized === "STRATEGY_SEED_SET") {
    return "seed-set";
  }
  if (normalized === "AGENT_PROFILE" || normalized === "STRATEGY_AGENT_PROFILE") {
    return "agent-profile";
  }
  if (normalized === "SCENARIO_TEMPLATE" || normalized === "STRATEGY_SCENARIO_TEMPLATE") {
    return "scenario-template";
  }
  if (normalized === "PUBLISH_POLICY" || normalized === "STRATEGY_PUBLISH_POLICY") {
    return "publish-policy";
  }
  return "";
}

export function buildMarketCenterRouteQuery(filters = {}) {
  const query = {
    tab: normalizeMarketCenterTab(filters?.tab)
  };
  const issueCode = String(filters?.issue_code || "").trim().toUpperCase();
  const qualityHours = normalizeMarketQualityLookbackHours(
    filters?.quality_hours ?? filters?.hours,
    0
  );
  const publishID = String(filters?.publish_id || "").trim();
  const policyID = String(filters?.policy_id || "").trim();
  const configType = normalizeMarketCenterConfigType(filters?.config_type);
  const configID = String(filters?.config_id || "").trim() || (configType === "publish-policy" ? policyID : "");
  const jobType = normalizeMarketCenterJobType(filters?.job_type);
  const view = normalizeMarketCenterView(filters?.view, Boolean(publishID));
  if (qualityHours > 0) {
    query.quality_hours = String(qualityHours);
  }
  if (issueCode) {
    query.issue_code = issueCode;
  }
  if (publishID) {
    query.publish_id = publishID;
  }
  if (view) {
    query.view = view;
  }
  if (jobType) {
    query.job_type = jobType;
  }
  if (policyID) {
    query.policy_id = policyID;
  }
  if (configType) {
    query.config_type = configType;
  }
  if (configID) {
    query.config_id = configID;
  }
  if (!query.policy_id && configType === "publish-policy" && configID) {
    query.policy_id = configID;
  }
  return query;
}

export function normalizeMarketCenterRouteState(query = {}) {
  const publishID = String(query?.publish_id || "").trim();
  const policyID = String(query?.policy_id || "").trim();
  const configType = normalizeMarketCenterConfigType(query?.config_type || (policyID ? "publish-policy" : ""));
  const configID = String(query?.config_id || "").trim() || (configType === "publish-policy" ? policyID : "");
  return {
    tab: normalizeMarketCenterTab(query?.tab),
    quality_hours: normalizeMarketQualityLookbackHours(query?.quality_hours),
    issue_code: String(query?.issue_code || "").trim().toUpperCase(),
    publish_id: publishID,
    view: normalizeMarketCenterView(query?.view, Boolean(publishID)),
    job_type: normalizeMarketCenterJobType(query?.job_type),
    policy_id: policyID,
    config_type: configType,
    config_id: configID
  };
}

export function normalizeMarketQualityRouteContext(query = {}) {
  const returnTo = String(query?.return_to || "").trim().toLowerCase();
  if (returnTo !== "market-center") {
    return {};
  }
  return {
    return_to: "market-center",
    return_tab: normalizeMarketCenterTab(query?.return_tab)
  };
}

export function formatMarketCenterReturnLabel(tab) {
  if (normalizeMarketCenterTab(tab) === "futures") {
    return "返回期货行情中心";
  }
  return "返回股票行情中心";
}

export function buildMarketCenterQualityReturnQuery(item = {}, hours) {
  const assetClass = String(item?.asset_class || "").trim().toUpperCase();
  return buildMarketCenterRouteQuery({
    tab: assetClass === "FUTURES" ? "futures" : "stocks",
    quality_hours: hours,
    issue_code: item?.issue_code || ""
  });
}

export function buildMarketCenterQualitySummaryReturnQuery(assetClass, summary = {}, hours, mode = "latest") {
  const normalizedAssetClass = String(assetClass || "").trim().toUpperCase();
  const issueCode =
    mode === "error"
      ? String(summary?.latest_error_issue_code || "").trim()
      : String(summary?.latest_issue_code || "").trim();
  return buildMarketCenterRouteQuery({
    tab: normalizedAssetClass === "FUTURES" ? "futures" : "stocks",
    quality_hours: hours,
    issue_code: issueCode
  });
}

export function buildMarketQualityDrillQuery(filters = {}) {
  return {
    return_to: "market-center",
    return_tab: normalizeMarketCenterTab(filters?.return_tab),
    ...buildMarketQualityRouteQuery(filters)
  };
}

export function marketQualitySeverityTagType(severity) {
  const normalized = String(severity || "").trim().toUpperCase();
  if (normalized === "ERROR") return "danger";
  if (normalized === "WARN") return "warning";
  if (normalized === "INFO") return "success";
  return "info";
}

export function formatTruthRebuildSuccessMessage(assetClass, result = {}) {
  const normalized = String(assetClass || result?.asset_class || "").trim().toUpperCase();
  const truthBarCount = Number(result?.truth_bar_count) || 0;
  if (normalized === "FUTURES") {
    return `期货主力映射重建完成，处理 ${truthBarCount} 条 truth 行情`;
  }
  return `股票状态 truth 重建完成，处理 ${truthBarCount} 条 truth 行情`;
}

export function formatMarketQualityPayload(payload) {
  const raw = String(payload || "").trim();
  if (!raw) {
    return "";
  }
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

export function collectMarketQualityIssueOptions(items, limit = 8) {
  const counts = new Map();
  (Array.isArray(items) ? items : []).forEach((item) => {
    const code = String(item?.issue_code || "").trim();
    if (!code) {
      return;
    }
    counts.set(code, (counts.get(code) || 0) + 1);
  });
  return Array.from(counts.entries())
    .sort((a, b) => {
      if (b[1] !== a[1]) {
        return b[1] - a[1];
      }
      return a[0].localeCompare(b[0]);
    })
    .slice(0, Math.max(0, Number(limit) || 0))
    .map(([value, count]) => ({
      value,
      label: `${value} (${count})`
    }));
}

export function buildStockGovernanceSummaryItems(summary = {}) {
  return [
    { key: "stock_master_coverage", label: "主数据覆盖", value: Number(summary?.stock_master_coverage) || 0 },
    { key: "stock_truth_coverage", label: "truth 覆盖", value: Number(summary?.stock_truth_coverage) || 0 },
    { key: "stock_daily_basic_coverage", label: "daily_basic 覆盖", value: Number(summary?.stock_daily_basic_coverage) || 0 },
    { key: "stock_moneyflow_coverage", label: "moneyflow 覆盖", value: Number(summary?.stock_moneyflow_coverage) || 0 },
    { key: "stock_news_coverage", label: "公告资讯覆盖", value: Number(summary?.stock_news_coverage) || 0 },
    { key: "canonical_key_gap_count", label: "canonical 缺口", value: Number(summary?.canonical_key_gap_count) || 0 },
    { key: "display_name_missing_count", label: "名称缺口", value: Number(summary?.display_name_missing_count) || 0 },
    { key: "list_date_missing_count", label: "上市日缺口", value: Number(summary?.list_date_missing_count) || 0 }
  ];
}

export function formatStockFallbackSourceSummary(summary = {}) {
  const raw = String(summary?.fallback_source_summary || "").trim();
  return raw || "暂无回填摘要";
}

export function buildMarketQualityDialogTitle(item = {}) {
  const assetClass = String(item?.asset_class || "").trim().toUpperCase();
  const issueCode = String(item?.issue_code || "").trim().toUpperCase();
  const parts = ["质量日志详情"];
  if (assetClass) {
    parts.push(assetClass);
  }
  if (issueCode) {
    parts.push(issueCode);
  }
  return parts.join(" · ");
}

export function buildMarketQualityRouteQuery(filters = {}) {
  const query = {};
  const assetClass = String(filters?.asset_class || "").trim().toUpperCase();
  const dataKind = String(filters?.data_kind || "").trim().toUpperCase();
  const severity = String(filters?.severity || "").trim().toUpperCase();
  const issueCode = String(filters?.issue_code || "").trim().toUpperCase();
  const hours = normalizeMarketQualityLookbackHours(filters?.hours, 0);

  if (assetClass) query.asset_class = assetClass;
  if (dataKind) query.data_kind = dataKind;
  if (severity) query.severity = severity;
  if (issueCode) query.issue_code = issueCode;
  if (hours > 0) {
    query.hours = String(hours);
  }
  return query;
}

export function normalizeMarketQualityRouteFilters(query = {}) {
  return {
    asset_class: String(query?.asset_class || "").trim().toUpperCase(),
    data_kind: String(query?.data_kind || "").trim().toUpperCase(),
    severity: String(query?.severity || "").trim().toUpperCase(),
    issue_code: String(query?.issue_code || "").trim().toUpperCase(),
    hours: normalizeMarketQualityLookbackHours(query?.hours)
  };
}

export function buildMarketQualityBucketSummary(items, resolver, limit = 5) {
  const counts = new Map();
  (Array.isArray(items) ? items : []).forEach((item) => {
    const rawValue = typeof resolver === "function" ? resolver(item) : "";
    const value = String(rawValue || "").trim().toUpperCase();
    if (!value) {
      return;
    }
    counts.set(value, (counts.get(value) || 0) + 1);
  });
  return Array.from(counts.entries())
    .sort((a, b) => {
      if (b[1] !== a[1]) {
        return b[1] - a[1];
      }
      return a[0].localeCompare(b[0]);
    })
    .slice(0, Math.max(0, Number(limit) || 0))
    .map(([value, count]) => ({
      value,
      count,
      label: value
    }));
}

export function areMarketQualityFiltersEqual(left = {}, right = {}) {
  const normalizedLeft = normalizeMarketQualityRouteFilters(left);
  const normalizedRight = normalizeMarketQualityRouteFilters(right);
  return (
    normalizedLeft.asset_class === normalizedRight.asset_class &&
    normalizedLeft.data_kind === normalizedRight.data_kind &&
    normalizedLeft.severity === normalizedRight.severity &&
    normalizedLeft.issue_code === normalizedRight.issue_code &&
    normalizedLeft.hours === normalizedRight.hours
  );
}

export function marketGovernanceSuggestionTagType(value) {
  const normalized = String(value || "").trim();
  if (normalized.includes("降级")) return "danger";
  if (normalized.includes("补同步")) return "warning";
  if (normalized.includes("主源")) return "success";
  return "info";
}

export function buildGovernanceProviderLabel(item = {}) {
  const providerKey = String(item?.provider_key || "").trim().toUpperCase();
  const providerName = String(item?.provider_name || "").trim();
  if (providerName) {
    return `${providerKey} · ${providerName}`;
  }
  return providerKey || "-";
}
