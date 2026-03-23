<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  listStockSelectionCandidates,
  listStockSelectionProfiles,
  listStockSelectionRuns,
  listStockSelectionTemplates
} from "../../api/admin";
import {
  averageStockSelectionMetric,
  formatStockSelectionDateTime,
  formatStockSelectionPercent
} from "../../lib/stock-selection";

const loading = ref(false);
const profiles = ref([]);
const templates = ref([]);
const runs = ref([]);
const selectedRunID = ref("");
const candidateRows = ref([]);

const defaultProfile = computed(() => profiles.value.find((item) => item.is_default) || profiles.value[0] || null);

const defaultTemplate = computed(() => {
  if (defaultProfile.value?.template_id) {
    const matched = templates.value.find((item) => item.id === defaultProfile.value.template_id);
    if (matched) {
      return matched;
    }
  }
  return templates.value.find((item) => item.is_default) || templates.value[0] || null;
});

const templateFactorConfig = computed(() => defaultTemplate.value?.factor_defaults_json || {});
const profileFactorConfig = computed(() => defaultProfile.value?.factor_config || {});

const factorRows = computed(() => [
  {
    label: "量化主分",
    template: formatStockSelectionPercent(templateFactorConfig.value.quant_weight ?? 0.7, 0),
    profile: formatStockSelectionPercent(
      profileFactorConfig.value.quant_weight ?? templateFactorConfig.value.quant_weight ?? 0.75,
      0
    )
  },
  {
    label: "事件确认",
    template: formatStockSelectionPercent(templateFactorConfig.value.event_weight ?? 0.1, 0),
    profile: formatStockSelectionPercent(
      profileFactorConfig.value.event_weight ?? templateFactorConfig.value.event_weight ?? 0.1,
      0
    )
  },
  {
    label: "共振确认",
    template: formatStockSelectionPercent(templateFactorConfig.value.resonance_weight ?? 0.1, 0),
    profile: formatStockSelectionPercent(
      profileFactorConfig.value.resonance_weight ?? templateFactorConfig.value.resonance_weight ?? 0.1,
      0
    )
  },
  {
    label: "流动性/风险修正",
    template: formatStockSelectionPercent(templateFactorConfig.value.liquidity_risk_weight ?? 0.1, 0),
    profile: formatStockSelectionPercent(
      profileFactorConfig.value.liquidity_risk_weight ?? templateFactorConfig.value.liquidity_risk_weight ?? 0.05,
      0
    )
  },
  {
    label: "回看窗口",
    template: `${Number(templateFactorConfig.value.lookback_days || 120)} 天`,
    profile: `${Number(profileFactorConfig.value.lookback_days || templateFactorConfig.value.lookback_days || 120)} 天`
  }
]);

const factorSummary = computed(() => [
  { label: "平均趋势", value: averageStockSelectionMetric(candidateRows.value, (item) => item.factor_breakdown_json?.trend) },
  { label: "平均资金流", value: averageStockSelectionMetric(candidateRows.value, (item) => item.factor_breakdown_json?.money_flow) },
  { label: "平均质量", value: averageStockSelectionMetric(candidateRows.value, (item) => item.factor_breakdown_json?.quality) },
  { label: "平均事件", value: averageStockSelectionMetric(candidateRows.value, (item) => item.factor_breakdown_json?.event) },
  { label: "平均共振", value: averageStockSelectionMetric(candidateRows.value, (item) => item.factor_breakdown_json?.resonance) },
  { label: "平均总分", value: averageStockSelectionMetric(candidateRows.value, (item) => item.factor_breakdown_json?.total_score) }
]);

const selectedRun = computed(() => runs.value.find((item) => item.run_id === selectedRunID.value) || null);

async function fetchBaseData() {
  loading.value = true;
  try {
    const [profileData, templateData, runData] = await Promise.all([
      listStockSelectionProfiles({ page: 1, page_size: 100 }),
      listStockSelectionTemplates({ page: 1, page_size: 100, status: "ACTIVE" }),
      listStockSelectionRuns({ status: "SUCCEEDED", page: 1, page_size: 50 })
    ]);
    profiles.value = Array.isArray(profileData?.items) ? profileData.items : [];
    templates.value = Array.isArray(templateData?.items) ? templateData.items : [];
    runs.value = Array.isArray(runData?.items) ? runData.items : [];
    if (!selectedRunID.value && runs.value.length > 0) {
      selectedRunID.value = runs.value[0].run_id;
    }
  } catch (error) {
    ElMessage.error(error?.message || "加载因子配置失败");
  } finally {
    loading.value = false;
  }
}

async function fetchCandidateRows() {
  if (!selectedRunID.value) {
    candidateRows.value = [];
    return;
  }
  loading.value = true;
  try {
    const data = await listStockSelectionCandidates(selectedRunID.value, { stage: "CANDIDATE_POOL" });
    candidateRows.value = Array.isArray(data?.items) ? data.items : [];
  } catch (error) {
    ElMessage.error(error?.message || "加载因子拆解失败");
    candidateRows.value = [];
  } finally {
    loading.value = false;
  }
}

watch(selectedRunID, fetchCandidateRows);

onMounted(async () => {
  await fetchBaseData();
  await fetchCandidateRows();
});
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股因子与权重"
    description="把默认模板、默认配置和最近成功 run 的候选因子拆解放到一个页面看，方便快速判断当前打分结构是否符合预期。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-select v-model="selectedRunID" placeholder="选择成功运行" style="width: 320px" filterable>
          <el-option
            v-for="run in runs"
            :key="run.run_id"
            :label="`${run.run_id} / ${run.trade_date} / ${run.template_name || run.profile_id}`"
            :value="run.run_id"
          />
        </el-select>
        <el-button :loading="loading" @click="fetchBaseData">刷新</el-button>
      </div>
    </template>

    <div class="factor-grid">
      <div class="card" v-loading="loading">
        <div class="card-title">默认模板因子</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="模板">
            {{ defaultTemplate?.name || "未找到默认模板" }}
          </el-descriptions-item>
          <el-descriptions-item label="市场状态偏好">
            {{ defaultTemplate?.market_regime_bias || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="说明">
            {{ defaultTemplate?.description || "-" }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">默认配置因子</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="配置方案">
            {{ defaultProfile?.name || "未找到默认配置" }}
          </el-descriptions-item>
          <el-descriptions-item label="模板来源">
            {{ defaultProfile?.template_id || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="说明">
            {{ defaultProfile?.description || "-" }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <div class="card" v-loading="loading">
      <div class="card-title">权重对照</div>
      <el-table :data="factorRows" border stripe size="small">
        <el-table-column prop="label" label="因子项" min-width="160" />
        <el-table-column prop="template" label="模板默认" min-width="140" />
        <el-table-column prop="profile" label="当前配置" min-width="140" />
      </el-table>
    </div>

    <div class="card" v-if="selectedRun" v-loading="loading">
      <div class="toolbar" style="justify-content: space-between; flex-wrap: wrap; margin-bottom: 12px">
        <div class="card-title">最近运行因子拆解</div>
        <div class="tag-wrap">
          <el-tag type="info">{{ selectedRun.run_id }}</el-tag>
          <el-tag type="warning">{{ selectedRun.trade_date }}</el-tag>
          <el-tag type="success">{{ selectedRun.template_name || selectedRun.profile_id }}</el-tag>
          <el-tag type="info">完成于 {{ formatStockSelectionDateTime(selectedRun.completed_at) }}</el-tag>
        </div>
      </div>

      <div class="summary-grid">
        <div
          v-for="item in factorSummary"
          :key="item.label"
          class="summary-card"
        >
          <div class="summary-label">{{ item.label }}</div>
          <div class="summary-value">{{ item.value == null ? "-" : Number(item.value).toFixed(2) }}</div>
        </div>
      </div>

      <el-table :data="candidateRows" border stripe size="small" empty-text="当前 run 没有候选因子拆解">
        <el-table-column prop="rank" label="排名" min-width="68" />
        <el-table-column prop="symbol" label="代码" min-width="110" />
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column label="趋势" min-width="90">
          <template #default="{ row }">{{ row.factor_breakdown_json?.trend ?? "-" }}</template>
        </el-table-column>
        <el-table-column label="资金流" min-width="90">
          <template #default="{ row }">{{ row.factor_breakdown_json?.money_flow ?? "-" }}</template>
        </el-table-column>
        <el-table-column label="质量" min-width="90">
          <template #default="{ row }">{{ row.factor_breakdown_json?.quality ?? "-" }}</template>
        </el-table-column>
        <el-table-column label="事件" min-width="90">
          <template #default="{ row }">{{ row.factor_breakdown_json?.event ?? "-" }}</template>
        </el-table-column>
        <el-table-column label="共振" min-width="90">
          <template #default="{ row }">{{ row.factor_breakdown_json?.resonance ?? "-" }}</template>
        </el-table-column>
        <el-table-column label="风险修正" min-width="90">
          <template #default="{ row }">{{ row.factor_breakdown_json?.risk_adjustment ?? "-" }}</template>
        </el-table-column>
        <el-table-column label="总分" min-width="90">
          <template #default="{ row }">{{ row.factor_breakdown_json?.total_score ?? row.quant_score ?? "-" }}</template>
        </el-table-column>
      </el-table>
    </div>
  </StockSelectionModuleShell>
</template>

<style scoped>
.factor-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.card-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}

.tag-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 12px;
}

.summary-card {
  padding: 12px;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.04);
}

.summary-label {
  font-size: 12px;
  color: #64748b;
}

.summary-value {
  margin-top: 6px;
  font-size: 20px;
  font-weight: 700;
}

@media (max-width: 1280px) {
  .factor-grid,
  .summary-grid {
    grid-template-columns: 1fr;
  }
}
</style>
