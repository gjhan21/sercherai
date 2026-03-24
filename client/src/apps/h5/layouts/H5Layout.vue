<template>
  <div class="h5-shell">
    <header class="h5-header">
      <div class="h5-header-inner">
        <div class="h5-header-main">
          <RouterLink class="h5-brand" to="/home" aria-label="返回首页">
            <span class="h5-brand-mark">S</span>
            <div class="h5-brand-copy">
              <span class="h5-brand-title">{{ shellScene.section }}</span>
              <strong>{{ shellScene.title }}</strong>
              <small>{{ shellScene.subtitle }}</small>
            </div>
          </RouterLink>
        </div>

        <div class="h5-header-actions">
          <span class="h5-shell-pulse">{{ shellScene.pulse }}</span>
          <span v-if="isLoggedIn" class="h5-user">{{ accountLabel }}</span>
          <RouterLink v-if="!isLoggedIn" class="h5-header-link" :to="{ path: '/auth', query: { redirect: route.fullPath } }">
            登录
          </RouterLink>
          <button v-else type="button" class="h5-header-link" :disabled="loggingOut" @click="handleLogout">
            {{ loggingOut ? '退出中' : '退出' }}
          </button>
        </div>
      </div>
    </header>

    <main class="h5-main">
      <RouterView />
    </main>

    <nav class="h5-tabbar" aria-label="H5 底部导航">
      <RouterLink
        v-for="item in tabs"
        :key="item.path"
        :to="item.path"
        class="h5-tabbar-link"
        :class="{ active: isTabActive(item.path) }"
      >
        <span class="h5-tabbar-icon" aria-hidden="true" v-html="tabIcon(item.icon)"></span>
        <span>{{ item.label }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { logout } from "../../../shared/api/auth";
import { clearClientAuthSession, useClientAuth } from "../../../shared/auth/client-auth";
import { resolveShellScene, resolveTabItems } from "../lib/shell-meta";

const TAB_ICON_MAP = {
  home: `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M4.5 10.5 12 4l7.5 6.5" />
      <path d="M6.5 10.5V20h11v-9.5" />
    </svg>
  `,
  news: `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M5 6.5h14" />
      <path d="M5 11.5h14" />
      <path d="M5 16.5h9" />
      <path d="M17 15.5h2v3h-2z" />
    </svg>
  `,
  insight: `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="M5 18h14" />
      <path d="M7 16V9" />
      <path d="M12 16V6" />
      <path d="M17 16v-4" />
    </svg>
  `,
  vip: `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <path d="m5 8 3 3 4-5 4 5 3-3" />
      <path d="M6 10.5 8 18h8l2-7.5" />
    </svg>
  `,
  user: `
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
      <circle cx="12" cy="8" r="3.2" />
      <path d="M5 19c1.8-3 4.2-4.5 7-4.5S17.2 16 19 19" />
    </svg>
  `
};

const route = useRoute();
const router = useRouter();
const loggingOut = ref(false);
const { session, isLoggedIn } = useClientAuth();

const tabs = resolveTabItems();
const shellScene = computed(() => resolveShellScene(route.fullPath));

const accountLabel = computed(() => {
  const phone = session.value?.phone || "";
  if (/^\d{11}$/.test(phone)) {
    return `${phone.slice(0, 3)}****${phone.slice(-2)}`;
  }
  const email = String(session.value?.email || "").trim();
  if (email.includes("@")) {
    return email.split("@")[0] || "当前用户";
  }
  return session.value?.userID || "当前用户";
});

function isTabActive(path) {
  if (path === "/home") {
    return route.path === "/home" || route.path === "/";
  }
  return route.path.startsWith(path);
}

function tabIcon(name) {
  return TAB_ICON_MAP[name] || TAB_ICON_MAP.home;
}

async function handleLogout() {
  if (loggingOut.value) {
    return;
  }
  loggingOut.value = true;
  try {
    const refreshToken = session.value?.refreshToken || "";
    if (refreshToken) {
      await logout(refreshToken);
    }
  } catch (error) {
    console.warn("h5 logout failed:", error?.message || error);
  } finally {
    clearClientAuthSession();
    loggingOut.value = false;
    await router.replace("/auth");
  }
}
</script>
