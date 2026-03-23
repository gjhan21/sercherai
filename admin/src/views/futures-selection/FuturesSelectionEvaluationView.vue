<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import FuturesSelectionModuleShell from "../../components/FuturesSelectionModuleShell.vue";
import {
  listFuturesSelectionEvaluationLeaderboard,
  listFuturesSelectionProfiles,
  listFuturesSelectionTemplates
} from "../../api/admin";
import {
  averageFuturesSelectionMetric,
  formatFuturesSelectionMarketRegime,
  formatFuturesSelectionPercent,
  futuresSelectionMarketRegimeOptions
} from "../../lib/futures-selection";

const loading = ref(false);
const leaderboard = ref([]);
const profiles = ref([]);
const templates = ref([]);

const filters = reactive({
  template_id: "",
  profile_id: "",
  market_regime: ""
});

const summaryCards = computed(() => {
  const items = Array.isArray(leaderboard.value) ? leaderboard.value : [];
  const bestFiveDay = [...items]
    .filter((item) => Number.isFinite(Number(item?.return_by_horizon?.["5"])))
    .sort((left, right) => Number(right?.return_by_horizon?.["5"] || -Infinity) - Number(left?.return_by_horizon?.["5"] || -Infinity))[0];

  return [
    {
      label: "榜单行数",
      value: String(items.length)
    },
    {
      label: "平均 5 日收益",
      value: formatFuturesSelectionPercent(averageFuturesSelectionMetric(items, (item) => item?.return_by_horizon?.["5"]))
    },
    {
      label: "平均 10 日命中率",
      value: formatFuturesSelectionPercent(averageFuturesSelectionMetric(items, (item) => item?.hit_rate_by_horizon?.["10"]))
    },
    {
      label: "当前最佳 5 日配置",
      value: bestFiveDay
        ? `${bestFiveDay.template_name || bestFiveDay.profile_name || "-"} / ${formatFuturesSelectionPercent(bestFiveDay?.return_by_horizon?.["5"])}`
        : "-"
    }
  ];
});

async function fetchMeta() {
  const [templateResp, profileResp] = await Promise.all([
    listFuturesSelectionTemplates({ page: 1, page_size: 100, status: "ACTIVE" }),
    listFuturesSelectionProfiles({ page: 1, page_size: 100 })
  ]);
  templates.value = Array.isArray(templateResp?.items) ? templateResp.items : [];
  profiles.value = Array.isArray(profileResp?.items) ? profileResp.items : [];
}

async function fetchLeaderboard() {
  loading.value = true;
  try {
    const data = await listFuturesSelectionEvaluationLeaderboard(filters);
    leaderboard.value = Array.isArray(data) ? data : Array.isArray(data?.items) ? data.items : [];
  } catch (error) {
    ElMessage.error(error?.message || "加载期货评估复盘失败");
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
  <FuturesSelectionModuleShell
    title="智能期货评估复盘"
    description="这里集中查看期货模板、配置方案和市场状态下的收益、命中率与回撤表现。当前先收口成榜单视图，方便快速比较。"
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
            v-for="item in futuresSelectionMarketRegimeOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-button type="primary" @click="fetchLeaderboard">查询</el-button>
      </div>

      <el-table :data="leaderboard" border stripe size="small" v-loading="loading" empty-text="暂无评估数据">
        <el-table-column prop="template_name" label="模板" min-width="140" />
        <el-table-column prop="profile_name" label="配置方案" min-width="160" />
        <el-table-column prop="market_regime" label="市场状态" min-width="120">
          <template #default="{ row }">{{ formatFuturesSelectionMarketRegime(row.market_regime) }}</template>
        </el-table-column>
        <el-table-column prop="sample_count" label="样本数" min-width="90" />
        <el-table-column label="1日收益" min-width="110">
          <template #default="{ row }">{{ formatFuturesSelectionPercent(row.return_by_horizon?.["1"]) }}</template>
        </el-table-column>
        <el-table-column label="3日收益" min-width="110">
          <template #default="{ row }">{{ formatFuturesSelectionPercent(row.return_by_horizon?.["3"]) }}</template>
        </el-table-column>
        <el-table-column label="5日收益" min-width="110">
          <template #default="{ row }">{{ formatFuturesSelectionPercent(row.return_by_horizon?.["5"]) }}</template>
        </el-table-column>
        <el-table-column label="10日收益" min-width="110">
          <template #default="{ row }">{{ formatFuturesSelectionPercent(row.return_by_horizon?.["10"]) }}</template>
        </el-table-column>
        <el-table-column label="20日收益" min-width="110">
          <template #default="{ row }">{{ formatFuturesSelectionPercent(row.return_by_horizon?.["20"]) }}</template>
        </el-table-column>
        <el-table-column label="5日命中率" min-width="110">
          <template #default="{ row }">{{ formatFuturesSelectionPercent(row.hit_rate_by_horizon?.["5"]) }}</template>
        </el-table-column>
        <el-table-column label="10日命中率" min-width="110">
          <template #default="{ row }">{{ formatFuturesSelectionPercent(row.hit_rate_by_horizon?.["10"]) }}</template>
        </el-table-column>
        <el-table-column label="最大回撤" min-width="110">
          <template #default="{ row }">{{ formatFuturesSelectionPercent(row.max_drawdown_pct) }}</template>
        </el-table-column>
      </el-table>
    </div>
  </FuturesSelectionModuleShell>
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
