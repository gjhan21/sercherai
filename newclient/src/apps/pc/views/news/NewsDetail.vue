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

      <div class="detail-content">
        <p v-for="(para, i) in article.content" :key="i">{{ para }}</p>
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
import { getNewsArticleDetail } from "@/api/news.js";

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const articleData = ref(null);

const article = computed(() => articleData.value);
const catTagClass = computed(() => {
  const map = { '宏观经济': 'tag tag-blue', '市场动态': 'tag-green', '公司新闻': 'tag-gold', '行业研究': 'tag-blue', '政策解读': 'tag-red', 'AI解读': 'tag-gold' };
  return map[article.value?.category] || 'tag-neutral';
});
const impactLabel = computed(() => article.value?.impact?.direction === 'positive' ? '利好' : article.value?.impact?.direction === 'negative' ? '利空' : '中性');
const levelLabel = computed(() => article.value?.impact?.level === 'high' ? '高' : article.value?.impact?.level === 'medium' ? '中' : '低');

function goIdentify(symbol) { router.push('/identify/' + symbol); }

async function loadArticle() {
  const id = route.params.id;
  if (!id) return;
  loading.value = true;
  try {
    const result = await getNewsArticleDetail(id);
    if (result) {
      articleData.value = {
        id: result.id, title: result.title, category: result.category_name || '资讯',
        source: result.source || '资讯中心', author: 'AI资讯',
        publishTime: result.published_at ? result.published_at.slice(0, 16).replace('T', ' ') : '-',
        views: 0, likes: 0, tags: [],
        aiSummary: (result.summary || '').slice(0, 200),
        content: [result.content || result.summary || '暂无内容'],
        impact: { direction: 'neutral', level: 'low', affectedStocks: [], analysis: 'AI 暂未生成影响分析' }
      };
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
</style>
