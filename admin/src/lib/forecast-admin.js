export const DEFAULT_FORECAST_ADMIN_CONFIG = {
  enabled: true,
  explanationEnabled: true,
  memoryFeedbackMinSamples: 5,
  advisoryPriorityThreshold: 0.55,
  l2Enabled: true,
  relationshipSnapshotEnabled: true,
  stableScenariosEnabled: true,
  vetoConfidenceThreshold: 0.35
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
    )
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
    "growth.forecast_l2.veto_confidence_threshold": config?.vetoConfidenceThreshold
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
