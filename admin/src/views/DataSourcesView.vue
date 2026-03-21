<script setup>
import { onMounted, reactive, ref } from "vue";
import { ElMessageBox } from "element-plus";
import {
  batchCheckDataSources,
  checkDataSourceHealth,
  createDataSource,
  deleteDataSource,
  listDataSourceHealthLogs,
  listDataSources,
  listSystemConfigs,
  upsertSystemConfig,
  updateDataSource
} from "../api/admin";
import { hasPermission } from "../lib/session";

const loading = ref(false);
const submitting = ref(false);
const batchChecking = ref(false);
const settingDefaultSource = ref(false);
const settingDefaultSourceKey = ref("");
const message = ref("");
const errorMessage = ref("");

const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const items = ref([]);

const healthMap = ref({});

const logsLoading = ref(false);
const logSourceKey = ref("");
const logItems = ref([]);

const formVisible = ref(false);
const formMode = ref("create");
const defaultStockSourceKey = ref("TUSHARE");
const editingConfigSnapshot = ref({});

const sourceTypeOptions = ["MARKET", "NEWS", "STOCK", "FUTURES", "SYSTEM"];
const statusOptions = ["ACTIVE", "DISABLED"];
const STOCK_DEFAULT_SOURCE_CONFIG_KEY = "stock.quotes.default_source_key";
const STOCK_DEFAULT_SOURCE_FALLBACK = "TUSHARE";
const canEditDataSources = hasPermission("data_source.edit");

function defaultForm() {
  return {
    source_key: "",
    name: "",
    source_type: "MARKET",
    status: "ACTIVE",
    endpoint: "",
    token: "",
    fail_threshold: 3,
    retry_times: 0,
    retry_interval_ms: 200,
    health_timeout_ms: 3000,
    alert_receiver_id: "admin_001"
  };
}

const form = reactive(defaultForm());

function resetForm() {
  Object.assign(form, defaultForm());
  formMode.value = "create";
  editingConfigSnapshot.value = {};
}

function toSafeInt(value, fallback) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? Math.trunc(parsed) : fallback;
}

function normalizeSourceKey(value) {
  return String(value || "")
    .trim()
    .toUpperCase();
}

function resolveDataSourceProvider(item) {
  return String(item?.config?.provider || item?.config?.vendor || "")
    .trim()
    .toUpperCase();
}

function supportsDefaultStockSource(item) {
  const sourceType = String(item?.source_type || "")
    .trim()
    .toUpperCase();
  if (sourceType === "STOCK") {
    return true;
  }
  const provider = resolveDataSourceProvider(item);
  return ["TUSHARE", "AKSHARE", "TICKERMD", "MOCK", "MYSELF"].includes(provider);
}

function isDefaultStockSource(sourceKey) {
  return normalizeSourceKey(sourceKey) === normalizeSourceKey(defaultStockSourceKey.value);
}

function cloneConfigMap(value) {
  return JSON.parse(JSON.stringify(value || {}));
}

function ensureCanEditDataSources() {
  if (canEditDataSources) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法修改数据源配置或执行健康检查";
  return false;
}

async function loadDefaultStockSourceKey(options = {}) {
  const { silent = false } = options;
  try {
    const data = await listSystemConfigs({
      keyword: STOCK_DEFAULT_SOURCE_CONFIG_KEY,
      page: 1,
      page_size: 50
    });
    const rows = Array.isArray(data?.items) ? data.items : [];
    const matched = rows.find(
      (item) =>
        normalizeSourceKey(item?.config_key) === normalizeSourceKey(STOCK_DEFAULT_SOURCE_CONFIG_KEY)
    );
    const configValue = normalizeSourceKey(matched?.config_value);
    defaultStockSourceKey.value = configValue || STOCK_DEFAULT_SOURCE_FALLBACK;
  } catch (error) {
    defaultStockSourceKey.value = STOCK_DEFAULT_SOURCE_FALLBACK;
    if (!silent) {
      throw error;
    }
  }
}

function buildPayload() {
  const sourceKey = form.source_key.trim().toUpperCase();
  const endpointValue = form.endpoint.trim();
  const config = {
    ...(formMode.value === "edit" ? cloneConfigMap(editingConfigSnapshot.value) : {}),
    endpoint: sourceKey === "TUSHARE" && !endpointValue ? "https://api.tushare.pro" : endpointValue,
    fail_threshold: toSafeInt(form.fail_threshold, 3),
    retry_times: toSafeInt(form.retry_times, 0),
    retry_interval_ms: toSafeInt(form.retry_interval_ms, 200),
    health_timeout_ms: toSafeInt(form.health_timeout_ms, 3000)
  };
  delete config.token;
  delete config.api_token;
  delete config.tushare_token;
  if (form.alert_receiver_id.trim()) {
    config.alert_receiver_id = form.alert_receiver_id.trim();
  } else {
    delete config.alert_receiver_id;
  }
  if (form.token.trim()) {
    config.token = form.token.trim();
  }
  return {
    source_key: sourceKey,
    name: form.name.trim(),
    source_type: form.source_type.trim(),
    status: form.status,
    config
  };
}

async function fetchDataSources(options = {}) {
  const { preserveFeedback = false } = options;
  loading.value = true;
  if (!preserveFeedback) {
    errorMessage.value = "";
    message.value = "";
  }
  try {
    const [data] = await Promise.all([
      listDataSources({ page: page.value, page_size: pageSize.value }),
      loadDefaultStockSourceKey({ silent: true })
    ]);
    items.value = data.items || [];
    total.value = data.total || 0;
  } catch (error) {
    errorMessage.value = error.message || "加载数据源失败";
  } finally {
    loading.value = false;
  }
}

async function submitForm() {
  if (!ensureCanEditDataSources()) {
    return;
  }
  submitting.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const payload = buildPayload();
    if (!payload.name || !payload.source_type) {
      throw new Error("请完整填写必填字段");
    }
    if (formMode.value === "create") {
      if (!payload.source_key) {
        throw new Error("source_key 不能为空");
      }
      await createDataSource(payload);
      message.value = `数据源 ${payload.source_key} 创建成功`;
    } else {
      await updateDataSource(payload.source_key, {
        name: payload.name,
        source_type: payload.source_type,
        status: payload.status,
        config: payload.config
      });
      message.value = `数据源 ${payload.source_key} 更新成功`;
    }
    formVisible.value = false;
    resetForm();
    await fetchDataSources({ preserveFeedback: true });
  } catch (error) {
    errorMessage.value = error.message || "提交失败";
  } finally {
    submitting.value = false;
  }
}

function handleCreate() {
  if (!ensureCanEditDataSources()) {
    return;
  }
  resetForm();
  formVisible.value = true;
}

function handleEdit(item) {
  if (!ensureCanEditDataSources()) {
    return;
  }
  const cfg = item.config || {};
  editingConfigSnapshot.value = cloneConfigMap(cfg);
  Object.assign(form, {
    source_key: item.source_key || "",
    name: item.name || "",
    source_type: item.source_type || "MARKET",
    status: item.status || "ACTIVE",
    endpoint: cfg.endpoint || "",
    token: cfg.token || cfg.api_token || cfg.tushare_token || "",
    fail_threshold: toSafeInt(cfg.fail_threshold, 3),
    retry_times: toSafeInt(cfg.retry_times, 0),
    retry_interval_ms: toSafeInt(cfg.retry_interval_ms, 200),
    health_timeout_ms: toSafeInt(cfg.health_timeout_ms, 3000),
    alert_receiver_id: cfg.alert_receiver_id || "admin_001"
  });
  formMode.value = "edit";
  formVisible.value = true;
}

async function handleDelete(sourceKey) {
  if (!ensureCanEditDataSources()) {
    return;
  }
  try {
    await ElMessageBox.confirm(`确认删除数据源 ${sourceKey}？`, "删除确认", {
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
    await deleteDataSource(sourceKey);
    message.value = `数据源 ${sourceKey} 已删除`;
    await fetchDataSources({ preserveFeedback: true });
  } catch (error) {
    errorMessage.value = error.message || "删除失败";
  }
}

async function handleSetDefaultStockSource(row) {
  if (!ensureCanEditDataSources()) {
    return;
  }
  if (!supportsDefaultStockSource(row)) {
    errorMessage.value = "仅支持设置可用于股票行情同步的数据源为默认行情源";
    return;
  }
  const sourceKey = normalizeSourceKey(row.source_key);
  if (!sourceKey) {
    errorMessage.value = "source_key 不能为空";
    return;
  }
  settingDefaultSource.value = true;
  settingDefaultSourceKey.value = sourceKey;
  errorMessage.value = "";
  message.value = "";
  try {
    await upsertSystemConfig({
      config_key: STOCK_DEFAULT_SOURCE_CONFIG_KEY,
      config_value: sourceKey,
      description: "股票行情默认数据源"
    });
    defaultStockSourceKey.value = sourceKey;
    message.value = `默认行情源已设置为 ${sourceKey}`;
  } catch (error) {
    errorMessage.value = error.message || "设置默认行情源失败";
  } finally {
    settingDefaultSource.value = false;
    settingDefaultSourceKey.value = "";
  }
}

async function handleCheckOne(sourceKey) {
  if (!ensureCanEditDataSources()) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await checkDataSourceHealth(sourceKey);
    healthMap.value[sourceKey] = result;
    message.value = `数据源 ${sourceKey} 健康检查完成`;
  } catch (error) {
    errorMessage.value = error.message || "健康检查失败";
  }
}

async function handleBatchCheckAll() {
  if (!ensureCanEditDataSources()) {
    return;
  }
  batchChecking.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const result = await batchCheckDataSources([]);
    const rows = result.items || [];
    const merged = { ...healthMap.value };
    rows.forEach((row) => {
      merged[row.source_key] = row;
    });
    healthMap.value = merged;
    message.value = `批量健康检查完成，共 ${rows.length} 个数据源`;
  } catch (error) {
    errorMessage.value = error.message || "批量健康检查失败";
  } finally {
    batchChecking.value = false;
  }
}

async function showLogs(sourceKey) {
  logsLoading.value = true;
  logSourceKey.value = sourceKey;
  errorMessage.value = "";
  try {
    const result = await listDataSourceHealthLogs(sourceKey, {
      page: 1,
      page_size: 20
    });
    logItems.value = result.items || [];
  } catch (error) {
    logItems.value = [];
    errorMessage.value = error.message || "加载健康日志失败";
  } finally {
    logsLoading.value = false;
  }
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "ACTIVE" || normalized === "HEALTHY") return "success";
  if (normalized === "DISABLED" || normalized === "UNHEALTHY") return "danger";
  if (normalized === "PENDING") return "warning";
  return "info";
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  fetchDataSources();
}

onMounted(fetchDataSources);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">数据源管理</h1>
        <p class="muted">管理数据源配置、健康检查与日志</p>
      </div>
      <div class="toolbar">
        <el-tag type="warning" effect="plain">
          默认行情源：{{ defaultStockSourceKey || "-" }}
        </el-tag>
        <el-button :loading="loading" @click="fetchDataSources">刷新列表</el-button>
        <el-button
          v-if="canEditDataSources"
          type="primary"
          plain
          :loading="batchChecking"
          @click="handleBatchCheckAll"
        >
          全部健康检查
        </el-button>
        <el-button v-if="canEditDataSources" type="primary" @click="handleCreate">新增数据源</el-button>
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

    <div class="card">
      <el-table :data="items" border stripe v-loading="loading" empty-text="暂无数据源">
        <el-table-column label="source_key" min-width="190">
          <template #default="{ row }">
            <div class="source-key-cell">
              <span>{{ row.source_key }}</span>
              <el-tag v-if="isDefaultStockSource(row.source_key)" size="small" type="warning" effect="plain">
                默认行情源
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="130" />
        <el-table-column prop="source_type" label="类型" min-width="110" />
        <el-table-column label="状态" min-width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="配置摘要" min-width="260">
          <template #default="{ row }">
            <div class="config-line">endpoint: {{ row.config?.endpoint || "-" }}</div>
            <div class="config-line config-line--muted">
              token: {{ row.config?.token ? "已配置" : "未配置" }}
            </div>
            <div class="config-line config-line--muted">
              阈值: {{ row.config?.fail_threshold ?? 3 }} / 重试: {{ row.config?.retry_times ?? 0 }}
            </div>
          </template>
        </el-table-column>
        <el-table-column label="最近健康状态" min-width="260">
          <template #default="{ row }">
            <template v-if="healthMap[row.source_key]">
              <el-tag :type="statusTagType(healthMap[row.source_key].status)">
                {{ healthMap[row.source_key].status }}
              </el-tag>
              <div class="config-line config-line--muted">
                {{ healthMap[row.source_key].message || "-" }} · {{ healthMap[row.source_key].latency_ms || 0 }}ms
              </div>
              <div class="config-line config-line--muted">
                尝试 {{ healthMap[row.source_key].attempts || 1 }}/{{ healthMap[row.source_key].max_attempts || 1 }} · 连续失败 {{ healthMap[row.source_key].consecutive_failures || 0 }}
              </div>
            </template>
            <el-text v-else type="info">未检查</el-text>
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" min-width="180" />
        <el-table-column label="操作" align="right" min-width="420">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-button
                v-if="canEditDataSources && supportsDefaultStockSource(row)"
                size="small"
                type="success"
                plain
                :loading="settingDefaultSource && settingDefaultSourceKey === normalizeSourceKey(row.source_key)"
                :disabled="isDefaultStockSource(row.source_key)"
                @click="handleSetDefaultStockSource(row)"
              >
                设为默认行情源
              </el-button>
              <el-button v-if="canEditDataSources" size="small" @click="handleCheckOne(row.source_key)">健康检查</el-button>
              <el-button size="small" @click="showLogs(row.source_key)">健康日志</el-button>
              <el-button v-if="canEditDataSources" size="small" @click="handleEdit(row)">编辑</el-button>
              <el-button
                v-if="canEditDataSources"
                size="small"
                type="danger"
                plain
                @click="handleDelete(row.source_key)"
              >
                删除
              </el-button>
            </div>
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

    <div class="card" style="margin-top: 12px">
      <div class="log-header">
        <h3 style="margin: 0">健康日志 {{ logSourceKey ? `(${logSourceKey})` : "" }}</h3>
        <el-text type="info">{{ logsLoading ? "加载中..." : "最近20条" }}</el-text>
      </div>
      <el-table :data="logItems" border stripe v-loading="logsLoading" empty-text="暂无日志，先执行健康检查">
        <el-table-column label="状态" min-width="120">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="reachable" min-width="110">
          <template #default="{ row }">
            {{ row.reachable ? "是" : "否" }}
          </template>
        </el-table-column>
        <el-table-column prop="http_status" label="HTTP" min-width="90">
          <template #default="{ row }">
            {{ row.http_status || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="latency_ms" label="latency(ms)" min-width="110" />
        <el-table-column prop="message" label="message" min-width="220">
          <template #default="{ row }">
            {{ row.message || "-" }}
          </template>
        </el-table-column>
        <el-table-column prop="checked_at" label="checked_at" min-width="180" />
      </el-table>
    </div>

    <el-dialog
      v-model="formVisible"
      :title="formMode === 'create' ? '新增数据源' : `编辑数据源：${form.source_key}`"
      width="860px"
      destroy-on-close
    >
      <el-form label-width="120px">
        <div class="dialog-grid">
          <el-form-item label="source_key" required>
            <el-input v-model="form.source_key" :disabled="formMode === 'edit'" placeholder="wind" />
          </el-form-item>
          <el-form-item label="名称" required>
            <el-input v-model="form.name" placeholder="Wind 数据源" />
          </el-form-item>
          <el-form-item label="类型" required>
            <el-select v-model="form.source_type">
              <el-option v-for="item in sourceTypeOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="form.status">
              <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="endpoint">
            <el-input v-model="form.endpoint" placeholder="Tushare: https://api.tushare.pro" />
          </el-form-item>
          <el-form-item label="token">
            <el-input
              v-model="form.token"
              type="password"
              show-password
              placeholder="Tushare Token（可留空使用后端环境变量）"
            />
          </el-form-item>
          <el-form-item label="fail_threshold">
            <el-input-number v-model="form.fail_threshold" :min="1" :step="1" controls-position="right" />
          </el-form-item>
          <el-form-item label="retry_times">
            <el-input-number v-model="form.retry_times" :min="0" :max="5" :step="1" controls-position="right" />
          </el-form-item>
          <el-form-item label="retry_interval_ms">
            <el-input-number v-model="form.retry_interval_ms" :min="0" :step="100" controls-position="right" />
          </el-form-item>
          <el-form-item label="health_timeout_ms">
            <el-input-number v-model="form.health_timeout_ms" :min="500" :step="100" controls-position="right" />
          </el-form-item>
          <el-form-item label="alert_receiver_id">
            <el-input v-model="form.alert_receiver_id" placeholder="admin_001" />
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button v-if="canEditDataSources" type="primary" :loading="submitting" @click="submitForm">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.config-line {
  line-height: 1.5;
  word-break: break-all;
}

.source-key-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.config-line--muted {
  color: #6b7280;
}

.inline-actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
}

.log-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.dialog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
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
