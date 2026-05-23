import {
  localizeForecastProbability,
  localizeForecastText
} from "./forecast-localization.js";

const STOCK_DIMENSION_ORDER = ["FUNDAMENTAL", "TECHNICAL", "FLOW", "VALUATION", "EVENT"];
const FUTURES_DIMENSION_ORDER = ["SUPPLY_DEMAND", "TERM_STRUCTURE", "TAPE_TECHNICAL", "POSITION_FLOW", "MACRO_EVENT"];

const DIMENSION_META = {
  FUNDAMENTAL: { label: "基本面", targetType: "STOCK" },
  TECHNICAL: { label: "技术面", targetType: "STOCK" },
  FLOW: { label: "资金面", targetType: "STOCK" },
  VALUATION: { label: "估值面", targetType: "STOCK" },
  EVENT: { label: "事件面", targetType: "STOCK" },
  SUPPLY_DEMAND: { label: "供需库存", targetType: "FUTURES" },
  TERM_STRUCTURE: { label: "期限结构", targetType: "FUTURES" },
  TAPE_TECHNICAL: { label: "盘面技术", targetType: "FUTURES" },
  POSITION_FLOW: { label: "持仓资金", targetType: "FUTURES" },
  MACRO_EVENT: { label: "宏观事件", targetType: "FUTURES" }
};

const STANCE_LABELS = {
  BULLISH: "偏多",
  BEARISH: "偏空",
  CONSTRUCTIVE: "偏积极",
  WATCH: "等待确认",
  CAUTION: "谨慎",
  NEUTRAL: "中性",
  POSITIVE: "偏积极",
  NEGATIVE: "偏谨慎"
};

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

function toText(value) {
  return localizeForecastText(value).trim();
}

function toDimensionKey(value) {
  return String(value || "").trim().toUpperCase();
}

function inferTargetType(targetType, items) {
  const normalizedTargetType = String(targetType || "").trim().toUpperCase();
  if (normalizedTargetType === "STOCK" || normalizedTargetType === "FUTURES") return normalizedTargetType;
  const firstKnown = items.find((item) => DIMENSION_META[item.dimension]?.targetType);
  return firstKnown ? DIMENSION_META[firstKnown.dimension].targetType : "";
}

function localizeEvidenceStance(value) {
  const key = String(value || "").trim().toUpperCase();
  return STANCE_LABELS[key] || toText(value) || "等待更多判断";
}

function formatEvidenceList(items, fallback) {
  const normalized = toArray(items).map((item) => toText(item)).filter(Boolean);
  return normalized.length ? normalized.join("；") : fallback;
}

function normalizeEvidenceItem(item) {
  const dimension = toDimensionKey(item?.dimension);
  return {
    dimension,
    label: DIMENSION_META[dimension]?.label || toText(item?.dimension) || "核心维度",
    summary: toText(item?.summary) || "当前未补更多维度摘要。",
    stanceLabel: localizeEvidenceStance(item?.stance),
    confidenceLabel: localizeForecastProbability(item?.confidence) || "-",
    supportingText: formatEvidenceList(item?.supporting_points, "等待更多支撑点。"),
    riskText: formatEvidenceList(item?.risk_points, "等待更多风险点。")
  };
}

export function buildForecastEvidenceSections({ targetType, dimensionEvidence } = {}) {
  const normalizedItems = toArray(dimensionEvidence).map(normalizeEvidenceItem).filter((item) => item.dimension);
  if (!normalizedItems.length) return [];

  const resolvedTargetType = inferTargetType(targetType, normalizedItems);
  const preferredOrder =
    resolvedTargetType === "STOCK"
      ? STOCK_DIMENSION_ORDER
      : resolvedTargetType === "FUTURES"
        ? FUTURES_DIMENSION_ORDER
        : normalizedItems.map((item) => item.dimension);

  const byDimension = new Map(normalizedItems.map((item) => [item.dimension, item]));
  const sections = [];

  preferredOrder.forEach((dimension) => {
    if (!byDimension.has(dimension)) return;
    const item = byDimension.get(dimension);
    sections.push({
      key: dimension,
      ...item
    });
  });

  normalizedItems.forEach((item) => {
    if (sections.some((section) => section.key === item.dimension)) return;
    sections.push({
      key: item.dimension,
      ...item
    });
  });

  return sections;
}
