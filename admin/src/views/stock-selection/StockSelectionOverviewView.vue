<script setup>
import { computed, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import { createStockSelectionRun, getStockSelectionOverview } from "../../api/admin";
import {
  formatStockSelectionLabel,
  formatStockSelectionEvaluationStatus,
  formatStockSelectionMarketRegime,
  formatStockSelectionMode,
  formatStockSelectionEvaluationScope,
  formatStockSelectionRunStatus,
  formatStockSelectionSource,
  formatStockSelectionUniverseScope
} from "../../lib/stock-selection";
import { hasPermission } from "../../lib/session";

const router = useRouter();
const canManage = hasPermission("stock_selection.manage");
const loading = ref(false);
const running = ref(false);
const errorMessage = ref("");
const overview = ref({
  default_profile: null,
  latest_trade_date: "",
  latest_run: null,
  latest_success_run: null,
  market_regime: "",
  template_summary: {},
  latest_approved_portfolio: [],
  pending_review_count: 0,
  data_freshness: {},
  evaluation_summary: {},
  evaluation_summary_1_3_5_10_20: {},
  market_analysis: {},
  candidate_pool_summary: {},
  head_summary: {},
  evaluation_backfill_state: {},
  evaluation_split_summary: {},
  warnings: []
});

const evaluationRows = computed(() =>
  Object.entries(
    overview.value?.evaluation_summary_1_3_5_10_20 || overview.value?.evaluation_summary || {}
  )
    .map(([window, item]) => ({
      window,
      ...(item || {})
    }))
    .filter((item) => /^\d+$/.test(String(item.window)))
    .sort((left, right) => Number(left.window) - Number(right.window))
);

const leaderboardRows = computed(() => {
  const items = overview.value?.evaluation_summary_1_3_5_10_20?.leaderboard_items;
  return Array.isArray(items) ? items : [];
});

const splitEvaluationRows = computed(() =>
  Object.entries(overview.value?.evaluation_split_summary || {}).map(([scope, item]) => ({
    scope,
    ...(item || {})
  }))
);

function formatDateTime(value) {
  const timestamp = Date.parse(value || "");
  if (Number.isNaN(timestamp)) {
    return "-";
  }
  return new Date(timestamp).toLocaleString("zh-CN", {
    hour12: false
  });
}

function formatPercent(value) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return "-";
  }
  return `${(numeric * 100).toFixed(2)}%`;
}

function resolveEvaluationMetric(row, primaryKey, legacyKey) {
  const primary = Number(row?.[primaryKey]);
  if (Number.isFinite(primary)) {
    return primary;
  }
  if (legacyKey) {
    const legacy = Number(row?.[legacyKey]);
    if (Number.isFinite(legacy)) {
      return legacy;
    }
  }
  return NaN;
}

function runTagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "SUCCEEDED") return "success";
  if (normalized === "FAILED") return "danger";
  if (normalized === "RUNNING") return "warning";
  return "info";
}

function evaluationBackfillTagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "READY") return "success";
  if (normalized === "PARTIAL") return "warning";
  if (normalized === "FAILED") return "danger";
  return "info";
}

const evaluationBackfillSummary = computed(() => {
  const state = overview.value?.evaluation_backfill_state || {};
  const readyCount = Number(state.ready_count || 0);
  const targetCount = Number(state.target_count || 0);
  const missingHorizons = Array.isArray(state.missing_horizons) ? state.missing_horizons : [];
  return {
    status: state.status || "PENDING",
    message: state.message || "-",
    progress: targetCount > 0 ? `${readyCount}/${targetCount}` : "-",
    missingText: missingHorizons.length ? missingHorizons.join(" / ") : "-",
    lastUpdatedAt: state.last_updated_at || ""
  };
});

async function fetchOverview() {
  loading.value = true;
  errorMessage.value = "";
  try {
    overview.value = await getStockSelectionOverview();
  } catch (error) {
    errorMessage.value = error?.message || "加载智能选股总览失败";
  } finally {
    loading.value = false;
  }
}

async function handleRunNow() {
  if (!overview.value?.default_profile?.id) {
    ElMessage.error("当前没有可用的默认配置方案");
    return;
  }
  running.value = true;
  try {
    const tradeDate =
      overview.value.latest_trade_date || new Date().toISOString().slice(0, 10);
    const run = await createStockSelectionRun({
      trade_date: tradeDate,
      profile_id: overview.value.default_profile.id
    });
    ElMessage.success(`已完成运行 ${run.run_id}`);
    router.push({
      name: "stock-selection-runs",
      query: { run_id: run.run_id }
    });
  } catch (error) {
    ElMessage.error(error?.message || "触发运行失败");
  } finally {
    running.value = false;
  }
}

onMounted(fetchOverview);
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股总览"
    description="智能选股是您的‘科研实验室’与‘生产线’，负责算法跑批与候选集生成。生成的结果将同步至策略中心进行最终审核与发布。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchOverview">刷新总览</el-button>
        <el-button type="success" plain @click="router.push({ path: '/market-center' })">
          前往策略中心 (运营台)
        </el-button>
        <el-button
          v-if="canManage"
          type="primary"
          :loading="running"
          @click="handleRunNow"
        >
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
        <div class="card-title">默认配置方案</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="名称">
            {{ overview.default_profile?.name || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="版本">
            v{{ overview.default_profile?.current_version || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="模式">
            {{ formatStockSelectionMode(overview.default_profile?.selection_mode_default) }}
          </el-descriptions-item>
          <el-descriptions-item label="股票池范围">
            {{ formatStockSelectionUniverseScope(overview.default_profile?.universe_scope) }}
          </el-descriptions-item>
          <el-descriptions-item label="默认模板">
            {{ overview.template_summary?.default_template_name || "-" }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">大盘结论区</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="趋势状态">
            {{ formatStockSelectionMarketRegime(overview.market_analysis?.trend_state || overview.market_regime) }}
          </el-descriptions-item>
          <el-descriptions-item label="次日节奏">
            {{ formatStockSelectionLabel(overview.market_analysis?.next_day_rhythm) }}
          </el-descriptions-item>
          <el-descriptions-item label="风控姿态">
            {{ overview.market_analysis?.risk_posture || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="主推荐数量">
            {{ overview.head_summary?.short_term_primary_count || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="辅助推荐数量">
            {{ overview.head_summary?.swing_auxiliary_count || 0 }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">待选池压缩摘要</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="粗筛数量">
            {{ overview.candidate_pool_summary?.coarse_count || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="聚焦数量">
            {{ overview.candidate_pool_summary?.focus_count || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="观察数量">
            {{ overview.candidate_pool_summary?.watch_count || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="上下文新鲜度">
            {{ overview.data_freshness?.selected_trade_date || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="行情来源">
            {{ formatStockSelectionSource(overview.data_freshness?.price_source) }}
          </el-descriptions-item>
          <el-descriptions-item label="待审核数量">
            {{ overview.pending_review_count || 0 }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <div class="card" v-loading="loading">
      <div class="card-title">最近运行</div>
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="运行编号">
          {{ overview.latest_run?.run_id || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="runTagType(overview.latest_run?.status)">
            {{ formatStockSelectionRunStatus(overview.latest_run?.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="交易日">
          {{ overview.latest_run?.trade_date || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="市场状态">
          {{ formatStockSelectionMarketRegime(overview.market_regime) }}
        </el-descriptions-item>
      </el-descriptions>
    </div>

    <div class="card" v-loading="loading">
      <div class="card-title">评估回填状态</div>
      <el-descriptions :column="1" border size="small">
        <el-descriptions-item label="状态">
          <el-tag :type="evaluationBackfillTagType(evaluationBackfillSummary.status)">
            {{ formatStockSelectionEvaluationStatus(evaluationBackfillSummary.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="完成度">
          {{ evaluationBackfillSummary.progress }}
        </el-descriptions-item>
        <el-descriptions-item label="缺失 Horizon">
          {{ evaluationBackfillSummary.missingText }}
        </el-descriptions-item>
        <el-descriptions-item label="说明">
          {{ evaluationBackfillSummary.message }}
        </el-descriptions-item>
        <el-descriptions-item label="最近回填时间">
          {{ formatDateTime(evaluationBackfillSummary.lastUpdatedAt) }}
        </el-descriptions-item>
      </el-descriptions>
    </div>

    <div class="card" v-if="Array.isArray(overview.latest_approved_portfolio) && overview.latest_approved_portfolio.length">
      <div class="card-title">最近已发布组合</div>
      <el-tag
        v-for="item in overview.latest_approved_portfolio"
        :key="`${item.symbol}-${item.rank}`"
        type="success"
        style="margin-right: 8px; margin-bottom: 8px"
      >
        {{ item.rank }}. {{ item.symbol }} / {{ item.name }} / {{ formatStockSelectionLabel(item.portfolio_role || "PORTFOLIO") }}
      </el-tag>
    </div>

    <div class="card" v-loading="loading">
      <div class="card-title">1 / 3 / 5 / 10 / 20 日研究评估摘要</div>
      <el-table :data="evaluationRows" border stripe size="small" empty-text="暂无评估摘要">
        <el-table-column prop="window" label="窗口" min-width="90">
          <template #default="{ row }">{{ row.window }} 日</template>
        </el-table-column>
        <el-table-column prop="sample_count" label="样本数" min-width="90" />
        <el-table-column label="平均收益" min-width="120">
          <template #default="{ row }">
            {{ formatPercent(resolveEvaluationMetric(row, "avg_return_pct", `avg_return_${row.window}`)) }}
          </template>
        </el-table-column>
        <el-table-column label="平均超额" min-width="120">
          <template #default="{ row }">
            {{ formatPercent(resolveEvaluationMetric(row, "avg_excess_return_pct", "")) }}
          </template>
        </el-table-column>
        <el-table-column label="命中率" min-width="120">
          <template #default="{ row }">
            {{ formatPercent(resolveEvaluationMetric(row, "hit_rate", `hit_rate_${row.window}`)) }}
          </template>
        </el-table-column>
        <el-table-column label="平均回撤" min-width="120">
          <template #default="{ row }">
            {{ formatPercent(resolveEvaluationMetric(row, "avg_max_drawdown_pct", `max_drawdown_${row.window}`)) }}
          </template>
        </el-table-column>
        <el-table-column prop="worst_max_drawdown_pct" label="最差回撤" min-width="120">
          <template #default="{ row }">{{ formatPercent(row.worst_max_drawdown_pct) }}</template>
        </el-table-column>
        <el-table-column prop="generated_at" label="生成时间" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.generated_at) }}</template>
        </el-table-column>
      </el-table>
    </div>

    <div class="card" v-loading="loading">
      <div class="card-title">分头评估摘要</div>
      <el-table :data="splitEvaluationRows" border stripe size="small" empty-text="暂无分头评估摘要">
        <el-table-column prop="scope" label="推荐头" min-width="150">
          <template #default="{ row }">{{ formatStockSelectionEvaluationScope(row.scope) }}</template>
        </el-table-column>
        <el-table-column prop="sample_count" label="样本数" min-width="100" />
        <el-table-column label="平均收益" min-width="120">
          <template #default="{ row }">{{ formatPercent(row.avg_return_pct) }}</template>
        </el-table-column>
        <el-table-column label="平均超额" min-width="120">
          <template #default="{ row }">{{ formatPercent(row.avg_excess_return_pct) }}</template>
        </el-table-column>
        <el-table-column label="命中率" min-width="110">
          <template #default="{ row }">{{ formatPercent(row.hit_rate) }}</template>
        </el-table-column>
        <el-table-column label="平均回撤" min-width="120">
          <template #default="{ row }">{{ formatPercent(row.avg_max_drawdown_pct) }}</template>
        </el-table-column>
      </el-table>
    </div>

    <div class="card" v-loading="loading">
      <div class="card-title">研究评估榜单摘要</div>
      <el-table :data="leaderboardRows" border stripe size="small" empty-text="暂无榜单数据">
        <el-table-column prop="template_name" label="模板" min-width="140" />
        <el-table-column prop="profile_name" label="配置方案" min-width="160" />
        <el-table-column prop="market_regime" label="市场状态" min-width="120">
          <template #default="{ row }">
            {{ formatStockSelectionMarketRegime(row.market_regime) }}
          </template>
        </el-table-column>
        <el-table-column prop="sample_count" label="样本数" min-width="90" />
        <el-table-column label="1日收益" min-width="110">
          <template #default="{ row }">{{ formatPercent(row.return_by_horizon?.["1"]) }}</template>
        </el-table-column>
        <el-table-column label="5日收益" min-width="110">
          <template #default="{ row }">{{ formatPercent(row.return_by_horizon?.["5"]) }}</template>
        </el-table-column>
        <el-table-column label="10日收益" min-width="110">
          <template #default="{ row }">{{ formatPercent(row.return_by_horizon?.["10"]) }}</template>
        </el-table-column>
        <el-table-column label="5日命中率" min-width="110">
          <template #default="{ row }">{{ formatPercent(row.hit_rate_by_horizon?.["5"]) }}</template>
        </el-table-column>
        <el-table-column label="10日命中率" min-width="120">
          <template #default="{ row }">{{ formatPercent(row.hit_rate_by_horizon?.["10"]) }}</template>
        </el-table-column>
      </el-table>
    </div>

    <div class="card" v-if="Array.isArray(overview.warnings) && overview.warnings.length">
      <div class="card-title">运行提醒</div>
      <el-tag
        v-for="warning in overview.warnings"
        :key="warning"
        type="warning"
        style="margin-right: 8px; margin-bottom: 8px"
      >
        {{ warning }}
      </el-tag>
    </div>
  </StockSelectionModuleShell>
</template>

<style scoped>
.stock-overview-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 12px;
  margin-bottom: 12px;
}

.card-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}
</style>
