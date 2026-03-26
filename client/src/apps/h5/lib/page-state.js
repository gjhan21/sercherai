function cleanValue(value) {
  if (value === undefined || value === null) {
    return undefined;
  }
  if (typeof value === "string") {
    const text = value.trim();
    return text || undefined;
  }
  return value;
}

export function mergeRecordById(records = [], nextRecord = {}) {
  const id = cleanValue(nextRecord?.id);
  if (!id) {
    return Array.isArray(records) ? [...records] : [];
  }

  const list = Array.isArray(records) ? [...records] : [];
  const normalized = Object.entries(nextRecord).reduce((result, [key, value]) => {
    const cleaned = cleanValue(value);
    if (cleaned !== undefined) {
      result[key] = cleaned;
    }
    return result;
  }, { id });

  const index = list.findIndex((item) => String(item?.id || "") === id);
  if (index === -1) {
    list.push(normalized);
    return list;
  }

  list[index] = {
    ...list[index],
    ...normalized
  };
  return list;
}

export function resolveNewsAccessState({ isLoggedIn = false, hasVipAccess = false, visibility = "" } = {}) {
  const needsVip = String(visibility || "").toUpperCase() === "VIP";
  if (!needsVip) {
    return { locked: false, message: "" };
  }
  if (!isLoggedIn) {
    return {
      locked: true,
      message: "请先登录，登录后可以继续阅读正文与附件。"
    };
  }
  if (!hasVipAccess) {
    return {
      locked: true,
      message: "该内容为 VIP 专享，请开通会员后查看完整正文与附件。"
    };
  }
  return { locked: false, message: "" };
}
