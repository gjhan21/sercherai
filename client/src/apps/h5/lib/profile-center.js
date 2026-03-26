import {
  formatDateTime,
  mapActivationState,
  mapInviteStatus,
  mapKYCStatus,
  mapMemberLevel,
  mapMessageType,
  mapPaymentStatus,
  mapVIPStatus,
  resolveVipStage
} from "./formatters.js";

function cleanText(value, fallback = "") {
  const text = String(value || "").replace(/\s+/g, " ").trim();
  return text || fallback;
}

function maskContact(profile = {}) {
  const phone = String(profile?.phone || "").trim();
  if (/^\d{11}$/.test(phone)) {
    return `${phone.slice(0, 3)}****${phone.slice(-4)}`;
  }
  const email = String(profile?.email || "").trim();
  if (email.includes("@")) {
    const [name, domain] = email.split("@");
    if (name.length <= 2) {
      return `${name[0] || "*"}***@${domain}`;
    }
    return `${name.slice(0, 2)}***@${domain}`;
  }
  return "我的账户";
}

function getUnreadCount(messages = []) {
  return messages.filter((item) => String(item?.read_status || "").toUpperCase() !== "READ").length;
}

function getLatestMessage(messages = []) {
  return messages[0] || null;
}

function getLatestOrder(orders = []) {
  return orders[0] || null;
}

function buildHero({ profile = {}, quota = {}, unreadCount = 0, inviteSummary = {} }) {
  const memberLevel = mapMemberLevel(profile?.member_level, quota?.member_level);
  const activationState = mapActivationState(quota?.activation_state || profile?.activation_state);
  const kycStatus = mapKYCStatus(profile?.kyc_status || quota?.kyc_status);
  const registeredCount = Number(inviteSummary?.last_7d_registered_count || 0);

  return {
    displayName: maskContact(profile),
    memberLevel,
    activationState,
    vipStatus: mapVIPStatus(quota?.vip_status, quota?.member_level || profile?.member_level),
    description: activationState === "待实名激活"
      ? "支付已经完成，补齐实名后会自动激活当前高级权益。"
      : resolveVipStage(quota)
        ? "账户状态已同步，常用入口、消息提醒和邀请关系都集中在这里。"
        : "先确认会员状态，再继续查看消息、订单和邀请转化。",
    metrics: [
      {
        label: "实名",
        value: kycStatus,
        note: activationState === "待实名激活" ? "支付后待完成" : "账户安全基础"
      },
      {
        label: "消息",
        value: `${unreadCount} 条`,
        note: unreadCount > 0 ? "未读待处理" : "暂无未读"
      },
      {
        label: "邀请",
        value: `${registeredCount} 人`,
        note: "7日注册"
      }
    ]
  };
}

function buildTodos({ quota = {}, unreadCount = 0, inviteSummary = {} }) {
  const activationState = String(quota?.activation_state || "").toUpperCase();
  const isVipActive = resolveVipStage(quota);
  const todos = [];

  if (activationState === "PAID_PENDING_KYC") {
    todos.push({
      id: "kyc",
      title: "完成实名激活",
      desc: "当前订单已支付，下一步优先提交实名，审核通过后高级权益自动生效。",
      badge: "高优先级",
      tone: "gold",
      actionLabel: "去实名"
    });
  } else {
    todos.push({
      id: "membership",
      title: isVipActive ? "确认当前会员权益" : "先开通会员权益",
      desc: isVipActive ? "检查有效期、激活状态和当前可继续使用的权限。" : "先确认适合的套餐与当前待完成动作。",
      badge: isVipActive ? "会员中心" : "优先处理",
      tone: isVipActive ? "brand" : "gold",
      actionLabel: "去会员页"
    });
  }

  todos.push({
    id: "messages",
    title: unreadCount > 0 ? "处理未读消息" : "查看最新提醒",
    desc: unreadCount > 0 ? "系统通知、策略提醒和风险告警已经聚合到消息区。" : "账户和策略提醒会集中显示在消息列表中。",
    badge: unreadCount > 0 ? "未读提醒" : "消息中心",
    tone: unreadCount > 0 ? "gold" : "brand",
    actionLabel: "看消息"
  });

  todos.push({
    id: "invite",
    title: "维护邀请关系",
    desc: Number(inviteSummary?.share_link_count || 0) > 0
      ? "继续查看近 7 天注册与首付转化，并维护已有分享链接。"
      : "先创建邀请码，再跟踪注册和首付转化。",
    badge: "邀请中心",
    tone: "brand",
    actionLabel: "看邀请"
  });

  return todos;
}

function buildServiceCards({ quota = {}, orders = [], messages = [], inviteSummary = {}, shareLinks = [] }) {
  const latestOrder = getLatestOrder(orders);
  const latestMessage = getLatestMessage(messages);
  const unreadCount = getUnreadCount(messages);
  const latestMessageTitle = cleanText(latestMessage?.title, mapMessageType(latestMessage?.type));

  return [
    {
      id: "membership",
      title: "会员与支付",
      summary: latestOrder
        ? `最近一笔会员订单${mapPaymentStatus(latestOrder?.status)}，可继续查看激活与权益状态`
        : "查看当前会员权益、支付状态和升级入口",
      tags: [
        mapMemberLevel(quota?.member_level, quota?.member_level),
        mapVIPStatus(quota?.vip_status, quota?.member_level),
        mapActivationState(quota?.activation_state)
      ].filter(Boolean)
    },
    {
      id: "messages",
      title: "消息与提醒",
      summary: latestMessage
        ? `未读 ${unreadCount} 条，最近一条来自${latestMessageTitle || "消息中心"}`
        : "未读消息、策略提醒和风险通知会显示在这里",
      tags: [
        unreadCount > 0 ? `${unreadCount} 条未读` : "消息已读",
        latestMessage ? mapMessageType(latestMessage.type) : "通知中心"
      ]
    },
    {
      id: "invite",
      title: "邀请与转化",
      summary: shareLinks.length
        ? `已创建 ${shareLinks.length} 个分享入口，近 7 日注册 ${Number(inviteSummary?.last_7d_registered_count || 0)} 人`
        : "先创建邀请码，再追踪近 7 日注册和首付转化",
      tags: [
        `${Number(inviteSummary?.last_7d_registered_count || 0)} 人注册`,
        `${Number(inviteSummary?.last_7d_first_paid_count || 0)} 人首付`
      ]
    }
  ];
}

function buildMessageCards(messages = []) {
  return messages.slice(0, 6).map((item) => ({
    id: item.id,
    title: cleanText(item.title, mapMessageType(item.type)),
    desc: cleanText(item.content, "-"),
    time: formatDateTime(item.created_at),
    read: String(item.read_status || "").toUpperCase() === "READ",
    typeLabel: mapMessageType(item.type),
    raw: item
  }));
}

function buildInviteCards(inviteRecords = []) {
  return inviteRecords.slice(0, 4).map((item) => ({
    id: item.id,
    title: cleanText(item.invitee_user_id, "邀请记录"),
    time: formatDateTime(item.register_at || item.first_pay_at),
    status: mapInviteStatus(item.status),
    desc: String(item.status || "").toUpperCase() === "FIRST_PAID"
      ? "已完成首单转化，可继续跟踪后续留存。"
      : "已完成注册，等待进一步付费转化。"
  }));
}

function buildInviteOverview({ inviteSummary = {}, shareLinks = [] }) {
  const primaryLink = shareLinks[0] || null;

  return {
    primaryCode: cleanText(primaryLink?.invite_code, "暂未创建邀请码"),
    primaryLink: cleanText(primaryLink?.url, "创建后这里会显示分享链接"),
    summary: `近 7 日注册 ${Number(inviteSummary?.last_7d_registered_count || 0)} 人 · 首付 ${Number(inviteSummary?.last_7d_first_paid_count || 0)} 人`,
    note: Number(inviteSummary?.share_link_count || 0) > 0
      ? `当前已有 ${Number(inviteSummary?.share_link_count || 0)} 个分享入口`
      : "还没有邀请码，建议先创建一个常用分享入口"
  };
}

function buildSticky({ quota = {}, unreadCount = 0 }) {
  const activationState = String(quota?.activation_state || "").toUpperCase();
  if (activationState === "PAID_PENDING_KYC") {
    return {
      title: "先完成实名激活，再继续账户操作",
      description: "审核通过后，高级权益会自动生效。",
      primaryLabel: "提交实名",
      primaryTarget: "kyc"
    };
  }
  if (!resolveVipStage(quota)) {
    return {
      title: "先确认会员状态，再处理其他账户动作",
      description: "会员权益、支付和激活状态已经集中到一个入口。",
      primaryLabel: "查看方案",
      primaryTarget: "membership"
    };
  }
  if (unreadCount > 0) {
    return {
      title: "账户已同步，优先处理未读提醒",
      description: "策略提醒、系统消息和风险通知都在消息区。",
      primaryLabel: "处理消息",
      primaryTarget: "messages"
    };
  }
  return {
    title: "账户状态已同步，可继续查看消息和邀请",
    description: "高频入口已经前置到首屏，常见动作都能一跳完成。",
    primaryLabel: "刷新账户",
    primaryTarget: "refresh"
  };
}

export function buildProfileCenterModel({
  profile = {},
  quota = {},
  orders = [],
  messages = [],
  shareLinks = [],
  inviteRecords = [],
  inviteSummary = {}
} = {}) {
  const unreadCount = getUnreadCount(messages);

  return {
    hero: buildHero({ profile, quota, unreadCount, inviteSummary }),
    todos: buildTodos({ quota, unreadCount, inviteSummary }),
    shortcuts: [
      { id: "membership", title: "会员中心", note: "权益与支付" },
      { id: "strategies", title: "策略观点", note: "推荐与风险" },
      { id: "news", title: "资讯正文", note: "研报与附件" },
      { id: "invite", title: "邀请中心", note: "分享与转化" }
    ],
    serviceCards: buildServiceCards({ quota, orders, messages, inviteSummary, shareLinks }),
    messageCards: buildMessageCards(messages),
    inviteCards: buildInviteCards(inviteRecords),
    inviteOverview: buildInviteOverview({ inviteSummary, shareLinks }),
    shareLinks: shareLinks.slice(0, 2).map((item) => ({
      id: item.id,
      code: cleanText(item.invite_code, "邀请码"),
      url: cleanText(item.url, "-")
    })),
    sticky: buildSticky({ quota, unreadCount })
  };
}
