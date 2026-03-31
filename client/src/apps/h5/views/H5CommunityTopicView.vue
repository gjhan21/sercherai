<template>
  <div class="h5-page fade-up topic-detail-page">
    <div class="h5-page-topline">
      <button type="button" class="h5-topline-back" @click="router.push('/community')">返回广场</button>
      <span>{{ detail.id ? "讨论详情" : "加载中..." }}</span>
    </div>

    <H5HeroCard
      :eyebrow="detailCard.typeLabel"
      :title="detailCard.title"
      :description="detailCard.summary"
      tone="accent"
    >
      <div class="h5-chip-list">
        <span class="h5-pill h5-pill-info">{{ detailCard.typeLabel }}</span>
        <span class="h5-pill" :class="detailCard.stanceClassH5">{{ detailCard.stanceLabel }}</span>
        <span class="h5-pill h5-pill-neutral">{{ detailCard.horizonLabel }}</span>
      </div>

      <div class="topic-detail-metrics h5-grid-3">
        <H5SummaryCard label="评论数" :value="`${commentsTotal} 条`" note="包含回复" tone="gold" />
        <H5SummaryCard label="点赞数" :value="String(detail.like_count || 0)" note="认可观点" tone="brand" />
        <H5SummaryCard label="作者" :value="detailCard.authorLabel" :note="detailCard.createdLabel" tone="soft" />
      </div>
    </H5HeroCard>

    <H5SectionBlock eyebrow="观点核心" title="看结论、依据与风险边界" tone="soft">
      <div class="topic-reading-stack">
        <article class="topic-reading-card accent">
          <div class="topic-reading-head">
            <strong>为什么这样看</strong>
            <span>依据</span>
          </div>
          <p>{{ detail.reason_text || "暂无文字补充" }}</p>
        </article>

        <article class="topic-reading-card danger">
          <div class="topic-reading-head">
            <strong>风险边界</strong>
            <span>失效条件</span>
          </div>
          <p>{{ detail.risk_text || "暂无风险标注" }}</p>
        </article>

        <article class="topic-reading-prose">
          <div class="topic-reading-head">
            <strong>完整正文</strong>
            <span>发表于 {{ detailCard.createdLabel }}</span>
          </div>
          <p>{{ detail.content || "暂无正文内容" }}</p>
        </article>
      </div>

      <div class="topic-interaction-row">
        <button
          type="button"
          class="h5-btn"
          :disabled="reactionLoading.like || !detail.id"
          @click="toggleTopicReaction('LIKE')"
        >
          {{ detail.liked_by_me ? "取消点赞" : "点赞支持" }}
        </button>
        <button
          type="button"
          class="h5-btn-secondary"
          :disabled="reactionLoading.favorite || !detail.id"
          @click="toggleTopicReaction('FAVORITE')"
        >
          {{ detail.favorited_by_me ? "已收藏" : "加入收藏" }}
        </button>
      </div>
    </H5SectionBlock>

    <H5SectionBlock
      v-if="detail.linked_target"
      eyebrow="关联对象"
      title="看看原始资讯或策略逻辑"
      tone="soft"
    >
      <div class="topic-linked-card">
        <div class="topic-linked-copy">
          <strong>{{ detailCard.linkTypeLabel }}</strong>
          <p>{{ detailCard.linkLabel }}</p>
        </div>
        <button type="button" class="h5-btn-ghost" @click="openLinkedSource">查看对象</button>
      </div>
    </H5SectionBlock>

    <H5SectionBlock id="comments-section" eyebrow="讨论跟帖" :title="`当前收到 ${commentsTotal} 条评论`" tone="soft">
      <div ref="commentComposerRef" class="topic-comment-composer">
        <div class="topic-composer-head">
          <span>{{ replyTarget ? `回复 ${replyTarget.authorLabel}` : "写评论" }}</span>
          <button v-if="replyTarget" type="button" class="h5-btn-ghost-sm" @click="clearReplyTarget">取消</button>
        </div>
        <textarea
          v-model.trim="commentDraft"
          :disabled="commentSubmitting"
          maxlength="600"
          placeholder="补充你的判断逻辑、证据或风险提醒..."
        />
        <div class="topic-composer-foot">
          <span>{{ commentDraft.length }}/600</span>
          <button type="button" class="h5-btn" :disabled="commentSubmitting || !commentDraft" @click="submitComment">
            {{ commentSubmitting ? "正在提交" : "发布评论" }}
          </button>
        </div>
      </div>

      <div v-if="commentRows.length" class="topic-comment-list">
        <article v-for="item in commentRows" :key="item.id" class="topic-comment-item">
          <div class="topic-comment-head">
            <span class="comment-author">{{ item.authorLabel }}</span>
            <span class="comment-time">{{ item.createdLabel }}</span>
          </div>
          <p class="comment-content">{{ item.content }}</p>
          <div class="topic-comment-foot">
            <button
              type="button"
              class="h5-btn-ghost-sm"
              :disabled="reactionLoading[`comment-${item.id}`]"
              @click="toggleCommentLike(item)"
            >
              {{ item.likedByMe ? "已赞" : "赞" }} {{ item.likeCount || 0 }}
            </button>
            <button type="button" class="h5-btn-ghost-sm" @click="setReplyTarget(item)">回复</button>
          </div>

          <div v-if="item.replies.length" class="topic-reply-list">
            <article v-for="reply in item.replies" :key="reply.id" class="topic-reply-item">
              <div class="topic-comment-head">
                <span class="comment-author">{{ reply.authorLabel }}</span>
                <span class="comment-time">{{ reply.createdLabel }}</span>
              </div>
              <p class="comment-content">
                <span v-if="reply.replyToLabel" class="reply-target">回复 @{{ reply.replyToLabel }}：</span>
                {{ reply.content }}
              </p>
              <div class="topic-comment-foot">
                <button
                  type="button"
                  class="h5-btn-ghost-sm"
                  :disabled="reactionLoading[`comment-${reply.id}`]"
                  @click="toggleCommentLike(reply)"
                >
                  {{ reply.likedByMe ? "已赞" : "赞" }} {{ reply.likeCount || 0 }}
                </button>
                <button type="button" class="h5-btn-ghost-sm" @click="setReplyTarget(reply)">回复</button>
              </div>
            </article>
          </div>
        </article>
      </div>

      <H5EmptyState
        v-else-if="!commentsLoading"
        title="暂无讨论跟帖"
        description="如果你有判断依据或风险提醒，可以写下第一条评论。"
      />
    </H5SectionBlock>

    <H5StickyCta
      v-if="!isLoggedIn"
      title="登录参与互动"
      description="登录后你可以点赞、收藏和发表你的判断。"
      primary-label="立即登录"
      @primary="router.push({ name: 'h5-auth', query: { redirect: route.fullPath } })"
    />
  </div>
</template>

<script setup>
import { computed, nextTick, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import H5HeroCard from "../components/H5HeroCard.vue";
import H5SectionBlock from "../components/H5SectionBlock.vue";
import H5SummaryCard from "../components/H5SummaryCard.vue";
import H5EmptyState from "../components/H5EmptyState.vue";
import H5StickyCta from "../components/H5StickyCta.vue";
import {
  createCommunityComment,
  createCommunityReaction,
  deleteCommunityReaction,
  getCommunityTopicDetail,
  listCommunityComments
} from "../../../api/community";
import { useClientAuth } from "../../../shared/auth/client-auth";
import { buildLinkedContentRoute } from "../../../lib/community-entry-links";

const route = useRoute();
const router = useRouter();
const { isLoggedIn } = useClientAuth();

const detailLoading = ref(false);
const commentsLoading = ref(false);
const commentSubmitting = ref(false);
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

const detailCard = computed(() => {
  const item = detail.value || {};
  return {
    id: item.id || "",
    title: item.title || "主题详情",
    summary: item.summary || "暂无摘要",
    typeLabel: mapTopicType(item.topic_type),
    stanceLabel: mapStance(item.stance),
    stanceClassH5: stanceClassH5(item.stance),
    linkTypeLabel: mapLinkType(item.linked_target?.target_type),
    linkLabel: item.linked_target?.target_snapshot || item.linked_target?.target_id || "未关联对象",
    authorLabel: maskUserID(item.user_id),
    createdLabel: formatDateTime(item.created_at),
    horizonLabel: mapHorizon(item.time_horizon),
    lastActiveLabel: formatDateTime(item.last_active_at)
  };
});

const linkedSourceRoute = computed(() =>
  buildLinkedContentRoute({
    targetType: detail.value?.linked_target?.target_type,
    targetID: detail.value?.linked_target?.target_id
  })
);

const commentRows = computed(() => {
  const mapped = comments.value.map(item => ({
    id: item.id,
    parentCommentID: item.parent_comment_id || "",
    userID: item.user_id,
    authorLabel: maskUserID(item.user_id),
    replyToLabel: item.reply_to_user_id ? maskUserID(item.reply_to_user_id) : "",
    content: item.content || "",
    likeCount: item.like_count || 0,
    likedByMe: !!item.liked_by_me,
    createdLabel: formatDateTime(item.created_at),
    statusLabel: mapStatus(item.status),
    replies: []
  }));

  const roots = [];
  const rootMap = new Map();

  mapped.forEach(item => {
    if (!item.parentCommentID) {
      rootMap.set(item.id, item);
      roots.push(item);
    }
  });

  mapped.forEach(item => {
    if (item.parentCommentID) {
      const parent = rootMap.get(item.parentCommentID);
      if (parent) parent.replies.push(item);
    }
  });

  return roots;
});

onMounted(() => {
  loadPageData();
});

watch(() => route.params.id, () => loadPageData());

async function loadPageData() {
  await Promise.all([loadTopicDetail(), loadComments()]);
}

async function loadTopicDetail() {
  detailLoading.value = true;
  try {
    detail.value = await getCommunityTopicDetail(route.params.id);
  } finally {
    detailLoading.value = false;
  }
}

async function loadComments() {
  commentsLoading.value = true;
  try {
    const res = await listCommunityComments(route.params.id, { page: 1, page_size: 100 });
    comments.value = res?.items || [];
    commentsTotal.value = Number(res?.total || 0);
  } finally {
    commentsLoading.value = false;
  }
}

async function submitComment() {
  if (!isLoggedIn.value) {
    router.push({ name: "h5-auth", query: { redirect: route.fullPath } });
    return;
  }
  if (!commentDraft.value) return;

  commentSubmitting.value = true;
  try {
    await createCommunityComment(route.params.id, {
      parent_comment_id: replyTarget.value?.id || "",
      reply_to_user_id: replyTarget.value?.userID || "",
      content: commentDraft.value
    });
    commentDraft.value = "";
    replyTarget.value = null;
    await loadComments();
  } finally {
    commentSubmitting.value = false;
  }
}

async function toggleTopicReaction(reactionType) {
  if (!isLoggedIn.value) {
    router.push({ name: "h5-auth", query: { redirect: route.fullPath } });
    return;
  }
  const key = reactionType === "FAVORITE" ? "favorite" : "like";
  reactionLoading[key] = true;
  try {
    const active = reactionType === "FAVORITE" ? !!detail.value.favorited_by_me : !!detail.value.liked_by_me;
    if (active) {
      await deleteCommunityReaction({ target_type: "TOPIC", target_id: detail.value.id, reaction_type: reactionType });
    } else {
      await createCommunityReaction({ target_type: "TOPIC", target_id: detail.value.id, reaction_type: reactionType });
    }
    await loadTopicDetail();
  } finally {
    reactionLoading[key] = false;
  }
}

async function toggleCommentLike(item) {
  if (!isLoggedIn.value) {
    router.push({ name: "h5-auth", query: { redirect: route.fullPath } });
    return;
  }
  const key = `comment-${item.id}`;
  reactionLoading[key] = true;
  try {
    if (item.likedByMe) {
      await deleteCommunityReaction({ target_type: "COMMENT", target_id: item.id, reaction_type: "LIKE" });
    } else {
      await createCommunityReaction({ target_type: "COMMENT", target_id: item.id, reaction_type: "LIKE" });
    }
    await loadComments();
  } finally {
    reactionLoading[key] = false;
  }
}

function setReplyTarget(item) {
  replyTarget.value = item;
  handleCommentFocus();
}

function clearReplyTarget() {
  replyTarget.value = null;
}

async function handleCommentFocus() {
  if (!isLoggedIn.value) {
    router.push({ name: "h5-auth", query: { redirect: route.fullPath } });
    return;
  }
  await nextTick();
  commentComposerRef.value?.scrollIntoView({ behavior: "smooth", block: "center" });
}

function openLinkedSource() {
  router.push(linkedSourceRoute.value);
}

// Helpers
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
function mapLinkType(v) {
  const m = { STOCK: "关联股票", FUTURES: "关联期货", NEWS_ARTICLE: "关联资讯", STRATEGY_ITEM: "关联策略" };
  return m[v] || "关联对象";
}
function mapHorizon(v) {
  const m = { SHORT: "短线", SWING: "波段", MID: "中线", LONG: "长线" };
  return m[v] || "未知";
}
function mapStatus(v) {
  const m = { PUBLISHED: "公开", PENDING_REVIEW: "审核中", HIDDEN: "隐藏" };
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
.topic-detail-page {
  gap: 12px;
}

.h5-topline-back {
  background: none;
  border: 0;
  padding: 0;
  color: var(--h5-brand, #102a56);
  font-size: 13px;
  font-weight: 700;
}

.h5-chip-list {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.topic-detail-metrics {
  margin-top: 4px;
}

.topic-reading-stack {
  display: grid;
  gap: 12px;
}

.topic-reading-card,
.topic-reading-prose {
  padding: 16px;
  border-radius: 12px;
  background: var(--h5-panel-bg, #f9f9f9);
  display: grid;
  gap: 8px;
}

.topic-reading-card.accent {
  border-left: 4px solid var(--h5-brand, #102a56);
}

.topic-reading-card.danger {
  border-left: 4px solid var(--h5-danger, #e64545);
}

.topic-reading-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.topic-reading-head strong {
  font-size: 14px;
  color: var(--h5-text, #333);
}

.topic-reading-head span {
  font-size: 10px;
  color: var(--h5-text-soft, #999);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.topic-reading-card p,
.topic-reading-prose p {
  margin: 0;
  font-size: 14px;
  line-height: 1.6;
  color: var(--h5-text-sub, #555);
}

.topic-interaction-row {
  margin-top: 12px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.topic-linked-card {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  background: var(--h5-panel-bg, #f9f9f9);
  border-radius: 12px;
}

.topic-linked-copy strong {
  display: block;
  font-size: 11px;
  color: var(--h5-brand, #102a56);
  margin-bottom: 2px;
}

.topic-linked-copy p {
  margin: 0;
  font-size: 14px;
  color: var(--h5-text, #333);
  font-weight: 700;
}

.topic-comment-composer {
  background: #fff;
  border: 1px solid var(--h5-line, #eee);
  border-radius: 16px;
  padding: 14px;
  display: grid;
  gap: 10px;
}

.topic-composer-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.topic-composer-head span {
  font-size: 12px;
  font-weight: 700;
  color: var(--h5-text-sub, #666);
}

.topic-comment-composer textarea {
  width: 100%;
  min-height: 80px;
  border: 1px solid var(--h5-line-soft, #f0f0f0);
  border-radius: 8px;
  padding: 10px;
  font-size: 14px;
  background: var(--h5-panel-bg, #f9f9f9);
}

.topic-composer-foot {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.topic-composer-foot span {
  font-size: 11px;
  color: var(--h5-text-soft, #999);
}

.topic-comment-list {
  margin-top: 20px;
  display: grid;
  gap: 20px;
}

.topic-comment-item {
  display: grid;
  gap: 8px;
}

.comment-author {
  font-size: 13px;
  font-weight: 700;
  color: var(--h5-text, #333);
}

.comment-time {
  font-size: 11px;
  color: var(--h5-text-soft, #999);
  margin-left: 8px;
}

.comment-content {
  margin: 0;
  font-size: 14px;
  line-height: 1.55;
  color: var(--h5-text-sub, #444);
}

.topic-comment-foot {
  display: flex;
  gap: 16px;
}

.topic-reply-list {
  margin-top: 8px;
  padding-left: 12px;
  border-left: 2px solid var(--h5-line-soft, #f0f0f0);
  display: grid;
  gap: 14px;
}

.reply-target {
  color: var(--h5-brand, #102a56);
  font-size: 13px;
  font-weight: 700;
}

.h5-btn-ghost-sm {
  background: none;
  border: 0;
  padding: 0;
  color: var(--h5-brand, #102a56);
  font-size: 12px;
  font-weight: 700;
}
</style>
