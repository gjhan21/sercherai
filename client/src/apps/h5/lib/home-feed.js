export function buildHomeFeedModel({ stocks = [], article = null } = {}) {
  const list = Array.isArray(stocks) ? stocks : [];
  const leadStock = list[0] || null;

  const marketPulse = leadStock ? [
    { label: "主线风险", value: leadStock.risk || "待确认" },
    { label: "建议仓位", value: leadStock.position || "轻仓观察" },
    { label: "目标区间", value: leadStock.takeProfit || "目标待补充" }
  ] : [];

  const watchlistPreview = list.slice(1, 4).map((item) => ({
    id: item.id,
    title: item.title || "未命名标的",
    summary: item.summary || "等待更多跟踪摘要",
    meta: item.meta || "-",
    risk: item.risk || "待确认",
    tags: Array.isArray(item.tags) ? item.tags.filter(Boolean).slice(0, 3) : []
  }));

  const newsBrief = article ? {
    id: article.id,
    title: article.title || "未命名资讯",
    summary: article.summary || "等待更多资讯摘要"
  } : null;

  return {
    leadStock,
    marketPulse,
    watchlistPreview,
    newsBrief
  };
}
