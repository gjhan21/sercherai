<template>
  <div class="h5-page fade-up home-page">
    <div class="h5-page-topline">
      <span class="h5-page-tagline">今日主线</span>
      <span>{{ todayText }}</span>
    </div>

    <H5HeroCard
      eyebrow="今日核心观点"
      :title="heroTitle"
      :description="heroDescription"
      :meta="heroMeta"
      tone="hero"
    >
      <template #actions>
        <H5ActionBar class="home-lead-actions">
          <button type="button" class="h5-btn" @click="goStrategies">查看完整观点</button>
          <button type="button" class="h5-btn-secondary" :disabled="!leadStock" @click="toggleWatch">
            {{ watchButtonText }}
          </button>
          <button type="button" class="h5-btn-ghost" :disabled="loading" @click="refreshPage">
            {{ loading ? "同步中..." : "刷新" }}
          </button>
        </H5ActionBar>
      </template>

      <div v-if="marketPulse.length" class="home-hero-metrics h5-grid-3">
        <H5SummaryCard
          v-for="(item, index) in marketPulse"
          :key="item.label"
          :label="item.label"
          :value="item.value"
          :tone="resolveHighlightTone(index)"
        />
      </div>
    </H5HeroCard>

    <H5SectionBlock eyebrow="今日脉冲" title="先看状态，再决定继续往哪刷" tone="soft">
      <div class="home-status-grid">
        <H5SummaryCard
          v-for="(item, index) in statusCards"
          :key="item.label"
          :label="item.label"
          :value="item.value"
          :note="item.note"
          :tone="resolveHighlightTone(index)"
        />
      </div>
      <p v-if="loadError" class="home-inline-note">{{ loadError }}</p>
    </H5SectionBlock>

    <H5SectionBlock eyebrow="为什么今天先看它" title="把解释压成手机端能一眼看懂的三段" tone="accent">
      <div v-if="reasonCards.length" class="home-reason-list">
        <article v-for="item in reasonCards" :key="item.title" class="home-reason-item">
          <span class="home-kicker">{{ item.title }}</span>
          <p>{{ item.desc }}</p>
          <div v-if="item.tags?.length" class="home-tag-row compact">
            <span v-for="tag in item.tags" :key="tag" class="home-chip subtle">{{ tag }}</span>
          </div>
        </article>
      </div>
      <H5EmptyState v-else title="主推荐理由待同步" description="刷新后会补齐核心结论、证据来源和风险边界。" />
    </H5SectionBlock>

    <H5SectionBlock eyebrow="延伸观察" title="主线看完后，再从候选里继续往下刷">
      <div v-if="watchlistPreview.length" class="home-watch-list">
        <article v-for="item in watchlistPreview" :key="item.id" class="home-watch-item">
          <div class="home-watch-topline">
            <div>
              <span>{{ item.meta }}</span>
              <strong>{{ item.title }}</strong>
            </div>
            <em>{{ item.risk }}</em>
          </div>
          <p>{{ item.summary }}</p>
          <div v-if="item.tags?.length" class="home-tag-row compact">
            <span v-for="tag in item.tags" :key="tag" class="home-chip subtle">{{ tag }}</span>
          </div>
        </article>
      </div>
      <H5EmptyState v-else title="观察清单待生成" description="登录后这里会出现今天值得继续跟踪的 2-4 个对象。" />
    </H5SectionBlock>

    <H5SectionBlock eyebrow="相关资讯" title="首页只放一条最值得继续读的资讯" tone="accent">
      <button v-if="newsBrief" type="button" class="home-news-brief" @click="openFeaturedArticle">
        <div class="home-news-topline">
          <span>{{ featuredArticle?.meta || "资讯" }}</span>
          <span>{{ featuredArticle?.badge || "公开" }}</span>
        </div>
        <strong>{{ newsBrief.title }}</strong>
        <p>{{ newsBrief.summary }}</p>
        <span class="home-news-cta">进入正文</span>
      </button>
      <H5EmptyState v-else title="资讯焦点待同步" description="稍后刷新，或直接进入资讯页查看最新文章流。" />
    </H5SectionBlock>

    <H5SectionBlock eyebrow="今日行动" title="把今天的下一步集中到一个操作区" tone="soft">
      <H5ActionBar column class="home-action-card">
        <button type="button" class="h5-btn block" @click="goStrategies">继续看观点详情</button>
        <button type="button" class="h5-btn-secondary block" @click="goNews">继续刷资讯</button>
        <button type="button" class="h5-btn-ghost block" @click="goMembership">看会员方案</button>
      </H5ActionBar>
    </H5SectionBlock>

    <H5StickyCta
      :title="stickyTitle"
      :description="stickyDescription"
      :primary-label="stickyPrimaryLabel"
      secondary-label="看资讯"
      @primary="handleStickyPrimary"
      @secondary="goNews"
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
import { getStockRecommendationDetail, getStockRecommendationInsight, listStockRecommendations } from "../../../api/market";
import { listNewsArticles } from "../../../api/news";
import { shouldUseDemoFallback } from "../../../lib/fallback-policy";
import { buildStrategyInsightSections, buildStrategyMetaText, buildStrategyProofTags } from "../../../lib/strategy-version";
import { WATCHLIST_EVENT, isWatchedStock, removeWatchedStock, saveWatchedStock } from "../../../lib/watchlist";
import { useClientAuth } from "../../../shared/auth/client-auth";
import { formatDateTime, mapMemberLevel, mapRiskLevel, mapVIPStatus, normalizeText, resolveVipStage, toArray, truncateText } from "../lib/formatters";
import { buildHomeFeedModel } from "../lib/home-feed.js";
import { fallbackNewsArticles, fallbackQuota, fallbackStockDetails, fallbackStockInsights, fallbackStockRecommendations } from "../lib/mock-data";
import { shapeStrategyDisplayTitle } from "../lib/display-copy.js";
import { resolveHighlightTone } from "../lib/surface-tone.js";

const router = useRouter();
const { isLoggedIn } = useClientAuth();
const useDemoFallback = shouldUseDemoFallback();

const loading = ref(false);
const loadError = ref("");
const rawStocks = ref(useDemoFallback ? [...fallbackStockRecommendations] : []);
const stockDetailMap = ref(useDemoFallback ? { ...fallbackStockDetails } : {});
const stockInsightMap = ref(useDemoFallback ? { ...fallbackStockInsights } : {});
const rawNews = ref(useDemoFallback ? [...fallbackNewsArticles] : []);
const rawQuota = ref(useDemoFallback ? { ...fallbackQuota } : {});
const watchVersion = ref(0);

const todayText = computed(() => new Date().toLocaleDateString("zh-CN", {
  month: "long",
  day: "numeric",
  weekday: "short"
}));

const mappedStocks = computed(() => {
  void watchVersion.value;
  return rawStocks.value.map((item) => {
    const detail = stockDetailMap.value[item.id] || {};
    const insight = stockInsightMap.value[item.id] || {};
    const sections = buildStrategyInsightSections(insight, item.reason_summary || "");
    return {
      id: item.id,
      symbol: item.symbol || "-",
      name: item.name || "未命名标的",
      title: shapeStrategyDisplayTitle({
        title: `${item.symbol || "-"} ${item.name || ""}`.trim(),
        symbol: item.symbol,
        name: item.name
      }),
      risk: mapRiskLevel(item.risk_level),
      summary: sections.whyNow || item.reason_summary || "今日理由待同步",
      tags: buildStrategyProofTags(insight, { limit: 3 }),
      meta: buildStrategyMetaText(insight, formatDateTime, { includeJob: false }) || formatDateTime(item.valid_from),
      takeProfit: detail.take_profit || "目标区间待补充",
      stopLoss: detail.stop_loss || "风控线待补充",
      position: detail.position_range || item.position_range || "轻仓观察",
      reason: sections.proofSource || sections.whyNow || item.reason_summary || "",
      riskBoundary: sections.riskBoundary || "当前未补更多风险边界。",
      insight,
      base: item
    };
  });
});

const featuredArticle = computed(() => {
  const item = rawNews.value[0];
  if (!item) {
    return null;
  }
  const visibility = String(item.visibility || "PUBLIC").toUpperCase();
  return {
    id: item.id,
    title: item.title || "未命名资讯",
    summary: item.summary || item.content || "-",
    meta: `${formatDateTime(item.published_at || item.created_at)} · ${item.category_name || item.category_id || "资讯"}`,
    badge: visibility === "VIP" ? "VIP" : "公开",
    tone: visibility === "VIP" ? "gold" : "brand",
    tags: [item.category_name || item.category_id || "资讯", visibility]
  };
});

const homeModel = computed(() => buildHomeFeedModel({
  stocks: mappedStocks.value,
  article: featuredArticle.value
}));

const leadStock = computed(() => homeModel.value.leadStock);
const marketPulse = computed(() => homeModel.value.marketPulse);
const watchlistPreview = computed(() => homeModel.value.watchlistPreview);
const newsBrief = computed(() => homeModel.value.newsBrief);

const heroTitle = computed(() => leadStock.value ? `今天先看 ${leadStock.value.title}` : "先看今日主推荐，再决定怎么做");
const heroDescription = computed(() => {
  if (!leadStock.value) {
    return isLoggedIn.value ? "主推荐同步后，这里会先给出结论、理由与风险边界。" : "未登录时优先展示移动端主线结构和操作顺序。";
  }
  return leadStock.value.summary;
});
const heroTags = computed(() => leadStock.value?.tags?.length ? leadStock.value.tags : []);
const heroMeta = computed(() => [leadStock.value?.symbol || "", leadStock.value?.meta || "等待同步", ...heroTags.value.slice(0, 2)].filter(Boolean));
const watchButtonText = computed(() => {
  if (!leadStock.value) {
    return "加入关注";
  }
  return isWatchedStock(leadStock.value.id) ? "取消关注" : "加入关注";
});

const reasonCards = computed(() => {
  const item = leadStock.value;
  if (!item) {
    return [];
  }
  return [
    {
      title: "核心结论",
      desc: item.summary,
      tags: item.tags
    },
    {
      title: "证据来源",
      desc: item.reason,
      tags: []
    },
    {
      title: "风险与边界",
      desc: item.riskBoundary,
      tags: [item.stopLoss]
    }
  ].filter((entry) => normalizeText(entry.desc));
});

const recommendationStatus = computed(() => ({
  value: leadStock.value ? "已生成" : isLoggedIn.value ? "待同步" : "登录后解锁",
  note: leadStock.value ? truncateText(leadStock.value.title, 28) : "股票与期货推荐会在登录后优先同步"
}));

const membershipSummary = computed(() => ({
  value: mapMemberLevel(rawQuota.value?.member_level, rawQuota.value?.member_level),
  note: resolveVipStage(rawQuota.value) ? mapVIPStatus(rawQuota.value?.vip_status, rawQuota.value?.member_level) : "未开通或待激活"
}));

const newsSummary = computed(() => ({
  value: featuredArticle.value ? "已同步" : "待刷新",
  note: featuredArticle.value ? truncateText(featuredArticle.value.title, 30) : "先查看焦点文章，再决定是否继续阅读"
}));

const statusCards = computed(() => [
  { label: "推荐状态", ...recommendationStatus.value },
  { label: "会员状态", ...membershipSummary.value },
  { label: "资讯节奏", ...newsSummary.value }
]);

const stickyTitle = computed(() => {
  if (!isLoggedIn.value) {
    return "登录后解锁完整推荐与阅读权限";
  }
  if (resolveVipStage(rawQuota.value)) {
    return "继续进入策略详情，完成今天的判断链";
  }
  return "开通会员后可继续查看更多正文、附件与策略解释";
});

const stickyDescription = computed(() => {
  if (!isLoggedIn.value) {
    return "当前先看公开结构，登录后自动回到 H5 原页面。";
  }
  if (resolveVipStage(rawQuota.value)) {
    return "策略页会继续展开核心结论、风险边界、版本信息与下一步动作。";
  }
  return "会员页会优先解释升级后能得到什么，而不是先给你一堆表格。";
});

const stickyPrimaryLabel = computed(() => {
  if (!isLoggedIn.value) {
    return "立即登录";
  }
  return resolveVipStage(rawQuota.value) ? "去策略页" : "看会员权益";
});

async function loadHomePage() {
  loading.value = true;
  loadError.value = "";

  if (!isLoggedIn.value) {
    rawQuota.value = {};
    rawStocks.value = [];
    stockDetailMap.value = {};
    stockInsightMap.value = {};
    rawNews.value = [];
    loading.value = false;
    return;
  }

  const [stockResult, newsResult, quotaResult] = await Promise.allSettled([
    listStockRecommendations({ page: 1, page_size: 4 }),
    listNewsArticles({ page: 1, page_size: 3 }),
    getMembershipQuota()
  ]);

  const errors = [];

  if (stockResult.status === "fulfilled") {
    rawStocks.value = toArray(stockResult.value?.items, []);
    await hydrateStockDetails(rawStocks.value.slice(0, 4));
  } else {
    errors.push(`推荐加载失败：${stockResult.reason?.message || "unknown error"}`);
  }

  if (newsResult.status === "fulfilled") {
    rawNews.value = toArray(newsResult.value?.items, []);
  } else {
    errors.push(`资讯加载失败：${newsResult.reason?.message || "unknown error"}`);
  }

  if (quotaResult.status === "fulfilled") {
    rawQuota.value = quotaResult.value || {};
  } else {
    errors.push(`会员状态加载失败：${quotaResult.reason?.message || "unknown error"}`);
  }

  if (!rawStocks.value.length && useDemoFallback) {
    rawStocks.value = [...fallbackStockRecommendations];
    stockDetailMap.value = { ...fallbackStockDetails };
    stockInsightMap.value = { ...fallbackStockInsights };
  }
  if (!rawNews.value.length && useDemoFallback) {
    rawNews.value = [...fallbackNewsArticles];
  }
  if (!Object.keys(rawQuota.value || {}).length && useDemoFallback) {
    rawQuota.value = { ...fallbackQuota };
  }

  loadError.value = errors.join("；");
  loading.value = false;
}

async function hydrateStockDetails(items) {
  if (!Array.isArray(items) || items.length === 0) {
    return;
  }
  const nextDetails = { ...stockDetailMap.value };
  const nextInsights = { ...stockInsightMap.value };

  await Promise.all(items.map(async (item) => {
    if (!item?.id) {
      return;
    }
    const [detailResult, insightResult] = await Promise.allSettled([
      getStockRecommendationDetail(item.id),
      getStockRecommendationInsight(item.id)
    ]);
    if (detailResult.status === "fulfilled" && detailResult.value) {
      nextDetails[item.id] = detailResult.value;
    }
    if (insightResult.status === "fulfilled" && insightResult.value) {
      nextInsights[item.id] = insightResult.value.explanation || insightResult.value;
    }
  }));

  stockDetailMap.value = nextDetails;
  stockInsightMap.value = nextInsights;
}

function refreshWatchState() {
  watchVersion.value += 1;
}

function toggleWatch() {
  if (!leadStock.value) {
    return;
  }
  if (isWatchedStock(leadStock.value.id)) {
    removeWatchedStock(leadStock.value.id);
  } else {
    saveWatchedStock({
      id: leadStock.value.id,
      symbol: leadStock.value.symbol,
      name: leadStock.value.name,
      risk_level: leadStock.value.base?.risk_level,
      status: leadStock.value.base?.status,
      reason_summary: leadStock.value.summary,
      valid_from: leadStock.value.base?.valid_from,
      valid_to: leadStock.value.base?.valid_to,
      baseline_take_profit: leadStock.value.takeProfit,
      baseline_stop_loss: leadStock.value.stopLoss,
      baseline_reason_summary: leadStock.value.summary,
      baseline_risk_level: leadStock.value.base?.risk_level,
      baseline_status: leadStock.value.base?.status,
      baseline_strategy_version: leadStock.value.insight?.strategy_version,
      baseline_publish_id: leadStock.value.insight?.publish_id,
      baseline_job_id: leadStock.value.insight?.job_id,
      baseline_trade_date: leadStock.value.insight?.trade_date,
      baseline_publish_version: leadStock.value.insight?.publish_version,
      baseline_agent_opinions: leadStock.value.insight?.agent_opinions
    });
  }
  refreshWatchState();
}

function refreshPage() {
  return loadHomePage();
}

function goStrategies() {
  router.push("/strategies");
}

function goNews() {
  if (featuredArticle.value?.id) {
    router.push({ path: "/news", query: { article: featuredArticle.value.id } });
    return;
  }
  router.push("/news");
}

function openFeaturedArticle() {
  goNews();
}

function goMembership() {
  router.push("/membership");
}

function handleStickyPrimary() {
  if (!isLoggedIn.value) {
    router.push({ path: "/auth", query: { redirect: "/home" } });
    return;
  }
  if (resolveVipStage(rawQuota.value)) {
    goStrategies();
    return;
  }
  goMembership();
}

onMounted(() => {
  loadHomePage();
  if (typeof window !== "undefined") {
    window.addEventListener(WATCHLIST_EVENT, refreshWatchState);
  }
});

onBeforeUnmount(() => {
  if (typeof window !== "undefined") {
    window.removeEventListener(WATCHLIST_EVENT, refreshWatchState);
  }
});
</script>

<style scoped>
.home-page {
  gap: 14px;
}

.home-hero-metrics,
.home-status-grid {
  display: grid;
  gap: 10px;
}

.home-status-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.home-inline-note {
  margin: 0;
  color: var(--h5-text-soft);
  font-size: 12px;
  line-height: 1.7;
}

.home-reason-list,
.home-watch-list {
  display: grid;
  gap: 10px;
}

.home-reason-item,
.home-watch-item {
  padding: 15px 14px;
  border-radius: var(--h5-radius-md);
  background: rgba(20, 52, 95, 0.045);
  display: grid;
  gap: 8px;
}

.home-kicker {
  color: var(--h5-brand);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.home-tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.home-chip {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  color: #fff;
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.08);
}

.home-chip.subtle {
  color: var(--h5-text-sub);
  border-color: rgba(16, 42, 86, 0.06);
  background: rgba(20, 52, 95, 0.04);
}

.home-watch-topline,
.home-news-topline {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.home-watch-topline span {
  color: var(--h5-text-soft);
  font-size: 12px;
}

.home-watch-topline strong {
  display: block;
  margin-top: 4px;
  color: var(--h5-text);
  font-size: 16px;
  line-height: 1.38;
}

.home-watch-topline em {
  font-style: normal;
  color: var(--h5-brand);
  font-size: 12px;
  font-weight: 700;
}

.home-watch-item p,
.home-reason-item p {
  margin: 0;
  color: var(--h5-text-sub);
  font-size: 13px;
  line-height: 1.72;
}

.home-news-brief {
  border: 0;
  width: 100%;
  padding: 18px 16px;
  border-radius: var(--h5-radius);
  background: var(--h5-surface-brand);
  display: grid;
  gap: 12px;
  text-align: left;
  cursor: pointer;
  box-shadow: 0 18px 38px rgba(16, 42, 86, 0.16);
}

.home-news-topline {
  color: rgba(255, 255, 255, 0.74);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.home-news-brief strong {
  margin: 0;
  color: #fff;
  font-size: 22px;
  line-height: 1.34;
}

.home-news-brief p {
  margin: 0;
  color: rgba(255, 255, 255, 0.82);
  font-size: 13px;
  line-height: 1.72;
}

.home-news-cta {
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}

.home-action-card {
  display: grid;
}

@media (max-width: 374px) {
  .home-status-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (prefers-reduced-motion: reduce) {
  .home-news-brief {
    transition: none;
  }
}
</style>
