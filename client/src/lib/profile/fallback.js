export const fallbackProfile = {
  id: "u_1772286896732911000",
  phone: "13800138000",
  email: "demo@sercherai.com",
  kyc_status: "APPROVED",
  member_level: "VIP3",
  status: "ACTIVE",
  activation_state: "ACTIVE",
  vip_expire_at: "2026-03-20T23:59:59+08:00",
  vip_status: "ACTIVE"
};

export const fallbackQuota = {
  member_level: "VIP3",
  kyc_status: "APPROVED",
  activation_state: "ACTIVE",
  period_key: "2026-02",
  doc_read_limit: 100,
  doc_read_used: 24,
  doc_read_remaining: 76,
  news_subscribe_limit: 50,
  news_subscribe_used: 12,
  news_subscribe_remaining: 38,
  reset_cycle: "MONTHLY",
  reset_at: "2026-03-01T00:00:00+08:00",
  vip_expire_at: "2026-03-20T23:59:59+08:00",
  vip_status: "ACTIVE",
  vip_remaining_days: 20
};

export const fallbackMembershipOrders = [
  {
    id: "mo_demo_001",
    order_no: "mo_demo_001",
    user_id: "u_demo_001",
    product_id: "mp_demo_001",
    amount: 99,
    pay_channel: "ALIPAY",
    status: "PAID",
    paid_at: "2026-02-24T11:00:00+08:00",
    created_at: "2026-02-24T10:50:00+08:00"
  }
];

export const fallbackRechargeRecords = [
  {
    id: "rc_demo_001",
    order_no: "O20260224001",
    amount: 99,
    pay_channel: "ALIPAY",
    status: "PAID",
    paid_at: "2026-02-24T11:00:00+08:00",
    remark: ""
  }
];

export const fallbackBrowseHistory = [
  {
    id: "h_demo_001",
    content_type: "NEWS",
    content_id: "news_001",
    title: "政策观察",
    source_page: "/news",
    viewed_at: "2026-02-24T10:00:00+08:00"
  }
];

export const fallbackSubscriptions = [
  {
    id: "sub_demo_001",
    type: "STOCK_RECO",
    scope: "ALL",
    frequency: "DAILY",
    status: "ACTIVE"
  }
];

export const fallbackMessages = [
  {
    id: "msg_demo_001",
    title: "策略提醒",
    content: "今日波动率升高，建议把初始仓位控制在计划仓位 70% 以内。",
    type: "STRATEGY",
    read_status: "UNREAD",
    created_at: "2026-02-27T10:20:00+08:00"
  }
];

export const fallbackShareLinks = [
  {
    id: "sl_demo_001",
    invite_code: "DEMO2026",
    url: "https://sercherai.local/invite/DEMO2026",
    channel: "APP",
    status: "ACTIVE",
    expired_at: ""
  }
];

export const fallbackInviteRecords = [
  {
    id: "ir_demo_001",
    invitee_user_id: "u_demo_002",
    status: "REGISTERED",
    register_at: "2026-02-24T10:00:00+08:00",
    first_pay_at: "",
    risk_flag: "NORMAL"
  }
];

export const fallbackInviteSummary = {
  share_link_count: 1,
  registered_count: 1,
  first_paid_count: 0,
  conversion_rate: 0,
  last_7d_registered_count: 1,
  last_7d_first_paid_count: 0,
  last_7d_conversion_rate: 0,
  last_30d_registered_count: 1,
  last_30d_first_paid_count: 0,
  last_30d_conversion_rate: 0
};
