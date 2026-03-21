const WATCHLIST_STORAGE_KEY = "sercherai:watchlist:stocks:v1";
export const WATCHLIST_EVENT = "sercherai-watchlist-updated";
const WATCHLIST_HISTORY_LIMIT = 12;

function canUseStorage() {
  return typeof window !== "undefined" && typeof window.localStorage !== "undefined";
}

function emitWatchlistChange() {
  if (typeof window !== "undefined") {
    window.dispatchEvent(new Event(WATCHLIST_EVENT));
  }
}

function normalizeAgentOpinions(agentOpinions) {
  if (!Array.isArray(agentOpinions)) {
    return [];
  }
  return agentOpinions
    .map((entry) => ({
      agent: entry?.agent || "",
      stance: entry?.stance || ""
    }))
    .filter((entry) => entry.agent || entry.stance);
}

function buildSnapshot(payload = {}, recordedAt = new Date().toISOString()) {
  return {
    reason_summary: payload.reason_summary || "",
    risk_level: payload.risk_level || "",
    status: payload.status || "",
    take_profit: payload.take_profit || "",
    stop_loss: payload.stop_loss || "",
    strategy_version: payload.strategy_version || "",
    publish_id: payload.publish_id || "",
    job_id: payload.job_id || "",
    trade_date: payload.trade_date || "",
    publish_version: Number(payload.publish_version) > 0 ? Number(payload.publish_version) : 0,
    agent_opinions: normalizeAgentOpinions(payload.agent_opinions),
    recorded_at: payload.recorded_at || recordedAt
  };
}

function isSameSnapshot(left, right) {
  if (!left || !right) {
    return false;
  }
  return (
    (left.reason_summary || "") === (right.reason_summary || "") &&
    (left.risk_level || "") === (right.risk_level || "") &&
    (left.status || "") === (right.status || "") &&
    (left.take_profit || "") === (right.take_profit || "") &&
    (left.stop_loss || "") === (right.stop_loss || "") &&
    (left.strategy_version || "") === (right.strategy_version || "") &&
    (left.publish_id || "") === (right.publish_id || "") &&
    (left.job_id || "") === (right.job_id || "") &&
    (left.trade_date || "") === (right.trade_date || "") &&
    Number(left.publish_version || 0) === Number(right.publish_version || 0) &&
    JSON.stringify(normalizeAgentOpinions(left.agent_opinions)) ===
      JSON.stringify(normalizeAgentOpinions(right.agent_opinions))
  );
}

export function listWatchedStocks() {
  if (!canUseStorage()) {
    return [];
  }
  try {
    const raw = window.localStorage.getItem(WATCHLIST_STORAGE_KEY);
    const parsed = JSON.parse(raw || "[]");
    if (!Array.isArray(parsed)) {
      return [];
    }
    return parsed.filter((item) => item && item.id);
  } catch {
    return [];
  }
}

export function isWatchedStock(id) {
  return listWatchedStocks().some((item) => item.id === id);
}

export function saveWatchedStock(item) {
  if (!item?.id || !canUseStorage()) {
    return false;
  }
  const items = listWatchedStocks();
  const addedAt = new Date().toISOString();
  const baselineSnapshot = buildSnapshot(
    {
      reason_summary: item.baseline_reason_summary || item.reason_summary || "",
      risk_level: item.baseline_risk_level || item.risk_level || "",
      status: item.baseline_status || item.status || "",
      take_profit: item.baseline_take_profit || item.take_profit || "",
      stop_loss: item.baseline_stop_loss || item.stop_loss || "",
      strategy_version: item.baseline_strategy_version || item.strategy_version || "",
      publish_id: item.baseline_publish_id || "",
      job_id: item.baseline_job_id || "",
      trade_date: item.baseline_trade_date || "",
      publish_version: item.baseline_publish_version || 0,
      agent_opinions: item.baseline_agent_opinions
    },
    addedAt
  );
  const next = [
    {
      id: item.id,
      symbol: item.symbol || "",
      name: item.name || "",
      risk_level: item.risk_level || "",
      status: item.status || "",
      reason_summary: item.reason_summary || "",
      valid_from: item.valid_from || "",
      valid_to: item.valid_to || "",
      snapshot: baselineSnapshot,
      history: [baselineSnapshot],
      added_at: addedAt
    },
    ...items.filter((entry) => entry.id !== item.id)
  ].slice(0, 50);
  window.localStorage.setItem(WATCHLIST_STORAGE_KEY, JSON.stringify(next));
  emitWatchlistChange();
  return true;
}

export function removeWatchedStock(id) {
  if (!id || !canUseStorage()) {
    return false;
  }
  const items = listWatchedStocks();
  const next = items.filter((item) => item.id !== id);
  window.localStorage.setItem(WATCHLIST_STORAGE_KEY, JSON.stringify(next));
  emitWatchlistChange();
  return true;
}

export function syncWatchedStockSnapshot(id, snapshot, options = {}) {
  if (!id || !canUseStorage()) {
    return false;
  }
  const items = listWatchedStocks();
  const index = items.findIndex((item) => item.id === id);
  if (index < 0) {
    return false;
  }

  const existing = items[index];
  const baseline = buildSnapshot(existing.snapshot || {}, existing.added_at);
  const normalizedHistory = Array.isArray(existing.history)
    ? existing.history.map((entry) => buildSnapshot(entry, entry?.recorded_at || existing.added_at))
    : [baseline];
  const nextSnapshot = buildSnapshot(snapshot);
  const lastSnapshot = normalizedHistory[normalizedHistory.length - 1];
  const nextHistory = isSameSnapshot(lastSnapshot, nextSnapshot)
    ? normalizedHistory
    : [...normalizedHistory, nextSnapshot].slice(-WATCHLIST_HISTORY_LIMIT);

  items[index] = {
    ...existing,
    history: nextHistory
  };

  window.localStorage.setItem(WATCHLIST_STORAGE_KEY, JSON.stringify(items));
  if (options.emit) {
    emitWatchlistChange();
  }
  return !isSameSnapshot(lastSnapshot, nextSnapshot);
}
