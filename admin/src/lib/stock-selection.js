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
  { label: "图谱增强", value: "GRAPH_ENRICHMENT" },
  { label: "题材/事件增强", value: "THEME_EVENT" },
  { label: "种子池", value: "SEED_POOL" },
  { label: "候选池", value: "CANDIDATE_POOL" },
  { label: "最终组合", value: "PORTFOLIO" },
  { label: "观察名单", value: "WATCHLIST" },
  { label: "审核发布载荷", value: "REVIEW_PAYLOAD" },
  { label: "前瞻评估", value: "FORWARD_EVALUATION" },
  { label: "记忆反馈", value: "MEMORY_FEEDBACK" }
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
  CORE: "核心",
  SATELLITE: "卫星",
  PORTFOLIO: "组合",
  UPTREND: "上升趋势",
  ROTATION: "轮动切换",
  EVENT_DRIVEN: "事件驱动",
  DEFENSIVE: "防御修复",
  RISK_OFF: "风险回避",
  MARKET_REGIME: "市场状态",
  UNIVERSE: "股票池",
  GRAPH_ENRICHMENT: "图谱增强",
  THEME_EVENT: "题材/事件增强",
  SEED_POOL: "种子池",
  CANDIDATE_POOL: "候选池",
  PORTFOLIO: "最终组合",
  WATCHLIST: "观察名单",
  REVIEW_PAYLOAD: "审核发布载荷",
  FORWARD_EVALUATION: "前瞻评估",
  MEMORY_FEEDBACK: "记忆反馈"
};

const graphEntityTypeMap = {
  STOCK: "股票",
  FUTURESCONTRACT: "期货合约",
  COMPANY: "公司",
  INDUSTRY: "行业",
  CONCEPTTHEME: "题材",
  COMMODITY: "商品",
  INDEX: "指数",
  POLICY: "政策/市场状态",
  EVENT: "事件",
  RESEARCHREPORT: "研报",
  SUPPLYCHAINNODE: "供应链节点"
};

const graphRelationTypeMap = {
  BELONGSTO: "属于",
  CONNECTEDTOFUTURES: "联动期货",
  AFFECTEDBY: "受影响于",
  BENEFITSFROM: "受益于",
  SUPPLIESTO: "供应到",
  COMPETESWITH: "竞争于",
  CORRELATEDWITH: "相关联",
  CONFIRMSSIGNAL: "确认信号",
  WEAKENSSIGNAL: "削弱信号",
  TRIGGERSROTATION: "触发轮动"
};

const assetDomainMap = {
  STOCK: "股票",
  FUTURES: "期货",
  CROSS: "跨资产"
};

const evaluationStatusMap = {
  PENDING: "评估中",
  DONE: "已完成",
  COMPLETED: "已完成",
  FAILED: "评估失败",
  SKIPPED: "未评估"
};

const evaluationScopeMap = {
  PORTFOLIO: "最终组合",
  CANDIDATE: "候选池",
  CANDIDATE_POOL: "候选池",
  WATCHLIST: "观察名单"
};

const graphWriteStatusMap = {
  WRITTEN: "已写入",
  SKIPPED: "已跳过",
  FAILED: "写入失败"
};

const sourceLabelMap = {
  AUTO: "自动回源链路",
  TUSHARE: "Tushare",
  AKSHARE: "AkShare",
  TICKERMD: "TickerMD",
  MYSELF: "自建回源",
  SINA: "新浪",
  TENCENT: "腾讯",
  MOCK: "模拟数据"
};

function normalizeGraphKey(value) {
  return String(value || "")
    .trim()
    .toUpperCase()
    .replace(/[^A-Z0-9]/g, "");
}

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

export function formatStockSelectionGraphEntityType(value) {
  const normalized = normalizeGraphKey(value);
  return graphEntityTypeMap[normalized] || value || "-";
}

export function formatStockSelectionGraphRelationType(value) {
  const normalized = normalizeGraphKey(value);
  return graphRelationTypeMap[normalized] || value || "-";
}

export function formatStockSelectionAssetDomain(value) {
  const normalized = normalizeGraphKey(value);
  return assetDomainMap[normalized] || value || "-";
}

export function formatStockSelectionGraphEntityKey(value) {
  const text = String(value || "").trim();
  if (!text) {
    return "-";
  }
  if (text.startsWith("REGIME:")) {
    return `${formatStockSelectionMarketRegime(text.slice("REGIME:".length))} 市场状态`;
  }
  if (text.startsWith("THEME:")) {
    return text.slice("THEME:".length);
  }
  if (text.startsWith("SECTOR:")) {
    return text.slice("SECTOR:".length);
  }
  return text;
}

export function formatStockSelectionEvaluationStatus(value) {
  const normalized = normalizeGraphKey(value);
  return evaluationStatusMap[normalized] || value || "-";
}

export function formatStockSelectionEvaluationScope(value) {
  const normalized = normalizeGraphKey(value);
  return evaluationScopeMap[normalized] || value || "-";
}

export function formatStockSelectionGraphWriteStatus(value) {
  const normalized = normalizeGraphKey(value);
  return graphWriteStatusMap[normalized] || value || "-";
}

export function formatStockSelectionSource(value) {
  const text = String(value || "").trim();
  if (!text) {
    return "-";
  }
  const tokens = text
    .split(/\s*(?:,|->|→)\s*/u)
    .map((item) => String(item || "").trim().toUpperCase())
    .filter(Boolean);
  if (!tokens.length) {
    return text;
  }
  return tokens
    .map((item) => sourceLabelMap[item] || item)
    .join(" → ");
}

export function formatStockSelectionStageDetail(value) {
  let text = String(value || "").trim();
  if (!text) {
    return "-";
  }
  text = text.replaceAll("CN_A_ALL", "A股全市场");
  text = text.replaceAll("CN_A_MAIN", "A股主板");
  text = text.replaceAll("REGIME:UPTREND", "上升趋势 市场状态");
  text = text.replaceAll("REGIME:ROTATION", "轮动切换 市场状态");
  text = text.replaceAll("REGIME:EVENT_DRIVEN", "事件驱动 市场状态");
  text = text.replaceAll("REGIME:DEFENSIVE", "防御修复 市场状态");
  text = text.replaceAll("REGIME:RISK_OFF", "风险回避 市场状态");
  Object.entries(sourceLabelMap)
    .filter(([key]) => key !== "AUTO")
    .forEach(([key, label]) => {
      text = text.replaceAll(key, label);
    });
  return text;
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
