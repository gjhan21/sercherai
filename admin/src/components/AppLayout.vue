<script setup>
import { computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { clearSession, getSession } from "../lib/session";

const route = useRoute();
const router = useRouter();
const session = computed(() => getSession());

const menus = [
  { name: "dashboard", label: "仪表盘", to: "/dashboard" },
  { name: "users", label: "用户管理", to: "/users" },
  { name: "news", label: "新闻管理", to: "/news" },
  { name: "review-center", label: "审核中心", to: "/review-center" },
  { name: "data-sources", label: "数据源管理", to: "/data-sources" },
  { name: "workflow-messages", label: "流程消息", to: "/workflow-messages" }
];

function isActive(menu) {
  return route.name === menu.name;
}

function logout() {
  clearSession();
  router.replace("/login");
}
</script>

<template>
  <div class="layout">
    <aside class="sider">
      <div class="brand">SercherAI Admin</div>
      <nav class="menu">
        <router-link
          v-for="menu in menus"
          :key="menu.name"
          :to="menu.to"
          class="menu-item"
          :class="{ 'menu-item-active': isActive(menu) }"
        >
          {{ menu.label }}
        </router-link>
      </nav>
    </aside>

    <div class="main">
      <header class="topbar">
        <div class="topbar-title">管理后台</div>
        <div class="topbar-user">
          <span class="muted">当前用户：{{ session?.userID || "-" }}</span>
          <span class="status-tag status-active">{{ session?.role || "ADMIN" }}</span>
          <button class="btn" @click="logout">退出登录</button>
        </div>
      </header>
      <router-view />
    </div>
  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  display: grid;
  grid-template-columns: 220px 1fr;
}

.sider {
  background: #111827;
  color: #f3f4f6;
  border-right: 1px solid #1f2937;
  padding: 12px;
}

.brand {
  font-size: 18px;
  font-weight: 700;
  padding: 10px 10px 14px;
}

.menu {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.menu-item {
  border-radius: 8px;
  padding: 10px;
  color: #d1d5db;
  font-size: 14px;
}

.menu-item:hover {
  background: #1f2937;
  color: #f9fafb;
}

.menu-item-active {
  background: #2563eb;
  color: #fff;
}

.main {
  min-width: 0;
}

.topbar {
  height: 56px;
  border-bottom: 1px solid #e5e7eb;
  background: #fff;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.topbar-title {
  font-size: 16px;
  font-weight: 600;
}

.topbar-user {
  display: flex;
  align-items: center;
  gap: 8px;
}

@media (max-width: 900px) {
  .layout {
    grid-template-columns: 1fr;
  }

  .sider {
    border-right: 0;
    border-bottom: 1px solid #1f2937;
  }

  .menu {
    flex-direction: row;
    flex-wrap: wrap;
  }
}
</style>
