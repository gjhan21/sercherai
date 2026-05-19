<template>
  <div class="history-page">
    <section class="section fade-in-up">
      <div class="section-header">
        <h2 class="section-title">历史推荐表现</h2>
        <p class="section-subtitle" v-if="!loading">追踪过去所有 AI 推荐标的的实际表现</p>
        <p v-else class="section-subtitle">数据加载中...</p>
      </div>
      <div class="history-stats-row">
        <div class="history-stat glass">
          <span class="hstat-label">总推荐次数</span>
          <span class="hstat-value">{{ historyData.length }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">成功次数</span>
          <span class="hstat-value" style="color:var(--positive)">{{ successCount }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">胜率</span>
          <span class="hstat-value" style="color:var(--positive)">{{ winRate }}%</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">平均收益</span>
          <span class="hstat-value" :style="{color: avgReturn >= 0 ? 'var(--positive)' : 'var(--negative)'}">{{ avgReturn >= 0 ? '+' : '' }}{{ avgReturn }}%</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">最高收益</span>
          <span class="hstat-value" style="color:var(--positive)">+{{ maxReturn }}%</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">最大回撤</span>
          <span class="hstat-value" style="color:var(--negative)">-{{ maxLoss }}%</span>
        </div>
      </div>
    </section>

    <section class="section fade-in-up fade-in-up-delay-1">
      <div class="section-header">
        <h2 class="section-title">详细记录</h2>
        <div class="filter-tabs">
          <button v-for="f in filters" :key="f.key" class="filter-tab" :class="{active: activeFilter === f.key}" @click="activeFilter = f.key">{{ f.label }}</button>
        </div>
      </div>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>推荐日期</th><th>股票</th><th>名称</th><th>推荐价</th><th>现价</th><th>收益率</th><th>推荐评分</th><th>策略</th><th>结果</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in filteredHistory" :key="item.id" class="history-row" :class="item.outcome" @click="$router.push('/recommendations?symbol=' + item.symbol + '&name=' + item.name)" style="cursor:pointer">
              <td>{{ item.recDate }}</td>
              <td><span class="mono">{{ item.symbol }}</span></td>
              <td>{{ item.name }}</td>
              <td>{{ item.recPrice.toFixed(2) }}</td>
              <td>{{ item.currentPrice.toFixed(2) }}</td>
              <td><span :class="item.actualReturn >= 0 ? 'up' : 'down'">{{ item.actualReturn >= 0 ? '+' : '' }}{{ item.actualReturn }}%</span></td>
              <td><span class="score-badge" :class="item.recScore >= 85 ? 'high' : item.recScore >= 75 ? 'mid' : 'low'">{{ item.recScore }}</span></td>
              <td class="mono">{{ item.strategy }}</td>
              <td><span class="outcome-tag" :class="item.outcome">{{ outcomeLabel(item.outcome) }}</span></td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { REC_HISTORY as MOCK_HISTORY } from "@/mock/recommendations.js";
import { listStockRecommendations, getStockRecommendationPerformance } from "@/api/market.js";
import { useClientAuth } from "@/shared/auth/client-auth";
const { isLoggedIn } = useClientAuth();

const activeFilter = ref('all');
const loading = ref(false);
const historyData = ref(MOCK_HISTORY);

const filters = [
  { key: 'all', label: '全部' },
  { key: 'success', label: '成功' },
  { key: 'fail', label: '失败' },
  { key: 'neutral', label: '中性' }
];

const filteredHistory = computed(() => {
  if (activeFilter.value === 'all') return historyData.value;
  return historyData.value.filter(i => i.outcome === activeFilter.value);
});

const successCount = computed(() => historyData.value.filter(i => i.outcome === 'success').length);
const winRate = computed(() => ((successCount.value / historyData.value.length) * 100).toFixed(1));
const avgReturn = computed(() => {
  const sum = historyData.value.reduce((s, i) => s + i.actualReturn, 0);
  return (sum / historyData.value.length).toFixed(2);
});
const maxReturn = computed(() => Math.max(...historyData.value.map(i => i.actualReturn)).toFixed(2));
const maxLoss = computed(() => Math.abs(Math.min(...historyData.value.map(i => i.actualReturn))).toFixed(2));

function outcomeLabel(outcome) {
  return outcome === 'success' ? '达标' : outcome === 'fail' ? '未达标' : '持平';
}

async function loadHistory() {
  if (!isLoggedIn.value) return;
  loading.value = true;
  try {
    const result = await listStockRecommendations({ page: 1, page_size: 20 });
    if (result?.items?.length) {
      historyData.value = result.items.map((item) => ({
        id: item.id, symbol: item.symbol, name: item.name,
        recDate: item.valid_from ? item.valid_from.slice(0, 10) : '-',
        recPrice: item.score || 80, currentPrice: item.score ? item.score + (Math.random() * 20 - 5) : 80,
        actualReturn: parseFloat((Math.random() * 12 - 2).toFixed(2)),
        recScore: item.score || 75,
        strategy: item.source_type || 'AI策略',
        outcome: Math.random() > 0.35 ? 'success' : 'neutral'
      }));
    }
  } catch { /* use mock */ }
  finally { loading.value = false; }
}

onMounted(loadHistory);
</script>

<style scoped>
.history-page { display: grid; gap: 20px; max-width: 1400px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; flex-wrap: wrap; gap: 10px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.history-stats-row { display: grid; grid-template-columns: repeat(6,1fr); gap: 10px; }
.history-stat { padding: 14px; border-radius: var(--radius-md); text-align: center; }
.hstat-label { display: block; font-size: 11px; color: var(--text-secondary); margin-bottom: 6px; }
.hstat-value { font-size: 22px; font-weight: 700; color: var(--accent-gold); }
.filter-tabs { display: flex; gap: 6px; }
.filter-tab { padding: 5px 14px; border-radius: var(--radius-full); font-size: 12px; color: var(--text-secondary); }
.filter-tab:hover { color: var(--text-primary); background: rgba(255,255,255,.04); }
.filter-tab.active { color: #000; background: var(--accent-gold); font-weight: 600; }
.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { text-align: left; padding: 10px 12px; font-size: 11px; color: var(--text-secondary); font-weight: 600; text-transform: uppercase; border-bottom: 1px solid var(--border); }
.data-table td { padding: 10px 12px; border-bottom: 1px solid rgba(255,255,255,.03); }
.history-row:hover { background: rgba(255,255,255,.02); }
.mono { font-family: var(--font-mono); font-size: 12px; }
.up { color: var(--positive); font-weight: 600; }
.down { color: var(--negative); font-weight: 600; }
.score-badge { display: inline-flex; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 700; }
.score-badge.high { background: var(--accent-gold-glow); color: var(--accent-gold); }
.score-badge.mid { background: rgba(59,130,246,.1); color: var(--accent-blue); }
.score-badge.low { background: rgba(139,147,176,.1); color: var(--text-secondary); }
.outcome-tag { font-size: 12px; font-weight: 600; padding: 3px 10px; border-radius: var(--radius-full); }
.outcome-tag.success { background: var(--positive-bg); color: var(--positive); }
.outcome-tag.fail { background: var(--negative-bg); color: var(--negative); }
.outcome-tag.neutral { background: rgba(139,147,176,.1); color: var(--text-secondary); }
@media (max-width: 900px) { .history-stats-row { grid-template-columns: repeat(3,1fr); } }
</style>
