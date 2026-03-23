<script setup>
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessageBox } from "element-plus";
import {
  batchCheckDataSources,
  checkDataSourceHealth,
  createDataSource,
  deleteDataSource,
  getMarketDataQualitySummary,
  getMarketDerivedTruthSummary,
  listDataSourceHealthLogs,
  listDataSources,
  listMarketDataQualityLogs,
  rebuildFuturesDerivedTruth,
  rebuildStockDerivedTruth,
  listSystemConfigs,
  syncFuturesInventory,
  syncFuturesQuotes,
  syncMarketNewsSource,
  syncStockQuotes,
  upsertSystemConfig,
  updateDataSource
} from "../api/admin";
import {
  DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS,
  MARKET_QUALITY_LOOKBACK_OPTIONS,
  areMarketQualityFiltersEqual,
  buildMarketQualityRouteQuery,
  buildMarketQualityBucketSummary,
  collectMarketQualityIssueOptions,
  formatMarketQualityLookbackLabel,
  formatMarketQualityPayload,
  formatTruthRebuildSuccessMessage,
  marketQualitySeverityTagType,
  normalizeMarketQualityRouteContext,
  normalizeMarketQualityLookbackHours,
  normalizeMarketQualityRouteFilters
} from "../lib/market-data-admin";
import { hasPermission } from "../lib/session";

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const submitting = ref(false);
const batchChecking = ref(false);
const settingDefaultSource = ref(false);
const settingDefaultSourceKey = ref("");
const message = ref("");
const errorMessage = ref("");

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const items = ref([]);

const healthMap = ref({});

const logsLoading = ref(false);
const logSourceKey = ref("");
const logItems = ref([]);
const qualityLogsLoading = ref(false);
const qualityLogItems = ref([]);
const qualityLogTotal = ref(0);
const qualityRouteInitialized = ref(false);
const qualityPayloadDialogVisible = ref(false);
const selectedQualityLog = ref(null);
const qualitySummaryAll = ref(null);
const qualitySummaryStock = ref(null);
const qualitySummaryFutures = ref(null);
const rebuildingStockTruth = ref(false);
const rebuildingFuturesTruth = ref(false);
const stockRebuildSummary = ref(null);
const futuresRebuildSummary = ref(null);
const syncingStockQuotes = ref(false);
const syncingFuturesQuotes = ref(false);
const syncingFuturesInventory = ref(false);
const syncingMarketNews = ref(false);
const stockLastSyncResult = ref(null);
const futuresLastSyncResult = ref(null);
const futuresInventoryLastSyncResult = ref(null);
const marketNewsLastSyncResult = ref(null);

const formVisible = ref(false);
const formMode = ref("create");
const defaultStockSourceKey = ref("TUSHARE");
const defaultFuturesSourceKey = ref("TUSHARE");
const defaultMarketNewsSourceKey = ref("AKSHARE");
const editingConfigSnapshot = ref({});
const FUTURES_DEFAULT_SOURCE_CONFIG_KEY = "futures.quotes.default_source_key";
const FUTURES_DEFAULT_SOURCE_FALLBACK = "TUSHARE";
const MARKET_NEWS_DEFAULT_SOURCE_CONFIG_KEY = "market.news.default_source_key";
const MARKET_NEWS_DEFAULT_SOURCE_FALLBACK = "AKSHARE";
const FUTURES_INVENTORY_SOURCE_FALLBACK = "TUSHARE";

const sourceTypeOptions = ["MARKET", "NEWS", "STOCK", "FUTURES", "SYSTEM"];
const statusOptions = ["ACTIVE", "DISABLED"];
const STOCK_DEFAULT_SOURCE_CONFIG_KEY = "stock.quotes.default_source_key";
const STOCK_DEFAULT_SOURCE_FALLBACK = "TUSHARE";
const canEditDataSources = hasPermission("data_source.edit");
const canEditMarket = hasPermission("market.edit");

const qualityFilters = reactive({
  asset_class: "",
  data_kind: "",
  severity: "",
  issue_code: "",
  hours: DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS
});

const stockSyncForm = reactive({
  source_key: "",
  symbols: "",
  days: 120
});

const futuresSyncForm = reactive({
  source_key: "",
  contracts: "",
  days: 120
});

const futuresInventorySyncForm = reactive({
  source_key: "",
  symbols: "",
  days: 30
});

const marketNewsSyncForm = reactive({
  source_key: "",
  symbols: "",
  days: 3,
  limit: 50
});

const rebuildForm = reactive({
  trade_date: "",
  days: 3
});

const assetClassOptions = [
  { label: "全部资产", value: "" },
  { label: "股票", value: "STOCK" },
  { label: "期货", value: "FUTURES" }
];

const qualityDataKindOptions = [
  { label: "全部类型", value: "" },
  { label: "日行情", value: "DAILY_BARS" },
  { label: "主数据", value: "INSTRUMENT_MASTER" },
  { label: "期货库存", value: "FUTURES_INVENTORY" }
];

const severityFilterOptions = [
  { label: "全部级别", value: "" },
  { label: "信息", value: "INFO" },
  { label: "警告", value: "WARN" },
  { label: "错误", value: "ERROR" }
];

const qualityIssueQuickOptions = computed(() =>
  collectMarketQualityIssueOptions(qualityLogItems.value, 8)
);
const qualitySourceBuckets = computed(() =>
  buildMarketQualityBucketSummary(qualityLogItems.value, (item) => item?.source_key, 3)
);
const qualityIssueBuckets = computed(() =>
  buildMarketQualityBucketSummary(qualityLogItems.value, (item) => item?.issue_code, 3)
);
const qualitySeverityBuckets = computed(() =>
  buildMarketQualityBucketSummary(qualityLogItems.value, (item) => item?.severity, 3)
);
const qualityLookbackLabel = computed(() => formatMarketQualityLookbackLabel(qualityFilters.hours));

const selectedQualityPayloadText = computed(() =>
  formatMarketQualityPayload(selectedQualityLog.value?.payload || "")
);

function defaultForm() {
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

const form = reactive(defaultForm());

function resetForm() {
  Object.assign(form, defaultForm());
  formMode.value = "create";
  editingConfigSnapshot.value = {};
}

function toSafeInt(value, fallback) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? Math.trunc(parsed) : fallback;
}

function resolveQualityLookbackHours(value) {
  return normalizeMarketQualityLookbackHours(value);
}

function normalizeSourceKey(value) {
  return String(value || "")
    .trim()
    .toUpperCase();
}

function resolveDataSourceProvider(item) {
  return String(item?.config?.provider || item?.config?.vendor || "")
    .trim()
    .toUpperCase();
}

function supportsDefaultStockSource(item) {
  const sourceType = String(item?.source_type || "")
    .trim()
    .toUpperCase();
  if (sourceType === "STOCK") {
    return true;
  }
  const provider = resolveDataSourceProvider(item);
  return ["TUSHARE", "AKSHARE", "TICKERMD", "MOCK", "MYSELF"].includes(provider);
}

function supportsSyncKind(item, kind) {
  const provider = resolveDataSourceProvider(item);
  switch (kind) {
    case "stock_quotes":
      return ["TUSHARE", "AKSHARE", "TICKERMD", "MOCK", "MYSELF"].includes(provider);
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

function buildSyncOptionLabel(item) {
  const sourceKey = normalizeSourceKey(item?.source_key);
  const name = String(item?.name || "").trim();
  const provider = resolveDataSourceProvider(item);
  const status = String(item?.status || "").trim().toUpperCase();
  const healthItem = healthMap.value[sourceKey];
  const healthStatus = String(healthItem?.status || "").trim().toUpperCase();
  const tags = [provider, status, healthStatus].filter(Boolean);
  return `${sourceKey}${name ? ` · ${name}` : ""}${tags.length ? ` (${tags.join("/")})` : ""}`;
}

function buildSyncSourceOptions(kind) {
  return items.value
    .filter((item) => supportsSyncKind(item, kind))
    .map((item) => ({
      value: normalizeSourceKey(item?.source_key),
      label: buildSyncOptionLabel(item),
      disabled: String(item?.status || "").trim().toUpperCase() !== "ACTIVE"
    }))
    .filter((item) => item.value);
}

const stockSyncOptions = computed(() => buildSyncSourceOptions("stock_quotes"));
const futuresSyncOptions = computed(() => buildSyncSourceOptions("futures_quotes"));
const futuresInventorySyncOptions = computed(() => buildSyncSourceOptions("futures_inventory"));
const marketNewsSyncOptions = computed(() => buildSyncSourceOptions("market_news"));

function ensurePreferredSource(form, options, preferredValue) {
  const current = normalizeSourceKey(form.source_key);
  const values = options.map((item) => item.value);
  if (current && (values.includes(current) || /[,;| ]/.test(current))) {
    form.source_key = current;
    return;
  }
  const preferred = normalizeSourceKey(preferredValue);
  form.source_key =
    options.find((item) => item.value === preferred)?.value ||
    options[0]?.value ||
    preferred;
}

function splitSyncInput(value) {
  return String(value || "")
    .split(/[\s,;\n]+/)
    .map((item) => item.trim().toUpperCase())
    .filter(Boolean);
}

function buildSyncMetricTags(result) {
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

function formatSyncResolvedSourceKeys(result, fallbackSourceKey = "") {
  const values = Array.isArray(result?.resolved_source_keys)
    ? result.resolved_source_keys.map((item) => normalizeSourceKey(item)).filter(Boolean)
    : [];
  return values.length > 0 ? values.join(" -> ") : normalizeSourceKey(fallbackSourceKey) || "-";
}

function formatRequestedSourceLabel(payload) {
  return normalizeSourceKey(payload?.requested_source_key) || "DEFAULT";
}

function buildLastSyncResult(data) {
  return {
    count: Number(data?.count) || 0,
    source_key: normalizeSourceKey(data?.source_key),
    requested_source_key: normalizeSourceKey(data?.requested_source_key),
    days: Number(data?.days) || 0,
    limit: Number(data?.limit) || 0,
    symbols: Array.isArray(data?.symbols) ? data.symbols : [],
    contracts: Array.isArray(data?.contracts) ? data.contracts : [],
    result: data?.result || null
  };
}

function formatSyncRequestScope(items, emptyLabel = "默认池") {
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

function isDefaultStockSource(sourceKey) {
  return normalizeSourceKey(sourceKey) === normalizeSourceKey(defaultStockSourceKey.value);
}

function cloneConfigMap(value) {
  return JSON.parse(JSON.stringify(value || {}));
}

function ensureCanEditDataSources() {
  if (canEditDataSources) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法修改数据源配置或执行健康检查";
  return false;
}

function ensureCanEditMarket() {
  if (canEditMarket) {
    return true;
  }
  errorMessage.value = "当前账号没有行情维护权限，无法执行 truth 派生重建";
  return false;
}

async function loadDefaultSourceKey(configKey, fallback, targetRef, options = {}) {
  const { silent = false } = options;
  try {
    const data = await listSystemConfigs({
      keyword: configKey,
      page: 1,
      page_size: 50
    });
    const rows = Array.isArray(data?.items) ? data.items : [];
    const matched = rows.find(
      (item) =>
        normalizeSourceKey(item?.config_key) === normalizeSourceKey(configKey)
    );
    const configValue = normalizeSourceKey(matched?.config_value);
    targetRef.value = configValue || fallback;
  } catch (error) {
    targetRef.value = fallback;
    if (!silent) {
      throw error;
    }
  }
}

async function loadDefaultSourceKeys(options = {}) {
  await Promise.all([
    loadDefaultSourceKey(STOCK_DEFAULT_SOURCE_CONFIG_KEY, STOCK_DEFAULT_SOURCE_FALLBACK, defaultStockSourceKey, options),
    loadDefaultSourceKey(FUTURES_DEFAULT_SOURCE_CONFIG_KEY, FUTURES_DEFAULT_SOURCE_FALLBACK, defaultFuturesSourceKey, options),
    loadDefaultSourceKey(MARKET_NEWS_DEFAULT_SOURCE_CONFIG_KEY, MARKET_NEWS_DEFAULT_SOURCE_FALLBACK, defaultMarketNewsSourceKey, options)
  ]);
}

function buildPayload() {
  const sourceKey = form.source_key.trim().toUpperCase();
  const endpointValue = form.endpoint.trim();
  const config = {
    ...(formMode.value === "edit" ? cloneConfigMap(editingConfigSnapshot.value) : {}),
    endpoint: sourceKey === "TUSHARE" && !endpointValue ? "https://api.tushare.pro" : endpointValue,
    fail_threshold: toSafeInt(form.fail_threshold, 3),
    retry_times: toSafeInt(form.retry_times, 0),
    retry_interval_ms: toSafeInt(form.retry_interval_ms, 200),
    health_timeout_ms: toSafeInt(form.health_timeout_ms, 3000)
  };
  delete config.token;
  delete config.api_token;
  delete config.tushare_token;
  if (form.alert_receiver_id.trim()) {
    config.alert_receiver_id = form.alert_receiver_id.trim();
  } else {
    delete config.alert_receiver_id;
  }
  if (form.token.trim()) {
    config.token = form.token.trim();
  }
  return {
    source_key: sourceKey,
    name: form.name.trim(),
    source_type: form.source_type.trim(),
    status: form.status,
    config
  };
}

async function fetchDataSources(options = {}) {
  const { preserveFeedback = false } = options;
  loading.value = true;
  if (!preserveFeedback) {
    errorMessage.value = "";
    message.value = "";
  }
  try {
    const [data] = await Promise.all([
      listDataSources({ page: page.value, page_size: pageSize.value }),
      loadDefaultSourceKeys({ silent: true })
    ]);
    items.value = data.items || [];
    total.value = data.total || 0;
    ensurePreferredSource(stockSyncForm, stockSyncOptions.value, defaultStockSourceKey.value);
    ensurePreferredSource(futuresSyncForm, futuresSyncOptions.value, defaultFuturesSourceKey.value);
    ensurePreferredSource(futuresInventorySyncForm, futuresInventorySyncOptions.value, FUTURES_INVENTORY_SOURCE_FALLBACK);
    ensurePreferredSource(marketNewsSyncForm, marketNewsSyncOptions.value, defaultMarketNewsSourceKey.value);
  } catch (error) {
    errorMessage.value = error.message || "加载数据源失败";
  } finally {
    loading.value = false;
  }
}

async function fetchQualityLogs(options = {}) {
  const { preserveFeedback = false } = options;
  qualityLogsLoading.value = true;
  if (!preserveFeedback) {
    errorMessage.value = "";
  }
  try {
    const data = await listMarketDataQualityLogs({
      ...qualityFilters,
      hours: resolveQualityLookbackHours(qualityFilters.hours),
      page: 1,
      page_size: 20
    });
    qualityLogItems.value = data.items || [];
    qualityLogTotal.value = data.total || 0;
  } catch (error) {
    qualityLogItems.value = [];
    qualityLogTotal.value = 0;
    errorMessage.value = error.message || "加载质量日志失败";
  } finally {
    qualityLogsLoading.value = false;
  }
}

function applyRouteQualityFilters() {
  Object.assign(qualityFilters, normalizeMarketQualityRouteFilters(route.query));
}

function syncQualityFiltersToRoute() {
  const nextQuery = {
    ...normalizeMarketQualityRouteContext(route.query),
    ...buildMarketQualityRouteQuery(qualityFilters)
  };
  if (areMarketQualityFiltersEqual(route.query, nextQuery)) {
    return false;
  }
  router.replace({
    name: "data-sources",
    query: nextQuery
  });
  return true;
}

function applyQualityIssueQuickFilter(issueCode) {
  qualityFilters.issue_code = String(issueCode || "").trim();
  if (!syncQualityFiltersToRoute()) {
    fetchQualityLogs({ preserveFeedback: true });
  }
}

function openQualityPayload(row) {
  selectedQualityLog.value = row || null;
  qualityPayloadDialogVisible.value = true;
}

async function submitForm() {
  if (!ensureCanEditDataSources()) {
    return;
  }
  submitting.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const payload = buildPayload();
    if (!payload.name || !payload.source_type) {
      throw new Error("请完整填写必填字段");
    }
    if (formMode.value === "create") {
      if (!payload.source_key) {
        throw new Error("source_key 不能为空");
      }
      await createDataSource(payload);
      message.value = `数据源 ${payload.source_key} 创建成功`;
    } else {
      await updateDataSource(payload.source_key, {
        name: payload.name,
        source_type: payload.source_type,
        status: payload.status,
        config: payload.config
      });
      message.value = `数据源 ${payload.source_key} 更新成功`;
    }
    formVisible.value = false;
    resetForm();
    await fetchDataSources({ preserveFeedback: true });
  } catch (error) {
    errorMessage.value = error.message || "提交失败";
  } finally {
    submitting.value = false;
  }
}

function handleCreate() {
  if (!ensureCanEditDataSources()) {
    return;
  }
  resetForm();
  formVisible.value = true;
}

function handleEdit(item) {
  if (!ensureCanEditDataSources()) {
    return;
  }
  const cfg = item.config || {};
  editingConfigSnapshot.value = cloneConfigMap(cfg);
  Object.assign(form, {
    source_key: item.source_key || "",
    name: item.name || "",
    source_type: item.source_type || "MARKET",
    status: item.status || "ACTIVE",
    endpoint: cfg.endpoint || "",
    token: cfg.token || cfg.api_token || cfg.tushare_token || "",
    fail_threshold: toSafeInt(cfg.fail_threshold, 3),
    retry_times: toSafeInt(cfg.retry_times, 0),
    retry_interval_ms: toSafeInt(cfg.retry_interval_ms, 200),
    health_timeout_ms: toSafeInt(cfg.health_timeout_ms, 3000),
    alert_receiver_id: cfg.alert_receiver_id || "admin_001"
  });
  formMode.value = "edit";
  formVisible.value = true;
}

async function handleDelete(sourceKey) {
  if (!ensureCanEditDataSources()) {
    return;
  }
  try {
    await ElMessageBox.confirm(`确认删除数据源 ${sourceKey}？`, "删除确认", {
      type: "warning",
      confirmButtonText: "删除",
      cancelButtonText: "取消"
    });
  } catch {
    return;
  }

  errorMessage.value = "";
  message.value = "";
  try {
    await deleteDataSource(sourceKey);
    message.value = `数据源 ${sourceKey} 已删除`;
    await fetchDataSources({ preserveFeedback: true });
  } catch (error) {
    errorMessage.value = error.message || "删除失败";
  }
}

async function handleSetDefaultStockSource(row) {
  if (!ensureCanEditDataSources()) {
    return;
  }
  if (!supportsDefaultStockSource(row)) {
    errorMessage.value = "仅支持设置可用于股票行情同步的数据源为默认行情源";
    return;
  }
  const sourceKey = normalizeSourceKey(row.source_key);
  if (!sourceKey) {
    errorMessage.value = "source_key 不能为空";
    return;
  }
  settingDefaultSource.value = true;
  settingDefaultSourceKey.value = sourceKey;
  errorMessage.value = "";
  message.value = "";
  try {
    await upsertSystemConfig({
      config_key: STOCK_DEFAULT_SOURCE_CONFIG_KEY,
      config_value: sourceKey,
      description: "股票行情默认数据源"
    });
    defaultStockSourceKey.value = sourceKey;
    message.value = `默认行情源已设置为 ${sourceKey}`;
  } catch (error) {
    errorMessage.value = error.message || "设置默认行情源失败";
  } finally {
    settingDefaultSource.value = false;
    settingDefaultSourceKey.value = "";
  }
}

async function handleCheckOne(sourceKey) {
  if (!ensureCanEditDataSources()) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await checkDataSourceHealth(sourceKey);
    healthMap.value[sourceKey] = result;
    message.value = `数据源 ${sourceKey} 健康检查完成`;
  } catch (error) {
    errorMessage.value = error.message || "健康检查失败";
  }
}

async function handleBatchCheckAll() {
  if (!ensureCanEditDataSources()) {
    return;
  }
  batchChecking.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await batchCheckDataSources([]);
    const rows = result.items || [];
    const merged = { ...healthMap.value };
    rows.forEach((row) => {
      merged[row.source_key] = row;
    });
    healthMap.value = merged;
    message.value = `批量健康检查完成，共 ${rows.length} 个数据源`;
  } catch (error) {
    errorMessage.value = error.message || "批量健康检查失败";
  } finally {
    batchChecking.value = false;
  }
}

async function showLogs(sourceKey) {
  logsLoading.value = true;
  logSourceKey.value = sourceKey;
  errorMessage.value = "";
  try {
    const result = await listDataSourceHealthLogs(sourceKey, {
      page: 1,
      page_size: 20
    });
    logItems.value = result.items || [];
  } catch (error) {
    logItems.value = [];
    errorMessage.value = error.message || "加载健康日志失败";
  } finally {
    logsLoading.value = false;
  }
}

function normalizeRebuildPayload() {
  return {
    trade_date: String(rebuildForm.trade_date || "").trim(),
    days: toSafeInt(rebuildForm.days, 3)
  };
}

function getRebuildSummaryRef(assetClass) {
  return assetClass === "FUTURES" ? futuresRebuildSummary : stockRebuildSummary;
}

function getQualitySummaryRef(assetClass) {
  if (assetClass === "STOCK") return qualitySummaryStock;
  if (assetClass === "FUTURES") return qualitySummaryFutures;
  return qualitySummaryAll;
}

function formatSummaryWarnings(summary) {
  return Array.isArray(summary?.warnings) && summary.warnings.length ? summary.warnings.join("；") : "-";
}

function formatQualitySummaryLatest(summary) {
  if (!summary?.latest_issue_code) {
    return "-";
  }
  const source = summary.latest_source_key || "-";
  const severity = summary.latest_severity || "-";
  return `${source} · ${severity} · ${summary.latest_issue_code}`;
}

function formatQualitySummaryLatestError(summary) {
  if (!summary?.latest_error_issue_code) {
    return "-";
  }
  const source = summary.latest_error_source_key || "-";
  return `${source} · ${summary.latest_error_issue_code}`;
}

async function fetchQualitySummary(assetClass, options = {}) {
  const { preserveFeedback = false } = options;
  const normalizedAssetClass = normalizeSourceKey(assetClass);
  const summaryRef = getQualitySummaryRef(normalizedAssetClass);
  if (!preserveFeedback) {
    errorMessage.value = "";
  }
  try {
    const params = { hours: resolveQualityLookbackHours(qualityFilters.hours) };
    if (normalizedAssetClass) {
      params.asset_class = normalizedAssetClass;
    }
    const data = await getMarketDataQualitySummary(params);
    summaryRef.value = data || null;
  } catch (error) {
    summaryRef.value = null;
    if (!preserveFeedback) {
      errorMessage.value = error.message || "加载质量概览失败";
    }
  }
}

async function fetchQualitySummaries(options = {}) {
  await Promise.all([
    fetchQualitySummary("", options),
    fetchQualitySummary("STOCK", options),
    fetchQualitySummary("FUTURES", options)
  ]);
}

async function fetchQualityDashboard(options = {}) {
  await Promise.all([
    fetchQualityLogs(options),
    fetchQualitySummaries(options)
  ]);
}

async function fetchDerivedTruthSummary(assetClass, options = {}) {
  const { preserveFeedback = false, preserveCurrent = false } = options;
  const normalizedAssetClass = normalizeSourceKey(assetClass);
  const summaryRef = getRebuildSummaryRef(normalizedAssetClass);
  if (!preserveFeedback) {
    errorMessage.value = "";
  }
  try {
    const data = await getMarketDerivedTruthSummary({ asset_class: normalizedAssetClass });
    summaryRef.value = data || null;
  } catch (error) {
    if (!preserveCurrent) {
      summaryRef.value = null;
    }
    if (!preserveFeedback) {
      errorMessage.value = error.message || "加载 truth 重建摘要失败";
    }
  }
}

async function fetchDerivedTruthSummaries(options = {}) {
  await Promise.all([
    fetchDerivedTruthSummary("STOCK", options),
    fetchDerivedTruthSummary("FUTURES", options)
  ]);
}

async function handleRebuildTruth(assetClass) {
  if (!ensureCanEditMarket()) {
    return;
  }
  const isStock = assetClass === "STOCK";
  const loadingRef = isStock ? rebuildingStockTruth : rebuildingFuturesTruth;
  const summaryRef = getRebuildSummaryRef(assetClass);
  loadingRef.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const data = isStock
      ? await rebuildStockDerivedTruth(normalizeRebuildPayload())
      : await rebuildFuturesDerivedTruth(normalizeRebuildPayload());
    summaryRef.value = data;
    message.value = formatTruthRebuildSuccessMessage(isStock ? "STOCK" : "FUTURES", data);
    await Promise.all([
      fetchQualityDashboard({ preserveFeedback: true }),
      fetchDerivedTruthSummary(assetClass, { preserveFeedback: true, preserveCurrent: true })
    ]);
  } catch (error) {
    errorMessage.value = error.message || `${isStock ? "股票" : "期货"} truth 重建失败`;
  } finally {
    loadingRef.value = false;
  }
}

async function handleSyncStockQuotes() {
  if (!ensureCanEditMarket()) {
    return;
  }
  syncingStockQuotes.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const data = await syncStockQuotes({
      source_key: String(stockSyncForm.source_key || "").trim(),
      symbols: splitSyncInput(stockSyncForm.symbols),
      days: Number(stockSyncForm.days) || 120
    });
    stockLastSyncResult.value = buildLastSyncResult(data);
    message.value = `股票行情同步完成，处理 ${stockLastSyncResult.value.count || 0} 条`;
  } catch (error) {
    errorMessage.value = error.message || "股票行情同步失败";
  } finally {
    syncingStockQuotes.value = false;
  }
}

async function handleSyncFuturesQuotes() {
  if (!ensureCanEditMarket()) {
    return;
  }
  syncingFuturesQuotes.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const data = await syncFuturesQuotes({
      source_key: String(futuresSyncForm.source_key || "").trim(),
      contracts: splitSyncInput(futuresSyncForm.contracts),
      days: Number(futuresSyncForm.days) || 120
    });
    futuresLastSyncResult.value = buildLastSyncResult(data);
    message.value = `期货行情同步完成，处理 ${futuresLastSyncResult.value.count || 0} 条`;
  } catch (error) {
    errorMessage.value = error.message || "期货行情同步失败";
  } finally {
    syncingFuturesQuotes.value = false;
  }
}

async function handleSyncFuturesInventory() {
  if (!ensureCanEditMarket()) {
    return;
  }
  syncingFuturesInventory.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const data = await syncFuturesInventory({
      source_key: String(futuresInventorySyncForm.source_key || "").trim(),
      symbols: splitSyncInput(futuresInventorySyncForm.symbols),
      days: Number(futuresInventorySyncForm.days) || 30
    });
    futuresInventoryLastSyncResult.value = buildLastSyncResult(data);
    message.value = `期货仓单同步完成，处理 ${futuresInventoryLastSyncResult.value.count || 0} 条`;
  } catch (error) {
    errorMessage.value = error.message || "期货仓单同步失败";
  } finally {
    syncingFuturesInventory.value = false;
  }
}

async function handleSyncMarketNews() {
  if (!ensureCanEditMarket()) {
    return;
  }
  syncingMarketNews.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const data = await syncMarketNewsSource({
      source_key: String(marketNewsSyncForm.source_key || "").trim(),
      symbols: splitSyncInput(marketNewsSyncForm.symbols),
      days: Number(marketNewsSyncForm.days) || 3,
      limit: Number(marketNewsSyncForm.limit) || 50
    });
    marketNewsLastSyncResult.value = buildLastSyncResult(data);
    message.value = `市场资讯同步完成，处理 ${marketNewsLastSyncResult.value.count || 0} 条`;
  } catch (error) {
    errorMessage.value = error.message || "市场资讯同步失败";
  } finally {
    syncingMarketNews.value = false;
  }
}

function resetQualityFilters() {
  Object.assign(qualityFilters, {
    asset_class: "",
    data_kind: "",
    severity: "",
    issue_code: "",
    hours: DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS
  });
  if (!syncQualityFiltersToRoute()) {
    fetchQualityDashboard({ preserveFeedback: true });
  }
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "ACTIVE" || normalized === "HEALTHY") return "success";
  if (normalized === "DISABLED" || normalized === "UNHEALTHY") return "danger";
  if (normalized === "PENDING") return "warning";
  return "info";
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  fetchDataSources();
}

onMounted(fetchDataSources);
onMounted(() => {
  fetchDerivedTruthSummaries({ preserveFeedback: true });
});

watch(
  () => route.query,
  (query) => {
    const nextFilters = normalizeMarketQualityRouteFilters(query);
    if (areMarketQualityFiltersEqual(nextFilters, qualityFilters) && qualityRouteInitialized.value) {
      return;
    }
    Object.assign(qualityFilters, nextFilters);
    qualityRouteInitialized.value = true;
    fetchQualityDashboard({ preserveFeedback: true });
  },
  { immediate: true }
);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">数据源管理</h1>
        <p class="muted">统一处理数据源配置、手动同步、健康检查、本地 truth 派生重建与质量日志，不再依赖策略中心承接这些动作。</p>
      </div>
      <div class="toolbar">
        <el-tag type="warning" effect="plain">
          默认行情源：{{ defaultStockSourceKey || "-" }}
        </el-tag>
        <el-button :loading="loading" @click="fetchDataSources">刷新列表</el-button>
        <el-button
          v-if="canEditDataSources"
          type="primary"
          plain
          :loading="batchChecking"
          @click="handleBatchCheckAll"
        >
          全部健康检查
        </el-button>
        <el-button v-if="canEditDataSources" type="primary" @click="handleCreate">新增数据源</el-button>
      </div>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />
    <el-alert
      v-if="message"
      :title="message"
      type="success"
      show-icon
      style="margin-bottom: 12px"
    />

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <div>
          <h3 style="margin: 0">数据同步操作台</h3>
          <p class="muted" style="margin: 6px 0 0">
            股票、期货、市场资讯的手动同步统一收口在这里；如果上游源失败，可直接切换默认源、AUTO 或自定义回退链路。
          </p>
        </div>
        <div class="inline-actions inline-actions--left">
          <el-tag type="warning" effect="plain">股票默认源：{{ defaultStockSourceKey || "-" }}</el-tag>
          <el-tag type="warning" effect="plain">期货默认源：{{ defaultFuturesSourceKey || "-" }}</el-tag>
          <el-tag type="warning" effect="plain">资讯默认源：{{ defaultMarketNewsSourceKey || "-" }}</el-tag>
        </div>
      </div>

      <div class="truth-summary-grid" style="margin-top: 12px">
        <div class="truth-summary-card">
          <div class="truth-summary-card__header">
            <div class="truth-summary-card__title">股票行情同步</div>
            <el-tag type="info" effect="plain">
              {{ stockSyncOptions.length ? `${stockSyncOptions.length} 个可用源` : "无可用源" }}
            </el-tag>
          </div>
          <div class="sync-form-grid">
            <el-select
              v-model="stockSyncForm.source_key"
              filterable
              allow-create
              default-first-option
              placeholder="选择股票行情源"
            >
              <el-option
                v-for="item in stockSyncOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
                :disabled="item.disabled"
              />
            </el-select>
            <el-input
              v-model="stockSyncForm.symbols"
              placeholder="股票代码，逗号或换行分隔；留空使用默认股票池"
            />
            <el-input-number v-model="stockSyncForm.days" :min="20" :max="365" controls-position="right" />
            <el-button type="primary" :loading="syncingStockQuotes" @click="handleSyncStockQuotes">同步股票行情</el-button>
          </div>
          <div class="sync-inline-hint">
            <el-text type="info" size="small">支持输入 `AUTO` 或自定义链路，如 `TUSHARE,AKSHARE,TICKERMD,MYSELF`。</el-text>
          </div>
          <div v-if="stockLastSyncResult" class="sync-last-result">
            <div class="sync-last-result__head">
              <strong>最近一次股票同步</strong>
              <span>{{ formatRequestedSourceLabel(stockLastSyncResult) }} -> {{ formatSyncResolvedSourceKeys(stockLastSyncResult.result, stockLastSyncResult.source_key) }}</span>
            </div>
            <div class="sync-result-tags">
              <el-tag v-for="item in buildSyncMetricTags(stockLastSyncResult.result)" :key="`stock-sync-${item.key}`" :type="item.type">
                {{ item.label }}
              </el-tag>
              <el-tag v-if="buildSyncMetricTags(stockLastSyncResult.result).length === 0" type="info">
                处理 {{ stockLastSyncResult.count || 0 }} 条
              </el-tag>
            </div>
            <el-text type="info" size="small">
              范围：{{ formatSyncRequestScope(stockLastSyncResult.symbols, "默认股票池") }}
            </el-text>
          </div>
        </div>

        <div class="truth-summary-card">
          <div class="truth-summary-card__header">
            <div class="truth-summary-card__title">期货行情同步</div>
            <el-tag type="info" effect="plain">
              {{ futuresSyncOptions.length ? `${futuresSyncOptions.length} 个可用源` : "无可用源" }}
            </el-tag>
          </div>
          <div class="sync-form-grid">
            <el-select
              v-model="futuresSyncForm.source_key"
              filterable
              allow-create
              default-first-option
              placeholder="选择期货行情源"
            >
              <el-option
                v-for="item in futuresSyncOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
                :disabled="item.disabled"
              />
            </el-select>
            <el-input
              v-model="futuresSyncForm.contracts"
              placeholder="合约代码，逗号或换行分隔；留空使用默认合约池"
            />
            <el-input-number v-model="futuresSyncForm.days" :min="20" :max="365" controls-position="right" />
            <el-button type="primary" :loading="syncingFuturesQuotes" @click="handleSyncFuturesQuotes">同步期货行情</el-button>
          </div>
          <div class="sync-inline-hint">
            <el-text type="info" size="small">支持输入 `AUTO` 或自定义链路，如 `TUSHARE,TICKERMD,AKSHARE,MYSELF`。</el-text>
          </div>
          <div v-if="futuresLastSyncResult" class="sync-last-result">
            <div class="sync-last-result__head">
              <strong>最近一次期货同步</strong>
              <span>{{ formatRequestedSourceLabel(futuresLastSyncResult) }} -> {{ formatSyncResolvedSourceKeys(futuresLastSyncResult.result, futuresLastSyncResult.source_key) }}</span>
            </div>
            <div class="sync-result-tags">
              <el-tag v-for="item in buildSyncMetricTags(futuresLastSyncResult.result)" :key="`futures-sync-${item.key}`" :type="item.type">
                {{ item.label }}
              </el-tag>
              <el-tag v-if="buildSyncMetricTags(futuresLastSyncResult.result).length === 0" type="info">
                处理 {{ futuresLastSyncResult.count || 0 }} 条
              </el-tag>
            </div>
            <el-text type="info" size="small">
              范围：{{ formatSyncRequestScope(futuresLastSyncResult.contracts, "默认合约池") }}
            </el-text>
          </div>
        </div>

        <div class="truth-summary-card">
          <div class="truth-summary-card__header">
            <div class="truth-summary-card__title">期货仓单同步</div>
            <el-tag type="info" effect="plain">
              {{ futuresInventorySyncOptions.length ? `${futuresInventorySyncOptions.length} 个可用源` : "无可用源" }}
            </el-tag>
          </div>
          <div class="sync-form-grid">
            <el-select
              v-model="futuresInventorySyncForm.source_key"
              filterable
              allow-create
              default-first-option
              placeholder="选择期货仓单源"
            >
              <el-option
                v-for="item in futuresInventorySyncOptions"
                :key="`inventory-${item.value}`"
                :label="item.label"
                :value="item.value"
                :disabled="item.disabled"
              />
            </el-select>
            <el-input
              v-model="futuresInventorySyncForm.symbols"
              placeholder="品种代码，逗号或换行分隔；留空使用默认品种池"
            />
            <el-input-number v-model="futuresInventorySyncForm.days" :min="1" :max="90" controls-position="right" />
            <el-button type="primary" :loading="syncingFuturesInventory" @click="handleSyncFuturesInventory">同步期货仓单</el-button>
          </div>
          <div class="sync-inline-hint">
            <el-text type="info" size="small">
              当前最小闭环先支持 `TUSHARE / MOCK / AUTO`；品种代码会自动归一到根符号，例如 `RB2405 -> RB`。
            </el-text>
          </div>
          <div v-if="futuresInventoryLastSyncResult" class="sync-last-result">
            <div class="sync-last-result__head">
              <strong>最近一次仓单同步</strong>
              <span>{{ formatRequestedSourceLabel(futuresInventoryLastSyncResult) }} -> {{ formatSyncResolvedSourceKeys(futuresInventoryLastSyncResult.result, futuresInventoryLastSyncResult.source_key) }}</span>
            </div>
            <div class="sync-result-tags">
              <el-tag
                v-for="item in buildSyncMetricTags(futuresInventoryLastSyncResult.result)"
                :key="`inventory-sync-${item.key}`"
                :type="item.type"
              >
                {{ item.label }}
              </el-tag>
              <el-tag v-if="buildSyncMetricTags(futuresInventoryLastSyncResult.result).length === 0" type="info">
                处理 {{ futuresInventoryLastSyncResult.count || 0 }} 条
              </el-tag>
            </div>
            <el-text type="info" size="small">
              范围：{{ formatSyncRequestScope(futuresInventoryLastSyncResult.symbols, "默认品种池") }}
            </el-text>
          </div>
        </div>

        <div class="truth-summary-card">
          <div class="truth-summary-card__header">
            <div class="truth-summary-card__title">市场资讯同步</div>
            <el-tag type="info" effect="plain">
              {{ marketNewsSyncOptions.length ? `${marketNewsSyncOptions.length} 个可用源` : "无可用源" }}
            </el-tag>
          </div>
          <div class="sync-form-grid">
            <el-select
              v-model="marketNewsSyncForm.source_key"
              filterable
              allow-create
              default-first-option
              placeholder="选择资讯源"
            >
              <el-option
                v-for="item in marketNewsSyncOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
                :disabled="item.disabled"
              />
            </el-select>
            <el-input
              v-model="marketNewsSyncForm.symbols"
              placeholder="关联股票代码，逗号或换行分隔；留空表示不限"
            />
            <div class="sync-inline-numbers">
              <el-input-number v-model="marketNewsSyncForm.days" :min="1" :max="30" controls-position="right" />
              <el-input-number v-model="marketNewsSyncForm.limit" :min="1" :max="500" controls-position="right" />
            </div>
            <el-button type="primary" :loading="syncingMarketNews" @click="handleSyncMarketNews">同步市场资讯</el-button>
          </div>
          <div class="sync-inline-hint">
            <el-text type="info" size="small">支持输入 `AUTO` 或自定义链路，如 `AKSHARE,TUSHARE`。</el-text>
          </div>
          <div v-if="marketNewsLastSyncResult" class="sync-last-result">
            <div class="sync-last-result__head">
              <strong>最近一次资讯同步</strong>
              <span>{{ formatRequestedSourceLabel(marketNewsLastSyncResult) }} -> {{ formatSyncResolvedSourceKeys(marketNewsLastSyncResult.result, marketNewsLastSyncResult.source_key) }}</span>
            </div>
            <div class="sync-result-tags">
              <el-tag v-for="item in buildSyncMetricTags(marketNewsLastSyncResult.result)" :key="`news-sync-${item.key}`" :type="item.type">
                {{ item.label }}
              </el-tag>
              <el-tag v-if="buildSyncMetricTags(marketNewsLastSyncResult.result).length === 0" type="info">
                处理 {{ marketNewsLastSyncResult.count || 0 }} 条
              </el-tag>
            </div>
            <el-text type="info" size="small">
              范围：{{ formatSyncRequestScope(marketNewsLastSyncResult.symbols, "不限股票代码") }}
            </el-text>
          </div>
        </div>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <div>
          <h3 style="margin: 0">本地 Truth 派生重建</h3>
          <p class="muted" style="margin: 6px 0 0">
            从本地 truth 行情重建股票状态和期货主力映射，不直接触发外部源拉取。
          </p>
        </div>
        <div class="inline-actions inline-actions--left">
          <el-date-picker
            v-model="rebuildForm.trade_date"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="可选：指定交易日"
            style="width: 180px"
          />
          <el-input-number v-model="rebuildForm.days" :min="1" :max="30" controls-position="right" />
          <el-button
            v-if="canEditMarket"
            type="primary"
            plain
            :loading="rebuildingStockTruth"
            @click="handleRebuildTruth('STOCK')"
          >
            重建股票状态
          </el-button>
          <el-button
            v-if="canEditMarket"
            type="primary"
            :loading="rebuildingFuturesTruth"
            @click="handleRebuildTruth('FUTURES')"
          >
            重建期货主力映射
          </el-button>
        </div>
      </div>

      <div class="truth-summary-grid" style="margin-top: 12px">
        <div class="truth-summary-card">
          <div class="truth-summary-card__title">最近一次股票状态重建</div>
          <template v-if="stockRebuildSummary">
            <ul class="truth-summary-list">
              <li>交易日：{{ stockRebuildSummary.trade_date || "-" }}</li>
              <li>窗口：{{ stockRebuildSummary.start_date || "-" }} ~ {{ stockRebuildSummary.end_date || "-" }}</li>
              <li>窗口天数：{{ stockRebuildSummary.days || 0 }}</li>
              <li>truth 行情：{{ stockRebuildSummary.truth_bar_count || 0 }}</li>
              <li>状态条数：{{ stockRebuildSummary.stock_status_count || 0 }}</li>
              <li>来源：{{ stockRebuildSummary.source_key || "LOCAL_TRUTH" }}</li>
              <li>记录时间：{{ stockRebuildSummary.created_at || "-" }}</li>
              <li>警告：{{ formatSummaryWarnings(stockRebuildSummary) }}</li>
            </ul>
          </template>
          <p v-else class="muted">暂无股票 truth 重建摘要</p>
        </div>

        <div class="truth-summary-card">
          <div class="truth-summary-card__title">最近一次期货主力映射重建</div>
          <template v-if="futuresRebuildSummary">
            <ul class="truth-summary-list">
              <li>交易日：{{ futuresRebuildSummary.trade_date || "-" }}</li>
              <li>窗口：{{ futuresRebuildSummary.start_date || "-" }} ~ {{ futuresRebuildSummary.end_date || "-" }}</li>
              <li>窗口天数：{{ futuresRebuildSummary.days || 0 }}</li>
              <li>truth 行情：{{ futuresRebuildSummary.truth_bar_count || 0 }}</li>
              <li>映射条数：{{ futuresRebuildSummary.futures_mapping_count || 0 }}</li>
              <li>来源：{{ futuresRebuildSummary.source_key || "LOCAL_TRUTH" }}</li>
              <li>记录时间：{{ futuresRebuildSummary.created_at || "-" }}</li>
              <li>警告：{{ formatSummaryWarnings(futuresRebuildSummary) }}</li>
            </ul>
          </template>
          <p v-else class="muted">暂无期货 truth 重建摘要</p>
        </div>
      </div>
    </div>

    <div class="card">
      <el-table :data="items" border stripe v-loading="loading" empty-text="暂无数据源">
        <el-table-column label="source_key" min-width="190">
          <template #default="{ row }">
            <div class="source-key-cell">
              <span>{{ row.source_key }}</span>
              <el-tag v-if="isDefaultStockSource(row.source_key)" size="small" type="warning" effect="plain">
                默认行情源
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="130" />
        <el-table-column prop="source_type" label="类型" min-width="110" />
        <el-table-column label="状态" min-width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="配置摘要" min-width="260">
          <template #default="{ row }">
            <div class="config-line">endpoint: {{ row.config?.endpoint || "-" }}</div>
            <div class="config-line config-line--muted">
              token: {{ row.config?.token ? "已配置" : "未配置" }}
            </div>
            <div class="config-line config-line--muted">
              阈值: {{ row.config?.fail_threshold ?? 3 }} / 重试: {{ row.config?.retry_times ?? 0 }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="最近健康状态" min-width="260">
          <template #default="{ row }">
            <template v-if="healthMap[row.source_key]">
              <el-tag :type="statusTagType(healthMap[row.source_key].status)">
                {{ healthMap[row.source_key].status }}
              </el-tag>
              <div class="config-line config-line--muted">
                {{ healthMap[row.source_key].message || "-" }} · {{ healthMap[row.source_key].latency_ms || 0 }}ms
              </div>
              <div class="config-line config-line--muted">
                尝试 {{ healthMap[row.source_key].attempts || 1 }}/{{ healthMap[row.source_key].max_attempts || 1 }} · 连续失败 {{ healthMap[row.source_key].consecutive_failures || 0 }}
              </div>
            </template>
            <el-text v-else type="info">未检查</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" min-width="180" />
        <el-table-column label="操作" align="right" min-width="420">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-button
                v-if="canEditDataSources && supportsDefaultStockSource(row)"
                size="small"
                type="success"
                plain
                :loading="settingDefaultSource && settingDefaultSourceKey === normalizeSourceKey(row.source_key)"
                :disabled="isDefaultStockSource(row.source_key)"
                @click="handleSetDefaultStockSource(row)"
              >
                设为默认行情源
              </el-button>
              <el-button v-if="canEditDataSources" size="small" @click="handleCheckOne(row.source_key)">健康检查</el-button>
              <el-button size="small" @click="showLogs(row.source_key)">健康日志</el-button>
              <el-button v-if="canEditDataSources" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button
                v-if="canEditDataSources"
                size="small"
                type="danger"
                plain
                @click="handleDelete(row.source_key)"
              >
                删除
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-text type="info">第 {{ page }} 页，共 {{ total }} 条</el-text>
        <el-pagination
          background
          layout="prev, pager, next"
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <div class="card" style="margin-top: 12px">
      <div class="log-header">
        <h3 style="margin: 0">健康日志 {{ logSourceKey ? `(${logSourceKey})` : "" }}</h3>
        <el-text type="info">{{ logsLoading ? "加载中..." : "最近20条" }}</el-text>
      </div>
      <el-table :data="logItems" border stripe v-loading="logsLoading" empty-text="暂无日志，先执行健康检查">
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="reachable" min-width="110">
          <template #default="{ row }">
            {{ row.reachable ? "是" : "否" }}
          </template>
        </el-table-column>
        <el-table-column prop="http_status" label="HTTP" min-width="90">
          <template #default="{ row }">
            {{ row.http_status || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="latency_ms" label="latency(ms)" min-width="110" />
        <el-table-column prop="message" label="message" min-width="220">
          <template #default="{ row }">
            {{ row.message || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="checked_at" label="checked_at" min-width="180" />
      </el-table>
    </div>

    <div class="card" style="margin-top: 12px">
      <div class="section-header">
        <div>
          <h3 style="margin: 0">市场数据质量日志</h3>
          <p class="muted" style="margin: 6px 0 0">
            查看多源行情、主数据和派生重建过程中产生的告警、回退与补写记录。
          </p>
        </div>
        <div class="inline-actions inline-actions--left">
          <el-select v-model="qualityFilters.hours" style="width: 140px">
            <el-option
              v-for="item in MARKET_QUALITY_LOOKBACK_OPTIONS"
              :key="`hours-${item.value}`"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-select v-model="qualityFilters.asset_class" style="width: 140px">
            <el-option
              v-for="item in assetClassOptions"
              :key="item.value || 'asset-all'"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-select v-model="qualityFilters.data_kind" style="width: 150px">
            <el-option
              v-for="item in qualityDataKindOptions"
              :key="item.value || 'kind-all'"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-select v-model="qualityFilters.severity" style="width: 140px">
            <el-option
              v-for="item in severityFilterOptions"
              :key="item.value || 'severity-all'"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
          <el-input
            v-model="qualityFilters.issue_code"
            clearable
            placeholder="输入 issue_code"
            style="width: 180px"
          />
          <el-button
            :loading="qualityLogsLoading"
            @click="
              () => {
                if (!syncQualityFiltersToRoute()) {
                  fetchQualityDashboard();
                }
              }
            "
          >
            查询日志
          </el-button>
          <el-button @click="resetQualityFilters">重置</el-button>
        </div>
      </div>

      <div class="truth-summary-grid" style="margin-top: 12px">
        <div class="truth-summary-card">
          <div class="truth-summary-card__title">{{ qualityLookbackLabel }}总体概览</div>
          <ul class="truth-summary-list">
            <li>日志总数：{{ qualitySummaryAll?.total_count || 0 }}</li>
            <li>错误：{{ qualitySummaryAll?.error_count || 0 }}</li>
            <li>警告：{{ qualitySummaryAll?.warn_count || 0 }}</li>
            <li>信息：{{ qualitySummaryAll?.info_count || 0 }}</li>
            <li>涉及来源：{{ qualitySummaryAll?.distinct_source_count || 0 }}</li>
            <li>最近事件：{{ formatQualitySummaryLatest(qualitySummaryAll) }}</li>
            <li>最近错误：{{ formatQualitySummaryLatestError(qualitySummaryAll) }}</li>
          </ul>
        </div>
        <div class="truth-summary-card">
          <div class="truth-summary-card__header">
            <div class="truth-summary-card__title">{{ qualityLookbackLabel }}股票质量概览</div>
          </div>
          <ul class="truth-summary-list">
            <li>日志总数：{{ qualitySummaryStock?.total_count || 0 }}</li>
            <li>错误：{{ qualitySummaryStock?.error_count || 0 }}</li>
            <li>警告：{{ qualitySummaryStock?.warn_count || 0 }}</li>
            <li>信息：{{ qualitySummaryStock?.info_count || 0 }}</li>
            <li>涉及来源：{{ qualitySummaryStock?.distinct_source_count || 0 }}</li>
            <li>最近事件：{{ formatQualitySummaryLatest(qualitySummaryStock) }}</li>
            <li>最近错误：{{ formatQualitySummaryLatestError(qualitySummaryStock) }}</li>
          </ul>
        </div>
        <div class="truth-summary-card">
          <div class="truth-summary-card__header">
            <div class="truth-summary-card__title">{{ qualityLookbackLabel }}期货质量概览</div>
          </div>
          <ul class="truth-summary-list">
            <li>日志总数：{{ qualitySummaryFutures?.total_count || 0 }}</li>
            <li>错误：{{ qualitySummaryFutures?.error_count || 0 }}</li>
            <li>警告：{{ qualitySummaryFutures?.warn_count || 0 }}</li>
            <li>信息：{{ qualitySummaryFutures?.info_count || 0 }}</li>
            <li>涉及来源：{{ qualitySummaryFutures?.distinct_source_count || 0 }}</li>
            <li>最近事件：{{ formatQualitySummaryLatest(qualitySummaryFutures) }}</li>
            <li>最近错误：{{ formatQualitySummaryLatestError(qualitySummaryFutures) }}</li>
          </ul>
        </div>
      </div>

      <div v-if="qualityIssueQuickOptions.length" class="issue-quick-filter">
        <span class="issue-quick-filter__label">快速筛选：</span>
        <el-tag
          v-for="item in qualityIssueQuickOptions"
          :key="item.value"
          class="issue-quick-filter__tag"
          :type="qualityFilters.issue_code === item.value ? 'primary' : 'info'"
          effect="plain"
          @click="applyQualityIssueQuickFilter(item.value)"
        >
          {{ item.label }}
        </el-tag>
        <el-button
          v-if="qualityFilters.issue_code"
          link
          type="primary"
          @click="applyQualityIssueQuickFilter('')"
        >
          清空 issue_code
        </el-button>
      </div>

      <div class="truth-summary-grid" style="margin-top: 12px">
        <div class="truth-summary-card">
          <div class="truth-summary-card__title">当前筛选结果 · 来源分布</div>
          <ul class="truth-summary-list">
            <li v-for="item in qualitySourceBuckets" :key="`source-bucket-${item.value}`">
              {{ item.label }}：{{ item.count }}
            </li>
            <li v-if="!qualitySourceBuckets.length">暂无统计</li>
          </ul>
        </div>
        <div class="truth-summary-card">
          <div class="truth-summary-card__title">当前筛选结果 · 问题分布</div>
          <ul class="truth-summary-list">
            <li v-for="item in qualityIssueBuckets" :key="`issue-bucket-${item.value}`">
              {{ item.label }}：{{ item.count }}
            </li>
            <li v-if="!qualityIssueBuckets.length">暂无统计</li>
          </ul>
        </div>
        <div class="truth-summary-card">
          <div class="truth-summary-card__title">当前筛选结果 · 级别分布</div>
          <ul class="truth-summary-list">
            <li v-for="item in qualitySeverityBuckets" :key="`severity-bucket-${item.value}`">
              {{ item.label }}：{{ item.count }}
            </li>
            <li v-if="!qualitySeverityBuckets.length">暂无统计</li>
          </ul>
        </div>
      </div>

      <el-table
        :data="qualityLogItems"
        border
        stripe
        v-loading="qualityLogsLoading"
        empty-text="暂无质量日志"
        style="margin-top: 12px"
      >
        <el-table-column prop="asset_class" label="资产" min-width="90" />
        <el-table-column prop="data_kind" label="数据类型" min-width="140" />
        <el-table-column prop="instrument_key" label="标的" min-width="140">
          <template #default="{ row }">
            {{ row.instrument_key || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="trade_date" label="交易日" min-width="110">
          <template #default="{ row }">
            {{ row.trade_date || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="source_key" label="来源" min-width="110">
          <template #default="{ row }">
            {{ row.source_key || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="级别" min-width="100">
          <template #default="{ row }">
            <el-tag :type="marketQualitySeverityTagType(row.severity)">{{ row.severity || "-" }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="issue_code" label="问题编码" min-width="180">
          <template #default="{ row }">
            <el-button link type="primary" @click="applyQualityIssueQuickFilter(row.issue_code)">
              {{ row.issue_code || "-" }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="issue_message" label="问题说明" min-width="260">
          <template #default="{ row }">
            {{ row.issue_message || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="180" />
        <el-table-column label="操作" min-width="220" align="center">
          <template #default="{ row }">
            <div class="inline-actions inline-actions--left">
              <el-button
                size="small"
                :disabled="!row.payload"
                @click="openQualityPayload(row)"
              >
                查看 payload
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-text type="info">共 {{ qualityLogTotal }} 条日志，当前展示最近 20 条</el-text>
      </div>
    </div>

    <el-dialog
      v-model="formVisible"
      :title="formMode === 'create' ? '新增数据源' : `编辑数据源：${form.source_key}`"
      width="860px"
      destroy-on-close
    >
      <el-form label-width="120px">
        <div class="dialog-grid">
          <el-form-item label="source_key" required>
            <el-input v-model="form.source_key" :disabled="formMode === 'edit'" placeholder="wind" />
          </el-form-item>
          <el-form-item label="名称" required>
            <el-input v-model="form.name" placeholder="Wind 数据源" />
          </el-form-item>
          <el-form-item label="类型" required>
            <el-select v-model="form.source_type">
              <el-option v-for="item in sourceTypeOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="form.status">
              <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="endpoint">
            <el-input v-model="form.endpoint" placeholder="Tushare: https://api.tushare.pro" />
          </el-form-item>
          <el-form-item label="token">
            <el-input
              v-model="form.token"
              type="password"
              show-password
              placeholder="Tushare Token（可留空使用后端环境变量）"
            />
          </el-form-item>
          <el-form-item label="fail_threshold">
            <el-input-number v-model="form.fail_threshold" :min="1" :step="1" controls-position="right" />
          </el-form-item>
          <el-form-item label="retry_times">
            <el-input-number v-model="form.retry_times" :min="0" :max="5" :step="1" controls-position="right" />
          </el-form-item>
          <el-form-item label="retry_interval_ms">
            <el-input-number v-model="form.retry_interval_ms" :min="0" :step="100" controls-position="right" />
          </el-form-item>
          <el-form-item label="health_timeout_ms">
            <el-input-number v-model="form.health_timeout_ms" :min="500" :step="100" controls-position="right" />
          </el-form-item>
          <el-form-item label="alert_receiver_id">
            <el-input v-model="form.alert_receiver_id" placeholder="admin_001" />
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button v-if="canEditDataSources" type="primary" :loading="submitting" @click="submitForm">提交</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="qualityPayloadDialogVisible"
      title="质量日志 payload 详情"
      width="760px"
      destroy-on-close
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item label="资产">
          {{ selectedQualityLog?.asset_class || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="来源">
          {{ selectedQualityLog?.source_key || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="问题编码">
          {{ selectedQualityLog?.issue_code || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="级别">
          {{ selectedQualityLog?.severity || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="交易日">
          {{ selectedQualityLog?.trade_date || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ selectedQualityLog?.created_at || "-" }}
        </el-descriptions-item>
      </el-descriptions>

      <div style="margin-top: 12px">
        <div class="truth-summary-card__title">问题说明</div>
        <p class="payload-message">{{ selectedQualityLog?.issue_message || "-" }}</p>
      </div>

      <div style="margin-top: 12px">
        <div class="truth-summary-card__title">payload</div>
        <pre class="payload-viewer">{{ selectedQualityPayloadText || "无 payload" }}</pre>
      </div>

      <template #footer>
        <el-button @click="qualityPayloadDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.config-line {
  line-height: 1.5;
  word-break: break-all;
}

.source-key-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.config-line--muted {
  color: #6b7280;
}

.inline-actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
}

.log-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.inline-actions--left {
  justify-content: flex-start;
}

.dialog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 0 12px;
}

.truth-summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 12px;
}

.truth-summary-card {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 14px 16px;
  background: #fafafa;
}

.truth-summary-card__title {
  font-weight: 600;
  margin-bottom: 10px;
}

.truth-summary-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.truth-summary-list {
  margin: 0;
  padding-left: 18px;
  line-height: 1.7;
}

.issue-quick-filter {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 12px;
}

.issue-quick-filter__label {
  color: #6b7280;
  font-size: 13px;
}

.issue-quick-filter__tag {
  cursor: pointer;
}

.sync-form-grid {
  display: grid;
  grid-template-columns: 1.3fr 1.4fr 120px auto;
  gap: 10px;
  align-items: center;
}

.sync-inline-numbers {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.sync-inline-hint {
  margin-top: 10px;
}

.sync-last-result {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed #d1d5db;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sync-last-result__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  color: #374151;
}

.sync-result-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.payload-message {
  margin: 8px 0 0;
  color: #374151;
  line-height: 1.7;
}

.payload-viewer {
  margin: 8px 0 0;
  padding: 12px;
  background: #111827;
  color: #e5e7eb;
  border-radius: 12px;
  max-height: 420px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}

:deep(.dialog-grid .el-form-item) {
  margin-bottom: 14px;
}

:deep(.dialog-grid .el-select),
:deep(.dialog-grid .el-input-number) {
  width: 100%;
}

@media (max-width: 1080px) {
  .sync-form-grid {
    grid-template-columns: 1fr;
  }

  .sync-inline-numbers {
    grid-template-columns: 1fr 1fr;
  }
}
</style>
