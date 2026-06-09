<template>
  <div class="h5-news-detail">
    <div v-if="article">
      <div class="h5-article-header">
        <div class="h5-article-top"><span class="tag tag-gold">{{ article.category }}</span><span>{{ article.publishTime }}</span></div>
        <h1>{{ article.title }}</h1>
        <div class="h5-article-source">{{ article.source }} · {{ article.author }}</div>
      </div>
      
      <div class="h5-ai-summary" v-if="article.aiSummary">
        <span class="h5-ai-label">AI 摘要</span>
        <p>{{ article.aiSummary }}</p>
      </div>

      <!-- VIP lock panel for H5 -->
      <div class="h5-vip-lock-panel" v-if="article.is_locked">
        <div class="h5-lock-icon-wrapper">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="36" height="36" class="h5-lock-icon">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
            <path d="M7 11V7a5 5 0 0110 0v4"/>
          </svg>
        </div>
        <h3>VIP 专属研报</h3>
        <p>本报告仅限 VIP 会员查看，包含详细内容及附件下载。</p>
        <button class="h5-btn-upgrade" @click="goVip">升级 VIP 查看</button>
      </div>

      <template v-else>
        <div class="h5-article-body">
          <p v-for="(p, i) in (article.content || [article.summary || '暂无内容'])" :key="i">{{ p }}</p>
        </div>

        <!-- H5 Attachments Panel -->
        <div v-if="attachments && attachments.length > 0" class="h5-attachments-panel">
          <div class="h5-attachments-title">附件下载</div>
          <div class="h5-attachments-list">
            <div v-for="att in attachments" :key="att.id" class="h5-attachment-item">
              <div class="h5-att-info">
                <svg class="h5-att-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="16" height="16">
                  <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                </svg>
                <span class="h5-att-name" :title="att.file_name">{{ att.file_name }}</span>
              </div>
              <button class="h5-btn-download" @click="downloadAttachment(att)">下载</button>
            </div>
          </div>
        </div>
      </template>
    </div>
    <div v-else class="h5-empty"><p>加载中...</p></div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getNewsById } from "@/mock/news.js";
import { getNewsArticleDetail, listNewsCategories, listNewsAttachments, getAttachmentSignedURL } from "@/api/news.js";

const route = useRoute();
const router = useRouter();
const articleData = ref(null);
const attachments = ref([]);
const article = computed(() => articleData.value);

function formatSize(bytes) {
  if (!bytes) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

async function downloadAttachment(att) {
  try {
    const res = await getAttachmentSignedURL(att.id);
    if (res && res.signed_url) {
      window.open(res.signed_url, '_blank');
    } else if (att.file_url) {
      window.open(att.file_url, '_blank');
    }
  } catch (e) {
    if (att.file_url) {
      window.open(att.file_url, '_blank');
    } else {
      alert('下载失败，请确认您已登录且会员等级足够。');
    }
  }
}

function goVip() {
  router.push('/profile/vip');
}

async function loadArticle() {
  const id = route.params.id;
  if (!id) return;
  attachments.value = [];
  try {
    const [result, categoriesRes] = await Promise.all([
      getNewsArticleDetail(id),
      listNewsCategories().catch(() => null)
    ]);
    if (result) {
      let categoryName = '资讯';
      if (categoriesRes && categoriesRes.items) {
        const cat = categoriesRes.items.find(c => c.id === result.category_id);
        if (cat) {
          categoryName = cat.name;
        }
      }
      articleData.value = {
        id: result.id, title: result.title, category: categoryName,
        source: result.source || '资讯中心', author: 'AI资讯',
        publishTime: result.published_at ? result.published_at.slice(0, 10) : '-',
        aiSummary: (result.summary || '').slice(0, 200),
        content: [result.content || result.summary || '暂无内容'],
        is_locked: result.is_locked
      };

      if (!result.is_locked) {
        try {
          const attRes = await listNewsAttachments(id);
          if (attRes && attRes.items) {
            attachments.value = attRes.items;
          }
        } catch (e) {
          console.error("Failed to load attachments:", e);
        }
      }
      return;
    }
  } catch { /* fall through */ }
  const mock = getNewsById(Number(id));
  if (mock) articleData.value = { ...mock };
}

onMounted(loadArticle);
</script>

<style scoped>
.h5-news-detail { display: grid; gap: 14px; }
.h5-article-header { padding: 16px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); }
.h5-article-top { display: flex; align-items: center; gap: 10px; margin-bottom: 8px; font-size: 11px; color: var(--text-muted); }
.h5-article-header h1 { font-size: 18px; font-weight: 800; line-height: 1.35; margin-bottom: 8px; }
.h5-article-source { font-size: 12px; color: var(--text-secondary); }
.h5-ai-summary { padding: 14px; border-radius: var(--radius-md); background: rgba(240,185,11,.05); border: 1px solid var(--border-gold); }
.h5-ai-label { display: inline-block; padding: 2px 8px; border-radius: 4px; background: var(--accent-gold); color: #000; font-size: 10px; font-weight: 700; margin-bottom: 8px; }
.h5-ai-summary p { font-size: 12px; color: var(--text-secondary); line-height: 1.6; }
.h5-article-body { padding: 0 4px; }
.h5-article-body p { font-size: 14px; line-height: 1.8; color: var(--text-secondary); margin-bottom: 14px; }
.h5-empty { text-align: center; padding: 40px; color: var(--text-secondary); }

/* VIP Lock for H5 */
.h5-vip-lock-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 30px 20px;
  text-align: center;
  border: 1px dashed var(--accent-gold);
  border-radius: var(--radius-lg);
  background: rgba(240,185,11,.03);
  margin: 10px 4px;
}
.h5-lock-icon-wrapper {
  background: rgba(240,185,11,.1);
  color: var(--accent-gold);
  padding: 12px;
  border-radius: 50%;
  margin-bottom: 12px;
}
.h5-lock-icon {
  display: block;
}
.h5-vip-lock-panel h3 {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 6px;
  color: var(--text-primary);
}
.h5-vip-lock-panel p {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.5;
  margin-bottom: 16px;
  max-width: 280px;
}
.h5-btn-upgrade {
  padding: 8px 20px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, var(--accent-gold), #d4a373);
  color: #000;
  font-weight: 700;
  font-size: 13px;
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 10px rgba(240,185,11,0.2);
}

/* Attachments for H5 */
.h5-attachments-panel {
  margin: 20px 4px;
  padding: 14px;
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
}
.h5-attachments-title {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 10px;
}
.h5-attachments-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.h5-attachment-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px;
  background: rgba(255,255,255,.01);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  gap: 10px;
}
.h5-att-info {
  display: flex;
  align-items: center;
  gap: 8px;
  overflow: hidden;
  flex: 1;
}
.h5-att-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}
.h5-att-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
}
.h5-btn-download {
  padding: 4px 10px;
  border-radius: var(--radius-full);
  background: rgba(255,255,255,.04);
  color: var(--text-primary);
  border: 1px solid var(--border);
  font-size: 11px;
  font-weight: 600;
  flex-shrink: 0;
}
.h5-btn-download:active {
  background: var(--accent-gold);
  color: #000;
  border-color: var(--accent-gold);
}
</style>
