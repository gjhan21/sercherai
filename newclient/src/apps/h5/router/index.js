import { createRouter, createWebHistory } from "vue-router";

const routes = [
  { path: "/", name: "home", component: () => import("../views/Home.vue"), meta: { tab: 0 } },
  { path: "/markets", name: "markets", component: () => import("../views/Markets.vue"), meta: { tab: 1 } },
  { path: "/markets/:symbol", name: "stock-detail", component: () => import("../views/StockDetail.vue"), meta: { tab: 1 } },
  { path: "/news", name: "news", component: () => import("../views/news/NewsList.vue"), meta: { tab: 2 } },
  { path: "/news/:id", name: "news-detail", component: () => import("../views/news/NewsDetail.vue"), meta: { tab: 2 } },
  { path: "/community", name: "community", component: () => import("../views/community/CommunityHome.vue"), meta: { tab: 3 } },
  { path: "/community/post/:id", name: "h5-post-detail", component: () => import("../views/community/PostDetail.vue"), meta: { tab: 3 } },
  { path: "/profile", name: "h5-profile", component: () => import("../views/Profile.vue"), meta: { tab: 4 } },
  { path: "/profile/vip", name: "h5-vip", component: () => import("../views/user/VipCenter.vue"), meta: {} },
  { path: "/profile/settings", name: "h5-settings", component: () => import("../views/user/Settings.vue"), meta: {} },
  { path: "/futures/arbitrage", name: "h5-futures-arb", component: () => import("../views/futures/FuturesArbitrage.vue"), meta: {} },
  { path: "/futures/arbitrage/:id", name: "h5-futures-arb-detail", component: () => import("../views/futures/FuturesArbitrageDetail.vue"), meta: {} },
  { path: "/forecast-lab", name: "h5-forecast-lab", component: () => import("../views/forecast/ForecastLabView.vue"), meta: {} },
  { path: "/forecast/:id", name: "h5-forecast-detail", component: () => import("../views/forecast/ForecastDetailView.vue"), meta: {} },
  { path: "/login", name: "h5-login", component: () => import("../views/user/Login.vue"), meta: { layout: "blank" } },
  { path: "/register", name: "h5-register", component: () => import("../views/user/Register.vue"), meta: { layout: "blank" } },
  { path: "/profile/forgot-password", name: "h5-forgot-password", component: () => import("../views/user/ForgotPassword.vue"), meta: { layout: "blank" } },
  { path: "/:pathMatch(.*)*", redirect: "/" }
];

const router = createRouter({
  history: createWebHistory("/m/"),
  routes,
  scrollBehavior() {
    return { top: 0 };
  }
});

export default router;
