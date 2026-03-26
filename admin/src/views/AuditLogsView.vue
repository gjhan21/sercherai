<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { getAuditEventSummary, listAuditEvents, listOperationLogs } from "../api/admin";
import { getAccessToken } from "../lib/session";

const router = useRouter();
const activeTab = ref("audit-events");
const loading = ref(false);
const exportingFiltered = ref(false);
const copyingDetail = ref(false);
const auditFeatureUnavailable = ref(false);
const auditFeatureError = ref("");

const errorMessage = ref("");
const message = ref("");

const operationFilters = reactive({
  module: "",
  action: "",
  operator_user_id: ""
});

const auditFilters = reactive({
  event_domain: "",
  event_type: "",
  level: "",
  module: "",
  object_type: "",
  status: ""
});

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const operationItems = ref([]);
const auditItems = ref([]);
const auditSummary = ref(null);

const detailVisible = ref(false);
const currentRecord = ref(null);

const moduleOptions = ["USER", "NEWS", "WORKFLOW", "MEMBERSHIP", "SYSTEM", "STOCK", "FUTURES", "RISK", "STRATEGY_ENGINE"];
const auditDomainOptions = ["RESEARCH", "DATA", "PUBLISH", "SYSTEM"];
const auditLevelOptions = ["INFO", "WARNING", "CRITICAL"];
const auditStatusOptions = ["OPEN", "RESOLVED"];
const auditObjectTypeOptions = ["REVIEW_TASK", "SCHEDULER_JOB", "DATA_SOURCE", "STRATEGY_JOB", "STRATEGY_PUBLISH_POLICY", "STRATEGY_SEED_SET", "STRATEGY_AGENT_PROFILE", "STRATEGY_SCENARIO_TEMPLATE"];
const auditMigrationSQLPath = "backend/migrations/20260324_00_admin_audit_events.sql";
const auditMigrationCommand = "./scripts/devctl.sh migrate audit";

const currentItems = computed(() => (activeTab.value === "audit-events" ? auditItems.value : operationItems.value));

const operationPageSummary = computed(() => {
  const modules = new Set();
  const actions = new Set();
  let withReason = 0;
  operationItems.value.forEach((item) => {
    if (item.module) {
      modules.add(item.module);
    }
    if (item.action) {
      actions.add(item.action);
    }
    if ((item.reason || "").trim()) {
      withReason += 1;
    }
  });
  return {
    modules: modules.size,
    actions: actions.size,
    withReason,
    withoutReason: Math.max(0, operationItems.value.length - withReason)
  };
});

const auditPageSummary = computed(() => {
  const domains = new Set();
  const eventTypes = new Set();
  let warningCount = 0;
  let openCount = 0;
  auditItems.value.forEach((item) => {
    if (item.event_domain) {
      domains.add(item.event_domain);
    }
    if (item.event_type) {
      eventTypes.add(item.event_type);
    }
    if (item.level === "WARNING" || item.level === "CRITICAL") {
      warningCount += 1;
    }
    if (item.status === "OPEN") {
      openCount += 1;
    }
  });
  return {
    domains: domains.size,
    eventTypes: eventTypes.size,
    warningCount,
    openCount
  };
});

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function isAuditFeatureUnavailableError(error) {
  const text = normalizeErrorMessage(error, "").toLowerCase();
  return text.includes("admin_audit_events") || text.includes("doesn't exist") || text.includes("1146");
}

function markAuditFeatureUnavailable(error) {
  const rootError = normalizeErrorMessage(error, "");
  auditFeatureUnavailable.value = true;
  auditFeatureError.value =
    rootError
      ? `统一审计事件当前不可用：${rootError}。请先执行 ${auditMigrationSQLPath}`
      : `统一审计事件依赖的数据表不存在，请先执行 ${auditMigrationSQLPath}`;
}

function clearAuditFeatureUnavailable() {
  auditFeatureUnavailable.value = false;
  auditFeatureError.value = "";
}

function resolvedAuditCount(summary) {
  const totalCount = Number(summary?.total_count) || 0;
  const openCount = Number(summary?.open_count) || 0;
  return Math.max(0, totalCount - openCount);
}

function normalizeAuditRouteToken(value) {
  return String(value || "")
    .trim()
    .toUpperCase();
}

function normalizeStrategyConfigRouteType(value) {
  const normalized = normalizeAuditRouteToken(value);
  if (normalized === "STRATEGY_SEED_SET") return "seed-set";
  if (normalized === "STRATEGY_AGENT_PROFILE") return "agent-profile";
  if (normalized === "STRATEGY_SCENARIO_TEMPLATE") return "scenario-template";
  if (normalized === "STRATEGY_PUBLISH_POLICY") return "publish-policy";
  return "";
}

function buildStrategyConfigRoute(objectType, objectID, policyID = "") {
  const configType = normalizeStrategyConfigRouteType(objectType);
  if (!configType) {
    return "";
  }
  const targetID = String(configType === "publish-policy" ? policyID || objectID : objectID).trim();
  if (!targetID) {
    return {
      name: "market-center",
      query: { tab: "engine-config" }
    };
  }
  return {
    name: "market-center",
    query: {
      tab: "engine-config",
      config_type: configType,
      config_id: targetID,
      ...(configType === "publish-policy" ? { policy_id: targetID } : {})
    }
  };
}

function resolveAuditObjectRoute(item) {
  const objectType = normalizeAuditRouteToken(item?.object_type);
  const module = normalizeAuditRouteToken(item?.module);
  const metadata = item?.metadata || {};
  const objectID = String(item?.object_id || metadata?.review_id || metadata?.run_id || metadata?.source_key || "").trim();
  const jobName = String(metadata?.job_name || "").trim();
  const publishID = String(metadata?.publish_id || "").trim();
  const policyID = String(metadata?.policy_id || objectID).trim();
  const jobType = normalizeAuditRouteToken(metadata?.job_type);
  const marketTab = jobType.includes("FUTURES") ? "futures" : "stocks";
  if (objectType === "REVIEW_TASK" || module === "WORKFLOW") {
    return objectID ? { name: "review-center", query: { review_id: objectID } } : { name: "review-center" };
  }
  if (objectType === "SCHEDULER_JOB" || module === "SYSTEM") {
    if (objectID) {
      return jobName
        ? { name: "system-jobs", query: { run_id: objectID, job_name: jobName } }
        : { name: "system-jobs", query: { run_id: objectID } };
    }
    return { name: "system-jobs" };
  }
  if (objectType === "DATA_SOURCE" || module === "DATA") {
    return objectID
      ? { name: "data-sources-registry", query: { source_key: objectID, action: "logs" } }
      : { name: "data-sources-governance" };
  }
  const strategyConfigRoute = buildStrategyConfigRoute(objectType, objectID, policyID);
  if (strategyConfigRoute) {
    return strategyConfigRoute;
  }
  if (objectType === "STRATEGY_JOB") {
    if (publishID) {
      return {
        name: "market-center",
        query: {
          tab: marketTab,
          publish_id: publishID,
          view: "detail",
          job_type: jobType || undefined
        }
      };
    }
    return objectID
      ? { name: "system-jobs", query: { run_id: objectID, job_name: jobName || jobType } }
      : { name: "system-jobs" };
  }
  if (module === "STOCK") {
    return objectID
      ? { name: "stock-selection-runs", query: { run_id: objectID } }
      : { name: "stock-selection-runs" };
  }
  if (module === "FUTURES") {
    return objectID
      ? { name: "futures-selection-runs", query: { run_id: objectID } }
      : { name: "futures-selection-runs" };
  }
  return "";
}

function resolveOperationObjectRoute(item) {
  const targetType = normalizeAuditRouteToken(item?.target_type);
  const module = normalizeAuditRouteToken(item?.module);
  const targetID = String(item?.target_id || "").trim();

  if (targetType === "REVIEW_TASK" || module === "WORKFLOW") {
    return targetID ? { name: "review-center", query: { review_id: targetID } } : { name: "review-center" };
  }
  if (targetType === "DATA_SOURCE") {
    return targetID
      ? { name: "data-sources-registry", query: { source_key: targetID, action: "logs" } }
      : { name: "data-sources-governance" };
  }
  if (targetType === "JOB_RUN" || targetType === "SCHEDULER_JOB") {
    return targetID ? { name: "system-jobs", query: { run_id: targetID } } : { name: "system-jobs" };
  }
  if (targetType === "JOB" || targetType === "JOB_DEFINITION") {
    return targetID ? { name: "system-jobs", query: { job_name: targetID } } : { name: "system-jobs" };
  }
  const strategyConfigRoute = buildStrategyConfigRoute(targetType, targetID, targetID);
  if (strategyConfigRoute) {
    return strategyConfigRoute;
  }
  if (targetType === "STRATEGY_JOB") {
    return targetID ? { name: "system-jobs", query: { run_id: targetID } } : { name: "system-jobs" };
  }
  return "";
}

function openAuditObject(item) {
  const route = resolveAuditObjectRoute(item);
  if (!route) {
    errorMessage.value = "当前事件暂未配置对象跳转页";
    return;
  }
  router.push(route);
}

function openOperationObject(item) {
  const route = resolveOperationObjectRoute(item);
  if (!route) {
    errorMessage.value = "当前日志暂未配置对象跳转页";
    return;
  }
  router.push(route);
}

async function fetchOperationLogs(options = {}) {
  const { keepMessage = false } = options;
  loading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }

  try {
    const data = await listOperationLogs({
      module: operationFilters.module.trim(),
      action: operationFilters.action.trim(),
      operator_user_id: operationFilters.operator_user_id.trim(),
      page: page.value,
      page_size: pageSize.value
    });
    operationItems.value = data.items || [];
    total.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载操作日志失败");
  } finally {
    loading.value = false;
  }
}

async function fetchAuditSummary(options = {}) {
  const { keepMessage = true } = options;
  if (!keepMessage) {
    message.value = "";
  }
  try {
    auditSummary.value = await getAuditEventSummary();
    clearAuditFeatureUnavailable();
  } catch (error) {
    if (isAuditFeatureUnavailableError(error)) {
      auditSummary.value = null;
      markAuditFeatureUnavailable(error);
      return;
    }
    errorMessage.value = normalizeErrorMessage(error, "加载审计摘要失败");
  }
}

async function fetchAuditEvents(options = {}) {
  const { keepMessage = false, forceAuditProbe = false } = options;
  if (auditFeatureUnavailable.value && !forceAuditProbe) {
    activeTab.value = "operation-logs";
    message.value = "统一审计事件当前不可用，已切换到旧操作日志";
    await fetchOperationLogs({ keepMessage: true });
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }

  try {
    await fetchAuditSummary({ keepMessage: true });
    if (auditFeatureUnavailable.value) {
      loading.value = false;
      activeTab.value = "operation-logs";
      message.value = "统一审计事件当前不可用，已切换到旧操作日志";
      await fetchOperationLogs({ keepMessage: true });
      return;
    }
    const data = await listAuditEvents({
      event_domain: auditFilters.event_domain.trim(),
      event_type: auditFilters.event_type.trim(),
      level: auditFilters.level.trim(),
      module: auditFilters.module.trim(),
      object_type: auditFilters.object_type.trim(),
      status: auditFilters.status.trim(),
      page: page.value,
      page_size: pageSize.value
    });
    auditItems.value = data.items || [];
    total.value = data.total || 0;
    clearAuditFeatureUnavailable();
  } catch (error) {
    if (isAuditFeatureUnavailableError(error)) {
      markAuditFeatureUnavailable(error);
      loading.value = false;
      activeTab.value = "operation-logs";
      message.value = "统一审计事件当前不可用，已切换到旧操作日志";
      await fetchOperationLogs({ keepMessage: true });
      return;
    }
    errorMessage.value = normalizeErrorMessage(error, "加载统一审计事件失败");
  } finally {
    loading.value = false;
  }
}

async function fetchCurrentTab(options = {}) {
  if (activeTab.value === "audit-events") {
    await fetchAuditEvents(options);
    return;
  }
  await fetchOperationLogs(options);
}

function refreshCurrentTab() {
  fetchCurrentTab({ forceAuditProbe: auditFeatureUnavailable.value });
}

function handleTabChange() {
  page.value = 1;
  if (activeTab.value === "audit-events" && auditFeatureUnavailable.value) {
    activeTab.value = "operation-logs";
    message.value = "统一审计事件当前不可用，请先执行数据库迁移";
    fetchOperationLogs({ keepMessage: true });
    return;
  }
  fetchCurrentTab();
}

function retryAuditFeature() {
  page.value = 1;
  activeTab.value = "audit-events";
  fetchAuditEvents({ keepMessage: true, forceAuditProbe: true });
}

function applyFilters() {
  page.value = 1;
  fetchCurrentTab();
}

function resetFilters() {
  if (activeTab.value === "audit-events") {
    auditFilters.event_domain = "";
    auditFilters.event_type = "";
    auditFilters.level = "";
    auditFilters.module = "";
    auditFilters.object_type = "";
    auditFilters.status = "";
  } else {
    operationFilters.module = "";
    operationFilters.action = "";
    operationFilters.operator_user_id = "";
  }
  page.value = 1;
  fetchCurrentTab();
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  fetchCurrentTab();
}

function csvEscape(value) {
  const text = String(value ?? "");
  if (/[",\n]/.test(text)) {
    return `"${text.replace(/"/g, '""')}"`;
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

function buildOperationCSVRows(logs) {
  const header = [
    "id",
    "module",
    "action",
    "target_type",
    "target_id",
    "operator_user_id",
    "before_value",
    "after_value",
    "reason",
    "created_at"
  ];
  const rows = logs.map((item) => [
    item.id || "",
    item.module || "",
    item.action || "",
    item.target_type || "",
    item.target_id || "",
    item.operator_user_id || "",
    item.before_value || "",
    item.after_value || "",
    item.reason || "",
    item.created_at || ""
  ]);
  return [header, ...rows].map((row) => row.map(csvEscape).join(",")).join("\n");
}

function exportCurrentPageCSV() {
  if (activeTab.value !== "operation-logs") {
    return;
  }
  const csv = buildOperationCSVRows(operationItems.value);
  const fileName = `admin_operation_logs_page_${new Date().toISOString().slice(0, 10)}.csv`;
  triggerCSVDownload(csv, fileName);
  message.value = `已导出当前页 CSV，共 ${operationItems.value.length} 条`;
}

async function exportFilteredCSV() {
  if (activeTab.value !== "operation-logs") {
    return;
  }
  exportingFiltered.value = true;
  errorMessage.value = "";
  message.value = "";

  try {
    const params = new URLSearchParams();
    if (operationFilters.module.trim()) params.set("module", operationFilters.module.trim());
    if (operationFilters.action.trim()) params.set("action", operationFilters.action.trim());
    if (operationFilters.operator_user_id.trim()) params.set("operator_user_id", operationFilters.operator_user_id.trim());

    const baseURL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
    const query = params.toString();
    const requestURL = `${baseURL}/admin/audit/operation-logs/export.csv${query ? `?${query}` : ""}`;

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
    const fileName = `admin_operation_logs_filtered_${new Date().toISOString().slice(0, 10)}.csv`;
    const anchor = document.createElement("a");
    anchor.href = blobURL;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(blobURL);

    message.value = "已发起筛选结果 CSV 下载";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "导出筛选结果失败");
  } finally {
    exportingFiltered.value = false;
  }
}

function previewText(value, maxLength = 26) {
  const text = String(value || "");
  if (!text) {
    return "-";
  }
  if (text.length <= maxLength) {
    return text;
  }
  return `${text.slice(0, maxLength)}...`;
}

function openDetail(row) {
  currentRecord.value = row;
  detailVisible.value = true;
}

function goToMessageCenter() {
  router.push("/workflow-messages");
}

function formatJsonBlock(value) {
  if (!value || typeof value !== "object") {
    return "-";
  }
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return "-";
  }
}

async function copyDetailText(text, label) {
  const value = String(text || "").trim();
  if (!value) {
    errorMessage.value = `${label}为空，无法复制`;
    return;
  }
  copyingDetail.value = true;
  errorMessage.value = "";

  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(value);
    } else {
      const textarea = document.createElement("textarea");
      textarea.value = value;
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
    copyingDetail.value = false;
  }
}

onMounted(fetchCurrentTab);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">审计与操作日志</h1>
        <p class="muted">统一审计事件优先作为真相源，旧操作日志继续保留兼容导出能力</p>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <el-button @click="goToMessageCenter">返回消息中心</el-button>
        <el-button :loading="loading" @click="refreshCurrentTab">刷新</el-button>
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
      <el-tabs v-model="activeTab" @tab-change="handleTabChange">
        <el-tab-pane :label="auditFeatureUnavailable ? '统一审计事件（未启用）' : '统一审计事件'" name="audit-events" :disabled="auditFeatureUnavailable" />
        <el-tab-pane label="旧操作日志" name="operation-logs" />
      </el-tabs>
    </div>

    <el-alert
      v-if="auditFeatureUnavailable"
      :title="auditFeatureError"
      type="warning"
      :closable="false"
      show-icon
      style="margin-bottom: 12px"
    />
    <div v-if="auditFeatureUnavailable" class="card" style="margin-bottom: 12px">
      <div class="inline-actions">
        <el-text type="info">请先执行迁移命令后再点击“迁移后重试”</el-text>
        <el-button @click="copyDetailText(auditMigrationCommand, '迁移命令')">复制迁移命令</el-button>
        <el-button @click="copyDetailText(auditMigrationSQLPath, '迁移文件路径')">复制 SQL 路径</el-button>
        <el-button type="primary" :loading="loading" @click="retryAuditFeature">迁移后重试</el-button>
      </div>
    </div>

    <div v-if="activeTab === 'audit-events'" class="card" style="margin-bottom: 12px">
      <div class="grid grid-4">
        <div class="metric-item">
          <div class="metric-label">事件总数</div>
          <div class="metric-value">{{ auditSummary?.total_count ?? auditItems.length }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">开放事件</div>
          <div class="metric-value">{{ auditSummary?.open_count ?? auditPageSummary.openCount }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">告警等级</div>
          <div class="metric-value">{{ auditSummary?.warning_count ?? auditPageSummary.warningCount }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">已关闭事件</div>
          <div class="metric-value">{{ resolvedAuditCount(auditSummary) }}</div>
        </div>
      </div>
    </div>

    <div v-else class="card" style="margin-bottom: 12px">
      <div class="grid grid-4">
        <div class="metric-item">
          <div class="metric-label">当前页记录数</div>
          <div class="metric-value">{{ operationItems.length }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">涉及模块数</div>
          <div class="metric-value">{{ operationPageSummary.modules }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">涉及动作数</div>
          <div class="metric-value">{{ operationPageSummary.actions }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">有备注 / 无备注</div>
          <div class="metric-value">{{ operationPageSummary.withReason }} / {{ operationPageSummary.withoutReason }}</div>
        </div>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div v-if="activeTab === 'audit-events'" class="toolbar" style="margin-bottom: 0">
        <el-select v-model="auditFilters.event_domain" clearable placeholder="事件域" style="width: 150px">
          <el-option v-for="item in auditDomainOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-input v-model="auditFilters.event_type" clearable placeholder="事件类型" style="width: 220px" />
        <el-select v-model="auditFilters.level" clearable placeholder="告警等级" style="width: 150px">
          <el-option v-for="item in auditLevelOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-select v-model="auditFilters.module" clearable filterable allow-create default-first-option placeholder="模块" style="width: 160px">
          <el-option v-for="item in moduleOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-select v-model="auditFilters.object_type" clearable filterable allow-create default-first-option placeholder="对象类型" style="width: 170px">
          <el-option v-for="item in auditObjectTypeOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-select v-model="auditFilters.status" clearable placeholder="状态" style="width: 150px">
          <el-option v-for="item in auditStatusOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
      <div v-else class="toolbar" style="margin-bottom: 0">
        <el-select
          v-model="operationFilters.module"
          clearable
          filterable
          allow-create
          default-first-option
          placeholder="模块（如 USER）"
          style="width: 180px"
        >
          <el-option v-for="item in moduleOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-input v-model="operationFilters.action" clearable placeholder="动作（如 UPDATE_STATUS）" style="width: 200px" />
        <el-input
          v-model="operationFilters.operator_user_id"
          clearable
          placeholder="操作人ID（可选）"
          style="width: 200px"
        />
        <el-button :loading="exportingFiltered" @click="exportFilteredCSV">导出筛选CSV</el-button>
        <el-button @click="exportCurrentPageCSV">导出当前页CSV</el-button>
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </div>

    <div class="card">
      <el-table v-if="activeTab === 'audit-events'" :data="auditItems" border stripe v-loading="loading" empty-text="暂无统一审计事件">
        <el-table-column prop="id" label="事件ID" min-width="150" />
        <el-table-column prop="event_domain" label="事件域" min-width="110" />
        <el-table-column prop="event_type" label="事件类型" min-width="200" />
        <el-table-column prop="level" label="告警等级" min-width="110" />
        <el-table-column prop="module" label="模块" min-width="120" />
        <el-table-column prop="object_type" label="对象类型" min-width="130" />
        <el-table-column prop="object_id" label="对象ID" min-width="150" />
        <el-table-column prop="status" label="状态" min-width="100" />
        <el-table-column label="摘要" min-width="220">
          <template #default="{ row }">
            <div class="cell-preview">{{ previewText(row.summary || row.title, 40) }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" min-width="180" />
        <el-table-column label="操作" align="right" min-width="100">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-button size="small" @click="openDetail(row)">详情</el-button>
              <el-button size="small" :disabled="!resolveAuditObjectRoute(row)" @click="openAuditObject(row)">跳转对象页</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <el-table v-else :data="operationItems" border stripe v-loading="loading" empty-text="暂无操作日志">
        <el-table-column prop="id" label="日志ID" min-width="140" />
        <el-table-column prop="module" label="模块" min-width="100" />
        <el-table-column prop="action" label="动作" min-width="160" />
        <el-table-column prop="target_type" label="目标类型" min-width="110" />
        <el-table-column prop="target_id" label="目标ID" min-width="140" />
        <el-table-column prop="operator_user_id" label="操作人" min-width="130" />
        <el-table-column label="变更前" min-width="170">
          <template #default="{ row }">
            <div class="cell-preview">{{ previewText(row.before_value) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="变更后" min-width="170">
          <template #default="{ row }">
            <div class="cell-preview">{{ previewText(row.after_value) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="备注" min-width="170">
          <template #default="{ row }">
            <div class="cell-preview">{{ previewText(row.reason) }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" min-width="180" />
        <el-table-column label="操作" align="right" min-width="100">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-button size="small" @click="openDetail(row)">详情</el-button>
              <el-button size="small" :disabled="!resolveOperationObjectRoute(row)" @click="openOperationObject(row)">跳转对象页</el-button>
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

    <el-drawer v-model="detailVisible" size="620px" destroy-on-close>
      <template #header>
        <div class="drawer-title">{{ activeTab === 'audit-events' ? '审计事件详情' : '操作日志详情' }}</div>
      </template>

      <template v-if="currentRecord">
        <el-descriptions v-if="activeTab === 'audit-events'" :column="1" border size="small">
          <el-descriptions-item label="事件ID">{{ currentRecord.id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="事件域">{{ currentRecord.event_domain || '-' }}</el-descriptions-item>
          <el-descriptions-item label="事件类型">{{ currentRecord.event_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="告警等级">{{ currentRecord.level || '-' }}</el-descriptions-item>
          <el-descriptions-item label="模块">{{ currentRecord.module || '-' }}</el-descriptions-item>
          <el-descriptions-item label="对象类型">{{ currentRecord.object_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="对象ID">{{ currentRecord.object_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="操作人">{{ currentRecord.actor_user_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ currentRecord.status || '-' }}</el-descriptions-item>
          <el-descriptions-item label="时间">{{ currentRecord.created_at || '-' }}</el-descriptions-item>
        </el-descriptions>

        <el-descriptions v-else :column="1" border size="small">
          <el-descriptions-item label="日志ID">{{ currentRecord.id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="模块">{{ currentRecord.module || '-' }}</el-descriptions-item>
          <el-descriptions-item label="动作">{{ currentRecord.action || '-' }}</el-descriptions-item>
          <el-descriptions-item label="目标类型">{{ currentRecord.target_type || '-' }}</el-descriptions-item>
          <el-descriptions-item label="目标ID">{{ currentRecord.target_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="操作人">{{ currentRecord.operator_user_id || '-' }}</el-descriptions-item>
          <el-descriptions-item label="时间">{{ currentRecord.created_at || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="detail-block-wrap">
          <div class="detail-title-row">
            <h4>{{ activeTab === 'audit-events' ? '摘要' : '对象跳转' }}</h4>
            <el-button
              v-if="activeTab === 'audit-events'"
              link
              type="primary"
              :disabled="!(currentRecord.summary || currentRecord.title || '').trim()"
              :loading="copyingDetail"
              @click="copyDetailText(currentRecord.summary || currentRecord.title, '摘要')"
            >
              复制
            </el-button>
            <el-button
              link
              type="primary"
              :disabled="activeTab === 'audit-events' ? !resolveAuditObjectRoute(currentRecord) : !resolveOperationObjectRoute(currentRecord)"
              @click="activeTab === 'audit-events' ? openAuditObject(currentRecord) : openOperationObject(currentRecord)"
            >
              跳转对象页
            </el-button>
          </div>
          <pre class="detail-block">{{ activeTab === 'audit-events' ? currentRecord.summary || currentRecord.title || '-' : '当前日志可跳转到对应对象页时，会复用统一消息中心的深链规则。' }}</pre>
        </div>

        <div class="detail-block-wrap">
          <div class="detail-title-row">
            <h4>{{ activeTab === 'audit-events' ? '详情' : '变更前' }}</h4>
            <el-button
              link
              type="primary"
              :disabled="!(activeTab === 'audit-events' ? currentRecord.detail : currentRecord.before_value || '').trim()"
              :loading="copyingDetail"
              @click="copyDetailText(activeTab === 'audit-events' ? currentRecord.detail : currentRecord.before_value, activeTab === 'audit-events' ? '详情' : '变更前内容')"
            >
              复制
            </el-button>
          </div>
          <pre class="detail-block">{{ activeTab === 'audit-events' ? currentRecord.detail || '-' : currentRecord.before_value || '-' }}</pre>
        </div>

        <div v-if="activeTab === 'audit-events'" class="detail-block-wrap">
          <div class="detail-title-row">
            <h4>metadata</h4>
            <el-button
              link
              type="primary"
              :disabled="formatJsonBlock(currentRecord.metadata) === '-'"
              :loading="copyingDetail"
              @click="copyDetailText(formatJsonBlock(currentRecord.metadata), 'metadata')"
            >
              复制
            </el-button>
          </div>
          <pre class="detail-block">{{ formatJsonBlock(currentRecord.metadata) }}</pre>
        </div>

        <template v-else>
          <div class="detail-block-wrap">
            <div class="detail-title-row">
              <h4>变更后</h4>
              <el-button
                link
                type="primary"
                :disabled="!(currentRecord.after_value || '').trim()"
                :loading="copyingDetail"
                @click="copyDetailText(currentRecord.after_value, '变更后内容')"
              >
                复制
              </el-button>
            </div>
            <pre class="detail-block">{{ currentRecord.after_value || '-' }}</pre>
          </div>

          <div class="detail-block-wrap">
            <div class="detail-title-row">
              <h4>备注</h4>
              <el-button
                link
                type="primary"
                :disabled="!(currentRecord.reason || '').trim()"
                :loading="copyingDetail"
                @click="copyDetailText(currentRecord.reason, '备注')"
              >
                复制
              </el-button>
            </div>
            <pre class="detail-block">{{ currentRecord.reason || '-' }}</pre>
          </div>
        </template>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.metric-item {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 12px;
  background: #fff;
}

.inline-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
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

.cell-preview {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.drawer-title {
  font-size: 16px;
  font-weight: 600;
}

.detail-block-wrap {
  margin-top: 14px;
}

.detail-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
}

.detail-title-row h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.detail-block {
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
</style>
