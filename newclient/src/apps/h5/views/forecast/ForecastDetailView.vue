<template>
  <div class="h5-forecast-page">
    <div class="h5-forecast-hero">
      <p class="h5-kicker">深度推演</p>
      <h1>深度推演报告</h1>
      <p class="h5-subtitle">{{ heroSubtitle }}</p>
      <div class="h5-badges">
        <span class="h5-status" :class="`tone-${statusTone}`">{{ statusLabel }}</span>
        <span class="h5-chip">{{ triggerTypeLabel }}</span>
      </div>
      <div class="h5-meta-grid">
        <p><span>发起时间</span><strong>{{ formatDateTime(runMeta.queuedAt) }}</strong></p>
        <p><span>完成时间</span><strong>{{ formatDateTime(runMeta.finishedAt) }}</strong></p>
      </div>
      <div class="h5-action-row">
        <button type="button" class="h5-btn" @click="loadDetail()">{{ loading ? "同步中..." : "刷新状态" }}</button>
        <button type="button" class="h5-btn-ghost" @click="goToLab">回推演工作台</button>
      </div>
    </div>

    <div v-if="detailState === 'not_found'" class="h5-card">
      <div class="h5-card-head">
        <strong>不存在或已失效</strong>
      </div>
      <p class="h5-copy">这份深推演记录不存在、已失效，或者你当前没有权限查看。可以回到工作台重新发起，或返回上一页继续查看来源内容。</p>
      <div class="h5-action-row compact">
        <button type="button" class="h5-btn" @click="goToLab">回工作台</button>
        <button type="button" class="h5-btn-ghost" @click="goBack">返回上一页</button>
      </div>
    </div>

    <div v-else-if="detailState === 'queued'" class="h5-card">
      <div class="h5-card-head">
        <strong>已进入推演队列</strong>
      </div>
      <p class="h5-copy">系统已经接收这次深推演请求，正在排队生成报告。你可以留在此页等待刷新，也可以先回来源页继续浏览。</p>
    </div>

    <div v-else-if="detailState === 'running'" class="h5-card">
      <div class="h5-card-head">
        <strong>正在生成深推演报告</strong>
      </div>
      <p class="h5-copy">推演已经开始执行，系统会自动刷新。你可以先看下面的运行证据，了解当前完成到了哪一步。</p>
    </div>

    <div v-else-if="detailState === 'failed'" class="h5-card">
      <div class="h5-card-head">
        <strong>本次深推演未完成</strong>
      </div>
      <p class="h5-copy">{{ failureNarrative }}</p>
    </div>

    <template v-else-if="detailState === 'succeeded' || detailState === 'idle'">
      <div class="h5-card">
        <div class="h5-card-head">
          <strong>当前状态</strong>
          <span>{{ statusLabel }}</span>
        </div>
        <p class="h5-copy">{{ stateAssessmentNarrative }}</p>
        <div class="h5-summary-grid">
          <div>
            <span>当前状态</span>
            <strong>{{ currentStateLabel }}</strong>
          </div>
          <div>
            <span>来源链路</span>
            <strong>{{ reportSourceLabel }}</strong>
          </div>
          <div>
            <span>上下文质量</span>
            <strong>{{ contextQualityLabel }}</strong>
          </div>
        </div>
      </div>

      <div class="h5-card">
        <div class="h5-card-head">
          <strong>核心判断</strong>
          <span>{{ targetTitle }}</span>
        </div>
        <p class="h5-copy">{{ headlineVerdict }}</p>
        <div class="h5-summary-grid">
          <div>
            <span>一句话结论</span>
            <strong>{{ localizeForecastText(forecastSummary?.summary) || localizeForecastText(report?.executive_summary) || "等待结果..." }}</strong>
          </div>
          <div>
            <span>发起原因</span>
            <strong>{{ localizeForecastText(runMeta.reason) || "-" }}</strong>
          </div>
          <div>
            <span>会员边界</span>
            <strong>{{ reportRequiresVip && !isVipUser ? "正文锁定" : "可读" }}</strong>
          </div>
        </div>
      </div>

      <div class="h5-card">
        <div class="h5-card-head">
          <strong>主情景</strong>
          <span>{{ scenarioConsistencyLabel }}</span>
        </div>
        <p class="h5-copy">{{ primaryScenarioNarrative }}</p>
        <div class="h5-summary-grid">
          <div>
            <span>主情景</span>
            <strong>{{ primaryScenarioLabel }}</strong>
          </div>
          <div>
            <span>次级情景</span>
            <strong>{{ secondaryScenarioText }}</strong>
          </div>
          <div>
            <span>触发条件</span>
            <strong>{{ triggerConditionText }}</strong>
          </div>
        </div>
      </div>

      <div class="h5-card">
        <div class="h5-card-head">
          <strong>风险边界</strong>
          <span>失效条件与风险复盘</span>
        </div>
        <p class="h5-copy">{{ riskBoundaryText }}</p>
        <div class="h5-summary-grid">
          <div>
            <span>风险边界</span>
            <strong>{{ riskBoundaryText }}</strong>
          </div>
          <div>
            <span>失效信号</span>
            <strong>{{ invalidationConditionText }}</strong>
          </div>
          <div>
            <span>风险复核</span>
            <strong>{{ validationRiskReviewText }}</strong>
          </div>
        </div>
      </div>

      <div class="h5-card">
        <div class="h5-card-head">
          <strong>后续操作建议</strong>
          <span>行动计划与模型建议</span>
        </div>
        <p class="h5-copy">{{ primaryActionGuidance }}</p>
        <div class="h5-summary-grid">
          <div>
            <span>动作建议</span>
            <strong>{{ actionPlanText }}</strong>
          </div>
          <div>
            <span>模型复核建议</span>
            <strong>{{ validationActionReviewText }}</strong>
          </div>
          <div>
            <span>完成时间</span>
            <strong>{{ formatDateTime(runMeta.finishedAt) }}</strong>
          </div>
        </div>
      </div>

      <div class="h5-card">
        <div class="h5-card-head">
          <strong>证据支撑</strong>
          <span>{{ evidenceSections.length }} 个研究维度</span>
        </div>
        <div v-if="evidenceSections.length" class="h5-evidence-list">
          <div v-for="item in evidenceSections" :key="item.key" class="h5-log-item">
            <strong>{{ item.label }}</strong>
            <p>{{ item.summary }}</p>
            <div class="h5-evidence-grid">
              <div>
                <span>当前立场</span>
                <strong>{{ item.stanceLabel }}</strong>
              </div>
              <div>
                <span>置信度</span>
                <strong>{{ item.confidenceLabel }}</strong>
              </div>
              <div>
                <span>支撑点</span>
                <strong>{{ item.supportingText }}</strong>
              </div>
              <div>
                <span>风险点</span>
                <strong>{{ item.riskText }}</strong>
              </div>
            </div>
          </div>
        </div>
        <p v-else class="h5-copy">当前还没有补齐结构化维度证据。</p>
      </div>

      <div class="h5-card">
        <div class="h5-card-head">
          <strong>模型复核</strong>
          <span>{{ validationStatusLabel }}</span>
        </div>
        <p class="h5-copy">{{ validationSummaryText }}</p>
        <div class="h5-summary-grid">
          <div>
            <span>模型结论</span>
            <strong>{{ validationVerdictText }}</strong>
          </div>
          <div>
            <span>支持证据</span>
            <strong>{{ validationSupportingText }}</strong>
          </div>
          <div>
            <span>反证与盲点</span>
            <strong>{{ validationCounterText }}</strong>
          </div>
        </div>
      </div>

      <div v-if="alternativeScenarios.length" class="h5-card">
        <div class="h5-card-head">
          <strong>备选情景</strong>
          <span>{{ alternativeScenarios.length }} 条</span>
        </div>
        <div class="h5-log-list">
          <div v-for="item in alternativeScenarios" :key="`${item.name}-${item.thesis}`" class="h5-log-item">
            <strong>{{ item.name }}</strong>
            <span>{{ item.probability }}</span>
            <p>{{ item.thesis }}</p>
            <small>动作建议：{{ item.action }}</small>
          </div>
        </div>
      </div>

      <div v-if="triggerChecklist.length" class="h5-card">
        <div class="h5-card-head">
          <strong>验证清单</strong>
          <span>{{ triggerChecklist.length }} 项</span>
        </div>
        <div class="h5-log-list">
          <div v-for="item in triggerChecklist" :key="`${item.label}-${item.trigger}`" class="h5-log-item">
            <strong>{{ item.label }}</strong>
            <span>{{ item.status }}</span>
            <p>{{ item.note }}</p>
            <small>触发条件：{{ item.trigger }}</small>
          </div>
        </div>
      </div>

      <div v-if="invalidationSignals.length" class="h5-card">
        <div class="h5-card-head">
          <strong>失效信号</strong>
          <span>{{ invalidationSignals.length }} 项</span>
        </div>
        <div class="h5-log-list">
          <div v-for="item in invalidationSignals" :key="item" class="h5-log-item">
            <p>{{ item }}</p>
          </div>
        </div>
      </div>

      <div v-if="roleDisagreements.length" class="h5-card">
        <div class="h5-card-head">
          <strong>角色分歧</strong>
          <span>{{ roleDisagreements.length }} 条</span>
        </div>
        <div class="h5-log-list">
          <div v-for="item in roleDisagreements" :key="`${item.role}-${item.summary}`" class="h5-log-item">
            <strong>{{ item.role }}</strong>
            <span>{{ item.stance }}</span>
            <p>{{ item.summary }}</p>
            <small>{{ item.veto ? "包含否决意见" : "未触发否决" }}</small>
          </div>
        </div>
      </div>

      <div class="h5-card">
        <div class="h5-card-head">
          <strong>运行证据</strong>
          <span>{{ logs.length }} 条</span>
        </div>
        <div v-if="logs.length" class="h5-log-list">
          <div v-for="item in logs" :key="item.id || item.created_at || item.step_key" class="h5-log-item">
            <strong>{{ localizeForecastStepKey(item.step_key || "阶段") }}</strong>
            <span>{{ localizeForecastStatusText(item.status || "-") }}</span>
            <p>{{ localizeForecastText(item.message || "当前未补更多步骤说明。") }}</p>
            <small>{{ formatDateTime(item.created_at) }}</small>
          </div>
        </div>
        <p v-else class="h5-copy">当前还没有步骤日志。</p>
      </div>

      <div class="h5-card">
        <div class="h5-card-head">
          <strong>原始报告全文</strong>
        </div>
        <p v-if="reportRequiresVip && !isVipUser" class="h5-copy">当前报告正文需要 VIP 权限。摘要和关键状态仍可查看。</p>
        <pre v-else class="h5-body">{{ localizeForecastText(report?.markdown_body) || "当前未生成正文。" }}</pre>
      </div>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getMembershipQuota } from "@/api/membership.js";
import { useForecastRunDetail } from "@/shared/composables/useForecastRunDetail.js";
import {
  localizeForecastChecklistStatus,
  localizeForecastContextQuality,
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
const targetTitle = computed(() => runMeta.value?.targetLabel || runMeta.value?.targetKey || "查看异步推演的运行状态与报告");
const stateAssessment = computed(() => report.value?.state_assessment || null);
const scenarioAssessment = computed(() => report.value?.scenario_assessment || null);
const validationReview = computed(() => report.value?.validation_review || null);
const evidenceSections = computed(() =>
  buildForecastEvidenceSections({
    targetType: run.value?.target_type,
    dimensionEvidence: report.value?.dimension_evidence
  })
);
const primaryActionGuidance = computed(() => {
  const actions = report.value?.action_guidance;
  if (Array.isArray(actions) && actions.length) return localizeForecastText(actions[0]);
  return localizeForecastText(forecastSummary.value?.actionGuidance) || "-";
});
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
  if (detailState.value === "running") return `${targetTitle.value} 正在生成深推演报告，系统会自动刷新。`;
  if (detailState.value === "failed") return `${targetTitle.value} 的深推演未能顺利完成。`;
  return targetTitle.value;
});
const failureNarrative = computed(() => {
  return localizeForecastText(errorMessage.value || run.value?.failure_reason) || "当前这次深推演没有顺利完成，建议回到工作台重新发起，或回来源页继续观察。";
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
.h5-forecast-page { display: grid; gap: 14px; }
.h5-forecast-hero, .h5-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 16px; }
.h5-kicker { color: var(--accent-gold); font-size: 11px; text-transform: uppercase; letter-spacing: .08em; margin-bottom: 6px; }
.h5-forecast-hero h1 { font-size: 22px; font-weight: 800; margin-bottom: 8px; }
.h5-subtitle, .h5-copy, .h5-body { color: var(--text-secondary); line-height: 1.7; }
.h5-badges { display: flex; gap: 8px; flex-wrap: wrap; margin: 10px 0; }
.h5-status, .h5-chip { padding: 6px 10px; border-radius: var(--radius-full); font-size: 11px; border: 1px solid transparent; }
.h5-chip { border-color: var(--border); color: var(--text-secondary); background: rgba(255,255,255,.03); }
.h5-meta-grid { display: grid; gap: 8px; }
.h5-meta-grid p { display: flex; justify-content: space-between; gap: 12px; }
.h5-meta-grid span, .h5-summary-grid span, .h5-log-item span { font-size: 11px; color: var(--text-muted); text-transform: uppercase; }
.h5-action-row { display: grid; gap: 10px; margin-top: 14px; }
.h5-action-row.compact { margin-top: 12px; }
.h5-btn, .h5-btn-ghost { width: 100%; padding: 12px; border-radius: var(--radius-full); font-size: 14px; font-weight: 600; cursor: pointer; }
.h5-btn { border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.h5-btn-ghost { border: 1px solid var(--border-gold); background: none; color: var(--accent-gold); }
.h5-card-head { display: flex; justify-content: space-between; align-items: center; gap: 10px; margin-bottom: 10px; }
.h5-summary-grid, .h5-log-list { display: grid; gap: 10px; }
.h5-summary-grid div, .h5-log-item { padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border); background: rgba(255,255,255,.02); }
.h5-summary-grid div { display: grid; gap: 8px; align-content: start; }
.h5-summary-grid strong { line-height: 1.6; white-space: normal; word-break: break-word; }
.h5-evidence-list { display: grid; gap: 10px; }
.h5-evidence-grid { display: grid; gap: 10px; margin-top: 10px; }
.h5-evidence-grid div { display: grid; gap: 6px; padding: 10px; border-radius: var(--radius-sm); border: 1px solid var(--border); background: rgba(255,255,255,.02); }
.h5-evidence-grid strong { line-height: 1.6; white-space: normal; word-break: break-word; }
.h5-log-item strong { display: block; margin-bottom: 4px; }
.h5-log-item p { color: var(--text-secondary); line-height: 1.6; margin: 6px 0; }
.h5-log-item small { color: var(--text-muted); }
.h5-body { white-space: pre-wrap; font-family: inherit; }
.tone-queued { color: var(--accent-gold); background: rgba(240,185,11,.08); border-color: rgba(240,185,11,.2); }
.tone-running { color: var(--accent-blue); background: rgba(59,130,246,.08); border-color: rgba(59,130,246,.2); }
.tone-success { color: var(--positive); background: rgba(0,200,151,.08); border-color: rgba(0,200,151,.2); }
.tone-failed { color: var(--negative); background: rgba(255,90,95,.08); border-color: rgba(255,90,95,.2); }
.tone-muted { color: var(--text-secondary); background: rgba(255,255,255,.04); border-color: var(--border); }
</style>
