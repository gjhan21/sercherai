<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { clearSession, getSession, hasPermission } from "../lib/session";

const route = useRoute();
const router = useRouter();
const session = computed(() => getSession());

const menus = [
  { name: "dashboard", label: "仪表盘", to: "/dashboard", permission: "dashboard.view" },
  { name: "users", label: "用户管理", to: "/users", permission: "users.view" },
  { name: "membership-center", label: "会员中心", to: "/membership-center", permission: "membership.view" },
  { name: "news", label: "新闻管理", to: "/news", permission: "news.view" },
  { name: "market-center", label: "策略中心", to: "/market-center", permission: "market.view" },
  { name: "review-center", label: "审核中心", to: "/review-center", permission: "review.view" },
  { name: "risk-center", label: "风控中心", to: "/risk-center", permission: "risk.view" },
  { name: "audit-logs", label: "操作日志", to: "/audit-logs", permission: "audit.view" },
  { name: "auth-security", label: "安全中心", to: "/auth-security", permission: "auth_security.view" },
  { name: "system-configs", label: "系统配置", to: "/system-configs", permission: "system_config.view" },
  { name: "system-jobs", label: "任务中心", to: "/system-jobs", permission: "system_job.view" },
  { name: "data-sources", label: "数据源管理", to: "/data-sources", permission: "data_source.view" },
  { name: "workflow-messages", label: "流程消息", to: "/workflow-messages", permission: "workflow.view" },
  { name: "admin-access", label: "管理员与权限", to: "/admin-access", permission: "access.view" }
];

const activeMenu = computed(() => route.path);
const visibleMenus = computed(() => menus.filter((menu) => hasPermission(menu.permission)));

function logout() {
  clearSession();
  router.replace("/login");
}
</script>

<template>
  <el-container class="layout-shell">
    <el-aside class="layout-aside" width="230px">
      <div class="brand">SercherAI Admin</div>
      <el-scrollbar class="menu-scroll">
        <el-menu :default-active="activeMenu" router class="side-menu">
          <el-menu-item v-for="menu in visibleMenus" :key="menu.name" :index="menu.to">
            {{ menu.label }}
          </el-menu-item>
        </el-menu>
      </el-scrollbar>
    </el-aside>

    <el-container>
      <el-header class="layout-header">
        <div class="header-title">管理后台</div>
        <div class="header-actions">
          <el-text type="info">当前用户：{{ session?.userID || "-" }}</el-text>
          <el-tag type="success">{{ session?.role || "ADMIN" }}</el-tag>
          <el-button plain @click="logout">退出登录</el-button>
        </div>
      </el-header>

      <el-main class="layout-main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.layout-shell {
  min-height: 100vh;
  background: linear-gradient(180deg, #f8fafc 0%, #eef2ff 100%);
}

.layout-aside {
  border-right: 1px solid #e4e7ed;
  background: #ffffff;
  box-shadow: 0 0 24px rgba(15, 23, 42, 0.06);
}

.brand {
  padding: 16px 16px 10px;
  font-size: 19px;
  font-weight: 700;
  letter-spacing: 0.2px;
  color: #1f2937;
}

.menu-scroll {
  height: calc(100vh - 54px);
}

.side-menu {
  border-right: none;
}

.layout-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e5e7eb;
  background: rgba(255, 255, 255, 0.85);
  backdrop-filter: blur(8px);
  padding: 0 18px;
}

.header-title {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.layout-main {
  padding: 0;
}

@media (max-width: 960px) {
  .layout-aside {
    display: none;
  }
}
</style>
