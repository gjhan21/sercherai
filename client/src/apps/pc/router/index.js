import { createRouter, createWebHistory } from "vue-router";
import { syncClientAuthSession } from "../../../shared/auth/client-auth";
import { buildProfileModuleRoute } from "../../../lib/profile-modules";

const PcLayout = () => import("../layouts/PcLayout.vue");
const PcAuthView = () => import("../views/PcAuthView.vue");
const PcHomeView = () => import("../views/PcHomeView.vue");
const PcStrategyView = () => import("../views/PcStrategyView.vue");
const PcArchiveView = () => import("../views/PcArchiveView.vue");
const PcWatchlistView = () => import("../views/PcWatchlistView.vue");
const PcNewsView = () => import("../views/PcNewsView.vue");
const PcSearchView = () => import("../views/PcSearchView.vue");
const PcCommunityView = () => import("../views/PcCommunityView.vue");
const PcCommunityTopicView = () => import("../views/PcCommunityTopicView.vue");
const PcCommunityComposeView = () => import("../views/PcCommunityComposeView.vue");
const PcMembershipView = () => import("../views/PcMembershipView.vue");
const PcProfileView = () => import("../views/PcProfileView.vue");
const PcForecastRunView = () => import("../views/PcForecastRunView.vue");


const routes = [
  {
    path: "/invite/:inviteCode",
    redirect: (to) => ({
      path: "/auth",
      query: {
        invite_code: String(to.params.inviteCode || ""),
        redirect: "/home"
      }
    })
  },
  {
    path: "/auth",
    name: "pc-auth",
    component: PcAuthView,
    meta: { guestOnly: true, clientPlatform: "pc" }
  },
  {
    path: "/",
    component: PcLayout,
    children: [
      { path: "", redirect: "/home" },
      { path: "home", name: "pc-home", component: PcHomeView },
      { path: "strategies", name: "pc-strategies", component: PcStrategyView },
      { path: "archive", name: "pc-archive", component: PcArchiveView },
      { path: "news", name: "pc-news", component: PcNewsView },
      { path: "search", name: "pc-search", component: PcSearchView },
      { path: "community", name: "pc-community", component: PcCommunityView },
      { path: "community/topics/:id", name: "pc-community-topic", component: PcCommunityTopicView },
      { path: "community/new", name: "pc-community-compose", component: PcCommunityComposeView, meta: { requiresAuth: true } },
      { path: "watchlist", redirect: buildProfileModuleRoute("watchlist") },
      { path: "membership", name: "pc-membership", component: PcMembershipView, meta: { requiresAuth: true } },
      { path: "profile", name: "pc-profile", component: PcProfileView, meta: { requiresAuth: true } },
      { path: "profile/watchlist", name: "pc-profile-watchlist", component: PcWatchlistView, meta: { requiresAuth: true } },
      { path: "forecast/runs/:id", name: "forecast-run", component: PcForecastRunView, meta: { requiresAuth: true } }

    ]
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/home"
  }
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
});

router.beforeEach((to) => {
  const session = syncClientAuthSession();
  const loggedIn = Boolean(session?.accessToken);
  if (to.meta?.requiresAuth && !loggedIn) {
    return {
      path: "/auth",
      query: { redirect: to.fullPath }
    };
  }
  if (to.meta?.guestOnly && loggedIn) {
    const redirect = String(to.query.redirect || "");
    if (redirect.startsWith("/")) {
      return redirect;
    }
    return "/home";
  }
  return true;
});

export default router;
