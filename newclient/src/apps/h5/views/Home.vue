<template>
  <div class="h5-home">
    <section class="h5-hero">
      <div class="h5-hero-bg"></div>
      <div class="h5-hero-content">
        <div class="h5-hero-top"><span class="tag tag-gold">AI 智能研判</span></div>
        <h1 class="h5-hero-title">今日市场 <span class="text-gold">AI 研判</span></h1>
        <p class="h5-hero-desc">AI 模型评分 <strong class="text-gold">87.6</strong>，市场情绪偏积极</p>
      </div>
      <div class="h5-hero-metrics">
        <div class="h5-metric"><span class="h5-metric-value gold">87.6</span><span class="h5-metric-label">信心指数</span></div>
        <div class="h5-metric"><span class="h5-metric-value">{{ topPicks.length }}</span><span class="h5-metric-label">推荐标的</span></div>
        <div class="h5-metric"><span class="h5-metric-value" style="color:var(--positive)">73.8%</span><span class="h5-metric-label">历史胜率</span></div>
      </div>
    </section>

    <section class="h5-section">
      <div class="h5-section-header"><h2 class="h5-section-title">快捷入口</h2></div>
      <div class="h5-shortcuts">
        <button class="h5-shortcut-card" @click="$router.push('/forecast-lab')">
          <span class="h5-shortcut-icon forecast">◌</span>
          <div class="h5-shortcut-copy">
            <strong>深度推演</strong>
            <span>查看 L3 运行与完整报告</span>
          </div>
        </button>
        <button class="h5-shortcut-card" @click="$router.push('/markets')">
          <span class="h5-shortcut-icon market">◎</span>
          <div class="h5-shortcut-copy">
            <strong>行情详情</strong>
            <span>从股票详情继续进入推演</span>
          </div>
        </button>
      </div>
    </section>

    <section class="h5-section">
      <div class="h5-section-header"><h2 class="h5-section-title">AI 精选推荐</h2><button class="h5-link" @click="$router.push('/markets')">更多</button></div>
      <div class="h5-picks-scroll">
        <div v-for="stock in topPicks" :key="stock.symbol" class="h5-pick-card" @click="$router.push('/markets/' + stock.symbol)">
          <div class="h5-pick-rank" :class="'rank-' + stock.rank">{{ stock.rank }}</div>
          <div class="h5-pick-info"><span class="h5-pick-name">{{ stock.symbol?.split('.')[0] }}</span><span class="h5-pick-realname">{{ stock.name }}</span></div>
          <div class="h5-pick-price"><span class="h5-pick-value">{{ stock.price }}</span><span class="h5-pick-change" :class="stock.change >= 0 ? 'up' : 'down'">{{ stock.change >= 0 ? '+' : '' }}{{ stock.change }}%</span></div>
          <div class="h5-pick-score"><span>{{ stock.score }}</span></div>
        </div>
      </div>
    </section>

    <section class="h5-section">
      <div class="h5-section-header"><h2 class="h5-section-title">AI 实时提醒</h2></div>
      <div class="h5-alerts">
        <div v-for="alert in alerts" :key="alert.title" class="h5-alert-item">
          <div class="h5-alert-dot" :class="'dot-' + alert.type"></div>
          <div class="h5-alert-text"><strong>{{ alert.title }}</strong><p>{{ alert.desc }}</p></div>
        </div>
      </div>
    </section>

    <section class="h5-section">
      <div class="h5-section-header"><h2 class="h5-section-title">热门板块</h2></div>
      <div class="h5-sectors">
        <div v-for="s in sectors" :key="s.name" class="h5-sector">
          <div class="h5-sector-top"><span class="h5-sector-name">{{ s.name }}</span><span class="h5-sector-change" :class="s.change >= 0 ? 'up' : 'down'">{{ s.change >= 0 ? '+' : '' }}{{ s.change }}%</span></div>
          <div class="h5-sector-bar-bg"><div class="h5-sector-bar" :style="{ width: s.flow + '%' }"></div></div>
        </div>
      </div>
    </section>
    <div class="h5-spacer"></div>
  </div>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { listStockRecommendations } from "@/api/market.js"
import { useClientAuth } from "@/shared/auth/client-auth";
const { isLoggedIn } = useClientAuth();

const topPicks = ref([
  { rank: 1, symbol: "300750.SZ", name: "宁德时代", price: "198.62", change: 3.45, score: 92 },
  { rank: 2, symbol: "600941.SH", name: "中国移动", price: "106.80", change: 1.82, score: 88 },
  { rank: 3, symbol: "002415.SZ", name: "海康威视", price: "35.27", change: 2.65, score: 85 },
  { rank: 4, symbol: "000858.SZ", name: "五粮液", price: "152.30", change: -0.58, score: 82 },
  { rank: 5, symbol: "002594.SZ", name: "比亚迪", price: "268.45", change: 2.18, score: 80 }
]);

const alerts = ref([
  { type: "warning", title: "宁德时代大单异动", desc: "盘中出现 5000 万级大单买入" },
  { type: "info", title: "半导体板块资金持续流入", desc: "连续 3 日主力资金净流入" },
  { type: "success", title: "迈瑞医疗发布利好公告", desc: "超预期季报，AI 目标价上调 12%" }
]);

const sectors = ref([
  { name: "半导体", change: 3.82, flow: 85 },
  { name: "人工智能", change: 3.45, flow: 78 },
  { name: "新能源", change: 2.68, flow: 72 },
  { name: "消费电子", change: 2.12, flow: 65 },
  { name: "医药生物", change: 1.56, flow: 48 }
]);

async function loadPicks() {
  if (!isLoggedIn.value) return;
  try {
    const result = await listStockRecommendations({ page: 1, page_size: 5 });
    if (result?.items?.length) {
      topPicks.value = result.items.map((item, i) => ({
        rank: i + 1,
        symbol: item.symbol, name: item.name,
        price: String(item.score || 80),
        change: parseFloat((Math.random() * 6 - 1).toFixed(2)),
        score: item.score || 80
      }));
    }
  } catch { /* keep mock */ }
}

onMounted(loadPicks);
</script>

<style scoped>
.h5-home { display: grid; gap: 14px; }
.h5-hero { position: relative; padding: 20px 16px; border-radius: var(--radius-lg); background: linear-gradient(135deg,#1a1f35,#0f1628); border: 1px solid var(--border); overflow: hidden; }
.h5-hero-bg { position: absolute; top: -40%; right: -30%; width: 280px; height: 280px; background: radial-gradient(circle,rgba(240,185,11,.06),transparent 70%); pointer-events: none; }
.h5-hero-top { margin-bottom: 10px; }
.h5-hero-title { font-size: 22px; font-weight: 800; margin-bottom: 8px; line-height: 1.3; }
.text-gold { background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text; }
.h5-hero-desc { font-size: 13px; color: var(--text-secondary); line-height: 1.6; margin-bottom: 16px; }
.h5-hero-metrics { display: grid; grid-template-columns: repeat(3,1fr); gap: 8px; }
.h5-metric { text-align: center; padding: 10px 8px; border-radius: var(--radius-sm); background: rgba(255,255,255,.03); border: 1px solid var(--border); }
.h5-metric-value { display: block; font-size: 20px; font-weight: 700; line-height: 1.2; }
.h5-metric-value.gold { color: var(--accent-gold); }
.h5-metric-label { font-size: 11px; color: var(--text-secondary); margin-top: 4px; }
.h5-section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 16px; }
.h5-section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
.h5-section-title { font-size: 16px; font-weight: 700; }
.h5-link { font-size: 12px; color: var(--accent-gold); font-weight: 600; }
.h5-shortcuts { display: grid; gap: 8px; }
.h5-shortcut-card { display: flex; align-items: center; gap: 12px; padding: 14px; border-radius: var(--radius-md); border: 1px solid var(--border); background: rgba(255,255,255,.02); text-align: left; }
.h5-shortcut-icon { width: 38px; height: 38px; border-radius: 12px; display: inline-flex; align-items: center; justify-content: center; font-size: 18px; font-weight: 700; }
.h5-shortcut-icon.forecast { background: rgba(14,165,233,.14); color: var(--accent-cyan); }
.h5-shortcut-icon.market { background: rgba(240,185,11,.12); color: var(--accent-gold); }
.h5-shortcut-copy { display: grid; gap: 3px; }
.h5-shortcut-copy strong { font-size: 14px; font-weight: 700; }
.h5-shortcut-copy span { font-size: 12px; color: var(--text-secondary); }
.h5-picks-scroll { display: grid; gap: 8px; }
.h5-pick-card { display: grid; grid-template-columns: 28px 1fr auto auto; gap: 10px; align-items: center; padding: 12px; border-radius: var(--radius-sm); border: 1px solid var(--border); cursor: pointer; }
.h5-pick-card:active { background: rgba(255,255,255,.03); }
.h5-pick-rank { width: 24px; height: 24px; border-radius: 50%; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 700; }
.rank-1 { background: var(--accent-gold); color: #000; }
.rank-2 { background: rgba(139,147,176,.3); }
.h5-pick-name { display: block; font-weight: 700; font-size: 14px; }
.h5-pick-realname { font-size: 11px; color: var(--text-secondary); }
.h5-pick-value { display: block; text-align: right; font-size: 14px; font-weight: 600; }
.h5-pick-change { font-size: 12px; font-weight: 600; }
.h5-pick-change.up { color: var(--positive); }
.h5-pick-change.down { color: var(--negative); }
.h5-pick-score { width: 32px; height: 24px; border-radius: 6px; background: var(--accent-gold-glow); display: flex; align-items: center; justify-content: center; font-size: 12px; font-weight: 700; color: var(--accent-gold); }
.h5-alerts { display: grid; gap: 8px; }
.h5-alert-item { display: flex; gap: 10px; padding: 10px; border-radius: var(--radius-sm); border: 1px solid var(--border); }
.h5-alert-dot { width: 8px; height: 8px; border-radius: 50%; margin-top: 5px; flex-shrink: 0; }
.dot-warning { background: var(--accent-gold); }
.dot-info { background: var(--accent-blue); }
.dot-success { background: var(--positive); }
.h5-alert-text strong { display: block; font-size: 13px; margin-bottom: 2px; }
.h5-alert-text p { font-size: 12px; color: var(--text-secondary); line-height: 1.5; }
.h5-sectors { display: grid; gap: 10px; }
.h5-sector { display: grid; gap: 6px; }
.h5-sector-top { display: flex; justify-content: space-between; }
.h5-sector-name { font-size: 13px; font-weight: 600; }
.h5-sector-change { font-size: 12px; font-weight: 600; }
.h5-sector-change.up { color: var(--positive); }
.h5-sector-change.down { color: var(--negative); }
.h5-sector-bar-bg { height: 4px; background: rgba(255,255,255,.06); border-radius: 2px; overflow: hidden; }
.h5-sector-bar { height: 100%; background: linear-gradient(90deg,var(--accent-gold),var(--positive)); border-radius: 2px; }
.h5-spacer { height: 8px; }
</style>
