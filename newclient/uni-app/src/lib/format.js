export function formatDateTime(value) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return String(value);
  }
  return new Intl.DateTimeFormat("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
}

export function formatDate(value) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return String(value);
  }
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit"
  }).format(date);
}

export function formatPercent(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  return `${(num * 100).toFixed(2)}%`;
}

export function formatScore(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  return num.toFixed(1);
}

export function formatCount(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "0";
  }
  return String(num);
}
