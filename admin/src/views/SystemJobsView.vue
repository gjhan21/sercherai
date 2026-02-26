<script setup>
import { onMounted, reactive, ref } from "vue";
import {
  createSchedulerJobDefinition,
  getSchedulerJobMetrics,
  listSchedulerJobDefinitions,
  listSchedulerJobRuns,
  retrySchedulerJobRun,
  triggerSchedulerJob,
  updateSchedulerJobDefinition,
  updateSchedulerJobDefinitionStatus
} from "../api/admin";

const metricsLoading = ref(false);
const defsLoading = ref(false);
const runsLoading = ref(false);
const submittingDefinition = ref(false);
const triggeringJob = ref(false);

const errorMessage = ref("");
const message = ref("");

const metricFilter = reactive({
  job_name: ""
});

const metrics = ref({
  today_total: 0,
  today_success: 0,
  today_failed: 0,
  today_running: 0
});

const definitionFilters = reactive({
  status: "",
  module: ""
});
const definitionPage = ref(1);
const definitionPageSize = ref(20);
const definitionTotal = ref(0);
const definitions = ref([]);
const definitionStatusMap = ref({});

const definitionFormVisible = ref(false);
const definitionFormMode = ref("create");
const definitionForm = reactive({
  id: "",
  job_name: "",
  display_name: "",
  module: "SYSTEM",
  cron_expr: "0 0 9 * * *",
  status: "ACTIVE"
});

const runFilters = reactive({
  job_name: "",
  status: ""
});
const runPage = ref(1);
const runPageSize = ref(20);
const runTotal = ref(0);
const runs = ref([]);

const retrySimMap = ref({});
const retrySummaryMap = ref({});
const retryErrorMap = ref({});

const triggerForm = reactive({
  job_name: "",
  trigger_source: "MANUAL",
  simulate_status: "",
  result_summary: "",
  error_message: ""
});

const moduleOptions = ["STOCK", "FUTURES", "NEWS", "SYSTEM"];
const definitionStatusOptions = ["ACTIVE", "DISABLED"];
const runStatusOptions = ["RUNNING", "SUCCESS", "FAILED"];
const simulateStatusOptions = ["SUCCESS", "FAILED"];

function resetDefinitionForm() {
  Object.assign(definitionForm, {
    id: "",
    job_name: "",
    display_name: "",
    module: "SYSTEM",
    cron_expr: "0 0 9 * * *",
    status: "ACTIVE"
  });
  definitionFormMode.value = "create";
}

function syncDefinitionDrafts() {
  const statusMap = {};
  definitions.value.forEach((item) => {
    statusMap[item.id] = item.status || "ACTIVE";
  });
  definitionStatusMap.value = statusMap;
}

function syncRunDrafts() {
  const simMap = {};
  const summaryMap = {};
  const errorMap = {};
  runs.value.forEach((item) => {
    simMap[item.id] = "";
    summaryMap[item.id] = "";
    errorMap[item.id] = "";
  });
  retrySimMap.value = simMap;
  retrySummaryMap.value = summaryMap;
  retryErrorMap.value = errorMap;
}

function cleanupPayload(raw) {
  const payload = {};
  Object.entries(raw).forEach(([key, value]) => {
    if (value === undefined || value === null || value === "") {
      return;
    }
    payload[key] = value;
  });
  return payload;
}

async function fetchMetrics() {
  metricsLoading.value = true;
  try {
    const data = await getSchedulerJobMetrics({ job_name: metricFilter.job_name.trim() });
    metrics.value = data || {};
  } catch {
    metrics.value = {
      today_total: 0,
      today_success: 0,
      today_failed: 0,
      today_running: 0
    };
  } finally {
    metricsLoading.value = false;
  }
}

async function fetchDefinitions() {
  defsLoading.value = true;
  errorMessage.value = "";
  try {
    const data = await listSchedulerJobDefinitions({
      status: definitionFilters.status,
      module: definitionFilters.module,
      page: definitionPage.value,
      page_size: definitionPageSize.value
    });
    definitions.value = data.items || [];
    definitionTotal.value = data.total || 0;
    syncDefinitionDrafts();
  } catch (error) {
    errorMessage.value = error.message || "加载任务定义失败";
  } finally {
    defsLoading.value = false;
  }
}

async function fetchRuns() {
  runsLoading.value = true;
  errorMessage.value = "";
  try {
    const data = await listSchedulerJobRuns({
      job_name: runFilters.job_name.trim(),
      status: runFilters.status,
      page: runPage.value,
      page_size: runPageSize.value
    });
    runs.value = data.items || [];
    runTotal.value = data.total || 0;
    syncRunDrafts();
  } catch (error) {
    errorMessage.value = error.message || "加载运行记录失败";
  } finally {
    runsLoading.value = false;
  }
}

async function refreshAll() {
  errorMessage.value = "";
  message.value = "";
  await Promise.all([fetchMetrics(), fetchDefinitions(), fetchRuns()]);
}

function openCreateDefinition() {
  resetDefinitionForm();
  definitionFormVisible.value = true;
}

function openEditDefinition(item) {
  Object.assign(definitionForm, {
    id: item.id,
    job_name: item.job_name || "",
    display_name: item.display_name || "",
    module: item.module || "SYSTEM",
    cron_expr: item.cron_expr || "",
    status: item.status || "ACTIVE"
  });
  definitionFormMode.value = "edit";
  definitionFormVisible.value = true;
}

async function submitDefinition() {
  const payload = {
    job_name: definitionForm.job_name.trim(),
    display_name: definitionForm.display_name.trim(),
    module: definitionForm.module,
    cron_expr: definitionForm.cron_expr.trim(),
    status: definitionForm.status
  };
  if (!payload.job_name || !payload.display_name || !payload.cron_expr) {
    errorMessage.value = "job_name/display_name/cron_expr 不能为空";
    return;
  }
  submittingDefinition.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    if (definitionFormMode.value === "create") {
      const result = await createSchedulerJobDefinition(payload);
      message.value = `任务定义已创建：${result.id || ""}`;
    } else {
      await updateSchedulerJobDefinition(definitionForm.id, payload);
      message.value = `任务定义 ${definitionForm.id} 已更新`;
    }
    definitionFormVisible.value = false;
    resetDefinitionForm();
    await fetchDefinitions();
  } catch (error) {
    errorMessage.value = error.message || "提交任务定义失败";
  } finally {
    submittingDefinition.value = false;
  }
}

async function updateDefinitionStatus(item) {
  const status = (definitionStatusMap.value[item.id] || "").trim();
  if (!status || status === item.status) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await updateSchedulerJobDefinitionStatus(item.id, status);
    message.value = `任务定义 ${item.id} 状态已更新为 ${status}`;
    await fetchDefinitions();
  } catch (error) {
    errorMessage.value = error.message || "更新任务定义状态失败";
  }
}

async function submitTrigger() {
  const payload = cleanupPayload({
    job_name: triggerForm.job_name.trim(),
    trigger_source: triggerForm.trigger_source,
    simulate_status: triggerForm.simulate_status,
    result_summary: triggerForm.result_summary.trim(),
    error_message: triggerForm.error_message.trim()
  });
  if (!payload.job_name || !payload.trigger_source) {
    errorMessage.value = "job_name 和 trigger_source 不能为空";
    return;
  }
  triggeringJob.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await triggerSchedulerJob(payload);
    message.value = `任务触发完成：run_id=${result.id || "-"} status=${result.status || "-"}`;
    await Promise.all([fetchMetrics(), fetchRuns()]);
  } catch (error) {
    errorMessage.value = error.message || "触发任务失败";
  } finally {
    triggeringJob.value = false;
  }
}

async function retryRun(runID) {
  const payload = cleanupPayload({
    simulate_status: retrySimMap.value[runID] || "",
    result_summary: (retrySummaryMap.value[runID] || "").trim(),
    error_message: (retryErrorMap.value[runID] || "").trim()
  });
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await retrySchedulerJobRun(runID, payload);
    message.value = `重跑已触发：new_run_id=${result.id || "-"} status=${result.status || "-"}`;
    await Promise.all([fetchMetrics(), fetchRuns()]);
  } catch (error) {
    errorMessage.value = error.message || "重跑任务失败";
  }
}

function applyDefinitionFilters() {
  definitionPage.value = 1;
  fetchDefinitions();
}

function resetDefinitionFilters() {
  definitionFilters.status = "";
  definitionFilters.module = "";
  definitionPage.value = 1;
  fetchDefinitions();
}

function handleDefinitionPageChange(nextPage) {
  if (nextPage === definitionPage.value) {
    return;
  }
  definitionPage.value = nextPage;
  fetchDefinitions();
}

function applyRunFilters() {
  runPage.value = 1;
  fetchRuns();
}

function resetRunFilters() {
  runFilters.job_name = "";
  runFilters.status = "";
  runPage.value = 1;
  fetchRuns();
}

function handleRunPageChange(nextPage) {
  if (nextPage === runPage.value) {
    return;
  }
  runPage.value = nextPage;
  fetchRuns();
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "SUCCESS" || normalized === "ACTIVE") return "success";
  if (normalized === "FAILED" || normalized === "DISABLED") return "danger";
  if (normalized === "RUNNING") return "warning";
  return "info";
}

onMounted(refreshAll);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">系统任务中心</h1>
        <p class="muted">管理定时任务定义、运行记录、触发与重跑</p>
      </div>
      <el-button :loading="defsLoading || runsLoading || metricsLoading" @click="refreshAll">刷新全部</el-button>
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
      <div class="toolbar">
        <el-input v-model="metricFilter.job_name" clearable placeholder="指标按 job_name 过滤（可选）" style="width: 260px" />
        <el-button :loading="metricsLoading" @click="fetchMetrics">刷新指标</el-button>
      </div>
      <div class="grid grid-4" v-loading="metricsLoading">
        <div class="metric-item">
          <div class="metric-label">今日总运行</div>
          <div class="metric-value">{{ metrics.today_total || 0 }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">今日成功</div>
          <div class="metric-value">{{ metrics.today_success || 0 }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">今日失败</div>
          <div class="metric-value">{{ metrics.today_failed || 0 }}</div>
        </div>
        <div class="metric-item">
          <div class="metric-label">今日运行中</div>
          <div class="metric-value">{{ metrics.today_running || 0 }}</div>
        </div>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <h3 style="margin: 0">手动触发任务</h3>
      </div>
      <el-form label-width="125px">
        <div class="dialog-grid">
          <el-form-item label="job_name" required>
            <el-input v-model="triggerForm.job_name" placeholder="daily_stock_recommendation" />
          </el-form-item>
          <el-form-item label="trigger_source" required>
            <el-select v-model="triggerForm.trigger_source">
              <el-option label="MANUAL" value="MANUAL" />
              <el-option label="SYSTEM" value="SYSTEM" />
            </el-select>
          </el-form-item>
          <el-form-item label="simulate_status">
            <el-select v-model="triggerForm.simulate_status" clearable placeholder="可选">
              <el-option v-for="item in simulateStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="result_summary">
            <el-input v-model="triggerForm.result_summary" placeholder="处理 120 条数据" />
          </el-form-item>
          <el-form-item label="error_message">
            <el-input v-model="triggerForm.error_message" placeholder="模拟失败信息" />
          </el-form-item>
        </div>
      </el-form>
      <div class="toolbar" style="margin-bottom: 8px">
        <el-button type="primary" :loading="triggeringJob" @click="submitTrigger">触发任务</el-button>
      </div>
      <el-alert
        title="simulate_status 仅在后端开启 ALLOW_JOB_SIMULATION=true 时生效"
        type="info"
        :closable="false"
        show-icon
      />
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <h3 style="margin: 0">任务定义</h3>
        <el-button type="primary" @click="openCreateDefinition">新增任务定义</el-button>
      </div>

      <div class="toolbar">
        <el-select v-model="definitionFilters.status" clearable placeholder="全部状态" style="width: 150px">
          <el-option v-for="item in definitionStatusOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-select v-model="definitionFilters.module" clearable placeholder="全部模块" style="width: 150px">
          <el-option v-for="item in moduleOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-button type="primary" plain @click="applyDefinitionFilters">查询</el-button>
        <el-button @click="resetDefinitionFilters">重置</el-button>
      </div>

      <el-table :data="definitions" border stripe v-loading="defsLoading" empty-text="暂无任务定义">
        <el-table-column prop="id" label="ID" min-width="120" />
        <el-table-column prop="job_name" label="job_name" min-width="180" />
        <el-table-column prop="display_name" label="display_name" min-width="150" />
        <el-table-column prop="module" label="module" min-width="100" />
        <el-table-column prop="cron_expr" label="cron_expr" min-width="180" />
        <el-table-column label="status" min-width="220">
          <template #default="{ row }">
            <div class="inline-actions inline-actions--left">
              <el-select v-model="definitionStatusMap[row.id]" size="small" style="width: 120px">
                <el-option v-for="item in definitionStatusOptions" :key="item" :label="item" :value="item" />
              </el-select>
              <el-button size="small" @click="updateDefinitionStatus(row)">保存状态</el-button>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="last_run_at" min-width="180">
          <template #default="{ row }">
            {{ row.last_run_at || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="updated_by" min-width="130">
          <template #default="{ row }">
            {{ row.updated_by || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="updated_at" min-width="180">
          <template #default="{ row }">
            {{ row.updated_at || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="操作" align="right" min-width="100">
          <template #default="{ row }">
            <el-button size="small" @click="openEditDefinition(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-text type="info">第 {{ definitionPage }} 页，共 {{ definitionTotal }} 条</el-text>
        <el-pagination
          background
          layout="prev, pager, next"
          :current-page="definitionPage"
          :page-size="definitionPageSize"
          :total="definitionTotal"
          @current-change="handleDefinitionPageChange"
        />
      </div>
    </div>

    <div class="card">
      <div class="section-header">
        <h3 style="margin: 0">运行记录</h3>
      </div>

      <div class="toolbar">
        <el-input v-model="runFilters.job_name" clearable placeholder="按 job_name 过滤" style="width: 220px" />
        <el-select v-model="runFilters.status" clearable placeholder="全部状态" style="width: 150px">
          <el-option v-for="item in runStatusOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-button type="primary" plain @click="applyRunFilters">查询</el-button>
        <el-button @click="resetRunFilters">重置</el-button>
      </div>

      <el-table :data="runs" border stripe v-loading="runsLoading" empty-text="暂无运行记录">
        <el-table-column prop="id" label="run_id" min-width="120" />
        <el-table-column label="parent_run_id" min-width="130">
          <template #default="{ row }">
            {{ row.parent_run_id || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="job_name" label="job_name" min-width="170" />
        <el-table-column label="status" min-width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="retry_count" label="retry_count" min-width="110" />
        <el-table-column prop="trigger_source" label="trigger_source" min-width="120" />
        <el-table-column prop="started_at" label="started_at" min-width="180" />
        <el-table-column label="finished_at" min-width="180">
          <template #default="{ row }">
            {{ row.finished_at || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="result_summary" min-width="180">
          <template #default="{ row }">
            {{ row.result_summary || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="error_message" min-width="180">
          <template #default="{ row }">
            {{ row.error_message || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="重跑" align="right" min-width="340">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-select v-model="retrySimMap[row.id]" size="small" clearable placeholder="simulate" style="width: 110px">
                <el-option v-for="item in simulateStatusOptions" :key="item" :label="item" :value="item" />
              </el-select>
              <el-input v-model="retrySummaryMap[row.id]" size="small" placeholder="summary" style="width: 120px" />
              <el-input v-model="retryErrorMap[row.id]" size="small" placeholder="error" style="width: 120px" />
              <el-button size="small" type="primary" @click="retryRun(row.id)">重跑</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-text type="info">第 {{ runPage }} 页，共 {{ runTotal }} 条</el-text>
        <el-pagination
          background
          layout="prev, pager, next"
          :current-page="runPage"
          :page-size="runPageSize"
          :total="runTotal"
          @current-change="handleRunPageChange"
        />
      </div>
    </div>

    <el-dialog
      v-model="definitionFormVisible"
      :title="definitionFormMode === 'create' ? '新增任务定义' : `编辑任务定义：${definitionForm.id}`"
      width="760px"
      destroy-on-close
    >
      <el-form label-width="110px">
        <div class="dialog-grid">
          <el-form-item label="job_name" required>
            <el-input v-model="definitionForm.job_name" />
          </el-form-item>
          <el-form-item label="display_name" required>
            <el-input v-model="definitionForm.display_name" />
          </el-form-item>
          <el-form-item label="module">
            <el-select v-model="definitionForm.module">
              <el-option v-for="item in moduleOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="cron_expr" required>
            <el-input v-model="definitionForm.cron_expr" placeholder="0 0 9 * * *" />
          </el-form-item>
          <el-form-item label="status">
            <el-select v-model="definitionForm.status">
              <el-option v-for="item in definitionStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="definitionFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="submittingDefinition" @click="submitDefinition">
          {{ definitionFormMode === "create" ? "创建" : "更新" }}
        </el-button>
      </template>
    </el-dialog>
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

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
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
</style>
