<template>
  <div class="finance-shell">
    <div class="finance-topbar fade-up">
      <div class="shell-container finance-topbar-inner">
        <div class="finance-topbar-links">
          <span>真实前端开发中</span>
          <span>中文站风格改造</span>
          <span>PC + H5 同时兼容</span>
        </div>
        <div class="finance-topbar-right">
          <span>{{ activeTab.signal }} · {{ activeTab.short }}</span>
        </div>
      </div>
    </div>

    <header class="finance-header fade-up">
      <div class="shell-container finance-header-inner">
        <RouterLink class="finance-brand" to="/home">
          <span class="finance-brand-mark">S</span>
          <div>
            <p class="finance-brand-title">SercherAI</p>
            <p class="finance-brand-desc">推荐、研报、复盘与账户体系一体化客户端</p>
          </div>
        </RouterLink>

        <div class="finance-header-note">
          当前真实页面正在按 demo 逐页收敛为更符合中文用户习惯的金融站结构，优先改决策链路，再改账户链路。
        </div>

        <div class="finance-header-side">
          <div class="finance-status-chip">
            <span>当前模块</span>
            <strong>{{ activeTab.label }}</strong>
          </div>
          <template v-if="isLoggedIn">
            <div class="finance-account-chip">
              <span>已登录</span>
              <strong>{{ accountLabel }}</strong>
            </div>
            <button class="finance-logout-btn" type="button" :disabled="loggingOut" @click="handleLogout">
              {{ loggingOut ? "退出中..." : "退出" }}
            </button>
          </template>
          <RouterLink
            v-else
            class="finance-login-btn"
            :to="{ path: '/auth', query: { redirect: route.fullPath } }"
          >
            登录 / 注册
          </RouterLink>
        </div>
      </div>
    </header>

    <nav class="finance-nav fade-up">
      <div class="shell-container finance-nav-inner">
        <div class="finance-nav-row">
          <RouterLink
            v-for="item in tabs"
            :key="item.path"
            :to="item.path"
            class="finance-nav-link"
            :class="{ active: route.path === item.path }"
          >
            {{ item.label }}
          </RouterLink>
        </div>
        <div class="finance-nav-note">{{ activeTab.focus }}</div>
      </div>
    </nav>

    <section class="finance-route-strip fade-up">
      <div class="shell-container finance-route-strip-inner">
        <div class="finance-route-copy">
          <p>{{ activeTab.label }} · 当前页</p>
          <strong>{{ activeTab.desc }}</strong>
        </div>
        <div v-if="activeHighlights.length" class="finance-route-tags">
          <span v-for="item in activeHighlights" :key="item">{{ item }}</span>
        </div>
      </div>
    </section>

    <main class="finance-view">
      <div class="shell-container">
        <RouterView />
      </div>
    </main>

    <footer class="finance-footer">
      <div class="shell-container finance-footer-inner">
        <div class="finance-footer-main">
          <div class="finance-footer-brand">
            <p class="finance-footer-title">SercherAI Client</p>
            <p class="finance-footer-desc">
              当前客户端按中文金融站风格持续收敛，优先承接推荐、研报、档案、关注与账户链路，不盲目扩成行情门户。
            </p>
          </div>

          <div class="finance-footer-columns">
            <section class="finance-footer-col">
              <p class="finance-footer-heading">页面结构</p>
              <div class="finance-footer-links">
                <RouterLink v-for="item in tabs" :key="`footer-${item.path}`" :to="item.path">
                  {{ item.label }}
                </RouterLink>
              </div>
            </section>

            <section class="finance-footer-col">
              <p class="finance-footer-heading">当前支持</p>
              <div class="finance-footer-tags">
                <span v-for="item in footerSupportItems" :key="item">{{ item }}</span>
              </div>
            </section>
          </div>
        </div>

        <div class="finance-footer-meta">
          <p>适配说明：同一套页面同时兼容 PC 与 H5，PC 以信息密度优先，H5 以阅读顺序压缩。</p>
          <p>风险提示：页面内容仅作研究展示与产品演示说明，不构成任何投资建议。</p>
        </div>
      </div>
    </footer>

    <nav class="finance-mobile-nav">
      <RouterLink
        v-for="item in mobileTabs"
        :key="item.path"
        :to="item.path"
        class="finance-mobile-link"
        :class="{ active: route.path === item.path }"
      >
        <span class="finance-mobile-dot" />
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
    short: "Home",
    signal: "推荐与研报",
    focus: "先看推荐，再看研报解读。",
    desc: "首页收敛为推荐区和研报区双主结构，先回答今天看什么，再回答为什么看。",
    highlights: ["主推荐", "历史样本", "研报解读"]
  },
  {
    path: "/strategies",
    label: "策略",
    short: "Strategy",
    signal: "深度解释",
    focus: "把首页推荐展开成可执行判断。",
    desc: "策略页负责解释推荐、参数、风险边界、版本与事件，不再只是列表工具页。",
    highlights: ["股票详情", "期货参数", "事件雷达"]
  },
  {
    path: "/archive",
    label: "档案",
    short: "Archive",
    signal: "历史复盘",
    focus: "把历史结果、原因和变化一起摆出来。",
    desc: "档案页负责建立信任：成功、失效、止盈和止损样本都应该能完整回看。",
    highlights: ["历史时间线", "版本差异", "来源说明"]
  },
  {
    path: "/watchlist",
    label: "关注",
    short: "Watchlist",
    signal: "变化跟踪",
    focus: "每次回来先看你的关注发生了什么变化。",
    desc: "关注页是回访工作台，重点看状态变化、资讯变化、风险边界变化和结论变化。",
    highlights: ["状态变化", "资讯变化", "版本对照"]
  },
  {
    path: "/news",
    label: "资讯",
    short: "News",
    signal: "深度阅读",
    focus: "首页导读，资讯页深读。",
    desc: "资讯页继续承接新闻、研报、期刊、正文、附件和权限，不做混合瀑布流。",
    highlights: ["研报", "正文", "附件"]
  },
  {
    path: "/membership",
    label: "会员",
    short: "Membership",
    signal: "转化承接",
    focus: "讲清楚为什么升级，而不是只展示价格。",
    desc: "会员页负责把推荐、研报、复盘和关注串成可感知的升级价值，并接住支付和实名激活。",
    highlights: ["升级价值", "订单状态", "实名激活"]
  },
  {
    path: "/profile",
    label: "我的",
    short: "Profile",
    signal: "账户管理",
    focus: "先看待办，再看查询。",
    desc: "个人中心是账户管理台，承接 VIP、支付、订阅、通知和邀请等账户级信息。",
    highlights: ["今日行动板", "查询中心", "快捷入口"]
  }
];

const activeTab = computed(() => tabs.find((item) => item.path === route.path) || tabs[0]);
const activeHighlights = computed(() => activeTab.value.highlights || []);
const mobileTabs = computed(() => tabs);
const footerSupportItems = [
  "股票推荐",
  "期货策略",
  "历史档案",
  "我的关注",
  "研报与资讯",
  "会员与账户"
];

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
