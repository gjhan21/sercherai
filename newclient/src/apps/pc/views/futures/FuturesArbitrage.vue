<template>
  <div class="futures-page">
    <section class="section fade-in-up">
      <div class="section-header">
        <h2 class="section-title">期货套利</h2>
        <p class="section-subtitle">{{ loading ? '加载中...' : 'AI 监测到的跨期与跨品种套利机会' }}</p>
      </div>
      <div class="arbitrage-list">
        <div v-for="item in arbitrageList" :key="item.id" class="arb-card glass" @click="$router.push('/futures/arbitrage/' + item.id)">
          <div class="arb-top">
            <div class="arb-pair">
              <span class="arb-contract-a">{{ item.contract_a }}</span>
              <span class="arb-vs">⇄</span>
              <span class="arb-contract-b">{{ item.contract_b }}</span>
              <span class="tag" :class="item.type === 'CALENDAR' ? 'tag-blue' : 'tag-gold'">{{ typeLabel(item.type) }}</span>
            </div>
            <div class="arb-status" :class="item.status?.toLowerCase()">{{ statusLabel(item.status) }}</div>
          </div>
          <div class="arb-grid">
            <div class="arb-stat"><span class="arb-stat-label">进入点位</span><span class="arb-stat-value">{{ item.entry_point }}</span></div>
            <div class="arb-stat"><span class="arb-stat-label">平仓点位</span><span class="arb-stat-value">{{ item.exit_point }}</span></div>
            <div class="arb-stat"><span class="arb-stat-label">止损点位</span><span class="arb-stat-value">{{ item.stop_point }}</span></div>
            <div class="arb-stat"><span class="arb-stat-label">分位数</span><span class="arb-stat-value" style="color:var(--accent-gold)">{{ ((item.percentile || 0) * 100).toFixed(0) }}%</span></div>
            <div class="arb-stat"><span class="arb-stat-label">风险</span><span class="arb-stat-value" :style="{color: item.risk_level === 'HIGH' ? 'var(--negative)' : 'var(--accent-gold)'}">{{ riskLabel(item.risk_level) }}</span></div>
            <div class="arb-stat"><span class="arb-stat-label">价差</span><span class="arb-stat-value">{{ item.spread || '-' }}</span></div>
          </div>
        </div>
        <div v-if="!arbitrageList.length && !loading" class="empty-box">暂无套利数据</div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { listFuturesArbitrage } from "@/api/market.js";

const MOCK = [
  { id: "arb_1", contract_a: "IF 主连", contract_b: "IC 主连", type: "CALENDAR", entry_point: 186, exit_point: 132, stop_point: 208, percentile: 0.86, z_score: 1.28, risk_level: "MEDIUM", status: "WATCH", spread: 54 },
  { id: "arb_2", contract_a: "IH 主连", contract_b: "IF 主连", type: "CROSS", entry_point: -76, exit_point: -42, stop_point: -94, percentile: 0.79, z_score: 0.95, risk_level: "MEDIUM", status: "ACTIVE", spread: 34 }
];

const loading = ref(false);
const arbitrageList = ref(MOCK);

function typeLabel(t) {
  const m = { CALENDAR: '跨期', CROSS: '跨品种', INTERTEMPORAL: '跨期套利' };
  return m[t] || t || '套利';
}
function statusLabel(s) {
  const m = { WATCH: '观察中', ACTIVE: '可执行', EXPIRED: '已失效' };
  return m[s] || s || '待确认';
}
function riskLabel(r) {
  const m = { HIGH: '高风险', MEDIUM: '中风险', LOW: '低风险' };
  return m[r] || r || '中';
}

async function loadData() {
  loading.value = true;
  try {
    const r = await listFuturesArbitrage({ page: 1, page_size: 20 });
    if (r?.items?.length) arbitrageList.value = r.items;
  } catch { /* use mock */ }
  finally { loading.value = false; }
}

onMounted(loadData);
</script>

<style scoped>
.futures-page { display: grid; gap: 20px; max-width: 1400px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.arbitrage-list { display: grid; gap: 10px; }
.arb-card { padding: 16px; border-radius: var(--radius-md); cursor: pointer; transition: all .2s; }
.arb-card:hover { border-color: var(--border-gold); box-shadow: var(--shadow-glow); }
.arb-top { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
.arb-pair { display: flex; align-items: center; gap: 8px; }
.arb-contract-a, .arb-contract-b { font-family: var(--font-mono); font-size: 16px; font-weight: 700; }
.arb-vs { color: var(--text-muted); font-size: 14px; }
.arb-status { font-size: 12px; padding: 3px 10px; border-radius: var(--radius-full); font-weight: 600; }
.arb-status.watch { background: rgba(139,147,176,.1); color: var(--text-secondary); }
.arb-status.active { background: var(--positive-bg); color: var(--positive); }
.arb-status.expired { background: var(--negative-bg); color: var(--negative); }
.arb-grid { display: grid; grid-template-columns: repeat(3,1fr); gap: 8px; }
.arb-stat { padding: 8px; border-radius: var(--radius-sm); background: rgba(255,255,255,.02); text-align: center; }
.arb-stat-label { display: block; font-size: 11px; color: var(--text-muted); margin-bottom: 4px; }
.arb-stat-value { font-size: 16px; font-weight: 700; }
.empty-box { text-align: center; padding: 40px; color: var(--text-secondary); }
</style>
