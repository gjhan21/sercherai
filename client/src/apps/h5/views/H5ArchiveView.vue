<template>
  <div class="h5-page fade-up archive-page">
    <div class="h5-page-topline">
      <span class="h5-page-tagline">历史档案</span>
      <span>{{ archiveStageLabel }}</span>
    </div>

    <H5HeroCard
      eyebrow="复盘中心"
      :title="heroTitle"
      :description="heroDescription"
      :meta="heroMeta"
      tone="accent"
    >
      <template #actions>
        <H5ActionBar>
          <button type="button" class="h5-btn-secondary" :disabled="loading" @click="loadArchivePage">
            {{ loading ? "同步中..." : "刷新档案" }}
          </button>
          <button type="button" class="h5-btn-ghost" @click="goNews">回资讯页</button>
        </H5ActionBar>
      </template>

      <div class="archive-hero-metrics h5-grid-3">
        <H5SummaryCard
          v-for="(item, index) in archiveModel.summaryCards"
          :key="item.label"
          :label="item.label"
          :value="item.value"
          :note="item.note"
          :tone="resolveHighlightTone(index)"
        />
      </div>
    </H5HeroCard>

    <H5SectionBlock eyebrow="阅读顺序" title="先看结果，再看理由和来源" tone="soft">
      <div class="archive-read-order">
        <article v-for="item in archiveModel.readOrder" :key="item.title" class="archive-order-item">
          <span>{{ item.title }}</span>
          <p>{{ item.desc }}</p>
        </article>
      </div>
    </H5SectionBlock>

    <H5SectionBlock eyebrow="优先复盘" title="先处理当前阶段最值得看的历史样本">
      <div v-if="archiveModel.leadItem" class="archive-lead-card">
        <div class="archive-item-topline">
          <div>
            <span>{{ archiveModel.leadItem.date }}</span>
            <strong>{{ archiveModel.leadItem.headline }}</strong>
          </div>
          <div class="archive-badge-row">
            <span class="archive-badge" :class="`status-${archiveModel.leadItem.statusClass}`">{{ archiveModel.leadItem.status }}</span>
            <span class="archive-badge subtle" :class="`source-${archiveModel.leadItem.sourceClass}`">{{ archiveModel.leadItem.sourceLabel }}</span>
          </div>
        </div>
        <p>{{ archiveModel.leadItem.explanationSummary }}</p>
        <div class="archive-detail-grid">
          <article class="archive-detail-item">
            <span>累计表现</span>
            <strong>{{ archiveModel.leadItem.cumulativeReturn }}</strong>
          </article>
          <article class="archive-detail-item">
            <span>超额表现</span>
            <strong>{{ archiveModel.leadItem.excessReturn }}</strong>
          </article>
          <article class="archive-detail-item">
            <span>推荐理由</span>
            <strong>{{ archiveModel.leadItem.reason }}</strong>
          </article>
          <article class="archive-detail-item">
            <span>结果标签</span>
            <strong>{{ archiveModel.leadItem.meta }}</strong>
          </article>
        </div>
        <div v-if="archiveModel.leadItem.proofTags.length" class="archive-tag-row">
          <span v-for="tag in archiveModel.leadItem.proofTags" :key="tag" class="archive-chip">{{ tag }}</span>
        </div>
        <div class="archive-item-actions">
          <button type="button" class="h5-btn" @click="handlePrimaryAction">{{ primaryActionLabel }}</button>
          <button type="button" class="h5-btn-ghost" @click="goStrategies">去策略页</button>
        </div>
      </div>
      <H5EmptyState
        v-else
        :title="archiveModel.emptyState.title"
        :description="archiveModel.emptyState.description"
      />
      <div v-if="!archiveModel.leadItem" class="archive-empty-actions">
        <button type="button" class="h5-btn" @click="handlePrimaryAction">{{ primaryActionLabel }}</button>
      </div>
      <p v-if="loadError" class="archive-inline-note">{{ loadError }}</p>
    </H5SectionBlock>

    <H5SectionBlock eyebrow="时间线" :title="timelineTitle" tone="accent">
      <p v-if="archiveHiddenCount > 0" class="archive-limit-note">
        当前还有 {{ archiveHiddenCount }} 条历史样本需要更高权限后继续查看。
      </p>

      <div v-if="archiveModel.timeline.length" class="archive-timeline">
        <article v-for="item in archiveModel.timeline" :key="item.id" class="archive-timeline-item">
          <div class="archive-item-topline">
            <div>
              <span>{{ item.date }}</span>
              <strong>{{ item.headline }}</strong>
            </div>
            <div class="archive-badge-row">
              <span class="archive-badge" :class="`status-${item.statusClass}`">{{ item.status }}</span>
              <span class="archive-badge subtle" :class="`source-${item.sourceClass}`">{{ item.sourceLabel }}</span>
            </div>
          </div>
          <p>{{ item.reason }}</p>
          <div class="archive-item-foot">
            <span>累计 {{ item.cumulativeReturn }}</span>
            <span>超额 {{ item.excessReturn }}</span>
          </div>
          <div v-if="item.proofTags.length" class="archive-tag-row">
            <span v-for="tag in item.proofTags" :key="tag" class="archive-chip subtle">{{ tag }}</span>
          </div>
        </article>
      </div>

      <H5EmptyState
        v-else
        title="当前没有更多历史样本"
        description="刷新后会把历史推荐结果、理由和来源说明收进这个时间线。"
      />
    </H5SectionBlock>

    <H5StickyCta
      :title="stickyTitle"
      :description="stickyDescription"
      :primary-label="primaryActionLabel"
      secondary-label="回资讯页"
      @primary="handlePrimaryAction"
      @secondary="goNews"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import H5ActionBar from "../components/H5ActionBar.vue";
import H5EmptyState from "../components/H5EmptyState.vue";
import H5HeroCard from "../components/H5HeroCard.vue";
import H5SectionBlock from "../components/H5SectionBlock.vue";
import H5StickyCta from "../components/H5StickyCta.vue";
import H5SummaryCard from "../components/H5SummaryCard.vue";
import { getMembershipQuota } from "../../../api/membership";
import {
  getStockRecommendationInsight,
  getStockRecommendationVersionHistory,
  listStockRecommendations
} from "../../../api/market";
import { shouldUseDemoFallback } from "../../../lib/fallback-policy";
import {
  buildFallbackStrategyVersionHistory,
  buildStrategyInsightSections,
  buildStrategyProofTags,
  mapStrategyVersionHistory
} from "../../../lib/strategy-version";
import { useClientAuth } from "../../../shared/auth/client-auth";
import { buildArchiveFeedModel } from "../lib/archive-feed.js";
import {
  fallbackStockInsights,
  fallbackStockRecommendations
} from "../lib/mock-data";
import { formatDate, formatDateTime, resolveVipStage, toArray } from "../lib/formatters";
import { resolveHighlightTone } from "../lib/surface-tone.js";

const fallbackArchiveInsights = {
  sr_local_001: {
    explanation: fallbackStockInsights.sr_local_001,
    performance_stats: {
      cumulative_return: 0.0232,
      excess_return: 0.0111,
      benchmark_source: "actual: CSI300"
    }
  },
  sr_local_002: {
    explanation: fallbackStockInsights.sr_local_002,
    performance_stats: {
      cumulative_return: -0.0182,
      excess_return: -0.0142,
      benchmark_source: "estimated: 55% of strategy daily return"
    }
  },
  sr_local_003: {
    explanation: fallbackStockInsights.sr_local_003,
    performance_stats: {
      cumulative_return: 0.0181,
      excess_return: 0.0091,
      benchmark_source: "simulated: fallback benchmark series"
    }
  }
};

const router = useRouter();
const { isLoggedIn } = useClientAuth();
const useDemoFallback = shouldUseDemoFallback();

const loading = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const rawItems = ref(useDemoFallback ? [...fallbackStockRecommendations] : []);
const insightMap = ref(useDemoFallback ? { ...fallbackArchiveInsights } : {});
const versionHistoryMap = ref({});
const rawQuota = ref({});

const hasVipAccess = computed(() => resolveVipStage(rawQuota.value));
const archiveAccessStage = computed(() => {
  if (!isLoggedIn.value) {
    return "VISITOR";
  }
  return hasVipAccess.value ? "VIP" : "REGISTERED";
});
const archiveStageLabel = computed(() => {
  if (archiveAccessStage.value === "VIP") {
    return "会员全量";
  }
  if (archiveAccessStage.value === "REGISTERED") {
    return "登录已同步";
  }
  return "游客试读";
});

const archiveRows = computed(() =>
  [...rawItems.value]
    .sort((left, right) => Date.parse(right?.valid_from || "") - Date.parse(left?.valid_from || ""))
    .map((item, index) => {
      const insight = insightMap.value[item.id] || {};
      const explanation = insight.explanation || fallbackArchiveInsights[item.id]?.explanation || buildFallbackArchiveExplanation(item, index);
      const sections = buildStrategyInsightSections(explanation, item.reason_summary || "暂无推荐理由");
      const stats = insight.performance_stats || fallbackArchiveInsights[item.id]?.performance_stats || {};
      const sourceMeta = resolveSourceMeta(stats.benchmark_source);
      const versionRows = resolveVersionRows(item, explanation);
      const versionTag = versionRows[0]?.version ? `版本 ${versionRows[0].version}` : "";
      return {
        id: item.id,
        name: `${item.symbol || "-"} ${item.name || ""}`.trim(),
        status: mapArchiveStatusLabel(item.status),
        statusClass: resolveArchiveStatusClass(item.status),
        sourceLabel: sourceMeta.label,
        sourceClass: sourceMeta.className,
        reason: item.reason_summary || "暂无推荐理由",
        explanationSummary: sections.whyNow || item.reason_summary || "暂无推荐理由",
        cumulativeReturn: formatSignedPercent(stats.cumulative_return),
        excessReturn: formatSignedPercent(stats.excess_return),
        date: formatDate(item.valid_from || item.created_at),
        proofTags: [...buildStrategyProofTags(explanation, { limit: 2, includeSeedCount: true }), versionTag]
          .filter(Boolean)
          .slice(0, 3)
      };
    })
);

const visibleArchiveRows = computed(() => {
  if (archiveAccessStage.value === "VIP") {
    return archiveRows.value;
  }
  if (archiveAccessStage.value === "REGISTERED") {
    return archiveRows.value.slice(0, 4);
  }
  return archiveRows.value.slice(0, 2);
});

const archiveHiddenCount = computed(() => Math.max(0, archiveRows.value.length - visibleArchiveRows.value.length));
const archiveModel = computed(() => buildArchiveFeedModel({
  items: visibleArchiveRows.value
}));

const heroTitle = computed(() => archiveModel.value.leadItem ? `先复盘 ${archiveModel.value.leadItem.headline}` : "先看历史样本，再决定是否继续深读");
const heroDescription = computed(() => {
  if (!archiveModel.value.timeline.length) {
    return archiveAccessStage.value === "VISITOR"
      ? "游客先看公开历史样本，登录后会继续展开更多时间线。"
      : "刷新后会补齐历史推荐结果、理由和来源说明。";
  }
  if (archiveHiddenCount.value > 0) {
    return archiveAccessStage.value === "VISITOR"
      ? "当前先展示公开历史样本，登录后会继续展开更多结果与复盘线索。"
      : "当前先展示你可见的历史样本，升级后会继续展开完整复盘时间线。";
  }
  return archiveAccessStage.value === "VIP"
    ? "会员阶段可完整回看结果、理由和来源说明。"
    : "当前阶段已展示全部可见历史样本，可以继续顺着时间线回看。";
});
const heroMeta = computed(() => [
  archiveStageLabel.value,
  archiveModel.value.summaryCards[0]?.value || "0 条",
  archiveHiddenCount.value > 0 ? `待解锁 ${archiveHiddenCount.value} 条` : lastUpdatedAt.value || "已同步"
]);
const timelineTitle = computed(() => archiveHiddenCount.value > 0 ? "当前阶段可读的历史样本" : "按时间顺序继续回看完整样本");

const stickyTitle = computed(() => {
  if (!archiveModel.value.timeline.length) {
    return archiveAccessStage.value === "VISITOR" ? "先登录，再回到这里继续查看历史档案" : "先去策略页生成新的推荐样本";
  }
  if (archiveAccessStage.value === "VIP") {
    return "完整档案已解锁，下一步可以回到我的关注继续跟踪";
  }
  if (archiveAccessStage.value === "REGISTERED") {
    return "已登录后可看更多样本，升级会员可继续看完整复盘";
  }
  return "游客先看部分历史样本，登录后继续解锁更多档案";
});
const stickyDescription = computed(() => {
  if (!archiveModel.value.timeline.length) {
    return archiveAccessStage.value === "VISITOR"
      ? "登录后会按 redirect 回到当前 H5 档案页，继续顺着时间线阅读。"
      : "新的推荐结果会继续回到档案页，按时间线沉淀下来。";
  }
  if (archiveAccessStage.value === "VIP") {
    return "看完结果后，可以把仍值得追踪的对象接到关注页继续观察。";
  }
  if (archiveAccessStage.value === "REGISTERED") {
    return "当前还有一部分历史样本和更完整复盘线索需要会员权限后继续展开。";
  }
  return "登录后会按 redirect 回到这个 H5 档案页，继续顺着时间线阅读。";
});
const primaryActionLabel = computed(() => {
  if (!archiveModel.value.timeline.length) {
    return archiveAccessStage.value === "VISITOR" ? "立即登录" : "去策略页";
  }
  if (archiveAccessStage.value === "VIP") {
    return "去我的关注";
  }
  if (archiveAccessStage.value === "REGISTERED") {
    return "升级会员看完整复盘";
  }
  return "登录查看更多档案";
});

async function loadArchivePage() {
  loading.value = true;
  loadError.value = "";

  if (!isLoggedIn.value) {
    applyFallbackArchiveState();
    rawQuota.value = {};
    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
    loading.value = false;
    return;
  }

  const errors = [];
  const [recommendationResult, quotaResult] = await Promise.allSettled([
    listStockRecommendations({ page: 1, page_size: 20 }),
    getMembershipQuota()
  ]);

  if (recommendationResult.status === "fulfilled") {
    const items = toArray(recommendationResult.value?.items, []);
    rawItems.value = items;
    const hydrationErrors = await hydrateArchiveRows(items);
    errors.push(...hydrationErrors);
  } else {
    errors.push(`历史档案加载失败：${recommendationResult.reason?.message || "unknown error"}`);
    if (useDemoFallback) {
      applyFallbackArchiveState();
    } else {
      rawItems.value = [];
      insightMap.value = {};
      versionHistoryMap.value = {};
    }
  }

  if (quotaResult.status === "fulfilled") {
    rawQuota.value = quotaResult.value || {};
  } else {
    rawQuota.value = {};
    errors.push(`会员状态加载失败：${quotaResult.reason?.message || "unknown error"}`);
  }

  loadError.value = errors.filter(Boolean).join("；");
  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  loading.value = false;
}

async function hydrateArchiveRows(items) {
  if (!Array.isArray(items) || items.length === 0) {
    insightMap.value = {};
    versionHistoryMap.value = {};
    return [];
  }

  const nextInsightMap = {};
  const nextVersionHistoryMap = {};
  const errors = [];

  await Promise.all(items.map(async (item) => {
    if (!item?.id) {
      return;
    }
    const [insightResult, versionHistoryResult] = await Promise.allSettled([
      getStockRecommendationInsight(item.id),
      getStockRecommendationVersionHistory(item.id)
    ]);

    if (insightResult.status === "fulfilled" && insightResult.value) {
      nextInsightMap[item.id] = insightResult.value;
    } else if (insightResult.status === "rejected") {
      errors.push(`${item.symbol || item.id} 档案补全失败：${insightResult.reason?.message || "unknown error"}`);
    }

    if (versionHistoryResult.status === "fulfilled") {
      nextVersionHistoryMap[item.id] = toArray(versionHistoryResult.value?.items, []);
    } else if (versionHistoryResult.status === "rejected") {
      errors.push(`${item.symbol || item.id} 版本轨迹补全失败：${versionHistoryResult.reason?.message || "unknown error"}`);
    }
  }));

  insightMap.value = nextInsightMap;
  versionHistoryMap.value = nextVersionHistoryMap;
  return errors;
}

function applyFallbackArchiveState() {
  rawItems.value = useDemoFallback ? [...fallbackStockRecommendations] : [];
  insightMap.value = useDemoFallback ? { ...fallbackArchiveInsights } : {};
  versionHistoryMap.value = {};
}

function resolveVersionRows(item, explanation) {
  const mappedRows = mapStrategyVersionHistory(versionHistoryMap.value[item.id] || [], formatDateTime, { limit: 2 });
  if (mappedRows.length) {
    return mappedRows;
  }
  return buildFallbackStrategyVersionHistory(explanation, {
    formatDateTime,
    limit: 1,
    reasonSummary: item.reason_summary || "",
    tradeDate: String(item.valid_from || "").slice(0, 10),
    strategyVersion: explanation.strategy_version || item.strategy_version || ""
  });
}

function handlePrimaryAction() {
  if (!archiveModel.value.timeline.length) {
    if (archiveAccessStage.value === "VISITOR") {
      goAuth();
      return;
    }
    goStrategies();
    return;
  }
  if (archiveAccessStage.value === "VIP") {
    goWatchlist();
    return;
  }
  if (archiveAccessStage.value === "REGISTERED") {
    goMembership();
    return;
  }
  goAuth();
}

function goAuth() {
  router.push({ path: "/auth", query: { redirect: "/archive" } });
}

function goMembership() {
  router.push("/membership");
}

function goNews() {
  router.push("/news");
}

function goStrategies() {
  router.push("/strategies");
}

function goWatchlist() {
  router.push("/watchlist");
}

function mapArchiveStatusLabel(status) {
  const value = String(status || "").toUpperCase();
  if (value === "TRACKING") return "持续跟踪";
  if (value === "PUBLISHED" || value === "ACTIVE") return "已发布";
  if (value === "HIT_TAKE_PROFIT") return "已止盈";
  if (value === "HIT_STOP_LOSS") return "已止损";
  if (value === "INVALIDATED") return "已失效";
  if (value === "REVIEWED") return "已复盘";
  return "待确认";
}

function resolveArchiveStatusClass(status) {
  const label = mapArchiveStatusLabel(status);
  if (label === "已止盈") return "success";
  if (label === "已止损" || label === "已失效") return "warning";
  if (label === "已复盘") return "neutral";
  return "brand";
}

function resolveSourceMeta(sourceText) {
  const source = String(sourceText || "").toLowerCase();
  if (source.includes("actual")) {
    return { label: "已验证", className: "verified" };
  }
  if (source.includes("estimated")) {
    return { label: "预估", className: "estimated" };
  }
  if (source.includes("simulated") || source.includes("fallback")) {
    return { label: "模拟", className: "simulated" };
  }
  return { label: "待说明", className: "pending" };
}

function buildFallbackArchiveExplanation(item, index = 0) {
  return {
    confidence_reason: item?.reason_summary || "系统会把历史推荐结果和理由收回到一个可回看的时间线里。",
    consensus_summary: "历史样本更适合先看结果，再回头验证最初为什么成立或失效。",
    risk_boundary: "复盘时要同时看结果、原因和边界，避免只看收益。",
    strategy_version: item?.strategy_version || `archive-demo-${index + 1}`,
    generated_at: item?.valid_from || "",
    workload_summary: {
      seed_count: 18 + index * 2,
      agent_count: 5,
      scenario_count: 4
    },
    seed_highlights: [
      item?.reason_summary || "推荐逻辑待同步",
      mapArchiveStatusLabel(item?.status),
      "来源说明待补充"
    ].filter(Boolean)
  };
}

function formatSignedPercent(value) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return "-";
  }
  return `${numeric > 0 ? "+" : ""}${(numeric * 100).toFixed(1)}%`;
}

onMounted(() => {
  loadArchivePage();
});
</script>

<style scoped>
.archive-page {
  gap: 14px;
}

.archive-hero-metrics,
.archive-read-order,
.archive-detail-grid,
.archive-timeline {
  display: grid;
  gap: 10px;
}

.archive-detail-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.archive-order-item,
.archive-lead-card,
.archive-timeline-item {
  display: grid;
  gap: 10px;
  padding: 15px 14px;
  border-radius: var(--h5-radius-md);
  background: rgba(20, 52, 95, 0.045);
}

.archive-order-item span,
.archive-item-topline span,
.archive-detail-item span {
  color: var(--h5-text-soft);
  font-size: 12px;
}

.archive-order-item p,
.archive-lead-card p,
.archive-timeline-item p {
  margin: 0;
  color: var(--h5-text-sub);
  font-size: 13px;
  line-height: 1.7;
}

.archive-order-item strong,
.archive-item-topline strong,
.archive-detail-item strong {
  color: var(--h5-text);
}

.archive-item-topline,
.archive-item-foot,
.archive-item-actions {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.archive-badge-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.archive-badge,
.archive-chip {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
}

.archive-badge.status-brand,
.archive-chip {
  color: var(--h5-brand);
  background: rgba(24, 58, 110, 0.08);
}

.archive-badge.status-success {
  color: var(--h5-success);
  background: rgba(31, 150, 113, 0.12);
}

.archive-badge.status-warning {
  color: var(--h5-warning);
  background: rgba(184, 137, 61, 0.14);
}

.archive-badge.status-neutral {
  color: var(--h5-text-sub);
  background: rgba(20, 52, 95, 0.08);
}

.archive-badge.subtle {
  background: rgba(255, 255, 255, 0.74);
}

.archive-badge.source-verified {
  color: var(--h5-success);
}

.archive-badge.source-estimated {
  color: var(--h5-warning);
}

.archive-badge.source-simulated {
  color: var(--h5-brand);
}

.archive-badge.source-pending {
  color: var(--h5-text-sub);
}

.archive-tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.archive-chip.subtle {
  color: var(--h5-text-sub);
  background: rgba(20, 52, 95, 0.05);
}

.archive-detail-item {
  display: grid;
  gap: 4px;
  padding: 12px;
  border-radius: var(--h5-radius-sm);
  background: rgba(255, 255, 255, 0.72);
}

.archive-item-foot {
  align-items: center;
  color: var(--h5-text-soft);
  font-size: 12px;
}

.archive-empty-actions {
  margin-top: 12px;
}

.archive-limit-note,
.archive-inline-note {
  margin: 0;
  color: var(--h5-text-soft);
  font-size: 12px;
  line-height: 1.7;
}

.archive-inline-note {
  color: var(--h5-warning);
}

@media (max-width: 380px) {
  .archive-detail-grid {
    grid-template-columns: 1fr;
  }

  .archive-item-topline,
  .archive-item-foot,
  .archive-item-actions {
    flex-wrap: wrap;
  }

  .archive-badge-row {
    justify-content: flex-start;
  }
}
</style>
