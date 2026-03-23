<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { ElMessage } from "element-plus";
import FuturesSelectionModuleShell from "../../components/FuturesSelectionModuleShell.vue";
import {
  listFuturesSelectionRuns,
  listStrategyAgentProfiles,
  listStrategyScenarioTemplates
} from "../../api/admin";
import {
  formatFuturesSelectionDateTime,
  formatFuturesSelectionGraphEntityType,
  formatFuturesSelectionMarketRegime
} from "../../lib/futures-selection";

const AGENT_LABELS = {
  trend: "趋势角色",
  event: "事件角色",
  liquidity: "流动性角色",
  risk: "风控角色",
  basis: "基差/结构角色",
  macro: "宏观角色",
  contrarian: "反方审查角色"
};

const AGENT_DESCRIPTIONS = {
  trend: "确认趋势是否还能延续，决定是否继续顺势。",
  event: "确认政策、资讯和事件面是否强化主逻辑。",
  liquidity: "确认合约流动性和执行条件，避免滑点过大。",
  risk: "专门负责否决和高风险拦截。",
  basis: "确认基差、期限结构、库存与定价锚点。",
  macro: "补充宏观与跨商品联动背景。",
  contrarian: "从反方视角验证观点漏洞。"
};

const loading = ref(false);
const runs = ref([]);
const agentProfiles = ref([]);
const scenarioTemplates = ref([]);
const selectedRunID = ref("");

const futuresAgentProfiles = computed(() =>
  agentProfiles.value.filter((item) => ["ALL", "FUTURES"].includes(String(item.target_type || "").toUpperCase()))
);

const futuresScenarioTemplates = computed(() =>
  scenarioTemplates.value.filter((item) => ["ALL", "FUTURES"].includes(String(item.target_type || "").toUpperCase()))
);

const defaultAgentProfile = computed(() => futuresAgentProfiles.value.find((item) => item.is_default && String(item.target_type).toUpperCase() === "FUTURES") || futuresAgentProfiles.value.find((item) => item.is_default) || futuresAgentProfiles.value[0] || null);
const defaultScenarioTemplate = computed(() => futuresScenarioTemplates.value.find((item) => item.is_default) || futuresScenarioTemplates.value[0] || null);
const selectedRun = computed(() => runs.value.find((item) => item.run_id === selectedRunID.value) || runs.value[0] || null);
const selectedRunContext = computed(() => selectedRun.value?.context_meta || {});

const agentRows = computed(() =>
  (defaultAgentProfile.value?.enabled_agents || []).map((agentKey) => ({
    key: agentKey,
    label: AGENT_LABELS[agentKey] || agentKey,
    description: AGENT_DESCRIPTIONS[agentKey] || "当前系统已启用该角色。"
  }))
);

const previewNotes = computed(() => {
  const notes = [];
  if (selectedRun.value?.market_regime) {
    notes.push(`当前运行市场状态为 ${formatFuturesSelectionMarketRegime(selectedRun.value.market_regime)}。`);
  }
  if (defaultAgentProfile.value) {
    notes.push(
      `当前默认启用了 ${agentRows.value.length} 个角色，正向阈值 ${defaultAgentProfile.value.positive_threshold}，反向阈值 ${defaultAgentProfile.value.negative_threshold}。`
    );
    if (defaultAgentProfile.value.allow_mock_fallback_on_short_history) {
      notes.push("角色配置允许在短历史场景下受控回退到模拟兜底，避免运行直接中断。");
    }
  }
  if (defaultScenarioTemplate.value?.items?.length) {
    notes.push(`当前场景模板包含 ${defaultScenarioTemplate.value.items.length} 个世界分支，可用于不同政策/供给/流动性冲击下的动作判断。`);
  } else {
    notes.push("当前没有读取到期货场景模板，页面先展示角色配置和最近运行上下文。");
  }
  if (selectedRunContext.value?.memory_feedback?.summary) {
    notes.push(`最近运行记忆反馈：${selectedRunContext.value.memory_feedback.summary}`);
  }
  return notes;
});

async function fetchData() {
  loading.value = true;
  try {
    const [runData, agentData, scenarioData] = await Promise.all([
      listFuturesSelectionRuns({ status: "SUCCEEDED", page: 1, page_size: 50 }),
      listStrategyAgentProfiles({ page: 1, page_size: 100, status: "ACTIVE" }),
      listStrategyScenarioTemplates({ page: 1, page_size: 100, status: "ACTIVE" })
    ]);
    runs.value = Array.isArray(runData?.items) ? runData.items : [];
    agentProfiles.value = Array.isArray(agentData?.items) ? agentData.items : [];
    scenarioTemplates.value = Array.isArray(scenarioData?.items) ? scenarioData.items : [];
    if (!selectedRunID.value && runs.value.length > 0) {
      selectedRunID.value = runs.value[0].run_id;
    }
  } catch (error) {
    ElMessage.error(error?.message || "加载期货多角色推演失败");
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
    title="智能期货多角色推演"
    description="把期货默认角色配置、场景模板和最近运行上下文放在一起看，帮助研究和审核统一风控与发布口径。"
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
        <div class="card-title">默认角色配置</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="名称">
            {{ defaultAgentProfile?.name || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="正向阈值">
            {{ defaultAgentProfile?.positive_threshold ?? "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="反向阈值">
            {{ defaultAgentProfile?.negative_threshold ?? "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="允许否决">
            {{ defaultAgentProfile?.allow_veto ? "允许" : "不允许" }}
          </el-descriptions-item>
          <el-descriptions-item label="短历史兜底">
            {{ defaultAgentProfile?.allow_mock_fallback_on_short_history ? "允许" : "关闭" }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">默认场景模板</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="名称">
            {{ defaultScenarioTemplate?.name || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="场景数量">
            {{ defaultScenarioTemplate?.items?.length || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="说明">
            {{ defaultScenarioTemplate?.description || "-" }}
          </el-descriptions-item>
        </el-descriptions>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">最近运行上下文</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="运行编号">
            {{ selectedRun?.run_id || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="市场状态">
            {{ formatFuturesSelectionMarketRegime(selectedRun?.market_regime) }}
          </el-descriptions-item>
          <el-descriptions-item label="完成时间">
            {{ formatFuturesSelectionDateTime(selectedRun?.completed_at) }}
          </el-descriptions-item>
          <el-descriptions-item label="图谱摘要">
            {{ selectedRunContext.graph_summary || "-" }}
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </div>

    <div class="roles-grid">
      <div class="card" v-loading="loading">
        <div class="card-title">启用角色</div>
        <div class="role-grid">
          <div v-for="item in agentRows" :key="item.key" class="role-card">
            <div class="role-title">{{ item.label }}</div>
            <div class="role-body">{{ item.description }}</div>
          </div>
        </div>
        <el-empty v-if="agentRows.length === 0" description="当前没有读取到角色配置" :image-size="72" />
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">场景分支</div>
        <el-table
          :data="defaultScenarioTemplate?.items || []"
          border
          stripe
          size="small"
          empty-text="当前没有读取到期货场景模板"
        >
          <el-table-column prop="label" label="场景" min-width="120" />
          <el-table-column prop="action" label="动作" min-width="100" />
          <el-table-column prop="risk_signal" label="风险信号" min-width="100" />
          <el-table-column prop="score_bias" label="分数偏置" min-width="100" />
          <el-table-column prop="thesis_template" label="推演主张" min-width="260" />
        </el-table>
      </div>
    </div>

    <div class="card" v-loading="loading">
      <div class="card-title">当前推演提示</div>
      <div class="note-list">
        <div v-for="note in previewNotes" :key="note" class="note-item">{{ note }}</div>
      </div>

      <div class="sub-card" v-if="Array.isArray(selectedRunContext.related_entities) && selectedRunContext.related_entities.length">
        <div class="sub-card__title">运行关联对象</div>
        <div class="tag-wrap">
          <el-tag
            v-for="entity in selectedRunContext.related_entities"
            :key="`${entity.entity_type}-${entity.entity_key || entity.label}`"
            type="info"
          >
            {{ entity.label || entity.entity_key }} / {{ formatFuturesSelectionGraphEntityType(entity.entity_type) }}
          </el-tag>
        </div>
      </div>

      <div class="sub-card" v-if="selectedRunContext.memory_feedback?.summary || (selectedRunContext.memory_feedback?.suggestions || []).length">
        <div class="sub-card__title">记忆反馈</div>
        <p class="muted">{{ selectedRunContext.memory_feedback?.summary || "-" }}</p>
        <div class="tag-wrap">
          <el-tag
            v-for="item in selectedRunContext.memory_feedback?.suggestions || []"
            :key="item"
            type="warning"
          >
            {{ item }}
          </el-tag>
        </div>
      </div>
    </div>
  </FuturesSelectionModuleShell>
</template>

<style scoped>
.overview-grid,
.roles-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 16px;
}

.roles-grid {
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.card-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}

.role-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.role-card {
  padding: 12px;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.04);
}

.role-title {
  font-weight: 700;
}

.role-body {
  margin-top: 6px;
  color: #475569;
  line-height: 1.5;
}

.note-list {
  display: grid;
  gap: 8px;
}

.note-item {
  padding: 12px;
  border-radius: 12px;
  background: rgba(15, 23, 42, 0.04);
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

.tag-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

@media (max-width: 1280px) {
  .overview-grid,
  .roles-grid,
  .role-grid {
    grid-template-columns: 1fr;
  }
}
</style>
