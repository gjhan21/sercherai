<template>
  <div class="finance-shell">
    <div v-if="!isDemoPage" class="finance-topbar fade-up">
      <div class="shell-container finance-topbar-inner">
        <div class="finance-topbar-links">
          <span>真实前端开发中</span>
          <span>中文站风格改造</span>
          <span>PC + H5 同时兼容</span>
        </div>
        <div class="finance-topbar-right">
          <span>{{ activeTab.signal }} · {{ activeTab.short }}</span>
        </div>
      </div>
    </div>

    <header v-if="!isDemoPage" class="finance-header fade-up">
      <div class="shell-container finance-header-inner">
        <RouterLink class="finance-brand" to="/home">
          <span class="finance-brand-mark">S</span>
          <div>
            <p class="finance-brand-title">SercherAI</p>
            <p class="finance-brand-desc">推荐、研报、复盘与账户体系一体化客户端</p>
          </div>
        </RouterLink>

        <div class="finance-header-note">
          当前真实页面正在按 demo 逐页收敛为更符合中文用户习惯的金融站结构，优先改决策链路，再改账户链路。
        </div>

        <div class="finance-header-side">
          <div class="finance-status-chip">
            <span>当前模块</span>
            <strong>{{ activeTab.label }}</strong>
          </div>
        </div>
      </div>
    </header>

    <nav v-if="!isDemoPage" class="finance-nav fade-up">
      <div class="shell-container finance-nav-inner">
        <div class="finance-nav-row">
          <RouterLink
            v-for="item in tabs"
            :key="item.path"
            :to="item.path"
            class="finance-nav-link"
            :class="{ active: isTabActive(item.path) }"
          >
            {{ item.label }}
          </RouterLink>
        </div>
        <div class="finance-nav-tools">
          <div v-if="showGlobalSearchBar" ref="searchBarRef" class="finance-nav-search">
            <form class="global-search-form" @submit.prevent="handleSearchSubmit">
              <div class="global-search-field-wrap">
                <input
                  v-model="searchKeyword"
                  class="global-search-input"
                  type="search"
                  name="global_search"
                  maxlength="40"
                  placeholder="输入股票/期货的代码或名称，体验 AI 智能分析与推演"
                  @focus="handleSearchFocus"
                  @keydown.esc.prevent="handleEscapeSearch"
                />
                <button
                  v-if="searchKeyword"
                  type="button"
                  class="global-search-clear"
                  aria-label="清空搜索"
                  @click="clearGlobalSearch"
                >
                  ×
                </button>
              </div>
              <button class="finance-mini-btn finance-mini-btn-primary" type="submit" :disabled="!canSubmitSearch">
                搜索
              </button>
            </form>
            <div v-if="searchDropdownVisible" class="global-search-dropdown finance-list-card finance-list-card-panel">
              <div class="global-search-dropdown-head">
                <div>
                  <p>全局搜索</p>
                  <strong>{{ activeSearchKeyword || "搜索结果" }}</strong>
                </div>
                <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ searchScopeLabel }}</span>
              </div>

              <p v-if="searchLoading" class="search-tip finance-note-strip finance-note-strip-info">正在检索股票、策略和资讯...</p>
              <p v-else-if="searchError" class="search-tip finance-note-strip finance-note-strip-warning">{{ searchError }}</p>
              <template v-else>
                <div v-if="hasSuggestionItems" class="global-search-dropdown-groups">
                  <article v-for="group in suggestionGroups" :key="group.key" class="global-search-group">
                    <header class="global-search-group-head">
                      <div>
                        <strong>{{ group.title }}</strong>
                        <span>{{ group.total }} 条</span>
                      </div>
                    </header>
                    <ul v-if="group.items.length" class="global-search-list">
                      <li
                        v-for="item in group.items.slice(0, 3)"
                        :key="`${group.key}-${item.id}`"
                        class="global-search-item"
                        @click="openSuggestedSearchItem(group.key, item)"
                      >
                        <h4>{{ item.title }}</h4>
                        <p>{{ item.summary }}</p>
                        <span>{{ item.meta }}</span>
                      </li>
                    </ul>
                    <p v-else class="global-search-empty">{{ group.emptyText }}</p>
                  </article>
                </div>
                <p v-else class="global-search-empty global-search-empty-panel">未找到匹配结果。</p>
              </template>

              <div class="global-search-dropdown-actions">
                <button type="button" class="finance-mini-btn finance-mini-btn-soft" @click="openSearchResultsPage()" style="width:100%; justify-content:center;">
                  🚀 查看 "{{ activeSearchKeyword }}" 的探索结果与 AI 深度推演
                </button>
              </div>
            </div>
          </div>
          <div class="finance-nav-account">
            <template v-if="isLoggedIn">
              <div class="finance-account-chip finance-account-chip-nav">
                <span>当前用户</span>
                <strong>{{ accountLabel }}</strong>
              </div>
              <button class="finance-logout-btn" type="button" :disabled="loggingOut" @click="handleLogout">
                {{ loggingOut ? "退出中..." : "退出" }}
              </button>
            </template>
            <RouterLink
              v-else
              class="finance-login-btn"
              :to="{ path: '/auth', query: { redirect: route.fullPath } }"
            >
              登录 / 注册
            </RouterLink>
          </div>
        </div>
      </div>
      <div class="shell-container finance-nav-meta">
        <div class="finance-nav-note">{{ activeTab.focus }}</div>
        <p v-if="showGlobalSearchBar" class="global-search-hint">
          {{ searchKeyword ? "输入时实时联想，回车或点搜索查看完整结果" : "先用顶部搜索条定位股票、期货策略和资讯" }}
        </p>
      </div>
    </nav>

    <section class="finance-route-strip fade-up">
      <div class="shell-container finance-route-strip-inner">
        <div class="finance-route-copy">
          <p>{{ activeTab.label }} · 当前页</p>
          <strong>{{ activeTab.desc }}</strong>
        </div>
        <div v-if="activeHighlights.length" class="finance-route-tags">
          <span v-for="item in activeHighlights" :key="item">{{ item }}</span>
        </div>
      </div>
    </section>

    <main class="finance-view">
      <div class="shell-container">
        <RouterView />
      </div>
    </main>

    <footer v-if="!isDemoPage" class="finance-footer">
      <div class="shell-container finance-footer-inner">
        <div class="finance-footer-main">
          <div class="finance-footer-brand">
            <p class="finance-footer-title">SercherAI Client</p>
            <p class="finance-footer-desc">
              当前客户端按中文金融站风格持续收敛，优先承接推荐、研报、档案、关注与账户链路，不盲目扩成行情门户。
            </p>
          </div>

          <div class="finance-footer-columns">
            <section class="finance-footer-col">
              <p class="finance-footer-heading">页面结构</p>
              <div class="finance-footer-links">
                <RouterLink v-for="item in tabs" :key="`footer-${item.path}`" :to="item.path">
                  {{ item.label }}
                </RouterLink>
              </div>
            </section>

            <section class="finance-footer-col">
              <p class="finance-footer-heading">当前支持</p>
              <div class="finance-footer-tags">
                <span v-for="item in footerSupportItems" :key="item">{{ item }}</span>
              </div>
            </section>
          </div>
        </div>

        <div class="finance-footer-meta">
          <p>适配说明：同一套页面同时兼容 PC 与 H5，PC 以信息密度优先，H5 以阅读顺序压缩。</p>
          <p>风险提示：页面内容仅作研究展示与产品演示说明，不构成任何投资建议。</p>
        </div>
      </div>
    </footer>

    <nav v-if="!isDemoPage" class="finance-mobile-nav">
      <RouterLink
        v-for="item in mobileTabs"
        :key="item.path"
        :to="item.path"
        class="finance-mobile-link"
        :class="{ active: isTabActive(item.path) }"
      >
        <span class="finance-mobile-dot" />
        <span>{{ item.label }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { logout } from "../api/auth";
import { searchGlobal, searchGlobalPublic } from "../api/search";
import { clearClientAuthSession, useClientAuth } from "../lib/client-auth";
import {
  buildGlobalSearchGroups,
  buildSearchPageQuery,
  normalizeGlobalSearchKeyword,
  normalizeGlobalSearchResult,
  resolveGlobalSearchScopeLabel,
  shouldRequestGlobalSearch
} from "../lib/global-search";

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

import { platformTabs as tabs } from "../lib/navigation-tabs";

const activeTab = computed(() => tabs.find((item) => isTabActive(item.path)) || tabs[0]);
const activeHighlights = computed(() => activeTab.value.highlights || []);
const mobileTabs = computed(() => tabs);
const footerSupportItems = [
  "股票推荐",
  "期货策略",
  "历史档案",
  "我的关注",
  "研报与资讯",
  "讨论广场",
  "会员与账户"
];

const accountLabel = computed(() => {
  const phone = session.value?.phone || "";
  if (/^\d{11}$/.test(phone)) {
    return `${phone.slice(0, 3)}****${phone.slice(-4)}`;
  }
  const email = String(session.value?.email || "").trim();
  if (email.includes("@")) {
    const parts = email.split("@");
    const local = parts[0] || "";
    const domain = parts[1] || "";
    if (local.length <= 2) {
      return `${local.slice(0, 1)}***@${domain}`;
    }
    return `${local.slice(0, 2)}***@${domain}`;
  }
  return session.value?.userID || "当前用户";
});

const isDemoPage = computed(() => {
  return route.path.startsWith("/demo") || route.path.includes("Demo");
});

const showGlobalSearchBar = computed(() => (route.path === "/home" || route.path === "/search") && !isDemoPage.value);
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
    console.warn("client logout failed:", error?.message || error);
  } finally {
    clearClientAuthSession();
    loggingOut.value = false;
    await router.replace("/auth");
  }
}

function isTabActive(path) {
  return route.path === path || route.path.startsWith(`${path}/`);
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
</script>
