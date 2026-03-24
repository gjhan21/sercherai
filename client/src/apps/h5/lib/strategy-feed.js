function cleanText(value, fallback = "") {
  const text = String(value || "").replace(/\s+/g, " ").trim();
  return text || fallback;
}

function normalizeLeadRow(item = {}, sectionLabel = "策略") {
  return {
    id: item.id,
    title: cleanText(item.title, "未命名策略"),
    summary: cleanText(item.summary, "等待更多观点摘要"),
    meta: cleanText(item.meta, "-"),
    signalLabel: cleanText(item.badge, "待跟踪"),
    tone: cleanText(item.tone, "brand"),
    sectionLabel: cleanText(sectionLabel, "策略"),
    tags: Array.isArray(item.tags) ? item.tags.filter(Boolean).slice(0, 3) : []
  };
}

export function buildStrategyFeedModel(rows = [], sectionLabel = "策略") {
  const list = Array.isArray(rows) ? rows : [];
  const lead = list[0] ? normalizeLeadRow(list[0], sectionLabel) : null;
  const digest = list
    .map((item) => cleanText(item?.title))
    .filter(Boolean)
    .slice(0, 3);

  const stream = list.slice(1).map((item) => ({
    id: item.id,
    title: cleanText(item.title, "未命名策略"),
    summary: cleanText(item.summary, "等待更多观点摘要"),
    meta: cleanText(item.meta, "-"),
    signalLabel: cleanText(item.badge, "待跟踪"),
    tone: cleanText(item.tone, "brand"),
    kicker: cleanText(sectionLabel, "策略"),
    tags: Array.isArray(item.tags) ? item.tags.filter(Boolean).slice(0, 3) : []
  }));

  return {
    lead,
    digest,
    stream
  };
}
