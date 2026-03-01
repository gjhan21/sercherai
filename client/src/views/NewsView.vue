<template>
  <section class="news-page fade-up">
    <header class="news-hero card">
      <div>
        <p class="hero-kicker">资讯中心 · API 联动</p>
        <h1 class="section-title">先看新闻，再看研报和期刊。</h1>
        <p class="section-subtitle">已接入资讯分类、文章详情和附件接口，支持二级导航切换。</p>
      </div>
      <div class="hero-status">
        <p>更新状态</p>
        <strong>{{ loading ? "同步中" : "已同步" }}</strong>
      </div>
    </header>

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
        <header>
          <h2 class="section-title">{{ currentCategory.label }}</h2>
          <p class="section-subtitle">{{ currentCategory.subtitle }}</p>
        </header>

        <div class="feed-list" v-if="currentFeed.length">
          <article
            v-for="item in currentFeed"
            :key="item.id"
            class="feed-item"
            :class="{ active: activeArticleID === item.id }"
            @click="openArticle(item.id)"
          >
            <div class="meta">
              <span class="time">{{ item.time }}</span>
              <span class="level" :class="item.levelClass">{{ item.level }}</span>
            </div>
            <h3>{{ item.title }}</h3>
            <p>{{ item.desc }}</p>
            <div class="tags">
              <span v-for="tag in item.tags" :key="tag">{{ tag }}</span>
            </div>
          </article>
        </div>
        <div v-else class="empty-box">当前栏目暂无内容</div>
        <div v-if="currentFeed.length" class="feed-footer">
          <p class="feed-progress">{{ feedProgressText }}</p>
          <button
            type="button"
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
        <article class="card detail-card">
          <header class="detail-head">
            <div>
              <p class="decision-tag">{{ currentCategory.label }}详情</p>
              <h2>文章详情</h2>
            </div>
            <button type="button" :disabled="detailLoading" @click="refreshActiveArticleDetail">
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
              <p v-else class="empty-inline">暂无附件</p>
            </div>
          </div>
          <div v-else class="empty-box">暂无可查看文章</div>

          <p v-if="detailErrorMessage" class="detail-error">{{ detailErrorMessage }}</p>
        </article>
      </aside>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import {
  getAttachmentSignedURL,
  getNewsArticleDetail,
  listNewsArticles,
  listNewsAttachments,
  listNewsCategories
} from "../api/news";
import { getAccessToken } from "../lib/session";

const categoryTemplates = [
  {
    key: "news",
    label: "新闻",
    subtitle: "聚合市场动态与政策消息，突出当日高影响事件。"
  },
  {
    key: "report",
    label: "研报",
    subtitle: "聚焦机构深度分析，帮助中短期策略参数优化。"
  },
  {
    key: "journal",
    label: "期刊",
    subtitle: "沉淀长期观点和方法论，支持策略体系迭代。"
  }
];

const categoryTemplateMap = categoryTemplates.reduce((result, item) => {
  result[item.key] = item;
  return result;
}, {});
const defaultCategoryTemplate = categoryTemplates[0];
const DEFAULT_PAGE_SIZE = 20;

const route = useRoute();

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
      errors.push("分类接口返回为空，请先在管理端创建并发布新闻分类");
    }
  } catch (error) {
    errors.push(`分类接口失败：${error?.message || "unknown error"}`);
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
        errors.push(`${category.label}接口失败：${error?.message || "unknown error"}`);
      }
    })
  );

  if (errors.length > 0) {
    errorMessage.value = errors.join("；");
  }
  lastUpdatedAt.value = formatDateTime(new Date().toISOString());
  loading.value = false;
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
  loadNewsData();
});
</script>

<style scoped>
.news-page {
  display: grid;
  gap: 12px;
}

.news-hero {
  border-radius: 20px;
  padding: 16px;
  display: grid;
  grid-template-columns: 1fr auto;
  align-items: end;
  gap: 10px;
}

.hero-kicker {
  margin: 0;
  color: var(--color-pine-600);
  font-size: 12px;
}

.hero-status {
  border-radius: 12px;
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(246, 244, 239, 0.9);
  padding: 8px 10px;
}

.hero-status p {
  margin: 0;
  font-size: 11px;
  color: var(--color-text-sub);
}

.hero-status strong {
  margin-top: 2px;
  display: block;
  color: var(--color-pine-700);
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
  border: 1px solid rgba(216, 223, 216, 0.9);
  background: rgba(252, 251, 247, 0.9);
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
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
  border-color: transparent;
}

.sub-item.active .label,
.sub-item.active .count {
  color: #fff;
}

.api-tip {
  border-radius: 10px;
  border: 1px dashed rgba(216, 223, 216, 0.95);
  background: rgba(246, 244, 239, 0.76);
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
  border: 0;
  border-radius: 10px;
  padding: 8px 12px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  color: #fff;
  background: linear-gradient(145deg, var(--color-pine-700), var(--color-pine-500));
}

.feed-footer button:disabled {
  opacity: 0.72;
  cursor: not-allowed;
}

.feed-item {
  border: 1px solid rgba(216, 223, 216, 0.9);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.9);
  padding: 10px;
  cursor: pointer;
}

.feed-item.active {
  border-color: rgba(63, 127, 113, 0.4);
  box-shadow: 0 8px 20px rgba(63, 127, 113, 0.12);
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
  border-radius: 999px;
  padding: 2px 8px;
  font-size: 11px;
}

.level.high {
  color: #8a3c2f;
  background: rgba(230, 194, 185, 0.65);
}

.level.mid {
  color: #775325;
  background: rgba(234, 215, 180, 0.66);
}

.level.low {
  color: var(--color-pine-700);
  background: rgba(223, 236, 230, 0.72);
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
  font-size: 11px;
  border-radius: 999px;
  padding: 2px 8px;
  background: rgba(239, 232, 218, 0.75);
  color: #775325;
}

.side {
  display: grid;
  gap: 12px;
  align-content: start;
}

.detail-card {
  background: linear-gradient(160deg, rgba(38, 88, 78, 0.96), rgba(45, 108, 95, 0.96));
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
  color: rgba(245, 250, 247, 0.76);
}

.detail-head h2 {
  margin: 7px 0 0;
  font-family: var(--font-serif);
  font-size: 22px;
}

.detail-head button {
  border: 0;
  border-radius: 10px;
  background: #f1dfbf;
  color: #664a26;
  padding: 8px 12px;
  font-weight: 600;
  cursor: pointer;
}

.detail-head button:disabled {
  opacity: 0.7;
  cursor: not-allowed;
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
  color: rgba(245, 250, 247, 0.74);
}

.detail-summary {
  margin: 10px 0 0;
  font-size: 13px;
  line-height: 1.62;
  color: rgba(245, 250, 247, 0.84);
}

.detail-content {
  margin: 8px 0 0;
  font-size: 13px;
  line-height: 1.62;
  color: rgba(245, 250, 247, 0.9);
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
  color: #f1dfbf;
  text-decoration: underline;
}

.detail-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 8px;
}

.attachment-box {
  margin-top: 10px;
  border-top: 1px solid rgba(241, 223, 191, 0.34);
  padding-top: 10px;
}

.attachment-title {
  margin: 0;
  font-size: 12px;
  color: rgba(245, 250, 247, 0.76);
}

.attachment-list {
  margin-top: 8px;
  display: grid;
  gap: 6px;
}

.attachment-item {
  border: 1px solid rgba(241, 223, 191, 0.4);
  background: rgba(246, 225, 188, 0.12);
  color: #f5faf7;
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
  color: rgba(245, 250, 247, 0.72);
}

.empty-inline {
  margin: 8px 0 0;
  font-size: 12px;
  color: rgba(245, 250, 247, 0.7);
}

.detail-error {
  margin: 10px 0 0;
  color: #f1dfbf;
  font-size: 12px;
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

@media (max-width: 980px) {
  .news-hero,
  .news-grid {
    grid-template-columns: 1fr;
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
