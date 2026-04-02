<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  assignReviewTask,
  getWorkflowMetrics,
  listSystemConfigs,
  listReviewTasks,
  reviewTaskDecision,
  getNewsArticleDetail,
  getStockEventCluster,
  getStockSelectionRun,
  getFuturesSelectionRun
} from "../api/admin";
import { parseForecastAdminConfigMap } from "../lib/forecast-admin";
import { getSession, hasPermission } from "../lib/session";

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const metricsLoading = ref(false);
const targetLoading = ref(false);

const errorMessage = ref("");
const message = ref("");

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const tasks = ref([]);
const reviewTableRef = ref(null);
const selectedRows = ref([]);
const nowTick = ref(Date.now());
const autoRefreshEnabled = ref(true);
const activeTab = ref("todo"); // todo, queue, history

const currentUserID = ref(getSession()?.userID || "");
const canEditReview = hasPermission("review.edit");

const filters = reactive({
  module: "",
  status: "", // This will be managed by activeTab mostly
  submitter_id: "",
  reviewer_id: ""
});

const metrics = ref({
  pending_reviews: 0,
  approved_today: 0,
  rejected_today: 0,
  unread_messages: 0,
  total_messages: 0
});

const forecastReviewConfig = ref(null);

const detailVisible = ref(false);
const currentTask = ref(null);
const targetDetail = ref(null);
const detailAssigning = ref(false);
const detailDeciding = ref(false);
const quickActionKey = ref("");

const detailForm = reactive({
  reviewer_id: "",
  decision_status: "APPROVED",
  decision_note: ""
});

const moduleOptions = ["NEWS", "STOCK_EVENT", "STOCK_SELECTION", "FUTURES_SELECTION"];
const slaWarnHours = 24;
const slaDangerHours = 48;

const slaStats = computed(() => {
  let pendingTotal = 0;
  let warningCount = 0;
  let dangerCount = 0;
  tasks.value.forEach((task) => {
    const hours = getPendingHours(task);
    if (hours === null) return;
    pendingTotal += 1;
    if (hours >= slaDangerHours) dangerCount += 1;
    else if (hours >= slaWarnHours) warningCount += 1;
  });
  return { pendingTotal, warningCount, dangerCount, normalCount: Math.max(0, pendingTotal - warningCount - dangerCount) };
});

function getPendingHours(task) {
  if ((task.status || "").toUpperCase() !== "PENDING") return null;
  const submittedAt = new Date((task.submitted_at || "").replace(" ", "T"));
  if (isNaN(submittedAt.getTime())) return null;
  return Math.max(0, (nowTick.value - submittedAt.getTime()) / (1000 * 60 * 60));
}

function formatDuration(hours) {
  if (hours === null) return "-";
  if (hours < 1) return "<1h";
  if (hours < 24) return `${Math.floor(hours)}h`;
  return `${Math.floor(hours / 24)}d ${Math.floor(hours % 24)}h`;
}

const timerRefs = { clock: null, refresh: null };

function setupTimers() {
  timerRefs.clock = setInterval(() => { nowTick.value = Date.now(); }, 30000);
  timerRefs.refresh = setInterval(() => { if (autoRefreshEnabled.value && !loading.value) refreshAll(); }, 60000);
}

// Fixed Refresh logic: activeTab drives the API filters
async function refreshAll() {
  loading.value = true;
  metricsLoading.value = true;
  
  const apiFilters = { 
    module: filters.module,
    submitter_id: filters.submitter_id.trim(),
    page: page.value,
    page_size: pageSize.value
  };

  if (activeTab.value === "todo") {
    apiFilters.status = "PENDING";
    apiFilters.reviewer_id = currentUserID.value;
  } else if (activeTab.value === "queue") {
    apiFilters.status = "PENDING";
    apiFilters.reviewer_id = "NULL"; // Backend now handles this
  } else if (activeTab.value === "history") {
    apiFilters.status = "APPROVED,REJECTED"; // Backend now handles multi-status via FIND_IN_SET
  }

  try {
    const [m, t, config] = await Promise.all([
      getWorkflowMetrics({ module: filters.module }),
      listReviewTasks(apiFilters),
      listSystemConfigs({ keyword: "growth.forecast_", page: 1, page_size: 50 })
    ]);
    if (m) metrics.value = { ...metrics.value, ...m };
    tasks.value = t.items || [];
    total.value = t.total || 0;
    const configMap = Object.fromEntries((config.items || []).map(i => [i.config_key.toLowerCase(), i.config_value]));
    forecastReviewConfig.value = parseForecastAdminConfigMap(configMap);
  } catch (err) {
    errorMessage.value = "数据加载失败: " + err.message;
  } finally {
    loading.value = false;
    metricsLoading.value = false;
  }
}

async function fetchTargetDetail(task) {
  if (!task?.target_id || !task?.module) return;
  targetLoading.value = true;
  targetDetail.value = null;
  try {
    let data = null;
    const module = task.module.toUpperCase();
    if (module === "NEWS") data = await getNewsArticleDetail(task.target_id);
    else if (module === "STOCK_EVENT") data = await getStockEventCluster(task.target_id);
    else if (module === "STOCK_SELECTION") data = await getStockSelectionRun(task.target_id);
    else if (module === "FUTURES_SELECTION") data = await getFuturesSelectionRun(task.target_id);
    targetDetail.value = data;
  } catch (err) {
    console.error("Fetch target failed:", err);
  } finally {
    targetLoading.value = false;
  }
}

function openDetail(task) {
  currentTask.value = task;
  detailForm.reviewer_id = task.reviewer_id || "";
  detailForm.decision_status = "APPROVED";
  detailForm.decision_note = task.review_note || "";
  detailVisible.value = true;
  fetchTargetDetail(task);
}

async function handleDecision(task, status, note = "") {
  if (!canEditReview) return;
  quickActionKey.value = `${task.id}:${status}`;
  try {
    await reviewTaskDecision(task.id, status, note);
    message.value = `任务 ${task.id} 已${status === 'APPROVED' ? '通过' : '驳回'}`;
    errorMessage.value = "";
    refreshAll();
  } catch (err) {
    errorMessage.value = err.message || "操作失败";
  } finally {
    quickActionKey.value = "";
  }
}

// Batch Actions
const batchLoading = ref(false);
function handleSelectionChange(rows) { selectedRows.value = rows || []; }
async function handleBatchApprove() {
  if (selectedRows.value.length === 0) return;
  batchLoading.value = true;
  try {
    for (const task of selectedRows.value) {
      if (task.status === 'PENDING') await reviewTaskDecision(task.id, 'APPROVED', 'Batch approved');
    }
    message.value = `批量通过了 ${selectedRows.value.length} 个任务`;
    refreshAll();
  } catch (err) {
    errorMessage.value = err.message || "批量操作失败";
  } finally {
    batchLoading.value = false;
  }
}

function exportCSV() {
  const header = ["ID", "Module", "TargetID", "Status", "SubmittedAt"].join(",");
  const rows = tasks.value.map(t => [t.id, t.module, t.target_id, t.status, t.submitted_at].join(","));
  const csv = [header, ...rows].join("\n");
  const blob = new Blob(["\ufeff" + csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.setAttribute("href", url);
  link.setAttribute("download", `review_tasks_${new Date().getTime()}.csv`);
  link.click();
}

onMounted(() => {
  refreshAll();
  setupTimers();
});

onBeforeUnmount(() => { Object.values(timerRefs).forEach(t => t && clearInterval(t)); });

watch(activeTab, () => { page.value = 1; refreshAll(); });
</script>

<template>
  <div class="review-center-pro">
    <div class="hero-section">
      <div class="hero-header">
        <h1 class="pro-title">智能运营审核中心 <span class="badge">PRO MAX</span></h1>
        <div class="hero-actions">
          <el-button-group>
            <el-button @click="exportCSV">导出 CSV</el-button>
            <el-button type="primary" :loading="loading" @click="refreshAll">强制刷新</el-button>
            <el-button @click="autoRefreshEnabled = !autoRefreshEnabled">
              {{ autoRefreshEnabled ? '自动巡航' : '手动模式' }}
            </el-button>
          </el-button-group>
        </div>
      </div>

      <el-row :gutter="16" class="metrics-hero">
        <el-col :xs="12" :sm="6">
          <div class="pro-card glass warning">
            <div class="card-content">
              <div class="label">⏳ 待处理 / SLA超时</div>
              <div class="value">{{ metrics.pending_reviews }} <span class="sub">/ {{ slaStats.dangerCount }}</span></div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="6">
          <div class="pro-card glass success">
            <div class="card-content">
              <div class="label">✅ 今日已通过</div>
              <div class="value">{{ metrics.approved_today }} <span class="sub">APPROVED</span></div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="6">
          <div class="pro-card glass info">
            <div class="card-content">
              <div class="label">👤 指派给我</div>
              <div class="value">{{ tasks.filter(t => t.reviewer_id === currentUserID).length }} <span class="sub">MY TASKS</span></div>
            </div>
          </div>
        </el-col>
        <el-col :xs="12" :sm="6">
          <div class="pro-card glass primary">
            <div class="card-content">
              <div class="label">✉️ 未读通知</div>
              <div class="value">{{ metrics.unread_messages }} <span class="sub">MESSAGES</span></div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div class="main-stage">
      <div class="tab-control">
        <el-tabs v-model="activeTab" class="pro-tabs">
          <el-tab-pane label="待我处理" name="todo" />
          <el-tab-pane label="未分配队列" name="queue" />
          <el-tab-pane label="全量历史" name="history" />
        </el-tabs>
        <div class="tab-filters">
          <el-input v-model="filters.submitter_id" placeholder="提交人ID" size="small" style="width: 120px; margin-right: 8px" @change="refreshAll" />
          <el-select v-model="filters.module" placeholder="全部模块" clearable size="small" style="width: 120px" @change="refreshAll">
            <el-option v-for="m in moduleOptions" :key="m" :label="m" :value="m" />
          </el-select>
        </div>
      </div>

      <div class="table-container glass">
        <div v-if="selectedRows.length > 0" class="bulk-bar glass">
          <span>已选中 {{ selectedRows.length }} 项</span>
          <el-button type="success" size="small" :loading="batchLoading" @click="handleBatchApprove">批量通过</el-button>
          <el-button type="info" size="small" @click="reviewTableRef.clearSelection()">取消</el-button>
        </div>
        <el-table 
          ref="reviewTableRef"
          :data="tasks" 
          v-loading="loading" 
          border 
          stripe 
          size="default" 
          class="pro-table"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="45" />
          <el-table-column prop="id" label="任务ID" width="120" />
          <el-table-column label="审核对象" min-width="160">
            <template #default="{ row }">
              <div class="target-info">
                <el-tag size="small" :type="row.module === 'NEWS' ? 'primary' : 'warning'">{{ row.module }}</el-tag>
                <span class="target-id">{{ row.target_id }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="submitter_id" label="提交员" width="100" />
          <el-table-column label="SLA" width="130">
            <template #default="{ row }">
              <div v-if="row.status === 'PENDING'" class="sla-monitor">
                <el-progress 
                  :percentage="Math.min(100, (getPendingHours(row) / slaDangerHours) * 100)" 
                  :status="getPendingHours(row) > slaDangerHours ? 'exception' : (getPendingHours(row) > slaWarnHours ? 'warning' : 'success')"
                  :show-text="false"
                  :stroke-width="4"
                />
                <span class="sla-text">{{ formatDuration(getPendingHours(row)) }}</span>
              </div>
              <el-tag v-else size="small" :type="row.status === 'APPROVED' ? 'success' : 'danger'">{{ row.status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="submit_note" label="提交备注" min-width="180" show-overflow-tooltip />
          <el-table-column label="操作" width="220" align="right">
            <template #default="{ row }">
              <div class="pro-row-actions">
                <el-button link type="primary" size="small" @click="openDetail(row)">详情</el-button>
                <template v-if="row.status === 'PENDING' && canEditReview">
                  <el-button 
                    type="success" 
                    size="small" 
                    plain 
                    :loading="quickActionKey === `${row.id}:APPROVED`"
                    @click="handleDecision(row, 'APPROVED')"
                  >通过</el-button>
                </template>
                <span v-else-if="!canEditReview" class="muted">只读</span>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="pro-pagination">
        <el-pagination
          v-model:current-page="page"
          :page-size="pageSize"
          layout="total, prev, pager, next"
          :total="total"
          @current-change="refreshAll"
        />
      </div>
    </div>

    <!-- Review & Preview Drawer -->
    <el-drawer v-model="detailVisible" size="60%" class="pro-drawer" destroy-on-close>
      <template #header>
        <div class="drawer-header">
          <h2 class="drawer-title">任务决策中心 <span class="task-id">#{{ currentTask?.id }}</span></h2>
        </div>
      </template>

      <div class="drawer-layout" v-loading="targetLoading">
        <div class="preview-stage glass">
          <div class="stage-label">上下文内容预览</div>
          <div v-if="targetDetail" class="content-preview">
            <template v-if="currentTask?.module === 'NEWS'">
              <h3 class="preview-title">{{ targetDetail.title }}</h3>
              <div class="preview-meta">
                 <el-tag size="small">{{ targetDetail.category_name }}</el-tag>
                 <span>{{ targetDetail.published_at }}</span>
              </div>
              <div class="preview-body" v-html="targetDetail.content || '暂无内容'"></div>
            </template>
            <template v-else-if="currentTask?.module === 'STOCK_EVENT'">
              <h3 class="preview-title">事件标题：{{ targetDetail.cluster_title }}</h3>
              <div class="preview-body">{{ targetDetail.summary }}</div>
              <div class="preview-ext">
                <div v-for="link in targetDetail.source_links" :key="link" class="link-item">
                  <el-link :href="link" target="_blank">{{ link }}</el-link>
                </div>
              </div>
            </template>
            <template v-else>
               <pre class="raw-data">{{ JSON.stringify(targetDetail, null, 2) }}</pre>
            </template>
          </div>
          <el-empty v-else description="无法加载预览数据，请核对 Target ID 是否正确" />
        </div>

        <div class="action-stage">
          <div class="stage-label">审批决策</div>
          <el-form label-position="top">
            <el-form-item label="任务分配">
              <el-input v-model="detailForm.reviewer_id" placeholder="填写 reviewer_id" />
            </el-form-item>
            <el-form-item label="审批结论">
              <el-radio-group v-model="detailForm.decision_status">
                <el-radio-button label="APPROVED">审核通过</el-radio-button>
                <el-radio-button label="REJECTED">打回驳回</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="审批备注 (驳回必填)">
              <el-input v-model="detailForm.decision_note" type="textarea" :rows="4" placeholder="请填写意见..." />
            </el-form-item>
            <div class="stage-btns">
              <el-button type="primary" size="large" @click="handleDecision(currentTask, detailForm.decision_status, detailForm.decision_note)">确认并提交决策</el-button>
            </div>
          </el-form>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.review-center-pro {
  padding: 16px;
  background: #f1f5f9;
  min-height: 100vh;
  font-family: 'Inter', system-ui, -apple-system, sans-serif;
}

.hero-section {
  margin-bottom: 24px;
}

.pro-title {
  font-size: 24px;
  font-weight: 800;
  color: #1e293b;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 12px;
}

.badge {
  font-size: 10px;
  background: linear-gradient(135deg, #3b82f6, #8b5cf6);
  color: white;
  padding: 2px 6px;
  border-radius: 4px;
  letter-spacing: 1px;
}

.hero-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.metrics-hero {
  margin-top: 12px;
}

.pro-card {
  padding: 16px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.glass {
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(231, 235, 241, 0.6);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.03);
}

.pro-card.warning { border-left: 4px solid #f59e0b; }
.pro-card.success { border-left: 4px solid #10b981; }
.pro-card.info { border-left: 4px solid #3b82f6; }
.pro-card.primary { border-left: 4px solid #8b5cf6; }

.label {
  font-size: 11px;
  color: #64748b;
  font-weight: 700;
  margin-bottom: 4px;
}

.value {
  font-size: 20px;
  font-weight: 800;
  color: #1e293b;
}

.sub {
  font-size: 11px;
  opacity: 0.5;
  margin-left: 4px;
}

.main-stage {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.02);
  overflow: hidden;
}

.tab-control {
  padding: 0 16px;
  border-bottom: 1px solid #f1f5f9;
  display: flex;
  justify-content: space-between;
  align-items: center;
  height: 56px;
}

.bulk-bar {
  margin: 12px 16px;
  padding: 8px 16px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  gap: 16px;
  font-size: 13px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
}

.target-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.target-id {
  font-family: 'SF Mono', 'Monaco', monospace;
  font-weight: 600;
  font-size: 13px;
  color: #475569;
}

.sla-monitor {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.sla-text {
  font-size: 11px;
  font-weight: 700;
  color: #64748b;
}

.drawer-layout {
  display: grid;
  grid-template-columns: 1.5fr 1fr;
  gap: 20px;
  height: calc(100vh - 100px);
  padding: 0 20px 20px;
}

.preview-stage {
  padding: 20px;
  border-radius: 12px;
  overflow-y: auto;
  border: 1px solid #f1f5f9;
}

.stage-label {
  font-size: 11px;
  font-weight: 800;
  color: #94a3b8;
  margin-bottom: 12px;
  text-transform: uppercase;
}

.preview-title {
  font-size: 18px;
  font-weight: 800;
  color: #0f172a;
  margin-bottom: 8px;
}

.preview-meta {
  margin-bottom: 16px;
  display: flex;
  gap: 12px;
  color: #64748b;
  font-size: 12px;
}

.preview-body {
  line-height: 1.6;
  color: #334155;
  font-size: 14px;
}

.raw-data {
  background: #f8fafc;
  padding: 12px;
  border-radius: 6px;
  font-size: 11px;
  font-family: monospace;
}

.action-stage {
  background: #f8fafc;
  padding: 20px;
  border-radius: 12px;
  border: 1px solid #f1f5f9;
}

.stage-btns {
  margin-top: 24px;
}

.pro-pagination {
  padding: 16px;
  display: flex;
  justify-content: flex-end;
}

:deep(.el-tabs__header) {
  margin: 0;
}

:deep(.el-tabs__item) {
  height: 56px;
  line-height: 56px;
  font-weight: 700;
}
</style>
