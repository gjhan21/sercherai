<script setup>
import { computed, onMounted, reactive, ref, watch } from "vue";
import { ElMessageBox } from "element-plus";
import { useRoute } from "vue-router";
import {
  createMarketDataBackfillRun,
  createSchedulerJobDefinition,
  deleteSchedulerJobDefinition,
  getMarketDataBackfillRun,
  getSchedulerJobMetrics,
  listMarketDataBackfillRunDetails,
  listMarketDataBackfillRuns,
  listMarketUniverseSnapshots,
  listOperationLogs,
  listNewsSyncRunDetails,
  listSupportedSchedulerJobs,
  listSystemConfigs,
  listSchedulerJobDefinitions,
  listSchedulerJobRuns,
  retryMarketDataBackfillRun,
  retryNewsSyncItem,
  retrySchedulerJobRun,
  triggerSchedulerJob,
  upsertSystemConfig,
  updateSchedulerJobDefinition,
  updateSchedulerJobDefinitionStatus
} from "../api/admin";
import {
  buildMarketBackfillGuideCards,
  buildMarketBackfillOverviewCards,
  buildSystemJobsActionCards,
  buildSystemJobsGuideCards,
  buildSystemJobsOverviewCards,
  buildSystemJobsTabOptions,
  buildSchedulerDefinitionCreateOptions,
  buildSchedulerDefinitionOptions,
  validateSchedulerDefinitionJobName
} from "../lib/system-jobs-admin";
import {
  buildUniverseSnapshotDigest,
  formatMarketAssetLabel,
  formatMarketAssetScopeSummary,
  formatMarketBackfillRunTypeLabel,
  formatMarketBackfillStageLabel,
  formatMarketBackfillStatusLabel,
  marketBackfillDetailStatusTagType,
  marketBackfillStatusTagType
} from "../lib/market-data-admin";
import { getAccessToken, hasPermission } from "../lib/session";

const route = useRoute();

const metricsLoading = ref(false);
const defsLoading = ref(false);
const supportedJobsLoading = ref(false);
const runsLoading = ref(false);
const exportingRuns = ref(false);
const autoRetryLoading = ref(false);
const autoRetryLogLoading = ref(false);
const savingAutoRetry = ref(false);
const submittingDefinition = ref(false);
const triggeringJob = ref(false);
const copyingRunText = ref(false);
const batchRetryingFailed = ref(false);
const batchRetryConcurrency = ref(3);

const errorMessage = ref("");
const message = ref("");
const canEditSystemJobs = hasPermission("system_job.edit");

const metricFilter = reactive({
  job_name: ""
});

function ensureCanEditSystemJobs() {
  if (canEditSystemJobs) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法修改任务配置、手动触发或重跑任务";
  return false;
}

const metrics = ref({
  today_total: 0,
  today_success: 0,
  today_failed: 0,
  today_running: 0,
  retry_total: 0,
  retry_success: 0,
  retry_failed: 0,
  retry_hit_rate: 0,
  avg_retry_count: 0,
  auto_retry_total: 0,
  recovery_total: 0,
  recovery_success: 0,
  recovery_hit_rate: 0,
  failure_reasons: [],
  job_retry_stats: [],
  job_failure_reasons: [],
  failure_reason_scope: ""
});
const failureReasonJobFilter = ref("");
const autoRetryConfigRows = ref([]);
const autoRetrySummary = ref({
  enabled: false,
  maxRetries: 0,
  backoffSeconds: 0,
  jobs: []
});
const autoRetryChangeLogs = ref([]);
const autoRetryForm = reactive({
  enabled: false,
  max_retries: 2,
  backoff_seconds: 2,
  jobs: []
});

const definitionFilters = reactive({
  status: "",
  module: ""
});
const definitionPage = ref(1);
const definitionPageSize = ref(20);
const definitionTotal = ref(0);
const definitions = ref([]);
const definitionStatusMap = ref({});
const definitionStatusSavingMap = ref({});
const definitionDeletingMap = ref({});
const supportedJobOptions = ref([]);

const definitionFormVisible = ref(false);
const definitionFormMode = ref("create");
const definitionForm = reactive({
  id: "",
  job_name: "",
  display_name: "",
  module: "SYSTEM",
  cron_expr: "0 0 9 * * *",
  status: "ACTIVE"
});
const definitionSchedule = reactive({
  mode: "DAILY",
  hour: 9,
  minute: 0,
  week_day: "MON",
  interval_minutes: 15,
  custom_expr: "0 0 9 * * *"
});

const runFilters = reactive({
  job_name: "",
  status: ""
});
const runPage = ref(1);
const runPageSize = ref(20);
const runTotal = ref(0);
const runs = ref([]);
const runDetailVisible = ref(false);
const currentRun = ref(null);
const newsSyncDetailsLoading = ref(false);
const retryNewsSyncItemLoading = ref(false);
const newsSyncDetailPage = ref(1);
const newsSyncDetailPageSize = ref(20);
const newsSyncDetailTotal = ref(0);
const newsSyncDetails = ref([]);
const newsSyncDetailFilters = reactive({
  sync_type: "",
  source: "",
  symbol: "",
  status: ""
});
const marketBackfillFilters = reactive({
  status: "",
  run_type: "",
  asset_type: "",
  source_key: ""
});
const marketBackfillDetailFilters = reactive({
  stage: "",
  asset_type: "",
  status: ""
});

const retrySimMap = ref({});
const retrySummaryMap = ref({});
const retryErrorMap = ref({});
const activeTab = ref("overview");
const marketBackfillLoaded = ref(false);
const marketBackfillCreating = ref(false);
const marketBackfillRunsLoading = ref(false);
const marketBackfillRuns = ref([]);
const marketBackfillRunPage = ref(1);
const marketBackfillRunPageSize = ref(10);
const marketBackfillRunTotal = ref(0);
const marketBackfillSnapshotsLoading = ref(false);
const marketBackfillSnapshots = ref([]);
const marketBackfillDetailVisible = ref(false);
const currentMarketBackfillRun = ref(null);
const marketBackfillRunDetailLoading = ref(false);
const marketBackfillDetailsLoading = ref(false);
const marketBackfillDetailPage = ref(1);
const marketBackfillDetailPageSize = ref(20);
const marketBackfillDetailTotal = ref(0);
const marketBackfillDetails = ref([]);
const marketBackfillRetryingMap = ref({});

const MARKET_BACKFILL_DEFAULT_STAGES = [
  "UNIVERSE",
  "MASTER",
  "QUOTES",
  "DAILY_BASIC",
  "MONEYFLOW",
  "TRUTH",
  "COVERAGE_SUMMARY"
];
const MARKET_BACKFILL_REBUILD_STAGES = ["TRUTH", "COVERAGE_SUMMARY"];

const triggerForm = reactive({
  job_name: "",
  trigger_source: "MANUAL",
  simulate_status: "",
  result_summary: "",
  error_message: "",
  news_sources_text: "",
  symbols_text: "",
  sync_types: [],
  batch_size: undefined
});
const marketBackfillForm = reactive({
  run_type: "FULL",
  asset_scope: ["STOCK", "INDEX", "ETF", "LOF", "CBOND"],
  source_key: "TUSHARE",
  trade_date_from: "",
  trade_date_to: "",
  batch_size: 200,
  stages: MARKET_BACKFILL_DEFAULT_STAGES.slice(),
  force_refresh_universe: false,
  rebuild_truth_after_sync: true
});

const moduleOptions = [
  { label: "股票", value: "STOCK" },
  { label: "期货", value: "FUTURES" },
  { label: "资讯", value: "NEWS" },
  { label: "系统", value: "SYSTEM" }
];
const definitionStatusOptions = [
  { label: "启用", value: "ACTIVE" },
  { label: "停用", value: "DISABLED" }
];
const runStatusOptions = [
  { label: "运行中", value: "RUNNING" },
  { label: "成功", value: "SUCCESS" },
  { label: "失败", value: "FAILED" }
];
const marketBackfillRunTypeOptions = [
  { label: "全量回填", value: "FULL" },
  { label: "增量回填", value: "INCREMENTAL" },
  { label: "仅重建", value: "REBUILD_ONLY" }
];
const marketBackfillAssetOptions = [
  { label: "股票", value: "STOCK" },
  { label: "指数", value: "INDEX" },
  { label: "ETF", value: "ETF" },
  { label: "LOF", value: "LOF" },
  { label: "可转债", value: "CBOND" }
];
const marketBackfillStageOptions = [
  { label: "Universe", value: "UNIVERSE" },
  { label: "主数据", value: "MASTER" },
  { label: "行情", value: "QUOTES" },
  { label: "Daily Basic", value: "DAILY_BASIC" },
  { label: "Moneyflow", value: "MONEYFLOW" },
  { label: "Truth", value: "TRUTH" },
  { label: "覆盖率汇总", value: "COVERAGE_SUMMARY" }
];
const marketBackfillStatusOptions = [
  { label: "待执行", value: "PENDING" },
  { label: "进行中", value: "RUNNING" },
  { label: "部分完成", value: "PARTIAL_SUCCESS" },
  { label: "已完成", value: "SUCCESS" },
  { label: "失败", value: "FAILED" },
  { label: "已取消", value: "CANCELLED" }
];
const marketBackfillDetailStatusOptions = [
  { label: "待执行", value: "PENDING" },
  { label: "进行中", value: "RUNNING" },
  { label: "已完成", value: "SUCCESS" },
  { label: "失败", value: "FAILED" },
  { label: "已跳过", value: "SKIPPED" }
];
const simulateStatusOptions = [
  { label: "模拟成功", value: "SUCCESS" },
  { label: "模拟失败", value: "FAILED" }
];
const quickJobOptions = [
  { label: "量化流水线", value: "daily_stock_quant_pipeline" },
  { label: "每日股票推荐", value: "daily_stock_recommendation" },
  { label: "每日期货策略", value: "daily_futures_strategy" },
  { label: "期货策略生成(别名)", value: "futures_strategy_generate" },
  { label: "期货策略评估", value: "futures_strategy_evaluate" },
  { label: "Tushare资讯增量", value: "tushare_news_incremental" }
];
const newsSyncTypeOptions = [
  { label: "新闻快讯", value: "NEWS_BRIEF" },
  { label: "新闻通讯", value: "NEWS_MAJOR" },
  { label: "券商研报", value: "RESEARCH_REPORT" },
  { label: "盈利预测", value: "REPORT_RC" },
  { label: "上市公司公告", value: "ANNOUNCEMENT" }
];
const newsSyncDetailStatusOptions = [
  { label: "成功", value: "SUCCESS" },
  { label: "失败", value: "FAILED" }
];
const scheduleModeOptions = [
  { label: "每日定时", value: "DAILY" },
  { label: "每周定时", value: "WEEKLY" },
  { label: "按分钟间隔", value: "INTERVAL_MINUTES" },
  { label: "自定义表达式", value: "CUSTOM" }
];
const scheduleWeekDayOptions = [
  { label: "周一", value: "MON" },
  { label: "周二", value: "TUE" },
  { label: "周三", value: "WED" },
  { label: "周四", value: "THU" },
  { label: "周五", value: "FRI" },
  { label: "周六", value: "SAT" },
  { label: "周日", value: "SUN" }
];
const scheduleWeekDayValueSet = new Set(scheduleWeekDayOptions.map((item) => item.value));

const jobLabelMap = {
  daily_stock_quant_pipeline: "每日股票量化流水线",
  daily_stock_recommendation: "每日股票推荐",
  daily_futures_strategy: "每日期货策略",
  futures_strategy_generate: "期货策略生成",
  futures_strategy_evaluate: "期货策略评估",
  tushare_news_incremental: "Tushare资讯增量同步",
  doc_fast_news_incremental: "DocFast资讯增量同步",
  vip_membership_lifecycle: "VIP会员生命周期任务"
};

const configKeyLabelMap = {
  "scheduler.auto_retry.enabled": "自动重试开关",
  "scheduler.auto_retry.max_retries": "最大重试次数",
  "scheduler.auto_retry.backoff_seconds": "重试退避秒数",
  "scheduler.auto_retry.jobs": "允许重试任务清单"
};

const failureReasonJobOptions = computed(() => {
  const names = new Set();
  (metrics.value.job_retry_stats || []).forEach((item) => {
    if ((item?.job_name || "").trim()) {
      names.add(item.job_name.trim());
    }
  });
  (metrics.value.job_failure_reasons || []).forEach((item) => {
    if ((item?.job_name || "").trim()) {
      names.add(item.job_name.trim());
    }
  });
  return Array.from(names).sort((a, b) => a.localeCompare(b, "zh-Hans-CN"));
});

const autoRetryJobOptions = computed(() => {
  const quickLabelMap = {};
  quickJobOptions.forEach((item) => {
    const value = String(item?.value || "").trim();
    if (!value) {
      return;
    }
    quickLabelMap[value] = item.label || value;
  });
  const allNames = new Set(Object.keys(quickLabelMap));
  definitions.value.forEach((item) => {
    const name = String(item?.job_name || "").trim();
    if (name) {
      allNames.add(name);
    }
  });
  (metrics.value.job_retry_stats || []).forEach((item) => {
    const name = String(item?.job_name || "").trim();
    if (name) {
      allNames.add(name);
    }
  });
  (autoRetrySummary.value.jobs || []).forEach((jobName) => {
    const name = String(jobName || "").trim();
    if (name) {
      allNames.add(name);
    }
  });
  return Array.from(allNames)
    .sort((a, b) => a.localeCompare(b, "zh-Hans-CN"))
    .map((jobName) => ({
      value: jobName,
      label: quickLabelMap[jobName] ? `${quickLabelMap[jobName]} (${jobName})` : jobName
    }));
});

const triggerJobOptions = computed(() => {
  const sourceMap = {};
  quickJobOptions.forEach((item) => {
    const value = String(item?.value || "").trim();
    if (!value) return;
    sourceMap[value] = item.label || formatJobName(value);
  });
  definitions.value.forEach((item) => {
    const value = String(item?.job_name || "").trim();
    if (!value) return;
    if (!sourceMap[value]) {
      sourceMap[value] = item.display_name || formatJobName(value);
    }
  });
  (metrics.value.job_retry_stats || []).forEach((item) => {
    const value = String(item?.job_name || "").trim();
    if (!value) return;
    if (!sourceMap[value]) {
      sourceMap[value] = formatJobName(value);
    }
  });
  return Object.keys(sourceMap)
    .sort((a, b) => a.localeCompare(b, "zh-Hans-CN"))
    .map((value) => ({
      value,
      label: `${sourceMap[value]} (${value})`
    }));
});

const filteredJobFailureReasons = computed(() => {
  const selectedJob = (failureReasonJobFilter.value || "").trim();
  const reasons = metrics.value.job_failure_reasons || [];
  if (!selectedJob) {
    return reasons;
  }
  return reasons.filter((item) => (item?.job_name || "").trim() === selectedJob);
});

const definitionJobOptions = computed(() => {
  return buildSchedulerDefinitionOptions(supportedJobOptions.value, definitions.value);
});

const definitionCreateJobOptions = computed(() =>
  buildSchedulerDefinitionCreateOptions(supportedJobOptions.value, definitions.value)
);

const definitionCronPreview = computed(() => buildDefinitionCronExpression());
const failedRunCount = computed(() =>
  (runs.value || []).filter((item) => String(item?.status || "").trim().toUpperCase() === "FAILED").length
);
const overviewCards = computed(() =>
  buildSystemJobsOverviewCards({
    metrics: metrics.value,
    autoRetrySummary: autoRetrySummary.value,
    definitionTotal: definitionTotal.value
  })
);
const guideCards = computed(() => buildSystemJobsGuideCards({ canEditSystemJobs }));
const actionCards = computed(() =>
  buildSystemJobsActionCards({
    canEditSystemJobs,
    failedRunCount: failedRunCount.value
  })
);
const systemJobTabs = computed(() => buildSystemJobsTabOptions({ canEditSystemJobs }));
const marketBackfillOverviewCards = computed(() =>
  buildMarketBackfillOverviewCards({
    runs: marketBackfillRuns.value,
    snapshots: marketBackfillSnapshots.value
  })
);
const marketBackfillGuideCards = computed(() =>
  buildMarketBackfillGuideCards({ canEditSystemJobs })
);

function clampInteger(value, minValue, maxValue, fallback) {
  const parsed = Number.parseInt(String(value), 10);
  if (!Number.isFinite(parsed)) {
    return fallback;
  }
  return Math.max(minValue, Math.min(maxValue, parsed));
}

function parseCronTokens(raw) {
  const tokens = String(raw || "")
    .trim()
    .split(/\s+/)
    .filter(Boolean);
  if (tokens.length !== 6) {
    return null;
  }
  return tokens;
}

function normalizeWeekDayValue(raw) {
  const key = String(raw || "").trim().toUpperCase();
  if (scheduleWeekDayValueSet.has(key)) {
    return key;
  }
  const numberMap = {
    "0": "SUN",
    "7": "SUN",
    "1": "MON",
    "2": "TUE",
    "3": "WED",
    "4": "THU",
    "5": "FRI",
    "6": "SAT"
  };
  return numberMap[key] || "";
}

function parseMinuteStepToken(raw) {
  const text = String(raw || "").trim();
  const matched = text.match(/^(?:\*|0)\/(\d+)$/i);
  if (!matched) {
    return NaN;
  }
  const step = Number.parseInt(matched[1], 10);
  if (!Number.isFinite(step) || step < 1 || step > 59) {
    return NaN;
  }
  return step;
}

function syncDefinitionScheduleFromCron(rawCronExpr) {
  const cronExpr = String(rawCronExpr || "").trim() || "0 0 9 * * *";
  const tokens = parseCronTokens(cronExpr);
  if (!tokens) {
    Object.assign(definitionSchedule, {
      mode: "CUSTOM",
      custom_expr: cronExpr
    });
    return;
  }
  const [seconds, minuteToken, hourToken, dayToken, monthToken, weekToken] = tokens;
  const minute = clampInteger(minuteToken, 0, 59, 0);
  const hour = clampInteger(hourToken, 0, 23, 9);

  if (
    seconds === "0" &&
    dayToken === "*" &&
    monthToken === "*" &&
    /^\d+$/.test(minuteToken) &&
    /^\d+$/.test(hourToken)
  ) {
    if (weekToken === "*") {
      Object.assign(definitionSchedule, {
        mode: "DAILY",
        hour,
        minute,
        custom_expr: cronExpr
      });
      return;
    }
    const weekDay = normalizeWeekDayValue(weekToken);
    if (weekDay) {
      Object.assign(definitionSchedule, {
        mode: "WEEKLY",
        hour,
        minute,
        week_day: weekDay,
        custom_expr: cronExpr
      });
      return;
    }
  }

  if (seconds === "0" && hourToken === "*" && dayToken === "*" && monthToken === "*" && weekToken === "*") {
    const intervalMinutes = parseMinuteStepToken(minuteToken);
    if (Number.isFinite(intervalMinutes)) {
      Object.assign(definitionSchedule, {
        mode: "INTERVAL_MINUTES",
        interval_minutes: intervalMinutes,
        custom_expr: cronExpr
      });
      return;
    }
  }

  Object.assign(definitionSchedule, {
    mode: "CUSTOM",
    custom_expr: cronExpr
  });
}

function buildDefinitionCronExpression() {
  const mode = String(definitionSchedule.mode || "DAILY").trim().toUpperCase();
  if (mode === "CUSTOM") {
    return String(definitionSchedule.custom_expr || "").trim();
  }
  if (mode === "INTERVAL_MINUTES") {
    const interval = clampInteger(definitionSchedule.interval_minutes, 1, 59, 15);
    definitionSchedule.interval_minutes = interval;
    return `0 */${interval} * * * *`;
  }
  const minute = clampInteger(definitionSchedule.minute, 0, 59, 0);
  const hour = clampInteger(definitionSchedule.hour, 0, 23, 9);
  definitionSchedule.minute = minute;
  definitionSchedule.hour = hour;
  if (mode === "WEEKLY") {
    const weekDay = normalizeWeekDayValue(definitionSchedule.week_day) || "MON";
    definitionSchedule.week_day = weekDay;
    return `0 ${minute} ${hour} * * ${weekDay}`;
  }
  return `0 ${minute} ${hour} * * *`;
}

function resetDefinitionForm() {
  Object.assign(definitionForm, {
    id: "",
    job_name: "",
    display_name: "",
    module: "SYSTEM",
    cron_expr: "0 0 9 * * *",
    status: "ACTIVE"
  });
  syncDefinitionScheduleFromCron(definitionForm.cron_expr);
  definitionFormMode.value = "create";
}

function syncDefinitionDrafts() {
  const statusMap = {};
  definitions.value.forEach((item) => {
    statusMap[item.id] = item.status || "ACTIVE";
  });
  definitionStatusMap.value = statusMap;
}

function syncRunDrafts() {
  const simMap = {};
  const summaryMap = {};
  const errorMap = {};
  runs.value.forEach((item) => {
    simMap[item.id] = "";
    summaryMap[item.id] = "";
    errorMap[item.id] = "";
  });
  retrySimMap.value = simMap;
  retrySummaryMap.value = summaryMap;
  retryErrorMap.value = errorMap;
}

function cleanupPayload(raw) {
  const payload = {};
  Object.entries(raw).forEach(([key, value]) => {
    if (value === undefined || value === null || value === "") {
      return;
    }
    payload[key] = value;
  });
  return payload;
}

function parseTextList(raw) {
  return Array.from(
    new Set(
      String(raw || "")
        .split(/[\n,，;；\s]+/)
        .map((item) => item.trim())
        .filter(Boolean)
    )
  );
}

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function actionButtonType(tone) {
  if (tone === "danger") {
    return "danger";
  }
  if (tone === "success") {
    return "success";
  }
  return "primary";
}

function handleActionCard(actionKey) {
  if (actionKey === "view-failed-runs") {
    activeTab.value = "runs";
    applyFailedRunFilter();
    return;
  }
  if (actionKey === "refresh-all") {
    refreshAll();
    return;
  }
  if (actionKey === "open-create-definition") {
    activeTab.value = "config";
    openCreateDefinition();
    return;
  }
  if (actionKey === "scroll-trigger") {
    activeTab.value = "trigger";
    return;
  }
  if (actionKey === "scroll-definitions") {
    activeTab.value = "config";
  }
}

function csvEscape(value) {
  const text = String(value ?? "");
  if (/[",\n]/.test(text)) {
    return `"${text.replace(/"/g, "\"\"")}"`;
  }
  return text;
}

function triggerCSVDownload(content, fileName) {
  const blob = new Blob([`\uFEFF${content}`], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = fileName;
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  URL.revokeObjectURL(url);
}

function buildRunCSVRows(items) {
  const header = [
    "运行ID",
    "上级运行ID",
    "任务编码",
    "任务名称",
    "状态",
    "重试次数",
    "触发来源",
    "开始时间",
    "结束时间",
    "执行耗时",
    "运行摘要",
    "错误信息",
    "操作人ID"
  ];
  const rows = items.map((item) => [
    item.id || "",
    item.parent_run_id || "",
    item.job_name || "",
    formatJobName(item.job_name),
    formatRunStatus(item.status),
    item.retry_count ?? "",
    formatTriggerSource(item.trigger_source),
    item.started_at || "",
    item.finished_at || "",
    formatRunDuration(item.started_at, item.finished_at, item.status),
    item.result_summary || "",
    item.error_message || "",
    item.operator_id || ""
  ]);
  return [header, ...rows].map((row) => row.map(csvEscape).join(",")).join("\n");
}

function exportRunCurrentPageCSV() {
  const csv = buildRunCSVRows(runs.value);
  const fileName = `system_job_runs_page_${new Date().toISOString().slice(0, 10)}.csv`;
  triggerCSVDownload(csv, fileName);
  message.value = `已导出当前页运行记录 CSV，共 ${runs.value.length} 条`;
}

async function exportRunFilteredCSV() {
  exportingRuns.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const params = new URLSearchParams();
    if (runFilters.job_name.trim()) params.set("job_name", runFilters.job_name.trim());
    if (runFilters.status) params.set("status", runFilters.status);
    const baseURL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
    const query = params.toString();
    const requestURL = `${baseURL}/admin/system/job-runs/export.csv${query ? `?${query}` : ""}`;
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
    const fileName = `system_job_runs_filtered_${new Date().toISOString().slice(0, 10)}.csv`;
    const anchor = document.createElement("a");
    anchor.href = blobURL;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(blobURL);
    message.value = "已发起筛选运行记录 CSV 下载";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "导出运行记录失败");
  } finally {
    exportingRuns.value = false;
  }
}

function isTushareRun(item) {
  return String(item?.job_name || "").trim() === "tushare_news_incremental";
}

function isFuturesGenerateRun(item) {
  const jobName = String(item?.job_name || "").trim();
  return jobName === "daily_futures_strategy" || jobName === "futures_strategy_generate";
}

function isFuturesEvaluateRun(item) {
  return String(item?.job_name || "").trim() === "futures_strategy_evaluate";
}

function isFuturesRun(item) {
  return isFuturesGenerateRun(item) || isFuturesEvaluateRun(item);
}

function resetNewsSyncDetailState() {
  newsSyncDetailPage.value = 1;
  newsSyncDetailTotal.value = 0;
  newsSyncDetails.value = [];
  newsSyncDetailFilters.sync_type = "";
  newsSyncDetailFilters.source = "";
  newsSyncDetailFilters.symbol = "";
  newsSyncDetailFilters.status = "";
}

async function fetchNewsSyncDetails() {
  if (!currentRun.value?.id || !isTushareRun(currentRun.value)) {
    newsSyncDetails.value = [];
    newsSyncDetailTotal.value = 0;
    return;
  }
  newsSyncDetailsLoading.value = true;
  try {
    const data = await listNewsSyncRunDetails(currentRun.value.id, {
      sync_type: newsSyncDetailFilters.sync_type,
      source: newsSyncDetailFilters.source.trim(),
      symbol: newsSyncDetailFilters.symbol.trim(),
      status: newsSyncDetailFilters.status,
      page: newsSyncDetailPage.value,
      page_size: newsSyncDetailPageSize.value
    });
    newsSyncDetails.value = data?.items || [];
    newsSyncDetailTotal.value = data?.total || 0;
  } catch (error) {
    newsSyncDetails.value = [];
    newsSyncDetailTotal.value = 0;
    errorMessage.value = normalizeErrorMessage(error, "加载同步明细失败");
  } finally {
    newsSyncDetailsLoading.value = false;
  }
}

function applyNewsSyncDetailFilters() {
  newsSyncDetailPage.value = 1;
  fetchNewsSyncDetails();
}

function resetNewsSyncDetailFilters() {
  newsSyncDetailFilters.sync_type = "";
  newsSyncDetailFilters.source = "";
  newsSyncDetailFilters.symbol = "";
  newsSyncDetailFilters.status = "";
  newsSyncDetailPage.value = 1;
  fetchNewsSyncDetails();
}

function handleNewsSyncDetailPageChange(nextPage) {
  if (nextPage === newsSyncDetailPage.value) {
    return;
  }
  newsSyncDetailPage.value = nextPage;
  fetchNewsSyncDetails();
}

async function retryNewsSyncDetailItem(row) {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  if (!currentRun.value?.id || !isTushareRun(currentRun.value)) {
    return;
  }
  retryNewsSyncItemLoading.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const payload = cleanupPayload({
      sync_type: row.sync_type,
      source: row.source || "",
      symbol: row.symbol || ""
    });
    const result = await retryNewsSyncItem(currentRun.value.id, payload);
    message.value = `明细重试已触发：新运行ID=${result.id || "-"}，状态=${formatRunStatus(result.status)}`;
    await Promise.all([fetchRuns(), fetchNewsSyncDetails(), fetchMetrics()]);
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "重试失败明细失败");
  } finally {
    retryNewsSyncItemLoading.value = false;
  }
}

async function openRunDetail(item) {
  currentRun.value = item;
  runDetailVisible.value = true;
  resetNewsSyncDetailState();
  if (isTushareRun(item)) {
    await fetchNewsSyncDetails();
  }
}

async function copyRunField(label, value) {
  const text = String(value || "").trim();
  if (!text) {
    errorMessage.value = `${label}为空，无法复制`;
    return;
  }
  copyingRunText.value = true;
  errorMessage.value = "";
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
    } else {
      const textarea = document.createElement("textarea");
      textarea.value = text;
      textarea.style.position = "fixed";
      textarea.style.opacity = "0";
      document.body.appendChild(textarea);
      textarea.focus();
      textarea.select();
      document.execCommand("copy");
      document.body.removeChild(textarea);
    }
    message.value = `${label}已复制`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, `复制${label}失败`);
  } finally {
    copyingRunText.value = false;
  }
}

function handleRunDetailClosed() {
  currentRun.value = null;
  resetNewsSyncDetailState();
}

async function fetchMetrics() {
  metricsLoading.value = true;
  const defaultMetrics = {
    today_total: 0,
    today_success: 0,
    today_failed: 0,
    today_running: 0,
    retry_total: 0,
    retry_success: 0,
    retry_failed: 0,
    retry_hit_rate: 0,
    avg_retry_count: 0,
    auto_retry_total: 0,
    recovery_total: 0,
    recovery_success: 0,
    recovery_hit_rate: 0,
    failure_reasons: [],
    job_retry_stats: [],
    job_failure_reasons: [],
    failure_reason_scope: ""
  };
  try {
    const data = await getSchedulerJobMetrics({ job_name: metricFilter.job_name.trim() });
    metrics.value = { ...defaultMetrics, ...(data || {}) };
    if (
      failureReasonJobFilter.value &&
      !failureReasonJobOptions.value.includes(failureReasonJobFilter.value)
    ) {
      failureReasonJobFilter.value = "";
    }
  } catch {
    metrics.value = defaultMetrics;
  } finally {
    metricsLoading.value = false;
  }
}

async function fetchAutoRetryConfigs() {
  autoRetryLoading.value = true;
  try {
    const data = await listSystemConfigs({ keyword: "scheduler.auto_retry", page: 1, page_size: 50 });
    const items = data?.items || [];
    autoRetryConfigRows.value = items;
    autoRetrySummary.value = parseAutoRetrySummary(items);
    syncAutoRetryForm(autoRetrySummary.value);
  } catch {
    autoRetryConfigRows.value = [];
    autoRetrySummary.value = {
      enabled: false,
      maxRetries: 0,
      backoffSeconds: 0,
      jobs: []
    };
    syncAutoRetryForm(autoRetrySummary.value);
  } finally {
    autoRetryLoading.value = false;
  }
}

async function fetchAutoRetryChangeLogs() {
  autoRetryLogLoading.value = true;
  try {
    const data = await listOperationLogs({
      module: "SYSTEM",
      action: "UPSERT_CONFIG",
      page: 1,
      page_size: 100
    });
    const items = (data?.items || []).filter((item) => {
      const targetType = String(item?.target_type || "").trim().toUpperCase();
      const targetID = String(item?.target_id || "").trim().toLowerCase();
      return targetType === "SYSTEM_CONFIG" && targetID.startsWith("scheduler.auto_retry");
    });
    autoRetryChangeLogs.value = items.slice(0, 40);
  } catch {
    autoRetryChangeLogs.value = [];
  } finally {
    autoRetryLogLoading.value = false;
  }
}

async function fetchSupportedJobs() {
  supportedJobsLoading.value = true;
  try {
    const data = await listSupportedSchedulerJobs();
    supportedJobOptions.value = data?.items || [];
  } catch {
    supportedJobOptions.value = [];
  } finally {
    supportedJobsLoading.value = false;
  }
}

async function fetchDefinitions() {
  defsLoading.value = true;
  errorMessage.value = "";
  try {
    const data = await listSchedulerJobDefinitions({
      status: definitionFilters.status,
      module: definitionFilters.module,
      page: definitionPage.value,
      page_size: definitionPageSize.value
    });
    definitions.value = data.items || [];
    definitionTotal.value = data.total || 0;
    syncDefinitionDrafts();
  } catch (error) {
    errorMessage.value = error.message || "加载任务定义失败";
  } finally {
    defsLoading.value = false;
  }
}

async function fetchRuns() {
  runsLoading.value = true;
  errorMessage.value = "";
  try {
    const data = await listSchedulerJobRuns({
      job_name: runFilters.job_name.trim(),
      status: runFilters.status,
      page: runPage.value,
      page_size: runPageSize.value
    });
    runs.value = data.items || [];
    runTotal.value = data.total || 0;
    syncRunDrafts();
  } catch (error) {
    errorMessage.value = error.message || "加载运行记录失败";
  } finally {
    runsLoading.value = false;
  }
}

function normalizeBackfillTab(rawTab) {
  const nextTab = String(rawTab || "").trim();
  return systemJobTabs.value.some((item) => item.key === nextTab) ? nextTab : "overview";
}

function resetMarketBackfillForm() {
  Object.assign(marketBackfillForm, {
    run_type: "FULL",
    asset_scope: ["STOCK", "INDEX", "ETF", "LOF", "CBOND"],
    source_key: "TUSHARE",
    trade_date_from: "",
    trade_date_to: "",
    batch_size: 200,
    stages: MARKET_BACKFILL_DEFAULT_STAGES.slice(),
    force_refresh_universe: false,
    rebuild_truth_after_sync: true
  });
}

function buildMarketBackfillPayload() {
  const batchSize = Number.parseInt(String(marketBackfillForm.batch_size || ""), 10);
  return cleanupPayload({
    run_type: String(marketBackfillForm.run_type || "").trim().toUpperCase(),
    asset_scope: Array.from(
      new Set(
        (marketBackfillForm.asset_scope || [])
          .map((item) => String(item || "").trim().toUpperCase())
          .filter(Boolean)
      )
    ),
    source_key: String(marketBackfillForm.source_key || "").trim().toUpperCase(),
    trade_date_from: String(marketBackfillForm.trade_date_from || "").trim(),
    trade_date_to: String(marketBackfillForm.trade_date_to || "").trim(),
    batch_size: Number.isFinite(batchSize) && batchSize > 0 ? batchSize : undefined,
    stages: Array.from(
      new Set(
        (marketBackfillForm.stages || [])
          .map((item) => String(item || "").trim().toUpperCase())
          .filter(Boolean)
      )
    ),
    force_refresh_universe: !!marketBackfillForm.force_refresh_universe,
    rebuild_truth_after_sync: !!marketBackfillForm.rebuild_truth_after_sync
  });
}

async function fetchMarketBackfillRuns(options = {}) {
  const { preserveFeedback = false } = options;
  marketBackfillRunsLoading.value = true;
  if (!preserveFeedback) {
    errorMessage.value = "";
  }
  try {
    const data = await listMarketDataBackfillRuns({
      status: marketBackfillFilters.status,
      run_type: marketBackfillFilters.run_type,
      asset_type: marketBackfillFilters.asset_type,
      source_key: String(marketBackfillFilters.source_key || "").trim().toUpperCase(),
      page: marketBackfillRunPage.value,
      page_size: marketBackfillRunPageSize.value
    });
    marketBackfillRuns.value = data?.items || [];
    marketBackfillRunTotal.value = data?.total || 0;
    marketBackfillLoaded.value = true;
  } catch (error) {
    marketBackfillRuns.value = [];
    marketBackfillRunTotal.value = 0;
    if (!preserveFeedback) {
      errorMessage.value = normalizeErrorMessage(error, "加载市场回填总单失败");
    }
  } finally {
    marketBackfillRunsLoading.value = false;
  }
}

async function fetchMarketBackfillSnapshots(options = {}) {
  const { preserveFeedback = false } = options;
  marketBackfillSnapshotsLoading.value = true;
  if (!preserveFeedback) {
    errorMessage.value = "";
  }
  try {
    const data = await listMarketUniverseSnapshots({ page: 1, page_size: 6 });
    marketBackfillSnapshots.value = data?.items || [];
    marketBackfillLoaded.value = true;
  } catch (error) {
    marketBackfillSnapshots.value = [];
    if (!preserveFeedback) {
      errorMessage.value = normalizeErrorMessage(error, "加载 Universe 快照失败");
    }
  } finally {
    marketBackfillSnapshotsLoading.value = false;
  }
}

async function fetchMarketBackfillWorkspace(options = {}) {
  const { preserveFeedback = false } = options;
  if (!preserveFeedback) {
    errorMessage.value = "";
  }
  await Promise.all([
    fetchMarketBackfillRuns({ preserveFeedback: true }),
    fetchMarketBackfillSnapshots({ preserveFeedback: true })
  ]).catch((error) => {
    if (!preserveFeedback) {
      errorMessage.value = normalizeErrorMessage(error, "加载市场数据工作台失败");
    }
  });
}

async function fetchCurrentMarketBackfillRun(runID, options = {}) {
  const { preserveFeedback = false } = options;
  if (!runID) {
    currentMarketBackfillRun.value = null;
    return;
  }
  marketBackfillRunDetailLoading.value = true;
  if (!preserveFeedback) {
    errorMessage.value = "";
  }
  try {
    const data = await getMarketDataBackfillRun(runID);
    currentMarketBackfillRun.value = data || null;
  } catch (error) {
    currentMarketBackfillRun.value = null;
    if (!preserveFeedback) {
      errorMessage.value = normalizeErrorMessage(error, "加载回填总单详情失败");
    }
  } finally {
    marketBackfillRunDetailLoading.value = false;
  }
}

async function fetchMarketBackfillRunDetails(options = {}) {
  const { preserveFeedback = false } = options;
  const runID = String(currentMarketBackfillRun.value?.id || "").trim();
  if (!runID) {
    marketBackfillDetails.value = [];
    marketBackfillDetailTotal.value = 0;
    return;
  }
  marketBackfillDetailsLoading.value = true;
  if (!preserveFeedback) {
    errorMessage.value = "";
  }
  try {
    const data = await listMarketDataBackfillRunDetails(runID, {
      stage: marketBackfillDetailFilters.stage,
      asset_type: marketBackfillDetailFilters.asset_type,
      status: marketBackfillDetailFilters.status,
      page: marketBackfillDetailPage.value,
      page_size: marketBackfillDetailPageSize.value
    });
    marketBackfillDetails.value = data?.items || [];
    marketBackfillDetailTotal.value = data?.total || 0;
  } catch (error) {
    marketBackfillDetails.value = [];
    marketBackfillDetailTotal.value = 0;
    if (!preserveFeedback) {
      errorMessage.value = normalizeErrorMessage(error, "加载回填批次明细失败");
    }
  } finally {
    marketBackfillDetailsLoading.value = false;
  }
}

async function handleCreateMarketBackfillRun() {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  const payload = buildMarketBackfillPayload();
  if (!payload.run_type || !Array.isArray(payload.asset_scope) || payload.asset_scope.length === 0) {
    errorMessage.value = "运行类型和资产范围不能为空";
    return;
  }
  marketBackfillCreating.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await createMarketDataBackfillRun(payload);
    message.value = `已创建回填任务：${result.run_id || "-"}，状态=${formatMarketBackfillStatusLabel(result.status)}`;
    marketBackfillRunPage.value = 1;
    await Promise.all([
      fetchMarketBackfillWorkspace({ preserveFeedback: true }),
      fetchMetrics(),
      fetchRuns()
    ]);
    if (result?.run_id) {
      await handleOpenMarketBackfillRun({ id: result.run_id });
    }
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "创建市场回填任务失败");
  } finally {
    marketBackfillCreating.value = false;
  }
}

async function handleOpenMarketBackfillRun(item) {
  const runID = String(item?.id || item?.run_id || "").trim();
  if (!runID) {
    errorMessage.value = "回填总单 ID 为空，无法查看详情";
    return;
  }
  currentMarketBackfillRun.value = { id: runID };
  marketBackfillDetailVisible.value = true;
  marketBackfillDetailPage.value = 1;
  marketBackfillDetailFilters.stage = "";
  marketBackfillDetailFilters.asset_type = "";
  marketBackfillDetailFilters.status = "";
  await Promise.all([
    fetchCurrentMarketBackfillRun(runID, { preserveFeedback: true }),
    fetchMarketBackfillRunDetails({ preserveFeedback: true })
  ]);
}

function handleMarketBackfillDetailClosed() {
  currentMarketBackfillRun.value = null;
  marketBackfillDetails.value = [];
  marketBackfillDetailTotal.value = 0;
}

async function handleRetryMarketBackfillRun(item) {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  const runID = String(item?.id || "").trim();
  if (!runID) {
    errorMessage.value = "回填总单 ID 为空，无法重试";
    return;
  }
  try {
    await ElMessageBox.confirm(
      "将仅重试失败或未完成的批次，不会重新从头全量执行，是否继续？",
      "确认重试市场回填",
      {
        type: "warning",
        confirmButtonText: "确认重试",
        cancelButtonText: "取消"
      }
    );
  } catch (error) {
    if (error === "cancel" || error === "close") {
      message.value = "已取消回填重试";
      return;
    }
    errorMessage.value = normalizeErrorMessage(error, "回填重试确认失败");
    return;
  }

  marketBackfillRetryingMap.value[runID] = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await retryMarketDataBackfillRun(runID, { retry_mode: "FAILED_ONLY" });
    message.value = `已发起回填重试：${result.run_id || "-"}，状态=${formatMarketBackfillStatusLabel(result.status)}`;
    await Promise.all([
      fetchMarketBackfillWorkspace({ preserveFeedback: true }),
      fetchMetrics(),
      fetchRuns()
    ]);
    if (String(currentMarketBackfillRun.value?.id || "").trim() === runID) {
      await Promise.all([
        fetchCurrentMarketBackfillRun(runID, { preserveFeedback: true }),
        fetchMarketBackfillRunDetails({ preserveFeedback: true })
      ]);
    }
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "重试市场回填任务失败");
  } finally {
    marketBackfillRetryingMap.value[runID] = false;
  }
}

function applyMarketBackfillFilters() {
  marketBackfillRunPage.value = 1;
  fetchMarketBackfillRuns();
}

function resetMarketBackfillFilters() {
  marketBackfillFilters.status = "";
  marketBackfillFilters.run_type = "";
  marketBackfillFilters.asset_type = "";
  marketBackfillFilters.source_key = "";
  marketBackfillRunPage.value = 1;
  fetchMarketBackfillRuns();
}

function handleMarketBackfillRunPageChange(nextPage) {
  if (nextPage === marketBackfillRunPage.value) {
    return;
  }
  marketBackfillRunPage.value = nextPage;
  fetchMarketBackfillRuns();
}

function applyMarketBackfillDetailFilters() {
  marketBackfillDetailPage.value = 1;
  fetchMarketBackfillRunDetails();
}

function resetMarketBackfillDetailFilters() {
  marketBackfillDetailFilters.stage = "";
  marketBackfillDetailFilters.asset_type = "";
  marketBackfillDetailFilters.status = "";
  marketBackfillDetailPage.value = 1;
  fetchMarketBackfillRunDetails();
}

function handleMarketBackfillDetailPageChange(nextPage) {
  if (nextPage === marketBackfillDetailPage.value) {
    return;
  }
  marketBackfillDetailPage.value = nextPage;
  fetchMarketBackfillRunDetails();
}

function canRetryMarketBackfillRun(item) {
  const status = String(item?.status || "").trim().toUpperCase();
  return status === "FAILED" || status === "PARTIAL_SUCCESS";
}

function formatMarketBackfillSummary(summary) {
  if (!summary) {
    return "暂无回填摘要";
  }
  if (typeof summary === "string") {
    return summary.trim() || "暂无回填摘要";
  }
  try {
    return JSON.stringify(summary, null, 2);
  } catch {
    return String(summary);
  }
}

function formatMarketBackfillBatchSample(row) {
  const items = Array.isArray(row?.symbol_sample) ? row.symbol_sample.filter(Boolean) : [];
  return items.length ? items.join(", ") : "-";
}

function formatMarketBackfillTimeRange(row) {
  const from = String(row?.trade_date_from || "").trim();
  const to = String(row?.trade_date_to || "").trim();
  if (from && to) {
    return `${from} -> ${to}`;
  }
  return from || to || "-";
}

function formatMarketStageProgressSummary(item) {
  const totalBatches = Number(item?.total_batches) || 0;
  const completedBatches = Number(item?.completed_batches) || 0;
  const failedBatches = Number(item?.failed_batches) || 0;
  const skippedBatches = Number(item?.skipped_batches) || 0;
  return `总 ${totalBatches} · 完成 ${completedBatches} · 失败 ${failedBatches} · 跳过 ${skippedBatches}`;
}

async function refreshAll() {
  errorMessage.value = "";
  message.value = "";
  await Promise.all([
    fetchMetrics(),
    fetchAutoRetryConfigs(),
    fetchAutoRetryChangeLogs(),
    fetchSupportedJobs(),
    fetchDefinitions(),
    fetchRuns(),
    fetchMarketBackfillWorkspace({ preserveFeedback: true })
  ]);
}

function openCreateDefinition() {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  if (!supportedJobOptions.value.length && !supportedJobsLoading.value) {
    fetchSupportedJobs();
  }
  resetDefinitionForm();
  definitionFormVisible.value = true;
}

function openEditDefinition(item) {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  if (!supportedJobOptions.value.length && !supportedJobsLoading.value) {
    fetchSupportedJobs();
  }
  Object.assign(definitionForm, {
    id: item.id,
    job_name: item.job_name || "",
    display_name: item.display_name || "",
    module: item.module || "SYSTEM",
    cron_expr: item.cron_expr || "",
    status: item.status || "ACTIVE"
  });
  syncDefinitionScheduleFromCron(definitionForm.cron_expr);
  definitionFormMode.value = "edit";
  definitionFormVisible.value = true;
}

async function submitDefinition() {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  const cronExpr = buildDefinitionCronExpression();
  const payload = {
    job_name: definitionForm.job_name.trim(),
    display_name: definitionForm.display_name.trim(),
    module: definitionForm.module,
    cron_expr: cronExpr,
    status: definitionForm.status
  };
  if (!payload.job_name || !payload.display_name || !payload.cron_expr) {
    errorMessage.value = "任务编码/任务名称/调度表达式不能为空";
    return;
  }
  const jobNameError = validateSchedulerDefinitionJobName(
    payload.job_name,
    definitions.value,
    definitionFormMode.value === "edit" ? definitionForm.id : ""
  );
  if (jobNameError) {
    errorMessage.value = jobNameError;
    return;
  }
  definitionForm.cron_expr = payload.cron_expr;
  submittingDefinition.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    if (definitionFormMode.value === "create") {
      const result = await createSchedulerJobDefinition(payload);
      message.value = `任务定义已创建：${result.id || ""}`;
    } else {
      await updateSchedulerJobDefinition(definitionForm.id, payload);
      message.value = `任务定义 ${definitionForm.id} 已更新`;
    }
    definitionFormVisible.value = false;
    resetDefinitionForm();
    await fetchDefinitions();
  } catch (error) {
    errorMessage.value = error.message || "提交任务定义失败";
  } finally {
    submittingDefinition.value = false;
  }
}

async function updateDefinitionStatus(item, nextStatus) {
  if (!ensureCanEditSystemJobs()) {
    definitionStatusMap.value[item.id] = item.status || "ACTIVE";
    return;
  }
  const status = String(nextStatus || definitionStatusMap.value[item.id] || "").trim();
  if (!status || status === item.status) {
    return;
  }
  definitionStatusSavingMap.value[item.id] = true;
  errorMessage.value = "";
  message.value = "";
  try {
    await updateSchedulerJobDefinitionStatus(item.id, status);
    message.value = `任务定义 ${item.id} 状态已更新为 ${formatDefinitionStatus(status)}(${status})`;
    await fetchDefinitions();
  } catch (error) {
    definitionStatusMap.value[item.id] = item.status || "ACTIVE";
    errorMessage.value = error.message || "更新任务定义状态失败";
  } finally {
    definitionStatusSavingMap.value[item.id] = false;
  }
}

function onDefinitionStatusSwitch(item, active) {
  if (!ensureCanEditSystemJobs()) {
    definitionStatusMap.value[item.id] = item.status || "ACTIVE";
    return;
  }
  const nextStatus = active ? "ACTIVE" : "DISABLED";
  definitionStatusMap.value[item.id] = nextStatus;
  updateDefinitionStatus(item, nextStatus);
}

async function removeDefinition(item) {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  const id = String(item?.id || "").trim();
  if (!id) {
    errorMessage.value = "任务定义ID为空，无法删除";
    return;
  }
  try {
    await ElMessageBox.confirm(
      `删除后不可恢复，确认删除任务定义“${item.display_name || id}”吗？`,
      "确认删除任务定义",
      {
        type: "warning",
        confirmButtonText: "确认删除",
        cancelButtonText: "取消"
      }
    );
  } catch (error) {
    if (error === "cancel" || error === "close") {
      message.value = "已取消删除";
      return;
    }
    errorMessage.value = normalizeErrorMessage(error, "删除确认失败");
    return;
  }

  definitionDeletingMap.value[id] = true;
  errorMessage.value = "";
  message.value = "";
  try {
    await deleteSchedulerJobDefinition(id);
    message.value = `任务定义 ${id} 已删除`;
    await Promise.all([fetchDefinitions(), fetchMetrics()]);
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "删除任务定义失败");
  } finally {
    definitionDeletingMap.value[id] = false;
  }
}

async function submitTrigger() {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  const newsSources = parseTextList(triggerForm.news_sources_text);
  const symbols = parseTextList(triggerForm.symbols_text).map((item) => item.toUpperCase());
  const syncTypes = Array.from(
    new Set(
      (triggerForm.sync_types || [])
        .map((item) => String(item || "").trim().toUpperCase())
        .filter(Boolean)
    )
  );
  const batchSize = Number.parseInt(String(triggerForm.batch_size || ""), 10);
  const payload = cleanupPayload({
    job_name: triggerForm.job_name.trim(),
    trigger_source: triggerForm.trigger_source,
    simulate_status: triggerForm.simulate_status,
    result_summary: triggerForm.result_summary.trim(),
    error_message: triggerForm.error_message.trim(),
    news_sources: newsSources,
    symbols,
    sync_types: syncTypes,
    batch_size: Number.isFinite(batchSize) && batchSize > 0 ? batchSize : undefined
  });
  if (!payload.job_name || !payload.trigger_source) {
    errorMessage.value = "任务编码和触发来源不能为空";
    return;
  }
  triggeringJob.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await triggerSchedulerJob(payload);
    const retryAttempts = Number(result?.retry_attempts || 0);
    message.value = `任务触发完成：运行ID=${result.id || "-"}，状态=${formatRunStatus(result.status)}，自动重试次数=${retryAttempts}`;
    await Promise.all([fetchMetrics(), fetchRuns()]);
  } catch (error) {
    errorMessage.value = error.message || "触发任务失败";
  } finally {
    triggeringJob.value = false;
  }
}

async function retryRun(runID) {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  const payload = cleanupPayload({
    simulate_status: retrySimMap.value[runID] || "",
    result_summary: (retrySummaryMap.value[runID] || "").trim(),
    error_message: (retryErrorMap.value[runID] || "").trim()
  });
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await retrySchedulerJobRun(runID, payload);
    const retryAttempts = Number(result?.retry_attempts || 0);
    message.value = `重跑已触发：新运行ID=${result.id || "-"}，状态=${formatRunStatus(result.status)}，自动重试次数=${retryAttempts}`;
    await Promise.all([fetchMetrics(), fetchRuns()]);
  } catch (error) {
    errorMessage.value = error.message || "重跑任务失败";
  }
}

function applyDefinitionFilters() {
  definitionPage.value = 1;
  fetchDefinitions();
}

function resetDefinitionFilters() {
  definitionFilters.status = "";
  definitionFilters.module = "";
  definitionPage.value = 1;
  fetchDefinitions();
}

function handleDefinitionPageChange(nextPage) {
  if (nextPage === definitionPage.value) {
    return;
  }
  definitionPage.value = nextPage;
  fetchDefinitions();
}

function applyRunFilters() {
  runPage.value = 1;
  fetchRuns();
}

function resetRunFilters() {
  runFilters.job_name = "";
  runFilters.status = "";
  runPage.value = 1;
  fetchRuns();
}

function handleRunPageChange(nextPage) {
  if (nextPage === runPage.value) {
    return;
  }
  runPage.value = nextPage;
  fetchRuns();
}

function applyFailedRunFilter() {
  runFilters.status = "FAILED";
  runPage.value = 1;
  fetchRuns();
}

async function retryFailedRunsCurrentPage() {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  const failedRuns = (runs.value || []).filter((item) => String(item?.status || "").toUpperCase() === "FAILED");
  if (failedRuns.length === 0) {
    errorMessage.value = "";
    message.value = "当前页没有失败任务";
    return;
  }
  const concurrency = Math.max(1, Math.min(10, Number.parseInt(String(batchRetryConcurrency.value || 1), 10) || 1));
  try {
    await ElMessageBox.confirm(
      `将按并发 ${concurrency} 触发当前页 ${failedRuns.length} 条失败任务重跑，是否继续？`,
      "确认批量重跑",
      {
        type: "warning",
        confirmButtonText: "确认重跑",
        cancelButtonText: "取消"
      }
    );
  } catch (error) {
    if (error === "cancel" || error === "close") {
      message.value = "已取消批量重跑";
      return;
    }
    errorMessage.value = normalizeErrorMessage(error, "批量重跑确认失败");
    return;
  }
  batchRetryingFailed.value = true;
  errorMessage.value = "";
  message.value = "";
  let successCount = 0;
  const failedIDs = [];
  try {
    let cursor = 0;
    async function worker() {
      while (cursor < failedRuns.length) {
        const index = cursor;
        cursor += 1;
        const item = failedRuns[index];
        if (!item) {
          continue;
        }
        try {
          await retrySchedulerJobRun(item.id, {});
          successCount += 1;
        } catch {
          failedIDs.push(item.id);
        }
      }
    }
    const workerCount = Math.min(concurrency, failedRuns.length);
    await Promise.all(Array.from({ length: workerCount }, () => worker()));
    await Promise.all([fetchMetrics(), fetchRuns()]);
    if (failedIDs.length === 0) {
      message.value = `已触发 ${successCount} 条失败任务重跑（并发 ${concurrency}）`;
      return;
    }
    message.value = `已触发 ${successCount} 条失败任务重跑（并发 ${concurrency}），${failedIDs.length} 条触发失败`;
    errorMessage.value = `触发失败运行ID：${failedIDs.join(", ")}`;
  } finally {
    batchRetryingFailed.value = false;
  }
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "SUCCESS" || normalized === "ACTIVE") return "success";
  if (normalized === "FAILED" || normalized === "DISABLED") return "danger";
  if (normalized === "RUNNING") return "warning";
  return "info";
}

function syncTypeLabel(syncType) {
  const normalized = String(syncType || "").trim().toUpperCase();
  const mapping = {
    NEWS_BRIEF: "新闻快讯",
    NEWS_MAJOR: "新闻通讯",
    RESEARCH_REPORT: "券商研报",
    REPORT_RC: "盈利预测",
    ANNOUNCEMENT: "上市公司公告"
  };
  return mapping[normalized] || normalized || "-";
}

function canRetryNewsSyncRow(row) {
  if (!row) {
    return false;
  }
  return String(row.status || "").trim().toUpperCase() === "FAILED" || Number(row.failed_count || 0) > 0;
}

function formatPercent(value, digits = 2) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return "-";
  }
  return `${(numeric * 100).toFixed(digits)}%`;
}

function parseTimeValue(value) {
  const text = String(value || "").trim();
  if (!text) {
    return NaN;
  }
  const normalized = text.includes("T") ? text : text.replace(" ", "T");
  return Date.parse(normalized);
}

function formatDuration(milliseconds) {
  const value = Number(milliseconds);
  if (!Number.isFinite(value) || value < 0) {
    return "-";
  }
  const totalSeconds = Math.floor(value / 1000);
  const hours = Math.floor(totalSeconds / 3600);
  const minutes = Math.floor((totalSeconds % 3600) / 60);
  const seconds = totalSeconds % 60;
  if (hours > 0) {
    return `${hours}小时${minutes}分${seconds}秒`;
  }
  if (minutes > 0) {
    return `${minutes}分${seconds}秒`;
  }
  return `${seconds}秒`;
}

function formatRunDuration(startedAt, finishedAt, status) {
  const startMs = parseTimeValue(startedAt);
  if (!Number.isFinite(startMs)) {
    return "-";
  }
  let endMs = parseTimeValue(finishedAt);
  if (!Number.isFinite(endMs) && String(status || "").trim().toUpperCase() === "RUNNING") {
    endMs = Date.now();
  }
  if (!Number.isFinite(endMs)) {
    return "-";
  }
  return formatDuration(Math.max(0, endMs - startMs));
}

function formatFailureReasonScope(scope) {
  const key = String(scope || "").trim().toUpperCase();
  const mapping = {
    TODAY: "今日",
    LAST_24_HOURS: "近24小时",
    LAST_3_DAYS: "近3天",
    LAST_7_DAYS: "近7天",
    LAST_30_DAYS: "近30天"
  };
  return mapping[key] || scope || "-";
}

function formatJobName(jobName) {
  const key = String(jobName || "").trim();
  if (!key) {
    return "-";
  }
  return jobLabelMap[key] || key;
}

function formatModule(module) {
  const key = String(module || "").trim().toUpperCase();
  const mapping = {
    STOCK: "股票",
    FUTURES: "期货",
    NEWS: "资讯",
    SYSTEM: "系统"
  };
  return mapping[key] || key || "-";
}

function formatDefinitionStatus(status) {
  const key = String(status || "").trim().toUpperCase();
  if (key === "ACTIVE") return "启用";
  if (key === "DISABLED") return "停用";
  return key || "-";
}

function formatRunStatus(status) {
  const key = String(status || "").trim().toUpperCase();
  if (key === "RUNNING") return "运行中";
  if (key === "SUCCESS") return "成功";
  if (key === "FAILED") return "失败";
  return key || "-";
}

function formatTriggerSource(source) {
  const key = String(source || "").trim().toUpperCase();
  if (key === "MANUAL") return "手动触发";
  if (key === "SYSTEM") return "系统触发";
  return key || "-";
}

function formatConfigKeyLabel(configKey) {
  const key = String(configKey || "").trim().toLowerCase();
  return configKeyLabelMap[key] || configKey || "-";
}

function formatAutoRetryJobs(value) {
  if (!Array.isArray(value) || value.length === 0) {
    return "-";
  }
  return value
    .map((jobName) => {
      const key = String(jobName || "").trim();
      if (!key) return "";
      return `${formatJobName(key)}(${key})`;
    })
    .filter(Boolean)
    .join("、");
}

function formatConfigDisplayValue(configKey, configValue) {
  const key = String(configKey || "").trim().toLowerCase();
  const value = String(configValue ?? "").trim();
  if (key === "scheduler.auto_retry.enabled") {
    return ["1", "true", "yes", "on", "y"].includes(value.toLowerCase()) ? "已开启" : "已关闭";
  }
  if (key === "scheduler.auto_retry.jobs") {
    if (!value) {
      return "不限制任务（全部）";
    }
    return value
      .split(",")
      .map((name) => name.trim())
      .filter(Boolean)
      .map((name) => `${formatJobName(name)}(${name})`)
      .join("、");
  }
  return value || "-";
}

function parseAutoRetrySummary(items) {
  const summary = {
    enabled: false,
    maxRetries: 2,
    backoffSeconds: 2,
    jobs: []
  };
  items.forEach((item) => {
    const key = String(item?.config_key || "").trim().toLowerCase();
    const value = String(item?.config_value || "").trim();
    if (!key) {
      return;
    }
    if (key === "scheduler.auto_retry.enabled") {
      summary.enabled = ["1", "true", "yes", "on", "y"].includes(value.toLowerCase());
      return;
    }
    if (key === "scheduler.auto_retry.max_retries") {
      const parsed = Number.parseInt(value, 10);
      if (Number.isFinite(parsed)) {
        summary.maxRetries = Math.max(0, parsed);
      }
      return;
    }
    if (key === "scheduler.auto_retry.backoff_seconds") {
      const parsed = Number.parseInt(value, 10);
      if (Number.isFinite(parsed)) {
        summary.backoffSeconds = Math.max(0, parsed);
      }
      return;
    }
    if (key === "scheduler.auto_retry.jobs") {
      summary.jobs = value
        .split(",")
        .map((name) => name.trim())
        .filter(Boolean);
    }
  });
  return summary;
}

function syncAutoRetryForm(summary) {
  const next = summary || {};
  autoRetryForm.enabled = !!next.enabled;
  autoRetryForm.max_retries = Number.isFinite(Number(next.maxRetries))
    ? Math.max(0, Math.min(5, Number(next.maxRetries)))
    : 2;
  autoRetryForm.backoff_seconds = Number.isFinite(Number(next.backoffSeconds))
    ? Math.max(0, Math.min(60, Number(next.backoffSeconds)))
    : 2;
  autoRetryForm.jobs = Array.isArray(next.jobs)
    ? next.jobs.map((name) => String(name || "").trim()).filter(Boolean)
    : [];
}

function resetAutoRetryForm() {
  syncAutoRetryForm(autoRetrySummary.value);
  message.value = "已重置为当前生效配置";
}

async function saveAutoRetryConfig() {
  if (!ensureCanEditSystemJobs()) {
    return;
  }
  const maxRetries = Math.max(0, Math.min(5, Number.parseInt(String(autoRetryForm.max_retries), 10) || 0));
  const backoffSeconds = Math.max(0, Math.min(60, Number.parseInt(String(autoRetryForm.backoff_seconds), 10) || 0));
  const jobs = Array.from(
    new Set(
      (autoRetryForm.jobs || [])
        .map((name) => String(name || "").trim())
        .filter(Boolean)
    )
  );
  const payloads = [
    {
      config_key: "scheduler.auto_retry.enabled",
      config_value: autoRetryForm.enabled ? "true" : "false",
      description: "是否开启任务自动重试"
    },
    {
      config_key: "scheduler.auto_retry.max_retries",
      config_value: String(maxRetries),
      description: "自动重试最大次数（0-5）"
    },
    {
      config_key: "scheduler.auto_retry.backoff_seconds",
      config_value: String(backoffSeconds),
      description: "自动重试退避秒数（0-60）"
    },
    {
      config_key: "scheduler.auto_retry.jobs",
      config_value: jobs.join(","),
      description: "允许自动重试的任务名，逗号分隔"
    }
  ];
  savingAutoRetry.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    await Promise.all(payloads.map((payload) => upsertSystemConfig(payload)));
    await Promise.all([fetchAutoRetryConfigs(), fetchAutoRetryChangeLogs(), fetchMetrics()]);
    message.value = "自动重试配置已保存并生效";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "保存自动重试配置失败");
  } finally {
    savingAutoRetry.value = false;
  }
}

function applyQuickTriggerJob(jobName) {
  triggerForm.job_name = String(jobName || "").trim();
  triggerForm.trigger_source = "MANUAL";
  triggerForm.simulate_status = "";
  triggerForm.result_summary = "";
  triggerForm.error_message = "";
  triggerForm.news_sources_text = "";
  triggerForm.symbols_text = "";
  triggerForm.sync_types = [];
  triggerForm.batch_size = undefined;
  message.value = `已选择任务：${triggerForm.job_name}`;
}

function parseRunSummary(rawSummary) {
  const text = String(rawSummary || "").trim();
  const result = {};
  if (!text) {
    return result;
  }
  text.split(/\s+/).forEach((part) => {
    const [key, value] = part.split("=");
    if (!key || value === undefined) {
      return;
    }
    result[key.trim()] = value.trim();
  });
  return result;
}

function runSummaryText(row, key, fallback = "-") {
  const summary = parseRunSummary(row?.result_summary);
  const value = String(summary[key] ?? "").trim();
  return value || fallback;
}

function runSummaryNumber(row, key, digits = 2) {
  const summary = parseRunSummary(row?.result_summary);
  const value = Number(summary[key]);
  if (!Number.isFinite(value)) {
    return "-";
  }
  return value.toFixed(digits);
}

function runSummaryPercent(row, key, digits = 2) {
  const summary = parseRunSummary(row?.result_summary);
  const value = Number(summary[key]);
  return formatPercent(value, digits);
}

watch(
  () => marketBackfillForm.run_type,
  (nextValue) => {
    const normalized = String(nextValue || "").trim().toUpperCase();
    if (normalized === "REBUILD_ONLY") {
      marketBackfillForm.stages = MARKET_BACKFILL_REBUILD_STAGES.slice();
      marketBackfillForm.rebuild_truth_after_sync = true;
      return;
    }
    const currentStages = Array.isArray(marketBackfillForm.stages) ? marketBackfillForm.stages : [];
    if (
      currentStages.length === 0 ||
      currentStages.every((item) => MARKET_BACKFILL_REBUILD_STAGES.includes(String(item || "").trim().toUpperCase()))
    ) {
      marketBackfillForm.stages = MARKET_BACKFILL_DEFAULT_STAGES.slice();
    }
  }
);

watch(activeTab, (nextTab) => {
  if (nextTab === "market-data" && !marketBackfillLoaded.value) {
    fetchMarketBackfillWorkspace({ preserveFeedback: true });
  }
});

onMounted(() => {
  activeTab.value = normalizeBackfillTab(route.query.tab);
  refreshAll();
});
</script>

<template>
  <div class="page">
    <div class="jobs-hero card">
      <div class="jobs-hero__top">
        <div>
          <div class="page-eyebrow">调度工作台</div>
          <h1 class="page-title">系统任务中心</h1>
          <p class="jobs-hero__desc">
            这里统一管理任务总览、自动重试、手动触发、任务定义和运行记录。先看异常，再做处理，操作路径会更顺。
          </p>
        </div>
        <div class="jobs-hero__top-actions">
          <el-tag :type="canEditSystemJobs ? 'success' : 'info'" effect="plain">
            {{ canEditSystemJobs ? "当前账号可配置、触发与重跑" : "当前账号仅支持查看" }}
          </el-tag>
          <el-tag type="warning" effect="plain">已配置 {{ definitionTotal }} 个任务定义</el-tag>
          <el-button :loading="defsLoading || runsLoading || metricsLoading" @click="refreshAll">刷新全部</el-button>
        </div>
      </div>

      <div class="jobs-action-grid">
        <div
          v-for="item in actionCards"
          :key="item.key"
          class="jobs-action-card"
          :class="`is-${item.tone}`"
        >
          <div class="jobs-action-card__title">{{ item.title }}</div>
          <div class="jobs-action-card__desc">{{ item.description }}</div>
          <el-button
            size="small"
            :type="actionButtonType(item.tone)"
            :plain="item.tone !== 'primary'"
            @click="handleActionCard(item.key)"
          >
            {{ item.actionText }}
          </el-button>
        </div>
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

    <div class="card jobs-tabs-shell">
      <el-tabs v-model="activeTab" type="border-card" class="jobs-tabs">
        <el-tab-pane
          v-for="tab in systemJobTabs"
          :key="tab.key"
          :label="tab.label"
          :name="tab.key"
        />
      </el-tabs>
      <div class="jobs-tab-summary">
        <div class="jobs-tab-summary__title">
          {{ systemJobTabs.find((item) => item.key === activeTab)?.label || "总览" }}
        </div>
        <div class="jobs-tab-summary__desc">
          {{ systemJobTabs.find((item) => item.key === activeTab)?.description || "看今日运行、失败原因和使用说明" }}
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'overview'" class="jobs-main-grid">
      <div class="card jobs-panel-card" style="margin-bottom: 12px">
        <div class="section-header section-header--stack">
          <div>
            <h3 style="margin: 0">运行总览</h3>
            <p class="section-copy">先看核心卡片，再往下看任务维度统计和失败原因，能更快判断今天该处理哪里。</p>
          </div>
          <div class="inline-actions inline-actions--left">
            <el-input
              v-model="metricFilter.job_name"
              clearable
              placeholder="按任务编码过滤指标（可选）"
              style="width: 260px"
            />
            <el-button :loading="metricsLoading" @click="fetchMetrics">刷新指标</el-button>
          </div>
        </div>

        <div class="jobs-overview-grid" v-loading="metricsLoading">
          <div
            v-for="item in overviewCards"
            :key="item.key"
            class="jobs-overview-card"
            :class="`is-${item.tone}`"
          >
            <div class="jobs-overview-card__title">{{ item.title }}</div>
            <div class="jobs-overview-card__value">{{ item.value }}</div>
            <div class="jobs-overview-card__helper">{{ item.helper }}</div>
          </div>
          <div class="jobs-overview-card is-info">
            <div class="jobs-overview-card__title">今日成功</div>
            <div class="jobs-overview-card__value">{{ metrics.today_success || 0 }}</div>
            <div class="jobs-overview-card__helper">结合失败数一起看，便于判断成功面是否稳定</div>
          </div>
          <div class="jobs-overview-card is-gold">
            <div class="jobs-overview-card__title">平均重试次数</div>
            <div class="jobs-overview-card__value">{{ Number(metrics.avg_retry_count || 0).toFixed(2) }}</div>
            <div class="jobs-overview-card__helper">重试次数持续升高时，建议复核任务配置或外部依赖</div>
          </div>
        </div>

        <div class="jobs-table-block">
          <div class="toolbar jobs-table-toolbar">
            <el-text type="primary">按任务维度的重试恢复统计</el-text>
            <span class="muted">用来判断问题集中在哪些任务，不要只盯总失败数。</span>
          </div>
          <el-table
            :data="metrics.job_retry_stats || []"
            border
            stripe
            size="small"
            empty-text="暂无按任务统计"
            v-loading="metricsLoading"
          >
            <el-table-column label="任务" min-width="220">
              <template #default="{ row }">
                {{ formatJobName(row.job_name) }} ({{ row.job_name || "-" }})
              </template>
            </el-table-column>
            <el-table-column prop="today_total" label="今日总运行" min-width="100" />
            <el-table-column prop="today_failed" label="今日失败" min-width="90" />
            <el-table-column prop="retry_total" label="重试数" min-width="80" />
            <el-table-column label="重试成功率" min-width="120">
              <template #default="{ row }">
                {{ formatPercent(row.retry_hit_rate) }}
              </template>
            </el-table-column>
            <el-table-column prop="auto_retry_total" label="自动重试触发" min-width="120" />
            <el-table-column label="恢复成功/触发" min-width="130">
              <template #default="{ row }">
                {{ row.recovery_success || 0 }}/{{ row.recovery_total || 0 }}
              </template>
            </el-table-column>
            <el-table-column label="恢复成功率" min-width="110">
              <template #default="{ row }">
                {{ formatPercent(row.recovery_hit_rate) }}
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="jobs-table-block">
          <div class="toolbar jobs-table-toolbar">
            <el-text type="primary">失败原因聚合</el-text>
            <el-tag type="info" effect="plain">{{ formatFailureReasonScope(metrics.failure_reason_scope || "LAST_7_DAYS") }}</el-tag>
            <span class="muted">先看共性原因，再决定是调配置、重跑，还是排查外部源。</span>
          </div>
          <el-table
            :data="metrics.failure_reasons || []"
            border
            stripe
            size="small"
            empty-text="暂无失败原因数据"
            v-loading="metricsLoading"
          >
            <el-table-column prop="reason" label="原因分类" min-width="280" />
            <el-table-column prop="count" label="次数" min-width="90" />
            <el-table-column prop="last_occurred_at" label="最近发生时间" min-width="190">
              <template #default="{ row }">
                {{ row.last_occurred_at || "-" }}
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="jobs-table-block">
          <div class="toolbar jobs-table-toolbar">
            <el-text type="primary">按任务失败原因</el-text>
            <el-select
              v-model="failureReasonJobFilter"
              clearable
              placeholder="按任务过滤失败原因"
              style="width: 260px"
            >
              <el-option
                v-for="name in failureReasonJobOptions"
                :key="name"
                :label="`${formatJobName(name)} (${name})`"
                :value="name"
              />
            </el-select>
            <span class="muted">适合定位某个任务的重复性故障。</span>
          </div>
          <el-table
            :data="filteredJobFailureReasons"
            border
            stripe
            size="small"
            empty-text="暂无按任务失败原因"
            v-loading="metricsLoading"
          >
            <el-table-column label="任务" min-width="220">
              <template #default="{ row }">
                {{ formatJobName(row.job_name) }} ({{ row.job_name || "-" }})
              </template>
            </el-table-column>
            <el-table-column prop="reason" label="原因分类" min-width="280" />
            <el-table-column prop="count" label="次数" min-width="90" />
            <el-table-column prop="last_occurred_at" label="最近发生时间" min-width="190">
              <template #default="{ row }">
                {{ row.last_occurred_at || "-" }}
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>

      <div class="jobs-side-column">
        <div class="card jobs-panel-card jobs-guide-panel">
          <div class="section-header section-header--stack">
            <div>
              <h3 style="margin: 0">使用说明</h3>
              <p class="section-copy">把常用处理顺序和权限边界放在右边，避免每次都去翻说明。</p>
            </div>
          </div>
          <div class="jobs-guide-list">
            <div v-for="card in guideCards" :key="card.key" class="jobs-guide-card">
              <div class="jobs-guide-card__title">{{ card.title }}</div>
              <ul class="jobs-guide-card__list">
                <li v-for="item in card.items" :key="item">{{ item }}</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'market-data'" class="jobs-main-grid">
      <div class="jobs-tab-stack">
        <div class="card jobs-panel-card">
          <div class="section-header section-header--stack">
            <div>
              <h3 style="margin: 0">市场数据回填工作台</h3>
              <p class="section-copy">
                先看最近回填总单和 Universe 快照，再决定是新建全量回填，还是只重试失败批次。
              </p>
            </div>
            <div class="inline-actions inline-actions--left">
              <el-button
                :loading="marketBackfillRunsLoading || marketBackfillSnapshotsLoading"
                @click="fetchMarketBackfillWorkspace"
              >
                刷新工作台
              </el-button>
              <el-button v-if="canEditSystemJobs" @click="resetMarketBackfillForm">重置新建表单</el-button>
            </div>
          </div>

          <div
            class="jobs-overview-grid"
            v-loading="marketBackfillRunsLoading || marketBackfillSnapshotsLoading"
          >
            <div
              v-for="item in marketBackfillOverviewCards"
              :key="item.key"
              class="jobs-overview-card"
              :class="`is-${item.tone}`"
            >
              <div class="jobs-overview-card__title">{{ item.title }}</div>
              <div class="jobs-overview-card__value">{{ item.value }}</div>
              <div class="jobs-overview-card__helper">{{ item.helper }}</div>
            </div>
          </div>
        </div>

        <div class="card jobs-panel-card">
          <div class="section-header section-header--stack">
            <div>
              <h3 style="margin: 0">新建回填任务</h3>
              <p class="section-copy">
                这里只发起真实回填总单。全量、增量和仅重建都走同一套链路，不再分散在多个入口。
              </p>
            </div>
          </div>

          <el-form label-position="top">
            <div class="market-backfill-form-grid">
              <el-form-item label="运行类型" required>
                <el-select v-model="marketBackfillForm.run_type">
                  <el-option
                    v-for="item in marketBackfillRunTypeOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="来源">
                <el-input v-model="marketBackfillForm.source_key" placeholder="默认 TUSHARE" />
              </el-form-item>
              <el-form-item label="批量大小">
                <el-input-number
                  v-model="marketBackfillForm.batch_size"
                  :min="20"
                  :max="500"
                  :step="20"
                  controls-position="right"
                />
              </el-form-item>
              <el-form-item label="资产范围" required>
                <el-select
                  v-model="marketBackfillForm.asset_scope"
                  multiple
                  collapse-tags
                  collapse-tags-tooltip
                  placeholder="选择要回填的资产"
                >
                  <el-option
                    v-for="item in marketBackfillAssetOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="阶段范围">
                <el-select
                  v-model="marketBackfillForm.stages"
                  multiple
                  collapse-tags
                  collapse-tags-tooltip
                  placeholder="不选则按默认阶段执行"
                >
                  <el-option
                    v-for="item in marketBackfillStageOptions"
                    :key="item.value"
                    :label="item.label"
                    :value="item.value"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="开始日期">
                <el-date-picker
                  v-model="marketBackfillForm.trade_date_from"
                  type="date"
                  value-format="YYYY-MM-DD"
                  format="YYYY-MM-DD"
                  placeholder="可选"
                />
              </el-form-item>
              <el-form-item label="结束日期">
                <el-date-picker
                  v-model="marketBackfillForm.trade_date_to"
                  type="date"
                  value-format="YYYY-MM-DD"
                  format="YYYY-MM-DD"
                  placeholder="可选"
                />
              </el-form-item>
              <el-form-item label="任务选项" class="market-backfill-switches">
                <el-switch
                  v-model="marketBackfillForm.force_refresh_universe"
                  active-text="强制刷新 Universe"
                />
                <el-switch
                  v-model="marketBackfillForm.rebuild_truth_after_sync"
                  active-text="同步后重建 Truth"
                />
              </el-form-item>
            </div>
          </el-form>

          <el-alert
            title="股票支持行情、daily_basic、moneyflow 和 truth；其他资产当前先走 universe、master、quotes 与 truth。"
            type="info"
            :closable="false"
            show-icon
          />

          <div class="toolbar" style="margin-top: 10px">
            <el-button
              v-if="canEditSystemJobs"
              type="primary"
              :loading="marketBackfillCreating"
              @click="handleCreateMarketBackfillRun"
            >
              发起回填
            </el-button>
            <el-text type="info">
              当前默认会创建总单并直接执行，执行结果会同步回写任务运行记录。
            </el-text>
          </div>
        </div>

        <div class="card jobs-panel-card">
          <div class="section-header section-header--stack">
            <div>
              <h3 style="margin: 0">回填总单</h3>
              <p class="section-copy">
                这里先看每次回填的状态、阶段和资产范围。真正定位问题时，再展开批次明细。
              </p>
            </div>
          </div>

          <div class="toolbar jobs-table-toolbar">
            <el-select
              v-model="marketBackfillFilters.status"
              clearable
              placeholder="全部状态"
              style="width: 150px"
            >
              <el-option
                v-for="item in marketBackfillStatusOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-select
              v-model="marketBackfillFilters.run_type"
              clearable
              placeholder="全部类型"
              style="width: 150px"
            >
              <el-option
                v-for="item in marketBackfillRunTypeOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-select
              v-model="marketBackfillFilters.asset_type"
              clearable
              placeholder="全部资产"
              style="width: 150px"
            >
              <el-option
                v-for="item in marketBackfillAssetOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-input
              v-model="marketBackfillFilters.source_key"
              clearable
              placeholder="按来源过滤"
              style="width: 180px"
            />
            <el-button type="primary" plain @click="applyMarketBackfillFilters">查询</el-button>
            <el-button @click="resetMarketBackfillFilters">重置</el-button>
          </div>

          <el-table
            :data="marketBackfillRuns"
            border
            stripe
            v-loading="marketBackfillRunsLoading"
            empty-text="暂无市场回填总单"
          >
            <el-table-column prop="id" label="总单ID" min-width="180" />
            <el-table-column label="运行类型" min-width="110">
              <template #default="{ row }">
                {{ formatMarketBackfillRunTypeLabel(row.run_type) }}
              </template>
            </el-table-column>
            <el-table-column label="资产范围" min-width="180">
              <template #default="{ row }">
                {{ formatMarketAssetScopeSummary(row.asset_scope) }}
              </template>
            </el-table-column>
            <el-table-column prop="source_key" label="来源" min-width="100" />
            <el-table-column label="当前阶段" min-width="120">
              <template #default="{ row }">
                {{ formatMarketBackfillStageLabel(row.current_stage) }}
              </template>
            </el-table-column>
            <el-table-column label="状态" min-width="110">
              <template #default="{ row }">
                <el-tag :type="marketBackfillStatusTagType(row.status)">
                  {{ formatMarketBackfillStatusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="scheduler_run_id" label="调度运行ID" min-width="170" />
            <el-table-column prop="created_at" label="创建时间" min-width="180" />
            <el-table-column label="操作" min-width="180" align="right">
              <template #default="{ row }">
                <div class="inline-actions">
                  <el-button size="small" @click="handleOpenMarketBackfillRun(row)">查看详情</el-button>
                  <el-button
                    v-if="canEditSystemJobs && canRetryMarketBackfillRun(row)"
                    size="small"
                    type="danger"
                    plain
                    :loading="Boolean(marketBackfillRetryingMap[row.id])"
                    @click="handleRetryMarketBackfillRun(row)"
                  >
                    重试失败批次
                  </el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">
              第 {{ marketBackfillRunPage }} 页，共 {{ marketBackfillRunTotal }} 条
            </el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="marketBackfillRunPage"
              :page-size="marketBackfillRunPageSize"
              :total="marketBackfillRunTotal"
              @current-change="handleMarketBackfillRunPageChange"
            />
          </div>
        </div>

        <div class="card jobs-panel-card">
          <div class="section-header section-header--stack">
            <div>
              <h3 style="margin: 0">最近 Universe 快照</h3>
              <p class="section-copy">
                快照是回填链路的第一步。先确认证券全集覆盖和来源，再决定后续补数范围。
              </p>
            </div>
          </div>

          <el-table
            :data="marketBackfillSnapshots"
            border
            stripe
            v-loading="marketBackfillSnapshotsLoading"
            empty-text="暂无 Universe 快照"
          >
            <el-table-column prop="snapshot_date" label="快照日期" min-width="110" />
            <el-table-column label="摘要" min-width="260">
              <template #default="{ row }">
                {{ buildUniverseSnapshotDigest(row) }}
              </template>
            </el-table-column>
            <el-table-column label="资产拆分" min-width="220">
              <template #default="{ row }">
                <div class="market-backfill-asset-summary">
                  <el-tag
                    v-for="asset in row.asset_summaries || []"
                    :key="`${row.id}-${asset.asset_type}`"
                    type="info"
                    effect="plain"
                  >
                    {{ formatMarketAssetLabel(asset.asset_type) }} {{ asset.item_count || 0 }}
                  </el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="created_by" label="创建人" min-width="120" />
            <el-table-column prop="created_at" label="创建时间" min-width="180" />
          </el-table>
        </div>
      </div>

      <div class="jobs-side-column">
        <div class="card jobs-panel-card jobs-guide-panel">
          <div class="section-header section-header--stack">
            <div>
              <h3 style="margin: 0">使用说明</h3>
              <p class="section-copy">把市场数据的处理顺序、支持范围和权限边界固定放在右边。</p>
            </div>
          </div>
          <div class="jobs-guide-list">
            <div v-for="card in marketBackfillGuideCards" :key="card.key" class="jobs-guide-card">
              <div class="jobs-guide-card__title">{{ card.title }}</div>
              <ul class="jobs-guide-card__list">
                <li v-for="item in card.items" :key="item">{{ item }}</li>
              </ul>
            </div>
          </div>
        </div>

        <div class="card jobs-panel-card">
          <div class="section-header section-header--stack">
            <div>
              <h3 style="margin: 0">当前窗口提示</h3>
              <p class="section-copy">方便快速判断最近一次回填和最近一次快照是否正常落库。</p>
            </div>
          </div>
          <div class="jobs-guide-list">
            <div class="jobs-guide-card">
              <div class="jobs-guide-card__title">最近回填</div>
              <p class="jobs-guide-card__desc">
                {{
                  marketBackfillRuns.length
                    ? `${formatMarketBackfillRunTypeLabel(marketBackfillRuns[0].run_type)} · ${formatMarketBackfillStatusLabel(marketBackfillRuns[0].status)}`
                    : "暂无回填总单"
                }}
              </p>
            </div>
            <div class="jobs-guide-card">
              <div class="jobs-guide-card__title">最近快照</div>
              <p class="jobs-guide-card__desc">
                {{ marketBackfillSnapshots.length ? buildUniverseSnapshotDigest(marketBackfillSnapshots[0]) : "暂无 Universe 快照" }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div v-show="activeTab === 'config' || activeTab === 'trigger'" class="jobs-workbench-grid">
    <div v-show="activeTab === 'config'" class="card jobs-panel-card" style="margin-bottom: 12px">
      <div class="section-header">
        <div>
          <h3 style="margin: 0">自动重试配置</h3>
          <p class="section-copy">先看当前生效状态，再决定要不要调整次数、退避和任务范围。</p>
        </div>
        <div class="inline-actions inline-actions--left">
          <el-button :loading="autoRetryLoading" @click="fetchAutoRetryConfigs">刷新配置</el-button>
          <el-button v-if="canEditSystemJobs" @click="resetAutoRetryForm">重置编辑</el-button>
          <el-button
            v-if="canEditSystemJobs"
            type="primary"
            :loading="savingAutoRetry"
            @click="saveAutoRetryConfig"
          >
            保存配置
          </el-button>
        </div>
      </div>
      <div class="grid grid-4" v-loading="autoRetryLoading || savingAutoRetry">
        <div class="metric-item">
          <div class="metric-label">自动重试状态</div>
          <div class="metric-value metric-value--small">
            <el-tag :type="autoRetrySummary.enabled ? 'success' : 'info'">
              {{ autoRetrySummary.enabled ? "已开启" : "已关闭" }}
            </el-tag>
          </div>
        </div>
        <div class="metric-item">
          <div class="metric-label">最大重试次数</div>
          <div class="metric-value">{{ autoRetrySummary.maxRetries || 0 }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">退避秒数</div>
          <div class="metric-value">{{ autoRetrySummary.backoffSeconds || 0 }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">允许任务</div>
          <div class="metric-value metric-value--small">
            {{ formatAutoRetryJobs(autoRetrySummary.jobs) }}
          </div>
        </div>
      </div>
      <el-form label-width="140px" style="margin-top: 10px">
        <div class="dialog-grid">
          <el-form-item label="启用自动重试">
            <el-switch v-model="autoRetryForm.enabled" :disabled="!canEditSystemJobs" />
          </el-form-item>
          <el-form-item label="最大重试次数(0-5)">
            <el-input-number
              v-model="autoRetryForm.max_retries"
              :min="0"
              :max="5"
              :step="1"
              controls-position="right"
              :disabled="!canEditSystemJobs"
            />
          </el-form-item>
          <el-form-item label="退避秒数(0-60)">
            <el-input-number
              v-model="autoRetryForm.backoff_seconds"
              :min="0"
              :max="60"
              :step="1"
              controls-position="right"
              :disabled="!canEditSystemJobs"
            />
          </el-form-item>
          <el-form-item label="允许自动重试任务">
            <el-select
              v-model="autoRetryForm.jobs"
              multiple
              filterable
              allow-create
              default-first-option
              clearable
              collapse-tags
              collapse-tags-tooltip
              :disabled="!canEditSystemJobs"
              placeholder="不选表示不过滤任务"
            >
              <el-option
                v-for="item in autoRetryJobOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
        </div>
      </el-form>
      <el-alert
        title="自动重试只对首次执行失败的任务生效；任务列表留空表示不限制任务名。"
        type="info"
        :closable="false"
        show-icon
        style="margin-top: 8px"
      />
      <el-table
        :data="autoRetryConfigRows"
        border
        stripe
        size="small"
        empty-text="暂无自动重试配置项"
        style="margin-top: 10px"
        v-loading="autoRetryLoading"
      >
        <el-table-column label="配置项" min-width="260">
          <template #default="{ row }">
            {{ formatConfigKeyLabel(row.config_key) }}
            <div class="muted">{{ row.config_key || "-" }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="config_value" label="配置值" min-width="160" />
        <el-table-column prop="description" label="说明" min-width="220" />
        <el-table-column prop="updated_by" label="更新人" min-width="120" />
        <el-table-column prop="updated_at" label="更新时间" min-width="180" />
      </el-table>

      <div class="toolbar" style="margin-top: 10px; margin-bottom: 6px">
        <el-text type="primary">自动重试配置变更记录</el-text>
        <el-button size="small" :loading="autoRetryLogLoading" @click="fetchAutoRetryChangeLogs">刷新记录</el-button>
      </div>
      <el-table
        :data="autoRetryChangeLogs"
        border
        stripe
        size="small"
        empty-text="暂无变更记录"
        v-loading="autoRetryLogLoading"
      >
        <el-table-column label="配置项" min-width="260">
          <template #default="{ row }">
            {{ formatConfigKeyLabel(row.target_id) }}
            <div class="muted">{{ row.target_id || "-" }}</div>
          </template>
        </el-table-column>
        <el-table-column label="变更后值" min-width="220">
          <template #default="{ row }">
            <span class="run-preview">{{ formatConfigDisplayValue(row.target_id, row.after_value) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="备注" min-width="180">
          <template #default="{ row }">
            {{ row.reason || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="operator_user_id" label="操作人" min-width="120" />
        <el-table-column prop="created_at" label="变更时间" min-width="180" />
      </el-table>
    </div>

    <div v-if="canEditSystemJobs" v-show="activeTab === 'trigger'" class="card jobs-panel-card" style="margin-bottom: 12px">
      <div class="section-header">
        <div>
          <h3 style="margin: 0">手动触发任务</h3>
          <p class="section-copy">用于临时补跑、联调验证或主动触发一次指定任务，不影响已有任务定义。</p>
        </div>
      </div>
      <div class="toolbar" style="margin-bottom: 8px">
        <el-text type="info">快捷选择：</el-text>
        <el-button
          v-for="item in quickJobOptions"
          :key="item.value"
          size="small"
          @click="applyQuickTriggerJob(item.value)"
        >
          {{ item.label }}
        </el-button>
      </div>
      <el-form label-width="125px">
        <div class="dialog-grid">
          <el-form-item label="任务编码" required>
            <el-select
              v-model="triggerForm.job_name"
              filterable
              allow-create
              default-first-option
              clearable
              placeholder="请选择或输入任务编码"
            >
              <el-option
                v-for="item in triggerJobOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="触发来源" required>
            <el-select v-model="triggerForm.trigger_source">
              <el-option label="手动触发" value="MANUAL" />
              <el-option label="系统触发" value="SYSTEM" />
            </el-select>
          </el-form-item>
          <el-form-item label="模拟结果">
            <el-select v-model="triggerForm.simulate_status" clearable placeholder="可选（仅模拟模式）">
              <el-option v-for="item in simulateStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="运行摘要">
            <el-input v-model="triggerForm.result_summary" placeholder="例如：处理 120 条数据" />
          </el-form-item>
          <el-form-item label="错误信息">
            <el-input v-model="triggerForm.error_message" placeholder="可填写模拟失败信息" />
          </el-form-item>
          <el-form-item label="资讯同步类型">
            <el-select
              v-model="triggerForm.sync_types"
              multiple
              clearable
              collapse-tags
              collapse-tags-tooltip
              placeholder="Tushare任务可选"
            >
              <el-option
                v-for="item in newsSyncTypeOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="资讯来源">
            <el-input
              v-model="triggerForm.news_sources_text"
              placeholder="来源筛选，逗号分隔，如 cls,财联社"
            />
          </el-form-item>
          <el-form-item label="股票代码">
            <el-input
              v-model="triggerForm.symbols_text"
              placeholder="股票代码筛选，逗号分隔，如 600519.SH,000001.SZ"
            />
          </el-form-item>
          <el-form-item label="批量大小">
            <el-input-number
              v-model="triggerForm.batch_size"
              :min="1"
              :max="1000"
              :step="10"
              controls-position="right"
              placeholder="可选"
            />
          </el-form-item>
        </div>
      </el-form>
      <div class="toolbar" style="margin-bottom: 8px">
        <el-button type="primary" :loading="triggeringJob" @click="submitTrigger">触发任务</el-button>
      </div>
      <el-alert
        title="“模拟结果”仅在后端开启 ALLOW_JOB_SIMULATION=true 时生效"
        type="info"
        :closable="false"
        show-icon
      />
    </div>
    </div>

    <div v-show="activeTab === 'config'" class="card jobs-panel-card" style="margin-bottom: 12px">
      <div class="section-header">
        <div>
          <h3 style="margin: 0">任务定义</h3>
          <p class="section-copy">维护任务编码、调度表达式、所属模块和启停状态。这里是任务中心的“配置台账”。</p>
        </div>
        <el-button v-if="canEditSystemJobs" type="primary" @click="openCreateDefinition">新增任务定义</el-button>
      </div>

      <div class="toolbar">
        <el-select v-model="definitionFilters.status" clearable placeholder="全部状态" style="width: 150px">
          <el-option v-for="item in definitionStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-select v-model="definitionFilters.module" clearable placeholder="全部模块" style="width: 150px">
          <el-option v-for="item in moduleOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-button type="primary" plain @click="applyDefinitionFilters">查询</el-button>
        <el-button @click="resetDefinitionFilters">重置</el-button>
      </div>

      <el-table :data="definitions" border stripe v-loading="defsLoading" empty-text="暂无任务定义">
        <el-table-column prop="id" label="任务ID" min-width="140" />
        <el-table-column label="任务编码" min-width="220">
          <template #default="{ row }">
            {{ formatJobName(row.job_name) }}
            <div class="muted">{{ row.job_name || "-" }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="display_name" label="任务名称" min-width="180" />
        <el-table-column label="所属模块" min-width="100">
          <template #default="{ row }">
            {{ formatModule(row.module) }}
          </template>
        </el-table-column>
        <el-table-column prop="cron_expr" label="调度表达式" min-width="180" />
        <el-table-column label="状态" min-width="150">
          <template #default="{ row }">
            <el-switch
              v-if="canEditSystemJobs"
              :model-value="(definitionStatusMap[row.id] || row.status) === 'ACTIVE'"
              inline-prompt
              active-text="启用"
              inactive-text="停用"
              :loading="Boolean(definitionStatusSavingMap[row.id])"
              @change="(active) => onDefinitionStatusSwitch(row, active)"
            />
            <el-tag v-else :type="statusTagType(row.status)">
              {{ formatDefinitionStatus(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最近运行时间" min-width="180">
          <template #default="{ row }">
            {{ row.last_run_at || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="更新人" min-width="130">
          <template #default="{ row }">
            {{ row.updated_by || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="180">
          <template #default="{ row }">
            {{ row.updated_at || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="操作" align="right" min-width="190">
          <template #default="{ row }">
            <div class="inline-actions">
              <template v-if="canEditSystemJobs">
                <el-button size="small" @click="openEditDefinition(row)">编辑</el-button>
                <el-button
                  size="small"
                  type="danger"
                  plain
                  :loading="Boolean(definitionDeletingMap[row.id])"
                  @click="removeDefinition(row)"
                >
                  删除
                </el-button>
              </template>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-text type="info">第 {{ definitionPage }} 页，共 {{ definitionTotal }} 条</el-text>
        <el-pagination
          background
          layout="prev, pager, next"
          :current-page="definitionPage"
          :page-size="definitionPageSize"
          :total="definitionTotal"
          @current-change="handleDefinitionPageChange"
        />
      </div>
    </div>

    <div v-show="activeTab === 'runs'" class="card jobs-panel-card">
      <div class="section-header">
        <div>
          <h3 style="margin: 0">运行记录</h3>
          <p class="section-copy">这里承接详情、失败重跑、批量重跑和导出，是日常排障和复核的主工作区。</p>
        </div>
      </div>

      <div class="toolbar">
        <el-input v-model="runFilters.job_name" clearable placeholder="按任务编码过滤" style="width: 220px" />
        <el-select v-model="runFilters.status" clearable placeholder="全部状态" style="width: 150px">
          <el-option v-for="item in runStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-text v-if="canEditSystemJobs" type="info">批量并发</el-text>
        <el-input-number
          v-if="canEditSystemJobs"
          v-model="batchRetryConcurrency"
          :min="1"
          :max="10"
          :step="1"
          controls-position="right"
          style="width: 180px"
        />
        <el-button
          v-if="canEditSystemJobs"
          :loading="batchRetryingFailed"
          @click="retryFailedRunsCurrentPage"
        >
          一键重跑当前页失败任务
        </el-button>
        <el-button @click="applyFailedRunFilter">只看失败</el-button>
        <el-button :loading="exportingRuns" @click="exportRunFilteredCSV">导出筛选CSV</el-button>
        <el-button @click="exportRunCurrentPageCSV">导出当前页CSV</el-button>
        <el-button type="primary" plain @click="applyRunFilters">查询</el-button>
        <el-button @click="resetRunFilters">重置</el-button>
      </div>

      <el-table :data="runs" border stripe v-loading="runsLoading" empty-text="暂无运行记录">
        <el-table-column prop="id" label="运行ID" min-width="120" />
        <el-table-column label="上级运行ID" min-width="130">
          <template #default="{ row }">
            {{ row.parent_run_id || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="任务" min-width="220">
          <template #default="{ row }">
            {{ formatJobName(row.job_name) }}
            <div class="muted">{{ row.job_name || "-" }}</div>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ formatRunStatus(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="retry_count" label="重试次数" min-width="110" />
        <el-table-column label="触发来源" min-width="120">
          <template #default="{ row }">
            {{ formatTriggerSource(row.trigger_source) }}
          </template>
        </el-table-column>
        <el-table-column prop="started_at" label="开始时间" min-width="180" />
        <el-table-column label="结束时间" min-width="180">
          <template #default="{ row }">
            {{ row.finished_at || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="执行耗时" min-width="120">
          <template #default="{ row }">
            {{ formatRunDuration(row.started_at, row.finished_at, row.status) }}
          </template>
        </el-table-column>
        <el-table-column label="运行摘要" min-width="220">
          <template #default="{ row }">
            <template v-if="row.job_name === 'daily_stock_quant_pipeline'">
              <div class="run-preview run-summary-tags">
                <el-tag size="small" type="info">交易日={{ runSummaryText(row, "trade_date") }}</el-tag>
                <el-tag size="small" type="warning">数据源={{ runSummaryText(row, "source") }}</el-tag>
                <el-tag size="small" type="success">行情数={{ runSummaryText(row, "quotes") }}</el-tag>
                <el-tag size="small" type="success">前十数={{ runSummaryText(row, "top") }}</el-tag>
                <el-tag size="small" type="success">推荐数={{ runSummaryText(row, "recommendations") }}</el-tag>
              </div>
            </template>
            <template v-else-if="isFuturesGenerateRun(row)">
              <div class="run-preview run-summary-tags">
                <el-tag size="small" type="info">交易日={{ runSummaryText(row, "trade_date") }}</el-tag>
                <el-tag size="small" type="success">生成数={{ runSummaryText(row, "generated") }}</el-tag>
              </div>
            </template>
            <template v-else-if="isFuturesEvaluateRun(row)">
              <div class="run-preview run-summary-tags">
                <el-tag size="small" type="info">交易日={{ runSummaryText(row, "trade_date") }}</el-tag>
                <el-tag size="small" type="success">成功={{ runSummaryText(row, "success") }}/{{ runSummaryText(row, "evaluated") }}</el-tag>
                <el-tag size="small" type="warning">均分={{ runSummaryNumber(row, "avg_score", 2) }}</el-tag>
                <el-tag size="small" type="success">胜率={{ runSummaryPercent(row, "avg_win_rate", 1) }}</el-tag>
                <el-tag size="small" type="danger">回撤={{ runSummaryPercent(row, "max_drawdown", 2) }}</el-tag>
              </div>
            </template>
            <template v-else>
              <span class="run-preview">{{ row.result_summary || "-" }}</span>
            </template>
          </template>
        </el-table-column>
        <el-table-column label="错误信息" min-width="220">
          <template #default="{ row }">
            <span class="run-preview">{{ row.error_message || "-" }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" align="right" min-width="470">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-button size="small" @click="openRunDetail(row)">
                {{ isTushareRun(row) ? "同步明细" : isFuturesRun(row) ? "评估明细" : "详情" }}
              </el-button>
              <template v-if="canEditSystemJobs">
                <el-select v-model="retrySimMap[row.id]" size="small" clearable placeholder="模拟状态" style="width: 130px">
                  <el-option v-for="item in simulateStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
                </el-select>
                <el-input v-model="retrySummaryMap[row.id]" size="small" placeholder="重跑摘要" style="width: 130px" />
                <el-input v-model="retryErrorMap[row.id]" size="small" placeholder="重跑错误信息" style="width: 140px" />
                <el-button size="small" type="primary" @click="retryRun(row.id)">重跑</el-button>
              </template>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-text type="info">第 {{ runPage }} 页，共 {{ runTotal }} 条</el-text>
        <el-pagination
          background
          layout="prev, pager, next"
          :current-page="runPage"
          :page-size="runPageSize"
          :total="runTotal"
          @current-change="handleRunPageChange"
        />
      </div>
    </div>

    <el-drawer
      v-model="marketBackfillDetailVisible"
      size="920px"
      destroy-on-close
      @closed="handleMarketBackfillDetailClosed"
    >
      <template #header>
        <div class="drawer-title">市场回填详情</div>
      </template>
      <template v-if="currentMarketBackfillRun">
        <div class="run-detail-title-row">
          <h4>总单概览</h4>
          <div class="inline-actions inline-actions--left">
            <el-tag :type="marketBackfillStatusTagType(currentMarketBackfillRun.status)">
              {{ formatMarketBackfillStatusLabel(currentMarketBackfillRun.status) }}
            </el-tag>
            <el-button
              size="small"
              :loading="marketBackfillRunDetailLoading || marketBackfillDetailsLoading"
              @click="
                Promise.all([
                  fetchCurrentMarketBackfillRun(currentMarketBackfillRun.id),
                  fetchMarketBackfillRunDetails()
                ])
              "
            >
              刷新详情
            </el-button>
            <el-button
              v-if="canEditSystemJobs && canRetryMarketBackfillRun(currentMarketBackfillRun)"
              size="small"
              type="danger"
              plain
              :loading="Boolean(marketBackfillRetryingMap[currentMarketBackfillRun.id])"
              @click="handleRetryMarketBackfillRun(currentMarketBackfillRun)"
            >
              重试失败批次
            </el-button>
          </div>
        </div>

        <el-descriptions :column="2" border size="small" v-loading="marketBackfillRunDetailLoading">
          <el-descriptions-item label="总单ID">{{ currentMarketBackfillRun.id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="调度运行ID">
            {{ currentMarketBackfillRun.scheduler_run_id || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="运行类型">
            {{ formatMarketBackfillRunTypeLabel(currentMarketBackfillRun.run_type) }}
          </el-descriptions-item>
          <el-descriptions-item label="来源">
            {{ currentMarketBackfillRun.source_key || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="资产范围">
            {{ formatMarketAssetScopeSummary(currentMarketBackfillRun.asset_scope) }}
          </el-descriptions-item>
          <el-descriptions-item label="当前阶段">
            {{ formatMarketBackfillStageLabel(currentMarketBackfillRun.current_stage) }}
          </el-descriptions-item>
          <el-descriptions-item label="Universe 快照">
            {{ currentMarketBackfillRun.universe_snapshot_id || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="批量大小">
            {{ currentMarketBackfillRun.batch_size || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="交易区间">
            {{ formatMarketBackfillTimeRange(currentMarketBackfillRun) }}
          </el-descriptions-item>
          <el-descriptions-item label="创建人">
            {{ currentMarketBackfillRun.created_by || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ currentMarketBackfillRun.created_at || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="完成时间">
            {{ currentMarketBackfillRun.finished_at || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="错误信息" :span="2">
            {{ currentMarketBackfillRun.error_message || "-" }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="run-detail-block">
          <div class="run-detail-title-row">
            <h4>阶段进度</h4>
            <el-text type="info">每个阶段都显示总批次、已完成、失败和跳过数量。</el-text>
          </div>
          <div class="market-backfill-stage-grid">
            <div
              v-for="item in currentMarketBackfillRun.stage_progress || []"
              :key="`${currentMarketBackfillRun.id}-${item.stage}`"
              class="market-backfill-stage-card"
              :class="`is-${marketBackfillDetailStatusTagType(item.status)}`"
            >
              <div class="market-backfill-stage-card__top">
                <div class="market-backfill-stage-card__title">
                  {{ formatMarketBackfillStageLabel(item.stage) }}
                </div>
                <el-tag size="small" :type="marketBackfillDetailStatusTagType(item.status)">
                  {{ formatMarketBackfillStatusLabel(item.status) }}
                </el-tag>
              </div>
              <div class="market-backfill-stage-card__desc">
                {{ formatMarketStageProgressSummary(item) }}
              </div>
            </div>
            <div
              v-if="!(currentMarketBackfillRun.stage_progress || []).length"
              class="market-backfill-empty"
            >
              暂无阶段进度
            </div>
          </div>
        </div>

        <div class="run-detail-block">
          <div class="run-detail-title-row">
            <h4>回填摘要</h4>
          </div>
          <pre class="run-detail-pre">{{ formatMarketBackfillSummary(currentMarketBackfillRun.summary) }}</pre>
        </div>

        <div class="run-detail-block">
          <div class="run-detail-title-row">
            <h4>批次明细</h4>
            <el-button size="small" :loading="marketBackfillDetailsLoading" @click="fetchMarketBackfillRunDetails">
              刷新明细
            </el-button>
          </div>
          <div class="toolbar">
            <el-select
              v-model="marketBackfillDetailFilters.stage"
              clearable
              placeholder="全部阶段"
              style="width: 160px"
            >
              <el-option
                v-for="item in marketBackfillStageOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-select
              v-model="marketBackfillDetailFilters.asset_type"
              clearable
              placeholder="全部资产"
              style="width: 150px"
            >
              <el-option
                v-for="item in marketBackfillAssetOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-select
              v-model="marketBackfillDetailFilters.status"
              clearable
              placeholder="全部状态"
              style="width: 150px"
            >
              <el-option
                v-for="item in marketBackfillDetailStatusOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-button type="primary" plain @click="applyMarketBackfillDetailFilters">查询</el-button>
            <el-button @click="resetMarketBackfillDetailFilters">重置</el-button>
          </div>
          <el-table
            :data="marketBackfillDetails"
            border
            stripe
            size="small"
            v-loading="marketBackfillDetailsLoading"
            empty-text="暂无批次明细"
          >
            <el-table-column label="阶段" min-width="120">
              <template #default="{ row }">
                {{ formatMarketBackfillStageLabel(row.stage) }}
              </template>
            </el-table-column>
            <el-table-column label="资产" min-width="90">
              <template #default="{ row }">
                {{ formatMarketAssetLabel(row.asset_type) }}
              </template>
            </el-table-column>
            <el-table-column prop="batch_key" label="批次键" min-width="160" />
            <el-table-column prop="source_key" label="来源" min-width="100" />
            <el-table-column prop="symbol_count" label="样本数" min-width="80" />
            <el-table-column label="样本示例" min-width="180">
              <template #default="{ row }">
                <span class="run-preview">{{ formatMarketBackfillBatchSample(row) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="交易区间" min-width="160">
              <template #default="{ row }">
                {{ formatMarketBackfillTimeRange(row) }}
              </template>
            </el-table-column>
            <el-table-column label="状态" min-width="90">
              <template #default="{ row }">
                <el-tag :type="marketBackfillDetailStatusTagType(row.status)">
                  {{ formatMarketBackfillStatusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="fetched_count" label="抓取" min-width="70" />
            <el-table-column prop="upserted_count" label="入库" min-width="70" />
            <el-table-column prop="truth_count" label="Truth" min-width="70" />
            <el-table-column label="告警/错误" min-width="240">
              <template #default="{ row }">
                <span class="run-preview">{{ row.error_text || row.warning_text || "-" }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="started_at" label="开始时间" min-width="168" />
            <el-table-column prop="finished_at" label="结束时间" min-width="168" />
          </el-table>
          <div class="pagination">
            <el-text type="info">
              第 {{ marketBackfillDetailPage }} 页，共 {{ marketBackfillDetailTotal }} 条
            </el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="marketBackfillDetailPage"
              :page-size="marketBackfillDetailPageSize"
              :total="marketBackfillDetailTotal"
              @current-change="handleMarketBackfillDetailPageChange"
            />
          </div>
        </div>
      </template>
    </el-drawer>

    <el-drawer v-model="runDetailVisible" size="820px" destroy-on-close @closed="handleRunDetailClosed">
      <template #header>
        <div class="drawer-title">任务运行详情</div>
      </template>
      <template v-if="currentRun">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="运行ID">{{ currentRun.id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="上级运行ID">{{ currentRun.parent_run_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="任务">
            {{ formatJobName(currentRun.job_name) }}
            <div class="muted">{{ currentRun.job_name || "-" }}</div>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(currentRun.status)">{{ formatRunStatus(currentRun.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="触发来源">{{ formatTriggerSource(currentRun.trigger_source) }}</el-descriptions-item>
          <el-descriptions-item label="重试次数">{{ currentRun.retry_count ?? "-" }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ currentRun.started_at || "-" }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{ currentRun.finished_at || "-" }}</el-descriptions-item>
          <el-descriptions-item label="执行耗时">
            {{ formatRunDuration(currentRun.started_at, currentRun.finished_at, currentRun.status) }}
          </el-descriptions-item>
          <el-descriptions-item label="操作人ID">{{ currentRun.operator_id || "-" }}</el-descriptions-item>
        </el-descriptions>

        <div v-if="isTushareRun(currentRun)" class="run-detail-block">
          <div class="run-detail-title-row">
            <h4>资讯同步明细</h4>
            <el-button
              size="small"
              :loading="newsSyncDetailsLoading"
              @click="fetchNewsSyncDetails"
            >
              刷新明细
            </el-button>
          </div>
          <div class="toolbar">
            <el-select
              v-model="newsSyncDetailFilters.sync_type"
              clearable
              placeholder="同步类型"
              style="width: 180px"
            >
              <el-option
                v-for="item in newsSyncTypeOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
            <el-input
              v-model="newsSyncDetailFilters.source"
              clearable
              placeholder="来源筛选"
              style="width: 160px"
            />
            <el-input
              v-model="newsSyncDetailFilters.symbol"
              clearable
              placeholder="股票代码"
              style="width: 150px"
            />
            <el-select
              v-model="newsSyncDetailFilters.status"
              clearable
              placeholder="状态"
              style="width: 130px"
            >
              <el-option
                v-for="status in newsSyncDetailStatusOptions"
                :key="status.value"
                :label="status.label"
                :value="status.value"
              />
            </el-select>
            <el-button type="primary" plain @click="applyNewsSyncDetailFilters">查询</el-button>
            <el-button @click="resetNewsSyncDetailFilters">重置</el-button>
          </div>
          <el-table
            :data="newsSyncDetails"
            border
            stripe
            size="small"
            v-loading="newsSyncDetailsLoading"
            empty-text="暂无同步明细"
          >
            <el-table-column label="同步类型" min-width="130">
              <template #default="{ row }">
                {{ syncTypeLabel(row.sync_type) }}
              </template>
            </el-table-column>
            <el-table-column prop="source" label="来源" min-width="120">
              <template #default="{ row }">
                {{ row.source || "-" }}
              </template>
            </el-table-column>
            <el-table-column prop="symbol" label="股票代码" min-width="100">
              <template #default="{ row }">
                {{ row.symbol || "-" }}
              </template>
            </el-table-column>
            <el-table-column prop="fetched_count" label="抓取" min-width="70" />
            <el-table-column prop="upserted_count" label="入库" min-width="70" />
            <el-table-column prop="failed_count" label="失败" min-width="70" />
            <el-table-column label="状态" min-width="90">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ formatRunStatus(row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="告警/错误" min-width="210">
              <template #default="{ row }">
                <span class="run-preview">{{ row.error_text || row.warning_text || "-" }}</span>
              </template>
            </el-table-column>
            <el-table-column label="开始时间" min-width="168">
              <template #default="{ row }">
                {{ row.started_at || "-" }}
              </template>
            </el-table-column>
            <el-table-column label="结束时间" min-width="168">
              <template #default="{ row }">
                {{ row.finished_at || "-" }}
              </template>
            </el-table-column>
            <el-table-column label="操作" align="right" min-width="120">
              <template #default="{ row }">
                <el-button
                  v-if="canEditSystemJobs"
                  size="small"
                  type="danger"
                  plain
                  :disabled="!canRetryNewsSyncRow(row)"
                  :loading="retryNewsSyncItemLoading"
                  @click="retryNewsSyncDetailItem(row)"
                >
                  重试失败项
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          <div class="pagination">
            <el-text type="info">第 {{ newsSyncDetailPage }} 页，共 {{ newsSyncDetailTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="newsSyncDetailPage"
              :page-size="newsSyncDetailPageSize"
              :total="newsSyncDetailTotal"
              @current-change="handleNewsSyncDetailPageChange"
            />
          </div>
        </div>

        <div v-if="isFuturesRun(currentRun)" class="run-detail-block">
          <div class="run-detail-title-row">
            <h4>期货策略明细</h4>
          </div>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="交易日">
              {{ runSummaryText(currentRun, "trade_date") }}
            </el-descriptions-item>
            <el-descriptions-item v-if="isFuturesGenerateRun(currentRun)" label="生成策略数">
              {{ runSummaryText(currentRun, "generated") }}
            </el-descriptions-item>
            <template v-if="isFuturesEvaluateRun(currentRun)">
              <el-descriptions-item label="评估总数">
                {{ runSummaryText(currentRun, "evaluated") }}
              </el-descriptions-item>
              <el-descriptions-item label="成功/失败">
                {{ runSummaryText(currentRun, "success") }}/{{ runSummaryText(currentRun, "failed") }}
              </el-descriptions-item>
              <el-descriptions-item label="平均评分">
                {{ runSummaryNumber(currentRun, "avg_score", 2) }}
              </el-descriptions-item>
              <el-descriptions-item label="平均胜率">
                {{ runSummaryPercent(currentRun, "avg_win_rate", 2) }}
              </el-descriptions-item>
              <el-descriptions-item label="平均超额收益">
                {{ runSummaryPercent(currentRun, "avg_excess", 2) }}
              </el-descriptions-item>
              <el-descriptions-item label="最大回撤">
                {{ runSummaryPercent(currentRun, "max_drawdown", 2) }}
              </el-descriptions-item>
              <el-descriptions-item label="真实基准覆盖">
                {{ runSummaryText(currentRun, "benchmark_actual") }}
              </el-descriptions-item>
              <el-descriptions-item label="最高分策略">
                {{ runSummaryText(currentRun, "top_id") }} / {{ runSummaryNumber(currentRun, "top_score", 2) }}
              </el-descriptions-item>
            </template>
          </el-descriptions>
        </div>

        <div class="run-detail-block">
          <div class="run-detail-title-row">
            <h4>运行摘要原文</h4>
            <el-button
              link
              type="primary"
              :disabled="!(currentRun.result_summary || '').trim()"
              :loading="copyingRunText"
              @click="copyRunField('运行摘要', currentRun.result_summary)"
            >
              复制
            </el-button>
          </div>
          <pre class="run-detail-pre">{{ currentRun.result_summary || "-" }}</pre>
        </div>

        <div class="run-detail-block">
          <div class="run-detail-title-row">
            <h4>错误信息原文</h4>
            <el-button
              link
              type="primary"
              :disabled="!(currentRun.error_message || '').trim()"
              :loading="copyingRunText"
              @click="copyRunField('错误信息', currentRun.error_message)"
            >
              复制
            </el-button>
          </div>
          <pre class="run-detail-pre">{{ currentRun.error_message || "-" }}</pre>
        </div>
      </template>
    </el-drawer>

    <el-dialog
      v-model="definitionFormVisible"
      :title="definitionFormMode === 'create' ? '新增任务定义' : `编辑任务定义：${definitionForm.id}`"
      width="760px"
      destroy-on-close
    >
      <el-form label-width="110px">
        <div class="dialog-grid">
          <el-form-item label="任务编码" required>
            <el-select
              v-model="definitionForm.job_name"
              filterable
              allow-create
              default-first-option
              reserve-keyword
              clearable
              :loading="supportedJobsLoading"
              :disabled="definitionFormMode === 'edit'"
              placeholder="请选择系统支持的任务，或输入自定义任务编码"
            >
              <el-option
                v-for="item in definitionFormMode === 'create' ? definitionCreateJobOptions : definitionJobOptions"
                :key="item.job_name"
                :label="`${item.display_name} [${formatModule(item.module)}] (${item.job_name})${item.alias_of ? ` · 别名→${item.alias_of}` : ''}${item.used ? ' · 已配置' : ''}`"
                :value="item.job_name"
              />
            </el-select>
          </el-form-item>
          <el-form-item v-if="definitionFormMode === 'create' && !definitionCreateJobOptions.length" label="任务编码提示">
            <el-text type="info">系统内置任务已全部配置，可直接输入自定义任务编码创建新定义。</el-text>
          </el-form-item>
          <el-form-item v-if="definitionFormMode === 'edit'" label="任务编码说明">
            <el-text type="info">编辑模式下任务编码不可修改，如需更换请删除后重新创建。</el-text>
          </el-form-item>
          <el-form-item label="任务名称" required>
            <el-input v-model="definitionForm.display_name" />
          </el-form-item>
          <el-form-item label="所属模块">
            <el-select v-model="definitionForm.module">
              <el-option v-for="item in moduleOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="调度方式" required>
            <el-select v-model="definitionSchedule.mode">
              <el-option
                v-for="item in scheduleModeOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item
            v-if="definitionSchedule.mode === 'DAILY' || definitionSchedule.mode === 'WEEKLY'"
            label="执行时间"
            required
          >
            <div class="inline-form-fields">
              <el-input-number
                v-model="definitionSchedule.hour"
                :min="0"
                :max="23"
                :step="1"
                controls-position="right"
                style="width: 120px"
              />
              <span class="muted">时</span>
              <el-input-number
                v-model="definitionSchedule.minute"
                :min="0"
                :max="59"
                :step="1"
                controls-position="right"
                style="width: 120px"
              />
              <span class="muted">分</span>
            </div>
          </el-form-item>
          <el-form-item v-if="definitionSchedule.mode === 'WEEKLY'" label="执行周期" required>
            <el-select v-model="definitionSchedule.week_day">
              <el-option
                v-for="item in scheduleWeekDayOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item v-if="definitionSchedule.mode === 'INTERVAL_MINUTES'" label="间隔分钟" required>
            <el-input-number
              v-model="definitionSchedule.interval_minutes"
              :min="1"
              :max="59"
              :step="1"
              controls-position="right"
              style="width: 180px"
            />
          </el-form-item>
          <el-form-item v-if="definitionSchedule.mode === 'CUSTOM'" label="调度表达式" required>
            <el-input
              v-model="definitionSchedule.custom_expr"
              placeholder="请输入 Cron 表达式，例如 0 0 9 * * *"
            />
          </el-form-item>
          <el-form-item label="表达式预览">
            <el-input :model-value="definitionCronPreview" readonly placeholder="自动生成表达式" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="definitionForm.status">
              <el-option v-for="item in definitionStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="definitionFormVisible = false">取消</el-button>
        <el-button
          v-if="canEditSystemJobs"
          type="primary"
          :loading="submittingDefinition"
          @click="submitDefinition"
        >
          {{ definitionFormMode === "create" ? "创建" : "更新" }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.jobs-hero {
  margin-bottom: 12px;
  border: 1px solid #dbeafe;
  background:
    radial-gradient(circle at top right, rgba(191, 219, 254, 0.9), transparent 32%),
    linear-gradient(135deg, #f8fbff 0%, #eef5ff 58%, #ffffff 100%);
}

.jobs-hero__top {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
}

.page-eyebrow {
  margin-bottom: 8px;
  color: #1d4ed8;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.08em;
}

.jobs-hero__desc {
  margin: 8px 0 0;
  max-width: 760px;
  color: #475569;
  font-size: 14px;
  line-height: 1.7;
}

.jobs-hero__top-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.jobs-action-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
  margin-top: 16px;
}

.jobs-action-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 156px;
  padding: 16px;
  border-radius: 16px;
  border: 1px solid #dbe5f3;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 14px 32px rgba(15, 23, 42, 0.06);
}

.jobs-action-card.is-primary {
  border-color: #bfdbfe;
  background: linear-gradient(180deg, #eff6ff 0%, #ffffff 100%);
}

.jobs-action-card.is-danger {
  border-color: #fecaca;
  background: linear-gradient(180deg, #fff4f4 0%, #ffffff 100%);
}

.jobs-action-card.is-gold {
  border-color: #fcd34d;
  background: linear-gradient(180deg, #fffbeb 0%, #ffffff 100%);
}

.jobs-action-card.is-info {
  border-color: #cbd5e1;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
}

.jobs-action-card__title {
  color: #0f172a;
  font-size: 16px;
  font-weight: 700;
}

.jobs-action-card__desc {
  color: #475569;
  font-size: 13px;
  line-height: 1.7;
}

.jobs-tabs-shell {
  margin-bottom: 12px;
  padding: 0;
  overflow: hidden;
}

.jobs-tab-summary {
  padding: 0 18px 18px;
}

.jobs-tab-summary__title {
  color: #0f172a;
  font-size: 18px;
  font-weight: 700;
}

.jobs-tab-summary__desc {
  margin-top: 6px;
  color: #64748b;
  font-size: 13px;
  line-height: 1.7;
}

.jobs-tab-stack {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.jobs-tabs :deep(.el-tabs__header) {
  margin: 0;
}

.jobs-tabs :deep(.el-tabs__content) {
  display: none;
}

.jobs-main-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.9fr) minmax(280px, 0.95fr);
  gap: 12px;
  align-items: start;
}

.jobs-side-column {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.jobs-panel-card {
  padding: 18px;
}

.section-header--stack {
  align-items: flex-start;
}

.section-copy {
  margin: 6px 0 0;
  color: #64748b;
  font-size: 13px;
  line-height: 1.6;
}

.jobs-overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}

.jobs-overview-card {
  min-height: 132px;
  padding: 14px;
  border-radius: 14px;
  border: 1px solid #dbe5f3;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
}

.jobs-overview-card.is-primary {
  border-color: #bfdbfe;
  background: linear-gradient(180deg, #eff6ff 0%, #ffffff 100%);
}

.jobs-overview-card.is-danger {
  border-color: #fecaca;
  background: linear-gradient(180deg, #fff5f5 0%, #ffffff 100%);
}

.jobs-overview-card.is-warning {
  border-color: #fde68a;
  background: linear-gradient(180deg, #fff8e6 0%, #ffffff 100%);
}

.jobs-overview-card.is-success {
  border-color: #bbf7d0;
  background: linear-gradient(180deg, #f0fdf4 0%, #ffffff 100%);
}

.jobs-overview-card.is-info {
  border-color: #cbd5e1;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
}

.jobs-overview-card.is-gold {
  border-color: #fcd34d;
  background: linear-gradient(180deg, #fffbeb 0%, #ffffff 100%);
}

.jobs-overview-card__title {
  color: #475569;
  font-size: 12px;
}

.jobs-overview-card__value {
  margin-top: 8px;
  color: #0f172a;
  font-size: 28px;
  font-weight: 700;
}

.jobs-overview-card__helper {
  margin-top: 10px;
  color: #64748b;
  font-size: 12px;
  line-height: 1.6;
}

.jobs-table-block + .jobs-table-block {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid #eef2f7;
}

.jobs-table-toolbar {
  align-items: center;
  margin-bottom: 8px;
}

.jobs-guide-panel {
  position: sticky;
  top: 20px;
}

.jobs-guide-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.jobs-guide-card {
  padding: 14px;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
}

.jobs-guide-card__title {
  color: #0f172a;
  font-size: 15px;
  font-weight: 700;
}

.jobs-guide-card__list {
  margin: 10px 0 0;
  padding-left: 18px;
  color: #475569;
  font-size: 13px;
  line-height: 1.75;
}

.jobs-guide-card__desc {
  margin: 10px 0 0;
  color: #475569;
  font-size: 13px;
  line-height: 1.75;
}

.jobs-workbench-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(420px, 1fr));
  gap: 12px;
  align-items: start;
}

.market-backfill-form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 0 12px;
}

.market-backfill-asset-summary {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.market-backfill-stage-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}

.market-backfill-stage-card {
  padding: 14px;
  border-radius: 14px;
  border: 1px solid #dbe5f3;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
}

.market-backfill-stage-card.is-success {
  border-color: #bbf7d0;
  background: linear-gradient(180deg, #f0fdf4 0%, #ffffff 100%);
}

.market-backfill-stage-card.is-warning {
  border-color: #fde68a;
  background: linear-gradient(180deg, #fff8e6 0%, #ffffff 100%);
}

.market-backfill-stage-card.is-danger {
  border-color: #fecaca;
  background: linear-gradient(180deg, #fff5f5 0%, #ffffff 100%);
}

.market-backfill-stage-card__top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.market-backfill-stage-card__title {
  color: #0f172a;
  font-size: 14px;
  font-weight: 700;
}

.market-backfill-stage-card__desc {
  margin-top: 10px;
  color: #475569;
  font-size: 12px;
  line-height: 1.65;
}

.market-backfill-empty {
  padding: 14px;
  border-radius: 14px;
  border: 1px dashed #cbd5e1;
  color: #64748b;
  background: #f8fafc;
}

.metric-item {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 12px;
  background: #fff;
}

.metric-label {
  color: #6b7280;
  font-size: 12px;
}

.metric-value {
  margin-top: 6px;
  font-size: 24px;
  font-weight: 700;
}

.metric-value--small {
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
  word-break: break-word;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.inline-actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
}

.inline-actions--left {
  justify-content: flex-start;
}

.inline-form-fields {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.run-preview {
  display: -webkit-box;
  overflow: hidden;
  text-overflow: ellipsis;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.run-summary-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.drawer-title {
  font-size: 16px;
  font-weight: 600;
}

.run-detail-block {
  margin-top: 14px;
}

.run-detail-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
}

.run-detail-title-row h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.run-detail-pre {
  margin: 0;
  border-radius: 10px;
  border: 1px solid #e5e7eb;
  background: #f8fafc;
  padding: 10px;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
  line-height: 1.5;
  max-height: 220px;
  overflow: auto;
}

.dialog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 0 12px;
}

:deep(.dialog-grid .el-form-item) {
  margin-bottom: 14px;
}

:deep(.dialog-grid .el-select),
:deep(.market-backfill-form-grid .el-select),
:deep(.market-backfill-form-grid .el-date-editor),
:deep(.market-backfill-form-grid .el-input-number) {
  width: 100%;
}

:deep(.market-backfill-switches .el-form-item__content) {
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
}

@media (max-width: 1280px) {
  .jobs-main-grid {
    grid-template-columns: 1fr;
  }

  .jobs-guide-panel {
    position: static;
  }
}

@media (max-width: 900px) {
  .jobs-hero__top {
    flex-direction: column;
  }

  .jobs-hero__top-actions {
    justify-content: flex-start;
  }

  .jobs-workbench-grid {
    grid-template-columns: 1fr;
  }

  .market-backfill-stage-grid {
    grid-template-columns: 1fr;
  }
}
</style>
