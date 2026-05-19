<template>
  <div class="h5-community">
    <div class="h5-com-tabs">
      <button v-for="t in tabs" :key="t.key" class="h5-com-tab" :class="{active: activeTab === t.key}" @click="activeTab = t.key">{{ t.label }}</button>
    </div>
    <div class="h5-room-bar">
      <button v-for="room in STOCK_ROOMS.slice(0,4)" :key="room.symbol" class="h5-room-chip" @click="$router.push('/community/post/' + (getFirstPost(room.symbol)?.id || ''))">{{ room.symbol.split('.')[0] }}</button>
    </div>
    <div class="h5-posts">
      <div v-for="post in filtered" :key="post.id" class="h5-post-card" @click="$router.push('/community/post/' + post.id)">
        <div class="h5-post-author">
          <span class="h5-post-avatar" :style="{background: (post.author?.color || 'var(--accent-gold)') + '22', color: post.author?.color || 'var(--accent-gold)'}">{{ (post.author?.nickname || '?')[0] }}</span>
          <div><strong>{{ post.author?.nickname || '用户' }}</strong><span class="h5-post-time">{{ (post.createdAt || '').slice(0, 10) }}</span></div>
        </div>
        <h3>{{ post.title }}</h3>
        <div class="h5-post-meta"><span>{{ post.likes || 0 }} ♥</span><span>{{ post.comments || 0 }} 评论</span></div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { POSTS as MOCK_POSTS, STOCK_ROOMS } from "@/mock/community.js";
import { listCommunityTopics } from "@/api/community.js";

const activeTab = ref('hot');
const posts = ref(MOCK_POSTS);
const tabs = [{ key: 'hot', label: '热门' }, { key: 'latest', label: '最新' }];

const filtered = computed(() => {
  const sorted = [...posts.value];
  if (activeTab.value === 'hot') sorted.sort((a, b) => (b.likes || 0) - (a.likes || 0));
  else sorted.sort((a, b) => new Date(b.createdAt || 0) - new Date(a.createdAt || 0));
  return sorted;
});

function getFirstPost(symbol) { return posts.value.find(p => p.stock === symbol); }

async function loadTopics() {
  try {
    const result = await listCommunityTopics({ page: 1, page_size: 20 });
    if (result?.items?.length) {
      posts.value = result.items.map(item => ({
        id: item.id, title: item.title, content: item.summary || '',
        author: { nickname: '用户', level: '', color: 'var(--accent-gold)' },
        createdAt: item.created_at || '', likes: item.like_count || 0,
        comments: item.comment_count || 0, views: (item.like_count || 0) * 10,
        stock: item.linked_target?.symbol || null
      }));
    }
  } catch { /* use mock */ }
}

onMounted(loadTopics);
</script>

<style scoped>
.h5-community { display: grid; gap: 10px; }
.h5-com-tabs { display: flex; gap: 6px; }
.h5-com-tab { padding: 5px 14px; border-radius: var(--radius-full); font-size: 12px; color: var(--text-secondary); }
.h5-com-tab.active { color: #000; background: var(--accent-gold); font-weight: 600; }
.h5-room-bar { display: flex; gap: 6px; overflow-x: auto; }
.h5-room-chip { padding: 5px 10px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 11px; color: var(--text-secondary); white-space: nowrap; font-family: var(--font-mono); }
.h5-posts { display: grid; gap: 8px; }
.h5-post-card { padding: 14px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-md); }
.h5-post-card:active { background: rgba(255,255,255,.03); }
.h5-post-author { display: flex; gap: 8px; margin-bottom: 8px; }
.h5-post-avatar { width: 28px; height: 28px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; flex-shrink: 0; }
.h5-post-author strong { font-size: 12px; display: block; }
.h5-post-time { font-size: 10px; color: var(--text-muted); }
.h5-post-card h3 { font-size: 14px; font-weight: 700; margin-bottom: 8px; line-height: 1.4; }
.h5-post-meta { display: flex; gap: 12px; font-size: 11px; color: var(--text-muted); }
</style>
