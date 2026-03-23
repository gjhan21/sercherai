<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  approveStockSelectionReview,
  getStockSelectionRun,
  getStockStrategyEnginePublishRecord,
  listStockSelectionPortfolio,
  listStockSelectionRuns,
  rejectStockSelectionReview
} from "../../api/admin";
import {
  formatStockSelectionDateTime,
  formatStockSelectionLabel,
  formatStockSelectionMarketRegime,
  formatStockSelectionReviewStatus,
  formatStockSelectionRiskLevel,
  formatStockSelectionRunStatus
} from "../../lib/stock-selection";
import { hasPermission } from "../../lib/session";

const route = useRoute();
const router = useRouter();
const canManage = hasPermission("stock_selection.manage");

const filters = ref({
  review_status: "PENDING"
});
const queueLoading = ref(false);
const detailLoading = ref(false);
const actionLoading = ref(false);
const runQueue = ref([]);
const selectedRunID = ref("");
const runDetail = ref(null);
const portfolioEntries = ref([]);
const publishRecord = ref(null);
const publishLoadError = ref("");

const publishedSnapshot = computed(() => {
  const snapshot = runDetail.value?.review?.published_portfolio_snapshot;
  return Array.isArray(snapshot) ? snapshot : [];
});

const reviewStatusSummary = computed(() => {
  const pending = runQueue.value.filter((item) => item.review_status === "PENDING").length;
  const approved = runQueue.value.filter((item) => item.review_status === "APPROVED").length;
  const rejected = runQueue.value.filter((item) => item.review_status === "REJECTED").length;
  return { pending, approved, rejected };
});

function tagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "APPROVED") return "success";
  if (normalized === "REJECTED") return "danger";
  if (normalized === "PENDING") return "warning";
  if (normalized === "SUCCEEDED") return "success";
  if (normalized === "FAILED") return "danger";
  return "info";
}

function syncRouteQuery(runID) {
  const nextQuery = { ...route.query };
  if (runID) {
    nextQuery.run_id = runID;
  } else {
    delete nextQuery.run_id;
  }
  router.replace({ query: nextQuery }).catch(() => {});
}

async function fetchRunQueue(preferredRunID = "") {
  queueLoading.value = true;
  try {
    const data = await listStockSelectionRuns({
      status: "SUCCEEDED",
      review_status: filters.value.review_status || undefined,
      page: 1,
      page_size: 50
    });
    runQueue.value = Array.isArray(data?.items) ? data.items : [];
    const queryRunID = String(preferredRunID || route.query.run_id || "").trim();
    const currentRunID = String(selectedRunID.value || "").trim();
    const nextRunID =
      (queryRunID && runQueue.value.find((item) => item.run_id === queryRunID)?.run_id) ||
      (currentRunID && runQueue.value.find((item) => item.run_id === currentRunID)?.run_id) ||
      runQueue.value[0]?.run_id ||
      "";
    selectedRunID.value = nextRunID;
    syncRouteQuery(nextRunID);
  } catch (error) {
    ElMessage.error(error?.message || "加载审核队列失败");
    runQueue.value = [];
    selectedRunID.value = "";
    syncRouteQuery("");
  } finally {
    queueLoading.value = false;
  }
}

async function fetchRunDetail() {
  if (!selectedRunID.value) {
    runDetail.value = null;
    portfolioEntries.value = [];
    publishRecord.value = null;
    publishLoadError.value = "";
    return;
  }
  detailLoading.value = true;
  try {
    const [run, portfolio] = await Promise.all([
      getStockSelectionRun(selectedRunID.value),
      listStockSelectionPortfolio(selectedRunID.value)
    ]);
    runDetail.value = run;
    portfolioEntries.value = Array.isArray(portfolio?.items) ? portfolio.items : [];
    const publishID = String(run?.review?.publish_id || run?.latest_publish_id || "").trim();
    publishRecord.value = null;
    publishLoadError.value = "";
    if (publishID) {
      try {
        publishRecord.value = await getStockStrategyEnginePublishRecord(publishID);
      } catch (error) {
        publishLoadError.value = error?.message || "加载发布归档失败";
      }
    }
  } catch (error) {
    ElMessage.error(error?.message || "加载审核详情失败");
    runDetail.value = null;
    portfolioEntries.value = [];
    publishRecord.value = null;
    publishLoadError.value = "";
  } finally {
    detailLoading.value = false;
  }
}

async function refreshPage(preferredRunID = "") {
  await fetchRunQueue(preferredRunID);
  await fetchRunDetail();
}

async function handleApprove(force = false) {
  if (!selectedRunID.value) {
    return;
  }
  actionLoading.value = true;
  try {
    const payload = {
      review_note: force ? "人工确认允许覆盖发布" : "审核工作台确认通过并发布",
      force,
      override_reason: ""
    };
    if (force) {
      const result = await ElMessageBox.prompt("请输入人工覆盖原因", "强制发布", {
        confirmButtonText: "发布",
        cancelButtonText: "取消"
      });
      payload.override_reason = result.value;
      payload.review_note = `人工覆盖发布：${result.value}`;
    }
    await approveStockSelectionReview(selectedRunID.value, payload);
    ElMessage.success(force ? "已完成强制发布" : "审核已通过并发布");
    await refreshPage(selectedRunID.value);
  } catch (error) {
    if (error === "cancel") {
      return;
    }
    ElMessage.error(error?.message || "审核通过失败");
  } finally {
    actionLoading.value = false;
  }
}

async function handleReject() {
  if (!selectedRunID.value) {
    return;
  }
  actionLoading.value = true;
  try {
    const result = await ElMessageBox.prompt("请输入驳回原因", "驳回审核", {
      confirmButtonText: "驳回",
      cancelButtonText: "取消"
    });
    await rejectStockSelectionReview(selectedRunID.value, {
      review_note: result.value
    });
    ElMessage.success("审核已驳回");
    await refreshPage(selectedRunID.value);
  } catch (error) {
    if (error === "cancel") {
      return;
    }
    ElMessage.error(error?.message || "驳回失败");
  } finally {
    actionLoading.value = false;
  }
}

watch(
  () => route.query.run_id,
  async (runID) => {
    const normalized = String(runID || "").trim();
    if (!normalized || normalized === selectedRunID.value) {
      return;
    }
    selectedRunID.value = normalized;
    await fetchRunDetail();
  }
);

watch(selectedRunID, async (value, oldValue) => {
  if (value === oldValue) {
    return;
  }
  syncRouteQuery(value);
  await fetchRunDetail();
});

onMounted(async () => {
  await refreshPage(String(route.query.run_id || "").trim());
});
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股审核与发布"
    description="把待审核 run、当前组合、发布归档和人工审核动作收口到独立工作台；候选页继续承接研究明细，这里专注审核闭环。"
  >
    <template #actions>
      <div class="toolbar review-toolbar">
        <el-select
          v-model="filters.review_status"
          placeholder="审核状态"
          clearable
          style="width: 160px"
          @change="refreshPage(selectedRunID)"
        >
          <el-option label="待审核" value="PENDING" />
          <el-option label="已通过" value="APPROVED" />
          <el-option label="已驳回" value="REJECTED" />
        </el-select>
        <el-button :loading="queueLoading || detailLoading" @click="refreshPage(selectedRunID)">刷新</el-button>
        <el-button
          v-if="selectedRunID"
          type="primary"
          plain
          @click="router.push({ name: 'stock-selection-candidates', query: { run_id: selectedRunID } })"
        >
          打开候选与组合
        </el-button>
      </div>
    </template>

    <div class="review-layout">
      <div class="card queue-card" v-loading="queueLoading">
        <div class="toolbar" style="justify-content: space-between; margin-bottom: 12px">
          <div class="card-title">审核队列</div>
          <div class="tag-wrap">
            <el-tag type="warning">待审核 {{ reviewStatusSummary.pending }}</el-tag>
            <el-tag type="success">已通过 {{ reviewStatusSummary.approved }}</el-tag>
            <el-tag type="danger">已驳回 {{ reviewStatusSummary.rejected }}</el-tag>
          </div>
        </div>
        <el-table
          :data="runQueue"
          border
          stripe
          size="small"
          height="620"
          empty-text="当前筛选条件下没有审核记录"
          row-key="run_id"
          highlight-current-row
          :current-row-key="selectedRunID"
          @row-click="(row) => selectedRunID = row.run_id"
        >
          <el-table-column prop="trade_date" label="交易日" min-width="104" />
          <el-table-column prop="run_id" label="运行编号" min-width="190" />
          <el-table-column label="审核状态" min-width="98">
            <template #default="{ row }">
              <el-tag :type="tagType(row.review_status)">
                {{ formatStockSelectionReviewStatus(row.review_status || "PENDING") }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="市场状态" min-width="112">
            <template #default="{ row }">{{ formatStockSelectionMarketRegime(row.market_regime) }}</template>
          </el-table-column>
          <el-table-column prop="publish_count" label="发布次数" min-width="90" />
          <el-table-column label="最近发布时间" min-width="160">
            <template #default="{ row }">{{ formatStockSelectionDateTime(row.latest_publish_at) }}</template>
          </el-table-column>
        </el-table>
      </div>

      <div class="detail-stack">
        <div class="card" v-loading="detailLoading">
          <template v-if="runDetail">
            <div class="toolbar" style="justify-content: space-between; flex-wrap: wrap">
              <div class="tag-wrap">
                <el-tag type="info">运行 {{ runDetail.run_id }}</el-tag>
                <el-tag :type="tagType(runDetail.status)">{{ formatStockSelectionRunStatus(runDetail.status) }}</el-tag>
                <el-tag :type="tagType(runDetail.review_status)">
                  {{ formatStockSelectionReviewStatus(runDetail.review_status || "PENDING") }}
                </el-tag>
                <el-tag type="warning">{{ formatStockSelectionMarketRegime(runDetail.market_regime) }}</el-tag>
                <el-tag type="success">组合 {{ runDetail.selected_count || 0 }}</el-tag>
                <el-tag type="info">发布 v{{ runDetail.latest_publish_version || 0 }}</el-tag>
              </div>
              <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
                <el-button
                  v-if="canManage && runDetail.review_status === 'PENDING'"
                  type="success"
                  :loading="actionLoading"
                  @click="handleApprove(false)"
                >
                  审核通过并发布
                </el-button>
                <el-button
                  v-if="canManage && runDetail.review_status === 'PENDING'"
                  type="warning"
                  :loading="actionLoading"
                  @click="handleApprove(true)"
                >
                  强制发布
                </el-button>
                <el-button
                  v-if="canManage && runDetail.review_status === 'PENDING'"
                  type="danger"
                  plain
                  :loading="actionLoading"
                  @click="handleReject"
                >
                  驳回
                </el-button>
              </div>
            </div>

            <el-descriptions :column="2" border size="small" style="margin-top: 12px">
              <el-descriptions-item label="结果摘要" :span="2">
                {{ runDetail.result_summary || "当前没有额外摘要。" }}
              </el-descriptions-item>
              <el-descriptions-item label="配置方案">
                {{ runDetail.profile_id }} / v{{ runDetail.profile_version }}
              </el-descriptions-item>
              <el-descriptions-item label="模板">
                {{ runDetail.template_name || "未指定模板" }}
              </el-descriptions-item>
              <el-descriptions-item label="审核人">
                {{ runDetail.review?.reviewer || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="审核备注">
                {{ runDetail.review?.review_note || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="最近发布记录">
                {{ runDetail.review?.publish_id || runDetail.latest_publish_id || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="最近发布时间">
                {{ formatStockSelectionDateTime(runDetail.review?.approved_at || runDetail.latest_publish_at) }}
              </el-descriptions-item>
              <el-descriptions-item label="版本差异" :span="2">
                新增：{{ (runDetail.compare_summary?.added_symbols || []).join("、") || "无" }}
                / 移除：{{ (runDetail.compare_summary?.removed_symbols || []).join("、") || "无" }}
              </el-descriptions-item>
            </el-descriptions>

            <div class="card tone-card" v-if="Array.isArray(runDetail.warning_messages) && runDetail.warning_messages.length">
              <div class="card-title">运行提醒</div>
              <div class="tag-wrap">
                <el-tag v-for="warning in runDetail.warning_messages" :key="warning" type="warning">
                  {{ warning }}
                </el-tag>
              </div>
            </div>
          </template>
          <el-empty v-else description="先从左侧选择一个待审核 run" :image-size="88" />
        </div>

        <div class="detail-grid" v-if="runDetail">
          <div class="card">
            <div class="card-title">当前组合</div>
            <el-table :data="portfolioEntries" border stripe size="small" height="280" empty-text="暂无组合结果">
              <el-table-column prop="rank" label="排名" min-width="68" />
              <el-table-column prop="symbol" label="代码" min-width="110" />
              <el-table-column prop="name" label="名称" min-width="120" />
              <el-table-column label="角色" min-width="90">
                <template #default="{ row }">{{ formatStockSelectionLabel(row.portfolio_role || "PORTFOLIO") }}</template>
              </el-table-column>
              <el-table-column prop="weight_suggestion" label="仓位建议" min-width="120" />
              <el-table-column label="风险" min-width="90">
                <template #default="{ row }">{{ formatStockSelectionRiskLevel(row.risk_level) }}</template>
              </el-table-column>
            </el-table>
          </div>

          <div class="card">
            <div class="card-title">已发布快照</div>
            <el-table
              v-if="publishedSnapshot.length"
              :data="publishedSnapshot"
              border
              stripe
              size="small"
              height="280"
            >
              <el-table-column prop="rank" label="排名" min-width="68" />
              <el-table-column prop="symbol" label="代码" min-width="110" />
              <el-table-column prop="name" label="名称" min-width="120" />
              <el-table-column label="角色" min-width="90">
                <template #default="{ row }">{{ formatStockSelectionLabel(row.portfolio_role || "PORTFOLIO") }}</template>
              </el-table-column>
              <el-table-column prop="weight_suggestion" label="仓位建议" min-width="120" />
            </el-table>
            <el-empty v-else description="当前 run 还没有固化的发布快照" :image-size="72" />
          </div>
        </div>

        <div class="card" v-if="runDetail">
          <div class="card-title">发布归档摘要</div>
          <el-alert
            v-if="publishLoadError"
            type="warning"
            show-icon
            :closable="false"
            :title="publishLoadError"
            style="margin-bottom: 12px"
          />
          <el-descriptions v-if="publishRecord" :column="2" border size="small">
            <el-descriptions-item label="发布记录">{{ publishRecord.publish_id }}</el-descriptions-item>
            <el-descriptions-item label="发布版本">v{{ publishRecord.version || 0 }}</el-descriptions-item>
            <el-descriptions-item label="归档时间">
              {{ formatStockSelectionDateTime(publishRecord.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="入选数量">
              {{ publishRecord.selected_count || 0 }}
            </el-descriptions-item>
            <el-descriptions-item label="发布方式">
              {{ publishRecord.replay?.force_publish ? "人工覆盖发布" : "按策略发布" }}
            </el-descriptions-item>
            <el-descriptions-item label="警告数量">
              {{ publishRecord.replay?.warning_count || 0 }}
            </el-descriptions-item>
            <el-descriptions-item label="摘要" :span="2">
              {{ publishRecord.report_summary || "当前没有额外归档摘要。" }}
            </el-descriptions-item>
            <el-descriptions-item label="覆盖原因" :span="2">
              {{ publishRecord.replay?.override_reason || "-" }}
            </el-descriptions-item>
          </el-descriptions>
          <el-empty v-else description="当前还没有可读取的发布归档。" :image-size="72" />
        </div>
      </div>
    </div>
  </StockSelectionModuleShell>
</template>

<style scoped>
.review-toolbar {
  margin-bottom: 0;
  flex-wrap: wrap;
}

.review-layout {
  display: grid;
  grid-template-columns: minmax(360px, 460px) minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.detail-stack {
  display: grid;
  gap: 16px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.queue-card {
  min-height: 700px;
}

.tag-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
}

.tone-card {
  margin-top: 12px;
}

@media (max-width: 1280px) {
  .review-layout,
  .detail-grid {
    grid-template-columns: 1fr;
  }

  .queue-card {
    min-height: auto;
  }
}
</style>
