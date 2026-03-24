export function resolveSurfaceToneClasses(tone = "default") {
  if (tone === "soft") {
    return ["h5-card-soft"];
  }
  if (tone === "accent") {
    return ["h5-card-accent"];
  }
  if (tone === "brand") {
    return ["h5-card-brand"];
  }
  if (tone === "hero") {
    return ["h5-card-brand", "h5-card-hero"];
  }
  return [];
}

export function resolveMetricToneClasses(tone = "default") {
  if (tone === "soft") {
    return ["h5-metric-card-soft"];
  }
  if (tone === "brand") {
    return ["h5-metric-card-brand"];
  }
  return [];
}

export function resolveHighlightTone(index, { emphasizeFirst = true } = {}) {
  if (!emphasizeFirst) {
    return "default";
  }
  if (index === 0) {
    return "brand";
  }
  if (index === 1) {
    return "soft";
  }
  return "default";
}

export function resolveStickyActionMode({
  primaryLabel = "",
  secondaryLabel = ""
} = {}) {
  return primaryLabel && secondaryLabel ? "stacked" : "single";
}
