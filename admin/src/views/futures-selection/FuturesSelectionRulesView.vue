<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import FuturesSelectionModuleShell from "../../components/FuturesSelectionModuleShell.vue";
import {
  listFuturesSelectionProfiles,
  listFuturesSelectionRuns,
  listFuturesSelectionTemplates
} from "../../api/admin";
import {
  formatFuturesSelectionContractScope,
  formatFuturesSelectionDateTime,
  formatFuturesSelectionStyle
} from "../../lib/futures-selection";

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

const comparisonRows = computed(() => [
  {
    label: "研究风格",
    template: formatFuturesSelectionStyle(templateUniverse.value.style),
    profile: formatFuturesSelectionStyle(defaultProfile.value?.style_default || profileUniverse.value.style),
    runtime: formatFuturesSelectionStyle(selectedRun.value?.style)
  },
  {
    label: "合约范围",
    template: formatFuturesSelectionContractScope(templateUniverse.value.contract_scope),
    profile: formatFuturesSelectionContractScope(defaultProfile.value?.contract_scope || profileUniverse.value.contract_scope),
    runtime: formatFuturesSelectionContractScope(selectedRun.value?.contract_scope)
  },
  {
    label: "短历史兜底",
    template: booleanText(templateUniverse.value.allow_mock_fallback_on_short_history),
    profile: booleanText(profileUniverse.value.allow_mock_fallback_on_short_history),
    runtime: booleanText(profileUniverse.value.allow_mock_fallback_on_short_history)
  },
  {
    label: "最近运行组合数",
    template: "-",
    profile: "-",
    runtime: valueOrDash(selectedRun.value?.selected_count)
  }
]);

const contractGroups = computed(() => [
  {
    label: "模板默认合约",
    items: toTags(templateUniverse.value.contracts)
  },
  {
    label: "当前配置合约",
    items: toTags(profileUniverse.value.contracts)
  }
]);

function valueOrDash(value) {
  return value === undefined || value === null || value === "" ? "-" : String(value);
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
      listFuturesSelectionProfiles({ page: 1, page_size: 100 }),
      listFuturesSelectionTemplates({ page: 1, page_size: 100, status: "ACTIVE" }),
      listFuturesSelectionRuns({ status: "SUCCEEDED", page: 1, page_size: 50 })
    ]);
    profiles.value = Array.isArray(profileData?.items) ? profileData.items : [];
    templates.value = Array.isArray(templateData?.items) ? templateData.items : [];
    runs.value = Array.isArray(runData?.items) ? runData.items : [];
    if (!selectedRunID.value && runs.value.length > 0) {
      selectedRunID.value = runs.value[0].run_id;
    }
  } catch (error) {
    ElMessage.error(error?.message || "加载合约池规则失败");
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
  <FuturesSelectionModuleShell
    title="智能期货合约池规则"
    description="把默认模板、默认配置和最近成功运行使用的合约池边界拆开查看，便于确认范围、风格和短历史兜底逻辑。"
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
          <el-descriptions-item label="研究风格">
            {{ formatFuturesSelectionStyle(defaultProfile?.style_default) }}
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
            {{ formatFuturesSelectionDateTime(selectedRun?.completed_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="候选 / 组合">
            {{ selectedRun?.candidate_count || 0 }} / {{ selectedRun?.selected_count || 0 }}
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
        <div class="card-title">手工合约列表</div>
        <div v-for="group in contractGroups" :key="group.label" class="restriction-group">
          <div class="restriction-label">{{ group.label }}</div>
          <div class="tag-wrap">
            <el-tag v-for="item in group.items" :key="`${group.label}-${item}`" type="info">{{ item }}</el-tag>
            <span v-if="group.items.length === 0" class="muted">未设置</span>
          </div>
        </div>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">最近运行回放</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="实际交易日">
            {{ selectedRun?.context_meta?.selected_trade_date || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="图谱摘要">
            {{ selectedRun?.context_meta?.graph_summary || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="相关实体数量">
            {{ (selectedRun?.context_meta?.related_entities || []).length || 0 }}
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
  </FuturesSelectionModuleShell>
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
