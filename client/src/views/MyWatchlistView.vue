<template>
  <section class="watchlist-page fade-up">
    <header class="watchlist-hero card">
      <div>
        <p class="watchlist-kicker">我的关注</p>
        <h1>查看关注标的的最新变化</h1>
        <p class="watchlist-desc">
          保存推荐股票后，可在这里查看状态、资讯和风险边界变化。
        </p>
        <div class="watchlist-tags">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">本地轻量存储</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">状态变化优先</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">支持快速移除</span>
        </div>
      </div>
      <div class="watchlist-actions">
        <button class="primary-btn finance-primary-btn" type="button" :disabled="loading" @click="loadWatchlist">
          {{ loading ? "同步中..." : "刷新我的关注" }}
        </button>
        <button class="ghost-btn finance-ghost-btn" type="button" @click="goStrategies">去策略中心</button>
      </div>
    </header>

    <section class="watchlist-summary card">
      <article>
        <p>关注股票</p>
        <strong>{{ watchedRows.length }} 只</strong>
      </article>
      <article>
        <p>可同步状态</p>
        <strong>{{ trackedCount }} 只</strong>
      </article>
      <article>
        <p>资讯更新</p>
        <strong>{{ totalNewsCount }} 条</strong>
      </article>
      <article>
        <p>最近刷新</p>
        <strong>{{ lastUpdatedAt || "-" }}</strong>
      </article>
    </section>

    <StatePanel
      :tone="watchlistAccessState.tone"
      :eyebrow="watchlistAccessState.label"
      :title="watchlistAccessState.title"
      :description="watchlistAccessState.desc"
    >
      <template #actions>
        <button type="button" @click="handleWatchlistPrimaryAction">{{ watchlistPrimaryActionText }}</button>
        <button type="button" class="ghost-btn finance-ghost-btn" @click="handleWatchlistSecondaryAction">
          {{ watchlistSecondaryActionText }}
        </button>
      </template>
    </StatePanel>

    <section class="watchlist-focus-layout finance-dual-rail">
      <article class="card watchlist-focus-card finance-section-card">
        <header class="watchlist-focus-head finance-section-head-grid">
          <div>
            <p class="watchlist-kicker">优先跟踪</p>
            <h2 class="section-title">先看变化，再决定下一步操作。</h2>
            <p class="section-subtitle">
              重点查看状态、资讯、风险边界和结论变化。
            </p>
          </div>
          <div class="watchlist-focus-actions finance-action-row">
            <button class="primary-btn finance-primary-btn" type="button" :disabled="loading" @click="loadWatchlist">
              {{ loading ? "同步中..." : "刷新关注列表" }}
            </button>
            <button class="ghost-btn finance-ghost-btn" type="button" @click="goStrategies">回策略中心</button>
          </div>
        </header>

        <div v-if="priorityWatchItem" class="watchlist-priority-panel finance-card-pale">
          <div class="watchlist-priority-head">
            <div>
              <p class="watchlist-date">优先跟踪样本</p>
              <h3>{{ priorityWatchItem.name }}</h3>
            </div>
            <div class="watchlist-priority-badges">
              <span class="finance-pill finance-pill-roomy finance-pill-info">{{ priorityWatchItem.status }}</span>
              <span
                class="finance-pill finance-pill-roomy"
                :class="
                  priorityWatchItem.performanceClass === 'up'
                    ? 'finance-pill-success'
                    : priorityWatchItem.performanceClass === 'down'
                      ? 'finance-pill-danger'
                      : 'finance-pill-neutral'
                "
              >
                {{ priorityWatchItem.performance }}
              </span>
              <span class="finance-pill finance-pill-roomy finance-pill-neutral">{{ priorityWatchItem.risk }}</span>
            </div>
          </div>
          <p class="watchlist-priority-summary">{{ priorityWatchItem.explanationSummary }}</p>
          <div v-if="priorityWatchItem.proofTags.length" class="watchlist-focus-tags">
            <span
              v-for="tag in priorityWatchItem.proofTags"
              :key="`focus-${tag}`"
              class="finance-pill finance-pill-roomy finance-pill-accent"
            >
              {{ tag }}
            </span>
          </div>
          <div class="watchlist-priority-grid">
            <article class="finance-list-card">
              <p>状态变化</p>
              <strong>{{ priorityWatchItem.statusSummary }}</strong>
              <span>{{ priorityWatchItem.statusNote }}</span>
            </article>
            <article class="finance-list-card">
              <p>资讯变化</p>
              <strong>{{ priorityWatchItem.latestNewsTitle }}</strong>
              <span>{{ priorityWatchItem.latestNewsTime }}</span>
            </article>
            <article class="finance-list-card">
              <p>风险边界</p>
              <strong>{{ priorityWatchItem.stopLoss }}</strong>
              <span>{{ priorityWatchItem.takeProfit }}</span>
            </article>
            <article class="finance-list-card">
              <p>下一动作</p>
              <strong>{{ priorityWatchItem.nextAction }}</strong>
              <span>{{ priorityWatchItem.reason }}</span>
            </article>
          </div>
          <div class="watchlist-change-guide-grid">
            <article class="finance-list-card">
              <p>结论变化</p>
              <strong>{{ priorityWatchItem.conclusionChange.title }}</strong>
              <span>{{ priorityWatchItem.conclusionChange.note }}</span>
            </article>
            <article class="finance-list-card">
              <p>风险边界变化</p>
              <strong>{{ priorityWatchItem.riskBoundaryChange.title }}</strong>
              <span>{{ priorityWatchItem.riskBoundaryChange.note }}</span>
            </article>
            <article class="finance-list-card">
              <p>角色分歧变化</p>
              <strong>{{ priorityWatchItem.roleChange.title }}</strong>
              <span>{{ priorityWatchItem.roleChange.note }}</span>
            </article>
          </div>
        </div>
        <StatePanel
          v-else
          tone="info"
          eyebrow="我的关注"
          title="你还没有可跟踪的关注样本"
          description="先在首页或策略页保存标的，这里才会显示跟踪变化。"
          compact
        >
          <template #actions>
            <button type="button" @click="goHome">回首页</button>
            <button type="button" class="ghost-btn finance-ghost-btn" @click="goStrategies">去策略中心</button>
          </template>
        </StatePanel>
      </article>

      <aside class="watchlist-side-stack finance-stack-tight finance-sticky-side">
        <article class="card watchlist-side-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">回访节奏</h2>
              <p class="section-subtitle">按固定顺序查看重点变化。</p>
            </div>
          </header>
          <div class="rhythm-list">
            <article v-for="item in revisitPlan" :key="item.title">
              <p>{{ item.time }}</p>
              <strong>{{ item.title }}</strong>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </article>

        <article class="card watchlist-side-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">功能范围</h2>
              <p class="section-subtitle">不同身份可查看的内容深度不同。</p>
            </div>
          </header>
          <div class="capability-bridge-list">
            <article
              v-for="item in watchCapabilityCards"
              :key="item.key"
              class="finance-list-card"
              :class="{ active: item.active, unlocked: item.unlocked }"
            >
              <div class="capability-bridge-head">
                <p>{{ item.stage }}</p>
                <span>{{ item.badge }}</span>
              </div>
              <strong>{{ item.title }}</strong>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </article>

        <article class="card watchlist-side-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">页面关系</h2>
              <p class="section-subtitle">关注页用于持续跟踪，不替代策略页和档案页。</p>
            </div>
          </header>
          <div class="watchlist-relation-list">
            <article v-for="item in watchPageRelationRows" :key="item.title" class="finance-list-card">
              <strong>{{ item.title }}</strong>
              <p>{{ item.desc }}</p>
            </article>
          </div>
        </article>
      </aside>
    </section>

    <p v-if="errorMessage" class="page-note warning">{{ errorMessage }}</p>
    <p v-else class="page-note">显示你已保存标的的最新变化。</p>

    <section class="watchlist-layout">
      <article class="card watchlist-card">
        <header class="section-head">
          <div>
            <h2 class="section-title">我的关注列表</h2>
            <p class="section-subtitle">优先查看状态、资讯和风险提醒。</p>
          </div>
        </header>

        <div v-if="watchedRows.length" class="watchlist-list">
          <article v-for="item in watchedRows" :key="item.id" class="watchlist-item finance-card-surface">
            <div class="watchlist-item-head">
              <div>
                <p class="watchlist-date">加入关注：{{ item.addedAt }}</p>
                <h3>{{ item.name }}</h3>
              </div>
              <span class="status finance-pill" :class="item.statusClass">{{ item.status }}</span>
            </div>

            <p class="watchlist-reason">{{ item.reason }}</p>

            <div class="watchlist-explanation-box finance-card-pale">
              <div class="watchlist-explanation-head">
                <p>继续跟踪原因</p>
                <span v-if="item.metaText">{{ item.metaText }}</span>
              </div>
              <p class="watchlist-explanation-summary">{{ item.explanationSummary }}</p>
              <div v-if="item.proofTags.length" class="watchlist-proof-tags">
                <span
                  v-for="tag in item.proofTags"
                  :key="`${item.id}-${tag}`"
                  class="finance-pill finance-pill-compact finance-pill-info"
                >
                  {{ tag }}
                </span>
              </div>
              <p v-if="item.consensusText" class="watchlist-consensus">{{ item.consensusText }}</p>
            </div>

            <div class="watchlist-metrics">
              <article class="finance-list-card finance-list-card-panel">
                <p>风险等级</p>
                <strong>{{ item.risk }}</strong>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>下一动作</p>
                <strong>{{ item.nextAction }}</strong>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>资讯变化</p>
                <strong>{{ item.newsCount }} 条</strong>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>累计表现</p>
                <strong :class="item.performanceClass">{{ item.performance }}</strong>
              </article>
            </div>

            <div class="watchlist-update-grid">
              <article class="finance-list-card finance-list-card-panel">
                <p>状态变化</p>
                <strong>{{ item.statusSummary }}</strong>
                <span>{{ item.statusNote }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>风险边界</p>
                <strong>{{ item.stopLoss }}</strong>
                <span>{{ item.takeProfit }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>最近资讯</p>
                <strong>{{ item.latestNewsTitle }}</strong>
                <span>{{ item.latestNewsTime }}</span>
              </article>
            </div>

            <div class="watchlist-change-grid">
              <article class="finance-list-card finance-list-card-panel">
                <p>结论变化</p>
                <strong>{{ item.conclusionChange.title }}</strong>
                <span>{{ item.conclusionChange.note }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>风险边界变化</p>
                <strong>{{ item.riskBoundaryChange.title }}</strong>
                <span>{{ item.riskBoundaryChange.note }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>角色分歧变化</p>
                <strong>{{ item.roleChange.title }}</strong>
                <span>{{ item.roleChange.note }}</span>
              </article>
            </div>

            <div class="watchlist-change-timeline">
              <article
                v-for="node in item.changeTimeline"
                :key="`${item.id}-${node.label}`"
                class="finance-list-card finance-list-card-panel"
              >
                <p>{{ node.label }}</p>
                <strong>{{ node.value }}</strong>
                <span>{{ node.note }}</span>
              </article>
            </div>

            <div v-if="item.versionDiff" class="watchlist-version-diff finance-card-pale">
              <div class="watchlist-version-head">
                <p>版本对照</p>
                <span>{{ item.versionCompareTitle || item.versionDiff.meta }}</span>
              </div>
              <p class="watchlist-version-hint">默认对比加入关注时版本，帮助你判断当前结论是不是变了。</p>
              <div class="watchlist-version-grid">
                <article class="finance-list-card finance-list-card-panel">
                  <p>加入关注时</p>
                  <strong>{{ item.versionDiff.beforeLabel }}</strong>
                  <span>{{ item.versionDiff.beforeNote }}</span>
                </article>
                <article class="finance-list-card finance-list-card-panel">
                  <p>当前 explanation</p>
                  <strong>{{ item.versionDiff.afterLabel }}</strong>
                  <span>{{ item.versionDiff.afterNote }}</span>
                </article>
                <article class="watchlist-version-grid__wide finance-list-card finance-list-card-panel">
                  <p>{{ item.versionDiff.diffLabel }}</p>
                  <strong>{{ item.versionDiff.diffNote }}</strong>
                </article>
              </div>
            </div>

            <div
              v-if="item.versionHistoryItems.length"
              class="watchlist-history-block watchlist-history-block--versions finance-list-card finance-list-card-panel"
            >
              <div class="watchlist-history-head">
                <p>后端版本轨迹</p>
                <span>最近 {{ item.versionHistoryItems.length }} 条</span>
              </div>
              <p class="watchlist-version-hint">这里展示的是后端版本轨迹，不只靠本地关注快照推断。</p>
              <div class="watchlist-history-list">
                <article
                  v-for="history in item.versionHistoryItems"
                  :key="`${item.id}-${history.key}`"
                  class="finance-list-card"
                >
                  <div class="watchlist-history-title">
                    <p>{{ history.title }}</p>
                    <strong>{{ history.version }}</strong>
                  </div>
                  <span>{{ history.note }}</span>
                </article>
              </div>
            </div>

            <div v-if="item.changeHistory.length" class="watchlist-history-block finance-list-card finance-list-card-panel">
              <div class="watchlist-history-head">
                <p>多次变化记录</p>
                <span>不只看加入关注时，而是连续看每次刷新抓到的变化。</span>
              </div>
              <div class="watchlist-history-list">
                <article
                  v-for="entry in item.changeHistory"
                  :key="`${item.id}-${entry.key}`"
                  class="finance-list-card"
                >
                  <div class="watchlist-history-title">
                    <p>{{ entry.time }}</p>
                    <strong>{{ entry.title }}</strong>
                  </div>
                  <span>{{ entry.note }}</span>
                </article>
              </div>
            </div>

            <div v-if="item.seedHighlights.length || item.roleHighlights.length || item.riskHighlights.length" class="watchlist-proof-grid">
              <article v-if="item.seedHighlights.length" class="finance-list-card finance-list-card-panel">
                <p>持续关注的种子</p>
                <div class="watchlist-chip-list">
                  <span
                    v-for="seed in item.seedHighlights"
                    :key="`${item.id}-seed-${seed}`"
                    class="finance-pill finance-pill-compact finance-pill-neutral"
                  >
                    {{ seed }}
                  </span>
                </div>
              </article>
              <article v-if="item.roleHighlights.length" class="finance-list-card finance-list-card-panel">
                <p>角色结论</p>
                <div class="watchlist-chip-list role">
                  <span
                    v-for="role in item.roleHighlights"
                    :key="`${item.id}-role-${role}`"
                    class="finance-pill finance-pill-compact finance-pill-info"
                  >
                    {{ role }}
                  </span>
                </div>
              </article>
              <article v-if="item.riskHighlights.length" class="finance-list-card finance-list-card-panel">
                <p>风险边界变化</p>
                <div class="watchlist-chip-list risk">
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

            <div class="watchlist-actions-row">
              <button type="button" class="ghost-btn finance-ghost-btn" @click="goStrategies">去策略中心</button>
              <button type="button" class="remove-btn" @click="removeItem(item.id)">移出关注</button>
            </div>
          </article>
        </div>
        <StatePanel
          v-else
          tone="info"
          eyebrow="我的关注"
          title="你还没有保存任何关注股票"
          description="可以先去首页或策略中心，把想持续跟踪的股票加入关注。"
          compact
        >
          <template #actions>
            <button type="button" @click="goHome">回首页</button>
            <button type="button" class="ghost-btn finance-ghost-btn" @click="goStrategies">去策略中心</button>
          </template>
        </StatePanel>
      </article>

      <aside class="card rhythm-card">
        <header class="section-head compact">
          <div>
            <h2 class="section-title">变化信息</h2>
            <p class="section-subtitle">按四类变化快速判断今日重点。</p>
          </div>
        </header>
        <div class="watchlist-relation-list watchlist-relation-list--dense">
          <article v-for="item in watchChangeGuideCards" :key="item.title">
            <strong>{{ item.title }}</strong>
            <p>{{ item.desc }}</p>
          </article>
        </div>
        <div class="capability-bridge">
          <p class="capability-bridge-kicker">使用建议</p>
          <div class="rhythm-list">
            <article v-for="item in watchUsageRows" :key="item.title">
              <p>{{ item.time }}</p>
              <strong>{{ item.title }}</strong>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </div>
        <div class="capability-bridge">
          <p class="capability-bridge-kicker">当前页作用</p>
          <div class="rhythm-list">
            <article v-for="item in watchWorkbenchCards" :key="item.title">
              <p>{{ item.label }}</p>
              <strong>{{ item.title }}</strong>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </div>
      </aside>
    </section>
  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import StatePanel from "../components/StatePanel.vue";
import { getMembershipQuota } from "../api/membership";
import { getStockRecommendationInsight, getStockRecommendationVersionHistory } from "../api/market";
import { useClientAuth } from "../lib/client-auth";
import {
  rememberExperimentAttributionSource,
  rememberPendingExperimentJourneySource,
  promotePendingExperimentJourneySources
} from "../lib/growth-analytics";
import { getExperimentVariant } from "../lib/growth-experiments";
import {
  buildStrategyBatchText,
  buildStrategyHistoryCompareState,
  buildStrategyInsightSections,
  buildStrategyMetaText,
  buildStrategyProofTags,
  buildStrategyRiskHighlights,
  buildFallbackStrategyVersionHistory,
  findMatchedStrategyHistoryItem,
  mapStrategyVersionHistory,
  toStrategyTradeDate
} from "../lib/strategy-version";
import { WATCHLIST_EVENT, listWatchedStocks, removeWatchedStock, syncWatchedStockSnapshot } from "../lib/watchlist";

const router = useRouter();
const { isLoggedIn } = useClientAuth();
const watchlistMembershipExperimentVariant = getExperimentVariant("watchlist_membership_entry", ["default"]);
const loading = ref(false);
const errorMessage = ref("");
const lastUpdatedAt = ref("");
const rawWatchedItems = ref([]);
const insightMap = ref({});
const versionHistoryMap = ref({});
const memberStageLoading = ref(false);
const isVIPUser = ref(false);

const watchedRows = computed(() =>
  rawWatchedItems.value
    .map((item) => {
    const insight = insightMap.value[item.id] || {};
    const explanation = insight.explanation || buildFallbackWatchExplanation(item, insight);
    const sections = buildStrategyInsightSections(
      explanation,
      insight.recommendation?.reason_summary || item.reason_summary || "等待同步推荐逻辑。"
    );
    const baseline = item.snapshot || {};
    const detail = insight.detail || {};
    const stats = insight.performance_stats || {};
    const relatedNews = Array.isArray(insight.related_news) ? insight.related_news : [];
    const statusMeta = mapStatus(insight.recommendation?.status || item.status);
    const performance = Number(stats.cumulative_return);
    const latestNews = relatedNews[0] || null;
    const conclusionChange = buildConclusionChange(
      baseline.reason_summary || item.reason_summary,
      sections.whyNow || insight.recommendation?.reason_summary || item.reason_summary
    );
    const riskBoundaryChange = buildRiskBoundaryChange(
      baseline,
      {
        risk_level: insight.recommendation?.risk_level || item.risk_level,
        stop_loss: detail.stop_loss || "",
        take_profit: detail.take_profit || ""
      }
    );
    const roleChange = buildRoleChange(baseline.agent_opinions, explanation.agent_opinions);
    const historySnapshots = buildHistorySnapshots(item);
    const versionHistoryItems = buildWatchVersionHistoryItems(item, explanation, versionHistoryMap.value[item.id]);
    const matchedHistoryItem = findWatchMatchedHistoryItem(item, versionHistoryItems);
    const historyCompare = buildStrategyHistoryCompareState({
      historyItems: versionHistoryItems,
      fallbackItem: matchedHistoryItem,
      explanation,
      selectedTitlePrefix: "当前对比 ",
      formatDateTime,
      fallbackRecordLabel: "watchlist",
      recordVersionFallback: item?.snapshot?.strategy_version || "watchlist",
      recordReasonFallback: item?.snapshot?.reason_summary || item?.reason_summary || "",
      upgradedText: "加入关注时的基线版本和当前 explanation 的结论都发生了变化。",
      reasonChangedText: "当前 explanation 对关注基线给出了新的收敛说明。",
      versionChangedText: "持续跟踪的主线还在，但后端版本已经发生刷新。"
    });

      return {
      id: item.id,
      name: `${item.symbol || "-"} ${item.name || ""}`.trim(),
      addedAt: formatDateTime(item.added_at),
      reason: insight.recommendation?.reason_summary || item.reason_summary || "等待同步推荐逻辑。",
      explanationSummary: sections.whyNow || insight.recommendation?.reason_summary || item.reason_summary || "等待同步推荐逻辑。",
      proofTags: buildWatchProofTags(explanation),
      consensusText: sections.proofSource || "",
      seedHighlights: normalizeTextList(explanation.seed_highlights).slice(0, 3),
      roleHighlights: buildRoleHighlights(explanation.agent_opinions).slice(0, 3),
      riskHighlights: buildStrategyRiskHighlights(explanation, {
        limit: 3,
        fallback: buildStatusNote(statusMeta.label, detail)
      }),
      metaText: buildWatchMetaText(explanation),
      conclusionChange,
      riskBoundaryChange,
      roleChange,
      changeTimeline: buildChangeTimeline(item, statusMeta, conclusionChange, riskBoundaryChange, roleChange, lastUpdatedAt.value),
      changeHistory: buildWatchChangeHistory(historySnapshots),
      versionHistoryItems: versionHistoryItems.slice(0, 3),
      versionDiff: historyCompare.diff,
      versionCompareTitle: historyCompare.selectedTitle || "当前对比 加入关注时版本",
      risk: mapRisk(insight.recommendation?.risk_level || item.risk_level),
      status: statusMeta.label,
      statusClass: statusMeta.className,
      statusSummary: buildStatusSummary(statusMeta.label, performance),
      statusNote: buildStatusNote(statusMeta.label, detail),
      nextAction: buildNextAction(statusMeta.label),
      newsCount: `${relatedNews.length}`,
      performance: formatPercent(performance),
      performanceClass: trendClassByNumber(performance),
      stopLoss: detail.stop_loss || "止损线待补充",
      takeProfit: detail.take_profit || "止盈计划待补充",
      latestNewsTitle: latestNews?.title || "暂无新增资讯",
      latestNewsTime: formatDateTime(latestNews?.published_at || latestNews?.time || latestNews?.created_at)
      };
    })
    .sort((a, b) => compareWatchPriority(a, b))
);

const trackedCount = computed(() => watchedRows.value.filter((item) => item.status !== "待同步").length);
const totalNewsCount = computed(() => watchedRows.value.reduce((sum, item) => sum + Number(item.newsCount || 0), 0));
const priorityWatchItem = computed(() => watchedRows.value[0] || null);

const revisitPlan = computed(() => [
  { time: "09:30", title: "先看状态变化", desc: "确认你的关注股票有没有进入跟踪、止盈、止损或失效阶段。" },
  { time: "11:30", title: "再看资讯变化", desc: "优先看关注股票有没有新的资讯支撑或风险提示。" },
  { time: "15:00", title: "最后看风险边界", desc: "对照止盈止损与失效条件，决定明天是否继续跟踪。" }
]);
const watchPageRelationRows = computed(() => [
  {
    title: "首页与策略页",
    desc: "先把想跟踪的标的加入关注，再回这里查看变化。"
  },
  {
    title: "历史档案页",
    desc: "进入止盈、止损或失效阶段后，可去档案页查看完整复盘。"
  },
  {
    title: "资讯页",
    desc: "需要进一步确认资讯变化时，可去资讯页查看正文和附件。"
  }
]);
const watchChangeGuideCards = computed(() => [
  {
    title: "状态变化",
    desc: "先确认是否进入跟踪、止盈、止损或失效阶段。"
  },
  {
    title: "资讯变化",
    desc: "再看是否有新的资讯支撑或风险提示。"
  },
  {
    title: "风险边界变化",
    desc: "查看止盈、止损和失效条件是否变化。"
  },
  {
    title: "结论变化",
    desc: "最后看推荐结论是否变化，再决定是否回策略页查看。"
  }
]);
const watchUsageRows = computed(() => [
  { time: "先做", title: "先找高优先样本", desc: "优先处理最需要动作或变化最明显的标的。" },
  { time: "再做", title: "再看风险边界", desc: "确认止盈止损和失效条件是否变化。" },
  { time: "最后", title: "最后决定去向", desc: "需要更详细内容时，再前往策略页、档案页或资讯页。" }
]);
const watchWorkbenchCards = computed(() => [
  {
    label: "变化",
    title: "优先显示重点标的",
    desc: "帮助你快速找到今天最需要处理的标的。"
  },
  {
    label: "节奏",
    title: "固定查看顺序",
    desc: "按盘中和收盘节奏查看变化。"
  },
  {
    label: "跳转",
    title: "继续查看相关页面",
    desc: "发现重要变化后，可继续前往策略、档案或资讯页。"
  }
]);
const watchlistAccessStage = computed(() => {
  if (isVIPUser.value) {
    return "VIP";
  }
  if (isLoggedIn.value) {
    return "REGISTERED";
  }
  return "VISITOR";
});
const watchlistAccessState = computed(() => {
  if (memberStageLoading.value) {
    return {
      tone: "info",
      label: "识别中",
      title: "正在确认你的关注页阶段",
      desc: "确认完成后，会把关注页 CTA 调整到对应的登录 / 会员状态。"
    };
  }
  if (watchlistAccessStage.value === "VIP") {
    return {
      tone: "success",
      label: "会员阶段",
      title: "你已可结合关注列表查看更完整的变化信息。",
      desc: "建议每天先看关注页，再去资讯中心确认变化。"
    };
  }
  if (watchlistAccessStage.value === "REGISTERED") {
    return {
      tone: "warning",
      label: "注册阶段",
      title: "你已形成自己的关注清单，可继续保持回访习惯。",
      desc: "先让关注页成为固定入口，再决定是否需要更多盘中变化和复盘内容。"
    };
  }
  return {
    tone: "info",
    label: "游客阶段",
    title: "可先用本地关注保存标的。",
    desc: "登录后可继续配合策略页和会员服务使用。"
  };
});
const watchlistPrimaryActionText = computed(() => {
  if (watchlistAccessStage.value === "VIP") {
    return "去会员中心";
  }
  if (watchlistAccessStage.value === "REGISTERED") {
    return "升级会员";
  }
  return "登录继续追踪";
});
const watchlistSecondaryActionText = computed(() => {
  if (watchlistAccessStage.value === "VIP") {
    return "去资讯中心";
  }
  if (watchlistAccessStage.value === "REGISTERED") {
    return "去资讯中心";
  }
  return "去策略中心";
});
const watchCapabilityCards = computed(() => [
  {
    key: "visitor",
    stage: "游客",
    badge: watchlistAccessStage.value === "VISITOR" ? "当前阶段" : "基础层",
    title: "先保存重点标的",
    desc: "先从关注列表开始，方便后续回看。",
    unlocked: true,
    active: watchlistAccessStage.value === "VISITOR"
  },
  {
    key: "registered",
    stage: "注册",
    badge: watchlistAccessStage.value === "REGISTERED" ? "当前阶段" : "已解锁",
    title: "开始看持续跟踪和风险边界",
    desc: "注册后可把关注、资讯和历史样本一起查看。",
    unlocked: watchlistAccessStage.value !== "VISITOR",
    active: watchlistAccessStage.value === "REGISTERED"
  },
  {
    key: "vip",
    stage: "会员",
    badge: watchlistAccessStage.value === "VIP" ? "当前阶段" : "升级后更完整",
    title: "查看更完整的跟踪与复盘信息",
    desc: "会员可查看更完整的解释、变化和复盘内容。",
    unlocked: watchlistAccessStage.value === "VIP",
    active: watchlistAccessStage.value === "VIP"
  }
]);

async function loadWatchlist() {
  loading.value = true;
  errorMessage.value = "";
  rawWatchedItems.value = listWatchedStocks();
  if (rawWatchedItems.value.length === 0) {
    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
    loading.value = false;
    return;
  }

  const nextInsightMap = {};
  const nextVersionHistoryMap = {};
  const errors = [];
  let didSyncHistory = false;
  await Promise.all(
    rawWatchedItems.value.map(async (item) => {
      const [insightResult, versionHistoryResult] = await Promise.allSettled([
        getStockRecommendationInsight(item.id),
        getStockRecommendationVersionHistory(item.id)
      ]);

      if (insightResult.status === "fulfilled" && insightResult.value) {
        nextInsightMap[item.id] = insightResult.value;
        didSyncHistory =
          syncWatchedStockSnapshot(item.id, buildWatchHistorySnapshot(item, insightResult.value), { emit: false }) ||
          didSyncHistory;
      } else if (insightResult.status === "rejected") {
        errors.push(`${item.symbol || item.id} 同步失败：${parseErrorMessage(insightResult.reason)}`);
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
  if (didSyncHistory) {
    rawWatchedItems.value = listWatchedStocks();
  }
  errorMessage.value = errors.join("；");
  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  loading.value = false;
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
    promoteWatchlistPostAuthAttribution();
  }
}

function handleWatchlistUpdated() {
  loadWatchlist();
}

function removeItem(id) {
  removeWatchedStock(id);
}

function goHome() {
  router.push("/home");
}

function goStrategies() {
  router.push("/strategies");
}

function goNews() {
  router.push("/news");
}

function rememberWatchlistMembershipEntry(targetKey, metadata = {}) {
  rememberExperimentAttributionSource({
    experimentKey: "watchlist_membership_entry",
    variantKey: watchlistMembershipExperimentVariant,
    pageKey: "watchlist",
    targetKey,
    userStage: watchlistAccessStage.value,
    metadata: {
      watched_count: watchedRows.value.length,
      tracked_count: trackedCount.value,
      ...metadata
    }
  });
}

function rememberWatchlistPendingMembershipEntry(targetKey, metadata = {}) {
  rememberPendingExperimentJourneySource({
    experimentKey: "watchlist_membership_entry",
    variantKey: watchlistMembershipExperimentVariant,
    pageKey: "watchlist",
    targetKey,
    userStage: "VISITOR",
    redirectPath: "/watchlist",
    metadata: {
      watched_count: watchedRows.value.length,
      tracked_count: trackedCount.value,
      ...metadata
    }
  });
}

function promoteWatchlistPostAuthAttribution() {
  if (!isLoggedIn.value) {
    return;
  }
  promotePendingExperimentJourneySources({
    experimentKey: "watchlist_membership_entry",
    pageKey: "watchlist",
    userStage: watchlistAccessStage.value,
    metadata: {
      watched_count: watchedRows.value.length,
      tracked_count: trackedCount.value
    }
  });
}

function handleWatchlistPrimaryAction() {
  if (watchlistAccessStage.value === "VIP") {
    router.push("/membership");
    return;
  }
  if (watchlistAccessStage.value === "REGISTERED") {
    rememberWatchlistMembershipEntry("primary_action", {
      destination: "/membership"
    });
    router.push("/membership");
    return;
  }
  rememberWatchlistPendingMembershipEntry("primary_action", {
    destination_after_auth: "/watchlist"
  });
  router.push({ path: "/auth", query: { redirect: "/watchlist" } });
}

function handleWatchlistSecondaryAction() {
  if (watchlistAccessStage.value === "VISITOR") {
    goStrategies();
    return;
  }
  goNews();
}

function compareWatchPriority(a, b) {
  const statusDiff = watchPriorityByLabel(b?.status) - watchPriorityByLabel(a?.status);
  if (statusDiff !== 0) {
    return statusDiff;
  }
  const newsDiff = Number(b?.newsCount || 0) - Number(a?.newsCount || 0);
  if (newsDiff !== 0) {
    return newsDiff;
  }
  return String(a?.addedAt || "").localeCompare(String(b?.addedAt || ""));
}

function watchPriorityByLabel(label) {
  const text = String(label || "");
  if (text === "已止损" || text === "已失效") {
    return 5;
  }
  if (text === "已止盈") {
    return 4;
  }
  if (text === "跟踪中" || text === "已发布") {
    return 3;
  }
  if (text === "已复盘") {
    return 2;
  }
  if (text === "待同步") {
    return 0;
  }
  return 1;
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
  if (!source) return { label: "待同步", className: "watch" };
  if (source === "PUBLISHED" || source === "ACTIVE") return { label: "已发布", className: "good" };
  if (source === "TRACKING") return { label: "跟踪中", className: "good" };
  if (source === "HIT_TAKE_PROFIT") return { label: "已止盈", className: "good" };
  if (source === "HIT_STOP_LOSS") return { label: "已止损", className: "risky" };
  if (source === "INVALIDATED") return { label: "已失效", className: "risky" };
  if (source === "REVIEWED") return { label: "已复盘", className: "normal" };
  return { label: source, className: "watch" };
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

function buildStatusSummary(label, performance) {
  if (label === "已止盈") {
    return "这条关注已经进入兑现阶段。";
  }
  if (label === "已止损") {
    return "这条关注已经触发风险边界。";
  }
  if (label === "已失效") {
    return "原始逻辑已不适合继续沿用。";
  }
  if (Number.isFinite(performance)) {
    return `当前累计表现 ${formatPercent(performance)}，可继续观察。`;
  }
  return "等待更多状态数据同步。";
}

function buildStatusNote(label, detail) {
  if (label === "已止盈") {
    return detail.take_profit || "建议回看止盈过程与后续节奏。";
  }
  if (label === "已止损") {
    return detail.stop_loss || "建议回看止损纪律是否执行到位。";
  }
  if (label === "已失效") {
    return detail.risk_note || "建议关注失效原因，而不是继续找理由。";
  }
  return detail.risk_note || "继续跟踪风险边界和资讯变化。";
}

function buildNextAction(label) {
  if (label === "已止盈") return "整理复盘";
  if (label === "已止损") return "暂停跟踪";
  if (label === "已失效") return "等待新逻辑";
  return "继续观察";
}

function buildFallbackWatchExplanation(item, insight) {
  const recommendation = insight?.recommendation || {};
  const detail = insight?.detail || {};
  const statusLabel = mapStatus(recommendation.status || item?.status).label;
  const riskNote = String(detail.risk_note || "").trim();
  return {
    seed_highlights: [
      recommendation.reason_summary || item?.reason_summary || "推荐逻辑待同步",
      riskNote || `${mapRisk(recommendation.risk_level || item?.risk_level)}边界`,
      "持续跟踪资讯变化"
    ].filter(Boolean),
    graph_summary: "系统会把推荐逻辑、资讯变化和风险边界串成一条持续跟踪链路。",
    consensus_summary: "继续跟踪不是重复看结果，而是反复验证原始逻辑是否还成立。",
    risk_flags: [buildStatusNote(statusLabel, detail)],
    invalidations: [detail.stop_loss || "跌破预设风险线或核心逻辑被证伪时停止跟踪"],
    confidence_reason: recommendation.reason_summary || item?.reason_summary || "等待同步推荐逻辑。",
    workload_summary: {
      seed_count: 12,
      agent_count: 5,
      scenario_count: 4
    },
    agent_opinions: [
      { agent: "趋势", stance: "继续观察" },
      { agent: "风险", stance: statusLabel === "已止损" || statusLabel === "已失效" ? "暂停跟踪" : "严守边界" }
    ],
    strategy_version: recommendation.strategy_version || "strategy-engine-followup",
    generated_at: recommendation.valid_from || item?.added_at || ""
  };
}

function buildWatchProofTags(explanation) {
  return buildStrategyProofTags(explanation, { includeSeedCount: true });
}

function buildRoleHighlights(agentOpinions) {
  if (!Array.isArray(agentOpinions)) {
    return [];
  }
  return agentOpinions
    .map((item) => {
      const agent = String(item?.agent || "").trim();
      const stance = String(item?.stance || "").trim();
      if (!agent || !stance) {
        return "";
      }
      return `${agent}：${stance}`;
    })
    .filter(Boolean);
}

function buildWatchMetaText(explanation) {
  return buildStrategyMetaText(explanation, formatDateTime, { includeBatch: true, includeJob: false });
}

function buildWatchHistorySnapshot(item, insight) {
  const recommendation = insight?.recommendation || {};
  const detail = insight?.detail || {};
  const explanation = insight?.explanation || buildFallbackWatchExplanation(item, insight);
  const sections = buildStrategyInsightSections(
    explanation,
    recommendation.reason_summary || item?.reason_summary || ""
  );
  return {
    reason_summary: sections.whyNow || recommendation.reason_summary || item?.reason_summary || "",
    risk_level: recommendation.risk_level || item?.risk_level || "",
    status: recommendation.status || item?.status || "",
    take_profit: detail.take_profit || "",
    stop_loss: detail.stop_loss || "",
    strategy_version: explanation.strategy_version || "",
    publish_id: explanation.publish_id || "",
    job_id: explanation.job_id || "",
    trade_date: explanation.trade_date || "",
    publish_version: explanation.publish_version || 0,
    agent_opinions: explanation.agent_opinions || []
  };
}

function normalizeTextList(items) {
  if (!Array.isArray(items)) {
    return [];
  }
  return items
    .map((item) => String(item || "").trim())
    .filter(Boolean);
}

function buildConclusionChange(previousReason, currentReason) {
  const prev = String(previousReason || "").trim();
  const next = String(currentReason || "").trim();
  if (!prev && !next) {
    return {
      title: "待同步",
      note: "加入关注时和当前都缺少足够的结论摘要。"
    };
  }
  if (!prev || prev === next) {
    return {
      title: "基本一致",
      note: next || prev || "当前结论与加入关注时没有明显变化。"
    };
  }
  return {
    title: "已变化",
    note: `加入时：${truncateText(prev, 28)}；当前：${truncateText(next, 28)}`
  };
}

function buildRiskBoundaryChange(previous, current) {
  const previousRisk = mapRisk(previous?.risk_level);
  const currentRisk = mapRisk(current?.risk_level);
  const previousStop = String(previous?.stop_loss || "").trim();
  const currentStop = String(current?.stop_loss || "").trim();

  if ((!previousStop && !currentStop) && previousRisk === currentRisk) {
    return {
      title: "待补更多边界",
      note: `当前风险等级 ${currentRisk || "待确认"}`
    };
  }
  if (previousRisk === currentRisk && previousStop === currentStop) {
    return {
      title: "边界稳定",
      note: currentStop ? `止损线维持 ${currentStop}` : `风险等级维持 ${currentRisk}`
    };
  }
  return {
    title: "已有调整",
    note: `加入时 ${previousRisk}${previousStop ? ` / ${previousStop}` : ""}；当前 ${currentRisk}${
      currentStop ? ` / ${currentStop}` : ""
    }`
  };
}

function buildRoleChange(previousAgents, currentAgents) {
  const before = summarizeAgentOpinions(previousAgents);
  const after = summarizeAgentOpinions(currentAgents);
  if (!before && !after) {
    return {
      title: "待同步",
      note: "当前还没有足够的角色结论用于比较。"
    };
  }
  if (!before) {
    return {
      title: "新增角色判断",
      note: after
    };
  }
  if (before === after) {
    return {
      title: "分歧稳定",
      note: after
    };
  }
  return {
    title: "角色变化",
    note: `加入时：${truncateText(before, 24)}；当前：${truncateText(after, 24)}`
  };
}

function buildChangeTimeline(item, statusMeta, conclusionChange, riskBoundaryChange, roleChange, refreshedAt) {
  const baseline = item?.snapshot || {};
  const baselineReason = baseline.reason_summary || item?.reason_summary || "加入关注时未记录更多结论";
  const baselineRisk = mapRisk(baseline.risk_level || item?.risk_level);
  const refreshNote = [conclusionChange.title, riskBoundaryChange.title, roleChange.title]
    .filter((value) => value && value !== "基本一致" && value !== "边界稳定" && value !== "分歧稳定")
    .join(" / ");

  return [
    {
      label: "加入关注",
      value: formatDateTime(item?.added_at),
      note: `${truncateText(baselineReason, 26)} · ${baselineRisk}`
    },
    {
      label: "最近刷新",
      value: refreshedAt || "待刷新",
      note: refreshNote || "本次刷新未发现明显变化"
    },
    {
      label: "当前判断",
      value: statusMeta.label,
      note: `下一动作：${buildNextAction(statusMeta.label)}`
    }
  ];
}

function buildHistorySnapshots(item) {
  const baseline = item?.snapshot
    ? {
        ...item.snapshot,
        recorded_at: item.snapshot.recorded_at || item?.added_at || ""
      }
    : null;
  const history = Array.isArray(item?.history) ? item.history : [];
  const snapshots = history.length ? history : baseline ? [baseline] : [];
  return snapshots
    .map((entry) => ({
      reason_summary: String(entry?.reason_summary || "").trim(),
      risk_level: String(entry?.risk_level || "").trim(),
      status: String(entry?.status || "").trim(),
      take_profit: String(entry?.take_profit || "").trim(),
      stop_loss: String(entry?.stop_loss || "").trim(),
      strategy_version: String(entry?.strategy_version || "").trim(),
      publish_id: String(entry?.publish_id || "").trim(),
      job_id: String(entry?.job_id || "").trim(),
      trade_date: String(entry?.trade_date || "").trim(),
      publish_version: Number(entry?.publish_version) > 0 ? Number(entry.publish_version) : 0,
      agent_opinions: Array.isArray(entry?.agent_opinions) ? entry.agent_opinions : [],
      recorded_at: entry?.recorded_at || ""
    }))
    .filter(
      (entry) =>
        entry.reason_summary ||
        entry.risk_level ||
        entry.status ||
        entry.strategy_version ||
        entry.publish_id ||
        entry.publish_version ||
        entry.recorded_at
    );
}

function buildWatchChangeHistory(historySnapshots) {
  if (!Array.isArray(historySnapshots) || historySnapshots.length === 0) {
    return [];
  }
  return historySnapshots
    .map((entry, index) => {
      const previous = index > 0 ? historySnapshots[index - 1] : null;
      return {
        key: `${entry.recorded_at || index}-${entry.strategy_version || index}`,
        time: formatDateTime(entry.recorded_at),
        title: buildHistoryTitle(previous, entry),
        note: buildHistoryNote(previous, entry)
      };
    })
    .reverse()
    .slice(0, 5);
}

function buildWatchVersionHistoryItems(item, explanation, historyItems) {
  const normalized = mapStrategyVersionHistory(historyItems, formatDateTime);
  if (normalized.length > 0) {
    return normalized;
  }
  return buildFallbackStrategyVersionHistory(explanation, {
    reasonSummary: item?.reason_summary || "",
    strategyVersion: explanation?.strategy_version || item?.strategy_version || "watchlist",
    tradeDate: explanation?.trade_date || toStrategyTradeDate(item?.valid_from),
    formatDateTime
  });
}

function findWatchMatchedHistoryItem(item, historyItems) {
  const baseline = item?.snapshot || {};
  return findMatchedStrategyHistoryItem(historyItems, {
    tradeDate: baseline.trade_date || item?.valid_from || item?.added_at,
    publishVersion: baseline.publish_version || item?.publish_version
  });
}

function buildHistoryTitle(previous, current) {
  const currentStatus = mapStatus(current?.status).label;
  if (!previous) {
    return "建立初始关注基线";
  }
  if (Number(current?.publish_version) > 0 && Number(previous?.publish_version) !== Number(current?.publish_version)) {
    return `同步到后端批次 V${current.publish_version}`;
  }
  if ((previous?.strategy_version || "") !== (current?.strategy_version || "") && current?.strategy_version) {
    return `切换到 ${current.strategy_version}`;
  }
  if ((previous?.status || "") !== (current?.status || "")) {
    return `状态变为${currentStatus}`;
  }
  if ((previous?.risk_level || "") !== (current?.risk_level || "") || (previous?.stop_loss || "") !== (current?.stop_loss || "")) {
    return "风险边界出现调整";
  }
  if ((previous?.reason_summary || "") !== (current?.reason_summary || "")) {
    return "结论摘要发生变化";
  }
  return "同步检查：结论延续";
}

function buildHistoryNote(previous, current) {
  const parts = [];
  const batchText = buildBatchText(current);
  if (batchText) {
    parts.push(batchText);
  }
  if (!previous) {
    parts.push(truncateText(current?.reason_summary || "加入关注时记录初始判断。", 34));
  } else {
    const conclusion = buildConclusionChange(previous?.reason_summary, current?.reason_summary);
    if (conclusion.title !== "基本一致") {
      parts.push(conclusion.note);
    }
    const risk = buildRiskBoundaryChange(previous, current);
    if (risk.title !== "边界稳定") {
      parts.push(risk.note);
    }
    const role = buildRoleChange(previous?.agent_opinions, current?.agent_opinions);
    if (role.title !== "分歧稳定") {
      parts.push(role.note);
    }
  }

  const extras = [];
  const riskLabel = mapRisk(current?.risk_level);
  if (riskLabel && riskLabel !== "-") {
    extras.push(riskLabel);
  }
  if (current?.stop_loss) {
    extras.push(`止损 ${current.stop_loss}`);
  }
  if (current?.take_profit) {
    extras.push(`止盈 ${current.take_profit}`);
  }
  if (extras.length) {
    parts.push(extras.join(" · "));
  }
  return parts.filter(Boolean).slice(0, 2).join("；") || "本次刷新主要用于确认结论是否延续。";
}

function buildBatchText(snapshot) {
  return buildStrategyBatchText(snapshot, { includeJob: true });
}

function summarizeAgentOpinions(agentOpinions) {
  if (!Array.isArray(agentOpinions)) {
    return "";
  }
  return agentOpinions
    .map((item) => {
      const agent = String(item?.agent || "").trim();
      const stance = String(item?.stance || "").trim();
      if (!agent || !stance) {
        return "";
      }
      return `${agent}:${stance}`;
    })
    .filter(Boolean)
    .slice(0, 2)
    .join(" / ");
}

function truncateText(value, maxLength = 30) {
  const text = String(value || "").trim();
  if (!text) {
    return "";
  }
  if (text.length <= maxLength) {
    return text;
  }
  return `${text.slice(0, maxLength)}...`;
}

function formatPercent(value) {
  if (!Number.isFinite(value)) return "-";
  const percent = (value * 100).toFixed(2);
  if (Number(percent) > 0) return `+${percent}%`;
  return `${percent}%`;
}

function trendClassByNumber(value) {
  if (!Number.isFinite(value)) return "";
  if (value > 0) return "up";
  if (value < 0) return "down";
  return "";
}

function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) return "-";
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

function parseErrorMessage(error) {
  if (!error) return "数据请求失败";
  if (typeof error === "string") return error;
  return error?.message || error?.response?.data?.message || "数据请求失败";
}

onMounted(() => {
  loadMembershipStage();
  loadWatchlist();
  if (typeof window !== "undefined") {
    window.addEventListener(WATCHLIST_EVENT, handleWatchlistUpdated);
  }
});

watch(
  () => isLoggedIn.value,
  () => {
    loadMembershipStage();
  }
);

onBeforeUnmount(() => {
  if (typeof window !== "undefined") {
    window.removeEventListener(WATCHLIST_EVENT, handleWatchlistUpdated);
  }
});
</script>

<style scoped>
.watchlist-page {
  display: grid;
  gap: 14px;
}

.watchlist-hero,
.watchlist-summary,
.watchlist-card,
.rhythm-card {
  padding: 14px;
}

.watchlist-focus-layout {
  --finance-main-column: minmax(0, 1.2fr);
  --finance-side-column: 340px;
}

.watchlist-priority-panel {
  margin-top: 12px;
  padding: 14px;
  display: grid;
  gap: 12px;
}

.watchlist-priority-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.watchlist-priority-head h3 {
  margin: 4px 0 0;
  font-size: 22px;
  line-height: 1.3;
}

.watchlist-priority-badges,
.watchlist-focus-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.watchlist-priority-summary {
  margin: 0;
  font-size: 13px;
  line-height: 1.72;
  color: var(--color-text-sub);
}

.watchlist-priority-grid,
.watchlist-change-guide-grid {
  display: grid;
  gap: 8px;
}

.watchlist-priority-grid {
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.watchlist-change-guide-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.watchlist-priority-grid article,
.watchlist-change-guide-grid article,
.watchlist-relation-list article {
  display: grid;
  gap: 4px;
}

.watchlist-priority-grid p,
.watchlist-change-guide-grid p,
.watchlist-relation-list p {
  margin: 0;
  font-size: 12px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.watchlist-priority-grid strong,
.watchlist-change-guide-grid strong,
.watchlist-relation-list strong {
  font-size: 14px;
  color: var(--color-text-main);
}

.watchlist-priority-grid span,
.watchlist-change-guide-grid span {
  font-size: 12px;
  line-height: 1.6;
  color: var(--color-text-sub);
}

.watchlist-side-stack {
  align-content: start;
}

.watchlist-relation-list {
  display: grid;
  gap: 8px;
}

.watchlist-relation-list--dense strong {
  margin-bottom: 2px;
}

.watchlist-hero {
  display: grid;
  gap: 14px;
  grid-template-columns: 1.1fr 0.9fr;
  background:
    radial-gradient(circle at 0% 0%, var(--color-focus-glow) 0%, transparent 34%),
    radial-gradient(circle at 100% 100%, var(--color-line-gold) 0%, transparent 36%),
    rgba(255, 255, 255, 0.94);
}

.watchlist-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.watchlist-hero h1 {
  margin: 8px 0 10px;
  font-family: var(--font-serif);
  font-size: clamp(28px, 3.4vw, 42px);
  line-height: 1.2;
}

.watchlist-desc,
.page-note,
.watchlist-reason {
  margin: 0;
  color: var(--color-text-sub);
  line-height: 1.72;
}

.watchlist-explanation-box {
  display: grid;
  gap: 8px;
}

.watchlist-explanation-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.watchlist-explanation-head p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.watchlist-explanation-head span {
  font-size: 11px;
  color: var(--color-text-sub);
}

.watchlist-explanation-summary,
.watchlist-consensus {
  margin: 0;
  font-size: 13px;
  line-height: 1.68;
  color: var(--color-text-sub);
}

.watchlist-history-block {
  display: grid;
  gap: 10px;
}

.watchlist-history-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.watchlist-history-head p,
.watchlist-history-title p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.watchlist-history-head span {
  font-size: 11px;
  color: var(--color-text-sub);
}

.watchlist-history-list {
  display: grid;
  gap: 8px;
}

.watchlist-history-list article {
  display: grid;
  gap: 6px;
}

.watchlist-history-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.watchlist-history-title strong {
  font-size: 13px;
  color: var(--color-text-main);
}

.watchlist-history-list article > span {
  font-size: 12px;
  line-height: 1.68;
  color: var(--color-text-sub);
}

.watchlist-version-diff {
  display: grid;
  gap: 10px;
}

.watchlist-version-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.watchlist-version-head p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.watchlist-version-head span {
  font-size: 11px;
  color: var(--color-text-sub);
}

.watchlist-version-hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.68;
  color: var(--color-text-sub);
}

.watchlist-version-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.watchlist-version-grid article {
  display: grid;
  gap: 6px;
}

.watchlist-version-grid article p,
.watchlist-version-grid article strong,
.watchlist-version-grid article span {
  margin: 0;
}

.watchlist-version-grid article p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.watchlist-version-grid article strong {
  font-size: 13px;
  color: var(--color-text-main);
}

.watchlist-version-grid article span {
  font-size: 12px;
  line-height: 1.68;
  color: var(--color-text-sub);
}

.watchlist-version-grid__wide {
  grid-column: 1 / -1;
}

.watchlist-history-block--versions {
  background: var(--color-surface-panel);
}

.watchlist-proof-tags,
.watchlist-chip-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.watchlist-proof-tags span,
.watchlist-chip-list span {
  font-weight: 700;
}

.watchlist-proof-tags span {
  border: 1px solid var(--color-border-soft-strong);
  background: var(--color-surface-panel-soft);
  color: var(--color-pine-700);
}

.page-note.warning {
  color: var(--color-warning);
}

.watchlist-tags {
  margin-top: 14px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.status {
  font-weight: 700;
}

.watchlist-actions {
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 10px;
}

.remove-btn {
  border: 0;
  border-radius: 12px;
  padding: 10px 14px;
  font-weight: 600;
  cursor: pointer;
}

.remove-btn {
  color: var(--color-fall);
  background: var(--color-fall-soft);
}

.watchlist-summary {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.watchlist-summary article,
.watchlist-metrics article,
.watchlist-update-grid article,
.watchlist-change-grid article,
.watchlist-change-timeline article,
.rhythm-list article {
  border-radius: 10px;
  border: 1px solid var(--color-border-soft);
  background: var(--color-surface-panel);
  padding: 8px 10px;
}

.watchlist-summary p,
.watchlist-metrics p,
.watchlist-update-grid p,
.watchlist-change-grid p,
.watchlist-change-timeline p,
.rhythm-list p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.watchlist-summary strong,
.watchlist-metrics strong,
.watchlist-update-grid strong,
.watchlist-change-grid strong,
.watchlist-change-timeline strong,
.rhythm-list strong {
  display: block;
  margin-top: 4px;
  font-size: 14px;
  color: var(--color-text-main);
}

.watchlist-update-grid span,
.watchlist-change-grid span,
.watchlist-change-timeline span,
.rhythm-list span {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: var(--color-text-sub);
  line-height: 1.55;
}

.watchlist-layout {
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

.watchlist-list,
.rhythm-list {
  margin-top: 12px;
  display: grid;
  gap: 12px;
}

.watchlist-item {
  display: grid;
  gap: 10px;
}

.watchlist-item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.watchlist-date {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.watchlist-item-head h3 {
  margin: 5px 0 0;
  font-size: 20px;
  line-height: 1.3;
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

.watchlist-metrics,
.watchlist-update-grid,
.watchlist-change-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.watchlist-update-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.watchlist-change-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.watchlist-change-timeline {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.watchlist-proof-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.watchlist-proof-grid article {
  display: grid;
  gap: 8px;
}

.watchlist-proof-grid p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.watchlist-actions-row {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.capability-bridge {
  margin-top: 12px;
  display: grid;
  gap: 8px;
}

.capability-bridge-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.capability-bridge-list {
  display: grid;
  gap: 8px;
}

.capability-bridge-list article {
  display: grid;
  gap: 6px;
  opacity: 0.88;
}

.capability-bridge-list article.unlocked {
  opacity: 1;
}

.capability-bridge-list article.active {
  border-color: var(--color-border-focus);
  background: var(--gradient-card-active);
  box-shadow: 0 12px 24px rgba(16, 42, 86, 0.08);
}

.capability-bridge-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.capability-bridge-head p,
.capability-bridge-list strong,
.capability-bridge-list span {
  margin: 0;
}

.capability-bridge-head p {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-pine-700);
}

.capability-bridge-head span {
  border-radius: 999px;
  padding: 4px 9px;
  background: var(--color-surface-panel-soft-strong);
  font-size: 11px;
  color: var(--color-text-sub);
}

.capability-bridge-list strong {
  font-size: 14px;
  color: var(--color-text-main);
}

.capability-bridge-list span {
  font-size: 12px;
  line-height: 1.58;
  color: var(--color-text-sub);
}

.up {
  color: var(--color-rise);
}

.down {
  color: var(--color-fall);
}

@media (max-width: 1080px) {
  .watchlist-focus-layout,
  .watchlist-hero,
  .watchlist-layout {
    grid-template-columns: 1fr;
  }

  .watchlist-focus-head,
  .watchlist-priority-head {
    display: grid;
    grid-template-columns: 1fr;
  }

  .watchlist-focus-actions {
    justify-content: flex-start;
  }

  .watchlist-side-stack {
    position: static;
  }
}

@media (max-width: 720px) {
  .watchlist-priority-badges,
  .watchlist-summary,
  .watchlist-priority-grid,
  .watchlist-change-guide-grid,
  .watchlist-metrics,
  .watchlist-update-grid,
  .watchlist-change-grid,
  .watchlist-change-timeline,
  .watchlist-proof-grid,
  .watchlist-version-grid {
    grid-template-columns: 1fr;
  }

  .watchlist-priority-badges {
    overflow-x: auto;
    scrollbar-width: none;
  }

  .watchlist-priority-badges::-webkit-scrollbar {
    display: none;
  }

  .watchlist-item-head,
  .watchlist-focus-head,
  .watchlist-priority-head,
  .section-head,
  .watchlist-actions-row,
  .watchlist-version-head {
    flex-direction: column;
  }
}
</style>
