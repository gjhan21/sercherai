<template>
  <div class="analysis-page">
    <!-- Stock Header -->
    <section class="stock-header glass fade-in-up">
      <div class="stock-header-main">
        <div class="stock-primary">
          <div class="stock-ticker">
            <span class="stock-symbol">300750.SZ</span>
            <span class="tag tag-gold">AI 精选</span>
            <span class="tag tag-green">强烈看多</span>
          </div>
          <h1 class="stock-name">宁德时代</h1>
          <div class="stock-price-area">
            <span class="stock-price">198.62</span>
            <span class="stock-change up">+6.62 (+3.45%)</span>
            <span class="stock-market">深证 | 新能源</span>
          </div>
        </div>
        <div class="stock-actions">
          <button class="btn-primary"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M20 21v-2a4 4 0 00-4-4H8a4 4 0 00-4 4v2"/><circle cx="12" cy="7" r="4"/></svg> 添加关注</button>
          <button class="btn-ghost"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M4 12v8a2 2 0 002 2h12a2 2 0 002-2v-8"/><polyline points="16 6 12 2 8 6"/><line x1="12" y1="2" x2="12" y2="15"/></svg> 分享</button>
        </div>
      </div>
      <div class="stock-metrics">
        <div class="metric">
          <span class="metric-label">今开</span>
          <span class="metric-value">192.00</span>
        </div>
        <div class="metric">
          <span class="metric-label">最高</span>
          <span class="metric-value">199.50</span>
        </div>
        <div class="metric">
          <span class="metric-label">最低</span>
          <span class="metric-value">191.20</span>
        </div>
        <div class="metric">
          <span class="metric-label">成交量</span>
          <span class="metric-value">2,158.6万</span>
        </div>
        <div class="metric">
          <span class="metric-label">成交额</span>
          <span class="metric-value">42.37亿</span>
        </div>
        <div class="metric">
          <span class="metric-label">换手率</span>
          <span class="metric-value">1.82%</span>
        </div>
        <div class="metric">
          <span class="metric-label">PE(TTM)</span>
          <span class="metric-value">21.45</span>
        </div>
        <div class="metric">
          <span class="metric-label">市值</span>
          <span class="metric-value">8,742亿</span>
        </div>
      </div>
    </section>

    <!-- AI Score Dashboard -->
    <div class="analysis-grid-2col fade-in-up fade-in-up-delay-1">
      <section class="section ai-score-section">
        <div class="section-header">
          <h2 class="section-title">AI 综合评分</h2>
          <span class="section-subtitle">多维度深度学习模型评估</span>
        </div>
        <div class="score-dashboard">
          <div class="score-main">
            <div class="score-circle-big">
              <svg viewBox="0 0 120 120" class="score-ring-big">
                <circle cx="60" cy="60" r="52" fill="none" stroke="rgba(255,255,255,0.06)" stroke-width="6" />
                <circle cx="60" cy="60" r="52" fill="none" stroke="url(#scoreGrad)" stroke-width="6" stroke-dasharray="326.7" stroke-dashoffset="39.2" stroke-linecap="round" />
              </svg>
              <div class="score-center">
                <span class="score-big">92</span>
                <span class="score-label-big">/100</span>
              </div>
            </div>
            <div class="score-text">
              <strong>AI 强烈看多</strong>
              <p>基于多因子模型评分，该标的在<span class="text-positive">技术面</span>、<span class="text-gold">资金面</span>得分突出</p>
            </div>
          </div>
          <div class="score-details">
            <div v-for="dim in dimensions" :key="dim.name" class="score-dim">
              <div class="dim-header">
                <span class="dim-name">{{ dim.name }}</span>
                <span class="dim-score" :class="dim.score >= 70 ? 'up' : dim.score >= 40 ? '' : 'down'">{{ dim.score }}</span>
              </div>
              <div class="dim-bar-bg">
                <div class="dim-bar" :style="{ width: dim.score + '%', background: dim.score >= 70 ? 'var(--positive)' : dim.score >= 40 ? 'var(--accent-gold)' : 'var(--negative)' }"></div>
              </div>
            </div>
          </div>
        </div>
      </section>

      <section class="section ai-insight-section">
        <div class="section-header">
          <h2 class="section-title">AI 核心观点</h2>
          <span class="section-subtitle">基于最新数据的智能研判</span>
        </div>
        <div class="insight-cards">
          <div v-for="insight in insights" :key="insight.title" class="insight-card">
            <div class="insight-icon" :style="{ color: insight.color }">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
                <circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h0"/>
              </svg>
            </div>
            <div>
              <strong>{{ insight.title }}</strong>
              <p>{{ insight.desc }}</p>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- AI Analysis Report -->
    <section class="section fade-in-up fade-in-up-delay-2">
      <div class="section-header">
        <h2 class="section-title">AI 深度分析报告</h2>
        <p class="section-subtitle">自然语言生成的全面投资分析</p>
      </div>
      <div class="report-body">
        <div class="report-section">
          <h3>一、核心逻辑</h3>
          <p>宁德时代作为全球动力电池龙头企业，当前受益于新能源汽车渗透率持续提升、储能业务高速增长两大核心驱动力。AI 模型综合评估认为，公司在技术路线（麒麟电池、钠离子电池）、产能布局（全球十大基地）和成本控制方面保持显著领先优势。</p>
        </div>
        <div class="report-section">
          <h3>二、技术面分析</h3>
          <p>日线级别看，股价已突破前期整理平台（185-192 元区间），MACD 金叉向上，成交量温和放大，RSI 处于 62 中性偏强区域，尚未超买。周线级别呈现底部抬升态势，中期趋势向好。AI 技术模型给出 84 分（满分 100）。</p>
        </div>
        <div class="report-section">
          <h3>三、资金面分析</h3>
          <p>近 5 个交易日主力资金净流入 12.6 亿元，其中今日超大单净流入 3.8 亿元。北向资金持股比例由 8.2% 提升至 8.7%。融资余额持续增加，市场参与度显著提高。AI 资金模型给出 91 分。</p>
        </div>
        <div class="report-section">
          <h3>四、风险提示</h3>
          <p>1) 行业竞争加剧，二三线电池厂价格战可能压缩利润率；2) 原材料价格波动风险，碳酸锂价格反弹可能影响成本；3) 海外政策不确定性，欧美电动车关税调整可能影响出口预期。AI 风控模型建议仓位控制在总资金 15% 以内。</p>
        </div>
        <div class="report-footer">
          <span class="tag tag-gold">AI 生成内容仅供参考</span>
          <span class="report-date">报告生成时间：{{ currentTime }}</span>
        </div>
      </div>
    </section>

    <!-- AI Sentiment & Technical -->
    <div class="analysis-grid-2col fade-in-up fade-in-up-delay-3">
      <section class="section">
        <div class="section-header">
          <h2 class="section-title">市场情绪分析</h2>
        </div>
        <div class="sentiment-chart">
          <div v-for="s in sentiment" :key="s.label" class="sentiment-item">
            <span class="sentiment-label">{{ s.label }}</span>
            <div class="sentiment-bar-bg">
              <div class="sentiment-bar" :style="{ width: s.value + '%', background: s.color }"></div>
            </div>
            <span class="sentiment-value">{{ s.value }}%</span>
          </div>
        </div>
      </section>

      <section class="section">
        <div class="section-header">
          <h2 class="section-title">关键价位</h2>
        </div>
        <div class="levels-grid">
          <div v-for="level in levels" :key="level.label" class="level-item">
            <span class="level-label">{{ level.label }}</span>
            <span class="level-price" :style="{ color: level.color }">{{ level.price }}</span>
            <span class="level-note">{{ level.note }}</span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue";

const currentTime = computed(() => {
  const d = new Date();
  return `${d.getFullYear()}-${(d.getMonth()+1).toString().padStart(2,'0')}-${d.getDate().toString().padStart(2,'0')} ${d.getHours().toString().padStart(2,'0')}:${d.getMinutes().toString().padStart(2,'0')}`;
});

const dimensions = [
  { name: "技术面", score: 84 },
  { name: "资金面", score: 91 },
  { name: "基本面", score: 78 },
  { name: "情绪面", score: 88 },
  { name: "估值面", score: 65 }
];

const insights = [
  { title: "趋势判断", desc: "日线、周线同步多头排列，中期上涨趋势确立", color: "var(--positive)" },
  { title: "资金信号", desc: "主力连续三日净流入，北向资金加速配置", color: "var(--accent-gold)" },
  { title: "风险提示", desc: "短期涨幅较大，注意回调风险，建议分批建仓", color: "var(--accent-cyan)" },
  { title: "目标价位", desc: "AI 模型预测 1 个月内目标价 215-228 元区间", color: "var(--accent-purple)" }
];

const sentiment = [
  { label: "积极", value: 62, color: "var(--positive)" },
  { label: "中性", value: 24, color: "var(--text-muted)" },
  { label: "消极", value: 14, color: "var(--negative)" }
];

const levels = [
  { label: "强阻力位", price: "228.00", note: "前高 + 筹码密集区", color: "var(--negative)" },
  { label: "弱阻力位", price: "210.00", note: "整数关口", color: "var(--accent-gold)" },
  { label: "当前价位", price: "198.62", note: "AI 推荐建仓区间", color: "var(--accent-cyan)" },
  { label: "弱支撑位", price: "185.00", note: "平台下沿", color: "var(--accent-blue)" },
  { label: "强支撑位", price: "175.00", note: "60 日均线", color: "var(--positive)" }
];
</script>

<style scoped>
.analysis-page {
  display: grid;
  gap: 20px;
  max-width: 1400px;
}

.section {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 24px;
}

.section-header {
  display: flex;
  align-items: baseline;
  gap: 12px;
  margin-bottom: 16px;
}

/* Stock Header */
.stock-header {
  padding: 24px;
  border-radius: var(--radius-lg);
}

.stock-header-main {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}

.stock-ticker {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.stock-symbol {
  font-size: 13px;
  color: var(--text-secondary);
  font-family: var(--font-mono);
}

.stock-name {
  font-size: 28px;
  font-weight: 800;
  margin-bottom: 8px;
}

.stock-price-area {
  display: flex;
  align-items: baseline;
  gap: 12px;
}

.stock-price {
  font-size: 36px;
  font-weight: 800;
  letter-spacing: -1px;
}

.stock-change {
  font-size: 16px;
  font-weight: 600;
}

.stock-change.up { color: var(--positive); }
.stock-change.down { color: var(--negative); }

.stock-market {
  font-size: 12px;
  color: var(--text-secondary);
}

.stock-actions {
  display: flex;
  gap: 8px;
}

.stock-metrics {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  gap: 8px;
}

.metric {
  text-align: center;
  padding: 8px;
  border-radius: var(--radius-sm);
  background: rgba(255, 255, 255, 0.03);
}

.metric-label {
  display: block;
  font-size: 11px;
  color: var(--text-secondary);
  margin-bottom: 4px;
}

.metric-value {
  font-size: 13px;
  font-weight: 600;
}

/* Analysis Grid */
.analysis-grid-2col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

/* Score Section */
.score-dashboard {
  display: grid;
  gap: 16px;
}

.score-main {
  display: flex;
  align-items: center;
  gap: 20px;
}

.score-circle-big {
  position: relative;
  width: 120px;
  height: 120px;
  flex-shrink: 0;
}

.score-ring-big {
  width: 120px;
  height: 120px;
}

.score-center {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
}

.score-big {
  font-size: 32px;
  font-weight: 800;
  color: var(--accent-gold);
  line-height: 1;
}

.score-label-big {
  font-size: 11px;
  color: var(--text-secondary);
}

.score-text strong {
  display: block;
  font-size: 18px;
  margin-bottom: 4px;
  color: var(--positive);
}

.score-text p {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.score-details {
  display: grid;
  gap: 10px;
}

.score-dim {
  display: grid;
  gap: 4px;
}

.dim-header {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
}

.dim-name {
  color: var(--text-secondary);
}

.dim-score.up { color: var(--positive); }
.dim-score.down { color: var(--negative); }

.dim-bar-bg {
  height: 6px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 3px;
  overflow: hidden;
}

.dim-bar {
  height: 100%;
  border-radius: 3px;
  transition: width 0.6s;
}

/* Insights */
.insight-cards {
  display: grid;
  gap: 10px;
}

.insight-card {
  display: flex;
  gap: 12px;
  padding: 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
  background: rgba(255, 255, 255, 0.02);
}

.insight-icon {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.04);
  display: flex;
  align-items: center;
  justify-content: center;
}

.insight-card strong {
  display: block;
  font-size: 13px;
  margin-bottom: 4px;
}

.insight-card p {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
}

/* Report Body */
.report-body {
  display: grid;
  gap: 20px;
}

.report-section h3 {
  font-size: 15px;
  font-weight: 700;
  margin-bottom: 8px;
  color: var(--accent-gold);
}

.report-section p {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.8;
}

.report-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 16px;
  border-top: 1px solid var(--border);
}

.report-date {
  font-size: 11px;
  color: var(--text-muted);
}

/* Sentiment */
.sentiment-chart {
  display: grid;
  gap: 12px;
}

.sentiment-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.sentiment-label {
  font-size: 13px;
  width: 40px;
}

.sentiment-bar-bg {
  flex: 1;
  height: 8px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 4px;
  overflow: hidden;
}

.sentiment-bar {
  height: 100%;
  border-radius: 4px;
  transition: width 0.5s;
}

.sentiment-value {
  font-size: 12px;
  color: var(--text-secondary);
  width: 40px;
  text-align: right;
}

/* Levels */
.levels-grid {
  display: grid;
  gap: 8px;
}

.level-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: var(--radius-sm);
  border: 1px solid var(--border);
}

.level-label {
  font-size: 12px;
  color: var(--text-secondary);
  width: 80px;
}

.level-price {
  font-size: 18px;
  font-weight: 700;
  flex: 1;
}

.level-note {
  font-size: 12px;
  color: var(--text-muted);
}

.btn-primary {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 18px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, var(--accent-gold), var(--accent-gold-dim));
  color: #000;
  font-size: 13px;
  font-weight: 600;
  transition: opacity 0.2s;
}

.btn-primary:hover { opacity: 0.9; }

.btn-ghost {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: var(--radius-full);
  border: 1px solid var(--border);
  font-size: 12px;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.btn-ghost:hover {
  border-color: var(--border-light);
  color: var(--text-primary);
}

.text-positive { color: var(--positive); }

@media (max-width: 900px) {
  .analysis-grid-2col {
    grid-template-columns: 1fr;
  }
  .stock-metrics {
    grid-template-columns: repeat(4, 1fr);
  }
}

@media (max-width: 600px) {
  .stock-metrics {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
