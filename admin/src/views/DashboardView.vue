<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import {
  countUnreadWorkflowMessages,
  getDashboardOverview,
  getSchedulerJobMetrics,
  getWorkflowMetrics,
  listDataSources,
  listMembershipOrders,
  listRiskHits,
  listStockRecommendations
} from "../api/admin";

const router = useRouter();

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

const ops = ref({
  unread_messages: 0,
  pending_reviews: 0,
  today_job_failed: 0,
  today_job_running: 0,
  pending_membership_orders: 0,
  pending_risk_hits: 0,
  published_stock_recos: 0,
  active_data_sources: 0
});

const coreCards = computed(() => [
  { label: "总用户数", value: overview.value.total_users },
  { label: "活跃用户", value: overview.value.active_users },
  { label: "实名通过", value: overview.value.kyc_approved_users },
  { label: "VIP 用户", value: overview.value.vip_users },
  { label: "今日新增", value: overview.value.today_new_users },
  { label: "今日付费订单", value: overview.value.today_paid_orders },
  { label: "今日发布股票推荐", value: overview.value.today_published_stocks },
  { label: "今日发布新闻", value: overview.value.today_published_news }
]);

const opsCards = computed(() => [
  { label: "未读流程消息", value: ops.value.unread_messages, type: "warning" },
  { label: "待审核任务", value: ops.value.pending_reviews, type: "warning" },
  { label: "今日失败任务", value: ops.value.today_job_failed, type: "danger" },
  { label: "运行中任务", value: ops.value.today_job_running, type: "info" },
  { label: "待处理会员订单", value: ops.value.pending_membership_orders, type: "warning" },
  { label: "待处理风险命中", value: ops.value.pending_risk_hits, type: "danger" },
  { label: "已发布股票推荐", value: ops.value.published_stock_recos, type: "success" },
  { label: "活跃数据源", value: ops.value.active_data_sources, type: "success" }
]);

const quickLinks = [
  { title: "审核中心", desc: "处理待审任务与批量审批", path: "/review-center" },
  { title: "流程消息", desc: "查看告警并批量已读", path: "/workflow-messages" },
  { title: "系统任务", desc: "查看任务运行与重跑", path: "/system-jobs" },
  { title: "数据源管理", desc: "查看健康状态与日志", path: "/data-sources" },
  { title: "会员中心", desc: "管理产品、订单、配额", path: "/membership-center" },
  { title: "策略中心", desc: "维护股票与期货策略", path: "/market-center" },
  { title: "风控中心", desc: "规则、命中、提现、对账", path: "/risk-center" },
  { title: "安全中心", desc: "登录风控与解锁审计", path: "/auth-security" },
  { title: "系统配置", desc: "维护平台配置项", path: "/system-configs" },
  { title: "操作日志", desc: "审计关键变更记录", path: "/audit-logs" }
];

function statusTagType(type) {
  if (type === "danger") return "danger";
  if (type === "warning") return "warning";
  if (type === "success") return "success";
  return "info";
}

function openLink(path) {
  router.push(path);
}

async function fetchOverview() {
  loading.value = true;
  errorMessage.value = "";

  try {
    const [
      dashboardData,
      unreadData,
      workflowData,
      schedulerData,
      pendingOrderData,
      pendingRiskData,
      stockRecoData,
      dataSourceData
    ] = await Promise.all([
      getDashboardOverview(),
      countUnreadWorkflowMessages({}),
      getWorkflowMetrics({}),
      getSchedulerJobMetrics({}),
      listMembershipOrders({ status: "PENDING", page: 1, page_size: 1 }),
      listRiskHits({ status: "PENDING", page: 1, page_size: 1 }),
      listStockRecommendations({ status: "PUBLISHED", page: 1, page_size: 1 }),
      listDataSources({ page: 1, page_size: 200 })
    ]);

    overview.value = dashboardData || {};
    const dataSourceItems = dataSourceData.items || [];
    const activeDataSourceCount = dataSourceItems.filter((item) => (item.status || "").toUpperCase() === "ACTIVE").length;

    ops.value = {
      unread_messages: unreadData.unread_count || 0,
      pending_reviews: workflowData.pending_reviews || 0,
      today_job_failed: schedulerData.today_failed || 0,
      today_job_running: schedulerData.today_running || 0,
      pending_membership_orders: pendingOrderData.total || 0,
      pending_risk_hits: pendingRiskData.total || 0,
      published_stock_recos: stockRecoData.total || 0,
      active_data_sources: activeDataSourceCount
    };

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
        <p class="muted">后端核心指标与运维入口总览</p>
      </div>
      <el-button type="primary" :loading="loading" @click="fetchOverview">刷新</el-button>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <h3 style="margin: 0">核心业务指标</h3>
        <el-text type="info">来源：dashboard overview</el-text>
      </div>
      <div class="grid grid-4" v-loading="loading">
        <div v-for="item in coreCards" :key="item.label" class="metric-item">
          <div class="metric-label">{{ item.label }}</div>
          <div class="metric-value">{{ item.value ?? 0 }}</div>
        </div>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <h3 style="margin: 0">运营监控指标</h3>
        <el-text type="info">来源：workflow / jobs / risk / membership</el-text>
      </div>
      <div class="grid grid-4" v-loading="loading">
        <div v-for="item in opsCards" :key="item.label" class="metric-item">
          <div class="metric-label">{{ item.label }}</div>
          <div class="metric-value-row">
            <div class="metric-value">{{ item.value ?? 0 }}</div>
            <el-tag :type="statusTagType(item.type)">{{ item.type || "info" }}</el-tag>
          </div>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="section-header">
        <h3 style="margin: 0">快捷入口</h3>
        <el-text type="info">最后刷新时间：{{ updatedAt || "-" }}</el-text>
      </div>
      <div class="quick-grid">
        <button
          v-for="item in quickLinks"
          :key="item.path"
          type="button"
          class="quick-card"
          @click="openLink(item.path)"
        >
          <div class="quick-title">{{ item.title }}</div>
          <div class="quick-desc">{{ item.desc }}</div>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.metric-item {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 12px;
  background: #fff;
}

.metric-label {
  color: #6b7280;
  font-size: 12px;
}

.metric-value-row {
  margin-top: 6px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.metric-value {
  font-size: 24px;
  font-weight: 700;
  color: #111827;
}

.quick-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 10px;
}

.quick-card {
  border: 1px solid #dbe5ff;
  border-radius: 12px;
  background: linear-gradient(180deg, #ffffff 0%, #f7f9ff 100%);
  padding: 12px;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
}

.quick-card:hover {
  transform: translateY(-1px);
  border-color: #93c5fd;
  box-shadow: 0 8px 20px rgba(37, 99, 235, 0.08);
}

.quick-title {
  font-size: 14px;
  font-weight: 600;
  color: #1f2937;
}

.quick-desc {
  margin-top: 6px;
  font-size: 12px;
  color: #6b7280;
  line-height: 1.5;
}
</style>
