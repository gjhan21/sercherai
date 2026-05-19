<template>
  <aside class="pc-sidebar">
    <div class="sidebar-brand">
      <div class="brand-icon">
        <svg viewBox="0 0 32 32" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="16" cy="16" r="14" />
          <path d="M10 20l4-6 4 3 4-7" stroke-width="2.5" />
        </svg>
      </div>
      <div class="brand-text">
        <span class="brand-name">智投AI</span>
        <span class="brand-desc">智能荐股平台</span>
      </div>
    </div>

    <nav class="sidebar-nav">
      <!-- Single items -->
      <RouterLink to="/" class="nav-item" :class="{ active: isActive('/', true) }" @click="closeAll">
        <span class="nav-icon" v-html="homeIcon"></span>
        <span class="nav-label">首页</span>
      </RouterLink>
      <RouterLink to="/search" class="nav-item" :class="{ active: isActive('/search') }" @click="closeAll">
        <span class="nav-icon" v-html="searchIcon"></span>
        <span class="nav-label">搜索</span>
      </RouterLink>

      <!-- Section: 推荐股票 -->
      <div class="sidebar-section">
        <button class="section-header" @click="toggleSection('recs')">
          <span class="section-icon" v-html="starIcon"></span>
          <span>推荐股票</span>
          <span class="section-arrow" :class="{ open: openSections.recs }">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M6 9l6 6 6-6"/></svg>
          </span>
        </button>
        <div class="section-items" :class="{ collapsed: !openSections.recs }">
          <RouterLink to="/recommendations" class="sub-item" :class="{ active: isActive('/recommendations', true) }">每日推荐</RouterLink>
          <RouterLink to="/recommendations/history" class="sub-item" :class="{ active: isActive('/recommendations/history') }">历史表现</RouterLink>
          <RouterLink to="/recommendations/strategies" class="sub-item" :class="{ active: isActive('/recommendations/strategies') }">交易策略</RouterLink>
          <RouterLink to="/recommendations/backtest" class="sub-item" :class="{ active: isActive('/recommendations/backtest') }">回测结果</RouterLink>
          <RouterLink to="/futures/arbitrage" class="sub-item" :class="{ active: isActive('/futures/arbitrage') }">期货套利</RouterLink>
        </div>
      </div>

      <!-- Section: 鉴别分析 -->
      <div class="sidebar-section">
        <button class="section-header" @click="toggleSection('identify')">
          <span class="section-icon" v-html="searchIcon"></span>
          <span>鉴别分析</span>
          <span class="section-arrow" :class="{ open: openSections.identify }">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M6 9l6 6 6-6"/></svg>
          </span>
        </button>
        <div class="section-items" :class="{ collapsed: !openSections.identify }">
          <RouterLink to="/identify" class="sub-item" :class="{ active: isActive('/identify', true) }">股票分析</RouterLink>
          <RouterLink to="/identify/logic" class="sub-item" :class="{ active: isActive('/identify/logic') }">推理逻辑</RouterLink>
        </div>
      </div>

      <!-- Single: 资讯中心 -->
      <RouterLink to="/news" class="nav-item" :class="{ active: isActive('/news', true) }">
        <span class="nav-icon" v-html="newsIcon"></span>
        <span class="nav-label">资讯中心</span>
      </RouterLink>

      <!-- Section: 社区 -->
      <div class="sidebar-section">
        <button class="section-header" @click="toggleSection('community')">
          <span class="section-icon" v-html="communityIcon"></span>
          <span>社区</span>
          <span class="section-arrow" :class="{ open: openSections.community }">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M6 9l6 6 6-6"/></svg>
          </span>
        </button>
        <div class="section-items" :class="{ collapsed: !openSections.community }">
          <RouterLink to="/community" class="sub-item" :class="{ active: isActive('/community', true) }">社区动态</RouterLink>
          <RouterLink to="/community/room/300750.SZ" class="sub-item" :class="{ active: isActive('/community/room') }">股票讨论圈</RouterLink>
        </div>
      </div>

      <!-- Section: 我的 -->
      <div class="sidebar-section">
        <button class="section-header" @click="toggleSection('user')">
          <span class="section-icon" v-html="userIcon"></span>
          <span>我的</span>
          <span class="section-arrow" :class="{ open: openSections.user }">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><path d="M6 9l6 6 6-6"/></svg>
          </span>
        </button>
        <div class="section-items" :class="{ collapsed: !openSections.user }">
          <RouterLink to="/user" class="sub-item" :class="{ active: isActive('/user', true) }">个人中心</RouterLink>
          <RouterLink to="/user/vip" class="sub-item" :class="{ active: isActive('/user/vip') }">VIP会员</RouterLink>
          <RouterLink to="/user/history" class="sub-item" :class="{ active: isActive('/user/history') }">使用记录</RouterLink>
          <RouterLink to="/user/settings" class="sub-item" :class="{ active: isActive('/user/settings') }">设置</RouterLink>
        </div>
      </div>
    </nav>

    <div class="sidebar-footer">
      <div class="ai-status">
        <span class="status-dot"></span>
        <span>AI 实时分析中</span>
      </div>
    </div>
  </aside>
</template>

<script setup>
import { computed, reactive } from "vue";
import { useRoute } from "vue-router";

const route = useRoute();

const homeIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M4 12l8-8 8 8"/><path d="M6 10v8h4v-4h4v4h4v-8"/></svg>';
const searchIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="11" cy="11" r="7"/><path d="M16.5 16.5L21 21"/></svg>';
const starIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>';
const newsIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M5 5h14v16H5z"/><path d="M8 9h8M8 13h6M8 17h4"/></svg>';
const communityIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2"/><circle cx="9" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87"/><path d="M16 3.13a4 4 0 010 7.75"/></svg>';
const userIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="8" r="4"/><path d="M4 20c1.8-4 4.2-6 8-6s6.2 2 8 6"/></svg>';

const openSections = reactive({
  recs: false,
  identify: false,
  community: false,
  user: false
});

// Auto-open section based on current route
const currentSection = computed(() => route.meta?.section);
Object.keys(openSections).forEach(key => {
  if (currentSection.value === key) openSections[key] = true;
});

function toggleSection(name) {
  openSections[name] = !openSections[name];
}

function isActive(path, exact) {
  if (exact) return route.path === path || route.path.startsWith(path + '/');
  return route.path.startsWith(path);
}

function closeAll() {
  Object.keys(openSections).forEach(k => { openSections[k] = false; });
}
</script>

<style scoped>
.pc-sidebar {
  width: 220px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 20px 0;
  flex-shrink: 0;
  overflow-y: auto;
}

.sidebar-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px 20px;
  border-bottom: 1px solid var(--border);
  margin-bottom: 12px;
}

.brand-icon {
  width: 36px;
  height: 36px;
  color: var(--accent-gold);
  filter: drop-shadow(0 0 8px rgba(240, 185, 11, 0.3));
}

.brand-text {
  display: flex;
  flex-direction: column;
}

.brand-name {
  font-size: 16px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--accent-gold), #fff);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.brand-desc {
  font-size: 11px;
  color: var(--text-secondary);
}

.sidebar-nav {
  flex: 1;
  padding: 0 8px;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  color: var(--text-secondary);
  transition: all 0.2s;
  cursor: pointer;
}

.nav-item:hover {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.04);
}

.nav-item.active {
  color: var(--accent-gold);
  background: var(--accent-gold-glow);
}

.nav-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.nav-icon svg {
  width: 18px;
  height: 18px;
}

/* Section */
.sidebar-section {
  display: flex;
  flex-direction: column;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border-radius: var(--radius-sm);
  font-size: 14px;
  color: var(--text-secondary);
  width: 100%;
  text-align: left;
  transition: all 0.2s;
}

.section-header:hover {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.04);
}

.section-icon {
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.section-icon svg {
  width: 18px;
  height: 18px;
}

.section-arrow {
  margin-left: auto;
  transition: transform 0.25s;
  display: flex;
  align-items: center;
}

.section-arrow.open {
  transform: rotate(180deg);
}

.section-items {
  overflow: hidden;
  max-height: 200px;
  transition: max-height 0.3s ease;
  padding-left: 20px;
}

.section-items.collapsed {
  max-height: 0;
}

.sub-item {
  display: block;
  padding: 7px 12px;
  font-size: 13px;
  color: var(--text-muted);
  border-radius: var(--radius-sm);
  transition: all 0.2s;
}

.sub-item:hover {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.03);
}

.sub-item.active {
  color: var(--accent-gold);
  background: var(--accent-gold-glow);
}

.sidebar-footer {
  padding: 16px 20px 0;
  border-top: 1px solid var(--border);
  margin-top: auto;
}

.ai-status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--positive);
  animation: pulse-glow 2s infinite;
}
</style>
