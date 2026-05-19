<template>
  <div class="strategies-page">
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
          <button class="btn-ghost strat-backtest-btn" @click="$router.push('/recommendations/backtest')">查看回测 →</button>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { STRATEGIES as MOCK_STRATEGIES } from "@/mock/recommendations.js";
import { listFuturesStrategies } from "@/api/market.js";

const loading = ref(false);
const strategyList = ref(MOCK_STRATEGIES);

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

onMounted(loadStrategies);
</script>

<style scoped>
.strategies-page { display: grid; gap: 20px; max-width: 1400px; }
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
.strat-backtest-btn { margin-top: 16px; width: 100%; justify-content: center; display: flex; }
.btn-ghost { padding: 8px 16px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 12px; color: var(--text-secondary); background: none; cursor: pointer; }
.btn-ghost:hover { border-color: var(--border-light); color: var(--text-primary); }
@media (max-width: 1200px) { .strategies-grid { grid-template-columns: 1fr; } }
</style>
