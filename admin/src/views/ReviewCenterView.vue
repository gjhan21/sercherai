<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from "vue";
import {
  assignReviewTask,
  getWorkflowMetrics,
  listReviewTasks,
  reviewTaskDecision,
  submitReviewTask
} from "../api/admin";
import { getSession } from "../lib/session";

const loading = ref(false);
const metricsLoading = ref(false);
const submitting = ref(false);

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
const batchReviewerID = ref("");
const batchRejectDialogVisible = ref(false);
const batchRejectReason = ref("");
const batchAssigning = ref(false);
const batchApproving = ref(false);
const batchRejecting = ref(false);
const timerRefs = {
  clock: null,
  refresh: null
};

const currentUserID = ref(getSession()?.userID || "");

const filters = reactive({
  module: "",
  status: "",
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

const submitForm = reactive({
  module: "NEWS",
  target_id: "",
  reviewer_id: "",
  submit_note: ""
});

const detailVisible = ref(false);
const currentTask = ref(null);
const detailAssigning = ref(false);
const detailDeciding = ref(false);
const quickActionKey = ref("");
const rejectDialogVisible = ref(false);
const rejectTask = ref(null);
const rejectReason = ref("");
const rejectSubmitting = ref(false);

const detailForm = reactive({
  reviewer_id: "",
  decision_status: "APPROVED",
  decision_note: ""
});

const moduleOptions = ["NEWS", "STOCK", "FUTURES"];
const decisionOptions = ["APPROVED", "REJECTED"];
const slaWarnHours = 24;
const slaDangerHours = 48;
const slaSortMode = ref("sla_desc");
const showOnlySLAWarning = ref(false);

const slaStats = computed(() => {
  let pendingTotal = 0;
  let warningCount = 0;
  let dangerCount = 0;
  tasks.value.forEach((task) => {
    const hours = getPendingHours(task);
    if (hours === null) {
      return;
    }
    pendingTotal += 1;
    if (hours >= slaDangerHours) {
      dangerCount += 1;
      return;
    }
    if (hours >= slaWarnHours) {
      warningCount += 1;
    }
  });
  return {
    pendingTotal,
    warningCount,
    dangerCount,
    normalCount: Math.max(0, pendingTotal - warningCount - dangerCount)
  };
});

const displayTasks = computed(() => {
  const source = [...tasks.value];
  if (showOnlySLAWarning.value) {
    return source
      .filter((task) => {
        const hours = getPendingHours(task);
        return hours !== null && hours >= slaWarnHours;
      })
      .sort((a, b) => (getPendingHours(b) || 0) - (getPendingHours(a) || 0));
  }
  if (slaSortMode.value === "sla_desc") {
    return source.sort((a, b) => {
      const left = getPendingHours(a);
      const right = getPendingHours(b);
      const leftHours = left === null ? -1 : left;
      const rightHours = right === null ? -1 : right;
      return rightHours - leftHours;
    });
  }
  return source;
});

const reviewEmptyText = computed(() => {
  if (!showOnlySLAWarning.value) {
    return "暂无审核任务";
  }
  return "当前页暂无超时/预警任务";
});

const selectedPendingRows = computed(() =>
  selectedRows.value.filter((task) => (task.status || "").toUpperCase() === "PENDING")
);
const selectedPendingCount = computed(() => selectedPendingRows.value.length);
const selectedEligibleDecisionRows = computed(() => selectedPendingRows.value.filter((task) => canQuickDecision(task)));
const selectedBlockedCount = computed(() => Math.max(0, selectedPendingCount.value - selectedEligibleDecisionRows.value.length));

function clearSelection() {
  selectedRows.value = [];
  nextTick(() => {
    reviewTableRef.value?.clearSelection();
  });
}

function handleSelectionChange(rows) {
  selectedRows.value = rows || [];
}

function updateNowTick() {
  nowTick.value = Date.now();
}

function clearTimers() {
  if (timerRefs.clock) {
    clearInterval(timerRefs.clock);
    timerRefs.clock = null;
  }
  if (timerRefs.refresh) {
    clearInterval(timerRefs.refresh);
    timerRefs.refresh = null;
  }
}

function setupTimers() {
  clearTimers();
  timerRefs.clock = setInterval(() => {
    updateNowTick();
  }, 30 * 1000);
  timerRefs.refresh = setInterval(() => {
    if (!autoRefreshEnabled.value || loading.value || metricsLoading.value) {
      return;
    }
    refreshAll();
  }, 60 * 1000);
}

async function fetchMetrics() {
  metricsLoading.value = true;
  try {
    const data = await getWorkflowMetrics({ module: filters.module });
    metrics.value = data || {};
  } catch {
    metrics.value = {
      pending_reviews: 0,
      approved_today: 0,
      rejected_today: 0,
      unread_messages: 0,
      total_messages: 0
    };
  } finally {
    metricsLoading.value = false;
  }
}

function syncCurrentTask() {
  if (!currentTask.value?.id) {
    return;
  }
  const found = tasks.value.find((task) => task.id === currentTask.value.id);
  if (found) {
    currentTask.value = found;
  }
}

async function fetchTasks() {
  loading.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const data = await listReviewTasks({
      module: filters.module,
      status: filters.status,
      submitter_id: filters.submitter_id.trim(),
      reviewer_id: filters.reviewer_id.trim(),
      page: page.value,
      page_size: pageSize.value
    });
    tasks.value = data.items || [];
    total.value = data.total || 0;
    syncCurrentTask();
  } catch (error) {
    errorMessage.value = error.message || "加载审核任务失败";
  } finally {
    loading.value = false;
  }
}

async function refreshAll() {
  await Promise.all([fetchMetrics(), fetchTasks()]);
  clearSelection();
  updateNowTick();
}

async function handleSubmitReview() {
  errorMessage.value = "";
  message.value = "";
  const payload = {
    module: submitForm.module,
    target_id: submitForm.target_id.trim(),
    reviewer_id: submitForm.reviewer_id.trim(),
    submit_note: submitForm.submit_note.trim()
  };
  if (!payload.target_id) {
    errorMessage.value = "target_id 不能为空";
    return;
  }
  submitting.value = true;
  try {
    const result = await submitReviewTask(payload);
    message.value = `审核任务已提交：${result.id || ""}`;
    submitForm.target_id = "";
    submitForm.submit_note = "";
    await refreshAll();
  } catch (error) {
    errorMessage.value = error.message || "提交审核任务失败";
  } finally {
    submitting.value = false;
  }
}

function openTaskDetail(task) {
  currentTask.value = { ...task };
  detailForm.reviewer_id = task.reviewer_id || "";
  detailForm.decision_status = "APPROVED";
  detailForm.decision_note = task.review_note || "";
  detailVisible.value = true;
}

function canQuickDecision(task) {
  if ((task.status || "").toUpperCase() !== "PENDING") {
    return false;
  }
  const reviewerID = (task.reviewer_id || "").trim();
  if (!reviewerID) {
    return true;
  }
  return reviewerID === currentUserID.value;
}

function isQuickActionLoading(task, status) {
  return quickActionKey.value === `${task.id}:${status}`;
}

async function handleQuickDecision(task, status, note = "") {
  if (!canQuickDecision(task)) {
    errorMessage.value = "任务已分配给其他审核员，请先在详情中重新分配";
    return false;
  }
  if (!decisionOptions.includes(status)) {
    errorMessage.value = "审核结果必须为 APPROVED 或 REJECTED";
    return false;
  }
  const normalizedNote = (note || "").trim();
  if (status === "REJECTED" && !normalizedNote) {
    errorMessage.value = "快速驳回必须填写原因";
    return false;
  }
  errorMessage.value = "";
  message.value = "";
  quickActionKey.value = `${task.id}:${status}`;
  try {
    await reviewTaskDecision(task.id, status, normalizedNote);
    message.value = `任务 ${task.id} 已${status === "APPROVED" ? "快速通过" : "快速驳回"}`;
    await refreshAll();
    return true;
  } catch (error) {
    errorMessage.value = error.message || "快捷审批失败";
    return false;
  } finally {
    quickActionKey.value = "";
  }
}

function openQuickReject(task) {
  if (!canQuickDecision(task)) {
    errorMessage.value = "任务已分配给其他审核员，请先在详情中重新分配";
    return;
  }
  rejectTask.value = task;
  rejectReason.value = "";
  rejectDialogVisible.value = true;
}

async function submitQuickReject() {
  if (!rejectTask.value?.id) {
    return;
  }
  const note = rejectReason.value.trim();
  if (!note) {
    errorMessage.value = "请填写驳回原因";
    return;
  }
  rejectSubmitting.value = true;
  try {
    const ok = await handleQuickDecision(rejectTask.value, "REJECTED", note);
    if (!ok) {
      return;
    }
    rejectDialogVisible.value = false;
    rejectTask.value = null;
    rejectReason.value = "";
  } finally {
    rejectSubmitting.value = false;
  }
}

async function handleBatchAssign() {
  const reviewerID = batchReviewerID.value.trim();
  if (!reviewerID) {
    errorMessage.value = "请先填写批量分配 reviewer_id";
    return;
  }
  if (selectedPendingCount.value <= 0) {
    errorMessage.value = "请先勾选待审核任务";
    return;
  }
  errorMessage.value = "";
  message.value = "";
  batchAssigning.value = true;
  let success = 0;
  let failed = 0;
  for (const task of selectedPendingRows.value) {
    try {
      await assignReviewTask(task.id, reviewerID);
      success += 1;
    } catch {
      failed += 1;
    }
  }
  batchAssigning.value = false;
  await refreshAll();
  message.value = `批量分配完成：成功 ${success}，失败 ${failed}`;
}

async function handleBatchApprove() {
  if (selectedPendingCount.value <= 0) {
    errorMessage.value = "请先勾选待审核任务";
    return;
  }
  errorMessage.value = "";
  message.value = "";
  batchApproving.value = true;
  let success = 0;
  let failed = 0;
  let skipped = 0;
  for (const task of selectedPendingRows.value) {
    if (!canQuickDecision(task)) {
      skipped += 1;
      continue;
    }
    try {
      await reviewTaskDecision(task.id, "APPROVED", "");
      success += 1;
    } catch {
      failed += 1;
    }
  }
  batchApproving.value = false;
  await refreshAll();
  message.value = `批量通过完成：成功 ${success}，失败 ${failed}，跳过 ${skipped}`;
}

function openBatchRejectDialog() {
  if (selectedPendingCount.value <= 0) {
    errorMessage.value = "请先勾选待审核任务";
    return;
  }
  batchRejectReason.value = "";
  batchRejectDialogVisible.value = true;
}

async function submitBatchReject() {
  const note = batchRejectReason.value.trim();
  if (!note) {
    errorMessage.value = "批量驳回必须填写原因";
    return;
  }
  errorMessage.value = "";
  message.value = "";
  batchRejecting.value = true;
  let success = 0;
  let failed = 0;
  let skipped = 0;
  for (const task of selectedPendingRows.value) {
    if (!canQuickDecision(task)) {
      skipped += 1;
      continue;
    }
    try {
      await reviewTaskDecision(task.id, "REJECTED", note);
      success += 1;
    } catch {
      failed += 1;
    }
  }
  batchRejecting.value = false;
  batchRejectDialogVisible.value = false;
  await refreshAll();
  message.value = `批量驳回完成：成功 ${success}，失败 ${failed}，跳过 ${skipped}`;
}

async function handleDetailAssign() {
  if (!currentTask.value?.id) {
    return;
  }
  const reviewerID = detailForm.reviewer_id.trim();
  if (!reviewerID) {
    errorMessage.value = "reviewer_id 不能为空";
    return;
  }
  errorMessage.value = "";
  message.value = "";
  detailAssigning.value = true;
  try {
    await assignReviewTask(currentTask.value.id, reviewerID);
    message.value = `任务 ${currentTask.value.id} 已分配给 ${reviewerID}`;
    await refreshAll();
    if (currentTask.value) {
      currentTask.value.reviewer_id = reviewerID;
    }
  } catch (error) {
    errorMessage.value = error.message || "分配审核任务失败";
  } finally {
    detailAssigning.value = false;
  }
}

async function handleDetailDecision() {
  if (!currentTask.value?.id) {
    return;
  }
  const status = (detailForm.decision_status || "").trim();
  const note = (detailForm.decision_note || "").trim();
  if (!decisionOptions.includes(status)) {
    errorMessage.value = "审核结果必须为 APPROVED 或 REJECTED";
    return;
  }
  if (status === "REJECTED" && !note) {
    errorMessage.value = "驳回必须填写审核备注";
    return;
  }
  errorMessage.value = "";
  message.value = "";
  detailDeciding.value = true;
  try {
    await reviewTaskDecision(currentTask.value.id, status, note);
    message.value = `任务 ${currentTask.value.id} 已 ${status}`;
    await refreshAll();
  } catch (error) {
    errorMessage.value = error.message || "提交审核结论失败";
  } finally {
    detailDeciding.value = false;
  }
}

function reviewerHint(task) {
  const reviewerID = (task.reviewer_id || "").trim();
  if (!reviewerID) {
    return "未分配审核员";
  }
  if (reviewerID === currentUserID.value) {
    return "当前任务分配给你";
  }
  return `当前任务分配给 ${reviewerID}`;
}

function parseTaskDateTime(value) {
  const raw = (value || "").trim();
  if (!raw) {
    return null;
  }
  const normalized = raw.includes("T") ? raw : raw.replace(" ", "T");
  const parsed = new Date(normalized);
  if (Number.isNaN(parsed.getTime())) {
    return null;
  }
  return parsed;
}

function getPendingHours(task) {
  if ((task.status || "").toUpperCase() !== "PENDING") {
    return null;
  }
  const submittedAt = parseTaskDateTime(task.submitted_at);
  if (!submittedAt) {
    return null;
  }
  return Math.max(0, (nowTick.value - submittedAt.getTime()) / (1000 * 60 * 60));
}

function formatPendingDuration(task) {
  const hours = getPendingHours(task);
  if (hours === null) {
    return "-";
  }
  if (hours < 1) {
    return "<1h";
  }
  if (hours < 24) {
    return `${Math.floor(hours)}h`;
  }
  const days = Math.floor(hours / 24);
  const remainHours = Math.floor(hours % 24);
  return `${days}d ${remainHours}h`;
}

function slaTagType(task) {
  const hours = getPendingHours(task);
  if (hours === null) {
    return "info";
  }
  if (hours >= slaDangerHours) {
    return "danger";
  }
  if (hours >= slaWarnHours) {
    return "warning";
  }
  return "success";
}

function slaLabel(task) {
  const hours = getPendingHours(task);
  if (hours === null) {
    return "-";
  }
  if (hours >= slaDangerHours) {
    return "超时严重";
  }
  if (hours >= slaWarnHours) {
    return "即将超时";
  }
  return "正常";
}

function reviewTableRowClassName({ row }) {
  const hours = getPendingHours(row);
  if (hours === null) {
    return "";
  }
  if (hours >= slaDangerHours) {
    return "row-sla-danger";
  }
  if (hours >= slaWarnHours) {
    return "row-sla-warning";
  }
  return "";
}

function applyFilters() {
  page.value = 1;
  refreshAll();
}

function resetFilters() {
  filters.module = "";
  filters.status = "";
  filters.submitter_id = "";
  filters.reviewer_id = "";
  slaSortMode.value = "sla_desc";
  showOnlySLAWarning.value = false;
  page.value = 1;
  refreshAll();
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  fetchTasks();
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "APPROVED") return "success";
  if (normalized === "REJECTED") return "danger";
  if (normalized === "PENDING") return "warning";
  return "info";
}

onMounted(() => {
  refreshAll();
  updateNowTick();
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
        <h1 class="page-title">审核中心</h1>
        <p class="muted">提交审核、分配审核员、处理通过/驳回</p>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <el-tag type="info">当前管理员：{{ currentUserID || "-" }}</el-tag>
        <el-button :loading="loading || metricsLoading" @click="refreshAll">刷新</el-button>
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

    <div class="card" style="margin-bottom: 12px" v-loading="metricsLoading">
      <div class="grid grid-4 metrics-grid">
        <div class="metric-item">
          <div class="metric-label">待审核任务</div>
          <div class="metric-value">{{ metrics.pending_reviews || 0 }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">今日通过</div>
          <div class="metric-value">{{ metrics.approved_today || 0 }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">今日驳回</div>
          <div class="metric-value">{{ metrics.rejected_today || 0 }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">流程消息（未读/总）</div>
          <div class="metric-value">{{ metrics.unread_messages || 0 }} / {{ metrics.total_messages || 0 }}</div>
        </div>
      </div>
      <div class="sla-overview">
        <el-tag type="success">SLA正常：{{ slaStats.normalCount }}</el-tag>
        <el-tag type="warning">SLA预警：{{ slaStats.warningCount }}</el-tag>
        <el-tag type="danger">SLA超时：{{ slaStats.dangerCount }}</el-tag>
        <el-text type="info">当前页待审核总数：{{ slaStats.pendingTotal }}</el-text>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <h3 style="margin: 0">提交审核任务</h3>
      </div>
      <el-form label-width="110px">
        <div class="dialog-grid">
          <el-form-item label="模块">
            <el-select v-model="submitForm.module">
              <el-option v-for="item in moduleOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="target_id" required>
            <el-input v-model="submitForm.target_id" placeholder="如 news_001 / sr_001" />
          </el-form-item>
          <el-form-item label="reviewer_id">
            <el-input v-model="submitForm.reviewer_id" placeholder="admin_002" />
          </el-form-item>
          <el-form-item label="提交备注">
            <el-input v-model="submitForm.submit_note" placeholder="请在今日内完成审核" />
          </el-form-item>
        </div>
      </el-form>
      <el-button type="primary" :loading="submitting" @click="handleSubmitReview">提交审核任务</el-button>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="toolbar" style="margin-bottom: 0">
        <el-select v-model="filters.module" clearable placeholder="全部模块" style="width: 150px">
          <el-option v-for="item in moduleOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-select v-model="filters.status" clearable placeholder="全部状态" style="width: 150px">
          <el-option label="PENDING" value="PENDING" />
          <el-option label="APPROVED" value="APPROVED" />
          <el-option label="REJECTED" value="REJECTED" />
        </el-select>
        <el-input v-model="filters.submitter_id" clearable placeholder="submitter_id" style="width: 180px" />
        <el-input v-model="filters.reviewer_id" clearable placeholder="reviewer_id" style="width: 180px" />
        <el-select v-model="slaSortMode" style="width: 160px">
          <el-option label="默认排序" value="default" />
          <el-option label="按SLA降序" value="sla_desc" />
        </el-select>
        <el-switch v-model="showOnlySLAWarning" inline-prompt active-text="只看超时" inactive-text="全部任务" />
        <el-switch v-model="autoRefreshEnabled" inline-prompt active-text="自动刷新" inactive-text="手动刷新" />
        <el-text type="info">SLA每30秒自动刷新，列表每60秒自动拉取</el-text>
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <h3 style="margin: 0">批量操作</h3>
        <el-text type="info">
          已选 {{ selectedRows.length }} 条，待审核 {{ selectedPendingCount }} 条，可审批 {{
            selectedEligibleDecisionRows.length
          }} 条，受限 {{ selectedBlockedCount }} 条
        </el-text>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <el-input v-model="batchReviewerID" placeholder="批量分配 reviewer_id" style="width: 220px" />
        <el-button :loading="batchAssigning" :disabled="selectedPendingCount <= 0" @click="handleBatchAssign">
          批量分配
        </el-button>
        <el-button
          type="success"
          plain
          :loading="batchApproving"
          :disabled="selectedEligibleDecisionRows.length <= 0"
          @click="handleBatchApprove"
        >
          批量通过
        </el-button>
        <el-button
          type="danger"
          plain
          :disabled="selectedEligibleDecisionRows.length <= 0"
          @click="openBatchRejectDialog"
        >
          批量驳回
        </el-button>
        <el-button @click="clearSelection">清空勾选</el-button>
      </div>
    </div>

    <div class="card">
      <el-table
        ref="reviewTableRef"
        :data="displayTasks"
        row-key="id"
        border
        stripe
        v-loading="loading"
        :empty-text="reviewEmptyText"
        :row-class-name="reviewTableRowClassName"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="52" reserve-selection />
        <el-table-column prop="id" label="ID" min-width="130" />
        <el-table-column prop="module" label="模块" min-width="100" />
        <el-table-column prop="target_id" label="目标ID" min-width="130" />
        <el-table-column prop="submitter_id" label="提交人" min-width="130" />
        <el-table-column label="审核人" min-width="130">
          <template #default="{ row }">
            {{ row.reviewer_id || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="SLA" min-width="140">
          <template #default="{ row }">
            <template v-if="row.status === 'PENDING'">
              <el-tag :type="slaTagType(row)">{{ slaLabel(row) }}</el-tag>
              <div class="sla-sub">{{ formatPendingDuration(row) }}</div>
            </template>
            <el-text v-else type="info">-</el-text>
          </template>
        </el-table-column>
        <el-table-column label="提交备注" min-width="180">
          <template #default="{ row }">
            {{ row.submit_note || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="审核备注" min-width="180">
          <template #default="{ row }">
            {{ row.review_note || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="submitted_at" label="提交时间" min-width="180" />
        <el-table-column label="审核时间" min-width="180">
          <template #default="{ row }">
            {{ row.reviewed_at || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="操作" align="right" min-width="360">
          <template #default="{ row }">
            <div class="operation-inline">
              <el-button size="small" @click="openTaskDetail(row)">详情</el-button>
              <el-button
                v-if="row.status === 'PENDING'"
                size="small"
                type="success"
                plain
                :disabled="!canQuickDecision(row)"
                :loading="isQuickActionLoading(row, 'APPROVED')"
                @click="handleQuickDecision(row, 'APPROVED', '')"
              >
                快速通过
              </el-button>
              <el-button
                v-if="row.status === 'PENDING'"
                size="small"
                type="danger"
                plain
                :disabled="!canQuickDecision(row)"
                :loading="isQuickActionLoading(row, 'REJECTED')"
                @click="openQuickReject(row)"
              >
                快速驳回
              </el-button>
            </div>
            <div v-if="row.status === 'PENDING'" class="reviewer-hint">
              {{ reviewerHint(row) }}
            </div>
            <el-text v-else type="info">已完成</el-text>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-text type="info">第 {{ page }} 页，当前展示 {{ displayTasks.length }} 条，共 {{ total }} 条</el-text>
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

    <el-dialog v-model="rejectDialogVisible" title="快速驳回任务" width="520px" destroy-on-close>
      <el-alert
        title="快速驳回会立即将任务状态置为 REJECTED，请填写明确原因。"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
      />
      <el-form label-width="80px">
        <el-form-item label="任务ID">
          <el-text>{{ rejectTask?.id || "-" }}</el-text>
        </el-form-item>
        <el-form-item label="驳回原因" required>
          <el-input
            v-model="rejectReason"
            type="textarea"
            :rows="4"
            maxlength="300"
            show-word-limit
            resize="vertical"
            placeholder="请填写具体驳回原因，例如：证据不足、字段缺失、内容不合规"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="rejectSubmitting" @click="submitQuickReject">确认驳回</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="batchRejectDialogVisible" title="批量驳回任务" width="560px" destroy-on-close>
      <el-alert
        title="会对已勾选且可审批的待审核任务批量执行驳回，请确认原因清晰。"
        type="warning"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
      />
      <el-form label-width="110px">
        <el-form-item label="任务统计">
          <el-text>待审核 {{ selectedPendingCount }} 条，可驳回 {{ selectedEligibleDecisionRows.length }} 条</el-text>
        </el-form-item>
        <el-form-item label="统一驳回原因" required>
          <el-input
            v-model="batchRejectReason"
            type="textarea"
            :rows="4"
            maxlength="300"
            show-word-limit
            resize="vertical"
            placeholder="请填写统一驳回原因，例如：材料不完整、校验不通过"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchRejectDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="batchRejecting" @click="submitBatchReject">确认批量驳回</el-button>
      </template>
    </el-dialog>

    <el-drawer v-model="detailVisible" size="560px" destroy-on-close>
      <template #header>
        <div class="drawer-title">审核任务详情</div>
      </template>

      <template v-if="currentTask">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="任务ID">{{ currentTask.id }}</el-descriptions-item>
          <el-descriptions-item label="模块">{{ currentTask.module }}</el-descriptions-item>
          <el-descriptions-item label="目标ID">{{ currentTask.target_id }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(currentTask.status)">{{ currentTask.status }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="SLA">
            <template v-if="(currentTask.status || '').toUpperCase() === 'PENDING'">
              <el-tag :type="slaTagType(currentTask)">{{ slaLabel(currentTask) }}</el-tag>
              <span class="sla-inline">{{ formatPendingDuration(currentTask) }}</span>
            </template>
            <template v-else>-</template>
          </el-descriptions-item>
          <el-descriptions-item label="提交人">{{ currentTask.submitter_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="审核人">{{ currentTask.reviewer_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="提交备注">{{ currentTask.submit_note || "-" }}</el-descriptions-item>
          <el-descriptions-item label="审核备注">{{ currentTask.review_note || "-" }}</el-descriptions-item>
          <el-descriptions-item label="提交时间">{{ currentTask.submitted_at || "-" }}</el-descriptions-item>
          <el-descriptions-item label="审核时间">{{ currentTask.reviewed_at || "-" }}</el-descriptions-item>
        </el-descriptions>

        <template v-if="currentTask.status === 'PENDING'">
          <div class="detail-section">
            <div class="detail-section-title">分配审核员</div>
            <div class="inline-actions inline-actions--left">
              <el-input v-model="detailForm.reviewer_id" placeholder="reviewer_id" style="width: 220px" />
              <el-button :loading="detailAssigning" @click="handleDetailAssign">保存分配</el-button>
            </div>
          </div>

          <div class="detail-section">
            <div class="detail-section-title">提交审核结论</div>
            <el-form label-width="88px">
              <el-form-item label="结论">
                <el-radio-group v-model="detailForm.decision_status">
                  <el-radio-button label="APPROVED">APPROVED</el-radio-button>
                  <el-radio-button label="REJECTED">REJECTED</el-radio-button>
                </el-radio-group>
              </el-form-item>
              <el-form-item label="备注">
                <el-input
                  v-model="detailForm.decision_note"
                  type="textarea"
                  :rows="4"
                  maxlength="300"
                  show-word-limit
                  resize="vertical"
                  placeholder="若驳回则必填，建议填写审核说明"
                />
              </el-form-item>
            </el-form>
            <el-text type="info">当结论为 REJECTED 时，备注为必填。</el-text>
            <el-button type="primary" :loading="detailDeciding" @click="handleDetailDecision">提交结论</el-button>
          </div>
        </template>

        <el-alert
          v-else
          title="该任务已完成，不能再次分配或审批。"
          type="info"
          :closable="false"
          show-icon
          style="margin-top: 14px"
        />
      </template>
    </el-drawer>
  </div>
</template>

<style scoped>
.metrics-grid {
  gap: 12px;
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

.sla-overview {
  margin-top: 12px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.operation-inline {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
}

.sla-sub {
  margin-top: 4px;
  font-size: 12px;
  color: #6b7280;
}

.sla-inline {
  margin-left: 8px;
  font-size: 12px;
  color: #6b7280;
}

.reviewer-hint {
  margin-top: 6px;
  font-size: 12px;
  color: #6b7280;
}

.detail-section {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid #e5e7eb;
}

.detail-section-title {
  margin-bottom: 10px;
  font-size: 14px;
  font-weight: 600;
  color: #111827;
}

.drawer-title {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
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

.dialog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 0 12px;
}

:deep(.dialog-grid .el-form-item) {
  margin-bottom: 14px;
}

:deep(.dialog-grid .el-select) {
  width: 100%;
}

:deep(.row-sla-warning > td.el-table__cell) {
  background: #fff7e6 !important;
}

:deep(.row-sla-danger > td.el-table__cell) {
  background: #fff1f0 !important;
}
</style>
