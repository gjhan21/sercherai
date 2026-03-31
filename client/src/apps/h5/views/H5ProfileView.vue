<template>
  <div class="h5-page fade-up profile-page">
    <div class="h5-page-topline">
      <span class="h5-page-tagline">账户中心</span>
      <span>{{ lastUpdatedAt || "我的" }}</span>
    </div>

    <section class="profile-hero">
      <div class="profile-hero-head">
        <div class="profile-hero-user">
          <div class="profile-avatar">AI</div>
          <div class="profile-hero-copy">
            <span class="profile-hero-kicker">SercherAI 会员账户</span>
            <h1>{{ profileModel.hero.displayName }}</h1>
            <p>{{ profileModel.hero.description }}</p>
          </div>
        </div>
        <button type="button" class="profile-refresh" :disabled="loading" @click="loadProfilePage">
          {{ loading ? "同步中" : "刷新" }}
        </button>
      </div>

      <div class="profile-status-strip">
        <span class="profile-status-pill brand">{{ profileModel.hero.memberLevel }}</span>
        <span class="profile-status-pill gold">{{ profileModel.hero.activationState }}</span>
        <span class="profile-status-pill soft">{{ profileModel.hero.vipStatus }}</span>
      </div>

      <div class="profile-metric-grid">
        <article v-for="item in profileModel.hero.metrics" :key="item.label" class="profile-metric-item">
          <span>{{ item.label }}</span>
          <strong>{{ item.value }}</strong>
          <p>{{ item.note }}</p>
        </article>
      </div>

      <div class="profile-hero-actions">
        <button type="button" class="h5-btn" @click="goMembership">会员中心</button>
        <button type="button" class="h5-btn-secondary" @click="handleTargetAction(profileModel.sticky.primaryTarget)">
          {{ profileModel.sticky.primaryLabel }}
        </button>
      </div>

      <p v-if="loadError" class="profile-inline-note">{{ loadError }}</p>
    </section>

    <section class="profile-card">
      <div class="profile-section-head">
        <div>
          <strong>今日待办</strong>
          <span>先处理最影响账户状态和继续阅读的动作</span>
        </div>
      </div>

      <div class="profile-todo-list">
        <article v-for="item in profileModel.todos" :key="item.id" class="profile-todo-item">
          <div class="profile-todo-copy">
            <div class="profile-todo-topline">
              <strong>{{ item.title }}</strong>
              <span class="h5-badge" :class="item.tone">{{ item.badge }}</span>
            </div>
            <p>{{ item.desc }}</p>
          </div>
          <button type="button" class="h5-btn-secondary" @click="handleTargetAction(item.id)">{{ item.actionLabel }}</button>
        </article>
      </div>
    </section>

    <section class="profile-card">
      <div class="profile-section-head">
        <div>
          <strong>常用入口</strong>
          <span>像 App 账户页一样，把高频入口前置成单手可达区</span>
        </div>
      </div>

      <div class="profile-shortcut-grid">
        <button
          v-for="item in profileModel.shortcuts"
          :key="item.id"
          type="button"
          class="profile-shortcut"
          @click="handleTargetAction(item.id)"
        >
          <span class="profile-shortcut-icon">{{ shortcutGlyph(item.id) }}</span>
          <strong>{{ item.title }}</strong>
          <small>{{ item.note }}</small>
        </button>
      </div>
    </section>

    <section ref="moduleSectionRef" class="profile-card">
      <div class="profile-section-head">
        <div>
          <strong>我的二级模块</strong>
          <span>社区负责发现与讨论，关注只作为“我的”里的二级模块承接</span>
        </div>
      </div>

      <div class="profile-module-grid">
        <button
          v-for="item in profileModuleCards"
          :key="item.id"
          type="button"
          class="profile-module-card"
          :class="{ active: item.active }"
          @click="handleTargetAction(item.id)"
        >
          <span>{{ item.note }}</span>
          <strong>{{ item.title }}</strong>
          <p>{{ item.desc }}</p>
        </button>
      </div>
    </section>

    <section ref="watchlistEntryRef" class="profile-card profile-card-accent">
      <div class="profile-section-head compact">
        <div>
          <strong>我的关注</strong>
          <span>先在“我的”里确认入口，再进入详情页继续回访</span>
        </div>
      </div>

      <div class="profile-service-list">
        <article class="profile-service-card active">
          <div class="profile-service-copy">
            <strong>{{ activeProfileSection === 'watchlist' ? '当前聚焦：我的关注' : '我的关注详情' }}</strong>
            <p>关注不再作为 H5 一级入口暴露，统一从个人中心进入，再查看变化工作台详情。</p>
          </div>
          <div class="profile-service-meta">
            <span class="h5-meta-chip">我的入口</span>
            <span class="h5-meta-chip">变化回访</span>
          </div>
        </article>
      </div>

      <div class="profile-hero-actions">
        <button type="button" class="h5-btn" @click="handleTargetAction('watchlist')">进入关注详情</button>
        <button type="button" class="h5-btn-secondary" @click="goStrategies">去策略页补充标的</button>
      </div>
    </section>


    <section class="profile-card">
      <div class="profile-section-head">
        <div>
          <strong>账户服务</strong>
          <span>会员、消息、邀请三类服务做成账户卡片流</span>
        </div>
      </div>

      <div class="profile-service-list">
        <button
          v-for="item in profileModel.serviceCards"
          :key="item.id"
          type="button"
          class="profile-service-card"
          @click="handleTargetAction(item.id)"
        >
          <div class="profile-service-copy">
            <strong>{{ item.title }}</strong>
            <p>{{ item.summary }}</p>
          </div>
          <div class="profile-service-meta">
            <span v-for="tag in item.tags" :key="tag" class="h5-meta-chip">{{ tag }}</span>
          </div>
        </button>
      </div>
    </section>

    <section ref="messageSectionRef" class="profile-card">
      <div class="profile-section-head">
        <div>
          <strong>消息提醒</strong>
          <span>未读优先，保持单列阅读节奏</span>
        </div>
      </div>

      <div v-if="profileModel.messageCards.length" class="profile-message-list">
        <article v-for="item in profileModel.messageCards" :key="item.id" class="profile-message-item">
          <div class="profile-message-topline">
            <div>
              <span>{{ item.time }}</span>
              <strong>{{ item.title }}</strong>
            </div>
            <span class="h5-badge" :class="item.read ? 'success' : 'gold'">{{ item.read ? "已读" : item.typeLabel }}</span>
          </div>
          <p>{{ item.desc }}</p>
          <button v-if="!item.read" type="button" class="h5-btn-secondary" @click="handleReadMessage(item.raw)">标记已读</button>
        </article>
      </div>
      <H5EmptyState v-else title="暂无消息" description="系统通知、策略提醒和风险告警会显示在这里。" />
    </section>

    <section ref="inviteSectionRef" class="profile-card">
      <div class="profile-section-head">
        <div>
          <strong>邀请中心</strong>
          <span>分享链接、近 7 日注册与转化都集中在这里</span>
        </div>
        <button type="button" class="h5-btn-secondary" :disabled="creatingShareLink" @click="handleCreateShareLink">
          {{ creatingShareLink ? "创建中..." : "新建链接" }}
        </button>
      </div>

      <article class="profile-invite-overview">
        <div class="profile-invite-code">
          <span>当前邀请码</span>
          <strong>{{ profileModel.inviteOverview.primaryCode }}</strong>
          <p>{{ profileModel.inviteOverview.summary }}</p>
        </div>
        <div class="profile-invite-actions">
          <button
            type="button"
            class="h5-btn"
            :disabled="!profileModel.shareLinks.length"
            @click="profileModel.shareLinks.length && copyInviteLink(profileModel.shareLinks[0])"
          >
            复制链接
          </button>
          <small>{{ profileModel.inviteOverview.note }}</small>
        </div>
      </article>

      <div v-if="profileModel.shareLinks.length" class="profile-share-list">
        <article v-for="item in profileModel.shareLinks" :key="item.id" class="profile-share-item">
          <div>
            <strong>{{ item.code }}</strong>
            <p>{{ item.url }}</p>
          </div>
          <button type="button" class="h5-btn-secondary" @click="copyInviteLink(item)">复制</button>
        </article>
      </div>

      <div v-if="profileModel.inviteCards.length" class="profile-invite-list">
        <article v-for="item in profileModel.inviteCards" :key="item.id" class="profile-invite-item">
          <div class="profile-message-topline">
            <div>
              <span>{{ item.time }}</span>
              <strong>{{ item.title }}</strong>
            </div>
            <span class="h5-badge brand">{{ item.status }}</span>
          </div>
          <p>{{ item.desc }}</p>
        </article>
      </div>
      <H5EmptyState v-else title="暂无邀请记录" description="创建分享链接后，注册与首单转化会显示在这里。" />

      <p v-if="inviteMessage" class="profile-inline-note">{{ inviteMessage }}</p>
    </section>

    <H5StickyCta
      :title="profileModel.sticky.title"
      :description="stickyDescription"
      :primary-label="profileModel.sticky.primaryLabel"
      secondary-label="去会员页"
      @primary="handleTargetAction(profileModel.sticky.primaryTarget)"
      @secondary="goMembership"
    />
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import H5EmptyState from "../components/H5EmptyState.vue";
import H5StickyCta from "../components/H5StickyCta.vue";
import {
  createShareLink,
  getInviteSummary,
  getMembershipQuota,
  getUserProfile,
  listInviteRecords,
  listMembershipOrders,
  listMessages,
  listShareLinks,
  readMessage
} from "../../../api/userCenter";
import { shouldUseDemoFallback } from "../../../lib/fallback-policy";
import { buildProfileModuleRoute, normalizeProfileModuleSection } from "../../../lib/profile-modules";
import { formatDateTime, mapActivationState, toArray } from "../lib/formatters";
import {
  fallbackInviteRecords,
  fallbackInviteSummary,
  fallbackMembershipOrders,
  fallbackMessages,
  fallbackProfile,
  fallbackQuota,
  fallbackShareLinks
} from "../lib/mock-data";
import { buildProfileCenterModel } from "../lib/profile-center.js";

const route = useRoute();
const router = useRouter();
const useDemoFallback = shouldUseDemoFallback();

const loading = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const rawProfile = ref(useDemoFallback ? { ...fallbackProfile } : {});
const rawQuota = ref(useDemoFallback ? { ...fallbackQuota } : {});
const rawOrders = ref(useDemoFallback ? [...fallbackMembershipOrders] : []);
const rawMessages = ref(useDemoFallback ? [...fallbackMessages] : []);
const rawShareLinks = ref(useDemoFallback ? [...fallbackShareLinks] : []);
const rawInviteRecords = ref(useDemoFallback ? [...fallbackInviteRecords] : []);
const rawInviteSummary = ref(useDemoFallback ? { ...fallbackInviteSummary } : {});
const inviteMessage = ref("");
const creatingShareLink = ref(false);
const messageSectionRef = ref(null);
const inviteSectionRef = ref(null);
const moduleSectionRef = ref(null);
const watchlistEntryRef = ref(null);

const profileModel = computed(() => buildProfileCenterModel({
  profile: rawProfile.value,
  quota: rawQuota.value,
  orders: rawOrders.value,
  messages: rawMessages.value,
  shareLinks: rawShareLinks.value,
  inviteRecords: rawInviteRecords.value,
  inviteSummary: rawInviteSummary.value
}));

const stickyDescription = computed(() => inviteMessage.value || loadError.value || profileModel.value.sticky.description);
const activeProfileSection = computed(() => normalizeProfileModuleSection(route.query.section));
const profileModuleCards = computed(() => [
  {
    id: "watchlist",
    title: "我的关注",
    note: activeProfileSection.value === "watchlist" ? "当前聚焦" : "回访入口",
    desc: "先在我的页聚焦关注模块，再进入详情继续看变化、风险边界和下一步动作。",
    active: activeProfileSection.value === "watchlist"
  },
  {
    id: "community",
    title: "我的讨论",
    note: "承接预留",
    desc: "H5 暂不单开社区一级入口，个人讨论回访口径统一收在“我的”里。",
    active: activeProfileSection.value === "community"
  }
]);

async function loadProfilePage() {
  loading.value = true;
  loadError.value = "";
  const tasks = await Promise.allSettled([
    getUserProfile(),
    getMembershipQuota(),
    listMembershipOrders({ page: 1, page_size: 20 }),
    listMessages({ page: 1, page_size: 20 }),
    listShareLinks(),
    listInviteRecords({ page: 1, page_size: 20 }),
    getInviteSummary()
  ]);

  const errors = [];
  if (tasks[0].status === "fulfilled") rawProfile.value = tasks[0].value || {};
  else errors.push(`用户资料加载失败：${tasks[0].reason?.message || "unknown error"}`);
  if (tasks[1].status === "fulfilled") rawQuota.value = tasks[1].value || {};
  else errors.push(`会员状态加载失败：${tasks[1].reason?.message || "unknown error"}`);
  if (tasks[2].status === "fulfilled") rawOrders.value = toArray(tasks[2].value?.items, []);
  else errors.push(`订单加载失败：${tasks[2].reason?.message || "unknown error"}`);
  if (tasks[3].status === "fulfilled") rawMessages.value = toArray(tasks[3].value?.items, []);
  else errors.push(`消息加载失败：${tasks[3].reason?.message || "unknown error"}`);
  if (tasks[4].status === "fulfilled") rawShareLinks.value = toArray(tasks[4].value?.items, []);
  else errors.push(`分享链接加载失败：${tasks[4].reason?.message || "unknown error"}`);
  if (tasks[5].status === "fulfilled") rawInviteRecords.value = toArray(tasks[5].value?.items, []);
  else errors.push(`邀请记录加载失败：${tasks[5].reason?.message || "unknown error"}`);
  if (tasks[6].status === "fulfilled") rawInviteSummary.value = tasks[6].value || {};
  else errors.push(`邀请汇总加载失败：${tasks[6].reason?.message || "unknown error"}`);

  if (!Object.keys(rawProfile.value || {}).length && useDemoFallback) rawProfile.value = { ...fallbackProfile };
  if (!Object.keys(rawQuota.value || {}).length && useDemoFallback) rawQuota.value = { ...fallbackQuota };
  if (!rawOrders.value.length && useDemoFallback) rawOrders.value = [...fallbackMembershipOrders];
  if (!rawMessages.value.length && useDemoFallback) rawMessages.value = [...fallbackMessages];
  if (!rawShareLinks.value.length && useDemoFallback) rawShareLinks.value = [...fallbackShareLinks];
  if (!rawInviteRecords.value.length && useDemoFallback) rawInviteRecords.value = [...fallbackInviteRecords];
  if (!Object.keys(rawInviteSummary.value || {}).length && useDemoFallback) rawInviteSummary.value = { ...fallbackInviteSummary };

  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  loadError.value = errors.join("；");
  loading.value = false;
}


async function handleReadMessage(item) {
  if (!item?.id) {
    return;
  }
  try {
    await readMessage(item.id);
    rawMessages.value = rawMessages.value.map((row) => row.id === item.id ? { ...row, read_status: "READ" } : row);
  } catch (error) {
    loadError.value = error?.message || "标记已读失败";
  }
}

async function handleCreateShareLink() {
  if (creatingShareLink.value) {
    return;
  }
  creatingShareLink.value = true;
  inviteMessage.value = "";
  try {
    const result = await createShareLink({ channel: "APP" });
    if (result?.id) {
      rawShareLinks.value = [result, ...rawShareLinks.value];
    }
    inviteMessage.value = "已创建新的分享链接";
  } catch (error) {
    inviteMessage.value = error?.message || "创建分享链接失败";
  } finally {
    creatingShareLink.value = false;
  }
}

async function copyInviteLink(item) {
  if (!item?.url || typeof navigator === "undefined" || !navigator.clipboard) {
    inviteMessage.value = "当前环境不支持自动复制";
    return;
  }
  await navigator.clipboard.writeText(item.url);
  inviteMessage.value = `已复制邀请码 ${item.code || item.invite_code}`;
}

function goMembership() {
  router.push({ name: "h5-membership" });
}

function goStrategies() {
  router.push({ name: "h5-strategies" });
}

function goNews() {
  router.push({ name: "h5-news" });
}

function goWatchlistDetail() {
  router.push({ name: "h5-profile-watchlist" });
}

function scrollIntoSection(targetRef) {
  targetRef?.value?.scrollIntoView({ behavior: "smooth", block: "start" });
}

function handleTargetAction(target) {
  if (target === "watchlist") {
    goWatchlistDetail();
    return;
  }
  if (target === "community") {
    scrollIntoSection(moduleSectionRef);
    return;
  }
  if (target === "membership") {
    goMembership();
    return;
  }
  if (target === "strategies") {
    goStrategies();
    return;
  }
  if (target === "news") {
    goNews();
    return;
  }
  if (target === "invite") {
    scrollIntoSection(inviteSectionRef);
    return;
  }
  if (target === "messages") {
    scrollIntoSection(messageSectionRef);
    return;
  }
  loadProfilePage();
}

function shortcutGlyph(id) {
  if (id === "watchlist") return "关";
  if (id === "membership") return "会";
  if (id === "strategies") return "策";
  if (id === "news") return "讯";
  if (id === "invite") return "邀";
  return "我";
}

function scrollIntoProfileSection(section) {
  if (section === "watchlist") {
    scrollIntoSection(watchlistEntryRef);
    return;
  }
  if (section === "community") {
    scrollIntoSection(moduleSectionRef);
  }
}

watch(
  () => route.query.section,
  async (value) => {
    const section = normalizeProfileModuleSection(value);
    if (section === "overview") {
      return;
    }
    await nextTick();
    scrollIntoProfileSection(section);
  },
  { immediate: true }
);

onMounted(async () => {
  loadProfilePage();
  if (activeProfileSection.value !== "overview") {
    await nextTick();
    scrollIntoProfileSection(activeProfileSection.value);
  }
});
</script>

<style scoped>
.profile-page {
  padding-bottom: 18px;
}

.profile-hero,
.profile-card {
  position: relative;
  overflow: hidden;
  border: 1px solid var(--h5-line);
  border-radius: var(--h5-radius);
  box-shadow: var(--h5-shadow);
}

.profile-hero {
  padding: 20px 18px;
  background:
    radial-gradient(circle at top right, rgba(184, 137, 61, 0.2), transparent 30%),
    linear-gradient(160deg, #173a6e 0%, #224d83 54%, #f5f8fc 54%, #ffffff 100%);
  display: grid;
  gap: 16px;
}

.profile-hero-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.profile-hero-user {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  min-width: 0;
}

.profile-avatar {
  width: 50px;
  height: 50px;
  border-radius: 18px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.28), rgba(255, 255, 255, 0.08));
  border: 1px solid rgba(255, 255, 255, 0.26);
  color: #fff;
  font-size: 18px;
  font-weight: 800;
  letter-spacing: 0.06em;
}

.profile-hero-copy {
  display: grid;
  gap: 6px;
  min-width: 0;
}

.profile-hero-kicker {
  color: rgba(255, 255, 255, 0.72);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.profile-hero-copy h1,
.profile-hero-copy p,
.profile-section-head strong,
.profile-section-head span,
.profile-todo-copy p,
.profile-todo-copy strong,
.profile-metric-item strong,
.profile-metric-item p,
.profile-service-copy strong,
.profile-service-copy p,
.profile-message-item p,
.profile-shortcut strong,
.profile-shortcut small {
  margin: 0;
}

.profile-hero-copy h1 {
  color: #fff;
  font-size: clamp(24px, 6.8vw, 30px);
  line-height: 1.2;
  letter-spacing: -0.02em;
}

.profile-hero-copy p {
  color: rgba(255, 255, 255, 0.82);
  font-size: 13px;
  line-height: 1.7;
}

.profile-refresh {
  min-height: 34px;
  padding: 0 12px;
  border-radius: 999px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}

.profile-status-strip {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.profile-status-pill {
  min-height: 30px;
  padding: 0 12px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  font-weight: 700;
}

.profile-status-pill.brand {
  background: rgba(255, 255, 255, 0.15);
  color: #fff;
}

.profile-status-pill.gold {
  background: rgba(184, 137, 61, 0.18);
  color: #ffebc3;
}

.profile-status-pill.soft {
  background: rgba(255, 255, 255, 0.82);
  color: #173a6e;
}

.profile-metric-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.profile-metric-item {
  padding: 14px 12px;
  border-radius: 18px;
  background: rgba(255, 255, 255, 0.9);
  display: grid;
  gap: 5px;
}

.profile-metric-item span {
  color: #62728b;
  font-size: 11px;
}

.profile-metric-item strong {
  color: #16263d;
  font-size: 18px;
  line-height: 1.2;
}

.profile-metric-item p {
  color: #6c7c92;
  font-size: 11px;
  line-height: 1.5;
}

.profile-hero-actions {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.profile-card {
  padding: 18px 16px;
  background: rgba(255, 255, 255, 0.95);
  display: grid;
  gap: 16px;
}

.profile-card-accent {
  background:
    linear-gradient(180deg, rgba(255, 250, 243, 0.96), rgba(255, 255, 255, 0.98)),
    radial-gradient(circle at top right, rgba(184, 137, 61, 0.14), transparent 30%);
}

.profile-section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.profile-section-head.compact {
  align-items: flex-start;
}

.profile-section-head > div {
  display: grid;
  gap: 6px;
}

.profile-section-head strong {
  color: #16263d;
  font-size: 18px;
  line-height: 1.3;
}

.profile-section-head span {
  color: #68788f;
  font-size: 12px;
  line-height: 1.6;
}

.profile-todo-list,
.profile-service-list,
.profile-message-list,
.profile-invite-list,
.profile-share-list {
  display: grid;
  gap: 10px;
}

.profile-todo-item,
.profile-service-card,
.profile-message-item,
.profile-invite-item,
.profile-share-item,
.profile-invite-overview {
  border: 1px solid rgba(16, 42, 86, 0.07);
  border-radius: var(--h5-radius-sm);
  background: rgba(245, 248, 252, 0.82);
}

.profile-todo-item {
  padding: 14px;
  display: grid;
  gap: 12px;
}

.profile-todo-topline,
.profile-message-topline {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.profile-todo-copy {
  display: grid;
  gap: 8px;
}

.profile-share-item strong {
  color: #16263d;
  font-size: 15px;
  line-height: 1.45;
}

.profile-todo-copy p,
.profile-inline-note {
  color: #5f7088;
  font-size: 13px;
  line-height: 1.7;
}

.profile-shortcut-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
}

.profile-shortcut {
  padding: 14px 8px 12px;
  border: 1px solid rgba(16, 42, 86, 0.07);
  border-radius: var(--h5-radius-md);
  background: linear-gradient(180deg, rgba(247, 249, 252, 0.96), rgba(255, 255, 255, 1));
  display: grid;
  justify-items: center;
  gap: 8px;
}

.profile-shortcut-icon {
  width: 38px;
  height: 38px;
  border-radius: 14px;
  display: grid;
  place-items: center;
  background: rgba(24, 58, 110, 0.08);
  color: #173a6e;
  font-size: 16px;
  font-weight: 800;
}

.profile-shortcut strong {
  color: #16263d;
  font-size: 13px;
  line-height: 1.3;
}

.profile-shortcut small {
  color: #77869a;
  font-size: 11px;
  line-height: 1.4;
  text-align: center;
}

.profile-module-grid {
  display: grid;
  gap: 10px;
}

.profile-module-card {
  padding: 14px;
  border: 1px solid rgba(16, 42, 86, 0.08);
  border-radius: var(--h5-radius-md);
  background: linear-gradient(180deg, rgba(247, 249, 252, 0.96), rgba(255, 255, 255, 1));
  display: grid;
  gap: 8px;
  text-align: left;
}

.profile-module-card span {
  color: #7a8798;
  font-size: 11px;
  line-height: 1.4;
}

.profile-module-card strong {
  color: #16263d;
  font-size: 15px;
  line-height: 1.45;
}

.profile-module-card p {
  color: #5f7088;
  font-size: 13px;
  line-height: 1.7;
}

.profile-module-card.active {
  border-color: rgba(23, 58, 110, 0.18);
  background:
    radial-gradient(circle at top right, rgba(184, 137, 61, 0.16), transparent 35%),
    linear-gradient(180deg, rgba(23, 58, 110, 0.08), rgba(255, 255, 255, 0.98));
  box-shadow: inset 0 0 0 1px rgba(23, 58, 110, 0.08);
}

.profile-kyc-hero {
  padding: 14px;
  border-radius: var(--h5-radius-sm);
  background: rgba(255, 255, 255, 0.82);
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.profile-kyc-form {
  display: grid;
  gap: 10px;
}

.profile-kyc-form input {
  min-height: 48px;
  padding: 0 14px;
  border: 1px solid rgba(18, 52, 95, 0.12);
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.96);
}

.profile-service-card {
  padding: 14px;
  display: grid;
  gap: 12px;
  text-align: left;
}

.profile-service-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.profile-message-item,
.profile-invite-item,
.profile-share-item {
  padding: 14px;
  display: grid;
  gap: 10px;
}

.profile-message-topline > div {
  display: grid;
  gap: 4px;
}

.profile-message-topline span {
  color: #7a8798;
  font-size: 11px;
}

.profile-invite-overview {
  padding: 16px;
  display: grid;
  gap: 14px;
  background: linear-gradient(180deg, rgba(23, 58, 110, 0.95), rgba(34, 77, 131, 0.95));
}

.profile-invite-code span,
.profile-invite-actions small {
  color: rgba(255, 255, 255, 0.72);
  font-size: 11px;
}

.profile-invite-code strong,
.profile-invite-code p {
  color: #fff;
}

.profile-invite-actions {
  display: grid;
  gap: 8px;
}

@media (min-width: 521px) {
  .profile-page {
    gap: 14px;
  }

  .profile-shortcut-grid {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .profile-module-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .profile-service-list,
  .profile-message-list,
  .profile-share-list,
  .profile-invite-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 380px) {
  .profile-metric-grid,
  .profile-shortcut-grid,
  .profile-hero-actions {
    grid-template-columns: 1fr;
  }
}
</style>
