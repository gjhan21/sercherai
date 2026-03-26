import { buildMembershipStatusView } from "./membership.js";

export function buildProfileSummary(profile = {}, quota = {}) {
  return {
    phone: profile.phone || "-",
    email: profile.email || "未填写",
    memberLevel: profile.member_level || quota.member_level || "FREE",
    kycStatus: profile.kyc_status || quota.kyc_status || "NONE",
    activationState: profile.activation_state || quota.activation_state || "NON_MEMBER",
    inviteCode: profile.invite_code || "-",
    registrationSource: profile.registration_source || "-",
    membership: buildMembershipStatusView(profile, quota)
  };
}
