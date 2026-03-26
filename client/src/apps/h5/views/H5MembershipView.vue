<template>
  <div class="h5-page fade-up membership-page">
    <div class="h5-page-topline">
      <span class="h5-page-tagline">会员中心</span>
      <span>{{ lastUpdatedAt || "收银台" }}</span>
    </div>

    <section class="membership-hero">
      <div class="membership-hero-topline">
        <span>SercherAI 会员收银台</span>
        <button type="button" class="membership-refresh" :disabled="loading" @click="loadMembershipPage">
          {{ loading ? "同步中" : "刷新" }}
        </button>
      </div>

      <div class="membership-hero-copy">
        <h1>{{ heroTitle }}</h1>
        <p>{{ heroDescription }}</p>
      </div>

      <div class="membership-status-strip">
        <span class="membership-status-pill brand">{{ currentLevelText }}</span>
        <span class="membership-status-pill gold">{{ activationStateText }}</span>
        <span class="membership-status-pill soft">{{ vipStatusText }}</span>
      </div>

      <div class="membership-status-grid">
        <article class="membership-status-item">
          <span>实名状态</span>
          <strong>{{ kycStatusText }}</strong>
          <p>{{ todoTitle }}</p>
        </article>
        <article class="membership-status-item">
          <span>剩余天数</span>
          <strong>{{ remainingDaysText }}</strong>
          <p>{{ vipExpireText }}</p>
        </article>
        <article class="membership-status-item">
          <span>当前通道</span>
          <strong>{{ payChannelLabel }}</strong>
          <p>支付动作会按当前通道发起</p>
        </article>
      </div>
    </section>

    <section v-if="selectedPlan" class="membership-focus-card">
      <div class="membership-section-head compact">
        <div>
          <strong>本次推荐</strong>
          <span>先看主推卡，再决定是否切换到其他套餐</span>
        </div>
        <span class="membership-focus-badge" :class="selectedPlan.current ? 'current' : selectedPlan.recommended ? 'recommended' : 'default'">
          {{ selectedPlan.badge }}
        </span>
      </div>

      <article class="membership-focus-main">
        <div class="membership-focus-copy">
          <span class="membership-plan-kicker">{{ selectedPlan.levelLabel }} · {{ selectedPlan.durationDays }} 天</span>
          <h2>{{ selectedPlan.displayName }}</h2>
          <p>{{ selectedPlan.sceneLabel }}</p>
        </div>

        <div class="membership-focus-price">
          <strong>{{ formatMoney(selectedPlan.price) }}</strong>
          <span>{{ selectedPlan.dailyPriceText || '会员深读权益按周期解锁' }}</span>
        </div>

        <div v-if="selectedPlan.highlights.length" class="membership-tag-row">
          <span v-for="item in selectedPlan.highlights" :key="item" class="membership-chip">{{ item }}</span>
        </div>

        <div class="membership-focus-actions">
          <button type="button" class="h5-btn block" :disabled="focusActionDisabled" @click="handleFocusAction">
            {{ focusActionLabel }}
          </button>
          <button type="button" class="h5-btn-secondary block" @click="router.push('/profile')">查看我的账户</button>
        </div>
      </article>
    </section>

    <section class="membership-benefit-card">
      <div class="membership-section-head compact">
        <div>
          <strong>升级后能得到什么</strong>
          <span>像中国金融 App 一样，只保留 3 个最重要的购买理由</span>
        </div>
      </div>

      <div class="membership-benefit-list">
        <article v-for="item in valueCards" :key="item.title" class="membership-benefit-item">
          <span class="membership-kicker">{{ item.title }}</span>
          <p>{{ item.desc }}</p>
          <div class="membership-tag-row compact">
            <span v-for="tag in item.tags" :key="tag" class="membership-chip subtle">{{ tag }}</span>
          </div>
        </article>
      </div>
    </section>

    <section class="membership-plan-card">
      <div class="membership-section-head compact">
        <div>
          <strong>套餐选择</strong>
          <span>点击卡片切换当前方案，主动作区始终跟随已选套餐</span>
        </div>
      </div>

      <div v-if="plans.length" class="membership-plan-list">
        <button
          v-for="plan in plans"
          :key="plan.id"
          type="button"
          class="membership-plan-item"
          :class="{ active: selectedPlan?.id === plan.id, current: plan.current, recommended: plan.recommended }"
          @click="selectedPlanID = plan.id"
        >
          <div class="membership-plan-row">
            <div class="membership-plan-copy">
              <div class="membership-plan-topline">
                <span>{{ plan.levelLabel }} · {{ plan.durationDays }} 天</span>
                <em>{{ plan.badge }}</em>
              </div>
              <strong>{{ plan.displayName }}</strong>
              <p>{{ plan.sceneLabel }}</p>
            </div>
            <div class="membership-plan-price">
              <strong>{{ formatMoney(plan.price) }}</strong>
              <small>{{ plan.dailyPriceText }}</small>
            </div>
          </div>

          <div v-if="plan.features.length" class="membership-tag-row compact">
            <span v-for="feature in plan.features" :key="feature" class="membership-chip subtle">{{ feature }}</span>
          </div>
        </button>
      </div>

      <H5EmptyState v-else title="套餐待同步" description="刷新后会拉取会员产品，展示当前可购买方案。" />
    </section>

    <section class="membership-action-card">
      <div class="membership-section-head compact">
        <div>
          <strong>支付与激活</strong>
          <span>支付通道、待办说明与主 CTA 放到一个操作区里完成</span>
        </div>
      </div>

      <div class="membership-mini-grid">
        <article class="membership-mini-item">
          <span>当前待办</span>
          <strong>{{ todoTitle }}</strong>
          <p>{{ todoDescription }}</p>
        </article>
        <article class="membership-mini-item">
          <span>已选方案</span>
          <strong>{{ selectedPlan?.displayName || '待选择' }}</strong>
          <p>{{ selectedPlan?.dailyPriceText || '先选择适合你的会员周期' }}</p>
        </article>
      </div>

      <div class="membership-channel-strip">
        <button
          v-for="item in payChannels"
          :key="item.value"
          type="button"
          class="membership-channel-pill"
          :class="{ active: payChannel === item.value }"
          @click="payChannel = item.value"
        >
          {{ item.label }}
        </button>
      </div>

      <div class="membership-action-buttons">
        <button type="button" class="h5-btn block" :disabled="focusActionDisabled" @click="handleFocusAction">{{ focusActionLabel }}</button>
        <button v-if="latestPaymentURL" type="button" class="h5-btn-secondary block" @click="openLatestPaymentPage">继续支付</button>
      </div>

      <p v-if="actionMessage" class="membership-inline-note">{{ actionMessage }}</p>
      <p v-if="loadError" class="membership-inline-note">{{ loadError }}</p>
    </section>

    <section class="membership-orders-card">
      <div class="membership-section-head compact">
        <div>
          <strong>最近订单</strong>
          <span>订单记录下沉展示，只保留支付结果和方案信息</span>
        </div>
      </div>

      <div v-if="orderCards.length" class="membership-order-list">
        <article v-for="item in orderCards" :key="item.id" class="membership-order-item">
          <div class="membership-order-topline">
            <div>
              <span>{{ item.time }}</span>
              <strong>{{ item.title }}</strong>
            </div>
            <em :class="item.tone">{{ item.status }}</em>
          </div>
          <p>{{ item.desc }}</p>
        </article>
      </div>
      <H5EmptyState v-else title="暂无订单" description="完成下单后，最近订单会显示在这里。" />
    </section>

    <H5StickyCta
      :title="stickyTitle"
      :description="stickyDescription"
      :primary-label="stickyPrimaryLabel"
      secondary-label="查看我的账户"
      @primary="handleStickyPrimary"
      @secondary="router.push('/profile')"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import H5EmptyState from "../components/H5EmptyState.vue";
import H5StickyCta from "../components/H5StickyCta.vue";
import { createMembershipOrder, getMembershipQuota, listMembershipOrders, listMembershipProducts } from "../../../api/membership";
import { shouldUseDemoFallback } from "../../../lib/fallback-policy";
import {
  formatDateTime,
  formatMoney,
  mapActivationState,
  mapKYCStatus,
  mapMemberLevel,
  mapPayChannel,
  mapPaymentStatus,
  mapVIPStatus,
  resolveVipStage,
  toArray
} from "../lib/formatters";
import { buildMembershipCashierModel } from "../lib/membership-cashier.js";
import { fallbackMembershipOrders, fallbackMembershipProducts, fallbackQuota } from "../lib/mock-data";

const router = useRouter();
const useDemoFallback = shouldUseDemoFallback();
const PAYMENT_PROTOCOL_ALLOWLIST = new Set(["http:", "https:", "alipay:", "alipays:", "weixin:", "wxp:", "unionpay:", "uppay:", "yolkpay:"]);

const loading = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const payChannel = ref("YOLKPAY");
const rawProducts = ref(useDemoFallback ? [...fallbackMembershipProducts] : []);
const rawQuota = ref(useDemoFallback ? { ...fallbackQuota } : {});
const rawOrders = ref(useDemoFallback ? [...fallbackMembershipOrders] : []);
const orderingProductID = ref("");
const actionMessage = ref("");
const latestPaymentAction = ref(null);
const selectedPlanID = ref("");

const payChannels = [
  { value: "YOLKPAY", label: "蛋黄支付" },
  { value: "ALIPAY", label: "支付宝" },
  { value: "WECHAT", label: "微信支付" },
  { value: "CARD", label: "银行卡" }
];

const currentLevelText = computed(() => mapMemberLevel(rawQuota.value?.member_level, rawQuota.value?.member_level));
const vipStatusText = computed(() => mapVIPStatus(rawQuota.value?.vip_status, rawQuota.value?.member_level));
const kycStatusText = computed(() => mapKYCStatus(rawQuota.value?.kyc_status));
const activationStateText = computed(() => mapActivationState(rawQuota.value?.activation_state));
const remainingDaysText = computed(() => `${Math.max(0, Number(rawQuota.value?.vip_remaining_days || 0))} 天`);
const vipExpireText = computed(() => rawQuota.value?.vip_expire_at ? `到期 ${formatDateTime(rawQuota.value.vip_expire_at)}` : "暂未开通");
const payChannelLabel = computed(() => mapPayChannel(payChannel.value));
const latestPaymentURL = computed(() => resolvePaymentActionURL(latestPaymentAction.value));

const heroTitle = computed(() => {
  if (rawQuota.value?.activation_state === "PAID_PENDING_KYC") {
    return `${currentLevelText.value} 已支付，待实名激活`;
  }
  return resolveVipStage(rawQuota.value) ? `${currentLevelText.value} 已开通，可继续升级` : "选定套餐后即可发起支付";
});
const heroDescription = computed(() => {
  if (rawQuota.value?.activation_state === "PAID_PENDING_KYC") {
    return "当前最优先的动作是完成实名，审核通过后高级权益会自动生效。";
  }
  if (resolveVipStage(rawQuota.value)) {
    return "当前权益已经生效，可以继续管理账户状态，也可以切到更高阶方案。";
  }
  return "像中国 App 收银台一样，先看推荐方案，再在一个动作区里完成支付。";
});

const valueCards = computed(() => [
  {
    title: "观点与策略深读",
    desc: "主推荐、策略理由、风险边界和版本信息会在移动端完整展开，不再只看结论。",
    tags: ["主推荐", "风险边界", "版本追踪"]
  },
  {
    title: "资讯正文与附件",
    desc: "VIP 资讯会解锁正文与附件，阅读节奏更完整，不需要频繁切页面找内容。",
    tags: ["正文解锁", "附件下载", "深度阅读"]
  },
  {
    title: "账户动作与复盘",
    desc: "会员状态、实名激活、订单记录和账户待办会被串成连续使用节奏。",
    tags: ["账户待办", "支付记录", "使用节奏"]
  }
]);

const cashierModel = computed(() => buildMembershipCashierModel({
  products: rawProducts.value,
  quota: rawQuota.value,
  mapMemberLevel: (value) => mapMemberLevel(value, value),
  resolveVipStage
}));

const plans = computed(() => cashierModel.value.plans);
const spotlightPlan = computed(() => cashierModel.value.spotlightPlan);
const selectedPlan = computed(() => plans.value.find((item) => item.id === selectedPlanID.value) || spotlightPlan.value || plans.value.find((item) => !item.disabled) || null);
const productNameMap = computed(() => Object.fromEntries(plans.value.map((item) => [item.id, item.displayName])));

const orderCards = computed(() => rawOrders.value.slice(0, 4).map((item) => ({
  id: item.id || item.order_no,
  title: productNameMap.value[item.product_id] || `${item.order_no || item.id || "-"}`,
  time: formatDateTime(item.paid_at || item.created_at),
  status: mapPaymentStatus(item.status),
  tone: String(item.status || "").toUpperCase() === "PAID" ? "success" : "gold",
  desc: `${mapPayChannel(item.pay_channel)} · ${formatMoney(item.amount)}`
})));

const todoTitle = computed(() => {
  if (rawQuota.value?.activation_state === "PAID_PENDING_KYC") {
    return "待实名激活";
  }
  if (!resolveVipStage(rawQuota.value)) {
    return "待支付开通";
  }
  return "权益生效中";
});
const todoDescription = computed(() => {
  if (rawQuota.value?.activation_state === "PAID_PENDING_KYC") {
    return "前往我的页提交实名信息，审核通过后自动激活。";
  }
  if (!resolveVipStage(rawQuota.value)) {
    return selectedPlan.value ? `当前建议先支付 ${selectedPlan.value.displayName}。` : "优先选择合适套餐并完成支付。";
  }
  return "当前账户已开通高级权益，可继续查看订单或选择更高阶方案。";
});

const focusActionLabel = computed(() => {
  if (rawQuota.value?.activation_state === "PAID_PENDING_KYC") {
    return "去完成实名";
  }
  if (resolveVipStage(rawQuota.value) && selectedPlan.value?.current) {
    return "查看我的账户";
  }
  return selectedPlan.value ? `开通 ${selectedPlan.value.displayName}` : "立即开通";
});
const focusActionDisabled = computed(() => Boolean(selectedPlan.value?.id && orderingProductID.value === selectedPlan.value.id));

const stickyTitle = computed(() => {
  if (rawQuota.value?.activation_state === "PAID_PENDING_KYC") {
    return "支付已完成，下一步先完成实名激活";
  }
  if (resolveVipStage(rawQuota.value)) {
    return selectedPlan.value?.current ? "当前方案已生效，可继续管理账户状态" : "权益已生效，仍可切换更高阶方案";
  }
  return "选定套餐后即可直接发起支付动作";
});
const stickyDescription = computed(() => actionMessage.value || (selectedPlan.value?.dailyPriceText ? `${selectedPlan.value.displayName} · ${selectedPlan.value.dailyPriceText}` : "套餐、支付通道和订单动作都集中在当前页面。"));
const stickyPrimaryLabel = computed(() => focusActionLabel.value);

async function loadMembershipPage() {
  loading.value = true;
  loadError.value = "";
  const [productsResult, quotaResult, ordersResult] = await Promise.allSettled([
    listMembershipProducts({ status: "ACTIVE", page: 1, page_size: 20 }),
    getMembershipQuota(),
    listMembershipOrders({ page: 1, page_size: 20 })
  ]);

  const errors = [];
  if (productsResult.status === "fulfilled") {
    rawProducts.value = toArray(productsResult.value?.items, []);
  } else {
    errors.push(`会员产品加载失败：${productsResult.reason?.message || "unknown error"}`);
  }
  if (quotaResult.status === "fulfilled") {
    rawQuota.value = quotaResult.value || {};
  } else {
    errors.push(`会员状态加载失败：${quotaResult.reason?.message || "unknown error"}`);
  }
  if (ordersResult.status === "fulfilled") {
    rawOrders.value = toArray(ordersResult.value?.items, []);
  } else {
    errors.push(`订单加载失败：${ordersResult.reason?.message || "unknown error"}`);
  }

  if (!rawProducts.value.length && useDemoFallback) {
    rawProducts.value = [...fallbackMembershipProducts];
  }
  if (!Object.keys(rawQuota.value || {}).length && useDemoFallback) {
    rawQuota.value = { ...fallbackQuota };
  }
  if (!rawOrders.value.length && useDemoFallback) {
    rawOrders.value = [...fallbackMembershipOrders];
  }

  if (!plans.value.find((item) => item.id === selectedPlanID.value)) {
    selectedPlanID.value = spotlightPlan.value?.id || plans.value[0]?.id || "";
  }

  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  loadError.value = errors.join("；");
  loading.value = false;
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
  try {
    const parsed = new URL(rawTarget, typeof window !== "undefined" ? window.location.origin : "http://localhost");
    return PAYMENT_PROTOCOL_ALLOWLIST.has(parsed.protocol) ? parsed.toString() : "";
  } catch {
    return "";
  }
}

function openPaymentLink(action) {
  const url = resolvePaymentActionURL(action);
  if (!url || typeof window === "undefined") {
    return false;
  }
  window.open(url, "_blank", "noopener,noreferrer");
  return true;
}

async function handleCreateOrder(plan) {
  if (!plan?.id || orderingProductID.value) {
    return;
  }
  orderingProductID.value = plan.id;
  actionMessage.value = "";
  try {
    const result = await createMembershipOrder({
      product_id: plan.id,
      pay_channel: payChannel.value
    });
    latestPaymentAction.value = result?.payment_action || null;
    const orderNo = result?.order?.order_no || result?.order_no || result?.id || "-";
    const opened = openPaymentLink(result?.payment_action);
    actionMessage.value = opened
      ? `${plan.displayName} 已下单，订单 ${orderNo} 的支付页已打开。`
      : `${plan.displayName} 已下单，订单 ${orderNo} 可在下方继续支付。`;
    await loadMembershipPage();
  } catch (error) {
    actionMessage.value = error?.message || "创建订单失败";
  } finally {
    orderingProductID.value = "";
  }
}

function openLatestPaymentPage() {
  if (!openPaymentLink(latestPaymentAction.value)) {
    actionMessage.value = "当前订单未返回可用的支付链接";
  }
}

function handleFocusAction() {
  if (rawQuota.value?.activation_state === "PAID_PENDING_KYC") {
    router.push("/profile");
    return;
  }
  if (resolveVipStage(rawQuota.value) && selectedPlan.value?.current) {
    router.push("/profile");
    return;
  }
  if (selectedPlan.value) {
    handleCreateOrder(selectedPlan.value);
  }
}

function handleStickyPrimary() {
  handleFocusAction();
}

onMounted(() => {
  loadMembershipPage();
});
</script>

<style scoped>
.membership-page {
  gap: 12px;
}

.membership-hero,
.membership-focus-card,
.membership-benefit-card,
.membership-plan-card,
.membership-action-card,
.membership-orders-card {
  border: 1px solid var(--h5-line);
  border-radius: var(--h5-radius);
  box-shadow: var(--h5-shadow);
  padding: 20px 18px;
  display: grid;
  gap: 16px;
  background: var(--h5-panel-bg);
}

.membership-hero {
  background:
    radial-gradient(circle at top right, rgba(219, 183, 101, 0.22), transparent 28%),
    linear-gradient(160deg, #173a6e 0%, #21497d 52%, #f6f9fc 52%, #ffffff 100%);
}

.membership-hero-topline,
.membership-plan-topline,
.membership-order-topline,
.membership-plan-row,
.membership-section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.membership-hero-topline {
  align-items: center;
}

.membership-refresh {
  min-height: 34px;
  padding: 0 12px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.22);
  background: rgba(255, 255, 255, 0.14);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}

.membership-hero-topline span,
.membership-hero-copy p,
.membership-status-item p,
.membership-status-item span,
.membership-inline-note,
.membership-plan-copy p,
.membership-benefit-item p,
.membership-mini-item p,
.membership-order-item p,
.membership-section-head span,
.membership-plan-price small {
  margin: 0;
}

.membership-hero-topline span {
  color: rgba(255, 255, 255, 0.78);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.membership-hero-copy,
.membership-focus-main,
.membership-benefit-list,
.membership-plan-list,
.membership-order-list,
.membership-action-buttons {
  display: grid;
  gap: 10px;
}

.membership-hero-copy h1,
.membership-focus-copy h2 {
  margin: 0;
  color: #16263d;
  font-size: clamp(28px, 7vw, 34px);
  line-height: 1.14;
  letter-spacing: -0.03em;
}

.membership-hero-copy p {
  color: #5f7088;
  font-size: 13px;
  line-height: 1.72;
}

.membership-status-strip,
.membership-tag-row,
.membership-channel-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.membership-status-pill {
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  font-weight: 700;
}

.membership-status-pill.brand {
  color: #173a6e;
  background: rgba(255, 255, 255, 0.9);
}

.membership-status-pill.gold {
  color: #9b6614;
  background: rgba(255, 230, 178, 0.92);
}

.membership-status-pill.soft {
  color: #4d5e76;
  background: rgba(255, 255, 255, 0.7);
}

.membership-status-grid,
.membership-mini-grid {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.membership-mini-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.membership-status-item,
.membership-mini-item,
.membership-benefit-item,
.membership-order-item,
.membership-plan-item {
  border-radius: var(--h5-radius-sm);
}

.membership-status-item {
  padding: 14px;
  background: rgba(255, 255, 255, 0.9);
  display: grid;
  gap: 6px;
}

.membership-status-item span,
.membership-mini-item span,
.membership-order-topline span,
.membership-plan-kicker,
.membership-kicker {
  color: #75849a;
  font-size: 11px;
}

.membership-status-item strong,
.membership-mini-item strong,
.membership-section-head strong,
.membership-plan-copy strong,
.membership-order-topline strong,
.membership-plan-price strong,
.membership-focus-price strong {
  margin: 0;
  color: #16263d;
}

.membership-status-item strong,
.membership-mini-item strong {
  font-size: 16px;
  line-height: 1.3;
}

.membership-focus-card {
  background:
    radial-gradient(circle at top right, rgba(219, 183, 101, 0.16), transparent 30%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.99), rgba(246, 249, 253, 0.97));
}

.membership-section-head.compact {
  align-items: center;
}

.membership-section-head > div {
  display: grid;
  gap: 6px;
}

.membership-section-head strong {
  font-size: 18px;
  line-height: 1.35;
}

.membership-section-head span,
.membership-inline-note,
.membership-order-item p,
.membership-mini-item p,
.membership-benefit-item p,
.membership-plan-copy p {
  color: #6b7b90;
  font-size: 12px;
  line-height: 1.7;
}

.membership-focus-badge {
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  font-weight: 700;
}

.membership-focus-badge.current {
  color: #2d7a61;
  background: rgba(45, 122, 97, 0.14);
}

.membership-focus-badge.recommended {
  color: #9b6614;
  background: rgba(184, 137, 61, 0.14);
}

.membership-focus-badge.default {
  color: #173a6e;
  background: rgba(24, 58, 110, 0.08);
}

.membership-focus-main {
  padding: 16px;
  border-radius: var(--h5-radius);
  background: linear-gradient(180deg, rgba(23, 58, 110, 0.97), rgba(35, 76, 128, 0.94));
}

.membership-focus-copy,
.membership-focus-price,
.membership-plan-copy {
  display: grid;
  gap: 6px;
}

.membership-focus-copy h2,
.membership-focus-copy p,
.membership-focus-price strong,
.membership-focus-price span,
.membership-plan-topline span,
.membership-plan-topline em {
  color: #fff;
}

.membership-focus-copy p,
.membership-focus-price span {
  font-size: 13px;
  line-height: 1.68;
  color: rgba(255, 255, 255, 0.82);
}

.membership-focus-price strong {
  font-size: 30px;
  line-height: 1.08;
}

.membership-plan-kicker {
  color: rgba(255, 255, 255, 0.7);
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.membership-focus-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.membership-benefit-item,
.membership-mini-item,
.membership-order-item {
  padding: 14px;
  background: rgba(20, 52, 95, 0.04);
  display: grid;
  gap: 8px;
}

.membership-plan-item {
  width: 100%;
  padding: 14px;
  border: 1px solid rgba(16, 42, 86, 0.08);
  background: rgba(249, 251, 253, 0.98);
  display: grid;
  gap: 10px;
  text-align: left;
}

.membership-plan-item.active {
  border-color: rgba(24, 58, 110, 0.24);
  background: linear-gradient(180deg, rgba(245, 248, 252, 1), rgba(255, 255, 255, 1));
  box-shadow: 0 12px 26px rgba(16, 42, 86, 0.08);
}

.membership-plan-item.current {
  border-color: rgba(45, 122, 97, 0.18);
}

.membership-plan-item.recommended {
  background: linear-gradient(180deg, rgba(255, 250, 243, 0.96), rgba(255, 255, 255, 1));
}

.membership-plan-copy strong {
  font-size: 16px;
  line-height: 1.4;
}

.membership-plan-copy p,
.membership-plan-price small,
.membership-plan-topline span,
.membership-plan-topline em {
  color: #68788f;
}

.membership-plan-price {
  display: grid;
  justify-items: end;
  gap: 4px;
  flex: 0 0 auto;
}

.membership-plan-price strong {
  font-size: 22px;
  line-height: 1.1;
}

.membership-plan-topline {
  align-items: center;
}

.membership-plan-topline em,
.membership-order-topline em {
  font-style: normal;
  font-size: 12px;
  font-weight: 700;
}

.membership-chip {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  font-size: 11px;
  font-weight: 700;
}

.membership-chip.subtle {
  background: rgba(20, 52, 95, 0.06);
  color: #51627b;
}

.membership-channel-pill {
  min-height: 42px;
  padding: 0 16px;
  border: 1px solid rgba(16, 42, 86, 0.08);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.96);
  color: #5c6d84;
  font-size: 13px;
  font-weight: 700;
}

.membership-channel-pill.active {
  color: #173a6e;
  border-color: rgba(24, 58, 110, 0.2);
  background: rgba(24, 58, 110, 0.08);
}

.membership-order-topline em.success {
  color: #2d7a61;
}

.membership-order-topline em.gold {
  color: #9b6614;
}

@media (max-width: 380px) {
  .membership-status-grid,
  .membership-mini-grid,
  .membership-focus-actions {
    grid-template-columns: 1fr;
  }

  .membership-plan-row {
    flex-direction: column;
  }

  .membership-plan-price {
    justify-items: start;
  }
}
</style>
