import { createRouter, createWebHistory } from "vue-router";
import AppLayout from "../components/AppLayout.vue";
import DashboardView from "../views/DashboardView.vue";
import DataSourcesView from "../views/DataSourcesView.vue";
import LoginView from "../views/LoginView.vue";
import WorkflowMessagesView from "../views/WorkflowMessagesView.vue";
import { hasSession } from "../lib/session";

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
        path: "workflow-messages",
        name: "workflow-messages",
        component: WorkflowMessagesView
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
