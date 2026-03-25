function normalizeOrigin(value) {
  const trimmed = String(value || "").trim();
  if (!trimmed) {
    return "";
  }
  return trimmed.replace(/\/+$/, "");
}

export function resolveClientOrigin({ envOrigin = "", locationOrigin = "" } = {}) {
  const explicitOrigin = normalizeOrigin(envOrigin);
  if (explicitOrigin) {
    return explicitOrigin;
  }

  const currentOrigin = normalizeOrigin(locationOrigin);
  if (!currentOrigin) {
    return "";
  }

  let parsed;
  try {
    parsed = new URL(currentOrigin);
  } catch {
    return "";
  }

  if (parsed.port === "5174") {
    parsed.port = "5175";
    return normalizeOrigin(parsed.toString());
  }
  if (parsed.port === "4174") {
    parsed.port = "4175";
    return normalizeOrigin(parsed.toString());
  }
  if (
    parsed.port &&
    (parsed.hostname === "127.0.0.1" ||
      parsed.hostname === "localhost" ||
      parsed.hostname.endsWith(".local"))
  ) {
    parsed.port = "5175";
    return normalizeOrigin(parsed.toString());
  }

  return currentOrigin;
}

export function buildCommunityTopicURL({
  topicID = "",
  commentID = "",
  envOrigin = "",
  locationOrigin = ""
} = {}) {
  const origin = resolveClientOrigin({ envOrigin, locationOrigin });
  const normalizedTopicID = String(topicID || "").trim();
  if (!origin || !normalizedTopicID) {
    return "";
  }

  const topicPath = `/community/topics/${encodeURIComponent(normalizedTopicID)}`;
  const normalizedCommentID = String(commentID || "").trim();
  const commentAnchor = normalizedCommentID
    ? `#comment-${encodeURIComponent(normalizedCommentID)}`
    : "";
  return `${origin}${topicPath}${commentAnchor}`;
}
