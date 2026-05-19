<template>
  <div class="post-detail-page" v-if="post">
    <button class="back-btn glass" @click="$router.push('/community')">← 返回社区</button>

    <article class="section fade-in-up">
      <div class="post-detail-header">
        <div class="post-author-row">
          <div class="post-avatar" :style="{background: (post.author?.color || 'var(--accent-gold)') + '22', color: post.author?.color || 'var(--accent-gold)'}">{{ (post.author?.nickname || '?')[0] }}</div>
          <div>
            <strong>{{ post.author?.nickname || '匿名' }}</strong>
            <div class="author-meta">
              <span>{{ post.author?.level || '' }}</span>
              <span>{{ (post.createdAt || '').replace('T', ' ') }}</span>
            </div>
          </div>
        </div>
      </div>

      <h1 class="post-title">{{ post.title }}</h1>
      <div class="post-tags">
        <span v-for="t in (post.tags || [])" :key="t" class="tag tag-neutral">{{ t }}</span>
        <span v-if="post.stock" class="tag tag-gold" @click="goStock(post.stock)">{{ post.stock }}</span>
      </div>
      <p class="post-full-content">{{ post.content || post.summary || '' }}</p>

      <div class="post-actions-bar">
        <button class="action-btn" :class="{ liked: liked }" @click="toggleLike">
          <span>{{ liked ? '♥' : '♡' }}</span>
          <span>{{ liked ? (post.likes || 0) + 1 : (post.likes || 0) }}</span>
        </button>
        <span class="action-btn">💬 {{ post.comments || post.comment_count || 0 }} 评论</span>
        <span class="action-btn">👁 {{ post.views || 0 }} 浏览</span>
      </div>

      <!-- Comment Form -->
      <div class="comment-form" v-if="isLoggedIn">
        <textarea v-model="commentText" placeholder="写下你的看法..." rows="3"></textarea>
        <div class="comment-form-actions">
          <button class="btn-primary" @click="submitComment" :disabled="!commentText.trim() || commenting">{{ commenting ? '提交中...' : '发表评论' }}</button>
        </div>
        <p v-if="commentError" class="comment-error">{{ commentError }}</p>
      </div>

      <div class="comments-section">
        <h3>评论 ({{ postComments.length }})</h3>
        <div v-for="c in postComments" :key="c.id" class="comment-card">
          <div class="comment-avatar" :style="{background: (c.author?.color || 'var(--accent-blue)') + '22', color: c.author?.color || 'var(--accent-blue)'}">{{ (c.author?.nickname || '?')[0] }}</div>
          <div class="comment-body">
            <div class="comment-header">
              <strong>{{ c.author?.nickname || '匿名' }}</strong>
              <span class="comment-time">{{ (c.createdAt || '').replace('T', ' ') }}</span>
            </div>
            <p class="comment-content">{{ c.content }}</p>
            <div class="comment-replies" v-if="c.replies?.length">
              <div v-for="r in c.replies" :key="r.id" class="reply-item">
                <strong>{{ r.author?.nickname || '匿名' }}</strong>
                <p>{{ r.content }}</p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </article>
  </div>

  <div v-else class="not-found">
    <p>{{ loading ? '加载中...' : '帖子未找到' }}</p>
    <button class="btn-primary" @click="$router.push('/community')">返回社区</button>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { POSTS, COMMENTS, USERS } from "@/mock/community.js";
import { getCommunityTopicDetail, listCommunityComments, createCommunityComment, createCommunityReaction, deleteCommunityReaction } from "@/api/community.js";
import { useClientAuth } from "@/shared/auth/client-auth";

const route = useRoute();
const router = useRouter();
const { isLoggedIn } = useClientAuth();
const liked = ref(false);
const loading = ref(false);
const postData = ref(null);
const commentList = ref([]);
const commentText = ref("");
const commenting = ref(false);
const commentError = ref("");

const post = computed(() => postData.value);
const postComments = computed(() => commentList.value);

function goStock(symbol) { router.push('/identify/' + symbol); }

async function toggleLike() {
  if (!post.value?.id) return;
  liked.value = !liked.value;
  try {
    if (liked.value) {
      await createCommunityReaction({ target_type: "TOPIC", target_id: post.value.id, reaction_type: "LIKE" });
    } else {
      await deleteCommunityReaction({ target_type: "TOPIC", target_id: post.value.id, reaction_type: "LIKE" });
    }
  } catch { liked.value = !liked.value; }
}

async function submitComment() {
  const text = commentText.value.trim();
  if (!text || !post.value?.id) return;
  commenting.value = true;
  commentError.value = "";
  try {
    await createCommunityComment(post.value.id, { content: text });
    commentText.value = "";
    // Reload comments
    try {
      const comments = await listCommunityComments(post.value.id, { page: 1, page_size: 20 });
      if (comments?.items) {
        commentList.value = comments.items.map(c => ({
          id: c.id, content: c.content || '',
          author: { nickname: '用户', color: 'var(--accent-blue)' },
          createdAt: c.created_at || '', likes: c.like_count || 0, replies: []
        }));
      }
    } catch { /* keep comment list as is */ }
  } catch (e) {
    commentError.value = e?.message || "评论失败";
  } finally { commenting.value = false; }
}

async function loadPost() {
  const id = route.params.id;
  if (!id) return;
  loading.value = true;

  try {
    const result = await getCommunityTopicDetail(id);
    if (result) {
      postData.value = {
        id: result.id, title: result.title, content: result.content || result.summary || '',
        author: { nickname: '用户', level: '', color: 'var(--accent-gold)' },
        createdAt: result.created_at || '',
        likes: result.like_count || 0, comments: result.comment_count || 0, views: (result.like_count || 0) * 10,
        tags: [result.topic_type || '讨论'], stock: result.linked_target?.symbol || null
      };
      // Load comments
      try {
        const comments = await listCommunityComments(id, { page: 1, page_size: 20 });
        if (comments?.items) {
          commentList.value = comments.items.map(c => ({
            id: c.id, content: c.content || '',
            author: { nickname: '用户', color: 'var(--accent-blue)' },
            createdAt: c.created_at || '', likes: c.like_count || 0, replies: []
          }));
        }
      } catch { /* use mock fallback for comments */ }
      loading.value = false;
      return;
    }
  } catch { /* fall through */ }

  // Mock fallback
  const mockPost = POSTS.find(p => p.id === id);
  if (mockPost) postData.value = { ...mockPost };
  if (COMMENTS[id]) commentList.value = [...COMMENTS[id]];
  loading.value = false;
}

onMounted(loadPost);
</script>

<style scoped>
.post-detail-page { max-width: 1000px; display: grid; gap: 16px; }
.back-btn { display: inline-flex; padding: 8px 16px; border-radius: var(--radius-full); font-size: 13px; color: var(--text-secondary); cursor: pointer; width: auto; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 28px; }
.post-detail-header { margin-bottom: 16px; }
.post-author-row { display: flex; gap: 12px; margin-bottom: 12px; align-items: center; }
.author-meta { display: flex; gap: 8px; font-size: 11px; color: var(--text-muted); margin-top: 2px; }
.post-title { font-size: 24px; font-weight: 800; margin-bottom: 8px; line-height: 1.35; }
.post-tags { display: flex; gap: 6px; margin-bottom: 16px; flex-wrap: wrap; }
.post-tags .tag-gold { cursor: pointer; }
.post-full-content { font-size: 15px; line-height: 1.9; color: var(--text-secondary); margin-bottom: 20px; }
.post-actions-bar { display: flex; gap: 16px; margin-bottom: 24px; padding: 12px 0; border-top: 1px solid var(--border); border-bottom: 1px solid var(--border); }
.action-btn { display: flex; align-items: center; gap: 6px; font-size: 13px; color: var(--text-secondary); cursor: pointer; }
.action-btn.liked { color: var(--negative); }
.comments-section h3 { font-size: 16px; font-weight: 700; margin-bottom: 16px; }
.comment-card { display: flex; gap: 10px; padding: 12px 0; border-bottom: 1px solid rgba(255,255,255,.03); }
.comment-avatar { width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px; flex-shrink: 0; }
.comment-body { flex: 1; }
.comment-header { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.comment-header strong { font-size: 13px; }
.comment-time { font-size: 11px; color: var(--text-muted); }
.comment-content { font-size: 13px; line-height: 1.6; color: var(--text-secondary); margin-bottom: 8px; }
.comment-replies { margin-top: 8px; padding-left: 16px; border-left: 2px solid var(--border); display: grid; gap: 8px; }
.reply-item { padding: 6px 0; }
.reply-item strong { font-size: 12px; }
.reply-item p { font-size: 12px; color: var(--text-secondary); margin-top: 2px; }
.comment-form { margin-bottom: 20px; display: grid; gap: 8px; }
.comment-form textarea { width: 100%; padding: 12px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 13px; resize: vertical; font-family: inherit; }
.comment-form textarea:focus { border-color: var(--accent-gold); }
.comment-form-actions { display: flex; justify-content: flex-end; }
.comment-error { color: var(--negative); font-size: 12px; }
.btn-primary { padding: 8px 18px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 13px; font-weight: 600; cursor: pointer; border: none; }
.not-found { text-align: center; padding: 60px; }
.not-found p { color: var(--text-secondary); margin-bottom: 16px; }
</style>
