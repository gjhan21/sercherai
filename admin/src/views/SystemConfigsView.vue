<script setup>
import { onMounted, reactive, ref } from "vue";
import { listSystemConfigs, upsertSystemConfig } from "../api/admin";

const loading = ref(false);
const submitting = ref(false);
const errorMessage = ref("");
const message = ref("");

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const items = ref([]);

const filters = reactive({
  keyword: ""
});

const dialogVisible = ref(false);
const form = reactive({
  config_key: "",
  config_value: "",
  description: ""
});

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function resetForm() {
  Object.assign(form, {
    config_key: "",
    config_value: "",
    description: ""
  });
}

async function fetchConfigs(options = {}) {
  const { keepMessage = false } = options;
  loading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }

  try {
    const data = await listSystemConfigs({
      keyword: filters.keyword.trim(),
      page: page.value,
      page_size: pageSize.value
    });
    items.value = data.items || [];
    total.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载系统配置失败");
  } finally {
    loading.value = false;
  }
}

function openCreateDialog() {
  resetForm();
  dialogVisible.value = true;
}

function openEditDialog(item) {
  Object.assign(form, {
    config_key: item.config_key || "",
    config_value: item.config_value || "",
    description: item.description || ""
  });
  dialogVisible.value = true;
}

async function submitForm() {
  const payload = {
    config_key: form.config_key.trim(),
    config_value: form.config_value,
    description: form.description.trim()
  };

  if (!payload.config_key || payload.config_value === "") {
    errorMessage.value = "config_key 和 config_value 不能为空";
    return;
  }

  submitting.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    await upsertSystemConfig(payload);
    dialogVisible.value = false;
    await fetchConfigs({ keepMessage: true });
    message.value = `系统配置 ${payload.config_key} 已保存`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "保存系统配置失败");
  } finally {
    submitting.value = false;
  }
}

function applyFilters() {
  page.value = 1;
  fetchConfigs();
}

function resetFilters() {
  filters.keyword = "";
  page.value = 1;
  fetchConfigs();
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  fetchConfigs();
}

onMounted(fetchConfigs);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">系统配置</h1>
        <p class="muted">维护后台运行参数，支持按关键字检索与更新</p>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <el-button :loading="loading" @click="fetchConfigs">刷新</el-button>
        <el-button type="primary" @click="openCreateDialog">新增配置</el-button>
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
      <div class="toolbar" style="margin-bottom: 0">
        <el-input v-model="filters.keyword" clearable placeholder="按 key/description 关键词检索" style="width: 260px" />
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </div>

    <div class="card">
      <el-table :data="items" border stripe v-loading="loading" empty-text="暂无系统配置">
        <el-table-column prop="config_key" label="配置键" min-width="220" />
        <el-table-column prop="config_value" label="配置值" min-width="320" />
        <el-table-column prop="description" label="描述" min-width="220" />
        <el-table-column prop="updated_by" label="更新人" min-width="120" />
        <el-table-column prop="updated_at" label="更新时间" min-width="180" />
        <el-table-column label="操作" align="right" min-width="120">
          <template #default="{ row }">
            <el-button size="small" @click="openEditDialog(row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-text type="info">第 {{ page }} 页，共 {{ total }} 条</el-text>
        <el-pagination
          background
          layout="prev, pager, next"
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <el-dialog v-model="dialogVisible" title="系统配置" width="620px" destroy-on-close>
      <el-form label-width="110px">
        <el-form-item label="配置键" required>
          <el-input v-model="form.config_key" placeholder="如 model.default" />
        </el-form-item>
        <el-form-item label="配置值" required>
          <el-input v-model="form.config_value" type="textarea" :rows="4" placeholder="配置内容" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" placeholder="用途说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>
