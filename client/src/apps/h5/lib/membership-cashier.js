function resolvePlanCycle(durationDays) {
  const days = Number(durationDays || 0);
  if (days >= 330) {
    return "年卡";
  }
  if (days >= 80) {
    return "季卡";
  }
  return "月卡";
}

function normalizePlanName(item, levelLabel) {
  const rawName = String(item?.name || "").trim();
  const looksInternal = !rawName || /\d{6,}|codex|qa|edit-|product/i.test(rawName);
  if (looksInternal) {
    return `${levelLabel} ${resolvePlanCycle(item?.duration_days)}`.trim();
  }
  return rawName;
}

function formatDailyPrice(price, durationDays) {
  const amount = Number(price || 0);
  const days = Number(durationDays || 0);
  if (!Number.isFinite(amount) || !Number.isFinite(days) || amount <= 0 || days <= 0) {
    return "";
  }
  return `约 ¥${(amount / days).toFixed(2)} / 天`;
}

function resolveSceneLabel(durationDays) {
  const days = Number(durationDays || 0);
  if (days >= 330) {
    return "适合长期持有深读与年度复盘";
  }
  if (days >= 80) {
    return "适合持续跟踪主线与资讯深读";
  }
  return "适合先体验完整会员能力";
}

export function buildMembershipCashierModel({
  products = [],
  quota = {},
  mapMemberLevel = (value) => value,
  resolveVipStage = () => false
} = {}) {
  const activeProducts = Array.isArray(products) ? products.slice().sort((a, b) => Number(a.price || 0) - Number(b.price || 0)) : [];
  const currentLevel = String(quota?.member_level || "").toUpperCase();
  const isActive = Boolean(resolveVipStage(quota));

  const plans = activeProducts.map((item, index, list) => {
    const level = String(item.member_level || "").toUpperCase();
    const levelLabel = mapMemberLevel(level);
    const current = isActive && currentLevel === level;
    const recommended = !current && list.length > 1 && index === Math.min(1, Math.max(0, list.length - 1));

    return {
      id: item.id,
      name: item.name || levelLabel,
      displayName: normalizePlanName(item, levelLabel),
      levelLabel,
      durationDays: Number(item.duration_days || 0),
      price: Number(item.price || 0),
      dailyPriceText: formatDailyPrice(item.price, item.duration_days),
      sceneLabel: resolveSceneLabel(item.duration_days),
      current,
      recommended,
      badge: current ? "当前方案" : recommended ? "推荐方案" : "立即开通",
      actionLabel: current ? "当前方案" : "立即下单",
      features: Array.isArray(item.feature_list) ? item.feature_list.slice(0, 3) : [],
      highlights: Array.isArray(item.feature_list) ? item.feature_list.filter(Boolean).slice(0, 2) : [],
      disabled: current || String(item.status || "").toUpperCase() !== "ACTIVE"
    };
  });

  const spotlightPlan = plans.find((item) => item.recommended) || plans.find((item) => !item.disabled) || plans[0] || null;

  return {
    plans,
    spotlightPlan
  };
}
