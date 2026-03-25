function toText(value) {
  return String(value || "").trim();
}

function normalizeStrategyDiffTextList(items, limit = 4) {
  if (!Array.isArray(items)) {
    return [];
  }
  const deduped = [];
  const seen = new Set();
  items.forEach((item) => {
    const text = toText(item);
    if (!text || seen.has(text)) {
      return;
    }
    seen.add(text);
    deduped.push(text);
  });
  return deduped.slice(0, limit);
}

function normalizeStrategyVersionDiff(raw) {
  if (!raw || typeof raw !== "object") {
    return null;
  }
  const added = normalizeStrategyDiffTextList(raw.added);
  const removed = normalizeStrategyDiffTextList(raw.removed);
  const promoted = normalizeStrategyDiffTextList(raw.promoted);
  const downgradeReasons = normalizeStrategyDiffTextList(raw.downgrade_reasons, 3);
  const summary = toText(raw.summary);
  const currentAssetChange = toText(raw.current_asset_change).toUpperCase();
  const comparePublishID = toText(raw.compare_publish_id);
  const compareVersion = Number(raw.compare_version || 0);
  if (!summary && !comparePublishID && compareVersion <= 0 && added.length === 0 && removed.length === 0 && promoted.length === 0 && downgradeReasons.length === 0 && !currentAssetChange) {
    return null;
  }
  return {
    added,
    removed,
    promoted,
    downgradeReasons,
    summary,
    currentAssetChange,
    comparePublishID,
    compareVersion
  };
}

function truncateStrategyText(value, maxLength = 12) {
  const text = toText(value);
  if (!text || text.length <= maxLength) {
    return text;
  }
  return `${text.slice(0, maxLength)}...`;
}

function compactId(value, maxLength = 18) {
  const text = toText(value);
  if (!text || text.length <= maxLength) {
    return text;
  }
  return `${text.slice(0, maxLength)}...`;
}

function formatStrategyMarketRegime(value) {
  return (
    {
      UPTREND: "上升趋势",
      ROTATION: "轮动切换",
      EVENT_DRIVEN: "事件驱动",
      DEFENSIVE: "防御修复",
      RISK_OFF: "风险回避",
      BASE: "基准情景",
      TREND_CONTINUE: "趋势延续",
      POLICY_POSITIVE: "政策利多",
      POLICY_NEGATIVE: "政策扰动",
      SUPPLY_SHOCK: "供给冲击",
      LIQUIDITY_SHOCK: "流动性冲击"
    }[toText(value).toUpperCase()] || toText(value)
  );
}

function formatStrategyPortfolioRole(value) {
  return (
    {
      CORE: "核心位",
      SATELLITE: "卫星位",
      WATCHLIST: "观察名单"
    }[toText(value).toUpperCase()] || toText(value)
  );
}

function firstEvidenceNote(explanation) {
  if (!Array.isArray(explanation?.evidence_cards)) {
    return "";
  }
  return firstMeaningfulStrategyText(
    explanation.evidence_cards.flatMap((item) => [item?.note, item?.value]).filter(Boolean)
  );
}

function firstEventEvidenceNote(explanation) {
  if (!Array.isArray(explanation?.event_evidence_cards)) {
    return "";
  }
  return firstMeaningfulStrategyText(
    explanation.event_evidence_cards.flatMap((item) => [item?.note, item?.value, item?.title]).filter(Boolean)
  );
}

function firstSupplyChainNote(explanation) {
  if (!Array.isArray(explanation?.supply_chain_notes)) {
    return "";
  }
  return firstMeaningfulStrategyText(explanation.supply_chain_notes);
}

export function buildStrategyRelatedEntities(explanation, options = {}) {
  const limit = Number.isFinite(Number(options.limit)) ? Number(options.limit) : 4;
  if (!Array.isArray(explanation?.related_entities)) {
    return [];
  }
  const result = [];
  const seen = new Set();
  explanation.related_entities.forEach((item) => {
    const label = toText(item?.label || item?.entity_key);
    if (!label || seen.has(label)) {
      return;
    }
    seen.add(label);
    result.push(label);
  });
  return result.slice(0, limit);
}

export function buildStrategyMemoryFeedbackText(explanation, fallback = "") {
  return firstMeaningfulStrategyText([
    explanation?.memory_feedback?.summary,
    firstMeaningfulStrategyText(explanation?.memory_feedback?.suggestions),
    fallback
  ]);
}

export function buildStrategyProofSourceText(explanation, fallback = "") {
  const relatedEntities = buildStrategyRelatedEntities(explanation, { limit: 3 });
  const relatedText = relatedEntities.length ? `关联 ${relatedEntities.join(" / ")}` : "";
  const relatedEventCount = Array.isArray(explanation?.related_events) ? explanation.related_events.length : 0;
  const eventText = relatedEventCount > 0 ? `审核事件 ${relatedEventCount} 条` : "";
  return firstMeaningfulStrategyText([
    explanation?.structure_factor_summary,
    explanation?.inventory_factor_summary,
    firstSupplyChainNote(explanation),
    firstEvidenceNote(explanation),
    firstEventEvidenceNote(explanation),
    eventText,
    relatedText,
    explanation?.consensus_summary,
    explanation?.graph_summary,
    buildStrategyMemoryFeedbackText(explanation),
    fallback
  ]);
}

export function buildStrategyRiskBoundaryText(explanation, fallback = "") {
  return firstMeaningfulStrategyText([
    explanation?.risk_boundary,
    firstMeaningfulStrategyText(explanation?.risk_flags),
    firstMeaningfulStrategyText(explanation?.invalidations),
    fallback
  ]);
}

export function buildStrategyRiskHighlights(explanation, options = {}) {
  const limit = Number.isFinite(Number(options.limit)) ? Number(options.limit) : 3;
  const fallback = toText(options.fallback);
  const items = [
    explanation?.risk_boundary,
    ...(Array.isArray(explanation?.risk_flags) ? explanation.risk_flags : []),
    ...(Array.isArray(explanation?.invalidations) ? explanation.invalidations : []),
    fallback
  ];
  const deduped = [];
  const seen = new Set();
  items.forEach((item) => {
    const text = toText(item);
    if (!text || seen.has(text)) {
      return;
    }
    seen.add(text);
    deduped.push(text);
  });
  return deduped.slice(0, limit);
}

export function buildStrategyEvaluationText(explanation) {
  const evaluation = explanation?.evaluation_meta || {};
  const status = toText(evaluation.status);
  if (!status) {
    return "";
  }
  if (status === "PENDING") {
    return "评估生成中";
  }
  const horizon5 = evaluation["5"];
  if (horizon5 && typeof horizon5 === "object") {
    const returnPct = Number(horizon5.return_pct);
    if (Number.isFinite(returnPct)) {
      return `5日评估 ${(returnPct * 100).toFixed(2)}%`;
    }
  }
  return status;
}

export function buildStrategyInsightSections(explanation, fallback = "") {
  return {
    whyNow: buildStrategySummaryText(explanation, fallback),
    proofSource: buildStrategyProofSourceText(explanation, fallback),
    riskBoundary: buildStrategyRiskBoundaryText(explanation, "当前未补更多风险边界。"),
    evaluation: buildStrategyEvaluationText(explanation)
  };
}

export function firstMeaningfulStrategyText(items) {
  if (!Array.isArray(items)) {
    return "";
  }
  return items
    .map((item) => toText(item))
    .find(Boolean) || "";
}

export function buildStrategyBatchText(source, options = {}) {
  if (!source) {
    return "";
  }
  const includeJob = options.includeJob !== false;
  const parts = [];
  const publishVersion = Number(source.publish_version);
  if (Number.isFinite(publishVersion) && publishVersion > 0) {
    parts.push(`批次 V${publishVersion}`);
  }
  if (toText(source.trade_date)) {
    parts.push(`交易日 ${toText(source.trade_date)}`);
  }
  if (includeJob && toText(source.job_id)) {
    parts.push(`任务 ${toText(source.job_id)}`);
  }
  return parts.join(" · ");
}

export function buildStrategySummaryText(explanation, fallback = "") {
  return firstMeaningfulStrategyText([
    explanation?.confidence_reason,
    firstEvidenceNote(explanation),
    buildStrategyMemoryFeedbackText(explanation),
    explanation?.consensus_summary,
    explanation?.graph_summary,
    fallback
  ]);
}

export function buildStrategyConsensusText(explanation, fallback = "") {
  return firstMeaningfulStrategyText([explanation?.consensus_summary, explanation?.graph_summary, fallback]);
}

export function buildStrategyMetaText(explanation, formatDateTime, options = {}) {
  if (!explanation) {
    return "";
  }
  const parts = [];
  if (typeof formatDateTime === "function" && explanation.generated_at) {
    parts.push(`生成 ${formatDateTime(explanation.generated_at)}`);
  }
  if (toText(explanation.market_regime)) {
    parts.push(formatStrategyMarketRegime(explanation.market_regime));
  }
  if (toText(explanation.portfolio_role)) {
    parts.push(formatStrategyPortfolioRole(explanation.portfolio_role));
  }
  if (toText(explanation.strategy_version)) {
    parts.push(`版本 ${toText(explanation.strategy_version)}`);
  }
  const evaluationText = buildStrategyEvaluationText(explanation);
  if (evaluationText) {
    parts.push(evaluationText);
  }
  if (options.includeBatch !== false) {
    const batchText = buildStrategyBatchText(explanation, { includeJob: options.includeJob });
    if (batchText) {
      parts.push(batchText);
    }
  }
  return parts.join(" · ");
}

export function buildStrategyProofTags(explanation, options = {}) {
  if (!explanation) {
    return [];
  }
  const limit = Number.isFinite(Number(options.limit)) ? Number(options.limit) : 4;
  const includeSeedCount = options.includeSeedCount === true;
  const includeVersion = options.includeVersion === true;
  const riskMaxLength = Number.isFinite(Number(options.riskMaxLength)) ? Number(options.riskMaxLength) : 12;
  const workload = explanation.workload_summary || {};
  const tags = [];
  const seedCount = Number(workload.seed_count);
  const agentCount = Number(workload.agent_count || explanation.agent_opinions?.length);
  const scenarioCount = Number(workload.scenario_count || explanation.simulations?.[0]?.scenarios?.length || 0);
  const riskText =
    buildStrategyRiskBoundaryText(explanation);
  const marketRegimeText = formatStrategyMarketRegime(explanation.market_regime);
  const portfolioRoleText = formatStrategyPortfolioRole(explanation.portfolio_role);
  const themeText = firstMeaningfulStrategyText(explanation.theme_tags);
  const relatedEntityText = firstMeaningfulStrategyText(buildStrategyRelatedEntities(explanation, { limit: 1 }));
  const relatedEventCount = Array.isArray(explanation?.related_events) ? explanation.related_events.length : 0;
  const evaluationText = buildStrategyEvaluationText(explanation);

  if (includeSeedCount && Number.isFinite(seedCount) && seedCount > 0) {
    tags.push(`种子 ${seedCount} 个`);
  }
  if (Number.isFinite(agentCount) && agentCount > 0) {
    tags.push(`评审 ${agentCount} 角`);
  }
  if (Number.isFinite(scenarioCount) && scenarioCount > 0) {
    tags.push(`场景 ${scenarioCount} 个`);
  }
  if (riskText) {
    tags.push(`边界 ${truncateStrategyText(riskText, riskMaxLength)}`);
  }
  if (marketRegimeText) {
    tags.push(marketRegimeText);
  }
  if (portfolioRoleText) {
    tags.push(portfolioRoleText);
  }
  if (themeText) {
    tags.push(`题材 ${truncateStrategyText(themeText, 10)}`);
  }
  if (relatedEntityText) {
    tags.push(`关联 ${truncateStrategyText(relatedEntityText, 10)}`);
  }
  if (relatedEventCount > 0) {
    tags.push(`事件 ${relatedEventCount} 条`);
  }
  if (evaluationText) {
    tags.push(evaluationText);
  }
  if (includeVersion && toText(explanation.strategy_version)) {
    tags.push(toText(explanation.strategy_version));
  }

  return tags.slice(0, limit);
}

export function buildStrategyEventEvidenceCards(explanation, options = {}) {
  if (!explanation) {
    return [];
  }
  const limit = Number.isFinite(Number(options.limit)) ? Number(options.limit) : 3;
  const cards = Array.isArray(explanation.event_evidence_cards) ? explanation.event_evidence_cards : [];
  return cards
    .map((item, index) => ({
      key: `${toText(item?.title)}-${toText(item?.value)}-${index}`,
      title: toText(item?.title) || "事件证据",
      value: toText(item?.value) || "系统已记录事件证据",
      note: toText(item?.note) || "已进入 reviewed event 审核证据层"
    }))
    .filter((item) => item.value)
    .slice(0, limit);
}

export function buildStrategySnapshotCard(recommendation, explanation, formatDateTime) {
  const workload = explanation?.workload_summary || {};
  return {
    name: `${recommendation?.symbol || recommendation?.contract || "-"} ${recommendation?.name || ""}`.trim(),
    summary: buildStrategySummaryText(
      explanation,
      recommendation?.reason_summary || "最近一次推荐 explanation 暂未返回更详细摘要。"
    ),
    consensus: buildStrategyConsensusText(
      explanation,
      "最近一次推荐会先做多视角验证，再收敛出最终结论。"
    ),
    meta: buildStrategyMetaText(explanation, formatDateTime, { includeBatch: false, includeJob: false }),
      workload: {
        seedCount: Number(workload.seed_count || 0),
        agentCount: Number(workload.agent_count || 0),
        scenarioCount: Number(workload.scenario_count || 0),
        candidateCount: Number(workload.candidate_count || 0)
      },
    seedHighlights: Array.isArray(explanation?.seed_highlights) ? explanation.seed_highlights.slice(0, 3) : [],
    sections: buildStrategyInsightSections(explanation, recommendation?.reason_summary || "")
  };
}

export function toStrategyTradeDate(value) {
  const text = toText(value);
  if (!text) {
    return "";
  }
  return text.slice(0, 10);
}

export function buildStrategyOriginCards(source, formatDateTime) {
  if (!source) {
    return [];
  }
  const cards = [];
  const publishVersion = Number(source.publish_version);
  const publishID = toText(source.publish_id);
  const jobID = toText(source.job_id);
  const tradeDate = toText(source.trade_date);
  if (Number.isFinite(publishVersion) && publishVersion > 0) {
    cards.push({
      label: "发布批次",
      value: `V${publishVersion}`,
      note: publishID ? `发布记录 ${compactId(publishID, 16)}` : "对应一次后端发布归档"
    });
  }
  if (tradeDate) {
    cards.push({
      label: "交易日",
      value: tradeDate,
      note:
        typeof formatDateTime === "function" && source.generated_at
          ? `生成 ${formatDateTime(source.generated_at)}`
          : "对应本次策略评估窗口"
    });
  }
  if (jobID) {
    cards.push({
      label: "任务追踪",
      value: compactId(jobID, 18),
      note: publishID ? `发布 ${compactId(publishID, 16)}` : "来源于 strategy-engine 任务"
    });
  }
  return cards.slice(0, 3);
}

export function mapStrategyVersionHistory(items, formatDateTime, options = {}) {
  if (!Array.isArray(items)) {
    return [];
  }
  const limit = Number.isFinite(Number(options.limit)) ? Number(options.limit) : 5;
  return items
    .map((item, index) => {
      const batchText = buildStrategyBatchText(item, { includeJob: false });
      const riskText =
        firstMeaningfulStrategyText(item?.risk_flags) ||
        firstMeaningfulStrategyText(item?.invalidations) ||
        "当前未补更多边界说明。";
      return {
        key: `${item?.publish_id || item?.job_id || item?.created_at || index}`,
        title:
          batchText ||
          (typeof formatDateTime === "function" && item?.created_at ? `生成 ${formatDateTime(item.created_at)}` : "当前版本"),
        version: item?.strategy_version || `V${item?.publish_version || "-"}`,
        recordVersion: item?.strategy_version || `V${item?.publish_version || "-"}`,
        recordReason: item?.confidence_reason || item?.reason_summary || "",
        publishVersion: Number(item?.publish_version || 0),
        tradeDate: item?.trade_date || "",
        versionDiff: normalizeStrategyVersionDiff(item?.version_diff),
        note: [
          item?.confidence_reason || item?.reason_summary || "当前版本未补更多摘要。",
          formatStrategyMarketRegime(item?.market_regime),
          item?.risk_boundary,
          riskText
        ]
          .filter(Boolean)
          .join(" · ")
      };
    })
    .slice(0, limit);
}

export function buildFallbackStrategyVersionHistory(explanation, options = {}) {
  if (!explanation) {
    return [];
  }
  return mapStrategyVersionHistory(
    [
      {
        publish_id: explanation.publish_id,
        job_id: explanation.job_id,
        trade_date: options.tradeDate || explanation.trade_date,
        publish_version: explanation.publish_version,
        created_at: explanation.generated_at,
        strategy_version: explanation.strategy_version || options.strategyVersion,
        reason_summary: options.reasonSummary || "",
        confidence_reason: explanation.confidence_reason,
        market_regime: explanation.market_regime,
        portfolio_role: explanation.portfolio_role,
        risk_boundary: explanation.risk_boundary,
        theme_tags: explanation.theme_tags,
        sector_tags: explanation.sector_tags,
        risk_flags: explanation.risk_flags,
        invalidations: explanation.invalidations,
        evaluation_meta: explanation.evaluation_meta
      }
    ],
    options.formatDateTime,
    options
  );
}

export function findMatchedStrategyHistoryItem(historyItems, options = {}) {
  if (!Array.isArray(historyItems) || historyItems.length === 0) {
    return null;
  }
  const tradeDate = toStrategyTradeDate(options.tradeDate);
  const publishVersion = Number(options.publishVersion || 0);
  if (tradeDate) {
    const matchedByDate = historyItems.find((history) => history.tradeDate === tradeDate);
    if (matchedByDate) {
      return matchedByDate;
    }
  }
  if (publishVersion > 0) {
    const matchedByVersion = historyItems.find((history) => history.publishVersion === publishVersion);
    if (matchedByVersion) {
      return matchedByVersion;
    }
  }
  return historyItems[0];
}

export function findSelectedStrategyHistoryItem(historyItems, selectedKey, fallbackItem) {
  if (!Array.isArray(historyItems) || historyItems.length === 0) {
    return fallbackItem || null;
  }
  if (toText(selectedKey)) {
    const selected = historyItems.find((history) => history.key === selectedKey);
    if (selected) {
      return selected;
    }
  }
  return fallbackItem || historyItems[0];
}

export function buildStrategyVersionDiff(config = {}) {
  const explanation = config.explanation;
  if (!explanation) {
    return null;
  }
  const structuredDiff = normalizeStrategyVersionDiff(config.versionDiff || explanation?.version_diff);
  const recordReason = toText(config.recordReason);
  const afterReason = toText(explanation.confidence_reason || recordReason);
  const fallbackRecordLabel = toText(config.fallbackRecordLabel) || "record";
  const beforeVersion = toText(config.recordVersion) || fallbackRecordLabel;
  const afterVersion = toText(explanation.strategy_version) || "strategy-engine";
  const riskNote =
    buildStrategyRiskBoundaryText(explanation) ||
    "当前未返回更多边界差异";
  const regimeText = formatStrategyMarketRegime(explanation.market_regime);

  let diffLabel = "解释补全";
  let diffNote = "当前 explanation 补充了更多过程信息。";
  if (beforeVersion !== afterVersion && recordReason !== afterReason) {
    diffLabel = "版本升级";
    diffNote = toText(config.upgradedText) || "记录版本和当前 explanation 的结论都发生了变化。";
  } else if (recordReason !== afterReason) {
    diffLabel = "结论改写";
    diffNote = toText(config.reasonChangedText) || "当前 explanation 对原始推荐理由做了新的收敛说明。";
  } else if (beforeVersion !== afterVersion) {
    diffLabel = "版本刷新";
    diffNote = toText(config.versionChangedText) || "结论主线仍在，但解释版本已经更新。";
  }

  const metaParts = [`当前 ${afterVersion}`];
  if (typeof config.formatDateTime === "function" && explanation.generated_at) {
    metaParts.push(config.formatDateTime(explanation.generated_at));
  }
  const batchText = buildStrategyBatchText(explanation, { includeJob: config.includeJobInMeta });
  if (batchText) {
    metaParts.push(batchText);
  }

  if (structuredDiff) {
    const diffParts = [];
    if (structuredDiff.summary) {
      diffParts.push(structuredDiff.summary);
    }
    if (structuredDiff.added.length) {
      diffParts.push(`新增 ${structuredDiff.added.join("、")}`);
    }
    if (structuredDiff.removed.length) {
      diffParts.push(`移除 ${structuredDiff.removed.join("、")}`);
    }
    if (structuredDiff.promoted.length) {
      diffParts.push(`上调优先级 ${structuredDiff.promoted.join("、")}`);
    }
    if (structuredDiff.downgradeReasons.length) {
      diffParts.push(`下降原因 ${structuredDiff.downgradeReasons.join("；")}`);
    }
    let diffLabel = "版本差异";
    if (structuredDiff.currentAssetChange === "ADDED") {
      diffLabel = "新增";
    } else if (structuredDiff.currentAssetChange === "PROMOTED") {
      diffLabel = "上调优先级";
    } else if (structuredDiff.currentAssetChange === "WEAKENED") {
      diffLabel = "下降原因";
    }
    return {
      meta: metaParts.join(" · "),
      beforeLabel: beforeVersion,
      beforeNote: recordReason || "记录版本未补更多摘要。",
      afterLabel: afterVersion,
      afterNote: afterReason || "当前 explanation 未补更多摘要。",
      diffLabel,
      diffNote: diffParts.filter(Boolean).join(" · ") || [regimeText, riskNote].filter(Boolean).join(" · ")
    };
  }

  return {
    meta: metaParts.join(" · "),
    beforeLabel: beforeVersion,
    beforeNote: recordReason || "记录版本未补更多摘要。",
    afterLabel: afterVersion,
    afterNote: afterReason || "当前 explanation 未补更多摘要。",
    diffLabel,
    diffNote: [diffNote, regimeText, riskNote].filter(Boolean).join(" · ")
  };
}

export function buildStrategyHistoryCompareState(config = {}) {
  const historyItems = Array.isArray(config.historyItems) ? config.historyItems : [];
  const fallbackItem = config.fallbackItem || historyItems[0] || null;
  const selectedItem = findSelectedStrategyHistoryItem(historyItems, config.selectedKey, fallbackItem);
  const explanation = config.explanation;
  const explanationPublishVersion = Number(explanation?.publish_version || 0);
  const selectedVersionDiff =
    selectedItem && explanationPublishVersion > 0 && selectedItem.publishVersion === explanationPublishVersion - 1
      ? explanation?.version_diff || selectedItem.versionDiff
      : selectedItem?.versionDiff;
  return {
    selectedItem,
    selectedKey: selectedItem?.key || "",
    selectedTitle: selectedItem?.title ? `${toText(config.selectedTitlePrefix)}${selectedItem.title}` : "",
    selectedNote: selectedItem?.note || "",
    isCustomSelected: Boolean(toText(config.selectedKey)) && toText(config.selectedKey) === selectedItem?.key,
    diff:
      selectedItem && explanation
        ? buildStrategyVersionDiff({
            recordVersion: selectedItem.recordVersion || config.recordVersionFallback,
            recordReason: selectedItem.recordReason || config.recordReasonFallback,
            explanation,
            formatDateTime: config.formatDateTime,
            fallbackRecordLabel: config.fallbackRecordLabel,
            upgradedText: config.upgradedText,
            reasonChangedText: config.reasonChangedText,
            versionChangedText: config.versionChangedText,
            includeJobInMeta: config.includeJobInMeta,
            versionDiff: selectedVersionDiff
          })
        : null
  };
}
