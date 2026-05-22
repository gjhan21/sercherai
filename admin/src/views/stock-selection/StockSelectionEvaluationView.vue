<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  getStockSelectionOverview,
  listStockSelectionEvaluationLeaderboard,
  listStockSelectionProfiles,
  listStockSelectionTemplates
} from "../../api/admin";
import {
  averageStockSelectionMetric,
  formatStockSelectionEvaluationStatus,
  formatStockSelectionEvaluationScope,
  formatStockSelectionMarketRegime,
  formatStockSelectionPercent,
  stockSelectionMarketRegimeOptions
} from "../../lib/stock-selection";

const loading = ref(false);
const leaderboard = ref([]);
const profiles = ref([]);
const templates = ref([]);
const evaluationBackfillState = ref({});

const filters = reactive({
  template_id: "",
  profile_id: "",
  market_regime: "",
  evaluation_scope: ""
});

const summaryCards = computed(() => {
  const items = Array.isArray(leaderboard.value) ? leaderboard.value : [];
  const bestByFiveDay = [...items]
    .filter((item) => Number.isFinite(Number(item?.return_by_horizon?.["5"])))
    .sort((left, right) => Number(right?.return_by_horizon?.["5"] || -Infinity) - Number(left?.return_by_horizon?.["5"] || -Infinity))[0];

  return [
    {
      label: "榜单行数",
      value: String(items.length)
    },
    {
      label: "平均 5 日收益",
      value: formatStockSelectionPercent(averageStockSelectionMetric(items, (item) => item?.return_by_horizon?.["5"]))
    },
    {
      label: "平均 10 日命中率",
      value: formatStockSelectionPercent(averageStockSelectionMetric(items, (item) => item?.hit_rate_by_horizon?.["10"]))
    },
    {
      label: "当前最佳 5 日模板",
      value: bestByFiveDay ? `${bestByFiveDay.template_name || bestByFiveDay.profile_name || "-"} / ${formatStockSelectionPercent(bestByFiveDay?.return_by_horizon?.["5"])}` : "-"
    }
  ];
});

async function fetchMeta() {
  const [templateResp, profileResp, overviewResp] = await Promise.all([
    listStockSelectionTemplates({ page: 1, page_size: 100 }),
    listStockSelectionProfiles({ page: 1, page_size: 100 }),
    getStockSelectionOverview()
  ]);
  templates.value = Array.isArray(templateResp?.items) ? templateResp.items : [];
  profiles.value = Array.isArray(profileResp?.items) ? profileResp.items : [];
  evaluationBackfillState.value =
    overviewResp?.evaluation_backfill_state && typeof overviewResp.evaluation_backfill_state === "object"
      ? overviewResp.evaluation_backfill_state
      : {};
}

function evaluationTagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "READY") return "success";
  if (normalized === "PARTIAL") return "warning";
  if (normalized === "FAILED") return "danger";
  return "info";
}

async function fetchLeaderboard() {
  loading.value = true;
  try {
    const data = await listStockSelectionEvaluationLeaderboard(filters);
    leaderboard.value = Array.isArray(data) ? data : Array.isArray(data?.items) ? data.items : [];
  } catch (error) {
    ElMessage.error(error?.message || "加载评估复盘失败");
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  await fetchMeta();
  await fetchLeaderboard();
});
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股评估复盘"
    description="这里集中查看模板、配置方案和市场状态下的收益、命中率与回撤表现。V2 当前先展示榜单聚合，后续再继续扩展最佳/最差样本与失效案例。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchLeaderboard">刷新复盘</el-button>
      </div>
    </template>

    <div class="card">
      <div class="summary-grid" style="margin-bottom: 12px">
        <div v-for="card in summaryCards" :key="card.label" class="summary-card">
          <div class="summary-label">{{ card.label }}</div>
          <div class="summary-value">{{ card.value }}</div>
        </div>
      </div>

      <el-descriptions :column="4" border size="small" style="margin-bottom: 12px">
        <el-descriptions-item label="评估回填状态">
          <el-tag :type="evaluationTagType(evaluationBackfillState.status)">
            {{ formatStockSelectionEvaluationStatus(evaluationBackfillState.status || "PENDING") }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="完成度">
          {{ Number(evaluationBackfillState.ready_count || 0) }}/{{ Number(evaluationBackfillState.target_count || 0) }}
        </el-descriptions-item>
        <el-descriptions-item label="缺失 Horizon">
          {{ Array.isArray(evaluationBackfillState.missing_horizons) && evaluationBackfillState.missing_horizons.length ? evaluationBackfillState.missing_horizons.join(" / ") : "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="说明">
          {{ evaluationBackfillState.message || "-" }}
        </el-descriptions-item>
      </el-descriptions>

      <div class="toolbar" style="flex-wrap: wrap">
        <el-select v-model="filters.template_id" clearable placeholder="策略模板" style="width: 220px">
          <el-option
            v-for="item in templates"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>
        <el-select v-model="filters.profile_id" clearable placeholder="策略配置" style="width: 220px">
          <el-option
            v-for="item in profiles"
            :key="item.id"
            :label="item.name"
            :value="item.id"
          />
        </el-select>
        <el-select v-model="filters.market_regime" clearable placeholder="市场状态" style="width: 180px">
          <el-option
            v-for="item in stockSelectionMarketRegimeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-select v-model="filters.evaluation_scope" clearable placeholder="推荐头" style="width: 180px">
          <el-option label="超短线主推荐" value="SHORT_TERM_PRIMARY" />
          <el-option label="短波段辅助" value="SWING_AUXILIARY" />
          <el-option label="候选池" value="CANDIDATE_POOL" />
        </el-select>
        <el-button type="primary" @click="fetchLeaderboard">查询</el-button>
      </div>

      <el-table :data="leaderboard" border stripe size="small" v-loading="loading" empty-text="暂无评估数据">
        <el-table-column prop="template_name" label="模板" min-width="140" />
        <el-table-column prop="profile_name" label="配置方案" min-width="160" />
        <el-table-column prop="evaluation_scope" label="推荐头" min-width="140">
          <template #default="{ row }">{{ formatStockSelectionEvaluationScope(row.evaluation_scope) }}</template>
        </el-table-column>
        <el-table-column prop="market_regime" label="市场状态" min-width="120">
          <template #default="{ row }">{{ formatStockSelectionMarketRegime(row.market_regime) }}</template>
        </el-table-column>
        <el-table-column prop="sample_count" label="样本数" min-width="90" />
        <el-table-column label="1日收益" min-width="110">
          <template #default="{ row }">{{ formatStockSelectionPercent(row.return_by_horizon?.["1"]) }}</template>
        </el-table-column>
        <el-table-column label="3日收益" min-width="110">
          <template #default="{ row }">{{ formatStockSelectionPercent(row.return_by_horizon?.["3"]) }}</template>
        </el-table-column>
        <el-table-column label="5日收益" min-width="110">
          <template #default="{ row }">{{ formatStockSelectionPercent(row.return_by_horizon?.["5"]) }}</template>
        </el-table-column>
        <el-table-column label="10日收益" min-width="110">
          <template #default="{ row }">{{ formatStockSelectionPercent(row.return_by_horizon?.["10"]) }}</template>
        </el-table-column>
        <el-table-column label="20日收益" min-width="110">
          <template #default="{ row }">{{ formatStockSelectionPercent(row.return_by_horizon?.["20"]) }}</template>
        </el-table-column>
        <el-table-column label="5日命中率" min-width="110">
          <template #default="{ row }">{{ formatStockSelectionPercent(row.hit_rate_by_horizon?.["5"]) }}</template>
        </el-table-column>
        <el-table-column label="10日命中率" min-width="110">
          <template #default="{ row }">{{ formatStockSelectionPercent(row.hit_rate_by_horizon?.["10"]) }}</template>
        </el-table-column>
        <el-table-column prop="max_drawdown_pct" label="最大回撤" min-width="110">
          <template #default="{ row }">{{ formatStockSelectionPercent(row.max_drawdown_pct) }}</template>
        </el-table-column>
      </el-table>
    </div>
  </StockSelectionModuleShell>
</template>

<style scoped>
.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
}

.summary-card {
  border: 1px solid var(--el-border-color-light);
  border-radius: 12px;
  padding: 14px 16px;
  background: var(--el-fill-color-blank);
}

.summary-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.summary-value {
  margin-top: 8px;
  font-size: 18px;
  font-weight: 700;
  color: var(--el-text-color-primary);
  line-height: 1.4;
}
</style>
