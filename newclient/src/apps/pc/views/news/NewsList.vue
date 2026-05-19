<template>
  <div class="news-page">
    <section class="section fade-in-up">
      <div class="section-header">
        <h2 class="section-title">资讯中心</h2>
        <p class="section-subtitle">AI 为您聚合市场最新动态与深度解读</p>
      </div>
      <div class="news-categories">
        <button v-for="cat in NEWS_CATEGORIES" :key="cat" class="cat-tab" :class="{ active: activeCat === cat }" @click="activeCat = cat">{{ cat }}</button>
      </div>

      <!-- Featured -->
      <article v-if="featured" class="featured-news glass card-hover" @click="$router.push('/news/' + featured.id)">
        <div class="featured-tags"><span class="tag tag-gold">{{ featured.category }}</span><span v-if="featured.tags?.includes('重磅')" class="tag tag-red">重磅</span></div>
        <h2 class="featured-title">{{ featured.title }}</h2>
        <p class="featured-ai-summary">{{ featured.aiSummary }}</p>
        <div class="featured-meta">
          <span>{{ featured.source }}</span>
          <span>{{ featured.publishTime }}</span>
          <span>{{ (featured.views || 0).toLocaleString() }} 阅读</span>
        </div>
      </article>

      <!-- Article List -->
      <div class="news-list">
        <article v-for="article in filteredNews" :key="article.id" class="news-card glass card-hover" @click="$router.push('/news/' + article.id)">
          <div class="news-card-top">
            <span class="tag" :class="catClass(article.category)">{{ article.category }}</span>
            <span class="news-time">{{ article.publishTime }}</span>
          </div>
          <h3 class="news-title">{{ article.title }}</h3>
          <p class="news-ai-summary">AI: {{ article.aiSummary }}</p>
          <div class="news-card-footer">
            <div class="news-meta">
              <span>{{ article.source }}</span>
              <span>{{ (article.views || 0).toLocaleString() }} 阅读</span>
            </div>
          </div>
        </article>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { NEWS_CATEGORIES, NEWS_ARTICLES as MOCK_ARTICLES } from "@/mock/news.js";
import { listNewsArticles, listNewsCategories } from "@/api/news.js";

const activeCat = ref('全部');
const articles = ref(MOCK_ARTICLES);

const featured = computed(() => articles.value.find(a => a.isFeatured));
const filteredNews = computed(() => {
  let list = articles.value.filter(a => a.id !== featured.value?.id);
  if (activeCat.value !== '全部') list = list.filter(a => a.category === activeCat.value);
  return list;
});

function catClass(cat) {
  const map = { '宏观经济': 'tag-blue', '市场动态': 'tag-green', '公司新闻': 'tag-gold', '行业研究': 'tag-blue', '政策解读': 'tag-red', 'AI解读': 'tag-gold' };
  return map[cat] || 'tag-neutral';
}

async function loadNews() {
  try {
    const [catResult, articleResult] = await Promise.allSettled([listNewsCategories(), listNewsArticles({ page: 1, page_size: 20 })]);
    if (articleResult.status === 'fulfilled' && articleResult.value?.items?.length) {
      const items = articleResult.value.items;
      articles.value = items.map((item, i) => ({
        id: item.id, title: item.title, category: item.category_name || '资讯', source: item.source || '资讯中心',
        publishTime: item.published_at ? item.published_at.slice(0, 10) : '-',
        views: 0, likes: 0, aiSummary: (item.summary || '').slice(0, 100),
        impact: { direction: 'neutral', level: 'low', affectedStocks: [] },
        isFeatured: i === 0, tags: [], content: [item.content || item.summary || '暂无内容']
      }));
    }
  } catch { /* use mock */ }
}

onMounted(loadNews);
</script>

<style scoped>
.news-page { display: grid; gap: 20px; max-width: 1400px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { margin-bottom: 16px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }
.news-categories { display: flex; gap: 6px; margin-bottom: 20px; flex-wrap: wrap; }
.cat-tab { padding: 6px 14px; border-radius: var(--radius-full); font-size: 13px; color: var(--text-secondary); }
.cat-tab:hover { color: var(--text-primary); background: rgba(255,255,255,.04); }
.cat-tab.active { color: #000; background: var(--accent-gold); font-weight: 600; }
.featured-news { padding: 20px; border-radius: var(--radius-lg); cursor: pointer; margin-bottom: 16px; }
.featured-tags { display: flex; gap: 6px; margin-bottom: 10px; }
.featured-title { font-size: 22px; font-weight: 700; line-height: 1.35; margin-bottom: 10px; }
.featured-ai-summary { font-size: 13px; color: var(--text-secondary); line-height: 1.7; margin-bottom: 12px; }
.featured-meta { display: flex; align-items: center; gap: 16px; font-size: 12px; color: var(--text-muted); }
.news-list { display: grid; gap: 10px; }
.news-card { padding: 16px; border-radius: var(--radius-md); cursor: pointer; }
.news-card-top { display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px; }
.news-time { font-size: 11px; color: var(--text-muted); }
.news-title { font-size: 15px; font-weight: 700; line-height: 1.4; margin-bottom: 8px; }
.news-ai-summary { font-size: 12px; color: var(--text-secondary); line-height: 1.6; margin-bottom: 10px; padding-left: 10px; border-left: 2px solid var(--accent-gold); }
.news-card-footer { display: flex; align-items: center; justify-content: space-between; }
.news-meta { display: flex; gap: 12px; font-size: 11px; color: var(--text-muted); }
</style>
