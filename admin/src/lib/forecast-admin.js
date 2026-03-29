export const DEFAULT_FORECAST_ADMIN_CONFIG = {
  enabled: true,
  explanationEnabled: true,
  memoryFeedbackMinSamples: 5,
  advisoryPriorityThreshold: 0.55,
  l2Enabled: true,
  relationshipSnapshotEnabled: true,
  stableScenariosEnabled: true,
  vetoConfidenceThreshold: 0.35,
  l3Enabled: false,
  l3AdminManualEnabled: true,
  l3UserRequestEnabled: false,
  l3AutoPriorityEnabled: false,
  l3ClientReadEnabled: true,
  l3RequireVipForFullReport: true,
  l3MaxActiveRuns: 2,
  l3MaxRunsPerDay: 24,
  l3MaxUserRunsPerDay: 1,
  l3MinPriorityThreshold: 0.7,
  l3DispatchEnabled: true,
  l3DispatchIntervalMinutes: 5,
  l3QualityEnabled: true,
  l3QualityIntervalMinutes: 60,
  l3DefaultEngineKey: "LOCAL_SYNTHESIS"
};

function parseConfigBool(raw, fallback) {
  const text = String(raw ?? "").trim().toLowerCase();
  if (!text) {
    return fallback;
  }
  if (["1", "true", "yes", "on", "y"].includes(text)) {
    return true;
  }
  if (["0", "false", "no", "off", "n"].includes(text)) {
    return false;
  }
  return fallback;
}

function parseConfigInt(raw, fallback, min, max) {
  const parsed = Number.parseInt(String(raw ?? "").trim(), 10);
  if (!Number.isFinite(parsed)) {
    return fallback;
  }
  return Math.max(min, Math.min(max, parsed));
}

function parseConfigFloat(raw, fallback, min, max) {
  const parsed = Number.parseFloat(String(raw ?? "").trim());
  if (!Number.isFinite(parsed)) {
    return fallback;
  }
  return Math.max(min, Math.min(max, parsed));
}

function parseConfigText(raw, fallback = "") {
  const text = String(raw ?? "").trim();
  return text || fallback;
}

function normalizeObject(value) {
  return value && typeof value === "object" && !Array.isArray(value) ? value : {};
}

function collectExplanationNodes(raw, results) {
  if (Array.isArray(raw)) {
    raw.forEach((item) => collectExplanationNodes(item, results));
    return;
  }
  if (!raw || typeof raw !== "object") {
    return;
  }
  const value = normalizeObject(raw);
  const hasResearch = Array.isArray(value.research_outline) && value.research_outline.length > 0;
  const hasActive = Array.isArray(value.active_thesis_cards) && value.active_thesis_cards.length > 0;
  const hasHistorical = Array.isArray(value.historical_thesis_cards) && value.historical_thesis_cards.length > 0;
  const hasWatch = Array.isArray(value.watch_signals) && value.watch_signals.length > 0;
  const memory = normalizeObject(value.memory_feedback);
  const hasMemory = Boolean(
    String(memory.summary || "").trim() ||
      (Array.isArray(memory.suggestions) && memory.suggestions.length > 0) ||
      (Array.isArray(memory.failure_signals) && memory.failure_signals.length > 0)
  );
  const calibration = normalizeObject(value.confidence_calibration);
  const hasCalibration =
    Number.isFinite(Number(calibration.adjusted_confidence)) || Array.isArray(calibration.drivers);
  const hasL2 =
    (value.relationship_snapshot && typeof value.relationship_snapshot === "object") ||
    (Array.isArray(value.scenario_snapshots) && value.scenario_snapshots.length > 0) ||
    (value.scenario_meta && typeof value.scenario_meta === "object") ||
    (Array.isArray(value.agent_opinions) && value.agent_opinions.length > 0);

  if (hasResearch || hasActive || hasHistorical || hasWatch || hasMemory || hasCalibration || hasL2) {
    results.push(value);
  }

  Object.values(value).forEach((item) => collectExplanationNodes(item, results));
}

export function parseForecastAdminConfigMap(map) {
  const source = normalizeObject(map);
  return {
    enabled: parseConfigBool(source["growth.forecast_l1.enabled"], DEFAULT_FORECAST_ADMIN_CONFIG.enabled),
    explanationEnabled: parseConfigBool(
      source["growth.forecast_l1.explanation_enabled"],
      DEFAULT_FORECAST_ADMIN_CONFIG.explanationEnabled
    ),
    memoryFeedbackMinSamples: parseConfigInt(
      source["growth.forecast_l1.memory_feedback_min_samples"],
      DEFAULT_FORECAST_ADMIN_CONFIG.memoryFeedbackMinSamples,
      1,
      100
    ),
    advisoryPriorityThreshold: Number(
      parseConfigFloat(
        source["growth.forecast_l1.advisory_priority_threshold"],
        DEFAULT_FORECAST_ADMIN_CONFIG.advisoryPriorityThreshold,
        0.1,
        0.95
      ).toFixed(2)
    ),
    l2Enabled: parseConfigBool(source["growth.forecast_l2.enabled"], DEFAULT_FORECAST_ADMIN_CONFIG.l2Enabled),
    relationshipSnapshotEnabled: parseConfigBool(
      source["growth.forecast_l2.relationship_snapshot_enabled"],
      DEFAULT_FORECAST_ADMIN_CONFIG.relationshipSnapshotEnabled
    ),
    stableScenariosEnabled: parseConfigBool(
      source["growth.forecast_l2.stable_scenarios_enabled"],
      DEFAULT_FORECAST_ADMIN_CONFIG.stableScenariosEnabled
    ),
    vetoConfidenceThreshold: Number(
      parseConfigFloat(
        source["growth.forecast_l2.veto_confidence_threshold"],
        DEFAULT_FORECAST_ADMIN_CONFIG.vetoConfidenceThreshold,
        0.05,
        0.95
      ).toFixed(2)
    ),
    l3Enabled: parseConfigBool(source["growth.forecast_l3.enabled"], DEFAULT_FORECAST_ADMIN_CONFIG.l3Enabled),
    l3AdminManualEnabled: parseConfigBool(
      source["growth.forecast_l3.admin_manual_enabled"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3AdminManualEnabled
    ),
    l3UserRequestEnabled: parseConfigBool(
      source["growth.forecast_l3.user_request_enabled"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3UserRequestEnabled
    ),
    l3AutoPriorityEnabled: parseConfigBool(
      source["growth.forecast_l3.auto_priority_enabled"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3AutoPriorityEnabled
    ),
    l3ClientReadEnabled: parseConfigBool(
      source["growth.forecast_l3.client_read_enabled"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3ClientReadEnabled
    ),
    l3RequireVipForFullReport: parseConfigBool(
      source["growth.forecast_l3.require_vip_for_full_report"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3RequireVipForFullReport
    ),
    l3MaxActiveRuns: parseConfigInt(
      source["growth.forecast_l3.max_active_runs"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3MaxActiveRuns,
      1,
      50
    ),
    l3MaxRunsPerDay: parseConfigInt(
      source["growth.forecast_l3.max_runs_per_day"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3MaxRunsPerDay,
      1,
      500
    ),
    l3MaxUserRunsPerDay: parseConfigInt(
      source["growth.forecast_l3.max_user_runs_per_day"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3MaxUserRunsPerDay,
      1,
      20
    ),
    l3MinPriorityThreshold: Number(
      parseConfigFloat(
        source["growth.forecast_l3.min_priority_threshold"],
        DEFAULT_FORECAST_ADMIN_CONFIG.l3MinPriorityThreshold,
        0.1,
        0.99
      ).toFixed(2)
    ),
    l3DispatchEnabled: parseConfigBool(
      source["growth.forecast_l3.dispatch.enabled"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3DispatchEnabled
    ),
    l3DispatchIntervalMinutes: parseConfigInt(
      source["growth.forecast_l3.dispatch.interval_minutes"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3DispatchIntervalMinutes,
      1,
      240
    ),
    l3QualityEnabled: parseConfigBool(
      source["growth.forecast_l3.quality.enabled"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3QualityEnabled
    ),
    l3QualityIntervalMinutes: parseConfigInt(
      source["growth.forecast_l3.quality.interval_minutes"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3QualityIntervalMinutes,
      5,
      1440
    ),
    l3DefaultEngineKey: parseConfigText(
      source["growth.forecast_l3.default_engine_key"],
      DEFAULT_FORECAST_ADMIN_CONFIG.l3DefaultEngineKey
    ).toUpperCase()
  };
}

export function buildForecastAdminConfigPayloads(config) {
  const normalized = parseForecastAdminConfigMap({
    "growth.forecast_l1.enabled": config?.enabled,
    "growth.forecast_l1.explanation_enabled": config?.explanationEnabled,
    "growth.forecast_l1.memory_feedback_min_samples": config?.memoryFeedbackMinSamples,
    "growth.forecast_l1.advisory_priority_threshold": config?.advisoryPriorityThreshold,
    "growth.forecast_l2.enabled": config?.l2Enabled,
    "growth.forecast_l2.relationship_snapshot_enabled": config?.relationshipSnapshotEnabled,
    "growth.forecast_l2.stable_scenarios_enabled": config?.stableScenariosEnabled,
    "growth.forecast_l2.veto_confidence_threshold": config?.vetoConfidenceThreshold,
    "growth.forecast_l3.enabled": config?.l3Enabled,
    "growth.forecast_l3.admin_manual_enabled": config?.l3AdminManualEnabled,
    "growth.forecast_l3.user_request_enabled": config?.l3UserRequestEnabled,
    "growth.forecast_l3.auto_priority_enabled": config?.l3AutoPriorityEnabled,
    "growth.forecast_l3.client_read_enabled": config?.l3ClientReadEnabled,
    "growth.forecast_l3.require_vip_for_full_report": config?.l3RequireVipForFullReport,
    "growth.forecast_l3.max_active_runs": config?.l3MaxActiveRuns,
    "growth.forecast_l3.max_runs_per_day": config?.l3MaxRunsPerDay,
    "growth.forecast_l3.max_user_runs_per_day": config?.l3MaxUserRunsPerDay,
    "growth.forecast_l3.min_priority_threshold": config?.l3MinPriorityThreshold,
    "growth.forecast_l3.dispatch.enabled": config?.l3DispatchEnabled,
    "growth.forecast_l3.dispatch.interval_minutes": config?.l3DispatchIntervalMinutes,
    "growth.forecast_l3.quality.enabled": config?.l3QualityEnabled,
    "growth.forecast_l3.quality.interval_minutes": config?.l3QualityIntervalMinutes,
    "growth.forecast_l3.default_engine_key": config?.l3DefaultEngineKey
  });
  return [
    {
      config_key: "growth.forecast_l1.enabled",
      config_value: normalized.enabled ? "true" : "false",
      description: "股票/期货预测增强 L1 全局开关（仅影响 explanation 增强与运营展示）"
    },
    {
      config_key: "growth.forecast_l1.explanation_enabled",
      config_value: normalized.explanationEnabled ? "true" : "false",
      description: "预测增强 explanation 展示总开关（不改排序、不改发布主链）"
    },
    {
      config_key: "growth.forecast_l1.memory_feedback_min_samples",
      config_value: String(normalized.memoryFeedbackMinSamples),
      description: "memory_feedback 进入 explanation 的最小样本阈值"
    },
    {
      config_key: "growth.forecast_l1.advisory_priority_threshold",
      config_value: normalized.advisoryPriorityThreshold.toFixed(2),
      description: "advisory priority 阈值（adjusted_confidence 低于该值视为高优先级样本）"
    },
    {
      config_key: "growth.forecast_l2.enabled",
      config_value: normalized.l2Enabled ? "true" : "false",
      description: "股票/期货预测增强 L2 全局开关（仅影响 scenario / relationship / veto 摘要展示）"
    },
    {
      config_key: "growth.forecast_l2.relationship_snapshot_enabled",
      config_value: normalized.relationshipSnapshotEnabled ? "true" : "false",
      description: "relationship snapshot 展示开关（只读，不改发布主流程）"
    },
    {
      config_key: "growth.forecast_l2.stable_scenarios_enabled",
      config_value: normalized.stableScenariosEnabled ? "true" : "false",
      description: "bull/base/bear 稳定三情景展示开关（只读，不改发布主流程）"
    },
    {
      config_key: "growth.forecast_l2.veto_confidence_threshold",
      config_value: normalized.vetoConfidenceThreshold.toFixed(2),
      description: "L2 veto 置信度阈值（低于该值仅做提示，不直接替代审核决策）"
    },
    {
      config_key: "growth.forecast_l3.enabled",
      config_value: normalized.l3Enabled ? "true" : "false",
      description: "L3 深推演全局开关（异步增强层，不替代推荐主链）"
    },
    {
      config_key: "growth.forecast_l3.admin_manual_enabled",
      config_value: normalized.l3AdminManualEnabled ? "true" : "false",
      description: "允许管理员手动触发 L3 深推演"
    },
    {
      config_key: "growth.forecast_l3.user_request_enabled",
      config_value: normalized.l3UserRequestEnabled ? "true" : "false",
      description: "允许前台用户主动请求 L3 深推演"
    },
    {
      config_key: "growth.forecast_l3.auto_priority_enabled",
      config_value: normalized.l3AutoPriorityEnabled ? "true" : "false",
      description: "允许按高优先级样本自动排队 L3 深推演"
    },
    {
      config_key: "growth.forecast_l3.client_read_enabled",
      config_value: normalized.l3ClientReadEnabled ? "true" : "false",
      description: "允许客户端 explanation/history 读取 L3 摘要引用"
    },
    {
      config_key: "growth.forecast_l3.require_vip_for_full_report",
      config_value: normalized.l3RequireVipForFullReport ? "true" : "false",
      description: "是否要求 VIP 才可阅读完整 L3 报告"
    },
    {
      config_key: "growth.forecast_l3.max_active_runs",
      config_value: String(normalized.l3MaxActiveRuns),
      description: "L3 同时运行中的最大任务数"
    },
    {
      config_key: "growth.forecast_l3.max_runs_per_day",
      config_value: String(normalized.l3MaxRunsPerDay),
      description: "L3 全站每日最大运行数"
    },
    {
      config_key: "growth.forecast_l3.max_user_runs_per_day",
      config_value: String(normalized.l3MaxUserRunsPerDay),
      description: "单用户每日最大 L3 请求数"
    },
    {
      config_key: "growth.forecast_l3.min_priority_threshold",
      config_value: normalized.l3MinPriorityThreshold.toFixed(2),
      description: "自动排队进入 L3 的最小 priority 阈值"
    },
    {
      config_key: "growth.forecast_l3.dispatch.enabled",
      config_value: normalized.l3DispatchEnabled ? "true" : "false",
      description: "L3 dispatch worker 开关"
    },
    {
      config_key: "growth.forecast_l3.dispatch.interval_minutes",
      config_value: String(normalized.l3DispatchIntervalMinutes),
      description: "L3 dispatch worker 轮询间隔（分钟）"
    },
    {
      config_key: "growth.forecast_l3.quality.enabled",
      config_value: normalized.l3QualityEnabled ? "true" : "false",
      description: "L3 quality backfill worker 开关"
    },
    {
      config_key: "growth.forecast_l3.quality.interval_minutes",
      config_value: String(normalized.l3QualityIntervalMinutes),
      description: "L3 quality backfill worker 间隔（分钟）"
    },
    {
      config_key: "growth.forecast_l3.default_engine_key",
      config_value: normalized.l3DefaultEngineKey,
      description: "L3 默认深推演引擎键"
    }
  ];
}

export function buildForecastPublishSummary(detail, advisoryThreshold = DEFAULT_FORECAST_ADMIN_CONFIG.advisoryPriorityThreshold) {
  const source = normalizeObject(detail);
  const explanationNodes = [];
  collectExplanationNodes(source.publish_payloads || [], explanationNodes);
  collectExplanationNodes(source.report_snapshot || {}, explanationNodes);

  const uniqueNodes = [];
  const seen = new Set();
  explanationNodes.forEach((item) => {
    const key = JSON.stringify(item);
    if (seen.has(key)) {
      return;
    }
    seen.add(key);
    uniqueNodes.push(item);
  });

  let researchOutlineCount = 0;
  let watchSignalCount = 0;
  let thesisCardCount = 0;
  let memoryFeedbackCount = 0;
  let highAdvisoryCount = 0;
  let scenarioSnapshotCount = 0;
  let relationshipNodeCount = 0;
  let vetoedCount = 0;
  let agentOpinionCount = 0;
  const primaryScenarios = new Set();

  uniqueNodes.forEach((item) => {
    researchOutlineCount += Array.isArray(item.research_outline) ? item.research_outline.length : 0;
    watchSignalCount += Array.isArray(item.watch_signals) ? item.watch_signals.length : 0;
    thesisCardCount +=
      (Array.isArray(item.active_thesis_cards) ? item.active_thesis_cards.length : 0) +
      (Array.isArray(item.historical_thesis_cards) ? item.historical_thesis_cards.length : 0);
    const memory = normalizeObject(item.memory_feedback);
    if (
      String(memory.summary || "").trim() ||
      (Array.isArray(memory.suggestions) && memory.suggestions.length > 0) ||
      (Array.isArray(memory.failure_signals) && memory.failure_signals.length > 0)
    ) {
      memoryFeedbackCount += 1;
    }
    const calibration = normalizeObject(item.confidence_calibration);
    const adjusted = Number(calibration.adjusted_confidence);
    if (calibration.advisory_only === true && Number.isFinite(adjusted) && adjusted <= advisoryThreshold) {
      highAdvisoryCount += 1;
    }
    scenarioSnapshotCount += Array.isArray(item.scenario_snapshots) ? item.scenario_snapshots.length : 0;
    relationshipNodeCount += Array.isArray(item.relationship_snapshot?.nodes) ? item.relationship_snapshot.nodes.length : 0;
    agentOpinionCount += Array.isArray(item.agent_opinions) ? item.agent_opinions.length : 0;
    if (item.scenario_meta?.vetoed === true) {
      vetoedCount += 1;
    }
    const primary = String(item.scenario_meta?.primary_scenario || "").trim();
    if (primary) {
      primaryScenarios.add(primary);
    }
  });

  const payloadCount = Math.max(0, Number.parseInt(String(source.payload_count ?? source.selected_count ?? 0), 10) || 0);
  const enhancedCount = uniqueNodes.filter((item) => {
    return (
      (Array.isArray(item.research_outline) && item.research_outline.length > 0) ||
      (Array.isArray(item.active_thesis_cards) && item.active_thesis_cards.length > 0) ||
      (Array.isArray(item.historical_thesis_cards) && item.historical_thesis_cards.length > 0) ||
      (Array.isArray(item.watch_signals) && item.watch_signals.length > 0)
    );
  }).length;

  const ratioBase = payloadCount > 0 ? payloadCount : enhancedCount;
  const coverageRatio = ratioBase > 0 ? `${((enhancedCount / ratioBase) * 100).toFixed(1)}%` : "0.0%";

  return {
    enhancedCount,
    researchOutlineCount,
    watchSignalCount,
    thesisCardCount,
    memoryFeedbackCount,
    highAdvisoryCount,
    scenarioSnapshotCount,
    relationshipNodeCount,
    vetoedCount,
    agentOpinionCount,
    primaryScenarios: Array.from(primaryScenarios),
    payloadCount,
    coverageRatio
  };
}

export function buildForecastL2Summary(raw) {
  const nodes = [];
  collectExplanationNodes(raw, nodes);
  const explanation = nodes[0];
  if (!explanation) {
    return null;
  }
  const relationshipCount = Array.isArray(explanation.relationship_snapshot?.nodes)
    ? explanation.relationship_snapshot.nodes.length
    : Number(explanation.relationship_snapshot?.relationship_count || 0);
  const scenarioCount = Array.isArray(explanation.scenario_snapshots) ? explanation.scenario_snapshots.length : 0;
  const primaryScenario = String(explanation.scenario_meta?.primary_scenario || "").trim();
  const consensusAction = String(explanation.scenario_meta?.consensus_action || "").trim();
  const vetoed = explanation.scenario_meta?.vetoed === true;
  const vetoReason = String(explanation.scenario_meta?.veto_reason || "").trim();
  const topRoles = Array.isArray(explanation.agent_opinions)
    ? explanation.agent_opinions
        .map((item) => String(item?.role || item?.agent || "").trim())
        .filter(Boolean)
        .slice(0, 3)
    : [];
  if (!relationshipCount && !scenarioCount && !primaryScenario && !consensusAction && !topRoles.length && !vetoed) {
    return null;
  }
  return {
    relationshipCount,
    scenarioCount,
    primaryScenario,
    consensusAction,
    vetoed,
    vetoReason,
    topRoles
  };
}
