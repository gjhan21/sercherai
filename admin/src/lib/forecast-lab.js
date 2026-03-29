const SUPPORTED_TARGET_TYPES = new Set(["STOCK", "FUTURES"]);
const SUPPORTED_STATUSES = new Set(["QUEUED", "RUNNING", "SUCCEEDED", "FAILED", "CANCELLED"]);
const SUPPORTED_TRIGGER_TYPES = new Set(["ADMIN_MANUAL", "AUTO_PRIORITY", "USER_REQUEST"]);

function normalizeText(raw) {
  return String(raw ?? "").trim();
}

function normalizeEnum(raw, supportedValues) {
  const value = normalizeText(raw).toUpperCase();
  return supportedValues.has(value) ? value : "";
}

export function normalizeForecastLabRouteState(query = {}) {
  return {
    runID: normalizeText(query.run_id),
    targetType: normalizeEnum(query.target_type, SUPPORTED_TARGET_TYPES),
    status: normalizeEnum(query.status, SUPPORTED_STATUSES),
    triggerType: normalizeEnum(query.trigger_type, SUPPORTED_TRIGGER_TYPES),
    userID: normalizeText(query.user_id)
  };
}

export function buildForecastLabRouteQuery(state = {}) {
  const query = {};
  const runID = normalizeText(state.runID);
  const targetType = normalizeEnum(state.targetType, SUPPORTED_TARGET_TYPES);
  const status = normalizeEnum(state.status, SUPPORTED_STATUSES);
  const triggerType = normalizeEnum(state.triggerType, SUPPORTED_TRIGGER_TYPES);
  const userID = normalizeText(state.userID);

  if (runID) {
    query.run_id = runID;
  }
  if (targetType) {
    query.target_type = targetType;
  }
  if (status) {
    query.status = status;
  }
  if (triggerType) {
    query.trigger_type = triggerType;
  }
  if (userID) {
    query.user_id = userID;
  }
  return query;
}
