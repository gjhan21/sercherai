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

function nextPage() {
  if (page.value * pageSize.value >= total.value) {
    return;
  }
  page.value += 1;
  fetchMessages();
}

function prevPage() {
  if (page.value <= 1) {
    return;
  }
  page.value -= 1;
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
          <span class="badge">{{ unreadCount }}</span>
        </h1>
        <p class="muted">查看审核/任务/数据源告警消息，并处理已读状态</p>
      </div>
      <div class="toolbar">
        <button class="btn" :disabled="loading" @click="fetchMessages">刷新</button>
        <button class="btn btn-primary" :disabled="loading" @click="markAllRead">全部标记已读</button>
      </div>
    </div>

    <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
    <div v-if="message" class="success-message">{{ message }}</div>

    <div class="card" style="margin-bottom: 12px">
      <div class="toolbar">
        <input v-model="filters.module" class="input" placeholder="模块，如 SYSTEM / STOCK" />
        <input v-model="filters.event_type" class="input" placeholder="事件，如 DATA_SOURCE_UNHEALTHY" />
        <select v-model="filters.is_read" class="select">
          <option value="">全部状态</option>
          <option value="false">未读</option>
          <option value="true">已读</option>
        </select>
        <button class="btn" @click="applyFilters">查询</button>
        <button class="btn" @click="resetFilters">重置</button>
      </div>
    </div>

    <div class="card">
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>事件</th>
              <th>模块</th>
              <th>标题</th>
              <th>内容</th>
              <th>接收人</th>
              <th>状态</th>
              <th>创建时间</th>
              <th class="text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.id">
              <td>{{ item.event_type }}</td>
              <td>{{ item.module }}</td>
              <td>{{ item.title }}</td>
              <td>{{ item.content }}</td>
              <td>{{ item.receiver_id || "-" }}</td>
              <td>
                <span class="status-tag" :class="item.is_read ? 'status-active' : 'status-disabled'">
                  {{ item.is_read ? "已读" : "未读" }}
                </span>
              </td>
              <td>{{ item.created_at }}</td>
              <td class="text-right">
                <button class="btn" @click="toggleRead(item)">
                  {{ item.is_read ? "取消已读" : "标记已读" }}
                </button>
              </td>
            </tr>
            <tr v-if="!loading && items.length === 0">
              <td colspan="8" class="muted">暂无流程消息</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="pagination">
        <span>第 {{ page }} 页，共 {{ total }} 条</span>
        <div class="toolbar">
          <button class="btn" :disabled="page <= 1 || loading" @click="prevPage">上一页</button>
          <button class="btn" :disabled="page * pageSize >= total || loading" @click="nextPage">
            下一页
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
