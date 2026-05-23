<template>
  <div class="analysis-page">
    <!-- Search -->
    <section class="search-section glass fade-in-up">
      <div class="search-section-content">
        <h1 class="search-title">AI 股票分析</h1>
        <p class="search-desc">输入股票代码或名称，AI 将从4个维度进行综合分析</p>
        <div class="search-form">
          <div class="search-input-wrap">
            <input v-model="searchQuery" type="text" placeholder="输入股票代码，如 300750" @keyup.enter="searchStock" />
            <button class="search-clear" v-if="searchQuery" @click="searchQuery = ''">×</button>
          </div>
          <button class="btn-primary" @click="searchStock" :disabled="!searchQuery.trim() || analyzing">
            🤖 AI 分析
          </button>
        </div>
        <div class="hot-searches">
          <span class="hot-label">热门：</span>
          <button v-for="hot in hotStocks" :key="hot.symbol" class="hot-chip" @click="quickAnalyze(hot)">{{ hot.name }}</button>
        </div>
      </div>
    </section>

    <!-- Loading -->
    <section v-if="analyzing" class="section loading-section fade-in-up">
      <div class="loading-content">
        <div class="loading-spinner">
          <svg viewBox="0 0 48 48" width="48" height="48"><circle cx="24" cy="24" r="20" fill="none" stroke="var(--accent-gold)" stroke-width="3" stroke-dasharray="125.6" stroke-dashoffset="0" class="spinner-ring"/></svg>
        </div>
        <h3>AI 正在分析...</h3>
        <p>正在调用大模型进行多维度数据分析</p>
      </div>
    </section>

    <!-- Results -->
    <template v-if="showResult && insight">
      <!-- Stock Header -->
      <div class="stock-header glass fade-in-up">
        <div class="stock-header-main">
          <div>
            <div class="stock-ticker">
              <span class="stock-symbol">{{ reco?.symbol || symbol }}</span>
              <span class="tag tag-gold">AI 分析</span>
              <span class="tag" :class="reco?.risk_level === 'HIGH' ? 'tag-red' : reco?.risk_level === 'LOW' ? 'tag-green' : 'tag-neutral'">{{ riskLabel(reco?.risk_level) }}</span>
            </div>
            <h1 class="stock-name">{{ reco?.name || symbol }}</h1>
          </div>
          <div class="stock-actions">
            <button class="btn-primary" @click="reset">重新分析</button>
          </div>
        </div>
        <p class="stock-reason" v-if="reco?.reason_summary">{{ reco.reason_summary }}</p>
      </div>

      <!-- Score Dashboard: 4 Dimensions -->
      <section class="section fade-in-up fade-in-up-delay-1">
        <div class="section-header">
          <h2 class="section-title">AI 综合评分</h2>
          <span class="score-method">{{ scoreMethod }}</span>
        </div>
        <div class="score-dashboard">
          <div class="score-circle-section">
            <div class="score-circle-wrap">
              <svg viewBox="0 0 120 120" class="score-ring-big">
                <circle cx="60" cy="60" r="52" fill="none" stroke="rgba(255,255,255,.06)" stroke-width="6"/>
                <circle cx="60" cy="60" r="52" fill="none" stroke="url(#scoreGrad)" stroke-width="6" :stroke-dasharray="326.7" :stroke-dashoffset="326.7 - 326.7 * totalScore / 100" stroke-linecap="round"/>
              </svg>
              <div class="score-center">
                <span class="score-number">{{ totalScore }}</span>
                <span class="score-total">/100</span>
              </div>
            </div>
            <div class="score-weights">
              <span v-for="f in factors" :key="f.key" class="weight-chip">
                {{ f.label }} <strong>{{ (f.weight * 100).toFixed(0) }}%</strong>
              </span>
            </div>
          </div>

          <div class="dimension-bars">
            <div v-for="dim in factors" :key="dim.key" class="dim-item">
              <div class="dim-header">
                <span class="dim-label">{{ dim.label }}</span>
                <div class="dim-scores">
                  <span class="dim-raw">{{ dim.score.toFixed(1) }}</span>
                  <span class="dim-contribution">+{{ dim.contribution.toFixed(1) }}</span>
                </div>
              </div>
              <div class="dim-bar-bg">
                <div class="dim-bar" :style="{ width: dim.score + '%', background: barColor(dim.score) }"></div>
              </div>
              <div class="dim-weight">权重 {{ (dim.weight * 100).toFixed(0) }}% · 贡献 {{ dim.contribution.toFixed(1) }} 分</div>
            </div>
          </div>
        </div>
        <div class="score-footer" v-if="insight?.ScoreFramework">
          <span>总分: {{ totalScore }} | 加权分: {{ weightedScore }} | 分差: {{ scoreGap >= 0 ? '+' : '' }}{{ scoreGap?.toFixed(1) }}</span>
          <span class="score-method-label">算法: {{ scoreMethod }}</span>
        </div>
      </section>

      <!-- Agent Opinions & Scenarios -->
      <div class="analysis-grid-2col fade-in-up fade-in-up-delay-2">
        <section class="section" v-if="agentOpinions.length">
          <div class="section-header"><h2 class="section-title">AI 分析师意见</h2></div>
          <DeepForecastSummaryCard
            v-if="forecastEntryVisible"
            :summary="forecastEntrySummary"
            :to="forecastEntryTo"
            mode="pc"
            heading="深度推演"
            style="margin-bottom: 12px;"
          />
          <button
            v-if="showResult && (reco?.symbol || symbol)"
            class="deep-forecast-lab-btn"
            @click="router.push(forecastLabEntryTo)"
          >
            带着完整上下文进入深度推演
          </button>
          <div class="agent-list">
            <div v-for="agent in agentOpinions" :key="agent.role" class="agent-card">
              <div class="agent-head">
                <span class="agent-icon">🧠</span>
                <div>
                  <strong>{{ agent.role }}分析师</strong>
                  <span class="agent-stance" :class="agent.stance?.toLowerCase()">{{ stanceLabel(agent.stance) }}</span>
                </div>
                <span class="agent-confidence">置信 {{ (agent.confidence * 100).toFixed(0) }}%</span>
              </div>
              <p>{{ agent.summary }}</p>
              <span v-if="agent.veto" class="agent-veto">⚠️ 该分析师行使了否决权</span>
            </div>
          </div>
        </section>

        <section class="section" v-if="scenarios.length">
          <div class="section-header"><h2 class="section-title">场景推演</h2></div>
          <div class="scenario-list">
            <div v-for="s in scenarios" :key="s.scenario" class="scenario-card" :class="'scenario-' + s.scenario">
              <div class="scenario-head">
                <span class="scenario-icon">{{ s.scenario === 'bull' ? '📈' : s.scenario === 'bear' ? '📉' : '📊' }}</span>
                <strong class="scenario-name">{{ scenarioLabel(s.scenario) }}</strong>
                <span class="scenario-conf">信 {{ (s.confidence * 100).toFixed(0) }}%</span>
              </div>
              <p>{{ s.thesis }}</p>
              <div class="scenario-detail">
                <span>触发: {{ s.trigger }}</span>
                <span>建议: {{ s.action_suggestion }}</span>
              </div>
            </div>
          </div>
        </section>
      </div>

      <!-- Confidence & Risk -->
      <div class="analysis-grid-2col fade-in-up fade-in-up-delay-3">
        <section class="section" v-if="calibration">
          <div class="section-header"><h2 class="section-title">信心校准</h2></div>
          <div class="calibration-display">
            <div class="cal-bar">
              <span>基础</span>
              <div class="cal-bar-bg"><div class="cal-bar-fill" :style="{ width: (calibration.base_confidence * 100) + '%' }"></div></div>
              <span>{{ (calibration.base_confidence * 100).toFixed(0) }}%</span>
            </div>
            <div class="cal-bar">
              <span>调整后</span>
              <div class="cal-bar-bg"><div class="cal-bar-fill gold" :style="{ width: (calibration.adjusted_confidence * 100) + '%' }"></div></div>
              <span style="color:var(--accent-gold)">{{ (calibration.adjusted_confidence * 100).toFixed(0) }}%</span>
            </div>
            <div class="cal-drivers" v-if="calibration.drivers?.length">
              <div v-for="d in calibration.drivers" :key="d" class="cal-driver">• {{ d }}</div>
            </div>
          </div>
        </section>

        <section class="section">
          <div class="section-header"><h2 class="section-title">风险提示</h2></div>
          <div class="risk-list">
            <div v-for="flag in riskFlags" :key="flag" class="risk-item">⚠ {{ flag }}</div>
            <div v-if="!riskFlags.length" class="no-risk">暂无风险提示</div>
          </div>
          <div class="invalidation-section" v-if="invalidations.length">
            <h4>失效条件</h4>
            <div v-for="inv in invalidations" :key="inv" class="inv-item">• {{ inv }}</div>
          </div>
        </section>
      </div>

      <!-- Analysis Report -->
      <section class="section fade-in-up fade-in-up-delay-4" v-if="reportSections.length">
        <div class="section-header"><h2 class="section-title">分析报告</h2></div>
        <div class="report-body">
          <div v-for="(sec, i) in reportSections" :key="i" class="report-segment">
            <h3>{{ sec.title }}</h3>
            <p>{{ sec.content }}</p>
          </div>
          <div class="report-footer">
            <span class="tag tag-gold">AI 生成内容仅供参考</span>
            <span class="report-date">生成时间: {{ generatedAt }}</span>
          </div>
        </div>
      </section>
    </template>

    <!-- Empty -->
    <section v-if="!analyzing && !showResult" class="section empty-section fade-in-up">
      <div class="empty-content">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="48" height="48" style="color:var(--text-muted)"><circle cx="11" cy="11" r="7"/><path d="M16.5 16.5L21 21"/></svg>
        <h3>输入股票开始分析</h3>
        <p>输入股票代码或名称，AI 从技术、基本面、情绪、资金流 4 个维度综合评分</p>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { STOCK_MASTER } from "@/mock/stocks.js";
import { getStockRecommendationInsight, listStockRecommendations } from "@/api/market.js";
import DeepForecastSummaryCard from "@/shared/components/deep-forecast/DeepForecastSummaryCard.vue";
import { useDeepForecastEntry } from "@/shared/composables/useDeepForecastEntry.js";
import { buildForecastContextQuery } from "@/shared/lib/forecast-context.js";

const route = useRoute();
const router = useRouter();
const searchQuery = ref("");
const analyzing = ref(false);
const showResult = ref(false);
const insight = ref(null);
const symbol = ref("");

const hotStocks = [
  { symbol: '300750.SZ', name: '宁德时代' },
  { symbol: '600941.SH', name: '中国移动' },
  { symbol: '002415.SZ', name: '海康威视' },
  { symbol: '688981.SH', name: '中芯国际' }
];

// Computed fields from insight
const reco = computed(() => insight.value?.Recommendation || null);
const factors = computed(() => {
  const sf = insight.value?.ScoreFramework;
  if (sf?.factors?.length) return sf.factors;
  return [];
});
const totalScore = computed(() => {
  const sf = insight.value?.ScoreFramework;
  return sf?.total_score ? sf.total_score.toFixed(1) : sf?.weighted_score?.toFixed(1) || '--';
});
const weightedScore = computed(() => insight.value?.ScoreFramework?.weighted_score?.toFixed(1) || '--');
const scoreGap = computed(() => insight.value?.ScoreFramework?.score_gap || 0);
const scoreMethod = computed(() => insight.value?.ScoreFramework?.method || 'multi-factor');

const agentOpinions = computed(() => insight.value?.Explanation?.agent_opinions || []);
const scenarios = computed(() => insight.value?.Explanation?.scenario_snapshots || []);
const calibration = computed(() => insight.value?.Explanation?.confidence_calibration || null);
const riskFlags = computed(() => insight.value?.Explanation?.risk_flags || []);
const invalidations = computed(() => insight.value?.Explanation?.invalidations || []);
const generatedAt = computed(() => {
  if (insight.value?.GeneratedAt) return insight.value.GeneratedAt.replace('T', ' ').slice(0, 16);
  return new Date().toLocaleString("zh-CN", { hour12: false });
});

const reportSections = computed(() => {
  const expl = insight.value?.Explanation;
  if (!expl) return [];
  return [
    { title: '核心逻辑', content: expl.seed_summary || 'AI 多维度分析完成' },
    { title: '技术面分析', content: expl.graph_summary || '技术指标分析中' },
    { title: '资金与情绪', content: expl.consensus_summary || '资金面情绪分析中' },
    { title: '风险提示', content: (expl.risk_flags || ['常规波动需控仓']).join('；') }
  ];
});
const forecastEntrySource = computed(() => insight.value?.Explanation || null);
const {
  summary: forecastEntrySummary,
  to: forecastEntryTo,
  visible: forecastEntryVisible
} = useDeepForecastEntry(forecastEntrySource, { mode: "pc" });
const forecastLabEntryTo = computed(() => {
  const recommendation = reco.value || {};
  const targetKey = recommendation.symbol || symbol.value;
  const targetLabel = recommendation.name || targetKey;
  return {
    path: "/forecast-lab",
    query: buildForecastContextQuery({
      targetType: "STOCK",
      targetId: recommendation.id || recommendation.recommendation_id || recommendation.reco_id || "",
      targetKey,
      targetLabel,
      source: "IDENTIFY",
      sourceId: recommendation.id || recommendation.recommendation_id || recommendation.reco_id || targetKey,
      sourcePath: targetKey ? `/identify/${targetKey}` : "/identify",
      from: "identify"
    })
  };
});

function riskLabel(r) { const m = { HIGH:'高风险', MEDIUM:'中风险', LOW:'低风险' }; return m[r] || r; }
function stanceLabel(s) {
  if (!s) return '观望';
  const u = s.toUpperCase();
  if (u === 'SUPPORT' || u === 'BULLISH') return '看多';
  if (u === 'WATCH' || u === 'NEUTRAL') return '观望';
  if (u === 'BEARISH') return '看空';
  return s;
}
function scenarioLabel(s) {
  if (s === 'bull') return '乐观场景';
  if (s === 'base') return '基准场景';
  if (s === 'bear') return '悲观场景';
  return s;
}
function barColor(score) {
  if (score >= 80) return 'var(--positive)';
  if (score >= 60) return 'var(--accent-gold)';
  return 'var(--negative)';
}

onMounted(() => {
  if (route.params.symbol) {
    searchQuery.value = route.params.symbol;
    analyzeStock(route.params.symbol);
  }
});

function searchStock() {
  const q = searchQuery.value.trim();
  if (q) analyzeStock(q);
}

function quickAnalyze(hot) {
  searchQuery.value = hot.symbol;
  analyzeStock(hot.symbol);
}

function reset() { showResult.value = false; analyzing.value = false; insight.value = null; }

async function analyzeStock(input) {
  symbol.value = input;
  analyzing.value = true;
  showResult.value = false;

  // Try to find the recommendation first
  let recoId = null;
  try {
    const list = await listStockRecommendations({ page: 1, page_size: 20 });
    const match = list?.items?.find(r =>
      r.symbol?.includes(input) || r.name?.includes(input)
    );
    if (match) recoId = match.id;
  } catch { /* continue */ }

  // Get insight
  if (recoId) {
    try {
      const result = await getStockRecommendationInsight(recoId);
      if (result) {
        insight.value = result;
        analyzing.value = false;
        showResult.value = true;
        return;
      }
    } catch { /* fall through */ }
  }

  // Mock fallback
  setTimeout(() => {
    insight.value = {
      Recommendation: {
        symbol: input.split('.')[0] || input, name: input,
        risk_level: 'MEDIUM', reason_summary: 'AI 多维度筛选推荐',
        score: 82, strategy_version: 'stock-l2-v1'
      },
      ScoreFramework: {
        method: 'growth-v1 (tech30 + fund30 + sentiment20 + flow20)',
        total_score: 82.5, weighted_score: 78.3, score_gap: 4.2,
        factors: [
          { key: 'tech', label: '技术因子', weight: 0.30, score: 84, contribution: 25.2 },
          { key: 'fund', label: '基本面因子', weight: 0.30, score: 78, contribution: 23.4 },
          { key: 'sentiment', label: '情绪因子', weight: 0.20, score: 88, contribution: 17.6 },
          { key: 'flow', label: '资金流因子', weight: 0.20, score: 76, contribution: 15.2 }
        ]
      },
      Explanation: {
        seed_summary: `${input} 基本面稳健，AI 从4个维度进行了综合评估。`,
        graph_summary: '技术面呈现震荡整理格局，短期方向尚不明确。',
        consensus_summary: '资金与情绪指标偏中性，等待明确信号。',
        risk_flags: ['市场整体波动风险', '行业政策变化风险'],
        invalidations: ['跌破关键支撑位'],
        agent_opinions: [
          { role: 'FLOW', stance: 'SUPPORT', confidence: 0.68, summary: '资金面表现中性偏积极', veto: false },
          { role: 'THEME', stance: 'WATCH', confidence: 0.61, summary: '主题概念热度一般', veto: false },
          { role: 'RISK', stance: 'WATCH', confidence: 0.73, summary: '风险可接受，注意仓位管理', veto: false }
        ],
        scenario_snapshots: [
          { scenario: 'bull', thesis: '趋势延续', trigger: '量价共振', action_suggestion: '顺势跟踪', confidence: 0.72 },
          { scenario: 'base', thesis: '核心逻辑维持', trigger: '常规波动', action_suggestion: '按计划执行', confidence: 0.64 },
          { scenario: 'bear', thesis: '风险边界被触发', trigger: '消息扰动', action_suggestion: '收缩风险暴露', confidence: 0.38 }
        ],
        confidence_calibration: { base_confidence: 0.52, adjusted_confidence: 0.65, drivers: ['事件佐证', '回撤校准'], advisory_only: true }
      },
      GeneratedAt: new Date().toISOString()
    };
    analyzing.value = false;
    showResult.value = true;
  }, 800);
}
</script>

<style scoped>
.analysis-page { display: grid; gap: 20px; max-width: 1400px; }
.search-section { padding: 28px; border-radius: var(--radius-lg); }
.search-title { font-size: 22px; font-weight: 700; margin-bottom: 6px; }
.search-desc { font-size: 13px; color: var(--text-secondary); margin-bottom: 16px; }
.search-form { display: flex; gap: 10px; margin-bottom: 12px; }
.search-input-wrap { flex: 1; position: relative; }
.search-input-wrap input { width: 100%; padding: 12px 16px; border-radius: var(--radius-md); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 14px; }
.search-input-wrap input:focus { border-color: var(--accent-gold); }
.search-clear { position: absolute; right: 10px; top: 50%; transform: translateY(-50%); color: var(--text-muted); font-size: 18px; }
.btn-primary { display: inline-flex; align-items: center; gap: 6px; padding: 12px 20px; border-radius: var(--radius-md); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-weight: 600; font-size: 14px; border: none; cursor: pointer; }
.btn-primary:disabled { opacity: .5; }
.hot-searches { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.hot-label { font-size: 12px; color: var(--text-muted); }
.hot-chip { padding: 5px 12px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 12px; color: var(--text-secondary); cursor: pointer; }
.hot-chip:hover { border-color: var(--border-gold); color: var(--accent-gold); }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { display: flex; align-items: baseline; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.section-title { font-size: 20px; font-weight: 700; }

/* Loading */
.loading-section { text-align: center; }
.loading-content { padding: 40px 20px; }
.loading-spinner { margin: 0 auto 20px; width: 48px; height: 48px; }
.spinner-ring { animation: spin 1.2s linear infinite; transform-origin: center; }
@keyframes spin { 100% { transform: rotate(360deg); } }
.loading-content h3 { font-size: 18px; margin-bottom: 8px; color: var(--accent-gold); }
.loading-content p { font-size: 13px; color: var(--text-secondary); }

/* Stock Header */
.stock-header { padding: 20px; border-radius: var(--radius-lg); }
.stock-header-main { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 8px; }
.stock-ticker { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.stock-symbol { font-size: 13px; color: var(--text-secondary); font-family: var(--font-mono); }
.stock-name { font-size: 28px; font-weight: 800; }
.stock-reason { font-size: 14px; color: var(--text-secondary); line-height: 1.6; }
.stock-actions { display: flex; gap: 8px; }

/* Score Dashboard */
.score-dashboard { display: flex; gap: 24px; align-items: flex-start; flex-wrap: wrap; }
.score-circle-section { text-align: center; }
.score-circle-wrap { position: relative; width: 120px; height: 120px; margin: 0 auto; }
.score-ring-big { width: 120px; height: 120px; }
.score-center { position: absolute; inset: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; }
.score-number { font-size: 36px; font-weight: 800; color: var(--accent-gold); line-height: 1; }
.score-total { font-size: 11px; color: var(--text-secondary); }
.score-weights { display: flex; flex-wrap: wrap; gap: 4px; justify-content: center; margin-top: 8px; }
.weight-chip { font-size: 10px; padding: 2px 6px; border-radius: 4px; background: rgba(255,255,255,.04); color: var(--text-muted); }
.weight-chip strong { color: var(--accent-gold); }
.score-method { font-size: 11px; color: var(--text-muted); font-family: var(--font-mono); }
.dimension-bars { flex: 1; min-width: 280px; display: grid; gap: 14px; }
.dim-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 4px; }
.dim-label { font-size: 14px; font-weight: 600; }
.dim-scores { display: flex; gap: 8px; align-items: baseline; }
.dim-raw { font-size: 18px; font-weight: 700; }
.dim-contribution { font-size: 12px; color: var(--text-muted); }
.dim-bar-bg { height: 8px; background: rgba(255,255,255,.06); border-radius: 4px; overflow: hidden; }
.dim-bar { height: 100%; border-radius: 4px; transition: width .6s; }
.dim-weight { font-size: 11px; color: var(--text-muted); margin-top: 2px; }
.score-footer { margin-top: 12px; display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px; font-size: 12px; color: var(--text-muted); }
.score-method-label { font-family: var(--font-mono); font-size: 11px; }

/* Grid */
.analysis-grid-2col { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; }

/* Agents */
.deep-forecast-lab-btn { width: 100%; margin-bottom: 12px; padding: 12px 16px; border-radius: var(--radius-full); border: 1px solid var(--border-gold); background: rgba(240,185,11,.08); color: var(--accent-gold); font-size: 13px; font-weight: 700; cursor: pointer; }
.agent-list { display: grid; gap: 8px; }
.agent-card { padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border); }
.agent-head { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.agent-icon { font-size: 16px; }
.agent-head strong { font-size: 13px; }
.agent-stance { font-size: 11px; padding: 2px 8px; border-radius: var(--radius-full); font-weight: 600; }
.agent-stance.support, .agent-stance.bullish { background: var(--positive-bg); color: var(--positive); }
.agent-stance.watch, .agent-stance.neutral { background: rgba(139,147,176,.1); color: var(--text-secondary); }
.agent-stance.bearish { background: var(--negative-bg); color: var(--negative); }
.agent-confidence { margin-left: auto; font-size: 11px; color: var(--text-muted); }
.agent-card p { font-size: 12px; color: var(--text-secondary); line-height: 1.5; }
.agent-veto { font-size: 11px; color: var(--negative); margin-top: 4px; display: block; }

/* Scenarios */
.scenario-list { display: grid; gap: 8px; }
.scenario-card { padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border); }
.scenario-head { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.scenario-icon { font-size: 16px; }
.scenario-name { font-size: 13px; font-weight: 600; }
.scenario-conf { margin-left: auto; font-size: 11px; color: var(--accent-gold); }
.scenario-card p { font-size: 12px; color: var(--text-secondary); line-height: 1.5; margin-bottom: 6px; }
.scenario-detail { display: flex; gap: 12px; font-size: 11px; color: var(--text-muted); flex-wrap: wrap; }

/* Confidence */
.calibration-display { display: grid; gap: 10px; }
.cal-bar { display: flex; align-items: center; gap: 10px; font-size: 13px; }
.cal-bar-bg { flex: 1; height: 6px; background: rgba(255,255,255,.06); border-radius: 3px; overflow: hidden; }
.cal-bar-fill { height: 100%; border-radius: 3px; background: var(--text-muted); }
.cal-bar-fill.gold { background: var(--accent-gold); }
.cal-drivers { padding: 8px; border-radius: var(--radius-sm); background: rgba(255,255,255,.02); font-size: 12px; color: var(--text-secondary); display: grid; gap: 4px; }

/* Risk */
.risk-list { display: grid; gap: 6px; }
.risk-item { font-size: 13px; color: var(--text-secondary); line-height: 1.6; }
.no-risk { font-size: 13px; color: var(--text-muted); }
.invalidation-section h4 { font-size: 14px; font-weight: 700; margin: 12px 0 8px; color: var(--negative); }
.inv-item { font-size: 13px; color: var(--text-secondary); line-height: 1.6; }

/* Report */
.report-body { display: grid; gap: 16px; }
.report-segment h3 { font-size: 15px; font-weight: 700; margin-bottom: 8px; color: var(--accent-gold); }
.report-segment p { font-size: 14px; color: var(--text-secondary); line-height: 1.8; }
.report-footer { display: flex; justify-content: space-between; align-items: center; padding-top: 16px; border-top: 1px solid var(--border); }
.report-date { font-size: 11px; color: var(--text-muted); }

/* Empty */
.empty-section { text-align: center; padding: 60px 20px; }
.empty-content svg { margin-bottom: 16px; }
.empty-content h3 { font-size: 18px; margin-bottom: 8px; }
.empty-content p { font-size: 13px; color: var(--text-secondary); max-width: 450px; margin: 0 auto; }

@media (max-width: 900px) { .analysis-grid-2col { grid-template-columns: 1fr; } }
</style>
