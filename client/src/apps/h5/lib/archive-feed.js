function cleanText(value, fallback = "") {
  const text = String(value || "").replace(/\s+/g, " ").trim();
  return text || fallback;
}

function createSummaryCards(items = []) {
  const leadItem = items[0] || null;

  if (!leadItem) {
    return [
      { label: "历史样本", value: "0 条", note: "公开历史样本会显示在这里" },
      { label: "优先查看", value: "暂无", note: "后续会按时间线展示" },
      { label: "来源标签", value: "待补充", note: "来源说明会跟随档案一起出现" }
    ];
  }

  return [
    { label: "历史样本", value: `${items.length} 条`, note: "成功与失效样本都会保留" },
    { label: "优先查看", value: cleanText(leadItem.status, "待确认"), note: "先看第一条档案的结果与理由" },
    { label: "来源标签", value: cleanText(leadItem.sourceLabel, "待补充"), note: "每条档案都带来源说明" }
  ];
}

function normalizeItem(item = {}) {
  const sourceLabel = cleanText(item.sourceLabel, "待补充");
  return {
    id: item.id,
    headline: cleanText(item.name, "未命名档案"),
    status: cleanText(item.status, "待确认"),
    statusClass: cleanText(item.statusClass, "neutral"),
    sourceLabel,
    sourceClass: cleanText(item.sourceClass, "pending"),
    reason: cleanText(item.reason, "暂无推荐理由"),
    explanationSummary: cleanText(item.explanationSummary, "暂无更多解释"),
    cumulativeReturn: cleanText(item.cumulativeReturn, "-"),
    excessReturn: cleanText(item.excessReturn, "-"),
    date: cleanText(item.date, "-"),
    meta: `${cleanText(item.date, "-")} · ${sourceLabel}`,
    proofTags: Array.isArray(item.proofTags) ? item.proofTags.filter(Boolean).slice(0, 3) : []
  };
}

export function buildArchiveFeedModel({ items = [] } = {}) {
  const normalizedItems = (Array.isArray(items) ? items : [])
    .filter((item) => item && item.id)
    .map(normalizeItem);

  return {
    summaryCards: createSummaryCards(normalizedItems),
    leadItem: normalizedItems[0] || null,
    timeline: normalizedItems,
    readOrder: [
      { title: "先看结果", desc: "先确认状态、累计收益和超额收益。" },
      { title: "再看理由", desc: "查看当时为什么推荐，以及现在怎么解释。" },
      { title: "最后看来源", desc: "结合来源标签判断档案的可信度与适用范围。" }
    ],
    emptyState: {
      title: "当前还没有可展示的历史档案",
      description: "刷新后会按时间顺序展示历史推荐、结果和来源说明。"
    }
  };
}
