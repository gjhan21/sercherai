<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from "vue";
import {
  bulkReadWorkflowMessages,
  countUnreadWorkflowMessages,
  listWorkflowMessages,
  updateWorkflowMessageRead
} from "../api/admin";
import { getAccessToken } from "../lib/session";

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

const items = ref([]);
const tableRef = ref(null);
const selectedRows = ref([]);

const autoRefreshEnabled = ref(true);
const timerRef = ref(null);

const detailVisible = ref(false);
const currentMessage = ref(null);

const filters = reactive({
  module: "",
  event_type: "",
  is_read: ""
});

const unreadInPage = computed(() => items.value.filter((item) => !item.is_read).length);
const readInPage = computed(() => Math.max(0, items.value.length - unreadInPage.value));
const selectedUnreadCount = computed(() => selectedRows.value.filter((item) => !item.is_read).length);
const selectedReadCount = computed(() => selectedRows.value.filter((item) => item.is_read).length);
const canBatchMarkRead = computed(() => selectedUnreadCount.value > 0);
const canBatchMarkUnread = computed(() => selectedReadCount.value > 0);

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

async function refreshAll(options = {}) {
  await fetchMessages(options);
}

async function toggleRead(item, targetRead = null) {
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
  currentMessage.value = item;
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
  fetchMessages();
}

function applyFilters() {
  page.value = 1;
  fetchMessages();
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  fetchMessages();
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
          流程消息
          <el-badge :value="unreadCount" class="title-badge" />
        </h1>
        <p class="muted">查看审核/任务/数据源告警消息，并处理已读状态</p>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <el-switch v-model="autoRefreshEnabled" inline-prompt active-text="自动刷新" inactive-text="手动刷新" />
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

    <div class="card" style="margin-bottom: 12px">
      <div class="message-stats">
        <el-tag type="warning">当前页未读 {{ unreadInPage }}</el-tag>
        <el-tag type="success">当前页已读 {{ readInPage }}</el-tag>
        <el-tag type="info">全量未读 {{ unreadCount }}</el-tag>
        <el-text type="info">每 60 秒可自动刷新一次列表</el-text>
      </div>
    </div>

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

    <div class="card" style="margin-bottom: 12px">
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
        <el-table-column type="selection" width="52" reserve-selection />
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
              <el-button size="small" @click="toggleRead(row)">
                {{ row.is_read ? "取消已读" : "标记已读" }}
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

    <el-drawer v-model="detailVisible" size="620px" destroy-on-close>
      <template #header>
        <div class="drawer-title">流程消息详情</div>
      </template>

      <template v-if="currentMessage">
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
