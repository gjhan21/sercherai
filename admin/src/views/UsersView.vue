<script setup>
import { onMounted, reactive, ref } from "vue";
import {
  listUsers,
  updateUserKYCStatus,
  updateUserMemberLevel,
  updateUserStatus
} from "../api/admin";

const loading = ref(false);
const errorMessage = ref("");
const message = ref("");

const filters = reactive({
  status: "",
  kyc_status: "",
  member_level: ""
});

const users = ref([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);

const draftStatusMap = ref({});
const draftKYCMap = ref({});
const draftLevelMap = ref({});

function syncDrafts() {
  const statusMap = {};
  const kycMap = {};
  const levelMap = {};
  users.value.forEach((user) => {
    statusMap[user.id] = user.status || "ACTIVE";
    kycMap[user.id] = user.kyc_status || "PENDING";
    levelMap[user.id] = user.member_level || "FREE";
  });
  draftStatusMap.value = statusMap;
  draftKYCMap.value = kycMap;
  draftLevelMap.value = levelMap;
}

async function fetchUsers() {
  loading.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const data = await listUsers({
      status: filters.status,
      kyc_status: filters.kyc_status,
      member_level: filters.member_level,
      page: page.value,
      page_size: pageSize.value
    });
    users.value = data.items || [];
    total.value = data.total || 0;
    syncDrafts();
  } catch (error) {
    errorMessage.value = error.message || "加载用户失败";
  } finally {
    loading.value = false;
  }
}

async function handleUpdateStatus(user) {
  const target = draftStatusMap.value[user.id];
  if (!target || target === user.status) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await updateUserStatus(user.id, target);
    message.value = `用户 ${user.id} 状态已更新为 ${target}`;
    await fetchUsers();
  } catch (error) {
    errorMessage.value = error.message || "更新用户状态失败";
  }
}

async function handleUpdateKYC(user) {
  const target = draftKYCMap.value[user.id];
  if (!target || target === user.kyc_status) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await updateUserKYCStatus(user.id, target);
    message.value = `用户 ${user.id} KYC 状态已更新为 ${target}`;
    await fetchUsers();
  } catch (error) {
    errorMessage.value = error.message || "更新 KYC 状态失败";
  }
}

async function handleUpdateMemberLevel(user) {
  const target = (draftLevelMap.value[user.id] || "").trim();
  if (!target || target === user.member_level) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await updateUserMemberLevel(user.id, target);
    message.value = `用户 ${user.id} 会员等级已更新为 ${target}`;
    await fetchUsers();
  } catch (error) {
    errorMessage.value = error.message || "更新会员等级失败";
  }
}

function applyFilters() {
  page.value = 1;
  fetchUsers();
}

function resetFilters() {
  filters.status = "";
  filters.kyc_status = "";
  filters.member_level = "";
  page.value = 1;
  fetchUsers();
}

function nextPage() {
  if (page.value * pageSize.value >= total.value) {
    return;
  }
  page.value += 1;
  fetchUsers();
}

function prevPage() {
  if (page.value <= 1) {
    return;
  }
  page.value -= 1;
  fetchUsers();
}

onMounted(fetchUsers);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">用户管理</h1>
        <p class="muted">用户状态、实名状态、会员等级维护</p>
      </div>
      <button class="btn" :disabled="loading" @click="fetchUsers">
        {{ loading ? "刷新中..." : "刷新" }}
      </button>
    </div>

    <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
    <div v-if="message" class="success-message">{{ message }}</div>

    <div class="card" style="margin-bottom: 12px">
      <div class="toolbar">
        <select v-model="filters.status" class="select">
          <option value="">全部用户状态</option>
          <option value="ACTIVE">ACTIVE</option>
          <option value="DISABLED">DISABLED</option>
          <option value="BANNED">BANNED</option>
        </select>
        <select v-model="filters.kyc_status" class="select">
          <option value="">全部 KYC 状态</option>
          <option value="PENDING">PENDING</option>
          <option value="APPROVED">APPROVED</option>
          <option value="REJECTED">REJECTED</option>
        </select>
        <input v-model="filters.member_level" class="input" placeholder="会员等级，如 VIP1" />
        <button class="btn" @click="applyFilters">查询</button>
        <button class="btn" @click="resetFilters">重置</button>
      </div>
    </div>

    <div class="card">
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>用户ID</th>
              <th>手机号</th>
              <th>邮箱</th>
              <th>用户状态</th>
              <th>KYC 状态</th>
              <th>会员等级</th>
              <th>创建时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td>{{ user.id }}</td>
              <td>{{ user.phone }}</td>
              <td>{{ user.email || "-" }}</td>
              <td>
                <div class="toolbar">
                  <select v-model="draftStatusMap[user.id]" class="select">
                    <option value="ACTIVE">ACTIVE</option>
                    <option value="DISABLED">DISABLED</option>
                    <option value="BANNED">BANNED</option>
                  </select>
                  <button class="btn" @click="handleUpdateStatus(user)">保存</button>
                </div>
              </td>
              <td>
                <div class="toolbar">
                  <select v-model="draftKYCMap[user.id]" class="select">
                    <option value="PENDING">PENDING</option>
                    <option value="APPROVED">APPROVED</option>
                    <option value="REJECTED">REJECTED</option>
                  </select>
                  <button class="btn" @click="handleUpdateKYC(user)">保存</button>
                </div>
              </td>
              <td>
                <div class="toolbar">
                  <input v-model="draftLevelMap[user.id]" class="input" />
                  <button class="btn" @click="handleUpdateMemberLevel(user)">保存</button>
                </div>
              </td>
              <td>{{ user.created_at }}</td>
            </tr>
            <tr v-if="!loading && users.length === 0">
              <td colspan="7" class="muted">暂无用户数据</td>
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
