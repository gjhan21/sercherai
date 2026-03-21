<script setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from "vue";
import { useRouter } from "vue-router";
import StrategyEngineConfigPanel from "../components/StrategyEngineConfigPanel.vue";
import {
  batchCheckDataSources,
  compareFuturesStrategyEnginePublishVersions,
  compareStockStrategyEnginePublishVersions,
  createMarketEvent,
  createFuturesStrategy,
  createStockRecommendation,
  ensureMarketRhythmTasks,
  generateDailyFuturesStrategies,
  generateDailyStockRecommendations,
  getFuturesStrategyEnginePublishRecord,
  getFuturesStrategyEnginePublishReplay,
  getStockStrategyEnginePublishRecord,
  getStockStrategyEnginePublishReplay,
  listDataSources,
  listFuturesStrategyEnginePublishHistory,
  listMarketRhythmTasks,
  listNewsArticles,
  listOperationLogs,
  listQuantEvaluation,
  listQuantTopStocks,
  listMarketEvents,
  listFuturesStrategies,
  listSystemConfigs,
  listStockStrategyEnginePublishHistory,
  listStockRecommendations,
  syncFuturesInventory,
  syncFuturesQuotes,
  syncMarketNewsSource,
  syncStockQuotes,
  updateMarketEvent,
  updateMarketRhythmTask,
  updateMarketRhythmTaskStatus,
  updateFuturesStrategyStatus,
  updateStockRecommendationStatus
} from "../api/admin";
import { sanitizeHTML } from "../lib/html";
import { getAccessToken, hasPermission } from "../lib/session";

const router = useRouter();
const activeTab = ref("stocks");
const strategyConfigPanelRef = ref(null);
const errorMessage = ref("");
const message = ref("");
const refreshingAll = ref(false);
const rhythmLoading = ref(false);
const rhythmTaskDate = ref(new Date().toISOString().slice(0, 10));
const rhythmStocks = ref([]);
const rhythmArticles = ref([]);
const rhythmTasks = ref([]);
const rhythmTaskDrafts = ref({});
const rhythmTaskSavingMap = ref({});
const canViewNewsRhythm = hasPermission("news.view");
const canEditMarket = hasPermission("market.edit");
const canEditMarketRhythm = canEditMarket;
const canEditNews = hasPermission("news.edit");
const canCheckDataSourceHealth = hasPermission("data_source.edit");

const stockLoading = ref(false);
const stockSubmitting = ref(false);
const stockGenerating = ref(false);
const stockPublishHistoryLoading = ref(false);
const stockPublishReplayLoading = ref(false);
const stockPublishCompareLoading = ref(false);
const stockPage = ref(1);
const stockPageSize = ref(20);
const stockTotal = ref(0);
const stocks = ref([]);
const stockPublishHistory = ref([]);
const stockSelectedPublishRows = ref([]);
const stockPublishDetailLoading = ref(false);
const stockPublishDetailDialogVisible = ref(false);
const stockPublishReplayDialogVisible = ref(false);
const stockPublishCompareResult = ref(null);
const stockPublishDetailActiveTab = ref("html");
const stockPublishDetailRecord = ref(null);
const stockPublishDetail = ref({
  publish_id: "",
  version: 0,
  created_at: "",
  trade_date: "",
  report_summary: "",
  selected_count: 0,
  asset_keys: [],
  payload_count: 0,
  markdown: "",
  html: "",
  publish_payloads: [],
  report_snapshot: {},
  replay: {
    publish_id: "",
    job_id: "",
    operator: "",
    force_publish: false,
    override_reason: "",
    policy_snapshot: {},
    created_at: "",
    storage_source: "",
    warning_count: 0,
    warning_messages: [],
    vetoed_assets: [],
    invalidated_assets: [],
    notes: []
  }
});
const stockPublishReplayRecord = ref(null);
const stockPublishReplay = ref({
  publish_id: "",
  job_id: "",
  operator: "",
  force_publish: false,
  override_reason: "",
  policy_snapshot: {},
  created_at: "",
  storage_source: "",
  warning_count: 0,
  warning_messages: [],
  vetoed_assets: [],
  invalidated_assets: [],
  notes: []
});
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
  status: "DRAFT",
  reason_summary: "",
  source_type: "MANUAL",
  strategy_version: "manual-v1",
  reviewer: "",
  publisher: "",
  review_note: "",
  performance_label: "PENDING"
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
const stockLastSyncResult = ref(null);
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
const stockPublishDetailHTML = computed(() => sanitizeHTML(stockPublishDetail.value?.html || "<p>暂无报告正文</p>"));

const futuresLoading = ref(false);
const futuresSubmitting = ref(false);
const futuresGenerating = ref(false);
const futuresPublishHistoryLoading = ref(false);
const futuresPublishReplayLoading = ref(false);
const futuresPublishCompareLoading = ref(false);
const futuresPage = ref(1);
const futuresPageSize = ref(20);
const futuresTotal = ref(0);
const futures = ref([]);
const futuresPublishHistory = ref([]);
const futuresSelectedPublishRows = ref([]);
const futuresPublishDetailLoading = ref(false);
const futuresPublishDetailDialogVisible = ref(false);
const futuresPublishReplayDialogVisible = ref(false);
const futuresPublishCompareResult = ref(null);
const futuresPublishDetailActiveTab = ref("html");
const futuresPublishDetailRecord = ref(null);
const futuresPublishDetail = ref({
  publish_id: "",
  version: 0,
  created_at: "",
  trade_date: "",
  report_summary: "",
  selected_count: 0,
  asset_keys: [],
  payload_count: 0,
  markdown: "",
  html: "",
  publish_payloads: [],
  report_snapshot: {},
  replay: {
    publish_id: "",
    job_id: "",
    operator: "",
    force_publish: false,
    override_reason: "",
    policy_snapshot: {},
    created_at: "",
    storage_source: "",
    warning_count: 0,
    warning_messages: [],
    vetoed_assets: [],
    invalidated_assets: [],
    notes: []
  }
});
const futuresPublishReplayRecord = ref(null);
const futuresPublishReplay = ref({
  publish_id: "",
  job_id: "",
  operator: "",
  force_publish: false,
  override_reason: "",
  policy_snapshot: {},
  created_at: "",
  storage_source: "",
  warning_count: 0,
  warning_messages: [],
  vetoed_assets: [],
  invalidated_assets: [],
  notes: []
});
const futuresFilters = reactive({
  status: "",
  contract: ""
});
const futuresTradeDate = ref("");
const futuresDraftStatusMap = ref({});
const futuresQuoteSyncing = ref(false);
const futuresInventorySyncing = ref(false);
const futuresSourceLoading = ref(false);
const futuresSourceOptions = ref([]);
const futuresSourceHealthMap = ref({});
const futuresDefaultSourceLoading = ref(false);
const futuresDefaultSourceKey = ref("TUSHARE");
const futuresLastSyncResult = ref(null);
const futuresInventoryLastSyncResult = ref(null);
const futuresQuoteSyncForm = reactive({
  source_key: "TUSHARE",
  contracts: "",
  days: 120
});
const futuresInventorySyncForm = reactive({
  source_key: "TUSHARE",
  symbols: "",
  days: 30
});
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
const futuresPublishDetailHTML = computed(() =>
  sanitizeHTML(futuresPublishDetail.value?.html || "<p>暂无报告正文</p>")
);

const marketNewsSyncing = ref(false);
const marketNewsSourceLoading = ref(false);
const marketNewsSourceOptions = ref([]);
const marketNewsSourceHealthMap = ref({});
const marketNewsDefaultSourceLoading = ref(false);
const marketNewsDefaultSourceKey = ref("AKSHARE");
const marketNewsLastSyncResult = ref(null);
const marketNewsSyncForm = reactive({
  source_key: "AKSHARE",
  symbols: "",
  days: 7,
  limit: 50
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

const stockStatusOptions = [
  "DRAFT",
  "REVIEW_PENDING",
  "PUBLISHED",
  "TRACKING",
  "HIT_TAKE_PROFIT",
  "HIT_STOP_LOSS",
  "INVALIDATED",
  "REVIEWED",
  "ACTIVE",
  "DISABLED"
];
const stockStatusTransitionMap = {
  DRAFT: ["REVIEW_PENDING", "PUBLISHED", "DISABLED", "INVALIDATED"],
  REVIEW_PENDING: ["DRAFT", "PUBLISHED", "DISABLED", "INVALIDATED"],
  PUBLISHED: ["REVIEW_PENDING", "TRACKING", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED", "REVIEWED", "DISABLED"],
  TRACKING: ["PUBLISHED", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED", "REVIEWED", "DISABLED"],
  HIT_TAKE_PROFIT: ["TRACKING", "REVIEWED", "DISABLED"],
  HIT_STOP_LOSS: ["TRACKING", "REVIEWED", "DISABLED"],
  INVALIDATED: ["DRAFT", "REVIEWED", "DISABLED"],
  REVIEWED: ["PUBLISHED", "TRACKING", "DISABLED"],
  ACTIVE: ["PUBLISHED", "REVIEW_PENDING", "TRACKING", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED", "REVIEWED", "DISABLED"],
  DISABLED: ["DRAFT", "REVIEW_PENDING", "PUBLISHED"]
};
const futuresStatusOptions = ["PUBLISHED", "ACTIVE", "DRAFT", "DISABLED"];
const stockSourceTypeOptions = ["MANUAL", "SYSTEM", "RESEARCH", "PARTNER"];
const stockPerformanceLabelOptions = ["PENDING", "ESTIMATED", "WATCH", "WIN", "LOSS", "FLAT"];
const riskLevelOptions = ["LOW", "MEDIUM", "HIGH"];
const directionOptions = ["LONG", "SHORT", "NEUTRAL"];
const marketEventTypeOptions = ["PRICE", "VOLUME", "VOLATILITY", "POLICY", "FLOW", "OTHER"];
const rhythmTaskStatusOptions = ["TODO", "IN_PROGRESS", "DONE", "BLOCKED"];
const STOCK_DEFAULT_SOURCE_CONFIG_KEY = "stock.quotes.default_source_key";
const STOCK_DEFAULT_SOURCE_FALLBACK = "TUSHARE";
const FUTURES_DEFAULT_SOURCE_CONFIG_KEY = "futures.quotes.default_source_key";
const FUTURES_DEFAULT_SOURCE_FALLBACK = "TUSHARE";
const MARKET_NEWS_DEFAULT_SOURCE_CONFIG_KEY = "market.news.default_source_key";
const MARKET_NEWS_DEFAULT_SOURCE_FALLBACK = "AKSHARE";

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function clearMessages() {
  errorMessage.value = "";
  message.value = "";
}

function ensureCanEditMarket() {
  if (canEditMarket) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法新增、生成、同步或更新策略数据";
  return false;
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

function prettyJSON(value) {
  return JSON.stringify(value ?? null, null, 2);
}

function publishReplayModeLabel(replay) {
  return replay?.force_publish ? "人工覆盖发布" : "按策略发布";
}

function publishReplaySourceLabel(replay) {
  switch (String(replay?.storage_source || "").toUpperCase()) {
    case "LOCAL_ARCHIVED":
      return "本地归档";
    case "REMOTE_ONLY":
      return "远端结果";
    default:
      return "未标记";
  }
}

function normalizeStatus(value) {
  return String(value || "").trim().toUpperCase();
}

function startOfDay(date = new Date()) {
  const copy = new Date(date);
  copy.setHours(0, 0, 0, 0);
  return copy;
}

function endOfDay(date = new Date()) {
  const copy = new Date(date);
  copy.setHours(23, 59, 59, 999);
  return copy;
}

function isWithinDay(value, dayOffset = 0) {
  const timestamp = Date.parse(value || "");
  if (Number.isNaN(timestamp)) {
    return false;
  }
  const target = new Date();
  target.setDate(target.getDate() + dayOffset);
  const start = startOfDay(target).getTime();
  const end = endOfDay(target).getTime();
  return timestamp >= start && timestamp <= end;
}

function isWithinLastDays(value, days) {
  const timestamp = Date.parse(value || "");
  if (Number.isNaN(timestamp)) {
    return false;
  }
  const now = Date.now();
  return now - timestamp <= days * 24 * 60 * 60 * 1000;
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

function formatRhythmTaskStatus(value) {
  const key = String(value || "").trim().toUpperCase();
  if (!key) {
    return "待处理";
  }
  const labels = {
    TODO: "待处理",
    IN_PROGRESS: "处理中",
    DONE: "已完成",
    BLOCKED: "阻塞"
  };
  return labels[key] || key;
}

function rhythmTaskStatusTagType(value) {
  const key = String(value || "").trim().toUpperCase();
  if (key === "DONE") {
    return "success";
  }
  if (key === "IN_PROGRESS") {
    return "warning";
  }
  if (key === "BLOCKED") {
    return "danger";
  }
  return "info";
}

function normalizeRhythmTaskLinks(rawValue) {
  return String(rawValue || "")
    .split(/[\n,，]+/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function buildRhythmTaskDraft(task = {}) {
  return {
    owner: task.owner || "",
    notes: task.notes || "",
    sourceLinksText: Array.isArray(task.source_links) ? task.source_links.join("\n") : "",
    status: String(task.status || "TODO").toUpperCase()
  };
}

function syncRhythmTaskDrafts(items) {
  const next = {};
  (Array.isArray(items) ? items : []).forEach((item) => {
    if (!item?.id) {
      return;
    }
    next[item.id] = buildRhythmTaskDraft(item);
  });
  rhythmTaskDrafts.value = next;
}

function setRhythmTaskSaving(id, loadingState) {
  const next = { ...rhythmTaskSavingMap.value };
  if (loadingState) {
    next[id] = true;
  } else {
    delete next[id];
  }
  rhythmTaskSavingMap.value = next;
}

function replaceRhythmTask(item) {
  if (!item?.id) {
    return;
  }
  rhythmTasks.value = (rhythmTasks.value || []).map((row) => (row.id === item.id ? item : row));
  rhythmTaskDrafts.value = {
    ...rhythmTaskDrafts.value,
    [item.id]: buildRhythmTaskDraft(item)
  };
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
  if (["ACTIVE", "PUBLISHED", "TRACKING", "REVIEWED", "SUCCESS", "LOW", "WIN"].includes(normalized)) {
    return "success";
  }
  if (["DRAFT", "REVIEW_PENDING", "PENDING", "RUNNING", "MEDIUM", "WATCH", "ESTIMATED"].includes(normalized)) {
    return "warning";
  }
  if (["DISABLED", "FAILED", "REJECTED", "HIT_STOP_LOSS", "INVALIDATED", "HIGH", "LOSS"].includes(normalized)) {
    return "danger";
  }
  if (["HIT_TAKE_PROFIT", "FLAT"].includes(normalized)) {
    return "info";
  }
  return "info";
}

function getAllowedStockStatusOptions(currentStatus) {
  const normalized = String(currentStatus || "").trim().toUpperCase();
  const allowed = stockStatusTransitionMap[normalized] || stockStatusOptions;
  return Array.from(new Set([normalized, ...allowed].filter(Boolean)));
}

function resolveDataSourceProvider(item) {
  return String(item?.config?.provider || item?.config?.vendor || "").trim().toUpperCase();
}

function supportsSyncKind(item, kind) {
  const provider = resolveDataSourceProvider(item);
  switch (kind) {
    case "stock_quotes":
      return ["TUSHARE", "AKSHARE", "TICKERMD", "MOCK", "MYSELF"].includes(provider);
    case "futures_quotes":
      return ["TUSHARE", "AKSHARE", "TICKERMD", "MOCK", "MYSELF"].includes(provider);
    case "market_news":
      return ["AKSHARE", "TUSHARE"].includes(provider);
    default:
      return false;
  }
}

function normalizeSourceOption(item, healthMap = {}) {
  const sourceKey = normalizeSourceKey(item?.source_key);
  const name = String(item?.name || "").trim();
  const provider = resolveDataSourceProvider(item);
  const sourceType = String(item?.source_type || "").trim().toUpperCase();
  const status = String(item?.status || "").trim().toUpperCase();
  const healthItem = healthMap[sourceKey] || null;
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

function compareSourceOptions(left, right, preferredValues = []) {
  const normalizedPreferred = preferredValues.map((item) => normalizeSourceKey(item));
  for (const value of normalizedPreferred) {
    if (left.value === value && right.value !== value) {
      return -1;
    }
    if (left.value !== value && right.value === value) {
      return 1;
    }
  }
  return left.value.localeCompare(right.value);
}

function buildSyntheticSourceOption(value, label, extra = {}) {
  return {
    value: normalizeSourceKey(value),
    label,
    status: "ACTIVE",
    source_type: "SYSTEM",
    health_status: "",
    health_message: "",
    disabled: false,
    synthetic: true,
    ...extra
  };
}

function buildAutoSourceOption(preferredValues = []) {
  const routeText = preferredValues.filter(Boolean).join(" -> ");
  return buildSyntheticSourceOption(
    "AUTO",
    `AUTO · 按优先级自动回退${routeText ? ` (${routeText})` : ""}`,
    {
      health_message: routeText ? `按系统优先级依次尝试：${routeText}` : "按系统优先级依次尝试多个数据源"
    }
  );
}

function isCompositeSourceKey(value) {
  return /[,;| ]/.test(String(value || "").trim());
}

function buildCustomChainSourceOption(value) {
  const normalized = normalizeSourceKey(value);
  return buildSyntheticSourceOption(normalized, `${normalized} · 自定义回退链路`, {
    health_message: "按输入顺序依次尝试多个数据源"
  });
}

async function fetchDefaultSourceKey(configKey, fallback, targetRef, options = {}) {
  const { keepMessage = false, loadingRef = null } = options;
  if (loadingRef) {
    loadingRef.value = true;
  }
  if (!keepMessage) {
    errorMessage.value = "";
    message.value = "";
  }
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
    const parsed = normalizeSourceKey(matched?.config_value);
    targetRef.value = parsed || fallback;
  } catch (error) {
    targetRef.value = fallback;
    if (!keepMessage) {
      errorMessage.value = normalizeErrorMessage(error, "加载默认行情源失败");
    }
  } finally {
    if (loadingRef) {
      loadingRef.value = false;
    }
  }
}

async function fetchSourceOptionsForKind(kind, options = {}) {
  const { keepMessage = false } = options;
  const sourceConfig =
    kind === "stock_quotes"
      ? {
          loadingRef: stockSourceLoading,
          optionsRef: stockSourceOptions,
          healthMapRef: stockSourceHealthMap,
          defaultKeyRef: stockDefaultSourceKey,
          defaultConfigKey: STOCK_DEFAULT_SOURCE_CONFIG_KEY,
          defaultFallback: STOCK_DEFAULT_SOURCE_FALLBACK,
          form: stockQuoteSyncForm,
          preferredValues: ["TUSHARE", "AKSHARE", "TICKERMD", "MOCK"],
          errorLabel: "加载股票行情数据源失败"
        }
      : kind === "futures_quotes"
        ? {
            loadingRef: futuresSourceLoading,
            optionsRef: futuresSourceOptions,
            healthMapRef: futuresSourceHealthMap,
            defaultKeyRef: futuresDefaultSourceKey,
            defaultConfigKey: FUTURES_DEFAULT_SOURCE_CONFIG_KEY,
            defaultFallback: FUTURES_DEFAULT_SOURCE_FALLBACK,
            form: futuresQuoteSyncForm,
            preferredValues: ["TUSHARE", "TICKERMD", "AKSHARE", "MOCK"],
            errorLabel: "加载期货行情数据源失败"
          }
        : {
            loadingRef: marketNewsSourceLoading,
            optionsRef: marketNewsSourceOptions,
            healthMapRef: marketNewsSourceHealthMap,
            defaultKeyRef: marketNewsDefaultSourceKey,
            defaultConfigKey: MARKET_NEWS_DEFAULT_SOURCE_CONFIG_KEY,
            defaultFallback: MARKET_NEWS_DEFAULT_SOURCE_FALLBACK,
            form: marketNewsSyncForm,
            preferredValues: ["AKSHARE", "TUSHARE"],
            errorLabel: "加载市场资讯数据源失败"
          };
  sourceConfig.loadingRef.value = true;
  if (!keepMessage) {
    errorMessage.value = "";
    message.value = "";
  }
  try {
    await fetchDefaultSourceKey(sourceConfig.defaultConfigKey, sourceConfig.defaultFallback, sourceConfig.defaultKeyRef, {
      keepMessage: true
    });
    const data = await listDataSources({ page: 1, page_size: 200 });
    const rows = Array.isArray(data?.items) ? data.items : [];
    const filteredRows = rows.filter((item) => supportsSyncKind(item, kind));
    const sourceKeys = filteredRows.map((item) => normalizeSourceKey(item?.source_key)).filter(Boolean);
    sourceConfig.healthMapRef.value = {};
    if (sourceKeys.length > 0 && canCheckDataSourceHealth) {
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
        sourceConfig.healthMapRef.value = healthMap;
      } catch (healthError) {
        console.warn(`batch check ${kind} data sources failed:`, healthError?.message || healthError);
      }
    }
    const normalized = filteredRows
      .map((item) => normalizeSourceOption(item, sourceConfig.healthMapRef.value))
      .filter((item) => item.value);
    const currentSourceKey = normalizeSourceKey(sourceConfig.form.source_key);
    const active = normalized.filter((item) => item.status === "ACTIVE");
    const sortedOptions = (active.length > 0 ? active : normalized)
      .reduce((acc, item) => {
        if (!acc.some((existing) => existing.value === item.value)) {
          acc.push(item);
        }
        return acc;
      }, [])
      .sort((left, right) => compareSourceOptions(left, right, sourceConfig.preferredValues));
    const presetOptions = [buildAutoSourceOption(sourceConfig.preferredValues)];
    if (currentSourceKey && isCompositeSourceKey(currentSourceKey)) {
      presetOptions.push(buildCustomChainSourceOption(currentSourceKey));
    }
    const finalOptions = [...presetOptions, ...sortedOptions].reduce((acc, item) => {
      if (!item?.value) {
        return acc;
      }
      if (!acc.some((existing) => existing.value === item.value)) {
        acc.push(item);
      }
      return acc;
    }, []);
    sourceConfig.optionsRef.value = finalOptions;

    const hasCurrent = finalOptions.some((item) => item.value === currentSourceKey);
    if (finalOptions.length > 0 && (!currentSourceKey || !hasCurrent)) {
      const preferred =
        finalOptions.find((item) => item.value === normalizeSourceKey(sourceConfig.defaultKeyRef.value)) ||
        finalOptions.find((item) => sourceConfig.preferredValues.includes(item.value)) ||
        finalOptions[0];
      sourceConfig.form.source_key = preferred.value;
    } else {
      sourceConfig.form.source_key = currentSourceKey || sourceConfig.form.source_key;
    }
  } catch (error) {
    if (!keepMessage) {
      errorMessage.value = normalizeErrorMessage(error, sourceConfig.errorLabel);
    }
  } finally {
    sourceConfig.loadingRef.value = false;
  }
}

async function fetchStockSourceOptions(options = {}) {
  await fetchSourceOptionsForKind("stock_quotes", options);
}

async function fetchFuturesSourceOptions(options = {}) {
  await fetchSourceOptionsForKind("futures_quotes", options);
}

async function fetchMarketNewsSourceOptions(options = {}) {
  await fetchSourceOptionsForKind("market_news", options);
}

function findCurrentSourceOption(optionsRef, currentValue) {
  const currentKey = normalizeSourceKey(currentValue);
  return optionsRef.value.find((item) => item.value === currentKey) || null;
}

const currentStockSourceOption = computed(() => {
  return findCurrentSourceOption(stockSourceOptions, stockQuoteSyncForm.source_key);
});

const currentFuturesSourceOption = computed(() => {
  return findCurrentSourceOption(futuresSourceOptions, futuresQuoteSyncForm.source_key);
});

const currentMarketNewsSourceOption = computed(() => {
  return findCurrentSourceOption(marketNewsSourceOptions, marketNewsSyncForm.source_key);
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

function syncExecutionStatusTagType(status) {
  const normalized = String(status || "").trim().toUpperCase();
  if (["SUCCESS", "ACTIVE", "HEALTHY"].includes(normalized)) {
    return "success";
  }
  if (["PARTIAL", "WARNING", "IN_PROGRESS"].includes(normalized)) {
    return "warning";
  }
  if (["FAILED", "ERROR", "UNHEALTHY"].includes(normalized)) {
    return "danger";
  }
  return "info";
}

function buildSyncMetricTags(result) {
  const items = [];
  const truthCount = Number(result?.truth_count) || 0;
  const barCount = Number(result?.bar_count) || 0;
  const newsCount = Number(result?.news_count) || 0;
  const inventoryCount = Number(result?.inventory_count) || 0;
  const snapshotCount = Number(result?.snapshot_count) || 0;
  if (truthCount > 0) {
    items.push({ key: "truth", label: `真相源 ${truthCount}`, type: "success" });
  }
  if (barCount > 0) {
    items.push({ key: "bars", label: `原始K线 ${barCount}`, type: "info" });
  }
  if (newsCount > 0) {
    items.push({ key: "news", label: `资讯 ${newsCount}`, type: "success" });
  }
  if (inventoryCount > 0) {
    items.push({ key: "inventory", label: `仓单 ${inventoryCount}`, type: "success" });
  }
  if (snapshotCount > 0) {
    items.push({ key: "snapshots", label: `快照 ${snapshotCount}`, type: "warning" });
  }
  return items;
}

function formatSyncResolvedSourceKeys(result, fallbackSourceKey = "") {
  const values = Array.isArray(result?.resolved_source_keys)
    ? result.resolved_source_keys.map((item) => normalizeSourceKey(item)).filter(Boolean)
    : [];
  if (values.length > 0) {
    return values.join(" -> ");
  }
  return normalizeSourceKey(fallbackSourceKey) || "-";
}

function formatSyncItemMetrics(item) {
  const parts = [];
  const truthCount = Number(item?.truth_count) || 0;
  const barCount = Number(item?.bar_count) || 0;
  const newsCount = Number(item?.news_count) || 0;
  const inventoryCount = Number(item?.inventory_count) || 0;
  const snapshotCount = Number(item?.snapshot_count) || 0;
  if (truthCount > 0) {
    parts.push(`真相源 ${truthCount}`);
  }
  if (barCount > 0) {
    parts.push(`K线 ${barCount}`);
  }
  if (newsCount > 0) {
    parts.push(`资讯 ${newsCount}`);
  }
  if (inventoryCount > 0) {
    parts.push(`仓单 ${inventoryCount}`);
  }
  if (snapshotCount > 0) {
    parts.push(`快照 ${snapshotCount}`);
  }
  return parts.join(" · ") || "无新增计数";
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

async function fetchStockDefaultSourceKey(options = {}) {
  await fetchDefaultSourceKey(STOCK_DEFAULT_SOURCE_CONFIG_KEY, STOCK_DEFAULT_SOURCE_FALLBACK, stockDefaultSourceKey, {
    ...options,
    loadingRef: stockDefaultSourceLoading
  });
}

async function fetchFuturesDefaultSourceKey(options = {}) {
  await fetchDefaultSourceKey(
    FUTURES_DEFAULT_SOURCE_CONFIG_KEY,
    FUTURES_DEFAULT_SOURCE_FALLBACK,
    futuresDefaultSourceKey,
    {
      ...options,
      loadingRef: futuresDefaultSourceLoading
    }
  );
}

async function fetchMarketNewsDefaultSourceKey(options = {}) {
  await fetchDefaultSourceKey(
    MARKET_NEWS_DEFAULT_SOURCE_CONFIG_KEY,
    MARKET_NEWS_DEFAULT_SOURCE_FALLBACK,
    marketNewsDefaultSourceKey,
    {
      ...options,
      loadingRef: marketNewsDefaultSourceLoading
    }
  );
}

function applyStockDefaultSource() {
  const nextSourceKey = normalizeSourceKey(stockDefaultSourceKey.value);
  if (!nextSourceKey) {
    return;
  }
  stockQuoteSyncForm.source_key = nextSourceKey;
  message.value = `已应用默认行情源：${nextSourceKey}`;
}

function applyFuturesDefaultSource() {
  const nextSourceKey = normalizeSourceKey(futuresDefaultSourceKey.value);
  if (!nextSourceKey) {
    return;
  }
  futuresQuoteSyncForm.source_key = nextSourceKey;
  message.value = `已应用期货默认行情源：${nextSourceKey}`;
}

function applyMarketNewsDefaultSource() {
  const nextSourceKey = normalizeSourceKey(marketNewsDefaultSourceKey.value);
  if (!nextSourceKey) {
    return;
  }
  marketNewsSyncForm.source_key = nextSourceKey;
  message.value = `已应用资讯默认数据源：${nextSourceKey}`;
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

async function refreshAfterStockSync() {
  stockSyncLogsPage.value = 1;
  await Promise.all([
    fetchStockQuantTop({ keepMessage: true }),
    fetchStockQuantEvaluation({ keepMessage: true }),
    fetchStocks({ keepMessage: true }),
    fetchStockSyncLogs({ keepMessage: true })
  ]);
}

function listStockFallbackSourceKeys(sourceKey) {
  if (!stockAutoFallback.value) {
    return [];
  }
  const normalized = normalizeSourceKey(sourceKey);
  const candidates = [];
  if (normalized !== "AUTO") {
    candidates.push("AUTO");
  }
  if (normalized !== "MOCK") {
    candidates.push("MOCK");
  }
  return candidates.filter((item, index, items) => item && item !== normalized && items.indexOf(item) === index);
}

function formatFallbackSourceLabel(sourceKey) {
  const normalized = normalizeSourceKey(sourceKey);
  if (normalized === "AUTO") {
    return "AUTO 优先级链路";
  }
  return normalized || "-";
}

async function trySyncStockQuotesFallback(payload, sourceKey, prefixText) {
  const failures = [];
  for (const fallbackSourceKey of listStockFallbackSourceKeys(sourceKey)) {
    try {
      const fallbackData = await syncStockQuotes({
        ...payload,
        source_key: fallbackSourceKey
      });
      stockLastSyncResult.value = buildLastSyncResult(fallbackData);
      message.value = `${prefixText}，已回退 ${formatFallbackSourceLabel(fallbackSourceKey)}，处理 ${fallbackData.count || 0} 条`;
      await refreshAfterStockSync();
      return true;
    } catch (fallbackError) {
      failures.push(`${formatFallbackSourceLabel(fallbackSourceKey)}：${normalizeErrorMessage(fallbackError, "同步失败")}`);
    }
  }
  if (failures.length > 0) {
    errorMessage.value = `${prefixText}；${failures.join("；")}`;
  }
  return false;
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

const rhythmBoardCards = computed(() => {
  const morningStocks = rhythmStocks.value.filter((item) => {
    const status = normalizeStatus(item.status);
    if (["DRAFT", "DISABLED"].includes(status)) {
      return false;
    }
    return isWithinDay(item.valid_from || item.created_at, 0);
  });
  const middayArticles = rhythmArticles.value.filter((item) =>
    isWithinDay(item.published_at || item.updated_at || item.created_at, 0)
  );
  const closeStocks = rhythmStocks.value.filter((item) => {
    const status = normalizeStatus(item.status);
    if (!["TRACKING", "REVIEWED", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED"].includes(status)) {
      return false;
    }
    return isWithinDay(item.updated_at || item.created_at || item.valid_to, 0);
  });
  const weekendItems = rhythmStocks.value.filter((item) => {
    const status = normalizeStatus(item.status);
    if (!["REVIEWED", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED"].includes(status)) {
      return false;
    }
    return isWithinLastDays(item.updated_at || item.created_at || item.valid_to, 7);
  });
  const weekendVIPArticles = rhythmArticles.value.filter(
    (item) =>
      normalizeStatus(item.visibility) === "VIP" &&
      isWithinLastDays(item.published_at || item.updated_at || item.created_at, 7)
  );

  const baseCards = [
    {
      slot: "08:30",
      taskKey: "morning_stock_publish",
      title: "今日主推荐发布",
      type: morningStocks.length > 0 ? "success" : "warning",
      stateText: morningStocks.length > 0 ? `已准备 ${morningStocks.length} 条` : "待生成今日主推荐",
      desc:
        morningStocks.length > 0
          ? "今天可投放的主推荐已进入列表，建议继续检查评分和状态。"
          : "先生成或补录今日主推荐，保证前台 08:30 入口有内容可看。",
      meta: `股票推荐 ${morningStocks.length} 条`,
      actionLabel: "切到股票推荐",
      action: { type: "tab", value: "stocks" }
    },
    {
      slot: "11:30",
      taskKey: "midday_news_publish",
      title: "午盘资讯更新",
      type: !canViewNewsRhythm ? "info" : middayArticles.length > 0 ? "success" : "warning",
      stateText: !canViewNewsRhythm
        ? "当前账号无新闻权限"
        : middayArticles.length > 0
          ? `已发布 ${middayArticles.length} 篇`
          : "待补午盘资讯",
      desc: !canViewNewsRhythm
        ? "如果由专门编辑负责资讯，这里仅提示当前节奏状态。"
        : middayArticles.length > 0
          ? "午盘资讯入口已有内容，建议继续检查摘要、封面和会员可见性。"
          : "建议在新闻管理补充午盘变化提醒，承接前台 11:30 回访。",
      meta: canViewNewsRhythm ? `资讯 ${middayArticles.length} 篇` : "需 news.view 权限",
      actionLabel: canViewNewsRhythm ? "去新闻管理" : "保持关注",
      action: canViewNewsRhythm ? { type: "route", value: "news" } : null
    },
    {
      slot: "15:30",
      taskKey: "close_tracking_review",
      title: "收盘跟踪处理",
      type: closeStocks.length > 0 ? "success" : "warning",
      stateText: closeStocks.length > 0 ? `已处理 ${closeStocks.length} 条` : "待更新收盘跟踪",
      desc:
        closeStocks.length > 0
          ? "已有推荐进入 TRACKING / REVIEWED 等收盘跟踪状态。"
          : "建议在收盘后推进状态流转，保证关注页和会员中心能看到进展。",
      meta: "查看状态流转与复核备注",
      actionLabel: "检查股票状态",
      action: { type: "tab", value: "stocks" }
    },
    {
      slot: "周末",
      taskKey: "weekend_review_digest",
      title: "周度复盘清单",
      type: weekendItems.length + weekendVIPArticles.length > 0 ? "success" : "info",
      stateText:
        weekendItems.length + weekendVIPArticles.length > 0
          ? `复盘素材 ${weekendItems.length + weekendVIPArticles.length} 项`
          : "建议提前准备复盘素材",
      desc:
        weekendItems.length + weekendVIPArticles.length > 0
          ? "股票闭环状态和 VIP 资讯已形成周末复盘素材。"
          : "建议提前沉淀已验证推荐与周观点文章，支撑周末留存入口。",
      meta: `股票复盘 ${weekendItems.length} 条 · VIP资讯 ${weekendVIPArticles.length} 篇`,
      actionLabel: canViewNewsRhythm ? "去新闻管理" : "回股票推荐",
      action: canViewNewsRhythm ? { type: "route", value: "news" } : { type: "tab", value: "stocks" }
    }
  ];

  const taskMap = {};
  rhythmTasks.value.forEach((item) => {
    const key = `${item.slot}::${item.task_key}`;
    taskMap[key] = item;
  });

  return baseCards.map((card) => {
    const task = taskMap[`${card.slot}::${card.taskKey}`] || null;
    const draft = task?.id
      ? rhythmTaskDrafts.value[task.id] || buildRhythmTaskDraft(task)
      : buildRhythmTaskDraft({ status: "TODO" });
    return {
      ...card,
      task,
      draft,
      taskStatus: String(task?.status || draft.status || "TODO").toUpperCase(),
      taskStatusLabel: formatRhythmTaskStatus(task?.status || draft.status || "TODO"),
      taskStatusType: rhythmTaskStatusTagType(task?.status || draft.status || "TODO"),
      ownerText: task?.owner || "待认领",
      completedAtText: task?.completed_at ? formatDateTime(task.completed_at) : "-",
      sourceLinks: Array.isArray(task?.source_links) ? task.source_links : [],
      saving: Boolean(task?.id && rhythmTaskSavingMap.value[task.id])
    };
  });
});

async function fetchRhythmBoard(options = {}) {
  const { keepMessage = false } = options;
  rhythmLoading.value = true;
  if (!keepMessage) {
    errorMessage.value = "";
  }
  const tasks = [
    canEditMarketRhythm
      ? ensureMarketRhythmTasks({ date: rhythmTaskDate.value })
      : listMarketRhythmTasks({ date: rhythmTaskDate.value }),
    listStockRecommendations({ page: 1, page_size: 100 }),
    canViewNewsRhythm ? listNewsArticles({ status: "PUBLISHED", page: 1, page_size: 100 }) : Promise.resolve({ items: [] })
  ];
  const [taskResult, stockResult, newsResult] = await Promise.allSettled(tasks);
  const errors = [];

  if (taskResult.status === "fulfilled") {
    rhythmTaskDate.value = String(taskResult.value?.date || rhythmTaskDate.value);
    rhythmTasks.value = Array.isArray(taskResult.value?.items) ? taskResult.value.items : [];
    syncRhythmTaskDrafts(rhythmTasks.value);
  } else {
    rhythmTasks.value = [];
    rhythmTaskDrafts.value = {};
    errors.push(normalizeErrorMessage(taskResult.reason, "加载运营任务失败"));
  }

  if (stockResult.status === "fulfilled") {
    rhythmStocks.value = Array.isArray(stockResult.value?.items) ? stockResult.value.items : [];
  } else {
    rhythmStocks.value = [];
    errors.push(normalizeErrorMessage(stockResult.reason, "加载节奏股票数据失败"));
  }

  if (newsResult.status === "fulfilled") {
    rhythmArticles.value = Array.isArray(newsResult.value?.items) ? newsResult.value.items : [];
  } else {
    rhythmArticles.value = [];
    if (canViewNewsRhythm) {
      errors.push(normalizeErrorMessage(newsResult.reason, "加载节奏资讯数据失败"));
    }
  }

  if (errors.length > 0) {
    errorMessage.value = errors.join("；");
  }
  rhythmLoading.value = false;
}

async function handleSaveRhythmTask(item) {
  const taskID = item?.task?.id;
  if (!taskID || !canEditMarketRhythm) {
    return;
  }
  setRhythmTaskSaving(taskID, true);
  clearMessages();
  try {
    const draft = rhythmTaskDrafts.value[taskID] || buildRhythmTaskDraft(item.task);
    const updated = await updateMarketRhythmTask(taskID, {
      owner: draft.owner,
      notes: draft.notes,
      source_links: normalizeRhythmTaskLinks(draft.sourceLinksText),
      status: draft.status
    });
    replaceRhythmTask(updated);
    message.value = `${item.title} 已保存为${formatRhythmTaskStatus(updated.status)}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "保存运营任务失败");
  } finally {
    setRhythmTaskSaving(taskID, false);
  }
}

async function handleRhythmTaskStatusAction(item, status) {
  const taskID = item?.task?.id;
  if (!taskID || !canEditMarketRhythm) {
    return;
  }
  setRhythmTaskSaving(taskID, true);
  clearMessages();
  try {
    const draft = rhythmTaskDrafts.value[taskID] || buildRhythmTaskDraft(item.task);
    const updated = await updateMarketRhythmTaskStatus(taskID, {
      status,
      owner: draft.owner,
      notes: draft.notes
    });
    replaceRhythmTask(updated);
    message.value = `${item.title} 状态已更新为${formatRhythmTaskStatus(updated.status)}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "更新运营任务状态失败");
  } finally {
    setRhythmTaskSaving(taskID, false);
  }
}

function handleRhythmAction(action) {
  if (!action?.type) {
    return;
  }
  if (action.type === "tab") {
    activeTab.value = action.value || "stocks";
    return;
  }
  if (action.type === "route" && action.value) {
    router.push({ name: action.value });
  }
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
    status: "DRAFT",
    reason_summary: "",
    source_type: "MANUAL",
    strategy_version: "manual-v1",
    reviewer: "",
    publisher: "",
    review_note: "",
    performance_label: "PENDING"
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

async function fetchStockPublishHistory(options = {}) {
  const { keepMessage = false } = options;
  stockPublishHistoryLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listStockStrategyEnginePublishHistory();
    stockPublishHistory.value = data.items || [];
    stockSelectedPublishRows.value = [];
    stockPublishCompareResult.value = null;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载发布归档失败");
  } finally {
    stockPublishHistoryLoading.value = false;
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
  if (!ensureCanEditMarket()) {
    return;
  }
  const sourceKey =
    normalizeSourceKey(stockQuoteSyncForm.source_key) ||
    normalizeSourceKey(stockDefaultSourceKey.value) ||
    "AUTO";
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
      const handled = await trySyncStockQuotesFallback(payload, sourceKey, `数据源 ${sourceKey} 当前不可用（${reason}）`);
      if (handled) {
        return;
      }
      errorMessage.value = `数据源 ${sourceKey} 当前不可用（${reason}），请切换可用数据源`;
      return;
    }

    const data = await syncStockQuotes(payload);
    stockLastSyncResult.value = buildLastSyncResult(data);
    message.value = `行情同步完成，处理 ${data.count || 0} 条`;
    await refreshAfterStockSync();
  } catch (error) {
    const primaryError = normalizeErrorMessage(error, "同步行情失败");
    const handled = await trySyncStockQuotesFallback(payload, sourceKey, `主数据源 ${sourceKey} 同步失败（${primaryError}）`);
    if (handled) {
      return;
    }
    errorMessage.value = primaryError;
  } finally {
    stockQuoteSyncing.value = false;
  }
}

async function handleSyncFuturesQuotes() {
  if (!ensureCanEditMarket()) {
    return;
  }
  const sourceKey =
    normalizeSourceKey(futuresQuoteSyncForm.source_key) ||
    normalizeSourceKey(futuresDefaultSourceKey.value) ||
    "AUTO";
  const payload = {
    source_key: sourceKey,
    contracts: splitSymbols(futuresQuoteSyncForm.contracts),
    days: Number(futuresQuoteSyncForm.days) || 120
  };
  futuresQuoteSyncing.value = true;
  clearMessages();
  try {
    const selectedSourceOption = futuresSourceOptions.value.find((item) => item.value === sourceKey) || null;
    if (selectedSourceOption?.disabled) {
      const reason = buildSourceUnavailableReason(selectedSourceOption);
      errorMessage.value = `数据源 ${sourceKey} 当前不可用（${reason}），建议改用 AUTO 或其他可用源`;
      return;
    }
    const data = await syncFuturesQuotes(payload);
    futuresLastSyncResult.value = buildLastSyncResult(data);
    message.value = `期货行情同步完成，处理 ${data.count || 0} 条`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "同步期货行情失败");
  } finally {
    futuresQuoteSyncing.value = false;
  }
}

async function handleSyncFuturesInventory() {
  if (!ensureCanEditMarket()) {
    return;
  }
  const sourceKey =
    normalizeSourceKey(futuresInventorySyncForm.source_key) ||
    normalizeSourceKey(futuresDefaultSourceKey.value) ||
    "AUTO";
  const payload = {
    source_key: sourceKey,
    symbols: splitSymbols(futuresInventorySyncForm.symbols),
    days: Number(futuresInventorySyncForm.days) || 30
  };
  futuresInventorySyncing.value = true;
  clearMessages();
  try {
    const selectedSourceOption = futuresSourceOptions.value.find((item) => item.value === sourceKey) || null;
    if (selectedSourceOption?.disabled) {
      const reason = buildSourceUnavailableReason(selectedSourceOption);
      errorMessage.value = `数据源 ${sourceKey} 当前不可用（${reason}），建议改用 AUTO 或其他可用源`;
      return;
    }
    const data = await syncFuturesInventory(payload);
    futuresInventoryLastSyncResult.value = buildLastSyncResult(data);
    message.value = `期货仓单同步完成，处理 ${data.count || 0} 条`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "同步期货仓单失败");
  } finally {
    futuresInventorySyncing.value = false;
  }
}

async function handleSyncMarketNews() {
  if (!ensureCanEditMarket()) {
    return;
  }
  const sourceKey =
    normalizeSourceKey(marketNewsSyncForm.source_key) ||
    normalizeSourceKey(marketNewsDefaultSourceKey.value) ||
    "AUTO";
  const payload = {
    source_key: sourceKey,
    symbols: splitSymbols(marketNewsSyncForm.symbols),
    days: Number(marketNewsSyncForm.days) || 7,
    limit: Number(marketNewsSyncForm.limit) || 50
  };
  marketNewsSyncing.value = true;
  clearMessages();
  try {
    const selectedSourceOption = marketNewsSourceOptions.value.find((item) => item.value === sourceKey) || null;
    if (selectedSourceOption?.disabled) {
      const reason = buildSourceUnavailableReason(selectedSourceOption);
      errorMessage.value = `数据源 ${sourceKey} 当前不可用（${reason}），建议改用 AUTO 或其他可用源`;
      return;
    }
    const data = await syncMarketNewsSource(payload);
    marketNewsLastSyncResult.value = buildLastSyncResult(data);
    message.value = `市场资讯同步完成，处理 ${data.count || 0} 条`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "同步市场资讯失败");
  } finally {
    marketNewsSyncing.value = false;
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

async function fetchFuturesPublishHistory(options = {}) {
  const { keepMessage = false } = options;
  futuresPublishHistoryLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listFuturesStrategyEnginePublishHistory();
    futuresPublishHistory.value = data.items || [];
    futuresSelectedPublishRows.value = [];
    futuresPublishCompareResult.value = null;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载期货发布归档失败");
  } finally {
    futuresPublishHistoryLoading.value = false;
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
  if (!ensureCanEditMarket()) {
    return;
  }
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
    reason_summary: stockForm.reason_summary.trim(),
    source_type: stockForm.source_type.trim().toUpperCase() || "MANUAL",
    strategy_version: stockForm.strategy_version.trim(),
    reviewer: stockForm.reviewer.trim(),
    publisher: stockForm.publisher.trim(),
    review_note: stockForm.review_note.trim(),
    performance_label: stockForm.performance_label.trim().toUpperCase() || "PENDING"
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
  if (!ensureCanEditMarket()) {
    return;
  }
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
  if (!ensureCanEditMarket()) {
    return;
  }
  eventDialogMode.value = "create";
  resetEventForm();
  eventDialogVisible.value = true;
}

function openEditEventDialog(row) {
  if (!ensureCanEditMarket()) {
    return;
  }
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
  if (!ensureCanEditMarket()) {
    return;
  }
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
  if (!ensureCanEditMarket()) {
    return;
  }
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
  if (!ensureCanEditMarket()) {
    return;
  }
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

async function handleGenerateDailyFutures() {
  if (!ensureCanEditMarket()) {
    return;
  }
  futuresGenerating.value = true;
  clearMessages();
  try {
    const data = await generateDailyFuturesStrategies(futuresTradeDate.value.trim());
    const notes = [];
    if (data.publish_id) {
      notes.push(`已归档到 ${data.publish_id}${data.publish_version ? ` (v${data.publish_version})` : ""}`);
    }
    if (data.report_summary) {
      notes.push(data.report_summary);
    }
    if (data.archive_enabled === false) {
      notes.push("当前环境未启用 Strategy Engine 归档，本次不会进入发布归档或作业中心");
    }
    message.value = `已生成每日期货策略 ${data.count || 0} 条${notes.length ? `；${notes.join("；")}` : ""}`;
    await Promise.all([fetchFutures({ keepMessage: true }), fetchFuturesPublishHistory({ keepMessage: true })]);
    await strategyConfigPanelRef.value?.refreshAll?.();
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "生成每日期货策略失败");
  } finally {
    futuresGenerating.value = false;
  }
}

function handleFuturesPublishSelectionChange(rows) {
  futuresSelectedPublishRows.value = Array.isArray(rows) ? rows : [];
}

async function handleViewFuturesPublishDetail(row) {
  const publishID = row?.publish_id;
  if (!publishID) {
    errorMessage.value = "缺少 publish_id，无法查看期货报告";
    return;
  }
  futuresPublishDetailLoading.value = true;
  clearMessages();
  futuresPublishDetailRecord.value = row;
  futuresPublishDetailActiveTab.value = "html";
  try {
    const data = await getFuturesStrategyEnginePublishRecord(publishID);
    futuresPublishDetail.value = {
      publish_id: data.publish_id || "",
      version: data.version || 0,
      created_at: data.created_at || "",
      trade_date: data.trade_date || "",
      report_summary: data.report_summary || "",
      selected_count: data.selected_count || 0,
      asset_keys: data.asset_keys || [],
      payload_count: data.payload_count || 0,
      markdown: data.markdown || "",
      html: data.html || "",
      publish_payloads: data.publish_payloads || [],
      report_snapshot: data.report_snapshot || {},
      replay: {
        publish_id: data.replay?.publish_id || data.publish_id || "",
        job_id: data.replay?.job_id || "",
        operator: data.replay?.operator || "",
        force_publish: Boolean(data.replay?.force_publish),
        override_reason: data.replay?.override_reason || "",
        policy_snapshot: data.replay?.policy_snapshot || {},
        created_at: data.replay?.created_at || "",
        storage_source: data.replay?.storage_source || "",
        warning_count: data.replay?.warning_count || 0,
        warning_messages: data.replay?.warning_messages || [],
        vetoed_assets: data.replay?.vetoed_assets || [],
        invalidated_assets: data.replay?.invalidated_assets || [],
        notes: data.replay?.notes || []
      }
    };
    futuresPublishDetailDialogVisible.value = true;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载期货发布详情失败");
  } finally {
    futuresPublishDetailLoading.value = false;
  }
}

async function handleViewFuturesPublishReplay(row) {
  const publishID = row?.publish_id;
  if (!publishID) {
    errorMessage.value = "缺少 publish_id，无法查看期货复盘";
    return;
  }
  futuresPublishReplayLoading.value = true;
  clearMessages();
  futuresPublishReplayRecord.value = row;
  try {
    const data = await getFuturesStrategyEnginePublishReplay(publishID);
    futuresPublishReplay.value = {
      publish_id: data.publish_id || publishID,
      job_id: data.job_id || "",
      operator: data.operator || "",
      force_publish: Boolean(data.force_publish),
      override_reason: data.override_reason || "",
      policy_snapshot: data.policy_snapshot || {},
      created_at: data.created_at || "",
      storage_source: data.storage_source || "",
      warning_count: data.warning_count || 0,
      warning_messages: data.warning_messages || [],
      vetoed_assets: data.vetoed_assets || [],
      invalidated_assets: data.invalidated_assets || [],
      notes: data.notes || []
    };
    futuresPublishReplayDialogVisible.value = true;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载期货发布复盘失败");
  } finally {
    futuresPublishReplayLoading.value = false;
  }
}

async function handleCompareFuturesPublishVersions() {
  if (futuresSelectedPublishRows.value.length !== 2) {
    errorMessage.value = "请选择两条期货发布归档后再对比";
    return;
  }
  futuresPublishCompareLoading.value = true;
  clearMessages();
  try {
    const [left, right] = [...futuresSelectedPublishRows.value].sort((a, b) => (a.version || 0) - (b.version || 0));
    futuresPublishCompareResult.value = await compareFuturesStrategyEnginePublishVersions({
      left_publish_id: left.publish_id,
      right_publish_id: right.publish_id
    });
    message.value = `已完成期货版本对比：v${left.version || "-"} -> v${right.version || "-"}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "对比期货发布版本失败");
  } finally {
    futuresPublishCompareLoading.value = false;
  }
}

async function handleGenerateDailyStocks() {
  if (!ensureCanEditMarket()) {
    return;
  }
  stockGenerating.value = true;
  clearMessages();
  try {
    const data = await generateDailyStockRecommendations(stockTradeDate.value.trim());
    const notes = [];
    if (data.publish_id) {
      notes.push(`已归档到 ${data.publish_id}${data.publish_version ? ` (v${data.publish_version})` : ""}`);
    }
    if (data.report_summary) {
      notes.push(data.report_summary);
    }
    if (data.archive_enabled === false) {
      notes.push("当前环境未启用 Strategy Engine 归档，本次不会进入发布归档或作业中心");
    }
    message.value = `已生成每日股票推荐 ${data.count || 0} 条${notes.length ? `；${notes.join("；")}` : ""}`;
    await Promise.all([fetchStocks({ keepMessage: true }), fetchStockPublishHistory({ keepMessage: true })]);
    await strategyConfigPanelRef.value?.refreshAll?.();
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "生成每日股票推荐失败");
  } finally {
    stockGenerating.value = false;
  }
}

function handleStockPublishSelectionChange(rows) {
  stockSelectedPublishRows.value = Array.isArray(rows) ? rows : [];
}

async function handleViewStockPublishDetail(row) {
  const publishID = row?.publish_id;
  if (!publishID) {
    errorMessage.value = "缺少 publish_id，无法查看股票报告";
    return;
  }
  stockPublishDetailLoading.value = true;
  clearMessages();
  stockPublishDetailRecord.value = row;
  stockPublishDetailActiveTab.value = "html";
  try {
    const data = await getStockStrategyEnginePublishRecord(publishID);
    stockPublishDetail.value = {
      publish_id: data.publish_id || "",
      version: data.version || 0,
      created_at: data.created_at || "",
      trade_date: data.trade_date || "",
      report_summary: data.report_summary || "",
      selected_count: data.selected_count || 0,
      asset_keys: data.asset_keys || [],
      payload_count: data.payload_count || 0,
      markdown: data.markdown || "",
      html: data.html || "",
      publish_payloads: data.publish_payloads || [],
      report_snapshot: data.report_snapshot || {},
      replay: {
        publish_id: data.replay?.publish_id || data.publish_id || "",
        job_id: data.replay?.job_id || "",
        operator: data.replay?.operator || "",
        force_publish: Boolean(data.replay?.force_publish),
        override_reason: data.replay?.override_reason || "",
        policy_snapshot: data.replay?.policy_snapshot || {},
        created_at: data.replay?.created_at || "",
        storage_source: data.replay?.storage_source || "",
        warning_count: data.replay?.warning_count || 0,
        warning_messages: data.replay?.warning_messages || [],
        vetoed_assets: data.replay?.vetoed_assets || [],
        invalidated_assets: data.replay?.invalidated_assets || [],
        notes: data.replay?.notes || []
      }
    };
    stockPublishDetailDialogVisible.value = true;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载股票发布详情失败");
  } finally {
    stockPublishDetailLoading.value = false;
  }
}

async function handleViewStockPublishReplay(row) {
  const publishID = row?.publish_id;
  if (!publishID) {
    errorMessage.value = "缺少 publish_id，无法查看复盘";
    return;
  }
  stockPublishReplayLoading.value = true;
  clearMessages();
  stockPublishReplayRecord.value = row;
  try {
    const data = await getStockStrategyEnginePublishReplay(publishID);
    stockPublishReplay.value = {
      publish_id: data.publish_id || publishID,
      job_id: data.job_id || "",
      operator: data.operator || "",
      force_publish: Boolean(data.force_publish),
      override_reason: data.override_reason || "",
      policy_snapshot: data.policy_snapshot || {},
      created_at: data.created_at || "",
      storage_source: data.storage_source || "",
      warning_count: data.warning_count || 0,
      warning_messages: data.warning_messages || [],
      vetoed_assets: data.vetoed_assets || [],
      invalidated_assets: data.invalidated_assets || [],
      notes: data.notes || []
    };
    stockPublishReplayDialogVisible.value = true;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载发布复盘失败");
  } finally {
    stockPublishReplayLoading.value = false;
  }
}

async function handleCompareStockPublishVersions() {
  if (stockSelectedPublishRows.value.length !== 2) {
    errorMessage.value = "请选择两条发布归档后再对比";
    return;
  }
  stockPublishCompareLoading.value = true;
  clearMessages();
  try {
    const [left, right] = [...stockSelectedPublishRows.value].sort((a, b) => (a.version || 0) - (b.version || 0));
    stockPublishCompareResult.value = await compareStockStrategyEnginePublishVersions({
      left_publish_id: left.publish_id,
      right_publish_id: right.publish_id
    });
    message.value = `已完成版本对比：v${left.version || "-"} -> v${right.version || "-"}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "对比发布版本失败");
  } finally {
    stockPublishCompareLoading.value = false;
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
      fetchRhythmBoard({ keepMessage: true }),
      fetchStockSourceOptions({ keepMessage: true }),
      fetchMarketNewsSourceOptions({ keepMessage: true }),
      fetchStocks(),
      fetchStockPublishHistory({ keepMessage: true }),
      fetchStockQuantTop(),
      fetchStockQuantEvaluation({ keepMessage: true }),
      fetchStockSyncLogs({ keepMessage: true })
    ]);
    return;
  }
  if (activeTab.value === "futures") {
    await Promise.all([
      fetchFuturesSourceOptions({ keepMessage: true }),
      fetchFutures(),
      fetchFuturesPublishHistory({ keepMessage: true })
    ]);
    return;
  }
  if (activeTab.value === "engine-config") {
    await strategyConfigPanelRef.value?.refreshAll?.();
    return;
  }
  await fetchEvents();
}

async function refreshAll(options = {}) {
  const { silentMessage = false } = options;
  refreshingAll.value = true;
  clearMessages();
  try {
    await Promise.all([
      fetchRhythmBoard({ keepMessage: true }),
      fetchStockSourceOptions({ keepMessage: true }),
      fetchFuturesSourceOptions({ keepMessage: true }),
      fetchMarketNewsSourceOptions({ keepMessage: true }),
      fetchStocks({ keepMessage: true }),
      fetchStockPublishHistory({ keepMessage: true }),
      fetchStockQuantTop({ keepMessage: true }),
      fetchStockQuantEvaluation({ keepMessage: true }),
      fetchStockSyncLogs({ keepMessage: true }),
      fetchFutures({ keepMessage: true }),
      fetchFuturesPublishHistory({ keepMessage: true }),
      fetchEvents({ keepMessage: true })
    ]);
    if (!silentMessage && !message.value && !errorMessage.value) {
      message.value = "策略中心数据已刷新";
    }
  } finally {
    refreshingAll.value = false;
  }
}

onMounted(() => {
  refreshAll({ silentMessage: true });
});

watch(activeTab, async (tab) => {
  if (tab !== "engine-config") {
    return;
  }
  await nextTick();
  await strategyConfigPanelRef.value?.refreshAll?.();
});
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

    <div class="card rhythm-board">
      <div class="toolbar rhythm-board-head">
        <div>
          <h2 class="rhythm-board-title">运营节奏检查台</h2>
          <p class="muted">把 08:30 / 11:30 / 15:30 / 周末 四个动作收敛成可检查、可编辑、可留痕的后台任务流。</p>
        </div>
        <div class="toolbar" style="margin-bottom: 0">
          <el-date-picker
            v-model="rhythmTaskDate"
            type="date"
            value-format="YYYY-MM-DD"
            format="YYYY-MM-DD"
            placeholder="选择日期"
            style="width: 160px"
          />
          <el-button :loading="rhythmLoading" @click="fetchRhythmBoard({ keepMessage: true })">
            {{ canEditMarketRhythm ? "刷新并确保任务" : "刷新检查台" }}
          </el-button>
          <el-button type="primary" plain @click="activeTab = 'stocks'">回股票推荐</el-button>
        </div>
      </div>
      <div class="rhythm-board-grid" v-loading="rhythmLoading">
        <article v-for="item in rhythmBoardCards" :key="item.slot" class="rhythm-board-card">
          <div class="rhythm-board-top">
            <div>
              <p class="rhythm-slot">{{ item.slot }}</p>
              <h3>{{ item.title }}</h3>
            </div>
            <div class="rhythm-board-tags">
              <el-tag :type="item.type">{{ item.stateText }}</el-tag>
              <el-tag :type="item.taskStatusType">{{ item.taskStatusLabel }}</el-tag>
            </div>
          </div>
          <p class="rhythm-desc">{{ item.desc }}</p>
          <p class="rhythm-meta">{{ item.meta }}</p>
          <div class="rhythm-task-summary">
            <span>负责人：{{ item.ownerText }}</span>
            <span>完成时间：{{ item.completedAtText }}</span>
          </div>
          <div v-if="item.sourceLinks.length > 0" class="rhythm-link-list">
            <a
              v-for="link in item.sourceLinks"
              :key="`${item.slot}-${link}`"
              :href="link"
              target="_blank"
              rel="noopener noreferrer"
            >
              {{ link }}
            </a>
          </div>
          <div class="rhythm-task-form">
            <el-select
              v-model="item.draft.status"
              :disabled="!canEditMarketRhythm || item.saving"
              placeholder="任务状态"
            >
              <el-option
                v-for="status in rhythmTaskStatusOptions"
                :key="status"
                :label="formatRhythmTaskStatus(status)"
                :value="status"
              />
            </el-select>
            <el-input
              v-model="item.draft.owner"
              :disabled="!canEditMarketRhythm || item.saving"
              placeholder="负责人"
            />
            <el-input
              v-model="item.draft.notes"
              :disabled="!canEditMarketRhythm || item.saving"
              type="textarea"
              :rows="3"
              placeholder="补充说明、阻塞原因或执行记录"
            />
            <el-input
              v-model="item.draft.sourceLinksText"
              :disabled="!canEditMarketRhythm || item.saving"
              type="textarea"
              :rows="3"
              placeholder="来源链接，支持换行或逗号分隔"
            />
          </div>
          <div v-if="canEditMarketRhythm" class="rhythm-status-actions">
            <el-button
              v-for="status in rhythmTaskStatusOptions"
              :key="`${item.slot}-${status}`"
              size="small"
              plain
              :type="rhythmTaskStatusTagType(status)"
              :disabled="item.saving"
              @click="handleRhythmTaskStatusAction(item, status)"
            >
              {{ formatRhythmTaskStatus(status) }}
            </el-button>
          </div>
          <div class="rhythm-task-actions">
            <el-button
              class="rhythm-button"
              size="small"
              :disabled="!item.action"
              @click="handleRhythmAction(item.action)"
            >
              {{ item.actionLabel }}
            </el-button>
            <el-button
              v-if="canEditMarketRhythm"
              size="small"
              type="primary"
              plain
              :loading="item.saving"
              @click="handleSaveRhythmTask(item)"
            >
              保存任务
            </el-button>
          </div>
        </article>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="toolbar rhythm-board-head">
        <div>
          <h2 class="rhythm-board-title">A/B 实验埋点看板已迁移</h2>
          <p class="muted">实验埋点分析已经独立到左侧菜单，策略中心只保留策略与运营节奏相关内容。</p>
        </div>
        <el-button type="primary" plain @click="router.push({ name: 'experiment-analytics' })">
          前往实验埋点看板
        </el-button>
      </div>
      <el-alert
        title="A/B 实验埋点看板已迁移到独立菜单“实验埋点看板”，后续请在新页面单独查看实验、分组、设备和用户阶段趋势。"
        type="info"
        :closable="false"
        show-icon
      />
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="toolbar rhythm-board-head">
        <div>
          <h2 class="rhythm-board-title">智能选股模块已独立</h2>
          <p class="muted">研究、运行、审核、发布闭环现在统一放到“智能选股”菜单，策略中心继续保留旧每日生成链路和迁移期兜底入口。</p>
        </div>
        <el-button type="primary" plain @click="router.push({ name: 'stock-selection-overview' })">
          前往智能选股
        </el-button>
      </div>
      <el-alert
        title="新模块已经拆分为总览、运行中心、策略配置、候选与组合四页；旧 /admin/stocks/recommendations/generate-daily 仍保留给调度与兜底。"
        type="success"
        :closable="false"
        show-icon
      />
    </div>

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
            <el-button
              v-if="canEditMarket"
              type="primary"
              plain
              :loading="stockGenerating"
              @click="handleGenerateDailyStocks"
            >
              生成每日推荐
            </el-button>
            <el-button type="primary" plain @click="applyStockFilters">查询</el-button>
            <el-button @click="resetStockFilters">重置</el-button>
            <el-button v-if="canEditMarket" type="primary" @click="stockDialogVisible = true">新增股票推荐</el-button>
          </div>
        </div>

        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 8px; flex-wrap: wrap">
            <el-text type="primary">Strategy Engine 发布归档</el-text>
            <el-button :loading="stockPublishHistoryLoading" @click="fetchStockPublishHistory({ keepMessage: true })">
              刷新归档
            </el-button>
            <el-button
              type="primary"
              plain
              :loading="stockPublishCompareLoading"
              :disabled="stockSelectedPublishRows.length !== 2"
              @click="handleCompareStockPublishVersions"
            >
              对比选中版本
            </el-button>
            <el-tag type="info">已选 {{ stockSelectedPublishRows.length }} 条</el-tag>
          </div>

          <el-table
            :data="stockPublishHistory"
            border
            stripe
            size="small"
            row-key="publish_id"
            v-loading="stockPublishHistoryLoading"
            empty-text="暂无发布归档，请先执行每日推荐生成"
            @selection-change="handleStockPublishSelectionChange"
          >
            <el-table-column type="selection" width="48" reserve-selection />
            <el-table-column prop="version" label="版本" width="80" />
            <el-table-column prop="publish_id" label="归档ID" min-width="220" />
            <el-table-column prop="trade_date" label="交易日" width="110" />
            <el-table-column label="生成时间" min-width="168">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column prop="selected_count" label="推荐数" width="86" />
            <el-table-column prop="payload_count" label="Payload" width="90" />
            <el-table-column label="标的" min-width="200">
              <template #default="{ row }">
                <div class="publish-asset-list">
                  <el-tag v-for="item in (row.asset_keys || []).slice(0, 4)" :key="item" size="small" effect="plain">
                    {{ item }}
                  </el-tag>
                  <el-text v-if="(row.asset_keys || []).length > 4" size="small" type="info">
                    +{{ (row.asset_keys || []).length - 4 }}
                  </el-text>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="report_summary" label="报告摘要" min-width="320" show-overflow-tooltip />
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" :loading="stockPublishDetailLoading" @click="handleViewStockPublishDetail(row)">
                  报告
                </el-button>
                <el-button link type="primary" :loading="stockPublishReplayLoading" @click="handleViewStockPublishReplay(row)">
                  复盘
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div v-if="stockPublishCompareResult" class="publish-compare-board">
            <div class="publish-compare-head">
              <div>
                <h3>版本对比结果</h3>
                <p class="muted">
                  v{{ stockPublishCompareResult.left_version || "-" }} -> v{{ stockPublishCompareResult.right_version || "-" }}
                </p>
              </div>
              <div class="publish-compare-tags">
                <el-tag type="success">推荐变化 {{ stockPublishCompareResult.selected_count_delta || 0 }}</el-tag>
                <el-tag type="warning">Payload变化 {{ stockPublishCompareResult.payload_count_delta || 0 }}</el-tag>
                <el-tag type="danger">告警变化 {{ stockPublishCompareResult.warning_count_delta || 0 }}</el-tag>
              </div>
            </div>
            <div class="publish-compare-grid">
              <div class="publish-compare-card">
                <h4>新增标的</h4>
                <div class="publish-asset-list">
                  <el-tag
                    v-for="item in stockPublishCompareResult.added_assets || []"
                    :key="`added-${item}`"
                    size="small"
                    type="success"
                    effect="plain"
                  >
                    {{ item }}
                  </el-tag>
                  <span v-if="!(stockPublishCompareResult.added_assets || []).length" class="muted">无</span>
                </div>
              </div>
              <div class="publish-compare-card">
                <h4>移除标的</h4>
                <div class="publish-asset-list">
                  <el-tag
                    v-for="item in stockPublishCompareResult.removed_assets || []"
                    :key="`removed-${item}`"
                    size="small"
                    type="danger"
                    effect="plain"
                  >
                    {{ item }}
                  </el-tag>
                  <span v-if="!(stockPublishCompareResult.removed_assets || []).length" class="muted">无</span>
                </div>
              </div>
              <div class="publish-compare-card">
                <h4>保留标的</h4>
                <div class="publish-asset-list">
                  <el-tag
                    v-for="item in stockPublishCompareResult.shared_assets || []"
                    :key="`shared-${item}`"
                    size="small"
                    type="info"
                    effect="plain"
                  >
                    {{ item }}
                  </el-tag>
                  <span v-if="!(stockPublishCompareResult.shared_assets || []).length" class="muted">无</span>
                </div>
              </div>
            </div>
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
              active-text="自动回退链路"
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
            <el-button
              v-if="canEditMarket"
              type="primary"
              plain
              :loading="stockQuoteSyncing"
              @click="handleSyncStockQuotes"
            >
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

          <div class="sync-inline-hint">
            <el-text type="info" size="small">
              支持 `AUTO`、`MYSELF` 或自定义链路，如 `TUSHARE,AKSHARE,TICKERMD,MYSELF`，用于按顺序回源。
            </el-text>
          </div>

          <div v-if="stockLastSyncResult" class="sync-result-board">
            <div class="sync-result-head">
              <div>
                <h3>最近一次股票行情同步</h3>
                <p class="muted">
                  请求源 {{ formatRequestedSourceLabel(stockLastSyncResult) }}
                  · 实际源 {{ stockLastSyncResult.source_key || "-" }}
                  · 执行链路 {{ formatSyncResolvedSourceKeys(stockLastSyncResult.result, stockLastSyncResult.source_key) }}
                </p>
              </div>
              <div class="publish-compare-tags">
                <el-tag
                  v-for="item in buildSyncMetricTags(stockLastSyncResult.result)"
                  :key="`stock-sync-tag-${item.key}`"
                  :type="item.type"
                >
                  {{ item.label }}
                </el-tag>
                <el-tag v-if="buildSyncMetricTags(stockLastSyncResult.result).length === 0" type="info">
                  处理 {{ stockLastSyncResult.count || 0 }} 条
                </el-tag>
              </div>
            </div>

            <div class="sync-result-grid">
              <div class="sync-result-card">
                <h4>请求参数</h4>
                <ul class="publish-replay-list">
                  <li>同步窗口：{{ stockLastSyncResult.days || "-" }} 天</li>
                  <li>股票范围：{{ formatSyncRequestScope(stockLastSyncResult.symbols, "默认股票池") }}</li>
                  <li>处理条数：{{ stockLastSyncResult.count || 0 }}</li>
                </ul>
              </div>

              <div class="sync-result-card">
                <h4>分源执行明细</h4>
                <div class="sync-result-source-list">
                  <div
                    v-for="item in (stockLastSyncResult.result?.results || [])"
                    :key="`stock-sync-item-${item.source_key}`"
                    class="sync-result-source-item"
                  >
                    <div class="sync-result-source-head">
                      <strong>{{ item.source_key || "-" }}</strong>
                      <el-tag size="small" :type="syncExecutionStatusTagType(item.status)">
                        {{ item.status || "UNKNOWN" }}
                      </el-tag>
                    </div>
                    <p class="sync-result-source-meta">{{ formatSyncItemMetrics(item) }}</p>
                    <p v-if="item.message && item.message !== 'ok'" class="sync-result-source-message">
                      {{ item.message }}
                    </p>
                  </div>
                  <p v-if="!(stockLastSyncResult.result?.results || []).length" class="muted">暂无分源执行记录</p>
                </div>
              </div>
            </div>
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
            <el-text type="primary">市场资讯抓取</el-text>
            <el-select
              v-model="marketNewsSyncForm.source_key"
              filterable
              allow-create
              default-first-option
              :loading="marketNewsSourceLoading"
              placeholder="选择资讯数据源"
              style="width: 280px"
            >
              <el-option
                v-for="item in marketNewsSourceOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
                :disabled="item.disabled"
              />
            </el-select>
            <el-button :loading="marketNewsSourceLoading" @click="fetchMarketNewsSourceOptions({ keepMessage: true })">
              刷新数据源
            </el-button>
            <el-tag type="warning" effect="plain">
              默认源：{{ marketNewsDefaultSourceKey || "-" }}
            </el-tag>
            <el-button
              :loading="marketNewsDefaultSourceLoading"
              :disabled="!marketNewsDefaultSourceKey"
              @click="applyMarketNewsDefaultSource"
            >
              使用默认源
            </el-button>
            <el-tag
              v-if="currentMarketNewsSourceOption"
              :type="sourceHealthTagType(currentMarketNewsSourceOption.health_status)"
            >
              {{
                currentMarketNewsSourceOption.health_status
                  ? `健康：${currentMarketNewsSourceOption.health_status}`
                  : "健康：未检查"
              }}
            </el-tag>
            <el-text v-if="currentMarketNewsSourceOption?.health_message" type="info" size="small">
              {{ currentMarketNewsSourceOption.health_message }}
            </el-text>
            <el-input
              v-model="marketNewsSyncForm.symbols"
              placeholder="股票代码，逗号分隔（可选）"
              style="width: 260px"
            />
            <el-input-number
              v-model="marketNewsSyncForm.days"
              :min="1"
              :max="90"
              :step="1"
              style="width: 140px"
            />
            <el-input-number
              v-model="marketNewsSyncForm.limit"
              :min="1"
              :max="500"
              :step="10"
              style="width: 140px"
            />
            <el-button
              v-if="canEditMarket"
              type="primary"
              plain
              :loading="marketNewsSyncing"
              @click="handleSyncMarketNews"
            >
              同步市场资讯
            </el-button>
          </div>

          <div class="sync-inline-hint">
            <el-text type="info" size="small">
              支持 `AUTO` 或自定义链路，如 `AKSHARE,TUSHARE`，用于资讯主源失败后的顺序回源。
            </el-text>
          </div>

          <div v-if="marketNewsLastSyncResult" class="sync-result-board">
            <div class="sync-result-head">
              <div>
                <h3>最近一次市场资讯同步</h3>
                <p class="muted">
                  请求源 {{ formatRequestedSourceLabel(marketNewsLastSyncResult) }}
                  · 实际源 {{ marketNewsLastSyncResult.source_key || "-" }}
                  · 执行链路 {{ formatSyncResolvedSourceKeys(marketNewsLastSyncResult.result, marketNewsLastSyncResult.source_key) }}
                </p>
              </div>
              <div class="publish-compare-tags">
                <el-tag
                  v-for="item in buildSyncMetricTags(marketNewsLastSyncResult.result)"
                  :key="`market-news-sync-tag-${item.key}`"
                  :type="item.type"
                >
                  {{ item.label }}
                </el-tag>
                <el-tag v-if="buildSyncMetricTags(marketNewsLastSyncResult.result).length === 0" type="info">
                  处理 {{ marketNewsLastSyncResult.count || 0 }} 条
                </el-tag>
              </div>
            </div>

            <div class="sync-result-grid">
              <div class="sync-result-card">
                <h4>请求参数</h4>
                <ul class="publish-replay-list">
                  <li>回看窗口：{{ marketNewsLastSyncResult.days || "-" }} 天</li>
                  <li>抓取上限：{{ marketNewsLastSyncResult.limit || "-" }} 条</li>
                  <li>关联股票：{{ formatSyncRequestScope(marketNewsLastSyncResult.symbols, "不限股票代码") }}</li>
                </ul>
              </div>

              <div class="sync-result-card">
                <h4>分源执行明细</h4>
                <div class="sync-result-source-list">
                  <div
                    v-for="item in (marketNewsLastSyncResult.result?.results || [])"
                    :key="`market-news-sync-item-${item.source_key}`"
                    class="sync-result-source-item"
                  >
                    <div class="sync-result-source-head">
                      <strong>{{ item.source_key || "-" }}</strong>
                      <el-tag size="small" :type="syncExecutionStatusTagType(item.status)">
                        {{ item.status || "UNKNOWN" }}
                      </el-tag>
                    </div>
                    <p class="sync-result-source-meta">{{ formatSyncItemMetrics(item) }}</p>
                    <p v-if="item.message && item.message !== 'ok'" class="sync-result-source-message">
                      {{ item.message }}
                    </p>
                  </div>
                  <p v-if="!(marketNewsLastSyncResult.result?.results || []).length" class="muted">暂无分源执行记录</p>
                </div>
              </div>
            </div>
          </div>
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
            <el-table-column label="来源/版本" min-width="180">
              <template #default="{ row }">
                <div class="stock-reason-cell">
                  <el-tag size="small" :type="statusTagType(row.source_type)">
                    {{ row.source_type || "-" }}
                  </el-tag>
                  <span class="muted">{{ row.strategy_version || "-" }}</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="状态" min-width="280">
              <template #default="{ row }">
                <div class="inline-actions">
                  <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
                  <template v-if="canEditMarket">
                    <el-select
                      v-model="stockDraftStatusMap[row.id]"
                      filterable
                      style="width: 120px"
                    >
                      <el-option
                        v-for="item in getAllowedStockStatusOptions(row.status)"
                        :key="item"
                        :label="item"
                        :value="item"
                      />
                    </el-select>
                    <el-button size="small" @click="saveStockStatus(row)">保存</el-button>
                  </template>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="责任人/表现" min-width="200">
              <template #default="{ row }">
                <div class="stock-reason-cell">
                  <div class="stock-reason-tags">
                    <el-tag size="small" type="info">发布: {{ row.publisher || "-" }}</el-tag>
                    <el-tag size="small" type="warning">复核: {{ row.reviewer || "-" }}</el-tag>
                    <el-tag size="small" :type="statusTagType(row.performance_label)">
                      {{ row.performance_label || "-" }}
                    </el-tag>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="复核备注" min-width="220">
              <template #default="{ row }">
                {{ row.review_note || "-" }}
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
            <el-input v-model="futuresTradeDate" clearable placeholder="trade_date(YYYY-MM-DD，可选)" style="width: 220px" />
            <el-button v-if="canEditMarket" type="primary" plain :loading="futuresGenerating" @click="handleGenerateDailyFutures">
              生成每日期货策略
            </el-button>
            <el-button type="primary" plain @click="applyFuturesFilters">查询</el-button>
            <el-button @click="resetFuturesFilters">重置</el-button>
            <el-button v-if="canEditMarket" type="primary" @click="futuresDialogVisible = true">新增期货策略</el-button>
          </div>
        </div>

        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 8px; flex-wrap: wrap">
            <el-text type="primary">Strategy Engine 期货发布归档</el-text>
            <el-button :loading="futuresPublishHistoryLoading" @click="fetchFuturesPublishHistory({ keepMessage: true })">
              刷新归档
            </el-button>
            <el-button
              type="primary"
              plain
              :loading="futuresPublishCompareLoading"
              :disabled="futuresSelectedPublishRows.length !== 2"
              @click="handleCompareFuturesPublishVersions"
            >
              对比选中版本
            </el-button>
            <el-tag type="info">已选 {{ futuresSelectedPublishRows.length }} 条</el-tag>
          </div>

          <el-table
            :data="futuresPublishHistory"
            border
            stripe
            size="small"
            row-key="publish_id"
            v-loading="futuresPublishHistoryLoading"
            empty-text="暂无期货发布归档，请先执行期货策略生成"
            @selection-change="handleFuturesPublishSelectionChange"
          >
            <el-table-column type="selection" width="48" reserve-selection />
            <el-table-column prop="version" label="版本" width="80" />
            <el-table-column prop="publish_id" label="归档ID" min-width="220" />
            <el-table-column prop="trade_date" label="交易日" width="110" />
            <el-table-column label="生成时间" min-width="168">
              <template #default="{ row }">
                {{ formatDateTime(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column prop="selected_count" label="策略数" width="86" />
            <el-table-column prop="payload_count" label="Payload" width="90" />
            <el-table-column label="合约" min-width="200">
              <template #default="{ row }">
                <div class="publish-asset-list">
                  <el-tag v-for="item in (row.asset_keys || []).slice(0, 4)" :key="item" size="small" effect="plain">
                    {{ item }}
                  </el-tag>
                  <el-text v-if="(row.asset_keys || []).length > 4" size="small" type="info">
                    +{{ (row.asset_keys || []).length - 4 }}
                  </el-text>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="report_summary" label="报告摘要" min-width="320" show-overflow-tooltip />
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" :loading="futuresPublishDetailLoading" @click="handleViewFuturesPublishDetail(row)">
                  报告
                </el-button>
                <el-button link type="primary" :loading="futuresPublishReplayLoading" @click="handleViewFuturesPublishReplay(row)">
                  复盘
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div v-if="futuresPublishCompareResult" class="publish-compare-board">
            <div class="publish-compare-head">
              <div>
                <h3>期货版本对比结果</h3>
                <p class="muted">
                  v{{ futuresPublishCompareResult.left_version || "-" }} -> v{{ futuresPublishCompareResult.right_version || "-" }}
                </p>
              </div>
              <div class="publish-compare-tags">
                <el-tag type="success">策略变化 {{ futuresPublishCompareResult.selected_count_delta || 0 }}</el-tag>
                <el-tag type="warning">Payload变化 {{ futuresPublishCompareResult.payload_count_delta || 0 }}</el-tag>
                <el-tag type="danger">告警变化 {{ futuresPublishCompareResult.warning_count_delta || 0 }}</el-tag>
              </div>
            </div>
            <div class="publish-compare-grid">
              <div class="publish-compare-card">
                <h4>新增合约</h4>
                <div class="publish-asset-list">
                  <el-tag
                    v-for="item in futuresPublishCompareResult.added_assets || []"
                    :key="`f-added-${item}`"
                    size="small"
                    type="success"
                    effect="plain"
                  >
                    {{ item }}
                  </el-tag>
                  <span v-if="!(futuresPublishCompareResult.added_assets || []).length" class="muted">无</span>
                </div>
              </div>
              <div class="publish-compare-card">
                <h4>移除合约</h4>
                <div class="publish-asset-list">
                  <el-tag
                    v-for="item in futuresPublishCompareResult.removed_assets || []"
                    :key="`f-removed-${item}`"
                    size="small"
                    type="danger"
                    effect="plain"
                  >
                    {{ item }}
                  </el-tag>
                  <span v-if="!(futuresPublishCompareResult.removed_assets || []).length" class="muted">无</span>
                </div>
              </div>
              <div class="publish-compare-card">
                <h4>保留合约</h4>
                <div class="publish-asset-list">
                  <el-tag
                    v-for="item in futuresPublishCompareResult.shared_assets || []"
                    :key="`f-shared-${item}`"
                    size="small"
                    type="info"
                    effect="plain"
                  >
                    {{ item }}
                  </el-tag>
                  <span v-if="!(futuresPublishCompareResult.shared_assets || []).length" class="muted">无</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 8px; flex-wrap: wrap">
            <el-text type="primary">期货行情同步</el-text>
            <el-select
              v-model="futuresQuoteSyncForm.source_key"
              filterable
              allow-create
              default-first-option
              :loading="futuresSourceLoading"
              placeholder="选择期货行情源"
              style="width: 280px"
            >
              <el-option
                v-for="item in futuresSourceOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
                :disabled="item.disabled"
              />
            </el-select>
            <el-button :loading="futuresSourceLoading" @click="fetchFuturesSourceOptions({ keepMessage: true })">
              刷新数据源
            </el-button>
            <el-tag type="warning" effect="plain">
              默认源：{{ futuresDefaultSourceKey || "-" }}
            </el-tag>
            <el-button
              :loading="futuresDefaultSourceLoading"
              :disabled="!futuresDefaultSourceKey"
              @click="applyFuturesDefaultSource"
            >
              使用默认源
            </el-button>
            <el-tag v-if="currentFuturesSourceOption" :type="sourceHealthTagType(currentFuturesSourceOption.health_status)">
              {{
                currentFuturesSourceOption.health_status
                  ? `健康：${currentFuturesSourceOption.health_status}`
                  : "健康：未检查"
              }}
            </el-tag>
            <el-text v-if="currentFuturesSourceOption?.health_message" type="info" size="small">
              {{ currentFuturesSourceOption.health_message }}
            </el-text>
            <el-input
              v-model="futuresQuoteSyncForm.contracts"
              placeholder="合约代码，逗号分隔（可选）"
              style="width: 280px"
            />
            <el-input-number
              v-model="futuresQuoteSyncForm.days"
              :min="20"
              :max="365"
              :step="5"
              style="width: 150px"
            />
            <el-button v-if="canEditMarket" type="primary" plain :loading="futuresQuoteSyncing" @click="handleSyncFuturesQuotes">
              同步期货行情
            </el-button>
          </div>

          <div class="sync-inline-hint">
            <el-text type="info" size="small">
              支持 `AUTO`、`MYSELF` 或自定义链路，如 `TUSHARE,TICKERMD,AKSHARE,MYSELF`，用于按顺序回源。
            </el-text>
          </div>

          <div v-if="futuresLastSyncResult" class="sync-result-board">
            <div class="sync-result-head">
              <div>
                <h3>最近一次期货行情同步</h3>
                <p class="muted">
                  请求源 {{ formatRequestedSourceLabel(futuresLastSyncResult) }}
                  · 实际源 {{ futuresLastSyncResult.source_key || "-" }}
                  · 执行链路 {{ formatSyncResolvedSourceKeys(futuresLastSyncResult.result, futuresLastSyncResult.source_key) }}
                </p>
              </div>
              <div class="publish-compare-tags">
                <el-tag
                  v-for="item in buildSyncMetricTags(futuresLastSyncResult.result)"
                  :key="`futures-sync-tag-${item.key}`"
                  :type="item.type"
                >
                  {{ item.label }}
                </el-tag>
                <el-tag v-if="buildSyncMetricTags(futuresLastSyncResult.result).length === 0" type="info">
                  处理 {{ futuresLastSyncResult.count || 0 }} 条
                </el-tag>
              </div>
            </div>

            <div class="sync-result-grid">
              <div class="sync-result-card">
                <h4>请求参数</h4>
                <ul class="publish-replay-list">
                  <li>同步窗口：{{ futuresLastSyncResult.days || "-" }} 天</li>
                  <li>合约范围：{{ formatSyncRequestScope(futuresLastSyncResult.contracts, "默认期货池") }}</li>
                  <li>处理条数：{{ futuresLastSyncResult.count || 0 }}</li>
                </ul>
              </div>

              <div class="sync-result-card">
                <h4>分源执行明细</h4>
                <div class="sync-result-source-list">
                  <div
                    v-for="item in (futuresLastSyncResult.result?.results || [])"
                    :key="`futures-sync-item-${item.source_key}`"
                    class="sync-result-source-item"
                  >
                    <div class="sync-result-source-head">
                      <strong>{{ item.source_key || "-" }}</strong>
                      <el-tag size="small" :type="syncExecutionStatusTagType(item.status)">
                        {{ item.status || "UNKNOWN" }}
                      </el-tag>
                    </div>
                    <p class="sync-result-source-meta">{{ formatSyncItemMetrics(item) }}</p>
                    <p v-if="item.message && item.message !== 'ok'" class="sync-result-source-message">
                      {{ item.message }}
                    </p>
                  </div>
                  <p v-if="!(futuresLastSyncResult.result?.results || []).length" class="muted">暂无分源执行记录</p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 8px; flex-wrap: wrap">
            <el-text type="primary">期货仓单同步</el-text>
            <el-select
              v-model="futuresInventorySyncForm.source_key"
              filterable
              allow-create
              default-first-option
              :loading="futuresSourceLoading"
              placeholder="选择仓单数据源"
              style="width: 280px"
            >
              <el-option
                v-for="item in futuresSourceOptions"
                :key="`inventory-source-${item.value}`"
                :label="item.label"
                :value="item.value"
                :disabled="item.disabled"
              />
            </el-select>
            <el-input
              v-model="futuresInventorySyncForm.symbols"
              placeholder="品种代码，逗号分隔（如 RB,AU,CU）"
              style="width: 280px"
            />
            <el-input-number
              v-model="futuresInventorySyncForm.days"
              :min="5"
              :max="365"
              :step="5"
              style="width: 150px"
            />
            <el-button v-if="canEditMarket" type="primary" plain :loading="futuresInventorySyncing" @click="handleSyncFuturesInventory">
              同步期货仓单
            </el-button>
          </div>

          <div class="sync-inline-hint">
            <el-text type="info" size="small">
              当前最小闭环先支持 TUSHARE / MOCK，品种代码会自动归一到根符号（如 `RB2405` -> `RB`）。
            </el-text>
          </div>

          <div v-if="futuresInventoryLastSyncResult" class="sync-result-board">
            <div class="sync-result-head">
              <div>
                <h3>最近一次期货仓单同步</h3>
                <p class="muted">
                  请求源 {{ formatRequestedSourceLabel(futuresInventoryLastSyncResult) }}
                  · 实际源 {{ futuresInventoryLastSyncResult.source_key || "-" }}
                  · 执行链路 {{ formatSyncResolvedSourceKeys(futuresInventoryLastSyncResult.result, futuresInventoryLastSyncResult.source_key) }}
                </p>
              </div>
              <div class="publish-compare-tags">
                <el-tag
                  v-for="item in buildSyncMetricTags(futuresInventoryLastSyncResult.result)"
                  :key="`futures-inventory-sync-tag-${item.key}`"
                  :type="item.type"
                >
                  {{ item.label }}
                </el-tag>
                <el-tag v-if="buildSyncMetricTags(futuresInventoryLastSyncResult.result).length === 0" type="info">
                  处理 {{ futuresInventoryLastSyncResult.count || 0 }} 条
                </el-tag>
              </div>
            </div>

            <div class="sync-result-grid">
              <div class="sync-result-card">
                <h4>请求参数</h4>
                <ul class="publish-replay-list">
                  <li>同步窗口：{{ futuresInventoryLastSyncResult.days || "-" }} 天</li>
                  <li>品种范围：{{ formatSyncRequestScope(futuresInventoryLastSyncResult.symbols, "默认仓单池") }}</li>
                  <li>处理条数：{{ futuresInventoryLastSyncResult.count || 0 }}</li>
                </ul>
              </div>

              <div class="sync-result-card">
                <h4>分源执行明细</h4>
                <div class="sync-result-source-list">
                  <div
                    v-for="item in (futuresInventoryLastSyncResult.result?.results || [])"
                    :key="`futures-inventory-sync-item-${item.source_key}`"
                    class="sync-result-source-item"
                  >
                    <div class="sync-result-source-head">
                      <strong>{{ item.source_key || "-" }}</strong>
                      <el-tag size="small" :type="syncExecutionStatusTagType(item.status)">
                        {{ item.status || "UNKNOWN" }}
                      </el-tag>
                    </div>
                    <p class="sync-result-source-meta">{{ formatSyncItemMetrics(item) }}</p>
                    <p v-if="item.message && item.message !== 'ok'" class="sync-result-source-message">
                      {{ item.message }}
                    </p>
                  </div>
                  <p v-if="!(futuresInventoryLastSyncResult.result?.results || []).length" class="muted">暂无分源执行记录</p>
                </div>
              </div>
            </div>
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
                  <template v-if="canEditMarket">
                    <el-select
                      v-model="futuresDraftStatusMap[row.id]"
                      filterable
                      style="width: 120px"
                    >
                      <el-option v-for="item in futuresStatusOptions" :key="item" :label="item" :value="item" />
                    </el-select>
                    <el-button size="small" @click="saveFuturesStatus(row)">保存</el-button>
                  </template>
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

      <el-tab-pane label="引擎配置" name="engine-config">
        <StrategyEngineConfigPanel ref="strategyConfigPanelRef" />
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
            <el-button v-if="canEditMarket" type="primary" @click="openCreateEventDialog">新增市场事件</el-button>
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
                <el-button v-if="canEditMarket" size="small" @click="openEditEventDialog(row)">编辑</el-button>
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

    <el-dialog
      v-model="stockPublishDetailDialogVisible"
      title="股票发布报告"
      width="960px"
      destroy-on-close
    >
      <div class="publish-detail-head">
        <div class="publish-detail-meta">
          <el-tag type="info">归档ID：{{ stockPublishDetail.publish_id || stockPublishDetailRecord?.publish_id || "-" }}</el-tag>
          <el-tag type="primary">版本：v{{ stockPublishDetail.version || stockPublishDetailRecord?.version || "-" }}</el-tag>
          <el-tag type="success">推荐数 {{ stockPublishDetail.selected_count || 0 }}</el-tag>
          <el-tag type="warning">Payload {{ stockPublishDetail.payload_count || 0 }}</el-tag>
        </div>
        <div class="publish-detail-meta">
          <el-tag>交易日：{{ stockPublishDetail.trade_date || "-" }}</el-tag>
          <el-tag>生成时间：{{ formatDateTime(stockPublishDetail.created_at) }}</el-tag>
        </div>
      </div>

      <div class="publish-detail-summary">
        <strong>摘要：</strong>{{ stockPublishDetail.report_summary || "暂无摘要" }}
      </div>

      <div class="publish-asset-list" style="margin-bottom: 12px">
        <el-tag v-for="item in stockPublishDetail.asset_keys || []" :key="`stock-detail-${item}`" size="small" effect="plain">
          {{ item }}
        </el-tag>
        <span v-if="!(stockPublishDetail.asset_keys || []).length" class="muted">暂无标的清单</span>
      </div>

      <div class="publish-detail-body" v-loading="stockPublishDetailLoading">
        <el-tabs v-model="stockPublishDetailActiveTab">
          <el-tab-pane label="HTML报告" name="html">
            <div class="publish-detail-html" v-html="stockPublishDetailHTML" />
          </el-tab-pane>
          <el-tab-pane label="Markdown" name="markdown">
            <pre class="publish-detail-pre">{{ stockPublishDetail.markdown || "暂无 Markdown 报告" }}</pre>
          </el-tab-pane>
          <el-tab-pane label="发布Payload" name="payload">
            <pre class="publish-detail-pre">{{ prettyJSON(stockPublishDetail.publish_payloads || []) }}</pre>
          </el-tab-pane>
          <el-tab-pane label="报告快照" name="snapshot">
            <pre class="publish-detail-pre">{{ prettyJSON(stockPublishDetail.report_snapshot || {}) }}</pre>
          </el-tab-pane>
        </el-tabs>

        <div class="publish-detail-replay">
          <div class="publish-replay-card">
            <h4>复盘摘要</h4>
            <ul class="publish-replay-list">
              <li>警告数：{{ stockPublishDetail.replay?.warning_count || 0 }}</li>
              <li>否决标的：{{ (stockPublishDetail.replay?.vetoed_assets || []).join("、") || "无" }}</li>
              <li>失效标的：{{ (stockPublishDetail.replay?.invalidated_assets || []).join("、") || "无" }}</li>
            </ul>
          </div>
        </div>
      </div>
    </el-dialog>

    <el-dialog
      v-model="stockPublishReplayDialogVisible"
      title="发布复盘"
      width="720px"
      destroy-on-close
    >
      <div class="publish-replay-head">
        <el-tag type="info">归档ID：{{ stockPublishReplay.publish_id || stockPublishReplayRecord?.publish_id || "-" }}</el-tag>
        <el-tag type="primary">版本：v{{ stockPublishReplayRecord?.version || "-" }}</el-tag>
        <el-tag :type="stockPublishReplay.force_publish ? 'danger' : 'success'">{{ publishReplayModeLabel(stockPublishReplay) }}</el-tag>
        <el-tag type="warning">警告 {{ stockPublishReplay.warning_count || 0 }}</el-tag>
      </div>

      <div class="publish-replay-grid" v-loading="stockPublishReplayLoading">
        <div class="publish-replay-card">
          <h4>审计信息</h4>
          <ul class="publish-replay-list">
            <li>来源：{{ publishReplaySourceLabel(stockPublishReplay) }}</li>
            <li>操作人：{{ stockPublishReplay.operator || "-" }}</li>
            <li>操作时间：{{ stockPublishReplay.created_at || "-" }}</li>
            <li>任务ID：{{ stockPublishReplay.job_id || "-" }}</li>
            <li>覆盖原因：{{ stockPublishReplay.override_reason || "无" }}</li>
          </ul>
        </div>
        <div class="publish-replay-card">
          <h4>发布策略快照</h4>
          <pre class="publish-detail-pre publish-detail-pre--compact">{{ prettyJSON(stockPublishReplay.policy_snapshot || {}) }}</pre>
        </div>
        <div class="publish-replay-card">
          <h4>警告信息</h4>
          <ul v-if="(stockPublishReplay.warning_messages || []).length" class="publish-replay-list">
            <li v-for="item in stockPublishReplay.warning_messages" :key="item">{{ item }}</li>
          </ul>
          <p v-else class="muted">无额外警告</p>
        </div>
        <div class="publish-replay-card">
          <h4>风控否决</h4>
          <div class="publish-asset-list">
            <el-tag v-for="item in stockPublishReplay.vetoed_assets || []" :key="item" size="small" type="danger" effect="plain">
              {{ item }}
            </el-tag>
            <span v-if="!(stockPublishReplay.vetoed_assets || []).length" class="muted">无</span>
          </div>
        </div>
        <div class="publish-replay-card">
          <h4>失效条件标的</h4>
          <div class="publish-asset-list">
            <el-tag
              v-for="item in stockPublishReplay.invalidated_assets || []"
              :key="item"
              size="small"
              type="warning"
              effect="plain"
            >
              {{ item }}
            </el-tag>
            <span v-if="!(stockPublishReplay.invalidated_assets || []).length" class="muted">无</span>
          </div>
        </div>
        <div class="publish-replay-card">
          <h4>复盘备注</h4>
          <ul v-if="(stockPublishReplay.notes || []).length" class="publish-replay-list">
            <li v-for="item in stockPublishReplay.notes" :key="item">{{ item }}</li>
          </ul>
          <p v-else class="muted">暂无备注</p>
        </div>
      </div>
    </el-dialog>

    <el-dialog v-model="stockDialogVisible" title="新增股票推荐" width="760px" destroy-on-close>
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
          <el-form-item label="来源类型">
            <el-select
              v-model="stockForm.source_type"
              filterable
              allow-create
              default-first-option
              style="width: 100%"
            >
              <el-option
                v-for="item in stockSourceTypeOptions"
                :key="item"
                :label="item"
                :value="item"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="策略版本">
            <el-input v-model="stockForm.strategy_version" placeholder="manual-v1 / daily-v1" />
          </el-form-item>
          <el-form-item label="发布人">
            <el-input v-model="stockForm.publisher" placeholder="留空默认当前操作人" />
          </el-form-item>
          <el-form-item label="复核人">
            <el-input v-model="stockForm.reviewer" placeholder="如 analyst_01" />
          </el-form-item>
          <el-form-item label="表现标签">
            <el-select
              v-model="stockForm.performance_label"
              filterable
              allow-create
              default-first-option
              style="width: 100%"
            >
              <el-option
                v-for="item in stockPerformanceLabelOptions"
                :key="item"
                :label="item"
                :value="item"
              />
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
        <el-form-item label="复核备注">
          <el-input
            v-model="stockForm.review_note"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
            placeholder="记录复核结论、发布依据或后续跟踪关注点"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stockDialogVisible = false">取消</el-button>
        <el-button v-if="canEditMarket" type="primary" :loading="stockSubmitting" @click="submitStock">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="futuresPublishDetailDialogVisible"
      title="期货发布报告"
      width="960px"
      destroy-on-close
    >
      <div class="publish-detail-head">
        <div class="publish-detail-meta">
          <el-tag type="info">归档ID：{{ futuresPublishDetail.publish_id || futuresPublishDetailRecord?.publish_id || "-" }}</el-tag>
          <el-tag type="primary">版本：v{{ futuresPublishDetail.version || futuresPublishDetailRecord?.version || "-" }}</el-tag>
          <el-tag type="success">策略数 {{ futuresPublishDetail.selected_count || 0 }}</el-tag>
          <el-tag type="warning">Payload {{ futuresPublishDetail.payload_count || 0 }}</el-tag>
        </div>
        <div class="publish-detail-meta">
          <el-tag>交易日：{{ futuresPublishDetail.trade_date || "-" }}</el-tag>
          <el-tag>生成时间：{{ formatDateTime(futuresPublishDetail.created_at) }}</el-tag>
        </div>
      </div>

      <div class="publish-detail-summary">
        <strong>摘要：</strong>{{ futuresPublishDetail.report_summary || "暂无摘要" }}
      </div>

      <div class="publish-asset-list" style="margin-bottom: 12px">
        <el-tag v-for="item in futuresPublishDetail.asset_keys || []" :key="`detail-${item}`" size="small" effect="plain">
          {{ item }}
        </el-tag>
        <span v-if="!(futuresPublishDetail.asset_keys || []).length" class="muted">暂无合约清单</span>
      </div>

      <div class="publish-detail-body" v-loading="futuresPublishDetailLoading">
        <el-tabs v-model="futuresPublishDetailActiveTab">
          <el-tab-pane label="HTML报告" name="html">
            <div class="publish-detail-html" v-html="futuresPublishDetailHTML" />
          </el-tab-pane>
          <el-tab-pane label="Markdown" name="markdown">
            <pre class="publish-detail-pre">{{ futuresPublishDetail.markdown || "暂无 Markdown 报告" }}</pre>
          </el-tab-pane>
          <el-tab-pane label="发布Payload" name="payload">
            <pre class="publish-detail-pre">{{ prettyJSON(futuresPublishDetail.publish_payloads || []) }}</pre>
          </el-tab-pane>
          <el-tab-pane label="报告快照" name="snapshot">
            <pre class="publish-detail-pre">{{ prettyJSON(futuresPublishDetail.report_snapshot || {}) }}</pre>
          </el-tab-pane>
        </el-tabs>

        <div class="publish-detail-replay">
          <div class="publish-replay-card">
            <h4>复盘摘要</h4>
            <ul class="publish-replay-list">
              <li>警告数：{{ futuresPublishDetail.replay?.warning_count || 0 }}</li>
              <li>否决合约：{{ (futuresPublishDetail.replay?.vetoed_assets || []).join("、") || "无" }}</li>
              <li>失效合约：{{ (futuresPublishDetail.replay?.invalidated_assets || []).join("、") || "无" }}</li>
            </ul>
          </div>
        </div>
      </div>
    </el-dialog>

    <el-dialog
      v-model="futuresPublishReplayDialogVisible"
      title="期货发布复盘"
      width="720px"
      destroy-on-close
    >
      <div class="publish-replay-head">
        <el-tag type="info">归档ID：{{ futuresPublishReplay.publish_id || futuresPublishReplayRecord?.publish_id || "-" }}</el-tag>
        <el-tag type="primary">版本：v{{ futuresPublishReplayRecord?.version || "-" }}</el-tag>
        <el-tag :type="futuresPublishReplay.force_publish ? 'danger' : 'success'">{{ publishReplayModeLabel(futuresPublishReplay) }}</el-tag>
        <el-tag type="warning">警告 {{ futuresPublishReplay.warning_count || 0 }}</el-tag>
      </div>

      <div class="publish-replay-grid" v-loading="futuresPublishReplayLoading">
        <div class="publish-replay-card">
          <h4>审计信息</h4>
          <ul class="publish-replay-list">
            <li>来源：{{ publishReplaySourceLabel(futuresPublishReplay) }}</li>
            <li>操作人：{{ futuresPublishReplay.operator || "-" }}</li>
            <li>操作时间：{{ futuresPublishReplay.created_at || "-" }}</li>
            <li>任务ID：{{ futuresPublishReplay.job_id || "-" }}</li>
            <li>覆盖原因：{{ futuresPublishReplay.override_reason || "无" }}</li>
          </ul>
        </div>
        <div class="publish-replay-card">
          <h4>发布策略快照</h4>
          <pre class="publish-detail-pre publish-detail-pre--compact">{{ prettyJSON(futuresPublishReplay.policy_snapshot || {}) }}</pre>
        </div>
        <div class="publish-replay-card">
          <h4>警告信息</h4>
          <ul v-if="(futuresPublishReplay.warning_messages || []).length" class="publish-replay-list">
            <li v-for="item in futuresPublishReplay.warning_messages" :key="item">{{ item }}</li>
          </ul>
          <p v-else class="muted">无额外警告</p>
        </div>
        <div class="publish-replay-card">
          <h4>风控否决</h4>
          <div class="publish-asset-list">
            <el-tag v-for="item in futuresPublishReplay.vetoed_assets || []" :key="item" size="small" type="danger" effect="plain">
              {{ item }}
            </el-tag>
            <span v-if="!(futuresPublishReplay.vetoed_assets || []).length" class="muted">无</span>
          </div>
        </div>
        <div class="publish-replay-card">
          <h4>失效条件合约</h4>
          <div class="publish-asset-list">
            <el-tag
              v-for="item in futuresPublishReplay.invalidated_assets || []"
              :key="item"
              size="small"
              type="warning"
              effect="plain"
            >
              {{ item }}
            </el-tag>
            <span v-if="!(futuresPublishReplay.invalidated_assets || []).length" class="muted">无</span>
          </div>
        </div>
        <div class="publish-replay-card">
          <h4>复盘备注</h4>
          <ul v-if="(futuresPublishReplay.notes || []).length" class="publish-replay-list">
            <li v-for="item in futuresPublishReplay.notes" :key="item">{{ item }}</li>
          </ul>
          <p v-else class="muted">暂无备注</p>
        </div>
      </div>
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
        <el-button v-if="canEditMarket" type="primary" :loading="futuresSubmitting" @click="submitFutures">创建</el-button>
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
        <el-button v-if="canEditMarket" type="primary" :loading="eventsSubmitting" @click="submitEvent">
          {{ eventDialogMode === "edit" ? "保存" : "创建" }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.publish-asset-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.publish-compare-board {
  margin-top: 12px;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
}

.publish-compare-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
  margin-bottom: 12px;
}

.publish-compare-head h3,
.publish-compare-card h4,
.publish-replay-card h4 {
  margin: 0 0 6px;
}

.publish-compare-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.publish-compare-grid,
.publish-replay-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.publish-compare-card,
.publish-replay-card {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 12px;
  background: #fff;
}

.publish-replay-head {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
}

.publish-detail-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.publish-detail-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.publish-detail-summary {
  margin-bottom: 12px;
  padding: 12px;
  border-radius: 10px;
  background: #f8fafc;
  color: #334155;
}

.publish-detail-body {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 280px;
  gap: 12px;
  align-items: start;
}

.publish-detail-html {
  min-height: 360px;
  max-height: 560px;
  overflow: auto;
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
  color: #0f172a;
}

.publish-detail-pre {
  min-height: 360px;
  max-height: 560px;
  overflow: auto;
  margin: 0;
  padding: 16px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  background: #0f172a;
  color: #e2e8f0;
  font-size: 12px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

.publish-detail-pre--compact {
  min-height: 180px;
  max-height: 260px;
}

.publish-detail-replay {
  position: sticky;
  top: 0;
}

.publish-replay-list {
  margin: 0;
  padding-left: 18px;
  color: #475569;
}

@media (max-width: 900px) {
  .publish-detail-body {
    grid-template-columns: 1fr;
  }

  .publish-detail-replay {
    position: static;
  }
}

.rhythm-board {
  margin-bottom: 12px;
}

.rhythm-board-head {
  align-items: flex-start;
}

.rhythm-board-title {
  margin: 0 0 4px;
  font-size: 18px;
}

.rhythm-board-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.rhythm-board-card {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 12px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fafc 100%);
}

.rhythm-board-top {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  align-items: flex-start;
}

.rhythm-board-tags {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 6px;
}

.rhythm-slot {
  margin: 0;
  font-size: 12px;
  font-weight: 600;
  color: #0f766e;
}

.rhythm-board-card h3 {
  margin: 4px 0 0;
  font-size: 16px;
  color: #0f172a;
}

.rhythm-desc {
  margin: 10px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: #475569;
}

.rhythm-meta {
  margin: 8px 0 0;
  font-size: 12px;
  color: #64748b;
}

.rhythm-task-summary {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  font-size: 12px;
  color: #475569;
}

.rhythm-link-list {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.rhythm-link-list a {
  color: #2563eb;
  word-break: break-all;
}

.rhythm-task-form {
  margin-top: 10px;
  display: grid;
  gap: 8px;
}

.rhythm-status-actions,
.rhythm-task-actions {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.rhythm-button {
  margin-top: 0;
}

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

.sync-inline-hint {
  margin-bottom: 10px;
}

.sync-result-board {
  margin-bottom: 12px;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
}

.sync-result-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
  margin-bottom: 12px;
}

.sync-result-head h3,
.sync-result-card h4 {
  margin: 0 0 4px;
}

.sync-result-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 12px;
}

.sync-result-card {
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 12px;
  background: #fff;
}

.sync-result-source-list {
  display: grid;
  gap: 8px;
}

.sync-result-source-item {
  padding: 10px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #f8fafc;
}

.sync-result-source-head {
  display: flex;
  justify-content: space-between;
  gap: 8px;
  align-items: center;
}

.sync-result-source-meta,
.sync-result-source-message {
  margin: 6px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: #475569;
}

.sync-result-source-message {
  color: #9f1239;
}

</style>
