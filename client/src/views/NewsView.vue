<template>
  <section class="news-page fade-up">
    <header class="news-hero card">
      <div class="news-hero-copy finance-copy-stack">
        <div class="finance-pill-row">
          <span class="finance-pill finance-pill-compact finance-pill-neutral">资讯页</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">资讯域内搜索</span>
          <span class="finance-pill finance-pill-compact finance-pill-info">导读进入深读</span>
        </div>
        <div>
          <p class="hero-kicker">资讯中心</p>
          <h1 class="section-title">新闻、研报与期刊</h1>
          <p class="section-subtitle">按栏目查看文章详情和附件。</p>
        </div>
      </div>
      <div class="hero-status finance-summary-pill">
        <p>更新状态</p>
        <strong>{{ loading ? "同步中" : "已同步" }}</strong>
      </div>
      <div class="news-hero-stats finance-hero-stat-grid">
        <article class="finance-hero-stat-card">
          <span>焦点研报</span>
          <strong>{{ featuredArticle?.title || "待同步焦点内容" }}</strong>
          <p>优先承接首页和策略页导入的深读动作。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>当前栏目</span>
          <strong>{{ currentCategory.label }}</strong>
          <p>研报、新闻、期刊分层阅读，不做内容瀑布流。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>阅读位置</span>
          <strong>{{ canReadActiveArticle ? "摘要 -> 正文 -> 附件" : "摘要 -> 登录 / VIP -> 继续阅读" }}</strong>
          <p>先看导读，再继续正文和附件，不打断阅读链。</p>
        </article>
        <article class="finance-hero-stat-card">
          <span>权限承接</span>
          <strong>{{ featuredArticle?.visibilityLabel || "游客 / 登录 / VIP" }}</strong>
          <p>游客先看摘要，登录后保留位置，VIP 解锁全文与附件。</p>
        </article>
      </div>
    </header>

    <StatePanel
      :tone="newsAccessState.tone"
      :eyebrow="newsAccessState.label"
      :title="newsAccessState.title"
      :description="newsAccessState.desc"
    >
      <template #actions>
        <button type="button" class="finance-primary-btn" @click="handleNewsPrimaryAction('state_panel_primary')">{{ newsPrimaryActionText }}</button>
        <button type="button" class="ghost finance-ghost-btn" @click="handleNewsSecondaryAction">{{ newsSecondaryActionText }}</button>
      </template>
    </StatePanel>

    <section class="news-focus-layout finance-dual-rail">
      <article class="card news-focus-card finance-section-card">
        <header class="news-focus-head finance-section-head-grid">
          <div>
            <p class="hero-kicker">今日焦点</p>
            <h2 class="section-title">查看今日重点文章，并继续阅读正文与附件。</h2>
            <p class="section-subtitle">
              左侧切换栏目，右侧查看详情。
            </p>
          </div>
          <div class="news-focus-actions finance-action-row">
            <button type="button" class="finance-primary-btn" :disabled="loading" @click="loadNewsData">
              {{ loading ? "同步中..." : "刷新资讯" }}
            </button>
            <button type="button" class="ghost finance-ghost-btn" @click="handleNewsPrimaryAction('focus_primary')">
              {{ newsPrimaryActionText }}
            </button>
          </div>
        </header>

        <div v-if="featuredArticle" class="news-featured-panel finance-card-pale">
          <div class="news-featured-head">
            <div>
              <p class="decision-tag">今日焦点{{ featuredArticleCategoryLabel }}</p>
              <h3>{{ featuredArticle.title }}</h3>
            </div>
            <div class="news-featured-badges">
              <span class="finance-pill finance-pill-roomy finance-pill-info">{{ featuredArticleCategoryLabel }}</span>
              <span class="finance-pill finance-pill-roomy finance-pill-info">{{ featuredArticle.time }}</span>
              <span class="finance-pill finance-pill-roomy finance-pill-info">{{ featuredArticle.visibilityLabel }}</span>
            </div>
          </div>
          <p class="news-featured-summary">{{ featuredArticle.desc }}</p>
          <div v-if="featuredArticle.tags.length" class="tags">
            <span
              v-for="tag in featuredArticle.tags"
              :key="`featured-${tag}`"
              class="finance-pill finance-pill-compact finance-pill-accent"
            >
              {{ tag }}
            </span>
          </div>
          <div class="news-feature-grid">
            <article class="finance-list-card finance-list-card-panel">
              <p>摘要</p>
              <strong>快速把握重点</strong>
              <span>先看摘要，再决定是否继续阅读全文。</span>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>正文</p>
              <strong>{{ canReadActiveArticle ? "可阅读全文" : "登录或升级后可阅读全文" }}</strong>
              <span>{{ activeArticleLockDesc }}</span>
            </article>
            <article class="finance-list-card finance-list-card-panel">
              <p>附件</p>
              <strong>{{ activeAttachments.length }} 个</strong>
              <span>附件保留在详情区查看。</span>
            </article>
          </div>
        </div>
        <StatePanel
          v-else
          tone="info"
          eyebrow="今日焦点"
          title="当前还没有可查看的资讯样本"
          description="刷新后将优先展示今日重点文章。"
          compact
        />
      </article>

      <aside class="news-focus-side finance-stack-tight finance-sticky-side">
        <article class="card news-side-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">栏目概览</h2>
              <p class="section-subtitle">新闻、研报、期刊分开展示。</p>
            </div>
          </header>
          <div class="news-role-list">
            <article v-for="item in contentRoleRows" :key="item.title" class="finance-list-card finance-list-card-panel">
              <div class="top-line">
                <p>{{ item.title }}</p>
                <span>{{ item.count }}</span>
              </div>
              <strong>{{ item.summary }}</strong>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </article>

        <article class="card news-side-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">权限说明</h2>
              <p class="section-subtitle">游客可看摘要，登录后保留阅读位置，VIP 解锁全文与附件。</p>
            </div>
          </header>
          <div class="news-role-list">
            <article v-for="item in newsReadingGuideRows" :key="item.title" class="finance-list-card finance-list-card-panel">
              <p>{{ item.title }}</p>
              <strong>{{ item.summary }}</strong>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </article>
      </aside>
    </section>

    <nav class="sub-nav card">
      <button
        v-for="item in categoryTabs"
        :key="item.key"
        type="button"
        class="sub-item"
        :class="{ active: activeTab === item.key }"
        @click="activeTab = item.key"
      >
        <span class="label">{{ item.label }}</span>
        <span class="count">{{ feedMap[item.key]?.length || 0 }}</span>
      </button>
    </nav>

    <div class="api-tip">
      <p v-if="loading">正在拉取资讯内容...</p>
      <p v-else-if="errorMessage">{{ errorMessage }}</p>
      <p v-else>
        数据更新时间：{{ lastUpdatedAt || "-" }}
        <template v-if="searchKeyword"> · 当前关键词：{{ searchKeyword }}</template>
      </p>
    </div>

    <div class="news-grid">
      <article class="card feed-card">
        <header class="finance-copy-stack">
          <h2 class="section-title">{{ currentCategory.label }}</h2>
          <p class="section-subtitle">{{ currentCategory.subtitle }}</p>
        </header>

        <div class="feed-list" v-if="currentFeed.length">
          <article
            v-for="item in currentFeed"
            :key="item.id"
            class="feed-item finance-list-card finance-list-card-interactive"
            :class="{ active: activeArticleID === item.id }"
            @click="openArticle(item.id)"
          >
            <div class="meta">
              <span class="time">{{ item.time }}</span>
              <span class="level finance-pill finance-pill-compact" :class="item.levelClass">{{ item.level }}</span>
            </div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.desc }}</p>
            <div class="tags">
              <span v-for="tag in item.tags" :key="tag" class="finance-pill finance-pill-compact finance-pill-accent">{{ tag }}</span>
            </div>
          </article>
        </div>
        <div v-else class="empty-box finance-empty-box">当前栏目暂无内容</div>
        <div v-if="currentFeed.length" class="feed-footer">
          <p class="feed-progress">{{ feedProgressText }}</p>
          <button
            type="button"
            class="finance-mini-btn finance-mini-btn-primary"
            :disabled="!canLoadMore || currentPagination.loading"
            @click="loadMoreForActiveCategory"
          >
            {{
              currentPagination.loading
                ? "加载中..."
                : canLoadMore
                  ? "加载更多"
                  : "已显示全部"
            }}
          </button>
        </div>
      </article>

      <aside class="side">
        <article class="card reading-guide-card finance-section-card">
          <header class="section-head compact">
            <div>
              <h2 class="section-title">阅读说明</h2>
              <p class="section-subtitle">左侧选文章，右侧查看摘要、正文和附件。</p>
            </div>
          </header>
          <div class="news-role-list">
            <article v-for="item in detailGuideRows" :key="item.title" class="finance-list-card finance-list-card-panel">
              <p>{{ item.title }}</p>
              <strong>{{ item.summary }}</strong>
              <span>{{ item.desc }}</span>
            </article>
          </div>
        </article>

        <article class="card detail-card">
          <header class="detail-head">
            <div>
              <p class="decision-tag">{{ currentCategory.label }}详情</p>
              <h2>文章详情</h2>
            </div>
            <button
              type="button"
              class="finance-mini-btn finance-mini-btn-accent"
              :disabled="detailLoading"
              @click="refreshActiveArticleDetail"
            >
              {{ detailLoading ? "同步中..." : "刷新详情" }}
            </button>
          </header>

          <div v-if="activeArticleDetail" class="detail-body">
            <p class="detail-title">{{ activeArticleDetail.title }}</p>
            <p class="detail-meta">
              {{ activeArticleDetail.publishedAt }} · {{ activeArticleDetail.visibilityLabel }} · 作者
              {{ activeArticleDetail.authorID || "-" }}
            </p>
            <p class="detail-summary">{{ activeArticleDetail.summary }}</p>
            <template v-if="canReadActiveArticle">
              <div class="detail-content" v-html="activeArticleDetail.contentHTML"></div>

              <div class="attachment-box">
                <p class="attachment-title">附件</p>
                <div v-if="activeAttachments.length" class="attachment-list">
                  <button
                    v-for="item in activeAttachments"
                    :key="item.id || item.file_url || item.file_name"
                    type="button"
                    class="attachment-item"
                    :disabled="downloadingAttachmentID === (item.id || item.file_url || '')"
                    @click="handleDownloadAttachment(item)"
                  >
                    <span>{{ item.file_name || "未命名附件" }}</span>
                    <small>{{ formatAttachmentSize(item.file_size) }}</small>
                  </button>
                </div>
                <p v-else class="empty-inline finance-empty-inline">暂无附件</p>
              </div>
            </template>
            <div v-else class="detail-lock-card">
              <p class="detail-lock-kicker">{{ activeArticleDetail.visibilityLabel }} 内容</p>
              <strong>{{ activeArticleLockTitle }}</strong>
              <span>{{ activeArticleLockDesc }}</span>
              <div class="detail-lock-actions">
                <button type="button" class="finance-primary-btn" @click="handleNewsPrimaryAction('detail_lock_primary')">{{ newsPrimaryActionText }}</button>
                <button type="button" class="ghost-btn finance-ghost-btn" @click="handleNewsSecondaryAction">
                  {{ newsSecondaryActionText }}
                </button>
              </div>
            </div>
          </div>
          <div v-else class="empty-box finance-empty-box">暂无可查看文章</div>

          <p v-if="detailErrorMessage" class="detail-error">{{ detailErrorMessage }}</p>
        </article>
      </aside>
    </div>

    <section v-if="isMobileView && isDetailModalOpen" class="mobile-detail-panel">
      <header class="mobile-detail-head">
        <button type="button" class="mobile-nav-btn finance-mini-btn finance-mini-btn-card" @click="closeDetailModal">返回</button>
        <div>
          <p>{{ currentCategory.label }}详情</p>
          <h2>{{ selectedArticleTitle }}</h2>
        </div>
        <button
          type="button"
          class="mobile-nav-btn finance-mini-btn finance-mini-btn-accent"
          :disabled="detailLoading"
          @click="refreshActiveArticleDetail"
        >
          {{ detailLoading ? "同步中" : "刷新" }}
        </button>
      </header>

      <div class="mobile-detail-body">
        <article v-if="activeArticle" class="mobile-detail-card">
          <div class="mobile-info-block">
            <p class="mobile-info-row">
              <span>标题</span>
              <strong>{{ selectedArticleTitle }}</strong>
            </p>
            <p class="mobile-info-row">
              <span>日期</span>
              <strong>{{ selectedPublishedAt }}</strong>
            </p>
            <p class="mobile-info-row">
              <span>发布人</span>
              <strong>{{ selectedPublisher }}</strong>
            </p>
          </div>

          <template v-if="activeArticleDetail">
            <section class="mobile-content-box">
              <p class="mobile-block-title">内容</p>
              <p class="mobile-detail-summary">{{ activeArticleDetail.summary }}</p>
              <div v-if="canReadActiveArticle" class="mobile-detail-content" v-html="activeArticleDetail.contentHTML"></div>
              <div v-else class="mobile-lock-card">
                <strong>{{ activeArticleLockTitle }}</strong>
                <span>{{ activeArticleLockDesc }}</span>
                <div class="mobile-lock-actions">
                  <button type="button" class="finance-primary-btn" @click="handleNewsPrimaryAction('mobile_lock_primary')">{{ newsPrimaryActionText }}</button>
                  <button type="button" class="ghost-btn finance-ghost-btn" @click="handleNewsSecondaryAction">
                    {{ newsSecondaryActionText }}
                  </button>
                </div>
              </div>
            </section>

            <div class="mobile-attachment-box">
              <p class="mobile-block-title">附件</p>
              <div v-if="canReadActiveArticle && activeAttachments.length" class="mobile-attachment-list">
                <button
                  v-for="item in activeAttachments"
                  :key="item.id || item.file_url || item.file_name"
                  type="button"
                  class="mobile-attachment-item"
                  :disabled="downloadingAttachmentID === (item.id || item.file_url || '')"
                  @click="handleDownloadAttachment(item)"
                >
                  <span>{{ item.file_name || "未命名附件" }}</span>
                  <small>{{ formatAttachmentSize(item.file_size) }}</small>
                </button>
              </div>
              <p v-else class="mobile-empty-inline finance-empty-inline">
                {{ canReadActiveArticle ? "暂无附件" : "开通对应权限后可查看附件" }}
              </p>
            </div>
          </template>
          <template v-else>
            <section class="mobile-content-box">
              <p class="mobile-block-title">内容</p>
              <p class="mobile-loading-tip">详情加载中...</p>
            </section>
            <div class="mobile-attachment-box">
              <p class="mobile-block-title">附件</p>
              <p class="mobile-empty-inline finance-empty-inline">详情加载后可查看附件</p>
            </div>
          </template>
        </article>
        <div v-else class="mobile-empty-box finance-empty-box">暂无可查看文章</div>
        <p v-if="detailErrorMessage" class="mobile-detail-error">{{ detailErrorMessage }}</p>
      </div>
    </section>
  </section>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import StatePanel from "../components/StatePanel.vue";
import { getMembershipQuota } from "../api/membership";
import {
  getAttachmentSignedURL,
  getNewsArticleDetail,
  listNewsArticles,
  listNewsAttachments,
  listNewsCategories
} from "../api/news";
import { useClientAuth } from "../lib/client-auth";
import {
  promotePendingExperimentJourneySources,
  rememberExperimentAttributionSource,
  rememberPendingExperimentJourneySource
} from "../lib/growth-analytics";
import { getExperimentVariant } from "../lib/growth-experiments";
import { getAccessToken } from "../lib/session";

const categoryTemplates = [
  {
    key: "news",
    label: "新闻",
    subtitle: "查看市场动态与政策消息。"
  },
  {
    key: "report",
    label: "研报",
    subtitle: "查看机构深度分析与重点观点。"
  },
  {
    key: "journal",
    label: "期刊",
    subtitle: "查看长期观点与方法总结。"
  }
];

const categoryTemplateMap = categoryTemplates.reduce((result, item) => {
  result[item.key] = item;
  return result;
}, {});
const defaultCategoryTemplate = categoryTemplates[0];
const DEFAULT_PAGE_SIZE = 20;

const route = useRoute();
const router = useRouter();
const { isLoggedIn } = useClientAuth();
const newsMembershipExperimentVariant = getExperimentVariant("news_membership_entry", ["default"]);

const activeTab = ref("");
const loading = ref(false);
const errorMessage = ref("");
const lastUpdatedAt = ref("");
const searchKeyword = ref(normalizeKeyword(route.query.keyword || ""));
const categoryTabs = ref([]);
const feedMap = ref({});
const paginationMap = ref({});
const detailMap = ref({});
const attachmentMap = ref({});
const activeArticleID = ref("");
const detailLoading = ref(false);
const detailErrorMessage = ref("");
const downloadingAttachmentID = ref("");
const isDetailModalOpen = ref(false);
const isMobileView = ref(false);
const memberStageLoading = ref(false);
const isVIPUser = ref(false);

const currentCategory = computed(
  () =>
    categoryTabs.value.find((item) => item.key === activeTab.value) ||
    categoryTabs.value[0] ||
    defaultCategoryTemplate
);
const currentFeed = computed(() => feedMap.value[activeTab.value] || []);
const currentPagination = computed(() => {
  const fallback = { page: 1, pageSize: DEFAULT_PAGE_SIZE, total: 0, loading: false };
  return paginationMap.value[activeTab.value] || fallback;
});
const canLoadMore = computed(() => {
  if (currentPagination.value.loading) {
    return false;
  }
  const loaded = currentFeed.value.length;
  const total = Number(currentPagination.value.total || 0);
  if (total > 0) {
    return loaded < total;
  }
  const page = Number(currentPagination.value.page || 1);
  const pageSize = Number(currentPagination.value.pageSize || DEFAULT_PAGE_SIZE);
  return loaded >= page * pageSize;
});
const feedProgressText = computed(() => {
  const loaded = currentFeed.value.length;
  const total = Number(currentPagination.value.total || 0);
  if (total > 0) {
    return `已展示 ${loaded} / ${total} 条`;
  }
  return `已展示 ${loaded} 条`;
});
const allFeedRows = computed(() => Object.values(feedMap.value || {}).flat());
const activeArticle = computed(
  () => currentFeed.value.find((item) => item.id === activeArticleID.value) || currentFeed.value[0] || null
);
const activeArticleDetail = computed(() => {
  const item = activeArticle.value;
  if (!item) {
    return null;
  }
  const detail = detailMap.value[item.id] || {};
  const visibilityRaw = String(detail.visibility || item.visibility || "").toUpperCase();
  const rawContent = detail.content || detail.summary || item.desc || "-";
  return {
    id: item.id,
    title: detail.title || item.title || "未命名资讯",
    summary: detail.summary || item.summary || item.desc || "-",
    content: rawContent,
    contentHTML: renderArticleHTML(rawContent),
    publishedAt: formatDateTime(detail.published_at || item.publishedAtRaw || item.time),
    visibilityLabel: visibilityRaw === "VIP" ? "VIP" : "公开",
    authorID: detail.author_id || item.authorID || "-"
  };
});
const activeAttachments = computed(() => {
  const id = activeArticle.value?.id;
  if (!id) {
    return [];
  }
  return attachmentMap.value[id] || [];
});
const selectedArticleTitle = computed(() => {
  const rawTitle = activeArticleDetail.value?.title || activeArticle.value?.title || "";
  return toPlainText(rawTitle) || "未命名资讯";
});
const selectedPublishedAt = computed(() => {
  if (activeArticleDetail.value?.publishedAt) {
    return activeArticleDetail.value.publishedAt;
  }
  if (!activeArticle.value) {
    return "-";
  }
  return formatDateTime(activeArticle.value.publishedAtRaw || activeArticle.value.time);
});
const selectedPublisher = computed(() => {
  if (activeArticleDetail.value?.authorID) {
    return activeArticleDetail.value.authorID;
  }
  return activeArticle.value?.authorID || "-";
});
const currentArticleNeedsVIP = computed(
  () => String(activeArticleDetail.value?.visibilityLabel || "").toUpperCase() === "VIP"
);
const canReadActiveArticle = computed(() => !currentArticleNeedsVIP.value || isVIPUser.value);
const featuredArticle = computed(() => {
  const reportItem = allFeedRows.value.find((item) =>
    (item.tags || []).some((tag) => /(研报|研究|深度|forecast|report|insight)/i.test(String(tag || "")))
  );
  return reportItem || activeArticle.value || allFeedRows.value[0] || null;
});
const featuredArticleCategoryLabel = computed(() => {
  const firstTag = featuredArticle.value?.tags?.[0];
  return firstTag || currentCategory.value?.label || "资讯";
});
const newsAccessStage = computed(() => {
  if (isVIPUser.value) {
    return "VIP";
  }
  if (isLoggedIn.value) {
    return "REGISTERED";
  }
  return "VISITOR";
});
const newsAccessState = computed(() => {
  if (memberStageLoading.value) {
    return {
      tone: "info",
      label: "识别中",
      title: "正在确认你的资讯访问阶段",
      desc: "确认完成后，会把资讯详情、附件和 CTA 调整到对应阶段。"
    };
  }
  if (newsAccessStage.value === "VIP") {
    return {
      tone: "success",
      label: "会员阶段",
      title: "你已解锁深度资讯，可直接阅读正文和附件。",
      desc: "可结合策略页和关注页继续查看。"
    };
  }
  if (newsAccessStage.value === "REGISTERED") {
    return {
      tone: "warning",
      label: "注册阶段",
      title: "可先查看公开资讯和摘要，再判断是否升级。",
      desc: "登录后会保留当前阅读位置。"
    };
  }
  return {
    tone: "info",
    label: "游客阶段",
    title: "可先查看公开资讯与摘要。",
    desc: "登录后可继续当前阅读，并保存关注方向。"
  };
});
const newsPrimaryActionText = computed(() => {
  if (newsAccessStage.value === "VIP") {
    return "去策略页查看";
  }
  if (newsAccessStage.value === "REGISTERED") {
    return "升级会员看深度内容";
  }
  return "先登录保存关注";
});
const newsSecondaryActionText = computed(() => {
  if (newsAccessStage.value === "VIP") {
    return "去我的关注";
  }
  if (newsAccessStage.value === "REGISTERED") {
    return "去策略页";
  }
  return "看历史档案";
});
const activeArticleLockTitle = computed(() =>
  newsAccessStage.value === "REGISTERED" ? "这篇深度内容需要 VIP 才能查看全文和附件。" : "登录后先进入注册路径，再决定是否升级 VIP。"
);
const activeArticleLockDesc = computed(() => {
  if (newsAccessStage.value === "VIP") {
    return "当前内容已解锁。";
  }
  if (newsAccessStage.value === "REGISTERED") {
    return "当前可查看摘要和公开资讯，升级后可解锁正文与附件。";
  }
  return "当前先展示摘要，登录后可继续当前阅读。";
});
const contentRoleRows = computed(() => [
  {
    title: "新闻",
    count: `${countFeedItemsByRole("news")} 条`,
    summary: "关注盘中动态和高影响事件",
    desc: "适合快速确认市场变化。"
  },
  {
    title: "研报",
    count: `${countFeedItemsByRole("report")} 条`,
    summary: "查看深度判断和参数分析",
    desc: "适合深入阅读重点观点。"
  },
  {
    title: "期刊",
    count: `${countFeedItemsByRole("journal")} 条`,
    summary: "查看长期观点与方法内容",
    desc: "适合做长期阅读和复盘。"
  }
]);
const newsReadingGuideRows = computed(() => [
  {
    title: "当前阶段",
    summary: newsAccessState.value.label,
    desc: newsAccessState.value.desc
  },
  {
    title: "正文阅读",
    summary: canReadActiveArticle.value ? "当前正文已可读" : "当前可先查看摘要",
    desc: "需要更多内容时，可继续登录或升级。"
  },
  {
    title: "继续查看",
    summary: "资讯页负责阅读正文",
    desc: "可继续前往策略页或关注页查看相关内容。"
  }
]);
const detailGuideRows = computed(() => [
  {
    title: "先看标题和摘要",
    summary: "先快速判断是否值得继续阅读",
    desc: "右侧详情区会先展示标题、时间和摘要。"
  },
  {
    title: "再看正文和附件",
    summary: canReadActiveArticle.value ? "当前可以继续读完整内容" : "登录或升级后可看完整内容",
    desc: "正文和附件都保留在详情区查看。"
  },
  {
    title: "继续查看相关内容",
    summary: "需要时可前往其他页面",
    desc: "看完可回策略页确认动作，或回关注页查看跟踪变化。"
  }
]);

watch(
  currentFeed,
  (items) => {
    if (!items.length) {
      activeArticleID.value = "";
      return;
    }
    if (!items.some((item) => item.id === activeArticleID.value)) {
      activeArticleID.value = items[0].id;
    }
  },
  { immediate: true }
);

watch(
  () => activeArticleID.value,
  () => {
    loadActiveArticleDetail({ silent: true });
  },
  { immediate: true }
);

watch(
  () => isMobileView.value,
  (mobile) => {
    if (!mobile) {
      isDetailModalOpen.value = false;
    }
  }
);

watch(
  () => [isDetailModalOpen.value, isMobileView.value],
  ([opened, mobile]) => {
    if (typeof document === "undefined") {
      return;
    }
    document.body.style.overflow = opened && mobile ? "hidden" : "";
  },
  { immediate: true }
);

watch(
  () => route.query.keyword,
  (value) => {
    const normalized = normalizeKeyword(value);
    if (normalized === searchKeyword.value) {
      return;
    }
    searchKeyword.value = normalized;
    loadNewsData();
  }
);

watch(
  () => isLoggedIn.value,
  () => {
    loadMembershipStage();
  }
);

async function loadNewsData() {
  loading.value = true;
  errorMessage.value = "";
  const errors = [];
  let resolvedCategories = [];
  try {
    const categoryData = await listNewsCategories();
    const backendCategories = normalizeBackendCategories(categoryData?.items || []);
    if (backendCategories.length > 0) {
      resolvedCategories = backendCategories;
    } else {
      errors.push("当前暂无可用资讯分类");
    }
  } catch (error) {
    errors.push(`资讯分类加载失败：${error?.message || "unknown error"}`);
  }

  categoryTabs.value = resolvedCategories;
  if (categoryTabs.value.length === 0) {
    activeTab.value = "";
    activeArticleID.value = "";
    feedMap.value = {};
    paginationMap.value = {};
    detailMap.value = {};
    attachmentMap.value = {};
    errorMessage.value = errors.join("；");
    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
    loading.value = false;
    return;
  }

  if (!categoryTabs.value.some((item) => item.key === activeTab.value)) {
    activeTab.value = categoryTabs.value[0]?.key || defaultCategoryTemplate.key;
  }

  const nextFeedMap = {};
  const nextPaginationMap = {};
  categoryTabs.value.forEach((category) => {
    nextFeedMap[category.key] = [];
    nextPaginationMap[category.key] = {
      page: 1,
      pageSize: DEFAULT_PAGE_SIZE,
      total: 0,
      loading: false
    };
  });
  feedMap.value = nextFeedMap;
  paginationMap.value = nextPaginationMap;

  await Promise.all(
    categoryTabs.value.map(async (category) => {
      try {
        await fetchCategoryArticles(category, { page: 1, reset: true });
      } catch (error) {
        errors.push(`${category.label}加载失败：${error?.message || "unknown error"}`);
      }
    })
  );

  if (errors.length > 0) {
    errorMessage.value = errors.join("；");
  }
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
  }
  promoteNewsPostAuthAttribution();
}

async function fetchCategoryArticles(category, options = {}) {
  const { page = 1, reset = false } = options;
  const categoryKey = category?.key;
  if (!categoryKey) {
    return;
  }

  setCategoryPagination(categoryKey, { loading: true });
  try {
    const params = buildArticleQuery(category, page);
    const data = await listNewsArticles(params);
    const items = Array.isArray(data?.items) ? data.items : [];
    const mappedRows = items.map((item) => mapArticleItem(item, category));

    const nextFeedMap = { ...feedMap.value };
    if (reset) {
      nextFeedMap[categoryKey] = mappedRows;
    } else {
      nextFeedMap[categoryKey] = mergeFeedRows(nextFeedMap[categoryKey] || [], mappedRows);
    }
    feedMap.value = nextFeedMap;

    const nextDetailMap = { ...detailMap.value };
    items.forEach((item) => {
      if (!item?.id) {
        return;
      }
      nextDetailMap[item.id] = {
        ...nextDetailMap[item.id],
        id: item.id,
        title: item.title || nextDetailMap[item.id]?.title,
        summary: item.summary || nextDetailMap[item.id]?.summary,
        visibility: item.visibility || nextDetailMap[item.id]?.visibility,
        status: item.status || nextDetailMap[item.id]?.status,
        published_at: item.published_at || item.created_at || item.updated_at,
        author_id: item.author_id || nextDetailMap[item.id]?.author_id
      };
    });
    detailMap.value = nextDetailMap;

    const total = Number(data?.total ?? data?.count ?? 0);
    const currentPage = Number(data?.page ?? page);
    const currentPageSize = Number(data?.page_size ?? data?.pageSize ?? params.page_size ?? DEFAULT_PAGE_SIZE);
    setCategoryPagination(categoryKey, {
      page: Number.isFinite(currentPage) && currentPage > 0 ? currentPage : page,
      pageSize: Number.isFinite(currentPageSize) && currentPageSize > 0 ? currentPageSize : DEFAULT_PAGE_SIZE,
      total: Number.isFinite(total) && total >= 0 ? total : 0,
      loading: false
    });
  } catch (error) {
    setCategoryPagination(categoryKey, { loading: false });
    throw error;
  }
}

function rememberNewsMembershipEntry(targetKey, metadata = {}) {
  rememberExperimentAttributionSource({
    experimentKey: "news_membership_entry",
    variantKey: newsMembershipExperimentVariant,
    pageKey: "news",
    targetKey,
    userStage: newsAccessStage.value,
    metadata: {
      active_category: currentCategory.value?.key || "",
      article_id: activeArticle.value?.id || "",
      article_visibility: currentArticleNeedsVIP.value ? "VIP" : "PUBLIC",
      ...metadata
    }
  });
}

function rememberNewsPendingMembershipEntry(targetKey, metadata = {}) {
  rememberPendingExperimentJourneySource({
    experimentKey: "news_membership_entry",
    variantKey: newsMembershipExperimentVariant,
    pageKey: "news",
    targetKey,
    userStage: "VISITOR",
    redirectPath: "/news",
    metadata: {
      active_category: currentCategory.value?.key || "",
      article_id: activeArticle.value?.id || "",
      article_visibility: currentArticleNeedsVIP.value ? "VIP" : "PUBLIC",
      ...metadata
    }
  });
}

function promoteNewsPostAuthAttribution() {
  if (!isLoggedIn.value) {
    return;
  }
  promotePendingExperimentJourneySources({
    experimentKey: "news_membership_entry",
    pageKey: "news",
    userStage: newsAccessStage.value,
    metadata: {
      active_category: currentCategory.value?.key || "",
      article_id: activeArticle.value?.id || "",
      article_visibility: currentArticleNeedsVIP.value ? "VIP" : "PUBLIC"
    }
  });
}

function handleNewsPrimaryAction(targetKey = "primary_action") {
  if (newsAccessStage.value === "VIP") {
    router.push("/strategies");
    return;
  }
  if (newsAccessStage.value === "REGISTERED") {
    rememberNewsMembershipEntry(targetKey, {
      destination: "/membership"
    });
    router.push("/membership");
    return;
  }
  rememberNewsPendingMembershipEntry(targetKey, {
    destination_after_auth: "/news"
  });
  router.push({ path: "/auth", query: { redirect: "/news" } });
}

function handleNewsSecondaryAction() {
  if (newsAccessStage.value === "VIP") {
    router.push("/watchlist");
    return;
  }
  if (newsAccessStage.value === "REGISTERED") {
    router.push("/strategies");
    return;
  }
  router.push("/archive");
}

function countFeedItemsByRole(role) {
  return allFeedRows.value.filter((item) => itemMatchesRole(item, role)).length;
}

async function loadMoreForActiveCategory() {
  const category = currentCategory.value;
  if (!category?.key || !canLoadMore.value || currentPagination.value.loading) {
    return;
  }
  try {
    await fetchCategoryArticles(category, {
      page: Number(currentPagination.value.page || 1) + 1,
      reset: false
    });
    lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  } catch (error) {
    errorMessage.value = error?.message || "加载更多失败";
  }
}

function setCategoryPagination(categoryKey, patch = {}) {
  const previous = paginationMap.value[categoryKey] || {
    page: 1,
    pageSize: DEFAULT_PAGE_SIZE,
    total: 0,
    loading: false
  };
  paginationMap.value = {
    ...paginationMap.value,
    [categoryKey]: {
      ...previous,
      ...patch
    }
  };
}

function mergeFeedRows(currentRows, nextRows) {
  const seen = new Set();
  const merged = [];
  [...currentRows, ...nextRows].forEach((item) => {
    const id = item?.id;
    if (!id || seen.has(id)) {
      return;
    }
    seen.add(id);
    merged.push(item);
  });
  return merged;
}

function openArticle(id) {
  activeArticleID.value = id;
  if (isMobileView.value) {
    isDetailModalOpen.value = true;
  }
}

function closeDetailModal() {
  isDetailModalOpen.value = false;
}

function syncViewport() {
  if (typeof window === "undefined") {
    return;
  }
  isMobileView.value = window.innerWidth <= 980;
}

async function loadActiveArticleDetail(options = {}) {
  const { silent = false } = options;
  detailErrorMessage.value = "";
  const article = activeArticle.value;
  if (!article?.id) {
    return;
  }
  if (article.id.includes("_local_")) {
    return;
  }
  if (!silent) {
    detailLoading.value = true;
  }
  try {
    const [detailRes, attachmentRes] = await Promise.allSettled([
      getNewsArticleDetail(article.id),
      listNewsAttachments(article.id)
    ]);

    if (detailRes.status === "fulfilled" && detailRes.value?.id) {
      detailMap.value = {
        ...detailMap.value,
        [article.id]: detailRes.value
      };
    } else if (detailRes.status === "rejected") {
      detailErrorMessage.value = resolveDetailErrorMessage(detailRes.reason, article.visibility);
    }

    if (attachmentRes.status === "fulfilled") {
      attachmentMap.value = {
        ...attachmentMap.value,
        [article.id]: normalizeAttachments(attachmentRes.value?.items || [])
      };
    } else if (attachmentRes.status === "rejected") {
      detailErrorMessage.value = resolveDetailErrorMessage(attachmentRes.reason, article.visibility);
    }
  } finally {
    if (!silent) {
      detailLoading.value = false;
    }
  }
}

function refreshActiveArticleDetail() {
  loadActiveArticleDetail();
}

async function handleDownloadAttachment(item) {
  if (!item) {
    return;
  }
  const id = item.id || item.file_url || "";
  downloadingAttachmentID.value = id;
  detailErrorMessage.value = "";
  try {
    let url = item.file_url || "";
    if (item.id && getAccessToken()) {
      try {
        const data = await getAttachmentSignedURL(item.id);
        url = data?.signed_url || url;
      } catch (error) {
        if (!url) {
          throw error;
        }
      }
    }
    if (!url) {
      throw new Error("附件链接为空");
    }
    if (typeof window !== "undefined") {
      window.open(url, "_blank", "noopener,noreferrer");
    }
  } catch (error) {
    detailErrorMessage.value = error?.message || "附件下载失败";
  } finally {
    downloadingAttachmentID.value = "";
  }
}

function normalizeAttachments(items) {
  return (items || []).map((item) => ({
    id: item.id,
    file_name: item.file_name,
    file_url: item.file_url,
    file_size: Number(item.file_size || 0),
    mime_type: item.mime_type,
    created_at: item.created_at
  }));
}

function normalizeBackendCategories(items) {
  return [...(items || [])]
    .filter((item) => Boolean(item?.id))
    .sort((a, b) => Number(a?.sort || 0) - Number(b?.sort || 0))
    .map((item, index) => toCategoryTab(item, index))
    .filter((item) => Boolean(item?.key));
}

function toCategoryTab(category, index) {
  if (!category?.id) {
    return null;
  }
  const template = resolveCategoryTemplate(category.name, category.slug);
  return {
    key: category.id,
    backendID: category.id,
    slug: category.slug || "",
    visibility: category.visibility || "",
    label: category.name || `分类${index + 1}`,
    subtitle: template.subtitle
  };
}

function resolveCategoryTemplate(name, slug) {
  const text = `${name || ""} ${slug || ""}`;
  if (/(研报|研究|深度|盈利预测|report|insight|forecast)/i.test(text)) {
    return categoryTemplateMap.report || defaultCategoryTemplate;
  }
  if (/(期刊|周刊|月刊|journal|weekly)/i.test(text)) {
    return categoryTemplateMap.journal || defaultCategoryTemplate;
  }
  return categoryTemplateMap.news || defaultCategoryTemplate;
}

function normalizeKeyword(value) {
  return String(value || "")
    .replace(/\s+/g, " ")
    .trim();
}

function buildArticleQuery(category, page = 1) {
  const keyword = normalizeKeyword(searchKeyword.value);
  const base = {
    page,
    page_size: DEFAULT_PAGE_SIZE
  };
  if (keyword) {
    base.keyword = keyword;
  }
  if (category?.backendID) {
    return {
      ...base,
      category_id: category.backendID
    };
  }
  if (category?.key === "report") {
    return { ...base, keyword: keyword || "研报" };
  }
  if (category?.key === "journal") {
    return { ...base, keyword: keyword || "期刊" };
  }
  return base;
}

function mapArticleItem(item, category) {
  const time = formatDateTime(item.published_at || item.created_at || item.updated_at);
  const visibility = String(item.visibility || "").toUpperCase();
  const levelClass = visibility === "VIP" ? "high" : "mid";
  const level = visibility === "VIP" ? "高" : "中";
  return {
    id: item.id,
    time,
    level,
    levelClass,
    title: item.title || "未命名资讯",
    desc: item.summary || item.content || "-",
    summary: item.summary || item.content || "-",
    tags: buildTags(item, category),
    visibility: item.visibility,
    status: item.status,
    publishedAtRaw: item.published_at || item.created_at || item.updated_at,
    authorID: item.author_id || ""
  };
}

function itemMatchesRole(item, role) {
  const text = `${item?.title || ""} ${(item?.tags || []).join(" ")}`;
  if (role === "report") {
    return /(研报|研究|深度|forecast|report|insight)/i.test(text);
  }
  if (role === "journal") {
    return /(期刊|周刊|月刊|journal|weekly)/i.test(text);
  }
  return !itemMatchesRole(item, "report") && !itemMatchesRole(item, "journal");
}

function buildTags(item, category) {
  const tags = [];
  if (category?.label) {
    tags.push(category.label);
  } else if (category?.key === "news") {
    tags.push("新闻");
  } else if (category?.key === "report") {
    tags.push("研报");
  } else if (category?.key === "journal") {
    tags.push("期刊");
  }
  if (item.visibility) tags.push(String(item.visibility).toUpperCase());
  if (item.status) tags.push(String(item.status).toUpperCase());
  return tags.slice(0, 3);
}

function resolveDetailErrorMessage(error, visibility) {
  const level = String(visibility || "").toUpperCase();
  const code = Number(error?.code || 0);
  if (code === 40101 || code === 40103 || code === 40105 || code === 401) {
    return "请先登录账号，登录后可查看详情和附件";
  }
  if (level === "VIP" && (code === 40401 || code === 404 || code === 40301 || code === 403)) {
    return "该内容为VIP专享，请登录并开通会员后查看完整内容";
  }
  return error?.message || "详情加载失败";
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

function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) return value || "-";
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

function formatAttachmentSize(value) {
  const size = Number(value || 0);
  if (!Number.isFinite(size) || size <= 0) {
    return "-";
  }
  if (size >= 1024 * 1024) {
    return `${(size / (1024 * 1024)).toFixed(1)} MB`;
  }
  if (size >= 1024) {
    return `${(size / 1024).toFixed(1)} KB`;
  }
  return `${size} B`;
}

function renderArticleHTML(content) {
  const text = String(content || "").trim();
  if (!text) {
    return "<p>-</p>";
  }
  if (containsHTMLTag(text)) {
    return sanitizeArticleHTML(text);
  }
  return `<p>${escapeHTML(text).replace(/\r?\n/g, "<br/>")}</p>`;
}

function containsHTMLTag(value) {
  return /<\/?[a-z][\s\S]*>/i.test(String(value || ""));
}

function escapeHTML(value) {
  return String(value || "")
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

function toPlainText(value) {
  return String(value || "")
    .replace(/<[^>]*>/g, " ")
    .replace(/\s+/g, " ")
    .trim();
}

function sanitizeArticleHTML(rawHTML) {
  if (typeof window === "undefined" || typeof DOMParser === "undefined") {
    return String(rawHTML || "")
      .replace(/<script[\s\S]*?>[\s\S]*?<\/script>/gi, "")
      .replace(/<style[\s\S]*?>[\s\S]*?<\/style>/gi, "");
  }
  const parser = new DOMParser();
  const doc = parser.parseFromString(String(rawHTML || ""), "text/html");
  const disallowedSelector =
    "script,style,iframe,object,embed,link,meta,form,input,button,textarea,select";
  doc.querySelectorAll(disallowedSelector).forEach((node) => node.remove());
  doc.querySelectorAll("*").forEach((element) => {
    Array.from(element.attributes).forEach((attr) => {
      const attrName = String(attr.name || "").toLowerCase();
      const attrValue = String(attr.value || "").trim();
      if (attrName.startsWith("on")) {
        element.removeAttribute(attr.name);
        return;
      }
      if (["href", "src", "xlink:href"].includes(attrName)) {
        const loweredValue = attrValue.toLowerCase();
        const safeValue =
          loweredValue.startsWith("http://") ||
          loweredValue.startsWith("https://") ||
          loweredValue.startsWith("/") ||
          loweredValue.startsWith("#") ||
          loweredValue.startsWith("mailto:") ||
          loweredValue.startsWith("tel:") ||
          loweredValue.startsWith("data:image/");
        if (!safeValue) {
          element.removeAttribute(attr.name);
        }
      }
    });
  });
  return doc.body.innerHTML || "<p>-</p>";
}

onMounted(() => {
  syncViewport();
  if (typeof window !== "undefined") {
    window.addEventListener("resize", syncViewport);
  }
  loadMembershipStage();
  loadNewsData();
});

onBeforeUnmount(() => {
  if (typeof window !== "undefined") {
    window.removeEventListener("resize", syncViewport);
  }
  if (typeof document !== "undefined") {
    document.body.style.overflow = "";
  }
});
</script>

<style scoped>
.news-page {
  display: grid;
  gap: 12px;
}

.news-focus-layout {
  --finance-main-column: minmax(0, 1.22fr);
  --finance-side-column: 340px;
}

.news-focus-actions button {
  width: auto;
}

.news-featured-panel {
  margin-top: 12px;
  padding: 14px;
  display: grid;
  gap: 12px;
}

.news-featured-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.news-featured-head h3 {
  margin: 4px 0 0;
  font-size: 22px;
  line-height: 1.35;
  color: var(--color-text-main);
}

.news-featured-badges {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.news-featured-badges span {
  box-shadow: inset 0 0 0 1px var(--color-focus-fill);
}

.news-featured-summary {
  margin: 0;
  font-size: 13px;
  line-height: 1.72;
  color: var(--color-text-sub);
}

.news-feature-grid,
.news-role-list {
  display: grid;
  gap: 8px;
}

.news-feature-grid {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.news-feature-grid article,
.news-role-list article {
  display: grid;
  gap: 4px;
}

.news-feature-grid p,
.news-role-list p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.news-feature-grid strong,
.news-role-list strong {
  font-size: 14px;
  color: var(--color-text-main);
}

.news-feature-grid span,
.news-role-list span {
  font-size: 12px;
  line-height: 1.6;
  color: var(--color-text-sub);
}

.news-focus-side {
  align-content: start;
}

.news-hero {
  border-radius: 20px;
  padding: 16px;
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: end;
  gap: 10px;
}

.news-hero-copy {
  min-width: 0;
}

.news-hero-stats {
  grid-column: 1 / -1;
}

.hero-kicker {
  margin: 0;
  color: var(--color-pine-600);
  font-size: 12px;
}

.hero-status {
  padding: 8px 10px;
}

.hero-status p {
  font-size: 11px;
}

.hero-status strong {
  margin-top: 2px;
  font-size: 16px;
}

.sub-nav {
  border-radius: 16px;
  padding: 8px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: 8px;
}

.sub-item {
  border: 1px solid var(--color-border-soft);
  background: var(--color-surface-panel-strong);
  border-radius: 11px;
  padding: 10px 12px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  transition: all 0.2s ease;
}

.sub-item .label {
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text-main);
}

.sub-item .count {
  font-size: 12px;
  color: var(--color-text-sub);
}

.sub-item.active {
  background: var(--gradient-primary);
  border-color: transparent;
}

.sub-item.active .label,
.sub-item.active .count {
  color: #fff;
}

.api-tip {
  border-radius: 10px;
  border: 1px dashed var(--color-border-soft-heavy);
  background: var(--color-surface-panel-soft-muted);
  padding: 8px 10px;
}

.api-tip p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.news-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: 1fr 0.88fr;
}

.feed-card,
.detail-card {
  padding: 14px;
}

.feed-list {
  margin-top: 10px;
  display: grid;
  gap: 8px;
}

.feed-footer {
  margin-top: 10px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  flex-wrap: wrap;
}

.feed-progress {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-sub);
}

.feed-footer button {
  width: auto;
}

.feed-footer button:disabled {
  opacity: 0.72;
}

.meta {
  display: flex;
  gap: 8px;
  align-items: center;
}

.time {
  color: var(--color-pine-600);
  font-size: 12px;
}

.level {
  font-weight: 700;
}

.level.high {
  color: var(--color-accent);
  background: var(--color-surface-gold-soft);
}

.level.mid {
  color: var(--color-pine-700);
  background: var(--color-surface-accent-glow);
}

.level.low {
  color: var(--color-text-sub);
  background: var(--color-surface-panel-soft-subtle);
}

h3 {
  margin: 6px 0 5px;
  line-height: 1.4;
  font-size: 18px;
}

.feed-item p {
  margin: 0;
  color: var(--color-text-sub);
  line-height: 1.62;
}

.tags {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.tags span {
  letter-spacing: 0.01em;
}

.side {
  display: grid;
  gap: 12px;
  align-content: start;
}

.reading-guide-card {
  background: var(--color-surface-card-strong);
}

.detail-card {
  background: linear-gradient(160deg, rgba(19, 54, 103, 0.96), rgba(30, 83, 161, 0.96));
  color: #f5faf7;
}

.detail-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 10px;
}

.decision-tag {
  margin: 0;
  font-size: 12px;
  color: rgba(248, 251, 255, 0.76);
}

.detail-head h2 {
  margin: 7px 0 0;
  font-family: var(--font-serif);
  font-size: 22px;
}

.detail-head button {
  width: auto;
}

.detail-head button:disabled {
  opacity: 0.7;
}

.detail-body {
  margin-top: 10px;
}

.detail-title {
  margin: 0;
  font-size: 18px;
  line-height: 1.45;
}

.detail-meta {
  margin: 6px 0 0;
  font-size: 12px;
  color: rgba(248, 251, 255, 0.74);
}

.detail-summary {
  margin: 10px 0 0;
  font-size: 13px;
  line-height: 1.62;
  color: rgba(248, 251, 255, 0.84);
}

.detail-lock-card {
  margin-top: 10px;
  border-radius: 14px;
  border: 1px solid var(--color-line-gold);
  background: var(--color-surface-gold-overlay);
  padding: 12px;
  display: grid;
  gap: 6px;
}

.detail-lock-kicker {
  margin: 0;
  font-size: 11px;
  letter-spacing: 0.08em;
  color: rgba(248, 251, 255, 0.72);
}

.detail-lock-card strong,
.detail-lock-card span {
  display: block;
}

.detail-lock-card strong {
  font-size: 16px;
  line-height: 1.5;
  color: #f8fbff;
}

.detail-lock-card span {
  font-size: 13px;
  line-height: 1.65;
  color: rgba(248, 251, 255, 0.82);
}

.detail-lock-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 2px;
}

.detail-lock-actions .ghost-btn {
  background: var(--color-surface-card-elevated);
}

.detail-content {
  margin: 8px 0 0;
  font-size: 13px;
  line-height: 1.62;
  color: rgba(248, 251, 255, 0.9);
}

.detail-content :deep(p) {
  margin: 0 0 10px;
}

.detail-content :deep(ul),
.detail-content :deep(ol) {
  margin: 6px 0 10px 18px;
  padding: 0;
}

.detail-content :deep(li) {
  margin-bottom: 4px;
}

.detail-content :deep(a) {
  color: var(--color-text-gold-soft);
  text-decoration: underline;
}

.detail-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
}

.attachment-box {
  margin-top: 10px;
  border-top: 1px solid var(--color-line-gold);
  padding-top: 10px;
}

.attachment-title {
  margin: 0;
  font-size: 12px;
  color: rgba(248, 251, 255, 0.76);
}

.attachment-list {
  margin-top: 8px;
  display: grid;
  gap: 6px;
}

.attachment-item {
  border: 1px solid var(--color-line-gold);
  background: var(--color-surface-gold-overlay);
  color: #f8fbff;
  border-radius: 9px;
  padding: 7px 9px;
  text-align: left;
  cursor: pointer;
  display: flex;
  justify-content: space-between;
  gap: 8px;
  align-items: center;
}

.attachment-item:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.attachment-item small {
  font-size: 11px;
  color: rgba(248, 251, 255, 0.72);
}

.empty-inline {
  margin: 8px 0 0;
  font-size: 12px;
  color: rgba(248, 251, 255, 0.7);
}

.detail-error {
  margin: 10px 0 0;
  color: var(--color-text-gold-soft);
  font-size: 12px;
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

@media (max-width: 980px) {
  .news-focus-layout,
  .news-hero,
  .news-grid {
    grid-template-columns: 1fr;
  }

  .news-focus-head,
  .news-featured-head,
  .news-hero {
    display: grid;
  }

  .news-feature-grid {
    grid-template-columns: 1fr;
  }

  .news-focus-actions {
    justify-content: flex-start;
  }

  .side {
    display: grid;
    gap: 12px;
  }

  .detail-card {
    display: none;
  }

  .mobile-detail-panel {
    position: fixed;
    inset: 0;
    z-index: 66;
    background: linear-gradient(180deg, var(--color-surface-soft) 0%, var(--color-champagne-100) 100%);
    display: grid;
    grid-template-rows: auto 1fr;
  }

  .mobile-detail-head {
    position: sticky;
    top: 0;
    z-index: 1;
    padding: 10px;
    border-bottom: 1px solid var(--color-border-soft);
    background: var(--color-surface-card-top);
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 8px;
  }

  .mobile-detail-head p {
    margin: 0;
    font-size: 11px;
    color: var(--color-text-sub);
  }

  .mobile-detail-head h2 {
    margin: 2px 0 0;
    font-size: 15px;
    font-family: var(--font-serif);
    color: var(--color-text-main);
    line-height: 1.4;
    overflow-wrap: anywhere;
  }

  .mobile-nav-btn {
    width: auto;
  }

  .mobile-nav-btn:disabled {
    opacity: 0.72;
  }

  .mobile-detail-body {
    overflow-y: auto;
    padding: 10px;
  }

  .mobile-detail-card {
    border: 1px solid var(--color-border-soft-heavy);
    border-radius: 14px;
    padding: 12px;
    background: var(--color-surface-card-top);
  }

  .mobile-info-block {
    border: 1px solid var(--color-border-soft-strong);
    border-radius: 11px;
    background: var(--color-surface-panel-soft);
    padding: 9px;
    display: grid;
    gap: 8px;
  }

  .mobile-info-row {
    margin: 0;
    display: grid;
    gap: 3px;
  }

  .mobile-info-row span {
    font-size: 11px;
    color: var(--color-text-sub);
  }

  .mobile-info-row strong {
    font-size: 14px;
    line-height: 1.45;
    color: var(--color-text-main);
    overflow-wrap: anywhere;
  }

  .mobile-content-box {
    margin-top: 10px;
  }

  .mobile-block-title {
    margin: 0;
    font-size: 12px;
    color: var(--color-text-sub);
  }

  .mobile-detail-summary {
    margin: 7px 0 0;
    font-size: 13px;
    line-height: 1.65;
    color: var(--color-text-main);
  }

  .mobile-detail-content {
    margin: 8px 0 0;
    font-size: 14px;
    line-height: 1.72;
    color: var(--color-text-main);
  }

  .mobile-lock-card {
    margin-top: 8px;
    border-radius: 12px;
    border: 1px solid var(--color-line-gold);
    background: var(--color-surface-gold-soft);
    padding: 10px;
    display: grid;
    gap: 6px;
  }

  .mobile-lock-card strong,
  .mobile-lock-card span {
    display: block;
  }

  .mobile-lock-card strong {
    font-size: 14px;
    line-height: 1.5;
    color: var(--color-text-main);
  }

  .mobile-lock-card span {
    font-size: 12px;
    line-height: 1.6;
    color: var(--color-text-sub);
  }

  .mobile-lock-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .mobile-lock-actions .ghost-btn {
    background: var(--color-surface-card-elevated);
  }

  .mobile-detail-content :deep(p) {
    margin: 0 0 10px;
  }

  .mobile-detail-content :deep(ul),
  .mobile-detail-content :deep(ol) {
    margin: 6px 0 10px 18px;
    padding: 0;
  }

  .mobile-detail-content :deep(li) {
    margin-bottom: 4px;
  }

  .mobile-detail-content :deep(a) {
    color: var(--color-pine-700);
    text-decoration: underline;
  }

  .mobile-detail-content :deep(img) {
    max-width: 100%;
    height: auto;
    border-radius: 8px;
  }

  .mobile-attachment-box {
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid var(--color-border-soft);
  }

  .mobile-attachment-list {
    margin-top: 8px;
    display: grid;
    gap: 6px;
  }

  .mobile-attachment-item {
    border: 1px solid var(--color-border-soft);
    border-radius: 9px;
    padding: 8px 9px;
    background: var(--color-surface-panel-soft);
    color: var(--color-text-main);
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 8px;
    cursor: pointer;
    text-align: left;
  }

  .mobile-attachment-item:disabled {
    opacity: 0.72;
    cursor: not-allowed;
  }

  .mobile-attachment-item small {
    font-size: 11px;
    color: var(--color-text-sub);
  }

  .mobile-empty-inline,
  .mobile-empty-box {
    margin: 8px 0 0;
    font-size: 12px;
    color: var(--color-text-sub);
  }

  .mobile-loading-tip {
    margin: 10px 0 0;
    font-size: 12px;
    color: var(--color-text-sub);
  }

  .mobile-detail-error {
    margin: 10px 2px 0;
    font-size: 12px;
    color: var(--color-fall);
  }
}

@media (max-width: 640px) {
  .sub-nav {
    display: flex;
    gap: 8px;
    overflow-x: auto;
    padding: 8px;
    scrollbar-width: none;
  }

  .sub-nav::-webkit-scrollbar {
    display: none;
  }

  .sub-item {
    flex: 0 0 auto;
    min-width: 112px;
  }

  .feed-footer {
    flex-direction: column;
    align-items: stretch;
  }

  .feed-footer button {
    width: 100%;
  }

  h3 {
    font-size: 16px;
  }
}
</style>
