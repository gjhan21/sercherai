<script setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import StrategyEngineConfigPanel from "../components/StrategyEngineConfigPanel.vue";
import {
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
  listFuturesStrategyEnginePublishHistory,
  listMarketRhythmTasks,
  listNewsArticles,
  listMarketEvents,
  listFuturesStrategies,
  listStockStrategyEnginePublishHistory,
  listStockRecommendations,
  updateMarketEvent,
  updateMarketRhythmTask,
  updateMarketRhythmTaskStatus,
  updateFuturesStrategyStatus,
  updateStockRecommendationStatus
} from "../api/admin";
import { DEFAULT_FORECAST_ADMIN_CONFIG, buildForecastL2Summary, buildForecastPublishSummary } from "../lib/forecast-admin";
import { sanitizeHTML } from "../lib/html";
import { normalizeMarketCenterRouteState } from "../lib/market-data-admin";
import { hasPermission } from "../lib/session";

const route = useRoute();
const router = useRouter();
const activeTab = ref("stocks");
const strategyConfigPanelRef = ref(null);
const marketCenterRouteFocusKey = ref("");
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
const stockPublishDetailHTML = computed(() => sanitizeHTML(stockPublishDetail.value?.html || "<p>暂无报告正文</p>"));
const stockForecastSummary = computed(() =>
  buildForecastPublishSummary(stockPublishDetail.value, DEFAULT_FORECAST_ADMIN_CONFIG.advisoryPriorityThreshold)
);
const stockForecastL2Summary = computed(() => buildForecastL2Summary(stockPublishDetail.value));

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
const futuresForecastSummary = computed(() =>
  buildForecastPublishSummary(futuresPublishDetail.value, DEFAULT_FORECAST_ADMIN_CONFIG.advisoryPriorityThreshold)
);
const futuresForecastL2Summary = computed(() => buildForecastL2Summary(futuresPublishDetail.value));

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

function applyMarketCenterRouteState(query = route.query) {
  const state = normalizeMarketCenterRouteState(query);
  activeTab.value = state.tab;
  return state;
}

function buildMarketCenterRouteFocusKey(state = {}) {
  return [state.tab, state.publish_id, state.view, state.job_type, state.policy_id, state.config_type, state.config_id].join("|");
}

function findPublishHistoryRow(rows, publishID) {
  return (Array.isArray(rows) ? rows : []).find((item) => String(item?.publish_id || "").trim() === publishID) || null;
}

async function applyMarketCenterObjectFocus(query = route.query, options = {}) {
  const state = normalizeMarketCenterRouteState(query);
  const focusKey = buildMarketCenterRouteFocusKey(state);
  if (!state.publish_id && !state.policy_id && !state.config_id) {
    marketCenterRouteFocusKey.value = "";
    return;
  }
  if (!options.force && focusKey && focusKey === marketCenterRouteFocusKey.value) {
    return;
  }

  if (state.policy_id || state.config_id) {
    activeTab.value = "engine-config";
    await nextTick();
    await strategyConfigPanelRef.value?.focusStrategyConfigItem?.(state.config_type || (state.policy_id ? "publish-policy" : ""), state.config_id || state.policy_id);
    marketCenterRouteFocusKey.value = focusKey;
    return;
  }

  const publishID = state.publish_id;
  const view = state.view || "detail";
  const targetTab = state.tab === "futures" || state.job_type.includes("FUTURES") ? "futures" : "stocks";
  activeTab.value = targetTab;

  if (targetTab === "futures") {
    let row = findPublishHistoryRow(futuresPublishHistory.value, publishID);
    if (!row) {
      await fetchFuturesPublishHistory({ keepMessage: true });
      row = findPublishHistoryRow(futuresPublishHistory.value, publishID);
    }
    if (row) {
      if (view === "replay") {
        await handleViewFuturesPublishReplay(row);
      } else {
        await handleViewFuturesPublishDetail(row);
      }
      marketCenterRouteFocusKey.value = focusKey;
    }
    return;
  }

  let row = findPublishHistoryRow(stockPublishHistory.value, publishID);
  if (!row) {
    await fetchStockPublishHistory({ keepMessage: true });
    row = findPublishHistoryRow(stockPublishHistory.value, publishID);
  }
  if (row) {
    if (view === "replay") {
      await handleViewStockPublishReplay(row);
    } else {
      await handleViewStockPublishDetail(row);
    }
    marketCenterRouteFocusKey.value = focusKey;
  }
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
      fetchStocks(),
      fetchStockPublishHistory({ keepMessage: true })
    ]);
    return;
  }
  if (activeTab.value === "futures") {
    await Promise.all([
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
      fetchStocks({ keepMessage: true }),
      fetchStockPublishHistory({ keepMessage: true }),
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

onMounted(async () => {
  applyMarketCenterRouteState();
  await refreshAll({ silentMessage: true });
  await applyMarketCenterObjectFocus(route.query, { force: true });
});

watch(activeTab, async (tab) => {
  if (tab !== "engine-config") {
    return;
  }
  await nextTick();
  await strategyConfigPanelRef.value?.refreshAll?.();
});

watch(
  () => route.query,
  async (query) => {
    applyMarketCenterRouteState(query);
    await applyMarketCenterObjectFocus(query);
  }
);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">策略中心</h1>
        <p class="muted">这里只保留运营节奏、市场事件和旧链路兜底入口；智能研究、数据同步和质量排查已拆到独立菜单。</p>
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

    <div class="legacy-entry-grid">
      <div class="card" style="margin-bottom: 12px">
        <div class="toolbar rhythm-board-head">
          <div>
            <h2 class="rhythm-board-title">智能选股已独立</h2>
            <p class="muted">股票研究、运行、候选审核与评估复盘请统一进入独立模块。</p>
          </div>
          <el-button type="primary" plain @click="router.push({ name: 'stock-selection-overview' })">
            前往智能选股
          </el-button>
        </div>
      </div>

      <div class="card" style="margin-bottom: 12px">
        <div class="toolbar rhythm-board-head">
          <div>
            <h2 class="rhythm-board-title">智能期货已独立</h2>
            <p class="muted">期货研究、运行、候选审核与评估复盘请统一进入独立模块。</p>
          </div>
          <el-button type="primary" plain @click="router.push({ name: 'futures-selection-overview' })">
            前往智能期货
          </el-button>
        </div>
      </div>

      <div class="card" style="margin-bottom: 12px">
        <div class="toolbar rhythm-board-head">
          <div>
            <h2 class="rhythm-board-title">数据链路已独立</h2>
            <p class="muted">行情/资讯同步、健康检查、truth 重建和质量日志统一迁到数据源管理。</p>
          </div>
          <el-button type="primary" plain @click="router.push({ name: 'data-sources-governance' })">
            前往数据源管理
          </el-button>
        </div>
      </div>
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
          <div class="section-header">
            <div>
              <h3 style="margin: 0">股票数据链路已迁移</h3>
              <p class="muted" style="margin: 6px 0 0">
                股票行情同步、市场资讯、truth 重建、质量排查、Top 股票量化与回测，统一迁到数据源管理或智能选股，不再在策略中心重复维护。
              </p>
            </div>
            <div class="inline-actions inline-actions--left">
              <el-button type="primary" plain @click="router.push({ name: 'data-sources-sync' })">
                前往数据源管理
              </el-button>
              <el-button type="primary" @click="router.push({ name: 'stock-selection-overview' })">
                前往智能选股
              </el-button>
            </div>
          </div>

          <div class="truth-summary-grid" style="margin-top: 12px">
            <div class="truth-summary-card">
              <div class="truth-summary-card__title">迁移到数据源管理</div>
              <ul class="truth-summary-list">
                <li>股票行情同步与默认源切换</li>
                <li>市场资讯同步与来源回退链路</li>
                <li>本地 truth 派生重建与质量日志排查</li>
                <li>多源健康检查、同步摘要和执行质量概览</li>
              </ul>
            </div>
            <div class="truth-summary-card">
              <div class="truth-summary-card__title">迁移到智能选股</div>
              <ul class="truth-summary-list">
                <li>研究运行、候选审核与发布闭环</li>
                <li>策略模板、因子与规则配置</li>
                <li>评估复盘、版本差异与研究证据</li>
                <li>默认作为股票研究工作台使用</li>
              </ul>
            </div>
            <div class="truth-summary-card">
              <div class="truth-summary-card__title">策略中心仍保留</div>
              <ul class="truth-summary-list">
                <li>旧每日推荐生成入口和人工补录兜底</li>
                <li>Strategy Engine 发布归档和版本对比</li>
                <li>最终推荐结果列表与状态维护</li>
                <li>运营节奏与市场事件协同入口</li>
              </ul>
            </div>
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
          <div class="section-header">
            <div>
              <h3 style="margin: 0">期货数据链路已迁移</h3>
              <p class="muted" style="margin: 6px 0 0">
                期货行情同步、主力映射 truth 重建、质量排查与仓单同步统一迁到数据源管理；研究运行、候选审核与发布统一迁到智能期货。
              </p>
            </div>
            <div class="inline-actions inline-actions--left">
              <el-button type="primary" plain @click="router.push({ name: 'data-sources-sync' })">
                前往数据源管理
              </el-button>
              <el-button type="primary" @click="router.push({ name: 'futures-selection-overview' })">
                前往智能期货
              </el-button>
            </div>
          </div>

          <div class="truth-summary-grid" style="margin-top: 12px">
            <div class="truth-summary-card">
              <div class="truth-summary-card__title">迁移到数据源管理</div>
              <ul class="truth-summary-list">
                <li>期货行情同步、默认源切换和回退链路</li>
                <li>主力映射 truth 重建与质量日志排查</li>
                <li>仓单与库存类数据同步入口</li>
                <li>多源健康检查、执行摘要和质量概览</li>
              </ul>
            </div>
            <div class="truth-summary-card">
              <div class="truth-summary-card__title">迁移到智能期货</div>
              <ul class="truth-summary-list">
                <li>研究运行、候选审核与发布闭环</li>
                <li>策略模板、规则与因子配置</li>
                <li>图谱证据、评估复盘与版本差异</li>
                <li>默认作为期货研究工作台使用</li>
              </ul>
            </div>
            <div class="truth-summary-card">
              <div class="truth-summary-card__title">策略中心仍保留</div>
              <ul class="truth-summary-list">
                <li>旧每日期货策略生成入口和人工补录兜底</li>
                <li>Strategy Engine 期货发布归档和版本对比</li>
                <li>最终策略列表与状态维护</li>
                <li>运营节奏与市场事件协同入口</li>
              </ul>
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

      <div class="publish-detail-summary-card" style="margin-bottom: 12px">
        <strong>预测增强摘要</strong>
        <div class="publish-summary-grid">
          <el-tag type="success">增强 explanation 覆盖 {{ stockForecastSummary.enhancedCount }} / {{ stockForecastSummary.payloadCount || 0 }}</el-tag>
          <el-tag type="info">覆盖率 {{ stockForecastSummary.coverageRatio }}</el-tag>
          <el-tag type="primary">研究编排 {{ stockForecastSummary.researchOutlineCount }}</el-tag>
          <el-tag type="warning">高 advisory 样本 {{ stockForecastSummary.highAdvisoryCount }}</el-tag>
          <el-tag>观察信号 {{ stockForecastSummary.watchSignalCount }}</el-tag>
          <el-tag>记忆反馈 {{ stockForecastSummary.memoryFeedbackCount }}</el-tag>
          <el-tag type="primary">L2 场景 {{ stockForecastSummary.scenarioSnapshotCount }}</el-tag>
          <el-tag>关系节点 {{ stockForecastSummary.relationshipNodeCount }}</el-tag>
          <el-tag :type="stockForecastSummary.vetoedCount > 0 ? 'danger' : 'info'">veto 提示 {{ stockForecastSummary.vetoedCount }}</el-tag>
        </div>
        <p v-if="stockForecastL2Summary" class="muted" style="margin: 8px 0 0">
          主情景 {{ stockForecastL2Summary.primaryScenario || "-" }} · 共识 {{ stockForecastL2Summary.consensusAction || "-" }} ·
          角色 {{ stockForecastL2Summary.topRoles.join(" / ") || "-" }}
          <template v-if="stockForecastL2Summary.vetoed">
            · veto {{ stockForecastL2Summary.vetoReason || "风险角色阻止直接执行" }}
          </template>
        </p>
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

      <div class="publish-detail-summary-card" style="margin-bottom: 12px">
        <strong>预测增强摘要</strong>
        <div class="publish-summary-grid">
          <el-tag type="success">增强 explanation 覆盖 {{ futuresForecastSummary.enhancedCount }} / {{ futuresForecastSummary.payloadCount || 0 }}</el-tag>
          <el-tag type="info">覆盖率 {{ futuresForecastSummary.coverageRatio }}</el-tag>
          <el-tag type="primary">研究编排 {{ futuresForecastSummary.researchOutlineCount }}</el-tag>
          <el-tag type="warning">高 advisory 样本 {{ futuresForecastSummary.highAdvisoryCount }}</el-tag>
          <el-tag>观察信号 {{ futuresForecastSummary.watchSignalCount }}</el-tag>
          <el-tag>记忆反馈 {{ futuresForecastSummary.memoryFeedbackCount }}</el-tag>
          <el-tag type="primary">L2 场景 {{ futuresForecastSummary.scenarioSnapshotCount }}</el-tag>
          <el-tag>关系节点 {{ futuresForecastSummary.relationshipNodeCount }}</el-tag>
          <el-tag :type="futuresForecastSummary.vetoedCount > 0 ? 'danger' : 'info'">veto 提示 {{ futuresForecastSummary.vetoedCount }}</el-tag>
        </div>
        <p v-if="futuresForecastL2Summary" class="muted" style="margin: 8px 0 0">
          主情景 {{ futuresForecastL2Summary.primaryScenario || "-" }} · 共识 {{ futuresForecastL2Summary.consensusAction || "-" }} ·
          角色 {{ futuresForecastL2Summary.topRoles.join(" / ") || "-" }}
          <template v-if="futuresForecastL2Summary.vetoed">
            · veto {{ futuresForecastL2Summary.vetoReason || "风险角色阻止直接执行" }}
          </template>
        </p>
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

.sync-result-issue-tags {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.sync-result-issue-tag {
  cursor: pointer;
}

.sync-result-issue-chip {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.sync-result-issue-link {
  padding: 0;
  min-height: auto;
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

.sync-result-source-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 8px;
}

.sync-result-card-title {
  font-weight: 600;
  color: #111827;
}

.quality-detail-message {
  margin: 8px 0 0;
  color: #374151;
  line-height: 1.7;
}

.publish-detail-summary-card {
  padding: 12px;
  border-radius: 12px;
  background: #f8fafc;
  border: 1px solid #e5e7eb;
}

.publish-summary-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

</style>
