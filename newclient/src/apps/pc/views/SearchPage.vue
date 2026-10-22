<template>
  <div class="search-page">
    <section class="section fade-in-up">
      <!-- Search Box Section -->
      <div class="search-box-container">
        <h2 class="search-page-title">智能搜索</h2>
        <p class="search-page-subtitle">探索全市场 AI 推荐标的、期货策略以及专业研报资讯</p>
        
        <div class="search-input-row">
          <div class="search-input-wrap">
            <svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" width="20" height="20">
              <circle cx="11" cy="11" r="7"/>
              <path d="M16.5 16.5L21 21"/>
            </svg>
            <input 
              type="text" 
              v-model="keyword" 
              placeholder="输入股票代码/名称、策略合约、资讯关键字..." 
              @keyup.enter="doSearch" 
              class="premium-input"
            />
            <button v-if="keyword" class="clear-btn" @click="clearKeyword">×</button>
          </div>
          <button class="btn-primary search-btn" @click="doSearch">
            <span>智能搜索</span>
          </button>
        </div>
        
        <!-- Trending & History (Only shown when not loading & no searched state OR search input is clean) -->
        <div v-if="!searched && !loading" class="search-tags-section">
          <div class="tags-group">
            <span class="tags-label">🔥 热门搜索</span>
            <div class="tags-list">
              <button 
                v-for="tag in trendingSearches" 
                :key="tag" 
                class="tag-capsule" 
                @click="clickTag(tag)"
              >
                {{ tag }}
              </button>
            </div>
          </div>
          
          <div v-if="recentSearches.length" class="tags-group">
            <div class="tags-header-row">
              <span class="tags-label">🕒 最近搜索</span>
              <button class="clear-history-btn" @click="clearRecentSearches">清除历史</button>
            </div>
            <div class="tags-list">
              <button 
                v-for="tag in recentSearches" 
                :key="tag" 
                class="tag-capsule history-tag" 
                @click="clickTag(tag)"
              >
                {{ tag }}
              </button>
            </div>
          </div>
        </div>

        <p class="search-results-hint" v-if="searched && !loading && totalResults > 0">
          为您找到 <span class="text-gold">{{ totalResults }}</span> 条相关结果
        </p>
        <p class="search-results-hint error-hint" v-else-if="searched && !loading">
          ⚠️ 未能找到与 “{{ lastQuery }}” 相关的结果，请尝试其他关键词
        </p>
      </div>

      <!-- Main Content Area -->
      
      <!-- 1. Default Exploration (Empty State) -->
      <div v-if="!searched && !loading" class="exploration-area fade-in-up-delay-2">
        <h3 class="exploration-title">✨ AI 智能探索</h3>
        <div class="exploration-grid">
          <!-- Trending Stocks Card -->
          <div class="explore-card glass card-hover">
            <div class="card-header">
              <span class="header-icon">📈</span>
              <h4>今日高分标的</h4>
            </div>
            <div class="trending-list">
              <div 
                v-for="stock in mockTrendingStocks" 
                :key="stock.symbol" 
                class="trending-stock-item"
                @click="clickStock(stock.symbol)"
              >
                <div class="stock-meta">
                  <span class="stock-code">{{ stock.symbol }}</span>
                  <span class="stock-name">{{ stock.name }}</span>
                </div>
                <div class="stock-stats">
                  <span class="stock-change" :class="stock.change >= 0 ? 'up' : 'down'">
                    {{ stock.change >= 0 ? '+' : '' }}{{ stock.change }}%
                  </span>
                  <span class="stock-score-badge">评分 {{ stock.score }}</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Featured Reports Card -->
          <div class="explore-card glass card-hover">
            <div class="card-header">
              <span class="header-icon">📑</span>
              <h4>热门研报资讯</h4>
            </div>
            <div class="trending-list">
              <div 
                v-for="report in mockFeaturedReports" 
                :key="report.id" 
                class="trending-report-item"
                @click="$router.push('/news')"
              >
                <div class="report-top">
                  <span class="tag" :class="report.tag === '深度' ? 'tag-gold' : 'tag-blue'">{{ report.tag }}</span>
                  <span class="report-time">{{ report.time }}</span>
                </div>
                <h5 class="report-title">{{ report.title }}</h5>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 2. Loading State: CSS Shimmer Skeletons -->
      <div v-if="loading" class="loading-skeletons">
        <div class="skeleton-title"></div>
        <div class="skeleton-tabs">
          <div class="skeleton-tab"></div>
          <div class="skeleton-tab"></div>
          <div class="skeleton-tab"></div>
        </div>
        <div class="skeleton-cards">
          <div v-for="n in 3" :key="n" class="skeleton-card glass">
            <div class="skeleton-line-header"></div>
            <div class="skeleton-line-body"></div>
            <div class="skeleton-line-footer"></div>
          </div>
        </div>
      </div>

      <!-- 3. Search Results State -->
      <div v-if="searched && !loading" class="search-content-area">
        <!-- Tabs Header -->
        <div class="search-tabs" v-if="groups.length">
          <button 
            v-for="g in groups" 
            :key="g.key" 
            class="search-tab" 
            :class="{ active: activeTab === g.key }" 
            @click="activeTab = g.key"
          >
            {{ g.title }} <span class="tab-count">{{ g.total }}</span>
          </button>
        </div>

        <!-- Group Listing -->
        <div class="search-results">
          <div v-for="g in filteredGroups" :key="g.key" class="result-group">
            <div class="result-group-header">
              <span class="group-dot"></span>
              <span>{{ g.title }}</span>
            </div>
            
            <div class="result-cards-grid">
              <!-- Item Card -->
              <div 
                v-for="item in g.items" 
                :key="item.id" 
                class="result-card glass card-hover" 
                @click="goItem(g.key, item)"
              >
                <!-- Stock Result Card Template -->
                <template v-if="g.key === 'stocks'">
                  <div class="card-top">
                    <div class="title-wrap">
                      <span class="card-main-title">{{ item.symbol }}</span>
                      <span class="card-subtitle">{{ item.name }}</span>
                    </div>
                    <span class="score-badge">★ AI 推荐分 {{ item.score }}</span>
                  </div>
                  <p class="card-desc">{{ item.summary || 'AI 精选多维选股模型今日推荐标的。' }}</p>
                  <div class="card-footer">
                    <span class="footer-tag" :class="'risk-' + item.riskClass">{{ item.risk }}</span>
                    <span v-for="tag in item.tags" :key="tag" class="footer-tag tag-blue">{{ tag }}</span>
                  </div>
                </template>

                <!-- Strategy Result Card Template -->
                <template v-else-if="g.key === 'strategies'">
                  <div class="card-top">
                    <div class="title-wrap">
                      <span class="card-main-title">{{ item.title }}</span>
                      <span class="card-subtitle">期货主力策略</span>
                    </div>
                    <span class="direction-badge" :class="item.direction === 'LONG' ? 'direction-long' : 'direction-short'">
                      {{ item.direction === 'LONG' ? '做多 BUY' : '做空 SELL' }}
                    </span>
                  </div>
                  <p class="card-desc">{{ item.summary || '基于多智能体决策系统的期货趋势跟踪策略模型。' }}</p>
                  <div class="card-footer">
                    <span class="footer-tag tag-neutral">最新信号</span>
                    <span class="footer-tag tag-gold">高胜率型</span>
                  </div>
                </template>

                <!-- News Result Card Template -->
                <template v-else>
                  <div class="card-top">
                    <h4 class="card-news-title">{{ item.title }}</h4>
                  </div>
                  <p class="card-desc">{{ item.summary }}...</p>
                  <div class="card-footer">
                    <span class="news-meta-item">🔒 {{ item.meta === 'PUBLIC' ? '公开免费' : 'VIP 专享' }}</span>
                    <span class="news-meta-item">📰 资讯中心</span>
                  </div>
                </template>
              </div>
            </div>
            
            <div v-if="!g.items.length" class="result-empty glass">
              <span class="empty-icon">📂</span>
              <p>{{ g.emptyText || '暂无该类目相关结果' }}</p>
            </div>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { searchGlobalPublic } from "@/api/search.js";

const route = useRoute();
const router = useRouter();
const keyword = ref("");
const lastQuery = ref("");
const loading = ref(false);
const searched = ref(false);
const result = ref(null);
const activeTab = ref("all");

// Trending and Recent Search Logic
const trendingSearches = ref(["宁德时代", "高股息", "AI智能研判", "深度推演", "半导体"]);
const recentSearches = ref([]);

const mockTrendingStocks = [
  { symbol: '300750.SZ', name: '宁德时代', score: 92, change: 3.45 },
  { symbol: '600941.SH', name: '中国移动', score: 88, change: 1.82 },
  { symbol: '002594.SZ', name: '比亚迪', score: 80, change: 2.18 }
];

const mockFeaturedReports = [
  { id: 'na_df_1', title: '半导体行业深度：周期筑底，AI芯片国产替代进程加速', tag: '深度', time: '06-10' },
  { id: 'na_df_2', title: '智能驾驶行业专题：L3级别推演与自动驾驶落地展望', tag: '解读', time: '06-09' }
];

const totalResults = computed(() => searchGroups.value.reduce((s, g) => s + g.total, 0));

const searchGroups = computed(() => {
  const r = result.value;
  if (!r) return [];
  return [
    { 
      key: 'stocks', 
      title: '股票推荐', 
      total: r.stocks?.total || 0, 
      emptyText: '未找到相关股票推荐',
      items: (r.stocks?.items || []).map(i => {
        const riskVal = i.risk_level || '中等';
        const riskClassMap = { '高': 'high', '中': 'medium', '低': 'low', '中等': 'medium' };
        return { 
          id: i.id || i.symbol, 
          symbol: i.symbol || '',
          name: i.name || '',
          title: `${i.symbol || ''} ${i.name || ''}`.trim(), 
          summary: i.reason_summary || '多因子量化模型精选标的', 
          score: i.score || 80,
          risk: riskVal,
          riskClass: riskClassMap[riskVal] || 'medium',
          tags: i.score >= 90 ? ['重点关注', '低估值'] : ['稳健配置']
        };
      }) 
    },
    { 
      key: 'strategies', 
      title: '期货策略', 
      total: r.strategies?.total || 0, 
      emptyText: '未找到相关期货策略',
      items: (r.strategies?.items || []).map(i => ({ 
        id: i.id, 
        title: i.contract || i.name || '期货策略', 
        summary: i.reason_summary || '多智能体自适应趋势追踪期货策略模型', 
        direction: i.direction || 'LONG',
        meta: i.direction || 'LONG'
      })) 
    },
    { 
      key: 'news', 
      title: '资讯研报', 
      total: r.news?.total || 0, 
      emptyText: '未找到相关研报资讯',
      items: (r.news?.items || []).map(i => ({ 
        id: i.id, 
        title: i.title || '', 
        summary: (i.summary || i.content || '').slice(0, 100), 
        meta: i.visibility || 'PUBLIC' 
      })) 
    }
  ];
});

const groups = computed(() => {
  if (totalResults.value === 0) return [];
  return [
    { key: "all", title: "全部", total: totalResults.value },
    ...searchGroups.value.filter(g => g.total > 0)
  ];
});

const filteredGroups = computed(() => {
  if (activeTab.value === 'all') return searchGroups.value.filter(g => g.total > 0);
  return searchGroups.value.filter(g => g.key === activeTab.value);
});

onMounted(() => {
  loadRecentSearches();
  if (route.query.q) { 
    keyword.value = route.query.q; 
    doSearch(); 
  }
});

async function doSearch() {
  const q = keyword.value.trim();
  if (q.length < 2) return;
  
  loading.value = true; 
  searched.value = true; 
  activeTab.value = 'all';
  lastQuery.value = q;
  saveSearchKeyword(q);

  try {
    const r = await searchGlobalPublic({ keyword: q, mode: 'full', limit: 10 });
    result.value = r;
  } catch { 
    result.value = null; 
  } finally { 
    loading.value = false; 
  }
}

function clearKeyword() {
  keyword.value = "";
}

function clickTag(tag) {
  keyword.value = tag;
  doSearch();
}

function clickStock(symbol) {
  router.push({ path: '/identify', query: { q: symbol } });
}

function loadRecentSearches() {
  try {
    const list = JSON.parse(localStorage.getItem("recent_searches") || "[]");
    recentSearches.value = list.slice(0, 5);
  } catch {
    recentSearches.value = [];
  }
}

function saveSearchKeyword(q) {
  if (!q) return;
  const list = [q, ...recentSearches.value.filter(k => k !== q)].slice(0, 5);
  recentSearches.value = list;
  localStorage.setItem("recent_searches", JSON.stringify(list));
}

function clearRecentSearches() {
  recentSearches.value = [];
  localStorage.removeItem("recent_searches");
}

function goItem(key, item) {
  if (key === 'stocks') {
    router.push({ path: '/identify/' + item.symbol });
  } else if (key === 'strategies') {
    router.push({ path: '/identify', query: { q: item.title } });
  } else {
    router.push('/news/' + item.id);
  }
}
</script>

<style scoped>
.search-page { 
  max-width: 1400px; 
  margin: 0 auto;
}

.section { 
  background: var(--bg-card); 
  border: 1px solid var(--border); 
  border-radius: var(--radius-lg); 
  padding: 32px; 
  box-shadow: 0 8px 32px rgba(0,0,0,0.15);
}

.search-box-container {
  margin-bottom: 32px;
}

.search-page-title {
  font-size: 24px;
  font-weight: 800;
  margin-bottom: 6px;
  background: linear-gradient(135deg, var(--accent-gold), var(--accent-gold-dim));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.search-page-subtitle {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 24px;
}

.search-input-row { 
  display: flex; 
  gap: 12px; 
  margin-bottom: 16px; 
}

.search-input-wrap { 
  flex: 1; 
  display: flex; 
  align-items: center; 
  gap: 10px; 
  padding: 12px 18px; 
  border-radius: var(--radius-full); 
  background: rgba(255, 255, 255, .03); 
  border: 1px solid var(--border); 
  box-shadow: inset 0 2px 4px rgba(0,0,0,0.2);
  transition: all 0.3s ease;
  position: relative;
}

.search-input-wrap:focus-within { 
  border-color: var(--accent-gold); 
  box-shadow: 0 0 12px var(--accent-gold-glow), inset 0 1px 2px rgba(0,0,0,0.1);
  background: rgba(255, 255, 255, .05);
}

.search-icon { 
  flex-shrink: 0; 
  color: var(--text-secondary); 
  transition: color 0.3s;
}

.search-input-wrap:focus-within .search-icon {
  color: var(--accent-gold);
}

.premium-input { 
  flex: 1; 
  background: none; 
  border: none;
  color: var(--text-primary); 
  font-size: 15px; 
  outline: none;
}

.clear-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  font-size: 18px;
  cursor: pointer;
  padding: 0 4px;
}

.clear-btn:hover {
  color: var(--text-primary);
}

.search-btn { 
  padding: 0 28px; 
  border-radius: var(--radius-full); 
  background: linear-gradient(135deg, var(--accent-gold), var(--accent-gold-dim)); 
  color: #000; 
  font-size: 14px; 
  font-weight: 700; 
  border: none; 
  cursor: pointer; 
  transition: all 0.3s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.search-btn:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(240, 185, 11, 0.3);
}

.search-results-hint { 
  font-size: 13px; 
  color: var(--text-muted); 
  margin-top: 8px;
}

.error-hint {
  color: var(--negative);
}

.search-tags-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-top: 20px;
  padding-top: 12px;
  border-top: 1px dashed var(--border);
}

.tags-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tags-header-row {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.clear-history-btn {
  font-size: 11px;
  color: var(--text-muted);
  background: none;
  border: none;
  cursor: pointer;
  text-decoration: underline;
  text-align: left;
}

.clear-history-btn:hover {
  color: var(--negative);
}

.tags-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  white-space: nowrap;
  min-width: 65px;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-capsule {
  padding: 6px 14px;
  border-radius: var(--radius-full);
  font-size: 12px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s ease;
}

.tag-capsule:hover {
  background: var(--accent-gold-glow);
  color: var(--accent-gold);
  border-color: var(--accent-gold);
}

.history-tag:hover {
  background: rgba(255,255,255,0.08);
  color: var(--text-primary);
  border-color: var(--border-light);
}

.exploration-area {
  margin-top: 24px;
  animation: fadeInUp 0.6s ease both;
}

.exploration-title {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 16px;
  color: var(--text-primary);
}

.exploration-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.explore-card {
  border-radius: var(--radius-lg);
  padding: 20px;
  border: 1px solid var(--border);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 16px;
  border-bottom: 1px solid rgba(255,255,255,0.06);
  padding-bottom: 10px;
}

.header-icon {
  font-size: 18px;
}

.card-header h4 {
  font-size: 15px;
  font-weight: 700;
}

.trending-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.trending-stock-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  background: rgba(255,255,255,0.02);
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.2s;
}

.trending-stock-item:hover {
  background: rgba(255,255,255,0.05);
  border-color: var(--border-light);
}

.stock-meta {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stock-code {
  font-family: var(--font-mono);
  font-weight: 700;
  font-size: 14px;
}

.stock-name {
  font-size: 12px;
  color: var(--text-secondary);
}

.stock-stats {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 4px;
}

.stock-change.up {
  color: var(--positive);
  font-size: 13px;
  font-weight: 600;
}

.stock-change.down {
  color: var(--negative);
  font-size: 13px;
  font-weight: 600;
}

.stock-score-badge {
  font-size: 10px;
  padding: 2px 6px;
  background: var(--accent-gold-glow);
  color: var(--accent-gold);
  border-radius: 4px;
}

.trending-report-item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 12px;
  border-radius: var(--radius-md);
  background: rgba(255,255,255,0.02);
  cursor: pointer;
  transition: all 0.2s;
}

.trending-report-item:hover {
  background: rgba(255,255,255,0.05);
}

.report-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.report-time {
  font-size: 11px;
  color: var(--text-muted);
}

.report-title {
  font-size: 13px;
  font-weight: 600;
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.loading-skeletons {
  margin-top: 24px;
}

.skeleton-title {
  height: 20px;
  width: 150px;
  background: linear-gradient(90deg, rgba(255,255,255,0.03) 25%, rgba(255,255,255,0.08) 50%, rgba(255,255,255,0.03) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
  margin-bottom: 20px;
}

.skeleton-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 24px;
}

.skeleton-tab {
  height: 32px;
  width: 80px;
  border-radius: var(--radius-full);
  background: linear-gradient(90deg, rgba(255,255,255,0.03) 25%, rgba(255,255,255,0.08) 50%, rgba(255,255,255,0.03) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

.skeleton-cards {
  display: grid;
  gap: 16px;
}

.skeleton-card {
  padding: 20px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.skeleton-line-header {
  height: 16px;
  width: 40%;
  border-radius: 4px;
  background: linear-gradient(90deg, rgba(255,255,255,0.03) 25%, rgba(255,255,255,0.08) 50%, rgba(255,255,255,0.03) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

.skeleton-line-body {
  height: 14px;
  width: 85%;
  border-radius: 4px;
  background: linear-gradient(90deg, rgba(255,255,255,0.03) 25%, rgba(255,255,255,0.08) 50%, rgba(255,255,255,0.03) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

.skeleton-line-footer {
  height: 12px;
  width: 25%;
  border-radius: 4px;
  background: linear-gradient(90deg, rgba(255,255,255,0.03) 25%, rgba(255,255,255,0.08) 50%, rgba(255,255,255,0.03) 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

.search-tabs { 
  display: flex; 
  gap: 8px; 
  margin-bottom: 24px; 
  border-bottom: 1px solid var(--border);
  padding-bottom: 8px;
}

.search-tab { 
  padding: 8px 18px; 
  border-radius: var(--radius-full); 
  font-size: 13px; 
  color: var(--text-secondary); 
  background: none;
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.3s ease;
  font-weight: 500;
}

.search-tab:hover {
  color: var(--text-primary);
  background: rgba(255,255,255,0.04);
}

.search-tab.active { 
  color: #000; 
  background: var(--accent-gold); 
  font-weight: 700; 
  box-shadow: 0 4px 10px var(--accent-gold-glow);
}

.tab-count {
  font-size: 11px;
  margin-left: 4px;
  opacity: 0.7;
}

.search-tab.active .tab-count {
  font-weight: 600;
}

.search-results { 
  display: flex;
  flex-direction: column;
  gap: 32px; 
}

.result-group-header { 
  font-size: 15px; 
  font-weight: 700;
  color: var(--text-primary); 
  margin-bottom: 16px; 
  display: flex;
  align-items: center;
  gap: 8px;
}

.group-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent-gold);
  box-shadow: 0 0 6px var(--accent-gold);
}

.result-cards-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.result-card { 
  padding: 20px; 
  border-radius: var(--radius-lg); 
  border: 1px solid var(--border); 
  cursor: pointer; 
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 140px;
}

.result-card:hover { 
  transform: translateY(-2px);
  border-color: var(--accent-gold);
  box-shadow: 0 8px 24px rgba(240,185,11,0.06);
}

.card-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 10px;
  gap: 12px;
}

.title-wrap {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.card-main-title {
  font-family: var(--font-mono);
  font-size: 18px;
  font-weight: 800;
  color: var(--text-primary);
}

.card-news-title {
  font-size: 15px;
  font-weight: 700;
  line-height: 1.45;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.card-subtitle {
  font-size: 12px;
  color: var(--text-secondary);
}

.score-badge {
  font-size: 11px;
  font-weight: 600;
  color: var(--accent-gold);
  background: var(--accent-gold-glow);
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  white-space: nowrap;
}

.direction-badge {
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  white-space: nowrap;
}

.direction-long {
  background: var(--positive-bg);
  color: var(--positive);
}

.direction-short {
  background: rgba(239, 68, 68, 0.15);
  color: var(--negative);
}

.card-desc { 
  font-size: 13px; 
  color: var(--text-secondary); 
  line-height: 1.6; 
  margin-bottom: 16px;
  flex: 1;
}

.card-footer {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.footer-tag {
  font-size: 10px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 600;
  text-transform: uppercase;
}

.risk-high {
  background: rgba(239, 68, 68, 0.1);
  color: var(--negative);
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.risk-medium {
  background: var(--accent-gold-glow);
  color: var(--accent-gold);
  border: 1px solid rgba(240, 185, 11, 0.2);
}

.risk-low {
  background: var(--positive-bg);
  color: var(--positive);
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.news-meta-item {
  font-size: 11px;
  color: var(--text-muted);
}

.result-empty { 
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
  border-radius: var(--radius-lg);
  border: 1px dashed var(--border);
  color: var(--text-muted);
  text-align: center;
}

.empty-icon {
  font-size: 32px;
  margin-bottom: 12px;
}

.result-empty p {
  font-size: 13px;
}

@media (max-width: 900px) {
  .exploration-grid, .result-cards-grid {
    grid-template-columns: 1fr;
  }
}
</style>
