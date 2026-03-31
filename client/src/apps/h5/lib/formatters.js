import {
  mapActivationState,
  mapContentType,
  mapInviteStatus,
  mapMemberLevel,
  mapMessageReadStatus,
  mapMessageType,
  mapPayChannel,
  mapPaymentStatus,
  mapSubscriptionFrequency,
  mapSubscriptionStatus,
  mapSubscriptionType,
  mapVIPStatus
} from "../../../lib/profile/helpers.js";
export {
  mapActivationState,
  mapContentType,
  mapInviteStatus,
  mapMemberLevel,
  mapMessageReadStatus,
  mapMessageType,
  mapPayChannel,
  mapPaymentStatus,
  mapSubscriptionFrequency,
  mapSubscriptionStatus,
  mapSubscriptionType,
  mapVIPStatus
};

export function toArray(value, fallback = []) {
  return Array.isArray(value) ? value : fallback;
}

export function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return value || "-";
  }
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

export function formatDate(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return value || "-";
  }
  return new Date(ts).toLocaleDateString("zh-CN");
}

export function formatMoney(value) {
  const amount = Number(value || 0);
  if (!Number.isFinite(amount)) {
    return "-";
  }
  return amount >= 1000 ? `¥${amount.toLocaleString("zh-CN")}` : `¥${amount}`;
}

export function formatPercent(value, digits = 1) {
  const amount = Number(value || 0);
  if (!Number.isFinite(amount)) {
    return "-";
  }
  return `${(amount * 100).toFixed(digits)}%`;
}

export function formatScore(value) {
  const amount = Number(value || 0);
  if (!Number.isFinite(amount) || amount <= 0) {
    return "-";
  }
  return amount.toFixed(1);
}

export function formatAttachmentSize(value) {
  const size = Number(value || 0);
  if (!Number.isFinite(size) || size <= 0) {
    return "-";
  }
  if (size >= 1024 * 1024) {
    return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  }
  if (size >= 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  }
  return `${size} B`;
}

export function mapRiskLevel(value) {
  const source = String(value || "").toUpperCase();
  if (source === "LOW") {
    return "低风险";
  }
  if (source === "MEDIUM") {
    return "中风险";
  }
  if (source === "HIGH") {
    return "高风险";
  }
  return value || "待确认";
}

export function mapTrendClass(value) {
  const numeric = Number(value);
  if (Number.isFinite(numeric)) {
    if (numeric > 0) {
      return "h5-risk-up";
    }
    if (numeric < 0) {
      return "h5-risk-down";
    }
  }
  const source = String(value || "").toUpperCase();
  if (["LONG", "UP", "RISE", "BULLISH"].includes(source)) {
    return "h5-risk-up";
  }
  if (["SHORT", "DOWN", "FALL", "BEARISH"].includes(source)) {
    return "h5-risk-down";
  }
  return "h5-risk-flat";
}

export function mapVipTone(quota) {
  return resolveVipStage(quota) ? "success" : "gold";
}

export function resolveVipStage(quota) {
  const activationState = String(quota?.activation_state || "").toUpperCase();
  if (activationState) {
    return activationState === "ACTIVE";
  }
  const vipStatus = String(quota?.vip_status || "").toUpperCase();
  if (vipStatus === "ACTIVE") {
    return true;
  }
  const memberLevel = String(quota?.member_level || "").toUpperCase();
  if (!memberLevel.startsWith("VIP")) {
    return false;
  }
  const remainingDays = Number(quota?.vip_remaining_days);
  return !Number.isFinite(remainingDays) || remainingDays > 0;
}

export function normalizeText(value, fallback = "") {
  const text = String(value || "").replace(/\s+/g, " ").trim();
  return text || fallback;
}

export function truncateText(value, maxLength = 60) {
  const text = normalizeText(value);
  if (!text || text.length <= maxLength) {
    return text;
  }
  return `${text.slice(0, maxLength)}...`;
}

export function containsHTMLTag(value) {
  return /<\/?[a-z][\s\S]*>/i.test(String(value || ""));
}

export function escapeHTML(value) {
  return String(value || "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/\"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

export function sanitizeArticleHTML(rawHTML) {
  if (typeof window === "undefined" || typeof DOMParser === "undefined") {
    return String(rawHTML || "")
      .replace(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, "")
      .replace(/<style[\s\S]*?>[\s\S]*?<\/style>/gi, "");
  }
  const parser = new DOMParser();
  const doc = parser.parseFromString(String(rawHTML || ""), "text/html");
  doc.querySelectorAll("script,style,iframe,object,embed,link,meta,form,input,button,textarea,select").forEach((node) => node.remove());
  doc.querySelectorAll("*").forEach((element) => {
    Array.from(element.attributes).forEach((attr) => {
      const attrName = String(attr.name || "").toLowerCase();
      const attrValue = String(attr.value || "").trim().toLowerCase();
      if (attrName.startsWith("on")) {
        element.removeAttribute(attr.name);
        return;
      }
      if (["href", "src", "xlink:href"].includes(attrName)) {
        const safe = attrValue.startsWith("http://") || attrValue.startsWith("https://") || attrValue.startsWith("/") || attrValue.startsWith("#") || attrValue.startsWith("mailto:");
        if (!safe) {
          element.removeAttribute(attr.name);
        }
      }
    });
  });
  return doc.body.innerHTML || "";
}

export function renderArticleHTML(content) {
  const text = String(content || "").trim();
  if (!text) {
    return "<p>-</p>";
  }
  if (containsHTMLTag(text)) {
    return sanitizeArticleHTML(text);
  }
  return `<p>${escapeHTML(text).replace(/\r?\n/g, "<br/>")}</p>`;
}

export function toPlainText(value) {
  return String(value || "")
    .replace(/<[^>]*>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}
