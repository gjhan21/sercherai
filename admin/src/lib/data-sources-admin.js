export const STOCK_DEFAULT_SOURCE_CONFIG_KEY = "stock.quotes.default_source_key";
export const STOCK_DEFAULT_SOURCE_FALLBACK = "TUSHARE";
export const FUTURES_DEFAULT_SOURCE_CONFIG_KEY = "futures.quotes.default_source_key";
export const FUTURES_DEFAULT_SOURCE_FALLBACK = "TUSHARE";
export const MARKET_NEWS_DEFAULT_SOURCE_CONFIG_KEY = "market.news.default_source_key";
export const MARKET_NEWS_DEFAULT_SOURCE_FALLBACK = "AKSHARE";
export const FUTURES_INVENTORY_SOURCE_FALLBACK = "TUSHARE";

export const sourceTypeOptions = ["MARKET", "NEWS", "STOCK", "FUTURES", "SYSTEM"];
export const statusOptions = ["ACTIVE", "DISABLED"];
export const assetClassOptions = [
  { label: "全部资产", value: "" },
  { label: "股票", value: "STOCK" },
  { label: "期货", value: "FUTURES" }
];
export const dataSourcesSectionItems = [
  { key: "governance", label: "治理", description: "供应商能力、路由策略与概览" },
  { key: "sync", label: "同步", description: "股票、期货、仓单与资讯同步控制台" },
  { key: "coverage", label: "覆盖", description: "覆盖摘要、Universe 快照与回填任务" },
  { key: "truth", label: "Truth", description: "本地派生真相源重建与摘要" },
  { key: "registry", label: "数据源", description: "数据源列表、默认源与健康日志" },
  { key: "quality", label: "质量", description: "质量日志、分布与 payload 查看" }
];
export const qualityDataKindOptions = [
  { label: "全部类型", value: "" },
  { label: "日行情", value: "DAILY_BARS" },
  { label: "主数据", value: "INSTRUMENT_MASTER" },
  { label: "期货库存", value: "FUTURES_INVENTORY" }
];
export const severityFilterOptions = [
  { label: "全部级别", value: "" },
  { label: "信息", value: "INFO" },
  { label: "警告", value: "WARN" },
  { label: "错误", value: "ERROR" }
];

const BUILTIN_PROVIDER_KEYS = new Set(["TUSHARE", "AKSHARE", "TICKERMD", "MOCK", "MYSELF"]);

export function defaultDataSourceForm() {
  return {
    source_key: "",
    name: "",
    source_type: "MARKET",
    status: "ACTIVE",
    endpoint: "",
    token: "",
    fail_threshold: 3,
    retry_times: 0,
    retry_interval_ms: 200,
    health_timeout_ms: 3000,
    alert_receiver_id: "admin_001"
  };
}

export function defaultRoutingPolicyForm() {
  return {
    policy_key: "",
    asset_class: "",
    data_kind: "",
    primary_provider_key: "",
    fallback_provider_keys: [],
    fallback_allowed: true,
    mock_allowed: false,
    quality_threshold: 0.8
  };
}

export function defaultRebuildForm() {
  return {
    trade_date: "",
    days: 3
  };
}

export function defaultStockSyncForm() {
  return {
    source_key: "",
    symbols: "",
    days: 120
  };
}

export function defaultFuturesSyncForm() {
  return {
    source_key: "",
    contracts: "",
    days: 120
  };
}

export function defaultFuturesInventorySyncForm() {
  return {
    source_key: "",
    symbols: "",
    days: 30
  };
}

export function defaultMarketNewsSyncForm() {
  return {
    source_key: "",
    symbols: "",
    days: 3,
    limit: 50
  };
}

export function cloneConfigMap(value) {
  return JSON.parse(JSON.stringify(value || {}));
}

export function toSafeInt(value, fallback) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? Math.trunc(parsed) : fallback;
}

export function normalizeSourceKey(value) {
  return String(value || "")
    .trim()
    .toUpperCase();
}

export function resolveBuiltinProviderKey(value) {
  const normalized = normalizeSourceKey(value);
  return BUILTIN_PROVIDER_KEYS.has(normalized) ? normalized : "";
}

export function buildDataSourcesSectionRouteName(sectionKey) {
  const normalized = String(sectionKey || "")
    .trim()
    .toLowerCase();
  return normalized ? `data-sources-${normalized}` : "data-sources-governance";
}

export function resolveDataSourcesSectionRoute(query = {}) {
  const action = String(query?.action || "")
    .trim()
    .toLowerCase();
  if (action === "logs" || query?.source_key) {
    return "data-sources-registry";
  }
  if (query?.asset_class || query?.data_kind || query?.severity || query?.issue_code || query?.hours) {
    return "data-sources-quality";
  }
  return "data-sources-governance";
}

export function normalizeProviderKeyList(items) {
  return Array.from(
    new Set(
      (Array.isArray(items) ? items : [])
        .map((item) => normalizeSourceKey(item))
        .filter(Boolean)
    )
  );
}

export function resolveDataSourceProvider(item) {
  const explicitProvider = String(item?.config?.provider || item?.config?.vendor || "")
    .trim()
    .toUpperCase();
  if (explicitProvider) {
    return explicitProvider;
  }
  return resolveBuiltinProviderKey(item?.source_key);
}

export function supportsDefaultStockSource(item) {
  const sourceType = String(item?.source_type || "")
    .trim()
    .toUpperCase();
  if (sourceType === "STOCK") {
    return true;
  }
  const provider = resolveDataSourceProvider(item);
  return ["TUSHARE", "AKSHARE", "TICKERMD", "MOCK", "MYSELF"].includes(provider);
}

export function supportsSyncKind(item, kind) {
  const provider = resolveDataSourceProvider(item);
  switch (kind) {
    case "stock_quotes":
    case "futures_quotes":
      return ["TUSHARE", "AKSHARE", "TICKERMD", "MOCK", "MYSELF"].includes(provider);
    case "futures_inventory":
      return ["TUSHARE", "MOCK"].includes(provider);
    case "market_news":
      return ["AKSHARE", "TUSHARE"].includes(provider);
    default:
      return false;
  }
}

export function buildSyncOptionLabel(item, healthMap = {}) {
  const sourceKey = normalizeSourceKey(item?.source_key);
  const name = String(item?.name || "").trim();
  const provider = resolveDataSourceProvider(item);
  const status = String(item?.status || "").trim().toUpperCase();
  const healthItem = healthMap[sourceKey];
  const healthStatus = String(healthItem?.status || "").trim().toUpperCase();
  const tags = [provider, status, healthStatus].filter(Boolean);
  return `${sourceKey}${name ? ` · ${name}` : ""}${tags.length ? ` (${tags.join("/")})` : ""}`;
}

export function buildSyncSourceOptions(items, healthMap, kind) {
  return (Array.isArray(items) ? items : [])
    .filter((item) => supportsSyncKind(item, kind))
    .map((item) => ({
      value: normalizeSourceKey(item?.source_key),
      label: buildSyncOptionLabel(item, healthMap),
      disabled: String(item?.status || "").trim().toUpperCase() !== "ACTIVE"
    }))
    .filter((item) => item.value);
}

export function ensurePreferredSource(form, options, preferredValue) {
  const current = normalizeSourceKey(form.source_key);
  const values = options.map((item) => item.value);
  if (current && (values.includes(current) || /[,;| ]/.test(current))) {
    form.source_key = current;
    return;
  }
  const preferred = normalizeSourceKey(preferredValue);
  form.source_key =
    options.find((item) => item.value === preferred)?.value || options[0]?.value || preferred;
}

export function splitSyncInput(value) {
  return String(value || "")
    .split(/[\s,;\n]+/)
    .map((item) => item.trim().toUpperCase())
    .filter(Boolean);
}

export function buildSyncMetricTags(result) {
  const items = [];
  const truthCount = Number(result?.truth_count) || 0;
  const barCount = Number(result?.bar_count) || 0;
  const newsCount = Number(result?.news_count) || 0;
  const inventoryCount = Number(result?.inventory_count) || 0;
  const snapshotCount = Number(result?.snapshot_count) || 0;
  if (truthCount > 0) items.push({ key: "truth", label: `真相源 ${truthCount}`, type: "success" });
  if (barCount > 0) items.push({ key: "bars", label: `原始K线 ${barCount}`, type: "info" });
  if (newsCount > 0) items.push({ key: "news", label: `资讯 ${newsCount}`, type: "success" });
  if (inventoryCount > 0) items.push({ key: "inventory", label: `仓单 ${inventoryCount}`, type: "success" });
  if (snapshotCount > 0) items.push({ key: "snapshots", label: `快照 ${snapshotCount}`, type: "warning" });
  return items;
}

export function formatSyncResolvedSourceKeys(result, fallbackSourceKey = "") {
  const values = Array.isArray(result?.resolved_source_keys)
    ? result.resolved_source_keys.map((item) => normalizeSourceKey(item)).filter(Boolean)
    : [];
  return values.length > 0 ? values.join(" -> ") : normalizeSourceKey(fallbackSourceKey) || "-";
}

export function formatRequestedSourceLabel(payload) {
  return normalizeSourceKey(payload?.requested_source_key) || "DEFAULT";
}

export function buildLastSyncResult(data) {
  const result = data?.result || null;
  const symbols = Array.isArray(data?.symbols) ? data.symbols : [];
  const contracts = Array.isArray(data?.contracts) ? data.contracts : symbols;
  const countCandidates = [
    data?.count,
    result?.truth_count,
    result?.bar_count,
    result?.news_count,
    result?.inventory_count,
    result?.snapshot_count
  ];
  let count = 0;
  for (const raw of countCandidates) {
    if (raw === undefined || raw === null || raw === "") {
      continue;
    }
    const parsed = Number(raw);
    count = Number.isFinite(parsed) ? parsed : 0;
    break;
  }
  return {
    count,
    source_key: normalizeSourceKey(data?.source_key || result?.selected_source || result?.requested_source_key),
    requested_source_key: normalizeSourceKey(data?.requested_source_key || result?.requested_source_key),
    days: Number(data?.days) || 0,
    limit: Number(data?.limit) || 0,
    symbols,
    contracts,
    result
  };
}

export function formatSyncRequestScope(items, emptyLabel = "默认池") {
  const list = Array.isArray(items) ? items.filter(Boolean) : [];
  if (list.length === 0) {
    return emptyLabel;
  }
  const preview = list.slice(0, 4).join("、");
  if (list.length > 4) {
    return `${preview} 等 ${list.length} 个`;
  }
  return preview;
}

export function statusTagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "ACTIVE" || normalized === "HEALTHY" || normalized === "SUCCESS") return "success";
  if (normalized === "DISABLED" || normalized === "UNHEALTHY" || normalized === "FAILED") return "danger";
  if (normalized === "PENDING" || normalized === "RUNNING") return "warning";
  return "info";
}

export function isDefaultSource(sourceKey, defaultKey) {
  return normalizeSourceKey(sourceKey) === normalizeSourceKey(defaultKey);
}

export function formatGovernanceSuggestion(item) {
  return String(item?.governance_suggestion || "").trim() || "暂无治理建议";
}

export function formatRoutingFallbackProviders(item) {
  return Array.isArray(item?.fallback_provider_keys) && item.fallback_provider_keys.length
    ? item.fallback_provider_keys.join(" / ")
    : "-";
}

export function formatSummaryWarnings(summary) {
  return Array.isArray(summary?.warnings) && summary.warnings.length ? summary.warnings.join("；") : "-";
}

export function formatQualitySummaryLatest(summary) {
  if (!summary?.latest_issue_code) {
    return "-";
  }
  const source = summary.latest_source_key || "-";
  const severity = summary.latest_severity || "-";
  return `${source} · ${severity} · ${summary.latest_issue_code}`;
}

export function formatQualitySummaryLatestError(summary) {
  if (!summary?.latest_error_issue_code) {
    return "-";
  }
  const source = summary.latest_error_source_key || "-";
  return `${source} · ${summary.latest_error_issue_code}`;
}

export function formatFallbackSources(items = [], limit = 4) {
  const sourceItems = Array.isArray(items) ? items : [];
  if (!sourceItems.length) {
    return "暂无回退源记录";
  }
  return sourceItems
    .slice(0, Math.max(0, Number(limit) || 0))
    .map((item) => `${normalizeSourceKey(item?.source_key) || "-"} ${Number(item?.count) || 0}`)
    .join(" / ");
}
