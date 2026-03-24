function cleanText(value, fallback = "") {
  const text = String(value || "").replace(/\s+/g, " ").trim();
  return text || fallback;
}

function createSummaryCards(items = []) {
  const leadItem = items[0] || null;

  if (!leadItem) {
    return [
      { label: "已保存", value: "0 项", note: "先从首页或策略页加入关注" },
      { label: "优先跟踪", value: "暂无", note: "保存后这里会出现回访主线" },
      { label: "最新动态", value: "待同步", note: "关注后的变化会集中显示" }
    ];
  }

  return [
    { label: "已保存", value: `${items.length} 项`, note: "本地关注可跨页面继续跟踪" },
    { label: "优先跟踪", value: cleanText(leadItem.status, "持续跟踪"), note: "默认展示第一条已保存对象" },
    {
      label: "最新动态",
      value: cleanText(leadItem.latestNewsTitle, "待同步"),
      note: cleanText(leadItem.latestNewsTime, "稍后刷新")
    }
  ];
}

function normalizeItem(item = {}) {
  return {
    id: item.id,
    headline: cleanText(item.name, "未命名标的"),
    summary: cleanText(item.summary, "等待更多跟踪摘要"),
    status: cleanText(item.status, "持续跟踪"),
    risk: cleanText(item.risk, "待确认"),
    nextAction: cleanText(item.nextAction, "继续观察"),
    performance: cleanText(item.performance, "-"),
    performanceClass: cleanText(item.performanceClass, "flat"),
    proofTags: Array.isArray(item.proofTags) ? item.proofTags.filter(Boolean).slice(0, 3) : [],
    latestNewsTitle: cleanText(item.latestNewsTitle, "暂无最新动态"),
    latestNewsTime: cleanText(item.latestNewsTime, "稍后刷新")
  };
}

export function buildWatchlistFeedModel({ items = [] } = {}) {
  const normalizedItems = (Array.isArray(items) ? items : [])
    .filter((item) => item && item.id)
    .map(normalizeItem);

  return {
    summaryCards: createSummaryCards(normalizedItems),
    leadItem: normalizedItems[0] || null,
    rows: normalizedItems.slice(1),
    emptyState: {
      title: "还没有保存任何关注对象",
      description: "可以先从首页或策略页加入关注，再回到这里持续跟踪变化。"
    }
  };
}
