import { createRouter, createWebHistory } from "vue-router";
import { NO_ACCESS_ROUTE_PATH, resolveFirstAccessibleRoute } from "../lib/admin-navigation";
import { hasPermission, hasSession } from "../lib/session";

const AppLayout = () => import("../components/AppLayout.vue");
const LoginView = () => import("../views/LoginView.vue");
const NoAccessView = () => import("../views/NoAccessView.vue");
const PublicHomePreviewView = () => import("../views/PublicHomePreviewView.vue");
const DashboardView = () => import("../views/DashboardView.vue");
const UsersView = () => import("../views/UsersView.vue");
const BrowseInsightsView = () => import("../views/BrowseInsightsView.vue");
const UserMessagesView = () => import("../views/UserMessagesView.vue");
const NewsView = () => import("../views/NewsView.vue");
const ReviewCenterView = () => import("../views/ReviewCenterView.vue");
const SystemJobsView = () => import("../views/SystemJobsView.vue");
const DataSourcesView = () => import("../views/DataSourcesView.vue");
const WorkflowMessagesView = () => import("../views/WorkflowMessagesView.vue");
const AuditLogsView = () => import("../views/AuditLogsView.vue");
const MembershipCenterView = () => import("../views/MembershipCenterView.vue");
const MarketCenterView = () => import("../views/MarketCenterView.vue");
const ExperimentAnalyticsView = () => import("../views/ExperimentAnalyticsView.vue");
const StockSelectionOverviewView = () => import("../views/stock-selection/StockSelectionOverviewView.vue");
const StockSelectionRunsView = () => import("../views/stock-selection/StockSelectionRunsView.vue");
const StockSelectionRulesView = () => import("../views/stock-selection/StockSelectionRulesView.vue");
const StockSelectionGraphView = () => import("../views/stock-selection/StockSelectionGraphView.vue");
const StockSelectionRolesView = () => import("../views/stock-selection/StockSelectionRolesView.vue");
const StockSelectionTemplatesView = () => import("../views/stock-selection/StockSelectionTemplatesView.vue");
const StockSelectionProfilesView = () => import("../views/stock-selection/StockSelectionProfilesView.vue");
const StockSelectionFactorsView = () => import("../views/stock-selection/StockSelectionFactorsView.vue");
const StockSelectionCandidatesView = () => import("../views/stock-selection/StockSelectionCandidatesView.vue");
const StockSelectionEvaluationView = () => import("../views/stock-selection/StockSelectionEvaluationView.vue");
const FuturesSelectionOverviewView = () => import("../views/futures-selection/FuturesSelectionOverviewView.vue");
const FuturesSelectionRunsView = () => import("../views/futures-selection/FuturesSelectionRunsView.vue");
const FuturesSelectionRulesView = () => import("../views/futures-selection/FuturesSelectionRulesView.vue");
const FuturesSelectionGraphView = () => import("../views/futures-selection/FuturesSelectionGraphView.vue");
const FuturesSelectionRolesView = () => import("../views/futures-selection/FuturesSelectionRolesView.vue");
const FuturesSelectionTemplatesView = () => import("../views/futures-selection/FuturesSelectionTemplatesView.vue");
const FuturesSelectionProfilesView = () => import("../views/futures-selection/FuturesSelectionProfilesView.vue");
const FuturesSelectionFactorsView = () => import("../views/futures-selection/FuturesSelectionFactorsView.vue");
const FuturesSelectionCandidatesView = () => import("../views/futures-selection/FuturesSelectionCandidatesView.vue");
const FuturesSelectionEvaluationView = () => import("../views/futures-selection/FuturesSelectionEvaluationView.vue");
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
    path: "/home-preview",
    name: "home-preview",
    component: PublicHomePreviewView,
    meta: { public: true, allowAuthed: true }
  },
  {
    path: "/",
    component: AppLayout,
    meta: { requiresAuth: true },
    children: [
      {
        path: "",
        redirect: () => resolveFirstAccessibleRoute()
      },
      {
        path: "no-access",
        name: "no-access",
        component: NoAccessView
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
        path: "browse-insights",
        name: "browse-insights",
        component: BrowseInsightsView,
        meta: { permission: "users.view" }
      },
      {
        path: "user-messages",
        name: "user-messages",
        component: UserMessagesView,
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
        path: "experiment-analytics",
        name: "experiment-analytics",
        component: ExperimentAnalyticsView,
        meta: { permission: "market.view" }
      },
      {
        path: "stock-selection/overview",
        name: "stock-selection-overview",
        component: StockSelectionOverviewView,
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "stock-selection/runs",
        name: "stock-selection-runs",
        component: StockSelectionRunsView,
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "stock-selection/rules",
        name: "stock-selection-rules",
        component: StockSelectionRulesView,
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "stock-selection/factors",
        name: "stock-selection-factors",
        component: StockSelectionFactorsView,
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "stock-selection/graph",
        name: "stock-selection-graph",
        component: StockSelectionGraphView,
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "stock-selection/roles",
        name: "stock-selection-roles",
        component: StockSelectionRolesView,
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "stock-selection/templates",
        name: "stock-selection-templates",
        component: StockSelectionTemplatesView,
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "stock-selection/profiles",
        name: "stock-selection-profiles",
        component: StockSelectionProfilesView,
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "stock-selection/candidates",
        name: "stock-selection-candidates",
        component: StockSelectionCandidatesView,
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "stock-selection/reviews",
        name: "stock-selection-reviews",
        redirect: (to) => ({
          name: "stock-selection-candidates",
          query: to.query
        }),
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "stock-selection/evaluation",
        name: "stock-selection-evaluation",
        component: StockSelectionEvaluationView,
        meta: { permission: "stock_selection.view", navKey: "/stock-selection/overview" }
      },
      {
        path: "futures-selection/overview",
        name: "futures-selection-overview",
        component: FuturesSelectionOverviewView,
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
      },
      {
        path: "futures-selection/runs",
        name: "futures-selection-runs",
        component: FuturesSelectionRunsView,
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
      },
      {
        path: "futures-selection/rules",
        name: "futures-selection-rules",
        component: FuturesSelectionRulesView,
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
      },
      {
        path: "futures-selection/factors",
        name: "futures-selection-factors",
        component: FuturesSelectionFactorsView,
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
      },
      {
        path: "futures-selection/graph",
        name: "futures-selection-graph",
        component: FuturesSelectionGraphView,
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
      },
      {
        path: "futures-selection/roles",
        name: "futures-selection-roles",
        component: FuturesSelectionRolesView,
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
      },
      {
        path: "futures-selection/templates",
        name: "futures-selection-templates",
        component: FuturesSelectionTemplatesView,
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
      },
      {
        path: "futures-selection/profiles",
        name: "futures-selection-profiles",
        component: FuturesSelectionProfilesView,
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
      },
      {
        path: "futures-selection/candidates",
        name: "futures-selection-candidates",
        component: FuturesSelectionCandidatesView,
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
      },
      {
        path: "futures-selection/reviews",
        name: "futures-selection-reviews",
        redirect: (to) => ({
          name: "futures-selection-candidates",
          query: to.query
        }),
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
      },
      {
        path: "futures-selection/evaluation",
        name: "futures-selection-evaluation",
        component: FuturesSelectionEvaluationView,
        meta: { permission: "futures_selection.view", navKey: "/futures-selection/overview" }
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
    redirect: () => (hasSession() ? resolveFirstAccessibleRoute() : "/login")
  }
];

const router = createRouter({
  history: createWebHistory(),
  routes
});

router.beforeEach((to) => {
  const authed = hasSession();
  const landingPath = resolveFirstAccessibleRoute();
  if (to.meta.public && authed && !to.meta.allowAuthed) {
    return landingPath;
  }
  if (to.meta.requiresAuth && !authed) {
    return "/login";
  }
  if (to.meta.permission && !hasPermission(to.meta.permission)) {
    return to.path === landingPath ? NO_ACCESS_ROUTE_PATH : landingPath;
  }
  return true;
});

export default router;
