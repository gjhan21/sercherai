import { createRouter, createWebHistory } from "vue-router";
import { syncClientAuthSession } from "../../../shared/auth/client-auth";

const H5Layout = () => import("../layouts/H5Layout.vue");
const H5AuthView = () => import("../views/H5AuthView.vue");
const H5HomeView = () => import("../views/H5HomeView.vue");
const H5StrategyView = () => import("../views/H5StrategyView.vue");
const H5NewsView = () => import("../views/H5NewsView.vue");
const H5MembershipView = () => import("../views/H5MembershipView.vue");
const H5ProfileView = () => import("../views/H5ProfileView.vue");

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
    name: "h5-auth",
    component: H5AuthView,
    meta: { guestOnly: true, clientPlatform: "h5" }
  },
  {
    path: "/",
    component: H5Layout,
    children: [
      { path: "", redirect: "/home" },
      { path: "home", name: "h5-home", component: H5HomeView },
      { path: "strategies", name: "h5-strategies", component: H5StrategyView },
      { path: "news", name: "h5-news", component: H5NewsView },
      { path: "membership", name: "h5-membership", component: H5MembershipView, meta: { requiresAuth: true } },
      { path: "profile", name: "h5-profile", component: H5ProfileView, meta: { requiresAuth: true } },
      { path: "archive", redirect: "/news" },
      { path: "watchlist", redirect: "/strategies" }
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
