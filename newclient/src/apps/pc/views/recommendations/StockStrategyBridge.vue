<template>
  <div class="bridge-page">
    <section class="flow-strip glass fade-in-up">
      <div class="flow-copy">
        <p class="flow-kicker">机会主线</p>
        <h2>第 2 步：执行策略</h2>
        <p>{{ bridgeSummary }}</p>
      </div>
      <div class="flow-steps">
        <span class="flow-step">1. 每日推荐</span>
        <span class="flow-step active">2. 执行策略</span>
        <span class="flow-step">3. 深度推演</span>
      </div>
      <div class="flow-actions">
        <button class="flow-btn ghost" @click="goToRecommendations">回每日推荐</button>
        <button class="flow-btn" :disabled="!canLaunchForecast" @click="goToForecastLab">进入深度推演</button>
      </div>
    </section>

    <section v-if="!hasContext" class="section glass fade-in-up">
      <div class="empty-state">
        <h3>请先从每日推荐进入策略承接页</h3>
        <p>这个页面只承接当前推荐标的的执行策略，不再展示无关的期货策略列表。</p>
        <button class="flow-btn" @click="goToRecommendations">返回每日推荐</button>
      </div>
    </section>

    <template v-else>
      <section class="section fade-in-up">
        <div class="section-header">
          <h2 class="section-title">策略结论</h2>
          <p class="section-subtitle">{{ loading ? "加载中..." : `${selectedName || selectedSymbol} 的执行结论与风险边界` }}</p>
        </div>
        <div v-if="loadError" class="error-state">
          <strong>策略数据加载失败</strong>
          <p>{{ loadError }}</p>
        </div>
        <div v-else-if="strategyCard" class="strategy-grid">
          <article class="strategy-card glass">
            <div class="strategy-head">
              <div>
                <p class="strategy-kicker">推荐标的</p>
                <h3>{{ strategyCard.name }}</h3>
                <p class="strategy-symbol">{{ strategyCard.symbol }}</p>
              </div>
              <div class="score-chip">{{ strategyCard.scoreLabel }}</div>
            </div>
            <div class="strategy-metrics">
              <div class="metric-item">
                <span>风险等级</span>
                <strong>{{ strategyCard.riskLevel }}</strong>
              </div>
              <div class="metric-item">
                <span>建议仓位</span>
                <strong>{{ strategyCard.positionRange }}</strong>
              </div>
              <div class="metric-item">
                <span>止盈</span>
                <strong class="positive">{{ strategyCard.takeProfit }}</strong>
              </div>
              <div class="metric-item">
                <span>止损</span>
                <strong class="negative">{{ strategyCard.stopLoss }}</strong>
              </div>
            </div>
            <div class="strategy-copy">
              <h4>策略理由</h4>
              <p>{{ strategyCard.reason }}</p>
            </div>
            <div class="strategy-copy" v-if="strategyCard.riskNote">
              <h4>风险备注</h4>
              <p>{{ strategyCard.riskNote }}</p>
            </div>
          </article>

          <article class="strategy-card glass">
            <div class="strategy-copy">
              <h4>执行计划</h4>
              <p>{{ executionPlanSummary }}</p>
            </div>
            <div class="execution-actions">
              <div class="execution-item">
                <span>建议动作</span>
                <strong>{{ executionPlanAction }}</strong>
              </div>
              <div class="execution-item">
                <span>去股票分析</span>
                <button class="mini-btn" @click="goToAnalysis">查看分析</button>
              </div>
              <div class="execution-item">
                <span>去深度推演</span>
                <button class="mini-btn primary" :disabled="!canLaunchForecast" @click="goToForecastLab">继续推进</button>
              </div>
            </div>
          </article>
        </div>
        <div v-else class="loading-state">正在加载策略承接信息...</div>
      </section>

      <section v-if="dimensionCards.length" class="section fade-in-up">
        <div class="section-header">
          <h2 class="section-title">证据维度</h2>
          <p class="section-subtitle">把推荐明细转成执行策略时，我们重点看的四个维度。</p>
        </div>
        <div class="dimension-grid">
          <article v-for="card in dimensionCards" :key="card.key" class="dimension-card glass">
            <div class="dimension-head">
              <h3>{{ card.title }}</h3>
              <span>{{ card.score }}</span>
            </div>
            <p>{{ card.summary }}</p>
          </article>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getStockRecommendationDetail, getStockRecommendationInsight } from "@/api/market.js";
import { buildForecastContextQuery } from "@/shared/lib/forecast-context.js";

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const loadError = ref("");
const recommendationDetail = ref(null);
const recommendationInsight = ref(null);

const selectedRecoId = computed(() => String(route.query.reco_id || "").trim());
const selectedSymbol = computed(() => String(route.query.symbol || "").trim().toUpperCase());
const selectedName = computed(() => String(route.query.name || "").trim());
const hasContext = computed(() => Boolean(selectedRecoId.value));
const canLaunchForecast = computed(() => hasContext.value && Boolean(selectedSymbol.value));

const bridgeSummary = computed(() => {
  if (!hasContext.value) {
    return "先从每日推荐锁定一只真实推荐标的，再在这里收敛执行计划，然后进入深度推演。";
  }
  return `你当前正在承接 ${selectedName.value || selectedSymbol.value} 的推荐机会，这里只展示这只股票的执行策略，不再混入期货策略列表。`;
});

const strategyCard = computed(() => {
  const detail = recommendationDetail.value;
  const insight = recommendationInsight.value;
  if (!detail && !insight) return null;
  const recommendation = insight?.recommendation || {};
  const detailInfo = insight?.detail || detail || {};
  return {
    name: recommendation.name || selectedName.value || selectedSymbol.value,
    symbol: recommendation.symbol || selectedSymbol.value,
    scoreLabel: recommendation.score ? `AI ${Math.round(recommendation.score)}` : "AI --",
    riskLevel: recommendation.risk_level || "待评估",
    positionRange: recommendation.position_range || "待确认",
    takeProfit: detailInfo.take_profit || "待确认",
    stopLoss: detailInfo.stop_loss || "待确认",
    riskNote: detailInfo.risk_note || "",
    reason: recommendation.reason_summary || insight?.explanation?.seed_summary || "系统会结合推荐详情与研究摘要给出执行策略。"
  };
});

const executionPlanSummary = computed(() => {
  const explanation = recommendationInsight.value?.explanation;
  return (
    explanation?.confidence_reason ||
    explanation?.seed_summary ||
    recommendationInsight.value?.recommendation?.reason_summary ||
    "当前更适合结合推荐详情、研究摘要与风险边界分批推进，先确认信号，再进入深度推演做最终复核。"
  );
});

const executionPlanAction = computed(() => {
  const riskLevel = strategyCard.value?.riskLevel || "";
  if (String(riskLevel).includes("高")) return "先观察，谨慎试探";
  if (String(riskLevel).includes("低")) return "分批跟随";
  return "等待确认后推进";
});

const dimensionCards = computed(() => {
  const detail = recommendationInsight.value?.detail || recommendationDetail.value;
  if (!detail) return [];
  const toScore = (value) => (typeof value === "number" ? Math.round(value) : "--");
  return [
    {
      key: "tech",
      title: "技术面",
      score: toScore(detail.tech_score),
      summary: "决定当前是否适合按趋势推进，重点看形态、节奏与确认强度。"
    },
    {
      key: "fund",
      title: "基本面代理",
      score: toScore(detail.fund_score),
      summary: "用推荐评分框架里的基本面代理信息，辅助判断这次机会是否有中期支撑。"
    },
    {
      key: "flow",
      title: "资金面",
      score: toScore(detail.money_flow_score),
      summary: "决定这次执行更适合试探还是加仓，重点看资金是否支持持续推进。"
    },
    {
      key: "sentiment",
      title: "情绪面",
      score: toScore(detail.sentiment_score),
      summary: "用于判断当前热度和共识度，避免在过热或过冷位置做错误动作。"
    }
  ];
});

async function loadBridgeData() {
  if (!hasContext.value) return;
  loading.value = true;
  loadError.value = "";
  try {
    const [detail, insight] = await Promise.all([
      getStockRecommendationDetail(selectedRecoId.value),
      getStockRecommendationInsight(selectedRecoId.value)
    ]);
    recommendationDetail.value = detail || null;
    recommendationInsight.value = insight || null;
  } catch (error) {
    loadError.value = error?.message || "请稍后重试";
  } finally {
    loading.value = false;
  }
}

function goToRecommendations() {
  router.push({
    path: "/recommendations",
    query: selectedSymbol.value
      ? { symbol: selectedSymbol.value, name: selectedName.value || "", from: "strategies" }
      : { from: "strategies" }
  });
}

function goToAnalysis() {
  if (selectedSymbol.value) {
    router.push(`/identify/${encodeURIComponent(selectedSymbol.value)}`);
    return;
  }
  router.push("/identify");
}

function goToForecastLab() {
  if (!canLaunchForecast.value) return;
  router.push({
    path: "/forecast-lab",
    query: buildForecastContextQuery({
      targetType: "STOCK",
      targetId: selectedRecoId.value,
      targetKey: selectedSymbol.value,
      targetLabel: selectedName.value || selectedSymbol.value,
      source: "STRATEGY",
      sourceId: selectedRecoId.value,
      sourcePath: "/recommendations/strategies",
      from: "strategies",
      strategyId: selectedRecoId.value
    })
  });
}

watch(
  () => selectedRecoId.value,
  () => {
    recommendationDetail.value = null;
    recommendationInsight.value = null;
    loadError.value = "";
    if (hasContext.value) loadBridgeData();
  }
);

onMounted(() => {
  if (hasContext.value) loadBridgeData();
});
</script>

<style scoped>
.bridge-page { display: grid; gap: 20px; max-width: 1400px; }
.flow-strip { display: grid; gap: 14px; padding: 18px 20px; border-radius: var(--radius-lg); border: 1px solid var(--border); background: linear-gradient(135deg, rgba(240,185,11,.08), rgba(255,255,255,.02)); }
.flow-copy h2 { font-size: 22px; font-weight: 800; margin: 4px 0 6px; }
.flow-copy p { color: var(--text-secondary); line-height: 1.7; }
.flow-kicker { font-size: 11px; color: var(--accent-gold); text-transform: uppercase; letter-spacing: .08em; }
.flow-steps { display: flex; gap: 8px; flex-wrap: wrap; }
.flow-step { padding: 8px 12px; border-radius: var(--radius-full); border: 1px solid var(--border); color: var(--text-secondary); font-size: 12px; }
.flow-step.active { border-color: var(--border-gold); color: var(--accent-gold); background: var(--accent-gold-glow); }
.flow-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.flow-btn { display: inline-flex; align-items: center; justify-content: center; padding: 10px 14px; border-radius: var(--radius-full); font-size: 13px; font-weight: 600; cursor: pointer; border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.flow-btn.ghost { border: 1px solid var(--border); background: none; color: var(--text-secondary); }
.flow-btn:disabled { opacity: .5; cursor: not-allowed; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.empty-state, .loading-state, .error-state { display: grid; gap: 10px; color: var(--text-secondary); }
.empty-state h3, .error-state strong { color: var(--text-primary); font-size: 18px; }
.strategy-grid { display: grid; gap: 16px; grid-template-columns: repeat(2, minmax(0, 1fr)); }
.strategy-card { display: grid; gap: 16px; padding: 20px; border-radius: var(--radius-md); }
.strategy-head { display: flex; justify-content: space-between; gap: 12px; align-items: flex-start; }
.strategy-kicker { font-size: 11px; color: var(--accent-gold); text-transform: uppercase; letter-spacing: .08em; margin-bottom: 4px; }
.strategy-head h3 { font-size: 22px; font-weight: 800; margin: 0 0 4px; }
.strategy-symbol { color: var(--text-secondary); font-size: 13px; }
.score-chip { padding: 8px 12px; border-radius: var(--radius-full); background: rgba(240,185,11,.08); color: var(--accent-gold); font-weight: 700; }
.strategy-metrics { display: grid; gap: 10px; grid-template-columns: repeat(2, minmax(0, 1fr)); }
.metric-item { padding: 12px 14px; border: 1px solid var(--border); border-radius: var(--radius-sm); background: rgba(255,255,255,.02); display: grid; gap: 6px; }
.metric-item span { color: var(--text-secondary); font-size: 12px; }
.metric-item strong { font-size: 15px; }
.positive { color: var(--positive); }
.negative { color: var(--negative); }
.strategy-copy { display: grid; gap: 8px; }
.strategy-copy h4 { font-size: 14px; color: var(--accent-gold); font-weight: 700; }
.strategy-copy p { color: var(--text-secondary); line-height: 1.7; }
.execution-actions { display: grid; gap: 12px; }
.execution-item { display: flex; justify-content: space-between; align-items: center; gap: 12px; padding: 12px 14px; border-radius: var(--radius-sm); background: rgba(255,255,255,.02); border: 1px solid var(--border); }
.execution-item span { color: var(--text-secondary); font-size: 13px; }
.mini-btn { border: 1px solid var(--border); background: none; color: var(--text-primary); border-radius: var(--radius-full); padding: 8px 12px; cursor: pointer; }
.mini-btn.primary { border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-weight: 700; }
.dimension-grid { display: grid; gap: 16px; grid-template-columns: repeat(4, minmax(0, 1fr)); }
.dimension-card { padding: 18px; border-radius: var(--radius-md); display: grid; gap: 10px; }
.dimension-head { display: flex; justify-content: space-between; align-items: center; gap: 8px; }
.dimension-head h3 { font-size: 16px; font-weight: 700; }
.dimension-head span { color: var(--accent-gold); font-weight: 700; }
.dimension-card p { color: var(--text-secondary); line-height: 1.7; font-size: 13px; }
@media (max-width: 1200px) {
  .strategy-grid,
  .dimension-grid { grid-template-columns: 1fr; }
}
</style>
