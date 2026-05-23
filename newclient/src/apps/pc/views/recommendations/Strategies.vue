<template>
  <div class="strategies-page">
    <section class="flow-strip glass fade-in-up">
      <div class="flow-copy">
        <p class="flow-kicker">机会主线</p>
        <h2>第 2 步：形成策略</h2>
        <p>{{ contextSummary }}</p>
      </div>
      <div class="flow-steps">
        <span class="flow-step">1. 每日推荐</span>
        <span class="flow-step active">2. 交易策略</span>
        <span class="flow-step">3. 深度推演</span>
      </div>
      <div class="flow-actions">
        <button class="flow-btn ghost" @click="goToRecommendations()">回看推荐来源</button>
        <button class="flow-btn" @click="goToForecastLab()">进入深度推演</button>
      </div>
    </section>

    <section class="section fade-in-up">
      <div class="section-header">
        <h2 class="section-title">交易策略</h2>
        <p class="section-subtitle">{{ loading ? '加载中...' : 'AI 为当前市场环境匹配的交易策略体系' }}</p>
      </div>
      <div class="strategies-grid">
        <article v-for="strategy in strategyList" :key="strategy.id" class="strategy-card glass" @click="$router.push('/futures/strategy/' + strategy.id)" style="cursor:pointer">
          <div class="strat-type-badge" :class="strategy.type || 'trend'">{{ strategy.typeLabel || '策略' }}</div>
          <h3 class="strat-name">{{ strategy.name }}</h3>
          <p class="strat-stock">{{ strategy.stockName || strategy.contract || '' }}{{ strategy.symbol ? '（' + strategy.symbol + '）' : '' }}</p>
          <div class="strat-details">
            <div class="strat-detail-block">
              <h4>方向</h4>
              <p>{{ strategy.directionText || strategy.direction || '中性' }}</p>
            </div>
            <div class="strat-row">
              <div class="strat-detail-block">
                <h4>风险等级</h4>
                <p :style="{color: strategy.riskLevel === 'HIGH' ? 'var(--negative)' : strategy.riskLevel === 'LOW' ? 'var(--positive)' : 'var(--accent-gold)'}">{{ strategy.riskLabel || strategy.risk_level || '中等' }}</p>
              </div>
              <div class="strat-detail-block">
                <h4>仓位</h4>
                <p style="color:var(--accent-gold);font-weight:600">{{ strategy.positionSize || strategy.position_range || '10%' }}</p>
              </div>
            </div>
            <p class="strat-reason">{{ strategy.reason_summary || strategy.reason || 'AI 策略推荐' }}</p>
          </div>
          <div class="strat-actions">
            <button class="btn-ghost strat-backtest-btn" @click.stop="$router.push('/recommendations/backtest')">查看回测 →</button>
            <button class="btn-primary strat-forecast-btn" @click.stop="goToForecastLab(strategy)">进入深度推演</button>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { STRATEGIES as MOCK_STRATEGIES } from "@/mock/recommendations.js";
import { listFuturesStrategies } from "@/api/market.js";
import { buildForecastContextQuery } from "@/shared/lib/forecast-context.js";

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const strategyList = ref(MOCK_STRATEGIES);
const selectedSymbol = computed(() => String(route.query.symbol || "").trim());
const selectedName = computed(() => String(route.query.name || "").trim());
const contextSummary = computed(() => {
  if (!selectedSymbol.value) {
    return "把推荐机会收敛成可执行策略，再继续进入深度推演确认运行与报告。";
  }
  return `你当前正在承接 ${selectedName.value || selectedSymbol.value} 的推荐机会，先看策略，再继续进入深度推演。`;
});

async function loadStrategies() {
  loading.value = true;
  try {
    const result = await listFuturesStrategies({ page: 1, page_size: 10 });
    if (result?.items?.length) {
      strategyList.value = result.items.map((s) => ({
        id: s.id, name: s.name, contract: s.contract,
        symbol: '', stockName: '',
        type: s.direction === 'LONG' ? 'breakout' : 'trend',
        typeLabel: s.direction === 'LONG' ? '做多策略' : s.direction === 'SHORT' ? '做空策略' : '中性策略',
        directionText: s.direction === 'LONG' ? '做多' : s.direction === 'SHORT' ? '做空' : '中性',
        riskLevel: s.risk_level, riskLabel: s.risk_level === 'HIGH' ? '高风险' : s.risk_level === 'LOW' ? '低风险' : '中风险',
        positionSize: s.position_range, reason: s.reason_summary,
        reason_summary: s.reason_summary
      }));
    }
  } catch { /* use mock */ }
  finally { loading.value = false; }
}

function goToRecommendations() {
  router.push({
    path: "/recommendations",
    query: selectedSymbol.value
      ? { symbol: selectedSymbol.value, name: selectedName.value || "", from: "strategies" }
      : { from: "strategies" }
  });
}

function goToForecastLab(strategy) {
  const target = selectedSymbol.value || strategy?.symbol || strategy?.contract || "";
  const targetType = strategy?.id ? "FUTURES" : "STOCK";
  const targetLabel = selectedName.value || strategy?.name || strategy?.contract || "";
  router.push({
    path: "/forecast-lab",
    query: buildForecastContextQuery({
      targetType,
      targetId: strategy?.id || "",
      targetKey: target,
      targetLabel,
      source: "STRATEGY",
      sourceId: strategy?.id || "",
      sourcePath: "/recommendations/strategies",
      from: "strategies",
      strategyId: strategy?.id || ""
    })
  });
}

onMounted(loadStrategies);
</script>

<style scoped>
.strategies-page { display: grid; gap: 20px; max-width: 1400px; }
.flow-strip { display: grid; gap: 14px; padding: 18px 20px; border-radius: var(--radius-lg); border: 1px solid var(--border); background: linear-gradient(135deg, rgba(240,185,11,.08), rgba(255,255,255,.02)); }
.flow-copy h2 { font-size: 22px; font-weight: 800; margin: 4px 0 6px; }
.flow-copy p { color: var(--text-secondary); line-height: 1.7; }
.flow-kicker { font-size: 11px; color: var(--accent-gold); text-transform: uppercase; letter-spacing: .08em; }
.flow-steps { display: flex; gap: 8px; flex-wrap: wrap; }
.flow-step { padding: 8px 12px; border-radius: var(--radius-full); border: 1px solid var(--border); color: var(--text-secondary); font-size: 12px; }
.flow-step.active { border-color: var(--border-gold); color: var(--accent-gold); background: var(--accent-gold-glow); }
.flow-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.flow-btn, .btn-primary { display: inline-flex; align-items: center; justify-content: center; padding: 10px 14px; border-radius: var(--radius-full); font-size: 13px; font-weight: 600; cursor: pointer; }
.flow-btn { border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.flow-btn.ghost, .btn-ghost { border: 1px solid var(--border); background: none; color: var(--text-secondary); }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.strategies-grid { display: grid; gap: 16px; grid-template-columns: repeat(3,1fr); }
.strategy-card { padding: 20px; border-radius: var(--radius-md); }
.strat-type-badge { display: inline-block; padding: 3px 10px; border-radius: var(--radius-full); font-size: 11px; font-weight: 600; margin-bottom: 10px; }
.strat-type-badge.breakout { background: var(--positive-bg); color: var(--positive); }
.strat-type-badge.defensive { background: rgba(59,130,246,.1); color: var(--accent-blue); }
.strat-type-badge.trend { background: var(--accent-gold-glow); color: var(--accent-gold); }
.strat-name { font-size: 18px; font-weight: 700; margin-bottom: 4px; }
.strat-stock { font-size: 12px; color: var(--text-secondary); margin-bottom: 16px; }
.strat-details { display: grid; gap: 12px; }
.strat-detail-block h4 { font-size: 12px; color: var(--text-muted); margin-bottom: 4px; text-transform: uppercase; letter-spacing: .5px; }
.strat-detail-block p { font-size: 13px; line-height: 1.6; color: var(--text-secondary); }
.strat-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.strat-reason { font-size: 12px; color: var(--text-muted); line-height: 1.6; padding: 8px; background: rgba(255,255,255,.02); border-radius: 6px; }
.strat-actions { display: grid; gap: 8px; margin-top: 16px; }
.strat-backtest-btn, .strat-forecast-btn { width: 100%; justify-content: center; display: flex; }
.strat-forecast-btn { border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.btn-ghost { padding: 8px 16px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 12px; color: var(--text-secondary); background: none; cursor: pointer; }
.btn-ghost:hover { border-color: var(--border-light); color: var(--text-primary); }
@media (max-width: 1200px) { .strategies-grid { grid-template-columns: 1fr; } }
</style>
