<script setup>
import { computed, onMounted, ref } from "vue";
import { countUnreadWorkflowMessages, getDashboardOverview } from "../api/admin";

const loading = ref(false);
const errorMessage = ref("");
const updatedAt = ref("");

const overview = ref({
  total_users: 0,
  active_users: 0,
  kyc_approved_users: 0,
  vip_users: 0,
  today_new_users: 0,
  today_paid_orders: 0,
  today_published_stocks: 0,
  today_published_news: 0
});

const unreadCount = ref(0);

const cards = computed(() => [
  { label: "总用户数", value: overview.value.total_users },
  { label: "活跃用户", value: overview.value.active_users },
  { label: "实名通过", value: overview.value.kyc_approved_users },
  { label: "VIP 用户", value: overview.value.vip_users },
  { label: "今日新增", value: overview.value.today_new_users },
  { label: "今日付费订单", value: overview.value.today_paid_orders },
  { label: "今日发布股票推荐", value: overview.value.today_published_stocks },
  { label: "今日发布新闻", value: overview.value.today_published_news },
  { label: "未读流程消息", value: unreadCount.value }
]);

async function fetchOverview() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const [dashboardData, unreadData] = await Promise.all([
      getDashboardOverview(),
      countUnreadWorkflowMessages({})
    ]);
    overview.value = dashboardData;
    unreadCount.value = unreadData.unread_count || 0;
    updatedAt.value = new Date().toLocaleString();
  } catch (error) {
    errorMessage.value = error.message || "加载仪表盘失败";
  } finally {
    loading.value = false;
  }
}

onMounted(fetchOverview);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">仪表盘</h1>
        <p class="muted">后端概览指标与流程消息状态</p>
      </div>
      <button class="btn btn-primary" :disabled="loading" @click="fetchOverview">
        {{ loading ? "刷新中..." : "刷新" }}
      </button>
    </div>

    <div v-if="errorMessage" class="error-message">{{ errorMessage }}</div>

    <div class="grid grid-4">
      <div v-for="item in cards" :key="item.label" class="metric-card">
        <div class="label">{{ item.label }}</div>
        <div class="value">{{ item.value ?? 0 }}</div>
      </div>
    </div>

    <p class="hint">最后刷新时间：{{ updatedAt || "-" }}</p>
  </div>
</template>
