<template>
  <div class="h5-markets">
    <div class="h5-market-search">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><circle cx="11" cy="11" r="7"/><path d="M16.5 16.5L21 21"/></svg>
      <input type="text" v-model="searchQuery" placeholder="搜索股票..." />
    </div>
    <div class="h5-market-tabs">
      <button v-for="tab in tabs" :key="tab.key" class="h5-market-tab" :class="{ active: activeTab === tab.key }" @click="activeTab = tab.key">{{ tab.label }}</button>
    </div>
    <div class="h5-stock-list">
      <div v-for="stock in filtered" :key="stock.symbol" class="h5-stock-item" @click="$router.push('/markets/' + stock.symbol)">
        <div class="h5-stock-left"><span class="h5-stock-symbol">{{ (stock.symbol || '').split('.')[0] }}</span><span class="h5-stock-name">{{ stock.name }}</span></div>
        <div class="h5-stock-price">{{ stock.price }}</div>
        <div class="h5-stock-right">
          <span class="h5-stock-change" :class="stock.change >= 0 ? 'up' : 'down'">{{ stock.change >= 0 ? '+' : '' }}{{ stock.change }}%</span>
          <span class="h5-stock-ai-badge" :class="(stock.score || 0) >= 85 ? 'hot' : ''">AI</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { STOCK_MASTER } from "@/mock/stocks.js";
import { listStockRecommendations } from "@/api/market.js"
import { useClientAuth } from "@/shared/auth/client-auth";
const { isLoggedIn } = useClientAuth();

const activeTab = ref("all");
const searchQuery = ref("");
const stockList = ref(STOCK_MASTER);

const tabs = [
  { key: "all", label: "全部" },
  { key: "hot", label: "热门" },
  { key: "gainers", label: "涨幅" },
  { key: "ai", label: "AI精选" }
];

const filtered = computed(() => {
  let list = [...stockList.value];
  if (activeTab.value === "hot") list.sort((a, b) => Math.abs(b.change) - Math.abs(a.change));
  else if (activeTab.value === "gainers") list.sort((a, b) => b.change - a.change);
  else if (activeTab.value === "ai") list = list.filter(s => (s.score || 0) >= 80);
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase();
    list = list.filter(s => (s.symbol||'').toLowerCase().includes(q) || (s.name||'').toLowerCase().includes(q));
  }
  return list.slice(0, 20);
});

async function loadStocks() {
  if (!isLoggedIn.value) return;
  try {
    const result = await listStockRecommendations({ page: 1, page_size: 20 });
    if (result?.items?.length) {
      stockList.value = result.items.map(item => ({
        symbol: item.symbol, name: item.name,
        price: String(item.score || 80),
        change: parseFloat((Math.random() * 6 - 1).toFixed(2)),
        score: item.score || 75
      }));
    }
  } catch { /* use mock */ }
}

onMounted(loadStocks);
</script>

<style scoped>
.h5-markets { display: grid; gap: 10px; }
.h5-market-search { display: flex; align-items: center; gap: 6px; padding: 8px 12px; border-radius: var(--radius-full); background: rgba(255,255,255,.04); border: 1px solid var(--border); }
.h5-market-search input { flex: 1; background: none; color: var(--text-primary); font-size: 13px; }
.h5-market-search input::placeholder { color: var(--text-muted); }
.h5-market-tabs { display: flex; gap: 6px; }
.h5-market-tab { padding: 5px 12px; border-radius: var(--radius-full); font-size: 12px; color: var(--text-secondary); }
.h5-market-tab.active { color: #000; background: var(--accent-gold); font-weight: 600; }
.h5-stock-list { display: grid; gap: 6px; }
.h5-stock-item { display: flex; align-items: center; gap: 10px; padding: 12px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-sm); }
.h5-stock-item:active { background: rgba(255,255,255,.03); }
.h5-stock-left { flex: 1; }
.h5-stock-symbol { display: block; font-weight: 700; font-size: 13px; font-family: var(--font-mono); }
.h5-stock-name { font-size: 10px; color: var(--text-secondary); }
.h5-stock-price { font-size: 14px; font-weight: 600; }
.h5-stock-right { text-align: right; }
.h5-stock-change { display: block; font-size: 12px; font-weight: 600; }
.h5-stock-change.up { color: var(--positive); }
.h5-stock-change.down { color: var(--negative); }
.h5-stock-ai-badge { display: none; font-size: 9px; padding: 1px 5px; border-radius: 3px; background: var(--accent-gold-glow); color: var(--accent-gold); font-weight: 700; margin-top: 2px; }
.h5-stock-ai-badge.hot { display: inline-block; }
</style>
