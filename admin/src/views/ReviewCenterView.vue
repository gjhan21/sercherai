<script setup>
import { onMounted, reactive, ref } from "vue";
import {
  assignReviewTask,
  getWorkflowMetrics,
  listReviewTasks,
  reviewTaskDecision,
  submitReviewTask
} from "../api/admin";

const loading = ref(false);
const metricsLoading = ref(false);
const submitting = ref(false);

const errorMessage = ref("");
const message = ref("");

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const tasks = ref([]);

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

const moduleOptions = ["NEWS", "STOCK", "FUTURES"];
const decisionOptions = ["APPROVED", "REJECTED"];

const assignDraftMap = ref({});
const decisionStatusMap = ref({});
const decisionNoteMap = ref({});

function syncDrafts() {
  const assignMap = {};
  const statusMap = {};
  const noteMap = {};
  tasks.value.forEach((task) => {
    assignMap[task.id] = task.reviewer_id || "";
    statusMap[task.id] = "APPROVED";
    noteMap[task.id] = "";
  });
  assignDraftMap.value = assignMap;
  decisionStatusMap.value = statusMap;
  decisionNoteMap.value = noteMap;
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
    syncDrafts();
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

async function handleAssign(task) {
  const reviewerID = (assignDraftMap.value[task.id] || "").trim();
  if (!reviewerID) {
    errorMessage.value = "reviewer_id 不能为空";
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await assignReviewTask(task.id, reviewerID);
    message.value = `任务 ${task.id} 已分配给 ${reviewerID}`;
    await refreshAll();
  } catch (error) {
    errorMessage.value = error.message || "分配审核任务失败";
  }
}

async function handleDecision(task) {
  const status = (decisionStatusMap.value[task.id] || "").trim();
  const note = (decisionNoteMap.value[task.id] || "").trim();
  if (!decisionOptions.includes(status)) {
    errorMessage.value = "审核结果必须为 APPROVED 或 REJECTED";
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await reviewTaskDecision(task.id, status, note);
    message.value = `任务 ${task.id} 已 ${status}`;
    await refreshAll();
  } catch (error) {
    errorMessage.value = error.message || "提交审核结论失败";
  }
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
      <el-button :loading="loading || metricsLoading" @click="refreshAll">刷新</el-button>
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
      <el-table :data="tasks" border stripe v-loading="loading" empty-text="暂无审核任务">
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
        <el-table-column label="操作" align="right" min-width="330">
          <template #default="{ row }">
            <div v-if="row.status === 'PENDING'" class="operation-block">
              <div class="inline-actions">
                <el-input v-model="assignDraftMap[row.id]" size="small" placeholder="reviewer_id" style="width: 130px" />
                <el-button size="small" @click="handleAssign(row)">分配</el-button>
              </div>
              <div class="inline-actions">
                <el-select v-model="decisionStatusMap[row.id]" size="small" style="width: 125px">
                  <el-option label="APPROVED" value="APPROVED" />
                  <el-option label="REJECTED" value="REJECTED" />
                </el-select>
                <el-input v-model="decisionNoteMap[row.id]" size="small" placeholder="review_note" style="width: 150px" />
                <el-button size="small" type="primary" @click="handleDecision(row)">提交结论</el-button>
              </div>
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

.operation-block {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.inline-actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
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
</style>
