<template>
  <div class="h5-news">
    <div class="h5-cat-scroll">
      <button v-for="cat in categoriesList" :key="cat" class="h5-cat-tab" :class="{active: activeCat === cat}" @click="selectCategory(cat)">{{ cat }}</button>
    </div>
    <div class="h5-news-list">
      <article v-for="article in filtered" :key="article.id" class="h5-news-card" @click="$router.push('/news/' + article.id)">
        <div class="h5-news-tag"><span class="tag" :class="catClass(article.category)">{{ article.category }}</span><span class="h5-news-time">{{ article.publishTime }}</span></div>
        <h3>{{ article.title }}</h3>
        <p class="h5-news-ai">AI: {{ article.aiSummary }}</p>
        <div class="h5-news-meta"><span>{{ article.source }}</span></div>
      </article>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { NEWS_CATEGORIES, NEWS_ARTICLES as MOCK_ARTICLES } from "@/mock/news.js";
import { listNewsArticles, listNewsCategories } from "@/api/news.js";

const activeCat = ref('全部');
const categoriesList = ref(['全部', ...NEWS_CATEGORIES.slice(1)]);
const articles = ref(MOCK_ARTICLES);
const categoryMap = ref({});
const categoryNameToIdMap = ref({});

const filtered = computed(() => {
  if (activeCat.value === '全部') return articles.value;
  return articles.value.filter(a => a.category === activeCat.value);
});

function catClass(cat) {
  const map = { '宏观经济': 'tag-blue', '市场动态': 'tag-green', '公司新闻': 'tag-gold', '行业研究': 'tag-blue', '政策解读': 'tag-red', 'AI解读': 'tag-gold' };
  return map[cat] || 'tag-neutral';
}

async function selectCategory(catName) {
  activeCat.value = catName;
  await loadNewsArticles();
}

async function loadCategories() {
  try {
    const res = await listNewsCategories();
    if (res?.items?.length) {
      const backendCats = res.items;
      categoriesList.value = ['全部', ...backendCats.map(c => c.name)];
      
      const idToName = {};
      const nameToId = {};
      backendCats.forEach(c => {
        idToName[c.id] = c.name;
        nameToId[c.name] = c.id;
      });
      categoryMap.value = idToName;
      categoryNameToIdMap.value = nameToId;
    }
  } catch (e) {
    console.error('Failed to load categories', e);
  }
}

async function loadNewsArticles() {
  try {
    const params = { page: 1, page_size: 20 };
    if (activeCat.value !== '全部') {
      const catId = categoryNameToIdMap.value[activeCat.value];
      if (catId) {
        params.category_id = catId;
      }
    }
    const result = await listNewsArticles(params);
    if (result?.items?.length) {
      const items = result.items;
      articles.value = items.map((item) => ({
        id: item.id, title: item.title, category: categoryMap.value[item.category_id] || item.category_name || '资讯',
        source: item.source || '资讯中心',
        publishTime: item.published_at ? item.published_at.slice(0, 10) : '-',
        aiSummary: (item.summary || '').slice(0, 100)
      }));
    } else {
      articles.value = [];
    }
  } catch (e) {
    // Fallback to mock
    if (activeCat.value === '全部') {
      articles.value = MOCK_ARTICLES;
    } else {
      articles.value = MOCK_ARTICLES.filter(a => a.category === activeCat.value);
    }
  }
}

async function init() {
  await loadCategories();
  await loadNewsArticles();
}

onMounted(init);
</script>

<style scoped>
.h5-news { display: grid; gap: 10px; }
.h5-cat-scroll { display: flex; gap: 6px; overflow-x: auto; }
.h5-cat-tab { padding: 5px 12px; border-radius: var(--radius-full); font-size: 12px; color: var(--text-secondary); white-space: nowrap; }
.h5-cat-tab.active { color: #000; background: var(--accent-gold); font-weight: 600; }
.h5-news-list { display: grid; gap: 8px; }
.h5-news-card { padding: 14px; background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-md); }
.h5-news-card:active { background: rgba(255,255,255,.03); }
.h5-news-tag { display: flex; align-items: center; justify-content: space-between; margin-bottom: 6px; }
.h5-news-time { font-size: 10px; color: var(--text-muted); }
.h5-news-card h3 { font-size: 14px; font-weight: 700; line-height: 1.4; margin-bottom: 6px; }
.h5-news-ai { font-size: 12px; color: var(--text-secondary); line-height: 1.5; padding-left: 8px; border-left: 2px solid var(--accent-gold); margin-bottom: 8px; }
.h5-news-meta { display: flex; justify-content: space-between; font-size: 11px; color: var(--text-muted); }
</style>
