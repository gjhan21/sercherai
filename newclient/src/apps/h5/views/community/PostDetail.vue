<template>
  <div class="h5-post-detail">
    <div v-if="post">
      <div class="h5-post-header">
        <div class="h5-post-author-row">
          <span class="h5-pa-avatar" style="background:var(--accent-gold-glow);color:var(--accent-gold)">{{ (post.author?.nickname || '?')[0] }}</span>
          <div><strong>{{ post.author?.nickname || '用户' }}</strong><span class="h5-pa-time">{{ (post.createdAt || '').replace('T', ' ') }}</span></div>
        </div>
      </div>
      <h2>{{ post.title }}</h2>
      <p class="h5-post-content">{{ post.content || post.summary || '' }}</p>
      <div class="h5-post-actions">
        <button class="h5-like-btn" :class="{ liked }" @click="liked = !liked">{{ liked ? '♥' : '♡' }} {{ liked ? (post.likes || 0) + 1 : (post.likes || 0) }}</button>
      </div>
      <div class="h5-comments">
        <h3>评论 ({{ postComments.length }})</h3>
        <div v-for="c in postComments" :key="c.id" class="h5-comment">
          <span class="h5-com-avatar" style="background:var(--accent-blue-glow);color:var(--accent-blue)">{{ (c.author?.nickname || '?')[0] }}</span>
          <div class="h5-com-body"><strong>{{ c.author?.nickname || '匿名' }}</strong><p>{{ c.content }}</p></div>
        </div>
      </div>
    </div>
    <div v-else class="h5-empty"><p>加载中...</p></div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { POSTS, COMMENTS } from "@/mock/community.js";
import { getCommunityTopicDetail } from "@/api/community.js";

const route = useRoute();
const liked = ref(false);
const postData = ref(null);
const commentList = ref([]);

const post = computed(() => postData.value);
const postComments = computed(() => commentList.value);

async function loadPost() {
  const id = route.params.id;
  if (!id) return;
  try {
    const result = await getCommunityTopicDetail(id);
    if (result) {
      postData.value = {
        id: result.id, title: result.title, content: result.content || result.summary || '',
        author: { nickname: '用户', level: '' }, createdAt: result.created_at || '',
        likes: result.like_count || 0, comments: result.comment_count || 0
      };
      return;
    }
  } catch { /* fall through */ }
  const mock = POSTS.find(p => p.id === id);
  if (mock) postData.value = { ...mock };
  if (COMMENTS[id]) commentList.value = [...COMMENTS[id]];
}

onMounted(loadPost);
</script>

<style scoped>
.h5-post-detail { display: grid; gap: 14px; }
.h5-post-header { padding: 14px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); }
.h5-post-author-row { display: flex; gap: 10px; align-items: center; }
.h5-pa-avatar { width: 32px; height: 32px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; flex-shrink: 0; }
.h5-pa-time { display: block; font-size: 10px; color: var(--text-muted); }
.h5-post-detail h2 { font-size: 18px; font-weight: 800; padding: 0 4px; }
.h5-post-content { font-size: 14px; color: var(--text-secondary); line-height: 1.8; padding: 0 4px; }
.h5-post-actions { padding: 10px 0; }
.h5-like-btn { padding: 8px 20px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 13px; background: none; color: var(--text-primary); cursor: pointer; }
.h5-like-btn.liked { background: var(--negative-bg); color: var(--negative); border-color: rgba(255,71,87,.3); }
.h5-comments h3 { font-size: 15px; font-weight: 700; margin-bottom: 10px; }
.h5-comment { display: flex; gap: 8px; padding: 10px 0; border-bottom: 1px solid rgba(255,255,255,.03); }
.h5-com-avatar { width: 26px; height: 26px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 10px; font-weight: 700; flex-shrink: 0; }
.h5-com-body strong { font-size: 12px; display: block; }
.h5-com-body p { font-size: 12px; color: var(--text-secondary); line-height: 1.5; margin-top: 2px; }
.h5-empty { text-align: center; padding: 40px; color: var(--text-secondary); }
</style>
