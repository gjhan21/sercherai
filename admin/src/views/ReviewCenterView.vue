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
  if (status !== "APPROVED" && status !== "REJECTED") {
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

function nextPage() {
  if (page.value * pageSize.value >= total.value) {
    return;
  }
  page.value += 1;
  fetchTasks();
}

function prevPage() {
  if (page.value <= 1) {
    return;
  }
  page.value -= 1;
  fetchTasks();
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
      <button class="btn" :disabled="loading || metricsLoading" @click="refreshAll">
        {{ loading || metricsLoading ? "刷新中..." : "刷新" }}
      </button>
    </div>

    <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
    <div v-if="message" class="success-message">{{ message }}</div>

    <div class="grid grid-4" style="margin-bottom: 12px">
      <div class="metric-card">
        <div class="label">待审核任务</div>
        <div class="value">{{ metrics.pending_reviews || 0 }}</div>
      </div>
      <div class="metric-card">
        <div class="label">今日通过</div>
        <div class="value">{{ metrics.approved_today || 0 }}</div>
      </div>
      <div class="metric-card">
        <div class="label">今日驳回</div>
        <div class="value">{{ metrics.rejected_today || 0 }}</div>
      </div>
      <div class="metric-card">
        <div class="label">流程消息（未读/总）</div>
        <div class="value">{{ metrics.unread_messages || 0 }} / {{ metrics.total_messages || 0 }}</div>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <h3 style="margin-top: 0">提交审核任务</h3>
      <div class="form-grid">
        <div class="form-item">
          <label>模块</label>
          <select v-model="submitForm.module" class="select">
            <option value="NEWS">NEWS</option>
            <option value="STOCK">STOCK</option>
            <option value="FUTURES">FUTURES</option>
          </select>
        </div>
        <div class="form-item">
          <label>target_id</label>
          <input v-model="submitForm.target_id" class="input" placeholder="如 news_001 / sr_001" />
        </div>
        <div class="form-item">
          <label>reviewer_id（可选）</label>
          <input v-model="submitForm.reviewer_id" class="input" placeholder="admin_002" />
        </div>
        <div class="form-item">
          <label>提交备注</label>
          <input v-model="submitForm.submit_note" class="input" placeholder="请在今日内完成审核" />
        </div>
      </div>
      <div class="form-actions">
        <button class="btn btn-primary" :disabled="submitting" @click="handleSubmitReview">
          {{ submitting ? "提交中..." : "提交审核任务" }}
        </button>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="toolbar">
        <select v-model="filters.module" class="select">
          <option value="">全部模块</option>
          <option value="NEWS">NEWS</option>
          <option value="STOCK">STOCK</option>
          <option value="FUTURES">FUTURES</option>
        </select>
        <select v-model="filters.status" class="select">
          <option value="">全部状态</option>
          <option value="PENDING">PENDING</option>
          <option value="APPROVED">APPROVED</option>
          <option value="REJECTED">REJECTED</option>
        </select>
        <input v-model="filters.submitter_id" class="input" placeholder="submitter_id" />
        <input v-model="filters.reviewer_id" class="input" placeholder="reviewer_id" />
        <button class="btn" @click="applyFilters">查询</button>
        <button class="btn" @click="resetFilters">重置</button>
      </div>
    </div>

    <div class="card">
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>模块</th>
              <th>目标ID</th>
              <th>提交人</th>
              <th>审核人</th>
              <th>状态</th>
              <th>提交备注</th>
              <th>审核备注</th>
              <th>提交时间</th>
              <th>审核时间</th>
              <th class="text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in tasks" :key="task.id">
              <td>{{ task.id }}</td>
              <td>{{ task.module }}</td>
              <td>{{ task.target_id }}</td>
              <td>{{ task.submitter_id }}</td>
              <td>{{ task.reviewer_id || "-" }}</td>
              <td>{{ task.status }}</td>
              <td>{{ task.submit_note || "-" }}</td>
              <td>{{ task.review_note || "-" }}</td>
              <td>{{ task.submitted_at }}</td>
              <td>{{ task.reviewed_at || "-" }}</td>
              <td class="text-right">
                <template v-if="task.status === 'PENDING'">
                  <div class="toolbar" style="justify-content: flex-end">
                    <input v-model="assignDraftMap[task.id]" class="input" style="width: 120px" placeholder="reviewer_id" />
                    <button class="btn" @click="handleAssign(task)">分配</button>
                  </div>
                  <div class="toolbar" style="justify-content: flex-end">
                    <select v-model="decisionStatusMap[task.id]" class="select">
                      <option value="APPROVED">APPROVED</option>
                      <option value="REJECTED">REJECTED</option>
                    </select>
                    <input v-model="decisionNoteMap[task.id]" class="input" style="width: 150px" placeholder="review_note" />
                    <button class="btn btn-primary" @click="handleDecision(task)">提交结论</button>
                  </div>
                </template>
                <span v-else class="muted">已完成</span>
              </td>
            </tr>
            <tr v-if="!loading && tasks.length === 0">
              <td colspan="11" class="muted">暂无审核任务</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination">
        <span>第 {{ page }} 页，共 {{ total }} 条</span>
        <div class="toolbar">
          <button class="btn" :disabled="page <= 1 || loading" @click="prevPage">上一页</button>
          <button class="btn" :disabled="page * pageSize >= total || loading" @click="nextPage">
            下一页
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
