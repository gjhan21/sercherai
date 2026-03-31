<template>
  <section class="community-page fade-up">
    <header class="community-hero card finance-section-card">
      <div class="finance-copy-stack">
        <p class="hero-kicker">社区主入口</p>
        <h1 class="section-title">围绕资讯、策略和标的发起结构化讨论</h1>
        <p class="section-subtitle">
          从策略、资讯和我的讨论继续承接，统一处理发现观点、跟进讨论和发布判断。
        </p>
        <div class="finance-chip-list">
          <span>股票</span>
          <span>期货</span>
          <span>资讯</span>
          <span>策略</span>
        </div>
        <div class="community-entry-grid">
          <RouterLink class="finance-list-card finance-list-card-interactive" to="/strategies">
            <p>从策略进入</p>
            <strong>围绕候选和解释继续讨论</strong>
            <span>适合接住“为什么选它”的后续判断。</span>
          </RouterLink>
          <RouterLink class="finance-list-card finance-list-card-interactive" to="/news">
            <p>从资讯进入</p>
            <strong>围绕新闻和研报继续发帖</strong>
            <span>适合从文章详情承接观点、风险和节奏。</span>
          </RouterLink>
          <RouterLink class="finance-list-card finance-list-card-interactive" :to="{ path: '/community', query: { mine: 'topics' } }">
            <p>从我的讨论进入</p>
            <strong>看我的主题与评论回访链</strong>
            <span>适合收盘后回看自己发过的观点和互动。</span>
          </RouterLink>
        </div>
      </div>

      <div class="community-hero-actions">
        <button type="button" class="finance-primary-btn" @click="handleCompose">
          发布我的观点
        </button>
        <button type="button" class="finance-ghost-btn" :disabled="loading" @click="loadCommunityData">
          {{ loading ? "刷新中..." : "刷新广场" }}
        </button>
      </div>
    </header>

    <StatePanel
      :tone="accessState.tone"
      :eyebrow="accessState.label"
      :title="accessState.title"
      :description="accessState.desc"
    >
      <template #actions>
        <RouterLink
          v-if="!isLoggedIn"
          class="primary"
          :to="{ path: '/auth', query: { redirect: '/community' } }"
        >
          登录后参与讨论
        </RouterLink>
        <button v-else type="button" class="finance-primary-btn" @click="handleCompose">
          发布我的观点
        </button>
        <RouterLink
          v-if="isLoggedIn"
          class="ghost"
          :to="{ path: '/community', query: { mine: 'topics' } }"
        >
          看我的主题
        </RouterLink>
        <RouterLink v-else class="ghost" to="/news">从资讯进入</RouterLink>
      </template>
    </StatePanel>

    <StatePanel
      v-if="entryContext.hasContext && !isMyCommentsView"
      compact
      tone="info"
      eyebrow="当前承接"
      :title="entryContextTitle"
      :description="entryContextDescription"
    >
      <template #actions>
        <RouterLink
          v-if="!isLoggedIn"
          class="primary"
          :to="{ path: '/auth', query: { redirect: composeRedirectPath } }"
        >
          登录后围绕当前对象发帖
        </RouterLink>
        <button v-else type="button" class="finance-primary-btn" @click="handleCompose">
          围绕当前对象发帖
        </button>
        <button type="button" class="ghost finance-ghost-btn" @click="clearEntryContext">
          回到默认广场
        </button>
      </template>
    </StatePanel>

    <section class="finance-dual-rail community-focus-layout">
      <article class="card finance-section-card community-focus-card">
        <header class="finance-section-head-grid">
          <div>
            <p class="hero-kicker">{{ focusEyebrow }}</p>
            <h2 class="section-title">{{ focusTitle }}</h2>
            <p class="section-subtitle">{{ focusSubtitle }}</p>
          </div>
          <div class="community-focus-meta finance-summary-pill">
            <p>{{ focusMetaLabel }}</p>
            <strong>{{ currentTotal }} 条</strong>
          </div>
        </header>

        <div v-if="focusTopic" class="community-hero-topic finance-card-pale">
          <div class="community-topic-head">
            <div class="community-topic-title-wrap">
              <p class="decision-tag">{{ focusTopic.typeLabel }} · {{ focusTopic.stanceLabel }}</p>
              <h3>{{ focusTopic.title }}</h3>
            </div>
            <div class="community-topic-pills">
              <span class="finance-pill finance-pill-roomy finance-pill-info">{{ focusTopic.linkLabel }}</span>
              <span class="finance-pill finance-pill-roomy" :class="focusTopic.stanceClass">{{ focusTopic.stanceLabel }}</span>
            </div>
          </div>
          <p class="community-topic-summary">{{ focusTopic.summary }}</p>
          <div class="community-topic-metrics">
            <article v-for="item in focusStats" :key="item.label" class="finance-list-card finance-list-card-panel">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>
          <div class="community-topic-actions">
            <RouterLink class="finance-primary-btn community-link-btn" :to="`/community/topics/${focusTopic.id}`">
              查看讨论详情
            </RouterLink>
            <RouterLink class="finance-ghost-btn community-link-btn" to="/strategies">
              回到策略页
            </RouterLink>
          </div>
        </div>

        <div v-else-if="focusComment" class="community-comment-focus finance-card-pale">
          <div class="community-comment-focus-head">
            <div class="community-topic-title-wrap">
              <p class="decision-tag">我的评论 · {{ focusComment.topicStatusLabel }}</p>
              <h3>{{ focusComment.topicTitle }}</h3>
            </div>
            <div class="community-topic-pills">
              <span class="finance-pill finance-pill-roomy finance-pill-info">{{ focusComment.linkLabel }}</span>
              <span class="finance-pill finance-pill-roomy finance-pill-neutral">{{ focusComment.createdAtLabel }}</span>
            </div>
          </div>
          <p class="community-topic-summary">{{ focusComment.content }}</p>
          <div class="community-topic-metrics">
            <article v-for="item in focusStats" :key="item.label" class="finance-list-card finance-list-card-panel">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>
          <div class="community-topic-actions">
            <button type="button" class="finance-primary-btn community-link-btn" @click="openTopic(focusComment.topicID, focusComment.id)">
              回到原讨论
            </button>
            <RouterLink class="finance-ghost-btn community-link-btn" :to="{ path: '/community', query: { mine: 'topics' } }">
              查看我的主题
            </RouterLink>
          </div>
        </div>

        <StatePanel
          v-else-if="!loading"
          compact
          tone="info"
          eyebrow="广场状态"
          :title="emptyState.title"
          :description="emptyState.description"
        />
      </article>

      <aside class="finance-stack-tight finance-sticky-side">
        <article class="card finance-section-card community-side-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">当前支持</h2>
              <p class="section-subtitle">只展示真实支持的讨论能力。</p>
            </div>
          </header>
          <div class="community-side-list">
            <article v-for="item in supportRows" :key="item.title" class="finance-list-card finance-list-card-panel">
              <p>{{ item.title }}</p>
              <strong>{{ item.summary }}</strong>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </article>

        <article class="card finance-section-card community-side-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">发帖约定</h2>
              <p class="section-subtitle">先讲观点，再讲依据和风险。</p>
            </div>
          </header>
          <div class="community-side-list">
            <article v-for="item in ruleRows" :key="item.title" class="finance-list-card finance-list-card-panel">
              <p>{{ item.title }}</p>
              <strong>{{ item.summary }}</strong>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </article>

        <article class="card finance-section-card community-side-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">站内承接</h2>
              <p class="section-subtitle">从哪些页面引流过来。</p>
            </div>
          </header>
          <div class="community-path-list">
            <RouterLink class="finance-list-card finance-list-card-interactive" to="/news">
              <p>资讯页</p>
              <strong>围绕新闻和研报继续讨论</strong>
              <span>适合从文章详情继续延伸观点。</span>
            </RouterLink>
            <RouterLink class="finance-list-card finance-list-card-interactive" to="/strategies">
              <p>策略页</p>
              <strong>围绕推荐标的补充判断</strong>
              <span>适合接住“为什么选它”的后续讨论。</span>
            </RouterLink>
            <RouterLink
              v-if="isLoggedIn"
              class="finance-list-card finance-list-card-interactive"
              :to="{ path: '/community', query: { mine: 'comments' } }"
            >
              <p>我的评论</p>
              <strong>回看自己参与过的讨论</strong>
              <span>适合收盘后集中检查已发表观点和后续互动。</span>
            </RouterLink>
          </div>
        </article>
      </aside>
    </section>

    <section class="card finance-section-card community-filter-card">
      <header class="section-head compact">
        <div>
          <h2 class="section-title">{{ filterTitle }}</h2>
          <p class="section-subtitle">{{ filterSubtitle }}</p>
        </div>
        <p class="community-filter-meta">
          <template v-if="errorMessage">{{ errorMessage }}</template>
          <template v-else-if="lastLoadedAt">更新时间：{{ lastLoadedAt }}</template>
          <template v-else>{{ filterMetaFallback }}</template>
        </p>
      </header>

      <div class="community-filter-group community-view-tabs">
        <button
          v-for="item in viewTabs"
          :key="item.value"
          type="button"
          class="finance-toggle-btn"
          :class="{ active: activeMine === item.value }"
          @click="handleViewChange(item.value)"
        >
          {{ item.label }}
        </button>
      </div>

      <div v-if="!isMyCommentsView" class="community-filter-row">
        <div class="community-filter-group">
          <button
            v-for="item in typeTabs"
            :key="item.value"
            type="button"
            class="finance-toggle-btn"
            :class="{ active: activeType === item.value }"
            @click="setActiveType(item.value)"
          >
            {{ item.label }}
          </button>
        </div>
        <div class="community-filter-group">
          <button
            v-for="item in sortTabs"
            :key="item.value"
            type="button"
            class="finance-toggle-btn"
            :class="{ active: activeSort === item.value }"
            @click="setActiveSort(item.value)"
          >
            {{ item.label }}
          </button>
        </div>
      </div>
    </section>

    <section v-if="!isMyCommentsView" class="community-topic-list">
      <article
        v-for="item in topicCards"
        :key="item.id"
        class="card finance-section-card community-topic-card"
        @click="openTopic(item.id)"
      >
        <div class="community-topic-head">
          <div class="community-topic-title-wrap">
            <div class="community-topic-labels">
              <span class="finance-pill finance-pill-compact finance-pill-info">{{ item.typeLabel }}</span>
              <span class="finance-pill finance-pill-compact" :class="item.stanceClass">{{ item.stanceLabel }}</span>
              <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ item.linkTypeLabel }}</span>
            </div>
            <h3>{{ item.title }}</h3>
          </div>
          <div class="community-topic-side">
            <p>{{ item.lastActiveLabel }}</p>
            <strong>{{ item.linkLabel }}</strong>
          </div>
        </div>

        <p class="community-topic-summary">{{ item.summary }}</p>

        <div class="community-topic-grid">
          <article class="finance-list-card finance-list-card-panel">
            <p>发帖人</p>
            <strong>{{ item.authorLabel }}</strong>
            <span>当前版本只展示 ID 匿名化。</span>
          </article>
          <article class="finance-list-card finance-list-card-panel">
            <p>评论数</p>
            <strong>{{ item.commentCount }}</strong>
            <span>先看有无继续跟帖价值。</span>
          </article>
          <article class="finance-list-card finance-list-card-panel">
            <p>点赞 / 收藏</p>
            <strong>{{ item.likeCount }} / {{ item.favoriteCount }}</strong>
            <span>帮助判断社区关注度。</span>
          </article>
          <article class="finance-list-card finance-list-card-panel">
            <p>当前状态</p>
            <strong>{{ item.statusLabel }}</strong>
            <span>{{ item.statusNote }}</span>
          </article>
        </div>
      </article>

      <StatePanel
        v-if="!loading && !topicCards.length"
        compact
        tone="info"
        eyebrow="筛选结果"
        :title="emptyState.title"
        :description="emptyState.description"
      />
    </section>

    <section v-else class="community-topic-list">
      <article
        v-for="item in commentCards"
        :key="item.id"
        class="card finance-section-card community-comment-card"
        @click="openTopic(item.topicID, item.id)"
      >
        <div class="community-topic-head">
          <div class="community-topic-title-wrap">
            <div class="community-topic-labels">
              <span class="finance-pill finance-pill-compact finance-pill-info">我的评论</span>
              <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ item.topicStatusLabel }}</span>
              <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ item.linkTypeLabel }}</span>
            </div>
            <h3>{{ item.topicTitle }}</h3>
          </div>
          <div class="community-topic-side">
            <p>{{ item.createdAtLabel }}</p>
            <strong>{{ item.linkLabel }}</strong>
          </div>
        </div>

        <p class="community-topic-summary">{{ item.content }}</p>

        <div class="community-topic-grid">
          <article class="finance-list-card finance-list-card-panel">
            <p>评论状态</p>
            <strong>{{ item.statusLabel }}</strong>
            <span>{{ item.statusNote }}</span>
          </article>
          <article class="finance-list-card finance-list-card-panel">
            <p>点赞数</p>
            <strong>{{ item.likeCount }}</strong>
            <span>用来判断这条评论是否继续跟进。</span>
          </article>
          <article class="finance-list-card finance-list-card-panel">
            <p>回复对象</p>
            <strong>{{ item.replyTarget }}</strong>
            <span>当前只支持单层回复链。</span>
          </article>
          <article class="finance-list-card finance-list-card-panel">
            <p>所属主题</p>
            <strong>{{ item.topicID }}</strong>
            <span>点击卡片可回到原始讨论继续查看。</span>
          </article>
        </div>
      </article>

      <StatePanel
        v-if="!loading && !commentCards.length"
        compact
        tone="info"
        eyebrow="我的评论"
        :title="emptyState.title"
        :description="emptyState.description"
      />
    </section>
  </section>
</template>

<script setup>
import { computed, ref, watch } from "vue";
import { RouterLink, useRoute, useRouter } from "vue-router";
import StatePanel from "../../../components/StatePanel.vue";
import {
  listCommunityTopics,
  listMyCommunityComments,
  listMyCommunityTopics
} from "../../../api/community";
import { useClientAuth } from "../../../lib/client-auth";
import {
  buildCommunityComposeRouteFromQuery,
  resolveCommunityEntryContext
} from "../../../lib/community-entry-links";
import { normalizeCommunityLoadError } from "../../../lib/community-error";

const router = useRouter();
const route = useRoute();
const { isLoggedIn } = useClientAuth();

const loading = ref(false);
const errorMessage = ref("");
const lastLoadedAt = ref("");
const topicTotal = ref(0);
const commentTotal = ref(0);
const topics = ref([]);
const comments = ref([]);

const viewTabs = [
  { value: "", label: "全部主题" },
  { value: "topics", label: "我的主题" },
  { value: "comments", label: "我的评论" }
];

const typeTabs = [
  { value: "", label: "全部主题" },
  { value: "STOCK", label: "股票" },
  { value: "FUTURES", label: "期货" },
  { value: "NEWS", label: "资讯" },
  { value: "STRATEGY", label: "策略" }
];

const sortTabs = [
  { value: "MOST_ACTIVE", label: "最活跃" },
  { value: "LATEST", label: "最新发布" }
];

const activeMine = computed(() => normalizeMine(route.query.mine));
const activeType = computed(() => normalizeTopicType(route.query.topic_type));
const activeSort = computed(() => normalizeSort(route.query.sort));
const isMyTopicsView = computed(() => activeMine.value === "topics");
const isMyCommentsView = computed(() => activeMine.value === "comments");
const currentTotal = computed(() => (isMyCommentsView.value ? commentTotal.value : topicTotal.value));
const entryContext = computed(() => resolveCommunityEntryContext(route.query));
const composeRoute = computed(() =>
  entryContext.value.hasContext ? buildCommunityComposeRouteFromQuery(route.query) : { path: "/community/new" }
);
const composeRedirectPath = computed(() => router.resolve(composeRoute.value).fullPath);

const accessState = computed(() =>
  isLoggedIn.value
    ? {
        tone: "info",
        label: "已登录",
        title: "你现在可以直接发起观点主题、点赞、收藏和评论。",
        desc: "写观点时请同时补充判断依据和风险边界，避免只留结论。"
      }
    : {
        tone: "warning",
        label: "游客模式",
        title: "当前可先查看主题摘要和评论结构，登录后再参与发言。",
        desc: "登录后会保留当前返回路径，不会强制跳回首页。"
      }
);

const supportRows = [
  {
    title: "当前主题类型",
    summary: "股票、期货、资讯、策略",
    desc: "只围绕现有站内真实内容建立讨论入口，不扩成泛社区。"
  },
  {
    title: "互动能力",
    summary: "主题帖、评论、点赞、收藏、举报",
    desc: "不做实时聊天室，不做广场热榜推荐流。"
  },
  {
    title: "回看方式",
    summary: "全部主题 / 我的主题 / 我的评论",
    desc: "同一套页面兼容广场浏览和个人回看，不拆独立 H5 子页。"
  }
];

const ruleRows = [
  {
    title: "先写判断",
    summary: "明确看多、看空还是观察",
    desc: "避免模糊表态，方便后续回看观点是否成立。"
  },
  {
    title: "补充依据",
    summary: "说清为什么这样看",
    desc: "可以结合研报、资讯、策略参数或标的背景来说明。"
  },
  {
    title: "写出风险",
    summary: "说明什么情况下失效",
    desc: "生产环境只保留有风险边界的观点讨论。"
  }
];

const entryContextTitle = computed(() => {
  const targetLabel = entryContext.value.targetSnapshot || entryContext.value.title;
  if (targetLabel) {
    return `当前从${mapEntrySource(entryContext.value.entrySource)}带着“${targetLabel}”进入`;
  }
  if (entryContext.value.topicType) {
    return `当前从${mapEntrySource(entryContext.value.entrySource)}进入${mapTopicType(entryContext.value.topicType)}讨论`;
  }
  return "当前从站内内容页进入讨论广场";
});

const entryContextDescription = computed(() => {
  const typeLabel = mapTopicType(entryContext.value.topicType);
  if (entryContext.value.targetSnapshot) {
    return `下方列表先按${typeLabel}主题展示真实内容；如果你要围绕“${entryContext.value.targetSnapshot}”补充个人判断，可直接发起主题。`;
  }
  return `下方列表先按${typeLabel}主题展示真实内容，不额外扩展虚构筛选能力。`;
});

const topicCards = computed(() =>
  topics.value.map((item) => {
    const typeLabel = mapTopicType(item.topic_type);
    const stanceLabel = mapStance(item.stance);
    const statusLabel = mapStatus(item.status);
    return {
      id: item.id,
      title: item.title || "未命名主题",
      summary: item.summary || "暂无摘要",
      typeLabel,
      stanceLabel,
      stanceClass: stanceClass(item.stance),
      commentCount: item.comment_count || 0,
      likeCount: item.like_count || 0,
      favoriteCount: item.favorite_count || 0,
      statusLabel,
      statusNote: statusLabel === "已发布" ? "公开可读，可继续讨论。" : "当前状态由管理员审核控制。",
      linkLabel: item.linked_target?.target_snapshot || item.linked_target?.target_id || "未关联对象",
      linkTypeLabel: mapLinkType(item.linked_target?.target_type),
      authorLabel: maskUserID(item.user_id),
      lastActiveLabel: formatDateTime(item.last_active_at)
    };
  })
);

const commentCards = computed(() =>
  comments.value.map((item) => {
    const statusLabel = mapStatus(item.status);
    const topicStatusLabel = mapStatus(item.topic_status);
    return {
      id: item.id,
      topicID: item.topic_id,
      topicTitle: item.topic_title || `主题 ${item.topic_id || "-"}`,
      topicStatusLabel,
      content: item.content || "暂无评论内容",
      statusLabel,
      statusNote: statusLabel === "已发布" ? "已对外展示，可继续收到互动。" : "评论状态会受审核结果影响。",
      likeCount: item.like_count || 0,
      createdAtLabel: formatDateTime(item.created_at),
      replyTarget: maskUserID(item.reply_to_user_id || item.parent_comment_id || "主题楼主"),
      linkLabel: item.linked_target?.target_snapshot || item.linked_target?.target_id || "未关联对象",
      linkTypeLabel: mapLinkType(item.linked_target?.target_type),
      topicStatus: item.topic_status || "",
      topicTitleRaw: item.topic_title || ""
    };
  })
);

const focusTopic = computed(() => (isMyCommentsView.value ? null : topicCards.value[0] || null));
const focusComment = computed(() => (isMyCommentsView.value ? commentCards.value[0] || null : null));

const focusEyebrow = computed(() => {
  if (isMyCommentsView.value) {
    return "我的评论回看";
  }
  if (isMyTopicsView.value) {
    return "我的主题回看";
  }
  return "今日广场焦点";
});

const focusTitle = computed(() => {
  if (focusTopic.value) {
    return focusTopic.value.title;
  }
  if (focusComment.value) {
    return focusComment.value.topicTitle;
  }
  if (isMyCommentsView.value) {
    return "优先回看你最近参与过的评论";
  }
  if (isMyTopicsView.value) {
    return "优先回看你发出的主题帖";
  }
  return "优先看当前最活跃的主题帖";
});

const focusSubtitle = computed(() => {
  if (isMyCommentsView.value) {
    return "先看评论内容、所属主题和当前状态，再决定是否回到原帖继续补充。";
  }
  if (isMyTopicsView.value) {
    return "先看你自己的观点沉淀、互动反馈和状态变化，再决定是否继续更新。";
  }
  return "先看观点摘要、关联对象和风险说明，再决定是否进入详情展开讨论。";
});

const focusMetaLabel = computed(() => {
  if (isMyCommentsView.value) {
    return "我的评论";
  }
  if (isMyTopicsView.value) {
    return "我的主题";
  }
  return "当前列表";
});

const focusStats = computed(() => {
  if (focusTopic.value) {
    return [
      {
        label: isMyTopicsView.value ? "最近活跃" : "最近活跃",
        value: focusTopic.value.lastActiveLabel,
        note: isMyTopicsView.value ? "适合优先回看这条主题是否需要补充。" : "适合优先查看讨论是否有新增补充。"
      },
      {
        label: "评论数",
        value: `${focusTopic.value.commentCount} 条`,
        note: "帮助判断主题是否已有充分跟帖。"
      },
      {
        label: "点赞 / 收藏",
        value: `${focusTopic.value.likeCount} / ${focusTopic.value.favoriteCount}`,
        note: "当前只提供基础互动，不做额外权重热榜。"
      }
    ];
  }
  if (focusComment.value) {
    return [
      {
        label: "评论时间",
        value: focusComment.value.createdAtLabel,
        note: "先确认这条评论是否还需要跟进。"
      },
      {
        label: "点赞数",
        value: `${focusComment.value.likeCount}`,
        note: "帮助判断这条评论是否获得持续关注。"
      },
      {
        label: "主题状态",
        value: focusComment.value.topicStatusLabel,
        note: "如果主题状态变化，回看优先级要跟着调整。"
      }
    ];
  }
  return [];
});

const filterTitle = computed(() => {
  if (isMyCommentsView.value) {
    return "我的评论回看";
  }
  if (isMyTopicsView.value) {
    return "我的主题筛选";
  }
  return "主题筛选";
});

const filterSubtitle = computed(() => {
  if (isMyCommentsView.value) {
    return "按评论时间回看自己参与过的讨论。";
  }
  if (isMyTopicsView.value) {
    return "按主题类型和排序回看自己的发帖记录。";
  }
  return "当前只支持按主题类型和排序查看。";
});

const filterMetaFallback = computed(() => {
  if (isMyCommentsView.value) {
    return "登录后可集中回看自己写过的评论。";
  }
  if (isMyTopicsView.value) {
    return "登录后可查看自己发表过的主题帖。";
  }
  return "首次进入将自动拉取主题列表。";
});

const emptyState = computed(() => {
  if (isMyCommentsView.value) {
    return {
      title: isLoggedIn.value ? "你还没有发表过评论" : "登录后查看我的评论",
      description: isLoggedIn.value
        ? "可以先从资讯页或策略页进入讨论详情，再留下自己的判断。"
        : "登录后可集中回看自己参与过的讨论。"
    };
  }
  if (isMyTopicsView.value) {
    return {
      title: isLoggedIn.value ? "你还没有发布过主题" : "登录后查看我的主题",
      description: isLoggedIn.value
        ? "可以先从资讯页、策略页或广场入口发起第一条观点主题。"
        : "登录后可查看自己发布过的主题和状态变化。"
    };
  }
  return {
    title: "当前筛选条件下暂无主题",
    description: "可以切换类型或排序重新查看。"
  };
});

watch(
  [activeMine, activeType, activeSort, isLoggedIn],
  () => {
    loadCommunityData();
  },
  { immediate: true }
);

async function loadCommunityData() {
  loading.value = true;
  errorMessage.value = "";
  try {
    if (isMyCommentsView.value) {
      if (!isLoggedIn.value) {
        comments.value = [];
        topics.value = [];
        topicTotal.value = 0;
        commentTotal.value = 0;
        return;
      }
      const result = await listMyCommunityComments({
        page: 1,
        page_size: 20
      });
      comments.value = Array.isArray(result?.items) ? result.items : [];
      commentTotal.value = Number(result?.total || 0);
      topics.value = [];
      topicTotal.value = 0;
    } else {
      const loader = isMyTopicsView.value ? listMyCommunityTopics : listCommunityTopics;
      const result = await loader({
        topic_type: activeType.value,
        sort: activeSort.value,
        page: 1,
        page_size: 20
      });
      topics.value = Array.isArray(result?.items) ? result.items : [];
      topicTotal.value = Number(result?.total || 0);
      comments.value = [];
      commentTotal.value = 0;
    }
    lastLoadedAt.value = formatDateTime(new Date().toISOString());
  } catch (error) {
    errorMessage.value = normalizeCommunityLoadError(error?.message);
    topics.value = [];
    comments.value = [];
    topicTotal.value = 0;
    commentTotal.value = 0;
  } finally {
    loading.value = false;
  }
}

function handleCompose() {
  const targetRoute = composeRoute.value;
  if (!isLoggedIn.value) {
    router.push({ path: "/auth", query: { redirect: composeRedirectPath.value } });
    return;
  }
  router.push(targetRoute);
}

function handleViewChange(value) {
  const targetMine = normalizeMine(value);
  if (targetMine && !isLoggedIn.value) {
    router.push({
      path: "/auth",
      query: { redirect: targetMine ? `/community?mine=${targetMine}` : "/community" }
    });
    return;
  }
  const query = { ...route.query };
  if (targetMine) {
    query.mine = targetMine;
  } else {
    delete query.mine;
  }
  router.replace({ path: "/community", query });
}

function clearEntryContext() {
  const query = {};
  if (activeMine.value) {
    query.mine = activeMine.value;
  }
  router.replace({ path: "/community", query });
}

function setActiveType(value) {
  replaceCommunityQuery({
    topic_type: normalizeTopicType(value)
  });
}

function setActiveSort(value) {
  replaceCommunityQuery({
    sort: normalizeSort(value)
  });
}

function openTopic(topicID, commentID = "") {
  const path = `/community/topics/${topicID}`;
  if (commentID) {
    router.push(`${path}#comment-${commentID}`);
    return;
  }
  router.push(path);
}

function normalizeMine(value) {
  const normalized = String(value || "").trim().toLowerCase();
  if (normalized === "topics" || normalized === "comments") {
    return normalized;
  }
  return "";
}

function normalizeTopicType(value) {
  const normalized = String(value || "").trim().toUpperCase();
  if (["STOCK", "FUTURES", "NEWS", "STRATEGY"].includes(normalized)) {
    return normalized;
  }
  return "";
}

function normalizeSort(value) {
  const normalized = String(value || "").trim().toUpperCase();
  if (normalized === "LATEST") {
    return "LATEST";
  }
  return "MOST_ACTIVE";
}

function replaceCommunityQuery(patch = {}) {
  const query = { ...route.query, ...patch };
  Object.keys(query).forEach((key) => {
    if (query[key] === undefined || query[key] === null || query[key] === "") {
      delete query[key];
    }
  });
  router.replace({ path: "/community", query });
}

function mapTopicType(value) {
  switch (String(value || "").toUpperCase()) {
    case "STOCK":
      return "股票";
    case "FUTURES":
      return "期货";
    case "NEWS":
      return "资讯";
    case "STRATEGY":
      return "策略";
    default:
      return "讨论";
  }
}

function mapEntrySource(value) {
  switch (String(value || "").trim().toLowerCase()) {
    case "news_detail":
      return "资讯页";
    case "home_stock":
      return "首页主推荐";
    case "home_research":
      return "首页研报导读";
    case "strategy_stock":
      return "策略页股票区";
    case "strategy_futures":
      return "策略页期货区";
    default:
      return "站内内容页";
  }
}

function mapLinkType(value) {
  switch (String(value || "").toUpperCase()) {
    case "STOCK":
      return "关联股票";
    case "FUTURES":
      return "关联期货";
    case "NEWS_ARTICLE":
      return "关联资讯";
    case "STRATEGY_ITEM":
      return "关联策略";
    default:
      return "关联对象";
  }
}

function mapStance(value) {
  switch (String(value || "").toUpperCase()) {
    case "BULLISH":
      return "看多";
    case "BEARISH":
      return "看空";
    case "WATCH":
      return "观察";
    default:
      return "待定";
  }
}

function stanceClass(value) {
  switch (String(value || "").toUpperCase()) {
    case "BULLISH":
      return "finance-pill-success";
    case "BEARISH":
      return "finance-pill-danger";
    default:
      return "finance-pill-warning";
  }
}

function mapStatus(value) {
  switch (String(value || "").toUpperCase()) {
    case "PUBLISHED":
      return "已发布";
    case "PENDING_REVIEW":
      return "待审核";
    case "HIDDEN":
      return "已隐藏";
    case "DELETED":
      return "已删除";
    default:
      return "处理中";
  }
}

function maskUserID(value) {
  const source = String(value || "").trim();
  if (!source) {
    return "匿名用户";
  }
  if (source.length <= 6) {
    return source;
  }
  return `${source.slice(0, 3)}***${source.slice(-2)}`;
}

function formatDateTime(value) {
  const date = value ? new Date(value) : null;
  if (!date || Number.isNaN(date.getTime())) {
    return "-";
  }
  return new Intl.DateTimeFormat("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit"
  }).format(date);
}
</script>

<style scoped>
.community-page {
  display: grid;
  gap: 16px;
}

.community-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 16px;
  align-items: start;
}

.community-hero-actions {
  display: grid;
  gap: 10px;
  min-width: 150px;
}

.community-focus-layout {
  --finance-main-column: minmax(0, 1.58fr);
  --finance-side-column: minmax(280px, 0.88fr);
}

.community-focus-card,
.community-side-card,
.community-filter-card,
.community-topic-card,
.community-comment-card {
  display: grid;
  gap: 14px;
}

.community-hero-topic,
.community-comment-focus {
  padding: 16px;
  display: grid;
  gap: 14px;
}

.community-topic-head,
.community-comment-focus-head {
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: start;
}

.community-topic-title-wrap,
.community-topic-side {
  min-width: 0;
  display: grid;
  gap: 6px;
}

.community-topic-title-wrap h3,
.community-topic-title-wrap h2 {
  margin: 0;
}

.community-topic-pills,
.community-topic-labels {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.community-topic-summary {
  margin: 0;
  line-height: 1.72;
  color: var(--color-text-sub);
}

.community-topic-metrics,
.community-topic-grid,
.community-side-list,
.community-path-list,
.community-entry-grid {
  display: grid;
  gap: 10px;
}

.community-topic-metrics,
.community-topic-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.community-topic-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.community-entry-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.community-link-btn {
  text-decoration: none;
}

.community-view-tabs {
  margin-bottom: 6px;
}

.community-filter-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: space-between;
  gap: 12px;
}

.community-filter-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.community-filter-meta {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.community-topic-list {
  display: grid;
  gap: 12px;
}

.community-topic-card,
.community-comment-card {
  cursor: pointer;
  transition:
    transform 0.2s ease,
    box-shadow 0.2s ease,
    border-color 0.2s ease;
}

.community-topic-card:hover,
.community-comment-card:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-soft);
  border-color: var(--color-border-focus-soft);
}

.community-topic-side {
  justify-items: end;
  text-align: right;
}

.community-topic-side p,
.community-topic-side strong {
  margin: 0;
}

.community-focus-meta strong {
  white-space: nowrap;
}

</style>
