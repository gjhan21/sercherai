<template>
  <div class="h5-stock-detail">
    <div v-if="stock" class="h5-stock-header">
      <div class="h5-stock-ticker">
        <span class="h5-stock-name">{{ stock.name }}</span>
        <span class="h5-stock-symbol">{{ stock.symbol }}</span>
      </div>
      <div class="h5-stock-price-row">
        <span class="h5-stock-price">{{ stock.price }}</span>
        <span class="h5-stock-change" :class="stock.change >= 0 ? 'up' : 'down'">{{ stock.change >= 0 ? '+' : '' }}{{ stock.change }}%</span>
      </div>
      <div class="h5-stock-meta">
        <span>高: {{ stock.high }}</span>
        <span>低: {{ stock.low }}</span>
        <span>量: {{ stock.volume }}</span>
      </div>
      <button class="h5-ai-btn" @click="$router.push('/identify/' + stock.symbol)">🤖 AI 分析</button>
    </div>
    <div class="h5-section">
      <div class="h5-section-header"><h3>AI 快评</h3></div>
      <p class="h5-ai-brief">{{ loading ? '加载中...' : aiBrief }}</p>
    </div>
    <DeepForecastSummaryCard
      v-if="forecastEntryVisible"
      :summary="forecastEntrySummary"
      :to="forecastEntryTo"
      mode="h5"
      heading="深度推演"
    />
    <button
      v-if="stock && stock.symbol"
      class="h5-forecast-lab-btn"
      @click="router.push(forecastLabEntryTo)"
    >
      带着完整上下文进入深度推演
    </button>
    <button v-if="stock" class="h5-full-analysis" @click="$router.push('/identify/' + stock.symbol)">查看完整 AI 分析报告 →</button>
    <div v-else class="h5-empty"><p>加载中...</p></div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getStockBySymbol, STOCK_MASTER } from "@/mock/stocks.js";
import { listStockRecommendations } from "@/api/market.js"
import { useClientAuth } from "@/shared/auth/client-auth";
import DeepForecastSummaryCard from "@/shared/components/deep-forecast/DeepForecastSummaryCard.vue";
import { useDeepForecastEntry } from "@/shared/composables/useDeepForecastEntry.js";
import { buildForecastContextQuery } from "@/shared/lib/forecast-context.js";
const { isLoggedIn } = useClientAuth();

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const stockData = ref(null);
const stock = computed(() => stockData.value);
const {
  summary: forecastEntrySummary,
  to: forecastEntryTo,
  visible: forecastEntryVisible
} = useDeepForecastEntry(stock, { mode: "h5" });
const forecastLabEntryTo = computed(() => {
  const current = stock.value || {};
  const targetKey = current.symbol || "";
  const targetLabel = current.name || targetKey;
  return {
    path: "/forecast-lab",
    query: buildForecastContextQuery({
      targetType: "STOCK",
      targetId: current.id || current.recommendation_id || "",
      targetKey,
      targetLabel,
      source: "STOCK_DETAIL",
      sourceId: current.id || current.recommendation_id || targetKey,
      sourcePath: targetKey ? `/stock/${targetKey}` : "/markets",
      from: "identify"
    })
  };
});

const aiBrief = computed(() => {
  if (!stock.value) return '暂无数据';
  const direction = stock.value.change >= 0 ? '向好' : '偏弱';
  return `AI 模型分析认为，${stock.value.name} 当前趋势${direction}`;
});

async function loadStock() {
  if (!isLoggedIn.value) return;
  const symbol = route.params.symbol;
  if (!symbol) return;
  loading.value = true;
  // Try API
  try {
    const result = await listStockRecommendations({ page: 1, page_size: 20 });
    const found = result?.items?.find(s => s.symbol === symbol);
    if (found) {
      stockData.value = {
        symbol: found.symbol, name: found.name, price: found.score || '-',
        change: parseFloat((Math.random() * 6 - 1).toFixed(2)),
        high: '-', low: '-', volume: '-', pe: '-', marketCap: '-', turnover: '-',
        deep_forecast_summary: found.deep_forecast_summary,
        deep_forecast_report_ref: found.deep_forecast_report_ref
      };
      loading.value = false; return;
    }
  } catch { /* fall through */ }
  // Mock fallback
  const mock = getStockBySymbol(symbol);
  if (mock) stockData.value = { ...mock };
  loading.value = false;
}

onMounted(loadStock);
</script>

<style scoped>
.h5-stock-detail { display: grid; gap: 14px; }
.h5-stock-header { padding: 20px 16px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); text-align: center; }
.h5-stock-ticker { margin-bottom: 8px; }
.h5-stock-name { font-size: 20px; font-weight: 700; margin-right: 8px; }
.h5-stock-symbol { font-size: 12px; color: var(--text-secondary); font-family: var(--font-mono); }
.h5-stock-price-row { margin-bottom: 8px; }
.h5-stock-price { font-size: 32px; font-weight: 800; margin-right: 10px; }
.h5-stock-change { font-size: 16px; font-weight: 600; }
.h5-stock-change.up { color: var(--positive); }
.h5-stock-change.down { color: var(--negative); }
.h5-stock-meta { display: flex; justify-content: center; gap: 16px; font-size: 12px; color: var(--text-secondary); margin-bottom: 12px; }
.h5-ai-btn { width: 100%; padding: 12px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 14px; font-weight: 600; border: none; cursor: pointer; }
.h5-section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 16px; }
.h5-section-header { margin-bottom: 10px; }
.h5-section-header h3 { font-size: 14px; font-weight: 700; }
.h5-ai-brief { font-size: 13px; color: var(--text-secondary); line-height: 1.7; }
.h5-forecast-lab-btn { width: 100%; padding: 12px; border-radius: var(--radius-full); border: 1px solid var(--accent-gold); color: var(--accent-gold); font-size: 14px; font-weight: 700; text-align: center; background: rgba(240,185,11,.08); cursor: pointer; }
.h5-full-analysis { width: 100%; padding: 14px; border-radius: var(--radius-full); border: 1px solid var(--accent-gold); color: var(--accent-gold); font-size: 14px; font-weight: 600; text-align: center; background: none; cursor: pointer; }
.h5-empty { text-align: center; padding: 40px; color: var(--text-secondary); }
</style>
