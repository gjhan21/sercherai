<template>
  <div class="corp-home-page">
    <div class="page-grid" />

    <header class="gateway-nav">
      <div class="brand-lockup">
        <span class="brand-mark">S</span>
        <div>
          <p class="brand-title">SercherAI Institutional Research</p>
          <p class="brand-meta">source-view: PcHomeView.vue</p>
        </div>
      </div>

      <nav class="gateway-links">
        <a href="#solutions">解决方案</a>
        <a href="#capabilities">核心能力</a>
        <a href="#evidence">数据证明</a>
        <a href="#next">后续动作</a>
      </nav>

      <div class="nav-actions">
        <RouterLink class="nav-btn nav-btn-ghost" to="/search">统一检索</RouterLink>
        <RouterLink class="nav-btn nav-btn-solid" to="/strategies">预约演示</RouterLink>
      </div>
    </header>

    <main class="site-shell">
      <section class="hero-section">
        <article class="hero-copy">
          <p class="eyebrow">TRUST & AUTHORITY</p>
          <h1>把今天最重要的研究结论，交付到一个可信的机构级首页。</h1>
          <p class="hero-desc">
            这版首页按高端金融官网重做，不沿用旧 demo 的终端或杂志布局。内容仍然映射
            <code>PcHomeView.vue</code>
            的真实首页职责：主推荐、观察清单、重点研报、历史样本、期货方案、社区承接和会员转化。
          </p>

          <div class="hero-actions">
            <RouterLink class="cta-btn cta-btn-solid" to="/strategies">查看今日主推荐</RouterLink>
            <RouterLink class="cta-btn cta-btn-ghost" to="/news">查看资讯中心</RouterLink>
            <RouterLink class="cta-btn cta-btn-ghost" to="/membership">查看会员权益</RouterLink>
          </div>

          <div class="mission-grid">
            <article v-for="item in missionPoints" :key="item.title" class="mission-card">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>

        <aside class="hero-board">
          <div class="board-head">
            <div>
              <p class="board-kicker">今日焦点</p>
              <h2>{{ primaryStock.symbol }} {{ primaryStock.name }}</h2>
            </div>
            <span class="risk-pill">{{ primaryStock.risk }}</span>
          </div>

          <p class="board-score">{{ primaryStock.score }}</p>
          <p class="board-range">{{ primaryStock.expected }}</p>
          <p class="board-copy">{{ primaryStock.reason }}</p>

          <div class="chart-shell">
            <div class="chart-head">
              <span>主推荐表现曲线</span>
              <strong>GET /stocks/recommendations/{id}/performance</strong>
            </div>
            <svg class="spark-chart" :viewBox="`0 0 ${sparkline.width} ${sparkline.height}`" role="img" aria-label="主推荐表现曲线">
              <defs>
                <linearGradient id="sparkFill" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="0%" stop-color="rgba(107,170,255,0.36)" />
                  <stop offset="100%" stop-color="rgba(107,170,255,0)" />
                </linearGradient>
              </defs>
              <path class="spark-area" :d="sparkline.area" fill="url(#sparkFill)" />
              <path class="spark-path" :d="sparkline.line" />
              <circle v-for="point in sparkline.dots" :key="point.key" class="spark-dot" :cx="point.x" :cy="point.y" r="4" />
            </svg>
          </div>

          <div class="board-metrics">
            <article v-for="item in boardProof" :key="item.title">
              <p>{{ item.title }}</p>
              <strong>{{ item.value }}</strong>
            </article>
          </div>
        </aside>
      </section>

      <section class="trust-strip">
        <div class="trust-logos">
          <span v-for="item in trustMarks" :key="item">{{ item }}</span>
        </div>
        <div class="trust-stats">
          <article v-for="item in heroStats" :key="item.label">
            <p>{{ item.label }}</p>
            <strong>{{ item.value }}</strong>
            <span>{{ item.note }}</span>
          </article>
        </div>
      </section>

      <section id="solutions" class="role-section">
        <div class="section-head">
          <div>
            <p class="section-kicker">Solutions By Role</p>
            <h2>让不同角色在同一个首页看到各自最关心的入口。</h2>
          </div>
          <p class="section-note">官网首页负责建立信任，再把用户分流到正确的深层页面。</p>
        </div>

        <div class="role-grid">
          <article v-for="item in roleCards" :key="item.title" class="role-card">
            <p>{{ item.kicker }}</p>
            <h3>{{ item.title }}</h3>
            <span>{{ item.desc }}</span>
            <RouterLink class="inline-link" :to="item.path">{{ item.cta }}</RouterLink>
          </article>
        </div>
      </section>

      <section id="capabilities" class="capability-grid">
        <article class="capability-card">
          <div class="section-label">
            <span>股票 / 期货推荐区</span>
            <RouterLink to="/strategies">进入策略页</RouterLink>
          </div>
          <h2>先看主推荐，再决定是否继续追踪观察清单。</h2>

          <div class="watch-list">
            <article v-for="item in watchlist" :key="item.symbol" class="watch-row">
              <div class="watch-top">
                <div>
                  <p>No.{{ item.rank }}</p>
                  <strong>{{ item.symbol }} {{ item.name }}</strong>
                </div>
                <span>{{ item.risk }}</span>
              </div>
              <p class="watch-expected">{{ item.expected }}</p>
              <p class="watch-note">{{ item.note }}</p>
            </article>
          </div>
        </article>

        <article class="capability-card">
          <div class="section-label">
            <span>今日研报解读</span>
            <RouterLink to="/news">进入资讯中心</RouterLink>
          </div>
          <h2>{{ featuredResearch.title }}</h2>
          <p class="card-copy">{{ featuredResearch.summary }}</p>

          <div class="insight-cards">
            <article v-for="item in insightCards" :key="item.title">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>

        <article class="capability-card">
          <div class="section-label">
            <span>会员承接</span>
            <RouterLink to="/membership">进入会员中心</RouterLink>
          </div>
          <h2>{{ vipOffer.name }}</h2>
          <p class="vip-price">{{ vipOffer.price }}</p>
          <p class="card-copy">{{ vipOffer.desc }}</p>

          <div class="stage-list">
            <article v-for="item in conversionStages" :key="item.title" :class="['stage-card', item.state]">
              <div class="stage-top">
                <strong>{{ item.title }}</strong>
                <span>{{ item.badge }}</span>
              </div>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>
      </section>

      <section id="evidence" class="evidence-grid">
        <article class="evidence-card pulse-card">
          <div class="section-label">
            <span>资讯热度时序</span>
            <strong>GET /public/news/articles</strong>
          </div>
          <h2>为什么今天看这些，不只靠一句推荐理由。</h2>
          <div class="pulse-bars">
            <article v-for="item in newsPulse" :key="item.time" class="pulse-bar-item">
              <div class="pulse-bar-track">
                <span class="pulse-bar-fill" :style="{ height: `${item.value}%` }" />
              </div>
              <strong>{{ item.time }}</strong>
              <p>{{ item.label }}</p>
            </article>
          </div>
        </article>

        <article class="evidence-card history-card">
          <div class="section-label">
            <span>历史样本</span>
            <RouterLink to="/archive">进入历史档案</RouterLink>
          </div>
          <h2>把历史命中重新接回今天的阅读链。</h2>
          <div class="history-bars">
            <article v-for="item in historyRows" :key="item.code" class="history-row">
              <div class="history-meta">
                <strong>{{ item.code }}</strong>
                <span>{{ item.date }} · {{ item.risk }}</span>
              </div>
              <div class="history-bar-track">
                <span class="history-bar-fill" :style="{ width: `${item.width}%` }" />
              </div>
              <p class="history-alpha">{{ item.alpha }}</p>
            </article>
          </div>
        </article>
      </section>

      <section class="evidence-grid secondary-grid">
        <article class="evidence-card futures-card">
          <div class="section-label">
            <span>期货方案</span>
            <RouterLink to="/strategies">查看完整方案</RouterLink>
          </div>
          <h2>{{ futuresPlan.name }}</h2>
          <p class="card-copy">{{ futuresPlan.summary }}</p>

          <div class="ladder-chart">
            <div v-for="item in futuresLadder" :key="item.label" class="ladder-row">
              <span>{{ item.label }}</span>
              <div class="ladder-line">
                <i :style="{ left: item.offset }" />
              </div>
              <strong>{{ item.value }}</strong>
            </div>
          </div>

          <div class="method-list">
            <article v-for="item in futuresSteps" :key="item.title">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>

        <article class="evidence-card proof-card">
          <div class="section-label">
            <span>产品价值证明</span>
            <strong>Homepage Mapping</strong>
          </div>
          <div class="proof-grid">
            <article v-for="item in evidenceCards" :key="item.title" class="proof-item">
              <p>{{ item.kicker }}</p>
              <h3>{{ item.title }}</h3>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </article>
      </section>

      <section id="next" class="next-grid">
        <article class="next-card">
          <p class="section-mini">社区承接</p>
          <h2>看完主推荐和研报后，继续去讨论广场补充判断。</h2>
          <p>首页只负责建立信任和分流，观点沉淀和分歧讨论继续留在社区完成。</p>
          <RouterLink class="inline-link" to="/community">进入讨论广场</RouterLink>
        </article>

        <article class="next-card">
          <p class="section-mini">统一检索</p>
          <h2>当今天的焦点不够明确，就让搜索先接手。</h2>
          <p>检索入口承接关键词、资讯、标的和主题，让官网首页维持机构级的清晰度与秩序。</p>
          <RouterLink class="inline-link" to="/search">进入统一检索</RouterLink>
        </article>
      </section>

      <footer class="final-band">
        <div>
          <p class="eyebrow">CONTACT SALES</p>
          <h2>先把公开价值讲清楚，再把用户交给策略、资讯和会员链路。</h2>
        </div>
        <div class="hero-actions">
          <RouterLink class="cta-btn cta-btn-solid" to="/membership">查看会员权益</RouterLink>
          <RouterLink class="cta-btn cta-btn-ghost" to="/strategies">继续查看策略</RouterLink>
        </div>
      </footer>
    </main>
  </div>
</template>

<script setup>
function buildLineChart(points, width = 360, height = 156, padding = 18) {
  const max = Math.max(...points);
  const min = Math.min(...points);
  const range = Math.max(max - min, 1);
  const innerWidth = width - padding * 2;
  const innerHeight = height - padding * 2;
  const dots = points.map((point, index) => {
    const x = padding + (innerWidth * index) / Math.max(points.length - 1, 1);
    const y = padding + innerHeight - ((point - min) / range) * innerHeight;
    return { key: `${index}-${point}`, x, y };
  });
  const line = dots.map((point, index) => `${index === 0 ? "M" : "L"} ${point.x} ${point.y}`).join(" ");
  const last = dots[dots.length - 1];
  const first = dots[0];
  const area = `${line} L ${last.x} ${height - padding} L ${first.x} ${height - padding} Z`;
  return { width, height, dots, line, area };
}

const missionPoints = [
  {
    title: "把今日决策前置",
    desc: "首屏直接承接主推荐与风险边界，不让用户先陷入信息噪音。"
  },
  {
    title: "把资讯解释接上",
    desc: "重点研报不是陪衬，而是用来解释今天为什么要先看这些标的。"
  },
  {
    title: "把转化放到后面",
    desc: "会员入口只在价值已经被看见后出现，不抢前面的阅读链。"
  }
];

const trustMarks = ["机构研究流程", "多接口证据链", "历史样本回看", "会员分层承接", "期货策略扩展", "统一检索入口"];

const heroStats = [
  { label: "主推荐评分", value: "92", note: "先看强度，再决定是否进入策略详情。" },
  { label: "观察清单", value: "4 只", note: "首页优先聚焦最值得继续跟踪的对象。" },
  { label: "资讯线索", value: "3 条", note: "焦点研报解释为什么今天先看这些标的。" },
  { label: "历史样本", value: "6 条", note: "历史命中率重新接回今天的阅读链。" }
];

const roleCards = [
  {
    kicker: "For Research Teams",
    title: "投研团队",
    desc: "更关心主推荐、解释链和历史样本，适合直接进入策略与档案页继续判断。",
    path: "/strategies",
    cta: "进入策略页"
  },
  {
    kicker: "For Advisors",
    title: "财富顾问",
    desc: "更关心资讯解释和会员转化路径，适合把首页作为展示和讲解入口。",
    path: "/news",
    cta: "进入资讯中心"
  },
  {
    kicker: "For Individual Users",
    title: "个人研究者",
    desc: "更关心检索、关注和社区承接，适合从首页分流到统一检索与讨论广场。",
    path: "/search",
    cta: "进入统一检索"
  }
];

const primaryStock = {
  symbol: "300750",
  name: "宁德时代",
  risk: "中风险",
  score: "92 / 100",
  expected: "目标区间 212 - 226 · 风控线 198",
  reason: "储能链订单确认继续强化，今天更适合先验证强势延续，而不是追逐弱修复。"
};

const sparkline = buildLineChart([18, 24, 23, 31, 30, 39, 37, 46]);

const boardProof = [
  { title: "解释证据", value: "版本已补齐" },
  { title: "收益序列", value: "8 段点列" },
  { title: "行动建议", value: "先确认强弱" }
];

const watchlist = [
  {
    rank: 1,
    symbol: "300750",
    name: "宁德时代",
    risk: "中风险",
    expected: "预期区间 212 - 226",
    note: "优先确认强势延续是否继续获得订单与情绪双重支持。"
  },
  {
    rank: 2,
    symbol: "600036",
    name: "招商银行",
    risk: "低风险",
    expected: "预期区间 37.8 - 39.6",
    note: "作为低波动防守线，适合做资金风险对冲参考。"
  },
  {
    rank: 3,
    symbol: "601899",
    name: "紫金矿业",
    risk: "中风险",
    expected: "预期区间 18.4 - 19.8",
    note: "金属价格韧性仍在，但需要继续盯住盘中量能确认。"
  },
  {
    rank: 4,
    symbol: "002371",
    name: "北方华创",
    risk: "中高风险",
    expected: "预期区间 334 - 356",
    note: "弹性更强，但必须把风险边界和止损纪律前置。"
  }
];

const featuredResearch = {
  title: "储能链与高股息同时占优，市场在寻找更确定的胜率资产",
  summary: "首页重点内容不再只是摆一篇文章，而是明确承接今日主推荐、资讯线索和后续社区讨论。"
};

const newsPulse = [
  { time: "08:30", label: "盘前导读", value: 46 },
  { time: "10:20", label: "热点确认", value: 72 },
  { time: "13:15", label: "午后分歧", value: 58 },
  { time: "15:30", label: "收盘复盘", value: 88 }
];

const insightCards = [
  {
    title: "这篇内容在说什么",
    desc: "把今天的主推荐放到更大的行业与情绪框架里解释，而不是单点推荐。"
  },
  {
    title: "它对今天推荐的影响",
    desc: "资讯继续强化储能与防守双主线，因此首页的推荐优先级也相应收敛。"
  },
  {
    title: "今天应该怎么用",
    desc: "先读摘要，再决定去资讯中心看全文，最后回到策略页确认动作边界。"
  }
];

const vipOffer = {
  name: "研究会员月度版",
  price: "¥199 / 月",
  desc: "解锁完整推荐池、午盘变化提醒、收盘复盘和更高配额。"
};

const conversionStages = [
  { title: "游客", badge: "当前入口", desc: "先看公开首页、主推荐和最新资讯。", state: "neutral" },
  { title: "注册用户", badge: "可保存关注", desc: "开始建立个人观察清单与回看路径。", state: "active" },
  { title: "会员", badge: "完整内容", desc: "进入完整推荐、复盘与更深层解释链。", state: "vip" }
];

const evidenceCards = [
  {
    kicker: "Why this homepage",
    title: "主推荐不是孤立信号",
    desc: "首页直接承接推荐理由、资讯线索和风险边界，让用户先理解再动作。"
  },
  {
    kicker: "Why this research",
    title: "重点研报负责解释逻辑是否延续",
    desc: "不是重复推荐，而是帮助判断今天的主线是否仍在强化。"
  },
  {
    kicker: "Why keep archives",
    title: "命中与失效都要留下复盘痕迹",
    desc: "这样首页才不是一次性推荐页，而是可持续的研究入口。"
  },
  {
    kicker: "Why conversion later",
    title: "转化放在阅读链之后",
    desc: "先让用户理解价值，再承接权益升级，不在首屏强推。"
  }
];

const historyRows = [
  { code: "300750", date: "06-18", risk: "中风险", alpha: "+12.4%", width: 84 },
  { code: "601899", date: "06-12", risk: "中风险", alpha: "+9.1%", width: 68 },
  { code: "600036", date: "06-05", risk: "低风险", alpha: "+5.8%", width: 49 },
  { code: "002371", date: "05-29", risk: "中高风险", alpha: "+14.2%", width: 92 },
  { code: "688041", date: "05-23", risk: "高风险", alpha: "+7.6%", width: 57 },
  { code: "603986", date: "05-15", risk: "中风险", alpha: "+10.3%", width: 74 }
];

const futuresPlan = {
  name: "RB2510 - HC2510 套利方案",
  summary: "首页只展示最关键的进场、平仓、止损区间，详细执行仍然留给深层策略页面。"
};

const futuresLadder = [
  { label: "止损", value: "-84", offset: "18%" },
  { label: "触发", value: "-46", offset: "42%" },
  { label: "目标一", value: "-12", offset: "68%" },
  { label: "目标二", value: "+8", offset: "82%" }
];

const futuresSteps = [
  { title: "步骤 1：确认触发区间", desc: "只有在价差进入触发区间且成交量同步配合时才执行首笔建仓。" },
  { title: "步骤 2：写入平仓目标", desc: "目标区间不靠主观推测，先写好，再决定是否分批止盈。" },
  { title: "步骤 3：严格执行止损", desc: "当价差进入失效区间时立即止损，不把首页展示误认为交易建议。" }
];
</script>

<style scoped>
.corp-home-page {
  --bg: #09111d;
  --bg-soft: #101927;
  --panel: #121d2c;
  --panel-soft: #182538;
  --ink: #edf3ff;
  --ink-soft: #9aaac2;
  --line: #273750;
  --line-strong: #355175;
  --accent: #6baaff;
  --accent-soft: #314f77;
  --cta: #ca8440;

  min-height: 100%;
  color: var(--ink);
  background:
    radial-gradient(circle at 0% 0%, rgba(107, 170, 255, 0.14), transparent 28%),
    radial-gradient(circle at 100% 10%, rgba(202, 132, 64, 0.12), transparent 26%),
    linear-gradient(180deg, var(--bg) 0%, var(--bg-soft) 100%);
}

.page-grid {
  position: fixed;
  inset: 0;
  pointer-events: none;
  opacity: 0.14;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.025) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.025) 1px, transparent 1px);
  background-size: 80px 80px;
  mask-image: radial-gradient(circle at center, black 55%, transparent 100%);
}

.gateway-nav,
.hero-copy,
.hero-board,
.trust-strip,
.role-section,
.capability-card,
.evidence-card,
.next-card,
.final-band {
  position: relative;
  z-index: 1;
  border: 1px solid var(--line);
  background: rgba(17, 27, 41, 0.88);
  backdrop-filter: blur(8px);
}

.gateway-nav {
  width: min(1420px, calc(100% - 32px));
  margin: 14px auto 0;
  padding: 14px 16px;
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 16px;
  box-shadow: 0 18px 44px rgba(0, 0, 0, 0.28);
}

.brand-lockup {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

.brand-mark {
  width: 34px;
  height: 34px;
  display: grid;
  place-items: center;
  border-radius: 10px;
  color: #08121e;
  font-weight: 800;
  background: linear-gradient(145deg, var(--accent), #c8e1ff 68%, var(--cta));
}

.brand-title {
  margin: 0;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.04em;
}

.brand-meta {
  margin: 2px 0 0;
  color: var(--ink-soft);
  font-size: 11px;
  font-family: "IBM Plex Mono", "SFMono-Regular", Menlo, Consolas, monospace;
}

.gateway-links {
  display: flex;
  justify-content: center;
  gap: 8px;
  flex-wrap: wrap;
}

.gateway-links a {
  color: #dce6f8;
  text-decoration: none;
  font-size: 12px;
  padding: 6px 12px;
  border: 1px solid #2f425f;
  border-radius: 999px;
  background: rgba(17, 28, 43, 0.84);
}

.gateway-links a:hover {
  border-color: #5e88bf;
}

.nav-actions,
.hero-actions {
  display: inline-flex;
  gap: 8px;
  flex-wrap: wrap;
}

.nav-btn,
.cta-btn {
  min-height: 40px;
  padding: 0 16px;
  border-radius: 999px;
  text-decoration: none;
  font-size: 13px;
  font-weight: 700;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

.nav-btn-ghost,
.cta-btn-ghost {
  color: #dce6f8;
  border: 1px solid #324865;
  background: rgba(15, 24, 37, 0.94);
}

.nav-btn-solid,
.cta-btn-solid {
  color: #09121f;
  border: 1px solid transparent;
  background: linear-gradient(145deg, var(--accent), #d5e8ff 68%, var(--cta));
}

.site-shell {
  width: min(1420px, calc(100% - 32px));
  margin: 0 auto;
  padding: 14px 0 28px;
  display: grid;
  gap: 14px;
}

.hero-section {
  display: grid;
  grid-template-columns: 1.25fr 0.85fr;
  gap: 12px;
}

.hero-copy,
.hero-board,
.role-section,
.capability-card,
.evidence-card,
.next-card,
.final-band,
.trust-strip {
  padding: 18px;
}

.hero-copy {
  background:
    radial-gradient(circle at 100% 0%, rgba(107, 170, 255, 0.12), transparent 32%),
    linear-gradient(180deg, rgba(19, 29, 45, 0.96), rgba(15, 24, 37, 0.96));
}

.eyebrow,
.section-kicker,
.section-mini,
.board-kicker,
.section-label span {
  margin: 0;
  color: #9ec9fb;
  font-size: 11px;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  font-weight: 700;
}

.hero-copy h1,
.section-head h2,
.capability-card h2,
.evidence-card h2,
.next-card h2,
.final-band h2,
.hero-board h2 {
  font-family: "IBM Plex Sans", "PingFang SC", "Microsoft YaHei", sans-serif;
}

.hero-copy h1 {
  margin: 12px 0 0;
  max-width: 760px;
  font-size: clamp(40px, 5.5vw, 72px);
  line-height: 0.94;
  letter-spacing: -0.04em;
  font-weight: 600;
}

.hero-desc,
.section-note,
.card-copy,
.next-card p {
  margin: 16px 0 0;
  max-width: 760px;
  color: var(--ink-soft);
  line-height: 1.7;
  font-size: 14px;
}

.hero-desc code,
.section-label strong,
.vip-foot code,
.chart-head strong {
  color: #d8e8ff;
  font-family: "IBM Plex Mono", "SFMono-Regular", Menlo, Consolas, monospace;
  font-size: 11px;
}

.hero-desc code {
  padding: 1px 6px;
  border-radius: 6px;
  border: 1px solid #355171;
  background: rgba(12, 24, 39, 0.88);
}

.hero-actions {
  margin-top: 18px;
}

.mission-grid {
  margin-top: 22px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.mission-card {
  border: 1px solid #2d415f;
  background: rgba(17, 29, 45, 0.75);
  padding: 12px;
}

.mission-card strong {
  font-size: 14px;
}

.mission-card p,
.proof-item span,
.method-list p,
.insight-cards p,
.watch-note,
.history-meta span,
.stage-card p {
  margin: 7px 0 0;
  color: var(--ink-soft);
  font-size: 12px;
  line-height: 1.55;
}

.hero-board {
  background:
    radial-gradient(circle at 100% 0%, rgba(202, 132, 64, 0.14), transparent 30%),
    linear-gradient(180deg, rgba(18, 27, 42, 0.96), rgba(14, 22, 34, 0.96));
}

.board-head,
.section-head,
.section-label,
.stage-top {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: baseline;
}

.hero-board h2 {
  margin: 8px 0 0;
  font-size: 30px;
  font-weight: 600;
}

.risk-pill {
  color: #f6d8a7;
  background: rgba(171, 117, 56, 0.18);
  border: 1px solid #805a2e;
  border-radius: 999px;
  padding: 5px 10px;
  font-size: 12px;
}

.board-score {
  margin: 20px 0 0;
  font-size: 52px;
  line-height: 1;
  letter-spacing: -0.04em;
  font-family: "IBM Plex Mono", "SFMono-Regular", Menlo, Consolas, monospace;
}

.board-range {
  margin: 8px 0 0;
  color: #d7e4f8;
  font-size: 14px;
}

.board-copy {
  margin: 12px 0 0;
  color: var(--ink-soft);
  font-size: 13px;
  line-height: 1.65;
}

.chart-shell {
  margin-top: 16px;
  border: 1px solid #2d415e;
  background: rgba(16, 28, 43, 0.82);
  padding: 12px;
}

.chart-head {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: baseline;
  color: #9fb4cf;
  font-size: 11px;
}

.spark-chart {
  width: 100%;
  height: 156px;
  margin-top: 8px;
}

.spark-path {
  fill: none;
  stroke: var(--accent);
  stroke-width: 3;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.spark-dot {
  fill: #dcecff;
  stroke: var(--accent);
  stroke-width: 2;
}

.board-metrics {
  margin-top: 14px;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.board-metrics article,
.trust-stats article {
  border: 1px solid #2f4361;
  background: rgba(16, 27, 41, 0.72);
  padding: 10px;
}

.board-metrics p,
.trust-stats p {
  margin: 0;
  color: #95a8c4;
  font-size: 11px;
}

.board-metrics strong,
.trust-stats strong {
  display: block;
  margin-top: 6px;
  font-size: 14px;
}

.trust-strip {
  display: grid;
  grid-template-columns: 1.2fr 1.8fr;
  gap: 12px;
}

.trust-logos {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-content: start;
}

.trust-logos span {
  color: #dce7f8;
  font-size: 12px;
  padding: 7px 10px;
  border: 1px solid #2f425f;
  background: rgba(17, 29, 45, 0.82);
}

.trust-stats {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.trust-stats span {
  display: block;
  margin-top: 7px;
  color: var(--ink-soft);
  font-size: 11px;
  line-height: 1.45;
}

.role-grid,
.capability-grid,
.evidence-grid,
.next-grid,
.proof-grid {
  display: grid;
  gap: 12px;
}

.role-section {
  display: grid;
  gap: 16px;
}

.section-head h2 {
  margin: 8px 0 0;
  font-size: 36px;
  line-height: 1.02;
  font-weight: 600;
}

.role-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.role-card {
  border: 1px solid #30425f;
  padding: 14px;
  background: rgba(18, 29, 44, 0.76);
}

.role-card p {
  margin: 0;
  color: #96abca;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.role-card h3,
.proof-item h3 {
  margin: 10px 0 0;
  font-size: 20px;
}

.role-card span {
  display: block;
  margin-top: 8px;
  color: var(--ink-soft);
  font-size: 13px;
  line-height: 1.6;
}

.capability-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.capability-card h2,
.evidence-card h2,
.next-card h2 {
  margin: 12px 0 0;
  font-size: 30px;
  line-height: 1.04;
  font-weight: 600;
}

.watch-list,
.insight-cards,
.stage-list,
.method-list {
  margin-top: 16px;
  display: grid;
  gap: 10px;
}

.watch-row,
.insight-cards article,
.stage-card,
.method-list article,
.proof-item {
  border-top: 1px solid #2d405d;
  padding-top: 10px;
}

.watch-top {
  display: flex;
  justify-content: space-between;
  gap: 10px;
}

.watch-top p,
.watch-top span {
  margin: 0;
  color: #8ea5c3;
  font-size: 11px;
}

.watch-top strong,
.insight-cards strong,
.method-list strong,
.stage-top strong {
  display: block;
  margin-top: 5px;
  font-size: 15px;
}

.watch-expected {
  margin: 8px 0 0;
  color: #e0efff;
  font-size: 13px;
}

.vip-price {
  margin: 14px 0 0;
  font-size: 38px;
  line-height: 1;
  color: #f4d8a7;
  letter-spacing: -0.03em;
}

.stage-top span {
  color: #9ab2d2;
  font-size: 11px;
}

.stage-card.neutral {
  border-color: #31435f;
}

.stage-card.active {
  border-color: #517db3;
}

.stage-card.vip {
  border-color: #816338;
}

.evidence-grid {
  grid-template-columns: 1.02fr 0.98fr;
}

.secondary-grid {
  grid-template-columns: 0.95fr 1.05fr;
}

.pulse-bars {
  margin-top: 18px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  align-items: end;
}

.pulse-bar-item {
  display: grid;
  gap: 7px;
}

.pulse-bar-track {
  height: 132px;
  border: 1px solid #30415d;
  background: linear-gradient(180deg, rgba(10, 16, 26, 0.28), rgba(10, 16, 26, 0.82));
  display: flex;
  align-items: end;
  padding: 6px;
}

.pulse-bar-fill {
  width: 100%;
  display: block;
  background: linear-gradient(180deg, #8fccff, #42678f 74%, #2a3c55);
}

.pulse-bar-item strong {
  font-size: 12px;
}

.pulse-bar-item p {
  margin: 0;
  color: var(--ink-soft);
  font-size: 11px;
}

.history-bars {
  margin-top: 18px;
  display: grid;
  gap: 12px;
}

.history-row {
  display: grid;
  grid-template-columns: 148px 1fr 72px;
  gap: 12px;
  align-items: center;
}

.history-meta strong {
  display: block;
  font-size: 14px;
}

.history-bar-track {
  height: 14px;
  border: 1px solid #2f3f58;
  background: rgba(10, 17, 28, 0.82);
  overflow: hidden;
}

.history-bar-fill {
  display: block;
  height: 100%;
  background: linear-gradient(90deg, #4f7cb0, #87cbff 72%, #d59c57);
}

.history-alpha {
  margin: 0;
  text-align: right;
  color: #d7e8ff;
  font-family: "IBM Plex Mono", "SFMono-Regular", Menlo, Consolas, monospace;
  font-size: 13px;
}

.ladder-chart {
  margin-top: 18px;
  display: grid;
  gap: 10px;
}

.ladder-row {
  display: grid;
  grid-template-columns: 64px 1fr 64px;
  gap: 10px;
  align-items: center;
}

.ladder-row span,
.ladder-row strong {
  font-size: 12px;
}

.ladder-line {
  position: relative;
  height: 12px;
  border-top: 1px solid #385173;
  border-bottom: 1px solid #213247;
  background: linear-gradient(90deg, rgba(107, 170, 255, 0.07), rgba(202, 132, 64, 0.07));
}

.ladder-line i {
  position: absolute;
  top: -4px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  transform: translateX(-50%);
  background: linear-gradient(145deg, var(--accent), var(--cta));
  box-shadow: 0 0 0 4px rgba(107, 170, 255, 0.1);
}

.proof-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin-top: 16px;
}

.proof-item p {
  margin: 0;
  color: #95aac9;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.12em;
}

.next-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.next-card {
  min-height: 220px;
  display: grid;
  align-content: start;
}

.inline-link {
  margin-top: 18px;
  display: inline-flex;
  color: #d9e8ff;
  text-decoration: none;
  font-size: 13px;
  font-weight: 700;
}

.final-band {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 16px;
  align-items: center;
  background:
    radial-gradient(circle at 0% 0%, rgba(107, 170, 255, 0.12), transparent 34%),
    linear-gradient(145deg, rgba(17, 28, 43, 0.96), rgba(14, 22, 34, 0.96));
}

.final-band h2 {
  margin: 10px 0 0;
  font-size: 40px;
  line-height: 1.02;
  font-weight: 600;
}

@media (max-width: 1240px) {
  .hero-section,
  .trust-strip,
  .capability-grid,
  .evidence-grid,
  .secondary-grid,
  .next-grid,
  .role-grid {
    grid-template-columns: 1fr;
  }

  .mission-grid,
  .trust-stats,
  .board-metrics,
  .proof-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 920px) {
  .gateway-nav,
  .final-band {
    grid-template-columns: 1fr;
  }

  .history-row,
  .mission-grid,
  .trust-stats,
  .board-metrics,
  .pulse-bars,
  .proof-grid {
    grid-template-columns: 1fr;
  }

  .history-row {
    gap: 8px;
  }
}

@media (max-width: 720px) {
  .gateway-nav,
  .site-shell {
    width: calc(100% - 16px);
  }

  .hero-copy,
  .hero-board,
  .trust-strip,
  .role-section,
  .capability-card,
  .evidence-card,
  .next-card,
  .final-band {
    padding: 14px;
  }

  .hero-copy h1,
  .final-band h2 {
    font-size: clamp(34px, 11vw, 54px);
  }

  .section-head h2,
  .capability-card h2,
  .evidence-card h2,
  .next-card h2 {
    font-size: 28px;
  }
}
</style>
