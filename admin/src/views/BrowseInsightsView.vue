<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import {
  createUserMessages,
  getBrowseHistorySummary,
  getBrowseHistoryTrend,
  listBrowseHistories,
  listBrowseUserSegments
} from "../api/admin";
import { getAccessToken, hasPermission } from "../lib/session";

const loading = ref(false);
const summaryLoading = ref(false);
const trendLoading = ref(false);
const segmentsLoading = ref(false);
const exporting = ref(false);

const sendingBatch = ref(false);
const sendingSingleID = ref("");
const sendingSegment = ref("");

const errorMessage = ref("");
const successMessage = ref("");
const sendFailures = ref([]);

const filters = reactive({
  user_id: "",
  content_type: "",
  keyword: ""
});

const rows = ref([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const selectedRows = ref([]);

const summary = ref({
  total_views: 0,
  unique_users: 0,
  news_views: 0,
  report_views: 0,
  journal_views: 0,
  today_views: 0,
  last_7d_views: 0
});

const segments = ref([]);
const trendPoints = ref([]);

const messageForm = reactive({
  type: "NEWS",
  title: "阅读提醒",
  content: "你关注的资讯有更新，建议及时查看并结合策略页调整执行计划。"
});

const contentTypeOptions = ["NEWS", "REPORT", "JOURNAL"];
const messageTypeOptions = ["NEWS", "SYSTEM", "STRATEGY", "ALERT"];
const canEditUsers = hasPermission("users.edit");

const selectedUserIDs = computed(() => {
  const ids = selectedRows.value.map((item) => item.user_id).filter(Boolean);
  return [...new Set(ids)];
});

const activeSegments = computed(() =>
  (segments.value || []).filter((item) => String(item.segment || "").toUpperCase() === "ACTIVE")
);
const silentSegments = computed(() =>
  (segments.value || []).filter((item) => String(item.segment || "").toUpperCase() === "SILENT")
);

const canSendBatch = computed(() => {
  return (
    selectedUserIDs.value.length > 0 &&
    messageForm.title.trim() !== "" &&
    messageForm.content.trim() !== "" &&
    messageForm.type.trim() !== ""
  );
});

const summaryCards = computed(() => [
  { label: "总阅读量", value: `${summary.value.total_views || 0}` },
  { label: "阅读用户数", value: `${summary.value.unique_users || 0}` },
  { label: "今日阅读", value: `${summary.value.today_views || 0}` },
  { label: "近7天阅读", value: `${summary.value.last_7d_views || 0}` },
  { label: "新闻阅读", value: `${summary.value.news_views || 0}` },
  { label: "研报阅读", value: `${summary.value.report_views || 0}` },
  { label: "期刊阅读", value: `${summary.value.journal_views || 0}` }
]);

const trendMax = computed(() => {
  const values = (trendPoints.value || []).map((item) => Number(item.total_views || 0));
  return values.length ? Math.max(...values, 1) : 1;
});

function mapContentType(contentType) {
  const value = String(contentType || "").toUpperCase();
  if (value === "NEWS") return "新闻";
  if (value === "REPORT") return "研报";
  if (value === "JOURNAL") return "期刊";
  return value || "-";
}

function mapMessageType(type) {
  const value = String(type || "").toUpperCase();
  if (value === "NEWS") return "资讯通知";
  if (value === "SYSTEM") return "系统通知";
  if (value === "STRATEGY") return "策略提醒";
  if (value === "ALERT") return "风险告警";
  return value || "-";
}

function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return value || "-";
  }
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

function applyReminderTemplate(kind) {
  if (kind === "ACTIVE") {
    messageForm.type = "STRATEGY";
    messageForm.title = "高活跃用户策略升级提醒";
    messageForm.content =
      "你近期阅读活跃度较高，策略中心已同步新的高优先级策略建议，建议及时查看并更新执行参数。";
    return;
  }
  if (kind === "SILENT") {
    messageForm.type = "NEWS";
    messageForm.title = "内容更新提醒";
    messageForm.content =
      "你关注的资讯和研报已有新增内容，建议回到资讯页查看最新解读，避免错过关键市场变化。";
  }
}

function handleSelectionChange(items) {
  selectedRows.value = items || [];
}

function ensureCanEditUsers() {
  if (canEditUsers) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法发送阅读提醒";
  return false;
}

async function fetchSummary() {
  summaryLoading.value = true;
  try {
    const data = await getBrowseHistorySummary();
    summary.value = {
      total_views: Number(data?.total_views || 0),
      unique_users: Number(data?.unique_users || 0),
      news_views: Number(data?.news_views || 0),
      report_views: Number(data?.report_views || 0),
      journal_views: Number(data?.journal_views || 0),
      today_views: Number(data?.today_views || 0),
      last_7d_views: Number(data?.last_7d_views || 0)
    };
  } catch (error) {
    errorMessage.value = error?.message || "加载阅读概览失败";
  } finally {
    summaryLoading.value = false;
  }
}

async function fetchTrend() {
  trendLoading.value = true;
  try {
    const data = await getBrowseHistoryTrend({ days: 7 });
    trendPoints.value = data?.items || [];
  } catch (error) {
    errorMessage.value = error?.message || "加载阅读趋势失败";
  } finally {
    trendLoading.value = false;
  }
}

async function fetchSegments() {
  segmentsLoading.value = true;
  try {
    const data = await listBrowseUserSegments({ limit: 10 });
    segments.value = data?.items || [];
  } catch (error) {
    errorMessage.value = error?.message || "加载用户分层失败";
  } finally {
    segmentsLoading.value = false;
  }
}

async function fetchRows() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const data = await listBrowseHistories({
      user_id: filters.user_id.trim(),
      content_type: filters.content_type,
      keyword: filters.keyword.trim(),
      page: page.value,
      page_size: pageSize.value
    });
    rows.value = data?.items || [];
    total.value = Number(data?.total || 0);
  } catch (error) {
    errorMessage.value = error?.message || "加载阅读记录失败";
  } finally {
    loading.value = false;
  }
}

async function refreshAll() {
  successMessage.value = "";
  await Promise.all([fetchSummary(), fetchTrend(), fetchSegments(), fetchRows()]);
}

function applyFilters() {
  page.value = 1;
  fetchRows();
}

function resetFilters() {
  filters.user_id = "";
  filters.content_type = "";
  filters.keyword = "";
  page.value = 1;
  fetchRows();
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  fetchRows();
}

async function sendMessage(userIDs, doneMessage) {
  const data = await createUserMessages({
    user_ids: userIDs,
    title: messageForm.title.trim(),
    content: messageForm.content.trim(),
    type: messageForm.type.trim().toUpperCase()
  });
  const sentCount = Number(data?.sent_count || 0);
  const failedCount = Number(data?.failed_count || 0);
  const failures = Array.isArray(data?.failures) ? data.failures : [];
  sendFailures.value = failures;
  if (failedCount > 0) {
    successMessage.value = `${doneMessage}：成功 ${sentCount}，失败 ${failedCount}`;
    return;
  }
  successMessage.value = `${doneMessage}：成功 ${sentCount}`;
}

async function handleSendBatch() {
  if (!ensureCanEditUsers()) {
    return;
  }
  if (!canSendBatch.value || sendingBatch.value) {
    return;
  }
  sendingBatch.value = true;
  errorMessage.value = "";
  successMessage.value = "";
  sendFailures.value = [];
  try {
    await sendMessage(selectedUserIDs.value, `批量发送阅读提醒`);
  } catch (error) {
    errorMessage.value = error?.message || "批量发送提醒失败";
  } finally {
    sendingBatch.value = false;
  }
}

async function handleSendSingle(row) {
  if (!ensureCanEditUsers()) {
    return;
  }
  const userID = row?.user_id;
  if (!userID || sendingSingleID.value) {
    return;
  }
  sendingSingleID.value = row.id;
  errorMessage.value = "";
  successMessage.value = "";
  sendFailures.value = [];
  try {
    await sendMessage([userID], `向用户 ${userID} 发送阅读提醒`);
  } catch (error) {
    errorMessage.value = error?.message || "发送提醒失败";
  } finally {
    sendingSingleID.value = "";
  }
}

async function handleSendSegment(segmentType) {
  if (!ensureCanEditUsers()) {
    return;
  }
  const normalized = String(segmentType || "").toUpperCase();
  const users =
    normalized === "ACTIVE"
      ? activeSegments.value.map((item) => item.user_id).filter(Boolean)
      : silentSegments.value.map((item) => item.user_id).filter(Boolean);
  const targetUserIDs = [...new Set(users)];
  if (targetUserIDs.length === 0 || sendingSegment.value) {
    return;
  }
  sendingSegment.value = normalized;
  errorMessage.value = "";
  successMessage.value = "";
  sendFailures.value = [];
  try {
    await sendMessage(targetUserIDs, `向${normalized === "ACTIVE" ? "高活跃" : "沉默"}用户发送提醒`);
  } catch (error) {
    errorMessage.value = error?.message || "分层提醒发送失败";
  } finally {
    sendingSegment.value = "";
  }
}

async function exportFilteredCSV() {
  exporting.value = true;
  errorMessage.value = "";
  successMessage.value = "";
  try {
    const params = new URLSearchParams();
    if (filters.user_id.trim()) params.set("user_id", filters.user_id.trim());
    if (filters.content_type.trim()) params.set("content_type", filters.content_type.trim());
    if (filters.keyword.trim()) params.set("keyword", filters.keyword.trim());

    const baseURL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
    const query = params.toString();
    const requestURL = `${baseURL}/admin/users/browse-histories/export.csv${query ? `?${query}` : ""}`;

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
    const fileName = `browse_histories_${new Date().toISOString().slice(0, 10)}.csv`;
    const anchor = document.createElement("a");
    anchor.href = blobURL;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(blobURL);
    successMessage.value = "已导出阅读明细 CSV";
  } catch (error) {
    errorMessage.value = error?.message || "导出阅读明细失败";
  } finally {
    exporting.value = false;
  }
}

function trendBarHeight(value) {
  const max = Number(trendMax.value || 1);
  const current = Number(value || 0);
  if (current <= 0) {
    return 4;
  }
  return Math.max(4, Math.round((current / max) * 80));
}

onMounted(refreshAll);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">阅读分析</h1>
        <p class="muted">查看客户端阅读行为并对目标用户发送提醒</p>
      </div>
      <el-button :loading="loading || summaryLoading || segmentsLoading" @click="refreshAll">刷新</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />
    <el-alert
      v-if="successMessage"
      :title="successMessage"
      type="success"
      show-icon
      style="margin-bottom: 12px"
    />

    <div class="grid grid-4" style="margin-bottom: 12px">
      <article v-for="item in summaryCards" :key="item.label" class="metric-card">
        <p class="label">{{ item.label }}</p>
        <p class="value">{{ item.value }}</p>
      </article>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-head">
        <h3>近7天阅读趋势</h3>
        <el-text type="info">总量峰值 {{ trendMax }}</el-text>
      </div>
      <div v-if="trendLoading" class="empty-text">正在加载趋势数据...</div>
      <div v-else class="trend-grid">
        <article v-for="item in trendPoints" :key="item.date" class="trend-item">
          <div class="trend-bars">
            <span class="bar total" :style="{ height: `${trendBarHeight(item.total_views)}px` }" />
            <span class="bar news" :style="{ height: `${trendBarHeight(item.news_views)}px` }" />
            <span class="bar report" :style="{ height: `${trendBarHeight(item.report_views)}px` }" />
            <span class="bar journal" :style="{ height: `${trendBarHeight(item.journal_views)}px` }" />
          </div>
          <p>{{ (item.date || "").slice(5) || "-" }}</p>
          <small>总 {{ item.total_views || 0 }}</small>
        </article>
      </div>
      <div class="trend-legend">
        <span><i class="dot total" />总量</span>
        <span><i class="dot news" />新闻</span>
        <span><i class="dot report" />研报</span>
        <span><i class="dot journal" />期刊</span>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-head">
        <h3>用户分层</h3>
        <el-text type="info">高活跃 {{ activeSegments.length }} 人 / 沉默 {{ silentSegments.length }} 人</el-text>
      </div>
      <div class="segment-grid">
        <article class="segment-card">
          <div class="segment-head">
            <p>高活跃用户</p>
            <el-button
              v-if="canEditUsers"
              size="small"
              :loading="sendingSegment === 'ACTIVE'"
              @click="applyReminderTemplate('ACTIVE')"
            >
              套用模板
            </el-button>
          </div>
          <div class="segment-list">
            <div v-for="item in activeSegments" :key="`${item.segment}-${item.user_id}`" class="segment-item">
              <div>
                <p>{{ item.user_id }}（{{ item.user_phone || "-" }}）</p>
                <small>7天阅读 {{ item.view_count_7d }} 次 · 最近 {{ formatDateTime(item.last_viewed_at) }}</small>
              </div>
            </div>
            <p v-if="activeSegments.length === 0" class="empty-text">暂无高活跃用户</p>
          </div>
          <el-button
            v-if="canEditUsers"
            type="primary"
            plain
            size="small"
            :disabled="activeSegments.length === 0"
            :loading="sendingSegment === 'ACTIVE'"
            @click="handleSendSegment('ACTIVE')"
          >
            给高活跃用户发送
          </el-button>
        </article>

        <article class="segment-card">
          <div class="segment-head">
            <p>沉默用户</p>
            <el-button
              v-if="canEditUsers"
              size="small"
              :loading="sendingSegment === 'SILENT'"
              @click="applyReminderTemplate('SILENT')"
            >
              套用模板
            </el-button>
          </div>
          <div class="segment-list">
            <div v-for="item in silentSegments" :key="`${item.segment}-${item.user_id}`" class="segment-item">
              <div>
                <p>{{ item.user_id }}（{{ item.user_phone || "-" }}）</p>
                <small>最近阅读 {{ formatDateTime(item.last_viewed_at) }}</small>
              </div>
            </div>
            <p v-if="silentSegments.length === 0" class="empty-text">暂无沉默用户</p>
          </div>
          <el-button
            v-if="canEditUsers"
            type="warning"
            plain
            size="small"
            :disabled="silentSegments.length === 0"
            :loading="sendingSegment === 'SILENT'"
            @click="handleSendSegment('SILENT')"
          >
            给沉默用户发送
          </el-button>
        </article>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="toolbar" style="margin-bottom: 0">
        <el-input v-model="filters.user_id" clearable placeholder="用户ID" style="width: 160px" />
        <el-select v-model="filters.content_type" clearable placeholder="内容类型" style="width: 160px">
          <el-option v-for="item in contentTypeOptions" :key="item" :label="mapContentType(item)" :value="item" />
        </el-select>
        <el-input v-model="filters.keyword" clearable placeholder="关键词（标题/内容ID/手机号）" style="width: 240px" />
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
        <el-button :loading="exporting" @click="exportFilteredCSV">导出CSV</el-button>
      </div>
    </div>

    <div v-if="canEditUsers" class="card" style="margin-bottom: 12px">
      <div class="section-head">
        <h3>阅读提醒</h3>
        <el-text type="info">已选用户 {{ selectedUserIDs.length }} 人</el-text>
      </div>
      <div class="form-grid">
        <div class="form-item">
          <span>消息类型</span>
          <el-select v-model="messageForm.type">
            <el-option
              v-for="item in messageTypeOptions"
              :key="item"
              :label="mapMessageType(item)"
              :value="item"
            />
          </el-select>
        </div>
        <div class="form-item">
          <span>消息标题</span>
          <el-input v-model="messageForm.title" maxlength="128" />
        </div>
      </div>
      <div class="form-item" style="margin-top: 8px">
        <span>消息内容</span>
        <el-input v-model="messageForm.content" type="textarea" :rows="3" maxlength="500" show-word-limit />
      </div>
      <div class="form-actions">
        <el-button type="primary" :disabled="!canSendBatch" :loading="sendingBatch" @click="handleSendBatch">
          给已选用户发送提醒
        </el-button>
      </div>
    </div>

    <div v-if="sendFailures.length > 0" class="card" style="margin-bottom: 12px">
      <div class="section-head">
        <h3>发送失败明细</h3>
        <el-text type="warning">失败 {{ sendFailures.length }} 条</el-text>
      </div>
      <el-table :data="sendFailures" border stripe size="small" max-height="240">
        <el-table-column prop="user_id" label="用户ID" min-width="150" />
        <el-table-column prop="reason" label="失败原因" min-width="220" />
      </el-table>
    </div>

    <div class="card">
      <el-table
        :data="rows"
        border
        stripe
        v-loading="loading"
        row-key="id"
        empty-text="暂无阅读记录"
        @selection-change="handleSelectionChange"
      >
        <el-table-column v-if="canEditUsers" type="selection" width="52" reserve-selection />
        <el-table-column prop="user_id" label="用户ID" min-width="140" />
        <el-table-column label="手机号" min-width="130">
          <template #default="{ row }">
            {{ row.user_phone || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="类型" min-width="90">
          <template #default="{ row }">
            {{ mapContentType(row.content_type) }}
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="220" />
        <el-table-column prop="content_id" label="内容ID" min-width="140" />
        <el-table-column prop="source_page" label="来源页面" min-width="110" />
        <el-table-column label="阅读时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.viewed_at) }}
          </template>
        </el-table-column>
        <el-table-column v-if="canEditUsers" label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button
              size="small"
              :loading="sendingSingleID === row.id"
              @click="handleSendSingle(row)"
            >
              发送提醒
            </el-button>
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
  </div>
</template>

<style scoped>
.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.section-head h3 {
  margin: 0;
  font-size: 16px;
}

.segment-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.trend-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(90px, 1fr));
  gap: 8px;
}

.trend-item {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  padding: 8px 6px;
  text-align: center;
  display: grid;
  gap: 6px;
}

.trend-bars {
  height: 86px;
  display: flex;
  align-items: flex-end;
  justify-content: center;
  gap: 3px;
}

.bar {
  width: 10px;
  border-radius: 6px 6px 2px 2px;
  background: #d1d5db;
}

.bar.total {
  background: #2563eb;
}

.bar.news {
  background: #16a34a;
}

.bar.report {
  background: #f59e0b;
}

.bar.journal {
  background: #7c3aed;
}

.trend-item p {
  margin: 0;
  font-size: 12px;
  color: #374151;
}

.trend-item small {
  color: #6b7280;
}

.trend-legend {
  margin-top: 8px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  font-size: 12px;
  color: #6b7280;
}

.dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 4px;
}

.dot.total {
  background: #2563eb;
}

.dot.news {
  background: #16a34a;
}

.dot.report {
  background: #f59e0b;
}

.dot.journal {
  background: #7c3aed;
}

.segment-card {
  border: 1px solid var(--el-border-color-light);
  border-radius: 10px;
  padding: 10px;
  background: #fff;
  display: grid;
  gap: 8px;
}

.segment-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.segment-head p {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.segment-list {
  display: grid;
  gap: 8px;
  min-height: 120px;
}

.segment-item {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  padding: 8px;
}

.segment-item p {
  margin: 0;
  font-size: 13px;
}

.segment-item small {
  color: #6b7280;
}

.empty-text {
  margin: 0;
  color: #9ca3af;
  font-size: 12px;
}

@media (max-width: 1200px) {
  .segment-grid {
    grid-template-columns: 1fr;
  }
}
</style>
