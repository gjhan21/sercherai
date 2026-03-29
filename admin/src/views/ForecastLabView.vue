<script setup>
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  cancelForecastRun,
  createForecastRun,
  getForecastRunDetail,
  listForecastQualitySummaries,
  listForecastRuns,
  retryForecastRun
} from "../api/admin";
import { buildForecastLabRouteQuery, normalizeForecastLabRouteState } from "../lib/forecast-lab.js";
import { sanitizeHTML } from "../lib/html";
import { hasPermission } from "../lib/session";

const route = useRoute();
const router = useRouter();

const canViewForecastLab = hasPermission("forecast_l3.view");
const canEditForecastLab = hasPermission("forecast_l3.edit");

const loading = ref(false);
const detailLoading = ref(false);
const qualityLoading = ref(false);
const creating = ref(false);
const actionLoading = ref(false);
const errorMessage = ref("");
const message = ref("");

const page = ref(1);
const pageSize = ref(12);
const total = ref(0);
const selectedRunID = ref("");
const runItems = ref([]);
const selectedRunDetail = ref(null);
const qualitySummaries = ref([]);

const runFilters = reactive({
  status: "",
  target_type: "",
  trigger_type: "",
  user_id: ""
});

const qualityFilters = reactive({
  target_type: "",
  days: 30
});

const createDialogVisible = ref(false);
const createForm = reactive({
  target_type: "STOCK",
  target_key: "",
  target_label: "",
  priority_score: 0.72,
  reason: "",
  trigger_type: "ADMIN_MANUAL"
});

const statusOptions = ["QUEUED", "RUNNING", "SUCCEEDED", "FAILED", "CANCELLED"];
const targetTypeOptions = ["STOCK", "FUTURES"];
const triggerTypeOptions = ["ADMIN_MANUAL", "AUTO_PRIORITY", "USER_REQUEST"];

const detailHTML = computed(() =>
  sanitizeHTML(selectedRunDetail.value?.report?.html_body || "<p>暂无 HTML 报告</p>")
);

const selectedRunSummary = computed(() => selectedRunDetail.value?.run?.summary || null);
const selectedRunQuality = computed(() => selectedRunDetail.value?.quality_summary || null);
const selectedRunRoleEffectiveness = computed(() =>
  Object.entries(selectedRunQuality.value?.role_effectiveness || {}).sort((left, right) => right[1] - left[1])
);

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function clearMessages() {
  errorMessage.value = "";
  message.value = "";
}

function formatDateTime(value) {
  const timestamp = Date.parse(String(value || ""));
  if (Number.isNaN(timestamp)) {
    return "-";
  }
  return new Date(timestamp).toLocaleString("zh-CN", { hour12: false });
}

function formatPercent(value) {
  const number = Number(value);
  if (!Number.isFinite(number)) {
    return "-";
  }
  return `${(number * 100).toFixed(1)}%`;
}

function formatTargetLabel(run) {
  const label = String(run?.target_label || "").trim();
  if (label) {
    return label;
  }
  return String(run?.target_key || "-").trim() || "-";
}

function statusTagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "SUCCEEDED") {
    return "success";
  }
  if (normalized === "FAILED" || normalized === "CANCELLED") {
    return "danger";
  }
  if (normalized === "RUNNING") {
    return "warning";
  }
  return "info";
}

function triggerTagType(triggerType) {
  const normalized = String(triggerType || "").toUpperCase();
  if (normalized === "ADMIN_MANUAL") {
    return "primary";
  }
  if (normalized === "AUTO_PRIORITY") {
    return "warning";
  }
  return "success";
}

async function fetchRunList() {
  if (!canViewForecastLab) {
    return;
  }
  loading.value = true;
  errorMessage.value = "";
  try {
    const data = await listForecastRuns({
      page: page.value,
      page_size: pageSize.value,
      status: runFilters.status,
      target_type: runFilters.target_type,
      trigger_type: runFilters.trigger_type,
      user_id: runFilters.user_id.trim()
    });
    runItems.value = data?.items || [];
    total.value = data?.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载深推演运行队列失败");
  } finally {
    loading.value = false;
  }
}

async function fetchRunDetail(runID = selectedRunID.value) {
  const nextRunID = String(runID || "").trim();
  if (!nextRunID || !canViewForecastLab) {
    selectedRunDetail.value = null;
    return;
  }
  detailLoading.value = true;
  try {
    const data = await getForecastRunDetail(nextRunID);
    selectedRunDetail.value = data || null;
    selectedRunID.value = nextRunID;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载深推演详情失败");
  } finally {
    detailLoading.value = false;
  }
}

async function fetchQualitySummaries() {
  if (!canViewForecastLab) {
    return;
  }
  qualityLoading.value = true;
  try {
    const data = await listForecastQualitySummaries({
      target_type: qualityFilters.target_type,
      days: qualityFilters.days
    });
    qualitySummaries.value = data?.items || [];
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载质量回写摘要失败");
  } finally {
    qualityLoading.value = false;
  }
}

async function refreshAll() {
  clearMessages();
  await Promise.all([fetchRunList(), fetchQualitySummaries()]);
  if (selectedRunID.value) {
    await fetchRunDetail(selectedRunID.value);
    return;
  }
  const firstRunID = String(runItems.value[0]?.id || "").trim();
  if (firstRunID) {
    selectedRunID.value = firstRunID;
    await fetchRunDetail(firstRunID);
    syncRouteQuery({
      runID: firstRunID,
      targetType: runItems.value[0]?.target_type || runFilters.target_type,
      status: runFilters.status,
      triggerType: runFilters.trigger_type,
      userID: runFilters.user_id
    });
  }
}

async function submitCreateRun() {
  if (!canEditForecastLab) {
    errorMessage.value = "当前账号只有查看权限，不能发起深推演任务";
    return;
  }
  if (!createForm.target_key.trim()) {
    errorMessage.value = "请填写股票代码、期货合约或统一对象键";
    return;
  }
  creating.value = true;
  clearMessages();
  try {
    const run = await createForecastRun({
      target_type: createForm.target_type,
      target_key: createForm.target_key.trim(),
      target_label: createForm.target_label.trim(),
      priority_score: Number(createForm.priority_score || 0),
      reason: createForm.reason.trim(),
      trigger_type: createForm.trigger_type
    });
    createDialogVisible.value = false;
    message.value = "深推演任务已进入队列";
    selectedRunID.value = run?.id || "";
    await Promise.all([fetchRunList(), fetchQualitySummaries()]);
    if (selectedRunID.value) {
      await fetchRunDetail(selectedRunID.value);
      syncRouteQuery({
        runID: selectedRunID.value,
        targetType: createForm.target_type,
        status: runFilters.status,
        triggerType: runFilters.trigger_type,
        userID: runFilters.user_id
      });
    }
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "创建深推演任务失败");
  } finally {
    creating.value = false;
  }
}

async function handleRetryRun(run) {
  if (!canEditForecastLab) {
    errorMessage.value = "当前账号只有查看权限，不能重跑深推演任务";
    return;
  }
  const runID = String(run?.id || selectedRunID.value || "").trim();
  if (!runID) {
    return;
  }
  const reason = window.prompt("请输入重跑原因", "人工复核后重跑深推演");
  if (reason === null) {
    return;
  }
  actionLoading.value = true;
  clearMessages();
  try {
    await retryForecastRun(runID, { reason: reason.trim() });
    message.value = "已提交重跑请求";
    await Promise.all([fetchRunList(), fetchQualitySummaries()]);
    await fetchRunDetail(runID);
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "重跑深推演任务失败");
  } finally {
    actionLoading.value = false;
  }
}

async function handleCancelRun(run) {
  if (!canEditForecastLab) {
    errorMessage.value = "当前账号只有查看权限，不能取消深推演任务";
    return;
  }
  const runID = String(run?.id || selectedRunID.value || "").trim();
  if (!runID) {
    return;
  }
  const reason = window.prompt("请输入取消原因", "人工取消深推演任务");
  if (reason === null || !reason.trim()) {
    return;
  }
  actionLoading.value = true;
  clearMessages();
  try {
    await cancelForecastRun(runID, { reason: reason.trim() });
    message.value = "任务已取消";
    await Promise.all([fetchRunList(), fetchQualitySummaries()]);
    await fetchRunDetail(runID);
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "取消深推演任务失败");
  } finally {
    actionLoading.value = false;
  }
}

function handleOpenCreateDialog(targetType = "") {
  createForm.target_type = targetTypeOptions.includes(targetType) ? targetType : createForm.target_type;
  createDialogVisible.value = true;
}

function handlePickRun(row) {
  const runID = String(row?.id || "").trim();
  if (!runID) {
    return;
  }
  selectedRunID.value = runID;
  syncRouteQuery({
    runID: runID,
    targetType: row?.target_type || runFilters.target_type,
    status: runFilters.status,
    triggerType: runFilters.trigger_type,
    userID: runFilters.user_id
  });
}

function handleRunPageChange(nextPage) {
  page.value = nextPage;
  fetchRunList();
}

function resetRunFilters() {
  runFilters.status = "";
  runFilters.target_type = "";
  runFilters.trigger_type = "";
  runFilters.user_id = "";
  page.value = 1;
  syncRouteQuery({ runID: selectedRunID.value || "" });
  fetchRunList();
}

function applyRunFilters() {
  page.value = 1;
  syncRouteQuery({
    runID: selectedRunID.value || "",
    targetType: runFilters.target_type,
    status: runFilters.status,
    triggerType: runFilters.trigger_type,
    userID: runFilters.user_id
  });
  fetchRunList();
}

function applyRouteState(query = {}) {
  const state = normalizeForecastLabRouteState(query);
  if (targetTypeOptions.includes(state.targetType)) {
    runFilters.target_type = state.targetType;
    qualityFilters.target_type = state.targetType;
    createForm.target_type = state.targetType;
  }
  if (statusOptions.includes(state.status)) {
    runFilters.status = state.status;
  }
  if (triggerTypeOptions.includes(state.triggerType)) {
    runFilters.trigger_type = state.triggerType;
  }
  runFilters.user_id = state.userID;
  if (state.runID) {
    selectedRunID.value = state.runID;
  }
}

function syncRouteQuery(nextState = {}) {
  router.replace({
    query: buildForecastLabRouteQuery(nextState)
  });
}

watch(
  () => route.query.run_id,
  (runID) => {
    const normalized = String(runID || "").trim();
    if (!normalized) {
      return;
    }
    if (normalized === selectedRunID.value && selectedRunDetail.value?.run?.id === normalized) {
      return;
    }
    fetchRunDetail(normalized);
  },
  { immediate: true }
);

watch(
  () => route.query.target_type,
  (targetType) => {
    const normalized = normalizeForecastLabRouteState({ target_type: targetType }).targetType;
    if (!targetTypeOptions.includes(normalized)) {
      return;
    }
    runFilters.target_type = normalized;
    qualityFilters.target_type = normalized;
    createForm.target_type = normalized;
  },
  { immediate: true }
);

watch(
  () => route.query.status,
  (status) => {
    const normalized = normalizeForecastLabRouteState({ status }).status;
    if (!statusOptions.includes(normalized)) {
      if (!String(status || "").trim()) {
        runFilters.status = "";
      }
      return;
    }
    runFilters.status = normalized;
  },
  { immediate: true }
);

watch(
  () => route.query.trigger_type,
  (triggerType) => {
    const normalized = normalizeForecastLabRouteState({ trigger_type: triggerType }).triggerType;
    if (!triggerTypeOptions.includes(normalized)) {
      if (!String(triggerType || "").trim()) {
        runFilters.trigger_type = "";
      }
      return;
    }
    runFilters.trigger_type = normalized;
  },
  { immediate: true }
);

watch(
  () => route.query.user_id,
  (userID) => {
    runFilters.user_id = String(userID || "").trim();
  },
  { immediate: true }
);

onMounted(async () => {
  if (!canViewForecastLab) {
    errorMessage.value = "当前账号没有 Forecast Lab 查看权限";
    return;
  }
  applyRouteState(route.query || {});
  await refreshAll();
});
</script>

<template>
  <div class="page-shell forecast-lab-page">
    <div class="page-header">
      <div>
        <h2>Forecast Lab</h2>
        <p class="muted">
          这里集中处理 L3 深推演的排队、运行、报告、质量回写与人工复核。主策略链继续留在策略中心，Forecast Lab 只负责增强层。
        </p>
      </div>
      <div class="toolbar">
        <el-button @click="refreshAll">刷新工作台</el-button>
        <el-button v-if="canEditForecastLab" type="primary" @click="handleOpenCreateDialog(runFilters.target_type)">
          管理员手动触发
        </el-button>
      </div>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      :closable="false"
      show-icon
      style="margin-bottom: 12px"
    />
    <el-alert
      v-else-if="message"
      :title="message"
      type="success"
      :closable="false"
      show-icon
      style="margin-bottom: 12px"
    />

    <div class="forecast-lab-grid">
      <section class="lab-main">
        <div class="card" style="margin-bottom: 12px">
          <div class="section-head">
            <div class="section-title">深推演运行队列</div>
            <div class="toolbar" style="margin-bottom: 0">
              <el-select v-model="runFilters.target_type" clearable placeholder="对象类型" style="width: 120px">
                <el-option v-for="item in targetTypeOptions" :key="item" :label="item" :value="item" />
              </el-select>
              <el-select v-model="runFilters.status" clearable placeholder="运行状态" style="width: 130px">
                <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
              </el-select>
              <el-select v-model="runFilters.trigger_type" clearable placeholder="触发方式" style="width: 150px">
                <el-option v-for="item in triggerTypeOptions" :key="item" :label="item" :value="item" />
              </el-select>
              <el-input v-model="runFilters.user_id" clearable placeholder="请求用户ID" style="width: 160px" />
              <el-button type="primary" plain @click="applyRunFilters">查询</el-button>
              <el-button @click="resetRunFilters">重置</el-button>
            </div>
          </div>

          <el-table
            :data="runItems"
            border
            stripe
            highlight-current-row
            v-loading="loading"
            empty-text="暂无深推演任务"
            @row-click="handlePickRun"
          >
            <el-table-column prop="target_type" label="对象" min-width="88">
              <template #default="{ row }">
                <el-tag :type="row.target_type === 'FUTURES' ? 'warning' : 'primary'">{{ row.target_type }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="标的" min-width="180">
              <template #default="{ row }">
                <div class="forecast-run-target">{{ formatTargetLabel(row) }}</div>
                <div class="muted">{{ row.target_key || "-" }}</div>
              </template>
            </el-table-column>
            <el-table-column label="状态" min-width="110">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ row.status || "-" }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="触发方式" min-width="120">
              <template #default="{ row }">
                <el-tag :type="triggerTagType(row.trigger_type)" effect="plain">{{ row.trigger_type || "-" }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="优先级" min-width="96">
              <template #default="{ row }">{{ Number(row.priority_score || 0).toFixed(2) }}</template>
            </el-table-column>
            <el-table-column label="更新时间" min-width="168">
              <template #default="{ row }">{{ formatDateTime(row.updated_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" min-width="190" fixed="right">
              <template #default="{ row }">
                <div class="table-actions">
                  <el-button size="small" @click.stop="handlePickRun(row)">查看详情</el-button>
                  <el-button
                    v-if="canEditForecastLab"
                    size="small"
                    type="primary"
                    plain
                    :loading="actionLoading"
                    @click.stop="handleRetryRun(row)"
                  >
                    重跑
                  </el-button>
                  <el-button
                    v-if="canEditForecastLab && ['QUEUED', 'RUNNING'].includes(row.status)"
                    size="small"
                    type="danger"
                    plain
                    :loading="actionLoading"
                    @click.stop="handleCancelRun(row)"
                  >
                    取消
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
              @current-change="handleRunPageChange"
            />
          </div>
        </div>

        <div class="card" v-loading="detailLoading">
          <div class="section-head">
            <div class="section-title">报告详情</div>
            <div class="toolbar" style="margin-bottom: 0">
              <el-tag v-if="selectedRunDetail?.run" :type="statusTagType(selectedRunDetail.run.status)">
                {{ selectedRunDetail.run.status }}
              </el-tag>
              <el-button v-if="selectedRunID" size="small" @click="fetchRunDetail(selectedRunID)">刷新详情</el-button>
            </div>
          </div>

          <div v-if="selectedRunDetail?.run" class="detail-stack">
            <div class="detail-kpi-grid">
              <div class="detail-kpi-card">
                <span class="detail-kpi-label">对象</span>
                <strong>{{ formatTargetLabel(selectedRunDetail.run) }}</strong>
                <div class="muted">{{ selectedRunDetail.run.target_type }} · {{ selectedRunDetail.run.target_key }}</div>
              </div>
              <div class="detail-kpi-card">
                <span class="detail-kpi-label">主情景</span>
                <strong>{{ selectedRunSummary?.primary_scenario || selectedRunSummary?.primaryScenario || "-" }}</strong>
                <div class="muted">引擎 {{ selectedRunDetail.run.engine_key || "-" }}</div>
              </div>
              <div class="detail-kpi-card">
                <span class="detail-kpi-label">动作建议</span>
                <strong>{{ selectedRunSummary?.action_guidance || selectedRunSummary?.actionGuidance || "-" }}</strong>
                <div class="muted">生成于 {{ formatDateTime(selectedRunDetail.run.finished_at || selectedRunDetail.run.updated_at) }}</div>
              </div>
            </div>

            <div class="detail-meta-grid">
              <div class="meta-card">
                <h4>运行摘要</h4>
                <p>{{ selectedRunSummary?.executive_summary || selectedRunSummary?.executiveSummary || selectedRunDetail.run.reason || "暂无摘要" }}</p>
              </div>
              <div class="meta-card">
                <h4>质量回写摘要</h4>
                <div v-if="selectedRunQuality">
                  <p>情景命中率：{{ formatPercent(selectedRunQuality.scenario_hit_rate || selectedRunQuality.scenarioHitRate) }}</p>
                  <p>触发命中率：{{ formatPercent(selectedRunQuality.trigger_hit_rate || selectedRunQuality.triggerHitRate) }}</p>
                  <p>提前失效识别率：{{ formatPercent(selectedRunQuality.invalidation_early_rate || selectedRunQuality.invalidationEarlyRate) }}</p>
                </div>
                <p v-else class="muted">当前运行还没有质量回写记录。</p>
              </div>
            </div>

            <div class="card-section">
              <h4>执行摘要</h4>
              <p>{{ selectedRunDetail.report?.executive_summary || "暂无执行摘要" }}</p>
            </div>

            <div class="card-section">
              <h4>主情景与备选情景</h4>
              <div class="scenario-grid">
                <div class="scenario-card">
                  <span class="detail-kpi-label">主情景</span>
                  <strong>{{ selectedRunDetail.report?.primary_scenario || "-" }}</strong>
                </div>
                <div
                  v-for="item in selectedRunDetail.report?.alternative_scenarios || []"
                  :key="`${item.name}-${item.action}`"
                  class="scenario-card"
                >
                  <span class="detail-kpi-label">{{ item.name || "备选情景" }}</span>
                  <strong>{{ item.action || "-" }}</strong>
                  <div class="muted">{{ item.thesis || "-" }}</div>
                </div>
              </div>
            </div>

            <div class="card-section">
              <h4>触发 checklist</h4>
              <el-table
                :data="selectedRunDetail.report?.trigger_checklist || []"
                size="small"
                border
                empty-text="暂无触发清单"
              >
                <el-table-column prop="label" label="检查项" min-width="180" />
                <el-table-column prop="status" label="状态" min-width="100" />
                <el-table-column prop="trigger" label="触发条件" min-width="180" />
                <el-table-column prop="note" label="备注" min-width="220" />
              </el-table>
            </div>

            <div class="card-section">
              <h4>失效信号与角色分歧</h4>
              <div class="signal-list">
                <el-tag
                  v-for="item in selectedRunDetail.report?.invalidation_signals || []"
                  :key="item"
                  type="danger"
                  effect="plain"
                >
                  {{ item }}
                </el-tag>
                <span v-if="!(selectedRunDetail.report?.invalidation_signals || []).length" class="muted">暂无失效信号</span>
              </div>
              <el-table
                :data="selectedRunDetail.report?.role_disagreements || []"
                size="small"
                border
                empty-text="暂无角色分歧"
                style="margin-top: 10px"
              >
                <el-table-column prop="role" label="角色" min-width="120" />
                <el-table-column prop="stance" label="立场" min-width="120" />
                <el-table-column prop="summary" label="分歧摘要" min-width="220" />
                <el-table-column label="Veto" min-width="80">
                  <template #default="{ row }">
                    <el-tag :type="row.veto ? 'danger' : 'info'">{{ row.veto ? "是" : "否" }}</el-tag>
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <div class="card-section">
              <h4>运行步骤日志</h4>
              <el-timeline>
                <el-timeline-item
                  v-for="item in selectedRunDetail.logs || []"
                  :key="item.id"
                  :timestamp="formatDateTime(item.created_at)"
                  :type="statusTagType(item.status)"
                >
                  <div class="log-item">
                    <strong>{{ item.step_key }}</strong>
                    <div>{{ item.message || "无步骤说明" }}</div>
                    <pre class="detail-pre">{{ JSON.stringify(item.payload || {}, null, 2) }}</pre>
                  </div>
                </el-timeline-item>
              </el-timeline>
            </div>

            <div class="card-section">
              <h4>报告正文</h4>
              <div class="detail-html" v-html="detailHTML" />
              <pre class="detail-pre">{{ selectedRunDetail.report?.markdown_body || "暂无 Markdown 正文" }}</pre>
            </div>
          </div>

          <el-empty v-else description="选择一条深推演任务后，这里会显示报告、质量回写和步骤日志。" />
        </div>
      </section>

      <aside class="lab-side">
        <div class="card" style="margin-bottom: 12px" v-loading="qualityLoading">
          <div class="section-head">
            <div class="section-title">质量回写摘要</div>
            <div class="toolbar" style="margin-bottom: 0">
              <el-select v-model="qualityFilters.target_type" clearable placeholder="对象类型" style="width: 120px">
                <el-option v-for="item in targetTypeOptions" :key="item" :label="item" :value="item" />
              </el-select>
              <el-input-number
                v-model="qualityFilters.days"
                :min="1"
                :max="365"
                :step="1"
                controls-position="right"
                style="width: 120px"
              />
              <el-button type="primary" plain @click="fetchQualitySummaries">刷新</el-button>
            </div>
          </div>

          <div v-if="qualitySummaries.length" class="quality-card-list">
            <div v-for="item in qualitySummaries" :key="item.target_type" class="quality-card">
              <div class="quality-card-head">
                <strong>{{ item.target_type }}</strong>
                <el-tag effect="plain">{{ item.succeeded_runs || item.succeededRuns || 0 }} / {{ item.total_runs || item.totalRuns || 0 }}</el-tag>
              </div>
              <p>情景命中率：{{ formatPercent(item.scenario_hit_rate || item.scenarioHitRate) }}</p>
              <p>触发命中率：{{ formatPercent(item.trigger_hit_rate || item.triggerHitRate) }}</p>
              <p>提前失效识别率：{{ formatPercent(item.invalidation_early_rate || item.invalidationEarlyRate) }}</p>
              <p class="muted">最近回写：{{ formatDateTime(item.last_learning_recorded_at || item.lastLearningRecordedAt) }}</p>
            </div>
          </div>
          <el-empty v-else description="暂无质量回写摘要" />
        </div>

        <div class="card" style="margin-bottom: 12px">
          <div class="section-title" style="margin-bottom: 12px">角色有效性</div>
          <div v-if="selectedRunRoleEffectiveness.length" class="role-score-list">
            <div v-for="[role, score] in selectedRunRoleEffectiveness" :key="role" class="role-score-row">
              <span>{{ role }}</span>
              <strong>{{ formatPercent(score) }}</strong>
            </div>
          </div>
          <el-empty v-else description="当前运行还没有角色有效性统计" />
        </div>

        <div class="card">
          <div class="section-title" style="margin-bottom: 12px">使用说明</div>
          <ul class="usage-list">
            <li>主链不变：Forecast Lab 只负责 L3 深推演，不改推荐排序和发布审核主流程。</li>
            <li>管理员手动触发适合高价值样本复核；用户主动请求是否开放，由配置中心控制。</li>
            <li>失败任务优先看“运行步骤日志”，再决定是重跑还是取消，避免盲目重复触发。</li>
          </ul>
        </div>
      </aside>
    </div>

    <el-dialog v-model="createDialogVisible" title="创建 L3 深推演任务" width="620px" destroy-on-close>
      <el-form label-width="120px">
        <el-form-item label="对象类型" required>
          <el-segmented v-model="createForm.target_type" :options="targetTypeOptions" />
        </el-form-item>
        <el-form-item label="对象键" required>
          <el-input v-model="createForm.target_key" placeholder="如 600519.SH / RB2601 / CU2409" />
        </el-form-item>
        <el-form-item label="展示名称">
          <el-input v-model="createForm.target_label" placeholder="可选，给运营和审核阅读使用" />
        </el-form-item>
        <el-form-item label="优先级分数">
          <el-input-number
            v-model="createForm.priority_score"
            :min="0"
            :max="1"
            :step="0.01"
            :precision="2"
            controls-position="right"
          />
        </el-form-item>
        <el-form-item label="触发方式">
          <el-select v-model="createForm.trigger_type" style="width: 100%">
            <el-option v-for="item in triggerTypeOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="触发原因">
          <el-input
            v-model="createForm.reason"
            type="textarea"
            :rows="4"
            placeholder="说明为什么现在需要做深推演，例如研报分歧大、版本变化快、需要补充情景推演。"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <div class="toolbar" style="margin-bottom: 0; justify-content: flex-end">
          <el-button @click="createDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="creating" @click="submitCreateRun">提交排队</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.forecast-lab-page {
  padding: 20px;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.page-header h2 {
  margin: 0 0 6px;
  font-size: 24px;
}

.forecast-lab-grid {
  display: grid;
  grid-template-columns: minmax(0, 1.7fr) minmax(320px, 0.9fr);
  gap: 16px;
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.section-title {
  font-size: 16px;
  font-weight: 700;
}

.toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.forecast-run-target {
  font-weight: 600;
}

.table-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.pagination {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-top: 12px;
}

.detail-stack {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.detail-kpi-grid,
.detail-meta-grid,
.scenario-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
}

.detail-kpi-card,
.meta-card,
.scenario-card,
.quality-card {
  padding: 14px;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
}

.detail-kpi-label {
  display: block;
  margin-bottom: 8px;
  color: #6b7280;
  font-size: 12px;
}

.card-section {
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  padding: 14px;
  background: #ffffff;
}

.card-section h4,
.meta-card h4 {
  margin: 0 0 10px;
}

.signal-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.detail-pre {
  margin: 10px 0 0;
  padding: 12px;
  border-radius: 12px;
  background: #0f172a;
  color: #e2e8f0;
  overflow: auto;
  font-size: 12px;
  line-height: 1.55;
}

.detail-html {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 14px;
  background: #ffffff;
}

.quality-card-list,
.role-score-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.quality-card-head,
.role-score-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.usage-list {
  margin: 0;
  padding-left: 18px;
  color: #475569;
  line-height: 1.7;
}

.muted {
  color: #6b7280;
}

@media (max-width: 1200px) {
  .forecast-lab-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .page-header,
  .section-head,
  .pagination {
    flex-direction: column;
    align-items: stretch;
  }

  .detail-kpi-grid,
  .detail-meta-grid,
  .scenario-grid {
    grid-template-columns: 1fr;
  }
}
</style>
