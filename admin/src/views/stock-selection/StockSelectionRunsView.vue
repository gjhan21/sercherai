<script setup>
import { onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  compareStockSelectionRuns,
  createStockSelectionRun,
  getStockSelectionRun,
  getStrategyGraphSnapshot,
  queryStrategyGraphSubgraph,
  listStockSelectionProfiles,
  listStockSelectionRuns
} from "../../api/admin";
import {
  formatStockSelectionAssetDomain,
  formatStockSelectionDateTime,
  formatStockSelectionGraphEntityKey,
  formatStockSelectionGraphEntityType,
  formatStockSelectionGraphRelationType,
  formatStockSelectionGraphWriteStatus,
  formatStockSelectionMarketRegime,
  formatStockSelectionMode,
  formatStockSelectionPercent,
  formatStockSelectionReviewStatus,
  formatStockSelectionRunStatus,
  formatStockSelectionSource,
  formatStockSelectionStageDetail,
  formatStockSelectionStage,
  formatStockSelectionUniverseScope
} from "../../lib/stock-selection";
import { hasPermission } from "../../lib/session";

const route = useRoute();
const router = useRouter();
const canManage = hasPermission("stock_selection.manage");
const loading = ref(false);
const creating = ref(false);
const detailLoading = ref(false);
const errorMessage = ref("");
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const runs = ref([]);
const profiles = ref([]);
const detailVisible = ref(false);
const selectedRun = ref(null);
const selectedCompareRunIDs = ref([]);
const compareResult = ref(null);
const graphLoading = ref(false);
const graphErrorMessage = ref("");
const graphSnapshot = ref(null);
const graphSubgraphVisible = ref(false);
const graphSubgraphLoading = ref(false);
const graphSubgraphErrorMessage = ref("");
const graphSubgraph = ref(null);
const graphSubgraphEntity = ref(null);

const filters = reactive({
  status: "",
  review_status: "",
  profile_id: ""
});

function formatDateTime(value) {
  return formatStockSelectionDateTime(value);
}

function tagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "SUCCEEDED" || normalized === "APPROVED") return "success";
  if (normalized === "FAILED" || normalized === "REJECTED") return "danger";
  if (normalized === "RUNNING" || normalized === "PENDING") return "warning";
  return "info";
}

function getRunContext(run) {
  return run?.context_meta && typeof run.context_meta === "object" ? run.context_meta : {};
}

function getRunGraphSnapshotID(run) {
  return String(getRunContext(run)?.graph_snapshot_id || "").trim();
}

function getRunGraphWriteStatus(run) {
  return String(getRunContext(run)?.graph_write_status || "").trim();
}

function getRunGraphSummary(run) {
  return String(getRunContext(run)?.graph_summary || "").trim();
}

function getRunRelatedEntities(run) {
  const items = getRunContext(run)?.related_entities;
  if (!Array.isArray(items)) {
    return [];
  }
  return items
    .map((item, index) => ({
      key:
        String(item?.entity_key || item?.label || "").trim() ||
        `entity-${index}`,
      entityKey: String(item?.entity_key || item?.label || "").trim(),
      label: String(item?.label || item?.entity_key || "").trim(),
      type: String(item?.entity_type || item?.kind || "").trim(),
      assetDomain: String(item?.asset_domain || "").trim()
    }))
    .filter((item) => item.label);
}

function getRunMemoryFeedback(run) {
  const contextFeedback = getRunContext(run)?.memory_feedback;
  if (contextFeedback && typeof contextFeedback === "object") {
    return {
      summary: String(contextFeedback.summary || "").trim(),
      suggestions: Array.isArray(contextFeedback.suggestions)
        ? contextFeedback.suggestions
            .map((item) => String(item || "").trim())
            .filter(Boolean)
        : []
    };
  }
  const stageLog = Array.isArray(run?.stage_logs)
    ? run.stage_logs.find((item) => item.stage_key === "MEMORY_FEEDBACK")
    : null;
  return {
    summary: String(stageLog?.detail_message || "").trim(),
    suggestions: Array.isArray(stageLog?.payload_snapshot?.suggestions)
      ? stageLog.payload_snapshot.suggestions
          .map((item) => String(item || "").trim())
          .filter(Boolean)
      : []
  };
}

function resetGraphSnapshot() {
  graphLoading.value = false;
  graphErrorMessage.value = "";
  graphSnapshot.value = null;
}

function resetGraphSubgraph() {
  graphSubgraphVisible.value = false;
  graphSubgraphLoading.value = false;
  graphSubgraphErrorMessage.value = "";
  graphSubgraph.value = null;
  graphSubgraphEntity.value = null;
}

async function loadGraphSnapshot(run = selectedRun.value) {
  const snapshotID = getRunGraphSnapshotID(run);
  if (!snapshotID) {
    resetGraphSnapshot();
    return;
  }
  graphLoading.value = true;
  graphErrorMessage.value = "";
  try {
    graphSnapshot.value = await getStrategyGraphSnapshot(snapshotID);
  } catch (error) {
    graphSnapshot.value = null;
    graphErrorMessage.value = error?.message || "读取图快照失败";
  } finally {
    graphLoading.value = false;
  }
}

function getGraphEntityTags(snapshot) {
  const items = Array.isArray(snapshot?.entities) ? snapshot.entities : [];
  return items
    .slice(0, 12)
    .map((item, index) => ({
      key: String(item?.entity_key || item?.label || "").trim() || `graph-entity-${index}`,
      entityKey: String(item?.entity_key || item?.label || "").trim(),
      label: String(item?.label || item?.entity_key || "").trim(),
      type: String(item?.entity_type || "").trim(),
      assetDomain: String(item?.asset_domain || "").trim()
    }))
    .filter((item) => item.label);
}

function getGraphRelationTags(snapshot) {
  const items = Array.isArray(snapshot?.relations) ? snapshot.relations : [];
  return items
    .slice(0, 8)
    .map((item, index) => ({
      key: `${item?.source_key || "source"}-${item?.relation_type || "relation"}-${item?.target_key || "target"}-${index}`,
      label: `${formatStockSelectionGraphEntityKey(item?.source_key)} ${formatStockSelectionGraphRelationType(item?.relation_type)} ${formatStockSelectionGraphEntityKey(item?.target_key)}`
    }));
}

async function handleInspectGraphEntity(entity) {
  if (!entity?.type || !entity?.entityKey) {
    return;
  }
  graphSubgraphVisible.value = true;
  graphSubgraphLoading.value = true;
  graphSubgraphErrorMessage.value = "";
  graphSubgraph.value = null;
  graphSubgraphEntity.value = entity;
  try {
    graphSubgraph.value = await queryStrategyGraphSubgraph({
      entity_type: entity.type,
      entity_key: entity.entityKey,
      asset_domain: entity.assetDomain || undefined,
      depth: 2
    });
  } catch (error) {
    graphSubgraph.value = null;
    graphSubgraphErrorMessage.value = error?.message || "读取图谱子图失败";
  } finally {
    graphSubgraphLoading.value = false;
  }
}

async function fetchProfiles() {
  const data = await listStockSelectionProfiles({ page: 1, page_size: 100 });
  profiles.value = Array.isArray(data?.items) ? data.items : [];
}

async function fetchRuns() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const data = await listStockSelectionRuns({
      ...filters,
      page: page.value,
      page_size: pageSize.value
    });
    runs.value = Array.isArray(data?.items) ? data.items : [];
    total.value = Number(data?.total) || 0;
  } catch (error) {
    errorMessage.value = error?.message || "加载运行中心失败";
  } finally {
    loading.value = false;
  }
}

async function openRunDetail(runID) {
  if (!runID) {
    return;
  }
  detailVisible.value = true;
  detailLoading.value = true;
  resetGraphSnapshot();
  resetGraphSubgraph();
  try {
    selectedRun.value = await getStockSelectionRun(runID);
    await loadGraphSnapshot(selectedRun.value);
  } catch (error) {
    ElMessage.error(error?.message || "加载运行详情失败");
  } finally {
    detailLoading.value = false;
  }
}

async function handleCompare() {
  if (selectedCompareRunIDs.value.length < 2 || selectedCompareRunIDs.value.length > 3) {
    ElMessage.warning("请选择 2 到 3 条运行记录做对比");
    return;
  }
  try {
    compareResult.value = await compareStockSelectionRuns(selectedCompareRunIDs.value);
  } catch (error) {
    ElMessage.error(error?.message || "加载运行对比失败");
  }
}

function handleSelectionChange(rows) {
  selectedCompareRunIDs.value = Array.isArray(rows)
    ? rows.map((item) => item.run_id).slice(0, 3)
    : [];
}

async function handleQuickRun() {
  creating.value = true;
  try {
    const run = await createStockSelectionRun({});
    ElMessage.success(`已完成运行 ${run.run_id}`);
    await fetchRuns();
    router.replace({
      name: "stock-selection-runs",
      query: { run_id: run.run_id }
    });
    await openRunDetail(run.run_id);
  } catch (error) {
    ElMessage.error(error?.message || "创建运行失败");
  } finally {
    creating.value = false;
  }
}

watch(
  () => route.query.run_id,
  async (runID) => {
    if (runID) {
      await openRunDetail(String(runID));
    }
  }
);

onMounted(async () => {
  await Promise.all([fetchProfiles(), fetchRuns()]);
  if (route.query.run_id) {
    await openRunDetail(String(route.query.run_id));
  }
});
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股运行中心"
    description="集中查看运行记录、阶段摘要、审核状态和策略引擎发布摘要；新运行仍然同步等待 Python 返回结果。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchRuns">刷新列表</el-button>
        <el-button v-if="canManage" type="primary" :loading="creating" @click="handleQuickRun">
          立即运行
        </el-button>
        <el-button type="primary" plain @click="router.push({ name: 'stock-selection-profiles' })">
          打开策略设计
        </el-button>
        <el-button :disabled="selectedCompareRunIDs.length < 2" @click="handleCompare">对比选中运行</el-button>
      </div>
    </template>

    <div class="card">
      <div class="toolbar" style="flex-wrap: wrap">
        <el-select v-model="filters.status" clearable placeholder="运行状态" style="width: 140px">
          <el-option label="运行中" value="RUNNING" />
          <el-option label="成功" value="SUCCEEDED" />
          <el-option label="失败" value="FAILED" />
        </el-select>
        <el-select v-model="filters.review_status" clearable placeholder="审核状态" style="width: 140px">
          <el-option label="待审核" value="PENDING" />
          <el-option label="已通过" value="APPROVED" />
          <el-option label="已驳回" value="REJECTED" />
        </el-select>
        <el-select v-model="filters.profile_id" clearable placeholder="配置方案" style="width: 240px">
          <el-option
            v-for="profile in profiles"
            :key="profile.id"
            :label="`${profile.name}（v${profile.current_version || 0}）`"
            :value="profile.id"
          />
        </el-select>
        <el-button type="primary" @click="page = 1; fetchRuns()">查询</el-button>
      </div>

      <el-alert
        v-if="errorMessage"
        :title="errorMessage"
        type="error"
        show-icon
        style="margin-bottom: 12px"
      />

      <el-table
        :data="runs"
        border
        stripe
        size="small"
        v-loading="loading"
        empty-text="暂无运行记录"
        @selection-change="handleSelectionChange"
        @row-click="(row) => openRunDetail(row.run_id)"
      >
        <el-table-column type="selection" width="48" />
        <el-table-column prop="run_id" label="运行编号" min-width="180" />
        <el-table-column prop="trade_date" label="交易日" min-width="110" />
        <el-table-column prop="profile_id" label="配置方案" min-width="180" />
        <el-table-column prop="template_name" label="模板" min-width="140" />
        <el-table-column prop="market_regime" label="市场状态" min-width="120">
          <template #default="{ row }">{{ formatStockSelectionMarketRegime(row.market_regime) }}</template>
        </el-table-column>
        <el-table-column label="状态" min-width="100">
          <template #default="{ row }">
            <el-tag :type="tagType(row.status)">{{ formatStockSelectionRunStatus(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="审核" min-width="100">
          <template #default="{ row }">
            <el-tag :type="tagType(row.review_status)">{{ formatStockSelectionReviewStatus(row.review_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="universe_count" label="股票池" min-width="90" />
        <el-table-column prop="seed_count" label="种子池" min-width="90" />
        <el-table-column prop="candidate_count" label="候选池" min-width="100" />
        <el-table-column prop="selected_count" label="最终组合" min-width="100" />
        <el-table-column prop="publish_count" label="发布次数" min-width="100" />
        <el-table-column prop="latest_publish_version" label="最近版本" min-width="100" />
        <el-table-column prop="latest_publish_at" label="最近发布" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.latest_publish_at) }}</template>
        </el-table-column>
        <el-table-column prop="job_id" label="作业编号" min-width="170" />
      </el-table>

      <div class="toolbar" style="justify-content: flex-end; margin-top: 12px">
        <el-pagination
          background
          layout="prev, pager, next, total"
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          @current-change="(value) => { page = value; fetchRuns(); }"
        />
      </div>
    </div>

    <div class="card" v-if="compareResult?.items?.length">
      <div class="card-title">运行对比</div>
      <el-table :data="compareResult.items" border stripe size="small" empty-text="暂无对比数据">
        <el-table-column prop="run_id" label="运行编号" min-width="180" />
        <el-table-column prop="template_name" label="模板" min-width="140" />
        <el-table-column prop="market_regime" label="市场状态" min-width="120">
          <template #default="{ row }">{{ formatStockSelectionMarketRegime(row.market_regime) }}</template>
        </el-table-column>
        <el-table-column prop="selected_count" label="组合数量" min-width="100" />
        <el-table-column prop="added_symbols" label="新增标的" min-width="180">
          <template #default="{ row }">{{ (row.added_symbols || []).join("、") || "-" }}</template>
        </el-table-column>
        <el-table-column prop="removed_symbols" label="移除标的" min-width="180">
          <template #default="{ row }">{{ (row.removed_symbols || []).join("、") || "-" }}</template>
        </el-table-column>
        <el-table-column prop="portfolio_symbols" label="组合快照" min-width="220">
          <template #default="{ row }">{{ (row.portfolio_symbols || []).join("、") || "-" }}</template>
        </el-table-column>
      </el-table>
    </div>

    <el-drawer v-model="detailVisible" size="760px" title="运行详情">
      <div v-loading="detailLoading">
        <template v-if="selectedRun">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="运行编号">{{ selectedRun.run_id }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="tagType(selectedRun.status)">{{ formatStockSelectionRunStatus(selectedRun.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="配置方案">
              {{ selectedRun.profile_id }} / v{{ selectedRun.profile_version }}
            </el-descriptions-item>
            <el-descriptions-item label="模板 / 市场状态">
              {{ selectedRun.template_name || "未指定模板" }} / {{ formatStockSelectionMarketRegime(selectedRun.market_regime) }}
            </el-descriptions-item>
            <el-descriptions-item label="审核">
              <el-tag :type="tagType(selectedRun.review_status)">{{ formatStockSelectionReviewStatus(selectedRun.review_status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="最近发布">
              {{ selectedRun.latest_publish_id || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="发布时间">
              {{ formatDateTime(selectedRun.latest_publish_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="运行模式">
              {{ formatStockSelectionMode(selectedRun.selection_mode) }}
            </el-descriptions-item>
            <el-descriptions-item label="股票池范围">
              {{ formatStockSelectionUniverseScope(selectedRun.universe_scope) }}
            </el-descriptions-item>
            <el-descriptions-item label="真实交易日">
              {{ selectedRun.context_meta?.selected_trade_date || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="行情来源">
              {{ formatStockSelectionSource(selectedRun.context_meta?.price_source) }}
            </el-descriptions-item>
            <el-descriptions-item label="图快照编号">
              {{ getRunGraphSnapshotID(selectedRun) || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="图写入状态">
              {{ formatStockSelectionGraphWriteStatus(getRunGraphWriteStatus(selectedRun)) }}
            </el-descriptions-item>
            <el-descriptions-item label="结果摘要" :span="2">
              {{ selectedRun.result_summary || "-" }}
            </el-descriptions-item>
          </el-descriptions>

          <div class="toolbar" style="margin-top: 12px; justify-content: flex-end; flex-wrap: wrap">
            <el-button
              type="primary"
              plain
              @click="router.push({ name: 'stock-selection-candidates', query: { run_id: selectedRun.run_id } })"
            >
              打开候选与审核发布
            </el-button>
          </div>

          <div class="card" style="margin-top: 12px" v-if="selectedRun.context_meta">
            <div class="card-title">上下文与过滤条件</div>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="列表真相交易日">
                {{ selectedRun.context_meta?.selected_trade_date || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="上市天数近似过滤">
                {{ selectedRun.context_meta?.listing_days_filter_applied === false ? "已跳过" : "已启用" }}
              </el-descriptions-item>
              <el-descriptions-item label="价格区间">
                {{ selectedRun.context_meta?.universe_filters?.price_min ?? "-" }} ~
                {{ selectedRun.context_meta?.universe_filters?.price_max ?? "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="波动率区间">
                {{ selectedRun.context_meta?.universe_filters?.volatility_min ?? "-" }} ~
                {{ selectedRun.context_meta?.universe_filters?.volatility_max ?? "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="最少上市天数">
                {{ selectedRun.context_meta?.universe_filters?.min_listing_days ?? "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="20日均成交额下限">
                {{ selectedRun.context_meta?.universe_filters?.min_avg_turnover ?? "-" }}
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <div
            class="card"
            style="margin-top: 12px"
            v-if="
              getRunGraphSnapshotID(selectedRun) ||
              getRunGraphSummary(selectedRun) ||
              getRunRelatedEntities(selectedRun).length ||
              getRunMemoryFeedback(selectedRun).summary ||
              getRunMemoryFeedback(selectedRun).suggestions.length
            "
          >
            <div class="card-title">图谱快照与记忆反馈</div>
            <div class="toolbar" style="margin-bottom: 12px; justify-content: flex-end">
              <el-button
                size="small"
                :loading="graphLoading"
                :disabled="!getRunGraphSnapshotID(selectedRun)"
                @click="loadGraphSnapshot(selectedRun)"
              >
                刷新图快照
              </el-button>
            </div>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="图快照编号">
                {{ getRunGraphSnapshotID(selectedRun) || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="图写入状态">
                {{ formatStockSelectionGraphWriteStatus(getRunGraphWriteStatus(selectedRun)) }}
              </el-descriptions-item>
              <el-descriptions-item label="图谱摘要" :span="2">
                {{ getRunGraphSummary(selectedRun) || "本次运行还没有落出额外图谱摘要。" }}
              </el-descriptions-item>
              <el-descriptions-item label="记忆反馈" :span="2">
                {{ getRunMemoryFeedback(selectedRun).summary || "本次运行还没有额外记忆反馈。" }}
              </el-descriptions-item>
              <el-descriptions-item label="下次建议" :span="2">
                {{ getRunMemoryFeedback(selectedRun).suggestions.join("；") || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="节点 / 关系" :span="2">
                {{
                  graphSnapshot
                    ? `${graphSnapshot.entities?.length || 0} 个节点 / ${graphSnapshot.relations?.length || 0} 条关系`
                    : graphLoading
                      ? "图快照加载中"
                      : graphErrorMessage || "尚未读取图快照详情"
                }}
              </el-descriptions-item>
            </el-descriptions>
            <div class="chip-wrap" style="margin-top: 12px" v-if="getRunRelatedEntities(selectedRun).length">
              <el-tag
                v-for="item in getRunRelatedEntities(selectedRun)"
                :key="item.key"
                type="info"
                class="clickable-chip"
                @click="handleInspectGraphEntity(item)"
              >
                {{ item.label }}<template v-if="item.type"> / {{ formatStockSelectionGraphEntityType(item.type) }}</template>
              </el-tag>
            </div>
            <div class="chip-wrap" style="margin-top: 12px" v-if="getGraphEntityTags(graphSnapshot).length">
              <el-tag
                v-for="item in getGraphEntityTags(graphSnapshot)"
                :key="item.key"
                type="success"
                class="clickable-chip"
                @click="handleInspectGraphEntity(item)"
              >
                {{ item.label }}<template v-if="item.type"> / {{ formatStockSelectionGraphEntityType(item.type) }}</template>
              </el-tag>
            </div>
            <div class="chip-wrap" style="margin-top: 12px" v-if="getGraphRelationTags(graphSnapshot).length">
              <el-tag
                v-for="item in getGraphRelationTags(graphSnapshot)"
                :key="item.key"
                type="warning"
              >
                {{ item.label }}
              </el-tag>
            </div>
          </div>

          <div class="card" style="margin-top: 12px" v-if="selectedRun.compare_summary">
            <div class="card-title">与最近发布版本对比</div>
            <div class="chip-wrap">
              <el-tag type="success">
                新增：{{ (selectedRun.compare_summary?.added_symbols || []).join("、") || "无" }}
              </el-tag>
              <el-tag type="danger">
                移除：{{ (selectedRun.compare_summary?.removed_symbols || []).join("、") || "无" }}
              </el-tag>
            </div>
            <div class="mini-text">
              当前组合：{{ (selectedRun.compare_summary?.current_symbols || []).join("、") || "-" }}
            </div>
          </div>

          <div class="card" style="margin-top: 12px" v-if="(selectedRun.warning_messages || []).length">
            <div class="card-title">运行提醒</div>
            <div class="chip-wrap">
              <el-tag
                v-for="warning in selectedRun.warning_messages || []"
                :key="warning"
                type="warning"
              >
                {{ warning }}
              </el-tag>
            </div>
          </div>

          <div class="card" style="margin-top: 12px">
            <div class="card-title">阶段计数</div>
            <el-tag
              v-for="(count, key) in selectedRun.stage_counts || {}"
              :key="key"
              type="info"
              style="margin-right: 8px; margin-bottom: 8px"
            >
              {{ formatStockSelectionStage(key) }}：{{ count }}
            </el-tag>
          </div>

          <div class="card" style="margin-top: 12px" v-if="selectedRun.stage_durations_ms">
            <div class="card-title">阶段耗时</div>
            <div class="chip-wrap">
              <el-tag
                v-for="(duration, key) in selectedRun.stage_durations_ms || {}"
                :key="`duration-${key}`"
                type="info"
              >
                {{ formatStockSelectionStage(key) }}：{{ duration }} ms
              </el-tag>
            </div>
          </div>

          <div class="card" style="margin-top: 12px">
            <div class="card-title">阶段日志</div>
            <el-table
              :data="selectedRun.stage_logs || []"
              border
              stripe
              size="small"
              empty-text="暂无阶段日志"
            >
              <el-table-column prop="stage_key" label="阶段" min-width="120">
                <template #default="{ row }">{{ formatStockSelectionStage(row.stage_key) }}</template>
              </el-table-column>
              <el-table-column prop="input_count" label="输入" min-width="80" />
              <el-table-column prop="output_count" label="输出" min-width="80" />
              <el-table-column prop="duration_ms" label="耗时(ms)" min-width="100" />
              <el-table-column prop="detail_message" label="说明" min-width="220">
                <template #default="{ row }">{{ formatStockSelectionStageDetail(row.detail_message) }}</template>
              </el-table-column>
              <el-table-column label="转化率" min-width="110">
                <template #default="{ row }">
                  {{
                    row.input_count > 0
                      ? formatStockSelectionPercent(row.output_count / row.input_count, 1)
                      : "-"
                  }}
                </template>
              </el-table-column>
            </el-table>
          </div>
        </template>
      </div>
    </el-drawer>

    <el-drawer
      v-model="graphSubgraphVisible"
      append-to-body
      size="520px"
      :title="graphSubgraphEntity ? `图谱证据：${graphSubgraphEntity.label}` : '图谱证据'"
    >
      <div v-loading="graphSubgraphLoading">
        <el-alert
          v-if="graphSubgraphErrorMessage"
          :title="graphSubgraphErrorMessage"
          type="error"
          show-icon
          style="margin-bottom: 12px"
        />

        <template v-if="graphSubgraph">
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="根实体">
              {{ graphSubgraph.entity?.label || graphSubgraphEntity?.label || "-" }}
              <template v-if="graphSubgraph.entity?.entity_type || graphSubgraphEntity?.type">
                / {{ formatStockSelectionGraphEntityType(graphSubgraph.entity?.entity_type || graphSubgraphEntity?.type) }}
              </template>
            </el-descriptions-item>
            <el-descriptions-item label="实体域">
              {{ formatStockSelectionAssetDomain(graphSubgraph.entity?.asset_domain || graphSubgraphEntity?.assetDomain) }}
            </el-descriptions-item>
            <el-descriptions-item label="匹配快照">
              {{ (graphSubgraph.matched_snapshot_ids || []).join("、") || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="节点 / 关系">
              {{ graphSubgraph.entities?.length || 0 }} 个节点 / {{ graphSubgraph.relations?.length || 0 }} 条关系
            </el-descriptions-item>
            <el-descriptions-item label="图后端">
              {{ graphSubgraph.backend || "-" }}
            </el-descriptions-item>
          </el-descriptions>

          <div class="card" style="margin-top: 12px" v-if="(graphSubgraph.entities || []).length">
            <div class="card-title">关联节点</div>
            <div class="chip-wrap">
              <el-tag
                v-for="item in graphSubgraph.entities || []"
                :key="`${item.entity_type}-${item.entity_key}`"
                type="success"
              >
                {{ item.label || item.entity_key }}
                <template v-if="item.entity_type"> / {{ formatStockSelectionGraphEntityType(item.entity_type) }}</template>
              </el-tag>
            </div>
          </div>

          <div class="card" style="margin-top: 12px" v-if="(graphSubgraph.relations || []).length">
            <div class="card-title">关联关系</div>
            <div class="chip-wrap">
              <el-tag
                v-for="(item, index) in graphSubgraph.relations || []"
                :key="`${item.source_key}-${item.relation_type}-${item.target_key}-${index}`"
                type="warning"
              >
                {{ formatStockSelectionGraphEntityKey(item.source_key) }}
                {{ formatStockSelectionGraphRelationType(item.relation_type) }}
                {{ formatStockSelectionGraphEntityKey(item.target_key) }}
              </el-tag>
            </div>
          </div>
        </template>
      </div>
    </el-drawer>
  </StockSelectionModuleShell>
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

.mini-text {
  margin-top: 10px;
  color: var(--el-text-color-regular);
  font-size: 13px;
}

.clickable-chip {
  cursor: pointer;
}
</style>
