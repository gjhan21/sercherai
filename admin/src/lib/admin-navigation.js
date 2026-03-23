import { hasPermission } from "./session.js";

export const NO_ACCESS_ROUTE_PATH = "/no-access";

export const adminNavigationItems = [
  { name: "dashboard", label: "仪表盘", to: "/dashboard", permission: "dashboard.view" },
  { name: "users", label: "用户管理", to: "/users", permission: "users.view" },
  { name: "browse-insights", label: "阅读分析", to: "/browse-insights", permission: "users.view" },
  { name: "user-messages", label: "用户消息", to: "/user-messages", permission: "users.view" },
  { name: "membership-center", label: "会员中心", to: "/membership-center", permission: "membership.view" },
  { name: "news", label: "新闻管理", to: "/news", permission: "news.view" },
  { name: "market-center", label: "策略中心", to: "/market-center", permission: "market.view" },
  { name: "stock-selection", label: "智能选股", to: "/stock-selection/overview", permission: "stock_selection.view" },
  { name: "futures-selection", label: "智能期货", to: "/futures-selection/overview", permission: "futures_selection.view" },
  { name: "data-sources", label: "数据源管理", to: "/data-sources", permission: "data_source.view" },
  { name: "experiment-analytics", label: "实验埋点看板", to: "/experiment-analytics", permission: "market.view" },
  { name: "review-center", label: "审核中心", to: "/review-center", permission: "review.view" },
  { name: "system-jobs", label: "任务中心", to: "/system-jobs", permission: "system_job.view" },
  { name: "risk-center", label: "风控中心", to: "/risk-center", permission: "risk.view" },
  { name: "audit-logs", label: "操作日志", to: "/audit-logs", permission: "audit.view" },
  { name: "auth-security", label: "安全中心", to: "/auth-security", permission: "auth_security.view" },
  { name: "system-configs", label: "配置中心", to: "/system-configs", permission: "system_config.view" },
  { name: "workflow-messages", label: "流程消息", to: "/workflow-messages", permission: "workflow.view" },
  { name: "admin-access", label: "管理员与权限", to: "/admin-access", permission: "access.view" }
];

export function getVisibleAdminNavigationItems() {
  return adminNavigationItems.filter((item) => hasPermission(item.permission));
}

export function resolveFirstAccessibleRoute() {
  return getVisibleAdminNavigationItems()[0]?.to || NO_ACCESS_ROUTE_PATH;
}
