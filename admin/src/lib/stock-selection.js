export const stockSelectionStatusOptions = [
  { label: "启用中", value: "ACTIVE" },
  { label: "已停用", value: "INACTIVE" }
];

export const stockSelectionRunStatusOptions = [
  { label: "运行中", value: "RUNNING" },
  { label: "运行成功", value: "SUCCEEDED" },
  { label: "运行失败", value: "FAILED" }
];

export const stockSelectionReviewStatusOptions = [
  { label: "待审核", value: "PENDING" },
  { label: "已通过", value: "APPROVED" },
  { label: "已驳回", value: "REJECTED" }
];

export const stockSelectionModeOptions = [
  { label: "自动选股", value: "AUTO" },
  { label: "手动指定", value: "MANUAL" },
  { label: "调试模式", value: "DEBUG" }
];

export const stockSelectionUniverseScopeOptions = [
  { label: "A股全市场", value: "CN_A_ALL" },
  { label: "A股主板", value: "CN_A_MAIN" }
];

export const stockSelectionRiskLevelOptions = [
  { label: "低风险", value: "LOW" },
  { label: "中风险", value: "MEDIUM" },
  { label: "高风险", value: "HIGH" }
];

export const stockSelectionMarketRegimeOptions = [
  { label: "上升趋势", value: "UPTREND" },
  { label: "轮动切换", value: "ROTATION" },
  { label: "事件驱动", value: "EVENT_DRIVEN" },
  { label: "防御修复", value: "DEFENSIVE" },
  { label: "风险回避", value: "RISK_OFF" }
];

export const stockSelectionStageOptions = [
  { label: "市场状态", value: "MARKET_REGIME" },
  { label: "股票池", value: "UNIVERSE" },
  { label: "题材/事件增强", value: "THEME_EVENT" },
  { label: "种子池", value: "SEED_POOL" },
  { label: "候选池", value: "CANDIDATE_POOL" },
  { label: "最终组合", value: "PORTFOLIO" },
  { label: "观察名单", value: "WATCHLIST" }
];

const labelMap = {
  ACTIVE: "启用中",
  INACTIVE: "已停用",
  RUNNING: "运行中",
  SUCCEEDED: "运行成功",
  FAILED: "运行失败",
  PENDING: "待审核",
  APPROVED: "已通过",
  REJECTED: "已驳回",
  AUTO: "自动选股",
  MANUAL: "手动指定",
  DEBUG: "调试模式",
  CN_A_ALL: "A股全市场",
  CN_A_MAIN: "A股主板",
  LOW: "低风险",
  MEDIUM: "中风险",
  HIGH: "高风险",
  UPTREND: "上升趋势",
  ROTATION: "轮动切换",
  EVENT_DRIVEN: "事件驱动",
  DEFENSIVE: "防御修复",
  RISK_OFF: "风险回避",
  MARKET_REGIME: "市场状态",
  UNIVERSE: "股票池",
  THEME_EVENT: "题材/事件增强",
  SEED_POOL: "种子池",
  CANDIDATE_POOL: "候选池",
  PORTFOLIO: "最终组合",
  WATCHLIST: "观察名单"
};

export function formatStockSelectionLabel(value) {
  const normalized = String(value || "").trim().toUpperCase();
  if (!normalized) {
    return "-";
  }
  return labelMap[normalized] || value;
}

export function formatStockSelectionStatus(value) {
  return formatStockSelectionLabel(value);
}

export function formatStockSelectionRunStatus(value) {
  return formatStockSelectionLabel(value);
}

export function formatStockSelectionReviewStatus(value) {
  return formatStockSelectionLabel(value);
}

export function formatStockSelectionMode(value) {
  return formatStockSelectionLabel(value);
}

export function formatStockSelectionUniverseScope(value) {
  return formatStockSelectionLabel(value);
}

export function formatStockSelectionRiskLevel(value) {
  return formatStockSelectionLabel(value);
}

export function formatStockSelectionStage(value) {
  return formatStockSelectionLabel(value);
}

export function formatStockSelectionMarketRegime(value) {
  return formatStockSelectionLabel(value);
}

export function joinStockSelectionSymbols(value) {
  if (!Array.isArray(value)) {
    return "";
  }
  return value.filter(Boolean).join("\n");
}

export function splitStockSelectionSymbols(text) {
  return Array.from(
    new Set(
      String(text || "")
        .split(/[\n,，\s]+/)
        .map((item) => item.trim().toUpperCase())
        .filter(Boolean)
      )
  );
}

export function joinStockSelectionTextList(value) {
  if (!Array.isArray(value)) {
    return "";
  }
  return value.filter(Boolean).join("\n");
}

export function splitStockSelectionTextList(text) {
  return Array.from(
    new Set(
      String(text || "")
        .split(/[\n,，]+/)
        .map((item) => item.trim())
        .filter(Boolean)
    )
  );
}

export function countExtraConfigKeys(extraConfigs = {}) {
  return Object.values(extraConfigs).reduce((total, item) => total + Object.keys(item || {}).length, 0);
}

export function formatStockSelectionPercent(value, digits = 2) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return "-";
  }
  return `${(numeric * 100).toFixed(digits)}%`;
}

export function formatStockSelectionDateTime(value) {
  const timestamp = Date.parse(value || "");
  if (Number.isNaN(timestamp)) {
    return "-";
  }
  return new Date(timestamp).toLocaleString("zh-CN", { hour12: false });
}

export function formatStockSelectionDiffStatus(diff = {}) {
  const status = String(diff?.status || "").trim().toUpperCase();
  if (status === "ADDED") {
    return "新增入选";
  }
  if (status === "REMOVED") {
    return "本次移除";
  }
  if (status === "UNCHANGED") {
    return "与上版延续";
  }
  return "暂无历史对比";
}

export function summarizeStockSelectionDiff(diff = {}) {
  const status = String(diff?.status || "").trim().toUpperCase();
  if (status === "ADDED") {
    return "这只股票是相对上一版新进入当前组合/候选的标的。";
  }
  if (status === "UNCHANGED") {
    const rank = Number(diff?.previous_rank);
    const previousReason = String(diff?.previous_reason || "").trim();
    const parts = [];
    if (Number.isFinite(rank) && rank > 0) {
      parts.push(`上一版排名第 ${rank} 位`);
    }
    if (previousReason) {
      parts.push(`上一版摘要：${previousReason}`);
    }
    return parts.join("；") || "该标的在上一版中也已出现。";
  }
  return "当前没有可用的历史版本差异说明。";
}

export function averageStockSelectionMetric(items = [], picker) {
  const values = (Array.isArray(items) ? items : [])
    .map((item) => Number(picker(item)))
    .filter((value) => Number.isFinite(value));
  if (values.length === 0) {
    return null;
  }
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}
