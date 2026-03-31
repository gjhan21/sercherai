<template>
  <section class="membership-page fade-up">
    <header class="member-hero card">
      <div class="member-hero-copy finance-copy-stack">
        <div class="finance-pill-row">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">会员页</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">升级价值解释</span>
        </div>
        <div>
          <p class="hero-kicker">会员中心</p>
          <h1 class="section-title">查看会员权益、状态和订单</h1>
          <p class="section-subtitle">{{ heroSubtitle }}</p>
        </div>
      </div>
      <div class="hero-pill finance-summary-pill">
        <p>当前套餐</p>
        <strong>{{ currentPlanName }}</strong>
        <small>{{ currentLevelLabel }} · {{ activationStateLabel }}</small>
      </div>
      <div class="member-hero-stats finance-hero-stat-grid">
        <article class="finance-hero-stat-card">
          <span>当前套餐</span>
          <strong>{{ currentPlanName }}</strong>
          <p>套餐和权益仍按真实账户状态展示。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>当前状态</span>
          <strong>{{ currentLevelLabel }} · {{ activationStateLabel }}</strong>
          <p>确认是否待支付、待续费或已正常开通。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>深读价值</span>
          <strong>{{ latestStrategySnapshot?.name ? `继续深读 ${latestStrategySnapshot.name}` : "全文 / 附件 / 连续跟踪" }}</strong>
          <p>会员页优先解释为什么现在需要升级，而不是先推价格。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>下一动作</span>
          <strong>{{ membershipJourney.primaryAction.label }}</strong>
          <p>{{ membershipJourney.summaryNote }}</p>
        </article>
      </div>
    </header>

    <div class="api-tip">
      <p v-if="loading">正在同步会员数据...</p>
      <p v-else-if="loadError">{{ loadError }}</p>
      <p v-else>数据更新时间：{{ lastUpdatedAt || "-" }}</p>
    </div>

    <StatePanel
      :tone="membershipStatus.tone"
      :eyebrow="membershipStatus.label"
      :title="membershipStatus.title"
      :description="membershipStatus.desc"
    >
      <template #actions>
        <button type="button" class="finance-primary-btn" @click="handleMembershipJourneyAction('journey_primary', membershipJourney.primaryAction)">
          {{ membershipJourney.primaryAction.label }}
        </button>
        <button class="ghost finance-ghost-btn" type="button" @click="handleMembershipJourneyAction('journey_secondary', membershipJourney.secondaryAction)">
          {{ membershipJourney.secondaryAction.label }}
        </button>
      </template>
    </StatePanel>

    <StatePanel
      compact
      :tone="renewalExperimentPanel.tone"
      :eyebrow="renewalExperimentPanel.eyebrow"
      :title="renewalExperimentPanel.title"
      :description="renewalExperimentPanel.desc"
    >
      <template #actions>
        <button type="button" class="finance-primary-btn" @click="handleRenewalAction('renewal_primary', renewalExperimentPanel.primaryAction)">
          {{ renewalExperimentPanel.primaryAction.label }}
        </button>
        <button class="ghost finance-ghost-btn" type="button" @click="handleRenewalAction('renewal_secondary', renewalExperimentPanel.secondaryAction)">
          {{ renewalExperimentPanel.secondaryAction.label }}
        </button>
      </template>
    </StatePanel>

    <section class="membership-focus-layout finance-dual-rail">
      <article class="card membership-focus-card finance-section-card">
        <header class="membership-focus-head finance-section-head-grid">
          <div>
            <p class="hero-kicker">当前状态</p>
            <h2 class="section-title">先确认权益状态，再决定升级或续费。</h2>
            <p class="section-subtitle">
              查看当前状态、升级价值和后续操作。
            </p>
          </div>
          <div class="membership-focus-actions finance-action-row">
            <button type="button" class="finance-primary-btn" :disabled="loading" @click="loadMembershipData">
              {{ loading ? "同步中..." : "刷新会员状态" }}
            </button>
            <button type="button" class="ghost finance-ghost-btn" @click="handleMembershipJourneyAction('focus_primary', membershipJourney.primaryAction)">
              {{ membershipJourney.primaryAction.label }}
            </button>
          </div>
        </header>

        <div class="membership-status-grid finance-card-grid finance-card-grid-3">
          <article v-for="item in membershipFocusRows" :key="item.label" class="finance-card-surface">
            <p>{{ item.label }}</p>
            <strong>{{ item.value }}</strong>
            <span>{{ item.note }}</span>
          </article>
        </div>

        <div class="membership-explain-grid finance-card-grid finance-card-grid-3">
          <article v-for="item in membershipValueGuideRows" :key="item.title" class="finance-card-surface">
            <p>{{ item.title }}</p>
            <strong>{{ item.summary }}</strong>
            <span>{{ item.desc }}</span>
          </article>
        </div>
      </article>

      <aside class="membership-focus-side finance-stack-tight finance-sticky-side">
        <article class="card membership-side-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">升级重点</h2>
              <p class="section-subtitle">先看权益差异，再决定套餐和支付。</p>
            </div>
          </header>
          <div class="membership-side-list finance-card-stack">
            <article v-for="item in membershipPageGuideRows" :key="item.title" class="finance-card-surface">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>

        <article class="card membership-side-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">当前状态卡</h2>
              <p class="section-subtitle">待支付、已开通等状态会优先显示。</p>
            </div>
          </header>
          <div class="membership-side-list finance-card-stack">
            <article v-for="item in membershipCurrentTodoRows" :key="item.title" class="finance-card-surface">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>
      </aside>
    </section>

    <article class="cadence-hub card">
      <header class="cadence-head">
        <div>
          <p class="hero-kicker">今日使用节奏</p>
          <span class="experiment-badge finance-pill finance-pill-roomy finance-pill-accent">{{ membershipExperimentLabel }}</span>
          <h2 class="section-title">{{ membershipJourney.title }}</h2>
          <p class="section-subtitle">{{ membershipJourney.desc }}</p>
        </div>
        <div class="cadence-summary finance-summary-pill">
          <p>{{ membershipJourney.summaryLabel }}</p>
          <strong>{{ membershipJourney.summaryValue }}</strong>
          <small>{{ membershipJourney.summaryNote }}</small>
        </div>
      </header>

      <div class="cadence-grid">
        <article v-for="entry in cadenceEntries" :key="entry.slot" class="cadence-card finance-card-surface">
          <p class="cadence-slot">{{ entry.slot }}</p>
          <h3>{{ entry.title }}</h3>
          <p class="cadence-desc">{{ entry.desc }}</p>
          <div class="cadence-meta">
            <span class="finance-pill finance-pill-compact finance-pill-info">{{ entry.highlight }}</span>
            <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ entry.supporting }}</span>
          </div>
          <div class="cadence-actions">
            <button type="button" class="finance-primary-btn" @click="handleCadenceAction(entry, 'primary')">
              {{ entry.primary.label }}
            </button>
            <button class="ghost finance-ghost-btn" type="button" @click="handleCadenceAction(entry, 'secondary')">
              {{ entry.secondary.label }}
            </button>
          </div>
        </article>
      </div>
    </article>

    <article class="insight-upgrade card">
      <header class="section-head">
        <div>
          <p class="hero-kicker">解释能力升级</p>
          <h2 class="section-title">升级后可查看更完整的策略解释和跟踪信息。</h2>
          <p class="section-subtitle">
            不同身份可查看的内容深度不同。
          </p>
        </div>
      </header>

      <div class="insight-stage-grid">
        <article
          v-for="item in explanationCapabilityCards"
          :key="item.key"
          class="insight-stage-card finance-card-surface"
          :class="{ active: item.active, unlocked: item.unlocked }"
        >
          <div class="insight-stage-head">
            <p>{{ item.stage }}</p>
            <span
              class="finance-pill finance-pill-compact"
              :class="item.active ? 'finance-pill-accent' : item.unlocked ? 'finance-pill-info' : 'finance-pill-neutral'"
            >
              {{ item.badge }}
            </span>
          </div>
          <strong>{{ item.title }}</strong>
          <p>{{ item.desc }}</p>
          <div class="insight-chip-list">
            <span
              v-for="point in item.points"
              :key="`${item.key}-${point}`"
              class="finance-pill finance-pill-compact finance-pill-info"
            >
              {{ point }}
            </span>
          </div>
        </article>
      </div>

      <div class="insight-value-grid">
        <article v-for="item in explanationValueRows" :key="item.label" class="finance-card-surface">
          <p>{{ item.label }}</p>
          <strong>{{ item.value }}</strong>
          <span>{{ item.note }}</span>
        </article>
      </div>

      <div class="insight-bridge-box">
        <p>我的关注配合</p>
        <strong>{{ watchlistBridgeSummary.title }}</strong>
        <span>{{ watchlistBridgeSummary.desc }}</span>
        <div class="insight-chip-list">
          <span
            v-for="point in watchlistBridgeSummary.points"
            :key="`bridge-${point}`"
            class="finance-pill finance-pill-compact finance-pill-info"
          >
            {{ point }}
          </span>
        </div>
      </div>

      <div class="runtime-proof-box finance-card-surface">
        <div class="runtime-proof-head">
          <div>
            <p>最近策略解释快照</p>
            <strong>{{ latestStrategySnapshot?.name || "等待同步最近策略" }}</strong>
          </div>
          <span>{{ latestStrategySnapshot?.meta || (strategySnapshotLoading ? "同步中..." : "-") }}</span>
        </div>
        <p v-if="strategySnapshotError" class="runtime-proof-note finance-note-strip finance-note-strip-warning">
          {{ strategySnapshotError }}
        </p>
        <p v-else class="runtime-proof-note finance-note-strip finance-note-strip-info">
          {{ latestStrategySnapshot?.summary || "同步最近一次策略 explanation 后，这里会展示更真实的升级价值。" }}
        </p>
        <div v-if="latestStrategySnapshot" class="runtime-proof-grid">
          <article v-for="item in strategySnapshotCards" :key="item.label" class="finance-list-card finance-list-card-panel">
            <p>{{ item.label }}</p>
            <strong>{{ item.value }}</strong>
            <span>{{ item.note }}</span>
          </article>
        </div>
        <div v-if="latestStrategySnapshot?.seedHighlights?.length" class="insight-chip-list">
          <span
            v-for="item in latestStrategySnapshot.seedHighlights"
            :key="`snapshot-${item}`"
            class="finance-pill finance-pill-compact finance-pill-info"
          >
            {{ item }}
          </span>
        </div>
      </div>
    </article>

    <article class="summary card">
      <header class="finance-copy-stack">
        <h2 class="section-title">我的会员配额</h2>
        <p class="section-subtitle">查看当前配额与使用情况</p>
      </header>
      <div class="summary-grid">
        <article v-for="item in quotaCards" :key="item.label">
          <p>{{ item.label }}</p>
          <strong>{{ item.value }}</strong>
        </article>
      </div>
      <div class="summary-actions">
        <label>
          支付方式
          <select v-model="payChannel">
            <option v-for="item in payChannelOptions" :key="item.value" :value="item.value">
              {{ item.label }}
            </option>
          </select>
        </label>
        <button type="button" class="finance-primary-btn" @click="loadMembershipData">立即刷新会员状态</button>
      </div>
      <div class="summary-steps">
        <span v-for="item in summarySteps" :key="item" class="finance-pill finance-pill-compact finance-pill-neutral">{{ item }}</span>
      </div>
      <p v-if="actionMessage" class="action-message finance-note-strip finance-note-strip-info">{{ actionMessage }}</p>
      <div v-if="latestPaymentAction" class="action-payment-entry">
        <button type="button" class="finance-primary-btn" @click="openLatestPaymentPage">
          {{ latestPaymentCTA }}
        </button>
        <a v-if="latestPaymentURL" :href="latestPaymentURL" target="_blank" rel="noopener">
          浏览器打开支付页
        </a>
      </div>
    </article>

    <div class="plans">
      <article
        v-for="plan in plans"
        :key="plan.id"
        class="plan card"
        :class="{ featured: plan.featured }"
      >
        <p class="tier">{{ plan.tier }}</p>
        <h2>{{ plan.name }}</h2>
        <p class="price">{{ plan.priceText }}</p>
        <p class="status-line">
          <span class="badge finance-pill finance-pill-compact" :class="plan.statusClass">{{ plan.statusLabel }}</span>
          <span>{{ plan.durationText }}</span>
        </p>
        <ul>
          <li v-for="feature in plan.features" :key="feature">{{ feature }}</li>
        </ul>
        <button
          class="finance-primary-btn"
          :disabled="plan.disabled || orderingProductID === plan.id"
          @click="handleCreateOrder(plan)"
        >
          {{ orderingProductID === plan.id ? "处理中..." : plan.actionText }}
        </button>
      </article>
    </div>

    <article class="ability card">
      <header class="finance-copy-stack">
        <h2 class="section-title">能力对比</h2>
        <p class="section-subtitle">按套餐状态、时长、等级和成本展示。</p>
      </header>
      <div class="ability-grid" v-if="abilityPlans.length > 0">
        <div class="ability-head" :style="abilityGridStyle">
          <span>能力项</span>
          <span v-for="plan in abilityPlans" :key="`head-${plan.id}`">{{ plan.name }}</span>
        </div>
        <div v-for="row in abilityRows" :key="row.name" class="ability-row" :style="abilityGridStyle">
          <span>{{ row.name }}</span>
          <span v-for="(value, idx) in row.values" :key="`${row.name}-${idx}`">{{ value }}</span>
        </div>
      </div>
      <div v-else class="empty-box finance-empty-box">暂无可对比的会员产品</div>
    </article>

    <article class="orders card">
      <header class="finance-copy-stack">
        <h2 class="section-title">订单记录</h2>
        <p class="section-subtitle">查看订单状态与支付进度</p>
      </header>

      <div class="order-summary-grid">
        <article v-for="item in orderSummaryCards" :key="item.label">
          <p>{{ item.label }}</p>
          <strong>{{ item.value }}</strong>
        </article>
      </div>

      <div class="order-toolbar">
        <label>
          <input v-model="autoTrackPendingOrders" type="checkbox" />
          自动追踪待支付订单（15 秒）
        </label>
        <button
          type="button"
          class="finance-mini-btn finance-mini-btn-primary"
          :disabled="manualRefreshing"
          @click="refreshOrdersManually"
        >
          {{ manualRefreshing ? "刷新中..." : "立即刷新订单" }}
        </button>
      </div>
      <p class="order-tip finance-note-strip finance-note-strip-info">
        当前待支付 {{ pendingOrderCount }} 笔，请在收银台完成真实支付
      </p>

      <div class="order-table-wrap finance-table-wrap" v-if="orderRows.length > 0">
        <table class="order-table finance-data-table">
          <thead>
            <tr>
              <th>订单号</th>
              <th>产品</th>
              <th>金额</th>
              <th>支付方式</th>
              <th>状态</th>
              <th>时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in orderRows" :key="item.orderNo">
              <td>{{ item.orderNo }}</td>
              <td>{{ item.productName }}</td>
              <td>{{ item.amount }}</td>
              <td>{{ item.payChannel }}</td>
              <td>
                <span class="badge finance-pill finance-pill-compact" :class="item.statusClass">{{ item.statusLabel }}</span>
              </td>
              <td>{{ item.time }}</td>
              <td class="order-actions">
                <button
                  v-if="canResumePayment(item)"
                  type="button"
                  class="mini-btn finance-mini-btn finance-mini-btn-primary"
                  @click="openLatestPaymentPage"
                >
                  继续支付
                </button>
                <span v-else-if="item.isPending">待支付</span>
                <span v-else>-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <StatePanel
        v-else
        tone="info"
        eyebrow="订单记录"
        title="暂无会员订单记录"
        description="选择任意套餐下单后，这里会展示最近的支付状态和订单结果。"
      >
        <template #actions>
          <button type="button" class="finance-primary-btn" @click="loadMembershipData">立即刷新会员状态</button>
        </template>
      </StatePanel>

    </article>
  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import StatePanel from "../../../components/StatePanel.vue";
import {
  createMembershipOrder,
  getMembershipQuota,
  listMembershipOrders,
  listMembershipProducts
} from "../../../api/membership";
import { getStockRecommendationInsight, listStockRecommendations } from "../../../api/market";
import { shouldUseDemoFallback } from "../../../lib/fallback-policy";
import {
  clearExperimentAttributionSources,
  createExperimentContext,
  listExperimentAttributionSources,
  trackExperimentEvent,
  trackExperimentExposureOnce
} from "../../../lib/growth-analytics";
import { getExperimentVariant } from "../../../lib/growth-experiments";
import { buildProfileModulePath } from "../../../lib/profile-modules";
import { buildStrategySnapshotCard } from "../../../lib/strategy-version";
import { WATCHLIST_EVENT, listWatchedStocks } from "../../../lib/watchlist";

const fallbackProducts = [
  {
    id: "mp_demo_001",
    name: "VIP月卡",
    price: 99,
    status: "ACTIVE",
    member_level: "VIP1",
    duration_days: 30
  },
  {
    id: "mp_demo_002",
    name: "VIP季卡",
    price: 269,
    status: "ACTIVE",
    member_level: "VIP2",
    duration_days: 90
  },
  {
    id: "mp_demo_003",
    name: "VIP年卡",
    price: 899,
    status: "ACTIVE",
    member_level: "VIP3",
    duration_days: 365
  }
];

const fallbackQuota = {
  member_level: "VIP1",
  activation_state: "ACTIVE",
  period_key: "2026-02",
  doc_read_limit: 100,
  doc_read_used: 24,
  doc_read_remaining: 76,
  news_subscribe_limit: 50,
  news_subscribe_used: 12,
  news_subscribe_remaining: 38,
  reset_cycle: "MONTHLY",
  reset_at: "2026-03-01T00:00:00+08:00",
  vip_expire_at: "2026-03-20T23:59:59+08:00",
  vip_status: "ACTIVE",
  vip_remaining_days: 20
};

const fallbackOrders = [
  {
    id: "mo_demo_001",
    order_no: "mo_demo_001",
    product_id: "mp_demo_001",
    amount: 99,
    pay_channel: "ALIPAY",
    status: "PAID",
    paid_at: "2026-02-24T11:00:00+08:00",
    created_at: "2026-02-24T10:50:00+08:00"
  }
];

const fallbackStrategySnapshot = {
  name: "600519.SH 贵州茅台",
  summary: "系统最近一次推荐解释显示，当前标的的资金、事件和风险边界仍能形成相对完整的共识链。",
  consensus: "多角色会先对趋势与风险做交叉验证，再决定是否进入推荐池。",
  meta: "生成 2026/02/27 09:30:00 · 版本 strategy-engine-demo",
  workload: {
    seedCount: 24,
    agentCount: 5,
    scenarioCount: 4,
    candidateCount: 8
  },
  seedHighlights: ["资金流延续", "板块轮动配合", "风险边界已定义"]
};

const payChannelOptions = [
  { value: "ALIPAY", label: "支付宝" },
  { value: "WECHAT", label: "微信支付" },
  { value: "CARD", label: "银行卡" },
  { value: "YOLKPAY", label: "蛋黄支付" }
];
const useDemoFallback = shouldUseDemoFallback();
const router = useRouter();
const watchlistModuleRoute = buildProfileModulePath("watchlist");
const membershipExperimentVariant = ref(getExperimentVariant("membership_copy", ["cadence", "proof"]));
const MEMBERSHIP_ATTRIBUTION_EXPERIMENT_KEYS = [
  "strategy_membership_cta",
  "home_membership_entry",
  "news_membership_entry",
  "archive_membership_entry"
];

const loading = ref(false);
const manualRefreshing = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const orderingProductID = ref("");
const actionMessage = ref("");
const latestPaymentAction = ref(null);
const latestPaymentOrderNo = ref("");
const payChannel = ref("YOLKPAY");
const autoTrackPendingOrders = ref(true);
const watchedStockCount = ref(0);
const strategySnapshotLoading = ref(false);
const strategySnapshotError = ref("");
const latestStrategySnapshot = ref(useDemoFallback ? fallbackStrategySnapshot : null);
let pendingOrderTimer = null;

const PAYMENT_PROTOCOL_ALLOWLIST = new Set([
  "http:",
  "https:",
  "alipay:",
  "alipays:",
  "weixin:",
  "wxp:",
  "unionpay:",
  "uppay:",
  "yolkpay:"
]);

const rawProducts = ref(useDemoFallback ? [...fallbackProducts] : []);
const rawQuota = ref(useDemoFallback ? { ...fallbackQuota } : {});
const rawOrders = ref(useDemoFallback ? [...fallbackOrders] : []);

const currentMemberLevel = computed(() => String(rawQuota.value?.member_level || "FREE").toUpperCase());
const currentActivationState = computed(() => {
  return String(rawQuota.value?.vip_status || "").toUpperCase() === "ACTIVE" ? "ACTIVE" : "NON_MEMBER";
});
const productMap = computed(() => {
  const mapping = {};
  rawProducts.value.forEach((item) => {
    if (item?.id) {
      mapping[item.id] = item;
    }
  });
  return mapping;
});

const plans = computed(() =>
  (rawProducts.value || [])
    .slice()
    .sort((a, b) => Number(a.price || 0) - Number(b.price || 0))
    .map((item, index) => {
      const level = String(item.member_level || "FREE").toUpperCase();
      const statusLabel = mapProductStatus(item.status);
      const statusClass = statusClassByStatus(item.status);
      const pending = hasPendingOrder(item.id);
      const isCurrent = level === currentMemberLevel.value;
      return {
        id: item.id || `plan_${index}`,
        tier: mapTier(level),
        name: item.name || level,
        priceText: formatPriceWithCycle(item.price, item.duration_days),
        durationText: formatDuration(item.duration_days),
        statusLabel,
        statusClass,
        features: buildPlanFeatures(item),
        featured: isCurrent,
        disabled: isCurrent || pending || String(item.status || "").toUpperCase() !== "ACTIVE",
        actionText: isCurrent
          ? "当前方案"
          : pending
            ? "先完成待支付订单"
            : String(item.status || "").toUpperCase() === "ACTIVE"
              ? "立即下单"
              : "暂不可购买"
      };
    })
);

const currentPlanName = computed(() => {
  const current = plans.value.find((item) => item.featured);
  return current?.name || mapLevelLabel(currentMemberLevel.value);
});

const currentLevelLabel = computed(() => `等级：${mapLevelLabel(currentMemberLevel.value)}`);
const vipStatusLabel = computed(() => mapVIPStatus(rawQuota.value?.vip_status, rawQuota.value?.member_level));
const activationStateLabel = computed(() => mapActivationState(currentActivationState.value));
const heroSubtitle = computed(() => "可查看套餐权益、订单记录，并直接创建会员订单。");
const vipExpireText = computed(() => formatDateTime(rawQuota.value?.vip_expire_at));
const vipRemainDaysText = computed(() => `${Math.max(0, Number(rawQuota.value?.vip_remaining_days || 0))} 天`);
const isVIPExpired = computed(() => String(rawQuota.value?.vip_status || "").toUpperCase() === "EXPIRED");
const isVIPActive = computed(() => {
  if (currentActivationState.value) {
    return currentActivationState.value === "ACTIVE";
  }
  const quotaStatus = String(rawQuota.value?.vip_status || "").toUpperCase();
  if (quotaStatus === "ACTIVE") {
    return true;
  }
  return currentMemberLevel.value.startsWith("VIP") && quotaStatus !== "EXPIRED";
});
const isProofVariant = computed(() => membershipExperimentVariant.value === "proof");
const membershipExperimentLabel = computed(() =>
  "会员建议"
);

const quotaCards = computed(() => [
  {
    label: "VIP状态",
    value: vipStatusLabel.value
  },
  {
    label: "VIP到期",
    value: vipExpireText.value
  },
  {
    label: "剩余天数",
    value: vipRemainDaysText.value
  },
  {
    label: "文档阅读",
    value: `${rawQuota.value.doc_read_used ?? 0} / ${rawQuota.value.doc_read_limit ?? 0}`
  },
  {
    label: "资讯订阅",
    value: `${rawQuota.value.news_subscribe_used ?? 0} / ${rawQuota.value.news_subscribe_limit ?? 0}`
  },
  {
    label: "重置周期",
    value: mapResetCycle(rawQuota.value.reset_cycle)
  },
  {
    label: "下次重置",
    value: formatDateTime(rawQuota.value.reset_at)
  }
]);

const abilityPlans = computed(() => plans.value.slice(0, 3));
const abilityGridStyle = computed(() => ({
  gridTemplateColumns: `1.1fr repeat(${abilityPlans.value.length}, minmax(0, 0.72fr))`
}));
const abilityRows = computed(() => [
  {
    name: "产品状态",
    values: abilityPlans.value.map((item) => item.statusLabel)
  },
  {
    name: "会员等级",
    values: abilityPlans.value.map((item) => item.tier.replace("版", ""))
  },
  {
    name: "服务时长",
    values: abilityPlans.value.map((item) => item.durationText)
  },
  {
    name: "价格",
    values: abilityPlans.value.map((item) => item.priceText)
  },
  {
    name: "月均成本",
    values: abilityPlans.value.map((item) => estimateMonthlyPrice(item.priceText))
  }
]);

const orderRows = computed(() =>
  (rawOrders.value || [])
    .slice()
    .sort((a, b) => {
      const ta = Date.parse(a.paid_at || a.created_at || "");
      const tb = Date.parse(b.paid_at || b.created_at || "");
      return (Number.isNaN(tb) ? 0 : tb) - (Number.isNaN(ta) ? 0 : ta);
    })
    .map((item) => ({
      orderNo: item.order_no || item.id || "-",
      productName: productMap.value[item.product_id]?.name || item.product_id || "-",
      amount: formatAmount(item.amount),
      amountValue: Number(item.amount || 0),
      payChannel: mapPayChannel(item.pay_channel),
      payChannelRaw: String(item.pay_channel || "").toUpperCase(),
      statusRaw: String(item.status || "").toUpperCase(),
      statusLabel: mapOrderStatus(item.status),
      statusClass: statusClassByStatus(item.status),
      time: formatDateTime(item.paid_at || item.created_at),
      isPending: isPendingStatus(item.status)
    }))
);

const pendingOrderCount = computed(
  () => orderRows.value.filter((item) => item.statusRaw === "PENDING").length
);
const latestPaymentURL = computed(() => resolvePaymentActionURL(latestPaymentAction.value));
const latestTrackedOrder = computed(() =>
  orderRows.value.find((item) => item.orderNo === latestPaymentOrderNo.value) || null
);
const latestPaymentCTA = computed(() => `继续去${mapPayChannel(latestPaymentAction.value?.channel)}支付`);
const summarySteps = computed(() => ["1 选择支付方式", "2 选择套餐并下单", "3 支付后回到本页刷新"]);
const membershipJourney = computed(() => {
  if (isVIPActive.value) {
    return {
      title: isProofVariant.value
        ? "会员已开通，建议先看历史兑现与复盘"
        : "会员已开通，建议按今日节奏使用内容",
      desc: isProofVariant.value
        ? "可先确认历史兑现和复盘价值，再回到今日内容继续查看。"
        : "先看开盘前主推，再跟午盘资讯，收盘回到关注列表，周末做历史复盘。",
      summaryLabel: "当前节奏",
      summaryValue: isProofVariant.value ? "历史复盘优先" : "会员使用中",
      summaryNote: isProofVariant.value
        ? `剩余 ${vipRemainDaysText.value} · 可先看历史兑现`
        : `剩余 ${vipRemainDaysText.value} · ${vipStatusLabel.value}`,
      primaryAction: { label: isProofVariant.value ? "先看历史档案" : "先看 08:30 主推荐", to: isProofVariant.value ? "/archive" : "/strategies" },
      secondaryAction: { label: isProofVariant.value ? "再回我的关注" : "15:30 回我的关注", to: watchlistModuleRoute }
    };
  }
  if (isVIPExpired.value) {
    return {
      title: isProofVariant.value ? "先看历史兑现样本，再决定是否续费" : "先恢复使用节奏，再决定是否续费",
      desc: isProofVariant.value
        ? "过期后可先看历史兑现和复盘质量，再决定是否续费。"
        : "建议先去历史档案看兑现情况，再到资讯页看盘中变化，确认价值后回到会员页续费。",
      summaryLabel: "当前状态",
      summaryValue: "会员待续费",
      summaryNote: `到期时间 ${vipExpireText.value}`,
      primaryAction: { label: "先看历史档案", to: "/archive" },
      secondaryAction: { label: "再看今日资讯", to: "/news" }
    };
  }
  return {
    title: isProofVariant.value ? "先看公开样本，再决定是否升级会员" : "先从今日主推荐开始，再决定是否升级会员",
    desc: isProofVariant.value
      ? "可先看历史样本和公开内容，再决定是否为盘中提醒和完整复盘升级。"
      : "可先查看今日主推荐、资讯和历史样本，再决定是否升级。",
    summaryLabel: "当前状态",
    summaryValue: isProofVariant.value ? "公开样本查看中" : "注册用户",
    summaryNote: isProofVariant.value ? "先看历史兑现，再决定升级" : "先看今日内容，再决定升级",
    primaryAction: { label: isProofVariant.value ? "查看历史样本" : "去看 08:30 主推荐", to: isProofVariant.value ? "/archive" : "/strategies" },
    secondaryAction: { label: isProofVariant.value ? "再看今日主推荐" : "查看历史样本", to: isProofVariant.value ? "/strategies" : "/archive" }
  };
});
const renewalExperimentPanel = computed(() => {
  if (isVIPExpired.value) {
    return {
      tone: "warning",
      eyebrow: membershipExperimentLabel.value,
      title: isProofVariant.value ? "先复盘历史兑现，再决定是否续费" : "先把节奏接回来，再决定是否续费",
      desc: isProofVariant.value
        ? "到期后可先看历史样本和复盘质量，再决定是否续费。"
        : "到期后可先恢复 08:30 / 11:30 / 15:30 / 周末 的查看节奏，再回会员页续费。",
      primaryAction: { label: "查看历史档案", to: "/archive" },
      secondaryAction: { label: "查看今日资讯", to: "/news" }
    };
  }
  if (isVIPActive.value) {
    return {
      tone: "success",
      eyebrow: membershipExperimentLabel.value,
      title: isProofVariant.value ? "在到期前先完成一轮历史复盘" : "在到期前先把今日节奏跑满",
      desc: isProofVariant.value
        ? "建议先看历史兑现、关注跟踪和复盘质量，再判断是否续费。"
        : "建议把会员能力用起来：看推荐、跟午盘、盯收盘、做周末复盘。",
      primaryAction: { label: isProofVariant.value ? "去看历史档案" : "去看策略页", to: isProofVariant.value ? "/archive" : "/strategies" },
      secondaryAction: { label: "去我的关注", to: watchlistModuleRoute }
    };
  }
  return {
    tone: "info",
    eyebrow: membershipExperimentLabel.value,
    title: isProofVariant.value ? "先验证公开样本，再做升级" : "先建立固定回访节奏，再做升级",
    desc: isProofVariant.value
      ? "可先看历史样本和兑现质量，再判断是否升级。"
      : "可先按固定节奏使用内容，再判断是否升级。",
    primaryAction: { label: isProofVariant.value ? "去看历史样本" : "去看主推荐", to: isProofVariant.value ? "/archive" : "/strategies" },
    secondaryAction: { label: "查看资讯页", to: "/news" }
  };
});
const cadenceEntries = computed(() => {
  const watchlistEntry = isVIPActive.value
    ? "会员收盘后回到我的关注页，继续跟进持仓与候选。"
    : "注册用户先在我的关注里形成自己的观察清单。";
  const archiveEntry = isVIPActive.value
    ? "周末用完整历史档案复盘推荐兑现和失效原因。"
    : "周末先看公开历史档案，决定是否升级解锁更完整复盘。";

  return [
    {
      slot: "08:30",
      title: "开盘前定方向",
      desc: isVIPActive.value
        ? "去策略页确认今日主推荐与解释逻辑，先完成一天的主决策。"
        : "先看今日主推荐，建立每天回来的第一个理由。",
      highlight: isVIPActive.value ? "入口：策略页主推荐" : "入口：试看版主推荐",
      supporting: isVIPActive.value ? "适合 3 分钟完成决策" : "看完再决定是否升级",
      primary: { label: "进入策略页", to: "/strategies" },
      secondary: { label: "回首页看节奏", to: "/home" }
    },
    {
      slot: "11:30",
      title: "午盘看变化",
      desc: "去资讯页确认午盘变化、相关新闻和盘中线索，不让早盘观点孤立存在。",
      highlight: "入口：资讯页",
      supporting: isVIPActive.value ? "会员可继续查看盘中信息" : "先建立盘中回访习惯",
      primary: { label: "进入资讯页", to: "/news" },
      secondary: { label: "去个人中心安排", to: "/profile" }
    },
    {
      slot: "15:30",
      title: "收盘后做跟踪",
      desc: watchlistEntry,
      highlight: "入口：我的关注",
      supporting: isVIPActive.value ? "把推荐转成持续跟踪" : "先把感兴趣标的留下来",
      primary: { label: "进入我的关注", to: watchlistModuleRoute },
      secondary: { label: "查看历史档案", to: "/archive" }
    },
    {
      slot: "周末",
      title: "用复盘建立信任",
      desc: archiveEntry,
      highlight: "入口：历史档案",
      supporting: isVIPActive.value ? "复盘胜率、节奏与退出原因" : "看公开样本再决定升级",
      primary: { label: "进入历史档案", to: "/archive" },
      secondary: { label: isVIPActive.value ? "去个人中心复盘" : "再看今日资讯", to: isVIPActive.value ? "/profile" : "/news" }
    }
  ];
});
const explanationCapabilityCards = computed(() => {
  const currentStage = resolveMembershipExperimentStage();
  return [
    {
      key: "REGISTERED",
      stage: "注册 / 游客",
      badge: currentStage === "REGISTERED" ? "当前阶段" : "基础解释层",
      title: "先看结论，再决定是否继续查看",
      desc: "可先看到主推荐、公开样本和基础理由。",
      points: ["首页结论层", "公开历史样本", "基础理由摘要"],
      unlocked: true,
      active: currentStage === "REGISTERED"
    },
    {
      key: "FOLLOWUP",
      stage: "持续跟踪",
      badge: currentStage === "VIP" ? "已解锁" : "登录后更完整",
      title: "把一次浏览变成持续跟踪",
      desc: "登录后可把标的留在我的关注，继续看跟踪原因、风险边界和资讯变化。",
      points: ["我的关注解释", "风险边界变化", "角色结论变化"],
      unlocked: true,
      active: currentStage !== "VIP"
    },
    {
      key: "VIP",
      stage: "会员",
      badge: currentStage === "VIP" ? "当前阶段" : "升级后解锁",
      title: "查看完整策略档案与复盘信息",
      desc: "会员可查看多角色模拟、场景推演、失效条件和历史复盘。",
      points: ["多角色模拟", "多场景推演", "失效条件与复盘"],
      unlocked: currentStage === "VIP",
      active: currentStage === "VIP"
    }
  ];
});
const explanationValueRows = computed(() => [
  {
    label: "升级后多了什么",
    value: isVIPActive.value ? "完整解释链" : "完整策略档案",
    note: isVIPActive.value
      ? "你现在能把首页、策略页、我的关注、历史档案串成一条完整解释链。"
      : "不仅知道选了什么，还能看到为什么选、什么情况下失效、后续怎么跟。"
  },
  {
    label: "节省的时间",
    value: isProofVariant.value ? "少走验证弯路" : "少做重复判断",
    note: isProofVariant.value
      ? "先看公开样本和历史兑现，再决定是否继续追踪或升级。"
      : "先给结论，再给过程和边界，不需要自己把碎片信息重新拼一遍。"
  },
  {
    label: "最关键的信任点",
    value: "边界清楚",
    note: "我们会把风险提醒、失效条件和角色分歧一起展示，而不是只说正向理由。"
  }
]);
const watchlistBridgeSummary = computed(() => {
  const count = watchedStockCount.value;
  if (isVIPActive.value) {
    return {
      title: count > 0 ? `当前有 ${count} 只关注股可继续跟踪` : "先把重点标的加入关注，再充分使用会员能力",
      desc:
        count > 0
          ? "可把会员内容落实到每日跟踪、风险边界和历史复盘。"
          : "升级后可把后续跟踪、风险边界和复盘内容一起查看。",
      points: ["我的关注持续跟踪", "风险边界变化", "历史复盘"]
    };
  }
  return {
    title: count > 0 ? `你已保存 ${count} 只关注股，下一步是把跟踪能力升级完整` : "先从保存关注开始，再决定是否升级会员",
    desc:
      count > 0
        ? "你已经开始形成自己的跟踪池，升级会员后会把多角色模拟、场景推演和复盘能力补齐。"
        : "先把感兴趣的标的留在我的关注，回来时再对照会员能力判断是否值得升级。",
    points: ["先保存关注", "再看持续跟踪", "最后补齐完整解释"]
  };
});
const strategySnapshotCards = computed(() => {
  const snapshot = latestStrategySnapshot.value;
  if (!snapshot) {
    return [];
  }
  return [
    {
      label: "种子处理",
      value: `${snapshot.workload?.seedCount || 0} 个`,
      note: "最近一次解释处理的市场种子数量"
    },
    {
      label: "参与角色",
      value: `${snapshot.workload?.agentCount || 0} 角`,
      note: snapshot.consensus || "多角色协同后再收敛结论"
    },
    {
      label: "场景推演",
      value: `${snapshot.workload?.scenarioCount || 0} 个`,
      note: "覆盖常态、顺势、回撤、冲击等不同情景"
    },
    {
      label: "候选筛选",
      value: `${snapshot.workload?.candidateCount || 0} 个`,
      note: "先过候选池，再决定是否发布到推荐池"
    }
  ];
});
const membershipFocusRows = computed(() => [
  {
    label: "当前套餐",
    value: currentPlanName.value,
    note: `${currentLevelLabel.value} · ${activationStateLabel.value}`
  },
  {
    label: "当前状态",
    value: membershipJourney.value.summaryValue,
    note: membershipJourney.value.summaryNote
  },
  {
    label: "待支付订单",
    value: `${pendingOrderCount.value} 笔`,
    note: pendingOrderCount.value > 0 ? "先完成待支付订单，再刷新会员状态。" : "当前没有待支付订单阻塞后续动作。"
  }
]);
const membershipValueGuideRows = computed(() => [
  {
    title: "为什么值得升级",
    summary: membershipJourney.value.title,
    desc: membershipJourney.value.desc
  },
  {
    title: "升级后多了什么",
    summary: explanationValueRows.value[0]?.value || "完整解释链",
    desc: explanationValueRows.value[0]?.note || "升级后可查看更完整的策略、关注、档案和资讯内容。"
  },
  {
    title: "当前最该做什么",
    summary: renewalExperimentPanel.value.title,
    desc: renewalExperimentPanel.value.desc
  }
]);
const membershipPageGuideRows = computed(() => [
  {
    title: "先看权益差异",
    desc: "先确认升级后能看到什么，再决定套餐与支付。"
  },
  {
    title: "状态清晰",
    desc: "已开通、待实名激活、待支付、已过期等状态都会明确显示。"
  },
  {
    title: "价格放在后面",
    desc: "先看升级后能获得的内容，再查看套餐与订单。"
  }
]);
const membershipActivationGuideDesc = computed(() => {
  if (isVIPActive.value) {
    return "当前状态已可直接使用今日内容，建议回到策略、资讯、关注和档案页继续查看。";
  }
  if (pendingOrderCount.value > 0) {
    return "先把待支付订单处理完，再回到本页刷新结果，避免支付状态和后续激活链路脱节。";
  }
  return "这页会显示套餐、订单、支付和激活状态，先明确今天要补齐哪一层能力，再决定是否升级。";
});
const membershipCurrentTodoRows = computed(() => [
  {
    title: "当前状态已可直接查看后续内容",
    desc: "当前可以直接去策略、资讯、关注和档案页把会员节奏跑起来。"
  },
  {
    title: pendingOrderCount.value > 0 ? "当前仍有待支付订单" : "当前没有待支付订单",
    desc: pendingOrderCount.value > 0
      ? "完成支付后回到本页刷新订单结果，避免一边看套餐一边丢失支付状态。"
      : "下单后这里会继续显示支付状态和支付入口。"
  },
  {
    title: "下一步操作",
    desc: `建议先${membershipJourney.value.primaryAction.label}，再回本页继续核对套餐、订单、支付和激活状态。`
  }
]);
const membershipStatus = computed(() => {
  if (loading.value) {
    return {
      tone: "info",
      label: "同步中",
      title: "正在刷新会员套餐、配额和订单",
      desc: "请稍候，再继续下单或刷新支付状态。"
    };
  }
  if (loadError.value) {
    return {
      tone: "warning",
      label: "需处理",
      title: "会员数据同步失败",
      desc: loadError.value
    };
  }
  if (latestTrackedOrder.value?.statusRaw === "PAID" || latestTrackedOrder.value?.statusRaw === "SUCCESS") {
    return {
      tone: "success",
      label: "已支付",
      title: `订单 ${latestTrackedOrder.value.orderNo} 已支付成功`,
      desc: "会员权益会在后续同步中生效，如有延迟可手动刷新会员数据。"
    };
  }
  if (latestTrackedOrder.value?.statusRaw === "PENDING") {
    return {
      tone: "warning",
      label: "待完成",
      title: `订单 ${latestTrackedOrder.value.orderNo} 仍待支付`,
      desc: latestPaymentURL.value
        ? "你可以继续前往收银台完成支付，完成后回到本页刷新结果。"
        : "当前暂无可恢复的支付链接，请联系管理员或重新下单。"
    };
  }
  if (
    latestTrackedOrder.value?.statusRaw === "FAILED" ||
    latestTrackedOrder.value?.statusRaw === "CANCELED" ||
    latestTrackedOrder.value?.statusRaw === "CANCELLED"
  ) {
    return {
      tone: "warning",
      label: "未完成",
      title: `订单 ${latestTrackedOrder.value.orderNo} 未支付成功`,
      desc: "建议重新选择套餐发起支付，或联系管理员确认支付通道状态。"
    };
  }
  if (actionMessage.value) {
    return {
      tone: latestPaymentURL.value ? "success" : "info",
      label: latestPaymentURL.value ? "下一步" : "已更新",
      title: actionMessage.value,
      desc: latestPaymentURL.value
        ? "完成收银台支付后，回到本页点击“手动刷新订单”确认结果。"
        : "当前可以继续选择套餐，或刷新会员数据查看最新状态。"
    };
  }
  if (pendingOrderCount.value > 0) {
    return {
      tone: "warning",
      label: "待支付",
      title: `还有 ${pendingOrderCount.value} 笔订单待完成`,
      desc: "完成支付后回到本页刷新订单状态，系统会自动追踪待支付订单。"
    };
  }
  return {
    tone: "success",
    label: "准备就绪",
    title: "套餐、配额和订单都已同步",
    desc: `当前支付方式：${mapPayChannel(payChannel.value)}，可以直接选择套餐下单。`
  };
});

const orderSummaryCards = computed(() => {
  const rows = orderRows.value;
  const paidRows = rows.filter((item) => item.statusRaw === "PAID" || item.statusRaw === "SUCCESS");
  const pendingRows = rows.filter((item) => item.statusRaw === "PENDING");
  const failedRows = rows.filter((item) => item.statusRaw === "FAILED" || item.statusRaw === "CANCELED");
  const refundRows = rows.filter((item) => item.statusRaw === "REFUNDED" || item.statusRaw === "REFUND");
  const paidAmount = paidRows.reduce((sum, item) => sum + (Number.isFinite(item.amountValue) ? item.amountValue : 0), 0);

  return [
    { label: "已支付", value: `${paidRows.length} 笔` },
    { label: "待支付", value: `${pendingRows.length} 笔` },
    { label: "失败/取消", value: `${failedRows.length} 笔` },
    { label: "已退款", value: `${refundRows.length} 笔` },
    { label: "累计支付", value: formatAmount(paidAmount) }
  ];
});

function resolveMembershipExperimentStage() {
  if (isVIPActive.value) {
    return "VIP";
  }
  if (isVIPExpired.value) {
    return "EXPIRED";
  }
  return "REGISTERED";
}

function normalizeAnalyticsKey(value) {
  return String(value || "")
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/^_+|_+$/g, "");
}

function trackMembershipExperiment(eventType, targetKey, metadata = {}) {
  return trackExperimentEvent({
    experimentKey: "membership_copy",
    variantKey: membershipExperimentVariant.value,
    eventType,
    pageKey: "membership",
    targetKey,
    userStage: resolveMembershipExperimentStage(),
    metadata
  });
}

function buildOrderExperimentPayload(plan) {
  const sourceExperiments = listExperimentAttributionSources({
    excludeExperimentKey: "membership_copy"
  });
  return createExperimentContext({
    experimentKey: "membership_copy",
    variantKey: membershipExperimentVariant.value,
    pageKey: "membership",
    targetKey: `order_${normalizeAnalyticsKey(plan?.id || "unknown")}`,
    userStage: resolveMembershipExperimentStage(),
    metadata: {
      product_id: plan?.id || "",
      plan_name: plan?.name || "",
      pay_channel: payChannel.value,
      source_experiments: sourceExperiments
    }
  });
}

function trackMembershipExposure() {
  const commonMetadata = {
    current_level: currentMemberLevel.value,
    vip_status: String(rawQuota.value?.vip_status || "").toUpperCase(),
    activation_state: currentActivationState.value,
    kyc_status: "APPROVED"
  };
  trackExperimentExposureOnce({
    experimentKey: "membership_copy",
    variantKey: membershipExperimentVariant.value,
    pageKey: "membership",
    targetKey: "journey_panel",
    userStage: resolveMembershipExperimentStage(),
    metadata: commonMetadata
  });
  trackExperimentExposureOnce({
    experimentKey: "membership_copy",
    variantKey: membershipExperimentVariant.value,
    pageKey: "membership",
    targetKey: "renewal_panel",
    userStage: resolveMembershipExperimentStage(),
    metadata: commonMetadata
  });
}

function handleMembershipJourneyAction(targetKey, action) {
  trackMembershipExperiment("CLICK", targetKey, {
    destination: action?.to || "",
    action_label: action?.label || ""
  });
  goToRoute(action?.to);
}

function handleRenewalAction(targetKey, action) {
  trackMembershipExperiment("CLICK", targetKey, {
    destination: action?.to || "",
    action_label: action?.label || ""
  });
  goToRoute(action?.to);
}

function handleCadenceAction(entry, actionKey) {
  const action = entry?.[actionKey];
  const slotKey = normalizeAnalyticsKey(entry?.slot || "slot");
  trackMembershipExperiment("CLICK", `cadence_${slotKey}_${actionKey}`, {
    destination: action?.to || "",
    action_label: action?.label || "",
    cadence_slot: entry?.slot || "",
    cadence_title: entry?.title || ""
  });
  goToRoute(action?.to);
}

function goToRoute(path) {
  if (!path) {
    return;
  }
  if (router.currentRoute.value.path === path) {
    return;
  }
  router.push(path);
}

function refreshWatchedStockCount() {
  watchedStockCount.value = listWatchedStocks().length;
}

async function loadMembershipData(options = {}) {
  const { keepActionMessage = false, silent = false } = options;
  if (silent && loading.value) {
    return;
  }
  if (!silent) {
    loading.value = true;
  }
  if (!silent) {
    loadError.value = "";
  }
  if (!keepActionMessage) {
    actionMessage.value = "";
  }

  const errors = [];
  const [productsResult, quotaResult, ordersResult] = await Promise.allSettled([
    listMembershipProducts({ status: "ACTIVE", page: 1, page_size: 20 }),
    getMembershipQuota(),
    listMembershipOrders({ page: 1, page_size: 20 })
  ]);

  if (productsResult.status === "fulfilled" && Array.isArray(productsResult.value?.items)) {
    if (productsResult.value.items.length > 0) {
      rawProducts.value = productsResult.value.items;
    }
  } else if (productsResult.status === "rejected") {
    errors.push(`会员产品加载失败：${productsResult.reason?.message || "unknown error"}`);
  }

  if (quotaResult.status === "fulfilled" && quotaResult.value) {
    rawQuota.value = {
      ...rawQuota.value,
      ...quotaResult.value
    };
  } else if (quotaResult.status === "rejected") {
    errors.push(`会员配额加载失败：${quotaResult.reason?.message || "unknown error"}`);
  }

  if (ordersResult.status === "fulfilled" && Array.isArray(ordersResult.value?.items)) {
    rawOrders.value = ordersResult.value.items;
  } else if (ordersResult.status === "rejected") {
    errors.push(`会员订单加载失败：${ordersResult.reason?.message || "unknown error"}`);
  }

  if (errors.length > 0) {
    loadError.value = errors.join("；");
  }
  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  await loadLatestStrategySnapshot({ silent: true });
  if (!silent) {
    loading.value = false;
  }
}

async function loadLatestStrategySnapshot(options = {}) {
  const { silent = false } = options;
  strategySnapshotError.value = "";
  if (!silent) {
    strategySnapshotLoading.value = true;
  }
  try {
    const data = await listStockRecommendations({ page: 1, page_size: 1 });
    const latest = Array.isArray(data?.items) ? data.items[0] : null;
    if (!latest?.id) {
      if (!useDemoFallback) {
        latestStrategySnapshot.value = null;
      }
      return;
    }
    const insight = await getStockRecommendationInsight(latest.id);
    latestStrategySnapshot.value = mapLatestStrategySnapshot(latest, insight);
  } catch (error) {
    strategySnapshotError.value = parseErrorMessage(error);
    if (!latestStrategySnapshot.value && useDemoFallback) {
      latestStrategySnapshot.value = fallbackStrategySnapshot;
    }
  } finally {
    if (!silent) {
      strategySnapshotLoading.value = false;
    }
  }
}

async function handleCreateOrder(plan) {
  if (!plan?.id || plan.disabled || orderingProductID.value) {
    return;
  }

  trackMembershipExperiment("UPGRADE_INTENT", `order_${normalizeAnalyticsKey(plan.id)}`, {
    product_id: plan.id,
    plan_name: plan.name,
    pay_channel: payChannel.value
  });
  orderingProductID.value = plan.id;
  actionMessage.value = "";
  latestPaymentAction.value = null;
  latestPaymentOrderNo.value = "";

  try {
    const result = await createMembershipOrder({
      product_id: plan.id,
      pay_channel: payChannel.value,
      experiment: buildOrderExperimentPayload(plan)
    });
    const order = result?.order || result || {};
    const orderNo = order?.order_no || order?.id || "-";
    latestPaymentOrderNo.value = orderNo;
    const initialized = result?.payment_initialized;
    const action = result?.payment_action || null;
    latestPaymentAction.value = action;
    if (initialized === false) {
      actionMessage.value = `订单 ${orderNo} 已创建，但支付下单失败：${result?.payment_error || "请稍后重试"}`;
    } else if (action) {
      const safePaymentURL = resolvePaymentActionURL(action);
      const opened = openPaymentLink(action);
      actionMessage.value = opened
        ? `订单 ${orderNo} 已创建，已为你打开支付页，请完成支付后返回刷新状态`
        : safePaymentURL
          ? `订单 ${orderNo} 已创建，请点击“立即前往收银台”完成支付`
          : `订单 ${orderNo} 已创建，但支付链接未通过安全校验，请联系管理员`;
    } else {
      actionMessage.value = `订单 ${orderNo} 已创建，当前状态：${mapOrderStatus(order?.status)}`;
    }
    MEMBERSHIP_ATTRIBUTION_EXPERIMENT_KEYS.forEach((experimentKey) => {
      clearExperimentAttributionSources(experimentKey);
    });
    await loadMembershipData({ keepActionMessage: true });
  } catch (error) {
    actionMessage.value = `创建订单失败：${error?.message || "unknown error"}`;
  } finally {
    orderingProductID.value = "";
  }
}

async function refreshOrdersManually() {
  if (manualRefreshing.value) {
    return;
  }
  manualRefreshing.value = true;
  try {
    await loadMembershipData({ keepActionMessage: true });
  } finally {
    manualRefreshing.value = false;
  }
}

function openLatestPaymentPage() {
  const rawPaymentTarget = getRawPaymentActionTarget(latestPaymentAction.value);
  if (!openPaymentLink(latestPaymentAction.value)) {
    actionMessage.value = rawPaymentTarget
      ? "当前订单的支付链接未通过安全校验，请联系管理员"
      : "当前订单未返回可打开的支付链接";
  }
}

function canResumePayment(item) {
  return Boolean(
    item?.statusRaw === "PENDING" && item.orderNo === latestPaymentOrderNo.value && latestPaymentURL.value
  );
}

function getRawPaymentActionTarget(action) {
  if (!action) {
    return "";
  }
  return String(action.pay_url || action.qrcode || action.urlscheme || "").trim();
}

function resolvePaymentActionURL(action) {
  const rawTarget = getRawPaymentActionTarget(action);
  if (!rawTarget) {
    return "";
  }
  if (rawTarget.startsWith("/")) {
    const origin =
      typeof window !== "undefined" && window.location?.origin ? window.location.origin : "http://localhost";
    try {
      const target = new URL(rawTarget, origin);
      return PAYMENT_PROTOCOL_ALLOWLIST.has(target.protocol) ? target.toString() : "";
    } catch {
      return "";
    }
  }
  try {
    const target = new URL(rawTarget);
    return PAYMENT_PROTOCOL_ALLOWLIST.has(target.protocol) ? target.toString() : "";
  } catch {
    return "";
  }
}

function openPaymentLink(action) {
  const target = resolvePaymentActionURL(action);
  if (!target) {
    return false;
  }
  if (typeof window === "undefined") {
    return false;
  }
  const opened = window.open(target, "_blank", "noopener,noreferrer");
  if (!opened) {
    window.location.assign(target);
  }
  return true;
}

function syncPendingOrderTracking() {
  stopPendingOrderTracking();
  if (!autoTrackPendingOrders.value || pendingOrderCount.value <= 0) {
    return;
  }
  pendingOrderTimer = window.setInterval(() => {
    loadMembershipData({ keepActionMessage: true, silent: true });
  }, 15000);
}

function stopPendingOrderTracking() {
  if (pendingOrderTimer) {
    window.clearInterval(pendingOrderTimer);
    pendingOrderTimer = null;
  }
}

function buildPlanFeatures(item) {
  const level = mapLevelLabel(item.member_level);
  return [
    `${level}能力包`,
    `${formatDuration(item.duration_days)}有效期`,
    `文档阅读与订阅配额按${level}生效`
  ];
}

function hasPendingOrder(productID) {
  return rawOrders.value.some((item) => {
    if (item.product_id !== productID) {
      return false;
    }
    const status = String(item.status || "").toUpperCase();
    return status === "PENDING";
  });
}

function mapTier(level) {
  const normalized = String(level || "").toUpperCase();
  if (normalized === "FREE") return "基础版";
  if (normalized === "VIP1") return "进阶版";
  if (normalized === "VIP2" || normalized === "VIP3") return "旗舰版";
  return "会员版";
}

function mapLevelLabel(level) {
  const normalized = String(level || "").toUpperCase();
  if (normalized === "FREE") return "Explorer";
  if (normalized === "VIP1") return "Pro";
  if (normalized === "VIP2") return "Elite";
  if (normalized === "VIP3") return "Summit";
  return normalized || "-";
}

function mapProductStatus(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "ACTIVE") return "可购买";
  if (normalized === "DISABLED") return "已下线";
  return normalized || "-";
}

function mapOrderStatus(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "PAID" || normalized === "SUCCESS") return "已支付";
  if (normalized === "PENDING") return "待支付";
  if (normalized === "FAILED") return "失败";
  if (normalized === "CANCELED" || normalized === "CANCELLED") return "已取消";
  if (normalized === "REFUNDED" || normalized === "REFUND") return "已退款";
  return normalized || "-";
}

function statusClassByStatus(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "PAID" || normalized === "SUCCESS" || normalized === "ACTIVE") return "success";
  if (normalized === "PENDING") return "warning";
  if (normalized === "DISABLED") return "muted";
  return "danger";
}

function isPendingStatus(status) {
  return String(status || "").toUpperCase() === "PENDING";
}

function mapPayChannel(channel) {
  const normalized = String(channel || "").toUpperCase();
  if (normalized === "ALIPAY") return "支付宝";
  if (normalized === "WECHAT") return "微信";
  if (normalized === "CARD" || normalized === "BANK") return "银行卡";
  if (normalized === "YOLKPAY") return "蛋黄支付";
  if (normalized === "UNIONPAY") return "银联";
  return normalized || "-";
}

function formatPriceWithCycle(price, durationDays) {
  const amount = formatAmount(price);
  const days = Number(durationDays);
  if (!Number.isFinite(days) || days <= 0) {
    return `${amount} / 周期`;
  }
  if (days >= 360) {
    return `${amount} / 年`;
  }
  if (days >= 85) {
    return `${amount} / 季`;
  }
  if (days >= 28) {
    return `${amount} / 月`;
  }
  return `${amount} / ${days}天`;
}

function formatDuration(durationDays) {
  const days = Number(durationDays);
  if (!Number.isFinite(days) || days <= 0) {
    return "周期未配置";
  }
  if (days >= 360) return "365天有效";
  if (days >= 85) return "90天有效";
  if (days >= 28) return "30天有效";
  return `${days}天有效`;
}

function formatAmount(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "¥0";
  }
  return `¥${num.toFixed(2).replace(/\.00$/, "")}`;
}

function estimateMonthlyPrice(priceText) {
  const matched = String(priceText || "").match(/\d+(\.\d+)?/);
  const amount = Number(matched?.[0] || 0);
  if (!Number.isFinite(amount) || amount <= 0) {
    return "-";
  }
  if (priceText.includes("/ 年")) {
    return `约 ¥${(amount / 12).toFixed(1)} / 月`;
  }
  if (priceText.includes("/ 季")) {
    return `约 ¥${(amount / 3).toFixed(1)} / 月`;
  }
  if (priceText.includes("/ 月")) {
    return `约 ¥${amount.toFixed(1)} / 月`;
  }
  return "-";
}

function mapResetCycle(value) {
  const normalized = String(value || "").toUpperCase();
  if (normalized === "MONTHLY") return "每月";
  if (normalized === "WEEKLY") return "每周";
  if (normalized === "DAILY") return "每天";
  return normalized || "-";
}

function mapVIPStatus(value, level) {
  const normalized = String(value || "").toUpperCase();
  if (normalized === "ACTIVE") return "生效中";
  if (normalized === "EXPIRED") return "已到期";
  const levelText = String(level || "").toUpperCase();
  if (levelText.startsWith("VIP")) {
    return "生效中";
  }
  return "未开通";
}


function mapActivationState(value) {
  const normalized = String(value || "").toUpperCase();
  if (normalized === "ACTIVE") return "已激活";
  return "已开通";
}

function mapLatestStrategySnapshot(recommendation, insight) {
  const explanation = insight?.explanation || {};
  return buildStrategySnapshotCard(recommendation, explanation, formatDateTime);
}

function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return value || "-";
  }
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

watch([autoTrackPendingOrders, pendingOrderCount], () => {
  syncPendingOrderTracking();
});

onMounted(async () => {
  refreshWatchedStockCount();
  await loadMembershipData();
  syncPendingOrderTracking();
  trackMembershipExposure();
  if (typeof window !== "undefined") {
    window.addEventListener(WATCHLIST_EVENT, refreshWatchedStockCount);
  }
});

onBeforeUnmount(() => {
  stopPendingOrderTracking();
  if (typeof window !== "undefined") {
    window.removeEventListener(WATCHLIST_EVENT, refreshWatchedStockCount);
  }
});
</script>

<style scoped>
.membership-page {
  display: grid;
  gap: 12px;
}

.member-hero {
  border-radius: 20px;
  padding: 16px;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 12px;
  align-items: end;
}

.member-hero-copy {
  min-width: 0;
}

.member-hero-stats {
  grid-column: 1 / -1;
}

.hero-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.hero-pill {
  padding: 8px 11px;
}

.hero-pill p {
  font-size: 11px;
}

.hero-pill strong {
  margin-top: 2px;
}

.hero-pill small {
  margin-top: 3px;
}

.api-tip {
  border-radius: 10px;
  border: 1px dashed var(--color-border-soft-heavy);
  background: var(--color-surface-panel-soft-muted);
  padding: 8px 10px;
}

.api-tip p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.membership-focus-layout {
  --finance-main-column: minmax(0, 1.7fr);
  --finance-side-column: minmax(280px, 0.9fr);
}

.membership-focus-actions button {
  width: auto;
}

.membership-focus-actions button:not(.ghost) {
  color: #fff;
}

.membership-focus-actions .ghost {
  color: var(--color-pine-700);
}

.membership-status-grid article,
.membership-explain-grid article,
.membership-side-list article {
  display: grid;
  gap: 6px;
}

.membership-status-grid p,
.membership-status-grid strong,
.membership-status-grid span,
.membership-explain-grid p,
.membership-explain-grid strong,
.membership-explain-grid span,
.membership-side-list strong,
.membership-side-list p {
  margin: 0;
}

.membership-status-grid p,
.membership-explain-grid p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.membership-status-grid strong,
.membership-explain-grid strong,
.membership-side-list strong {
  font-size: 16px;
  line-height: 1.45;
  color: var(--color-pine-700);
}

.membership-status-grid span,
.membership-explain-grid span,
.membership-side-list p {
  font-size: 13px;
  line-height: 1.68;
  color: var(--color-text-sub);
}

.experiment-badge {
  margin-top: 8px;
  letter-spacing: 0.01em;
}

.cadence-hub {
  padding: 14px;
}

.insight-upgrade,
.summary {
  padding: 14px;
}

.cadence-head {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 12px;
  align-items: end;
}

.cadence-summary {
  min-width: 200px;
}

.cadence-summary p,
.cadence-summary small {
  margin: 0;
}

.cadence-summary strong {
  margin: 4px 0;
}

.cadence-grid {
  margin-top: 12px;
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.cadence-card {
  display: grid;
  gap: 8px;
}

.cadence-slot {
  margin: 0;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.12em;
  color: var(--color-pine-700);
}

.cadence-card h3 {
  margin: 0;
  font-size: 18px;
  line-height: 1.3;
}

.cadence-desc {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--color-text-sub);
}

.cadence-meta {
  display: grid;
  gap: 6px;
}

.cadence-actions {
  margin-top: auto;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.cadence-actions button {
  width: 100%;
}

.cadence-actions button:not(.ghost) {
  color: #fff;
}

.cadence-actions .ghost {
  color: var(--color-pine-700);
}

.insight-stage-grid {
  margin-top: 12px;
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.insight-stage-card,
.insight-value-grid article {
  display: grid;
  gap: 8px;
}

.insight-stage-card {
  opacity: 0.9;
}

.insight-stage-card.unlocked {
  opacity: 1;
}

.insight-stage-card.active {
  border-color: var(--color-border-focus);
  background: var(--gradient-card-active);
  box-shadow: var(--shadow-card-active);
}

.insight-stage-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.insight-stage-head p,
.insight-stage-card strong,
.insight-stage-card p:last-of-type {
  margin: 0;
}

.insight-stage-head p {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-pine-700);
}

.insight-stage-card strong {
  font-size: 16px;
  line-height: 1.45;
  color: var(--color-text-main);
}

.insight-stage-card p:last-of-type {
  font-size: 13px;
  line-height: 1.68;
  color: var(--color-text-sub);
}

.insight-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.insight-value-grid {
  margin-top: 12px;
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.insight-value-grid p,
.insight-value-grid strong,
.insight-value-grid span {
  margin: 0;
}

.insight-value-grid p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.insight-value-grid strong {
  font-size: 16px;
  color: var(--color-pine-700);
}

.insight-value-grid span {
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.insight-bridge-box {
  margin-top: 12px;
  border-radius: 14px;
  border: 1px solid var(--color-border-soft);
  background: linear-gradient(160deg, rgba(244, 248, 255, 0.94), rgba(255, 255, 255, 0.96));
  padding: 12px;
  display: grid;
  gap: 8px;
}

.insight-bridge-box p,
.insight-bridge-box strong,
.insight-bridge-box span {
  margin: 0;
}

.insight-bridge-box p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.insight-bridge-box strong {
  font-size: 16px;
  color: var(--color-pine-700);
}

.insight-bridge-box > span {
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.runtime-proof-box {
  margin-top: 12px;
  display: grid;
  gap: 10px;
}

.runtime-proof-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.runtime-proof-head p,
.runtime-proof-head strong,
.runtime-proof-head span,
.runtime-proof-note {
  margin: 0;
}

.runtime-proof-head p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.runtime-proof-head strong {
  display: block;
  margin-top: 4px;
  font-size: 16px;
  color: var(--color-pine-700);
}

.runtime-proof-head span {
  font-size: 12px;
  line-height: 1.6;
  color: var(--color-text-sub);
}

.runtime-proof-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.runtime-proof-grid article {
  border-radius: 10px;
  border: 1px solid var(--color-border-soft);
  background: var(--color-surface-panel);
  padding: 8px 10px;
  display: grid;
  gap: 4px;
}

.runtime-proof-grid p,
.runtime-proof-grid strong,
.runtime-proof-grid span {
  margin: 0;
}

.runtime-proof-grid p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.runtime-proof-grid strong {
  font-size: 15px;
  color: var(--color-text-main);
}

.runtime-proof-grid span {
  font-size: 12px;
  line-height: 1.55;
  color: var(--color-text-sub);
}

.summary-grid {
  margin-top: 10px;
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.summary-grid article {
  border: 1px solid var(--color-border-soft);
  border-radius: 11px;
  background: var(--color-surface-card-soft);
  padding: 9px 10px;
}

.summary-grid p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.summary-grid strong {
  margin-top: 4px;
  display: block;
  font-size: 16px;
  color: var(--color-pine-700);
}

.summary-actions {
  margin-top: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.summary-actions label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--color-text-sub);
  font-size: 13px;
}

.summary-actions select {
  border-radius: 9px;
  border: 1px solid var(--color-border-soft-heavy);
  background: #fff;
  padding: 5px 9px;
}

.summary-actions button {
  width: auto;
}

.summary-steps {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.action-message {
  margin: 10px 0 0;
  font-size: 12px;
  color: var(--color-pine-700);
}

.action-payment-entry {
  margin-top: 8px;
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.action-payment-entry button {
  width: auto;
}

.action-payment-entry a {
  color: var(--color-pine-600);
  font-size: 12px;
}

.plans {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.plan {
  padding: 14px;
  border-radius: 16px;
  display: flex;
  flex-direction: column;
}

.plan.featured {
  border-color: var(--color-border-focus-strong);
  background: linear-gradient(160deg, rgba(230, 238, 252, 0.82), rgba(255, 255, 255, 0.94));
  box-shadow: var(--shadow-strong);
}

.tier {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

h2 {
  margin: 8px 0 6px;
  font-family: var(--font-serif);
  font-size: 24px;
}

.price {
  margin: 0 0 10px;
  font-size: 24px;
  font-weight: 700;
}

.status-line {
  margin: 0 0 10px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  color: var(--color-text-sub);
  font-size: 12px;
}

ul {
  margin: 0;
  padding-left: 18px;
  display: grid;
  gap: 7px;
  color: var(--color-text-sub);
  font-size: 14px;
  line-height: 1.55;
  flex: 1;
}

.plan button {
  margin-top: 14px;
}

.plan.featured button {
  background: linear-gradient(145deg, #d2a766, #b57e38);
}

.plan button:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.badge {
  font-weight: 700;
}

.badge.success {
  color: var(--color-success);
  background: rgba(201, 229, 211, 0.72);
  border-color: rgba(46, 125, 50, 0.16);
}

.badge.warning {
  color: var(--color-warning);
  background: rgba(243, 228, 194, 0.84);
  border-color: rgba(184, 130, 48, 0.16);
}

.badge.danger {
  color: var(--color-danger);
  background: rgba(237, 198, 190, 0.68);
  border-color: rgba(178, 58, 42, 0.14);
}

.badge.muted {
  color: #52616d;
  background: rgba(232, 238, 246, 0.74);
  border-color: var(--color-border-soft);
}

.ability {
  padding: 14px;
}

.ability-grid {
  margin-top: 10px;
  display: grid;
  gap: 8px;
}

.ability-head,
.ability-row {
  border-radius: 11px;
  border: 1px solid var(--color-border-soft);
  padding: 8px 10px;
  display: grid;
  gap: 8px;
}

.ability-head {
  background: rgba(244, 248, 255, 0.84);
  color: var(--color-text-sub);
  font-size: 12px;
}

.ability-row {
  background: rgba(255, 255, 255, 0.88);
  font-size: 13px;
}

.ability-row span:not(:first-child) {
  color: var(--color-pine-700);
  font-weight: 600;
}

.empty-box {
  margin-top: 10px;
  border-radius: 11px;
  border: 1px dashed var(--color-border-soft-heavy);
  background: var(--color-surface-panel-soft-faint);
  padding: 10px;
  color: var(--color-text-sub);
  font-size: 12px;
}

.orders {
  padding: 14px;
}

.order-summary-grid {
  margin-top: 10px;
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(5, minmax(0, 1fr));
}

.order-summary-grid article {
  border: 1px solid var(--color-border-soft);
  border-radius: 10px;
  background: var(--color-surface-card-soft);
  padding: 8px 9px;
}

.order-summary-grid p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.order-summary-grid strong {
  margin-top: 3px;
  display: block;
  font-size: 15px;
  color: var(--color-pine-700);
}

.order-toolbar {
  margin-top: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
}

.order-toolbar label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--color-text-sub);
  font-size: 12px;
}

.order-toolbar input[type="checkbox"] {
  accent-color: var(--color-pine-600);
}

.order-toolbar button {
  flex-shrink: 0;
}

.order-toolbar button:disabled {
  opacity: 0.72;
}

.order-tip {
  margin: 7px 0 0;
}

.order-table {
  min-width: 720px;
}

.order-actions {
  min-width: 110px;
}

.mini-btn {
  min-width: 76px;
}

.mini-btn:disabled {
  opacity: 0.72;
}

</style>
