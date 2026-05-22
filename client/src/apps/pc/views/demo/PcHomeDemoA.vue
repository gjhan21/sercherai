<template>
  <div class="brief-page">
    <DemoVariantSwitcher current="a" />

    <div class="page-guides" />

    <div class="page-shell">
      <header class="site-header">
        <div class="brand-block">
          <span class="brand-mark">S</span>
          <div>
            <p class="brand-title">SercherAI Global Research</p>
            <p class="brand-meta">方案 A · Global Markets Brief</p>
          </div>
        </div>

        <nav class="site-nav">
          <a href="#brief">今日摘要</a>
          <a href="#coverage">研究覆盖</a>
          <a href="#records">历史记录</a>
          <a href="#next">下一步</a>
        </nav>

        <div class="header-actions">
          <RouterLink class="ghost-btn" to="/news">阅读焦点研报</RouterLink>
          <RouterLink class="solid-btn" to="/strategies">进入策略页</RouterLink>
        </div>
      </header>

      <main>
        <section id="brief" class="hero-section">
          <article class="hero-copy">
            <p class="eyebrow">GLOBAL MARKETS BRIEF</p>
            <p class="dateline">Institutional Homepage Briefing · 2026.04.04</p>
            <h1>把今天最重要的研究判断，交给一个真正像机构首页的入口。</h1>
            <p class="hero-desc">
              这版首页不再模仿产品面板，而是回到国际投行官网的表达方式：先交付旗舰观点，再给出解释链与长期记录，最后把用户送往真正承担动作的深层页面。
            </p>

            <div class="hero-actions">
              <RouterLink class="solid-btn" to="/strategies">查看今日旗舰观点</RouterLink>
              <RouterLink class="ghost-btn" to="/archive">查看历史档案</RouterLink>
            </div>

            <div class="hero-principles">
              <article v-for="item in principles" :key="item.title">
                <strong>{{ item.title }}</strong>
                <p>{{ item.desc }}</p>
              </article>
            </div>
          </article>

          <aside class="thesis-board">
            <div class="board-head">
              <div>
                <p class="board-kicker">Today&apos;s Flagship Thesis</p>
                <h2>{{ data.primaryStock.symbol }} {{ data.primaryStock.name }}</h2>
              </div>
              <span class="board-risk">{{ data.primaryStock.risk }}</span>
            </div>

            <p class="board-score">{{ data.primaryStock.score }}</p>
            <p class="board-range">{{ data.primaryStock.expected }} · {{ data.primaryStock.stopLoss }}</p>
            <p class="board-reason">{{ data.primaryStock.reason }}</p>

            <div class="board-chart">
              <div class="chart-head">
                <span>研究附图</span>
                <strong>主推荐走势摘要</strong>
              </div>
              <svg :viewBox="`0 0 ${sparkline.width} ${sparkline.height}`" role="img" aria-label="研究走势摘要">
                <defs>
                  <linearGradient id="briefSparkFill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="rgba(198, 167, 121, 0.28)" />
                    <stop offset="100%" stop-color="rgba(198, 167, 121, 0)" />
                  </linearGradient>
                </defs>
                <path :d="sparkline.area" fill="url(#briefSparkFill)" />
                <path :d="sparkline.line" class="spark-line" />
              </svg>
            </div>

            <div class="board-notes">
              <article>
                <p>目标区间</p>
                <strong>{{ data.primaryStock.expected }}</strong>
              </article>
              <article>
                <p>止盈建议</p>
                <strong>{{ data.primaryStock.takeProfit }}</strong>
              </article>
              <article>
                <p>风控边界</p>
                <strong>{{ data.primaryStock.stopLoss }}</strong>
              </article>
            </div>
          </aside>
        </section>

        <section class="proof-strip">
          <article v-for="item in data.trustStats" :key="item.label" class="proof-item">
            <p>{{ item.label }}</p>
            <strong>{{ item.value }}</strong>
            <span>{{ item.note }}</span>
          </article>
          <div class="proof-foot">
            <span v-for="item in data.trustMarks" :key="item">{{ item }}</span>
          </div>
        </section>

        <section id="coverage" class="coverage-section">
          <div class="section-head">
            <div>
              <p class="section-kicker">Research Coverage</p>
              <h2>首页只保留最关键的研究覆盖，不把深层功能搬到首屏上堆叠。</h2>
            </div>
            <p>主推荐负责给出今天的方向，焦点研报负责解释原因，多市场与会员承接则说明平台的覆盖深度与访问层级。</p>
          </div>

          <div class="coverage-layout">
            <article class="panel panel-wide">
              <div class="panel-top">
                <div>
                  <p class="panel-kicker">Extension Watchlist</p>
                  <h3>当旗舰观点成立，首页只继续保留最值得跟踪的延伸对象。</h3>
                </div>
                <RouterLink class="inline-link" to="/strategies">查看完整策略池</RouterLink>
              </div>

              <div class="watch-grid">
                <article v-for="item in data.watchlist" :key="item.symbol" class="watch-row">
                  <div class="watch-main">
                    <p>No.{{ item.rank }}</p>
                    <strong>{{ item.symbol }} {{ item.name }}</strong>
                  </div>
                  <p class="watch-range">{{ item.expected }}</p>
                  <span class="watch-risk">{{ item.risk }}</span>
                  <p class="watch-note">{{ item.note }}</p>
                </article>
              </div>
            </article>

            <div class="side-stack">
              <article class="panel">
                <div class="panel-top">
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

              <article class="panel split-panel">
                <div class="split-card">
                  <p class="panel-kicker">Cross-Market Coverage</p>
                  <h3>{{ data.futuresPlan.name }}</h3>
                  <p class="panel-copy">{{ data.futuresPlan.summary }}</p>

                  <div class="steps-list">
                    <article v-for="item in data.futuresSteps" :key="item.title">
                      <strong>{{ item.title }}</strong>
                      <p>{{ item.desc }}</p>
                    </article>
                  </div>
                </div>

                <div class="split-card split-access">
                  <p class="panel-kicker">Depth Of Access</p>
                  <h3>{{ data.vipOffer.name }}</h3>
                  <p class="access-price">{{ data.vipOffer.price }}</p>
                  <p class="panel-copy">{{ data.vipOffer.desc }}</p>

                  <div class="access-list">
                    <article v-for="item in data.conversionStages" :key="item.title">
                      <strong>{{ item.title }}</strong>
                      <span>{{ item.badge }}</span>
                    </article>
                  </div>
                </div>
              </article>
            </div>
          </div>
        </section>

        <section id="records" class="records-section">
          <div class="section-head records-head">
            <div>
              <p class="section-kicker">Track Record</p>
              <h2>把过往样本整理成规则化记录，让首页像一家有长期方法论的研究机构。</h2>
            </div>
            <RouterLink class="inline-link" to="/archive">查看完整历史档案</RouterLink>
          </div>

          <div class="records-layout">
            <article class="records-intro">
              <p class="records-lead">History is not used here as decoration. It is used to show discipline, repeatability, and a visible record of prior judgment.</p>
              <div class="news-pulse">
                <div class="pulse-head">
                  <span>资讯热度摘要</span>
                  <strong>今日解释链</strong>
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
              </div>
            </article>

            <div class="records-grid">
              <article v-for="item in data.historyRows" :key="`${item.code}-${item.date}`" class="record-card">
                <p>{{ item.date }}</p>
                <h3>{{ item.code }}</h3>
                <span>{{ item.risk }}</span>
                <strong>{{ item.alpha }}</strong>
                <i :style="{ width: `${item.width}%` }" />
              </article>
            </div>
          </div>
        </section>

        <section id="next" class="next-section">
          <div class="next-copy">
            <p class="section-kicker">Next Step</p>
            <h2>首页的职责到这里结束。理解今天的判断之后，请进入真正承接动作的下一层页面。</h2>
            <div class="next-actions">
              <RouterLink class="solid-btn" to="/strategies">进入策略页</RouterLink>
              <RouterLink class="ghost-btn" to="/news">进入资讯页</RouterLink>
            </div>
          </div>

          <div class="next-links">
            <RouterLink v-for="item in secondaryLinks" :key="item.title" class="next-link" :to="item.path">
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

const principles = [
  {
    title: "旗舰观点先行",
    desc: "首屏只交付最重要的研究结论，不先让用户陷入信息分散。"
  },
  {
    title: "解释链随后补齐",
    desc: "焦点研报与资讯热度负责解释今天为什么优先看这个方向。"
  },
  {
    title: "首页只负责分流",
    desc: "真正的操作、阅读与权益承接继续交给下一层页面完成。"
  }
];

const secondaryLinks = data.nextActions.filter((item) => item.path !== "/news");
</script>

<style scoped>
.brief-page {
  --bg: #09111b;
  --bg-soft: #0d1623;
  --panel: #0f1826;
  --panel-soft: #121d2d;
  --line: rgba(148, 163, 184, 0.14);
  --line-strong: rgba(198, 167, 121, 0.26);
  --text: #edf3f9;
  --muted: #94a2b5;
  --muted-strong: #b9c4d2;
  --accent: #c6a779;
  --accent-soft: rgba(198, 167, 121, 0.12);
  --shadow: 0 28px 54px rgba(0, 0, 0, 0.22);

  min-height: 100%;
  color: var(--text);
  background:
    radial-gradient(circle at 12% 0%, rgba(198, 167, 121, 0.08), transparent 24%),
    radial-gradient(circle at 88% 0%, rgba(104, 128, 164, 0.08), transparent 22%),
    linear-gradient(180deg, var(--bg) 0%, var(--bg-soft) 56%, #0a121c 100%);
  font-family: "IBM Plex Sans", "Avenir Next", "Segoe UI", sans-serif;
}

.page-guides {
  position: fixed;
  inset: 0;
  pointer-events: none;
  opacity: 0.12;
  background-image:
    linear-gradient(90deg, rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(rgba(255, 255, 255, 0.02) 1px, transparent 1px);
  background-size: 96px 96px;
  mask-image: linear-gradient(180deg, black 0%, rgba(0, 0, 0, 0.4) 62%, transparent 100%);
}

.page-shell {
  position: relative;
  z-index: 1;
  width: min(1360px, calc(100% - 40px));
  margin: 0 auto;
  padding: 26px 0 44px;
}

.site-header {
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 18px;
  padding: 0 0 18px;
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
  border: 1px solid var(--line-strong);
  border-radius: 10px;
  color: #101821;
  font-weight: 800;
  background: linear-gradient(135deg, #e2c79f, #c09b65);
}

.brand-title,
.brand-meta,
.eyebrow,
.dateline,
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
.hero-actions,
.next-actions {
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
  transition: transform 180ms ease, border-color 180ms ease, background-color 180ms ease;
}

.solid-btn {
  color: #111922;
  background: linear-gradient(135deg, #e6cb9f, #bd935a);
}

.ghost-btn {
  color: var(--text);
  border: 1px solid var(--line-strong);
  background: rgba(255, 255, 255, 0.02);
}

.solid-btn:hover,
.ghost-btn:hover,
.next-link:hover {
  transform: translateY(-1px);
}

.hero-section {
  padding: 52px 0 46px;
  display: grid;
  grid-template-columns: minmax(0, 1.25fr) minmax(360px, 0.78fr);
  gap: 48px;
  align-items: stretch;
}

.eyebrow,
.board-kicker,
.section-kicker,
.panel-kicker {
  color: var(--accent);
  font-size: 12px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
}

.dateline {
  margin-top: 10px;
  color: var(--muted);
  font-size: 12px;
  letter-spacing: 0.06em;
}

.hero-copy h1 {
  max-width: 11ch;
  margin: 22px 0 18px;
  font-size: clamp(54px, 6vw, 90px);
  line-height: 0.94;
  letter-spacing: -0.055em;
  font-weight: 600;
}

.hero-desc {
  max-width: 620px;
  color: var(--muted-strong);
  font-size: 17px;
  line-height: 1.84;
}

.hero-actions {
  margin-top: 28px;
}

.hero-principles {
  margin-top: 42px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.hero-principles article,
.thesis-board,
.proof-strip,
.panel,
.records-intro,
.record-card,
.next-link {
  border: 1px solid var(--line);
  background: rgba(12, 20, 31, 0.82);
  box-shadow: var(--shadow);
}

.hero-principles article {
  padding: 20px 18px;
}

.hero-principles strong,
.insight-list strong,
.steps-list strong,
.access-list strong {
  display: block;
  margin-bottom: 8px;
}

.hero-principles p,
.panel-copy,
.insight-list p,
.steps-list p,
.watch-note,
.records-lead,
.pulse-list span,
.next-link span,
.proof-item span {
  color: var(--muted);
  line-height: 1.76;
}

.thesis-board {
  padding: 26px;
  border-top: 1px solid var(--line-strong);
}

.board-head {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  align-items: flex-start;
}

.board-head h2 {
  margin: 10px 0 0;
  font-size: 30px;
  line-height: 1.15;
}

.board-risk,
.watch-risk,
.record-card span,
.access-list span {
  display: inline-flex;
  align-items: center;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 12px;
  color: #ead6b9;
  background: var(--accent-soft);
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
  font-size: 12px;
  color: var(--muted);
}

.spark-line {
  fill: none;
  stroke: #d7bb92;
  stroke-width: 2.6;
  stroke-linecap: round;
}

.board-notes {
  margin-top: 18px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1px;
  background: var(--line);
}

.board-notes article {
  padding: 14px 12px;
  background: rgba(10, 17, 27, 0.9);
}

.board-notes p {
  margin: 0 0 8px;
  color: var(--muted);
  font-size: 12px;
}

.board-notes strong {
  font-size: 14px;
}

.proof-strip {
  padding: 0;
  overflow: hidden;
}

.proof-strip {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 1px;
  background: var(--line);
}

.proof-item,
.proof-foot {
  background: rgba(10, 17, 27, 0.9);
}

.proof-item {
  padding: 22px 22px 20px;
}

.proof-item p {
  margin: 0;
  color: var(--muted);
  font-size: 13px;
}

.proof-item strong {
  display: block;
  margin: 10px 0 8px;
  font-size: 38px;
  line-height: 1;
  letter-spacing: -0.04em;
}

.proof-foot {
  grid-column: 1 / -1;
  padding: 14px 22px;
  display: flex;
  gap: 14px;
  flex-wrap: wrap;
  border-top: 1px solid var(--line);
  color: var(--muted-strong);
  font-size: 12px;
  letter-spacing: 0.03em;
}

.coverage-section,
.records-section,
.next-section {
  padding-top: 84px;
}

.section-head {
  display: flex;
  justify-content: space-between;
  align-items: end;
  gap: 24px;
  margin-bottom: 30px;
}

.section-head h2 {
  max-width: 12ch;
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

.coverage-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.08fr) minmax(360px, 0.92fr);
  gap: 20px;
}

.panel {
  padding: 24px;
}

.panel-wide {
  padding-bottom: 14px;
}

.panel-top {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: start;
}

.panel h3 {
  margin: 10px 0 0;
  font-size: 28px;
  line-height: 1.2;
}

.watch-grid {
  margin-top: 22px;
  display: grid;
}

.watch-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px 18px;
  padding: 18px 0;
  border-top: 1px solid var(--line);
}

.watch-row:first-child {
  border-top: none;
  padding-top: 0;
}

.watch-main p,
.watch-main strong,
.record-card p,
.record-card h3 {
  margin: 0;
}

.watch-main p {
  color: var(--muted);
  font-size: 12px;
  letter-spacing: 0.08em;
}

.watch-main strong {
  display: block;
  margin-top: 6px;
  font-size: 20px;
}

.watch-range {
  align-self: center;
  justify-self: start;
  margin: 0;
  color: var(--muted-strong);
}

.watch-note {
  grid-column: 1 / -1;
  margin: 0;
}

.side-stack {
  display: grid;
  gap: 20px;
}

.insight-list,
.steps-list,
.access-list {
  margin-top: 18px;
  display: grid;
  gap: 14px;
}

.insight-list article,
.steps-list article,
.access-list article {
  padding-top: 14px;
  border-top: 1px solid var(--line);
}

.split-panel {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0;
  padding: 0;
  overflow: hidden;
}

.split-card {
  padding: 24px;
}

.split-card + .split-card {
  border-left: 1px solid var(--line);
}

.access-price {
  margin: 14px 0 8px;
  font-size: 34px;
  color: #ebd1ad;
  line-height: 1;
}

.access-list article {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}

.records-layout {
  display: grid;
  grid-template-columns: minmax(300px, 0.78fr) minmax(0, 1.22fr);
  gap: 20px;
}

.records-intro {
  padding: 24px;
}

.records-lead {
  margin: 0;
  font-size: 16px;
}

.news-pulse {
  margin-top: 34px;
}

.pulse-head {
  display: flex;
  justify-content: space-between;
  gap: 14px;
  font-size: 12px;
  color: var(--muted);
}

.pulse-list {
  margin-top: 16px;
  display: grid;
  gap: 14px;
}

.pulse-list article {
  padding-top: 14px;
  border-top: 1px solid var(--line);
}

.pulse-meta {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
}

.pulse-meta strong {
  font-size: 13px;
}

.pulse-list i,
.record-card i {
  display: block;
  height: 7px;
  border-radius: 999px;
  background: linear-gradient(90deg, #af8755, #dec39d);
}

.records-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.record-card {
  padding: 20px 18px;
}

.record-card p {
  color: var(--muted);
  font-size: 12px;
  letter-spacing: 0.06em;
}

.record-card h3 {
  margin-top: 10px;
  font-size: 28px;
  line-height: 1;
}

.record-card strong {
  display: block;
  margin: 18px 0 14px;
  font-size: 24px;
  color: #f0d4ad;
  line-height: 1;
}

.next-section {
  padding-bottom: 8px;
  display: grid;
  grid-template-columns: minmax(0, 0.86fr) minmax(360px, 1.14fr);
  gap: 20px;
  align-items: start;
}

.next-copy h2 {
  max-width: 13ch;
  margin: 12px 0 0;
  font-size: clamp(32px, 4vw, 52px);
  line-height: 1.08;
  letter-spacing: -0.04em;
  font-weight: 600;
}

.next-actions {
  margin-top: 28px;
}

.next-links {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.next-link {
  padding: 22px;
  text-decoration: none;
  transition: transform 180ms ease, border-color 180ms ease;
}

.next-link strong {
  display: block;
  margin-bottom: 10px;
  color: var(--text);
  font-size: 18px;
}

.next-link span {
  display: block;
}

@media (max-width: 1240px) {
  .site-header,
  .hero-section,
  .coverage-layout,
  .records-layout,
  .next-section,
  .next-links,
  .records-grid,
  .hero-principles,
  .proof-strip {
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

  .split-panel {
    grid-template-columns: 1fr;
  }

  .split-card + .split-card {
    border-left: none;
    border-top: 1px solid var(--line);
  }

  .board-notes {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .page-shell {
    width: min(100% - 24px, 1360px);
    padding-top: 72px;
  }

  .hero-copy h1,
  .section-head h2,
  .next-copy h2 {
    max-width: none;
  }

  .watch-row {
    grid-template-columns: 1fr;
  }
}
</style>
