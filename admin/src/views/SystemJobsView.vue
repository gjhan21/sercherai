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
import { getAccessToken } from "../lib/session";

const metricsLoading = ref(false);
const defsLoading = ref(false);
const runsLoading = ref(false);
const exportingRuns = ref(false);
const submittingDefinition = ref(false);
const triggeringJob = ref(false);
const copyingRunText = ref(false);

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
const runDetailVisible = ref(false);
const currentRun = ref(null);

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

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function csvEscape(value) {
  const text = String(value ?? "");
  if (/[",\n]/.test(text)) {
    return `"${text.replace(/"/g, "\"\"")}"`;
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

function buildRunCSVRows(items) {
  const header = [
    "id",
    "parent_run_id",
    "job_name",
    "status",
    "retry_count",
    "trigger_source",
    "started_at",
    "finished_at",
    "result_summary",
    "error_message",
    "operator_id"
  ];
  const rows = items.map((item) => [
    item.id || "",
    item.parent_run_id || "",
    item.job_name || "",
    item.status || "",
    item.retry_count ?? "",
    item.trigger_source || "",
    item.started_at || "",
    item.finished_at || "",
    item.result_summary || "",
    item.error_message || "",
    item.operator_id || ""
  ]);
  return [header, ...rows].map((row) => row.map(csvEscape).join(",")).join("\n");
}

function exportRunCurrentPageCSV() {
  const csv = buildRunCSVRows(runs.value);
  const fileName = `system_job_runs_page_${new Date().toISOString().slice(0, 10)}.csv`;
  triggerCSVDownload(csv, fileName);
  message.value = `已导出当前页运行记录 CSV，共 ${runs.value.length} 条`;
}

async function exportRunFilteredCSV() {
  exportingRuns.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const params = new URLSearchParams();
    if (runFilters.job_name.trim()) params.set("job_name", runFilters.job_name.trim());
    if (runFilters.status) params.set("status", runFilters.status);
    const baseURL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
    const query = params.toString();
    const requestURL = `${baseURL}/admin/system/job-runs/export.csv${query ? `?${query}` : ""}`;
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
    const fileName = `system_job_runs_filtered_${new Date().toISOString().slice(0, 10)}.csv`;
    const anchor = document.createElement("a");
    anchor.href = blobURL;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(blobURL);
    message.value = "已发起筛选运行记录 CSV 下载";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "导出运行记录失败");
  } finally {
    exportingRuns.value = false;
  }
}

function openRunDetail(item) {
  currentRun.value = item;
  runDetailVisible.value = true;
}

async function copyRunField(label, value) {
  const text = String(value || "").trim();
  if (!text) {
    errorMessage.value = `${label}为空，无法复制`;
    return;
  }
  copyingRunText.value = true;
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
    message.value = `${label}已复制`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, `复制${label}失败`);
  } finally {
    copyingRunText.value = false;
  }
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
        <el-button :loading="exportingRuns" @click="exportRunFilteredCSV">导出筛选CSV</el-button>
        <el-button @click="exportRunCurrentPageCSV">导出当前页CSV</el-button>
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
            <span class="run-preview">{{ row.result_summary || "-" }}</span>
          </template>
        </el-table-column>
        <el-table-column label="error_message" min-width="180">
          <template #default="{ row }">
            <span class="run-preview">{{ row.error_message || "-" }}</span>
          </template>
        </el-table-column>
        <el-table-column label="重跑" align="right" min-width="430">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-button size="small" @click="openRunDetail(row)">详情</el-button>
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

    <el-drawer v-model="runDetailVisible" size="620px" destroy-on-close>
      <template #header>
        <div class="drawer-title">任务运行详情</div>
      </template>
      <template v-if="currentRun">
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="run_id">{{ currentRun.id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="parent_run_id">{{ currentRun.parent_run_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="job_name">{{ currentRun.job_name || "-" }}</el-descriptions-item>
          <el-descriptions-item label="status">
            <el-tag :type="statusTagType(currentRun.status)">{{ currentRun.status || "-" }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="trigger_source">{{ currentRun.trigger_source || "-" }}</el-descriptions-item>
          <el-descriptions-item label="retry_count">{{ currentRun.retry_count ?? "-" }}</el-descriptions-item>
          <el-descriptions-item label="started_at">{{ currentRun.started_at || "-" }}</el-descriptions-item>
          <el-descriptions-item label="finished_at">{{ currentRun.finished_at || "-" }}</el-descriptions-item>
          <el-descriptions-item label="operator_id">{{ currentRun.operator_id || "-" }}</el-descriptions-item>
        </el-descriptions>

        <div class="run-detail-block">
          <div class="run-detail-title-row">
            <h4>result_summary</h4>
            <el-button
              link
              type="primary"
              :disabled="!(currentRun.result_summary || '').trim()"
              :loading="copyingRunText"
              @click="copyRunField('result_summary', currentRun.result_summary)"
            >
              复制
            </el-button>
          </div>
          <pre class="run-detail-pre">{{ currentRun.result_summary || "-" }}</pre>
        </div>

        <div class="run-detail-block">
          <div class="run-detail-title-row">
            <h4>error_message</h4>
            <el-button
              link
              type="primary"
              :disabled="!(currentRun.error_message || '').trim()"
              :loading="copyingRunText"
              @click="copyRunField('error_message', currentRun.error_message)"
            >
              复制
            </el-button>
          </div>
          <pre class="run-detail-pre">{{ currentRun.error_message || "-" }}</pre>
        </div>
      </template>
    </el-drawer>

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

.run-preview {
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

.run-detail-block {
  margin-top: 14px;
}

.run-detail-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
}

.run-detail-title-row h4 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.run-detail-pre {
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
