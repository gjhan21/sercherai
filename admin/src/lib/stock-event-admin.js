function toArray(value) {
  return Array.isArray(value) ? value : [];
}

export function buildStockEventSubgraphViewModel(payload) {
  const normalized = payload && typeof payload === "object" ? payload : {};
  const nodes = toArray(normalized.nodes).length > 0 ? toArray(normalized.nodes) : toArray(normalized.entities);
  const edges = toArray(normalized.edges).length > 0 ? toArray(normalized.edges) : toArray(normalized.relations);
  return {
    ...normalized,
    nodes,
    edges,
    warning_message: String(normalized.warning_message || "").trim()
  };
}
