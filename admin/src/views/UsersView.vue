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

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
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
      <el-button :loading="loading" @click="fetchUsers">刷新</el-button>
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
        <el-select v-model="filters.status" clearable placeholder="全部用户状态" style="width: 160px">
          <el-option label="ACTIVE" value="ACTIVE" />
          <el-option label="DISABLED" value="DISABLED" />
          <el-option label="BANNED" value="BANNED" />
        </el-select>
        <el-select v-model="filters.kyc_status" clearable placeholder="全部 KYC 状态" style="width: 170px">
          <el-option label="PENDING" value="PENDING" />
          <el-option label="APPROVED" value="APPROVED" />
          <el-option label="REJECTED" value="REJECTED" />
        </el-select>
        <el-input v-model="filters.member_level" clearable placeholder="会员等级，如 VIP1" style="width: 180px" />
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </div>

    <div class="card">
      <el-table :data="users" border stripe v-loading="loading" empty-text="暂无用户数据">
        <el-table-column prop="id" label="用户ID" min-width="160" />
        <el-table-column prop="phone" label="手机号" min-width="130" />
        <el-table-column label="邮箱" min-width="180">
          <template #default="{ row }">
            {{ row.email || "-" }}
          </template>
        </el-table-column>

        <el-table-column label="用户状态" min-width="220">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-select v-model="draftStatusMap[row.id]" style="width: 120px">
                <el-option label="ACTIVE" value="ACTIVE" />
                <el-option label="DISABLED" value="DISABLED" />
                <el-option label="BANNED" value="BANNED" />
              </el-select>
              <el-button size="small" @click="handleUpdateStatus(row)">保存</el-button>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="KYC 状态" min-width="220">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-select v-model="draftKYCMap[row.id]" style="width: 120px">
                <el-option label="PENDING" value="PENDING" />
                <el-option label="APPROVED" value="APPROVED" />
                <el-option label="REJECTED" value="REJECTED" />
              </el-select>
              <el-button size="small" @click="handleUpdateKYC(row)">保存</el-button>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="会员等级" min-width="240">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-input v-model="draftLevelMap[row.id]" style="width: 120px" />
              <el-button size="small" @click="handleUpdateMemberLevel(row)">保存</el-button>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="创建时间" min-width="180" />
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
.inline-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
