<template>
  <div class="strategies-page">
    <section class="flow-strip glass fade-in-up">
      <div class="flow-copy">
        <p class="flow-kicker">期货策略中心</p>
        <h2>期货策略</h2>
        <p>查看最新发布与进行中的期货策略，并带着完整上下文继续进入深度推演。</p>
      </div>
      <div class="flow-actions">
        <button class="flow-btn" @click="router.push('/forecast-lab')">进入深度推演</button>
      </div>
    </section>

    <section class="section fade-in-up">
      <div class="section-header">
        <h2 class="section-title">期货策略列表</h2>
        <p class="section-subtitle">{{ loading ? "加载中..." : "当前可读的期货策略与执行建议" }}</p>
      </div>
      <div v-if="loadError" class="state-card error-state">
        <strong>期货策略加载失败</strong>
        <p>{{ loadError }}</p>
      </div>
      <div v-else-if="!loading && !strategyList.length" class="state-card empty-state">
        <strong>当前暂无可展示的期货策略</strong>
        <p>稍后再来查看最新发布的期货策略，或先进入深度推演研究中心。</p>
      </div>
      <div v-else class="strategies-grid">
        <article v-for="strategy in strategyList" :key="strategy.id" class="strategy-card glass" @click="router.push('/futures/strategy/' + strategy.id)" style="cursor:pointer">
          <div class="strat-type-badge" :class="strategy.type || 'trend'">{{ strategy.typeLabel || "策略" }}</div>
          <h3 class="strat-name">{{ strategy.name }}</h3>
          <p class="strat-stock">{{ strategy.contract || "" }}</p>
          <div class="strat-details">
            <div class="strat-detail-block">
              <h4>方向</h4>
              <p>{{ strategy.directionText || strategy.direction || "中性" }}</p>
            </div>
            <div class="strat-row">
              <div class="strat-detail-block">
                <h4>风险等级</h4>
                <p :style="{color: strategy.riskLevel === 'HIGH' ? 'var(--negative)' : strategy.riskLevel === 'LOW' ? 'var(--positive)' : 'var(--accent-gold)'}">{{ strategy.riskLabel || strategy.risk_level || "中等" }}</p>
              </div>
              <div class="strat-detail-block">
                <h4>仓位</h4>
                <p style="color:var(--accent-gold);font-weight:600">{{ strategy.positionSize || strategy.position_range || "10%" }}</p>
              </div>
            </div>
            <p class="strat-reason">{{ strategy.reason_summary || strategy.reason || "AI 策略推荐" }}</p>
          </div>
          <div class="strat-actions">
            <button class="btn-primary strat-forecast-btn" @click.stop="goToForecastLab(strategy)">进入深度推演</button>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { listFuturesStrategies } from "@/api/market.js";
import { buildForecastContextQuery } from "@/shared/lib/forecast-context.js";

const router = useRouter();
const loading = ref(false);
const loadError = ref("");
const strategyList = ref([]);

async function loadStrategies() {
  loading.value = true;
  loadError.value = "";
  try {
    const result = await listFuturesStrategies({ page: 1, page_size: 10 });
    const items = Array.isArray(result?.items) ? result.items : [];
    strategyList.value = items.map((s) => ({
      id: s.id,
      name: s.name,
      contract: s.contract,
      type: s.direction === "LONG" ? "breakout" : "trend",
      typeLabel: s.direction === "LONG" ? "做多策略" : s.direction === "SHORT" ? "做空策略" : "中性策略",
      directionText: s.direction === "LONG" ? "做多" : s.direction === "SHORT" ? "做空" : "中性",
      riskLevel: s.risk_level,
      riskLabel: s.risk_level === "HIGH" ? "高风险" : s.risk_level === "LOW" ? "低风险" : "中风险",
      positionSize: s.position_range,
      reason: s.reason_summary,
      reason_summary: s.reason_summary
    }));
  } catch (error) {
    loadError.value = error?.message || "请稍后重试";
    strategyList.value = [];
  } finally {
    loading.value = false;
  }
}

function goToForecastLab(strategy) {
  router.push({
    path: "/forecast-lab",
    query: buildForecastContextQuery({
      targetType: "FUTURES",
      targetId: strategy?.id || "",
      targetKey: strategy?.contract || "",
      targetLabel: strategy?.name || strategy?.contract || "",
      source: "STRATEGY",
      sourceId: strategy?.id || "",
      sourcePath: "/futures/strategies",
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
.flow-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.flow-btn, .btn-primary { display: inline-flex; align-items: center; justify-content: center; padding: 10px 14px; border-radius: var(--radius-full); font-size: 13px; font-weight: 600; cursor: pointer; }
.flow-btn, .btn-primary { border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.state-card { display: grid; gap: 10px; color: var(--text-secondary); }
.state-card strong { color: var(--text-primary); font-size: 18px; }
.strategies-grid { display: grid; gap: 16px; grid-template-columns: repeat(3,1fr); }
.strategy-card { padding: 20px; border-radius: var(--radius-md); }
.strat-type-badge { display: inline-block; padding: 3px 10px; border-radius: var(--radius-full); font-size: 11px; font-weight: 600; margin-bottom: 10px; }
.strat-type-badge.breakout { background: var(--positive-bg); color: var(--positive); }
.strat-type-badge.trend { background: var(--accent-gold-glow); color: var(--accent-gold); }
.strat-name { font-size: 18px; font-weight: 700; margin-bottom: 4px; }
.strat-stock { font-size: 12px; color: var(--text-secondary); margin-bottom: 16px; }
.strat-details { display: grid; gap: 12px; }
.strat-detail-block h4 { font-size: 12px; color: var(--text-muted); margin-bottom: 4px; text-transform: uppercase; letter-spacing: .5px; }
.strat-detail-block p { font-size: 13px; line-height: 1.6; color: var(--text-secondary); }
.strat-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
.strat-reason { font-size: 12px; color: var(--text-muted); line-height: 1.6; padding: 8px; background: rgba(255,255,255,.02); border-radius: 6px; }
.strat-actions { display: grid; gap: 8px; margin-top: 16px; }
.strat-forecast-btn { width: 100%; justify-content: center; display: flex; }
@media (max-width: 1200px) { .strategies-grid { grid-template-columns: 1fr; } }
</style>
