<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { listOperationLogs } from "../api/admin";
import { getAccessToken } from "../lib/session";

const loading = ref(false);
const exportingFiltered = ref(false);
const copyingDetail = ref(false);

const errorMessage = ref("");
const message = ref("");

const filters = reactive({
  module: "",
  action: "",
  operator_user_id: ""
});

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const items = ref([]);

const detailVisible = ref(false);
const currentLog = ref(null);

const moduleOptions = ["USER", "NEWS", "WORKFLOW", "MEMBERSHIP", "SYSTEM", "STOCK", "FUTURES", "RISK"];

const pageSummary = computed(() => {
  const modules = new Set();
  const actions = new Set();
  let withReason = 0;
  items.value.forEach((item) => {
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
    withoutReason: Math.max(0, items.value.length - withReason)
  };
});

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

async function fetchLogs(options = {}) {
  const { keepMessage = false } = options;
  loading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }

  try {
    const data = await listOperationLogs({
      module: filters.module.trim(),
      action: filters.action.trim(),
      operator_user_id: filters.operator_user_id.trim(),
      page: page.value,
      page_size: pageSize.value
    });
    items.value = data.items || [];
    total.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载操作日志失败");
  } finally {
    loading.value = false;
  }
}

function applyFilters() {
  page.value = 1;
  fetchLogs();
}

function resetFilters() {
  filters.module = "";
  filters.action = "";
  filters.operator_user_id = "";
  page.value = 1;
  fetchLogs();
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  fetchLogs();
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

function buildCSVRows(logs) {
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
  const csv = buildCSVRows(items.value);
  const fileName = `admin_operation_logs_page_${new Date().toISOString().slice(0, 10)}.csv`;
  triggerCSVDownload(csv, fileName);
  message.value = `已导出当前页 CSV，共 ${items.value.length} 条`;
}

async function exportFilteredCSV() {
  exportingFiltered.value = true;
  errorMessage.value = "";
  message.value = "";

  try {
    const params = new URLSearchParams();
    if (filters.module.trim()) params.set("module", filters.module.trim());
    if (filters.action.trim()) params.set("action", filters.action.trim());
    if (filters.operator_user_id.trim()) params.set("operator_user_id", filters.operator_user_id.trim());

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
  currentLog.value = row;
  detailVisible.value = true;
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

onMounted(fetchLogs);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">操作日志</h1>
        <p class="muted">追踪后台关键变更，支持筛选和导出</p>
      </div>
      <el-button :loading="loading" @click="fetchLogs">刷新</el-button>
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
      <div class="grid grid-4">
        <div class="metric-item">
          <div class="metric-label">当前页记录数</div>
          <div class="metric-value">{{ items.length }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">涉及模块数</div>
          <div class="metric-value">{{ pageSummary.modules }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">涉及动作数</div>
          <div class="metric-value">{{ pageSummary.actions }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">有备注 / 无备注</div>
          <div class="metric-value">{{ pageSummary.withReason }} / {{ pageSummary.withoutReason }}</div>
        </div>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="toolbar" style="margin-bottom: 0">
        <el-select
          v-model="filters.module"
          clearable
          filterable
          allow-create
          default-first-option
          placeholder="模块（如 USER）"
          style="width: 180px"
        >
          <el-option v-for="item in moduleOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-input v-model="filters.action" clearable placeholder="动作（如 UPDATE_STATUS）" style="width: 200px" />
        <el-input
          v-model="filters.operator_user_id"
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
      <el-table :data="items" border stripe v-loading="loading" empty-text="暂无操作日志">
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
            <el-button size="small" @click="openDetail(row)">详情</el-button>
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
        <div class="drawer-title">操作日志详情</div>
      </template>

      <template v-if="currentLog">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="日志ID">{{ currentLog.id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="模块">{{ currentLog.module || "-" }}</el-descriptions-item>
          <el-descriptions-item label="动作">{{ currentLog.action || "-" }}</el-descriptions-item>
          <el-descriptions-item label="目标类型">{{ currentLog.target_type || "-" }}</el-descriptions-item>
          <el-descriptions-item label="目标ID">{{ currentLog.target_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="操作人">{{ currentLog.operator_user_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="时间">{{ currentLog.created_at || "-" }}</el-descriptions-item>
        </el-descriptions>

        <div class="detail-block-wrap">
          <div class="detail-title-row">
            <h4>变更前</h4>
            <el-button
              link
              type="primary"
              :disabled="!(currentLog.before_value || '').trim()"
              :loading="copyingDetail"
              @click="copyDetailText(currentLog.before_value, '变更前内容')"
            >
              复制
            </el-button>
          </div>
          <pre class="detail-block">{{ currentLog.before_value || "-" }}</pre>
        </div>

        <div class="detail-block-wrap">
          <div class="detail-title-row">
            <h4>变更后</h4>
            <el-button
              link
              type="primary"
              :disabled="!(currentLog.after_value || '').trim()"
              :loading="copyingDetail"
              @click="copyDetailText(currentLog.after_value, '变更后内容')"
            >
              复制
            </el-button>
          </div>
          <pre class="detail-block">{{ currentLog.after_value || "-" }}</pre>
        </div>

        <div class="detail-block-wrap">
          <div class="detail-title-row">
            <h4>备注</h4>
            <el-button
              link
              type="primary"
              :disabled="!(currentLog.reason || '').trim()"
              :loading="copyingDetail"
              @click="copyDetailText(currentLog.reason, '备注')"
            >
              复制
            </el-button>
          </div>
          <pre class="detail-block">{{ currentLog.reason || "-" }}</pre>
        </div>
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
