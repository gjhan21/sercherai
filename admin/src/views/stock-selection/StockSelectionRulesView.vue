<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  listStockSelectionProfiles,
  listStockSelectionRuns,
  listStockSelectionTemplates
} from "../../api/admin";
import {
  formatStockSelectionDateTime,
  formatStockSelectionMode,
  formatStockSelectionUniverseScope
} from "../../lib/stock-selection";

const loading = ref(false);
const profiles = ref([]);
const templates = ref([]);
const runs = ref([]);
const selectedRunID = ref("");

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

const selectedRun = computed(() => runs.value.find((item) => item.run_id === selectedRunID.value) || runs.value[0] || null);
const profileUniverse = computed(() => defaultProfile.value?.universe_config || {});
const templateUniverse = computed(() => defaultTemplate.value?.universe_defaults_json || {});
const runFilters = computed(() => selectedRun.value?.context_meta?.universe_filters || {});

const comparisonRows = computed(() => [
  {
    label: "股票池范围",
    template: formatStockSelectionUniverseScope(templateUniverse.value.universe_scope || defaultTemplate.value?.universe_scope),
    profile: formatStockSelectionUniverseScope(defaultProfile.value?.universe_scope || profileUniverse.value.universe_scope),
    runtime: formatStockSelectionUniverseScope(selectedRun.value?.universe_scope || profileUniverse.value.universe_scope)
  },
  {
    label: "默认模式",
    template: "-",
    profile: formatStockSelectionMode(defaultProfile.value?.selection_mode_default),
    runtime: formatStockSelectionMode(selectedRun.value?.selection_mode)
  },
  {
    label: "最少上市天数",
    template: valueOrDash(templateUniverse.value.min_listing_days),
    profile: valueOrDash(profileUniverse.value.min_listing_days),
    runtime: valueOrDash(runFilters.value.min_listing_days)
  },
  {
    label: "20日均成交额下限",
    template: valueOrDash(templateUniverse.value.min_avg_turnover),
    profile: valueOrDash(profileUniverse.value.min_avg_turnover),
    runtime: valueOrDash(runFilters.value.min_avg_turnover)
  },
  {
    label: "价格区间",
    template: rangeText(templateUniverse.value.price_min, templateUniverse.value.price_max),
    profile: rangeText(profileUniverse.value.price_min, profileUniverse.value.price_max),
    runtime: rangeText(runFilters.value.price_min, runFilters.value.price_max)
  },
  {
    label: "波动率区间",
    template: rangeText(templateUniverse.value.volatility_min, templateUniverse.value.volatility_max),
    profile: rangeText(profileUniverse.value.volatility_min, profileUniverse.value.volatility_max),
    runtime: rangeText(runFilters.value.volatility_min, runFilters.value.volatility_max)
  },
  {
    label: "剔除 ST",
    template: booleanText(templateUniverse.value.exclude_st),
    profile: booleanText(profileUniverse.value.exclude_st),
    runtime: booleanText(profileUniverse.value.exclude_st)
  },
  {
    label: "剔除停牌/零成交",
    template: booleanText(templateUniverse.value.exclude_suspended),
    profile: booleanText(profileUniverse.value.exclude_suspended),
    runtime: booleanText(profileUniverse.value.exclude_suspended)
  }
]);

const restrictionGroups = computed(() => [
  {
    label: "行业白名单",
    template: toTags(templateUniverse.value.industry_whitelist),
    profile: toTags(profileUniverse.value.industry_whitelist)
  },
  {
    label: "行业黑名单",
    template: toTags(templateUniverse.value.industry_blacklist),
    profile: toTags(profileUniverse.value.industry_blacklist)
  },
  {
    label: "板块白名单",
    template: toTags(templateUniverse.value.sector_whitelist),
    profile: toTags(profileUniverse.value.sector_whitelist)
  },
  {
    label: "板块黑名单",
    template: toTags(templateUniverse.value.sector_blacklist),
    profile: toTags(profileUniverse.value.sector_blacklist)
  },
  {
    label: "题材白名单",
    template: toTags(templateUniverse.value.theme_whitelist),
    profile: toTags(profileUniverse.value.theme_whitelist)
  },
  {
    label: "题材黑名单",
    template: toTags(templateUniverse.value.theme_blacklist),
    profile: toTags(profileUniverse.value.theme_blacklist)
  },
  {
    label: "显式排除股票",
    template: toTags(templateUniverse.value.excluded_symbols),
    profile: toTags(profileUniverse.value.excluded_symbols)
  }
]);

function valueOrDash(value) {
  return value === undefined || value === null || value === "" ? "-" : String(value);
}

function rangeText(minValue, maxValue) {
  const left = valueOrDash(minValue);
  const right = valueOrDash(maxValue);
  return left === "-" && right === "-" ? "-" : `${left} ~ ${right}`;
}

function booleanText(value) {
  if (value === undefined || value === null || value === "") {
    return "-";
  }
  return value ? "启用" : "关闭";
}

function toTags(value) {
  return (Array.isArray(value) ? value : []).map((item) => String(item || "").trim()).filter(Boolean);
}

async function fetchData() {
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
    ElMessage.error(error?.message || "加载股票池规则失败");
  } finally {
    loading.value = false;
  }
}

watch(
  runs,
  (items) => {
    if (selectedRunID.value && items.some((item) => item.run_id === selectedRunID.value)) {
      return;
    }
    selectedRunID.value = items[0]?.run_id || "";
  },
  { deep: true }
);

onMounted(fetchData);
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股股票池规则"
    description="把默认模板、默认配置和最近成功运行实际生效的过滤条件拆开来看，方便确认股票池边界是否按预期执行。"
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
        <el-button :loading="loading" @click="fetchData">刷新</el-button>
      </div>
    </template>

    <div class="overview-grid">
      <div class="card" v-loading="loading">
        <div class="card-title">默认模板</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="模板名称">
            {{ defaultTemplate?.name || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="模板说明">
            {{ defaultTemplate?.description || "-" }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">默认配置</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="配置名称">
            {{ defaultProfile?.name || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="运行模式">
            {{ formatStockSelectionMode(defaultProfile?.selection_mode_default) }}
          </el-descriptions-item>
          <el-descriptions-item label="说明">
            {{ defaultProfile?.description || "-" }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">最近运行</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="运行编号">
            {{ selectedRun?.run_id || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="交易日">
            {{ selectedRun?.trade_date || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="完成时间">
            {{ formatStockSelectionDateTime(selectedRun?.completed_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="结果流转">
            {{ selectedRun?.universe_count || 0 }} -> {{ selectedRun?.seed_count || 0 }} -> {{ selectedRun?.candidate_count || 0 }} -> {{ selectedRun?.selected_count || 0 }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <div class="card" v-loading="loading">
      <div class="card-title">规则对照</div>
      <el-table :data="comparisonRows" border stripe size="small">
        <el-table-column prop="label" label="规则项" min-width="160" />
        <el-table-column prop="template" label="模板默认" min-width="160" />
        <el-table-column prop="profile" label="当前配置" min-width="160" />
        <el-table-column prop="runtime" label="最近运行生效值" min-width="180" />
      </el-table>
    </div>

    <div class="restriction-grid">
      <div class="card" v-loading="loading">
        <div class="card-title">模板/配置限制</div>
        <div
          v-for="group in restrictionGroups"
          :key="group.label"
          class="restriction-group"
        >
          <div class="restriction-label">{{ group.label }}</div>
          <div class="restriction-row">
            <span class="restriction-side">模板</span>
            <div class="tag-wrap">
              <el-tag v-for="item in group.template" :key="`${group.label}-template-${item}`" type="info">{{ item }}</el-tag>
              <span v-if="group.template.length === 0" class="muted">未设置</span>
            </div>
          </div>
          <div class="restriction-row">
            <span class="restriction-side">配置</span>
            <div class="tag-wrap">
              <el-tag v-for="item in group.profile" :key="`${group.label}-profile-${item}`" type="success">{{ item }}</el-tag>
              <span v-if="group.profile.length === 0" class="muted">未设置</span>
            </div>
          </div>
        </div>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">最近运行回放</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="实际交易日">
            {{ selectedRun?.context_meta?.selected_trade_date || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="上市天数近似过滤">
            {{ selectedRun?.context_meta?.listing_days_filter_applied === false ? "已跳过" : "已启用" }}
          </el-descriptions-item>
          <el-descriptions-item label="价格区间">
            {{ rangeText(runFilters.price_min, runFilters.price_max) }}
          </el-descriptions-item>
          <el-descriptions-item label="波动率区间">
            {{ rangeText(runFilters.volatility_min, runFilters.volatility_max) }}
          </el-descriptions-item>
          <el-descriptions-item label="最少上市天数">
            {{ valueOrDash(runFilters.min_listing_days) }}
          </el-descriptions-item>
          <el-descriptions-item label="20日均成交额下限">
            {{ valueOrDash(runFilters.min_avg_turnover) }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="sub-card" v-if="Array.isArray(selectedRun?.warning_messages) && selectedRun.warning_messages.length">
          <div class="sub-card__title">运行提醒</div>
          <div class="tag-wrap">
            <el-tag v-for="warning in selectedRun.warning_messages" :key="warning" type="warning">
              {{ warning }}
            </el-tag>
          </div>
        </div>
      </div>
    </div>
  </StockSelectionModuleShell>
</template>

<style scoped>
.overview-grid,
.restriction-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.restriction-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.card-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}

.restriction-group + .restriction-group {
  margin-top: 14px;
  padding-top: 14px;
  border-top: 1px solid rgba(148, 163, 184, 0.2);
}

.restriction-label {
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
}

.restriction-row {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  margin-bottom: 8px;
}

.restriction-side {
  width: 44px;
  flex: none;
  color: #64748b;
}

.tag-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.sub-card {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(148, 163, 184, 0.2);
}

.sub-card__title {
  margin-bottom: 8px;
  font-weight: 600;
}

@media (max-width: 1280px) {
  .overview-grid,
  .restriction-grid {
    grid-template-columns: 1fr;
  }
}
</style>
