const TAB_ITEMS = [
  { path: "/home", label: "首页", icon: "home" },
  { path: "/news", label: "资讯", icon: "news" },
  { path: "/strategies", label: "策略", icon: "insight" },
  { path: "/membership", label: "会员", icon: "vip" },
  { path: "/profile", label: "我的", icon: "user" }
];

function normalizePath(input = "") {
  const value = String(input || "");
  const [pathname] = value.split("?");
  return pathname || "/home";
}

export function resolveTabItems() {
  return TAB_ITEMS;
}

export function resolveShellScene(path) {
  const pathname = normalizePath(path);

  if (pathname.startsWith("/news")) {
    return {
      section: "资讯",
      title: "市场资讯",
      subtitle: "栏目切换后直接进入正文与内容流",
      pulse: "阅读中"
    };
  }

  if (pathname.startsWith("/strategies")) {
    return {
      section: "策略",
      title: "精选观点",
      subtitle: "像内容 App 一样查看结论、理由和风险边界",
      pulse: "观点流"
    };
  }

  if (pathname.startsWith("/membership")) {
    return {
      section: "会员",
      title: "会员中心",
      subtitle: "套餐、支付与激活状态集中在一页完成",
      pulse: "收银台"
    };
  }

  if (pathname.startsWith("/profile")) {
    return {
      section: "我的",
      title: "账户中心",
      subtitle: "消息、订单、实名和会员状态统一管理",
      pulse: "账户"
    };
  }

  return {
    section: "首页",
    title: "今日观点",
    subtitle: "先看核心判断，再顺着内容流继续阅读",
    pulse: "内容优先"
  };
}
