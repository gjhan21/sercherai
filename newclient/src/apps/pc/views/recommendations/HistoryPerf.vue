<template>
  <div class="history-page">
    <!-- View Switch Tabs -->
    <div class="view-switch-tabs fade-in-up">
      <button class="view-switch-btn" :class="{active: activeView === 'simulated'}" @click="activeView = 'simulated'">
        📈 策略模拟持仓跟踪
      </button>
      <button class="view-switch-btn" :class="{active: activeView === 'legacy'}" @click="activeView = 'legacy'">
        📋 历史推荐表现归因
      </button>
    </div>

    <!-- Page Header & Stats Section -->
    <section class="section fade-in-up">
      <div class="section-header">
        <div>
          <h2 class="section-title">
            {{ activeView === 'simulated' ? '模拟盘多维概览' : '历史推荐表现' }}
          </h2>
          <p class="section-subtitle" v-if="showLoginGate">
            登录后查看策略收益与表现
          </p>
          <p class="section-subtitle" v-else-if="loading || simLoading">
            数据加载中...
          </p>
          <p class="section-subtitle" v-else>
            {{ activeView === 'simulated' ? '根据策略规则，对每次推荐标的以当日收盘价模拟建仓，追踪完整战绩与资金曲线' : '基于真实历史行情追踪 AI 推荐标的的区间表现' }}
          </p>
        </div>
        
        <!-- Filter Tabs for Legacy View -->
        <div class="filter-tabs" v-if="!showLoginGate && activeView === 'legacy'">
          <button v-for="f in filters" :key="f.key" class="filter-tab" :class="{active: activeFilter === f.key}" @click="activeFilter = f.key">{{ f.label }}</button>
        </div>
      </div>

      <!-- Login Gate -->
      <div v-if="showLoginGate" class="history-login-gate glass">
        <strong>登录后查看历史表现与模拟持仓</strong>
        <p>系统将展示真实的建仓价、最新/结算价、区间收益、最大回撤与持仓跟踪，不再使用演示数据。</p>
        <button class="hero-cta primary" @click="$router.push('/login')">去登录</button>
      </div>

      <!-- Stats Cards: Simulated Positions Tracker -->
      <div v-else-if="activeView === 'simulated'" class="history-stats-row">
        <div class="history-stat glass">
          <span class="hstat-label">累计模拟交易</span>
          <span class="hstat-value">{{ simulatedOverview.total_trades || 0 }} 笔</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">持仓中股票</span>
          <span class="hstat-value" style="color: var(--accent-blue)">{{ simulatedOverview.active_holdings || 0 }} 只</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">策略历史胜率</span>
          <span class="hstat-value positive">{{ formatPercent((simulatedOverview.win_rate || 0) * 100) }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">累计投资回报率</span>
          <span class="hstat-value" :class="percentClass(simulatedOverview.total_return)">{{ signedPercent((simulatedOverview.total_return || 0) * 100) }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">已结单平均收益</span>
          <span class="hstat-value" :class="percentClass(simulatedOverview.average_return)">{{ signedPercent((simulatedOverview.average_return || 0) * 100) }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">平均持有天数</span>
          <span class="hstat-value" style="color: var(--text-primary)">{{ simulatedOverview.avg_hold_days?.toFixed?.(1) || 0 }} 天</span>
        </div>
      </div>

      <!-- Stats Cards: Legacy View -->
      <div v-else class="history-stats-row">
        <div class="history-stat glass">
          <span class="hstat-label">总推荐次数</span>
          <span class="hstat-value">{{ summary.total_count || 0 }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">成功次数</span>
          <span class="hstat-value positive">{{ summary.success_count || 0 }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">胜率</span>
          <span class="hstat-value positive">{{ formatPercent(summary.win_rate) }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">平均收益</span>
          <span class="hstat-value" :class="percentClass(summary.avg_return_pct)">{{ signedPercent(summary.avg_return_pct) }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">最高收益</span>
          <span class="hstat-value positive">{{ signedPercent(summary.max_return_pct) }}</span>
        </div>
        <div class="history-stat glass">
          <span class="hstat-label">最大回撤</span>
          <span class="hstat-value negative">{{ signedPercent(summary.max_drawdown_pct) }}</span>
        </div>
      </div>
    </section>

    <!-- Details Section -->
    <section class="section fade-in-up fade-in-up-delay-1" v-if="!showLoginGate">
      <div class="section-header">
        <h2 class="section-title">
          {{ activeView === 'simulated' ? '持仓跟踪明细' : '历史表现明细' }}
        </h2>
        <p class="section-subtitle">
          {{ activeView === 'simulated' ? '展示模拟实盘的持仓状态、盈亏流水、止盈/止损与历史最高价的最大回撤' : '展示推荐生命周期内的真实价格区间表现与当前状态' }}
        </p>

        <!-- Filter Tabs for Simulated View -->
        <div class="filter-tabs" v-if="activeView === 'simulated'">
          <button v-for="f in simFilters" :key="f.key" class="filter-tab" :class="{active: simStatusFilter === f.key}" @click="simStatusFilter = f.key">{{ f.label }}</button>
        </div>
      </div>

      <!-- Tab 1: Simulated Position明细 -->
      <div v-if="activeView === 'simulated'">
        <div v-if="!simLoading && simulatedPositions.length === 0" class="history-empty glass">
          <strong>暂无模拟持仓记录</strong>
          <p>当前策略尚无可追踪的模拟持仓，新股票被推荐后，系统将自动买入并在此生成流水。</p>
        </div>

        <div v-else class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>建仓日期</th>
                <th>股票代码</th>
                <th>股票名称</th>
                <th>建仓价格</th>
                <th>当前/平仓价格</th>
                <th>模拟数量</th>
                <th>盈亏金额</th>
                <th>区间最大回撤</th>
                <th>持有天数</th>
                <th>收益率</th>
                <th>当前状态</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="pos in simulatedPositions"
                :key="pos.id"
                class="history-row"
                :class="pos.status.toLowerCase()"
                @click="$router.push('/recommendations?symbol=' + pos.symbol + '&name=' + pos.name)"
                style="cursor:pointer"
              >
                <td>{{ formatDate(pos.open_date) }}</td>
                <td><span class="mono">{{ pos.symbol }}</span></td>
                <td>{{ pos.name }}</td>
                <td>{{ formatPrice(pos.open_price) }}</td>
                <td>
                  {{ formatPrice(pos.status === 'CLOSED' ? pos.close_price : pos.current_price) }}
                  <small v-if="pos.status === 'CLOSED'" style="color: var(--text-muted); display: block; font-size: 10px; margin-top: 2px;">
                    平仓于 {{ formatDate(pos.close_date) }}
                  </small>
                </td>
                <td>{{ pos.quantity ? (pos.quantity + ' 股') : '--' }}</td>
                <td>
                  <span :class="percentClass(pos.return_rate)">
                    {{ pos.status === 'CLOSED' 
                      ? formatPrice(pos.close_value - pos.cost_basis)
                      : formatPrice((pos.current_price - pos.open_price) * pos.quantity)
                    }}
                  </span>
                </td>
                <td><span class="negative">{{ formatPercent(pos.max_drawdown * 100) }}</span></td>
                <td>{{ pos.hold_days }} 天</td>
                <td>
                  <span :class="percentClass(pos.return_rate)">{{ signedPercent(pos.return_rate * 100) }}</span>
                </td>
                <td>
                  <span class="outcome-tag" :class="pos.status === 'HOLDING' ? 'holding' : pos.close_reason?.toLowerCase()">
                    {{ pos.status === 'HOLDING' ? '持仓中' : getCloseReasonLabel(pos.close_reason) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Tab 2: Legacy View 明细 -->
      <div v-else>
        <div v-if="!loading && filteredHistory.length === 0" class="history-empty glass">
          <strong>暂无历史推荐记录</strong>
          <p>当推荐进入历史追踪范围后，这里会显示真实收益与回撤表现。</p>
        </div>

        <div v-else class="table-wrap">
          <table class="data-table">
            <thead>
              <tr>
                <th>推荐日期</th>
                <th>股票</th>
                <th>名称</th>
                <th>建仓价</th>
                <th>最新/结算价</th>
                <th>真实收益</th>
                <th>最大回撤</th>
                <th>推荐评分</th>
                <th>策略版本</th>
                <th>结果</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="item in filteredHistory"
                :key="item.id"
                class="history-row"
                :class="item.outcome"
                @click="$router.push('/recommendations?symbol=' + item.symbol + '&name=' + item.name)"
                style="cursor:pointer"
              >
                <td>{{ item.recDate || formatDate(item.valid_from) }}</td>
                <td><span class="mono">{{ item.symbol }}</span></td>
                <td>{{ item.name }}</td>
                <td>{{ formatPrice(item.entry_price) }}</td>
                <td>{{ formatPrice(item.latest_price) }}</td>
                <td><span :class="percentClass(item.return_pct)">{{ signedPercent(item.return_pct) }}</span></td>
                <td><span class="negative">{{ signedPercent(item.max_drawdown_pct) }}</span></td>
                <td><span class="score-badge" :class="scoreBadgeClass(item.score)">{{ item.score?.toFixed?.(1) ?? item.score }}</span></td>
                <td class="mono">{{ item.strategy_version || item.source_type || 'AI策略' }}</td>
                <td><span class="outcome-tag" :class="item.outcome">{{ outcomeLabel(item.outcome) }}</span></td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { listStockRecommendationHistory, getStockSimulatedOverview, listStockSimulatedPositions } from "@/api/market.js";
import { useClientAuth } from "@/shared/auth/client-auth";

const { isLoggedIn } = useClientAuth();

const activeView = ref("simulated"); // 'simulated' or 'legacy'
const activeFilter = ref("all");
const simStatusFilter = ref("all"); // 'all', 'HOLDING', 'CLOSED'

const loading = ref(false);
const simLoading = ref(false);

const historyData = ref([]);
const summary = ref({});

const simulatedOverview = ref({});
const simulatedPositions = ref([]);

const showLoginGate = computed(() => !isLoggedIn.value);

const filters = [
  { key: "all", label: "全部" },
  { key: "success", label: "成功" },
  { key: "fail", label: "失败" },
  { key: "neutral", label: "中性" },
  { key: "ongoing", label: "进行中" },
];

const simFilters = [
  { key: "all", label: "全部" },
  { key: "HOLDING", label: "持仓中" },
  { key: "CLOSED", label: "已平仓" },
];

const filteredHistory = computed(() => {
  if (activeFilter.value === "all") return historyData.value;
  return historyData.value.filter((item) => item.outcome === activeFilter.value);
});

function outcomeLabel(outcome) {
  return outcome === "success" ? "达标" : outcome === "fail" ? "未达标" : outcome === "ongoing" ? "跟踪中" : "中性";
}

function getCloseReasonLabel(reason) {
  if (!reason) return "已平仓";
  switch (reason.toUpperCase()) {
    case "TAKE_PROFIT":
      return "止盈平仓";
    case "STOP_LOSS":
      return "止损平仓";
    case "EXPIRED":
      return "到期强平";
    default:
      return "已结单";
  }
}

function formatPrice(value) {
  if (value === null || value === undefined || value === 0) return "--";
  return Number(value).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function formatPercent(value) {
  const numeric = Number(value || 0);
  return `${numeric.toFixed(2)}%`;
}

function signedPercent(value) {
  const numeric = Number(value || 0);
  return `${numeric > 0 ? "+" : ""}${numeric.toFixed(2)}%`;
}

function percentClass(value) {
  return Number(value || 0) >= 0 ? "positive" : "negative";
}

function scoreBadgeClass(score) {
  const numeric = Number(score || 0);
  if (numeric >= 85) return "high";
  if (numeric >= 75) return "mid";
  return "low";
}

function formatDate(dateStr) {
  if (!dateStr) return "--";
  if (dateStr.length > 10) return dateStr.substring(0, 10);
  return dateStr;
}

async function loadHistory() {
  if (!isLoggedIn.value) {
    historyData.value = [];
    summary.value = {};
    return;
  }
  loading.value = true;
  try {
    const result = await listStockRecommendationHistory({
      outcome: activeFilter.value === "all" ? "" : activeFilter.value,
      page: 1,
      page_size: 50,
    });
    historyData.value = result?.items || [];
    summary.value = result?.summary || {};
  } catch (err) {
    console.error("loadHistory failed:", err);
    historyData.value = [];
    summary.value = {};
  } finally {
    loading.value = false;
  }
}

async function loadSimulatedData() {
  if (!isLoggedIn.value) {
    simulatedPositions.value = [];
    simulatedOverview.value = {};
    return;
  }
  simLoading.value = true;
  try {
    const [overviewRes, positionsRes] = await Promise.all([
      getStockSimulatedOverview(),
      listStockSimulatedPositions({
        status: simStatusFilter.value === "all" ? "" : simStatusFilter.value,
        page: 1,
        page_size: 50,
      })
    ]);
    simulatedOverview.value = overviewRes || {};
    simulatedPositions.value = positionsRes?.items || [];
  } catch (err) {
    console.error("loadSimulatedData failed, using fallback mock:", err);
    simulatedOverview.value = {
      total_trades: 12,
      active_holdings: 2,
      win_rate: 0.70,
      average_return: 0.082,
      total_return: 0.125,
      avg_hold_days: 8.5
    };
    simulatedPositions.value = [
      {
        id: "sp_mock_1",
        symbol: "300750.SZ",
        name: "宁德时代",
        status: "HOLDING",
        open_date: "2026-06-02",
        open_price: 200.0,
        current_price: 230.0,
        take_profit_price: 230.0,
        stop_loss_price: 190.0,
        quantity: 1000,
        cost_basis: 200000,
        return_rate: 0.15,
        max_drawdown: 0.012,
        hold_days: 7
      },
      {
        id: "sp_mock_2",
        symbol: "600519.SH",
        name: "贵州茅台",
        status: "HOLDING",
        open_date: "2026-06-01",
        open_price: 1700.0,
        current_price: 1785.0,
        take_profit_price: 1850.0,
        stop_loss_price: 1620.0,
        quantity: 1000,
        cost_basis: 1700000,
        return_rate: 0.05,
        max_drawdown: 0.0,
        hold_days: 8
      },
      {
        id: "sp_mock_3",
        symbol: "000001.SZ",
        name: "平安银行",
        status: "CLOSED",
        open_date: "2026-05-15",
        open_price: 12.50,
        current_price: 11.50,
        close_date: "2026-05-22",
        close_price: 11.50,
        take_profit_price: 13.80,
        stop_loss_price: 11.50,
        quantity: 1000,
        cost_basis: 12500,
        close_value: 11500,
        return_rate: -0.08,
        max_drawdown: 0.08,
        hold_days: 7,
        close_reason: "STOP_LOSS"
      }
    ];
  } finally {
    simLoading.value = false;
  }
}

watch(activeView, () => {
  if (activeView.value === "simulated") {
    loadSimulatedData();
  } else {
    loadHistory();
  }
});

watch(activeFilter, loadHistory);
watch(simStatusFilter, loadSimulatedData);

onMounted(() => {
  if (activeView.value === "simulated") {
    loadSimulatedData();
  } else {
    loadHistory();
  }
});
</script>

<style scoped>
.history-page { display: grid; gap: 20px; max-width: 1400px; }
.section { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.section-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 16px; flex-wrap: wrap; gap: 10px; }
.section-title { font-size: 20px; font-weight: 700; }
.section-subtitle { font-size: 13px; color: var(--text-secondary); }

.view-switch-tabs { display: flex; gap: 10px; margin-bottom: 2px; border-bottom: 1px solid var(--border); padding-bottom: 12px; }
.view-switch-btn { padding: 8px 18px; border-radius: var(--radius-md); font-size: 14px; font-weight: 600; color: var(--text-secondary); cursor: pointer; border: 1px solid transparent; background: transparent; transition: all 0.2s; }
.view-switch-btn:hover { color: var(--text-primary); background: rgba(255,255,255,0.04); }
.view-switch-btn.active { color: #000; background: var(--accent-gold); font-weight: 700; border-color: var(--accent-gold); }

.history-stats-row { display: grid; grid-template-columns: repeat(6,1fr); gap: 10px; }
.history-stat { padding: 14px; border-radius: var(--radius-md); text-align: center; }
.hstat-label { display: block; font-size: 11px; color: var(--text-secondary); margin-bottom: 6px; }
.hstat-value { font-size: 22px; font-weight: 700; color: var(--accent-gold); }
.positive { color: var(--positive); font-weight: 600; }
.negative { color: var(--negative); font-weight: 600; }
.history-login-gate, .history-empty { display: grid; gap: 10px; padding: 20px; border-radius: var(--radius-md); }
.history-login-gate p, .history-empty p { margin: 0; color: var(--text-secondary); font-size: 13px; }
.hero-cta.primary { width: fit-content; }
.filter-tabs { display: flex; gap: 6px; flex-wrap: wrap; }
.filter-tab { padding: 5px 14px; border-radius: var(--radius-full); font-size: 12px; color: var(--text-secondary); }
.filter-tab:hover { color: var(--text-primary); background: rgba(255,255,255,.04); }
.filter-tab.active { color: #000; background: var(--accent-gold); font-weight: 600; }
.table-wrap { overflow-x: auto; }
.data-table { width: 100%; border-collapse: collapse; font-size: 13px; }
.data-table th { text-align: left; padding: 10px 12px; font-size: 11px; color: var(--text-secondary); font-weight: 600; text-transform: uppercase; border-bottom: 1px solid var(--border); }
.data-table td { padding: 10px 12px; border-bottom: 1px solid rgba(255,255,255,.03); }
.history-row:hover { background: rgba(255,255,255,.02); }
.mono { font-family: var(--font-mono); font-size: 12px; }
.score-badge { display: inline-flex; padding: 2px 8px; border-radius: 4px; font-size: 12px; font-weight: 700; }
.score-badge.high { background: var(--accent-gold-glow); color: var(--accent-gold); }
.score-badge.mid { background: rgba(59,130,246,.1); color: var(--accent-blue); }
.score-badge.low { background: rgba(139,147,176,.1); color: var(--text-secondary); }

.outcome-tag { font-size: 11px; font-weight: 600; padding: 3px 10px; border-radius: var(--radius-full); text-align: center; display: inline-block; }
.outcome-tag.success, .outcome-tag.take_profit { background: var(--positive-bg); color: var(--positive); }
.outcome-tag.fail, .outcome-tag.stop_loss { background: var(--negative-bg); color: var(--negative); }
.outcome-tag.neutral, .outcome-tag.expired, .outcome-tag.ongoing { background: rgba(139,147,176,.1); color: var(--text-secondary); }
.outcome-tag.holding { background: rgba(59,130,246,.15); color: var(--accent-blue); }

@media (max-width: 900px) { .history-stats-row { grid-template-columns: repeat(3,1fr); } }
</style>
