<script setup>
import { computed, nextTick, onMounted, reactive, ref } from "vue";
import {
  listUsers,
  updateUserKYCStatus,
  updateUserMemberLevel,
  updateUserStatus
} from "../api/admin";
import { getAccessToken } from "../lib/session";

const loading = ref(false);
const exportingFiltered = ref(false);
const retryingFailed = ref(false);
const copyingFailedDetails = ref(false);

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

const userTableRef = ref(null);
const selectedRows = ref([]);

const draftStatusMap = ref({});
const draftKYCMap = ref({});
const draftLevelMap = ref({});

const batchStatus = ref("ACTIVE");
const batchKYCStatus = ref("APPROVED");
const batchMemberLevel = ref("VIP1");
const batchUpdatingStatus = ref(false);
const batchUpdatingKYC = ref(false);
const batchUpdatingLevel = ref(false);

const batchResultVisible = ref(false);
const batchResultTitle = ref("");
const batchResultRows = ref([]);
const batchResultFilter = ref("all");

const statusOptions = ["ACTIVE", "DISABLED", "BANNED"];
const kycStatusOptions = ["PENDING", "APPROVED", "REJECTED"];

const selectedCount = computed(() => selectedRows.value.length);
const failedBatchRows = computed(() => batchResultRows.value.filter((row) => row.result === "FAILED"));
const canBatchUpdateStatus = computed(() => selectedCount.value > 0 && batchStatus.value.trim() !== "");
const canBatchUpdateKYC = computed(() => selectedCount.value > 0 && batchKYCStatus.value.trim() !== "");
const canBatchUpdateLevel = computed(() => selectedCount.value > 0 && batchMemberLevel.value.trim() !== "");
const batchResultStats = computed(() => {
  const stats = {
    total: batchResultRows.value.length,
    success: 0,
    failed: 0,
    skipped: 0
  };
  batchResultRows.value.forEach((row) => {
    const result = (row.result || "").toUpperCase();
    if (result === "SUCCESS") {
      stats.success += 1;
      return;
    }
    if (result === "FAILED") {
      stats.failed += 1;
      return;
    }
    if (result === "SKIPPED") {
      stats.skipped += 1;
    }
  });
  return stats;
});
const displayBatchResultRows = computed(() => {
  if (batchResultFilter.value === "failed") {
    return batchResultRows.value.filter((row) => row.result === "FAILED");
  }
  if (batchResultFilter.value === "skipped") {
    return batchResultRows.value.filter((row) => row.result === "SKIPPED");
  }
  return batchResultRows.value;
});

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

function clearSelection() {
  selectedRows.value = [];
  nextTick(() => {
    userTableRef.value?.clearSelection();
  });
}

function handleSelectionChange(rows) {
  selectedRows.value = rows || [];
}

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

async function fetchUsers(options = {}) {
  const { keepMessage = false } = options;
  loading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
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
    clearSelection();
  } catch (error) {
    errorMessage.value = error.message || "加载用户失败";
  } finally {
    loading.value = false;
  }
}

async function handleUpdateStatus(user) {
  const target = (draftStatusMap.value[user.id] || "").trim();
  if (!target || target === user.status) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await updateUserStatus(user.id, target);
    await fetchUsers({ keepMessage: true });
    message.value = `用户 ${user.id} 状态已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = error.message || "更新用户状态失败";
  }
}

async function handleUpdateKYC(user) {
  const target = (draftKYCMap.value[user.id] || "").trim();
  if (!target || target === user.kyc_status) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await updateUserKYCStatus(user.id, target);
    await fetchUsers({ keepMessage: true });
    message.value = `用户 ${user.id} KYC 状态已更新为 ${target}`;
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
    await fetchUsers({ keepMessage: true });
    message.value = `用户 ${user.id} 会员等级已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = error.message || "更新会员等级失败";
  }
}

function openBatchResultDialog(title, rows) {
  batchResultTitle.value = title;
  batchResultRows.value = rows;
  batchResultFilter.value = "all";
  batchResultVisible.value = true;
}

async function runBatchUpdate(config) {
  const target = (config.targetValue || "").trim();
  if (!target) {
    errorMessage.value = config.emptyTargetMessage || "请先设置目标值";
    return;
  }
  if (selectedRows.value.length <= 0) {
    errorMessage.value = "请先勾选用户";
    return;
  }

  errorMessage.value = "";
  message.value = "";
  config.loadingRef.value = true;

  let success = 0;
  let failed = 0;
  let skipped = 0;
  const resultRows = [];

  try {
    for (const user of selectedRows.value) {
      const currentValue = (config.currentValue(user) || "").trim();
      if (currentValue === target) {
        skipped += 1;
        resultRows.push({
          id: user.id,
          action: config.action,
          action_key: config.actionKey,
          target,
          result: "SKIPPED",
          reason: config.skippedReason || "已是目标值"
        });
        continue;
      }

      try {
        await config.executor(user.id, target);
        success += 1;
        resultRows.push({
          id: user.id,
          action: config.action,
          action_key: config.actionKey,
          target,
          result: "SUCCESS",
          reason: config.successReason(target)
        });
      } catch (error) {
        failed += 1;
        resultRows.push({
          id: user.id,
          action: config.action,
          action_key: config.actionKey,
          target,
          result: "FAILED",
          reason: normalizeErrorMessage(error, "更新失败")
        });
      }
    }
  } finally {
    config.loadingRef.value = false;
  }

  await fetchUsers({ keepMessage: true });
  message.value = `${config.action}完成：成功 ${success}，失败 ${failed}，跳过 ${skipped}`;
  openBatchResultDialog(config.title, resultRows);
}

async function handleBatchUpdateStatus() {
  await runBatchUpdate({
    title: "批量更新用户状态结果",
    action: "批量更新状态",
    actionKey: "USER_STATUS",
    targetValue: batchStatus.value,
    loadingRef: batchUpdatingStatus,
    emptyTargetMessage: "请先选择目标用户状态",
    currentValue: (user) => user.status,
    skippedReason: "用户状态已是目标值",
    successReason: (target) => `状态已更新为 ${target}`,
    executor: updateUserStatus
  });
}

async function handleBatchUpdateKYC() {
  await runBatchUpdate({
    title: "批量更新 KYC 状态结果",
    action: "批量更新KYC",
    actionKey: "KYC_STATUS",
    targetValue: batchKYCStatus.value,
    loadingRef: batchUpdatingKYC,
    emptyTargetMessage: "请先选择目标 KYC 状态",
    currentValue: (user) => user.kyc_status,
    skippedReason: "KYC 状态已是目标值",
    successReason: (target) => `KYC 状态已更新为 ${target}`,
    executor: updateUserKYCStatus
  });
}

async function handleBatchUpdateLevel() {
  await runBatchUpdate({
    title: "批量更新会员等级结果",
    action: "批量更新会员等级",
    actionKey: "MEMBER_LEVEL",
    targetValue: batchMemberLevel.value,
    loadingRef: batchUpdatingLevel,
    emptyTargetMessage: "请先输入目标会员等级",
    currentValue: (user) => user.member_level,
    skippedReason: "会员等级已是目标值",
    successReason: (target) => `会员等级已更新为 ${target}`,
    executor: updateUserMemberLevel
  });
}

async function executeBatchResultRow(row) {
  const actionKey = (row.action_key || "").toUpperCase();
  const target = (row.target || "").trim();
  if (!target) {
    throw new Error("缺少目标值");
  }

  if (actionKey === "USER_STATUS") {
    await updateUserStatus(row.id, target);
    return `状态已更新为 ${target}`;
  }
  if (actionKey === "KYC_STATUS") {
    await updateUserKYCStatus(row.id, target);
    return `KYC 状态已更新为 ${target}`;
  }
  if (actionKey === "MEMBER_LEVEL") {
    await updateUserMemberLevel(row.id, target);
    return `会员等级已更新为 ${target}`;
  }
  throw new Error(`未知动作: ${actionKey || "-"}`);
}

async function retryFailedBatchRows() {
  if (failedBatchRows.value.length <= 0) {
    return;
  }

  retryingFailed.value = true;
  errorMessage.value = "";
  message.value = "";

  let success = 0;
  let failed = 0;
  const resultRows = [];

  try {
    for (const row of failedBatchRows.value) {
      try {
        const tip = await executeBatchResultRow(row);
        success += 1;
        resultRows.push({
          ...row,
          action: `${row.action}重试`,
          result: "SUCCESS",
          reason: tip
        });
      } catch (error) {
        failed += 1;
        resultRows.push({
          ...row,
          action: `${row.action}重试`,
          result: "FAILED",
          reason: normalizeErrorMessage(error, "重试失败")
        });
      }
    }
  } finally {
    retryingFailed.value = false;
  }

  await fetchUsers({ keepMessage: true });
  message.value = `失败任务重试完成：成功 ${success}，失败 ${failed}`;
  openBatchResultDialog("失败任务重试结果", resultRows);
}

function buildFailedDetailsText() {
  return failedBatchRows.value
    .map((row) => {
      return [
        `用户ID=${row.id || "-"}`,
        `动作=${row.action || "-"}`,
        `目标=${row.target || "-"}`,
        `结果=${row.result || "-"}`,
        `原因=${row.reason || "-"}`
      ].join(" | ");
    })
    .join("\n");
}

async function copyFailedDetails() {
  if (failedBatchRows.value.length <= 0) {
    return;
  }
  const text = buildFailedDetailsText();
  if (!text) {
    return;
  }

  copyingFailedDetails.value = true;
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
    } else {
      const textarea = document.createElement("textarea");
      textarea.value = text;
      textarea.style.position = "fixed";
      textarea.style.opacity = "0";
      document.body.appendChild(textarea);
      textarea.focus();
      textarea.select();
      document.execCommand("copy");
      document.body.removeChild(textarea);
    }
    message.value = `已复制失败明细，共 ${failedBatchRows.value.length} 条`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "复制失败明细失败");
  } finally {
    copyingFailedDetails.value = false;
  }
}

function csvEscape(value) {
  const text = String(value ?? "");
  if (/[",\n]/.test(text)) {
    return `"${text.replace(/"/g, '""')}"`;
  }
  return text;
}

function triggerCSVDownload(content, fileName) {
  const blob = new Blob([`\uFEFF${content}`], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = fileName;
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  URL.revokeObjectURL(url);
}

function buildCSVRows(items) {
  const header = ["id", "phone", "email", "status", "kyc_status", "member_level", "created_at"];
  const rows = items.map((item) => [
    item.id || "",
    item.phone || "",
    item.email || "",
    item.status || "",
    item.kyc_status || "",
    item.member_level || "",
    item.created_at || ""
  ]);
  return [header, ...rows].map((row) => row.map(csvEscape).join(",")).join("\n");
}

function exportCurrentPageCSV() {
  const csv = buildCSVRows(users.value);
  const fileName = `admin_users_page_${new Date().toISOString().slice(0, 10)}.csv`;
  triggerCSVDownload(csv, fileName);
  message.value = `已导出当前页 CSV，共 ${users.value.length} 条`;
}

async function exportFilteredCSV() {
  exportingFiltered.value = true;
  errorMessage.value = "";
  message.value = "";

  try {
    const params = new URLSearchParams();
    if (filters.status) params.set("status", filters.status);
    if (filters.kyc_status) params.set("kyc_status", filters.kyc_status);
    if (filters.member_level.trim()) params.set("member_level", filters.member_level.trim());

    const baseURL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
    const query = params.toString();
    const requestURL = `${baseURL}/admin/users/export.csv${query ? `?${query}` : ""}`;

    const headers = {};
    const token = getAccessToken();
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    const response = await fetch(requestURL, { method: "GET", headers });
    if (!response.ok) {
      const text = await response.text();
      throw new Error(text || `导出失败(${response.status})`);
    }

    const blob = await response.blob();
    const blobURL = URL.createObjectURL(blob);
    const fileName = `admin_users_filtered_${new Date().toISOString().slice(0, 10)}.csv`;
    const anchor = document.createElement("a");
    anchor.href = blobURL;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(blobURL);
    message.value = "已发起筛选结果 CSV 下载";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "导出筛选结果失败");
  } finally {
    exportingFiltered.value = false;
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

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "ACTIVE") return "success";
  if (normalized === "DISABLED" || normalized === "BANNED") return "danger";
  return "info";
}

function kycTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "APPROVED") return "success";
  if (normalized === "REJECTED") return "danger";
  if (normalized === "PENDING") return "warning";
  return "info";
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
          <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-select v-model="filters.kyc_status" clearable placeholder="全部 KYC 状态" style="width: 170px">
          <el-option v-for="item in kycStatusOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-input v-model="filters.member_level" clearable placeholder="会员等级，如 VIP1" style="width: 180px" />
        <el-button :loading="exportingFiltered" @click="exportFilteredCSV">导出筛选CSV</el-button>
        <el-button @click="exportCurrentPageCSV">导出当前页CSV</el-button>
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <h3 style="margin: 0">批量操作</h3>
        <el-text type="info">已勾选 {{ selectedCount }} 个用户</el-text>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <div class="field-stack">
          <el-select v-model="batchStatus" style="width: 140px">
            <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
          </el-select>
          <el-popconfirm
            width="300"
            :title="`确认将选中用户状态批量更新为 ${batchStatus || '-'} 吗？`"
            @confirm="handleBatchUpdateStatus"
          >
            <template #reference>
              <el-button :loading="batchUpdatingStatus" :disabled="!canBatchUpdateStatus">批量更新状态</el-button>
            </template>
          </el-popconfirm>
        </div>

        <div class="field-stack">
          <el-select v-model="batchKYCStatus" style="width: 140px">
            <el-option v-for="item in kycStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
          <el-popconfirm
            width="300"
            :title="`确认将选中用户 KYC 状态批量更新为 ${batchKYCStatus || '-'} 吗？`"
            @confirm="handleBatchUpdateKYC"
          >
            <template #reference>
              <el-button :loading="batchUpdatingKYC" :disabled="!canBatchUpdateKYC">批量更新KYC</el-button>
            </template>
          </el-popconfirm>
        </div>

        <div class="field-stack">
          <el-input v-model="batchMemberLevel" placeholder="目标会员等级" style="width: 140px" />
          <el-popconfirm
            width="300"
            :title="`确认将选中用户会员等级批量更新为 ${batchMemberLevel || '-'} 吗？`"
            @confirm="handleBatchUpdateLevel"
          >
            <template #reference>
              <el-button :loading="batchUpdatingLevel" :disabled="!canBatchUpdateLevel">批量更新等级</el-button>
            </template>
          </el-popconfirm>
        </div>

        <el-button @click="clearSelection">清空勾选</el-button>
      </div>
    </div>

    <div class="card">
      <el-table
        ref="userTableRef"
        :data="users"
        row-key="id"
        border
        stripe
        v-loading="loading"
        empty-text="暂无用户数据"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="52" reserve-selection />
        <el-table-column prop="id" label="用户ID" min-width="160" />
        <el-table-column prop="phone" label="手机号" min-width="130" />
        <el-table-column label="邮箱" min-width="180">
          <template #default="{ row }">
            {{ row.email || "-" }}
          </template>
        </el-table-column>

        <el-table-column label="用户状态" min-width="300">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-tag :type="statusTagType(row.status)">{{ row.status || "-" }}</el-tag>
              <el-select v-model="draftStatusMap[row.id]" style="width: 120px">
                <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
              </el-select>
              <el-button size="small" @click="handleUpdateStatus(row)">保存</el-button>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="KYC 状态" min-width="300">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-tag :type="kycTagType(row.kyc_status)">{{ row.kyc_status || "-" }}</el-tag>
              <el-select v-model="draftKYCMap[row.id]" style="width: 120px">
                <el-option v-for="item in kycStatusOptions" :key="item" :label="item" :value="item" />
              </el-select>
              <el-button size="small" @click="handleUpdateKYC(row)">保存</el-button>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="会员等级" min-width="260">
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

    <el-dialog v-model="batchResultVisible" :title="batchResultTitle" width="760px" destroy-on-close>
      <div class="batch-result-summary">
        <el-tag type="success">成功 {{ batchResultStats.success }}</el-tag>
        <el-tag type="danger">失败 {{ batchResultStats.failed }}</el-tag>
        <el-tag type="warning">跳过 {{ batchResultStats.skipped }}</el-tag>
        <el-text type="info">总计 {{ batchResultStats.total }} 条</el-text>
      </div>
      <div class="batch-result-toolbar">
        <el-select v-model="batchResultFilter" style="width: 180px">
          <el-option label="查看全部结果" value="all" />
          <el-option label="仅看失败" value="failed" />
          <el-option label="仅看跳过" value="skipped" />
        </el-select>
        <el-button
          :disabled="failedBatchRows.length <= 0"
          :loading="copyingFailedDetails"
          @click="copyFailedDetails"
        >
          复制失败明细
        </el-button>
      </div>
      <el-table :data="displayBatchResultRows" border stripe max-height="360" empty-text="当前筛选条件下暂无结果">
        <el-table-column prop="id" label="用户ID" min-width="140" />
        <el-table-column prop="action" label="动作" min-width="130" />
        <el-table-column prop="target" label="目标值" min-width="120" />
        <el-table-column label="结果" min-width="110">
          <template #default="{ row }">
            <el-tag :type="row.result === 'SUCCESS' ? 'success' : row.result === 'SKIPPED' ? 'warning' : 'danger'">
              {{ row.result }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="说明" min-width="260" />
      </el-table>
      <template #footer>
        <el-button
          type="warning"
          plain
          :disabled="failedBatchRows.length <= 0"
          :loading="retryingFailed"
          @click="retryFailedBatchRows"
        >
          重试失败任务
        </el-button>
        <el-button type="primary" @click="batchResultVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.inline-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.field-stack {
  display: flex;
  align-items: center;
  gap: 8px;
}

.batch-result-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.batch-result-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}
</style>
