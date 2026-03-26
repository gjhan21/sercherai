function normalizeText(value) {
  return String(value || "")
    .replace(/\s+/g, " ")
    .trim();
}

export function normalizeGlobalSearchKeyword(value) {
  return normalizeText(value);
}

export function shouldRequestGlobalSearch(value) {
  return normalizeGlobalSearchKeyword(value).length >= 2;
}

export function buildSearchPageQuery(keyword, options = {}) {
  const query = {};
  const normalizedKeyword = normalizeGlobalSearchKeyword(keyword);
  if (normalizedKeyword) {
    query.q = normalizedKeyword;
  }
  const focusType = normalizeText(options.focusType || options.focus_type);
  const focusID = normalizeText(options.focusID || options.focus_id);
  if (focusType && focusID) {
    query.focus_type = focusType;
    query.focus_id = focusID;
  }
  return query;
}

export function normalizeGlobalSearchItems(items) {
  return Array.isArray(items) ? items : [];
}

export function normalizeGlobalSearchResult(payload, keyword) {
  const result = payload && typeof payload === "object" ? payload : {};
  return {
    keyword: normalizeGlobalSearchKeyword(result.keyword || keyword),
    scope: String(result.scope || "").trim().toUpperCase(),
    stocks: {
      items: normalizeGlobalSearchItems(result?.stocks?.items),
      total: Number(result?.stocks?.total || 0)
    },
    strategies: {
      items: normalizeGlobalSearchItems(result?.strategies?.items),
      total: Number(result?.strategies?.total || 0)
    },
    news: {
      items: normalizeGlobalSearchItems(result?.news?.items),
      total: Number(result?.news?.total || 0)
    }
  };
}

function summarizeText(value) {
  const text = normalizeText(String(value || "").replace(/<[^>]*>/g, " "));
  if (!text) {
    return "暂无摘要";
  }
  if (text.length <= 56) {
    return text;
  }
  return `${text.slice(0, 56)}...`;
}

function formatDate(value) {
  const text = normalizeText(value);
  if (!text) {
    return "";
  }
  const timestamp = Date.parse(text);
  if (Number.isNaN(timestamp)) {
    return "";
  }
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit"
  }).format(new Date(timestamp));
}

function formatDateTime(value) {
  const text = normalizeText(value);
  if (!text) {
    return "";
  }
  const timestamp = Date.parse(text);
  if (Number.isNaN(timestamp)) {
    return "";
  }
  return new Intl.DateTimeFormat("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false
  }).format(new Date(timestamp));
}

function mapRiskLevel(value) {
  const key = String(value || "").trim().toUpperCase();
  if (key === "LOW") return "低风险";
  if (key === "HIGH") return "高风险";
  if (key === "MEDIUM") return "中风险";
  return key || "";
}

function buildStockSearchMeta(item) {
  const pieces = [mapRiskLevel(item?.risk_level)];
  const score = Number(item?.score);
  if (Number.isFinite(score)) {
    pieces.push(`评分 ${score.toFixed(1)}`);
  }
  return pieces.filter(Boolean).join(" · ") || "推荐";
}

function buildStrategySearchMeta(item) {
  const pieces = [normalizeText(item?.direction), mapRiskLevel(item?.risk_level)];
  const validFrom = formatDate(item?.valid_from);
  if (validFrom) {
    pieces.push(validFrom);
  }
  return pieces.filter(Boolean).join(" · ") || "策略";
}

function buildNewsSearchMeta(item) {
  const pieces = [String(item?.visibility || "").trim().toUpperCase() === "VIP" ? "VIP" : "公开"];
  const publishedAt = formatDateTime(item?.published_at);
  if (publishedAt) {
    pieces.push(publishedAt);
  }
  return pieces.join(" · ");
}

export function buildGlobalSearchGroups(result) {
  const payload = normalizeGlobalSearchResult(result);
  return [
    {
      key: "stocks",
      title: "股票推荐",
      total: Number(payload?.stocks?.total || 0),
      emptyText: "当前关键词未命中股票推荐。",
      items: normalizeGlobalSearchItems(payload?.stocks?.items).map((item) => ({
        id: item.id || item.symbol || "",
        title: [item.symbol, item.name].filter(Boolean).join(" ") || "未命名股票",
        summary: item.reason_summary || "暂无推荐理由",
        meta: buildStockSearchMeta(item)
      }))
    },
    {
      key: "strategies",
      title: "期货策略",
      total: Number(payload?.strategies?.total || 0),
      emptyText: "当前关键词未命中期货策略。",
      items: normalizeGlobalSearchItems(payload?.strategies?.items).map((item) => ({
        id: item.id || item.contract || "",
        title: item.contract || item.name || "未命名策略",
        summary: item.reason_summary || "暂无策略摘要",
        meta: buildStrategySearchMeta(item)
      }))
    },
    {
      key: "news",
      title: "资讯研报",
      total: Number(payload?.news?.total || 0),
      emptyText: "当前关键词未命中资讯内容。",
      items: normalizeGlobalSearchItems(payload?.news?.items).map((item) => ({
        id: item.id || "",
        title: item.title || "未命名资讯",
        summary: summarizeText(item.summary || item.content),
        meta: buildNewsSearchMeta(item)
      }))
    }
  ];
}

export function getGlobalSearchTotal(result) {
  return buildGlobalSearchGroups(result).reduce((sum, group) => sum + Number(group.total || 0), 0);
}

export function resolveGlobalSearchScopeLabel(scope) {
  return String(scope || "").trim().toUpperCase() === "USER" ? "登录态全量结果" : "公开结果";
}

export function buildSearchItemRoute(groupKey, item = {}) {
  if (groupKey === "stocks") {
    return { path: "/strategies", query: { stock_id: item.id || "" } };
  }
  if (groupKey === "strategies") {
    return { path: "/strategies", query: { futures_id: item.id || "" } };
  }
  return { path: "/news", query: { article_id: item.id || "" } };
}
