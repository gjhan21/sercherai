<template>
  <section class="profile-page fade-up">
    <header class="account-card card">
      <div class="identity">
        <div class="avatar">U</div>
        <div>
          <h1>{{ displayProfile.name }}</h1>
          <p>
            {{ vipInfo.level }} · KYC {{ displayProfile.kycStatus }} · 最近更新
            {{ lastUpdatedAt || "-" }}
          </p>
        </div>
      </div>
      <div class="actions">
        <button class="primary" type="button" @click="loadUserCenterData">刷新数据</button>
        <button class="ghost" type="button" @click="openSecurityPanel">账户安全</button>
      </div>
    </header>

    <article class="card query-card">
      <header class="query-head">
        <div>
          <h2 class="section-title">客户信息查询中心</h2>
          <p class="section-subtitle">支持查询 VIP、支付、阅读、订阅及其他信息。</p>
        </div>
        <div class="range-switch">
          <button
            v-for="range in timeRanges"
            :key="range"
            type="button"
            :class="{ active: activeRange === range }"
            @click="activeRange = range"
          >
            {{ range }}
          </button>
        </div>
      </header>

      <nav class="query-nav">
        <button
          v-for="item in modules"
          :key="item.key"
          type="button"
          :class="{ active: activeModule === item.key }"
          @click="activeModule = item.key"
        >
          {{ item.label }}
        </button>
      </nav>

      <div class="query-tip">
        <p>当前查询：{{ currentModule.label }}</p>
        <p>时间范围：{{ activeRange }}</p>
      </div>

      <div v-if="loading" class="state-box">正在加载 API 数据...</div>
      <div v-else-if="loadError" class="state-box warning">API 加载失败：{{ loadError }}</div>

      <section class="query-body">
        <template v-if="activeModule === 'vip'">
          <div class="vip-panel">
            <article class="vip-main">
              <p class="vip-level">{{ vipInfo.level }}</p>
              <h3>VIP 有效期至 {{ vipInfo.expireAt }}</h3>
              <p>下次续费时间：{{ vipInfo.nextRenewAt }}，剩余 {{ vipInfo.remainingDays }} 天</p>
            </article>
            <div class="summary-grid">
              <article v-for="item in vipMetrics" :key="item.label">
                <p>{{ item.label }}</p>
                <strong>{{ item.value }}</strong>
              </article>
            </div>
          </div>
          <div class="benefits-grid">
            <article v-for="item in vipBenefits" :key="item.title">
              <h4>{{ item.title }}</h4>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </template>

        <template v-else-if="activeModule === 'payment'">
          <div class="summary-grid">
            <article v-for="item in paymentSummary" :key="item.label">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
            </article>
          </div>

          <div class="payment-table-wrap">
            <table class="payment-table">
              <thead>
                <tr>
                  <th>订单号</th>
                  <th>时间</th>
                  <th>项目</th>
                  <th>金额</th>
                  <th>方式</th>
                  <th>状态</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in paymentRecords" :key="item.orderNo">
                  <td>{{ item.orderNo }}</td>
                  <td>{{ item.time }}</td>
                  <td>{{ item.product }}</td>
                  <td>{{ item.amount }}</td>
                  <td>{{ item.method }}</td>
                  <td>
                    <span class="status" :class="paymentStatusClass(item.status)">
                      {{ item.status }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="payment-mobile">
            <article v-for="item in paymentRecords" :key="`m-${item.orderNo}`">
              <div class="top-line">
                <p>{{ item.product }}</p>
                <span>{{ item.amount }}</span>
              </div>
              <div class="meta-line">
                <span>{{ item.time }}</span>
                <span>{{ item.method }}</span>
                <span class="status" :class="paymentStatusClass(item.status)">{{ item.status }}</span>
              </div>
              <p class="order">订单号：{{ item.orderNo }}</p>
            </article>
          </div>
        </template>

        <template v-else-if="activeModule === 'reading'">
          <div class="summary-grid">
            <article v-for="item in readingStats" :key="item.label">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
            </article>
          </div>

          <div class="log-list">
            <article v-for="item in readingLogs" :key="item.id">
              <div class="top-line">
                <p>{{ item.title }}</p>
                <span>{{ item.type }}</span>
              </div>
              <p class="desc">{{ item.desc }}</p>
              <div class="meta-line">
                <span>{{ item.time }}</span>
                <span>阅读时长 {{ item.duration }}</span>
                <span>完成度 {{ item.progress }}</span>
              </div>
            </article>
          </div>
        </template>

        <template v-else-if="activeModule === 'subscription'">
          <div class="subscription-create">
            <label>
              订阅类型
              <select v-model="newSubscriptionForm.type">
                <option v-for="item in subscriptionTypeOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label>
              推送频率
              <select v-model="newSubscriptionForm.frequency">
                <option v-for="item in subscriptionFrequencyOptions" :key="item.value" :value="item.value">
                  {{ item.label }}
                </option>
              </select>
            </label>
            <label class="scope-input">
              订阅范围
              <input v-model.trim="newSubscriptionForm.scope" placeholder="如：ALL / A股 / 沪深300" />
            </label>
            <button type="button" :disabled="creatingSubscription" @click="handleCreateSubscription">
              {{ creatingSubscription ? "创建中..." : "新增订阅" }}
            </button>
          </div>
          <p v-if="subscriptionActionError" class="state-box warning">{{ subscriptionActionError }}</p>
          <p v-else-if="subscriptionActionMessage" class="state-box">{{ subscriptionActionMessage }}</p>

          <div class="subscription-grid">
            <article v-for="item in subscriptionItems" :key="item.id" class="subscription-item">
              <div class="top-line">
                <p>{{ item.name }}</p>
                <span class="status" :class="subscriptionStatusClass(item.status)">
                  {{ item.status }}
                </span>
              </div>
              <p class="desc">{{ item.desc }}</p>
              <div class="meta-line">
                <span>周期：{{ item.cycle }}</span>
                <span>范围：{{ item.scope }}</span>
                <span>{{ item.price }}</span>
              </div>
              <div class="subscription-actions">
                <button
                  type="button"
                  class="secondary"
                  :disabled="item.saving"
                  @click="handleRotateSubscriptionFrequency(item)"
                >
                  {{ item.saving ? "处理中..." : "切换频率" }}
                </button>
                <button
                  type="button"
                  :disabled="item.saving"
                  @click="handleToggleSubscriptionStatus(item)"
                >
                  {{ item.statusRaw === "ACTIVE" ? "暂停订阅" : "恢复订阅" }}
                </button>
              </div>
            </article>
          </div>
        </template>

        <template v-else-if="activeModule === 'message'">
          <div class="summary-grid">
            <article v-for="item in messageStats" :key="item.label">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
            </article>
          </div>

          <div class="message-list">
            <article v-for="item in messageItems" :key="item.id" class="message-item">
              <div class="top-line">
                <p>{{ item.title }}</p>
                <span class="status" :class="item.readStatusRaw === 'READ' ? 'success' : 'pending'">
                  {{ item.readStatus }}
                </span>
              </div>
              <p class="desc">{{ item.content }}</p>
              <div class="meta-line">
                <span>{{ item.type }}</span>
                <span>{{ item.time }}</span>
              </div>
              <div class="message-actions">
                <button
                  type="button"
                  class="secondary"
                  :disabled="item.readStatusRaw === 'READ' || item.loading"
                  @click="handleReadMessage(item)"
                >
                  {{ item.readStatusRaw === "READ" ? "已读" : item.loading ? "处理中..." : "标记已读" }}
                </button>
              </div>
            </article>
          </div>
        </template>

        <template v-else-if="activeModule === 'invite'">
          <div class="summary-grid">
            <article v-for="item in inviteStats" :key="item.label">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
            </article>
          </div>

          <div class="other-grid">
            <article>
              <h4>我的注册来源</h4>
              <div class="kv-list">
                <p>
                  <span>注册来源</span>
                  <strong>{{ inviteSourceInfo.registrationSource }}</strong>
                </p>
                <p>
                  <span>邀请人ID</span>
                  <strong>{{ inviteSourceInfo.inviterUserID }}</strong>
                </p>
                <p>
                  <span>邀请码</span>
                  <strong>{{ inviteSourceInfo.inviteCode }}</strong>
                </p>
                <p>
                  <span>注册绑定时间</span>
                  <strong>{{ inviteSourceInfo.invitedAt }}</strong>
                </p>
              </div>
            </article>
            <article>
              <h4>我的分享链接</h4>
              <div class="invite-create">
                <label>
                  渠道
                  <select v-model="newShareLinkChannel">
                    <option v-for="item in shareChannelOptions" :key="item.value" :value="item.value">
                      {{ item.label }}
                    </option>
                  </select>
                </label>
                <button type="button" :disabled="creatingShareLink" @click="handleCreateShareLink">
                  {{ creatingShareLink ? "创建中..." : "新增分享链接" }}
                </button>
              </div>
              <p v-if="inviteActionError" class="state-box warning">{{ inviteActionError }}</p>
              <p v-else-if="inviteActionMessage" class="state-box">{{ inviteActionMessage }}</p>
              <div class="kv-list">
                <p v-if="shareLinks.length === 0">
                  <span>链接状态</span>
                  <strong>暂无分享链接</strong>
                </p>
                <p v-for="item in shareLinks" :key="item.id" class="invite-link-row">
                  <span>{{ item.code }} · {{ mapShareChannel(item.channel) }} · {{ item.status }}</span>
                  <strong>
                    <button type="button" :disabled="item.copying" @click="handleCopyInviteLink(item)">
                      {{ item.copying ? "复制中..." : "复制链接" }}
                    </button>
                  </strong>
                </p>
              </div>
            </article>
          </div>

          <div class="payment-table-wrap">
            <table class="payment-table">
              <thead>
                <tr>
                  <th>被邀请用户</th>
                  <th>注册时间</th>
                  <th>首单支付</th>
                  <th>状态</th>
                  <th>风控</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in inviteRecords" :key="item.id">
                  <td>{{ item.inviteeUser }}</td>
                  <td>{{ item.registerAt }}</td>
                  <td>{{ item.firstPayAt }}</td>
                  <td>{{ item.status }}</td>
                  <td>{{ item.riskFlag }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="payment-mobile">
            <article v-if="inviteRecords.length === 0">
              <div class="top-line">
                <p>暂无邀请记录</p>
              </div>
            </article>
            <article v-for="item in inviteRecords" :key="`invite-${item.id}`">
              <div class="top-line">
                <p>{{ item.inviteeUser }}</p>
                <span>{{ item.status }}</span>
              </div>
              <div class="meta-line">
                <span>注册：{{ item.registerAt }}</span>
                <span>首单：{{ item.firstPayAt }}</span>
                <span>风控：{{ item.riskFlag }}</span>
              </div>
            </article>
          </div>
        </template>

        <template v-else>
          <div class="other-grid">
            <article v-for="item in otherInfos" :key="item.title">
              <h4>{{ item.title }}</h4>
              <div class="kv-list">
                <p v-for="row in item.rows" :key="`${item.title}-${row.key}`">
                  <span>{{ row.key }}</span>
                  <strong>{{ row.value }}</strong>
                </p>
              </div>
            </article>
          </div>
        </template>
      </section>
    </article>

    <div class="bottom-grid">
      <article class="card todo-card">
        <header>
          <h2 class="section-title">待办中心</h2>
          <p class="section-subtitle">把高频操作前置，减少跳转。</p>
        </header>
        <ul>
          <li v-for="todo in todos" :key="todo.title">
            <span class="dot" :class="todo.level" />
            <div>
              <p class="title">{{ todo.title }}</p>
              <p class="note">{{ todo.note }}</p>
            </div>
          </li>
        </ul>
      </article>

      <article class="card quick-card">
        <header>
          <h2 class="section-title">快捷入口</h2>
          <p class="section-subtitle">常用操作统一收敛到个人中心。</p>
        </header>
        <div class="quick-grid">
          <article v-for="item in menus" :key="item.title" class="quick-item">
            <h3>{{ item.title }}</h3>
            <p>{{ item.desc }}</p>
          </article>
        </div>
      </article>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import {
  createShareLink,
  createSubscription,
  getInviteSummary,
  getMembershipQuota,
  getUserProfile,
  listBrowseHistory,
  listInviteRecords,
  listMessages,
  listMembershipOrders,
  listRechargeRecords,
  listShareLinks,
  listSubscriptions,
  readMessage,
  updateSubscription
} from "../api/userCenter";
import { shouldUseDemoFallback } from "../lib/fallback-policy";
import {
  menus,
  modules,
  shareChannelOptions,
  subscriptionFrequencyOptions,
  subscriptionTypeOptions,
  timeRanges
} from "./profile/constants";
import {
  buildInviteURL,
  copyText,
  formatAmount,
  formatDateTime,
  inRange,
  mapContentType,
  mapInviteLinkStatus,
  mapInviteRiskFlag,
  mapInviteStatus,
  mapKYCStatus,
  mapMemberLevel,
  mapMessageReadStatus,
  mapMessageType,
  mapPayChannel,
  mapPaymentStatus,
  mapProduct,
  mapRegistrationSource,
  mapResetCycle,
  mapShareChannel,
  mapSubscriptionFrequency,
  mapSubscriptionStatus,
  mapSubscriptionType,
  mapVIPStatus,
  nextSubscriptionFrequency,
  paymentStatusClass,
  subscriptionStatusClass,
  toArray,
  toTimestamp
} from "./profile/helpers";
import {
  fallbackBrowseHistory,
  fallbackInviteRecords,
  fallbackInviteSummary,
  fallbackMembershipOrders,
  fallbackMessages,
  fallbackProfile,
  fallbackQuota,
  fallbackRechargeRecords,
  fallbackShareLinks,
  fallbackSubscriptions
} from "./profile/fallback";

const useDemoFallback = shouldUseDemoFallback();

const activeModule = ref("vip");
const activeRange = ref(timeRanges[1] || timeRanges[0] || "全部");

const loading = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const creatingSubscription = ref(false);
const subscriptionActionMessage = ref("");
const subscriptionActionError = ref("");
const subscriptionSavingMap = ref({});
const messageActionLoadingMap = ref({});
const creatingShareLink = ref(false);
const inviteActionMessage = ref("");
const inviteActionError = ref("");
const copyInviteID = ref("");
const newShareLinkChannel = ref(shareChannelOptions[0]?.value || "APP");

const newSubscriptionForm = ref({
  type: subscriptionTypeOptions[0]?.value || "STOCK_RECO",
  frequency: "DAILY",
  scope: "ALL"
});

const rawProfile = ref(useDemoFallback ? { ...fallbackProfile } : {});
const rawQuota = ref(useDemoFallback ? { ...fallbackQuota } : {});
const rawMembershipOrders = ref(useDemoFallback ? [...fallbackMembershipOrders] : []);
const rawRechargeRecords = ref(useDemoFallback ? [...fallbackRechargeRecords] : []);
const rawBrowseHistory = ref(useDemoFallback ? [...fallbackBrowseHistory] : []);
const rawSubscriptions = ref(useDemoFallback ? [...fallbackSubscriptions] : []);
const rawMessages = ref(useDemoFallback ? [...fallbackMessages] : []);
const rawShareLinks = ref(useDemoFallback ? [...fallbackShareLinks] : []);
const rawInviteRecords = ref(useDemoFallback ? [...fallbackInviteRecords] : []);
const rawInviteSummary = ref(useDemoFallback ? { ...fallbackInviteSummary } : {});

const currentModule = computed(() => modules.find((item) => item.key === activeModule.value) || modules[0]);

const displayProfile = computed(() => ({
  name: rawProfile.value?.id ? `用户 ${rawProfile.value.id}` : "当前用户",
  phone: rawProfile.value?.phone || "-",
  email: rawProfile.value?.email || "-",
  kycStatus: mapKYCStatus(rawProfile.value?.kyc_status),
  memberLevel: rawProfile.value?.member_level || "FREE"
}));

const paymentRecords = computed(() => {
  const orders = (rawMembershipOrders.value || []).map((item) => ({
    source: "membership",
    orderNo: item.order_no || item.id || "-",
    time: formatDateTime(item.paid_at || item.created_at),
    rawTime: item.paid_at || item.created_at || "",
    product: mapProduct(item.product_id),
    amount: formatAmount(item.amount),
    amountValue: Number(item.amount || 0),
    method: mapPayChannel(item.pay_channel),
    status: mapPaymentStatus(item.status)
  }));

  const recharges = (rawRechargeRecords.value || []).map((item) => ({
    source: "recharge",
    orderNo: item.order_no || item.id || "-",
    time: formatDateTime(item.paid_at || item.created_at),
    rawTime: item.paid_at || item.created_at || "",
    product: "账户充值",
    amount: formatAmount(item.amount),
    amountValue: Number(item.amount || 0),
    method: mapPayChannel(item.pay_channel),
    status: mapPaymentStatus(item.status)
  }));

  return [...orders, ...recharges]
    .filter((item) => inRange(item.rawTime, activeRange.value))
    .sort((a, b) => toTimestamp(b.rawTime) - toTimestamp(a.rawTime));
});

const readingLogs = computed(() =>
  (rawBrowseHistory.value || [])
    .filter((item) => inRange(item.viewed_at, activeRange.value))
    .sort((a, b) => toTimestamp(b.viewed_at) - toTimestamp(a.viewed_at))
    .map((item) => ({
      id: item.id,
      title: item.title || "未命名内容",
      type: mapContentType(item.content_type),
      desc: `来源页面 ${item.source_page || "-"}，内容ID ${item.content_id || "-"}`,
      time: formatDateTime(item.viewed_at),
      duration: resolveReadingDuration(item),
      progress: resolveReadingProgress(item)
    }))
);

const subscriptionItems = computed(() =>
  (rawSubscriptions.value || []).map((item) => {
    const statusRaw = String(item.status || "").toUpperCase();
    const frequencyRaw = String(item.frequency || "").toUpperCase();
    return {
      id: item.id,
      name: mapSubscriptionType(item.type),
      status: mapSubscriptionStatus(statusRaw),
      statusRaw: statusRaw || "ACTIVE",
      desc: `订阅范围：${item.scope || "ALL"}，推送频率：${mapSubscriptionFrequency(frequencyRaw)}`,
      cycle: mapSubscriptionFrequency(frequencyRaw),
      frequencyRaw: frequencyRaw || "DAILY",
      scope: item.scope || "ALL",
      price: "按会员权益",
      saving: !!subscriptionSavingMap.value[item.id]
    };
  })
);

const vipInfo = computed(() => {
  const levelText = mapMemberLevel(displayProfile.value.memberLevel, rawQuota.value?.member_level);
  const expireRaw = rawProfile.value?.vip_expire_at || rawQuota.value?.vip_expire_at || "";
  const expireTs = toTimestamp(expireRaw);
  const computedRemaining = expireTs > 0 ? Math.max(0, Math.ceil((expireTs - Date.now()) / (24 * 3600 * 1000))) : 0;
  const serverRemaining = Number(rawQuota.value?.vip_remaining_days ?? rawProfile.value?.vip_remaining_days ?? 0);
  const remainingDays = Number.isFinite(serverRemaining) && serverRemaining > 0 ? serverRemaining : computedRemaining;
  const expireAt = formatDateTime(expireRaw);
  const nextRenewAt = expireAt;
  const status = mapVIPStatus(rawQuota.value?.vip_status || rawProfile.value?.vip_status, displayProfile.value.memberLevel);

  return {
    level: levelText,
    status,
    expireAt,
    nextRenewAt,
    remainingDays
  };
});

const vipMetrics = computed(() => [
  { label: "会员等级", value: vipInfo.value.level },
  { label: "会员状态", value: vipInfo.value.status },
  {
    label: "文档阅读配额",
    value: `${rawQuota.value?.doc_read_used ?? 0}/${rawQuota.value?.doc_read_limit ?? 0}`
  },
  {
    label: "资讯订阅余量",
    value: `${rawQuota.value?.news_subscribe_remaining ?? 0}`
  },
  { label: "KYC状态", value: displayProfile.value.kycStatus }
]);

const vipBenefits = computed(() => [
  {
    title: "文档阅读额度",
    desc: `周期 ${rawQuota.value?.period_key || "-"}，已用 ${
      rawQuota.value?.doc_read_used ?? 0
    } / ${rawQuota.value?.doc_read_limit ?? 0}。`
  },
  {
    title: "资讯订阅额度",
    desc: `已用 ${rawQuota.value?.news_subscribe_used ?? 0} / ${
      rawQuota.value?.news_subscribe_limit ?? 0
    }，剩余 ${rawQuota.value?.news_subscribe_remaining ?? 0}。`
  },
  {
    title: "重置周期",
    desc: `周期 ${mapResetCycle(rawQuota.value?.reset_cycle)}，下次重置 ${formatDateTime(
      rawQuota.value?.reset_at
    )}。`
  },
  {
    title: "会员续费状态",
    desc: `当前会员状态：${vipInfo.value.status}，到期时间：${vipInfo.value.expireAt}。`
  }
]);

const paymentSummary = computed(() => {
  const paid = paymentRecords.value.filter((item) => item.status === "已支付");
  const pending = paymentRecords.value.filter((item) => item.status === "处理中");
  const refund = paymentRecords.value.filter((item) => item.status === "已退款");

  const paidAmount = paid.reduce((acc, item) => acc + item.amountValue, 0);
  const refundAmount = refund.reduce((acc, item) => acc + item.amountValue, 0);

  return [
    { label: `${activeRange.value}支付总额`, value: formatAmount(paidAmount) },
    { label: "成功支付笔数", value: `${paid.length} 笔` },
    { label: "待处理订单", value: `${pending.length} 笔` },
    { label: "退款金额", value: formatAmount(refundAmount) }
  ];
});

const readingStats = computed(() => {
  const items = readingLogs.value;
  const reportCount = items.filter((item) => item.type === "研报").length;
  const journalCount = items.filter((item) => item.type === "期刊").length;
  const newsCount = items.filter((item) => item.type === "新闻").length;

  return [
    { label: `${activeRange.value}阅读总量`, value: `${items.length} 篇` },
    { label: "研报阅读", value: `${reportCount} 篇` },
    { label: "期刊阅读", value: `${journalCount} 篇` },
    { label: "新闻阅读", value: `${newsCount} 篇` }
  ];
});

const messageItems = computed(() =>
  (rawMessages.value || [])
    .filter((item) => inRange(item.created_at, activeRange.value))
    .sort((a, b) => toTimestamp(b.created_at) - toTimestamp(a.created_at))
    .map((item) => {
      const readStatusRaw = String(item.read_status || "UNREAD").toUpperCase();
      return {
        id: item.id,
        title: item.title || "未命名通知",
        content: item.content || "-",
        type: mapMessageType(item.type),
        time: formatDateTime(item.created_at),
        readStatus: mapMessageReadStatus(readStatusRaw),
        readStatusRaw,
        loading: !!messageActionLoadingMap.value[item.id]
      };
    })
);

const messageStats = computed(() => {
  const list = messageItems.value;
  const unread = list.filter((item) => item.readStatusRaw !== "READ").length;
  const read = list.length - unread;
  const strategy = list.filter((item) => item.type === "策略提醒").length;
  return [
    { label: `${activeRange.value}通知总量`, value: `${list.length} 条` },
    { label: "未读通知", value: `${unread} 条` },
    { label: "已读通知", value: `${read} 条` },
    { label: "策略提醒", value: `${strategy} 条` }
  ];
});

const shareLinks = computed(() =>
  (rawShareLinks.value || []).map((item) => ({
    id: item.id,
    code: item.invite_code || "-",
    channel: item.channel || "-",
    status: mapInviteLinkStatus(item.status),
    rawStatus: item.status || "",
    expiredAt: formatDateTime(item.expired_at),
    url: item.url || "",
    shareURL: buildInviteURL(item.url, item.invite_code),
    copying: copyInviteID.value === item.id
  }))
);

const inviteRecords = computed(() =>
  (rawInviteRecords.value || [])
    .filter((item) => inRange(item.register_at, activeRange.value))
    .sort((a, b) => toTimestamp(b.register_at) - toTimestamp(a.register_at))
    .map((item) => ({
      id: item.id,
      inviteeUser: item.invitee_user_id || "-",
      registerAt: formatDateTime(item.register_at),
      firstPayAt: formatDateTime(item.first_pay_at),
      status: mapInviteStatus(item.status),
      riskFlag: mapInviteRiskFlag(item.risk_flag),
      statusRaw: String(item.status || "").toUpperCase()
    }))
);

const inviteSourceInfo = computed(() => ({
  registrationSource: mapRegistrationSource(rawProfile.value?.registration_source),
  inviterUserID: rawProfile.value?.inviter_user_id || "-",
  inviteCode: rawProfile.value?.invite_code || "-",
  invitedAt: formatDateTime(rawProfile.value?.invited_at)
}));

const inviteStats = computed(() => {
  const summary = rawInviteSummary.value || {};
  let invitedCount = Number(summary.registered_count || 0);
  let convertedCount = Number(summary.first_paid_count || 0);
  let conversionRate = Number(summary.conversion_rate || 0);

  if (activeRange.value === "近7天") {
    invitedCount = Number(summary.last_7d_registered_count || 0);
    convertedCount = Number(summary.last_7d_first_paid_count || 0);
    conversionRate = Number(summary.last_7d_conversion_rate || 0);
  } else if (activeRange.value === "近30天") {
    invitedCount = Number(summary.last_30d_registered_count || 0);
    convertedCount = Number(summary.last_30d_first_paid_count || 0);
    conversionRate = Number(summary.last_30d_conversion_rate || 0);
  }

  const activeShareLinks = Number(summary.share_link_count || 0);
  const window7Rate = Number(summary.last_7d_conversion_rate || 0);
  const window30Rate = Number(summary.last_30d_conversion_rate || 0);
  return [
    { label: `${activeRange.value}邀请注册`, value: `${invitedCount} 人` },
    { label: "首单转化", value: `${convertedCount} 人` },
    { label: "生效分享链接", value: `${activeShareLinks} 条` },
    { label: `${activeRange.value}转化率`, value: `${(conversionRate * 100).toFixed(1)}%` },
    { label: "近7天转化率", value: `${(window7Rate * 100).toFixed(1)}%` },
    { label: "近30天转化率", value: `${(window30Rate * 100).toFixed(1)}%` },
    { label: "我的注册来源", value: inviteSourceInfo.value.registrationSource }
  ];
});

const otherInfos = computed(() => [
  {
    title: "账户基础信息",
    rows: [
      { key: "客户编号", value: rawProfile.value?.id || "-" },
      { key: "手机号", value: rawProfile.value?.phone || "-" },
      { key: "邮箱", value: rawProfile.value?.email || "-" }
    ]
  },
  {
    title: "会员与配额",
    rows: [
      { key: "会员等级", value: vipInfo.value.level },
      { key: "文档配额剩余", value: `${rawQuota.value?.doc_read_remaining ?? 0}` },
      { key: "资讯订阅剩余", value: `${rawQuota.value?.news_subscribe_remaining ?? 0}` }
    ]
  },
  {
    title: "记录统计",
    rows: [
      { key: "支付记录", value: `${paymentRecords.value.length} 条` },
      { key: "阅读记录", value: `${readingLogs.value.length} 条` },
      { key: "订阅项", value: `${subscriptionItems.value.length} 条` },
      { key: "通知消息", value: `${messageItems.value.length} 条` },
      { key: "邀请记录", value: `${inviteRecords.value.length} 条` }
    ]
  },
  {
    title: "邀请关系",
    rows: [
      { key: "注册来源", value: inviteSourceInfo.value.registrationSource },
      { key: "邀请人", value: inviteSourceInfo.value.inviterUserID },
      { key: "邀请码", value: inviteSourceInfo.value.inviteCode },
      { key: "分享链接数", value: `${shareLinks.value.length} 条` }
    ]
  }
]);

const todos = computed(() => [
  { title: "开启二次验证", note: "建议 24 小时内完成", level: "high" },
  {
    title: "检查订阅自动续费",
    note: `当前有效订阅 ${subscriptionItems.value.filter((item) => item.status === "生效中").length} 项`,
    level: "mid"
  },
  { title: "导出支付与阅读记录", note: "可用于周度复盘和对账", level: "low" }
]);

async function loadUserCenterData() {
  loading.value = true;
  loadError.value = "";
  subscriptionActionMessage.value = "";
  subscriptionActionError.value = "";
  inviteActionMessage.value = "";
  inviteActionError.value = "";
  const tasks = [
    {
      key: "profile",
      label: "用户资料",
      request: getUserProfile(),
      apply: (data) => {
        rawProfile.value = data || rawProfile.value;
      }
    },
    {
      key: "quota",
      label: "会员配额",
      request: getMembershipQuota(),
      apply: (data) => {
        rawQuota.value = data || rawQuota.value;
      }
    },
    {
      key: "orders",
      label: "会员订单",
      request: listMembershipOrders({ page: 1, page_size: 50 }),
      apply: (data) => {
        rawMembershipOrders.value = toArray(data?.items, rawMembershipOrders.value);
      }
    },
    {
      key: "recharges",
      label: "充值记录",
      request: listRechargeRecords({ page: 1, page_size: 50 }),
      apply: (data) => {
        rawRechargeRecords.value = toArray(data?.items, rawRechargeRecords.value);
      }
    },
    {
      key: "browses",
      label: "阅读记录",
      request: listBrowseHistory({ page: 1, page_size: 100 }),
      apply: (data) => {
        rawBrowseHistory.value = toArray(data?.items, rawBrowseHistory.value);
      }
    },
    {
      key: "subscriptions",
      label: "订阅列表",
      request: listSubscriptions({ page: 1, page_size: 50 }),
      apply: (data) => {
        applySubscriptionItems(toArray(data?.items, rawSubscriptions.value));
      }
    },
    {
      key: "messages",
      label: "通知消息",
      request: listMessages({ page: 1, page_size: 100 }),
      apply: (data) => {
        rawMessages.value = toArray(data?.items, rawMessages.value);
        messageActionLoadingMap.value = {};
      }
    },
    {
      key: "shareLinks",
      label: "分享链接",
      request: listShareLinks(),
      apply: (data) => {
        rawShareLinks.value = toArray(data?.items, rawShareLinks.value);
      }
    },
    {
      key: "invites",
      label: "邀请记录",
      request: listInviteRecords({ page: 1, page_size: 100 }),
      apply: (data) => {
        rawInviteRecords.value = toArray(data?.items, rawInviteRecords.value);
      }
    },
    {
      key: "inviteSummary",
      label: "邀请汇总",
      request: getInviteSummary(),
      apply: (data) => {
        rawInviteSummary.value = data || rawInviteSummary.value;
      }
    }
  ];

  const results = await Promise.allSettled(tasks.map((item) => item.request));
  const errors = [];
  results.forEach((result, index) => {
    const task = tasks[index];
    if (result.status === "fulfilled") {
      task.apply(result.value);
      return;
    }
    errors.push(`${task.label}加载失败：${result.reason?.message || "unknown error"}`);
  });

  loadError.value = errors.join("；");
  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  loading.value = false;
}

async function refreshInviteData() {
  const tasks = [
    {
      label: "用户资料",
      request: getUserProfile(),
      apply: (data) => {
        rawProfile.value = data || rawProfile.value;
      }
    },
    {
      label: "分享链接",
      request: listShareLinks(),
      apply: (data) => {
        rawShareLinks.value = toArray(data?.items, rawShareLinks.value);
      }
    },
    {
      label: "邀请记录",
      request: listInviteRecords({ page: 1, page_size: 100 }),
      apply: (data) => {
        rawInviteRecords.value = toArray(data?.items, rawInviteRecords.value);
      }
    },
    {
      label: "邀请汇总",
      request: getInviteSummary(),
      apply: (data) => {
        rawInviteSummary.value = data || rawInviteSummary.value;
      }
    }
  ];

  const results = await Promise.allSettled(tasks.map((item) => item.request));
  const errors = [];
  results.forEach((result, index) => {
    const task = tasks[index];
    if (result.status === "fulfilled") {
      task.apply(result.value);
      return;
    }
    errors.push(`${task.label}刷新失败：${result.reason?.message || "unknown error"}`);
  });
  if (errors.length > 0) {
    throw new Error(errors.join("；"));
  }
}

async function handleCreateShareLink() {
  if (creatingShareLink.value) {
    return;
  }
  creatingShareLink.value = true;
  inviteActionMessage.value = "";
  inviteActionError.value = "";
  try {
    const payload = {
      channel: newShareLinkChannel.value
    };
    const result = await createShareLink(payload);
    if (result?.id) {
      rawShareLinks.value = [result, ...(rawShareLinks.value || [])];
    } else {
      await refreshInviteData();
    }
    inviteActionMessage.value = `已创建分享链接（${mapShareChannel(newShareLinkChannel.value)}）`;
  } catch (error) {
    inviteActionError.value = error?.message || "创建分享链接失败";
  } finally {
    creatingShareLink.value = false;
  }
}

async function handleCopyInviteLink(item) {
  if (!item?.id || !item?.shareURL) {
    return;
  }
  copyInviteID.value = item.id;
  inviteActionMessage.value = "";
  inviteActionError.value = "";
  try {
    await copyText(item.shareURL);
    inviteActionMessage.value = `已复制邀请码 ${item.code} 的分享链接`;
  } catch (error) {
    inviteActionError.value = error?.message || "复制失败，请手动复制";
  } finally {
    copyInviteID.value = "";
  }
}

async function handleReadMessage(item) {
  if (!item?.id || item.readStatusRaw === "READ") {
    return;
  }
  setMessageSaving(item.id, true);
  loadError.value = "";
  try {
    await readMessage(item.id);
    rawMessages.value = (rawMessages.value || []).map((row) => {
      if (row.id === item.id) {
        return { ...row, read_status: "READ" };
      }
      return row;
    });
  } catch (error) {
    loadError.value = error?.message || "标记已读失败";
  } finally {
    setMessageSaving(item.id, false);
  }
}

async function refreshSubscriptions() {
  const subscriptionData = await listSubscriptions({ page: 1, page_size: 50 });
  applySubscriptionItems(toArray(subscriptionData?.items, rawSubscriptions.value));
}

async function handleCreateSubscription() {
  if (creatingSubscription.value) {
    return;
  }
  creatingSubscription.value = true;
  subscriptionActionMessage.value = "";
  subscriptionActionError.value = "";
  try {
    const payload = {
      type: newSubscriptionForm.value.type,
      frequency: newSubscriptionForm.value.frequency,
      scope: newSubscriptionForm.value.scope || "ALL"
    };
    const result = await createSubscription(payload);
    await refreshSubscriptions();
    const createdID = result?.id || "-";
    subscriptionActionMessage.value = `订阅创建成功：${createdID}`;
  } catch (error) {
    subscriptionActionError.value = error?.message || "创建订阅失败";
  } finally {
    creatingSubscription.value = false;
  }
}

async function handleToggleSubscriptionStatus(item) {
  const targetStatus = item.statusRaw === "ACTIVE" ? "PAUSED" : "ACTIVE";
  await updateSubscriptionWithFeedback(
    item,
    { frequency: item.frequencyRaw, status: targetStatus },
    `订阅状态已更新为 ${mapSubscriptionStatus(targetStatus)}`
  );
}

async function handleRotateSubscriptionFrequency(item) {
  const nextFrequency = nextSubscriptionFrequency(item.frequencyRaw);
  await updateSubscriptionWithFeedback(
    item,
    { frequency: nextFrequency, status: item.statusRaw },
    `订阅频率已更新为 ${mapSubscriptionFrequency(nextFrequency)}`
  );
}

async function updateSubscriptionWithFeedback(item, payload, successMessage) {
  if (!item?.id) {
    return;
  }
  setSubscriptionSaving(item.id, true);
  subscriptionActionMessage.value = "";
  subscriptionActionError.value = "";
  try {
    await updateSubscription(item.id, payload);
    await refreshSubscriptions();
    subscriptionActionMessage.value = successMessage;
  } catch (error) {
    subscriptionActionError.value = error?.message || "更新订阅失败";
  } finally {
    setSubscriptionSaving(item.id, false);
  }
}

function setSubscriptionSaving(id, loadingState) {
  const next = { ...subscriptionSavingMap.value };
  if (loadingState) {
    next[id] = true;
  } else {
    delete next[id];
  }
  subscriptionSavingMap.value = next;
}

function setMessageSaving(id, loadingState) {
  const next = { ...messageActionLoadingMap.value };
  if (loadingState) {
    next[id] = true;
  } else {
    delete next[id];
  }
  messageActionLoadingMap.value = next;
}

function applySubscriptionItems(items) {
  rawSubscriptions.value = toArray(items, rawSubscriptions.value);
  subscriptionSavingMap.value = {};
}

function resolveReadingDuration(item) {
  const secondCandidates = [
    item?.duration_seconds,
    item?.duration_sec,
    item?.read_duration_seconds,
    item?.read_seconds
  ];
  const msCandidates = [item?.duration_ms, item?.read_duration_ms];
  for (const value of secondCandidates) {
    const seconds = Number(value);
    if (Number.isFinite(seconds) && seconds > 0) {
      return formatDuration(seconds);
    }
  }
  for (const value of msCandidates) {
    const milliseconds = Number(value);
    if (Number.isFinite(milliseconds) && milliseconds > 0) {
      return formatDuration(milliseconds / 1000);
    }
  }
  return "-";
}

function resolveReadingProgress(item) {
  const candidates = [
    item?.progress,
    item?.progress_rate,
    item?.completion_rate,
    item?.read_percent,
    item?.finish_percent
  ];
  for (const value of candidates) {
    const num = Number(value);
    if (!Number.isFinite(num) || num < 0) {
      continue;
    }
    const normalized = num > 1 ? num / 100 : num;
    const clipped = Math.max(0, Math.min(1, normalized));
    return `${Math.round(clipped * 100)}%`;
  }
  return "100%";
}

function formatDuration(secondsValue) {
  const seconds = Math.max(0, Math.round(Number(secondsValue)));
  if (!Number.isFinite(seconds) || seconds <= 0) {
    return "-";
  }
  const hour = Math.floor(seconds / 3600);
  const minute = Math.floor((seconds % 3600) / 60);
  const second = seconds % 60;
  if (hour > 0) {
    return `${hour}小时${minute}分`;
  }
  if (minute > 0) {
    return `${minute}分${second}秒`;
  }
  return `${second}秒`;
}

function openSecurityPanel() {
  activeModule.value = "other";
  if (typeof window === "undefined") {
    return;
  }
  const target = document.querySelector(".query-card");
  if (target) {
    target.scrollIntoView({ behavior: "smooth", block: "start" });
  }
}

onMounted(() => {
  loadUserCenterData();
});
</script>

<style scoped>
.profile-page {
  display: grid;
  gap: 12px;
}

.account-card {
  border-radius: 20px;
  padding: 15px;
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
  align-items: center;
  background:
    radial-gradient(circle at 100% 0%, rgba(63, 127, 113, 0.16) 0%, transparent 36%),
    radial-gradient(circle at 0% 100%, rgba(234, 215, 180, 0.2) 0%, transparent 34%),
    rgba(255, 255, 255, 0.93);
}

.identity {
  display: flex;
  align-items: center;
  gap: 10px;
}

.avatar {
  width: 58px;
  height: 58px;
  border-radius: 16px;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
  color: #fff;
  font-size: 24px;
  font-family: var(--font-serif);
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

h1 {
  margin: 0;
  font-size: 24px;
}

.identity p {
  margin: 4px 0 0;
  color: var(--color-text-sub);
  font-size: 13px;
}

.actions {
  display: inline-flex;
  gap: 8px;
}

.actions button {
  border: 0;
  border-radius: 10px;
  padding: 9px 12px;
  cursor: pointer;
  font-weight: 600;
}

.actions .primary {
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.actions .ghost {
  color: var(--color-pine-700);
  background: rgba(239, 232, 218, 0.82);
}

.query-card {
  padding: 14px;
}

.query-head {
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: end;
  gap: 10px;
}

.range-switch {
  display: inline-flex;
  gap: 6px;
}

.range-switch button {
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(252, 251, 247, 0.9);
  color: var(--color-text-sub);
  border-radius: 9px;
  padding: 6px 10px;
  cursor: pointer;
  font-size: 12px;
}

.range-switch button.active {
  color: #fff;
  border-color: transparent;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.query-nav {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
  gap: 8px;
}

.query-nav button {
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(252, 251, 247, 0.9);
  border-radius: 11px;
  padding: 9px 12px;
  cursor: pointer;
  font-weight: 600;
  color: var(--color-text-sub);
}

.query-nav button.active {
  color: #fff;
  border-color: transparent;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.query-tip {
  margin-top: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  border-radius: 10px;
  border: 1px dashed rgba(216, 223, 216, 0.95);
  background: rgba(246, 244, 239, 0.76);
  padding: 8px 10px;
}

.query-tip p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.state-box {
  margin-top: 10px;
  border-radius: 10px;
  padding: 8px 10px;
  font-size: 12px;
  color: var(--color-pine-700);
  background: rgba(223, 236, 230, 0.66);
}

.state-box.warning {
  color: #7f5f36;
  background: rgba(234, 215, 180, 0.56);
}

.query-body {
  margin-top: 10px;
  display: grid;
  gap: 10px;
}

.vip-panel {
  display: grid;
  gap: 10px;
  grid-template-columns: 0.95fr 1.05fr;
}

.vip-main {
  border-radius: 13px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: linear-gradient(160deg, rgba(36, 83, 73, 0.95), rgba(45, 109, 95, 0.95));
  color: #f5faf7;
  padding: 12px;
}

.vip-level {
  margin: 0;
  font-size: 12px;
  color: rgba(245, 250, 247, 0.76);
}

.vip-main h3 {
  margin: 6px 0;
  font-size: 24px;
  font-family: var(--font-serif);
  line-height: 1.3;
}

.vip-main p {
  margin: 0;
  font-size: 13px;
  color: rgba(245, 250, 247, 0.82);
}

.summary-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.summary-grid article {
  border-radius: 11px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 9px;
}

.summary-grid p {
  margin: 0;
  color: var(--color-text-sub);
  font-size: 12px;
}

.summary-grid strong {
  margin-top: 5px;
  display: block;
  color: var(--color-pine-700);
  font-size: 18px;
}

.benefits-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.benefits-grid article {
  border-radius: 11px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
}

.benefits-grid h4 {
  margin: 0;
  font-size: 14px;
}

.benefits-grid p {
  margin: 5px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.55;
}

.payment-table-wrap {
  overflow-x: auto;
}

.payment-table {
  width: 100%;
  border-collapse: collapse;
  min-width: 760px;
}

.payment-table th,
.payment-table td {
  border-bottom: 1px solid rgba(216, 223, 216, 0.75);
  padding: 9px 8px;
  text-align: left;
  font-size: 13px;
  white-space: nowrap;
}

.payment-table th {
  font-size: 12px;
  color: var(--color-text-sub);
  background: rgba(246, 244, 239, 0.72);
}

.status {
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
}

.status.success {
  color: var(--color-pine-700);
  background: rgba(223, 236, 230, 0.72);
}

.status.pending {
  color: #775325;
  background: rgba(234, 215, 180, 0.72);
}

.status.refund,
.status.inactive,
.status.fail {
  color: #8a3c2f;
  background: rgba(230, 194, 185, 0.62);
}

.payment-mobile {
  display: none;
}

.payment-mobile article,
.log-list article,
.subscription-item,
.message-item,
.other-grid article,
.todo-card li,
.quick-item {
  border-radius: 11px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
}

.top-line {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.top-line p {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.top-line span {
  font-size: 12px;
  color: var(--color-text-sub);
}

.meta-line {
  margin-top: 6px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.meta-line span {
  font-size: 12px;
  color: var(--color-text-sub);
}

.order {
  margin: 6px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.log-list {
  display: grid;
  gap: 8px;
}

.message-list {
  display: grid;
  gap: 8px;
}

.log-list .desc {
  margin: 5px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.58;
}

.message-actions {
  margin-top: 8px;
  display: flex;
  justify-content: flex-end;
}

.message-actions button {
  border: 0;
  border-radius: 8px;
  padding: 7px 12px;
  cursor: pointer;
  color: var(--color-pine-700);
  background: rgba(223, 236, 230, 0.72);
}

.message-actions button:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.subscription-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.subscription-create {
  margin-bottom: 8px;
  border-radius: 11px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(246, 244, 239, 0.78);
  padding: 10px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 8px;
  align-items: end;
}

.subscription-create label {
  display: grid;
  gap: 4px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.subscription-create select,
.subscription-create input {
  border-radius: 8px;
  border: 1px solid rgba(216, 223, 216, 0.95);
  background: #fff;
  padding: 7px 8px;
  color: var(--color-text-main);
}

.subscription-create .scope-input {
  grid-column: span 2;
}

.subscription-create button {
  border: 0;
  border-radius: 10px;
  padding: 9px 12px;
  cursor: pointer;
  color: #fff;
  font-weight: 600;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.subscription-create button:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.invite-create {
  margin: 8px 0;
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

.invite-create label {
  display: grid;
  gap: 4px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.invite-create select {
  border-radius: 8px;
  border: 1px solid rgba(216, 223, 216, 0.95);
  background: #fff;
  padding: 7px 8px;
  color: var(--color-text-main);
}

.invite-create button {
  border: 0;
  border-radius: 9px;
  padding: 8px 11px;
  color: #fff;
  cursor: pointer;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.invite-create button:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.invite-link-row button {
  border: 0;
  border-radius: 8px;
  padding: 5px 8px;
  color: var(--color-pine-700);
  background: rgba(223, 236, 230, 0.72);
  cursor: pointer;
}

.invite-link-row button:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.subscription-item .desc {
  margin: 5px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.58;
}

.subscription-actions {
  margin-top: 10px;
  display: flex;
  gap: 8px;
}

.subscription-item button {
  flex: 1;
  border: 0;
  border-radius: 10px;
  padding: 8px 10px;
  cursor: pointer;
  color: #fff;
  font-weight: 600;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.subscription-item button.secondary {
  color: var(--color-pine-700);
  background: rgba(223, 236, 230, 0.72);
}

.subscription-item button:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.other-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.other-grid h4 {
  margin: 0;
  font-size: 14px;
}

.kv-list {
  margin-top: 7px;
  display: grid;
  gap: 6px;
}

.kv-list p {
  margin: 0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.kv-list strong {
  color: var(--color-text-main);
  font-size: 13px;
}

.bottom-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: 1fr 1fr;
}

.todo-card,
.quick-card {
  padding: 14px;
}

.todo-card ul {
  margin: 10px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 8px;
}

.todo-card li {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 8px;
}

.dot {
  margin-top: 5px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.dot.high {
  background: #e8c07d;
}

.dot.mid {
  background: #90d0bd;
}

.dot.low {
  background: #bcd7cf;
}

.title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.note {
  margin: 3px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.quick-grid {
  margin-top: 10px;
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.quick-item h3 {
  margin: 0;
  font-size: 16px;
}

.quick-item p {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.56;
}

@media (max-width: 1080px) {
  .query-nav {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .subscription-create {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .vip-panel,
  .subscription-grid,
  .other-grid {
    grid-template-columns: 1fr;
  }

  .summary-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 980px) {
  .account-card,
  .query-head,
  .bottom-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .identity {
    align-items: flex-start;
  }

  .identity p {
    line-height: 1.5;
  }

  .actions {
    width: 100%;
    display: grid;
    grid-template-columns: 1fr;
  }

  .range-switch,
  .query-nav {
    display: flex;
    overflow-x: auto;
    padding-bottom: 2px;
    scrollbar-width: none;
  }

  .range-switch::-webkit-scrollbar,
  .query-nav::-webkit-scrollbar {
    display: none;
  }

  .range-switch button,
  .query-nav button {
    flex: 0 0 auto;
    min-width: 92px;
  }

  .query-tip {
    flex-direction: column;
    align-items: flex-start;
  }

  .summary-grid,
  .benefits-grid,
  .quick-grid {
    grid-template-columns: 1fr;
  }

  .subscription-create {
    grid-template-columns: 1fr;
  }

  .subscription-create .scope-input {
    grid-column: auto;
  }

  .subscription-actions {
    flex-direction: column;
  }

  .payment-table-wrap {
    display: none;
  }

  .payment-mobile {
    display: grid;
    gap: 8px;
  }
}
</style>
