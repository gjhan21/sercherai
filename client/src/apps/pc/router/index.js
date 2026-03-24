import { createRouter, createWebHistory } from "vue-router";
import { syncClientAuthSession } from "../../../shared/auth/client-auth";

const PcLayout = () => import("../layouts/PcLayout.vue");
const AuthView = () => import("../../../views/AuthView.vue");
const HomeView = () => import("../../../views/HomeView.vue");
const StrategyView = () => import("../../../views/StrategyView.vue");
const RecommendationArchiveView = () => import("../../../views/RecommendationArchiveView.vue");
const MyWatchlistView = () => import("../../../views/MyWatchlistView.vue");
const NewsView = () => import("../../../views/NewsView.vue");
const MembershipView = () => import("../../../views/MembershipView.vue");
const ProfileView = () => import("../../../views/ProfileView.vue");

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
    component: AuthView,
    meta: { guestOnly: true, clientPlatform: "pc" }
  },
  {
    path: "/",
    component: PcLayout,
    children: [
      { path: "", redirect: "/home" },
      { path: "home", name: "pc-home", component: HomeView },
      { path: "strategies", name: "pc-strategies", component: StrategyView },
      { path: "archive", name: "pc-archive", component: RecommendationArchiveView },
      { path: "watchlist", name: "pc-watchlist", component: MyWatchlistView },
      { path: "news", name: "pc-news", component: NewsView },
      { path: "membership", name: "pc-membership", component: MembershipView, meta: { requiresAuth: true } },
      { path: "profile", name: "pc-profile", component: ProfileView, meta: { requiresAuth: true } }
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
