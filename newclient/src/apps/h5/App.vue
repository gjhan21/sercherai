<template>
  <div v-if="isBlank" class="h5-blank">
    <RouterView />
  </div>
  <div v-else class="h5-root">
    <header class="h5-header">
      <div class="h5-header-inner">
        <div class="h5-brand">
          <div class="h5-brand-icon">
            <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="16" cy="16" r="14" />
              <path d="M10 20l4-6 4 3 4-7" stroke-width="2.5" />
            </svg>
          </div>
          <span class="h5-brand-name">智投AI</span>
        </div>
        <div class="h5-market-status">
          <span class="h5-clock-dot" :class="marketClass"></span>
          <span>{{ marketStatusText }}</span>
        </div>
      </div>
    </header>

    <main class="h5-content">
      <RouterView />
    </main>

    <nav class="h5-tabbar">
      <RouterLink v-for="(tab, i) in tabs" :key="tab.path" :to="tab.path" class="h5-tab" :class="{ active: activeTab === i }">
        <span class="h5-tab-icon" v-html="tab.icon"></span>
        <span class="h5-tab-label">{{ tab.label }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

const isBlank = computed(() => route.meta?.layout === "blank");

const now = new Date();
const hour = now.getHours();
const marketClass = computed(() => (hour >= 9 && hour < 15) ? "open" : "closed");
const marketStatusText = computed(() => (hour >= 9 && hour < 15) ? "交易中" : "已收盘");

const activeTab = computed(() => {
  const tab = route.meta?.tab;
  return tab !== undefined ? tab : 0;
});

const tabs = [
  {
    path: "/",
    label: "首页",
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 12l8-8 8 8"/><path d="M6 10v8h4v-4h4v4h4v-8"/></svg>`
  },
  {
    path: "/markets",
    label: "行情",
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M3 3v18h18"/><path d="M7 16l4-6 4 3 5-7"/></svg>`
  },
  {
    path: "/news",
    label: "资讯",
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M5 5h14v16H5z"/><path d="M8 9h8M8 13h6M8 17h4"/></svg>`
  },
  {
    path: "/community",
    label: "社区",
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87"/><path d="M16 3.13a4 4 0 010 7.75"/></svg>`
  },
  {
    path: "/profile",
    label: "我的",
    icon: `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="8" r="4"/><path d="M4 20c1.8-4 4.2-6 8-6s6.2 2 8 6"/></svg>`
  }
];
</script>

<style scoped>
.h5-blank {
  min-height: 100vh;
  background: var(--bg-primary);
}

.h5-root {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--bg-primary);
}

.h5-header {
  flex-shrink: 0;
  padding: 12px 16px;
  padding-top: max(12px, env(safe-area-inset-top));
  border-bottom: 1px solid var(--border);
  background: var(--bg-secondary);
}

.h5-header-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.h5-brand {
  display: flex;
  align-items: center;
  gap: 8px;
}

.h5-brand-icon {
  width: 28px;
  height: 28px;
  color: var(--accent-gold);
  filter: drop-shadow(0 0 6px rgba(240, 185, 11, 0.3));
}

.h5-brand-name {
  font-size: 16px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--accent-gold), #fff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.h5-market-status {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
}

.h5-clock-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.h5-clock-dot.open { background: var(--positive); box-shadow: 0 0 8px rgba(0, 200, 151, 0.5); }
.h5-clock-dot.closed { background: var(--neutral); }

.h5-content {
  flex: 1;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
  padding: 12px 16px;
  padding-bottom: 8px;
}

.h5-tabbar {
  flex-shrink: 0;
  display: flex;
  background: var(--bg-secondary);
  border-top: 1px solid var(--border);
  padding: 6px 0;
  padding-bottom: max(6px, env(safe-area-inset-bottom));
}

.h5-tab {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  padding: 4px;
  color: var(--text-secondary);
  transition: color 0.2s;
  -webkit-tap-highlight-color: transparent;
}

.h5-tab.active { color: var(--accent-gold); }

.h5-tab-icon {
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.h5-tab-icon svg { width: 22px; height: 22px; }

.h5-tab-label { font-size: 10px; font-weight: 500; }
</style>
