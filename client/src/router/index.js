import { createRouter, createWebHistory } from "vue-router";
import { syncClientAuthSession } from "../lib/client-auth";

const ClientLayout = () => import("../components/ClientLayout.vue");
const AuthView = () => import("../views/AuthView.vue");
const HomeView = () => import("../views/HomeView.vue");
const StrategyView = () => import("../views/StrategyView.vue");
const NewsView = () => import("../views/NewsView.vue");
const MembershipView = () => import("../views/MembershipView.vue");
const ProfileView = () => import("../views/ProfileView.vue");

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
    name: "auth",
    component: AuthView,
    meta: { guestOnly: true }
  },
  {
    path: "/",
    component: ClientLayout,
    children: [
      { path: "", redirect: "/home" },
      { path: "home", name: "home", component: HomeView },
      { path: "strategies", name: "strategies", component: StrategyView },
      { path: "news", name: "news", component: NewsView },
      { path: "membership", name: "membership", component: MembershipView, meta: { requiresAuth: true } },
      { path: "profile", name: "profile", component: ProfileView, meta: { requiresAuth: true } }
    ]
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/home"
  }
];

const router = createRouter({
  history: createWebHistory(),
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
