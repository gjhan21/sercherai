function toText(value) {
  return String(value || "").trim();
}

const exactTextMap = new Map([
  ["Forecast L3", "深度推演"],
  ["Forecast L3 Demo", "深度推演示例"],
  ["NOT FOUND", "记录失效"],
  ["QUEUED", "排队中"],
  ["RUNNING", "推演中"],
  ["FAILED", "未完成"],
  ["STEP", "阶段"],
  ["LOCAL_SYNTHESIS", "本地综合推演引擎"],
  ["forecast run not found", "深推演记录不存在或已失效"],
  ["from recommendations focused handoff", "来自每日推荐的聚焦承接"],
  ["from strategies focused handoff", "来自交易策略的聚焦承接"],
  ["from forecast-lab focused handoff", "来自深度推演工作台的聚焦承接"],
  ["Keep position sizing disciplined and wait for confirmation.", "控制仓位，等待确认信号后再行动。"],
  ["Add only after confirmation.", "仅在确认信号出现后再加仓。"],
  ["Hold and verify.", "先持有，并持续验证主逻辑。"],
  ["Reduce and reassess.", "降低仓位，并重新评估交易设定。"],
  ["Main thesis still needs confirmation from flow and events.", "核心逻辑仍需等待资金面与事件面的进一步确认。"],
  ["If the risk boundary breaks, the current setup is invalidated.", "一旦风险边界被击穿，当前交易设定即告失效。"],
  ["Primary scenario confirmation", "主情景确认"],
  ["Wait for the main evidence chain to confirm.", "等待主证据链进一步确认。"],
  ["wait for confirmation.", "等待确认信号。"],
  ["Wait for confirmation.", "等待确认信号。"],
  ["Confirm price, flow and event alignment.", "确认价格、资金与事件三条线共振。"],
  ["Primary evidence chain fails to confirm.", "主证据链未能得到确认。"],
  ["trend remains intact.", "趋势主线仍然完整。"],
  ["Trend remains intact.", "趋势主线仍然完整。"],
  ["Follow the trend with tighter risk control.", "控制风险的前提下顺势跟踪。"],
  ["Observe and confirm.", "先观察，再确认。"],
  ["Reduce exposure quickly.", "快速降低仓位暴露。"],
  ["Wait for confirmation from spread, inventory and flow.", "等待价差、库存与资金流三条线进一步确认。"],
  ["Failure of the main evidence chain can force a fast reversal.", "主证据链一旦失效，价格可能快速反转。"],
  ["Track whether this signal remains aligned with the thesis.", "持续跟踪该信号是否仍与核心逻辑保持一致。"],
  ["Compare post-publish performance with current setup.", "将发布后的表现与当前推演设定持续对照。"],
  ["Evaluation feedback", "复盘反馈"],
  ["Review score", "复盘评分"],
  ["review score", "复盘评分"],
  ["history comparison loaded", "历史对比数据已加载"],
  ["history review loaded", "历史复盘数据已加载"],
  ["history runs loaded", "历史运行数据已加载"],
  ["Risk boundary: ", "风险边界："],
  ["target context loaded", "标的上下文已加载"],
  ["research pack assembled", "研究资料包已组装完成"],
  ["local synthesis completed", "本地综合推演已完成"],
  ["local synthesis finished", "本地综合推演已完成"],
  ["structured report built", "结构化报告已生成"],
  ["report persisted", "报告已保存"],
  ["publish history and explanation loaded", "发布历史与解释数据已加载"],
  ["SUCCESS", "已完成"],
  ["WATCH", "观察中"],
  ["READY", "已就绪"],
  ["PASS", "已满足"],
  ["DONE", "已完成"],
  ["PENDING", "待观察"],
  ["CONSTRUCTIVE", "偏积极"],
  ["NEUTRAL", "中性"],
  ["CAUTION", "谨慎"],
  ["BULLISH", "看多"],
  ["BEARISH", "看空"],
  ["INDUSTRY", "行业面"],
  ["FLOW", "资金面"],
  ["EVENT", "事件面"],
  ["MACRO", "宏观面"],
  ["RISK", "风险面"],
  ["FUNDAMENTAL", "基本面"],
  ["TECHNICAL", "技术面"],
  ["VALUATION", "估值面"],
  ["SUPPLY_DEMAND", "供需库存"],
  ["TERM_STRUCTURE", "期限结构"],
  ["TAPE_TECHNICAL", "盘面技术"],
  ["POSITION_FLOW", "持仓资金"],
  ["MACRO_EVENT", "宏观事件"],
  ["PARTIAL", "部分上下文"],
  ["FULL", "完整上下文"],
  ["SKIPPED", "未触发模型复核"],
  ["PENDING", "等待模型复核"],
  ["COMPLETED", "模型复核已完成"],
  ["DEGRADED", "模型复核降级完成"],
  ["UNAVAILABLE", "模型复核暂不可用"],
  ["MEDIUM", "中"],
  ["LOW", "低"],
  ["HIGH", "高"]
]);

const regexTextMap = [
  [/\bAlternative Scenarios\b/g, "备选情景"],
  [/\bExecutive summary:\s*/g, "执行摘要："],
  [/\bAction:\s*/g, "动作建议："],
  [/\bbull\b/g, "乐观情景"],
  [/\bbase\b/g, "基准情景"],
  [/\bbear\b/g, "悲观情景"],
  [/深度推演 L3 报告/g, "深度推演报告"],
  [/核心主线推演:\s*乐观情景/g, "核心主线推演：乐观情景"],
  [/核心主线推演:\s*基准情景/g, "核心主线推演：基准情景"],
  [/核心主线推演:\s*悲观情景/g, "核心主线推演：悲观情景"]
];

export function localizeForecastText(value) {
  let text = toText(value);
  if (!text) return "";
  for (const [from, to] of exactTextMap.entries()) {
    text = text.split(from).join(to);
  }
  for (const [pattern, replacement] of regexTextMap) {
    text = text.replace(pattern, replacement);
  }
  return text;
}

export function localizeForecastEngine(engineKey) {
  return localizeForecastText(engineKey) || "-";
}

export function localizeForecastStepKey(stepKey) {
  const key = toText(stepKey).toUpperCase();
  const mapping = {
    LOAD_CONTEXT: "加载标的上下文",
    BUILD_RESEARCH_PACK: "组装研究资料",
    RUN_DEEP_FORECAST: "运行深度推演",
    BUILD_REPORT: "生成结构化报告",
    PERSIST_REPORT: "保存报告"
  };
  return mapping[key] || localizeForecastText(stepKey) || "阶段";
}

export function localizeForecastStatusText(status) {
  const key = toText(status).toUpperCase();
  const mapping = {
    SUCCESS: "已完成",
    SUCCEEDED: "已完成",
    FAILED: "已失败",
    RUNNING: "推演中",
    QUEUED: "排队中",
    CANCELLED: "已取消"
  };
  return mapping[key] || localizeForecastText(status) || "-";
}

export function localizeForecastConfidence(value) {
  return localizeForecastText(value) || "";
}

export function localizeForecastScenarioName(value) {
  const key = toText(value).toLowerCase();
  const mapping = {
    bull: "乐观情景",
    base: "基准情景",
    bear: "悲观情景"
  };
  return mapping[key] || localizeForecastText(value) || "-";
}

export function localizeForecastChecklistStatus(value) {
  const key = toText(value).toUpperCase();
  const mapping = {
    READY: "已就绪",
    WATCH: "观察中",
    PASS: "已满足",
    DONE: "已完成",
    FAILED: "未满足",
    FAIL: "未满足",
    PENDING: "待观察"
  };
  return mapping[key] || localizeForecastText(value) || "-";
}

export function localizeForecastProbability(value) {
  if (value === null || typeof value === "undefined" || value === "") return "-";
  const num = Number(value);
  if (!Number.isFinite(num)) return localizeForecastText(value) || "-";
  const normalized = num <= 1 ? num * 100 : num;
  return `${Math.round(normalized)}%`;
}

export function localizeForecastDimension(value) {
  return localizeForecastText(value) || "核心维度";
}

export function localizeForecastSource(value) {
  const key = toText(value).toUpperCase();
  const mapping = {
    RECOMMENDATION: "每日推荐",
    STRATEGY: "交易策略",
    ADMIN_CONSOLE: "管理后台",
    USER_REQUEST: "用户主动发起",
    ADMIN_MANUAL: "人工触发",
    AUTO_PRIORITY: "系统优先级触发"
  };
  return mapping[key] || localizeForecastText(value) || "-";
}

export function localizeForecastContextQuality(value) {
  const key = toText(value).toUpperCase();
  const mapping = {
    FULL: "完整上下文",
    PARTIAL: "部分上下文"
  };
  return mapping[key] || localizeForecastText(value) || "-";
}

export function localizeForecastValidationStatus(value) {
  const key = toText(value).toUpperCase();
  const mapping = {
    SKIPPED: "未触发模型复核",
    PENDING: "等待模型复核",
    COMPLETED: "模型复核已完成",
    DEGRADED: "模型复核降级完成",
    UNAVAILABLE: "模型复核暂不可用"
  };
  return mapping[key] || localizeForecastText(value) || "-";
}

export function localizeForecastHistoryChangeLabel(value) {
  const key = toText(value).toUpperCase();
  const mapping = {
    ADDED: "新增",
    REMOVED: "移除",
    CHANGED: "变化",
    UPDATED: "变化",
    STABLE: "稳定",
    "基本不变": "基本不变",
    "新增": "新增",
    "弱化": "弱化",
    "变化": "变化",
    "未提供": "未提供"
  };
  return mapping[key] || localizeForecastText(value) || "变化";
}

export function localizeForecastReviewGrade(value) {
  const key = toText(value).toUpperCase();
  const mapping = {
    A: "A级",
    B: "B级",
    C: "C级",
    D: "D级"
  };
  return mapping[key] || localizeForecastText(value) || "-";
}
