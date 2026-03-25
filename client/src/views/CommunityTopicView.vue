<template>
  <section class="community-topic-page fade-up">
    <header class="card finance-section-card topic-hero">
      <div class="finance-copy-stack">
        <p class="hero-kicker">讨论详情</p>
        <h1 class="section-title">{{ detailCard.title }}</h1>
        <p class="section-subtitle">{{ detailCard.summary }}</p>
        <div class="topic-hero-tags">
          <span class="finance-pill finance-pill-compact finance-pill-info">{{ detailCard.typeLabel }}</span>
          <span class="finance-pill finance-pill-compact" :class="detailCard.stanceClass">{{ detailCard.stanceLabel }}</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ detailCard.linkTypeLabel }}</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ detailCard.statusLabel }}</span>
        </div>
      </div>

      <div class="topic-hero-actions">
        <button type="button" class="finance-ghost-btn" @click="router.push('/community')">返回广场</button>
        <button type="button" class="finance-primary-btn" :disabled="detailLoading" @click="loadTopicPage">
          {{ detailLoading ? "刷新中..." : "刷新详情" }}
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
        <button type="button" class="finance-primary-btn" @click="handleCommentFocus">
          {{ isLoggedIn ? "参与评论" : "登录后评论" }}
        </button>
        <button
          type="button"
          class="ghost finance-ghost-btn"
          :disabled="reactionLoading.like || !detail.id"
          @click="toggleTopicReaction('LIKE')"
        >
          {{ detail.liked_by_me ? "取消点赞" : "点赞主题" }}
        </button>
      </template>
    </StatePanel>

    <section class="finance-dual-rail topic-detail-layout">
      <div class="finance-stack-tight">
        <article class="card finance-section-card topic-core-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">主题观点</h2>
              <p class="section-subtitle">先看结论，再看依据和风险边界。</p>
            </div>
            <div class="topic-stat-strip">
              <span class="finance-pill finance-pill-roomy finance-pill-info">评论 {{ detail.comment_count || 0 }}</span>
              <span class="finance-pill finance-pill-roomy finance-pill-info">点赞 {{ detail.like_count || 0 }}</span>
              <span class="finance-pill finance-pill-roomy finance-pill-info">收藏 {{ detail.favorite_count || 0 }}</span>
            </div>
          </header>

          <div class="topic-core-grid">
            <article class="finance-list-card finance-list-card-panel">
              <p>观点方向</p>
              <strong>{{ detailCard.stanceLabel }}</strong>
              <span>{{ detailCard.typeLabel }}主题 · {{ detailCard.horizonLabel }}</span>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>关联对象</p>
              <strong>{{ detailCard.linkLabel }}</strong>
              <span>{{ detailCard.linkTypeLabel }}</span>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>最近活跃</p>
              <strong>{{ detailCard.lastActiveLabel }}</strong>
              <span>发帖人 {{ detailCard.authorLabel }}</span>
            </article>
          </div>

          <div class="topic-reading-stack">
            <section class="finance-card-pale topic-reading-card">
              <div class="topic-reading-head">
                <p>为什么这样看</p>
                <span>{{ detailCard.typeLabel }}判断</span>
              </div>
              <strong>{{ detail.reason_text || "暂无补充" }}</strong>
              <p>{{ detail.summary || "暂无摘要" }}</p>
            </section>

            <section class="finance-card-pale topic-reading-card">
              <div class="topic-reading-head">
                <p>风险边界</p>
                <span>失效条件</span>
              </div>
              <strong>{{ detail.risk_text || "暂无补充" }}</strong>
              <p>讨论页只保留“观点 + 依据 + 风险”三段式结构，不支持只给结论。</p>
            </section>

            <section class="finance-card-soft topic-content-card">
              <div class="topic-reading-head">
                <p>完整正文</p>
                <span>{{ detailCard.createdLabel }}</span>
              </div>
              <p class="topic-content-text">{{ detail.content || "暂无正文" }}</p>
            </section>
          </div>

          <div class="topic-action-row">
            <button
              type="button"
              class="finance-primary-btn"
              :disabled="reactionLoading.like || !detail.id"
              @click="toggleTopicReaction('LIKE')"
            >
              {{ detail.liked_by_me ? "取消点赞" : "点赞主题" }}
            </button>
            <button
              type="button"
              class="finance-ghost-btn"
              :disabled="reactionLoading.favorite || !detail.id"
              @click="toggleTopicReaction('FAVORITE')"
            >
              {{ detail.favorited_by_me ? "取消收藏" : "收藏主题" }}
            </button>
          </div>
        </article>

        <article class="card finance-section-card topic-comment-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">评论区</h2>
              <p class="section-subtitle">当前支持一层回复，适合围绕主题继续补充证据和风险提示。</p>
            </div>
            <div class="topic-comment-meta">
              <span>{{ commentsTotal }} 条评论</span>
              <span v-if="commentErrorMessage">{{ commentErrorMessage }}</span>
            </div>
          </header>

          <div ref="commentComposerRef" class="topic-comment-composer finance-card-pale">
            <div class="topic-comment-head">
              <div>
                <p>写评论</p>
                <strong>{{ replyTarget ? `回复 ${replyTarget.authorLabel}` : "补充你的判断、证据和风险" }}</strong>
              </div>
              <button
                v-if="replyTarget"
                type="button"
                class="finance-mini-btn finance-mini-btn-soft"
                @click="clearReplyTarget"
              >
                取消回复
              </button>
            </div>
            <textarea
              v-model.trim="commentDraft"
              :disabled="commentSubmitting"
              maxlength="600"
              placeholder="请写明你为什么这样判断，以及什么情况下你的观点会失效。"
            />
            <div class="topic-comment-actions">
              <button type="button" class="finance-primary-btn" :disabled="commentSubmitting" @click="submitComment">
                {{ commentSubmitting ? "提交中..." : isLoggedIn ? "提交评论" : "登录后评论" }}
              </button>
              <p>{{ commentDraft.length }}/600</p>
            </div>
          </div>

          <div v-if="commentRows.length" class="topic-comment-list">
            <article v-for="item in commentRows" :key="item.id" class="finance-list-card topic-comment-item">
              <div class="topic-comment-top">
                <div>
                  <p>{{ item.authorLabel }}</p>
                  <strong>{{ item.createdLabel }}</strong>
                </div>
                <div class="topic-comment-side">
                  <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ item.statusLabel }}</span>
                  <span class="finance-pill finance-pill-compact finance-pill-info">赞 {{ item.likeCount }}</span>
                </div>
              </div>
              <p class="topic-comment-content">{{ item.content }}</p>
              <div class="topic-comment-actions">
                <button
                  type="button"
                  class="finance-mini-btn finance-mini-btn-soft"
                  :disabled="reactionLoading[`comment-${item.id}`]"
                  @click="toggleCommentLike(item)"
                >
                  {{ item.likedByMe ? "取消点赞" : "点赞" }}
                </button>
                <button type="button" class="finance-mini-btn finance-mini-btn-soft" @click="setReplyTarget(item)">
                  回复
                </button>
              </div>

              <div v-if="item.replies.length" class="topic-reply-list">
                <article v-for="reply in item.replies" :key="reply.id" class="finance-list-card finance-list-card-panel topic-reply-item">
                  <div class="topic-comment-top">
                    <div>
                      <p>{{ reply.authorLabel }}</p>
                      <strong>{{ reply.createdLabel }}</strong>
                    </div>
                    <div class="topic-comment-side">
                      <span class="finance-pill finance-pill-compact finance-pill-info">赞 {{ reply.likeCount }}</span>
                    </div>
                  </div>
                  <p class="topic-comment-content">
                    <template v-if="reply.replyToLabel">回复 {{ reply.replyToLabel }}：</template>{{ reply.content }}
                  </p>
                  <div class="topic-comment-actions">
                    <button
                      type="button"
                      class="finance-mini-btn finance-mini-btn-soft"
                      :disabled="reactionLoading[`comment-${reply.id}`]"
                      @click="toggleCommentLike(reply)"
                    >
                      {{ reply.likedByMe ? "取消点赞" : "点赞" }}
                    </button>
                    <button type="button" class="finance-mini-btn finance-mini-btn-soft" @click="setReplyTarget(reply)">
                      回复
                    </button>
                  </div>
                </article>
              </div>
            </article>
          </div>

          <StatePanel
            v-else-if="!commentsLoading"
            compact
            tone="info"
            eyebrow="评论区"
            title="当前还没有跟帖内容"
            description="如果你有判断依据或风险提醒，可以先发第一条评论。"
          />
        </article>
      </div>

      <aside class="finance-stack-tight finance-sticky-side">
        <article class="card finance-section-card topic-side-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">关联对象</h2>
              <p class="section-subtitle">讨论来自站内真实内容对象。</p>
            </div>
          </header>
          <div class="finance-card-pale topic-side-highlight">
            <p>{{ detailCard.linkTypeLabel }}</p>
            <strong>{{ detailCard.linkLabel }}</strong>
            <span>{{ linkedSourceHint }}</span>
            <div class="topic-side-highlight-actions">
              <button type="button" class="finance-mini-btn finance-mini-btn-soft" @click="openLinkedSource">
                {{ linkedSourceActionLabel }}
              </button>
            </div>
          </div>
        </article>

        <article class="card finance-section-card topic-side-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">互动说明</h2>
              <p class="section-subtitle">优先保留高频、真实可运营的动作。</p>
            </div>
          </header>
          <div class="topic-side-list">
            <article v-for="item in interactionRows" :key="item.title" class="finance-list-card finance-list-card-panel">
              <p>{{ item.title }}</p>
              <strong>{{ item.summary }}</strong>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </article>

        <article class="card finance-section-card topic-side-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">举报处理</h2>
              <p class="section-subtitle">发现有问题的内容时可提交复核。</p>
            </div>
          </header>
          <textarea
            v-model.trim="reportReason"
            maxlength="120"
            :disabled="reportSubmitting"
            placeholder="例如：标题过度情绪化、缺少风险提示、与正文不符。"
          />
          <button type="button" class="finance-primary-btn" :disabled="reportSubmitting" @click="submitReport">
            {{ reportSubmitting ? "提交中..." : "提交举报" }}
          </button>
          <p class="topic-side-note">当前举报只支持主题级别，管理员审核后再做处理。</p>
        </article>
      </aside>
    </section>
  </section>
</template>

<script setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import StatePanel from "../components/StatePanel.vue";
import {
  createCommunityComment,
  createCommunityReaction,
  createCommunityReport,
  deleteCommunityReaction,
  getCommunityTopicDetail,
  listCommunityComments
} from "../api/community";
import { useClientAuth } from "../lib/client-auth";
import { buildLinkedContentRoute } from "../lib/community-entry-links";

const route = useRoute();
const router = useRouter();
const { isLoggedIn } = useClientAuth();

const detailLoading = ref(false);
const commentsLoading = ref(false);
const commentSubmitting = ref(false);
const reportSubmitting = ref(false);
const commentErrorMessage = ref("");
const reportReason = ref("");
const commentDraft = ref("");
const commentComposerRef = ref(null);
const reactionLoading = reactive({
  like: false,
  favorite: false
});

const detail = ref({});
const comments = ref([]);
const commentsTotal = ref(0);
const replyTarget = ref(null);

const interactionRows = [
  {
    title: "点赞 / 收藏",
    summary: "只保留高频动作",
    desc: "点赞用于认可观点，收藏用于后续继续跟踪。"
  },
  {
    title: "单层回复",
    summary: "避免讨论结构失控",
    desc: "当前只支持一层回复，适合围绕一个核心判断继续延展。"
  },
  {
    title: "主题举报",
    summary: "先接住基础治理",
    desc: "举报会进入管理员复核，不开放前台即时下架。"
  }
];

const accessState = computed(() =>
  isLoggedIn.value
    ? {
        tone: "info",
        label: "已登录",
        title: "你可以继续评论、回复、点赞、收藏和举报。",
        desc: "互动动作只围绕当前真实支持能力展开，不扩成聊天产品。"
      }
    : {
        tone: "warning",
        label: "游客模式",
        title: "当前可阅读主题与评论，登录后再参与互动。",
        desc: "登录后会回到当前讨论详情页，保持阅读位置。"
      }
);

const detailCard = computed(() => {
  const item = detail.value || {};
  return {
    id: item.id || "",
    title: item.title || "主题详情",
    summary: item.summary || "暂无摘要",
    typeLabel: mapTopicType(item.topic_type),
    stanceLabel: mapStance(item.stance),
    stanceClass: stanceClass(item.stance),
    statusLabel: mapStatus(item.status),
    linkTypeLabel: mapLinkType(item.linked_target?.target_type),
    linkLabel: item.linked_target?.target_snapshot || item.linked_target?.target_id || "未关联对象",
    authorLabel: maskUserID(item.user_id),
    lastActiveLabel: formatDateTime(item.last_active_at),
    createdLabel: formatDateTime(item.created_at),
    horizonLabel: mapHorizon(item.time_horizon)
  };
});

const linkedSourceRoute = computed(() =>
  buildLinkedContentRoute({
    targetType: detail.value?.linked_target?.target_type,
    targetID: detail.value?.linked_target?.target_id
  })
);

const linkedSourceActionLabel = computed(() => {
  switch (String(detail.value?.linked_target?.target_type || "").toUpperCase()) {
    case "NEWS_ARTICLE":
      return "回资讯页查看原文";
    case "STOCK":
      return "回策略页查看股票";
    case "FUTURES":
      return "回策略页查看期货";
    case "STRATEGY_ITEM":
      return "回策略页";
    default:
      return "回讨论广场";
  }
});

const linkedSourceHint = computed(() => {
  switch (String(detail.value?.linked_target?.target_type || "").toUpperCase()) {
    case "NEWS_ARTICLE":
      return "可回到资讯页继续读原文，再和当前讨论一起对照。";
    case "STOCK":
      return "可回到策略页继续看这只股票的推荐逻辑和风险边界。";
    case "FUTURES":
      return "可回到策略页继续看当前期货参数和执行边界。";
    case "STRATEGY_ITEM":
      return "可回到策略页继续查看原始策略内容。";
    default:
      return "当前只保留对象快照，不做跨页实时联动改写。";
  }
});

const commentRows = computed(() => {
  const mapped = comments.value.map((item) => {
    const parentID = item.parent_comment_id || "";
    return {
      id: item.id,
      parentCommentID: parentID,
      rootCommentID: parentID || item.id,
      userID: item.user_id,
      authorLabel: maskUserID(item.user_id),
      replyToLabel: item.reply_to_user_id ? maskUserID(item.reply_to_user_id) : "",
      content: item.content || "",
      likeCount: item.like_count || 0,
      likedByMe: !!item.liked_by_me,
      createdLabel: formatDateTime(item.created_at),
      statusLabel: mapCommentStatus(item.status),
      replies: []
    };
  });

  const rootMap = new Map();
  const rows = [];

  mapped.forEach((item) => {
    if (!item.parentCommentID) {
      rootMap.set(item.id, item);
      rows.push(item);
    }
  });

  mapped.forEach((item) => {
    if (!item.parentCommentID) {
      return;
    }
    const parent = rootMap.get(item.parentCommentID);
    if (parent) {
      parent.replies.push(item);
      return;
    }
    rows.push(item);
  });

  return rows;
});

watch(
  () => route.params.id,
  () => {
    loadTopicPage();
  }
);

onMounted(() => {
  loadTopicPage();
});

async function loadTopicPage() {
  await Promise.all([loadTopicDetail(), loadComments()]);
}

async function loadTopicDetail() {
  detailLoading.value = true;
  try {
    detail.value = await getCommunityTopicDetail(route.params.id);
  } catch (error) {
    detail.value = {};
    commentErrorMessage.value = error?.message || "主题详情加载失败";
  } finally {
    detailLoading.value = false;
  }
}

async function loadComments() {
  commentsLoading.value = true;
  commentErrorMessage.value = "";
  try {
    const result = await listCommunityComments(route.params.id, { page: 1, page_size: 50 });
    comments.value = Array.isArray(result?.items) ? result.items : [];
    commentsTotal.value = Number(result?.total || 0);
  } catch (error) {
    comments.value = [];
    commentsTotal.value = 0;
    commentErrorMessage.value = error?.message || "评论加载失败";
  } finally {
    commentsLoading.value = false;
  }
}

async function submitComment() {
  if (!isLoggedIn.value) {
    router.push({ path: "/auth", query: { redirect: route.fullPath } });
    return;
  }
  if (!commentDraft.value) {
    commentErrorMessage.value = "请先输入评论内容";
    return;
  }

  commentSubmitting.value = true;
  commentErrorMessage.value = "";
  try {
    await createCommunityComment(route.params.id, {
      parent_comment_id: replyTarget.value?.parentPayloadID || "",
      reply_to_user_id: replyTarget.value?.userID || "",
      content: commentDraft.value
    });
    commentDraft.value = "";
    replyTarget.value = null;
    await loadTopicPage();
  } catch (error) {
    commentErrorMessage.value = error?.message || "评论提交失败";
  } finally {
    commentSubmitting.value = false;
  }
}

async function toggleTopicReaction(reactionType) {
  if (!isLoggedIn.value) {
    router.push({ path: "/auth", query: { redirect: route.fullPath } });
    return;
  }
  const reactionKey = reactionType === "FAVORITE" ? "favorite" : "like";
  reactionLoading[reactionKey] = true;
  try {
    const active =
      reactionType === "FAVORITE" ? !!detail.value.favorited_by_me : !!detail.value.liked_by_me;
    if (active) {
      await deleteCommunityReaction({
        target_type: "TOPIC",
        target_id: detail.value.id,
        reaction_type: reactionType
      });
    } else {
      await createCommunityReaction({
        target_type: "TOPIC",
        target_id: detail.value.id,
        reaction_type: reactionType
      });
    }
    await loadTopicDetail();
  } catch (error) {
    commentErrorMessage.value = error?.message || "互动操作失败";
  } finally {
    reactionLoading[reactionKey] = false;
  }
}

async function toggleCommentLike(item) {
  if (!isLoggedIn.value) {
    router.push({ path: "/auth", query: { redirect: route.fullPath } });
    return;
  }
  const key = `comment-${item.id}`;
  reactionLoading[key] = true;
  try {
    if (item.likedByMe) {
      await deleteCommunityReaction({
        target_type: "COMMENT",
        target_id: item.id,
        reaction_type: "LIKE"
      });
    } else {
      await createCommunityReaction({
        target_type: "COMMENT",
        target_id: item.id,
        reaction_type: "LIKE"
      });
    }
    await loadComments();
  } catch (error) {
    commentErrorMessage.value = error?.message || "评论点赞失败";
  } finally {
    reactionLoading[key] = false;
  }
}

function setReplyTarget(item) {
  replyTarget.value = {
    id: item.id,
    userID: item.userID,
    authorLabel: item.authorLabel,
    parentPayloadID: item.parentCommentID || item.id
  };
  handleCommentFocus();
}

function clearReplyTarget() {
  replyTarget.value = null;
}

async function submitReport() {
  if (!isLoggedIn.value) {
    router.push({ path: "/auth", query: { redirect: route.fullPath } });
    return;
  }
  if (!reportReason.value) {
    commentErrorMessage.value = "请先填写举报原因";
    return;
  }
  reportSubmitting.value = true;
  try {
    await createCommunityReport({
      target_type: "TOPIC",
      target_id: detail.value.id,
      reason: reportReason.value
    });
    reportReason.value = "";
    await loadTopicDetail();
  } catch (error) {
    commentErrorMessage.value = error?.message || "举报提交失败";
  } finally {
    reportSubmitting.value = false;
  }
}

async function handleCommentFocus() {
  if (!isLoggedIn.value) {
    router.push({ path: "/auth", query: { redirect: route.fullPath } });
    return;
  }
  await nextTick();
  commentComposerRef.value?.scrollIntoView({ behavior: "smooth", block: "center" });
}

function openLinkedSource() {
  router.push(linkedSourceRoute.value);
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

function mapCommentStatus(value) {
  switch (String(value || "").toUpperCase()) {
    case "PUBLISHED":
      return "公开";
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

function mapHorizon(value) {
  switch (String(value || "").toUpperCase()) {
    case "SHORT":
      return "短线";
    case "SWING":
      return "波段";
    case "MID":
      return "中线";
    case "LONG":
      return "中长线";
    default:
      return "时间未标注";
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
.community-topic-page {
  display: grid;
  gap: 16px;
}

.topic-hero {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 16px;
  align-items: start;
}

.topic-hero-tags,
.topic-stat-strip,
.topic-action-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.topic-hero-actions {
  display: grid;
  gap: 10px;
  min-width: 148px;
}

.topic-detail-layout {
  --finance-main-column: minmax(0, 1.56fr);
  --finance-side-column: minmax(280px, 0.88fr);
}

.topic-core-card,
.topic-comment-card,
.topic-side-card {
  display: grid;
  gap: 14px;
}

.topic-core-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.topic-reading-stack,
.topic-side-list {
  display: grid;
  gap: 10px;
}

.topic-reading-card,
.topic-content-card,
.topic-comment-composer,
.topic-side-highlight {
  padding: 14px;
}

.topic-reading-head,
.topic-comment-head {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 10px;
}

.topic-reading-head p,
.topic-reading-head span,
.topic-comment-head p,
.topic-comment-head strong {
  margin: 0;
}

.topic-reading-card strong,
.topic-content-text {
  display: block;
  margin: 0;
  line-height: 1.72;
}

.topic-reading-card p,
.topic-side-note {
  margin: 0;
  line-height: 1.68;
  color: var(--color-text-sub);
}

.topic-side-highlight-actions {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.topic-content-text {
  white-space: pre-wrap;
  color: var(--color-text-main);
}

.topic-comment-meta {
  display: grid;
  gap: 4px;
  justify-items: end;
  font-size: 12px;
  color: var(--color-text-sub);
}

.topic-comment-composer {
  display: grid;
  gap: 12px;
}

.topic-comment-composer textarea,
.topic-side-card textarea {
  min-height: 120px;
  border: 1px solid var(--color-border-soft);
  border-radius: 14px;
  padding: 12px;
  resize: vertical;
  font: inherit;
  color: var(--color-text-main);
  background: rgba(255, 255, 255, 0.92);
}

.topic-comment-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.topic-comment-actions p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.topic-comment-list,
.topic-reply-list {
  display: grid;
  gap: 10px;
}

.topic-comment-item,
.topic-reply-item {
  display: grid;
  gap: 10px;
}

.topic-comment-top {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: 10px;
}

.topic-comment-top p,
.topic-comment-top strong,
.topic-comment-content {
  margin: 0;
}

.topic-comment-side {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.topic-comment-content {
  line-height: 1.72;
  color: var(--color-text-main);
}

.topic-reply-list {
  padding-left: 14px;
  border-left: 2px solid var(--color-border-soft);
}

@media (max-width: 1080px) {
  .topic-core-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .topic-hero {
    grid-template-columns: minmax(0, 1fr);
  }

  .topic-hero-actions {
    min-width: 0;
  }

  .topic-core-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .topic-comment-top {
    flex-direction: column;
  }

  .topic-comment-side {
    justify-content: flex-start;
  }
}
</style>
