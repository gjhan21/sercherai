<template>
  <div class="daily-recs-page">
    <section class="flow-strip glass fade-in-up">
      <div class="flow-copy">
        <p class="flow-kicker">机会主线</p>
        <h2>第 1 步：发现机会</h2>
        <p>先从今日 AI 精选里确认机会，再进入交易策略与深度推演完成决策闭环。</p>
      </div>
      <div class="flow-steps">
        <span class="flow-step active">1. 每日推荐</span>
        <span class="flow-step">2. 交易策略</span>
        <span class="flow-step">3. 深度推演</span>
      </div>
      <div class="flow-actions">
        <button class="flow-btn" @click="goToStrategies()">查看交易策略</button>
        <button class="flow-btn ghost" @click="goToForecastLab()">查看深度推演</button>
      </div>
    </section>

    <!-- AI Picks -->
    <section class="section fade-in-up">
      <div class="section-header">
        <div>
          <div class="picks-header-row">
            <h2 class="section-title">今日 AI 精选</h2>
            <span class="tag tag-gold">{{ todayDate }}</span>
          </div>
          <p class="section-subtitle">全市场 {{ STOCK_MASTER.length }} 只标的 · AI 多因子模型筛选</p>
        </div>
        <div class="recs-summary-stats">
          <div class="recs-stat"><span class="recs-stat-value" style="color:var(--accent-gold)">{{ avgScore }}</span><span class="recs-stat-label">平均评分</span></div>
          <div class="recs-stat"><span class="recs-stat-value">{{ riskLevelSummary }}</span><span class="recs-stat-label">主风险等级</span></div>
          <div class="recs-stat"><span class="recs-stat-value">{{ dailyRecs.length }}</span><span class="recs-stat-label">推荐数量</span></div>
        </div>
      </div>
      <div v-if="showLoginGate" class="daily-recs-empty glass">
        <div class="daily-recs-empty-title">登录后查看今日 AI 精选</div>
        <p class="daily-recs-empty-copy">系统会基于当日有效推荐集合展示真实的 AI 精选标的、风险等级、建议仓位与止盈止损。</p>
        <button class="flow-btn" @click="$router.push('/login')">去登录</button>
      </div>
      <div v-else-if="dailyRecs.length" class="picks-grid">
        <article v-for="stock in dailyRecs" :key="stock.symbol" class="pick-card glass" :class="{ active: chartSymbol === stock.symbol }" @click="showDetail(stock)">
          <div class="pick-rank" :class="'rank-' + stock.rank">{{ stock.rank }}</div>
          <div class="pick-info">
            <div class="pick-name-row">
              <span class="pick-symbol">{{ stock.symbol?.split('.')[0] }}</span>
              <span class="pick-name">{{ stock.name }}</span>
            </div>
            <div class="pick-meta-row">
              <span class="pick-meta">风险等级：{{ stock.riskLevel || '待评估' }}</span>
              <span class="pick-meta">建议仓位：{{ stock.positionSize || '待确认' }}</span>
            </div>
          </div>
          <div class="pick-score-ring">
            <svg viewBox="0 0 36 36" class="score-svg">
              <circle cx="18" cy="18" r="15.5" fill="none" stroke="rgba(255,255,255,0.06)" stroke-width="2.5" />
              <circle cx="18" cy="18" r="15.5" fill="none" stroke="var(--accent-gold)" stroke-width="2.5" stroke-dasharray="97.4" :stroke-dashoffset="97.4 - 97.4 * stock.score / 100" stroke-linecap="round" />
            </svg>
            <span class="score-num">{{ stock.score }}</span>
          </div>
          <p class="pick-reason"><span class="reason-label">AI 理由：</span>{{ stock.aiReason }}</p>
          <div class="pick-score-badges">
            <span class="pick-badge" :class="stock.score >= 85 ? 'badge-high' : stock.score >= 75 ? 'badge-mid' : 'badge-low'">AI {{ stock.score }}</span>
            <span v-if="stock.riskLevel" class="pick-badge risk-badge">{{ stock.riskLevel }}</span>
          </div>
          <div class="pick-plan-grid">
            <div><span class="plan-label">止盈</span><span class="plan-value plan-positive">{{ stock.takeProfit || '待确认' }}</span></div>
            <div><span class="plan-label">止损</span><span class="plan-value plan-negative">{{ stock.stopLoss || '待确认' }}</span></div>
          </div>
          <div class="pick-footer">
            <button class="pick-detail-btn" @click.stop="$router.push('/identify?q=' + stock.symbol)">分析 →</button>
            <span class="pick-strategy">策略: {{ stock.positionSize || '待确认' }}</span>
          </div>
          <div class="pick-journey">
            <button class="journey-btn" @click.stop="goToStrategies(stock)">查看对应策略</button>
            <button class="journey-btn ghost" @click.stop="goToForecastLab(stock)">进入深度推演</button>
          </div>
        </article>
      </div>
      <div v-else class="daily-recs-empty glass">
        <div class="daily-recs-empty-title">今日暂无可发布推荐</div>
        <p class="daily-recs-empty-copy">当前交易日还没有可展示的 AI 精选推荐，稍后再来查看。</p>
      </div>
    </section>

    <!-- Real K-line Chart -->
    <section class="section fade-in-up fade-in-up-delay-2">
      <div class="section-header">
        <h2 class="section-title">行情图表 — {{ chartStock?.name || '--' }}</h2>
        <button class="predict-btn" @click="loadPrediction" :disabled="predictLoading" v-if="klineData.length && !prediction">
          <svg v-if="!predictLoading" viewBox="0 0 20 20" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16"><circle cx="10" cy="10" r="8"/><path d="M10 4v4l2.5 2.5"/><path d="M6 14l2-2M14 14l-2-2"/></svg>
          <span v-if="!predictLoading">AI 预测未来 7 天</span>
          <span v-else class="predict-loading">
            <span class="predict-spinner"></span>
            AI 分析中...
          </span>
        </button>
      </div>
      <div class="chart-container glass">
        <div class="chart-header">
          <div>
            <span class="chart-stock-name">{{ chartStock?.name }}</span>
            <span class="chart-stock-price">{{ chartStock?.score ? `AI 评分 ${chartStock.score}` : '--' }}</span>
            <span class="chart-stock-change">{{ chartStock?.riskLevel || '待评估' }}</span>
          </div>
          <div class="chart-status">{{ klineLoading ? '加载中...' : (klineData.length + ' 天') }}</div>
        </div>
        <KlineChart :data="klineData" :width="chartWidth" :height="380" :prediction="prediction" v-if="klineData.length" />
        <div v-else class="chart-empty">暂无K线数据</div>
      </div>
      <div v-if="prediction" class="prediction-panel glass">
        <div class="pred-header">
          <span>📈 历史形态预测 — K线图上金色区域为未来 7 天置信区间</span>
          <button class="pred-close" @click="prediction = null">×</button>
        </div>
        <div class="pred-stats">
          <span>平均收益: <strong :style="{color: prediction.avg_return >= 0 ? 'var(--positive)' : 'var(--negative)'}">{{ prediction.avg_return >= 0 ? '+' : '' }}{{ prediction.avg_return }}%</strong></span>
          <span>上涨概率: <strong>{{ prediction.win_rate }}%</strong></span>
          <span>匹配数量: <strong>{{ matches.length }}</strong></span>
        </div>
        <div class="pred-matches" v-if="matches.length">
          <div class="pred-matches-title">Top 匹配</div>
          <div v-for="m in matches.slice(0, 5)" :key="m.rank" class="pred-match-row">
            <span>#{{ m.rank }}</span>
            <span>{{ m.stock }}</span>
            <span>相似 {{ (m.similarity * 100).toFixed(0) }}%</span>
            <span :style="{color: m.return_7d >= 0 ? 'var(--positive)' : 'var(--negative)'}">{{ m.return_7d >= 0 ? '+' : '' }}{{ m.return_7d.toFixed(1) }}%</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Today's Strategy Summary -->
    <section class="section fade-in-up fade-in-up-delay-3">
      <div class="section-header">
        <h2 class="section-title">今日策略要点</h2>
        <p class="section-subtitle">AI 为每只精选标的匹配的交易策略</p>
      </div>
      <div class="strategy-summary-grid">
        <div v-for="stock in dailyRecs" :key="'strat-' + stock.symbol" class="strat-mini-card glass">
          <div class="strat-mini-header">
            <span class="strat-mini-symbol">{{ stock.symbol.split('.')[0] }}</span>
            <span class="strat-mini-name">{{ stock.name }}</span>
          </div>
          <div class="strat-mini-grid">
            <div><span class="strat-label">风险等级</span><span class="strat-val">{{ stock.riskLevel || '待评估' }}</span></div>
            <div><span class="strat-label">止损</span><span class="strat-val" style="color:var(--negative)">{{ stock.stopLoss || '待确认' }}</span></div>
            <div><span class="strat-label">止盈</span><span class="strat-val" style="color:var(--positive)">{{ stock.takeProfit || '待确认' }}</span></div>
            <div><span class="strat-label">仓位</span><span class="strat-val">{{ stock.positionSize || '待确认' }}</span></div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { STOCK_MASTER } from "@/mock/stocks.js";
import { listStockRecommendations, getStockKline, getStockPatternMatch } from "@/api/market.js"
import KlineChart from "@/apps/pc/components/KlineChart.vue"
import { buildForecastContextQuery } from "@/shared/lib/forecast-context.js";
import { useClientAuth } from "@/shared/auth/client-auth";
const { isLoggedIn } = useClientAuth();

const router = useRouter();
const route = useRoute();
const chartSymbol = ref(route.query.symbol || '300750.SZ');
const hoverBar = ref(null);
const dailyRecs = ref([]);
const apiLoaded = ref(false);
const klineData = ref([]);
const klineLoading = ref(false);
const chartWidth = ref(800);
const prediction = ref(null);
const predictLoading = ref(false);
const matches = ref([]);
const todayTradeDate = computed(() => {
  const d = new Date();
  const year = d.getFullYear();
  const month = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
});
const showLoginGate = computed(() => !isLoggedIn.value);

const todayDate = computed(() => {
  const d = new Date();
  return `${d.getFullYear()}年${d.getMonth()+1}月${d.getDate()}日`;
});

const avgScore = computed(() => {
  if (!dailyRecs.value.length) return "--";
  return (dailyRecs.value.reduce((s, i) => s + i.score, 0) / dailyRecs.value.length).toFixed(0);
});
const riskLevelSummary = computed(() => dailyRecs.value[0]?.riskLevel || "--");

const selectedStock = ref(null);
const chartStock = computed(() => {
  const s = dailyRecs.value.find(r => r.symbol === chartSymbol.value);
  if (s) return s;
  // Fallback: show symbol even if not in recommendations
  const sym = chartSymbol.value;
  if (sym) return { symbol: sym, name: route.query.name || sym.replace('.SZ','').replace('.SH',''), score: null, riskLevel: "" };
  return dailyRecs.value[0] || null;
});

const chartLabels = ['05/05','05/07','05/09','05/12','05/14'];

const chartData = computed(() => {
  const base = 198;
  return [
    { open: base - 2, close: base + 3, high: base + 5, low: base - 4, annotation: null },
    { open: base + 3, close: base + 1, high: base + 4, low: base - 1, annotation: { text: 'AI 监测到资金流入', color: 'var(--accent-gold)' } },
    { open: base + 1, close: base + 5, high: base + 7, low: base, annotation: null },
    { open: base + 5, close: base + 8, high: base + 10, low: base + 3, annotation: { text: '突破前高，AI 推荐买入', color: 'var(--positive)' } },
    { open: base + 8, close: base + 6, high: base + 9, low: base + 4, annotation: null }
  ];
});

function getCandleHeight(bar) {
  return Math.abs(bar.close - bar.open) * 0.3 + 30;
}
function getCandleTop(bar) {
  return Math.max(bar.open, bar.close) * 0.3;
}

function showDetail(stock) {
  selectedStock.value = stock;
  chartSymbol.value = stock.symbol;
}

function getRecommendationTargetId(stock) {
  return String(stock?.id || stock?.recommendation_id || stock?.reco_id || "").trim();
}

function goToStrategies(stock) {
  const target = stock?.symbol || chartSymbol.value || "";
  const targetId = getRecommendationTargetId(stock);
  router.push({
    path: "/recommendations/strategies",
    query: target
      ? { reco_id: targetId, symbol: target, name: stock?.name || chartStock.value?.name || "", from: "recommendations" }
      : { from: "recommendations" }
  });
}

function goToForecastLab(stock) {
  const target = stock?.symbol || chartSymbol.value || "";
  const targetLabel = stock?.name || chartStock.value?.name || "";
  const targetId = getRecommendationTargetId(stock);
  router.push({
    path: "/forecast-lab",
    query: buildForecastContextQuery(
      target
        ? {
            targetType: "STOCK",
            targetId: targetId,
            targetKey: target,
            targetLabel,
            source: "RECOMMENDATION",
            sourceId: targetId,
            sourcePath: "/recommendations",
            from: "recommendations"
          }
        : {
            targetType: "STOCK",
            source: "RECOMMENDATION",
            sourcePath: "/recommendations",
            from: "recommendations"
          }
    )
  });
}

async function loadDailyRecs() {
  if (!isLoggedIn.value) return;
  try {
    const result = await listStockRecommendations({ trade_date: todayTradeDate.value, page: 1, page_size: 6 });
    if (result?.items?.length) {
      dailyRecs.value = result.items.map((item, i) => ({
        id: item.id || item.recommendation_id || item.reco_id || "",
        recommendation_id: item.recommendation_id || item.id || "",
        reco_id: item.reco_id || item.id || "",
        symbol: item.symbol,
        name: item.name,
        rank: i + 1,
        score: item.score || 80,
        aiReason: item.reason_summary || 'AI 模型筛选',
        riskLevel: item.risk_level || '',
        positionSize: item.position_range || '待确认',
        takeProfit: item.take_profit || '',
        stopLoss: item.stop_loss || '',
        strategy: {
          stopLoss: item.stop_loss || '待确认',
          takeProfit: item.take_profit || '待确认',
          positionSize: item.position_range || '待确认'
        },
        tags: ['AI精选']
      }));
      apiLoaded.value = true;
      selectedStock.value = dailyRecs.value[0] || null;
      if (dailyRecs.value.length) loadKline(dailyRecs.value[0].symbol);
    }
  } catch { /* use mock */ }
}

onMounted(() => { loadDailyRecs(); loadKline(chartSymbol.value); });
import { watch } from "vue";
watch(chartSymbol, (val) => { if (val) loadKline(val); });
watch(() => route.query.symbol, (val) => { if (val && val !== chartSymbol.value) { chartSymbol.value = val; } });

async function loadKline(symbol) {
  if (!symbol) return;
  klineLoading.value = true;
  try {
    const result = await getStockKline({ symbol, days: 60 });
    if (result?.points) klineData.value = result.points;
  } catch { klineData.value = []; }
  finally { klineLoading.value = false; }
}

async function loadPrediction() {
  const sym = chartSymbol.value;
  if (!sym) return;
  predictLoading.value = true;
  try {
    const result = await getStockPatternMatch({ symbol: sym, lookback: 20, top: 10 });
    if (result) {
      prediction.value = result.prediction;
      matches.value = result.matches || [];
    }
  } catch { /* ignore */ }
  finally { predictLoading.value = false; }
}

</script>

<style scoped>
.daily-recs-page { display: grid; gap: 20px; max-width: 1400px; }

.flow-strip { display: grid; gap: 14px; padding: 18px 20px; border-radius: var(--radius-lg); border: 1px solid var(--border); background: linear-gradient(135deg, rgba(240,185,11,.08), rgba(255,255,255,.02)); }
.flow-copy h2 { font-size: 22px; font-weight: 800; margin: 4px 0 6px; }
.flow-copy p { color: var(--text-secondary); line-height: 1.7; }
.flow-kicker { font-size: 11px; color: var(--accent-gold); text-transform: uppercase; letter-spacing: .08em; }
.flow-steps { display: flex; gap: 8px; flex-wrap: wrap; }
.flow-step { padding: 8px 12px; border-radius: var(--radius-full); border: 1px solid var(--border); color: var(--text-secondary); font-size: 12px; }
.flow-step.active { border-color: var(--border-gold); color: var(--accent-gold); background: var(--accent-gold-glow); }
.flow-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.flow-btn, .journey-btn { display: inline-flex; align-items: center; justify-content: center; padding: 10px 14px; border-radius: var(--radius-full); font-size: 13px; font-weight: 600; cursor: pointer; border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.flow-btn.ghost, .journey-btn.ghost { background: none; border: 1px solid var(--border-gold); color: var(--accent-gold); }

.recs-summary-stats { display: flex; gap: 16px; }
.recs-stat { padding: 12px 20px; border-radius: var(--radius-md); background: rgba(255,255,255,0.03); border: 1px solid var(--border); }
.recs-stat-value { display: block; font-size: 22px; font-weight: 700; color: var(--accent-gold); }
.recs-stat-label { font-size: 11px; color: var(--text-secondary); }
.picks-header-row { display: flex; align-items: center; gap: 10px; margin-bottom: 4px; }

.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { display: flex; align-items: baseline; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.section-actions { margin-left: auto; }

.picks-grid { display: grid; grid-template-columns: repeat(3,1fr); gap: 12px; }
.pick-card { padding: 16px; border-radius: var(--radius-md); cursor: pointer; transition: all .25s; position: relative; }
.pick-card.active { border-color: var(--accent-gold); box-shadow: 0 0 16px rgba(240,185,11,.2); }
.pick-card:hover { transform: translateY(-2px); border-color: var(--border-gold); box-shadow: var(--shadow-glow); }
.pick-rank { position: absolute; top: 12px; right: 12px; width: 24px; height: 24px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; }
.rank-1 { background: var(--accent-gold); color: #000; }
.rank-2 { background: rgba(139,147,176,.3); }
.rank-3 { background: rgba(139,147,176,.2); }

.pick-info { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 10px; }
.pick-symbol { display: block; font-size: 15px; font-weight: 700; }
.pick-name { font-size: 12px; color: var(--text-secondary); }
.pick-price { display: block; font-size: 18px; font-weight: 600; }
.pick-change { font-size: 13px; font-weight: 600; }
.pick-change.up { color: var(--positive); }
.pick-change.down { color: var(--negative); }

.pick-score-ring { position: relative; width: 40px; height: 40px; display: flex; align-items: center; justify-content: center; margin-bottom: 8px; }
.score-svg { position: absolute; width: 40px; height: 40px; }
.score-num { font-size: 11px; font-weight: 700; color: var(--accent-gold); }

.pick-reason { font-size: 12px; color: var(--text-secondary); line-height: 1.6; margin-bottom: 10px; }
.reason-label { color: var(--accent-gold); font-weight: 600; }
.pick-dim-bars { display: grid; gap: 4px; margin-bottom: 10px; }
.pick-score-badges { display: flex; gap: 6px; margin-bottom: 8px; }
.pick-badge { padding: 2px 8px; border-radius: 4px; font-size: 11px; font-weight: 700; }
.pick-badge.badge-high { background: var(--accent-gold-glow); color: var(--accent-gold); }
.pick-badge.badge-mid { background: rgba(59,130,246,.1); color: var(--accent-blue); }
.pick-badge.badge-low { background: rgba(139,147,176,.1); color: var(--text-secondary); }
.pick-badge.risk-badge { background: rgba(255,255,255,.04); color: var(--text-secondary); }
.pick-footer { display: flex; align-items: center; justify-content: space-between; }
.pick-journey { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-top: 12px; }
.journey-btn { padding: 9px 12px; font-size: 12px; }
.pick-tags { display: flex; gap: 6px; }
.pick-strategy { font-size: 11px; color: var(--text-muted); }
.pick-detail-btn { padding: 4px 12px; border-radius: var(--radius-full); border: 1px solid var(--accent-gold); color: var(--accent-gold); font-size: 11px; font-weight: 600; background: none; cursor: pointer; }
.pick-detail-btn:hover { background: var(--accent-gold-glow); }

/* Chart section */
.chart-stock-select { padding: 6px 12px; border-radius: var(--radius-sm); background: rgba(255,255,255,0.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 13px; }
.chart-container { padding: 20px; border-radius: var(--radius-md); }
.chart-header { margin-bottom: 20px; }
.chart-stock-name { font-size: 18px; font-weight: 700; margin-right: 12px; }
.chart-stock-price { font-size: 24px; font-weight: 700; margin-right: 8px; }
.chart-stock-change { font-size: 14px; font-weight: 600; }
.chart-stock-change.up { color: var(--positive); }

.sim-chart { position: relative; height: 200px; }
.chart-grid { position: absolute; inset: 0; }
.chart-gridline { position: absolute; left: 0; right: 0; border-top: 1px solid rgba(255,255,255,0.04); }
.chart-bars { position: absolute; inset: 0; display: flex; align-items: flex-end; justify-content: space-around; padding: 0 20px; }
.candle-wrap { position: relative; display: flex; flex-direction: column; align-items: center; width: 40px; }
.candle { position: relative; width: 16px; border-radius: 2px; }
.candle.up { background: var(--positive); }
.candle.down { background: var(--negative); }
.wick { position: absolute; left: 50%; width: 2px; background: var(--text-muted); transform: translateX(-50%); }
.chart-annotation { position: absolute; }
.anno-dot { display: block; width: 6px; height: 6px; border-radius: 50%; margin: 0 auto; }
.anno-tooltip { display: none; position: absolute; bottom: 10px; left: 50%; transform: translateX(-50%); white-space: nowrap; padding: 4px 8px; border-radius: 4px; background: var(--bg-elevated); font-size: 11px; color: var(--text-primary); border: 1px solid var(--border); }
.candle-wrap:hover .anno-tooltip { display: block; }
.chart-labels { position: absolute; bottom: -20px; left: 0; right: 0; display: flex; justify-content: space-around; padding: 0 20px; }
.chart-labels span { font-size: 10px; color: var(--text-muted); }

/* Strategy summary */
.strategy-summary-grid { display: grid; grid-template-columns: repeat(3,1fr); gap: 10px; }
.strat-mini-card { padding: 14px; border-radius: var(--radius-md); }
.strat-mini-header { display: flex; align-items: center; gap: 8px; margin-bottom: 10px; }
.strat-mini-symbol { font-family: var(--font-mono); font-size: 12px; color: var(--text-secondary); }
.strat-mini-name { font-weight: 700; font-size: 14px; }
.strat-mini-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.strat-label { display: block; font-size: 10px; color: var(--text-muted); }
.strat-val { font-size: 13px; font-weight: 600; }

@media (max-width: 1200px) { .picks-grid, .strategy-summary-grid { grid-template-columns: repeat(2,1fr); } }
@media (max-width: 900px) { .picks-grid, .strategy-summary-grid { grid-template-columns: 1fr; } .pick-journey { grid-template-columns: 1fr; } }
.predict-btn { display: inline-flex; align-items: center; gap: 8px; padding: 9px 22px; border-radius: var(--radius-full); background: #f0b90b; background: linear-gradient(135deg, #f0b90b, #d4a00a); color: #0b0e1a; font-size: 13px; font-weight: 700; border: none; cursor: pointer; transition: all .25s; box-shadow: 0 2px 12px rgba(240,185,11,.25); margin-left: auto; flex-shrink: 0; }
.predict-btn:hover { transform: translateY(-1px); box-shadow: 0 4px 20px rgba(240,185,11,.35); }
.predict-btn:disabled { opacity: .5; cursor: default; transform: none; }
.predict-loading { display: inline-flex; align-items: center; gap: 6px; }
.predict-spinner { display: inline-block; width: 14px; height: 14px; border: 2px solid rgba(0,0,0,.15); border-top-color: #0b0e1a; border-radius: 50%; animation: spin .6s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }
.pred-chart { display: flex; align-items: flex-end; gap: 10px; margin-bottom: 12px; }
.pred-bars { display: flex; align-items: flex-end; gap: 4px; flex: 1; height: 80px; }
.pred-bar-group { display: flex; flex-direction: column; align-items: center; flex: 1; }
.pred-bar { width: 100%; max-width: 24px; border-radius: 3px 3px 0 0; min-height: 2px; transition: height .5s; }
.pred-bar-up { background: linear-gradient(to top, var(--positive), rgba(0,200,151,.4)); }
.pred-bar-down { background: linear-gradient(to top, var(--negative), rgba(255,71,87,.4)); }
.pred-bar-label { font-size: 9px; color: var(--text-muted); margin-top: 2px; }
.pred-avg { font-size: 20px; font-weight: 800; color: var(--positive); white-space: nowrap; }
.pred-stats { display: flex; gap: 12px; flex-wrap: wrap; font-size: 13px; margin-bottom: 12px; color: var(--text-secondary); }
.pred-matches-title { font-size: 12px; color: var(--text-muted); margin-bottom: 6px; }
.pred-match-row { display: flex; gap: 12px; font-size: 12px; padding: 4px 0; border-bottom: 1px solid rgba(255,255,255,.03); }
.pred-match-row span { flex: 1; }

</style>
