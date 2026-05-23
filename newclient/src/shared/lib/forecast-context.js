function clean(value) {
  return String(value || "").trim();
}

export function buildForecastContextQuery({
  targetType = "STOCK",
  targetId = "",
  targetKey = "",
  targetLabel = "",
  source = "",
  sourceId = "",
  sourcePath = "",
  from = "",
  strategyId = ""
} = {}) {
  const query = {
    target_type: clean(targetType || "STOCK").toUpperCase() || "STOCK"
  };
  if (clean(targetId)) query.target_id = clean(targetId);
  if (clean(targetKey)) {
    query.target_key = clean(targetKey).toUpperCase();
    query.symbol = clean(targetKey).toUpperCase();
  }
  if (clean(targetLabel)) {
    query.target_label = clean(targetLabel);
    query.name = clean(targetLabel);
  }
  if (clean(source)) query.source = clean(source).toUpperCase();
  if (clean(sourceId)) query.source_id = clean(sourceId);
  if (clean(sourcePath)) query.source_path = clean(sourcePath);
  if (clean(from)) query.from = clean(from);
  if (clean(strategyId)) query.strategy_id = clean(strategyId);
  return query;
}

export function resolveForecastContextFromRoute(route) {
  const query = route?.query || {};
  return {
    targetId: clean(query.target_id),
    targetKey: clean(query.target_key || query.symbol).toUpperCase(),
    targetLabel: clean(query.target_label || query.name),
    targetType: clean(query.target_type || "STOCK").toUpperCase() || "STOCK",
    source: clean(query.source).toUpperCase(),
    sourceId: clean(query.source_id),
    sourcePath: clean(query.source_path),
    from: clean(query.from),
    strategyId: clean(query.strategy_id)
  };
}

export function canLaunchForecastContext(context) {
  return Boolean(clean(context?.targetId) && clean(context?.source) && clean(context?.sourcePath));
}
