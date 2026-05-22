import { computed } from "vue";
import {
  buildStrategyDeepForecastSummary,
  buildFallbackStrategyVersionHistory,
  buildStrategyAgentOpinionRows,
  buildStrategyConfidenceCalibrationSummary,
  buildStrategyHistoryCompareState,
  buildStrategyInsightSections,
  buildStrategyOriginCards,
  buildStrategyProofTags,
  buildStrategyRelationshipSnapshotSummary,
  buildStrategyResearchOutlineRows,
  buildStrategyRiskBoundaryText,
  buildStrategyScenarioMetaSummary,
  buildStrategyScenarioSnapshotRows,
  buildStrategyThesisCardRows,
  buildStrategyVersionDiff,
  buildStrategyWatchSignalRows,
  firstMeaningfulStrategyText,
  mapStrategyVersionHistory
} from "../lib/strategy-version";
import {
  formatDateTime,
  formatScore,
  mapRisk
} from "../utils/finance";

/**
 * Composable for strategy insight and version history logic
 */
export function useStrategyInsight(activeView, activeBase, explanationMap, versionHistoryMap, selectedVersionKey) {
  const explanation = computed(() => {
    const id = activeView.value?.id;
    return id ? explanationMap.value[id] || null : null;
  });

  const insightSections = computed(() =>
    buildStrategyInsightSections(explanation.value, activeView.value?.reason || "")
  );

  const proofTags = computed(() => buildStrategyProofTags(explanation.value, { limit: 4 }));
  
  const seedHighlights = computed(() => explanation.value?.seed_highlights || []);
  
  const scenarioCards = computed(() => buildStrategyScenarioSnapshotRows(explanation.value));
  
  const scenarioMeta = computed(() => buildStrategyScenarioMetaSummary(explanation.value));
  
  const relationshipSummary = computed(() => buildStrategyRelationshipSnapshotSummary(explanation.value));
  
  const agentOpinions = computed(() => buildStrategyAgentOpinionRows(explanation.value));
  
  const riskCards = computed(() => buildStrategyRiskCardsLocal(explanation.value));

  const explanationCards = computed(() => {
    const exp = explanation.value || {};
    const insightSection = insightSections.value;
    const evidenceCount = Array.isArray(exp.evidence_cards) ? exp.evidence_cards.length : 0;
    const agentCount = Number(exp.workload_summary?.agent_count ?? 0);
    const scenarioCount = Number(exp.workload_summary?.scenario_count ?? 0);
    
    const researchOutline = buildStrategyResearchOutlineRows(exp, { limit: 1 })[0];
    const activeThesis = buildStrategyThesisCardRows(exp, "active", { limit: 1 })[0];
    const historicalThesis = buildStrategyThesisCardRows(exp, "historical", { limit: 1 })[0];
    const watchSignal = buildStrategyWatchSignalRows(exp, { limit: 1 })[0];
    const calibration = buildStrategyConfidenceCalibrationSummary(exp);

    return [
      {
        label: "种子输入",
        value: Array.isArray(exp.seed_highlights) ? exp.seed_highlights.length : 0,
        note: exp.seed_summary || insightSection.whyNow || "系统会先处理种子输入，再逐步筛选。"
      },
      {
        label: "证据卡片",
        value: evidenceCount,
        note: insightSection.proofSource || "系统会把证据和多角色结论汇总后展示。"
      },
      {
        label: "评审覆盖",
        value: `${agentCount} 角 / ${scenarioCount} 景`,
        note: buildStrategyRiskBoundaryText(exp, "当前未补更多风险边界。")
      },
      researchOutline && {
        label: "研究拆解",
        value: researchOutline.title,
        note: researchOutline.summary
      },
      activeThesis && {
        label: "当前有效理由",
        value: activeThesis.title,
        note: activeThesis.summary
      },
      historicalThesis && {
        label: "历史弱化理由",
        value: historicalThesis.title,
        note: historicalThesis.summary
      },
      (watchSignal || calibration) && {
        label: "观察与校准",
        value: watchSignal?.title || calibration?.summary || "继续观察",
        note: firstMeaningfulStrategyText([watchSignal?.trigger, calibration?.deltaLabel, calibration?.note])
      }
    ].filter(Boolean);
  });

  const originCards = computed(() => buildStrategyOriginCards(explanation.value, formatDateTime));

  const versionHistoryItems = computed(() => {
    const id = activeView.value?.id;
    const items = id ? versionHistoryMap.value[id] : [];
    if (Array.isArray(items) && items.length > 0) {
      return mapStrategyVersionHistory(items, formatDateTime);
    }
    return buildFallbackStrategyVersionHistory(explanation.value, {
      reasonSummary: activeBase.value?.reason_summary || "",
      strategyVersion: activeBase.value?.strategy_version || "record",
      formatDateTime
    });
  });

  const historyCompare = computed(() => {
    const items = versionHistoryItems.value;
    return buildStrategyHistoryCompareState({
      historyItems: items,
      selectedKey: selectedVersionKey.value,
      fallbackItem: items[1] || items[0] || null,
      explanation: explanation.value,
      selectedTitlePrefix: "当前对比 ",
      formatDateTime,
      fallbackRecordLabel: "record",
      upgradedText: "所选历史版本和当前 explanation 的结论都发生了变化。",
      reasonChangedText: "当前 explanation 对所选历史版本做了新的收敛。",
      versionChangedText: "核心结论延续，但版本已发生刷新。"
    });
  });

  const versionDiff = computed(() => {
    const base = activeBase.value;
    const exp = explanation.value;
    if (!base || !exp) return null;
    return buildStrategyVersionDiff({
      recordVersion: base.strategy_version || "record",
      recordReason: base.reason_summary || "",
      explanation: exp,
      formatDateTime,
      fallbackRecordLabel: "record",
      upgradedText: "记录版本和当前 explanation 的结论都发生了变化。",
      reasonChangedText: "当前 explanation 对原始推荐理由做了新的收敛。",
      versionChangedText: "核心结论仍在，但解释版本已经更新。"
    });
  });

  const deepForecast = computed(() => buildStrategyDeepForecastSummary(explanation.value));

  return {
    explanation,
    insightSections,
    proofTags,
    seedHighlights,
    scenarioCards,
    scenarioMeta,
    relationshipSummary,
    agentOpinions,
    riskCards,
    explanationCards,
    originCards,
    versionHistoryItems,
    historyCompare,
    versionDiff,
    deepForecast
  };
}

// Local helper duplicated from PcStrategyView for extraction
function buildStrategyRiskCardsLocal(explanation) {
  if (!explanation) return [];
  const riskBoundary = explanation.risk_boundary || {};
  return [
    { label: "止损触发", value: riskBoundary.stop_loss_trigger || "未设定" },
    { label: "失效条件", value: riskBoundary.invalid_condition || "未设定" },
    { label: "仓位控制", value: riskBoundary.position_guidance || "正常" }
  ].filter(Boolean);
}
