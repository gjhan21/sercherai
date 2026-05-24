<template>
  <div class="history-page">
    <section class="section fade-in-up">
      <div class="section-header">
        <div>
          <h2 class="section-title">历史推荐表现</h2>
          <p class="section-subtitle" v-if="showLoginGate">登录后查看历史推荐表现</p>
          <p class="section-subtitle" v-else-if="loading">数据加载中...</p>
          <p class="section-subtitle" v-else>基于真实历史行情追踪 AI 推荐标的的区间表现</p>
        </div>
        <div class="filter-tabs" v-if="!showLoginGate">
          <button v-for="f in filters" :key="f.key" class="filter-tab" :class="{active: activeFilter === f.key}" @click="activeFilter = f.key">{{ f.label }}</button>
        </div>
      </div>

      <div v-if="showLoginGate" class="history-login-gate glass">
        <strong>登录后查看历史推荐表现</strong>
        <p>系统将展示真实的建仓价、最新/结算价、区间收益、最大回撤与推荐结果，不再使用演示数据。</p>
        <button class="hero-cta primary" @click="$router.push('/login')">去登录</button>
      </div>

      <div v-else class="history-stats-row">
        <div class="history-stat glass">
          <span class="hstat-label">总推荐次数</span>
          <span class="hstat-value">{{ summary.total_count || 0 }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">成功次数</span>
          <span class="hstat-value positive">{{ summary.success_count || 0 }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">胜率</span>
          <span class="hstat-value positive">{{ formatPercent(summary.win_rate) }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">平均收益</span>
          <span class="hstat-value" :class="percentClass(summary.avg_return_pct)">{{ signedPercent(summary.avg_return_pct) }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">最高收益</span>
          <span class="hstat-value positive">{{ signedPercent(summary.max_return_pct) }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">最大回撤</span>
          <span class="hstat-value negative">{{ signedPercent(summary.max_drawdown_pct) }}</span>
        </div>
      </div>
    </section>

    <section class="section fade-in-up fade-in-up-delay-1" v-if="!showLoginGate">
      <div class="section-header">
        <h2 class="section-title">详细记录</h2>
        <p class="section-subtitle">展示推荐生命周期内的真实价格区间表现与当前状态</p>
      </div>

      <div v-if="!loading && filteredHistory.length === 0" class="history-empty glass">
        <strong>暂无历史推荐记录</strong>
        <p>当推荐进入历史追踪范围后，这里会显示真实收益与回撤表现。</p>
      </div>

      <div v-else class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>推荐日期</th>
              <th>股票</th>
              <th>名称</th>
              <th>建仓价</th>
              <th>最新/结算价</th>
              <th>真实收益</th>
              <th>最大回撤</th>
              <th>推荐评分</th>
              <th>策略版本</th>
              <th>结果</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="item in filteredHistory"
              :key="item.id"
              class="history-row"
              :class="item.outcome"
              @click="$router.push('/recommendations?symbol=' + item.symbol + '&name=' + item.name)"
              style="cursor:pointer"
            >
              <td>{{ item.recDate }}</td>
              <td><span class="mono">{{ item.symbol }}</span></td>
              <td>{{ item.name }}</td>
              <td>{{ formatPrice(item.entry_price) }}</td>
              <td>{{ formatPrice(item.latest_price) }}</td>
              <td><span :class="percentClass(item.return_pct)">{{ signedPercent(item.return_pct) }}</span></td>
              <td><span class="negative">{{ signedPercent(item.max_drawdown_pct) }}</span></td>
              <td><span class="score-badge" :class="scoreBadgeClass(item.score)">{{ item.score?.toFixed?.(1) ?? item.score }}</span></td>
              <td class="mono">{{ item.strategy_version || item.source_type || 'AI策略' }}</td>
              <td><span class="outcome-tag" :class="item.outcome">{{ outcomeLabel(item.outcome) }}</span></td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { listStockRecommendationHistory } from "@/api/market.js";
import { useClientAuth } from "@/shared/auth/client-auth";

const { isLoggedIn } = useClientAuth();

const activeFilter = ref("all");
const loading = ref(false);
const historyData = ref([]);
const summary = ref({});
const showLoginGate = computed(() => !isLoggedIn.value);

const filters = [
  { key: "all", label: "全部" },
  { key: "success", label: "成功" },
  { key: "fail", label: "失败" },
  { key: "neutral", label: "中性" },
  { key: "ongoing", label: "进行中" },
];

const filteredHistory = computed(() => {
  if (activeFilter.value === "all") return historyData.value;
  return historyData.value.filter((item) => item.outcome === activeFilter.value);
});

function outcomeLabel(outcome) {
  return outcome === "success" ? "达标" : outcome === "fail" ? "未达标" : outcome === "ongoing" ? "跟踪中" : "中性";
}

function formatPrice(value) {
  if (value === null || value === undefined || value === 0) return "--";
  return Number(value).toFixed(2);
}

function formatPercent(value) {
  const numeric = Number(value || 0);
  return `${numeric.toFixed(2)}%`;
}

function signedPercent(value) {
  const numeric = Number(value || 0);
  return `${numeric > 0 ? "+" : ""}${numeric.toFixed(2)}%`;
}

function percentClass(value) {
  return Number(value || 0) >= 0 ? "positive" : "negative";
}

function scoreBadgeClass(score) {
  const numeric = Number(score || 0);
  if (numeric >= 85) return "high";
  if (numeric >= 75) return "mid";
  return "low";
}

async function loadHistory() {
  if (!isLoggedIn.value) {
    historyData.value = [];
    summary.value = {};
    return;
  }
  loading.value = true;
  try {
    const result = await listStockRecommendationHistory({
      outcome: activeFilter.value === "all" ? "" : activeFilter.value,
      page: 1,
      page_size: 20,
    });
    historyData.value = result?.items || [];
    summary.value = result?.summary || {};
  } catch {
    historyData.value = [];
    summary.value = {};
  } finally {
    loading.value = false;
  }
}

watch(activeFilter, () => {
  loadHistory();
});

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
.positive { color: var(--positive); font-weight: 600; }
.negative { color: var(--negative); font-weight: 600; }
.history-login-gate, .history-empty { display: grid; gap: 10px; padding: 20px; border-radius: var(--radius-md); }
.history-login-gate p, .history-empty p { margin: 0; color: var(--text-secondary); font-size: 13px; }
.hero-cta.primary { width: fit-content; }
.filter-tabs { display: flex; gap: 6px; flex-wrap: wrap; }
.filter-tab { padding: 5px 14px; border-radius: var(--radius-full); font-size: 12px; color: var(--text-secondary); }
.filter-tab:hover { color: var(--text-primary); background: rgba(255,255,255,.04); }
.filter-tab.active { color: #000; background: var(--accent-gold); font-weight: 600; }
.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { text-align: left; padding: 10px 12px; font-size: 11px; color: var(--text-secondary); font-weight: 600; text-transform: uppercase; border-bottom: 1px solid var(--border); }
.data-table td { padding: 10px 12px; border-bottom: 1px solid rgba(255,255,255,.03); }
.history-row:hover { background: rgba(255,255,255,.02); }
.mono { font-family: var(--font-mono); font-size: 12px; }
.score-badge { display: inline-flex; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 700; }
.score-badge.high { background: var(--accent-gold-glow); color: var(--accent-gold); }
.score-badge.mid { background: rgba(59,130,246,.1); color: var(--accent-blue); }
.score-badge.low { background: rgba(139,147,176,.1); color: var(--text-secondary); }
.outcome-tag { font-size: 12px; font-weight: 600; padding: 3px 10px; border-radius: var(--radius-full); }
.outcome-tag.success { background: var(--positive-bg); color: var(--positive); }
.outcome-tag.fail { background: var(--negative-bg); color: var(--negative); }
.outcome-tag.neutral,
.outcome-tag.ongoing { background: rgba(139,147,176,.1); color: var(--text-secondary); }
@media (max-width: 900px) { .history-stats-row { grid-template-columns: repeat(3,1fr); } }
</style>
