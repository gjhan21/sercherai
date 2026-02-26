<script setup>
import { onMounted, reactive, ref } from "vue";
import {
  getAuthRiskConfig,
  listAuthLoginLogs,
  listAuthRiskConfigLogs,
  listAuthUnlockLogs,
  unlockAuthRiskState,
  updateAuthRiskConfig
} from "../api/admin";
import { getAccessToken } from "../lib/session";

const activeTab = ref("login-logs");

const errorMessage = ref("");
const message = ref("");
const refreshingAll = ref(false);

const loginLoading = ref(false);
const loginExporting = ref(false);
const loginPage = ref(1);
const loginPageSize = ref(20);
const loginTotal = ref(0);
const loginItems = ref([]);
const loginFilters = reactive({
  action: "",
  status: "",
  date_from: "",
  date_to: ""
});

const riskConfigLoading = ref(false);
const riskConfigSubmitting = ref(false);
const riskConfig = reactive({
  phone_fail_threshold: 5,
  ip_fail_threshold: 20,
  ip_phone_threshold: 5,
  lock_seconds: 900
});

const riskLogLoading = ref(false);
const riskLogPage = ref(1);
const riskLogPageSize = ref(20);
const riskLogTotal = ref(0);
const riskLogItems = ref([]);

const unlockSubmitting = ref(false);
const unlockForm = reactive({
  phone: "",
  ip: "",
  reason: ""
});
const unlockLogLoading = ref(false);
const unlockLogPage = ref(1);
const unlockLogPageSize = ref(20);
const unlockLogTotal = ref(0);
const unlockLogItems = ref([]);
const unlockLogFilters = reactive({
  phone: "",
  ip: "",
  operator_user_id: ""
});

const actionOptions = ["LOGIN", "REGISTER", "REFRESH", "LOGOUT", "LOGOUT_ALL", "MOCK_LOGIN"];
const authStatusOptions = ["SUCCESS", "FAILED"];

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function toSafeInt(value, fallback) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? Math.trunc(parsed) : fallback;
}

function clearMessages() {
  errorMessage.value = "";
  message.value = "";
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (["SUCCESS", "ACTIVE"].includes(normalized)) return "success";
  if (["FAILED", "DISABLED"].includes(normalized)) return "danger";
  if (["PENDING", "RUNNING"].includes(normalized)) return "warning";
  return "info";
}

async function fetchLoginLogs(options = {}) {
  const { keepMessage = false } = options;
  loginLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }

  try {
    const data = await listAuthLoginLogs({
      action: loginFilters.action,
      status: loginFilters.status,
      page: loginPage.value,
      page_size: loginPageSize.value
    });
    loginItems.value = data.items || [];
    loginTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载登录日志失败");
  } finally {
    loginLoading.value = false;
  }
}

async function exportLoginLogsCSV() {
  loginExporting.value = true;
  clearMessages();

  try {
    const params = new URLSearchParams();
    if (loginFilters.action) params.set("action", loginFilters.action);
    if (loginFilters.status) params.set("status", loginFilters.status);
    if (loginFilters.date_from.trim()) params.set("date_from", loginFilters.date_from.trim());
    if (loginFilters.date_to.trim()) params.set("date_to", loginFilters.date_to.trim());

    const baseURL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
    const query = params.toString();
    const requestURL = `${baseURL}/admin/auth/login-logs/export.csv${query ? `?${query}` : ""}`;

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
    const fileName = `auth_login_logs_${new Date().toISOString().slice(0, 10)}.csv`;
    const anchor = document.createElement("a");
    anchor.href = blobURL;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(blobURL);

    message.value = "已发起登录日志 CSV 下载";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "导出登录日志失败");
  } finally {
    loginExporting.value = false;
  }
}

function applyLoginFilters() {
  loginPage.value = 1;
  fetchLoginLogs();
}

function resetLoginFilters() {
  loginFilters.action = "";
  loginFilters.status = "";
  loginFilters.date_from = "";
  loginFilters.date_to = "";
  loginPage.value = 1;
  fetchLoginLogs();
}

function handleLoginPageChange(nextPage) {
  if (nextPage === loginPage.value) {
    return;
  }
  loginPage.value = nextPage;
  fetchLoginLogs();
}

async function fetchRiskConfig(options = {}) {
  const { keepMessage = false } = options;
  riskConfigLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }

  try {
    const data = await getAuthRiskConfig();
    Object.assign(riskConfig, {
      phone_fail_threshold: data.phone_fail_threshold ?? 5,
      ip_fail_threshold: data.ip_fail_threshold ?? 20,
      ip_phone_threshold: data.ip_phone_threshold ?? 5,
      lock_seconds: data.lock_seconds ?? 900
    });
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载登录风控参数失败");
  } finally {
    riskConfigLoading.value = false;
  }
}

async function submitRiskConfig() {
  const payload = {
    phone_fail_threshold: toSafeInt(riskConfig.phone_fail_threshold, 5),
    ip_fail_threshold: toSafeInt(riskConfig.ip_fail_threshold, 20),
    ip_phone_threshold: toSafeInt(riskConfig.ip_phone_threshold, 5),
    lock_seconds: toSafeInt(riskConfig.lock_seconds, 900)
  };

  if (
    payload.phone_fail_threshold <= 0 ||
    payload.ip_fail_threshold <= 0 ||
    payload.ip_phone_threshold <= 0 ||
    payload.lock_seconds <= 0
  ) {
    errorMessage.value = "所有风控参数必须为正整数";
    return;
  }

  riskConfigSubmitting.value = true;
  clearMessages();
  try {
    await updateAuthRiskConfig(payload);
    await Promise.all([fetchRiskConfig({ keepMessage: true }), fetchRiskLogs({ keepMessage: true })]);
    message.value = "登录风控参数已更新";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "更新登录风控参数失败");
  } finally {
    riskConfigSubmitting.value = false;
  }
}

async function fetchRiskLogs(options = {}) {
  const { keepMessage = false } = options;
  riskLogLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }

  try {
    const data = await listAuthRiskConfigLogs({
      page: riskLogPage.value,
      page_size: riskLogPageSize.value
    });
    riskLogItems.value = data.items || [];
    riskLogTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载风控配置日志失败");
  } finally {
    riskLogLoading.value = false;
  }
}

function handleRiskLogPageChange(nextPage) {
  if (nextPage === riskLogPage.value) {
    return;
  }
  riskLogPage.value = nextPage;
  fetchRiskLogs();
}

async function submitUnlock() {
  const payload = {
    phone: unlockForm.phone.trim(),
    ip: unlockForm.ip.trim(),
    reason: unlockForm.reason.trim()
  };

  if (!payload.phone && !payload.ip) {
    errorMessage.value = "手机号和IP至少填写一个";
    return;
  }

  unlockSubmitting.value = true;
  clearMessages();
  try {
    await unlockAuthRiskState(payload);
    Object.assign(unlockForm, {
      phone: "",
      ip: "",
      reason: ""
    });
    await fetchUnlockLogs({ keepMessage: true });
    message.value = "已提交风控解锁";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "提交解锁失败");
  } finally {
    unlockSubmitting.value = false;
  }
}

async function fetchUnlockLogs(options = {}) {
  const { keepMessage = false } = options;
  unlockLogLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }

  try {
    const data = await listAuthUnlockLogs({
      phone: unlockLogFilters.phone.trim(),
      ip: unlockLogFilters.ip.trim(),
      operator_user_id: unlockLogFilters.operator_user_id.trim(),
      page: unlockLogPage.value,
      page_size: unlockLogPageSize.value
    });
    unlockLogItems.value = data.items || [];
    unlockLogTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载解锁日志失败");
  } finally {
    unlockLogLoading.value = false;
  }
}

function applyUnlockLogFilters() {
  unlockLogPage.value = 1;
  fetchUnlockLogs();
}

function resetUnlockLogFilters() {
  unlockLogFilters.phone = "";
  unlockLogFilters.ip = "";
  unlockLogFilters.operator_user_id = "";
  unlockLogPage.value = 1;
  fetchUnlockLogs();
}

function handleUnlockLogPageChange(nextPage) {
  if (nextPage === unlockLogPage.value) {
    return;
  }
  unlockLogPage.value = nextPage;
  fetchUnlockLogs();
}

async function refreshCurrentTab() {
  if (activeTab.value === "login-logs") {
    await fetchLoginLogs();
    return;
  }
  if (activeTab.value === "risk-config") {
    await Promise.all([fetchRiskConfig(), fetchRiskLogs()]);
    return;
  }
  await fetchUnlockLogs();
}

async function refreshAll() {
  refreshingAll.value = true;
  clearMessages();
  try {
    await Promise.all([
      fetchLoginLogs({ keepMessage: true }),
      fetchRiskConfig({ keepMessage: true }),
      fetchRiskLogs({ keepMessage: true }),
      fetchUnlockLogs({ keepMessage: true })
    ]);
    message.value = "安全中心数据已刷新";
  } finally {
    refreshingAll.value = false;
  }
}

onMounted(refreshAll);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">安全中心</h1>
        <p class="muted">登录日志、风控参数配置、锁定状态解锁与审计</p>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <el-button :loading="refreshingAll" @click="refreshAll">刷新全部</el-button>
        <el-button type="primary" plain :loading="refreshingAll" @click="refreshCurrentTab">刷新当前页签</el-button>
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

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="登录日志" name="login-logs">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select v-model="loginFilters.action" clearable placeholder="动作" style="width: 160px">
              <el-option v-for="item in actionOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-select v-model="loginFilters.status" clearable placeholder="结果" style="width: 140px">
              <el-option v-for="item in authStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-input v-model="loginFilters.date_from" clearable placeholder="date_from (YYYY-MM-DD)" style="width: 180px" />
            <el-input v-model="loginFilters.date_to" clearable placeholder="date_to (YYYY-MM-DD)" style="width: 180px" />
            <el-button :loading="loginExporting" @click="exportLoginLogsCSV">导出筛选CSV</el-button>
            <el-button type="primary" plain @click="applyLoginFilters">查询</el-button>
            <el-button @click="resetLoginFilters">重置</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="loginItems" border stripe v-loading="loginLoading" empty-text="暂无登录日志">
            <el-table-column prop="id" label="日志ID" min-width="160" />
            <el-table-column prop="action" label="动作" min-width="100" />
            <el-table-column prop="status" label="结果" min-width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="user_id" label="用户ID" min-width="130" />
            <el-table-column prop="phone" label="手机号" min-width="130" />
            <el-table-column prop="ip" label="IP" min-width="130" />
            <el-table-column prop="reason" label="原因" min-width="180" />
            <el-table-column label="UA" min-width="200">
              <template #default="{ row }">
                <span class="ua-text">{{ row.user_agent || "-" }}</span>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="时间" min-width="180" />
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ loginPage }} 页，共 {{ loginTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="loginPage"
              :page-size="loginPageSize"
              :total="loginTotal"
              @current-change="handleLoginPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="风控参数" name="risk-config">
        <div class="card" style="margin-bottom: 12px" v-loading="riskConfigLoading">
          <div class="section-header">
            <h3 style="margin: 0">登录风控参数</h3>
          </div>
          <el-form label-width="180px">
            <div class="dialog-grid">
              <el-form-item label="手机号失败阈值" required>
                <el-input-number
                  v-model="riskConfig.phone_fail_threshold"
                  :min="1"
                  :step="1"
                  controls-position="right"
                  style="width: 100%"
                />
              </el-form-item>
              <el-form-item label="IP失败阈值" required>
                <el-input-number
                  v-model="riskConfig.ip_fail_threshold"
                  :min="1"
                  :step="1"
                  controls-position="right"
                  style="width: 100%"
                />
              </el-form-item>
              <el-form-item label="IP+手机号阈值" required>
                <el-input-number
                  v-model="riskConfig.ip_phone_threshold"
                  :min="1"
                  :step="1"
                  controls-position="right"
                  style="width: 100%"
                />
              </el-form-item>
              <el-form-item label="锁定时长(秒)" required>
                <el-input-number
                  v-model="riskConfig.lock_seconds"
                  :min="1"
                  :step="60"
                  controls-position="right"
                  style="width: 100%"
                />
              </el-form-item>
            </div>
          </el-form>
          <el-button type="primary" :loading="riskConfigSubmitting" @click="submitRiskConfig">保存风控参数</el-button>
        </div>

        <div class="card">
          <div class="section-header">
            <h3 style="margin: 0">风控参数变更日志</h3>
            <el-button :loading="riskLogLoading" @click="fetchRiskLogs">刷新日志</el-button>
          </div>
          <el-table :data="riskLogItems" border stripe v-loading="riskLogLoading" empty-text="暂无风控参数日志">
            <el-table-column prop="id" label="日志ID" min-width="150" />
            <el-table-column prop="operator_user_id" label="操作人" min-width="120" />
            <el-table-column label="旧参数" min-width="250">
              <template #default="{ row }">
                <div class="config-line">phone_fail: {{ row.old_config?.phone_fail_threshold ?? '-' }}</div>
                <div class="config-line">ip_fail: {{ row.old_config?.ip_fail_threshold ?? '-' }}</div>
                <div class="config-line">ip_phone: {{ row.old_config?.ip_phone_threshold ?? '-' }}</div>
                <div class="config-line">lock_seconds: {{ row.old_config?.lock_seconds ?? '-' }}</div>
              </template>
            </el-table-column>
            <el-table-column label="新参数" min-width="250">
              <template #default="{ row }">
                <div class="config-line">phone_fail: {{ row.new_config?.phone_fail_threshold ?? '-' }}</div>
                <div class="config-line">ip_fail: {{ row.new_config?.ip_fail_threshold ?? '-' }}</div>
                <div class="config-line">ip_phone: {{ row.new_config?.ip_phone_threshold ?? '-' }}</div>
                <div class="config-line">lock_seconds: {{ row.new_config?.lock_seconds ?? '-' }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="时间" min-width="180" />
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ riskLogPage }} 页，共 {{ riskLogTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="riskLogPage"
              :page-size="riskLogPageSize"
              :total="riskLogTotal"
              @current-change="handleRiskLogPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="解锁操作" name="unlock-logs">
        <div class="card" style="margin-bottom: 12px">
          <div class="section-header">
            <h3 style="margin: 0">提交解锁</h3>
          </div>
          <el-form label-width="130px">
            <div class="dialog-grid">
              <el-form-item label="手机号">
                <el-input v-model="unlockForm.phone" placeholder="可选" />
              </el-form-item>
              <el-form-item label="IP地址">
                <el-input v-model="unlockForm.ip" placeholder="可选" />
              </el-form-item>
            </div>
            <el-form-item label="解锁原因">
              <el-input
                v-model="unlockForm.reason"
                type="textarea"
                :rows="3"
                maxlength="200"
                show-word-limit
                placeholder="可选，记录解锁原因"
              />
            </el-form-item>
          </el-form>
          <el-button type="primary" :loading="unlockSubmitting" @click="submitUnlock">提交解锁</el-button>
        </div>

        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-input v-model="unlockLogFilters.phone" clearable placeholder="手机号" style="width: 180px" />
            <el-input v-model="unlockLogFilters.ip" clearable placeholder="IP" style="width: 180px" />
            <el-input
              v-model="unlockLogFilters.operator_user_id"
              clearable
              placeholder="操作人ID"
              style="width: 180px"
            />
            <el-button type="primary" plain @click="applyUnlockLogFilters">查询</el-button>
            <el-button @click="resetUnlockLogFilters">重置</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="unlockLogItems" border stripe v-loading="unlockLogLoading" empty-text="暂无解锁日志">
            <el-table-column prop="id" label="日志ID" min-width="150" />
            <el-table-column prop="operator_user_id" label="操作人" min-width="130" />
            <el-table-column prop="phone" label="手机号" min-width="130" />
            <el-table-column prop="ip" label="IP" min-width="130" />
            <el-table-column prop="reason" label="原因" min-width="220" />
            <el-table-column prop="created_at" label="时间" min-width="180" />
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ unlockLogPage }} 页，共 {{ unlockLogTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="unlockLogPage"
              :page-size="unlockLogPageSize"
              :total="unlockLogTotal"
              @current-change="handleUnlockLogPageChange"
            />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped>
.ua-text {
  display: -webkit-box;
  overflow: hidden;
  text-overflow: ellipsis;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.dialog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 0 12px;
}

.config-line {
  font-size: 12px;
  line-height: 1.5;
}
</style>
