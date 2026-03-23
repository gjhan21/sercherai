export function resolveReviewDialogMeta(mode) {
  const normalized = String(mode || "approve").trim().toLowerCase();
  if (normalized === "force") {
    return {
      title: "强制发布确认",
      primaryText: "确认强制发布",
      primaryType: "warning",
      summaryTone: "warning"
    };
  }
  if (normalized === "reject") {
    return {
      title: "驳回审核",
      primaryText: "确认驳回",
      primaryType: "danger",
      summaryTone: "danger"
    };
  }
  if (normalized === "blocked") {
    return {
      title: "默认发布已被拦截",
      primaryText: "改为强制发布",
      primaryType: "primary",
      summaryTone: "warning"
    };
  }
  return {
    title: "审核通过并发布",
    primaryText: "确认发布",
    primaryType: "primary",
    summaryTone: "primary"
  };
}

export function extractReviewConflictReason(error) {
  const raw = String(error?.message || "").trim();
  const payloadDetail = String(error?.payload?.detail || "").trim();
  const conflictType = String(error?.payload?.conflict_type || "").trim();
  if (Number(error?.code) === 40901 || conflictType === "PUBLISH_POLICY_BLOCKED") {
    return payloadDetail || raw;
  }
  if (!raw) {
    return "";
  }
  const lower = raw.toLowerCase();
  if (!lower.includes("returned 409 when publishing job") && !raw.includes("发布策略拦截")) {
    return "";
  }
  const jsonStart = raw.indexOf("{");
  if (jsonStart >= 0) {
    try {
      const payload = JSON.parse(raw.slice(jsonStart));
      if (typeof payload?.detail === "string" && payload.detail.trim()) {
        return payload.detail.trim();
      }
    } catch {
      // Ignore JSON parse failures and fall back to string slicing.
    }
  }
  const marker = raw.lastIndexOf(":");
  if (marker >= 0 && marker + 1 < raw.length) {
    return raw.slice(marker + 1).trim();
  }
  return raw;
}
