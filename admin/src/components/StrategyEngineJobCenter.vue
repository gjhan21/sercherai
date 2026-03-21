<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { getStrategyEngineJob, listStrategyEngineJobs, publishStrategyEngineJob } from "../api/admin";
import { hasPermission } from "../lib/session";

const loading = ref(false);
const detailLoading = ref(false);
const detailDialogVisible = ref(false);
const detailTab = ref("overview");
const publishSubmitting = ref(false);
const jobs = ref([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(10);
const errorMessage = ref("");
const detail = ref(null);
const canEditMarket = hasPermission("market.edit");

const filters = reactive({
  job_type: "",
  status: ""
});

const statusTagTypeMap = {
  QUEUED: "info",
  RUNNING: "warning",
  SUCCEEDED: "success",
  FAILED: "danger"
};

const riskRankMap = {
  LOW: 1,
  MEDIUM: 2,
  HIGH: 3
};

const configRefOrder = [
  { key: "seed_set", label: "种子集" },
  { key: "agent_profile", label: "角色配置" },
  { key: "scenario_template", label: "场景模板" },
  { key: "publish_policy", label: "发布策略" }
];

const statusOptions = [
  { label: "全部状态", value: "" },
  { label: "排队中", value: "QUEUED" },
  { label: "运行中", value: "RUNNING" },
  { label: "成功", value: "SUCCEEDED" },
  { label: "失败", value: "FAILED" }
];

const jobTypeOptions = [
  { label: "全部任务", value: "" },
  { label: "股票选股", value: "stock-selection" },
  { label: "期货策略", value: "futures-strategy" }
];

const statusSummary = computed(() => {
  const counts = {
    QUEUED: 0,
    RUNNING: 0,
    SUCCEEDED: 0,
    FAILED: 0
  };
  (jobs.value || []).forEach((item) => {
    const key = String(item?.status || "").toUpperCase();
    if (Object.prototype.hasOwnProperty.call(counts, key)) {
      counts[key] += 1;
    }
  });
  return counts;
});

const detailPayload = computed(() => detail.value?.payload || {});
const detailReport = computed(() => detail.value?.result?.artifacts?.report || {});
const detailContextMeta = computed(() => {
  const meta = detailReport.value?.context_meta;
  return meta && typeof meta === "object" ? meta : {};
});
const detailWarnings = computed(() => (Array.isArray(detail.value?.result?.warnings) ? detail.value.result.warnings : []));
const detailSimulations = computed(() => (Array.isArray(detailReport.value?.simulations) ? detailReport.value.simulations : []));
const detailAssetEntries = computed(() => {
  if (Array.isArray(detailReport.value?.candidates)) {
    return detailReport.value.candidates;
  }
  if (Array.isArray(detailReport.value?.strategies)) {
    return detailReport.value.strategies;
  }
  return [];
});
const detailPublishPayloadCount = computed(() => {
  const topLevelCount = Number(detail.value?.payload_count);
  if (Number.isFinite(topLevelCount) && topLevelCount >= 0) {
    return topLevelCount;
  }
  const items = Array.isArray(detailReport.value?.publish_payloads) ? detailReport.value.publish_payloads : [];
  return items.length;
});
const detailWarningCount = computed(() => {
  const topLevelCount = Number(detail.value?.warning_count);
  if (Number.isFinite(topLevelCount) && topLevelCount >= 0) {
    return topLevelCount;
  }
  return detailWarnings.value.length;
});
const detailSelectedCount = computed(() => {
  const topLevelCount = Number(detail.value?.selected_count);
  if (Number.isFinite(topLevelCount) && topLevelCount >= 0) {
    return topLevelCount;
  }
  const reportCount = Number(detailReport.value?.selected_count);
  if (Number.isFinite(reportCount)) {
    return reportCount;
  }
  return detailPublishPayloadCount.value;
});
const detailConsensusSummary = computed(() =>
  String(detailReport.value?.consensus_summary || detail.value?.result_summary || detail.value?.result?.summary || "").trim()
);
const detailContextRows = computed(() => [
  { label: "行情交易日", value: detailContextMeta.value?.selected_trade_date || "-" },
  { label: "行情来源", value: detailContextMeta.value?.price_source || "-" },
  {
    label: "资讯窗口",
    value: Number.isFinite(Number(detailContextMeta.value?.news_window_days))
      ? `${detailContextMeta.value.news_window_days} 天`
      : "-"
  }
]);
const detailConfigRefs = computed(() => {
  const refs = detailPayload.value?.config_refs;
  return refs && typeof refs === "object" ? refs : {};
});
const detailConfigCards = computed(() =>
  configRefOrder.map((item) => ({
    key: item.key,
    label: item.label,
    data: detailConfigRefs.value?.[item.key] || null
  }))
);
const detailInlineConfigRows = computed(() => {
  const payload = detailPayload.value || {};
  const seedItems = Array.isArray(payload.seed_symbols)
    ? payload.seed_symbols
    : Array.isArray(payload.contracts)
      ? payload.contracts
      : [];
  return [
    { label: "交易日", value: payload.trade_date || "-" },
    { label: "种子输入", value: seedItems.length ? seedItems.join(" / ") : "-" },
    {
      label: "角色列表",
      value: Array.isArray(payload.enabled_agents) && payload.enabled_agents.length ? payload.enabled_agents.join(" / ") : "-"
    },
    {
      label: "正向阈值",
      value: Number.isFinite(Number(payload.positive_threshold)) ? String(payload.positive_threshold) : "-"
    },
    {
      label: "负向阈值",
      value: Number.isFinite(Number(payload.negative_threshold)) ? String(payload.negative_threshold) : "-"
    },
    {
      label: "允许 veto",
      value: typeof payload.allow_veto === "boolean" ? (payload.allow_veto ? "是" : "否") : "-"
    }
  ];
});
const detailVetoedAssets = computed(() =>
  detailSimulations.value
    .filter((item) => Boolean(item?.vetoed))
    .map((item) => resolveSimulationAssetKey(item))
    .filter(Boolean)
);
const detailInvalidatedAssets = computed(() => {
  const invalidated = [];
  detailAssetEntries.value.forEach((item) => {
    if (!Array.isArray(item?.invalidations) || item.invalidations.length === 0) {
      return;
    }
    invalidated.push(resolveAssetKey(item));
  });
  return invalidated.filter(Boolean);
});
const detailPublishPolicyPreview = computed(() => {
  const preview = detailPayload.value?.publish_policy_preview;
  return preview && typeof preview === "object" ? preview : null;
});
const detailReplays = computed(() => sortReplays(detail.value?.replays));
const detailHighestRisk = computed(() => {
  const levels = detailAssetEntries.value
    .map((item) => String(item?.risk_level || "").toUpperCase())
    .filter((item) => riskRankMap[item]);
  if (levels.length === 0) {
    return "";
  }
  return levels.reduce((highest, current) => (riskRankMap[current] > riskRankMap[highest] ? current : highest), levels[0]);
});
const detailPublishState = computed(() => {
  const preview = detailPublishPolicyPreview.value;
  if (!preview) {
    return {
      key: "UNKNOWN",
      type: "info",
      label: "UNKNOWN",
      summary: "该任务未携带发布策略快照，当前只能展示运行结果，不能严格判断是否可发布。",
      reasons: ["缺少 publish_policy_preview"]
    };
  }

  const reasons = [];
  const maxRiskLevel = String(preview.max_risk_level || "").toUpperCase();
  const highestRisk = detailHighestRisk.value;
  if (maxRiskLevel && highestRisk && riskRankMap[highestRisk] > riskRankMap[maxRiskLevel]) {
    reasons.push(`最高风险等级 ${highestRisk} 超过发布上限 ${maxRiskLevel}`);
  }
  if (!preview.allow_vetoed_publish && detailVetoedAssets.value.length > 0) {
    reasons.push(`存在 ${detailVetoedAssets.value.length} 个 veto 标的，当前策略不允许发布`);
  }
  const maxWarningCount = Number(preview.max_warning_count);
  if (Number.isFinite(maxWarningCount) && detailWarnings.value.length > maxWarningCount) {
    reasons.push(`警告数 ${detailWarnings.value.length} 超过上限 ${maxWarningCount}`);
  }

  if (reasons.length > 0) {
    return {
      key: "BLOCKED",
      type: "danger",
      label: "BLOCKED",
      summary: "当前任务命中了发布门槛，按默认策略应先拦截或人工覆盖后再发布。",
      reasons
    };
  }

  return {
    key: "READY",
    type: "success",
    label: "READY",
    summary: "当前任务未命中默认发布门槛，可按策略发布。",
    reasons: []
  };
});

const detailStorageState = computed(() => {
  const source = String(detail.value?.storage_source || "").toUpperCase();
  const syncedAt = String(detail.value?.synced_at || "").trim();
  switch (source) {
    case "LOCAL_ARCHIVED":
      return {
        label: "已归档快照",
        type: "success",
        note: syncedAt ? `当前详情来自 Go 后端本地快照，归档时间 ${syncedAt}。` : "当前详情来自 Go 后端本地快照。"
      };
    case "REMOTE_BACKFILLED":
      return {
        label: "远端回源已归档",
        type: "warning",
        note: syncedAt ? `本次详情由 strategy-engine 回源后补归档，归档时间 ${syncedAt}。` : "本次详情由 strategy-engine 回源后补归档。"
      };
    case "REMOTE_ONLY":
      return {
        label: "远端临时结果",
        type: "info",
        note: "当前仅拿到远端临时结果，本地归档尚未成功，请尽快检查后端快照表或迁移状态。"
      };
    default:
      return {
        label: "来源未标记",
        type: "info",
        note: "当前记录未返回来源标记，默认按接口返回结果展示。"
      };
  }
});

function prettyJson(value) {
  return JSON.stringify(value || {}, null, 2);
}

function statusTagType(status) {
  return statusTagTypeMap[String(status || "").toUpperCase()] || "info";
}

function getTradeDate(row) {
  return row?.trade_date || row?.payload?.trade_date || row?.result?.payload_echo?.trade_date || row?.result?.artifacts?.report?.trade_date || "-";
}

function getSelectedCount(row) {
  const topLevelCount = Number(row?.selected_count);
  if (Number.isFinite(topLevelCount) && topLevelCount >= 0) {
    return topLevelCount;
  }
  return row?.result?.artifacts?.report?.selected_count ?? "-";
}

function getWarningCount(row) {
  const topLevelCount = Number(row?.warning_count);
  if (Number.isFinite(topLevelCount) && topLevelCount >= 0) {
    return topLevelCount;
  }
  return Array.isArray(row?.result?.warnings) ? row.result.warnings.length : 0;
}

function getResultSummary(row) {
  return row?.result_summary || row?.result?.summary || row?.error_message || "-";
}

function resolveAssetKey(item) {
  return String(item?.symbol || item?.contract || item?.asset_key || "").trim();
}

function resolveSimulationAssetKey(item) {
  return String(item?.asset_key || item?.symbol || item?.contract || "").trim();
}

function formatConfidence(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  return num.toFixed(1);
}

function formatConfigRefNote(data) {
  if (!data) {
    return "该任务未携带这类配置快照。";
  }
  const parts = [];
  if (data.target_type) {
    parts.push(String(data.target_type));
  }
  parts.push(data.is_default ? "默认配置" : "非默认配置");
  if (data.source) {
    parts.push(String(data.source));
  }
  return parts.join(" · ");
}

function sortReplays(items) {
  if (!Array.isArray(items)) {
    return [];
  }
  return [...items].sort((left, right) => String(right?.created_at || "").localeCompare(String(left?.created_at || "")));
}

function replayModeLabel(item) {
  return item?.force_publish ? "人工覆盖发布" : "按策略发布";
}

function publishModeLabel(value) {
  return String(value || "").toUpperCase() === "OVERRIDE" ? "人工覆盖发布" : "按策略发布";
}

function replaySourceLabel(item) {
  switch (String(item?.storage_source || "").toUpperCase()) {
    case "LOCAL_ARCHIVED":
      return "本地归档";
    case "REMOTE_BACKFILLED":
      return "远端回填";
    case "REMOTE_ONLY":
      return "远端结果";
    default:
      return "未标记";
  }
}

function publishSourceLabel(value) {
  switch (String(value || "").toUpperCase()) {
    case "LOCAL_ARCHIVED":
      return "本地归档";
    case "REMOTE_BACKFILLED":
      return "远端回填";
    case "REMOTE_ONLY":
      return "远端结果";
    default:
      return "未标记";
  }
}

function ensureCanPublish() {
  if (canEditMarket) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法执行策略发布";
  return false;
}

function getReplayCount(row) {
  const topLevelCount = Number(row?.publish_count);
  if (Number.isFinite(topLevelCount) && topLevelCount >= 0) {
    return topLevelCount;
  }
  return sortReplays(row?.replays).length;
}

function getLatestReplay(row) {
  return sortReplays(row?.replays)[0] || null;
}

function getLatestReplaySummary(row) {
  const latestPublishAt = String(row?.latest_publish_at || "").trim();
  const latestPublishID = String(row?.latest_publish_id || "").trim();
  const topLevelCount = Number(row?.publish_count);
  if ((Number.isFinite(topLevelCount) && topLevelCount > 0) || latestPublishAt || latestPublishID) {
    const parts = [publishModeLabel(row?.latest_publish_mode), publishSourceLabel(row?.latest_publish_source)];
    if (row?.latest_publish_version) {
      parts.push(`V${row.latest_publish_version}`);
    }
    if (latestPublishAt) {
      parts.push(latestPublishAt);
    }
    return parts.filter(Boolean).join(" · ");
  }
  const replay = getLatestReplay(row);
  if (!replay) {
    return "未发布";
  }
  const parts = [replayModeLabel(replay), replaySourceLabel(replay)];
  if (replay.publish_version) {
    parts.push(`V${replay.publish_version}`);
  }
  if (replay.created_at) {
    parts.push(replay.created_at);
  }
  return parts.filter(Boolean).join(" · ");
}

async function refreshJobs(targetPage = page.value) {
  loading.value = true;
  errorMessage.value = "";
  try {
    const data = await listStrategyEngineJobs({
      page: targetPage,
      page_size: pageSize.value,
      job_type: filters.job_type,
      status: filters.status
    });
    jobs.value = data.items || [];
    total.value = Number(data.total || 0);
    page.value = Number(data.page || targetPage || 1);
    pageSize.value = Number(data.page_size || pageSize.value);
  } catch (error) {
    errorMessage.value = error?.message || "加载策略作业失败";
  } finally {
    loading.value = false;
  }
}

async function openDetail(row) {
  if (!row?.job_id) {
    return;
  }
  detailLoading.value = true;
  errorMessage.value = "";
  detailTab.value = "overview";
  try {
    detail.value = await getStrategyEngineJob(row.job_id);
    detailDialogVisible.value = true;
  } catch (error) {
    errorMessage.value = error?.message || "加载作业详情失败";
  } finally {
    detailLoading.value = false;
  }
}

async function handlePublish(force = false) {
  if (!ensureCanPublish()) {
    return;
  }
  if (!detail.value?.job_id) {
    return;
  }
  let overrideReason = "";
  if (force) {
    const result = await ElMessageBox.prompt("请输入人工覆盖发布原因，便于后续复盘。", "人工覆盖发布", {
      confirmButtonText: "确认发布",
      cancelButtonText: "取消",
      inputType: "textarea",
      inputPlaceholder: "例如：今日热点事件驱动明确，允许在警告存在时人工放行。",
      inputValidator: (value) => String(value || "").trim().length > 0,
      inputErrorMessage: "请填写覆盖原因"
    }).catch(() => null);
    if (!result) {
      return;
    }
    overrideReason = result.value;
  }

  publishSubmitting.value = true;
  try {
    const record = await publishStrategyEngineJob(detail.value.job_id, {
      force,
      override_reason: overrideReason
    });
    ElMessage.success(`已生成发布记录 ${record.publish_id}`);
    await openDetail(detail.value);
    await refreshJobs(page.value);
  } catch (error) {
    ElMessage.error(error?.message || "发布失败");
  } finally {
    publishSubmitting.value = false;
  }
}

function handleFilter() {
  refreshJobs(1);
}

function handlePageChange(nextPage) {
  refreshJobs(nextPage);
}

onMounted(() => {
  refreshJobs(1);
});

defineExpose({ refreshJobs });
</script>

<template>
  <section class="card" v-loading="loading || detailLoading">
    <div class="toolbar panel-head">
      <div>
        <h3>作业中心</h3>
        <p class="muted">查看策略任务的输入快照、运行状态、结果摘要和风险提醒，后台真正能追溯每次生成过程。</p>
      </div>
      <el-button @click="refreshJobs(1)">刷新</el-button>
    </div>

    <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon class="panel-alert" />

    <div class="job-toolbar">
      <div class="job-filters">
        <el-select v-model="filters.job_type" class="job-filter" @change="handleFilter">
          <el-option v-for="item in jobTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
        <el-select v-model="filters.status" class="job-filter" @change="handleFilter">
          <el-option v-for="item in statusOptions" :key="item.value" :label="item.label" :value="item.value" />
        </el-select>
      </div>
      <div class="job-status-summary">
        <el-tag type="info">排队 {{ statusSummary.QUEUED }}</el-tag>
        <el-tag type="warning">运行 {{ statusSummary.RUNNING }}</el-tag>
        <el-tag type="success">成功 {{ statusSummary.SUCCEEDED }}</el-tag>
        <el-tag type="danger">失败 {{ statusSummary.FAILED }}</el-tag>
      </div>
    </div>

    <el-table :data="jobs" border>
      <el-table-column prop="created_at" label="创建时间" min-width="180" />
      <el-table-column prop="job_type" label="任务类型" min-width="150" />
      <el-table-column label="状态" width="110">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ row.status || "-" }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="requested_by" label="发起方" min-width="120" />
      <el-table-column label="交易日" width="120">
        <template #default="{ row }">{{ getTradeDate(row) }}</template>
      </el-table-column>
      <el-table-column label="入选数" width="90">
        <template #default="{ row }">{{ getSelectedCount(row) }}</template>
      </el-table-column>
      <el-table-column label="警告" width="90">
        <template #default="{ row }">{{ getWarningCount(row) }}</template>
      </el-table-column>
      <el-table-column label="发布次数" width="90">
        <template #default="{ row }">{{ getReplayCount(row) }}</template>
      </el-table-column>
      <el-table-column label="最近发布" min-width="220">
        <template #default="{ row }">
          <span>{{ getLatestReplaySummary(row) }}</span>
        </template>
      </el-table-column>
      <el-table-column label="结果摘要" min-width="260">
        <template #default="{ row }">{{ getResultSummary(row) }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pager-row">
      <el-pagination
        background
        layout="total, prev, pager, next"
        :current-page="page"
        :page-size="pageSize"
        :total="total"
        @current-change="handlePageChange"
      />
    </div>

    <el-dialog v-model="detailDialogVisible" title="策略作业详情" width="1080px">
      <template v-if="detail">
        <el-descriptions :column="3" border class="job-detail-overview">
          <el-descriptions-item label="任务编号">{{ detail.job_id }}</el-descriptions-item>
          <el-descriptions-item label="任务类型">{{ detail.job_type }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(detail.status)">{{ detail.status || "-" }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="发起方">{{ detail.requested_by || "-" }}</el-descriptions-item>
          <el-descriptions-item label="Trace ID">{{ detail.trace_id || "-" }}</el-descriptions-item>
          <el-descriptions-item label="交易日">{{ getTradeDate(detail) }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ detail.created_at || "-" }}</el-descriptions-item>
          <el-descriptions-item label="开始时间">{{ detail.started_at || "-" }}</el-descriptions-item>
          <el-descriptions-item label="结束时间">{{ detail.finished_at || "-" }}</el-descriptions-item>
          <el-descriptions-item label="数据来源">
            <el-tag :type="detailStorageState.type">{{ detailStorageState.label }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="归档时间">{{ detail.synced_at || "-" }}</el-descriptions-item>
        </el-descriptions>

        <el-tabs v-model="detailTab" class="detail-tabs">
          <el-tab-pane label="概览" name="overview">
            <div class="detail-block">
              <el-alert :title="detailStorageState.note" :type="detailStorageState.type" show-icon :closable="false" class="detail-alert" />
              <div class="detail-grid detail-grid--4">
                <div class="detail-metric">
                  <span class="metric-label">入选结果</span>
                  <strong>{{ detailSelectedCount }}</strong>
                </div>
                <div class="detail-metric">
                  <span class="metric-label">警告数量</span>
                  <strong>{{ detailWarningCount }}</strong>
                </div>
                <div class="detail-metric">
                  <span class="metric-label">发布 payload</span>
                  <strong>{{ detailPublishPayloadCount }}</strong>
                </div>
                <div class="detail-metric">
                  <span class="metric-label">发布次数</span>
                  <strong>{{ getReplayCount(detail) }}</strong>
                </div>
              </div>
              <div class="detail-grid detail-grid--2">
                <div class="detail-metric">
                  <span class="metric-label">发布就绪</span>
                  <strong>{{ detailPublishState.label }}</strong>
                </div>
                <div class="detail-metric">
                  <span class="metric-label">最近发布</span>
                  <strong>{{ getLatestReplaySummary(detail) }}</strong>
                </div>
              </div>
              <div class="detail-card">
                <p class="detail-card__label">真实数据上下文</p>
                <div class="inline-grid">
                  <article v-for="item in detailContextRows" :key="item.label">
                    <p>{{ item.label }}</p>
                    <strong>{{ item.value }}</strong>
                  </article>
                </div>
              </div>
              <div class="detail-card">
                <p class="detail-card__label">多角色收敛</p>
                <strong>{{ detailConsensusSummary || "暂无收敛摘要" }}</strong>
                <span>{{ getResultSummary(detail) }}</span>
              </div>
              <div v-if="detailReplays.length" class="detail-card">
                <p class="detail-card__label">最近一次发布动作</p>
                <strong>{{ replayModeLabel(detailReplays[0]) }}</strong>
                <span>{{ detailReplays[0].created_at || "暂无时间" }} · {{ replaySourceLabel(detailReplays[0]) }}</span>
                <span>操作人 {{ detailReplays[0].operator || "-" }} · 覆盖原因 {{ detailReplays[0].override_reason || "无" }}</span>
              </div>
              <el-alert
                v-if="detail.error_message"
                :title="detail.error_message"
                type="error"
                show-icon
                class="detail-alert"
              />
              <div v-if="detailWarnings.length" class="warning-list">
                <el-tag v-for="item in detailWarnings" :key="item" type="warning" effect="plain">{{ item }}</el-tag>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="配置来源" name="config">
            <div class="detail-block">
              <div class="config-grid">
                <article v-for="item in detailConfigCards" :key="item.key" class="detail-card">
                  <p class="detail-card__label">{{ item.label }}</p>
                  <strong>{{ item.data?.name || "未携带快照" }}</strong>
                  <span>{{ formatConfigRefNote(item.data) }}</span>
                  <em v-if="item.data?.updated_at">{{ item.data.updated_at }}</em>
                </article>
              </div>
              <div class="detail-card">
                <p class="detail-card__label">直传参数回退</p>
                <div class="inline-grid">
                  <article v-for="item in detailInlineConfigRows" :key="item.label">
                    <p>{{ item.label }}</p>
                    <strong>{{ item.value }}</strong>
                  </article>
                </div>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="角色评审" name="agents">
            <div v-if="detailSimulations.length" class="simulation-grid">
              <article v-for="item in detailSimulations" :key="resolveSimulationAssetKey(item) || item.asset_type" class="simulation-card">
                <div class="simulation-card__head">
                  <div>
                    <p>{{ resolveSimulationAssetKey(item) || "未命名标的" }}</p>
                    <strong>{{ item.consensus_action || "-" }}</strong>
                  </div>
                  <el-tag :type="item.vetoed ? 'danger' : 'success'">{{ item.vetoed ? "已 veto" : "可继续" }}</el-tag>
                </div>
                <p class="simulation-card__note">{{ item.veto_reason || "当前无 veto 原因。" }}</p>
                <div v-if="Array.isArray(item.agents) && item.agents.length" class="agent-list">
                  <article v-for="agent in item.agents" :key="`${resolveSimulationAssetKey(item)}-${agent.agent}`">
                    <p>{{ agent.agent || "agent" }}</p>
                    <strong>{{ agent.stance || "-" }} · {{ formatConfidence(agent.confidence) }}</strong>
                    <span>{{ agent.summary || "暂无角色说明" }}</span>
                  </article>
                </div>
                <el-empty v-else description="暂无角色输出" :image-size="64" />
              </article>
            </div>
            <el-empty v-else description="当前作业没有 simulations，暂无法展示角色评审。" :image-size="88" />
          </el-tab-pane>

          <el-tab-pane label="场景输出" name="scenarios">
            <div v-if="detailSimulations.length" class="simulation-grid">
              <article v-for="item in detailSimulations" :key="`${resolveSimulationAssetKey(item)}-scenarios`" class="simulation-card">
                <div class="simulation-card__head">
                  <div>
                    <p>{{ resolveSimulationAssetKey(item) || "未命名标的" }}</p>
                    <strong>{{ item.asset_type || detail.job_type }}</strong>
                  </div>
                  <span class="muted">{{ (item.scenarios || []).length }} 个场景</span>
                </div>
                <div v-if="Array.isArray(item.scenarios) && item.scenarios.length" class="scenario-list">
                  <article v-for="scenario in item.scenarios" :key="`${resolveSimulationAssetKey(item)}-${scenario.scenario}`">
                    <p>{{ scenario.scenario || "scenario" }}</p>
                    <strong>{{ scenario.action || "-" }}</strong>
                    <span>{{ scenario.thesis || "暂无 thesis" }}</span>
                    <em>风险 {{ scenario.risk_signal || "-" }} · 分数 {{ scenario.score_adjustment ?? "-" }}</em>
                  </article>
                </div>
                <el-empty v-else description="暂无场景输出" :image-size="64" />
              </article>
            </div>
            <el-empty v-else description="当前作业没有 simulations，暂无法展示场景输出。" :image-size="88" />
          </el-tab-pane>

          <el-tab-pane label="风险与发布" name="risk-publish">
            <div class="detail-block">
              <div class="detail-card detail-card--highlight">
                <div class="status-row">
                  <div>
                    <p class="detail-card__label">发布就绪状态</p>
                    <strong>{{ detailPublishState.label }}</strong>
                    <span>{{ detailPublishState.summary }}</span>
                  </div>
                  <el-tag :type="detailPublishState.type">{{ detailPublishState.label }}</el-tag>
                </div>
                <div v-if="detailPublishState.reasons.length" class="reason-list">
                  <el-tag v-for="item in detailPublishState.reasons" :key="item" type="danger" effect="plain">{{ item }}</el-tag>
                </div>
              </div>

              <div class="detail-grid detail-grid--2">
                <div class="detail-card">
                  <p class="detail-card__label">风险输出</p>
                  <div class="tag-block">
                    <span class="tag-title">warnings</span>
                    <div class="warning-list">
                      <el-tag v-for="item in detailWarnings" :key="item" type="warning" effect="plain">{{ item }}</el-tag>
                      <span v-if="!detailWarnings.length" class="muted">暂无 warnings</span>
                    </div>
                  </div>
                  <div class="tag-block">
                    <span class="tag-title">vetoed_assets</span>
                    <div class="warning-list">
                      <el-tag v-for="item in detailVetoedAssets" :key="item" type="danger" effect="plain">{{ item }}</el-tag>
                      <span v-if="!detailVetoedAssets.length" class="muted">暂无 veto 资产</span>
                    </div>
                  </div>
                  <div class="tag-block">
                    <span class="tag-title">invalidated_assets</span>
                    <div class="warning-list">
                      <el-tag v-for="item in detailInvalidatedAssets" :key="item" type="info" effect="plain">{{ item }}</el-tag>
                      <span v-if="!detailInvalidatedAssets.length" class="muted">暂无 invalidations</span>
                    </div>
                  </div>
                </div>

                <div class="detail-card">
                  <p class="detail-card__label">发布策略快照</p>
                  <div v-if="detailPublishPolicyPreview" class="inline-grid">
                    <article>
                      <p>最高风险</p>
                      <strong>{{ detailPublishPolicyPreview.max_risk_level || "-" }}</strong>
                    </article>
                    <article>
                      <p>警告上限</p>
                      <strong>{{ detailPublishPolicyPreview.max_warning_count ?? "-" }}</strong>
                    </article>
                    <article>
                      <p>允许 veto 发布</p>
                      <strong>{{ detailPublishPolicyPreview.allow_vetoed_publish ? "是" : "否" }}</strong>
                    </article>
                    <article>
                      <p>默认发布者</p>
                      <strong>{{ detailPublishPolicyPreview.default_publisher || "-" }}</strong>
                    </article>
                  </div>
                  <span v-else class="muted">该任务未携带发布策略快照。</span>
                </div>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="发布审计" name="publish-audit">
            <div v-if="detailReplays.length" class="detail-block">
              <article v-for="item in detailReplays" :key="`${item.publish_id}-${item.created_at}`" class="detail-card">
                <div class="status-row">
                  <div>
                    <p class="detail-card__label">归档 {{ item.publish_id || "-" }}</p>
                    <strong>{{ replayModeLabel(item) }}</strong>
                    <span>{{ item.created_at || "暂无时间" }} · {{ replaySourceLabel(item) }}</span>
                  </div>
                  <el-tag :type="item.force_publish ? 'danger' : 'success'">{{ replayModeLabel(item) }}</el-tag>
                </div>
                <div class="inline-grid">
                  <article>
                    <p>操作人</p>
                    <strong>{{ item.operator || "-" }}</strong>
                  </article>
                  <article>
                    <p>任务ID</p>
                    <strong>{{ item.job_id || detail.job_id || "-" }}</strong>
                  </article>
                  <article>
                    <p>警告数</p>
                    <strong>{{ item.warning_count || 0 }}</strong>
                  </article>
                  <article>
                    <p>覆盖原因</p>
                    <strong>{{ item.override_reason || "无" }}</strong>
                  </article>
                </div>
                <div class="tag-block">
                  <span class="tag-title">notes</span>
                  <div class="warning-list">
                    <el-tag v-for="note in item.notes || []" :key="note" type="info" effect="plain">{{ note }}</el-tag>
                    <span v-if="!(item.notes || []).length" class="muted">暂无 replay notes</span>
                  </div>
                </div>
                <div class="tag-block">
                  <span class="tag-title">policy_snapshot</span>
                  <pre class="json-block json-block--compact">{{ prettyJson(item.policy_snapshot || {}) }}</pre>
                </div>
              </article>
            </div>
            <el-empty v-else description="当前任务还没有发布审计记录。" :image-size="88" />
          </el-tab-pane>

          <el-tab-pane label="原始 JSON" name="raw-json">
            <div class="detail-block">
              <div>
                <p class="json-title">payload</p>
                <pre class="json-block">{{ prettyJson(detailPayload) }}</pre>
              </div>
              <div>
                <p class="json-title">artifacts.report</p>
                <pre class="json-block">{{ prettyJson(detailReport) }}</pre>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </template>
      <template #footer>
        <el-button v-if="canEditMarket" :loading="publishSubmitting" type="primary" plain @click="handlePublish(false)">
          按策略发布
        </el-button>
        <el-button v-if="canEditMarket" :loading="publishSubmitting" type="warning" plain @click="handlePublish(true)">
          人工覆盖发布
        </el-button>
        <el-button @click="detailDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </section>
</template>

<style scoped>
.job-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
  flex-wrap: wrap;
}

.job-filters {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.job-filter {
  width: 180px;
}

.job-status-summary {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.pager-row {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.job-detail-overview {
  margin-bottom: 16px;
}

.detail-tabs {
  margin-top: 16px;
}

.detail-block {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.detail-grid--4 {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.detail-grid--2 {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.detail-metric,
.detail-card,
.simulation-card {
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  padding: 16px;
  background: var(--el-fill-color-lighter);
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.detail-card--highlight {
  background: linear-gradient(135deg, rgba(64, 158, 255, 0.08), rgba(103, 194, 58, 0.08));
}

.metric-label,
.detail-card__label,
.simulation-card__note,
.tag-title,
.json-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.detail-summary {
  margin: 0;
  line-height: 1.7;
  color: var(--el-text-color-primary);
}

.detail-alert {
  margin: 0;
}

.warning-list,
.reason-list {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.config-grid,
.simulation-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.inline-grid,
.agent-list,
.scenario-list {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.inline-grid article,
.agent-list article,
.scenario-list article {
  border: 1px solid var(--el-border-color-light);
  border-radius: 10px;
  padding: 12px;
  background: var(--el-bg-color);
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.inline-grid p,
.inline-grid strong,
.agent-list p,
.agent-list strong,
.agent-list span,
.scenario-list p,
.scenario-list strong,
.scenario-list span,
.scenario-list em {
  margin: 0;
}

.simulation-card__head,
.status-row {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: flex-start;
}

.simulation-card__head p,
.simulation-card__head strong,
.status-row p,
.status-row strong,
.status-row span {
  margin: 0;
}

.tag-block {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.muted {
  color: var(--el-text-color-secondary);
}

.json-block {
  margin: 0;
  padding: 16px;
  border-radius: 12px;
  background: #0f172a;
  color: #e2e8f0;
  font-size: 12px;
  line-height: 1.6;
  overflow: auto;
  max-height: 420px;
}

.json-block--compact {
  max-height: 220px;
}

@media (max-width: 1080px) {
  .detail-grid--4,
  .config-grid,
  .simulation-grid,
  .inline-grid,
  .agent-list,
  .scenario-list {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 960px) {
  .detail-grid,
  .detail-grid--2 {
    grid-template-columns: 1fr;
  }

  .job-filter {
    width: 100%;
  }
}
</style>
