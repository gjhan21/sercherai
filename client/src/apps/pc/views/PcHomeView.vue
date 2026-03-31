<template>
  <div class="home-page fade-up">
    <section class="hero-stage card">
      <div class="hero-copy">
        <div class="finance-pill-row">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">首页</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">今日决策</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">研报与策略协同</span>
        </div>
        <p class="hero-kicker">今日决策首页 · {{ todayDateText }}</p>
        <h1>
          {{ primaryStock ? `今天先看 ${primaryStock.symbol} ${primaryStock.name}` : "先看今日主推荐，再决定怎么做" }}
        </h1>
        <p class="hero-desc">
          {{
            primaryStock
              ? `${primaryStock.reason}。${marketPulse.summary}`
              : "首页汇总今日推荐、重点研报和历史样本。"
          }}
        </p>
        <div class="hero-actions">
          <button class="primary-btn finance-primary-btn" type="button" @click="goStrategyCenter">
            查看推荐详情
          </button>
          <button class="ghost-btn finance-ghost-btn" type="button" :disabled="!primaryStock" @click="togglePrimaryStockWatch">
            {{ primaryWatchActionText }}
          </button>
          <button class="ghost-btn finance-ghost-btn" type="button" :disabled="loading" @click="loadHomeData">
            {{ loading ? "同步中..." : "刷新今日决策" }}
          </button>
          <button class="ghost-btn finance-ghost-btn" type="button" @click="scrollToInsights">查看今日资讯</button>
        </div>
        <div class="hero-tags">
          <span v-for="tag in heroTags" :key="tag" class="finance-pill finance-pill-compact finance-pill-neutral">{{ tag }}</span>
        </div>
        <StatePanel
          :tone="homeStatus.tone"
          :eyebrow="homeStatus.label"
          :title="homeStatus.title"
          :description="homeStatus.desc"
          compact
        />
        <div class="home-hero-stats finance-hero-stat-grid">
          <article class="finance-hero-stat-card">
            <span>主推荐评分</span>
            <strong>{{ primaryStock?.score || "-" }}</strong>
            <p>先看主推荐强度，再决定是否进入策略详情页。</p>
          </article>
          <article class="finance-hero-stat-card">
            <span>观察股票</span>
            <strong>{{ decisionWatchlist.length }} 只</strong>
            <p>今天值得继续跟踪的对象会优先汇总在首页第一屏。</p>
          </article>
          <article class="finance-hero-stat-card">
            <span>资讯线索</span>
            <strong>{{ decisionNewsLinks.length }} 条</strong>
            <p>焦点研报和市场资讯会解释为什么今天先看这些标的。</p>
          </article>
          <article class="finance-hero-stat-card">
            <span>历史样本</span>
            <strong>{{ historyStocks.length }} 条</strong>
            <p>保留历史样本预览，方便把复盘重新接回今天的阅读链。</p>
          </article>
        </div>
      </div>

      <aside class="hero-panel">
        <div class="panel-head">
          <p>今日主推荐</p>
          <strong>{{ todayDateText }}</strong>
        </div>
        <div class="lead-stock-card">
          <div class="lead-stock-top">
            <div>
              <p class="lead-stock-symbol">{{ primaryStock?.symbol || "推荐池待更新" }}</p>
              <h2>{{ primaryStock?.name || "暂无今日主推荐" }}</h2>
            </div>
            <span class="lead-stock-risk">{{ primaryStock?.risk || "待确认" }}</span>
          </div>
          <p class="lead-stock-expected" :class="primaryStock?.direction || 'flat'">
            {{ primaryStock?.expected || "区间待更新" }}
          </p>
          <p class="lead-stock-note">
            {{ primaryStock?.takeProfit || "目标区间待补充" }} · {{ primaryStock?.stopLoss || "风控线待补充" }}
          </p>
          <p v-if="recommendationAccessLocked && recommendationAccessNote" class="lead-stock-why">
            {{ recommendationAccessNote }}
          </p>
          <p v-if="primaryStockWhyText" class="lead-stock-why">{{ primaryStockWhyText }}</p>
          <div v-if="primaryStockProofTags.length" class="lead-stock-proof-list">
            <span
              v-for="tag in primaryStockProofTags"
              :key="`primary-proof-${tag}`"
              class="finance-pill finance-pill-compact finance-pill-accent"
            >
              {{ tag }}
            </span>
          </div>
          <p v-if="primaryStockMetaText" class="lead-stock-meta">{{ primaryStockMetaText }}</p>
        </div>
        <ul class="task-list">
          <li v-for="task in tasks" :key="task.title">
            <span class="task-dot" :class="task.level" />
            <div>
              <p class="task-title">{{ task.title }}</p>
              <p class="task-note">{{ task.note }}</p>
            </div>
          </li>
        </ul>
      </aside>
    </section>

    <section class="decision-grid">
      <article class="card watch-card">
        <header class="decision-head">
          <div>
            <p class="search-kicker">股票 / 期货推荐区</p>
            <h2>查看今日主推荐、观察清单和期货方案。</h2>
          </div>
          <button type="button" class="search-jump-btn finance-mini-btn finance-mini-btn-soft" @click="goStrategyCenter">进入策略页</button>
        </header>
        <div v-if="decisionWatchlist.length" class="watch-list">
          <article v-for="item in decisionWatchlist" :key="item.id" class="watch-item finance-list-card">
            <div class="watch-item-top">
              <div>
                <p class="watch-rank">No.{{ item.rank }}</p>
                <h3>{{ item.symbol }} {{ item.name }}</h3>
              </div>
              <span class="watch-risk">{{ item.risk }}</span>
            </div>
            <p class="watch-expected" :class="item.direction">{{ item.expected }}</p>
            <p class="watch-reason">{{ item.note }}</p>
            <p v-if="item.whySelected" class="watch-why">{{ item.whySelected }}</p>
            <div v-if="item.proofTags.length" class="watch-proof-list">
              <span
                v-for="tag in item.proofTags"
                :key="`${item.id}-${tag}`"
                class="finance-pill finance-pill-compact finance-pill-info"
              >
                {{ tag }}
              </span>
            </div>
            <p class="watch-action">{{ item.action }}</p>
            <p v-if="item.meta" class="watch-meta">{{ item.meta }}</p>
          </article>
        </div>
        <StatePanel
          v-else
          tone="info"
          eyebrow="观察清单"
          :title="recommendationAccessLocked ? '登录后解锁今日观察清单' : '今日观察清单待生成'"
          :description="
            recommendationAccessLocked
              ? '当前先展示公开页面结构；登录后这里会自动补齐今日推荐与观察清单。'
              : '同步推荐数据后，这里会自动展示今天最值得先看的股票。'
          "
          compact
        />
      </article>

      <article class="card pulse-card research-card">
        <header class="decision-head">
          <div>
            <p class="search-kicker">今日研报解读</p>
            <h2>{{ featuredResearch?.title || "先看今日重点研报或资讯。" }}</h2>
          </div>
          <button type="button" class="search-jump-btn finance-mini-btn finance-mini-btn-soft" @click="openNewsModule">进入资讯中心</button>
        </header>
        <p class="pulse-summary">
          {{ featuredResearch?.summary || "优先展示一篇重点研报或资讯，帮助你快速把握今日重点。" }}
        </p>
        <div class="pulse-signals">
          <span class="finance-pill finance-pill-compact finance-pill-info">{{ featuredResearch?.category || "资讯" }}</span>
          <span
            class="finance-pill finance-pill-compact"
            :class="featuredResearch?.visibility === 'VIP' ? 'finance-pill-accent' : 'finance-pill-neutral'"
          >
            {{ featuredResearch?.visibility === "VIP" ? "VIP全文" : "公开内容" }}
          </span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">{{ featuredResearch?.time || newsUpdatedAt || "-" }}</span>
          <span class="finance-pill finance-pill-compact finance-pill-neutral">
            {{ primaryStock ? `关联 ${primaryStock.symbol}` : "关联主推荐" }}
          </span>
        </div>
        <div class="research-interpret-grid">
          <article v-for="item in featuredResearchCards" :key="item.title" class="research-interpret-item finance-card-surface">
            <strong>{{ item.title }}</strong>
            <p>{{ item.desc }}</p>
          </article>
        </div>
        <form class="insight-search" @submit.prevent="handleInsightSearch">
          <input
            v-model.trim="insightKeyword"
            type="text"
            maxlength="40"
            placeholder="输入关键词，快速筛选研报和市场资讯"
          />
          <button type="submit" class="finance-mini-btn finance-mini-btn-primary" :disabled="newsLoading">搜索</button>
          <button
            type="button"
            class="ghost finance-mini-btn finance-mini-btn-soft"
            :disabled="newsLoading || (!appliedInsightKeyword && !insightKeyword)"
            @click="clearInsightSearch"
          >
            清空
          </button>
        </form>
        <p class="search-tip finance-note-strip finance-note-strip-info">
          <template v-if="appliedInsightKeyword">当前关键词：{{ appliedInsightKeyword }}</template>
          <template v-else>未设置关键词，将优先展示最新研报和市场资讯</template>
        </p>
        <div v-if="researchSecondaryRows.length" class="pulse-news-list">
          <article
            v-for="item in researchSecondaryRows"
            :key="item.id"
            class="pulse-news-item finance-list-card"
            @click="openNewsModule"
          >
            <div class="pulse-news-meta">
              <span>{{ item.category }}</span>
              <strong>{{ item.visibility }}</strong>
            </div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.summary }}</p>
          </article>
        </div>
      </article>
    </section>

    <section class="community-home-section">
      <article class="card community-home-card">
        <header class="decision-head">
          <div>
            <p class="search-kicker">观点延伸区</p>
            <h2>看完主推荐和研报后，继续去讨论广场补充个人判断。</h2>
          </div>
          <button type="button" class="search-jump-btn finance-mini-btn finance-mini-btn-soft" @click="openCommunityPlaza">
            进入讨论广场
          </button>
        </header>
        <div class="community-home-grid">
          <article
            v-for="item in homeCommunityEntryCards"
            :key="item.key"
            class="finance-list-card finance-list-card-panel community-home-entry"
          >
            <div class="community-home-entry-top">
              <p>{{ item.kicker }}</p>
              <span class="finance-pill finance-pill-compact" :class="item.badgeClass">{{ item.badge }}</span>
            </div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.desc }}</p>
            <div class="community-home-entry-actions">
              <button
                type="button"
                class="finance-mini-btn finance-mini-btn-soft"
                :disabled="item.disabled"
                @click="openHomeCommunityList(item.key)"
              >
                先看讨论
              </button>
              <button
                type="button"
                class="finance-mini-btn finance-mini-btn-accent"
                :disabled="item.disabled"
                @click="openHomeCommunityComposer(item.key)"
              >
                发起观点
              </button>
            </div>
          </article>
        </div>
      </article>

      <aside class="card community-home-side">
        <header class="finance-copy-stack">
          <h2 class="section-title">讨论模块当前支持</h2>
          <p class="section-subtitle">只承接站内真实内容，不扩聊天室式即时刷屏。</p>
        </header>
        <div class="community-home-side-list">
          <article v-for="item in homeCommunitySupportRows" :key="item.title" class="finance-list-card finance-list-card-panel">
            <p>{{ item.title }}</p>
            <strong>{{ item.summary }}</strong>
            <span>{{ item.desc }}</span>
          </article>
        </div>
      </aside>
    </section>

    <section class="mobile-quick card">
      <button
        v-for="item in mobileQuickActions"
        :key="item.key"
        type="button"
        :disabled="item.disabled"
        @click="handleMobileQuickAction(item.key)"
      >
        <p>{{ item.title }}</p>
        <span>{{ item.desc }}</span>
      </button>
      <p
        class="mobile-quick-hint finance-note-strip"
        :class="loadError || newsError ? 'finance-note-strip-warning' : 'finance-note-strip-info'"
      >
        {{ mobileQuickHint }}
      </p>
    </section>

    <section id="insight-section" class="insight-section" :class="{ 'single-column': !shouldShowVipPromo }">
      <article class="card insight-card">
        <header class="insight-head">
          <div>
            <h2 class="section-title">研报 / 市场资讯</h2>
            <p class="section-subtitle">汇总新闻、研报和期刊，方便快速查看今日重点。</p>
          </div>
          <div class="insight-actions">
            <button type="button" class="finance-primary-btn" :disabled="newsLoading" @click="loadHomeInsights">
              {{ newsLoading ? "同步中..." : "立即刷新资讯" }}
            </button>
            <button type="button" class="ghost finance-ghost-btn" @click="openNewsModule">进入资讯中心</button>
          </div>
        </header>
        <p v-if="newsLoading" class="insight-status finance-note-strip finance-note-strip-info">正在同步资讯内容...</p>
        <p v-else-if="newsError" class="insight-status finance-note-strip finance-note-strip-warning">{{ newsError }}</p>
        <p v-else class="insight-status finance-note-strip finance-note-strip-info">
          资讯更新时间：{{ newsUpdatedAt || "-" }}
          <template v-if="appliedInsightKeyword"> · 当前关键词：{{ appliedInsightKeyword }}</template>
        </p>

        <div class="insight-tags" v-if="newsCategoryTags.length">
          <span v-for="tag in newsCategoryTags" :key="tag" class="finance-pill finance-pill-compact finance-pill-info">{{ tag }}</span>
        </div>

        <div v-if="newsHighlights.length > 0" class="insight-list">
          <article v-for="item in newsHighlights" :key="item.id" class="insight-item" @click="openNewsModule">
            <div class="insight-item-head">
              <span class="category">{{ item.category }}</span>
              <span class="time">{{ item.time }}</span>
            </div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.summary }}</p>
            <div class="insight-meta">
              <span>{{ item.source }}</span>
              <span class="vip-mark finance-pill finance-pill-compact" :class="{ active: item.visibility === 'VIP' }">
                {{ item.visibility === "VIP" ? "VIP专享" : "公开" }}
              </span>
            </div>
          </article>
        </div>
        <StatePanel
          v-else
          tone="info"
          eyebrow="资讯中心"
          title="暂无资讯内容"
          description="可以立即刷新资讯，或进入资讯中心查看更多内容。"
          compact
        >
          <template #actions>
            <button type="button" class="finance-primary-btn" @click="loadHomeInsights">立即刷新资讯</button>
            <button type="button" class="ghost finance-ghost-btn" @click="openNewsModule">进入资讯中心</button>
          </template>
        </StatePanel>
      </article>

      <aside v-if="shouldShowVipPromo" class="card vip-promo-card">
        <p class="promo-kicker">研报与深度权限</p>
        <h2>开通 VIP，解锁研报正文、附件和更完整的解释链。</h2>
        <p class="promo-price">{{ vipPromoPriceText }}</p>
        <p class="promo-note">{{ vipPromoDesc }}</p>
        <ul class="promo-benefits">
          <li>VIP 文章详情与附件下载</li>
          <li>更完整的推荐解释与事件提醒</li>
          <li>更高阅读与订阅配额</li>
        </ul>
        <button type="button" class="promo-btn finance-primary-btn" @click="goMembershipCenter('vip_promo_card')">
          {{ vipPromoCTA }}
        </button>
      </aside>
    </section>

    <section id="stock-section" class="stock-section">
      <article class="card history-card">
        <header class="section-head">
          <div>
            <h2 class="section-title">历史推荐档案预览</h2>
            <p class="section-subtitle">
              查看历史推荐结果和表现概况。基准表现为同周期模拟值。
            </p>
          </div>
          <button type="button" class="ghost-btn finance-ghost-btn section-ghost-btn" @click="goArchiveCenter">查看完整档案</button>
        </header>

        <div class="table-wrap finance-table-wrap" v-if="historyStocks.length">
          <table class="stock-table finance-data-table">
            <thead>
              <tr>
                <th>股票</th>
                <th>推荐日期</th>
                <th>综合评分</th>
                <th>风险等级</th>
                <th>历史涨跌幅</th>
                <th>基准表现</th>
                <th>超额收益</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in historyStocks" :key="`${row.code}-${row.date}`">
                <td>{{ row.code }}</td>
                <td>{{ row.date }}</td>
                <td>{{ row.score }}</td>
                <td>{{ row.risk }}</td>
                <td :class="getTrendClass(row.change)">{{ row.change }}</td>
                <td>{{ row.benchmark }}</td>
                <td :class="getTrendClass(row.alpha)">{{ row.alpha }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="empty-box finance-empty-box">暂无历史推荐数据</div>

        <div class="mobile-history-list">
          <article
            v-for="row in historyStocks"
            :key="`mobile-${row.code}-${row.date}`"
            class="mobile-history-item finance-list-card finance-list-card-compact"
          >
            <div class="top">
              <p>{{ row.code }}</p>
              <span>{{ row.date }}</span>
            </div>
            <div class="line">
              <span>评分 {{ row.score }}</span>
              <span>{{ row.risk }}</span>
            </div>
            <div class="line">
              <span :class="getTrendClass(row.change)">涨跌幅 {{ row.change }}</span>
              <span :class="getTrendClass(row.alpha)">超额 {{ row.alpha }}</span>
            </div>
          </article>
        </div>
      </article>

      <article class="card today-card">
        <header class="finance-copy-stack">
          <h2 class="section-title">今日观察清单补充</h2>
          <p class="section-subtitle">结合止盈止损提示，辅助判断今日重点关注标的。</p>
        </header>

        <div class="today-list" v-if="todayStocks.length">
          <article v-for="item in todayStocks" :key="item.code" class="today-item finance-list-card">
            <div class="today-main">
              <p class="code">{{ item.code }}</p>
              <p class="name">{{ item.name }}</p>
            </div>
            <p class="expect" :class="item.side">{{ item.expected }}</p>
            <p class="condition">{{ item.condition }}</p>
          </article>
        </div>
        <div v-else class="empty-box finance-empty-box">暂无今日推荐数据</div>
      </article>
    </section>

    <section class="futures-section" id="arbitrage-section">
      <article class="card arbitrage-card">
        <header class="finance-copy-stack">
          <h2 class="section-title">期货套利方案</h2>
          <p class="section-subtitle">展示进入点位、平仓点位、止损和仓位建议。</p>
        </header>

        <div class="arbitrage-list" v-if="arbitragePlans.length">
          <article v-for="plan in arbitragePlans" :key="plan.name" class="plan-item finance-list-card">
            <div class="plan-title">
              <p>{{ plan.name }}</p>
              <span>{{ plan.tag }}</span>
            </div>
            <div class="plan-grid">
              <div>
                <p>进入点位</p>
                <strong>{{ plan.entry }}</strong>
              </div>
              <div>
                <p>平仓点位</p>
                <strong>{{ plan.close }}</strong>
              </div>
              <div>
                <p>止损点位</p>
                <strong>{{ plan.stop }}</strong>
              </div>
              <div>
                <p>建议仓位</p>
                <strong>{{ plan.position }}</strong>
              </div>
            </div>
          </article>
        </div>
        <div v-else class="empty-box finance-empty-box">暂无期货套利计划</div>
      </article>

      <aside class="card method-card">
        <header class="finance-copy-stack">
          <h2 class="section-title">操作方法</h2>
          <p class="section-subtitle">按固定步骤执行，减少临盘决策失误。</p>
        </header>
        <ol class="method-list">
          <li v-for="step in methods" :key="step.title">
            <p class="step-title">{{ step.title }}</p>
            <p class="step-desc">{{ step.desc }}</p>
          </li>
        </ol>
        <p class="risk-tip finance-note-strip finance-note-strip-warning">
          风险提示：正式实盘前仍需结合行情、撮合规则和风控阈值确认。
        </p>
      </aside>
    </section>

    <section class="home-search-banner card">
      <div class="search-banner-head">
        <div>
          <p class="search-kicker">后续操作</p>
          <h2>看完推荐和研报后，可继续查看关注、历史档案和会员服务。</h2>
        </div>
        <button type="button" class="search-jump-btn finance-mini-btn finance-mini-btn-soft" @click="goMembershipCenter('action_banner_membership')">
          {{ vipPromoCTA }}
        </button>
      </div>
      <p class="search-tip finance-note-strip finance-note-strip-info">
        {{
          isVIPUser
            ? "你已拥有会员权益，可继续查看完整推荐、研报正文和会员状态。"
            : "可先从今日主推荐、历史样本和重点研报开始使用。"
        }}
      </p>
      <div class="action-pill-list">
        <button type="button" class="action-pill finance-toggle-btn" @click="goStrategyCenter">查看完整推荐逻辑</button>
        <button type="button" class="action-pill finance-toggle-btn" @click="goWatchlistCenter">进入我的关注</button>
        <button type="button" class="action-pill finance-toggle-btn" @click="goArchiveCenter">查看历史推荐档案</button>
        <button type="button" class="action-pill finance-toggle-btn" @click="openNewsModule">查看资讯详情</button>
      </div>
    </section>

    <section class="conversion-grid">
      <article class="card conversion-card">
        <header class="decision-head">
          <div>
            <p class="search-kicker">权益分层</p>
            <h2>{{ conversionHeadline }}</h2>
          </div>
          <button type="button" class="search-jump-btn finance-mini-btn finance-mini-btn-soft" @click="handlePrimaryConversionAction">
            {{ primaryConversionActionText }}
          </button>
        </header>
        <p class="search-tip finance-note-strip finance-note-strip-info">{{ conversionSubline }}</p>
        <div class="conversion-lanes">
          <article
            v-for="item in conversionStageCards"
            :key="item.key"
            class="conversion-lane"
            :class="{ active: item.active, unlocked: item.unlocked }"
          >
            <div class="conversion-lane-head">
              <p>{{ item.title }}</p>
              <span
                class="finance-pill finance-pill-compact"
                :class="item.active ? 'finance-pill-accent' : item.unlocked ? 'finance-pill-info' : 'finance-pill-neutral'"
              >
                {{ item.badge }}
              </span>
            </div>
            <strong>{{ item.summary }}</strong>
            <p>{{ item.desc }}</p>
          </article>
        </div>
      </article>

      <article class="card cadence-card">
        <header class="decision-head">
          <div>
            <p class="search-kicker">内容节奏</p>
            <h2>让用户每天都有固定回访理由。</h2>
          </div>
          <button type="button" class="search-jump-btn finance-mini-btn finance-mini-btn-soft" @click="goArchiveCenter">看历史节奏</button>
        </header>
        <div class="cadence-list">
          <article v-for="item in contentCadenceRows" :key="item.time" class="cadence-item finance-list-card">
            <p class="cadence-time">{{ item.time }}</p>
            <div>
              <strong>{{ item.title }}</strong>
              <p>{{ item.note }}</p>
            </div>
            <span
              class="cadence-access finance-pill finance-pill-compact"
              :class="
                item.accessClass === 'vip'
                  ? 'finance-pill-accent'
                  : item.accessClass === 'registered'
                    ? 'finance-pill-info'
                    : 'finance-pill-neutral'
              "
            >
              {{ item.access }}
            </span>
          </article>
        </div>
      </article>
    </section>
  </div>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import StatePanel from "../../../components/StatePanel.vue";
import {
  getFuturesGuidance,
  getStockRecommendationDetail,
  getStockRecommendationInsight,
  getStockRecommendationPerformance,
  listFuturesArbitrage,
  listStockRecommendations
} from "../../../api/market";
import { listNewsArticles, listNewsCategories } from "../../../api/news";
import { getMembershipQuota, listMembershipProducts } from "../../../api/membership";
import { useClientAuth } from "../../../lib/client-auth";
import { shouldUseDemoFallback } from "../../../lib/fallback-policy";
import {
  rememberExperimentAttributionSource,
  rememberPendingExperimentJourneySource,
  promotePendingExperimentJourneySources
} from "../../../lib/growth-analytics";
import { getExperimentVariant } from "../../../lib/growth-experiments";
import {
  buildStrategyConfidenceCalibrationSummary,
  buildStrategyInsightSections,
  buildStrategyMetaText,
  buildStrategyProofTags,
  buildStrategyThesisCardRows,
  buildStrategyWatchSignalRows,
  firstMeaningfulStrategyText
} from "../../../lib/strategy-version";
import {
  buildCommunityComposeRoute,
  buildCommunityListRoute
} from "../../../lib/community-entry-links";
import { buildProfileModuleRedirectPath, buildProfileModuleRoute } from "../../../lib/profile-modules";
import { WATCHLIST_EVENT, isWatchedStock, saveWatchedStock, removeWatchedStock } from "../../../lib/watchlist";

const fallbackStockRecommendations = [
  {
    id: "sr_local_001",
    symbol: "600036.SH",
    name: "招商银行",
    score: 91.2,
    risk_level: "MEDIUM",
    valid_from: "2026-02-10T09:00:00+08:00",
    reason_summary: "量价结构共振"
  },
  {
    id: "sr_local_002",
    symbol: "300750.SZ",
    name: "宁德时代",
    score: 88.5,
    risk_level: "MEDIUM",
    valid_from: "2026-02-12T09:00:00+08:00",
    reason_summary: "趋势延续，资金净流入"
  },
  {
    id: "sr_local_003",
    symbol: "601012.SH",
    name: "隆基绿能",
    score: 82.4,
    risk_level: "HIGH",
    valid_from: "2026-02-17T09:00:00+08:00",
    reason_summary: "反弹确认阶段"
  },
  {
    id: "sr_local_004",
    symbol: "000858.SZ",
    name: "五粮液",
    score: 86.8,
    risk_level: "LOW",
    valid_from: "2026-02-19T09:00:00+08:00",
    reason_summary: "估值修复 + 防御属性"
  }
];

const fallbackDetails = {
  sr_local_001: { take_profit: "上涨6%-10%分批止盈", stop_loss: "跌破 33.80 止损" },
  sr_local_002: { take_profit: "上涨5%-9%分批止盈", stop_loss: "跌破 139.20 止损" },
  sr_local_003: { take_profit: "上涨3%-6%分批止盈", stop_loss: "跌破 21.50 止损" },
  sr_local_004: { take_profit: "上涨4%-7%分批止盈", stop_loss: "跌破 145.00 止损" }
};

const fallbackPerformances = {
  sr_local_001: [
    { date: "2026-02-11", return: 0.018 },
    { date: "2026-02-12", return: 0.011 },
    { date: "2026-02-13", return: 0.012 }
  ],
  sr_local_002: [
    { date: "2026-02-13", return: 0.009 },
    { date: "2026-02-14", return: 0.013 },
    { date: "2026-02-15", return: 0.011 }
  ],
  sr_local_003: [
    { date: "2026-02-18", return: -0.011 },
    { date: "2026-02-19", return: -0.008 },
    { date: "2026-02-20", return: 0.003 }
  ],
  sr_local_004: [
    { date: "2026-02-20", return: 0.010 },
    { date: "2026-02-21", return: 0.007 },
    { date: "2026-02-22", return: 0.012 }
  ]
};

const fallbackArbitragePlans = [
  {
    id: "arb_local_001",
    type: "CALENDAR",
    contract_a: "IF 主连",
    contract_b: "IC 主连",
    entry_point: 186,
    exit_point: 132,
    stop_point: 208,
    percentile: 0.86,
    status: "WATCH"
  },
  {
    id: "arb_local_002",
    type: "CROSS",
    contract_a: "IH 主连",
    contract_b: "IF 主连",
    entry_point: -76,
    exit_point: -42,
    stop_point: -94,
    percentile: 0.79,
    status: "WATCH"
  }
];

const fallbackGuidance = {
  contract: "IF 主连",
  guidance_direction: "LONG_SPREAD",
  position_level: "LIGHT",
  entry_range: "120-150",
  take_profit_range: "80-100",
  stop_loss_range: "165-180",
  risk_level: "MEDIUM",
  invalid_condition: "相关性失效",
  valid_to: "2026-03-01T15:00:00+08:00"
};

const fallbackNewsHighlights = [
  {
    id: "news_local_001",
    category: "新闻",
    time: "2026/02/28 09:30:00",
    title: "盘前资金风格切换，银行与高股息板块走强。",
    summary: "市场风险偏好回落，早盘更关注防御属性板块轮动。",
    source: "编辑部",
    visibility: "PUBLIC"
  },
  {
    id: "news_local_002",
    category: "研报",
    time: "2026/02/28 08:55:00",
    title: "机构上调两融余额敏感行业评级。",
    summary: "研报建议关注景气向上赛道的估值切换窗口。",
    source: "研究团队",
    visibility: "VIP"
  },
  {
    id: "news_local_003",
    category: "期刊",
    time: "2026/02/27 21:10:00",
    title: "周度策略期刊：波动率管理与仓位纪律。",
    summary: "强调先控制回撤，再放大胜率，维持组合弹性。",
    source: "策略周刊",
    visibility: "PUBLIC"
  }
];

const fallbackVipProducts = [
  {
    id: "mp_local_vip_month",
    name: "VIP月卡",
    price: 99,
    duration_days: 30,
    status: "ACTIVE",
    member_level: "VIP1"
  }
];

const useDemoFallback = shouldUseDemoFallback();

const router = useRouter();
const { isLoggedIn } = useClientAuth();
const homeMembershipExperimentVariant = getExperimentVariant("home_membership_entry", ["default"]);

const loading = ref(false);
const loadError = ref("");
const lastUpdatedAt = ref("");
const newsLoading = ref(false);
const newsError = ref("");
const newsUpdatedAt = ref("");
const insightKeyword = ref("");
const appliedInsightKeyword = ref("");
const newsCategoryTags = ref([]);
const newsHighlights = ref(useDemoFallback ? [...fallbackNewsHighlights] : []);
const vipStateLoading = ref(false);
const isVIPUser = ref(false);
const vipOffer = ref(useDemoFallback ? fallbackVipProducts[0] : null);
const recommendationAccessLocked = ref(false);
const recommendationAccessNote = ref("");

const rawStockRecommendations = ref(useDemoFallback ? fallbackStockRecommendations : []);
const stockDetailMap = ref(useDemoFallback ? { ...fallbackDetails } : {});
const stockPerformanceMap = ref(useDemoFallback ? { ...fallbackPerformances } : {});
const stockExplanationMap = ref(
  useDemoFallback ? buildFallbackStockExplanationMap(fallbackStockRecommendations) : {}
);
const rawArbitragePlans = ref(useDemoFallback ? fallbackArbitragePlans : []);
const futuresGuidance = ref(useDemoFallback ? fallbackGuidance : null);
const watchedPrimaryStock = ref(false);

const todayDateText = computed(() =>
  new Date().toLocaleDateString("zh-CN", { year: "numeric", month: "2-digit", day: "2-digit" })
);

const primaryStock = computed(() => {
  const item = rawStockRecommendations.value[0];
  if (!item) {
    return null;
  }

  const detail = stockDetailMap.value[item.id] || {};
  const historyReturn = calcCumulativeReturn(stockPerformanceMap.value[item.id] || []);
  const expected = inferExpectedRange(detail.take_profit, historyReturn);

  return {
    id: item.id,
    symbol: item.symbol || item.id || "-",
    name: item.name || "未命名标的",
    score: formatScore(item.score),
    risk: mapRiskLevel(item.risk_level),
    reason: item.reason_summary || "满足当前策略触发条件",
    expected,
    direction: inferExpectedSide(expected),
    takeProfit: detail.take_profit || "目标区间待补充",
    stopLoss: detail.stop_loss || "止损线待补充",
    position: item.position_range || "建议轻仓跟踪"
  };
});

const primaryStockExplanation = computed(() => {
  const id = primaryStock.value?.id;
  if (!id) {
    return null;
  }
  return stockExplanationMap.value[id] || null;
});

const primaryStockInsightSections = computed(() =>
  buildStrategyInsightSections(primaryStockExplanation.value, primaryStock.value?.reason || "")
);

const primaryStockWhyText = computed(() => {
  const explanation = primaryStockExplanation.value;
  const activeThesis = buildStrategyThesisCardRows(explanation, "active", { limit: 1 })[0];
  return firstMeaningfulStrategyText([activeThesis?.summary, primaryStockInsightSections.value.whyNow]);
});

const primaryStockProofTags = computed(() =>
  buildStrategyProofTags(primaryStockExplanation.value, { includeVersion: true })
);

const primaryStockMetaText = computed(() =>
  buildStrategyMetaText(primaryStockExplanation.value, formatDateTime, { includeBatch: false, includeJob: false })
);

const decisionWatchlist = computed(() =>
  rawStockRecommendations.value.slice(0, 4).map((item, index) => {
    const detail = stockDetailMap.value[item.id] || {};
    const explanation = stockExplanationMap.value[item.id] || null;
    const sections = buildStrategyInsightSections(explanation, item.reason_summary || "等待更多触发信号确认");
    const historyReturn = calcCumulativeReturn(stockPerformanceMap.value[item.id] || []);
    const expected = inferExpectedRange(detail.take_profit, historyReturn);
    const activeThesis = buildStrategyThesisCardRows(explanation, "active", { limit: 1 })[0];
    const watchSignal = buildStrategyWatchSignalRows(explanation, { limit: 1 })[0];
    const calibration = buildStrategyConfidenceCalibrationSummary(explanation);

    return {
      id: item.id,
      rank: index + 1,
      symbol: item.symbol || item.id || "-",
      name: item.name || "未命名标的",
      risk: mapRiskLevel(item.risk_level),
      expected,
      direction: inferExpectedSide(expected),
      note: firstMeaningfulStrategyText([activeThesis?.summary, sections.whyNow, item.reason_summary, "等待更多触发信号确认"]),
      whySelected: sections.proofSource || "",
      proofTags: buildStrategyProofTags(explanation, { limit: 3, includeVersion: true }),
      action:
        firstMeaningfulStrategyText([watchSignal?.trigger, calibration?.summary, sections.riskBoundary]) ||
        (detail.stop_loss ? `风险线：${detail.stop_loss}` : `观察日：${formatDate(item.valid_from)}`),
      meta: buildStrategyMetaText(explanation, formatDateTime, { includeBatch: false, includeJob: false })
    };
  })
);

const historyStocks = computed(() =>
  rawStockRecommendations.value.slice(0, 6).map((item) => {
    const points = stockPerformanceMap.value[item.id] || [];
    const historyReturn = calcCumulativeReturn(points);
    const benchmarkReturn = estimateBenchmarkReturn(historyReturn);
    const alphaReturn =
      historyReturn === null || benchmarkReturn === null ? null : historyReturn - benchmarkReturn;

    return {
      code: item.symbol || item.id || "-",
      date: formatDate(item.valid_from),
      score: formatScore(item.score),
      risk: mapRiskLevel(item.risk_level),
      change: formatPercent(historyReturn),
      benchmark: formatPercent(benchmarkReturn),
      alpha: formatPercent(alphaReturn),
      rawChange: historyReturn
    };
  })
);

const todayStocks = computed(() =>
  rawStockRecommendations.value.slice(0, 3).map((item) => {
    const detail = stockDetailMap.value[item.id] || {};
    const historyReturn = calcCumulativeReturn(stockPerformanceMap.value[item.id] || []);
    const expected = inferExpectedRange(detail.take_profit, historyReturn);
    const side = inferExpectedSide(expected);

    return {
      code: item.symbol || item.id || "-",
      name: item.name || "未命名标的",
      expected,
      side,
      condition: detail.stop_loss
        ? `条件：${detail.stop_loss}`
        : `条件：${item.reason_summary || "满足策略触发条件"}`
    };
  })
);

const arbitragePlans = computed(() =>
  rawArbitragePlans.value.slice(0, 4).map((item) => ({
    name: `${item.contract_a || "-"} - ${item.contract_b || "-"}`,
    tag: `${mapArbitrageType(item.type)} · ${mapArbitrageStatus(item.status)}`,
    entry: pointWithRule(item.entry_point, "触发"),
    close: pointWithRule(item.exit_point, "止盈"),
    stop: pointWithRule(item.stop_point, "止损"),
    position: suggestPosition(item.percentile, item.risk_level)
  }))
);

const methods = computed(() => {
  const guidance = futuresGuidance.value;
  if (!guidance) {
    return [
      { title: "步骤 1：确认触发条件", desc: "仅在价差进入触发区间且成交量同步放大时建仓。" },
      { title: "步骤 2：分两笔进场", desc: "首笔 60%，二次确认后补足剩余 40%，避免单点误差。" },
      { title: "步骤 3：预设平仓与止损", desc: "下单同时写入平仓/止损点位，不临盘手动决策。" },
      { title: "步骤 4：复盘偏差原因", desc: "记录滑点、执行延迟和信号有效性，用于次日优化。" }
    ];
  }

  return [
    {
      title: "步骤 1：确认进入区间",
      desc: `当价差进入 ${guidance.entry_range || "-"}，且方向为 ${mapGuidanceDirection(
        guidance.guidance_direction
      )} 时执行首笔建仓。`
    },
    {
      title: "步骤 2：设置平仓目标",
      desc: `平仓参考区间 ${guidance.take_profit_range || "-"}，达到目标后分批止盈。`
    },
    {
      title: "步骤 3：执行止损纪律",
      desc: `当价差进入 ${guidance.stop_loss_range || "-"} 时立即止损，不做主观加仓。`
    },
    {
      title: "步骤 4：校验失效条件",
      desc: `若出现“${guidance.invalid_condition || "模型失效"}”，立即暂停策略并复盘。`
    }
  ];
});

const decisionNewsLinks = computed(() =>
  newsHighlights.value.slice(0, 3).map((item) => ({
    id: item.id,
    category: item.category || "资讯",
    title: item.title || "未命名资讯",
    summary: item.summary || "暂无摘要",
    visibility: item.visibility === "VIP" ? "VIP" : "公开"
  }))
);

const featuredResearch = computed(() => {
  if (newsHighlights.value.length === 0) {
    return null;
  }
  return (
    newsHighlights.value.find((item) => isResearchLike(item)) ||
    newsHighlights.value[0] ||
    null
  );
});

const featuredResearchCards = computed(() => {
  const item = featuredResearch.value;
  const stockLabel = primaryStock.value
    ? `${primaryStock.value.symbol} ${primaryStock.value.name}`
    : "今日主推荐";

  return [
    {
      title: "这篇内容在说什么",
      desc:
        item?.summary ||
        "优先展示一篇重点内容，帮助你快速把握今日重点。"
    },
    {
      title: "它对今天推荐的影响",
      desc: item
        ? `${item.category || "这条内容"}可作为 ${stockLabel} 的补充参考，帮助确认当前逻辑是否延续。`
        : "先看内容是否继续支持主推荐逻辑。"
    },
    {
      title: "今天应该怎么用",
      desc:
        item?.visibility === "VIP"
          ? "先看导读，再进入资讯页查看全文和附件；确认后回策略页继续判断。"
          : "先看导读，再进入资讯页查看正文；确认后回策略页或我的关注继续跟踪。"
    }
  ];
});

const researchSecondaryRows = computed(() =>
  newsHighlights.value
    .filter((item) => item.id !== featuredResearch.value?.id)
    .slice(0, 3)
    .map((item) => ({
      id: item.id,
      category: item.category || "资讯",
      title: item.title || "未命名资讯",
      summary: item.summary || "暂无摘要",
      visibility: item.visibility === "VIP" ? "VIP" : "公开"
    }))
);

const homeCommunityEntryCards = computed(() => [
  {
    key: "stock",
    kicker: "围绕主推荐",
    badge: "股票",
    badgeClass: "finance-pill-info",
    title: primaryStock.value ? `${primaryStock.value.symbol} ${primaryStock.value.name}` : "今日主推荐待同步",
    desc: primaryStock.value
      ? "先看股票主题，再决定是否围绕当前主推荐补充自己的依据、风险边界和后续动作。"
      : "同步主推荐后，这里会自动承接到股票讨论入口。",
    disabled: !primaryStock.value
  },
  {
    key: "research",
    kicker: "围绕焦点研报",
    badge: featuredResearch.value?.category || "资讯",
    badgeClass: featuredResearch.value?.visibility === "VIP" ? "finance-pill-accent" : "finance-pill-neutral",
    title: featuredResearch.value?.title || "焦点研报待同步",
    desc: featuredResearch.value
      ? "先进入资讯讨论区，再决定是否围绕当前研报继续发起结构化观点主题。"
      : "同步资讯后，这里会自动承接到研报和资讯讨论入口。",
    disabled: !featuredResearch.value?.id
  }
]);

const homeCommunitySupportRows = [
  {
    title: "当前能力",
    summary: "主题帖、评论、点赞、收藏、举报",
    desc: "讨论区先做结构化表达，不做实时聊天室。"
  },
  {
    title: "首页承接",
    summary: "主推荐和焦点研报可直接带上下文进入",
    desc: "进入后列表会先展示对应主题类型，并保留发帖预填信息。"
  },
  {
    title: "后续回看",
    summary: "个人中心已支持我的讨论入口",
    desc: "适合收盘后回看自己发过的主题和参与过的评论。"
  }
];

const marketPulse = computed(() => {
  const scores = rawStockRecommendations.value
    .map((item) => Number(item?.score))
    .filter((value) => Number.isFinite(value));
  const avgScore = scores.length > 0 ? scores.reduce((sum, value) => sum + value, 0) / scores.length : null;
  const highRiskCount = rawStockRecommendations.value.filter(
    (item) => String(item?.risk_level || "").toUpperCase() === "HIGH"
  ).length;

  if (Number.isFinite(avgScore) && avgScore >= 88 && highRiskCount <= 1) {
    return {
      headline: "今天偏向精选出手，但先做强弱判断。",
      summary: "推荐池评分较集中，适合先盯主推荐和前 4 只观察清单，再看资讯是否继续强化逻辑。",
      signals: [
        `主推荐评分 ${primaryStock.value?.score || "-"}`,
        `观察清单 ${decisionWatchlist.value.length} 只`,
        `相关资讯 ${decisionNewsLinks.value.length} 条`
      ]
    };
  }

  if (highRiskCount >= 2) {
    return {
      headline: "今天更适合先观察，不宜凭情绪追涨。",
      summary: "推荐池中高风险标的偏多，适合先看主推荐与风险边界，再决定是否行动。",
      signals: [
        `高风险标的 ${highRiskCount} 只`,
        `历史命中 ${calcHitRate(historyStocks.value)}`,
        `相关资讯 ${decisionNewsLinks.value.length} 条`
      ]
    };
  }

  return {
    headline: "今天以跟踪确认优先，先看节奏再做动作。",
    summary: "先盯今日主推荐与观察清单，再通过资讯和历史档案确认逻辑是否延续。",
    signals: [
      `主推荐风险 ${primaryStock.value?.risk || "-"}`,
      `历史命中 ${calcHitRate(historyStocks.value)}`,
      `观察清单 ${decisionWatchlist.value.length} 只`
    ]
  };
});

const primaryWatchActionText = computed(() => (watchedPrimaryStock.value ? "移出我的关注" : "保存到我的关注"));

const heroTags = computed(() => [
  `主推荐 ${primaryStock.value?.risk || "待确认"}`,
  `历史命中 ${calcHitRate(historyStocks.value)}`,
  `相关资讯 ${decisionNewsLinks.value.length} 条`
]);

const mobileQuickActions = computed(() => [
  {
    key: "stocks",
    title: "看推荐",
    desc: `${decisionWatchlist.value.length} 只待确认`,
    disabled: false
  },
  {
    key: "insights",
    title: "看资讯",
    desc: `${decisionNewsLinks.value.length} 条线索`,
    disabled: false
  },
  {
    key: "refresh",
    title: "刷新首页",
    desc: loadError.value || newsError.value ? "点击重试" : "点击刷新",
    disabled: loading.value || newsLoading.value || vipStateLoading.value
  }
]);
const mobileQuickHint = computed(() => {
  if (loading.value || newsLoading.value || vipStateLoading.value) {
    return "正在同步推荐、资讯和会员状态，请稍候...";
  }
  if (recommendationAccessLocked.value) {
    return recommendationAccessNote.value || "游客模式下先看公开资讯；登录后会自动补齐今日推荐与期货方案。";
  }
  if (loadError.value || newsError.value) {
    return "部分数据加载失败，建议先点击“刷新首页”再继续浏览。";
  }
  return "移动端可直接跳到推荐、资讯，或一键刷新今日决策。";
});
const homeStatus = computed(() => {
  if (loading.value || newsLoading.value || vipStateLoading.value) {
    return {
      tone: "info",
      label: "同步中",
      title: "首页关键数据正在更新",
      desc: "推荐、资讯和会员状态会在同步完成后自动刷新。"
    };
  }
  if (recommendationAccessLocked.value) {
    return {
      tone: "info",
      label: "游客模式",
      title: "登录后解锁今日推荐与期货方案",
      desc: recommendationAccessNote.value || "当前先展示公开资讯与页面结构，登录后会自动补齐决策数据。"
    };
  }
  const issues = [loadError.value, newsError.value].filter(Boolean);
  if (issues.length > 0) {
    return {
      tone: "warning",
      label: "需处理",
      title: "首页部分数据加载失败",
      desc: issues.join("；")
    };
  }
  return {
    tone: "success",
    label: "准备就绪",
    title: "推荐、资讯和会员状态已同步",
    desc: `首页 ${lastUpdatedAt.value || "-"} · 资讯 ${newsUpdatedAt.value || "-"}`
  };
});

const shouldShowVipPromo = computed(() => !vipStateLoading.value && !isVIPUser.value);

const vipPromoPriceText = computed(() => {
  const offer = vipOffer.value;
  const price = Number(offer?.price);
  if (!Number.isFinite(price) || price <= 0) {
    return "VIP权益套餐";
  }
  const duration = Number(offer?.duration_days);
  if (Number.isFinite(duration) && duration >= 360) {
    return `¥${price.toFixed(0)} / 年`;
  }
  if (Number.isFinite(duration) && duration >= 85) {
    return `¥${price.toFixed(0)} / 季`;
  }
  if (Number.isFinite(duration) && duration >= 28) {
    return `¥${price.toFixed(0)} / 月`;
  }
  return `¥${price.toFixed(0)} 起`;
});

const vipPromoDesc = computed(() => {
  const offer = vipOffer.value;
  if (!offer?.name) {
    return "会员到期自动失效，续费后立即恢复权益。";
  }
  const duration = Number(offer.duration_days);
  if (Number.isFinite(duration) && duration > 0) {
    return `${offer.name}，权益有效期 ${duration} 天。`;
  }
  return `${offer.name}，权益即时生效。`;
});

const vipPromoCTA = computed(() => (isLoggedIn.value ? "去会员中心开通 VIP" : "登录后开通 VIP"));

const conversionStageKey = computed(() => {
  if (isVIPUser.value) {
    return "VIP";
  }
  if (isLoggedIn.value) {
    return "REGISTERED";
  }
  return "VISITOR";
});

const conversionHeadline = computed(() => {
  if (conversionStageKey.value === "VIP") {
    return "你已开通会员，可直接使用完整内容与复盘服务。";
  }
  if (conversionStageKey.value === "REGISTERED") {
    return "你已完成注册，可继续保存关注并查看更多内容。";
  }
  return "从公开内容开始，逐步建立自己的观察节奏。";
});

const conversionSubline = computed(() => {
  if (conversionStageKey.value === "VIP") {
    return "建议固定查看推荐、盘中变化和收盘复盘。";
  }
  if (conversionStageKey.value === "REGISTERED") {
    return "可保存关注、查看更多历史样本，并按需升级会员。";
  }
  return "可先查看主推荐、公开资讯和历史样本。";
});

const primaryConversionActionText = computed(() => {
  if (conversionStageKey.value === "VIP") {
    return "进入会员中心";
  }
  if (conversionStageKey.value === "REGISTERED") {
    return "去会员中心";
  }
  return "先登录保存关注";
});

const conversionStageCards = computed(() => [
  {
    key: "VISITOR",
    title: "游客",
    badge: conversionStageKey.value === "VISITOR" ? "当前阶段" : "入口层",
    summary: "查看今日主推荐、公开资讯和最近历史档案。",
    desc: "先熟悉今日内容，再决定是否继续使用。",
    unlocked: true,
    active: conversionStageKey.value === "VISITOR"
  },
  {
    key: "REGISTERED",
    title: "注册用户",
    badge: conversionStageKey.value === "REGISTERED" ? "当前阶段" : isLoggedIn.value ? "已解锁" : "登录后解锁",
    summary: "保存关注、查看更多历史样本和推荐逻辑。",
    desc: "适合开始建立自己的观察清单。",
    unlocked: isLoggedIn.value,
    active: conversionStageKey.value === "REGISTERED"
  },
  {
    key: "VIP",
    title: "会员",
    badge: conversionStageKey.value === "VIP" ? "当前阶段" : isVIPUser.value ? "已解锁" : "升级后解锁",
    summary: "查看完整推荐池、盘中变化提醒、收盘复盘和更高配额。",
    desc: "适合需要更完整内容和更高效率判断的用户。",
    unlocked: isVIPUser.value,
    active: conversionStageKey.value === "VIP"
  }
]);

const contentCadenceRows = computed(() => [
  {
    time: "08:30",
    title: "今日重点推荐",
    note: "游客也能先看到主推荐，快速判断今天该先跟哪条线。",
    access: "公开",
    accessClass: "public"
  },
  {
    time: "11:30",
    title: "午盘变化提醒",
    note: isLoggedIn.value ? "登录后可把午盘变化和自选关注一起查看。" : "登录后可以把午盘变化和关注列表一起查看。",
    access: isLoggedIn.value ? "已解锁" : "登录可见",
    accessClass: isLoggedIn.value ? "registered" : "locked"
  },
  {
    time: "15:30",
    title: "收盘复盘",
    note: isVIPUser.value ? "会员可直接复盘命中、失效和次日观察点。" : "会员重点内容，帮助用户把当天判断沉淀成明天动作。",
    access: isVIPUser.value ? "会员已解锁" : "VIP核心",
    accessClass: isVIPUser.value ? "vip" : "locked"
  },
  {
    time: "周末",
    title: "下周关注方向",
    note: "把历史档案、资讯变化和跟踪状态汇总成下周清单，形成固定节奏。",
    access: isLoggedIn.value ? "注册可追踪" : "登录后解锁",
    accessClass: isLoggedIn.value ? "registered" : "locked"
  }
]);

const tasks = computed(() => {
  const firstStock = primaryStock.value;
  const firstNews = decisionNewsLinks.value[0];
  return [
    {
      title: "09:35 确认主推荐强弱",
      note: `优先确认 ${firstStock?.symbol || "首个推荐标的"} 是否继续满足触发条件`,
      level: "high"
    },
    {
      title: "11:30 跟踪资讯变化",
      note: `重点核对“${firstNews?.title || "最新资讯"}”是否强化当前推荐逻辑`,
      level: "mid"
    },
    {
      title: "14:50 执行收盘复核",
      note: "记录主推荐是否延续、失效，形成明日复盘输入",
      level: "low"
    }
  ];
});

async function loadHomeData() {
  loading.value = true;
  loadError.value = "";
  recommendationAccessLocked.value = false;
  recommendationAccessNote.value = "";

  if (!isLoggedIn.value) {
    recommendationAccessLocked.value = true;
    recommendationAccessNote.value = "股票推荐需要登录后查看；期货方案需要登录后查看，当前先展示公开资讯与页面结构。";
    rawStockRecommendations.value = useDemoFallback ? fallbackStockRecommendations : [];
    rawArbitragePlans.value = useDemoFallback ? fallbackArbitragePlans : [];
    stockDetailMap.value = useDemoFallback ? { ...fallbackDetails } : {};
    stockPerformanceMap.value = useDemoFallback ? { ...fallbackPerformances } : {};
    stockExplanationMap.value = useDemoFallback ? buildFallbackStockExplanationMap(fallbackStockRecommendations) : {};
    futuresGuidance.value = useDemoFallback ? fallbackGuidance : null;
    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
    loading.value = false;
    return;
  }

  const errors = [];
  const accessHints = [];
  try {
    const [stockResult, arbitrageResult] = await Promise.allSettled([
      listStockRecommendations({ page: 1, page_size: 8 }),
      listFuturesArbitrage({ page: 1, page_size: 8 })
    ]);

    if (stockResult.status === "fulfilled" && Array.isArray(stockResult.value?.items)) {
      if (stockResult.value.items.length > 0) {
        rawStockRecommendations.value = stockResult.value.items;
      }
      await hydrateStockDetails(rawStockRecommendations.value.slice(0, 6));
    } else if (stockResult.status === "rejected") {
      if (!isLoggedIn.value && isAuthBlockedError(stockResult.reason)) {
        accessHints.push("股票推荐需要登录后查看");
      } else {
        errors.push(`股票推荐加载失败：${stockResult.reason?.message || "unknown error"}`);
      }
    }

    if (arbitrageResult.status === "fulfilled" && Array.isArray(arbitrageResult.value?.items)) {
      if (arbitrageResult.value.items.length > 0) {
        rawArbitragePlans.value = arbitrageResult.value.items;
      }
      await hydrateFuturesGuidance(rawArbitragePlans.value[0]?.contract_a);
    } else if (arbitrageResult.status === "rejected") {
      if (!isLoggedIn.value && isAuthBlockedError(arbitrageResult.reason)) {
        accessHints.push("期货方案需要登录后查看");
      } else {
        errors.push(`期货套利加载失败：${arbitrageResult.reason?.message || "unknown error"}`);
      }
    }

    if (accessHints.length > 0) {
      recommendationAccessLocked.value = true;
      recommendationAccessNote.value = `${accessHints.join("；")}，当前先展示公开资讯与页面结构。`;
    }
    if (errors.length > 0) {
      loadError.value = errors.join("；");
    }
    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  } catch (error) {
    loadError.value = error?.message || "数据请求失败";
  } finally {
    loading.value = false;
  }
}

async function loadHomeInsights(options = {}) {
  newsLoading.value = true;
  newsError.value = "";
  const keyword = normalizeInsightKeyword(options.keyword ?? insightKeyword.value);
  const query = {
    page: 1,
    page_size: 8
  };
  if (keyword) {
    query.keyword = keyword;
  }
  const [categoryResult, newsResult] = await Promise.allSettled([
    listNewsCategories(),
    listNewsArticles(query)
  ]);
  const errors = [];
  const categoryMap = {};
  if (categoryResult.status === "fulfilled") {
    const items = Array.isArray(categoryResult.value?.items) ? categoryResult.value.items : [];
    newsCategoryTags.value = items
      .slice()
      .sort((a, b) => Number(a?.sort || 0) - Number(b?.sort || 0))
      .slice(0, 6)
      .map((item) => item.name)
      .filter(Boolean);
    items.forEach((item) => {
      if (item?.id) {
        categoryMap[item.id] = item.name || "资讯";
      }
    });
  } else {
    errors.push(`资讯分类加载失败：${categoryResult.reason?.message || "unknown error"}`);
  }

  if (newsResult.status === "fulfilled") {
    const items = Array.isArray(newsResult.value?.items) ? newsResult.value.items : [];
    const mapped = items.map((item) => mapNewsHighlight(item, categoryMap)).filter(Boolean);
    if (mapped.length > 0) {
      newsHighlights.value = mapped;
    } else if (!useDemoFallback) {
      newsHighlights.value = [];
    }
    appliedInsightKeyword.value = keyword;
  } else {
    errors.push(`资讯内容加载失败：${newsResult.reason?.message || "unknown error"}`);
  }

  newsError.value = errors.join("；");
  newsUpdatedAt.value = formatDateTime(new Date().toISOString());
  newsLoading.value = false;
}

async function loadVIPState() {
  vipStateLoading.value = true;
  try {
    if (!isLoggedIn.value) {
      isVIPUser.value = false;
      if (!useDemoFallback) {
        vipOffer.value = null;
      }
      return;
    }
    const [quotaResult, productsResult] = await Promise.allSettled([
      getMembershipQuota(),
      listMembershipProducts({ status: "ACTIVE", page: 1, page_size: 20 })
    ]);

    if (quotaResult.status === "fulfilled" && quotaResult.value) {
      isVIPUser.value = resolveVIPState(quotaResult.value);
    } else {
      isVIPUser.value = false;
    }

    if (productsResult.status === "fulfilled" && Array.isArray(productsResult.value?.items)) {
      const sorted = productsResult.value.items
        .filter((item) => String(item?.status || "").toUpperCase() === "ACTIVE")
        .sort((a, b) => Number(a?.price || 0) - Number(b?.price || 0));
      if (sorted.length > 0) {
        vipOffer.value = sorted[0];
      } else if (!useDemoFallback) {
        vipOffer.value = null;
      }
    } else if (!useDemoFallback) {
      vipOffer.value = null;
    }
  } finally {
    vipStateLoading.value = false;
    promoteHomePostAuthAttribution();
  }
}

async function hydrateStockDetails(items) {
  if (!Array.isArray(items) || items.length === 0) {
    return;
  }

  const nextDetailMap = { ...stockDetailMap.value };
  const nextPerfMap = { ...stockPerformanceMap.value };
  const nextExplanationMap = { ...stockExplanationMap.value };

  await Promise.all(
    items.map(async (item) => {
      if (!item?.id) {
        return;
      }
      const [detailRes, perfRes, insightRes] = await Promise.allSettled([
        getStockRecommendationDetail(item.id),
        getStockRecommendationPerformance(item.id),
        getStockRecommendationInsight(item.id)
      ]);

      if (detailRes.status === "fulfilled" && detailRes.value) {
        nextDetailMap[item.id] = detailRes.value;
      }
      if (perfRes.status === "fulfilled" && Array.isArray(perfRes.value?.points)) {
        nextPerfMap[item.id] = perfRes.value.points;
      }
      if (insightRes.status === "fulfilled" && insightRes.value?.explanation) {
        nextExplanationMap[item.id] = insightRes.value.explanation;
      } else if (!nextExplanationMap[item.id]) {
        nextExplanationMap[item.id] = buildFallbackStockExplanation(item);
      }
    })
  );

  stockDetailMap.value = nextDetailMap;
  stockPerformanceMap.value = nextPerfMap;
  stockExplanationMap.value = nextExplanationMap;
}

async function hydrateFuturesGuidance(contract) {
  if (!contract) {
    return;
  }
  try {
    const guidance = await getFuturesGuidance(contract);
    if (guidance) {
      futuresGuidance.value = guidance;
    }
  } catch {
    // guidance 接口失败时保留当前数据
  }
}

function scrollToArbitrage() {
  scrollToSection("arbitrage-section");
}

function scrollToInsights() {
  scrollToSection("insight-section");
}

function scrollToStocks() {
  scrollToSection("stock-section");
}

function scrollToSection(id) {
  const target = document.getElementById(id);
  if (!target) {
    return;
  }
  target.scrollIntoView({ behavior: "smooth", block: "start" });
}

async function handleMobileQuickAction(key) {
  if (key === "stocks") {
    scrollToStocks();
    return;
  }
  if (key === "insights") {
    scrollToInsights();
    return;
  }
  if (key === "refresh") {
    await refreshMobileQuickData();
  }
}

async function refreshMobileQuickData() {
  const keyword = normalizeInsightKeyword(appliedInsightKeyword.value || insightKeyword.value);
  await Promise.all([loadHomeData(), loadHomeInsights({ keyword }), loadVIPState()]);
}

function openNewsModule() {
  const keyword = normalizeInsightKeyword(appliedInsightKeyword.value);
  if (keyword) {
    router.push({ path: "/news", query: { keyword } });
    return;
  }
  router.push("/news");
}

function openCommunityPlaza() {
  router.push("/community");
}

function goStrategyCenter() {
  router.push("/strategies");
}

function goArchiveCenter() {
  router.push("/archive");
}

function goWatchlistCenter() {
  router.push(buildProfileModuleRoute("watchlist"));
}

function buildHomeStockCommunityDraft() {
  const item = primaryStock.value;
  if (!item?.id) {
    return null;
  }
  const snapshot = `${item.symbol || ""} ${item.name || ""}`.trim();
  return {
    topicType: "STOCK",
    sort: "LATEST",
    entrySource: "home_stock",
    targetType: "STOCK",
    targetID: item.symbol || item.id,
    targetSnapshot: snapshot,
    title: `${snapshot || item.name} 当前值得继续跟踪吗`,
    stance: "WATCH"
  };
}

function buildHomeResearchCommunityDraft() {
  const item = featuredResearch.value;
  if (!item?.id) {
    return null;
  }
  const title = String(item.title || "").trim();
  return {
    topicType: "NEWS",
    sort: "LATEST",
    entrySource: "home_research",
    targetType: "NEWS_ARTICLE",
    targetID: item.id,
    targetSnapshot: title,
    title: title ? `围绕《${title}》继续讨论` : "围绕当前资讯继续讨论",
    stance: "WATCH"
  };
}

function openHomeCommunityList(key) {
  const draft = key === "research" ? buildHomeResearchCommunityDraft() : buildHomeStockCommunityDraft();
  router.push(draft ? buildCommunityListRoute(draft) : "/community");
}

function openHomeCommunityComposer(key) {
  const draft = key === "research" ? buildHomeResearchCommunityDraft() : buildHomeStockCommunityDraft();
  router.push(draft ? buildCommunityComposeRoute(draft) : "/community/new");
}

function refreshWatchState() {
  watchedPrimaryStock.value = Boolean(primaryStock.value?.id) && isWatchedStock(primaryStock.value.id);
}

function togglePrimaryStockWatch() {
  const item = rawStockRecommendations.value.find((entry) => entry.id === primaryStock.value?.id);
  if (!item) {
    return;
  }
  if (isWatchedStock(item.id)) {
    removeWatchedStock(item.id);
  } else {
    const detail = stockDetailMap.value[item.id] || {};
    const explanation = stockExplanationMap.value[item.id] || {};
    const sections = buildStrategyInsightSections(explanation, item.reason_summary || "");
    saveWatchedStock({
      ...item,
      baseline_reason_summary: sections.whyNow || item.reason_summary || "",
      baseline_risk_level: item.risk_level || "",
      baseline_status: item.status || "",
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
  refreshWatchState();
}

function resolveHomeExperimentStage() {
  if (isVIPUser.value) {
    return "VIP";
  }
  if (isLoggedIn.value) {
    return "REGISTERED";
  }
  return "VISITOR";
}

function rememberHomeMembershipEntry(targetKey, metadata = {}) {
  rememberExperimentAttributionSource({
    experimentKey: "home_membership_entry",
    variantKey: homeMembershipExperimentVariant,
    pageKey: "home",
    targetKey,
    userStage: resolveHomeExperimentStage(),
    metadata: {
      primary_stock_id: primaryStock.value?.id || "",
      primary_stock_symbol: primaryStock.value?.symbol || "",
      ...metadata
    }
  });
}

function rememberHomePendingMembershipEntry(targetKey, metadata = {}) {
  rememberPendingExperimentJourneySource({
    experimentKey: "home_membership_entry",
    variantKey: homeMembershipExperimentVariant,
    pageKey: "home",
    targetKey,
    userStage: "VISITOR",
    redirectPath: "/home",
    metadata: {
      primary_stock_id: primaryStock.value?.id || "",
      primary_stock_symbol: primaryStock.value?.symbol || "",
      ...metadata
    }
  });
}

function promoteHomePostAuthAttribution() {
  if (!isLoggedIn.value) {
    return;
  }
  promotePendingExperimentJourneySources({
    experimentKey: "home_membership_entry",
    pageKey: "home",
    userStage: resolveHomeExperimentStage(),
    metadata: {
      primary_stock_id: primaryStock.value?.id || "",
      primary_stock_symbol: primaryStock.value?.symbol || ""
    }
  });
}

function goMembershipCenter(targetKey = "default_membership_entry") {
  if (isLoggedIn.value) {
    rememberHomeMembershipEntry(targetKey, {
      destination: "/membership"
    });
    router.push("/membership");
    return;
  }
  rememberHomePendingMembershipEntry(targetKey, {
    destination_after_auth: "/home"
  });
  router.push({ path: "/auth", query: { redirect: "/home" } });
}

function goAuthCenter(redirect = buildProfileModuleRedirectPath("watchlist")) {
  router.push({ path: "/auth", query: { redirect } });
}

function handlePrimaryConversionAction() {
  if (conversionStageKey.value === "VIP") {
    goMembershipCenter();
    return;
  }
  if (conversionStageKey.value === "REGISTERED") {
    goMembershipCenter();
    return;
  }
  goAuthCenter(buildProfileModuleRedirectPath("watchlist"));
}

function resolveVIPState(quota) {
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

function mapNewsHighlight(item, categoryMap = {}) {
  if (!item?.id) {
    return null;
  }
  const visibility = String(item.visibility || "").toUpperCase();
  return {
    id: item.id,
    category: categoryMap[item.category_id] || resolveNewsCategory(item),
    time: formatDateTime(item.published_at || item.created_at || item.updated_at),
    title: item.title || "未命名资讯",
    summary: summarizeText(item.summary || item.content),
    source: item.source || "资讯中心",
    visibility: visibility === "VIP" ? "VIP" : "PUBLIC"
  };
}

function resolveNewsCategory(item) {
  const text = `${item?.category_name || ""} ${item?.title || ""}`;
  if (/(研报|研究|券商)/i.test(text)) {
    return "研报";
  }
  if (/(期刊|周刊|月刊)/i.test(text)) {
    return "期刊";
  }
  return "新闻";
}

function summarizeText(value) {
  const text = String(value || "")
    .replace(/<[^>]*>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
  if (!text) {
    return "暂无摘要";
  }
  if (text.length <= 56) {
    return text;
  }
  return `${text.slice(0, 56)}...`;
}

async function handleInsightSearch() {
  const keyword = normalizeInsightKeyword(insightKeyword.value);
  if (!keyword) {
    await clearInsightSearch();
    return;
  }
  await loadHomeInsights({ keyword });
  scrollToInsights();
}

async function clearInsightSearch() {
  insightKeyword.value = "";
  await loadHomeInsights({ keyword: "" });
  scrollToInsights();
}

function normalizeInsightKeyword(value) {
  return String(value || "")
    .replace(/\s+/g, " ")
    .trim();
}

function isAuthBlockedError(error) {
  const status = Number(error?.code || error?.original?.response?.status || 0);
  const message = String(error?.message || error?.original?.response?.data?.message || "")
    .replace(/\s+/g, " ")
    .trim()
    .toLowerCase();
  return status === 401 || message.includes("authorization") || message.includes("unauthorized");
}

function calcCumulativeReturn(points) {
  if (!Array.isArray(points) || points.length === 0) {
    return null;
  }
  let acc = 1;
  let hasValid = false;
  points.forEach((point) => {
    const value = Number(point?.return);
    if (Number.isFinite(value)) {
      acc *= 1 + value;
      hasValid = true;
    }
  });
  if (!hasValid) {
    return null;
  }
  return acc - 1;
}

function estimateBenchmarkReturn(value) {
  if (!Number.isFinite(value)) {
    return null;
  }
  const bench = value * 0.58;
  return Math.max(Math.min(bench, 0.25), -0.25);
}

function inferExpectedRange(takeProfitText, historyReturn) {
  const rangeFromTakeProfit = parsePercentRange(takeProfitText);
  if (rangeFromTakeProfit) {
    return rangeFromTakeProfit;
  }

  if (Number.isFinite(historyReturn)) {
    const low = historyReturn * 0.6;
    const high = historyReturn * 1.2;
    return `${formatPercent(low)} ~ ${formatPercent(high)}`;
  }

  return "区间待更新";
}

function isResearchLike(item) {
  const text = `${item?.category || ""} ${item?.title || ""} ${item?.summary || ""}`;
  return /(研报|研究|深度|report|insight|forecast)/i.test(text);
}

function inferExpectedSide(expected) {
  if (expected.startsWith("+")) {
    return "up";
  }
  if (expected.startsWith("-")) {
    return "down";
  }
  return "flat";
}

function parsePercentRange(text) {
  if (!text || typeof text !== "string") {
    return "";
  }
  const matched = text.match(/-?\d+(\.\d+)?%/g);
  if (!matched || matched.length === 0) {
    return "";
  }
  if (matched.length >= 2) {
    const first = normalizePercentText(matched[0]);
    const second = normalizePercentText(matched[1]);
    return `${first} ~ ${second}`;
  }
  return normalizePercentText(matched[0]);
}

function normalizePercentText(value) {
  if (!value) {
    return "";
  }
  if (value.startsWith("+") || value.startsWith("-")) {
    return value;
  }
  return `+${value}`;
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

function formatScore(value) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return "-";
  }
  return num.toFixed(1);
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

function mapRiskLevel(level) {
  const source = String(level || "").toUpperCase();
  if (source === "LOW") {
    return "低风险";
  }
  if (source === "MEDIUM") {
    return "中风险";
  }
  if (source === "HIGH") {
    return "高风险";
  }
  return source || "-";
}

function buildFallbackStockExplanationMap(items) {
  const map = {};
  (items || []).forEach((item, index) => {
    if (!item?.id) {
      return;
    }
    map[item.id] = buildFallbackStockExplanation(item, index);
  });
  return map;
}

function buildFallbackStockExplanation(item, index = 0) {
  const reason = item?.reason_summary || "系统从多维信号中筛出当前候选。";
  const generatedAt = item?.valid_from || new Date().toISOString();
  return {
    seed_summary: `已处理与 ${item?.symbol || item?.name || "当前标的"} 相关的资金、技术、事件和情绪种子。`,
    seed_highlights: [reason, `${mapRiskLevel(item?.risk_level)}边界`, "资讯验证"].filter(Boolean),
    graph_summary: "系统会把标的、板块、事件和风险边界串成一条判断链。",
    consensus_summary: "趋势、事件和风控视角达成了以观察确认后执行的共识。",
    simulations: [
      {
        scenarios: [
          { scenario: "base", action: "按计划跟踪", thesis: "主逻辑延续", risk_signal: "常规波动", score_adjustment: 0 },
          { scenario: "shock", action: "缩减仓位", thesis: "若突发事件加大波动，优先守边界", risk_signal: "消息扰动", score_adjustment: -6 }
        ]
      }
    ],
    agent_opinions: [
      { agent: "趋势", stance: "支持", confidence: 0.78, summary: "量价结构仍支持继续观察。", veto: false },
      { agent: "风险", stance: "谨慎", confidence: 0.66, summary: "需严格执行止损与仓位边界。", veto: false }
    ],
    risk_flags: [`${mapRiskLevel(item?.risk_level)}波动需控仓`],
    invalidations: ["跌破预设风险线或核心催化落空时停止沿用旧结论"],
    confidence_reason: reason,
    workload_summary: {
      seed_count: 18 + index * 3,
      candidate_count: 6,
      selected_count: 1,
      agent_count: 5,
      scenario_count: 4,
      filter_steps: ["seed-ingest", "feature-factory", "scenario-engine", "agent-review"]
    },
    strategy_version: "strategy-engine-demo",
    generated_at: generatedAt
  };
}

function mapArbitrageType(type) {
  const source = String(type || "").toUpperCase();
  if (source === "CALENDAR") {
    return "跨期价差";
  }
  if (source === "CROSS") {
    return "跨品种价差";
  }
  if (source === "INTERTEMPORAL") {
    return "跨期套利";
  }
  return source || "套利";
}

function mapArbitrageStatus(status) {
  const source = String(status || "").toUpperCase();
  if (source === "WATCH") {
    return "观察中";
  }
  if (source === "ACTIVE") {
    return "可执行";
  }
  if (source === "EXPIRED") {
    return "已失效";
  }
  return source || "待确认";
}

function mapGuidanceDirection(direction) {
  const source = String(direction || "").toUpperCase();
  if (source === "LONG_SPREAD") {
    return "做多价差";
  }
  if (source === "SHORT_SPREAD") {
    return "做空价差";
  }
  return source || "-";
}

function pointWithRule(value, label) {
  const num = Number(value);
  if (!Number.isFinite(num)) {
    return `${label}区间待更新`;
  }
  return `${label} ${num.toFixed(2)} 点`;
}

function suggestPosition(percentile, riskLevel) {
  const p = Number(percentile);
  const risk = String(riskLevel || "").toUpperCase();
  if (risk === "HIGH") {
    return "总资金 15%-20%";
  }
  if (Number.isFinite(p) && p >= 0.9) {
    return "总资金 20%-25%";
  }
  if (Number.isFinite(p) && p >= 0.8) {
    return "总资金 25%-30%";
  }
  return "总资金 30%-35%";
}

function calcHitRate(rows) {
  if (!Array.isArray(rows) || rows.length === 0) {
    return "-";
  }
  const positive = rows.filter((item) => Number.isFinite(item.rawChange) && item.rawChange > 0).length;
  const rate = (positive / rows.length) * 100;
  return `${Math.round(rate)}%`;
}

function getTrendClass(value) {
  if (typeof value !== "string") {
    return "";
  }
  if (value.startsWith("+")) {
    return "up";
  }
  if (value.startsWith("-")) {
    return "down";
  }
  return "";
}

watch(
  () => isLoggedIn.value,
  () => {
    loadVIPState();
  },
  { immediate: true }
);

watch(
  () => primaryStock.value?.id,
  () => {
    refreshWatchState();
  },
  { immediate: true }
);

onMounted(() => {
  loadHomeData();
  loadHomeInsights();
  if (typeof window !== "undefined") {
    window.addEventListener(WATCHLIST_EVENT, refreshWatchState);
  }
});

onBeforeUnmount(() => {
  if (typeof window !== "undefined") {
    window.removeEventListener(WATCHLIST_EVENT, refreshWatchState);
  }
});
</script>

<style scoped>
.home-page {
  display: grid;
  gap: 14px;
}

.hero-stage {
  border-radius: 22px;
  padding: clamp(16px, 2.2vw, 24px);
  display: grid;
  gap: 12px;
  grid-template-columns: 1.05fr 0.95fr;
  background:
    radial-gradient(circle at 0% 100%, var(--color-line-gold-soft) 0%, transparent 36%),
    radial-gradient(circle at 100% 0%, rgba(64, 112, 196, 0.18) 0%, transparent 34%),
    rgba(255, 255, 255, 0.93);
}

.hero-kicker {
  margin: 0;
  font-size: 12px;
  letter-spacing: 0.3px;
  color: var(--color-pine-600);
}

.hero-copy {
  min-width: 0;
  display: grid;
  align-content: start;
}

.hero-copy h1 {
  margin: 8px 0 10px;
  font-family: var(--font-serif);
  font-size: clamp(30px, 3.5vw, 46px);
  line-height: 1.15;
}

.hero-desc {
  margin: 0;
  color: var(--color-text-sub);
  max-width: 580px;
  line-height: 1.72;
}

.hero-actions {
  margin-top: 18px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.hero-tags {
  margin-top: 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.home-hero-stats {
  margin-top: 16px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.hero-panel {
  border-radius: 18px;
  padding: 14px;
  background: linear-gradient(160deg, rgba(17, 47, 94, 0.95), rgba(30, 83, 161, 0.92));
  color: #f6fbf8;
  display: grid;
  gap: 12px;
}

.panel-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
}

.panel-head p {
  margin: 0;
  font-size: 13px;
}

.panel-head strong {
  font-size: 12px;
  color: rgba(246, 251, 248, 0.78);
}

.lead-stock-card {
  border-radius: 14px;
  border: 1px solid rgba(245, 251, 248, 0.18);
  background: rgba(255, 255, 255, 0.08);
  padding: 12px;
}

.lead-stock-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.lead-stock-symbol {
  margin: 0;
  font-size: 12px;
  letter-spacing: 0.4px;
  color: rgba(246, 251, 248, 0.7);
}

.lead-stock-top h2 {
  margin: 6px 0 0;
  font-size: 22px;
  line-height: 1.2;
}

.lead-stock-risk {
  border-radius: 999px;
  padding: 4px 10px;
  background: rgba(255, 255, 255, 0.14);
  font-size: 12px;
  color: #f6fbf8;
}

.lead-stock-expected {
  margin: 12px 0 0;
  font-size: 28px;
  font-weight: 700;
}

.lead-stock-note {
  margin: 8px 0 0;
  font-size: 12px;
  line-height: 1.6;
  color: rgba(246, 251, 248, 0.76);
}

.lead-stock-why {
  margin: 10px 0 0;
  font-size: 13px;
  line-height: 1.65;
  color: #f6fbf8;
}

.lead-stock-proof-list {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.lead-stock-meta {
  margin: 8px 0 0;
  font-size: 11px;
  color: rgba(246, 251, 248, 0.7);
}

.task-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: 8px;
}

.task-list li {
  border-radius: 10px;
  border: 1px solid rgba(245, 251, 248, 0.2);
  background: rgba(255, 255, 255, 0.05);
  padding: 8px 10px;
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 8px;
}

.task-dot {
  margin-top: 4px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.task-dot.high {
  background: var(--color-priority-high);
}

.task-dot.mid {
  background: var(--color-priority-mid);
}

.task-dot.low {
  background: var(--color-priority-low);
}

.task-title {
  margin: 0;
  font-size: 13px;
}

.task-note {
  margin: 3px 0 0;
  font-size: 12px;
  color: rgba(246, 251, 248, 0.75);
}

.decision-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: 1.1fr 0.9fr;
}

.watch-card,
.pulse-card {
  padding: 14px;
}

.decision-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.decision-head h2 {
  margin: 6px 0 0;
  font-family: var(--font-serif);
  font-size: clamp(22px, 2.3vw, 30px);
  line-height: 1.28;
}

.watch-list {
  margin-top: 12px;
  display: grid;
  gap: 10px;
}

.watch-item {
  min-width: 0;
}

.watch-item-top {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.watch-rank {
  margin: 0;
  font-size: 11px;
  letter-spacing: 0.3px;
  color: var(--color-pine-600);
}

.watch-item-top h3 {
  margin: 4px 0 0;
  font-size: 18px;
  line-height: 1.35;
}

.watch-risk {
  border-radius: 999px;
  border: 1px solid var(--color-border-soft);
  background: var(--color-surface-panel-soft-strong);
  padding: 4px 9px;
  font-size: 12px;
  color: var(--color-text-sub);
}

.watch-expected {
  margin: 10px 0 0;
  font-size: 20px;
  font-weight: 700;
}

.watch-reason,
.watch-action {
  margin: 6px 0 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--color-text-sub);
}

.watch-why {
  margin: 8px 0 0;
  font-size: 12px;
  line-height: 1.65;
  color: var(--color-text-main);
}

.watch-proof-list {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.watch-meta {
  margin: 6px 0 0;
  font-size: 11px;
  color: var(--color-text-sub);
}

.pulse-summary {
  margin: 12px 0 0;
  font-size: 14px;
  line-height: 1.72;
  color: var(--color-text-sub);
}

.research-card {
  background:
    radial-gradient(circle at 100% 0%, var(--color-focus-fill) 0%, transparent 34%),
    linear-gradient(180deg, rgba(255, 255, 255, 0.96), rgba(247, 250, 255, 0.96));
}

.research-card .pulse-summary {
  color: var(--color-text-main);
}

.pulse-signals {
  margin-top: 10px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.research-interpret-grid {
  margin-top: 12px;
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.research-interpret-item {
  min-width: 0;
}

.research-interpret-item strong {
  display: block;
  font-size: 14px;
  line-height: 1.4;
}

.research-interpret-item p {
  margin: 6px 0 0;
  font-size: 12px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.pulse-news-list {
  margin-top: 12px;
  display: grid;
  gap: 8px;
}

.pulse-news-item {
  cursor: pointer;
}

.pulse-news-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 11px;
  color: var(--color-text-sub);
}

.pulse-news-item h3 {
  margin: 6px 0 4px;
  font-size: 15px;
  line-height: 1.45;
}

.pulse-news-item p {
  margin: 0;
  font-size: 13px;
  line-height: 1.58;
  color: var(--color-text-sub);
}

.community-home-section {
  display: grid;
  gap: 12px;
  grid-template-columns: 1.08fr 0.92fr;
}

.community-home-card,
.community-home-side {
  padding: 14px;
}

.community-home-grid,
.community-home-side-list {
  margin-top: 12px;
  display: grid;
  gap: 10px;
}

.community-home-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.community-home-entry {
  min-width: 0;
}

.community-home-entry-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.community-home-entry-top p,
.community-home-entry strong,
.community-home-entry p {
  margin: 0;
}

.community-home-entry-top p {
  font-size: 12px;
  color: var(--color-pine-600);
}

.community-home-entry strong {
  display: block;
  margin-top: 6px;
  font-size: 16px;
  line-height: 1.5;
  color: var(--color-text-main);
}

.community-home-entry p:last-child {
  margin-top: 6px;
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.community-home-entry-actions {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.community-home-entry-actions button {
  width: auto;
}

.home-search-banner {
  border-radius: 20px;
  padding: 14px 16px;
  border: 1px solid var(--color-border-soft);
  background:
    radial-gradient(circle at 0% 0%, rgba(36, 87, 167, 0.2) 0%, transparent 36%),
    radial-gradient(circle at 100% 100%, rgba(184, 130, 48, 0.34) 0%, transparent 38%),
    var(--color-surface-card-elevated);
}

.search-banner-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.search-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-pine-600);
}

.search-banner-head h2 {
  margin: 6px 0 0;
  font-family: var(--font-serif);
  font-size: clamp(21px, 2.2vw, 30px);
  line-height: 1.25;
}

.search-jump-btn {
  width: auto;
}

.home-search-banner .insight-search {
  margin-top: 12px;
}

.home-search-banner .insight-search input {
  font-size: 14px;
  padding: 11px 12px;
}

.search-tip {
  margin: 8px 0 0;
}

.action-pill-list {
  margin-top: 12px;
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.action-pill {
  width: auto;
}

.conversion-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: 1.05fr 0.95fr;
}

.conversion-card,
.cadence-card {
  padding: 14px;
}

.conversion-lanes {
  margin-top: 12px;
  display: grid;
  gap: 10px;
}

.conversion-lane {
  border-radius: 14px;
  border: 1px solid var(--color-border-soft);
  background: var(--color-surface-card-soft);
  padding: 12px;
  display: grid;
  gap: 6px;
  opacity: 0.8;
}

.conversion-lane.unlocked {
  opacity: 1;
}

.conversion-lane.active {
  border-color: var(--color-border-focus);
  background:
    linear-gradient(135deg, rgba(236, 243, 255, 0.96), rgba(248, 251, 255, 0.96));
  box-shadow: var(--shadow-card-active);
}

.conversion-lane-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.conversion-lane-head p,
.conversion-lane strong,
.conversion-lane p:last-child {
  margin: 0;
}

.conversion-lane-head p {
  font-size: 12px;
  font-weight: 700;
  color: var(--color-pine-700);
}

.conversion-lane strong {
  font-size: 16px;
  line-height: 1.45;
  color: var(--color-text-main);
}

.conversion-lane p:last-child {
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.cadence-list {
  margin-top: 12px;
  display: grid;
  gap: 10px;
}

.cadence-item {
  min-width: 0;
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr) auto;
  gap: 10px;
  align-items: start;
}

.cadence-time {
  margin: 0;
  font-size: 18px;
  font-weight: 700;
  color: var(--color-pine-700);
}

.cadence-item strong,
.cadence-item p {
  margin: 0;
}

.cadence-item div {
  display: grid;
  gap: 4px;
}

.cadence-item div p {
  font-size: 13px;
  line-height: 1.65;
  color: var(--color-text-sub);
}

.cadence-access {
  align-self: start;
}

.mobile-quick {
  display: none;
  padding: 10px;
  border-radius: 16px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 8px;
}

.mobile-quick button {
  border: 1px solid var(--color-border-soft);
  background: var(--color-surface-card-soft);
  border-radius: 12px;
  padding: 9px;
  text-align: left;
  cursor: pointer;
}

.mobile-quick p {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.mobile-quick span {
  margin-top: 3px;
  display: block;
  font-size: 11px;
  color: var(--color-text-sub);
}

.mobile-quick-hint {
  grid-column: 1 / -1;
  margin: 2px 2px 0;
}

.insight-section {
  display: grid;
  gap: 12px;
  grid-template-columns: 1.15fr 0.85fr;
}

.insight-section.single-column {
  grid-template-columns: 1fr;
}

.insight-card,
.vip-promo-card {
  padding: 14px;
}

.insight-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.insight-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.insight-actions button {
  width: auto;
}

.insight-actions button.ghost {
  color: var(--color-pine-700);
}

.insight-actions button:disabled {
  opacity: 0.72;
}

.insight-search {
  margin-top: 10px;
  display: grid;
  grid-template-columns: 1fr auto auto;
  gap: 8px;
}

.insight-search input {
  border: 1px solid var(--color-border-soft-heavy);
  border-radius: 10px;
  padding: 9px 11px;
  background: var(--color-surface-card-elevated);
  color: var(--color-text-main);
}

.insight-search input:focus {
  outline: none;
  border-color: var(--color-pine-500);
  box-shadow: var(--shadow-focus);
}

.insight-search button {
  width: auto;
}

.insight-search button.ghost {
  color: var(--color-pine-700);
}

.insight-search button:disabled {
  opacity: 0.7;
}

.insight-status {
  margin: 8px 0 0;
}

.insight-tags {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.insight-list {
  margin-top: 10px;
  display: grid;
  gap: 8px;
}

.insight-item {
  border: 1px solid var(--color-border-soft);
  border-radius: 12px;
  background: var(--color-surface-card-soft);
  padding: 10px;
  cursor: pointer;
}

.insight-item-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.insight-item .category {
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
  color: var(--color-pine-700);
  background: var(--color-surface-accent-strong);
}

.insight-item .time {
  font-size: 11px;
  color: var(--color-text-sub);
}

.insight-item h3 {
  margin: 7px 0 5px;
  font-size: 16px;
  line-height: 1.45;
}

.insight-item p {
  margin: 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.58;
}

.insight-meta {
  margin-top: 8px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  font-size: 11px;
  color: var(--color-text-sub);
}

.vip-mark {
  background: var(--color-surface-accent-strong);
  color: var(--color-warning);
}

.vip-mark.active {
  background: rgba(192, 139, 60, 0.2);
  color: var(--color-accent);
}

.vip-promo-card {
  border: 1px solid rgba(214, 222, 234, 0.86);
  background:
    radial-gradient(circle at 100% 0%, rgba(192, 139, 60, 0.26) 0%, transparent 36%),
    radial-gradient(circle at 0% 100%, var(--color-focus-glow) 0%, transparent 38%),
    var(--color-surface-card-elevated);
}

.promo-kicker {
  margin: 0;
  font-size: 12px;
  color: var(--color-accent);
}

.vip-promo-card h2 {
  margin: 8px 0 0;
  font-family: var(--font-serif);
  font-size: 25px;
  line-height: 1.3;
}

.promo-price {
  margin: 10px 0 0;
  font-size: 28px;
  color: var(--color-pine-700);
  font-weight: 700;
}

.promo-note {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
}

.promo-benefits {
  margin: 12px 0 0;
  padding-left: 18px;
  display: grid;
  gap: 7px;
  color: var(--color-text-sub);
  font-size: 13px;
}

.promo-btn {
  margin-top: 14px;
  width: 100%;
  background: linear-gradient(145deg, #c08b3c, #a8742e);
}

.stock-section {
  display: grid;
  gap: 12px;
  grid-template-columns: 1.3fr 0.7fr;
}

.history-card,
.today-card,
.arbitrage-card,
.method-card {
  padding: 14px;
}

.history-card header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.section-ghost-btn {
  flex-shrink: 0;
}

.stock-table {
  min-width: 680px;
}

.mobile-history-list {
  display: none;
}

.mobile-history-item .top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.mobile-history-item .top p {
  margin: 0;
  font-weight: 700;
  font-size: 13px;
}

.mobile-history-item .top span {
  font-size: 12px;
  color: var(--color-text-sub);
}

.mobile-history-item .line {
  margin-top: 5px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-size: 12px;
}

.up {
  color: var(--color-rise);
  font-weight: 600;
}

.down {
  color: var(--color-fall);
  font-weight: 600;
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

.today-list {
  margin-top: 10px;
  display: grid;
  gap: 8px;
}

.today-main {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.code {
  margin: 0;
  font-weight: 700;
}

.name {
  margin: 0;
  font-size: 13px;
  color: var(--color-text-sub);
}

.expect {
  margin: 6px 0 0;
  font-size: 18px;
  font-weight: 700;
}

.expect.up {
  color: var(--color-rise);
}

.expect.flat {
  color: var(--color-warning);
}

.expect.down {
  color: var(--color-fall);
}

.condition {
  margin: 4px 0 0;
  font-size: 12px;
  color: var(--color-text-sub);
  line-height: 1.5;
}

.futures-section {
  display: grid;
  gap: 12px;
  grid-template-columns: 1.1fr 0.9fr;
}

.arbitrage-list {
  margin-top: 10px;
  display: grid;
  gap: 10px;
}

.plan-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.plan-title p {
  margin: 0;
  font-weight: 600;
}

.plan-title span {
  font-size: 12px;
  color: var(--color-warning);
  background: rgba(227, 207, 173, 0.55);
  border-radius: 999px;
  padding: 2px 8px;
}

.plan-grid {
  margin-top: 9px;
  display: grid;
  gap: 8px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.plan-grid div {
  border-radius: 10px;
  border: 1px solid rgba(214, 222, 234, 0.85);
  background: rgba(246, 249, 255, 0.84);
  padding: 8px;
}

.plan-grid p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.plan-grid strong {
  display: block;
  margin-top: 4px;
  font-size: 16px;
  color: var(--color-pine-700);
}

.method-list {
  margin: 10px 0 0;
  padding-left: 18px;
  display: grid;
  gap: 9px;
}

.step-title {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
}

.step-desc {
  margin: 3px 0 0;
  font-size: 13px;
  color: var(--color-text-sub);
  line-height: 1.56;
}

.risk-tip {
  margin: 12px 0 0;
}

@media (max-width: 1080px) {
  .hero-stage,
  .decision-grid,
  .community-home-section,
  .conversion-grid,
  .insight-section,
  .stock-section,
  .futures-section {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 720px) {
  .hero-copy h1 {
    font-size: 32px;
  }

  .search-banner-head {
    flex-direction: column;
  }

  .search-jump-btn {
    width: 100%;
  }

  .decision-head {
    flex-direction: column;
  }

  .cadence-item {
    grid-template-columns: 1fr;
  }

  .mobile-quick {
    display: grid;
  }

  .insight-head {
    flex-direction: column;
  }

  .research-interpret-grid {
    grid-template-columns: 1fr;
  }

  .community-home-grid {
    grid-template-columns: 1fr;
  }

  .insight-actions {
    width: 100%;
  }

  .insight-actions button {
    flex: 1;
  }

  .insight-search {
    grid-template-columns: 1fr 1fr;
  }

  .insight-search input {
    grid-column: 1 / -1;
  }

  .vip-promo-card h2 {
    font-size: 22px;
  }

  .history-card header {
    flex-direction: column;
  }

  .action-pill-list {
    flex-direction: column;
  }

  .action-pill {
    width: 100%;
  }

  .community-home-entry-actions {
    display: grid;
    grid-template-columns: 1fr;
  }

  .community-home-entry-actions button {
    width: 100%;
  }

  .table-wrap {
    display: none;
  }

  .mobile-history-list {
    margin-top: 10px;
    display: grid;
    gap: 8px;
  }

  .plan-grid {
    grid-template-columns: 1fr;
  }
}
</style>