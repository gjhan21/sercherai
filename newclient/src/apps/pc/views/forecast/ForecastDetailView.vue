<template>
  <section class="forecast-page">
    <div class="forecast-hero">
      <div class="hero-main">
        <p class="forecast-kicker">深度推演</p>
        <h1>深推演报告</h1>
        <p class="forecast-subtitle">{{ heroSubtitle }}</p>
        <div class="hero-badges">
          <span class="status-pill" :class="`tone-${statusTone}`">{{ statusLabel }}</span>
          <span class="meta-pill">{{ triggerTypeLabel }}</span>
          <span v-if="summaryConfidence" class="meta-pill">置信度 {{ summaryConfidence }}</span>
        </div>
      </div>
      <div class="hero-side">
        <div class="hero-times">
          <p><span>发起时间</span><strong>{{ formatDateTime(runMeta.queuedAt) }}</strong></p>
          <p><span>完成时间</span><strong>{{ formatDateTime(runMeta.finishedAt) }}</strong></p>
          <p><span>引擎</span><strong>{{ localizeForecastEngine(runMeta.engineKey) }}</strong></p>
        </div>
        <div class="forecast-actions">
          <button class="btn-primary" type="button" @click="loadDetail()">{{ loading ? "同步中..." : "刷新深推演" }}</button>
          <button class="btn-ghost" type="button" @click="goToLab">回深度推演工作台</button>
        </div>
      </div>
    </div>

    <section v-if="detailState === 'not_found'" class="forecast-card state-card">
      <div class="forecast-head">
        <strong>不存在或已失效</strong>
        <span class="forecast-status tone-failed">记录失效</span>
      </div>
      <p class="forecast-summary">
        这份深推演记录不存在、已失效，或者你当前没有权限查看。你可以回到深度推演工作台，或者回到来源页重新发起。
      </p>
      <div class="state-actions">
        <button class="btn-primary" type="button" @click="goToLab">回深度推演工作台</button>
        <button class="btn-ghost" type="button" @click="goBack">返回上一页</button>
      </div>
    </section>

    <section v-else-if="detailState === 'queued'" class="forecast-card state-card">
      <div class="forecast-head">
        <strong>已进入推演队列</strong>
        <span class="forecast-status tone-queued">排队中</span>
      </div>
      <p class="forecast-summary">
        系统已经收到这次深推演请求，正在排队生成报告。你可以留在此页等待自动刷新，也可以先回来源页继续查看推荐或策略。
      </p>
      <div class="forecast-grid">
        <article>
          <span>当前标的</span>
          <strong>{{ targetTitle }}</strong>
        </article>
        <article>
          <span>触发来源</span>
          <strong>{{ triggerTypeLabel }}</strong>
        </article>
        <article>
          <span>最近更新时间</span>
          <strong>{{ formatDateTime(lastUpdatedAt) }}</strong>
        </article>
      </div>
    </section>

    <section v-else-if="detailState === 'running'" class="forecast-card state-card">
      <div class="forecast-head">
        <strong>正在生成深推演报告</strong>
        <span class="forecast-status tone-running">推演中</span>
      </div>
      <p class="forecast-summary">
        推演已开始执行，系统会自动刷新状态。你现在可以先看运行证据区，了解已完成的阶段和最近一步处理结果。
      </p>
      <div class="forecast-grid">
        <article>
          <span>当前标的</span>
          <strong>{{ targetTitle }}</strong>
        </article>
        <article>
          <span>开始时间</span>
          <strong>{{ formatDateTime(runMeta.startedAt) }}</strong>
        </article>
        <article>
          <span>最近更新时间</span>
          <strong>{{ formatDateTime(lastUpdatedAt) }}</strong>
        </article>
      </div>
    </section>

    <section v-else-if="detailState === 'failed'" class="forecast-card state-card">
      <div class="forecast-head">
        <strong>本次深推演未完成</strong>
        <span class="forecast-status tone-failed">未完成</span>
      </div>
      <p class="forecast-summary">
        {{ failureNarrative }}
      </p>
      <div class="state-actions">
        <button class="btn-primary" type="button" @click="loadDetail()">重新同步状态</button>
        <button class="btn-ghost" type="button" @click="goToLab">回深度推演工作台</button>
      </div>
    </section>

    <template v-else-if="detailState === 'succeeded' || detailState === 'idle'">
      <section class="forecast-card">
        <div class="forecast-head">
          <strong>当前状态</strong>
          <span class="forecast-status" :class="`tone-${statusTone}`">{{ statusLabel }}</span>
        </div>
        <p class="forecast-summary">{{ stateAssessmentNarrative }}</p>
        <div class="forecast-grid">
          <article>
            <span>当前状态</span>
            <strong>{{ currentStateLabel }}</strong>
          </article>
          <article>
            <span>来源链路</span>
            <strong>{{ reportSourceLabel }}</strong>
          </article>
          <article>
            <span>上下文质量</span>
            <strong>{{ contextQualityLabel }}</strong>
          </article>
        </div>
      </section>

      <section class="forecast-card">
        <div class="forecast-head">
          <strong>核心判断</strong>
          <span class="forecast-meta">{{ targetTitle }}</span>
        </div>
        <p class="forecast-summary">{{ headlineVerdict }}</p>
        <div class="forecast-grid">
          <article>
            <span>一句话结论</span>
            <strong>{{ localizeForecastText(forecastSummary?.summary) || localizeForecastText(report?.executive_summary) || "等待结果..." }}</strong>
          </article>
          <article>
            <span>发起原因</span>
            <strong>{{ localizeForecastText(runMeta.reason) || "-" }}</strong>
          </article>
          <article>
            <span>会员边界</span>
            <strong>{{ reportRequiresVip && !isVipUser ? "正文锁定" : "可读" }}</strong>
          </article>
        </div>
      </section>

      <section class="forecast-card">
        <div class="forecast-head">
          <strong>主情景</strong>
          <span class="forecast-meta">{{ scenarioConsistencyLabel }}</span>
        </div>
        <p class="forecast-summary">{{ primaryScenarioNarrative }}</p>
        <div class="forecast-grid">
          <article>
            <span>主情景</span>
            <strong>{{ primaryScenarioLabel }}</strong>
          </article>
          <article>
            <span>次级情景</span>
            <strong>{{ secondaryScenarioText }}</strong>
          </article>
          <article>
            <span>触发条件</span>
            <strong>{{ triggerConditionText }}</strong>
          </article>
        </div>
      </section>

      <section class="forecast-card">
        <div class="forecast-head">
          <strong>风险边界</strong>
          <span class="forecast-meta">失效条件与风险复盘</span>
        </div>
        <p class="forecast-summary">{{ riskBoundaryText }}</p>
        <div class="forecast-grid">
          <article>
            <span>风险边界</span>
            <strong>{{ riskBoundaryText }}</strong>
          </article>
          <article>
            <span>失效信号</span>
            <strong>{{ invalidationConditionText }}</strong>
          </article>
          <article>
            <span>风险复核</span>
            <strong>{{ validationRiskReviewText }}</strong>
          </article>
        </div>
      </section>

      <section class="forecast-card">
        <div class="forecast-head">
          <strong>后续操作建议</strong>
          <span class="forecast-meta">行动计划与模型建议</span>
        </div>
        <p class="forecast-summary">{{ primaryActionGuidance }}</p>
        <div class="forecast-grid">
          <article>
            <span>动作建议</span>
            <strong>{{ actionPlanText }}</strong>
          </article>
          <article>
            <span>模型复核建议</span>
            <strong>{{ validationActionReviewText }}</strong>
          </article>
          <article>
            <span>完成时间</span>
            <strong>{{ formatDateTime(runMeta.finishedAt) }}</strong>
          </article>
        </div>
      </section>

      <section class="forecast-card">
        <div class="forecast-head">
          <strong>证据支撑</strong>
          <span class="forecast-meta">{{ evidenceSections.length }} 个研究维度</span>
        </div>
        <div v-if="evidenceSections.length" class="evidence-grid">
          <article v-for="item in evidenceSections" :key="item.key" class="evidence-card">
            <div class="evidence-head">
              <span>{{ item.label }}</span>
              <strong>{{ item.summary }}</strong>
            </div>
            <dl class="evidence-metrics">
              <div>
                <dt>当前立场</dt>
                <dd>{{ item.stanceLabel }}</dd>
              </div>
              <div>
                <dt>置信度</dt>
                <dd>{{ item.confidenceLabel }}</dd>
              </div>
              <div>
                <dt>支撑点</dt>
                <dd>{{ item.supportingText }}</dd>
              </div>
              <div>
                <dt>风险点</dt>
                <dd>{{ item.riskText }}</dd>
              </div>
            </dl>
          </article>
        </div>
        <p v-else class="forecast-empty">当前还没有补齐结构化维度证据。</p>
      </section>

      <section class="forecast-card">
        <div class="forecast-head">
          <strong>模型复核</strong>
          <span class="forecast-meta">{{ validationStatusLabel }}</span>
        </div>
        <p class="forecast-summary">{{ validationSummaryText }}</p>
        <div class="forecast-grid">
          <article>
            <span>模型结论</span>
            <strong>{{ validationVerdictText }}</strong>
          </article>
          <article>
            <span>支持证据</span>
            <strong>{{ validationSupportingText }}</strong>
          </article>
          <article>
            <span>反证与盲点</span>
            <strong>{{ validationCounterText }}</strong>
          </article>
        </div>
      </section>

      <section v-if="alternativeScenarios.length" class="forecast-card">
        <div class="forecast-head">
          <strong>备选情景</strong>
          <span class="forecast-meta">{{ alternativeScenarios.length }} 条</span>
        </div>
        <div class="forecast-grid">
          <article v-for="item in alternativeScenarios" :key="`${item.name}-${item.thesis}`">
            <span>{{ item.name }}</span>
            <strong>{{ item.probability }}</strong>
            <p class="forecast-summary">{{ item.thesis }}</p>
            <small class="forecast-meta">动作建议：{{ item.action }}</small>
          </article>
        </div>
      </section>

      <section v-if="triggerChecklist.length || invalidationSignals.length" class="forecast-card">
        <div class="forecast-head">
          <strong>验证清单</strong>
          <span class="forecast-meta">主线确认与失效边界</span>
        </div>
        <div class="forecast-grid">
          <article v-for="item in triggerChecklist" :key="`${item.label}-${item.trigger}`">
            <span>{{ item.label }}</span>
            <strong>{{ item.status }}</strong>
            <p class="forecast-summary">{{ item.note }}</p>
            <small class="forecast-meta">触发条件：{{ item.trigger }}</small>
          </article>
          <article v-if="invalidationSignals.length">
            <span>失效信号</span>
            <strong>{{ invalidationSignals.length }} 项</strong>
            <p class="forecast-summary">{{ invalidationSignals.join("；") }}</p>
          </article>
        </div>
      </section>

      <section v-if="roleDisagreements.length" class="forecast-card">
        <div class="forecast-head">
          <strong>角色分歧</strong>
          <span class="forecast-meta">多视角意见摘要</span>
        </div>
        <div class="forecast-grid">
          <article v-for="item in roleDisagreements" :key="`${item.role}-${item.summary}`">
            <span>{{ item.role }}</span>
            <strong>{{ item.stance }}</strong>
            <p class="forecast-summary">{{ item.summary }}</p>
            <small class="forecast-meta">{{ item.veto ? "包含否决意见" : "未触发否决" }}</small>
          </article>
        </div>
      </section>

      <section class="forecast-card">
        <div class="forecast-head">
          <strong>运行证据</strong>
          <span class="forecast-meta">{{ logs.length }} 个阶段</span>
        </div>
        <div v-if="logs.length" class="forecast-log-list">
          <article v-for="item in logs" :key="item.id || item.created_at || item.step_key" class="forecast-log-item">
            <div class="log-top">
              <span>{{ localizeForecastStepKey(item.step_key || "阶段") }}</span>
              <strong>{{ localizeForecastStatusText(item.status || "-") }}</strong>
            </div>
            <p>{{ localizeForecastText(item.message || "当前未补更多步骤说明。") }}</p>
            <small>{{ formatDateTime(item.created_at) }}</small>
          </article>
        </div>
        <p v-else class="forecast-empty">当前还没有步骤日志。</p>
      </section>

      <section class="forecast-card">
        <div class="forecast-head">
          <strong>原始报告全文</strong>
        </div>
        <p v-if="reportRequiresVip && !isVipUser" class="forecast-lock">当前报告正文需要 VIP 权限。摘要和关键状态仍可查看。</p>
        <pre v-else class="forecast-body">{{ localizeForecastText(report?.markdown_body) || "当前未生成正文。" }}</pre>
      </section>
    </template>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getMembershipQuota } from "@/api/membership.js";
import { useForecastRunDetail } from "@/shared/composables/useForecastRunDetail.js";
import {
  localizeForecastChecklistStatus,
  localizeForecastContextQuality,
  localizeForecastConfidence,
  localizeForecastEngine,
  localizeForecastProbability,
  localizeForecastScenarioName,
  localizeForecastSource,
  localizeForecastStatusText,
  localizeForecastStepKey,
  localizeForecastText,
  localizeForecastValidationStatus
} from "@/shared/lib/forecast-localization.js";
import { buildForecastEvidenceSections } from "@/shared/lib/forecast-report-view-model.js";
import { buildDeepForecastSummary } from "@/shared/lib/forecast-summary.js";

const route = useRoute();
const router = useRouter();
const runId = computed(() => String(route.params.id || ""));
const isVipUser = ref(false);

const {
  loading,
  errorMessage,
  run,
  report,
  logs,
  detailState,
  statusLabel,
  statusTone,
  lastUpdatedAt,
  runMeta,
  loadDetail
} = useForecastRunDetail(runId);

const forecastSummary = computed(() =>
  buildDeepForecastSummary({
    deep_forecast_summary: run.value?.summary,
    deep_forecast_report_ref: run.value?.report_ref
  })
);
const reportRequiresVip = computed(() => runMeta.value?.reportRequiresVip === true);
const targetTitle = computed(() => runMeta.value?.targetLabel || runMeta.value?.targetKey || "查看异步深推演结果与运行日志");
const summaryConfidence = computed(() => localizeForecastConfidence(run.value?.summary?.confidence_label));
const stateAssessment = computed(() => report.value?.state_assessment || null);
const scenarioAssessment = computed(() => report.value?.scenario_assessment || null);
const validationReview = computed(() => report.value?.validation_review || null);
const evidenceSections = computed(() =>
  buildForecastEvidenceSections({
    targetType: run.value?.target_type,
    dimensionEvidence: report.value?.dimension_evidence
  })
);
const headlineVerdict = computed(() =>
  localizeForecastText(report.value?.headline_verdict || report.value?.executive_summary || forecastSummary.value?.summary) || "当前未生成核心判断。"
);
const currentStateLabel = computed(() =>
  localizeForecastText(stateAssessment.value?.current_state || scenarioAssessment.value?.current_state || report.value?.primary_scenario) || "-"
);
const reportSourceLabel = computed(() =>
  localizeForecastSource(stateAssessment.value?.source || run.value?.source || runMeta.value?.triggerType) || "-"
);
const contextQualityLabel = computed(() =>
  localizeForecastContextQuality(stateAssessment.value?.context_quality || run.value?.context_quality) || "-"
);
const stateAssessmentNarrative = computed(() => {
  const parts = [
    `${targetTitle.value} 当前处于 ${currentStateLabel.value} 状态。`,
    reportSourceLabel.value !== "-" ? `来源链路为 ${reportSourceLabel.value}。` : "",
    contextQualityLabel.value !== "-" ? `上下文质量为 ${contextQualityLabel.value}。` : ""
  ].filter(Boolean);
  return parts.join("");
});
const primaryScenarioLabel = computed(() =>
  localizeForecastText(scenarioAssessment.value?.primary_scenario || forecastSummary.value?.scenario || report.value?.primary_scenario) || "-"
);
const secondaryScenarioText = computed(() =>
  formatBulletSummary(scenarioAssessment.value?.secondary_scenarios, "当前未补更多次级情景。")
);
const triggerConditionText = computed(() =>
  formatBulletSummary(scenarioAssessment.value?.trigger_conditions, "等待更多触发条件。")
);
const invalidationConditionText = computed(() =>
  formatBulletSummary(scenarioAssessment.value?.invalidation_conditions || report.value?.invalidation_signals, "当前未补更多失效信号。")
);
const actionPlanText = computed(() =>
  formatBulletSummary(scenarioAssessment.value?.action_plan || report.value?.action_guidance, primaryActionGuidance.value || "等待更多动作建议。")
);
const scenarioConsistencyLabel = computed(() =>
  localizeForecastText(validationReview.value?.scenario_consistency) || "等待模型复核"
);
const primaryScenarioNarrative = computed(() => {
  const parts = [
    `当前主情景为 ${primaryScenarioLabel.value}。`,
    scenarioConsistencyLabel.value ? `模型复核：${scenarioConsistencyLabel.value}。` : "",
    secondaryScenarioText.value && secondaryScenarioText.value !== "当前未补更多次级情景。" ? `次级情景包括 ${secondaryScenarioText.value}。` : ""
  ].filter(Boolean);
  return parts.join("");
});
const riskBoundaryText = computed(() =>
  localizeForecastText(stateAssessment.value?.risk_boundary) || formatBulletSummary(report.value?.invalidation_signals, "当前未补更多风险边界。")
);
const validationRiskReviewText = computed(() =>
  formatBulletSummary(validationReview.value?.risk_review, "当前未补更多风险复核。")
);
const validationActionReviewText = computed(() =>
  formatBulletSummary(validationReview.value?.action_review, "当前未补更多模型动作建议。")
);
const validationStatusLabel = computed(() =>
  localizeForecastValidationStatus(validationReview.value?.status || run.value?.validation_status) || "未触发模型复核"
);
const validationSummaryText = computed(() =>
  localizeForecastText(validationReview.value?.llm_summary) ||
  localizeForecastText(validationReview.value?.scenario_consistency) ||
  "当前未完成模型复核，不影响主报告阅读。"
);
const validationVerdictText = computed(() =>
  localizeForecastText(validationReview.value?.verdict) || "当前未补更多模型结论。"
);
const validationSupportingText = computed(() =>
  formatBulletSummary(validationReview.value?.supporting_evidence, "当前未补更多支持证据。")
);
const validationCounterText = computed(() => {
  const text = [
    formatBulletSummary(validationReview.value?.counter_evidence, ""),
    formatBulletSummary(validationReview.value?.blind_spots, "")
  ].filter(Boolean).join("；");
  return text || "当前未补更多反证或盲点。";
});
const primaryActionGuidance = computed(() => {
  const actions = report.value?.action_guidance;
  if (Array.isArray(actions) && actions.length) return localizeForecastText(actions[0]);
  return localizeForecastText(forecastSummary.value?.actionGuidance) || "-";
});
const alternativeScenarios = computed(() =>
  (Array.isArray(report.value?.alternative_scenarios) ? report.value.alternative_scenarios : []).map((item) => ({
    name: localizeForecastScenarioName(item?.name),
    probability: localizeForecastProbability(item?.probability),
    thesis: localizeForecastText(item?.thesis) || "当前未补更多情景说明。",
    action: localizeForecastText(item?.action) || "延续主情景观察。"
  }))
);
const triggerChecklist = computed(() =>
  (Array.isArray(report.value?.trigger_checklist) ? report.value.trigger_checklist : []).map((item) => ({
    label: localizeForecastText(item?.label) || "待观察项",
    status: localizeForecastChecklistStatus(item?.status),
    note: localizeForecastText(item?.note) || "当前未补更多验证说明。",
    trigger: localizeForecastText(item?.trigger) || "等待更多触发条件。"
  }))
);
const invalidationSignals = computed(() =>
  (Array.isArray(report.value?.invalidation_signals) ? report.value.invalidation_signals : [])
    .map((item) => localizeForecastText(item))
    .filter(Boolean)
);
const roleDisagreements = computed(() =>
  (Array.isArray(report.value?.role_disagreements) ? report.value.role_disagreements : []).map((item) => ({
    role: localizeForecastText(item?.role) || "角色",
    stance: localizeForecastText(item?.stance) || "中性",
    summary: localizeForecastText(item?.summary) || "当前未补更多分歧说明。",
    veto: item?.veto === true
  }))
);
const triggerTypeLabel = computed(() => {
  const type = String(runMeta.value?.triggerType || "").trim().toUpperCase();
  switch (type) {
    case "USER_REQUEST":
      return "主动发起";
    case "ADMIN_MANUAL":
      return "人工触发";
    case "AUTO_PRIORITY":
      return "系统触发";
    default:
      return "深推演";
  }
});
const heroSubtitle = computed(() => {
  if (detailState.value === "not_found") return "这份深推演记录不存在或已失效。";
  if (detailState.value === "queued") return `${targetTitle.value} 已进入深推演队列，等待系统生成报告。`;
  if (detailState.value === "running") return `${targetTitle.value} 正在生成深推演报告，系统会自动刷新当前状态。`;
  if (detailState.value === "failed") return `${targetTitle.value} 的深推演未能顺利完成，可查看原因并稍后重试。`;
  return targetTitle.value;
});
const failureNarrative = computed(() => {
  return localizeForecastText(errorMessage.value || run.value?.failure_reason) || "当前这次深推演没有顺利完成，建议回到深度推演工作台重新发起，或回来源页继续观察。";
});

async function loadVipState() {
  try {
    const quota = await getMembershipQuota();
    isVipUser.value = String(quota?.activation_state || "").toUpperCase() === "ACTIVE";
  } catch {
    isVipUser.value = false;
  }
}

function formatDateTime(value) {
  const raw = String(value || "").trim();
  if (!raw) return "-";
  const parsed = new Date(raw);
  if (Number.isNaN(parsed.getTime())) return raw;
  return parsed.toLocaleString("zh-CN", {
    hour12: false,
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  });
}

function formatBulletSummary(items, fallback = "-") {
  if (Array.isArray(items)) {
    const normalized = items.map((item) => localizeForecastText(item)).filter(Boolean);
    if (normalized.length) return normalized.join("；");
  }
  const text = localizeForecastText(items);
  return text || fallback;
}

function goBack() {
  router.back();
}

function goToLab() {
  router.push("/forecast-lab");
}

onMounted(() => {
  loadDetail();
  loadVipState();
});
</script>

<style scoped>
.forecast-page { display: grid; gap: 16px; max-width: 1100px; }
.forecast-hero, .forecast-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.forecast-hero { display: flex; justify-content: space-between; gap: 18px; align-items: flex-start; }
.hero-main { display: grid; gap: 8px; }
.forecast-kicker { font-size: 12px; color: var(--accent-gold); margin-bottom: 2px; text-transform: uppercase; letter-spacing: .08em; }
.forecast-hero h1 { font-size: 28px; font-weight: 800; }
.forecast-subtitle, .forecast-summary, .forecast-empty, .forecast-lock, .forecast-body { color: var(--text-secondary); line-height: 1.7; }
.hero-badges { display: flex; gap: 8px; flex-wrap: wrap; }
.status-pill, .meta-pill, .forecast-status { padding: 7px 11px; border-radius: var(--radius-full); font-size: 12px; border: 1px solid transparent; }
.meta-pill { border-color: var(--border); color: var(--text-secondary); background: rgba(255,255,255,.03); }
.hero-side { display: grid; gap: 14px; min-width: 290px; }
.hero-times { display: grid; gap: 8px; }
.hero-times p { display: flex; justify-content: space-between; gap: 12px; }
.hero-times span, .forecast-grid span, .forecast-log-item span { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: .06em; }
.hero-times strong, .forecast-grid strong, .forecast-log-item strong { text-align: right; }
.forecast-actions, .state-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.btn-primary, .btn-ghost { padding: 10px 18px; border-radius: var(--radius-full); font-size: 13px; font-weight: 600; cursor: pointer; }
.btn-primary { border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.btn-ghost { border: 1px solid var(--border-gold); background: none; color: var(--accent-gold); }
.forecast-head { display: flex; justify-content: space-between; align-items: center; gap: 12px; margin-bottom: 12px; }
.forecast-meta { font-size: 12px; color: var(--text-muted); }
.forecast-grid { display: grid; grid-template-columns: repeat(3,1fr); gap: 12px; }
.forecast-grid article, .forecast-log-item, .evidence-card { padding: 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.02); border: 1px solid var(--border); }
.forecast-grid article { display: grid; gap: 8px; align-content: start; }
.evidence-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12px; }
.evidence-card { display: grid; gap: 12px; }
.evidence-head { display: grid; gap: 8px; }
.evidence-head span { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: .06em; }
.evidence-head strong { line-height: 1.6; white-space: normal; word-break: break-word; }
.evidence-metrics { display: grid; gap: 10px; }
.evidence-metrics div { display: grid; gap: 6px; }
.evidence-metrics dt { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: .06em; }
.evidence-metrics dd { margin: 0; color: var(--text-secondary); line-height: 1.6; white-space: normal; word-break: break-word; }
.forecast-log-list { display: grid; gap: 10px; }
.log-top { display: flex; justify-content: space-between; gap: 10px; margin-bottom: 8px; }
.forecast-log-item p { color: var(--text-secondary); line-height: 1.6; margin-bottom: 6px; }
.forecast-log-item small { color: var(--text-muted); }
.forecast-grid strong { text-align: left; line-height: 1.6; white-space: normal; word-break: break-word; }
.forecast-grid small { line-height: 1.6; white-space: normal; word-break: break-word; }
.forecast-body { white-space: pre-wrap; font-family: inherit; }
.state-card { background: linear-gradient(135deg, rgba(255,255,255,.02), rgba(240,185,11,.04)); }
.tone-queued { color: var(--accent-gold); background: rgba(240,185,11,.08); border-color: rgba(240,185,11,.2); }
.tone-running { color: var(--accent-blue); background: rgba(59,130,246,.08); border-color: rgba(59,130,246,.2); }
.tone-success { color: var(--positive); background: rgba(0,200,151,.08); border-color: rgba(0,200,151,.2); }
.tone-failed { color: var(--negative); background: rgba(255,90,95,.08); border-color: rgba(255,90,95,.2); }
.tone-muted { color: var(--text-secondary); background: rgba(255,255,255,.04); border-color: var(--border); }

@media (max-width: 900px) {
  .forecast-hero,
  .forecast-grid,
  .evidence-grid {
    display: grid;
    grid-template-columns: 1fr;
  }

  .hero-side {
    min-width: 0;
  }
}
</style>
