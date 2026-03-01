<template>
  <section class="strategy-page fade-up">
    <header class="strategy-hero card">
      <div>
        <p class="hero-kicker">策略中心 · API 实时同步</p>
        <h1 class="section-title">先看推荐，再看参数和风险。</h1>
        <p class="section-subtitle">
          已接入股票推荐与期货策略接口，优先展示可执行策略和仓位建议。
        </p>
        <div class="tag-row">
          <span v-for="tag in tags" :key="tag">{{ tag }}</span>
        </div>
        <p v-if="loading" class="api-state">正在同步策略数据...</p>
        <p v-else-if="errorMessage" class="api-state warning">{{ errorMessage }}</p>
        <p v-else class="api-state">更新时间：{{ lastUpdatedAt || "-" }}</p>
      </div>
      <div class="hero-actions">
        <button class="primary-btn" type="button" :disabled="loading" @click="loadStrategies">
          {{ loading ? "同步中..." : "刷新策略" }}
        </button>
      </div>
    </header>

    <div class="strategy-grid">
      <article class="card matrix-card">
        <header>
          <h2 class="section-title">股票推荐策略</h2>
          <p class="section-subtitle">按评分排序，突出风险等级和建议仓位。</p>
        </header>

        <div class="matrix-head">
          <span>策略名称</span>
          <span>评分</span>
          <span>风险</span>
          <span>仓位建议</span>
          <span>状态</span>
        </div>

        <div class="matrix-list" v-if="stockRows.length">
          <button
            v-for="item in stockRows"
            :key="item.id"
            type="button"
            class="matrix-item"
            :class="{ active: activeStockID === item.id }"
            @click="openStockDetail(item.id, $event)"
          >
            <div>
              <p class="name">{{ item.name }}</p>
              <p class="desc">{{ item.desc }}</p>
            </div>
            <p class="score">{{ item.score }}</p>
            <p>{{ item.risk }}</p>
            <p>{{ item.position }}</p>
            <span class="status" :class="item.statusClass">{{ item.status }}</span>
          </button>
        </div>
        <p v-if="stockRows.length" class="matrix-tip">点击任意推荐可查看推荐理由和数据支撑。</p>
        <div v-else class="empty-box">暂无股票推荐策略</div>
      </article>

      <aside class="card cycle-card">
        <header>
          <h2 class="section-title">期货策略清单</h2>
          <p class="section-subtitle">每个策略保留关键执行参数。</p>
        </header>
        <div class="cycle-list" v-if="futuresRows.length">
          <button
            v-for="step in futuresRows"
            :key="step.id"
            type="button"
            class="cycle-item"
            :class="{ active: activeFuturesID === step.id }"
            @click="openFuturesInsight(step.id, $event)"
          >
            <p class="step">{{ step.contract }}</p>
            <p class="title">{{ step.title }}</p>
            <p class="note">{{ step.note }}</p>
          </button>
        </div>
        <p v-if="futuresRows.length" class="matrix-tip">点击任意期货策略可查看评分体系与执行细节。</p>
        <div v-else class="empty-box">暂无期货策略</div>
      </aside>
    </div>

    <section class="event-section">
      <article class="card event-list-card">
        <header>
          <h2 class="section-title">市场事件雷达</h2>
          <p class="section-subtitle">接入 market/events，统一查看策略触发事件。</p>
        </header>

        <div class="event-type-tabs">
          <button
            v-for="item in eventTypeOptions"
            :key="item.value"
            type="button"
            :class="{ active: activeEventType === item.value }"
            @click="activeEventType = item.value"
          >
            {{ item.label }}
          </button>
        </div>

        <div class="event-list" v-if="filteredEventRows.length > 0">
          <button
            v-for="item in filteredEventRows"
            :key="item.id"
            type="button"
            class="event-item"
            :class="{ active: activeEventID === item.id }"
            @click="activeEventID = item.id"
          >
            <div class="event-top">
              <p>{{ item.symbol }}</p>
              <span class="event-level" :class="item.levelClass">{{ item.levelText }}</span>
            </div>
            <p class="event-title">{{ item.title }}</p>
            <p class="event-meta">{{ item.typeLabel }} · {{ item.time }}</p>
          </button>
        </div>
        <div v-else class="empty-box">当前条件下暂无市场事件</div>
      </article>

      <article class="card event-detail-card">
        <header>
          <h2 class="section-title">事件详情</h2>
          <p class="section-subtitle">结合触发规则判断是否需要调整策略。</p>
        </header>

        <div v-if="activeEventView" class="event-detail">
          <p class="detail-kicker">
            {{ activeEventView.typeLabel }} · {{ activeEventView.symbol }} · {{ activeEventView.time }}
          </p>
          <h3>{{ activeEventView.title }}</h3>
          <p class="detail-summary">{{ activeEventView.summary }}</p>

          <div class="detail-grid">
            <article>
              <p>触发规则</p>
              <strong>{{ activeEventView.triggerRule }}</strong>
            </article>
            <article>
              <p>来源</p>
              <strong>{{ activeEventView.source || "-" }}</strong>
            </article>
          </div>

          <div class="detail-actions">
            <button type="button" :disabled="eventDetailLoading" @click="refreshActiveEventDetail">
              {{ eventDetailLoading ? "同步中..." : "刷新事件详情" }}
            </button>
          </div>
        </div>
        <div v-else class="empty-box">暂无可展示的事件详情</div>
      </article>
    </section>

    <transition name="detail-fade">
      <div
        v-if="stockDetailDialogVisible && activeStockView"
        class="stock-detail-dialog-mask"
        @click.self="closeStockDetail"
      >
        <section ref="stockDetailDialogRef" class="stock-detail-dialog card" :style="stockDetailDialogStyle">
          <header class="stock-detail-head">
            <div>
              <p class="detail-kicker">推荐详情</p>
              <h3>{{ activeStockView.name }}</h3>
            </div>
            <div class="stock-detail-actions">
              <button type="button" :disabled="stockDetailLoading" @click="refreshActiveStockDetail">
                {{ stockDetailLoading ? "同步中..." : "刷新业绩" }}
              </button>
              <button type="button" class="close-btn" @click="closeStockDetail">关闭</button>
            </div>
          </header>

          <p class="stock-reason">{{ activeStockView.reason }}</p>
          <p v-if="activeStockScoreFramework?.method" class="stock-reason-meta">
            评分模型：{{ activeStockScoreFramework.method }}
          </p>

          <div class="stock-detail-grid">
            <article>
              <p>综合评分</p>
              <strong>{{ activeStockView.score }}</strong>
            </article>
            <article>
              <p>风险等级</p>
              <strong>{{ activeStockView.risk }}</strong>
            </article>
            <article>
              <p>建议仓位</p>
              <strong>{{ activeStockView.position }}</strong>
            </article>
            <article>
              <p>有效时间</p>
              <strong>{{ activeStockView.validRange }}</strong>
            </article>
            <article>
              <p>止盈建议</p>
              <strong>{{ activeStockView.takeProfit }}</strong>
            </article>
            <article>
              <p>止损建议</p>
              <strong>{{ activeStockView.stopLoss }}</strong>
            </article>
          </div>

          <div class="reason-support-grid">
            <article v-for="item in activeStockReasonSupports" :key="item.label">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div class="stock-news-box">
            <div class="stock-news-head">
              <p>相关资讯支撑</p>
              <span>{{ activeStockRelatedNews.length }} 条</span>
            </div>
            <div v-if="stockRelatedNewsLoading" class="empty-inline">正在同步相关资讯...</div>
            <div v-else-if="activeStockRelatedNews.length > 0" class="stock-news-list">
              <article v-for="item in activeStockRelatedNews" :key="item.id" class="stock-news-item">
                <h4>{{ item.title }}</h4>
                <p class="stock-news-meta">
                  {{ item.source }} · {{ item.time }} · {{ item.visibility }} · 相关度 {{ formatScore01(item.relevanceScore) }}
                </p>
                <p class="stock-news-summary">{{ item.summary }}</p>
              </article>
            </div>
            <div v-else class="empty-inline">暂无可匹配的资讯数据</div>
          </div>

          <div class="stock-performance-box">
            <div class="stock-performance-head">
              <p>历史推荐业绩</p>
              <span>{{ activeStockPerformanceSummary }}</span>
            </div>
            <div v-if="activeStockPerformanceRows.length > 0" class="performance-table-wrap">
              <table class="performance-table">
                <thead>
                  <tr>
                    <th>日期</th>
                    <th>单日收益</th>
                    <th>累计收益</th>
                    <th>基准累计</th>
                    <th>累计超额</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in activeStockPerformanceRows" :key="`${activeStockView.id}-${item.date}`">
                    <td>{{ item.date }}</td>
                    <td :class="item.dailyClass">{{ item.dailyReturn }}</td>
                    <td :class="item.cumulativeClass">{{ item.cumulativeReturn }}</td>
                    <td :class="item.benchmarkClass">{{ item.benchmarkReturn }}</td>
                    <td :class="item.excessClass">{{ item.excessReturn }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="empty-inline">暂无历史推荐业绩</div>
          </div>

          <p v-if="stockDetailErrorMessage" class="detail-warning">{{ stockDetailErrorMessage }}</p>
        </section>
      </div>
    </transition>

    <transition name="detail-fade">
      <div
        v-if="futuresDetailDialogVisible && activeFuturesView"
        class="futures-detail-dialog-mask"
        @click.self="closeFuturesDetail"
      >
        <section ref="futuresDetailDialogRef" class="futures-detail-dialog card" :style="futuresDetailDialogStyle">
          <header class="stock-detail-head">
            <div>
              <p class="detail-kicker">期货策略详情</p>
              <h3>{{ activeFuturesView.name }}</h3>
            </div>
            <div class="stock-detail-actions">
              <button type="button" :disabled="futuresInsightLoading" @click="refreshActiveFuturesInsight">
                {{ futuresInsightLoading ? "同步中..." : "刷新详情" }}
              </button>
              <button type="button" class="close-btn" @click="closeFuturesDetail">关闭</button>
            </div>
          </header>

          <p class="stock-reason">{{ activeFuturesView.reason }}</p>
          <p v-if="activeFuturesScoreFramework?.method" class="stock-reason-meta">
            评分模型：{{ activeFuturesScoreFramework.method }}
          </p>

          <div class="futures-detail-grid">
            <article>
              <p>方向</p>
              <strong>{{ activeFuturesView.direction }}</strong>
            </article>
            <article>
              <p>风险等级</p>
              <strong>{{ activeFuturesView.risk }}</strong>
            </article>
            <article>
              <p>建议仓位</p>
              <strong>{{ activeFuturesView.position }}</strong>
            </article>
            <article>
              <p>有效时间</p>
              <strong>{{ activeFuturesView.validRange }}</strong>
            </article>
            <article>
              <p>入场区间</p>
              <strong>{{ activeFuturesView.entryRange }}</strong>
            </article>
            <article>
              <p>止盈区间</p>
              <strong>{{ activeFuturesView.takeProfitRange }}</strong>
            </article>
            <article>
              <p>止损区间</p>
              <strong>{{ activeFuturesView.stopLossRange }}</strong>
            </article>
            <article>
              <p>失效条件</p>
              <strong>{{ activeFuturesView.invalidCondition }}</strong>
            </article>
          </div>

          <div class="reason-support-grid">
            <article v-for="item in activeFuturesReasonSupports" :key="`fut-${item.label}`">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div class="stock-news-box">
            <div class="stock-news-head">
              <p>相关资讯支撑</p>
              <span>{{ activeFuturesRelatedNews.length }} 条</span>
            </div>
            <div v-if="futuresInsightLoading" class="empty-inline">正在同步相关资讯...</div>
            <div v-else-if="activeFuturesRelatedNews.length > 0" class="stock-news-list">
              <article v-for="item in activeFuturesRelatedNews" :key="`fut-news-${item.id}`" class="stock-news-item">
                <h4>{{ item.title }}</h4>
                <p class="stock-news-meta">
                  {{ item.source }} · {{ formatDateTime(item.published_at || item.time) }} ·
                  {{ mapVisibilityLabel(item.visibility) }} · 相关度 {{ formatScore01(item.relevance_score ?? item.relevanceScore) }}
                </p>
                <p class="stock-news-summary">{{ summarizeNewsText(item.summary) }}</p>
              </article>
            </div>
            <div v-else class="empty-inline">暂无可匹配资讯</div>
          </div>

          <div class="futures-event-box">
            <div class="stock-news-head">
              <p>相关市场事件</p>
              <span>{{ activeFuturesRelatedEvents.length }} 条</span>
            </div>
            <div v-if="activeFuturesRelatedEvents.length > 0" class="futures-event-list">
              <article v-for="item in activeFuturesRelatedEvents" :key="`fut-event-${item.id}`" class="futures-event-item">
                <p>{{ mapEventType(item.event_type || item.type) }} · {{ item.symbol || "-" }}</p>
                <strong>{{ item.summary || "-" }}</strong>
                <span>{{ formatDateTime(item.created_at || item.time) }}</span>
              </article>
            </div>
            <div v-else class="empty-inline">暂无相关事件</div>
          </div>

          <div class="stock-performance-box">
            <div class="stock-performance-head">
              <p>历史策略业绩</p>
              <span>{{ activeFuturesPerformanceSummary }}</span>
            </div>
            <div v-if="activeFuturesPerformanceRows.length > 0" class="performance-table-wrap">
              <table class="performance-table">
                <thead>
                  <tr>
                    <th>日期</th>
                    <th>单日收益</th>
                    <th>累计收益</th>
                    <th>基准累计</th>
                    <th>累计超额</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in activeFuturesPerformanceRows" :key="`fut-perf-${activeFuturesView.id}-${item.date}`">
                    <td>{{ item.date }}</td>
                    <td :class="item.dailyClass">{{ item.dailyReturn }}</td>
                    <td :class="item.cumulativeClass">{{ item.cumulativeReturn }}</td>
                    <td :class="item.benchmarkClass">{{ item.benchmarkReturn }}</td>
                    <td :class="item.excessClass">{{ item.excessReturn }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div v-else class="empty-inline">暂无历史策略业绩</div>
          </div>

          <p v-if="futuresInsightErrorMessage" class="detail-warning">{{ futuresInsightErrorMessage }}</p>
        </section>
      </div>
    </transition>
  </section>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import {
  getFuturesStrategyInsight,
  getStockRecommendationInsight,
  getStockRecommendationDetail,
  getStockRecommendationPerformance,
  getMarketEventDetail,
  listFuturesStrategies,
  listMarketEvents,
  listStockRecommendations
} from "../api/market";
import { listNewsArticles } from "../api/news";
import { shouldUseDemoFallback } from "../lib/fallback-policy";

const loading = ref(false);
const errorMessage = ref("");
const lastUpdatedAt = ref("");

const useDemoFallback = shouldUseDemoFallback();

const rawStocks = ref(
  useDemoFallback
    ? [
        {
          id: "sr_local_001",
          symbol: "600519.SH",
          name: "贵州茅台",
          score: 91.2,
          risk_level: "MEDIUM",
          position_range: "10%-15%",
          valid_from: "2026-02-27T09:30:00+08:00",
          valid_to: "2026-03-04T15:00:00+08:00",
          status: "PUBLISHED",
          reason_summary: "基本面和资金流共振"
        }
      ]
    : []
);

const rawFutures = ref(
  useDemoFallback
    ? [
        {
          id: "fs_local_001",
          contract: "IF2603",
          name: "股指趋势跟踪",
          direction: "LONG",
          risk_level: "MEDIUM",
          position_range: "20%-30%",
          status: "PUBLISHED",
          reason_summary: "趋势与量价结构一致"
        }
      ]
    : []
);

const fallbackStockDetails = {
  sr_local_001: {
    reco_id: "sr_local_001",
    tech_score: 89.4,
    fund_score: 92.1,
    sentiment_score: 84.8,
    money_flow_score: 90.6,
    take_profit: "上涨 6%-10% 分批止盈",
    stop_loss: "跌破 1590 止损",
    risk_note: "白酒板块受政策与资金风格影响，建议分批进出。"
  }
};

const fallbackStockPerformance = {
  sr_local_001: [
    { date: "2026-02-24", return: 0.009 },
    { date: "2026-02-25", return: 0.012 },
    { date: "2026-02-26", return: -0.004 },
    { date: "2026-02-27", return: 0.011 }
  ]
};

const fallbackStockRelatedNews = {
  sr_local_001: [
    {
      id: "news_local_001",
      title: "白酒板块资金回流，龙头股维持高景气度",
      source: "市场快讯",
      time: "2026-02-27 09:41:00",
      visibility: "公开",
      summary: "主力资金连续净流入，机构维持中长期配置观点。"
    },
    {
      id: "news_local_002",
      title: "消费修复预期增强，龙头估值回归历史中枢",
      source: "券商研报",
      time: "2026-02-26 20:30:00",
      visibility: "VIP",
      summary: "盈利稳定性改善，渠道反馈显示需求恢复节奏较稳。"
    }
  ]
};

const fallbackStockBenchmark = {
  sr_local_001: [
    { date: "2026-02-24", return: 0.005 },
    { date: "2026-02-25", return: 0.007 },
    { date: "2026-02-26", return: -0.002 },
    { date: "2026-02-27", return: 0.006 }
  ]
};

const fallbackStockScoreFramework = {
  sr_local_001: {
    method: "growth-v1 (tech30 + fund30 + sentiment20 + flow20)",
    total_score: 91.2,
    weighted_score: 89.7,
    score_gap: 1.5,
    factors: [
      { key: "tech", label: "技术因子", weight: 0.3, score: 89.4, contribution: 26.82 },
      { key: "fund", label: "基本面因子", weight: 0.3, score: 92.1, contribution: 27.63 },
      { key: "sentiment", label: "情绪因子", weight: 0.2, score: 84.8, contribution: 16.96 },
      { key: "flow", label: "资金流因子", weight: 0.2, score: 90.6, contribution: 18.12 }
    ]
  }
};

const fallbackStockPerformanceStats = {
  sr_local_001: {
    sample_days: 4,
    win_rate: 0.75,
    avg_daily_return: 0.007,
    cumulative_return: 0.0281,
    benchmark_cumulative_return: 0.0161,
    excess_return: 0.012,
    max_drawdown: 0.004,
    benchmark_symbol: "000300.SH",
    benchmark_source: "estimated: 55% of strategy daily return"
  }
};

const fallbackFuturesInsights = {
  fs_local_001: {
    strategy: {
      id: "fs_local_001",
      contract: "IF2603",
      name: "股指趋势跟踪",
      direction: "LONG",
      risk_level: "MEDIUM",
      position_range: "20%-30%",
      valid_from: "2026-02-27T09:30:00+08:00",
      valid_to: "2026-03-04T15:00:00+08:00",
      status: "PUBLISHED",
      reason_summary: "趋势与量价结构一致，等待突破确认后执行。"
    },
    guidance: {
      contract: "IF2603",
      guidance_direction: "LONG_SPREAD",
      position_level: "MEDIUM",
      entry_range: "3520 - 3545",
      take_profit_range: "3590 - 3620",
      stop_loss_range: "3470 - 3490",
      risk_level: "MEDIUM",
      invalid_condition: "15分钟跌破3470且量能放大",
      valid_to: "2026-03-04T15:00:00+08:00"
    },
    score_framework: {
      method: "futures-v1 (trend25 + structure20 + flow15 + risk20 + news10 + performance10)",
      total_score: 84.6,
      weighted_score: 83.9,
      score_gap: 0.7,
      factors: [
        { key: "trend", label: "趋势因子", weight: 0.25, score: 86, contribution: 21.5 },
        { key: "structure", label: "结构因子", weight: 0.2, score: 82, contribution: 16.4 },
        { key: "flow", label: "资金与事件因子", weight: 0.15, score: 80, contribution: 12 },
        { key: "risk", label: "风险控制因子", weight: 0.2, score: 84, contribution: 16.8 },
        { key: "news", label: "资讯因子", weight: 0.1, score: 81, contribution: 8.1 },
        { key: "performance", label: "绩效因子", weight: 0.1, score: 91, contribution: 9.1 }
      ]
    },
    related_news: [
      {
        id: "news_fut_local_001",
        title: "股指期货成交放量，短线趋势延续",
        summary: "IF 主力合约成交量明显上升，量价结构偏多。",
        source: "新闻快讯",
        visibility: "PUBLIC",
        published_at: "2026-02-27T10:20:00+08:00",
        relevance_score: 0.86
      }
    ],
    related_events: [
      {
        id: "me_fut_local_001",
        event_type: "PRICE",
        symbol: "IF2603",
        summary: "15分钟级别突破关键压力位。",
        trigger_rule: "15分钟涨幅超过1.2%",
        source: "system",
        created_at: "2026-02-27T10:16:00+08:00"
      }
    ],
    performance: [
      { date: "2026-02-24", return: 0.008 },
      { date: "2026-02-25", return: 0.012 },
      { date: "2026-02-26", return: -0.004 },
      { date: "2026-02-27", return: 0.011 }
    ],
    benchmark: [
      { date: "2026-02-24", return: 0.004 },
      { date: "2026-02-25", return: 0.006 },
      { date: "2026-02-26", return: -0.002 },
      { date: "2026-02-27", return: 0.005 }
    ],
    performance_stats: {
      sample_days: 4,
      win_rate: 0.75,
      avg_daily_return: 0.0068,
      cumulative_return: 0.0271,
      benchmark_cumulative_return: 0.0131,
      excess_return: 0.014,
      max_drawdown: 0.004,
      benchmark_symbol: "000300.SH",
      benchmark_source: "actual: CSI300"
    },
    generated_at: "2026-02-27T10:30:00+08:00"
  }
};

const fallbackEvents = [
  {
    id: "me_local_001",
    event_type: "PRICE",
    symbol: "IF2603",
    summary: "股指期货盘中突破关键压力位，短线情绪转强。",
    trigger_rule: "15分钟涨幅超过 1.2%",
    source: "system",
    created_at: "2026-02-27T10:16:00+08:00"
  },
  {
    id: "me_local_002",
    event_type: "VOLUME",
    symbol: "IC2603",
    summary: "成交量明显放大，价量共振信号增强。",
    trigger_rule: "5分钟量能超过近20周期均值 2 倍",
    source: "system",
    created_at: "2026-02-27T10:28:00+08:00"
  },
  {
    id: "me_local_003",
    event_type: "POLICY",
    symbol: "ALL",
    summary: "交易所发布盘后保证金调整通知。",
    trigger_rule: "重要规则变更监控命中",
    source: "exchange",
    created_at: "2026-02-27T08:42:00+08:00"
  }
];

const fallbackEventDetails = {
  me_local_001: {
    id: "me_local_001",
    event_type: "PRICE",
    symbol: "IF2603",
    summary: "指数期货突破后，短线可提高跟随策略权重。",
    trigger_rule: "15分钟涨幅超过 1.2%",
    source: "system",
    created_at: "2026-02-27T10:16:00+08:00"
  },
  me_local_002: {
    id: "me_local_002",
    event_type: "VOLUME",
    symbol: "IC2603",
    summary: "量能激增说明资金参与度提升，策略执行可适度前置。",
    trigger_rule: "5分钟量能超过近20周期均值 2 倍",
    source: "system",
    created_at: "2026-02-27T10:28:00+08:00"
  },
  me_local_003: {
    id: "me_local_003",
    event_type: "POLICY",
    symbol: "ALL",
    summary: "政策类事件会改变风险参数，建议先审查仓位与保证金占用。",
    trigger_rule: "重要规则变更监控命中",
    source: "exchange",
    created_at: "2026-02-27T08:42:00+08:00"
  }
};

const rawEvents = ref(useDemoFallback ? [...fallbackEvents] : []);
const eventDetailMap = ref(useDemoFallback ? { ...fallbackEventDetails } : {});
const loadedEventDetailIDs = new Set();
const stockDetailMap = ref(useDemoFallback ? { ...fallbackStockDetails } : {});
const stockPerformanceMap = ref(useDemoFallback ? { ...fallbackStockPerformance } : {});
const stockBenchmarkMap = ref(useDemoFallback ? { ...fallbackStockBenchmark } : {});
const stockRelatedNewsMap = ref(useDemoFallback ? { ...fallbackStockRelatedNews } : {});
const stockScoreFrameworkMap = ref(useDemoFallback ? { ...fallbackStockScoreFramework } : {});
const stockPerformanceStatsMap = ref(useDemoFallback ? { ...fallbackStockPerformanceStats } : {});
const loadedStockInsightIDs = new Set();
const loadedStockDetailIDs = new Set();
const loadedStockPerformanceIDs = new Set();
const loadedStockRelatedNewsIDs = new Set();
const activeStockID = ref(useDemoFallback ? rawStocks.value[0]?.id || "" : "");
const stockDetailLoading = ref(false);
const stockDetailErrorMessage = ref("");
const stockDetailDialogVisible = ref(false);
const stockDetailDialogRef = ref(null);
const stockDetailDialogStyle = ref({});
const stockDetailAnchorRect = ref(null);
const stockRelatedNewsLoading = ref(false);
const futuresInsightMap = ref(useDemoFallback ? { ...fallbackFuturesInsights } : {});
const loadedFuturesInsightIDs = new Set();
const activeFuturesID = ref(useDemoFallback ? rawFutures.value[0]?.id || "" : "");
const futuresDetailDialogVisible = ref(false);
const futuresDetailDialogRef = ref(null);
const futuresDetailDialogStyle = ref({});
const futuresDetailAnchorRect = ref(null);
const futuresInsightLoading = ref(false);
const futuresInsightErrorMessage = ref("");
const activeEventType = ref("ALL");
const activeEventID = ref(useDemoFallback ? fallbackEvents[0].id : "");
const eventDetailLoading = ref(false);

const stockRows = computed(() =>
  (rawStocks.value || []).map((item) => ({
    id: item.id,
    name: `${item.symbol || "-"} ${item.name || ""}`.trim(),
    desc: item.reason_summary || "-",
    score: formatScore(item.score),
    risk: mapRisk(item.risk_level),
    position: item.position_range || "-",
    status: mapStatus(item.status).label,
    statusClass: mapStatus(item.status).className
  }))
);

const activeStockBase = computed(() => {
  const items = rawStocks.value || [];
  if (items.length === 0) {
    return null;
  }
  if (activeStockID.value) {
    const matched = items.find((item) => item.id === activeStockID.value);
    if (matched) {
      return matched;
    }
  }
  return items[0];
});

const activeStockView = computed(() => {
  const base = activeStockBase.value;
  if (!base?.id) {
    return null;
  }
  const detail = stockDetailMap.value[base.id] || {};
  const reasonParts = [base.reason_summary, detail.risk_note, buildStockScoreNote(detail)].filter(Boolean);
  return {
    id: base.id,
    name: `${base.symbol || "-"} ${base.name || ""}`.trim(),
    reason: reasonParts.join(" "),
    score: formatScore(base.score),
    risk: mapRisk(base.risk_level),
    position: base.position_range || "-",
    validRange: formatDateRange(base.valid_from, base.valid_to),
    takeProfit: detail.take_profit || "待补充",
    stopLoss: detail.stop_loss || "待补充"
  };
});

const activeStockPerformanceRows = computed(() => {
  const stockID = activeStockView.value?.id;
  if (!stockID) {
    return [];
  }
  const points = Array.isArray(stockPerformanceMap.value[stockID])
    ? [...stockPerformanceMap.value[stockID]]
    : [];
  const benchmarkPoints = Array.isArray(stockBenchmarkMap.value[stockID])
    ? [...stockBenchmarkMap.value[stockID]]
    : [];
  if (points.length === 0) {
    return [];
  }

  points.sort((a, b) => compareDateAsc(a?.date, b?.date));
  const benchmarkMap = new Map();
  benchmarkPoints.forEach((item) => {
    if (item?.date) {
      benchmarkMap.set(item.date, Number(item.return));
    }
  });

  let acc = 1;
  let benchmarkAcc = 1;
  let hasValidValue = false;
  return points.map((point) => {
    const daily = Number(point?.return);
    const benchmarkDaily = benchmarkMap.has(point?.date) ? Number(benchmarkMap.get(point?.date)) : NaN;
    if (Number.isFinite(daily)) {
      acc *= 1 + daily;
      hasValidValue = true;
    }
    if (Number.isFinite(benchmarkDaily)) {
      benchmarkAcc *= 1 + benchmarkDaily;
    }
    const cumulative = hasValidValue ? acc - 1 : null;
    const benchmarkCumulative = Number.isFinite(benchmarkDaily) ? benchmarkAcc - 1 : null;
    const excess = Number.isFinite(cumulative) && Number.isFinite(benchmarkCumulative)
      ? cumulative - benchmarkCumulative
      : null;
    return {
      date: formatDate(point?.date),
      dailyReturn: formatPercent(daily),
      cumulativeReturn: formatPercent(cumulative),
      benchmarkReturn: formatPercent(benchmarkCumulative),
      excessReturn: formatPercent(excess),
      dailyClass: trendClassByNumber(daily),
      cumulativeClass: trendClassByNumber(cumulative),
      benchmarkClass: trendClassByNumber(benchmarkCumulative),
      excessClass: trendClassByNumber(excess),
      dailyRaw: daily,
      cumulativeRaw: cumulative,
      benchmarkRaw: benchmarkCumulative,
      excessRaw: excess
    };
  });
});

const activeStockPerformanceSummary = computed(() => {
  const stockID = activeStockView.value?.id;
  if (stockID) {
    const stats = stockPerformanceStatsMap.value[stockID];
    if (stats) {
      const benchmarkLabel = stats.benchmark_symbol ? `基准(${stats.benchmark_symbol})` : "基准";
      return `样本 ${Number(stats.sample_days || 0)} 日 · 胜率 ${formatRate(stats.win_rate)} · 累计 ${formatPercent(
        stats.cumulative_return
      )} · ${benchmarkLabel} ${formatPercent(stats.benchmark_cumulative_return)} · 超额 ${formatPercent(
        stats.excess_return
      )} · 回撤 ${formatPercent(stats.max_drawdown)}`;
    }
  }
  const rows = activeStockPerformanceRows.value.filter((item) => Number.isFinite(item.dailyRaw));
  if (rows.length === 0) {
    return "暂无统计";
  }
  const positiveDays = rows.filter((item) => item.dailyRaw > 0).length;
  const cumulative = rows[rows.length - 1].cumulativeRaw;
  return `样本 ${rows.length} 日 · 胜率 ${formatRate(positiveDays / rows.length)} · 累计 ${formatPercent(cumulative)}`;
});

const activeStockScoreFramework = computed(() => {
  const stockID = activeStockView.value?.id;
  if (!stockID) {
    return null;
  }
  return stockScoreFrameworkMap.value[stockID] || null;
});

const activeStockPerformanceStats = computed(() => {
  const stockID = activeStockView.value?.id;
  if (!stockID) {
    return null;
  }
  return stockPerformanceStatsMap.value[stockID] || null;
});

const activeStockRelatedNews = computed(() => {
  const stockID = activeStockView.value?.id;
  if (!stockID) {
    return [];
  }
  const items = stockRelatedNewsMap.value[stockID];
  if (!Array.isArray(items)) {
    return [];
  }
  return items;
});

const activeStockReasonSupports = computed(() => {
  const stock = activeStockBase.value;
  if (!stock?.id) {
    return [];
  }
  const detail = stockDetailMap.value[stock.id] || {};
  const rows = activeStockPerformanceRows.value.filter((item) => Number.isFinite(item.dailyRaw));
  const framework = activeStockScoreFramework.value;
  const stats = activeStockPerformanceStats.value;

  const result = [];
  const factorItems = Array.isArray(framework?.factors)
    ? framework.factors
    : [
        { label: "技术因子", weight: 0.3, score: detail?.tech_score, note: "趋势强度与动量结构" },
        { label: "基本面因子", weight: 0.3, score: detail?.fund_score, note: "盈利能力与估值质量" },
        { label: "情绪因子", weight: 0.2, score: detail?.sentiment_score, note: "舆情与事件偏好" },
        { label: "资金流因子", weight: 0.2, score: detail?.money_flow_score, note: "主力净流入趋势" }
      ];
  factorItems.forEach((item) => {
    const value = Number(item?.score);
    if (!Number.isFinite(value)) {
      return;
    }
    result.push({
      label: item?.label || "因子",
      value: `${value.toFixed(1)} 分`,
      note: `权重 ${formatRate(item?.weight)} · ${(item?.note || "").trim() || "评分因子"} · 判定 ${describeScoreLevel(value)}`
    });
  });

  const totalScore = Number(framework?.total_score ?? stock.score);
  const weightedScore = Number(framework?.weighted_score);
  const scoreGap = Number(framework?.score_gap);
  if (Number.isFinite(totalScore) || Number.isFinite(weightedScore)) {
    result.push({
      label: "后端评分体系",
      value: `${formatScore(Number.isFinite(totalScore) ? totalScore : weightedScore)} 分`,
      note: `${framework?.method || "四因子综合模型"}${
        Number.isFinite(scoreGap) ? ` · 总分偏差 ${formatSignedNumber(scoreGap, 1)} 分` : ""
      }`
    });
  }

  if (stats) {
    const benchmarkInfo = stats.benchmark_symbol ? `${stats.benchmark_symbol} · ` : "";
    result.push({
      label: "基准对比",
      value: `超额 ${formatPercent(stats.excess_return)}`,
      note: `基准累计 ${formatPercent(stats.benchmark_cumulative_return)} · 回撤 ${formatPercent(
        stats.max_drawdown
      )} · ${benchmarkInfo}${stats.benchmark_source || "基准来源未标注"}`
    });
  }

  if (rows.length > 0) {
    const cumulative = rows[rows.length - 1].cumulativeRaw;
    const positive = rows.filter((item) => item.dailyRaw > 0).length;
    const avgDaily = rows.reduce((sum, item) => sum + item.dailyRaw, 0) / rows.length;
    result.push({
      label: "历史绩效",
      value: `胜率 ${formatRate(positive / rows.length)}`,
      note: `${rows.length} 日均收益 ${formatPercent(avgDaily)}，累计 ${formatPercent(cumulative)}`
    });
  }

  if (activeStockRelatedNews.value.length > 0) {
    const latestNews = activeStockRelatedNews.value[0];
    result.push({
      label: "资讯支撑",
      value: `${activeStockRelatedNews.value.length} 条`,
      note: `最近一条：${latestNews.title || "资讯"}（${latestNews.source || "资讯中心"}）`
    });
  }

  if (result.length === 0) {
    result.push({
      label: "数据状态",
      value: "等待同步",
      note: "请点击“刷新业绩”拉取推荐详情和历史表现。"
    });
  }

  return result;
});

const futuresRows = computed(() =>
  (rawFutures.value || []).map((item) => ({
    id: item.id,
    contract: item.contract || "-",
    title: `${item.name || "-"} · ${mapDirection(item.direction)}`,
    note: `风险 ${mapRisk(item.risk_level)}，仓位 ${item.position_range || "-"}，状态 ${
      mapStatus(item.status).label
    }`
  }))
);

const activeFuturesBase = computed(() => {
  const items = rawFutures.value || [];
  if (items.length === 0) {
    return null;
  }
  if (activeFuturesID.value) {
    const matched = items.find((item) => item.id === activeFuturesID.value);
    if (matched) {
      return matched;
    }
  }
  return items[0];
});

const activeFuturesInsight = computed(() => {
  const strategyID = activeFuturesBase.value?.id;
  if (!strategyID) {
    return null;
  }
  return futuresInsightMap.value[strategyID] || null;
});

const activeFuturesScoreFramework = computed(() => activeFuturesInsight.value?.score_framework || null);

const activeFuturesPerformanceStats = computed(() => activeFuturesInsight.value?.performance_stats || null);

const activeFuturesRelatedNews = computed(() => {
  const items = activeFuturesInsight.value?.related_news;
  return Array.isArray(items) ? items : [];
});

const activeFuturesRelatedEvents = computed(() => {
  const items = activeFuturesInsight.value?.related_events;
  return Array.isArray(items) ? items : [];
});

const activeFuturesView = computed(() => {
  const base = activeFuturesBase.value;
  if (!base) {
    return null;
  }
  const insight = activeFuturesInsight.value || {};
  const strategy = insight.strategy || base;
  const guidance = insight.guidance || {};
  const reasonParts = [strategy.reason_summary || base.reason_summary, guidance.invalid_condition].filter(Boolean);
  return {
    id: strategy.id || base.id,
    name: `${strategy.contract || base.contract || "-"} ${strategy.name || base.name || ""}`.trim(),
    reason: reasonParts.join("；") || "等待策略同步",
    direction: mapDirection(strategy.direction),
    risk: mapRisk(strategy.risk_level),
    position: strategy.position_range || base.position_range || "-",
    validRange: formatDateRange(strategy.valid_from || base.valid_from, strategy.valid_to || base.valid_to),
    entryRange: guidance.entry_range || "-",
    takeProfitRange: guidance.take_profit_range || "-",
    stopLossRange: guidance.stop_loss_range || "-",
    invalidCondition: guidance.invalid_condition || "-"
  };
});

const activeFuturesPerformanceRows = computed(() => {
  const insight = activeFuturesInsight.value || {};
  const points = Array.isArray(insight.performance) ? [...insight.performance] : [];
  const benchmarkPoints = Array.isArray(insight.benchmark) ? [...insight.benchmark] : [];
  if (points.length === 0) {
    return [];
  }
  points.sort((a, b) => compareDateAsc(a?.date, b?.date));
  const benchmarkMap = new Map();
  benchmarkPoints.forEach((item) => {
    if (item?.date) {
      benchmarkMap.set(item.date, Number(item.return));
    }
  });
  let acc = 1;
  let benchmarkAcc = 1;
  return points.map((point) => {
    const daily = Number(point?.return);
    const benchmarkDaily = benchmarkMap.has(point?.date) ? Number(benchmarkMap.get(point?.date)) : NaN;
    if (Number.isFinite(daily)) {
      acc *= 1 + daily;
    }
    if (Number.isFinite(benchmarkDaily)) {
      benchmarkAcc *= 1 + benchmarkDaily;
    }
    const cumulative = Number.isFinite(daily) ? acc - 1 : null;
    const benchmarkCumulative = Number.isFinite(benchmarkDaily) ? benchmarkAcc - 1 : null;
    const excess = Number.isFinite(cumulative) && Number.isFinite(benchmarkCumulative)
      ? cumulative - benchmarkCumulative
      : null;
    return {
      date: formatDate(point?.date),
      dailyReturn: formatPercent(daily),
      cumulativeReturn: formatPercent(cumulative),
      benchmarkReturn: formatPercent(benchmarkCumulative),
      excessReturn: formatPercent(excess),
      dailyClass: trendClassByNumber(daily),
      cumulativeClass: trendClassByNumber(cumulative),
      benchmarkClass: trendClassByNumber(benchmarkCumulative),
      excessClass: trendClassByNumber(excess),
      dailyRaw: daily
    };
  });
});

const activeFuturesPerformanceSummary = computed(() => {
  const stats = activeFuturesPerformanceStats.value;
  if (stats) {
    const benchmarkLabel = stats.benchmark_symbol ? `基准(${stats.benchmark_symbol})` : "基准";
    return `样本 ${Number(stats.sample_days || 0)} 日 · 胜率 ${formatRate(stats.win_rate)} · 累计 ${formatPercent(
      stats.cumulative_return
    )} · ${benchmarkLabel} ${formatPercent(stats.benchmark_cumulative_return)} · 超额 ${formatPercent(
      stats.excess_return
    )}`;
  }
  const rows = activeFuturesPerformanceRows.value.filter((item) => Number.isFinite(item.dailyRaw));
  if (rows.length === 0) {
    return "暂无统计";
  }
  const wins = rows.filter((item) => item.dailyRaw > 0).length;
  return `样本 ${rows.length} 日 · 胜率 ${formatRate(wins / rows.length)}`;
});

const activeFuturesReasonSupports = computed(() => {
  const framework = activeFuturesScoreFramework.value;
  const stats = activeFuturesPerformanceStats.value;
  const result = [];
  const factors = Array.isArray(framework?.factors) ? framework.factors : [];
  factors.forEach((factor) => {
    const score = Number(factor?.score);
    if (!Number.isFinite(score)) {
      return;
    }
    result.push({
      label: factor.label || "因子",
      value: `${score.toFixed(1)} 分`,
      note: `权重 ${formatRate(factor.weight)} · 贡献 ${formatSignedNumber(factor.contribution, 2)}`
    });
  });
  if (framework) {
    result.push({
      label: "后端评分体系",
      value: `${formatScore(framework.total_score)} 分`,
      note: `${framework.method || "futures strategy model"} · 总分偏差 ${formatSignedNumber(framework.score_gap, 2)}`
    });
  }
  if (stats) {
    result.push({
      label: "基准对比",
      value: `超额 ${formatPercent(stats.excess_return)}`,
      note: `${stats.benchmark_symbol || "-"} · ${stats.benchmark_source || "基准来源未标注"}`
    });
  }
  if (result.length === 0) {
    result.push({
      label: "数据状态",
      value: "等待同步",
      note: "请点击刷新详情"
    });
  }
  return result;
});

const eventRows = computed(() =>
  (rawEvents.value || []).map((item) => {
    const detail = eventDetailMap.value[item.id] || {};
    const eventType = String(item.event_type || detail.event_type || "").toUpperCase();
    return {
      id: item.id,
      symbol: item.symbol || detail.symbol || "-",
      title: summarizeEventTitle(item.summary || detail.summary || "-"),
      summary: detail.summary || item.summary || "-",
      triggerRule: detail.trigger_rule || item.trigger_rule || "-",
      source: detail.source || item.source || "-",
      time: formatDateTime(detail.created_at || item.created_at),
      type: eventType,
      typeLabel: mapEventType(eventType),
      levelClass: mapEventLevelClass(eventType),
      levelText: mapEventLevelText(eventType)
    };
  })
);

const eventTypeOptions = computed(() => {
  const all = [{ value: "ALL", label: "全部" }];
  const uniqueTypes = new Set(eventRows.value.map((item) => item.type).filter(Boolean));
  uniqueTypes.forEach((value) => {
    all.push({
      value,
      label: mapEventType(value)
    });
  });
  return all;
});

const filteredEventRows = computed(() => {
  if (activeEventType.value === "ALL") {
    return eventRows.value;
  }
  return eventRows.value.filter((item) => item.type === activeEventType.value);
});

const activeEventView = computed(() => {
  const current =
    filteredEventRows.value.find((item) => item.id === activeEventID.value) || filteredEventRows.value[0];
  return current || null;
});

const tags = computed(() => [
  `股票策略 ${stockRows.value.length} 条`,
  `期货策略 ${futuresRows.value.length} 条`,
  `市场事件 ${eventRows.value.length} 条`,
  `高优先策略 ${stockRows.value.filter((item) => item.statusClass === "good").length} 条`
]);

async function loadStrategies() {
  loading.value = true;
  errorMessage.value = "";
  const errors = [];
  try {
    const [stockRes, futuresRes, eventsRes] = await Promise.allSettled([
      listStockRecommendations({ page: 1, page_size: 20 }),
      listFuturesStrategies({ page: 1, page_size: 20 }),
      listMarketEvents({ page: 1, page_size: 30 })
    ]);

    if (stockRes.status === "fulfilled" && Array.isArray(stockRes.value?.items)) {
      rawStocks.value = stockRes.value.items;
    } else if (stockRes.status === "rejected") {
      errors.push(`股票策略接口失败：${stockRes.reason?.message || "unknown error"}`);
    }

    if (futuresRes.status === "fulfilled" && Array.isArray(futuresRes.value?.items)) {
      rawFutures.value = futuresRes.value.items;
    } else if (futuresRes.status === "rejected") {
      errors.push(`期货策略接口失败：${futuresRes.reason?.message || "unknown error"}`);
    }

    if (eventsRes.status === "fulfilled" && Array.isArray(eventsRes.value?.items)) {
      rawEvents.value = eventsRes.value.items;
      const nextDetailMap = { ...eventDetailMap.value };
      eventsRes.value.items.forEach((item) => {
        if (item?.id) {
          nextDetailMap[item.id] = {
            ...nextDetailMap[item.id],
            ...item
          };
        }
      });
      eventDetailMap.value = nextDetailMap;
    } else if (eventsRes.status === "rejected") {
      errors.push(`市场事件接口失败：${eventsRes.reason?.message || "unknown error"}`);
    }

    if (errors.length > 0) {
      errorMessage.value = errors.join("；");
    }

    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  } catch (error) {
    errorMessage.value = error?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

async function loadFuturesInsight(id, options = {}) {
  const { silent = false, force = false } = options;
  if (!id || id.startsWith("fs_local_")) {
    if (!silent) {
      futuresInsightErrorMessage.value = "";
    }
    return;
  }
  if (!force && loadedFuturesInsightIDs.has(id) && futuresInsightMap.value[id]) {
    return;
  }
  if (!silent) {
    futuresInsightLoading.value = true;
    futuresInsightErrorMessage.value = "";
  }
  try {
    const insight = await getFuturesStrategyInsight(id);
    if (insight) {
      futuresInsightMap.value = {
        ...futuresInsightMap.value,
        [id]: insight
      };
      loadedFuturesInsightIDs.add(id);
    }
  } catch (error) {
    if (!silent) {
      futuresInsightErrorMessage.value = parseErrorMessage(error);
    }
  } finally {
    if (!silent) {
      futuresInsightLoading.value = false;
      await nextTick();
      repositionFuturesDetailDialog();
    }
  }
}

async function openFuturesInsight(id, event) {
  if (!id) {
    return;
  }
  const anchor = event?.currentTarget?.getBoundingClientRect?.();
  if (anchor) {
    futuresDetailAnchorRect.value = {
      top: anchor.top,
      left: anchor.left,
      right: anchor.right,
      bottom: anchor.bottom
    };
  } else {
    futuresDetailAnchorRect.value = null;
  }
  stockDetailDialogVisible.value = false;
  stockDetailAnchorRect.value = null;
  futuresDetailDialogVisible.value = true;
  activeFuturesID.value = id;
  await nextTick();
  repositionFuturesDetailDialog();
  loadFuturesInsight(id);
}

function refreshActiveFuturesInsight() {
  if (activeFuturesView.value?.id) {
    loadFuturesInsight(activeFuturesView.value.id, { force: true });
  }
}

function closeFuturesDetail() {
  futuresDetailDialogVisible.value = false;
  futuresDetailAnchorRect.value = null;
}

async function loadActiveStockDetail(id, options = {}) {
  const { silent = false, force = false } = options;
  if (!id) {
    return;
  }
  if (id.startsWith("sr_local_")) {
    if (!silent) {
      stockDetailErrorMessage.value = "";
    }
    return;
  }

  const hasDetail = !!stockDetailMap.value[id];
  const hasPerformance = Array.isArray(stockPerformanceMap.value[id]);
  const hasRelatedNews = Array.isArray(stockRelatedNewsMap.value[id]);
  const hasFramework = !!stockScoreFrameworkMap.value[id];
  const hasStats = !!stockPerformanceStatsMap.value[id];
  const shouldLoadInsight = force || !loadedStockInsightIDs.has(id) || !hasFramework || !hasStats;
  const shouldLoadDetail = force || !loadedStockDetailIDs.has(id) || !hasDetail;
  const shouldLoadPerformance = force || !loadedStockPerformanceIDs.has(id) || !hasPerformance;
  const shouldLoadRelatedNews = force || !loadedStockRelatedNewsIDs.has(id) || !hasRelatedNews;
  if (!shouldLoadInsight && !shouldLoadDetail && !shouldLoadPerformance && !shouldLoadRelatedNews) {
    return;
  }

  const nextDetailMap = { ...stockDetailMap.value };
  const nextPerformanceMap = { ...stockPerformanceMap.value };
  const nextBenchmarkMap = { ...stockBenchmarkMap.value };
  const nextRelatedNewsMap = { ...stockRelatedNewsMap.value };
  const nextScoreFrameworkMap = { ...stockScoreFrameworkMap.value };
  const nextPerformanceStatsMap = { ...stockPerformanceStatsMap.value };
  const errors = [];

  if (!silent) {
    stockDetailLoading.value = true;
    stockRelatedNewsLoading.value = true;
    stockDetailErrorMessage.value = "";
  }

  let insightLoaded = false;
  if (shouldLoadInsight) {
    try {
      const insight = await getStockRecommendationInsight(id);
      if (insight?.detail) {
        nextDetailMap[id] = insight.detail;
        loadedStockDetailIDs.add(id);
      }
      if (Array.isArray(insight?.performance)) {
        nextPerformanceMap[id] = insight.performance;
        loadedStockPerformanceIDs.add(id);
      }
      if (Array.isArray(insight?.benchmark)) {
        nextBenchmarkMap[id] = insight.benchmark;
      }
      if (insight?.score_framework) {
        nextScoreFrameworkMap[id] = insight.score_framework;
      }
      if (insight?.performance_stats) {
        nextPerformanceStatsMap[id] = insight.performance_stats;
      }
      if (Array.isArray(insight?.related_news)) {
        nextRelatedNewsMap[id] = insight.related_news.map((item) => mapStockRelatedNewsItem(item)).filter(Boolean);
        loadedStockRelatedNewsIDs.add(id);
      }
      loadedStockInsightIDs.add(id);
      insightLoaded = true;
    } catch (error) {
      errors.push(`推荐解释包同步失败：${parseErrorMessage(error)}`);
    }
  }

  const tasks = [];
  if (!insightLoaded) {
    if (shouldLoadDetail) {
      tasks.push(
        getStockRecommendationDetail(id)
          .then((data) => {
            if (data) {
              nextDetailMap[id] = data;
              loadedStockDetailIDs.add(id);
            }
          })
          .catch((error) => {
            errors.push(`推荐详情同步失败：${parseErrorMessage(error)}`);
          })
      );
    }

    if (shouldLoadPerformance) {
      tasks.push(
        getStockRecommendationPerformance(id)
          .then((data) => {
            if (Array.isArray(data?.points)) {
              nextPerformanceMap[id] = data.points;
              loadedStockPerformanceIDs.add(id);
            }
          })
          .catch((error) => {
            errors.push(`历史业绩同步失败：${parseErrorMessage(error)}`);
          })
      );
    }

    if (shouldLoadRelatedNews) {
      tasks.push(
        loadStockRelatedNews(id, { force, silent: true })
          .then((items) => {
            nextRelatedNewsMap[id] = items;
            loadedStockRelatedNewsIDs.add(id);
          })
          .catch((error) => {
            errors.push(`相关资讯同步失败：${parseErrorMessage(error)}`);
          })
      );
    }
  }

  await Promise.all(tasks);
  stockDetailMap.value = nextDetailMap;
  stockPerformanceMap.value = nextPerformanceMap;
  stockBenchmarkMap.value = nextBenchmarkMap;
  stockRelatedNewsMap.value = nextRelatedNewsMap;
  stockScoreFrameworkMap.value = nextScoreFrameworkMap;
  stockPerformanceStatsMap.value = nextPerformanceStatsMap;

  if (!silent) {
    stockDetailErrorMessage.value = errors.join("；");
    stockDetailLoading.value = false;
    stockRelatedNewsLoading.value = false;
    await nextTick();
    repositionStockDetailDialog();
  }
}

function refreshActiveStockDetail() {
  if (activeStockView.value?.id) {
    loadActiveStockDetail(activeStockView.value.id, { force: true });
  }
}

async function openStockDetail(id, event) {
  if (!id) {
    return;
  }
  closeFuturesDetail();
  const anchor = event?.currentTarget?.getBoundingClientRect?.();
  if (anchor) {
    stockDetailAnchorRect.value = {
      top: anchor.top,
      left: anchor.left,
      right: anchor.right,
      bottom: anchor.bottom
    };
  } else {
    stockDetailAnchorRect.value = null;
  }
  stockDetailDialogVisible.value = true;
  activeStockID.value = id;
  await nextTick();
  repositionStockDetailDialog();
  loadActiveStockDetail(id, { silent: false });
}

function closeStockDetail() {
  stockDetailDialogVisible.value = false;
  stockDetailAnchorRect.value = null;
}

async function loadStockRelatedNews(stockID, options = {}) {
  const { force = false, silent = false } = options;
  if (!stockID) {
    return [];
  }
  if (stockID.startsWith("sr_local_")) {
    return Array.isArray(stockRelatedNewsMap.value[stockID]) ? stockRelatedNewsMap.value[stockID] : [];
  }

  const base = (rawStocks.value || []).find((item) => item.id === stockID);
  if (!base) {
    return [];
  }

  if (!force && loadedStockRelatedNewsIDs.has(stockID) && Array.isArray(stockRelatedNewsMap.value[stockID])) {
    return stockRelatedNewsMap.value[stockID];
  }

  if (!silent) {
    stockRelatedNewsLoading.value = true;
  }

  const symbols = deriveStockKeywords(base.symbol);
  const keywordCandidates = [base.name, symbols[0], symbols[1]].filter(Boolean);
  const dedupKeywords = Array.from(new Set(keywordCandidates));

  const collected = [];
  const seen = new Set();

  for (const keyword of dedupKeywords) {
    if (!keyword || collected.length >= 6) {
      continue;
    }
    try {
      const response = await listNewsArticles({
        page: 1,
        page_size: 6,
        keyword
      });
      const items = Array.isArray(response?.items) ? response.items : [];
      for (const item of items) {
        const mapped = mapStockRelatedNewsItem(item);
        if (!mapped || seen.has(mapped.id)) {
          continue;
        }
        seen.add(mapped.id);
        collected.push(mapped);
        if (collected.length >= 6) {
          break;
        }
      }
    } catch {
      // 忽略单个关键词的拉取失败，继续尝试其他关键词
    }
  }

  collected.sort((a, b) => compareDateAsc(b.rawTime, a.rawTime));
  const sliced = collected.slice(0, 5);
  if (!silent) {
    stockRelatedNewsLoading.value = false;
  }
  return sliced;
}

function repositionStockDetailDialog() {
  if (!stockDetailDialogVisible.value) {
    return;
  }
  const dialogEl = stockDetailDialogRef.value;
  if (!dialogEl) {
    return;
  }

  const viewportWidth = window.innerWidth || 1280;
  const viewportHeight = window.innerHeight || 800;
  const horizontalMargin = 12;
  const verticalMargin = 12;

  const width = Math.min(980, viewportWidth - horizontalMargin * 2);
  const height = Math.min(dialogEl.offsetHeight || 620, viewportHeight - verticalMargin * 2);

  const anchor = stockDetailAnchorRect.value;
  if (!anchor) {
    stockDetailDialogStyle.value = {
      top: `${Math.max(verticalMargin, (viewportHeight - height) / 2)}px`,
      left: `${Math.max(horizontalMargin, (viewportWidth - width) / 2)}px`,
      width: `${Math.round(width)}px`
    };
    return;
  }

  let left = anchor.left;
  if (left + width > viewportWidth - horizontalMargin) {
    left = viewportWidth - width - horizontalMargin;
  }
  if (left < horizontalMargin) {
    left = horizontalMargin;
  }

  let top = anchor.bottom + 8;
  if (top + height > viewportHeight - verticalMargin) {
    top = anchor.top - height - 8;
  }
  if (top < verticalMargin) {
    top = verticalMargin;
  }

  stockDetailDialogStyle.value = {
    top: `${Math.round(top)}px`,
    left: `${Math.round(left)}px`,
    width: `${Math.round(width)}px`
  };
}

function repositionFuturesDetailDialog() {
  if (!futuresDetailDialogVisible.value) {
    return;
  }
  const dialogEl = futuresDetailDialogRef.value;
  if (!dialogEl) {
    return;
  }

  const viewportWidth = window.innerWidth || 1280;
  const viewportHeight = window.innerHeight || 800;
  const horizontalMargin = 12;
  const verticalMargin = 12;

  const width = Math.min(860, viewportWidth - horizontalMargin * 2);
  const height = Math.min(dialogEl.offsetHeight || 640, viewportHeight - verticalMargin * 2);

  const anchor = futuresDetailAnchorRect.value;
  if (!anchor) {
    futuresDetailDialogStyle.value = {
      top: `${Math.max(verticalMargin, (viewportHeight - height) / 2)}px`,
      left: `${Math.max(horizontalMargin, (viewportWidth - width) / 2)}px`,
      width: `${Math.round(width)}px`
    };
    return;
  }

  const anchorCenterX = (anchor.left + anchor.right) / 2;
  let left = anchorCenterX - width / 2;
  if (anchorCenterX > viewportWidth * 0.6) {
    left = anchor.right - width;
  } else if (anchorCenterX < viewportWidth * 0.4) {
    left = anchor.left;
  }
  if (left + width > viewportWidth - horizontalMargin) {
    left = viewportWidth - width - horizontalMargin;
  }
  if (left < horizontalMargin) {
    left = horizontalMargin;
  }

  let top = anchor.bottom + 8;
  if (top + height > viewportHeight - verticalMargin) {
    top = anchor.top - height - 8;
  }
  if (top < verticalMargin) {
    top = verticalMargin;
  }

  futuresDetailDialogStyle.value = {
    top: `${Math.round(top)}px`,
    left: `${Math.round(left)}px`,
    width: `${Math.round(width)}px`
  };
}

async function loadMarketEventDetail(id, options = {}) {
  const { silent = false } = options;
  if (!id) {
    return;
  }
  if (!silent) {
    eventDetailLoading.value = true;
  }
  try {
    const data = await getMarketEventDetail(id);
    if (data?.id) {
      loadedEventDetailIDs.add(id);
      eventDetailMap.value = {
        ...eventDetailMap.value,
        [id]: data
      };
    }
  } catch {
    // 详情接口失败时保留当前事件摘要
  } finally {
    if (!silent) {
      eventDetailLoading.value = false;
    }
  }
}

function refreshActiveEventDetail() {
  if (activeEventView.value?.id) {
    loadMarketEventDetail(activeEventView.value.id);
  }
}

function summarizeEventTitle(summary) {
  const text = String(summary || "").trim();
  if (!text) {
    return "未命名事件";
  }
  if (text.length <= 28) {
    return text;
  }
  return `${text.slice(0, 28)}...`;
}

function buildStockScoreNote(detail) {
  const scoreEntries = [
    ["技术", detail?.tech_score],
    ["基本面", detail?.fund_score],
    ["情绪", detail?.sentiment_score],
    ["资金流", detail?.money_flow_score]
  ]
    .map(([label, value]) => {
      const num = Number(value);
      if (!Number.isFinite(num)) {
        return "";
      }
      return `${label}${num.toFixed(1)}`;
    })
    .filter(Boolean);
  if (scoreEntries.length === 0) {
    return "";
  }
  return `子分项：${scoreEntries.join(" / ")}`;
}

function mapStockRelatedNewsItem(item) {
  if (!item?.id) {
    return null;
  }
  const visibilityText = String(item.visibility || "").toUpperCase() === "VIP" ? "VIP" : "公开";
  const rawTime = item.published_at || item.time || item.created_at || item.updated_at || "";
  const relevanceScore = Number(item.relevance_score ?? item.relevanceScore);
  return {
    id: item.id,
    title: item.title || "未命名资讯",
    source: item.source || "资讯中心",
    time: formatDateTime(rawTime),
    rawTime,
    visibility: visibilityText,
    summary: summarizeNewsText(item.summary || item.content),
    relevanceScore: Number.isFinite(relevanceScore) ? relevanceScore : null
  };
}

function summarizeNewsText(value) {
  const text = String(value || "")
    .replace(/<[^>]*>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
  if (!text) {
    return "暂无摘要";
  }
  if (text.length <= 80) {
    return text;
  }
  return `${text.slice(0, 80)}...`;
}

function deriveStockKeywords(symbol) {
  const raw = String(symbol || "").trim().toUpperCase();
  if (!raw) {
    return [];
  }
  const compact = raw.replace(/\.(SH|SZ|BJ)$/i, "");
  if (compact === raw) {
    return [raw];
  }
  return [raw, compact];
}

function compareDateAsc(a, b) {
  const left = Date.parse(a || "");
  const right = Date.parse(b || "");
  if (Number.isNaN(left) && Number.isNaN(right)) return 0;
  if (Number.isNaN(left)) return 1;
  if (Number.isNaN(right)) return -1;
  return left - right;
}

function formatScore(value) {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  return num.toFixed(1);
}

function mapRisk(level) {
  const source = String(level || "").toUpperCase();
  if (source === "LOW") return "低风险";
  if (source === "MEDIUM") return "中风险";
  if (source === "HIGH") return "高风险";
  return source || "-";
}

function mapDirection(direction) {
  const source = String(direction || "").toUpperCase();
  if (source === "LONG") return "做多";
  if (source === "SHORT") return "做空";
  if (source === "NEUTRAL") return "中性";
  return source || "-";
}

function mapStatus(status) {
  const source = String(status || "").toUpperCase();
  if (source === "PUBLISHED" || source === "ACTIVE") {
    return { label: "推荐", className: "good" };
  }
  if (source === "WATCH" || source === "DRAFT") {
    return { label: "观察", className: "watch" };
  }
  if (source === "DISABLED" || source === "EXPIRED") {
    return { label: "谨慎", className: "risky" };
  }
  return { label: source || "-", className: "normal" };
}

function mapEventType(value) {
  const source = String(value || "").toUpperCase();
  if (source === "PRICE") return "价格异动";
  if (source === "VOLUME") return "量能异动";
  if (source === "VOLATILITY") return "波动率异动";
  if (source === "POLICY") return "政策事件";
  if (source === "FLOW") return "资金流事件";
  if (source === "OTHER") return "其他事件";
  return source || "-";
}

function mapEventLevelClass(eventType) {
  const source = String(eventType || "").toUpperCase();
  if (source === "POLICY" || source === "VOLATILITY") return "high";
  if (source === "PRICE" || source === "FLOW") return "mid";
  return "low";
}

function mapEventLevelText(eventType) {
  const source = mapEventLevelClass(eventType);
  if (source === "high") return "高影响";
  if (source === "mid") return "中影响";
  return "低影响";
}

function mapVisibilityLabel(value) {
  const source = String(value || "").toUpperCase();
  if (source === "VIP") {
    return "VIP";
  }
  if (source === "PUBLIC") {
    return "公开";
  }
  return source || "-";
}

function describeScoreLevel(score) {
  const value = Number(score);
  if (!Number.isFinite(value)) {
    return "缺少评分";
  }
  if (value >= 85) {
    return "强势";
  }
  if (value >= 70) {
    return "中性偏强";
  }
  if (value >= 55) {
    return "中性";
  }
  return "偏弱";
}

function formatSignedNumber(value, digits = 2) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  const fixed = num.toFixed(digits);
  if (num > 0) {
    return `+${fixed}`;
  }
  return fixed;
}

function formatScore01(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  return num.toFixed(2);
}

function trendClassByNumber(value) {
  if (!Number.isFinite(value)) {
    return "";
  }
  if (value > 0) {
    return "up";
  }
  if (value < 0) {
    return "down";
  }
  return "";
}

function formatRate(value) {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  return `${(num * 100).toFixed(1)}%`;
}

function formatPercent(value) {
  if (value === null || value === undefined || value === "") {
    return "-";
  }
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  const percent = (num * 100).toFixed(2);
  if (num > 0) {
    return `+${percent}%`;
  }
  return `${percent}%`;
}

function formatDate(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return value || "-";
  }
  return new Date(ts).toLocaleDateString("zh-CN");
}

function formatDateRange(start, end) {
  const startText = formatDate(start);
  const endText = formatDate(end);
  if (startText === "-" && endText === "-") {
    return "-";
  }
  if (endText === "-") {
    return `${startText} 起`;
  }
  if (startText === "-") {
    return `截至 ${endText}`;
  }
  return `${startText} ~ ${endText}`;
}

function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) return "-";
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

function parseErrorMessage(error) {
  if (!error) {
    return "unknown error";
  }
  const responseMessage =
    error?.response?.data?.message || error?.response?.data?.error || error?.response?.statusText;
  return responseMessage || error?.message || "unknown error";
}

function handleStockDialogViewportChange() {
  if (!stockDetailDialogVisible.value) {
    return;
  }
  repositionStockDetailDialog();
}

function handleFuturesDialogViewportChange() {
  if (!futuresDetailDialogVisible.value) {
    return;
  }
  repositionFuturesDetailDialog();
}

watch(
  () => stockDetailDialogVisible.value,
  (visible) => {
    if (visible) {
      window.addEventListener("resize", handleStockDialogViewportChange);
      window.addEventListener("scroll", handleStockDialogViewportChange, true);
      nextTick(() => {
        repositionStockDetailDialog();
      });
      return;
    }
    window.removeEventListener("resize", handleStockDialogViewportChange);
    window.removeEventListener("scroll", handleStockDialogViewportChange, true);
  }
);

watch(
  () => futuresDetailDialogVisible.value,
  (visible) => {
    if (visible) {
      window.addEventListener("resize", handleFuturesDialogViewportChange);
      window.addEventListener("scroll", handleFuturesDialogViewportChange, true);
      nextTick(() => {
        repositionFuturesDetailDialog();
      });
      return;
    }
    window.removeEventListener("resize", handleFuturesDialogViewportChange);
    window.removeEventListener("scroll", handleFuturesDialogViewportChange, true);
  }
);

watch(
  stockRows,
  (rows) => {
    if (!rows.length) {
      activeStockID.value = "";
      return;
    }
    if (!rows.some((item) => item.id === activeStockID.value)) {
      activeStockID.value = rows[0].id;
    }
  },
  { immediate: true }
);

watch(
  futuresRows,
  (rows) => {
    if (!rows.length) {
      activeFuturesID.value = "";
      return;
    }
    if (!rows.some((item) => item.id === activeFuturesID.value)) {
      activeFuturesID.value = rows[0].id;
    }
  },
  { immediate: true }
);

watch(
  () => activeFuturesID.value,
  (id) => {
    if (!id) {
      return;
    }
    loadFuturesInsight(id, { silent: true });
  },
  { immediate: true }
);

watch(
  () => activeStockID.value,
  (id) => {
    if (!id) {
      return;
    }
    if (stockDetailDialogVisible.value) {
      return;
    }
    loadActiveStockDetail(id, { silent: true });
  },
  { immediate: true }
);

watch(
  () => [activeStockReasonSupports.value.length, activeStockPerformanceRows.value.length, activeStockRelatedNews.value.length],
  () => {
    if (!stockDetailDialogVisible.value) {
      return;
    }
    nextTick(() => {
      repositionStockDetailDialog();
    });
  }
);

watch(
  () => [
    activeFuturesReasonSupports.value.length,
    activeFuturesPerformanceRows.value.length,
    activeFuturesRelatedNews.value.length,
    activeFuturesRelatedEvents.value.length
  ],
  () => {
    if (!futuresDetailDialogVisible.value) {
      return;
    }
    nextTick(() => {
      repositionFuturesDetailDialog();
    });
  }
);

watch(
  filteredEventRows,
  (rows) => {
    if (!rows.length) {
      activeEventID.value = "";
      return;
    }
    if (!rows.some((item) => item.id === activeEventID.value)) {
      activeEventID.value = rows[0].id;
    }
  },
  { immediate: true }
);

watch(
  () => activeEventID.value,
  (id) => {
    if (!id || id.startsWith("me_local_")) {
      return;
    }
    if (loadedEventDetailIDs.has(id)) {
      return;
    }
    loadMarketEventDetail(id, { silent: true });
  },
  { immediate: true }
);

onMounted(() => {
  loadStrategies();
});

onBeforeUnmount(() => {
  window.removeEventListener("resize", handleStockDialogViewportChange);
  window.removeEventListener("scroll", handleStockDialogViewportChange, true);
  window.removeEventListener("resize", handleFuturesDialogViewportChange);
  window.removeEventListener("scroll", handleFuturesDialogViewportChange, true);
});
</script>

<style scoped>
.strategy-page {
  display: grid;
  gap: 12px;
}

.strategy-hero {
  border-radius: 20px;
  padding: 16px;
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: end;
  gap: 12px;
  background:
    radial-gradient(circle at 100% 0%, rgba(63, 127, 113, 0.16) 0%, transparent 34%),
    radial-gradient(circle at 0% 100%, rgba(234, 215, 180, 0.22) 0%, transparent 36%),
    rgba(255, 255, 255, 0.93);
}

.hero-kicker {
  margin: 0;
  color: var(--color-pine-600);
  font-size: 12px;
}

.tag-row {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.tag-row span {
  border-radius: 999px;
  padding: 4px 10px;
  font-size: 12px;
  border: 1px solid rgba(216, 223, 216, 0.92);
  color: var(--color-text-sub);
  background: rgba(252, 251, 247, 0.88);
}

.hero-actions {
  display: flex;
  gap: 8px;
}

.primary-btn {
  border: 0;
  border-radius: 10px;
  padding: 9px 13px;
  cursor: pointer;
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.primary-btn:disabled {
  cursor: not-allowed;
  opacity: 0.72;
}

.api-state {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--color-pine-700);
}

.api-state.warning {
  color: #7f5f36;
}

.strategy-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: 1fr 0.88fr;
}

.matrix-card,
.cycle-card {
  padding: 14px;
}

.matrix-head {
  margin-top: 12px;
  border-radius: 10px;
  padding: 8px 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(246, 244, 239, 0.8);
  display: grid;
  grid-template-columns: 1.4fr repeat(4, 0.6fr);
  gap: 10px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.matrix-list {
  margin-top: 8px;
  display: grid;
  gap: 8px;
}

.matrix-item {
  width: 100%;
  text-align: left;
  font: inherit;
  cursor: pointer;
  border-radius: 12px;
  border: 1px solid rgba(216, 223, 216, 0.92);
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
  display: grid;
  grid-template-columns: 1.4fr repeat(4, 0.6fr);
  align-items: center;
  gap: 10px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.matrix-item:hover {
  border-color: rgba(63, 127, 113, 0.28);
}

.matrix-item.active {
  border-color: rgba(63, 127, 113, 0.46);
  box-shadow: 0 8px 20px rgba(63, 127, 113, 0.13);
}

.name {
  margin: 0;
  font-weight: 600;
}

.desc {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.matrix-item p {
  margin: 0;
  font-size: 13px;
}

.score {
  font-weight: 700;
  color: var(--color-pine-700);
}

.status {
  justify-self: start;
  font-size: 12px;
  border-radius: 999px;
  padding: 3px 9px;
}

.status.good {
  color: var(--color-pine-700);
  background: rgba(221, 236, 229, 0.72);
}

.status.normal {
  color: #5a5e2f;
  background: rgba(233, 231, 190, 0.72);
}

.status.watch {
  color: #775325;
  background: rgba(234, 215, 180, 0.72);
}

.status.risky {
  color: #8a3c2f;
  background: rgba(230, 194, 185, 0.62);
}

.matrix-tip {
  margin: 9px 0 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.stock-detail-dialog-mask,
.futures-detail-dialog-mask {
  position: fixed;
  inset: 0;
  background: rgba(15, 28, 26, 0.38);
  display: block;
  z-index: 30;
}

.stock-detail-dialog,
.futures-detail-dialog {
  position: fixed;
  width: min(980px, calc(100vw - 24px));
  max-height: min(90vh, 840px);
  overflow: auto;
  border-radius: 14px;
  border: 1px solid rgba(216, 223, 216, 0.92);
  background: rgba(246, 244, 239, 0.97);
  padding: 14px;
  display: grid;
  gap: 10px;
}

.futures-detail-dialog {
  width: min(860px, calc(100vw - 24px));
  max-height: min(88vh, 820px);
}

.stock-detail-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.stock-detail-head h3 {
  margin: 4px 0 0;
  font-size: 18px;
}

.stock-detail-actions {
  display: flex;
  gap: 8px;
}

.stock-detail-head button {
  border: 1px solid rgba(216, 223, 216, 0.9);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.95);
  color: var(--color-pine-700);
  padding: 8px 11px;
  cursor: pointer;
}

.stock-detail-head button:disabled {
  cursor: not-allowed;
  opacity: 0.68;
}

.close-btn {
  color: var(--color-text-sub);
}

.stock-reason {
  margin: 0;
  line-height: 1.75;
  font-size: 13px;
  color: var(--color-text-sub);
}

.stock-reason-meta {
  margin: -2px 0 2px;
  font-size: 12px;
  color: var(--color-pine-600);
}

.stock-detail-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.stock-detail-grid article {
  border-radius: 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.92);
  padding: 8px 10px;
}

.stock-detail-grid p {
  margin: 0 0 4px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.stock-detail-grid strong {
  font-size: 13px;
  color: var(--color-text-main);
}

.reason-support-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.reason-support-grid article {
  border-radius: 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.92);
  padding: 8px 10px;
  display: grid;
  gap: 4px;
}

.reason-support-grid p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.reason-support-grid strong {
  font-size: 14px;
  color: var(--color-text-main);
}

.reason-support-grid span {
  font-size: 12px;
  color: var(--color-text-sub);
}

.stock-news-box {
  border-radius: 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
  display: grid;
  gap: 8px;
}

.stock-news-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.stock-news-head p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.stock-news-head span {
  font-size: 12px;
  color: var(--color-text-sub);
}

.stock-news-list {
  display: grid;
  gap: 8px;
}

.stock-news-item {
  border-radius: 9px;
  border: 1px solid rgba(216, 223, 216, 0.8);
  background: rgba(252, 251, 247, 0.9);
  padding: 8px 10px;
}

.stock-news-item h4 {
  margin: 0;
  font-size: 13px;
  line-height: 1.45;
}

.stock-news-meta {
  margin: 5px 0 0;
  font-size: 11px;
  color: var(--color-text-sub);
}

.stock-news-summary {
  margin: 5px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
  line-height: 1.6;
}

.stock-performance-box {
  border-radius: 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
}

.stock-performance-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.stock-performance-head p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.stock-performance-head span {
  font-size: 12px;
  color: var(--color-text-sub);
}

.performance-table-wrap {
  margin-top: 8px;
  overflow-x: auto;
}

.performance-table {
  width: 100%;
  min-width: 560px;
  border-collapse: collapse;
  font-size: 12px;
}

.performance-table th,
.performance-table td {
  border-bottom: 1px solid rgba(216, 223, 216, 0.7);
  padding: 7px 5px;
  text-align: left;
}

.performance-table th {
  color: var(--color-text-sub);
  font-weight: 500;
}

.performance-table .up {
  color: #1f6c4d;
}

.performance-table .down {
  color: #8a3c2f;
}

.empty-inline {
  margin-top: 8px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.detail-warning {
  margin: 0;
  font-size: 12px;
  color: #7f5f36;
}

.detail-fade-enter-active,
.detail-fade-leave-active {
  transition: opacity 0.2s ease;
}

.detail-fade-enter-from,
.detail-fade-leave-to {
  opacity: 0;
}

.cycle-list {
  margin-top: 12px;
  display: grid;
  gap: 8px;
}

.cycle-item {
  width: 100%;
  text-align: left;
  font: inherit;
  cursor: pointer;
  border-radius: 12px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.cycle-item:hover {
  border-color: rgba(63, 127, 113, 0.28);
}

.cycle-item.active {
  border-color: rgba(63, 127, 113, 0.45);
  box-shadow: 0 8px 20px rgba(63, 127, 113, 0.12);
}

.step {
  margin: 0;
  font-size: 11px;
  color: var(--color-pine-600);
}

.title {
  margin: 5px 0 4px;
  font-weight: 600;
}

.note {
  margin: 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.6;
}

.futures-detail-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.futures-detail-grid article {
  border-radius: 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.92);
  padding: 8px 10px;
}

.futures-detail-grid p {
  margin: 0 0 4px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.futures-detail-grid strong {
  font-size: 13px;
  color: var(--color-text-main);
}

.futures-event-box {
  border-radius: 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
  display: grid;
  gap: 8px;
}

.futures-event-list {
  display: grid;
  gap: 8px;
}

.futures-event-item {
  border-radius: 8px;
  border: 1px solid rgba(216, 223, 216, 0.82);
  background: rgba(252, 251, 247, 0.9);
  padding: 8px 10px;
  display: grid;
  gap: 4px;
}

.futures-event-item p {
  margin: 0;
  font-size: 11px;
  color: var(--color-pine-600);
}

.futures-event-item strong {
  font-size: 13px;
  line-height: 1.5;
}

.futures-event-item span {
  font-size: 11px;
  color: var(--color-text-sub);
}

.empty-box {
  margin-top: 10px;
  border-radius: 11px;
  border: 1px dashed rgba(216, 223, 216, 0.95);
  background: rgba(246, 244, 239, 0.7);
  padding: 10px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.event-section {
  display: grid;
  gap: 12px;
  grid-template-columns: 0.95fr 1.05fr;
}

.event-list-card,
.event-detail-card {
  padding: 14px;
}

.event-type-tabs {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.event-type-tabs button {
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(246, 244, 239, 0.74);
  color: var(--color-text-sub);
  border-radius: 999px;
  padding: 5px 11px;
  font-size: 12px;
  cursor: pointer;
}

.event-type-tabs button.active {
  border-color: rgba(63, 127, 113, 0.34);
  color: var(--color-pine-700);
  background: rgba(221, 236, 229, 0.76);
}

.event-list {
  margin-top: 10px;
  display: grid;
  gap: 8px;
  max-height: 430px;
  overflow-y: auto;
  padding-right: 2px;
}

.event-item {
  width: 100%;
  text-align: left;
  border-radius: 12px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
  cursor: pointer;
}

.event-item.active {
  border-color: rgba(63, 127, 113, 0.45);
  box-shadow: 0 8px 20px rgba(63, 127, 113, 0.12);
}

.event-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.event-top p {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.event-level {
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
}

.event-level.high {
  color: #8a3c2f;
  background: rgba(230, 194, 185, 0.62);
}

.event-level.mid {
  color: #775325;
  background: rgba(234, 215, 180, 0.68);
}

.event-level.low {
  color: var(--color-pine-700);
  background: rgba(221, 236, 229, 0.72);
}

.event-title {
  margin: 6px 0 4px;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-main);
}

.event-meta {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.event-detail {
  margin-top: 8px;
}

.detail-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.event-detail h3 {
  margin: 6px 0 8px;
  font-size: 18px;
}

.detail-summary {
  margin: 0;
  line-height: 1.7;
  color: var(--color-text-sub);
}

.detail-grid {
  margin-top: 10px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
}

.detail-grid article {
  border-radius: 10px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(246, 244, 239, 0.8);
  padding: 8px 10px;
}

.detail-grid p {
  margin: 0 0 4px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.detail-grid strong {
  font-size: 13px;
  color: var(--color-text-main);
}

.detail-actions {
  margin-top: 10px;
}

.detail-actions button {
  border: 1px solid rgba(216, 223, 216, 0.9);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.95);
  color: var(--color-pine-700);
  padding: 8px 11px;
  cursor: pointer;
}

.detail-actions button:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

@media (max-width: 980px) {
  .strategy-hero,
  .strategy-grid,
  .event-section {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 760px) {
  .tag-row {
    flex-wrap: nowrap;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .tag-row::-webkit-scrollbar {
    display: none;
  }

  .tag-row span {
    flex: 0 0 auto;
  }

  .matrix-head {
    display: none;
  }

  .matrix-item {
    grid-template-columns: 1fr 1fr;
    gap: 8px;
  }

  .status {
    grid-column: 1 / -1;
  }

  .stock-detail-head {
    flex-direction: column;
  }

  .stock-detail-actions {
    width: 100%;
    flex-direction: column;
  }

  .stock-detail-head button {
    width: 100%;
  }

  .stock-detail-grid {
    grid-template-columns: 1fr;
  }

  .futures-detail-grid {
    grid-template-columns: 1fr;
  }

  .reason-support-grid {
    grid-template-columns: 1fr;
  }

  .stock-performance-head {
    flex-direction: column;
    align-items: flex-start;
  }

  .event-type-tabs {
    flex-wrap: nowrap;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .event-type-tabs::-webkit-scrollbar {
    display: none;
  }

  .event-type-tabs button {
    flex: 0 0 auto;
  }

  .event-list {
    max-height: none;
  }

  .detail-grid {
    grid-template-columns: 1fr;
  }
}
</style>
