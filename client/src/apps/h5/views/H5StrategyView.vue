<template>
  <div class="h5-page fade-up strategy-page">
    <div class="h5-page-topline">
      <span class="h5-page-tagline">观点流</span>
      <span>{{ activeDetailID ? "策略简报" : activeKindLabel }}</span>
    </div>

    <template v-if="!activeDetailID">
      <H5HeroCard eyebrow="精选观点" :title="activeKindLabel" :description="listHeroDescription" :meta="strategyHeroMeta" tone="accent">
        <template #actions>
          <H5ActionBar>
            <button type="button" class="h5-btn-secondary" :disabled="loading" @click="loadStrategyPage">
              {{ loading ? "同步中..." : "刷新策略" }}
            </button>
            <button type="button" class="h5-btn-ghost" @click="router.push('/membership')">会员权益</button>
          </H5ActionBar>
        </template>

        <div class="strategy-hero-metrics h5-grid-3">
          <H5SummaryCard
            v-for="(item, index) in strategyHeroStats"
            :key="item.label"
            :label="item.label"
            :value="item.value"
            :note="item.note"
            :tone="resolveHighlightTone(index)"
          />
        </div>
      </H5HeroCard>

      <H5SectionBlock eyebrow="栏目切换" title="先确定要看的策略类型，再沿着内容流继续往下刷" tone="soft">
        <div class="strategy-kind-strip" aria-label="策略栏目">
          <button
            v-for="item in tabs"
            :key="item.key"
            type="button"
            class="strategy-kind-pill"
            :class="{ active: activeKind === item.key }"
            @click="selectKind(item.key)"
          >
            {{ item.label }}
          </button>
        </div>
      </H5SectionBlock>

      <H5SectionBlock v-if="digestItems.length" eyebrow="热看" title="当前栏目最值得先扫一眼的重点" tone="accent">
        <div class="strategy-digest-list">
          <span v-for="item in digestItems" :key="item" class="strategy-digest-item">{{ item }}</span>
        </div>
      </H5SectionBlock>

      <button v-if="leadRow" type="button" class="strategy-lead-card" @click="openDetail(leadRow)">
        <div class="strategy-lead-topline">
          <span>{{ leadRow.sectionLabel }}</span>
          <span>{{ leadRow.meta }}</span>
        </div>
        <h3>{{ leadRow.title }}</h3>
        <p>{{ leadRow.summary }}</p>
        <div class="strategy-lead-footer">
          <div class="strategy-tag-row">
            <span class="strategy-badge" :class="leadRow.tone">{{ leadRow.signalLabel }}</span>
            <span v-for="tag in leadRow.tags" :key="tag" class="strategy-chip">{{ tag }}</span>
          </div>
          <span class="strategy-lead-cta">继续阅读</span>
        </div>
      </button>

      <H5SectionBlock eyebrow="继续看观点" title="像手机内容 App 一样顺着往下刷，先看结论，再点进详情">
        <div v-if="streamRows.length" class="strategy-feed-list">
          <button
            v-for="item in streamRows"
            :key="item.id"
            type="button"
            class="strategy-feed-item"
            @click="openDetail(item)"
          >
            <div class="strategy-feed-item-head">
              <span class="strategy-kicker">{{ item.kicker }}</span>
              <span class="strategy-feed-meta">{{ item.meta }}</span>
            </div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.summary }}</p>
            <div class="strategy-feed-item-foot">
              <div class="strategy-tag-row">
                <span class="strategy-badge" :class="item.tone">{{ item.signalLabel }}</span>
                <span v-for="tag in item.tags" :key="tag" class="strategy-chip subtle">{{ tag }}</span>
              </div>
              <span class="strategy-readmore">查看详情</span>
            </div>
          </button>
        </div>

        <H5EmptyState v-else title="当前没有可读策略" description="刷新后会同步当前策略、期货方案或事件提醒。" />
      </H5SectionBlock>
    </template>

    <template v-else>
      <section class="strategy-detail-head">
        <button type="button" class="strategy-detail-back" @click="closeDetail">返回列表</button>
        <div class="strategy-detail-meta">
          <span>{{ activeKindLabel }}</span>
          <span>{{ detailMetaLine }}</span>
        </div>
        <h1>{{ detailCoreTitle }}</h1>
        <p>{{ detailSummary }}</p>
        <div class="strategy-tag-row">
          <span class="strategy-badge strong" :class="detailToneClass">{{ detailSignalLabel }}</span>
          <span v-for="tag in detailTags" :key="tag" class="strategy-chip">{{ tag }}</span>
        </div>
      </section>

      <section class="strategy-brief-card">
        <div class="strategy-feed-head">
          <div>
            <strong>执行摘要</strong>
            <p>手机端先给结论、执行区间和风险边界，不做桌面式工作台。</p>
          </div>
        </div>
        <div class="strategy-brief-grid">
          <article v-for="item in detailQuickFacts" :key="item.label" class="strategy-brief-item">
            <span>{{ item.label }}</span>
            <strong>{{ item.value }}</strong>
            <p>{{ item.note }}</p>
          </article>
        </div>
      </section>

      <section class="strategy-detail-card">
        <div class="strategy-feed-head">
          <div>
            <strong>为什么看它</strong>
            <p>把策略解释压缩成手机端能快速读完的两段。</p>
          </div>
        </div>
        <div class="strategy-detail-list">
          <article class="strategy-note-item">
            <span class="strategy-kicker">核心理由</span>
            <strong>{{ detailWhyTitle }}</strong>
            <p>{{ detailWhyText }}</p>
          </article>
          <article v-if="detailGuidanceText" class="strategy-note-item">
            <span class="strategy-kicker">执行区间</span>
            <strong>{{ detailGuidanceTitle }}</strong>
            <p>{{ detailGuidanceText }}</p>
          </article>
        </div>
      </section>

      <section class="strategy-detail-card">
        <div class="strategy-feed-head">
          <div>
            <strong>{{ detailRiskLabel }}</strong>
            <p>把边界提前说清楚，避免只看结论不看风控。</p>
          </div>
        </div>
        <article class="strategy-risk-panel" :class="detailToneClass">
          <strong>{{ detailRiskTitle }}</strong>
          <p>{{ detailRiskText }}</p>
        </article>
      </section>

      <section class="strategy-detail-card">
        <div class="strategy-feed-head">
          <div>
            <strong>版本与事件</strong>
            <p>版本时间线和相关事件合并成单列阅读，不做横向对照面板。</p>
          </div>
        </div>

        <div v-if="detailVersionRows.length || detailEventRows.length" class="strategy-timeline">
          <article v-for="item in detailVersionRows" :key="item.key" class="strategy-timeline-item">
            <span class="strategy-timeline-dot" />
            <div class="strategy-timeline-copy">
              <strong>{{ item.title }}</strong>
              <p>{{ item.note }}</p>
            </div>
          </article>
          <article v-for="item in detailEventRows" :key="item.id" class="strategy-timeline-item">
            <span class="strategy-timeline-dot gold" />
            <div class="strategy-timeline-copy">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </div>
          </article>
        </div>

        <H5EmptyState v-else title="暂无更多版本或事件" description="若后端返回版本历史或事件，这里会补齐时间线卡片。" />
      </section>

      <section class="strategy-actions-card">
        <button type="button" class="h5-btn block" :disabled="detailPrimaryDisabled" @click="handlePrimaryAction">
          {{ primaryActionLabel }}
        </button>
        <button v-if="showWatchlistEntry" type="button" class="h5-btn-secondary block" @click="goWatchlist">去我的 &gt; 我的关注</button>
        <button type="button" class="h5-btn-secondary block" @click="router.push('/membership')">查看会员权益</button>
        <button type="button" class="h5-btn-ghost block" @click="closeDetail">返回观点流</button>
      </section>
    </template>

    <H5StickyCta
      :title="stickyTitle"
      :description="stickyDescription"
      :primary-label="primaryActionLabel"
      :secondary-label="secondaryActionLabel"
      @primary="handlePrimaryAction"
      @secondary="handleSecondaryAction"
    />
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import H5ActionBar from "../components/H5ActionBar.vue";
import H5EmptyState from "../components/H5EmptyState.vue";
import H5HeroCard from "../components/H5HeroCard.vue";
import H5SectionBlock from "../components/H5SectionBlock.vue";
import H5StickyCta from "../components/H5StickyCta.vue";
import H5SummaryCard from "../components/H5SummaryCard.vue";
import {
  getFuturesStrategyInsight,
  getFuturesStrategyVersionHistory,
  getMarketEventDetail,
  getStockRecommendationDetail,
  getStockRecommendationInsight,
  getStockRecommendationVersionHistory,
  listFuturesStrategies,
  listMarketEvents,
  listStockRecommendations
} from "../../../api/market";
import { getMembershipQuota } from "../../../api/membership";
import { shouldUseDemoFallback } from "../../../lib/fallback-policy";
import { buildProfileModuleRoute } from "../../../lib/profile-modules";
import { WATCHLIST_EVENT, isWatchedStock, removeWatchedStock, saveWatchedStock } from "../../../lib/watchlist";
import { useClientAuth } from "../../../shared/auth/client-auth";
import { shapeStrategyDisplayTitle } from "../lib/display-copy.js";
import { resolveStrategyAccessSummary } from "../lib/membership-display.js";
import {
  buildFallbackStrategyVersionHistory,
  buildStrategyInsightSections,
  buildStrategyMetaText,
  buildStrategyProofTags,
  mapStrategyVersionHistory
} from "../../../lib/strategy-version";
import { formatDateTime, mapRiskLevel, resolveVipStage, toArray } from "../lib/formatters";
import { mergeRecordById } from "../lib/page-state";
import { buildStrategyFeedModel } from "../lib/strategy-feed.js";
import { resolveHighlightTone } from "../lib/surface-tone.js";
import {
  fallbackMarketEvents,
  fallbackQuota,
  fallbackStockDetails,
  fallbackStockInsights,
  fallbackStockRecommendations
} from "../lib/mock-data";

const router = useRouter();
const route = useRoute();
const { isLoggedIn } = useClientAuth();
const useDemoFallback = shouldUseDemoFallback();

const loading = ref(false);
const rawQuota = ref(useDemoFallback ? { ...fallbackQuota } : {});
const rawStocks = ref(useDemoFallback ? [...fallbackStockRecommendations] : []);
const rawFutures = ref(useDemoFallback ? [{
  id: "fs_local_001",
  contract: "IF2603",
  name: "股指趋势跟踪",
  direction: "LONG",
  risk_level: "MEDIUM",
  reason_summary: "趋势与量价结构一致",
  valid_from: "2026-02-27T09:30:00+08:00"
}] : []);
const rawEvents = ref(useDemoFallback ? [...fallbackMarketEvents] : []);
const stockDetailMap = ref(useDemoFallback ? { ...fallbackStockDetails } : {});
const stockInsightMap = ref(useDemoFallback ? { ...fallbackStockInsights } : {});
const stockVersionMap = ref({});
const futuresInsightMap = ref(useDemoFallback ? {
  fs_local_001: {
    strategy: { id: "fs_local_001", contract: "IF2603", name: "股指趋势跟踪", direction: "LONG", risk_level: "MEDIUM", reason_summary: "趋势与量价结构一致，等待突破确认后执行。" },
    guidance: { entry_range: "3520 - 3545", take_profit_range: "3590 - 3620", stop_loss_range: "3470 - 3490", invalid_condition: "15分钟跌破3470且量能放大" },
    explanation: { confidence_reason: "趋势和量价结构同步改善，适合继续跟踪突破确认。", risk_boundary: "跌破关键区间则停止沿用旧逻辑。", strategy_version: "futures-demo", generated_at: "2026-02-27T10:30:00+08:00" }
  }
} : {});
const futuresVersionMap = ref({});
const eventDetailMap = ref(useDemoFallback ? Object.fromEntries(fallbackMarketEvents.map((item) => [item.id, item])) : {});
const watchVersion = ref(0);

const tabs = [
  { key: "stock", label: "股票策略" },
  { key: "futures", label: "期货策略" },
  { key: "event", label: "事件提醒" }
];

const activeKind = computed(() => String(route.query.kind || "stock"));
const activeKindLabel = computed(() => tabs.find((item) => item.key === activeKind.value)?.label || "股票策略");
const activeDetailID = computed(() => String(route.query.id || ""));
const currentRows = computed(() => {
  if (activeKind.value === "futures") {
    return rawFutures.value.map((item) => ({
      id: item.id,
      title: shapeStrategyDisplayTitle({
        title: `${item.contract || "-"} ${item.name || ""}`.trim(),
        contract: item.contract,
        name: item.name
      }),
      summary: item.reason_summary || "等待更多执行理由",
      meta: formatDateTime(item.valid_from || item.created_at),
      badge: mapRiskLevel(item.risk_level),
      tone: "brand",
      tags: [item.direction || "方向待定", mapRiskLevel(item.risk_level)]
    }));
  }
  if (activeKind.value === "event") {
    return rawEvents.value.map((item) => ({
      id: item.id,
      title: `${item.symbol || "ALL"} · ${item.event_type || "事件"}`,
      summary: item.summary || "-",
      meta: formatDateTime(item.created_at),
      badge: "事件",
      tone: "gold",
      tags: [item.source || "system"]
    }));
  }
  return rawStocks.value.map((item) => {
    const insight = stockInsightMap.value[item.id] || {};
    const sections = buildStrategyInsightSections(insight, item.reason_summary || "");
    return {
      id: item.id,
      title: shapeStrategyDisplayTitle({
        title: `${item.symbol || "-"} ${item.name || ""}`.trim(),
        symbol: item.symbol,
        name: item.name
      }),
      summary: sections.whyNow || item.reason_summary || "等待更多理由",
      meta: buildStrategyMetaText(insight, formatDateTime, { includeJob: false }) || formatDateTime(item.valid_from),
      badge: mapRiskLevel(item.risk_level),
      tone: "brand",
      tags: buildStrategyProofTags(insight, { limit: 3 })
    };
  });
});
const feedModel = computed(() => buildStrategyFeedModel(currentRows.value, activeKindLabel.value));
const leadRow = computed(() => feedModel.value.lead);
const streamRows = computed(() => feedModel.value.stream);
const digestItems = computed(() => feedModel.value.digest);
const strategyHeroMeta = computed(() => [
  activeKindLabel.value,
  leadRow.value?.meta || "",
  leadRow.value?.signalLabel || ""
].filter(Boolean));
const strategyHeroStats = computed(() => [
  {
    label: "当前栏目",
    value: activeKindLabel.value,
    note: leadRow.value ? "已生成主策略卡" : "等待内容同步"
  },
  {
    label: "策略数量",
    value: `${currentRows.value.length} 条`,
    note: "先看主结论，再进入详情"
  },
  {
    label: "阅读状态",
    ...resolveStrategyAccessSummary(rawQuota.value, { isLoggedIn: isLoggedIn.value, loading: loading.value })
  }
]);

const activeStock = computed(() => rawStocks.value.find((item) => item.id === activeDetailID.value) || null);
const activeFutures = computed(() => rawFutures.value.find((item) => item.id === activeDetailID.value) || null);
const activeEvent = computed(() => rawEvents.value.find((item) => item.id === activeDetailID.value) || eventDetailMap.value[activeDetailID.value] || null);
const activeStockInsight = computed(() => stockInsightMap.value[activeDetailID.value] || null);
const activeFuturesBundle = computed(() => futuresInsightMap.value[activeDetailID.value] || null);
const activeEventDetail = computed(() => eventDetailMap.value[activeDetailID.value] || activeEvent.value || null);
const isVIPUser = computed(() => resolveVipStage(rawQuota.value));

const listHeroDescription = computed(() => {
  if (activeKind.value === "event") {
    return "先看触发中的事件提醒，再决定要不要继续追踪相关策略。";
  }
  if (activeKind.value === "futures") {
    return "期货观点更强调方向、执行区间和失效条件，适合单列快速阅读。";
  }
  return "把当下值得先看的股票观点压成一条流，先看结论，再看原因和边界。";
});

const detailMetaLine = computed(() => {
  if (activeKind.value === "event") {
    return formatDateTime(activeEventDetail.value?.created_at);
  }
  if (activeKind.value === "futures") {
    return buildStrategyMetaText(activeFuturesBundle.value?.explanation, formatDateTime, { includeJob: false }) || formatDateTime(activeFutures.value?.valid_from);
  }
  return buildStrategyMetaText(activeStockInsight.value, formatDateTime, { includeJob: false }) || formatDateTime(activeStock.value?.valid_from);
});
const detailCoreTitle = computed(() => {
  if (activeKind.value === "event") {
    return `${activeEventDetail.value?.symbol || "ALL"} · ${activeEventDetail.value?.event_type || "事件提醒"}`;
  }
  if (activeKind.value === "futures") {
    return shapeStrategyDisplayTitle({
      title: `${activeFutures.value?.contract || "-"} ${activeFutures.value?.name || ""}`.trim(),
      contract: activeFutures.value?.contract,
      name: activeFutures.value?.name
    });
  }
  return shapeStrategyDisplayTitle({
    title: `${activeStock.value?.symbol || "-"} ${activeStock.value?.name || ""}`.trim(),
    symbol: activeStock.value?.symbol,
    name: activeStock.value?.name
  });
});
const detailSummary = computed(() => {
  if (activeKind.value === "event") {
    return activeEventDetail.value?.summary || "事件详情待同步";
  }
  if (activeKind.value === "futures") {
    return activeFuturesBundle.value?.explanation?.confidence_reason || activeFutures.value?.reason_summary || "期货方案理由待同步";
  }
  return activeStockInsight.value?.confidence_reason || activeStock.value?.reason_summary || "股票策略理由待同步";
});
const detailTags = computed(() => {
  if (activeKind.value === "event") {
    return [activeEventDetail.value?.source || "system"].filter(Boolean);
  }
  if (activeKind.value === "futures") {
    return [activeFutures.value?.direction || "方向待定", mapRiskLevel(activeFutures.value?.risk_level)];
  }
  return buildStrategyProofTags(activeStockInsight.value, { limit: 4 });
});
const detailWhyText = computed(() => {
  if (activeKind.value === "event") {
    return activeEventDetail.value?.trigger_rule || "等待更多事件触发说明";
  }
  if (activeKind.value === "futures") {
    return activeFuturesBundle.value?.strategy?.reason_summary || activeFuturesBundle.value?.explanation?.confidence_reason || "等待更多期货方案解释";
  }
  return buildStrategyInsightSections(activeStockInsight.value, activeStock.value?.reason_summary || "").proofSource || detailSummary.value;
});
const detailGuidanceText = computed(() => {
  if (activeKind.value === "futures") {
    const guidance = activeFuturesBundle.value?.guidance || {};
    return `入场 ${guidance.entry_range || "-"} · 止盈 ${guidance.take_profit_range || "-"} · 止损 ${guidance.stop_loss_range || "-"}`;
  }
  if (activeKind.value === "stock") {
    const detail = stockDetailMap.value[activeDetailID.value] || {};
    return `目标 ${detail.take_profit || "-"} · 止损 ${detail.stop_loss || "-"}${detail.position_range ? ` · 仓位 ${detail.position_range}` : ""}`;
  }
  return "";
});
const detailRiskLabel = computed(() => activeKind.value === "event" ? "触发条件" : "风险边界");
const detailRiskText = computed(() => {
  if (activeKind.value === "event") {
    return activeEventDetail.value?.trigger_rule || "等待更多触发条件";
  }
  if (activeKind.value === "futures") {
    return activeFuturesBundle.value?.guidance?.invalid_condition || activeFuturesBundle.value?.explanation?.risk_boundary || "等待更多风险边界";
  }
  return buildStrategyInsightSections(activeStockInsight.value, activeStock.value?.reason_summary || "").riskBoundary || "等待更多风险边界";
});
const detailVersionRows = computed(() => {
  if (activeKind.value === "event") {
    return [];
  }
  if (activeKind.value === "futures") {
    return futuresVersionMap.value[activeDetailID.value] || [];
  }
  return stockVersionMap.value[activeDetailID.value] || [];
});
const detailEventRows = computed(() => activeKind.value === "event" ? [] : rawEvents.value.slice(0, 3).map((item) => ({
  id: item.id,
  title: `${item.symbol || "ALL"} · ${item.event_type || "事件"}`,
  desc: item.summary || item.trigger_rule || "-"
})));
const detailSignalLabel = computed(() => {
  if (activeKind.value === "event") {
    return "事件提醒";
  }
  if (activeKind.value === "futures") {
    return mapRiskLevel(activeFutures.value?.risk_level);
  }
  return mapRiskLevel(activeStock.value?.risk_level);
});
const detailToneClass = computed(() => activeKind.value === "event" ? "gold" : "brand");
const detailWhyTitle = computed(() => activeKind.value === "event" ? "触发原因" : "推荐理由");
const detailGuidanceTitle = computed(() => activeKind.value === "futures" ? "执行区间" : "计划区间");
const detailRiskTitle = computed(() => activeKind.value === "event" ? "需要关注的触发条件" : "失效与回撤边界");
const showWatchlistEntry = computed(() => activeDetailID.value && activeKind.value === "stock");
const activeStockWatchLabel = computed(() => {
  void watchVersion.value;
  if (activeKind.value !== "stock" || !activeStock.value) {
    return "仅股票可关注";
  }
  return isWatchedStock(activeStock.value.id) ? "取消关注" : "加入关注";
});
const detailPrimaryDisabled = computed(() => showWatchlistEntry.value && !activeStock.value);

const detailQuickFacts = computed(() => {
  if (activeKind.value === "event") {
    return [
      { label: "事件级别", value: activeEventDetail.value?.event_type || "事件", note: activeEventDetail.value?.source || "system" },
      { label: "触发对象", value: activeEventDetail.value?.symbol || "ALL", note: "优先确认是否影响当前持仓或观察清单" },
      { label: "建议动作", value: "先核对影响", note: "事件提醒更适合先确认再联动后续策略" }
    ];
  }
  if (activeKind.value === "futures") {
    return [
      { label: "方向", value: activeFutures.value?.direction || "待定", note: "先确认方向，再看执行区间" },
      { label: "风险", value: mapRiskLevel(activeFutures.value?.risk_level), note: "波动更快，边界一定要前置" },
      { label: "优先动作", value: "观察突破确认", note: "先看价格是否进入计划区间" }
    ];
  }
  return [
    { label: "风险", value: mapRiskLevel(activeStock.value?.risk_level), note: "先看风险，再决定是否继续跟踪" },
    { label: "仓位", value: stockDetailMap.value[activeDetailID.value]?.position_range || activeStock.value?.position_range || "轻仓观察", note: "仓位建议只保留手机端最关键字段" },
    { label: "优先动作", value: "继续跟踪", note: "确认理由与边界后再决定是否进入会员深读" }
  ];
});

const stickyTitle = computed(() => {
  if (showWatchlistEntry.value) {
    return isWatchedStock(activeStock.value?.id) ? "这条策略已进入你的关注清单" : "这条策略值得继续看时，可以先加入关注";
  }
  return activeDetailID.value ? "看完策略简报后，把动作收在最后一个 CTA" : "先选一条观点，再进入单页策略简报";
});
const stickyDescription = computed(() => {
  if (showWatchlistEntry.value) {
    return "我的关注详情会把状态、资讯和风险边界集中到一条持续跟踪链路里。";
  }
  return isVIPUser.value ? "当前会员状态已同步，可继续阅读和跟踪。" : "开通会员后，可继续解锁更多正文、附件和策略解释。";
});
const primaryActionLabel = computed(() => {
  if (!activeDetailID.value) {
    return "查看一条观点";
  }
  if (showWatchlistEntry.value) {
    return activeStockWatchLabel.value;
  }
  return "查看会员权益";
});
const secondaryActionLabel = computed(() => {
  if (showWatchlistEntry.value) {
    return "去我的 > 我的关注";
  }
  return activeDetailID.value ? "返回列表" : "去资讯页";
});

async function loadStrategyPage() {
  loading.value = true;
  const [stockResult, futuresResult, eventResult, quotaResult] = await Promise.allSettled([
    listStockRecommendations({ page: 1, page_size: 8 }),
    listFuturesStrategies({ page: 1, page_size: 8 }),
    listMarketEvents({ page: 1, page_size: 8 }),
    getMembershipQuota()
  ]);

  if (stockResult.status === "fulfilled") {
    rawStocks.value = toArray(stockResult.value?.items, []);
    await hydrateStockInsights(rawStocks.value.slice(0, 6));
  } else if (useDemoFallback) {
    rawStocks.value = [...fallbackStockRecommendations];
    stockDetailMap.value = { ...fallbackStockDetails };
    stockInsightMap.value = { ...fallbackStockInsights };
  }

  if (futuresResult.status === "fulfilled") {
    rawFutures.value = toArray(futuresResult.value?.items, []);
  }
  if (eventResult.status === "fulfilled") {
    rawEvents.value = toArray(eventResult.value?.items, []);
  } else if (useDemoFallback) {
    rawEvents.value = [...fallbackMarketEvents];
  }
  if (quotaResult.status === "fulfilled") {
    rawQuota.value = quotaResult.value || {};
  }

  loading.value = false;
}

async function hydrateStockInsights(items) {
  const nextDetailMap = { ...stockDetailMap.value };
  const nextInsightMap = { ...stockInsightMap.value };
  await Promise.all(items.map(async (item) => {
    if (!item?.id) return;
    const [detailResult, insightResult] = await Promise.allSettled([
      getStockRecommendationDetail(item.id),
      getStockRecommendationInsight(item.id)
    ]);
    if (detailResult.status === "fulfilled") {
      nextDetailMap[item.id] = detailResult.value;
    }
    if (insightResult.status === "fulfilled") {
      nextInsightMap[item.id] = insightResult.value?.explanation || insightResult.value;
    }
  }));
  stockDetailMap.value = nextDetailMap;
  stockInsightMap.value = nextInsightMap;
}

function selectKind(kind) {
  router.replace({ query: { kind, id: undefined } });
}

function openDetail(item) {
  router.replace({ query: { kind: activeKind.value, id: item.id } });
}

function closeDetail() {
  router.replace({ query: { kind: activeKind.value, id: undefined } });
}

async function loadDetail(kind, id) {
  if (!id) return;
  if (kind === "stock") {
    const [detailResult, insightResult, versionResult] = await Promise.allSettled([
      getStockRecommendationDetail(id),
      getStockRecommendationInsight(id),
      getStockRecommendationVersionHistory(id)
    ]);
    if (detailResult.status === "fulfilled") {
      stockDetailMap.value = { ...stockDetailMap.value, [id]: detailResult.value };
      rawStocks.value = mergeRecordById(rawStocks.value, {
        id,
        symbol: detailResult.value?.symbol || detailResult.value?.stock_symbol,
        name: detailResult.value?.name || detailResult.value?.stock_name,
        risk_level: detailResult.value?.risk_level,
        reason_summary: detailResult.value?.reason_summary,
        valid_from: detailResult.value?.valid_from,
        position_range: detailResult.value?.position_range
      });
    }
    if (insightResult.status === "fulfilled") {
      stockInsightMap.value = { ...stockInsightMap.value, [id]: insightResult.value?.explanation || insightResult.value };
      rawStocks.value = mergeRecordById(rawStocks.value, {
        id,
        symbol: insightResult.value?.detail?.symbol || insightResult.value?.detail?.stock_symbol,
        name: insightResult.value?.detail?.name || insightResult.value?.detail?.stock_name,
        risk_level: insightResult.value?.detail?.risk_level,
        reason_summary: insightResult.value?.detail?.reason_summary || insightResult.value?.explanation?.confidence_reason,
        valid_from: insightResult.value?.detail?.valid_from || insightResult.value?.explanation?.generated_at,
        position_range: insightResult.value?.detail?.position_range
      });
      if (versionResult.status !== "fulfilled" && stockInsightMap.value[id]) {
        stockVersionMap.value = {
          ...stockVersionMap.value,
          [id]: buildFallbackStrategyVersionHistory(stockInsightMap.value[id], { formatDateTime })
        };
      }
    }
    if (versionResult.status === "fulfilled") {
      stockVersionMap.value = {
        ...stockVersionMap.value,
        [id]: mapStrategyVersionHistory(versionResult.value?.items || [], formatDateTime)
      };
    }
    return;
  }
  if (kind === "futures") {
    const [insightResult, versionResult] = await Promise.allSettled([
      getFuturesStrategyInsight(id),
      getFuturesStrategyVersionHistory(id)
    ]);
    if (insightResult.status === "fulfilled") {
      futuresInsightMap.value = { ...futuresInsightMap.value, [id]: insightResult.value };
      rawFutures.value = mergeRecordById(rawFutures.value, {
        id,
        contract: insightResult.value?.strategy?.contract,
        name: insightResult.value?.strategy?.name,
        direction: insightResult.value?.strategy?.direction,
        risk_level: insightResult.value?.strategy?.risk_level,
        reason_summary: insightResult.value?.strategy?.reason_summary || insightResult.value?.explanation?.confidence_reason,
        valid_from: insightResult.value?.strategy?.valid_from || insightResult.value?.explanation?.generated_at
      });
    }
    if (versionResult.status === "fulfilled") {
      futuresVersionMap.value = {
        ...futuresVersionMap.value,
        [id]: mapStrategyVersionHistory(versionResult.value?.items || [], formatDateTime)
      };
    }
    return;
  }
  const eventResult = await getMarketEventDetail(id).catch(() => null);
  if (eventResult) {
    eventDetailMap.value = { ...eventDetailMap.value, [id]: eventResult };
  }
}

function handlePrimaryAction() {
  if (!activeDetailID.value) {
    if (leadRow.value) {
      openDetail(leadRow.value);
    }
    return;
  }
  if (showWatchlistEntry.value) {
    toggleActiveStockWatch();
    return;
  }
  router.push("/membership");
}

function handleSecondaryAction() {
  if (showWatchlistEntry.value) {
    goWatchlist();
    return;
  }
  if (activeDetailID.value) {
    closeDetail();
    return;
  }
  goNews();
}

function refreshWatchState() {
  watchVersion.value += 1;
}

function toggleActiveStockWatch() {
  if (activeKind.value !== "stock" || !activeStock.value) {
    return;
  }
  if (isWatchedStock(activeStock.value.id)) {
    removeWatchedStock(activeStock.value.id);
    refreshWatchState();
    return;
  }

  saveWatchedStock({
    id: activeStock.value.id,
    symbol: activeStock.value.symbol,
    name: activeStock.value.name,
    risk_level: activeStock.value.risk_level,
    status: activeStock.value.status,
    reason_summary: detailSummary.value,
    valid_from: activeStock.value.valid_from,
    valid_to: activeStock.value.valid_to,
    baseline_take_profit: stockDetailMap.value[activeDetailID.value]?.take_profit,
    baseline_stop_loss: stockDetailMap.value[activeDetailID.value]?.stop_loss,
    baseline_reason_summary: detailSummary.value,
    baseline_risk_level: activeStock.value.risk_level,
    baseline_status: activeStock.value.status,
    baseline_strategy_version: activeStockInsight.value?.strategy_version,
    baseline_publish_id: activeStockInsight.value?.publish_id,
    baseline_job_id: activeStockInsight.value?.job_id,
    baseline_trade_date: activeStockInsight.value?.trade_date,
    baseline_publish_version: activeStockInsight.value?.publish_version,
    baseline_agent_opinions: activeStockInsight.value?.agent_opinions
  });
  refreshWatchState();
}

function goWatchlist() {
  router.push(buildProfileModuleRoute("watchlist"));
}

function goNews() {
  router.push("/news");
}

watch(() => route.query.id, (value) => {
  if (value) {
    loadDetail(activeKind.value, String(value));
  }
}, { immediate: true });

onMounted(() => {
  loadStrategyPage().then(() => {
    if (activeDetailID.value) {
      return loadDetail(activeKind.value, activeDetailID.value);
    }
    return null;
  });
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
.strategy-page {
  gap: 12px;
}

.strategy-hero-metrics {
  display: grid;
  gap: 10px;
}

.strategy-stream-hero,
.strategy-digest-card,
.strategy-feed-card,
.strategy-detail-head,
.strategy-brief-card,
.strategy-detail-card,
.strategy-actions-card {
  border: 1px solid var(--h5-line);
  border-radius: var(--h5-radius);
  background: var(--h5-panel-bg);
  box-shadow: var(--h5-shadow);
}

.strategy-stream-hero,
.strategy-feed-card,
.strategy-detail-head,
.strategy-brief-card,
.strategy-detail-card,
.strategy-actions-card {
  padding: 20px 18px;
}

.strategy-stream-hero,
.strategy-stream-copy,
.strategy-feed-card,
.strategy-detail-head,
.strategy-brief-card,
.strategy-detail-card,
.strategy-actions-card,
.strategy-digest-card {
  display: grid;
  gap: 14px;
}

.strategy-kicker {
  color: var(--h5-brand);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.strategy-stream-copy h2,
.strategy-detail-head h1 {
  margin: 0;
  color: var(--h5-text);
  font-size: clamp(28px, 7vw, 36px);
  line-height: 1.18;
  letter-spacing: -0.03em;
}

.strategy-stream-copy p,
.strategy-feed-head p,
.strategy-detail-head p,
.strategy-brief-item p,
.strategy-note-item p,
.strategy-timeline-copy p,
.strategy-risk-panel p {
  margin: 0;
  color: var(--h5-text-sub);
  font-size: 13px;
  line-height: 1.72;
}

.strategy-stream-actions,
.strategy-lead-topline,
.strategy-lead-footer,
.strategy-feed-head,
.strategy-feed-item-head,
.strategy-feed-item-foot,
.strategy-detail-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.strategy-stream-actions,
.strategy-feed-item-foot,
.strategy-tag-row {
  flex-wrap: wrap;
}

.strategy-kind-strip {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 2px;
  scrollbar-width: none;
}

.strategy-kind-strip::-webkit-scrollbar {
  display: none;
}

.strategy-kind-pill {
  flex: 0 0 auto;
  min-height: 38px;
  padding: 0 16px;
  border: 1px solid rgba(16, 42, 86, 0.08);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  color: var(--h5-text-sub);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.strategy-kind-pill.active {
  color: #fff;
  background: linear-gradient(180deg, #15345f 0%, #214879 100%);
  border-color: transparent;
  box-shadow: 0 10px 22px rgba(16, 42, 86, 0.16);
}

.strategy-digest-card {
  padding: 14px 16px;
}

.strategy-digest-list {
  display: grid;
  gap: 8px;
}

.strategy-digest-item {
  position: relative;
  padding-left: 14px;
  color: var(--h5-text-sub);
  font-size: 13px;
  line-height: 1.6;
}

.strategy-digest-item::before {
  content: "";
  position: absolute;
  left: 0;
  top: 8px;
  width: 5px;
  height: 5px;
  border-radius: 999px;
  background: var(--h5-gold);
}

.strategy-lead-card {
  width: 100%;
  border: 0;
  padding: 20px 18px;
  border-radius: 28px;
  background: var(--h5-surface-brand);
  color: #fff;
  display: grid;
  gap: 12px;
  text-align: left;
  cursor: pointer;
  box-shadow: 0 20px 42px rgba(16, 42, 86, 0.18);
}

.strategy-lead-topline,
.strategy-detail-meta {
  flex-wrap: wrap;
  color: rgba(255, 255, 255, 0.74);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.strategy-lead-card h3,
.strategy-feed-item strong,
.strategy-feed-head strong,
.strategy-brief-item strong,
.strategy-note-item strong,
.strategy-risk-panel strong,
.strategy-timeline-copy strong {
  margin: 0;
  color: inherit;
}

.strategy-lead-card h3 {
  font-size: 24px;
  line-height: 1.28;
}

.strategy-lead-card p {
  margin: 0;
  color: rgba(255, 255, 255, 0.82);
  font-size: 14px;
  line-height: 1.78;
}

.strategy-tag-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.strategy-badge,
.strategy-chip {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  white-space: nowrap;
}

.strategy-badge {
  color: var(--h5-brand);
  background: rgba(20, 52, 95, 0.08);
}

.strategy-badge.gold {
  color: var(--h5-warning);
  background: rgba(184, 137, 61, 0.14);
}

.strategy-badge.strong.brand {
  color: #fff;
  background: rgba(255, 255, 255, 0.14);
}

.strategy-badge.strong.gold {
  color: #fff;
  background: rgba(219, 183, 101, 0.22);
}

.strategy-chip {
  color: rgba(255, 255, 255, 0.88);
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.08);
}

.strategy-chip.subtle {
  color: var(--h5-text-sub);
  border: 1px solid rgba(16, 42, 86, 0.06);
  background: rgba(20, 52, 95, 0.04);
}

.strategy-lead-cta,
.strategy-readmore,
.strategy-feed-meta {
  font-size: 12px;
  font-weight: 700;
}

.strategy-feed-meta,
.strategy-readmore {
  color: var(--h5-text-soft);
}

.strategy-feed-list,
.strategy-detail-list,
.strategy-timeline {
  display: grid;
  gap: 10px;
}

.strategy-feed-item {
  padding: 14px 0;
  border: 0;
  border-bottom: 1px solid rgba(16, 42, 86, 0.08);
  background: transparent;
  display: grid;
  gap: 8px;
  text-align: left;
  cursor: pointer;
}

.strategy-feed-list .strategy-feed-item:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}

.strategy-feed-item:first-child {
  padding-top: 0;
}

.strategy-feed-item strong {
  color: var(--h5-text);
  font-size: 17px;
  line-height: 1.4;
}

.strategy-feed-item p {
  margin: 0;
  color: var(--h5-text-sub);
  font-size: 13px;
  line-height: 1.7;
}

.strategy-feed-head {
  align-items: flex-start;
}

.strategy-feed-head strong,
.strategy-note-item strong,
.strategy-risk-panel strong,
.strategy-timeline-copy strong {
  color: var(--h5-text);
  font-size: 17px;
  line-height: 1.4;
}

.strategy-feed-head span {
  color: var(--h5-text-soft);
  font-size: 12px;
}

.strategy-detail-head {
  background: var(--h5-surface-brand);
  color: #fff;
}

.strategy-detail-back {
  min-height: 34px;
  width: fit-content;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.strategy-brief-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.strategy-brief-item {
  padding: 14px;
  border-radius: var(--h5-radius-sm);
  background: rgba(20, 52, 95, 0.045);
  display: grid;
  gap: 6px;
}

.strategy-brief-item span {
  color: var(--h5-text-soft);
  font-size: 11px;
}

.strategy-brief-item strong {
  color: var(--h5-text);
  font-size: 16px;
}

.strategy-note-item {
  padding: 15px 14px;
  border-radius: var(--h5-radius-md);
  background: rgba(20, 52, 95, 0.045);
  display: grid;
  gap: 8px;
}

.strategy-risk-panel {
  padding: 16px;
  border-radius: var(--h5-radius-md);
  display: grid;
  gap: 8px;
}

.strategy-risk-panel.brand {
  background: rgba(20, 52, 95, 0.06);
}

.strategy-risk-panel.gold {
  background: rgba(184, 137, 61, 0.12);
}

.strategy-timeline-item {
  display: grid;
  grid-template-columns: 14px minmax(0, 1fr);
  gap: 12px;
  align-items: start;
}

.strategy-timeline-dot {
  width: 10px;
  height: 10px;
  margin-top: 7px;
  border-radius: 999px;
  background: var(--h5-brand);
  box-shadow: 0 0 0 4px rgba(20, 52, 95, 0.08);
}

.strategy-timeline-dot.gold {
  background: var(--h5-gold);
  box-shadow: 0 0 0 4px rgba(184, 137, 61, 0.12);
}

.strategy-timeline-copy {
  padding-bottom: 12px;
  border-bottom: 1px solid rgba(16, 42, 86, 0.08);
  display: grid;
  gap: 6px;
}

.strategy-timeline .strategy-timeline-item:last-child .strategy-timeline-copy {
  border-bottom: 0;
  padding-bottom: 0;
}

.strategy-actions-card {
  gap: 10px;
}

@media (min-width: 521px) {
  .strategy-page {
    gap: 14px;
  }

  .strategy-hero-metrics {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .strategy-kind-strip {
    flex-wrap: wrap;
    overflow: visible;
  }

  .strategy-digest-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .strategy-actions-card {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    align-items: stretch;
  }

  .strategy-actions-card .block {
    min-width: 0;
  }

  .strategy-actions-card .block:last-child {
    grid-column: 1 / -1;
  }
}

@media (max-width: 374px) {
  .strategy-brief-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (prefers-reduced-motion: reduce) {
  .strategy-kind-pill,
  .strategy-feed-item,
  .strategy-lead-card {
    transition: none;
  }
}
</style>
