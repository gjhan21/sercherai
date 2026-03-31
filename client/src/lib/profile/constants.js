export const modules = [
  { key: "vip", label: "VIP情况" },
  { key: "payment", label: "支付情况" },
  { key: "reading", label: "阅读情况" },
  { key: "subscription", label: "订阅情况" },
  { key: "message", label: "通知情况" },
  { key: "invite", label: "邀请关系" },
  { key: "other", label: "其他信息" }
];

export const timeRanges = ["近7天", "近30天", "全部"];

export const rangeDayMap = {
  "近7天": 7,
  "近30天": 30,
  "全部": null
};

export const shareChannelOptions = [
  { value: "APP", label: "App内分享" },
  { value: "WECHAT", label: "微信" },
  { value: "WEIBO", label: "微博" },
  { value: "QQ", label: "QQ" }
];

export const subscriptionTypeOptions = [
  { value: "STOCK_RECO", label: "股票推荐" },
  { value: "FUTURES_STRATEGY", label: "期货策略" },
  { value: "ARBITRAGE", label: "套利信号" },
  { value: "EVENT", label: "事件提醒" }
];

export const subscriptionFrequencyOptions = [
  { value: "INSTANT", label: "实时" },
  { value: "DAILY", label: "每日" },
  { value: "WEEKLY", label: "每周" }
];

export const menus = [
  { title: "账户与安全", desc: "密码、登录设备、验证方式配置" },
  { title: "VIP权益中心", desc: "查看权益明细、续费记录和到期提醒" },
  { title: "支付记录导出", desc: "按时间范围导出订单与支付流水" },
  { title: "阅读偏好设置", desc: "资讯、研报和期刊推送策略配置" }
];
