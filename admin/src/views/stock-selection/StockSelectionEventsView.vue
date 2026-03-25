<script setup>
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  getStockEventCluster,
  listStockEventClusters,
  queryStockEventSubgraph,
  reviewStockEventCluster
} from "../../api/admin";
import { buildStockEventSubgraphViewModel } from "../../lib/stock-event-admin";
import { hasPermission } from "../../lib/session";

const route = useRoute();
const router = useRouter();
const canManage = hasPermission("stock_selection.manage");

const loading = ref(false);
const detailLoading = ref(false);
const graphLoading = ref(false);
const actionLoading = ref(false);

const clusters = ref([]);
const total = ref(0);
const selectedClusterID = ref("");
const selectedCluster = ref(null);
const graphData = ref(null);

const page = ref(1);
const pageSize = ref(20);

const filters = reactive({
  review_status: "PENDING",
  review_priority: "",
  event_type: "",
  symbol: "",
  sector: "",
  topic: ""
});

const reviewForm = reactive({
  review_status: "APPROVED",
  review_note: "",
  reviewer: "",
  review_metadata_text: '{\n  "manual": true\n}'
});

const reviewReasonOptions = [
  { label: "全部优先级", value: "" },
  { label: "HIGH", value: "HIGH" },
  { label: "NORMAL", value: "NORMAL" }
];

const reviewStatusOptions = [
  { label: "全部状态", value: "" },
  { label: "PENDING", value: "PENDING" },
  { label: "APPROVED", value: "APPROVED" },
  { label: "REJECTED", value: "REJECTED" }
];

const eventTypeOptions = [
  { label: "全部事件", value: "" },
  { label: "NEWS", value: "NEWS" },
  { label: "ANNOUNCEMENT", value: "ANNOUNCEMENT" },
  { label: "EARNINGS", value: "EARNINGS" },
  { label: "POLICY", value: "POLICY" },
  { label: "INDUSTRY_THEME", value: "INDUSTRY_THEME" },
  { label: "SUPPLY_CHAIN_EVENT", value: "SUPPLY_CHAIN_EVENT" }
];

const selectedItems = computed(() => (Array.isArray(selectedCluster.value?.items) ? selectedCluster.value.items : []));
const selectedEntities = computed(() => (Array.isArray(selectedCluster.value?.entities) ? selectedCluster.value.entities : []));
const selectedReviewReasons = computed(() => {
  const value = selectedCluster.value?.metadata?.review_reason_codes;
  return Array.isArray(value) ? value : [];
});
const graphNodeCount = computed(() => Number(graphData.value?.nodes?.length || 0));
const graphEdgeCount = computed(() => Number(graphData.value?.edges?.length || 0));

function statusTagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "APPROVED" || normalized === "REVIEWED") return "success";
  if (normalized === "REJECTED") return "danger";
  if (normalized === "PENDING" || normalized === "CLUSTERED") return "warning";
  return "info";
}

function formatDateTime(value) {
  const raw = String(value || "").trim();
  if (!raw) {
    return "-";
  }
  const date = new Date(raw);
  if (Number.isNaN(date.getTime())) {
    return raw;
  }
  return date.toLocaleString("zh-CN", {
    hour12: false
  });
}

function normalizeReviewNote(cluster) {
  if (!cluster) {
    return "";
  }
  if ((cluster.review_status || "").toUpperCase() === "APPROVED") {
    return "事件归因成立，允许进入 reviewed event 证据层";
  }
  return "事件证据不足，暂不进入 reviewed event 真相源";
}

function syncRouteQuery(clusterID) {
  router.replace({
    query: {
      ...route.query,
      event_id: clusterID || undefined
    }
  });
}

async function fetchClusters() {
  loading.value = true;
  try {
    const data = await listStockEventClusters({
      ...filters,
      page: page.value,
      page_size: pageSize.value
    });
    clusters.value = Array.isArray(data?.items) ? data.items : [];
    total.value = Number(data?.total || 0);
    const routeClusterID = String(route.query.event_id || "").trim();
    const candidateID =
      routeClusterID ||
      selectedClusterID.value ||
      (clusters.value.length > 0 ? clusters.value[0].id : "");
    if (candidateID) {
      await selectCluster(candidateID);
      return;
    }
    selectedClusterID.value = "";
    selectedCluster.value = null;
    graphData.value = null;
  } catch (error) {
    ElMessage.error(error?.message || "加载股票事件审核队列失败");
  } finally {
    loading.value = false;
  }
}

async function loadGraph(clusterID) {
  if (!clusterID) {
    graphData.value = null;
    return;
  }
  graphLoading.value = true;
  try {
    const payload = await queryStockEventSubgraph(clusterID, { depth: 2 });
    graphData.value = buildStockEventSubgraphViewModel(payload);
  } catch (error) {
    graphData.value = buildStockEventSubgraphViewModel({
      warning_message: error?.message || "图谱增强失败，可稍后重试"
    });
  } finally {
    graphLoading.value = false;
  }
}

async function selectCluster(clusterID) {
  const normalized = String(clusterID || "").trim();
  if (!normalized) {
    selectedClusterID.value = "";
    selectedCluster.value = null;
    graphData.value = null;
    syncRouteQuery("");
    return;
  }
  selectedClusterID.value = normalized;
  detailLoading.value = true;
  try {
    const data = await getStockEventCluster(normalized);
    selectedCluster.value = data;
    reviewForm.review_status = "APPROVED";
    reviewForm.review_note = normalizeReviewNote(data);
    syncRouteQuery(normalized);
    await loadGraph(normalized);
  } catch (error) {
    ElMessage.error(error?.message || "加载事件详情失败");
  } finally {
    detailLoading.value = false;
  }
}

function resetFilters() {
  filters.review_status = "PENDING";
  filters.review_priority = "";
  filters.event_type = "";
  filters.symbol = "";
  filters.sector = "";
  filters.topic = "";
  page.value = 1;
  fetchClusters();
}

function handlePageChange(nextPage) {
  page.value = nextPage;
  fetchClusters();
}

async function submitReview() {
  if (!selectedClusterID.value) {
    return;
  }
  const reviewNote = reviewForm.review_note.trim();
  if (!reviewNote) {
    ElMessage.warning("请填写审核说明");
    return;
  }

  let reviewMetadata = {};
  const reviewMetadataText = String(reviewForm.review_metadata_text || "").trim();
  if (reviewMetadataText) {
    try {
      reviewMetadata = JSON.parse(reviewMetadataText);
    } catch (error) {
      ElMessage.warning("review_metadata 需要是合法 JSON");
      return;
    }
  }

  actionLoading.value = true;
  try {
    const data = await reviewStockEventCluster(selectedClusterID.value, {
      review_status: reviewForm.review_status,
      review_note: reviewNote,
      reviewer: reviewForm.reviewer.trim(),
      review_metadata: reviewMetadata
    });
    selectedCluster.value = data;
    ElMessage.success(reviewForm.review_status === "APPROVED" ? "事件已审核通过" : "事件已驳回");
    await fetchClusters();
  } catch (error) {
    ElMessage.error(error?.message || "提交事件审核失败");
  } finally {
    actionLoading.value = false;
  }
}

watch(
  () => route.query.event_id,
  async (value) => {
    const normalized = String(value || "").trim();
    if (!normalized || normalized === selectedClusterID.value) {
      return;
    }
    await selectCluster(normalized);
  }
);

onMounted(fetchClusters);
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股事件审核台"
    description="把新闻聚类后的股票事件收口成可审核队列，左侧看优先级和 SLA，中间看事件成员与实体，右侧直接执行 APPROVED / REJECTED。"
  >
    <template #actions>
      <div class="toolbar stock-event-actions">
        <el-button :loading="loading" @click="fetchClusters">刷新队列</el-button>
      </div>
    </template>

    <div class="stock-event-layout">
      <div class="card stock-event-panel stock-event-panel--list">
        <div class="card-title">事件队列</div>
        <div class="toolbar stock-event-filter-grid">
          <el-select v-model="filters.review_status" placeholder="审核状态">
            <el-option v-for="item in reviewStatusOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-select v-model="filters.review_priority" placeholder="优先级">
            <el-option v-for="item in reviewReasonOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-select v-model="filters.event_type" placeholder="事件类型">
            <el-option v-for="item in eventTypeOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-input v-model="filters.symbol" placeholder="股票代码，如 600519.SH" clearable />
          <el-input v-model="filters.sector" placeholder="行业标签" clearable />
          <el-input v-model="filters.topic" placeholder="主题标签" clearable />
        </div>
        <div class="toolbar stock-event-toolbar-end">
          <el-button type="primary" @click="page = 1; fetchClusters()">应用筛选</el-button>
          <el-button @click="resetFilters">重置</el-button>
        </div>
        <el-table
          :data="clusters"
          border
          stripe
          size="small"
          highlight-current-row
          row-key="id"
          :current-row-key="selectedClusterID"
          v-loading="loading"
          empty-text="暂无待审核股票事件"
          @row-click="(row) => selectCluster(row.id)"
        >
          <el-table-column prop="title" label="事件标题" min-width="180" />
          <el-table-column prop="event_type" label="类型" min-width="120" />
          <el-table-column prop="primary_symbol" label="主股票" min-width="110" />
          <el-table-column label="优先级" min-width="110">
            <template #default="{ row }">
              <el-tag :type="row.metadata?.review_priority === 'HIGH' ? 'danger' : 'info'">
                {{ row.metadata?.review_priority || "NORMAL" }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="审核状态" min-width="110">
            <template #default="{ row }">
              <el-tag :type="statusTagType(row.review_status)">{{ row.review_status }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="新闻数" min-width="90">
            <template #default="{ row }">{{ row.news_count || 0 }}</template>
          </el-table-column>
        </el-table>
        <div class="pagination">
          <el-text type="info">第 {{ page }} 页，共 {{ total }} 条</el-text>
          <el-pagination
            background
            layout="prev, pager, next"
            :current-page="page"
            :page-size="pageSize"
            :total="total"
            @current-change="handlePageChange"
          />
        </div>
      </div>

      <div class="card stock-event-panel" v-loading="detailLoading">
        <div class="card-title">事件详情</div>
        <el-empty v-if="!selectedCluster" description="请选择左侧事件 cluster" />
        <template v-else>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="Cluster ID">{{ selectedCluster.id }}</el-descriptions-item>
            <el-descriptions-item label="事件类型">{{ selectedCluster.event_type }}</el-descriptions-item>
            <el-descriptions-item label="主股票">{{ selectedCluster.primary_symbol || "-" }}</el-descriptions-item>
            <el-descriptions-item label="主题 / 行业">
              {{ selectedCluster.topic_label || "-" }} / {{ selectedCluster.sector_label || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="聚类状态">
              <el-tag :type="statusTagType(selectedCluster.status)">{{ selectedCluster.status }}</el-tag>
              <el-tag :type="statusTagType(selectedCluster.review_status)" style="margin-left: 8px">
                {{ selectedCluster.review_status }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="审核优先级">
              <el-tag :type="selectedCluster.metadata?.review_priority === 'HIGH' ? 'danger' : 'info'">
                {{ selectedCluster.metadata?.review_priority || "NORMAL" }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="原因码">
              <div class="tag-row">
                <el-tag v-for="item in selectedReviewReasons" :key="item" size="small" effect="plain">{{ item }}</el-tag>
                <el-text v-if="selectedReviewReasons.length === 0" type="info">无</el-text>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="更新时间">{{ formatDateTime(selectedCluster.updated_at) }}</el-descriptions-item>
            <el-descriptions-item label="摘要">
              {{ selectedCluster.summary || "当前聚类没有摘要，需依靠成员新闻原文审核。" }}
            </el-descriptions-item>
          </el-descriptions>

          <div class="stock-event-section">
            <div class="section-title">成员新闻</div>
            <el-timeline>
              <el-timeline-item
                v-for="item in selectedItems"
                :key="item.id"
                :timestamp="formatDateTime(item.published_at)"
                placement="top"
              >
                <div class="member-title">{{ item.title }}</div>
                <div class="member-meta">{{ item.source_key || "-" }} / {{ item.primary_symbol || "-" }}</div>
                <div class="member-summary">{{ item.summary || "无摘要" }}</div>
              </el-timeline-item>
            </el-timeline>
          </div>

          <div class="stock-event-section">
            <div class="section-title">事件实体</div>
            <div class="tag-row">
              <el-tag
                v-for="entity in selectedEntities"
                :key="entity.id"
                effect="plain"
                size="small"
              >
                {{ entity.entity_type }} · {{ entity.label || entity.entity_key }}
              </el-tag>
            </div>
          </div>

          <div class="stock-event-section">
            <div class="section-title">最新审核记录</div>
            <el-descriptions :column="1" border size="small">
              <el-descriptions-item label="审核人">
                {{ selectedCluster.latest_review?.reviewer || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="审核状态">
                {{ selectedCluster.latest_review?.review_status || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="审核说明">
                {{ selectedCluster.latest_review?.review_note || "-" }}
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </template>
      </div>

      <div class="card stock-event-panel">
        <div class="card-title">审核动作与图谱预览</div>
        <el-empty v-if="!selectedCluster" description="选中事件后可执行审核与查看子图" />
        <template v-else>
          <el-alert
            v-if="selectedCluster.metadata?.graph_sync_warning"
            :title="selectedCluster.metadata.graph_sync_warning"
            type="warning"
            :closable="false"
            show-icon
            style="margin-bottom: 16px"
          />
          <el-alert
            :title="selectedCluster.review_status === 'PENDING' ? '当前事件仍在审核队列，可直接审批。' : '当前事件已完成审核，仍可查看图谱增强结果。'"
            :type="selectedCluster.review_status === 'PENDING' ? 'warning' : 'info'"
            :closable="false"
            show-icon
            style="margin-bottom: 16px"
          />

          <el-form label-width="84px" class="stock-event-review-form">
            <el-form-item label="审核结论">
              <el-radio-group v-model="reviewForm.review_status" :disabled="!canManage || actionLoading">
                <el-radio-button value="APPROVED">APPROVED</el-radio-button>
                <el-radio-button value="REJECTED">REJECTED</el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="审核人">
              <el-input v-model="reviewForm.reviewer" placeholder="可留空，默认使用当前管理员" :disabled="!canManage || actionLoading" />
            </el-form-item>
            <el-form-item label="审核说明">
              <el-input
                v-model="reviewForm.review_note"
                type="textarea"
                :rows="4"
                maxlength="300"
                show-word-limit
                resize="vertical"
                placeholder="告诉系统为什么批准 / 驳回这个事件"
                :disabled="!canManage || actionLoading"
              />
            </el-form-item>
            <el-form-item label="元数据">
              <el-input
                v-model="reviewForm.review_metadata_text"
                type="textarea"
                :rows="5"
                resize="vertical"
                placeholder='{"manual":true}'
                :disabled="!canManage || actionLoading"
              />
            </el-form-item>
          </el-form>
          <div class="toolbar stock-event-toolbar-end">
            <el-button :disabled="actionLoading" @click="loadGraph(selectedClusterID)">刷新子图</el-button>
            <el-button type="primary" :disabled="!canManage" :loading="actionLoading" @click="submitReview">
              提交审核
            </el-button>
          </div>

          <div class="stock-event-section">
            <div class="section-title">Stock Event 子图</div>
            <el-skeleton v-if="graphLoading" :rows="4" animated />
            <template v-else>
              <el-alert
                v-if="graphData?.warning_message"
                :title="graphData.warning_message"
                type="warning"
                :closable="false"
                show-icon
                style="margin-bottom: 12px"
              />
              <el-descriptions :column="1" border size="small">
                <el-descriptions-item label="节点数">{{ graphNodeCount }}</el-descriptions-item>
                <el-descriptions-item label="关系数">{{ graphEdgeCount }}</el-descriptions-item>
              </el-descriptions>
              <div class="stock-event-section" style="padding-top: 0">
                <div class="section-subtitle">节点预览</div>
                <div class="tag-row">
                  <el-tag
                    v-for="node in graphData?.nodes || []"
                    :key="`${node.entity_type}-${node.entity_key}`"
                    size="small"
                    effect="plain"
                  >
                    {{ node.entity_type }} · {{ node.label || node.entity_key }}
                  </el-tag>
                </div>
              </div>
            </template>
          </div>
        </template>
      </div>
    </div>
  </StockSelectionModuleShell>
</template>

<style scoped>
.stock-event-layout {
  display: grid;
  grid-template-columns: 1.1fr 1.1fr 0.9fr;
  gap: 16px;
  align-items: start;
}

.stock-event-panel {
  min-height: 720px;
}

.stock-event-panel--list {
  overflow: hidden;
}

.stock-event-actions,
.stock-event-toolbar-end {
  justify-content: flex-end;
}

.stock-event-filter-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.stock-event-section {
  margin-top: 16px;
}

.section-title {
  font-weight: 600;
  margin-bottom: 10px;
}

.section-subtitle {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin: 12px 0 8px;
}

.tag-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.member-title {
  font-weight: 600;
  margin-bottom: 4px;
}

.member-meta {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-bottom: 4px;
}

.member-summary {
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-regular);
}

@media (max-width: 1400px) {
  .stock-event-layout {
    grid-template-columns: 1fr;
  }

  .stock-event-panel {
    min-height: auto;
  }
}
</style>
