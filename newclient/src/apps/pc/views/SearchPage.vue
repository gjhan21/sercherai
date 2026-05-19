<template>
  <div class="search-page">
    <section class="section fade-in-up">
      <div class="search-header">
        <div class="search-input-row">
          <div class="search-input-wrap">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18"><circle cx="11" cy="11" r="7"/><path d="M16.5 16.5L21 21"/></svg>
            <input type="text" v-model="keyword" placeholder="搜索股票、策略、资讯..." @keyup.enter="doSearch" />
          </div>
          <button class="btn-primary" @click="doSearch">搜索</button>
        </div>
        <p class="search-hint" v-if="totalResults > 0">共 {{ totalResults }} 条结果</p>
        <p class="search-hint" v-else-if="searched">未找到结果</p>
      </div>

      <div class="search-tabs" v-if="groups.length">
        <button v-for="g in groups" :key="g.key" class="search-tab" :class="{ active: activeTab === g.key }" @click="activeTab = g.key">
          {{ g.title }} ({{ g.total }})
        </button>
      </div>

      <div v-if="loading" class="search-status">搜索中...</div>

      <div class="search-results">
        <div v-for="g in filteredGroups" :key="g.key" class="result-group">
          <div class="result-group-header">{{ g.title }} · {{ g.total }} 条</div>
          <div v-for="item in g.items" :key="item.id" class="result-item" @click="goItem(g.key, item)">
            <strong>{{ item.title }}</strong>
            <p>{{ item.summary }}</p>
            <span class="result-meta">{{ item.meta }}</span>
          </div>
          <div v-if="!g.items.length" class="result-empty">{{ g.emptyText || '暂无结果' }}</div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { searchGlobalPublic } from "@/api/search.js";

const route = useRoute();
const router = useRouter();
const keyword = ref("");
const loading = ref(false);
const searched = ref(false);
const result = ref(null);
const activeTab = ref("all");

const totalResults = computed(() => searchGroups.value.reduce((s, g) => s + g.total, 0));

const searchGroups = computed(() => {
  const r = result.value;
  if (!r) return [];
  return [
    { key: 'stocks', title: '股票推荐', total: r.stocks?.total || 0, emptyText: '未找到股票推荐',
      items: (r.stocks?.items || []).map(i => ({ id: i.id || i.symbol, title: `${i.symbol || ''} ${i.name || ''}`.trim(), summary: i.reason_summary || '', meta: `评分 ${i.score || '-'} · ${i.risk_level || ''}` })) },
    { key: 'strategies', title: '期货策略', total: r.strategies?.total || 0, emptyText: '未找到策略',
      items: (r.strategies?.items || []).map(i => ({ id: i.id, title: i.contract || i.name || '', summary: i.reason_summary || '', meta: i.direction || '' })) },
    { key: 'news', title: '资讯研报', total: r.news?.total || 0, emptyText: '未找到资讯',
      items: (r.news?.items || []).map(i => ({ id: i.id, title: i.title || '', summary: (i.summary || '').slice(0,80), meta: i.visibility || '公开' })) }
  ].filter(g => totalResults > 0);
});

const filteredGroups = computed(() => {
  if (activeTab.value === 'all') return searchGroups.value;
  return searchGroups.value.filter(g => g.key === activeTab.value);
});

onMounted(() => {
  if (route.query.q) { keyword.value = route.query.q; doSearch(); }
});

async function doSearch() {
  const q = keyword.value.trim();
  if (q.length < 2) return;
  loading.value = true; searched.value = true; activeTab.value = 'all';
  try {
    const r = await searchGlobalPublic({ keyword: q, mode: 'full', limit: 10 });
    result.value = r;
  } catch { result.value = null; }
  finally { loading.value = false; }
}

function goItem(key, item) {
  if (key === 'stocks' || key === 'strategies') router.push({ path: '/identify', query: { q: item.title.split(' ')[0] || item.title } });
  else router.push('/news/' + item.id);
}
</script>

<style scoped>
.search-page { max-width: 1000px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.search-header { margin-bottom: 16px; }
.search-input-row { display: flex; gap: 10px; margin-bottom: 8px; }
.search-input-wrap { flex: 1; display: flex; align-items: center; gap: 8px; padding: 10px 14px; border-radius: var(--radius-md); background: rgba(255,255,255,.04); border: 1px solid var(--border); }
.search-input-wrap:focus-within { border-color: var(--accent-gold); }
.search-input-wrap svg { flex-shrink: 0; color: var(--text-muted); }
.search-input-wrap input { flex: 1; background: none; color: var(--text-primary); font-size: 14px; }
.search-hint { font-size: 13px; color: var(--text-muted); }
.btn-primary { padding: 10px 24px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 14px; font-weight: 600; border: none; cursor: pointer; }
.search-tabs { display: flex; gap: 6px; margin-bottom: 16px; flex-wrap: wrap; }
.search-tab { padding: 6px 14px; border-radius: var(--radius-full); font-size: 13px; color: var(--text-secondary); }
.search-tab.active { color: #000; background: var(--accent-gold); font-weight: 600; }
.search-status { text-align: center; padding: 20px; color: var(--text-muted); font-size: 13px; }
.search-results { display: grid; gap: 16px; }
.result-group-header { font-size: 13px; color: var(--text-muted); margin-bottom: 8px; }
.result-item { padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border); cursor: pointer; }
.result-item:hover { background: rgba(255,255,255,.03); }
.result-item strong { display: block; font-size: 14px; margin-bottom: 4px; }
.result-item p { font-size: 13px; color: var(--text-secondary); line-height: 1.5; margin-bottom: 4px; }
.result-meta { font-size: 11px; color: var(--text-muted); }
.result-empty { font-size: 13px; color: var(--text-muted); padding: 8px 0; }
</style>
