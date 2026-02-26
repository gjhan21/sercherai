import { createRouter, createWebHistory } from "vue-router";
import { hasSession } from "../lib/session";

const AppLayout = () => import("../components/AppLayout.vue");
const LoginView = () => import("../views/LoginView.vue");
const DashboardView = () => import("../views/DashboardView.vue");
const UsersView = () => import("../views/UsersView.vue");
const NewsView = () => import("../views/NewsView.vue");
const ReviewCenterView = () => import("../views/ReviewCenterView.vue");
const SystemJobsView = () => import("../views/SystemJobsView.vue");
const DataSourcesView = () => import("../views/DataSourcesView.vue");
const WorkflowMessagesView = () => import("../views/WorkflowMessagesView.vue");
const AuditLogsView = () => import("../views/AuditLogsView.vue");
const MembershipCenterView = () => import("../views/MembershipCenterView.vue");
const MarketCenterView = () => import("../views/MarketCenterView.vue");
const RiskCenterView = () => import("../views/RiskCenterView.vue");
const AuthSecurityView = () => import("../views/AuthSecurityView.vue");
const SystemConfigsView = () => import("../views/SystemConfigsView.vue");

const routes = [
  {
    path: "/login",
    name: "login",
    component: LoginView,
    meta: { public: true }
  },
  {
    path: "/",
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: "",
        redirect: "/dashboard"
      },
      {
        path: "dashboard",
        name: "dashboard",
        component: DashboardView
      },
      {
        path: "data-sources",
        name: "data-sources",
        component: DataSourcesView
      },
      {
        path: "users",
        name: "users",
        component: UsersView
      },
      {
        path: "news",
        name: "news",
        component: NewsView
      },
      {
        path: "review-center",
        name: "review-center",
        component: ReviewCenterView
      },
      {
        path: "market-center",
        name: "market-center",
        component: MarketCenterView
      },
      {
        path: "membership-center",
        name: "membership-center",
        component: MembershipCenterView
      },
      {
        path: "risk-center",
        name: "risk-center",
        component: RiskCenterView
      },
      {
        path: "system-jobs",
        name: "system-jobs",
        component: SystemJobsView
      },
      {
        path: "system-configs",
        name: "system-configs",
        component: SystemConfigsView
      },
      {
        path: "auth-security",
        name: "auth-security",
        component: AuthSecurityView
      },
      {
        path: "workflow-messages",
        name: "workflow-messages",
        component: WorkflowMessagesView
      },
      {
        path: "audit-logs",
        name: "audit-logs",
        component: AuditLogsView
      }
    ]
  },
  {
    path: "/:pathMatch(.*)*",
    redirect: "/dashboard"
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach((to) => {
  const authed = hasSession();
  if (to.meta.public && authed) {
    return "/dashboard";
  }
  if (to.meta.requiresAuth && !authed) {
    return "/login";
  }
  return true;
});

export default router;
