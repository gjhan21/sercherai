<template>
  <section class="forecast-run-page fade-up">
    <header class="forecast-run-hero card">
      <div class="finance-copy-stack">
        <div class="finance-pill-row">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">深推演报告</span>
          <span class="finance-pill finance-pill-compact" :class="forecastToneClass">{{ forecastStatusLabel }}</span>
        </div>
        <div>
          <p class="hero-kicker">Forecast L3</p>
          <h1 class="section-title">{{ targetTitle }}</h1>
          <p class="section-subtitle">
            查看异步深推演的运行状态、结构化报告和关键步骤日志。
          </p>
          <p v-if="loading" class="api-state">正在同步深推演结果...</p>
          <p v-else-if="errorMessage" class="api-state warning">{{ errorMessage }}</p>
          <p v-else class="api-state">运行编号：{{ run?.id || route.params.id }}</p>
        </div>
      </div>
      <div class="hero-actions">
        <button class="primary-btn finance-primary-btn" type="button" :disabled="loading" @click="loadDetail">
          {{ loading ? "同步中..." : "刷新深推演" }}
        </button>
        <button class="ghost-btn finance-ghost-btn" type="button" @click="goBack">返回上一页</button>
        <button
          v-if="reportRequiresVIP && !isVIPUser"
          class="ghost-btn finance-ghost-btn"
          type="button"
          @click="goMembership"
        >
          升级会员看全文
        </button>
      </div>
      <div class="forecast-run-stats finance-hero-stat-grid">
        <article class="finance-hero-stat-card">
          <span>运行状态</span>
          <strong>{{ forecastStatusLabel }}</strong>
          <p>{{ forecastSummary?.summary || "当前还没有可展示的深推演摘要。" }}</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>主情景</span>
          <strong>{{ forecastSummary?.scenario || "-" }}</strong>
          <p>{{ forecastSummary?.actionGuidance || "等待更多运行结果返回。" }}</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>报告可读性</span>
          <strong>{{ reportReadableLabel }}</strong>
          <p>{{ reportGateNote }}</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>步骤日志</span>
          <strong>{{ logs.length }} 条</strong>
          <p>保留研究包构建、深推演执行和报告落库的关键过程。</p>
        </article>
      </div>
    </header>

    <section class="forecast-run-layout">
      <article class="card forecast-main-card">
        <header class="section-head">
          <div>
            <h2 class="section-title">结构化报告</h2>
            <p class="section-subtitle">摘要先读，报告正文次之，保持当前策略阅读链节奏。</p>
          </div>
        </header>

        <div v-if="reportSummaryCard" class="forecast-summary-box finance-card-pale">
          <div class="stock-news-head">
            <p>执行摘要</p>
            <span>{{ forecastStatusLabel }}</span>
          </div>
          <p class="forecast-summary-text">{{ reportSummaryCard.summary }}</p>
          <div class="reason-support-grid">
            <article class="finance-list-card finance-list-card-panel">
              <p>主情景</p>
              <strong>{{ reportSummaryCard.scenario || "-" }}</strong>
              <span>{{ reportSummaryCard.actionGuidance || "等待更多动作建议。" }}</span>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>生成时间</p>
              <strong>{{ formatDateTime(reportSummaryCard.generatedAt) || "-" }}</strong>
              <span>{{ run?.engine_key || run?.summary?.engine_key || "LOCAL_SYNTHESIS" }}</span>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>目标对象</p>
              <strong>{{ run?.target_label || run?.target_key || "-" }}</strong>
              <span>{{ run?.target_type || "-" }} · {{ run?.trigger_type || "-" }}</span>
            </article>
          </div>
        </div>

        <StatePanel
          v-if="!report && !loading && !errorMessage"
          tone="info"
          eyebrow="深推演状态"
          :title="statePanelTitle"
          :description="statePanelDescription"
          compact
        >
          <template #actions>
            <button type="button" @click="loadDetail">刷新结果</button>
            <button
              v-if="reportRequiresVIP && !isVIPUser"
              type="button"
              class="ghost-btn finance-ghost-btn"
              @click="goMembership"
            >
              升级会员
            </button>
          </template>
        </StatePanel>

        <template v-else-if="report">
          <section class="forecast-report-section finance-card-surface">
            <div class="stock-news-head">
              <p>关键结论</p>
              <span>{{ report.primary_scenario || "-" }}</span>
            </div>
            <div class="reason-support-grid">
              <article class="finance-list-card finance-list-card-panel">
                <p>执行摘要</p>
                <strong>{{ report.executive_summary || "当前未补更多摘要。" }}</strong>
                <span>{{ report.summary?.confidence_label || forecastSummary?.statusLabel || "" }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>动作建议</p>
                <strong>{{ report.action_guidance?.[0] || forecastSummary?.actionGuidance || "等待更多动作建议。" }}</strong>
                <span>{{ report.invalidation_signals?.[0] || "当前未补更多失效信号。" }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>角色分歧</p>
                <strong>{{ report.role_disagreements?.length || 0 }} 条</strong>
                <span>{{ report.role_disagreements?.[0]?.summary || "当前未补更多角色分歧。" }}</span>
              </article>
            </div>
          </section>

          <section v-if="report.trigger_checklist?.length" class="forecast-report-section finance-card-surface">
            <div class="stock-news-head">
              <p>触发检查清单</p>
              <span>{{ report.trigger_checklist.length }} 条</span>
            </div>
            <div class="scenario-grid">
              <article
                v-for="item in report.trigger_checklist"
                :key="`${item.label}-${item.trigger}`"
                class="scenario-item finance-list-card finance-list-card-panel"
              >
                <p>{{ item.label }}</p>
                <strong>{{ item.status }}</strong>
                <span>{{ item.note || item.trigger || "等待更多触发信号。" }}</span>
              </article>
            </div>
          </section>

          <section v-if="report.role_disagreements?.length" class="forecast-report-section finance-card-surface">
            <div class="stock-news-head">
              <p>角色分歧与 veto</p>
              <span>{{ report.role_disagreements.length }} 条</span>
            </div>
            <div class="agent-opinion-list">
              <article
                v-for="item in report.role_disagreements"
                :key="`${item.role}-${item.stance}`"
                class="agent-opinion-item finance-list-card finance-list-card-panel"
              >
                <p>{{ item.role }}</p>
                <strong>{{ item.stance || "WATCH" }}</strong>
                <span>{{ item.summary }}</span>
                <em v-if="item.veto">已触发 veto</em>
              </article>
            </div>
          </section>

          <section class="forecast-report-section finance-card-surface">
            <div class="stock-news-head">
              <p>报告正文</p>
              <span>{{ reportReadableLabel }}</span>
            </div>
            <p v-if="reportRequiresVIP && !isVIPUser" class="forecast-lock-note">
              当前报告正文需要会员权限。已保留摘要、主情景和关键步骤日志，升级后可继续查看完整正文。
            </p>
            <pre v-else class="forecast-report-body">{{ report.markdown_body || "当前未生成正文。" }}</pre>
          </section>
        </template>
      </article>

      <aside class="forecast-side-stack">
        <article class="card finance-card-surface">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">运行日志</h2>
              <p class="section-subtitle">保留关键步骤，便于诊断和后续追踪。</p>
            </div>
          </header>
          <div v-if="logs.length" class="forecast-log-list">
            <article
              v-for="item in logs"
              :key="item.id"
              class="finance-list-card finance-list-card-panel forecast-log-item"
            >
              <p>{{ item.step_key }}</p>
              <strong>{{ item.status }}</strong>
              <span>{{ item.message || "当前未补更多步骤说明。" }}</span>
              <em>{{ formatDateTime(item.created_at) || "-" }}</em>
            </article>
          </div>
          <div v-else class="empty-inline finance-empty-inline">当前还没有步骤日志。</div>
        </article>

        <article class="card finance-card-surface">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">阅读提示</h2>
              <p class="section-subtitle">L3 是增强层，没有结果时继续回到原有 L2。</p>
            </div>
          </header>
          <div class="forecast-note-list">
            <article class="finance-list-card finance-list-card-panel">
              <p>当前对象</p>
              <strong>{{ run?.target_label || run?.target_key || "-" }}</strong>
              <span>{{ run?.target_type || "-" }} · {{ run?.trigger_type || "-" }}</span>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>回退策略</p>
              <strong>没有 L3 时继续读 L2</strong>
              <span>不会打断原有推荐、版本历史和档案阅读链。</span>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>会员边界</p>
              <strong>{{ reportReadableLabel }}</strong>
              <span>{{ reportGateNote }}</span>
            </article>
          </div>
        </article>
      </aside>
    </section>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import StatePanel from "../components/StatePanel.vue";
import { getForecastRunDetail } from "../api/forecast";
import { getMembershipQuota } from "../api/membership";
import { buildStrategyDeepForecastSummary } from "../lib/strategy-version";

const route = useRoute();
const router = useRouter();

const loading = ref(false);
const errorMessage = ref("");
const detail = ref(null);
const isVIPUser = ref(false);

const run = computed(() => detail.value?.run || null);
const report = computed(() => detail.value?.report || null);
const logs = computed(() => (Array.isArray(detail.value?.logs) ? detail.value.logs : []));
const forecastSummary = computed(() =>
  buildStrategyDeepForecastSummary({
    deep_forecast_summary: run.value?.summary,
    deep_forecast_report_ref: run.value?.report_ref
  })
);
const reportSummaryCard = computed(() =>
  buildStrategyDeepForecastSummary({
    deep_forecast_summary: report.value?.summary || run.value?.summary,
    deep_forecast_report_ref: run.value?.report_ref
  })
);
const forecastStatusLabel = computed(() => forecastSummary.value?.statusLabel || "等待结果");
const forecastToneClass = computed(() => {
  const tone = forecastSummary.value?.tone || "muted";
  if (tone === "success") {
    return "finance-pill-success";
  }
  if (tone === "running") {
    return "finance-pill-info";
  }
  if (tone === "queued") {
    return "finance-pill-neutral";
  }
  if (tone === "failed") {
    return "finance-pill-warning";
  }
  return "finance-pill-neutral";
});
const targetTitle = computed(() => run.value?.target_label || run.value?.target_key || "深推演详情");
const reportRequiresVIP = computed(() => run.value?.report_ref?.requires_vip === true);
const reportReadableLabel = computed(() => {
  if (!report.value) {
    return "等待报告生成";
  }
  if (reportRequiresVIP.value && !isVIPUser.value) {
    return "摘要可读";
  }
  return "全文可读";
});
const reportGateNote = computed(() => {
  if (!report.value) {
    return "当前还没有完整报告正文，先看运行状态和步骤日志。";
  }
  if (reportRequiresVIP.value && !isVIPUser.value) {
    return "当前先开放摘要和状态，完整正文需要会员权限。";
  }
  return "当前账号可直接查看完整报告正文。";
});
const statePanelTitle = computed(() => {
  if (forecastSummary.value?.status === "FAILED") {
    return "本次深推演没有成功完成";
  }
  if (forecastSummary.value?.status === "RUNNING" || forecastSummary.value?.status === "QUEUED") {
    return "深推演仍在执行中";
  }
  return "当前还没有可展示的报告正文";
});
const statePanelDescription = computed(() => {
  if (run.value?.failure_reason) {
    return run.value.failure_reason;
  }
  return forecastSummary.value?.summary || "请稍后刷新查看运行结果。";
});

async function loadDetail() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const data = await getForecastRunDetail(route.params.id);
    detail.value = data || null;
  } catch (error) {
    errorMessage.value = error?.message || "深推演详情加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadVIPState() {
  try {
    const quota = await getMembershipQuota();
    isVIPUser.value = resolveVIPStage(quota);
  } catch {
    isVIPUser.value = false;
  }
}

function resolveVIPStage(quota) {
  const membership = quota?.membership || quota || {};
  const level = String(membership?.tier || membership?.level || membership?.plan_key || "").toUpperCase();
  const status = String(membership?.status || "").toUpperCase();
  return Boolean(level.includes("VIP") || status === "ACTIVE");
}

function formatDateTime(value) {
  const text = String(value || "").trim();
  if (!text) {
    return "";
  }
  const date = new Date(text);
  if (Number.isNaN(date.getTime())) {
    return text;
  }
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, "0")}-${String(date.getDate()).padStart(2, "0")} ${String(
    date.getHours()
  ).padStart(2, "0")}:${String(date.getMinutes()).padStart(2, "0")}`;
}

function goBack() {
  const from = String(route.query.from || "");
  if (from.startsWith("/")) {
    router.push(from);
    return;
  }
  router.push("/strategies");
}

function goMembership() {
  router.push("/membership");
}

watch(
  () => route.params.id,
  () => {
    if (route.params.id) {
      loadDetail();
    }
  }
);

onMounted(() => {
  loadDetail();
  loadVIPState();
});
</script>

<style scoped>
.forecast-run-page {
  display: grid;
  gap: 24px;
}

.forecast-run-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.8fr) minmax(280px, 0.9fr);
  gap: 20px;
}

.forecast-main-card,
.forecast-side-stack {
  display: grid;
  gap: 18px;
}

.forecast-summary-box,
.forecast-report-section {
  display: grid;
  gap: 16px;
}

.forecast-summary-text {
  margin: 0;
  color: var(--color-text-main);
  line-height: 1.7;
}

.forecast-report-body {
  margin: 0;
  padding: 16px;
  border-radius: 20px;
  background: rgba(13, 37, 84, 0.06);
  color: var(--color-text-main);
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.7;
}

.forecast-log-list,
.forecast-note-list {
  display: grid;
  gap: 12px;
}

.forecast-log-item {
  gap: 6px;
}

.forecast-log-item em {
  color: var(--color-text-muted);
  font-style: normal;
  font-size: 12px;
}

.forecast-lock-note {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.7;
}

@media (max-width: 960px) {
  .forecast-run-layout {
    grid-template-columns: 1fr;
  }
}
</style>
