<template>
  <div class="review-page">
    <DemoVariantSwitcher current="c" />

    <div class="page-shell">
      <header class="site-header">
        <div class="brand-block">
          <span class="brand-mark">S</span>
          <div>
            <p class="brand-title">SercherAI Capital Markets Review</p>
            <p class="brand-meta">方案 C · Capital Markets Review</p>
          </div>
        </div>

        <nav class="site-nav">
          <a href="#brief">市场摘要</a>
          <a href="#review">研究回顾</a>
          <a href="#record">记录样本</a>
          <a href="#close">收口动作</a>
        </nav>

        <div class="header-actions">
          <RouterLink class="ghost-btn" to="/archive">历史档案</RouterLink>
          <RouterLink class="solid-btn" to="/strategies">今日策略</RouterLink>
        </div>
      </header>

      <main>
        <section id="brief" class="hero-section">
          <article class="hero-copy">
            <p class="eyebrow">CAPITAL MARKETS REVIEW</p>
            <p class="hero-note">A homepage treatment closer to a flagship research publication.</p>
            <h1>让首页像一份机构市场评论，而不是一张产品菜单。</h1>
            <p class="hero-desc">
              这版仍然属于国际投行官网语气，但更强调“研究首页”的品牌感。首屏像一份市场摘要封面，下面再用回顾、样本和访问层级把今天的研究逻辑展开。
            </p>

            <div class="hero-actions">
              <RouterLink class="solid-btn" to="/news">阅读研究评论</RouterLink>
              <RouterLink class="ghost-btn" to="/community">进入讨论广场</RouterLink>
            </div>

            <div class="hero-ledger">
              <article v-for="item in ledgerStats" :key="item.title">
                <p>{{ item.title }}</p>
                <strong>{{ item.value }}</strong>
              </article>
            </div>
          </article>

          <aside class="thesis-board">
            <div class="board-head">
              <div>
                <p class="board-kicker">Lead Market View</p>
                <h2>{{ data.primaryStock.symbol }} {{ data.primaryStock.name }}</h2>
              </div>
              <span class="board-risk">{{ data.primaryStock.risk }}</span>
            </div>

            <p class="board-score">{{ data.primaryStock.score }}</p>
            <p class="board-range">{{ data.primaryStock.expected }} · {{ data.primaryStock.stopLoss }}</p>
            <p class="board-reason">{{ data.primaryStock.reason }}</p>

            <div class="board-chart">
              <svg :viewBox="`0 0 ${sparkline.width} ${sparkline.height}`" role="img" aria-label="首要观点附图">
                <defs>
                  <linearGradient id="reviewSparkFill" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stop-color="rgba(208, 176, 128, 0.34)" />
                    <stop offset="100%" stop-color="rgba(208, 176, 128, 0)" />
                  </linearGradient>
                </defs>
                <path :d="sparkline.area" fill="url(#reviewSparkFill)" />
                <path :d="sparkline.line" class="spark-line" />
              </svg>
            </div>

            <div class="board-summary">
              <article>
                <p>观点摘要</p>
                <strong>旗舰观点先行</strong>
              </article>
              <article>
                <p>解释方式</p>
                <strong>焦点研报补齐</strong>
              </article>
              <article>
                <p>下一步</p>
                <strong>进入深层页</strong>
              </article>
            </div>
          </aside>
        </section>

        <section class="trust-strip">
          <span v-for="item in data.trustMarks" :key="item">{{ item }}</span>
        </section>

        <section id="review" class="review-section">
          <div class="section-head">
            <div>
              <p class="section-kicker">Research Review</p>
              <h2>首页作为研究评论首页，更强调“回顾与解释”，而不是功能入口密度。</h2>
            </div>
            <p>这一版和 B 的区别在于：它更像研究主页、更像市场回顾栏目，但仍然保持投资机构官网的冷静与秩序。</p>
          </div>

          <div class="review-layout">
            <article class="panel article-panel">
              <div class="panel-head">
                <div>
                  <p class="panel-kicker">Lead Commentary</p>
                  <h3>{{ data.featuredResearch.title }}</h3>
                </div>
                <RouterLink class="inline-link" to="/news">阅读全文</RouterLink>
              </div>

              <p class="panel-copy">{{ data.featuredResearch.summary }}</p>

              <div class="article-points">
                <article v-for="item in data.insightCards" :key="item.title">
                  <strong>{{ item.title }}</strong>
                  <p>{{ item.desc }}</p>
                </article>
              </div>
            </article>

            <article class="panel list-panel">
              <div class="panel-head">
                <div>
                  <p class="panel-kicker">Tracked Names</p>
                  <h3>观察清单以“今日继续跟踪的标的”方式出现。</h3>
                </div>
                <RouterLink class="inline-link" to="/strategies">查看完整策略池</RouterLink>
              </div>

              <div class="watch-list">
                <article v-for="item in data.watchlist" :key="item.symbol">
                  <div class="watch-top">
                    <strong>{{ item.symbol }} {{ item.name }}</strong>
                    <span>{{ item.risk }}</span>
                  </div>
                  <p>{{ item.expected }}</p>
                  <i>{{ item.note }}</i>
                </article>
              </div>
            </article>
          </div>
        </section>

        <section id="record" class="record-section">
          <div class="section-head">
            <div>
              <p class="section-kicker">Recorded Evidence</p>
              <h2>历史记录、资讯热度与研究覆盖，在这里被组织成一组更品牌化的证据系统。</h2>
            </div>
            <RouterLink class="inline-link" to="/archive">进入历史档案</RouterLink>
          </div>

          <div class="record-layout">
            <article class="panel pulse-panel">
              <div class="panel-head">
                <div>
                  <p class="panel-kicker">News Pulse</p>
                  <h3>资讯线索的强弱变化，决定解释链是否足够完整。</h3>
                </div>
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

        <section class="coverage-section">
          <div class="section-head">
            <div>
              <p class="section-kicker">Coverage Depth</p>
              <h2>首页最后再说明平台能覆盖到哪里，而不是一开始就把产品功能全推到前面。</h2>
            </div>
          </div>

          <div class="coverage-grid">
            <article class="panel coverage-card">
              <p class="panel-kicker">Cross-Market Capability</p>
              <h3>{{ data.futuresPlan.name }}</h3>
              <p class="panel-copy">{{ data.futuresPlan.summary }}</p>
              <div class="steps-list">
                <article v-for="item in data.futuresSteps" :key="item.title">
                  <strong>{{ item.title }}</strong>
                  <p>{{ item.desc }}</p>
                </article>
              </div>
            </article>

            <article class="panel coverage-card">
              <p class="panel-kicker">Access Layer</p>
              <h3>{{ data.vipOffer.name }}</h3>
              <p class="price">{{ data.vipOffer.price }}</p>
              <p class="panel-copy">{{ data.vipOffer.desc }}</p>
              <div class="access-list">
                <article v-for="item in data.conversionStages" :key="item.title">
                  <strong>{{ item.title }}</strong>
                  <span>{{ item.badge }}</span>
                </article>
              </div>
            </article>

            <article class="panel coverage-card">
              <p class="panel-kicker">Role Entry</p>
              <h3>不同角色在同一个首页里看到不同的后续动作。</h3>
              <div class="role-list">
                <article v-for="item in data.roleCards" :key="item.title">
                  <strong>{{ item.title }}</strong>
                  <p>{{ item.desc }}</p>
                </article>
              </div>
            </article>
          </div>
        </section>

        <section id="close" class="closing-section">
          <div class="closing-copy">
            <p class="section-kicker">Closing Note</p>
            <h2>这版首页更像旗舰研究栏目首页，适合强调品牌化的研究表达。</h2>
          </div>

          <div class="closing-links">
            <RouterLink class="next-link" to="/strategies">
              <strong>进入策略页</strong>
              <span>从首页摘要继续进入完整策略与风险边界。</span>
            </RouterLink>
            <RouterLink class="next-link" to="/membership">
              <strong>查看会员权益</strong>
              <span>在理解公开价值之后，再承接更深层的研究访问权限。</span>
            </RouterLink>
            <RouterLink class="next-link" to="/search">
              <strong>进入统一检索</strong>
              <span>当今天的旗舰观点不足以覆盖你的问题时，由搜索继续接手。</span>
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

const ledgerStats = [
  { title: "主推荐评分", value: data.trustStats[0].value },
  { title: "观察清单", value: data.trustStats[1].value },
  { title: "资讯线索", value: data.trustStats[2].value }
];
</script>

<style scoped>
.review-page {
  --bg: #0b0d11;
  --bg-soft: #13161c;
  --panel: rgba(21, 23, 29, 0.86);
  --panel-strong: rgba(18, 20, 25, 0.94);
  --line: rgba(183, 167, 142, 0.16);
  --line-strong: rgba(208, 176, 128, 0.24);
  --text: #f3efe7;
  --muted: #a79d90;
  --muted-strong: #d3cbbf;
  --gold: #d0b080;
  --shadow: 0 30px 56px rgba(0, 0, 0, 0.26);

  min-height: 100%;
  color: var(--text);
  background:
    radial-gradient(circle at 16% 0%, rgba(208, 176, 128, 0.1), transparent 22%),
    linear-gradient(180deg, var(--bg) 0%, var(--bg-soft) 58%, #0e1116 100%);
  font-family: "IBM Plex Sans", "Avenir Next", "Segoe UI", sans-serif;
}

.page-shell {
  width: min(1340px, calc(100% - 40px));
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
  border: 1px solid var(--line-strong);
  color: #17120e;
  font-weight: 800;
  background: linear-gradient(135deg, #ecd5b0, #ba8e58);
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
  color: #17120e;
  background: linear-gradient(135deg, #f0d6af, #bb8d54);
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
  padding: 48px 0 44px;
  display: grid;
  grid-template-columns: minmax(0, 1.12fr) minmax(360px, 0.88fr);
  gap: 36px;
  align-items: stretch;
}

.eyebrow,
.board-kicker,
.section-kicker,
.panel-kicker {
  color: var(--gold);
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
  font-size: clamp(50px, 5.5vw, 82px);
  line-height: 0.97;
  letter-spacing: -0.05em;
  font-weight: 600;
}

.hero-desc {
  max-width: 630px;
  color: var(--muted-strong);
  font-size: 17px;
  line-height: 1.84;
}

.hero-actions {
  margin-top: 26px;
}

.hero-ledger {
  margin-top: 38px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1px;
  background: var(--line);
}

.hero-ledger article,
.thesis-board,
.trust-strip,
.panel,
.next-link {
  border: 1px solid var(--line);
  background: var(--panel);
  box-shadow: var(--shadow);
}

.hero-ledger article {
  padding: 18px;
  background: rgba(17, 20, 26, 0.94);
}

.hero-ledger p,
.panel-copy,
.article-points p,
.watch-list p,
.watch-list i,
.role-list p,
.steps-list p,
.next-link span {
  color: var(--muted);
  line-height: 1.76;
}

.hero-ledger strong {
  display: block;
  margin-top: 10px;
  font-size: 30px;
  line-height: 1;
  letter-spacing: -0.04em;
}

.thesis-board {
  padding: 26px;
  background: var(--panel-strong);
  border-top: 1px solid var(--line-strong);
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
.watch-top span,
.record-card span,
.access-list span {
  display: inline-flex;
  align-items: center;
  padding: 6px 10px;
  border-radius: 999px;
  font-size: 12px;
  color: #f0dec4;
  background: rgba(208, 176, 128, 0.12);
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

.spark-line {
  fill: none;
  stroke: #d8bb92;
  stroke-width: 2.6;
  stroke-linecap: round;
}

.board-summary {
  margin-top: 18px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 1px;
  background: var(--line);
}

.board-summary article {
  padding: 14px 12px;
  background: rgba(16, 18, 23, 0.95);
}

.board-summary p {
  margin: 0 0 8px;
  color: var(--muted);
  font-size: 12px;
}

.trust-strip {
  margin-top: 4px;
  padding: 16px 20px;
  display: flex;
  gap: 18px;
  flex-wrap: wrap;
  justify-content: center;
  color: var(--muted-strong);
  font-size: 12px;
  letter-spacing: 0.04em;
}

.review-section,
.record-section,
.coverage-section,
.closing-section {
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
  font-size: clamp(32px, 4vw, 50px);
  line-height: 1.08;
  letter-spacing: -0.04em;
  font-weight: 600;
}

.section-head > p {
  max-width: 430px;
  color: var(--muted);
  line-height: 1.8;
}

.review-layout,
.record-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.06fr) minmax(340px, 0.94fr);
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

.panel h3 {
  margin: 12px 0 10px;
  font-size: 28px;
  line-height: 1.18;
}

.article-points,
.watch-list,
.pulse-list,
.steps-list,
.access-list,
.role-list {
  margin-top: 18px;
  display: grid;
  gap: 14px;
}

.article-points article,
.watch-list article,
.pulse-list article,
.steps-list article,
.access-list article,
.role-list article {
  padding-top: 14px;
  border-top: 1px solid var(--line);
}

.watch-top,
.pulse-meta,
.access-list article {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
}

.watch-top strong,
.role-list strong,
.steps-list strong {
  display: block;
}

.pulse-meta {
  margin-bottom: 10px;
}

.pulse-list i,
.record-card i {
  display: block;
  height: 7px;
  border-radius: 999px;
  background: linear-gradient(90deg, #b18453, #e5c59b);
}

.records-grid,
.coverage-grid,
.closing-links {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 18px;
}

.record-card {
  padding: 18px;
  border: 1px solid var(--line);
  background: rgba(16, 18, 23, 0.76);
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
  color: #f0d5ae;
}

.price {
  margin: 14px 0 8px;
  font-size: 34px;
  color: #f0d4ad;
  line-height: 1;
}

.closing-copy h2 {
  max-width: 15ch;
  margin: 12px 0 0;
  font-size: clamp(32px, 4vw, 50px);
  line-height: 1.08;
  letter-spacing: -0.04em;
  font-weight: 600;
}

.closing-links {
  margin-top: 26px;
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

@media (max-width: 1240px) {
  .site-header,
  .hero-section,
  .hero-ledger,
  .review-layout,
  .record-layout,
  .records-grid,
  .coverage-grid,
  .closing-links {
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

  .board-summary {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .page-shell {
    width: min(100% - 24px, 1340px);
    padding-top: 72px;
  }

  .hero-copy h1,
  .section-head h2,
  .closing-copy h2 {
    max-width: none;
  }
}
</style>
