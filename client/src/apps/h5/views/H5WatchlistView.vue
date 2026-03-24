<template>
  <div class="h5-page fade-up watchlist-page">
    <div class="h5-page-topline">
      <span class="h5-page-tagline">我的关注</span>
      <span>{{ lastUpdatedText }}</span>
    </div>

    <H5HeroCard
      eyebrow="变化工作台"
      :title="heroTitle"
      :description="heroDescription"
      :meta="heroMeta"
      tone="accent"
    >
      <template #actions>
        <H5ActionBar>
          <button type="button" class="h5-btn-secondary" :disabled="loading" @click="loadWatchlistPage">
            {{ loading ? "同步中..." : "刷新关注" }}
          </button>
          <button type="button" class="h5-btn-ghost" @click="goStrategies">去策略页</button>
        </H5ActionBar>
      </template>

      <div class="watchlist-hero-metrics h5-grid-3">
        <H5SummaryCard
          v-for="(item, index) in watchlistModel.summaryCards"
          :key="item.label"
          :label="item.label"
          :value="item.value"
          :note="item.note"
          :tone="resolveHighlightTone(index)"
        />
      </div>
    </H5HeroCard>

    <H5SectionBlock eyebrow="优先跟踪" title="先处理今天最需要动作的一项" tone="soft">
      <div v-if="watchlistModel.leadItem" class="watchlist-lead-card">
        <div class="watchlist-item-topline">
          <div>
            <span>{{ leadMetaText }}</span>
            <strong>{{ watchlistModel.leadItem.headline }}</strong>
          </div>
          <em>{{ watchlistModel.leadItem.risk }}</em>
        </div>
        <p>{{ watchlistModel.leadItem.summary }}</p>
        <div v-if="watchlistModel.leadItem.proofTags.length" class="watchlist-tag-row">
          <span v-for="tag in watchlistModel.leadItem.proofTags" :key="tag" class="watchlist-chip">{{ tag }}</span>
        </div>
        <div class="watchlist-detail-grid">
          <article class="watchlist-detail-item">
            <span>当前状态</span>
            <strong>{{ watchlistModel.leadItem.status }}</strong>
          </article>
          <article class="watchlist-detail-item">
            <span>下一动作</span>
            <strong>{{ watchlistModel.leadItem.nextAction }}</strong>
          </article>
          <article class="watchlist-detail-item">
            <span>累计表现</span>
            <strong :class="`trend-${watchlistModel.leadItem.performanceClass}`">{{ watchlistModel.leadItem.performance }}</strong>
          </article>
          <article class="watchlist-detail-item">
            <span>最新动态</span>
            <strong>{{ watchlistModel.leadItem.latestNewsTitle }}</strong>
          </article>
        </div>
        <div class="watchlist-item-actions">
          <button
            v-if="!isLoggedIn"
            type="button"
            class="h5-btn"
            @click="goAuth"
          >
            登录同步更多变化
          </button>
          <button
            v-else
            type="button"
            class="h5-btn-secondary"
            @click="goNews"
          >
            去资讯页确认
          </button>
          <button type="button" class="h5-btn-ghost" @click="removeItem(watchlistModel.leadItem.id)">移出关注</button>
        </div>
      </div>
      <H5EmptyState
        v-else
        :title="watchlistModel.emptyState.title"
        :description="watchlistModel.emptyState.description"
      />
      <div v-if="!watchlistModel.leadItem" class="watchlist-empty-actions">
        <button v-if="!isLoggedIn" type="button" class="h5-btn" @click="goAuth">登录后同步更多变化</button>
        <button v-else type="button" class="h5-btn-secondary" @click="goStrategies">去策略页加入关注</button>
      </div>
      <p v-if="loadError" class="watchlist-inline-note">{{ loadError }}</p>
    </H5SectionBlock>

    <H5SectionBlock eyebrow="继续回访" title="顺着单列清单往下看，先确认变化，再决定是否继续读">
      <div v-if="watchlistModel.rows.length" class="watchlist-list">
        <article v-for="item in watchlistModel.rows" :key="item.id" class="watchlist-list-item">
          <div class="watchlist-item-topline">
            <div>
              <span>{{ item.latestNewsTime }}</span>
              <strong>{{ item.headline }}</strong>
            </div>
            <em>{{ item.risk }}</em>
          </div>
          <p>{{ item.summary }}</p>
          <div class="watchlist-item-foot">
            <span class="watchlist-badge">{{ item.status }}</span>
            <span :class="`trend-${item.performanceClass}`">{{ item.performance }}</span>
          </div>
          <div v-if="item.proofTags.length" class="watchlist-tag-row">
            <span v-for="tag in item.proofTags" :key="tag" class="watchlist-chip subtle">{{ tag }}</span>
          </div>
          <div class="watchlist-item-actions compact">
            <button v-if="!isLoggedIn" type="button" class="h5-btn-secondary" @click="goAuth">登录</button>
            <button v-else type="button" class="h5-btn-secondary" @click="goNews">看资讯</button>
            <button type="button" class="h5-btn-ghost" @click="removeItem(item.id)">移除</button>
          </div>
        </article>
      </div>
      <H5EmptyState
        v-else
        title="当前没有更多关注项"
        description="保存新的关注对象后，这里会继续显示回访清单。"
      />
    </H5SectionBlock>

    <H5SectionBlock eyebrow="回访节奏" title="把每天的关注动作收成固定顺序" tone="accent">
      <div class="watchlist-rhythm-list">
        <article v-for="item in revisitPlan" :key="item.title" class="watchlist-rhythm-item">
          <span>{{ item.time }}</span>
          <strong>{{ item.title }}</strong>
          <p>{{ item.desc }}</p>
        </article>
      </div>
    </H5SectionBlock>

    <H5StickyCta
      :title="stickyTitle"
      :description="stickyDescription"
      :primary-label="stickyPrimaryLabel"
      secondary-label="去策略页"
      @primary="handlePrimaryAction"
      @secondary="goStrategies"
    />
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import H5ActionBar from "../components/H5ActionBar.vue";
import H5EmptyState from "../components/H5EmptyState.vue";
import H5HeroCard from "../components/H5HeroCard.vue";
import H5SectionBlock from "../components/H5SectionBlock.vue";
import H5StickyCta from "../components/H5StickyCta.vue";
import H5SummaryCard from "../components/H5SummaryCard.vue";
import { getMembershipQuota } from "../../../api/membership";
import { getStockRecommendationInsight } from "../../../api/market";
import { buildStrategyInsightSections, buildStrategyProofTags } from "../../../lib/strategy-version";
import { WATCHLIST_EVENT, listWatchedStocks, removeWatchedStock } from "../../../lib/watchlist";
import { useClientAuth } from "../../../shared/auth/client-auth";
import { formatDateTime, mapRiskLevel, resolveVipStage } from "../lib/formatters";
import { buildWatchlistFeedModel } from "../lib/watchlist-feed.js";
import { resolveHighlightTone } from "../lib/surface-tone.js";

const router = useRouter();
const { isLoggedIn } = useClientAuth();

const loading = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const rawWatchedItems = ref([]);
const insightMap = ref({});
const rawQuota = ref({});
const watchVersion = ref(0);

const hydratedWatchRows = computed(() => {
  void watchVersion.value;
  return rawWatchedItems.value.map((item) => {
    const insight = insightMap.value[item.id] || {};
    const recommendation = insight.recommendation || {};
    const explanation = insight.explanation || {};
    const stats = insight.performance_stats || {};
    const relatedNews = Array.isArray(insight.related_news) ? insight.related_news : [];
    const latestNews = relatedNews[0] || null;
    const performance = Number(stats.cumulative_return);
    const sections = buildStrategyInsightSections(
      explanation,
      recommendation.reason_summary || item.reason_summary || ""
    );

    return {
      id: item.id,
      name: `${item.symbol || "-"} ${item.name || ""}`.trim(),
      summary: sections.whyNow || recommendation.reason_summary || item.reason_summary || "继续跟踪变化与风险边界。",
      status: mapWatchStatusLabel(recommendation.status || item.status),
      risk: mapRiskLevel(recommendation.risk_level || item.risk_level),
      nextAction: resolveNextAction(recommendation.status || item.status),
      performance: formatPerformance(performance),
      performanceClass: resolveTrendClass(performance),
      proofTags: buildStrategyProofTags(explanation, { limit: 3 }),
      latestNewsTitle: latestNews?.title || "暂无新增资讯",
      latestNewsTime: formatDateTime(latestNews?.published_at || latestNews?.created_at || item.added_at),
      addedAt: formatDateTime(item.added_at)
    };
  });
});

const watchlistModel = computed(() => buildWatchlistFeedModel({
  items: hydratedWatchRows.value
}));

const lastUpdatedText = computed(() => lastUpdatedAt.value || "关注清单");
const heroTitle = computed(() => watchlistModel.value.leadItem ? `先看 ${watchlistModel.value.leadItem.headline}` : "先建立一份自己的关注清单");
const heroDescription = computed(() => {
  if (!rawWatchedItems.value.length) {
    return "关注页会把你保存过的标的集中起来，方便每天按固定顺序回访。";
  }
  if (!isLoggedIn.value) {
    return "你已经保存了本地关注对象，登录后会补齐更多变化、资讯和风险边界信息。";
  }
  if (resolveVipStage(rawQuota.value)) {
    return "当前会员状态已同步，可以继续把关注、资讯和历史复盘串成一条连续阅读链。";
  }
  return "当前先展示关注主线，会员状态同步后会继续补齐更深的阅读与复盘能力。";
});
const heroMeta = computed(() => [
  `${rawWatchedItems.value.length} 项关注`,
  isLoggedIn.value ? "已登录" : "本地保存",
  resolveVipStage(rawQuota.value) ? "会员已同步" : "公开跟踪"
]);
const leadMetaText = computed(() => watchlistModel.value.leadItem?.latestNewsTime || watchlistModel.value.leadItem?.addedAt || "最近刷新");

const revisitPlan = [
  { time: "09:30", title: "先看状态变化", desc: "确认今天最该继续跟踪、止盈还是暂停观察。" },
  { time: "11:30", title: "再看资讯变化", desc: "用新增资讯确认原始逻辑是否还成立。" },
  { time: "15:00", title: "最后看边界", desc: "收盘前再核对风险线和下一步动作。" }
];

const stickyTitle = computed(() => {
  if (!rawWatchedItems.value.length) {
    return "先从首页或策略页加入关注";
  }
  if (!isLoggedIn.value) {
    return "登录后同步更多变化和资讯";
  }
  if (resolveVipStage(rawQuota.value)) {
    return "顺着关注页继续进入资讯或历史复盘";
  }
  return "继续查看会员权益，补齐更多跟踪能力";
});
const stickyDescription = computed(() => {
  if (!rawWatchedItems.value.length) {
    return "保存后的对象会回到这个页面，按单列节奏继续跟踪。";
  }
  if (!isLoggedIn.value) {
    return "本地保存的数据会保留，登录后自动接到更多服务能力。";
  }
  if (resolveVipStage(rawQuota.value)) {
    return "现在可以把关注、资讯和历史样本自然串起来。";
  }
  return "会员页会优先解释状态、待办和升级价值，不会直接把价格压到最前面。";
});
const stickyPrimaryLabel = computed(() => {
  if (!rawWatchedItems.value.length) {
    return "去策略页";
  }
  if (!isLoggedIn.value) {
    return "立即登录";
  }
  return resolveVipStage(rawQuota.value) ? "去资讯页" : "看会员权益";
});

async function loadWatchlistPage() {
  loading.value = true;
  loadError.value = "";
  rawWatchedItems.value = listWatchedStocks();

  if (!isLoggedIn.value) {
    rawQuota.value = {};
    insightMap.value = {};
    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
    loading.value = false;
    return;
  }

  const [quotaResult, ...insightResults] = await Promise.allSettled([
    getMembershipQuota(),
    ...rawWatchedItems.value.map((item) => getStockRecommendationInsight(item.id))
  ]);

  if (quotaResult.status === "fulfilled") {
    rawQuota.value = quotaResult.value || {};
  } else {
    loadError.value = quotaResult.reason?.message || "会员状态同步失败";
  }

  const nextInsightMap = {};
  rawWatchedItems.value.forEach((item, index) => {
    const result = insightResults[index];
    if (result?.status === "fulfilled" && result.value) {
      nextInsightMap[item.id] = result.value;
      return;
    }
    if (result?.status === "rejected") {
      loadError.value = [loadError.value, `${item.symbol || item.id} 同步失败：${result.reason?.message || "unknown error"}`]
        .filter(Boolean)
        .join("；");
    }
  });

  insightMap.value = nextInsightMap;
  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  loading.value = false;
}

function refreshWatchlistState() {
  rawWatchedItems.value = listWatchedStocks();
  watchVersion.value += 1;
  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
}

function removeItem(id) {
  removeWatchedStock(id);
}

function handleWatchlistUpdate() {
  refreshWatchlistState();
  if (isLoggedIn.value) {
    loadWatchlistPage();
  }
}

function goStrategies() {
  router.push("/strategies");
}

function goNews() {
  router.push("/news");
}

function goAuth() {
  router.push({ path: "/auth", query: { redirect: "/watchlist" } });
}

function handlePrimaryAction() {
  if (!rawWatchedItems.value.length) {
    goStrategies();
    return;
  }
  if (!isLoggedIn.value) {
    goAuth();
    return;
  }
  if (resolveVipStage(rawQuota.value)) {
    goNews();
    return;
  }
  router.push("/membership");
}

onMounted(() => {
  loadWatchlistPage();
  if (typeof window !== "undefined") {
    window.addEventListener(WATCHLIST_EVENT, handleWatchlistUpdate);
  }
});

onBeforeUnmount(() => {
  if (typeof window !== "undefined") {
    window.removeEventListener(WATCHLIST_EVENT, handleWatchlistUpdate);
  }
});

function mapWatchStatusLabel(status) {
  const value = String(status || "").toUpperCase();
  if (value === "TRACKING") return "持续跟踪";
  if (value === "PUBLISHED" || value === "ACTIVE") return "已发布";
  if (value === "HIT_TAKE_PROFIT") return "已止盈";
  if (value === "HIT_STOP_LOSS") return "已止损";
  if (value === "INVALIDATED") return "已失效";
  if (value === "REVIEWED") return "已复盘";
  return "等待确认";
}

function resolveNextAction(status) {
  const label = mapWatchStatusLabel(status);
  if (label === "已止盈") return "整理复盘";
  if (label === "已止损") return "暂停跟踪";
  if (label === "已失效") return "等待新逻辑";
  return "继续观察";
}

function resolveTrendClass(value) {
  if (!Number.isFinite(value)) {
    return "flat";
  }
  if (value > 0) return "up";
  if (value < 0) return "down";
  return "flat";
}

function formatPerformance(value) {
  if (!Number.isFinite(value)) {
    return "-";
  }
  const percent = (value * 100).toFixed(1);
  return `${value > 0 ? "+" : ""}${percent}%`;
}
</script>

<style scoped>
.watchlist-page {
  gap: 14px;
}

.watchlist-hero-metrics,
.watchlist-detail-grid,
.watchlist-rhythm-list {
  display: grid;
  gap: 10px;
}

.watchlist-detail-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.watchlist-lead-card,
.watchlist-list-item,
.watchlist-rhythm-item {
  display: grid;
  gap: 10px;
  padding: 15px 14px;
  border-radius: var(--h5-radius-md);
  background: rgba(20, 52, 95, 0.045);
}

.watchlist-list {
  display: grid;
  gap: 10px;
}

.watchlist-empty-actions {
  margin-top: 12px;
}

.watchlist-item-topline,
.watchlist-item-foot,
.watchlist-item-actions {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.watchlist-item-actions {
  align-items: center;
}

.watchlist-item-actions.compact {
  justify-content: flex-start;
}

.watchlist-item-topline span,
.watchlist-rhythm-item span {
  color: var(--h5-text-soft);
  font-size: 12px;
}

.watchlist-item-topline strong,
.watchlist-rhythm-item strong,
.watchlist-detail-item strong {
  display: block;
  color: var(--h5-text);
}

.watchlist-item-topline strong {
  margin-top: 4px;
  font-size: 17px;
  line-height: 1.35;
}

.watchlist-item-topline em {
  font-style: normal;
  color: var(--h5-warning);
  font-size: 12px;
  font-weight: 700;
}

.watchlist-detail-item {
  display: grid;
  gap: 4px;
  padding: 12px;
  border-radius: var(--h5-radius-sm);
  background: rgba(255, 255, 255, 0.72);
}

.watchlist-detail-item span {
  color: var(--h5-text-soft);
  font-size: 12px;
}

.watchlist-tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.watchlist-chip,
.watchlist-badge {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
}

.watchlist-chip {
  color: var(--h5-brand);
  background: rgba(24, 58, 110, 0.08);
}

.watchlist-chip.subtle {
  color: var(--h5-text-sub);
  background: rgba(20, 52, 95, 0.05);
}

.watchlist-badge {
  color: var(--h5-brand);
  background: rgba(24, 58, 110, 0.08);
}

.watchlist-inline-note {
  margin: 0;
  color: var(--h5-warning);
  font-size: 12px;
  line-height: 1.7;
}

.trend-up {
  color: var(--h5-success);
}

.trend-down {
  color: var(--h5-danger);
}

.trend-flat {
  color: var(--h5-text-sub);
}

@media (max-width: 380px) {
  .watchlist-detail-grid {
    grid-template-columns: 1fr;
  }

  .watchlist-item-actions {
    flex-wrap: wrap;
  }
}
</style>
