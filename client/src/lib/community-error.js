export function normalizeCommunityLoadError(message) {
  const text = String(message || "").trim();
  if (!text) {
    return "讨论列表加载失败";
  }

  const lowered = text.toLowerCase();
  if (
    lowered.includes("discussion_topics") ||
    lowered.includes("discussion_comments") ||
    lowered.includes("error 1146") ||
    lowered.includes("doesn't exist")
  ) {
    return "社区讨论数据尚未初始化，当前先展示入口和说明。";
  }

  return text;
}
