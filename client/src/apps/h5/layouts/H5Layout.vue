<template>
  <div class="h5-shell">
    <header class="h5-header">
      <div class="h5-header-inner">
        <div class="h5-header-top">
          <RouterLink class="h5-brand" to="/home" aria-label="返回首页">
            <span class="h5-brand-mark">S</span>
            <div class="h5-brand-copy">
              <strong>SercherAI</strong>
              <span>移动投研终端</span>
            </div>
          </RouterLink>

          <div class="h5-header-actions">
            <span v-if="isLoggedIn" class="h5-user">{{ accountLabel }}</span>
            <RouterLink v-if="!isLoggedIn" class="h5-header-link" :to="{ name: 'h5-auth', query: { redirect: route.fullPath } }">
              注册 / 登录
            </RouterLink>
            <button v-else type="button" class="h5-header-link" :disabled="loggingOut" @click="handleLogout">
              {{ loggingOut ? '退出中' : '退出' }}
            </button>
          </div>
        </div>

        <div class="h5-header-scene">
          <div class="h5-scene-chip-row">
            <span class="h5-scene-chip">{{ shellScene.section }}</span>
            <span class="h5-shell-pulse">{{ shellScene.pulse }}</span>
          </div>
          <strong>{{ shellScene.title }}</strong>
          <small>{{ shellScene.subtitle }}</small>
        </div>

        <div v-if="showGlobalSearchBar" ref="searchBarRef" class="h5-search-shell">
          <form class="h5-search-form" @submit.prevent="handleSearchSubmit">
            <div class="h5-search-field">
              <input
                v-model="searchKeyword"
                class="h5-search-input"
                type="search"
                name="h5_global_search"
                maxlength="40"
                placeholder="输入股票/期货的代码或名称，体验 AI 智能分析与推演"
                @focus="handleSearchFocus"
                @keydown.esc.prevent="handleEscapeSearch"
              />
              <button
                v-if="searchKeyword"
                type="button"
                class="h5-search-clear"
                aria-label="清空搜索"
                @click="clearGlobalSearch"
              >
                ×
              </button>
            </div>
            <button type="submit" class="h5-search-submit" :disabled="!canSubmitSearch">搜索</button>
          </form>

          <div v-if="searchDropdownVisible" class="h5-search-dropdown">
            <div class="h5-search-dropdown-head">
              <div>
                <p>全局搜索</p>
                <strong>{{ activeSearchKeyword || "搜索结果" }}</strong>
              </div>
              <span class="h5-search-scope">{{ searchScopeLabel }}</span>
            </div>

            <p v-if="searchLoading" class="h5-search-tip">正在检索股票、策略和资讯...</p>
            <p v-else-if="searchError" class="h5-search-tip h5-search-tip-error">{{ searchError }}</p>
            <template v-else>
              <div v-if="hasSuggestionItems" class="h5-search-groups">
                <article v-for="group in suggestionGroups" :key="group.key" class="h5-search-group">
                  <header class="h5-search-group-head">
                    <strong>{{ group.title }}</strong>
                    <span>{{ group.total }} 条</span>
                  </header>
                  <ul v-if="group.items.length" class="h5-search-list">
                    <li
                      v-for="item in group.items.slice(0, 3)"
                      :key="`${group.key}-${item.id}`"
                      class="h5-search-item"
                      @click="openSuggestedSearchItem(group.key, item)"
                    >
                      <h4>{{ item.title }}</h4>
                      <p>{{ item.summary }}</p>
                      <span>{{ item.meta }}</span>
                    </li>
                  </ul>
                  <p v-else class="h5-search-empty">{{ group.emptyText }}</p>
                </article>
              </div>
              <p v-else class="h5-search-empty h5-search-empty-panel">未找到匹配结果。</p>
            </template>

            <div class="h5-search-actions">
              <button type="button" class="h5-search-more" @click="openSearchResultsPage()">
                🚀 查看 "{{ activeSearchKeyword }}" 的探索结果与 AI 深度推演
              </button>
            </div>
          </div>
        </div>
      </div>
    </header>

    <main class="h5-main">
      <RouterView />
    </main>

    <nav class="h5-tabbar" aria-label="H5 底部导航">
      <RouterLink
        v-for="item in tabs"
        :key="item.path"
        :to="item.path"
        class="h5-tabbar-link"
        :class="{ active: isTabActive(item.path) }"
      >
        <span class="h5-tabbar-icon" aria-hidden="true" v-html="tabIcon(item.icon)"></span>
        <span>{{ item.label }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { searchGlobal, searchGlobalPublic } from "../../../api/search";
import { logout } from "../../../shared/api/auth";
import { clearClientAuthSession, useClientAuth } from "../../../shared/auth/client-auth";
import {
  buildGlobalSearchGroups,
  buildSearchPageQuery,
  normalizeGlobalSearchKeyword,
  normalizeGlobalSearchResult,
  resolveGlobalSearchScopeLabel,
  shouldRequestGlobalSearch
} from "../../../lib/global-search";
import { resolveShellScene, resolveTabItems } from "../lib/shell-meta";

const TAB_ICON_MAP = {
  home: `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M4.5 10.5 12 4l7.5 6.5" />
      <path d="M6.5 10.5V20h11v-9.5" />
    </svg>
  `,
  news: `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M5 6.5h14" />
      <path d="M5 11.5h14" />
      <path d="M5 16.5h9" />
      <path d="M17 15.5h2v3h-2z" />
    </svg>
  `,
  insight: `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M5 18h14" />
      <path d="M7 16V9" />
      <path d="M12 16V6" />
      <path d="M17 16v-4" />
    </svg>
  `,
  vip: `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="m5 8 3 3 4-5 4 5 3-3" />
      <path d="M6 10.5 8 18h8l2-7.5" />
    </svg>
  `,
  user: `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="12" cy="8" r="3.2" />
      <path d="M5 19c1.8-3 4.2-4.5 7-4.5S17.2 16 19 19" />
    </svg>
  `
};

const route = useRoute();
const router = useRouter();
const loggingOut = ref(false);
const { session, isLoggedIn } = useClientAuth();
const searchBarRef = ref(null);
const searchKeyword = ref(normalizeGlobalSearchKeyword(route.query.q || ""));
const searchLoading = ref(false);
const searchError = ref("");
const searchResult = ref(null);
const searchDropdownRequested = ref(false);
let searchTimer = null;
let latestSearchRequestID = 0;

const tabs = resolveTabItems();
const shellScene = computed(() => resolveShellScene(route.fullPath));
const showGlobalSearchBar = computed(() => route.path === "/home" || route.path === "/search");
const activeSearchKeyword = computed(() => normalizeGlobalSearchKeyword(searchKeyword.value));
const searchDropdownVisible = computed(() =>
  showGlobalSearchBar.value && searchDropdownRequested.value && shouldRequestGlobalSearch(activeSearchKeyword.value)
);
const suggestionGroups = computed(() => buildGlobalSearchGroups(searchResult.value));
const hasSuggestionItems = computed(() => suggestionGroups.value.some((group) => group.items.length > 0));
const searchScopeLabel = computed(() => resolveGlobalSearchScopeLabel(searchResult.value?.scope));
const canSubmitSearch = computed(() => shouldRequestGlobalSearch(activeSearchKeyword.value));

const accountLabel = computed(() => {
  const phone = session.value?.phone || "";
  if (/^\d{11}$/.test(phone)) {
    return `${phone.slice(0, 3)}****${phone.slice(-2)}`;
  }
  const email = String(session.value?.email || "").trim();
  if (email.includes("@")) {
    return email.split("@")[0] || "当前用户";
  }
  return session.value?.userID || "当前用户";
});

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

function isTabActive(path) {
  if (path === "/home") {
    return route.path === "/home" || route.path === "/";
  }
  return route.path.startsWith(path);
}

function tabIcon(name) {
  return TAB_ICON_MAP[name] || TAB_ICON_MAP.home;
}

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

async function handleLogout() {
  if (loggingOut.value) {
    return;
  }
  loggingOut.value = true;
  try {
    const refreshToken = session.value?.refreshToken || "";
    if (refreshToken) {
      await logout(refreshToken);
    }
  } catch (error) {
    console.warn("h5 logout failed:", error?.message || error);
  } finally {
    clearClientAuthSession();
    loggingOut.value = false;
    await router.replace({ name: "h5-auth" });
  }
}
</script>
