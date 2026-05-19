<template>
  <div class="room-page" v-if="room">
    <button class="back-btn glass" @click="$router.push('/community')">← 返回社区</button>

    <section class="room-header glass fade-in-up">
      <div class="room-header-info">
        <h1 class="room-name">{{ room.roomName }}</h1>
        <p class="room-desc">{{ room.description }}</p>
        <div class="room-stats">
          <span>👥 {{ room.memberCount }} 人</span>
          <span>📝 {{ roomPosts.length }} 活跃帖子</span>
          <span>🕐 最近活跃: {{ room.lastActive }}</span>
        </div>
      </div>
      <div class="room-stock-badge">
        <span class="room-stock-symbol">{{ (room.symbol || '').split('.')[0] }}</span>
        <span class="room-stock-name">{{ room.name }}</span>
      </div>
    </section>

    <section class="section fade-in-up fade-in-up-delay-1">
      <div class="section-header">
        <h2 class="section-title">相关讨论</h2>
        <span class="section-subtitle">{{ roomPosts.length }} 个帖子</span>
      </div>
      <div class="post-feed">
        <div v-for="post in roomPosts" :key="post.id" class="post-card glass card-hover" @click="$router.push('/community/post/' + post.id)">
          <div class="post-card-left">
            <div class="post-avatar" :style="{background: (post.author?.color || 'var(--accent-gold)') + '22', color: post.author?.color || 'var(--accent-gold)'}">{{ (post.author?.nickname || '?')[0] }}</div>
          </div>
          <div class="post-card-body">
            <div class="post-card-header">
              <span class="post-author">{{ post.author?.nickname || '用户' }}</span>
              <span class="post-time">{{ (post.createdAt || '').replace('T', ' ') }}</span>
            </div>
            <h3 class="post-title">{{ post.title }}</h3>
            <p class="post-excerpt">{{ (post.content || '').slice(0, 60) }}...</p>
            <div class="post-meta-row">
              <span>{{ post.likes || 0 }} ♥</span>
              <span>{{ post.comments || 0 }} 评论</span>
              <span>{{ post.views || 0 }} 浏览</span>
            </div>
          </div>
        </div>
        <div v-if="!roomPosts.length" class="empty-state"><p>暂无讨论</p></div>
      </div>
    </section>
  </div>

  <div v-else class="not-found">
    <p>讨论圈未找到</p>
    <button class="btn-primary" @click="$router.push('/community')">返回社区</button>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getRoomBySymbol, POSTS as MOCK_POSTS } from "@/mock/community.js";
import { listCommunityTopics } from "@/api/community.js";

const route = useRoute();
const router = useRouter();
const postsData = ref([]);
const symbol = computed(() => route.params.symbol);
const room = computed(() => getRoomBySymbol(symbol.value));
const roomPosts = computed(() => postsData.value.length ? postsData.value : MOCK_POSTS.filter(p => p.stock === symbol.value));

async function loadPosts() {
  try {
    const result = await listCommunityTopics({ page: 1, page_size: 20 });
    if (result?.items?.length) {
      postsData.value = result.items
        .filter(item => item.linked_target?.symbol === symbol.value)
        .map(item => ({
          id: item.id, title: item.title, content: item.summary || item.content || '',
          author: { nickname: '用户', level: '', color: 'var(--accent-gold)' },
          createdAt: item.created_at || '',
          likes: item.like_count || 0, comments: item.comment_count || 0, views: (item.like_count || 0) * 10
        }));
    }
  } catch { /* use mock */ }
}

onMounted(loadPosts);
</script>

<style scoped>
.room-page { max-width: 1000px; display: grid; gap: 16px; }
.back-btn { display: inline-flex; padding: 8px 16px; border-radius: var(--radius-full); font-size: 13px; color: var(--text-secondary); cursor: pointer; width: auto; }
.room-header { display: flex; justify-content: space-between; padding: 24px; border-radius: var(--radius-lg); }
.room-name { font-size: 24px; font-weight: 800; margin-bottom: 8px; }
.room-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.6; margin-bottom: 12px; max-width: 500px; }
.room-stats { display: flex; gap: 16px; font-size: 12px; color: var(--text-muted); }
.room-stock-badge { text-align: right; }
.room-stock-symbol { display: block; font-size: 14px; color: var(--accent-gold); font-family: var(--font-mono); }
.room-stock-name { font-size: 16px; font-weight: 700; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.post-feed { display: grid; gap: 10px; }
.post-card { display: flex; gap: 14px; padding: 14px; border-radius: var(--radius-md); cursor: pointer; }
.post-card-left { flex-shrink: 0; }
.post-avatar { width: 36px; height: 36px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 14px; flex-shrink: 0; }
.post-card-body { flex: 1; }
.post-card-header { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.post-author { font-size: 13px; font-weight: 600; }
.post-time { font-size: 11px; color: var(--text-muted); }
.post-title { font-size: 15px; font-weight: 700; margin-bottom: 4px; line-height: 1.4; }
.post-excerpt { font-size: 12px; color: var(--text-secondary); margin-bottom: 8px; line-height: 1.5; }
.post-meta-row { display: flex; gap: 12px; font-size: 11px; color: var(--text-muted); }
.empty-state { text-align: center; padding: 40px; color: var(--text-secondary); }
.btn-primary { padding: 8px 18px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 13px; font-weight: 600; cursor: pointer; border: none; }
.not-found { text-align: center; padding: 60px; }
.not-found p { color: var(--text-secondary); margin-bottom: 16px; }
</style>
