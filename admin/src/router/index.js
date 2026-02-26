import { createRouter, createWebHistory } from "vue-router";
import { hasPermission, hasSession } from "../lib/session";

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
const AdminAccessView = () => import("../views/AdminAccessView.vue");

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
        component: DashboardView,
        meta: { permission: "dashboard.view" }
      },
      {
        path: "data-sources",
        name: "data-sources",
        component: DataSourcesView,
        meta: { permission: "data_source.view" }
      },
      {
        path: "users",
        name: "users",
        component: UsersView,
        meta: { permission: "users.view" }
      },
      {
        path: "news",
        name: "news",
        component: NewsView,
        meta: { permission: "news.view" }
      },
      {
        path: "review-center",
        name: "review-center",
        component: ReviewCenterView,
        meta: { permission: "review.view" }
      },
      {
        path: "market-center",
        name: "market-center",
        component: MarketCenterView,
        meta: { permission: "market.view" }
      },
      {
        path: "membership-center",
        name: "membership-center",
        component: MembershipCenterView,
        meta: { permission: "membership.view" }
      },
      {
        path: "risk-center",
        name: "risk-center",
        component: RiskCenterView,
        meta: { permission: "risk.view" }
      },
      {
        path: "system-jobs",
        name: "system-jobs",
        component: SystemJobsView,
        meta: { permission: "system_job.view" }
      },
      {
        path: "system-configs",
        name: "system-configs",
        component: SystemConfigsView,
        meta: { permission: "system_config.view" }
      },
      {
        path: "auth-security",
        name: "auth-security",
        component: AuthSecurityView,
        meta: { permission: "auth_security.view" }
      },
      {
        path: "workflow-messages",
        name: "workflow-messages",
        component: WorkflowMessagesView,
        meta: { permission: "workflow.view" }
      },
      {
        path: "audit-logs",
        name: "audit-logs",
        component: AuditLogsView,
        meta: { permission: "audit.view" }
      },
      {
        path: "admin-access",
        name: "admin-access",
        component: AdminAccessView,
        meta: { permission: "access.view" }
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
  if (to.meta.permission && !hasPermission(to.meta.permission)) {
    return "/dashboard";
  }
  return true;
});

export default router;
