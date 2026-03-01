<template>
  <div class="client-shell">
    <header class="top-nav fade-up">
      <div class="top-nav-inner card">
        <div class="brand-block">
          <span class="brand-mark">S</span>
          <div>
            <p class="brand-name">SercherAI Client</p>
            <p class="brand-note">策略执行与风险协同</p>
          </div>
        </div>

        <nav class="desktop-nav">
          <RouterLink
            v-for="item in tabs"
            :key="item.path"
            :to="item.path"
            class="nav-link"
            :class="{ active: route.path === item.path }"
          >
            <span class="nav-label">{{ item.label }}</span>
            <span class="nav-sub">{{ item.short }}</span>
          </RouterLink>
        </nav>

        <div class="nav-right">
          <div class="pulse-box">
            <p>当前模块</p>
            <strong>{{ activeTab.signal }}</strong>
          </div>
          <template v-if="isLoggedIn">
            <div class="account-chip">
              <p>已登录</p>
              <strong>{{ accountLabel }}</strong>
            </div>
            <button class="logout-btn" type="button" :disabled="loggingOut" @click="handleLogout">
              {{ loggingOut ? "退出中..." : "退出" }}
            </button>
          </template>
          <RouterLink
            v-else
            class="launch-btn"
            :to="{ path: '/auth', query: { redirect: route.fullPath } }"
          >
            登录 / 注册
          </RouterLink>
        </div>
      </div>
    </header>

    <section class="context-banner card fade-up" :class="{ 'home-focus': activeTab.path === '/home' }">
      <div class="banner-main">
        <p class="banner-eyebrow">{{ activeTab.label }} · 模块焦点</p>
        <h1>{{ activeTab.focus }}</h1>
        <p class="banner-desc">{{ activeTab.desc }}</p>
        <div v-if="activeHighlights.length" class="banner-highlights">
          <span v-for="item in activeHighlights" :key="item">{{ item }}</span>
        </div>
        <div v-if="activeFocusSteps.length" class="banner-steps">
          <article v-for="(item, index) in activeFocusSteps" :key="item.title">
            <p>步骤 {{ index + 1 }}</p>
            <strong>{{ item.title }}</strong>
            <span>{{ item.desc }}</span>
          </article>
        </div>
        <div v-if="activeActions.length" class="banner-actions">
          <RouterLink
            v-for="item in activeActions"
            :key="`${activeTab.path}-${item.to}-${item.label}`"
            :to="item.to"
            class="banner-action"
          >
            <p>{{ item.label }}</p>
            <span>{{ item.desc }}</span>
          </RouterLink>
        </div>
      </div>
      <div class="banner-stats">
        <article v-for="item in bannerStats" :key="item.label">
          <p>{{ item.label }}</p>
          <strong>{{ item.value }}</strong>
        </article>
      </div>
    </section>

    <main class="view-wrap">
      <RouterView />
    </main>

    <nav class="mobile-tab">
      <RouterLink
        v-for="item in tabs"
        :key="item.path"
        :to="item.path"
        class="mobile-link"
        :class="{ active: route.path === item.path }"
      >
        <span class="mobile-dot" />
        <span>{{ item.label }}</span>
      </RouterLink>
    </nav>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { logout } from "../api/auth";
import { clearClientAuthSession, useClientAuth } from "../lib/client-auth";

const route = useRoute();
const router = useRouter();
const loggingOut = ref(false);
const { session, isLoggedIn } = useClientAuth();

const tabs = [
  {
    path: "/home",
    label: "首页",
    short: "Overview",
    signal: "决策中枢",
    focus: "先检索资讯，再执行策略。",
    desc: "首页聚合资讯搜索、行情推荐与VIP入口，先完成信息筛选，再进入具体交易动作。",
    highlights: ["资讯搜索直达", "焦点内容同屏", "VIP入口智能展示"],
    steps: [
      { title: "先搜资讯", desc: "输入关键词过滤新闻、研报、公告线索" },
      { title: "再看策略", desc: "对比推荐股票与套利区间，锁定执行对象" },
      { title: "最后执行", desc: "进入策略或会员页完成下一步操作" }
    ],
    actions: [
      { label: "去资讯中心", desc: "查看完整资讯与分页内容", to: "/news" },
      { label: "去策略中心", desc: "查看推荐与事件联动", to: "/strategies" },
      { label: "去会员中心", desc: "开通VIP解锁专享内容", to: "/membership" }
    ]
  },
  {
    path: "/strategies",
    label: "策略",
    short: "Strategy",
    signal: "策略池",
    focus: "优先推荐当前阶段最适合的策略。",
    desc: "按风险等级和执行成本排序，先看关键策略，再看参数细节。",
    highlights: ["按风险排序", "参数清晰可比", "事件触发联动"],
    actions: [
      { label: "回首页", desc: "从首页统一筛选后再执行", to: "/home" },
      { label: "看资讯线索", desc: "结合资讯再确认策略方向", to: "/news" }
    ]
  },
  {
    path: "/news",
    label: "资讯",
    short: "News",
    signal: "风险线索",
    focus: "信息不过载，只保留可执行线索。",
    desc: "为每条资讯标注影响方向和处理建议，避免只看信息不落地。",
    highlights: ["分类切换查看", "支持关键词检索", "支持分页加载"],
    actions: [
      { label: "回首页搜索", desc: "从首页入口继续筛选资讯", to: "/home" },
      { label: "看策略落地", desc: "将资讯线索转为策略动作", to: "/strategies" }
    ]
  },
  {
    path: "/membership",
    label: "会员",
    short: "Membership",
    signal: "权益提升",
    focus: "清楚展示升级后你会得到什么。",
    desc: "用能力分层取代营销堆砌，让用户快速判断是否值得升级。",
    highlights: ["会员权益透明", "订单状态跟踪", "支付流程闭环"],
    actions: [
      { label: "回首页", desc: "继续查看焦点资讯与推荐", to: "/home" },
      { label: "去个人中心", desc: "查看订单、配额与订阅状态", to: "/profile" }
    ]
  },
  {
    path: "/profile",
    label: "我的",
    short: "Profile",
    signal: "账户状态",
    focus: "账户健康、数据、操作入口集中管理。",
    desc: "把高频个人操作前置，减少路径层级并保留安全控制。",
    highlights: ["账户数据聚合", "邀请关系管理", "订阅状态可控"],
    actions: [
      { label: "回首页", desc: "返回决策中枢继续执行", to: "/home" },
      { label: "去会员中心", desc: "查看权益与支付状态", to: "/membership" }
    ]
  }
];

const routeStats = {
  "/home": [
    { label: "资讯搜索", value: "已启用" },
    { label: "焦点资讯", value: "实时更新" }
  ],
  "/strategies": [
    { label: "可用策略", value: "16 套" },
    { label: "本周更新", value: "5 次" }
  ],
  "/news": [
    { label: "高优先资讯", value: "3 条" },
    { label: "风险警示", value: "2 条" }
  ],
  "/membership": [
    { label: "当前等级", value: "基础版" },
    { label: "可升级权益", value: "8 项" }
  ],
  "/profile": [
    { label: "账户安全", value: "B+" },
    { label: "本周执行", value: "9 次" }
  ]
};

const activeTab = computed(() => tabs.find((item) => item.path === route.path) || tabs[0]);

const bannerStats = computed(() => routeStats[activeTab.value.path] || routeStats["/home"]);
const activeHighlights = computed(() => activeTab.value.highlights || []);
const activeFocusSteps = computed(() => activeTab.value.steps || []);
const activeActions = computed(() => activeTab.value.actions || []);

const accountLabel = computed(() => {
  const phone = session.value?.phone || "";
  if (/^\d{11}$/.test(phone)) {
    return `${phone.slice(0, 3)}****${phone.slice(-4)}`;
  }
  const email = String(session.value?.email || "").trim();
  if (email.includes("@")) {
    const parts = email.split("@");
    const local = parts[0] || "";
    const domain = parts[1] || "";
    if (local.length <= 2) {
      return `${local.slice(0, 1)}***@${domain}`;
    }
    return `${local.slice(0, 2)}***@${domain}`;
  }
  return session.value?.userID || "当前用户";
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
    console.warn("client logout failed:", error?.message || error);
  } finally {
    clearClientAuthSession();
    loggingOut.value = false;
    await router.replace("/auth");
  }
}
</script>

<style scoped>
.client-shell {
  min-height: 100vh;
  padding: 14px 0 102px;
}

.top-nav {
  width: min(1180px, calc(100% - 28px));
  margin: 0 auto;
}

.top-nav-inner {
  border-radius: var(--radius-xl);
  padding: 12px 14px;
  display: grid;
  grid-template-columns: auto 1fr auto;
  align-items: center;
  gap: 12px;
}

.brand-block {
  display: flex;
  align-items: center;
  gap: 10px;
}

.brand-mark {
  width: 44px;
  height: 44px;
  border-radius: 15px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(155deg, var(--color-pine-700), var(--color-pine-500));
  color: #fff;
  font-size: 20px;
  font-family: var(--font-serif);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.24);
}

.brand-name {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  letter-spacing: 0.3px;
}

.brand-note {
  margin: 3px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.desktop-nav {
  justify-self: center;
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 6px;
  border-radius: 14px;
  border: 1px solid rgba(216, 223, 216, 0.82);
  background: rgba(252, 250, 245, 0.9);
}

.nav-link {
  min-width: 86px;
  padding: 7px 8px;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1px;
  color: var(--color-text-sub);
  transition: all 0.2s ease;
}

.nav-link:hover {
  color: var(--color-pine-700);
}

.nav-link.active {
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.2);
}

.nav-label {
  font-size: 14px;
  font-weight: 600;
}

.nav-sub {
  font-size: 11px;
  opacity: 0.76;
}

.nav-right {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.account-chip {
  border-radius: 12px;
  padding: 7px 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(252, 251, 247, 0.9);
}

.account-chip p {
  margin: 0;
  font-size: 11px;
  color: var(--color-text-sub);
}

.account-chip strong {
  margin-top: 2px;
  display: block;
  font-size: 13px;
  color: var(--color-pine-700);
}

.pulse-box {
  border-radius: 12px;
  padding: 7px 10px;
  border: 1px solid rgba(63, 127, 113, 0.3);
  background: rgba(235, 245, 241, 0.82);
}

.pulse-box p {
  margin: 0;
  font-size: 11px;
  color: var(--color-text-sub);
}

.pulse-box strong {
  margin-top: 2px;
  display: block;
  font-size: 13px;
  color: var(--color-pine-700);
}

.launch-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 0;
  border-radius: 12px;
  padding: 10px 14px;
  background: linear-gradient(145deg, #d7b67f, #b4833e);
  color: #fff;
  cursor: pointer;
  font-weight: 600;
}

.logout-btn {
  border: 1px solid rgba(63, 127, 113, 0.35);
  border-radius: 12px;
  padding: 9px 12px;
  background: rgba(236, 245, 242, 0.92);
  color: var(--color-pine-700);
  cursor: pointer;
  font-weight: 600;
}

.logout-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.context-banner {
  width: min(1180px, calc(100% - 28px));
  margin: 12px auto 18px;
  border-radius: 20px;
  padding: 14px 16px;
  display: grid;
  gap: 10px;
  grid-template-columns: 1fr auto;
  align-items: center;
  background:
    radial-gradient(circle at 100% 0%, rgba(63, 127, 113, 0.16) 0%, transparent 35%),
    radial-gradient(circle at 0% 100%, rgba(234, 215, 180, 0.3) 0%, transparent 34%),
    rgba(255, 255, 255, 0.92);
}

.context-banner.home-focus {
  background:
    radial-gradient(circle at 100% 0%, rgba(63, 127, 113, 0.24) 0%, transparent 33%),
    radial-gradient(circle at 0% 100%, rgba(234, 215, 180, 0.38) 0%, transparent 36%),
    rgba(255, 255, 255, 0.94);
}

.banner-eyebrow {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.banner-main h1 {
  margin: 5px 0 4px;
  font-size: clamp(22px, 2.4vw, 30px);
  font-family: var(--font-serif);
  line-height: 1.2;
}

.banner-desc {
  margin: 0;
  color: var(--color-text-sub);
  font-size: 14px;
  max-width: 700px;
}

.banner-highlights {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.banner-highlights span {
  border-radius: 999px;
  border: 1px solid rgba(216, 223, 216, 0.92);
  background: rgba(252, 251, 247, 0.88);
  padding: 4px 10px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.banner-steps {
  margin-top: 10px;
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.banner-steps article {
  border-radius: 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(252, 251, 247, 0.86);
  padding: 8px 9px;
}

.banner-steps p {
  margin: 0;
  font-size: 11px;
  color: var(--color-pine-600);
}

.banner-steps strong {
  margin-top: 3px;
  display: block;
  font-size: 13px;
  color: var(--color-text-main);
}

.banner-steps span {
  margin-top: 3px;
  display: block;
  font-size: 12px;
  color: var(--color-text-sub);
  line-height: 1.45;
}

.banner-actions {
  margin-top: 10px;
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.banner-action {
  border-radius: 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 9px 10px;
  transition: all 0.2s ease;
}

.banner-action:hover {
  border-color: rgba(63, 127, 113, 0.4);
  box-shadow: 0 8px 20px rgba(63, 127, 113, 0.12);
}

.banner-action p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-pine-700);
}

.banner-action span {
  margin-top: 3px;
  display: block;
  font-size: 12px;
  color: var(--color-text-sub);
  line-height: 1.45;
}

.banner-stats {
  display: inline-grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.banner-stats article {
  min-width: 118px;
  border-radius: 12px;
  padding: 9px 10px;
  border: 1px solid rgba(216, 223, 216, 0.88);
  background: rgba(252, 251, 247, 0.9);
}

.banner-stats p {
  margin: 0;
  font-size: 11px;
  color: var(--color-text-sub);
}

.banner-stats strong {
  margin-top: 3px;
  display: block;
  font-size: 17px;
  color: var(--color-pine-700);
}

.view-wrap {
  width: min(1180px, calc(100% - 28px));
  margin: 0 auto;
}

.mobile-tab {
  position: fixed;
  left: 50%;
  bottom: 8px;
  transform: translateX(-50%);
  width: min(640px, calc(100% - 16px));
  border-radius: 18px;
  padding: 8px 6px 10px;
  border: 1px solid rgba(216, 223, 216, 0.92);
  background: rgba(255, 255, 255, 0.95);
  box-shadow: 0 12px 26px rgba(22, 40, 34, 0.16);
  display: none;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 4px;
  z-index: 30;
}

.mobile-link {
  min-height: 50px;
  border-radius: 10px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 5px;
  color: var(--color-text-sub);
  font-size: 12px;
}

.mobile-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: rgba(63, 127, 113, 0.24);
}

.mobile-link.active {
  color: var(--color-pine-700);
  background: rgba(226, 238, 232, 0.66);
}

.mobile-link.active .mobile-dot {
  background: var(--color-pine-600);
}

@media (max-width: 1080px) {
  .desktop-nav {
    gap: 2px;
  }

  .nav-link {
    min-width: 78px;
  }

  .launch-btn {
    padding: 9px 11px;
    font-size: 13px;
  }

  .account-chip {
    display: none;
  }
}

@media (max-width: 900px) {
  .client-shell {
    padding-top: 8px;
  }

  .top-nav,
  .context-banner,
  .view-wrap {
    width: calc(100% - 20px);
  }

  .top-nav-inner {
    grid-template-columns: 1fr auto;
    border-radius: 16px;
  }

  .brand-note {
    display: none;
  }

  .desktop-nav {
    display: none;
  }

  .pulse-box {
    display: none;
  }

  .context-banner {
    grid-template-columns: 1fr;
    gap: 12px;
    padding: 13px;
  }

  .banner-main h1 {
    font-size: 24px;
  }

  .banner-desc {
    display: none;
  }

  .banner-steps,
  .banner-actions {
    grid-template-columns: 1fr;
  }

  .banner-stats {
    width: 100%;
  }

  .banner-stats article {
    min-width: auto;
  }

  .mobile-tab {
    display: grid;
  }
}
</style>
