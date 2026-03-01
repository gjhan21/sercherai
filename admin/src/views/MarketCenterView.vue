<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import {
  batchCheckDataSources,
  createMarketEvent,
  createFuturesStrategy,
  createStockRecommendation,
  generateDailyStockRecommendations,
  listDataSources,
  listOperationLogs,
  listQuantEvaluation,
  listQuantTopStocks,
  listMarketEvents,
  listFuturesStrategies,
  listSystemConfigs,
  listStockRecommendations,
  syncStockQuotes,
  updateMarketEvent,
  updateFuturesStrategyStatus,
  updateStockRecommendationStatus
} from "../api/admin";
import { getAccessToken } from "../lib/session";

const activeTab = ref("stocks");
const errorMessage = ref("");
const message = ref("");
const refreshingAll = ref(false);

const stockLoading = ref(false);
const stockSubmitting = ref(false);
const stockGenerating = ref(false);
const stockPage = ref(1);
const stockPageSize = ref(20);
const stockTotal = ref(0);
const stocks = ref([]);
const stockFilters = reactive({
  status: ""
});
const stockDraftStatusMap = ref({});
const stockTradeDate = ref("");
const stockDialogVisible = ref(false);
const stockForm = reactive({
  symbol: "",
  name: "",
  score: 80,
  risk_level: "MEDIUM",
  position_range: "10%-20%",
  valid_from: "",
  valid_to: "",
  status: "PUBLISHED",
  reason_summary: ""
});
const stockQuantLoading = ref(false);
const stockQuoteSyncing = ref(false);
const stockSourceLoading = ref(false);
const stockAutoFallback = ref(true);
const stockQuantUpdatedAt = ref("");
const stockQuantList = ref([]);
const stockEvalLoading = ref(false);
const stockEvalExporting = ref(false);
const stockEvalList = ref([]);
const stockEvalRiskList = ref([]);
const stockEvalRotationList = ref([]);
const stockEvalSummary = ref({
  window_days: 60,
  top_n: 10,
  sample_days: 0,
  sample_count: 0,
  avg_return_5: 0,
  hit_rate_5: 0,
  max_drawdown_5: 0,
  avg_return_10: 0,
  hit_rate_10: 0,
  max_drawdown_10: 0,
  benchmark_avg_return_5: 0,
  benchmark_avg_return_10: 0
});
const stockSourceOptions = ref([]);
const stockSourceHealthMap = ref({});
const stockDefaultSourceLoading = ref(false);
const stockDefaultSourceKey = ref("TUSHARE");
const stockSyncLogsLoading = ref(false);
const stockSyncLogsPage = ref(1);
const stockSyncLogsPageSize = ref(10);
const stockSyncLogsTotal = ref(0);
const stockSyncLogs = ref([]);
const stockQuoteSyncForm = reactive({
  source_key: "TUSHARE",
  symbols: "",
  days: 120
});
const stockQuantQuery = reactive({
  limit: 10,
  lookback_days: 120
});
const stockEvalQuery = reactive({
  days: 60,
  top_n: 10
});

const futuresLoading = ref(false);
const futuresSubmitting = ref(false);
const futuresPage = ref(1);
const futuresPageSize = ref(20);
const futuresTotal = ref(0);
const futures = ref([]);
const futuresFilters = reactive({
  status: "",
  contract: ""
});
const futuresDraftStatusMap = ref({});
const futuresDialogVisible = ref(false);
const futuresForm = reactive({
  contract: "",
  name: "",
  direction: "LONG",
  risk_level: "MEDIUM",
  position_range: "10%-20%",
  valid_from: "",
  valid_to: "",
  status: "PUBLISHED",
  reason_summary: ""
});

const eventsLoading = ref(false);
const eventsSubmitting = ref(false);
const eventsPage = ref(1);
const eventsPageSize = ref(20);
const eventsTotal = ref(0);
const events = ref([]);
const eventsFilters = reactive({
  event_type: "",
  symbol: ""
});
const eventDialogVisible = ref(false);
const eventDialogMode = ref("create");
const eventForm = reactive({
  id: "",
  event_type: "PRICE",
  symbol: "",
  summary: "",
  trigger_rule: "",
  source: ""
});

const stockStatusOptions = ["PUBLISHED", "ACTIVE", "DRAFT", "DISABLED"];
const futuresStatusOptions = ["PUBLISHED", "ACTIVE", "DRAFT", "DISABLED"];
const riskLevelOptions = ["LOW", "MEDIUM", "HIGH"];
const directionOptions = ["LONG", "SHORT", "NEUTRAL"];
const marketEventTypeOptions = ["PRICE", "VOLUME", "VOLATILITY", "POLICY", "FLOW", "OTHER"];
const STOCK_DEFAULT_SOURCE_CONFIG_KEY = "stock.quotes.default_source_key";
const STOCK_DEFAULT_SOURCE_FALLBACK = "TUSHARE";

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function clearMessages() {
  errorMessage.value = "";
  message.value = "";
}

function normalizeToRFC3339(value) {
  const text = String(value || "").trim();
  if (!text) {
    return "";
  }
  const timestamp = Date.parse(text);
  if (!Number.isNaN(timestamp)) {
    return new Date(timestamp).toISOString();
  }
  if (/^\d{4}-\d{2}-\d{2}$/.test(text)) {
    const timestampDateOnly = Date.parse(`${text}T00:00:00`);
    if (!Number.isNaN(timestampDateOnly)) {
      return new Date(timestampDateOnly).toISOString();
    }
  }
  return "";
}

function splitSymbols(rawText) {
  const set = new Set();
  String(rawText || "")
    .split(/[\s,，;\n]+/)
    .map((item) => item.trim().toUpperCase())
    .filter(Boolean)
    .forEach((item) => set.add(item));
  return Array.from(set);
}

function formatDateTime(value) {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) {
    return "-";
  }
  const date = new Date(timestamp);
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  const hour = `${date.getHours()}`.padStart(2, "0");
  const minute = `${date.getMinutes()}`.padStart(2, "0");
  return `${year}-${month}-${day} ${hour}:${minute}`;
}

function formatPercent(value, digits = 2, signed = false) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return "-";
  }
  const percent = numeric * 100;
  const prefix = signed && percent > 0 ? "+" : "";
  return `${prefix}${percent.toFixed(digits)}%`;
}

function buildCurveModel(rows, mappings) {
  const width = 760;
  const height = 220;
  const padding = 28;
  if (!Array.isArray(rows) || rows.length === 0) {
    return {
      width,
      height,
      minValue: 0,
      maxValue: 0,
      paths: mappings.map((mapping) => ({ ...mapping, d: "" }))
    };
  }
  const values = [];
  rows.forEach((row) => {
    mappings.forEach((mapping) => {
      const value = Number(mapping.selector(row));
      if (Number.isFinite(value)) {
        values.push(value);
      }
    });
  });
  let minValue = Math.min(0, ...values);
  let maxValue = Math.max(0, ...values);
  if (!Number.isFinite(minValue)) {
    minValue = 0;
  }
  if (!Number.isFinite(maxValue)) {
    maxValue = 0;
  }
  if (Math.abs(maxValue - minValue) < 1e-9) {
    maxValue += 0.01;
    minValue -= 0.01;
  }
  const innerWidth = width - padding * 2;
  const innerHeight = height - padding * 2;
  const toPoint = (index, value) => {
    const x = padding + (rows.length <= 1 ? innerWidth / 2 : (innerWidth * index) / (rows.length - 1));
    const ratio = (value - minValue) / (maxValue - minValue);
    const y = padding + (1 - ratio) * innerHeight;
    return [x, y];
  };
  const paths = mappings.map((mapping) => {
    const points = rows
      .map((row, index) => {
        const value = Number(mapping.selector(row));
        if (!Number.isFinite(value)) {
          return null;
        }
        return toPoint(index, value);
      })
      .filter(Boolean);
    const d = points
      .map((point, index) => `${index === 0 ? "M" : "L"} ${point[0].toFixed(2)} ${point[1].toFixed(2)}`)
      .join(" ");
    return { ...mapping, d };
  });
  return { width, height, minValue, maxValue, paths };
}

const stockEvalCurve5 = computed(() =>
  buildCurveModel(stockEvalList.value, [
    { key: "strategy", label: "策略累计收益(5日)", color: "#0f766e", selector: (row) => row.cumulative_return_5 },
    { key: "benchmark", label: "基准累计收益(5日)", color: "#9a3412", selector: (row) => row.cumulative_benchmark_5 },
    { key: "excess", label: "超额收益(5日)", color: "#1d4ed8", selector: (row) => row.cumulative_excess_5 }
  ])
);

const stockEvalCurve10 = computed(() =>
  buildCurveModel(stockEvalList.value, [
    { key: "strategy", label: "策略累计收益(10日)", color: "#047857", selector: (row) => row.cumulative_return_10 },
    { key: "benchmark", label: "基准累计收益(10日)", color: "#b45309", selector: (row) => row.cumulative_benchmark_10 },
    { key: "excess", label: "超额收益(10日)", color: "#2563eb", selector: (row) => row.cumulative_excess_10 }
  ])
);

function curveZeroLineY(curve) {
  const minValue = Number(curve?.minValue);
  const maxValue = Number(curve?.maxValue);
  const height = Number(curve?.height) || 220;
  if (!Number.isFinite(minValue) || !Number.isFinite(maxValue) || Math.abs(maxValue - minValue) < 1e-9) {
    return height - 28;
  }
  const ratio = (maxValue - 0) / (maxValue - minValue);
  return ratio * (height - 56) + 28;
}

function normalizeSourceKey(value) {
  return String(value || "")
    .trim()
    .toUpperCase();
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (["ACTIVE", "PUBLISHED", "SUCCESS", "LOW"].includes(normalized)) return "success";
  if (["DRAFT", "PENDING", "RUNNING", "MEDIUM"].includes(normalized)) return "warning";
  if (["DISABLED", "FAILED", "REJECTED", "HIGH"].includes(normalized)) return "danger";
  return "info";
}

function normalizeStockSourceOption(item) {
  const sourceKey = normalizeSourceKey(item?.source_key);
  const name = String(item?.name || "").trim();
  const provider = String(item?.config?.provider || item?.config?.vendor || "").trim().toUpperCase();
  const sourceType = String(item?.source_type || "").trim().toUpperCase();
  const status = String(item?.status || "").trim().toUpperCase();
  const healthItem = stockSourceHealthMap.value[sourceKey] || null;
  const healthStatus = String(healthItem?.status || "").trim().toUpperCase();
  const healthMessage = String(healthItem?.message || "").trim();
  const tags = [];
  if (provider) {
    tags.push(provider);
  }
  if (sourceType) {
    tags.push(sourceType);
  }
  if (status) {
    tags.push(status);
  }
  if (healthStatus) {
    tags.push(`HEALTH:${healthStatus}`);
  }
  const disabled = status !== "ACTIVE" || healthStatus === "UNHEALTHY";
  return {
    value: sourceKey,
    label: `${sourceKey}${name ? ` · ${name}` : ""}${tags.length ? ` (${tags.join("/")})` : ""}`,
    status,
    source_type: sourceType,
    health_status: healthStatus,
    health_message: healthMessage,
    disabled
  };
}

function compareStockSourceOption(left, right) {
  if (left.value === "TUSHARE" && right.value !== "TUSHARE") {
    return -1;
  }
  if (left.value !== "TUSHARE" && right.value === "TUSHARE") {
    return 1;
  }
  return left.value.localeCompare(right.value);
}

async function fetchStockSourceOptions(options = {}) {
  const { keepMessage = false } = options;
  stockSourceLoading.value = true;
  if (!keepMessage) {
    errorMessage.value = "";
    message.value = "";
  }
  try {
    await fetchStockDefaultSourceKey({ keepMessage: true });
    const data = await listDataSources({ page: 1, page_size: 200 });
    const rows = Array.isArray(data?.items) ? data.items : [];
    const stockRows = rows.filter((item) => String(item?.source_type || "").trim().toUpperCase() === "STOCK");
    const sourceKeys = stockRows.map((item) => normalizeSourceKey(item?.source_key)).filter(Boolean);
    stockSourceHealthMap.value = {};
    if (sourceKeys.length > 0) {
      try {
        const healthData = await batchCheckDataSources(sourceKeys);
        const healthItems = Array.isArray(healthData?.items) ? healthData.items : [];
        const healthMap = {};
        healthItems.forEach((healthItem) => {
          const key = normalizeSourceKey(healthItem?.source_key);
          if (!key) {
            return;
          }
          healthMap[key] = healthItem;
        });
        stockSourceHealthMap.value = healthMap;
      } catch (healthError) {
        console.warn("batch check stock data sources failed:", healthError?.message || healthError);
      }
    }
    const normalized = stockRows
      .map((item) => normalizeStockSourceOption(item))
      .filter((item) => item.value);
    const active = normalized.filter((item) => item.status === "ACTIVE");
    const finalOptions = (active.length > 0 ? active : normalized)
      .reduce((acc, item) => {
        if (!acc.some((existing) => existing.value === item.value)) {
          acc.push(item);
        }
        return acc;
      }, [])
      .sort(compareStockSourceOption);
    stockSourceOptions.value = finalOptions;

    const currentSourceKey = normalizeSourceKey(stockQuoteSyncForm.source_key);
    const hasCurrent = finalOptions.some((item) => item.value === currentSourceKey);
    if (finalOptions.length > 0 && (!currentSourceKey || !hasCurrent)) {
      const preferred =
        finalOptions.find((item) => item.value === normalizeSourceKey(stockDefaultSourceKey.value)) ||
        finalOptions.find((item) => item.value === "TUSHARE") ||
        finalOptions[0];
      stockQuoteSyncForm.source_key = preferred.value;
    } else {
      stockQuoteSyncForm.source_key = currentSourceKey || stockQuoteSyncForm.source_key;
    }
  } catch (error) {
    if (!keepMessage) {
      errorMessage.value = normalizeErrorMessage(error, "加载行情数据源失败");
    }
  } finally {
    stockSourceLoading.value = false;
  }
}

const currentStockSourceOption = computed(() => {
  const currentKey = normalizeSourceKey(stockQuoteSyncForm.source_key);
  return stockSourceOptions.value.find((item) => item.value === currentKey) || null;
});

function sourceHealthTagType(status) {
  const normalized = String(status || "").trim().toUpperCase();
  if (normalized === "HEALTHY") {
    return "success";
  }
  if (normalized === "UNHEALTHY") {
    return "danger";
  }
  return "info";
}

function parseSyncCount(afterValue) {
  const text = String(afterValue || "");
  const matched = text.match(/count\s*=\s*(\d+)/i);
  if (!matched) {
    return 0;
  }
  return Number(matched[1]) || 0;
}

async function fetchStockDefaultSourceKey(options = {}) {
  const { keepMessage = false } = options;
  stockDefaultSourceLoading.value = true;
  if (!keepMessage) {
    errorMessage.value = "";
    message.value = "";
  }
  try {
    const data = await listSystemConfigs({
      keyword: STOCK_DEFAULT_SOURCE_CONFIG_KEY,
      page: 1,
      page_size: 50
    });
    const rows = Array.isArray(data?.items) ? data.items : [];
    const matched = rows.find(
      (item) =>
        normalizeSourceKey(item?.config_key) === normalizeSourceKey(STOCK_DEFAULT_SOURCE_CONFIG_KEY)
    );
    const parsed = normalizeSourceKey(matched?.config_value);
    stockDefaultSourceKey.value = parsed || STOCK_DEFAULT_SOURCE_FALLBACK;
  } catch (error) {
    stockDefaultSourceKey.value = STOCK_DEFAULT_SOURCE_FALLBACK;
    if (!keepMessage) {
      errorMessage.value = normalizeErrorMessage(error, "加载默认行情源失败");
    }
  } finally {
    stockDefaultSourceLoading.value = false;
  }
}

function applyStockDefaultSource() {
  const nextSourceKey = normalizeSourceKey(stockDefaultSourceKey.value);
  if (!nextSourceKey) {
    return;
  }
  stockQuoteSyncForm.source_key = nextSourceKey;
  message.value = `已应用默认行情源：${nextSourceKey}`;
}

async function fetchStockSyncLogs(options = {}) {
  const { keepMessage = false } = options;
  stockSyncLogsLoading.value = true;
  if (!keepMessage) {
    errorMessage.value = "";
    message.value = "";
  }
  try {
    const data = await listOperationLogs({
      module: "STOCK",
      action: "SYNC_QUOTES",
      page: stockSyncLogsPage.value,
      page_size: stockSyncLogsPageSize.value
    });
    const rows = Array.isArray(data?.items) ? data.items : [];
    stockSyncLogs.value = rows.map((item) => ({
      ...item,
      sync_count: parseSyncCount(item?.after_value)
    }));
    stockSyncLogsTotal.value = data?.total || 0;
  } catch (error) {
    if (!keepMessage) {
      errorMessage.value = normalizeErrorMessage(error, "加载行情同步日志失败");
    }
  } finally {
    stockSyncLogsLoading.value = false;
  }
}

function handleStockSyncLogPageChange(nextPage) {
  if (nextPage === stockSyncLogsPage.value) {
    return;
  }
  stockSyncLogsPage.value = nextPage;
  fetchStockSyncLogs({ keepMessage: true });
}

function buildSourceUnavailableReason(option) {
  const reasons = [];
  const status = String(option?.status || "").trim().toUpperCase();
  const healthStatus = String(option?.health_status || "").trim().toUpperCase();
  if (status && status !== "ACTIVE") {
    reasons.push(`状态=${status}`);
  }
  if (healthStatus === "UNHEALTHY") {
    const healthMessage = String(option?.health_message || "").trim();
    reasons.push(`健康=${healthStatus}${healthMessage ? `(${healthMessage})` : ""}`);
  }
  if (reasons.length === 0) {
    return "状态不可用";
  }
  return reasons.join("，");
}

function syncStockDrafts() {
  const map = {};
  stocks.value.forEach((item) => {
    map[item.id] = item.status || "DRAFT";
  });
  stockDraftStatusMap.value = map;
}

function syncFuturesDrafts() {
  const map = {};
  futures.value.forEach((item) => {
    map[item.id] = item.status || "ACTIVE";
  });
  futuresDraftStatusMap.value = map;
}

function resetStockForm() {
  Object.assign(stockForm, {
    symbol: "",
    name: "",
    score: 80,
    risk_level: "MEDIUM",
    position_range: "10%-20%",
    valid_from: "",
    valid_to: "",
    status: "PUBLISHED",
    reason_summary: ""
  });
}

function resetFuturesForm() {
  Object.assign(futuresForm, {
    contract: "",
    name: "",
    direction: "LONG",
    risk_level: "MEDIUM",
    position_range: "10%-20%",
    valid_from: "",
    valid_to: "",
    status: "PUBLISHED",
    reason_summary: ""
  });
}

function resetEventForm() {
  Object.assign(eventForm, {
    id: "",
    event_type: "PRICE",
    symbol: "",
    summary: "",
    trigger_rule: "",
    source: ""
  });
}

async function fetchStocks(options = {}) {
  const { keepMessage = false } = options;
  stockLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listStockRecommendations({
      status: stockFilters.status,
      page: stockPage.value,
      page_size: stockPageSize.value
    });
    stocks.value = data.items || [];
    stockTotal.value = data.total || 0;
    syncStockDrafts();
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载股票推荐失败");
  } finally {
    stockLoading.value = false;
  }
}

async function fetchStockQuantTop(options = {}) {
  const { keepMessage = false } = options;
  stockQuantLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listQuantTopStocks({
      limit: Number(stockQuantQuery.limit) || 10,
      lookback_days: Number(stockQuantQuery.lookback_days) || 120
    });
    stockQuantList.value = data.items || [];
    stockQuantUpdatedAt.value = new Date().toISOString();
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载量化Top股票失败");
  } finally {
    stockQuantLoading.value = false;
  }
}

async function fetchStockQuantEvaluation(options = {}) {
  const { keepMessage = false } = options;
  stockEvalLoading.value = true;
  if (!keepMessage) {
    errorMessage.value = "";
    message.value = "";
  }
  try {
    const data = await listQuantEvaluation({
      days: Number(stockEvalQuery.days) || 60,
      top_n: Number(stockEvalQuery.top_n) || 10
    });
    stockEvalList.value = Array.isArray(data?.items) ? data.items : [];
    stockEvalRiskList.value = Array.isArray(data?.risk_items) ? data.risk_items : [];
    stockEvalRotationList.value = Array.isArray(data?.rotation_items) ? data.rotation_items : [];
    stockEvalSummary.value = data?.summary || stockEvalSummary.value;
  } catch (error) {
    if (!keepMessage) {
      errorMessage.value = normalizeErrorMessage(error, "加载量化回测评估失败");
    }
  } finally {
    stockEvalLoading.value = false;
  }
}

async function handleSyncStockQuotes() {
  const sourceKey = normalizeSourceKey(stockQuoteSyncForm.source_key) || "MOCK";
  const payload = {
    source_key: sourceKey,
    symbols: splitSymbols(stockQuoteSyncForm.symbols),
    days: Number(stockQuoteSyncForm.days) || 120
  };
  stockQuoteSyncing.value = true;
  clearMessages();
  try {
    const selectedSourceOption = stockSourceOptions.value.find((item) => item.value === sourceKey) || null;
    if (selectedSourceOption?.disabled) {
      const reason = buildSourceUnavailableReason(selectedSourceOption);
      const canFallbackDirectly = stockAutoFallback.value && sourceKey !== "MOCK";
      if (canFallbackDirectly) {
        const fallbackData = await syncStockQuotes({
          ...payload,
          source_key: "MOCK"
        });
        message.value = `数据源 ${sourceKey} 当前不可用（${reason}），已回退 MOCK，处理 ${fallbackData.count || 0} 条`;
        stockSyncLogsPage.value = 1;
        await Promise.all([
          fetchStockQuantTop({ keepMessage: true }),
          fetchStockQuantEvaluation({ keepMessage: true }),
          fetchStocks({ keepMessage: true }),
          fetchStockSyncLogs({ keepMessage: true })
        ]);
        return;
      }
      errorMessage.value = `数据源 ${sourceKey} 当前不可用（${reason}），请切换可用数据源`;
      return;
    }

    const data = await syncStockQuotes(payload);
    message.value = `行情同步完成，处理 ${data.count || 0} 条`;
    stockSyncLogsPage.value = 1;
    await Promise.all([
      fetchStockQuantTop({ keepMessage: true }),
      fetchStockQuantEvaluation({ keepMessage: true }),
      fetchStocks({ keepMessage: true }),
      fetchStockSyncLogs({ keepMessage: true })
    ]);
  } catch (error) {
    const primaryError = normalizeErrorMessage(error, "同步行情失败");
    const canFallback = stockAutoFallback.value && sourceKey !== "MOCK";
    if (canFallback) {
      try {
        const fallbackData = await syncStockQuotes({
          ...payload,
          source_key: "MOCK"
        });
        message.value = `主数据源 ${sourceKey} 同步失败，已回退 MOCK，处理 ${fallbackData.count || 0} 条（${primaryError}）`;
        stockSyncLogsPage.value = 1;
        await Promise.all([
          fetchStockQuantTop({ keepMessage: true }),
          fetchStockQuantEvaluation({ keepMessage: true }),
          fetchStocks({ keepMessage: true }),
          fetchStockSyncLogs({ keepMessage: true })
        ]);
        return;
      } catch (fallbackError) {
        const fallbackMessage = normalizeErrorMessage(fallbackError, "MOCK 回退失败");
        errorMessage.value = `主数据源 ${sourceKey} 失败：${primaryError}；回退 MOCK 失败：${fallbackMessage}`;
        return;
      }
    }
    errorMessage.value = primaryError;
  } finally {
    stockQuoteSyncing.value = false;
  }
}

async function fetchFutures(options = {}) {
  const { keepMessage = false } = options;
  futuresLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listFuturesStrategies({
      status: futuresFilters.status,
      contract: futuresFilters.contract.trim(),
      page: futuresPage.value,
      page_size: futuresPageSize.value
    });
    futures.value = data.items || [];
    futuresTotal.value = data.total || 0;
    syncFuturesDrafts();
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载期货策略失败");
  } finally {
    futuresLoading.value = false;
  }
}

async function fetchEvents(options = {}) {
  const { keepMessage = false } = options;
  eventsLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listMarketEvents({
      event_type: eventsFilters.event_type.trim(),
      symbol: eventsFilters.symbol.trim(),
      page: eventsPage.value,
      page_size: eventsPageSize.value
    });
    events.value = data.items || [];
    eventsTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载市场事件失败");
  } finally {
    eventsLoading.value = false;
  }
}

async function submitStock() {
  const validFrom = normalizeToRFC3339(stockForm.valid_from);
  const validTo = normalizeToRFC3339(stockForm.valid_to);
  if (!validFrom || !validTo) {
    errorMessage.value = "生效起止时间格式错误，请重新选择日期时间";
    return;
  }
  if (Date.parse(validTo) <= Date.parse(validFrom)) {
    errorMessage.value = "生效止时间必须晚于生效起时间";
    return;
  }
  const payload = {
    symbol: stockForm.symbol.trim(),
    name: stockForm.name.trim(),
    score: Number(stockForm.score),
    risk_level: stockForm.risk_level,
    position_range: stockForm.position_range.trim(),
    valid_from: validFrom,
    valid_to: validTo,
    status: stockForm.status,
    reason_summary: stockForm.reason_summary.trim()
  };
  if (!payload.symbol || !payload.name || !payload.valid_from || !payload.valid_to) {
    errorMessage.value = "请完整填写股票推荐必填字段";
    return;
  }
  stockSubmitting.value = true;
  clearMessages();
  try {
    await createStockRecommendation(payload);
    stockDialogVisible.value = false;
    resetStockForm();
    await fetchStocks({ keepMessage: true });
    message.value = `股票推荐 ${payload.symbol} 已创建`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "创建股票推荐失败");
  } finally {
    stockSubmitting.value = false;
  }
}

async function submitFutures() {
  const validFrom = normalizeToRFC3339(futuresForm.valid_from);
  const validTo = normalizeToRFC3339(futuresForm.valid_to);
  if (!validFrom || !validTo) {
    errorMessage.value = "生效起止时间格式错误，请重新选择日期时间";
    return;
  }
  if (Date.parse(validTo) <= Date.parse(validFrom)) {
    errorMessage.value = "生效止时间必须晚于生效起时间";
    return;
  }
  const payload = {
    contract: futuresForm.contract.trim(),
    name: futuresForm.name.trim(),
    direction: futuresForm.direction,
    risk_level: futuresForm.risk_level,
    position_range: futuresForm.position_range.trim(),
    valid_from: validFrom,
    valid_to: validTo,
    status: futuresForm.status,
    reason_summary: futuresForm.reason_summary.trim()
  };
  if (!payload.contract || !payload.direction || !payload.valid_from || !payload.valid_to) {
    errorMessage.value = "请完整填写期货策略必填字段";
    return;
  }
  futuresSubmitting.value = true;
  clearMessages();
  try {
    await createFuturesStrategy(payload);
    futuresDialogVisible.value = false;
    resetFuturesForm();
    await fetchFutures({ keepMessage: true });
    message.value = `期货策略 ${payload.contract} 已创建`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "创建期货策略失败");
  } finally {
    futuresSubmitting.value = false;
  }
}

function openCreateEventDialog() {
  eventDialogMode.value = "create";
  resetEventForm();
  eventDialogVisible.value = true;
}

function openEditEventDialog(row) {
  eventDialogMode.value = "edit";
  Object.assign(eventForm, {
    id: row.id || "",
    event_type: row.event_type || "PRICE",
    symbol: row.symbol || "",
    summary: row.summary || "",
    trigger_rule: row.trigger_rule || "",
    source: row.source || ""
  });
  eventDialogVisible.value = true;
}

async function submitEvent() {
  const payload = {
    event_type: eventForm.event_type.trim(),
    symbol: eventForm.symbol.trim(),
    summary: eventForm.summary.trim(),
    trigger_rule: eventForm.trigger_rule.trim(),
    source: eventForm.source.trim()
  };
  if (!payload.event_type || !payload.symbol || !payload.summary || !payload.trigger_rule) {
    errorMessage.value = "请完整填写市场事件必填字段";
    return;
  }

  eventsSubmitting.value = true;
  clearMessages();
  try {
    if (eventDialogMode.value === "edit") {
      await updateMarketEvent(eventForm.id, payload);
      message.value = `市场事件 ${eventForm.id} 已更新`;
    } else {
      await createMarketEvent(payload);
      message.value = `市场事件 ${payload.symbol} 已创建`;
    }
    eventDialogVisible.value = false;
    resetEventForm();
    await fetchEvents({ keepMessage: true });
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "保存市场事件失败");
  } finally {
    eventsSubmitting.value = false;
  }
}

async function saveStockStatus(item) {
  const target = (stockDraftStatusMap.value[item.id] || "").trim();
  if (!target || target === item.status) {
    return;
  }
  clearMessages();
  try {
    await updateStockRecommendationStatus(item.id, target);
    await fetchStocks({ keepMessage: true });
    message.value = `股票推荐 ${item.id} 状态已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "更新股票推荐状态失败");
  }
}

async function saveFuturesStatus(item) {
  const target = (futuresDraftStatusMap.value[item.id] || "").trim();
  if (!target || target === item.status) {
    return;
  }
  clearMessages();
  try {
    await updateFuturesStrategyStatus(item.id, target);
    await fetchFutures({ keepMessage: true });
    message.value = `期货策略 ${item.id} 状态已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "更新期货策略状态失败");
  }
}

async function handleGenerateDailyStocks() {
  stockGenerating.value = true;
  clearMessages();
  try {
    const data = await generateDailyStockRecommendations(stockTradeDate.value.trim());
    message.value = `已生成每日股票推荐 ${data.count || 0} 条`;
    await fetchStocks({ keepMessage: true });
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "生成每日股票推荐失败");
  } finally {
    stockGenerating.value = false;
  }
}

function refreshStockQuantTop() {
  fetchStockQuantTop();
}

function refreshStockQuantEvaluation() {
  fetchStockQuantEvaluation();
}

async function exportStockQuantEvaluationCSV() {
  stockEvalExporting.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const params = new URLSearchParams();
    params.set("days", String(Number(stockEvalQuery.days) || 60));
    params.set("top_n", String(Number(stockEvalQuery.top_n) || 10));
    const baseURL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
    const requestURL = `${baseURL}/admin/stocks/quant/evaluation/export.csv?${params.toString()}`;
    const headers = {};
    const token = getAccessToken();
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    const response = await fetch(requestURL, { method: "GET", headers });
    if (!response.ok) {
      const text = await response.text();
      throw new Error(text || `导出失败(${response.status})`);
    }
    const blob = await response.blob();
    const blobURL = URL.createObjectURL(blob);
    const fileName = `stock_quant_evaluation_${new Date().toISOString().slice(0, 10)}.csv`;
    const anchor = document.createElement("a");
    anchor.href = blobURL;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(blobURL);
    message.value = "量化评估CSV已导出";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "导出量化评估CSV失败");
  } finally {
    stockEvalExporting.value = false;
  }
}

function applyStockFilters() {
  stockPage.value = 1;
  fetchStocks();
}

function resetStockFilters() {
  stockFilters.status = "";
  stockPage.value = 1;
  fetchStocks();
}

function handleStockPageChange(nextPage) {
  if (nextPage === stockPage.value) {
    return;
  }
  stockPage.value = nextPage;
  fetchStocks();
}

function applyFuturesFilters() {
  futuresPage.value = 1;
  fetchFutures();
}

function resetFuturesFilters() {
  futuresFilters.status = "";
  futuresFilters.contract = "";
  futuresPage.value = 1;
  fetchFutures();
}

function handleFuturesPageChange(nextPage) {
  if (nextPage === futuresPage.value) {
    return;
  }
  futuresPage.value = nextPage;
  fetchFutures();
}

function applyEventFilters() {
  eventsPage.value = 1;
  fetchEvents();
}

function resetEventFilters() {
  eventsFilters.event_type = "";
  eventsFilters.symbol = "";
  eventsPage.value = 1;
  fetchEvents();
}

function handleEventPageChange(nextPage) {
  if (nextPage === eventsPage.value) {
    return;
  }
  eventsPage.value = nextPage;
  fetchEvents();
}

async function refreshCurrentTab() {
  if (activeTab.value === "stocks") {
    await Promise.all([
      fetchStockSourceOptions({ keepMessage: true }),
      fetchStocks(),
      fetchStockQuantTop(),
      fetchStockQuantEvaluation({ keepMessage: true }),
      fetchStockSyncLogs({ keepMessage: true })
    ]);
    return;
  }
  if (activeTab.value === "futures") {
    await fetchFutures();
    return;
  }
  await fetchEvents();
}

async function refreshAll() {
  refreshingAll.value = true;
  clearMessages();
  try {
    await Promise.all([
      fetchStockSourceOptions({ keepMessage: true }),
      fetchStocks({ keepMessage: true }),
      fetchStockQuantTop({ keepMessage: true }),
      fetchStockQuantEvaluation({ keepMessage: true }),
      fetchStockSyncLogs({ keepMessage: true }),
      fetchFutures({ keepMessage: true }),
      fetchEvents({ keepMessage: true })
    ]);
    message.value = "策略中心数据已刷新";
  } finally {
    refreshingAll.value = false;
  }
}

onMounted(refreshAll);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">策略中心</h1>
        <p class="muted">股票推荐、期货策略与市场事件维护发布</p>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <el-button :loading="refreshingAll" @click="refreshAll">刷新全部</el-button>
        <el-button type="primary" plain :loading="refreshingAll" @click="refreshCurrentTab">刷新当前页签</el-button>
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

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="股票推荐" name="stocks">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select
              v-model="stockFilters.status"
              clearable
              filterable
              placeholder="状态"
              style="width: 160px"
            >
              <el-option v-for="item in stockStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-input v-model="stockTradeDate" clearable placeholder="trade_date(YYYY-MM-DD，可选)" style="width: 220px" />
            <el-button type="primary" plain :loading="stockGenerating" @click="handleGenerateDailyStocks">
              生成每日推荐
            </el-button>
            <el-button type="primary" plain @click="applyStockFilters">查询</el-button>
            <el-button @click="resetStockFilters">重置</el-button>
            <el-button type="primary" @click="stockDialogVisible = true">新增股票推荐</el-button>
          </div>
        </div>

        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 8px; flex-wrap: wrap">
            <el-select
              v-model="stockQuoteSyncForm.source_key"
              filterable
              allow-create
              default-first-option
              :loading="stockSourceLoading"
              placeholder="选择行情数据源"
              style="width: 260px"
            >
              <el-option
                v-for="item in stockSourceOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
                :disabled="item.disabled"
              />
            </el-select>
            <el-button :loading="stockSourceLoading" @click="fetchStockSourceOptions({ keepMessage: true })">
              刷新数据源
            </el-button>
            <el-tag type="warning" effect="plain">
              默认源：{{ stockDefaultSourceKey || "-" }}
            </el-tag>
            <el-button
              :loading="stockDefaultSourceLoading"
              :disabled="!stockDefaultSourceKey"
              @click="applyStockDefaultSource"
            >
              使用默认源
            </el-button>
            <el-switch
              v-model="stockAutoFallback"
              inline-prompt
              active-text="自动回退MOCK"
              inactive-text="不回退"
            />
            <el-tag v-if="currentStockSourceOption" :type="sourceHealthTagType(currentStockSourceOption.health_status)">
              {{
                currentStockSourceOption.health_status
                  ? `健康：${currentStockSourceOption.health_status}`
                  : "健康：未检查"
              }}
            </el-tag>
            <el-text v-if="currentStockSourceOption?.health_message" type="info" size="small">
              {{ currentStockSourceOption.health_message }}
            </el-text>
            <el-input
              v-model="stockQuoteSyncForm.symbols"
              placeholder="股票代码，逗号分隔（可选）"
              style="width: 280px"
            />
            <el-input-number
              v-model="stockQuoteSyncForm.days"
              :min="20"
              :max="365"
              :step="5"
              style="width: 150px"
            />
            <el-button type="primary" plain :loading="stockQuoteSyncing" @click="handleSyncStockQuotes">
              同步行情
            </el-button>

            <el-divider direction="vertical" />

            <el-input-number
              v-model="stockQuantQuery.lookback_days"
              :min="30"
              :max="365"
              :step="10"
              style="width: 150px"
            />
            <el-input-number
              v-model="stockQuantQuery.limit"
              :min="1"
              :max="50"
              style="width: 120px"
            />
            <el-button type="success" plain :loading="stockQuantLoading" @click="refreshStockQuantTop">
              计算Top股票
            </el-button>
            <el-text type="info">最近更新：{{ stockQuantUpdatedAt ? formatDateTime(stockQuantUpdatedAt) : "-" }}</el-text>
          </div>

          <el-table
            :data="stockQuantList"
            border
            stripe
            size="small"
            v-loading="stockQuantLoading"
            empty-text="暂无量化评分结果，请先同步行情"
          >
            <el-table-column prop="rank" label="排名" width="66" />
            <el-table-column prop="symbol" label="代码" min-width="100" />
            <el-table-column prop="name" label="名称" min-width="120" />
            <el-table-column prop="score" label="总分" min-width="90" />
            <el-table-column prop="trend_score" label="趋势分" min-width="88" />
            <el-table-column prop="flow_score" label="资金分" min-width="88" />
            <el-table-column prop="value_score" label="估值分" min-width="88" />
            <el-table-column prop="news_score" label="资讯分" min-width="88" />
            <el-table-column prop="momentum_20" label="20日动量(%)" min-width="110" />
            <el-table-column prop="momentum_5" label="5日动量(%)" min-width="100" />
            <el-table-column prop="volume_ratio" label="量比" min-width="80" />
            <el-table-column prop="net_mf_amount" label="主力净流入" min-width="110" />
            <el-table-column prop="pe_ttm" label="PE(TTM)" min-width="88" />
            <el-table-column prop="pb" label="PB" min-width="72" />
            <el-table-column prop="volatility_20" label="20日波动(%)" min-width="110" />
            <el-table-column prop="drawdown_20" label="20日回撤(%)" min-width="110" />
            <el-table-column prop="risk_level" label="风险" min-width="80">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.risk_level)">{{ row.risk_level }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="信号摘要" min-width="280">
              <template #default="{ row }">
                <div class="stock-reason-cell">
                  <span class="stock-reason-summary">{{ row.reason_summary || "-" }}</span>
                  <div v-if="Array.isArray(row.reasons) && row.reasons.length > 0" class="stock-reason-tags">
                    <el-tag v-for="reason in row.reasons.slice(0, 3)" :key="reason" size="small" type="info">
                      {{ reason }}
                    </el-tag>
                  </div>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 8px; flex-wrap: wrap">
            <el-text type="primary">量化回测评估（TopN组合）</el-text>
            <el-input-number
              v-model="stockEvalQuery.days"
              :min="20"
              :max="365"
              :step="10"
              style="width: 150px"
            />
            <el-input-number
              v-model="stockEvalQuery.top_n"
              :min="1"
              :max="30"
              style="width: 120px"
            />
            <el-button type="success" plain :loading="stockEvalLoading" @click="refreshStockQuantEvaluation">
              刷新评估
            </el-button>
            <el-button plain :loading="stockEvalExporting" @click="exportStockQuantEvaluationCSV">
              导出CSV
            </el-button>
            <el-tag type="info">
              样本日：{{ stockEvalSummary.sample_days || 0 }}，样本数：{{ stockEvalSummary.sample_count || 0 }}
            </el-tag>
            <el-tag type="success">5日均收益：{{ formatPercent(stockEvalSummary.avg_return_5, 2, true) }}</el-tag>
            <el-tag type="success">10日均收益：{{ formatPercent(stockEvalSummary.avg_return_10, 2, true) }}</el-tag>
            <el-tag type="warning">5日命中率：{{ formatPercent(stockEvalSummary.hit_rate_5) }}</el-tag>
            <el-tag type="warning">10日命中率：{{ formatPercent(stockEvalSummary.hit_rate_10) }}</el-tag>
            <el-tag type="danger">5日最大回撤：{{ formatPercent(stockEvalSummary.max_drawdown_5) }}</el-tag>
            <el-tag type="danger">10日最大回撤：{{ formatPercent(stockEvalSummary.max_drawdown_10) }}</el-tag>
          </div>
          <div class="stock-eval-curves">
            <div class="stock-eval-curve-card">
              <div class="stock-eval-curve-title">5日累计收益与超额曲线</div>
              <svg
                :viewBox="`0 0 ${stockEvalCurve5.width} ${stockEvalCurve5.height}`"
                xmlns="http://www.w3.org/2000/svg"
                class="stock-eval-svg"
              >
                <line x1="28" :y1="stockEvalCurve5.height - 28" :x2="stockEvalCurve5.width - 28" :y2="stockEvalCurve5.height - 28" stroke="#d1d5db" />
                <line x1="28" y1="28" x2="28" :y2="stockEvalCurve5.height - 28" stroke="#d1d5db" />
                <line
                  x1="28"
                  :y1="curveZeroLineY(stockEvalCurve5)"
                  :x2="stockEvalCurve5.width - 28"
                  :y2="curveZeroLineY(stockEvalCurve5)"
                  stroke="#e5e7eb"
                  stroke-dasharray="4 4"
                />
                <path
                  v-for="line in stockEvalCurve5.paths"
                  :key="line.key"
                  :d="line.d"
                  fill="none"
                  :stroke="line.color"
                  stroke-width="2.2"
                  stroke-linecap="round"
                />
              </svg>
              <div class="stock-eval-legend">
                <span v-for="line in stockEvalCurve5.paths" :key="line.key" class="stock-eval-legend-item">
                  <i class="stock-eval-legend-dot" :style="{ backgroundColor: line.color }" />
                  {{ line.label }}
                </span>
              </div>
            </div>
            <div class="stock-eval-curve-card">
              <div class="stock-eval-curve-title">10日累计收益与超额曲线</div>
              <svg
                :viewBox="`0 0 ${stockEvalCurve10.width} ${stockEvalCurve10.height}`"
                xmlns="http://www.w3.org/2000/svg"
                class="stock-eval-svg"
              >
                <line x1="28" :y1="stockEvalCurve10.height - 28" :x2="stockEvalCurve10.width - 28" :y2="stockEvalCurve10.height - 28" stroke="#d1d5db" />
                <line x1="28" y1="28" x2="28" :y2="stockEvalCurve10.height - 28" stroke="#d1d5db" />
                <line
                  x1="28"
                  :y1="curveZeroLineY(stockEvalCurve10)"
                  :x2="stockEvalCurve10.width - 28"
                  :y2="curveZeroLineY(stockEvalCurve10)"
                  stroke="#e5e7eb"
                  stroke-dasharray="4 4"
                />
                <path
                  v-for="line in stockEvalCurve10.paths"
                  :key="line.key"
                  :d="line.d"
                  fill="none"
                  :stroke="line.color"
                  stroke-width="2.2"
                  stroke-linecap="round"
                />
              </svg>
              <div class="stock-eval-legend">
                <span v-for="line in stockEvalCurve10.paths" :key="line.key" class="stock-eval-legend-item">
                  <i class="stock-eval-legend-dot" :style="{ backgroundColor: line.color }" />
                  {{ line.label }}
                </span>
              </div>
            </div>
          </div>
          <el-table
            :data="stockEvalList"
            border
            stripe
            size="small"
            v-loading="stockEvalLoading"
            empty-text="暂无评估数据，请先计算Top股票并生成历史快照"
          >
            <el-table-column prop="trade_date" label="交易日" min-width="110" />
            <el-table-column prop="sample_count" label="样本数" min-width="80" />
            <el-table-column label="5日均收益" min-width="100">
              <template #default="{ row }">
                {{ formatPercent(row.avg_return_5, 2, true) }}
              </template>
            </el-table-column>
            <el-table-column label="5日命中率" min-width="100">
              <template #default="{ row }">
                {{ formatPercent(row.hit_rate_5) }}
              </template>
            </el-table-column>
            <el-table-column label="10日均收益" min-width="110">
              <template #default="{ row }">
                {{ formatPercent(row.avg_return_10, 2, true) }}
              </template>
            </el-table-column>
            <el-table-column label="10日命中率" min-width="110">
              <template #default="{ row }">
                {{ formatPercent(row.hit_rate_10) }}
              </template>
            </el-table-column>
            <el-table-column label="基准5日收益" min-width="110">
              <template #default="{ row }">
                {{ formatPercent(row.benchmark_return, 2, true) }}
              </template>
            </el-table-column>
          </el-table>

          <div class="toolbar" style="margin-top: 10px; margin-bottom: 6px">
            <el-text type="primary">按风险等级表现</el-text>
            <el-text type="info">按 TopN 历史样本分组统计</el-text>
          </div>
          <el-table
            :data="stockEvalRiskList"
            border
            stripe
            size="small"
            v-loading="stockEvalLoading"
            empty-text="暂无风险分组数据"
          >
            <el-table-column prop="risk_level" label="风险等级" min-width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.risk_level)">{{ row.risk_level || "-" }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="sample_count" label="样本数" min-width="80" />
            <el-table-column label="5日均收益" min-width="110">
              <template #default="{ row }">
                {{ formatPercent(row.avg_return_5, 2, true) }}
              </template>
            </el-table-column>
            <el-table-column label="5日命中率" min-width="110">
              <template #default="{ row }">
                {{ formatPercent(row.hit_rate_5) }}
              </template>
            </el-table-column>
            <el-table-column label="10日均收益" min-width="120">
              <template #default="{ row }">
                {{ formatPercent(row.avg_return_10, 2, true) }}
              </template>
            </el-table-column>
            <el-table-column label="10日命中率" min-width="120">
              <template #default="{ row }">
                {{ formatPercent(row.hit_rate_10) }}
              </template>
            </el-table-column>
          </el-table>

          <div class="toolbar" style="margin-top: 10px; margin-bottom: 6px">
            <el-text type="primary">Top10 成分轮动</el-text>
            <el-text type="info">观察每日新增/移除，跟踪稳定性</el-text>
          </div>
          <el-table
            :data="stockEvalRotationList"
            border
            stripe
            size="small"
            v-loading="stockEvalLoading"
            empty-text="暂无成分轮动数据"
          >
            <el-table-column prop="trade_date" label="交易日" min-width="110" />
            <el-table-column prop="stayed_count" label="延续数" min-width="80" />
            <el-table-column prop="changed_count" label="变动数" min-width="80" />
            <el-table-column label="新增成分" min-width="220">
              <template #default="{ row }">
                <div class="stock-reason-tags">
                  <el-tag v-for="symbol in row.entered || []" :key="`${row.trade_date}-in-${symbol}`" size="small" type="success">
                    {{ symbol }}
                  </el-tag>
                  <span v-if="!(row.entered || []).length">-</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="移除成分" min-width="220">
              <template #default="{ row }">
                <div class="stock-reason-tags">
                  <el-tag v-for="symbol in row.exited || []" :key="`${row.trade_date}-out-${symbol}`" size="small" type="danger">
                    {{ symbol }}
                  </el-tag>
                  <span v-if="!(row.exited || []).length">-</span>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 8px">
            <el-text type="primary">行情同步执行日志</el-text>
            <el-text type="info">仅展示 STOCK / SYNC_QUOTES</el-text>
            <el-button :loading="stockSyncLogsLoading" @click="fetchStockSyncLogs({ keepMessage: true })">
              刷新日志
            </el-button>
          </div>
          <el-table
            :data="stockSyncLogs"
            border
            stripe
            size="small"
            v-loading="stockSyncLogsLoading"
            empty-text="暂无行情同步日志"
          >
            <el-table-column label="时间" min-width="165">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column prop="operator_user_id" label="操作人" min-width="120">
              <template #default="{ row }">
                {{ row.operator_user_id || "-" }}
              </template>
            </el-table-column>
            <el-table-column prop="target_id" label="实际数据源" min-width="120" />
            <el-table-column prop="before_value" label="请求数据源" min-width="120">
              <template #default="{ row }">
                {{ row.before_value || "-" }}
              </template>
            </el-table-column>
            <el-table-column prop="sync_count" label="同步条数" min-width="95" />
            <el-table-column prop="reason" label="参数摘要" min-width="180">
              <template #default="{ row }">
                {{ row.reason || "-" }}
              </template>
            </el-table-column>
            <el-table-column prop="after_value" label="执行结果" min-width="120">
              <template #default="{ row }">
                {{ row.after_value || "-" }}
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ stockSyncLogsPage }} 页，共 {{ stockSyncLogsTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="stockSyncLogsPage"
              :page-size="stockSyncLogsPageSize"
              :total="stockSyncLogsTotal"
              @current-change="handleStockSyncLogPageChange"
            />
          </div>
        </div>

        <div class="card">
          <el-table :data="stocks" border stripe v-loading="stockLoading" empty-text="暂无股票推荐">
            <el-table-column prop="id" label="推荐ID" min-width="130" />
            <el-table-column prop="symbol" label="代码" min-width="90" />
            <el-table-column prop="name" label="名称" min-width="120" />
            <el-table-column prop="score" label="评分" min-width="80" />
            <el-table-column prop="risk_level" label="风险等级" min-width="90" />
            <el-table-column prop="position_range" label="仓位建议" min-width="110" />
            <el-table-column prop="valid_from" label="生效起" min-width="150" />
            <el-table-column prop="valid_to" label="生效止" min-width="150" />
            <el-table-column label="状态" min-width="250">
              <template #default="{ row }">
                <div class="inline-actions">
                  <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
                  <el-select
                    v-model="stockDraftStatusMap[row.id]"
                    filterable
                    style="width: 120px"
                  >
                    <el-option v-for="item in stockStatusOptions" :key="item" :label="item" :value="item" />
                  </el-select>
                  <el-button size="small" @click="saveStockStatus(row)">保存</el-button>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="推荐理由" min-width="220">
              <template #default="{ row }">
                {{ row.reason_summary || "-" }}
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ stockPage }} 页，共 {{ stockTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="stockPage"
              :page-size="stockPageSize"
              :total="stockTotal"
              @current-change="handleStockPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="期货策略" name="futures">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select
              v-model="futuresFilters.status"
              clearable
              filterable
              placeholder="状态"
              style="width: 160px"
            >
              <el-option v-for="item in futuresStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-input v-model="futuresFilters.contract" clearable placeholder="合约代码" style="width: 180px" />
            <el-button type="primary" plain @click="applyFuturesFilters">查询</el-button>
            <el-button @click="resetFuturesFilters">重置</el-button>
            <el-button type="primary" @click="futuresDialogVisible = true">新增期货策略</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="futures" border stripe v-loading="futuresLoading" empty-text="暂无期货策略">
            <el-table-column prop="id" label="策略ID" min-width="130" />
            <el-table-column prop="contract" label="合约" min-width="110" />
            <el-table-column prop="name" label="名称" min-width="120" />
            <el-table-column prop="direction" label="方向" min-width="90" />
            <el-table-column prop="risk_level" label="风险等级" min-width="90" />
            <el-table-column prop="position_range" label="仓位建议" min-width="110" />
            <el-table-column prop="valid_from" label="生效起" min-width="150" />
            <el-table-column prop="valid_to" label="生效止" min-width="150" />
            <el-table-column label="状态" min-width="250">
              <template #default="{ row }">
                <div class="inline-actions">
                  <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
                  <el-select
                    v-model="futuresDraftStatusMap[row.id]"
                    filterable
                    style="width: 120px"
                  >
                    <el-option v-for="item in futuresStatusOptions" :key="item" :label="item" :value="item" />
                  </el-select>
                  <el-button size="small" @click="saveFuturesStatus(row)">保存</el-button>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="策略说明" min-width="220">
              <template #default="{ row }">
                {{ row.reason_summary || "-" }}
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ futuresPage }} 页，共 {{ futuresTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="futuresPage"
              :page-size="futuresPageSize"
              :total="futuresTotal"
              @current-change="handleFuturesPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="市场事件" name="events">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select
              v-model="eventsFilters.event_type"
              clearable
              filterable
              placeholder="事件类型"
              style="width: 170px"
            >
              <el-option v-for="item in marketEventTypeOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-input v-model="eventsFilters.symbol" clearable placeholder="标的代码" style="width: 200px" />
            <el-button type="primary" plain @click="applyEventFilters">查询</el-button>
            <el-button @click="resetEventFilters">重置</el-button>
            <el-button type="primary" @click="openCreateEventDialog">新增市场事件</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="events" border stripe v-loading="eventsLoading" empty-text="暂无市场事件">
            <el-table-column prop="id" label="事件ID" min-width="130" />
            <el-table-column prop="event_type" label="类型" min-width="100" />
            <el-table-column prop="symbol" label="标的" min-width="110" />
            <el-table-column prop="summary" label="事件摘要" min-width="220" />
            <el-table-column prop="trigger_rule" label="触发规则" min-width="220" />
            <el-table-column prop="source" label="来源" min-width="120">
              <template #default="{ row }">
                {{ row.source || "-" }}
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" min-width="180" />
            <el-table-column label="操作" min-width="100" fixed="right">
              <template #default="{ row }">
                <el-button size="small" @click="openEditEventDialog(row)">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ eventsPage }} 页，共 {{ eventsTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="eventsPage"
              :page-size="eventsPageSize"
              :total="eventsTotal"
              @current-change="handleEventPageChange"
            />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="stockDialogVisible" title="新增股票推荐" width="620px" destroy-on-close>
      <el-form label-width="120px">
        <div class="dialog-grid">
          <el-form-item label="代码" required>
            <el-input v-model="stockForm.symbol" placeholder="如 600519.SH" />
          </el-form-item>
          <el-form-item label="名称" required>
            <el-input v-model="stockForm.name" placeholder="贵州茅台" />
          </el-form-item>
          <el-form-item label="评分" required>
            <el-input-number v-model="stockForm.score" :min="0" :max="100" :step="1" style="width: 100%" />
          </el-form-item>
          <el-form-item label="风险等级" required>
            <el-select v-model="stockForm.risk_level" style="width: 100%">
              <el-option v-for="item in riskLevelOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="仓位建议">
            <el-input v-model="stockForm.position_range" placeholder="10%-20%" />
          </el-form-item>
          <el-form-item label="状态" required>
            <el-select
              v-model="stockForm.status"
              filterable
              style="width: 100%"
            >
              <el-option v-for="item in stockStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="生效起" required>
            <el-date-picker
              v-model="stockForm.valid_from"
              type="datetime"
              placeholder="请选择生效起时间"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
              format="YYYY-MM-DD HH:mm:ss"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="生效止" required>
            <el-date-picker
              v-model="stockForm.valid_to"
              type="datetime"
              placeholder="请选择生效止时间"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
              format="YYYY-MM-DD HH:mm:ss"
              style="width: 100%"
            />
          </el-form-item>
        </div>
        <el-form-item label="推荐理由">
          <el-input
            v-model="stockForm.reason_summary"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
            placeholder="趋势强、估值合理、行业景气上行"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stockDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="stockSubmitting" @click="submitStock">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="futuresDialogVisible" title="新增期货策略" width="620px" destroy-on-close>
      <el-form label-width="120px">
        <div class="dialog-grid">
          <el-form-item label="合约" required>
            <el-input v-model="futuresForm.contract" placeholder="如 IF2406" />
          </el-form-item>
          <el-form-item label="名称">
            <el-input v-model="futuresForm.name" placeholder="股指趋势策略" />
          </el-form-item>
          <el-form-item label="方向" required>
            <el-select v-model="futuresForm.direction" style="width: 100%">
              <el-option v-for="item in directionOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="风险等级" required>
            <el-select v-model="futuresForm.risk_level" style="width: 100%">
              <el-option v-for="item in riskLevelOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="仓位建议">
            <el-input v-model="futuresForm.position_range" placeholder="10%-20%" />
          </el-form-item>
          <el-form-item label="状态" required>
            <el-select
              v-model="futuresForm.status"
              filterable
              style="width: 100%"
            >
              <el-option v-for="item in futuresStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="生效起" required>
            <el-date-picker
              v-model="futuresForm.valid_from"
              type="datetime"
              placeholder="请选择生效起时间"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
              format="YYYY-MM-DD HH:mm:ss"
              style="width: 100%"
            />
          </el-form-item>
          <el-form-item label="生效止" required>
            <el-date-picker
              v-model="futuresForm.valid_to"
              type="datetime"
              placeholder="请选择生效止时间"
              value-format="YYYY-MM-DDTHH:mm:ssZ"
              format="YYYY-MM-DD HH:mm:ss"
              style="width: 100%"
            />
          </el-form-item>
        </div>
        <el-form-item label="策略说明">
          <el-input
            v-model="futuresForm.reason_summary"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
            placeholder="趋势突破+波动率收敛，按风险等级动态止损"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="futuresDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="futuresSubmitting" @click="submitFutures">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="eventDialogVisible"
      :title="eventDialogMode === 'edit' ? `编辑市场事件 ${eventForm.id}` : '新增市场事件'"
      width="620px"
      destroy-on-close
    >
      <el-form label-width="120px">
        <div class="dialog-grid">
          <el-form-item label="事件类型" required>
            <el-select
              v-model="eventForm.event_type"
              filterable
              style="width: 100%"
            >
              <el-option v-for="item in marketEventTypeOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="标的代码" required>
            <el-input v-model="eventForm.symbol" placeholder="如 IF2603 / ALL" />
          </el-form-item>
          <el-form-item label="来源">
            <el-input v-model="eventForm.source" placeholder="exchange/system/manual" />
          </el-form-item>
        </div>
        <el-form-item label="事件摘要" required>
          <el-input
            v-model="eventForm.summary"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
            placeholder="描述事件核心影响"
          />
        </el-form-item>
        <el-form-item label="触发规则" required>
          <el-input
            v-model="eventForm.trigger_rule"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
            placeholder="如：5分钟成交量超过均值2倍"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="eventDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="eventsSubmitting" @click="submitEvent">
          {{ eventDialogMode === "edit" ? "保存" : "创建" }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.inline-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.dialog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 0 12px;
}

.stock-reason-cell {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stock-reason-summary {
  line-height: 1.35;
}

.stock-reason-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.stock-eval-curves {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 10px;
  margin-bottom: 10px;
}

.stock-eval-curve-card {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  padding: 8px 10px 10px;
  background: #fcfcfd;
}

.stock-eval-curve-title {
  font-size: 13px;
  color: #334155;
  margin-bottom: 6px;
}

.stock-eval-svg {
  width: 100%;
  height: 190px;
}

.stock-eval-legend {
  margin-top: 4px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.stock-eval-legend-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #475569;
}

.stock-eval-legend-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}
</style>
