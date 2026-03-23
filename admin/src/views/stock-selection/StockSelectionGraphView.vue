<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  getStrategyGraphSnapshot,
  listStockSelectionRuns,
  queryStrategyGraphSubgraph
} from "../../api/admin";
import {
  formatStockSelectionAssetDomain,
  formatStockSelectionDateTime,
  formatStockSelectionGraphEntityKey,
  formatStockSelectionGraphEntityType,
  formatStockSelectionGraphRelationType,
  formatStockSelectionMarketRegime,
  formatStockSelectionRunStatus
} from "../../lib/stock-selection";

const graphEntityTypeOptions = [
  { label: "股票", value: "Stock" },
  { label: "题材", value: "ConceptTheme" },
  { label: "行业", value: "Industry" },
  { label: "公司", value: "Company" },
  { label: "指数", value: "Index" },
  { label: "政策/市场状态", value: "Policy" },
  { label: "事件", value: "Event" },
  { label: "研报", value: "ResearchReport" },
  { label: "供应链节点", value: "SupplyChainNode" }
];

const graphDepthOptions = [
  { label: "一跳关系", value: 1 },
  { label: "两跳关系", value: 2 }
];

const loading = ref(false);
const runs = ref([]);
const selectedRunID = ref("");
const selectedRun = ref(null);
const snapshotLoading = ref(false);
const snapshotErrorMessage = ref("");
const snapshot = ref(null);
const querying = ref(false);
const queryErrorMessage = ref("");
const subgraph = ref(null);
const hasQueried = ref(false);

const queryForm = reactive({
  entity_type: "",
  entity_key: "",
  asset_domain: "",
  depth: 2
});

const recentRuns = computed(() =>
  (Array.isArray(runs.value) ? runs.value : []).filter((item) => {
    const context = item?.context_meta && typeof item.context_meta === "object" ? item.context_meta : {};
    return Boolean(context.graph_snapshot_id) || Array.isArray(context.related_entities);
  })
);

const selectedRunContext = computed(() =>
  selectedRun.value?.context_meta && typeof selectedRun.value.context_meta === "object"
    ? selectedRun.value.context_meta
    : {}
);

const selectedRunRelatedEntities = computed(() =>
  (Array.isArray(selectedRunContext.value.related_entities) ? selectedRunContext.value.related_entities : [])
    .map((item, index) => ({
      key: String(item?.entity_key || item?.label || "").trim() || `run-entity-${index}`,
      entityType: String(item?.entity_type || "").trim(),
      entityKey: String(item?.entity_key || item?.label || "").trim(),
      label: String(item?.label || item?.entity_key || "").trim(),
      assetDomain: String(item?.asset_domain || "").trim()
    }))
    .filter((item) => item.entityType && item.entityKey)
);

function getSnapshotEntityTags(items = []) {
  return (Array.isArray(items) ? items : [])
    .slice(0, 18)
    .map((item, index) => ({
      key: String(item?.entity_key || item?.label || "").trim() || `snapshot-entity-${index}`,
      entityType: String(item?.entity_type || "").trim(),
      entityKey: String(item?.entity_key || item?.label || "").trim(),
      label: String(item?.label || item?.entity_key || "").trim(),
      assetDomain: String(item?.asset_domain || "").trim()
    }))
    .filter((item) => item.entityType && item.entityKey);
}

function getRelationTags(items = []) {
  return (Array.isArray(items) ? items : []).slice(0, 20).map((item, index) => ({
    key: `${item?.source_key || "source"}-${item?.relation_type || "relation"}-${item?.target_key || "target"}-${index}`,
    label: `${formatStockSelectionGraphEntityKey(item?.source_key)} ${formatStockSelectionGraphRelationType(item?.relation_type)} ${formatStockSelectionGraphEntityKey(item?.target_key)}`
  }));
}

async function fetchRuns() {
  loading.value = true;
  try {
    const data = await listStockSelectionRuns({ page: 1, page_size: 12 });
    runs.value = Array.isArray(data?.items) ? data.items : [];
    if (!selectedRunID.value && recentRuns.value.length > 0) {
      await handleSelectRun(recentRuns.value[0].run_id);
    } else if (selectedRunID.value) {
      selectedRun.value = recentRuns.value.find((item) => item.run_id === selectedRunID.value) || null;
    }
  } catch (error) {
    ElMessage.error(error?.message || "加载最近选股运行失败");
  } finally {
    loading.value = false;
  }
}

async function loadSnapshot(run = selectedRun.value) {
  snapshot.value = null;
  snapshotErrorMessage.value = "";
  const snapshotID = String(run?.context_meta?.graph_snapshot_id || "").trim();
  if (!snapshotID) {
    return;
  }
  snapshotLoading.value = true;
  try {
    snapshot.value = await getStrategyGraphSnapshot(snapshotID);
  } catch (error) {
    snapshotErrorMessage.value =
      error?.message || "当前图服务没有返回快照详情，可能是本地服务重启后旧快照已失效。";
  } finally {
    snapshotLoading.value = false;
  }
}

async function handleSelectRun(runID) {
  selectedRunID.value = runID;
  selectedRun.value = recentRuns.value.find((item) => item.run_id === runID) || null;
  subgraph.value = null;
  queryErrorMessage.value = "";
  hasQueried.value = false;
  if (selectedRunRelatedEntities.value.length > 0) {
    const first = selectedRunRelatedEntities.value[0];
    queryForm.entity_type = first.entityType;
    queryForm.entity_key = first.entityKey;
    queryForm.asset_domain = first.assetDomain || "";
  } else {
    queryForm.entity_type = "";
    queryForm.entity_key = "";
    queryForm.asset_domain = "";
  }
  await loadSnapshot(selectedRun.value);
}

async function submitQuery() {
  if (!String(queryForm.entity_type || "").trim() || !String(queryForm.entity_key || "").trim()) {
    ElMessage.warning("请先选择实体类型并填写实体键");
    return;
  }
  querying.value = true;
  queryErrorMessage.value = "";
  hasQueried.value = true;
  try {
    subgraph.value = await queryStrategyGraphSubgraph({
      entity_type: queryForm.entity_type,
      entity_key: queryForm.entity_key,
      asset_domain: queryForm.asset_domain || undefined,
      depth: queryForm.depth
    });
  } catch (error) {
    subgraph.value = null;
    queryErrorMessage.value = error?.message || "查询图谱子图失败";
  } finally {
    querying.value = false;
  }
}

async function handleUseEntity(entity, autoQuery = true) {
  queryForm.entity_type = entity.entityType || "";
  queryForm.entity_key = entity.entityKey || "";
  queryForm.asset_domain = entity.assetDomain || "";
  if (autoQuery) {
    await submitQuery();
  }
}

onMounted(fetchRuns);
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股事件图谱"
    description="集中查看最近选股运行的图谱摘要、快照状态和子图查询结果，让市场状态、题材和个股关系能在独立工作台里复盘。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchRuns">刷新最近运行</el-button>
      </div>
    </template>

    <div class="graph-layout">
      <div class="card">
        <div class="card-title">最近选股运行</div>
        <el-table
          :data="recentRuns"
          border
          stripe
          size="small"
          v-loading="loading"
          empty-text="暂无带图谱上下文的选股运行"
          highlight-current-row
          :current-row-key="selectedRunID"
          row-key="run_id"
          @row-click="(row) => handleSelectRun(row.run_id)"
        >
          <el-table-column prop="run_id" label="运行编号" min-width="180" />
          <el-table-column prop="template_name" label="模板" min-width="120" />
          <el-table-column prop="market_regime" label="市场状态" min-width="110">
            <template #default="{ row }">{{ formatStockSelectionMarketRegime(row.market_regime) }}</template>
          </el-table-column>
          <el-table-column label="状态" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionRunStatus(row.status) }}</template>
          </el-table-column>
        </el-table>
      </div>

      <div class="card">
        <div class="card-title">快照与运行摘要</div>
        <el-empty v-if="!selectedRun" description="请选择左侧运行记录" />
        <template v-else>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="运行编号">{{ selectedRun.run_id }}</el-descriptions-item>
            <el-descriptions-item label="图快照编号">
              {{ selectedRunContext.graph_snapshot_id || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="图谱摘要">
              {{ selectedRunContext.graph_summary || "当前运行还没有图谱摘要。" }}
            </el-descriptions-item>
            <el-descriptions-item label="最近完成时间">
              {{ formatStockSelectionDateTime(selectedRun.completed_at) }}
            </el-descriptions-item>
          </el-descriptions>

          <div class="toolbar" style="margin-top: 12px; justify-content: flex-end">
            <el-button
              size="small"
              :loading="snapshotLoading"
              :disabled="!selectedRunContext.graph_snapshot_id"
              @click="loadSnapshot(selectedRun)"
            >
              读取图快照
            </el-button>
          </div>

          <el-alert
            v-if="snapshotErrorMessage"
            :title="snapshotErrorMessage"
            type="warning"
            show-icon
            style="margin-top: 12px"
          />

          <el-descriptions v-if="snapshot" :column="2" border size="small" style="margin-top: 12px">
            <el-descriptions-item label="快照域">
              {{ formatStockSelectionAssetDomain(snapshot.asset_domain) }}
            </el-descriptions-item>
            <el-descriptions-item label="交易日">
              {{ snapshot.trade_date || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="快照摘要" :span="2">
              {{ snapshot.summary || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="节点 / 关系" :span="2">
              {{ (snapshot.entities || []).length }} 个节点 / {{ (snapshot.relations || []).length }} 条关系
            </el-descriptions-item>
          </el-descriptions>

          <div class="section-subtitle" style="margin-top: 14px">最近运行关联实体</div>
          <div class="chip-wrap">
            <el-tag
              v-for="item in selectedRunRelatedEntities"
              :key="item.key"
              type="info"
              class="clickable-chip"
              @click="handleUseEntity(item)"
            >
              {{ item.label }} / {{ formatStockSelectionGraphEntityType(item.entityType) }}
            </el-tag>
          </div>

          <template v-if="snapshot">
            <div class="section-subtitle" style="margin-top: 14px">快照节点</div>
            <div class="chip-wrap">
              <el-tag
                v-for="item in getSnapshotEntityTags(snapshot.entities)"
                :key="item.key"
                type="success"
                class="clickable-chip"
                @click="handleUseEntity(item)"
              >
                {{ item.label }} / {{ formatStockSelectionGraphEntityType(item.entityType) }}
              </el-tag>
            </div>

            <div class="section-subtitle" style="margin-top: 14px">快照关系</div>
            <div class="chip-wrap">
              <el-tag
                v-for="item in getRelationTags(snapshot.relations)"
                :key="item.key"
                type="warning"
              >
                {{ item.label }}
              </el-tag>
            </div>
          </template>
        </template>
      </div>
    </div>

    <div class="card" style="margin-top: 12px">
      <div class="card-title">子图查询</div>
      <div class="query-grid">
        <el-form-item label="实体类型">
          <el-select v-model="queryForm.entity_type" placeholder="请选择实体类型">
            <el-option
              v-for="item in graphEntityTypeOptions"
              :key="item.value"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="实体键">
          <el-input v-model="queryForm.entity_key" placeholder="例如 REGIME:DEFENSIVE 或 600519.SH" />
        </el-form-item>
        <el-form-item label="资产域">
          <el-input v-model="queryForm.asset_domain" placeholder="例如 跨资产（cross）或 股票（stock）" />
        </el-form-item>
        <el-form-item label="关系深度">
          <el-segmented v-model="queryForm.depth" :options="graphDepthOptions" />
        </el-form-item>
      </div>

      <div class="toolbar" style="justify-content: flex-end">
        <el-button type="primary" :loading="querying" @click="submitQuery">查询子图</el-button>
      </div>

      <el-alert
        v-if="queryErrorMessage"
        :title="queryErrorMessage"
        type="error"
        show-icon
        style="margin-top: 12px"
      />

      <div class="card graph-result-card" v-if="hasQueried" style="margin-top: 12px">
        <div class="card-title">查询结果</div>
        <el-descriptions :column="1" border size="small">
          <el-descriptions-item label="根实体">
            {{ subgraph?.entity?.label || queryForm.entity_key || "-" }}
            <template v-if="subgraph?.entity?.entity_type || queryForm.entity_type">
              / {{ formatStockSelectionGraphEntityType(subgraph?.entity?.entity_type || queryForm.entity_type) }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="实体域">
            {{ formatStockSelectionAssetDomain(subgraph?.entity?.asset_domain || queryForm.asset_domain) }}
          </el-descriptions-item>
          <el-descriptions-item label="匹配快照">
            {{ (subgraph?.matched_snapshot_ids || []).join("、") || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="节点 / 关系">
            {{ (subgraph?.entities || []).length }} 个节点 / {{ (subgraph?.relations || []).length }} 条关系
          </el-descriptions-item>
          <el-descriptions-item label="图后端">
            {{ subgraph?.backend || "-" }}
          </el-descriptions-item>
        </el-descriptions>

        <el-empty
          v-if="subgraph && !(subgraph.entities || []).length && !(subgraph.relations || []).length"
          description="当前查询没有返回可展开的节点或关系，可以换一个实体键继续尝试。"
          style="margin-top: 12px"
        />

        <template v-else-if="subgraph">
          <div class="section-subtitle" style="margin-top: 14px">关联节点</div>
          <div class="chip-wrap">
            <el-tag
              v-for="item in getSnapshotEntityTags(subgraph.entities)"
              :key="item.key"
              type="success"
              class="clickable-chip"
              @click="handleUseEntity(item)"
            >
              {{ item.label }} / {{ formatStockSelectionGraphEntityType(item.entityType) }}
            </el-tag>
          </div>

          <div class="section-subtitle" style="margin-top: 14px">关联关系</div>
          <div class="chip-wrap">
            <el-tag
              v-for="item in getRelationTags(subgraph.relations)"
              :key="item.key"
              type="warning"
            >
              {{ item.label }}
            </el-tag>
          </div>
        </template>
      </div>
    </div>
  </StockSelectionModuleShell>
</template>

<style scoped>
.graph-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr);
  gap: 12px;
}

.card-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}

.section-subtitle {
  margin-bottom: 8px;
  color: var(--el-text-color-primary);
  font-size: 13px;
  font-weight: 600;
}

.query-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.chip-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.clickable-chip {
  cursor: pointer;
}

.graph-result-card {
  background: linear-gradient(180deg, rgba(249, 250, 251, 0.92), rgba(243, 244, 246, 0.88));
}

@media (max-width: 1100px) {
  .graph-layout,
  .query-grid {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
