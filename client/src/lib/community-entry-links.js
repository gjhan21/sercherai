const TOPIC_TYPES = new Set(["STOCK", "FUTURES", "NEWS", "STRATEGY"]);
const TARGET_TYPES = new Set(["STOCK", "FUTURES", "NEWS_ARTICLE", "STRATEGY_ITEM"]);
const STANCES = new Set(["BULLISH", "BEARISH", "WATCH"]);
const TIME_HORIZONS = new Set(["SHORT", "SWING", "MID", "LONG"]);
const SORT_VALUES = new Set(["MOST_ACTIVE", "LATEST"]);

function normalizeText(value) {
  return String(value || "")
    .replace(/\s+/g, " ")
    .trim();
}

function normalizeEnum(value, allowedValues) {
  const normalized = normalizeText(value).toUpperCase();
  if (!normalized || !allowedValues.has(normalized)) {
    return "";
  }
  return normalized;
}

function buildDraftQuery({
  topicType = "",
  targetType = "",
  targetID = "",
  targetSnapshot = "",
  title = "",
  stance = "",
  timeHorizon = ""
} = {}) {
  const query = {};
  const normalizedTopicType = normalizeEnum(topicType, TOPIC_TYPES);
  const normalizedTargetType = normalizeEnum(targetType, TARGET_TYPES);
  const normalizedTargetID = normalizeText(targetID);
  const normalizedTargetSnapshot = normalizeText(targetSnapshot);
  const normalizedTitle = normalizeText(title);
  const normalizedStance = normalizeEnum(stance, STANCES);
  const normalizedTimeHorizon = normalizeEnum(timeHorizon, TIME_HORIZONS);

  if (normalizedTopicType) {
    query.topic_type = normalizedTopicType;
  }
  if (normalizedTargetType) {
    query.target_type = normalizedTargetType;
  }
  if (normalizedTargetID) {
    query.target_id = normalizedTargetID;
  }
  if (normalizedTargetSnapshot) {
    query.target_snapshot = normalizedTargetSnapshot;
  }
  if (normalizedTitle) {
    query.title = normalizedTitle;
  }
  if (normalizedStance) {
    query.stance = normalizedStance;
  }
  if (normalizedTimeHorizon) {
    query.time_horizon = normalizedTimeHorizon;
  }

  return query;
}

export function buildCommunityListRoute({
  topicType = "",
  sort = "",
  entrySource = "",
  targetType = "",
  targetID = "",
  targetSnapshot = "",
  title = "",
  stance = "",
  timeHorizon = ""
} = {}) {
  const query = buildDraftQuery({
    topicType,
    targetType,
    targetID,
    targetSnapshot,
    title,
    stance,
    timeHorizon
  });
  const normalizedSort = normalizeEnum(sort, SORT_VALUES);
  const normalizedEntrySource = normalizeText(entrySource);

  if (normalizedSort) {
    query.sort = normalizedSort;
  }
  if (normalizedEntrySource) {
    query.entry_source = normalizedEntrySource;
  }

  if (Object.keys(query).length === 0) {
    return { path: "/community" };
  }
  return { path: "/community", query };
}

export function buildCommunityComposeRoute(options = {}) {
  const query = buildDraftQuery(options);
  if (Object.keys(query).length === 0) {
    return { path: "/community/new" };
  }
  return { path: "/community/new", query };
}

export function buildCommunityComposeRouteFromQuery(query = {}) {
  return buildCommunityComposeRoute({
    topicType: query.topic_type,
    targetType: query.target_type,
    targetID: query.target_id,
    targetSnapshot: query.target_snapshot,
    title: query.title,
    stance: query.stance,
    timeHorizon: query.time_horizon
  });
}

export function resolveCommunityEntryContext(query = {}) {
  const topicType = normalizeEnum(query.topic_type, TOPIC_TYPES);
  const targetType = normalizeEnum(query.target_type, TARGET_TYPES);
  const targetID = normalizeText(query.target_id);
  const targetSnapshot = normalizeText(query.target_snapshot);
  const title = normalizeText(query.title);
  const stance = normalizeEnum(query.stance, STANCES);
  const timeHorizon = normalizeEnum(query.time_horizon, TIME_HORIZONS);
  const entrySource = normalizeText(query.entry_source);
  const sort = normalizeEnum(query.sort, SORT_VALUES);

  return {
    topicType,
    targetType,
    targetID,
    targetSnapshot,
    title,
    stance,
    timeHorizon,
    entrySource,
    sort,
    hasContext: Boolean(topicType || targetType || targetID || targetSnapshot || title)
  };
}

export function buildLinkedContentRoute({ targetType = "", targetID = "" } = {}) {
  const normalizedTargetType = normalizeEnum(targetType, TARGET_TYPES);
  const normalizedTargetID = normalizeText(targetID);

  if (normalizedTargetType === "NEWS_ARTICLE" && normalizedTargetID) {
    return {
      path: "/news",
      query: {
        article_id: normalizedTargetID
      }
    };
  }

  if (normalizedTargetType === "STOCK" && normalizedTargetID) {
    return {
      path: "/strategies",
      query: {
        stock_id: normalizedTargetID
      }
    };
  }

  if (normalizedTargetType === "FUTURES" && normalizedTargetID) {
    return {
      path: "/strategies",
      query: {
        futures_id: normalizedTargetID
      }
    };
  }

  return normalizedTargetType === "STRATEGY_ITEM" ? { path: "/strategies" } : { path: "/community" };
}

export function findNewsArticleLocation(feedMap = {}, requestedArticleID = "") {
  const normalizedRequestedID = normalizeText(requestedArticleID);
  if (!normalizedRequestedID) {
    return null;
  }

  for (const categoryKey of Object.keys(feedMap || {})) {
    const items = Array.isArray(feedMap?.[categoryKey]) ? feedMap[categoryKey] : [];
    const matched = items.find((item) => normalizeText(item?.id) === normalizedRequestedID);
    if (matched) {
      return {
        categoryKey: normalizeText(matched.categoryKey) || categoryKey,
        articleID: normalizedRequestedID
      };
    }
  }

  return null;
}

export function findEntityIDByTarget(items = [], requestedTarget = "", keys = ["id"]) {
  const normalizedRequestedTarget = normalizeText(requestedTarget);
  if (!normalizedRequestedTarget) {
    return "";
  }

  const normalizedKeys = Array.isArray(keys) ? keys : ["id"];
  const matched = (items || []).find((item) =>
    normalizedKeys.some((key) => normalizeText(item?.[key]) === normalizedRequestedTarget)
  );
  return normalizeText(matched?.id);
}
