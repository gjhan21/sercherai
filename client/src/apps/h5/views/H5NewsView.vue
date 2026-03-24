<template>
  <div class="h5-page fade-up news-page">
    <div class="h5-page-topline">
      <span class="h5-page-tagline">资讯流</span>
      <span>{{ activeArticleID ? "正文阅读" : activeTabLabel }}</span>
    </div>

    <template v-if="activeArticleID">
      <section class="news-detail-head">
        <button type="button" class="news-detail-back" @click="closeArticle">返回栏目</button>
        <div class="news-detail-meta">
          <span>{{ detailCategoryLabel }}</span>
          <span>{{ detailMetaLine }}</span>
        </div>
        <h1>{{ detailTitle }}</h1>
        <p>{{ activeArticle?.summary || "当前文章摘要待补充。" }}</p>
        <div class="news-detail-tags">
          <span v-for="tag in detailTags" :key="tag" class="news-detail-chip">{{ tag }}</span>
        </div>
      </section>

      <section class="news-reading-card">
        <div class="news-reading-topline">
          <strong>正文</strong>
          <span>{{ articleAccessState.locked ? "会员专享" : "移动阅读" }}</span>
        </div>

        <div v-if="articleAccessState.locked" class="news-lock-panel">
          <div class="news-lock-copy">
            <strong>该内容需会员权限</strong>
            <p>{{ articleAccessState.message }}</p>
          </div>
          <div class="news-lock-actions">
            <button type="button" class="h5-btn" @click="goPrimaryAction">{{ primaryActionLabel }}</button>
            <button type="button" class="h5-btn-secondary" @click="goSecondaryAction">{{ secondaryActionLabel }}</button>
          </div>
        </div>

        <div v-else-if="activeArticleDetail" class="news-body-prose" v-html="articleHtml" />

        <H5EmptyState
          v-else
          title="正文待同步"
          :description="detailLoading ? '正在同步详情...' : detailErrorMessage || '当前文章还没有更多正文内容。'"
        />
      </section>

      <section class="news-support-card">
        <div class="news-support-head">
          <div>
            <strong>附件与延伸</strong>
            <p>附件、关联动作收在正文之后，不打断阅读顺序。</p>
          </div>
        </div>

        <div v-if="articleAccessState.locked" class="news-inline-tip">
          <strong>开通会员后可查看附件</strong>
          <p>{{ articleAccessState.message }}</p>
        </div>

        <div v-else-if="activeAttachments.length" class="news-attachment-list">
          <article v-for="item in activeAttachments" :key="item.id || item.file_name" class="news-attachment-item">
            <div class="news-attachment-copy">
              <strong>{{ item.file_name || "未命名附件" }}</strong>
              <p>{{ formatAttachmentSize(item.file_size) }} · {{ item.mime_type || "文件" }}</p>
            </div>
            <button
              type="button"
              class="h5-btn-secondary"
              :disabled="downloadingAttachmentID === (item.id || item.file_name)"
              @click="handleDownloadAttachment(item)"
            >
              {{ downloadingAttachmentID === (item.id || item.file_name) ? "准备中..." : "下载" }}
            </button>
          </article>
        </div>

        <H5EmptyState
          v-else
          title="当前没有附件"
          description="如该文章有附件权限，下载入口会显示在这里。"
        />

        <div class="news-next-actions">
          <button type="button" class="h5-btn block" @click="goPrimaryAction">{{ primaryActionLabel }}</button>
          <button type="button" class="h5-btn-secondary block" @click="goSecondaryAction">{{ secondaryActionLabel }}</button>
          <button type="button" class="h5-btn-ghost block" @click="closeArticle">返回资讯流</button>
        </div>
      </section>
    </template>

    <template v-else>
      <H5HeroCard eyebrow="市场资讯" :title="activeTabLabel" :description="leadSummary" :meta="newsHeroMeta" tone="accent">
        <template #actions>
          <H5ActionBar>
            <button type="button" class="h5-btn-ghost news-refresh-btn" :disabled="loading" @click="refreshPage">
              {{ loading ? "同步中..." : "刷新" }}
            </button>
          </H5ActionBar>
        </template>

        <div class="news-hero-metrics h5-grid-3">
          <H5SummaryCard
            v-for="(item, index) in newsHeroStats"
            :key="item.label"
            :label="item.label"
            :value="item.value"
            :note="item.note"
            :tone="resolveHighlightTone(index)"
          />
        </div>
      </H5HeroCard>

      <H5SectionBlock eyebrow="栏目筛选" title="搜索、切栏目，然后继续顺着内容流往下刷" tone="soft">
        <form class="news-search-bar" @submit.prevent="searchArticles">
          <input
            v-model.trim="keyword"
            type="text"
            maxlength="40"
            enterkeyhint="search"
            placeholder="搜索资讯、研报、公告"
          />
          <button type="submit" class="h5-btn">搜索</button>
        </form>

        <div class="news-category-strip" aria-label="资讯栏目">
          <button
            v-for="item in categoryTabs"
            :key="item.key"
            type="button"
            class="news-category-pill"
            :class="{ active: activeTab === item.key }"
            @click="selectTab(item.key)"
          >
            {{ item.label }}
          </button>
        </div>
      </H5SectionBlock>

      <H5SectionBlock v-if="tickerItems.length" eyebrow="热读" title="当前栏目最值得先扫一眼的三条" tone="accent">
        <div class="news-ticker-list">
          <span v-for="item in tickerItems" :key="item" class="news-ticker-item">{{ item }}</span>
        </div>
      </H5SectionBlock>

      <button v-if="leadArticle" type="button" class="news-lead-card" @click="openArticle(leadArticle.id)">
        <div class="news-lead-meta">
          <span>{{ leadArticle.tags?.[0] || activeTabLabel }}</span>
          <span>{{ leadArticle.meta }}</span>
        </div>
        <h3>{{ leadArticle.title }}</h3>
        <p>{{ leadArticle.summary }}</p>
        <div class="news-lead-footer">
          <div class="news-detail-tags">
            <span v-for="tag in leadArticle.tags" :key="tag" class="news-detail-chip">{{ tag }}</span>
          </div>
          <span class="news-lead-cta">{{ leadArticle.visibility === "VIP" ? "会员阅读" : "立即阅读" }}</span>
        </div>
      </button>

      <H5SectionBlock eyebrow="最新内容" title="像国内资讯 App 一样，按时间顺序继续往下刷">
        <div v-if="feedItems.length" class="news-feed-list">
          <button v-for="item in feedItems" :key="item.id" type="button" class="news-feed-item" @click="openArticle(item.id)">
            <div class="news-feed-item-head">
              <span class="news-feed-section">{{ item.primaryTag }}</span>
              <span class="news-feed-time">{{ item.meta }}</span>
            </div>
            <strong>{{ item.title }}</strong>
            <p>{{ item.summary }}</p>
            <div class="news-feed-item-foot">
              <span class="news-feed-badge" :class="item.tone">{{ item.accessLabel }}</span>
            </div>
          </button>
        </div>

        <H5EmptyState v-else title="当前栏目暂无内容" description="尝试切换栏目或重新搜索关键词。" />

        <div class="news-feed-actions">
          <button type="button" class="h5-btn-secondary" :disabled="currentPagination.loading || !canLoadMore" @click="loadMore">
            {{ currentPagination.loading ? "加载中..." : canLoadMore ? "继续加载" : "已显示全部" }}
          </button>
          <button type="button" class="h5-btn" @click="goArchive">查看历史档案</button>
          <button type="button" class="h5-btn-ghost" @click="router.push('/strategies')">去看策略</button>
        </div>
      </H5SectionBlock>
    </template>

    <H5StickyCta
      :title="stickyTitle"
      :description="stickyDescription"
      :primary-label="primaryActionLabel"
      :secondary-label="activeArticleID ? '返回资讯' : '去策略页'"
      @primary="goPrimaryAction"
      @secondary="activeArticleID ? closeArticle() : router.push('/strategies')"
    />
  </div>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import H5ActionBar from "../components/H5ActionBar.vue";
import H5EmptyState from "../components/H5EmptyState.vue";
import H5HeroCard from "../components/H5HeroCard.vue";
import H5SectionBlock from "../components/H5SectionBlock.vue";
import H5StickyCta from "../components/H5StickyCta.vue";
import H5SummaryCard from "../components/H5SummaryCard.vue";
import { getMembershipQuota } from "../../../api/membership";
import { getAttachmentSignedURL, getNewsArticleDetail, listNewsArticles, listNewsAttachments, listNewsCategories } from "../../../api/news";
import { getAccessToken } from "../../../shared/auth/session";
import { useClientAuth } from "../../../shared/auth/client-auth";
import { shouldUseDemoFallback } from "../../../lib/fallback-policy";
import { shapeNewsDisplayTitle } from "../lib/display-copy.js";
import { formatAttachmentSize, formatDateTime, renderArticleHTML, resolveVipStage, toArray, toPlainText, truncateText } from "../lib/formatters";
import { buildNewsFeedRows, sortNewsCategories } from "../lib/news-feed.js";
import { resolveNewsAccessState } from "../lib/page-state";
import { fallbackNewsArticles, fallbackNewsAttachments, fallbackNewsCategories } from "../lib/mock-data";
import { resolveHighlightTone } from "../lib/surface-tone.js";

const DEFAULT_PAGE_SIZE = 8;
const router = useRouter();
const route = useRoute();
const { isLoggedIn } = useClientAuth();
const useDemoFallback = shouldUseDemoFallback();

const loading = ref(false);
const detailLoading = ref(false);
const keyword = ref(String(route.query.keyword || ""));
const activeTab = ref(String(route.query.tab || "news"));
const rawCategories = ref(useDemoFallback ? [...fallbackNewsCategories] : []);
const feedMap = ref({});
const paginationMap = ref({});
const detailMap = ref({});
const attachmentMap = ref(useDemoFallback ? { ...fallbackNewsAttachments } : {});
const rawQuota = ref({});
const detailErrorMessage = ref("");
const downloadingAttachmentID = ref("");

const categoryTabs = computed(() => sortNewsCategories(rawCategories.value, fallbackNewsCategories));
const activeTabLabel = computed(() => categoryTabs.value.find((item) => item.key === activeTab.value)?.label || "资讯");
const activeArticleID = computed(() => String(route.query.article || ""));
const currentFeed = computed(() => feedMap.value[activeTab.value] || []);
const newsFeed = computed(() => buildNewsFeedRows(currentFeed.value));
const leadArticle = computed(() => newsFeed.value.lead);
const tickerItems = computed(() => newsFeed.value.tickerItems);
const feedItems = computed(() => newsFeed.value.feedItems);
const activeArticle = computed(() => {
  if (!activeArticleID.value) {
    return null;
  }
  const rows = Object.values(feedMap.value).flat();
  return rows.find((item) => item.id === activeArticleID.value) || null;
});
const activeArticleDetail = computed(() => detailMap.value[activeArticleID.value] || null);
const activeAttachments = computed(() => attachmentMap.value[activeArticleID.value] || []);
const currentArticleNeedsVIP = computed(() => String(activeArticle.value?.visibility || "").toUpperCase() === "VIP");
const hasVipAccess = computed(() => resolveVipStage(rawQuota.value));
const articleAccessState = computed(() => resolveNewsAccessState({
  isLoggedIn: isLoggedIn.value,
  hasVipAccess: hasVipAccess.value,
  visibility: activeArticle.value?.visibility
}));
const currentPagination = computed(() => paginationMap.value[activeTab.value] || { page: 1, total: 0, loading: false, pageSize: DEFAULT_PAGE_SIZE });
const canLoadMore = computed(() => {
  const total = Number(currentPagination.value.total || 0);
  return total > currentFeed.value.length;
});
const detailMetaLine = computed(() => {
  if (!activeArticle.value) {
    return "返回列表后可继续切换栏目。";
  }
  return `${activeArticle.value.meta} · ${currentArticleNeedsVIP.value ? "VIP全文" : "公开内容"}`;
});
const detailTitle = computed(() => activeArticle.value?.title || "文章详情");
const detailCategoryLabel = computed(() => activeArticle.value?.tags?.[0] || activeTabLabel.value);
const detailTags = computed(() => activeArticle.value?.tags?.length ? activeArticle.value.tags : [currentArticleNeedsVIP.value ? "VIP全文" : "公开内容"]);
const leadSummary = computed(() => leadArticle.value?.summary || "先看一条主线资讯，再顺着内容流继续阅读。");
const newsHeroMeta = computed(() => [
  activeTabLabel.value,
  keyword.value ? `关键词 ${keyword.value}` : "",
  leadArticle.value?.meta || ""
].filter(Boolean));
const newsHeroStats = computed(() => [
  {
    label: "当前栏目",
    value: activeTabLabel.value,
    note: leadArticle.value ? "已生成主阅读卡" : "等待内容同步"
  },
  {
    label: "可读内容",
    value: `${currentFeed.value.length} 条`,
    note: keyword.value ? `关键词：${keyword.value}` : "按时间顺序继续浏览"
  },
  {
    label: "阅读权限",
    value: !isLoggedIn.value ? "未登录" : hasVipAccess.value ? "会员已同步" : "登录已同步",
    note: !isLoggedIn.value ? "登录后保留阅读位置" : hasVipAccess.value ? "可继续查看正文与附件" : "VIP 内容仍需升级"
  }
]);
const articleHtml = computed(() => {
  if (articleAccessState.value.locked) {
    return "";
  }
  return renderArticleHTML(activeArticleDetail.value?.content || activeArticleDetail.value?.summary || activeArticle.value?.summary || "");
});

const stickyTitle = computed(() => {
  if (!isLoggedIn.value) {
    return "登录后保留阅读位置并解锁更多详情";
  }
  if (articleAccessState.value.locked) {
    return "开通会员后可查看正文与附件";
  }
  return activeArticleID.value ? "读完这篇后，继续关联策略或升级阅读权限" : "先读资讯主线，再决定是否切去策略";
});

const stickyDescription = computed(() => {
  if (!isLoggedIn.value) {
    return "登录后会按 redirect 回到当前 H5 阅读页。";
  }
  if (articleAccessState.value.locked) {
    return articleAccessState.value.message;
  }
  return activeArticleID.value ? "下一步可以跳去策略页，或者继续返回资讯流。" : "移动端优先保留单列内容节奏，不堆网页式功能块。";
});

const primaryActionLabel = computed(() => {
  if (!isLoggedIn.value) {
    return "立即登录";
  }
  if (articleAccessState.value.locked) {
    return "去会员页";
  }
  return activeArticleID.value ? "看关联策略" : "阅读主线资讯";
});

const secondaryActionLabel = computed(() => currentArticleNeedsVIP.value ? "看策略页" : "返回列表");

function mapArticle(item, tabKey) {
  const visibility = String(item.visibility || "PUBLIC").toUpperCase();
  const categoryLabel = tabKey === "report" ? "研报" : tabKey === "journal" ? "期刊" : tabKey === "book" ? "图书" : "新闻";
  return {
    id: item.id,
    title: shapeNewsDisplayTitle(item.title || "未命名资讯", categoryLabel),
    summary: truncateText(item.summary || toPlainText(item.content) || "-", 82),
    content: item.content || "",
    visibility,
    meta: `${formatDateTime(item.published_at || item.created_at || item.updated_at)} · ${categoryLabel}`,
    badge: visibility === "VIP" ? "VIP" : "公开",
    tone: visibility === "VIP" ? "gold" : "brand",
    tags: [categoryLabel, visibility]
  };
}

async function fetchCategories() {
  try {
    const response = await listNewsCategories();
    rawCategories.value = toArray(response?.items, []);
  } catch {
    rawCategories.value = useDemoFallback ? [...fallbackNewsCategories] : [];
  }
}

async function fetchArticles(tabKey, options = {}) {
  const { page = 1, reset = false } = options;
  paginationMap.value = {
    ...paginationMap.value,
    [tabKey]: {
      ...(paginationMap.value[tabKey] || { page: 1, total: 0, pageSize: DEFAULT_PAGE_SIZE }),
      loading: true
    }
  };

  const query = {
    page,
    page_size: DEFAULT_PAGE_SIZE,
    keyword: keyword.value,
    category_id: tabKey
  };

  try {
    const response = await listNewsArticles(query);
    const nextRows = toArray(response?.items, []).map((item) => mapArticle(item, tabKey));
    const previous = reset ? [] : toArray(feedMap.value[tabKey], []);
    const merged = [...previous, ...nextRows].filter((item, index, list) => list.findIndex((entry) => entry.id === item.id) === index);
    feedMap.value = {
      ...feedMap.value,
      [tabKey]: merged
    };
    paginationMap.value = {
      ...paginationMap.value,
      [tabKey]: {
        page,
        total: Number(response?.total || merged.length),
        pageSize: DEFAULT_PAGE_SIZE,
        loading: false
      }
    };
  } catch (error) {
    if (useDemoFallback) {
      const fallbackRows = fallbackNewsArticles.filter((item) => item.category_id === tabKey).map((item) => mapArticle(item, tabKey));
      feedMap.value = {
        ...feedMap.value,
        [tabKey]: fallbackRows
      };
      paginationMap.value = {
        ...paginationMap.value,
        [tabKey]: {
          page: 1,
          total: fallbackRows.length,
          pageSize: DEFAULT_PAGE_SIZE,
          loading: false
        }
      };
      return;
    }
    paginationMap.value = {
      ...paginationMap.value,
      [tabKey]: {
        ...(paginationMap.value[tabKey] || { page: 1, total: 0, pageSize: DEFAULT_PAGE_SIZE }),
        loading: false
      }
    };
    throw error;
  }
}

async function loadInitialData() {
  loading.value = true;
  const tasks = [fetchCategories()];
  if (isLoggedIn.value) {
    tasks.push(
      getMembershipQuota()
        .then((quota) => {
          rawQuota.value = quota || {};
        })
        .catch(() => {
          rawQuota.value = {};
        })
    );
  } else {
    rawQuota.value = {};
  }
  await Promise.all(tasks);
  const tabs = categoryTabs.value.map((item) => item.key);
  const targetTabs = tabs.length ? tabs : ["news", "report", "journal"];
  await Promise.all(targetTabs.map((tabKey) => fetchArticles(tabKey, { page: 1, reset: true }).catch(() => null)));
  if (!targetTabs.includes(activeTab.value)) {
    activeTab.value = targetTabs[0];
  }
  if (activeArticleID.value) {
    await loadArticleDetail(activeArticleID.value);
  }
  loading.value = false;
}

async function loadArticleDetail(id) {
  const article = Object.values(feedMap.value).flat().find((item) => item.id === id);
  if (!article?.id) {
    detailErrorMessage.value = "当前文章不存在或暂未同步，请返回列表后重新选择。";
    return;
  }
  detailLoading.value = true;
  detailErrorMessage.value = "";
  const accessState = resolveNewsAccessState({
    isLoggedIn: isLoggedIn.value,
    hasVipAccess: hasVipAccess.value,
    visibility: article.visibility
  });
  if (accessState.locked) {
    detailErrorMessage.value = accessState.message;
    detailLoading.value = false;
    return;
  }
  try {
    const [detailResult, attachmentResult] = await Promise.allSettled([
      getNewsArticleDetail(article.id),
      listNewsAttachments(article.id)
    ]);

    if (detailResult.status === "fulfilled" && detailResult.value) {
      detailMap.value = {
        ...detailMap.value,
        [article.id]: detailResult.value
      };
    } else if (detailResult.status === "rejected") {
      detailErrorMessage.value = resolveDetailErrorMessage(article.visibility, detailResult.reason?.message || "详情加载失败");
    }

    if (attachmentResult.status === "fulfilled") {
      attachmentMap.value = {
        ...attachmentMap.value,
        [article.id]: toArray(attachmentResult.value?.items, [])
      };
    } else if (attachmentResult.status === "rejected") {
      detailErrorMessage.value = resolveDetailErrorMessage(article.visibility, attachmentResult.reason?.message || "附件加载失败");
    }
  } finally {
    detailLoading.value = false;
  }
}

function resolveDetailErrorMessage(visibility, fallbackMessage) {
  if (!isLoggedIn.value) {
    return "请先登录，登录后可以保留阅读位置并继续解锁详情。";
  }
  if (String(visibility || "").toUpperCase() === "VIP") {
    return "该内容为 VIP 专享，请开通会员后查看完整正文与附件。";
  }
  return fallbackMessage;
}

function selectTab(tabKey) {
  activeTab.value = tabKey;
  router.replace({ query: { ...route.query, tab: tabKey, article: undefined } });
}

function openArticle(id) {
  router.replace({ query: { ...route.query, tab: activeTab.value, article: id } });
}

function closeArticle() {
  router.replace({ query: { ...route.query, article: undefined } });
}

async function searchArticles() {
  await Promise.all(categoryTabs.value.map((item) => fetchArticles(item.key, { page: 1, reset: true }).catch(() => null)));
  router.replace({ query: { ...route.query, keyword: keyword.value || undefined, article: undefined } });
}

async function loadMore() {
  if (!canLoadMore.value || currentPagination.value.loading) {
    return;
  }
  await fetchArticles(activeTab.value, { page: Number(currentPagination.value.page || 1) + 1, reset: false });
}

async function refreshPage() {
  await loadInitialData();
}

function goArchive() {
  router.push("/archive");
}

async function handleDownloadAttachment(item) {
  if (!item) {
    return;
  }
  const targetID = item.id || item.file_name || "attachment";
  downloadingAttachmentID.value = targetID;
  try {
    let url = item.file_url || "";
    if (item.id && getAccessToken()) {
      const signed = await getAttachmentSignedURL(item.id);
      url = signed?.signed_url || url;
    }
    if (url && typeof window !== "undefined") {
      window.open(url, "_blank", "noopener,noreferrer");
    }
  } finally {
    downloadingAttachmentID.value = "";
  }
}

function goPrimaryAction() {
  if (!isLoggedIn.value) {
    router.push({ path: "/auth", query: { redirect: route.fullPath } });
    return;
  }
  if (articleAccessState.value.locked) {
    router.push("/membership");
    return;
  }
  if (activeArticleID.value) {
    router.push("/strategies");
    return;
  }
  if (leadArticle.value?.id) {
    openArticle(leadArticle.value.id);
  }
}

function goSecondaryAction() {
  if (articleAccessState.value.locked) {
    router.push("/strategies");
    return;
  }
  closeArticle();
}

watch(() => route.query.article, (value) => {
  if (value) {
    loadArticleDetail(String(value));
  }
}, { immediate: false });

watch(() => route.query.tab, (value) => {
  if (value) {
    activeTab.value = String(value);
  }
});

onMounted(() => {
  loadInitialData();
});
</script>

<style scoped>
.news-page {
  gap: 12px;
}

.news-hero-metrics {
  display: grid;
  gap: 10px;
}

.news-stream-hero,
.news-ticker-card,
.news-feed-card,
.news-detail-head,
.news-reading-card,
.news-support-card {
  border: 1px solid var(--h5-line);
  border-radius: var(--h5-radius);
  background: var(--h5-panel-bg);
  box-shadow: var(--h5-shadow);
}

.news-stream-hero,
.news-feed-card,
.news-detail-head,
.news-reading-card,
.news-support-card {
  padding: 20px 18px;
}

.news-stream-hero {
  display: grid;
  gap: 14px;
}

.news-stream-copy {
  display: grid;
  gap: 8px;
}

.news-stream-kicker,
.news-ticker-label,
.news-feed-section {
  color: var(--h5-brand);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.news-stream-copy h2,
.news-detail-head h1 {
  margin: 0;
  color: var(--h5-text);
  font-size: clamp(28px, 7.5vw, 36px);
  line-height: 1.18;
  letter-spacing: -0.03em;
}

.news-stream-copy p,
.news-detail-head p,
.news-feed-head p,
.news-support-head p,
.news-attachment-copy p,
.news-lock-copy p,
.news-inline-tip p {
  margin: 0;
  color: var(--h5-text-sub);
  font-size: 13px;
  line-height: 1.72;
}

.news-refresh-btn {
  justify-self: start;
}

.news-search-bar {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  gap: 10px;
}

.news-search-bar input {
  min-width: 0;
  min-height: 48px;
  padding: 0 16px;
  border: 1px solid var(--h5-line);
  border-radius: var(--h5-radius-sm);
  background: rgba(255, 255, 255, 0.92);
  color: var(--h5-text);
  font-size: 14px;
}

.news-search-bar input::placeholder {
  color: var(--h5-text-soft);
}

.news-search-bar input:focus-visible {
  outline: 2px solid rgba(32, 91, 168, 0.22);
  outline-offset: 2px;
}

.news-category-strip {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 2px;
  scrollbar-width: none;
}

.news-category-strip::-webkit-scrollbar {
  display: none;
}

.news-category-pill {
  flex: 0 0 auto;
  min-height: 38px;
  padding: 0 16px;
  border: 1px solid rgba(16, 42, 86, 0.08);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.92);
  color: var(--h5-text-sub);
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.news-category-pill.active {
  color: #fff;
  background: linear-gradient(180deg, #15345f 0%, #214879 100%);
  border-color: transparent;
  box-shadow: 0 10px 22px rgba(16, 42, 86, 0.16);
}

.news-ticker-card {
  padding: 14px 16px;
  display: grid;
  gap: 10px;
}

.news-ticker-list {
  display: grid;
  gap: 8px;
}

.news-ticker-item {
  position: relative;
  padding-left: 14px;
  color: var(--h5-text-sub);
  font-size: 13px;
  line-height: 1.6;
}

.news-ticker-item::before {
  content: "";
  position: absolute;
  left: 0;
  top: 8px;
  width: 5px;
  height: 5px;
  border-radius: 999px;
  background: var(--h5-gold);
}

.news-lead-card {
  border: 0;
  width: 100%;
  padding: 20px 18px;
  border-radius: 28px;
  background: var(--h5-surface-brand);
  color: #fff;
  display: grid;
  gap: 12px;
  text-align: left;
  cursor: pointer;
  box-shadow: 0 20px 42px rgba(16, 42, 86, 0.18);
}

.news-lead-meta,
.news-detail-meta,
.news-feed-item-head,
.news-feed-item-foot,
.news-support-head,
.news-reading-topline,
.news-lead-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.news-lead-meta,
.news-detail-meta {
  flex-wrap: wrap;
  color: rgba(255, 255, 255, 0.74);
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.05em;
  text-transform: uppercase;
}

.news-lead-card h3,
.news-feed-item strong,
.news-attachment-copy strong,
.news-lock-copy strong,
.news-feed-head strong,
.news-support-head strong {
  margin: 0;
  color: inherit;
  font-size: 22px;
  line-height: 1.34;
}

.news-lead-card p {
  margin: 0;
  color: rgba(255, 255, 255, 0.82);
  font-size: 14px;
  line-height: 1.78;
}

.news-lead-cta {
  color: #fff;
  font-size: 12px;
  font-weight: 700;
}

.news-detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.news-detail-chip,
.news-feed-badge {
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
}

.news-detail-chip {
  border: 1px solid rgba(255, 255, 255, 0.14);
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
}

.news-feed-card {
  display: grid;
  gap: 14px;
}

.news-feed-head {
  align-items: flex-start;
}

.news-feed-head strong,
.news-support-head strong {
  color: var(--h5-text);
  font-size: 17px;
}

.news-feed-head span,
.news-reading-topline span {
  color: var(--h5-text-soft);
  font-size: 12px;
}

.news-feed-list,
.news-attachment-list,
.news-next-actions {
  display: grid;
  gap: 10px;
}

.news-feed-item,
.news-attachment-item {
  padding: 14px 0;
  border: 0;
  border-bottom: 1px solid rgba(16, 42, 86, 0.08);
  background: transparent;
  display: grid;
  gap: 8px;
  text-align: left;
  cursor: pointer;
}

.news-feed-list .news-feed-item:last-child,
.news-attachment-list .news-attachment-item:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}

.news-feed-item:first-child,
.news-attachment-item:first-child {
  padding-top: 0;
}

.news-feed-item strong {
  color: var(--h5-text);
  font-size: 17px;
}

.news-feed-item p {
  margin: 0;
  color: var(--h5-text-sub);
  font-size: 13px;
  line-height: 1.7;
}

.news-feed-time {
  color: var(--h5-text-soft);
  font-size: 12px;
}

.news-feed-badge {
  color: var(--h5-brand);
  background: rgba(20, 52, 95, 0.06);
}

.news-feed-badge.gold {
  color: var(--h5-warning);
  background: rgba(184, 137, 61, 0.12);
}

.news-feed-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  padding-top: 4px;
}

.news-detail-head {
  display: grid;
  gap: 12px;
  background: var(--h5-surface-brand);
  color: #fff;
}

.news-detail-back {
  min-height: 34px;
  width: fit-content;
  padding: 0 14px;
  border: 1px solid rgba(255, 255, 255, 0.16);
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.08);
  color: #fff;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
}

.news-reading-topline strong {
  color: var(--h5-text);
  font-size: 17px;
}

.news-lock-panel,
.news-inline-tip {
  padding: 14px;
  border-radius: var(--h5-radius-md);
  background: rgba(20, 52, 95, 0.045);
  display: grid;
  gap: 12px;
}

.news-lock-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.news-body-prose {
  color: var(--h5-text);
  font-size: 16px;
  line-height: 1.92;
}

.news-body-prose :deep(*) {
  max-width: 100%;
}

.news-body-prose :deep(h1),
.news-body-prose :deep(h2),
.news-body-prose :deep(h3),
.news-body-prose :deep(h4) {
  margin: 1.4em 0 0.65em;
  color: var(--h5-text);
  line-height: 1.38;
}

.news-body-prose :deep(p),
.news-body-prose :deep(li),
.news-body-prose :deep(blockquote) {
  margin: 0 0 1em;
  color: var(--h5-text-sub);
  font-size: 16px;
  line-height: 1.92;
}

.news-body-prose :deep(ul),
.news-body-prose :deep(ol) {
  margin: 0 0 1em;
  padding-left: 1.4em;
}

.news-body-prose :deep(img) {
  display: block;
  width: 100%;
  height: auto;
  margin: 16px 0;
  border-radius: 18px;
}

.news-body-prose :deep(a) {
  color: var(--h5-brand);
  word-break: break-word;
}

.news-attachment-item {
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
}

.news-attachment-copy {
  display: grid;
  gap: 6px;
}

.news-attachment-copy strong {
  color: var(--h5-text);
  font-size: 15px;
}

@media (prefers-reduced-motion: reduce) {
  .news-category-pill,
  .news-feed-item,
  .news-lead-card {
    transition: none;
  }
}
</style>
