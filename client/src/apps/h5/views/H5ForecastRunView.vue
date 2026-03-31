<template>
  <div class="h5-page fade-up">
    <div class="h5-page-topline">
      <span class="h5-page-tagline">深度推演报告</span>
      <span>Forecast L3</span>
    </div>

    <header class="h5-card h5-card-brand h5-card-hero h5-hero-card">
      <div class="h5-card-body">
        <div class="h5-section-head">
          <p class="h5-eyebrow">当前状态: {{ forecastStatusLabel }}</p>
          <h1 class="h5-title">{{ targetTitle }}</h1>
          <p class="h5-subtitle">查看异步推演的运行状态、结构化建议与后续发展预测。</p>
        </div>

        <div class="h5-meta-list hero-meta-list">
          <span class="h5-meta-chip">运行编号: {{ run?.id || route.params.id }}</span>
          <span v-if="run?.engine_key" class="h5-meta-chip">{{ run.engine_key }}</span>
        </div>

        <div class="h5-action-row" style="margin-top: 20px;">
          <button type="button" class="h5-btn block" :disabled="loading" @click="loadDetail">
            {{ loading ? "同步中..." : "刷新状态" }}
          </button>
          <button type="button" class="h5-btn-ghost block" @click="goBack">返回上一页</button>
        </div>
      </div>
    </header>

    <section class="h5-section-block">
      <div class="h5-section-head">
        <p class="h5-eyebrow">运行摘要</p>
        <h2 class="h5-title" style="font-size: 20px;">关键运行指标</h2>
      </div>
      <div class="h5-grid-2">
        <article class="h5-metric-card h5-summary-card">
          <p class="h5-summary-label">当前进度</p>
          <strong class="h5-summary-value">{{ forecastStatusLabel }}</strong>
          <small class="h5-summary-note">{{ forecastSummary?.summary || "等待运行..." }}</small>
        </article>
        <article class="h5-metric-card h5-summary-card">
          <p class="h5-summary-label">主情景预判</p>
          <strong class="h5-summary-value">{{ forecastSummary?.scenario || "-" }}</strong>
          <small class="h5-summary-note">{{ forecastSummary?.actionGuidance || "收集论据中..." }}</small>
        </article>
      </div>
    </section>

    <template v-if="report">
      <section class="h5-section-block">
        <div class="h5-section-head">
          <p class="h5-eyebrow">核心结论</p>
          <h2 class="h5-title" style="font-size: 20px;">推演执行建议</h2>
        </div>
        <div class="h5-list">
          <article class="h5-list-card h5-card-soft">
            <div class="h5-list-topline">
              <strong>执行摘要</strong>
              <span class="h5-badge brand">{{ report.primary_scenario || "-" }}</span>
            </div>
            <p class="h5-list-note">{{ report.executive_summary || "暂无详细摘要" }}</p>
          </article>

          <article v-if="report.action_guidance?.length" class="h5-list-card">
            <div class="h5-list-topline">
              <strong>动作建议</strong>
            </div>
            <ul class="h5-list-note" style="padding-left: 18px; margin: 0;">
              <li v-for="act in report.action_guidance" :key="act">{{ act }}</li>
            </ul>
          </article>
        </div>
      </section>

      <section v-if="report.alternative_scenarios?.length" class="h5-section-block">
        <div class="h5-section-head">
          <p class="h5-eyebrow">后续发展预测</p>
          <h2 class="h5-title" style="font-size: 20px;">备选剧本推演</h2>
        </div>
        <div class="h5-list">
          <article v-for="alt in report.alternative_scenarios" :key="alt.name" class="h5-list-card h5-card-accent">
            <div class="h5-list-topline">
              <strong>{{ alt.name }}</strong>
              <span class="h5-badge gold">概率 {{ (alt.probability * 100).toFixed(0) }}%</span>
            </div>
            <p class="h5-list-note"><strong>逻辑:</strong> {{ alt.thesis }}</p>
            <p class="h5-list-note" style="color: var(--h5-warning); font-weight: 700;"><strong>指引:</strong> {{ alt.action }}</p>
          </article>
        </div>
      </section>

      <section class="h5-section-block">
        <div class="h5-section-head">
          <p class="h5-eyebrow">报告正文</p>
        </div>
        <div class="h5-card">
          <div class="h5-card-body h5-body-html">
             <p v-if="reportRequiresVIP && !isVIPUser" class="h5-lock-card">
               <strong>会员内容锁定</strong>
               <span>当前报告正文需要 VIP 权限。升级后可查看完整推演逻辑。</span>
               <button type="button" class="h5-btn block" style="margin-top: 10px;" @click="goMembership">升级会员</button>
             </p>
             <div v-else class="h5-report-markdown">
               <pre style="white-space: pre-wrap; font-family: inherit; font-size: 14px; line-height: 1.6; color: var(--h5-text-sub);">{{ report.markdown_body || "正文生成中..." }}</pre>
             </div>
          </div>
        </div>
      </section>
    </template>

    <StatePanel
      v-if="!report && !loading && !errorMessage"
      tone="info"
      :title="statePanelTitle"
      :description="statePanelDescription"
    >
      <template #actions>
        <button type="button" class="h5-btn" @click="loadDetail">刷新重试</button>
      </template>
    </StatePanel>

    <div v-if="logs.length" class="h5-section-block">
      <div class="h5-section-head">
        <p class="h5-eyebrow">运行日志</p>
      </div>
      <div class="h5-list">
        <div v-for="log in logs.slice().reverse()" :key="log.id" class="h5-list-card h5-card-soft" style="padding: 12px; gap: 4px;">
           <div class="h5-list-topline" style="font-size: 11px;">
             <strong>{{ log.step_key }}</strong>
             <span>{{ formatDateTime(log.created_at) }}</span>
           </div>
           <p style="margin:0; font-size: 12px; color: var(--h5-text-sub);">{{ log.message }}</p>
        </div>
      </div>
    </div>

    <div class="h5-sticky-cta-space"></div>
  </div>
</template>

<script setup>
import { computed, onMounted, onBeforeUnmount, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import StatePanel from "../../../components/StatePanel.vue";
import { getForecastRunDetail } from "../../../api/forecast";
import { getMembershipQuota } from "../../../api/membership";
import { buildStrategyDeepForecastSummary } from "../../../lib/strategy-version";

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
const forecastStatusLabel = computed(() => forecastSummary.value?.statusLabel || "等待结果");
const targetTitle = computed(() => run.value?.target_label || run.value?.target_key || "深度推演报告");
const reportRequiresVIP = computed(() => run.value?.report_ref?.requires_vip === true);

const statePanelTitle = computed(() => {
  if (forecastSummary.value?.status === "FAILED") return "推演未成功完成";
  if (forecastSummary.value?.status === "RUNNING") return "AI 正在深度推演中";
  return "等待报告同步";
});
const statePanelDescription = computed(() => run.value?.failure_reason || forecastSummary.value?.summary || "推演正在异步执行中，请稍后刷新。");

let pollTimer = null;
function cleanPoll() { if (pollTimer) { clearTimeout(pollTimer); pollTimer = null; } }

async function loadDetail(quiet = false) {
  if (!quiet) loading.value = true;
  try {
    const data = await getForecastRunDetail(route.params.id);
    detail.value = data || null;
    const status = data?.run?.status || "";
    if (status === "QUEUED" || status === "RUNNING") {
      cleanPoll();
      pollTimer = setTimeout(() => loadDetail(true), 5000);
    }
  } catch (err) {
    errorMessage.value = err?.message || "同步失败";
  } finally {
    if (!quiet) loading.value = false;
  }
}

async function loadVIPState() {
  try {
    const quota = await getMembershipQuota();
    const activationState = String(quota?.activation_state || "").toUpperCase();
    isVIPUser.value = (activationState === "ACTIVE");
  } catch { isVIPUser.value = false; }
}

function formatDateTime(v) {
  if (!v) return "";
  const d = new Date(v);
  return `${d.getHours().toString().padStart(2,"0")}:${d.getMinutes().toString().padStart(2,"0")}`;
}

function goBack() {
  const from = String(route.query.from || "");
  if (from.startsWith("/")) router.push(from);
  else router.push("/strategies");
}
function goMembership() { router.push("/membership"); }

onMounted(() => { loadDetail(); loadVIPState(); });
onBeforeUnmount(() => { cleanPoll(); });
</script>

<style scoped>
.h5-report-markdown {
  overflow-x: hidden;
}
</style>
