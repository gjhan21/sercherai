<template>
  <div class="backtest-page">
    <section class="section fade-in-up">
      <div class="section-header">
        <h2 class="section-title">回测结果</h2>
        <p class="section-subtitle">{{ BACKTEST_RESULTS.strategy }} · {{ BACKTEST_RESULTS.period }}</p>
      </div>
      <div class="backtest-metrics">
        <div v-for="m in metrics" :key="m.label" class="metric-card glass">
          <span class="metric-label">{{ m.label }}</span>
          <span class="metric-value" :style="{color: m.color || 'var(--accent-gold)'}">{{ m.value }}</span>
        </div>
      </div>
    </section>

    <section class="section fade-in-up fade-in-up-delay-1">
      <div class="section-header">
        <h2 class="section-title">交易明细</h2>
        <p class="section-subtitle">共 {{ BACKTEST_RESULTS.trades.length }} 笔交易</p>
      </div>
      <div class="table-wrap">
        <table class="data-table">
          <thead>
            <tr>
              <th>#</th><th>入场日期</th><th>入场价格</th><th>出场日期</th><th>出场价格</th><th>收益率</th><th>持有时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="t in BACKTEST_RESULTS.trades" :key="t.id" class="trade-row">
              <td>{{ t.id }}</td>
              <td>{{ t.entryDate }}</td>
              <td class="mono">{{ t.entryPrice.toFixed(2) }}</td>
              <td>{{ t.exitDate }}</td>
              <td class="mono">{{ t.exitPrice.toFixed(2) }}</td>
              <td><span :class="t.returnPct >= 0 ? 'up' : 'down'">{{ t.returnPct >= 0 ? '+' : '' }}{{ t.returnPct }}%</span></td>
              <td>{{ t.duration }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { BACKTEST_RESULTS } from "@/mock/recommendations.js";

const s = BACKTEST_RESULTS.summary;
const metrics = [
  { label: '总收益率', value: s.totalReturn, color: 'var(--positive)' },
  { label: '年化收益', value: s.annualized, color: 'var(--positive)' },
  { label: '胜率', value: s.winRate, color: 'var(--positive)' },
  { label: '最大回撤', value: s.maxDrawdown, color: 'var(--negative)' },
  { label: '夏普比率', value: s.sharpeRatio, color: 'var(--accent-gold)' },
  { label: '交易次数', value: s.totalTrades, color: 'var(--accent-cyan)' },
  { label: '盈亏比', value: s.profitFactor, color: 'var(--accent-gold)' }
];
</script>

<style scoped>
.backtest-page { display: grid; gap: 20px; max-width: 1400px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.backtest-metrics { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 12px; }
.metric-card { padding: 16px; border-radius: var(--radius-md); text-align: center; }
.metric-label { display: block; font-size: 11px; color: var(--text-secondary); margin-bottom: 8px; text-transform: uppercase; letter-spacing: .5px; }
.metric-value { display: block; font-size: 24px; font-weight: 700; }
.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { text-align: left; padding: 10px 12px; font-size: 11px; color: var(--text-secondary); font-weight: 600; text-transform: uppercase; letter-spacing: .5px; border-bottom: 1px solid var(--border); }
.data-table td { padding: 10px 12px; border-bottom: 1px solid rgba(255,255,255,.03); }
.trade-row:hover { background: rgba(255,255,255,.02); }
.mono { font-family: var(--font-mono); }
.up { color: var(--positive); font-weight: 600; }
.down { color: var(--negative); font-weight: 600; }
</style>
