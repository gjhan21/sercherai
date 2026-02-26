<script setup>
import { onMounted, reactive, ref } from "vue";
import {
  createNewsArticle,
  createNewsAttachment,
  createNewsCategory,
  deleteNewsAttachment,
  listNewsArticles,
  listNewsAttachments,
  listNewsCategories,
  publishNewsArticle,
  updateNewsArticle,
  updateNewsCategory
} from "../api/admin";

const loadingCategories = ref(false);
const loadingArticles = ref(false);
const loadingAttachments = ref(false);

const errorMessage = ref("");
const message = ref("");

const categories = ref([]);
const articles = ref([]);
const totalArticles = ref(0);

const articlePage = ref(1);
const articlePageSize = ref(20);

const articleFilters = reactive({
  status: "",
  category_id: ""
});

const categoryFormVisible = ref(false);
const categoryFormMode = ref("create");
const categoryForm = reactive({
  id: "",
  name: "",
  slug: "",
  sort: 0,
  visibility: "PUBLIC",
  status: "DRAFT"
});

const articleFormVisible = ref(false);
const articleFormMode = ref("create");
const articleForm = reactive({
  id: "",
  category_id: "",
  title: "",
  summary: "",
  content: "",
  visibility: "PUBLIC",
  status: "DRAFT"
});

const selectedArticle = ref(null);
const attachments = ref([]);
const attachmentForm = reactive({
  file_name: "",
  file_url: "",
  file_size: 0,
  mime_type: ""
});

function resetCategoryForm() {
  Object.assign(categoryForm, {
    id: "",
    name: "",
    slug: "",
    sort: 0,
    visibility: "PUBLIC",
    status: "DRAFT"
  });
  categoryFormMode.value = "create";
}

function resetArticleForm() {
  Object.assign(articleForm, {
    id: "",
    category_id: categories.value[0]?.id || "",
    title: "",
    summary: "",
    content: "",
    visibility: "PUBLIC",
    status: "DRAFT"
  });
  articleFormMode.value = "create";
}

function resetAttachmentForm() {
  Object.assign(attachmentForm, {
    file_name: "",
    file_url: "",
    file_size: 0,
    mime_type: ""
  });
}

async function fetchCategories() {
  loadingCategories.value = true;
  try {
    const data = await listNewsCategories({ page: 1, page_size: 200 });
    categories.value = data.items || [];
    if (!articleForm.category_id && categories.value.length > 0) {
      articleForm.category_id = categories.value[0].id;
    }
  } finally {
    loadingCategories.value = false;
  }
}

async function fetchArticles() {
  loadingArticles.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const data = await listNewsArticles({
      status: articleFilters.status,
      category_id: articleFilters.category_id,
      page: articlePage.value,
      page_size: articlePageSize.value
    });
    articles.value = data.items || [];
    totalArticles.value = data.total || 0;
  } catch (error) {
    errorMessage.value = error.message || "加载新闻文章失败";
  } finally {
    loadingArticles.value = false;
  }
}

async function fetchAttachments(article) {
  if (!article?.id) {
    selectedArticle.value = null;
    attachments.value = [];
    return;
  }
  selectedArticle.value = article;
  loadingAttachments.value = true;
  errorMessage.value = "";
  try {
    const data = await listNewsAttachments(article.id);
    attachments.value = data.items || [];
  } catch (error) {
    errorMessage.value = error.message || "加载附件失败";
    attachments.value = [];
  } finally {
    loadingAttachments.value = false;
  }
}

async function submitCategory() {
  errorMessage.value = "";
  message.value = "";
  try {
    const payload = {
      name: categoryForm.name.trim(),
      slug: categoryForm.slug.trim(),
      sort: Number(categoryForm.sort) || 0,
      visibility: categoryForm.visibility,
      status: categoryForm.status
    };
    if (!payload.name || !payload.slug) {
      throw new Error("分类名称和 slug 不能为空");
    }
    if (categoryFormMode.value === "create") {
      await createNewsCategory(payload);
      message.value = "新闻分类创建成功";
    } else {
      await updateNewsCategory(categoryForm.id, payload);
      message.value = "新闻分类更新成功";
    }
    categoryFormVisible.value = false;
    resetCategoryForm();
    await fetchCategories();
  } catch (error) {
    errorMessage.value = error.message || "提交新闻分类失败";
  }
}

async function submitArticle() {
  errorMessage.value = "";
  message.value = "";
  try {
    const payload = {
      category_id: articleForm.category_id,
      title: articleForm.title.trim(),
      summary: articleForm.summary.trim(),
      content: articleForm.content.trim(),
      visibility: articleForm.visibility,
      status: articleForm.status
    };
    if (!payload.category_id || !payload.title || !payload.content) {
      throw new Error("分类、标题、正文不能为空");
    }
    if (articleFormMode.value === "create") {
      await createNewsArticle(payload);
      message.value = "新闻文章创建成功";
    } else {
      await updateNewsArticle(articleForm.id, payload);
      message.value = "新闻文章更新成功";
    }
    articleFormVisible.value = false;
    resetArticleForm();
    await fetchArticles();
  } catch (error) {
    errorMessage.value = error.message || "提交新闻文章失败";
  }
}

async function handlePublishArticle(article) {
  errorMessage.value = "";
  message.value = "";
  try {
    await publishNewsArticle(article.id);
    message.value = `文章 ${article.id} 已发布`;
    await fetchArticles();
  } catch (error) {
    errorMessage.value = error.message || "发布文章失败";
  }
}

async function submitAttachment() {
  if (!selectedArticle.value?.id) {
    errorMessage.value = "请先选择文章";
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    const payload = {
      file_name: attachmentForm.file_name.trim(),
      file_url: attachmentForm.file_url.trim(),
      file_size: Number(attachmentForm.file_size) || 0,
      mime_type: attachmentForm.mime_type.trim()
    };
    if (!payload.file_name || !payload.file_url || payload.file_size <= 0) {
      throw new Error("附件名称、URL、大小不能为空");
    }
    await createNewsAttachment(selectedArticle.value.id, payload);
    message.value = `文章 ${selectedArticle.value.id} 附件添加成功`;
    resetAttachmentForm();
    await fetchAttachments(selectedArticle.value);
  } catch (error) {
    errorMessage.value = error.message || "添加附件失败";
  }
}

async function handleDeleteAttachment(id) {
  if (!window.confirm(`确认删除附件 ${id}？`)) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await deleteNewsAttachment(id);
    message.value = `附件 ${id} 已删除`;
    await fetchAttachments(selectedArticle.value);
  } catch (error) {
    errorMessage.value = error.message || "删除附件失败";
  }
}

function openCreateCategory() {
  resetCategoryForm();
  categoryFormVisible.value = true;
}

function openEditCategory(category) {
  Object.assign(categoryForm, {
    id: category.id,
    name: category.name || "",
    slug: category.slug || "",
    sort: Number(category.sort) || 0,
    visibility: category.visibility || "PUBLIC",
    status: category.status || "DRAFT"
  });
  categoryFormMode.value = "edit";
  categoryFormVisible.value = true;
}

function openCreateArticle() {
  resetArticleForm();
  articleFormVisible.value = true;
}

function openEditArticle(article) {
  Object.assign(articleForm, {
    id: article.id,
    category_id: article.category_id || categories.value[0]?.id || "",
    title: article.title || "",
    summary: article.summary || "",
    content: article.content || "",
    visibility: article.visibility || "PUBLIC",
    status: article.status || "DRAFT"
  });
  articleFormMode.value = "edit";
  articleFormVisible.value = true;
}

function applyArticleFilters() {
  articlePage.value = 1;
  fetchArticles();
}

function resetArticleFilters() {
  articleFilters.status = "";
  articleFilters.category_id = "";
  articlePage.value = 1;
  fetchArticles();
}

function nextArticlePage() {
  if (articlePage.value * articlePageSize.value >= totalArticles.value) {
    return;
  }
  articlePage.value += 1;
  fetchArticles();
}

function prevArticlePage() {
  if (articlePage.value <= 1) {
    return;
  }
  articlePage.value -= 1;
  fetchArticles();
}

function resolveCategoryName(categoryID) {
  const found = categories.value.find((item) => item.id === categoryID);
  return found?.name || categoryID || "-";
}

onMounted(async () => {
  try {
    await fetchCategories();
    await fetchArticles();
  } catch (error) {
    errorMessage.value = error.message || "初始化新闻管理失败";
  }
});
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">新闻管理</h1>
        <p class="muted">分类管理、文章管理、发布与附件维护</p>
      </div>
      <div class="toolbar">
        <button class="btn" :disabled="loadingCategories || loadingArticles" @click="fetchCategories">
          刷新分类
        </button>
        <button class="btn" :disabled="loadingArticles" @click="fetchArticles">刷新文章</button>
      </div>
    </div>

    <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
    <div v-if="message" class="success-message">{{ message }}</div>

    <div class="card" style="margin-bottom: 12px">
      <div class="page-header" style="margin-bottom: 10px">
        <h3 style="margin: 0">新闻分类</h3>
        <button class="btn btn-primary" @click="openCreateCategory">新增分类</button>
      </div>

      <div v-if="categoryFormVisible" class="card" style="margin-bottom: 12px">
        <div class="form-grid">
          <div class="form-item">
            <label>名称</label>
            <input v-model="categoryForm.name" class="input" placeholder="资讯快讯" />
          </div>
          <div class="form-item">
            <label>slug</label>
            <input v-model="categoryForm.slug" class="input" placeholder="flash-news" />
          </div>
          <div class="form-item">
            <label>排序</label>
            <input v-model.number="categoryForm.sort" class="input" type="number" />
          </div>
          <div class="form-item">
            <label>可见性</label>
            <select v-model="categoryForm.visibility" class="select">
              <option value="PUBLIC">PUBLIC</option>
              <option value="VIP">VIP</option>
            </select>
          </div>
          <div class="form-item">
            <label>状态</label>
            <select v-model="categoryForm.status" class="select">
              <option value="DRAFT">DRAFT</option>
              <option value="PUBLISHED">PUBLISHED</option>
              <option value="DISABLED">DISABLED</option>
            </select>
          </div>
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" @click="submitCategory">
            {{ categoryFormMode === "create" ? "创建分类" : "更新分类" }}
          </button>
          <button class="btn" @click="categoryFormVisible = false">取消</button>
        </div>
      </div>

      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>名称</th>
              <th>slug</th>
              <th>排序</th>
              <th>可见性</th>
              <th>状态</th>
              <th class="text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="category in categories" :key="category.id">
              <td>{{ category.id }}</td>
              <td>{{ category.name }}</td>
              <td>{{ category.slug }}</td>
              <td>{{ category.sort }}</td>
              <td>{{ category.visibility }}</td>
              <td>{{ category.status }}</td>
              <td class="text-right">
                <button class="btn" @click="openEditCategory(category)">编辑</button>
              </td>
            </tr>
            <tr v-if="!loadingCategories && categories.length === 0">
              <td colspan="7" class="muted">暂无分类</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="page-header" style="margin-bottom: 10px">
        <h3 style="margin: 0">新闻文章</h3>
        <button class="btn btn-primary" @click="openCreateArticle">新增文章</button>
      </div>

      <div class="toolbar" style="margin-bottom: 12px">
        <select v-model="articleFilters.status" class="select">
          <option value="">全部状态</option>
          <option value="DRAFT">DRAFT</option>
          <option value="PUBLISHED">PUBLISHED</option>
          <option value="DISABLED">DISABLED</option>
        </select>
        <select v-model="articleFilters.category_id" class="select">
          <option value="">全部分类</option>
          <option v-for="category in categories" :key="category.id" :value="category.id">
            {{ category.name }}
          </option>
        </select>
        <button class="btn" @click="applyArticleFilters">查询</button>
        <button class="btn" @click="resetArticleFilters">重置</button>
      </div>

      <div v-if="articleFormVisible" class="card" style="margin-bottom: 12px">
        <div class="form-grid">
          <div class="form-item">
            <label>分类</label>
            <select v-model="articleForm.category_id" class="select">
              <option v-for="category in categories" :key="category.id" :value="category.id">
                {{ category.name }}
              </option>
            </select>
          </div>
          <div class="form-item">
            <label>标题</label>
            <input v-model="articleForm.title" class="input" placeholder="请输入标题" />
          </div>
          <div class="form-item">
            <label>摘要</label>
            <input v-model="articleForm.summary" class="input" placeholder="请输入摘要" />
          </div>
          <div class="form-item">
            <label>可见性</label>
            <select v-model="articleForm.visibility" class="select">
              <option value="PUBLIC">PUBLIC</option>
              <option value="VIP">VIP</option>
            </select>
          </div>
          <div class="form-item">
            <label>状态</label>
            <select v-model="articleForm.status" class="select">
              <option value="DRAFT">DRAFT</option>
              <option value="PUBLISHED">PUBLISHED</option>
              <option value="DISABLED">DISABLED</option>
            </select>
          </div>
        </div>
        <div class="form-item" style="margin-top: 10px">
          <label>正文</label>
          <textarea v-model="articleForm.content" class="textarea" rows="6" placeholder="请输入正文内容" />
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" @click="submitArticle">
            {{ articleFormMode === "create" ? "创建文章" : "更新文章" }}
          </button>
          <button class="btn" @click="articleFormVisible = false">取消</button>
        </div>
      </div>

      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>分类</th>
              <th>标题</th>
              <th>可见性</th>
              <th>状态</th>
              <th>发布时间</th>
              <th>作者</th>
              <th class="text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="article in articles" :key="article.id">
              <td>{{ article.id }}</td>
              <td>{{ resolveCategoryName(article.category_id) }}</td>
              <td>
                <div>{{ article.title }}</div>
                <div class="muted">{{ article.summary || "-" }}</div>
              </td>
              <td>{{ article.visibility }}</td>
              <td>{{ article.status }}</td>
              <td>{{ article.published_at || "-" }}</td>
              <td>{{ article.author_id || "-" }}</td>
              <td class="text-right">
                <div class="toolbar" style="justify-content: flex-end">
                  <button class="btn" @click="openEditArticle(article)">编辑</button>
                  <button class="btn" @click="handlePublishArticle(article)">发布</button>
                  <button class="btn" @click="fetchAttachments(article)">附件</button>
                </div>
              </td>
            </tr>
            <tr v-if="!loadingArticles && articles.length === 0">
              <td colspan="8" class="muted">暂无文章</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination">
        <span>第 {{ articlePage }} 页，共 {{ totalArticles }} 条</span>
        <div class="toolbar">
          <button class="btn" :disabled="articlePage <= 1 || loadingArticles" @click="prevArticlePage">
            上一页
          </button>
          <button
            class="btn"
            :disabled="articlePage * articlePageSize >= totalArticles || loadingArticles"
            @click="nextArticlePage"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="page-header" style="margin-bottom: 10px">
        <h3 style="margin: 0">
          附件管理 {{ selectedArticle ? `(文章: ${selectedArticle.id})` : "" }}
        </h3>
      </div>

      <div v-if="selectedArticle" class="card" style="margin-bottom: 12px">
        <div class="form-grid">
          <div class="form-item">
            <label>文件名</label>
            <input v-model="attachmentForm.file_name" class="input" placeholder="report.pdf" />
          </div>
          <div class="form-item">
            <label>文件URL</label>
            <input v-model="attachmentForm.file_url" class="input" placeholder="https://example.com/report.pdf" />
          </div>
          <div class="form-item">
            <label>文件大小(bytes)</label>
            <input v-model.number="attachmentForm.file_size" class="input" type="number" min="1" />
          </div>
          <div class="form-item">
            <label>MIME Type</label>
            <input v-model="attachmentForm.mime_type" class="input" placeholder="application/pdf" />
          </div>
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" @click="submitAttachment">新增附件</button>
          <button class="btn" @click="resetAttachmentForm">清空</button>
        </div>
      </div>

      <div v-else class="hint">请在上方文章列表中点击“附件”选择文章。</div>

      <div class="table-wrap" style="margin-top: 10px">
        <table class="table">
          <thead>
            <tr>
              <th>ID</th>
              <th>文件名</th>
              <th>URL</th>
              <th>大小</th>
              <th>MIME</th>
              <th>创建时间</th>
              <th class="text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in attachments" :key="item.id">
              <td>{{ item.id }}</td>
              <td>{{ item.file_name }}</td>
              <td>{{ item.file_url }}</td>
              <td>{{ item.file_size }}</td>
              <td>{{ item.mime_type || "-" }}</td>
              <td>{{ item.created_at || "-" }}</td>
              <td class="text-right">
                <button class="btn btn-danger" @click="handleDeleteAttachment(item.id)">删除</button>
              </td>
            </tr>
            <tr v-if="!loadingAttachments && attachments.length === 0">
              <td colspan="7" class="muted">暂无附件</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
