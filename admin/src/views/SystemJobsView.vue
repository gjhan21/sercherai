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

function nextDefinitionPage() {
  if (definitionPage.value * definitionPageSize.value >= definitionTotal.value) {
    return;
  }
  definitionPage.value += 1;
  fetchDefinitions();
}

function prevDefinitionPage() {
  if (definitionPage.value <= 1) {
    return;
  }
  definitionPage.value -= 1;
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

function nextRunPage() {
  if (runPage.value * runPageSize.value >= runTotal.value) {
    return;
  }
  runPage.value += 1;
  fetchRuns();
}

function prevRunPage() {
  if (runPage.value <= 1) {
    return;
  }
  runPage.value -= 1;
  fetchRuns();
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
      <button class="btn" :disabled="defsLoading || runsLoading || metricsLoading" @click="refreshAll">
        {{ defsLoading || runsLoading || metricsLoading ? "刷新中..." : "刷新全部" }}
      </button>
    </div>

    <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
    <div v-if="message" class="success-message">{{ message }}</div>

    <div class="card" style="margin-bottom: 12px">
      <div class="toolbar">
        <input v-model="metricFilter.job_name" class="input" placeholder="指标按 job_name 过滤（可选）" />
        <button class="btn" :disabled="metricsLoading" @click="fetchMetrics">
          {{ metricsLoading ? "加载中..." : "刷新指标" }}
        </button>
      </div>
      <div class="grid grid-4">
        <div class="metric-card">
          <div class="label">今日总运行</div>
          <div class="value">{{ metrics.today_total || 0 }}</div>
        </div>
        <div class="metric-card">
          <div class="label">今日成功</div>
          <div class="value">{{ metrics.today_success || 0 }}</div>
        </div>
        <div class="metric-card">
          <div class="label">今日失败</div>
          <div class="value">{{ metrics.today_failed || 0 }}</div>
        </div>
        <div class="metric-card">
          <div class="label">今日运行中</div>
          <div class="value">{{ metrics.today_running || 0 }}</div>
        </div>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <h3 style="margin-top: 0">手动触发任务</h3>
      <div class="form-grid">
        <div class="form-item">
          <label>job_name</label>
          <input v-model="triggerForm.job_name" class="input" placeholder="daily_stock_recommendation" />
        </div>
        <div class="form-item">
          <label>trigger_source</label>
          <select v-model="triggerForm.trigger_source" class="select">
            <option value="MANUAL">MANUAL</option>
            <option value="SYSTEM">SYSTEM</option>
          </select>
        </div>
        <div class="form-item">
          <label>simulate_status（可选）</label>
          <select v-model="triggerForm.simulate_status" class="select">
            <option value="">(空)</option>
            <option value="SUCCESS">SUCCESS</option>
            <option value="FAILED">FAILED</option>
          </select>
        </div>
        <div class="form-item">
          <label>result_summary（可选）</label>
          <input v-model="triggerForm.result_summary" class="input" placeholder="处理 120 条数据" />
        </div>
        <div class="form-item">
          <label>error_message（可选）</label>
          <input v-model="triggerForm.error_message" class="input" placeholder="模拟失败信息" />
        </div>
      </div>
      <div class="form-actions">
        <button class="btn btn-primary" :disabled="triggeringJob" @click="submitTrigger">
          {{ triggeringJob ? "触发中..." : "触发任务" }}
        </button>
      </div>
      <p class="hint">
        `simulate_status` 仅在后端开启 `ALLOW_JOB_SIMULATION=true` 时生效。
      </p>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="page-header" style="margin-bottom: 10px">
        <h3 style="margin: 0">任务定义</h3>
        <button class="btn btn-primary" @click="openCreateDefinition">新增任务定义</button>
      </div>

      <div class="toolbar">
        <select v-model="definitionFilters.status" class="select">
          <option value="">全部状态</option>
          <option value="ACTIVE">ACTIVE</option>
          <option value="DISABLED">DISABLED</option>
        </select>
        <select v-model="definitionFilters.module" class="select">
          <option value="">全部模块</option>
          <option value="STOCK">STOCK</option>
          <option value="FUTURES">FUTURES</option>
          <option value="NEWS">NEWS</option>
          <option value="SYSTEM">SYSTEM</option>
        </select>
        <button class="btn" @click="applyDefinitionFilters">查询</button>
        <button class="btn" @click="resetDefinitionFilters">重置</button>
      </div>

      <div v-if="definitionFormVisible" class="card" style="margin: 12px 0">
        <div class="form-grid">
          <div class="form-item">
            <label>job_name</label>
            <input v-model="definitionForm.job_name" class="input" />
          </div>
          <div class="form-item">
            <label>display_name</label>
            <input v-model="definitionForm.display_name" class="input" />
          </div>
          <div class="form-item">
            <label>module</label>
            <select v-model="definitionForm.module" class="select">
              <option value="STOCK">STOCK</option>
              <option value="FUTURES">FUTURES</option>
              <option value="NEWS">NEWS</option>
              <option value="SYSTEM">SYSTEM</option>
            </select>
          </div>
          <div class="form-item">
            <label>cron_expr</label>
            <input v-model="definitionForm.cron_expr" class="input" placeholder="0 0 9 * * *" />
          </div>
          <div class="form-item">
            <label>status</label>
            <select v-model="definitionForm.status" class="select">
              <option value="ACTIVE">ACTIVE</option>
              <option value="DISABLED">DISABLED</option>
            </select>
          </div>
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" :disabled="submittingDefinition" @click="submitDefinition">
            {{ submittingDefinition ? "提交中..." : definitionFormMode === "create" ? "创建" : "更新" }}
          </button>
          <button class="btn" :disabled="submittingDefinition" @click="definitionFormVisible = false">取消</button>
        </div>
      </div>

      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>job_name</th>
              <th>display_name</th>
              <th>module</th>
              <th>cron_expr</th>
              <th>status</th>
              <th>last_run_at</th>
              <th>updated_by</th>
              <th>updated_at</th>
              <th class="text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in definitions" :key="item.id">
              <td>{{ item.id }}</td>
              <td>{{ item.job_name }}</td>
              <td>{{ item.display_name }}</td>
              <td>{{ item.module }}</td>
              <td>{{ item.cron_expr }}</td>
              <td>
                <div class="toolbar">
                  <select v-model="definitionStatusMap[item.id]" class="select">
                    <option value="ACTIVE">ACTIVE</option>
                    <option value="DISABLED">DISABLED</option>
                  </select>
                  <button class="btn" @click="updateDefinitionStatus(item)">保存状态</button>
                </div>
              </td>
              <td>{{ item.last_run_at || "-" }}</td>
              <td>{{ item.updated_by || "-" }}</td>
              <td>{{ item.updated_at || "-" }}</td>
              <td class="text-right">
                <button class="btn" @click="openEditDefinition(item)">编辑</button>
              </td>
            </tr>
            <tr v-if="!defsLoading && definitions.length === 0">
              <td colspan="10" class="muted">暂无任务定义</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination">
        <span>第 {{ definitionPage }} 页，共 {{ definitionTotal }} 条</span>
        <div class="toolbar">
          <button class="btn" :disabled="definitionPage <= 1 || defsLoading" @click="prevDefinitionPage">
            上一页
          </button>
          <button
            class="btn"
            :disabled="definitionPage * definitionPageSize >= definitionTotal || defsLoading"
            @click="nextDefinitionPage"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <div class="card">
      <h3 style="margin-top: 0">运行记录</h3>
      <div class="toolbar">
        <input v-model="runFilters.job_name" class="input" placeholder="按 job_name 过滤" />
        <select v-model="runFilters.status" class="select">
          <option value="">全部状态</option>
          <option value="RUNNING">RUNNING</option>
          <option value="SUCCESS">SUCCESS</option>
          <option value="FAILED">FAILED</option>
        </select>
        <button class="btn" @click="applyRunFilters">查询</button>
        <button class="btn" @click="resetRunFilters">重置</button>
      </div>

      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>run_id</th>
              <th>parent_run_id</th>
              <th>job_name</th>
              <th>status</th>
              <th>retry_count</th>
              <th>trigger_source</th>
              <th>started_at</th>
              <th>finished_at</th>
              <th>result_summary</th>
              <th>error_message</th>
              <th class="text-right">重跑</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in runs" :key="item.id">
              <td>{{ item.id }}</td>
              <td>{{ item.parent_run_id || "-" }}</td>
              <td>{{ item.job_name }}</td>
              <td>{{ item.status }}</td>
              <td>{{ item.retry_count }}</td>
              <td>{{ item.trigger_source }}</td>
              <td>{{ item.started_at }}</td>
              <td>{{ item.finished_at || "-" }}</td>
              <td>{{ item.result_summary || "-" }}</td>
              <td>{{ item.error_message || "-" }}</td>
              <td class="text-right">
                <div class="toolbar" style="justify-content: flex-end">
                  <select v-model="retrySimMap[item.id]" class="select">
                    <option value="">(空)</option>
                    <option value="SUCCESS">SUCCESS</option>
                    <option value="FAILED">FAILED</option>
                  </select>
                  <input v-model="retrySummaryMap[item.id]" class="input" style="width: 130px" placeholder="summary" />
                  <input v-model="retryErrorMap[item.id]" class="input" style="width: 130px" placeholder="error" />
                  <button class="btn btn-primary" @click="retryRun(item.id)">重跑</button>
                </div>
              </td>
            </tr>
            <tr v-if="!runsLoading && runs.length === 0">
              <td colspan="11" class="muted">暂无运行记录</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination">
        <span>第 {{ runPage }} 页，共 {{ runTotal }} 条</span>
        <div class="toolbar">
          <button class="btn" :disabled="runPage <= 1 || runsLoading" @click="prevRunPage">上一页</button>
          <button class="btn" :disabled="runPage * runPageSize >= runTotal || runsLoading" @click="nextRunPage">
            下一页
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
