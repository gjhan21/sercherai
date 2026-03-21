import { rangeDayMap } from "./constants";

export function nextSubscriptionFrequency(current) {
  const order = ["INSTANT", "DAILY", "WEEKLY"];
  const currentValue = String(current || "").toUpperCase();
  const index = order.indexOf(currentValue);
  if (index < 0) {
    return "DAILY";
  }
  return order[(index + 1) % order.length];
}

export function toArray(value, fallback = []) {
  return Array.isArray(value) ? value : fallback;
}

export function mapMemberLevel(level, quotaLevel) {
  const source = String(level || quotaLevel || "FREE").toUpperCase();
  if (source === "VIP1") {
    return "VIP 1";
  }
  if (source === "VIP2") {
    return "VIP 2";
  }
  if (source === "VIP3") {
    return "VIP 3";
  }
  return source === "FREE" ? "普通会员" : source;
}

export function mapVIPStatus(status, level) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "ACTIVE") {
    return "生效中";
  }
  if (normalized === "EXPIRED") {
    return "已到期";
  }
  const levelText = String(level || "").toUpperCase();
  if (levelText.startsWith("VIP")) {
    return "生效中";
  }
  return "未开通";
}

export function mapKYCStatus(status) {
  const source = String(status || "").toUpperCase();
  if (source === "APPROVED" || source === "VERIFIED") {
    return "已认证";
  }
  if (source === "REJECTED") {
    return "未通过";
  }
  if (source === "PENDING") {
    return "审核中";
  }
  return source || "-";
}

export function mapActivationState(state) {
  const source = String(state || "").toUpperCase();
  if (source === "ACTIVE") {
    return "已激活";
  }
  if (source === "PAID_PENDING_KYC") {
    return "待实名激活";
  }
  if (source === "NON_MEMBER") {
    return "未开通";
  }
  return source || "-";
}

export function mapResetCycle(cycle) {
  const source = String(cycle || "").toUpperCase();
  if (source === "MONTHLY") {
    return "每月";
  }
  if (source === "WEEKLY") {
    return "每周";
  }
  return source || "-";
}

export function mapProduct(productID) {
  if (!productID) {
    return "会员订单";
  }
  const mapping = {
    mp_demo_001: "VIP月卡",
    mp_demo_002: "VIP季卡",
    mp_demo_003: "VIP年卡"
  };
  return mapping[productID] || `会员产品 ${productID}`;
}

export function mapPayChannel(channel) {
  const source = String(channel || "").toUpperCase();
  if (source === "ALIPAY") {
    return "支付宝";
  }
  if (source === "WECHAT" || source === "WECHAT_PAY") {
    return "微信支付";
  }
  if (source === "CARD" || source === "BANK") {
    return "银行卡";
  }
  if (source === "YOLKPAY") {
    return "蛋黄支付";
  }
  if (source === "BALANCE") {
    return "余额";
  }
  return source || "-";
}

export function mapPaymentStatus(status) {
  const source = String(status || "").toUpperCase();
  if (source === "PAID" || source === "SUCCESS") {
    return "已支付";
  }
  if (source === "REFUND" || source === "REFUNDED") {
    return "已退款";
  }
  if (source === "FAILED" || source === "CLOSED" || source === "CANCELLED" || source === "CANCELED") {
    return "失败";
  }
  return "处理中";
}

export function mapSubscriptionType(type) {
  const source = String(type || "").toUpperCase();
  const mapping = {
    STOCK_RECO: "股票推荐订阅",
    FUTURES_STRATEGY: "期货策略订阅",
    ARBITRAGE: "套利信号订阅",
    EVENT: "事件提醒订阅",
    FUTURES_ALERT: "期货提醒订阅",
    NEWS_DIGEST: "资讯摘要订阅"
  };
  return mapping[source] || source || "未命名订阅";
}

export function mapSubscriptionFrequency(frequency) {
  const source = String(frequency || "").toUpperCase();
  const mapping = {
    INSTANT: "实时",
    REALTIME: "实时",
    DAILY: "每日",
    WEEKLY: "每周"
  };
  return mapping[source] || source || "-";
}

export function mapSubscriptionStatus(status) {
  const source = String(status || "").toUpperCase();
  if (source === "ACTIVE") {
    return "生效中";
  }
  if (source === "PAUSED" || source === "INACTIVE") {
    return "已暂停";
  }
  if (source === "TRIAL") {
    return "试用中";
  }
  return source || "-";
}

export function mapContentType(contentType) {
  const source = String(contentType || "").toUpperCase();
  if (source === "NEWS") {
    return "新闻";
  }
  if (source === "REPORT") {
    return "研报";
  }
  if (source === "JOURNAL") {
    return "期刊";
  }
  return source || "其他";
}

export function mapMessageType(messageType) {
  const source = String(messageType || "").toUpperCase();
  if (source === "SYSTEM") {
    return "系统通知";
  }
  if (source === "STRATEGY") {
    return "策略提醒";
  }
  if (source === "ALERT") {
    return "风险告警";
  }
  if (source === "NEWS") {
    return "资讯通知";
  }
  return source || "其他通知";
}

export function mapMessageReadStatus(readStatus) {
  return String(readStatus || "").toUpperCase() === "READ" ? "已读" : "未读";
}

export function mapInviteStatus(status) {
  const source = String(status || "").toUpperCase();
  if (source === "REGISTERED") {
    return "已注册";
  }
  if (source === "FIRST_PAID") {
    return "首单完成";
  }
  if (source === "INVALID") {
    return "失效";
  }
  return source || "-";
}

export function mapInviteRiskFlag(value) {
  const source = String(value || "").toUpperCase();
  if (source === "NORMAL") {
    return "正常";
  }
  if (source === "RISK") {
    return "疑似风险";
  }
  if (source === "BLOCKED") {
    return "已拦截";
  }
  return source || "-";
}

export function mapRegistrationSource(sourceValue) {
  const source = String(sourceValue || "").toUpperCase();
  if (source === "INVITED") {
    return "邀请注册";
  }
  if (source === "DIRECT") {
    return "自然注册";
  }
  return source || "-";
}

export function mapInviteLinkStatus(status) {
  const source = String(status || "").toUpperCase();
  if (source === "ACTIVE") {
    return "生效中";
  }
  if (source === "DISABLED") {
    return "已停用";
  }
  if (source === "EXPIRED") {
    return "已过期";
  }
  return source || "-";
}

export function mapShareChannel(channel) {
  const source = String(channel || "").toUpperCase();
  if (source === "APP") {
    return "App内分享";
  }
  if (source === "WECHAT") {
    return "微信";
  }
  if (source === "WEIBO") {
    return "微博";
  }
  if (source === "QQ") {
    return "QQ";
  }
  return source || "-";
}

export function buildInviteURL(urlValue, inviteCode) {
  const direct = String(urlValue || "").trim();
  const code = String(inviteCode || "").trim();
  if (direct) {
    if (direct.includes("example.com/invite/") && code) {
      return buildInviteURL("", code);
    }
    if (direct.startsWith("/") && typeof window !== "undefined" && window.location?.origin) {
      return `${window.location.origin}${direct}`;
    }
    return direct;
  }
  if (!code) {
    return "";
  }
  if (typeof window !== "undefined" && window.location?.origin) {
    return `${window.location.origin}/invite/${encodeURIComponent(code)}`;
  }
  return `/invite/${encodeURIComponent(code)}`;
}

export async function copyText(text) {
  if (navigator?.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }
  const input = document.createElement("textarea");
  input.value = text;
  input.style.position = "fixed";
  input.style.opacity = "0";
  document.body.appendChild(input);
  input.focus();
  input.select();
  document.execCommand("copy");
  document.body.removeChild(input);
}

export function toTimestamp(value) {
  const parsed = Date.parse(value || "");
  return Number.isNaN(parsed) ? 0 : parsed;
}

export function inRange(value, range) {
  const days = rangeDayMap[range];
  if (!days) {
    return true;
  }
  const ts = toTimestamp(value);
  if (!ts) {
    return false;
  }
  const now = Date.now();
  return now - ts <= days * 24 * 60 * 60 * 1000;
}

export function formatDateTime(value) {
  const ts = toTimestamp(value);
  if (!ts) {
    return "-";
  }
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

export function formatAmount(value) {
  const num = Number(value || 0);
  return `¥${num.toFixed(2).replace(/\.00$/, "")}`;
}

export function paymentStatusClass(status) {
  if (status === "已支付") {
    return "success";
  }
  if (status === "已退款") {
    return "refund";
  }
  if (status === "失败") {
    return "fail";
  }
  return "pending";
}

export function subscriptionStatusClass(status) {
  if (status === "生效中") {
    return "success";
  }
  if (status === "已暂停") {
    return "inactive";
  }
  return "pending";
}
