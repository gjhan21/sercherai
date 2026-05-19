<template>
  <div v-if="isBlankLayout" class="blank-root">
    <RouterView />
  </div>

  <div v-else class="pc-root">
    <SidebarNav />

    <div class="pc-main">
      <header class="pc-topbar">
        <div class="topbar-left">
          <h1 class="page-title">{{ currentPageTitle }}</h1>
          <span class="page-date">{{ currentDate }}</span>
        </div>
        <div class="topbar-right">
          <div class="market-clock">
            <span class="clock-dot" :class="marketStatus"></span>
            <span>{{ marketStatusText }}</span>
          </div>
          <div class="search-box" ref="searchRef">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
              <circle cx="11" cy="11" r="7" /><path d="M16.5 16.5L21 21" />
            </svg>
            <input type="text" placeholder="搜索股票/策略/资讯..." v-model="searchQuery" @focus="onSearchFocus" @keydown.escape.prevent="closeSearch" />
            <button v-if="searchQuery" class="search-clear" @click="clearSearch">×</button>

            <!-- Search Dropdown -->
            <div v-if="showDropdown" class="search-dropdown">
              <div v-if="searchLoading" class="sd-status">搜索中...</div>
              <div v-else-if="searchError" class="sd-status sd-error">{{ searchError }}</div>
              <template v-else>
                <div v-for="group in searchGroups" :key="group.key" class="sd-group">
                  <div class="sd-group-title">{{ group.title }} <span class="sd-group-count">{{ group.total }} 条</span></div>
                  <div v-for="item in group.items.slice(0, 3)" :key="item.id" class="sd-item" @click="goToSearchResult(group.key, item)">
                    <strong>{{ item.title }}</strong>
                    <p>{{ item.summary }}</p>
                    <span class="sd-meta">{{ item.meta }}</span>
                  </div>
                  <div v-if="!group.items.length" class="sd-empty">{{ group.emptyText }}</div>
                </div>
              </template>
              <div class="sd-footer" @click="handleSearchSubmit">
                查看更多 "{{ searchQuery }}" 的结果 →
              </div>
            </div>
          </div>
        </div>
      </header>

      <main class="pc-content">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import SidebarNav from "./components/SidebarNav.vue";
import { searchGlobal, searchGlobalPublic } from "@/api/search.js";

const route = useRoute();
const router = useRouter();
const searchQuery = ref("");
const searchRef = ref(null);
const searchLoading = ref(false);
const searchError = ref("");
const searchResult = ref(null);
const showDropdown = ref(false);
const searchFocused = ref(false);

let searchTimer = null;

const isBlankLayout = computed(() => route.meta?.layout === "blank");

const sectionTitles = {
  home: '智投首页', recs: '推荐股票', identify: '鉴别分析',
  news: '资讯中心', community: '社区', user: '我的'
};

const currentPageTitle = computed(() => {
  const metaTitle = route.meta?.title;
  if (metaTitle) return metaTitle;
  const section = route.meta?.section;
  return sectionTitles[section] || '智投首页';
});

const currentDate = computed(() => {
  const d = new Date();
  return `${d.getFullYear()}年${d.getMonth() + 1}月${d.getDate()}日 ${['周日','周一','周二','周三','周四','周五','周六'][d.getDay()]}`;
});

const now = new Date();
const hour = now.getHours();
const marketStatus = computed(() => (hour >= 9 && hour < 15) ? "open" : "closed");
const marketStatusText = computed(() => (hour >= 9 && hour < 15) ? "交易中" : "已收盘");

const searchGroups = computed(() => {
  const r = searchResult.value;
  if (!r) return [];
  return [
    {
      key: 'stocks', title: '股票推荐',
      total: r.stocks?.total || 0, emptyText: '未找到股票推荐',
      items: (r.stocks?.items || []).map(item => ({
        id: item.id || item.symbol, title: `${item.symbol || ''} ${item.name || ''}`.trim(),
        summary: item.reason_summary || '',
        meta: item.risk_level ? `风险: ${item.risk_level}` : ''
      }))
    },
    {
      key: 'strategies', title: '期货策略',
      total: r.strategies?.total || 0, emptyText: '未找到期货策略',
      items: (r.strategies?.items || []).map(item => ({
        id: item.id, title: item.contract || item.name || '',
        summary: item.reason_summary || '',
        meta: item.direction || ''
      }))
    },
    {
      key: 'news', title: '资讯研报',
      total: r.news?.total || 0, emptyText: '未找到资讯',
      items: (r.news?.items || []).map(item => ({
        id: item.id, title: item.title || '',
        summary: (item.summary || '').slice(0, 60),
        meta: item.visibility || '公开'
      }))
    }
  ].filter(g => g.total > 0);
});

watch(searchQuery, (val) => {
  const keyword = (val || '').trim();
  clearTimeout(searchTimer);
  if (keyword.length < 2) { searchResult.value = null; showDropdown.value = false; return; }
  searchTimer = setTimeout(() => performSearch(keyword), 300);
});

watch(searchFocused, (val) => {
  if (val && searchResult.value) showDropdown.value = true;
});

onMounted(() => document.addEventListener('click', handleOutsideClick));
onBeforeUnmount(() => document.removeEventListener('click', handleOutsideClick));

function onSearchFocus() { searchFocused.value = true; if (searchResult.value) showDropdown.value = true; }

function closeSearch() { showDropdown.value = false; searchFocused.value = false; }

function clearSearch() { searchQuery.value = ''; searchResult.value = null; showDropdown.value = false; }

function handleOutsideClick(e) {
  if (searchRef.value && !searchRef.value.contains(e.target)) closeSearch();
}

async function performSearch(keyword) {
  searchLoading.value = true;
  searchError.value = "";
  try {
    const fn = searchGlobalPublic;
    const result = await fn({ keyword, mode: 'suggest', limit: 6 });
    searchResult.value = result;
    showDropdown.value = searchFocused.value;
  } catch (e) {
    searchError.value = e?.message || '搜索失败';
    showDropdown.value = true;
  } finally {
    searchLoading.value = false;
  }
}

function handleSearchSubmit() {
  const keyword = searchQuery.value.trim();
  if (!keyword) return;
  closeSearch();
  router.push({ path: '/search', query: { q: keyword } });
}

function goToSearchResult(groupKey, item) {
  closeSearch();
  if (groupKey === 'stocks' || groupKey === 'strategies') {
    router.push({ path: '/identify', query: { q: item.title.split(' ')[0] || item.title } });
  } else {
    router.push('/news/' + item.id);
  }
}

function handleSearch() { handleSearchSubmit(); }
</script>

<style scoped>
.blank-root { min-height: 100vh; background: var(--bg-primary); }
.pc-root { display: flex; height: 100vh; overflow: hidden; }
.pc-main { flex: 1; display: flex; flex-direction: column; overflow: hidden; }
.pc-topbar { display: flex; align-items: center; justify-content: space-between; padding: 16px 28px; border-bottom: 1px solid var(--border); background: var(--bg-secondary); flex-shrink: 0; }
.topbar-left { display: flex; align-items: baseline; gap: 12px; }
.page-title { font-size: 18px; font-weight: 600; }
.page-date { font-size: 12px; color: var(--text-secondary); }
.topbar-right { display: flex; align-items: center; gap: 16px; }
.market-clock { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-secondary); padding: 6px 12px; border-radius: var(--radius-full); background: rgba(255,255,255,.04); }
.clock-dot { width: 6px; height: 6px; border-radius: 50%; }
.clock-dot.open { background: var(--positive); box-shadow: 0 0 8px rgba(0,200,151,.5); }
.clock-dot.closed { background: var(--neutral); }

.search-box { position: relative; display: flex; align-items: center; gap: 8px; padding: 8px 14px; border-radius: var(--radius-md); background: rgba(255,255,255,.04); border: 1px solid var(--border); width: 280px; transition: border-color .2s; }
.search-box:focus-within { border-color: var(--accent-gold); background: rgba(240,185,11,.04); }
.search-icon { color: var(--text-muted); flex-shrink: 0; }
.search-box input { flex: 1; background: none; color: var(--text-primary); font-size: 13px; }
.search-box input::placeholder { color: var(--text-muted); }
.search-clear { color: var(--text-muted); font-size: 16px; line-height: 1; padding: 0 4px; }

.search-dropdown { position: absolute; top: calc(100% + 6px); left: 0; right: 0; background: var(--bg-elevated); border: 1px solid var(--border); border-radius: var(--radius-md); z-index: 100; max-height: 420px; overflow-y: auto; box-shadow: var(--shadow-lg); }
.sd-status { padding: 14px; text-align: center; font-size: 13px; color: var(--text-secondary); }
.sd-error { color: var(--negative); }
.sd-group { border-bottom: 1px solid var(--border); padding: 10px 12px; }
.sd-group-title { font-size: 12px; font-weight: 700; color: var(--text-muted); text-transform: uppercase; letter-spacing: .5px; margin-bottom: 6px; }
.sd-group-count { font-weight: 400; color: var(--text-muted); margin-left: 6px; }
.sd-item { padding: 8px; border-radius: var(--radius-sm); cursor: pointer; }
.sd-item:hover { background: rgba(255,255,255,.04); }
.sd-item strong { display: block; font-size: 13px; line-height: 1.3; }
.sd-item p { font-size: 12px; color: var(--text-secondary); line-height: 1.4; margin: 2px 0; }
.sd-meta { font-size: 11px; color: var(--text-muted); }
.sd-empty { font-size: 12px; color: var(--text-muted); padding: 4px 0; }
.sd-footer { padding: 10px 12px; text-align: center; font-size: 13px; color: var(--accent-gold); cursor: pointer; font-weight: 600; }
.sd-footer:hover { background: rgba(240,185,11,.05); }

.pc-content { flex: 1; overflow-y: auto; padding: 24px 28px; }
</style>
