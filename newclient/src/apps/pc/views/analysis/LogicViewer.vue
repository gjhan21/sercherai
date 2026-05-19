<template>
  <div class="logic-page">
    <section class="section fade-in-up">
      <div class="section-header">
        <h2 class="section-title">AI 推理逻辑</h2>
        <p class="section-subtitle">透明化展示 AI 的4维度评分、分析师投票与场景推演</p>
      </div>

      <div class="logic-intro glass">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20" style="color:var(--accent-gold);flex-shrink:0"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h0"/></svg>
        <p>下列数据来自 AI 模型的多因子评分、多分析师投票和场景模拟，展示了完整的推理链路。</p>
      </div>

      <!-- Step 1: Multi-factor scoring -->
      <div class="logic-step glass" :class="{ expanded: expandedStep === 0 }">
        <button class="logic-step-header" @click="toggleStep(0)">
          <span class="step-number">1</span>
          <div class="step-info">
            <span class="step-title">多因子评分</span>
            <span class="step-status" v-if="factors.length">已完成</span>
          </div>
          <span class="step-arrow" :class="{ open: expandedStep === 0 }">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M6 9l6 6 6-6"/></svg>
          </span>
        </button>
        <div class="logic-step-body" :class="{ collapsed: expandedStep !== 0 }">
          <div class="factor-dashboard">
            <div v-for="f in factors" :key="f.key" class="factor-item">
              <div class="factor-head">
                <span class="factor-label">{{ f.label }}</span>
                <span class="factor-score">{{ f.score.toFixed(1) }}</span>
                <span class="factor-weight">权重 {{ (f.weight * 100).toFixed(0) }}%</span>
                <span class="factor-contribution">贡献 {{ f.contribution.toFixed(1) }}</span>
              </div>
              <div class="factor-bar-bg">
                <div class="factor-bar" :style="{ width: f.score + '%', background: f.score >= 80 ? 'var(--positive)' : f.score >= 60 ? 'var(--accent-gold)' : 'var(--negative)' }"></div>
              </div>
              <p class="factor-desc" v-if="factorDesc(f.key)">{{ factorDesc(f.key) }}</p>
            </div>
          </div>
          <div class="logic-output" v-if="factors.length">
            <span>总分</span><strong>{{ totalScore }}</strong> |
            <span>加权分</span><strong>{{ weightedScore }}</strong> |
            <span>{{ scoreMethod }}</span>
          </div>
        </div>
      </div>

      <!-- Step 2: Agent opinions -->
      <div class="logic-step glass" :class="{ expanded: expandedStep === 1 }">
        <button class="logic-step-header" @click="toggleStep(1)">
          <span class="step-number">2</span>
          <div class="step-info">
            <span class="step-title">多分析师投票</span>
            <span class="step-status" v-if="agentOpinions.length">{{ agentOpinions.length }} 位分析师</span>
          </div>
          <span class="step-arrow" :class="{ open: expandedStep === 1 }">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M6 9l6 6 6-6"/></svg>
          </span>
        </button>
        <div class="logic-step-body" :class="{ collapsed: expandedStep !== 1 }">
          <p class="step-prompt">系统向每位分析师输入了相同的因子数据和市场信息，由各分析师独立给出判断：</p>
          <div class="agent-grid">
            <div v-for="a in agentOpinions" :key="a.role" class="agent-card">
              <div class="agent-head">
                <span class="agent-icon">{{ agentIcon(a.role) }}</span>
                <strong>{{ a.role }}分析师</strong>
                <span class="agent-stance" :class="a.stance?.toLowerCase()">{{ stanceLabel(a.stance) }}</span>
                <span class="agent-conf">置信: {{ (a.confidence * 100).toFixed(0) }}%</span>
              </div>
              <p>{{ a.summary }}</p>
              <span v-if="a.veto" class="agent-veto">⚠ 行使了否决权</span>
            </div>
          </div>
          <div class="logic-footer-small" v-if="agentOpinions.length">
            <span>共识决策: <strong style="color:var(--positive)">{{ consensusAction }}</strong></span>
            <span>是否否决: {{ hasVeto ? '是' : '否' }}</span>
          </div>
        </div>
      </div>

      <!-- Step 3: Scenario simulation -->
      <div class="logic-step glass" :class="{ expanded: expandedStep === 2 }">
        <button class="logic-step-header" @click="toggleStep(2)">
          <span class="step-number">3</span>
          <div class="step-info">
            <span class="step-title">场景模拟</span>
            <span class="step-status" v-if="scenarios.length">{{ scenarios.length }} 种场景</span>
          </div>
          <span class="step-arrow" :class="{ open: expandedStep === 2 }">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M6 9l6 6 6-6"/></svg>
          </span>
        </button>
        <div class="logic-step-body" :class="{ collapsed: expandedStep !== 2 }">
          <p class="step-prompt">AI 模拟了不同市场环境下该推荐标的的可能表现：</p>
          <div class="scenario-grid">
            <div v-for="s in scenarios" :key="s.scenario" class="scenario-card" :class="'sc-' + s.scenario">
              <div class="sc-head">
                <span class="sc-icon">{{ scenarioIcon(s.scenario) }}</span>
                <strong>{{ scenarioLabel(s.scenario) }}</strong>
                <span class="sc-conf">P = {{ (s.confidence * 100).toFixed(0) }}%</span>
              </div>
              <div class="sc-row"><span>论题:</span><span>{{ s.thesis }}</span></div>
              <div class="sc-row"><span>触发:</span><span>{{ s.trigger }}</span></div>
              <div class="sc-row"><span>建议:</span><span>{{ s.action_suggestion }}</span></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Step 4: Risk assessment -->
      <div class="logic-step glass" :class="{ expanded: expandedStep === 3 }">
        <button class="logic-step-header" @click="toggleStep(3)">
          <span class="step-number">4</span>
          <div class="step-info">
            <span class="step-title">风险与信心校准</span>
            <span class="step-status" v-if="riskFlags.length || calibration">已评估</span>
          </div>
          <span class="step-arrow" :class="{ open: expandedStep === 3 }">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M6 9l6 6 6-6"/></svg>
          </span>
        </button>
        <div class="logic-step-body" :class="{ collapsed: expandedStep !== 3 }">
          <div class="risk-section" v-if="riskFlags.length">
            <h4>风险标志</h4>
            <div v-for="f in riskFlags" :key="f" class="risk-flag">⚠ {{ f }}</div>
          </div>
          <div class="risk-section" v-if="invalidations.length">
            <h4>失效条件</h4>
            <div v-for="inv in invalidations" :key="inv" class="inv-item">• {{ inv }}</div>
          </div>
          <div class="confidence-section" v-if="calibration">
            <h4>信心校准</h4>
            <div class="cal-bars">
              <div class="cal-bar"><span>基础</span><div class="cal-bar-bg"><div class="cal-bar-fill" :style="{ width: (calibration.base_confidence * 100) + '%' }"></div></div><span>{{ (calibration.base_confidence * 100).toFixed(0) }}%</span></div>
              <div class="cal-bar"><span>调整</span><div class="cal-bar-bg"><div class="cal-bar-fill gold" :style="{ width: (calibration.adjusted_confidence * 100) + '%' }"></div></div><span style="color:var(--accent-gold)">{{ (calibration.adjusted_confidence * 100).toFixed(0) }}%</span></div>
            </div>
            <p class="cal-advisory" v-if="calibration.advisory_only">⚠ 此建议仅供参考，不构成投资意见</p>
          </div>
        </div>
      </div>

      <div class="logic-footer glass">
        <span class="tag tag-gold">AI 生成内容仅供参考</span>
        <p>以上推理链路展示了 AI 的数据分析全流程，不构成投资建议。</p>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { getStockRecommendationInsight, listStockRecommendations } from "@/api/market.js";
import { onMounted } from "vue";
import { buildLogicChain } from "@/mock/analysis.js";

const expandedStep = ref(0);
const insight = ref(null);

const factors = computed(() => insight.value?.ScoreFramework?.factors || []);
const totalScore = computed(() => insight.value?.ScoreFramework?.total_score?.toFixed(1) || '--');
const weightedScore = computed(() => insight.value?.ScoreFramework?.weighted_score?.toFixed(1) || '--');
const scoreMethod = computed(() => insight.value?.ScoreFramework?.method || '');
const agentOpinions = computed(() => insight.value?.Explanation?.agent_opinions || []);
const scenarios = computed(() => insight.value?.Explanation?.scenario_snapshots || []);
const calibration = computed(() => insight.value?.Explanation?.confidence_calibration || null);
const riskFlags = computed(() => insight.value?.Explanation?.risk_flags || []);
const invalidations = computed(() => insight.value?.Explanation?.invalidations || []);
const hasVeto = computed(() => agentOpinions.value.some(a => a.veto));
const consensusAction = computed(() => {
  const supports = agentOpinions.value.filter(a => a.stance === 'SUPPORT' || a.stance === 'BULLISH').length;
  if (hasVeto.value) return '降低风险暴露';
  if (supports >= agentOpinions.value.length / 2) return '按计划执行';
  return '观望等待';
});

function toggleStep(i) { expandedStep.value = expandedStep.value === i ? -1 : i; }

function factorDesc(key) {
  const descs = {
    tech: '计算5日/20日动量、均线偏离度、波动率和回撤',
    fund: '评估PE/PB估值水平、换手率合理性',
    sentiment: '分析新闻热度与正负面情绪比例',
    flow: '监测主力资金净流向与成交量变化'
  };
  return descs[key] || '';
}

function stanceLabel(s) {
  const m = { SUPPORT: '看多', BULLISH: '看多', WATCH: '观望', NEUTRAL: '观望', BEARISH: '看空' };
  return m[s] || s || '观望';
}

function scenarioLabel(s) {
  return s === 'bull' ? '乐观场景' : s === 'base' ? '基准场景' : s === 'bear' ? '悲观场景' : s;
}

function scenarioIcon(s) {
  return s === 'bull' ? '📈' : s === 'bear' ? '📉' : '📊';
}

function agentIcon(role) {
  const icons = { FLOW: '💰', THEME: '🎯', RISK: '🛡️', DIRECTION: '🧭', SUPPLY: '📦' };
  return icons[role] || '🧠';
}

async function loadInsight() {
  try {
    const list = await listStockRecommendations({ page: 1, page_size: 5 });
    const first = list?.items?.[0];
    if (first) {
      const result = await getStockRecommendationInsight(first.id);
      if (result) { insight.value = result; return; }
    }
  } catch { /* use mock */ }
  // mock fallback
  insight.value = {
    ScoreFramework: { method: 'growth-v1', total_score: 82, weighted_score: 78, score_gap: 4, factors: [{ key:'tech', label:'技术因子', weight:.3, score:84, contribution:25.2 }, { key:'fund', label:'基本面因子', weight:.3, score:78, contribution:23.4 }, { key:'sentiment', label:'情绪因子', weight:.2, score:88, contribution:17.6 }, { key:'flow', label:'资金流因子', weight:.2, score:76, contribution:15.2 }] },
    Explanation: {
      agent_opinions: [{ role:'FLOW', stance:'SUPPORT', confidence:.68, summary:'资金面中性偏积极', veto:false }, { role:'THEME', stance:'WATCH', confidence:.61, summary:'主题热度一般', veto:false }, { role:'RISK', stance:'WATCH', confidence:.73, summary:'风险可接受', veto:false }],
      scenario_snapshots: [{ scenario:'bull', thesis:'趋势延续', trigger:'量价共振', action_suggestion:'顺势跟踪', confidence:.72 }, { scenario:'base', thesis:'核心逻辑维持', trigger:'常规波动', action_suggestion:'按计划执行', confidence:.64 }, { scenario:'bear', thesis:'风险边界被触发', trigger:'消息扰动', action_suggestion:'收缩风险暴露', confidence:.38 }],
      risk_flags: ['市场整体波动风险', '行业政策变化风险'],
      invalidations: ['跌破关键支撑位'],
      confidence_calibration: { base_confidence:.52, adjusted_confidence:.65, drivers:['事件佐证','收益校准'], advisory_only:true }
    }
  };
}

onMounted(loadInsight);
</script>

<style scoped>
.logic-page { max-width: 1000px; display: grid; gap: 16px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.logic-intro { display: flex; gap: 10px; padding: 14px; border-radius: var(--radius-md); margin-bottom: 8px; font-size: 13px; color: var(--text-secondary); line-height: 1.6; }
.logic-step { padding: 0; border-radius: var(--radius-md); overflow: hidden; }
.logic-step-header { display: flex; align-items: center; gap: 12px; padding: 14px 16px; width: 100%; text-align: left; font-size: 14px; }
.logic-step-header:hover { background: rgba(255,255,255,.02); }
.step-number { width: 28px; height: 28px; border-radius: 50%; background: var(--accent-gold); color: #000; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; flex-shrink: 0; }
.step-info { flex: 1; }
.step-title { display: block; font-weight: 600; }
.step-status { font-size: 11px; color: var(--positive); }
.step-arrow { transition: transform .2s; color: var(--text-muted); }
.step-arrow.open { transform: rotate(180deg); }
.logic-step-body { padding: 0 16px 16px; }
.logic-step-body.collapsed { display: none; }
.step-prompt { font-size: 13px; color: var(--text-secondary); margin-bottom: 12px; line-height: 1.6; }

/* Factor bars */
.factor-dashboard { display: grid; gap: 12px; }
.factor-item { display: grid; gap: 4px; }
.factor-head { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.factor-label { font-weight: 600; min-width: 80px; }
.factor-score { font-size: 18px; font-weight: 700; }
.factor-weight { font-size: 11px; color: var(--text-muted); margin-left: auto; }
.factor-contribution { font-size: 11px; color: var(--accent-gold); }
.factor-bar-bg { height: 6px; background: rgba(255,255,255,.06); border-radius: 3px; overflow: hidden; }
.factor-bar { height: 100%; border-radius: 3px; transition: width .6s; }
.factor-desc { font-size: 12px; color: var(--text-muted); }
.logic-output { margin-top: 10px; padding: 8px; border-radius: var(--radius-sm); background: rgba(255,255,255,.02); font-size: 12px; color: var(--text-muted); display: flex; gap: 8px; flex-wrap: wrap; }

/* Agents */
.agent-grid { display: grid; gap: 8px; }
.agent-card { padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border); }
.agent-head { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; flex-wrap: wrap; }
.agent-icon { font-size: 16px; }
.agent-head strong { font-size: 13px; }
.agent-stance { font-size: 11px; padding: 1px 8px; border-radius: var(--radius-full); font-weight: 600; }
.agent-stance.support, .agent-stance.bullish { background: var(--positive-bg); color: var(--positive); }
.agent-stance.watch, .agent-stance.neutral { background: rgba(139,147,176,.1); color: var(--text-secondary); }
.agent-stance.bearish { background: var(--negative-bg); color: var(--negative); }
.agent-conf { margin-left: auto; font-size: 11px; color: var(--text-muted); }
.agent-card p { font-size: 12px; color: var(--text-secondary); line-height: 1.5; }
.agent-veto { font-size: 11px; color: var(--negative); display: block; margin-top: 4px; }
.logic-footer-small { margin-top: 10px; display: flex; gap: 16px; font-size: 12px; color: var(--text-secondary); }

/* Scenarios */
.scenario-grid { display: grid; gap: 8px; }
.scenario-card { padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border); }
.sc-head { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.sc-icon { font-size: 16px; }
.sc-head strong { font-size: 13px; }
.sc-conf { margin-left: auto; font-size: 12px; color: var(--accent-gold); font-weight: 600; }
.sc-row { display: flex; gap: 8px; font-size: 12px; padding: 3px 0; }
.sc-row span:first-child { color: var(--text-muted); min-width: 50px; }
.sc-row span:last-child { color: var(--text-secondary); }

/* Risk */
.risk-section { margin-bottom: 12px; }
.risk-section h4 { font-size: 14px; font-weight: 700; margin-bottom: 8px; }
.risk-flag { font-size: 13px; color: var(--text-secondary); padding: 4px 0; }
.inv-item { font-size: 13px; color: var(--negative); padding: 2px 0; }

/* Confidence */
.confidence-section h4 { font-size: 14px; font-weight: 700; margin-bottom: 8px; }
.cal-bars { display: grid; gap: 8px; }
.cal-bar { display: flex; align-items: center; gap: 8px; font-size: 13px; }
.cal-bar-bg { flex: 1; height: 6px; background: rgba(255,255,255,.06); border-radius: 3px; overflow: hidden; }
.cal-bar-fill { height: 100%; border-radius: 3px; background: var(--text-muted); }
.cal-bar-fill.gold { background: var(--accent-gold); }
.cal-advisory { margin-top: 8px; font-size: 12px; color: var(--accent-gold); }

.logic-footer { padding: 14px; border-radius: var(--radius-md); display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
.logic-footer p { font-size: 12px; color: var(--text-secondary); }
</style>
