<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { createUserMessages, listUserMessages } from "../api/admin";
import { hasPermission } from "../lib/session";

const loading = ref(false);
const sending = ref(false);
const errorMessage = ref("");
const successMessage = ref("");
const sendFailures = ref([]);

const filters = reactive({
  user_id: "",
  type: "",
  read_status: ""
});

const messages = ref([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);

const composeForm = reactive({
  send_mode: "SINGLE",
  user_ids: "",
  type: "SYSTEM",
  title: "",
  content: ""
});

const typeOptions = ["SYSTEM", "STRATEGY", "ALERT", "NEWS"];
const readStatusOptions = ["UNREAD", "READ"];
const canEditUsers = hasPermission("users.edit");

const canSend = computed(() => {
  if (!composeForm.title.trim() || !composeForm.content.trim() || !composeForm.type.trim()) {
    return false;
  }
  if (composeForm.send_mode === "SINGLE") {
    return parseUserIDs(composeForm.user_ids).length > 0;
  }
  return true;
});

function parseUserIDs(raw) {
  const values = String(raw || "")
    .split(/[\s,，;；\n]+/)
    .map((item) => item.trim())
    .filter(Boolean);
  return [...new Set(values)];
}

function mapTypeLabel(value) {
  const normalized = String(value || "").toUpperCase();
  if (normalized === "SYSTEM") return "系统通知";
  if (normalized === "STRATEGY") return "策略提醒";
  if (normalized === "ALERT") return "风险告警";
  if (normalized === "NEWS") return "资讯通知";
  return normalized || "-";
}

function mapReadStatusLabel(value) {
  return String(value || "").toUpperCase() === "READ" ? "已读" : "未读";
}

function readStatusTagType(value) {
  return String(value || "").toUpperCase() === "READ" ? "success" : "warning";
}

function formatDateTime(value) {
  const ts = Date.parse(value || "");
  if (Number.isNaN(ts)) {
    return value || "-";
  }
  return new Date(ts).toLocaleString("zh-CN", { hour12: false });
}

function ensureCanEditUsers() {
  if (canEditUsers) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法发送用户消息";
  return false;
}

async function fetchMessages() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const data = await listUserMessages({
      user_id: filters.user_id.trim(),
      type: filters.type,
      read_status: filters.read_status,
      page: page.value,
      page_size: pageSize.value
    });
    messages.value = data?.items || [];
    total.value = data?.total || 0;
  } catch (error) {
    errorMessage.value = error?.message || "加载用户消息失败";
  } finally {
    loading.value = false;
  }
}

function applyFilters() {
  page.value = 1;
  fetchMessages();
}

function resetFilters() {
  filters.user_id = "";
  filters.type = "";
  filters.read_status = "";
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

async function handleSendMessages() {
  if (!ensureCanEditUsers()) {
    return;
  }
  if (sending.value || !canSend.value) {
    return;
  }

  sending.value = true;
  errorMessage.value = "";
  successMessage.value = "";
  sendFailures.value = [];

  try {
    const userIDs = composeForm.send_mode === "SINGLE" ? parseUserIDs(composeForm.user_ids) : [];
    const data = await createUserMessages({
      user_ids: userIDs,
      title: composeForm.title.trim(),
      content: composeForm.content.trim(),
      type: composeForm.type
    });
    const sentCount = Number(data?.sent_count || 0);
    const failedCount = Number(data?.failed_count || 0);
    sendFailures.value = Array.isArray(data?.failures) ? data.failures : [];
    successMessage.value = composeForm.send_mode === "ALL_ACTIVE"
      ? `已向全部活跃用户推送消息：成功 ${sentCount}，失败 ${failedCount}`
      : `用户消息发送完成：成功 ${sentCount}，失败 ${failedCount}`;
    composeForm.title = "";
    composeForm.content = "";
    await fetchMessages();
  } catch (error) {
    errorMessage.value = error?.message || "发送用户消息失败";
  } finally {
    sending.value = false;
  }
}

onMounted(fetchMessages);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">用户消息中心</h1>
        <p class="muted">统一发送客户端通知，并跟踪消息已读状态</p>
      </div>
      <el-button :loading="loading" @click="fetchMessages">刷新</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />
    <el-alert
      v-if="successMessage"
      :title="successMessage"
      type="success"
      show-icon
      style="margin-bottom: 12px"
    />

    <div v-if="sendFailures.length > 0" class="card" style="margin-bottom: 12px">
      <div class="section-title">发送失败明细（{{ sendFailures.length }}）</div>
      <el-table :data="sendFailures" border stripe size="small" max-height="220">
        <el-table-column prop="user_id" label="用户ID" min-width="140" />
        <el-table-column prop="reason" label="失败原因" min-width="220" />
      </el-table>
    </div>

    <div v-if="canEditUsers" class="card" style="margin-bottom: 12px">
      <div class="section-title">发送消息</div>
      <div class="send-grid">
        <label class="field">
          <span>发送范围</span>
          <el-radio-group v-model="composeForm.send_mode">
            <el-radio-button value="SINGLE">指定用户</el-radio-button>
            <el-radio-button value="ALL_ACTIVE">全部活跃用户</el-radio-button>
          </el-radio-group>
        </label>

        <label v-if="composeForm.send_mode === 'SINGLE'" class="field">
          <span>用户ID（支持逗号/换行）</span>
          <el-input
            v-model="composeForm.user_ids"
            type="textarea"
            :rows="2"
            placeholder="例如：u_001, u_002"
          />
        </label>

        <label class="field">
          <span>消息类型</span>
          <el-select v-model="composeForm.type" style="width: 180px">
            <el-option v-for="item in typeOptions" :key="item" :label="mapTypeLabel(item)" :value="item" />
          </el-select>
        </label>

        <label class="field">
          <span>标题</span>
          <el-input v-model="composeForm.title" maxlength="128" placeholder="请输入消息标题" />
        </label>

        <label class="field">
          <span>内容</span>
          <el-input
            v-model="composeForm.content"
            type="textarea"
            :rows="3"
            maxlength="1000"
            show-word-limit
            placeholder="请输入消息内容"
          />
        </label>
      </div>
      <div class="send-actions">
        <el-button type="primary" :disabled="!canSend" :loading="sending" @click="handleSendMessages">
          发送消息
        </el-button>
      </div>
    </div>

    <div class="card">
      <div class="toolbar">
        <el-input v-model="filters.user_id" clearable placeholder="按用户ID筛选" style="width: 180px" />
        <el-select v-model="filters.type" clearable placeholder="全部类型" style="width: 180px">
          <el-option v-for="item in typeOptions" :key="item" :label="mapTypeLabel(item)" :value="item" />
        </el-select>
        <el-select v-model="filters.read_status" clearable placeholder="全部已读状态" style="width: 180px">
          <el-option
            v-for="item in readStatusOptions"
            :key="item"
            :label="mapReadStatusLabel(item)"
            :value="item"
          />
        </el-select>
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>

      <el-table :data="messages" border stripe v-loading="loading" empty-text="暂无用户消息">
        <el-table-column prop="id" label="消息ID" min-width="180" />
        <el-table-column prop="user_id" label="用户ID" min-width="140" />
        <el-table-column label="手机号" min-width="130">
          <template #default="{ row }">
            {{ row.user_phone || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="类型" min-width="110">
          <template #default="{ row }">
            {{ mapTypeLabel(row.type) }}
          </template>
        </el-table-column>
        <el-table-column prop="title" label="标题" min-width="180" />
        <el-table-column label="内容" min-width="260">
          <template #default="{ row }">
            <span class="content-cell">{{ row.content || "-" }}</span>
          </template>
        </el-table-column>
        <el-table-column label="已读状态" min-width="100">
          <template #default="{ row }">
            <el-tag :type="readStatusTagType(row.read_status)">
              {{ mapReadStatusLabel(row.read_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
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
.section-title {
  margin-bottom: 10px;
  font-size: 15px;
  font-weight: 600;
}

.send-grid {
  display: grid;
  gap: 10px;
}

.field {
  display: grid;
  gap: 6px;
}

.field > span {
  color: #4b5563;
  font-size: 13px;
}

.send-actions {
  margin-top: 10px;
  display: flex;
  justify-content: flex-end;
}

.content-cell {
  display: -webkit-box;
  overflow: hidden;
  text-overflow: ellipsis;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  line-height: 1.5;
}
</style>
