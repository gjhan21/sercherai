function toNumber(value) {
  const num = Number(value);
  return Number.isFinite(num) ? num : 0;
}

export function mapActivationStateLabel(state) {
  const source = String(state || "").toUpperCase();
  if (source === "ACTIVE") {
    return "已激活";
  }
  if (source === "PAID_PENDING_KYC") {
    return "待实名激活";
  }
  if (source === "NON_MEMBER") {
    return "未开通";
  }
  return source || "未开通";
}

export function buildMembershipStatusView(profile = {}, quota = {}) {
  const memberLevel = String(profile.member_level || quota.member_level || "FREE").toUpperCase();
  const activationState = String(profile.activation_state || quota.activation_state || "NON_MEMBER").toUpperCase();
  const kycStatus = String(profile.kyc_status || quota.kyc_status || "NONE").toUpperCase();

  let badge = mapActivationStateLabel(activationState);
  let title = "当前未开通会员";
  let description = "登录后可查看套餐、订单与激活状态。";

  if (activationState === "PAID_PENDING_KYC") {
    title = "会员已开通，待实名激活高级权益";
    description = "你已完成开通，但实名状态尚未完成，完成实名认证后可解锁完整深读、附件与高级权益。";
  } else if (activationState === "ACTIVE" || (memberLevel !== "FREE" && kycStatus === "APPROVED")) {
    badge = "已激活";
    title = "会员权益已激活";
    description = "当前可继续查看完整策略解释、研报附件与更深度的历史版本信息。";
  }

  return {
    badge,
    title,
    description,
    stats: [
      { label: "会员层级", value: memberLevel },
      { label: "研报额度", value: String(toNumber(quota.doc_read_remaining)) },
      { label: "资讯额度", value: String(toNumber(quota.news_subscribe_remaining)) }
    ]
  };
}
