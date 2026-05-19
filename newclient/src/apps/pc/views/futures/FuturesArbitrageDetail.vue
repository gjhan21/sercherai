<template>
  <div class="detail-page" v-if="item">
    <button class="back-btn glass" @click="$router.push('/futures/arbitrage')">← 返回套利列表</button>
    <section class="section fade-in-up">
      <div class="detail-header">
        <div class="detail-pair">
          <span class="detail-a">{{ item.contract_a }}</span>
          <span class="detail-vs">⇄</span>
          <span class="detail-b">{{ item.contract_b }}</span>
          <span class="tag" :class="item.type === 'CALENDAR' ? 'tag-blue' : 'tag-gold'">{{ typeLabel(item.type) }}</span>
        </div>
        <div class="arb-status-large" :class="item.status?.toLowerCase()">{{ statusLabel(item.status) }}</div>
      </div>
      <div class="detail-grid">
        <div class="detail-stat"><span class="ds-label">进入点位</span><span class="ds-value">{{ item.entry_point }}</span></div>
        <div class="detail-stat"><span class="ds-label">平仓点位</span><span class="ds-value">{{ item.exit_point }}</span></div>
        <div class="detail-stat"><span class="ds-label">止损点位</span><span class="ds-value">{{ item.stop_point }}</span></div>
        <div class="detail-stat"><span class="ds-label">分位数</span><span class="ds-value" style="color:var(--accent-gold)">{{ ((item.percentile || 0) * 100).toFixed(1) }}%</span></div>
        <div class="detail-stat"><span class="ds-label">Z 值</span><span class="ds-value">{{ item.z_score || '-' }}</span></div>
        <div class="detail-stat"><span class="ds-label">风险等级</span><span class="ds-value" :style="{color: item.risk_level === 'HIGH' ? 'var(--negative)' : 'var(--accent-gold)'}">{{ riskLabel(item.risk_level) }}</span></div>
      </div>
      <div class="detail-meta" v-if="item.created_at || item.updated_at">
        <span>创建: {{ item.created_at?.slice(0,10) || '-' }}</span>
        <span>更新: {{ item.updated_at?.slice(0,10) || '-' }}</span>
      </div>
      <div class="detail-actions"><button class="btn-primary" @click="showAlert = !showAlert">🔔 {{ showAlert ? '取消提醒' : '设置提醒' }}</button></div>
      <div v-if="showAlert" class="alert-form glass">
        <p>提醒将在价差触及阈值时推送</p>
        <div class="alert-input-row">
          <input v-model="alertThreshold" placeholder="阈值点数" type="number" />
          <button class="btn-primary" @click="createAlert">确认</button>
        </div>
        <p v-if="alertMsg" class="alert-msg">{{ alertMsg }}</p>
      </div>
    </section>
  </div>
  <div v-else class="not-found"><p>加载中...</p></div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getFuturesArbitrageDetail, createFuturesAlert } from "@/api/market.js";

const MOCK_ARB = [
  { id: "arb_1", contract_a: "IF 主连", contract_b: "IC 主连", type: "CALENDAR", entry_point: 186, exit_point: 132, stop_point: 208, percentile: 0.86, z_score: 1.28, risk_level: "MEDIUM", status: "WATCH", spread: 54 },
  { id: "arb_2", contract_a: "IH 主连", contract_b: "IF 主连", type: "CROSS", entry_point: -76, exit_point: -42, stop_point: -94, percentile: 0.79, z_score: 0.95, risk_level: "MEDIUM", status: "ACTIVE", spread: 34 }
];

const route = useRoute();
const router = useRouter();
const item = ref(null);
const showAlert = ref(false);
const alertThreshold = ref("");
const alertMsg = ref("");

function typeLabel(t) { const m = { CALENDAR:'跨期', CROSS:'跨品种' }; return m[t] || t || '套利'; }
function statusLabel(s) { const m = { WATCH:'观察中', ACTIVE:'可执行', EXPIRED:'已失效' }; return m[s] || s; }
function riskLabel(r) { const m = { HIGH:'高风险', MEDIUM:'中风险', LOW:'低风险' }; return m[r] || r; }

async function loadDetail() {
  const id = route.params.id;
  if (!id) return;
  try {
    const r = await getFuturesArbitrageDetail(id);
    if (r) item.value = r;
  } catch { item.value = MOCK_ARB.find(a => String(a.id) === id) || MOCK_ARB[0]; }
}

async function createAlert() {
  if (!alertThreshold.value) return;
  try {
    await createFuturesAlert({ contract: item.value.contract_a, alert_type: 'SPREAD', threshold: Number(alertThreshold.value) });
    alertMsg.value = '提醒设置成功！';
  } catch { alertMsg.value = '设置失败，请重试'; }
}

onMounted(loadDetail);
</script>

<style scoped>
.detail-page { max-width: 900px; display: grid; gap: 16px; }
.back-btn { display: inline-flex; padding: 8px 16px; border-radius: var(--radius-full); font-size: 13px; color: var(--text-secondary); cursor: pointer; width: auto; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.detail-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.detail-pair { display: flex; align-items: center; gap: 10px; }
.detail-a, .detail-b { font-family: var(--font-mono); font-size: 22px; font-weight: 800; }
.detail-vs { color: var(--text-muted); font-size: 18px; }
.arb-status-large { padding: 6px 16px; border-radius: var(--radius-full); font-size: 14px; font-weight: 600; }
.arb-status-large.watch { background: rgba(139,147,176,.1); color: var(--text-secondary); }
.arb-status-large.active { background: var(--positive-bg); color: var(--positive); }
.arb-status-large.expired { background: var(--negative-bg); color: var(--negative); }
.detail-grid { display: grid; grid-template-columns: repeat(3,1fr); gap: 12px; margin-bottom: 16px; }
.detail-stat { padding: 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.02); border: 1px solid var(--border); text-align: center; }
.ds-label { display: block; font-size: 12px; color: var(--text-muted); margin-bottom: 4px; text-transform: uppercase; }
.ds-value { font-size: 22px; font-weight: 700; }
.detail-meta { display: flex; gap: 16px; font-size: 12px; color: var(--text-muted); margin-bottom: 16px; }
.btn-primary { padding: 10px 20px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 13px; font-weight: 600; border: none; cursor: pointer; }
.alert-form { padding: 16px; border-radius: var(--radius-md); margin-top: 12px; }
.alert-form p { font-size: 13px; color: var(--text-secondary); margin-bottom: 10px; }
.alert-input-row { display: flex; gap: 10px; }
.alert-input-row input { flex: 1; padding: 10px 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 14px; }
.alert-msg { margin-top: 8px; font-size: 12px; color: var(--positive); }
.not-found { text-align: center; padding: 60px; color: var(--text-secondary); }
</style>
