<template>
  <div class="pc-shell">
    <header class="pc-header">
      <div class="pc-container pc-header-inner">
        <RouterLink class="pc-brand" to="/home">
          <span class="pc-brand-mark">S</span>
          <div>
            <strong>SercherAI</strong>
            <span>PC 决策工作台</span>
          </div>
        </RouterLink>

        <nav class="pc-nav">
          <RouterLink
            v-for="item in tabs"
            :key="item.path"
            :to="item.path"
            class="pc-nav-link"
            :class="{ active: route.path === item.path }"
          >
            {{ item.label }}
          </RouterLink>
        </nav>

        <div class="pc-actions">
          <span class="pc-header-chip primary">{{ currentTabLabel }}</span>
          <template v-if="isLoggedIn">
            <span class="pc-user">{{ accountLabel }}</span>
            <button type="button" class="pc-ghost-btn" :disabled="loggingOut" @click="handleLogout">
              {{ loggingOut ? "退出中..." : "退出" }}
            </button>
          </template>
          <RouterLink v-else class="pc-primary-btn" :to="{ path: '/auth', query: { redirect: route.fullPath } }">
            登录 / 注册
          </RouterLink>
        </div>
      </div>
    </header>

    <main class="pc-main">
      <div class="pc-container">
        <RouterView />
      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { logout } from "../../../shared/api/auth";
import { clearClientAuthSession, useClientAuth } from "../../../shared/auth/client-auth";

const route = useRoute();
const router = useRouter();
const loggingOut = ref(false);
const { session, isLoggedIn } = useClientAuth();

const tabs = [
  { path: "/home", label: "首页" },
  { path: "/strategies", label: "策略" },
  { path: "/archive", label: "档案" },
  { path: "/watchlist", label: "关注" },
  { path: "/news", label: "资讯" },
  { path: "/membership", label: "会员" },
  { path: "/profile", label: "我的" }
];

const accountLabel = computed(() => {
  const phone = session.value?.phone || "";
  if (/^\d{11}$/.test(phone)) {
    return `${phone.slice(0, 3)}****${phone.slice(-4)}`;
  }
  const email = String(session.value?.email || "").trim();
  if (email.includes("@")) {
    const [local, domain] = email.split("@");
    if ((local || "").length <= 2) {
      return `${(local || "").slice(0, 1)}***@${domain || ""}`;
    }
    return `${local.slice(0, 2)}***@${domain || ""}`;
  }
  return session.value?.userID || "当前用户";
});

const currentTabLabel = computed(() => {
  return tabs.find((item) => route.path === item.path)?.label || "桌面终端";
});

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
    console.warn("pc logout failed:", error?.message || error);
  } finally {
    clearClientAuthSession();
    loggingOut.value = false;
    await router.replace("/auth");
  }
}
</script>
