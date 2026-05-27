<template>
  <div class="home-page">
    <!-- AI Market Pulse Hero -->
    <section class="market-hero fade-in-up">
      <div class="hero-bg-glow"></div>
      <div class="hero-content">
        <div class="hero-top">
          <span class="tag tag-gold">AI 智能研判</span>
          <span class="hero-time">数据更新于 {{ lastUpdate }}</span>
        </div>
        <h1 class="hero-title">今日市场 <span class="text-gold">AI 研判</span></h1>
        <p class="hero-desc">基于多维度数据分析，AI 认为今日市场偏<span class="text-positive">积极</span>，建议关注科技与消费板块机会</p>
        <div class="hero-stats">
          <div class="hero-stat" @click="$router.push('/recommendations')">
            <span class="stat-label">AI 信心指数</span>
            <span class="stat-value gold">87.6</span>
            <span class="stat-change up">+2.3%</span>
          </div>
          <div class="hero-stat" @click="$router.push('/recommendations')">
            <span class="stat-label">推荐标的</span>
            <span class="stat-value">12</span><span class="stat-sub">只精选</span>
          </div>
          <div class="hero-stat" @click="$router.push('/recommendations/history')">
            <span class="stat-label">历史胜率</span>
            <span class="stat-value">73.8%</span><span class="stat-sub">近30日</span>
          </div>
          <div class="hero-stat" @click="$router.push('/user/vip')">
            <span class="stat-label">风险等级</span>
            <span class="stat-value" style="color:var(--accent-cyan)">中等</span><span class="stat-sub">建议均衡配置</span>
          </div>
        </div>
      </div>
    </section>

    <!-- Quick Links to Modules -->
    <div class="quick-modules fade-in-up fade-in-up-delay-2">
      <button class="quick-module glass card-hover" @click="$router.push('/recommendations')">
        <span class="qm-icon" style="background:var(--accent-gold-glow);color:var(--accent-gold)">★</span>
        <span class="qm-title">每日推荐</span>
        <span class="qm-desc">AI 精选 6 只标的</span>
      </button>
      <button class="quick-module glass card-hover" @click="$router.push('/identify')">
        <span class="qm-icon" style="background:rgba(59,130,246,.1);color:var(--accent-blue)">◎</span>
        <span class="qm-title">鉴别分析</span>
        <span class="qm-desc">输入股票 AI 分析</span>
      </button>
      <button class="quick-module glass card-hover" @click="$router.push('/forecast-lab')">
        <span class="qm-icon" style="background:rgba(14,165,233,.14);color:var(--accent-cyan)">◌</span>
        <span class="qm-title">深度推演</span>
        <span class="qm-desc">查看 L3 运行与报告</span>
      </button>
      <button class="quick-module glass card-hover" @click="$router.push('/news')">
        <span class="qm-icon" style="background:var(--positive-bg);color:var(--positive)">📰</span>
        <span class="qm-title">资讯中心</span>
        <span class="qm-desc">市场最新动态</span>
      </button>
      <button class="quick-module glass card-hover" @click="$router.push('/community')">
        <span class="qm-icon" style="background:rgba(139,92,246,.1);color:#8b5cf6">💬</span>
        <span class="qm-title">社区</span>
        <span class="qm-desc">投资者交流</span>
      </button>
    </div>

    <!-- AI Top Picks -->
    <section class="section fade-in-up fade-in-up-delay-2">
      <div class="section-header">
        <div><h2 class="section-title">AI 精选推荐</h2><p class="section-subtitle">多因子选股模型今日推荐</p></div>
        <div class="section-actions"><button class="btn-primary" @click="$router.push('/recommendations')">查看全部</button></div>
      </div>
      <div class="picks-grid">
        <article v-for="stock in topPicks" :key="stock.symbol" class="pick-card glass card-hover" @click="$router.push('/identify/' + stock.symbol)">
          <div class="pick-rank" :class="'rank-' + stock.rank">{{ stock.rank }}</div>
          <div class="pick-info">
            <div><span class="pick-symbol">{{ stock.symbol }}</span><span class="pick-name">{{ stock.name }}</span></div>
            <div class="pick-price-row"><span class="pick-price">{{ stock.price }}</span><span class="pick-change" :class="stock.change >= 0 ? 'up' : 'down'">{{ stock.change >= 0 ? '+' : '' }}{{ stock.change }}%</span></div>
          </div>
          <div class="pick-score-ring"><svg viewBox="0 0 36 36" class="score-svg"><circle cx="18" cy="18" r="15.5" fill="none" stroke="rgba(255,255,255,0.06)" stroke-width="2.5"/><circle cx="18" cy="18" r="15.5" fill="none" stroke="var(--accent-gold)" stroke-width="2.5" stroke-dasharray="97.4" :stroke-dashoffset="97.4 - 97.4 * stock.score / 100" stroke-linecap="round"/></svg><span class="score-num">{{ stock.score }}</span></div>
          <p class="pick-reason"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" width="14" height="14" style="color:var(--accent-gold);flex-shrink:0;margin-top:2px"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4"/></svg><span>{{ stock.reason }}</span></p>
          <div class="pick-tags"><span v-for="tag in stock.tags" :key="tag" class="tag" :class="tag === 'VIP' ? 'tag-gold' : 'tag-blue'">{{ tag }}</span></div>
        </article>
      </div>
    </section>

    <!-- Hot Sectors + Alerts -->
    <div class="home-grid-2col fade-in-up fade-in-up-delay-3">
      <section class="section">
        <div class="section-header"><div><h2 class="section-title">热门板块</h2></div></div>
        <div class="sectors-list">
          <div v-for="sector in sectors" :key="sector.name" class="sector-item glass">
            <div class="sector-left"><span class="sector-name">{{ sector.name }}</span><div class="sector-bar-bg"><div class="sector-bar" :style="{ width: sector.flow + '%' }"></div></div></div>
            <div class="sector-right"><span class="sector-change" :class="sector.change >= 0 ? 'up' : 'down'">{{ sector.change >= 0 ? '+' : '' }}{{ sector.change }}%</span><span class="sector-flow">资金 +{{ sector.flow }}%</span></div>
          </div>
        </div>
      </section>

      <section class="section">
        <div class="section-header"><div><h2 class="section-title">AI 预警提醒</h2></div></div>
        <div class="alerts-list">
          <div v-for="alert in alerts" :key="alert.title" class="alert-item">
            <span class="alert-dot" :class="'dot-' + alert.type"></span>
            <div class="alert-text"><strong>{{ alert.title }}</strong><p>{{ alert.desc }}</p></div>
            <span class="alert-time">{{ alert.time }}</span>
          </div>
        </div>
      </section>
    </div>

    <!-- Latest Reports + Community -->
    <div class="home-grid-2col fade-in-up fade-in-up-delay-4">
      <section class="section">
        <div class="section-header"><div><h2 class="section-title">最新 AI 研报</h2></div><div class="section-actions"><button class="btn-ghost" @click="$router.push('/news')">更多</button></div></div>
        <div class="reports-list">
          <article v-for="r in reports" :key="r.title" class="report-item glass card-hover" @click="$router.push('/news')">
            <div class="report-item-top"><span class="tag" :class="r.tag === '深度' ? 'tag-gold' : 'tag-blue'">{{ r.tag }}</span><span class="report-time">{{ r.time }}</span></div>
            <h3>{{ r.title }}</h3>
            <p>{{ r.summary }}</p>
          </article>
        </div>
      </section>

      <section class="section">
        <div class="section-header"><div><h2 class="section-title">社区热帖</h2></div><div class="section-actions"><button class="btn-ghost" @click="$router.push('/community')">更多</button></div></div>
        <div class="hot-posts">
          <div v-for="post in hotPosts" :key="post.id" class="hot-post glass card-hover" @click="$router.push('/community/post/' + post.id)">
            <div class="hot-post-author">{{ post.author.nickname[0] }}</div>
            <div class="hot-post-body">
              <strong>{{ post.title }}</strong>
              <span>{{ post.likes }} ♥ · {{ post.comments }} 评论</span>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- VIP CTA -->
    <section class="vip-cta glass fade-in-up fade-in-up-delay-5" @click="$router.push('/user/vip')">
      <div class="vip-cta-content">
        <span class="tag tag-gold">VIP 会员</span>
        <h2>升级白银会员，解锁无限 AI 分析</h2>
        <p>仅 ¥99/月，享受进阶技术指标、Level-2 行情等专属权益</p>
      </div>
      <button class="btn-primary">了解详情</button>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from "vue";
import { POSTS } from "@/mock/community.js";
import { NEWS_ARTICLES } from "@/mock/news.js";
import { listStockRecommendations } from "@/api/market.js"
import { useClientAuth } from "@/shared/auth/client-auth";
const { isLoggedIn } = useClientAuth();

const lastUpdate = computed(() => {
  const d = new Date();
  return `${d.getHours().toString().padStart(2,'0')}:${d.getMinutes().toString().padStart(2,'0')}`;
});

const topPicks = ref([
  { rank: 1, symbol: '300750.SZ', name: '宁德时代', price: '198.62', change: 3.45, score: 92, reason: 'AI 检测到主力资金持续流入，技术形态突破', tags: ['VIP', '资金流入'] },
  { rank: 2, symbol: '600941.SH', name: '中国移动', price: '106.80', change: 1.82, score: 88, reason: '高股息防御属性，AI 模型评分持续走高', tags: ['高股息', '稳健'] },
  { rank: 3, symbol: '002415.SZ', name: '海康威视', price: '35.27', change: 2.65, score: 85, reason: 'AI 概念催化，量价齐升突破平台', tags: ['AI概念', '放量'] },
  { rank: 4, symbol: '000858.SZ', name: '五粮液', price: '152.30', change: -0.58, score: 82, reason: '估值处于历史低位，AI 判断反弹窗口临近', tags: ['消费', '估值修复'] },
  { rank: 5, symbol: '002594.SZ', name: '比亚迪', price: '268.45', change: 2.18, score: 80, reason: '新能源政策利好，AI 趋势模型看多', tags: ['新能源', '趋势'] },
  { rank: 6, symbol: '688981.SH', name: '中芯国际', price: '56.78', change: 4.12, score: 79, reason: '半导体周期回暖，AI 情绪模型积极', tags: ['半导体', '热点'] }
]);

const sectors = [
  { name: '半导体', change: 3.82, flow: 85 },
  { name: '人工智能', change: 3.45, flow: 78 },
  { name: '新能源', change: 2.68, flow: 72 },
  { name: '消费电子', change: 2.12, flow: 65 },
  { name: '医药生物', change: 1.56, flow: 48 },
  { name: '金融', change: 0.85, flow: 35 }
];

const alerts = [
  { type: 'warning', title: '宁德时代出现大单异动', desc: '盘中出现 5000 万级大单买入', time: '10:32' },
  { type: 'info', title: '半导体板块资金持续流入', desc: '连续 3 日主力资金净流入，AI 评级上调', time: '10:15' },
  { type: 'success', title: '迈瑞医疗发布超预期季报', desc: 'AI 预测目标价上调 12%', time: '09:48' }
];

const reports = NEWS_ARTICLES.filter(a => a.isFeatured).slice(0, 3);
const hotPosts = [...POSTS].sort((a, b) => b.likes - a.likes).slice(0, 4);

async function loadHomeData() {
  if (!isLoggedIn.value) return;
  try {
    const result = await listStockRecommendations({ page: 1, page_size: 6 });
    if (result?.items?.length) {
      topPicks.value = result.items.map((item, i) => ({
        rank: i + 1, symbol: item.symbol, name: item.name,
        price: String(item.score || Math.random() * 100).slice(0, 6),
        change: parseFloat((Math.random() * 6 - 1).toFixed(2)),
        score: item.score || 80,
        reason: item.reason_summary || 'AI 模型筛选推荐',
        tags: ['AI精选']
      }));
    }
  } catch { /* keep mock data */ }
}

onMounted(loadHomeData);
</script>

<style scoped>
.home-page { display: grid; gap: 20px; max-width: 1400px; }

.market-hero { position: relative; border-radius: var(--radius-xl); padding: 28px; background: linear-gradient(135deg,#1a1f35,#0f1628,#1a1f35); border: 1px solid var(--border); overflow: hidden; }
.hero-bg-glow { position: absolute; top: -50%; right: -20%; width: 500px; height: 500px; background: radial-gradient(circle,rgba(240,185,11,.06),transparent 70%); pointer-events: none; }
.hero-top { display: flex; align-items: center; gap: 12px; margin-bottom: 12px; }
.hero-time { font-size: 12px; color: var(--text-secondary); }
.hero-title { font-size: 32px; font-weight: 800; margin-bottom: 10px; }
.text-gold { background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
.text-positive { color: var(--positive); }
.hero-desc { font-size: 14px; color: var(--text-secondary); margin-bottom: 20px; }
.hero-stats { display: grid; grid-template-columns: repeat(4,1fr); gap: 10px; }
.hero-stat { background: rgba(255,255,255,.03); border: 1px solid var(--border); border-radius: var(--radius-md); padding: 12px; cursor: pointer; transition: all .2s; }
.hero-stat:hover { background: rgba(255,255,255,.06); }
.stat-label { display: block; font-size: 11px; color: var(--text-secondary); margin-bottom: 4px; text-transform: uppercase; letter-spacing: .5px; }
.stat-value { font-size: 22px; font-weight: 700; }
.stat-value.gold { color: var(--accent-gold); }
.stat-change { font-size: 12px; font-weight: 600; }
.stat-change.up { color: var(--positive); }
.stat-sub { font-size: 12px; color: var(--text-secondary); margin-left: 4px; }

.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 20px; }
.section-header { display: flex; align-items: flex-start; justify-content: space-between; margin-bottom: 14px; }
.section-title { font-size: 18px; font-weight: 700; }
.section-subtitle { font-size: 12px; color: var(--text-secondary); }
.section-actions { display: flex; gap: 8px; }
.btn-ghost { padding: 6px 14px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 12px; color: var(--text-secondary); }
.btn-ghost:hover { border-color: var(--border-light); color: var(--text-primary); }
.btn-primary { padding: 8px 18px; border-radius: var(--radius-full); background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; font-size: 13px; font-weight: 600; }

/* Quick Modules */
.quick-modules { display: grid; grid-template-columns: repeat(4,1fr); gap: 10px; }
.quick-module { display: flex; flex-direction: column; align-items: center; gap: 8px; padding: 20px; border-radius: var(--radius-lg); cursor: pointer; text-align: center; }
.qm-icon { width: 48px; height: 48px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 20px; }
.qm-title { font-size: 15px; font-weight: 700; }
.qm-desc { font-size: 12px; color: var(--text-secondary); }

.picks-grid { display: grid; grid-template-columns: repeat(3,1fr); gap: 10px; }
.pick-card { padding: 14px; border-radius: var(--radius-md); cursor: pointer; position: relative; }
.pick-rank { position: absolute; top: 10px; right: 10px; width: 22px; height: 22px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 10px; font-weight: 700; }
.rank-1 { background: var(--accent-gold); color: #000; }
.rank-2 { background: rgba(139,147,176,.3); }
.rank-3 { background: rgba(139,147,176,.2); }
.pick-info { display: flex; justify-content: space-between; margin-bottom: 8px; }
.pick-symbol { display: block; font-size: 14px; font-weight: 700; }
.pick-name { font-size: 11px; color: var(--text-secondary); }
.pick-price { font-size: 16px; font-weight: 600; }
.pick-change { font-size: 12px; font-weight: 600; }
.pick-change.up { color: var(--positive); }
.pick-change.down { color: var(--negative); }
.pick-score-ring { position: relative; width: 36px; height: 36px; display: flex; align-items: center; justify-content: center; margin-bottom: 8px; }
.score-svg { position: absolute; width: 36px; height: 36px; }
.score-num { font-size: 10px; font-weight: 700; color: var(--accent-gold); }
.pick-reason { display: flex; gap: 6px; font-size: 11px; color: var(--text-secondary); line-height: 1.6; margin-bottom: 8px; }
.pick-tags { display: flex; gap: 4px; flex-wrap: wrap; }

.home-grid-2col { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; }

.sectors-list { display: grid; gap: 6px; }
.sector-item { display: flex; align-items: center; justify-content: space-between; padding: 10px 12px; border-radius: var(--radius-sm); }
.sector-name { display: block; font-size: 12px; font-weight: 600; margin-bottom: 4px; }
.sector-bar-bg { height: 3px; background: rgba(255,255,255,.06); border-radius: 2px; overflow: hidden; max-width: 120px; }
.sector-bar { height: 100%; background: linear-gradient(90deg,var(--accent-gold),var(--positive)); border-radius: 2px; }
.sector-change { display: block; font-size: 13px; font-weight: 600; }
.sector-change.up { color: var(--positive); }
.sector-change.down { color: var(--negative); }
.sector-flow { font-size: 10px; color: var(--text-secondary); }

.alerts-list { display: grid; gap: 6px; }
.alert-item { display: flex; gap: 10px; padding: 10px; border-radius: var(--radius-sm); border: 1px solid var(--border); align-items: flex-start; }
.alert-dot { width: 8px; height: 8px; border-radius: 50%; margin-top: 4px; flex-shrink: 0; }
.dot-warning { background: var(--accent-gold); }
.dot-info { background: var(--accent-blue); }
.dot-success { background: var(--positive); }
.alert-text strong { display: block; font-size: 12px; margin-bottom: 2px; }
.alert-text p { font-size: 11px; color: var(--text-secondary); line-height: 1.5; }
.alert-time { font-size: 10px; color: var(--text-muted); white-space: nowrap; }

.reports-list { display: grid; gap: 8px; }
.report-item { padding: 12px; border-radius: var(--radius-sm); cursor: pointer; }
.report-item h3 { font-size: 13px; font-weight: 700; margin: 6px 0; }
.report-item p { font-size: 11px; color: var(--text-secondary); line-height: 1.5; }
.report-item-top { display: flex; align-items: center; justify-content: space-between; }
.report-time { font-size: 10px; color: var(--text-muted); }

.hot-posts { display: grid; gap: 6px; }
.hot-post { display: flex; gap: 10px; align-items: center; padding: 10px; border-radius: var(--radius-sm); cursor: pointer; }
.hot-post-author { width: 28px; height: 28px; border-radius: 50%; background: var(--accent-gold-glow); color: var(--accent-gold); display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; flex-shrink: 0; }
.hot-post-body strong { display: block; font-size: 12px; line-height: 1.3; }
.hot-post-body span { font-size: 10px; color: var(--text-muted); }

.vip-cta { display: flex; align-items: center; justify-content: space-between; padding: 20px 24px; border-radius: var(--radius-lg); cursor: pointer; }
.vip-cta-content h2 { font-size: 18px; font-weight: 700; margin: 8px 0; }
.vip-cta-content p { font-size: 13px; color: var(--text-secondary); }

@media (max-width:1100px) { .picks-grid { grid-template-columns: repeat(2,1fr); } .quick-modules { grid-template-columns: repeat(2,1fr); } }
@media (max-width:900px) { .home-grid-2col { grid-template-columns: 1fr; } .hero-stats { grid-template-columns: repeat(2,1fr); } }
@media (max-width:600px) { .picks-grid { grid-template-columns: 1fr; } .hero-stats { grid-template-columns: 1fr 1fr; } }
</style>
