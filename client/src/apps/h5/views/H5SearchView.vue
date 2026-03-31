<template>
  <div class="h5-page fade-up">
    <div class="h5-page-topline">
      <span class="h5-page-tagline">统一搜索结果</span>
      <span>共 {{ totalCount }} 条结果</span>
    </div>

    <!-- Hero区域: 搜索概览 -->
    <header class="h5-card h5-card-brand h5-card-hero h5-hero-card">
      <div class="h5-card-body">
        <div class="h5-section-head">
          <p class="h5-eyebrow">{{ scopeLabel }}</p>
          <h1 class="h5-title">{{ searchPageTitle }}</h1>
          <p class="h5-subtitle">{{ searchPageSubtitle }}</p>
        </div>
        <div class="h5-meta-list hero-meta-list">
          <span class="h5-meta-chip">{{ resultSummaryLabel }}</span>
        </div>
      </div>
    </header>

    <!-- 最佳命中 -->
    <H5SectionBlock v-if="bestMatch" eyebrow="最佳命中" title="AI 定位结果" tone="accent">
      <article class="h5-list-card h5-card-accent interactive" @click="openSearchItem(bestMatch.group.key, bestMatch.item)">
        <div class="h5-list-topline">
          <strong>{{ bestMatch.item.title }}</strong>
          <span class="h5-badge gold">{{ bestMatch.group.title }}</span>
        </div>
        <p class="h5-list-note">{{ bestMatch.item.summary }}</p>
        <div class="h5-list-submeta">原因: {{ bestMatch.reason }}</div>
        
        <H5ActionBar style="margin-top: 10px;">
          <button
            v-if="isLoggedIn && isForecastable(bestMatch.group.key)"
            type="button"
            class="h5-btn block"
            @click.stop="requestForecast(bestMatch.group.key, bestMatch.item)"
          >
            {{ submitting ? "提交中..." : "深度推演" }}
          </button>
          <button type="button" class="h5-btn-secondary block">查看详情</button>
        </H5ActionBar>
      </article>
    </H5SectionBlock>

    <!-- TAB 导航 -->
    <div class="h5-tabs">
      <button
        v-for="item in tabs"
        :key="item.key"
        type="button"
        class="h5-tab"
        :class="{ active: activeTab === item.key }"
        @click="activeTab = item.key"
      >
        {{ item.label }}
        <small v-if="item.count > 0">({{ item.count }})</small>
      </button>
    </div>

    <!-- 搜索状态 -->
    <H5EmptyState
      v-if="showEmptyPrompt"
      title="准备开始搜索"
      description="在顶部搜索条输入至少两个字符，即可实时联想并获取结果。"
    />
    <H5EmptyState
      v-else-if="loading"
      title="正在检索..."
      description="正在为您同步股票、期货和资讯相关数据。"
    />
    <H5EmptyState
      v-else-if="errorMessage"
      title="搜索失败"
      :description="errorMessage"
    />

    <!-- 结果列表 -->
    <template v-else>
      <section v-for="group in visibleGroups" :key="group.key" class="h5-section-block">
        <div class="h5-section-head">
          <p class="h5-eyebrow">{{ group.title }}</p>
          <h2 class="h5-title" style="font-size: 20px;">{{ describeSearchGroupCount(group) }}</h2>
        </div>
        
        <div v-if="group.items.length" class="h5-list">
          <article
            v-for="item in group.items"
            :key="`${group.key}-${item.id}`"
            class="h5-list-card interactive"
            :class="{ 'h5-list-card-highlight': focusKey === `${group.key}:${item.id}` }"
            @click="openSearchItem(group.key, item)"
          >
            <div class="h5-list-topline">
              <strong style="font-size: 15px;">{{ item.title }}</strong>
              <span>{{ item.meta }}</span>
            </div>
            <p class="h5-list-note" style="display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden;">
              {{ item.summary }}
            </p>
          </article>
          
          <button
            v-if="group.preview && activeTab === 'all' && group.items.length >= 3"
            type="button"
            class="h5-btn-ghost block"
            style="margin-top: 8px;"
            @click="activateTab(group.key)"
          >
            查看全部 {{ group.total }} 条结果
          </button>
        </div>
        <H5EmptyState v-else :title="group.emptyText" description="暂未找到匹配内容" />
      </section>
    </template>

    <div class="h5-sticky-cta-space"></div>
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import H5ActionBar from "../components/H5ActionBar.vue";
import H5EmptyState from "../components/H5EmptyState.vue";
import H5SectionBlock from "../components/H5SectionBlock.vue";
import { searchGlobal, searchGlobalPublic } from "../../../api/search";
import { createForecastRun } from "../../../api/forecast";
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
} from "../../../lib/global-search";
import { useClientAuth } from "../../../shared/auth/client-auth";

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
const searchPageTitle = computed(() => (keyword.value ? `“${keyword.value}” 的搜索结果` : "统一搜索"));
const searchPageSubtitle = computed(() =>
  keyword.value ? "按照股票推荐、期货策略和资讯结果纵览全局。" : "从顶部搜索框输入后，结果会同步呈现。"
);
const resultSummaryLabel = computed(() => {
  if (bestMatch.value?.group?.title) {
    return `${bestMatch.value.group.title}优先 · 共 ${totalCount.value} 条`;
  }
  return `找到 ${totalCount.value} 条匹配结果`;
});

watch(() => route.query.q, () => loadFullSearch(), { immediate: true });

watch(() => [route.query.q, route.query.focus_type, route.query.focus_id, searchResult.value], () => {
  activeTab.value = resolveSearchInitialTab(searchResult.value, {
    focusType: route.query.focus_type,
    focusID: route.query.focus_id
  });
});

watch(() => [route.query.focus_type, route.query.focus_id, searchResult.value], () => {
  void focusSearchItemFromRoute();
});

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
    if (requestID !== latestRequestID) return;
    searchResult.value = normalizeGlobalSearchResult(result, keyword.value);
  } catch (error) {
    if (requestID !== latestRequestID) return;
    searchResult.value = null;
    errorMessage.value = error?.message || "搜索失败";
  } finally {
    if (requestID === latestRequestID) loading.value = false;
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
  highlightTimer = window.setTimeout(() => { focusKey.value = ""; }, 1800);
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
    router.push({ path: "/auth", query: { redirect: route.fullPath } });
    return;
  }
  if (submitting.value) return;
  submitting.value = true;
  errorMessage.value = "";

  try {
    const targetType = groupKey === "stocks" ? "STOCK" : "FUTURES";
    const data = await createForecastRun({
      target_type: targetType,
      target_id: item.id,
      target_key: item.id,
      target_label: item.title,
      trigger_type: "USER_REQUEST"
    });
    const runId = data?.id || data?.run_id || data?.runID || (data?.run?.id);
    if (runId) {
      router.push("/forecast/runs/" + runId);
    } else {
      errorMessage.value = "推演已发起，同步编号中...";
    }
  } catch (error) {
    errorMessage.value = error?.message || "推演发起失败";
  } finally {
    submitting.value = false;
  }
}

onMounted(() => {
  if (route.query.q) loadFullSearch();
});
</script>

<style scoped>
.h5-list-card-highlight {
  border-color: var(--h5-brand) !important;
  box-shadow: 0 0 0 4px var(--h5-brand-soft) !important;
}

.h5-tab small {
  font-weight: 400;
  opacity: 0.8;
  margin-left: 2px;
}
</style>
