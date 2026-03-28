const TAB_ITEMS = [
  { path: "/home", label: "首页", icon: "home" },
  { path: "/news", label: "资讯", icon: "news" },
  { path: "/strategies", label: "策略", icon: "insight" },
  { path: "/membership", label: "会员", icon: "vip" },
  { path: "/profile", label: "我的", icon: "user" }
];

function toURL(input = "") {
  const value = String(input || "").trim();
  try {
    return new URL(value.startsWith("/") ? value : `/${value}`, "https://sercherai.local");
  } catch {
    return new URL("/home", "https://sercherai.local");
  }
}

export function resolveTabItems() {
  return TAB_ITEMS;
}

export function resolveShellScene(path) {
  const targetURL = toURL(path);
  const pathname = targetURL.pathname;
  const section = String(targetURL.searchParams.get("section") || "").trim().toLowerCase();

  if (pathname.startsWith("/watchlist")) {
    return {
      section: "我的",
      title: "我的关注详情",
      subtitle: "关注详情从个人中心进入，继续查看变化工作台",
      pulse: "二级模块"
    };
  }

  if (pathname.startsWith("/archive")) {
    return {
      section: "档案",
      title: "历史档案",
      subtitle: "历史样本、结果和版本变化统一按时间线查看",
      pulse: "复盘"
    };
  }

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
    if (section === "watchlist") {
      return {
        section: "我的",
        title: "我的关注",
        subtitle: "个人中心先承接关注模块，再进入详情继续回访",
        pulse: "模块聚焦"
      };
    }
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
