<script setup>
import { onMounted, ref } from "vue";
import { getExperimentAnalyticsSummary } from "../api/admin";

const experimentSummaryLoading = ref(false);
const experimentSummaryDays = ref(7);
const errorMessage = ref("");
const experimentOverview = ref({
  days: 7,
  total_events: 0,
  total_experiments: 0,
  exposure_count: 0,
  click_count: 0,
  upgrade_intent_count: 0,
  payment_success_count: 0,
  renewal_success_count: 0,
  click_through_rate: 0,
  upgrade_per_click_rate: 0,
  upgrade_per_exposure_rate: 0,
  paid_per_upgrade_rate: 0,
  paid_per_click_rate: 0,
  paid_per_exposure_rate: 0,
  last_event_at: ""
});
const experimentItems = ref([]);
const experimentPageBreakdown = ref([]);
const experimentDailyTrend = ref([]);
const experimentPayChannelBreakdown = ref([]);
const experimentDeviceBreakdown = ref([]);
const experimentUserStageBreakdown = ref([]);
const experimentVariantDailyTrend = ref([]);

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function formatDateTime(value) {
  const timestamp = Date.parse(value);
  if (Number.isNaN(timestamp)) {
    return "-";
  }
  const date = new Date(timestamp);
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  const hour = `${date.getHours()}`.padStart(2, "0");
  const minute = `${date.getMinutes()}`.padStart(2, "0");
  return `${year}-${month}-${day} ${hour}:${minute}`;
}

function formatPercent(value, digits = 2) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return "-";
  }
  return `${(numeric * 100).toFixed(digits)}%`;
}

function formatExperimentPageLabel(value) {
  const key = String(value || "").trim().toLowerCase();
  if (!key) {
    return "-";
  }
  const labels = {
    membership: "会员页",
    strategy: "策略页",
    home: "首页",
    news: "资讯页",
    archive: "历史档案页"
  };
  return labels[key] || key;
}

function formatPayChannelLabel(value) {
  const key = String(value || "").trim().toUpperCase();
  if (!key) {
    return "-";
  }
  const labels = {
    ALIPAY: "支付宝",
    WECHAT: "微信支付",
    CARD: "银行卡",
    YOLKPAY: "蛋黄支付",
    UNKNOWN: "未记录"
  };
  return labels[key] || key;
}

function formatDeviceTypeLabel(value) {
  const key = String(value || "").trim().toLowerCase();
  if (!key) {
    return "未记录";
  }
  const labels = {
    mobile: "移动端",
    desktop: "桌面端",
    tablet: "平板",
    unknown: "未记录"
  };
  return labels[key] || key;
}

function formatUserStageLabel(value) {
  const key = String(value || "").trim().toUpperCase();
  if (!key) {
    return "未记录";
  }
  const labels = {
    VISITOR: "游客",
    REGISTERED: "注册用户",
    VIP: "已激活会员",
    EXPIRED: "过期会员",
    PAID_PENDING_KYC: "付费待实名激活",
    UNKNOWN: "未记录"
  };
  return labels[key] || key;
}

async function fetchExperimentSummary() {
  experimentSummaryLoading.value = true;
  errorMessage.value = "";
  try {
    const data = await getExperimentAnalyticsSummary({
      days: Number(experimentSummaryDays.value) || 7
    });
    experimentOverview.value = data?.overview || experimentOverview.value;
    experimentItems.value = Array.isArray(data?.items) ? data.items : [];
    experimentPageBreakdown.value = Array.isArray(data?.page_breakdown) ? data.page_breakdown : [];
    experimentDailyTrend.value = Array.isArray(data?.daily_trend) ? data.daily_trend : [];
    experimentPayChannelBreakdown.value = Array.isArray(data?.pay_channel_breakdown) ? data.pay_channel_breakdown : [];
    experimentDeviceBreakdown.value = Array.isArray(data?.device_breakdown) ? data.device_breakdown : [];
    experimentUserStageBreakdown.value = Array.isArray(data?.user_stage_breakdown) ? data.user_stage_breakdown : [];
    experimentVariantDailyTrend.value = Array.isArray(data?.variant_daily_trend) ? data.variant_daily_trend : [];
    experimentSummaryDays.value = Number(data?.days) || experimentSummaryDays.value;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载实验埋点看板失败");
  } finally {
    experimentSummaryLoading.value = false;
  }
}

onMounted(() => {
  fetchExperimentSummary();
});
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">A/B 实验埋点看板</h1>
        <p class="muted">把实验、分组、页面、设备、用户阶段和支付结果拆开看，方便单独追踪投放与转化质量。</p>
      </div>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-select v-model="experimentSummaryDays" style="width: 120px">
          <el-option :value="3" label="近 3 天" />
          <el-option :value="7" label="近 7 天" />
          <el-option :value="14" label="近 14 天" />
          <el-option :value="30" label="近 30 天" />
        </el-select>
        <el-button :loading="experimentSummaryLoading" @click="fetchExperimentSummary">刷新埋点看板</el-button>
      </div>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />

    <div class="card">
      <div class="toolbar" style="margin-bottom: 8px; flex-wrap: wrap">
        <el-tag type="info">实验 {{ experimentOverview.total_experiments || 0 }} 个</el-tag>
        <el-tag type="success">曝光 {{ experimentOverview.exposure_count || 0 }}</el-tag>
        <el-tag type="warning">点击 {{ experimentOverview.click_count || 0 }}</el-tag>
        <el-tag type="danger">升级意图 {{ experimentOverview.upgrade_intent_count || 0 }}</el-tag>
        <el-tag type="success">支付成功 {{ experimentOverview.payment_success_count || 0 }}</el-tag>
        <el-tag type="warning">续费成功 {{ experimentOverview.renewal_success_count || 0 }}</el-tag>
        <el-tag type="info">CTR {{ formatPercent(experimentOverview.click_through_rate || 0) }}</el-tag>
        <el-tag type="success">升级/点击 {{ formatPercent(experimentOverview.upgrade_per_click_rate || 0) }}</el-tag>
        <el-tag type="warning">升级/曝光 {{ formatPercent(experimentOverview.upgrade_per_exposure_rate || 0) }}</el-tag>
        <el-tag type="danger">支付/升级 {{ formatPercent(experimentOverview.paid_per_upgrade_rate || 0) }}</el-tag>
        <el-tag type="info">支付/点击 {{ formatPercent(experimentOverview.paid_per_click_rate || 0) }}</el-tag>
        <el-tag type="success">支付/曝光 {{ formatPercent(experimentOverview.paid_per_exposure_rate || 0) }}</el-tag>
        <el-text type="info">
          最近事件：{{ experimentOverview.last_event_at ? formatDateTime(experimentOverview.last_event_at) : "-" }}
        </el-text>
      </div>

      <el-table
        :data="experimentItems"
        border
        stripe
        size="small"
        v-loading="experimentSummaryLoading"
        empty-text="暂无实验埋点数据"
      >
        <el-table-column prop="experiment_key" label="实验" min-width="180" />
        <el-table-column prop="variant_key" label="分组" min-width="100" />
        <el-table-column label="页面" min-width="110">
          <template #default="{ row }">
            {{ formatExperimentPageLabel(row.page_key) }}
          </template>
        </el-table-column>
        <el-table-column prop="user_stage" label="用户阶段" min-width="110" />
        <el-table-column prop="exposure_count" label="曝光" min-width="88" />
        <el-table-column prop="click_count" label="点击" min-width="88" />
        <el-table-column prop="upgrade_intent_count" label="升级意图" min-width="100" />
        <el-table-column prop="payment_success_count" label="支付成功" min-width="100" />
        <el-table-column prop="renewal_success_count" label="续费成功" min-width="100" />
        <el-table-column label="CTR" min-width="90">
          <template #default="{ row }">
            {{ formatPercent(row.click_through_rate || 0) }}
          </template>
        </el-table-column>
        <el-table-column label="升级/点击" min-width="100">
          <template #default="{ row }">
            {{ formatPercent(row.upgrade_per_click_rate || 0) }}
          </template>
        </el-table-column>
        <el-table-column label="升级/曝光" min-width="100">
          <template #default="{ row }">
            {{ formatPercent(row.upgrade_per_exposure_rate || 0) }}
          </template>
        </el-table-column>
        <el-table-column label="支付/升级" min-width="100">
          <template #default="{ row }">
            {{ formatPercent(row.paid_per_upgrade_rate || 0) }}
          </template>
        </el-table-column>
        <el-table-column label="支付/点击" min-width="100">
          <template #default="{ row }">
            {{ formatPercent(row.paid_per_click_rate || 0) }}
          </template>
        </el-table-column>
        <el-table-column label="支付/曝光" min-width="100">
          <template #default="{ row }">
            {{ formatPercent(row.paid_per_exposure_rate || 0) }}
          </template>
        </el-table-column>
        <el-table-column prop="last_event_at" label="最近事件" min-width="170">
          <template #default="{ row }">
            {{ row.last_event_at ? formatDateTime(row.last_event_at) : "-" }}
          </template>
        </el-table-column>
      </el-table>

      <div class="experiment-insight-grid">
        <div class="experiment-insight-card">
          <div class="experiment-insight-head">
            <h3>来源页面拆解</h3>
            <span>看哪个入口最会带支付</span>
          </div>
          <el-table :data="experimentPageBreakdown" border stripe size="small" empty-text="暂无页面拆解数据">
            <el-table-column label="页面" min-width="110">
              <template #default="{ row }">
                {{ formatExperimentPageLabel(row.page_key) }}
              </template>
            </el-table-column>
            <el-table-column prop="exposure_count" label="曝光" min-width="80" />
            <el-table-column prop="upgrade_intent_count" label="升级意图" min-width="96" />
            <el-table-column label="支付成功" min-width="96">
              <template #default="{ row }">
                {{ (row.payment_success_count || 0) + (row.renewal_success_count || 0) }}
              </template>
            </el-table-column>
            <el-table-column label="支付/升级" min-width="96">
              <template #default="{ row }">
                {{ formatPercent(row.paid_per_upgrade_rate || 0) }}
              </template>
            </el-table-column>
            <el-table-column label="支付/曝光" min-width="96">
              <template #default="{ row }">
                {{ formatPercent(row.paid_per_exposure_rate || 0) }}
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="experiment-insight-card">
          <div class="experiment-insight-head">
            <h3>支付渠道拆解</h3>
            <span>看成交集中在哪个渠道</span>
          </div>
          <el-table :data="experimentPayChannelBreakdown" border stripe size="small" empty-text="暂无支付渠道数据">
            <el-table-column label="支付渠道" min-width="110">
              <template #default="{ row }">
                {{ formatPayChannelLabel(row.pay_channel) }}
              </template>
            </el-table-column>
            <el-table-column prop="payment_success_count" label="新购成功" min-width="90" />
            <el-table-column prop="renewal_success_count" label="续费成功" min-width="90" />
            <el-table-column prop="paid_success_count" label="合计" min-width="80" />
            <el-table-column label="支付占比" min-width="96">
              <template #default="{ row }">
                {{ formatPercent(row.paid_share_rate || 0) }}
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="experiment-insight-card">
          <div class="experiment-insight-head">
            <h3>设备拆解</h3>
            <span>看移动端和桌面端的转化差异</span>
          </div>
          <el-table :data="experimentDeviceBreakdown" border stripe size="small" empty-text="暂无设备拆解数据">
            <el-table-column prop="experiment_key" label="实验" min-width="150" />
            <el-table-column prop="variant_key" label="分组" min-width="90" />
            <el-table-column label="页面" min-width="110">
              <template #default="{ row }">
                {{ formatExperimentPageLabel(row.page_key) }}
              </template>
            </el-table-column>
            <el-table-column label="设备" min-width="100">
              <template #default="{ row }">
                {{ formatDeviceTypeLabel(row.device_type) }}
              </template>
            </el-table-column>
            <el-table-column prop="exposure_count" label="曝光" min-width="80" />
            <el-table-column prop="upgrade_intent_count" label="升级意图" min-width="96" />
            <el-table-column label="支付/曝光" min-width="96">
              <template #default="{ row }">
                {{ formatPercent(row.paid_per_exposure_rate || 0) }}
              </template>
            </el-table-column>
            <el-table-column prop="last_event_at" label="最近事件" min-width="160">
              <template #default="{ row }">
                {{ row.last_event_at ? formatDateTime(row.last_event_at) : "-" }}
              </template>
            </el-table-column>
          </el-table>
        </div>

        <div class="experiment-insight-card">
          <div class="experiment-insight-head">
            <h3>用户阶段拆解</h3>
            <span>看游客、注册、待实名激活和已激活会员分别怎么转化</span>
          </div>
          <el-table :data="experimentUserStageBreakdown" border stripe size="small" empty-text="暂无用户阶段数据">
            <el-table-column prop="experiment_key" label="实验" min-width="150" />
            <el-table-column prop="variant_key" label="分组" min-width="90" />
            <el-table-column label="页面" min-width="110">
              <template #default="{ row }">
                {{ formatExperimentPageLabel(row.page_key) }}
              </template>
            </el-table-column>
            <el-table-column label="用户阶段" min-width="130">
              <template #default="{ row }">
                {{ formatUserStageLabel(row.user_stage) }}
              </template>
            </el-table-column>
            <el-table-column prop="exposure_count" label="曝光" min-width="80" />
            <el-table-column prop="upgrade_intent_count" label="升级意图" min-width="96" />
            <el-table-column label="支付/升级" min-width="96">
              <template #default="{ row }">
                {{ formatPercent(row.paid_per_upgrade_rate || 0) }}
              </template>
            </el-table-column>
            <el-table-column prop="last_event_at" label="最近事件" min-width="160">
              <template #default="{ row }">
                {{ row.last_event_at ? formatDateTime(row.last_event_at) : "-" }}
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>

      <div class="experiment-trend-card">
        <div class="experiment-insight-head">
          <h3>日期趋势</h3>
          <span>看转化节奏是否连续放大</span>
        </div>
        <el-table :data="experimentDailyTrend" border stripe size="small" empty-text="暂无日期趋势数据">
          <el-table-column prop="date" label="日期" min-width="110" />
          <el-table-column prop="exposure_count" label="曝光" min-width="80" />
          <el-table-column prop="click_count" label="点击" min-width="80" />
          <el-table-column prop="upgrade_intent_count" label="升级意图" min-width="96" />
          <el-table-column label="支付成功" min-width="96">
            <template #default="{ row }">
              {{ (row.payment_success_count || 0) + (row.renewal_success_count || 0) }}
            </template>
          </el-table-column>
          <el-table-column label="CTR" min-width="90">
            <template #default="{ row }">
              {{ formatPercent(row.click_through_rate || 0) }}
            </template>
          </el-table-column>
          <el-table-column label="升级/点击" min-width="96">
            <template #default="{ row }">
              {{ formatPercent(row.upgrade_per_click_rate || 0) }}
            </template>
          </el-table-column>
          <el-table-column label="支付/曝光" min-width="96">
            <template #default="{ row }">
              {{ formatPercent(row.paid_per_exposure_rate || 0) }}
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="experiment-trend-card">
        <div class="experiment-insight-head">
          <h3>实验组趋势</h3>
          <span>按实验、分组、页面、设备和用户阶段继续下钻</span>
        </div>
        <el-table :data="experimentVariantDailyTrend" border stripe size="small" empty-text="暂无实验组趋势数据">
          <el-table-column prop="date" label="日期" min-width="110" />
          <el-table-column prop="experiment_key" label="实验" min-width="150" />
          <el-table-column prop="variant_key" label="分组" min-width="90" />
          <el-table-column label="页面" min-width="110">
            <template #default="{ row }">
              {{ formatExperimentPageLabel(row.page_key) }}
            </template>
          </el-table-column>
          <el-table-column label="设备" min-width="100">
            <template #default="{ row }">
              {{ formatDeviceTypeLabel(row.device_type) }}
            </template>
          </el-table-column>
          <el-table-column label="用户阶段" min-width="130">
            <template #default="{ row }">
              {{ formatUserStageLabel(row.user_stage) }}
            </template>
          </el-table-column>
          <el-table-column prop="exposure_count" label="曝光" min-width="80" />
          <el-table-column prop="click_count" label="点击" min-width="80" />
          <el-table-column prop="upgrade_intent_count" label="升级意图" min-width="96" />
          <el-table-column label="支付成功" min-width="96">
            <template #default="{ row }">
              {{ (row.payment_success_count || 0) + (row.renewal_success_count || 0) }}
            </template>
          </el-table-column>
          <el-table-column label="支付/曝光" min-width="96">
            <template #default="{ row }">
              {{ formatPercent(row.paid_per_exposure_rate || 0) }}
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-bottom: 16px;
}

.page-title {
  margin: 0;
  font-size: 24px;
  font-weight: 700;
  color: #111827;
}

.muted {
  margin: 6px 0 0;
  color: #64748b;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
}

.card {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  background: #fff;
  padding: 16px;
}

.experiment-insight-grid {
  margin-top: 12px;
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(360px, 1fr));
  gap: 12px;
}

.experiment-insight-card,
.experiment-trend-card {
  margin-top: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 10px;
  padding: 10px;
  background: #fcfcfd;
}

.experiment-insight-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.experiment-insight-head h3 {
  margin: 0;
  font-size: 14px;
  color: #111827;
}

.experiment-insight-head span {
  font-size: 12px;
  color: #64748b;
}

@media (max-width: 960px) {
  .page-header {
    flex-direction: column;
  }
}
</style>
