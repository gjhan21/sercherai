<template>
  <section class="archive-page fade-up">
    <header class="archive-hero card">
      <div>
        <p class="archive-kicker">历史推荐档案</p>
        <h1>查看历史推荐结果与复盘信息</h1>
        <p class="archive-desc">
          统一查看历史推荐轨迹、状态、表现和数据来源。
        </p>
        <div class="archive-tags">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">完整历史轨迹</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">来源标识清晰</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">不只展示成功样本</span>
        </div>
      </div>
      <div class="archive-actions">
        <button class="primary-btn finance-primary-btn" type="button" :disabled="loading" @click="loadArchive">
          {{ loading ? "同步中..." : "刷新历史档案" }}
        </button>
        <button class="ghost-btn finance-ghost-btn" type="button" @click="goStrategies">去策略中心</button>
      </div>
      <div class="archive-hero-stats finance-hero-stat-grid">
        <article class="finance-hero-stat-card">
          <span>历史样本</span>
          <strong>{{ archiveRows.length }} 条</strong>
          <p>成功、失效、止盈与止损样本都应长期保留。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>当前展示</span>
          <strong>{{ visibleArchiveRows.length }} 条</strong>
          <p>
            <template v-if="archiveHiddenCount > 0">还有 {{ archiveHiddenCount }} 条等待更高权限展开。</template>
            <template v-else>当前筛选结果已全部展示。</template>
          </p>
        </article>
        <article class="finance-hero-stat-card">
          <span>当前筛选</span>
          <strong>{{ statusOptions.find((item) => item.value === selectedStatus)?.label || "全部样本" }}</strong>
          <p>{{ sourceOptions.find((item) => item.value === selectedSource)?.label || "全部来源" }} · {{ sourceGuide.length }} 类来源标签</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>阅读重点</span>
          <strong>理由 -> 结果 -> 后续处理</strong>
          <p>先看当时为什么选，再看结果和后续处理，不把档案页做成收益榜。</p>
        </article>
      </div>
    </header>

    <section class="archive-toolbar card">
      <div class="filter-group">
        <p>状态筛选</p>
        <div class="filter-row">
          <button
            v-for="item in statusOptions"
            :key="item.value"
            type="button"
            class="finance-toggle-btn"
            :class="{ active: selectedStatus === item.value }"
            @click="selectedStatus = item.value"
          >
            {{ item.label }}
          </button>
        </div>
      </div>
      <div class="filter-group">
        <p>业绩来源</p>
        <div class="filter-row">
          <button
            v-for="item in sourceOptions"
            :key="item.value"
            type="button"
            class="finance-toggle-btn"
            :class="{ active: selectedSource === item.value }"
            @click="selectedSource = item.value"
          >
            {{ item.label }}
          </button>
        </div>
      </div>
      <p v-if="loading" class="toolbar-note finance-note-strip finance-note-strip-info">正在同步历史推荐...</p>
      <p v-else-if="errorMessage" class="toolbar-note finance-note-strip finance-note-strip-warning">{{ errorMessage }}</p>
      <p v-else class="toolbar-note finance-note-strip finance-note-strip-info">
        已展示 {{ visibleArchiveRows.length }} 条档案
        <template v-if="archiveHiddenCount > 0"> · 还有 {{ archiveHiddenCount }} 条需要更高权限</template>
        · 更新时间：{{ lastUpdatedAt || "-" }}
      </p>
    </section>

    <StatePanel
      :tone="archiveAccessState.tone"
      :eyebrow="archiveAccessState.label"
      :title="archiveAccessState.title"
      :description="archiveAccessState.desc"
    >
      <template #actions>
        <button type="button" @click="handleArchivePrimaryAction('state_panel_primary')">{{ archivePrimaryActionText }}</button>
        <button type="button" class="ghost-btn finance-ghost-btn" @click="handleArchiveSecondaryAction">
          {{ archiveSecondaryActionText }}
        </button>
      </template>
    </StatePanel>

    <section class="archive-focus-layout">
      <article class="card archive-focus-card">
        <header class="archive-focus-head">
          <div>
            <p class="archive-kicker">档案概览</p>
            <h2 class="section-title">查看推荐结果、理由变化和复盘信息。</h2>
            <p class="section-subtitle">
              先看结果，再查看理由、版本变化和数据来源。
            </p>
          </div>
          <div class="archive-focus-actions">
            <button class="primary-btn finance-primary-btn" type="button" :disabled="loading" @click="loadArchive">
              {{ loading ? "同步中..." : "刷新完整档案" }}
            </button>
            <button class="ghost-btn finance-ghost-btn" type="button" @click="goStrategies">回策略中心</button>
          </div>
        </header>

        <div class="archive-focus-grid">
          <article class="finance-card-surface">
            <p>完整样本</p>
            <strong>{{ archiveRows.length }} 条</strong>
            <span>成功、失效、止盈与止损样本都应该长期保留。</span>
          </article>
          <article class="finance-card-surface">
            <p>当前展示</p>
            <strong>{{ visibleArchiveRows.length }} 条</strong>
            <span>
              <template v-if="archiveHiddenCount > 0">还有 {{ archiveHiddenCount }} 条因阶段权限未展开。</template>
              <template v-else>当前筛选下已全部展示。</template>
            </span>
          </article>
          <article class="finance-card-surface">
            <p>来源标签</p>
            <strong>{{ sourceGuide.length }} 类</strong>
            <span>已验证、预估、模拟、待补充都会标注清楚。</span>
          </article>
        </div>

        <div class="archive-read-order">
          <article class="finance-card-surface">
            <strong>先看结果</strong>
            <p>先看状态、累计收益和超额收益。</p>
          </article>
          <article class="finance-card-surface">
            <strong>再看理由与变化</strong>
            <p>查看当时推荐理由，以及后续版本变化。</p>
          </article>
          <article class="finance-card-surface">
            <strong>最后看数据来源</strong>
            <p>确认来源是否清晰，以及止盈止损说明是否完整。</p>
          </article>
        </div>
      </article>

      <aside class="archive-focus-side">
        <article class="card archive-side-card finance-card-surface">
          <h3>首页与档案页</h3>
          <p>首页展示历史样本预览，档案页展示完整时间线、版本轨迹和处理结果。</p>
        </article>
        <article class="card archive-side-card finance-card-surface">
          <h3>复盘重点</h3>
          <p>除了收益，也要查看推荐理由、风险边界和后续处理。</p>
        </article>
        <article class="card archive-side-card finance-card-surface">
          <h3>查看范围</h3>
          <p>
            <template v-if="archiveHiddenCount > 0">{{ archiveUpgradeHint }}</template>
            <template v-else>当前筛选结果已全部展示，可继续前往我的关注查看持续跟踪内容。</template>
          </p>
        </article>
      </aside>
    </section>

    <section class="archive-layout">
      <article class="card archive-list-card">
        <header class="section-head">
          <div>
            <h2 class="section-title">历史推荐时间线</h2>
            <p class="section-subtitle">按推荐时间倒序展示，让用户完整看到“推荐 -> 跟踪 -> 结果”的过程。</p>
          </div>
          <span class="section-badge finance-pill finance-pill-roomy finance-pill-neutral">{{ visibleArchiveRows.length }} 条</span>
        </header>

        <div v-if="visibleArchiveRows.length" class="archive-list">
          <article v-for="item in visibleArchiveRows" :key="item.id" class="archive-item finance-card-surface">
            <div class="archive-item-head">
              <div>
                <p class="archive-date">{{ item.date }}</p>
                <h3>{{ item.name }}</h3>
              </div>
              <div class="archive-badges">
                <span class="status finance-pill" :class="item.statusClass">{{ item.status }}</span>
                <span class="source-badge finance-pill finance-pill-roomy" :class="item.sourceClass">{{ item.sourceLabel }}</span>
              </div>
            </div>

            <p class="archive-reason">{{ item.reason }}</p>

            <div class="archive-explanation-box finance-card-pale">
              <div class="archive-explanation-head">
                <p>当时为什么选</p>
                <span v-if="item.metaText">{{ item.metaText }}</span>
              </div>
              <p class="archive-explanation-summary">{{ item.explanationSummary }}</p>
              <div v-if="item.proofTags.length" class="archive-proof-tags">
                <span
                  v-for="tag in item.proofTags"
                  :key="`${item.id}-${tag}`"
                  class="finance-pill finance-pill-compact finance-pill-info"
                >
                  {{ tag }}
                </span>
              </div>
              <p v-if="item.consensusText" class="archive-consensus">{{ item.consensusText }}</p>
            </div>

            <div v-if="item.seedHighlights.length || item.riskHighlights.length" class="archive-proof-grid">
              <article v-if="item.seedHighlights.length" class="finance-list-card finance-list-card-panel">
                <p>关键种子</p>
                <div class="archive-chip-list">
                  <span
                    v-for="seed in item.seedHighlights"
                    :key="`${item.id}-seed-${seed}`"
                    class="finance-pill finance-pill-compact finance-pill-neutral"
                  >
                    {{ seed }}
                  </span>
                </div>
              </article>
              <article v-if="item.riskHighlights.length" class="finance-list-card finance-list-card-panel">
                <p>风险与失效</p>
                <div class="archive-chip-list risk">
                  <span
                    v-for="risk in item.riskHighlights"
                    :key="`${item.id}-risk-${risk}`"
                    class="finance-pill finance-pill-compact finance-pill-warning"
                  >
                    {{ risk }}
                  </span>
                </div>
              </article>
            </div>

            <div v-if="item.versionDiff" class="archive-version-box finance-card-pale">
              <div class="archive-version-head">
                <p>版本差异</p>
                <span>{{ item.selectedHistoryTitle || item.versionDiff.meta }}</span>
              </div>
              <div class="archive-version-grid">
                <article class="finance-list-card finance-list-card-panel">
                  <p>{{ item.isCustomHistorySelected ? "所选版本" : "档案记录" }}</p>
                  <strong>{{ item.versionDiff.beforeLabel }}</strong>
                  <span>{{ item.versionDiff.beforeNote }}</span>
                </article>
                <article class="finance-list-card finance-list-card-panel">
                  <p>当前解释</p>
                  <strong>{{ item.versionDiff.afterLabel }}</strong>
                  <span>{{ item.versionDiff.afterNote }}</span>
                </article>
                <article class="finance-list-card finance-list-card-panel">
                  <p>为什么变了</p>
                  <strong>{{ item.versionDiff.diffLabel }}</strong>
                  <span>{{ item.versionDiff.diffNote }}</span>
                </article>
              </div>
            </div>

            <div v-if="item.versionHistoryItems.length" class="archive-history-box finance-list-card finance-list-card-panel">
              <div class="archive-version-head">
                <p>版本轨迹</p>
                <div class="archive-history-actions">
                  <span>后端 {{ item.versionHistoryItems.length }} 条</span>
                  <button
                    v-if="item.isCustomHistorySelected"
                    type="button"
                    class="archive-history-reset finance-mini-btn finance-mini-btn-soft"
                    @click="resetArchiveHistorySelection(item.id)"
                  >
                    回到默认对比
                  </button>
                </div>
              </div>
              <p class="archive-history-hint">点一条历史版本，直接和当前 explanation 对比。</p>
              <div class="archive-history-list">
                <button
                  v-for="history in item.versionHistoryItems"
                  :key="`${item.id}-${history.key}`"
                  class="archive-history-item finance-list-card finance-list-card-compact finance-list-card-interactive"
                  :class="{ active: item.selectedHistoryKey === history.key }"
                  type="button"
                  @click="selectArchiveHistory(item.id, history.key)"
                >
                  <p>{{ history.title }}</p>
                  <strong>{{ history.version }}</strong>
                  <span>{{ history.note }}</span>
                </button>
              </div>
            </div>

            <div v-if="item.l2Summary || item.relationshipSummary || item.agentOpinionRows.length" class="archive-proof-grid">
              <article v-if="item.l2Summary" class="finance-list-card finance-list-card-panel">
                <p>L2 情景摘要</p>
                <strong>{{ item.l2Summary.summary }}</strong>
                <span>{{ item.l2Summary.note || "当前未补更多主情景说明。" }}</span>
              </article>
              <article v-if="item.relationshipSummary" class="finance-list-card finance-list-card-panel">
                <p>关系快照</p>
                <strong>{{ item.relationshipSummary.summary }}</strong>
                <span>{{ item.relationshipSummary.note || "当前未补更多关系节点说明。" }}</span>
              </article>
              <article v-if="item.agentOpinionRows.length" class="finance-list-card finance-list-card-panel">
                <p>角色评审</p>
                <strong>{{ item.agentOpinionRows[0].role }} · {{ item.agentOpinionRows[0].stance }}</strong>
                <span>{{ item.agentOpinionRows[0].summary }}</span>
              </article>
            </div>

            <div v-if="item.originCards.length" class="archive-origin-grid">
              <article
                v-for="origin in item.originCards"
                :key="`${item.id}-${origin.label}`"
                class="finance-list-card finance-list-card-panel"
              >
                <p>{{ origin.label }}</p>
                <strong>{{ origin.value }}</strong>
                <span>{{ origin.note }}</span>
              </article>
            </div>

            <div class="archive-metrics">
              <article>
                <p>风险等级</p>
                <strong>{{ item.risk }}</strong>
              </article>
              <article>
                <p>累计收益</p>
                <strong :class="item.cumulativeClass">{{ item.cumulativeReturn }}</strong>
              </article>
              <article>
                <p>基准累计</p>
                <strong :class="item.benchmarkClass">{{ item.benchmarkReturn }}</strong>
              </article>
              <article>
                <p>累计超额</p>
                <strong :class="item.excessClass">{{ item.excessReturn }}</strong>
              </article>
            </div>

            <div class="archive-timeline">
              <article v-for="step in item.timeline" :key="`${item.id}-${step.label}`">
                <p>{{ step.label }}</p>
                <strong>{{ step.value }}</strong>
                <span>{{ step.note }}</span>
              </article>
            </div>

            <p class="archive-source-note">{{ item.sourceNote }}</p>

            <div class="archive-actions-row">
              <button type="button" class="ghost-btn finance-ghost-btn" @click="goStrategies">去策略中心</button>
            </div>
          </article>
        </div>
        <StatePanel
          v-else
          tone="info"
          eyebrow="历史档案"
          :title="archiveAccessLocked ? '登录后解锁完整历史档案' : '暂无可展示的历史推荐'"
          :description="
            archiveAccessLocked
              ? '当前先展示档案结构与复盘说明；登录后会自动补齐历史推荐时间线。'
              : '可以先刷新历史档案，或去策略中心查看当前主推荐。'
          "
          compact
        >
          <template #actions>
            <button type="button" @click="loadArchive">刷新历史档案</button>
            <button type="button" class="ghost-btn finance-ghost-btn" @click="goStrategies">去策略中心</button>
          </template>
        </StatePanel>
      </article>

      <aside class="card trust-card">
        <header class="section-head compact">
          <div>
            <h2 class="section-title">如何读懂来源标识</h2>
            <p class="section-subtitle">不同来源类型会对应不同参考强度。</p>
          </div>
        </header>
        <div class="trust-list">
          <article v-for="item in sourceGuide" :key="item.label" class="finance-card-surface">
            <div class="trust-top">
              <span class="source-badge finance-pill finance-pill-roomy" :class="item.className">{{ item.label }}</span>
              <strong>{{ item.title }}</strong>
            </div>
            <p>{{ item.desc }}</p>
          </article>
        </div>
        <div class="trust-note finance-card-surface">
          <p>查看原则</p>
          <strong>所有结果都会附带来源标识。</strong>
          <span>查看历史表现时，请结合来源类型一起判断。</span>
        </div>
        <div v-if="archiveHiddenCount > 0" class="trust-note upgrade-note finance-card-surface">
          <p>阶段边界</p>
          <strong>{{ archiveHiddenCount }} 条历史档案暂未展示。</strong>
          <span>{{ archiveUpgradeHint }}</span>
        </div>
      </aside>
    </section>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import StatePanel from "../components/StatePanel.vue";
import { getMembershipQuota } from "../api/membership";
import {
  getStockRecommendationInsight,
  getStockRecommendationVersionHistory,
  listStockRecommendations
} from "../api/market";
import { useClientAuth } from "../lib/client-auth";
import { shouldUseDemoFallback } from "../lib/fallback-policy";
import {
  promotePendingExperimentJourneySources,
  rememberExperimentAttributionSource,
  rememberPendingExperimentJourneySource
} from "../lib/growth-analytics";
import { getExperimentVariant } from "../lib/growth-experiments";
import {
  buildStrategyConfidenceCalibrationSummary,
  buildFallbackStrategyVersionHistory,
  buildStrategyAgentOpinionRows,
  buildStrategyHistoryCompareState,
  buildStrategyInsightSections,
  buildStrategyMetaText,
  buildStrategyOriginCards,
  buildStrategyProofTags,
  buildStrategyRelationshipSnapshotSummary,
  buildStrategyScenarioMetaSummary,
  buildStrategyScenarioSnapshotRows,
  buildStrategyThesisCardRows,
  buildStrategyRiskHighlights,
  buildStrategyWatchSignalRows,
  findMatchedStrategyHistoryItem,
  firstMeaningfulStrategyText,
  mapStrategyVersionHistory,
  toStrategyTradeDate
} from "../lib/strategy-version";
import { buildProfileModuleRoute } from "../lib/profile-modules";

const fallbackRecommendations = [
  {
    id: "sr_archive_001",
    symbol: "600519.SH",
    name: "贵州茅台",
    score: 91.2,
    risk_level: "MEDIUM",
    position_range: "10%-15%",
    valid_from: "2026-02-20T09:30:00+08:00",
    valid_to: "2026-02-27T15:00:00+08:00",
    status: "REVIEWED",
    reason_summary: "消费龙头估值修复，资金流与基本面共振。"
  },
  {
    id: "sr_archive_002",
    symbol: "300750.SZ",
    name: "宁德时代",
    score: 87.4,
    risk_level: "HIGH",
    position_range: "8%-12%",
    valid_from: "2026-02-18T09:30:00+08:00",
    valid_to: "2026-02-24T15:00:00+08:00",
    status: "INVALIDATED",
    reason_summary: "趋势尝试修复，但量能与板块情绪未持续确认。"
  },
  {
    id: "sr_archive_003",
    symbol: "601318.SH",
    name: "中国平安",
    score: 84.8,
    risk_level: "LOW",
    position_range: "12%-18%",
    valid_from: "2026-02-14T09:30:00+08:00",
    valid_to: "2026-02-21T15:00:00+08:00",
    status: "HIT_TAKE_PROFIT",
    reason_summary: "高股息与低估值防御逻辑成立。"
  }
];

const fallbackInsights = {
  sr_archive_001: {
    performance: [
      { date: "2026-02-21", return: 0.008 },
      { date: "2026-02-24", return: 0.009 },
      { date: "2026-02-25", return: 0.006 }
    ],
    benchmark: [
      { date: "2026-02-21", return: 0.004 },
      { date: "2026-02-24", return: 0.005 },
      { date: "2026-02-25", return: 0.003 }
    ],
    performance_stats: {
      sample_days: 3,
      cumulative_return: 0.0232,
      benchmark_cumulative_return: 0.0121,
      excess_return: 0.0111,
      benchmark_symbol: "000300.SH",
      benchmark_source: "actual: CSI300"
    }
  },
  sr_archive_002: {
    performance: [
      { date: "2026-02-19", return: 0.004 },
      { date: "2026-02-20", return: -0.013 },
      { date: "2026-02-21", return: -0.009 }
    ],
    benchmark: [
      { date: "2026-02-19", return: 0.003 },
      { date: "2026-02-20", return: -0.004 },
      { date: "2026-02-21", return: -0.003 }
    ],
    performance_stats: {
      sample_days: 3,
      cumulative_return: -0.0182,
      benchmark_cumulative_return: -0.0040,
      excess_return: -0.0142,
      benchmark_symbol: "创业板指",
      benchmark_source: "estimated: 55% of strategy daily return"
    }
  },
  sr_archive_003: {
    performance: [
      { date: "2026-02-17", return: 0.007 },
      { date: "2026-02-18", return: 0.006 },
      { date: "2026-02-19", return: 0.005 }
    ],
    benchmark: [
      { date: "2026-02-17", return: 0.004 },
      { date: "2026-02-18", return: 0.003 },
      { date: "2026-02-19", return: 0.002 }
    ],
    performance_stats: {
      sample_days: 3,
      cumulative_return: 0.0181,
      benchmark_cumulative_return: 0.0090,
      excess_return: 0.0091,
      benchmark_symbol: "上证50",
      benchmark_source: "simulated: fallback benchmark series"
    }
  }
};

const sourceOptions = [
  { value: "ALL", label: "全部来源" },
  { value: "VERIFIED", label: "已验证" },
  { value: "ESTIMATED", label: "预估" },
  { value: "SIMULATED", label: "模拟" },
  { value: "PENDING", label: "待说明" }
];

const sourceGuide = [
  {
    label: "已验证",
    title: "优先参考",
    desc: "说明基准或表现来源已标识为实际数据，可作为更强的跟踪依据。",
    className: "verified"
  },
  {
    label: "预估",
    title: "谨慎参考",
    desc: "说明表现或基准包含估算逻辑，适合作为趋势判断，不宜理解为已验证结果。",
    className: "estimated"
  },
  {
    label: "模拟",
    title: "仅作参考",
    desc: "说明展示值来自模拟数据，适合作为趋势参考，不代表真实验证结果。",
    className: "simulated"
  },
  {
    label: "待说明",
    title: "需要补充",
    desc: "说明当前缺少足够来源信息，需要补充后再判断。",
    className: "pending"
  }
];

const useDemoFallback = shouldUseDemoFallback();
const router = useRouter();
const { isLoggedIn } = useClientAuth();
const archiveMembershipExperimentVariant = getExperimentVariant("archive_membership_entry", ["default"]);
const loading = ref(false);
const errorMessage = ref("");
const lastUpdatedAt = ref("");
const selectedStatus = ref("ALL");
const selectedSource = ref("ALL");
const rawItems = ref(useDemoFallback ? [...fallbackRecommendations] : []);
const insightMap = ref(useDemoFallback ? { ...fallbackInsights } : {});
const versionHistoryMap = ref({});
const selectedArchiveHistoryKeys = ref({});
const memberStageLoading = ref(false);
const isVIPUser = ref(false);
const archiveAccessLocked = ref(false);

const archiveRows = computed(() =>
  (rawItems.value || []).map((item, index) => {
    const insight = insightMap.value[item.id] || {};
    const explanation = insight.explanation || buildFallbackArchiveExplanation(item, index);
    const sections = buildStrategyInsightSections(explanation, item.reason_summary || "暂无推荐理由");
    const versionHistoryItems = buildArchiveHistoryItems(item, explanation, versionHistoryMap.value[item.id]);
    const matchedHistoryItem = findArchiveMatchedHistoryItem(item, versionHistoryItems);
    const historyCompare = buildStrategyHistoryCompareState({
      historyItems: versionHistoryItems,
      selectedKey: selectedArchiveHistoryKeys.value[item.id],
      fallbackItem: matchedHistoryItem,
      explanation,
      selectedTitlePrefix: "当前对比 ",
      formatDateTime,
      fallbackRecordLabel: "record",
      upgradedText: "原始档案和当前 explanation 的版本与结论都发生了变化。",
      reasonChangedText: "当前 explanation 对原始推荐理由做了新的收敛说明。",
      versionChangedText: "结论主线仍在，但解释版本已经更新。"
    });
    const stats = insight.performance_stats || {};
    const sourceMeta = resolveSourceMeta(stats.benchmark_source);
    const cumulative = Number(stats.cumulative_return);
    const benchmark = Number(stats.benchmark_cumulative_return);
    const excess = Number(stats.excess_return);
    const historicalThesis = buildStrategyThesisCardRows(explanation, "historical", { limit: 1 })[0];
    const watchSignal = buildStrategyWatchSignalRows(explanation, { limit: 1 })[0];
    const calibration = buildStrategyConfidenceCalibrationSummary(explanation);
    const summarySource = historyCompare.selectedItem || explanation;
    const scenarioMetaSummary = buildStrategyScenarioMetaSummary(summarySource);
    const relationshipSummary = buildStrategyRelationshipSnapshotSummary(summarySource);
    const agentOpinionRows = buildStrategyAgentOpinionRows(summarySource, { limit: 2 });
    const scenarioRows = buildStrategyScenarioSnapshotRows(summarySource, { limit: 2 });
    return {
      id: item.id,
      name: `${item.symbol || "-"} ${item.name || ""}`.trim(),
      date: formatDate(item.valid_from),
      reason: item.reason_summary || "暂无推荐理由",
      explanationSummary:
        firstMeaningfulStrategyText([historicalThesis?.summary, sections.whyNow, item.reason_summary]) || "暂无推荐理由",
      proofTags: buildArchiveProofTags(explanation),
      consensusText: firstMeaningfulStrategyText([
        historicalThesis?.title,
        scenarioMetaSummary?.summary,
        scenarioRows[0]?.action,
        calibration?.summary,
        sections.proofSource
      ]),
      l2Summary: scenarioMetaSummary,
      relationshipSummary,
      agentOpinionRows,
      seedHighlights: Array.isArray(explanation.seed_highlights) ? explanation.seed_highlights.slice(0, 3) : [],
      riskHighlights: buildStrategyRiskHighlights(explanation, {
        limit: 3,
        fallback: firstMeaningfulStrategyText([watchSignal?.trigger, calibration?.deltaLabel, buildStatusNote(item.status)])
      }),
      metaText: buildArchiveMetaText(explanation),
      versionDiff: historyCompare.diff,
      versionHistoryItems: versionHistoryItems.slice(0, 3),
      selectedHistoryKey: historyCompare.selectedKey,
      selectedHistoryTitle: historyCompare.selectedTitle,
      isCustomHistorySelected: historyCompare.isCustomSelected,
      originCards: buildStrategyOriginCards(explanation, formatDateTime),
      risk: mapRisk(item.risk_level),
      status: mapStatus(item.status).label,
      statusClass: mapStatus(item.status).className,
      sourceType: sourceMeta.key,
      sourceLabel: sourceMeta.label,
      sourceClass: sourceMeta.className,
      sourceNote: buildSourceNote(stats, sourceMeta),
      cumulativeReturn: formatPercent(cumulative),
      benchmarkReturn: formatPercent(benchmark),
      excessReturn: formatPercent(excess),
      cumulativeClass: trendClassByNumber(cumulative),
      benchmarkClass: trendClassByNumber(benchmark),
      excessClass: trendClassByNumber(excess),
      timeline: [
        {
          label: "推荐生效",
          value: formatDate(item.valid_from),
          note: `观察期 ${formatDateRange(item.valid_from, item.valid_to)}`
        },
        {
          label: "当前状态",
          value: mapStatus(item.status).label,
          note: buildStatusNote(item.status)
        },
        {
          label: "来源说明",
          value: sourceMeta.label,
          note: stats.benchmark_source || "当前缺少更细的来源字段说明"
        }
      ]
    };
  })
);

const statusOptions = computed(() => {
  const seen = new Map([["ALL", "全部状态"]]);
  archiveRows.value.forEach((item) => {
    if (!seen.has(item.status)) {
      seen.set(item.status, item.status);
    }
  });
  return Array.from(seen.entries()).map(([value, label]) => ({ value, label }));
});

const filteredRows = computed(() =>
  archiveRows.value.filter((item) => {
    if (selectedStatus.value !== "ALL" && item.status !== selectedStatus.value) {
      return false;
    }
    if (selectedSource.value !== "ALL" && item.sourceType !== selectedSource.value) {
      return false;
    }
    return true;
  })
);
const archiveAccessStage = computed(() => {
  if (isVIPUser.value) {
    return "VIP";
  }
  if (isLoggedIn.value) {
    return "REGISTERED";
  }
  return "VISITOR";
});
const visibleArchiveRows = computed(() => {
  if (archiveAccessStage.value === "VIP") {
    return filteredRows.value;
  }
  if (archiveAccessStage.value === "REGISTERED") {
    return filteredRows.value.slice(0, 4);
  }
  return filteredRows.value.slice(0, 2);
});
const archiveHiddenCount = computed(() => Math.max(0, filteredRows.value.length - visibleArchiveRows.value.length));
const archiveAccessState = computed(() => {
  if (memberStageLoading.value) {
    return {
      tone: "info",
      label: "识别中",
      title: "正在确认你的历史档案访问阶段",
      desc: "确认完成后，会按游客 / 注册 / 会员显示对应数量和 CTA。"
    };
  }
  if (archiveAccessStage.value === "VIP") {
    return {
      tone: "success",
      label: "会员阶段",
      title: "你已解锁完整历史档案，可查看来源和复盘信息。",
      desc: "建议结合我的关注继续查看后续跟踪内容。"
    };
  }
  if (archiveAccessStage.value === "REGISTERED") {
    return {
      tone: "warning",
      label: "注册阶段",
      title: "你现在能看更多历史样本，可进一步判断是否需要完整复盘。",
      desc: "可先比较成功、失效和止损样本。"
    };
  }
  return {
    tone: "info",
    label: "游客阶段",
    title: "可先查看部分历史档案。",
    desc: "登录后会解锁更多历史样本。"
  };
});
const archivePrimaryActionText = computed(() => {
  if (archiveAccessStage.value === "VIP") {
    return "去我的关注";
  }
  if (archiveAccessStage.value === "REGISTERED") {
    return "升级会员看完整复盘";
  }
  return "登录查看更多档案";
});
const archiveSecondaryActionText = computed(() => {
  if (archiveAccessStage.value === "VIP") {
    return "回策略页";
  }
  if (archiveAccessStage.value === "REGISTERED") {
    return "回策略页";
  }
  return "回首页";
});
const archiveUpgradeHint = computed(() => {
  if (archiveAccessStage.value === "REGISTERED") {
    return "升级会员后，可继续查看完整历史档案和更完整的复盘信息。";
  }
  return "登录后可查看更多历史档案，再决定是否升级会员查看完整复盘。";
});

async function loadArchive() {
  loading.value = true;
  errorMessage.value = "";
  archiveAccessLocked.value = false;
  if (!isLoggedIn.value) {
    archiveAccessLocked.value = true;
    rawItems.value = useDemoFallback ? [...fallbackRecommendations] : [];
    insightMap.value = useDemoFallback ? { ...fallbackInsights } : {};
    versionHistoryMap.value = {};
    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
    loading.value = false;
    return;
  }
  try {
    const data = await listStockRecommendations({ page: 1, page_size: 20 });
    const items = Array.isArray(data?.items) ? data.items : [];
    if (items.length > 0) {
      rawItems.value = items;
      await hydrateInsights(items);
    } else if (!useDemoFallback) {
      rawItems.value = [];
    }
    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  } catch (error) {
    if (!isLoggedIn.value && isAuthBlockedError(error)) {
      archiveAccessLocked.value = true;
      rawItems.value = [];
      errorMessage.value = "";
      lastUpdatedAt.value = formatDateTime(new Date().toISOString());
    } else {
      errorMessage.value = parseErrorMessage(error);
    }
  } finally {
    loading.value = false;
  }
}

function selectArchiveHistory(itemID, historyKey) {
  if (!itemID || !historyKey) {
    return;
  }
  selectedArchiveHistoryKeys.value = {
    ...selectedArchiveHistoryKeys.value,
    [itemID]: historyKey
  };
}

function resetArchiveHistorySelection(itemID) {
  if (!itemID || !selectedArchiveHistoryKeys.value[itemID]) {
    return;
  }
  const next = { ...selectedArchiveHistoryKeys.value };
  delete next[itemID];
  selectedArchiveHistoryKeys.value = next;
}

async function loadMembershipStage() {
  memberStageLoading.value = true;
  try {
    if (!isLoggedIn.value) {
      isVIPUser.value = false;
      return;
    }
    const quota = await getMembershipQuota();
    isVIPUser.value = resolveVIPStage(quota);
  } catch {
    isVIPUser.value = false;
  } finally {
    memberStageLoading.value = false;
  }
  promoteArchivePostAuthAttribution();
}

async function hydrateInsights(items) {
  const nextInsightMap = { ...insightMap.value };
  const nextVersionHistoryMap = { ...versionHistoryMap.value };
  const errors = [];
  await Promise.all(
    items.map(async (item) => {
      if (!item?.id) {
        return;
      }
      const [insightResult, versionHistoryResult] = await Promise.allSettled([
        getStockRecommendationInsight(item.id),
        getStockRecommendationVersionHistory(item.id)
      ]);

      if (insightResult.status === "fulfilled" && insightResult.value) {
        nextInsightMap[item.id] = insightResult.value;
      } else if (insightResult.status === "rejected") {
        errors.push(`${item.symbol || item.id} 档案补全失败：${parseErrorMessage(insightResult.reason)}`);
      }

      if (versionHistoryResult.status === "fulfilled") {
        nextVersionHistoryMap[item.id] = Array.isArray(versionHistoryResult.value?.items)
          ? versionHistoryResult.value.items
          : [];
      } else if (versionHistoryResult.status === "rejected") {
        errors.push(`${item.symbol || item.id} 版本轨迹补全失败：${parseErrorMessage(versionHistoryResult.reason)}`);
      }
    })
  );
  insightMap.value = nextInsightMap;
  versionHistoryMap.value = nextVersionHistoryMap;
  if (errors.length > 0) {
    errorMessage.value = errors.join("；");
  }
}

function resolveSourceMeta(sourceText) {
  const source = String(sourceText || "").toLowerCase();
  if (source.includes("actual")) {
    return { key: "VERIFIED", label: "已验证", className: "verified" };
  }
  if (source.includes("estimated")) {
    return { key: "ESTIMATED", label: "预估", className: "estimated" };
  }
  if (source.includes("simulated") || source.includes("fallback")) {
    return { key: "SIMULATED", label: "模拟", className: "simulated" };
  }
  if (!source) {
    return { key: "PENDING", label: "待说明", className: "pending" };
  }
  return { key: "PENDING", label: "待说明", className: "pending" };
}

function buildSourceNote(stats, sourceMeta) {
  const benchmark = stats?.benchmark_symbol ? `${stats.benchmark_symbol} ` : "";
  if (sourceMeta.key === "VERIFIED") {
    return `当前基准 ${benchmark}已标识为实际来源，可作为更强的跟踪参考。`;
  }
  if (sourceMeta.key === "ESTIMATED") {
    return `当前累计与基准对比包含预估逻辑，适合辅助判断，不宜视为已验证结果。`;
  }
  if (sourceMeta.key === "SIMULATED") {
    return `当前展示值含模拟或 fallback 成分，主要用于说明推荐过程和比较框架。`;
  }
  return `当前缺少充分的来源标识，后续应优先补充后端来源字段。`;
}

function buildStatusNote(status) {
  const mapped = mapStatus(status).label;
  if (mapped === "已止盈") {
    return "推荐已进入兑现阶段，适合回看止盈过程。";
  }
  if (mapped === "已止损") {
    return "推荐已触发风险边界，应优先查看失效原因。";
  }
  if (mapped === "已失效") {
    return "逻辑失效不等于简单亏损，关键是看为什么失效。";
  }
  if (mapped === "已复盘") {
    return "推荐周期已经结束，可作为完整历史样本查看。";
  }
  return "推荐仍在历史样本池中，可结合档案和详情继续回看。";
}

function buildFallbackArchiveExplanation(item, index = 0) {
  const reason = item?.reason_summary || "系统从当日市场信号中筛出该标的。";
  return {
    seed_summary: `已处理与 ${item?.symbol || item?.name || "标的"} 相关的资金、技术、板块和事件种子。`,
    seed_highlights: [reason, `${mapRisk(item?.risk_level)}边界`, "板块确认"].filter(Boolean),
    graph_summary: "系统会把板块、事件、资金和风险边界串成可回看的因果链。",
    consensus_summary: "趋势、事件和风险视角先形成共识，再决定是否发布到推荐池。",
    risk_flags: [`${mapRisk(item?.risk_level)}条件下需严格执行仓位纪律`],
    invalidations: [buildStatusNote(item?.status)],
    confidence_reason: reason,
    workload_summary: {
      seed_count: 22 + index * 3,
      candidate_count: 8,
      selected_count: 1,
      agent_count: 5,
      scenario_count: 4,
      filter_steps: ["seed-ingest", "market-graph", "scenario-engine", "agent-review"]
    },
    strategy_version: item?.strategy_version || "strategy-engine-archive",
    generated_at: item?.valid_from || ""
  };
}

function buildArchiveProofTags(explanation) {
  return buildStrategyProofTags(explanation, { includeSeedCount: true });
}

function buildArchiveMetaText(explanation) {
  return buildStrategyMetaText(explanation, formatDateTime, { includeBatch: true, includeJob: false });
}

function buildArchiveHistoryItems(item, explanation, historyItems) {
  const normalized = mapStrategyVersionHistory(historyItems, formatDateTime);
  if (normalized.length > 0) {
    return normalized;
  }
  return buildFallbackStrategyVersionHistory(explanation, {
    reasonSummary: item?.reason_summary || "",
    strategyVersion: explanation?.strategy_version || item?.strategy_version || "record",
    tradeDate: explanation?.trade_date || toStrategyTradeDate(item?.valid_from),
    formatDateTime
  });
}

function normalizeTextList(items) {
  if (!Array.isArray(items)) {
    return [];
  }
  return items
    .map((item) => String(item || "").trim())
    .filter(Boolean)
    .slice(0, 3);
}

function findArchiveMatchedHistoryItem(item, historyItems) {
  return findMatchedStrategyHistoryItem(historyItems, {
    tradeDate: item?.valid_from,
    publishVersion: item?.publish_version
  });
}

function goStrategies() {
  router.push("/strategies");
}

function goHome() {
  router.push("/home");
}

function rememberArchiveMembershipEntry(targetKey, metadata = {}) {
  rememberExperimentAttributionSource({
    experimentKey: "archive_membership_entry",
    variantKey: archiveMembershipExperimentVariant,
    pageKey: "archive",
    targetKey,
    userStage: archiveAccessStage.value,
    metadata: {
      selected_status: selectedStatus.value,
      selected_source: selectedSource.value,
      visible_count: visibleArchiveRows.value.length,
      hidden_count: archiveHiddenCount.value,
      ...metadata
    }
  });
}

function rememberArchivePendingMembershipEntry(targetKey, metadata = {}) {
  rememberPendingExperimentJourneySource({
    experimentKey: "archive_membership_entry",
    variantKey: archiveMembershipExperimentVariant,
    pageKey: "archive",
    targetKey,
    userStage: "VISITOR",
    redirectPath: "/archive",
    metadata: {
      selected_status: selectedStatus.value,
      selected_source: selectedSource.value,
      visible_count: visibleArchiveRows.value.length,
      hidden_count: archiveHiddenCount.value,
      ...metadata
    }
  });
}

function promoteArchivePostAuthAttribution() {
  if (!isLoggedIn.value) {
    return;
  }
  promotePendingExperimentJourneySources({
    experimentKey: "archive_membership_entry",
    pageKey: "archive",
    userStage: archiveAccessStage.value,
    metadata: {
      selected_status: selectedStatus.value,
      selected_source: selectedSource.value,
      visible_count: visibleArchiveRows.value.length,
      hidden_count: archiveHiddenCount.value
    }
  });
}

function handleArchivePrimaryAction(targetKey = "primary_action") {
  if (archiveAccessStage.value === "VIP") {
    router.push(buildProfileModuleRoute("watchlist"));
    return;
  }
  if (archiveAccessStage.value === "REGISTERED") {
    rememberArchiveMembershipEntry(targetKey, {
      destination: "/membership"
    });
    router.push("/membership");
    return;
  }
  rememberArchivePendingMembershipEntry(targetKey, {
    destination_after_auth: "/archive"
  });
  router.push({ path: "/auth", query: { redirect: "/archive" } });
}

function handleArchiveSecondaryAction() {
  if (archiveAccessStage.value === "VISITOR") {
    goHome();
    return;
  }
  goStrategies();
}

function mapRisk(level) {
  const source = String(level || "").toUpperCase();
  if (source === "LOW") return "低风险";
  if (source === "MEDIUM") return "中风险";
  if (source === "HIGH") return "高风险";
  return source || "-";
}

function mapStatus(status) {
  const source = String(status || "").toUpperCase();
  if (source === "PUBLISHED" || source === "ACTIVE") {
    return { label: "已发布", className: "good" };
  }
  if (source === "TRACKING") {
    return { label: "跟踪中", className: "good" };
  }
  if (source === "HIT_TAKE_PROFIT") {
    return { label: "已止盈", className: "good" };
  }
  if (source === "HIT_STOP_LOSS") {
    return { label: "已止损", className: "risky" };
  }
  if (source === "INVALIDATED") {
    return { label: "已失效", className: "risky" };
  }
  if (source === "REVIEWED") {
    return { label: "已复盘", className: "normal" };
  }
  return { label: source || "待确认", className: "watch" };
}

function resolveVIPStage(quota) {
  const activationState = String(quota?.activation_state || "").toUpperCase();
  if (activationState) {
    return activationState === "ACTIVE";
  }
  const status = String(quota?.vip_status || "").toUpperCase();
  if (status === "ACTIVE") {
    return true;
  }
  const level = String(quota?.member_level || "").toUpperCase();
  if (!level.startsWith("VIP")) {
    return false;
  }
  const remainingDays = Number(quota?.vip_remaining_days);
  if (Number.isFinite(remainingDays)) {
    return remainingDays > 0;
  }
  return true;
}

function formatPercent(value) {
  if (!Number.isFinite(value)) {
    return "-";
  }
  const percent = (value * 100).toFixed(2);
  if (Number(percent) > 0) {
    return `+${percent}%`;
  }
  return `${percent}%`;
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

function formatDate(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return "-";
  }
  return new Date(ts).toLocaleDateString("zh-CN");
}

function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return "-";
  }
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

function formatDateRange(start, end) {
  return `${formatDate(start)} - ${formatDate(end)}`;
}

function parseErrorMessage(error) {
  if (!error) {
    return "数据请求失败";
  }
  if (typeof error === "string") {
    return error;
  }
  return error?.message || error?.response?.data?.message || "数据请求失败";
}

function isAuthBlockedError(error) {
  const status = Number(error?.code || error?.original?.response?.status || 0);
  const message = String(error?.message || error?.original?.response?.data?.message || "")
    .replace(/\s+/g, " ")
    .trim()
    .toLowerCase();
  return status === 401 || message.includes("authorization") || message.includes("unauthorized");
}

onMounted(() => {
  loadMembershipStage();
  loadArchive();
});

watch(
  () => isLoggedIn.value,
  () => {
    loadMembershipStage();
  }
);
</script>

<style scoped>
.archive-page {
  display: grid;
  gap: 14px;
}

.archive-hero,
.archive-toolbar,
.archive-list-card,
.trust-card {
  padding: 14px;
}

.archive-hero {
  display: grid;
  gap: 14px;
  grid-template-columns: 1.1fr 0.9fr;
  background:
    radial-gradient(circle at 0% 0%, var(--color-focus-glow) 0%, transparent 34%),
    radial-gradient(circle at 100% 100%, var(--color-line-gold) 0%, transparent 36%),
    rgba(255, 255, 255, 0.94);
}

.archive-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.archive-hero h1 {
  margin: 8px 0 10px;
  font-family: var(--font-serif);
  font-size: clamp(28px, 3.4vw, 42px);
  line-height: 1.2;
}

.archive-desc {
  margin: 0;
  color: var(--color-text-sub);
  line-height: 1.72;
}

.archive-tags {
  margin-top: 14px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.archive-hero-stats {
  grid-column: 1 / -1;
}

.section-badge,
.source-badge {
  font-weight: 700;
}

.archive-actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
  justify-content: center;
}

.archive-toolbar {
  display: grid;
  gap: 12px;
}

.filter-group {
  display: grid;
  gap: 8px;
}

.filter-group p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.filter-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-row button {
  white-space: nowrap;
}

.filter-row button.active {
  border-color: var(--color-border-focus-medium);
  box-shadow: inset 0 0 0 1px var(--color-focus-fill);
}

.toolbar-note {
  margin: 0;
}

.archive-focus-layout {
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, 1.2fr) 320px;
}

.archive-focus-card,
.archive-side-card {
  padding: 14px;
}

.archive-focus-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.archive-focus-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.archive-focus-grid,
.archive-read-order,
.archive-focus-side {
  display: grid;
  gap: 10px;
}

.archive-focus-grid {
  margin-top: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.archive-focus-grid p,
.archive-read-order p,
.archive-side-card p {
  margin: 0;
  color: var(--color-text-sub);
  line-height: 1.7;
}

.archive-focus-grid p {
  font-size: 12px;
}

.archive-focus-grid strong,
.archive-read-order strong,
.archive-side-card h3 {
  color: var(--color-text-main);
}

.archive-focus-grid strong {
  display: block;
  margin-top: 6px;
  font-size: 22px;
}

.archive-focus-grid span {
  display: block;
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.7;
  color: var(--color-text-sub);
}

.archive-read-order {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.archive-read-order strong {
  display: block;
  margin-bottom: 6px;
  font-size: 14px;
}

.archive-side-card h3 {
  margin: 0 0 8px;
  font-size: 15px;
}

.archive-layout {
  display: grid;
  gap: 12px;
  grid-template-columns: 1.1fr 0.9fr;
}

.section-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.section-head.compact {
  margin-bottom: 10px;
}

.archive-list {
  margin-top: 12px;
  display: grid;
  gap: 12px;
}

.archive-item {
  display: grid;
  gap: 10px;
}

.archive-item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.archive-date {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.archive-item-head h3 {
  margin: 5px 0 0;
  font-size: 20px;
  line-height: 1.3;
}

.archive-badges {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.status {
  font-weight: 700;
}

.status.good {
  color: var(--color-pine-700);
  background: var(--color-surface-accent-soft);
  border-color: var(--color-border-focus-soft);
}

.status.watch {
  color: var(--color-warning);
  background: var(--color-surface-accent-strong);
}

.status.risky {
  color: var(--color-fall);
  background: var(--color-fall-soft);
}

.status.normal {
  color: var(--color-text-sub);
  background: var(--color-surface-panel-soft-subtle);
}

.source-badge.verified {
  color: var(--color-rise);
  background: var(--color-rise-soft);
}

.source-badge.estimated {
  color: var(--color-accent);
  background: rgba(184, 130, 48, 0.7);
}

.source-badge.simulated {
  color: var(--color-pine-700);
  background: var(--color-surface-accent-glow);
}

.source-badge.pending {
  color: var(--color-text-sub);
  background: var(--color-surface-panel-soft-subtle);
}

.archive-reason,
.archive-source-note {
  margin: 0;
  color: var(--color-text-sub);
  line-height: 1.7;
}

.archive-explanation-box {
  display: grid;
  gap: 8px;
}

.archive-explanation-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.archive-explanation-head p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.archive-explanation-head span {
  font-size: 11px;
  color: var(--color-text-sub);
}

.archive-explanation-summary,
.archive-consensus {
  margin: 0;
  font-size: 13px;
  line-height: 1.68;
  color: var(--color-text-sub);
}

.archive-proof-tags,
.archive-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.archive-proof-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.archive-proof-grid article {
  display: grid;
  gap: 8px;
}

.archive-origin-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.archive-origin-grid article {
  display: grid;
  gap: 6px;
}

.archive-version-box {
  display: grid;
  gap: 8px;
}

.archive-history-box {
  display: grid;
  gap: 8px;
}

.archive-version-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.archive-version-head p,
.archive-version-head span {
  margin: 0;
}

.archive-version-head p {
  font-size: 13px;
  font-weight: 600;
}

.archive-version-head span {
  font-size: 11px;
  color: var(--color-text-sub);
}

.archive-history-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.archive-history-reset {
  white-space: nowrap;
}

.archive-version-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.archive-history-list {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.archive-history-hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.archive-history-item {
  appearance: none;
  text-align: left;
  display: grid;
  gap: 4px;
}

.archive-version-grid article {
  display: grid;
  gap: 4px;
}

.archive-version-grid p,
.archive-version-grid strong,
.archive-version-grid span,
.archive-history-item p,
.archive-history-item strong,
.archive-history-item span {
  margin: 0;
}

.archive-version-grid p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.archive-history-item p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.archive-version-grid strong {
  font-size: 14px;
  color: var(--color-text-main);
}

.archive-history-item strong {
  font-size: 14px;
  color: var(--color-text-main);
}

.archive-version-grid span {
  font-size: 12px;
  line-height: 1.55;
  color: var(--color-text-sub);
}

.archive-history-item span {
  font-size: 12px;
  line-height: 1.55;
  color: var(--color-text-sub);
}

.archive-proof-grid p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.archive-metrics,
.archive-timeline {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.archive-timeline {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.archive-metrics article,
.archive-timeline article {
  border-radius: 10px;
  border: 1px solid var(--color-border-soft);
  background: var(--color-surface-panel);
  padding: 8px 10px;
}

.trust-note {
  display: grid;
  gap: 6px;
}

.archive-metrics p,
.archive-timeline p,
.trust-note p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.archive-metrics strong,
.archive-timeline strong,
.trust-note strong {
  display: block;
  margin-top: 4px;
  font-size: 14px;
  color: var(--color-text-main);
}

.archive-timeline span,
.trust-note span {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-text-sub);
  line-height: 1.55;
}

.archive-actions-row {
  display: flex;
  justify-content: flex-end;
}

.up {
  color: var(--color-rise);
}

.down {
  color: var(--color-fall);
}

.trust-list {
  display: grid;
  gap: 10px;
}

.trust-top {
  display: flex;
  align-items: center;
  gap: 8px;
}

.trust-top strong {
  font-size: 14px;
}

.trust-list p {
  margin: 8px 0 0;
  color: var(--color-text-sub);
  line-height: 1.65;
}

@media (max-width: 1080px) {
  .archive-focus-layout,
  .archive-hero,
  .archive-layout {
    grid-template-columns: 1fr;
  }

  .archive-focus-head {
    display: grid;
  }

  .archive-focus-actions {
    justify-content: flex-start;
  }
}

@media (max-width: 720px) {
  .archive-actions,
  .archive-focus-actions,
  .archive-badges,
  .archive-actions-row {
    width: 100%;
  }

  .archive-item-head,
  .archive-focus-head,
  .section-head {
    flex-direction: column;
  }

  .archive-focus-grid,
  .archive-read-order,
  .archive-metrics,
  .archive-timeline,
  .archive-version-grid,
  .archive-history-list,
  .archive-proof-grid,
  .archive-origin-grid {
    grid-template-columns: 1fr;
  }
}
</style>
