<script setup>
import { onMounted, reactive, ref } from "vue";
import {
  adjustUserQuota,
  createMembershipProduct,
  createVIPQuotaConfig,
  listMembershipOrders,
  listMembershipProducts,
  listUserQuotas,
  listVIPQuotaConfigs,
  updateMembershipOrderStatus,
  updateMembershipProductStatus,
  updateVIPQuotaConfig
} from "../api/admin";
import { getAccessToken } from "../lib/session";

const activeTab = ref("products");

const errorMessage = ref("");
const message = ref("");

const refreshingAll = ref(false);

const productLoading = ref(false);
const productSubmitting = ref(false);
const productPage = ref(1);
const productPageSize = ref(20);
const productTotal = ref(0);
const products = ref([]);
const productFilters = reactive({
  status: ""
});
const draftProductStatusMap = ref({});
const productFormVisible = ref(false);
const productForm = reactive({
  name: "",
  price: 99,
  status: "ACTIVE",
  member_level: "VIP1",
  duration_days: 30
});

const orderLoading = ref(false);
const orderExporting = ref(false);
const orderPage = ref(1);
const orderPageSize = ref(20);
const orderTotal = ref(0);
const orders = ref([]);
const orderFilters = reactive({
  status: "",
  user_id: ""
});
const draftOrderStatusMap = ref({});

const quotaLoading = ref(false);
const quotaSubmitting = ref(false);
const quotaPage = ref(1);
const quotaPageSize = ref(20);
const quotaTotal = ref(0);
const quotaItems = ref([]);
const quotaFilters = reactive({
  member_level: "",
  status: ""
});
const quotaDialogVisible = ref(false);
const quotaFormMode = ref("create");
const quotaForm = reactive({
  id: "",
  member_level: "VIP1",
  doc_read_limit: 500,
  news_subscribe_limit: 200,
  reset_cycle: "MONTHLY",
  status: "ACTIVE",
  effective_at: ""
});

const usageLoading = ref(false);
const usageAdjusting = ref(false);
const usagePage = ref(1);
const usagePageSize = ref(20);
const usageTotal = ref(0);
const usageItems = ref([]);
const usageFilters = reactive({
  user_id: "",
  period_key: ""
});
const usageAdjustDialogVisible = ref(false);
const usageAdjustForm = reactive({
  user_id: "",
  period_key: "",
  doc_read_delta: 0,
  news_subscribe_delta: 0,
  reason: ""
});

const productStatusOptions = ["ACTIVE", "DISABLED"];
const orderStatusOptions = ["PENDING", "PAID", "CANCELED", "REFUNDED", "FAILED"];
const productMemberLevelOptions = ["VIP1", "VIP2"];
const memberLevelOptions = ["VIP1", "VIP2", "VIP3", "VIP4", "FREE"];
const resetCycleOptions = ["MONTHLY", "WEEKLY", "DAILY"];

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function toSafeInt(value, fallback = 0) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? Math.trunc(parsed) : fallback;
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (["ACTIVE", "PAID", "APPROVED"].includes(normalized)) return "success";
  if (["PENDING", "REVIEWING"].includes(normalized)) return "warning";
  if (["FAILED", "CANCELED", "DISABLED", "REJECTED", "REFUNDED"].includes(normalized)) return "danger";
  return "info";
}

function clearMessages() {
  errorMessage.value = "";
  message.value = "";
}

function resetProductForm() {
  Object.assign(productForm, {
    name: "",
    price: 99,
    status: "ACTIVE",
    member_level: "VIP1",
    duration_days: 30
  });
}

function syncProductDrafts() {
  const statusMap = {};
  products.value.forEach((item) => {
    statusMap[item.id] = item.status || "ACTIVE";
  });
  draftProductStatusMap.value = statusMap;
}

async function fetchProducts(options = {}) {
  const { keepMessage = false } = options;
  productLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listMembershipProducts({
      status: productFilters.status,
      page: productPage.value,
      page_size: productPageSize.value
    });
    products.value = data.items || [];
    productTotal.value = data.total || 0;
    syncProductDrafts();
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载会员产品失败");
  } finally {
    productLoading.value = false;
  }
}

async function submitProduct() {
  const payload = {
    name: productForm.name.trim(),
    price: Number(productForm.price),
    status: productForm.status,
    member_level: productForm.member_level,
    duration_days: toSafeInt(productForm.duration_days, 30)
  };
  if (!payload.name || !Number.isFinite(payload.price) || payload.price <= 0) {
    errorMessage.value = "请正确填写产品名称和价格";
    return;
  }

  productSubmitting.value = true;
  clearMessages();
  try {
    await createMembershipProduct(payload);
    productFormVisible.value = false;
    resetProductForm();
    await fetchProducts({ keepMessage: true });
    message.value = `会员产品 ${payload.name} 已创建`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "创建会员产品失败");
  } finally {
    productSubmitting.value = false;
  }
}

async function updateProductStatus(item) {
  const target = (draftProductStatusMap.value[item.id] || "").trim();
  if (!target || target === item.status) {
    return;
  }
  clearMessages();
  try {
    await updateMembershipProductStatus(item.id, target);
    await fetchProducts({ keepMessage: true });
    message.value = `产品 ${item.id} 状态已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "更新产品状态失败");
  }
}

function applyProductFilters() {
  productPage.value = 1;
  fetchProducts();
}

function resetProductFilters() {
  productFilters.status = "";
  productPage.value = 1;
  fetchProducts();
}

function handleProductPageChange(nextPage) {
  if (nextPage === productPage.value) {
    return;
  }
  productPage.value = nextPage;
  fetchProducts();
}

function syncOrderDrafts() {
  const map = {};
  orders.value.forEach((item) => {
    map[item.id] = item.status || "PENDING";
  });
  draftOrderStatusMap.value = map;
}

async function fetchOrders(options = {}) {
  const { keepMessage = false } = options;
  orderLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listMembershipOrders({
      status: orderFilters.status,
      user_id: orderFilters.user_id.trim(),
      page: orderPage.value,
      page_size: orderPageSize.value
    });
    orders.value = data.items || [];
    orderTotal.value = data.total || 0;
    syncOrderDrafts();
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载会员订单失败");
  } finally {
    orderLoading.value = false;
  }
}

async function updateOrderStatus(item) {
  const target = (draftOrderStatusMap.value[item.id] || "").trim();
  if (!target || target === item.status) {
    return;
  }
  clearMessages();
  try {
    await updateMembershipOrderStatus(item.id, target);
    await fetchOrders({ keepMessage: true });
    message.value = `订单 ${item.id} 状态已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "更新订单状态失败");
  }
}

async function exportOrdersFilteredCSV() {
  orderExporting.value = true;
  clearMessages();

  try {
    const params = new URLSearchParams();
    if (orderFilters.status) params.set("status", orderFilters.status);
    if (orderFilters.user_id.trim()) params.set("user_id", orderFilters.user_id.trim());

    const baseURL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
    const query = params.toString();
    const requestURL = `${baseURL}/admin/membership/orders/export.csv${query ? `?${query}` : ""}`;

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
    const fileName = `membership_orders_filtered_${new Date().toISOString().slice(0, 10)}.csv`;
    const anchor = document.createElement("a");
    anchor.href = blobURL;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(blobURL);

    message.value = "已发起会员订单筛选CSV下载";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "导出会员订单失败");
  } finally {
    orderExporting.value = false;
  }
}

function applyOrderFilters() {
  orderPage.value = 1;
  fetchOrders();
}

function resetOrderFilters() {
  orderFilters.status = "";
  orderFilters.user_id = "";
  orderPage.value = 1;
  fetchOrders();
}

function handleOrderPageChange(nextPage) {
  if (nextPage === orderPage.value) {
    return;
  }
  orderPage.value = nextPage;
  fetchOrders();
}

function resetQuotaForm() {
  Object.assign(quotaForm, {
    id: "",
    member_level: "VIP1",
    doc_read_limit: 500,
    news_subscribe_limit: 200,
    reset_cycle: "MONTHLY",
    status: "ACTIVE",
    effective_at: ""
  });
  quotaFormMode.value = "create";
}

function openCreateQuotaDialog() {
  resetQuotaForm();
  quotaDialogVisible.value = true;
}

function openEditQuotaDialog(item) {
  Object.assign(quotaForm, {
    id: item.id || "",
    member_level: item.member_level || "VIP1",
    doc_read_limit: item.doc_read_limit ?? 0,
    news_subscribe_limit: item.news_subscribe_limit ?? 0,
    reset_cycle: item.reset_cycle || "MONTHLY",
    status: item.status || "ACTIVE",
    effective_at: item.effective_at || ""
  });
  quotaFormMode.value = "edit";
  quotaDialogVisible.value = true;
}

async function fetchQuotaConfigs(options = {}) {
  const { keepMessage = false } = options;
  quotaLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listVIPQuotaConfigs({
      member_level: quotaFilters.member_level.trim(),
      status: quotaFilters.status,
      page: quotaPage.value,
      page_size: quotaPageSize.value
    });
    quotaItems.value = data.items || [];
    quotaTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载VIP配额配置失败");
  } finally {
    quotaLoading.value = false;
  }
}

async function submitQuotaForm() {
  const basePayload = {
    doc_read_limit: toSafeInt(quotaForm.doc_read_limit, 0),
    news_subscribe_limit: toSafeInt(quotaForm.news_subscribe_limit, 0),
    reset_cycle: quotaForm.reset_cycle,
    status: quotaForm.status,
    effective_at: quotaForm.effective_at.trim()
  };

  if (!basePayload.effective_at) {
    errorMessage.value = "effective_at 不能为空，格式示例：2026-12-31T00:00:00Z";
    return;
  }

  quotaSubmitting.value = true;
  clearMessages();
  try {
    if (quotaFormMode.value === "create") {
      await createVIPQuotaConfig({
        member_level: quotaForm.member_level,
        ...basePayload
      });
      message.value = `会员等级 ${quotaForm.member_level} 配额配置已创建`;
    } else {
      await updateVIPQuotaConfig(quotaForm.id, basePayload);
      message.value = `配额配置 ${quotaForm.id} 已更新`;
    }

    quotaDialogVisible.value = false;
    resetQuotaForm();
    await fetchQuotaConfigs({ keepMessage: true });
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "提交配额配置失败");
  } finally {
    quotaSubmitting.value = false;
  }
}

function applyQuotaFilters() {
  quotaPage.value = 1;
  fetchQuotaConfigs();
}

function resetQuotaFilters() {
  quotaFilters.member_level = "";
  quotaFilters.status = "";
  quotaPage.value = 1;
  fetchQuotaConfigs();
}

function handleQuotaPageChange(nextPage) {
  if (nextPage === quotaPage.value) {
    return;
  }
  quotaPage.value = nextPage;
  fetchQuotaConfigs();
}

async function fetchUserQuotaUsages(options = {}) {
  const { keepMessage = false } = options;
  usageLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listUserQuotas({
      user_id: usageFilters.user_id.trim(),
      period_key: usageFilters.period_key.trim(),
      page: usagePage.value,
      page_size: usagePageSize.value
    });
    usageItems.value = data.items || [];
    usageTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载用户配额失败");
  } finally {
    usageLoading.value = false;
  }
}

function openAdjustUsageDialog(item) {
  Object.assign(usageAdjustForm, {
    user_id: item.user_id,
    period_key: item.period_key,
    doc_read_delta: 0,
    news_subscribe_delta: 0,
    reason: ""
  });
  usageAdjustDialogVisible.value = true;
}

async function submitAdjustUsage() {
  const payload = {
    period_key: usageAdjustForm.period_key,
    doc_read_delta: toSafeInt(usageAdjustForm.doc_read_delta, 0),
    news_subscribe_delta: toSafeInt(usageAdjustForm.news_subscribe_delta, 0),
    reason: usageAdjustForm.reason.trim()
  };

  if (!usageAdjustForm.user_id || !payload.period_key) {
    errorMessage.value = "用户ID与周期不能为空";
    return;
  }

  usageAdjusting.value = true;
  clearMessages();
  try {
    await adjustUserQuota(usageAdjustForm.user_id, payload);
    usageAdjustDialogVisible.value = false;
    await fetchUserQuotaUsages({ keepMessage: true });
    message.value = `用户 ${usageAdjustForm.user_id} 配额已调整`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "调整用户配额失败");
  } finally {
    usageAdjusting.value = false;
  }
}

function applyUsageFilters() {
  usagePage.value = 1;
  fetchUserQuotaUsages();
}

function resetUsageFilters() {
  usageFilters.user_id = "";
  usageFilters.period_key = "";
  usagePage.value = 1;
  fetchUserQuotaUsages();
}

function handleUsagePageChange(nextPage) {
  if (nextPage === usagePage.value) {
    return;
  }
  usagePage.value = nextPage;
  fetchUserQuotaUsages();
}

async function refreshCurrentTab() {
  if (activeTab.value === "products") {
    await fetchProducts();
    return;
  }
  if (activeTab.value === "orders") {
    await fetchOrders();
    return;
  }
  if (activeTab.value === "quotas") {
    await fetchQuotaConfigs();
    return;
  }
  await fetchUserQuotaUsages();
}

async function refreshAll() {
  refreshingAll.value = true;
  clearMessages();
  try {
    await Promise.all([
      fetchProducts({ keepMessage: true }),
      fetchOrders({ keepMessage: true }),
      fetchQuotaConfigs({ keepMessage: true }),
      fetchUserQuotaUsages({ keepMessage: true })
    ]);
    message.value = "会员中心数据已刷新";
  } finally {
    refreshingAll.value = false;
  }
}

onMounted(() => {
  refreshAll();
});
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">会员中心</h1>
        <p class="muted">会员产品、订单、VIP配额与用户配额调整</p>
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
      <el-tab-pane label="会员产品" name="products">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select v-model="productFilters.status" clearable placeholder="全部状态" style="width: 140px">
              <el-option v-for="item in productStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-button type="primary" plain @click="applyProductFilters">查询</el-button>
            <el-button @click="resetProductFilters">重置</el-button>
            <el-button type="primary" @click="productFormVisible = true">新增产品</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="products" border stripe v-loading="productLoading" empty-text="暂无会员产品">
            <el-table-column prop="id" label="产品ID" min-width="140" />
            <el-table-column prop="name" label="名称" min-width="160" />
            <el-table-column prop="price" label="价格" min-width="100" />
            <el-table-column prop="member_level" label="会员等级" min-width="110" />
            <el-table-column prop="duration_days" label="时长(天)" min-width="100" />
            <el-table-column label="状态" min-width="250">
              <template #default="{ row }">
                <div class="inline-actions">
                  <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
                  <el-select v-model="draftProductStatusMap[row.id]" style="width: 120px">
                    <el-option v-for="item in productStatusOptions" :key="item" :label="item" :value="item" />
                  </el-select>
                  <el-button size="small" @click="updateProductStatus(row)">保存</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ productPage }} 页，共 {{ productTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="productPage"
              :page-size="productPageSize"
              :total="productTotal"
              @current-change="handleProductPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="会员订单" name="orders">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select v-model="orderFilters.status" clearable placeholder="全部状态" style="width: 160px">
              <el-option v-for="item in orderStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-input v-model="orderFilters.user_id" clearable placeholder="用户ID" style="width: 180px" />
            <el-button :loading="orderExporting" @click="exportOrdersFilteredCSV">导出筛选CSV</el-button>
            <el-button type="primary" plain @click="applyOrderFilters">查询</el-button>
            <el-button @click="resetOrderFilters">重置</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="orders" border stripe v-loading="orderLoading" empty-text="暂无会员订单">
            <el-table-column prop="id" label="订单ID" min-width="140" />
            <el-table-column prop="order_no" label="订单号" min-width="150" />
            <el-table-column prop="user_id" label="用户ID" min-width="130" />
            <el-table-column prop="product_id" label="产品ID" min-width="130" />
            <el-table-column prop="amount" label="金额" min-width="90" />
            <el-table-column prop="pay_channel" label="支付渠道" min-width="100" />
            <el-table-column label="状态" min-width="250">
              <template #default="{ row }">
                <div class="inline-actions">
                  <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
                  <el-select v-model="draftOrderStatusMap[row.id]" style="width: 120px">
                    <el-option v-for="item in orderStatusOptions" :key="item" :label="item" :value="item" />
                  </el-select>
                  <el-button size="small" @click="updateOrderStatus(row)">保存</el-button>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="paid_at" label="支付时间" min-width="170" />
            <el-table-column prop="created_at" label="创建时间" min-width="170" />
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ orderPage }} 页，共 {{ orderTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="orderPage"
              :page-size="orderPageSize"
              :total="orderTotal"
              @current-change="handleOrderPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="VIP配额配置" name="quotas">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select v-model="quotaFilters.member_level" clearable placeholder="会员等级" style="width: 160px">
              <el-option v-for="item in memberLevelOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-select v-model="quotaFilters.status" clearable placeholder="状态" style="width: 140px">
              <el-option v-for="item in productStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-button type="primary" plain @click="applyQuotaFilters">查询</el-button>
            <el-button @click="resetQuotaFilters">重置</el-button>
            <el-button type="primary" @click="openCreateQuotaDialog">新增配额配置</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="quotaItems" border stripe v-loading="quotaLoading" empty-text="暂无VIP配额配置">
            <el-table-column prop="id" label="配置ID" min-width="150" />
            <el-table-column prop="member_level" label="会员等级" min-width="110" />
            <el-table-column prop="doc_read_limit" label="文档阅读上限" min-width="120" />
            <el-table-column prop="news_subscribe_limit" label="新闻订阅上限" min-width="120" />
            <el-table-column prop="reset_cycle" label="重置周期" min-width="100" />
            <el-table-column label="状态" min-width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="effective_at" label="生效时间" min-width="170" />
            <el-table-column prop="updated_at" label="更新时间" min-width="170" />
            <el-table-column label="操作" align="right" min-width="120">
              <template #default="{ row }">
                <el-button size="small" @click="openEditQuotaDialog(row)">编辑</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ quotaPage }} 页，共 {{ quotaTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="quotaPage"
              :page-size="quotaPageSize"
              :total="quotaTotal"
              @current-change="handleQuotaPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="用户配额" name="user-quotas">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-input v-model="usageFilters.user_id" clearable placeholder="用户ID" style="width: 180px" />
            <el-input v-model="usageFilters.period_key" clearable placeholder="周期，如 2026-02" style="width: 180px" />
            <el-button type="primary" plain @click="applyUsageFilters">查询</el-button>
            <el-button @click="resetUsageFilters">重置</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="usageItems" border stripe v-loading="usageLoading" empty-text="暂无用户配额数据">
            <el-table-column prop="user_id" label="用户ID" min-width="130" />
            <el-table-column prop="member_level" label="会员等级" min-width="110" />
            <el-table-column prop="period_key" label="周期" min-width="120" />
            <el-table-column prop="doc_read_limit" label="文档上限" min-width="100" />
            <el-table-column prop="doc_read_used" label="文档已用" min-width="100" />
            <el-table-column label="文档剩余" min-width="100">
              <template #default="{ row }">
                {{ (row.doc_read_limit || 0) - (row.doc_read_used || 0) }}
              </template>
            </el-table-column>
            <el-table-column prop="news_subscribe_limit" label="新闻上限" min-width="100" />
            <el-table-column prop="news_subscribe_used" label="新闻已用" min-width="100" />
            <el-table-column label="新闻剩余" min-width="100">
              <template #default="{ row }">
                {{ (row.news_subscribe_limit || 0) - (row.news_subscribe_used || 0) }}
              </template>
            </el-table-column>
            <el-table-column prop="updated_at" label="更新时间" min-width="170" />
            <el-table-column label="操作" align="right" min-width="120">
              <template #default="{ row }">
                <el-button size="small" @click="openAdjustUsageDialog(row)">调整配额</el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ usagePage }} 页，共 {{ usageTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="usagePage"
              :page-size="usagePageSize"
              :total="usageTotal"
              @current-change="handleUsagePageChange"
            />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="productFormVisible" title="新增会员产品" width="520px" destroy-on-close>
      <el-form label-width="120px">
        <el-form-item label="产品名称" required>
          <el-input v-model="productForm.name" placeholder="如 VIP1 月卡" />
        </el-form-item>
        <el-form-item label="价格" required>
          <el-input-number v-model="productForm.price" :min="0.01" :step="1" :precision="2" style="width: 100%" />
        </el-form-item>
        <el-form-item label="会员等级">
          <el-select v-model="productForm.member_level" style="width: 100%">
            <el-option v-for="item in productMemberLevelOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="时长(天)">
          <el-input-number v-model="productForm.duration_days" :min="1" :step="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="productForm.status" style="width: 100%">
            <el-option v-for="item in productStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="productFormVisible = false">取消</el-button>
        <el-button type="primary" :loading="productSubmitting" @click="submitProduct">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="quotaDialogVisible"
      :title="quotaFormMode === 'create' ? '新增VIP配额配置' : '编辑VIP配额配置'"
      width="580px"
      destroy-on-close
    >
      <el-alert
        title="effective_at 需使用 RFC3339，例如 2026-12-31T00:00:00Z"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 12px"
      />
      <el-form label-width="130px">
        <el-form-item label="会员等级" required>
          <el-select v-model="quotaForm.member_level" style="width: 100%" :disabled="quotaFormMode === 'edit'">
            <el-option v-for="item in memberLevelOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="文档阅读上限" required>
          <el-input-number v-model="quotaForm.doc_read_limit" :min="0" :step="10" style="width: 100%" />
        </el-form-item>
        <el-form-item label="新闻订阅上限" required>
          <el-input-number v-model="quotaForm.news_subscribe_limit" :min="0" :step="10" style="width: 100%" />
        </el-form-item>
        <el-form-item label="重置周期" required>
          <el-select v-model="quotaForm.reset_cycle" style="width: 100%">
            <el-option v-for="item in resetCycleOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" required>
          <el-select v-model="quotaForm.status" style="width: 100%">
            <el-option v-for="item in productStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="生效时间" required>
          <el-input v-model="quotaForm.effective_at" placeholder="2026-12-31T00:00:00Z" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="quotaDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="quotaSubmitting" @click="submitQuotaForm">
          {{ quotaFormMode === "create" ? "创建" : "更新" }}
        </el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="usageAdjustDialogVisible" title="调整用户配额" width="560px" destroy-on-close>
      <el-form label-width="130px">
        <el-form-item label="用户ID">
          <el-text>{{ usageAdjustForm.user_id || "-" }}</el-text>
        </el-form-item>
        <el-form-item label="周期">
          <el-text>{{ usageAdjustForm.period_key || "-" }}</el-text>
        </el-form-item>
        <el-form-item label="文档增量">
          <el-input-number v-model="usageAdjustForm.doc_read_delta" :step="10" style="width: 100%" />
        </el-form-item>
        <el-form-item label="新闻增量">
          <el-input-number v-model="usageAdjustForm.news_subscribe_delta" :step="10" style="width: 100%" />
        </el-form-item>
        <el-form-item label="调整原因">
          <el-input
            v-model="usageAdjustForm.reason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="例如：活动补偿"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="usageAdjustDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="usageAdjusting" @click="submitAdjustUsage">确认调整</el-button>
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
</style>
