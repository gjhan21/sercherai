<script setup>
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import FuturesSelectionModuleShell from "../../components/FuturesSelectionModuleShell.vue";
import { createFuturesSelectionRun, getFuturesSelectionOverview } from "../../api/admin";
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

onMounted(fetchOverview);
</script>

<template>
  <FuturesSelectionModuleShell
    title="智能期货总览"
    description="集中查看默认期货配置、最近运行、图谱上下文新鲜度和待审核数量，先确认状态，再决定是否重跑。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchOverview">刷新总览</el-button>
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
