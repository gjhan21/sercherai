function normalizeRunStatus(value) {
  return String(value || "").trim().toUpperCase();
}

function normalizeDateValue(run) {
  return String(run?.finished_at || run?.finishedAt || run?.created_at || run?.createdAt || "").trim();
}

function normalizeDimensionKey(item) {
  return String(item?.dimension || item?.key || "").trim().toUpperCase();
}

function extractReviewScore(run) {
  const reviewScore = Number(run?.review?.review_score ?? run?.report?.review_score ?? run?.summary?.review_score);
  return Number.isFinite(reviewScore) ? reviewScore : null;
}

function extractHeadline(run) {
  return String(
    run?.headline_verdict ||
    run?.report?.headline_verdict ||
    run?.summary?.executive_summary ||
    run?.executive_summary ||
    ""
  ).trim();
}

function extractScenario(run) {
  return String(
    run?.primary_scenario ||
    run?.report?.primary_scenario ||
    run?.summary?.scenario ||
    run?.summary?.primary_scenario ||
    ""
  ).trim();
}

function extractEvidenceMap(run) {
  const raw = Array.isArray(run?.dimension_evidence)
    ? run.dimension_evidence
    : Array.isArray(run?.report?.dimension_evidence)
      ? run.report.dimension_evidence
      : [];
  const map = new Map();
  raw.forEach((item) => {
    const key = normalizeDimensionKey(item);
    if (!key) return;
    const summary = String(item?.summary || "").trim();
    if (!summary) return;
    map.set(key, summary);
  });
  return map;
}

function buildConclusionChange(latest, previous) {
  const latestHeadline = extractHeadline(latest);
  const previousHeadline = extractHeadline(previous);
  const latestScenario = extractScenario(latest);
  const previousScenario = extractScenario(previous);
  const changed = latestHeadline !== previousHeadline || latestScenario !== previousScenario;
  const summary = changed
    ? `核心判断由“${previousHeadline || "未给出"}”变为“${latestHeadline || "未给出"}”，主情景从“${previousScenario || "未给出"}”切到“${latestScenario || "未给出"}”。`
    : `两次深推演的核心判断整体保持一致，当前仍以“${latestHeadline || "未给出"}”为主。`;
  return { changed, summary };
}

function buildEvidenceChange(latest, previous) {
  const latestMap = extractEvidenceMap(latest);
  const previousMap = extractEvidenceMap(previous);
  const dimensions = new Set([...latestMap.keys(), ...previousMap.keys()]);
  const changedDimensions = [];
  dimensions.forEach((key) => {
    if ((latestMap.get(key) || "") !== (previousMap.get(key) || "")) changedDimensions.push(key);
  });
  const changed = changedDimensions.length > 0;
  const summary = changed
    ? `证据变化主要出现在 ${changedDimensions.join("、")} 维度。`
    : "证据维度整体没有出现明显变化。";
  return { changed, summary, changedDimensions };
}

function sortSuccessfulRuns(runs) {
  return (Array.isArray(runs) ? runs : [])
    .filter((run) => {
      const status = normalizeRunStatus(run?.status);
      if (!status) {
        return Boolean(run?.run_id || run?.id);
      }
      return status === "SUCCESS" || status === "SUCCEEDED";
    })
    .slice()
    .sort((left, right) => {
      const leftDate = normalizeDateValue(left);
      const rightDate = normalizeDateValue(right);
      if (leftDate === rightDate) {
        return String(right?.id || "").localeCompare(String(left?.id || ""));
      }
      return rightDate.localeCompare(leftDate);
    });
}

function mapTimelineItem(run) {
  return {
    id: run?.id || run?.run_id || "",
    title: run?.target_label || run?.targetLabel || run?.target_key || run?.targetKey || "未命名标的",
    verdict: extractHeadline(run),
    scenario: extractScenario(run),
    reviewScore: extractReviewScore(run),
    finishedAt: normalizeDateValue(run)
  };
}

export function getLatestSuccessfulForecastRun(runs = []) {
  return sortSuccessfulRuns(runs)[0] || null;
}

function buildReviewScores(runs) {
  return runs
    .slice()
    .reverse()
    .map((run) => ({
      id: run?.id || run?.run_id || "",
      score: extractReviewScore(run),
      label: extractHeadline(run) || extractScenario(run) || "深推演复盘"
    }))
    .filter((item) => item.score !== null);
}

export function buildForecastHistoryViewModel({ targetKey = "", targetLabel = "", runs = [] } = {}) {
  const successfulRuns = sortSuccessfulRuns(runs);
  const latest = successfulRuns[0] || null;
  const previous = successfulRuns[1] || null;
  const hasHistory = successfulRuns.length > 1;
  const compareSummary = hasHistory
    ? {
        title: "最新一次 vs 上一次",
        conclusionChange: buildConclusionChange(latest, previous),
        evidenceChange: buildEvidenceChange(latest, previous)
      }
    : {
        title: "最新一次 vs 上一次",
        emptyMessage: "当前仅有 1 次成功深推演，暂不能形成历史对比"
      };

  return {
    targetKey: String(targetKey || latest?.target_key || latest?.targetKey || "").trim(),
    targetLabel: String(targetLabel || latest?.target_label || latest?.targetLabel || "").trim(),
    hasHistory,
    successRunCount: successfulRuns.length,
    comparePair: {
      latest,
      previous
    },
    compareSummary,
    reviewScores: buildReviewScores(successfulRuns),
    timeline: successfulRuns.map(mapTimelineItem)
  };
}
