<template>
  <div class="gateway-page">
    <DemoVariantSwitcher current="b" />

    <div class="page-grid" />

    <div class="page-shell">
      <header class="site-header">
        <div class="brand-block">
          <span class="brand-mark">S</span>
          <div>
            <p class="brand-title">SercherAI Institutional Gateway</p>
            <p class="brand-meta">方案 B · Institutional Research Gateway</p>
          </div>
        </div>

        <nav class="site-nav">
          <a href="#brief">今日焦点</a>
          <a href="#framework">研究框架</a>
          <a href="#solutions">解决方案</a>
          <a href="#evidence">证据记录</a>
        </nav>

        <div class="header-actions">
          <RouterLink class="ghost-btn" to="/search">统一检索</RouterLink>
          <RouterLink class="solid-btn" to="/membership">查看会员权益</RouterLink>
        </div>
      </header>

      <main>
        <section id="brief" class="hero-section">
          <article class="hero-copy">
            <p class="eyebrow">INSTITUTIONAL RESEARCH GATEWAY</p>
            <p class="hero-note">For research teams, advisors, and professional individual investors.</p>
            <h1>把机构级研究首页，做成一个可以真正承接后续动作的入口。</h1>
            <p class="hero-desc">
              这版比 A 更像真实产品首页，但仍然保持金融机构官网的气质。首页先给出今日旗舰观点，再明确研究路径、角色入口和下一层页面，不用营销话术，也不做产品面板堆叠。
            </p>

            <div class="hero-actions">
              <RouterLink class="solid-btn" to="/strategies">查看今日观点</RouterLink>
              <RouterLink class="ghost-btn" to="/news">阅读焦点研报</RouterLink>
            </div>

            <div class="trust-bullets">
              <article v-for="item in trustBullets" :key="item.title">
                <strong>{{ item.title }}</strong>
                <p>{{ item.desc }}</p>
              </article>
            </div>
          </article>

          <aside class="thesis-board">
            <div class="board-head">
              <div>
                <p class="board-kicker">Flagship View</p>
                <h2>{{ data.primaryStock.symbol }} {{ data.primaryStock.name }}</h2>
              </div>
              <span class="board-risk">{{ data.primaryStock.risk }}</span>
            </div>

            <p class="board-score">{{ data.primaryStock.score }}</p>
            <p class="board-range">{{ data.primaryStock.expected }} · {{ data.primaryStock.stopLoss }}</p>
            <p class="board-reason">{{ data.primaryStock.reason }}</p>

            <div class="board-chart">
              <div class="chart-head">
                <span>Research Summary</span>
                <strong>主推荐走势附图</strong>
              </div>
              <svg :viewBox="`0 0 ${sparkline.width} ${sparkline.height}`" role="img" aria-label="主推荐走势附图">
                <defs>
                  <linearGradient id="gatewaySparkFill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="rgba(116, 150, 211, 0.34)" />
                    <stop offset="100%" stop-color="rgba(116, 150, 211, 0)" />
                  </linearGradient>
                </defs>
                <path :d="sparkline.area" fill="url(#gatewaySparkFill)" />
                <path :d="sparkline.line" class="spark-line" />
              </svg>
            </div>

            <div class="board-metrics">
              <article v-for="item in boardMetrics" :key="item.title">
                <p>{{ item.title }}</p>
                <strong>{{ item.value }}</strong>
              </article>
            </div>
          </aside>
        </section>

        <section class="proof-strip">
          <article v-for="item in data.trustStats" :key="item.label">
            <p>{{ item.label }}</p>
            <strong>{{ item.value }}</strong>
            <span>{{ item.note }}</span>
          </article>
        </section>

        <section id="framework" class="framework-section">
          <div class="section-head">
            <div>
              <p class="section-kicker">Homepage Framework</p>
              <h2>首页先建立判断框架，再把用户送到正确的深层页面。</h2>
            </div>
            <p>这版不强调“内容多”，而强调首页应该如何把主推荐、解释链、证据记录和下一步动作组织得更像真正的机构门户。</p>
          </div>

          <div class="framework-grid">
            <article v-for="item in frameworkSteps" :key="item.title" class="framework-card">
              <p>{{ item.step }}</p>
              <h3>{{ item.title }}</h3>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </section>

        <section id="solutions" class="solutions-section">
          <div class="section-head">
            <div>
              <p class="section-kicker">Solutions By Role</p>
              <h2>同一个首页，为不同角色保留不同的后续路径，但保持统一的机构表达。</h2>
            </div>
            <RouterLink class="inline-link" to="/search">查看统一检索</RouterLink>
          </div>

          <div class="role-grid">
            <article v-for="item in data.roleCards" :key="item.title" class="role-card">
              <p>{{ item.kicker }}</p>
              <h3>{{ item.title }}</h3>
              <span>{{ item.desc }}</span>
              <RouterLink class="inline-link" :to="item.path">{{ item.cta }}</RouterLink>
            </article>
          </div>

          <div class="product-grid">
            <article class="panel panel-wide">
              <div class="panel-head">
                <div>
                  <p class="panel-kicker">Research Extension</p>
                  <h3>观察清单只保留最值得继续追踪的延伸对象。</h3>
                </div>
                <RouterLink class="inline-link" to="/strategies">进入策略页</RouterLink>
              </div>

              <div class="watch-grid">
                <article v-for="item in data.watchlist" :key="item.symbol" class="watch-row">
                  <div class="watch-main">
                    <strong>{{ item.symbol }} {{ item.name }}</strong>
                    <span>{{ item.risk }}</span>
                  </div>
                  <p>{{ item.expected }}</p>
                  <i>{{ item.note }}</i>
                </article>
              </div>
            </article>

            <div class="stack">
              <article class="panel">
                <div class="panel-head">
                  <div>
                    <p class="panel-kicker">Focused Research Note</p>
                    <h3>{{ data.featuredResearch.title }}</h3>
                  </div>
                  <RouterLink class="inline-link" to="/news">进入资讯中心</RouterLink>
                </div>

                <p class="panel-copy">{{ data.featuredResearch.summary }}</p>
                <div class="insight-list">
                  <article v-for="item in data.insightCards" :key="item.title">
                    <strong>{{ item.title }}</strong>
                    <p>{{ item.desc }}</p>
                  </article>
                </div>
              </article>

              <article class="panel dual-panel">
                <div class="dual-card">
                  <p class="panel-kicker">Cross-Market</p>
                  <h3>{{ data.futuresPlan.name }}</h3>
                  <p class="panel-copy">{{ data.futuresPlan.summary }}</p>
                </div>

                <div class="dual-card">
                  <p class="panel-kicker">Access Layer</p>
                  <h3>{{ data.vipOffer.name }}</h3>
                  <p class="price">{{ data.vipOffer.price }}</p>
                  <p class="panel-copy">{{ data.vipOffer.desc }}</p>
                </div>
              </article>
            </div>
          </div>
        </section>

        <section id="evidence" class="evidence-section">
          <div class="section-head">
            <div>
              <p class="section-kicker">Evidence & Record</p>
              <h2>让资讯热度、历史样本和后续动作组成一个完整的证据链。</h2>
            </div>
            <RouterLink class="inline-link" to="/archive">查看完整档案</RouterLink>
          </div>

          <div class="evidence-layout">
            <article class="panel pulse-panel">
              <div class="panel-head">
                <div>
                  <p class="panel-kicker">News Pulse</p>
                  <h3>资讯线索解释今天的判断为何成立。</h3>
                </div>
                <strong class="panel-api">GET /public/news/articles</strong>
              </div>

              <div class="pulse-list">
                <article v-for="item in data.newsPulse" :key="item.time">
                  <div class="pulse-meta">
                    <strong>{{ item.time }}</strong>
                    <span>{{ item.label }}</span>
                  </div>
                  <i :style="{ width: `${item.value}%` }" />
                </article>
              </div>
            </article>

            <article class="panel records-panel">
              <div class="panel-head">
                <div>
                  <p class="panel-kicker">Historic Records</p>
                  <h3>历史样本让首页具有长期记录，而不是只有当天观点。</h3>
                </div>
              </div>

              <div class="records-grid">
                <article v-for="item in data.historyRows" :key="`${item.code}-${item.date}`" class="record-card">
                  <p>{{ item.date }}</p>
                  <h4>{{ item.code }}</h4>
                  <span>{{ item.risk }}</span>
                  <strong>{{ item.alpha }}</strong>
                  <i :style="{ width: `${item.width}%` }" />
                </article>
              </div>
            </article>
          </div>
        </section>

        <section class="next-section">
          <div class="next-copy">
            <p class="section-kicker">Next Step</p>
            <h2>首页应该结束在正确的分流节点，而不是强行承担全部产品职责。</h2>
          </div>

          <div class="next-links">
            <RouterLink v-for="item in data.nextActions" :key="item.title" class="next-link" :to="item.path">
              <strong>{{ item.title }}</strong>
              <span>{{ item.desc }}</span>
            </RouterLink>
          </div>
        </section>
      </main>
    </div>
  </div>
</template>

<script setup>
import DemoVariantSwitcher from "./components/DemoVariantSwitcher.vue";
import { homeDemoData as data, sparkline } from "./homeDemoContent";

const trustBullets = [
  {
    title: "像官网，不像后台",
    desc: "更接近真实机构门户，而不是把首页做成面板集合。"
  },
  {
    title: "价值清楚，但不营销",
    desc: "仍然清楚表达产品价值，只是不使用 SaaS 化叙事。"
  },
  {
    title: "首页负责分流",
    desc: "策略、资讯、检索与会员都被组织成更合适的后续入口。"
  }
];

const frameworkSteps = [
  {
    step: "01",
    title: "今日观点先行",
    desc: "先交付最重要的研究判断，让首页具备机构级的主叙事。"
  },
  {
    step: "02",
    title: "解释链随后补齐",
    desc: "用焦点研报与资讯热度说明今天为什么优先看这些方向。"
  },
  {
    step: "03",
    title: "最后导向深层页面",
    desc: "首页完成判断与分流，不把全部产品职责都留在首屏。"
  }
];

const boardMetrics = [
  { title: "目标区间", value: data.primaryStock.expected },
  { title: "止盈建议", value: data.primaryStock.takeProfit },
  { title: "风控边界", value: data.primaryStock.stopLoss }
];
</script>

<style scoped>
.gateway-page {
  --bg: #08101b;
  --bg-soft: #0d1522;
  --panel: rgba(13, 21, 33, 0.84);
  --panel-strong: rgba(12, 19, 30, 0.94);
  --line: rgba(138, 157, 182, 0.16);
  --line-strong: rgba(110, 143, 194, 0.26);
  --text: #eef4ff;
  --muted: #9aa9bd;
  --muted-strong: #c6d4e8;
  --blue: #7ea7f0;
  --gold: #c8a06e;
  --shadow: 0 30px 58px rgba(0, 0, 0, 0.24);

  min-height: 100%;
  color: var(--text);
  background:
    radial-gradient(circle at 8% 0%, rgba(126, 167, 240, 0.12), transparent 22%),
    radial-gradient(circle at 90% 0%, rgba(200, 160, 110, 0.08), transparent 18%),
    linear-gradient(180deg, var(--bg) 0%, var(--bg-soft) 58%, #09111b 100%);
  font-family: "IBM Plex Sans", "Avenir Next", "Segoe UI", sans-serif;
}

.page-grid {
  position: fixed;
  inset: 0;
  pointer-events: none;
  opacity: 0.1;
  background-image:
    linear-gradient(90deg, rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(rgba(255, 255, 255, 0.02) 1px, transparent 1px);
  background-size: 92px 92px;
  mask-image: linear-gradient(180deg, black 0%, rgba(0, 0, 0, 0.45) 66%, transparent 100%);
}

.page-shell {
  position: relative;
  z-index: 1;
  width: min(1380px, calc(100% - 40px));
  margin: 0 auto;
  padding: 24px 0 44px;
}

.site-header {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 18px;
  padding-bottom: 18px;
  border-bottom: 1px solid var(--line);
}

.brand-block {
  display: inline-flex;
  align-items: center;
  gap: 12px;
}

.brand-mark {
  width: 40px;
  height: 40px;
  display: grid;
  place-items: center;
  border-radius: 10px;
  border: 1px solid rgba(200, 160, 110, 0.3);
  color: #0d1622;
  font-weight: 800;
  background: linear-gradient(135deg, #cfb289, #98b9f5);
}

.brand-title,
.brand-meta,
.eyebrow,
.hero-note,
.board-kicker,
.section-kicker,
.panel-kicker {
  margin: 0;
}

.brand-title {
  font-size: 14px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.brand-meta {
  margin-top: 4px;
  color: var(--muted);
  font-size: 12px;
}

.site-nav {
  display: flex;
  justify-content: center;
  gap: 28px;
  flex-wrap: wrap;
}

.site-nav a,
.inline-link {
  color: var(--muted-strong);
  text-decoration: none;
}

.site-nav a {
  font-size: 13px;
}

.inline-link {
  font-size: 13px;
}

.header-actions,
.hero-actions {
  display: inline-flex;
  gap: 12px;
  flex-wrap: wrap;
}

.solid-btn,
.ghost-btn {
  padding: 11px 18px;
  border-radius: 999px;
  text-decoration: none;
  font-size: 13px;
  transition: transform 180ms ease, border-color 180ms ease;
}

.solid-btn {
  color: #101925;
  background: linear-gradient(135deg, #b9d1fa, #d6b38a);
}

.ghost-btn {
  color: var(--text);
  border: 1px solid rgba(126, 167, 240, 0.26);
  background: rgba(255, 255, 255, 0.02);
}

.solid-btn:hover,
.ghost-btn:hover,
.next-link:hover {
  transform: translateY(-1px);
}

.hero-section {
  padding: 48px 0 44px;
  display: grid;
  grid-template-columns: minmax(0, 1.16fr) minmax(360px, 0.84fr);
  gap: 38px;
  align-items: stretch;
}

.eyebrow,
.board-kicker,
.section-kicker,
.panel-kicker {
  color: var(--blue);
  font-size: 12px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
}

.hero-note {
  margin-top: 10px;
  color: var(--muted);
  font-size: 12px;
  letter-spacing: 0.06em;
}

.hero-copy h1 {
  max-width: 12ch;
  margin: 20px 0 18px;
  font-size: clamp(50px, 5.7vw, 84px);
  line-height: 0.96;
  letter-spacing: -0.05em;
  font-weight: 600;
}

.hero-desc {
  max-width: 640px;
  color: var(--muted-strong);
  font-size: 17px;
  line-height: 1.82;
}

.hero-actions {
  margin-top: 26px;
}

.trust-bullets {
  margin-top: 40px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.trust-bullets article,
.thesis-board,
.proof-strip article,
.framework-card,
.role-card,
.panel,
.next-link {
  border: 1px solid var(--line);
  background: var(--panel);
  box-shadow: var(--shadow);
}

.trust-bullets article {
  padding: 18px;
}

.trust-bullets strong,
.insight-list strong {
  display: block;
  margin-bottom: 8px;
}

.trust-bullets p,
.panel-copy,
.watch-row p,
.watch-row i,
.insight-list p,
.role-card span,
.framework-card span,
.next-link span,
.proof-strip span {
  color: var(--muted);
  line-height: 1.75;
}

.thesis-board {
  padding: 26px;
  background: var(--panel-strong);
  border-top: 1px solid rgba(126, 167, 240, 0.24);
}

.board-head {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: start;
}

.board-head h2 {
  margin: 10px 0 0;
  font-size: 30px;
}

.board-risk,
.watch-main span,
.record-card span {
  display: inline-flex;
  align-items: center;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 12px;
  color: #dbe7ff;
  background: rgba(126, 167, 240, 0.12);
}

.board-score {
  margin: 18px 0 10px;
  font-size: 48px;
  line-height: 1;
  letter-spacing: -0.05em;
  font-weight: 600;
}

.board-range,
.board-reason {
  margin: 0 0 12px;
  color: var(--muted-strong);
  line-height: 1.8;
}

.board-chart {
  margin-top: 20px;
  padding-top: 18px;
  border-top: 1px solid var(--line);
}

.chart-head {
  margin-bottom: 12px;
  display: flex;
  justify-content: space-between;
  gap: 14px;
  color: var(--muted);
  font-size: 12px;
}

.spark-line {
  fill: none;
  stroke: #8bb1f7;
  stroke-width: 2.6;
  stroke-linecap: round;
}

.board-metrics {
  margin-top: 18px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1px;
  background: var(--line);
}

.board-metrics article {
  padding: 14px 12px;
  background: rgba(10, 17, 27, 0.92);
}

.board-metrics p {
  margin: 0 0 8px;
  color: var(--muted);
  font-size: 12px;
}

.proof-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 16px;
}

.proof-strip article {
  padding: 22px;
}

.proof-strip p {
  margin: 0;
  color: var(--muted);
  font-size: 13px;
}

.proof-strip strong {
  display: block;
  margin: 12px 0 8px;
  font-size: 36px;
  line-height: 1;
  letter-spacing: -0.04em;
}

.framework-section,
.solutions-section,
.evidence-section,
.next-section {
  padding-top: 80px;
}

.section-head {
  display: flex;
  justify-content: space-between;
  align-items: end;
  gap: 24px;
  margin-bottom: 28px;
}

.section-head h2 {
  max-width: 13ch;
  margin: 12px 0 0;
  font-size: clamp(32px, 4vw, 52px);
  line-height: 1.08;
  letter-spacing: -0.04em;
  font-weight: 600;
}

.section-head > p {
  max-width: 450px;
  color: var(--muted);
  line-height: 1.82;
}

.framework-grid,
.role-grid,
.next-links {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.framework-card,
.role-card,
.next-link {
  padding: 22px;
}

.framework-card p:first-child,
.role-card p:first-child {
  color: var(--gold);
  font-size: 12px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
}

.framework-card h3,
.role-card h3,
.panel h3 {
  margin: 12px 0 10px;
  font-size: 28px;
  line-height: 1.18;
}

.solutions-section .section-head {
  margin-bottom: 30px;
}

.product-grid {
  margin-top: 22px;
  display: grid;
  grid-template-columns: minmax(0, 1.02fr) minmax(360px, 0.98fr);
  gap: 18px;
}

.panel {
  padding: 24px;
}

.panel-head {
  display: flex;
  justify-content: space-between;
  gap: 18px;
  align-items: start;
}

.watch-grid,
.insight-list,
.pulse-list {
  margin-top: 18px;
  display: grid;
  gap: 14px;
}

.watch-row,
.insight-list article,
.pulse-list article {
  padding-top: 14px;
  border-top: 1px solid var(--line);
}

.watch-row:first-child {
  border-top: none;
  padding-top: 0;
}

.watch-main {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}

.watch-main strong {
  font-size: 18px;
}

.watch-row p,
.watch-row i {
  margin: 8px 0 0;
  display: block;
}

.stack {
  display: grid;
  gap: 18px;
}

.dual-panel {
  padding: 0;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  overflow: hidden;
}

.dual-card {
  padding: 24px;
}

.dual-card + .dual-card {
  border-left: 1px solid var(--line);
}

.price {
  margin: 14px 0 8px;
  font-size: 34px;
  color: #f0d2a6;
  line-height: 1;
}

.evidence-layout {
  display: grid;
  grid-template-columns: minmax(340px, 0.76fr) minmax(0, 1.24fr);
  gap: 18px;
}

.panel-api {
  color: var(--muted);
  font-size: 12px;
  font-weight: 500;
}

.pulse-meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.pulse-list i,
.record-card i {
  display: block;
  height: 7px;
  border-radius: 999px;
  background: linear-gradient(90deg, #7296d3, #d5b186);
}

.records-grid {
  margin-top: 18px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.record-card {
  padding: 18px;
  border: 1px solid var(--line);
  background: rgba(10, 17, 27, 0.72);
}

.record-card p,
.record-card h4 {
  margin: 0;
}

.record-card p {
  color: var(--muted);
  font-size: 12px;
}

.record-card h4 {
  margin-top: 10px;
  font-size: 24px;
  line-height: 1;
}

.record-card strong {
  display: block;
  margin: 16px 0 12px;
  font-size: 22px;
  color: #ecd0a7;
}

.next-section {
  padding-bottom: 10px;
}

.next-copy h2 {
  max-width: 15ch;
  margin: 12px 0 0;
  font-size: clamp(32px, 4vw, 50px);
  line-height: 1.08;
  letter-spacing: -0.04em;
  font-weight: 600;
}

.next-links {
  margin-top: 26px;
}

.next-link {
  text-decoration: none;
  transition: transform 180ms ease, border-color 180ms ease;
}

.next-link strong {
  display: block;
  margin-bottom: 10px;
  color: var(--text);
  font-size: 18px;
}

@media (max-width: 1240px) {
  .site-header,
  .hero-section,
  .proof-strip,
  .framework-grid,
  .role-grid,
  .product-grid,
  .evidence-layout,
  .records-grid,
  .next-links,
  .trust-bullets {
    grid-template-columns: 1fr;
  }

  .site-header,
  .section-head {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }

  .site-nav {
    justify-content: flex-start;
  }

  .dual-panel {
    grid-template-columns: 1fr;
  }

  .dual-card + .dual-card {
    border-left: none;
    border-top: 1px solid var(--line);
  }

  .board-metrics {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .page-shell {
    width: min(100% - 24px, 1380px);
    padding-top: 72px;
  }

  .hero-copy h1,
  .section-head h2,
  .next-copy h2 {
    max-width: none;
  }
}
</style>
