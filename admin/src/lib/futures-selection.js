export const futuresSelectionStatusOptions = [
  { label: "启用中", value: "ACTIVE" },
  { label: "已停用", value: "INACTIVE" }
];

export const futuresSelectionRunStatusOptions = [
  { label: "运行中", value: "RUNNING" },
  { label: "运行成功", value: "SUCCEEDED" },
  { label: "运行失败", value: "FAILED" }
];

export const futuresSelectionReviewStatusOptions = [
  { label: "待审核", value: "PENDING" },
  { label: "已通过", value: "APPROVED" },
  { label: "已驳回", value: "REJECTED" }
];

export const futuresSelectionStyleOptions = [
  { label: "均衡研究", value: "balanced" },
  { label: "趋势优先", value: "trend" }
];

export const futuresSelectionContractScopeOptions = [
  { label: "全市场主力合约", value: "DOMINANT_ALL" },
  { label: "股指期货优先", value: "INDEX_FUTURES" },
  { label: "商品期货优先", value: "COMMODITY_FUTURES" },
  { label: "手工指定合约", value: "MANUAL" }
];

export const futuresSelectionRiskLevelOptions = [
  { label: "低风险", value: "LOW" },
  { label: "中风险", value: "MEDIUM" },
  { label: "高风险", value: "HIGH" }
];

export const futuresSelectionMarketRegimeOptions = [
  { label: "基准状态", value: "BASE" },
  { label: "趋势延续", value: "TREND_CONTINUE" },
  { label: "政策利多", value: "POLICY_POSITIVE" },
  { label: "政策利空", value: "POLICY_NEGATIVE" },
  { label: "供给冲击", value: "SUPPLY_SHOCK" },
  { label: "流动性冲击", value: "LIQUIDITY_SHOCK" }
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
  LOW: "低风险",
  MEDIUM: "中风险",
  HIGH: "高风险",
  LONG: "做多",
  SHORT: "做空",
  NEUTRAL: "中性",
  CORE: "核心",
  SATELLITE: "卫星",
  WATCHLIST: "观察",
  DONE: "已生成",
  BASE: "基准状态",
  TREND_CONTINUE: "趋势延续",
  POLICY_POSITIVE: "政策利多",
  POLICY_NEGATIVE: "政策利空",
  SUPPLY_SHOCK: "供给冲击",
  LIQUIDITY_SHOCK: "流动性冲击",
  MARKET_REGIME: "市场状态",
  UNIVERSE: "合约池",
  GRAPH_ENRICHMENT: "图谱增强",
  CANDIDATE_POOL: "候选池",
  PORTFOLIO: "最终组合",
  REVIEW_PAYLOAD: "审核发布载荷",
  FORWARD_EVALUATION: "前瞻评估",
  MEMORY_FEEDBACK: "记忆反馈",
  BALANCED: "均衡研究",
  TREND: "趋势优先",
  DOMINANT_ALL: "全市场主力合约",
  INDEX_FUTURES: "股指期货优先",
  COMMODITY_FUTURES: "商品期货优先",
  MANUAL: "手工指定合约"
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

export function formatFuturesSelectionLabel(value) {
  const normalized = String(value || "").trim().toUpperCase();
  if (!normalized) {
    return "-";
  }
  return labelMap[normalized] || value;
}

export function formatFuturesSelectionStatus(value) {
  return formatFuturesSelectionLabel(value);
}

export function formatFuturesSelectionRunStatus(value) {
  return formatFuturesSelectionLabel(value);
}

export function formatFuturesSelectionReviewStatus(value) {
  return formatFuturesSelectionLabel(value);
}

export function formatFuturesSelectionMarketRegime(value) {
  return formatFuturesSelectionLabel(value);
}

export function formatFuturesSelectionStage(value) {
  return formatFuturesSelectionLabel(value);
}

export function formatFuturesSelectionRiskLevel(value) {
  return formatFuturesSelectionLabel(value);
}

export function formatFuturesSelectionDirection(value) {
  return formatFuturesSelectionLabel(value);
}

export function formatFuturesSelectionStyle(value) {
  return formatFuturesSelectionLabel(value);
}

export function formatFuturesSelectionContractScope(value) {
  return formatFuturesSelectionLabel(value);
}

export function formatFuturesSelectionGraphEntityType(value) {
  const normalized = normalizeGraphKey(value);
  return graphEntityTypeMap[normalized] || value || "-";
}

export function formatFuturesSelectionGraphRelationType(value) {
  const normalized = normalizeGraphKey(value);
  return graphRelationTypeMap[normalized] || value || "-";
}

export function formatFuturesSelectionAssetDomain(value) {
  const normalized = normalizeGraphKey(value);
  return assetDomainMap[normalized] || value || "-";
}

export function formatFuturesSelectionGraphEntityKey(value) {
  const text = String(value || "").trim();
  if (!text) {
    return "-";
  }
  if (text.startsWith("FUTURES_REGIME:")) {
    return `${formatFuturesSelectionMarketRegime(text.slice("FUTURES_REGIME:".length))} 期货状态`;
  }
  if (text.startsWith("THEME:")) {
    return text.slice("THEME:".length);
  }
  if (text.startsWith("SPREAD:")) {
    return text.slice("SPREAD:".length);
  }
  return text;
}

export function formatFuturesSelectionEvaluationStatus(value) {
  const normalized = normalizeGraphKey(value);
  return evaluationStatusMap[normalized] || value || "-";
}

export function formatFuturesSelectionEvaluationScope(value) {
  const normalized = normalizeGraphKey(value);
  return evaluationScopeMap[normalized] || value || "-";
}

export function formatFuturesSelectionGraphWriteStatus(value) {
  const normalized = normalizeGraphKey(value);
  return graphWriteStatusMap[normalized] || value || "-";
}

export function formatFuturesSelectionSource(value) {
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

export function formatFuturesSelectionStageDetail(value) {
  let text = String(value || "").trim();
  if (!text) {
    return "-";
  }
  text = text.replaceAll("balanced", "均衡研究");
  text = text.replaceAll("trend", "趋势优先");
  text = text.replaceAll("DOMINANT_ALL", "全市场主力合约");
  text = text.replaceAll("INDEX_FUTURES", "股指期货优先");
  text = text.replaceAll("COMMODITY_FUTURES", "商品期货优先");
  text = text.replaceAll("MANUAL", "手工指定合约");
  text = text.replaceAll("FUTURES_REGIME:BASE", "基准状态 期货状态");
  text = text.replaceAll("FUTURES_REGIME:TREND_CONTINUE", "趋势延续 期货状态");
  text = text.replaceAll("FUTURES_REGIME:POLICY_POSITIVE", "政策利多 期货状态");
  text = text.replaceAll("FUTURES_REGIME:POLICY_NEGATIVE", "政策利空 期货状态");
  text = text.replaceAll("FUTURES_REGIME:SUPPLY_SHOCK", "供给冲击 期货状态");
  text = text.replaceAll("FUTURES_REGIME:LIQUIDITY_SHOCK", "流动性冲击 期货状态");
  Object.entries(sourceLabelMap)
    .filter(([key]) => key !== "AUTO")
    .forEach(([key, label]) => {
      text = text.replaceAll(key, label);
    });
  return text;
}

export function joinFuturesSelectionContracts(value) {
  if (!Array.isArray(value)) {
    return "";
  }
  return value.filter(Boolean).join("\n");
}

export function splitFuturesSelectionContracts(text) {
  return Array.from(
    new Set(
      String(text || "")
        .split(/[\n,，\s]+/)
        .map((item) => item.trim().toUpperCase())
        .filter(Boolean)
    )
  );
}

export function averageFuturesSelectionMetric(items = [], picker) {
  const values = (Array.isArray(items) ? items : [])
    .map((item) => Number(picker(item)))
    .filter((value) => Number.isFinite(value));
  if (values.length === 0) {
    return null;
  }
  return values.reduce((sum, value) => sum + value, 0) / values.length;
}

export function formatFuturesSelectionPercent(value, digits = 2) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return "-";
  }
  return `${(numeric * 100).toFixed(digits)}%`;
}

export function formatFuturesSelectionDateTime(value) {
  const timestamp = Date.parse(value || "");
  if (Number.isNaN(timestamp)) {
    return "-";
  }
  return new Date(timestamp).toLocaleString("zh-CN", { hour12: false });
}

export function formatFuturesSelectionDiffStatus(diff = {}) {
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

export function summarizeFuturesSelectionDiff(diff = {}) {
  const status = String(diff?.status || "").trim().toUpperCase();
  if (status === "ADDED") {
    return "这只合约是相对上一版新进入当前组合/候选的标的。";
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
    return parts.join("；") || "该合约在上一版中也已出现。";
  }
  return "当前没有可用的历史版本差异说明。";
}
