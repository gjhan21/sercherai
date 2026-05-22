/**
 * Finance-specific formatting and parsing utilities
 */

export function formatScore(value) {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  return num.toFixed(1);
}

export function formatRate(num) {
  if (num === null || num === undefined || num === "" || Number.isNaN(num)) {
    return "-";
  }
  return `${(num * 100).toFixed(1)}%`;
}

export function formatPercent(value) {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  const percent = (num * 100).toFixed(2);
  if (num > 0) {
    return `+${percent}%`;
  }
  return `${percent}%`;
}

export function formatDate(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return value || "-";
  }
  return new Date(ts).toLocaleDateString("zh-CN");
}

export function formatDateRange(start, end) {
  const startText = formatDate(start);
  const endText = formatDate(end);
  if (startText === "-" && endText === "-") {
    return "-";
  }
  if (endText === "-") {
    return `${startText} 起`;
  }
  if (startText === "-") {
    return `截至 ${endText}`;
  }
  return `${startText} ~ ${endText}`;
}

export function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) return "-";
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

export function resolveVIPStage(quota) {
  const activationState = String(quota?.activation_state || "").toUpperCase();
  if (activationState) {
    return activationState === "ACTIVE";
  }
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

export function trendClassByNumber(num) {
  if (!Number.isFinite(num) || num === 0) {
    return "";
  }
  return num > 0 ? "up" : "down";
}

export function mapRisk(level) {
  const l = String(level || "").toUpperCase();
  if (l === "HIGH") return "高风险";
  if (l === "LOW") return "低风险";
  return "中风险";
}

export function mapStatus(status) {
  const s = String(status || "").toUpperCase();
  if (s === "PUBLISHED") return { label: "已发布", className: "good" };
  if (s === "ARCHIVED") return { label: "已存档", className: "normal" };
  return { label: s || "未知", className: "watch" };
}

export function compareDateAsc(a, b) {
  const tsa = Date.parse(a || "");
  const tsb = Date.parse(b || "");
  if (Number.isNaN(tsa)) return 1;
  if (Number.isNaN(tsb)) return -1;
  return tsa - tsb;
}

export function parseErrorMessage(error) {
  if (!error) {
    return "unknown error";
  }
  const responseMessage =
    error?.response?.data?.message || error?.response?.data?.error || error?.response?.statusText;
  return responseMessage || error?.message || "unknown error";
}
