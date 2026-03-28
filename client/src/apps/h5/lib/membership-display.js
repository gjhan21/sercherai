import { mapMemberLevel, mapVIPStatus, resolveVipStage } from "./formatters.js";

export function resolveHomeMembershipSummary(quota = {}, { isLoggedIn = false, loading = false } = {}) {
  if (isLoggedIn && loading && !Object.keys(quota || {}).length) {
    return {
      value: "同步中",
      note: "正在确认会员状态"
    };
  }

  return {
    value: mapMemberLevel(quota?.member_level, quota?.member_level),
    note: resolveVipStage(quota) ? mapVIPStatus(quota?.vip_status, quota?.member_level) : "未开通或待激活"
  };
}

export function resolveStrategyAccessSummary(quota = {}, { isLoggedIn = false, loading = false } = {}) {
  if (isLoggedIn && loading && !Object.keys(quota || {}).length) {
    return {
      value: "同步中",
      note: "正在确认会员能力"
    };
  }

  if (resolveVipStage(quota)) {
    return {
      value: "会员已同步",
      note: "可继续深读更多内容"
    };
  }

  return {
    value: "普通浏览",
    note: "升级后可串联更多正文与附件"
  };
}
