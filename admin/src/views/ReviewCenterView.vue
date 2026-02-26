<script setup>
import { onMounted, reactive, ref } from "vue";
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
  return Math.max(0, (Date.now() - submittedAt.getTime()) / (1000 * 60 * 60));
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

onMounted(refreshAll);
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
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </div>

    <div class="card">
      <el-table
        :data="tasks"
        border
        stripe
        v-loading="loading"
        empty-text="暂无审核任务"
        :row-class-name="reviewTableRowClassName"
      >
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
