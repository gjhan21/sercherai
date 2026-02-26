<script setup>
import { onMounted, reactive, ref } from "vue";
import {
  batchCheckDataSources,
  checkDataSourceHealth,
  createDataSource,
  deleteDataSource,
  listDataSourceHealthLogs,
  listDataSources,
  updateDataSource
} from "../api/admin";

const loading = ref(false);
const submitting = ref(false);
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

function defaultForm() {
  return {
    source_key: "",
    name: "",
    source_type: "MARKET",
    status: "ACTIVE",
    endpoint: "",
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
}

function toSafeInt(value, fallback) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? Math.trunc(parsed) : fallback;
}

function buildPayload() {
  const config = {
    endpoint: form.endpoint.trim(),
    fail_threshold: toSafeInt(form.fail_threshold, 3),
    retry_times: toSafeInt(form.retry_times, 0),
    retry_interval_ms: toSafeInt(form.retry_interval_ms, 200),
    health_timeout_ms: toSafeInt(form.health_timeout_ms, 3000)
  };
  if (form.alert_receiver_id.trim()) {
    config.alert_receiver_id = form.alert_receiver_id.trim();
  }
  return {
    source_key: form.source_key.trim(),
    name: form.name.trim(),
    source_type: form.source_type.trim(),
    status: form.status,
    config
  };
}

async function fetchDataSources() {
  loading.value = true;
  errorMessage.value = "";
  message.value = "";
  try {
    const data = await listDataSources({ page: page.value, page_size: pageSize.value });
    items.value = data.items || [];
    total.value = data.total || 0;
  } catch (error) {
    errorMessage.value = error.message || "加载数据源失败";
  } finally {
    loading.value = false;
  }
}

async function submitForm() {
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
    await fetchDataSources();
  } catch (error) {
    errorMessage.value = error.message || "提交失败";
  } finally {
    submitting.value = false;
  }
}

function handleCreate() {
  resetForm();
  formVisible.value = true;
}

function handleEdit(item) {
  const cfg = item.config || {};
  Object.assign(form, {
    source_key: item.source_key || "",
    name: item.name || "",
    source_type: item.source_type || "MARKET",
    status: item.status || "ACTIVE",
    endpoint: cfg.endpoint || "",
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
  if (!window.confirm(`确认删除数据源 ${sourceKey}？`)) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await deleteDataSource(sourceKey);
    message.value = `数据源 ${sourceKey} 已删除`;
    await fetchDataSources();
  } catch (error) {
    errorMessage.value = error.message || "删除失败";
  }
}

async function handleCheckOne(sourceKey) {
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

function statusClass(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "ACTIVE" || normalized === "HEALTHY") return "status-active";
  if (normalized === "DISABLED" || normalized === "UNHEALTHY") return "status-disabled";
  return "status-unknown";
}

function nextPage() {
  if (page.value * pageSize.value >= total.value) {
    return;
  }
  page.value += 1;
  fetchDataSources();
}

function prevPage() {
  if (page.value <= 1) {
    return;
  }
  page.value -= 1;
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
        <button class="btn" :disabled="loading" @click="fetchDataSources">刷新列表</button>
        <button class="btn" :disabled="loading" @click="handleBatchCheckAll">全部健康检查</button>
        <button class="btn btn-primary" @click="handleCreate">新增数据源</button>
      </div>
    </div>

    <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>
    <div v-if="message" class="success-message">{{ message }}</div>

    <div v-if="formVisible" class="card" style="margin-bottom: 12px">
      <h3 style="margin-top: 0">
        {{ formMode === "create" ? "新增数据源" : `编辑数据源：${form.source_key}` }}
      </h3>
      <div class="form-grid">
        <div class="form-item">
          <label>source_key</label>
          <input v-model="form.source_key" class="input" :disabled="formMode === 'edit'" placeholder="wind" />
        </div>
        <div class="form-item">
          <label>名称</label>
          <input v-model="form.name" class="input" placeholder="Wind 数据源" />
        </div>
        <div class="form-item">
          <label>类型</label>
          <select v-model="form.source_type" class="select">
            <option value="MARKET">MARKET</option>
            <option value="NEWS">NEWS</option>
            <option value="STOCK">STOCK</option>
            <option value="FUTURES">FUTURES</option>
            <option value="SYSTEM">SYSTEM</option>
          </select>
        </div>
        <div class="form-item">
          <label>状态</label>
          <select v-model="form.status" class="select">
            <option value="ACTIVE">ACTIVE</option>
            <option value="DISABLED">DISABLED</option>
          </select>
        </div>
        <div class="form-item">
          <label>endpoint</label>
          <input v-model="form.endpoint" class="input" placeholder="http://127.0.0.1:8080/healthz" />
        </div>
        <div class="form-item">
          <label>fail_threshold</label>
          <input v-model.number="form.fail_threshold" class="input" type="number" min="1" />
        </div>
        <div class="form-item">
          <label>retry_times</label>
          <input v-model.number="form.retry_times" class="input" type="number" min="0" max="5" />
        </div>
        <div class="form-item">
          <label>retry_interval_ms</label>
          <input v-model.number="form.retry_interval_ms" class="input" type="number" min="0" />
        </div>
        <div class="form-item">
          <label>health_timeout_ms</label>
          <input v-model.number="form.health_timeout_ms" class="input" type="number" min="500" />
        </div>
        <div class="form-item">
          <label>alert_receiver_id</label>
          <input v-model="form.alert_receiver_id" class="input" placeholder="admin_001" />
        </div>
      </div>
      <div class="form-actions">
        <button class="btn btn-primary" :disabled="submitting" @click="submitForm">
          {{ submitting ? "提交中..." : "提交" }}
        </button>
        <button class="btn" :disabled="submitting" @click="formVisible = false">取消</button>
      </div>
    </div>

    <div class="card">
      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>source_key</th>
              <th>名称</th>
              <th>类型</th>
              <th>状态</th>
              <th>配置摘要</th>
              <th>最近健康状态</th>
              <th>更新时间</th>
              <th class="text-right">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in items" :key="item.source_key">
              <td>{{ item.source_key }}</td>
              <td>{{ item.name }}</td>
              <td>{{ item.source_type }}</td>
              <td>
                <span class="status-tag" :class="statusClass(item.status)">
                  {{ item.status }}
                </span>
              </td>
              <td>
                <div>endpoint: {{ item.config?.endpoint || "-" }}</div>
                <div class="muted">
                  阈值: {{ item.config?.fail_threshold ?? 3 }} / 重试: {{ item.config?.retry_times ?? 0 }}
                </div>
              </td>
              <td>
                <template v-if="healthMap[item.source_key]">
                  <span class="status-tag" :class="statusClass(healthMap[item.source_key].status)">
                    {{ healthMap[item.source_key].status }}
                  </span>
                  <div class="muted">
                    {{ healthMap[item.source_key].message || "-" }} ·
                    {{ healthMap[item.source_key].latency_ms || 0 }}ms
                  </div>
                  <div class="muted">
                    尝试 {{ healthMap[item.source_key].attempts || 1 }}/{{
                      healthMap[item.source_key].max_attempts || 1
                    }}
                    · 连续失败 {{ healthMap[item.source_key].consecutive_failures || 0 }}
                  </div>
                </template>
                <span v-else class="muted">未检查</span>
              </td>
              <td>{{ item.updated_at || "-" }}</td>
              <td class="text-right">
                <div class="toolbar" style="justify-content: flex-end">
                  <button class="btn" @click="handleCheckOne(item.source_key)">健康检查</button>
                  <button class="btn" @click="showLogs(item.source_key)">健康日志</button>
                  <button class="btn" @click="handleEdit(item)">编辑</button>
                  <button class="btn btn-danger" @click="handleDelete(item.source_key)">删除</button>
                </div>
              </td>
            </tr>
            <tr v-if="!loading && items.length === 0">
              <td colspan="8" class="muted">暂无数据源</td>
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

    <div class="card" style="margin-top: 12px">
      <div class="page-header" style="margin-bottom: 10px">
        <h3 style="margin: 0">健康日志 {{ logSourceKey ? `(${logSourceKey})` : "" }}</h3>
        <span v-if="logsLoading" class="muted">加载中...</span>
      </div>

      <div class="table-wrap">
        <table class="table">
          <thead>
            <tr>
              <th>状态</th>
              <th>reachable</th>
              <th>HTTP</th>
              <th>latency(ms)</th>
              <th>message</th>
              <th>checked_at</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in logItems" :key="row.id">
              <td>
                <span class="status-tag" :class="statusClass(row.status)">{{ row.status }}</span>
              </td>
              <td>{{ row.reachable }}</td>
              <td>{{ row.http_status || "-" }}</td>
              <td>{{ row.latency_ms }}</td>
              <td>{{ row.message || "-" }}</td>
              <td>{{ row.checked_at }}</td>
            </tr>
            <tr v-if="!logsLoading && logItems.length === 0">
              <td colspan="6" class="muted">暂无日志，先执行健康检查</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
