<template>
  <section class="search-page fade-up" :class="{ 'search-page-h5': isH5SearchSurface }">
    <header class="search-hero card">
      <div>
        <div class="finance-pill-row">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">统一搜索</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">{{ scopeLabel }}</span>
        </div>
        <p class="hero-kicker">搜索结果承接页</p>
        <h1>{{ searchPageTitle }}</h1>
        <p class="section-subtitle">{{ searchPageSubtitle }}</p>
      </div>
      <div class="search-hero-meta">
        <span class="finance-pill finance-pill-roomy finance-pill-info">{{ resultSummaryLabel }}</span>
      </div>
    </header>

    <article v-if="bestMatch" class="card search-best-match">
      <div class="search-best-match-copy">
        <p class="hero-kicker">{{ bestMatch.isFocused ? "已定位目标结果" : "最佳命中" }}</p>
        <h2>{{ bestMatch.item.title }}</h2>
        <p>{{ bestMatch.item.summary }}</p>
        <div class="search-best-match-meta">
          <span class="finance-pill finance-pill-compact finance-pill-info">{{ bestMatch.group.title }}</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ bestMatch.item.meta }}</span>
        </div>
        <p class="search-best-match-reason">{{ bestMatch.reason }}</p>
      </div>
      <div class="search-best-match-actions">
        <button type="button" class="finance-mini-btn finance-mini-btn-primary" @click="openSearchItem(bestMatch.group.key, bestMatch.item)">
          进入详情
        </button>
        <button
          v-if="isLoggedIn && isForecastable(bestMatch.group.key)"
          type="button"
          class="finance-mini-btn finance-mini-btn-primary"
          style="background: #0f172a; border-color: #0f172a;"
          :disabled="submitting"
          @click="requestForecast(bestMatch.group.key, bestMatch.item)"
        >
          {{ submitting ? "提交分析中..." : "一键深度推演" }}
        </button>
        <button
          v-if="activeTab !== bestMatch.group.key"
          type="button"
          class="finance-mini-btn finance-mini-btn-soft"
          @click="activateTab(bestMatch.group.key)"
        >
          查看该分组
        </button>
      </div>
    </article>

    <nav class="search-tabs card">
      <button
        v-for="item in tabs"
        :key="item.key"
        type="button"
        class="search-tab"
        :class="{ active: activeTab === item.key }"
        @click="activeTab = item.key"
      >
        <span>{{ item.label }}</span>
        <strong>{{ item.count }}</strong>
      </button>
    </nav>

    <article
      v-if="!showEmptyPrompt && !loading && !errorMessage"
      class="card search-overview"
    >
      <article v-for="item in overviewItems" :key="item.key" class="search-overview-item">
        <span>{{ item.label }}</span>
        <strong>{{ item.count }}</strong>
        <p>{{ item.note }}</p>
      </article>
    </article>

    <StatePanel
      v-if="showEmptyPrompt"
      tone="info"
      eyebrow="统一搜索"
      title="请输入至少两个字符开始搜索"
      description="顶部搜索条会实时联想，进入结果页后按分组查看完整结果。"
      compact
    />
    <StatePanel
      v-else-if="loading"
      tone="info"
      eyebrow="统一搜索"
      title="正在加载搜索结果"
      description="股票、期货策略和资讯结果正在同步。"
      compact
    />
    <StatePanel
      v-else-if="errorMessage"
      tone="warning"
      eyebrow="统一搜索"
      title="搜索结果加载失败"
      :description="errorMessage"
      compact
    />

    <section v-else class="search-results-grid">
      <article
        v-for="group in visibleGroups"
        :key="group.key"
        class="card search-group-card"
      >
        <header class="search-group-head">
          <div>
            <p class="hero-kicker">{{ group.title }}</p>
            <h2 class="section-title">{{ describeSearchGroupCount(group) }}</h2>
          </div>
          <button
            v-if="group.preview && activeTab === 'all'"
            type="button"
            class="finance-mini-btn finance-mini-btn-soft"
            @click="activateTab(group.key)"
          >
            查看该分组全部结果
          </button>
        </header>

        <div v-if="group.items.length" class="search-card-list">
          <article
            v-for="item in group.items"
            :key="`${group.key}-${item.id}`"
            :data-focus-key="`${group.key}:${item.id}`"
            class="finance-list-card finance-list-card-interactive search-result-card"
            :class="{ 'search-result-card-highlight': focusKey === `${group.key}:${item.id}` }"
            @click="openSearchItem(group.key, item)"
          >
            <div class="search-result-head">
              <h3>{{ item.title }}</h3>
              <span>{{ item.meta }}</span>
            </div>
            <p>{{ item.summary }}</p>
          </article>
        </div>
        <p v-else class="search-group-empty">{{ group.emptyText }}</p>
      </article>
    </section>
  </section>
</template>

<script setup>
import { computed, nextTick, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { searchGlobal, searchGlobalPublic } from "../api/search";
import { createForecastRun } from "../api/forecast";
import StatePanel from "../components/StatePanel.vue";
import { useClientAuth } from "../lib/client-auth";
import {
  buildGlobalSearchGroups,
  buildSearchPreviewGroups,
  buildSearchItemRoute,
  dedupeVisibleSearchGroups,
  describeSearchGroupCount,
  getGlobalSearchTotal,
  normalizeGlobalSearchKeyword,
  normalizeGlobalSearchResult,
  pickBestSearchMatch,
  resolveGlobalSearchScopeLabel,
  resolveSearchInitialTab,
  shouldRequestGlobalSearch
} from "../lib/global-search";

const route = useRoute();
const router = useRouter();
const { isLoggedIn } = useClientAuth();
const loading = ref(false);
const errorMessage = ref("");
const searchResult = ref(null);
const activeTab = ref("all");
const focusKey = ref("");
const submitting = ref(false);
let latestRequestID = 0;
let highlightTimer = null;

const keyword = computed(() => normalizeGlobalSearchKeyword(route.query.q || ""));
const isH5SearchSurface = computed(() => String(route.name || "").startsWith("h5-"));
const groups = computed(() => buildGlobalSearchGroups(searchResult.value));
const totalCount = computed(() => getGlobalSearchTotal(searchResult.value));
const scopeLabel = computed(() => resolveGlobalSearchScopeLabel(searchResult.value?.scope));
const bestMatch = computed(() =>
  pickBestSearchMatch(searchResult.value, {
    focusType: route.query.focus_type,
    focusID: route.query.focus_id
  })
);
const tabs = computed(() => {
  const [stocks, strategies, news] = groups.value;
  return [
    { key: "all", label: "全部", count: totalCount.value },
    { key: "stocks", label: "股票", count: stocks?.total || 0 },
    { key: "strategies", label: "期货策略", count: strategies?.total || 0 },
    { key: "news", label: "资讯", count: news?.total || 0 }
  ];
});
const overviewItems = computed(() =>
  tabs.value
    .filter((item) => item.key !== "all")
    .map((item) => ({
      key: item.key,
      label: item.label,
      count: `${item.count} 条`,
      note: item.count > 0 ? "可继续展开查看" : "当前没有命中"
    }))
);
const visibleGroups = computed(() => {
  const dedupedGroups = dedupeVisibleSearchGroups(groups.value, {
    focusType: route.query.focus_type,
    focusID: route.query.focus_id
  });
  if (activeTab.value === "all") {
    return buildSearchPreviewGroups(dedupedGroups, {
      bestMatchGroupKey: bestMatch.value?.group?.key,
      perGroupLimit: 3
    });
  }
  return dedupedGroups.filter((group) => group.key === activeTab.value);
});
const showEmptyPrompt = computed(() => !shouldRequestGlobalSearch(keyword.value));
const searchPageTitle = computed(() => (keyword.value ? `“${keyword.value}” 的搜索结果` : "统一搜索结果"));
const searchPageSubtitle = computed(() =>
  keyword.value ? "先看最佳命中，再按分组查看股票推荐、期货策略和资讯结果。" : "从顶部固定搜索条输入关键词后，这里承接完整结果。"
);
const resultSummaryLabel = computed(() => {
  if (bestMatch.value?.group?.title) {
    return `${bestMatch.value.group.title}优先，共 ${totalCount.value} 条`;
  }
  return `共 ${totalCount.value} 条`;
});

watch(
  () => route.query.q,
  () => {
    void loadFullSearch();
  },
  { immediate: true }
);

watch(
  () => [route.query.q, route.query.focus_type, route.query.focus_id, searchResult.value],
  () => {
    activeTab.value = resolveSearchInitialTab(searchResult.value, {
      focusType: route.query.focus_type,
      focusID: route.query.focus_id
    });
  }
);

watch(
  () => [route.query.focus_type, route.query.focus_id, searchResult.value],
  () => {
    void focusSearchItemFromRoute();
  }
);

async function loadFullSearch() {
  window.clearTimeout(highlightTimer);
  focusKey.value = "";
  if (!shouldRequestGlobalSearch(keyword.value)) {
    searchResult.value = null;
    errorMessage.value = "";
    loading.value = false;
    return;
  }
  const requestID = latestRequestID + 1;
  latestRequestID = requestID;
  loading.value = true;
  errorMessage.value = "";
  const searchAction = isLoggedIn.value ? searchGlobal : searchGlobalPublic;
  try {
    const result = await searchAction({ keyword: keyword.value, mode: "full", limit: 20 });
    if (requestID !== latestRequestID) {
      return;
    }
    searchResult.value = normalizeGlobalSearchResult(result, keyword.value);
  } catch (error) {
    if (requestID !== latestRequestID) {
      return;
    }
    searchResult.value = null;
    errorMessage.value = error?.message || "搜索失败，请稍后再试";
  } finally {
    if (requestID === latestRequestID) {
      loading.value = false;
    }
  }
}

async function focusSearchItemFromRoute() {
  const focusType = String(route.query.focus_type || "").trim();
  const focusID = String(route.query.focus_id || "").trim();
  if (!focusType || !focusID || !searchResult.value) {
    focusKey.value = "";
    return;
  }
  const nextFocusKey = `${focusType}:${focusID}`;
  const matchedGroup = groups.value.find((group) => group.key === focusType && group.items.some((item) => item.id === focusID));
  if (!matchedGroup) {
    focusKey.value = "";
    return;
  }
  activeTab.value = focusType;
  focusKey.value = nextFocusKey;
  await nextTick();
  const row = document.querySelector(`[data-focus-key="${nextFocusKey}"]`);
  row?.scrollIntoView?.({ behavior: "smooth", block: "center" });
  highlightTimer = window.setTimeout(() => {
    focusKey.value = "";
  }, 1800);
}

function openSearchItem(groupKey, item) {
  router.push(buildSearchItemRoute(groupKey, item));
}

function activateTab(tabKey) {
  activeTab.value = tabKey;
}

function isForecastable(groupKey) {
  return groupKey === "stocks" || groupKey === "strategies";
}

async function requestForecast(groupKey, item) {
  if (!isLoggedIn.value) {
    router.push("/login?redirect=" + encodeURIComponent(route.fullPath));
    return;
  }
  if (submitting.value) return;
  submitting.value = true;
  errorMessage.value = "";

  try {
    const targetType = groupKey === "stocks" ? "stock" : "futures_strategy";
    const data = await createForecastRun({
      target_type: targetType,
      target_id: item.id,
      target_key: item.id,
      target_label: item.title,
      trigger_type: "USER_REQUEST"
    });
    const runId = data?.id || data?.run?.id;
    if (runId) {
      router.push("/forecast/runs/" + runId);
    }
  } catch (error) {
    errorMessage.value = error?.message || "深度推演请求失败，请重试";
  } finally {
    submitting.value = false;
  }
}
</script>

<style scoped>
.search-page {
  display: grid;
  gap: 16px;
}

.search-hero,
.search-tabs,
.search-overview,
.search-group-card {
  border-radius: 22px;
}

.search-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
}

.search-best-match {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 18px;
  border-radius: 22px;
  border: 1px solid rgba(37, 99, 235, 0.14);
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.06), rgba(255, 255, 255, 0.94));
}

.search-best-match-copy {
  display: grid;
  gap: 10px;
}

.search-best-match-copy h2,
.search-best-match-copy p {
  margin: 0;
}

.search-best-match-copy h2 {
  font-size: 24px;
}

.search-best-match-copy p {
  color: var(--color-text-sub);
}

.search-best-match-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.search-best-match-reason {
  font-size: 13px;
  color: var(--color-brand-primary) !important;
}

.search-best-match-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.search-hero-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.search-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  padding: 12px;
}

.search-overview {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 12px;
  padding: 14px;
}

.search-overview-item {
  display: grid;
  gap: 6px;
  padding: 12px 14px;
  border-radius: 18px;
  border: 1px solid rgba(15, 23, 42, 0.08);
  background: linear-gradient(180deg, rgba(248, 251, 255, 0.98), rgba(255, 255, 255, 0.94));
}

.search-overview-item span,
.search-overview-item p {
  margin: 0;
}

.search-overview-item span {
  color: var(--color-text-soft);
  font-size: 12px;
}

.search-overview-item strong {
  font-size: 18px;
  color: var(--color-text-main);
}

.search-overview-item p {
  color: var(--color-text-sub);
  font-size: 12px;
}

.search-tab {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  min-height: 42px;
  padding: 0 14px;
  border: 1px solid var(--color-border-soft-subtle);
  border-radius: 999px;
  background: #fff;
  color: var(--color-text-main);
}

.search-tab.active {
  border-color: var(--color-brand-primary);
  background: rgba(37, 99, 235, 0.08);
  color: var(--color-brand-primary);
}

.search-results-grid {
  display: grid;
  gap: 16px;
}

.search-group-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 14px;
}

.search-card-list {
  display: grid;
  gap: 12px;
}

.search-result-card {
  transition: box-shadow 0.2s ease, border-color 0.2s ease, transform 0.2s ease;
}

.search-result-card:hover {
  transform: translateY(-1px);
}

.search-result-card-highlight {
  border-color: rgba(37, 99, 235, 0.36);
  box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.12);
}

.search-result-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.search-result-head h3 {
  margin: 0;
  font-size: 16px;
}

.search-result-head span {
  color: var(--color-text-soft);
  font-size: 12px;
  white-space: nowrap;
}

.search-group-empty {
  margin: 0;
  color: var(--color-text-soft);
}

.search-page-h5 {
  gap: 14px;
}

.search-page-h5 .search-hero,
.search-page-h5 .search-tabs,
.search-page-h5 .search-group-card,
.search-page-h5 .search-best-match {
  border-radius: 18px;
}

.search-page-h5 .search-hero {
  padding: 18px 16px;
  border: 1px solid rgba(24, 56, 108, 0.08);
  background:
    radial-gradient(circle at top right, rgba(202, 162, 74, 0.16), transparent 26%),
    linear-gradient(180deg, rgba(248, 251, 255, 0.98), rgba(239, 245, 254, 0.96));
}

.search-page-h5 .search-tabs {
  padding: 10px;
  border: 1px solid rgba(24, 56, 108, 0.08);
  background: rgba(255, 255, 255, 0.94);
  box-shadow: 0 18px 38px rgba(15, 35, 73, 0.08);
}

.search-page-h5 .search-overview {
  gap: 10px;
  padding: 12px;
  border: 1px solid rgba(24, 56, 108, 0.08);
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 16px 32px rgba(15, 35, 73, 0.06);
}

.search-page-h5 .search-overview-item {
  gap: 4px;
  padding: 10px 12px;
  border-radius: 16px;
}

.search-page-h5 .search-overview-item strong {
  font-size: 16px;
}

.search-page-h5 .search-tab {
  min-height: 40px;
  font-size: 13px;
  font-weight: 700;
}

.search-page-h5 .search-best-match {
  padding: 18px 16px;
  border-color: rgba(24, 56, 108, 0.1);
  background:
    radial-gradient(circle at top right, rgba(202, 162, 74, 0.12), transparent 24%),
    linear-gradient(180deg, rgba(248, 251, 255, 0.98), rgba(239, 245, 254, 0.94));
  box-shadow: 0 18px 40px rgba(12, 31, 64, 0.08);
}

.search-page-h5 .search-best-match-copy h2 {
  font-size: 22px;
  line-height: 1.36;
}

.search-page-h5 .search-group-card {
  padding: 18px 16px;
  border: 1px solid rgba(24, 56, 108, 0.08);
  background: linear-gradient(180deg, rgba(248, 251, 255, 0.98), rgba(239, 245, 254, 0.96));
  box-shadow: 0 18px 38px rgba(15, 35, 73, 0.08);
}

.search-page-h5 .search-result-card {
  border-radius: 16px;
  padding: 12px 13px;
}

.search-page-h5 .search-result-head {
  gap: 8px;
}

.search-page-h5 .search-result-head h3 {
  font-size: 15px;
  line-height: 1.45;
}

.search-page-h5 .search-result-head span {
  font-size: 11px;
}

.search-page-h5 .search-result-card p {
  display: -webkit-box;
  overflow: hidden;
  margin: 8px 0 0;
  color: var(--color-text-sub);
  font-size: 13px;
  line-height: 1.6;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

@media (max-width: 768px) {
  .search-hero,
  .search-best-match,
  .search-result-head {
    flex-direction: column;
  }

  .search-overview {
    grid-template-columns: 1fr;
  }

  .search-tab {
    flex: 1 1 calc(50% - 10px);
    justify-content: space-between;
  }

  .search-best-match-actions {
    width: 100%;
    flex-direction: row;
    flex-wrap: wrap;
  }

  .search-page-h5 .search-hero-meta,
  .search-page-h5 .search-best-match-actions {
    width: 100%;
  }

  .search-page-h5 .search-best-match-actions .finance-mini-btn {
    flex: 1 1 calc(50% - 10px);
    justify-content: center;
  }
}
</style>
