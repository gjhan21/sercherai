<script setup>
import { onMounted, reactive, ref } from "vue";
import { ElMessageBox } from "element-plus";
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
const savingCategory = ref(false);
const savingArticle = ref(false);
const savingAttachment = ref(false);

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

const visibilityOptions = ["PUBLIC", "VIP"];
const statusOptions = ["DRAFT", "PUBLISHED", "DISABLED"];

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
  errorMessage.value = "";
  try {
    const data = await listNewsCategories({ page: 1, page_size: 200 });
    categories.value = data.items || [];
    if (!articleForm.category_id && categories.value.length > 0) {
      articleForm.category_id = categories.value[0].id;
    }
  } catch (error) {
    errorMessage.value = error.message || "加载新闻分类失败";
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
  savingCategory.value = true;
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
  } finally {
    savingCategory.value = false;
  }
}

async function submitArticle() {
  savingArticle.value = true;
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
  } finally {
    savingArticle.value = false;
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
  savingAttachment.value = true;
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
  } finally {
    savingAttachment.value = false;
  }
}

async function handleDeleteAttachment(id) {
  try {
    await ElMessageBox.confirm(`确认删除附件 ${id}？`, "删除确认", {
      type: "warning",
      confirmButtonText: "删除",
      cancelButtonText: "取消"
    });
  } catch {
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

function handleArticlePageChange(nextPage) {
  if (nextPage === articlePage.value) {
    return;
  }
  articlePage.value = nextPage;
  fetchArticles();
}

function resolveCategoryName(categoryID) {
  const found = categories.value.find((item) => item.id === categoryID);
  return found?.name || categoryID || "-";
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "PUBLISHED" || normalized === "ACTIVE") return "success";
  if (normalized === "DISABLED" || normalized === "REJECTED") return "danger";
  if (normalized === "DRAFT" || normalized === "PENDING") return "warning";
  return "info";
}

function visibilityTagType(visibility) {
  return visibility === "VIP" ? "danger" : "success";
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
        <el-button :loading="loadingCategories" @click="fetchCategories">刷新分类</el-button>
        <el-button :loading="loadingArticles" @click="fetchArticles">刷新文章</el-button>
      </div>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />
    <el-alert
      v-if="message"
      :title="message"
      type="success"
      show-icon
      style="margin-bottom: 12px"
    />

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <h3 style="margin: 0">新闻分类</h3>
        <el-button type="primary" @click="openCreateCategory">新增分类</el-button>
      </div>

      <el-table :data="categories" border stripe v-loading="loadingCategories" empty-text="暂无分类">
        <el-table-column prop="id" label="ID" min-width="130" />
        <el-table-column prop="name" label="名称" min-width="160" />
        <el-table-column prop="slug" label="slug" min-width="140" />
        <el-table-column prop="sort" label="排序" min-width="90" />
        <el-table-column label="可见性" min-width="110">
          <template #default="{ row }">
            <el-tag :type="visibilityTagType(row.visibility)">{{ row.visibility }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" align="right" min-width="120">
          <template #default="{ row }">
            <el-button size="small" @click="openEditCategory(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <h3 style="margin: 0">新闻文章</h3>
        <el-button type="primary" @click="openCreateArticle">新增文章</el-button>
      </div>

      <div class="toolbar" style="margin-bottom: 12px">
        <el-select v-model="articleFilters.status" clearable placeholder="全部状态" style="width: 160px">
          <el-option value="DRAFT" label="DRAFT" />
          <el-option value="PUBLISHED" label="PUBLISHED" />
          <el-option value="DISABLED" label="DISABLED" />
        </el-select>
        <el-select v-model="articleFilters.category_id" clearable placeholder="全部分类" style="width: 200px">
          <el-option v-for="category in categories" :key="category.id" :label="category.name" :value="category.id" />
        </el-select>
        <el-button type="primary" plain @click="applyArticleFilters">查询</el-button>
        <el-button @click="resetArticleFilters">重置</el-button>
      </div>

      <el-table :data="articles" border stripe v-loading="loadingArticles" empty-text="暂无文章">
        <el-table-column prop="id" label="ID" min-width="130" />
        <el-table-column label="分类" min-width="140">
          <template #default="{ row }">
            {{ resolveCategoryName(row.category_id) }}
          </template>
        </el-table-column>
        <el-table-column label="标题" min-width="260">
          <template #default="{ row }">
            <div class="article-title">{{ row.title }}</div>
            <div class="article-summary">{{ row.summary || "-" }}</div>
          </template>
        </el-table-column>
        <el-table-column label="可见性" min-width="110">
          <template #default="{ row }">
            <el-tag :type="visibilityTagType(row.visibility)">{{ row.visibility }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="published_at" label="发布时间" min-width="180">
          <template #default="{ row }">
            {{ row.published_at || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="author_id" label="作者" min-width="140">
          <template #default="{ row }">
            {{ row.author_id || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="操作" align="right" min-width="230">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-button size="small" @click="openEditArticle(row)">编辑</el-button>
              <el-button
                size="small"
                type="primary"
                plain
                :disabled="row.status === 'PUBLISHED'"
                @click="handlePublishArticle(row)"
              >
                发布
              </el-button>
              <el-button size="small" @click="fetchAttachments(row)">附件</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-text type="info">第 {{ articlePage }} 页，共 {{ totalArticles }} 条</el-text>
        <el-pagination
          background
          layout="prev, pager, next"
          :current-page="articlePage"
          :page-size="articlePageSize"
          :total="totalArticles"
          @current-change="handleArticlePageChange"
        />
      </div>
    </div>

    <div class="card">
      <div class="section-header">
        <h3 style="margin: 0">附件管理 {{ selectedArticle ? `(文章: ${selectedArticle.id})` : "" }}</h3>
        <el-button v-if="selectedArticle" :loading="loadingAttachments" @click="fetchAttachments(selectedArticle)">
          刷新附件
        </el-button>
      </div>

      <div v-if="selectedArticle" class="attachment-editor">
        <el-form label-width="90px">
          <div class="dialog-grid">
            <el-form-item label="文件名" required>
              <el-input v-model="attachmentForm.file_name" placeholder="report.pdf" />
            </el-form-item>
            <el-form-item label="文件URL" required>
              <el-input v-model="attachmentForm.file_url" placeholder="https://example.com/report.pdf" />
            </el-form-item>
            <el-form-item label="文件大小" required>
              <el-input-number v-model="attachmentForm.file_size" :min="1" :step="1024" controls-position="right" />
            </el-form-item>
            <el-form-item label="MIME Type">
              <el-input v-model="attachmentForm.mime_type" placeholder="application/pdf" />
            </el-form-item>
          </div>
        </el-form>
        <div class="toolbar" style="margin-bottom: 0">
          <el-button type="primary" :loading="savingAttachment" @click="submitAttachment">新增附件</el-button>
          <el-button @click="resetAttachmentForm">清空</el-button>
        </div>
      </div>
      <el-empty v-else description="请在上方文章列表中点击“附件”选择文章" />

      <el-table
        :data="attachments"
        border
        stripe
        v-loading="loadingAttachments"
        empty-text="暂无附件"
        style="margin-top: 12px"
      >
        <el-table-column prop="id" label="ID" min-width="130" />
        <el-table-column prop="file_name" label="文件名" min-width="150" />
        <el-table-column prop="file_url" label="URL" min-width="220">
          <template #default="{ row }">
            <el-link :href="row.file_url" target="_blank" type="primary">{{ row.file_url }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="file_size" label="大小(bytes)" min-width="120" />
        <el-table-column prop="mime_type" label="MIME" min-width="140">
          <template #default="{ row }">
            {{ row.mime_type || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ row.created_at || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="操作" align="right" min-width="120">
          <template #default="{ row }">
            <el-button size="small" type="danger" plain @click="handleDeleteAttachment(row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog
      v-model="categoryFormVisible"
      :title="categoryFormMode === 'create' ? '新增新闻分类' : `编辑分类：${categoryForm.id}`"
      width="620px"
      destroy-on-close
    >
      <el-form label-width="96px">
        <div class="dialog-grid">
          <el-form-item label="名称" required>
            <el-input v-model="categoryForm.name" placeholder="资讯快讯" />
          </el-form-item>
          <el-form-item label="slug" required>
            <el-input v-model="categoryForm.slug" placeholder="flash-news" />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="categoryForm.sort" :step="1" controls-position="right" />
          </el-form-item>
          <el-form-item label="可见性">
            <el-select v-model="categoryForm.visibility">
              <el-option v-for="item in visibilityOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="categoryForm.status">
              <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="categoryFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingCategory" @click="submitCategory">
          {{ categoryFormMode === "create" ? "创建分类" : "更新分类" }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="articleFormVisible"
      :title="articleFormMode === 'create' ? '新增新闻文章' : `编辑文章：${articleForm.id}`"
      width="860px"
      destroy-on-close
    >
      <el-form label-width="90px">
        <div class="dialog-grid">
          <el-form-item label="分类" required>
            <el-select v-model="articleForm.category_id" placeholder="请选择分类">
              <el-option v-for="category in categories" :key="category.id" :label="category.name" :value="category.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="标题" required>
            <el-input v-model="articleForm.title" placeholder="请输入标题" />
          </el-form-item>
          <el-form-item label="摘要">
            <el-input v-model="articleForm.summary" placeholder="请输入摘要" />
          </el-form-item>
          <el-form-item label="可见性">
            <el-select v-model="articleForm.visibility">
              <el-option v-for="item in visibilityOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="articleForm.status">
              <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
        </div>
        <el-form-item label="正文" required>
          <el-input
            v-model="articleForm.content"
            type="textarea"
            :rows="8"
            resize="vertical"
            placeholder="请输入正文内容"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="articleFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="savingArticle" @click="submitArticle">
          {{ articleFormMode === "create" ? "创建文章" : "更新文章" }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.article-title {
  line-height: 1.4;
}

.article-summary {
  margin-top: 2px;
  color: #6b7280;
  font-size: 12px;
}

.inline-actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
}

.attachment-editor {
  margin-bottom: 12px;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  background: #f9fafb;
}

.dialog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: 0 12px;
}

:deep(.dialog-grid .el-form-item) {
  margin-bottom: 14px;
}

:deep(.dialog-grid .el-select),
:deep(.dialog-grid .el-input-number) {
  width: 100%;
}
</style>
