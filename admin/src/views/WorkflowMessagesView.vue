<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import {
  bulkReadWorkflowMessages,
  countUnreadWorkflowMessages,
  getAuditEventSummary,
  listAuditEvents,
  listWorkflowMessages,
  updateWorkflowMessageRead
} from "../api/admin";
import { getAccessToken, hasPermission } from "../lib/session";

const router = useRouter();

const loading = ref(false);
const batchUpdating = ref(false);
const exportingFiltered = ref(false);
const copyingContent = ref(false);

const errorMessage = ref("");
const message = ref("");

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const unreadCount = ref(0);
const auditSummary = ref(null);
const relatedAuditItems = ref([]);
const messageCenterTab = ref("workflow");

const items = ref([]);
const tableRef = ref(null);
const selectedRows = ref([]);

const autoRefreshEnabled = ref(true);
const timerRef = ref(null);

const detailVisible = ref(false);
const detailMode = ref("workflow");
const currentMessage = ref(null);
const currentAuditItem = ref(null);
const canEditWorkflow = hasPermission("workflow.edit");

const filters = reactive({
  module: "",
  event_type: "",
  is_read: ""
});
const auditFilters = reactive({
  event_domain: "",
  event_type: "",
  level: "",
  module: "",
  object_type: "",
  status: "OPEN"
});

const unreadInPage = computed(() => items.value.filter((item) => !item.is_read).length);
const readInPage = computed(() => Math.max(0, items.value.length - unreadInPage.value));
const selectedUnreadCount = computed(() => selectedRows.value.filter((item) => !item.is_read).length);
const selectedReadCount = computed(() => selectedRows.value.filter((item) => item.is_read).length);
const canBatchMarkRead = computed(() => selectedUnreadCount.value > 0);
const canBatchMarkUnread = computed(() => selectedReadCount.value > 0);
const relatedAuditOpenCount = computed(() => auditSummary.value?.open_count ?? relatedAuditItems.value.length);
const messageCenterTabOptions = [
  { label: "流程待办", value: "workflow" },
  { label: "开放事件", value: "audit" }
];
const auditDomainOptions = ["RESEARCH", "DATA", "PUBLISH", "SYSTEM"];
const auditLevelOptions = ["INFO", "WARNING", "CRITICAL"];
const auditStatusOptions = ["OPEN", "RESOLVED"];
const auditObjectTypeOptions = ["REVIEW_TASK", "SCHEDULER_JOB", "DATA_SOURCE", "STRATEGY_JOB", "STRATEGY_PUBLISH_POLICY"];
const detailTitle = computed(() => (detailMode.value === "audit" ? "审计事件详情" : "流程消息详情"));

function ensureCanEditWorkflow() {
  if (canEditWorkflow) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法修改流程消息已读状态";
  return false;
}

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function clearMessages() {
  errorMessage.value = "";
  message.value = "";
}

function clearSelection() {
  selectedRows.value = [];
  nextTick(() => {
    tableRef.value?.clearSelection();
  });
}

function handleSelectionChange(rows) {
  selectedRows.value = rows || [];
}

async function fetchUnreadCount() {
  try {
    const data = await countUnreadWorkflowMessages({
      module: filters.module,
      event_type: filters.event_type
    });
    unreadCount.value = data.unread_count || 0;
  } catch {
    unreadCount.value = 0;
  }
}

async function fetchMessages(options = {}) {
  const { keepMessage = false } = options;
  loading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listWorkflowMessages({
      module: filters.module,
      event_type: filters.event_type,
      is_read: filters.is_read,
      page: page.value,
      page_size: pageSize.value
    });
    items.value = data.items || [];
    total.value = data.total || 0;
    clearSelection();
    await fetchUnreadCount();
  } catch (error) {
    errorMessage.value = error.message || "加载流程消息失败";
  } finally {
    loading.value = false;
  }
}

async function fetchAuditSummary() {
  try {
    auditSummary.value = await getAuditEventSummary();
  } catch {
    auditSummary.value = null;
  }
}

async function fetchRelatedAuditEvents() {
  try {
    const data = await listAuditEvents({
      event_domain: auditFilters.event_domain,
      event_type: auditFilters.event_type,
      level: auditFilters.level,
      module: auditFilters.module,
      object_type: auditFilters.object_type,
      status: auditFilters.status,
      page: 1,
      page_size: 5
    });
    relatedAuditItems.value = data.items || [];
  } catch {
    relatedAuditItems.value = [];
  }
}

async function refreshAll(options = {}) {
  await Promise.all([fetchMessages(options), fetchAuditSummary(), fetchRelatedAuditEvents()]);
}

async function toggleRead(item, targetRead = null) {
  if (!ensureCanEditWorkflow()) {
    return;
  }
  const nextRead = targetRead === null ? !item.is_read : !!targetRead;
  if (nextRead === item.is_read) {
    return;
  }
  clearMessages();
  try {
    await updateWorkflowMessageRead(item.id, nextRead);
    await fetchMessages({ keepMessage: true });
    message.value = `消息 ${item.id} 已${nextRead ? "标记已读" : "标记未读"}`;
  } catch (error) {
    errorMessage.value = error.message || "更新状态失败";
  }
}

async function markAllRead() {
  if (!ensureCanEditWorkflow()) {
    return;
  }
  clearMessages();
  try {
    const result = await bulkReadWorkflowMessages({
      module: filters.module,
      event_type: filters.event_type
    });
    await fetchMessages({ keepMessage: true });
    message.value = `批量已读完成，影响 ${result.affected || 0} 条`;
  } catch (error) {
    errorMessage.value = error.message || "批量已读失败";
  }
}

async function batchMarkSelected(targetRead) {
  if (!ensureCanEditWorkflow()) {
    return;
  }
  if (selectedRows.value.length <= 0) {
    errorMessage.value = "请先勾选消息";
    return;
  }

  batchUpdating.value = true;
  clearMessages();

  let success = 0;
  let failed = 0;
  let skipped = 0;

  try {
    for (const row of selectedRows.value) {
      if (!!row.is_read === !!targetRead) {
        skipped += 1;
        continue;
      }
      try {
        await updateWorkflowMessageRead(row.id, !!targetRead);
        success += 1;
      } catch {
        failed += 1;
      }
    }
  } finally {
    batchUpdating.value = false;
  }

  await fetchMessages({ keepMessage: true });
  message.value = `批量标记${targetRead ? "已读" : "未读"}完成：成功 ${success}，失败 ${failed}，跳过 ${skipped}`;
}

function openMessageDetail(item) {
  detailMode.value = "workflow";
  currentMessage.value = item;
  currentAuditItem.value = null;
  detailVisible.value = true;
}

function openAuditDetail(item) {
  detailMode.value = "audit";
  currentAuditItem.value = item;
  currentMessage.value = null;
  detailVisible.value = true;
}

async function copyContent() {
  const text = (currentMessage.value?.content || "").trim();
  if (!text) {
    errorMessage.value = "消息内容为空，无法复制";
    return;
  }

  copyingContent.value = true;
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
    message.value = "消息内容已复制";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "复制消息内容失败");
  } finally {
    copyingContent.value = false;
  }
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
      ? { name: "data-sources", query: { source_key: objectID, action: "logs" } }
      : { name: "data-sources" };
  }
  if (objectType === "STRATEGY_PUBLISH_POLICY") {
    return policyID
      ? { name: "market-center", query: { tab: "engine-config", policy_id: policyID } }
      : { name: "market-center", query: { tab: "engine-config" } };
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

function openAuditObject(item) {
  const route = resolveAuditObjectRoute(item);
  if (!route) {
    errorMessage.value = "当前事件暂未配置对象跳转页";
    return;
  }
  router.push(route);
}

function goToAuditLogs() {
  router.push("/audit-logs");
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

function buildCSVRows(rows) {
  const header = [
    "id",
    "module",
    "event_type",
    "title",
    "content",
    "target_id",
    "receiver_id",
    "sender_id",
    "is_read",
    "created_at",
    "read_at"
  ];
  const lines = rows.map((item) => [
    item.id || "",
    item.module || "",
    item.event_type || "",
    item.title || "",
    item.content || "",
    item.target_id || "",
    item.receiver_id || "",
    item.sender_id || "",
    item.is_read ? "true" : "false",
    item.created_at || "",
    item.read_at || ""
  ]);
  return [header, ...lines].map((row) => row.map(csvEscape).join(",")).join("\n");
}

function exportCurrentPageCSV() {
  const csv = buildCSVRows(items.value);
  const fileName = `workflow_messages_page_${new Date().toISOString().slice(0, 10)}.csv`;
  triggerCSVDownload(csv, fileName);
  message.value = `已导出当前页 CSV，共 ${items.value.length} 条`;
}

async function exportFilteredCSV() {
  exportingFiltered.value = true;
  clearMessages();

  try {
    const params = new URLSearchParams();
    if (filters.module.trim()) params.set("module", filters.module.trim());
    if (filters.event_type.trim()) params.set("event_type", filters.event_type.trim());
    if (filters.is_read !== "") params.set("is_read", filters.is_read);

    const baseURL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
    const query = params.toString();
    const requestURL = `${baseURL}/admin/workflow/messages/export.csv${query ? `?${query}` : ""}`;

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
    const fileName = `workflow_messages_filtered_${new Date().toISOString().slice(0, 10)}.csv`;
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

function resetFilters() {
  filters.module = "";
  filters.event_type = "";
  filters.is_read = "";
  page.value = 1;
  refreshAll();
}

function applyFilters() {
  page.value = 1;
  refreshAll();
}

function applyAuditFilters() {
  fetchRelatedAuditEvents();
}

function resetAuditFilters() {
  auditFilters.event_domain = "";
  auditFilters.event_type = "";
  auditFilters.level = "";
  auditFilters.module = "";
  auditFilters.object_type = "";
  auditFilters.status = "OPEN";
  fetchRelatedAuditEvents();
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  refreshAll();
}

function auditLevelTagType(level) {
  switch ((level || "").toUpperCase()) {
    case "CRITICAL":
      return "danger";
    case "WARNING":
      return "warning";
    default:
      return "info";
  }
}

function clearTimers() {
  if (timerRef.value) {
    clearInterval(timerRef.value);
    timerRef.value = null;
  }
}

function setupTimers() {
  clearTimers();
  timerRef.value = setInterval(() => {
    if (!autoRefreshEnabled.value || loading.value || batchUpdating.value || exportingFiltered.value) {
      return;
    }
    refreshAll({ keepMessage: true });
  }, 60 * 1000);
}

onMounted(() => {
  refreshAll();
  setupTimers();
});

onBeforeUnmount(() => {
  clearTimers();
});
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">
          消息中心
          <el-badge :value="unreadCount" class="title-badge" />
        </h1>
        <p class="muted">把流程待办与统一审计事件收口到一个入口里查看和处理</p>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <el-switch v-model="autoRefreshEnabled" inline-prompt active-text="自动刷新" inactive-text="手动刷新" />
        <el-button @click="goToAuditLogs">前往审计日志</el-button>
        <el-button :loading="loading" @click="refreshAll">刷新</el-button>
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
    <el-alert
      v-if="!canEditWorkflow"
      title="当前账号为只读流程消息权限，可查看消息详情和导出结果，但不能修改已读状态。"
      type="info"
      :closable="false"
      show-icon
      style="margin-bottom: 12px"
    />

    <div class="card" style="margin-bottom: 12px">
      <div class="message-stats">
        <el-tag type="warning">当前页未读 {{ unreadInPage }}</el-tag>
        <el-tag type="success">当前页已读 {{ readInPage }}</el-tag>
        <el-tag type="info">全量未读 {{ unreadCount }}</el-tag>
        <el-text type="info">每 60 秒可自动刷新一次列表</el-text>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header" style="margin-bottom: 0">
        <h3 style="margin: 0">消息中心视图</h3>
        <el-segmented v-model="messageCenterTab" :options="messageCenterTabOptions" />
      </div>
    </div>

    <template v-if="messageCenterTab === 'audit'">
      <div class="card" style="margin-bottom: 12px">
        <div class="section-header">
          <h3 style="margin: 0">统一事件摘要</h3>
          <el-text type="info">流程消息开始回读统一审计事件主链</el-text>
        </div>
        <div class="metric-grid">
          <div class="metric-card">
            <div class="metric-label">开放事件</div>
            <div class="metric-value">{{ relatedAuditOpenCount }}</div>
          </div>
          <div class="metric-card">
            <div class="metric-label">Warning</div>
            <div class="metric-value">{{ auditSummary?.warning_count ?? 0 }}</div>
          </div>
          <div class="metric-card">
            <div class="metric-label">Critical</div>
            <div class="metric-value">{{ auditSummary?.critical_count ?? 0 }}</div>
          </div>
          <div class="metric-card">
            <div class="metric-label">已关闭事件</div>
            <div class="metric-value">{{ resolvedAuditCount(auditSummary) }}</div>
          </div>
        </div>
      </div>

      <div class="card" style="margin-bottom: 12px">
        <div class="toolbar" style="margin-bottom: 0">
          <el-select v-model="auditFilters.event_domain" clearable placeholder="事件域" style="width: 150px">
            <el-option v-for="item in auditDomainOptions" :key="item" :label="item" :value="item" />
          </el-select>
          <el-input v-model="auditFilters.event_type" clearable placeholder="事件类型" style="width: 220px" />
          <el-select v-model="auditFilters.level" clearable placeholder="告警等级" style="width: 150px">
            <el-option v-for="item in auditLevelOptions" :key="item" :label="item" :value="item" />
          </el-select>
          <el-input v-model="auditFilters.module" clearable placeholder="模块" style="width: 140px" />
          <el-select v-model="auditFilters.object_type" clearable filterable allow-create default-first-option placeholder="对象类型" style="width: 170px">
            <el-option v-for="item in auditObjectTypeOptions" :key="item" :label="item" :value="item" />
          </el-select>
          <el-select v-model="auditFilters.status" clearable placeholder="状态" style="width: 150px">
            <el-option v-for="item in auditStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
          <el-button type="primary" plain @click="applyAuditFilters">查询</el-button>
          <el-button @click="resetAuditFilters">重置</el-button>
        </div>
      </div>

      <div class="card" style="margin-bottom: 12px">
        <div class="section-header">
          <h3 style="margin: 0">待处理审计事件</h3>
          <div class="inline-actions">
            <el-text type="info">按当前模块/事件筛选回读最近 5 条 open audit events</el-text>
            <el-button link type="primary" @click="goToAuditLogs">前往审计日志</el-button>
          </div>
        </div>
        <div v-if="relatedAuditItems.length" class="audit-event-list">
          <div v-for="item in relatedAuditItems" :key="item.id" class="audit-event-item">
            <div class="audit-event-topline">
              <el-tag size="small" :type="auditLevelTagType(item.level)">{{ item.level || "INFO" }}</el-tag>
              <el-tag size="small" effect="plain">{{ item.event_domain || "SYSTEM" }}</el-tag>
              <el-text type="info">{{ item.module || "-" }}</el-text>
              <el-text type="info">{{ item.created_at || "-" }}</el-text>
            </div>
            <div class="audit-event-title">{{ item.title || "-" }}</div>
            <div class="audit-event-summary">{{ item.summary || item.detail || "-" }}</div>
            <div class="inline-actions" style="margin-top: 8px">
              <el-button size="small" @click="openAuditDetail(item)">详情</el-button>
              <el-button size="small" :disabled="!resolveAuditObjectRoute(item)" @click="openAuditObject(item)">跳转对象页</el-button>
            </div>
          </div>
        </div>
        <el-empty v-else description="当前筛选下暂无 open audit events" />
      </div>
    </template>

    <template v-else>
      <div class="card" style="margin-bottom: 12px">
        <div class="toolbar">
          <el-input v-model="filters.module" clearable placeholder="模块，如 SYSTEM / STOCK" style="width: 220px" />
          <el-input
            v-model="filters.event_type"
            clearable
            placeholder="事件，如 DATA_SOURCE_UNHEALTHY"
            style="width: 260px"
          />
          <el-select v-model="filters.is_read" clearable placeholder="全部状态" style="width: 130px">
            <el-option label="未读" value="false" />
            <el-option label="已读" value="true" />
          </el-select>
          <el-button :loading="exportingFiltered" @click="exportFilteredCSV">导出筛选CSV</el-button>
          <el-button @click="exportCurrentPageCSV">导出当前页CSV</el-button>
          <el-button type="primary" plain @click="applyFilters">查询</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
      </div>

      <div v-if="canEditWorkflow" class="card" style="margin-bottom: 12px">
        <div class="section-header">
          <h3 style="margin: 0">批量处理</h3>
          <el-text type="info">
            已选 {{ selectedRows.length }} 条，未读 {{ selectedUnreadCount }} 条，已读 {{ selectedReadCount }} 条
          </el-text>
        </div>
        <div class="toolbar" style="margin-bottom: 0">
          <el-button type="primary" :loading="batchUpdating" @click="markAllRead">筛选结果全部已读</el-button>
          <el-button :disabled="!canBatchMarkRead" :loading="batchUpdating" @click="batchMarkSelected(true)">
            选中标记已读
          </el-button>
          <el-button :disabled="!canBatchMarkUnread" :loading="batchUpdating" @click="batchMarkSelected(false)">
            选中标记未读
          </el-button>
          <el-button @click="clearSelection">清空勾选</el-button>
        </div>
      </div>

      <div class="card">
        <el-table
          ref="tableRef"
          :data="items"
          row-key="id"
          border
          stripe
          v-loading="loading"
          empty-text="暂无流程消息"
          @selection-change="handleSelectionChange"
        >
          <el-table-column v-if="canEditWorkflow" type="selection" width="52" reserve-selection />
          <el-table-column prop="event_type" label="事件" min-width="170" />
          <el-table-column prop="module" label="模块" min-width="100" />
          <el-table-column prop="title" label="标题" min-width="160" />
          <el-table-column label="内容" min-width="260">
            <template #default="{ row }">
              <span class="content-preview">{{ row.content || "-" }}</span>
            </template>
          </el-table-column>
          <el-table-column label="接收人" min-width="140">
            <template #default="{ row }">
              {{ row.receiver_id || "-" }}
            </template>
          </el-table-column>
          <el-table-column label="状态" min-width="100">
            <template #default="{ row }">
              <el-tag :type="row.is_read ? 'success' : 'warning'">
                {{ row.is_read ? "已读" : "未读" }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" min-width="180" />
          <el-table-column label="操作" min-width="210" align="right">
            <template #default="{ row }">
              <div class="inline-actions inline-actions--right">
                <el-button size="small" @click="openMessageDetail(row)">详情</el-button>
                <el-button v-if="canEditWorkflow" size="small" @click="toggleRead(row)">
                  {{ row.is_read ? "取消已读" : "标记已读" }}
                </el-button>
                <el-text v-else type="info">只读</el-text>
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
    </template>

    <el-drawer v-model="detailVisible" size="620px" destroy-on-close>
      <template #header>
        <div class="drawer-title">{{ detailTitle }}</div>
      </template>

      <template v-if="detailMode === 'workflow' && currentMessage">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="消息ID">{{ currentMessage.id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="模块">{{ currentMessage.module || "-" }}</el-descriptions-item>
          <el-descriptions-item label="事件">{{ currentMessage.event_type || "-" }}</el-descriptions-item>
          <el-descriptions-item label="目标ID">{{ currentMessage.target_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="接收人">{{ currentMessage.receiver_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="发送人">{{ currentMessage.sender_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="currentMessage.is_read ? 'success' : 'warning'">
              {{ currentMessage.is_read ? "已读" : "未读" }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ currentMessage.created_at || "-" }}</el-descriptions-item>
          <el-descriptions-item label="已读时间">{{ currentMessage.read_at || "-" }}</el-descriptions-item>
        </el-descriptions>

        <div class="detail-title-row">
          <h4>标题</h4>
        </div>
        <pre class="detail-block">{{ currentMessage.title || "-" }}</pre>

        <div class="detail-title-row">
          <h4>内容</h4>
          <el-button
            link
            type="primary"
            :disabled="!(currentMessage.content || '').trim()"
            :loading="copyingContent"
            @click="copyContent"
          >
            复制内容
          </el-button>
        </div>
        <pre class="detail-block">{{ currentMessage.content || "-" }}</pre>
      </template>

      <template v-else-if="currentAuditItem">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="事件ID">{{ currentAuditItem.id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="事件域">{{ currentAuditItem.event_domain || "-" }}</el-descriptions-item>
          <el-descriptions-item label="事件类型">{{ currentAuditItem.event_type || "-" }}</el-descriptions-item>
          <el-descriptions-item label="告警等级">{{ currentAuditItem.level || "-" }}</el-descriptions-item>
          <el-descriptions-item label="模块">{{ currentAuditItem.module || "-" }}</el-descriptions-item>
          <el-descriptions-item label="对象类型">{{ currentAuditItem.object_type || "-" }}</el-descriptions-item>
          <el-descriptions-item label="对象ID">{{ currentAuditItem.object_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ currentAuditItem.status || "-" }}</el-descriptions-item>
          <el-descriptions-item label="时间">{{ currentAuditItem.created_at || "-" }}</el-descriptions-item>
        </el-descriptions>

        <div class="detail-title-row">
          <h4>摘要</h4>
        </div>
        <pre class="detail-block">{{ currentAuditItem.summary || currentAuditItem.title || "-" }}</pre>

        <div class="detail-title-row">
          <h4>详情</h4>
          <el-button link type="primary" :disabled="!resolveAuditObjectRoute(currentAuditItem)" @click="openAuditObject(currentAuditItem)">
            跳转对象页
          </el-button>
        </div>
        <pre class="detail-block">{{ currentAuditItem.detail || "-" }}</pre>

        <div class="detail-title-row">
          <h4>metadata</h4>
        </div>
        <pre class="detail-block">{{ formatJsonBlock(currentAuditItem.metadata) }}</pre>
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.title-badge {
  margin-left: 8px;
  vertical-align: middle;
}

.message-stats {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 12px;
}

.metric-card {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  background: #f8fafc;
  padding: 12px;
}

.metric-label {
  color: #64748b;
  font-size: 12px;
  margin-bottom: 6px;
}

.metric-value {
  font-size: 24px;
  font-weight: 700;
  color: #0f172a;
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
  gap: 8px;
  flex-wrap: wrap;
}

.inline-actions--right {
  justify-content: flex-end;
}

.content-preview {
  display: -webkit-box;
  overflow: hidden;
  text-overflow: ellipsis;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.audit-event-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.audit-event-item {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 12px;
  background: #fff;
}

.audit-event-topline {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}

.audit-event-title {
  font-size: 14px;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 6px;
}

.audit-event-summary {
  color: #475569;
  font-size: 13px;
  line-height: 1.5;
}

.drawer-title {
  font-size: 16px;
  font-weight: 600;
}

.detail-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 14px;
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
