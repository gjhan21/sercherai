export const DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS = 24;

export const MARKET_QUALITY_LOOKBACK_OPTIONS = [
  { label: "近 24 小时", value: 24 },
  { label: "近 3 天", value: 72 },
  { label: "近 7 天", value: 168 }
];

export function normalizeMarketQualityLookbackHours(value, fallback = DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS) {
  const hours = Number(value);
  if (Number.isFinite(hours) && hours > 0) {
    return Math.trunc(hours);
  }
  return fallback;
}

export function formatMarketQualityLookbackLabel(value) {
  const hours = normalizeMarketQualityLookbackHours(value);
  return (
    MARKET_QUALITY_LOOKBACK_OPTIONS.find((item) => item.value === hours)?.label ||
    `近 ${hours} 小时`
  );
}

export function normalizeMarketCenterTab(value) {
  const normalized = String(value || "").trim().toLowerCase();
  if (normalized === "futures") return "futures";
  if (normalized === "engine-config") return "engine-config";
  return "stocks";
}

export function buildMarketCenterRouteQuery(filters = {}) {
  const query = {
    tab: normalizeMarketCenterTab(filters?.tab)
  };
  const issueCode = String(filters?.issue_code || "").trim().toUpperCase();
  const qualityHours = normalizeMarketQualityLookbackHours(
    filters?.quality_hours ?? filters?.hours,
    0
  );
  if (qualityHours > 0) {
    query.quality_hours = String(qualityHours);
  }
  if (issueCode) {
    query.issue_code = issueCode;
  }
  return query;
}

export function normalizeMarketCenterRouteState(query = {}) {
  return {
    tab: normalizeMarketCenterTab(query?.tab),
    quality_hours: normalizeMarketQualityLookbackHours(query?.quality_hours),
    issue_code: String(query?.issue_code || "").trim().toUpperCase()
  };
}

export function normalizeMarketQualityRouteContext(query = {}) {
  const returnTo = String(query?.return_to || "").trim().toLowerCase();
  if (returnTo !== "market-center") {
    return {};
  }
  return {
    return_to: "market-center",
    return_tab: normalizeMarketCenterTab(query?.return_tab)
  };
}

export function formatMarketCenterReturnLabel(tab) {
  if (normalizeMarketCenterTab(tab) === "futures") {
    return "返回期货行情中心";
  }
  return "返回股票行情中心";
}

export function buildMarketCenterQualityReturnQuery(item = {}, hours) {
  const assetClass = String(item?.asset_class || "").trim().toUpperCase();
  return buildMarketCenterRouteQuery({
    tab: assetClass === "FUTURES" ? "futures" : "stocks",
    quality_hours: hours,
    issue_code: item?.issue_code || ""
  });
}

export function buildMarketCenterQualitySummaryReturnQuery(assetClass, summary = {}, hours, mode = "latest") {
  const normalizedAssetClass = String(assetClass || "").trim().toUpperCase();
  const issueCode =
    mode === "error"
      ? String(summary?.latest_error_issue_code || "").trim()
      : String(summary?.latest_issue_code || "").trim();
  return buildMarketCenterRouteQuery({
    tab: normalizedAssetClass === "FUTURES" ? "futures" : "stocks",
    quality_hours: hours,
    issue_code: issueCode
  });
}

export function buildMarketQualityDrillQuery(filters = {}) {
  return {
    return_to: "market-center",
    return_tab: normalizeMarketCenterTab(filters?.return_tab),
    ...buildMarketQualityRouteQuery(filters)
  };
}

export function marketQualitySeverityTagType(severity) {
  const normalized = String(severity || "").trim().toUpperCase();
  if (normalized === "ERROR") return "danger";
  if (normalized === "WARN") return "warning";
  if (normalized === "INFO") return "success";
  return "info";
}

export function formatTruthRebuildSuccessMessage(assetClass, result = {}) {
  const normalized = String(assetClass || result?.asset_class || "").trim().toUpperCase();
  const truthBarCount = Number(result?.truth_bar_count) || 0;
  if (normalized === "FUTURES") {
    return `期货主力映射重建完成，处理 ${truthBarCount} 条 truth 行情`;
  }
  return `股票状态 truth 重建完成，处理 ${truthBarCount} 条 truth 行情`;
}

export function formatMarketQualityPayload(payload) {
  const raw = String(payload || "").trim();
  if (!raw) {
    return "";
  }
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}

export function collectMarketQualityIssueOptions(items, limit = 8) {
  const counts = new Map();
  (Array.isArray(items) ? items : []).forEach((item) => {
    const code = String(item?.issue_code || "").trim();
    if (!code) {
      return;
    }
    counts.set(code, (counts.get(code) || 0) + 1);
  });
  return Array.from(counts.entries())
    .sort((a, b) => {
      if (b[1] !== a[1]) {
        return b[1] - a[1];
      }
      return a[0].localeCompare(b[0]);
    })
    .slice(0, Math.max(0, Number(limit) || 0))
    .map(([value, count]) => ({
      value,
      label: `${value} (${count})`
    }));
}

export function buildMarketQualityDialogTitle(item = {}) {
  const assetClass = String(item?.asset_class || "").trim().toUpperCase();
  const issueCode = String(item?.issue_code || "").trim().toUpperCase();
  const parts = ["质量日志详情"];
  if (assetClass) {
    parts.push(assetClass);
  }
  if (issueCode) {
    parts.push(issueCode);
  }
  return parts.join(" · ");
}

export function buildMarketQualityRouteQuery(filters = {}) {
  const query = {};
  const assetClass = String(filters?.asset_class || "").trim().toUpperCase();
  const dataKind = String(filters?.data_kind || "").trim().toUpperCase();
  const severity = String(filters?.severity || "").trim().toUpperCase();
  const issueCode = String(filters?.issue_code || "").trim().toUpperCase();
  const hours = normalizeMarketQualityLookbackHours(filters?.hours, 0);

  if (assetClass) query.asset_class = assetClass;
  if (dataKind) query.data_kind = dataKind;
  if (severity) query.severity = severity;
  if (issueCode) query.issue_code = issueCode;
  if (hours > 0) {
    query.hours = String(hours);
  }
  return query;
}

export function normalizeMarketQualityRouteFilters(query = {}) {
  return {
    asset_class: String(query?.asset_class || "").trim().toUpperCase(),
    data_kind: String(query?.data_kind || "").trim().toUpperCase(),
    severity: String(query?.severity || "").trim().toUpperCase(),
    issue_code: String(query?.issue_code || "").trim().toUpperCase(),
    hours: normalizeMarketQualityLookbackHours(query?.hours)
  };
}

export function buildMarketQualityBucketSummary(items, resolver, limit = 5) {
  const counts = new Map();
  (Array.isArray(items) ? items : []).forEach((item) => {
    const rawValue = typeof resolver === "function" ? resolver(item) : "";
    const value = String(rawValue || "").trim().toUpperCase();
    if (!value) {
      return;
    }
    counts.set(value, (counts.get(value) || 0) + 1);
  });
  return Array.from(counts.entries())
    .sort((a, b) => {
      if (b[1] !== a[1]) {
        return b[1] - a[1];
      }
      return a[0].localeCompare(b[0]);
    })
    .slice(0, Math.max(0, Number(limit) || 0))
    .map(([value, count]) => ({
      value,
      count,
      label: value
    }));
}

export function areMarketQualityFiltersEqual(left = {}, right = {}) {
  const normalizedLeft = normalizeMarketQualityRouteFilters(left);
  const normalizedRight = normalizeMarketQualityRouteFilters(right);
  return (
    normalizedLeft.asset_class === normalizedRight.asset_class &&
    normalizedLeft.data_kind === normalizedRight.data_kind &&
    normalizedLeft.severity === normalizedRight.severity &&
    normalizedLeft.issue_code === normalizedRight.issue_code &&
    normalizedLeft.hours === normalizedRight.hours
  );
}
