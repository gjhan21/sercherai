<template>
  <div class="h5-futures">
    <div class="h5-section-title">期货套利</div>
    <div class="h5-arb-list">
      <div v-for="item in arbitrageList" :key="item.id" class="h5-arb-card" @click="$router.push('/futures/arbitrage/' + item.id)">
        <div class="h5-arb-pair">
          <span class="h5-arb-a">{{ item.contract_a }}</span>
          <span class="h5-arb-v">⇄</span>
          <span class="h5-arb-b">{{ item.contract_b }}</span>
          <span class="h5-arb-type">{{ typeLabel(item.type) }}</span>
        </div>
        <div class="h5-arb-grid">
          <span>进: {{ item.entry_point }}</span><span>止: {{ item.stop_point }}</span>
          <span>分位: {{ ((item.percentile||0)*100).toFixed(0) }}%</span>
        </div>
      </div>
      <div v-if="!arbitrageList.length && !loading" class="h5-empty">暂无套利数据</div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { listFuturesArbitrage } from "@/api/market.js";

const MOCK_ARB = [{id:"arb_1",contract_a:"IF 主连",contract_b:"IC 主连",type:"CALENDAR",entry_point:186,exit_point:132,stop_point:208,percentile:0.86,risk_level:"MEDIUM"},{id:"arb_2",contract_a:"IH 主连",contract_b:"IF 主连",type:"CROSS",entry_point:-76,exit_point:-42,stop_point:-94,percentile:0.79,risk_level:"MEDIUM"}];

const loading = ref(false);
const arbitrageList = ref(MOCK_ARB);
function typeLabel(t) { return t === 'CALENDAR' ? '跨期' : '跨品种'; }

async function loadData() {
  loading.value = true;
  try { const r = await listFuturesArbitrage({page:1,page_size:20}); if (r?.items?.length) arbitrageList.value = r.items; }
  catch { /* mock */ }
  finally { loading.value = false; }
}
onMounted(loadData);
</script>

<style scoped>
.h5-futures { display: grid; gap: 14px; }
.h5-section-title { font-size: 18px; font-weight: 700; padding: 0 4px; }
.h5-arb-list { display: grid; gap: 8px; }
.h5-arb-card { padding: 14px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-md); }
.h5-arb-card:active { background: rgba(255,255,255,.03); }
.h5-arb-pair { display: flex; align-items: center; gap: 6px; margin-bottom: 10px; }
.h5-arb-a, .h5-arb-b { font-family: var(--font-mono); font-size: 15px; font-weight: 700; }
.h5-arb-v { color: var(--text-muted); }
.h5-arb-type { margin-left: auto; font-size: 11px; padding: 2px 8px; border-radius: var(--radius-full); background: var(--accent-gold-glow); color: var(--accent-gold); }
.h5-arb-grid { display: flex; gap: 12px; font-size: 12px; color: var(--text-secondary); }
.h5-empty { text-align: center; padding: 40px; color: var(--text-secondary); }
</style>
