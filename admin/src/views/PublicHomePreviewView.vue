<script setup>
const navItems = ["首页", "策略中心", "资讯", "会员", "关于我们"];

const pulseCards = [
  { title: "今日热点追踪", value: "128", unit: "条", trend: "+12.6%", trendType: "up" },
  { title: "策略信号触发", value: "43", unit: "次", trend: "+8.4%", trendType: "up" },
  { title: "风险预警提醒", value: "9", unit: "条", trend: "-18.2%", trendType: "down" },
  { title: "会员策略覆盖", value: "92", unit: "%", trend: "+3.1%", trendType: "up" }
];

const strategyCards = [
  {
    type: "股票策略",
    title: "价值修复组合",
    summary: "基于估值分位和资金强度，筛选低回撤标的。",
    tags: ["低波动", "中周期", "VIP可见"]
  },
  {
    type: "期货策略",
    title: "趋势跟踪篮子",
    summary: "结合量价结构和持仓变化，动态调整仓位区间。",
    tags: ["日内风控", "自动提醒", "可回测"]
  },
  {
    type: "套利机会",
    title: "跨期价差监控",
    summary: "监测异常价差，给出入场、止盈和风控区间。",
    tags: ["实时监测", "高频信号", "多品类"]
  }
];

const quickNews = [
  { time: "09:20", title: "宏观流动性边际改善，成长板块活跃度提升", level: "重点" },
  { time: "10:05", title: "三大行业龙头公布业绩预告，超预期占比上升", level: "快讯" },
  { time: "13:30", title: "商品端波动率抬升，建议缩短短线持仓周期", level: "提醒" },
  { time: "15:10", title: "盘后资金流向更新，北向净流入连续三日增长", level: "重点" }
];

const plans = [
  { name: "体验版", price: "0", desc: "快速体验核心资讯与公开策略", perks: ["每日快讯", "公开策略", "基础消息提醒"] },
  {
    name: "专业版",
    price: "199",
    desc: "适合高频投资者的策略与风控组合",
    perks: ["全量策略", "高级筛选", "风险预警与回测报告"],
    highlight: true
  },
  { name: "机构版", price: "699", desc: "面向团队协作与多账户管理", perks: ["团队协作", "权限管理", "专属顾问与定制报表"] }
];

const mobileTabs = ["首页", "策略", "资讯", "会员", "我的"];
</script>

<template>
  <div class="home-preview">
    <div class="ambient ambient-a" />
    <div class="ambient ambient-b" />

    <header class="topbar">
      <div class="brand-wrap">
        <div class="brand-mark">S</div>
        <div class="brand-text">
          <strong>SercherAI</strong>
          <span>投资研究平台</span>
        </div>
      </div>
      <nav class="desktop-nav">
        <a v-for="item in navItems" :key="item" href="#">{{ item }}</a>
      </nav>
      <button class="btn btn-primary">立即体验</button>
    </header>

    <main class="main">
      <section class="hero reveal reveal-1">
        <div class="hero-copy">
          <p class="eyebrow">松柏绿 + 香槟米 视觉提案</p>
          <h1>一眼看清市场节奏，策略与资讯同步到达</h1>
          <p class="subtitle">
            这是前台主页样式稿（静态展示），用于确认 Web 与 H5 的统一视觉语言、信息结构和交互层级。
          </p>
          <div class="hero-actions">
            <button class="btn btn-primary">查看今日策略</button>
            <button class="btn btn-ghost">浏览资讯中心</button>
          </div>
        </div>
        <div class="hero-card">
          <div class="hero-card-head">
            <span>今日综合得分</span>
            <strong>84.6</strong>
          </div>
          <div class="hero-meter">
            <div class="hero-meter-fill" />
          </div>
          <div class="hero-stat-grid">
            <div>
              <em>股票信号</em>
              <strong>26</strong>
            </div>
            <div>
              <em>期货信号</em>
              <strong>11</strong>
            </div>
            <div>
              <em>套利跟踪</em>
              <strong>6</strong>
            </div>
            <div>
              <em>风险提醒</em>
              <strong>3</strong>
            </div>
          </div>
        </div>
      </section>

      <section class="section reveal reveal-2">
        <div class="section-head">
          <h2>市场脉搏</h2>
          <span>实时概览（静态示例）</span>
        </div>
        <div class="pulse-grid">
          <article v-for="card in pulseCards" :key="card.title" class="pulse-card">
            <p>{{ card.title }}</p>
            <div class="pulse-value">
              <strong>{{ card.value }}</strong>
              <span>{{ card.unit }}</span>
            </div>
            <small :class="card.trendType === 'up' ? 'trend-up' : 'trend-down'">{{ card.trend }}</small>
          </article>
        </div>
      </section>

      <section class="section mixed reveal reveal-3">
        <div class="strategy-panel">
          <div class="section-head">
            <h2>策略推荐</h2>
            <span>适配 Web 高密度 + H5 卡片流</span>
          </div>
          <div class="strategy-list">
            <article v-for="item in strategyCards" :key="item.title" class="strategy-card">
              <span class="strategy-type">{{ item.type }}</span>
              <h3>{{ item.title }}</h3>
              <p>{{ item.summary }}</p>
              <div class="strategy-tags">
                <span v-for="tag in item.tags" :key="tag">{{ tag }}</span>
              </div>
            </article>
          </div>
        </div>

        <aside class="news-panel">
          <div class="section-head">
            <h2>资讯快讯</h2>
            <span>新闻摘要 + 时间线</span>
          </div>
          <ul>
            <li v-for="item in quickNews" :key="`${item.time}-${item.title}`">
              <time>{{ item.time }}</time>
              <div>
                <p>{{ item.title }}</p>
                <small>{{ item.level }}</small>
              </div>
            </li>
          </ul>
        </aside>
      </section>

      <section class="section reveal reveal-4">
        <div class="section-head">
          <h2>会员方案</h2>
          <span>明确分层，便于转化</span>
        </div>
        <div class="plan-grid">
          <article v-for="plan in plans" :key="plan.name" class="plan-card" :class="{ highlight: plan.highlight }">
            <h3>{{ plan.name }}</h3>
            <div class="price-line">
              <strong>{{ plan.price }}</strong>
              <span>元 / 月</span>
            </div>
            <p class="plan-desc">{{ plan.desc }}</p>
            <ul>
              <li v-for="perk in plan.perks" :key="perk">{{ perk }}</li>
            </ul>
            <button class="btn" :class="plan.highlight ? 'btn-primary' : 'btn-ghost'">立即开通</button>
          </article>
        </div>
      </section>

      <section class="section mobile-preview reveal reveal-5">
        <div class="section-head">
          <h2>H5 端交互预览</h2>
          <span>底部导航 + 卡片化内容布局</span>
        </div>
        <div class="phone-frame">
          <div class="phone-head">
            <span>09:41</span>
            <span>5G</span>
          </div>
          <div class="phone-hero">
            <strong>今日策略摘要</strong>
            <p>涨跌节奏平稳，建议优先关注防御与高股息方向。</p>
          </div>
          <div class="phone-cards">
            <article>
              <span>热点资讯</span>
              <strong>12 条更新</strong>
            </article>
            <article>
              <span>信号触发</span>
              <strong>4 条待查看</strong>
            </article>
          </div>
          <div class="phone-tabs">
            <span v-for="(item, index) in mobileTabs" :key="item" :class="{ active: index === 0 }">{{ item }}</span>
          </div>
        </div>
      </section>
    </main>

    <footer class="footer">
      <p>Homepage UI Preview - Pine & Champagne Theme</p>
      <small>仅视觉与交互展示，不接入真实数据</small>
    </footer>
  </div>
</template>

<style scoped>
.home-preview {
  --pine: #2f6f62;
  --pine-dark: #275c52;
  --accent: #c08b3c;
  --bg: #f6f4ef;
  --surface: #ffffff;
  --text: #1f2933;
  --muted: #5f6b76;
  --border: #d8dfd8;
  --success: #2e7d32;
  --danger: #b23a2a;
  --hero-serif: "Source Han Serif SC", "Noto Serif SC", "STSong", serif;
  --body-sans: "Source Han Sans SC", "PingFang SC", "HarmonyOS Sans SC", "Microsoft YaHei", sans-serif;
  position: relative;
  min-height: 100vh;
  padding: 24px clamp(16px, 3vw, 44px) 120px;
  overflow: hidden;
  background:
    radial-gradient(circle at 8% 10%, #e7efe8 0%, transparent 28%),
    radial-gradient(circle at 92% 2%, #f2e8d5 0%, transparent 24%),
    linear-gradient(160deg, #f8f6f1 0%, #f3f0e9 45%, #efebe2 100%);
  color: var(--text);
  font-family: var(--body-sans);
}

.ambient {
  position: absolute;
  z-index: 0;
  border-radius: 999px;
  filter: blur(2px);
}

.ambient-a {
  width: 340px;
  height: 340px;
  right: -120px;
  top: -120px;
  background: rgba(47, 111, 98, 0.13);
}

.ambient-b {
  width: 280px;
  height: 280px;
  left: -110px;
  bottom: 120px;
  background: rgba(192, 139, 60, 0.13);
}

.topbar,
.main,
.footer {
  position: relative;
  z-index: 1;
}

.topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 20px;
  padding: 14px 18px;
  margin-bottom: 18px;
  border: 1px solid rgba(216, 223, 216, 0.75);
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.84);
  backdrop-filter: blur(8px);
}

.brand-wrap {
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-mark {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: grid;
  place-items: center;
  color: #fff;
  font-weight: 700;
  background: linear-gradient(145deg, var(--pine), var(--pine-dark));
  box-shadow: 0 8px 16px rgba(39, 92, 82, 0.32);
}

.brand-text {
  display: flex;
  flex-direction: column;
  line-height: 1.1;
}

.brand-text strong {
  font-size: 16px;
}

.brand-text span {
  font-size: 11px;
  color: var(--muted);
}

.desktop-nav {
  display: flex;
  gap: 20px;
}

.desktop-nav a {
  position: relative;
  font-size: 14px;
  color: #324148;
}

.desktop-nav a::after {
  content: "";
  position: absolute;
  left: 0;
  bottom: -7px;
  width: 0;
  height: 2px;
  border-radius: 999px;
  background: var(--accent);
  transition: width 0.25s ease;
}

.desktop-nav a:hover::after {
  width: 100%;
}

.main {
  display: grid;
  gap: 16px;
}

.section {
  border: 1px solid rgba(216, 223, 216, 0.78);
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.88);
  padding: clamp(16px, 2vw, 24px);
  box-shadow: 0 14px 28px rgba(47, 66, 60, 0.06);
}

.hero {
  display: grid;
  grid-template-columns: 1.2fr 1fr;
  gap: 16px;
  align-items: stretch;
}

.eyebrow {
  margin: 0 0 10px;
  color: var(--pine);
  font-weight: 600;
  letter-spacing: 0.2px;
}

.hero-copy h1 {
  margin: 0;
  font-family: var(--hero-serif);
  font-size: clamp(28px, 4vw, 46px);
  line-height: 1.18;
  letter-spacing: 0.3px;
}

.subtitle {
  margin: 14px 0 0;
  max-width: 620px;
  color: var(--muted);
  font-size: 15px;
}

.hero-actions {
  margin-top: 22px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.hero-card {
  padding: 16px;
  border-radius: 16px;
  border: 1px solid #cfd8ce;
  background: linear-gradient(180deg, #fefcf7 0%, #f7f3e8 100%);
}

.hero-card-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.hero-card-head span {
  color: var(--muted);
  font-size: 13px;
}

.hero-card-head strong {
  color: var(--pine-dark);
  font-size: 32px;
}

.hero-meter {
  margin: 14px 0;
  height: 9px;
  border-radius: 999px;
  background: rgba(47, 111, 98, 0.2);
  overflow: hidden;
}

.hero-meter-fill {
  width: 84.6%;
  height: 100%;
  border-radius: inherit;
  background: linear-gradient(90deg, var(--pine), #4c8f80);
}

.hero-stat-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.hero-stat-grid div {
  border: 1px solid rgba(216, 223, 216, 0.9);
  border-radius: 12px;
  padding: 10px;
  background: #fff;
}

.hero-stat-grid em {
  display: block;
  color: var(--muted);
  font-size: 12px;
  font-style: normal;
}

.hero-stat-grid strong {
  font-size: 22px;
  color: #2b4f49;
}

.section-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 14px;
}

.section-head h2 {
  margin: 0;
  font-size: clamp(18px, 2vw, 24px);
  font-family: var(--hero-serif);
}

.section-head span {
  color: var(--muted);
  font-size: 12px;
}

.pulse-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.pulse-card {
  border: 1px solid rgba(216, 223, 216, 0.88);
  border-radius: 14px;
  background: #fff;
  padding: 14px;
}

.pulse-card p {
  margin: 0;
  color: var(--muted);
  font-size: 13px;
}

.pulse-value {
  margin: 8px 0 6px;
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.pulse-value strong {
  font-size: 30px;
  line-height: 1;
  color: #21473f;
}

.pulse-value span {
  color: var(--muted);
  font-size: 13px;
}

.trend-up {
  color: var(--success);
}

.trend-down {
  color: var(--danger);
}

.mixed {
  display: grid;
  grid-template-columns: 1.1fr 0.9fr;
  gap: 14px;
}

.strategy-list {
  display: grid;
  gap: 10px;
}

.strategy-card {
  border: 1px solid rgba(216, 223, 216, 0.85);
  border-radius: 14px;
  background: linear-gradient(180deg, #ffffff 0%, #fcfbf8 100%);
  padding: 14px;
}

.strategy-type {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 999px;
  padding: 2px 10px;
  background: rgba(47, 111, 98, 0.12);
  color: var(--pine-dark);
  font-size: 12px;
  font-weight: 600;
}

.strategy-card h3 {
  margin: 8px 0 6px;
  font-size: 20px;
  color: #223338;
}

.strategy-card p {
  margin: 0;
  color: var(--muted);
  font-size: 13px;
}

.strategy-tags {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.strategy-tags span {
  border-radius: 999px;
  border: 1px solid #d9decc;
  padding: 2px 8px;
  font-size: 11px;
  color: #5d6652;
  background: #f9faef;
}

.news-panel ul {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 10px;
}

.news-panel li {
  display: grid;
  grid-template-columns: 54px 1fr;
  gap: 10px;
  border: 1px solid rgba(216, 223, 216, 0.86);
  border-radius: 12px;
  background: #fff;
  padding: 10px;
}

.news-panel time {
  color: var(--pine-dark);
  font-size: 13px;
  font-weight: 600;
}

.news-panel p {
  margin: 0;
  line-height: 1.4;
}

.news-panel small {
  color: var(--accent);
  font-size: 11px;
}

.plan-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.plan-card {
  border: 1px solid rgba(216, 223, 216, 0.9);
  border-radius: 16px;
  background: #fff;
  padding: 16px;
}

.plan-card.highlight {
  border-color: rgba(47, 111, 98, 0.5);
  box-shadow: 0 14px 26px rgba(39, 92, 82, 0.14);
  transform: translateY(-2px);
}

.plan-card h3 {
  margin: 0;
  font-size: 20px;
}

.price-line {
  margin-top: 10px;
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.price-line strong {
  font-size: 36px;
  color: #274740;
}

.price-line span {
  color: var(--muted);
}

.plan-desc {
  margin: 8px 0 10px;
  color: var(--muted);
  min-height: 38px;
}

.plan-card ul {
  margin: 0 0 14px;
  padding-left: 18px;
  color: #33454a;
  display: grid;
  gap: 4px;
}

.mobile-preview {
  display: grid;
  justify-items: center;
}

.phone-frame {
  width: min(360px, 100%);
  border: 1px solid #cad4ca;
  border-radius: 28px;
  background: #fffdf7;
  padding: 14px;
  box-shadow: 0 20px 30px rgba(31, 41, 51, 0.13);
}

.phone-head {
  display: flex;
  justify-content: space-between;
  color: #6b7280;
  font-size: 12px;
}

.phone-hero {
  margin-top: 10px;
  border: 1px solid rgba(47, 111, 98, 0.2);
  border-radius: 14px;
  background: linear-gradient(150deg, #edf5ef 0%, #f8f3e6 100%);
  padding: 12px;
}

.phone-hero strong {
  font-size: 17px;
}

.phone-hero p {
  margin: 6px 0 0;
  color: var(--muted);
  font-size: 13px;
}

.phone-cards {
  margin-top: 10px;
  display: grid;
  gap: 8px;
}

.phone-cards article {
  border: 1px solid #dbe2d8;
  border-radius: 12px;
  background: #fff;
  padding: 10px;
  display: flex;
  justify-content: space-between;
}

.phone-cards span {
  color: var(--muted);
  font-size: 13px;
}

.phone-cards strong {
  color: #2a4b43;
}

.phone-tabs {
  margin-top: 12px;
  border-top: 1px solid #dce3da;
  padding-top: 8px;
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  text-align: center;
}

.phone-tabs span {
  color: #7b8591;
  font-size: 12px;
}

.phone-tabs span.active {
  color: var(--pine-dark);
  font-weight: 700;
}

.footer {
  margin-top: 16px;
  text-align: center;
  color: var(--muted);
}

.footer p {
  margin: 0;
  font-size: 13px;
}

.footer small {
  display: block;
  margin-top: 4px;
  font-size: 12px;
}

.btn {
  border-radius: 11px;
  border: 1px solid transparent;
  padding: 8px 14px;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-primary {
  background: linear-gradient(135deg, var(--pine), var(--pine-dark));
  color: #fff;
  box-shadow: 0 10px 20px rgba(39, 92, 82, 0.25);
}

.btn-primary:hover {
  transform: translateY(-1px);
}

.btn-ghost {
  border-color: #cdd8cb;
  background: rgba(255, 255, 255, 0.72);
  color: #314046;
}

.btn-ghost:hover {
  border-color: var(--pine);
  color: var(--pine-dark);
}

.mobile-tabbar {
  display: none;
}

.reveal {
  opacity: 0;
  transform: translateY(10px);
  animation: reveal 0.65s ease forwards;
}

.reveal-1 {
  animation-delay: 0.06s;
}

.reveal-2 {
  animation-delay: 0.14s;
}

.reveal-3 {
  animation-delay: 0.22s;
}

.reveal-4 {
  animation-delay: 0.3s;
}

.reveal-5 {
  animation-delay: 0.38s;
}

@keyframes reveal {
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@media (max-width: 1160px) {
  .pulse-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .plan-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 860px) {
  .topbar {
    padding: 12px;
  }

  .desktop-nav {
    display: none;
  }

  .hero {
    grid-template-columns: 1fr;
  }

  .mixed {
    grid-template-columns: 1fr;
  }

  .plan-grid {
    grid-template-columns: 1fr;
  }

  .mobile-tabbar {
    position: fixed;
    left: 14px;
    right: 14px;
    bottom: 12px;
    z-index: 15;
    display: grid;
    grid-template-columns: repeat(5, minmax(0, 1fr));
    border: 1px solid rgba(216, 223, 216, 0.95);
    border-radius: 14px;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(8px);
    padding: 8px 6px;
    text-align: center;
    box-shadow: 0 14px 28px rgba(23, 33, 36, 0.15);
    color: #7a848b;
    font-size: 12px;
  }

  .mobile-tabbar span:first-child {
    color: var(--pine-dark);
    font-weight: 700;
  }
}
</style>
