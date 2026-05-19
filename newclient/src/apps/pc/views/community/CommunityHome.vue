<template>
  <div class="community-page">
    <section class="section fade-in-up">
      <div class="section-header">
        <h2 class="section-title">社区动态</h2>
        <p class="section-subtitle">与万千投资者一起交流分享</p>
        <div class="section-stats">
          <span>{{ allPosts.length }} 帖子</span>
          <button v-if="isLoggedIn" class="btn-primary create-post-btn" @click="showCreateModal = true">+ 发起讨论</button>
        </div>
      </div>
      <div class="community-tabs">
        <button v-for="t in tabs" :key="t.key" class="com-tab" :class="{active: activeTab === t.key}" @click="activeTab = t.key">{{ t.label }}</button>
      </div>

      <div class="rooms-row">
        <div v-for="room in STOCK_ROOMS.slice(0,4)" :key="room.symbol" class="room-mini glass card-hover" @click="$router.push('/community/room/' + room.symbol)">
          <div class="room-mini-top">
            <span class="room-mini-symbol">{{ room.symbol.split('.')[0] }}</span>
            <span class="room-mini-name">{{ room.name }}</span>
          </div>
          <div class="room-mini-meta"><span>{{ room.memberCount }} 人</span><span>{{ room.activePostCount }} 帖</span></div>
        </div>
      </div>

      <div class="post-feed">
        <div v-for="post in filteredPosts" :key="post.id" class="post-card glass card-hover" @click="$router.push('/community/post/' + post.id)">
          <div class="post-card-left">
            <div class="post-avatar" :style="{background: post.author.color + '22', color: post.author.color}">{{ post.author.nickname[0] }}</div>
          </div>
          <div class="post-card-body">
            <div class="post-card-header">
              <span class="post-author">{{ post.author.nickname }}</span>
              <span class="post-level">{{ post.author.level }}</span>
              <span class="post-time">{{ post.createdAt?.replace('T', ' ') }}</span>
              <span v-if="post.isPinned" class="pin-badge">📌 置顶</span>
            </div>
            <h3 class="post-title">{{ post.title }}</h3>
            <p class="post-excerpt">{{ (post.content || '').slice(0, 80) }}{{ (post.content || '').length > 80 ? '...' : '' }}</p>
            <div class="post-card-footer">
              <div class="post-tags">
                <span v-for="t in post.tags" :key="t" class="tag tag-neutral">{{ t }}</span>
                <span v-if="post.stock" class="tag tag-gold">{{ post.stock }}</span>
              </div>
              <div class="post-stats">
                <span>{{ post.likes }} ♥</span>
                <span>{{ post.comments }} 评论</span>
                <span>{{ post.views }} 浏览</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <!-- Create Post Modal -->
      <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
        <div class="modal-card glass">
          <div class="modal-header">
            <h3>发起讨论</h3>
            <button class="modal-close" @click="showCreateModal = false">×</button>
          </div>
          <div class="modal-body">
            <div class="modal-field">
              <label>标题</label>
              <input type="text" v-model="newPost.title" placeholder="输入标题" maxlength="60" />
            </div>
            <div class="modal-field">
              <label>内容</label>
              <textarea v-model="newPost.content" placeholder="写下你的观点..." rows="5"></textarea>
            </div>
            <div class="modal-field">
              <label>类型</label>
              <select v-model="newPost.topic_type">
                <option value="STOCK">股票讨论</option>
                <option value="STRATEGY">策略分享</option>
                <option value="NEWS">资讯点评</option>
              </select>
            </div>
            <div class="modal-field">
              <label>关联股票（可选）</label>
              <input type="text" v-model="newPost.target_id" placeholder="股票代码，如 300750.SZ" />
            </div>
            <p v-if="postError" class="post-error">{{ postError }}</p>
          </div>
          <div class="modal-footer">
            <button class="btn-ghost" @click="showCreateModal = false">取消</button>
            <button class="btn-primary" @click="submitPost" :disabled="!newPost.title || !newPost.content || posting">
              {{ posting ? '发布中...' : '发布' }}
            </button>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { POSTS as MOCK_POSTS, USERS, STOCK_ROOMS } from "@/mock/community.js";
import { listCommunityTopics, createCommunityTopic } from "@/api/community.js";
import { useClientAuth } from "@/shared/auth/client-auth";

const { isLoggedIn } = useClientAuth();
const activeTab = ref('hot');
const allPosts = ref(MOCK_POSTS);
const showCreateModal = ref(false);
const posting = ref(false);
const postError = ref("");
const newPost = ref({ title: '', content: '', topic_type: 'STOCK', target_id: '' });

const tabs = [{ key: 'hot', label: '热门' }, { key: 'latest', label: '最新' }];

const filteredPosts = computed(() => {
  const sorted = [...allPosts.value];
  if (activeTab.value === 'hot') sorted.sort((a, b) => b.likes - a.likes);
  else sorted.sort((a, b) => new Date(b.createdAt) - new Date(a.createdAt));
  return sorted;
});

async function loadTopics() {
  try {
    const result = await listCommunityTopics({ page: 1, page_size: 20 });
    if (result?.items?.length) {
      allPosts.value = result.items.map((item) => ({
        id: item.id, title: item.title, content: item.summary || item.content || '',
        author: USERS[Math.floor(Math.random() * USERS.length)],
        stock: item.linked_target?.symbol || null,
        category: '讨论', createdAt: item.created_at || new Date().toISOString(),
        likes: item.like_count || 0, comments: item.comment_count || 0, views: (item.like_count || 0) * 10,
        isPinned: false, tags: ['社区'], isHot: false
      }));
    }
  } catch { /* use mock */ }
}

async function submitPost() {
  if (!newPost.value.title || !newPost.value.content) return;
  posting.value = true;
  postError.value = "";
  try {
    await createCommunityTopic({
      title: newPost.value.title,
      content: newPost.value.content,
      topic_type: newPost.value.topic_type,
      stance: "WATCH",
      reason_text: "AI 分析参考",
      risk_text: "市场有风险",
      target_type: "STOCK",
      target_id: newPost.value.target_id || newPost.value.topic_type
    });
    showCreateModal.value = false;
    newPost.value = { title: '', content: '', topic_type: 'STOCK', target_id: '' };
    await loadTopics();
  } catch (e) {
    postError.value = e?.message || "发布失败";
  } finally { posting.value = false; }
}

onMounted(loadTopics);
</script>

<style scoped>
.community-page { display: grid; gap: 20px; max-width: 1400px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; flex-wrap: wrap; gap: 10px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.section-stats { display: flex; gap: 12px; font-size: 12px; color: var(--text-muted); }
.community-tabs { display: flex; gap: 6px; margin-bottom: 16px; }
.com-tab { padding: 6px 16px; border-radius: var(--radius-full); font-size: 13px; color: var(--text-secondary); }
.com-tab.active { color: #000; background: var(--accent-gold); font-weight: 600; }
.rooms-row { display: flex; gap: 8px; margin-bottom: 20px; overflow-x: auto; padding-bottom: 4px; }
.room-mini { padding: 12px 16px; border-radius: var(--radius-md); cursor: pointer; flex-shrink: 0; }
.room-mini-top { display: flex; align-items: center; gap: 8px; margin-bottom: 6px; }
.room-mini-symbol { font-family: var(--font-mono); font-size: 12px; color: var(--text-secondary); }
.room-mini-name { font-size: 13px; font-weight: 600; white-space: nowrap; }
.room-mini-meta { display: flex; gap: 10px; font-size: 11px; color: var(--text-muted); }
.post-feed { display: grid; gap: 10px; }
.post-card { display: flex; gap: 14px; padding: 16px; border-radius: var(--radius-md); cursor: pointer; }
.post-card-left { flex-shrink: 0; }
.post-card-body { flex: 1; min-width: 0; }
.post-card-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; flex-wrap: wrap; }
.post-author { font-size: 13px; font-weight: 600; }
.post-level { font-size: 10px; color: var(--accent-gold); padding: 1px 6px; border-radius: 4px; background: var(--accent-gold-glow); }
.post-time { font-size: 11px; color: var(--text-muted); }
.pin-badge { font-size: 11px; }
.post-title { font-size: 16px; font-weight: 700; margin-bottom: 6px; line-height: 1.4; }
.post-excerpt { font-size: 13px; color: var(--text-secondary); line-height: 1.6; margin-bottom: 10px; }
.post-card-footer { display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px; }
.post-tags { display: flex; gap: 6px; flex-wrap: wrap; }
.post-stats { display: flex; gap: 10px; font-size: 11px; color: var(--text-muted); }
.section-stats { display: flex; align-items: center; gap: 10px; }
.create-post-btn { flex-shrink: 0; }
.btn-primary { padding: 8px 18px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 13px; font-weight: 600; border: none; cursor: pointer; }
.btn-ghost { padding: 8px 18px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 13px; color: var(--text-secondary); background: none; cursor: pointer; }
.modal-overlay { position: fixed; inset: 0; background: rgba(0,0,0,.6); display: flex; align-items: center; justify-content: center; z-index: 1000; }
.modal-card { width: 520px; max-width: 90vw; max-height: 80vh; overflow-y: auto; border-radius: var(--radius-xl); padding: 0; }
.modal-header { display: flex; align-items: center; justify-content: space-between; padding: 20px 24px 0; }
.modal-header h3 { font-size: 18px; font-weight: 700; }
.modal-close { font-size: 24px; color: var(--text-muted); }
.modal-body { padding: 20px 24px; display: grid; gap: 14px; }
.modal-field { display: grid; gap: 4px; }
.modal-field label { font-size: 13px; color: var(--text-secondary); }
.modal-field input, .modal-field textarea, .modal-field select { padding: 10px 12px; border-radius: var(--radius-sm); background: rgba(255,255,255,.04); border: 1px solid var(--border); color: var(--text-primary); font-size: 13px; font-family: inherit; }
.modal-field input:focus, .modal-field textarea:focus, .modal-field select:focus { border-color: var(--accent-gold); }
.modal-field select { appearance: auto; }
.post-error { color: var(--negative); font-size: 12px; }
.modal-footer { display: flex; justify-content: flex-end; gap: 10px; padding: 0 24px 20px; }
</style>
