<template>
  <section class="search-page fade-up">
    <header class="search-hero card">
      <div>
        <div class="finance-pill-row">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">搜索页</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">统一结果承接</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">股票 / 期货 / 资讯</span>
        </div>
        <p class="hero-kicker">统一搜索结果</p>
        <h1>{{ searchPageTitle }}</h1>
        <p class="section-subtitle">{{ searchPageSubtitle }}</p>
      </div>
      <div class="search-hero-meta">
        <span class="finance-pill finance-pill-roomy finance-pill-neutral">{{ scopeLabel }}</span>
        <span class="finance-pill finance-pill-roomy finance-pill-info">共 {{ totalCount }} 条</span>
      </div>
    </header>

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
            <h2 class="section-title">{{ group.total }} 条结果</h2>
          </div>
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
import StatePanel from "../components/StatePanel.vue";
import { useClientAuth } from "../lib/client-auth";
import {
  buildGlobalSearchGroups,
  buildSearchItemRoute,
  getGlobalSearchTotal,
  normalizeGlobalSearchKeyword,
  normalizeGlobalSearchResult,
  resolveGlobalSearchScopeLabel,
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
let latestRequestID = 0;
let highlightTimer = null;

const keyword = computed(() => normalizeGlobalSearchKeyword(route.query.q || ""));
const groups = computed(() => buildGlobalSearchGroups(searchResult.value));
const totalCount = computed(() => getGlobalSearchTotal(searchResult.value));
const scopeLabel = computed(() => resolveGlobalSearchScopeLabel(searchResult.value?.scope));
const tabs = computed(() => {
  const [stocks, strategies, news] = groups.value;
  return [
    { key: "all", label: "全部", count: totalCount.value },
    { key: "stocks", label: "股票", count: stocks?.total || 0 },
    { key: "strategies", label: "期货策略", count: strategies?.total || 0 },
    { key: "news", label: "资讯", count: news?.total || 0 }
  ];
});
const visibleGroups = computed(() => {
  if (activeTab.value === "all") {
    return groups.value;
  }
  return groups.value.filter((group) => group.key === activeTab.value);
});
const showEmptyPrompt = computed(() => !shouldRequestGlobalSearch(keyword.value));
const searchPageTitle = computed(() => (keyword.value ? `“${keyword.value}” 的搜索结果` : "统一搜索结果"));
const searchPageSubtitle = computed(() =>
  keyword.value ? "按分组查看股票推荐、期货策略和资讯结果。" : "从顶部固定搜索条输入关键词后，这里承接完整结果。"
);

watch(
  () => route.query.q,
  () => {
    void loadFullSearch();
  },
  { immediate: true }
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
</script>

<style scoped>
.search-page {
  display: grid;
  gap: 16px;
}

.search-hero,
.search-tabs,
.search-group-card {
  border-radius: 22px;
}

.search-hero {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
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

@media (max-width: 768px) {
  .search-hero,
  .search-result-head {
    flex-direction: column;
  }

  .search-tab {
    flex: 1 1 calc(50% - 10px);
    justify-content: space-between;
  }
}
</style>
