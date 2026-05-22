<template>
  <section class="forecast-page">
    <div class="forecast-hero">
      <div>
        <p class="forecast-kicker">Forecast L3</p>
        <h1>深推演报告</h1>
        <p class="forecast-subtitle">{{ targetTitle }}</p>
      </div>
      <div class="forecast-actions">
        <button class="btn-primary" type="button" @click="loadDetail()">{{ loading ? "同步中..." : "刷新深推演" }}</button>
        <button class="btn-ghost" type="button" @click="goBack">返回上一页</button>
      </div>
    </div>

    <section class="forecast-card">
      <div class="forecast-head">
        <strong>执行摘要</strong>
        <span class="forecast-status">{{ statusLabel }}</span>
      </div>
      <p class="forecast-summary">{{ forecastSummary?.summary || errorMessage || "等待结果..." }}</p>
      <div class="forecast-grid">
        <article>
          <span>主情景</span>
          <strong>{{ forecastSummary?.scenario || "-" }}</strong>
        </article>
        <article>
          <span>动作建议</span>
          <strong>{{ forecastSummary?.actionGuidance || "-" }}</strong>
        </article>
        <article>
          <span>会员边界</span>
          <strong>{{ reportRequiresVip && !isVipUser ? "正文锁定" : "可读" }}</strong>
        </article>
      </div>
    </section>

    <section class="forecast-card">
      <div class="forecast-head">
        <strong>运行日志</strong>
      </div>
      <div v-if="logs.length" class="forecast-log-list">
        <article v-for="item in logs" :key="item.id || item.created_at || item.step_key" class="forecast-log-item">
          <span>{{ item.step_key || "STEP" }}</span>
          <strong>{{ item.status || "-" }}</strong>
          <p>{{ item.message || "当前未补更多步骤说明。" }}</p>
        </article>
      </div>
      <p v-else class="forecast-empty">当前还没有步骤日志。</p>
    </section>

    <section class="forecast-card">
      <div class="forecast-head">
        <strong>报告正文</strong>
      </div>
      <p v-if="reportRequiresVip && !isVipUser" class="forecast-lock">当前报告正文需要 VIP 权限。摘要和关键状态仍可查看。</p>
      <pre v-else class="forecast-body">{{ report?.markdown_body || "当前未生成正文。" }}</pre>
    </section>
  </section>
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
const targetTitle = computed(() => run.value?.target_label || run.value?.target_key || "查看异步深推演结果与运行日志");

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
.forecast-page { display: grid; gap: 16px; max-width: 1100px; }
.forecast-hero, .forecast-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.forecast-hero { display: flex; justify-content: space-between; gap: 16px; align-items: flex-start; }
.forecast-kicker { font-size: 12px; color: var(--accent-gold); margin-bottom: 6px; text-transform: uppercase; letter-spacing: .08em; }
.forecast-hero h1 { font-size: 28px; font-weight: 800; margin-bottom: 8px; }
.forecast-subtitle { color: var(--text-secondary); line-height: 1.7; }
.forecast-actions { display: flex; gap: 10px; }
.btn-primary, .btn-ghost { padding: 10px 18px; border-radius: var(--radius-full); font-size: 13px; font-weight: 600; cursor: pointer; }
.btn-primary { border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.btn-ghost { border: 1px solid var(--border-gold); background: none; color: var(--accent-gold); }
.forecast-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.forecast-status { font-size: 12px; color: var(--accent-gold); }
.forecast-summary { color: var(--text-secondary); line-height: 1.8; margin-bottom: 14px; }
.forecast-grid { display: grid; grid-template-columns: repeat(3,1fr); gap: 12px; }
.forecast-grid article, .forecast-log-item { padding: 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.02); border: 1px solid var(--border); }
.forecast-grid span, .forecast-log-item span { display: block; font-size: 11px; color: var(--text-muted); margin-bottom: 6px; text-transform: uppercase; }
.forecast-grid strong, .forecast-log-item strong { display: block; margin-bottom: 6px; }
.forecast-log-list { display: grid; gap: 10px; }
.forecast-log-item p { color: var(--text-secondary); line-height: 1.6; }
.forecast-empty, .forecast-lock { color: var(--text-secondary); line-height: 1.7; }
.forecast-body { white-space: pre-wrap; color: var(--text-secondary); line-height: 1.7; font-family: inherit; }
@media (max-width: 900px) { .forecast-hero, .forecast-grid { grid-template-columns: 1fr; display: grid; } .forecast-actions { flex-wrap: wrap; } }
</style>
