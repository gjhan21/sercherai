<script setup>
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import FuturesSelectionModuleShell from "../../components/FuturesSelectionModuleShell.vue";
import {
  createFuturesSelectionRun,
  getFuturesSelectionOverview,
  getFuturesSimulatedOverview,
  listFuturesSimulatedPositions,
  triggerFuturesSimulatedSettlement
} from "../../api/admin";
import {
  formatFuturesSelectionDateTime,
  formatFuturesSelectionLabel,
  formatFuturesSelectionMarketRegime,
  formatFuturesSelectionRunStatus,
  formatFuturesSelectionSource
} from "../../lib/futures-selection";
import { hasPermission } from "../../lib/session";

const router = useRouter();
const canManage = hasPermission("futures_selection.manage");
const loading = ref(false);
const running = ref(false);
const errorMessage = ref("");
const overview = ref({
  default_profile: null,
  latest_trade_date: "",
  latest_run: null,
  latest_success_run: null,
  market_regime: "",
  latest_approved_portfolio: [],
  pending_review_count: 0,
  data_freshness: {},
  warnings: []
});

// 期货模拟做单与持仓跟踪系统状态
const activeTab = ref("holding");
const settleDate = ref(new Date().toISOString().slice(0, 10));
const settling = ref(false);
const simLoading = ref(false);
const simulatedOverview = ref({
  total_trades: 0,
  active_holdings: 0,
  win_rate: 0,
  average_return: 0,
  total_return: 0,
  avg_hold_days: 0
});
const holdingPositions = ref([]);
const holdingTotal = ref(0);
const holdingPage = ref(1);

const closedPositions = ref([]);
const closedTotal = ref(0);
const closedPage = ref(1);
const pageSize = ref(10);

function runTagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "SUCCEEDED") return "success";
  if (normalized === "FAILED") return "danger";
  if (normalized === "RUNNING") return "warning";
  return "info";
}

async function fetchOverview() {
  loading.value = true;
  errorMessage.value = "";
  try {
    overview.value = await getFuturesSelectionOverview();
  } catch (error) {
    errorMessage.value = error?.message || "加载智能期货总览失败";
  } finally {
    loading.value = false;
  }
}

async function handleRunNow() {
  if (!overview.value?.default_profile?.id) {
    ElMessage.error("当前没有可用的默认期货配置");
    return;
  }
  running.value = true;
  try {
    const run = await createFuturesSelectionRun({
      trade_date: overview.value.latest_trade_date || new Date().toISOString().slice(0, 10),
      profile_id: overview.value.default_profile.id
    });
    ElMessage.success(`已完成运行 ${run.run_id}`);
    router.push({ name: "futures-selection-runs", query: { run_id: run.run_id } });
  } catch (error) {
    ElMessage.error(error?.message || "触发期货运行失败");
  } finally {
    running.value = false;
  }
}

async function fetchSimulatedOverview() {
  try {
    simulatedOverview.value = await getFuturesSimulatedOverview();
  } catch (error) {
    console.error("加载模拟做单概览失败", error);
  }
}

async function fetchHoldingPositions() {
  simLoading.value = true;
  try {
    const res = await listFuturesSimulatedPositions({
      status: "HOLDING",
      page: holdingPage.value,
      page_size: pageSize.value
    });
    holdingPositions.value = res?.items || [];
    holdingTotal.value = res?.total || 0;
  } catch (error) {
    ElMessage.error(error?.message || "加载模拟持仓失败");
  } finally {
    simLoading.value = false;
  }
}

async function fetchClosedPositions() {
  simLoading.value = true;
  try {
    const res = await listFuturesSimulatedPositions({
      status: "CLOSED",
      page: closedPage.value,
      page_size: pageSize.value
    });
    closedPositions.value = res?.items || [];
    closedTotal.value = res?.total || 0;
  } catch (error) {
    ElMessage.error(error?.message || "加载历史模拟成交记录失败");
  } finally {
    simLoading.value = false;
  }
}

async function handleTriggerSettle() {
  settling.value = true;
  try {
    await triggerFuturesSimulatedSettlement(settleDate.value);
    ElMessage.success("触发模拟结算跑批成功");
    await loadAllSimulatedData();
  } catch (error) {
    ElMessage.error(error?.message || "触发模拟结算跑批失败");
  } finally {
    settling.value = false;
  }
}

async function loadAllSimulatedData() {
  await fetchSimulatedOverview();
  await Promise.all([
    fetchHoldingPositions(),
    fetchClosedPositions()
  ]);
}

function getReasonLabel(reason) {
  if (!reason) return "已结单";
  switch (reason.toUpperCase()) {
    case "TAKE_PROFIT":
      return "止盈平仓";
    case "STOP_LOSS":
      return "止损平仓";
    case "EXPIRED":
      return "到期强平";
    case "MANUAL":
      return "手动平仓";
    default:
      return reason;
  }
}

function getReasonTagType(reason) {
  if (!reason) return "info";
  switch (reason.toUpperCase()) {
    case "TAKE_PROFIT":
      return "success";
    case "STOP_LOSS":
      return "danger";
    case "EXPIRED":
      return "warning";
    case "MANUAL":
      return "info";
    default:
      return "info";
  }
}

onMounted(async () => {
  await fetchOverview();
  await loadAllSimulatedData();
});
</script>

<template>
  <FuturesSelectionModuleShell
    title="智能期货总览"
    description="智能期货是您的‘科研实验室’与‘生产线’，负责算法跑批与候选集生成。生成的结果将同步至策略中心进行最终审核与发布。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchOverview">刷新总览</el-button>
        <el-button type="success" plain @click="router.push({ path: '/market-center' })">
          前往策略中心 (运营台)
        </el-button>
        <el-button v-if="canManage" type="primary" :loading="running" @click="handleRunNow">
          立即运行
        </el-button>
      </div>
    </template>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />

    <div class="stock-overview-grid">
      <div class="card" v-loading="loading">
        <div class="card-title">默认配置</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="名称">
            {{ overview.default_profile?.name || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="版本">
            v{{ overview.default_profile?.current_version || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="风格">
            {{ overview.default_profile?.style_default || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="合约范围">
            {{ overview.default_profile?.contract_scope || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="默认模板">
            {{ overview.default_profile?.template_name || "-" }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">最近运行</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="运行编号">
            {{ overview.latest_run?.run_id || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="runTagType(overview.latest_run?.status)">
              {{ formatFuturesSelectionRunStatus(overview.latest_run?.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="交易日">
            {{ overview.latest_run?.trade_date || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="组合数量">
            {{ overview.latest_run?.selected_count || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="市场状态">
            {{ formatFuturesSelectionMarketRegime(overview.market_regime) }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">最近成功运行</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="运行编号">
            {{ overview.latest_success_run?.run_id || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="发布次数">
            {{ overview.latest_success_run?.publish_count || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="图快照">
            {{ overview.data_freshness?.graph_snapshot_id || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="上下文交易日">
            {{ overview.data_freshness?.selected_trade_date || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="行情来源">
            {{ formatFuturesSelectionSource(overview.data_freshness?.price_source) }}
          </el-descriptions-item>
          <el-descriptions-item label="待审核数量">
            {{ overview.pending_review_count || 0 }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <div class="card" v-if="Array.isArray(overview.latest_approved_portfolio) && overview.latest_approved_portfolio.length">
      <div class="card-title">最近已发布组合</div>
      <el-tag
        v-for="item in overview.latest_approved_portfolio"
        :key="`${item.contract}-${item.rank}`"
        type="success"
        style="margin-right: 8px; margin-bottom: 8px"
      >
        {{ item.rank }}. {{ item.contract }} / {{ item.direction || "-" }} / {{ formatFuturesSelectionLabel(item.portfolio_role || "SATELLITE") }}
      </el-tag>
    </div>

    <div class="card" v-if="Array.isArray(overview.warnings) && overview.warnings.length">
      <div class="card-title">运行提醒</div>
      <div class="chip-wrap">
        <el-tag v-for="warning in overview.warnings" :key="warning" type="warning">
          {{ warning }}
        </el-tag>
      </div>
    </div>

    <div class="card">
      <div class="card-title">数据新鲜度</div>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="图谱摘要" :span="2">
          {{ overview.data_freshness?.graph_summary || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="图写入状态">
          {{ overview.data_freshness?.graph_write_status || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="最近完成时间">
          {{ formatFuturesSelectionDateTime(overview.data_freshness?.completed_at || overview.latest_success_run?.completed_at) }}
        </el-descriptions-item>
      </el-descriptions>
    </div>

    <!-- 期货模拟交易与持仓跟踪系统 -->
    <div class="card" v-loading="simLoading" style="margin-top: 20px;">
      <div class="card-header-flex" style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <span class="card-title" style="margin-bottom: 0;">模拟做单与持仓跟踪系统</span>
        <div style="display: flex; align-items: center; gap: 8px;">
          <span style="font-size: 13px; color: var(--el-text-color-secondary);">交易日:</span>
          <el-date-picker
            v-model="settleDate"
            type="date"
            value-format="YYYY-MM-DD"
            placeholder="选择交易日"
            size="small"
            style="width: 130px;"
          />
          <el-button type="warning" size="small" :loading="settling" @click="handleTriggerSettle">
            触发模拟结算跑批
          </el-button>
        </div>
      </div>

      <!-- Overview Stats Cards -->
      <div class="sim-stats-grid" style="display: grid; grid-template-columns: repeat(6, 1fr); gap: 12px; margin-bottom: 20px;">
        <div class="stat-card" style="background: var(--el-fill-color-blank); border: 1px solid var(--el-border-color-light); border-radius: 8px; padding: 12px; text-align: center;">
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-bottom: 4px;">累计模拟做单</div>
          <div style="font-size: 18px; font-weight: bold; color: var(--el-text-color-primary);">{{ simulatedOverview.total_trades || 0 }} 笔</div>
        </div>
        <div class="stat-card" style="background: var(--el-fill-color-blank); border: 1px solid var(--el-border-color-light); border-radius: 8px; padding: 12px; text-align: center;">
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-bottom: 4px;">当前持仓中</div>
          <div style="font-size: 18px; font-weight: bold; color: var(--el-color-primary);">{{ simulatedOverview.active_holdings || 0 }} 笔</div>
        </div>
        <div class="stat-card" style="background: var(--el-fill-color-blank); border: 1px solid var(--el-border-color-light); border-radius: 8px; padding: 12px; text-align: center;">
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-bottom: 4px;">已结单胜率</div>
          <div style="font-size: 18px; font-weight: bold; color: #f56c6c;">{{ (simulatedOverview.win_rate * 100).toFixed(2) }}%</div>
        </div>
        <div class="stat-card" style="background: var(--el-fill-color-blank); border: 1px solid var(--el-border-color-light); border-radius: 8px; padding: 12px; text-align: center;">
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-bottom: 4px;">累计盈亏率</div>
          <div :style="{ fontSize: '18px', fontWeight: 'bold', color: simulatedOverview.total_return >= 0 ? '#f56c6c' : '#67c23a' }">
            {{ simulatedOverview.total_return >= 0 ? '+' : '' }}{{ (simulatedOverview.total_return * 100).toFixed(2) }}%
          </div>
        </div>
        <div class="stat-card" style="background: var(--el-fill-color-blank); border: 1px solid var(--el-border-color-light); border-radius: 8px; padding: 12px; text-align: center;">
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-bottom: 4px;">已结单平均收益</div>
          <div :style="{ fontSize: '18px', fontWeight: 'bold', color: simulatedOverview.average_return >= 0 ? '#f56c6c' : '#67c23a' }">
            {{ simulatedOverview.average_return >= 0 ? '+' : '' }}{{ (simulatedOverview.average_return * 100).toFixed(2) }}%
          </div>
        </div>
        <div class="stat-card" style="background: var(--el-fill-color-blank); border: 1px solid var(--el-border-color-light); border-radius: 8px; padding: 12px; text-align: center;">
          <div style="font-size: 12px; color: var(--el-text-color-secondary); margin-bottom: 4px;">平均持有天数</div>
          <div style="font-size: 18px; font-weight: bold; color: var(--el-text-color-primary);">{{ simulatedOverview.avg_hold_days ? simulatedOverview.avg_hold_days.toFixed(1) : '0.0' }} 天</div>
        </div>
      </div>

      <!-- Tabbed Position Tracking -->
      <el-tabs v-model="activeTab" class="simulated-tabs">
        <el-tab-pane label="当前模拟持仓 (HOLDING)" name="holding">
          <el-table :data="holdingPositions" border stripe size="small" style="width: 100%">
            <el-table-column prop="open_date" label="建仓日期" width="100" />
            <el-table-column prop="contract" label="合约代码" width="100" />
            <el-table-column prop="name" label="合约名称" min-width="120" />
            <el-table-column prop="direction" label="方向" width="95">
              <template #default="{ row }">
                <el-tag :type="row.direction === 'LONG' ? 'danger' : 'primary'" size="small">
                  {{ row.direction === 'LONG' ? '多头 LONG' : '空头 SHORT' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="open_price" label="建仓价" width="90" />
            <el-table-column prop="current_price" label="当前价" width="90" />
            <el-table-column label="浮动盈亏率" width="100">
              <template #default="{ row }">
                <span :style="row.return_rate >= 0 ? 'color: #f56c6c; font-weight: 600;' : 'color: #67c23a; font-weight: 600;'">
                  {{ row.return_rate >= 0 ? '+' : '' }}{{ (row.return_rate * 100).toFixed(2) }}%
                </span>
              </template>
            </el-table-column>
            <el-table-column label="浮动盈亏" min-width="110">
              <template #default="{ row }">
                <span :style="row.return_rate >= 0 ? 'color: #f56c6c; font-weight: 600;' : 'color: #67c23a; font-weight: 600;'">
                  {{ row.return_rate >= 0 ? '+' : '' }}{{ (row.cost_basis * row.return_rate).toFixed(2) }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="take_profit_price" label="止盈价" width="90" />
            <el-table-column prop="stop_loss_price" label="止损价" width="90" />
            <el-table-column prop="quantity" label="交易量 (手)" width="100" />
            <el-table-column label="最大回撤" width="90">
              <template #default="{ row }">
                <span style="color: #67c23a;">{{ (row.max_drawdown * 100).toFixed(2) }}%</span>
              </template>
            </el-table-column>
            <el-table-column prop="hold_days" label="持有天数" width="90">
              <template #default="{ row }">
                {{ row.hold_days }} 天
              </template>
            </el-table-column>
          </el-table>
          <div style="display: flex; justify-content: flex-end; margin-top: 12px;">
            <el-pagination
              v-model:current-page="holdingPage"
              :page-size="pageSize"
              :total="holdingTotal"
              layout="total, prev, pager, next"
              @current-change="fetchHoldingPositions"
            />
          </div>
        </el-tab-pane>

        <el-tab-pane label="历史模拟交易记录 (CLOSED)" name="closed">
          <el-table :data="closedPositions" border stripe size="small" style="width: 100%">
            <el-table-column prop="open_date" label="建仓日期" width="100" />
            <el-table-column prop="close_date" label="平仓日期" width="100" />
            <el-table-column prop="contract" label="合约代码" width="100" />
            <el-table-column prop="name" label="合约名称" min-width="120" />
            <el-table-column prop="direction" label="方向" width="95">
              <template #default="{ row }">
                <el-tag :type="row.direction === 'LONG' ? 'danger' : 'primary'" size="small">
                  {{ row.direction === 'LONG' ? '多头 LONG' : '空头 SHORT' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="open_price" label="建仓价" width="90" />
            <el-table-column prop="close_price" label="平仓价" width="90" />
            <el-table-column label="平仓收益率" width="100">
              <template #default="{ row }">
                <span :style="row.return_rate >= 0 ? 'color: #f56c6c; font-weight: 600;' : 'color: #67c23a; font-weight: 600;'">
                  {{ row.return_rate >= 0 ? '+' : '' }}{{ (row.return_rate * 100).toFixed(2) }}%
                </span>
              </template>
            </el-table-column>
            <el-table-column label="实现盈亏" min-width="110">
              <template #default="{ row }">
                <span :style="row.return_rate >= 0 ? 'color: #f56c6c; font-weight: 600;' : 'color: #67c23a; font-weight: 600;'">
                  {{ row.return_rate >= 0 ? '+' : '' }}{{ (row.close_value - row.cost_basis).toFixed(2) }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="hold_days" label="持有天数" width="90">
              <template #default="{ row }">
                {{ row.hold_days }} 天
              </template>
            </el-table-column>
            <el-table-column prop="close_reason" label="平仓原因" width="110">
              <template #default="{ row }">
                <el-tag :type="getReasonTagType(row.close_reason)" size="small">
                  {{ getReasonLabel(row.close_reason) }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
          <div style="display: flex; justify-content: flex-end; margin-top: 12px;">
            <el-pagination
              v-model:current-page="closedPage"
              :page-size="pageSize"
              :total="closedTotal"
              layout="total, prev, pager, next"
              @current-change="fetchClosedPositions"
            />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </FuturesSelectionModuleShell>
</template>

<style scoped>
.card-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}

.chip-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
</style>
