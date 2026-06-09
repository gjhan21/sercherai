<template>
  <div class="news-detail-page" v-if="article">
    <button class="back-btn glass" @click="$router.push('/news')">← 返回资讯列表</button>

    <article class="section fade-in-up">
      <div class="detail-top">
        <div class="detail-tags">
          <span class="tag" :class="catTagClass">{{ article.category }}</span>
          <span v-if="article.tags" v-for="t in article.tags" :key="t" class="tag tag-neutral">{{ t }}</span>
        </div>
        <h1 class="detail-title">{{ article.title }}</h1>
        <div class="detail-meta">
          <span>{{ article.source }}</span>
          <span>{{ article.author }}</span>
          <span>{{ article.publishTime }}</span>
          <span>{{ (article.views || 0).toLocaleString() }} 阅读</span>
        </div>
      </div>

      <div class="ai-summary-panel glass">
        <div class="asp-header">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="18" height="18" style="color:var(--accent-gold)"><circle cx="12" cy="12" r="10"/><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83"/></svg>
          <span>AI 智能摘要</span>
        </div>
        <p class="asp-content">{{ article.aiSummary }}</p>
      </div>

      <!-- Restricted / VIP Locked state -->
      <div class="vip-lock-panel glass" v-if="article.is_locked">
        <div class="lock-icon-wrapper">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="48" height="48" class="lock-icon">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
            <path d="M7 11V7a5 5 0 0110 0v4"/>
          </svg>
        </div>
        <h3>此内容为 VIP 专属研报</h3>
        <p>本深度报告仅限 VIP 会员查看，包含详细数据透视及 PDF 附件下载。</p>
        <button class="btn-upgrade" @click="goVip">升级 VIP 查看完整内容</button>
      </div>

      <template v-else>
        <div class="detail-content">
          <p v-for="(para, i) in article.content" :key="i">{{ para }}</p>
        </div>

        <!-- Attachments Panel -->
        <div v-if="attachments && attachments.length > 0" class="attachments-panel glass">
          <h3>附件下载</h3>
          <div class="attachments-list">
            <div v-for="att in attachments" :key="att.id" class="attachment-item">
              <div class="att-info">
                <svg class="att-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="20" height="20">
                  <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
                  <polyline points="14 2 14 8 20 8"/>
                  <line x1="16" y1="13" x2="8" y2="13"/>
                  <line x1="16" y1="17" x2="8" y2="17"/>
                </svg>
                <div class="att-name-size">
                  <span class="att-name" :title="att.file_name">{{ att.file_name }}</span>
                  <span class="att-size">{{ formatSize(att.file_size) }}</span>
                </div>
              </div>
              <button class="btn-download" @click="downloadAttachment(att)">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
                  <path d="M21 15v4a2 2 0 01-2 2H5a2 2 0 01-2-2v-4M7 10l5 5 5-5M12 15V3"/>
                </svg>
                下载
              </button>
            </div>
          </div>
        </div>

        <div class="impact-panel glass">
          <h3>市场影响分析</h3>
          <div class="impact-header">
            <span class="impact-badge-large" :class="article.impact.direction">{{ impactLabel }}</span>
            <span class="impact-level">影响等级: {{ levelLabel }}</span>
          </div>
          <p class="impact-text">{{ article.impact.analysis }}</p>
          <div class="impact-stocks">
            <span>关联股票：</span>
            <button v-for="sym in article.impact.affectedStocks" :key="sym" class="stock-chip" @click="goIdentify(sym)">{{ sym }}</button>
          </div>
        </div>
      </template>
    </article>
  </div>

  <div v-else class="not-found">
    <p>{{ loading ? '加载中...' : '资讯未找到' }}</p>
    <button class="btn-primary" @click="$router.push('/news')">返回资讯列表</button>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getNewsById } from "@/mock/news.js";
import { getNewsArticleDetail, listNewsCategories, listNewsAttachments, getAttachmentSignedURL } from "@/api/news.js";

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const articleData = ref(null);
const attachments = ref([]);

const article = computed(() => articleData.value);
const catTagClass = computed(() => {
  const map = { '宏观经济': 'tag tag-blue', '市场动态': 'tag-green', '公司新闻': 'tag-gold', '行业研究': 'tag-blue', '政策解读': 'tag-red', 'AI解读': 'tag-gold' };
  return map[article.value?.category] || 'tag-neutral';
});
const impactLabel = computed(() => article.value?.impact?.direction === 'positive' ? '利好' : article.value?.impact?.direction === 'negative' ? '利空' : '中性');
const levelLabel = computed(() => article.value?.impact?.level === 'high' ? '高' : article.value?.impact?.level === 'medium' ? '中' : '低');

function goIdentify(symbol) { router.push('/identify/' + symbol); }

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
  router.push('/user/vip');
}

async function loadArticle() {
  const id = route.params.id;
  if (!id) return;
  loading.value = true;
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
        publishTime: result.published_at ? result.published_at.slice(0, 16).replace('T', ' ') : '-',
        views: 0, likes: 0, tags: [],
        aiSummary: (result.summary || '').slice(0, 200),
        content: [result.content || result.summary || '暂无内容'],
        impact: { direction: 'neutral', level: 'low', affectedStocks: [], analysis: 'AI 暂未生成影响分析' },
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
      
      loading.value = false;
      return;
    }
  } catch { /* fall through */ }
  // Mock fallback
  const mock = getNewsById(Number(id));
  if (mock) articleData.value = { ...mock };
  loading.value = false;
}

onMounted(loadArticle);
</script>

<style scoped>
.news-detail-page { max-width: 1000px; display: grid; gap: 16px; }
.back-btn { display: inline-flex; padding: 8px 16px; border-radius: var(--radius-full); font-size: 13px; color: var(--text-secondary); cursor: pointer; width: auto; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 28px; }
.detail-top { margin-bottom: 20px; }
.detail-tags { display: flex; gap: 6px; margin-bottom: 12px; flex-wrap: wrap; }
.detail-title { font-size: 26px; font-weight: 800; line-height: 1.3; margin-bottom: 12px; }
.detail-meta { display: flex; gap: 16px; font-size: 12px; color: var(--text-muted); }
.ai-summary-panel { padding: 16px; border-radius: var(--radius-md); margin-bottom: 20px; }
.asp-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; font-weight: 600; font-size: 14px; }
.asp-content { font-size: 14px; color: var(--text-secondary); line-height: 1.7; }
.detail-content { margin-bottom: 20px; }
.detail-content p { font-size: 15px; line-height: 1.9; color: var(--text-secondary); margin-bottom: 16px; }
.impact-panel { padding: 20px; border-radius: var(--radius-md); }
.impact-panel h3 { font-size: 16px; font-weight: 700; margin-bottom: 12px; }
.impact-header { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.impact-badge-large { padding: 4px 12px; border-radius: 4px; font-size: 14px; font-weight: 700; }
.impact-badge-large.positive { background: var(--positive-bg); color: var(--positive); }
.impact-badge-large.negative { background: var(--negative-bg); color: var(--negative); }
.impact-badge-large.neutral { background: rgba(139,147,176,.1); color: var(--text-secondary); }
.impact-level { font-size: 12px; color: var(--text-secondary); }
.impact-text { font-size: 14px; color: var(--text-secondary); line-height: 1.7; margin-bottom: 12px; }
.impact-stocks { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; font-size: 13px; color: var(--text-secondary); }
.stock-chip { padding: 4px 10px; border-radius: 4px; background: var(--accent-gold-glow); color: var(--accent-gold); font-size: 12px; font-weight: 600; cursor: pointer; }
.stock-chip:hover { background: rgba(240,185,11,.2); }
.btn-primary { padding: 8px 18px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 13px; font-weight: 600; cursor: pointer; border: none; }
.not-found { text-align: center; padding: 60px 20px; }
.not-found p { font-size: 16px; color: var(--text-secondary); margin-bottom: 16px; }

/* VIP lock panel */
.vip-lock-panel {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  text-align: center;
  border: 1px dashed var(--accent-gold);
  border-radius: var(--radius-lg);
  background: rgba(240,185,11,.03);
  margin-top: 20px;
}
.lock-icon-wrapper {
  background: rgba(240,185,11,.1);
  color: var(--accent-gold);
  padding: 16px;
  border-radius: 50%;
  margin-bottom: 16px;
}
.lock-icon {
  display: block;
}
.vip-lock-panel h3 {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 8px;
  color: var(--text-primary);
}
.vip-lock-panel p {
  font-size: 14px;
  color: var(--text-secondary);
  max-width: 400px;
  line-height: 1.6;
  margin-bottom: 20px;
}
.btn-upgrade {
  padding: 10px 24px;
  border-radius: var(--radius-full);
  background: linear-gradient(135deg, var(--accent-gold), #d4a373);
  color: #000;
  font-weight: 700;
  border: none;
  cursor: pointer;
  box-shadow: 0 4px 15px rgba(240,185,11,0.2);
  transition: all 0.2s ease;
}
.btn-upgrade:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(240,185,11,0.3);
}

/* Attachments Panel */
.attachments-panel {
  padding: 20px;
  border-radius: var(--radius-md);
  margin-bottom: 20px;
  background: rgba(255,255,255,.01);
  border: 1px solid var(--border);
}
.attachments-panel h3 {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 14px;
}
.attachments-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.attachment-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: rgba(255,255,255,.02);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  transition: background 0.2s;
}
.attachment-item:hover {
  background: rgba(255,255,255,.04);
}
.att-info {
  display: flex;
  align-items: center;
  gap: 12px;
  overflow: hidden;
}
.att-icon {
  color: var(--text-muted);
  flex-shrink: 0;
}
.att-name-size {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.att-name {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  text-overflow: ellipsis;
  white-space: nowrap;
  overflow: hidden;
}
.att-size {
  font-size: 11px;
  color: var(--text-muted);
}
.btn-download {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border-radius: var(--radius-full);
  background: rgba(255,255,255,.04);
  color: var(--text-primary);
  border: 1px solid var(--border);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.btn-download:hover {
  background: var(--accent-gold);
  color: #000;
  border-color: var(--accent-gold);
}
</style>
