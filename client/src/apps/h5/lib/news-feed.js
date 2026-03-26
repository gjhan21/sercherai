function cleanText(value) {
  return String(value || "").replace(/\s+/g, " ").trim();
}

function toPlainText(value) {
  return String(value || "")
    .replace(/<[^>]*>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

export function sortNewsCategories(rawCategories = [], fallbackCategories = []) {
  const source = Array.isArray(rawCategories) && rawCategories.length ? rawCategories : fallbackCategories;
  return source
    .slice()
    .sort((a, b) => Number(a?.sort || 0) - Number(b?.sort || 0))
    .map((item) => ({
      key: item.id,
      label: item.name || "资讯"
    }));
}

export function buildNewsFeedRows(items = []) {
  const rows = Array.isArray(items) ? items : [];
  const lead = rows[0] || null;
  const tickerItems = rows
    .map((item) => cleanText(item?.title))
    .filter(Boolean)
    .slice(0, 3);

  const feedItems = rows.slice(1).map((item) => ({
    id: item.id,
    title: cleanText(item.title) || "未命名资讯",
    summary: cleanText(item.summary) || cleanText(toPlainText(item.content)) || "-",
    meta: cleanText(item.meta) || "-",
    primaryTag: cleanText(item.tags?.[0]) || "资讯",
    accessLabel: cleanText(item.badge) || cleanText(item.tags?.[1]) || cleanText(item.visibility) || "公开",
    tone: cleanText(item.tone) || "brand",
    visibility: cleanText(item.visibility) || "PUBLIC"
  }));

  return {
    lead,
    tickerItems,
    feedItems
  };
}
