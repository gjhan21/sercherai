<template>
  <div class="h5-forecast-page">
    <div class="h5-forecast-hero">
      <p class="h5-kicker">Forecast L3</p>
      <h1>深度推演报告</h1>
      <p class="h5-subtitle">{{ targetTitle }}</p>
      <div class="h5-action-row">
        <button type="button" class="h5-btn" @click="loadDetail()">{{ loading ? "同步中..." : "刷新状态" }}</button>
        <button type="button" class="h5-btn-ghost" @click="goBack">返回上一页</button>
      </div>
    </div>

    <div class="h5-card">
      <div class="h5-card-head">
        <strong>执行摘要</strong>
        <span>{{ statusLabel }}</span>
      </div>
      <p class="h5-copy">{{ forecastSummary?.summary || errorMessage || "等待结果..." }}</p>
    </div>

    <div class="h5-card">
      <div class="h5-card-head">
        <strong>运行日志</strong>
      </div>
      <div v-if="logs.length" class="h5-log-list">
        <div v-for="item in logs" :key="item.id || item.created_at || item.step_key" class="h5-log-item">
          <strong>{{ item.step_key || "STEP" }}</strong>
          <span>{{ item.status || "-" }}</span>
          <p>{{ item.message || "当前未补更多步骤说明。" }}</p>
        </div>
      </div>
      <p v-else class="h5-copy">当前还没有步骤日志。</p>
    </div>

    <div class="h5-card">
      <div class="h5-card-head">
        <strong>报告正文</strong>
      </div>
      <p v-if="reportRequiresVip && !isVipUser" class="h5-copy">当前报告正文需要 VIP 权限。摘要和关键状态仍可查看。</p>
      <pre v-else class="h5-body">{{ report?.markdown_body || "当前未生成正文。" }}</pre>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getMembershipQuota } from "@/api/membership.js";
import { useForecastRunDetail } from "@/shared/composables/useForecastRunDetail.js";
import { buildDeepForecastSummary } from "@/shared/lib/forecast-summary.js";

const route = useRoute();
const router = useRouter();
const runId = computed(() => String(route.params.id || ""));
const isVipUser = ref(false);

const { loading, errorMessage, run, report, logs, loadDetail } = useForecastRunDetail(runId);

const forecastSummary = computed(() =>
  buildDeepForecastSummary({
    deep_forecast_summary: run.value?.summary,
    deep_forecast_report_ref: run.value?.report_ref
  })
);
const statusLabel = computed(() => forecastSummary.value?.statusLabel || "等待结果");
const reportRequiresVip = computed(() => run.value?.report_ref?.requires_vip === true);
const targetTitle = computed(() => run.value?.target_label || run.value?.target_key || "查看异步推演的运行状态与报告");

async function loadVipState() {
  try {
    const quota = await getMembershipQuota();
    isVipUser.value = String(quota?.activation_state || "").toUpperCase() === "ACTIVE";
  } catch {
    isVipUser.value = false;
  }
}

function goBack() {
  router.back();
}

onMounted(() => {
  loadDetail();
  loadVipState();
});
</script>

<style scoped>
.h5-forecast-page { display: grid; gap: 14px; }
.h5-forecast-hero, .h5-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 16px; }
.h5-kicker { color: var(--accent-gold); font-size: 11px; text-transform: uppercase; letter-spacing: .08em; margin-bottom: 6px; }
.h5-forecast-hero h1 { font-size: 22px; font-weight: 800; margin-bottom: 8px; }
.h5-subtitle, .h5-copy { color: var(--text-secondary); line-height: 1.7; }
.h5-action-row { display: grid; gap: 10px; margin-top: 14px; }
.h5-btn, .h5-btn-ghost { width: 100%; padding: 12px; border-radius: var(--radius-full); font-size: 14px; font-weight: 600; cursor: pointer; }
.h5-btn { border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.h5-btn-ghost { border: 1px solid var(--border-gold); background: none; color: var(--accent-gold); }
.h5-card-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.h5-card-head span { color: var(--accent-gold); font-size: 12px; }
.h5-log-list { display: grid; gap: 10px; }
.h5-log-item { padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border); background: rgba(255,255,255,.02); }
.h5-log-item strong { display: block; margin-bottom: 4px; }
.h5-log-item span { display: block; color: var(--text-muted); font-size: 11px; margin-bottom: 6px; text-transform: uppercase; }
.h5-log-item p { color: var(--text-secondary); line-height: 1.6; }
.h5-body { white-space: pre-wrap; font-family: inherit; color: var(--text-secondary); line-height: 1.7; }
</style>
