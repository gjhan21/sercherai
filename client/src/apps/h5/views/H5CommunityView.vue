<template>
  <div class="h5-page fade-up community-page">
    <div class="h5-page-topline">
      <span class="h5-page-tagline">讨论广场</span>
      <span>{{ activeMineLabel }}</span>
    </div>

    <H5HeroCard
      eyebrow="社区主入口"
      title="围绕资讯、策略和标的发起结构化讨论"
      description="从策略、资讯和我的讨论继续承接，统一处理发现观点、跟进讨论和发布判断。"
      tone="accent"
    >
      <div class="h5-chip-list">
        <span>股票</span>
        <span>期货</span>
        <span>资讯</span>
        <span>策略</span>
      </div>
      
      <template #actions>
        <H5ActionBar>
          <button type="button" class="h5-btn" @click="handleCompose">发布我的观点</button>
          <button type="button" class="h5-btn-ghost" :disabled="loading" @click="loadCommunityData">
            {{ loading ? "刷新中..." : "刷新列表" }}
          </button>
        </H5ActionBar>
      </template>

      <div class="community-hero-metrics h5-grid-3">
        <H5SummaryCard
          label="当前列表"
          :value="`${currentTotal} 条`"
          note="按时间顺序排列"
          tone="gold"
        />
        <H5SummaryCard
          v-if="isLoggedIn"
          label="我的参与"
          value="已登录"
          note="支持完整互动"
          tone="brand"
        />
        <H5SummaryCard
          v-else
          label="游客模式"
          value="待登录"
          note="登录后可评论"
          tone="soft"
        />
      </div>
    </H5HeroCard>

    <H5SectionBlock eyebrow="视角切换" title="在全部广场与个人足迹之间快速跳转" tone="soft">
      <div class="community-view-tabs">
        <button
          v-for="item in viewTabs"
          :key="item.value"
          type="button"
          class="h5-toggle-pill"
          :class="{ active: activeMine === item.value }"
          @click="handleViewChange(item.value)"
        >
          {{ item.label }}
        </button>
      </div>
    </H5SectionBlock>

    <H5SectionBlock v-if="!isMyCommentsView" eyebrow="内容筛选" title="按类型和热度寻找讨论干货" tone="soft">
      <div class="community-filter-scroll">
        <button
          v-for="item in typeTabs"
          :key="item.value"
          type="button"
          class="h5-toggle-pill"
          :class="{ active: activeType === item.value }"
          @click="setActiveType(item.value)"
        >
          {{ item.label }}
        </button>
      </div>
      <div class="community-filter-row">
        <button
          v-for="item in sortTabs"
          :key="item.value"
          type="button"
          class="h5-toggle-pill"
          :class="{ active: activeSort === item.value }"
          @click="setActiveSort(item.value)"
        >
          {{ item.label }}
        </button>
      </div>
    </H5SectionBlock>

    <H5SectionBlock :eyebrow="focusEyebrow" :title="focusSubtitle">
      <div v-if="!isMyCommentsView" class="community-list">
        <article
          v-for="item in topicCards"
          :key="item.id"
          class="community-card topic-card"
          @click="openTopic(item.id)"
        >
          <div class="community-card-head">
            <div class="community-card-labels">
              <span class="h5-pill h5-pill-info">{{ item.typeLabel }}</span>
              <span class="h5-pill" :class="item.stanceClassH5">{{ item.stanceLabel }}</span>
            </div>
            <span class="community-card-time">{{ item.lastActiveLabel }}</span>
          </div>
          <h3 class="community-card-title">{{ item.title }}</h3>
          <p class="community-card-summary">{{ item.summary }}</p>
          <div class="community-card-foot">
            <span class="community-card-author">作者: {{ item.authorLabel }}</span>
            <div class="community-card-stats">
              <span>评 {{ item.commentCount }}</span>
              <span>赞 {{ item.likeCount }}</span>
            </div>
          </div>
        </article>
        
        <H5EmptyState
          v-if="!loading && !topicCards.length"
          title="暂无主题"
          description="尝试切换筛选条件，或者发布你的第一条观点。"
        />
      </div>

      <div v-else class="community-list">
        <article
          v-for="item in commentCards"
          :key="item.id"
          class="community-card comment-card"
          @click="openTopic(item.topicID, item.id)"
        >
          <div class="community-card-head">
            <span class="h5-pill h5-pill-info">我的评论</span>
            <span class="community-card-time">{{ item.createdAtLabel }}</span>
          </div>
          <p class="community-comment-content">{{ item.content }}</p>
          <div class="community-comment-context">
            <strong>原文讨论: {{ item.topicTitle }}</strong>
          </div>
          <div class="community-card-foot">
            <span class="community-card-status">{{ item.statusLabel }}</span>
            <div class="community-card-stats">
              <span>赞 {{ item.likeCount }}</span>
            </div>
          </div>
        </article>

        <H5EmptyState
          v-if="!loading && !commentCards.length"
          title="暂无评论"
          description="登录后可在这里查看你参与过的所有讨论。"
        />
      </div>

      <div v-if="loading" class="h5-loading-box">
        <span>同步数据中...</span>
      </div>
    </H5SectionBlock>

    <H5StickyCta
      v-if="!isLoggedIn"
      title="登录并参与讨论"
      description="登录后你可以发布主线观点、点赞和评论。"
      primary-label="立即登录"
      @primary="router.push({ name: 'h5-auth', query: { redirect: route.fullPath } })"
    />
    <H5StickyCta
      v-else
      title="发现并分享你的专业判断"
      description="你可以从资讯、策略页关联发起讨论，也可以直接在这里发帖。"
      primary-label="发起新讨论"
      @primary="handleCompose"
    />
  </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import H5HeroCard from "../components/H5HeroCard.vue";
import H5SectionBlock from "../components/H5SectionBlock.vue";
import H5ActionBar from "../components/H5ActionBar.vue";
import H5SummaryCard from "../components/H5SummaryCard.vue";
import H5EmptyState from "../components/H5EmptyState.vue";
import H5StickyCta from "../components/H5StickyCta.vue";
import {
  listCommunityTopics,
  listMyCommunityComments,
  listMyCommunityTopics
} from "../../../api/community";
import { useClientAuth } from "../../../shared/auth/client-auth";
import { buildCommunityComposeRouteFromQuery } from "../../../lib/community-entry-links";

const router = useRouter();
const route = useRoute();
const { isLoggedIn } = useClientAuth();

const loading = ref(false);
const topicTotal = ref(0);
const commentTotal = ref(0);
const topics = ref([]);
const comments = ref([]);

const viewTabs = [
  { value: "", label: "全部广场" },
  { value: "topics", label: "我的主题" },
  { value: "comments", label: "我的评论" }
];

const typeTabs = [
  { value: "", label: "全部" },
  { value: "STOCK", label: "股票" },
  { value: "FUTURES", label: "期货" },
  { value: "NEWS", label: "资讯" },
  { value: "STRATEGY", label: "策略" }
];

const sortTabs = [
  { value: "MOST_ACTIVE", label: "最活跃" },
  { value: "LATEST", label: "最新发布" }
];

const activeMine = computed(() => route.query.mine || "");
const activeType = computed(() => route.query.topic_type || "");
const activeSort = computed(() => route.query.sort || "MOST_ACTIVE");
const isMyCommentsView = computed(() => activeMine.value === "comments");
const isMyTopicsView = computed(() => activeMine.value === "topics");
const activeMineLabel = computed(() => viewTabs.find(t => t.value === activeMine.value)?.label || "讨论");
const currentTotal = computed(() => (isMyCommentsView.value ? commentTotal.value : topicTotal.value));

const focusEyebrow = computed(() => isMyCommentsView.value ? "我的评论记录" : isMyTopicsView.value ? "我的发帖记录" : "内容流");
const focusSubtitle = computed(() => isMyCommentsView.value ? "按时间回看我参与过的评论" : isMyTopicsView.value ? "按时间切回我发表过的主题" : "按类型和热度查看广场最新讨论");

const topicCards = computed(() => topics.value.map(item => ({
  id: item.id,
  title: item.title || "未命名主题",
  summary: item.summary || "暂无摘要",
  typeLabel: mapTopicType(item.topic_type),
  stanceLabel: mapStance(item.stance),
  stanceClassH5: stanceClassH5(item.stance),
  commentCount: item.comment_count || 0,
  likeCount: item.like_count || 0,
  authorLabel: maskUserID(item.user_id),
  lastActiveLabel: formatDateTime(item.last_active_at)
})));

const commentCards = computed(() => comments.value.map(item => ({
  id: item.id,
  topicID: item.topic_id,
  topicTitle: item.topic_title || `主题 ${item.topic_id || "-"}`,
  content: item.content || "暂无评论内容",
  statusLabel: mapStatus(item.status),
  likeCount: item.like_count || 0,
  createdAtLabel: formatDateTime(item.created_at)
})));

watch(
  [activeMine, activeType, activeSort, isLoggedIn],
  () => {
    loadCommunityData();
  },
  { immediate: true }
);

async function loadCommunityData() {
  loading.value = true;
  try {
    if (isMyCommentsView.value) {
      if (!isLoggedIn.value) {
        comments.value = [];
        commentTotal.value = 0;
        return;
      }
      const result = await listMyCommunityComments({ page: 1, page_size: 20 });
      comments.value = result?.items || [];
      commentTotal.value = Number(result?.total || 0);
    } else {
      const loader = isMyTopicsView.value ? listMyCommunityTopics : listCommunityTopics;
      const result = await loader({
        topic_type: activeType.value,
        sort: activeSort.value,
        page: 1,
        page_size: 20
      });
      topics.value = result?.items || [];
      topicTotal.value = Number(result?.total || 0);
    }
  } finally {
    loading.value = false;
  }
}

function handleViewChange(val) {
  router.replace({ query: { ...route.query, mine: val || undefined } });
}

function setActiveType(val) {
  router.replace({ query: { ...route.query, topic_type: val || undefined } });
}

function setActiveSort(val) {
  router.replace({ query: { ...route.query, sort: val || undefined } });
}

function handleCompose() {
  if (!isLoggedIn.value) {
    router.push({ path: "/auth", query: { redirect: route.fullPath } });
    return;
  }
  const target = buildCommunityComposeRouteFromQuery(route.query);
  router.push(target || { path: "/community/new" });
}

function openTopic(id, commentID) {
  router.push({ 
    path: `/community/topics/${id}`, 
    query: commentID ? { comment_id: commentID } : undefined 
  });
}

function mapTopicType(v) {
  const m = { STOCK: "股票", FUTURES: "期货", NEWS: "资讯", STRATEGY: "策略" };
  return m[v] || "讨论";
}

function mapStance(v) {
  const m = { BULLISH: "看多", BEARISH: "看空", WATCH: "观察" };
  return m[v] || "待定";
}

function stanceClassH5(v) {
  const m = { BULLISH: "h5-pill-success", BEARISH: "h5-pill-danger" };
  return m[v] || "h5-pill-warning";
}

function mapStatus(v) {
  const m = { PUBLISHED: "已公开", PENDING_REVIEW: "待审核", HIDDEN: "已隐藏" };
  return m[v] || "处理中";
}

function maskUserID(v) {
  const s = String(v || "").trim();
  return s.length > 6 ? `${s.slice(0, 3)}***${s.slice(-2)}` : s || "匿名用户";
}

function formatDateTime(v) {
  if (!v) return "-";
  const d = new Date(v);
  return `${d.getMonth() + 1}-${d.getDate()} ${d.getHours()}:${d.getMinutes().toString().padStart(2, '0')}`;
}
</script>

<style scoped>
.community-page {
  gap: 12px;
}

.h5-chip-list {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.h5-chip-list span {
  padding: 4px 10px;
  background: rgba(255, 255, 255, 0.15);
  border-radius: 99px;
  font-size: 11px;
  color: #fff;
}

.community-view-tabs,
.community-filter-scroll {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  scrollbar-width: none;
  padding-bottom: 4px;
}

.community-view-tabs::-webkit-scrollbar,
.community-filter-scroll::-webkit-scrollbar {
  display: none;
}

.community-filter-row {
  display: flex;
  gap: 10px;
  margin-top: 10px;
}

.community-list {
  display: grid;
  gap: 12px;
}

.community-card {
  padding: 16px;
  background: var(--h5-card-bg, #fff);
  border: 1px solid var(--h5-line, #eee);
  border-radius: 16px;
  display: grid;
  gap: 10px;
}

.community-card-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.community-card-labels {
  display: flex;
  gap: 6px;
}

.community-card-time {
  font-size: 11px;
  color: var(--h5-text-soft, #999);
}

.community-card-title {
  margin: 0;
  font-size: 16px;
  line-height: 1.4;
  color: var(--h5-text, #333);
}

.community-card-summary,
.community-comment-content {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--h5-text-sub, #666);
}

.community-comment-context {
  background: var(--h5-panel-bg, #f9f9f9);
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 12px;
}

.community-card-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px solid var(--h5-line-soft, #f0f0f0);
}

.community-card-author,
.community-card-status {
  font-size: 11px;
  color: var(--h5-text-soft, #999);
}

.community-card-stats {
  display: flex;
  gap: 10px;
  font-size: 11px;
  color: var(--h5-brand, #102a56);
  font-weight: 700;
}

.h5-loading-box {
  padding: 20px;
  text-align: center;
  color: var(--h5-text-soft, #999);
  font-size: 12px;
}
</style>
