<template>
  <section class="membership-page fade-up">
    <header class="member-hero card">
      <div>
        <p class="hero-kicker">会员中心 · API 联动</p>
        <h1 class="section-title">会员产品、配额和订单已接入后端接口。</h1>
        <p class="section-subtitle">支持查看套餐能力、订单记录，并可直接创建会员订单。</p>
      </div>
      <div class="hero-pill">
        <p>当前套餐</p>
        <strong>{{ currentPlanName }}</strong>
        <small>{{ currentLevelLabel }}</small>
      </div>
    </header>

    <div class="api-tip">
      <p v-if="loading">正在同步会员数据...</p>
      <p v-else-if="loadError">{{ loadError }}</p>
      <p v-else>数据更新时间：{{ lastUpdatedAt || "-" }}</p>
    </div>

    <article class="summary card">
      <header>
        <h2 class="section-title">我的会员配额</h2>
        <p class="section-subtitle">来自 membership/quota 接口</p>
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
        <button type="button" @click="loadMembershipData">刷新会员数据</button>
      </div>
      <p v-if="actionMessage" class="action-message">{{ actionMessage }}</p>
      <div v-if="latestPaymentAction" class="action-payment-entry">
        <button type="button" @click="openLatestPaymentPage">
          立即前往{{ mapPayChannel(latestPaymentAction.channel) }}收银台
        </button>
        <a
          v-if="latestPaymentAction.pay_url || latestPaymentAction.qrcode || latestPaymentAction.urlscheme"
          :href="latestPaymentAction.pay_url || latestPaymentAction.qrcode || latestPaymentAction.urlscheme"
          target="_blank"
          rel="noopener"
        >
          打开支付链接
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
          <span class="badge" :class="plan.statusClass">{{ plan.statusLabel }}</span>
          <span>{{ plan.durationText }}</span>
        </p>
        <ul>
          <li v-for="feature in plan.features" :key="feature">{{ feature }}</li>
        </ul>
        <button
          :disabled="plan.disabled || orderingProductID === plan.id"
          @click="handleCreateOrder(plan)"
        >
          {{ orderingProductID === plan.id ? "处理中..." : plan.actionText }}
        </button>
      </article>
    </div>

    <article class="ability card">
      <header>
        <h2 class="section-title">能力对比</h2>
        <p class="section-subtitle">按套餐状态、时长、等级和成本展示（来自会员产品 API）。</p>
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
      <div v-else class="empty-box">暂无可对比的会员产品</div>
    </article>

    <article class="orders card">
      <header>
        <h2 class="section-title">订单记录</h2>
        <p class="section-subtitle">来自 membership/orders 接口</p>
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
        <button type="button" :disabled="manualRefreshing" @click="refreshOrdersManually">
          {{ manualRefreshing ? "刷新中..." : "手动刷新订单" }}
        </button>
      </div>
      <p class="order-tip">
        当前待支付 {{ pendingOrderCount }} 笔，请在收银台完成真实支付
      </p>

      <div class="order-table-wrap" v-if="orderRows.length > 0">
        <table class="order-table">
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
                <span class="badge" :class="item.statusClass">{{ item.statusLabel }}</span>
              </td>
              <td>{{ item.time }}</td>
              <td class="order-actions">
                <span v-if="item.isPending">待支付</span>
                <span v-else>-</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-else class="empty-box">暂无会员订单记录</div>

      <div class="order-mobile-list">
        <article v-for="item in orderRows" :key="`mobile-${item.orderNo}`">
          <div class="top-line">
            <p>{{ item.productName }}</p>
            <span>{{ item.amount }}</span>
          </div>
          <div class="meta-line">
            <span>{{ item.payChannel }}</span>
            <span>{{ item.time }}</span>
            <span class="badge" :class="item.statusClass">{{ item.statusLabel }}</span>
          </div>
          <p class="order-no">订单号：{{ item.orderNo }}</p>
        </article>
      </div>
    </article>
  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import {
  createMembershipOrder,
  getMembershipQuota,
  listMembershipOrders,
  listMembershipProducts
} from "../api/membership";
import { shouldUseDemoFallback } from "../lib/fallback-policy";

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

const payChannelOptions = [
  { value: "ALIPAY", label: "支付宝" },
  { value: "WECHAT", label: "微信支付" },
  { value: "CARD", label: "银行卡" },
  { value: "YOLKPAY", label: "蛋黄支付" }
];
const useDemoFallback = shouldUseDemoFallback();

const loading = ref(false);
const manualRefreshing = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const orderingProductID = ref("");
const actionMessage = ref("");
const latestPaymentAction = ref(null);
const payChannel = ref("YOLKPAY");
const autoTrackPendingOrders = ref(true);
let pendingOrderTimer = null;

const rawProducts = ref(useDemoFallback ? [...fallbackProducts] : []);
const rawQuota = ref(useDemoFallback ? { ...fallbackQuota } : {});
const rawOrders = ref(useDemoFallback ? [...fallbackOrders] : []);

const currentMemberLevel = computed(() => String(rawQuota.value?.member_level || "FREE").toUpperCase());
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
            ? "已有待支付订单"
            : String(item.status || "").toUpperCase() === "ACTIVE"
              ? "开通会员"
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
const vipExpireText = computed(() => formatDateTime(rawQuota.value?.vip_expire_at));
const vipRemainDaysText = computed(() => `${Math.max(0, Number(rawQuota.value?.vip_remaining_days || 0))} 天`);

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
    errors.push(`会员产品接口失败：${productsResult.reason?.message || "unknown error"}`);
  }

  if (quotaResult.status === "fulfilled" && quotaResult.value) {
    rawQuota.value = {
      ...rawQuota.value,
      ...quotaResult.value
    };
  } else if (quotaResult.status === "rejected") {
    errors.push(`会员配额接口失败：${quotaResult.reason?.message || "unknown error"}`);
  }

  if (ordersResult.status === "fulfilled" && Array.isArray(ordersResult.value?.items)) {
    rawOrders.value = ordersResult.value.items;
  } else if (ordersResult.status === "rejected") {
    errors.push(`会员订单接口失败：${ordersResult.reason?.message || "unknown error"}`);
  }

  if (errors.length > 0) {
    loadError.value = errors.join("；");
  }
  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  if (!silent) {
    loading.value = false;
  }
}

async function handleCreateOrder(plan) {
  if (!plan?.id || plan.disabled || orderingProductID.value) {
    return;
  }

  orderingProductID.value = plan.id;
  actionMessage.value = "";
  latestPaymentAction.value = null;

  try {
    const result = await createMembershipOrder({
      product_id: plan.id,
      pay_channel: payChannel.value
    });
    const order = result?.order || result || {};
    const orderNo = order?.order_no || order?.id || "-";
    const initialized = result?.payment_initialized;
    const action = result?.payment_action || null;
    latestPaymentAction.value = action;
    if (initialized === false) {
      actionMessage.value = `订单 ${orderNo} 已创建，但支付下单失败：${result?.payment_error || "请稍后重试"}`;
    } else if (action) {
      const opened = openPaymentLink(action);
      actionMessage.value = opened
        ? `订单 ${orderNo} 已创建，已为你打开支付页，请完成支付后返回刷新状态`
        : `订单 ${orderNo} 已创建，请点击“立即前往收银台”完成支付`;
    } else {
      actionMessage.value = `订单 ${orderNo} 已创建，当前状态：${mapOrderStatus(order?.status)}`;
    }
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
  if (!openPaymentLink(latestPaymentAction.value)) {
    actionMessage.value = "当前订单未返回可打开的支付链接";
  }
}

function resolvePaymentActionURL(action) {
  if (!action) {
    return "";
  }
  return String(action.pay_url || action.qrcode || action.urlscheme || "").trim();
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
  await loadMembershipData();
  syncPendingOrderTracking();
});

onBeforeUnmount(() => {
  stopPendingOrderTracking();
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

.hero-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.hero-pill {
  border-radius: 12px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(246, 244, 239, 0.9);
  padding: 8px 11px;
}

.hero-pill p {
  margin: 0;
  font-size: 11px;
  color: var(--color-text-sub);
}

.hero-pill strong {
  margin-top: 2px;
  display: block;
  font-size: 18px;
  color: var(--color-pine-700);
}

.hero-pill small {
  display: block;
  margin-top: 3px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.api-tip {
  border-radius: 10px;
  border: 1px dashed rgba(216, 223, 216, 0.95);
  background: rgba(246, 244, 239, 0.76);
  padding: 8px 10px;
}

.api-tip p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.summary {
  padding: 14px;
}

.summary-grid {
  margin-top: 10px;
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.summary-grid article {
  border: 1px solid rgba(216, 223, 216, 0.9);
  border-radius: 11px;
  background: rgba(255, 255, 255, 0.9);
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
  border: 1px solid rgba(216, 223, 216, 0.95);
  background: #fff;
  padding: 5px 9px;
}

.summary-actions button {
  border: 0;
  border-radius: 10px;
  padding: 8px 12px;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
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
  border: 0;
  border-radius: 10px;
  padding: 7px 12px;
  color: #fff;
  font-weight: 600;
  cursor: pointer;
  background: linear-gradient(145deg, #2f6d60, #3f7f71);
}

.action-payment-entry a {
  color: #2f6d60;
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
  border-color: rgba(63, 127, 113, 0.45);
  background: linear-gradient(160deg, rgba(223, 236, 230, 0.72), rgba(255, 255, 255, 0.92));
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
  border: 0;
  border-radius: 10px;
  padding: 10px 12px;
  cursor: pointer;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.plan.featured button {
  background: linear-gradient(145deg, #d2a766, #b57e38);
}

.plan button:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.badge {
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
}

.badge.success {
  color: #1f6a40;
  background: rgba(187, 230, 204, 0.6);
}

.badge.warning {
  color: #8b5f1d;
  background: rgba(244, 223, 183, 0.72);
}

.badge.danger {
  color: #93332a;
  background: rgba(237, 198, 190, 0.72);
}

.badge.muted {
  color: #52616d;
  background: rgba(224, 228, 231, 0.72);
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
  border: 1px solid rgba(216, 223, 216, 0.9);
  padding: 8px 10px;
  display: grid;
  gap: 8px;
}

.ability-head {
  background: rgba(246, 244, 239, 0.84);
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
  border: 1px dashed rgba(216, 223, 216, 0.95);
  background: rgba(246, 244, 239, 0.72);
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
  border: 1px solid rgba(216, 223, 216, 0.9);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.9);
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
  border: 0;
  border-radius: 9px;
  padding: 7px 10px;
  cursor: pointer;
  color: #fff;
  font-weight: 600;
  font-size: 12px;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.order-toolbar button:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.order-tip {
  margin: 7px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.order-table-wrap {
  margin-top: 10px;
  overflow: auto;
}

.order-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 720px;
}

.order-table th,
.order-table td {
  border-bottom: 1px solid rgba(216, 223, 216, 0.9);
  padding: 9px 8px;
  text-align: left;
  font-size: 13px;
}

.order-table th {
  color: var(--color-text-sub);
  background: rgba(246, 244, 239, 0.8);
  font-size: 12px;
}

.order-actions {
  min-width: 110px;
}

.mini-btn {
  border: 0;
  border-radius: 8px;
  padding: 6px 9px;
  cursor: pointer;
  color: #fff;
  font-weight: 600;
  font-size: 12px;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.mini-btn:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.order-mobile-list {
  display: none;
  margin-top: 10px;
  gap: 8px;
}

.order-mobile-list article {
  border: 1px solid rgba(216, 223, 216, 0.9);
  border-radius: 11px;
  padding: 9px;
  background: rgba(255, 255, 255, 0.9);
}

.top-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.top-line p {
  margin: 0;
  font-weight: 600;
}

.top-line span {
  color: var(--color-pine-700);
  font-weight: 600;
}

.meta-line {
  margin-top: 6px;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  color: var(--color-text-sub);
  font-size: 12px;
}

.order-no {
  margin: 7px 0 0;
  color: var(--color-text-sub);
  font-size: 12px;
}

.order-mobile-actions {
  margin-top: 8px;
}

@media (max-width: 980px) {
  .member-hero,
  .plans,
  .summary-grid,
  .order-summary-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .hero-pill {
    width: fit-content;
  }

  .plan button {
    width: 100%;
  }

  .ability-head {
    display: none;
  }

  .ability-head,
  .ability-row {
    grid-template-columns: 1fr 1fr;
  }

  .ability-row span:first-child {
    grid-column: 1 / -1;
    font-weight: 600;
    color: var(--color-text-main);
  }

  .order-table-wrap {
    display: none;
  }

  .order-mobile-list {
    display: grid;
  }

  .order-mobile-actions .mini-btn {
    width: 100%;
  }
}
</style>
