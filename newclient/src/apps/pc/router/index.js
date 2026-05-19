import { createRouter, createWebHistory } from "vue-router";

const routes = [
  // Home
  { path: "/", name: "home", component: () => import("../views/Home.vue"), meta: { title: "智投首页", section: "home", icon: "home" } },

  // Module 1: Recommendations
  { path: "/recommendations", name: "daily-recs", component: () => import("../views/recommendations/DailyRecs.vue"), meta: { title: "每日推荐", section: "recs", icon: "star", sub: "daily" } },
  { path: "/recommendations/history", name: "rec-history", component: () => import("../views/recommendations/HistoryPerf.vue"), meta: { title: "历史表现", section: "recs", sub: "history" } },
  { path: "/recommendations/strategies", name: "strategies", component: () => import("../views/recommendations/Strategies.vue"), meta: { title: "交易策略", section: "recs", sub: "strategies" } },
  { path: "/recommendations/backtest", name: "backtest", component: () => import("../views/recommendations/BacktestView.vue"), meta: { title: "回测结果", section: "recs", sub: "backtest" } },

  // Futures arbitrage
  { path: "/futures/arbitrage", name: "futures-arbitrage", component: () => import("../views/futures/FuturesArbitrage.vue"), meta: { title: "期货套利", section: "recs", icon: "trade" } },
  { path: "/futures/arbitrage/:id", name: "futures-arb-detail", component: () => import("../views/futures/FuturesArbitrageDetail.vue"), meta: { title: "套利详情", section: "recs" } },
  { path: "/futures/strategy/:id", name: "futures-strategy-detail", component: () => import("../views/futures/FuturesStrategyDetail.vue"), meta: { title: "策略详情", section: "recs" } },

  // Module 2: Analysis
  { path: "/identify", name: "identify", component: () => import("../views/analysis/StockAnalysis.vue"), meta: { title: "股票分析", section: "identify", icon: "search", sub: "analysis" } },
  { path: "/identify/logic", name: "logic-viewer", component: () => import("../views/analysis/LogicViewer.vue"), meta: { title: "推理逻辑", section: "identify", sub: "logic" } },
  { path: "/identify/:symbol", name: "identify-symbol", component: () => import("../views/analysis/StockAnalysis.vue"), meta: { title: "股票分析", section: "identify" } },

  // Module 3: News
  { path: "/news", name: "news", component: () => import("../views/news/NewsList.vue"), meta: { title: "资讯中心", section: "news", icon: "news" } },
  { path: "/news/:id", name: "news-detail", component: () => import("../views/news/NewsDetail.vue"), meta: { title: "资讯详情", section: "news" } },

  // Module 4: Community
  { path: "/community", name: "community", component: () => import("../views/community/CommunityHome.vue"), meta: { title: "社区动态", section: "community", icon: "community", sub: "feed" } },
  { path: "/community/post/:id", name: "post-detail", component: () => import("../views/community/PostDetail.vue"), meta: { title: "帖子详情", section: "community" } },
  { path: "/community/room/:symbol", name: "stock-room", component: () => import("../views/community/StockRoom.vue"), meta: { title: "股票讨论圈", section: "community" } },

  // Module 5: User
  { path: "/user", name: "profile", component: () => import("../views/user/Profile.vue"), meta: { title: "个人中心", section: "user", icon: "user", sub: "profile" } },
  { path: "/user/vip", name: "vip-center", component: () => import("../views/user/VipCenter.vue"), meta: { title: "VIP会员", section: "user", sub: "vip" } },
  { path: "/user/vip/purchase", name: "vip-purchase", component: () => import("../views/user/VipPurchase.vue"), meta: { title: "开通会员", section: "user" } },
  { path: "/user/history", name: "usage-history", component: () => import("../views/user/UsageHistory.vue"), meta: { title: "使用记录", section: "user", sub: "history" } },
  { path: "/user/messages", name: "user-messages", component: () => import("../views/user/Messages.vue"), meta: { title: "消息通知", section: "user", sub: "messages" } },
  { path: "/user/settings", name: "settings", component: () => import("../views/user/Settings.vue"), meta: { title: "设置", section: "user", sub: "settings" } },

  // Auth (blank layout)
  { path: "/login", name: "login", component: () => import("../views/user/Login.vue"), meta: { title: "登录", layout: "blank" } },
  { path: "/register", name: "register", component: () => import("../views/user/Register.vue"), meta: { title: "注册", layout: "blank" } },

  // Search
  { path: "/search", name: "search", component: () => import("../views/SearchPage.vue"), meta: { title: "搜索", section: "home" } },

  // Legacy redirect
  { path: "/analysis", redirect: "/identify" },
  { path: "/stocks", redirect: "/recommendations" },
  { path: "/:pathMatch(.*)*", redirect: "/" }
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior() {
    return { top: 0 };
  }
});

export default router;
