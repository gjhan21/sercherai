<template>
  <div class="h5-arb-detail" v-if="item">
    <div class="h5-arb-header">
      <div class="h5-arb-pair">
        <span class="h5-arb-a">{{ item.contract_a }}</span>
        <span class="h5-arb-v">⇄</span>
        <span class="h5-arb-b">{{ item.contract_b }}</span>
      </div>
      <div class="h5-arb-type">{{ typeLabel(item.type) }}</div>
    </div>
    <div class="h5-arb-grid">
      <div><span>进入</span><strong>{{ item.entry_point }}</strong></div>
      <div><span>平仓</span><strong>{{ item.exit_point }}</strong></div>
      <div><span>止损</span><strong>{{ item.stop_point }}</strong></div>
      <div><span>分位数</span><strong style="color:var(--accent-gold)">{{ ((item.percentile||0)*100).toFixed(1) }}%</strong></div>
      <div><span>Z值</span><strong>{{ item.z_score || '-' }}</strong></div>
      <div><span>风险</span><strong :style="{color:item.risk_level==='HIGH'?'var(--negative)':'var(--accent-gold)'}">{{ riskLabel(item.risk_level) }}</strong></div>
    </div>
    <div class="h5-forecast-entry">
      <strong>深度推演</strong>
      <p>如当前期货机会已生成 L3 推演，可先查看完整报告；若需要新的研究任务，也可以带着这次机会的完整上下文进入深度推演入口。</p>
      <button class="h5-forecast-link h5-forecast-entry-btn" @click="router.push(forecastLabEntryTo)">带着完整上下文进入深度推演</button>
    </div>
    <button class="h5-alert-btn" @click="showAlert = !showAlert">🔔 {{ showAlert ? '取消' : '设置提醒' }}</button>
    <div v-if="showAlert" class="h5-alert-form">
      <input v-model="alertThreshold" placeholder="阈值点数" type="number" />
      <button @click="createAlert">确认</button>
      <p v-if="alertMsg" class="h5-alert-msg">{{ alertMsg }}</p>
    </div>
  </div>
  <div v-else class="h5-empty"><p>加载中...</p></div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getFuturesArbitrageDetail, createFuturesAlert } from "@/api/market.js";
import { buildForecastContextQuery } from "@/shared/lib/forecast-context.js";

const MOCK_ARB = [{id:"arb_1",contract_a:"IF 主连",contract_b:"IC 主连",type:"CALENDAR",entry_point:186,exit_point:132,stop_point:208,percentile:0.86},{id:"arb_2",contract_a:"IH 主连",contract_b:"IF 主连",type:"CROSS",entry_point:-76,exit_point:-42,stop_point:-94,percentile:0.79}];

const route = useRoute();
const router = useRouter();
const item = ref(null);
const showAlert = ref(false);
const alertThreshold = ref("");
const alertMsg = ref("");
const forecastLabEntryTo = computed(() => {
  const current = item.value || {};
  const targetKey = current.contract_a || current.contract_b || "";
  const targetLabel = [current.contract_a, current.contract_b].filter(Boolean).join(" / ") || targetKey;
  return {
    path: "/forecast-lab",
    query: buildForecastContextQuery({
      targetType: "FUTURES",
      targetId: current.id || route.params.id,
      targetKey,
      targetLabel,
      source: "FUTURES_DETAIL",
      sourceId: current.id || route.params.id,
      sourcePath: current.id ? `/futures/arbitrage/${current.id}` : "/markets",
      from: "strategies"
    })
  };
});

function typeLabel(t) { return t === 'CALENDAR' ? '跨期' : '跨品种'; }
function riskLabel(r) { const m = { HIGH:'高风险', MEDIUM:'中风险', LOW:'低风险' }; return m[r] || r; }

async function loadDetail() {
  const id = route.params.id;
  try { const r = await getFuturesArbitrageDetail(id); if (r) item.value = r; }
  catch { item.value = MOCK_ARB.find(a => String(a.id) === id) || MOCK_ARB[0]; }
}

async function createAlert() {
  if (!alertThreshold.value) return;
  try { await createFuturesAlert({ contract: item.value.contract_a, alert_type: 'SPREAD', threshold: Number(alertThreshold.value) }); alertMsg.value = '设置成功！'; }
  catch { alertMsg.value = '设置失败'; }
}

onMounted(loadDetail);
</script>

<style scoped>
.h5-arb-detail { display: grid; gap: 14px; }
.h5-arb-header { display: flex; justify-content: space-between; align-items: center; padding: 16px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); }
.h5-arb-pair { display: flex; align-items: center; gap: 8px; }
.h5-arb-a, .h5-arb-b { font-family: var(--font-mono); font-size: 18px; font-weight: 700; }
.h5-arb-v { color: var(--text-muted); }
.h5-arb-type { font-size: 12px; padding: 3px 10px; border-radius: var(--radius-full); background: var(--accent-gold-glow); color: var(--accent-gold); }
.h5-arb-grid { display: grid; grid-template-columns: repeat(3,1fr); gap: 8px; }
.h5-arb-grid div { padding: 12px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-sm); text-align: center; }
.h5-arb-grid span { display: block; font-size: 11px; color: var(--text-muted); margin-bottom: 4px; }
.h5-arb-grid strong { font-size: 16px; font-weight: 700; }
.h5-alert-btn { width: 100%; padding: 12px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 14px; background: none; color: var(--text-primary); cursor: pointer; }
.h5-alert-form { padding: 14px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-md); display: grid; gap: 10px; }
.h5-alert-form input { padding: 10px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 14px; }
.h5-alert-form button { padding: 10px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-weight: 600; border: none; cursor: pointer; }
.h5-alert-msg { font-size: 12px; color: var(--positive); }
.h5-empty { text-align: center; padding: 40px; color: var(--text-secondary); }
.h5-forecast-entry { padding: 14px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-md); display: grid; gap: 8px; }
.h5-forecast-entry p { font-size: 13px; color: var(--text-secondary); line-height: 1.7; }
.h5-forecast-link { display: inline-flex; justify-content: center; padding: 10px; border-radius: var(--radius-full); border: 1px solid var(--accent-gold); color: var(--accent-gold); text-decoration: none; font-size: 13px; font-weight: 600; }
.h5-forecast-entry-btn { width: 100%; background: rgba(240,185,11,.08); cursor: pointer; }
</style>
