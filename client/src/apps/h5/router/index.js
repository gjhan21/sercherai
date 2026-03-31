import { createRouter, createWebHistory } from "vue-router";
import { syncClientAuthSession } from "../../../shared/auth/client-auth";
import { buildProfileModuleRoute } from "../../../lib/profile-modules";

const H5Layout = () => import("../layouts/H5Layout.vue");
const H5AuthView = () => import("../views/H5AuthView.vue");
const H5HomeView = () => import("../views/H5HomeView.vue");
const H5StrategyView = () => import("../views/H5StrategyView.vue");
const H5NewsView = () => import("../views/H5NewsView.vue");
const H5WatchlistView = () => import("../views/H5WatchlistView.vue");
const H5ArchiveView = () => import("../views/H5ArchiveView.vue");
const H5MembershipView = () => import("../views/H5MembershipView.vue");
const H5ProfileView = () => import("../views/H5ProfileView.vue");
const H5SearchView = () => import("../views/H5SearchView.vue");
const H5CommunityView = () => import("../views/H5CommunityView.vue");
const H5CommunityTopicView = () => import("../views/H5CommunityTopicView.vue");
const H5CommunityComposeView = () => import("../views/H5CommunityComposeView.vue");
const ForecastRunView = () => import("../views/H5ForecastRunView.vue");


const routes = [
  {
    path: "/m/invite/:inviteCode",
    redirect: (to) => ({
      path: "/m/auth",
      query: {
        invite_code: String(to.params.inviteCode || ""),
        redirect: "/m/home"
      }
    })
  },
  {
    path: "/m/auth",
    name: "h5-auth",
    component: H5AuthView,
    meta: { guestOnly: true, clientPlatform: "h5" }
  },
  {
    path: "/m",
    component: H5Layout,
    children: [
      { path: "", redirect: "/m/home" },
      { path: "home", name: "h5-home", component: H5HomeView },
      { path: "strategies", name: "h5-strategies", component: H5StrategyView },
      { path: "news", name: "h5-news", component: H5NewsView },
      { path: "search", name: "h5-search", component: H5SearchView },
      { path: "watchlist", redirect: buildProfileModuleRoute("watchlist") },
      { path: "archive", name: "h5-archive", component: H5ArchiveView },
      { path: "membership", name: "h5-membership", component: H5MembershipView, meta: { requiresAuth: true } },
      { path: "profile", name: "h5-profile", component: H5ProfileView, meta: { requiresAuth: true } },
      { path: "profile/watchlist", name: "h5-profile-watchlist", component: H5WatchlistView, meta: { requiresAuth: true } },
      { path: "community", name: "h5-community", component: H5CommunityView },
      { path: "community/topics/:id", name: "h5-community-topic", component: H5CommunityTopicView },
      { path: "community/new", name: "h5-community-new", component: H5CommunityComposeView, meta: { requiresAuth: true } },
      { path: "forecast/runs/:id", name: "forecast-run", component: ForecastRunView, meta: { requiresAuth: true } }

    ]
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/home"
  }
];

const routerInstance = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
});

routerInstance.beforeEach((to) => {
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

export default routerInstance;
