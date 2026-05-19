<template>
  <div class="detail-page" v-if="strategy">
    <button class="back-btn glass" @click="$router.push('/strategies')">← 返回策略列表</button>

    <!-- Strategy Header -->
    <section class="section fade-in-up">
      <div class="strat-header">
        <div class="strat-primary">
          <span class="strat-contract">{{ strategy.contract || '-' }}</span>
          <span class="strat-name">{{ strategy.name || '未命名策略' }}</span>
          <div class="strat-tags">
            <span class="tag" :class="strategy.direction === 'LONG' ? 'tag-green' : strategy.direction === 'SHORT' ? 'tag-red' : 'tag-neutral'">{{ directionLabel(strategy.direction) }}</span>
            <span class="tag" :class="strategy.risk_level === 'HIGH' ? 'tag-red' : strategy.risk_level === 'LOW' ? 'tag-green' : 'tag-gold'">{{ riskLabel(strategy.risk_level) }}</span>
          </div>
        </div>
        <div class="strat-status" :class="strategy.status?.toLowerCase()">{{ strategy.status === 'ACTIVE' ? '进行中' : strategy.status }}</div>
      </div>
      <div class="strat-meta-row">
        <div class="strat-meta-item"><span class="sm-label">仓位建议</span><span class="sm-value">{{ strategy.position_range || '10%' }}</span></div>
        <div class="strat-meta-item"><span class="sm-label">有效期</span><span class="sm-value">{{ (strategy.valid_from || '').slice(0,10) }} ~ {{ (strategy.valid_to || '').slice(0,10) }}</span></div>
      </div>
      <div class="strat-reason" v-if="strategy.reason_summary">
        <h3>策略理由</h3>
        <p>{{ strategy.reason_summary }}</p>
      </div>
    </section>

    <!-- AI Insight -->
    <section v-if="insight" class="section fade-in-up fade-in-up-delay-1">
      <div class="section-header"><h2 class="section-title">AI 深度分析</h2></div>
      <div class="insight-section">
        <h3>核心结论</h3>
        <p>{{ insight.explanation?.seed_summary || insight.strategy?.reason_summary || 'AI 分析中' }}</p>
      </div>
      <div class="insight-grid">
        <div class="insight-card" v-if="insight.guidance">
          <h4>操作指引</h4>
          <div class="guidance-grid">
            <div><span>方向</span><strong>{{ insight.guidance.guidance_direction === 'LONG_SPREAD' ? '做多价差' : '做空价差' }}</strong></div>
            <div><span>仓位</span><strong>{{ insight.guidance.position_level === 'LIGHT' ? '轻仓' : '重仓' }}</strong></div>
            <div><span>进入区间</span><strong>{{ insight.guidance.entry_range || '-' }}</strong></div>
            <div><span>止盈区间</span><strong>{{ insight.guidance.take_profit_range || '-' }}</strong></div>
            <div><span>止损区间</span><strong>{{ insight.guidance.stop_loss_range || '-' }}</strong></div>
          </div>
        </div>
      </div>
    </section>

    <div class="empty-section" v-if="!strategy"><p>加载中...</p></div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { getFuturesStrategyDetail, getFuturesStrategyInsight } from "@/api/market.js";

const MOCK_STRAT = { id:"strat_1", contract:"IF 主连", name:"IF跨期套利策略", direction:"LONG", risk_level:"MEDIUM", position_range:"10%-15%", valid_from:"2026-01-01", valid_to:"2026-06-30", reason_summary:"基于历史价差回归规律，当前价差处于近年高分位，开仓做空价差" };

const route = useRoute();
const strategy = ref(null);
const insight = ref(null);

function directionLabel(d) { return d === 'LONG' ? '做多' : d === 'SHORT' ? '做空' : '中性'; }
function riskLabel(r) { const m = { HIGH:'高风险', MEDIUM:'中风险', LOW:'低风险' }; return m[r] || r; }

async function loadDetail() {
  const id = route.params.id;
  if (!id) return;
  try {
    const [s, i] = await Promise.allSettled([getFuturesStrategyDetail(id), getFuturesStrategyInsight(id)]);
    if (s.status === 'fulfilled' && s.value) strategy.value = s.value;
    if (i.status === 'fulfilled' && i.value) insight.value = i.value;
  } catch { strategy.value = MOCK_STRAT; }
}

onMounted(loadDetail);
</script>

<style scoped>
.detail-page { max-width: 1000px; display: grid; gap: 16px; }
.back-btn { display: inline-flex; padding: 8px 16px; border-radius: var(--radius-full); font-size: 13px; color: var(--text-secondary); width: auto; cursor: pointer; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.strat-header { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 16px; }
.strat-contract { display: block; font-family: var(--font-mono); font-size: 13px; color: var(--text-secondary); margin-bottom: 4px; }
.strat-name { display: block; font-size: 24px; font-weight: 800; margin-bottom: 8px; }
.strat-tags { display: flex; gap: 6px; }
.strat-status { padding: 6px 14px; border-radius: var(--radius-full); font-size: 13px; font-weight: 600; }
.strat-status.active { background: var(--positive-bg); color: var(--positive); }
.strat-meta-row { display: flex; gap: 16px; margin-bottom: 16px; }
.strat-meta-item { padding: 12px 16px; border-radius: var(--radius-sm); background: rgba(255,255,255,.02); border: 1px solid var(--border); min-width: 160px; }
.sm-label { display: block; font-size: 11px; color: var(--text-muted); margin-bottom: 4px; }
.sm-value { font-size: 16px; font-weight: 700; }
.strat-reason h3 { font-size: 14px; font-weight: 700; margin-bottom: 6px; color: var(--accent-gold); }
.strat-reason p { font-size: 14px; color: var(--text-secondary); line-height: 1.7; }
.insight-section { margin-bottom: 16px; }
.insight-section h3 { font-size: 14px; font-weight: 700; margin-bottom: 8px; color: var(--accent-gold); }
.insight-section p { font-size: 14px; color: var(--text-secondary); line-height: 1.7; }
.insight-grid { display: grid; gap: 12px; }
.insight-card { padding: 16px; border-radius: var(--radius-md); background: rgba(255,255,255,.02); border: 1px solid var(--border); }
.insight-card h4 { font-size: 14px; font-weight: 700; margin-bottom: 12px; }
.guidance-grid { display: grid; gap: 8px; }
.guidance-grid div { display: flex; justify-content: space-between; font-size: 13px; padding: 4px 0; border-bottom: 1px solid var(--border); }
.guidance-grid span { color: var(--text-secondary); }
.empty-section { text-align: center; padding: 60px; color: var(--text-secondary); }
</style>
