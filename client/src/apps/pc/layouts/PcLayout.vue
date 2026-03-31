<template>
  <div class="pc-shell">
    <header class="pc-header">
      <div class="pc-container pc-header-inner">
        <RouterLink class="pc-brand" to="/home">
          <span class="pc-brand-mark">S</span>
          <div>
            <strong>SercherAI</strong>
            <span>PC 决策工作台</span>
          </div>
        </RouterLink>

        <nav class="pc-nav">
          <RouterLink
            v-for="item in tabs"
            :key="item.path"
            :to="item.path"
            class="pc-nav-link"
            :class="{ active: isTabActive(item) }"
          >
            {{ item.label }}
          </RouterLink>
        </nav>

        <div class="pc-actions">
          <div v-if="showGlobalSearchBar" ref="searchBarRef" class="pc-search">
            <form class="pc-search-form" @submit.prevent="handleSearchSubmit">
              <div class="pc-search-field">
                <input
                  v-model="searchKeyword"
                  class="pc-search-input"
                  type="search"
                  name="pc_global_search"
                  maxlength="40"
                  placeholder="输入股票/期货的代码或名称，体验 AI 智能分析与推演"
                  @focus="handleSearchFocus"
                  @keydown.esc.prevent="handleEscapeSearch"
                />
                <button
                  v-if="searchKeyword"
                  type="button"
                  class="pc-search-clear"
                  aria-label="清空搜索"
                  @click="clearGlobalSearch"
                >
                  ×
                </button>
              </div>
              <button type="submit" class="pc-search-submit" :disabled="!canSubmitSearch">搜索</button>
            </form>

            <div v-if="searchDropdownVisible" class="pc-search-dropdown">
              <div class="pc-search-dropdown-head">
                <div>
                  <p>全局搜索</p>
                  <strong>{{ activeSearchKeyword || "搜索结果" }}</strong>
                </div>
                <span class="pc-search-scope">{{ searchScopeLabel }}</span>
              </div>

              <p v-if="searchLoading" class="pc-search-tip">正在检索股票、策略和资讯...</p>
              <p v-else-if="searchError" class="pc-search-tip pc-search-tip-error">{{ searchError }}</p>
              <template v-else>
                <div v-if="hasSuggestionItems" class="pc-search-groups">
                  <article v-for="group in suggestionGroups" :key="group.key" class="pc-search-group">
                    <header class="pc-search-group-head">
                      <strong>{{ group.title }}</strong>
                      <span>{{ group.total }} 条</span>
                    </header>
                    <ul v-if="group.items.length" class="pc-search-list">
                      <li
                        v-for="item in group.items.slice(0, 3)"
                        :key="`${group.key}-${item.id}`"
                        class="pc-search-item"
                        @click="openSuggestedSearchItem(group.key, item)"
                      >
                        <h4>{{ item.title }}</h4>
                        <p>{{ item.summary }}</p>
                        <span>{{ item.meta }}</span>
                      </li>
                    </ul>
                    <p v-else class="pc-search-empty">{{ group.emptyText }}</p>
                  </article>
                </div>
                <p v-else class="pc-search-empty pc-search-empty-panel">未找到匹配结果。</p>
              </template>

              <div class="pc-search-actions">
                <button type="button" class="pc-search-more" @click="openSearchResultsPage()">
                  🚀 查看 "{{ activeSearchKeyword }}" 的探索结果与 AI 深度推演
                </button>
              </div>
            </div>
          </div>
          <RouterLink class="pc-primary-btn" :to="{ path: '/auth', query: { redirect: route.fullPath } }">
            登录 / 注册
          </RouterLink>
        </div>
      </div>
    </header>

    <main class="pc-main">
      <div class="pc-container">
        <RouterView />
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { searchGlobal, searchGlobalPublic } from "../../../api/search";
import {
  buildGlobalSearchGroups,
  buildSearchPageQuery,
  normalizeGlobalSearchKeyword,
  normalizeGlobalSearchResult,
  resolveGlobalSearchScopeLabel,
  shouldRequestGlobalSearch
} from "../../../lib/global-search";
import { useClientAuth } from "../../../shared/auth/client-auth";

const route = useRoute();
const router = useRouter();
const { isLoggedIn } = useClientAuth();
const searchBarRef = ref(null);
const searchKeyword = ref(normalizeGlobalSearchKeyword(route.query.q || ""));
const searchLoading = ref(false);
const searchError = ref("");
const searchResult = ref(null);
const searchDropdownRequested = ref(false);
let searchTimer = null;
let latestSearchRequestID = 0;

const tabs = [
  { path: "/home", label: "首页" },
  { path: "/strategies", label: "策略" },
  { path: "/archive", label: "档案" },
  { path: "/community", label: "社区", match: ["/community"] },
  { path: "/news", label: "资讯" },
  { path: "/membership", label: "会员" },
  { path: "/profile", label: "我的", match: ["/profile"] }
];

const showGlobalSearchBar = computed(() => route.path === "/home" || route.path === "/search");
const activeSearchKeyword = computed(() => normalizeGlobalSearchKeyword(searchKeyword.value));
const searchDropdownVisible = computed(() =>
  showGlobalSearchBar.value && searchDropdownRequested.value && shouldRequestGlobalSearch(activeSearchKeyword.value)
);
const suggestionGroups = computed(() => buildGlobalSearchGroups(searchResult.value));
const hasSuggestionItems = computed(() => suggestionGroups.value.some((group) => group.items.length > 0));
const searchScopeLabel = computed(() => resolveGlobalSearchScopeLabel(searchResult.value?.scope));
const canSubmitSearch = computed(() => shouldRequestGlobalSearch(activeSearchKeyword.value));

watch(
  () => route.query.q,
  (value) => {
    const normalized = normalizeGlobalSearchKeyword(value || "");
    if (normalized !== searchKeyword.value) {
      searchKeyword.value = normalized;
    }
    if (!showGlobalSearchBar.value || !shouldRequestGlobalSearch(normalized)) {
      clearSuggestionState();
    }
  },
  { immediate: true }
);

watch(showGlobalSearchBar, (visible) => {
  if (!visible) {
    clearSuggestionState();
  }
});

watch(searchKeyword, (value) => {
  if (!showGlobalSearchBar.value) {
    return;
  }
  const keyword = normalizeGlobalSearchKeyword(value);
  syncSearchKeywordToRoute(keyword);
  window.clearTimeout(searchTimer);
  if (!shouldRequestGlobalSearch(keyword)) {
    clearSuggestionState();
    return;
  }
  searchDropdownRequested.value = true;
  searchTimer = window.setTimeout(() => {
    void loadSuggestedSearch(keyword);
  }, 250);
});

onMounted(() => {
  document.addEventListener("mousedown", handleDocumentPointerDown);
});

onBeforeUnmount(() => {
  document.removeEventListener("mousedown", handleDocumentPointerDown);
  window.clearTimeout(searchTimer);
});

function clearSuggestionState() {
  searchLoading.value = false;
  searchError.value = "";
  searchResult.value = null;
  searchDropdownRequested.value = false;
}

function handleSearchFocus() {
  if (shouldRequestGlobalSearch(activeSearchKeyword.value)) {
    searchDropdownRequested.value = true;
  }
}

function handleEscapeSearch() {
  searchDropdownRequested.value = false;
}

function handleDocumentPointerDown(event) {
  if (!searchBarRef.value?.contains?.(event.target)) {
    searchDropdownRequested.value = false;
  }
}

function syncSearchKeywordToRoute(keyword) {
  const normalized = normalizeGlobalSearchKeyword(keyword);
  const current = normalizeGlobalSearchKeyword(route.query.q || "");
  if (normalized === current) {
    return;
  }
  const nextQuery = { ...route.query };
  if (normalized) {
    nextQuery.q = normalized;
  } else {
    delete nextQuery.q;
  }
  delete nextQuery.focus_type;
  delete nextQuery.focus_id;
  void router.replace({ path: route.path, query: nextQuery });
}

async function loadSuggestedSearch(keyword) {
  const requestID = latestSearchRequestID + 1;
  latestSearchRequestID = requestID;
  searchLoading.value = true;
  searchError.value = "";
  const searchAction = isLoggedIn.value ? searchGlobal : searchGlobalPublic;
  try {
    const result = await searchAction({ keyword, mode: "suggest", limit: 6 });
    if (requestID !== latestSearchRequestID) {
      return;
    }
    searchResult.value = normalizeGlobalSearchResult(result, keyword);
  } catch (error) {
    if (requestID !== latestSearchRequestID) {
      return;
    }
    searchResult.value = null;
    searchError.value = error?.message || "搜索失败，请稍后再试";
  } finally {
    if (requestID === latestSearchRequestID) {
      searchLoading.value = false;
    }
  }
}

function clearGlobalSearch() {
  searchKeyword.value = "";
  if (route.path === "/search") {
    void router.replace({ path: "/search" });
  }
}

function openSearchResultsPage(options = {}) {
  const query = buildSearchPageQuery(activeSearchKeyword.value, options);
  searchDropdownRequested.value = false;
  void router.push({ path: "/search", query });
}

function openSuggestedSearchItem(groupKey, item) {
  openSearchResultsPage({
    focusType: groupKey,
    focusID: item?.id || ""
  });
}

function handleSearchSubmit() {
  if (!canSubmitSearch.value) {
    return;
  }
  openSearchResultsPage();
}

function isTabActive(item) {
  if (route.path === item.path) {
    return true;
  }
  if (!Array.isArray(item.match)) {
    return false;
  }
  return item.match.some((prefix) => route.path.startsWith(prefix));
}
</script>
