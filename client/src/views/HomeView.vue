<template>
  <div class="home-page fade-up">
    <section class="hero-stage card">
      <div class="hero-copy">
        <p class="hero-kicker">首页重点 · 推荐股票 + 期货套利</p>
        <h1>先看历史胜率，再看今天怎么做。</h1>
        <p class="hero-desc">
          首页已对接 API：历史推荐股票涨跌幅对比、今日推荐预期区间、期货套利进出场点位和操作方法同屏展示。
        </p>
        <div class="hero-actions">
          <button class="primary-btn" type="button" :disabled="loading" @click="loadHomeData">
            {{ loading ? "同步中..." : "刷新推荐数据" }}
          </button>
          <button class="ghost-btn" type="button" @click="scrollToArbitrage">查看套利计划</button>
          <button class="ghost-btn" type="button" @click="scrollToInsights">查看资讯重点</button>
        </div>
        <div class="hero-tags">
          <span v-for="tag in heroTags" :key="tag">{{ tag }}</span>
        </div>
        <p v-if="loading" class="api-state">正在同步实时数据...</p>
        <p v-else-if="loadError" class="api-state warning">{{ loadError }}</p>
        <p v-else class="api-state">数据更新时间：{{ lastUpdatedAt || "-" }}</p>
      </div>

      <aside class="hero-panel">
        <div class="panel-head">
          <p>今日摘要</p>
          <strong>{{ todayDateText }}</strong>
        </div>
        <div class="panel-stats">
          <article v-for="item in panelStats" :key="item.label">
            <p>{{ item.label }}</p>
            <strong>{{ item.value }}</strong>
          </article>
        </div>
        <ul class="task-list">
          <li v-for="task in tasks" :key="task.title">
            <span class="task-dot" :class="task.level" />
            <div>
              <p class="task-title">{{ task.title }}</p>
              <p class="task-note">{{ task.note }}</p>
            </div>
          </li>
        </ul>
      </aside>
    </section>

    <section class="home-search-banner card">
      <div class="search-banner-head">
        <div>
          <p class="search-kicker">首页搜索</p>
          <h2>快速检索资讯、研报与公告线索</h2>
        </div>
        <button type="button" class="search-jump-btn" @click="openNewsModule">查看完整资讯库</button>
      </div>
      <form class="insight-search" @submit.prevent="handleInsightSearch">
        <input
          v-model.trim="insightKeyword"
          type="text"
          maxlength="40"
          placeholder="输入关键词立即搜索（如：新能源、券商研报、上市公司公告）"
        />
        <button type="submit" :disabled="newsLoading">搜索</button>
        <button
          type="button"
          class="ghost"
          :disabled="newsLoading || (!appliedInsightKeyword && !insightKeyword)"
          @click="clearInsightSearch"
        >
          清空
        </button>
      </form>
      <p class="search-tip">
        <template v-if="appliedInsightKeyword">当前关键词：{{ appliedInsightKeyword }}</template>
        <template v-else>未设置关键词，将展示最新资讯</template>
      </p>
    </section>

    <section class="mobile-quick card">
      <button v-for="item in mobileQuickActions" :key="item.title" type="button">
        <p>{{ item.title }}</p>
        <span>{{ item.desc }}</span>
      </button>
    </section>

    <section id="insight-section" class="insight-section" :class="{ 'single-column': !shouldShowVipPromo }">
      <article class="card insight-card">
        <header class="insight-head">
          <div>
            <h2 class="section-title">资讯焦点</h2>
            <p class="section-subtitle">同步新闻、研报、期刊，让行情判断和信息判断在首页闭环。</p>
          </div>
          <div class="insight-actions">
            <button type="button" :disabled="newsLoading" @click="loadHomeInsights">
              {{ newsLoading ? "同步中..." : "刷新资讯" }}
            </button>
            <button type="button" class="ghost" @click="openNewsModule">进入资讯中心</button>
          </div>
        </header>
        <p v-if="newsLoading" class="insight-status">正在同步资讯内容...</p>
        <p v-else-if="newsError" class="insight-status warning">{{ newsError }}</p>
        <p v-else class="insight-status">
          资讯更新时间：{{ newsUpdatedAt || "-" }}
          <template v-if="appliedInsightKeyword"> · 当前关键词：{{ appliedInsightKeyword }}</template>
        </p>

        <div class="insight-tags" v-if="newsCategoryTags.length">
          <span v-for="tag in newsCategoryTags" :key="tag">{{ tag }}</span>
        </div>

        <div v-if="newsHighlights.length > 0" class="insight-list">
          <article v-for="item in newsHighlights" :key="item.id" class="insight-item" @click="openNewsModule">
            <div class="insight-item-head">
              <span class="category">{{ item.category }}</span>
              <span class="time">{{ item.time }}</span>
            </div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.summary }}</p>
            <div class="insight-meta">
              <span>{{ item.source }}</span>
              <span class="vip-mark" :class="{ active: item.visibility === 'VIP' }">
                {{ item.visibility === "VIP" ? "VIP专享" : "公开" }}
              </span>
            </div>
          </article>
        </div>
        <div v-else class="empty-box">暂无资讯内容</div>
      </article>

      <aside v-if="shouldShowVipPromo" class="card vip-promo-card">
        <p class="promo-kicker">会员广告位</p>
        <h2>开通 VIP，解锁深度研判和专享内容。</h2>
        <p class="promo-price">{{ vipPromoPriceText }}</p>
        <p class="promo-note">{{ vipPromoDesc }}</p>
        <ul class="promo-benefits">
          <li>VIP文章详情与附件下载</li>
          <li>优先策略推送与事件提醒</li>
          <li>更高阅读与订阅配额</li>
        </ul>
        <button type="button" class="promo-btn" @click="goMembershipCenter">
          {{ vipPromoCTA }}
        </button>
      </aside>
    </section>

    <section class="stock-section">
      <article class="card history-card">
        <header>
          <h2 class="section-title">历史推荐股票涨跌幅对比</h2>
          <p class="section-subtitle">
            基于推荐历史收益点计算累计涨跌幅；基准表现为同周期模拟值，后续可接入指数行情 API。
          </p>
        </header>

        <div class="table-wrap" v-if="historyStocks.length">
          <table class="stock-table">
            <thead>
              <tr>
                <th>股票</th>
                <th>推荐日期</th>
                <th>综合评分</th>
                <th>风险等级</th>
                <th>历史涨跌幅</th>
                <th>基准表现</th>
                <th>超额收益</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in historyStocks" :key="`${row.code}-${row.date}`">
                <td>{{ row.code }}</td>
                <td>{{ row.date }}</td>
                <td>{{ row.score }}</td>
                <td>{{ row.risk }}</td>
                <td :class="getTrendClass(row.change)">{{ row.change }}</td>
                <td>{{ row.benchmark }}</td>
                <td :class="getTrendClass(row.alpha)">{{ row.alpha }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="empty-box">暂无历史推荐数据</div>

        <div class="mobile-history-list">
          <article v-for="row in historyStocks" :key="`mobile-${row.code}-${row.date}`" class="mobile-history-item">
            <div class="top">
              <p>{{ row.code }}</p>
              <span>{{ row.date }}</span>
            </div>
            <div class="line">
              <span>评分 {{ row.score }}</span>
              <span>{{ row.risk }}</span>
            </div>
            <div class="line">
              <span :class="getTrendClass(row.change)">涨跌幅 {{ row.change }}</span>
              <span :class="getTrendClass(row.alpha)">超额 {{ row.alpha }}</span>
            </div>
          </article>
        </div>
      </article>

      <article class="card today-card">
        <header>
          <h2 class="section-title">今日推荐股票预期</h2>
          <p class="section-subtitle">结合推荐详情（止盈/止损）和历史表现生成预期区间。</p>
        </header>

        <div class="today-list" v-if="todayStocks.length">
          <article v-for="item in todayStocks" :key="item.code" class="today-item">
            <div class="today-main">
              <p class="code">{{ item.code }}</p>
              <p class="name">{{ item.name }}</p>
            </div>
            <p class="expect" :class="item.side">{{ item.expected }}</p>
            <p class="condition">{{ item.condition }}</p>
          </article>
        </div>
        <div v-else class="empty-box">暂无今日推荐数据</div>
      </article>
    </section>

    <section class="futures-section" id="arbitrage-section">
      <article class="card arbitrage-card">
        <header>
          <h2 class="section-title">期货套利方案</h2>
          <p class="section-subtitle">来自期货套利接口，展示进入点位、平仓点位、止损和仓位建议。</p>
        </header>

        <div class="arbitrage-list" v-if="arbitragePlans.length">
          <article v-for="plan in arbitragePlans" :key="plan.name" class="plan-item">
            <div class="plan-title">
              <p>{{ plan.name }}</p>
              <span>{{ plan.tag }}</span>
            </div>
            <div class="plan-grid">
              <div>
                <p>进入点位</p>
                <strong>{{ plan.entry }}</strong>
              </div>
              <div>
                <p>平仓点位</p>
                <strong>{{ plan.close }}</strong>
              </div>
              <div>
                <p>止损点位</p>
                <strong>{{ plan.stop }}</strong>
              </div>
              <div>
                <p>建议仓位</p>
                <strong>{{ plan.position }}</strong>
              </div>
            </div>
          </article>
        </div>
        <div v-else class="empty-box">暂无期货套利计划</div>
      </article>

      <aside class="card method-card">
        <header>
          <h2 class="section-title">操作方法</h2>
          <p class="section-subtitle">结合期货指导接口，固定流程，降低临盘决策失误。</p>
        </header>
        <ol class="method-list">
          <li v-for="step in methods" :key="step.title">
            <p class="step-title">{{ step.title }}</p>
            <p class="step-desc">{{ step.desc }}</p>
          </li>
        </ol>
        <p class="risk-tip">
          风险提示：接口已接入，正式实盘前仍需联调行情源、撮合规则和风控阈值。
        </p>
      </aside>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import {
  getFuturesGuidance,
  getStockRecommendationDetail,
  getStockRecommendationPerformance,
  listFuturesArbitrage,
  listStockRecommendations
} from "../api/market";
import { listNewsArticles, listNewsCategories } from "../api/news";
import { getMembershipQuota, listMembershipProducts } from "../api/membership";
import { useClientAuth } from "../lib/client-auth";
import { shouldUseDemoFallback } from "../lib/fallback-policy";

const fallbackStockRecommendations = [
  {
    id: "sr_local_001",
    symbol: "600036.SH",
    name: "招商银行",
    score: 91.2,
    risk_level: "MEDIUM",
    valid_from: "2026-02-10T09:00:00+08:00",
    reason_summary: "量价结构共振"
  },
  {
    id: "sr_local_002",
    symbol: "300750.SZ",
    name: "宁德时代",
    score: 88.5,
    risk_level: "MEDIUM",
    valid_from: "2026-02-12T09:00:00+08:00",
    reason_summary: "趋势延续，资金净流入"
  },
  {
    id: "sr_local_003",
    symbol: "601012.SH",
    name: "隆基绿能",
    score: 82.4,
    risk_level: "HIGH",
    valid_from: "2026-02-17T09:00:00+08:00",
    reason_summary: "反弹确认阶段"
  },
  {
    id: "sr_local_004",
    symbol: "000858.SZ",
    name: "五粮液",
    score: 86.8,
    risk_level: "LOW",
    valid_from: "2026-02-19T09:00:00+08:00",
    reason_summary: "估值修复 + 防御属性"
  }
];

const fallbackDetails = {
  sr_local_001: { take_profit: "上涨6%-10%分批止盈", stop_loss: "跌破 33.80 止损" },
  sr_local_002: { take_profit: "上涨5%-9%分批止盈", stop_loss: "跌破 139.20 止损" },
  sr_local_003: { take_profit: "上涨3%-6%分批止盈", stop_loss: "跌破 21.50 止损" },
  sr_local_004: { take_profit: "上涨4%-7%分批止盈", stop_loss: "跌破 145.00 止损" }
};

const fallbackPerformances = {
  sr_local_001: [
    { date: "2026-02-11", return: 0.018 },
    { date: "2026-02-12", return: 0.011 },
    { date: "2026-02-13", return: 0.012 }
  ],
  sr_local_002: [
    { date: "2026-02-13", return: 0.009 },
    { date: "2026-02-14", return: 0.013 },
    { date: "2026-02-15", return: 0.011 }
  ],
  sr_local_003: [
    { date: "2026-02-18", return: -0.011 },
    { date: "2026-02-19", return: -0.008 },
    { date: "2026-02-20", return: 0.003 }
  ],
  sr_local_004: [
    { date: "2026-02-20", return: 0.010 },
    { date: "2026-02-21", return: 0.007 },
    { date: "2026-02-22", return: 0.012 }
  ]
};

const fallbackArbitragePlans = [
  {
    id: "arb_local_001",
    type: "CALENDAR",
    contract_a: "IF 主连",
    contract_b: "IC 主连",
    entry_point: 186,
    exit_point: 132,
    stop_point: 208,
    percentile: 0.86,
    status: "WATCH"
  },
  {
    id: "arb_local_002",
    type: "CROSS",
    contract_a: "IH 主连",
    contract_b: "IF 主连",
    entry_point: -76,
    exit_point: -42,
    stop_point: -94,
    percentile: 0.79,
    status: "WATCH"
  }
];

const fallbackGuidance = {
  contract: "IF 主连",
  guidance_direction: "LONG_SPREAD",
  position_level: "LIGHT",
  entry_range: "120-150",
  take_profit_range: "80-100",
  stop_loss_range: "165-180",
  risk_level: "MEDIUM",
  invalid_condition: "相关性失效",
  valid_to: "2026-03-01T15:00:00+08:00"
};

const fallbackNewsHighlights = [
  {
    id: "news_local_001",
    category: "新闻",
    time: "2026/02/28 09:30:00",
    title: "盘前资金风格切换，银行与高股息板块走强。",
    summary: "市场风险偏好回落，早盘更关注防御属性板块轮动。",
    source: "编辑部",
    visibility: "PUBLIC"
  },
  {
    id: "news_local_002",
    category: "研报",
    time: "2026/02/28 08:55:00",
    title: "机构上调两融余额敏感行业评级。",
    summary: "研报建议关注景气向上赛道的估值切换窗口。",
    source: "研究团队",
    visibility: "VIP"
  },
  {
    id: "news_local_003",
    category: "期刊",
    time: "2026/02/27 21:10:00",
    title: "周度策略期刊：波动率管理与仓位纪律。",
    summary: "强调先控制回撤，再放大胜率，维持组合弹性。",
    source: "策略周刊",
    visibility: "PUBLIC"
  }
];

const fallbackVipProducts = [
  {
    id: "mp_local_vip_month",
    name: "VIP月卡",
    price: 99,
    duration_days: 30,
    status: "ACTIVE",
    member_level: "VIP1"
  }
];

const useDemoFallback = shouldUseDemoFallback();

const router = useRouter();
const { isLoggedIn } = useClientAuth();

const loading = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const newsLoading = ref(false);
const newsError = ref("");
const newsUpdatedAt = ref("");
const insightKeyword = ref("");
const appliedInsightKeyword = ref("");
const newsCategoryTags = ref([]);
const newsHighlights = ref(useDemoFallback ? [...fallbackNewsHighlights] : []);
const vipStateLoading = ref(false);
const isVIPUser = ref(false);
const vipOffer = ref(useDemoFallback ? fallbackVipProducts[0] : null);

const rawStockRecommendations = ref(useDemoFallback ? fallbackStockRecommendations : []);
const stockDetailMap = ref(useDemoFallback ? { ...fallbackDetails } : {});
const stockPerformanceMap = ref(useDemoFallback ? { ...fallbackPerformances } : {});
const rawArbitragePlans = ref(useDemoFallback ? fallbackArbitragePlans : []);
const futuresGuidance = ref(useDemoFallback ? fallbackGuidance : null);

const todayDateText = computed(() =>
  new Date().toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" })
);

const historyStocks = computed(() =>
  rawStockRecommendations.value.slice(0, 6).map((item) => {
    const points = stockPerformanceMap.value[item.id] || [];
    const historyReturn = calcCumulativeReturn(points);
    const benchmarkReturn = estimateBenchmarkReturn(historyReturn);
    const alphaReturn =
      historyReturn === null || benchmarkReturn === null ? null : historyReturn - benchmarkReturn;

    return {
      code: item.symbol || item.id || "-",
      date: formatDate(item.valid_from),
      score: formatScore(item.score),
      risk: mapRiskLevel(item.risk_level),
      change: formatPercent(historyReturn),
      benchmark: formatPercent(benchmarkReturn),
      alpha: formatPercent(alphaReturn),
      rawChange: historyReturn
    };
  })
);

const todayStocks = computed(() =>
  rawStockRecommendations.value.slice(0, 3).map((item) => {
    const detail = stockDetailMap.value[item.id] || {};
    const historyReturn = calcCumulativeReturn(stockPerformanceMap.value[item.id] || []);
    const expected = inferExpectedRange(detail.take_profit, historyReturn);
    const side = inferExpectedSide(expected);

    return {
      code: item.symbol || item.id || "-",
      name: item.name || "未命名标的",
      expected,
      side,
      condition: detail.stop_loss
        ? `条件：${detail.stop_loss}`
        : `条件：${item.reason_summary || "满足策略触发条件"}`
    };
  })
);

const arbitragePlans = computed(() =>
  rawArbitragePlans.value.slice(0, 4).map((item) => ({
    name: `${item.contract_a || "-"} - ${item.contract_b || "-"}`,
    tag: `${mapArbitrageType(item.type)} · ${mapArbitrageStatus(item.status)}`,
    entry: pointWithRule(item.entry_point, "触发"),
    close: pointWithRule(item.exit_point, "止盈"),
    stop: pointWithRule(item.stop_point, "止损"),
    position: suggestPosition(item.percentile, item.risk_level)
  }))
);

const methods = computed(() => {
  const guidance = futuresGuidance.value;
  if (!guidance) {
    return [
      { title: "步骤 1：确认触发条件", desc: "仅在价差进入触发区间且成交量同步放大时建仓。" },
      { title: "步骤 2：分两笔进场", desc: "首笔 60%，二次确认后补足剩余 40%，避免单点误差。" },
      { title: "步骤 3：预设平仓与止损", desc: "下单同时写入平仓/止损点位，不临盘手动决策。" },
      { title: "步骤 4：复盘偏差原因", desc: "记录滑点、执行延迟和信号有效性，用于次日优化。" }
    ];
  }

  return [
    {
      title: "步骤 1：确认进入区间",
      desc: `当价差进入 ${guidance.entry_range || "-"}，且方向为 ${mapGuidanceDirection(
        guidance.guidance_direction
      )} 时执行首笔建仓。`
    },
    {
      title: "步骤 2：设置平仓目标",
      desc: `平仓参考区间 ${guidance.take_profit_range || "-"}，达到目标后分批止盈。`
    },
    {
      title: "步骤 3：执行止损纪律",
      desc: `当价差进入 ${guidance.stop_loss_range || "-"} 时立即止损，不做主观加仓。`
    },
    {
      title: "步骤 4：校验失效条件",
      desc: `若出现“${guidance.invalid_condition || "模型失效"}”，立即暂停策略并复盘。`
    }
  ];
});

const heroTags = computed(() => [
  `历史胜率 ${calcHitRate(historyStocks.value)}`,
  `今日推荐 ${todayStocks.value.length} 只`,
  `套利计划 ${arbitragePlans.value.length} 套`
]);

const panelStats = computed(() => [
  { label: "历史推荐胜率", value: calcHitRate(historyStocks.value) },
  { label: "今日推荐股票", value: `${todayStocks.value.length} 只` },
  { label: "套利计划", value: `${arbitragePlans.value.length} 套` }
]);

const mobileQuickActions = computed(() => [
  { title: "今日推荐", desc: `${todayStocks.value.length} 只待确认` },
  { title: "套利进场", desc: `${arbitragePlans.value.length} 套可执行` },
  { title: "数据状态", desc: loadError.value ? "部分失败" : "已同步" }
]);

const shouldShowVipPromo = computed(() => !vipStateLoading.value && !isVIPUser.value);

const vipPromoPriceText = computed(() => {
  const offer = vipOffer.value;
  const price = Number(offer?.price);
  if (!Number.isFinite(price) || price <= 0) {
    return "VIP权益套餐";
  }
  const duration = Number(offer?.duration_days);
  if (Number.isFinite(duration) && duration >= 360) {
    return `¥${price.toFixed(0)} / 年`;
  }
  if (Number.isFinite(duration) && duration >= 85) {
    return `¥${price.toFixed(0)} / 季`;
  }
  if (Number.isFinite(duration) && duration >= 28) {
    return `¥${price.toFixed(0)} / 月`;
  }
  return `¥${price.toFixed(0)} 起`;
});

const vipPromoDesc = computed(() => {
  const offer = vipOffer.value;
  if (!offer?.name) {
    return "会员到期自动失效，续费后立即恢复权益。";
  }
  const duration = Number(offer.duration_days);
  if (Number.isFinite(duration) && duration > 0) {
    return `${offer.name}，权益有效期 ${duration} 天。`;
  }
  return `${offer.name}，权益即时生效。`;
});

const vipPromoCTA = computed(() => (isLoggedIn.value ? "立即开通VIP" : "登录后开通VIP"));

const tasks = computed(() => {
  const firstStock = todayStocks.value[0];
  const firstPlan = arbitragePlans.value[0];
  return [
    {
      title: "09:35 更新分时信号",
      note: `优先确认 ${firstStock?.code || "首个推荐标的"} 是否满足触发条件`,
      level: "high"
    },
    {
      title: "10:20 检查套利价差",
      note: `关注 ${firstPlan?.name || "套利组合"} 的进入点位`,
      level: "mid"
    },
    {
      title: "14:50 执行收盘复核",
      note: "平仓后写入交易复盘与偏差说明",
      level: "low"
    }
  ];
});

async function loadHomeData() {
  loading.value = true;
  loadError.value = "";

  const errors = [];
  try {
    const [stockResult, arbitrageResult] = await Promise.allSettled([
      listStockRecommendations({ page: 1, page_size: 8 }),
      listFuturesArbitrage({ page: 1, page_size: 8 })
    ]);

    if (stockResult.status === "fulfilled" && Array.isArray(stockResult.value?.items)) {
      if (stockResult.value.items.length > 0) {
        rawStockRecommendations.value = stockResult.value.items;
      }
      await hydrateStockDetails(rawStockRecommendations.value.slice(0, 6));
    } else if (stockResult.status === "rejected") {
      errors.push(`股票推荐接口失败：${stockResult.reason?.message || "unknown error"}`);
    }

    if (arbitrageResult.status === "fulfilled" && Array.isArray(arbitrageResult.value?.items)) {
      if (arbitrageResult.value.items.length > 0) {
        rawArbitragePlans.value = arbitrageResult.value.items;
      }
      await hydrateFuturesGuidance(rawArbitragePlans.value[0]?.contract_a);
    } else if (arbitrageResult.status === "rejected") {
      errors.push(`期货套利接口失败：${arbitrageResult.reason?.message || "unknown error"}`);
    }

    if (errors.length > 0) {
      loadError.value = errors.join("；");
    }
    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  } catch (error) {
    loadError.value = error?.message || "接口请求失败";
  } finally {
    loading.value = false;
  }
}

async function loadHomeInsights(options = {}) {
  newsLoading.value = true;
  newsError.value = "";
  const keyword = normalizeInsightKeyword(options.keyword ?? insightKeyword.value);
  const query = {
    page: 1,
    page_size: 8
  };
  if (keyword) {
    query.keyword = keyword;
  }
  const [categoryResult, newsResult] = await Promise.allSettled([
    listNewsCategories(),
    listNewsArticles(query)
  ]);
  const errors = [];
  const categoryMap = {};
  if (categoryResult.status === "fulfilled") {
    const items = Array.isArray(categoryResult.value?.items) ? categoryResult.value.items : [];
    newsCategoryTags.value = items
      .slice()
      .sort((a, b) => Number(a?.sort || 0) - Number(b?.sort || 0))
      .slice(0, 6)
      .map((item) => item.name)
      .filter(Boolean);
    items.forEach((item) => {
      if (item?.id) {
        categoryMap[item.id] = item.name || "资讯";
      }
    });
  } else {
    errors.push(`分类接口失败：${categoryResult.reason?.message || "unknown error"}`);
  }

  if (newsResult.status === "fulfilled") {
    const items = Array.isArray(newsResult.value?.items) ? newsResult.value.items : [];
    const mapped = items.map((item) => mapNewsHighlight(item, categoryMap)).filter(Boolean);
    if (mapped.length > 0) {
      newsHighlights.value = mapped;
    } else if (!useDemoFallback) {
      newsHighlights.value = [];
    }
    appliedInsightKeyword.value = keyword;
  } else {
    errors.push(`资讯接口失败：${newsResult.reason?.message || "unknown error"}`);
  }

  newsError.value = errors.join("；");
  newsUpdatedAt.value = formatDateTime(new Date().toISOString());
  newsLoading.value = false;
}

async function loadVIPState() {
  vipStateLoading.value = true;
  try {
    if (!isLoggedIn.value) {
      isVIPUser.value = false;
      if (!useDemoFallback) {
        vipOffer.value = null;
      }
      return;
    }
    const [quotaResult, productsResult] = await Promise.allSettled([
      getMembershipQuota(),
      listMembershipProducts({ status: "ACTIVE", page: 1, page_size: 20 })
    ]);

    if (quotaResult.status === "fulfilled" && quotaResult.value) {
      isVIPUser.value = resolveVIPState(quotaResult.value);
    } else {
      isVIPUser.value = false;
    }

    if (productsResult.status === "fulfilled" && Array.isArray(productsResult.value?.items)) {
      const sorted = productsResult.value.items
        .filter((item) => String(item?.status || "").toUpperCase() === "ACTIVE")
        .sort((a, b) => Number(a?.price || 0) - Number(b?.price || 0));
      if (sorted.length > 0) {
        vipOffer.value = sorted[0];
      } else if (!useDemoFallback) {
        vipOffer.value = null;
      }
    } else if (!useDemoFallback) {
      vipOffer.value = null;
    }
  } finally {
    vipStateLoading.value = false;
  }
}

async function hydrateStockDetails(items) {
  if (!Array.isArray(items) || items.length === 0) {
    return;
  }

  const nextDetailMap = { ...stockDetailMap.value };
  const nextPerfMap = { ...stockPerformanceMap.value };

  await Promise.all(
    items.map(async (item) => {
      if (!item?.id) {
        return;
      }
      const [detailRes, perfRes] = await Promise.allSettled([
        getStockRecommendationDetail(item.id),
        getStockRecommendationPerformance(item.id)
      ]);

      if (detailRes.status === "fulfilled" && detailRes.value) {
        nextDetailMap[item.id] = detailRes.value;
      }
      if (perfRes.status === "fulfilled" && Array.isArray(perfRes.value?.points)) {
        nextPerfMap[item.id] = perfRes.value.points;
      }
    })
  );

  stockDetailMap.value = nextDetailMap;
  stockPerformanceMap.value = nextPerfMap;
}

async function hydrateFuturesGuidance(contract) {
  if (!contract) {
    return;
  }
  try {
    const guidance = await getFuturesGuidance(contract);
    if (guidance) {
      futuresGuidance.value = guidance;
    }
  } catch {
    // guidance 接口失败时保留当前数据
  }
}

function scrollToArbitrage() {
  const target = document.getElementById("arbitrage-section");
  if (target) {
    target.scrollIntoView({ behavior: "smooth", block: "start" });
  }
}

function scrollToInsights() {
  const target = document.getElementById("insight-section");
  if (target) {
    target.scrollIntoView({ behavior: "smooth", block: "start" });
  }
}

function openNewsModule() {
  const keyword = normalizeInsightKeyword(appliedInsightKeyword.value);
  if (keyword) {
    router.push({ path: "/news", query: { keyword } });
    return;
  }
  router.push("/news");
}

function goMembershipCenter() {
  if (isLoggedIn.value) {
    router.push("/membership");
    return;
  }
  router.push({ path: "/auth", query: { redirect: "/membership" } });
}

function resolveVIPState(quota) {
  const status = String(quota?.vip_status || "").toUpperCase();
  if (status === "ACTIVE") {
    return true;
  }
  const level = String(quota?.member_level || "").toUpperCase();
  if (!level.startsWith("VIP")) {
    return false;
  }
  const remainingDays = Number(quota?.vip_remaining_days);
  if (Number.isFinite(remainingDays)) {
    return remainingDays > 0;
  }
  return true;
}

function mapNewsHighlight(item, categoryMap = {}) {
  if (!item?.id) {
    return null;
  }
  const visibility = String(item.visibility || "").toUpperCase();
  return {
    id: item.id,
    category: categoryMap[item.category_id] || resolveNewsCategory(item),
    time: formatDateTime(item.published_at || item.created_at || item.updated_at),
    title: item.title || "未命名资讯",
    summary: summarizeText(item.summary || item.content),
    source: item.source || "资讯中心",
    visibility: visibility === "VIP" ? "VIP" : "PUBLIC"
  };
}

function resolveNewsCategory(item) {
  const text = `${item?.category_name || ""} ${item?.title || ""}`;
  if (/(研报|研究|券商)/i.test(text)) {
    return "研报";
  }
  if (/(期刊|周刊|月刊)/i.test(text)) {
    return "期刊";
  }
  return "新闻";
}

function summarizeText(value) {
  const text = String(value || "")
    .replace(/<[^>]*>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
  if (!text) {
    return "暂无摘要";
  }
  if (text.length <= 56) {
    return text;
  }
  return `${text.slice(0, 56)}...`;
}

async function handleInsightSearch() {
  await loadHomeInsights({ keyword: insightKeyword.value });
  scrollToInsights();
}

async function clearInsightSearch() {
  insightKeyword.value = "";
  await loadHomeInsights({ keyword: "" });
  scrollToInsights();
}

function normalizeInsightKeyword(value) {
  return String(value || "")
    .replace(/\s+/g, " ")
    .trim();
}

function calcCumulativeReturn(points) {
  if (!Array.isArray(points) || points.length === 0) {
    return null;
  }
  let acc = 1;
  let hasValid = false;
  points.forEach((point) => {
    const value = Number(point?.return);
    if (Number.isFinite(value)) {
      acc *= 1 + value;
      hasValid = true;
    }
  });
  if (!hasValid) {
    return null;
  }
  return acc - 1;
}

function estimateBenchmarkReturn(value) {
  if (!Number.isFinite(value)) {
    return null;
  }
  const bench = value * 0.58;
  return Math.max(Math.min(bench, 0.25), -0.25);
}

function inferExpectedRange(takeProfitText, historyReturn) {
  const rangeFromTakeProfit = parsePercentRange(takeProfitText);
  if (rangeFromTakeProfit) {
    return rangeFromTakeProfit;
  }

  if (Number.isFinite(historyReturn)) {
    const low = historyReturn * 0.6;
    const high = historyReturn * 1.2;
    return `${formatPercent(low)} ~ ${formatPercent(high)}`;
  }

  return "区间待更新";
}

function inferExpectedSide(expected) {
  if (expected.startsWith("+")) {
    return "up";
  }
  if (expected.startsWith("-")) {
    return "down";
  }
  return "flat";
}

function parsePercentRange(text) {
  if (!text || typeof text !== "string") {
    return "";
  }
  const matched = text.match(/-?\d+(\.\d+)?%/g);
  if (!matched || matched.length === 0) {
    return "";
  }
  if (matched.length >= 2) {
    const first = normalizePercentText(matched[0]);
    const second = normalizePercentText(matched[1]);
    return `${first} ~ ${second}`;
  }
  return normalizePercentText(matched[0]);
}

function normalizePercentText(value) {
  if (!value) {
    return "";
  }
  if (value.startsWith("+") || value.startsWith("-")) {
    return value;
  }
  return `+${value}`;
}

function formatPercent(value) {
  if (!Number.isFinite(value)) {
    return "-";
  }
  const percent = (value * 100).toFixed(2);
  if (Number(percent) > 0) {
    return `+${percent}%`;
  }
  return `${percent}%`;
}

function formatScore(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  return num.toFixed(1);
}

function formatDate(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return "-";
  }
  return new Date(ts).toLocaleDateString("zh-CN");
}

function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return "-";
  }
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

function mapRiskLevel(level) {
  const source = String(level || "").toUpperCase();
  if (source === "LOW") {
    return "低风险";
  }
  if (source === "MEDIUM") {
    return "中风险";
  }
  if (source === "HIGH") {
    return "高风险";
  }
  return source || "-";
}

function mapArbitrageType(type) {
  const source = String(type || "").toUpperCase();
  if (source === "CALENDAR") {
    return "跨期价差";
  }
  if (source === "CROSS") {
    return "跨品种价差";
  }
  if (source === "INTERTEMPORAL") {
    return "跨期套利";
  }
  return source || "套利";
}

function mapArbitrageStatus(status) {
  const source = String(status || "").toUpperCase();
  if (source === "WATCH") {
    return "观察中";
  }
  if (source === "ACTIVE") {
    return "可执行";
  }
  if (source === "EXPIRED") {
    return "已失效";
  }
  return source || "待确认";
}

function mapGuidanceDirection(direction) {
  const source = String(direction || "").toUpperCase();
  if (source === "LONG_SPREAD") {
    return "做多价差";
  }
  if (source === "SHORT_SPREAD") {
    return "做空价差";
  }
  return source || "-";
}

function pointWithRule(value, label) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return `${label}区间待更新`;
  }
  return `${label} ${num.toFixed(2)} 点`;
}

function suggestPosition(percentile, riskLevel) {
  const p = Number(percentile);
  const risk = String(riskLevel || "").toUpperCase();
  if (risk === "HIGH") {
    return "总资金 15%-20%";
  }
  if (Number.isFinite(p) && p >= 0.9) {
    return "总资金 20%-25%";
  }
  if (Number.isFinite(p) && p >= 0.8) {
    return "总资金 25%-30%";
  }
  return "总资金 30%-35%";
}

function calcHitRate(rows) {
  if (!Array.isArray(rows) || rows.length === 0) {
    return "-";
  }
  const positive = rows.filter((item) => Number.isFinite(item.rawChange) && item.rawChange > 0).length;
  const rate = (positive / rows.length) * 100;
  return `${Math.round(rate)}%`;
}

function getTrendClass(value) {
  if (typeof value !== "string") {
    return "";
  }
  if (value.startsWith("+")) {
    return "up";
  }
  if (value.startsWith("-")) {
    return "down";
  }
  return "";
}

watch(
  () => isLoggedIn.value,
  () => {
    loadVIPState();
  },
  { immediate: true }
);

onMounted(() => {
  loadHomeData();
  loadHomeInsights();
});
</script>

<style scoped>
.home-page {
  display: grid;
  gap: 14px;
}

.hero-stage {
  border-radius: 22px;
  padding: clamp(16px, 2.2vw, 24px);
  display: grid;
  gap: 12px;
  grid-template-columns: 1.05fr 0.95fr;
  background:
    radial-gradient(circle at 0% 100%, rgba(234, 215, 180, 0.2) 0%, transparent 36%),
    radial-gradient(circle at 100% 0%, rgba(112, 164, 148, 0.18) 0%, transparent 34%),
    rgba(255, 255, 255, 0.93);
}

.hero-kicker {
  margin: 0;
  font-size: 12px;
  letter-spacing: 0.3px;
  color: var(--color-pine-600);
}

.hero-copy h1 {
  margin: 8px 0 10px;
  font-family: var(--font-serif);
  font-size: clamp(30px, 3.5vw, 46px);
  line-height: 1.15;
}

.hero-desc {
  margin: 0;
  color: var(--color-text-sub);
  max-width: 580px;
  line-height: 1.72;
}

.hero-actions {
  margin-top: 18px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.primary-btn,
.ghost-btn {
  border: 0;
  border-radius: 12px;
  padding: 10px 16px;
  cursor: pointer;
  font-weight: 600;
}

.primary-btn:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.primary-btn {
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.ghost-btn {
  color: var(--color-pine-700);
  background: rgba(239, 232, 218, 0.72);
}

.hero-tags {
  margin-top: 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.hero-tags span {
  border-radius: 999px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(252, 251, 247, 0.82);
  padding: 4px 10px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.api-state {
  margin: 10px 0 0;
  font-size: 12px;
  color: var(--color-pine-700);
}

.api-state.warning {
  color: #7f5f36;
}

.hero-panel {
  border-radius: 18px;
  border: 1px solid rgba(216, 223, 216, 0.86);
  background: linear-gradient(180deg, rgba(33, 78, 69, 0.95), rgba(39, 92, 82, 0.92));
  color: #f6fbf8;
  padding: 14px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.panel-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.panel-head p {
  margin: 0;
  font-size: 13px;
}

.panel-head strong {
  font-size: 12px;
  color: rgba(246, 251, 248, 0.78);
}

.panel-stats {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.panel-stats article {
  border-radius: 11px;
  border: 1px solid rgba(245, 251, 248, 0.2);
  background: rgba(255, 255, 255, 0.08);
  padding: 8px;
}

.panel-stats p {
  margin: 0;
  font-size: 11px;
  color: rgba(246, 251, 248, 0.72);
}

.panel-stats strong {
  display: block;
  margin-top: 4px;
  font-size: 16px;
}

.task-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 8px;
}

.task-list li {
  border-radius: 10px;
  border: 1px solid rgba(245, 251, 248, 0.2);
  background: rgba(255, 255, 255, 0.05);
  padding: 8px 10px;
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 8px;
}

.task-dot {
  margin-top: 4px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.task-dot.high {
  background: #e8c07d;
}

.task-dot.mid {
  background: #90d0bd;
}

.task-dot.low {
  background: #bcd7cf;
}

.task-title {
  margin: 0;
  font-size: 13px;
}

.task-note {
  margin: 3px 0 0;
  font-size: 12px;
  color: rgba(246, 251, 248, 0.75);
}

.home-search-banner {
  border-radius: 20px;
  padding: 14px 16px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background:
    radial-gradient(circle at 0% 0%, rgba(63, 127, 113, 0.2) 0%, transparent 36%),
    radial-gradient(circle at 100% 100%, rgba(234, 215, 180, 0.34) 0%, transparent 38%),
    rgba(255, 255, 255, 0.95);
}

.search-banner-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.search-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.search-banner-head h2 {
  margin: 6px 0 0;
  font-family: var(--font-serif);
  font-size: clamp(21px, 2.2vw, 30px);
  line-height: 1.25;
}

.search-jump-btn {
  border: 0;
  border-radius: 10px;
  padding: 9px 12px;
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
}

.home-search-banner .insight-search {
  margin-top: 12px;
}

.home-search-banner .insight-search input {
  font-size: 14px;
  padding: 11px 12px;
}

.search-tip {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.mobile-quick {
  display: none;
  padding: 10px;
  border-radius: 16px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.mobile-quick button {
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  border-radius: 12px;
  padding: 9px;
  text-align: left;
  cursor: pointer;
}

.mobile-quick p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.mobile-quick span {
  margin-top: 3px;
  display: block;
  font-size: 11px;
  color: var(--color-text-sub);
}

.insight-section {
  display: grid;
  gap: 12px;
  grid-template-columns: 1.15fr 0.85fr;
}

.insight-section.single-column {
  grid-template-columns: 1fr;
}

.insight-card,
.vip-promo-card {
  padding: 14px;
}

.insight-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.insight-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.insight-actions button {
  border: 0;
  border-radius: 10px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.insight-actions button.ghost {
  color: var(--color-pine-700);
  background: rgba(239, 232, 218, 0.72);
}

.insight-actions button:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.insight-search {
  margin-top: 10px;
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 8px;
}

.insight-search input {
  border: 1px solid rgba(216, 223, 216, 0.95);
  border-radius: 10px;
  padding: 9px 11px;
  background: rgba(255, 255, 255, 0.94);
  color: var(--color-text-main);
}

.insight-search input:focus {
  outline: none;
  border-color: var(--color-pine-500);
  box-shadow: 0 0 0 3px rgba(63, 127, 113, 0.14);
}

.insight-search button {
  border: 0;
  border-radius: 10px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.insight-search button.ghost {
  color: var(--color-pine-700);
  background: rgba(239, 232, 218, 0.72);
}

.insight-search button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.insight-status {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--color-pine-700);
}

.insight-status.warning {
  color: #7f5f36;
}

.insight-tags {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.insight-tags span {
  border-radius: 999px;
  border: 1px solid rgba(216, 223, 216, 0.92);
  background: rgba(252, 251, 247, 0.9);
  padding: 4px 9px;
  font-size: 11px;
  color: var(--color-text-sub);
}

.insight-list {
  margin-top: 10px;
  display: grid;
  gap: 8px;
}

.insight-item {
  border: 1px solid rgba(216, 223, 216, 0.9);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
  cursor: pointer;
}

.insight-item-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.insight-item .category {
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
  color: var(--color-pine-700);
  background: rgba(223, 236, 230, 0.82);
}

.insight-item .time {
  font-size: 11px;
  color: var(--color-text-sub);
}

.insight-item h3 {
  margin: 7px 0 5px;
  font-size: 16px;
  line-height: 1.45;
}

.insight-item p {
  margin: 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.58;
}

.insight-meta {
  margin-top: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: var(--color-text-sub);
}

.vip-mark {
  border-radius: 999px;
  padding: 2px 8px;
  background: rgba(239, 232, 218, 0.82);
  color: #7f5f36;
}

.vip-mark.active {
  background: rgba(192, 139, 60, 0.2);
  color: #8a5b16;
}

.vip-promo-card {
  border: 1px solid rgba(216, 223, 216, 0.86);
  background:
    radial-gradient(circle at 100% 0%, rgba(192, 139, 60, 0.26) 0%, transparent 36%),
    radial-gradient(circle at 0% 100%, rgba(63, 127, 113, 0.16) 0%, transparent 38%),
    rgba(255, 255, 255, 0.95);
}

.promo-kicker {
  margin: 0;
  font-size: 12px;
  color: #8a5b16;
}

.vip-promo-card h2 {
  margin: 8px 0 0;
  font-family: var(--font-serif);
  font-size: 25px;
  line-height: 1.3;
}

.promo-price {
  margin: 10px 0 0;
  font-size: 28px;
  color: var(--color-pine-700);
  font-weight: 700;
}

.promo-note {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
}

.promo-benefits {
  margin: 12px 0 0;
  padding-left: 18px;
  display: grid;
  gap: 7px;
  color: var(--color-text-sub);
  font-size: 13px;
}

.promo-btn {
  margin-top: 14px;
  width: 100%;
  border: 0;
  border-radius: 12px;
  padding: 11px 12px;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  background: linear-gradient(145deg, #c08b3c, #a8742e);
}

.stock-section {
  display: grid;
  gap: 12px;
  grid-template-columns: 1.3fr 0.7fr;
}

.history-card,
.today-card,
.arbitrage-card,
.method-card {
  padding: 14px;
}

.table-wrap {
  margin-top: 10px;
  overflow-x: auto;
}

.stock-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 680px;
}

.stock-table th,
.stock-table td {
  border-bottom: 1px solid rgba(216, 223, 216, 0.75);
  padding: 9px 8px;
  text-align: left;
  font-size: 13px;
  white-space: nowrap;
}

.stock-table th {
  font-size: 12px;
  color: var(--color-text-sub);
  background: rgba(246, 244, 239, 0.7);
}

.mobile-history-list {
  display: none;
}

.mobile-history-item {
  border-radius: 12px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 9px;
}

.mobile-history-item .top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.mobile-history-item .top p {
  margin: 0;
  font-weight: 700;
  font-size: 13px;
}

.mobile-history-item .top span {
  font-size: 12px;
  color: var(--color-text-sub);
}

.mobile-history-item .line {
  margin-top: 5px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
}

.up {
  color: #267f62;
  font-weight: 600;
}

.down {
  color: #b23a2a;
  font-weight: 600;
}

.empty-box {
  margin-top: 10px;
  border-radius: 11px;
  border: 1px dashed rgba(216, 223, 216, 0.95);
  background: rgba(246, 244, 239, 0.7);
  padding: 10px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.today-list {
  margin-top: 10px;
  display: grid;
  gap: 8px;
}

.today-item {
  border-radius: 12px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
}

.today-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.code {
  margin: 0;
  font-weight: 700;
}

.name {
  margin: 0;
  font-size: 13px;
  color: var(--color-text-sub);
}

.expect {
  margin: 6px 0 0;
  font-size: 18px;
  font-weight: 700;
}

.expect.up {
  color: #267f62;
}

.expect.flat {
  color: #7f5f36;
}

.expect.down {
  color: #b23a2a;
}

.condition {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
  line-height: 1.5;
}

.futures-section {
  display: grid;
  gap: 12px;
  grid-template-columns: 1.1fr 0.9fr;
}

.arbitrage-list {
  margin-top: 10px;
  display: grid;
  gap: 10px;
}

.plan-item {
  border-radius: 12px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.88);
  padding: 10px;
}

.plan-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.plan-title p {
  margin: 0;
  font-weight: 600;
}

.plan-title span {
  font-size: 12px;
  color: #7f5f36;
  background: rgba(227, 207, 173, 0.55);
  border-radius: 999px;
  padding: 2px 8px;
}

.plan-grid {
  margin-top: 9px;
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.plan-grid div {
  border-radius: 10px;
  border: 1px solid rgba(216, 223, 216, 0.85);
  background: rgba(252, 251, 247, 0.84);
  padding: 8px;
}

.plan-grid p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.plan-grid strong {
  display: block;
  margin-top: 4px;
  font-size: 16px;
  color: var(--color-pine-700);
}

.method-list {
  margin: 10px 0 0;
  padding-left: 18px;
  display: grid;
  gap: 9px;
}

.step-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.step-desc {
  margin: 3px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.56;
}

.risk-tip {
  margin: 12px 0 0;
  border-radius: 10px;
  border: 1px dashed rgba(216, 223, 216, 0.95);
  background: rgba(246, 244, 239, 0.8);
  padding: 9px;
  font-size: 12px;
  color: var(--color-text-sub);
  line-height: 1.6;
}

@media (max-width: 1080px) {
  .hero-stage,
  .insight-section,
  .stock-section,
  .futures-section {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .hero-copy h1 {
    font-size: 32px;
  }

  .search-banner-head {
    flex-direction: column;
  }

  .search-jump-btn {
    width: 100%;
  }

  .mobile-quick {
    display: grid;
  }

  .insight-head {
    flex-direction: column;
  }

  .insight-actions {
    width: 100%;
  }

  .insight-actions button {
    flex: 1;
  }

  .insight-search {
    grid-template-columns: 1fr 1fr;
  }

  .insight-search input {
    grid-column: 1 / -1;
  }

  .vip-promo-card h2 {
    font-size: 22px;
  }

  .table-wrap {
    display: none;
  }

  .mobile-history-list {
    margin-top: 10px;
    display: grid;
    gap: 8px;
  }

  .panel-stats {
    grid-template-columns: 1fr;
  }

  .plan-grid {
    grid-template-columns: 1fr;
  }
}
</style>
