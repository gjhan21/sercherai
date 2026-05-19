<template>
  <div class="stocks-page">
    <!-- Filters -->
    <section class="filters-bar glass fade-in-up">
      <div class="filter-tabs">
        <button v-for="tab in tabs" :key="tab.key" class="filter-tab" :class="{ active: activeTab === tab.key }" @click="activeTab = tab.key">
          {{ tab.label }}
        </button>
      </div>
      <div class="filter-search">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14"><circle cx="11" cy="11" r="7"/><path d="M16.5 16.5L21 21"/></svg>
        <input type="text" v-model="searchQuery" placeholder="筛选股票..." />
      </div>
    </section>

    <!-- Stock Table -->
    <section class="section fade-in-up fade-in-up-delay-1">
      <div class="stocks-table-wrap">
        <table class="stocks-table">
          <thead>
            <tr>
              <th></th>
              <th>代码</th>
              <th>名称</th>
              <th>现价</th>
              <th>涨跌幅</th>
              <th>AI评分</th>
              <th>AI判断</th>
              <th>主力资金</th>
              <th>换手率</th>
              <th>关注</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(stock, i) in filteredStocks" :key="stock.symbol" class="stock-row" @click="$router.push('/analysis')">
              <td><span class="stock-index">{{ i + 1 }}</span></td>
              <td><span class="stock-code">{{ stock.symbol }}</span></td>
              <td><span class="stock-name">{{ stock.name }}</span></td>
              <td><span class="stock-price">{{ stock.price }}</span></td>
              <td><span class="stock-change" :class="stock.change >= 0 ? 'up' : 'down'">{{ stock.change >= 0 ? '+' : '' }}{{ stock.change }}%</span></td>
              <td>
                <div class="score-pill" :class="scoreClass(stock.score)">
                  <span>{{ stock.score }}</span>
                </div>
              </td>
              <td><span class="tag" :class="stock.judgment === '看多' ? 'tag-green' : stock.judgment === '看空' ? 'tag-red' : 'tag-neutral'">{{ stock.judgment }}</span></td>
              <td>
                <div class="fund-flow">
                  <div class="flow-bar-bg">
                    <div class="flow-bar" :style="{ width: Math.abs(stock.flow) + '%', background: stock.flow >= 0 ? 'var(--positive)' : 'var(--negative)' }"></div>
                  </div>
                  <span class="flow-text" :class="stock.flow >= 0 ? 'up' : 'down'">{{ stock.flow >= 0 ? '+' : '' }}{{ stock.flow }}%</span>
                </div>
              </td>
              <td><span class="turnover">{{ stock.turnover }}%</span></td>
              <td @click.stop>
                <button class="watch-btn" :class="{ watched: stock.watched }" @click="toggleWatch(stock)">
                  <svg viewBox="0 0 24 24" :fill="stock.watched ? 'var(--accent-gold)' : 'none'" stroke="currentColor" stroke-width="2" width="16" height="16"><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></svg>
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, ref } from "vue";

const activeTab = ref("all");
const searchQuery = ref("");

const tabs = [
  { key: "all", label: "全部" },
  { key: "ai-pick", label: "AI 精选" },
  { key: "hot", label: "热门" },
  { key: "breakout", label: "突破" },
  { key: "volume", label: "放量异动" },
  { key: "watchlist", label: "我的关注" }
];

const stocks = ref([
  { symbol: "300750.SZ", name: "宁德时代", price: "198.62", change: 3.45, score: 92, judgment: "看多", flow: 85, turnover: 1.82, watched: false, tag: "ai-pick" },
  { symbol: "600941.SH", name: "中国移动", price: "106.80", change: 1.82, score: 88, judgment: "看多", flow: 62, turnover: 0.45, watched: false, tag: "ai-pick" },
  { symbol: "002415.SZ", name: "海康威视", price: "35.27", change: 2.65, score: 85, judgment: "看多", flow: 73, turnover: 1.25, watched: false, tag: "hot" },
  { symbol: "688981.SH", name: "中芯国际", price: "56.78", change: 4.12, score: 79, judgment: "看多", flow: 91, turnover: 3.56, watched: false, tag: "hot" },
  { symbol: "000858.SZ", name: "五粮液", price: "152.30", change: -0.58, score: 82, judgment: "中性", flow: -12, turnover: 0.68, watched: false, tag: "" },
  { symbol: "002594.SZ", name: "比亚迪", price: "268.45", change: 2.18, score: 80, judgment: "看多", flow: 45, turnover: 1.12, watched: false, tag: "ai-pick" },
  { symbol: "601318.SH", name: "中国平安", price: "48.26", change: 0.85, score: 72, judgment: "中性", flow: 18, turnover: 0.32, watched: false, tag: "" },
  { symbol: "000001.SZ", name: "平安银行", price: "12.45", change: -1.12, score: 65, judgment: "看空", flow: -38, turnover: 0.89, watched: false, tag: "volume" },
  { symbol: "600036.SH", name: "招商银行", price: "36.78", change: 0.92, score: 75, judgment: "中性", flow: 22, turnover: 0.28, watched: false, tag: "" },
  { symbol: "300059.SZ", name: "东方财富", price: "16.82", change: 5.68, score: 86, judgment: "看多", flow: 95, turnover: 4.82, watched: false, tag: "breakout" },
  { symbol: "002230.SZ", name: "科大讯飞", price: "48.90", change: 3.28, score: 83, judgment: "看多", flow: 78, turnover: 2.15, watched: false, tag: "hot" },
  { symbol: "600519.SH", name: "贵州茅台", price: "1,685.00", change: 0.45, score: 76, judgment: "中性", flow: 8, turnover: 0.18, watched: false, tag: "" },
  { symbol: "688012.SH", name: "中微公司", price: "168.20", change: 6.82, score: 87, judgment: "看多", flow: 88, turnover: 4.56, watched: false, tag: "volume" },
  { symbol: "002475.SZ", name: "立讯精密", price: "32.56", change: 2.45, score: 78, judgment: "看多", flow: 56, turnover: 1.68, watched: false, tag: "" },
  { symbol: "300124.SZ", name: "汇川技术", price: "68.90", change: 1.82, score: 74, judgment: "中性", flow: 32, turnover: 0.75, watched: false, tag: "" }
]);

const filteredStocks = computed(() => {
  let list = stocks.value;
  if (activeTab.value === "ai-pick") {
    list = list.filter(s => s.tag === "ai-pick");
  } else if (activeTab.value === "hot") {
    list = list.filter(s => s.tag === "hot");
  } else if (activeTab.value === "breakout") {
    list = list.filter(s => s.tag === "breakout");
  } else if (activeTab.value === "volume") {
    list = list.filter(s => s.tag === "volume");
  } else if (activeTab.value === "watchlist") {
    list = list.filter(s => s.watched);
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase();
    list = list.filter(s => s.symbol.toLowerCase().includes(q) || s.name.toLowerCase().includes(q));
  }
  return list;
});

function scoreClass(score) {
  if (score >= 85) return "score-high";
  if (score >= 70) return "score-mid";
  return "score-low";
}

function toggleWatch(stock) {
  stock.watched = !stock.watched;
}
</script>

<style scoped>
.stocks-page {
  display: grid;
  gap: 16px;
  max-width: 1400px;
}

.section {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 20px;
}

/* Filters */
.filters-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-radius: var(--radius-lg);
  gap: 16px;
}

.filter-tabs {
  display: flex;
  gap: 4px;
}

.filter-tab {
  padding: 6px 14px;
  border-radius: var(--radius-full);
  font-size: 13px;
  color: var(--text-secondary);
  transition: all 0.2s;
}

.filter-tab:hover {
  color: var(--text-primary);
  background: rgba(255, 255, 255, 0.04);
}

.filter-tab.active {
  color: #000;
  background: var(--accent-gold);
  font-weight: 600;
}

.filter-search {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: var(--radius-full);
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid var(--border);
}

.filter-search input {
  background: none;
  color: var(--text-primary);
  font-size: 13px;
  width: 140px;
}

.filter-search input::placeholder {
  color: var(--text-muted);
}

/* Table */
.stocks-table-wrap {
  overflow-x: auto;
}

.stocks-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.stocks-table th {
  text-align: left;
  padding: 10px 12px;
  font-size: 11px;
  color: var(--text-secondary);
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border);
  white-space: nowrap;
}

.stock-row {
  cursor: pointer;
  transition: background 0.15s;
}

.stock-row:hover {
  background: rgba(255, 255, 255, 0.03);
}

.stock-row td {
  padding: 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  vertical-align: middle;
}

.stock-index {
  color: var(--text-muted);
  font-size: 12px;
}

.stock-code {
  color: var(--text-secondary);
  font-family: var(--font-mono);
  font-size: 12px;
}

.stock-name {
  font-weight: 600;
}

.stock-price {
  font-weight: 600;
  font-family: var(--font-mono);
}

.stock-change {
  font-weight: 600;
}

.stock-change.up { color: var(--positive); }
.stock-change.down { color: var(--negative); }

/* Score pill */
.score-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 24px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
}

.score-high { background: var(--accent-gold-glow); color: var(--accent-gold); }
.score-mid { background: rgba(59, 130, 246, 0.1); color: var(--accent-blue); }
.score-low { background: rgba(139, 147, 176, 0.1); color: var(--text-secondary); }

/* Fund Flow */
.fund-flow {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 100px;
}

.flow-bar-bg {
  flex: 1;
  height: 4px;
  background: rgba(255, 255, 255, 0.06);
  border-radius: 2px;
  overflow: hidden;
}

.flow-bar {
  height: 100%;
  border-radius: 2px;
}

.flow-text {
  font-size: 11px;
  font-weight: 600;
  width: 45px;
  text-align: right;
}

.flow-text.up { color: var(--positive); }
.flow-text.down { color: var(--negative); }

.turnover {
  color: var(--text-secondary);
  font-size: 12px;
}

.watch-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: 50%;
  color: var(--text-muted);
  transition: all 0.2s;
}

.watch-btn:hover {
  background: rgba(255, 255, 255, 0.06);
  color: var(--accent-gold);
}

.watch-btn.watched {
  color: var(--accent-gold);
}
</style>
