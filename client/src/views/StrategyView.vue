<template>
  <section class="strategy-page fade-up">
    <header class="strategy-hero card">
      <div class="strategy-hero-copy finance-copy-stack">
        <div class="finance-pill-row">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">策略页</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">主推荐优先</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">风险边界前置</span>
        </div>
        <div>
          <p class="hero-kicker">策略中心</p>
          <h1 class="section-title">查看今日股票与期货策略</h1>
          <p class="section-subtitle">
            优先展示今日主推荐、风险边界和执行参数。
          </p>
          <div class="tag-row">
            <span v-for="tag in tags" :key="tag" class="finance-pill finance-pill-compact finance-pill-neutral">{{ tag }}</span>
          </div>
          <p v-if="loading" class="api-state">正在同步策略数据...</p>
          <p v-else-if="errorMessage" class="api-state warning">{{ errorMessage }}</p>
          <p v-else class="api-state">更新时间：{{ lastUpdatedAt || "-" }}</p>
        </div>
      </div>
      <div class="hero-actions">
        <button class="primary-btn finance-primary-btn" type="button" :disabled="loading" @click="loadStrategies">
          {{ loading ? "同步中..." : "刷新策略" }}
        </button>
        <button class="ghost-btn finance-ghost-btn" type="button" @click="goStrategyArchive">看历史档案</button>
        <button class="ghost-btn finance-ghost-btn" type="button" @click="goStrategyWatchlist">去我的关注</button>
      </div>
      <div class="strategy-hero-stats finance-hero-stat-grid">
        <article class="finance-hero-stat-card">
          <span>今日主推荐</span>
          <strong>{{ activeStockView?.name || "待选择股票推荐" }}</strong>
          <p>先看股票主推荐，再继续看理由、风险边界和版本变化。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>股票策略</span>
          <strong>{{ stockRows.length }} 条</strong>
          <p>保留主推荐、观察样本和执行区间，不改现有策略数据行为。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>期货策略</span>
          <strong>{{ futuresRows.length }} 条</strong>
          <p>事件驱动与期货方案继续并列展示，方便从股票切到期货视角。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>阅读顺序</span>
          <strong>{{ activeStockTrackingState.nextAction || "先看主推荐" }}</strong>
          <p>主推荐 -> 风险边界 -> 版本变化 -> 历史档案，保持 demo 的工作台节奏。</p>
        </article>
      </div>
    </header>

    <section class="strategy-focus-layout">
      <article class="card focus-detail-card">
        <header class="focus-detail-head">
          <div>
            <p class="hero-kicker">股票推荐详情</p>
            <h2 class="section-title">
              {{ activeStockView ? `${activeStockView.name} · 推荐详情` : "请选择一条推荐查看详情" }}
            </h2>
            <p class="section-subtitle">
              查看推荐理由、风控边界、版本变化、资讯支撑和历史表现。
            </p>
          </div>
          <div class="focus-detail-actions">
            <button class="primary-btn finance-primary-btn" type="button" :disabled="stockDetailLoading || !activeStockView" @click="refreshActiveStockDetail">
              {{ stockDetailLoading ? "同步中..." : "刷新详情" }}
            </button>
            <button class="ghost-btn finance-ghost-btn" type="button" :disabled="!activeStockView" @click="openStrategyCommunity">
              看股票讨论
            </button>
            <button class="ghost-btn finance-ghost-btn" type="button" :disabled="!activeStockView" @click="openStrategyDiscussionComposer">
              围绕当前标的发观点
            </button>
            <button class="ghost-btn finance-ghost-btn" type="button" :disabled="!activeStockView" @click="toggleActiveStockWatch">
              {{ isActiveStockWatched ? "移出关注" : "加入关注" }}
            </button>
            <button class="ghost-btn finance-ghost-btn" type="button" :disabled="!activeStockView" @click="openActiveStockDetailDialog">
              查看浮层详情
            </button>
          </div>
        </header>

        <div v-if="activeStockView" class="focus-detail-content">
          <section class="focus-summary-panel finance-card-pale">
            <div class="focus-summary-head">
              <div>
                <p class="detail-kicker">今日主推荐</p>
                <h3>{{ activeStockView.name }}</h3>
              </div>
              <div class="focus-badge-row">
                <span class="finance-pill finance-pill-roomy finance-pill-info">评分 {{ activeStockView.score }}</span>
                <span class="finance-pill finance-pill-roomy finance-pill-neutral">{{ activeStockView.risk }}</span>
                <span class="finance-pill finance-pill-roomy finance-pill-info">仓位 {{ activeStockView.position }}</span>
              </div>
            </div>
            <p class="focus-summary-text">
              {{ activeStockInsightSections.whyNow || activeStockView.reason }}
            </p>
            <div v-if="activeStockProofTags.length" class="focus-proof-list">
              <span
                v-for="tag in activeStockProofTags"
                :key="`stock-proof-${tag}`"
                class="finance-pill finance-pill-roomy finance-pill-accent"
              >
                {{ tag }}
              </span>
            </div>
            <div class="focus-kpi-grid">
              <article class="finance-list-card finance-list-card-panel">
                <p>有效时间</p>
                <strong>{{ activeStockView.validRange }}</strong>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>止盈建议</p>
                <strong>{{ activeStockView.takeProfit }}</strong>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>止损建议</p>
                <strong>{{ activeStockView.stopLoss }}</strong>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>下一动作</p>
                <strong>{{ activeStockTrackingState.nextAction }}</strong>
              </article>
            </div>
          </section>

          <div class="tracking-status-box finance-card-surface">
            <div class="tracking-status-head">
              <p>当前跟踪状态</p>
              <span class="status tracking-badge finance-pill" :class="activeStockTrackingState.className">
                {{ activeStockTrackingState.label }}
              </span>
            </div>
            <p class="tracking-status-summary">{{ activeStockTrackingState.summary }}</p>
            <div class="tracking-timeline">
              <article v-for="item in activeStockTrackingTimeline" :key="item.label" class="finance-list-card finance-list-card-panel">
                <p>{{ item.label }}</p>
                <strong>{{ item.value }}</strong>
                <span>{{ item.note }}</span>
              </article>
            </div>
          </div>

          <div class="reason-support-grid">
            <article v-for="item in activeStockReasonSupports" :key="item.label" class="finance-list-card finance-list-card-panel">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div class="risk-boundary-grid">
            <article v-for="item in activeStockRiskBoundaries" :key="item.label" class="finance-list-card finance-list-card-panel">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div v-if="activeStockVersionDiff" class="strategy-version-box finance-card-pale">
            <div class="stock-news-head">
              <p>版本差异</p>
              <span>{{ activeStockVersionDiff.meta }}</span>
            </div>
            <div class="strategy-version-grid">
              <article class="finance-list-card finance-list-card-panel">
                <p>记录版本</p>
                <strong>{{ activeStockVersionDiff.beforeLabel }}</strong>
                <span>{{ activeStockVersionDiff.beforeNote }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>当前解释</p>
                <strong>{{ activeStockVersionDiff.afterLabel }}</strong>
                <span>{{ activeStockVersionDiff.afterNote }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>为什么变了</p>
                <strong>{{ activeStockVersionDiff.diffLabel }}</strong>
                <span>{{ activeStockVersionDiff.diffNote }}</span>
              </article>
            </div>
          </div>

          <div v-if="activeStockOriginCards.length" class="reason-support-grid">
            <article
              v-for="item in activeStockOriginCards"
              :key="`stock-origin-${item.label}`"
              class="finance-list-card finance-list-card-panel"
            >
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div v-if="activeStockVersionHistoryItems.length" class="strategy-history-box finance-card-surface">
            <div class="stock-news-head">
              <p>历史版本</p>
              <div class="strategy-history-head-actions">
                <span>{{ activeStockVersionHistoryItems.length }} 条</span>
                <button
                  v-if="activeStockHistoryCompare?.isCustomSelected"
                  type="button"
                  class="strategy-history-reset finance-mini-btn finance-mini-btn-soft"
                  @click="resetStockVersionHistorySelection"
                >
                  回到默认对比
                </button>
              </div>
            </div>
            <p class="strategy-history-hint">点一条历史版本，直接和当前 explanation 对比。</p>
            <div class="strategy-history-list">
              <button
                v-for="item in activeStockVersionHistoryItems"
                :key="`stock-history-${item.key}`"
                type="button"
                class="strategy-history-item finance-list-card finance-list-card-interactive"
                :class="{ active: activeSelectedStockVersionHistory?.key === item.key }"
                @click="selectStockVersionHistory(item.key)"
              >
                <p>{{ item.title }}</p>
                <strong>{{ item.version }}</strong>
                <span>{{ item.note }}</span>
              </button>
            </div>
            <div v-if="activeStockHistoryCompare?.diff" class="strategy-version-box compact finance-card-pale">
              <div class="stock-news-head">
                <p>版本对照</p>
                <span>{{ activeStockHistoryCompare.selectedTitle }}</span>
              </div>
              <p class="explanation-note">{{ activeStockHistoryCompare.selectedNote }}</p>
              <div class="strategy-version-grid">
                <article class="finance-list-card finance-list-card-panel">
                  <p>{{ activeStockHistoryCompare.isCustomSelected ? "所选版本" : "默认对比版本" }}</p>
                  <strong>{{ activeStockHistoryCompare.diff.beforeLabel }}</strong>
                  <span>{{ activeStockHistoryCompare.diff.beforeNote }}</span>
                </article>
                <article class="finance-list-card finance-list-card-panel">
                  <p>当前解释</p>
                  <strong>{{ activeStockHistoryCompare.diff.afterLabel }}</strong>
                  <span>{{ activeStockHistoryCompare.diff.afterNote }}</span>
                </article>
                <article class="finance-list-card finance-list-card-panel">
                  <p>变化总结</p>
                  <strong>{{ activeStockHistoryCompare.diff.diffLabel }}</strong>
                  <span>{{ activeStockHistoryCompare.diff.diffNote }}</span>
                </article>
              </div>
            </div>
          </div>

          <div v-if="activeStockExplanation" class="strategy-explanation-stack">
            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>为什么选它</p>
                <span>{{ activeStockExplanation.strategy_version || "strategy-engine" }}</span>
              </div>
              <p class="explanation-summary">
                {{ activeStockInsightSections.whyNow || activeStockView.reason }}
              </p>
              <div class="reason-support-grid">
                <article
                  v-for="item in activeStockExplanationCards"
                  :key="`stock-explain-${item.label}`"
                  class="finance-list-card finance-list-card-panel"
                >
                  <p>{{ item.label }}</p>
                  <strong>{{ item.value }}</strong>
                  <span>{{ item.note }}</span>
                </article>
              </div>
              <div class="chip-group" v-if="activeStockSeedHighlights.length > 0">
                <span
                  v-for="item in activeStockSeedHighlights"
                  :key="`stock-seed-${item}`"
                  class="finance-pill finance-pill-compact finance-pill-info"
                >
                  {{ item }}
                </span>
              </div>
            </section>

            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>多场景推演</p>
                <span>{{ activeStockScenarioCards.length }} 个场景</span>
              </div>
              <p class="explanation-note">{{ activeStockInsightSections.proofSource || "系统会从多个场景验证这次推荐。" }}</p>
              <div class="scenario-grid">
                <article
                  v-for="item in activeStockScenarioCards"
                  :key="`stock-scenario-${item.scenario}`"
                  class="scenario-item finance-list-card finance-list-card-panel"
                >
                  <p>{{ item.scenario }}</p>
                  <strong>{{ item.action }}</strong>
                  <span>{{ item.thesis }}</span>
                  <em>风险 {{ item.risk_signal }} · 调整 {{ item.score_adjustment }}</em>
                </article>
              </div>
            </section>

            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>角色评审</p>
                <span>{{ activeStockAgentOpinions.length }} 个视角</span>
              </div>
              <div class="agent-opinion-list">
                <article
                  v-for="item in activeStockAgentOpinions"
                  :key="`stock-agent-${item.agent}`"
                  class="agent-opinion-item finance-list-card finance-list-card-panel"
                >
                  <p>{{ item.agent }}</p>
                  <strong>{{ item.stance }} · {{ formatScore(item.confidence) }}</strong>
                  <span>{{ item.summary }}</span>
                </article>
              </div>
            </section>

            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>风险与失效条件</p>
                <span>{{ activeStockRiskCards.length }} 条</span>
              </div>
              <div class="risk-flag-list">
                <article
                  v-for="item in activeStockRiskCards"
                  :key="`stock-risk-${item.label}-${item.text}`"
                  class="risk-flag-item finance-list-card finance-list-card-panel"
                  :class="{ subtle: item.subtle }"
                >
                  <strong>{{ item.label }}</strong>
                  <span>{{ item.text }}</span>
                </article>
              </div>
            </section>
          </div>

          <div class="stock-news-box finance-card-surface">
            <div class="stock-news-head">
              <p>相关资讯支撑</p>
              <span>{{ activeStockRelatedNews.length }} 条</span>
            </div>
            <div v-if="stockRelatedNewsLoading" class="empty-inline finance-empty-inline">正在同步相关资讯...</div>
            <div v-else-if="activeStockRelatedNews.length > 0" class="stock-news-list">
              <article
                v-for="item in activeStockRelatedNews"
                :key="item.id"
                class="stock-news-item finance-list-card finance-list-card-panel"
              >
                <h4>{{ item.title }}</h4>
                <p class="stock-news-meta">
                  {{ item.source }} · {{ item.time }} · {{ item.visibility }} · 相关度 {{ formatScore01(item.relevanceScore) }}
                </p>
                <p class="stock-news-summary">{{ item.summary }}</p>
              </article>
            </div>
            <div v-else class="empty-inline finance-empty-inline">暂无可匹配的资讯数据</div>
          </div>

          <div class="stock-news-box finance-card-surface">
            <div class="stock-news-head">
              <p>事件证据卡</p>
              <span>{{ activeStockEventEvidenceCards.length }} 条</span>
            </div>
            <div v-if="activeStockEventEvidenceCards.length > 0" class="stock-news-list">
              <article
                v-for="item in activeStockEventEvidenceCards"
                :key="item.key"
                class="stock-news-item finance-list-card finance-list-card-panel"
              >
                <h4>{{ item.value }}</h4>
                <p class="stock-news-meta">{{ item.title }}</p>
                <p class="stock-news-summary">{{ item.note }}</p>
              </article>
            </div>
            <div v-else class="empty-inline finance-empty-inline">暂无已审核事件证据</div>
          </div>

          <div class="stock-performance-box finance-card-surface">
            <div class="stock-performance-head">
              <p>历史推荐业绩</p>
              <span>{{ activeStockPerformanceSummary }}</span>
            </div>
            <p class="stock-performance-note">{{ activeStockPerformanceNote }}</p>
            <div v-if="activeStockPerformanceRows.length > 0" class="performance-table-wrap finance-table-wrap">
              <table class="performance-table finance-data-table finance-data-table-compact">
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
            <div v-else class="empty-inline finance-empty-inline">暂无历史推荐业绩</div>
          </div>

          <p v-if="stockDetailErrorMessage" class="detail-warning">{{ stockDetailErrorMessage }}</p>
        </div>
        <StatePanel
          v-else
          tone="info"
          eyebrow="股票推荐详情"
          title="请选择一条推荐查看详情"
          description="左侧会展示当前选中推荐的评分、资讯和历史表现。"
          compact
        />
      </article>

      <aside class="focus-side-stack">
        <article class="card focus-side-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">期货策略侧看板</h2>
              <p class="section-subtitle">不抢主推荐位置，但保留今天的参数入口和执行边界。</p>
            </div>
          </header>
          <div v-if="activeFuturesView" class="side-highlight-box">
            <p class="detail-kicker">当前选中期货</p>
            <h3>{{ activeFuturesView.name }}</h3>
            <div class="focus-badge-row">
              <span class="finance-pill finance-pill-roomy finance-pill-info">{{ activeFuturesView.direction }}</span>
              <span class="finance-pill finance-pill-roomy finance-pill-neutral">{{ activeFuturesView.risk }}</span>
              <span class="finance-pill finance-pill-roomy finance-pill-info">仓位 {{ activeFuturesView.position }}</span>
            </div>
            <p class="side-highlight-text">
              {{ activeFuturesInsightSections.whyNow || activeFuturesView.reason }}
            </p>
            <div v-if="activeFuturesProofTags.length" class="focus-proof-list">
              <span
                v-for="tag in activeFuturesProofTags"
                :key="`futures-proof-${tag}`"
                class="finance-pill finance-pill-roomy finance-pill-accent"
              >
                {{ tag }}
              </span>
            </div>
            <div class="side-mini-grid">
              <article class="finance-list-card finance-list-card-panel">
                <p>入场</p>
                <strong>{{ activeFuturesView.entryRange }}</strong>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>止盈</p>
                <strong>{{ activeFuturesView.takeProfitRange }}</strong>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>止损</p>
                <strong>{{ activeFuturesView.stopLossRange }}</strong>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>失效</p>
                <strong>{{ activeFuturesView.invalidCondition }}</strong>
              </article>
            </div>
            <div class="focus-link-row">
              <button type="button" class="ghost-btn finance-ghost-btn" :disabled="futuresInsightLoading" @click="refreshActiveFuturesInsight">
                {{ futuresInsightLoading ? "同步中..." : "刷新期货详情" }}
              </button>
              <button type="button" class="ghost-btn finance-ghost-btn" :disabled="!activeFuturesView" @click="openFuturesCommunity">
                看期货讨论
              </button>
              <button type="button" class="ghost-btn finance-ghost-btn" :disabled="!activeFuturesView" @click="openFuturesDiscussionComposer">
                围绕当前期货发观点
              </button>
              <button type="button" class="ghost-btn finance-ghost-btn" @click="openActiveFuturesDetailDialog">完整查看</button>
            </div>
          </div>
          <StatePanel
            v-else
            tone="info"
            eyebrow="期货策略"
            title="当前还没有可选中的期货策略"
            description="同步期货策略后，这里会展示当前选中的参数摘要。"
            compact
          />
        </article>

        <article class="card focus-side-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">市场事件提示</h2>
              <p class="section-subtitle">事件模块不回塞首页，而是在策略页负责强化或削弱判断。</p>
            </div>
          </header>
          <div v-if="activeEventView" class="side-highlight-box">
            <div class="event-top">
              <p>{{ activeEventView.symbol }}</p>
              <span class="event-level finance-pill finance-pill-compact" :class="activeEventView.levelClass">{{ activeEventView.levelText }}</span>
            </div>
            <h3 class="side-event-title">{{ activeEventView.title }}</h3>
            <p class="side-highlight-text">{{ activeEventView.summary }}</p>
            <div class="side-mini-grid side-mini-grid--single">
              <article class="finance-list-card finance-list-card-panel">
                <p>触发规则</p>
                <strong>{{ activeEventView.triggerRule }}</strong>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>来源</p>
                <strong>{{ activeEventView.source || "-" }}</strong>
              </article>
            </div>
          </div>
          <div v-else class="empty-inline finance-empty-inline">当前没有可展示的事件摘要。</div>
        </article>
      </aside>
    </section>

    <div class="strategy-grid">
      <article class="card matrix-card">
        <header class="finance-copy-stack">
          <h2 class="section-title">股票推荐策略</h2>
          <p class="section-subtitle">保留排序感，但更强调切换研究对象，而不是点完就弹窗。</p>
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
            class="matrix-item finance-list-card finance-list-card-interactive"
            :class="{ active: activeStockID === item.id }"
            @click="focusStock(item.id)"
          >
            <div class="matrix-main">
              <p class="name">{{ item.name }}</p>
              <p class="desc">{{ item.desc }}</p>
              <div v-if="item.proofTags.length" class="list-proof-tags">
                <span v-for="tag in item.proofTags" :key="`${item.id}-${tag}`">{{ tag }}</span>
              </div>
              <p v-if="item.supportingText" class="list-proof-note">{{ item.supportingText }}</p>
            </div>
            <p class="score">{{ item.score }}</p>
            <p>{{ item.risk }}</p>
            <p>{{ item.position }}</p>
            <span class="status finance-pill" :class="item.statusClass">{{ item.status }}</span>
            <div v-if="item.metaText" class="matrix-proof">
              <span>{{ item.metaText }}</span>
            </div>
          </button>
        </div>
        <p v-if="stockRows.length" class="matrix-tip">点击任意推荐，左侧主详情区会切换到对应解释链和历史样本。</p>
        <div v-else class="empty-box finance-empty-box">暂无股票推荐策略，可先刷新或去历史档案看最近批次。</div>
      </article>

      <aside class="card cycle-card">
        <header class="finance-copy-stack">
          <h2 class="section-title">期货策略清单</h2>
          <p class="section-subtitle">每个策略保留关键执行参数，右上侧看板会同步展示当前选中项。</p>
        </header>
        <div class="cycle-list" v-if="futuresRows.length">
          <button
            v-for="step in futuresRows"
            :key="step.id"
            type="button"
            class="cycle-item finance-list-card finance-list-card-interactive"
            :class="{ active: activeFuturesID === step.id }"
            @click="focusFutures(step.id)"
          >
            <p class="step">{{ step.contract }}</p>
            <p class="title">{{ step.title }}</p>
            <p class="note">{{ step.note }}</p>
            <div v-if="step.proofTags.length" class="list-proof-tags compact">
              <span v-for="tag in step.proofTags" :key="`${step.id}-${tag}`">{{ tag }}</span>
            </div>
            <p v-if="step.supportingText" class="cycle-proof-note">{{ step.supportingText }}</p>
            <p v-if="step.metaText" class="cycle-proof-meta">{{ step.metaText }}</p>
          </button>
        </div>
        <p v-if="futuresRows.length" class="matrix-tip">点击任意期货策略，侧边摘要和下方事件判断会同步切换。</p>
        <div v-else class="empty-box finance-empty-box">暂无期货策略，可先刷新或去历史档案看最近批次。</div>
      </aside>
    </div>

    <section class="event-section">
      <article class="card event-list-card">
        <header class="finance-copy-stack">
          <h2 class="section-title">市场事件雷达</h2>
          <p class="section-subtitle">接入 market/events，统一查看策略触发事件。</p>
        </header>

        <div class="event-type-tabs">
          <button
            v-for="item in eventTypeOptions"
            :key="item.value"
            type="button"
            class="finance-toggle-btn"
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
            class="event-item finance-list-card finance-list-card-interactive"
            :class="{ active: activeEventID === item.id }"
            @click="activeEventID = item.id"
          >
            <div class="event-top">
              <p>{{ item.symbol }}</p>
              <span class="event-level finance-pill finance-pill-compact" :class="item.levelClass">{{ item.levelText }}</span>
            </div>
            <p class="event-title">{{ item.title }}</p>
            <p class="event-meta">{{ item.typeLabel }} · {{ item.time }}</p>
          </button>
        </div>
        <div v-else class="empty-box finance-empty-box">当前条件下暂无市场事件</div>
      </article>

      <article class="card event-detail-card">
        <header class="finance-copy-stack">
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
            <button
              type="button"
              class="finance-mini-btn finance-mini-btn-soft"
              :disabled="eventDetailLoading"
              @click="refreshActiveEventDetail"
            >
              {{ eventDetailLoading ? "同步中..." : "刷新事件详情" }}
            </button>
          </div>
        </div>
        <div v-else class="empty-box finance-empty-box">暂无可展示的事件详情</div>
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
              <p class="detail-kicker">推荐档案</p>
              <h3>{{ activeStockView.name }}</h3>
            </div>
            <div class="stock-detail-actions">
              <button type="button" class="watch-btn finance-mini-btn finance-mini-btn-accent" @click="toggleActiveStockWatch">
                {{ isActiveStockWatched ? "移出关注" : "加入关注" }}
              </button>
              <button
                type="button"
                class="finance-mini-btn finance-mini-btn-card"
                :disabled="stockDetailLoading"
                @click="refreshActiveStockDetail"
              >
                {{ stockDetailLoading ? "同步中..." : "刷新业绩" }}
              </button>
              <button type="button" class="close-btn finance-mini-btn finance-mini-btn-card" @click="closeStockDetail">关闭</button>
            </div>
          </header>

          <p class="stock-reason">{{ activeStockInsightSections.whyNow || activeStockView.reason }}</p>
          <p v-if="activeStockScoreFramework?.method" class="stock-reason-meta">
            评分模型：{{ activeStockScoreFramework.method }}
          </p>

          <div class="stock-detail-grid">
            <article class="finance-list-card finance-list-card-panel">
              <p>综合评分</p>
              <strong>{{ activeStockView.score }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>风险等级</p>
              <strong>{{ activeStockView.risk }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>建议仓位</p>
              <strong>{{ activeStockView.position }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>有效时间</p>
              <strong>{{ activeStockView.validRange }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>止盈建议</p>
              <strong>{{ activeStockView.takeProfit }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>止损建议</p>
              <strong>{{ activeStockView.stopLoss }}</strong>
            </article>
          </div>

          <div class="tracking-status-box finance-card-surface">
            <div class="tracking-status-head">
              <p>当前跟踪状态</p>
              <span class="status tracking-badge finance-pill" :class="activeStockTrackingState.className">
                {{ activeStockTrackingState.label }}
              </span>
            </div>
            <p class="tracking-status-summary">{{ activeStockTrackingState.summary }}</p>
            <div class="tracking-timeline">
              <article v-for="item in activeStockTrackingTimeline" :key="item.label" class="finance-list-card finance-list-card-panel">
                <p>{{ item.label }}</p>
                <strong>{{ item.value }}</strong>
                <span>{{ item.note }}</span>
              </article>
            </div>
          </div>

          <div class="risk-boundary-grid">
            <article v-for="item in activeStockRiskBoundaries" :key="item.label" class="finance-list-card finance-list-card-panel">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div class="reason-support-grid">
            <article v-for="item in activeStockReasonSupports" :key="item.label" class="finance-list-card finance-list-card-panel">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div v-if="activeStockVersionDiff" class="strategy-version-box finance-card-pale">
            <div class="stock-news-head">
              <p>版本差异</p>
              <span>{{ activeStockVersionDiff.meta }}</span>
            </div>
            <div class="strategy-version-grid">
              <article class="finance-list-card finance-list-card-panel">
                <p>记录版本</p>
                <strong>{{ activeStockVersionDiff.beforeLabel }}</strong>
                <span>{{ activeStockVersionDiff.beforeNote }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>当前解释</p>
                <strong>{{ activeStockVersionDiff.afterLabel }}</strong>
                <span>{{ activeStockVersionDiff.afterNote }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>为什么变了</p>
                <strong>{{ activeStockVersionDiff.diffLabel }}</strong>
                <span>{{ activeStockVersionDiff.diffNote }}</span>
              </article>
            </div>
          </div>

          <div v-if="activeStockOriginCards.length" class="reason-support-grid">
            <article
              v-for="item in activeStockOriginCards"
              :key="`stock-origin-${item.label}`"
              class="finance-list-card finance-list-card-panel"
            >
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div v-if="activeStockVersionHistoryItems.length" class="strategy-history-box finance-card-surface">
            <div class="stock-news-head">
              <p>历史版本</p>
              <div class="strategy-history-head-actions">
                <span>{{ activeStockVersionHistoryItems.length }} 条</span>
                <button
                  v-if="activeStockHistoryCompare?.isCustomSelected"
                  type="button"
                  class="strategy-history-reset finance-mini-btn finance-mini-btn-soft"
                  @click="resetStockVersionHistorySelection"
                >
                  回到默认对比
                </button>
              </div>
            </div>
            <p class="strategy-history-hint">点一条历史版本，直接和当前 explanation 对比。</p>
            <div class="strategy-history-list">
              <button
                v-for="item in activeStockVersionHistoryItems"
                :key="`stock-history-${item.key}`"
                type="button"
                class="strategy-history-item finance-list-card finance-list-card-interactive"
                :class="{ active: activeSelectedStockVersionHistory?.key === item.key }"
                @click="selectStockVersionHistory(item.key)"
              >
                <p>{{ item.title }}</p>
                <strong>{{ item.version }}</strong>
                <span>{{ item.note }}</span>
              </button>
            </div>
            <div v-if="activeStockHistoryCompare?.diff" class="strategy-version-box compact finance-card-pale">
              <div class="stock-news-head">
                <p>版本对照</p>
                <span>{{ activeStockHistoryCompare.selectedTitle }}</span>
              </div>
              <p class="explanation-note">{{ activeStockHistoryCompare.selectedNote }}</p>
              <div class="strategy-version-grid">
                <article class="finance-list-card finance-list-card-panel">
                  <p>{{ activeStockHistoryCompare.isCustomSelected ? "所选版本" : "默认对比版本" }}</p>
                  <strong>{{ activeStockHistoryCompare.diff.beforeLabel }}</strong>
                  <span>{{ activeStockHistoryCompare.diff.beforeNote }}</span>
                </article>
                <article class="finance-list-card finance-list-card-panel">
                  <p>当前解释</p>
                  <strong>{{ activeStockHistoryCompare.diff.afterLabel }}</strong>
                  <span>{{ activeStockHistoryCompare.diff.afterNote }}</span>
                </article>
                <article class="finance-list-card finance-list-card-panel">
                  <p>变化总结</p>
                  <strong>{{ activeStockHistoryCompare.diff.diffLabel }}</strong>
                  <span>{{ activeStockHistoryCompare.diff.diffNote }}</span>
                </article>
              </div>
            </div>
          </div>

          <div v-if="activeStockExplanation" class="strategy-explanation-stack">
            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>为什么选它</p>
                <span>{{ activeStockExplanation.strategy_version || "strategy-engine" }}</span>
              </div>
              <p class="explanation-summary">
                {{ activeStockInsightSections.whyNow || activeStockView.reason }}
              </p>
              <div class="reason-support-grid">
                <article
                  v-for="item in activeStockExplanationCards"
                  :key="`stock-explain-${item.label}`"
                  class="finance-list-card finance-list-card-panel"
                >
                  <p>{{ item.label }}</p>
                  <strong>{{ item.value }}</strong>
                  <span>{{ item.note }}</span>
                </article>
              </div>
              <div class="chip-group" v-if="activeStockSeedHighlights.length > 0">
                <span
                  v-for="item in activeStockSeedHighlights"
                  :key="`stock-seed-${item}`"
                  class="finance-pill finance-pill-compact finance-pill-info"
                >
                  {{ item }}
                </span>
              </div>
            </section>

            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>多场景推演</p>
                <span>{{ activeStockScenarioCards.length }} 个场景</span>
              </div>
              <p class="explanation-note">{{ activeStockInsightSections.proofSource || "系统会从多个场景验证这次推荐。" }}</p>
              <div class="scenario-grid">
                <article
                  v-for="item in activeStockScenarioCards"
                  :key="`stock-scenario-${item.scenario}`"
                  class="scenario-item finance-list-card finance-list-card-panel"
                >
                  <p>{{ item.scenario }}</p>
                  <strong>{{ item.action }}</strong>
                  <span>{{ item.thesis }}</span>
                  <em>风险 {{ item.risk_signal }} · 调整 {{ item.score_adjustment }}</em>
                </article>
              </div>
            </section>

            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>角色评审</p>
                <span>{{ activeStockAgentOpinions.length }} 个视角</span>
              </div>
              <div class="agent-opinion-list">
                <article
                  v-for="item in activeStockAgentOpinions"
                  :key="`stock-agent-${item.agent}`"
                  class="agent-opinion-item finance-list-card finance-list-card-panel"
                >
                  <p>{{ item.agent }}</p>
                  <strong>{{ item.stance }} · {{ formatScore(item.confidence) }}</strong>
                  <span>{{ item.summary }}</span>
                </article>
              </div>
            </section>

            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>风险与失效条件</p>
                <span>{{ activeStockRiskCards.length }} 条</span>
              </div>
              <div class="risk-flag-list">
                <article
                  v-for="item in activeStockRiskCards"
                  :key="`stock-risk-${item.label}-${item.text}`"
                  class="risk-flag-item finance-list-card finance-list-card-panel"
                  :class="{ subtle: item.subtle }"
                >
                  <strong>{{ item.label }}</strong>
                  <span>{{ item.text }}</span>
                </article>
              </div>
            </section>
          </div>

          <div class="stock-news-box finance-card-surface">
            <div class="stock-news-head">
              <p>相关资讯支撑</p>
              <span>{{ activeStockRelatedNews.length }} 条</span>
            </div>
            <div v-if="stockRelatedNewsLoading" class="empty-inline finance-empty-inline">正在同步相关资讯...</div>
            <div v-else-if="activeStockRelatedNews.length > 0" class="stock-news-list">
              <article
                v-for="item in activeStockRelatedNews"
                :key="item.id"
                class="stock-news-item finance-list-card finance-list-card-panel"
              >
                <h4>{{ item.title }}</h4>
                <p class="stock-news-meta">
                  {{ item.source }} · {{ item.time }} · {{ item.visibility }} · 相关度 {{ formatScore01(item.relevanceScore) }}
                </p>
                <p class="stock-news-summary">{{ item.summary }}</p>
              </article>
            </div>
            <div v-else class="empty-inline finance-empty-inline">暂无可匹配的资讯数据</div>
          </div>

          <div class="stock-news-box finance-card-surface">
            <div class="stock-news-head">
              <p>事件证据卡</p>
              <span>{{ activeStockEventEvidenceCards.length }} 条</span>
            </div>
            <div v-if="activeStockEventEvidenceCards.length > 0" class="stock-news-list">
              <article
                v-for="item in activeStockEventEvidenceCards"
                :key="`dialog-${item.key}`"
                class="stock-news-item finance-list-card finance-list-card-panel"
              >
                <h4>{{ item.value }}</h4>
                <p class="stock-news-meta">{{ item.title }}</p>
                <p class="stock-news-summary">{{ item.note }}</p>
              </article>
            </div>
            <div v-else class="empty-inline finance-empty-inline">暂无已审核事件证据</div>
          </div>

          <div class="stock-performance-box finance-card-surface">
            <div class="stock-performance-head">
              <p>历史推荐业绩</p>
              <span>{{ activeStockPerformanceSummary }}</span>
            </div>
            <p class="stock-performance-note">{{ activeStockPerformanceNote }}</p>
            <div v-if="activeStockPerformanceRows.length > 0" class="performance-table-wrap finance-table-wrap">
              <table class="performance-table finance-data-table finance-data-table-compact">
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
            <div v-else class="empty-inline finance-empty-inline">暂无历史推荐业绩</div>
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
              <button
                type="button"
                class="finance-mini-btn finance-mini-btn-card"
                :disabled="futuresInsightLoading"
                @click="refreshActiveFuturesInsight"
              >
                {{ futuresInsightLoading ? "同步中..." : "刷新详情" }}
              </button>
              <button type="button" class="close-btn finance-mini-btn finance-mini-btn-card" @click="closeFuturesDetail">关闭</button>
            </div>
          </header>

          <p class="stock-reason">{{ activeFuturesInsightSections.whyNow || activeFuturesView.reason }}</p>
          <p v-if="activeFuturesScoreFramework?.method" class="stock-reason-meta">
            评分模型：{{ activeFuturesScoreFramework.method }}
          </p>

          <div class="futures-detail-grid">
            <article class="finance-list-card finance-list-card-panel">
              <p>方向</p>
              <strong>{{ activeFuturesView.direction }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>风险等级</p>
              <strong>{{ activeFuturesView.risk }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>建议仓位</p>
              <strong>{{ activeFuturesView.position }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>有效时间</p>
              <strong>{{ activeFuturesView.validRange }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>入场区间</p>
              <strong>{{ activeFuturesView.entryRange }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>止盈区间</p>
              <strong>{{ activeFuturesView.takeProfitRange }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>止损区间</p>
              <strong>{{ activeFuturesView.stopLossRange }}</strong>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>失效条件</p>
              <strong>{{ activeFuturesView.invalidCondition }}</strong>
            </article>
          </div>

          <div class="reason-support-grid">
            <article v-for="item in activeFuturesReasonSupports" :key="`fut-${item.label}`" class="finance-list-card finance-list-card-panel">
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div v-if="activeFuturesVersionDiff" class="strategy-version-box finance-card-pale">
            <div class="stock-news-head">
              <p>版本差异</p>
              <span>{{ activeFuturesVersionDiff.meta }}</span>
            </div>
            <div class="strategy-version-grid">
              <article class="finance-list-card finance-list-card-panel">
                <p>记录版本</p>
                <strong>{{ activeFuturesVersionDiff.beforeLabel }}</strong>
                <span>{{ activeFuturesVersionDiff.beforeNote }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>当前解释</p>
                <strong>{{ activeFuturesVersionDiff.afterLabel }}</strong>
                <span>{{ activeFuturesVersionDiff.afterNote }}</span>
              </article>
              <article class="finance-list-card finance-list-card-panel">
                <p>为什么变了</p>
                <strong>{{ activeFuturesVersionDiff.diffLabel }}</strong>
                <span>{{ activeFuturesVersionDiff.diffNote }}</span>
              </article>
            </div>
          </div>

          <div v-if="activeFuturesOriginCards.length" class="reason-support-grid">
            <article
              v-for="item in activeFuturesOriginCards"
              :key="`futures-origin-${item.label}`"
              class="finance-list-card finance-list-card-panel"
            >
              <p>{{ item.label }}</p>
              <strong>{{ item.value }}</strong>
              <span>{{ item.note }}</span>
            </article>
          </div>

          <div v-if="activeFuturesVersionHistoryItems.length" class="strategy-history-box finance-card-surface">
            <div class="stock-news-head">
              <p>历史版本</p>
              <div class="strategy-history-head-actions">
                <span>{{ activeFuturesVersionHistoryItems.length }} 条</span>
                <button
                  v-if="activeFuturesHistoryCompare?.isCustomSelected"
                  type="button"
                  class="strategy-history-reset finance-mini-btn finance-mini-btn-soft"
                  @click="resetFuturesVersionHistorySelection"
                >
                  回到默认对比
                </button>
              </div>
            </div>
            <p class="strategy-history-hint">点一条历史版本，直接和当前 explanation 对比。</p>
            <div class="strategy-history-list">
              <button
                v-for="item in activeFuturesVersionHistoryItems"
                :key="`futures-history-${item.key}`"
                type="button"
                class="strategy-history-item finance-list-card finance-list-card-interactive"
                :class="{ active: activeSelectedFuturesVersionHistory?.key === item.key }"
                @click="selectFuturesVersionHistory(item.key)"
              >
                <p>{{ item.title }}</p>
                <strong>{{ item.version }}</strong>
                <span>{{ item.note }}</span>
              </button>
            </div>
            <div v-if="activeFuturesHistoryCompare?.diff" class="strategy-version-box compact finance-card-pale">
              <div class="stock-news-head">
                <p>版本对照</p>
                <span>{{ activeFuturesHistoryCompare.selectedTitle }}</span>
              </div>
              <p class="explanation-note">{{ activeFuturesHistoryCompare.selectedNote }}</p>
              <div class="strategy-version-grid">
                <article class="finance-list-card finance-list-card-panel">
                  <p>{{ activeFuturesHistoryCompare.isCustomSelected ? "所选版本" : "默认对比版本" }}</p>
                  <strong>{{ activeFuturesHistoryCompare.diff.beforeLabel }}</strong>
                  <span>{{ activeFuturesHistoryCompare.diff.beforeNote }}</span>
                </article>
                <article class="finance-list-card finance-list-card-panel">
                  <p>当前解释</p>
                  <strong>{{ activeFuturesHistoryCompare.diff.afterLabel }}</strong>
                  <span>{{ activeFuturesHistoryCompare.diff.afterNote }}</span>
                </article>
                <article class="finance-list-card finance-list-card-panel">
                  <p>变化总结</p>
                  <strong>{{ activeFuturesHistoryCompare.diff.diffLabel }}</strong>
                  <span>{{ activeFuturesHistoryCompare.diff.diffNote }}</span>
                </article>
              </div>
            </div>
          </div>

          <div v-if="activeFuturesExplanation" class="strategy-explanation-stack">
            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>为什么选这个策略</p>
                <span>{{ activeFuturesExplanation.strategy_version || "strategy-engine" }}</span>
              </div>
              <p class="explanation-summary">
                {{ activeFuturesInsightSections.whyNow || activeFuturesView.reason }}
              </p>
              <div class="reason-support-grid">
                <article
                  v-for="item in activeFuturesExplanationCards"
                  :key="`fut-explain-${item.label}`"
                  class="finance-list-card finance-list-card-panel"
                >
                  <p>{{ item.label }}</p>
                  <strong>{{ item.value }}</strong>
                  <span>{{ item.note }}</span>
                </article>
              </div>
              <div class="chip-group" v-if="activeFuturesSeedHighlights.length > 0">
                <span
                  v-for="item in activeFuturesSeedHighlights"
                  :key="`fut-seed-${item}`"
                  class="finance-pill finance-pill-compact finance-pill-info"
                >
                  {{ item }}
                </span>
              </div>
            </section>

            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>商品链证据</p>
                <span>{{ activeFuturesSupplyChainEntities.length }} 个实体</span>
              </div>
              <div class="reason-support-grid">
                <article
                  v-for="item in activeFuturesSupplyChainSummaryCards"
                  :key="`fut-supply-${item.label}`"
                  class="finance-list-card finance-list-card-panel"
                >
                  <p>{{ item.label }}</p>
                  <strong>{{ item.value }}</strong>
                  <span>{{ item.note }}</span>
                </article>
              </div>
              <div v-if="activeFuturesSupplyChainEntities.length" class="chip-group">
                <span
                  v-for="item in activeFuturesSupplyChainEntities"
                  :key="`fut-supply-entity-${item.key}`"
                  class="finance-pill finance-pill-compact finance-pill-accent"
                >
                  {{ item.typeLabel }} · {{ item.label }}
                </span>
              </div>
              <div v-if="activeFuturesSupplyChainEvidenceCards.length" class="stock-news-list" style="margin-top: 16px">
                <article
                  v-for="item in activeFuturesSupplyChainEvidenceCards"
                  :key="`fut-supply-card-${item.key}`"
                  class="stock-news-item finance-list-card finance-list-card-panel"
                >
                  <h4>{{ item.value }}</h4>
                  <p class="stock-news-meta">{{ item.title }}</p>
                  <p class="stock-news-summary">{{ item.note }}</p>
                </article>
              </div>
            </section>

            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>多场景推演</p>
                <span>{{ activeFuturesScenarioCards.length }} 个场景</span>
              </div>
              <p class="explanation-note">{{ activeFuturesInsightSections.proofSource || "系统会在不同市场情景下验证策略。 " }}</p>
              <div class="scenario-grid">
                <article
                  v-for="item in activeFuturesScenarioCards"
                  :key="`fut-scenario-${item.scenario}`"
                  class="scenario-item finance-list-card finance-list-card-panel"
                >
                  <p>{{ item.scenario }}</p>
                  <strong>{{ item.action }}</strong>
                  <span>{{ item.thesis }}</span>
                  <em>风险 {{ item.risk_signal }} · 调整 {{ item.score_adjustment }}</em>
                </article>
              </div>
            </section>

            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>角色评审</p>
                <span>{{ activeFuturesAgentOpinions.length }} 个视角</span>
              </div>
              <div class="agent-opinion-list">
                <article
                  v-for="item in activeFuturesAgentOpinions"
                  :key="`fut-agent-${item.agent}`"
                  class="agent-opinion-item finance-list-card finance-list-card-panel"
                >
                  <p>{{ item.agent }}</p>
                  <strong>{{ item.stance }} · {{ formatScore(item.confidence) }}</strong>
                  <span>{{ item.summary }}</span>
                </article>
              </div>
            </section>

            <section class="strategy-explanation-box finance-card-pale">
              <div class="stock-news-head">
                <p>风险与失效条件</p>
                <span>{{ activeFuturesRiskCards.length }} 条</span>
              </div>
              <div class="risk-flag-list">
                <article
                  v-for="item in activeFuturesRiskCards"
                  :key="`fut-risk-${item.label}-${item.text}`"
                  class="risk-flag-item finance-list-card finance-list-card-panel"
                  :class="{ subtle: item.subtle }"
                >
                  <strong>{{ item.label }}</strong>
                  <span>{{ item.text }}</span>
                </article>
              </div>
            </section>
          </div>

          <div class="stock-news-box finance-card-surface">
            <div class="stock-news-head">
              <p>相关资讯支撑</p>
              <span>{{ activeFuturesRelatedNews.length }} 条</span>
            </div>
            <div v-if="futuresInsightLoading" class="empty-inline finance-empty-inline">正在同步相关资讯...</div>
            <div v-else-if="activeFuturesRelatedNews.length > 0" class="stock-news-list">
              <article
                v-for="item in activeFuturesRelatedNews"
                :key="`fut-news-${item.id}`"
                class="stock-news-item finance-list-card finance-list-card-panel"
              >
                <h4>{{ item.title }}</h4>
                <p class="stock-news-meta">
                  {{ item.source }} · {{ formatDateTime(item.published_at || item.time) }} ·
                  {{ mapVisibilityLabel(item.visibility) }} · 相关度 {{ formatScore01(item.relevance_score ?? item.relevanceScore) }}
                </p>
                <p class="stock-news-summary">{{ summarizeNewsText(item.summary) }}</p>
              </article>
            </div>
            <div v-else class="empty-inline finance-empty-inline">暂无可匹配资讯</div>
          </div>

          <div class="futures-event-box finance-card-surface">
            <div class="stock-news-head">
              <p>相关市场事件</p>
              <span>{{ activeFuturesRelatedEvents.length }} 条</span>
            </div>
            <div v-if="activeFuturesRelatedEvents.length > 0" class="futures-event-list">
              <article
                v-for="item in activeFuturesRelatedEvents"
                :key="`fut-event-${item.id}`"
                class="futures-event-item finance-list-card finance-list-card-compact finance-list-card-panel"
              >
                <p>{{ mapEventType(item.event_type || item.type) }} · {{ item.symbol || "-" }}</p>
                <strong>{{ item.summary || "-" }}</strong>
                <span>{{ formatDateTime(item.created_at || item.time) }}</span>
              </article>
            </div>
            <div v-else class="empty-inline finance-empty-inline">暂无相关事件</div>
          </div>

          <div class="stock-performance-box finance-card-surface">
            <div class="stock-performance-head">
              <p>历史策略业绩</p>
              <span>{{ activeFuturesPerformanceSummary }}</span>
            </div>
            <div v-if="activeFuturesPerformanceRows.length > 0" class="performance-table-wrap finance-table-wrap">
              <table class="performance-table finance-data-table finance-data-table-compact">
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
            <div v-else class="empty-inline finance-empty-inline">暂无历史策略业绩</div>
          </div>

          <p v-if="futuresInsightErrorMessage" class="detail-warning">{{ futuresInsightErrorMessage }}</p>
        </section>
      </div>
    </transition>
  </section>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import StatePanel from "../components/StatePanel.vue";
import {
  getFuturesStrategyInsight,
  getFuturesStrategyVersionHistory,
  getStockRecommendationInsight,
  getStockRecommendationDetail,
  getStockRecommendationPerformance,
  getStockRecommendationVersionHistory,
  getMarketEventDetail,
  listFuturesStrategies,
  listMarketEvents,
  listStockRecommendations
} from "../api/market";
import { getMembershipQuota } from "../api/membership";
import { listNewsArticles } from "../api/news";
import {
  rememberExperimentAttributionSource,
  rememberPendingExperimentJourneySource,
  promotePendingExperimentJourneySources,
  trackExperimentEvent,
  trackExperimentExposureOnce
} from "../lib/growth-analytics";
import { getExperimentVariant } from "../lib/growth-experiments";
import { shouldUseDemoFallback } from "../lib/fallback-policy";
import {
  buildFallbackStrategyVersionHistory,
  buildStrategyEventEvidenceCards,
  buildStrategyHistoryCompareState,
  buildStrategyBatchText,
  buildStrategyInsightSections,
  buildStrategyMetaText,
  buildStrategyOriginCards,
  buildStrategyProofTags,
  buildStrategyRiskBoundaryText,
  buildStrategyVersionDiff,
  firstMeaningfulStrategyText,
  mapStrategyVersionHistory
} from "../lib/strategy-version";
import {
  buildCommunityComposeRoute,
  buildCommunityListRoute,
  findEntityIDByTarget
} from "../lib/community-entry-links";
import { useClientAuth } from "../lib/client-auth";
import { WATCHLIST_EVENT, isWatchedStock, removeWatchedStock, saveWatchedStock } from "../lib/watchlist";

const route = useRoute();
const router = useRouter();
const { isLoggedIn } = useClientAuth();
const strategyExperimentVariant = getExperimentVariant("strategy_membership_cta", ["cadence", "proof"]);

const loading = ref(false);
const errorMessage = ref("");
const lastUpdatedAt = ref("");
const memberStageLoading = ref(false);
const isVIPUser = ref(false);

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
const stockExplanationMap = ref({});
const stockVersionHistoryMap = ref({});
const loadedStockInsightIDs = new Set();
const loadedStockDetailIDs = new Set();
const loadedStockPerformanceIDs = new Set();
const loadedStockRelatedNewsIDs = new Set();
const loadedStockVersionHistoryIDs = new Set();
const activeStockID = ref(useDemoFallback ? rawStocks.value[0]?.id || "" : "");
const stockDetailLoading = ref(false);
const stockDetailErrorMessage = ref("");
const stockDetailDialogVisible = ref(false);
const stockDetailDialogRef = ref(null);
const stockDetailDialogStyle = ref({});
const stockDetailAnchorRect = ref(null);
const stockRelatedNewsLoading = ref(false);
const selectedStockVersionHistoryKey = ref("");
const watchlistVersion = ref(0);
const futuresInsightMap = ref(useDemoFallback ? { ...fallbackFuturesInsights } : {});
const futuresExplanationMap = ref({});
const futuresVersionHistoryMap = ref({});
const loadedFuturesInsightIDs = new Set();
const loadedFuturesVersionHistoryIDs = new Set();
const activeFuturesID = ref(useDemoFallback ? rawFutures.value[0]?.id || "" : "");
const selectedFuturesVersionHistoryKey = ref("");
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
  (rawStocks.value || []).map((item) => {
    const explanation = stockExplanationMap.value[item.id] || null;
    const sections = buildStrategyInsightSections(explanation, item.reason_summary || "-");
    const status = mapStatus(item.status);
    return {
      id: item.id,
      name: `${item.symbol || "-"} ${item.name || ""}`.trim(),
      desc: sections.whyNow || item.reason_summary || "-",
      proofTags: buildListProofTags(explanation, { limit: 3 }),
      supportingText: sections.proofSource || sections.riskBoundary || "",
      metaText: buildListMetaText(explanation),
      score: formatScore(item.score),
      risk: mapRisk(item.risk_level),
      position: item.position_range || "-",
      status: status.label,
      statusClass: status.className
    };
  })
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
  return {
    id: base.id,
    name: `${base.symbol || "-"} ${base.name || ""}`.trim(),
    reason: base.reason_summary || "等待同步推荐逻辑说明。",
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

const activeStockExplanation = computed(() => {
  const stockID = activeStockView.value?.id;
  if (!stockID) {
    return null;
  }
  return stockExplanationMap.value[stockID] || null;
});

const activeStockInsightSections = computed(() =>
  buildStrategyInsightSections(activeStockExplanation.value, activeStockView.value?.reason || "")
);

const activeStockExplanationCards = computed(() => {
  const explanation = activeStockExplanation.value || {};
  const evidenceCount = Array.isArray(explanation.evidence_cards) ? explanation.evidence_cards.length : 0;
  const agentCount = Number(explanation.workload_summary?.agent_count ?? 0);
  const scenarioCount = Number(explanation.workload_summary?.scenario_count ?? 0);
  return [
    {
      label: "种子输入",
      value: Array.isArray(explanation.seed_highlights) ? explanation.seed_highlights.length : 0,
      note: explanation.seed_summary || activeStockInsightSections.value.whyNow || "系统会先处理种子输入，再逐步筛选。"
    },
    {
      label: "证据卡片",
      value: evidenceCount,
      note: activeStockInsightSections.value.proofSource || "系统会把证据和多角色结论汇总后展示。"
    },
    {
      label: "评审覆盖",
      value: `${agentCount} 角 / ${scenarioCount} 景`,
      note: buildStrategyRiskBoundaryText(explanation, "当前未补更多风险边界。")
    }
  ];
});

const activeStockProofTags = computed(() => buildListProofTags(activeStockExplanation.value, { limit: 4 }));
const activeStockSeedHighlights = computed(() => activeStockExplanation.value?.seed_highlights || []);
const activeStockScenarioCards = computed(() => activeStockExplanation.value?.simulations?.[0]?.scenarios || []);
const activeStockAgentOpinions = computed(() => activeStockExplanation.value?.agent_opinions || []);
const activeStockRiskCards = computed(() => buildStrategyRiskCards(activeStockExplanation.value));
const activeStockOriginCards = computed(() => buildStrategyOriginCards(activeStockExplanation.value, formatDateTime));
const activeStockVersionHistoryItems = computed(() => {
  const stockID = activeStockView.value?.id;
  const items = stockID ? stockVersionHistoryMap.value[stockID] : [];
  if (Array.isArray(items) && items.length > 0) {
    return mapStrategyVersionHistory(items, formatDateTime);
  }
  return buildFallbackStrategyVersionHistory(activeStockExplanation.value, {
    reasonSummary: activeStockBase.value?.reason_summary || "",
    strategyVersion: activeStockBase.value?.strategy_version || "record",
    formatDateTime
  });
});
const activeSelectedStockVersionHistory = computed(() => {
  return activeStockHistoryCompare.value?.selectedItem || null;
});
const activeStockHistoryCompare = computed(() => {
  const items = activeStockVersionHistoryItems.value;
  return buildStrategyHistoryCompareState({
    historyItems: items,
    selectedKey: selectedStockVersionHistoryKey.value,
    fallbackItem: items[1] || items[0] || null,
    explanation: activeStockExplanation.value,
    selectedTitlePrefix: "当前对比 ",
    formatDateTime,
    fallbackRecordLabel: "record",
    upgradedText: "所选历史版本和当前 explanation 的结论都发生了变化。",
    reasonChangedText: "当前 explanation 对所选历史版本做了新的收敛。",
    versionChangedText: "核心结论延续，但版本已发生刷新。"
  });
});
const activeStockVersionDiff = computed(() => {
  const base = activeStockBase.value;
  const explanation = activeStockExplanation.value;
  if (!base || !explanation) {
    return null;
  }
  return buildStrategyVersionDiff({
    recordVersion: base.strategy_version || "record",
    recordReason: base.reason_summary || "",
    explanation,
    formatDateTime,
    fallbackRecordLabel: "record",
    upgradedText: "记录版本和当前 explanation 的结论都发生了变化。",
    reasonChangedText: "当前 explanation 对原始推荐理由做了新的收敛。",
    versionChangedText: "核心结论仍在，但解释版本已经更新。"
  });
});

const activeFuturesExplanation = computed(() => {
  const strategyID = activeFuturesView.value?.id;
  if (!strategyID) {
    return null;
  }
  return futuresExplanationMap.value[strategyID] || null;
});

const activeFuturesInsightSections = computed(() =>
  buildStrategyInsightSections(activeFuturesExplanation.value, activeFuturesView.value?.reason || "")
);

const activeFuturesExplanationCards = computed(() => {
  const explanation = activeFuturesExplanation.value || {};
  const evidenceCount = Array.isArray(explanation.evidence_cards) ? explanation.evidence_cards.length : 0;
  const agentCount = Number(explanation.workload_summary?.agent_count ?? 0);
  const scenarioCount = Number(explanation.workload_summary?.scenario_count ?? 0);
  return [
    {
      label: "种子输入",
      value: Array.isArray(explanation.seed_highlights) ? explanation.seed_highlights.length : 0,
      note: explanation.seed_summary || activeFuturesInsightSections.value.whyNow || "系统先处理合约种子，再收敛出可执行策略。"
    },
    {
      label: "证据卡片",
      value: evidenceCount,
      note: activeFuturesInsightSections.value.proofSource || "系统会把证据和多角色结论汇总后展示。"
    },
    {
      label: "评审覆盖",
      value: `${agentCount} 角 / ${scenarioCount} 景`,
      note: buildStrategyRiskBoundaryText(explanation, "当前未补更多风险边界。")
    }
  ];
});

const activeFuturesSupplyChainEntities = computed(() => {
  const explanation = activeFuturesExplanation.value || {};
  const allowed = new Set(["Commodity", "SupplyChainNode", "SpreadPair", "Index", "DeliveryPlace", "Warehouse", "Brand", "Grade"]);
  const typeLabels = {
    Commodity: "商品",
    SupplyChainNode: "商品链节点",
    SpreadPair: "价差对",
    Index: "指数",
    DeliveryPlace: "交割地",
    Warehouse: "仓库",
    Brand: "品牌",
    Grade: "等级"
  };
  const seen = new Set();
  return (Array.isArray(explanation.related_entities) ? explanation.related_entities : [])
    .map((item, index) => ({
      key: String(item?.entity_key || item?.label || `entity-${index}`),
      entityType: String(item?.entity_type || "").trim(),
      label: String(item?.label || item?.entity_key || "").trim()
    }))
    .filter((item) => item.entityType && item.label && allowed.has(item.entityType))
    .filter((item) => {
      const dedupeKey = `${item.entityType}:${item.label}`;
      if (seen.has(dedupeKey)) {
        return false;
      }
      seen.add(dedupeKey);
      return true;
    })
    .map((item) => ({
      ...item,
      typeLabel: typeLabels[item.entityType] || item.entityType
    }));
});
const activeFuturesSupplyChainNotes = computed(() => {
  const explanation = activeFuturesExplanation.value || {};
  const notes = Array.isArray(explanation.supply_chain_notes) ? explanation.supply_chain_notes : [];
  const deduped = [];
  const seen = new Set();
  notes.forEach((item) => {
    const text = String(item || "").trim();
    if (!text || seen.has(text)) {
      return;
    }
    seen.add(text);
    deduped.push(text);
  });
  return deduped.slice(0, 4);
});
const activeFuturesSupplyChainEvidenceCards = computed(() => {
  const cards = Array.isArray(activeFuturesExplanation.value?.evidence_cards) ? activeFuturesExplanation.value.evidence_cards : [];
  return cards
    .filter((item) => ["库存画像", "结构联动", "商品链"].includes(String(item?.title || "").trim()))
    .map((item, index) => ({
      key: `${item?.title || "card"}-${item?.value || index}`,
      title: item?.title || "证据",
      value: item?.value || "-",
      note: item?.note || "-"
    }));
});
const activeFuturesSupplyChainSummaryCards = computed(() => {
  const explanation = activeFuturesExplanation.value || {};
  const entities = activeFuturesSupplyChainEntities.value;
  const notes = activeFuturesSupplyChainNotes.value;
  return [
    {
      label: "结构联动摘要",
      value: explanation.structure_factor_summary || "待同步",
      note: notes[0] || "优先查看期限结构、价差与商品链的联动方向。"
    },
    {
      label: "库存画像摘要",
      value: explanation.inventory_factor_summary || "待同步",
      note: notes[1] || "结合仓库、品牌、等级与连续去库/累库节奏判断。"
    },
    {
      label: "商品链节点",
      value: entities.length ? entities.slice(0, 4).map((item) => item.label).join(" / ") : "待同步",
      note: notes[2] || "当前会把商品、仓库、品牌、等级和价差对一起纳入解释。"
    }
  ];
});

const activeFuturesProofTags = computed(() => buildListProofTags(activeFuturesExplanation.value, { limit: 4 }));
const activeFuturesSeedHighlights = computed(() => activeFuturesExplanation.value?.seed_highlights || []);
const activeFuturesScenarioCards = computed(() => activeFuturesExplanation.value?.simulations?.[0]?.scenarios || []);
const activeFuturesAgentOpinions = computed(() => activeFuturesExplanation.value?.agent_opinions || []);
const activeFuturesRiskCards = computed(() => buildStrategyRiskCards(activeFuturesExplanation.value));
const activeFuturesOriginCards = computed(() => buildStrategyOriginCards(activeFuturesExplanation.value, formatDateTime));
const activeFuturesVersionHistoryItems = computed(() => {
  const strategyID = activeFuturesView.value?.id;
  const items = strategyID ? futuresVersionHistoryMap.value[strategyID] : [];
  if (Array.isArray(items) && items.length > 0) {
    return mapStrategyVersionHistory(items, formatDateTime);
  }
  return buildFallbackStrategyVersionHistory(activeFuturesExplanation.value, {
    reasonSummary: activeFuturesBase.value?.reason_summary || "",
    strategyVersion: "futures-mvp-v1",
    formatDateTime
  });
});
const activeSelectedFuturesVersionHistory = computed(() => {
  return activeFuturesHistoryCompare.value?.selectedItem || null;
});
const activeFuturesHistoryCompare = computed(() => {
  const items = activeFuturesVersionHistoryItems.value;
  return buildStrategyHistoryCompareState({
    historyItems: items,
    selectedKey: selectedFuturesVersionHistoryKey.value,
    fallbackItem: items[1] || items[0] || null,
    explanation: activeFuturesExplanation.value,
    selectedTitlePrefix: "当前对比 ",
    formatDateTime,
    fallbackRecordLabel: "record",
    upgradedText: "所选历史版本和当前 explanation 的结论都发生了变化。",
    reasonChangedText: "当前 explanation 对所选期货历史版本做了新的收敛。",
    versionChangedText: "核心策略主线延续，但版本已发生刷新。"
  });
});
const activeFuturesVersionDiff = computed(() => {
  const base = activeFuturesBase.value;
  const insight = activeFuturesInsight.value || {};
  const strategy = insight.strategy || {};
  const explanation = activeFuturesExplanation.value;
  if (!base || !explanation) {
    return null;
  }
  return buildStrategyVersionDiff({
    recordVersion: strategy.strategy_version || base.strategy_version || "record",
    recordReason: strategy.reason_summary || base.reason_summary || "",
    explanation,
    formatDateTime,
    fallbackRecordLabel: "record",
    upgradedText: "记录版本和当前 explanation 的结论都发生了变化。",
    reasonChangedText: "当前 explanation 对原始期货策略理由做了新的收敛。",
    versionChangedText: "核心策略主线仍在，但解释版本已经更新。"
  });
});

const activeStockTrackingState = computed(() => {
  const base = activeStockBase.value;
  if (!base?.id) {
    return {
      label: "待同步",
      className: "normal",
      summary: "等待同步推荐状态。",
      nextAction: "刷新详情"
    };
  }

  const status = mapStatus(base.status);
  const rows = activeStockPerformanceRows.value.filter((item) => Number.isFinite(item.dailyRaw));
  const latest = rows[rows.length - 1] || null;
  const stats = activeStockPerformanceStats.value;

  let summary = "推荐已生成，等待更多跟踪数据。";
  let nextAction = "继续跟踪逻辑与风险边界";

  if (status.key === "HIT_TAKE_PROFIT") {
    summary = "推荐已达到止盈阶段，当前更适合回看兑现过程而不是继续追价。";
    nextAction = "整理复盘并关注是否有新一轮入场机会";
  } else if (status.key === "HIT_STOP_LOSS") {
    summary = "推荐已触发止损，当前重点是确认亏损控制是否执行到位。";
    nextAction = "停止主观加仓，回到历史档案做复盘";
  } else if (status.key === "INVALIDATED") {
    summary = "推荐逻辑已失效，当前更适合记录失效原因，不再沿用旧判断。";
    nextAction = "等待新的推荐逻辑或新的触发信号";
  } else if (rows.length > 0 && latest) {
    summary = `最近样本日 ${latest.date}，累计 ${latest.cumulativeReturn}，当前处于持续跟踪阶段。`;
    nextAction =
      stats && Number.isFinite(Number(stats.excess_return)) && Number(stats.excess_return) > 0
        ? "优先跟踪超额收益是否继续扩大"
        : "优先跟踪是否重新转强并站稳风险线";
  }

  return {
    label: status.label,
    className: status.className,
    summary,
    nextAction
  };
});

const activeStockRiskBoundaries = computed(() => {
  const base = activeStockBase.value;
  if (!base?.id) {
    return [];
  }
  const detail = stockDetailMap.value[base.id] || {};
  const invalidationCondition = buildStockInvalidationCondition(base, detail);
  return [
    {
      label: "止盈计划",
      value: detail.take_profit || "待补充",
      note: "优先分批兑现，不把推荐理解为无限持有。"
    },
    {
      label: "止损边界",
      value: detail.stop_loss || "待补充",
      note: "跌破风险线时，优先执行纪律，而不是继续找理由。"
    },
    {
      label: "失效条件",
      value: invalidationCondition,
      note: detail.risk_note || "当价格、量能或事件环境背离当前逻辑时，应停止沿用本推荐。"
    }
  ];
});

const activeStockTrackingTimeline = computed(() => {
  const base = activeStockBase.value;
  if (!base?.id) {
    return [];
  }
  return [
    {
      label: "推荐生效",
      value: formatDate(base.valid_from),
      note: `有效期 ${formatDateRange(base.valid_from, base.valid_to)}`
    },
    {
      label: "当前阶段",
      value: activeStockTrackingState.value.label,
      note: activeStockTrackingState.value.summary
    },
    {
      label: "下一动作",
      value: activeStockTrackingState.value.nextAction,
      note: activeStockRiskBoundaries.value[2]?.value || "继续观察风险边界"
    }
  ];
});

const activeStockPerformanceNote = computed(() => {
  const stats = activeStockPerformanceStats.value;
  if (!stats) {
    return "业绩仅作推荐跟踪参考，建议结合止盈止损和逻辑变化一起判断。";
  }
  const benchmarkLabel = stats.benchmark_symbol ? `${stats.benchmark_symbol} ` : "";
  const source = stats.benchmark_source || "基准来源未标注";
  return `当前展示的业绩与 ${benchmarkLabel}基准对比结果仅作推荐跟踪参考，来源说明：${source}。`;
});

const isActiveStockWatched = computed(() => {
  watchlistVersion.value;
  return Boolean(activeStockBase.value?.id) && isWatchedStock(activeStockBase.value.id);
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

const activeStockRelatedEvents = computed(() => {
  const items = activeStockExplanation.value?.related_events;
  return Array.isArray(items) ? items : [];
});

const activeStockEventEvidenceCards = computed(() =>
  buildStrategyEventEvidenceCards(activeStockExplanation.value, { limit: 3 })
);

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
  if (activeStockRelatedEvents.value.length > 0) {
    const latestEvent = activeStockRelatedEvents.value[0];
    result.push({
      label: "事件证据",
      value: `${activeStockRelatedEvents.value.length} 条`,
      note: `最近一条：${latestEvent.title || latestEvent.event_type || "审核事件"}`
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
  (rawFutures.value || []).map((item) => {
    const explanation = futuresExplanationMap.value[item.id] || null;
    const sections = buildStrategyInsightSections(explanation, item.reason_summary || "");
    return {
      id: item.id,
      contract: item.contract || "-",
      title: `${item.name || "-"} · ${mapDirection(item.direction)}`,
      note:
        sections.whyNow ||
        `风险 ${mapRisk(item.risk_level)}，仓位 ${item.position_range || "-"}，状态 ${mapStatus(item.status).label}`,
      proofTags: buildListProofTags(explanation, { limit: 3 }),
      supportingText: sections.proofSource || sections.riskBoundary || "",
      metaText: buildListMetaText(explanation)
    };
  })
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

const strategyAccessStage = computed(() => {
  if (isVIPUser.value) {
    return "VIP";
  }
  if (isLoggedIn.value) {
    return "REGISTERED";
  }
  return "VISITOR";
});

function resolveStrategyStage() {
  return strategyAccessStage.value;
}

function resolveRouteTarget(route) {
  if (!route) {
    return "";
  }
  if (typeof route === "string") {
    return route;
  }
  return route.path || route.name || "";
}

function trackStrategyExperiment(eventType, targetKey, metadata = {}) {
  return trackExperimentEvent({
    experimentKey: "strategy_membership_cta",
    variantKey: strategyExperimentVariant,
    eventType,
    pageKey: "strategy",
    targetKey,
    userStage: resolveStrategyStage(),
    metadata
  });
}

function navigateWithStrategyTracking(route, targetKey, options = {}) {
  const destination = resolveRouteTarget(route);
  const metadata = {
    destination,
    cta_label: options.ctaLabel || "",
    experiment_variant: strategyExperimentVariant,
    ...options.metadata
  };
  trackStrategyExperiment("CLICK", targetKey, metadata);
  if (options.upgradeIntent) {
    trackStrategyExperiment("UPGRADE_INTENT", targetKey, metadata);
  }
  if (options.rememberAttribution) {
    rememberExperimentAttributionSource({
      experimentKey: "strategy_membership_cta",
      variantKey: strategyExperimentVariant,
      pageKey: "strategy",
      targetKey,
      userStage: resolveStrategyStage(),
      metadata
    });
  }
  if (options.rememberPendingAttribution) {
    rememberPendingExperimentJourneySource({
      experimentKey: "strategy_membership_cta",
      variantKey: strategyExperimentVariant,
      pageKey: "strategy",
      targetKey,
      userStage: "VISITOR",
      redirectPath: options.redirectPath || "/strategies",
      metadata
    });
  }
  router.push(route);
}

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
      errors.push(`股票策略加载失败：${stockRes.reason?.message || "unknown error"}`);
    }

    if (futuresRes.status === "fulfilled" && Array.isArray(futuresRes.value?.items)) {
      rawFutures.value = futuresRes.value.items;
    } else if (futuresRes.status === "rejected") {
      errors.push(`期货策略加载失败：${futuresRes.reason?.message || "unknown error"}`);
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
      errors.push(`市场事件加载失败：${eventsRes.reason?.message || "unknown error"}`);
    }

    if (errors.length > 0) {
      errorMessage.value = errors.join("；");
    }

    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
    void hydrateStrategyListInsights();
  } catch (error) {
    errorMessage.value = error?.message || "加载失败";
  } finally {
    loading.value = false;
  }
}

async function hydrateStrategyListInsights() {
  const stockTargets = (rawStocks.value || [])
    .slice(0, 6)
    .map((item) => item?.id)
    .filter(Boolean);
  const futuresTargets = (rawFutures.value || [])
    .slice(0, 4)
    .map((item) => item?.id)
    .filter(Boolean);

  await Promise.allSettled([
    ...stockTargets.map((id) => loadActiveStockDetail(id, { silent: true })),
    ...futuresTargets.map((id) => loadFuturesInsight(id, { silent: true }))
  ]);
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
    promoteStrategyPostAuthAttribution();
  }
}

function promoteStrategyPostAuthAttribution() {
  if (!isLoggedIn.value) {
    return;
  }
  promotePendingExperimentJourneySources({
    experimentKey: "strategy_membership_cta",
    pageKey: "strategy",
    userStage: resolveStrategyStage(),
    metadata: {
      stock_count: stockRows.value.length,
      futures_count: futuresRows.value.length
    }
  });
}

async function loadFuturesInsight(id, options = {}) {
  const { silent = false, force = false } = options;
  if (!id || id.startsWith("fs_local_")) {
    if (!silent) {
      futuresInsightErrorMessage.value = "";
    }
    return;
  }
  const hasVersionHistory = Array.isArray(futuresVersionHistoryMap.value[id]);
  const shouldLoadHistory = force || !loadedFuturesVersionHistoryIDs.has(id) || !hasVersionHistory;
  if (!force && loadedFuturesInsightIDs.has(id) && futuresInsightMap.value[id] && !shouldLoadHistory) {
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
      if (insight?.explanation) {
        futuresExplanationMap.value = {
          ...futuresExplanationMap.value,
          [id]: insight.explanation
        };
      }
      loadedFuturesInsightIDs.add(id);
    }
    if (shouldLoadHistory) {
      const historyRes = await getFuturesStrategyVersionHistory(id);
      futuresVersionHistoryMap.value = {
        ...futuresVersionHistoryMap.value,
        [id]: Array.isArray(historyRes?.items) ? historyRes.items : []
      };
      loadedFuturesVersionHistoryIDs.add(id);
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

function selectFuturesVersionHistory(key) {
  selectedFuturesVersionHistoryKey.value = key || "";
}

function resetFuturesVersionHistorySelection() {
  selectedFuturesVersionHistoryKey.value = "";
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
  const hasVersionHistory = Array.isArray(stockVersionHistoryMap.value[id]);
  const shouldLoadInsight = force || !loadedStockInsightIDs.has(id) || !hasFramework || !hasStats;
  const shouldLoadDetail = force || !loadedStockDetailIDs.has(id) || !hasDetail;
  const shouldLoadPerformance = force || !loadedStockPerformanceIDs.has(id) || !hasPerformance;
  const shouldLoadRelatedNews = force || !loadedStockRelatedNewsIDs.has(id) || !hasRelatedNews;
  const shouldLoadVersionHistory = force || !loadedStockVersionHistoryIDs.has(id) || !hasVersionHistory;
  if (!shouldLoadInsight && !shouldLoadDetail && !shouldLoadPerformance && !shouldLoadRelatedNews && !shouldLoadVersionHistory) {
    return;
  }

  const nextDetailMap = { ...stockDetailMap.value };
  const nextPerformanceMap = { ...stockPerformanceMap.value };
  const nextBenchmarkMap = { ...stockBenchmarkMap.value };
  const nextRelatedNewsMap = { ...stockRelatedNewsMap.value };
  const nextScoreFrameworkMap = { ...stockScoreFrameworkMap.value };
  const nextPerformanceStatsMap = { ...stockPerformanceStatsMap.value };
  const nextExplanationMap = { ...stockExplanationMap.value };
  const nextVersionHistoryMap = { ...stockVersionHistoryMap.value };
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
      if (insight?.explanation) {
        nextExplanationMap[id] = insight.explanation;
      }
      loadedStockInsightIDs.add(id);
      insightLoaded = true;
    } catch (error) {
      errors.push(`推荐解释包同步失败：${parseErrorMessage(error)}`);
    }
  }

  if (shouldLoadVersionHistory) {
    try {
      const historyRes = await getStockRecommendationVersionHistory(id);
      nextVersionHistoryMap[id] = Array.isArray(historyRes?.items) ? historyRes.items : [];
      loadedStockVersionHistoryIDs.add(id);
    } catch (error) {
      errors.push(`版本历史同步失败：${parseErrorMessage(error)}`);
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
  stockExplanationMap.value = nextExplanationMap;
  stockVersionHistoryMap.value = nextVersionHistoryMap;

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

function selectStockVersionHistory(key) {
  selectedStockVersionHistoryKey.value = key || "";
}

function resetStockVersionHistorySelection() {
  selectedStockVersionHistoryKey.value = "";
}

function toggleActiveStockWatch() {
  const base = activeStockBase.value;
  if (!base?.id) {
    return;
  }
  if (isWatchedStock(base.id)) {
    removeWatchedStock(base.id);
  } else {
    const detail = stockDetailMap.value[base.id] || {};
    const explanation = stockExplanationMap.value[base.id] || {};
    const sections = buildStrategyInsightSections(explanation, base.reason_summary || "");
    saveWatchedStock({
      ...base,
      baseline_reason_summary: sections.whyNow || base.reason_summary || "",
      baseline_risk_level: base.risk_level || "",
      baseline_status: base.status || "",
      baseline_take_profit: detail.take_profit || "",
      baseline_stop_loss: detail.stop_loss || "",
      baseline_strategy_version: explanation.strategy_version || "",
      baseline_publish_id: explanation.publish_id || "",
      baseline_job_id: explanation.job_id || "",
      baseline_trade_date: explanation.trade_date || "",
      baseline_publish_version: explanation.publish_version || 0,
      baseline_agent_opinions: explanation.agent_opinions || []
    });
  }
  watchlistVersion.value += 1;
}

function handleWatchlistUpdate() {
  watchlistVersion.value += 1;
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

function buildListProofTags(explanation, options = {}) {
  return buildStrategyProofTags(explanation, {
    limit: Number(options.limit) > 0 ? Number(options.limit) : 3
  });
}

function buildListMetaText(explanation) {
  return buildStrategyMetaText(explanation, formatDateTime, { includeBatch: true, includeJob: false });
}

function buildStrategyRiskCards(explanation) {
  const boundary = buildStrategyRiskBoundaryText(explanation);
  const riskFlags = Array.isArray(explanation?.risk_flags) ? explanation.risk_flags : [];
  const invalidations = Array.isArray(explanation?.invalidations) ? explanation.invalidations : [];
  const cards = [];
  const seen = new Set();

  function pushCard(label, text, subtle = false) {
    const value = String(text || "").trim();
    if (!value || seen.has(value)) {
      return;
    }
    seen.add(value);
    cards.push({ label, text: value, subtle });
  }

  pushCard("风险边界", boundary);
  riskFlags.forEach((item) => pushCard("风险提醒", item));
  invalidations.forEach((item) => pushCard("失效条件", item, true));

  if (cards.length === 0) {
    cards.push({
      label: "风险边界",
      text: "当前未补更多风险边界。",
      subtle: false
    });
  }

  return cards;
}

function truncateText(value, maxLength = 16) {
  const text = String(value || "").trim();
  if (!text) {
    return "";
  }
  if (text.length <= maxLength) {
    return text;
  }
  return `${text.slice(0, maxLength)}...`;
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

function buildStockInvalidationCondition(base, detail) {
  const stopLoss = String(detail?.stop_loss || "").trim();
  const riskNote = String(detail?.risk_note || "").trim();
  if (stopLoss && riskNote) {
    return `${stopLoss}，且 ${riskNote}`;
  }
  if (stopLoss) {
    return stopLoss;
  }
  if (riskNote) {
    return riskNote;
  }
  const symbol = String(base?.symbol || base?.name || "推荐标的").trim();
  return `${symbol} 的趋势、量能或事件环境明显背离当前推荐逻辑`;
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
    return { key: source, label: "已发布", className: "good" };
  }
  if (source === "TRACKING") {
    return { key: source, label: "跟踪中", className: "good" };
  }
  if (source === "HIT_TAKE_PROFIT") {
    return { key: source, label: "已止盈", className: "good" };
  }
  if (source === "HIT_STOP_LOSS") {
    return { key: source, label: "已止损", className: "risky" };
  }
  if (source === "INVALIDATED") {
    return { key: source, label: "已失效", className: "risky" };
  }
  if (source === "REVIEWED") {
    return { key: source, label: "已复盘", className: "normal" };
  }
  if (source === "WATCH" || source === "DRAFT") {
    return { key: source, label: "观察", className: "watch" };
  }
  if (source === "REVIEW_PENDING") {
    return { key: source, label: "待审核", className: "watch" };
  }
  if (source === "DISABLED" || source === "EXPIRED") {
    return { key: source, label: "谨慎", className: "risky" };
  }
  return { key: source, label: source || "-", className: "normal" };
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

function goStrategyArchive() {
  navigateWithStrategyTracking("/archive", "focus_archive", {
    ctaLabel: "看历史档案"
  });
}

function goStrategyWatchlist() {
  navigateWithStrategyTracking("/watchlist", "focus_watchlist", {
    ctaLabel: "去我的关注"
  });
}

function goStrategyNews() {
  navigateWithStrategyTracking("/news", "focus_news", {
    ctaLabel: "去资讯页"
  });
}

function focusStock(id) {
  if (!id || activeStockID.value === id) {
    return;
  }
  activeStockID.value = id;
}

function focusFutures(id) {
  if (!id || activeFuturesID.value === id) {
    return;
  }
  activeFuturesID.value = id;
}

function syncRequestedStockFocus() {
  const requestedTarget = normalizeRouteValue(route.query.stock_id);
  if (!requestedTarget) {
    return;
  }
  const matchedID = findEntityIDByTarget(rawStocks.value, requestedTarget, ["id", "symbol"]);
  if (matchedID && activeStockID.value !== matchedID) {
    activeStockID.value = matchedID;
  }
}

function syncRequestedFuturesFocus() {
  const requestedTarget = normalizeRouteValue(route.query.futures_id);
  if (!requestedTarget) {
    return;
  }
  const matchedID = findEntityIDByTarget(rawFutures.value, requestedTarget, ["id", "contract"]);
  if (matchedID && activeFuturesID.value !== matchedID) {
    activeFuturesID.value = matchedID;
  }
}

function openActiveStockDetailDialog() {
  if (!activeStockView.value?.id) {
    return;
  }
  openStockDetail(activeStockView.value.id);
}

function buildActiveStockCommunityDraft() {
  const stock = activeStockView.value;
  const stockBase = activeStockBase.value;
  if (!stock?.id) {
    return null;
  }
  const snapshot = stock.name || `${stockBase?.symbol || ""} ${stockBase?.name || ""}`.trim();
  return {
    topicType: "STOCK",
    sort: "LATEST",
    entrySource: "strategy_stock",
    targetType: "STOCK",
    targetID: stockBase?.symbol || stock.id,
    targetSnapshot: snapshot,
    title: `${snapshot || "当前标的"} 当前值得继续跟踪吗`,
    stance: "WATCH"
  };
}

function openStrategyCommunity() {
  const draft = buildActiveStockCommunityDraft();
  router.push(draft ? buildCommunityListRoute(draft) : "/community");
}

function openStrategyDiscussionComposer() {
  const draft = buildActiveStockCommunityDraft();
  router.push(draft ? buildCommunityComposeRoute(draft) : "/community/new");
}

function openActiveFuturesDetailDialog() {
  if (!activeFuturesView.value?.id) {
    return;
  }
  openFuturesInsight(activeFuturesView.value.id);
}

function buildActiveFuturesCommunityDraft() {
  const futures = activeFuturesView.value;
  const futuresBase = activeFuturesBase.value;
  if (!futures?.id) {
    return null;
  }
  const snapshot = futures.name || `${futuresBase?.contract || ""} ${futuresBase?.name || ""}`.trim();
  return {
    topicType: "FUTURES",
    sort: "LATEST",
    entrySource: "strategy_futures",
    targetType: "FUTURES",
    targetID: futuresBase?.contract || futures.id,
    targetSnapshot: snapshot,
    title: `围绕 ${snapshot || "当前期货策略"} 继续判断`,
    stance: "WATCH"
  };
}

function openFuturesCommunity() {
  const draft = buildActiveFuturesCommunityDraft();
  router.push(draft ? buildCommunityListRoute(draft) : "/community");
}

function openFuturesDiscussionComposer() {
  const draft = buildActiveFuturesCommunityDraft();
  router.push(draft ? buildCommunityComposeRoute(draft) : "/community/new");
}

function parseErrorMessage(error) {
  if (!error) {
    return "unknown error";
  }
  const responseMessage =
    error?.response?.data?.message || error?.response?.data?.error || error?.response?.statusText;
  return responseMessage || error?.message || "unknown error";
}

function normalizeRouteValue(value) {
  return String(value || "").trim();
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
  () => isLoggedIn.value,
  () => {
    loadMembershipStage();
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
  [() => rawStocks.value, () => route.query.stock_id],
  () => {
    syncRequestedStockFocus();
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
  [() => rawFutures.value, () => route.query.futures_id],
  () => {
    syncRequestedFuturesFocus();
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

watch(
  [memberStageLoading, strategyAccessStage],
  ([loadingState]) => {
    if (loadingState) {
      return;
    }
    trackExperimentExposureOnce({
      experimentKey: "strategy_membership_cta",
      variantKey: strategyExperimentVariant,
      pageKey: "strategy",
      targetKey: "access_panel",
      userStage: resolveStrategyStage(),
      metadata: {
        stock_count: stockRows.value.length,
        futures_count: futuresRows.value.length
      }
    });
  },
  { immediate: true }
);

onMounted(() => {
  loadStrategies();
  loadMembershipStage();
  window.addEventListener(WATCHLIST_EVENT, handleWatchlistUpdate);
});

onBeforeUnmount(() => {
  window.removeEventListener(WATCHLIST_EVENT, handleWatchlistUpdate);
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
    radial-gradient(circle at 100% 0%, var(--color-focus-glow) 0%, transparent 34%),
    radial-gradient(circle at 0% 100%, rgba(184, 130, 48, 0.22) 0%, transparent 36%),
    rgba(255, 255, 255, 0.93);
}

.strategy-hero-copy {
  min-width: 0;
}

.strategy-hero-stats {
  grid-column: 1 / -1;
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

.hero-actions {
  display: flex;
  gap: 8px;
}

.api-state {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--color-pine-700);
}

.api-state.warning {
  color: var(--color-warning);
}

.strategy-focus-layout {
  display: grid;
  gap: 12px;
  grid-template-columns: minmax(0, 1.28fr) 360px;
}

.focus-detail-card,
.focus-side-card {
  padding: 14px;
}

.focus-detail-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.focus-detail-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.focus-detail-content {
  margin-top: 12px;
  display: grid;
  gap: 12px;
}

.focus-summary-panel {
  padding: 14px;
  display: grid;
  gap: 12px;
}

.focus-summary-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.focus-summary-head h3,
.side-highlight-box h3 {
  margin: 4px 0 0;
  font-size: 20px;
}

.focus-badge-row,
.focus-proof-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.focus-summary-text,
.side-highlight-text {
  margin: 0;
  font-size: 13px;
  line-height: 1.75;
  color: var(--color-text-sub);
}

.focus-kpi-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(4, minmax(0, 1fr));
}

.focus-kpi-grid article,
.side-mini-grid article {
  display: grid;
  gap: 4px;
}

.focus-kpi-grid p,
.side-mini-grid p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.focus-kpi-grid strong,
.side-mini-grid strong {
  font-size: 14px;
  color: var(--color-text-main);
}

.focus-side-stack {
  display: grid;
  gap: 12px;
  align-content: start;
}

.side-highlight-box {
  display: grid;
  gap: 10px;
}

.focus-link-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.focus-link-row .ghost-btn {
  flex: 1 1 0;
  min-width: 92px;
}

.side-mini-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.side-mini-grid--single {
  grid-template-columns: 1fr;
}

.side-event-title {
  margin: 0;
  font-size: 15px;
  line-height: 1.55;
  color: var(--color-text-main);
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
  border: 1px solid var(--color-border-soft);
  background: rgba(244, 248, 255, 0.8);
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
  display: grid;
  grid-template-columns: 1.4fr repeat(4, 0.6fr);
  align-items: center;
  gap: 10px;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.matrix-item:hover {
  border-color: var(--color-border-focus-soft);
}

.matrix-item.active {
  border-color: var(--color-border-focus-stronger);
  box-shadow: 0 8px 20px rgba(36, 87, 167, 0.13);
}

.matrix-main {
  min-width: 0;
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

.list-proof-tags {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.list-proof-tags span {
  border-radius: 999px;
  border: 1px solid var(--color-border-soft-strong);
  background: var(--color-surface-panel-soft);
  padding: 3px 8px;
  font-size: 11px;
  color: var(--color-pine-700);
}

.list-proof-tags.compact {
  margin-top: 7px;
}

.list-proof-note,
.cycle-proof-note,
.cycle-proof-meta {
  margin: 7px 0 0;
  font-size: 12px;
  line-height: 1.6;
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
}

.status.good {
  color: var(--color-pine-700);
  background: var(--color-surface-accent-soft);
}

.status.normal {
  color: var(--color-text-sub);
  background: var(--color-surface-panel-soft-subtle);
}

.status.watch {
  color: var(--color-warning);
  background: var(--color-surface-gold-soft);
}

.status.risky {
  color: var(--color-fall);
  background: var(--color-fall-soft);
}

.matrix-proof {
  grid-column: 1 / -1;
  margin-top: -2px;
  padding-top: 4px;
  border-top: 1px dashed var(--color-border-soft-muted);
  font-size: 11px;
  color: var(--color-text-sub);
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
  background: rgba(15, 23, 42, 0.38);
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
  border: 1px solid var(--color-border-soft-strong);
  background: rgba(244, 248, 255, 0.97);
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
  width: auto;
}

.stock-detail-head button:disabled {
  opacity: 0.68;
}

.close-btn {
  color: var(--color-text-sub);
}

.watch-btn {
  color: var(--color-accent);
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
  display: grid;
  gap: 4px;
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

.tracking-status-box {
  display: grid;
  gap: 10px;
}

.tracking-status-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.tracking-status-head p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.tracking-badge {
  min-width: auto;
}

.tracking-status-summary {
  margin: 0;
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.tracking-timeline {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.tracking-timeline article,
.risk-boundary-grid article {
  display: grid;
  gap: 4px;
}

.tracking-timeline p,
.risk-boundary-grid p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.tracking-timeline strong,
.risk-boundary-grid strong {
  font-size: 13px;
  color: var(--color-text-main);
}

.tracking-timeline span,
.risk-boundary-grid span {
  font-size: 12px;
  line-height: 1.55;
  color: var(--color-text-sub);
}

.risk-boundary-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.reason-support-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.reason-support-grid article {
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

.strategy-explanation-stack {
  display: grid;
  gap: 10px;
}

.strategy-version-box {
  display: grid;
  gap: 10px;
}

.strategy-version-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.strategy-version-grid article {
  display: grid;
  gap: 4px;
}

.strategy-version-grid p,
.strategy-version-grid strong,
.strategy-version-grid span {
  margin: 0;
}

.strategy-version-grid p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.strategy-version-grid strong {
  font-size: 14px;
  color: var(--color-text-main);
}

.strategy-version-grid span {
  font-size: 12px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.strategy-history-box {
  display: grid;
  gap: 10px;
}

.strategy-history-list {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.strategy-history-item {
  text-align: left;
  cursor: pointer;
}

.strategy-history-list article,
.strategy-history-item {
  display: grid;
  gap: 4px;
}

.strategy-history-list p,
.strategy-history-item p,
.strategy-history-list strong,
.strategy-history-item strong,
.strategy-history-list span,
.strategy-history-item span {
  margin: 0;
}

.strategy-history-list p {
  font-size: 12px;
  color: var(--color-text-sub);
}

.strategy-history-list strong,
.strategy-history-item strong {
  font-size: 14px;
  color: var(--color-text-main);
}

.strategy-history-list span,
.strategy-history-item span {
  font-size: 12px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.strategy-history-item.active {
  box-shadow: 0 0 0 1px var(--color-focus-ring);
}

.strategy-history-hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.strategy-history-head-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.strategy-history-reset {
  white-space: nowrap;
}

.strategy-explanation-box {
  display: grid;
  gap: 10px;
}

.explanation-summary,
.explanation-note {
  margin: 0;
  font-size: 13px;
  line-height: 1.75;
  color: var(--color-text-main);
}

.chip-group {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.scenario-grid {
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.scenario-item,
.agent-opinion-item,
.risk-flag-item {
  display: grid;
  gap: 4px;
}

.scenario-item p,
.agent-opinion-item p,
.risk-flag-item strong {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.scenario-item strong,
.agent-opinion-item strong {
  font-size: 14px;
  color: var(--color-text-main);
}

.scenario-item span,
.agent-opinion-item span,
.risk-flag-item span {
  font-size: 12px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.scenario-item em {
  font-style: normal;
  font-size: 12px;
  color: var(--color-text-sub);
}

.agent-opinion-list,
.risk-flag-list {
  display: grid;
  gap: 8px;
}

.risk-flag-item.subtle {
  background: var(--color-surface-panel-soft-strong);
}

.stock-news-box {
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
  min-width: 0;
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
  display: grid;
  gap: 8px;
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

.stock-performance-note {
  margin: 8px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: var(--color-text-sub);
}

.performance-table-wrap {
  margin-top: 8px;
}

.performance-table {
  min-width: 560px;
}

.performance-table .up {
  color: var(--color-rise);
}

.performance-table .down {
  color: var(--color-fall);
}

.empty-inline {
  margin-top: 8px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.detail-warning {
  margin: 0;
  font-size: 12px;
  color: var(--color-warning);
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
  display: grid;
  gap: 4px;
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
  display: grid;
  gap: 8px;
}

.futures-event-list {
  display: grid;
  gap: 8px;
}

.futures-event-item {
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
  border: 1px dashed var(--color-border-soft-heavy);
  background: var(--color-surface-panel-tint);
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
  white-space: nowrap;
}

.event-type-tabs button.active {
  border-color: var(--color-border-focus-medium);
  box-shadow: inset 0 0 0 1px var(--color-focus-fill);
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
  font-weight: 700;
}

.event-level.high {
  color: var(--color-accent);
  background: var(--color-surface-gold-soft);
}

.event-level.mid {
  color: var(--color-pine-700);
  background: var(--color-surface-accent);
}

.event-level.low {
  color: var(--color-text-sub);
  background: var(--color-surface-panel-soft-subtle);
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
  border: 1px solid var(--color-border-soft);
  background: rgba(244, 248, 255, 0.8);
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
  width: auto;
}

.detail-actions button:disabled {
  opacity: 0.7;
}

@media (max-width: 980px) {
  .strategy-focus-layout,
  .strategy-hero,
  .strategy-grid,
  .event-section {
    grid-template-columns: 1fr;
  }

  .focus-detail-head,
  .focus-summary-head {
    display: grid;
  }

  .strategy-hero {
    align-items: start;
  }

  .focus-detail-actions {
    justify-content: flex-start;
  }
}

@media (max-width: 760px) {
  .hero-actions,
  .focus-detail-actions {
    display: grid;
    grid-template-columns: 1fr;
  }

  .hero-actions > *,
  .focus-detail-actions > * {
    width: 100%;
  }

  .focus-badge-row,
  .tag-row {
    flex-wrap: nowrap;
    overflow-x: auto;
    scrollbar-width: none;
  }

  .focus-badge-row::-webkit-scrollbar,
  .tag-row::-webkit-scrollbar {
    display: none;
  }

  .focus-badge-row span,
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

  .matrix-proof {
    margin-top: 0;
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

  .focus-kpi-grid,
  .tracking-timeline,
  .risk-boundary-grid {
    grid-template-columns: 1fr;
  }

  .futures-detail-grid {
    grid-template-columns: 1fr;
  }

  .side-mini-grid,
  .reason-support-grid {
    grid-template-columns: 1fr;
  }

  .scenario-grid {
    grid-template-columns: 1fr;
  }

  .strategy-version-grid {
    grid-template-columns: 1fr;
  }

  .strategy-history-list {
    grid-template-columns: 1fr;
  }

  .strategy-history-head-actions,
  .stock-news-head {
    flex-direction: column;
    align-items: flex-start;
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

  .focus-link-row .ghost-btn {
    flex-basis: calc(50% - 4px);
  }
}
</style>
