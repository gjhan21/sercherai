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
      <div class="h5-article-body">
        <p v-for="(p, i) in (article.content || [article.summary || '暂无内容'])" :key="i">{{ p }}</p>
      </div>
    </div>
    <div v-else class="h5-empty"><p>加载中...</p></div>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute } from "vue-router";
import { getNewsById } from "@/mock/news.js";
import { getNewsArticleDetail } from "@/api/news.js";

const route = useRoute();
const articleData = ref(null);
const article = computed(() => articleData.value);

async function loadArticle() {
  const id = route.params.id;
  if (!id) return;
  try {
    const result = await getNewsArticleDetail(id);
    if (result) {
      articleData.value = {
        id: result.id, title: result.title, category: result.category_name || '资讯',
        source: result.source || '资讯中心', author: 'AI资讯',
        publishTime: result.published_at ? result.published_at.slice(0, 10) : '-',
        aiSummary: (result.summary || '').slice(0, 200),
        content: [result.content || result.summary || '暂无内容']
      };
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
</style>
