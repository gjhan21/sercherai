<script setup>
import { onMounted, reactive, ref } from "vue";
import {
  bulkReadWorkflowMessages,
  countUnreadWorkflowMessages,
  listWorkflowMessages,
  updateWorkflowMessageRead
} from "../api/admin";

const loading = ref(false);
const errorMessage = ref("");
const message = ref("");

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const unreadCount = ref(0);

const items = ref([]);
const filters = reactive({
  module: "",
  event_type: "",
  is_read: ""
});

async function fetchUnreadCount() {
  try {
    const data = await countUnreadWorkflowMessages({
      module: filters.module,
      event_type: filters.event_type
    });
    unreadCount.value = data.unread_count || 0;
  } catch {
    unreadCount.value = 0;
  }
}

async function fetchMessages() {
  loading.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const data = await listWorkflowMessages({
      module: filters.module,
      event_type: filters.event_type,
      is_read: filters.is_read,
      page: page.value,
      page_size: pageSize.value
    });
    items.value = data.items || [];
    total.value = data.total || 0;
    await fetchUnreadCount();
  } catch (error) {
    errorMessage.value = error.message || "加载流程消息失败";
  } finally {
    loading.value = false;
  }
}

async function toggleRead(item) {
  errorMessage.value = "";
  message.value = "";
  try {
    await updateWorkflowMessageRead(item.id, !item.is_read);
    message.value = `消息 ${item.id} 已${item.is_read ? "取消已读" : "标记已读"}`;
    await fetchMessages();
  } catch (error) {
    errorMessage.value = error.message || "更新状态失败";
  }
}

async function markAllRead() {
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await bulkReadWorkflowMessages({
      module: filters.module,
      event_type: filters.event_type
    });
    message.value = `批量已读完成，影响 ${result.affected || 0} 条`;
    await fetchMessages();
  } catch (error) {
    errorMessage.value = error.message || "批量已读失败";
  }
}

function resetFilters() {
  filters.module = "";
  filters.event_type = "";
  filters.is_read = "";
  page.value = 1;
  fetchMessages();
}

function applyFilters() {
  page.value = 1;
  fetchMessages();
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  fetchMessages();
}

onMounted(fetchMessages);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">
          流程消息
          <el-badge :value="unreadCount" class="title-badge" />
        </h1>
        <p class="muted">查看审核/任务/数据源告警消息，并处理已读状态</p>
      </div>
      <div class="toolbar">
        <el-button :loading="loading" @click="fetchMessages">刷新</el-button>
        <el-button type="primary" :loading="loading" @click="markAllRead">全部标记已读</el-button>
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
      <div class="toolbar">
        <el-input v-model="filters.module" clearable placeholder="模块，如 SYSTEM / STOCK" style="width: 220px" />
        <el-input
          v-model="filters.event_type"
          clearable
          placeholder="事件，如 DATA_SOURCE_UNHEALTHY"
          style="width: 260px"
        />
        <el-select v-model="filters.is_read" clearable placeholder="全部状态" style="width: 130px">
          <el-option label="未读" value="false" />
          <el-option label="已读" value="true" />
        </el-select>
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </div>

    <div class="card">
      <el-table :data="items" border stripe v-loading="loading" empty-text="暂无流程消息">
        <el-table-column prop="event_type" label="事件" min-width="170" />
        <el-table-column prop="module" label="模块" min-width="100" />
        <el-table-column prop="title" label="标题" min-width="160" />
        <el-table-column prop="content" label="内容" min-width="240" />
        <el-table-column label="接收人" min-width="140">
          <template #default="{ row }">
            {{ row.receiver_id || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="100">
          <template #default="{ row }">
            <el-tag :type="row.is_read ? 'success' : 'warning'">
              {{ row.is_read ? "已读" : "未读" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" min-width="180" />
        <el-table-column label="操作" min-width="140" align="right">
          <template #default="{ row }">
            <el-button size="small" @click="toggleRead(row)">
              {{ row.is_read ? "取消已读" : "标记已读" }}
            </el-button>
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
  </div>
</template>

<style scoped>
.title-badge {
  margin-left: 8px;
  vertical-align: middle;
}
</style>
