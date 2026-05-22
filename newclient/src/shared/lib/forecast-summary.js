function toText(value) {
  return String(value || "").trim();
}

function formatStatusLabel(status) {
  return (
    {
      QUEUED: "排队中",
      RUNNING: "推演中",
      SUCCEEDED: "已完成",
      FAILED: "已失败",
      CANCELLED: "已取消"
    }[toText(status).toUpperCase()] || "L3 状态"
  );
}

function formatStatusTone(status) {
  return (
    {
      QUEUED: "queued",
      RUNNING: "running",
      SUCCEEDED: "success",
      FAILED: "failed",
      CANCELLED: "muted"
    }[toText(status).toUpperCase()] || "muted"
  );
}

export function buildDeepForecastSummary(source) {
  if (!source || typeof source !== "object") {
    return null;
  }

  const summary = source.deep_forecast_summary || source.summary;
  const reportRef = source.deep_forecast_report_ref || source.report_ref;
  if ((!summary || typeof summary !== "object") && (!reportRef || typeof reportRef !== "object")) {
    return null;
  }

  const runId = toText(summary?.run_id || summary?.runID || reportRef?.run_id || reportRef?.runID);
  const reportId = toText(reportRef?.report_id || reportRef?.reportID);
  const status = toText(summary?.status || reportRef?.status).toUpperCase();
  const scenario = toText(summary?.primary_scenario || summary?.primaryScenario);
  const actionGuidance = toText(summary?.action_guidance || summary?.actionGuidance);
  const executiveSummary = toText(summary?.executive_summary || summary?.executiveSummary);
  const generatedAt = toText(summary?.generated_at || summary?.generatedAt || reportRef?.generated_at || reportRef?.generatedAt);
  const reportAvailable = Boolean(summary?.report_available ?? summary?.reportAvailable ?? reportId);
  const requiresVip = reportRef?.requires_vip === true || reportRef?.requiresVIP === true;
  const fullReadable = reportRef?.full_readable === true || reportRef?.fullReadable === true;
  const statusLabel = formatStatusLabel(status);

  if (!runId && !reportId && !executiveSummary && !scenario && !actionGuidance) {
    return null;
  }

  return {
    runId,
    reportId,
    status,
    statusLabel,
    tone: formatStatusTone(status),
    summary:
      executiveSummary ||
      [scenario ? `主情景 ${scenario}` : "", actionGuidance].filter(Boolean).join(" · ") ||
      "当前未补更多深推演摘要。",
    scenario,
    actionGuidance,
    generatedAt,
    reportAvailable,
    requiresVip,
    fullReadable,
    note: [scenario ? `主情景 ${scenario}` : "", actionGuidance].filter(Boolean).join(" · ")
  };
}
